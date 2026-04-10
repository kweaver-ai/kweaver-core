package cdaenum

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

// 文档源字段类型
type DocSourceFieldType string

const (
	DocSourceFieldTypeFolder DocSourceFieldType = "folder"
	DocSourceFieldTypeFile   DocSourceFieldType = "file"
)

func (b DocSourceFieldType) EnumCheck() (err error) {
	if !cutil.ExistsGeneric([]DocSourceFieldType{DocSourceFieldTypeFolder, DocSourceFieldTypeFile}, b) {
		err = errors.New("[DocSourceFieldType]: invalid doc source field type")
		return
	}

	return
}
