package skill

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

type testSkillIndexBuildCreateLocker struct {
	ok  bool
	err error
}

func (l *testSkillIndexBuildCreateLocker) Lock(ctx context.Context) (bool, error) {
	return l.ok, l.err
}

func (l *testSkillIndexBuildCreateLocker) Unlock(ctx context.Context) {}

type testSkillIndexBuildInspector struct {
	getTaskInfo      func(queue, id string) (*asynq.TaskInfo, error)
	deleteTask       func(queue, id string) error
	cancelProcessing func(id string) error
}

func (i *testSkillIndexBuildInspector) GetTaskInfo(queue, id string) (*asynq.TaskInfo, error) {
	return i.getTaskInfo(queue, id)
}

func (i *testSkillIndexBuildInspector) DeleteTask(queue, id string) error {
	if i.deleteTask == nil {
		return nil
	}
	return i.deleteTask(queue, id)
}

func (i *testSkillIndexBuildInspector) CancelProcessing(id string) error {
	if i.cancelProcessing == nil {
		return nil
	}
	return i.cancelProcessing(id)
}

func (i *testSkillIndexBuildInspector) Close() error { return nil }

type testSkillIndexBuildEnqueuer struct {
	enqueue func(ctx context.Context, payload *skillIndexBuildTaskPayload) error
}

func (e *testSkillIndexBuildEnqueuer) Enqueue(ctx context.Context, payload *skillIndexBuildTaskPayload) error {
	if e.enqueue == nil {
		return nil
	}
	return e.enqueue(ctx, payload)
}

