// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"fmt"
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

func MockNewActionTypeRestHandler(appSetting *common.AppSetting,
	as interfaces.AuthService,
	ats interfaces.ActionTypeService,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		as:         as,
		ats:        ats,
		kns:        kns,
	}
	return r
}

func Test_ActionTypeRestHandler_CreateActionTypes(t *testing.T) {
	Convey("Test ActionTypeHandler CreateActionTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, as, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/action-types"

		actionType := &interfaces.ActionType{
			ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
				ATID:         "at1",
				ATName:       "action1",
				ObjectTypeID: "ot1",
				ActionType:   interfaces.ACTION_TYPE_ADD,
			},
		}
		requestData := struct {
			Entries []*interfaces.ActionType `json:"entries"`
		}{
			Entries: []*interfaces.ActionType{actionType},
		}

		Convey("Success CreateActionTypes \n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CreateActionTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"at1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateActionTypes ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.ActionType{*actionType})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty entries\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			emptyRequestData := struct {
				Entries []*interfaces.ActionType `json:"entries"`
			}{
				Entries: []*interfaces.ActionType{},
			}
			reqParamByte, _ := sonic.Marshal(emptyRequestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("CheckKNExistByID failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ActionType_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, expectedErr)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("CreateActionTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ActionType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CreateActionTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, err)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionTypeRestHandler_UpdateActionType(t *testing.T) {
	Convey("Test ActionTypeHandler UpdateActionType\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, as, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		atID := "at1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/action-types/" + atID

		actionType := interfaces.ActionType{
			ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
				ATID:         atID,
				ATName:       "action1",
				ObjectTypeID: "ot1",
				ActionType:   interfaces.ACTION_TYPE_MODIFY,
			},
		}

		Convey("Success UpdateActionType\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("old_action1", true, nil)
			ats.EXPECT().CheckActionTypeExistByName(gomock.Any(), knID, gomock.Any(), actionType.ATName).Return("", false, nil)
			ats.EXPECT().UpdateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateActionType ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("ActionType not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_ActionTypeRestHandler_DeleteActionTypes(t *testing.T) {
	Convey("Test ActionTypeHandler DeleteActionTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, as, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		atIDs := "at1,at2"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/action-types/" + atIDs

		Convey("Success DeleteActionTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), "at1").Return("action1", true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), "at2").Return("action2", true, nil)
			ats.EXPECT().DeleteActionTypesByIDs(gomock.Any(), gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil)

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

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("ActionType not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), "at1").Return("action1", true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), "at2").Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_ActionTypeRestHandler_ListActionTypes(t *testing.T) {
	Convey("Test ActionTypeHandler ListActionTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, as, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/action-types"

		Convey("Success ListActionTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{}, 0, nil)

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

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

	})
}

func Test_ActionTypeRestHandler_GetActionTypes(t *testing.T) {
	Convey("Test ActionTypeHandler GetActionTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, as, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		atIDs := "at1,at2"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/action-types/" + atIDs

		Convey("Success GetActionTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().GetActionTypesByIDs(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:       "at1",
						ATName:     "action1",
						ActionType: interfaces.ACTION_TYPE_ADD,
					},
				},
				{
					ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
						ATID:       "at2",
						ATName:     "action2",
						ActionType: interfaces.ACTION_TYPE_ADD,
					},
				},
			}, nil)

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

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("GetActionTypesByIDs failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ActionType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().GetActionTypesByIDs(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil, err)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

	})
}

func Test_ActionTypeRestHandler_SearchActionTypes(t *testing.T) {
	Convey("Test ActionTypeHandler SearchActionTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, as, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/action-types"

		query := interfaces.ConceptsQuery{
			Limit: 10,
		}

		Convey("Success SearchActionTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().SearchActionTypes(gomock.Any(), gomock.Any()).Return(interfaces.ActionTypes{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed SearchActionTypes ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("SearchActionTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ActionType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().SearchActionTypes(gomock.Any(), gomock.Any()).Return(interfaces.ActionTypes{}, err)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

	})
}

func Test_ActionTypeRestHandler_HandleActionTypeGetOverride(t *testing.T) {
	Convey("Test ActionTypeHandler HandleActionTypeGetOverrideByEx\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, as, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		urlEx := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/action-types"

		actionType := &interfaces.ActionType{
			ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
				ATID:         "at1",
				ATName:       "action1",
				ObjectTypeID: "ot1",
				ActionType:   interfaces.ACTION_TYPE_ADD,
			},
		}
		requestData := struct {
			Entries []*interfaces.ActionType `json:"entries"`
		}{
			Entries: []*interfaces.ActionType{actionType},
		}

		Convey("HandleActionTypeGetOverrideByEx - Success with POST method (default)\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CreateActionTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"at1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, urlEx, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("HandleActionTypeGetOverrideByEx - Success with GET override method\n", func() {
			query := interfaces.ConceptsQuery{
				Limit: 10,
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().SearchActionTypes(gomock.Any(), gomock.Any()).Return(interfaces.ActionTypes{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, urlEx, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("HandleActionTypeGetOverrideByEx - Failed with invalid override method\n", func() {
			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, urlEx, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "PUT")
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func newActionTypeTestHandler(t *testing.T) (*restHandler, *gomock.Controller, *gin.Engine, *bmock.MockActionTypeService, *bmock.MockKNService) {
	t.Helper()
	mockCtrl := gomock.NewController(t)
	engine := gin.New()
	engine.Use(gin.Recovery())
	appSetting := &common.AppSetting{}
	as := bmock.NewMockAuthService(mockCtrl)
	ats := bmock.NewMockActionTypeService(mockCtrl)
	kns := bmock.NewMockKNService(mockCtrl)
	handler := MockNewActionTypeRestHandler(appSetting, as, ats, kns)
	handler.RegisterPublic(engine)
	return handler, mockCtrl, engine, ats, kns
}

func Test_ActionTypeRestHandler_CreateActionTypesByIn(t *testing.T) {
	Convey("Test ActionTypeHandler CreateActionTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, ats, kns := newActionTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CreateActionTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"at1"}, nil)

			actionType := &interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         "at1",
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_ADD,
				},
			}
			requestData := struct {
				Entries []*interfaces.ActionType `json:"entries"`
			}{
				Entries: []*interfaces.ActionType{actionType},
			}
			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/action-types", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_ActionTypeRestHandler_UpdateActionTypeByIn(t *testing.T) {
	Convey("Test ActionTypeHandler UpdateActionTypeByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, ats, kns := newActionTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		atID := "at1"

		Convey("Success\n", func() {
			actionType := interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         atID,
					ATName:       "action1",
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_MODIFY,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("old_action1", true, nil)
			ats.EXPECT().CheckActionTypeExistByName(gomock.Any(), knID, gomock.Any(), actionType.ATName).Return("", false, nil)
			ats.EXPECT().UpdateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/action-types/"+atID, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_ActionTypeRestHandler_ListActionTypesByIn(t *testing.T) {
	Convey("Test ActionTypeHandler ListActionTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, ats, kns := newActionTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{}, 0, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/action-types", nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_ActionTypeRestHandler_GetActionTypesByIn(t *testing.T) {
	Convey("Test ActionTypeHandler GetActionTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, ats, kns := newActionTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		atIDs := "at1,at2"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().GetActionTypesByIDs(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{}, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/action-types/"+atIDs, nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_ActionTypeRestHandler_SearchActionTypesByIn(t *testing.T) {
	Convey("Test ActionTypeHandler SearchActionTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, ats, kns := newActionTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().SearchActionTypes(gomock.Any(), gomock.Any()).Return(interfaces.ActionTypes{}, nil)

			query := interfaces.ConceptsQuery{Limit: 10}
			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/action-types", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_ActionTypeRestHandler_HandleActionTypeGetOverrideByIn(t *testing.T) {
	Convey("Test ActionTypeHandler HandleActionTypeGetOverrideByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, ats, kns := newActionTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		urlIn := "/api/bkn-backend/in/v1/knowledge-networks/" + knID + "/action-types"

		actionType := &interfaces.ActionType{
			ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
				ATID:         "at1",
				ATName:       "action1",
				ObjectTypeID: "ot1",
				ActionType:   interfaces.ACTION_TYPE_ADD,
			},
		}
		requestData := struct {
			Entries []*interfaces.ActionType `json:"entries"`
		}{
			Entries: []*interfaces.ActionType{actionType},
		}

		Convey("HandleActionTypeGetOverrideByIn - Success with POST method (default)\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CreateActionTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"at1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, urlIn, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("HandleActionTypeGetOverrideByIn - Success with GET override method\n", func() {
			query := interfaces.ConceptsQuery{
				Limit: 10,
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().SearchActionTypes(gomock.Any(), gomock.Any()).Return(interfaces.ActionTypes{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, urlIn, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("HandleActionTypeGetOverrideByIn - Failed with invalid override method\n", func() {
			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, urlIn, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "PUT")
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func Test_ActionTypeRestHandler_UpdateActionType_extraCases(t *testing.T) {
	Convey("Test ActionTypeHandler UpdateActionType extra cases\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, as, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		atID := "at1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/action-types/" + atID

		actionType := interfaces.ActionType{
			ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
				ATID:         atID,
				ATName:       "action1",
				ObjectTypeID: "ot1",
				ActionType:   interfaces.ACTION_TYPE_MODIFY,
			},
		}

		Convey("Failed when CheckKNExistByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionType_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, httpErr)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when CheckActionTypeExistByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionType_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("", false, httpErr)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed ValidateActionType with empty ATName\n", func() {
			atNoName := interfaces.ActionType{
				ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
					ATID:         atID,
					ObjectTypeID: "ot1",
					ActionType:   interfaces.ACTION_TYPE_MODIFY,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("old_action1", true, nil)

			reqParamByte, _ := sonic.Marshal(atNoName)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when CheckActionTypeExistByName returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionType_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("old_name", true, nil)
			ats.EXPECT().CheckActionTypeExistByName(gomock.Any(), knID, gomock.Any(), actionType.ATName).Return("", false, httpErr)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when action type name already exists\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("old_name", true, nil)
			ats.EXPECT().CheckActionTypeExistByName(gomock.Any(), knID, gomock.Any(), actionType.ATName).Return(atID, true, nil)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Failed when UpdateActionType service returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionType_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), atID).Return("old_name", true, nil)
			ats.EXPECT().CheckActionTypeExistByName(gomock.Any(), knID, gomock.Any(), actionType.ATName).Return("", false, nil)
			ats.EXPECT().UpdateActionType(gomock.Any(), gomock.Any(), gomock.Any()).Return(httpErr)

			reqParamByte, _ := sonic.Marshal(actionType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ActionTypeRestHandler_DeleteActionTypes_extraCases(t *testing.T) {
	Convey("Test ActionTypeHandler DeleteActionTypes extra cases\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewActionTypeRestHandler(appSetting, as, ats, kns)
		handler.RegisterPublic(engine)

		knID := "kn1"
		atIDs := "at1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/action-types/" + atIDs

		Convey("Failed when VerifyToken returns error\n", func() {
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(hydra.Visitor{}, fmt.Errorf("token invalid"))

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Failed when CheckKNExistByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionType_InternalError},
			}
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(hydra.Visitor{}, nil)
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, httpErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when CheckActionTypeExistByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionType_InternalError},
			}
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(hydra.Visitor{}, nil)
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), "at1").Return("", false, httpErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when DeleteActionTypesByIDs service returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ActionType_InternalError},
			}
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(hydra.Visitor{}, nil)
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ats.EXPECT().CheckActionTypeExistByID(gomock.Any(), knID, gomock.Any(), "at1").Return("action1", true, nil)
			ats.EXPECT().DeleteActionTypesByIDs(gomock.Any(), gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(httpErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
