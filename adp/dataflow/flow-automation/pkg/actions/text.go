package actions

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

// StringMatchType 匹配类型
type StringMatchType string

const (
	textSplit string = "@internal/text/split"
	textJoin  string = "@internal/text/join"
	textMatch string = "@internal/text/match"

	// NumberType 数字类型
	NumberType StringMatchType = "NUMBER"
	// EmailType 邮件类型
	EmailType StringMatchType = "EMAIL"
	// CardType 身份证类型
	CardType StringMatchType = "CN_ID_CARD"
	// PhoneType 手机号类型
	PhoneType StringMatchType = "CN_PHONE_NUMBER"
	// BankCardType 银行卡类型
	BankCardType StringMatchType = "CN_BANK_CARD_NUMBER"
)

// TextSplit text 分割
type TextSplit struct {
}

// TextSplitParam text 分割 参数
type TextSplitParam struct {
	Text      string `json:"text"`
	Custom    string `json:"custom"`
	Separator string `json:"separator"`
}

// Name 操作名称
func (a *TextSplit) Name() string {
	return textSplit
}

// Run 操作方法
func (a *TextSplit) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	input := params.(*TextSplitParam)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	var sep = input.Separator
	if sep == "custom" {
		sep = input.Custom
	}
	slices := strings.Split(input.Text, sep)
	sMap := make(map[int]string, 0)
	for index, s := range slices {
		sMap[index] = s
	}
	slicesByte, _ := json.Marshal(sMap)
	data := map[string]string{
		"slices": string(slicesByte),
	}
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *TextSplit) ParameterNew() interface{} {
	return &TextSplitParam{}
}

// TextJoin text 合并
type TextJoin struct {
}

// TextJoinParam text 合并 参数
type TextJoinParam struct {
	Texts     []string `json:"texts"`
	Custom    string   `json:"custom"`
	Separator string   `json:"separator"`
}

// Name 操作名称
func (a *TextJoin) Name() string {
	return textJoin
}

// Run 操作方法
func (a *TextJoin) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	input := params.(*TextJoinParam)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	var sep = input.Separator
	if sep == "custom" {
		sep = input.Custom
	}
	data := map[string]string{
		"text": strings.Join(input.Texts, sep),
	}
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *TextJoin) ParameterNew() interface{} {
	return &TextJoinParam{}
}

// TextMatch text 匹配
type TextMatch struct {
}

// TextMatchParam text 合并 参数
type TextMatchParam struct {
	Text      string `json:"text"`
	MatchType string `json:"matchtype"`
}

// Name 操作名称
func (a *TextMatch) Name() string {
	return textMatch
}

// Run 操作方法
func (a *TextMatch) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	input := params.(*TextMatchParam)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	var text = input.Text
	var matchType = StringMatchType(input.MatchType)
	var matchedText = ""

	tikaClient := drivenadapters.NewTika()
	var isFastTextAnalisysOK = true
	err = tikaClient.CheckFastTextAnalysys(ctx.Context())

	if err != nil {
		// fate-text-analysis服务不可用时，自行进行验证
		isFastTextAnalisysOK = false
	}

	if isFastTextAnalisysOK && matchType != NumberType {
		var methodMap = map[string]string{
			"KEYWORD": "KWD",
			"REG":     "REG",
		}
		var tpl = make(map[string]interface{})

		tpl["name"] = matchType
		tpl["expression"] = ""
		tpl["auxiliary_words"] = ""
		method, ok := methodMap[input.MatchType]
		if ok {
			tpl["method"] = method
		} else {
			tpl["method"] = "KWD"
		}

		var textByte = []byte(text)
		matchedRes, err := tikaClient.MatchContent(ctx.Context(), &textByte, tpl)

		if err != nil {
			return nil, err
		}

		if matchedRes.HasPrivateInfo {
			if matchedRes.Results[input.MatchType].Hit {
				matchedText = matchedRes.Results[input.MatchType].Info[0].Content
			}
		}
	} else {
		switch matchType {
		case NumberType:
			reg := regexp.MustCompile(`\d+`)
			matchedText = reg.FindString(text)
		case EmailType:
			reg := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
			matchedText = reg.FindString(text)
		case CardType:
			reg := regexp.MustCompile(`\d{6}(?:18|19|20)\d{2}(?:0[1-9]|1[0-2])(?:0[1-9]|[1-2][0-9]|3[0-1])\d{3}[0-9Xx]`)
			matchedText = reg.FindString(text)
		case PhoneType:
			reg := regexp.MustCompile(`1[3-9]\d{9}`)
			matchedText = reg.FindString(text)
		case BankCardType:
			reg := regexp.MustCompile(`\d{16,19}`)
			matchedText = reg.FindString(text)
		}
	}

	data := map[string]string{
		"matched": matchedText,
	}
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *TextMatch) ParameterNew() interface{} {
	return &TextMatchParam{}
}
