package publishvo

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/stretchr/testify/assert"
)

func TestPublishInfo_NewInstance(t *testing.T) {
	t.Parallel()

	info := &PublishInfo{}

	assert.NotNil(t, info)
	assert.Nil(t, info.CategoryIDs)
	assert.Equal(t, "", info.Description)
	assert.Nil(t, info.PublishToWhere)
	assert.Nil(t, info.PmsControl)
	assert.Nil(t, info.PublishToBes)
}

func TestPublishInfo_WithCategoryIDs(t *testing.T) {
	t.Parallel()

	info := &PublishInfo{
		CategoryIDs: []string{"cat_1", "cat_2", "cat_3"},
	}

	assert.Len(t, info.CategoryIDs, 3)
	assert.Equal(t, "cat_1", info.CategoryIDs[0])
	assert.Equal(t, "cat_2", info.CategoryIDs[1])
	assert.Equal(t, "cat_3", info.CategoryIDs[2])
}

func TestPublishInfo_WithEmptyCategoryIDs(t *testing.T) {
	t.Parallel()

	info := &PublishInfo{
		CategoryIDs: []string{},
	}

	assert.NotNil(t, info.CategoryIDs)
	assert.Len(t, info.CategoryIDs, 0)
}

func TestPublishInfo_WithDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "normal description",
			description: "This is a test agent description",
		},
		{
			name:        "empty description",
			description: "",
		},
		{
			name:        "description with special characters",
			description: "Description with @#$% special chars",
		},
		{
			name:        "unicode description",
			description: "这是一个测试描述",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info := &PublishInfo{
				Description: tt.description,
			}

			assert.Equal(t, tt.description, info.Description)
		})
	}
}

func TestPublishInfo_WithPublishToWhere(t *testing.T) {
	t.Parallel()

	info := &PublishInfo{
		PublishToWhere: []daenum.PublishToWhere{
			daenum.PublishToWhereCustomSpace,
			daenum.PublishToWhereSquare,
		},
	}

	assert.Len(t, info.PublishToWhere, 2)
	assert.Equal(t, daenum.PublishToWhereCustomSpace, info.PublishToWhere[0])
	assert.Equal(t, daenum.PublishToWhereSquare, info.PublishToWhere[1])
}

func TestPublishInfo_WithPublishToBes(t *testing.T) {
	t.Parallel()

	info := &PublishInfo{
		PublishToBes: []cdaenum.PublishToBe{
			cdaenum.PublishToBeSkillAgent,
		},
	}

	assert.Len(t, info.PublishToBes, 1)
	assert.Equal(t, cdaenum.PublishToBeSkillAgent, info.PublishToBes[0])
}

func TestPublishInfo_WithEmptyPublishToBes(t *testing.T) {
	t.Parallel()

	info := &PublishInfo{
		PublishToBes: []cdaenum.PublishToBe{},
	}

	assert.NotNil(t, info.PublishToBes)
	assert.Len(t, info.PublishToBes, 0)
}

func TestPublishInfo_FullStructure(t *testing.T) {
	t.Parallel()

	info := &PublishInfo{
		CategoryIDs: []string{"cat_1", "cat_2"},
		Description: "Test description",
		PublishToWhere: []daenum.PublishToWhere{
			daenum.PublishToWhereSquare,
		},
		PublishToBes: []cdaenum.PublishToBe{
			cdaenum.PublishToBeSkillAgent,
		},
	}

	assert.Len(t, info.CategoryIDs, 2)
	assert.Equal(t, "Test description", info.Description)
	assert.Len(t, info.PublishToWhere, 1)
	assert.Len(t, info.PublishToBes, 1)
}
