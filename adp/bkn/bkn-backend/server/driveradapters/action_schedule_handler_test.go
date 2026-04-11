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

func MockNewActionScheduleRestHandler(
	appSetting *common.AppSetting,
	as interfaces.AuthService,
	kns interfaces.KNService,
	ass interfaces.ActionScheduleService,
) *restHandler {
	return &restHandler{
		appSetting: appSetting,
		as:         as,
		kns:        kns,
		ass:        ass,
	}
}

func Test_ActionScheduleRestHandler_CreateActionSchedule(t *testing.T) {
	Convey("Test ActionScheduleHandler CreateActionSchedule\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)
		ass := bmock.NewMockActionScheduleService(mockCtrl)

		handler := MockNewActionScheduleRestHandler(appSetting, as, kns, ass)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/knowledge-networks/kn1/action-schedules"

		validReq := interfaces.ActionScheduleCreateRequest{
			Name:               "test-schedule",
			ActionTypeID:       "at1",
			CronExpression:     "0 * * * *",
			Status:             interfaces.ScheduleStatusInactive,
			InstanceIdentities: []map[string]any{{"id": "instance1"}},
		}

		Convey("Success creating action schedule\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH).Return("kn1", true, nil)
			ass.EXPECT().CreateSchedule(gomock.Any(), gomock.Any()).Return("sched1", nil)

			body, _ := sonic.Marshal(validReq)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed when KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH).Return("", false, nil)

			body, _ := sonic.Marshal(validReq)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Failed when KN check returns error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH).Return("", false, &rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionSchedule_InternalError},
			})

			body, _ := sonic.Marshal(validReq)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when ShouldBind error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH).Return("kn1", true, nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("invalid-json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when CreateSchedule returns error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH).Return("kn1", true, nil)
			ass.EXPECT().CreateSchedule(gomock.Any(), gomock.Any()).Return("", &rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionSchedule_InternalError},
			})

			body, _ := sonic.Marshal(validReq)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionScheduleRestHandler_UpdateActionSchedule(t *testing.T) {
	Convey("Test ActionScheduleHandler UpdateActionSchedule\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)
		ass := bmock.NewMockActionScheduleService(mockCtrl)

		handler := MockNewActionScheduleRestHandler(appSetting, as, kns, ass)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/knowledge-networks/kn1/action-schedules/sched1"

		existingSchedule := &interfaces.ActionSchedule{
			ID:     "sched1",
			Name:   "test-schedule",
			KNID:   "kn1",
			Branch: interfaces.MAIN_BRANCH,
		}

		validReq := interfaces.ActionScheduleUpdateRequest{
			Name:           "updated-schedule",
			CronExpression: "0 0 * * *",
		}

		Convey("Success updating action schedule\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), "sched1").Return(existingSchedule, nil)
			ass.EXPECT().UpdateSchedule(gomock.Any(), "sched1", gomock.Any()).Return(nil)

			body, _ := sonic.Marshal(validReq)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed when schedule not found\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), "sched1").Return(nil, &rest.HTTPError{
				HTTPCode:  http.StatusNotFound,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionSchedule_NotFound},
			})

			body, _ := sonic.Marshal(validReq)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Failed when schedule belongs to different KN\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), "sched1").Return(&interfaces.ActionSchedule{
				ID: "sched1", KNID: "other-kn", Branch: interfaces.MAIN_BRANCH,
			}, nil)

			body, _ := sonic.Marshal(validReq)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Failed when UpdateSchedule returns error\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), "sched1").Return(existingSchedule, nil)
			ass.EXPECT().UpdateSchedule(gomock.Any(), "sched1", gomock.Any()).Return(&rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionSchedule_InternalError},
			})

			body, _ := sonic.Marshal(validReq)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionScheduleRestHandler_UpdateActionScheduleStatus(t *testing.T) {
	Convey("Test ActionScheduleHandler UpdateActionScheduleStatus\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)
		ass := bmock.NewMockActionScheduleService(mockCtrl)

		handler := MockNewActionScheduleRestHandler(appSetting, as, kns, ass)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/knowledge-networks/kn1/action-schedules/sched1/status"

		existingSchedule := &interfaces.ActionSchedule{
			ID: "sched1", Name: "test-schedule", KNID: "kn1", Branch: interfaces.MAIN_BRANCH,
		}

		Convey("Success activating schedule\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), "sched1").Return(existingSchedule, nil)
			ass.EXPECT().UpdateScheduleStatus(gomock.Any(), "sched1", interfaces.ScheduleStatusActive).Return(nil)

			body, _ := sonic.Marshal(interfaces.ActionScheduleStatusRequest{Status: interfaces.ScheduleStatusActive})
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed when schedule belongs to different KN\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), "sched1").Return(&interfaces.ActionSchedule{
				ID: "sched1", KNID: "other-kn", Branch: interfaces.MAIN_BRANCH,
			}, nil)

			body, _ := sonic.Marshal(interfaces.ActionScheduleStatusRequest{Status: interfaces.ScheduleStatusActive})
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Failed when UpdateScheduleStatus returns error\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), "sched1").Return(existingSchedule, nil)
			ass.EXPECT().UpdateScheduleStatus(gomock.Any(), "sched1", gomock.Any()).Return(&rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionSchedule_InternalError},
			})

			body, _ := sonic.Marshal(interfaces.ActionScheduleStatusRequest{Status: interfaces.ScheduleStatusActive})
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionScheduleRestHandler_DeleteActionSchedules(t *testing.T) {
	Convey("Test ActionScheduleHandler DeleteActionSchedules\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)
		ass := bmock.NewMockActionScheduleService(mockCtrl)

		handler := MockNewActionScheduleRestHandler(appSetting, as, kns, ass)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/knowledge-networks/kn1/action-schedules/sched1"

		Convey("Success deleting schedules\n", func() {
			ass.EXPECT().GetSchedules(gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ActionSchedule{
				"sched1": {ID: "sched1", Name: "test-schedule"},
			}, nil)
			ass.EXPECT().DeleteSchedules(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed when GetSchedules returns error\n", func() {
			ass.EXPECT().GetSchedules(gomock.Any(), gomock.Any()).Return(nil, &rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionSchedule_InternalError},
			})

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when DeleteSchedules returns error\n", func() {
			ass.EXPECT().GetSchedules(gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ActionSchedule{}, nil)
			ass.EXPECT().DeleteSchedules(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, gomock.Any()).Return(&rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionSchedule_InternalError},
			})

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionScheduleRestHandler_ListActionSchedules(t *testing.T) {
	Convey("Test ActionScheduleHandler ListActionSchedules\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)
		ass := bmock.NewMockActionScheduleService(mockCtrl)

		handler := MockNewActionScheduleRestHandler(appSetting, as, kns, ass)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/knowledge-networks/kn1/action-schedules"

		Convey("Success listing schedules\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH).Return("kn1", true, nil)
			ass.EXPECT().ListSchedules(gomock.Any(), gomock.Any()).Return([]*interfaces.ActionSchedule{}, int64(0), nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed when KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH).Return("", false, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Failed with invalid status filter\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH).Return("kn1", true, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?status=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when ListSchedules returns error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH).Return("kn1", true, nil)
			ass.EXPECT().ListSchedules(gomock.Any(), gomock.Any()).Return(nil, int64(0), &rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionSchedule_InternalError},
			})

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionScheduleRestHandler_GetActionSchedule(t *testing.T) {
	Convey("Test ActionScheduleHandler GetActionSchedule\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)
		ass := bmock.NewMockActionScheduleService(mockCtrl)

		handler := MockNewActionScheduleRestHandler(appSetting, as, kns, ass)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/knowledge-networks/kn1/action-schedules/sched1"

		Convey("Success getting schedule\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), "sched1").Return(&interfaces.ActionSchedule{
				ID: "sched1", KNID: "kn1", Branch: interfaces.MAIN_BRANCH,
			}, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed when schedule not found\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), "sched1").Return(nil, &rest.HTTPError{
				HTTPCode:  http.StatusNotFound,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionSchedule_NotFound},
			})

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Failed when schedule belongs to different KN\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), "sched1").Return(&interfaces.ActionSchedule{
				ID: "sched1", KNID: "other-kn", Branch: interfaces.MAIN_BRANCH,
			}, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func newActionScheduleTestHandler(t *testing.T) (*restHandler, *gomock.Controller, *gin.Engine, *bmock.MockKNService, *bmock.MockActionScheduleService) {
	t.Helper()
	mockCtrl := gomock.NewController(t)
	engine := gin.New()
	engine.Use(gin.Recovery())
	appSetting := &common.AppSetting{}
	as := bmock.NewMockAuthService(mockCtrl)
	kns := bmock.NewMockKNService(mockCtrl)
	ass := bmock.NewMockActionScheduleService(mockCtrl)
	handler := MockNewActionScheduleRestHandler(appSetting, as, kns, ass)
	handler.RegisterPublic(engine)
	return handler, mockCtrl, engine, kns, ass
}

func Test_ActionScheduleRestHandler_CreateActionScheduleByIn(t *testing.T) {
	Convey("Test ActionScheduleHandler CreateActionScheduleByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, kns, ass := newActionScheduleTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		validReq := interfaces.ActionScheduleCreateRequest{
			Name:               "test-schedule",
			ActionTypeID:       "at1",
			CronExpression:     "0 * * * *",
			Status:             interfaces.ScheduleStatusInactive,
			InstanceIdentities: []map[string]any{{"id": "instance1"}},
		}

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, interfaces.MAIN_BRANCH).Return(knID, true, nil)
			ass.EXPECT().CreateSchedule(gomock.Any(), gomock.Any()).Return("sched1", nil)

			body, _ := sonic.Marshal(validReq)
			req := httptest.NewRequest(http.MethodPost, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/action-schedules", bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_ActionScheduleRestHandler_UpdateActionScheduleByIn(t *testing.T) {
	Convey("Test ActionScheduleHandler UpdateActionScheduleByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, _, ass := newActionScheduleTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		schedID := "sched1"

		Convey("Success\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), schedID).Return(&interfaces.ActionSchedule{
				ID: schedID, Name: "test-schedule", KNID: knID, Branch: interfaces.MAIN_BRANCH,
			}, nil)
			ass.EXPECT().UpdateSchedule(gomock.Any(), schedID, gomock.Any()).Return(nil)

			body, _ := sonic.Marshal(interfaces.ActionScheduleUpdateRequest{Name: "updated", CronExpression: "0 0 * * *"})
			req := httptest.NewRequest(http.MethodPut, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/action-schedules/"+schedID, bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_ActionScheduleRestHandler_UpdateActionScheduleStatusByIn(t *testing.T) {
	Convey("Test ActionScheduleHandler UpdateActionScheduleStatusByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, _, ass := newActionScheduleTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		schedID := "sched1"

		Convey("Success\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), schedID).Return(&interfaces.ActionSchedule{
				ID: schedID, Name: "test-schedule", KNID: knID, Branch: interfaces.MAIN_BRANCH,
			}, nil)
			ass.EXPECT().UpdateScheduleStatus(gomock.Any(), schedID, interfaces.ScheduleStatusActive).Return(nil)

			body, _ := sonic.Marshal(interfaces.ActionScheduleStatusRequest{Status: interfaces.ScheduleStatusActive})
			req := httptest.NewRequest(http.MethodPut, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/action-schedules/"+schedID+"/status", bytes.NewReader(body))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_ActionScheduleRestHandler_DeleteActionSchedulesByIn(t *testing.T) {
	Convey("Test ActionScheduleHandler DeleteActionSchedulesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, _, ass := newActionScheduleTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		schedID := "sched1"

		Convey("Success\n", func() {
			ass.EXPECT().GetSchedules(gomock.Any(), gomock.Any()).Return(map[string]*interfaces.ActionSchedule{
				schedID: {ID: schedID, Name: "test-schedule"},
			}, nil)
			ass.EXPECT().DeleteSchedules(gomock.Any(), knID, interfaces.MAIN_BRANCH, gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodDelete, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/action-schedules/"+schedID, nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_ActionScheduleRestHandler_ListActionSchedulesByIn(t *testing.T) {
	Convey("Test ActionScheduleHandler ListActionSchedulesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, kns, ass := newActionScheduleTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, interfaces.MAIN_BRANCH).Return(knID, true, nil)
			ass.EXPECT().ListSchedules(gomock.Any(), gomock.Any()).Return([]*interfaces.ActionSchedule{}, int64(0), nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/action-schedules", nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_ActionScheduleRestHandler_GetActionScheduleByIn(t *testing.T) {
	Convey("Test ActionScheduleHandler GetActionScheduleByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, _, ass := newActionScheduleTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		schedID := "sched1"

		Convey("Success\n", func() {
			ass.EXPECT().GetSchedule(gomock.Any(), schedID).Return(&interfaces.ActionSchedule{
				ID: schedID, KNID: knID, Branch: interfaces.MAIN_BRANCH,
			}, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/action-schedules/"+schedID, nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
