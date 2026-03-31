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

func MockNewKnowledgeNetworkRestHandler(appSetting *common.AppSetting,
	as interfaces.AuthService,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		as:         as,
		kns:        kns,
	}
	return r
}

func Test_KnowledgeNetworkRestHandler_CreateKN(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler CreateKN\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, as, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/knowledge-networks"

		kn := interfaces.KN{
			KNName:  "kn1",
			Comment: "test comment",
			Branch:  "main",
		}

		Convey("Success CreateKN \n", func() {
			kns.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", nil)

			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateKN ShouldBind Error\n", func() {
			reqParamByte, _ := sonic.Marshal([]interfaces.KN{kn})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN name is null\n", func() {
			reqParamByte, _ := sonic.Marshal(interfaces.KN{})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Business domain is empty, proceeds with empty domain\n", func() {
			kns.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", nil)

			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("CreateKN failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_KnowledgeNetwork_InternalError,
				},
			}

			kns.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", err)

			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

	})
}

func Test_KnowledgeNetworkRestHandler_UpdateKN(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler UpdateKN\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, as, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID

		kn := interfaces.KN{
			KNID:   knID,
			KNName: "kn1",
			Branch: interfaces.MAIN_BRANCH,
		}

		Convey("Success UpdateKN\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("kn2", true, nil)
			kns.EXPECT().CheckKNExistByName(gomock.Any(), kn.KNName, gomock.Any()).Return("", false, nil)
			kns.EXPECT().UpdateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateKN ShouldBind Error\n", func() {
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_DeleteKN(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler DeleteKN\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, as, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID

		Convey("Success DeleteKN\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(&interfaces.KN{
				KNID:   knID,
				KNName: "kn1",
			}, nil)
			kns.EXPECT().DeleteKN(gomock.Any(), gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_ListKNs(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler ListKNs\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, as, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/knowledge-networks"

		Convey("Success ListKNs\n", func() {
			kns.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return([]*interfaces.KN{}, 0, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?business_domain=domain1", nil)
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Business domain is empty, proceeds with empty domain\n", func() {
			kns.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return([]*interfaces.KN{}, 0, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_GetKN(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler GetKN\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, as, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID

		Convey("Success GetKN\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(&interfaces.KN{}, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("KN not found\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusNotFound,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_KnowledgeNetwork_NotFound,
				},
			}

			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil, err)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_GetRelationTypePaths(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler GetRelationTypePaths\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, as, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/relation-type-paths"

		query := interfaces.RelationTypePathsBaseOnSource{
			SourceObjecTypeId: "ot1",
			Direction:         interfaces.DIRECTION_FORWARD,
			PathLength:        2,
		}

		Convey("Success GetRelationTypePaths\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(&interfaces.KN{
				KNID:   knID,
				KNName: "kn1",
				Branch: interfaces.MAIN_BRANCH,
			}, nil)
			kns.EXPECT().GetRelationTypePaths(gomock.Any(), gomock.Any()).Return([]interfaces.RelationTypePath{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed GetRelationTypePaths ShouldBind Error\n", func() {

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

func newKNTestHandler(t *testing.T) (*restHandler, *gomock.Controller, *gin.Engine, *bmock.MockKNService) {
	t.Helper()
	mockCtrl := gomock.NewController(t)
	engine := gin.New()
	engine.Use(gin.Recovery())
	appSetting := &common.AppSetting{}
	as := bmock.NewMockAuthService(mockCtrl)
	kns := bmock.NewMockKNService(mockCtrl)
	handler := MockNewKnowledgeNetworkRestHandler(appSetting, as, kns)
	handler.RegisterPublic(engine)
	return handler, mockCtrl, engine, kns
}

func Test_KnowledgeNetworkRestHandler_CreateKNByIn(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler CreateKNByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, kns := newKNTestHandler(t)
		defer mockCtrl.Finish()

		Convey("Success\n", func() {
			kns.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", nil)

			kn := interfaces.KN{KNName: "kn1", Comment: "test comment", Branch: "main"}
			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPost, "/api/bkn-backend/in/v1/knowledge-networks", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_UpdateKNByIn(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler UpdateKNByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, kns := newKNTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("kn2", true, nil)
			kns.EXPECT().CheckKNExistByName(gomock.Any(), "kn1", gomock.Any()).Return("", false, nil)
			kns.EXPECT().UpdateKN(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(interfaces.KN{KNID: knID, KNName: "kn1", Branch: interfaces.MAIN_BRANCH})
			req := httptest.NewRequest(http.MethodPut, "/api/bkn-backend/in/v1/knowledge-networks/"+knID, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_ListKNsByIn(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler ListKNsByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, kns := newKNTestHandler(t)
		defer mockCtrl.Finish()

		Convey("Success\n", func() {
			kns.EXPECT().ListKNs(gomock.Any(), gomock.Any()).Return([]*interfaces.KN{}, 0, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks", nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_GetKNByIn(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler GetKNByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, kns := newKNTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(&interfaces.KN{}, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks/"+knID, nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_GetRelationTypePathsByIn(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler GetRelationTypePathsByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, kns := newKNTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		query := interfaces.RelationTypePathsBaseOnSource{
			SourceObjecTypeId: "ot1",
			Direction:         interfaces.DIRECTION_FORWARD,
			PathLength:        2,
		}

		Convey("Success\n", func() {
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(&interfaces.KN{
				KNID: knID, KNName: "kn1", Branch: interfaces.MAIN_BRANCH,
			}, nil)
			kns.EXPECT().GetRelationTypePaths(gomock.Any(), gomock.Any()).Return([]interfaces.RelationTypePath{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/relation-type-paths", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_CreateKN_extraCases(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler CreateKN extra cases\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewKnowledgeNetworkRestHandler(appSetting, as, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/knowledge-networks"
		kn := interfaces.KN{KNName: "kn1", Branch: "main"}

		Convey("Failed when validate_dependency has invalid value\n", func() {
			reqParamByte, _ := sonic.Marshal(kn)
			req := httptest.NewRequest(http.MethodPost, url+"?validate_dependency=notabool", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when ModuleType is set to non-KN value\n", func() {
			knWithBadType := interfaces.KN{KNName: "kn1", Branch: "main", ModuleType: "object_type"}
			reqParamByte, _ := sonic.Marshal(knWithBadType)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_KnowledgeNetworkRestHandler_DeleteKN_extraCases(t *testing.T) {
	Convey("Test KnowledgeNetworkHandler DeleteKN extra cases\n", t, func() {
		test := setGinMode()
		defer test()
		engine := gin.New()
		engine.Use(gin.Recovery())
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)
		handler := MockNewKnowledgeNetworkRestHandler(appSetting, as, kns)
		handler.RegisterPublic(engine)
		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)
		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID

		Convey("Failed when GetKNByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_KnowledgeNetwork_InternalError},
			}
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil, httpErr)
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when DeleteKN service returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_KnowledgeNetwork_InternalError},
			}
			kns.EXPECT().GetKNByID(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(&interfaces.KN{
				KNID: knID, KNName: "kn1",
			}, nil)
			kns.EXPECT().DeleteKN(gomock.Any(), gomock.Any()).Return(httpErr)
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
