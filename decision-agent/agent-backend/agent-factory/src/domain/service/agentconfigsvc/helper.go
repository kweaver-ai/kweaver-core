package v3agentconfigsvc

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

func (s *dataAgentConfigSvc) getTx(ctx context.Context) (tx *sql.Tx, err error) {
	tx, err = s.agentTplRepo.BeginTx(ctx)
	if err != nil {
		err = errors.Wrap(err, "开启事务失败")
		return
	}

	return
}
