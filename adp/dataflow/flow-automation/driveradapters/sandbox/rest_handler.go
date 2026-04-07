package sandbox

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/sandbox"
)

const sandboxExecutionSchema = "base/sandbox_execution.json"

type RESTHandler interface {
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once sync.Once
	rh   RESTHandler
)

type restHandler struct {
	sandboxHandler sandbox.SandboxHandler
}

func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			sandboxHandler: sandbox.NewSandbox(),
		}
	})
	return rh
}

func (h *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.POST("/sandbox-execution", middleware.TokenAuth(), h.executeSandbox)
}

func (h *restHandler) executeSandbox(c *gin.Context) {
	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, sandboxExecutionSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param sandbox.SandboxExecuteRequest
	err = json.Unmarshal(data, &param)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	if param.Language == "" {
		param.Language = "python"
	}

	result, err := h.sandboxHandler.Execute(c.Request.Context(), &param)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
