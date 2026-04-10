package callback

import (
	"io"
	"sync"

	"github.com/gin-gonic/gin"
	commonLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/dependency"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
)

type RESTHandler interface {
	RegisterPrivateAPI(engine *gin.RouterGroup)
}

var (
	once sync.Once
	rh   RESTHandler
)

type restHandler struct {
	mgnt         mgnt.MgntHandler
	docConverter dependency.DocumentConverter
}

func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			mgnt:         mgnt.NewMgnt(),
			docConverter: dependency.NewDocumentConverter(),
		}
	})

	return rh
}

func (h *restHandler) RegisterPrivateAPI(engine *gin.RouterGroup) {
	engine.POST("/gotenberg/callback/success", h.gotenbergSuccessCallback)
	engine.POST("/gotenberg/callback/error", h.gotenbergErrorCallback)
}

func (h *restHandler) gotenbergSuccessCallback(c *gin.Context) {
	ctx, span := trace.StartConsumerSpan(c.Request.Context())
	defer func() { trace.TelemetrySpanEnd(span, nil) }()

	taskID := c.Request.Header.Get("X-Task-ID")
	docID := c.Request.Header.Get("X-Source-Doc-ID")
	fileName := c.Request.Header.Get("X-Result-File-Name")
	contentType := c.Request.Header.Get("Content-Type")
	log := commonLog.NewLogger()

	result, err := h.docConverter.HandleGotenbergCallback(ctx, &dependency.GotenbergCallbackRequest{
		TaskID:      taskID,
		DocID:       docID,
		FileName:    fileName,
		ContentType: contentType,
		Body:        c.Request.Body,
		Size:        c.Request.ContentLength,
	})
	if err != nil {
		if log != nil {
			log.Warnf("[gotenbergCallback] handle callback failed, taskID: %s, docID: %s, err: %s", taskID, docID, err.Error())
		}
		if taskID != "" {
			_ = h.mgnt.ContinueBlockInstances(ctx, []string{taskID}, map[string]interface{}{
				"code":    "gotenberg_callback_failed",
				"message": err.Error(),
			}, entity.TaskInstanceStatusFailed)
		}
		c.Status(200)
		return
	}

	if err = h.mgnt.ContinueBlockInstances(ctx, []string{taskID}, result, entity.TaskInstanceStatusSuccess); err != nil {
		if log != nil {
			log.Warnf("[gotenbergCallback] ContinueBlockInstances failed, taskID: %s, err: %s", taskID, err.Error())
		}
		c.JSON(200, gin.H{"code": 0, "message": "accepted"})
		return
	}

	c.Status(200)
}

func (h *restHandler) gotenbergErrorCallback(c *gin.Context) {
	ctx, span := trace.StartConsumerSpan(c.Request.Context())
	defer func() { trace.TelemetrySpanEnd(span, nil) }()

	taskID := c.Request.Header.Get("X-Task-ID")
	log := commonLog.NewLogger()
	body, _ := io.ReadAll(c.Request.Body)

	if taskID != "" {
		err := h.mgnt.ContinueBlockInstances(ctx, []string{taskID}, map[string]interface{}{
			"code":    "gotenberg_callback_failed",
			"message": string(body),
		}, entity.TaskInstanceStatusFailed)
		if err != nil && log != nil {
			log.Warnf("[gotenbergCallback] ContinueBlockInstances failed, taskID: %s, err: %s", taskID, err.Error())
		}
	}

	c.Status(200)
}
