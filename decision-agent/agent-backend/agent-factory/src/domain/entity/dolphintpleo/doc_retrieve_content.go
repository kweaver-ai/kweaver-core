package dolphintpleo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
)

type DocRetrieveContent struct {
	Content  string `json:"content"`
	IsEnable bool   `json:"is_enable"`
}

func NewDocRetrieveContent() *DocRetrieveContent {
	return &DocRetrieveContent{
		Content:  "",
		IsEnable: false,
	}
}

func (d *DocRetrieveContent) LoadFromConfig(config *daconfvalobj.Config, isBuiltInDocQAAgent bool) {
	if isBuiltInDocQAAgent {
		d.IsEnable = false
		d.Content = ""

		return
	}

	if config.DataSource != nil && len(config.DataSource.Doc) > 0 {
		d.IsEnable = true
		d.Content = `
/judge/(tools=["doc_qa"], history=True)判断【$query】是否需要到文档中召回，如果不需要召回，则直接返回\"不需要文档召回\"，否则执行工具对【$query】进行召回 -> doc_retrieval_res
`
	}
}

func (d *DocRetrieveContent) ToString() (str string) {
	str = d.Content
	return
}

func (d *DocRetrieveContent) ToDolphinTplEo() *DolphinTplEo {
	key := cdaenum.DolphinTplKeyDocRetrieve

	return &DolphinTplEo{
		Key:   key,
		Name:  key.GetName(),
		Value: d.ToString(),
	}
}
