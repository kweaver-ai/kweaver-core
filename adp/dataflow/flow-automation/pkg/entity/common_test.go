package entity

import (
	"testing"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/store"
	"github.com/stretchr/testify/assert"
)

func TestBaseInfo_Initial(t *testing.T) {
	store.InitFlakeGenerator(100)
	bi := &BaseInfo{}
	bi.Initial()
	assert.NotEmpty(t, bi.ID)
	assert.NotZero(t, bi.CreatedAt)
	assert.NotZero(t, bi.UpdatedAt)

	bi = &BaseInfo{ID: "test"}
	bi.Initial()
	assert.Equal(t, "test", bi.ID)
	assert.NotZero(t, bi.CreatedAt)
	assert.NotZero(t, bi.UpdatedAt)
}

func TestBaseInfo_Update(t *testing.T) {
	bi := &BaseInfo{}
	bi.Update()
	assert.Zero(t, bi.CreatedAt)
	assert.NotZero(t, bi.UpdatedAt)
}
