package apimiddleware

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/squaresvc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func CheckAgentUsePms() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 方法3：使用 gin 的 GetRawData
		body, err := c.GetRawData()
		if err != nil {
			httpErr := capierr.New400Err(c, "[CheckAgentUsePms] get raw data failed")
			rest.ReplyError(c, httpErr)
			c.Abort()

			return
		}

		var data map[string]interface{}
		if err := sonic.Unmarshal(body, &data); err != nil {
			httpErr := capierr.New400Err(c, "[CheckAgentUsePms] unmarshal body failed")
			rest.ReplyError(c, httpErr)
			c.Abort()

			return
		}

		agentID := ""
		// 首先尝试从请求体中获取
		if agentValue, ok := data["agent_id"].(string); ok {
			agentID = agentValue
		} else if agentValue, ok := data["agent_key"].(string); ok {
			// NOTE: 通过AgentKey查询AgentID
			resolvedID, err := squaresvc.NewSquareService().CheckAndGetID(c.Request.Context(), agentValue)
			if err != nil {
				httpErr := capierr.New400Err(c, "[CheckAgentUsePms] get agent id by agent key failed")
				rest.ReplyError(c, httpErr)
				c.Abort()

				return
			}

			agentID = resolvedID
		} else {
			// 如果请求体中获取不到，尝试从URL路径参数中获取
			if agentValue := c.Param("agent_id"); agentValue != "" {
				agentID = agentValue
			} else if agentValue := c.Param("agent_key"); agentValue != "" {
				// NOTE: 通过AgentKey查询AgentID
				resolvedID, err := squaresvc.NewSquareService().CheckAndGetID(c.Request.Context(), agentValue)
				if err != nil {
					httpErr := capierr.New400Err(c, "[CheckAgentUsePms] get agent id by agent key from path param failed")
					rest.ReplyError(c, httpErr)
					c.Abort()

					return
				}

				agentID = resolvedID
			} else {
				httpErr := capierr.New400Err(c, "[CheckAgentUsePms] one of agent_id and agent_key is required,type must be string, can be in request body or path param")
				rest.ReplyError(c, httpErr)
				c.Abort()

				return
			}
		}
		// 重新设置请求体
		// NOTE: 读取完请求体后需要重新设置，否则后续的处理器无法读取
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		visitor := chelper.GetVisitorFromCtx(c)
		if visitor == nil {
			httpErr := capierr.New401Err(c, "[CheckAgentUsePms] user not found")
			rest.ReplyError(c, httpErr)
			c.Abort()

			return
		}

		var req *capimiddleware.CheckPmsReq
		if visitor.Type == rest.VisitorType_App {
			req = capimiddleware.NewCheckAgentUsePmsReq(agentID, "", visitor.ID)
		} else {
			req = capimiddleware.NewCheckAgentUsePmsReq(agentID, visitor.ID, "")
		}

		handler := capimiddleware.CheckPms(req, func(c *gin.Context, hasPms bool) {
			if !hasPms {
				httpErr := capierr.NewCustom403Err(c, apierr.AgentAPP_Forbidden_PermissionDenied, fmt.Sprintf("user %s has no permission to use agent %s", visitor.ID, agentID))
				rest.ReplyError(c, httpErr)
				c.Abort()

				return
			}
		})

		handler(c)

		c.Next()
	}
}

func CheckAgentUsePmsInternal() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 方法3：使用 gin 的 GetRawData
		body, err := c.GetRawData()
		if err != nil {
			httpErr := capierr.New400Err(c, "[CheckAgentUsePmsInternal] get raw data failed")
			rest.ReplyError(c, httpErr)
			c.Abort()

			return
		}

		var data map[string]interface{}
		if err := sonic.Unmarshal(body, &data); err != nil {
			httpErr := capierr.New400Err(c, "[CheckAgentUsePmsInternal] unmarshal body failed")
			rest.ReplyError(c, httpErr)
			c.Abort()

			return
		}

		agentID := ""
		// 首先尝试从请求体中获取
		if agentValue, ok := data["agent_id"].(string); ok {
			agentID = agentValue
		} else if agentValue, ok := data["agent_key"].(string); ok {
			// NOTE: 通过AgentKey查询AgentID,一一对应与agent_version无关
			resolvedID, err := squaresvc.NewSquareService().CheckAndGetID(c.Request.Context(), agentValue)
			if err != nil {
				httpErr := capierr.New400Err(c, "[CheckAgentUsePmsInternal] get agent id by agent key failed")
				rest.ReplyError(c, httpErr)
				c.Abort()

				return
			}

			agentID = resolvedID
		} else {
			// 如果请求体中获取不到，尝试从URL路径参数中获取
			if agentValue := c.Param("agent_id"); agentValue != "" {
				agentID = agentValue
			} else if agentValue := c.Param("agent_key"); agentValue != "" {
				// NOTE: 通过AgentKey查询AgentID,一一对应与agent_version无关
				resolvedID, err := squaresvc.NewSquareService().CheckAndGetID(c.Request.Context(), agentValue)
				if err != nil {
					httpErr := capierr.New400Err(c, "[CheckAgentUsePmsInternal] get agent id by agent key from path param failed")
					rest.ReplyError(c, httpErr)
					c.Abort()

					return
				}

				agentID = resolvedID
			} else {
				httpErr := capierr.New400Err(c, "[CheckAgentUsePmsInternal] one of agent_id and agent_key is required,type must be string, can be in request body or path param")
				rest.ReplyError(c, httpErr)
				c.Abort()

				return
			}
		}
		// 重新设置请求体
		// NOTE: 读取完请求体后需要重新设置，否则后续的处理器无法读取
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		if global.GConfig != nil && global.GConfig.SwitchFields != nil && global.GConfig.SwitchFields.DisablePmsCheck {
			c.Next()
			return
		}

		userID := c.Request.Header.Get("x-account-id")

		if userID == "" {
			httpErr := capierr.New401Err(c, "[CheckAgentUsePmsInternal] user not found")
			rest.ReplyError(c, httpErr)
			c.Abort()

			return
		}

		accountType := c.Request.Header.Get("x-account-type")
		// NOTE: 如果visitorType为空，则默认是实名用户
		if accountType == "" {
			accountType = "user"
		}

		var req *capimiddleware.CheckPmsReq
		if accountType == "app" {
			req = capimiddleware.NewCheckAgentUsePmsReq(agentID, "", userID)
		} else if accountType == "user" || accountType == "anonymous" {
			req = capimiddleware.NewCheckAgentUsePmsReq(agentID, userID, "")
		} else {
			httpErr := capierr.New400Err(c, "[CheckAgentUsePmsInternal] account type not found")
			rest.ReplyError(c, httpErr)
			c.Abort()

			return
		}

		handler := capimiddleware.CheckPms(req, func(c *gin.Context, hasPms bool) {
			if !hasPms {
				httpErr := capierr.NewCustom403Err(c, apierr.AgentAPP_Forbidden_PermissionDenied,
					fmt.Sprintf("user %s has no permission to use agent %s", userID, agentID))

				rest.ReplyError(c, httpErr)
				c.Abort()

				return
			}
		})

		handler(c)
		c.Next()
	}
}
