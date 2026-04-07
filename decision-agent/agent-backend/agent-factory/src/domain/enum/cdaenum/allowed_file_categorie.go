package cdaenum

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type AllowedFileCategory string

var AllAllowedFileCategories = []AllowedFileCategory{
	"document", "spreadsheet", "presentation", "pdf", "text", "audio", "other", "video", "wikidoc", "faq",
}

type AllowedFileCategories []AllowedFileCategory

func (c AllowedFileCategories) EnumCheck() (err error) {
	for _, category := range c {
		if !cutil.ExistsGeneric(AllAllowedFileCategories, category) {
			err = errors.New("[AllowedFileCategories]: invalid category")
			return
		}
	}

	return
}

func (c AllowedFileCategories) GetAllowedFileTypes() (allowedFileTypes []string, err error) {
	for _, category := range c {
		if fileTypes, ok := fileMap[string(category)]; ok {
			allowedFileTypes = append(allowedFileTypes, fileTypes...)
		} else {
			// 如果类别不存在，返回错误或处理逻辑
			err = errors.New("[AllowedFileCategories]: invalid category")
			return
		}
	}

	return
}

var fileMap = map[string][]string{
	"document":     document,
	"spreadsheet":  spreadsheet,
	"presentation": presentation,
	"text":         text,
	"audio":        audio,
	"pdf":          pdf,
	"other":        other,
	"video":        video,
	"wikidoc":      wikidoc,
	"faq":          faq,
}

var (
	document     = []string{"docx", "dotx", "dot", "doc", "odt", "wps", "docm", "dotm"}
	spreadsheet  = []string{"xlsx", "xlsm", "xlsb", "xls", "et", "xla", "xlam", "xltm", "xltx", "xlt", "ods", "csv"}
	presentation = []string{"pptx", "ppt", "pot", "pps", "ppsx", "dps", "potm", "ppsm", "potx", "pptm", "odp"}
	text         = []string{"txt", "html"}
	audio        = []string{"aac", "ape", "flac", "m4a", "mp3", "wav", "wma", "ogg"}
	pdf          = []string{"pdf"}
	other        = []string{"dwg"}
	video        = []string{"3gp", "avi", "asf", "flv", "mov", "m2ts", "mkv", "mp4", "mpeg", "mpg", "mts", "rm", "rmvb", "wmv"}
	wikidoc      = []string{"wikidoc"}
	faq          = []string{"faq"}
)

func GetFileExtMap() map[string][]string {
	return fileMap
}
