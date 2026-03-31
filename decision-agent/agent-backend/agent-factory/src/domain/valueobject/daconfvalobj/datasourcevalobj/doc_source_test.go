package datasourcevalobj

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocSource_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ds       *DocSource
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "valid doc source with fields",
			ds: &DocSource{
				DsID: "ds-1",
				Fields: []*DocSourceField{
					{
						Name:   "field1",
						Path:   "path1",
						Source: "gns://abc123/def456",
						Type:   cdaenum.DocSourceFieldTypeFile,
					},
				},
			},
			wantErr:  false,
			checkErr: nil,
		},
		{
			name:    "nil doc source",
			ds:      nil,
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "empty ds id",
			ds: &DocSource{
				DsID:   "",
				Fields: []*DocSourceField{{}},
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "ds_id is required")
			},
		},
		{
			name: "nil fields",
			ds: &DocSource{
				DsID:   "ds-1",
				Fields: nil,
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "fields is required")
			},
		},
		{
			name: "empty fields",
			ds: &DocSource{
				DsID:   "ds-1",
				Fields: []*DocSourceField{},
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "fields is required")
			},
		},
		{
			name: "invalid field in fields",
			ds: &DocSource{
				DsID: "ds-1",
				Fields: []*DocSourceField{
					{
						Name:   "",
						Path:   "path1",
						Source: "gns://abc123/def456",
						Type:   cdaenum.DocSourceFieldTypeFile,
					},
				},
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "field is invalid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.ds.ValObjCheck()
			if tt.wantErr {
				require.Error(t, err)

				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDocSource_SetDatasetId(t *testing.T) {
	t.Parallel()

	ds := &DocSource{
		Datasets: []string{},
	}

	// Add first dataset
	ds.SetDatasetId("dataset-1")
	assert.Len(t, ds.Datasets, 1)
	assert.Equal(t, "dataset-1", ds.Datasets[0])

	// Add second dataset
	ds.SetDatasetId("dataset-2")
	assert.Len(t, ds.Datasets, 2)
	assert.Equal(t, "dataset-2", ds.Datasets[1])

	// Try to add empty dataset (should not add)
	ds.SetDatasetId("")
	assert.Len(t, ds.Datasets, 2)
}

func TestDocSource_GetFirstDatasetId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		datasets   []string
		expectedId string
	}{
		{
			name:       "single dataset",
			datasets:   []string{"dataset-1"},
			expectedId: "dataset-1",
		},
		{
			name:       "multiple datasets",
			datasets:   []string{"dataset-1", "dataset-2"},
			expectedId: "dataset-1",
		},
		{
			name:       "empty datasets",
			datasets:   []string{},
			expectedId: "",
		},
		{
			name:       "nil datasets",
			datasets:   nil,
			expectedId: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ds := &DocSource{
				Datasets: tt.datasets,
			}
			got := ds.GetFirstDatasetId()
			assert.Equal(t, tt.expectedId, got)
		})
	}
}

func TestDocSourceField_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    *DocSourceField
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "valid field",
			field: &DocSourceField{
				Name:   "field1",
				Path:   "path/to/file",
				Source: "gns://abc123/def456",
				Type:   cdaenum.DocSourceFieldTypeFile,
			},
			wantErr: false,
		},
		{
			name:    "nil field",
			field:   nil,
			wantErr: true,
		},
		{
			name: "empty name",
			field: &DocSourceField{
				Name:   "",
				Path:   "path1",
				Source: "source1",
				Type:   cdaenum.DocSourceFieldTypeFile,
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "name is required")
			},
		},
		{
			name: "empty path",
			field: &DocSourceField{
				Name:   "field1",
				Path:   "",
				Source: "source1",
				Type:   cdaenum.DocSourceFieldTypeFile,
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "path is required")
			},
		},
		{
			name: "empty source",
			field: &DocSourceField{
				Name:   "field1",
				Path:   "path1",
				Source: "",
				Type:   cdaenum.DocSourceFieldTypeFile,
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "source is required")
			},
		},
		{
			name: "invalid type",
			field: &DocSourceField{
				Name:   "field1",
				Path:   "path1",
				Source: "gns://abc123/def456",
				Type:   cdaenum.DocSourceFieldType("invalid"),
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "type is invalid")
			},
		},
		{
			name: "empty type",
			field: &DocSourceField{
				Name:   "field1",
				Path:   "path1",
				Source: "gns://abc123/def456",
				Type:   "",
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "type is required")
			},
		},
		{
			name: "invalid source - ends with slash",
			field: &DocSourceField{
				Name:   "field1",
				Path:   "path1",
				Source: "gns://abc123/",
				Type:   cdaenum.DocSourceFieldTypeFile,
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "source is invalid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.field.ValObjCheck()
			if tt.wantErr {
				require.Error(t, err)

				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDocSourceField_GetDirObjID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		source     string
		expectedID string
	}{
		{
			name:       "simple source",
			source:     "gns://abc123/def456",
			expectedID: "def456",
		},
		{
			name:       "source with multiple slashes",
			source:     "gns://abc123/def456/ghi789",
			expectedID: "ghi789",
		},
		{
			name:       "source without slash",
			source:     "nofilename",
			expectedID: "nofilename",
		},
		{
			name:       "empty source",
			source:     "",
			expectedID: "",
		},
		{
			name:       "source with trailing slash",
			source:     "gns://abc123/def456/",
			expectedID: "",
		},
		{
			name:       "source with only slashes",
			source:     "///",
			expectedID: "",
		},
		{
			name:       "single slash",
			source:     "/",
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			field := &DocSourceField{
				Source: tt.source,
			}
			got := field.GetDirObjID()
			assert.Equal(t, tt.expectedID, got)
		})
	}
}
