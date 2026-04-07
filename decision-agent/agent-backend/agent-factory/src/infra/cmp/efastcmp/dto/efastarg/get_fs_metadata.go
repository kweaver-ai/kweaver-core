package efastarg

import (
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// IbField 查询哪一个信息（查询字段）
// IB: item batch
type IbField string

type IbFields []IbField

func (f IbFields) ToPathString() string {
	tmp := make([]string, len(f))
	for i := range f {
		tmp[i] = string(f[i])
	}

	return strings.Join(tmp, ",")
}

// 可以参考efast的文档根据需要在此处添加（内部接口）
// http://{host}:{port}/api/efast/v1/items-batch/{fields}
const (
	IbFieldName        IbField = "names"
	IbFieldDocLibTypes IbField = "doc_lib_types"
	IbFieldPaths       IbField = "paths"
)

// GetFsMetadataArgDto 传过来的参数
type GetFsMetadataArgDto struct {
	IDs    []string `json:"ids"` // 文件或目录的id
	ObjIDs []string `json:"obj_ids"`
	Fields IbFields `json:"fields"`
}

// GetFsMetadataEFArgDto 访问efast的参数
// 由GetFsMetadataArgDto转换而来
type GetFsMetadataEFArgDto struct {
	IDs    []string `json:"ids,omitempty"`
	ObjIDs []string `json:"obj_ids,omitempty"`
	Method string   `json:"method"`
}

func NewGetFsMetadataEFArgDto(dto *GetFsMetadataArgDto) *GetFsMetadataEFArgDto {
	var ids []string
	if len(dto.IDs) != 0 {
		ids = cutil.DeduplGeneric(dto.IDs)

		return &GetFsMetadataEFArgDto{
			IDs:    ids,
			Method: "GET",
		}
	} else if len(dto.ObjIDs) != 0 {
		ids = cutil.DeduplGeneric(dto.ObjIDs)

		return &GetFsMetadataEFArgDto{
			ObjIDs: ids,
			Method: "GET",
		}
	}

	return &GetFsMetadataEFArgDto{}
}
