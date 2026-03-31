// Package cronjob 定时任务
package cronjob

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// DefaultCacheSize 默认缓冲区大小
	DefaultCacheSize = 10
	// defaultFolder 日志存储目录
	defaultFolder = "/osslog"
	// defaultKeepHistory 默认保留最近记录的时间, 单位: d
	defaultKeepHistory = 7
	// defaultDagInsThreshold 默认运行实例历史记录阈值
	defaultDagInsThreshold = 1000000
	// defaultTaskInsThreshold 默认运行实例详情历史记录阈值
	defaultTaskInsThreshold = 2000000
	// maxFileSize 默认文件大小
	maxFileSize = 500 * 1024 * 1024 // 500MB
)

// DefaultOSSNotFound 默认对象存储未找到
const DefaultOSSNotFound = float64(404031002)

// DumpLog 日志转储接口
type DumpLog interface {
	DumpLogToOSS(ctx context.Context)
}

type dumpLog struct {
	config     *common.Config
	store      mod.Store
	ossGateway drivenadapters.OssGateWay
}

var (
	dOnce               sync.Once
	dl                  DumpLog
	filterDagInsStatus  = []string{string(entity.DagInstanceStatusCancled), string(entity.DagInstanceStatusFailed), string(entity.DagInstanceStatusSuccess)}
	filterTaskInsStatus = []string{string(entity.TaskInstanceStatusCanceled), string(entity.TaskInstanceStatusEnding), string(entity.TaskInstanceStatusSkipped),
		string(entity.TaskInstanceStatusSuccess), string(entity.TaskInstanceStatusFailed)}
)

// ParseBSON 解析BSON结构体函数
type ParseBSON func(bson.M) ([]string, error)

// DeleteDataFunc 清除数据库记录函数
type DeleteDataFunc func(ctx context.Context, params map[string]interface{}) error

// DumpLogTask 日志转储入参结构体
type DumpLogTask struct {
	fileName    string
	dataChannel chan []bson.M
	wg          *sync.WaitGroup
	// 解析函数
	parseFunc ParseBSON
	// 任务状态
	status []string
	// 更新时间
	updatedAt int64
	// deleteOpt 0: delete dag 1: delete task
	deleteOpt int
	// 对象存储id
	OSSID string
	// 文件索引
	Index int
	// 日志持久化路径列表
	FilePaths []string
	// 分片中最后一条记录id
	ShardStartIDs []string
	// 分片中任务id列表
	ShardTaskMap map[string][]string
}

// NewDumpLog new dump log instance
func NewDumpLog() DumpLog {
	dOnce.Do(func() {
		dl = &dumpLog{
			config:     common.NewConfig(),
			store:      mod.GetStore(),
			ossGateway: drivenadapters.NewOssGateWay(),
		}
	})

	return dl
}

