package categorysvc

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/categoryacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
)

var (
	categorySvcOnce sync.Once
	categorySvcImpl iv3portdriver.ICategorySvc
)

type categorySvc struct {
	*service.SvcBase
	categoryRepo idbaccess.ICategoryRepo
}

var _ iv3portdriver.ICategorySvc = &categorySvc{}

func NewCategorySvc() iv3portdriver.ICategorySvc {
	categorySvcOnce.Do(func() {
		categorySvcImpl = &categorySvc{
			SvcBase:      service.NewSvcBase(),
			categoryRepo: categoryacc.NewCategoryRepo(),
		}
	})

	return categorySvcImpl
}
