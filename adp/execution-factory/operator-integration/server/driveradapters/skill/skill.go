package skill

import (
	"fmt"
	"net/http"

	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

func (h *skillHandler) RegisterSkill(c *gin.Context) {
	req := &interfaces.RegisterSkillReq{}
	if err := c.ShouldBindHeader(req); err != nil {
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error()))
		return
	}
	switch c.ContentType() {
	case "application/x-www-form-urlencoded":
		if err := utils.GetBindFormRaw(c, req); err != nil {
			rest.ReplyError(c, err)
			return
		}
	case "multipart/form-data":
		fileBytes, err := utils.GetBindMultipartFormRaw(c, req, "file", 0)
		if err != nil {
			rest.ReplyError(c, err)
			return
		}
		req.File = fileBytes
	default:
		rest.ReplyError(c, errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "unsupported content type"))
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
	resp, err := h.Registry.RegisterSkill(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) DeleteSkill(c *gin.Context) {
	req := &interfaces.DeleteSkillReq{}
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
	if err := h.Registry.DeleteSkill(c.Request.Context(), req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, gin.H{"skill_id": req.SkillID, "deleted": true})
}

func (h *skillHandler) DownloadSkill(c *gin.Context) {
	req := &interfaces.DownloadSkillReq{}
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
	resp, err := h.Registry.DownloadSkill(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", resp.FileName))
	c.Data(http.StatusOK, "application/zip", resp.Content)
}

func (h *skillHandler) QuerySkillList(c *gin.Context) {
	req := &interfaces.QuerySkillListReq{}
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
	resp, err := h.Registry.QuerySkillList(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) QuerySkillMarketList(c *gin.Context) {
	req := &interfaces.QuerySkillMarketListReq{}
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
	resp, err := h.Market.QuerySkillMarketList(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) GetSkillMarketDetail(c *gin.Context) {
	req := &interfaces.GetSkillMarketDetailReq{}
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
	resp, err := h.Market.GetSkillMarketDetail(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) GetSkillDetail(c *gin.Context) {
	req := &interfaces.GetSkillDetailReq{}
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
	resp, err := h.Registry.GetSkillDetail(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) GetSkillContent(c *gin.Context) {
	req := &interfaces.GetSkillContentReq{}
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
	resp, err := h.Reader.GetSkillContent(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) ReadSkillFile(c *gin.Context) {
	req := &interfaces.ReadSkillFileReq{}
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
	resp, err := h.Reader.ReadSkillFile(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) UpdateSkillStatus(c *gin.Context) {
	req := &interfaces.UpdateSkillStatusReq{}
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
	resp, err := h.Registry.UpdateSkillStatus(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}

func (h *skillHandler) ExecuteSkill(c *gin.Context) {
	req := &interfaces.ExecuteSkillReq{}
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
	resp, err := h.Registry.ExecuteSkill(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}
