package datasourcevalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKgSource_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		kg       *KgSource
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "valid kg source",
			kg: &KgSource{
				KgID:   "kg-1",
				Fields: []string{"field1", "field2"},
			},
			wantErr:  false,
			checkErr: nil,
		},
		{
			name:    "nil kg source",
			kg:      nil,
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "empty kg id",
			kg: &KgSource{
				KgID:   "",
				Fields: []string{"field1"},
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "kg_id is required")
			},
		},
		{
			name: "nil fields",
			kg: &KgSource{
				KgID:   "kg-1",
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
			kg: &KgSource{
				KgID:   "kg-1",
				Fields: []string{},
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "fields is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.kg.ValObjCheck()
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

func TestKgSource_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	kg := &KgSource{}

	msgMap := kg.GetErrMsgMap()

	assert.NotNil(t, msgMap)
	assert.Equal(t, `"kg_id"不能为空`, msgMap["KgID.required"])
	assert.Equal(t, `"fields"不能为空`, msgMap["Fields.required"])
}
