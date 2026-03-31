package skillenum

import (
	"github.com/pkg/errors"
)

type Datasource string

const (
	DatasourceInheritMain    Datasource = "inherit_main"    // 继承主 Agent 数据源
	DatasourceSelfConfigured Datasource = "self_configured" // 使用自身配置（默认逻辑）
)

func (b Datasource) EnumCheck() (err error) {
	if b != DatasourceInheritMain && b != DatasourceSelfConfigured {
		err = errors.New("[Datasource]: invalid skill agent datasource")
		return
	}

	return
}

// 仅当 type = inherit_main 时生效，指定继承范围：
type DatasourceSpecificInherit string

const (
	DatasourceInheritDocs  DatasourceSpecificInherit = "docs_only"  // docs_only：仅继承文档数据源
	DatasourceInheritGraph DatasourceSpecificInherit = "graph_only" // graph_only：仅继承图谱数据源
	DatasourceInheritAll   DatasourceSpecificInherit = "all"        // all：继承所有类型数据源
)

func (b DatasourceSpecificInherit) EnumCheck() (err error) {
	if b != DatasourceInheritDocs && b != DatasourceInheritGraph && b != DatasourceInheritAll {
		err = errors.New("数据源继承类型不合法")
	}

	return
}
