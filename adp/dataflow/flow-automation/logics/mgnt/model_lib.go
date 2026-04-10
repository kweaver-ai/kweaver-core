package mgnt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/dependency"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
)

const extractFileSizeLimit int64 = 20 * 1024 * 1024

// ModelLibHandler 模型库接口
type ModelLibHandler interface {
	RecognizeText(ctx context.Context, params map[string]interface{}, userInfo *drivenadapters.UserInfo) (map[string]interface{}, error)
	AudioTransfer(ctx context.Context, docID string, userInfo *drivenadapters.UserInfo) (map[string]interface{}, error)
	EntityExtract(ctx context.Context, trainID, docID string, userInfo *drivenadapters.UserInfo) (map[string]interface{}, error)
	TrainModule(ctx context.Context, trainID string, userInfo *drivenadapters.UserInfo) (string, error)
	UploadTrainFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, int, int, error)
	CreateTagsRule(ctx context.Context, userInfo *drivenadapters.UserInfo, params *TagRulesMol) (uint64, error)
	GetModelInfoByID(ctx context.Context, id, userid string) (*ModelInfo, error)
	DeleteModelInfoByID(ctx context.Context, id string, userInfo *drivenadapters.UserInfo) error
	ListModelInfo(ctx context.Context, userid string, status, limit, offset int64) ([]ModelInfo, error)
	UpdateModelInfo(ctx context.Context, id string, userInfo *drivenadapters.UserInfo, params *TagRulesMol) error
	ExtractTagsByRule(ctx context.Context, params *TagExtractionParams, userInfo *drivenadapters.UserInfo) ([]string, error)
}

var (
	mlOnce sync.Once
	ml     ModelLibHandler
)

type modelLib struct {
	efast          drivenadapters.Efast
	oss            drivenadapters.OssGateWay
	dependency     dependency.Repo
	executeMethods entity.ExecuteMethods
	logger         drivenadapters.Logger
}

// Entity 实体信息
type Entity struct {
	Text        string  `json:"text"`
	Start       int64   `json:"start"`
	End         int64   `json:"end"`
	Probability float64 `json:"probability"`
}

type TrainLog struct {
	ID     uint64 `json:"id"`
	TaskID int    `json:"task_id"`
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	Status string `json:"status"`
}

