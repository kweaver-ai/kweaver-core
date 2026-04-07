package dbacccom

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

// ==================== test helpers ====================

type testLogger struct{}

func (testLogger) Infof(string, ...interface{})  {}
func (testLogger) Infoln(...interface{})         {}
func (testLogger) Debugf(string, ...interface{}) {}
func (testLogger) Debugln(...interface{})        {}
func (testLogger) Errorf(string, ...interface{}) {}
func (testLogger) Errorln(...interface{})        {}
func (testLogger) Warnf(string, ...interface{})  {}
func (testLogger) Warnln(...interface{})         {}
func (testLogger) Panicf(string, ...interface{}) {}
func (testLogger) Panicln(...interface{})        {}
func (testLogger) Fatalf(string, ...interface{}) {}
func (testLogger) Fatalln(...interface{})        {}

type testTable struct{}

func (testTable) TableName() string { return "t_test" }

// ==================== NewUniqUlidHelper ====================

func TestNewUniqUlidHelper_Success(t *testing.T) {
	t.Parallel()

	db, _, err := sqlx.New()
	require.NoError(t, err)
	defer db.Close()

	h := NewUniqUlidHelper(&UniqUlidHelper{
		Po:     testTable{},
		DB:     db,
		Logger: testLogger{},
	})

	assert.NotNil(t, h)
	assert.Equal(t, "f_id", h.Pk) // default PK
}

func TestNewUniqUlidHelper_CustomPk(t *testing.T) {
	t.Parallel()

	db, _, err := sqlx.New()
	require.NoError(t, err)
	defer db.Close()

	h := NewUniqUlidHelper(&UniqUlidHelper{
		Po:     testTable{},
		Pk:     "f_uid",
		DB:     db,
		Logger: testLogger{},
	})

	assert.Equal(t, "f_uid", h.Pk)
}

func TestNewUniqUlidHelper_WithTx(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlx.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()

	tx, txErr := db.Begin()
	require.NoError(t, txErr)

	h := NewUniqUlidHelper(&UniqUlidHelper{
		Po:     testTable{},
		Tx:     tx,
		Logger: testLogger{},
	})

	assert.NotNil(t, h)
	assert.NotNil(t, h.Tx)
}

func TestNewUniqUlidHelper_PanicNilDto(t *testing.T) {
	t.Parallel()

	assert.PanicsWithValue(t, "[UniqUlidHelper][NewUniqUlidHelper]: dto is nil", func() {
		NewUniqUlidHelper(nil)
	})
}

func TestNewUniqUlidHelper_PanicNilPo(t *testing.T) {
	t.Parallel()

	db, _, err := sqlx.New()
	require.NoError(t, err)
	defer db.Close()

	assert.PanicsWithValue(t, "[UniqUlidHelper][NewUniqUlidHelper]: dto.Po is nil", func() {
		NewUniqUlidHelper(&UniqUlidHelper{
			DB:     db,
			Logger: testLogger{},
		})
	})
}

func TestNewUniqUlidHelper_PanicNilDBAndTx(t *testing.T) {
	t.Parallel()

	assert.PanicsWithValue(t, "[UniqUlidHelper][NewUniqUlidHelper]: dto.DB is nil and dto.Tx is nil", func() {
		NewUniqUlidHelper(&UniqUlidHelper{
			Po:     testTable{},
			Logger: testLogger{},
		})
	})
}

// ==================== GenDBID ====================

func TestGenDBID_Success_NotExists(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlx.New()
	require.NoError(t, err)
	defer db.Close()

	// ULID 不存在于表中 → Exists() 内部 SELECT 1 返回 sql.ErrNoRows
	mock.ExpectQuery(`(?i)select .* from t_test`).WillReturnError(sql.ErrNoRows)

	h := &UniqUlidHelper{
		Po:     testTable{},
		Pk:     "f_id",
		DB:     db,
		Logger: testLogger{},
	}

	id, genErr := h.GenDBID(context.Background())
	assert.NoError(t, genErr)
	assert.NotEmpty(t, id)
}

func TestGenDBID_Success_WithTx(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlx.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()

	tx, txErr := db.Begin()
	require.NoError(t, txErr)

	mock.ExpectQuery(`(?i)select .* from t_test`).WillReturnError(sql.ErrNoRows)

	h := &UniqUlidHelper{
		Po:     testTable{},
		Pk:     "f_id",
		Tx:     tx,
		Logger: testLogger{},
	}

	id, genErr := h.GenDBID(context.Background())
	assert.NoError(t, genErr)
	assert.NotEmpty(t, id)
}

func TestGenDBID_QueryError_ThenSuccess(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlx.New()
	require.NoError(t, err)
	defer db.Close()

	// 第一次查询返回真正的 DB 错误，第二次成功（不存在）
	mock.ExpectQuery(`(?i)select .* from t_test`).
		WillReturnError(errors.New("db connection error"))
	mock.ExpectQuery(`(?i)select .* from t_test`).WillReturnError(sql.ErrNoRows)

	h := &UniqUlidHelper{
		Po:     testTable{},
		Pk:     "f_id",
		DB:     db,
		Logger: testLogger{},
	}

	id, genErr := h.GenDBID(context.Background())
	assert.NoError(t, genErr)
	assert.NotEmpty(t, id)
}

func TestGenDBID_AllRetriesFail(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlx.New()
	require.NoError(t, err)
	defer db.Close()

	// 所有 50 次重试都失败
	for i := 0; i < 50; i++ {
		mock.ExpectQuery(`(?i)select .* from t_test`).
			WillReturnError(errors.New("db error"))
	}

	h := &UniqUlidHelper{
		Po:     testTable{},
		Pk:     "f_id",
		DB:     db,
		Logger: testLogger{},
	}

	id, genErr := h.GenDBID(context.Background())
	assert.Error(t, genErr)
	assert.Empty(t, id)
	assert.Contains(t, genErr.Error(), "failed to generate unique id")
}
