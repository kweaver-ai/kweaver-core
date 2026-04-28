// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	llq "github.com/emirpasic/gods/queues/linkedlistqueue"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

func TestNewJobExecutor(t *testing.T) {
	Convey("Test NewJobExecutor", t, func() {
		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				ReloadJobEnabled:   true,
				MaxConcurrentTasks: 5,
			},
		}

		executor1 := NewJobExecutor(appSetting)
		executor2 := NewJobExecutor(appSetting)

		Convey("Should return singleton instance", func() {
			So(executor1, ShouldNotBeNil)
			So(executor2, ShouldEqual, executor1)
		})
	})
}

func TestJobExecutor_reloadJobs(t *testing.T) {
	Convey("Test reloadJobs", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				ReloadJobEnabled: true,
			},
		}

		ja := bmock.NewMockJobAccess(mockCtrl)
		db, _, _ := sqlmock.New()

		je := &jobExecutor{
			appSetting: appSetting,
			ja:         ja,
			db:         db,
		}

		Convey("Success reloading jobs", func() {
			tasks := []*interfaces.TaskInfo{
				{
					ID: "task1",
					TaskStateInfo: interfaces.TaskStateInfo{
						State: interfaces.TaskStateRunning,
					},
				},
			}

			jobs := []*interfaces.JobInfo{
				{
					ID: "job1",
					JobStateInfo: interfaces.JobStateInfo{
						State: interfaces.JobStateRunning,
					},
					CreateTime: time.Now().UnixMilli(),
				},
			}

			ja.EXPECT().ListTasks(ctx, gomock.Any()).Return(tasks, nil)
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(ctx, gomock.Any()).Return(jobs, nil)
			ja.EXPECT().UpdateJobState(ctx, nil, "job1", gomock.Any()).Return(nil)

			err := je.reloadJobs()
			So(err, ShouldBeNil)
		})

		Convey("Failed to list tasks", func() {
			ja.EXPECT().ListTasks(ctx, gomock.Any()).Return(nil, errors.New("db error"))

			err := je.reloadJobs()
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to list jobs", func() {
			ja.EXPECT().ListTasks(ctx, gomock.Any()).Return([]*interfaces.TaskInfo{}, nil)
			ja.EXPECT().ListJobs(ctx, gomock.Any()).Return(nil, errors.New("db error"))

			err := je.reloadJobs()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestJobExecutor_AddJob(t *testing.T) {
	Convey("Test AddJob", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				ViewDataLimit:    100,
				JobMaxRetryTimes: 3,
			},
		}

		ja := bmock.NewMockJobAccess(mockCtrl)
		ota := bmock.NewMockObjectTypeAccess(mockCtrl)

		je := &jobExecutor{
			appSetting: appSetting,
			ja:         ja,
			ota:        ota,
			mJobs:      make(map[string]*Job),
			mTaskQueue: llq.New(),
		}

		jobInfo := &interfaces.JobInfo{
			ID:     "job1",
			KNID:   "kn1",
			Branch: "main",
			TaskInfos: map[string]*interfaces.TaskInfo{
				"task1": {
					ID:          "task1",
					JobID:       "job1",
					ConceptID:   "ot1",
					ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				},
			},
		}

		je.mTaskQueue = llq.New()

		Convey("Success adding job", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
				},
			}

			ota.EXPECT().GetObjectTypeByID(ctx, gomock.Any(), "kn1", "main", "ot1").Return(objectType, nil)
			ja.EXPECT().UpdateJobState(ctx, nil, "job1", gomock.Any()).Return(nil)

			err := je.AddJob(ctx, jobInfo)
			So(err, ShouldBeNil)
			So(je.mJobs["job1"], ShouldNotBeNil)
		})

		Convey("Failed to get object type", func() {
			ota.EXPECT().GetObjectTypeByID(ctx, gomock.Any(), "kn1", "main", "ot1").Return(nil, errors.New("db error"))

			err := je.AddJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to update job state", func() {
			objectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
				},
			}

			ota.EXPECT().GetObjectTypeByID(ctx, gomock.Any(), "kn1", "main", "ot1").Return(objectType, nil)
			ja.EXPECT().UpdateJobState(ctx, nil, "job1", gomock.Any()).Return(errors.New("db error"))

			err := je.AddJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestJobExecutor_HandleTaskCallback(t *testing.T) {
	Convey("Test HandleTaskCallback", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}

		ja := bmock.NewMockJobAccess(mockCtrl)
		ota := bmock.NewMockObjectTypeAccess(mockCtrl)
		db, smock, _ := sqlmock.New()

		je := &jobExecutor{
			appSetting: appSetting,
			ja:         ja,
			ota:        ota,
			db:         db,
			mJobs:      make(map[string]*Job),
		}

		jobInfo := &interfaces.JobInfo{
			ID:         "job1",
			KNID:       "kn1",
			Branch:     "main",
			CreateTime: time.Now().UnixMilli(),
			Creator:    interfaces.AccountInfo{},
		}

		taskInfo := &interfaces.TaskInfo{
			ID:          "task1",
			Name:        "task1",
			JobID:       "job1",
			ConceptID:   "ot1",
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			TaskStateInfo: interfaces.TaskStateInfo{
				State:       interfaces.TaskStateCompleted,
				StateDetail: "",
			},
		}

		objectTypeTask := &ObjectTypeTask{
			taskInfo: taskInfo,
			objectTypeStatus: &interfaces.ObjectTypeStatus{
				IndexAvailable: true,
				DocCount:       10,
			},
		}

		job := &Job{
			mJobInfo:     jobInfo,
			mTasks:       map[string]Task{"task1": objectTypeTask},
			mFinishCount: 0,
		}

		je.mJobs["job1"] = job

		Convey("Success handling task callback with completed job", func() {
			// 设置 mFinishCount 为 len(mTasks) - 1，这样调用后会触发完成逻辑
			job.mFinishCount = len(job.mTasks) - 1

			smock.ExpectBegin()
			ota.EXPECT().UpdateObjectTypeStatus(gomock.Any(), gomock.Any(), "kn1", "main", "ot1", gomock.Any()).Return(nil)
			ja.EXPECT().UpdateJobState(gomock.Any(), gomock.Any(), "job1", gomock.Any()).Return(nil)
			smock.ExpectCommit()

			je.HandleTaskCallback(objectTypeTask)
			So(je.mJobs["job1"], ShouldBeNil)
		})

		Convey("Success handling task callback with failed task", func() {
			failedTaskInfo := &interfaces.TaskInfo{
				ID:          "task1",
				Name:        "task1",
				JobID:       "job1",
				ConceptID:   "ot1",
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				TaskStateInfo: interfaces.TaskStateInfo{
					State:       interfaces.TaskStateFailed,
					StateDetail: "task failed",
				},
			}
			failedTask := &ObjectTypeTask{
				taskInfo: failedTaskInfo,
				objectTypeStatus: &interfaces.ObjectTypeStatus{
					IndexAvailable: true,
					DocCount:       10,
				},
			}
			failedJob := &Job{
				mJobInfo:     jobInfo,
				mTasks:       map[string]Task{"task1": failedTask},
				mFinishCount: 0, // 设置为 len(mTasks) - 1 来触发完成逻辑
			}
			je.mJobs["job1"] = failedJob
			failedJob.mFinishCount = len(failedJob.mTasks) - 1

			ja.EXPECT().UpdateJobState(gomock.Any(), nil, "job1", gomock.Any()).Return(nil)

			je.HandleTaskCallback(failedTask)
			So(je.mJobs["job1"], ShouldBeNil)
		})

		Convey("Job not found", func() {
			delete(je.mJobs, "job1")

			je.HandleTaskCallback(objectTypeTask)
			// Should not panic
		})
	})
}

