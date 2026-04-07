package dbaulid

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// GenUniqID 生成一个唯一ID
func (repo *ulidRepo) GenUniqID(ctx context.Context, flag cconstant.UniqueIDFlag) (id string, err error) {
	maxRetry := 5
	for i := 0; i < maxRetry; i++ {
		id, err = repo.genUniqID(ctx, flag)
		if err != nil {
			continue
		}

		if id != "" {
			break
		}
	}

	if id == "" {
		err = fmt.Errorf("[%s]: failed to generate unique id, err: %w", "GenUniqID", err)
	}

	return
}

//nolint:unparam
func (repo *ulidRepo) genUniqID(ctx context.Context, flag cconstant.UniqueIDFlag) (id string, err error) {
	_po := &UniqueID{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	id = cutil.UlidMake()
	_po.ID = id
	_po.Flag = flag

	_, err = sr.FromPo(_po).InsertStruct(_po)
	if err != nil {
		id = ""
	}

	return
}

func (repo *ulidRepo) DelUniqID(ctx context.Context, flag cconstant.UniqueIDFlag, id string) (err error) {
	_po := &UniqueID{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	_, err = sr.FromPo(_po).
		WhereEqual("f_id", id).
		WhereEqual("f_flag", flag).
		Delete()

	return
}
