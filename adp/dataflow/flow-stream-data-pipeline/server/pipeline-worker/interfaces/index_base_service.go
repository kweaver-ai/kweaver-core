package interfaces

import (
	"context"
)

//go:generate mockgen -source ../interfaces/index_base_service.go -destination ../interfaces/mock/mock_index_base_service.go
type IndexBaseService interface {
	GetIndexBaseByBaseType(ctx context.Context, baseType string) (*IndexBaseInfo, error)
}
