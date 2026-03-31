package conversationsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// Delete implements iportdriver.IConversation.
func (svc *conversationSvc) Delete(ctx context.Context, id string) (err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	o11y.SetAttributes(ctx, attribute.String("conversation_id", id))
	// 1. 获取数据
	_, err = svc.conversationRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			svc.logger.Errorf("[Delete] conversation not found, id: %s", id)
			o11y.Error(ctx, fmt.Sprintf("[Delete] conversation not found, id: %s", id))
			err = capierr.NewCustom404Err(ctx, apierr.ConversationNotFound, "数据智能体配置不存在")

			return
		}

		o11y.Error(ctx, fmt.Sprintf("[Delete] get conversation by id error, id: %s, err: %v", id, err))

		return
	}

	// 2. 开始事务
	tx, err := svc.conversationRepo.BeginTx(ctx)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[Delete] begin tx error, id: %s, err: %v", id, err))
		return
	}

	defer chelper.TxRollbackOrCommit(tx, &err, svc.logger)

	// 3. 调用repo层删除数据
	err = svc.conversationRepo.Delete(ctx, tx, id)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[Delete] delete conversation error, id: %s, err: %v", id, err))
		err = errors.Wrapf(err, "删除对话数据失败")

		return
	}

	err = svc.conversationMsgRepo.DeleteByConversationID(ctx, tx, id)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[Delete] delete conversation msg error, id: %s, err: %v", id, err))
		err = errors.Wrapf(err, "删除对话消息数据失败")

		return
	}

	return
}

// DeleteByAppKey implements iportdriver.IConversation.
func (svc *conversationSvc) DeleteByAppKey(ctx context.Context, appKey string) (err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	o11y.SetAttributes(ctx, attribute.String("app_key", appKey))
	// 1. 开始事务
	tx, err := svc.conversationRepo.BeginTx(ctx)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[DeleteByAppKey] begin tx error, appKey: %s, err: %v", appKey, err))
		return
	}

	defer chelper.TxRollbackOrCommit(tx, &err, svc.logger)

	// 2. 调用repo层删除数据
	err = svc.conversationRepo.DeleteByAPPKey(ctx, tx, appKey)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[DeleteByAppKey] delete conversation error, appKey: %s, err: %v", appKey, err))
		return errors.Wrapf(err, "删除对话数据失败")
	}

	err = svc.conversationMsgRepo.DeleteByAPPKey(ctx, tx, appKey)
	if err != nil {
		o11y.Error(ctx, fmt.Sprintf("[DeleteByAppKey] delete conversation msg error, appKey: %s, err: %v", appKey, err))
		return errors.Wrapf(err, "删除对话消息数据失败")
	}

	return
}
