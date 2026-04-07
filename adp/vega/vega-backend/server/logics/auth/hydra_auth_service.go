package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"

	"vega-backend/common"
	"vega-backend/interfaces"
	"vega-backend/logics"
)

type hydraAuthService struct {
	appSetting *common.AppSetting
	aa         interfaces.AuthAccess
}

func NewHydraAuthService(appSetting *common.AppSetting) interfaces.AuthService {
	return &hydraAuthService{
		appSetting: appSetting,
		aa:         logics.AA,
	}
}

func (s *hydraAuthService) VerifyToken(ctx context.Context, c *gin.Context) (hydra.Visitor, error) {
	return s.aa.VerifyToken(ctx, c)
}
