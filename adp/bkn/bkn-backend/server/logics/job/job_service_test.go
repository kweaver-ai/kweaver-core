// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package job

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

func Test_jobService_CreateJob(t *testing.T) {
	Convey("Test CreateJob\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ja := bmock.NewMockJobAccess(mockCtrl)
		je := bmock.NewMockJobExecutor(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		ums := bmock.NewMockUserMgmtService(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &jobService{
			appSetting: appSetting,
			db:         db,
			ja:         ja,
			je:         je,
			ps:         ps,
			ots:        ots,
			ums:        ums,
		}

		Convey("Success creating job without JobConceptConfig, auto-generate from object types\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						OTName:      "object_type1",
						PrimaryKeys: []string{"pk1"},
						DataSource:  &interfaces.ResourceInfo{},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			smock.ExpectBegin()
			ja.EXPECT().CreateJob(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().CreateTasks(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()
			je.EXPECT().AddJob(gomock.Any(), gomock.Any()).Return(nil)

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldBeNil)
			So(jobID, ShouldEqual, "job1")
		})

		Convey("Success creating job without JobConceptConfig with multiple object types\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						OTName:      "object_type1",
						PrimaryKeys: []string{"pk1"},
						DataSource:  &interfaces.ResourceInfo{},
					},
				},
				"ot2": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot2",
						OTName:      "object_type2",
						PrimaryKeys: []string{"pk2"},
						DataSource:  &interfaces.ResourceInfo{},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			smock.ExpectBegin()
			ja.EXPECT().CreateJob(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().CreateTasks(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()
			je.EXPECT().AddJob(gomock.Any(), gomock.Any()).Return(nil)

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldBeNil)
			So(jobID, ShouldEqual, "job1")
		})

		Convey("Success creating job with auto-generated ID\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						OTName:      "object_type1",
						PrimaryKeys: []string{"pk1"},
						DataSource:  &interfaces.ResourceInfo{},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			smock.ExpectBegin()
			ja.EXPECT().CreateJob(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, tx *sql.Tx, job *interfaces.JobInfo) {
				So(job.ID, ShouldNotBeEmpty)
			}).Return(nil)
			ja.EXPECT().CreateTasks(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit()
			je.EXPECT().AddJob(gomock.Any(), gomock.Any()).Return(nil)

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldBeNil)
			So(jobID, ShouldNotBeEmpty)
		})

		Convey("Failed when permission check fails\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_Job_InternalError))

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
			So(jobID, ShouldEqual, "")
		})

		Convey("Failed when there is already a running or pending job\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			existingJobs := []*interfaces.JobInfo{
				{
					ID: "job2",
					JobStateInfo: interfaces.JobStateInfo{
						State: interfaces.JobStateRunning,
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return(existingJobs, nil)

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
			So(jobID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_CreateConflict)
		})

		Convey("Failed when object type has no primary key\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						OTName:      "object_type1",
						PrimaryKeys: []string{},
						DataSource:  &interfaces.ResourceInfo{},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
			So(jobID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InvalidObjectType)
		})

		Convey("Failed when object type has no data source\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						OTName:      "object_type1",
						PrimaryKeys: []string{"pk1"},
						DataSource:  nil,
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
			So(jobID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_NoneConceptType)
		})

		Convey("Failed when JobConceptConfig is empty and no object types\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
			So(jobID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_NoneConceptType)
		})

		Convey("Failed when only resource-backed object types exist\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						OTName:      "object_type1",
						PrimaryKeys: []string{"pk1"},
						DataSource: &interfaces.ResourceInfo{
							Type: interfaces.DATA_SOURCE_TYPE_RESOURCE,
							ID:   "res1",
						},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
			So(jobID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_NoneConceptType)
		})

		Convey("Success creating job skips resource object types and keeps data_view tasks\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						OTName:      "dv_ot",
						PrimaryKeys: []string{"pk1"},
						DataSource: &interfaces.ResourceInfo{
							Type: interfaces.DATA_SOURCE_TYPE_DATA_VIEW,
							ID:   "dv1",
						},
					},
				},
				"ot2": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot2",
						OTName:      "res_ot",
						PrimaryKeys: []string{"pk1"},
						DataSource: &interfaces.ResourceInfo{
							Type: interfaces.DATA_SOURCE_TYPE_RESOURCE,
							ID:   "res1",
						},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			smock.ExpectBegin()
			ja.EXPECT().CreateJob(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().CreateTasks(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, tx *sql.Tx, tasks map[string]*interfaces.TaskInfo) {
				So(len(tasks), ShouldEqual, 1)
				for _, ti := range tasks {
					So(ti.ConceptID, ShouldEqual, "ot1")
				}
			}).Return(nil)
			smock.ExpectCommit()
			je.EXPECT().AddJob(gomock.Any(), gomock.Any()).Return(nil)

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldBeNil)
			So(jobID, ShouldEqual, "job1")
		})

		Convey("Failed when begin transaction fails\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						OTName:      "object_type1",
						PrimaryKeys: []string{"pk1"},
						DataSource:  &interfaces.ResourceInfo{},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			smock.ExpectBegin().WillReturnError(errors.New("begin transaction error"))

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
			So(jobID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError_BeginTransactionFailed)
		})

		Convey("Failed when CreateJob fails\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						OTName:      "object_type1",
						PrimaryKeys: []string{"pk1"},
						DataSource:  &interfaces.ResourceInfo{},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			smock.ExpectBegin()
			ja.EXPECT().CreateJob(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("create job error"))
			smock.ExpectRollback()

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
			So(jobID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError)
		})

		Convey("Failed when CreateTasks fails\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						OTName:      "object_type1",
						PrimaryKeys: []string{"pk1"},
						DataSource:  &interfaces.ResourceInfo{},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			smock.ExpectBegin()
			ja.EXPECT().CreateJob(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().CreateTasks(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("create tasks error"))
			smock.ExpectRollback()

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
			So(jobID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError)
		})

		Convey("Failed when commit transaction fails\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			objectTypes := map[string]*interfaces.ObjectType{
				"ot1": {
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						OTName:      "object_type1",
						PrimaryKeys: []string{"pk1"},
						DataSource:  &interfaces.ResourceInfo{},
					},
				},
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(objectTypes, nil)
			smock.ExpectBegin()
			ja.EXPECT().CreateJob(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().CreateTasks(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectCommit().WillReturnError(errors.New("commit error"))

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
			So(jobID, ShouldEqual, "")
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError_CommitTransactionFailed)
		})

		Convey("Failed when GetAllObjectTypesByKnID fails\n", func() {
			jobInfo := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, nil)
			ots.EXPECT().GetAllObjectTypesByKnID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("get object types error"))

			jobID, err := service.CreateJob(ctx, jobInfo)
			So(err, ShouldNotBeNil)
			So(jobID, ShouldEqual, "")
		})
	})
}