func TestJobExecutor_UpdateTaskStateFailed(t *testing.T) {
	Convey("Test UpdateTaskStateFailed", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ja := bmock.NewMockJobAccess(mockCtrl)

		je := &jobExecutor{
			ja: ja,
		}

		taskInfo := &interfaces.TaskInfo{
			ID: "task1",
			TaskStateInfo: interfaces.TaskStateInfo{
				StartTime: time.Now().UnixMilli(),
			},
		}

		Convey("Success updating task state to failed", func() {
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(nil)

			je.UpdateTaskStateFailed(ctx, taskInfo, errors.New("task error"))
			So(taskInfo.State, ShouldEqual, interfaces.TaskStateFailed)
		})

		Convey("Failed to update task state", func() {
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(errors.New("db error"))

			je.UpdateTaskStateFailed(ctx, taskInfo, errors.New("task error"))
			// Should not panic
		})
	})
}

func TestJobExecutor_UpdateTaskStateCompleted(t *testing.T) {
	Convey("Test UpdateTaskStateCompleted", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ja := bmock.NewMockJobAccess(mockCtrl)

		je := &jobExecutor{
			ja: ja,
		}

		taskInfo := &interfaces.TaskInfo{
			ID: "task1",
			TaskStateInfo: interfaces.TaskStateInfo{
				StartTime: time.Now().UnixMilli(),
			},
		}

		Convey("Success updating task state to completed", func() {
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(nil)

			je.UpdateTaskStateCompleted(ctx, taskInfo)
			So(taskInfo.State, ShouldEqual, interfaces.TaskStateCompleted)
		})

		Convey("Failed to update task state", func() {
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(errors.New("db error"))

			je.UpdateTaskStateCompleted(ctx, taskInfo)
			// Should not panic
		})
	})
}

