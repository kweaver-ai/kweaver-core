package dependency

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
)

// 依赖倒置的接口
// Repo 被驱动器代理
type Repo interface {
	// CheckSpeechModel 检查SpeechModel服务是否可用
	CheckSpeechModel(ctx context.Context) error
	// AudioTransfer 音频转文字
	AudioTransfer(ctx context.Context, sizeLimit float64, wbbHook string, content *[]byte, attr *drivenadapters.DocAttr) (map[string]interface{}, error)
	// RecognizeText ocr文字识别
	RecognizeText(ctx context.Context, params map[string]interface{}, attr *drivenadapters.DocAttr) (map[string]interface{}, error)
	// CreateTrainFile 创建训练文件存储oss信息
	CreateTrainFile(ctx context.Context, data *rds.AiModel, trainFile *rds.TrainFileOSSInfo) error
	// StartTrainModule 模型训练
	StartTrainModule(ctx context.Context, trainID string, userID string) (string, error)
	// StartInfer 模型预测
	StartInfer(ctx context.Context, trainID, content, token, loginIP string, sizeLimit int64) (map[string]interface{}, error)
	// CreateTagsRule 创建标签提取规则
	CreateTagsRule(ctx context.Context, name, userid, rule, description string, status int) (uint64, error)
	// GetModelInfoByID 根据id获取标签处理规则
	GetModelInfoByID(ctx context.Context, conditions *QueryCondition) (*rds.AiModel, error)
	// GetModelInfoByID 根据id获取标签处理规则
	DeleteModelInfoByID(ctx context.Context, conditions *QueryCondition) error
	// ListModelInfo 列举标签处理规则
	ListModelInfo(ctx context.Context, param *ListParams, offset, limit int64) (*[]rds.AiModel, error)
	// UpdateModelInfo 更新标签规则
	UpdateModelInfo(ctx context.Context, conditions *UpdateCondition, data *UpdateParams) error
	// // ExtractTagsByID 提取标签
	// ExtractTagsByID(ctx context.Context, docID, content, userID, token, loginIP string, ruleID uint64) ([]string, error)
	// ExtractTagsByRule 提取标签
	ExtractTagsByRule(ctx context.Context, params *TagExtractionParams, userInfo *drivenadapters.UserInfo, sizeLimit int64) ([]string, error)
	// VerifyTaskUnique 校验任务唯一性
	VerifyTaskUnique(ctx context.Context) (bool, error)
	// CheckDupName 检查重名
	CheckDupName(ctx context.Context, name string) (bool, error)
	// GetModelTypeByID 根据id获取模型类型
	GetModelTypeByID(ctx context.Context, id string) (int, error)
	// GetDatabaseTableList 获取数据库表列表
	GetDatabaseTableList(ctx context.Context, dataSourceID, token, ipStr string) ([]TableMetadata, error)
	// ExtractFullText 提取文本
	ExtractFullText(ctx context.Context, docID string) (map[string]any, error)
	// ConvertToPDF 转换为pdf
	ConvertToPDF(ctx context.Context, taskID, docID string) error
	// HandleGotenbergCallback 处理Gotenberg回调
	HandleGotenbergCallback(ctx context.Context, req *GotenbergCallbackRequest) (map[string]any, error)

	DocumentConverter() DocumentConverter
}

// repo 委托对象
type repo struct {
	codeRunner           CodeRunner
	speechModel          SpeechModel
	aiModel              AiModelService
	databaseTableService DatabaseTableService
	documentConverter    DocumentConverter
}

// NewDriven driven 被驱动对象
func NewDriven() Repo {
	ur := &repo{
		codeRunner:           NewCoderunner(),
		speechModel:          NewSpeechModel(),
		aiModel:              NewAiModelService(),
		databaseTableService: NewDatabaseTableService(),
		documentConverter:    NewDocumentConverter(),
	}
	return ur
}

// CheckSpeechModel 检查SpeechModel是否可用
func (r *repo) CheckSpeechModel(ctx context.Context) error {
	return r.speechModel.CheckSpeechModel(ctx)
}

// AudioTransfer 音频转文字
func (r *repo) AudioTransfer(ctx context.Context, sizeLimit float64, wbbHook string, content *[]byte, attr *drivenadapters.DocAttr) (map[string]interface{}, error) {
	return r.speechModel.AudioTransfer(ctx, sizeLimit, wbbHook, content, attr)
}