func Test_jobService_DeleteJobsByIDs(t *testing.T) {
	Convey("Test DeleteJobsByIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ja := bmock.NewMockJobAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		db, smock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &jobService{
			appSetting: appSetting,
			db:         db,
			ja:         ja,
			ps:         ps,
		}

		Convey("Success deleting jobs\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			jobIDs := []string{"job1", "job2"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ja.EXPECT().DeleteJobsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			ja.EXPECT().DeleteTasksByJobIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			smock.ExpectCommit()

			err := service.DeleteJobsByIDs(ctx, knID, branch, jobIDs)
			So(err, ShouldBeNil)
		})

		Convey("Failed when permission check fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			jobIDs := []string{"job1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_Job_InternalError))

			err := service.DeleteJobsByIDs(ctx, knID, branch, jobIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when begin transaction fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			jobIDs := []string{"job1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin().WillReturnError(errors.New("begin transaction error"))

			err := service.DeleteJobsByIDs(ctx, knID, branch, jobIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteJobsByIDs fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			jobIDs := []string{"job1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ja.EXPECT().DeleteJobsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(int64(0), errors.New("delete jobs error"))
			smock.ExpectRollback()

			err := service.DeleteJobsByIDs(ctx, knID, branch, jobIDs)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError)
		})

		Convey("Failed when DeleteTasksByJobIDs fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			jobIDs := []string{"job1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ja.EXPECT().DeleteJobsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			ja.EXPECT().DeleteTasksByJobIDs(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(int64(0), errors.New("delete tasks error"))
			smock.ExpectRollback()

			err := service.DeleteJobsByIDs(ctx, knID, branch, jobIDs)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError)
		})

		Convey("Failed when commit transaction fails\n", func() {
			knID := "kn1"
			branch := interfaces.MAIN_BRANCH
			jobIDs := []string{"job1"}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			smock.ExpectBegin()
			ja.EXPECT().DeleteJobsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			ja.EXPECT().DeleteTasksByJobIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil)
			smock.ExpectCommit().WillReturnError(errors.New("commit error"))

			err := service.DeleteJobsByIDs(ctx, knID, branch, jobIDs)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_jobService_ListJobs(t *testing.T) {
	Convey("Test ListJobs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ja := bmock.NewMockJobAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		ums := bmock.NewMockUserMgmtService(mockCtrl)
		db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &jobService{
			appSetting: appSetting,
			db:         db,
			ja:         ja,
			ps:         ps,
			ums:        ums,
		}

		Convey("Success listing jobs\n", func() {
			queryParams := interfaces.JobsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			jobs := []*interfaces.JobInfo{
				{
					ID:     "job1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
					Creator: interfaces.AccountInfo{
						ID:   "user1",
						Type: "user",
						Name: "user1",
					},
				},
			}
			total := int64(1)
			accountInfos := []*interfaces.AccountInfo{
				&jobs[0].Creator,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return(jobs, nil)
			ja.EXPECT().GetJobsTotal(gomock.Any(), gomock.Any()).Return(total, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), accountInfos).Return(nil)

			result, resultTotal, err := service.ListJobs(ctx, queryParams)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(resultTotal, ShouldEqual, total)
		})

		Convey("Failed when permission check fails\n", func() {
			queryParams := interfaces.JobsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_Job_InternalError))

			result, resultTotal, err := service.ListJobs(ctx, queryParams)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(resultTotal, ShouldEqual, 0)
		})

		Convey("Failed when ListJobs fails\n", func() {
			queryParams := interfaces.JobsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return(nil, errors.New("list jobs error"))

			result, resultTotal, err := service.ListJobs(ctx, queryParams)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(resultTotal, ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError)
		})

		Convey("Failed when GetJobsTotal fails\n", func() {
			queryParams := interfaces.JobsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			jobs := []*interfaces.JobInfo{}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return(jobs, nil)
			ja.EXPECT().GetJobsTotal(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("get jobs total error"))

			result, resultTotal, err := service.ListJobs(ctx, queryParams)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(resultTotal, ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError)
		})

		Convey("Failed when GetAccountNames fails\n", func() {
			queryParams := interfaces.JobsQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			jobs := []*interfaces.JobInfo{
				{
					ID:     "job1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
					Creator: interfaces.AccountInfo{
						ID:   "user1",
						Type: "user",
						Name: "user1",
					},
				},
			}
			total := int64(1)
			accountInfos := []*interfaces.AccountInfo{
				&jobs[0].Creator,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return(jobs, nil)
			ja.EXPECT().GetJobsTotal(gomock.Any(), gomock.Any()).Return(total, nil)
			ums.EXPECT().GetAccountNames(gomock.Any(), accountInfos).Return(errors.New("get account names error"))

			result, resultTotal, err := service.ListJobs(ctx, queryParams)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
			So(resultTotal, ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError)
		})
	})
}

