package bizdomainsvc

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/stretchr/testify/assert"
)

func TestNewBizDomainService(t *testing.T) {
	t.Parallel()

	t.Run("creates service with all dependencies", func(t *testing.T) {
		t.Parallel()

		dto := &NewBizDomainSvcDto{
			SvcBase: service.NewSvcBase(),
		}

		svc := NewBizDomainService(dto)

		assert.NotNil(t, svc)
		assert.IsType(t, &BizDomainSvc{}, svc)
	})

	t.Run("creates service with minimal dependencies", func(t *testing.T) {
		t.Parallel()

		dto := &NewBizDomainSvcDto{
			SvcBase:       service.NewSvcBase(),
			Logger:        nil,
			BizDomainHttp: nil,
		}

		svc := NewBizDomainService(dto)

		assert.NotNil(t, svc)
	})
}
