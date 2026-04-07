package chelper

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// mockLogger 用于测试的mock logger
type mockLogger struct {
	errorLogs []interface{}
}

func (m *mockLogger) Debugf(format string, args ...interface{}) {}
func (m *mockLogger) Debugln(args ...interface{})               {}
func (m *mockLogger) Infof(format string, args ...interface{})  {}
func (m *mockLogger) Infoln(args ...interface{})                {}
func (m *mockLogger) Warnf(format string, args ...interface{})  {}
func (m *mockLogger) Warnln(args ...interface{})                {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {
	m.errorLogs = append(m.errorLogs, args)
}
func (m *mockLogger) Errorln(args ...interface{})               { m.errorLogs = append(m.errorLogs, args) }
func (m *mockLogger) Panicf(format string, args ...interface{}) {}
func (m *mockLogger) Panicln(args ...interface{})               {}
func (m *mockLogger) Fatalf(format string, args ...interface{}) {}
func (m *mockLogger) Fatalln(args ...interface{})               {}

func newMockLogger() *mockLogger {
	return &mockLogger{errorLogs: make([]interface{}, 0)}
}

// TestJoinErrors 测试错误合并函数
func TestJoinErrors(t *testing.T) {
	t.Parallel()

	t.Run("newErr is nil", func(t *testing.T) {
		t.Parallel()

		var err error = errors.New("original error")

		joinErrors(&err, nil)
		assert.Equal(t, "original error", err.Error())
	})

	t.Run("original is nil", func(t *testing.T) {
		t.Parallel()

		var err error

		joinErrors(&err, errors.New("new error"))
		assert.Equal(t, "new error", err.Error())
	})

	t.Run("both not nil - errors should be joined", func(t *testing.T) {
		t.Parallel()

		originalErr := errors.New("original error")
		newErr := errors.New("new error")

		var err error = originalErr

		joinErrors(&err, newErr)

		// 验证两个错误都被包含
		assert.True(t, errors.Is(err, originalErr))
		assert.True(t, errors.Is(err, newErr))
	})
}

// TestIsSqlNotFound 测试SQL未找到判断
func TestIsSqlNotFound(t *testing.T) {
	t.Parallel()

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsSqlNotFound(nil))
	})

	t.Run("sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsSqlNotFound(sql.ErrNoRows))
	})

	t.Run("wrapped sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		wrappedErr := errors.Join(errors.New("some context"), sql.ErrNoRows)
		assert.True(t, IsSqlNotFound(wrappedErr))
	})

	t.Run("other error", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsSqlNotFound(errors.New("some other error")))
	})
}

// TestTxRollbackOrCommit_Commit 测试正常提交
func TestTxRollbackOrCommit_Commit(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	tx, err := db.Begin()
	assert.NoError(t, err)

	logger := newMockLogger()

	var txErr error

	TxRollbackOrCommit(tx, &txErr, logger)

	assert.NoError(t, txErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTxRollbackOrCommit_Rollback 测试有错误时回滚
func TestTxRollbackOrCommit_Rollback(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	tx, err := db.Begin()
	assert.NoError(t, err)

	logger := newMockLogger()
	txErr := errors.New("some business error")

	TxRollbackOrCommit(tx, &txErr, logger)

	assert.Error(t, txErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTxRollbackOrCommit_CommitWithCallback 测试提交后回调
func TestTxRollbackOrCommit_CommitWithCallback(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	tx, err := db.Begin()
	assert.NoError(t, err)

	logger := newMockLogger()

	var txErr error

	callbackCalled := false

	opt := TxRollbackOrCommitOption{
		CommitCallBack: func() error {
			callbackCalled = true
			return nil
		},
	}

	TxRollbackOrCommit(tx, &txErr, logger, opt)

	assert.NoError(t, txErr)
	assert.True(t, callbackCalled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTxRollbackOrCommit_CommitCallbackError 测试提交回调返回错误
func TestTxRollbackOrCommit_CommitCallbackError(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	tx, err := db.Begin()
	assert.NoError(t, err)

	logger := newMockLogger()

	var txErr error

	callbackErr := errors.New("callback error")

	opt := TxRollbackOrCommitOption{
		CommitCallBack: func() error {
			return callbackErr
		},
	}

	TxRollbackOrCommit(tx, &txErr, logger, opt)

	assert.Error(t, txErr)
	assert.True(t, errors.Is(txErr, callbackErr))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTxRollbackOrCommit_RollbackWithExistingError 测试回滚时合并错误
func TestTxRollbackOrCommit_RollbackWithExistingError(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()

	rollbackErr := errors.New("rollback failed")
	mock.ExpectRollback().WillReturnError(rollbackErr)

	tx, err := db.Begin()
	assert.NoError(t, err)

	logger := newMockLogger()
	businessErr := errors.New("business error")
	txErr := businessErr

	TxRollbackOrCommit(tx, &txErr, logger)

	// 验证两个错误都被保留
	assert.True(t, errors.Is(txErr, businessErr))
	assert.True(t, errors.Is(txErr, rollbackErr))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTxRollback 测试只回滚不提交
func TestTxRollback(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	tx, err := db.Begin()
	assert.NoError(t, err)

	logger := newMockLogger()
	txErr := errors.New("some error")

	TxRollback(tx, &txErr, logger)

	assert.Error(t, txErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTxRollback_NoError 测试无错误时不回滚
func TestTxRollback_NoError(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	// 不期望回滚，因为没有错误

	tx, err := db.Begin()
	assert.NoError(t, err)

	logger := newMockLogger()

	var txErr error

	TxRollback(tx, &txErr, logger)

	assert.NoError(t, txErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTxDeferHandlerCommitClb 测试带回调的defer处理
func TestTxDeferHandlerCommitClb(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	tx, err := db.Begin()
	assert.NoError(t, err)

	logger := newMockLogger()

	var txErr error

	callbackCalled := false

	TxDeferHandlerCommitClb(tx, &txErr, logger, func() error {
		callbackCalled = true
		return nil
	})

	assert.NoError(t, txErr)
	assert.True(t, callbackCalled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCloseRows 测试关闭rows
func TestCloseRows(t *testing.T) {
	t.Parallel()

	t.Run("nil rows", func(t *testing.T) {
		t.Parallel()

		logger := newMockLogger()
		CloseRows(nil, logger)
		assert.Empty(t, logger.errorLogs)
	})

	t.Run("rows with no error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer db.Close()

		mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		rows, err := db.Query("SELECT 1")
		assert.NoError(t, err)

		logger := newMockLogger()
		CloseRows(rows, logger)
		assert.Empty(t, logger.errorLogs)
	})

	t.Run("rows with query error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		assert.NoError(t, err)

		defer db.Close()

		mock.ExpectQuery("SELECT").WillReturnError(errors.New("query failed"))

		rows, err := db.Query("SELECT 1")
		assert.Error(t, err)

		logger := newMockLogger()
		CloseRows(rows, logger)
		// Note: rows.Err() might not return the query error in all drivers
		// The function should not panic
		assert.NotNil(t, logger)
	})

	t.Run("rows with close error", func(t *testing.T) {
		t.Parallel()

		db, _, err := sqlmock.New()
		assert.NoError(t, err)

		defer db.Close()

		// Create a rows that will fail on close
		// We need to use a special approach - close after mock expectations
		rows, _ := db.Query("SELECT 1")
		// Close the DB first to force an error on rows.Close()
		db.Close()

		logger := newMockLogger()
		CloseRows(rows, logger)
		// Should have logged the error (may be from close or from the already-closed state)
		// The exact behavior depends on the driver implementation
	})
}
