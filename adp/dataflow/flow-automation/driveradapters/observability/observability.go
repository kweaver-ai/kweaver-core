package observability

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/observability"
)

type RESTHandler interface {
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once                sync.Once
	rh                  RESTHandler
	observabilitySchema = "base/observability.json"
)

type observabilityRESTHandler struct {
	observability observability.ObservabilityHandler
}

func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &observabilityRESTHandler{
			observability: observability.NewObservability(),
		}
	})

	return rh
}

func (h *observabilityRESTHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.GET("/observability/full-view", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.fullView)
	engine.GET("/observability/runtime-view", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.runtimeView)
	engine.GET("/observability/recent", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.recent)
	engine.GET("/observability/visible", middleware.TokenAuth(), h.visible)
}

func (h *observabilityRESTHandler) fullView(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	query, err := common.ParseQuery(c)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	queryByte, _ := json.Marshal(query)
	err = common.JSONSchemaValidV2(c.Request.Context(), queryByte, observabilitySchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var params observability.ObservabilityQueryParams
	err = json.Unmarshal(queryByte, &params)
	if err != nil {
		errors.ReplyError(c, ierr.NewPublicRestError(c.Request.Context(), ierr.PErrorBadRequest, ierr.PErrorBadRequest, err.Error()))
		return
	}

	params.BizDomainID = c.GetString("bizDomainID")
	res, err := h.observability.FullView(c.Request.Context(), params, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *observabilityRESTHandler) runtimeView(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	query, err := common.ParseQuery(c)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	queryByte, _ := json.Marshal(query)
	err = common.JSONSchemaValidV2(c.Request.Context(), queryByte, observabilitySchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var params observability.ObservabilityQueryParams
	err = json.Unmarshal(queryByte, &params)
	if err != nil {
		errors.ReplyError(c, ierr.NewPublicRestError(c.Request.Context(), ierr.PErrorBadRequest, ierr.PErrorBadRequest, err.Error()))
		return
	}

	params.BizDomainID = c.GetString("bizDomainID")
	res, err := h.observability.RuntimeView(c.Request.Context(), params, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *observabilityRESTHandler) recent(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	query, err := common.ParseQuery(c)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	queryByte, _ := json.Marshal(query)
	err = common.JSONSchemaValidV2(c.Request.Context(), queryByte, observabilitySchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	params := observability.ObservabilityQueryParams{
		BizDomainID: c.GetString("bizDomainID"),
	}

	if _, ok := query["trigger"]; ok {
		trigger := query["trigger"].(string)
		params.Trigger = &trigger
	}

	res, err := h.observability.RecentRunView(c.Request.Context(), params, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)

}

// visible 概览可见性判断
func (h *observabilityRESTHandler) visible(c *gin.Context) {
	isVisible := h.observability.IsVisible(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{"visible": isVisible})
}
