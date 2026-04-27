// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

func MockNewJobRestHandler(appSetting *common.AppSetting,
	as interfaces.AuthService,
	js interfaces.JobService,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		as:         as,
		js:         js,
		kns:        kns,
	}
	return r
}

func Test_JobRestHandler_CreateJob(t *testing.T) {
	Convey("Test JobHandler CreateJob\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		js := bmock.NewMockJobService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewJobRestHandler(appSetting, as, js, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/jobs"

		jobInfo := interfaces.JobInfo{
			Name:    "job1",
			Branch:  interfaces.MAIN_BRANCH,
			JobType: interfaces.JobTypeFull,
		}

		Convey("Success CreateJob \n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().CreateJob(gomock.Any(), gomock.Any()).Return("job1", nil)

			reqParamByte, _ := sonic.Marshal(jobInfo)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateJob ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.JobInfo{jobInfo})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Job name is null\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			reqParamByte, _ := sonic.Marshal(interfaces.JobInfo{})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(jobInfo)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("CheckKNExistByID failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_Job_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, expectedErr)

			reqParamByte, _ := sonic.Marshal(jobInfo)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("CreateJob failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_Job_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().CreateJob(gomock.Any(), gomock.Any()).Return("", err)

			reqParamByte, _ := sonic.Marshal(jobInfo)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_JobRestHandler_DeleteJobs(t *testing.T) {
	Convey("Test JobHandler DeleteJobs\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		js := bmock.NewMockJobService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewJobRestHandler(appSetting, as, js, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		jobIDs := "job1,job2"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/jobs/" + jobIDs

		Convey("Success DeleteJobs\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobsByIDs(gomock.Any(), gomock.Any()).Return(map[string]*interfaces.JobInfo{
				"job1": {ID: "job1", Name: "job1", KNID: knID, Branch: interfaces.MAIN_BRANCH},
				"job2": {ID: "job2", Name: "job2", KNID: knID, Branch: interfaces.MAIN_BRANCH},
			}, nil)
			js.EXPECT().DeleteJobsByIDs(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Job not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobsByIDs(gomock.Any(), gomock.Any()).Return(map[string]*interfaces.JobInfo{
				"job1": {ID: "job1", Name: "job1", KNID: knID, Branch: interfaces.MAIN_BRANCH},
			}, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func Test_JobRestHandler_ListJobs(t *testing.T) {
	Convey("Test JobHandler ListJobs\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		js := bmock.NewMockJobService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewJobRestHandler(appSetting, as, js, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/jobs"

		Convey("Success ListJobs\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Invalid pagination parameters\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?offset=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid job_type\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?job_type=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid state\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?state=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ListJobs failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_Job_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return(nil, int64(0), expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_JobRestHandler_ListTasks(t *testing.T) {
	Convey("Test JobHandler ListTasks\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		js := bmock.NewMockJobService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewJobRestHandler(appSetting, as, js, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		jobID := "job1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/jobs/" + jobID + "/tasks"

		Convey("Success ListTasks\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(&interfaces.JobInfo{
				ID:     jobID,
				KNID:   knID,
				Branch: interfaces.MAIN_BRANCH,
			}, nil)
			js.EXPECT().ListTasks(gomock.Any(), gomock.Any()).Return([]*interfaces.TaskInfo{}, int64(0), nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Job not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(nil, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("GetJobByID failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_Job_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(nil, expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Invalid pagination parameters\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(&interfaces.JobInfo{
				ID:     jobID,
				KNID:   knID,
				Branch: interfaces.MAIN_BRANCH,
			}, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?offset=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid concept_type\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(&interfaces.JobInfo{
				ID:     jobID,
				KNID:   knID,
				Branch: interfaces.MAIN_BRANCH,
			}, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?concept_type=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid state\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(&interfaces.JobInfo{
				ID:     jobID,
				KNID:   knID,
				Branch: interfaces.MAIN_BRANCH,
			}, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?state=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ListTasks failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_Job_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(&interfaces.JobInfo{
				ID:     jobID,
				KNID:   knID,
				Branch: interfaces.MAIN_BRANCH,
			}, nil)
			js.EXPECT().ListTasks(gomock.Any(), gomock.Any()).Return(nil, int64(0), expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func newJobTestHandler(t *testing.T) (*restHandler, *gomock.Controller, *gin.Engine, *bmock.MockJobService, *bmock.MockKNService) {
	t.Helper()
	mockCtrl := gomock.NewController(t)
	engine := gin.New()
	engine.Use(gin.Recovery())
	appSetting := &common.AppSetting{}
	as := bmock.NewMockAuthService(mockCtrl)
	js := bmock.NewMockJobService(mockCtrl)
	kns := bmock.NewMockKNService(mockCtrl)
	handler := MockNewJobRestHandler(appSetting, as, js, kns)
	handler.RegisterPublic(engine)
	return handler, mockCtrl, engine, js, kns
}

func Test_JobRestHandler_CreateJobByIn(t *testing.T) {
	Convey("Test JobHandler CreateJobByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, js, kns := newJobTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().CreateJob(gomock.Any(), gomock.Any()).Return("job1", nil)

			jobInfo := interfaces.JobInfo{Name: "job1", Branch: interfaces.MAIN_BRANCH, JobType: interfaces.JobTypeFull}
			reqParamByte, _ := sonic.Marshal(jobInfo)
			req := httptest.NewRequest(http.MethodPost, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/jobs", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_JobRestHandler_DeleteJobsByIn(t *testing.T) {
	Convey("Test JobHandler DeleteJobsByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, js, kns := newJobTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		jobIDs := "job1,job2"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobsByIDs(gomock.Any(), gomock.Any()).Return(map[string]*interfaces.JobInfo{
				"job1": {ID: "job1", Name: "job1", KNID: knID, Branch: interfaces.MAIN_BRANCH},
				"job2": {ID: "job2", Name: "job2", KNID: knID, Branch: interfaces.MAIN_BRANCH},
			}, nil)
			js.EXPECT().DeleteJobsByIDs(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodDelete, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/jobs/"+jobIDs, nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_JobRestHandler_ListJobsByIn(t *testing.T) {
	Convey("Test JobHandler ListJobsByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, js, kns := newJobTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().ListJobs(gomock.Any(), gomock.Any()).Return([]*interfaces.JobInfo{}, int64(0), nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/jobs", nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_JobRestHandler_ListTasksByIn(t *testing.T) {
	Convey("Test JobHandler ListTasksByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, js, kns := newJobTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		jobID := "job1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			js.EXPECT().GetJobByID(gomock.Any(), jobID).Return(&interfaces.JobInfo{
				ID: jobID, KNID: knID, Branch: interfaces.MAIN_BRANCH,
			}, nil)
			js.EXPECT().ListTasks(gomock.Any(), gomock.Any()).Return([]*interfaces.TaskInfo{}, int64(0), nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/jobs/"+jobID+"/tasks", nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
