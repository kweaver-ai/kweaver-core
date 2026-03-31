package operators

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	aerr "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/policy"
)

type RESTHandler interface {
	RegisterAPI(engine *gin.RouterGroup)
	RegisterPrivateAPI(engine *gin.RouterGroup)
}

var (
	once                 sync.Once
	rh                   RESTHandler
	createOperatorSchema = "base/create_operator.json"
	updateOperatorSchema = "base/update_operator.json"
	listOperatorSchema   = "base/list_operator.json"
	importOperatorSchema = "base/import_operator.json"
)

type OperatorsRESTHandler struct {
	mgnt   mgnt.MgntHandler
	policy policy.Handler
}

func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &OperatorsRESTHandler{
			mgnt:   mgnt.NewMgnt(),
			policy: policy.NewPolicy(),
		}
	})

	return rh
}

func (h *OperatorsRESTHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.POST("/operators/:id/executions", middleware.TokenAuth(), h.execute)
	engine.GET("/executions/:id", middleware.TokenAuth(), h.getExecution)
	engine.POST("/continuations/:id", h.continueExecution)
	engine.POST("/operators", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.registerOperator)
	// 接口无人使用暂时去除
	// engine.GET("/operators", middleware.TokenAuth(), h.listOperator)
	engine.PUT("/operators/:id", middleware.TokenAuth(), h.updateOperator)
}

// 注册私有API
func (h *OperatorsRESTHandler) RegisterPrivateAPI(engine *gin.RouterGroup) {
	engine.GET("/operators/configs/export", h.export)
	engine.PUT("/operators/configs/import", h.importOp)
}

func (h *OperatorsRESTHandler) registerOperator(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	ctx := c.Request.Context()
	err := common.JSONSchemaValidV2(ctx, data, createOperatorSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.ComboOperatorReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		errors.ReplyError(c, ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, err.Error()))
		return
	}

	param.BizDomainID = c.GetString("bizDomainID")
	dagID, operatorID, err := h.mgnt.CreateComboOperator(ctx, &param, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          dagID,
		"operator_id": operatorID,
	})

}

func (h *OperatorsRESTHandler) updateOperator(c *gin.Context) {
	id := c.Param("id")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	ctx := c.Request.Context()
	err := common.JSONSchemaValidV2(ctx, data, updateOperatorSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.OptionalComboOperatorReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		errors.ReplyError(c, ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, err.Error()))
		return
	}

	param.DagID = id
	err = h.mgnt.UpdateComboOperator(ctx, &param, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *OperatorsRESTHandler) listOperator(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	query, err := common.ParseQuery(c)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	ctx := c.Request.Context()
	queryByte, _ := json.Marshal(query)
	err = common.JSONSchemaValidV2(ctx, queryByte, listOperatorSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	ops, err := h.mgnt.ListComboOperator(ctx, query, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, ops)
}

func (h *OperatorsRESTHandler) export(c *gin.Context) {
	ids := strings.Split(strings.TrimSpace(c.Query("id")), ",")

	exportRes, err := h.mgnt.ExportOperator(c.Request.Context(), ids)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, exportRes)
}

func (h *OperatorsRESTHandler) importOp(c *gin.Context) {
	userID := c.GetHeader("X-User")

	if userID == "" {
		errors.ReplyError(c, ierr.NewPublicRestError(c.Request.Context(), ierr.PErrorForbidden, aerr.DescKeyMissingUserInfo, nil))
		return
	}

	userInfo := &drivenadapters.UserInfo{
		UserID: userID,
	}

	data, _ := io.ReadAll(c.Request.Body)

	ctx := c.Request.Context()
	err := common.JSONSchemaValidV2(ctx, data, importOperatorSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.ImportOperatorReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		errors.ReplyError(c, ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, err.Error()))
		return
	}

	err = h.mgnt.ImportOperator(ctx, &param, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}
