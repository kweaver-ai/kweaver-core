package datasourcevalobj

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRetrieverDataSource(t *testing.T) {
	t.Parallel()

	ds := NewRetrieverDataSource()

	assert.NotNil(t, ds)
	assert.NotNil(t, ds.Kg)
	assert.NotNil(t, ds.Doc)
	assert.NotNil(t, ds.Metric)
	assert.NotNil(t, ds.KnEntry)
	assert.NotNil(t, ds.KnowledgeNetwork)
	assert.NotNil(t, ds.AdvancedConfig)
	assert.Empty(t, ds.Kg)
	assert.Empty(t, ds.Doc)
	assert.Empty(t, ds.Metric)
	assert.Empty(t, ds.KnEntry)
	assert.Empty(t, ds.KnowledgeNetwork)
}

func TestRetrieverDataSource_IsNotSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ds   *RetrieverDataSource
		want bool
	}{
		{
			name: "all empty",
			ds: &RetrieverDataSource{
				Kg:     []*KgSource{},
				Doc:    []*DocSource{},
				Metric: []*MetricSource{},
			},
			want: true,
		},
		{
			name: "has kg source",
			ds: &RetrieverDataSource{
				Kg:     []*KgSource{{}},
				Doc:    []*DocSource{},
				Metric: []*MetricSource{},
			},
			want: false,
		},
		{
			name: "has doc source",
			ds: &RetrieverDataSource{
				Kg:     []*KgSource{},
				Doc:    []*DocSource{{}},
				Metric: []*MetricSource{},
			},
			want: false,
		},
		{
			name: "has metric source",
			ds: &RetrieverDataSource{
				Kg:     []*KgSource{},
				Doc:    []*DocSource{},
				Metric: []*MetricSource{{}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.ds.IsNotSet()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRetrieverDataSource_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	ds := &RetrieverDataSource{}
	msgMap := ds.GetErrMsgMap()

	assert.NotNil(t, msgMap)
	assert.Empty(t, msgMap)
}

func TestRetrieverDataSource_ValObjCheckWithCtx_Empty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ds := &RetrieverDataSource{
		Kg:     []*KgSource{},
		Doc:    []*DocSource{},
		Metric: []*MetricSource{},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.NoError(t, err)
}

func TestRetrieverDataSource_ValObjCheckWithCtx_WithInvalidKg(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ds := &RetrieverDataSource{
		Kg: []*KgSource{{}}, // Invalid - missing required fields
		AdvancedConfig: &RetrieverAdvancedConfig{
			KG: &KGAdvancedConfig{},
		},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kg is invalid")
}

func TestRetrieverDataSource_ValObjCheckWithCtx_WithValidKg(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validKG := 60
	simThreshold := -5.5
	graphRagTopK := 25
	longTextLength := 256
	retrievalMaxLength := 1000

	ds := &RetrieverDataSource{
		Kg: []*KgSource{{KgID: "test-kg", Fields: []string{"field1"}}},
		AdvancedConfig: &RetrieverAdvancedConfig{
			KG: &KGAdvancedConfig{
				TextMatchEntityNums:   &validKG,
				VectorMatchEntityNums: &validKG,
				GraphRagTopK:          &graphRagTopK,
				LongTextLength:        &longTextLength,
				RerankerSimThreshold:  &simThreshold,
				RetrievalMaxLength:    &retrievalMaxLength,
			},
		},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.NoError(t, err)
}

func TestRetrieverDataSource_ValObjCheckWithCtx_MissingAdvancedConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ds := &RetrieverDataSource{
		Kg:             []*KgSource{{KgID: "test-kg", Fields: []string{"field1"}}},
		AdvancedConfig: nil,
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "advanced_config is required")
}

func TestRetrieverDataSource_ValObjCheckWithCtx_MissingKgAdvancedConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ds := &RetrieverDataSource{
		Kg: []*KgSource{{KgID: "test-kg", Fields: []string{"field1"}}},
		AdvancedConfig: &RetrieverAdvancedConfig{
			KG: nil,
		},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "advanced_config.kg is required")
}

func TestRetrieverDataSource_ValObjCheckWithCtx_WithInvalidDoc(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ds := &RetrieverDataSource{
		Doc: []*DocSource{{DsID: ""}}, // Invalid - missing required fields
		AdvancedConfig: &RetrieverAdvancedConfig{
			Doc: &DocAdvancedConfig{},
		},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "doc is invalid")
}

func TestRetrieverDataSource_ValObjCheckWithCtx_WithValidDoc(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	retrievalSlicesNum := 100
	maxSlicePerCite := 5
	rerankTopK := 10
	sliceHeadNum := 2
	sliceTailNum := 0
	documentsNum := 8
	docThreshold := -5.5
	retrievalMaxLength := 1000

	// Create a valid DocSourceField with all required fields
	validField := &DocSourceField{
		Name:   "test_field",
		Path:   "test/path",
		Source: "gns://92EE2D87255142B78A6F1DFB6BBB836B/B08AC060A758422583A851C601C0A89B",
		Type:   cdaenum.DocSourceFieldTypeFile,
	}

	ds := &RetrieverDataSource{
		Doc: []*DocSource{{DsID: "test-doc", Fields: []*DocSourceField{validField}}},
		AdvancedConfig: &RetrieverAdvancedConfig{
			Doc: &DocAdvancedConfig{
				RetrievalSlicesNum: &retrievalSlicesNum,
				MaxSlicePerCite:    &maxSlicePerCite,
				RerankTopK:         &rerankTopK,
				SliceHeadNum:       &sliceHeadNum,
				SliceTailNum:       &sliceTailNum,
				DocumentsNum:       &documentsNum,
				DocumentThreshold:  &docThreshold,
				RetrievalMaxLength: &retrievalMaxLength,
			},
		},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.NoError(t, err)
}

func TestRetrieverDataSource_ValObjCheckWithCtx_MissingDocAdvancedConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// Create a valid DocSourceField with all required fields
	validField := &DocSourceField{
		Name:   "test_field",
		Path:   "test/path",
		Source: "gns://92EE2D87255142B78A6F1DFB6BBB836B/B08AC060A758422583A851C601C0A89B",
		Type:   cdaenum.DocSourceFieldTypeFile,
	}

	ds := &RetrieverDataSource{
		Doc: []*DocSource{{DsID: "test-doc", Fields: []*DocSourceField{validField}}},
		AdvancedConfig: &RetrieverAdvancedConfig{
			Doc: nil,
		},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "advanced_config.doc is required")
}

func TestRetrieverDataSource_ValObjCheckWithCtx_KgAdvancedConfigInvalidWhenKgEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ds := &RetrieverDataSource{
		Kg: []*KgSource{},
		AdvancedConfig: &RetrieverAdvancedConfig{
			KG: &KGAdvancedConfig{},
		},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "advanced_config.kg is invalid when data_source.kg is empty")
}

func TestRetrieverDataSource_ValObjCheckWithCtx_DocAdvancedConfigInvalidWhenDocEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ds := &RetrieverDataSource{
		Doc: []*DocSource{},
		AdvancedConfig: &RetrieverAdvancedConfig{
			Doc: &DocAdvancedConfig{},
		},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "advanced_config.doc is invalid when data_source.doc is empty")
}

func TestRetrieverDataSource_GetBuiltInDocDataSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ds   *RetrieverDataSource
		want *DocSource
	}{
		{
			name: "no doc sources",
			ds: &RetrieverDataSource{
				Doc: []*DocSource{},
			},
			want: nil,
		},
		{
			name: "has built-in doc source (DsID=0)",
			ds: &RetrieverDataSource{
				Doc: []*DocSource{
					{DsID: "1", Fields: []*DocSourceField{{Name: "field1"}}},
					{DsID: "0", Fields: []*DocSourceField{{Name: "field2"}}},
				},
			},
			want: &DocSource{DsID: "0"},
		},
		{
			name: "no built-in doc source",
			ds: &RetrieverDataSource{
				Doc: []*DocSource{
					{DsID: "1", Fields: []*DocSourceField{{Name: "field1"}}},
					{DsID: "2", Fields: []*DocSourceField{{Name: "field2"}}},
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.ds.GetBuiltInDocDataSource()
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.DsID, got.DsID)
			}
		})
	}
}

