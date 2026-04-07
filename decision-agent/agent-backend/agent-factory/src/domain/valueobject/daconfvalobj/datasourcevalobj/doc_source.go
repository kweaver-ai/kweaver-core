package datasourcevalobj

import (
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/pkg/errors"
)

// DocSource 表示文档类型数据源
type DocSource struct {
	DsID     string            `json:"ds_id" binding:"required"`  // 数据源ID
	Fields   []*DocSourceField `json:"fields" binding:"required"` // 数据源范围列表
	Datasets []string          `json:"datasets"`                  // 数据集列表
}

// ValObjCheck 验证数据源配置
func (p *DocSource) ValObjCheck() (err error) {
	if p == nil {
		err = errors.New("[DocSource]: cannot be nil")
		return
	}
	// 检查DsID是否为空
	if p.DsID == "" {
		err = errors.New("[DocSource]: ds_id is required")
		return
	}

	// 检查Fields是否为空
	if len(p.Fields) == 0 {
		err = errors.New("[DocSource]: fields is required")
		return
	}

	// 验证每个字段的有效性
	for _, field := range p.Fields {
		if err = field.ValObjCheck(); err != nil {
			// 包装错误信息，提供更详细的上下文
			err = errors.Wrap(err, "[DocSource]: field is invalid")
			return
		}
	}

	//// 检查Datasets是否为空
	//if len(p.Datasets) > 0 {
	//	//不需要通过接口来传递这个，这里置空。后面会有逻辑给这个赋值
	//	p.Datasets=[]string{}
	//	//err = errors.New("[DocSource]: datasets cannot be passed in by api")
	//	return
	//}

	return
}

func (p *DocSource) SetDatasetId(datasetId string) {
	if datasetId == "" {
		return
	}

	p.Datasets = append(p.Datasets, datasetId)
}

func (p *DocSource) GetFirstDatasetId() string {
	if len(p.Datasets) == 0 {
		return ""
	}

	return p.Datasets[0]
}

// DocSourceField 表示文档源中的字段
// 示例：
//
//	{
//		"name": "爱数介绍：Who are we (2025).pptx",
//		"path": "测试数据/爱数介绍：Who are we (2025).pptx",
//		"source": "gns://92EE2D87255142B78A6F1DFB6BBB836B/B08AC060A758422583A851C601C0A89B"
//	}
type DocSourceField struct {
	Name   string `json:"name" binding:"required"`   // 字段名称
	Path   string `json:"path" binding:"required"`   // 字段路径
	Source string `json:"source" binding:"required"` // 字段来源

	Type cdaenum.DocSourceFieldType `json:"type" binding:"required"` // 字段类型
}

// ValObjCheck 验证文档源字段
func (p *DocSourceField) ValObjCheck() (err error) {
	if p == nil {
		err = errors.New("[DocSourceField]: cannot be nil")
		return
	}
	// 检查Name是否为空
	if p.Name == "" {
		err = errors.New("[DocSourceField]: name is required")
		return
	}

	// 检查Path是否为空
	if p.Path == "" {
		err = errors.New("[DocSourceField]: path is required")
		return
	}

	// 检查Source是否为空
	if p.Source == "" {
		err = errors.New("[DocSourceField]: source is required")
		return
	}

	// 检查Type是否为空
	if p.Type == "" {
		err = errors.New("[DocSourceField]: type is required")
		return
	}

	// 验证Type枚举值的有效性
	if err = p.Type.EnumCheck(); err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[DocSourceField]: type is invalid")
		return
	}

	// 检查Source的格式是否正确
	id := p.GetDirObjID()
	if id == "" {
		err = errors.New("[DocSourceField]: source is invalid")
		return
	}

	return
}

func (p *DocSourceField) GetDirObjID() string {
	// source的最后一段
	parts := strings.Split(p.Source, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}
