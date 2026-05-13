package skill

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

func (h *skillHandler) GetManagementContent(c *gin.Context) {
	req := &interfaces.GetManagementContentReq{}
	if err := c.ShouldBindHeader(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := c.ShouldBindUri(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := c.ShouldBindQuery(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := validator.New().Struct(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := h.MgmtReader.GetManagementContent(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) ReadManagementFile(c *gin.Context) {
	req := &interfaces.ReadManagementFileReq{}
	if err := c.ShouldBindHeader(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	if err := c.ShouldBindUri(req); err != nil {
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
	resp, err := h.MgmtReader.ReadManagementFile(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) DownloadManagementSkill(c *gin.Context) {
	req := &interfaces.DownloadManagementSkillReq{}
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
	resp, err := h.MgmtReader.DownloadManagementSkill(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", resp.FileName))
	c.Data(http.StatusOK, "application/zip", resp.Content)
}
