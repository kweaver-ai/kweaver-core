package daconfeo

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentTpl_GetObjName(t *testing.T) {
	t.Parallel()

	dat := &DataAgentTpl{
		DataAgentTplPo: dapo.DataAgentTplPo{
			Name: "Test Template",
		},
	}

	assert.Equal(t, "Test Template", dat.GetObjName())
}

func TestDataAgentTpl_GetObjName_Empty(t *testing.T) {
	t.Parallel()

	dat := &DataAgentTpl{}

	assert.Empty(t, dat.GetObjName())
}

func TestDataAgentTpl_AuditMngLogCreate(t *testing.T) {
	t.Parallel()

	dat := &DataAgentTpl{}

	// This method is a stub, just call it to ensure no panic
	assert.NotPanics(t, func() {
		dat.AuditMngLogCreate(context.Background())
	})
}

func TestDataAgentTpl_AuditMngLogUpdate(t *testing.T) {
	t.Parallel()

	dat := &DataAgentTpl{}

	// This method is a stub, just call it to ensure no panic
	assert.NotPanics(t, func() {
		dat.AuditMngLogUpdate(context.Background())
	})
}

func TestDataAgentTpl_AuditMngLogDelete(t *testing.T) {
	t.Parallel()

	dat := &DataAgentTpl{}

	// This method is a stub, just call it to ensure no panic
	assert.NotPanics(t, func() {
		dat.AuditMngLogDelete(context.Background())
	})
}

func TestDataAgentTpl_Fields(t *testing.T) {
	t.Parallel()

	config := &daconfvalobj.Config{}
	dat := &DataAgentTpl{
		ProductName: "Test Product",
		Config:      config,
	}

	assert.Equal(t, "Test Product", dat.ProductName)
	assert.Equal(t, config, dat.Config)
}

func TestDataAgentTpl_Empty(t *testing.T) {
	t.Parallel()

	dat := &DataAgentTpl{}

	assert.Empty(t, dat.ProductName)
	assert.Nil(t, dat.Config)
}
