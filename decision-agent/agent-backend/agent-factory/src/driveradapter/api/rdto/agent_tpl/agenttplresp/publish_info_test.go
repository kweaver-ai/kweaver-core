package agenttplresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishInfoRes_StructFields(t *testing.T) {
	t.Parallel()

	resp := PublishInfoRes{
		Categories: []CategoryInfo{
			{
				ID:   "cat-1",
				Name: "Category 1",
			},
			{
				ID:   "cat-2",
				Name: "Category 2",
			},
		},
	}

	assert.Len(t, resp.Categories, 2)
	assert.Equal(t, "cat-1", resp.Categories[0].ID)
	assert.Equal(t, "Category 1", resp.Categories[0].Name)
	assert.Equal(t, "cat-2", resp.Categories[1].ID)
	assert.Equal(t, "Category 2", resp.Categories[1].Name)
}

func TestPublishInfoRes_Empty(t *testing.T) {
	t.Parallel()

	resp := PublishInfoRes{}

	assert.Nil(t, resp.Categories)
}

func TestPublishInfoRes_WithEmptyCategories(t *testing.T) {
	t.Parallel()

	resp := PublishInfoRes{
		Categories: []CategoryInfo{},
	}

	assert.NotNil(t, resp.Categories)
	assert.Len(t, resp.Categories, 0)
}

func TestCategoryInfo_StructFields(t *testing.T) {
	t.Parallel()

	cat := CategoryInfo{
		ID:   "cat-123",
		Name: "Test Category",
	}

	assert.Equal(t, "cat-123", cat.ID)
	assert.Equal(t, "Test Category", cat.Name)
}

func TestCategoryInfo_Empty(t *testing.T) {
	t.Parallel()

	cat := CategoryInfo{}

	assert.Empty(t, cat.ID)
	assert.Empty(t, cat.Name)
}

func TestPublishInfoRes_WithSingleCategory(t *testing.T) {
	t.Parallel()

	resp := PublishInfoRes{
		Categories: []CategoryInfo{
			{
				ID:   "cat-single",
				Name: "Single Category",
			},
		},
	}

	assert.Len(t, resp.Categories, 1)
	assert.Equal(t, "cat-single", resp.Categories[0].ID)
}

func TestPublishInfoRes_WithMultipleCategories(t *testing.T) {
	t.Parallel()

	resp := PublishInfoRes{
		Categories: []CategoryInfo{
			{ID: "cat-1", Name: "Category One"},
			{ID: "cat-2", Name: "Category Two"},
			{ID: "cat-3", Name: "Category Three"},
		},
	}

	assert.Len(t, resp.Categories, 3)
}

func TestPublishInfoRes_Append(t *testing.T) {
	t.Parallel()

	resp := PublishInfoRes{
		Categories: []CategoryInfo{},
	}

	resp.Categories = append(resp.Categories, CategoryInfo{
		ID:   "cat-new",
		Name: "New Category",
	})

	assert.Len(t, resp.Categories, 1)
}

func TestCategoryInfo_WithDifferentIDs(t *testing.T) {
	t.Parallel()

	ids := []string{
		"cat-001",
		"cat-xyz",
		"分类-123",
		"",
	}

	for _, id := range ids {
		cat := CategoryInfo{
			ID: id,
		}
		assert.Equal(t, id, cat.ID)
	}
}

func TestCategoryInfo_WithDifferentNames(t *testing.T) {
	t.Parallel()

	names := []string{
		"Test Category",
		"测试分类",
		"Category with numbers 123",
		"Category with special chars !@#$%",
		"",
	}

	for _, name := range names {
		cat := CategoryInfo{
			Name: name,
		}
		assert.Equal(t, name, cat.Name)
	}
}

func TestPublishInfoRes_WithChineseCategories(t *testing.T) {
	t.Parallel()

	resp := PublishInfoRes{
		Categories: []CategoryInfo{
			{
				ID:   "cat-中文",
				Name: "中文分类",
			},
		},
	}

	assert.Equal(t, "cat-中文", resp.Categories[0].ID)
	assert.Equal(t, "中文分类", resp.Categories[0].Name)
}

func TestPublishInfoRes_SliceOperations(t *testing.T) {
	t.Parallel()

	resp := PublishInfoRes{
		Categories: []CategoryInfo{
			{ID: "cat-1", Name: "Category 1"},
			{ID: "cat-2", Name: "Category 2"},
			{ID: "cat-3", Name: "Category 3"},
		},
	}

	// Test length
	assert.Len(t, resp.Categories, 3)

	// Test slicing
	subCategories := resp.Categories[1:3]
	assert.Len(t, subCategories, 2)
	assert.Equal(t, "cat-2", subCategories[0].ID)

	// Test iteration
	count := 0

	for _, cat := range resp.Categories {
		assert.NotEmpty(t, cat.ID)

		count++
	}

	assert.Equal(t, 3, count)
}

func TestCategoryInfo_WithAllFields(t *testing.T) {
	t.Parallel()

	cat := CategoryInfo{
		ID:   "cat-complete",
		Name: "Complete Category Name",
	}

	assert.Equal(t, "cat-complete", cat.ID)
	assert.Equal(t, "Complete Category Name", cat.Name)
}