func Test_jobService_ListTasks(t *testing.T) {
	Convey("Test ListTasks\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ja := bmock.NewMockJobAccess(mockCtrl)
		ps := bmock.NewMockPermissionService(mockCtrl)
		db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &jobService{
			appSetting: appSetting,
			db:         db,
			ja:         ja,
			ps:         ps,
		}

		Convey("Success listing tasks\n", func() {
			queryParams := interfaces.TasksQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
				JobID:  "job1",
			}
			tasks := []*interfaces.TaskInfo{
				{
					ID:          "task1",
					JobID:       "job1",
					ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
					ConceptID:   "ot1",
				},
			}
			total := int64(1)

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListTasks(gomock.Any(), gomock.Any()).Return(tasks, nil)
			ja.EXPECT().GetTasksTotal(gomock.Any(), gomock.Any()).Return(total, nil)

			result, resultTotal, err := service.ListTasks(ctx, queryParams)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(resultTotal, ShouldEqual, total)
		})

		Convey("Failed when permission check fails\n", func() {
			queryParams := interfaces.TasksQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(rest.NewHTTPError(ctx, 403, berrors.BknBackend_Job_InternalError))

			result, resultTotal, err := service.ListTasks(ctx, queryParams)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(resultTotal, ShouldEqual, 0)
		})

		Convey("Failed when ListTasks fails\n", func() {
			queryParams := interfaces.TasksQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListTasks(gomock.Any(), gomock.Any()).Return(nil, errors.New("list tasks error"))

			result, resultTotal, err := service.ListTasks(ctx, queryParams)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(resultTotal, ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError)
		})

		Convey("Failed when GetTasksTotal fails\n", func() {
			queryParams := interfaces.TasksQueryParams{
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}
			tasks := []*interfaces.TaskInfo{}

			ps.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			ja.EXPECT().ListTasks(gomock.Any(), gomock.Any()).Return(tasks, nil)
			ja.EXPECT().GetTasksTotal(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("get tasks total error"))

			result, resultTotal, err := service.ListTasks(ctx, queryParams)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(resultTotal, ShouldEqual, 0)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError)
		})
	})
}