// DumpLogToOSS 日志转储
func (d *dumpLog) DumpLogToOSS(ctx context.Context) {
	var (
		err        error
		folderPath string
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	folderPath, err = utils.CreateFolder(defaultFolder)
	if err != nil {
		log.Warnf("[DumpLogToOSS] initParentFolder failed, detail: %s", err.Error())
		return
	}

	suffix := generateSuffix()
	sevenDaysAgoTimestamp := time.Now().Add(d.keepRecentlyRecordTime()).Unix()

	// dag ins 数量统计
	params := map[string]interface{}{
		"$and": []bson.M{
			{"status": bson.M{"$in": filterDagInsStatus}},
			{"updatedAt": bson.M{"$lte": sevenDaysAgoTimestamp}},
		},
		"log_clean_ttl": 10 * time.Minute,
	}
	dagInsCount, err := d.store.GetDagInstanceCount(ctx, params)
	if err != nil {
		log.Warnf("[DumpLogToOSS] GetDagInstanceCount failed, detail: %s", err.Error())
		return
	}
	dagThreshold := d.config.DumpLog.DagThreshold
	if dagThreshold <= 0 {
		dagThreshold = defaultDagInsThreshold
	}
	log.Infof("[DumpLogToOSS] GetDagInstanceCount count: %v, threshold: %v", dagInsCount, dagThreshold)

	// task ins 数量统计
	params = map[string]interface{}{
		"$and": []bson.M{
			{"status": bson.M{"$in": filterTaskInsStatus}},
			{"updatedAt": bson.M{"$lte": time.Now().Unix()}},
		},
		"log_clean_ttl": 10 * time.Minute,
	}
	taskInsCount, err := d.store.GetTaskInstanceCount(ctx, params)
	if err != nil {
		log.Warnf("[DumpLogToOSS] GetTaskInstanceCount failed, detail: %s", err.Error())
		return
	}
	taskThreshold := d.config.DumpLog.TaskThreshold
	if taskThreshold <= 0 {
		taskThreshold = defaultTaskInsThreshold
	}
	log.Infof("[DumpLogToOSS] GetTaskInstanceCount count: %v, threshold: %v", taskInsCount, taskThreshold)

	if dagInsCount < dagThreshold && taskInsCount < taskThreshold {
		return
	}

	// 删除dagIns
	start := time.Now()
	fileName := fmt.Sprintf("%s/automationDagInsLog_%v", folderPath, suffix)
	d.buildHistoryDagInsData(ctx, fileName, sevenDaysAgoTimestamp)
	cost := time.Since(start)
	log.Infof("[DumpLogToOSS] execute cron job build history dag ins data cost time: %s", cost)

	// 删除taskIns
	start = time.Now()
	fileName = fmt.Sprintf("%s/automationTaskInsLog_%v", folderPath, suffix)
	d.buildHistoryTaskInsData(ctx, fileName, sevenDaysAgoTimestamp)
	cost = time.Since(start)
	log.Infof("[DumpLogToOSS] execute cron job build history task ins data cost time: %s", cost)
}

func generateSuffix() string {
	// 生成文件统一后缀
	id, _ := utils.GetUniqueID()
	return fmt.Sprintf("%v_%v", time.Now().Format("2006-01-02"), id)
}

// keepRecentlyRecordTime 获取保存最近多久的历史记录时间
func (d *dumpLog) keepRecentlyRecordTime() time.Duration {
	var keepHistoryStr = strings.TrimSpace(d.config.DumpLog.KeepHistory)
	if strings.HasSuffix(keepHistoryStr, "d") {
		// 按天保留
		keepHistory, err := strconv.ParseInt(strings.TrimSuffix(keepHistoryStr, "d"), 10, 64)
		if err != nil || keepHistory <= 0 {
			keepHistory = defaultKeepHistory
		}
		return -1 * time.Duration(keepHistory) * 24 * time.Hour
	} else if strings.HasSuffix(keepHistoryStr, "h") {
		// 按小时
		keepHistory, err := strconv.ParseInt(strings.TrimSuffix(keepHistoryStr, "h"), 10, 64)
		if err != nil || keepHistory <= 0 {
			return -1 * time.Duration(defaultKeepHistory) * 24 * time.Hour
		}
		return -1 * time.Duration(keepHistory) * time.Hour
	} else {
		// 默认情况 7天
		return -1 * time.Duration(defaultKeepHistory) * 24 * time.Hour
	}
}

func (d *dumpLog) buildHistoryDagInsData(ctx context.Context, fileName string, timeStamp int64) { //nolint
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dataChannel := make(chan []bson.M, DefaultCacheSize)
	dumpLogTask := &DumpLogTask{
		fileName:      fileName,
		dataChannel:   dataChannel,
		wg:            &sync.WaitGroup{},
		parseFunc:     parseDagIns,
		status:        filterDagInsStatus,
		updatedAt:     timeStamp,
		deleteOpt:     0,
		Index:         0,
		FilePaths:     []string{},
		ShardStartIDs: []string{},
		ShardTaskMap:  map[string][]string{},
	}

	dumpLogTask.wg.Add(1)
	subCtx, cancle := context.WithCancel(ctx)
	go d.dumpLogToOSS(subCtx, dumpLogTask, cancle, d.store.DeleteDagInsByID)

	// 获取dag实例历史数据
	params := map[string]interface{}{"status": dumpLogTask.status, "updatedAt": dumpLogTask.updatedAt}
	err = d.store.ListHistoryDagIns(subCtx, params, dataChannel)
	if err != nil {
		log.Warnf("[buildHistoryDagInsData] ListHistoryDagIns failed, detail: %s", err.Error())
		close(dataChannel)
		return
	}

	dumpLogTask.wg.Wait()
}

func (d *dumpLog) buildHistoryTaskInsData(ctx context.Context, fileName string, timeStamp int64) { //nolint
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dataChannel := make(chan []bson.M, DefaultCacheSize)
	dumpLogTask := &DumpLogTask{
		fileName:      fileName,
		dataChannel:   dataChannel,
		wg:            &sync.WaitGroup{},
		parseFunc:     parseTaskIns,
		status:        filterTaskInsStatus,
		updatedAt:     timeStamp,
		deleteOpt:     1,
		OSSID:         "",
		Index:         0,
		FilePaths:     []string{},
		ShardStartIDs: []string{},
		ShardTaskMap:  map[string][]string{},
	}

	dumpLogTask.wg.Add(1)
	subCtx, cancle := context.WithCancel(ctx)
	go d.dumpLogToOSS(subCtx, dumpLogTask, cancle, d.store.DeleteTaskInsByID)

	// 获取dag实例历史数据
	params := map[string]interface{}{"status": dumpLogTask.status, "updatedAt": dumpLogTask.updatedAt}
	err = d.store.ListHistoryTaskIns(subCtx, params, dataChannel)
	if err != nil {
		log.Warnf("[buildHistoryTaskInsData] ListHistoryTaskIns failed, detail: %s", err.Error())
		close(dataChannel)
		return
	}

	dumpLogTask.wg.Wait()
}

func (d *dumpLog) dumpLogToOSS(ctx context.Context, dumpLogTask *DumpLogTask, cancle context.CancelFunc, deleteDataFunc DeleteDataFunc) {
	var err error
	defer func() {
		cancle()
		for _, filePath := range dumpLogTask.FilePaths {
			_ = os.Remove(filePath)
		}
		dumpLogTask.wg.Done()
	}()

	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 创建CSV写入器
	fileName := fmt.Sprintf("%v_%v.csv", dumpLogTask.fileName, dumpLogTask.Index)
	file, writer, err := d.createCSVWriter(fileName)
	defer func() { _ = file.Close() }()
	defer writer.Flush()
	if err != nil {
		log.Warnf("[dumpLogToOSS] createCSVWriter failed, index: %v, detail: %v", dumpLogTask.Index, err.Error())
		return
	}

	dumpLogTask.FilePaths = append(dumpLogTask.FilePaths, fileName)

	// 获取oss信息
	dumpLogTask.OSSID, err = d.ossGateway.GetAvaildOSS(ctx)
	if err != nil {
		log.Warnf("[dumpLogToOSS] GetAvaildOSS failed, detail: %v", err.Error())
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case values, ok := <-dumpLogTask.dataChannel:
			if !ok {
				// 调用 Flush, 确保数据缓冲区未满时也写入
				writer.Flush()
				// 检查是否有写入错误
				if err = writer.Error(); err != nil {
					log.Warnf("[dumpLogToOSS] Error writing row to CSV: %v", err.Error())
					return
				}
				_ = file.Close()

				// 推送到oss
				// 记录oss相关信息
				// 删除历史记录
				err = d.uploadFileToOSS(ctx, dumpLogTask, deleteDataFunc)
				if err != nil {
					log.Infof("[dumpLogToOSS] uploadFileToOSS filed, detail: %s", err.Error())
					return
				}
				return
			}
			var finishedDags, runningDags, allDags []string
			if dumpLogTask.deleteOpt == 1 {
				// taskins判断dag是否执行完成
				for _, value := range values {
					allDags = append(allDags, fmt.Sprintf("%v", value["dagInsId"]))
				}
				allDags = utils.RemoveRepByMap(allDags)
				runningDags, err = d.store.ListExistDagInsID(ctx, allDags)
				if err != nil {
					log.Warnf("[dumpLogToOSS] ListExistDagInsID failed, detail: %v", err.Error())
					return
				}
				_, finishedDags = utils.Arrcmp(allDags, runningDags)
			}

			if dumpLogTask.deleteOpt == 1 && len(finishedDags) == 0 {
				continue
			}

			for index, value := range values {
				rows, err := dumpLogTask.parseFunc(value)
				if err != nil {
					return
				}
				// 记录分片中最后一条记录id
				if index == len(values)-1 {
					dumpLogTask.ShardStartIDs = append(dumpLogTask.ShardStartIDs, rows[0])
					dumpLogTask.ShardTaskMap[rows[0]] = finishedDags
				}

				// 如果当前记录dagins还存在则不删除
				if utils.IsContain(rows[4], runningDags) {
					continue
				}
				err = writer.Write(rows)
				if err != nil {
					log.Warnf("[dumpLogToOSS] Error writing row to CSV: %v", err.Error())
					return
				}

				// 日志文件大小限制
				fileHeader, _ := file.Stat()
				fileSize := fileHeader.Size()
				// 获取文件信息, 大于500MB，重新生成文件存储
				if fileSize >= maxFileSize {
					dumpLogTask.Index++
					fileName = fmt.Sprintf("%v_%v.csv", dumpLogTask.fileName, dumpLogTask.Index)
					dumpLogTask.FilePaths = append(dumpLogTask.FilePaths, fileName)
					writer.Flush()
					_ = file.Close()
					file, writer, err = d.createCSVWriter(fileName)
					if err != nil {
						log.Warnf("[dumpLogToOSS] createCSVWriter failed, index: %v, detail: %v", dumpLogTask.Index, err.Error())
						return
					}
				}
			}
		}
	}
}

