package skill

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

func (h *skillHandler) CreateSkillIndexBuildTask(c *gin.Context) {
	req := &interfaces.CreateSkillIndexBuildTaskReq{}
	if err := c.ShouldBindHeader(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := utils.GetBindJSONRaw(c, req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	if err := validator.New().Struct(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.IndexBuildService.CreateTask(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) QuerySkillIndexBuildTaskList(c *gin.Context) {
	req := &interfaces.QuerySkillIndexBuildTaskListReq{}
	if err := c.ShouldBindHeader(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := c.ShouldBindQuery(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := defaults.Set(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := validator.New().Struct(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.IndexBuildService.QueryTaskList(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) GetSkillIndexBuildTask(c *gin.Context) {
	req := &interfaces.GetSkillIndexBuildTaskReq{}
	if err := c.ShouldBindHeader(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := c.ShouldBindUri(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := validator.New().Struct(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.IndexBuildService.GetTask(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) CancelSkillIndexBuildTask(c *gin.Context) {
	req := &interfaces.CancelSkillIndexBuildTaskReq{}
	if err := c.ShouldBindHeader(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := c.ShouldBindUri(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := validator.New().Struct(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.IndexBuildService.CancelTask(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) RetrySkillIndexBuildTask(c *gin.Context) {
	req := &interfaces.RetrySkillIndexBuildTaskReq{}
	if err := c.ShouldBindHeader(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := c.ShouldBindUri(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := validator.New().Struct(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.IndexBuildService.RetryTask(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}
