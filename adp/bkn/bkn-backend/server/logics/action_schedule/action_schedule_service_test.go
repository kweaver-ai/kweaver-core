// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_schedule

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/robfig/cron/v3"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

// newTestService constructs an actionScheduleService with injected mocks.
func newTestService(t *testing.T) (*actionScheduleService, *gomock.Controller, *bmock.MockActionScheduleAccess, *bmock.MockActionTypeAccess) {
	t.Helper()
	mockCtrl := gomock.NewController(t)
	asa := bmock.NewMockActionScheduleAccess(mockCtrl)
	ata := bmock.NewMockActionTypeAccess(mockCtrl)
	svc := &actionScheduleService{
		appSetting: &common.AppSetting{},
		asa:        asa,
		ata:        ata,
		cronParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
	return svc, mockCtrl, asa, ata
}

// ---------- ValidateCronExpression ----------

func Test_actionScheduleService_ValidateCronExpression(t *testing.T) {
	Convey("Test ValidateCronExpression\n", t, func() {
		svc, mockCtrl, _, _ := newTestService(t)
		defer mockCtrl.Finish()

		Convey("Success with standard every-minute expression\n", func() {
			err := svc.ValidateCronExpression("* * * * *")
			So(err, ShouldBeNil)
		})

		Convey("Success with specific schedule expression\n", func() {
			err := svc.ValidateCronExpression("0 9 * * 1-5")
			So(err, ShouldBeNil)
		})

		Convey("Failed with completely invalid expression\n", func() {
			err := svc.ValidateCronExpression("invalid")
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with 6-field expression (not supported by 5-field parser)\n", func() {
			err := svc.ValidateCronExpression("0 * * * * *")
			So(err, ShouldNotBeNil)
		})
	})
}

// ---------- CalculateNextRunTime ----------

func Test_actionScheduleService_CalculateNextRunTime(t *testing.T) {
	Convey("Test CalculateNextRunTime\n", t, func() {
		svc, mockCtrl, _, _ := newTestService(t)
		defer mockCtrl.Finish()

		Convey("Success: next run time is after from time\n", func() {
			from := time.Now().UnixMilli()
			next, err := svc.CalculateNextRunTime("* * * * *", from)
			So(err, ShouldBeNil)
			So(next, ShouldBeGreaterThan, from)
		})

		Convey("Failed with invalid cron expression\n", func() {
			_, err := svc.CalculateNextRunTime("invalid", time.Now().UnixMilli())
			So(err, ShouldNotBeNil)
		})
	})
}

// ---------- CreateSchedule ----------

func Test_actionScheduleService_CreateSchedule(t *testing.T) {
	Convey("Test CreateSchedule\n", t, func() {
		ctx := context.Background()
		svc, mockCtrl, asa, ata := newTestService(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		branch := interfaces.MAIN_BRANCH

		Convey("Success with inactive status (default)\n", func() {
			schedule := &interfaces.ActionSchedule{
				KNID:           knID,
				Branch:         branch,
				ActionTypeID:   "at1",
				CronExpression: "* * * * *",
				Status:         interfaces.ScheduleStatusInactive,
			}
			ata.EXPECT().GetActionTypesByIDs(gomock.Any(), knID, branch, []string{"at1"}).
				Return([]*interfaces.ActionType{{ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{ATID: "at1"}}}, nil)
			asa.EXPECT().CreateSchedule(gomock.Any(), nil, gomock.Any()).Return(nil)

			id, err := svc.CreateSchedule(ctx, schedule)
			So(err, ShouldBeNil)
			So(id, ShouldNotBeEmpty)
		})

		Convey("Success with active status (calculates next run time)\n", func() {
			schedule := &interfaces.ActionSchedule{
				KNID:           knID,
				Branch:         branch,
				ActionTypeID:   "at1",
				CronExpression: "* * * * *",
				Status:         interfaces.ScheduleStatusActive,
			}
			ata.EXPECT().GetActionTypesByIDs(gomock.Any(), knID, branch, []string{"at1"}).
				Return([]*interfaces.ActionType{{ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{ATID: "at1"}}}, nil)
			asa.EXPECT().CreateSchedule(gomock.Any(), nil, gomock.Any()).Return(nil)

			id, err := svc.CreateSchedule(ctx, schedule)
			So(err, ShouldBeNil)
			So(id, ShouldNotBeEmpty)
			So(schedule.NextRunTime, ShouldBeGreaterThan, int64(0))
		})

		Convey("Failed with invalid cron expression\n", func() {
			schedule := &interfaces.ActionSchedule{
				KNID:           knID,
				Branch:         branch,
				ActionTypeID:   "at1",
				CronExpression: "invalid",
			}

			_, err := svc.CreateSchedule(ctx, schedule)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_InvalidCronExpression)
		})

		Convey("Failed when GetActionTypesByIDs returns error\n", func() {
			schedule := &interfaces.ActionSchedule{
				KNID:           knID,
				Branch:         branch,
				ActionTypeID:   "at1",
				CronExpression: "* * * * *",
			}
			ata.EXPECT().GetActionTypesByIDs(gomock.Any(), knID, branch, []string{"at1"}).
				Return(nil, errors.New("db error"))

			_, err := svc.CreateSchedule(ctx, schedule)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_GetActionTypeFailed)
		})

		Convey("Failed when action type not found\n", func() {
			schedule := &interfaces.ActionSchedule{
				KNID:           knID,
				Branch:         branch,
				ActionTypeID:   "at_missing",
				CronExpression: "* * * * *",
			}
			ata.EXPECT().GetActionTypesByIDs(gomock.Any(), knID, branch, []string{"at_missing"}).
				Return([]*interfaces.ActionType{}, nil)

			_, err := svc.CreateSchedule(ctx, schedule)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_ActionTypeNotFound)
		})

		Convey("Failed when asa.CreateSchedule returns error\n", func() {
			schedule := &interfaces.ActionSchedule{
				KNID:           knID,
				Branch:         branch,
				ActionTypeID:   "at1",
				CronExpression: "* * * * *",
			}
			ata.EXPECT().GetActionTypesByIDs(gomock.Any(), knID, branch, []string{"at1"}).
				Return([]*interfaces.ActionType{{ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{ATID: "at1"}}}, nil)
			asa.EXPECT().CreateSchedule(gomock.Any(), nil, gomock.Any()).Return(errors.New("db error"))

			_, err := svc.CreateSchedule(ctx, schedule)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_CreateFailed)
		})
	})
}

