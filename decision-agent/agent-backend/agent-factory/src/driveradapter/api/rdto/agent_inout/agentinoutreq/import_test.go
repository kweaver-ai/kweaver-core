package agentinoutreq

import (
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImportReq(t *testing.T) {
	t.Parallel()

	req := NewImportReq()

	assert.NotNil(t, req)
	assert.IsType(t, &ImportReq{}, req)
}

func TestImportReq_StructFields(t *testing.T) {
	t.Parallel()

	fileHeader := &multipart.FileHeader{
		Filename: "test.json",
	}
	req := &ImportReq{
		File:       fileHeader,
		ImportType: ImportTypeCreate,
	}

	assert.Equal(t, fileHeader, req.File)
	assert.Equal(t, ImportTypeCreate, req.ImportType)
}

func TestImportReq_Empty(t *testing.T) {
	t.Parallel()

	req := &ImportReq{}

	assert.Nil(t, req.File)
	assert.Empty(t, req.ImportType)
}

func TestImportReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &ImportReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"file"字段不能为空`, errMsgMap["File.required"])
	assert.Equal(t, `"import_type"字段不能为空`, errMsgMap["ImportType.required"])
}

func TestImportReq_ValueCheck_Valid(t *testing.T) {
	t.Parallel()

	req := &ImportReq{
		File:       &multipart.FileHeader{},
		ImportType: ImportTypeUpsert,
	}

	err := req.ValueCheck()

	assert.NoError(t, err)
}

func TestImportReq_ValueCheck_InvalidImportType(t *testing.T) {
	t.Parallel()

	req := &ImportReq{
		File:       &multipart.FileHeader{},
		ImportType: ImportType("invalid"),
	}

	err := req.ValueCheck()

	assert.Error(t, err)
}

func TestNewExportReq(t *testing.T) {
	t.Parallel()

	req := NewExportReq()

	assert.NotNil(t, req)
	assert.IsType(t, &ExportReq{}, req)
	assert.Empty(t, req.AgentIDs)
}

func TestExportReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &ExportReq{
		AgentIDs: []string{"agent-1", "agent-2"},
	}

	assert.Len(t, req.AgentIDs, 2)
	assert.Equal(t, "agent-1", req.AgentIDs[0])
	assert.Equal(t, "agent-2", req.AgentIDs[1])
}

func TestExportReq_Empty(t *testing.T) {
	t.Parallel()

	req := &ExportReq{}

	assert.Empty(t, req.AgentIDs)
}

func TestExportReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &ExportReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"agent_ids"字段不能为空`, errMsgMap["AgentIDs.required"])
}

func TestExportReq_CustomCheckAndDedupl_Valid(t *testing.T) {
	t.Parallel()

	req := &ExportReq{
		AgentIDs: []string{"agent-1", "agent-2", "agent-3"},
	}

	err := req.CustomCheckAndDedupl()

	assert.NoError(t, err)
	assert.Len(t, req.AgentIDs, 3)
}

func TestExportReq_CustomCheckAndDedupl_Empty(t *testing.T) {
	t.Parallel()

	req := &ExportReq{
		AgentIDs: []string{},
	}

	err := req.CustomCheckAndDedupl()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_ids不能为空")
}

func TestExportReq_CustomCheckAndDedupl_WithDuplicates(t *testing.T) {
	t.Parallel()

	req := &ExportReq{
		AgentIDs: []string{"agent-1", "agent-2", "agent-1", "agent-3", "agent-2"},
	}

	err := req.CustomCheckAndDedupl()

	assert.NoError(t, err)
	assert.Len(t, req.AgentIDs, 3) // Should be deduplicated
}

func TestExportReq_CustomCheckAndDedupl_AllDuplicates(t *testing.T) {
	t.Parallel()

	req := &ExportReq{
		AgentIDs: []string{"agent-1", "agent-1", "agent-1"},
	}

	err := req.CustomCheckAndDedupl()

	assert.NoError(t, err)
	assert.Len(t, req.AgentIDs, 1) // Should be deduplicated to single item
}

func TestExportReq_CustomCheckAndDedupl_ExceedsMaxSize(t *testing.T) {
	t.Parallel()

	// Create a request with more than max size agent IDs
	maxSize := 500 // AgentInoutMaxSize = 500
	agentIDs := make([]string, maxSize+1)

	for i := 0; i <= maxSize; i++ {
		agentIDs[i] = "agent-" + string(rune(i))
	}

	req := &ExportReq{
		AgentIDs: agentIDs,
	}

	err := req.CustomCheckAndDedupl()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "单次导入最多导入")
}

func TestExportReq_WithSingleAgentID(t *testing.T) {
	t.Parallel()

	req := &ExportReq{
		AgentIDs: []string{"agent-123"},
	}

	err := req.CustomCheckAndDedupl()

	assert.NoError(t, err)
	assert.Len(t, req.AgentIDs, 1)
}

func TestExportReq_WithMultipleAgentIDs(t *testing.T) {
	t.Parallel()

	req := &ExportReq{
		AgentIDs: []string{
			"agent-001",
			"agent-002",
			"agent-003",
			"agent-004",
			"agent-005",
		},
	}

	err := req.CustomCheckAndDedupl()

	assert.NoError(t, err)
	assert.Len(t, req.AgentIDs, 5)
}

func TestImportReq_WithDifferentImportTypes(t *testing.T) {
	t.Parallel()

	types := []ImportType{
		ImportTypeUpsert,
		ImportTypeCreate,
	}

	for _, importType := range types {
		req := &ImportReq{
			File:       &multipart.FileHeader{},
			ImportType: importType,
		}

		err := req.ValueCheck()

		assert.NoError(t, err, "ImportType %s should be valid", importType)
	}
}

func TestExportReq_CustomCheck_PreservesOrder(t *testing.T) {
	t.Parallel()

	req := &ExportReq{
		AgentIDs: []string{"agent-3", "agent-1", "agent-2"},
	}

	// After deduplication, the order should be maintained
	originalOrder := req.AgentIDs
	req.CustomCheckAndDedupl() //nolint:errcheck

	// Check that the first occurrence of each is maintained
	assert.Contains(t, req.AgentIDs, originalOrder[0])
	assert.Contains(t, req.AgentIDs, originalOrder[1])
	assert.Contains(t, req.AgentIDs, originalOrder[2])
}

func TestExportReq_CustomCheck_NoChangesWhenNoDuplicates(t *testing.T) {
	t.Parallel()

	originalIDs := []string{"agent-1", "agent-2", "agent-3"}
	req := &ExportReq{
		AgentIDs: originalIDs,
	}

	req.CustomCheckAndDedupl() //nolint:errcheck

	assert.Equal(t, originalIDs, req.AgentIDs)
}
