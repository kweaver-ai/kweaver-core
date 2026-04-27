package visitor

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"

	"vega-backend/interfaces"
)

func GenerateVisitor(c *gin.Context) hydra.Visitor {
	accountInfo := interfaces.AccountInfo{
		ID:   c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_ID),
		Type: c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_TYPE),
	}

	visitor := hydra.Visitor{
		ID:         accountInfo.ID,
		Type:       hydra.VisitorType(accountInfo.Type),
		TokenID:    "",
		IP:         c.ClientIP(),
		Mac:        c.GetHeader("X-Request-MAC"),
		UserAgent:  c.GetHeader("User-Agent"),
		ClientType: hydra.ClientType_Linux,
	}
	return visitor
}
