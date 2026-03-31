package conversationsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// MarkRead implements iportdriver.IConversation.
func (svc *conversationSvc) MarkRead(ctx context.Context, id string, lastestReadIdx int) (err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	o11y.SetAttributes(ctx, attribute.String("conversation_id", id))
	o11y.SetAttributes(ctx, attribute.Int("lastest_read_idx", lastestReadIdx))

	_, err = svc.conversationRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			o11y.Error(ctx, fmt.Sprintf("[MarkRead] get conversation error, id: %s, err: %v", id, err))
			err = capierr.NewCustom404Err(ctx, apierr.ConversationNotFound, fmt.Sprintf("[MarkRead] get conversation error, id: %s, err: %v", id, err))

			return
		}

		o11y.Error(ctx, fmt.Sprintf("[MarkRead] get conversation error, id: %s, err: %v", id, err))

		return
	}

	// 更新最新读取索引
	err = svc.conversationRepo.Update(ctx, &dapo.ConversationPO{ID: id, ReadMessageIndex: lastestReadIdx})
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[MarkRead] update conversation error, id: %s, err: %v", id, err))
		return errors.Wrapf(err, "[MarkRead] update conversation error, id: %s, err: %v", id, err)
	}

	return
}