func (e *testSkillIndexBuildEnqueuer) Close() error { return nil }

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

		Convey("CreateTask rejects when create lock is held", func() {
			svc := &skillIndexBuildService{
				logger: logger.DefaultLogger(),
				createLocker: func() skillIndexBuildCreateLocker {
					return &testSkillIndexBuildCreateLocker{ok: false}
				},
			}

			resp, err := svc.CreateTask(context.Background(), &interfaces.CreateSkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				ExecuteType:      interfaces.SkillIndexBuildExecuteTypeFull,
			})
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("CreateTask returns internal error when create lock fails", func() {
			svc := &skillIndexBuildService{
				logger: logger.DefaultLogger(),
				createLocker: func() skillIndexBuildCreateLocker {
					return &testSkillIndexBuildCreateLocker{err: errors.New("lock failed")}
				},
			}

			resp, err := svc.CreateTask(context.Background(), &interfaces.CreateSkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				ExecuteType:      interfaces.SkillIndexBuildExecuteTypeFull,
			})
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("RetryTask creates a new task for failed source task", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
				createLocker: func() skillIndexBuildCreateLocker {
					return &testSkillIndexBuildCreateLocker{ok: true}
				},
				enqueuer: &testSkillIndexBuildEnqueuer{},
			}
			gomock.InOrder(
				mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-failed").Return(&model.SkillIndexBuildTaskDB{
					TaskID:      "task-failed",
					Status:      interfaces.SkillIndexBuildStatusFailed.String(),
					ExecuteType: interfaces.SkillIndexBuildExecuteTypeIncremental.String(),
				}, nil),
				mockTaskRepo.EXPECT().SelectRunningTask(gomock.Any(), gomock.Nil()).Return(nil, nil),
				mockTaskRepo.EXPECT().SelectLatestCompletedIncrementalTask(gomock.Any(), gomock.Nil()).Return(nil, nil),
				mockTaskRepo.EXPECT().Insert(gomock.Any(), gomock.Nil(), gomock.Any()).Return(nil),
			)

			resp, err := svc.RetryTask(context.Background(), &interfaces.RetrySkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				TaskID:           "task-failed",
			})
			So(err, ShouldBeNil)
			So(resp.SourceTaskID, ShouldEqual, "task-failed")
			So(resp.TaskID, ShouldNotBeBlank)
			So(resp.Status, ShouldEqual, interfaces.SkillIndexBuildStatusPending)
			So(resp.ExecuteType, ShouldEqual, interfaces.SkillIndexBuildExecuteTypeIncremental.String())
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

		Convey("GetTask returns queue state from inspector", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
				inspector: &testSkillIndexBuildInspector{
					getTaskInfo: func(queue, id string) (*asynq.TaskInfo, error) {
						So(queue, ShouldEqual, skillIndexBuildQueueName)
						So(id, ShouldEqual, "task-1")
						return &asynq.TaskInfo{State: asynq.TaskStateRetry}, nil
					},
				},
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
			So(resp.QueueState, ShouldEqual, "retry")
		})

		Convey("QueryTaskList returns paged task list with queue states", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
				inspector: &testSkillIndexBuildInspector{
					getTaskInfo: func(queue, id string) (*asynq.TaskInfo, error) {
						if id == "task-l1" {
							return &asynq.TaskInfo{State: asynq.TaskStatePending}, nil
						}
						return &asynq.TaskInfo{State: asynq.TaskStateCompleted}, nil
					},
				},
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
			So(resp.Data[0].QueueState, ShouldEqual, "pending")
			So(resp.Data[1].QueueState, ShouldEqual, "completed")
		})

		Convey("CancelTask deletes queued task and marks task failed", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
				inspector: &testSkillIndexBuildInspector{
					getTaskInfo: func(queue, id string) (*asynq.TaskInfo, error) {
						return &asynq.TaskInfo{State: asynq.TaskStatePending}, nil
					},
					deleteTask: func(queue, id string) error {
						So(queue, ShouldEqual, skillIndexBuildQueueName)
						So(id, ShouldEqual, "task-2")
						return nil
					},
				},
			}
			mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-2").Return(&model.SkillIndexBuildTaskDB{
				TaskID: "task-2",
				Status: interfaces.SkillIndexBuildStatusPending.String(),
			}, nil)
			mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
					So(task.Status, ShouldEqual, interfaces.SkillIndexBuildStatusFailed.String())
					So(task.ErrorMsg, ShouldEqual, "task canceled by user")
					return nil
				})

			resp, err := svc.CancelTask(context.Background(), &interfaces.CancelSkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				TaskID:           "task-2",
			})
			So(err, ShouldBeNil)
			So(resp.Action, ShouldEqual, "delete_queue_task")
			So(resp.QueueState, ShouldEqual, "pending")
		})

		Convey("CancelTask cancels active processing task", func() {
			mockTaskRepo := mocks.NewMockISkillIndexBuildTaskDB(ctrl)
			svc := &skillIndexBuildService{
				logger:   logger.DefaultLogger(),
				taskRepo: mockTaskRepo,
				inspector: &testSkillIndexBuildInspector{
					getTaskInfo: func(queue, id string) (*asynq.TaskInfo, error) {
						return &asynq.TaskInfo{State: asynq.TaskStateActive}, nil
					},
					cancelProcessing: func(id string) error {
						So(id, ShouldEqual, "task-3")
						return nil
					},
				},
			}
			mockTaskRepo.EXPECT().SelectByTaskID(gomock.Any(), gomock.Nil(), "task-3").Return(&model.SkillIndexBuildTaskDB{
				TaskID: "task-3",
				Status: interfaces.SkillIndexBuildStatusRunning.String(),
			}, nil)
			mockTaskRepo.EXPECT().UpdateByTaskID(gomock.Any(), gomock.Nil(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
					So(task.ErrorMsg, ShouldEqual, "cancel requested by user")
					return nil
				})

			resp, err := svc.CancelTask(context.Background(), &interfaces.CancelSkillIndexBuildTaskReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				TaskID:           "task-3",
			})
			So(err, ShouldBeNil)
			So(resp.Action, ShouldEqual, "cancel_processing")
			So(resp.QueueState, ShouldEqual, "active")
		})

		Convey("RecoverRunningTasks is noop when task recovery is delegated to asynq", func() {
			svc := &skillIndexBuildService{
				logger: logger.DefaultLogger(),
			}
			err := svc.RecoverRunningTasks(context.Background())
			So(err, ShouldBeNil)
		})
	})
}
