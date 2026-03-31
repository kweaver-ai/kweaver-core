package conversationsvc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/conversationp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// List implements iportdriver.IConversation.
func (svc *conversationSvc) List(ctx context.Context, req conversationreq.ListReq) (conversationList conversationresp.ListConversationResp, count int64, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	o11y.SetAttributes(ctx, attribute.String("app_key", req.AgentAPPKey))

	conversationListEmpty := conversationresp.ListConversationResp{}
	// 1. 获取数据
	rt, count, err := svc.conversationRepo.List(ctx, req)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[List] get conversation list error, app_key: %s, err: %v", req.AgentAPPKey, err))
		return conversationListEmpty, 0, errors.Wrapf(err, "[List] get conversation list error, app_key: %s, err: %v", req.AgentAPPKey, err)
	}

	// 2. PO转EO
	eos, err := conversationp2e.Conversations(ctx, rt, svc.conversationMsgRepo)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[List] convert PO to EO error, app_key: %s, err: %v", req.AgentAPPKey, err))
		return conversationListEmpty, 0, errors.Wrapf(err, "[List] convert PO to EO error, app_key: %s, err: %v", req.AgentAPPKey, err)
	}
	// 3. 转换为响应DTO

	conversationList = make([]conversationresp.ConversationDetail, len(eos))

	for i, eo := range eos {
		conversationDetail := conversationresp.NewConversationDetail()

		err := conversationDetail.LoadFromEo(eo)
		if err != nil {
			o11y.Error(ctx, fmt.Sprintf("[List] convert EO to DTO error, app_key: %s, err: %v", req.AgentAPPKey, err))
			return conversationListEmpty, 0, errors.Wrapf(err, "[List] convert EO to DTO error, app_key: %s, err: %v", req.AgentAPPKey, err)
		}

		conversationList[i] = *conversationDetail
	}

	// NOTE: 获取会话最新消息的状态
	for index, conversation := range conversationList {
		po, err := svc.conversationMsgRepo.GetLatestMsgByConversationID(ctx, conversation.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				conversationList[index].Status = cdaenum.ConvStatusCompleted
			} else {
				return conversationListEmpty, 0, errors.Wrapf(err, "获取会话最新消息失败")
			}
		} else {
			if po.Status == cdaenum.MsgStatusProcessing {
				conversationList[index].Status = cdaenum.ConvStatusProcessing
			} else if po.Status == cdaenum.MsgStatusSucceded {
				conversationList[index].Status = cdaenum.ConvStatusCompleted
			} else if po.Status == cdaenum.MsgStatusCancelled {
				conversationList[index].Status = cdaenum.ConvStatusCancelled
			} else {
				conversationList[index].Status = cdaenum.ConvStatusFailed
			}
		}
	}

	return
}
