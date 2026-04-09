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

func MockNewObjectTypeRestHandler(appSetting *common.AppSetting,
	as interfaces.AuthService,
	ots interfaces.ObjectTypeService,
	rts interfaces.RelationTypeService,
	ats interfaces.ActionTypeService,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		as:         as,
		ots:        ots,
		rts:        rts,
		ats:        ats,
		kns:        kns,
	}
	return r
}

func Test_ObjectTypeRestHandler_CreateObjectTypes(t *testing.T) {
	Convey("Test ObjectTypeHandler CreateObjectTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewObjectTypeRestHandler(appSetting, as, ots, rts, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/object-types"

		objectType := &interfaces.ObjectType{
			ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
				OTID:   "ot1",
				OTName: "object1",
				DataProperties: []*interfaces.DataProperty{
					{
						Name:        "prop1",
						Type:        "string",
						DisplayName: "prop1",
					},
				},
				PrimaryKeys: []string{"prop1"},
				DisplayKey:  "prop1",
			},
		}
		requestData := struct {
			Entries []*interfaces.ObjectType `json:"entries"`
		}{
			Entries: []*interfaces.ObjectType{objectType},
		}

		Convey("Success CreateObjectTypes \n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CreateObjectTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"ot1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("Failed CreateObjectTypes ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			reqParamByte, _ := sonic.Marshal([]interfaces.ObjectType{*objectType})
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty entries\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			emptyRequestData := struct {
				Entries []*interfaces.ObjectType `json:"entries"`
			}{
				Entries: []*interfaces.ObjectType{},
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
					ErrorCode: berrors.BknBackend_ObjectType_InternalError,
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

		Convey("CreateObjectTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ObjectType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CreateObjectTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, err)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

	})
}

func Test_ObjectTypeRestHandler_UpdateObjectType(t *testing.T) {
	Convey("Test ObjectTypeHandler UpdateObjectType\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewObjectTypeRestHandler(appSetting, as, ots, rts, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		otID := "ot1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/object-types/" + otID

		objectType := &interfaces.ObjectType{
			ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
				OTID:   otID,
				OTName: "object1",
				DataProperties: []*interfaces.DataProperty{
					{
						Name:        "prop1",
						Type:        "string",
						DisplayName: "prop1",
					},
				},
				PrimaryKeys: []string{"prop1"},
				DisplayKey:  "prop1",
			},
		}

		Convey("Success UpdateObjectType\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), otID).Return("object2", true, nil)
			ots.EXPECT().CheckObjectTypeExistByName(gomock.Any(), knID, gomock.Any(), objectType.OTName).Return("", false, nil)
			ots.EXPECT().UpdateObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(objectType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Failed UpdateObjectType ShouldBind Error\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)

			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte("invalid json")))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(objectType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("ObjectType not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), otID).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(objectType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("ObjectType name already exists\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), otID).Return("oldname", true, nil)
			ots.EXPECT().CheckObjectTypeExistByName(gomock.Any(), knID, gomock.Any(), objectType.OTName).Return("object1", true, nil)

			reqParamByte, _ := sonic.Marshal(objectType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("UpdateObjectType failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ObjectType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), otID).Return("object1", true, nil)
			ots.EXPECT().UpdateObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(err)

			reqParamByte, _ := sonic.Marshal(objectType)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ObjectTypeRestHandler_DeleteObjectTypes(t *testing.T) {
	Convey("Test ObjectTypeHandler DeleteObjectTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewObjectTypeRestHandler(appSetting, as, ots, rts, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		otIDs := "ot1,ot2"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/object-types/" + otIDs

		Convey("Success DeleteObjectTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), "ot1").Return("object1", true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), "ot2").Return("object2", true, nil)
			ots.EXPECT().DeleteObjectTypesByIDs(gomock.Any(), gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, 0, nil)
			ats.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{}, 0, nil)

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

		Convey("ObjectType not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), "ot1").Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("DeleteObjectTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ObjectType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), "ot1").Return("object1", true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), "ot2").Return("object2", true, nil)
			ots.EXPECT().DeleteObjectTypesByIDs(gomock.Any(), gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(err)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, 0, nil)
			ats.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{}, 0, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_ObjectTypeRestHandler_UpdateDataProperties(t *testing.T) {
	Convey("Test ObjectTypeHandler UpdateDataProperties\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewObjectTypeRestHandler(appSetting, as, ots, rts, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		otID := "ot1"
		propertyNames := "prop1,prop2"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/object-types/" + otID + "/data_properties/" + propertyNames

		dataProperties := []*interfaces.DataProperty{
			{
				Name:        "prop1",
				Type:        "string",
				DisplayName: "prop1",
			},
			{
				Name:        "prop2",
				Type:        "integer",
				DisplayName: "prop2",
			},
		}
		requestData := struct {
			Entries []*interfaces.DataProperty `json:"entries"`
		}{
			Entries: dataProperties,
		}

		Convey("Success UpdateDataProperties\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), knID, gomock.Any(), otID).Return(&interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   otID,
					OTName: "object1",
				},
			}, nil)
			ots.EXPECT().UpdateDataProperties(gomock.Any(), gomock.Any(), gomock.Any(), true).Return(nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Success UpdateDataProperties with strict_mode=false\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), knID, gomock.Any(), otID).Return(&interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   otID,
					OTName: "object1",
				},
			}, nil)
			ots.EXPECT().UpdateDataProperties(gomock.Any(), gomock.Any(), gomock.Any(), false).Return(nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPut, url+"?strict_mode=false", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Invalid strict_mode returns 400\n", func() {
			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPut, url+"?strict_mode=maybe", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("KN not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("ObjectType not found\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().GetObjectTypeByID(gomock.Any(), gomock.Any(), knID, gomock.Any(), otID).Return(nil, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})
	})
}

