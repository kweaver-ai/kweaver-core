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

func MockNewConceptGroupRestHandler(appSetting *common.AppSetting,
	as interfaces.AuthService,
	cgs interfaces.ConceptGroupService,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		as:         as,
		cgs:        cgs,
		kns:        kns,
	}
	return r
}

func Test_ConceptGroupRestHandler_CreateConceptGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler CreateConceptGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups"

		conceptGroup := interfaces.ConceptGroup{
			CGName: "group1",
			CommonInfo: interfaces.CommonInfo{
				Comment: "test comment",
			},
		}

		Convey("Success CreateConceptGroup \n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("cg1", nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateConceptGroup ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.ConceptGroup{conceptGroup})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("CG name is null\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			reqParamByte, _ := sonic.Marshal(interfaces.ConceptGroup{})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
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
					ErrorCode: berrors.BknBackend_ConceptGroup_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, expectedErr)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

	})
}

func Test_ConceptGroupRestHandler_UpdateConceptGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler UpdateConceptGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID

		conceptGroup := interfaces.ConceptGroup{
			CGID:   cgID,
			CGName: "group1",
		}

		Convey("Success UpdateConceptGroup\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group2", true, nil)
			cgs.EXPECT().CheckConceptGroupExistByName(gomock.Any(), knID, gomock.Any(), conceptGroup.CGName).Return("", false, nil)
			cgs.EXPECT().UpdateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateConceptGroup ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("ConceptGroup not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

	})
}

