// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"ontology-query/common"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	omock "ontology-query/interfaces/mock"
)

func Test_RestHandler_GetObjectsInObjectTypeByIn(t *testing.T) {
	Convey("Test RestHandler GetObjectsInObjectTypeByIn", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := omock.NewMockAuthService(mockCtrl)
		ats := omock.NewMockActionTypeService(mockCtrl)
		kns := omock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, as, ats, kns, ots)
		handler.RegisterPublic(engine)

		knID := "kn1"
		otID := "ot1"
		url := "/api/ontology-query/in/v1/knowledge-networks/" + knID + "/object-types/" + otID

		objectQuery := interfaces.ObjectQueryBaseOnObjectType{
			PageQuery: interfaces.PageQuery{
				Limit: 10,
			},
		}

		Convey("成功 - 获取对象数据", func() {
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{
				Datas: []map[string]any{
					{"id": "1", "name": "obj1"},
				},
				TotalCount: 1,
			}, nil)

			reqParamByte, _ := sonic.Marshal(objectQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - 参数绑定失败", func() {
			reqParamByte := []byte("invalid json")
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("失败 - Service返回错误", func() {
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{}, rest.NewHTTPError(context.TODO(), http.StatusInternalServerError, oerrors.OntologyQuery_ObjectType_InternalError_GetObjectTypesByIDFailed))

			reqParamByte, _ := sonic.Marshal(objectQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_RestHandler_GetObjectsInObjectTypeByEx(t *testing.T) {
	Convey("Test RestHandler GetObjectsInObjectTypeByEx", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := omock.NewMockAuthService(mockCtrl)
		ats := omock.NewMockActionTypeService(mockCtrl)
		kns := omock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, as, ats, kns, ots)
		handler.RegisterPublic(engine)

		knID := "kn1"
		otID := "ot1"
		url := "/api/ontology-query/v1/knowledge-networks/" + knID + "/object-types/" + otID

		objectQuery := interfaces.ObjectQueryBaseOnObjectType{
			PageQuery: interfaces.PageQuery{
				Limit: 10,
			},
		}

		Convey("成功 - Token验证通过，获取对象数据", func() {
			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(visitor, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{
				Datas: []map[string]any{
					{"id": "1", "name": "obj1"},
				},
				TotalCount: 1,
			}, nil)

			reqParamByte, _ := sonic.Marshal(objectQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - Token验证失败", func() {
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(hydra.Visitor{}, rest.NewHTTPError(context.TODO(), http.StatusUnauthorized, rest.PublicError_Unauthorized))

			reqParamByte, _ := sonic.Marshal(objectQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

func Test_RestHandler_GetObjectsInObjectType(t *testing.T) {
	Convey("Test RestHandler GetObjectsInObjectType", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := omock.NewMockAuthService(mockCtrl)
		ats := omock.NewMockActionTypeService(mockCtrl)
		kns := omock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, as, ats, kns, ots)

		knID := "kn1"
		otID := "ot1"
		url := "/api/ontology-query/v1/knowledge-networks/" + knID + "/object-types/" + otID

		objectQuery := interfaces.ObjectQueryBaseOnObjectType{
			PageQuery: interfaces.PageQuery{
				Limit: 10,
			},
		}

		Convey("成功 - 参数验证通过，获取对象数据", func() {
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{
				Datas: []map[string]any{
					{"id": "1", "name": "obj1"},
				},
				TotalCount: 1,
			}, nil)

			reqParamByte, _ := sonic.Marshal(objectQuery)
			req := httptest.NewRequest(http.MethodPost, url+"?branch=main", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
				{Key: "ot_id", Value: otID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsInObjectType(c, visitor)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - includeTypeInfo参数无效", func() {
			reqParamByte, _ := sonic.Marshal(objectQuery)
			req := httptest.NewRequest(http.MethodPost, url+"?include_type_info=invalid", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
				{Key: "ot_id", Value: otID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsInObjectType(c, visitor)

			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("失败 - Limit无效", func() {
			invalidQuery := interfaces.ObjectQueryBaseOnObjectType{
				PageQuery: interfaces.PageQuery{
					Limit: 0,
				},
			}
			reqParamByte, _ := sonic.Marshal(invalidQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
				{Key: "ot_id", Value: otID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsInObjectType(c, visitor)

			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("失败 - Service返回错误", func() {
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{}, rest.NewHTTPError(context.TODO(), http.StatusInternalServerError, oerrors.OntologyQuery_ObjectType_InternalError_GetObjectTypesByIDFailed))

			reqParamByte, _ := sonic.Marshal(objectQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
				{Key: "ot_id", Value: otID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsInObjectType(c, visitor)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_RestHandler_GetObjectsPropertiesByIn(t *testing.T) {
	Convey("Test RestHandler GetObjectsPropertiesByIn", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := omock.NewMockAuthService(mockCtrl)
		ats := omock.NewMockActionTypeService(mockCtrl)
		kns := omock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, as, ats, kns, ots)
		handler.RegisterPublic(engine)

		knID := "kn1"
		otID := "ot1"
		url := "/api/ontology-query/in/v1/knowledge-networks/" + knID + "/object-types/" + otID + "/properties"

		propertyQuery := interfaces.ObjectPropertyValueQuery{
			InstanceIdentities: []map[string]any{
				{"id": "1"},
			},
			Properties: []string{"prop1", "prop2"},
		}

		Convey("成功 - 获取对象属性值", func() {
			ots.EXPECT().GetObjectPropertyValue(gomock.Any(), gomock.Any()).Return(interfaces.Objects{
				Datas: []map[string]any{
					{"id": "1", "prop1": "value1", "prop2": "value2"},
				},
				TotalCount: 1,
			}, nil)

			reqParamByte, _ := sonic.Marshal(propertyQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - 唯一标识为空", func() {
			invalidQuery := interfaces.ObjectPropertyValueQuery{
				InstanceIdentities: []map[string]any{},
				Properties:         []string{"prop1"},
			}
			reqParamByte, _ := sonic.Marshal(invalidQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("失败 - 属性列表为空", func() {
			invalidQuery := interfaces.ObjectPropertyValueQuery{
				InstanceIdentities: []map[string]any{
					{"id": "1"},
				},
				Properties: []string{},
			}
			reqParamByte, _ := sonic.Marshal(invalidQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func Test_RestHandler_GetObjectsPropertiesByEx(t *testing.T) {
	Convey("Test RestHandler GetObjectsPropertiesByEx", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := omock.NewMockAuthService(mockCtrl)
		ats := omock.NewMockActionTypeService(mockCtrl)
		kns := omock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, as, ats, kns, ots)
		handler.RegisterPublic(engine)

		knID := "kn1"
		otID := "ot1"
		url := "/api/ontology-query/v1/knowledge-networks/" + knID + "/object-types/" + otID + "/properties"

		propertyQuery := interfaces.ObjectPropertyValueQuery{
			InstanceIdentities: []map[string]any{
				{"id": "1"},
			},
			Properties: []string{"prop1", "prop2"},
		}

		Convey("成功 - Token验证通过，获取对象属性值", func() {
			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(visitor, nil)
			ots.EXPECT().GetObjectPropertyValue(gomock.Any(), gomock.Any()).Return(interfaces.Objects{
				Datas: []map[string]any{
					{"id": "1", "prop1": "value1", "prop2": "value2"},
				},
				TotalCount: 1,
			}, nil)

			reqParamByte, _ := sonic.Marshal(propertyQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - Token验证失败", func() {
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(hydra.Visitor{}, rest.NewHTTPError(context.TODO(), http.StatusUnauthorized, rest.PublicError_Unauthorized))

			reqParamByte, _ := sonic.Marshal(propertyQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

func Test_RestHandler_GetObjectsProperties(t *testing.T) {
	Convey("Test RestHandler GetObjectsProperties", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := omock.NewMockAuthService(mockCtrl)
		ats := omock.NewMockActionTypeService(mockCtrl)
		kns := omock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, as, ats, kns, ots)

		knID := "kn1"
		otID := "ot1"
		url := "/api/ontology-query/v1/knowledge-networks/" + knID + "/object-types/" + otID + "/properties"

		propertyQuery := interfaces.ObjectPropertyValueQuery{
			InstanceIdentities: []map[string]any{
				{"id": "1"},
			},
			Properties: []string{"prop1", "prop2"},
		}

		Convey("成功 - 参数验证通过，获取对象属性值", func() {
			ots.EXPECT().GetObjectPropertyValue(gomock.Any(), gomock.Any()).Return(interfaces.Objects{
				Datas: []map[string]any{
					{"id": "1", "prop1": "value1", "prop2": "value2"},
				},
				TotalCount: 1,
			}, nil)

			reqParamByte, _ := sonic.Marshal(propertyQuery)
			req := httptest.NewRequest(http.MethodPost, url+"?branch=main", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
				{Key: "ot_id", Value: otID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsProperties(c, visitor)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - includeTypeInfo参数无效", func() {
			reqParamByte, _ := sonic.Marshal(propertyQuery)
			req := httptest.NewRequest(http.MethodPost, url+"?include_type_info=invalid", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
				{Key: "ot_id", Value: otID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsProperties(c, visitor)

			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("失败 - Service返回错误", func() {
			ots.EXPECT().GetObjectPropertyValue(gomock.Any(), gomock.Any()).Return(interfaces.Objects{}, rest.NewHTTPError(context.TODO(), http.StatusInternalServerError, oerrors.OntologyQuery_ObjectType_InternalError_GetObjectTypesByIDFailed))

			reqParamByte, _ := sonic.Marshal(propertyQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
				{Key: "ot_id", Value: otID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsProperties(c, visitor)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
