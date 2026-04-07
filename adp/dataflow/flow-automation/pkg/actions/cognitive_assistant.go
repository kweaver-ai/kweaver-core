package actions

import (
	"fmt"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

var (
	docSummarizeID  = "1012"
	meetSummarizeID = "1011"
)

// DocSummarize 基于大模型总结文档
type DocSummarize struct {
	DocID string `json:"docid"`
}

// Name 方法名称
func (*DocSummarize) Name() string {
	return common.OpCognitiveAssistantDocSummarize
}

// ParameterNew 方法参数
func (*DocSummarize) ParameterNew() interface{} {
	return &DocSummarize{}
}

// Run 方法执行逻辑
func (*DocSummarize) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) { // nolint
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	log := traceLog.WithContext(ctx.Context())
	input := params.(*DocSummarize)

	var result = map[string]interface{}{
		"result": "",
	}

	content, err := ctx.NewASDoc().GetDocSetSubdocContent(ctx.Context(), drivenadapters.DocSetSubdocParams{
		DocID: input.DocID,
		Type:  "full_text",
	}, -1, time.Second, 10*1024*1024)

	if err != nil {
		log.Warnf("[CustomPrompt] GetDocSetSubdocContent failed, err: %v", err)
		return nil, err
	}

	if len(content) > 0 {
		cognitiveAssistant := drivenadapters.NewCognitiveAssistant()

		result["result"], err = cognitiveAssistant.CustomPrompt(ctx.Context(), docSummarizeID, content)
	}

	ctx.Trace(ctx.Context(), "run end")

	if err != nil {
		log.Warnf("[CustomPrompt] CustomPrompt failed, err: %v", err)
		return nil, err
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, result)

	return result, nil
}

// MeetSummarize 基于大模型做会议纪要总结
type MeetSummarize struct {
	PromptServiceID string `json:"prompt_service_id"`
	DocID           string `json:"docid"`
}

// Name 方法名称
func (*MeetSummarize) Name() string {
	return common.OpCognitiveAssistantMeetSummarize
}

// ParameterNew 方法参数
func (*MeetSummarize) ParameterNew() interface{} {
	return &MeetSummarize{}
}

// Run 方法执行逻辑
func (*MeetSummarize) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) { // nolint
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	log := traceLog.WithContext(ctx.Context())
	input := params.(*MeetSummarize)

	var result = map[string]interface{}{
		"result": "",
	}

	content, err := ctx.NewASDoc().GetDocSetSubdocContent(ctx.Context(), drivenadapters.DocSetSubdocParams{
		DocID: input.DocID,
		Type:  "full_text",
	}, -1, time.Second, 10*1024*1024)

	if err != nil {
		log.Warnf("[CustomPrompt] GetDocSetSubdocContent failed, err: %v", err)
		return nil, err
	}

	if len(content) > 0 {
		cognitiveAssistant := drivenadapters.NewCognitiveAssistant()

		result["result"], err = cognitiveAssistant.CustomPrompt(ctx.Context(), meetSummarizeID, content)
	}

	ctx.Trace(ctx.Context(), "run end")

	if err != nil {
		log.Warnf("[CustomPrompt] CustomPrompt failed, err: %v", err)
		return nil, err
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, result)

	return result, nil
}

func getCustomPromptID(result interface{}, key string) string {
	if promptList, ok := result.([]interface{}); ok {
		for _, prompt := range promptList {
			if promptMap, ok := prompt.(map[string]interface{}); ok {
				if promptMap["class_name"] != "WorkCenter" {
					continue
				}
				if _, ok := promptMap["prompt"].([]interface{}); !ok {
					continue
				}

				for _, prompt := range promptMap["prompt"].([]interface{}) {
					if promptDetailMap, ok := prompt.(map[string]interface{}); ok {
						if promptDetailMap["prompt_name"] == key {
							return fmt.Sprintf("%v", promptDetailMap["prompt_service_id"])
						}
					}
				}
			}
		}
	}

	return ""
}
