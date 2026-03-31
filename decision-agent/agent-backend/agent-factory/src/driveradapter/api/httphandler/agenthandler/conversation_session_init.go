package agenthandler

// import (
// 	"fmt"
// 	"net/http"

// 	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
// 	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
// 	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
// 	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
// 	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
// 	"github.com/kweaver-ai/kweaver-go-lib/rest"
// 	"github.com/gin-gonic/gin"
// )

// func (h *agentHTTPHandler) ConversationSessionInit(c *gin.Context) {
// 	var req agentreq.ConversationSessionInitReq
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		httpErr := capierr.New400Err(c, fmt.Sprintf("[ConversationSessionInit] should bind json err: %v", err))
// 		o11y.Error(c, fmt.Sprintf("[ConversationSessionInit] should bind json err: %v", err))
// 		h.logger.Errorf("[ConversationSessionInit] should bind json err: %v", err)
// 		rest.ReplyError(c, httpErr)

// 		return
// 	}

// 	if err := req.Check(); err != nil {
// 		httpErr := capierr.New400Err(c, fmt.Sprintf("[ConversationSessionInit] check req err: %v", err))
// 		o11y.Error(c, fmt.Sprintf("[ConversationSessionInit] check req err: %v", err))
// 		h.logger.Errorf("[ConversationSessionInit] check req err: %v", err)
// 		rest.ReplyError(c, httpErr)

// 		return
// 	}

// 	visitor := chelper.GetVisitorFromCtx(c)
// 	if visitor == nil {
// 		httpErr := capierr.New401Err(c, "[ConversationSessionInit] visitor not found")
// 		o11y.Error(c, "[ConversationSessionInit] visitor not found")
// 		h.logger.Errorf("[ConversationSessionInit] visitor not found")
// 		rest.ReplyError(c, httpErr)

// 		return
// 	}

// 	req.UserID = visitor.ID
// 	req.XAccountID = visitor.ID
// 	req.XAccountType.LoadFromMDLVisitorType(visitor.Type)
// 	req.XBusinessDomainID = chelper.GetBizDomainIDFromCtx(c)

// 	rt, err := h.agentSvc.ConversationSessionInit(c.Request.Context(), &req)
// 	if err != nil {
// 		httpErr := rest.NewHTTPError(c.Request.Context(), http.StatusInternalServerError,
// 			apierr.AgentAPP_Agent_SessionInitFailed).WithErrorDetails(fmt.Sprintf("[ConversationSessionInit] conversation session init err: %v", err))

// 		o11y.Error(c, fmt.Sprintf("[ConversationSessionInit] conversation session init err: %v", err))
// 		h.logger.Errorf("[ConversationSessionInit] conversation session init err: %v", err)
// 		rest.ReplyError(c, httpErr)

// 		return
// 	}

// 	rest.ReplyOK(c, http.StatusOK, rt)
// }
