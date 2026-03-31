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

func Test_RestHandler_GetObjectsSubgraphByIn(t *testing.T) {
	Convey("Test RestHandler GetObjectsSubgraphByIn", t, func() {
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
		url := "/api/ontology-query/in/v1/knowledge-networks/" + knID + "/subgraph"

		subgraphQuery := interfaces.SubGraphQueryBaseOnSource{
			SourceObjecTypeId: "ot1",
			Direction:         interfaces.DIRECTION_FORWARD,
			PathLength:        2,
			PageQuery: interfaces.PageQuery{
				Limit: 100,
			},
		}

		Convey("成功 - 基于起点获取子图", func() {
			kns.EXPECT().SearchSubgraph(gomock.Any(), gomock.Any()).Return(interfaces.ObjectSubGraph{
				Objects:    map[string]interfaces.ObjectInfoInSubgraph{},
				TotalCount: 0,
			}, nil)

			reqParamByte, _ := sonic.Marshal(subgraphQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("成功 - 基于路径获取子图", func() {
			pathsQuery := interfaces.QueryRelationTypePaths{
				TypePaths: []interfaces.QueryRelationTypePath{
					{
						ObjectTypes: []interfaces.ObjectTypeWithKeyField{
							{OTID: "ot1"},
							{OTID: "ot2"},
						},
						Edges: []interfaces.TypeEdge{
							{
								RelationTypeId:     "rt1",
								SourceObjectTypeId: "ot1",
								TargetObjectTypeId: "ot2",
							},
						},
					},
				},
			}
			kns.EXPECT().SearchSubgraphByTypePath(gomock.Any(), gomock.Any()).Return(interfaces.PathsEntries{
				Entries: []interfaces.ObjectSubGraph{},
			}, nil)

			reqParamByte, _ := sonic.Marshal(pathsQuery)
			req := httptest.NewRequest(http.MethodPost, url+"?query_type="+interfaces.QUERY_TYPE_RELATION_TYPE_PATH, bytes.NewReader(reqParamByte))
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
	})
}

func Test_RestHandler_GetObjectsSubgraphByEx(t *testing.T) {
	Convey("Test RestHandler GetObjectsSubgraphByEx", t, func() {
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
		url := "/api/ontology-query/v1/knowledge-networks/" + knID + "/subgraph"

		subgraphQuery := interfaces.SubGraphQueryBaseOnSource{
			SourceObjecTypeId: "ot1",
			Direction:         interfaces.DIRECTION_FORWARD,
			PathLength:        2,
			PageQuery: interfaces.PageQuery{
				Limit: 100,
			},
		}

		Convey("成功 - Token验证通过，获取子图", func() {
			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(visitor, nil)
			kns.EXPECT().SearchSubgraph(gomock.Any(), gomock.Any()).Return(interfaces.ObjectSubGraph{
				Objects:    map[string]interfaces.ObjectInfoInSubgraph{},
				TotalCount: 0,
			}, nil)

			reqParamByte, _ := sonic.Marshal(subgraphQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - Token验证失败", func() {
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(hydra.Visitor{}, rest.NewHTTPError(context.TODO(), http.StatusUnauthorized, rest.PublicError_Unauthorized))

			reqParamByte, _ := sonic.Marshal(subgraphQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

func Test_RestHandler_GetObjectsSubgraph(t *testing.T) {
	Convey("Test RestHandler GetObjectsSubgraph", t, func() {
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
		url := "/api/ontology-query/v1/knowledge-networks/" + knID + "/subgraph"

		subgraphQuery := interfaces.SubGraphQueryBaseOnSource{
			SourceObjecTypeId: "ot1",
			Direction:         interfaces.DIRECTION_FORWARD,
			PathLength:        2,
			PageQuery: interfaces.PageQuery{
				Limit: 100,
			},
		}

		Convey("成功 - 参数验证通过，获取子图", func() {
			kns.EXPECT().SearchSubgraph(gomock.Any(), gomock.Any()).Return(interfaces.ObjectSubGraph{
				Objects:    map[string]interfaces.ObjectInfoInSubgraph{},
				TotalCount: 0,
			}, nil)

			reqParamByte, _ := sonic.Marshal(subgraphQuery)
			req := httptest.NewRequest(http.MethodPost, url+"?branch=main", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsSubgraph(c, visitor)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - includeLogicParams参数无效", func() {
			reqParamByte, _ := sonic.Marshal(subgraphQuery)
			req := httptest.NewRequest(http.MethodPost, url+"?include_logic_params=invalid", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsSubgraph(c, visitor)

			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("失败 - 起点对象类ID为空", func() {
			invalidQuery := interfaces.SubGraphQueryBaseOnSource{
				SourceObjecTypeId: "",
				Direction:         interfaces.DIRECTION_FORWARD,
				PathLength:        2,
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
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsSubgraph(c, visitor)

			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("失败 - Service返回错误", func() {
			kns.EXPECT().SearchSubgraph(gomock.Any(), gomock.Any()).Return(interfaces.ObjectSubGraph{}, rest.NewHTTPError(context.TODO(), http.StatusInternalServerError, oerrors.OntologyQuery_KnowledgeNetwork_InternalError_GetKnowledgeNetworksByIDFailed))

			reqParamByte, _ := sonic.Marshal(subgraphQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsSubgraph(c, visitor)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_RestHandler_GetObjectsSubgraphByTypePath(t *testing.T) {
	Convey("Test RestHandler GetObjectsSubgraphByTypePath", t, func() {
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
		url := "/api/ontology-query/v1/knowledge-networks/" + knID + "/subgraph"

		pathsQuery := interfaces.QueryRelationTypePaths{
			TypePaths: []interfaces.QueryRelationTypePath{
				{
					ObjectTypes: []interfaces.ObjectTypeWithKeyField{
						{OTID: "ot1"},
						{OTID: "ot2"},
					},
					Edges: []interfaces.TypeEdge{
						{
							RelationTypeId:     "rt1",
							SourceObjectTypeId: "ot1",
							TargetObjectTypeId: "ot2",
						},
					},
				},
			},
		}

		Convey("成功 - 参数验证通过，基于路径获取子图", func() {
			kns.EXPECT().SearchSubgraphByTypePath(gomock.Any(), gomock.Any()).Return(interfaces.PathsEntries{
				Entries: []interfaces.ObjectSubGraph{},
			}, nil)

			reqParamByte, _ := sonic.Marshal(pathsQuery)
			req := httptest.NewRequest(http.MethodPost, url+"?branch=main", bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsSubgraphByTypePath(c, visitor)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - 对象类型为空", func() {
			invalidQuery := interfaces.QueryRelationTypePaths{
				TypePaths: []interfaces.QueryRelationTypePath{
					{
						ObjectTypes: []interfaces.ObjectTypeWithKeyField{},
						Edges: []interfaces.TypeEdge{
							{
								RelationTypeId:     "rt1",
								SourceObjectTypeId: "ot1",
								TargetObjectTypeId: "ot2",
							},
						},
					},
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
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsSubgraphByTypePath(c, visitor)

			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("失败 - Service返回错误", func() {
			kns.EXPECT().SearchSubgraphByTypePath(gomock.Any(), gomock.Any()).Return(interfaces.PathsEntries{}, rest.NewHTTPError(context.TODO(), http.StatusInternalServerError, oerrors.OntologyQuery_KnowledgeNetwork_InternalError_GetKnowledgeNetworksByIDFailed))

			reqParamByte, _ := sonic.Marshal(pathsQuery)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			req.Header.Set(interfaces.HTTP_HEADER_METHOD_OVERRIDE, "GET")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "kn_id", Value: knID},
			}

			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			handler.GetObjectsSubgraphByTypePath(c, visitor)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