type TagRulesMol struct {
	Rules       *[]TagRules `json:"rules,omitempty"`
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	ID          uint64      `json:"id"`
	Status      *int64      `json:"status"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	Rules       interface{} `json:"rules,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	ID          string      `json:"id"`
	Status      int64       `json:"status"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
	Type        int64       `json:"type,omitempty"`
}

type TagExtractionParams = dependency.TagExtractionParams

type TagRules = dependency.TagRules

type TagRuleDetail = dependency.TagRuleDetail

type UpdateParams = dependency.UpdateParams

// FileMetaData 训练文件元数据信息
type FileMetaData struct {
	FileName string `json:"fileName"`
	FileSize string `json:"fileSize"`
	Schema   string `json:"schema"`
}

// NewModelLib 实例化模型库
func NewModelLib() ModelLibHandler {
	mlOnce.Do(func() {
		ml = &modelLib{
			efast:      drivenadapters.NewEfast(),
			oss:        drivenadapters.NewOssGateWay(),
			dependency: dependency.NewDriven(),
			executeMethods: entity.ExecuteMethods{
				Publish: mod.NewMQHandler().Publish,
			},
			logger: drivenadapters.NewLogger(),
		}
	})

	return ml
}

// RecognizeText OCR 识别图片
func (ml *modelLib) RecognizeText(ctx context.Context, params map[string]interface{}, userInfo *drivenadapters.UserInfo) (map[string]interface{}, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)

	docID := params["docid"].(string)
	attr, err := ml.efast.GetDocMsg(ctx, docID)
	if err != nil {
		tLog.Warnf("[logic.RecognizeText] GetDocMsg err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err.Error())
	}

	perm, err := ml.efast.CheckPerm(ctx, docID, "display", strings.TrimPrefix(userInfo.TokenID, "Bearer "), userInfo.LoginIP)
	if err != nil {
		tLog.Warnf("[logic.RecognizeText] CheckPerm err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err.Error())
	}

	if perm != 0 {
		err = ierrors.NewIError(ierrors.NoPermission, "", map[string]interface{}{
			"info": "has no perm to get doc metadata",
			"doc": map[string]string{
				"docid":   docID,
				"docname": attr.Name,
			},
		})
		return nil, err
	}

	res, err := ml.dependency.RecognizeText(ctx, params, attr)
	if err != nil {
		tLog.Warnf("[logic.RecognizeText] RecognizeText err, detail: %s", err.Error())
		return res, err
	}
	return res, nil
}

// AudioTransfer OCR 识别图片
func (ml *modelLib) AudioTransfer(ctx context.Context, docID string, userInfo *drivenadapters.UserInfo) (map[string]interface{}, error) {
	var err error
	var sizeLimmit int64 = 10 * 1024 * 1024
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)

	attr, err := ml.efast.GetDocMsg(ctx, docID)
	if err != nil {
		tLog.Warnf("[logic.AudioTransfer] GetDocMsg err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err.Error())
	}

	body, err := ml.efast.DownloadFile(ctx, docID, "", strings.TrimPrefix(userInfo.TokenID, "Bearer "), userInfo.LoginIP)
	if err != nil {
		tLog.Warnf("[logic.AudioTransfer] DownloadFile err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", err.Error())
	}

	res, err := ml.dependency.AudioTransfer(ctx, float64(sizeLimmit), "", body, attr)
	if err != nil {
		tLog.Warnf("[logic.AudioTransfer] AudioTransfer err, detail: %s", err.Error())
		return res, err
	}
	return res, nil
}

// EntityExtract 实体抽取
func (ml *modelLib) EntityExtract(ctx context.Context, trainID, docID string, userInfo *drivenadapters.UserInfo) (map[string]interface{}, error) {
	var (
		err error
		res = make(map[string]interface{}, 0)
	)

	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)

	entitys, err := ml.dependency.StartInfer(ctx, trainID, docID, strings.TrimPrefix(userInfo.TokenID, "Bearer "), userInfo.LoginIP, extractFileSizeLimit)
	if err != nil {
		tLog.Warnf("[logic.EntityExtract] StartInfer err, detail: %s", err.Error())
		return res, err
	}
	items := entitys["res"].([]interface{})
	for _, val := range items {
		itemMap := val.(map[string]interface{})
		for key, entity := range itemMap {
			// var vals []string
			var entitys []Entity
			entityByte, _ := json.Marshal(entity)
			_ = json.Unmarshal(entityByte, &entitys)
			// for _, entity := range entitys {
			// 	vals = append(vals, entity.Text)
			// }
			res[key] = entitys
		}
	}
	return res, nil
}

// TrainModule 模型训练
func (ml *modelLib) TrainModule(ctx context.Context, trainID string, userInfo *drivenadapters.UserInfo) (string, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)

	status, err := ml.dependency.StartTrainModule(ctx, trainID, userInfo.UserID)
	if err != nil {
		tLog.Warnf("[logic.EntityExtract] StartInfer err, detail: %s", err.Error())
		return status, err
	}
	return status, nil
}

// UploadTrainFile 创建
func (ml *modelLib) UploadTrainFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, int, int, error) {
	var (
		lines   int
		err     error
		trainID uint64
		schema  []string
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)

	if !strings.HasSuffix(fileHeader.Filename, ".json") {
		return "", lines, len(schema), ierrors.NewIError(ierrors.InvalidParameter, ierrors.UnSupportedFileType, map[string]interface{}{
			"name": fileHeader.Filename,
			"type": "support file type: json",
		})
	}

	// 查询已模型数量, 数量大于0,则禁止再创建
	exist, err := ml.dependency.VerifyTaskUnique(ctx)
	if err != nil {
		tLog.Warnf("[UploadTrainFile] VerifyTaskUnique failed, detail: %v", err.Error())
		return "", lines, len(schema), err
	}
	if exist {
		return "", lines, len(schema), ierrors.NewIError(ierrors.OperationDenied, ierrors.NumberOfTasksLimited, nil)
	}
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", lines, len(schema), err
	}

	lines = len(strings.Split(strings.TrimSpace(string(fileBytes)), "\n"))

	re := regexp.MustCompile(`"label":\s*"([^"]+)"`) // 匹配 "label": "value"
	matches := re.FindAllStringSubmatch(string(fileBytes), -1)
	for _, match := range matches {
		schema = append(schema, match[1])
	}
	schema = utils.RemoveRepByMap(schema)

	var metadata = map[string]interface{}{
		"fileName": fileHeader.Filename,
		"fileSize": fileHeader.Size,
		"lines":    lines,
		"schema":   strings.Join(schema, ","),
	}

	metadataBytes, _ := json.Marshal(metadata)

	file.Seek(0, 0)
	var files []common.File
	files = append(files, common.File{Reader: file, FileName: fileHeader.Filename, FileSize: fileHeader.Size})
	files = append(files, common.File{Reader: bytes.NewReader(metadataBytes), FileName: "metadata.txt", FileSize: int64(len(metadataBytes))})

	var zipBuf bytes.Buffer
	if err := common.Compress(&zipBuf, files); err != nil {
		return "", lines, len(schema), err
	}

	ossID, err := ml.oss.GetAvaildOSS(ctx)
	key := uuid.New().String()
	// 推送到oss
	err = ml.oss.UploadFile(ctx, ossID, key, true, bytes.NewReader(zipBuf.Bytes()), int64(zipBuf.Len()))
	if err != nil {
		tLog.Warnf("[UploadTrainFile] UploadFile failed, detail: %v", err.Error())
		return "", lines, len(schema), err
	}

	trainID, _ = utils.GetUniqueID()
	createdAt := time.Now().UnixNano() / 1e6
	data := &rds.AiModel{
		ID:          trainID,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		TrainStatus: common.TrainStatusInit,
		Status:      common.Init,
		Rule:        "",
		Type:        common.UIEType,
	}

	id, _ := utils.GetUniqueID()
	trainFileInfo := &rds.TrainFileOSSInfo{
		ID:        id,
		TrainID:   trainID,
		OSSID:     ossID,
		Key:       key,
		CreatedAt: createdAt,
	}

	err = ml.dependency.CreateTrainFile(ctx, data, trainFileInfo)
	if err != nil {
		tLog.Warnf("[UploadTrainFile] CreateTrainFile failed, detail: %v", err.Error())
	}

	return fmt.Sprintf("%v", trainID), lines, len(schema), err
}

