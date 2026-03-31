package conversationsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/conversationp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// ListByAgentID implements iportdriver.IConversationSvc.
func (svc *conversationSvc) ListByAgentID(ctx context.Context, agentID, title string, page, size int, startTime, endTime int64) (conversationList []conversationresp.ConversationDetail, count int64, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	o11y.SetAttributes(ctx, attribute.String("agentID", agentID))

	// 1. 获取数据
	rt, count, err := svc.conversationRepo.ListByAgentID(ctx, agentID, title, page, size)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[ListByAgentID] get conversation list by agentID error, agentID: %s, err: %v", agentID, err))
		return nil, 0, errors.Wrapf(err, "[ListByAgentID] get conversation list by agentID error, agentID: %s, err: %v", agentID, err)
	}

	// 2. PO转EO
	eos, err := conversationp2e.Conversations(ctx, rt, svc.conversationMsgRepo)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[ListByAgentID] convert PO to EO error, agentID: %s, err: %v", agentID, err))
		return nil, 0, errors.Wrapf(err, "[ListByAgentID] convert PO to EO error, agentID: %s, err: %v", agentID, err)
	}

	// 3. 转换为响应DTO
	conversationList = make([]conversationresp.ConversationDetail, len(eos))

	for i, eo := range eos {
		conversationDetail := conversationresp.NewConversationDetail()

		err := conversationDetail.LoadFromEo(eo)
		if err != nil {
			o11y.Error(ctx, fmt.Sprintf("[ListByAgentID] convert EO to DTO error, agentID: %s, err: %v", agentID, err))
			return nil, 0, errors.Wrapf(err, "[ListByAgentID] convert EO to DTO error, agentID: %s, err: %v", agentID, err)
		}

		conversationList[i] = *conversationDetail
	}

	// TODO:  对于每个conversation，需要查询一下
	return
}