func Test_jobService_GetJobsByIDs(t *testing.T) {
	Convey("Test GetJobsByIDs\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ja := bmock.NewMockJobAccess(mockCtrl)
		db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &jobService{
			appSetting: appSetting,
			db:         db,
			ja:         ja,
		}

		Convey("Success getting jobs\n", func() {
			jobIDs := []string{"job1", "job2"}
			jobs := map[string]*interfaces.JobInfo{
				"job1": {
					ID:     "job1",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
				"job2": {
					ID:     "job2",
					KNID:   "kn1",
					Branch: interfaces.MAIN_BRANCH,
				},
			}

			ja.EXPECT().GetJobsByIDs(gomock.Any(), gomock.Any()).Return(jobs, nil)

			result, err := service.GetJobsByIDs(ctx, jobIDs)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
			So(result["job1"], ShouldNotBeNil)
			So(result["job2"], ShouldNotBeNil)
		})

		Convey("Success with empty job IDs\n", func() {
			jobIDs := []string{}
			jobs := map[string]*interfaces.JobInfo{}

			ja.EXPECT().GetJobsByIDs(gomock.Any(), gomock.Any()).Return(jobs, nil)

			result, err := service.GetJobsByIDs(ctx, jobIDs)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Failed when GetJobsByIDs fails\n", func() {
			jobIDs := []string{"job1"}

			ja.EXPECT().GetJobsByIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("get jobs error"))

			result, err := service.GetJobsByIDs(ctx, jobIDs)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError)
		})
	})
}

func Test_jobService_GetJobByID(t *testing.T) {
	Convey("Test GetJobByID\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ja := bmock.NewMockJobAccess(mockCtrl)
		db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

		service := &jobService{
			appSetting: appSetting,
			db:         db,
			ja:         ja,
		}

		Convey("Success getting job\n", func() {
			jobID := "job1"
			job := &interfaces.JobInfo{
				ID:     "job1",
				KNID:   "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}

			ja.EXPECT().GetJobByID(gomock.Any(), gomock.Any()).Return(job, nil)

			result, err := service.GetJobByID(ctx, jobID)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.ID, ShouldEqual, jobID)
		})

		Convey("Failed when GetJobByID fails\n", func() {
			jobID := "job1"

			ja.EXPECT().GetJobByID(gomock.Any(), gomock.Any()).Return(nil, errors.New("get job error"))

			result, err := service.GetJobByID(ctx, jobID)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_Job_InternalError)
		})
	})
}

func Test_jobService_DeleteJobsByKnID(t *testing.T) {
	Convey("Test DeleteJobsByKnID\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		ja := bmock.NewMockJobAccess(mockCtrl)
		service := &jobService{appSetting: &common.AppSetting{}, ja: ja}

		knID := "kn1"
		branch := interfaces.MAIN_BRANCH
		jobIDs := []string{"job1", "job2"}

		Convey("Failed when tx is nil\n", func() {
			err := service.DeleteJobsByKnID(context.Background(), nil, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetJobIDsByKnID access returns error\n", func() {
			tx := new(sql.Tx)
			ja.EXPECT().GetJobIDsByKnID(gomock.Any(), tx, knID, branch).Return(nil, errors.New("get ids error"))
			err := service.DeleteJobsByKnID(context.Background(), tx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteJobsByIDs access returns error\n", func() {
			tx := new(sql.Tx)
			ja.EXPECT().GetJobIDsByKnID(gomock.Any(), tx, knID, branch).Return(jobIDs, nil)
			ja.EXPECT().DeleteJobsByIDs(gomock.Any(), tx, jobIDs).Return(int64(0), errors.New("delete jobs error"))
			err := service.DeleteJobsByKnID(context.Background(), tx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when DeleteTasksByJobIDs access returns error\n", func() {
			tx := new(sql.Tx)
			ja.EXPECT().GetJobIDsByKnID(gomock.Any(), tx, knID, branch).Return(jobIDs, nil)
			ja.EXPECT().DeleteJobsByIDs(gomock.Any(), tx, jobIDs).Return(int64(2), nil)
			ja.EXPECT().DeleteTasksByJobIDs(gomock.Any(), tx, jobIDs).Return(int64(0), errors.New("delete tasks error"))
			err := service.DeleteJobsByKnID(context.Background(), tx, knID, branch)
			So(err, ShouldNotBeNil)
		})

		Convey("Success\n", func() {
			tx := new(sql.Tx)
			ja.EXPECT().GetJobIDsByKnID(gomock.Any(), tx, knID, branch).Return(jobIDs, nil)
			ja.EXPECT().DeleteJobsByIDs(gomock.Any(), tx, jobIDs).Return(int64(2), nil)
			ja.EXPECT().DeleteTasksByJobIDs(gomock.Any(), tx, jobIDs).Return(int64(5), nil)
			err := service.DeleteJobsByKnID(context.Background(), tx, knID, branch)
			So(err, ShouldBeNil)
		})
	})
}
