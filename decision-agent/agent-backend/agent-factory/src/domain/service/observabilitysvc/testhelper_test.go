package observabilitysvc

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iuniqueryhttp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iuniqueryhttp/uniquerymock"
	"go.uber.org/mock/gomock"
)

// 类型别名，让其他测试文件能引用
type (
	uniquerymock_iface = iuniqueryhttp.IUniquery //nolint:unused
	cmpmock_iface      = icmp.Logger             //nolint:unused
)

// uniquerymock_new creates a new MockIUniquery
func uniquerymock_new(ctrl *gomock.Controller) *uniquerymock.MockIUniquery {
	return uniquerymock.NewMockIUniquery(ctrl)
}

// cmpmock_new creates a new MockLogger
func cmpmock_new(ctrl *gomock.Controller) *cmpmock.MockLogger {
	return cmpmock.NewMockLogger(ctrl)
}

// newTestSvc creates a test observabilitySvc with mock Uniquery and Logger (no SquareSvc)
func newTestSvc(ctrl *gomock.Controller) (*observabilitySvc, *uniquerymock.MockIUniquery, *cmpmock.MockLogger) {
	mockUniquery := uniquerymock_new(ctrl)
	mockLogger := cmpmock_new(ctrl)
	svc := &observabilitySvc{
		logger:   mockLogger,
		uniquery: mockUniquery,
	}

	return svc, mockUniquery, mockLogger
}

// makeEntry builds a nested map simulating a UniQuery GetDataView entry.
// Numeric values are float64 (JSON default); start_time/end_time are also float64
// since the service does int64(attributes[...]["Data"].(float64)).
func makeEntry(sessionID, runID, conversationID, agentID, status string,
	totalTime, ttft, toolCallCount, toolCallFailedCount float64,
	startTime, endTime float64,
) map[string]any {
	return map[string]any{
		"Attributes": map[string]any{
			"session_id":             map[string]any{"Data": sessionID},
			"run_id":                 map[string]any{"Data": runID},
			"conversation_id":        map[string]any{"Data": conversationID},
			"agent_id":               map[string]any{"Data": agentID},
			"status":                 map[string]any{"Data": status},
			"total_time":             map[string]any{"Data": totalTime},
			"ttft":                   map[string]any{"Data": ttft},
			"tool_call_count":        map[string]any{"Data": toolCallCount},
			"tool_call_failed_count": map[string]any{"Data": toolCallFailedCount},
			"start_time":             map[string]any{"Data": startTime},
			"end_time":               map[string]any{"Data": endTime},
			"input_message":          map[string]any{"Data": "test input"},
			"agent_version":          map[string]any{"Data": "v1"},
		},
		"Resource": map[string]any{
			"service": map[string]any{
				"name": "agent-app",
			},
		},
	}
}

// makeRunEntry builds an entry specifically shaped for RunList (with run-level fields).
func makeRunEntry(agentID, runID, sessionID, status string, totalTime float64) map[string]any {
	return map[string]any{
		"Attributes": map[string]any{
			"agent_id":      map[string]any{"Data": agentID},
			"run_id":        map[string]any{"Data": runID},
			"session_id":    map[string]any{"Data": sessionID},
			"status":        map[string]any{"Data": status},
			"total_time":    map[string]any{"Data": totalTime},
			"input_message": map[string]any{"Data": "test input"},
		},
	}
}