func (d *dumpLog) uploadFileToOSS(ctx context.Context, dumpLogTask *DumpLogTask, deleteDataFunc DeleteDataFunc) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	currentDir, _ := os.Getwd()
	var entityLogs []*entity.Log
	for i := 0; i < len(dumpLogTask.FilePaths); i++ {
		file, err := os.Open(dumpLogTask.FilePaths[i])
		if err != nil {
			return err
		}
		fileInfo, err := file.Stat()
		if err != nil {
			return err
		}

		fileSize := fileInfo.Size()
		if fileSize <= 0 {
			continue
		}
		key := uuid.New().String()

		err = d.ossGateway.UploadFile(ctx, dumpLogTask.OSSID, key, true, file, fileSize)
		if err != nil {
			return err
		}

		entityLogs = append(entityLogs, &entity.Log{
			OssID:    dumpLogTask.OSSID,
			Key:      key,
			FileName: strings.TrimPrefix(dumpLogTask.FilePaths[i], currentDir),
		})
	}

	if len(entityLogs) == 0 {
		log.Warnf("[dumpLogToOSS] CreateLog entityLogs is empty")
		return nil
	}

	err = d.store.CreateLogs(ctx, entityLogs)
	if err != nil {
		log.Warnf("[dumpLogToOSS] CreateLog failed, detail: %v", err.Error())
		return err
	}

	params := map[string]interface{}{
		"status":    dumpLogTask.status,
		"updatedAt": dumpLogTask.updatedAt,
	}
	for i := 0; i < len(dumpLogTask.ShardStartIDs); i++ {
		params["_id"] = dumpLogTask.ShardStartIDs[i]
		params["dagInsIDs"] = dumpLogTask.ShardTaskMap[dumpLogTask.ShardStartIDs[i]]
		err = deleteDataFunc(ctx, params)
		if err != nil {
			log.Warnf("[dumpLogToOSS] deleteDataFunc failed, detail: %v", err.Error())
			return err
		}
	}
	return nil
}