// CreateTagsRule 创建标签提取规则
func (ml *modelLib) CreateTagsRule(ctx context.Context, userInfo *drivenadapters.UserInfo, params *TagRulesMol) (uint64, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)
	r, _ := json.Marshal(*params.Rules)

	// 重名检测, 名称全局唯一
	exist, err := ml.dependency.CheckDupName(ctx, *params.Name)
	if err != nil {
		tLog.Warnf("[logic.CreateTagsRule] CheckDupName err, detail: %s", err.Error())
		return 0, err
	}
	if exist {
		return 0, ierrors.NewIError(ierrors.DuplicatedName, "", nil)
	}

	if params.Description == nil {
		*params.Description = ""
	}

	id, err := ml.dependency.CreateTagsRule(ctx, *params.Name, userInfo.UserID, string(r), *params.Description, int(*params.Status))
	if err != nil {
		tLog.Warnf("[logic.CreateTagsRule] CreateTagsRule err, detail: %s", err.Error())
		return id, err
	}

	go func() {
		detail, extMsg := common.GetLogBody("createCustomCapabily", []interface{}{*params.Name},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		ml.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: fmt.Sprintf("%v", id),
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		}, &drivenadapters.JSONLogWriter{SendFunc: ml.executeMethods.Publish})
	}()

	return id, err
}

