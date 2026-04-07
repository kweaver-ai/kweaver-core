package productresp

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/producteo"
	"github.com/stretchr/testify/assert"
)

func TestNewListRes(t *testing.T) {
	t.Parallel()

	list := NewListRes()

	assert.NotNil(t, list)
	assert.NotNil(t, list.Entries)
	assert.Empty(t, list.Entries)
	assert.Equal(t, 0, list.Total)
}

func TestListRes_StructFields(t *testing.T) {
	t.Parallel()

	item1 := &ListItem{
		ID:        1,
		Name:      "Product 1",
		Key:       "product-1",
		Profile:   "Profile 1",
		CreatedAt: 1000,
		UpdatedAt: 2000,
	}
	item2 := &ListItem{
		ID:        2,
		Name:      "Product 2",
		Key:       "product-2",
		Profile:   "Profile 2",
		CreatedAt: 3000,
		UpdatedAt: 4000,
	}

	list := &ListRes{
		Entries: []*ListItem{item1, item2},
		Total:   2,
	}

	assert.Len(t, list.Entries, 2)
	assert.Equal(t, 2, list.Total)
	assert.Equal(t, item1, list.Entries[0])
	assert.Equal(t, item2, list.Entries[1])
}

func TestListRes_Empty(t *testing.T) {
	t.Parallel()

	list := &ListRes{
		Entries: []*ListItem{},
		Total:   0,
	}

	assert.Empty(t, list.Entries)
	assert.Equal(t, 0, list.Total)
}

func TestListRes_WithEntries(t *testing.T) {
	t.Parallel()

	list := &ListRes{
		Entries: make([]*ListItem, 0),
		Total:   0,
	}

	// Add entries
	list.Entries = append(list.Entries, &ListItem{ID: 1, Name: "Item 1"})
	list.Entries = append(list.Entries, &ListItem{ID: 2, Name: "Item 2"})
	list.Total = 2

	assert.Len(t, list.Entries, 2)
	assert.Equal(t, 2, list.Total)
	assert.Equal(t, int64(1), list.Entries[0].ID)
	assert.Equal(t, "Item 1", list.Entries[0].Name)
	assert.Equal(t, int64(2), list.Entries[1].ID)
	assert.Equal(t, "Item 2", list.Entries[1].Name)
}

func TestListItem_StructFields(t *testing.T) {
	t.Parallel()

	item := &ListItem{
		ID:        123,
		Name:      "Test Product",
		Key:       "test-product",
		Profile:   "Test Profile",
		CreatedAt: 1234567890,
		UpdatedAt: 1234567899,
	}

	assert.Equal(t, int64(123), item.ID)
	assert.Equal(t, "Test Product", item.Name)
	assert.Equal(t, "test-product", item.Key)
	assert.Equal(t, "Test Profile", item.Profile)
	assert.Equal(t, int64(1234567890), item.CreatedAt)
	assert.Equal(t, int64(1234567899), item.UpdatedAt)
}

func TestListItem_Empty(t *testing.T) {
	t.Parallel()

	item := &ListItem{}

	assert.Equal(t, int64(0), item.ID)
	assert.Empty(t, item.Name)
	assert.Empty(t, item.Key)
	assert.Empty(t, item.Profile)
	assert.Equal(t, int64(0), item.CreatedAt)
	assert.Equal(t, int64(0), item.UpdatedAt)
}

func TestListRes_TotalField(t *testing.T) {
	t.Parallel()

	list := &ListRes{
		Entries: []*ListItem{},
		Total:   10,
	}

	assert.Equal(t, 10, list.Total)
	assert.Empty(t, list.Entries)
}

func TestListRes_LoadFromEo(t *testing.T) {
	t.Parallel()

	t.Run("load from product eos", func(t *testing.T) {
		t.Parallel()

		list := NewListRes()

		eo1 := &producteo.Product{}
		eo1.ID = 1
		eo1.Name = "Product 1"
		eo1.Key = "product-1"
		eo1.Profile = "Profile 1"
		eo1.CreatedAt = 1000
		eo1.UpdatedAt = 2000

		eo2 := &producteo.Product{}
		eo2.ID = 2
		eo2.Name = "Product 2"
		eo2.Key = "product-2"
		eo2.Profile = "Profile 2"
		eo2.CreatedAt = 3000
		eo2.UpdatedAt = 4000

		eos := []*producteo.Product{eo1, eo2}

		err := list.LoadFromEo(eos)
		assert.NoError(t, err)
		assert.Len(t, list.Entries, 2)
		assert.Equal(t, int64(1), list.Entries[0].ID)
		assert.Equal(t, "Product 1", list.Entries[0].Name)
		assert.Equal(t, int64(2), list.Entries[1].ID)
		assert.Equal(t, "Product 2", list.Entries[1].Name)
	})

	t.Run("load from empty slice", func(t *testing.T) {
		t.Parallel()

		list := NewListRes()

		eos := []*producteo.Product{}

		err := list.LoadFromEo(eos)
		assert.NoError(t, err)
		assert.Len(t, list.Entries, 0)
	})

	t.Run("load from nil slice", func(t *testing.T) {
		t.Parallel()

		list := NewListRes()

		err := list.LoadFromEo(nil)
		assert.NoError(t, err)
		assert.Len(t, list.Entries, 0)
	})
}
