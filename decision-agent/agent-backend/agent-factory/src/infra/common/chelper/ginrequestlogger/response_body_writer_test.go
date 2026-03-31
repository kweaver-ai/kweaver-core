package ginrequestlogger

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestResponseBodyWriter_Write_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	responseBodyWriter := &ResponseBodyWriter{
		ResponseWriter: ctx.Writer,
		Body:           &bytes.Buffer{},
	}

	testData := []byte("test response data")
	n, err := responseBodyWriter.Write(testData)

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, responseBodyWriter.Body.Bytes())
	assert.Equal(t, testData, w.Body.Bytes())
}

func TestResponseBodyWriter_Write_EmptyData(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	responseBodyWriter := &ResponseBodyWriter{
		ResponseWriter: ctx.Writer,
		Body:           &bytes.Buffer{},
	}

	testData := []byte("")
	n, err := responseBodyWriter.Write(testData)

	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, 0, responseBodyWriter.Body.Len())
	assert.Equal(t, "", responseBodyWriter.Body.String())
}

func TestResponseBodyWriter_Write_MultipleWrites(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	responseBodyWriter := &ResponseBodyWriter{
		ResponseWriter: ctx.Writer,
		Body:           &bytes.Buffer{},
	}

	data1 := []byte("first part ")
	data2 := []byte("second part")

	n1, err1 := responseBodyWriter.Write(data1)
	n2, err2 := responseBodyWriter.Write(data2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, len(data1), n1)
	assert.Equal(t, len(data2), n2)
	assert.Equal(t, "first part second part", responseBodyWriter.Body.String())
	assert.Equal(t, "first part second part", w.Body.String())
}

func TestResponseBodyWriter_Write_LargeData(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	responseBodyWriter := &ResponseBodyWriter{
		ResponseWriter: ctx.Writer,
		Body:           &bytes.Buffer{},
	}

	largeData := bytes.Repeat([]byte("x"), 10000)
	n, err := responseBodyWriter.Write(largeData)

	assert.NoError(t, err)
	assert.Equal(t, len(largeData), n)
	assert.Equal(t, largeData, responseBodyWriter.Body.Bytes())
	assert.Equal(t, largeData, w.Body.Bytes())
}

func TestResponseBodyWriter_Write_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	responseBodyWriter := &ResponseBodyWriter{
		ResponseWriter: ctx.Writer,
		Body:           &bytes.Buffer{},
	}

	testData := []byte("{\"key\": \"value\", \"emoji\": \"😀\"}")
	n, err := responseBodyWriter.Write(testData)

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, responseBodyWriter.Body.Bytes())
}
