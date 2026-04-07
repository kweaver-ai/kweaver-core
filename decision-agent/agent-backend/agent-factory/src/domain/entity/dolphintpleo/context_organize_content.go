package dolphintpleo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
)

type OtherTplStruct struct {
	MemoryRetrieve *MemoryRetrieveContent `json:"memory_retrieve"`
	DocRetrieve    *DocRetrieveContent    `json:"doc_retrieve"`
	GraphRetrieve  *GraphRetrieveContent  `json:"graph_retrieve"`
}

type ContextOrganizeContent struct {
	OtherTplStruct *OtherTplStruct `json:"other_tpl_struct"`

	DocRetrieveContent   string `json:"doc_retrieve_content"`
	GraphRetrieveContent string `json:"graph_retrieve_content"`
	TempZoneContent      string `json:"temp_zone_content"`
	Other                string `json:"other"`
	IsEnable             bool   `json:"is_enable"`

	referenceEnable bool
}

func NewContextOrganizeContent(otherTplStruct *OtherTplStruct) *ContextOrganizeContent {
	return &ContextOrganizeContent{
		OtherTplStruct: otherTplStruct,
	}
}

func (c *ContextOrganizeContent) LoadFromConfig(config *daconfvalobj.Config) {
	// 如果下面DocRetrieve、GraphRetrieve、TempFileProcess这三个dolphin tpl有一个开启，则referenceEnable为true
	referenceEnable := false

	// 1. 根据config中的pre_dolphin和post_dolphin，判断一个dolphin tpl是否被用户禁用
	isDocRetrieveDisabledFromConfig := config.IsOneDolphinTplDisabled(cdaenum.DolphinTplKeyDocRetrieve)
	isGraphRetrieveDisabledFromConfig := config.IsOneDolphinTplDisabled(cdaenum.DolphinTplKeyGraphRetrieve)

	// 2. 如果配置了文档召回数据源，并没有禁用文档召回dolphin tpl，则添加文档召回内容
	if c.OtherTplStruct.DocRetrieve.IsEnable && !isDocRetrieveDisabledFromConfig {
		c.DocRetrieveContent = `
/if/ "result" in $doc_retrieval_res['answer'] and $doc_retrieval_res['answer']['result']:
    $reference + "文档召回的内容：" + $doc_retrieval_res['answer']['result'] + "\n" -> reference
/end/
`
		referenceEnable = true
	}

	// 3. 如果配置了业务知识网络召回数据源，并没有禁用业务知识网络召回dolphin tpl，则添加业务知识网络召回内容
	if c.OtherTplStruct.GraphRetrieve.IsEnable && !isGraphRetrieveDisabledFromConfig {
		c.GraphRetrieveContent = `
/if/ "result" in $graph_retrieval_res['answer'] and $graph_retrieval_res['answer']['result']:
    $reference + "业务知识网络召回的内容：" + $graph_retrieval_res['answer']['result'] + "\n" -> reference
/end/
`
		referenceEnable = true
	}

	// 5. 如果开启了参考文档，则添加参考文档内容
	if referenceEnable {
		c.Other = `
{"reference": $reference, "query": "用户的问题为: "+$query} -> context
`
	} else {
		c.Other = `
{"query": "用户的问题为: "+$query} -> context
`
	}

	// 因为始终有内容，所以这里直接设置为true
	c.IsEnable = true

	c.referenceEnable = referenceEnable
}

func (c *ContextOrganizeContent) ToString() (str string) {
	if c.referenceEnable {
		str = `
"如果有参考文档，结合参考文档回答用户的问题。如果没有参考文档，根据用户的问题回答。\n" -> reference
`
	}

	if c.DocRetrieveContent != "" {
		str += c.DocRetrieveContent
	}

	if c.GraphRetrieveContent != "" {
		str += c.GraphRetrieveContent
	}

	if c.TempZoneContent != "" {
		str += c.TempZoneContent
	}

	if c.Other != "" {
		str += c.Other
	}

	return
}

func (c *ContextOrganizeContent) ToDolphinTplEo() *DolphinTplEo {
	key := cdaenum.DolphinTplKeyContextOrganize

	return &DolphinTplEo{
		Key:   key,
		Name:  key.GetName(),
		Value: c.ToString(),
	}
}
