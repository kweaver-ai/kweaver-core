package dependency

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
)

// CodeRunner ocr接口
type CodeRunner interface {
	RecognizeText(ctx context.Context, params map[string]interface{}, attr *drivenadapters.DocAttr) (map[string]interface{}, error)
}

type coderunner struct {
	ocrType    string
	ocrAdapter drivenadapters.CodeRunner
}

type externalOCRResult struct {
	TaskID  string `json:"task_id"`
	RecType string `json:"rec_type"`
}

var (
	oOnce sync.Once
	o     CodeRunner
)

var supportMap = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"bmp":  true,
	"png":  true,
	"tif":  true,
	"tiff": true,
	"JPG":  true,
	"JPEG": true,
	"BMP":  true,
	"PNG":  true,
	"TIF":  true,
	"TIFF": true,
}

// MaxFileSize 最大能处理的文件大小
var MaxFileSize int64 = 100 * 1024 * 1024

// NewCoderunner 实例化
func NewCoderunner() CodeRunner {
	oOnce.Do(func() {
		config := common.NewConfig()
		o = &coderunner{
			ocrType:    config.T4th.Type,
			ocrAdapter: drivenadapters.NewCodeRunner(),
		}
	})
	return o
}

func (c *coderunner) RecognizeText(ctx context.Context, params map[string]interface{}, attr *drivenadapters.DocAttr) (map[string]interface{}, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 文件大小超过100M，跳过处理
	if attr.Size > 100*1024*1024 {
		return nil, errors.NewIError(errors.FileSizeExceed, "", map[string]interface{}{
			"doc": map[string]interface{}{
				"docid": attr.DocID,
				"limit": MaxFileSize,
			},
		})
	}

	fileExtension := ""
	dotIndex := strings.LastIndex(attr.Name, ".")
	if dotIndex != -1 {
		fileExtension = attr.Name[dotIndex+1:]
	}

	switch c.ocrType {
	case "fileReader":
		if !supportMap[fileExtension] && fileExtension != "pdf" && fileExtension != "PDF" {
			return nil, errors.NewIError(errors.FileTypeNotSupported, "", map[string]interface{}{
				"doc": map[string]interface{}{
					"docid":       attr.DocID,
					"supportType": "jpg/jpeg/bmp/png/tif/tiff/JPG/JPEG/BMP/PNG/TIF/TIFF/PDF/pdf",
					"docname":     attr.Name,
				},
			})
		}
		res, err := c.ocrAdapter.RecognizeTextByExternal(ctx, params)
		if err != nil {
			log.Warnf("[RecognizeText] RecognizeTextByExternal failed, detail: %s", err.Error())
			return nil, errors.NewIError(errors.InternalError, "", err.Error())
		}
		var ocrResult externalOCRResult
		resByte, _ := json.Marshal(res)
		_ = json.Unmarshal(resByte, &ocrResult)
		for {
			code, res, err := c.ocrAdapter.GetRecognizitionResult(ctx, ocrResult.TaskID, ocrResult.RecType)
			if err != nil {
				log.Warnf("[RecognizeText] GetRecognizitionResult failed, detail: %s", err.Error())
				return nil, errors.NewIError(errors.InternalError, "", err.Error())
			}
			if code == http.StatusAccepted {
				time.Sleep(1 * time.Second)
				continue
			}
			err = c.ocrAdapter.DeleteRecognizeTask(ctx, []string{ocrResult.TaskID})
			if err != nil {
				log.Warnf("[RecognizeText] DeleteRecognizeTask failed, detail: %s", err.Error())
			}
			return res, nil
		}
	case "ocr":
		if !supportMap[fileExtension] {
			return nil, errors.NewIError(errors.FileTypeNotSupported, "", map[string]interface{}{
				"doc": map[string]interface{}{
					"docid":       attr.DocID,
					"supportType": "jpg/jpeg/bmp/png/tif/tiff/JPG/JPEG/BMP/PNG/TIF/TIFF",
					"docname":     attr.Name,
				},
			})
		}
		res, err := c.ocrAdapter.RecognizeTextByBuildIn(ctx, params)
		if err != nil {
			log.Warnf("[RecognizeText] RecognizeTextByBuildIn failed, detail: %s", err.Error())
			return res, errors.NewIError(errors.InternalError, "", err.Error())
		}
		return res, nil
	default:
		log.Warnf("[RecognizeText] unsupported ocr type")
		return nil, errors.NewIError(errors.InvalidParameter, "", fmt.Errorf("unsupported ocr type"))
	}
}
