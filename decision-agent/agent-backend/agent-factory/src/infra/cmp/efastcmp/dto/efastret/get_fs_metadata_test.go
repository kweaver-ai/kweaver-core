package efastret

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestFsMetadata_StructFields(t *testing.T) {
	t.Parallel()

	metadata := FsMetadata{
		ID:         "file-123",
		Name:       "Test File",
		DocLibType: cenum.DocLibTypeStrPersonal,
		Path:       "/path/to/file",
		Size:       1024,
	}

	assert.Equal(t, "file-123", metadata.ID)
	assert.Equal(t, "Test File", metadata.Name)
	assert.Equal(t, cenum.DocLibTypeStrPersonal, metadata.DocLibType)
	assert.Equal(t, "/path/to/file", metadata.Path)
	assert.Equal(t, int64(1024), metadata.Size)
}

func TestFsMetadata_Empty(t *testing.T) {
	t.Parallel()

	metadata := FsMetadata{}

	assert.Empty(t, metadata.ID)
	assert.Empty(t, metadata.Name)
	assert.Equal(t, cenum.DocLibType(""), metadata.DocLibType)
	assert.Empty(t, metadata.Path)
	assert.Equal(t, int64(0), metadata.Size)
}

func TestFsMetadata_WithChineseName(t *testing.T) {
	t.Parallel()

	metadata := FsMetadata{
		ID:         "文件-123",
		Name:       "测试文件",
		DocLibType: cenum.DocLibTypeStrPersonal,
		Path:       "/路径/到/文件",
		Size:       2048,
	}

	assert.Equal(t, "文件-123", metadata.ID)
	assert.Equal(t, "测试文件", metadata.Name)
	assert.Equal(t, "/路径/到/文件", metadata.Path)
	assert.Equal(t, int64(2048), metadata.Size)
}

func TestFsMetadata_WithDocLibTypeDepartment(t *testing.T) {
	t.Parallel()

	metadata := FsMetadata{
		ID:         "folder-123",
		Name:       "Test Folder",
		DocLibType: cenum.DocLibTypeStrDepartment,
		Path:       "/path/to/folder",
		Size:       0,
	}

	assert.Equal(t, cenum.DocLibTypeStrDepartment, metadata.DocLibType)
	assert.Equal(t, int64(0), metadata.Size)
}

func TestFsMetadata_WithLargeSize(t *testing.T) {
	t.Parallel()

	sizes := []int64{0, 1024, 1024 * 1024, 1024 * 1024 * 1024}

	for _, size := range sizes {
		metadata := FsMetadata{
			ID:   "file-123",
			Size: size,
		}
		assert.Equal(t, size, metadata.Size)
	}
}

func TestFsMetadata_WithSpecialCharactersInPath(t *testing.T) {
	t.Parallel()

	metadata := FsMetadata{
		ID:   "file-123",
		Name: "Test File",
		Path: "/path/with spaces/to/file-@#$%.txt",
		Size: 512,
	}

	assert.Equal(t, "/path/with spaces/to/file-@#$%.txt", metadata.Path)
}

func TestGetFsMetadataRetDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataRetDto{
		&FsMetadata{
			ID:         "file-1",
			Name:       "File 1",
			DocLibType: cenum.DocLibTypeStrPersonal,
			Path:       "/path/file1",
			Size:       1024,
		},
		&FsMetadata{
			ID:         "file-2",
			Name:       "File 2",
			DocLibType: cenum.DocLibTypeStrPersonal,
			Path:       "/path/file2",
			Size:       2048,
		},
	}

	assert.Len(t, dto, 2)
	assert.Equal(t, "file-1", dto[0].ID)
	assert.Equal(t, "file-2", dto[1].ID)
}

func TestGetFsMetadataRetDto_Empty(t *testing.T) {
	t.Parallel()

	var dto GetFsMetadataRetDto

	assert.Nil(t, dto)
	assert.Len(t, dto, 0)
}

