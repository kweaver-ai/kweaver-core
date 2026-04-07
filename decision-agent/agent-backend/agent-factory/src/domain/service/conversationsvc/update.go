package conversationsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

func (svc *conversationSvc) Update(ctx context.Context, req conversationreq.UpdateReq) (err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetAttributes(ctx, attribute.String("conversation_id", req.ID))

	_, err = svc.conversationRepo.GetByID(ctx, req.ID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			otellog.LogError(ctx, fmt.Sprintf("[Update] get conversation error, id: %s, err: %v", req.ID, err), err)
			err = capierr.NewCustom404Err(ctx, apierr.ConversationNotFound, fmt.Sprintf("[Update] get conversation error, id: %s, err: %v", req.ID, err))

			return
		}

		return
	}

	currentTimestamp := cutil.GetCurrentMSTimestamp()
	if req.Title != "" {
		runes := []rune(req.Title)
		if len(runes) < 50 {
			req.Title = string(runes)
		} else {
			req.Title = string(runes[:50])
		}
	}

	err = svc.conversationRepo.Update(ctx, &dapo.ConversationPO{ID: req.ID, Title: req.Title, UpdateTime: currentTimestamp})
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[Update] update conversation error, id: %s, err: %v", req.ID, err), err)
		return errors.Wrapf(err, "[Update] update conversation error, id: %s, err: %v", req.ID, err)
	}

	return
}
