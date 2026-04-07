package bizdomainhttpres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetItemIDs_Empty(t *testing.T) {
	t.Parallel()

	res := &QueryResourceAssociationsRes{
		Items: []*ResourceAssociationItem{},
	}
	ids := res.GetItemIDs()
	assert.Empty(t, ids)
}

func TestGetItemIDs_MultipleItems(t *testing.T) {
	t.Parallel()

	res := &QueryResourceAssociationsRes{
		Items: []*ResourceAssociationItem{
			{ID: "id-1", BdID: "bd-1", Type: "agent"},
			{ID: "id-2", BdID: "bd-2", Type: "tpl"},
			{ID: "id-3", BdID: "bd-1", Type: "agent"},
		},
	}
	ids := res.GetItemIDs()
	assert.Equal(t, []string{"id-1", "id-2", "id-3"}, ids)
}

func TestGetItemIDs_Nil(t *testing.T) {
	t.Parallel()

	res := &QueryResourceAssociationsRes{}
	ids := res.GetItemIDs()
	assert.Empty(t, ids)
}
