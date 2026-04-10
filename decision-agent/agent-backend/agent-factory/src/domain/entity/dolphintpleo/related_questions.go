package dolphintpleo

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
)

type RelatedQuestionsContent struct {
	Content  string `json:"content"`
	IsEnable bool   `json:"is_enable"`
}

func NewRelatedQuestionsContent() *RelatedQuestionsContent {
	return &RelatedQuestionsContent{
		Content:  "",
		IsEnable: false,
	}
}

func (d *RelatedQuestionsContent) LoadFromConfig(config *daconfvalobj.Config) {
	if config != nil && config.RelatedQuestion != nil && config.RelatedQuestion.IsEnabled {
		d.IsEnable = true
		d.Content = `
/prompt/(flags='{"debug": true}')请根据原始用户问题和上下文信息，更进一步的生成3个问题和答案对。所生成的3个问题与原始用户问题呈递进关系，3个问题之间则相互独立，用户可以使用这些问题深入挖掘上下文中的话题。
===
$query
===
要求：
1. 根据上下文信息生成问题答案对，禁止推测；若上下文信息为空。则直接返回空。
2. 生成的问题不要和原始用户问题重复。
3. 确保生成的问题主谓宾完整，长度不超过25个字。
4. 不要输出问题对应的答案。
5. 输出格式为：
["第一个问题", "第二个问题", "第三个问题"]
6. 如果无法生成问题，则返回空列表 []
7. 示例:
原始问题:你能帮我写一首描写春天的唐诗吗？
相关问题:["春天的唐诗中有哪些典型的意象？", "这首唐诗中如何体现春天的生机与活力？", "唐诗中春天的描写与现代人对春天的感受有何不同？"]
8. 不要输出多余的内容。
-> related_questions
eval($related_questions.answer) -> related_questions
`
	}
}

func (d *RelatedQuestionsContent) ToString() (str string) {
	str = d.Content
	return
}

func (d *RelatedQuestionsContent) ToDolphinTplEo() *DolphinTplEo {
	key := cdaenum.DolphinTplKeyRelatedQuestions

	return &DolphinTplEo{
		Key:   key,
		Name:  key.GetName(),
		Value: d.ToString(),
	}
}
