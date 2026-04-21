package skill

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

type testSkillIndexBuildAssignLocker struct {
	lockOK  bool
	lockErr error
	locked  bool
}

func (l *testSkillIndexBuildAssignLocker) Lock(ctx context.Context) (bool, error) {
	l.locked = true
	return l.lockOK, l.lockErr
}

func (l *testSkillIndexBuildAssignLocker) Unlock(ctx context.Context) {
	l.locked = false
}

func TestSkillIndexBuildService(t *testing.T) {
	Convey("SkillIndexBuildService", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("CreateTask rejects when another task is running", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			mockTaskRepo.EXPECT().SelectRunningTask(gomock.Any(), gomock.Nil()).
				Return(&model.SkillIndexBuildTaskDB{TaskID: "task-1", Status: interfaces.SkillIndexBuildStatusRunning.String()}, nil)

			resp, err := svc.CreateTask(context.Background(), &interfaces.CreateSkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				ExecuteType:      interfaces.SkillIndexBuildExecuteTypeFull,
			})
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("CreateTask sets max retry to 3", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			mockTaskRepo.EXPECT().SelectRunningTask(gomock.Any(), gomock.Nil()).Return(nil, nil)
			mockTaskRepo.EXPECT().Insert(gomock.Any(), gomock.Nil(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
					So(task.MaxRetry, ShouldEqual, 3)
					So(task.Status, ShouldEqual, interfaces.SkillIndexBuildStatusPending.String())
					return nil
				})

			resp, err := svc.CreateTask(context.Background(), &interfaces.CreateSkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				ExecuteType:      interfaces.SkillIndexBuildExecuteTypeFull,
			})
			So(err, ShouldBeNil)
			So(resp.Status, ShouldEqual, interfaces.SkillIndexBuildStatusPending)
		})

		Convey("RetryTask creates a new task for failed source task", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-failed").Return(&model.SkillIndexBuildTaskDB{
				TaskID:           "task-failed",
				Status:           interfaces.SkillIndexBuildStatusFailed.String(),
				ExecuteType:      interfaces.SkillIndexBuildExecuteTypeIncremental.String(),
				TotalCount:       10,
				SuccessCount:     6,
				DeleteCount:      2,
				FailedCount:      2,
				RetryCount:       3,
				CursorUpdateTime: 999,
				CursorSkillID:    "skill-cursor",
				ErrorMsg:         "boom",
				LastFinishedTime: 12345,
			}, nil)
			mockTaskRepo.EXPECT().SelectRunningTask(gomock.Any(), gomock.Nil()).Return(nil, nil)
			mockTaskRepo.EXPECT().SelectLatestCompletedIncrementalTask(gomock.Any(), gomock.Nil()).Return(&model.SkillIndexBuildTaskDB{
				TaskID:           "last-completed",
				CursorUpdateTime: 100,
				CursorSkillID:    "skill-100",
			}, nil)
			mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
					So(task.TaskID, ShouldEqual, "task-failed")
					So(task.Status, ShouldEqual, interfaces.SkillIndexBuildStatusPending.String())
					So(task.TotalCount, ShouldEqual, 0)
					So(task.SuccessCount, ShouldEqual, 0)
					So(task.DeleteCount, ShouldEqual, 0)
					So(task.FailedCount, ShouldEqual, 0)
					So(task.RetryCount, ShouldEqual, 0)
					So(task.ErrorMsg, ShouldEqual, "")
					So(task.LastFinishedTime, ShouldEqual, 0)
					So(task.CursorUpdateTime, ShouldEqual, 100)
					So(task.CursorSkillID, ShouldEqual, "skill-100")
					return nil
				})

			resp, err := svc.RetryTask(context.Background(), &interfaces.RetrySkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				TaskID:           "task-failed",
			})
			So(err, ShouldBeNil)
			So(resp.SourceTaskID, ShouldEqual, "task-failed")
			So(resp.TaskID, ShouldEqual, "task-failed")
			So(resp.Status, ShouldEqual, interfaces.SkillIndexBuildStatusPending)
			So(resp.ExecuteType, ShouldEqual, interfaces.SkillIndexBuildExecuteTypeIncremental.String())
		})

		Convey("RetryTask creates a new task for canceled source task", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-canceled").Return(&model.SkillIndexBuildTaskDB{
				TaskID:           "task-canceled",
				Status:           interfaces.SkillIndexBuildStatusCanceled.String(),
				ExecuteType:      interfaces.SkillIndexBuildExecuteTypeFull.String(),
				TotalCount:       3,
				SuccessCount:     1,
				DeleteCount:      1,
				FailedCount:      1,
				RetryCount:       2,
				CursorUpdateTime: 777,
				CursorSkillID:    "skill-777",
				ErrorMsg:         "task canceled by user",
				LastFinishedTime: 88,
			}, nil)
			mockTaskRepo.EXPECT().SelectRunningTask(gomock.Any(), gomock.Nil()).Return(nil, nil)
			mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
					So(task.TaskID, ShouldEqual, "task-canceled")
					So(task.Status, ShouldEqual, interfaces.SkillIndexBuildStatusPending.String())
					So(task.TotalCount, ShouldEqual, 0)
					So(task.SuccessCount, ShouldEqual, 0)
					So(task.DeleteCount, ShouldEqual, 0)
					So(task.FailedCount, ShouldEqual, 0)
					So(task.RetryCount, ShouldEqual, 0)
					So(task.ErrorMsg, ShouldEqual, "")
					So(task.LastFinishedTime, ShouldEqual, 0)
					So(task.CursorUpdateTime, ShouldEqual, 0)
					So(task.CursorSkillID, ShouldEqual, "")
					return nil
				})

			resp, err := svc.RetryTask(context.Background(), &interfaces.RetrySkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				TaskID:           "task-canceled",
			})
			So(err, ShouldBeNil)
			So(resp.SourceTaskID, ShouldEqual, "task-canceled")
			So(resp.TaskID, ShouldEqual, "task-canceled")
			So(resp.Status, ShouldEqual, interfaces.SkillIndexBuildStatusPending)
			So(resp.ExecuteType, ShouldEqual, interfaces.SkillIndexBuildExecuteTypeFull.String())
		})

		Convey("RetryTask rejects non failed source task", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-running").Return(&model.SkillIndexBuildTaskDB{
				TaskID:      "task-running",
				Status:      interfaces.SkillIndexBuildStatusRunning.String(),
				ExecuteType: interfaces.SkillIndexBuildExecuteTypeFull.String(),
			}, nil)

			resp, err := svc.RetryTask(context.Background(), &interfaces.RetrySkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				TaskID:           "task-running",
			})
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("handleSkill deletes editing draft without release", func() {
			mockReleaseRepo := mocks.NewMockISkillReleaseDB(ctrl)
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			svc := &skillIndexBuildService{
				logger:      logger.DefaultLogger(),
				releaseRepo: mockReleaseRepo,
				indexSync:   mockIndexSync,
			}
			mockReleaseRepo.EXPECT().SelectBySkillID(gomock.Any(), gomock.Nil(), "skill-editing").Return(nil, nil)
			mockIndexSync.EXPECT().DeleteSkill(gomock.Any(), "skill-editing").Return(nil)

			action, err := svc.handleSkill(context.Background(), &model.SkillRepositoryDB{
				SkillID: "skill-editing",
				Status:  interfaces.BizStatusEditing.String(),
			})
			So(err, ShouldBeNil)
			So(action, ShouldEqual, "delete")
		})

		Convey("handleSkill upserts published release snapshot for editing skill", func() {
			mockReleaseRepo := mocks.NewMockISkillReleaseDB(ctrl)
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			svc := &skillIndexBuildService{
				logger:      logger.DefaultLogger(),
				releaseRepo: mockReleaseRepo,
				indexSync:   mockIndexSync,
			}
			mockReleaseRepo.EXPECT().SelectBySkillID(gomock.Any(), gomock.Nil(), "skill-1").Return(&model.SkillReleaseDB{
				SkillID:     "skill-1",
				Name:        "release-name",
				Description: "release-desc",
				Version:     "v1",
				Status:      interfaces.BizStatusPublished.String(),
				UpdateTime:  200,
			}, nil)
			mockIndexSync.EXPECT().UpsertSkill(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, skill *model.SkillRepositoryDB) error {
				So(skill.SkillID, ShouldEqual, "skill-1")
				So(skill.Name, ShouldEqual, "release-name")
				So(skill.Description, ShouldEqual, "release-desc")
				So(skill.Version, ShouldEqual, "v1")
				return nil
			})

			action, err := svc.handleSkill(context.Background(), &model.SkillRepositoryDB{
				SkillID: "skill-1",
				Status:  interfaces.BizStatusEditing.String(),
			})
			So(err, ShouldBeNil)
			So(action, ShouldEqual, "upsert")
		})

		Convey("handleSkill deletes soft deleted skill", func() {
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			svc := &skillIndexBuildService{
				logger:    logger.DefaultLogger(),
				indexSync: mockIndexSync,
			}
			mockIndexSync.EXPECT().DeleteSkill(gomock.Any(), "skill-deleted").Return(nil)

			action, err := svc.handleSkill(context.Background(), &model.SkillRepositoryDB{
				SkillID:   "skill-deleted",
				IsDeleted: true,
			})
			So(err, ShouldBeNil)
			So(action, ShouldEqual, "delete")
		})

		Convey("GetTask returns task without queue state", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-1").Return(&model.SkillIndexBuildTaskDB{
				TaskID: "task-1",
				Status: interfaces.SkillIndexBuildStatusFailed.String(),
			}, nil)

			resp, err := svc.GetTask(context.Background(), &interfaces.GetSkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				TaskID:           "task-1",
			})
			So(err, ShouldBeNil)
			So(resp.QueueState, ShouldEqual, "")
		})

		Convey("QueryTaskList returns paged task list without queue states", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			mockTaskRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).Return(int64(2), nil)
			mockTaskRepo.EXPECT().SelectListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).Return([]*model.SkillIndexBuildTaskDB{
				{TaskID: "task-l1", Status: interfaces.SkillIndexBuildStatusPending.String(), ExecuteType: interfaces.SkillIndexBuildExecuteTypeFull.String()},
				{TaskID: "task-l2", Status: interfaces.SkillIndexBuildStatusCompleted.String(), ExecuteType: interfaces.SkillIndexBuildExecuteTypeIncremental.String()},
			}, nil)

			resp, err := svc.QueryTaskList(context.Background(), &interfaces.QuerySkillIndexBuildTaskListReq{
				BusinessDomainID: "bd-1",
				CommonPageParams: interfaces.CommonPageParams{
					Page:      1,
					PageSize:  10,
					SortBy:    "update_time",
					SortOrder: "desc",
				},
			})
			So(err, ShouldBeNil)
			So(resp.TotalCount, ShouldEqual, 2)
			So(len(resp.Data), ShouldEqual, 2)
			So(resp.Data[0].QueueState, ShouldEqual, "")
			So(resp.Data[1].QueueState, ShouldEqual, "")
		})

		Convey("CancelTask marks pending task canceled", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-2").Return(&model.SkillIndexBuildTaskDB{
				TaskID: "task-2",
				Status: interfaces.SkillIndexBuildStatusPending.String(),
			}, nil)
			mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
					So(task.Status, ShouldEqual, interfaces.SkillIndexBuildStatusCanceled.String())
					So(task.ErrorMsg, ShouldEqual, "task canceled by user")
					return nil
				})

			resp, err := svc.CancelTask(context.Background(), &interfaces.CancelSkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				TaskID:           "task-2",
			})
			So(err, ShouldBeNil)
			So(resp.Action, ShouldEqual, "cancel_task")
			So(resp.QueueState, ShouldEqual, "")
		})

		Convey("CancelTask marks running task canceled", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-3").Return(&model.SkillIndexBuildTaskDB{
				TaskID: "task-3",
				Status: interfaces.SkillIndexBuildStatusRunning.String(),
			}, nil)
			mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
					So(task.Status, ShouldEqual, interfaces.SkillIndexBuildStatusCanceled.String())
					So(task.ErrorMsg, ShouldEqual, "task canceled by user")
					return nil
				})

			resp, err := svc.CancelTask(context.Background(), &interfaces.CancelSkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				TaskID:           "task-3",
			})
			So(err, ShouldBeNil)
			So(resp.Action, ShouldEqual, "cancel_task")
			So(resp.QueueState, ShouldEqual, "")
		})

		Convey("tryStartPendingTask marks task running after acquiring lock", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			locker := &testSkillIndexBuildAssignLocker{lockOK: true}
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
				assignLockerFactory: func(taskID string) skillIndexBuildAssignLocker {
					So(taskID, ShouldEqual, "task-4")
					return locker
				},
			}
			gomock.InOrder(
				mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-4").Return(&model.SkillIndexBuildTaskDB{
					TaskID: "task-4",
					Status: interfaces.SkillIndexBuildStatusPending.String(),
				}, nil),
				mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
						So(task.Status, ShouldEqual, interfaces.SkillIndexBuildStatusRunning.String())
						return nil
					}),
			)

			ok, err := svc.tryStartPendingTask(context.Background(), "task-4")
			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)
			So(locker.locked, ShouldBeFalse)
		})

		Convey("tryStartPendingTask skips task when lock is not acquired", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
				assignLockerFactory: func(taskID string) skillIndexBuildAssignLocker {
					return &testSkillIndexBuildAssignLocker{lockOK: false}
				},
			}

			ok, err := svc.tryStartPendingTask(context.Background(), "task-5")
			So(err, ShouldBeNil)
			So(ok, ShouldBeFalse)
		})

		Convey("tryStartPendingTask returns lock errors", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
				assignLockerFactory: func(taskID string) skillIndexBuildAssignLocker {
					return &testSkillIndexBuildAssignLocker{lockErr: errors.New("lock failed")}
				},
			}

			ok, err := svc.tryStartPendingTask(context.Background(), "task-6")
			So(err, ShouldNotBeNil)
			So(ok, ShouldBeFalse)
		})

		Convey("recoverStaleRunningTask marks stale running task failed", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			staleTime := time.Now().Add(-skillIndexBuildRunningTimeout - time.Minute).UnixNano()
			mockTaskRepo.EXPECT().SelectRunningTask(gomock.Any(), gomock.Nil()).Return(&model.SkillIndexBuildTaskDB{
				TaskID:     "task-7",
				Status:     interfaces.SkillIndexBuildStatusRunning.String(),
				UpdateTime: staleTime,
			}, nil)
			mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
					So(task.Status, ShouldEqual, interfaces.SkillIndexBuildStatusFailed.String())
					So(task.ErrorMsg, ShouldEqual, "stale running task recovered as failed")
					return nil
				})

			err := svc.recoverStaleRunningTask(context.Background())
			So(err, ShouldBeNil)
		})

		Convey("failTask reschedules task when retries remain", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			mockTaskRepo.EXPECT().SelectLatestCompletedIncrementalTask(gomock.Any(), gomock.Nil()).Return(&model.SkillIndexBuildTaskDB{
				TaskID:           "completed",
				CursorUpdateTime: 321,
				CursorSkillID:    "skill-321",
			}, nil)
			mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
					So(task.Status, ShouldEqual, interfaces.SkillIndexBuildStatusPending.String())
					So(task.RetryCount, ShouldEqual, 1)
					So(task.TotalCount, ShouldEqual, 0)
					So(task.SuccessCount, ShouldEqual, 0)
					So(task.DeleteCount, ShouldEqual, 0)
					So(task.FailedCount, ShouldEqual, 0)
					So(task.ErrorMsg, ShouldEqual, "")
					So(task.LastFinishedTime, ShouldEqual, 0)
					So(task.CursorUpdateTime, ShouldEqual, 321)
					So(task.CursorSkillID, ShouldEqual, "skill-321")
					return nil
				})

			err := svc.failTask(context.Background(), &model.SkillIndexBuildTaskDB{
				TaskID:           "task-8",
				Status:           interfaces.SkillIndexBuildStatusRunning.String(),
				ExecuteType:      interfaces.SkillIndexBuildExecuteTypeIncremental.String(),
				TotalCount:       11,
				SuccessCount:     7,
				DeleteCount:      2,
				FailedCount:      2,
				RetryCount:       0,
				MaxRetry:         3,
				CursorUpdateTime: 999,
				CursorSkillID:    "skill-999",
			}, errors.New("transient"))
			So(err, ShouldNotBeNil)
		})

		Convey("failTask marks task failed after retries exhausted", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
			}
			mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
					So(task.Status, ShouldEqual, interfaces.SkillIndexBuildStatusFailed.String())
					So(task.RetryCount, ShouldEqual, 3)
					So(task.ErrorMsg, ShouldEqual, "still broken")
					return nil
				})

			err := svc.failTask(context.Background(), &model.SkillIndexBuildTaskDB{
				TaskID:      "task-9",
				Status:      interfaces.SkillIndexBuildStatusRunning.String(),
				ExecuteType: interfaces.SkillIndexBuildExecuteTypeFull.String(),
				RetryCount:  3,
				MaxRetry:    3,
			}, errors.New("still broken"))
			So(err, ShouldNotBeNil)
		})

		Convey("runTask persists accumulated counters", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockReleaseRepo := mocks.NewMockISkillReleaseDB(ctrl)
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			svc := &skillIndexBuildService{
				logger:      logger.DefaultLogger(),
				taskRepo:    mockTaskRepo,
				skillRepo:   mockSkillRepo,
				releaseRepo: mockReleaseRepo,
				indexSync:   mockIndexSync,
			}
			gomock.InOrder(
				mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-10").Return(&model.SkillIndexBuildTaskDB{
					TaskID:      "task-10",
					Status:      interfaces.SkillIndexBuildStatusRunning.String(),
					ExecuteType: interfaces.SkillIndexBuildExecuteTypeFull.String(),
				}, nil),
				mockIndexSync.EXPECT().EnsureInitialized(gomock.Any()).Return(nil),
				mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-10").Return(&model.SkillIndexBuildTaskDB{
					TaskID:      "task-10",
					Status:      interfaces.SkillIndexBuildStatusRunning.String(),
					ExecuteType: interfaces.SkillIndexBuildExecuteTypeFull.String(),
				}, nil),
				mockSkillRepo.EXPECT().SelectSkillBuildPage(gomock.Any(), gomock.Nil(), int64(0), "", skillIndexBuildBatchSize).Return([]*model.SkillRepositoryDB{
					{SkillID: "skill-upsert", Status: interfaces.BizStatusPublished.String(), UpdateTime: 100},
					{SkillID: "skill-delete", Status: interfaces.BizStatusOffline.String(), UpdateTime: 101},
				}, nil),
				mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-10").Return(&model.SkillIndexBuildTaskDB{
					TaskID:      "task-10",
					Status:      interfaces.SkillIndexBuildStatusRunning.String(),
					ExecuteType: interfaces.SkillIndexBuildExecuteTypeFull.String(),
				}, nil),
				mockReleaseRepo.EXPECT().SelectBySkillID(gomock.Any(), gomock.Nil(), "skill-upsert").Return(nil, nil),
				mockIndexSync.EXPECT().UpsertSkill(gomock.Any(), gomock.Any()).Return(nil),
				mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-10").Return(&model.SkillIndexBuildTaskDB{
					TaskID:      "task-10",
					Status:      interfaces.SkillIndexBuildStatusRunning.String(),
					ExecuteType: interfaces.SkillIndexBuildExecuteTypeFull.String(),
				}, nil),
				mockReleaseRepo.EXPECT().SelectBySkillID(gomock.Any(), gomock.Nil(), "skill-delete").Return(nil, nil),
				mockIndexSync.EXPECT().DeleteSkill(gomock.Any(), "skill-delete").Return(nil),
				mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
						So(task.TotalCount, ShouldEqual, 2)
						So(task.SuccessCount, ShouldEqual, 1)
						So(task.DeleteCount, ShouldEqual, 1)
						So(task.FailedCount, ShouldEqual, 0)
						So(task.CursorUpdateTime, ShouldEqual, 101)
						So(task.CursorSkillID, ShouldEqual, "skill-delete")
						return nil
					}),
				mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-10").Return(&model.SkillIndexBuildTaskDB{
					TaskID:           "task-10",
					Status:           interfaces.SkillIndexBuildStatusRunning.String(),
					ExecuteType:      interfaces.SkillIndexBuildExecuteTypeFull.String(),
					TotalCount:       2,
					SuccessCount:     1,
					DeleteCount:      1,
					FailedCount:      0,
					CursorUpdateTime: 101,
					CursorSkillID:    "skill-delete",
				}, nil),
				mockSkillRepo.EXPECT().SelectSkillBuildPage(gomock.Any(), gomock.Nil(), int64(101), "skill-delete", skillIndexBuildBatchSize).Return([]*model.SkillRepositoryDB{}, nil),
				mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
						So(task.Status, ShouldEqual, interfaces.SkillIndexBuildStatusCompleted.String())
						So(task.TotalCount, ShouldEqual, 2)
						So(task.SuccessCount, ShouldEqual, 1)
						So(task.DeleteCount, ShouldEqual, 1)
						So(task.FailedCount, ShouldEqual, 0)
						return nil
					}),
			)

			err := svc.runTask(context.Background(), "task-10")
			So(err, ShouldBeNil)
		})
	})
}
