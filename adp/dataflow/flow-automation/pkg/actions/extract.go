package actions

import (
	"encoding/json"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/dependency"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"gorm.io/gorm"
)

const (
	customExtract        = "@docinfo/entity/extract"
	extractFileSizeLimit = 100 * 1024 * 1024
)

// CustomExtract 模型提取节点
type CustomExtract struct {
	ModelID string `json:"modelid"`
	Content string `json:"content"`
}

// Name 操作名称
func (a *CustomExtract) Name() string {
	return customExtract
}

// Run 操作方法
func (a *CustomExtract) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) { //nolint
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	tLog := traceLog.WithContext(ctx.Context())

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*CustomExtract)

	// 判断模型是否发布
	modelInfo, err := ctx.NewRepo().GetModelInfoByID(ctx.Context(), &rds.QueryCondition{ID: &input.ModelID})
	if err != nil {
		tLog.Warnf("[Extract.Run] GetModelInfoByID failed, detail: %s", err.Error())
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewIError(errors.TaskNotFound, "", map[string]interface{}{
				"id":     input.ModelID,
				"detail": err.Error(),
			})
		}
		return nil, err
	}

	repo := ctx.NewRepo()
	result := map[string]interface{}{}
	switch modelInfo.Type {
	case common.UIEType:
		if modelInfo.Status != common.Publish {
			err = errors.NewIError(errors.Forbidden, "", map[string]interface{}{
				"status": "model is not publish",
			})
			return result, err
		}
		entitys, err := repo.StartInfer(ctx.Context(), input.ModelID, input.Content, token.Token, token.LoginIP, extractFileSizeLimit)
		if err != nil {
			tLog.Warnf("[Extract.Run] StartInfer failed, detail: %s", err.Error())
			return result, err
		}

		if _res, ok := entitys["res"].([]interface{}); ok {
			if len(_res) > 0 {
				resBytes, _ := json.Marshal(_res[0])
				result["result"] = string(resBytes)
			}
		}
	case common.TagRuleType:
		target := map[string]string{}
		if strings.HasPrefix(input.Content, "gns://") {
			target["docid"] = input.Content
		} else {
			target["content"] = input.Content
		}
		entitys, err := repo.ExtractTagsByRule(ctx.Context(), &dependency.TagExtractionParams{
			Target: target,
			RuleID: input.ModelID,
		}, &drivenadapters.UserInfo{
			LoginIP: token.LoginIP,
			TokenID: token.Token,
		}, extractFileSizeLimit)
		if err != nil {
			tLog.Warnf("[Extract.Run] ExtractTagsByRule failed, detail: %s", err.Error())
			return result, err
		}
		resBytes, _ := json.Marshal(entitys)
		result["result"] = string(resBytes)
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, result)
	ctx.Trace(ctx.Context(), "run end")

	return result, nil
}

// ParameterNew 初始化参数
func (a *CustomExtract) ParameterNew() interface{} {
	return &CustomExtract{}
}
