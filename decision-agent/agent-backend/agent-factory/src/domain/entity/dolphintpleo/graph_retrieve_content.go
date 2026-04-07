package dolphintpleo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
)

type GraphRetrieveContent struct {
	Content  string `json:"content"`
	IsEnable bool   `json:"is_enable"`
}

func NewGraphRetrieveContent() *GraphRetrieveContent {
	return &GraphRetrieveContent{
		Content:  "",
		IsEnable: false,
	}
}

func (g *GraphRetrieveContent) LoadFromConfig(config *daconfvalobj.Config, isBuiltInGraphQAAgent bool) {
	if isBuiltInGraphQAAgent {
		g.IsEnable = false
		g.Content = ""

		return
	}

	if config.DataSource != nil && len(config.DataSource.Kg) > 0 {
		g.IsEnable = true
		g.Content = `
/judge/(tools=["graph_qa"], history=True)判断【$query】是否需要到业务知识网络中召回，如果不需要召回，则直接返回“不需要业务知识网络召回”，否则执行工具对【$query】进行召回 -> graph_retrieval_res
`
	}
}

func (g *GraphRetrieveContent) ToString() (str string) {
	str = g.Content
	return
}

func (g *GraphRetrieveContent) ToDolphinTplEo() *DolphinTplEo {
	key := cdaenum.DolphinTplKeyGraphRetrieve

	return &DolphinTplEo{
		Key:   key,
		Name:  key.GetName(),
		Value: g.ToString(),
	}
}
