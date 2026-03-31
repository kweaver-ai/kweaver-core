package conversationsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/conversationp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

func (svc *conversationSvc) Detail(ctx context.Context, id string) (res conversationresp.ConversationDetail, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	o11y.SetAttributes(ctx, attribute.String("conversation_id", id))

	conversationDetailEmpty := *conversationresp.NewConversationDetail()
	// 1. 获取数据
	po, err := svc.conversationRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			o11y.Error(ctx, fmt.Sprintf("[Detail] conversation not found, id: %s", id))
			return conversationDetailEmpty, errors.Wrapf(err, "数据智能体配置不存在")
		}

		o11y.Error(ctx, fmt.Sprintf("[Detail] get conversation by id error, id: %s, err: %v", id, err))

		return conversationDetailEmpty, errors.Wrapf(err, "获取数据失败")
	}

	// 2. PO转EO
	eo, err := conversationp2e.Conversation(ctx, po, svc.conversationMsgRepo, true)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[Detail] conversation p2e error, id: %s, err: %v", id, err))
		return conversationDetailEmpty, errors.Wrapf(err, "PO转EO失败")
	}

	// 3. 转换为响应DTO
	conversationDetail := conversationresp.NewConversationDetail()
	_ = conversationDetail.LoadFromEo(eo)
	res = *conversationDetail

	return
}

func (svc *conversationSvc) DetailWithLimit(ctx context.Context, id string, limit int) (res conversationresp.ConversationDetail, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	o11y.SetAttributes(ctx, attribute.String("conversation_id", id), attribute.Int("limit", limit))

	conversationDetailEmpty := *conversationresp.NewConversationDetail()
	// 1. 获取数据
	po, err := svc.conversationRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			o11y.Error(ctx, fmt.Sprintf("[DetailWithLimit] conversation not found, id: %s", id))
			return conversationDetailEmpty, errors.Wrapf(err, "数据智能体配置不存在")
		}

		o11y.Error(ctx, fmt.Sprintf("[DetailWithLimit] get conversation by id error, id: %s, err: %v", id, err))

		return conversationDetailEmpty, errors.Wrapf(err, "获取数据失败")
	}

	// 2. PO转EO（带limit）
	eo, err := conversationp2e.ConversationWithLimit(ctx, po, svc.conversationMsgRepo, limit)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[DetailWithLimit] conversation p2e error, id: %s, err: %v", id, err))
		return conversationDetailEmpty, errors.Wrapf(err, "PO转EO失败")
	}

	// 3. 转换为响应DTO
	conversationDetail := conversationresp.NewConversationDetail()
	_ = conversationDetail.LoadFromEo(eo)
	res = *conversationDetail

	return
}
