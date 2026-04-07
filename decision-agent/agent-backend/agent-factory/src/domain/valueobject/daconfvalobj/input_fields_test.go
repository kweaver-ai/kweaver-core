package daconfvalobj

import (
	"strings"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
)

func TestFields_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fields  Fields
		wantErr bool
	}{
		{
			name: "有效字段",
			fields: Fields{
				&Field{Name: "field1", Type: cdaenum.InputFieldTypeString},
				&Field{Name: "field2", Type: cdaenum.InputFieldTypeJSONObject},
			},
			wantErr: false,
		},
		{
			name:    "空字段列表",
			fields:  Fields{},
			wantErr: false,
		},
		{
			name: "重名字段",
			fields: Fields{
				&Field{Name: "field1", Type: cdaenum.InputFieldTypeString},
				&Field{Name: "field1", Type: cdaenum.InputFieldTypeJSONObject},
			},
			wantErr: true,
		},
		{
			name: "多个文件类型字段",
			fields: Fields{
				&Field{Name: "file1", Type: cdaenum.InputFieldTypeFile},
				&Field{Name: "file2", Type: cdaenum.InputFieldTypeFile},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.fields.ValObjCheck()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValObjCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFields_IsEnabledTempZone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		fields Fields
		want   bool
	}{
		{
			name: "包含文件类型字段",
			fields: Fields{
				&Field{Name: "file1", Type: cdaenum.InputFieldTypeFile},
				&Field{Name: "text1", Type: cdaenum.InputFieldTypeString},
			},
			want: true,
		},
		{
			name: "不包含文件类型字段",
			fields: Fields{
				&Field{Name: "text1", Type: cdaenum.InputFieldTypeString},
				&Field{Name: "obj1", Type: cdaenum.InputFieldTypeJSONObject},
			},
			want: false,
		},
		{
			name:   "空字段列表",
			fields: Fields{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.fields.IsEnabledTempZone(); got != tt.want {
				t.Errorf("IsEnabledTempZone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFields_IsFieldNameRepeat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		fields Fields
		want   bool
	}{
		{
			name: "无重名字段",
			fields: Fields{
				&Field{Name: "field1"},
				&Field{Name: "field2"},
				&Field{Name: "field3"},
			},
			want: false,
		},
		{
			name: "有重名字段",
			fields: Fields{
				&Field{Name: "field1"},
				&Field{Name: "field2"},
				&Field{Name: "field1"},
			},
			want: true,
		},
		{
			name:   "空字段列表",
			fields: Fields{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.fields.IsFieldNameRepeat(); got != tt.want {
				t.Errorf("IsFieldNameRepeat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFields_GenNotFileDolphinStr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		fields         Fields
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "普通字段",
			fields: Fields{
				&Field{Name: "field1", Type: cdaenum.InputFieldTypeString},
				&Field{Name: "field2", Type: cdaenum.InputFieldTypeJSONObject},
			},
			wantContains:   []string{`"field1: " + $field1`, `"field2: " + $field2`, " -> all_inputs \n"},
			wantNotContain: []string{},
		},
		{
			name: "包含文件类型字段",
			fields: Fields{
				&Field{Name: "file1", Type: cdaenum.InputFieldTypeFile},
				&Field{Name: "text1", Type: cdaenum.InputFieldTypeString},
			},
			wantContains:   []string{`"text1: " + $text1`, " -> all_inputs \n"},
			wantNotContain: []string{`file1`},
		},
		{
			name: "包含系统保留字段",
			fields: Fields{
				&Field{Name: "history", Type: cdaenum.InputFieldTypeString},
				&Field{Name: "tool", Type: cdaenum.InputFieldTypeString},
				&Field{Name: "field1", Type: cdaenum.InputFieldTypeString},
			},
			wantContains:   []string{`"field1: " + $field1`, " -> all_inputs \n"},
			wantNotContain: []string{`history`, `tool`},
		},
		{
			name:           "只有文件字段",
			fields:         Fields{&Field{Name: "file1", Type: cdaenum.InputFieldTypeFile}},
			wantContains:   []string{" -> all_inputs \n"},
			wantNotContain: []string{`file1`},
		},
		{
			name:           "空字段列表",
			fields:         Fields{},
			wantContains:   []string{" -> all_inputs \n"},
			wantNotContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.fields.GenNotFileDolphinStr()
			for _, s := range tt.wantContains {
				if !strings.Contains(got, s) {
					t.Errorf("GenNotFileDolphinStr() 应包含: %s, 实际: %s", s, got)
				}
			}

			for _, s := range tt.wantNotContain {
				if strings.Contains(got, s) {
					t.Errorf("GenNotFileDolphinStr() 不应包含: %s, 实际: %s", s, got)
				}
			}
		})
	}
}

func TestFields_GenFileDolphinStr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		fields         Fields
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "包含文件类型字段",
			fields: Fields{
				&Field{Name: "file1", Type: cdaenum.InputFieldTypeFile},
				&Field{Name: "text1", Type: cdaenum.InputFieldTypeString},
			},
			wantContains:   []string{`@process_file_intelligent(query=$query, file_infos=$file1)`},
			wantNotContain: []string{`text1`},
		},
		{
			name:           "不包含文件类型字段",
			fields:         Fields{&Field{Name: "text1", Type: cdaenum.InputFieldTypeString}},
			wantContains:   []string{},
			wantNotContain: []string{`process_file_intelligent`},
		},
		{
			name:           "空字段列表",
			fields:         Fields{},
			wantContains:   []string{},
			wantNotContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, filedName := tt.fields.GenFileDolphinStr()
			for _, s := range tt.wantContains {
				if !strings.Contains(got, s) {
					t.Errorf("GenFileDolphinStr() 应包含: %s, 实际: %s", s, got)
				}
			}

			for _, s := range tt.wantNotContain {
				if strings.Contains(got, s) {
					t.Errorf("GenFileDolphinStr() 不应包含: %s, 实际: %s", s, got)
				}
			}

			if tt.fields.IsEnabledTempZone() && filedName == "" {
				t.Errorf("GenFileDolphinStr() 应返回字段名")
			}
		})
	}
}
