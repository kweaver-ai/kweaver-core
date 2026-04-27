package tplutils

import (
	"strings"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		templateStr string
		data        map[string]interface{}
		want        string
		wantErr     bool
	}{
		{
			name:        "基本变量替换",
			templateStr: "Hello, {{.Name}}!",
			data: map[string]interface{}{
				"Name": "张三",
			},
			want:    "Hello, 张三!",
			wantErr: false,
		},
		{
			name:        "条件语句",
			templateStr: "{{if .Show}}显示{{else}}隐藏{{end}}",
			data: map[string]interface{}{
				"Show": true,
			},
			want:    "显示",
			wantErr: false,
		},
		{
			name:        "循环语句",
			templateStr: "{{range .Items}}<li>{{.}}</li>{{end}}",
			data: map[string]interface{}{
				"Items": []string{"项目1", "项目2"},
			},
			want:    "<li>项目1</li><li>项目2</li>",
			wantErr: false,
		},
		{
			name:        "嵌套数据",
			templateStr: "{{.User.Name}} - {{.User.Age}}岁",
			data: map[string]interface{}{
				"User": map[string]interface{}{
					"Name": "李四",
					"Age":  25,
				},
			},
			want:    "李四 - 25岁",
			wantErr: false,
		},
		{
			name:        "模板语法错误",
			templateStr: "{{.Name}", // 缺少结束括号
			data: map[string]interface{}{
				"Name": "张三",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := RenderTemplate(tt.templateStr, tt.data)

			// 检查错误
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 如果期望错误，不需要检查结果
			if tt.wantErr {
				return
			}

			// 去除空白字符后比较结果
			got = strings.TrimSpace(got)
			want := strings.TrimSpace(tt.want)

			if got != want {
				t.Errorf("RenderTemplate() = %v, want %v", got, want)
			}
		})
	}
}
