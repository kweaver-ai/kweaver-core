package conversationsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// MarkRead implements iportdriver.IConversation.
func (svc *conversationSvc) MarkRead(ctx context.Context, id string, lastestReadIdx int) (err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetAttributes(ctx, attribute.String("conversation_id", id))
	oteltrace.SetAttributes(ctx, attribute.Int("lastest_read_idx", lastestReadIdx))

	_, err = svc.conversationRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			otellog.LogError(ctx, fmt.Sprintf("[MarkRead] get conversation error, id: %s, err: %v", id, err), err)
			err = capierr.NewCustom404Err(ctx, apierr.ConversationNotFound, fmt.Sprintf("[MarkRead] get conversation error, id: %s, err: %v", id, err))

			return
		}

		otellog.LogError(ctx, fmt.Sprintf("[MarkRead] get conversation error, id: %s, err: %v", id, err), err)

		return
	}

	err = svc.conversationRepo.Update(ctx, &dapo.ConversationPO{ID: id, ReadMessageIndex: lastestReadIdx})
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[MarkRead] update conversation error, id: %s, err: %v", id, err), err)
		return errors.Wrapf(err, "[MarkRead] update conversation error, id: %s, err: %v", id, err)
	}

	return
}