func Test_ConceptGroupRestHandler_DeleteConceptGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler DeleteConceptGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID

		Convey("Success DeleteConceptGroup\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().DeleteConceptGroupByID(gomock.Any(), gomock.Any(), knID, gomock.Any(), cgID).Return(nil)

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

		Convey("ConceptGroup not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_ConceptGroupRestHandler_ListConceptGroups(t *testing.T) {
	Convey("Test ConceptGroupHandler ListConceptGroups\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups"

		Convey("Success ListConceptGroups\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return([]*interfaces.ConceptGroup{}, 0, nil)

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

func Test_ConceptGroupRestHandler_GetConceptGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler GetConceptGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID

		Convey("Success GetConceptGroup\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().GetConceptGroupByID(gomock.Any(), knID, gomock.Any(), cgID, gomock.Any()).Return(&interfaces.ConceptGroup{}, nil)

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

		Convey("Invalid mode\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?mode=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid include_statistics\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?include_statistics=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetConceptGroupByID failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ConceptGroup_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().GetConceptGroupByID(gomock.Any(), knID, gomock.Any(), cgID, gomock.Any()).Return(nil, expectedErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("GetStatByConceptGroup failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ConceptGroup_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().GetConceptGroupByID(gomock.Any(), knID, gomock.Any(), cgID, gomock.Any()).Return(&interfaces.ConceptGroup{}, nil)
			cgs.EXPECT().GetStatByConceptGroup(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			req := httptest.NewRequest(http.MethodGet, url+"?include_statistics=true", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Success with include_statistics\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().GetConceptGroupByID(gomock.Any(), knID, gomock.Any(), cgID, gomock.Any()).Return(&interfaces.ConceptGroup{}, nil)
			cgs.EXPECT().GetStatByConceptGroup(gomock.Any(), gomock.Any()).Return(&interfaces.Statistics{}, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?include_statistics=true", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

	})
}

func Test_ConceptGroupRestHandler_AddObjectTypesToConceptGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler AddObjectTypesToConceptGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID + "/object-types"

		requestData := struct {
			Entries []interfaces.ID `json:"entries"`
		}{
			Entries: []interfaces.ID{{ID: "ot1"}},
		}

		Convey("Success AddObjectTypesToConceptGroup\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().AddObjectTypesToConceptGroup(gomock.Any(), gomock.Any(), knID, gomock.Any(), cgID, gomock.Any(), gomock.Any()).Return([]string{"ot1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
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

		Convey("ConceptGroup not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("CheckConceptGroupExistByID failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ConceptGroup_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("", false, expectedErr)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("ShouldBindJSON failed\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("AddObjectTypesToConceptGroup failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ConceptGroup_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().AddObjectTypesToConceptGroup(gomock.Any(), gomock.Any(), knID, gomock.Any(), cgID, gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

	})
}

func Test_ConceptGroupRestHandler_DeleteObjectTypesFromGroup(t *testing.T) {
	Convey("Test ConceptGroupHandler DeleteObjectTypesFromGroup\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		otIDs := "ot1,ot2"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID + "/object-types/" + otIDs

		Convey("Success DeleteObjectTypesFromGroup\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, query interface{}) ([]interfaces.ConceptGroupRelation, error) {
				return []interfaces.ConceptGroupRelation{
					{
						ID:          "rel1",
						KNID:        knID,
						Branch:      interfaces.MAIN_BRANCH,
						CGID:        cgID,
						ConceptID:   "ot1",
						ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
					},
					{
						ID:          "rel2",
						KNID:        knID,
						Branch:      interfaces.MAIN_BRANCH,
						CGID:        cgID,
						ConceptID:   "ot2",
						ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
					},
				}, nil
			})
			cgs.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), knID, gomock.Any(), cgID, gomock.Any()).Return(nil)

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

		Convey("ConceptGroup not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("CheckConceptGroupExistByID failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ConceptGroup_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("", false, expectedErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("DeleteObjectTypesFromGroup failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ConceptGroup_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, query interface{}) ([]interfaces.ConceptGroupRelation, error) {
				return []interfaces.ConceptGroupRelation{
					{
						ID:          "rel1",
						KNID:        knID,
						Branch:      interfaces.MAIN_BRANCH,
						CGID:        cgID,
						ConceptID:   "ot1",
						ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
					},
					{
						ID:          "rel2",
						KNID:        knID,
						Branch:      interfaces.MAIN_BRANCH,
						CGID:        cgID,
						ConceptID:   "ot2",
						ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
					},
				}, nil
			})
			cgs.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), knID, gomock.Any(), cgID, gomock.Any()).Return(expectedErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("ListConceptGroupRelations failed\n", func() {
			expectedErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ConceptGroup_InternalError,
				},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("ConceptGroupRelation not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, query interface{}) ([]interfaces.ConceptGroupRelation, error) {
				return []interfaces.ConceptGroupRelation{
					{
						CGID:        cgID,
						ConceptID:   "ot1",
						ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
					},
				}, nil
			})

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

	})
}

func newConceptGroupTestHandler(t *testing.T) (*restHandler, *gomock.Controller, *gin.Engine, *bmock.MockConceptGroupService, *bmock.MockKNService) {
	t.Helper()
	mockCtrl := gomock.NewController(t)
	engine := gin.New()
	engine.Use(gin.Recovery())
	appSetting := &common.AppSetting{}
	as := bmock.NewMockAuthService(mockCtrl)
	cgs := bmock.NewMockConceptGroupService(mockCtrl)
	kns := bmock.NewMockKNService(mockCtrl)
	handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
	handler.RegisterPublic(engine)
	return handler, mockCtrl, engine, cgs, kns
}

func Test_ConceptGroupRestHandler_CreateConceptGroupByIn(t *testing.T) {
	Convey("Test ConceptGroupHandler CreateConceptGroupByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, cgs, kns := newConceptGroupTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		conceptGroup := interfaces.ConceptGroup{
			CGName: "group1",
			CommonInfo: interfaces.CommonInfo{
				Comment: "test comment",
			},
		}

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("cg1", nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPost, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/concept-groups", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_ConceptGroupRestHandler_UpdateConceptGroupByIn(t *testing.T) {
	Convey("Test ConceptGroupHandler UpdateConceptGroupByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, cgs, kns := newConceptGroupTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		cgID := "cg1"
		conceptGroup := interfaces.ConceptGroup{
			CGID:   cgID,
			CGName: "group1",
		}

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("old_group1", true, nil)
			cgs.EXPECT().CheckConceptGroupExistByName(gomock.Any(), knID, gomock.Any(), conceptGroup.CGName).Return("", false, nil)
			cgs.EXPECT().UpdateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/concept-groups/"+cgID, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_ConceptGroupRestHandler_ListConceptGroupsByIn(t *testing.T) {
	Convey("Test ConceptGroupHandler ListConceptGroupsByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, cgs, kns := newConceptGroupTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return([]*interfaces.ConceptGroup{}, 0, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/concept-groups", nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_ConceptGroupRestHandler_GetConceptGroupByIn(t *testing.T) {
	Convey("Test ConceptGroupHandler GetConceptGroupByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, cgs, kns := newConceptGroupTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		cgID := "cg1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().GetConceptGroupByID(gomock.Any(), knID, gomock.Any(), cgID, gomock.Any()).Return(&interfaces.ConceptGroup{}, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/concept-groups/"+cgID, nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_ConceptGroupRestHandler_AddObjectTypesToConceptGroupByIn(t *testing.T) {
	Convey("Test ConceptGroupHandler AddObjectTypesToConceptGroupByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, cgs, kns := newConceptGroupTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		cgID := "cg1"
		requestData := struct {
			Entries []interfaces.ID `json:"entries"`
		}{
			Entries: []interfaces.ID{{ID: "ot1"}},
		}

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().AddObjectTypesToConceptGroup(gomock.Any(), gomock.Any(), knID, gomock.Any(), cgID, gomock.Any(), gomock.Any()).Return([]string{"ot1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/concept-groups/"+cgID+"/object-types", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_ConceptGroupRestHandler_DeleteObjectTypesFromGroupByIn(t *testing.T) {
	Convey("Test ConceptGroupHandler DeleteObjectTypesFromGroupByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, cgs, kns := newConceptGroupTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		cgID := "cg1"
		otIDs := "ot1,ot2"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().ListConceptGroupRelations(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, query interface{}) ([]interfaces.ConceptGroupRelation, error) {
				return []interfaces.ConceptGroupRelation{
					{
						CGID:        cgID,
						ConceptID:   "ot1",
						ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
					},
					{
						CGID:        cgID,
						ConceptID:   "ot2",
						ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
					},
				}, nil
			})
			cgs.EXPECT().DeleteObjectTypesFromGroup(gomock.Any(), gomock.Any(), knID, gomock.Any(), cgID, gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodDelete, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/concept-groups/"+cgID+"/object-types/"+otIDs, nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_ConceptGroupRestHandler_CreateConceptGroup_extraCases(t *testing.T) {
	Convey("Test ConceptGroupHandler CreateConceptGroup extra cases\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups"

		conceptGroup := interfaces.ConceptGroup{
			CGName: "group1",
			CommonInfo: interfaces.CommonInfo{
				Comment: "test comment",
			},
		}

		Convey("Failed when validate_dependency is invalid\n", func() {
			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPost, url+"?validate_dependency=notbool", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when module_type is wrong\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cg := interfaces.ConceptGroup{
				CGName:     "group1",
				ModuleType: "wrong_module",
				CommonInfo: interfaces.CommonInfo{Comment: "test"},
			}
			reqParamByte, _ := sonic.Marshal(cg)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Failed when CreateConceptGroup service returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ConceptGroup_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CreateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", httpErr)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ConceptGroupRestHandler_UpdateConceptGroup_extraCases(t *testing.T) {
	Convey("Test ConceptGroupHandler UpdateConceptGroup extra cases\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID

		conceptGroup := interfaces.ConceptGroup{
			CGID:   cgID,
			CGName: "group1",
		}

		Convey("Failed when CheckKNExistByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ConceptGroup_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, httpErr)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when CheckConceptGroupExistByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ConceptGroup_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("", false, httpErr)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed ValidateConceptGroup with empty CGName\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("old_group", true, nil)

			reqParamByte, _ := sonic.Marshal(interfaces.ConceptGroup{})
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when CheckConceptGroupExistByName returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ConceptGroup_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("old_name", true, nil)
			cgs.EXPECT().CheckConceptGroupExistByName(gomock.Any(), knID, gomock.Any(), conceptGroup.CGName).Return("", false, httpErr)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when concept group name already exists\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("old_name", true, nil)
			cgs.EXPECT().CheckConceptGroupExistByName(gomock.Any(), knID, gomock.Any(), conceptGroup.CGName).Return(cgID, true, nil)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Failed when UpdateConceptGroup service returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ConceptGroup_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("old_name", true, nil)
			cgs.EXPECT().CheckConceptGroupExistByName(gomock.Any(), knID, gomock.Any(), conceptGroup.CGName).Return("", false, nil)
			cgs.EXPECT().UpdateConceptGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(httpErr)

			reqParamByte, _ := sonic.Marshal(conceptGroup)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ConceptGroupRestHandler_DeleteConceptGroup_extraCases(t *testing.T) {
	Convey("Test ConceptGroupHandler DeleteConceptGroup extra cases\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		cgID := "cg1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID

		Convey("Failed when CheckKNExistByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ConceptGroup_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, httpErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when CheckConceptGroupExistByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ConceptGroup_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("", false, httpErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when DeleteConceptGroupByID service returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ConceptGroup_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().CheckConceptGroupExistByID(gomock.Any(), knID, gomock.Any(), cgID).Return("group1", true, nil)
			cgs.EXPECT().DeleteConceptGroupByID(gomock.Any(), gomock.Any(), knID, gomock.Any(), cgID).Return(httpErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ConceptGroupRestHandler_ListConceptGroups_extraCases(t *testing.T) {
	Convey("Test ConceptGroupHandler ListConceptGroups extra cases\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups"

		Convey("Failed when CheckKNExistByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ConceptGroup_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, httpErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when pagination parameters are invalid\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?limit=invalid", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when ListConceptGroups service returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ConceptGroup_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			cgs.EXPECT().ListConceptGroups(gomock.Any(), gomock.Any()).Return(nil, 0, httpErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ConceptGroupRestHandler_DeleteConceptGroup_authFail(t *testing.T) {
	Convey("Test ConceptGroupHandler DeleteConceptGroup auth fail\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		cgs := bmock.NewMockConceptGroupService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewConceptGroupRestHandler(appSetting, as, cgs, kns)
		handler.RegisterPublic(engine)

		knID := "kn1"
		cgID := "cg1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/concept-groups/" + cgID

		Convey("Failed when VerifyToken returns error\n", func() {
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(hydra.Visitor{}, fmt.Errorf("token invalid"))

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusUnauthorized)
		})
	})
}