func (d *dumpLog) createCSVWriter(fileName string) (*os.File, *csv.Writer, error) {
	file, err := os.Create(fileName)
	if err != nil {
		return nil, nil, err
	}

	writer := csv.NewWriter(file)
	return file, writer, nil
}

func parseDagIns(data bson.M) ([]string, error) {
	var result []string
	dataByte, err := bson.Marshal(data)
	if err != nil {
		return result, err
	}
	var dagIns entity.DagInstance
	err = bson.Unmarshal(dataByte, &dagIns)
	if err != nil {
		return result, err
	}
	vars, _ := json.Marshal(dagIns.Vars)
	shareData, _ := json.Marshal(dagIns.ShareData)
	result = []string{
		dagIns.ID,
		fmt.Sprintf("%v", dagIns.CreatedAt),
		fmt.Sprintf("%v", dagIns.UpdatedAt),
		dagIns.DagID,
		fmt.Sprintf("%v", dagIns.Trigger),
		dagIns.Worker,
		string(vars),
		string(shareData),
		fmt.Sprintf("%v", dagIns.Status),
		dagIns.UserID,
		dagIns.Reason,
		fmt.Sprintf("%v", dagIns.EndedAt),
	}
	return result, nil
}

func parseTaskIns(data bson.M) ([]string, error) {
	var result []string
	dataByte, err := bson.Marshal(data)
	if err != nil {
		return result, err
	}
	var taskIns entity.TaskInstance
	err = bson.Unmarshal(dataByte, &taskIns)
	if err != nil {
		return result, err
	}
	params, _ := json.Marshal(taskIns.Params)
	reason, _ := json.Marshal(taskIns.Reason)
	result = []string{
		taskIns.ID,
		fmt.Sprintf("%v", taskIns.CreatedAt),
		fmt.Sprintf("%v", taskIns.UpdatedAt),
		taskIns.TaskID,
		taskIns.DagInsID,
		fmt.Sprintf("%v", taskIns.DependOn),
		taskIns.ActionName,
		fmt.Sprintf("%v", taskIns.TimeoutSecs),
		string(params),
		fmt.Sprintf("%v", taskIns.Status),
		string(reason),
	}
	return result, nil
}
