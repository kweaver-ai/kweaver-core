package dependency

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

// SpeechModel 音频转文字接口
type SpeechModel interface {
	// AudioTransfer 音频转文字
	AudioTransfer(ctx context.Context, sizeLimit float64, wbbHook string, content *[]byte, attr *drivenadapters.DocAttr) (map[string]interface{}, error)
	// CheckSpeechModel 检查SpeechModel是否可用
	CheckSpeechModel(ctx context.Context) error
}

var audioSupportMap = map[string]bool{
	".mp3": true,
	".wav": true,
	".m4a": true,
	".mp4": true,
}

type speechModel struct {
	speechModel drivenadapters.SpeechModel
}

var (
	sOnce sync.Once
	sm    SpeechModel
)

// NewSpeechModel 实例化音频转文字实例
func NewSpeechModel() SpeechModel {
	sOnce.Do(func() {
		sm = &speechModel{
			speechModel: drivenadapters.NewSpeechModel(),
		}
	})
	return sm
}

// AudioTransfer 音频转文字
func (c *speechModel) AudioTransfer(ctx context.Context, sizeLimit float64, wbbHook string, content *[]byte, attr *drivenadapters.DocAttr) (map[string]interface{}, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var result = map[string]interface{}{}

	docName := attr.Name
	fileExtension := utils.GetFileExtension(docName)

	if !audioSupportMap[fileExtension] {
		return nil, errors.NewIError(errors.FileTypeNotSupported, "", map[string]interface{}{
			"doc": map[string]interface{}{
				"docid":       attr.DocID,
				"supportType": "mp3/wav/m4a",
				"docname":     attr.Name,
			},
		})
	}

	if attr.Size > sizeLimit {
		return result, errors.NewIError(errors.FileSizeExceed, "", map[string]interface{}{
			"doc": map[string]interface{}{
				"docid": attr.DocID,
				"limit": sizeLimit,
			},
		})
	}

	speechModelAdapter := drivenadapters.NewSpeechModel()
	taskID, err := speechModelAdapter.AudioTransfer(ctx, attr.Name, wbbHook, content)
	if err != nil {
		log.Warnf("[AudioTransfer] AudioTransfer failed, taskID %v, err: %v", "applyID", err.Error())
		return nil, err
	}
	for {
		code, res, err := speechModelAdapter.GetAudioTransferResult(ctx, taskID)
		if err != nil {
			log.Warnf("[AudioTransfer] GetAudioTransferResult failed, detail: %s", err.Error())
			return nil, err
		}
		if code == http.StatusAccepted {
			time.Sleep(10 * time.Second)
			continue
		}
		return res, nil
	}
}

// CheckSpeechModel 检查CheckSpeechModel是否可用
func (c *speechModel) CheckSpeechModel(ctx context.Context) error {
	return c.speechModel.CheckSpeechModel(ctx)
}