func Test_ObjectTypeRestHandler_ListObjectTypes(t *testing.T) {
	Convey("Test ObjectTypeHandler ListObjectTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewObjectTypeRestHandler(appSetting, as, ots, rts, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/object-types"

		Convey("Success ListObjectTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.ObjectType{}, 0, nil)

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

		Convey("ListObjectTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ObjectType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, err)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

	})
}

func Test_ObjectTypeRestHandler_GetObjectTypes(t *testing.T) {
	Convey("Test ObjectTypeHandler GetObjectTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewObjectTypeRestHandler(appSetting, as, ots, rts, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		otIDs := "ot1,ot2"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/object-types/" + otIDs

		Convey("Success GetObjectTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().GetObjectTypesByIDs(gomock.Any(), gomock.Any(), knID, gomock.Any(), gomock.Any()).Return([]*interfaces.ObjectType{}, nil)

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

		Convey("GetObjectTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ObjectType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().GetObjectTypesByIDs(gomock.Any(), gomock.Any(), knID, gomock.Any(), gomock.Any()).Return(nil, err)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

	})
}

func Test_ObjectTypeRestHandler_SearchObjectTypes(t *testing.T) {
	Convey("Test ObjectTypeHandler SearchObjectTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewObjectTypeRestHandler(appSetting, as, ots, rts, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/object-types"

		query := interfaces.ConceptsQuery{
			Limit: 10,
		}

		Convey("Success SearchObjectTypes\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().SearchObjectTypes(gomock.Any(), gomock.Any()).Return(interfaces.ObjectTypes{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed SearchObjectTypes ShouldBind Error\n", func() {
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

		Convey("SearchObjectTypes failed\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_ObjectType_InternalError,
				},
			}

			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().SearchObjectTypes(gomock.Any(), gomock.Any()).Return(interfaces.ObjectTypes{}, err)

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

func newObjectTypeTestHandler(t *testing.T) (*restHandler, *gomock.Controller, *gin.Engine, *bmock.MockObjectTypeService, *bmock.MockRelationTypeService, *bmock.MockActionTypeService, *bmock.MockKNService) {
	t.Helper()
	mockCtrl := gomock.NewController(t)
	engine := gin.New()
	engine.Use(gin.Recovery())
	appSetting := &common.AppSetting{}
	as := bmock.NewMockAuthService(mockCtrl)
	ots := bmock.NewMockObjectTypeService(mockCtrl)
	rts := bmock.NewMockRelationTypeService(mockCtrl)
	ats := bmock.NewMockActionTypeService(mockCtrl)
	kns := bmock.NewMockKNService(mockCtrl)
	handler := MockNewObjectTypeRestHandler(appSetting, as, ots, rts, ats, kns)
	handler.RegisterPublic(engine)
	return handler, mockCtrl, engine, ots, rts, ats, kns
}

func Test_ObjectTypeRestHandler_CreateObjectTypesByIn(t *testing.T) {
	Convey("Test ObjectTypeHandler CreateObjectTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, ots, _, _, kns := newObjectTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		requestData := struct {
			Entries []*interfaces.ObjectType `json:"entries"`
		}{
			Entries: []*interfaces.ObjectType{
				{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object1",
						DataProperties: []*interfaces.DataProperty{
							{Name: "prop1", Type: "string", DisplayName: "prop1"},
						},
						PrimaryKeys: []string{"prop1"},
						DisplayKey:  "prop1",
					},
				},
			},
		}

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CreateObjectTypes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{"ot1"}, nil)

			reqParamByte, _ := sonic.Marshal(requestData)
			req := httptest.NewRequest(http.MethodPost, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/object-types", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_ObjectTypeRestHandler_UpdateObjectTypeByIn(t *testing.T) {
	Convey("Test ObjectTypeHandler UpdateObjectTypeByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, ots, _, _, kns := newObjectTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		otID := "ot1"
		objectType := &interfaces.ObjectType{
			ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
				OTID:   otID,
				OTName: "object1",
				DataProperties: []*interfaces.DataProperty{
					{Name: "prop1", Type: "string", DisplayName: "prop1"},
				},
				PrimaryKeys: []string{"prop1"},
				DisplayKey:  "prop1",
			},
		}

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), otID).Return("object2", true, nil)
			ots.EXPECT().CheckObjectTypeExistByName(gomock.Any(), knID, gomock.Any(), objectType.OTName).Return("", false, nil)
			ots.EXPECT().UpdateObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(objectType)
			req := httptest.NewRequest(http.MethodPut, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/object-types/"+otID, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_ObjectTypeRestHandler_ListObjectTypesByIn(t *testing.T) {
	Convey("Test ObjectTypeHandler ListObjectTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, ots, _, _, kns := newObjectTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().ListObjectTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.ObjectType{}, 0, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/object-types", nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_ObjectTypeRestHandler_GetObjectTypesByIn(t *testing.T) {
	Convey("Test ObjectTypeHandler GetObjectTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, ots, _, _, kns := newObjectTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		otIDs := "ot1,ot2"

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().GetObjectTypesByIDs(gomock.Any(), gomock.Any(), knID, gomock.Any(), gomock.Any()).Return([]*interfaces.ObjectType{}, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/object-types/"+otIDs, nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_ObjectTypeRestHandler_SearchObjectTypesByIn(t *testing.T) {
	Convey("Test ObjectTypeHandler SearchObjectTypesByIn\n", t, func() {
		test := setGinMode()
		defer test()
		_, mockCtrl, engine, ots, _, _, kns := newObjectTypeTestHandler(t)
		defer mockCtrl.Finish()

		knID := "kn1"
		query := interfaces.ConceptsQuery{Limit: 10}

		Convey("Success\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().SearchObjectTypes(gomock.Any(), gomock.Any()).Return(interfaces.ObjectTypes{}, nil)

			reqParamByte, _ := sonic.Marshal(query)
			req := httptest.NewRequest(http.MethodPost, "/api/bkn-backend/in/v1/knowledge-networks/"+knID+"/object-types", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodGet)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_ObjectTypeRestHandler_HandleObjectTypeGetOverride_default(t *testing.T) {
	Convey("Test HandleObjectTypeGetOverride default method branch\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewObjectTypeRestHandler(appSetting, as, ots, rts, ats, kns)
		handler.RegisterPublic(engine)

		knID := "kn1"

		Convey("HandleObjectTypeGetOverrideByIn returns 400 for invalid method override\n", func() {
			url := "/api/bkn-backend/in/v1/knowledge-networks/" + knID + "/object-types"
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPut)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("HandleObjectTypeGetOverrideByEx returns 400 for invalid method override\n", func() {
			url := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/object-types"
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, http.MethodPut)
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func Test_ObjectTypeRestHandler_DeleteObjectTypes_extraCases(t *testing.T) {
	Convey("Test ObjectTypeHandler DeleteObjectTypes extra cases\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		ots := bmock.NewMockObjectTypeService(mockCtrl)
		rts := bmock.NewMockRelationTypeService(mockCtrl)
		ats := bmock.NewMockActionTypeService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewObjectTypeRestHandler(appSetting, as, ots, rts, ats, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		knID := "kn1"
		otIDs := "ot1"
		baseURL := "/api/bkn-backend/v1/knowledge-networks/" + knID + "/object-types/" + otIDs

		Convey("Failed when CheckKNExistByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_KnowledgeNetwork_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return("", false, httpErr)
			req := httptest.NewRequest(http.MethodDelete, baseURL, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when force_delete has invalid value\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			req := httptest.NewRequest(http.MethodDelete, baseURL+"?force_delete=notbool", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when CheckObjectTypeExistByID returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ObjectType_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), "ot1").Return("", false, httpErr)
			req := httptest.NewRequest(http.MethodDelete, baseURL, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when ListRelationTypes returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ObjectType_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), "ot1").Return("ot1", true, nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return(nil, 0, httpErr)
			req := httptest.NewRequest(http.MethodDelete, baseURL, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when object type is bound by relation type\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), "ot1").Return("ot1", true, nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{
				{RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{RTName: "rt1"}},
			}, 1, nil)
			req := httptest.NewRequest(http.MethodDelete, baseURL, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when ListActionTypes returns error\n", func() {
			httpErr := &rest.HTTPError{
				HTTPCode:  http.StatusInternalServerError,
				Language:  rest.DefaultLanguage,
				BaseError: rest.BaseError{ErrorCode: berrors.BknBackend_ObjectType_InternalError},
			}
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), "ot1").Return("ot1", true, nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, 0, nil)
			ats.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return(nil, 0, httpErr)
			req := httptest.NewRequest(http.MethodDelete, baseURL, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Failed when object type is bound by action type\n", func() {
			kns.EXPECT().CheckKNExistByID(gomock.Any(), knID, gomock.Any()).Return(knID, true, nil)
			ots.EXPECT().CheckObjectTypeExistByID(gomock.Any(), knID, gomock.Any(), "ot1").Return("ot1", true, nil)
			rts.EXPECT().ListRelationTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.RelationType{}, 0, nil)
			ats.EXPECT().ListActionTypes(gomock.Any(), gomock.Any()).Return([]*interfaces.ActionType{
				{ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{ATName: "at1"}},
			}, 1, nil)
			req := httptest.NewRequest(http.MethodDelete, baseURL, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}