func TestJobExecutor_HandleTask(t *testing.T) {
	Convey("Test HandleTask", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				ViewDataLimit:    100,
				JobMaxRetryTimes: 3,
			},
		}

		ja := bmock.NewMockJobAccess(mockCtrl)

		je := &jobExecutor{
			appSetting:        appSetting,
			ja:                ja,
			mJobs:             make(map[string]*Job),
			mTaskCallbackChan: make(chan Task, 100),
		}

		jobInfo := &interfaces.JobInfo{
			ID:      "job1",
			KNID:    "kn1",
			Branch:  "main",
			Creator: interfaces.AccountInfo{},
		}

		taskInfo := &interfaces.TaskInfo{
			ID:          "task1",
			JobID:       "job1",
			ConceptID:   "ot1",
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			TaskStateInfo: interfaces.TaskStateInfo{
				State: interfaces.TaskStatePending,
			},
		}

		objectType := &interfaces.ObjectType{
			ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
				OTID:        "ot1",
				PrimaryKeys: []string{"pk1"},
				DataSource: &interfaces.ResourceInfo{
					Type: "data_view",
					ID:   "dv1",
				},
				DataProperties: []*interfaces.DataProperty{
					{
						Name: "pk1",
						Type: "string",
						MappedField: &interfaces.Field{
							Name: "field1",
							Type: "string",
						},
					},
				},
			},
		}

		objectTypeTask := NewObjectTypeTask(appSetting, taskInfo, objectType)

		job := &Job{
			mJobInfo: jobInfo,
			mTasks:   map[string]Task{"task1": objectTypeTask},
		}

		je.mJobs["job1"] = job

		Convey("Failed with task not in pending state", func() {
			taskInfo.State = interfaces.TaskStateRunning
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(nil)
			err := je.HandleTask(ctx, objectTypeTask)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with job not found", func() {
			delete(je.mJobs, "job1")
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(nil)
			err := je.HandleTask(ctx, objectTypeTask)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to update task state to running", func() {
			ja.EXPECT().UpdateTaskState(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("db error")).AnyTimes()

			err := je.HandleTask(ctx, objectTypeTask)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestJobExecutor_reloadJobs_Errors(t *testing.T) {
	Convey("Test reloadJobs error cases\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				ReloadJobEnabled: true,
			},
		}

		ja := bmock.NewMockJobAccess(mockCtrl)
		db, _, _ := sqlmock.New()

		je := &jobExecutor{
			appSetting: appSetting,
			ja:         ja,
			db:         db,
		}

		Convey("Failed to update task state\n", func() {
			tasks := []*interfaces.TaskInfo{
				{
					ID: "task1",
					TaskStateInfo: interfaces.TaskStateInfo{
						State: interfaces.TaskStateRunning,
					},
				},
			}

			ja.EXPECT().ListTasks(ctx, gomock.Any()).Return(tasks, nil)
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(errors.New("db error"))

			err := je.reloadJobs()
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to update job state\n", func() {
			tasks := []*interfaces.TaskInfo{}
			jobs := []*interfaces.JobInfo{
				{
					ID: "job1",
					JobStateInfo: interfaces.JobStateInfo{
						State: interfaces.JobStateRunning,
					},
					CreateTime: time.Now().UnixMilli(),
				},
			}

			ja.EXPECT().ListTasks(ctx, gomock.Any()).Return(tasks, nil)
			ja.EXPECT().ListJobs(ctx, gomock.Any()).Return(jobs, nil)
			ja.EXPECT().UpdateJobState(ctx, nil, "job1", gomock.Any()).Return(errors.New("db error"))

			err := je.reloadJobs()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestJobExecutor_HandleTaskCallback_Errors(t *testing.T) {
	Convey("Test HandleTaskCallback error cases\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}

		ja := bmock.NewMockJobAccess(mockCtrl)
		ota := bmock.NewMockObjectTypeAccess(mockCtrl)
		db, smock, _ := sqlmock.New()

		je := &jobExecutor{
			appSetting: appSetting,
			ja:         ja,
			ota:        ota,
			db:         db,
			mJobs:      make(map[string]*Job),
		}

		jobInfo := &interfaces.JobInfo{
			ID:         "job1",
			KNID:       "kn1",
			Branch:     "main",
			CreateTime: time.Now().UnixMilli(),
			Creator:    interfaces.AccountInfo{},
		}

		taskInfo := &interfaces.TaskInfo{
			ID:          "task1",
			Name:        "task1",
			JobID:       "job1",
			ConceptID:   "ot1",
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			TaskStateInfo: interfaces.TaskStateInfo{
				State:       interfaces.TaskStateCompleted,
				StateDetail: "",
			},
		}

		objectTypeTask := &ObjectTypeTask{
			taskInfo: taskInfo,
			objectTypeStatus: &interfaces.ObjectTypeStatus{
				IndexAvailable: true,
				DocCount:       10,
			},
		}

		job := &Job{
			mJobInfo:     jobInfo,
			mTasks:       map[string]Task{"task1": objectTypeTask},
			mFinishCount: 0,
		}

		je.mJobs["job1"] = job

		Convey("Failed when UpdateObjectTypeStatus returns error\n", func() {
			job.mFinishCount = len(job.mTasks) - 1

			smock.ExpectBegin()
			ota.EXPECT().UpdateObjectTypeStatus(gomock.Any(), gomock.Any(), "kn1", "main", "ot1", gomock.Any()).Return(errors.New("db error"))
			smock.ExpectRollback()

			je.HandleTaskCallback(objectTypeTask)
			So(je.mJobs["job1"], ShouldNotBeNil) // Job should still exist
		})

		Convey("Failed when UpdateJobState returns error\n", func() {
			job.mFinishCount = len(job.mTasks) - 1

			smock.ExpectBegin()
			ota.EXPECT().UpdateObjectTypeStatus(gomock.Any(), gomock.Any(), "kn1", "main", "ot1", gomock.Any()).Return(nil)
			ja.EXPECT().UpdateJobState(gomock.Any(), gomock.Any(), "job1", gomock.Any()).Return(errors.New("db error"))
			smock.ExpectRollback()

			je.HandleTaskCallback(objectTypeTask)
			So(je.mJobs["job1"], ShouldNotBeNil) // Job should still exist
		})

		Convey("Failed when transaction commit returns error\n", func() {
			job.mFinishCount = len(job.mTasks) - 1

			smock.ExpectBegin()
			ota.EXPECT().UpdateObjectTypeStatus(gomock.Any(), gomock.Any(), "kn1", "main", "ot1", gomock.Any()).Return(nil)
			ja.EXPECT().UpdateJobState(gomock.Any(), gomock.Any(), "job1", gomock.Any()).Return(nil)
			smock.ExpectCommit().WillReturnError(errors.New("commit error"))

			je.HandleTaskCallback(objectTypeTask)
			So(je.mJobs["job1"], ShouldNotBeNil) // Job should still exist
		})
	})
}
