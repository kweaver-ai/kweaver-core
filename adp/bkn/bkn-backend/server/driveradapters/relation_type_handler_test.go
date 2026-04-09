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

func MockNewRelationTypeRestHandler(appSetting *common.AppSetting,
	as interfaces.AuthService,
	rts interfaces.RelationTypeService,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		as:         as,
		rts:        rts,
		kns:        kns,
	}
	return r
}

func Test_RelationTypeRestHandler_CreateRelationTypes(t *testing.T) {
	Convey("Test RelationTypeHandler CreateRelationTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewRelationTypeRestHandler(appSetting, as, rts, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/relation-types"

		relationType := &interfaces.RelationType{
			RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
				RTID:               "rt1",
				RTName:             "relation1",
				SourceObjectTypeID: "ot1",
				TargetObjectTypeID: "ot2",
				Type:               interfaces.RELATION_TYPE_DIRECT,
				MappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "prop1"},
						TargetProp: interfaces.SimpleProperty{Name: "prop2"},
					},
				},
			},
			CommonInfo: interfaces.CommonInfo{
				Tags:          []string{"tag1", "tag2", "tag3"},
				Comment:       "test comment",
				Icon:          "icon1",
				Color:         "color1",
				BKNRawContent: "bkn1",
			},
		}
		requestData := struct {
			Entries []*interfaces.RelationType `json:"entries"`
		}{
			Entries: []*interfaces.RelationType{relationType},
		}

		Convey("Success CreateRelationTypes \n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CreateRelationTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"rt1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateRelationTypes ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			reqParamByte, _ := sonic.Marshal([]interfaces.RelationType{*relationType})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty entries\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			emptyRequestData := struct {
				Entries []*interfaces.RelationType `json:"entries"`
			}{
				Entries: []*interfaces.RelationType{},
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
					ErrorCode: berrors.BknBackend_RelationType_InternalError,
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

		Convey("CreateRelationTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_RelationType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CreateRelationTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, err)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_RelationTypeRestHandler_UpdateRelationType(t *testing.T) {
	Convey("Test RelationTypeHandler UpdateRelationType\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewRelationTypeRestHandler(appSetting, as, rts, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		rtID := "rt1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/relation-types/" + rtID

		relationType := &interfaces.RelationType{
			RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
				RTID:               rtID,
				RTName:             "relation1",
				SourceObjectTypeID: "ot1",
				TargetObjectTypeID: "ot2",
				Type:               interfaces.RELATION_TYPE_DIRECT,
				MappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "prop1"},
						TargetProp: interfaces.SimpleProperty{Name: "prop2"},
					},
				},
			},
			CommonInfo: interfaces.CommonInfo{
				Tags:          []string{"tag1", "tag2", "tag3"},
				Comment:       "test comment",
				Icon:          "icon1",
				Color:         "color1",
				BKNRawContent: "bkn1",
			},
		}

		Convey("Success UpdateRelationType\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CheckRelationTypeExistByID(gomock.Any(), knID, gomock.Any(), rtID).Return("relation2", true, nil)
			rts.EXPECT().UpdateRelationType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(relationType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateRelationType ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(relationType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("RelationType not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CheckRelationTypeExistByID(gomock.Any(), knID, gomock.Any(), rtID).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(relationType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Update to existing name succeeds (no name uniqueness check)\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CheckRelationTypeExistByID(gomock.Any(), knID, gomock.Any(), rtID).Return("oldname", true, nil)
			rts.EXPECT().UpdateRelationType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(relationType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("UpdateRelationType failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_RelationType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CheckRelationTypeExistByID(gomock.Any(), knID, gomock.Any(), rtID).Return("relation2", true, nil)
			rts.EXPECT().UpdateRelationType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(err)

			reqParamByte, _ := sonic.Marshal(relationType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("UpdateRelationTypeByIn - Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CheckRelationTypeExistByID(gomock.Any(), knID, gomock.Any(), rtID).Return("old_relation1", true, nil)
			rts.EXPECT().UpdateRelationType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			urlIn := "/api/bkn-backend/in/v1/knowledge-networks/" + knID + "/relation-types/" + rtID
			reqParamByte, _ := sonic.Marshal(relationType)
			req := httptest.NewRequest(http.MethodPut, urlIn, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_RelationTypeRestHandler_DeleteRelationTypes(t *testing.T) {
	Convey("Test RelationTypeHandler DeleteRelationTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewRelationTypeRestHandler(appSetting, as, rts, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		rtIDs := "rt1,rt2"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/relation-types/" + rtIDs

		Convey("Success DeleteRelationTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CheckRelationTypeExistByID(gomock.Any(), knID, gomock.Any(), "rt1").Return("relation1", true, nil)
			rts.EXPECT().CheckRelationTypeExistByID(gomock.Any(), knID, gomock.Any(), "rt2").Return("relation2", true, nil)
			rts.EXPECT().DeleteRelationTypesByIDs(gomock.Any(), gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil)

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

		Convey("RelationType not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CheckRelationTypeExistByID(gomock.Any(), knID, gomock.Any(), "rt1").Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("DeleteRelationTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_RelationType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CheckRelationTypeExistByID(gomock.Any(), knID, gomock.Any(), "rt1").Return("relation1", true, nil)
			rts.EXPECT().CheckRelationTypeExistByID(gomock.Any(), knID, gomock.Any(), "rt2").Return("relation2", true, nil)
			rts.EXPECT().DeleteRelationTypesByIDs(gomock.Any(), gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(err)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_RelationTypeRestHandler_ListRelationTypes(t *testing.T) {
	Convey("Test RelationTypeHandler ListRelationTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewRelationTypeRestHandler(appSetting, as, rts, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/relation-types"

		Convey("Success ListRelationTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, 0, nil)

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

		Convey("ListRelationTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_RelationType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return(nil, 0, err)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_RelationTypeRestHandler_GetRelationTypes(t *testing.T) {
	Convey("Test RelationTypeHandler GetRelationTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewRelationTypeRestHandler(appSetting, as, rts, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		rtIDs := "rt1,rt2"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/relation-types/" + rtIDs

		Convey("Success GetRelationTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().GetRelationTypesByIDs(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, nil)

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

		Convey("GetRelationTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_RelationType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().GetRelationTypesByIDs(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil, err)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_RelationTypeRestHandler_SearchRelationTypes(t *testing.T) {
	Convey("Test RelationTypeHandler SearchRelationTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewRelationTypeRestHandler(appSetting, as, rts, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/relation-types"

		query := interfaces.ConceptsQuery{
			Limit: 10,
		}

		Convey("Success SearchRelationTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().SearchRelationTypes(gomock.Any(), gomock.Any()).Return(interfaces.RelationTypes{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed SearchRelationTypes ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
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
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("SearchRelationTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_RelationType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().SearchRelationTypes(gomock.Any(), gomock.Any()).Return(interfaces.RelationTypes{}, err)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_RelationTypeRestHandler_HandleRelationTypeGetOverride(t *testing.T) {
	Convey("Test RelationTypeHandler HandleRelationTypeGetOverrideByEx\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewRelationTypeRestHandler(appSetting, as, rts, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		urlEx := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/relation-types"

		relationType := &interfaces.RelationType{
			RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
				RTID:               "rt1",
				RTName:             "relation1",
				SourceObjectTypeID: "ot1",
				TargetObjectTypeID: "ot2",
				Type:               interfaces.RELATION_TYPE_DIRECT,
				MappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "prop1"},
						TargetProp: interfaces.SimpleProperty{Name: "prop2"},
					},
				},
			},
			CommonInfo: interfaces.CommonInfo{
				Tags:          []string{"tag1", "tag2", "tag3"},
				Comment:       "test comment",
				Icon:          "icon1",
				Color:         "color1",
				BKNRawContent: "bkn1",
			},
		}
		requestData := struct {
			Entries []*interfaces.RelationType `json:"entries"`
		}{
			Entries: []*interfaces.RelationType{relationType},
		}

		Convey("HandleRelationTypeGetOverrideByEx - Success with POST method (default)\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CreateRelationTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"rt1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, urlEx, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("HandleRelationTypeGetOverrideByEx - Success with GET override method\n", func() {
			query := interfaces.ConceptsQuery{
				Limit: 10,
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().SearchRelationTypes(gomock.Any(), gomock.Any()).Return(interfaces.RelationTypes{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, urlEx, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("HandleRelationTypeGetOverrideByEx - Failed with invalid override method\n", func() {
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

func newRelationTypeTestHandler(t *testing.T) (*restHandler, *gomock.Controller, *gin.Engine, *bmock.MockRelationTypeService, *bmock.MockKNService) {
	t.Helper()
	mockCtrl := gomock.NewController(t)
	engine := gin.New()
	engine.Use(gin.Recovery())
	appSetting := &common.AppSetting{}
	as := bmock.NewMockAuthService(mockCtrl)
	rts := bmock.NewMockRelationTypeService(mockCtrl)
	kns := bmock.NewMockKNService(mockCtrl)
	handler := MockNewRelationTypeRestHandler(appSetting, as, rts, kns)
	handler.RegisterPublic(engine)
	return handler, mockCtrl, engine, rts, kns
}

func Test_RelationTypeRestHandler_CreateRelationTypesByIn(t *testing.T) {
	Convey("Test RelationTypeHandler CreateRelationTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, rts, kns := newRelationTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		relationType := &interfaces.RelationType{
			RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
				RTID:               "rt1",
				RTName:             "relation1",
				SourceObjectTypeID: "ot1",
				TargetObjectTypeID: "ot2",
				Type:               interfaces.RELATION_TYPE_DIRECT,
				MappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "prop1"},
						TargetProp: interfaces.SimpleProperty{Name: "prop2"},
					},
				},
			},
			CommonInfo: interfaces.CommonInfo{
				Tags:          []string{"tag1", "tag2", "tag3"},
				Comment:       "test comment",
				Icon:          "icon1",
				Color:         "color1",
				BKNRawContent: "bkn1",
			},
		}
		requestData := struct {
			Entries []*interfaces.RelationType `json:"entries"`
		}{
			Entries: []*interfaces.RelationType{relationType},
		}

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CreateRelationTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"rt1"}, nil)

			urlIn := "/api/bkn-backend/in/v1/knowledge-networks/" + knID + "/relation-types"
			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, urlIn, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_RelationTypeRestHandler_UpdateRelationTypeByIn(t *testing.T) {
	Convey("Test RelationTypeHandler UpdateRelationTypeByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, rts, kns := newRelationTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		rtID := "rt1"

		relationType := &interfaces.RelationType{
			RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
				RTID:               rtID,
				RTName:             "relation1",
				SourceObjectTypeID: "ot1",
				TargetObjectTypeID: "ot2",
				Type:               interfaces.RELATION_TYPE_DIRECT,
				MappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "prop1"},
						TargetProp: interfaces.SimpleProperty{Name: "prop2"},
					},
				},
			},
			CommonInfo: interfaces.CommonInfo{
				Tags:          []string{"tag1", "tag2", "tag3"},
				Comment:       "test comment",
				Icon:          "icon1",
				Color:         "color1",
				BKNRawContent: "bkn1",
			},
		}

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CheckRelationTypeExistByID(gomock.Any(), knID, gomock.Any(), rtID).Return("old_relation1", true, nil)
			rts.EXPECT().UpdateRelationType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			urlIn := "/api/bkn-backend/in/v1/knowledge-networks/" + knID + "/relation-types/" + rtID
			reqParamByte, _ := sonic.Marshal(relationType)
			req := httptest.NewRequest(http.MethodPut, urlIn, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_RelationTypeRestHandler_ListRelationTypesByIn(t *testing.T) {
	Convey("Test RelationTypeHandler ListRelationTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, rts, kns := newRelationTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, 0, nil)

			urlIn := "/api/bkn-backend/in/v1/knowledge-networks/" + knID + "/relation-types"
			req := httptest.NewRequest(http.MethodGet, urlIn, nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_RelationTypeRestHandler_GetRelationTypesByIn(t *testing.T) {
	Convey("Test RelationTypeHandler GetRelationTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, rts, kns := newRelationTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		rtIDs := "rt1,rt2"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().GetRelationTypesByIDs(gomock.Any(), knID, gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, nil)

			urlIn := "/api/bkn-backend/in/v1/knowledge-networks/" + knID + "/relation-types/" + rtIDs
			req := httptest.NewRequest(http.MethodGet, urlIn, nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_RelationTypeRestHandler_SearchRelationTypesByIn(t *testing.T) {
	Convey("Test RelationTypeHandler SearchRelationTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, rts, kns := newRelationTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().SearchRelationTypes(gomock.Any(), gomock.Any()).Return(interfaces.RelationTypes{}, nil)

			query := interfaces.ConceptsQuery{
				Limit: 10,
			}
			urlIn := "/api/bkn-backend/in/v1/knowledge-networks/" + knID + "/relation-types"
			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, urlIn, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_RelationTypeRestHandler_HandleRelationTypeGetOverrideByIn(t *testing.T) {
	Convey("Test RelationTypeHandler HandleRelationTypeGetOverrideByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, rts, kns := newRelationTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		urlIn := "/api/bkn-backend/in/v1/knowledge-networks/" + knID + "/relation-types"

		relationType := &interfaces.RelationType{
			RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
				RTID:               "rt1",
				RTName:             "relation1",
				SourceObjectTypeID: "ot1",
				TargetObjectTypeID: "ot2",
				Type:               interfaces.RELATION_TYPE_DIRECT,
				MappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "prop1"},
						TargetProp: interfaces.SimpleProperty{Name: "prop2"},
					},
				},
			},
			CommonInfo: interfaces.CommonInfo{
				Tags:          []string{"tag1", "tag2", "tag3"},
				Comment:       "test comment",
				Icon:          "icon1",
				Color:         "color1",
				BKNRawContent: "bkn1",
			},
		}
		requestData := struct {
			Entries []*interfaces.RelationType `json:"entries"`
		}{
			Entries: []*interfaces.RelationType{relationType},
		}

		Convey("HandleRelationTypeGetOverrideByIn - Success with POST method (default)\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().CreateRelationTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"rt1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, urlIn, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("HandleRelationTypeGetOverrideByIn - Success with GET override method\n", func() {
			query := interfaces.ConceptsQuery{
				Limit: 10,
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			rts.EXPECT().SearchRelationTypes(gomock.Any(), gomock.Any()).Return(interfaces.RelationTypes{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, urlIn, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("HandleRelationTypeGetOverrideByIn - Failed with invalid override method\n", func() {
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
