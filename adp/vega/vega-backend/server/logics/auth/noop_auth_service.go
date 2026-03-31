package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"

	"vega-backend/common"
	"vega-backend/common/visitor"
	"vega-backend/interfaces"
)

type NoopAuthService struct {
	appSetting *common.AppSetting
}

func NewNoopAuthService(appSetting *common.AppSetting) interfaces.AuthService {
	return &NoopAuthService{
		appSetting: appSetting,
	}
}

func (n *NoopAuthService) VerifyToken(ctx context.Context, c *gin.Context) (hydra.Visitor, error) {
	return visitor.GenerateVisitor(c), nil
}