// GetModelInfoByID 根据id获取标签处理规则
func (ml *modelLib) GetModelInfoByID(ctx context.Context, id, userid string) (*ModelInfo, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)

	conditions, err := ml.getQueryConditionByModelType(ctx, id, userid)
	if err != nil {
		tLog.Warnf("[logic.GetModelInfoByID] getQueryConditionByModelType err, detail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	info, err := ml.dependency.GetModelInfoByID(ctx, conditions)
	if err != nil {
		tLog.Warnf("[logic.GetModelInfoByID] GetTrainFileInfo err, detail: %s", err.Error())
		return nil, err
	}

	res := &ModelInfo{
		ID:          fmt.Sprintf("%v", info.ID),
		Name:        info.Name,
		CreatedAt:   info.CreatedAt,
		UpdatedAt:   info.UpdatedAt,
		Description: info.Description,
		Status:      int64(info.Status),
		Type:        int64(info.Type),
	}

	switch info.Type {
	case common.UIEType:
		schemas := strings.Split(info.Rule, ",")
		res.Rules = schemas
	case common.TagRuleType:
		var rules []TagRules
		err = json.Unmarshal([]byte(info.Rule), &rules)
		if err != nil {
			tLog.Warnf("[logic.GetModelInfoByID] parsed rule err, detail: %s", err.Error())
			return nil, err
		}
		res.Rules = rules
	}
	return res, nil
}

// DeleteModelInfoByID 根据id获取标签处理规则
func (ml *modelLib) DeleteModelInfoByID(ctx context.Context, id string, userInfo *drivenadapters.UserInfo) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)

	conditions, err := ml.getQueryConditionByModelType(ctx, id, userInfo.UserID)
	if err != nil {
		tLog.Warnf("[logic.DeleteModelInfoByID] getQueryConditionByModelType err, detail: %s", err.Error())
		return err
	}

	info, err := ml.dependency.GetModelInfoByID(ctx, conditions)
	if err != nil {
		tLog.Warnf("[logic.DeleteModelInfoByID] GetModelInfoByID err, detail: %s", err.Error())
		return err
	}

	err = ml.dependency.DeleteModelInfoByID(ctx, conditions)
	if err != nil {
		tLog.Warnf("[logic.DeleteModelInfoByID] DeleteModelInfoByID err, detail: %s", err.Error())
		return err
	}

	go func() {
		detail, extMsg := common.GetLogBody("deleteCustomCapabily", []interface{}{info.Name},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		ml.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: id,
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_WARN,
		}, &drivenadapters.JSONLogWriter{SendFunc: ml.executeMethods.Publish})
	}()

	return nil
}

// ListModelInfo 列举标签规则
func (ml *modelLib) ListModelInfo(ctx context.Context, userid string, status, offset, limit int64) ([]ModelInfo, error) {
	var (
		err        error
		modelInfos = make([]ModelInfo, 0)
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)
	listParams := &dependency.ListParams{
		UserID: &userid,
	}
	if status != -1 {
		listParams.Status = &status
	}
	models, err := ml.dependency.ListModelInfo(ctx, listParams, offset*limit, limit)
	if err != nil {
		tLog.Warnf("[logic.ListModelInfo] ListModelInfo err, detail: %s", err.Error())
		return nil, err
	}

	for _, val := range *models {
		modelInfos = append(modelInfos, ModelInfo{
			ID:        fmt.Sprintf("%v", val.ID),
			Name:      val.Name,
			CreatedAt: val.CreatedAt,
			UpdatedAt: val.UpdatedAt,
			Status:    int64(val.Status),
			Type:      int64(val.Type),
		})
	}

	return modelInfos, nil
}

