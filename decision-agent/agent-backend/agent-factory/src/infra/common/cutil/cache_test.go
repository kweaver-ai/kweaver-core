package cutil

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string
	Value int
}

// UnmarshalableStruct is a struct that cannot be marshaled to JSON
type UnmarshalableStruct struct {
	Func func()
}

func TestSetCache(t *testing.T) {
	t.Parallel()

	db, mock := redismock.NewClientMock()
	ctx := context.Background()

	testCases := []struct {
		name    string
		key     string
		value   any
		expire  time.Duration
		wantErr bool
		setup   func()
	}{
		{
			name:   "正常设置缓存",
			key:    "test_key",
			value:  TestStruct{Name: "test", Value: 123},
			expire: time.Hour,
			setup: func() {
				data, _ := json.Marshal(TestStruct{Name: "test", Value: 123})
				mock.ExpectSet("test_key", data, time.Hour).SetVal("OK")
			},
		},
		{
			name:    "Redis设置缓存错误",
			key:     "error_key",
			value:   TestStruct{Name: "test", Value: 123},
			expire:  time.Hour,
			wantErr: true,
			setup: func() {
				data, _ := json.Marshal(TestStruct{Name: "test", Value: 123})
				mock.ExpectSet("error_key", data, time.Hour).SetErr(errors.New("redis error"))
			},
		},
		{
			name:    "JSON序列化错误",
			key:     "marshal_error_key",
			value:   UnmarshalableStruct{Func: func() {}},
			expire:  time.Hour,
			wantErr: true,
			setup: func() {
				// No Redis expectation needed since marshal fails before Redis call
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.setup()

			err := SetCache(ctx, db, tc.key, tc.value, tc.expire)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetCache(t *testing.T) {
	t.Parallel()

	db, mock := redismock.NewClientMock()
	ctx := context.Background()

	testCases := []struct {
		name     string
		key      string
		mockData TestStruct
		wantErr  bool
		setup    func()
	}{
		{
			name: "正常获取缓存",
			key:  "test_key",
			mockData: TestStruct{
				Name:  "test",
				Value: 123,
			},
			setup: func() {
				data, _ := json.Marshal(TestStruct{Name: "test", Value: 123})
				mock.ExpectGet("test_key").SetVal(string(data))
			},
		},
		{
			name:    "键不存在",
			key:     "not_exist_key",
			wantErr: false,
			setup: func() {
				mock.ExpectGet("not_exist_key").RedisNil()
			},
		},
		{
			name:    "Redis获取错误",
			key:     "error_key",
			wantErr: true,
			setup: func() {
				mock.ExpectGet("error_key").SetErr(errors.New("redis error"))
			},
		},
		{
			name:    "JSON反序列化错误",
			key:     "invalid_json_key",
			wantErr: true,
			setup: func() {
				mock.ExpectGet("invalid_json_key").SetVal("invalid json")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.setup()

			result, err := GetCache[TestStruct](ctx, db, tc.key)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tc.key == "not_exist_key" {
					assert.Nil(t, result)
				} else {
					assert.Equal(t, &tc.mockData, result)
				}
			}
		})
	}
}

func TestDelCache(t *testing.T) {
	t.Parallel()

	db, mock := redismock.NewClientMock()
	ctx := context.Background()

	testCases := []struct {
		name    string
		key     string
		wantErr bool
		setup   func()
	}{
		{
			name: "正常删除缓存",
			key:  "test_key",
			setup: func() {
				mock.ExpectDel("test_key").SetVal(1)
			},
		},
		{
			name:    "Redis删除错误",
			key:     "error_key",
			wantErr: true,
			setup: func() {
				mock.ExpectDel("error_key").SetErr(errors.New("redis error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.setup()

			err := DelCache(ctx, db, tc.key)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
