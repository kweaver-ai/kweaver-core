package conversationsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/conversationp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// ListByAgentID implements iportdriver.IConversationSvc.
func (svc *conversationSvc) ListByAgentID(ctx context.Context, agentID, title string, page, size int, startTime, endTime int64) (conversationList []conversationresp.ConversationDetail, count int64, err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetAttributes(ctx, attribute.String("agentID", agentID))

	rt, count, err := svc.conversationRepo.ListByAgentID(ctx, agentID, title, page, size)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[ListByAgentID] get conversation list by agentID error, agentID: %s, err: %v", agentID, err), err)
		return nil, 0, errors.Wrapf(err, "[ListByAgentID] get conversation list by agentID error, agentID: %s, err: %v", agentID, err)
	}

	eos, err := conversationp2e.Conversations(ctx, rt, svc.conversationMsgRepo)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[ListByAgentID] convert PO to EO error, agentID: %s, err: %v", agentID, err), err)
		return nil, 0, errors.Wrapf(err, "[ListByAgentID] convert PO to EO error, agentID: %s, err: %v", agentID, err)
	}

	conversationList = make([]conversationresp.ConversationDetail, len(eos))

	for i, eo := range eos {
		conversationDetail := conversationresp.NewConversationDetail()

		err := conversationDetail.LoadFromEo(eo)
		if err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[ListByAgentID] convert EO to DTO error, agentID: %s, err: %v", agentID, err), err)
			return nil, 0, errors.Wrapf(err, "[ListByAgentID] convert EO to DTO error, agentID: %s, err: %v", agentID, err)
		}

		conversationList[i] = *conversationDetail
	}

	return
}
