package dependency

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

const (
	GET_SUBDOC_TYPE = "full_text"
	TRAIN_FILES_DIR = "/trainfiles"
)

var (
	SUPPORT_FILE_EXT = []string{"doc", "docx", "pptx", "ppt", "pdf", "txt"}
)

// FileMetaData 训练文件元数据信息
type FileMetaData struct {
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
	Schema   string `json:"schema"`
	Lines    int    `json:"lines"`
}

type TagRules struct {
	ID   string        `json:"tag_id"`
	Path string        `json:"tag_path"`
	Rule TagRuleDetail `json:"rule"`
}

type TagRuleDetail struct {
	OR [][]string `json:"or"`
}

type TagExtractionParams struct {
	Target map[string]string `json:"target"`
	RuleID string            `json:"rule_id"`
	Rules  []TagRules        `json:"rules"`
}

type AiModelService interface {
	// CreateTagsRule 创建标签提取规则
	CreateTagsRule(ctx context.Context, name, userid, rule, description string, status int) (uint64, error)
	// GetModelInfoByID 根据id获取标签处理规则
	GetModelInfoByID(ctx context.Context, conditions *QueryCondition) (*rds.AiModel, error)
	// DeleteModelInfoByID 根据id删除标签处理规则
	DeleteModelInfoByID(ctx context.Context, conditions *QueryCondition) error
	// ListModelInfo 列举标签处理规则
	ListModelInfo(ctx context.Context, param *rds.ListParams, offset, limit int64) (*[]rds.AiModel, error)
	// UpdateModelInfo 更新标签规则
	UpdateModelInfo(ctx context.Context, conditions *UpdateCondition, data *UpdateParams) error
	// ExtractTagsByRule 提取标签
	ExtractTagsByRule(ctx context.Context, params *TagExtractionParams, userInfo *drivenadapters.UserInfo, sizeLimit int64) ([]string, error)
	// CreateTrainFile 创建训练文件存储oss信息
	CreateTrainFile(ctx context.Context, data *rds.AiModel, trainFile *rds.TrainFileOSSInfo) error
	// StartTrainModule 模型训练
	StartTrainModule(ctx context.Context, trainID string, userID string) (string, error)
	// StartInfer 模型预测
	StartInfer(ctx context.Context, trainID, content, token, loginIP string, sizeLimit int64) (map[string]interface{}, error)
	// VerifyTaskUnique 校验任务唯一性
	VerifyTaskUnique(ctx context.Context) (bool, error)
	// CheckDupName 校验名称是否重复
	CheckDupName(ctx context.Context, name string) (bool, error)
	// GetModelTypeByID 根据id获取模型类型
	GetModelTypeByID(ctx context.Context, id string) (int, error)
}

type aiModel struct {
	uieHandle  drivenadapters.Uie
	oss        drivenadapters.OssGateWay
	codeRunner drivenadapters.CodeRunner
	efast      drivenadapters.Efast
	ecotag     drivenadapters.EcoTag
	aiModel    rds.AiModelDao
	cache      *cache.Cache
}

var (
	tOnce sync.Once
	t     AiModelService
)

type ListParams = rds.ListParams
type UpdateCondition = rds.UpdateCondition
type QueryCondition = rds.QueryCondition
type UpdateParams = rds.UpdateParams

// NewAiModelService 实例化Ai能力实例
func NewAiModelService() AiModelService {
	tOnce.Do(func() {
		t = &aiModel{
			uieHandle:  drivenadapters.NewUie(),
			oss:        drivenadapters.NewOssGateWay(),
			codeRunner: drivenadapters.NewCodeRunner(),
			efast:      drivenadapters.NewEfast(),
			ecotag:     drivenadapters.NewEcoTag(),
			aiModel:    rds.GetAiModelDao(),
			cache:      cache.New(3*time.Minute, 6*time.Minute),
		}
	})
	return t
}

