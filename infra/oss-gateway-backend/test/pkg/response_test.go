package pkg_test

import (
	"net/http"
	"net/http/httptest"
	"oss-gateway/pkg/errors"
	"oss-gateway/pkg/response"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestSuccess(t *testing.T) {
	c, w := setupTestContext()

	data := map[string]string{
		"id":   "123",
		"name": "test",
	}

	response.Success(c, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"id\":\"123\"")
}

func TestSuccessWithCount(t *testing.T) {
	c, w := setupTestContext()

	data := []map[string]string{
		{"id": "1", "name": "item1"},
		{"id": "2", "name": "item2"},
	}

	response.SuccessWithCount(c, data, 2)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"count\":2")
}

func TestError(t *testing.T) {
	c, w := setupTestContext()

	response.Error(c, http.StatusBadRequest, "400001", "Invalid parameter",
		"The parameter is invalid", "Please check your input", "missing field")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "\"code\":\"400001\"")
	assert.Contains(t, w.Body.String(), "\"message\":\"Invalid parameter\"")
}

func TestBadRequest(t *testing.T) {
	c, w := setupTestContext()

	response.BadRequest(c, "Invalid JSON")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid JSON")
}

func TestNotFound(t *testing.T) {
	c, w := setupTestContext()

	response.NotFound(c, "Resource not found")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Resource not found")
}

func TestInternalError(t *testing.T) {
	c, w := setupTestContext()

	response.InternalError(c, "Database connection failed")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Database connection failed")
}

func TestInvalidParam(t *testing.T) {
	c, w := setupTestContext()

	response.InvalidParam(c, "storage_id")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "storage_id")
}

func TestStorageNotFound(t *testing.T) {
	c, w := setupTestContext()

	response.StorageNotFound(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), errors.StorageNotFound.Code)
}

func TestConnectionFailed(t *testing.T) {
	c, w := setupTestContext()

	response.ConnectionFailed(c, "timeout")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "timeout")
}

func TestStorageNameExist(t *testing.T) {
	c, w := setupTestContext()

	response.StorageNameExist(c, "my-storage")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "my-storage")
	assert.Contains(t, w.Body.String(), errors.StorageNameExists.Code)
}

func TestStorageExist(t *testing.T) {
	c, w := setupTestContext()

	description := "Bucket(test-bucket) with endpoint(https://oss.aliyuncs.com) already exists"
	response.StorageExist(c, description)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "test-bucket")
	assert.Contains(t, w.Body.String(), errors.StorageExists.Code)
}

func TestTooManyKeys(t *testing.T) {
	c, w := setupTestContext()

	response.TooManyKeys(c, 100)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "100")
	assert.Contains(t, w.Body.String(), errors.TooManyKeys.Code)
}

func TestInvalidVendorType(t *testing.T) {
	c, w := setupTestContext()

	response.InvalidVendorType(c, "INVALID_TYPE")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_TYPE")
	assert.Contains(t, w.Body.String(), errors.InvalidVendorType.Code)
}

func TestErrorWithCode(t *testing.T) {
	c, w := setupTestContext()

	response.ErrorWithCode(c, http.StatusBadRequest, &errors.BadRequest, "Custom message")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Custom message")
	assert.Contains(t, w.Body.String(), errors.BadRequest.Code)
}

func TestErrorWithCode_DefaultMessage(t *testing.T) {
	c, w := setupTestContext()

	response.ErrorWithCode(c, http.StatusBadRequest, &errors.BadRequest, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), errors.BadRequest.Description)
}

func TestSuccess_WithNilData(t *testing.T) {
	c, w := setupTestContext()

	response.Success(c, nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSuccess_WithEmptyData(t *testing.T) {
	c, w := setupTestContext()

	response.Success(c, map[string]string{})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "data")
}

func TestSuccessWithCount_ZeroCount(t *testing.T) {
	c, w := setupTestContext()

	response.SuccessWithCount(c, []interface{}{}, 0)

	assert.Equal(t, http.StatusOK, w.Code)
	// count=0时可能不显示在JSON中
}

func TestSuccessWithCount_LargeCount(t *testing.T) {
	c, w := setupTestContext()

	response.SuccessWithCount(c, []interface{}{}, 1000000)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"count\":1000000")
}
