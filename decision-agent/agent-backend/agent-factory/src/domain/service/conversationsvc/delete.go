package conversationsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// Delete implements iportdriver.IConversation.
func (svc *conversationSvc) Delete(ctx context.Context, id string) (err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetAttributes(ctx, attribute.String("conversation_id", id))

	_, err = svc.conversationRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			svc.logger.Errorf("[Delete] conversation not found, id: %s", id)
			otellog.LogError(ctx, fmt.Sprintf("[Delete] conversation not found, id: %s", id), err)
			err = capierr.NewCustom404Err(ctx, apierr.ConversationNotFound, "数据智能体配置不存在")

			return
		}

		otellog.LogError(ctx, fmt.Sprintf("[Delete] get conversation by id error, id: %s, err: %v", id, err), err)

		return
	}

	tx, err := svc.conversationRepo.BeginTx(ctx)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[Delete] begin tx error, id: %s, err: %v", id, err), err)
		return
	}

	defer chelper.TxRollbackOrCommit(tx, &err, svc.logger)

	err = svc.conversationRepo.Delete(ctx, tx, id)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[Delete] delete conversation error, id: %s, err: %v", id, err), err)
		err = errors.Wrapf(err, "删除对话数据失败")

		return
	}

	err = svc.conversationMsgRepo.DeleteByConversationID(ctx, tx, id)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[Delete] delete conversation msg error, id: %s, err: %v", id, err), err)
		err = errors.Wrapf(err, "删除对话消息数据失败")

		return
	}

	return
}

// DeleteByAppKey implements iportdriver.IConversation.
func (svc *conversationSvc) DeleteByAppKey(ctx context.Context, appKey string) (err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetAttributes(ctx, attribute.String("app_key", appKey))

	tx, err := svc.conversationRepo.BeginTx(ctx)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[DeleteByAppKey] begin tx error, appKey: %s, err: %v", appKey, err), err)
		return
	}

	defer chelper.TxRollbackOrCommit(tx, &err, svc.logger)

	err = svc.conversationRepo.DeleteByAPPKey(ctx, tx, appKey)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[DeleteByAppKey] delete conversation error, appKey: %s, err: %v", appKey, err), err)
		return errors.Wrapf(err, "删除对话数据失败")
	}

	err = svc.conversationMsgRepo.DeleteByAPPKey(ctx, tx, appKey)
	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[DeleteByAppKey] delete conversation msg error, appKey: %s, err: %v", appKey, err), err)
		return errors.Wrapf(err, "删除对话消息数据失败")
	}

	return
}
