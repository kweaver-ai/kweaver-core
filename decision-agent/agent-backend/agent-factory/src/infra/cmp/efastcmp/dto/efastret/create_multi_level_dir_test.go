package efastret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMultiLevelDirRsp_StructFields(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		DocId:    "doc-123",
		Rev:      "rev-456",
		Modified: 1234567890,
	}

	assert.Equal(t, "doc-123", rsp.DocId)
	assert.Equal(t, "rev-456", rsp.Rev)
	assert.Equal(t, int64(1234567890), rsp.Modified)
}

func TestCreateMultiLevelDirRsp_Empty(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{}

	assert.Empty(t, rsp.DocId)
	assert.Empty(t, rsp.Rev)
	assert.Equal(t, int64(0), rsp.Modified)
}

func TestCreateMultiLevelDirRsp_WithChineseDocId(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		DocId:    "文档-123",
		Rev:      "版本-456",
		Modified: 1234567890,
	}

	assert.Equal(t, "文档-123", rsp.DocId)
	assert.Equal(t, "版本-456", rsp.Rev)
}

func TestCreateMultiLevelDirRsp_WithDifferentModifiedTimes(t *testing.T) {
	t.Parallel()

	modifiedTimes := []int64{0, 1234567890, 9876543210, 1640000000}

	for _, modified := range modifiedTimes {
		rsp := CreateMultiLevelDirRsp{
			Modified: modified,
		}
		assert.Equal(t, modified, rsp.Modified)
	}
}

func TestCreateMultiLevelDirRsp_WithSpecialCharactersInRev(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		DocId:    "doc-123",
		Rev:      "rev-@#$%",
		Modified: 1234567890,
	}

	assert.Equal(t, "rev-@#$%", rsp.Rev)
}

func TestCreateMultiLevelDirRsp_WithOnlyDocId(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		DocId: "doc-123",
	}

	assert.Equal(t, "doc-123", rsp.DocId)
	assert.Empty(t, rsp.Rev)
	assert.Equal(t, int64(0), rsp.Modified)
}

func TestCreateMultiLevelDirRsp_WithOnlyRev(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		Rev: "rev-456",
	}

	assert.Empty(t, rsp.DocId)
	assert.Equal(t, "rev-456", rsp.Rev)
	assert.Equal(t, int64(0), rsp.Modified)
}

func TestCreateMultiLevelDirRsp_WithOnlyModified(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		Modified: 1234567890,
	}

	assert.Empty(t, rsp.DocId)
	assert.Empty(t, rsp.Rev)
	assert.Equal(t, int64(1234567890), rsp.Modified)
}

func TestCreateMultiLevelDirRsp_WithLargeModified(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		Modified: 9999999999,
	}

	assert.Equal(t, int64(9999999999), rsp.Modified)
}

func TestCreateMultiLevelDirRsp_WithNegativeModified(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		Modified: -1,
	}

	assert.Equal(t, int64(-1), rsp.Modified)
}

func TestCreateMultiLevelDirRsp_AllFieldsSet(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		DocId:    "test-doc",
		Rev:      "test-rev",
		Modified: 1234567890,
	}

	assert.Equal(t, "test-doc", rsp.DocId)
	assert.Equal(t, "test-rev", rsp.Rev)
	assert.Equal(t, int64(1234567890), rsp.Modified)
}

func TestCreateMultiLevelDirRsp_WithEmptyDocId(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		DocId:    "",
		Rev:      "rev-456",
		Modified: 1234567890,
	}

	assert.Empty(t, rsp.DocId)
	assert.Equal(t, "rev-456", rsp.Rev)
}

func TestCreateMultiLevelDirRsp_WithEmptyRev(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		DocId:    "doc-123",
		Rev:      "",
		Modified: 1234567890,
	}

	assert.Equal(t, "doc-123", rsp.DocId)
	assert.Empty(t, rsp.Rev)
}

func TestCreateMultiLevelDirRsp_WithZeroModified(t *testing.T) {
	t.Parallel()

	rsp := CreateMultiLevelDirRsp{
		DocId:    "doc-123",
		Rev:      "rev-456",
		Modified: 0,
	}

	assert.Equal(t, "doc-123", rsp.DocId)
	assert.Equal(t, "rev-456", rsp.Rev)
	assert.Equal(t, int64(0), rsp.Modified)
}
