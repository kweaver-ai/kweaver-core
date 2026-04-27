package agentconfigreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestTmpReq_StructFields(t *testing.T) {
	t.Parallel()

	req := TestTmpReq{
		TestFlag: "test-flag-123",
		Params:   map[string]interface{}{"key": "value"},
	}

	assert.Equal(t, "test-flag-123", req.TestFlag)
	assert.NotNil(t, req.Params)
}

func TestTestTmpReq_Empty(t *testing.T) {
	t.Parallel()

	req := TestTmpReq{}

	assert.Empty(t, req.TestFlag)
	assert.Nil(t, req.Params)
}

func TestTestTmpReq_WithOnlyTestFlag(t *testing.T) {
	t.Parallel()

	req := TestTmpReq{
		TestFlag: "test-flag",
	}

	assert.Equal(t, "test-flag", req.TestFlag)
	assert.Nil(t, req.Params)
}

func TestTestTmpReq_WithOnlyParams(t *testing.T) {
	t.Parallel()

	params := map[string]interface{}{"key": "value"}
	req := TestTmpReq{
		Params: params,
	}

	assert.Empty(t, req.TestFlag)
	assert.Equal(t, params, req.Params)
}

func TestTestTmpReq_WithChineseTestFlag(t *testing.T) {
	t.Parallel()

	req := TestTmpReq{
		TestFlag: "测试标志",
		Params:   map[string]interface{}{"键": "值"},
	}

	assert.Equal(t, "测试标志", req.TestFlag)
}

func TestTestTmpReq_WithNilParams(t *testing.T) {
	t.Parallel()

	req := TestTmpReq{
		TestFlag: "test-flag",
		Params:   nil,
	}

	assert.Equal(t, "test-flag", req.TestFlag)
	assert.Nil(t, req.Params)
}

func TestTestTmpReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := TestTmpReq{}
	errMap := req.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Len(t, errMap, 2)
	assert.Contains(t, errMap, "TestFlag.required")
	assert.Contains(t, errMap, "Params.required")
	assert.Equal(t, "test_flag is required", errMap["TestFlag.required"])
	assert.Equal(t, "params is required", errMap["Params.required"])
}

func TestTestTmpReq_WithDifferentParamsTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params interface{}
	}{
		{
			name:   "string params",
			params: "test-string",
		},
		{
			name:   "map params",
			params: map[string]interface{}{"key": "value"},
		},
		{
			name:   "slice params",
			params: []string{"item1", "item2"},
		},
		{
			name:   "int params",
			params: 123,
		},
		{
			name:   "nil params",
			params: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := TestTmpReq{
				TestFlag: "test-flag",
				Params:   tt.params,
			}
			assert.Equal(t, tt.params, req.Params)
		})
	}
}

func TestTestTmpReq_WithComplexParams(t *testing.T) {
	t.Parallel()

	params := map[string]interface{}{
		"string": "value",
		"number": 123,
		"bool":   true,
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	req := TestTmpReq{
		TestFlag: "test-flag",
		Params:   params,
	}

	assert.Equal(t, params, req.Params)
	assert.Equal(t, "value", req.Params.(map[string]interface{})["string"])
	assert.Equal(t, 123, req.Params.(map[string]interface{})["number"])
}

func TestTestTmpReq_WithEmptyTestFlag(t *testing.T) {
	t.Parallel()

	req := TestTmpReq{
		TestFlag: "",
		Params:   map[string]interface{}{},
	}

	assert.Empty(t, req.TestFlag)
	assert.NotNil(t, req.Params)
}

func TestTestTmpReq_WithEmptyMapParams(t *testing.T) {
	t.Parallel()

	req := TestTmpReq{
		TestFlag: "test-flag",
		Params:   map[string]interface{}{},
	}

	assert.Equal(t, "test-flag", req.TestFlag)
	assert.NotNil(t, req.Params)
	assert.Empty(t, req.Params.(map[string]interface{}))
}

func TestTestTmpReq_GetErrMsgMap_Keys(t *testing.T) {
	t.Parallel()

	req := TestTmpReq{}
	errMap := req.GetErrMsgMap()

	assert.Equal(t, "test_flag is required", errMap["TestFlag.required"])
	assert.Equal(t, "params is required", errMap["Params.required"])
}

func TestTestTmpReq_AllFieldsSet(t *testing.T) {
	t.Parallel()

	req := TestTmpReq{
		TestFlag: "test-flag-123",
		Params:   map[string]interface{}{"key": "value"},
	}

	assert.Equal(t, "test-flag-123", req.TestFlag)
	assert.NotNil(t, req.Params)
	assert.Equal(t, "value", req.Params.(map[string]interface{})["key"])
}
