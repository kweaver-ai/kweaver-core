package dbaulid

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

var (
	ulidRepoOnce sync.Once
	ulidRepoImpl idbaccess.UlidRepo
)

type ulidRepo struct {
	db     *sqlx.DB
	logger icmp.Logger
}

type UniqueID struct {
	ID   string                 `json:"id" db:"f_id"`
	Flag cconstant.UniqueIDFlag `json:"flag" db:"f_flag"`
}

func (p *UniqueID) TableName() string {
	return "t_stc_unique_id"
}

var _ idbaccess.UlidRepo = &ulidRepo{}

func NewUlidRepo() idbaccess.UlidRepo {
	ulidRepoOnce.Do(func() {
		ulidRepoImpl = &ulidRepo{
			db:     cglobal.GDB,
			logger: logger.GetLogger(),
		}
	})

	return ulidRepoImpl
}
