package datasourcevalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnEntrySource_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	knEntry := &KnEntrySource{
		KnEntryID: "entry-123",
	}

	err := knEntry.ValObjCheck()
	assert.NoError(t, err)
}

func TestKnEntrySource_ValObjCheck_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kn   *KnEntrySource
	}{
		{
			name: "empty kn entry id",
			kn: &KnEntrySource{
				KnEntryID: "",
			},
		},
		{
			name: "nil struct",
			kn:   &KnEntrySource{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.kn.ValObjCheck()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "kn_entry_id is required")
		})
	}
}

func TestKnEntrySource_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	knEntry := &KnEntrySource{}
	errMap := knEntry.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Len(t, errMap, 1)
	assert.Equal(t, `"kn_entry_id"不能为空`, errMap["KnEntryID.required"])
}

func TestKnEntrySource_Fields(t *testing.T) {
	t.Parallel()

	knEntry := &KnEntrySource{
		KnEntryID: "entry-456",
	}

	assert.Equal(t, "entry-456", knEntry.KnEntryID)
}
