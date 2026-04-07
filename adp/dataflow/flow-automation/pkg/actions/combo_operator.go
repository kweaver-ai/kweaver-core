package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

type ComboOperatorRequest struct {
	Parameters []any `json:"parameters"`
	Body       any   `json:"body"`
}

type ComboOperator struct {
	Operator string                `json:"operator"`
	Version  string                `json:"version"`
	Request  *ComboOperatorRequest `json:"request"`
}

func (*ComboOperator) Name() string {
	return "@operator"
}

func (*ComboOperator) ParameterNew() interface{} {
	return &ComboOperator{}
}

func (c *ComboOperator) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {

	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	log := traceLog.WithContext(ctx.Context())

	input := params.(*ComboOperator)
	taskIns := ctx.GetTaskInstance()

	if !strings.HasPrefix(input.Operator, c.Name()) {
		err = fmt.Errorf("invalid operator %s", input.Operator)
		log.Warnf("[ComboOperator] err %s, taskId %s", err.Error(), taskIns.ID)
		return nil, err
	}

	operatorID := strings.TrimPrefix(input.Operator, "@operator/")

	userInfo := &drivenadapters.UserInfo{}
	if token != nil {
		userInfo.UserID = token.UserID
		userInfo.AccountType = common.User.ToString()
		if token.IsApp {
			userInfo.AccountType = common.APP.ToString()
		}
	}
	agentOperatorIntegration := drivenadapters.NewAgentOperatorIntegration()
	operator, err := agentOperatorIntegration.GetOperatorInfo(ctx.Context(), operatorID, input.Version, taskIns.RelatedDagInstance.BizDomainID, userInfo)

	if err != nil {
		log.Warnf("[CallOperator] GetOperatorInfo err %s, taskId %s", err.Error(), taskIns.ID)
		return nil, err
	}

	data, err := callOperator(
		ctx.Context(),
		taskIns,
		token,
		operator,
		input.Request.Parameters,
		input.Request.Body,
		ctx.IsDebug(),
	)

	if err != nil {
		return nil, err
	}

	result := map[string]any{"data": data}

	if operator.OperatorInfo.ExecutionMode == "async" {
		statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
		ctx.ShareData().Set(statusKey, entity.TaskInstanceStatusBlocked)
	} else {
		ctx.ShareData().Set(ctx.GetTaskID(), result)
	}

	return result, nil
}

func (c *ComboOperator) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	taskIns := ctx.GetTaskInstance()
	statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
	status, ok := ctx.ShareData().Get(statusKey)
	if ok && status == entity.TaskInstanceStatusBlocked {
		return entity.TaskInstanceStatusBlocked, nil
	}

	return entity.TaskInstanceStatusSuccess, nil
}

var _ entity.Action = (*ComboOperator)(nil)

