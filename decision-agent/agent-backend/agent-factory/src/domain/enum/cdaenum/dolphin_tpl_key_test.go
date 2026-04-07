package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDolphinTplKey_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		b       DolphinTplKey
		wantErr bool
	}{
		{
			name:    "记忆召回",
			b:       DolphinTplKeyMemoryRetrieve,
			wantErr: false,
		},
		{
			name:    "临时文件处理",
			b:       DolphinTplKeyTempFileProcess,
			wantErr: false,
		},
		{
			name:    "文档召回",
			b:       DolphinTplKeyDocRetrieve,
			wantErr: false,
		},
		{
			name:    "图谱召回",
			b:       DolphinTplKeyGraphRetrieve,
			wantErr: false,
		},
		{
			name:    "上下文组织",
			b:       DolphinTplKeyContextOrganize,
			wantErr: false,
		},
		{
			name:    "相关问题",
			b:       DolphinTplKeyRelatedQuestions,
			wantErr: false,
		},
		{
			name:    "无效key",
			b:       DolphinTplKey("invalid"),
			wantErr: true,
		},
		{
			name:    "空字符串",
			b:       DolphinTplKey(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.b.EnumCheck()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

func TestDolphinTplKey_GetName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		b    DolphinTplKey
		want string
	}{
		{
			name: "记忆召回",
			b:    DolphinTplKeyMemoryRetrieve,
			want: "记忆召回模块",
		},
		{
			name: "临时文件处理",
			b:    DolphinTplKeyTempFileProcess,
			want: "临时区文件处理模块",
		},
		{
			name: "文档召回",
			b:    DolphinTplKeyDocRetrieve,
			want: "文档召回模块",
		},
		{
			name: "图谱召回",
			b:    DolphinTplKeyGraphRetrieve,
			want: "业务知识网络召回模块",
		},
		{
			name: "上下文组织",
			b:    DolphinTplKeyContextOrganize,
			want: "上下文组织模块",
		},
		{
			name: "相关问题",
			b:    DolphinTplKeyRelatedQuestions,
			want: "相关问题模块",
		},
		{
			name: "无效key",
			b:    DolphinTplKey("invalid"),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.b.GetName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDolphinTplKey_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		b    DolphinTplKey
		want string
	}{
		{
			name: "记忆召回",
			b:    DolphinTplKeyMemoryRetrieve,
			want: "memory_retrieve",
		},
		{
			name: "临时文件处理",
			b:    DolphinTplKeyTempFileProcess,
			want: "temp_file_process",
		},
		{
			name: "文档召回",
			b:    DolphinTplKeyDocRetrieve,
			want: "doc_retrieve",
		},
		{
			name: "图谱召回",
			b:    DolphinTplKeyGraphRetrieve,
			want: "graph_retrieve",
		},
		{
			name: "上下文组织",
			b:    DolphinTplKeyContextOrganize,
			want: "context_organize",
		},
		{
			name: "相关问题",
			b:    DolphinTplKeyRelatedQuestions,
			want: "related_questions",
		},
		{
			name: "自定义key",
			b:    DolphinTplKey("custom_key"),
			want: "custom_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.b.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
