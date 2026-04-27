package dbaulid

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/bwmarrin/snowflake"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
)

var node *snowflake.Node

func init() {
	node, _ = snowflake.NewNode(0)
}

// 雪花ID统一用19位数字转字符串
func (repo *ulidRepo) GenSnowID(ctx context.Context) (id string, err error) {
	return strconv.FormatInt(node.Generate().Int64(), 10), nil
}

// GenDBID 生成一个ID
func (repo *ulidRepo) GenDBID(ctx context.Context, tx *sql.Tx) (id string, err error) {
	maxRetry := 5
	for i := 0; i < maxRetry; i++ {
		id, err = repo.genDBID(ctx, tx)
		if err != nil {
			continue
		}

		if id != "" {
			break
		}
	}

	if id == "" {
		err = fmt.Errorf("[%s]: failed to generate unique id, err: %w", "GenDBID", err)
	}

	return
}

//nolint:unparam
func (repo *ulidRepo) genDBID(ctx context.Context, tx *sql.Tx) (id string, err error) {
	_po := &UniqueID{}
	sr := dbhelper2.TxSr(tx, repo.logger)

	id = cutil.UlidMake()
	_po.ID = id
	_po.Flag = cconstant.UniqueIDFlagDB

	_, err = sr.FromPo(_po).InsertStruct(_po)
	if err != nil {
		id = ""
	}

	return
}

// BatchGenDBID 批量生成数据库ID
func (repo *ulidRepo) BatchGenDBID(ctx context.Context, tx *sql.Tx, num int) (ids []string, err error) {
	maxRetry := 5

	for i := 0; i < maxRetry; i++ {
		ids, err = repo.batchGenDBID(ctx, tx, num)
		if err != nil {
			continue
		}

		return
	}

	if err != nil {
		err = fmt.Errorf("[%s]: failed to batch generate unique id, err: %w", "BatchGenDBID", err)
	}

	return
}

func (repo *ulidRepo) batchGenDBID(ctx context.Context, tx *sql.Tx, num int) (ids []string, err error) {
	maxPerSize := 500

	defer func() {
		if err != nil {
			ids = nil
		}
	}()

	for {
		if num > maxPerSize {
			var _ids []string

			_ids, err = repo.doBatchGenID(ctx, tx, maxPerSize)
			if err != nil {
				return
			}

			num -= maxPerSize

			ids = append(ids, _ids...)
		} else {
			var _ids []string

			_ids, err = repo.doBatchGenID(ctx, tx, num)
			if err != nil {
				return
			}

			ids = append(ids, _ids...)

			return
		}
	}
}

//nolint:unparam
func (repo *ulidRepo) doBatchGenID(ctx context.Context, tx *sql.Tx, num int) (ids []string, err error) {
	_po := &UniqueID{}

	sr := dbhelper2.TxSr(tx, repo.logger)

	ids = make([]string, num)
	pos := make([]UniqueID, num)

	for i := 0; i < num; i++ {
		ids[i] = cutil.UlidMake()
		pos[i].ID = ids[i]
		pos[i].Flag = cconstant.UniqueIDFlagDB
	}

	_, err = sr.FromPo(_po).InsertStructs(pos)
	if err != nil {
		ids = make([]string, 0)
	}

	return
}
