package toolbox

import (
	"net/http"

	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

// CreateToolBox 创建工具箱
func (h *toolBoxHandler) CreateToolBox(c *gin.Context) {
	req := &interfaces.CreateToolBoxReq{
		OpenAPIInput: &interfaces.OpenAPIInput{},
	}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	switch c.ContentType() {
	case "application/json":
		err = utils.GetBindJSONRaw(c, req)
	case "application/x-www-form-urlencoded":
		err = utils.GetBindFormRaw(c, req)
	case "multipart/form-data":
		req.Data, err = utils.GetBindMultipartFormRaw(c, req, "data", 0)
	default:
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "unsupported content type")
	}
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	// 检验传参大小
	resp, err := h.ToolService.CreateToolBox(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// UpdateToolBox 更新工具箱
func (h *toolBoxHandler) UpdateToolBox(c *gin.Context) {
	req := &interfaces.UpdateToolBoxReq{
		OpenAPIInput: &interfaces.OpenAPIInput{},
	}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	switch c.ContentType() {
	case "application/json":
		err = utils.GetBindJSONRaw(c, req)
	case "application/x-www-form-urlencoded":
		err = utils.GetBindFormRaw(c, req)
	case "multipart/form-data":
		req.Data, err = utils.GetBindMultipartFormRaw(c, req, "data", 0)
	default:
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "unsupported content type")
	}
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	// 校验参数
	err = h.Validator.ValidatorToolBoxName(c.Request.Context(), req.BoxName)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = h.Validator.ValidatorToolBoxDesc(c.Request.Context(), req.BoxDesc)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.UpdateToolBox(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// QueryToolBox 查询工具箱
func (h *toolBoxHandler) QueryToolBox(c *gin.Context) {
	req := &interfaces.GetToolBoxReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.GetToolBox(c.Request.Context(), req, false)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *toolBoxHandler) UpdateToolBoxStatus(c *gin.Context) {
	req := &interfaces.UpdateToolBoxStatusReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = utils.GetBindJSONRaw(c, req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.UpdateToolBoxStatus(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// DeleteToolBox 删除工具箱
func (h *toolBoxHandler) DeleteToolBox(c *gin.Context) {
	req := &interfaces.DeleteBoxReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.DeleteBoxByID(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// QueryToolBoxPage 查询工具箱分页
func (h *toolBoxHandler) QueryToolBoxPage(c *gin.Context) {
	req := &interfaces.QueryToolBoxListReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindQuery(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.QueryToolBoxList(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// CreateTool 创建工具
func (h *toolBoxHandler) CreateTool(c *gin.Context) {
	req := &interfaces.CreateToolReq{
		OpenAPIInput: &interfaces.OpenAPIInput{},
		FunctionInput: &interfaces.FunctionInput{
			Dependencies: []*interfaces.DependencyInfo{},
		},
	}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	switch c.ContentType() {
	case "application/json":
		err = utils.GetBindJSONRaw(c, req)
	case "application/x-www-form-urlencoded":
		err = utils.GetBindFormRaw(c, req)
	case "multipart/form-data":
		req.Data, err = utils.GetBindMultipartFormRaw(c, req, "data", 0)
	default:
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "unsupported content type")
	}
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.CreateTool(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// UpdateTool 更新工具
func (h *toolBoxHandler) UpdateTool(c *gin.Context) {
	req := &interfaces.UpdateToolReq{
		OpenAPIInput: &interfaces.OpenAPIInput{},
		FunctionInputEdit: &interfaces.FunctionInputEdit{
			Dependencies: []*interfaces.DependencyInfo{},
		},
	}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	switch c.ContentType() {
	case "application/json":
		err = utils.GetBindJSONRaw(c, req)
	case "application/x-www-form-urlencoded":
		err = utils.GetBindFormRaw(c, req)
	case "multipart/form-data":
		req.Data, err = utils.GetBindMultipartFormRaw(c, req, "data", 0)
	default:
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "unsupported content type")
	}
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	// 参数校验
	err = h.Validator.ValidatorToolName(c.Request.Context(), req.ToolName)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = h.Validator.ValidatorToolDesc(c.Request.Context(), req.ToolDesc)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.UpdateTool(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// QueryTool 查询工具
func (h *toolBoxHandler) QueryTool(c *gin.Context) {
	req := &interfaces.GetToolReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.GetBoxTool(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// DeleteBoxTool 删除工具
func (h *toolBoxHandler) DeleteBoxTool(c *gin.Context) {
	req := &interfaces.BatchDeleteToolReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindJSON(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.DeleteBoxTool(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// QueryBoxToolPage 查询工具分页
func (h *toolBoxHandler) QueryBoxToolPage(c *gin.Context) {
	req := &interfaces.QueryToolListReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindQuery(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.QueryToolList(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// UpdateToolStatus 更新工具状态
func (h *toolBoxHandler) UpdateToolStatus(c *gin.Context) {
	req := &interfaces.UpdateToolStatusReq{
		ToolStatusList: []*interfaces.ToolStatus{},
	}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = utils.GetBindJSONRaw(c, &req.ToolStatusList)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	for _, toolStatus := range req.ToolStatusList {
		err = validator.New().Struct(toolStatus)
		if err != nil {
			rest.ReplyError(c, err)
			return
		}
	}
	resp, err := h.ToolService.UpdateToolStatus(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// GetMarketToolList 获取所有工具
func (h *toolBoxHandler) GetMarketToolList(c *gin.Context) {
	req := &interfaces.QueryMarketToolListReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindQuery(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.GetMarketToolList(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// DebugTool 调试工具
func (h *toolBoxHandler) DebugTool(c *gin.Context) {
	req := &interfaces.ExecuteToolReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindJSON(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.DebugTool(c.Request.Context(), req)
	rest.ReplyWithExecutionMode(c, resp, err)
}

// ExecuteTool 执行工具
func (h *toolBoxHandler) ExecuteTool(c *gin.Context) {
	req := &interfaces.ExecuteToolReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindJSON(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.ExecuteTool(c.Request.Context(), req)
	rest.ReplyWithExecutionMode(c, resp, err)
}

// OperatorToTool 算子转换成工具
func (h *toolBoxHandler) OperatorToTool(c *gin.Context) {
	req := &interfaces.ConvertOperatorToToolReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindJSON(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.ConvertOperatorToTool(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

// 获取发布工具箱信息
func (h *toolBoxHandler) GetReleaseToolBoxInfo(c *gin.Context) {
	req := &interfaces.GetReleaseToolBoxInfoReq{}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = c.ShouldBindUri(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.ToolService.GetReleaseToolBoxInfo(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}