func TestGetFsMetadataRetDto_WithNil(t *testing.T) {
	t.Parallel()

	var dto GetFsMetadataRetDto

	assert.Nil(t, dto)
}

func TestGetFsMetadataRetDto_Append(t *testing.T) {
	t.Parallel()

	dto := make(GetFsMetadataRetDto, 0)

	dto = append(dto, &FsMetadata{
		ID:   "file-1",
		Name: "File 1",
	})
	dto = append(dto, &FsMetadata{
		ID:   "file-2",
		Name: "File 2",
	})

	assert.Len(t, dto, 2)
	assert.Equal(t, "file-1", dto[0].ID)
	assert.Equal(t, "file-2", dto[1].ID)
}

func TestGetFsMetadataRetDto_WithMultipleEntries(t *testing.T) {
	t.Parallel()

	dto := make(GetFsMetadataRetDto, 50)
	for i := 0; i < 50; i++ {
		dto[i] = &FsMetadata{
			ID:   "file-" + string(rune(i)),
			Name: "File " + string(rune(i)),
		}
	}

	assert.Len(t, dto, 50)
}

func TestGetFsMetadataRetDto_Iteration(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataRetDto{
		&FsMetadata{ID: "file-1", Name: "File 1"},
		&FsMetadata{ID: "file-2", Name: "File 2"},
		&FsMetadata{ID: "file-3", Name: "File 3"},
	}

	count := 0

	for _, metadata := range dto {
		assert.NotEmpty(t, metadata.ID)
		assert.NotEmpty(t, metadata.Name)

		count++
	}

	assert.Equal(t, 3, count)
}

func TestGetFsMetadataRetDto_SliceOperations(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataRetDto{
		&FsMetadata{ID: "file-1", Name: "File 1"},
		&FsMetadata{ID: "file-2", Name: "File 2"},
		&FsMetadata{ID: "file-3", Name: "File 3"},
	}

	// Test slicing
	subDto := dto[1:3]
	assert.Len(t, subDto, 2)
	assert.Equal(t, "file-2", subDto[0].ID)
}

func TestFsMetadata_AllFieldsSet(t *testing.T) {
	t.Parallel()

	metadata := FsMetadata{
		ID:         "test-id",
		Name:       "test-name",
		DocLibType: cenum.DocLibTypeStrCustom,
		Path:       "/test/path",
		Size:       9999,
	}

	assert.Equal(t, "test-id", metadata.ID)
	assert.Equal(t, "test-name", metadata.Name)
	assert.Equal(t, cenum.DocLibTypeStrCustom, metadata.DocLibType)
	assert.Equal(t, "/test/path", metadata.Path)
	assert.Equal(t, int64(9999), metadata.Size)
}

func TestFsMetadata_WithEmptyPath(t *testing.T) {
	t.Parallel()

	metadata := FsMetadata{
		ID:         "file-123",
		Name:       "Test File",
		DocLibType: cenum.DocLibTypeStrKnowledge,
		Path:       "",
		Size:       1024,
	}

	assert.Empty(t, metadata.Path)
}

func TestFsMetadata_WithNegativeSize(t *testing.T) {
	t.Parallel()

	metadata := FsMetadata{
		ID:   "file-123",
		Name: "Test File",
		Size: -1,
	}

	assert.Equal(t, int64(-1), metadata.Size)
}

func TestGetFsMetadataRetDto_WithMixedTypes(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataRetDto{
		&FsMetadata{
			ID:         "file-1",
			Name:       "File 1",
			DocLibType: cenum.DocLibTypeStrPersonal,
			Size:       1024,
		},
		&FsMetadata{
			ID:         "folder-1",
			Name:       "Folder 1",
			DocLibType: cenum.DocLibTypeStrDepartment,
			Size:       0,
		},
	}

	assert.Equal(t, cenum.DocLibTypeStrPersonal, dto[0].DocLibType)
	assert.Equal(t, cenum.DocLibTypeStrDepartment, dto[1].DocLibType)
}
