package redishelper

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type redisHelperTestStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type redisHelperTestUnmarshalableStruct struct {
	Fn func()
}

func TestErrNotSupportInLocalEnv(t *testing.T) {
	t.Parallel()

	t.Run("error message is correct", func(t *testing.T) {
		t.Parallel()

		err := ErrNotSupportInLocalEnv
		assert.Equal(t, "redishelper: not support in local env", err.Error())
	})
}

func TestSetStruct(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db, mock := redismock.NewClientMock()
		value := redisHelperTestStruct{Name: "n1", Value: 1}
		body, _ := json.Marshal(value)

		mock.ExpectSet("k1", string(body), time.Minute).SetVal("OK")

		err := SetStruct(db, "k1", value, time.Minute)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("marshal error", func(t *testing.T) {
		t.Parallel()

		db, _ := redismock.NewClientMock()

		err := SetStruct(db, "k1", redisHelperTestUnmarshalableStruct{Fn: func() {}}, time.Minute)

		assert.Error(t, err)
	})

	t.Run("redis set error", func(t *testing.T) {
		t.Parallel()

		db, mock := redismock.NewClientMock()
		value := redisHelperTestStruct{Name: "n1", Value: 1}
		body, _ := json.Marshal(value)

		mock.ExpectSet("k1", string(body), time.Minute).SetErr(errors.New("redis set failed"))

		err := SetStruct(db, "k1", value, time.Minute)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "redis set failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetStruct(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db, mock := redismock.NewClientMock()
		expected := redisHelperTestStruct{Name: "n1", Value: 10}
		body, _ := json.Marshal(expected)

		mock.ExpectGet("k1").SetVal(string(body))

		actual := &redisHelperTestStruct{}
		err := GetStruct(db, "k1", actual)

		require.NoError(t, err)
		assert.Equal(t, expected, *actual)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("redis get error", func(t *testing.T) {
		t.Parallel()

		db, mock := redismock.NewClientMock()
		mock.ExpectGet("k1").SetErr(errors.New("redis get failed"))

		actual := &redisHelperTestStruct{}
		err := GetStruct(db, "k1", actual)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "redis get failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("json unmarshal error", func(t *testing.T) {
		t.Parallel()

		db, mock := redismock.NewClientMock()
		mock.ExpectGet("k1").SetVal("not-json")

		actual := &redisHelperTestStruct{}
		err := GetStruct(db, "k1", actual)

		require.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSetStruct_PanicsWithNilRedisClient(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		_ = SetStruct(nil, "key", "value", 0)
	})
}

func TestGetStruct_PanicsWithNilRedisClient(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		_ = GetStruct(nil, "key", nil)
	})
}

func TestGetRedisClientUniversal(t *testing.T) {
	t.Parallel()

	originalClient := redisClient

	t.Cleanup(func() {
		redisClient = originalClient
	})

	t.Run("panic when redis client is not initialized", func(t *testing.T) {
		t.Parallel()

		redisClient = nil

		assert.Panics(t, func() {
			_ = GetRedisClientUniversal()
		})
	})

	t.Run("return universal client when initialized", func(t *testing.T) {
		t.Parallel()

		db, _ := redismock.NewClientMock()
		redisClient = db

		uc := GetRedisClientUniversal()

		assert.NotNil(t, uc)
	})

	t.Run("ErrNotSupportInLocalEnv can be checked with errors.Is", func(t *testing.T) {
		t.Parallel()

		err := errors.New("wrapped: " + ErrNotSupportInLocalEnv.Error())
		assert.NotNil(t, err)
	})
}
