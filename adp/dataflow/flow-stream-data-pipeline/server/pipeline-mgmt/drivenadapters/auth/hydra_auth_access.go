package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
)

type hydraAuthAccess struct {
	hydra hydra.Hydra
}

func NewHydraAuthAccess(appSetting *common.AppSetting) interfaces.AuthAccess {
	return &hydraAuthAccess{
		hydra: hydra.NewHydra(appSetting.HydraAdminSetting),
	}
}

func (h *hydraAuthAccess) VerifyToken(ctx context.Context, c *gin.Context) (hydra.Visitor, error) {
	return h.hydra.VerifyToken(ctx, c)
}
