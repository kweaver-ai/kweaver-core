package conversationp2e

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/conversationeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation_message/conversationmsgreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/pkg/errors"
)

func Conversation(ctx context.Context, _po *dapo.ConversationPO, conversationMsgRepo idbaccess.IConversationMsgRepo, withMsg bool) (eo *conversationeo.Conversation, err error) {
	eo = &conversationeo.Conversation{
		ConversationPO: _po,
	}

	if withMsg {
		msgPOList, err := conversationMsgRepo.List(ctx, conversationmsgreq.ListReq{ConversationID: _po.ID})
		if err != nil {
			return nil, errors.Wrapf(err, "查询对话消息失败")
		}

		eo.Messages = msgPOList
	}

	return
}

func ConversationWithLimit(ctx context.Context, _po *dapo.ConversationPO, conversationMsgRepo idbaccess.IConversationMsgRepo, limit int) (eo *conversationeo.Conversation, err error) {
	eo = &conversationeo.Conversation{
		ConversationPO: _po,
	}

	msgPOList, err := conversationMsgRepo.GetRecentMessages(ctx, _po.ID, limit)
	if err != nil {
		return nil, errors.Wrapf(err, "查询对话消息失败")
	}

	eo.Messages = msgPOList

	return
}

// DataAgents 批量PO转EO
func Conversations(ctx context.Context, _pos []*dapo.ConversationPO, conversationMsgRepo idbaccess.IConversationMsgRepo) (eos []*conversationeo.Conversation, err error) {
	eos = make([]*conversationeo.Conversation, 0, len(_pos))

	for i := range _pos {
		var eo *conversationeo.Conversation

		if eo, err = Conversation(ctx, _pos[i], conversationMsgRepo, false); err != nil {
			return
		}

		eos = append(eos, eo)
	}

	return
}
