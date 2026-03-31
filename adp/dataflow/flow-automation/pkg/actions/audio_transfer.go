package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

// AudioTransfer 音频转文字节点
type AudioTransfer struct {
	DocID   string `json:"docid"`
	Version string `json:"version"`
}

// Name 操作名称
func (a *AudioTransfer) Name() string {
	return common.AudioTransfer
}

// Run 操作方法
func (a *AudioTransfer) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)

	input := params.(*AudioTransfer)

	// 创建执行器包装器
	executor := &audioTransferExecutor{
		action:  a,
		input:   input,
		token:   token,
		context: ctx,
	}

	// 使用通用的异步任务缓存管理器
	manager := NewAsyncTaskManager(ctx.NewExecuteMethods()).
		WithLockPrefix("automation:audio_transfer")

	return manager.Run(ctx, executor)
}

func (a *AudioTransfer) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	manager := NewAsyncTaskManager(ctx.NewExecuteMethods())
	return manager.RunAfter(ctx)
}

// ParameterNew 初始化参数
func (a *AudioTransfer) ParameterNew() interface{} {
	return &AudioTransfer{}
}

// audioTransferExecutor 实现 AsyncTaskExecutor 接口
type audioTransferExecutor struct {
	action  *AudioTransfer
	input   *AudioTransfer
	token   *entity.Token
	context entity.ExecuteContext
}

func (e *audioTransferExecutor) GetTaskType() string {
	return e.action.Name()
}

func (e *audioTransferExecutor) GetHashContent() string {
	return fmt.Sprintf("%s:%s:%s", e.action.Name(), e.input.DocID, e.input.Version)
}

func (e *audioTransferExecutor) GetExpireSeconds() int64 {
	config := common.NewConfig()
	return config.ActionConfig.AudioTransfer.ExpireSec
}

func (e *audioTransferExecutor) GetResultFileExt() string {
	return ".json"
}

func (e *audioTransferExecutor) Execute(ctx context.Context) (map[string]any, error) {
	log := traceLog.WithContext(ctx)

	downloadInfo, err := GetFileDownloadInfoWithDocAttr(ctx, e.context, e.input.DocID, e.input.Version)
	if err != nil {
		log.Warnf("[audioTransferExecutor] GetFileDownloadInfoWithDocAttr failed, detail: %s", err.Error())
		return nil, err
	}

	// 获取文件二进制内容
	body, err := getFileStream(downloadInfo.URL, 15*60*time.Second)
	if err != nil {
		log.Warnf("[audioTransferExecutor] getFileStream failed, detail: %s", err.Error())
		return nil, err
	}

	var sizeLimit int64 = 500 * 1024 * 1024
	result, err := e.context.NewRepo().AudioTransfer(ctx, float64(sizeLimit), "", body, downloadInfo.DocAttr)
	if err != nil {
		log.Warnf("[audioTransferExecutor] AudioTransfer failed, detail: %s", err.Error())
		return nil, err
	}

	return result, nil
}
