package auth

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"vega-backend/common"
	"vega-backend/interfaces"
)

var (
	hAccessOnce sync.Once
	hAccess     interfaces.AuthAccess
)

type hydraAuthAccess struct {
	appSetting *common.AppSetting
	hydra      hydra.Hydra
}

func NewHydraAuthAccess(appSetting *common.AppSetting) interfaces.AuthAccess {
	hAccessOnce.Do(func() {
		hAccess = &hydraAuthAccess{
			appSetting: appSetting,
			hydra:      hydra.NewHydra(appSetting.HydraAdminSetting),
		}
	})

	return hAccess
}

func (ha *hydraAuthAccess) VerifyToken(ctx context.Context, c *gin.Context) (hydra.Visitor, error) {
	visitor, err := ha.hydra.VerifyToken(ctx, c)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, 401, rest.PublicError_Unauthorized).
			WithErrorDetails(err.Error())
		logger.Errorf("VerifyToken failed: %v", err)
		return visitor, httpErr
	}

	return visitor, nil
}
