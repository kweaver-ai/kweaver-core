package conversationsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

func (svc *conversationSvc) Update(ctx context.Context, req conversationreq.UpdateReq) (err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	o11y.SetAttributes(ctx, attribute.String("conversation_id", req.ID))

	// 1. 获取数据
	_, err = svc.conversationRepo.GetByID(ctx, req.ID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			o11y.Error(ctx, fmt.Sprintf("[Update] get conversation error, id: %s, err: %v", req.ID, err))
			err = capierr.NewCustom404Err(ctx, apierr.ConversationNotFound, fmt.Sprintf("[Update] get conversation error, id: %s, err: %v", req.ID, err))

			return
		}

		return
	}

	currentTimestamp := cutil.GetCurrentMSTimestamp()
	// 截断title
	if req.Title != "" {
		runes := []rune(req.Title)
		if len(runes) < 50 {
			req.Title = string(runes)
		} else {
			req.Title = string(runes[:50])
		}
	}

	// 2. 更新标题
	err = svc.conversationRepo.Update(ctx, &dapo.ConversationPO{ID: req.ID, Title: req.Title, UpdateTime: currentTimestamp})
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[Update] update conversation error, id: %s, err: %v", req.ID, err))
		return errors.Wrapf(err, "[Update] update conversation error, id: %s, err: %v", req.ID, err)
	}

	return
}
