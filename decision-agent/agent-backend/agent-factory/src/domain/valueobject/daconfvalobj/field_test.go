package daconfvalobj

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestField_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	field := &Field{}
	errMsgMap := field.GetErrMsgMap()
	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"name"不能为空`, errMsgMap["Name.required"])
}

func TestField_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		field   *Field
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Field (String type)",
			field: &Field{
				Name: "param1",
				Type: cdaenum.InputFieldTypeString,
				Desc: "A string parameter",
			},
			wantErr: false,
		},
		{
			name: "Valid Field (File type)",
			field: &Field{
				Name: "param2",
				Type: cdaenum.InputFieldTypeFile,
				Desc: "A file parameter",
			},
			wantErr: false,
		},
		{
			name: "Missing Name",
			field: &Field{
				Name: "",
				Type: cdaenum.InputFieldTypeString,
			},
			wantErr: true,
			errMsg:  "[Field]: name is required",
		},
		{
			name: "Invalid Type",
			field: &Field{
				Name: "param3",
				Type: cdaenum.InputFieldType("invalid_type"),
			},
			wantErr: true,
			errMsg:  "[Field]: type is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.field.ValObjCheck()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
