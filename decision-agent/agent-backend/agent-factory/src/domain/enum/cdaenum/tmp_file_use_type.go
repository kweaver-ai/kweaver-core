package cdaenum

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

// TmpFileUseType 临时文件使用类型
type TmpFileUseType string

const (
	// TmpFileUseTypeUpload 直接上传
	TmpFileUseTypeUpload TmpFileUseType = "upload"

	// TmpFileUseTypeSelectFromTempZone 从临时区选择
	TmpFileUseTypeSelectFromTempZone TmpFileUseType = "select_from_temp_zone"
)

func (t TmpFileUseType) EnumCheck() (err error) {
	if !cutil.ExistsGeneric([]TmpFileUseType{TmpFileUseTypeUpload, TmpFileUseTypeSelectFromTempZone}, t) {
		err = errors.New("[TmpFileUseType]: invalid tmp_file_use_type")
		return
	}

	return
}
