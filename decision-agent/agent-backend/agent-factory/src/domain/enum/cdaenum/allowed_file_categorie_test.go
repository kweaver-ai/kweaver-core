package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllowedFileCategory_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		c       AllowedFileCategory
		wantErr bool
	}{
		{
			name:    "文档",
			c:       AllowedFileCategory("document"),
			wantErr: false,
		},
		{
			name:    "电子表格",
			c:       AllowedFileCategory("spreadsheet"),
			wantErr: false,
		},
		{
			name:    "演示文稿",
			c:       AllowedFileCategory("presentation"),
			wantErr: false,
		},
		{
			name:    "PDF",
			c:       AllowedFileCategory("pdf"),
			wantErr: false,
		},
		{
			name:    "文本",
			c:       AllowedFileCategory("text"),
			wantErr: false,
		},
		{
			name:    "音频",
			c:       AllowedFileCategory("audio"),
			wantErr: false,
		},
		{
			name:    "视频",
			c:       AllowedFileCategory("video"),
			wantErr: false,
		},
		{
			name:    "其他",
			c:       AllowedFileCategory("other"),
			wantErr: false,
		},
		{
			name:    "Wiki文档",
			c:       AllowedFileCategory("wikidoc"),
			wantErr: false,
		},
		{
			name:    "FAQ",
			c:       AllowedFileCategory("faq"),
			wantErr: false,
		},
		{
			name:    "无效类别",
			c:       AllowedFileCategory("invalid"),
			wantErr: true,
		},
		{
			name:    "空字符串",
			c:       AllowedFileCategory(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cats := AllowedFileCategories{tt.c}

			err := cats.EnumCheck()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

func TestAllowedFileCategories_GetAllowedFileTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		categories AllowedFileCategories
		wantTypes  []string
		wantErr    bool
	}{
		{
			name:       "文档类别",
			categories: AllowedFileCategories{"document"},
			wantTypes:  []string{"docx", "dotx", "dot", "doc", "odt", "wps", "docm", "dotm"},
			wantErr:    false,
		},
		{
			name:       "电子表格类别",
			categories: AllowedFileCategories{"spreadsheet"},
			wantTypes:  []string{"xlsx", "xlsm", "xlsb", "xls", "et", "xla", "xlam", "xltm", "xltx", "xlt", "ods", "csv"},
			wantErr:    false,
		},
		{
			name:       "演示文稿类别",
			categories: AllowedFileCategories{"presentation"},
			wantTypes:  []string{"pptx", "ppt", "pot", "pps", "ppsx", "dps", "potm", "ppsm", "potx", "pptm", "odp"},
			wantErr:    false,
		},
		{
			name:       "PDF类别",
			categories: AllowedFileCategories{"pdf"},
			wantTypes:  []string{"pdf"},
			wantErr:    false,
		},
		{
			name:       "文本类别",
			categories: AllowedFileCategories{"text"},
			wantTypes:  []string{"txt", "html"},
			wantErr:    false,
		},
		{
			name:       "音频类别",
			categories: AllowedFileCategories{"audio"},
			wantTypes:  []string{"aac", "ape", "flac", "m4a", "mp3", "wav", "wma", "ogg"},
			wantErr:    false,
		},
		{
			name:       "视频类别",
			categories: AllowedFileCategories{"video"},
			wantTypes:  []string{"3gp", "avi", "asf", "flv", "mov", "m2ts", "mkv", "mp4", "mpeg", "mpg", "mts", "rm", "rmvb", "wmv"},
			wantErr:    false,
		},
		{
			name:       "其他类别",
			categories: AllowedFileCategories{"other"},
			wantTypes:  []string{"dwg"},
			wantErr:    false,
		},
		{
			name:       "Wiki文档类别",
			categories: AllowedFileCategories{"wikidoc"},
			wantTypes:  []string{"wikidoc"},
			wantErr:    false,
		},
		{
			name:       "FAQ类别",
			categories: AllowedFileCategories{"faq"},
			wantTypes:  []string{"faq"},
			wantErr:    false,
		},
		{
			name:       "无效类别",
			categories: AllowedFileCategories{"invalid"},
			wantTypes:  nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			types, err := tt.categories.GetAllowedFileTypes()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.ElementsMatch(t, tt.wantTypes, types)
			}
		})
	}
}

func TestGetFileExtMap(t *testing.T) {
	t.Parallel()

	fileMap := GetFileExtMap()

	tests := []struct {
		name        string
		category    string
		wantPresent bool
	}{
		{
			name:        "文档类别",
			category:    "document",
			wantPresent: true,
		},
		{
			name:        "电子表格类别",
			category:    "spreadsheet",
			wantPresent: true,
		},
		{
			name:        "演示文稿类别",
			category:    "presentation",
			wantPresent: true,
		},
		{
			name:        "PDF类别",
			category:    "pdf",
			wantPresent: true,
		},
		{
			name:        "文本类别",
			category:    "text",
			wantPresent: true,
		},
		{
			name:        "音频类别",
			category:    "audio",
			wantPresent: true,
		},
		{
			name:        "视频类别",
			category:    "video",
			wantPresent: true,
		},
		{
			name:        "其他类别",
			category:    "other",
			wantPresent: true,
		},
		{
			name:        "Wiki文档类别",
			category:    "wikidoc",
			wantPresent: true,
		},
		{
			name:        "FAQ类别",
			category:    "faq",
			wantPresent: true,
		},
		{
			name:        "不存在的类别",
			category:    "nonexistent",
			wantPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, present := fileMap[tt.category]
			assert.Equal(t, tt.wantPresent, present, "category presence should match expected")
		})
	}
}