// ---------- UpdateSchedule ----------

func Test_actionScheduleService_UpdateSchedule(t *testing.T) {
	Convey("Test UpdateSchedule\n", t, func() {
		ctx := context.Background()
		svc, mockCtrl, asa, _ := newTestService(t)
		defer mockCtrl.Finish()

		scheduleID := "sched1"
		existing := &interfaces.ActionSchedule{
			ID:             scheduleID,
			CronExpression: "* * * * *",
			Status:         interfaces.ScheduleStatusInactive,
		}

		Convey("Success without cron change\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(existing, nil)
			asa.EXPECT().UpdateSchedule(gomock.Any(), nil, gomock.Any()).Return(nil)

			err := svc.UpdateSchedule(ctx, scheduleID, &interfaces.ActionScheduleUpdateRequest{Name: "new name"})
			So(err, ShouldBeNil)
		})

		Convey("Success with cron change on active schedule (recalculates next run)\n", func() {
			activeExisting := &interfaces.ActionSchedule{
				ID:             scheduleID,
				CronExpression: "0 8 * * *",
				Status:         interfaces.ScheduleStatusActive,
			}
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(activeExisting, nil)
			asa.EXPECT().UpdateSchedule(gomock.Any(), nil, gomock.Any()).Return(nil)

			err := svc.UpdateSchedule(ctx, scheduleID, &interfaces.ActionScheduleUpdateRequest{CronExpression: "0 9 * * *"})
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetSchedule returns error\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(nil, errors.New("db error"))

			err := svc.UpdateSchedule(ctx, scheduleID, &interfaces.ActionScheduleUpdateRequest{Name: "x"})
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_GetFailed)
		})

		Convey("Failed when schedule not found\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(nil, nil)

			err := svc.UpdateSchedule(ctx, scheduleID, &interfaces.ActionScheduleUpdateRequest{Name: "x"})
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_NotFound)
		})

		Convey("Failed with invalid new cron expression\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(existing, nil)

			err := svc.UpdateSchedule(ctx, scheduleID, &interfaces.ActionScheduleUpdateRequest{CronExpression: "invalid"})
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_InvalidCronExpression)
		})

		Convey("Failed when asa.UpdateSchedule returns error\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(existing, nil)
			asa.EXPECT().UpdateSchedule(gomock.Any(), nil, gomock.Any()).Return(errors.New("db error"))

			err := svc.UpdateSchedule(ctx, scheduleID, &interfaces.ActionScheduleUpdateRequest{Name: "x"})
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_UpdateFailed)
		})
	})
}

