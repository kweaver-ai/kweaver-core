package dbaulid

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cconstant"
	"github.com/stretchr/testify/assert"
)

func TestNewUlidRepo(t *testing.T) {
	t.Parallel()

	repo := NewUlidRepo()

	assert.NotNil(t, repo, "NewUlidRepo should return non-nil repo")
}

func TestUniqueID_TableName(t *testing.T) {
	t.Parallel()

	uid := &UniqueID{}

	tableName := uid.TableName()

	assert.Equal(t, "t_stc_unique_id", tableName, "TableName should return correct table name")
}

func TestUniqueID_Fields(t *testing.T) {
	t.Parallel()

	uid := &UniqueID{
		ID:   "test123",
		Flag: cconstant.UniqueIDFlag(1),
	}

	assert.Equal(t, "test123", uid.ID, "ID field should be set")
	assert.Equal(t, cconstant.UniqueIDFlag(1), uid.Flag, "Flag field should be set")
}
