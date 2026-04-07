package efastarg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMultiLevelDirReq_StructFields(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "folder1/folder2/folder3",
	}

	assert.Equal(t, "doc-123", req.DocId)
	assert.Equal(t, "folder1/folder2/folder3", req.Path)
}

func TestCreateMultiLevelDirReq_Empty(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{}

	assert.Empty(t, req.DocId)
	assert.Empty(t, req.Path)
}

func TestCreateMultiLevelDirReq_WithChinesePath(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "文档-123",
		Path:  "文件夹1/文件夹2/文件夹3",
	}

	assert.Equal(t, "文档-123", req.DocId)
	assert.Equal(t, "文件夹1/文件夹2/文件夹3", req.Path)
}

func TestCreateMultiLevelDirReq_WithSingleLevel(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "single-folder",
	}

	assert.Equal(t, "single-folder", req.Path)
}

func TestCreateMultiLevelDirReq_WithTwoLevels(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "folder1/folder2",
	}

	assert.Equal(t, "folder1/folder2", req.Path)
}

func TestCreateMultiLevelDirReq_WithMultipleLevels(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "level1/level2/level3/level4/level5",
	}

	assert.Equal(t, "level1/level2/level3/level4/level5", req.Path)
}

func TestCreateMultiLevelDirReq_WithSpecialCharactersInPath(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "folder@#$%/folder-123",
	}

	assert.Equal(t, "folder@#$%/folder-123", req.Path)
}

func TestCreateMultiLevelDirReq_WithSpacesInPath(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "folder 1/folder 2/folder 3",
	}

	assert.Equal(t, "folder 1/folder 2/folder 3", req.Path)
}

func TestCreateMultiLevelDirReq_WithOnlyDocId(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
	}

	assert.Equal(t, "doc-123", req.DocId)
	assert.Empty(t, req.Path)
}

func TestCreateMultiLevelDirReq_WithOnlyPath(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		Path: "folder1/folder2",
	}

	assert.Empty(t, req.DocId)
	assert.Equal(t, "folder1/folder2", req.Path)
}

func TestCreateMultiLevelDirReq_WithEmptyDocId(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "",
		Path:  "folder1/folder2",
	}

	assert.Empty(t, req.DocId)
	assert.Equal(t, "folder1/folder2", req.Path)
}

func TestCreateMultiLevelDirReq_WithEmptyPath(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "",
	}

	assert.Equal(t, "doc-123", req.DocId)
	assert.Empty(t, req.Path)
}

func TestCreateMultiLevelDirReq_WithBackwardSlashPath(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "folder1\\folder2\\folder3",
	}

	assert.Equal(t, "folder1\\folder2\\folder3", req.Path)
}

func TestCreateMultiLevelDirReq_WithMixedSlashPath(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "folder1/folder2\\folder3",
	}

	assert.Equal(t, "folder1/folder2\\folder3", req.Path)
}

func TestCreateMultiLevelDirReq_WithLeadingSlash(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "/folder1/folder2",
	}

	assert.Equal(t, "/folder1/folder2", req.Path)
}

func TestCreateMultiLevelDirReq_WithTrailingSlash(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "folder1/folder2/",
	}

	assert.Equal(t, "folder1/folder2/", req.Path)
}

func TestCreateMultiLevelDirReq_WithDotsInPath(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "folder1/./folder2/../folder3",
	}

	assert.Equal(t, "folder1/./folder2/../folder3", req.Path)
}

func TestCreateMultiLevelDirReq_AllFieldsSet(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "test-doc",
		Path:  "test/path",
	}

	assert.Equal(t, "test-doc", req.DocId)
	assert.Equal(t, "test/path", req.Path)
}

func TestCreateMultiLevelDirReq_WithUnderscoresInPath(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "folder_1/folder_2/folder_3",
	}

	assert.Equal(t, "folder_1/folder_2/folder_3", req.Path)
}

func TestCreateMultiLevelDirReq_WithHyphensInPath(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "folder-1/folder-2/folder-3",
	}

	assert.Equal(t, "folder-1/folder-2/folder-3", req.Path)
}

func TestCreateMultiLevelDirReq_WithNumbersInPath(t *testing.T) {
	t.Parallel()

	req := CreateMultiLevelDirReq{
		DocId: "doc-123",
		Path:  "folder1/folder2/folder3",
	}

	assert.Equal(t, "folder1/folder2/folder3", req.Path)
}