// UpdateModelInfo 创建标签提取规则
func (ml *modelLib) UpdateModelInfo(ctx context.Context, id string, userInfo *drivenadapters.UserInfo, params *TagRulesMol) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)

	queryCondition, err := ml.getQueryConditionByModelType(ctx, id, userInfo.UserID)
	if err != nil {
		tLog.Warnf("[logic.DeleteModelInfoByID] getQueryConditionByModelType err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// 判断模型信息是否存在
	info, err := ml.dependency.GetModelInfoByID(ctx, queryCondition)
	if err != nil {
		tLog.Warnf("[logic.UpdateModelInfo] GetModelInfoByID err, detail: %s", err.Error())
		return err
	}

	conditions := &dependency.UpdateCondition{
		ID: &id,
	}
	if info.Type == common.TagRuleType {
		conditions.UserID = &userInfo.UserID
	}

	data := &dependency.UpdateParams{
		Name:        params.Name,
		Status:      params.Status,
		Description: params.Description,
	}

	// 只有标签提取时才可能更新rule
	if params.Rules != nil && info.Type == common.TagRuleType {
		r, _ := json.Marshal(params.Rules)
		strRule := string(r)
		data.Rule = &strRule
	}

	// 模型训练失败禁止用户保存
	if params.Status != nil && info.Type == common.UIEType && *params.Status == common.Publish && info.TrainStatus != common.TrainStatusFinished {
		return ierrors.NewIError(ierrors.Forbidden, ierrors.ModelTrainFailed, map[string]interface{}{"status": info.TrainStatus})
	}

	// 重名检测, 名称全局唯一
	if data.Name != nil && info.Name != *data.Name {
		exist, err := ml.dependency.CheckDupName(ctx, *data.Name)
		if err != nil {
			tLog.Warnf("[logic.UpdateModelInfo] CheckDupName err, detail: %s", err.Error())
			return err
		}
		if exist {
			return ierrors.NewIError(ierrors.DuplicatedName, "", nil)
		}
	}
	err = ml.dependency.UpdateModelInfo(ctx, conditions, data)
	if err != nil {
		tLog.Warnf("[logic.UpdateModelInfo] UpdateModelInfo err, detail: %s", err.Error())
		return err
	}

	var detail, extMsg string
	if info.Type == common.UIEType && len(info.Name) == 0 {
		detail, extMsg = common.GetLogBody("createCustomCapabily", []interface{}{*params.Name}, []interface{}{})
	} else {
		detail, extMsg = common.GetLogBody("updateCustomCapabily", []interface{}{*params.Name}, []interface{}{})
	}

	go func() {
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		ml.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: id,
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_WARN,
		}, &drivenadapters.JSONLogWriter{SendFunc: ml.executeMethods.Publish})
	}()

	return err
}

func (ml *modelLib) ExtractTagsByRule(ctx context.Context, params *TagExtractionParams, userInfo *drivenadapters.UserInfo) ([]string, error) {
	var (
		err  error
		tags = []string{}
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)
	tags, err = ml.dependency.ExtractTagsByRule(ctx, params, userInfo, extractFileSizeLimit)
	if err != nil {
		tLog.Warnf("[logic.ExtractTagsByRule] ExtractTagsByRule err, detail: %s", err.Error())
	}

	return tags, err
}

func (ml *modelLib) getQueryConditionByModelType(ctx context.Context, id, userid string) (*rds.QueryCondition, error) {
	taskType, err := ml.dependency.GetModelTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	conditions := &rds.QueryCondition{
		ID: &id,
	}
	if taskType == common.TagRuleType {
		conditions.UserID = &userid
	}

	return conditions, nil
}