// CreateTagsRule 创建标签处理规则
func (ai *aiModel) CreateTagsRule(ctx context.Context, name, userid, rule, description string, status int) (uint64, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)
	id, _ := utils.GetUniqueID()
	data := &rds.AiModel{
		ID:          id,
		Name:        name,
		Rule:        rule,
		UserID:      userid,
		Description: description,
		Status:      status,
		Type:        common.TagRuleType,
		CreatedAt:   time.Now().UnixMicro() / 1000,
		UpdatedAt:   time.Now().UnixMicro() / 1000,
	}
	err = ai.aiModel.CreateTagsRule(ctx, data)
	if err != nil {
		tLog.Warnf("[dependency.CreateTagsRule] CreateTagsRule err, detail: %s", err.Error())
	}
	return id, err
}

// GetModelInfoByID 根据id和用户id获取模型信息
func (ai *aiModel) GetModelInfoByID(ctx context.Context, conditions *QueryCondition) (*rds.AiModel, error) {
	aiModel, err := ai.aiModel.GetModelInfoByID(ctx, conditions)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[dependency.GetModelInfoByID] GetModelInfoByID err, detail: %s", err.Error())
		if err == gorm.ErrRecordNotFound {
			return &aiModel, ierrors.NewIError(ierrors.TaskNotFound, "", err.Error())
		}
		return &aiModel, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	return &aiModel, nil
}

