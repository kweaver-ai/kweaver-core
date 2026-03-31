package observabilitysvc

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/uniqueryaccess/uniquerydto"
	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	observabilityresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/resp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func (s *observabilitySvc) RunList(ctx context.Context, req *observabilityreq.RunListReq) (*observabilityresp.RunListResp, error) {
	// 1. 构建过滤条件
	conditions := []uniquerydto.Condition{
		{
			Field:     "Resource.service.name",
			Operation: "==",
			Value:     "agent-app",
			ValueFrom: "const",
		},
		{
			Field:     "SeverityText",
			Operation: "==",
			Value:     "Info",
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.call_type.Data",
			Operation: "!=",
			Value:     "api_chat",
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.agent_version.Data",
			Operation: "!=",
			Value:     "v0",
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.session_id.Data",
			Operation: "==",
			Value:     req.SessionID,
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.start_time.Data",
			Operation: ">=",
			Value:     req.StartTime,
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.end_time.Data",
			Operation: "<",
			Value:     req.EndTime,
			ValueFrom: "const",
		},
	}

	// 如果agent_id不为空，添加agent_id过滤条件
	if req.AgentID != "" {
		conditions = append(conditions, uniquerydto.Condition{
			Field:     "Attributes.agent_id.Data",
			Operation: "==",
			Value:     req.AgentID,
			ValueFrom: "const",
		})
	}

	// 如果conversation_id不为空，添加conversation_id过滤条件
	if req.ConversationID != "" {
		conditions = append(conditions, uniquerydto.Condition{
			Field:     "Attributes.conversation_id.Data",
			Operation: "==",
			Value:     req.ConversationID,
			ValueFrom: "const",
		})
	}

	// 2. 构建uniquery请求
	uniqueryReq := uniquerydto.ReqDataView{
		Limit:        req.Size,
		Offset:       (req.Page - 1) * req.Size,
		Start:        req.StartTime,
		End:          req.EndTime,
		Format:       "original",
		XAccountID:   req.XAccountID,
		XAccountType: string(req.XAccountType),
		Filters: &uniquerydto.GlobalFilter{
			Operation:     "and",
			SubConditions: conditions,
		},
	}

	// 3. 调用uniquery GetDataView接口
	viewResults, err := s.uniquery.GetDataView(ctx, "__dip_o11y_log", uniqueryReq)
	// viewResults, err := s.uniquery.GetDataViewMock(ctx, "__dip_o11y_log", uniqueryReq, "run_list")
	if err != nil {
		s.logger.Errorf("[RunList] GetDataView failed: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
	}

	entries := viewResults.Entries

	// 4. 转换为RunListItem列表
	var runListItems []observabilityresp.RunListItem

	for _, entry := range entries {
		entryMap := entry.(map[string]any)
		attributes := entryMap["Attributes"].(map[string]any)

		runListItem := observabilityresp.RunListItem{
			AgentID:      attributes["agent_id"].(map[string]any)["Data"].(string),
			RunID:        attributes["run_id"].(map[string]any)["Data"].(string),
			InputMessage: attributes["input_message"].(map[string]any)["Data"].(string),
			Status:       attributes["status"].(map[string]any)["Data"].(string),
			TotalTime:    int(attributes["total_time"].(map[string]any)["Data"].(float64)),
		}

		runListItems = append(runListItems, runListItem)
	}

	// 5. 返回响应
	return &observabilityresp.RunListResp{
		Entries:    runListItems,
		TotalCount: len(runListItems),
	}, nil
}