// ---------- UpdateScheduleStatus ----------

func Test_actionScheduleService_UpdateScheduleStatus(t *testing.T) {
	Convey("Test UpdateScheduleStatus\n", t, func() {
		ctx := context.Background()
		svc, mockCtrl, asa, _ := newTestService(t)
		defer mockCtrl.Finish()

		scheduleID := "sched1"
		existing := &interfaces.ActionSchedule{
			ID:             scheduleID,
			CronExpression: "* * * * *",
			Status:         interfaces.ScheduleStatusInactive,
		}

		Convey("Success activating schedule\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(existing, nil)
			asa.EXPECT().UpdateScheduleStatus(gomock.Any(), scheduleID, interfaces.ScheduleStatusActive, gomock.Any()).Return(nil)

			err := svc.UpdateScheduleStatus(ctx, scheduleID, interfaces.ScheduleStatusActive)
			So(err, ShouldBeNil)
		})

		Convey("Success deactivating schedule\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(existing, nil)
			asa.EXPECT().UpdateScheduleStatus(gomock.Any(), scheduleID, interfaces.ScheduleStatusInactive, int64(0)).Return(nil)

			err := svc.UpdateScheduleStatus(ctx, scheduleID, interfaces.ScheduleStatusInactive)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid status\n", func() {
			err := svc.UpdateScheduleStatus(ctx, scheduleID, "unknown")
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_InvalidStatus)
		})

		Convey("Failed when GetSchedule returns error\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(nil, errors.New("db error"))

			err := svc.UpdateScheduleStatus(ctx, scheduleID, interfaces.ScheduleStatusActive)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_GetFailed)
		})

		Convey("Failed when schedule not found\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(nil, nil)

			err := svc.UpdateScheduleStatus(ctx, scheduleID, interfaces.ScheduleStatusActive)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_NotFound)
		})

		Convey("Failed when UpdateScheduleStatus access returns error\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(existing, nil)
			asa.EXPECT().UpdateScheduleStatus(gomock.Any(), scheduleID, interfaces.ScheduleStatusInactive, int64(0)).Return(errors.New("db error"))

			err := svc.UpdateScheduleStatus(ctx, scheduleID, interfaces.ScheduleStatusInactive)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_UpdateFailed)
		})
	})
}

// ---------- DeleteSchedules ----------

func Test_actionScheduleService_DeleteSchedules(t *testing.T) {
	Convey("Test DeleteSchedules\n", t, func() {
		ctx := context.Background()
		svc, mockCtrl, asa, _ := newTestService(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		branch := interfaces.MAIN_BRANCH

		Convey("Success deleting schedules\n", func() {
			asa.EXPECT().GetSchedules(gomock.Any(), []string{"s1", "s2"}).Return(map[string]*interfaces.ActionSchedule{
				"s1": {ID: "s1", KNID: knID, Branch: branch},
				"s2": {ID: "s2", KNID: knID, Branch: branch},
			}, nil)
			asa.EXPECT().DeleteSchedules(gomock.Any(), nil, []string{"s1", "s2"}).Return(nil)

			err := svc.DeleteSchedules(ctx, knID, branch, []string{"s1", "s2"})
			So(err, ShouldBeNil)
		})

		Convey("Success with empty IDs (no-op)\n", func() {
			err := svc.DeleteSchedules(ctx, knID, branch, []string{})
			So(err, ShouldBeNil)
		})

		Convey("Failed when GetSchedules returns error\n", func() {
			asa.EXPECT().GetSchedules(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

			err := svc.DeleteSchedules(ctx, knID, branch, []string{"s1"})
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_GetFailed)
		})

		Convey("Failed when schedule not found in result\n", func() {
			asa.EXPECT().GetSchedules(gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ActionSchedule{}, nil)

			err := svc.DeleteSchedules(ctx, knID, branch, []string{"s_missing"})
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_NotFound)
		})

		Convey("Failed when schedule belongs to different kn\n", func() {
			asa.EXPECT().GetSchedules(gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ActionSchedule{
				"s1": {ID: "s1", KNID: "other-kn", Branch: branch},
			}, nil)

			err := svc.DeleteSchedules(ctx, knID, branch, []string{"s1"})
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_NotFound)
		})

		Convey("Failed when DeleteSchedules access returns error\n", func() {
			asa.EXPECT().GetSchedules(gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ActionSchedule{
				"s1": {ID: "s1", KNID: knID, Branch: branch},
			}, nil)
			asa.EXPECT().DeleteSchedules(gomock.Any(), nil, gomock.Any()).Return(errors.New("db error"))

			err := svc.DeleteSchedules(ctx, knID, branch, []string{"s1"})
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_DeleteFailed)
		})
	})
}

