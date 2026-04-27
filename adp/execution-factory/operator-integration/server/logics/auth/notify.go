package auth

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

// NotifyResourceChange 通知资源变更
func (s *authServiceImpl) NotifyResourceChange(ctx context.Context, authResource *interfaces.AuthResource) error {
	jsonData, err := json.Marshal(authResource)
	if err != nil {
		s.logger.Errorf("marshal auth resource %v error: %v", authResource, err)
		return err
	}
	err = s.mqClient.Publish(ctx, interfaces.AuthResourceNameModifyTopic, jsonData)
	if err != nil {
		s.logger.Errorf("publish auth resource %v error: %v", authResource, err)
		return err
	}
	return nil
}

// BatchNotifyResourceChange(ctx context.Context, authResource []AuthResource) error
