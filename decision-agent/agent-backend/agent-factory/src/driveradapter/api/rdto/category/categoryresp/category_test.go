package categoryresp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryResp_StructFields(t *testing.T) {
	t.Parallel()

	category := CategoryResp{
		ID:          "cat-123",
		Name:        "TestCategory",
		Description: "Test category description",
	}

	assert.Equal(t, "cat-123", category.ID)
	assert.Equal(t, "TestCategory", category.Name)
	assert.Equal(t, "Test category description", category.Description)
}

func TestCategoryResp_JSONTags(t *testing.T) {
	t.Parallel()

	category := CategoryResp{
		ID:          "cat-456",
		Name:        "JSON Category",
		Description: "JSON description",
	}

	// Marshal to JSON
	data, err := json.Marshal(category)
	require.NoError(t, err)

	// Unmarshal to map to check JSON tags
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "cat-456", result["category_id"])
	assert.Equal(t, "JSON Category", result["name"])
	assert.Equal(t, "JSON description", result["description"])
}

func TestCategoryResp_EmptyValues(t *testing.T) {
	t.Parallel()

	category := CategoryResp{}

	assert.Empty(t, category.ID)
	assert.Empty(t, category.Name)
	assert.Empty(t, category.Description)
}

func TestListResp_Type(t *testing.T) {
	t.Parallel()

	// ListResp is a slice type
	var list ListResp

	assert.Nil(t, list)
	assert.IsType(t, ListResp{}, list)
}

func TestListResp_WithMultipleCategories(t *testing.T) {
	t.Parallel()

	list := ListResp{
		{
			ID:          "cat-1",
			Name:        "Category 1",
			Description: "Description 1",
		},
		{
			ID:          "cat-2",
			Name:        "Category 2",
			Description: "Description 2",
		},
		{
			ID:          "cat-3",
			Name:        "Category 3",
			Description: "Description 3",
		},
	}

	assert.Len(t, list, 3)
	assert.Equal(t, "cat-1", list[0].ID)
	assert.Equal(t, "cat-2", list[1].ID)
	assert.Equal(t, "cat-3", list[2].ID)
}

func TestListResp_Empty(t *testing.T) {
	t.Parallel()

	list := ListResp{}

	assert.Empty(t, list)
	assert.Len(t, list, 0)
}

func TestListResp_Append(t *testing.T) {
	t.Parallel()

	list := ListResp{}

	// Append categories
	list = append(list, CategoryResp{ID: "cat-1", Name: "Category 1"})
	list = append(list, CategoryResp{ID: "cat-2", Name: "Category 2"})

	assert.Len(t, list, 2)
	assert.Equal(t, "Category 1", list[0].Name)
	assert.Equal(t, "Category 2", list[1].Name)
}

func TestListResp_JSONMarshaling(t *testing.T) {
	t.Parallel()

	list := ListResp{
		{
			ID:          "cat-1",
			Name:        "Category 1",
			Description: "Description 1",
		},
		{
			ID:          "cat-2",
			Name:        "Category 2",
			Description: "Description 2",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(list)
	require.NoError(t, err)

	// Unmarshal back
	var result ListResp
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, "cat-1", result[0].ID)
	assert.Equal(t, "cat-2", result[1].ID)
}

func TestCategoryResp_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	category := CategoryResp{
		ID:          "cat-中文-123",
		Name:        "分类名称",
		Description: "This is a description with \"quotes\" and 'apostrophes'",
	}

	assert.Equal(t, "cat-中文-123", category.ID)
	assert.Equal(t, "分类名称", category.Name)
	assert.Contains(t, category.Description, "quotes")
}

func TestCategoryResp_LongValues(t *testing.T) {
	t.Parallel()

	longID := string(make([]byte, 1000))
	longName := "This is a very long category name that exceeds normal length but should still work"
	longDesc := string(make([]byte, 5000))

	category := CategoryResp{
		ID:          longID,
		Name:        longName,
		Description: longDesc,
	}

	assert.Len(t, category.ID, 1000)
	assert.Equal(t, longName, category.Name)
	assert.Len(t, category.Description, 5000)
}
