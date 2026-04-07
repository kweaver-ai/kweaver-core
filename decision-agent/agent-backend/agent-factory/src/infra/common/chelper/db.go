package chelper

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
)

// joinErrors 合并错误，如果原错误不为nil，则将新错误与原错误合并
func joinErrors(original *error, newErr error) {
	if newErr == nil {
		return
	}

	if *original == nil {
		*original = newErr
	} else {
		*original = errors.Join(*original, newErr)
	}
}

type TxRollbackOrCommitOption struct {
	CommitCallBack func() error
}

// TxRollbackOrCommit 事务回滚或提交
//
//nolint:gocritic
func TxRollbackOrCommit(tx *sql.Tx, err *error, logger icmp.Logger, opt ...TxRollbackOrCommitOption) {
	re := recover()
	if *err != nil || re != nil {
		if re != nil {
			logger.Errorf("db panic: %v", re)

			// 如果是panic并且err为nil，将panic转换为error
			if *err == nil {
				//nolint:goerr113
				*err = fmt.Errorf("%v", re)
			}
		}

		joinErrors(err, tx.Rollback())
	} else {
		joinErrors(err, tx.Commit())

		if len(opt) > 0 && opt[0].CommitCallBack != nil {
			joinErrors(err, opt[0].CommitCallBack())
		}
	}
}

// TxRollback 回滚事务，不提交
func TxRollback(tx *sql.Tx, err *error, logger icmp.Logger) {
	re := recover()
	if *err != nil || re != nil {
		if re != nil {
			logger.Errorf("db panic: %v", re)

			// 如果是panic并且err为nil，将panic转换为error
			if *err == nil {
				//nolint:goerr113
				*err = fmt.Errorf("%v", re)
			}
		}

		joinErrors(err, tx.Rollback())
	}
}

func TxDeferHandlerCommitClb(tx *sql.Tx, err *error, logger icmp.Logger, commitCallBack func() error) {
	opt := TxRollbackOrCommitOption{
		CommitCallBack: commitCallBack,
	}

	TxRollbackOrCommit(tx, err, logger, opt)
}

func IsSqlNotFound(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, sql.ErrNoRows)
}

func CloseRows(rows *sql.Rows, logger icmp.Logger) {
	if rows != nil {
		if rowsErr := rows.Err(); rowsErr != nil {
			logger.Errorln(rowsErr)
		}

		if closeErr := rows.Close(); closeErr != nil {
			logger.Errorln(closeErr)
		}
	}
}
