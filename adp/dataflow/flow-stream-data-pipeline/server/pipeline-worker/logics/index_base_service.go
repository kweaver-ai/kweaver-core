package logics

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-worker/interfaces"
)

var (
	imServiceOnce sync.Once
	imService     interfaces.IndexBaseService
)

type indexBaseService struct {
	appSetting *common.AppSetting
	ima        interfaces.IndexBaseAccess
}

func NewIndexBaseService(appSetting *common.AppSetting) interfaces.IndexBaseService {
	imServiceOnce.Do(func() {
		imService = &indexBaseService{
			appSetting: appSetting,
			ima:        IBAccess,
		}
	})
	return imService
}

func (ibs *indexBaseService) GetIndexBaseByBaseType(ctx context.Context, baseType string) (*interfaces.IndexBaseInfo, error) {
	indexbases, err := ibs.ima.GetIndexBasesByTypes(ctx, []string{baseType})
	if err != nil {
		logger.Errorf("failed to get index base by base type, error: %v", err)
		return nil, err
	}

	if len(indexbases) == 0 {
		return nil, fmt.Errorf("no index base found for base type %s", baseType)
	}

	return indexbases[0], nil
}
