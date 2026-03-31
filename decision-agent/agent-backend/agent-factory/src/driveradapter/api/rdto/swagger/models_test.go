package swagger

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIError_StructFields(t *testing.T) {
	t.Parallel()

	err := APIError{
		Code:    400,
		Message: "请求参数错误",
		Details: "详细错误信息",
	}

	assert.Equal(t, 400, err.Code)
	assert.Equal(t, "请求参数错误", err.Message)
	assert.Equal(t, "详细错误信息", err.Details)
}

func TestAPIError_Empty(t *testing.T) {
	t.Parallel()

	err := APIError{}

	assert.Equal(t, 0, err.Code)
	assert.Empty(t, err.Message)
	assert.Empty(t, err.Details)
}

func TestAPIError_WithoutDetails(t *testing.T) {
	t.Parallel()

	err := APIError{
		Code:    404,
		Message: "未找到资源",
	}

	assert.Equal(t, 404, err.Code)
	assert.Equal(t, "未找到资源", err.Message)
	assert.Empty(t, err.Details)
}

func TestAPIError_DifferentCodes(t *testing.T) {
	t.Parallel()

	codes := []int{
		200, 201, 204, // Success
		400, 401, 403, 404, // Client errors
		500, 502, 503, // Server errors
	}

	for _, code := range codes {
		err := APIError{
			Code:    code,
			Message: "Test message",
		}

		assert.Equal(t, code, err.Code)
	}
}

func TestAPIError_JSONTags(t *testing.T) {
	t.Parallel()

	apiErr := APIError{
		Code:    400,
		Message: "请求参数错误",
		Details: "详细错误信息",
	}

	// Marshal to JSON
	data, err := json.Marshal(apiErr)
	require.NoError(t, err)

	// Unmarshal to map to check JSON tags
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, float64(400), result["code"])
	assert.Equal(t, "请求参数错误", result["message"])
	assert.Equal(t, "详细错误信息", result["details"])
}

func TestAPIError_WithChineseMessage(t *testing.T) {
	t.Parallel()

	err := APIError{
		Code:    500,
		Message: "服务器内部错误",
		Details: "请稍后重试",
	}

	assert.Equal(t, "服务器内部错误", err.Message)
	assert.Equal(t, "请稍后重试", err.Details)
}

func TestAPIError_LongMessages(t *testing.T) {
	t.Parallel()

	longMessage := "这是一个非常长的错误消息，包含了很多详细的描述信息，用于测试长文本的处理能力"
	longDetails := "详细错误信息：" + longMessage

	err := APIError{
		Code:    400,
		Message: longMessage,
		Details: longDetails,
	}

	assert.Equal(t, longMessage, err.Message)
	assert.Equal(t, longDetails, err.Details)
	assert.Contains(t, err.Details, "详细错误信息")
}

func TestPaginatedResponse_StructFields(t *testing.T) {
	t.Parallel()

	list := []string{"item1", "item2", "item3"}
	resp := PaginatedResponse{
		Total: 100,
		List:  list,
	}

	assert.Equal(t, int64(100), resp.Total)
	assert.NotNil(t, resp.List)
	assert.Len(t, resp.List, 3)
}

func TestPaginatedResponse_Empty(t *testing.T) {
	t.Parallel()

	resp := PaginatedResponse{
		Total: 0,
		List:  []string{},
	}

	assert.Equal(t, int64(0), resp.Total)
	assert.Empty(t, resp.List)
}

func TestPaginatedResponse_NilList(t *testing.T) {
	t.Parallel()

	resp := PaginatedResponse{
		Total: 50,
		List:  nil,
	}

	assert.Equal(t, int64(50), resp.Total)
	assert.Nil(t, resp.List)
}

func TestPaginatedResponse_WithDifferentTypes(t *testing.T) {
	t.Parallel()

	// Test with string slice
	stringList := []string{"a", "b", "c"}
	resp1 := PaginatedResponse{
		Total: 3,
		List:  stringList,
	}
	assert.Equal(t, int64(3), resp1.Total)
	assert.Len(t, resp1.List, 3)

	// Test with int slice
	intList := []int{1, 2, 3, 4, 5}
	resp2 := PaginatedResponse{
		Total: 5,
		List:  intList,
	}
	assert.Equal(t, int64(5), resp2.Total)
	assert.Len(t, resp2.List, 5)

	// Test with map slice
	mapList := []map[string]interface{}{
		{"id": 1, "name": "item1"},
		{"id": 2, "name": "item2"},
	}
	resp3 := PaginatedResponse{
		Total: 2,
		List:  mapList,
	}
	assert.Equal(t, int64(2), resp3.Total)
	assert.Len(t, resp3.List, 2)
}

func TestPaginatedResponse_LargeTotal(t *testing.T) {
	t.Parallel()

	resp := PaginatedResponse{
		Total: 999999999,
		List:  []string{"item"},
	}

	assert.Equal(t, int64(999999999), resp.Total)
}

func TestPaginatedResponse_WithEmptyList(t *testing.T) {
	t.Parallel()

	resp := PaginatedResponse{
		Total: 100,
		List:  []interface{}{},
	}

	assert.Equal(t, int64(100), resp.Total)
	assert.Empty(t, resp.List)
	assert.NotNil(t, resp.List)
}

func TestPaginatedResponse_JSONTags(t *testing.T) {
	t.Parallel()

	list := []string{"item1", "item2"}
	resp := PaginatedResponse{
		Total: 2,
		List:  list,
	}

	// Marshal to JSON
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	// Unmarshal to map to check JSON tags
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, float64(2), result["total"])
	assert.Contains(t, result, "list")
}

func TestAPIError_CommonErrors(t *testing.T) {
	t.Parallel()

	commonErrors := []struct {
		code    int
		message string
	}{
		{400, "请求参数错误"},
		{401, "未授权"},
		{403, "禁止访问"},
		{404, "未找到资源"},
		{500, "服务器内部错误"},
		{503, "服务不可用"},
	}

	for _, ce := range commonErrors {
		err := APIError{
			Code:    ce.code,
			Message: ce.message,
		}

		assert.Equal(t, ce.code, err.Code)
		assert.Equal(t, ce.message, err.Message)
	}
}
