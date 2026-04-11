package v3agentconfighandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// @Summary      Agent列表（benchmarch）
// @Description  - Agent列表（供模型工厂做 Agent 的 Benchmarch 使用） - 只提供`内部接口`
// @Tags         agent-ignore
// @Accept       json
// @Produce      json
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent [get]
func (h *daConfHTTPHandler) AgentListListForBenchmark(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)
	req := &agentconfigreq.ListForBenchmarkReq{}

	if err := c.ShouldBindQuery(req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, req))
		_ = c.Error(httpErr)

		return
	}

	resp, err := h.daConfSvc.ListForBenchmark(ctx, req)
	if err != nil {
		h.logger.Errorf("AgentList error cause: %v, err trace: %+v\n", errors.Cause(err), err)

		_ = c.Error(err)

		return
	}

	// 返回成功
	rest.ReplyOK(c, http.StatusOK, resp)
}
