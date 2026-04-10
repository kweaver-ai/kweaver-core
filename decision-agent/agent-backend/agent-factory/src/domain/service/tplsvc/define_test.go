package tplsvc

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewDataAgentTplService(t *testing.T) {
	t.Parallel()

	t.Run("creates service with all dependencies", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dto := &NewDaTplSvcDto{
			RedisCmp:          nil,
			SvcBase:           service.NewSvcBase(),
			AgentTplRepo:      idbaccessmock.NewMockIDataAgentTplRepo(ctrl),
			PublishedTplRepo:  idbaccessmock.NewMockIPublishedTplRepo(ctrl),
			AgentConfRepo:     idbaccessmock.NewMockIDataAgentConfigRepo(ctrl),
			Logger:            nil,
			UmHttp:            httpaccmock.NewMockUmHttpAcc(ctrl),
			CategorySvc:       nil,
			ProductRepo:       idbaccessmock.NewMockIProductRepo(ctrl),
			CategoryRepo:      idbaccessmock.NewMockICategoryRepo(ctrl),
			PmsSvc:            nil,
			BizDomainHttp:     bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl),
			BdAgentTplRelRepo: idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl),
		}

		svc := NewDataAgentTplService(dto)

		assert.NotNil(t, svc)
		assert.IsType(t, &dataAgentTplSvc{}, svc)
	})

	t.Run("creates service with minimal dependencies", func(t *testing.T) {
		t.Parallel()

		dto := &NewDaTplSvcDto{
			SvcBase:           service.NewSvcBase(),
			AgentTplRepo:      nil,
			PublishedTplRepo:  nil,
			AgentConfRepo:     nil,
			Logger:            nil,
			UmHttp:            nil,
			CategorySvc:       nil,
			ProductRepo:       nil,
			CategoryRepo:      nil,
			PmsSvc:            nil,
			BizDomainHttp:     nil,
			BdAgentTplRelRepo: nil,
		}

		svc := NewDataAgentTplService(dto)

		assert.NotNil(t, svc)
	})
}
