package toolbox

import (
	"context"
	"net/http"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

// HandleOperatorDeleteEvent 算子删除事件
func (s *ToolServiceImpl) HandleOperatorDeleteEvent(ctx context.Context, message []byte) error {
	// 记录可观测
	var err error
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"topic":   interfaces.OperatorDeleteEventTopic,
		"message": string(message),
	})
	s.Logger.WithContext(ctx).Debugf("handle operator delete event topic: %s, message: %s", interfaces.OperatorDeleteEventTopic, string(message))
	defer func() {
		if err != nil {
			s.Logger.WithContext(ctx).Debugf("handle operator delete event topic: %s, failed: message: %s, err: %v", interfaces.OperatorDeleteEventTopic, string(message), err)
		}
	}()
	// 解析消息，消息格式解析失败打印报错
	operatorDeleteEvent := &interfaces.OperatorDeleteEvent{}
	err = utils.StringToObject(string(message), operatorDeleteEvent)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("parse operator delete event failed, message: %s, err: %v", string(message), err)
		return nil
	}

	// 1. 根据OperatorID查询工具信息
	toolDBs, err := s.ToolDB.SelectToolBySource(ctx, model.SourceTypeOperator, operatorDeleteEvent.OperatorID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select tool by source failed, err: %v", err)
		return err
	}
	// 没有依赖改算子的工具不需要处理
	if len(toolDBs) == 0 {
		return nil
	}
	// 2. 根据工具信息删除工具
	for _, toolDB := range toolDBs {
		// 如果工具状态为禁用，直接跳过
		if toolDB.Status == interfaces.ToolStatusTypeDisabled.String() {
			continue
		}
		// 将工具置为禁用, 失败了直接返回报错，等待消息重新投递
		err = s.ToolDB.UpdateToolStatus(ctx, nil, toolDB.ToolID, interfaces.ToolStatusTypeDisabled.String(), operatorDeleteEvent.UpdateUser)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("update tool status failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return err
		}
	}
	return nil
}
