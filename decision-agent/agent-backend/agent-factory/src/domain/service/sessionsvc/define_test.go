package sessionsvc

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/iredisaccess/isessionredis/isessionredismock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewSessionService(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dto := &NewSessionSvcDto{
		SessionRedis: isessionredismock.NewMockISessionRedisAcc(ctrl),
	}

	svc := NewSessionService(dto)

	assert.NotNil(t, svc)
}

func TestNewSessionService_WithLogger(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dto := &NewSessionSvcDto{
		Logger:       nil,
		SessionRedis: isessionredismock.NewMockISessionRedisAcc(ctrl),
	}

	svc := NewSessionService(dto)

	assert.NotNil(t, svc)
	assert.IsType(t, &sessionSvc{}, svc)
}

func TestNewSessionService_WithMinimalDependencies(t *testing.T) {
	t.Parallel()

	dto := &NewSessionSvcDto{
		SessionRedis: nil,
		Logger:       nil,
	}

	svc := NewSessionService(dto)

	assert.NotNil(t, svc)
}

func TestNewSessionService_WithNilDependencies(t *testing.T) {
	t.Parallel()

	dto := &NewSessionSvcDto{}

	svc := NewSessionService(dto)

	assert.NotNil(t, svc)
}