// DeleteModelInfoByID 根据id删除模型信息
func (ai *aiModel) DeleteModelInfoByID(ctx context.Context, conditions *QueryCondition) error {
	var (
		err error
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)
	err = ai.aiModel.DeleteModelInfoByID(ctx, conditions)
	if err != nil {
		tLog.Warnf("[dependency.DeleteModelInfoByID] DeleteModelInfoByID err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	return err
}

// ListModelInfo 列举模型信息
func (ai *aiModel) ListModelInfo(ctx context.Context, param *rds.ListParams, offset, limit int64) (*[]rds.AiModel, error) {
	models, err := ai.aiModel.ListModelInfo(ctx, param, offset, limit)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[dependency.ListModelInfo] ListModelInfo err, detail: %s", err.Error())
		return &models, ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	return &models, err
}

// UpdateModelInfo 更新模型信息
func (ai *aiModel) UpdateModelInfo(ctx context.Context, conditions *UpdateCondition, data *UpdateParams) error {
	err := ai.aiModel.UpdateModelInfo(ctx, conditions, data)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[dependency.UpdateModelInfo] UpdateModelInfo err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	return nil
}

func (ai *aiModel) ExtractTagsByID(ctx context.Context, docID, content, userID, token, loginIP string, ruleID uint64, sizeLimit int64) ([]string, error) {
	var (
		err  error
		tags = []string{}
		con  = content
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)
	if docID != "" {
		con, err = ai.getFileContent(ctx, tLog, docID, token, loginIP, SUPPORT_FILE_EXT, sizeLimit)
	}

	ruleIDstr := fmt.Sprintf("%v", ruleID)
	rule, err := ai.aiModel.GetModelInfoByID(ctx, &QueryCondition{
		ID:     &ruleIDstr,
		UserID: &userID,
	})
	if err != nil {
		tLog.Warnf("[dependency.GetModelInfoByID] ExtractTags err, detail: %s", err.Error())
	}
	tags, err = ai.codeRunner.ExtractTags(ctx, con, rule)
	if err != nil {
		tLog.Warnf("[dependency.ExtractTagsByID] ExtractTags err, detail: %s", err.Error())
	}
	return tags, err
}

func (ai *aiModel) ExtractTagsByRule(ctx context.Context, params *TagExtractionParams, userInfo *drivenadapters.UserInfo, sizeLimit int64) ([]string, error) {
	var (
		err    error
		tags   = []string{}
		con    = params.Target["content"]
		docID  = params.Target["docid"]
		rules  = params.Rules
		ruleID = params.RuleID
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)
	if docID != "" {
		con, err = ai.getFileContent(ctx, tLog, docID, userInfo.TokenID, userInfo.LoginIP, SUPPORT_FILE_EXT, sizeLimit)
		if err != nil {
			tLog.Warnf("[dependency.ExtractTagsByRule] getFileContent err, detail: %s", err.Error())
			return tags, err
		}
	}
	if ruleID != "" {
		// 此处不应该按照用户id查询模型信息，工作流触发不一定是模型创建者
		rule, err := ai.aiModel.GetModelInfoByID(ctx, &QueryCondition{
			ID: &ruleID,
			// UserID: &userInfo.UserID,
		})
		if err != nil {
			tLog.Warnf("[dependency.ExtractTagsByRule] GetModelInfoByID err, detail: %s", err.Error())
			return tags, err
		}
		fmt.Println(rule.Rule)
		err = json.Unmarshal([]byte(rule.Rule), &rules)
		if err != nil {
			tLog.Warnf("[dependency.ExtractTagsByRule] parsed rules err, detail: %s", err.Error())
			return tags, err
		}
		ai.parseTags(ctx, &rules)
	}

	tags, err = ai.codeRunner.ExtractTags(ctx, con, rules)
	if err != nil {
		tLog.Warnf("[dependency.ExtractTagsByRule] ExtractTags err, detail: %s", err.Error())
	}
	return tags, err
}

func (ai *aiModel) getFileContent(ctx context.Context, tLog traceLog.Logger, docID, token, loginIP string, supportExtension []string, sizeLimit int64) (string, error) {
	perm, err := ai.efast.CheckPerm(ctx, docID, "download", token, loginIP)
	if err != nil {
		tLog.Warnf("[dependency.getFileContent] CheckPerm err, detail: %s", err.Error())
		return "", err
	}
	if perm != 0 {
		err = ierrors.NewIError(ierrors.NoPermission, "", map[string]interface{}{
			"info": "has no perm to download doc",
			"doc": map[string]string{
				"docid": docID,
			},
		})
		return "", err
	}
	attr, err := ai.efast.GetDocMsg(ctx, docID)
	if err != nil {
		tLog.Warnf("[dependency.getFileContent] GetDocMsg err, detail: %s", err.Error())
		return "", err
	}

	fileExtension := utils.GetFileExtension(attr.Name)
	if !utils.IsContain(strings.TrimPrefix(fileExtension, "."), supportExtension) {
		return "", ierrors.NewIError(ierrors.FileTypeNotSupported, "", map[string]interface{}{
			"doc": map[string]interface{}{
				"docid":       attr.DocID,
				"supportType": strings.Join(supportExtension, "/"),
				"docname":     attr.Name,
			},
		})
	}

	content, err := ai.efast.GetDocSetSubdocContent(ctx, drivenadapters.DocSetSubdocParams{
		DocID: docID,
		Type:  GET_SUBDOC_TYPE,
	}, 3, 15*time.Second, float64(sizeLimit))
	if err != nil {
		return "", err
	}
	return content, nil
}

func (ai *aiModel) parseTags(ctx context.Context, rules *[]TagRules) (*[]TagRules, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)
	if data, found := ai.cache.Get("tagMap"); found {
		r, ok := data.([]TagRules)
		if ok {
			return &r, nil
		}
	}
	tagTree, err := ai.ecotag.GetTagTrees(ctx)
	if err != nil {
		tLog.Warnf("[dependency.parseTags] GetTagTrees err, detail: %s", err.Error())
		return rules, err
	}

	// 创建一个映射以优化查找性能
	idToPath := make(map[string]string)
	var buildMap func(tags []*drivenadapters.TagTree)
	buildMap = func(tags []*drivenadapters.TagTree) {
		for _, tag := range tags {
			idToPath[tag.ID] = tag.Path
			if len(tag.ChildTags) > 0 {
				buildMap(tag.ChildTags)
			}
		}
	}
	buildMap(tagTree)

	for i, _ := range *rules {
		r := *rules
		r[i].Path = idToPath[r[i].ID]
	}
	ai.cache.Set("tagMap", *rules, cache.DefaultExpiration)
	return rules, nil
}

// CreateTrainFile 创建训练文件存储oss信息
func (ai *aiModel) CreateTrainFile(ctx context.Context, data *rds.AiModel, trainFile *rds.TrainFileOSSInfo) error {
	err := ai.aiModel.CreateTrainFile(ctx, data, trainFile)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[dependency.CreateTrainFile] CreateTrainFile err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	return nil
}

// VerifyTaskUnique 校验任务唯一性
func (ai *aiModel) VerifyTaskUnique(ctx context.Context) (bool, error) {
	exist, err := ai.aiModel.VerifyTaskUnique(ctx)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[dependency.VerifyTaskUnique] VerifyTaskUnique err, detail: %s", err.Error())
		return exist, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	return exist, nil
}

// CheckDupName 校验名称是否重复
func (ai *aiModel) CheckDupName(ctx context.Context, name string) (bool, error) {
	exist, err := ai.aiModel.CheckDupName(ctx, name)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[dependency.CheckDupName] CheckDupName err, detail: %s", err.Error())
		return exist, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	return exist, nil
}

// GetModelTypeByID 根据id获取模型类型
func (ai *aiModel) GetModelTypeByID(ctx context.Context, id string) (int, error) {
	taskType, err := ai.aiModel.GetModelTypeByID(ctx, id)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[dependency.GetModelTypeByID] GetModelTypeByID err, detail: %s", err.Error())
		return taskType, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	return taskType, nil
}

// StartTrainModule 模型训练
func (ai *aiModel) StartTrainModule(ctx context.Context, trainID string, userID string) (string, error) {
	var (
		taskID int
		err    error
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)

	fileInfo, err := ai.aiModel.GetTrainFileInfo(ctx, trainID)
	if err != nil {
		tLog.Warnf("[dependency.StartTrainModule] GetTrainFileInfo err, detail: %s", err.Error())
		if err == gorm.ErrRecordNotFound {
			return "", ierrors.NewIError(ierrors.TaskNotFound, "", err.Error())
		}
		return "", ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	parent, err := utils.CreateFolder(fmt.Sprintf("%s/%s", TRAIN_FILES_DIR, fileInfo.Key))
	if err != nil {
		return "", err
	}
	// 结束自定删除本地文件
	defer os.RemoveAll(parent)

	path := fmt.Sprintf("%s/%s.tar.gz", parent, fileInfo.Key)
	_, err = ai.oss.DownloadFile2Local(ctx, fileInfo.OSSID, fileInfo.Key, true, path)
	if err != nil {
		tLog.Warnf("[dependency.StartTrainModule] DownloadFile2Local err, detail: %s", err.Error())
		return "", err
	}

	err = common.Decompress(path, parent)
	if err != nil {
		return "", err
	}

	metadataPath := fmt.Sprintf("%s/metadata.txt", parent)
	buf, err := utils.ReadFile(metadataPath)
	if err != nil {
		return "", err
	}
	var metadata FileMetaData
	_ = json.Unmarshal(buf, &metadata)

	trainFilePath := fmt.Sprintf("%s/%s", parent, metadata.FileName)
	contentByte, err := utils.ReadFile(trainFilePath)
	if err != nil {
		return "", err
	}
	contentByte = bytes.TrimSpace(contentByte)

	taskID, err = ai.uieHandle.StartTrainModule(ctx, &contentByte)
	if err != nil {
		tLog.Warnf("[dependency.StartTrainModule] StartTrainModule err, detail: %s", err.Error())
		return "", err
	}

	if taskID == -1 {
		return "", ierrors.NewIError(ierrors.Forbidden, ierrors.TrainingInProgress, nil)
	}

	TrainIDUint64, _ := strconv.ParseUint(trainID, 10, 64)
	currentTime := time.Now().UnixNano() / 1e6
	data := &rds.AiModel{
		ID:        TrainIDUint64,
		UpdatedAt: currentTime,
		Rule:      metadata.Schema,
		UserID:    userID,
	}

	ai.reCycleCheckTrainStatus(ctx, data, taskID)

	if data.TrainStatus == common.TrainStatusFailed {
		return data.TrainStatus, ierrors.NewIError(ierrors.InternalError, ierrors.ModelTrainFailed, "")
	}

	err = ai.aiModel.UpdateTrainLog(ctx, data)
	if err != nil {
		tLog.Warnf("[dependency.StartTrainModule] UpdateTrainLog err, detail: %s", err.Error())
		return "", ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	return data.TrainStatus, nil
}

func (ai *aiModel) reCycleCheckTrainStatus(ctx context.Context, data *rds.AiModel, taskID int) {
	var (
		retryTime       int
		err             error
		trainRecordFlag bool
		trainLogs       []drivenadapters.TrainLog
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)

L:
	for {
		trainLogs, err = ai.uieHandle.ListTrainLog(ctx)
		if err != nil {
			tLog.Warnf("[dependency.recycleCheckTrainStatus] ListTrainLog err, detail: %s, retry:", err.Error(), retryTime)
			if retryTime < 3 {
				retryTime++
				time.Sleep(15 * time.Second)
				continue
			} else {
				data.TrainStatus = common.TrainStatusFailed
				return
			}
		} else {
			for i := len(trainLogs) - 1; i >= 0; i-- {
				trainLog := trainLogs[i]
				if trainLog.ID == taskID {
					if trainLog.Status == "running" {
						time.Sleep(3 * time.Second)
						goto L
					}
					data.TrainStatus = trainLog.Status
					trainRecordFlag = true
					break
				}
			}
			break
		}
	}

	// 如果训练状态为空，此次日志记录不添加至数据库
	if !trainRecordFlag {
		data.TrainStatus = common.TrainStatusFailed
	}
}

// StartInfer 模型预测
func (ai *aiModel) StartInfer(ctx context.Context, trainID, content, token, loginIP string, sizeLimit int64) (map[string]interface{}, error) {
	var (
		err error
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tLog := traceLog.WithContext(ctx)

	// 如果文件以gns:// 开头则表示为文档中心文件
	if strings.HasPrefix(content, "gns://") {
		content, err = ai.getFileContent(ctx, tLog, content, token, loginIP, SUPPORT_FILE_EXT, sizeLimit)
	}

	if err != nil {
		tLog.Warnf("[dependency.StartInfer] getFileContent err, detail: %s", err.Error())
		return nil, err
	}

	schema, err := ai.getInferSchema(ctx, trainID)
	if err != nil {
		tLog.Warnf("[dependency.EntityExtract] GetInferSchema err, detail: %s", err.Error())
		return nil, err
	}

	datas, err := ai.uieHandle.StartInfer(ctx, schema, content)
	if err != nil {
		tLog.Warnf("[dependency.EntityExtract] StartInfer err, detail: %s", err.Error())
		return datas, err
	}

	if schema == nil {
		return datas, nil
	}

	schemaArr, _ := schema.([]string)
	items := datas["res"].([]interface{})
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		for _, val := range schemaArr {
			if _, ok := itemMap[val]; ok {
				continue
			}
			itemMap[val] = make([]map[string]interface{}, 0)
		}
	}

	return datas, nil
}

// GetInferSchema 获取模型推理schema
func (ai *aiModel) getInferSchema(ctx context.Context, trainID string) (interface{}, error) {
	var err error
	tLog := traceLog.WithContext(ctx)

	schemaStr, err := ai.aiModel.GetInferSchema(ctx, trainID)
	if err != nil {
		tLog.Warnf("[dependency.EntityExtract] StartInfer err, detail: %s", err.Error())
		return nil, err
	}
	if len(schemaStr) == 0 {
		return nil, nil
	}
	// 当前仅处理简单实体schema,后续事件抽取可在此扩展
	schemas := strings.Split(schemaStr, ",")

	return schemas, nil
}