// ---------- GetSchedule ----------

func Test_actionScheduleService_GetSchedule(t *testing.T) {
	Convey("Test GetSchedule\n", t, func() {
		ctx := context.Background()
		svc, mockCtrl, asa, _ := newTestService(t)
		defer mockCtrl.Finish()

		scheduleID := "sched1"

		Convey("Success\n", func() {
			expected := &interfaces.ActionSchedule{ID: scheduleID}
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(expected, nil)

			result, err := svc.GetSchedule(ctx, scheduleID)
			So(err, ShouldBeNil)
			So(result.ID, ShouldEqual, scheduleID)
		})

		Convey("Failed when access returns error\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(nil, errors.New("db error"))

			_, err := svc.GetSchedule(ctx, scheduleID)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_GetFailed)
		})

		Convey("Failed when schedule not found (nil)\n", func() {
			asa.EXPECT().GetSchedule(gomock.Any(), scheduleID).Return(nil, nil)

			_, err := svc.GetSchedule(ctx, scheduleID)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_NotFound)
		})
	})
}

// ---------- GetSchedules ----------

func Test_actionScheduleService_GetSchedules(t *testing.T) {
	Convey("Test GetSchedules\n", t, func() {
		ctx := context.Background()
		svc, mockCtrl, asa, _ := newTestService(t)
		defer mockCtrl.Finish()

		Convey("Success\n", func() {
			expected := map[string]*interfaces.ActionSchedule{
				"s1": {ID: "s1"},
				"s2": {ID: "s2"},
			}
			asa.EXPECT().GetSchedules(gomock.Any(), []string{"s1", "s2"}).Return(expected, nil)

			result, err := svc.GetSchedules(ctx, []string{"s1", "s2"})
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
		})

		Convey("Failed when access returns error\n", func() {
			asa.EXPECT().GetSchedules(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

			_, err := svc.GetSchedules(ctx, []string{"s1"})
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_GetFailed)
		})
	})
}

// ---------- ListSchedules ----------

func Test_actionScheduleService_ListSchedules(t *testing.T) {
	Convey("Test ListSchedules\n", t, func() {
		ctx := context.Background()
		svc, mockCtrl, asa, _ := newTestService(t)
		defer mockCtrl.Finish()

		params := interfaces.ActionScheduleQueryParams{}

		Convey("Success\n", func() {
			asa.EXPECT().ListSchedules(gomock.Any(), params).Return([]*interfaces.ActionSchedule{{ID: "s1"}}, nil)
			asa.EXPECT().GetSchedulesTotal(gomock.Any(), params).Return(int64(1), nil)

			schedules, total, err := svc.ListSchedules(ctx, params)
			So(err, ShouldBeNil)
			So(len(schedules), ShouldEqual, 1)
			So(total, ShouldEqual, int64(1))
		})

		Convey("Failed when ListSchedules access returns error\n", func() {
			asa.EXPECT().ListSchedules(gomock.Any(), params).Return(nil, errors.New("db error"))

			_, _, err := svc.ListSchedules(ctx, params)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_GetFailed)
		})

		Convey("Failed when GetSchedulesTotal access returns error\n", func() {
			asa.EXPECT().ListSchedules(gomock.Any(), params).Return([]*interfaces.ActionSchedule{}, nil)
			asa.EXPECT().GetSchedulesTotal(gomock.Any(), params).Return(int64(0), errors.New("db error"))

			_, _, err := svc.ListSchedules(ctx, params)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ActionSchedule_GetFailed)
		})
	})
}
