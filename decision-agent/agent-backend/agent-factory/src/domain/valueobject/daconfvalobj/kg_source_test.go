package daconfvalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKgSource_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kg   *KgSource
	}{
		{
			name: "valid kg source with fields",
			kg: &KgSource{
				KgID:   "kg-123",
				Fields: []string{"field1", "field2"},
			},
		},
		{
			name: "valid kg source with empty fields slice",
			kg: &KgSource{
				KgID:   "kg-456",
				Fields: []string{},
			},
		},
		{
			name: "valid kg source with all fields",
			kg: &KgSource{
				KgID:            "kg-789",
				Fields:          []string{"Person", "Company"},
				FieldProperties: map[string][]string{"Person": {"name", "age"}},
				OutputFields:    []string{"name"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.kg.ValObjCheck()
			assert.NoError(t, err)
		})
	}
}

func TestKgSource_ValObjCheck_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		kg          *KgSource
		expectedErr string
	}{
		{
			name: "empty kg_id",
			kg: &KgSource{
				KgID:   "",
				Fields: []string{"field1"},
			},
			expectedErr: "kg_id is required",
		},
		{
			name: "nil fields",
			kg: &KgSource{
				KgID:   "kg-123",
				Fields: nil,
			},
			expectedErr: "fields is required",
		},
		{
			name:        "both empty kg_id and nil fields",
			kg:          &KgSource{},
			expectedErr: "kg_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.kg.ValObjCheck()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestKgSource_ValObjCheck_Nil(t *testing.T) {
	t.Parallel()

	var kg *KgSource
	// Nil pointer will panic, so we test for that
	assert.Panics(t, func() {
		kg.ValObjCheck() //nolint:errcheck
	})
}

func TestKgSource_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	kg := &KgSource{}
	errMap := kg.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Len(t, errMap, 2)
	assert.Equal(t, `"kg_id"不能为空`, errMap["KgID.required"])
	assert.Equal(t, `"fields"不能为空`, errMap["Fields.required"])
}

func TestKgSource_Fields(t *testing.T) {
	t.Parallel()

	fieldProperties := map[string][]string{
		"Person":  {"name", "age", "email"},
		"Company": {"name", "address"},
	}

	kg := &KgSource{
		KgID:            "kg-test",
		Fields:          []string{"Person", "Company"},
		FieldProperties: fieldProperties,
		OutputFields:    []string{"name", "address"},
	}

	assert.Equal(t, "kg-test", kg.KgID)
	assert.Len(t, kg.Fields, 2)
	assert.Equal(t, "Person", kg.Fields[0])
	assert.Equal(t, "Company", kg.Fields[1])
	assert.Len(t, kg.FieldProperties, 2)
	assert.Len(t, kg.OutputFields, 2)
}