// RecognizeText ocr文字识别
func (r *repo) RecognizeText(ctx context.Context, params map[string]interface{}, attr *drivenadapters.DocAttr) (map[string]interface{}, error) {
	return r.codeRunner.RecognizeText(ctx, params, attr)
}

// StartTrainModule 模型训练
func (r *repo) StartTrainModule(ctx context.Context, trainID string, userID string) (string, error) {
	return r.aiModel.StartTrainModule(ctx, trainID, userID)
}

// StartInfer 模型预测
func (r *repo) StartInfer(ctx context.Context, trainID, content, token, loginIP string, sizeLimit int64) (map[string]interface{}, error) {
	return r.aiModel.StartInfer(ctx, trainID, content, token, loginIP, sizeLimit)
}

// CreateTrainFile 上传模型训练文件
func (r *repo) CreateTrainFile(ctx context.Context, data *rds.AiModel, trainFile *rds.TrainFileOSSInfo) error {
	return r.aiModel.CreateTrainFile(ctx, data, trainFile)
}

// CreateTagsRule 创建标签提取规则
func (r *repo) CreateTagsRule(ctx context.Context, name, userid, rule, description string, status int) (uint64, error) {
	return r.aiModel.CreateTagsRule(ctx, name, userid, rule, description, status)
}

// GetModelInfoByID 根据id获取标签处理规则
func (r *repo) GetModelInfoByID(ctx context.Context, conditions *QueryCondition) (*rds.AiModel, error) {
	return r.aiModel.GetModelInfoByID(ctx, conditions)
}

// DeleteModelInfoByID 根据id获取标签处理规则
func (r *repo) DeleteModelInfoByID(ctx context.Context, conditions *QueryCondition) error {
	return r.aiModel.DeleteModelInfoByID(ctx, conditions)
}

// ListModelInfo 列举标签处理规则
func (r *repo) ListModelInfo(ctx context.Context, param *ListParams, offset, limit int64) (*[]rds.AiModel, error) {
	return r.aiModel.ListModelInfo(ctx, param, offset, limit)
}

// UpdateModelInfo 更新标签规则
func (r *repo) UpdateModelInfo(ctx context.Context, conditions *UpdateCondition, data *UpdateParams) error {
	return r.aiModel.UpdateModelInfo(ctx, conditions, data)
}

// // ExtractTagsByID 提取标签
// func (r *repo) ExtractTagsByID(ctx context.Context, docid, content, userID, token, loginIP string, ruleID uint64) ([]string, error) {
// 	return r.aiModel.ExtractTagsByID(ctx, docid, content, userID, token, loginIP, ruleID)
// }

// ExtractTagsByRule 提取标签
func (r *repo) ExtractTagsByRule(ctx context.Context, params *TagExtractionParams, userInfo *drivenadapters.UserInfo, sizeLimit int64) ([]string, error) {
	return r.aiModel.ExtractTagsByRule(ctx, params, userInfo, sizeLimit)
}

// VerifyTaskUnique 校验任务唯一性
func (r *repo) VerifyTaskUnique(ctx context.Context) (bool, error) {
	return r.aiModel.VerifyTaskUnique(ctx)
}

// CheckDupName 检查重名
func (r *repo) CheckDupName(ctx context.Context, name string) (bool, error) {
	return r.aiModel.CheckDupName(ctx, name)
}

// GetModelTypeByID 根据id获取模型类型
func (r *repo) GetModelTypeByID(ctx context.Context, id string) (int, error) {
	return r.aiModel.GetModelTypeByID(ctx, id)
}

// GetDatabaseTableList 获取数据库表列表
func (r *repo) GetDatabaseTableList(ctx context.Context, dataSourceID, token, ipStr string) ([]TableMetadata, error) {
	return r.databaseTableService.ListTables(ctx, dataSourceID, token, ipStr)
}

// ExtractFullText 全文提取
func (r *repo) ExtractFullText(ctx context.Context, docID string) (map[string]any, error) {
	return r.documentConverter.ExtractFullText(ctx, docID)
}

// ConvertToPDF 转换为pdf
func (r *repo) ConvertToPDF(ctx context.Context, taskID, docID string) error {
	return r.documentConverter.ConvertToPDF(ctx, taskID, docID)
}

// HandleGotenbergCallback 处理gotenberg回调
func (r *repo) HandleGotenbergCallback(ctx context.Context, req *GotenbergCallbackRequest) (map[string]any, error) {
	return r.documentConverter.HandleGotenbergCallback(ctx, req)
}

func (r *repo) DocumentConverter() DocumentConverter {
	return r.documentConverter
}
