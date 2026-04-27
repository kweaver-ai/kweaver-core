package daconfeo

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestDataAgent_GetObjName(t *testing.T) {
	t.Parallel()

	da := &DataAgent{
		DataAgentPo: dapo.DataAgentPo{
			Name: "Test Agent",
		},
	}

	assert.Equal(t, "Test Agent", da.GetObjName())
}

func TestDataAgent_GetObjName_Empty(t *testing.T) {
	t.Parallel()

	da := &DataAgent{}

	assert.Empty(t, da.GetObjName())
}

func TestDataAgent_AuditMngLogCreate(t *testing.T) {
	t.Parallel()

	da := &DataAgent{}

	// This method is a stub, just call it to ensure no panic
	assert.NotPanics(t, func() {
		da.AuditMngLogCreate(context.Background())
	})
}

func TestDataAgent_AuditMngLogUpdate(t *testing.T) {
	t.Parallel()

	da := &DataAgent{}

	// This method is a stub, just call it to ensure no panic
	assert.NotPanics(t, func() {
		da.AuditMngLogUpdate(context.Background())
	})
}

func TestDataAgent_AuditMngLogDelete(t *testing.T) {
	t.Parallel()

	da := &DataAgent{}

	// This method is a stub, just call it to ensure no panic
	assert.NotPanics(t, func() {
		da.AuditMngLogDelete(context.Background())
	})
}

func TestDataAgent_SetDatasetId_NilConfig(t *testing.T) {
	t.Parallel()

	da := &DataAgent{
		Config: nil,
	}

	// Should not panic
	assert.NotPanics(t, func() {
		da.SetDatasetId("dataset-123")
	})
}

func TestDataAgent_SetDatasetId_NilDataSource(t *testing.T) {
	t.Parallel()

	da := &DataAgent{
		Config: &daconfvalobj.Config{
			DataSource: nil,
		},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		da.SetDatasetId("dataset-123")
	})
}

func TestDataAgent_SetDatasetId_NilDocSource(t *testing.T) {
	t.Parallel()

	da := &DataAgent{
		Config: &daconfvalobj.Config{
			DataSource: &datasourcevalobj.RetrieverDataSource{
				Doc: nil,
			},
		},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		da.SetDatasetId("dataset-123")
	})
}

func TestDataAgent_SetDatasetId_EmptyDocSource(t *testing.T) {
	t.Parallel()

	da := &DataAgent{
		Config: &daconfvalobj.Config{
			DataSource: &datasourcevalobj.RetrieverDataSource{
				Doc: []*datasourcevalobj.DocSource{},
			},
		},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		da.SetDatasetId("dataset-123")
	})
}

func TestDataAgent_SetDatasetId_NoBuiltInDataSource(t *testing.T) {
	t.Parallel()

	da := &DataAgent{
		Config: &daconfvalobj.Config{
			DataSource: &datasourcevalobj.RetrieverDataSource{
				Doc: []*datasourcevalobj.DocSource{
					{
						DsID: "1",
						Fields: []*datasourcevalobj.DocSourceField{
							{
								Name:   "test",
								Path:   "/test",
								Source: "gns://test/id",
								Type:   cdaenum.DocSourceFieldTypeFile,
							},
						},
					},
				},
			},
		},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		da.SetDatasetId("dataset-123")
	})
}

func TestDataAgent_SetDatasetId_ValidBuiltInDataSource(t *testing.T) {
	t.Parallel()

	da := &DataAgent{
		Config: &daconfvalobj.Config{
			DataSource: &datasourcevalobj.RetrieverDataSource{
				Doc: []*datasourcevalobj.DocSource{
					{
						DsID: "0", // Built-in data source
						Fields: []*datasourcevalobj.DocSourceField{
							{
								Name:   "test",
								Path:   "/test",
								Source: "gns://test/id",
								Type:   cdaenum.DocSourceFieldTypeFile,
							},
						},
					},
				},
			},
		},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		da.SetDatasetId("dataset-123")
	})

	// Verify the dataset ID was set
	assert.Len(t, da.Config.DataSource.Doc[0].Datasets, 1)
	assert.Equal(t, "dataset-123", da.Config.DataSource.Doc[0].Datasets[0])
}

func TestDataAgent_Fields(t *testing.T) {
	t.Parallel()

	config := &daconfvalobj.Config{}
	da := &DataAgent{
		ProductName:   "Test Product",
		CreatedByName: "User 1",
		UpdatedByName: "User 2",
		Config:        config,
	}

	assert.Equal(t, "Test Product", da.ProductName)
	assert.Equal(t, "User 1", da.CreatedByName)
	assert.Equal(t, "User 2", da.UpdatedByName)
	assert.Equal(t, config, da.Config)
}

func TestDataAgent_Empty(t *testing.T) {
	t.Parallel()

	da := &DataAgent{}

	assert.Empty(t, da.ProductName)
	assert.Empty(t, da.CreatedByName)
	assert.Empty(t, da.UpdatedByName)
	assert.Nil(t, da.Config)
}
