package daconfvalobj

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestInput_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	input := &Input{}
	msgMap := input.GetErrMsgMap()

	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "Fields.required",
			key:  "Fields.required",
			want: `"fields"不能为空`,
		},
		{
			name: "不存在的key",
			key:  "Unknown.key",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := msgMap[tt.key]
			if got != tt.want {
				t.Errorf("GetErrMsgMap()[%q] = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestInput_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   *Input
		wantErr bool
	}{
		{
			name: "有效配置",
			input: &Input{
				Fields: Fields{
					&Field{Name: "field1", Type: cdaenum.InputFieldTypeString},
				},
			},
			wantErr: false,
		},
		{
			name: "Fields为空",
			input: &Input{
				Fields: nil,
			},
			wantErr: true,
		},
		{
			name: "包含Rewrite且有效",
			input: &Input{
				Fields: Fields{
					&Field{Name: "field1", Type: cdaenum.InputFieldTypeString},
				},
				Rewrite: &Rewrite{
					Enable: func() *bool { b := false; return &b }(),
				},
			},
			wantErr: false,
		},
		{
			name: "包含Augment且有效",
			input: &Input{
				Fields: Fields{
					&Field{Name: "field1", Type: cdaenum.InputFieldTypeString},
				},
				Augment: &Augment{
					Enable: func() *bool { b := false; return &b }(),
					DataSource: &AugmentDataSource{
						Kg: []KgSource{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "所有配置有效",
			input: &Input{
				Fields: Fields{
					&Field{Name: "field1", Type: cdaenum.InputFieldTypeString},
				},
				Rewrite: &Rewrite{
					Enable: func() *bool { b := false; return &b }(),
				},
				Augment: &Augment{
					Enable: func() *bool { b := false; return &b }(),
					DataSource: &AugmentDataSource{
						Kg: []KgSource{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Fields验证失败",
			input: &Input{
				Fields: Fields{
					&Field{Name: "", Type: cdaenum.InputFieldTypeString}, // Name is empty
				},
			},
			wantErr: true,
		},
		{
			name: "Rewrite验证失败",
			input: &Input{
				Fields: Fields{
					&Field{Name: "field1", Type: cdaenum.InputFieldTypeString},
				},
				Rewrite: &Rewrite{
					Enable: func() *bool { b := true; return &b }(),
					// Pattern is empty when Enable is true
				},
			},
			wantErr: true,
		},
		{
			name: "Augment验证失败",
			input: &Input{
				Fields: Fields{
					&Field{Name: "field1", Type: cdaenum.InputFieldTypeString},
				},
				Augment: &Augment{
					Enable: func() *bool { b := true; return &b }(),
					// DataSource is nil when Enable is true
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.input.ValObjCheck()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}