func callOperator(ctx context.Context,
	taskIns *entity.TaskInstance,
	token *entity.Token,
	operator *drivenadapters.OperatorResponse,
	parameters []any,
	body any,
	isDebug bool) (data any, err error) {

	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	isAsync := operator.OperatorInfo.ExecutionMode == "async"

	authorization := drivenadapters.NewAuthorization()

	isApp, err := drivenadapters.NewUserManagement().IsApp(taskIns.RelatedDagInstance.UserID)
	if err != nil {
		log.Warnf("[Run.ComboOperator] IsApp err %s, taskId %s", err.Error(), taskIns.ID)
		return nil, err
	}

	hasPerm, err := authorization.OperationPermCheck(ctx, drivenadapters.OperationPermCheckParams{
		Accessor: drivenadapters.Vistor{
			ID:   taskIns.RelatedDagInstance.UserID,
			Type: utils.IfNot(isApp, "app", "user"),
		},
		Resource: drivenadapters.Resource{
			ID:   operator.OperatorID,
			Type: "operator",
		},
		Operation: []string{"execute"},
		Method:    "GET",
	})

	if err != nil {
		log.Warnf("[ComboOperator] GetOperatorInfo err %s, taskId %s", err.Error(), taskIns.ID)
		return nil, err
	}

	if !hasPerm {
		return nil, ierr.NewPublicRestError(ctx, ierr.PErrorForbidden, errors.NoPermission, map[string]any{"operator_id": []any{operator.OperatorID}})
	}

	reqUrl := fmt.Sprintf("%s/%s",
		strings.TrimSuffix(operator.Metadata.ServerURL, "/"),
		strings.TrimPrefix(operator.Metadata.Path, "/"))
	headers := make(http.Header)
	query := make(url.Values)
	cookies := make([]*http.Cookie, 0)

	for index, p := range operator.Metadata.APISpec.Parameters {
		if len(parameters) <= index {
			log.Warnf("[CallOperator] parameters.%d is missing, taskId %s", index, taskIns.ID)
			return nil, fmt.Errorf("parameters.%d is required", index)
		}

		value := parameters[index]

		if value == nil {
			value = ""
		}

		switch p.In {
		case "path":
			reqUrl = strings.Replace(reqUrl, "{"+p.Name+"}", fmt.Sprintf("%v", value), 1)
		case "header":
			headers.Add(p.Name, fmt.Sprintf("%v", value))
		case "query":
			query.Add(p.Name, fmt.Sprintf("%v", value))
		case "cookie":
			cookies = append(cookies, &http.Cookie{
				Name:  p.Name,
				Value: fmt.Sprintf("%v", value),
			})
		}
	}

	if operator.OperatorInfo.OperatorType == "composite" {
		headers.Set("X-Parent-Execution-ID", taskIns.RelatedDagInstance.ID)
	}

	u, err := url.Parse(reqUrl)

	if err != nil {
		log.Warnf("[CallOperator] url.Parse err %s, taskId %s", err.Error(), taskIns.ID)
		return nil, err
	}

	u.RawQuery = query.Encode()

	var bodyBytes []byte

	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	req, err := http.NewRequestWithContext(ctx, operator.Metadata.Method, u.String(), bytes.NewReader(bodyBytes))

	if err != nil {
		log.Warnf("[CallOperator] Request err %s, taskId %s", err.Error(), taskIns.ID)
		return nil, err
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	hasAuth := false

	for k, values := range headers {
		for _, v := range values {
			if strings.ToLower(k) == "authorization" {
				if v = strings.TrimSpace(v); v != "" {
					hasAuth = true
					req.Header.Add(k, v)
				}
				continue
			}
			req.Header.Add(k, v)
		}
	}

	if !hasAuth {
		req.Header.Add("Authorization", "Bearer "+token.Token)
	}

	client := drivenadapters.NewOtelRawHTTPClient()

	if isAsync {
		config := common.NewConfig()
		taskID := taskIns.ID
		if isDebug {
			taskID = fmt.Sprintf("DEBUG:%v", taskIns.ID)
		}

		callback := fmt.Sprintf("%s://%s:%s/api/automation/v1/continuations/%s", config.AccessAddress.Schema, config.AccessAddress.Host, config.AccessAddress.Port, taskID)
		req.Header.Add("X-Callback-URL", callback)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Warnf("[CallOperator] Request err %s, taskId %s", err.Error(), taskIns.ID)
		return nil, err
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			log.Warnf("[CallOperator] err %s", closeErr.Error())
		}
	}()

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Warnf("[CallOperator] Read body err %s, taskId %s", err.Error(), taskIns.ID)
		return nil, err
	}

	if resp.StatusCode >= 400 {
		log.Warnf("[CallOperator] Response status %d, resp %s, taskId %s", resp.StatusCode, string(respBody), taskIns.ID)
		return nil, errors.ExHTTPError{
			Body:   string(respBody),
			Status: resp.StatusCode,
		}
	}

	if isAsync {
		return map[string]any{"data": nil}, nil
	}

	_ = json.Unmarshal(respBody, &data)

	if dataMap, ok := data.(map[string]any); ok && len(dataMap) == 0 {
		data = nil
	}

	return data, nil
}
