package actions

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
)

// AnyshareDocLibQuotaScale 配额扩容参数体
type AnyshareDocLibQuotaScale struct {
	User      interface{} `json:"user"`
	ScaleSize interface{} `json:"scale_size"`
}

// Name 操作名称
func (a *AnyshareDocLibQuotaScale) Name() string {
	return common.AnyshareDocLibQuotaScaleOpt
}

// Run 操作方法
func (a *AnyshareDocLibQuotaScale) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	tLog := traceLog.WithContext(ctx.Context())
	input := params.(*AnyshareDocLibQuotaScale)

	var user Accessor
	switch parsedUser := input.User.(type) {
	case string:
		err = json.Unmarshal([]byte(parsedUser), &user)
		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}
	case map[string]interface{}:
		user = Accessor{
			Type: parsedUser["type"].(string),
			ID:   parsedUser["id"].(string),
			Name: parsedUser["name"].(string),
		}
	}

	var scaleSize int64
	switch parsedScaleSize := input.ScaleSize.(type) {
	case string:
		scaleSize, err = strconv.ParseInt(parsedScaleSize, 10, 64)
		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}
	case int:
		scaleSize = int64(parsedScaleSize)
	case float64:
		scaleSize = int64(parsedScaleSize)
	case int64:
		scaleSize = parsedScaleSize
	default:
		return nil, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{
			"val":         input.ScaleSize,
			"type":        fmt.Sprintf("%v", reflect.TypeOf(input.ScaleSize)),
			"supportType": "string/int/int64/float64",
		})
	}

	if scaleSize < 0 {
		return nil, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{
			"val":  input.ScaleSize,
			"size": "scale size must be greater than zero",
		})
	}

	userID := ctx.GetTaskInstance().RelatedDagInstance.UserID
	isAdmin, err := rds.GetContentAdminDao().CheckAdminExistByUSerID(ctx.Context(), userID)
	if err != nil {
		tLog.Warnf("[AnyshareDocLibQuotaScale.Run] CheckAdminExistByUSerID failed, detail: %s", err.Error())
		return nil, err
	}

	if !isAdmin {
		return nil, errors.NewIError(errors.Forbidden, errors.NonProcessManager, map[string]interface{}{
			"userid": userID,
		})
	}

	docLibInfo, err := ctx.NewASDoc().GetUserDocLib(ctx.Context(), user.ID, token.Token)
	if err != nil {
		tLog.Warnf("[AnyshareDocLibQuotaScale.Run] GetUserDocLib failed, detail: %s", err.Error())
		return nil, err
	}

	// 扩容后的总容量
	totalQuato := docLibInfo.Quota.Allocated + scaleSize*1024*1024*1024
	err = ctx.NewASDoc().SetUserDocLibQuota(ctx.Context(), docLibInfo.ID, token.Token, totalQuato)
	if err != nil {
		tLog.Warnf("[AnyshareDocLibQuotaScale.Run] SetUserDocLibQuota failed, detail: %s", err.Error())
		return nil, err
	}

	ctx.Trace(ctx.Context(), "run end")
	return nil, nil
}

// ParameterNew 初始化参数
func (a *AnyshareDocLibQuotaScale) ParameterNew() interface{} {
	return &AnyshareDocLibQuotaScale{}
}
