package squaresvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/daconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRecordVisitLog(t *testing.T) {
	t.Parallel()

	t.Run("returns_early_when_IsVisit_is_false", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVisitHistoryRepo := idbaccessmock.NewMockIVisitHistoryRepo(ctrl)

		svc := &squareSvc{
			SvcBase:          service.NewSvcBase(),
			visitHistoryRepo: mockVisitHistoryRepo,
		}

		ctx := context.Background()
		req := &squarereq.AgentInfoReq{
			AgentID:      "agent-123",
			AgentVersion: "v1.0.0",
			IsVisit:      false, // Not a visit
		}

		err := svc.RecordVisitLog(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("records_visit_for_published_version", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVisitHistoryRepo := idbaccessmock.NewMockIVisitHistoryRepo(ctrl)
		mockVisitHistoryRepo.EXPECT().IncVisitCount(gomock.Any(), gomock.Any()).Return(nil)

		svc := &squareSvc{
			SvcBase:          service.NewSvcBase(),
			visitHistoryRepo: mockVisitHistoryRepo,
		}

		ctx := context.Background()
		req := &squarereq.AgentInfoReq{
			AgentID:      "agent-123",
			AgentVersion: "v1.0.0",
			UserID:       "user-456",
			IsVisit:      true,
		}

		err := svc.RecordVisitLog(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("records_visit_for_unpublished_version", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVisitHistoryRepo := idbaccessmock.NewMockIVisitHistoryRepo(ctrl)
		mockVisitHistoryRepo.EXPECT().IncVisitCount(gomock.Any(), gomock.Any()).Return(nil)

		svc := &squareSvc{
			SvcBase:          service.NewSvcBase(),
			visitHistoryRepo: mockVisitHistoryRepo,
		}

		ctx := context.Background()
		req := &squarereq.AgentInfoReq{
			AgentID:      "agent-123",
			AgentVersion: daconstant.AgentVersionUnpublished,
			UserID:       "user-456",
			IsVisit:      true,
		}

		err := svc.RecordVisitLog(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("returns_error_from_IncVisitCount", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVisitHistoryRepo := idbaccessmock.NewMockIVisitHistoryRepo(ctrl)
		mockVisitHistoryRepo.EXPECT().IncVisitCount(gomock.Any(), gomock.Any()).Return(errors.New("database error"))

		svc := &squareSvc{
			SvcBase:          service.NewSvcBase(),
			visitHistoryRepo: mockVisitHistoryRepo,
		}

		ctx := context.Background()
		req := &squarereq.AgentInfoReq{
			AgentID:      "agent-123",
			AgentVersion: "v1.0.0",
			UserID:       "user-456",
			IsVisit:      true,
		}

		err := svc.RecordVisitLog(ctx, req)

		assert.Error(t, err) // Should return error from IncVisitCount
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("records_visit_with_empty_custom_space_id", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVisitHistoryRepo := idbaccessmock.NewMockIVisitHistoryRepo(ctrl)
		mockVisitHistoryRepo.EXPECT().IncVisitCount(gomock.Any(), gomock.Any()).Return(nil)

		svc := &squareSvc{
			SvcBase:          service.NewSvcBase(),
			visitHistoryRepo: mockVisitHistoryRepo,
		}

		ctx := context.Background()
		req := &squarereq.AgentInfoReq{
			AgentID:      "agent-123",
			AgentVersion: "v1.0.0",
			UserID:       "user-456",
			IsVisit:      true,
		}

		err := svc.RecordVisitLog(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("records_visit_with_empty_user_id", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVisitHistoryRepo := idbaccessmock.NewMockIVisitHistoryRepo(ctrl)
		mockVisitHistoryRepo.EXPECT().IncVisitCount(gomock.Any(), gomock.Any()).Return(nil)

		svc := &squareSvc{
			SvcBase:          service.NewSvcBase(),
			visitHistoryRepo: mockVisitHistoryRepo,
		}

		ctx := context.Background()
		req := &squarereq.AgentInfoReq{
			AgentID:      "agent-123",
			AgentVersion: "v1.0.0",
			UserID:       "",
			IsVisit:      true,
		}

		err := svc.RecordVisitLog(ctx, req)

		assert.NoError(t, err)
	})
}