func TestRetrieverDataSource_GetBuiltInDsDocSourceFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ds        *RetrieverDataSource
		wantCount int
	}{
		{
			name: "no doc sources",
			ds: &RetrieverDataSource{
				Doc: []*DocSource{},
			},
			wantCount: 0,
		},
		{
			name: "has built-in doc source with fields",
			ds: &RetrieverDataSource{
				Doc: []*DocSource{
					{DsID: "1", Fields: []*DocSourceField{{Name: "field1"}}},
					{
						DsID:   "0",
						Fields: []*DocSourceField{{Name: "field1"}, {Name: "field2"}},
					},
				},
			},
			wantCount: 2,
		},
		{
			name: "no built-in doc source",
			ds: &RetrieverDataSource{
				Doc: []*DocSource{
					{DsID: "1", Fields: []*DocSourceField{{Name: "field1"}}},
					{DsID: "2", Fields: []*DocSourceField{{Name: "field2"}}},
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.ds.GetBuiltInDsDocSourceFields()
			assert.Len(t, got, tt.wantCount)
		})
	}
}

func TestRetrieverDataSource_GetFirstDocDatasetId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ds   *RetrieverDataSource
		want string
	}{
		{
			name: "no doc sources",
			ds: &RetrieverDataSource{
				Doc: []*DocSource{},
			},
			want: "",
		},
		{
			name: "has built-in doc source",
			ds: &RetrieverDataSource{
				Doc: []*DocSource{
					{DsID: "1", Fields: []*DocSourceField{{Name: "field1"}}},
					{
						DsID:     "0",
						Fields:   []*DocSourceField{{Name: "field2"}},
						Datasets: []string{"dataset1", "dataset2"},
					},
				},
			},
			want: "dataset1",
		},
		{
			name: "no built-in doc source",
			ds: &RetrieverDataSource{
				Doc: []*DocSource{
					{DsID: "1", Fields: []*DocSourceField{{Name: "field1"}}},
					{DsID: "2", Fields: []*DocSourceField{{Name: "field2"}}},
				},
			},
			want: "",
		},
		{
			name: "built-in doc source has no datasets",
			ds: &RetrieverDataSource{
				Doc: []*DocSource{
					{DsID: "0", Fields: []*DocSourceField{{Name: "field1"}}, Datasets: []string{}},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.ds.GetFirstDocDatasetId()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRetrieverDataSource_ValObjCheckWithCtx_InvalidKnowledgeNetwork(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ds := &RetrieverDataSource{
		KnowledgeNetwork: []*KnowledgeNetworkSource{
			{}, // Invalid - missing required fields
		},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "knowledge_network is invalid")
}

func TestRetrieverDataSource_ValObjCheckWithCtx_ValidKnowledgeNetwork(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ds := &RetrieverDataSource{
		KnowledgeNetwork: []*KnowledgeNetworkSource{
			{
				KnowledgeNetworkID: "test-kn",
				ObjectTypes: []*ObjectType{
					{ObjectTypeID: "type1"},
					{ObjectTypeID: "type2"},
				},
			},
		},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.NoError(t, err)
}

func TestRetrieverDataSource_ValObjCheckWithCtx_InvalidKnEntry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ds := &RetrieverDataSource{
		KnEntry: []*KnEntrySource{
			{}, // Invalid - missing required fields
		},
	}

	err := ds.ValObjCheckWithCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kn_entry is invalid")
}
