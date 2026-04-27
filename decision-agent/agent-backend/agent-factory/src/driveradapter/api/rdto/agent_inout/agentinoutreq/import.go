package agentinoutreq

import (
	"mime/multipart"
)

// ImportReq 导入agent请求
type ImportReq struct {
	File       *multipart.FileHeader `form:"file" binding:"required" label:"导入文件"`
	ImportType ImportType            `form:"import_type" binding:"required" label:"导入类型"`
}

func NewImportReq() *ImportReq {
	return &ImportReq{}
}

// GetErrMsgMap 返回错误信息映射
func (r *ImportReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"File.required":       `"file"字段不能为空`,
		"ImportType.required": `"import_type"字段不能为空`,
	}
}

func (r *ImportReq) ValueCheck() (err error) {
	err = r.ImportType.EnumCheck()
	if err != nil {
		return
	}

	return
}
