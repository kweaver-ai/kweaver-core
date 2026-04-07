package interfaces

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
)

//go:generate mockgen -source ../interfaces/auth_service.go -destination ../interfaces/mock/mock_auth_service.go
type AuthService interface {
	VerifyToken(ctx context.Context, c *gin.Context) (hydra.Visitor, error)
}
