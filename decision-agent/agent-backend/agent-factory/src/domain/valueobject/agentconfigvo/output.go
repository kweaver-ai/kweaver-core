package agentconfigvo

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
)

type Variable struct {
	AnswerVar           string   `json:"answer_var"`
	DocRetrievalVar     string   `json:"doc_retrieval_var"`
	GraphRetrievalVar   string   `json:"graph_retrieval_var"`
	RelatedQuestionsVar string   `json:"related_questions_var"`
	OtherVars           []string `json:"other_vars"`

	MiddleOutputVars []string `json:"middle_output_vars"`
}

type OutputVariablesS Variable

func NewOutputVariablesS() *OutputVariablesS {
	return &OutputVariablesS{}
}

func (v *OutputVariablesS) LoadFromConfig(config *daconfvalobj.Config) (err error) {
	// 1. 拿到output配置
	err = cutil.CopyStructUseJSON(v, config.Output.Variables)
	if err != nil {
		return
	}

	// 2. 拿到dolphin中的中间需要在页面上展示的变量
	if len(v.MiddleOutputVars) > 0 {
		return
	}

	dolphin := config.Dolphin
	isDolphinMode := config.IsDolphinMode

	if dolphin != "" && isDolphinMode == 1 {
		var middleOutputVars []string

		middleOutputVars, err = ExtractOutputsFromText(dolphin)
		if err != nil {
			return
		}

		v.MiddleOutputVars = middleOutputVars
	}

	return
}

func (v *OutputVariablesS) ToVariable() (variable *Variable, err error) {
	variable = &Variable{}
	err = cutil.CopyStructUseJSON(variable, v)

	return
}

// -----------提取output_xxx  start-----------

// extractOutputFromLine 从单行文本中提取output_xx
func extractOutputFromLine(line string) string {
	// 创建正则表达式匹配包含->或>>且包含output_的行
	re := regexp.MustCompile(`(?:->|>>)\s*(output_\w+)`)

	// 查找匹配
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		// matches[0]是完整匹配，matches[1]是第一个捕获组(output_xx)
		return matches[1]
	}

	return ""
}

// extractOutputs 从任意Reader中提取所有匹配的output_xx
func extractOutputs(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	// 初始化为空切片而不是nil
	outputs := []string{}

	// 逐行读取
	for scanner.Scan() {
		line := scanner.Text()

		// 使用匹配函数处理每一行
		if output := extractOutputFromLine(line); output != "" {
			outputs = append(outputs, output)
		}
	}

	if err := scanner.Err(); err != nil {
		return []string{}, err
	}

	return outputs, nil
}

// ExtractOutputsFromText 从文本字符串中提取所有匹配的output_xx
func ExtractOutputsFromText(text string) ([]string, error) {
	return extractOutputs(strings.NewReader(text))
}

// -----------提取output_xxx  end-----------
