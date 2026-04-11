package conversationsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/p2e/conversationp2e"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

func (svc *conversationSvc) Detail(ctx context.Context, id string) (res conversationresp.ConversationDetail, err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetAttributes(ctx, attribute.String("conversation_id", id))

	conversationDetailEmpty := *conversationresp.NewConversationDetail()

	po, err := svc.conversationRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			otellog.LogError(ctx, fmt.Sprintf("[Detail] conversation not found, id: %s", id), err)
			return conversationDetailEmpty, errors.Wrapf(err, "数据智能体配置不存在")
		}

		otellog.LogError(ctx, fmt.Sprintf("[Detail] get conversation by id error, id: %s, err: %v", id, err), err)

		return conversationDetailEmpty, errors.Wrapf(err, "获取数据失败")
	}

	eo, err := conversationp2e.Conversation(ctx, po, svc.conversationMsgRepo, true)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[Detail] conversation p2e error, id: %s, err: %v", id, err), err)
		return conversationDetailEmpty, errors.Wrapf(err, "PO转EO失败")
	}

	conversationDetail := conversationresp.NewConversationDetail()
	_ = conversationDetail.LoadFromEo(eo)
	res = *conversationDetail

	return
}

func (svc *conversationSvc) DetailWithLimit(ctx context.Context, id string, limit int) (res conversationresp.ConversationDetail, err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetAttributes(ctx, attribute.String("conversation_id", id), attribute.Int("limit", limit))

	conversationDetailEmpty := *conversationresp.NewConversationDetail()

	po, err := svc.conversationRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			otellog.LogError(ctx, fmt.Sprintf("[DetailWithLimit] conversation not found, id: %s", id), err)
			return conversationDetailEmpty, errors.Wrapf(err, "数据智能体配置不存在")
		}

		otellog.LogError(ctx, fmt.Sprintf("[DetailWithLimit] get conversation by id error, id: %s, err: %v", id, err), err)

		return conversationDetailEmpty, errors.Wrapf(err, "获取数据失败")
	}

	eo, err := conversationp2e.ConversationWithLimit(ctx, po, svc.conversationMsgRepo, limit)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[DetailWithLimit] conversation p2e error, id: %s, err: %v", id, err), err)
		return conversationDetailEmpty, errors.Wrapf(err, "PO转EO失败")
	}

	conversationDetail := conversationresp.NewConversationDetail()
	_ = conversationDetail.LoadFromEo(eo)
	res = *conversationDetail

	return
}
