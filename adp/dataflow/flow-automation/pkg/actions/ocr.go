package actions

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

const (
	ocr         = "@anyshare/ocr/general"
	eleInvoice  = "@anyshare/ocr/eleinvoice"
	idCard      = "@anyshare/ocr/idcard"
	ocrTaskType = int64(100)
	pdfTaskType = int64(1)
)

var (
	mu  sync.Mutex
	emu sync.Mutex
	imu sync.Mutex
)

// OCR 图像识别
type OCR struct {
	DocID string `json:"docid"`
}

// Name 操作名称
func (a *OCR) Name() string {
	return ocr
}

// Run 操作方法
func (a *OCR) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) { //nolint
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	newCtx = context.WithValue(newCtx, common.Authorization, token.Token)
	ctx.SetContext(newCtx)
	log := traceLog.WithContext(ctx.Context())

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*OCR)

	err = checkOCRAvailable(ctx, token)
	if err != nil {
		log.Warnf("[OCR.Run] checkOCRAvailable failed, detail: %s", err.Error())
		return nil, err
	}

	mu.Lock()
	defer mu.Unlock()
	result, err := handleOCR(ctx, token, input.DocID, "general", "handwriting", "/lab/ocr/predict/general", pdfTaskType)
	if err != nil {
		log.Warnf("[OCR.Run] handleOCR failed, detail: %s", err.Error())
	}
	return result, err
}

// ParameterNew 初始化参数
func (a *OCR) ParameterNew() interface{} {
	return &OCR{}
}

// EleInvoice 电子票据识别
type EleInvoice struct {
	DocID string `json:"docid"`
}

// Name 操作名称
func (ei *EleInvoice) Name() string {
	return eleInvoice
}

// Run 操作方法
func (ei *EleInvoice) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) { //nolint
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	newCtx = context.WithValue(newCtx, common.Authorization, token.Token)
	ctx.SetContext(newCtx)
	log := traceLog.WithContext(ctx.Context())

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*EleInvoice)

	err = checkOCRAvailable(ctx, token)
	if err != nil {
		log.Warnf("[EleInvoice.Run] checkOCRAvailable failed, detail: %s", err.Error())
		return nil, err
	}

	emu.Lock()
	defer emu.Unlock()
	result, err := handleOCR(ctx, token, input.DocID, "eleinvoice", "eleinvoice", "/lab/ocr/predict/ticket", ocrTaskType)
	if err != nil {
		log.Warnf("[EleInvoice.Run] handleOCR failed, detail: %s", err.Error())
	}
	return result, err
}

// ParameterNew 初始化参数
func (ei *EleInvoice) ParameterNew() interface{} {
	return &EleInvoice{}
}

// IDCard 身份证识别
type IDCard struct {
	DocID string `json:"docid"`
}

// Name 操作名称
func (ic *IDCard) Name() string {
	return idCard
}

// Run 操作方法
func (ic *IDCard) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) { //nolint
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	newCtx = context.WithValue(newCtx, common.Authorization, token.Token)
	ctx.SetContext(newCtx)
	log := traceLog.WithContext(ctx.Context())

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*IDCard)

	err = checkOCRAvailable(ctx, token)
	if err != nil {
		log.Warnf("[IDCard.Run] checkOCRAvailable failed, detail: %s", err.Error())
		return nil, err
	}

	imu.Lock()
	defer imu.Unlock()
	result, err := handleOCR(ctx, token, input.DocID, "idcard", "idcard", "/lab/ocr/predict/ticket", ocrTaskType)
	if err != nil {
		log.Warnf("[IDCard.Run] handleOCR failed, detail: %s", err.Error())
	}
	return result, err
}

// ParameterNew 初始化参数
func (ic *IDCard) ParameterNew() interface{} {
	return &IDCard{}
}

type OCRNew struct {
	DocID   string `json:"docid"`
	Version string `json:"version"`
}

func (a *OCRNew) Name() string {
	return common.OpOCRNew
}

func (a *OCRNew) ParameterNew() any {
	return &OCRNew{}
}

func (a *OCRNew) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)

	input := params.(*OCRNew)

	// 创建执行器包装器
	executor := &ocrNewExecutor{
		action:  a,
		input:   input,
		token:   token,
		context: ctx,
	}

	// 使用通用的异步任务缓存管理器
	manager := NewAsyncTaskManager(ctx.NewExecuteMethods()).
		WithLockPrefix("automation:ocr")

	return manager.Run(ctx, executor)
}

func (a *OCRNew) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	manager := NewAsyncTaskManager(ctx.NewExecuteMethods())
	return manager.RunAfter(ctx)
}

// ocrNewExecutor 实现 AsyncTaskExecutor 接口
type ocrNewExecutor struct {
	action  *OCRNew
	input   *OCRNew
	token   *entity.Token
	context entity.ExecuteContext
}

func (e *ocrNewExecutor) GetTaskType() string {
	return e.action.Name()
}

func (e *ocrNewExecutor) GetHashContent() string {
	return fmt.Sprintf("%s:%s:%s", e.action.Name(), e.input.DocID, e.input.Version)
}

func (e *ocrNewExecutor) GetExpireSeconds() int64 {
	config := common.NewConfig()
	return config.ActionConfig.OCR.ExpireSec
}

func (e *ocrNewExecutor) GetResultFileExt() string {
	return ".json"
}

func (e *ocrNewExecutor) Execute(ctx context.Context) (map[string]any, error) {
	log := traceLog.WithContext(ctx)

	downloadInfo, err := GetFileDownloadInfo(ctx, e.context, e.input.DocID, e.input.Version)
	if err != nil {
		log.Warnf("[ocrNewExecutor] GetFileDownloadInfo err %s, docid %s, version %s", err.Error(), e.input.DocID, e.input.Version)
		return nil, err
	}

	ocr := drivenadapters.NewOcr()
	text, err := ocr.RecognizeText(ctx, downloadInfo.URL, downloadInfo.Filename)

	if err != nil {
		log.Warnf("[ocrNewExecutor] RecognizeText err %s, docid %s, version %s", err.Error(), e.input.DocID, e.input.Version)
		return nil, err
	}

	result := map[string]any{
		"text": text,
	}

	return result, nil
}
