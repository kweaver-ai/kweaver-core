// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knowledge_network

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	omock "ontology-query/interfaces/mock"
	"ontology-query/logics"
)

func Test_NewKnowledgeNetworkService(t *testing.T) {
	Convey("Test NewKnowledgeNetworkService", t, func() {
		appSetting := &common.AppSetting{}

		Convey("成功 - 创建服务实例", func() {
			service := NewKnowledgeNetworkService(appSetting)
			So(service, ShouldNotBeNil)
		})

		Convey("成功 - 单例模式", func() {
			service1 := NewKnowledgeNetworkService(appSetting)
			service2 := NewKnowledgeNetworkService(appSetting)
			So(service1, ShouldEqual, service2)
		})
	})
}

func Test_knowledgeNetworkService_SearchSubgraph(t *testing.T) {
	Convey("Test knowledgeNetworkService SearchSubgraph", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		omAccess := omock.NewMockOntologyManagerAccess(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)
		uAccess := omock.NewMockUniqueryAccess(mockCtrl)

		logics.OMA = omAccess
		logics.UA = uAccess

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			omAccess:   omAccess,
			ots:        ots,
			uAccess:    uAccess,
		}

		ctx := context.Background()
		knID := "kn1"
		branch := "main"
		sourceObjectTypeID := "ot1"

		Convey("失败 - 获取路径错误", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:              knID,
				Branch:            branch,
				SourceObjecTypeId: sourceObjectTypeID,
				Direction:         interfaces.DIRECTION_FORWARD,
				PathLength:        2,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetRelationTypePathsBaseOnSource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.SearchSubgraph(ctx, query)
			So(err, ShouldNotBeNil)
			So(result.RelationPaths, ShouldBeNil)
		})

		Convey("成功 - 查询子图", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:              knID,
				Branch:            branch,
				SourceObjecTypeId: sourceObjectTypeID,
				Direction:         interfaces.DIRECTION_FORWARD,
				PathLength:        2,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
				PathQuotaManager: &interfaces.PathQuotaManager{
					TotalLimit: 100,
				},
			}

			typePaths := []interfaces.RelationTypePath{
				{
					ID: 1,
					ObjectTypes: []interfaces.ObjectTypeWithKeyField{
						{OTID: sourceObjectTypeID},
						{OTID: "ot2"},
					},
					TypeEdges: []interfaces.TypeEdge{
						{
							RelationTypeId:     "rt1",
							SourceObjectTypeId: sourceObjectTypeID,
							TargetObjectTypeId: "ot2",
						},
					},
				},
			}

			startObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				TotalCount: 1,
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        sourceObjectTypeID,
						PrimaryKeys: []string{"id"},
					},
				},
			}

			omAccess.EXPECT().GetRelationTypePathsBaseOnSource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(typePaths, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(startObjects, nil)

			result, err := service.SearchSubgraph(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 1)
		})

		Convey("成功 - limit为0时使用默认值", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:              knID,
				Branch:            branch,
				SourceObjecTypeId: sourceObjectTypeID,
				Direction:         interfaces.DIRECTION_FORWARD,
				PathLength:        2,
				PageQuery: interfaces.PageQuery{
					Limit: 0,
				},
				PathQuotaManager: &interfaces.PathQuotaManager{
					TotalLimit: 100,
				},
			}

			typePaths := []interfaces.RelationTypePath{}
			startObjects := interfaces.Objects{
				Datas:      []map[string]any{},
				TotalCount: 0,
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: sourceObjectTypeID,
					},
				},
			}

			omAccess.EXPECT().GetRelationTypePathsBaseOnSource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(typePaths, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, q *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
				So(q.Limit, ShouldEqual, interfaces.DEFAULT_LIMIT)
				return startObjects, nil
			})

			result, err := service.SearchSubgraph(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
		})

		Convey("失败 - 获取起点对象错误", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:              knID,
				Branch:            branch,
				SourceObjecTypeId: sourceObjectTypeID,
				Direction:         interfaces.DIRECTION_FORWARD,
				PathLength:        2,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			typePaths := []interfaces.RelationTypePath{
				{
					ID: 1,
					ObjectTypes: []interfaces.ObjectTypeWithKeyField{
						{OTID: sourceObjectTypeID},
					},
				},
			}

			omAccess.EXPECT().GetRelationTypePathsBaseOnSource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(typePaths, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.SearchSubgraph(ctx, query)
			So(err, ShouldNotBeNil)
			So(result.RelationPaths, ShouldBeNil)
		})

		Convey("失败 - 构建对象子图错误", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:              knID,
				Branch:            branch,
				SourceObjecTypeId: sourceObjectTypeID,
				Direction:         interfaces.DIRECTION_FORWARD,
				PathLength:        2,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
				PathQuotaManager: &interfaces.PathQuotaManager{
					TotalLimit:         100,
					GlobalCount:        0,
					UsedQuota:          sync.Map{},
					RequestPathTypeNum: 1,
				},
			}

			typePaths := []interfaces.RelationTypePath{
				{
					ID: 1,
					ObjectTypes: []interfaces.ObjectTypeWithKeyField{
						{OTID: sourceObjectTypeID},
						{OTID: "ot2"},
					},
					TypeEdges: []interfaces.TypeEdge{
						{
							RelationTypeId:     "rt1",
							SourceObjectTypeId: sourceObjectTypeID,
							TargetObjectTypeId: "ot2",
							Direction:          interfaces.DIRECTION_FORWARD,
							RelationType: interfaces.RelationType{
								SourceObjectTypeID: sourceObjectTypeID,
								TargetObjectTypeID: "ot2",
								MappingRules: []interfaces.Mapping{
									{
										SourceProp: interfaces.SimpleProperty{Name: "id"},
										TargetProp: interfaces.SimpleProperty{Name: "target_id"},
									},
								},
							},
						},
					},
				},
			}

			startObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				TotalCount: 1,
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        sourceObjectTypeID,
						PrimaryKeys: []string{"id"},
					},
				},
			}

			omAccess.EXPECT().GetRelationTypePathsBaseOnSource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(typePaths, nil)

			// 第一次调用：获取起点对象（在 SearchSubgraph 中）
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, q *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
				if q.ObjectTypeID == sourceObjectTypeID {
					// 获取起点对象
					return startObjects, nil
				} else if q.ObjectTypeID == "ot2" {
					// 在 buildObjectSubgraph -> getNextObjectsBatchByRelation 中获取下一层对象，返回错误
					return interfaces.Objects{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError)
				}
				return interfaces.Objects{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError)
			}).AnyTimes()

			result, err := service.SearchSubgraph(ctx, query)
			So(err, ShouldNotBeNil)
			So(result.RelationPaths, ShouldBeNil)
		})
	})
}

func Test_knowledgeNetworkService_SearchSubgraphByTypePath(t *testing.T) {
	Convey("Test knowledgeNetworkService SearchSubgraphByTypePath", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		omAccess := omock.NewMockOntologyManagerAccess(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)
		uAccess := omock.NewMockUniqueryAccess(mockCtrl)

		logics.OMA = omAccess
		logics.UA = uAccess

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			omAccess:   omAccess,
			ots:        ots,
			uAccess:    uAccess,
		}

		ctx := context.Background()
		knID := "kn1"
		branch := "main"

		Convey("成功 - 查询路径子图", func() {
			query := &interfaces.SubGraphQueryBaseOnTypePath{
				KNID:   knID,
				Branch: branch,
				Paths: interfaces.QueryRelationTypePaths{
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
				},
			}

			relationType := interfaces.RelationType{
				RTID:               "rt1",
				RTName:             "relation1",
				SourceObjectTypeID: "ot1",
				TargetObjectTypeID: "ot2",
			}

			startObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				TotalCount: 1,
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						PrimaryKeys: []string{"id"},
					},
				},
			}

			omAccess.EXPECT().GetRelationType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(relationType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(startObjects, nil)

			result, err := service.SearchSubgraphByTypePath(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Entries), ShouldEqual, 1)
		})

		Convey("失败 - 关系类不存在", func() {
			query := &interfaces.SubGraphQueryBaseOnTypePath{
				KNID:   knID,
				Branch: branch,
				Paths: interfaces.QueryRelationTypePaths{
					TypePaths: []interfaces.QueryRelationTypePath{
						{
							Edges: []interfaces.TypeEdge{
								{
									RelationTypeId: "rt1",
								},
							},
						},
					},
				},
			}

			omAccess.EXPECT().GetRelationType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.RelationType{}, false, nil)

			result, err := service.SearchSubgraphByTypePath(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
			So(result.Entries, ShouldBeNil)
		})

		Convey("失败 - 获取关系类错误", func() {
			query := &interfaces.SubGraphQueryBaseOnTypePath{
				KNID:   knID,
				Branch: branch,
				Paths: interfaces.QueryRelationTypePaths{
					TypePaths: []interfaces.QueryRelationTypePath{
						{
							Edges: []interfaces.TypeEdge{
								{
									RelationTypeId: "rt1",
								},
							},
						},
					},
				},
			}

			omAccess.EXPECT().GetRelationType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.RelationType{}, false, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.SearchSubgraphByTypePath(ctx, query)
			So(err, ShouldNotBeNil)
			So(result.Entries, ShouldBeNil)
		})

		Convey("失败 - 获取起点对象错误", func() {
			query := &interfaces.SubGraphQueryBaseOnTypePath{
				KNID:   knID,
				Branch: branch,
				Paths: interfaces.QueryRelationTypePaths{
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
				},
			}

			relationType := interfaces.RelationType{
				RTID:               "rt1",
				RTName:             "relation1",
				SourceObjectTypeID: "ot1",
				TargetObjectTypeID: "ot2",
			}

			omAccess.EXPECT().GetRelationType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(relationType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.SearchSubgraphByTypePath(ctx, query)
			So(err, ShouldNotBeNil)
			So(result.Entries, ShouldBeNil)
		})

		Convey("成功 - 反向方向的关系", func() {
			query := &interfaces.SubGraphQueryBaseOnTypePath{
				KNID:   knID,
				Branch: branch,
				Paths: interfaces.QueryRelationTypePaths{
					TypePaths: []interfaces.QueryRelationTypePath{
						{
							ObjectTypes: []interfaces.ObjectTypeWithKeyField{
								{OTID: "ot1"},
								{OTID: "ot2"},
							},
							Edges: []interfaces.TypeEdge{
								{
									RelationTypeId:     "rt1",
									SourceObjectTypeId: "ot2", // 反向：路径的源是关系类的目标
									TargetObjectTypeId: "ot1",
								},
							},
						},
					},
				},
			}

			relationType := interfaces.RelationType{
				RTID:               "rt1",
				RTName:             "relation1",
				SourceObjectTypeID: "ot1", // 关系类的源是 ot1
				TargetObjectTypeID: "ot2", // 关系类的目标是 ot2
			}

			startObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				TotalCount: 1,
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot2",
						PrimaryKeys: []string{"id"},
					},
				},
			}

			omAccess.EXPECT().GetRelationType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(relationType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(startObjects, nil)

			result, err := service.SearchSubgraphByTypePath(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Entries), ShouldEqual, 1)
		})
	})
}

func Test_knowledgeNetworkService_isPathEndsWith(t *testing.T) {
	Convey("Test knowledgeNetworkService isPathEndsWith", t, func() {
		service := &knowledgeNetworkService{}

		Convey("成功 - 路径以指定对象ID结尾", func() {
			path := interfaces.RelationPath{
				Relations: []interfaces.Relation{
					{
						SourceObjectId: "obj1",
						TargetObjectId: "obj2",
					},
				},
			}

			result := service.isPathEndsWith(path, "obj2")
			So(result, ShouldBeTrue)
		})

		Convey("失败 - 路径不以指定对象ID结尾", func() {
			path := interfaces.RelationPath{
				Relations: []interfaces.Relation{
					{
						SourceObjectId: "obj1",
						TargetObjectId: "obj2",
					},
				},
			}

			result := service.isPathEndsWith(path, "obj3")
			So(result, ShouldBeFalse)
		})

		Convey("成功 - 空路径", func() {
			path := interfaces.RelationPath{
				Relations: []interfaces.Relation{},
			}

			result := service.isPathEndsWith(path, "obj1")
			So(result, ShouldBeTrue) // 空路径返回true
		})
	})
}

func Test_knowledgeNetworkService_extendPathsWithNewEdge(t *testing.T) {
	Convey("Test knowledgeNetworkService extendPathsWithNewEdge", t, func() {
		service := &knowledgeNetworkService{}

		Convey("成功 - 扩展路径", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				BatchQueryState: interfaces.BatchQueryState{
					Visited: make(map[string]bool),
				},
			}

			paths := []interfaces.RelationPath{
				{
					Relations: []interfaces.Relation{
						{
							SourceObjectId: "obj1",
							TargetObjectId: "obj2",
						},
					},
					Length: 1,
				},
			}

			edge := interfaces.TypeEdge{
				RelationTypeId: "rt1",
				RelationType: interfaces.RelationType{
					RTName: "relation1",
				},
			}

			newPaths, pathExisted := service.extendPathsWithNewEdge(query, paths, "obj2", "obj3", edge)
			So(len(newPaths), ShouldEqual, 1)
			So(newPaths[0].Length, ShouldEqual, 2)
			So(len(newPaths[0].Relations), ShouldEqual, 2)
			So(pathExisted, ShouldBeFalse)
		})

		Convey("成功 - 路径不匹配", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				BatchQueryState: interfaces.BatchQueryState{
					Visited: make(map[string]bool),
				},
			}

			paths := []interfaces.RelationPath{
				{
					Relations: []interfaces.Relation{
						{
							SourceObjectId: "obj1",
							TargetObjectId: "obj2",
						},
					},
					Length: 1,
				},
			}

			edge := interfaces.TypeEdge{
				RelationTypeId: "rt1",
				RelationType: interfaces.RelationType{
					RTName: "relation1",
				},
			}

			newPaths, pathExisted := service.extendPathsWithNewEdge(query, paths, "obj999", "obj3", edge)
			So(len(newPaths), ShouldEqual, 0)
			So(pathExisted, ShouldBeFalse)
		})

		Convey("成功 - 检测到重复路径", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				BatchQueryState: interfaces.BatchQueryState{
					Visited: make(map[string]bool),
				},
			}

			paths := []interfaces.RelationPath{
				{
					Relations: []interfaces.Relation{
						{
							RelationTypeId: "rt1",
							SourceObjectId: "obj1",
							TargetObjectId: "obj2",
						},
					},
					Length: 1,
				},
			}

			edge := interfaces.TypeEdge{
				RelationTypeId: "rt1",
				RelationType: interfaces.RelationType{
					RTName: "relation1",
				},
			}

			// 先添加一次路径
			newPaths1, pathExisted1 := service.extendPathsWithNewEdge(query, paths, "obj2", "obj3", edge)
			So(len(newPaths1), ShouldEqual, 1)
			So(pathExisted1, ShouldBeFalse)

			// 再次添加相同路径，应该检测到重复
			newPaths2, pathExisted2 := service.extendPathsWithNewEdge(query, paths, "obj2", "obj3", edge)
			So(len(newPaths2), ShouldEqual, 1)
			So(pathExisted2, ShouldBeTrue)
		})
	})
}

func Test_knowledgeNetworkService_isObjectRelated(t *testing.T) {
	Convey("Test knowledgeNetworkService isObjectRelated", t, func() {
		service := &knowledgeNetworkService{}

		Convey("成功 - 直接映射匹配", func() {
			currentObjectData := map[string]any{
				"id": "123",
			}
			nextObject := map[string]any{
				"target_id": "123",
			}
			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "id"},
							TargetProp: interfaces.SimpleProperty{Name: "target_id"},
						},
					},
				},
			}

			result := service.isObjectRelated(currentObjectData, nextObject, edge, true, nil)
			So(result, ShouldBeTrue)
		})

		Convey("失败 - 直接映射不匹配", func() {
			currentObjectData := map[string]any{
				"id": "123",
			}
			nextObject := map[string]any{
				"target_id": "456",
			}
			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "id"},
							TargetProp: interfaces.SimpleProperty{Name: "target_id"},
						},
					},
				},
			}

			result := service.isObjectRelated(currentObjectData, nextObject, edge, true, nil)
			So(result, ShouldBeFalse)
		})

		Convey("成功 - 间接映射匹配", func() {
			currentObjectData := map[string]any{
				"id": "123",
			}
			nextObject := map[string]any{
				"target_id": "456",
			}
			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					MappingRules: &interfaces.InDirectMapping{
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "id"},
								TargetProp: interfaces.SimpleProperty{Name: "view_id"},
							},
						},
						TargetMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "view_target_id"},
								TargetProp: interfaces.SimpleProperty{Name: "target_id"},
							},
						},
					},
				},
			}
			viewData := []map[string]any{
				{
					"view_id":        "123",
					"view_target_id": "456",
				},
			}

			result := service.isObjectRelated(currentObjectData, nextObject, edge, true, viewData)
			So(result, ShouldBeTrue)
		})
	})
}

func Test_knowledgeNetworkService_mapResultsToObjects(t *testing.T) {
	Convey("Test knowledgeNetworkService mapResultsToObjects", t, func() {
		service := &knowledgeNetworkService{}

		Convey("成功 - 映射结果到对象", func() {
			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			nextObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"target_id": "123", "name": "test"},
					{"target_id": "456", "name": "test2"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: "ot2",
					},
				},
			}

			result := make(map[string]interfaces.Objects)
			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "id"},
							TargetProp: interfaces.SimpleProperty{Name: "target_id"},
						},
					},
				},
			}

			service.mapResultsToObjects(currentLevelObjects, nextObjects, result, edge, true, nil)
			So(len(result), ShouldEqual, 1)
			So(result["obj1"], ShouldNotBeNil)
			So(len(result["obj1"].Datas), ShouldEqual, 1)
		})
	})
}

func Test_knowledgeNetworkService_buildObjectSubgraph(t *testing.T) {
	Convey("Test knowledgeNetworkService buildObjectSubgraph", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ots := omock.NewMockObjectTypeService(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			ots:        ots,
		}

		ctx := context.Background()
		knID := "kn1"
		sourceObjectTypeID := "ot1"

		Convey("成功 - 构建对象子图", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:              knID,
				SourceObjecTypeId: sourceObjectTypeID,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
				PathQuotaManager: &interfaces.PathQuotaManager{
					TotalLimit:         100,
					GlobalCount:        0,
					RequestPathTypeNum: 1,
				},
			}

			typePaths := []interfaces.RelationTypePath{
				{
					ID: 1,
					ObjectTypes: []interfaces.ObjectTypeWithKeyField{
						{OTID: sourceObjectTypeID},
						{OTID: "ot2"},
					},
					TypeEdges: []interfaces.TypeEdge{
						{
							RelationTypeId:     "rt1",
							SourceObjectTypeId: sourceObjectTypeID,
							TargetObjectTypeId: "ot2",
							RelationType: interfaces.RelationType{
								MappingRules: []interfaces.Mapping{
									{
										SourceProp: interfaces.SimpleProperty{Name: "id"},
										TargetProp: interfaces.SimpleProperty{Name: "target_id"},
									},
								},
							},
						},
					},
				},
			}

			startObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        sourceObjectTypeID,
						PrimaryKeys: []string{"id"},
					},
				},
			}

			nextObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"target_id": "123", "name": "test"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot2",
						PrimaryKeys: []string{"target_id"},
					},
				},
			}

			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(nextObjects, nil)

			result, err := service.buildObjectSubgraph(ctx, query, typePaths, startObjects)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Objects, ShouldNotBeNil)
		})

		Convey("成功 - 空路径列表", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:              knID,
				SourceObjecTypeId: sourceObjectTypeID,
				PathQuotaManager: &interfaces.PathQuotaManager{
					TotalLimit:         100,
					GlobalCount:        0,
					RequestPathTypeNum: 0,
				},
			}

			typePaths := []interfaces.RelationTypePath{}
			startObjects := interfaces.Objects{
				Datas: []map[string]any{},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: sourceObjectTypeID,
					},
				},
			}

			result, err := service.buildObjectSubgraph(ctx, query, typePaths, startObjects)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.RelationPaths), ShouldEqual, 0)
		})

		Convey("失败 - 获取下一层对象错误", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:              knID,
				SourceObjecTypeId: sourceObjectTypeID,
				PathQuotaManager: &interfaces.PathQuotaManager{
					TotalLimit:         100,
					GlobalCount:        0,
					RequestPathTypeNum: 1,
				},
			}

			typePaths := []interfaces.RelationTypePath{
				{
					ID: 1,
					ObjectTypes: []interfaces.ObjectTypeWithKeyField{
						{OTID: sourceObjectTypeID},
						{OTID: "ot2"},
					},
					TypeEdges: []interfaces.TypeEdge{
						{
							RelationTypeId:     "rt1",
							SourceObjectTypeId: sourceObjectTypeID,
							TargetObjectTypeId: "ot2",
							RelationType: interfaces.RelationType{
								MappingRules: []interfaces.Mapping{
									{
										SourceProp: interfaces.SimpleProperty{Name: "id"},
										TargetProp: interfaces.SimpleProperty{Name: "target_id"},
									},
								},
							},
						},
					},
				},
			}

			startObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        sourceObjectTypeID,
						PrimaryKeys: []string{"id"},
					},
				},
			}

			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.buildObjectSubgraph(ctx, query, typePaths, startObjects)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func Test_knowledgeNetworkService_expandObjectPathsBatch(t *testing.T) {
	Convey("Test knowledgeNetworkService expandObjectPathsBatch", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ots := omock.NewMockObjectTypeService(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			ots:        ots,
		}

		ctx := context.Background()
		knID := "kn1"
		sourceObjectTypeID := "ot1"

		Convey("成功 - 扩展对象路径", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:              knID,
				SourceObjecTypeId: sourceObjectTypeID,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
				PathQuotaManager: &interfaces.PathQuotaManager{
					TotalLimit:         100,
					GlobalCount:        0,
					RequestPathTypeNum: 1,
				},
				BatchQueryState: interfaces.BatchQueryState{
					BatchSize: 50,
					Visited:   make(map[string]bool),
				},
			}

			typePath := interfaces.RelationTypePath{
				ID: 1,
				ObjectTypes: []interfaces.ObjectTypeWithKeyField{
					{OTID: sourceObjectTypeID},
					{OTID: "ot2"},
				},
				TypeEdges: []interfaces.TypeEdge{
					{
						RelationTypeId:     "rt1",
						SourceObjectTypeId: sourceObjectTypeID,
						TargetObjectTypeId: "ot2",
						Direction:          interfaces.DIRECTION_FORWARD,
						RelationType: interfaces.RelationType{
							MappingRules: []interfaces.Mapping{
								{
									SourceProp: interfaces.SimpleProperty{Name: "id"},
									TargetProp: interfaces.SimpleProperty{Name: "target_id"},
								},
							},
						},
					},
				},
			}

			startObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        sourceObjectTypeID,
						PrimaryKeys: []string{"id"},
					},
				},
			}

			nextObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"target_id": "123", "name": "test"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot2",
						PrimaryKeys: []string{"target_id"},
					},
				},
			}

			objectsMap := make(map[string]interfaces.ObjectInfoInSubgraph)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(nextObjects, nil)

			result, err := service.expandObjectPathsBatch(ctx, query, typePath, startObjects, objectsMap)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("成功 - 达到路径终点", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:              knID,
				SourceObjecTypeId: sourceObjectTypeID,
				PathQuotaManager: &interfaces.PathQuotaManager{
					TotalLimit:         100,
					GlobalCount:        0,
					RequestPathTypeNum: 1,
				},
				BatchQueryState: interfaces.BatchQueryState{
					BatchSize: 50,
				},
			}

			typePath := interfaces.RelationTypePath{
				ID: 1,
				ObjectTypes: []interfaces.ObjectTypeWithKeyField{
					{OTID: sourceObjectTypeID},
				},
				TypeEdges: []interfaces.TypeEdge{},
			}

			startObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        sourceObjectTypeID,
						PrimaryKeys: []string{"id"},
					},
				},
			}

			objectsMap := make(map[string]interfaces.ObjectInfoInSubgraph)

			result, err := service.expandObjectPathsBatch(ctx, query, typePath, startObjects, objectsMap)
			So(err, ShouldBeNil)
			So(len(result), ShouldBeGreaterThanOrEqualTo, 0)
		})

		Convey("成功 - 达到配额限制", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:              knID,
				SourceObjecTypeId: sourceObjectTypeID,
				PathQuotaManager: &interfaces.PathQuotaManager{
					TotalLimit:         0,
					GlobalCount:        0,
					RequestPathTypeNum: 1,
				},
				BatchQueryState: interfaces.BatchQueryState{
					BatchSize: 50,
				},
			}

			typePath := interfaces.RelationTypePath{
				ID: 1,
				ObjectTypes: []interfaces.ObjectTypeWithKeyField{
					{OTID: sourceObjectTypeID},
					{OTID: "ot2"},
				},
				TypeEdges: []interfaces.TypeEdge{
					{
						RelationTypeId:     "rt1",
						SourceObjectTypeId: sourceObjectTypeID,
						TargetObjectTypeId: "ot2",
						Direction:          interfaces.DIRECTION_FORWARD,
						RelationType: interfaces.RelationType{
							MappingRules: []interfaces.Mapping{
								{
									SourceProp: interfaces.SimpleProperty{Name: "id"},
									TargetProp: interfaces.SimpleProperty{Name: "target_id"},
								},
							},
						},
					},
				},
			}

			startObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        sourceObjectTypeID,
						PrimaryKeys: []string{"id"},
					},
				},
			}

			objectsMap := make(map[string]interfaces.ObjectInfoInSubgraph)

			result, err := service.expandObjectPathsBatch(ctx, query, typePath, startObjects, objectsMap)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})
	})
}

func Test_knowledgeNetworkService_getNextObjectsBatchByRelation(t *testing.T) {
	Convey("Test knowledgeNetworkService getNextObjectsBatchByRelation", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		ots := omock.NewMockObjectTypeService(mockCtrl)

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			ots:        ots,
		}

		ctx := context.Background()
		knID := "kn1"
		branch := "main"

		Convey("成功 - 正向关系获取下一层对象", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:   knID,
				Branch: branch,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			batch := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationTypeId:     "rt1",
				SourceObjectTypeId: "ot1",
				TargetObjectTypeId: "ot2",
				Direction:          interfaces.DIRECTION_FORWARD,
				RelationType: interfaces.RelationType{
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "id"},
							TargetProp: interfaces.SimpleProperty{Name: "target_id"},
						},
					},
				},
			}

			objectType := interfaces.ObjectTypeWithKeyField{
				OTID: "ot2",
			}

			nextObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"target_id": "123", "name": "test"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot2",
						PrimaryKeys: []string{"target_id"},
					},
				},
			}

			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(nextObjects, nil)

			result, err := service.getNextObjectsBatchByRelation(ctx, query, batch, edge, objectType)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result), ShouldEqual, 1)
		})

		Convey("成功 - 反向关系获取下一层对象-1", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:   knID,
				Branch: branch,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			batch := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationTypeId:     "rt1",
				SourceObjectTypeId: "ot2",
				TargetObjectTypeId: "ot1",
				Direction:          interfaces.DIRECTION_BACKWARD,
				RelationType: interfaces.RelationType{
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "id"},
							TargetProp: interfaces.SimpleProperty{Name: "target_id"},
						},
					},
				},
			}

			objectType := interfaces.ObjectTypeWithKeyField{
				OTID: "ot1",
			}

			nextObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123", "name": "test"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:        "ot1",
						PrimaryKeys: []string{"id"},
					},
				},
			}

			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(nextObjects, nil)

			result, err := service.getNextObjectsBatchByRelation(ctx, query, batch, edge, objectType)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("成功 - 无查询条件", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:   knID,
				Branch: branch,
			}

			batch := []interfaces.LevelObject{
				{
					ObjectID:   "obj1",
					ObjectData: map[string]any{},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationTypeId:     "rt1",
				SourceObjectTypeId: "ot1",
				TargetObjectTypeId: "ot2",
				Direction:          interfaces.DIRECTION_FORWARD,
				RelationType: interfaces.RelationType{
					MappingRules: []interfaces.Mapping{},
				},
			}

			objectType := interfaces.ObjectTypeWithKeyField{
				OTID: "ot2",
			}

			result, err := service.getNextObjectsBatchByRelation(ctx, query, batch, edge, objectType)
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("失败 - 获取对象错误", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:   knID,
				Branch: branch,
			}

			batch := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationTypeId:     "rt1",
				SourceObjectTypeId: "ot1",
				TargetObjectTypeId: "ot2",
				Direction:          interfaces.DIRECTION_FORWARD,
				RelationType: interfaces.RelationType{
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "id"},
							TargetProp: interfaces.SimpleProperty{Name: "target_id"},
						},
					},
				},
			}

			objectType := interfaces.ObjectTypeWithKeyField{
				OTID: "ot2",
			}

			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.getNextObjectsBatchByRelation(ctx, query, batch, edge, objectType)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功 - 反向关系获取下一层对象", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:   knID,
				Branch: branch,
			}

			batch := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationTypeId:     "rt1",
				SourceObjectTypeId: "ot2",
				TargetObjectTypeId: "ot1",
				Direction:          interfaces.DIRECTION_BACKWARD,
				RelationType: interfaces.RelationType{
					SourceObjectTypeID: "ot1",
					TargetObjectTypeID: "ot2",
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "id"},
							TargetProp: interfaces.SimpleProperty{Name: "target_id"},
						},
					},
				},
			}

			objectType := interfaces.ObjectTypeWithKeyField{
				OTID: "ot1",
			}

			nextObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"target_id": "123"},
				},
				TotalCount: 1,
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: "ot1",
					},
				},
			}

			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, q *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
				So(q.ObjectTypeID, ShouldEqual, "ot1")
				return nextObjects, nil
			})

			result, err := service.getNextObjectsBatchByRelation(ctx, query, batch, edge, objectType)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("成功 - 多个条件和对象类过滤条件", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID:   knID,
				Branch: branch,
			}

			batch := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
				{
					ObjectID: "obj2",
					ObjectData: map[string]any{
						"id": "456",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationTypeId:     "rt1",
				SourceObjectTypeId: "ot1",
				TargetObjectTypeId: "ot2",
				Direction:          interfaces.DIRECTION_FORWARD,
				RelationType: interfaces.RelationType{
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "id"},
							TargetProp: interfaces.SimpleProperty{Name: "target_id"},
						},
					},
				},
			}

			objectType := interfaces.ObjectTypeWithKeyField{
				OTID: "ot2",
				ActualCondition: &cond.CondCfg{
					Name:      "status",
					Operation: "==",
					ValueOptCfg: cond.ValueOptCfg{
						Value: "active",
					},
				},
				PageQuery: interfaces.PageQuery{
					Sort: []*interfaces.SortParams{
						{
							Field:     "name",
							Direction: interfaces.ASC_DIRECTION,
						},
					},
					Limit: 100,
				},
			}

			nextObjects := interfaces.Objects{
				Datas: []map[string]any{
					{"target_id": "123", "status": "active"},
					{"target_id": "456", "status": "active"},
				},
				TotalCount: 2,
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: "ot2",
					},
				},
			}

			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, q *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
				So(q.ActualCondition, ShouldNotBeNil)
				So(q.Sort, ShouldNotBeNil)
				So(q.Limit, ShouldEqual, 100)
				return nextObjects, nil
			})

			result, err := service.getNextObjectsBatchByRelation(ctx, query, batch, edge, objectType)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})
	})
}

func Test_knowledgeNetworkService_buildBatchConditions(t *testing.T) {
	Convey("Test knowledgeNetworkService buildBatchConditions", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		uAccess := omock.NewMockUniqueryAccess(mockCtrl)

		logics.UA = uAccess

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			uAccess:    uAccess,
		}

		ctx := context.Background()
		knID := "kn1"

		Convey("成功 - 直接映射构建条件", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "id"},
							TargetProp: interfaces.SimpleProperty{Name: "target_id"},
						},
					},
				},
			}

			conditions, viewDataMap, err := service.buildBatchConditions(ctx, query, currentLevelObjects, edge, true)
			So(err, ShouldBeNil)
			So(conditions, ShouldNotBeNil)
			So(len(conditions), ShouldBeGreaterThan, 0)
			So(viewDataMap, ShouldNotBeNil)
		})

		Convey("成功 - 间接映射构建条件", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					MappingRules: &interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							ID: "view1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "id"},
								TargetProp: interfaces.SimpleProperty{Name: "view_id"},
							},
						},
						TargetMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "view_target_id"},
								TargetProp: interfaces.SimpleProperty{Name: "target_id"},
							},
						},
					},
				},
			}

			viewData := interfaces.ViewData{
				Datas: []map[string]any{
					{
						"view_id":        "123",
						"view_target_id": "456",
					},
				},
			}

			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(viewData, nil)

			conditions, viewDataMap, err := service.buildBatchConditions(ctx, query, currentLevelObjects, edge, true)
			So(err, ShouldBeNil)
			So(conditions, ShouldNotBeNil)
			So(viewDataMap, ShouldNotBeNil)
		})

		Convey("成功 - 混合映射", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
				{
					ObjectID: "obj2",
					ObjectData: map[string]any{
						"id": "456",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					MappingRules: []interfaces.Mapping{
						{
							SourceProp: interfaces.SimpleProperty{Name: "id"},
							TargetProp: interfaces.SimpleProperty{Name: "target_id"},
						},
					},
				},
			}

			conditions, viewDataMap, err := service.buildBatchConditions(ctx, query, currentLevelObjects, edge, true)
			So(err, ShouldBeNil)
			So(conditions, ShouldNotBeNil)
			So(viewDataMap, ShouldNotBeNil)
		})
	})
}

func Test_knowledgeNetworkService_buildIndirectBatchConditions(t *testing.T) {
	Convey("Test knowledgeNetworkService buildIndirectBatchConditions", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		uAccess := omock.NewMockUniqueryAccess(mockCtrl)

		logics.UA = uAccess

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			uAccess:    uAccess,
		}

		ctx := context.Background()
		knID := "kn1"

		Convey("失败 - 视图ID为空", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					RTName: "relation1",
					MappingRules: &interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							ID: "",
						},
					},
				},
			}

			conditions, viewDataMap, err := service.buildIndirectBatchConditions(ctx, query, currentLevelObjects, edge, true)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(conditions, ShouldBeNil)
			So(viewDataMap, ShouldBeNil)
		})

		Convey("成功 - 构建间接映射条件", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					RTName: "relation1",
					MappingRules: &interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							ID: "view1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "id"},
								TargetProp: interfaces.SimpleProperty{Name: "view_id"},
							},
						},
						TargetMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "view_target_id"},
								TargetProp: interfaces.SimpleProperty{Name: "target_id"},
							},
						},
					},
				},
			}

			viewData := interfaces.ViewData{
				Datas: []map[string]any{
					{
						"view_id":        "123",
						"view_target_id": "456",
					},
				},
			}

			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(viewData, nil)

			conditions, viewDataMap, err := service.buildIndirectBatchConditions(ctx, query, currentLevelObjects, edge, true)
			So(err, ShouldBeNil)
			So(conditions, ShouldNotBeNil)
			So(viewDataMap, ShouldNotBeNil)
			So(len(viewDataMap), ShouldBeGreaterThan, 0)
		})

		Convey("失败 - 获取视图数据错误", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					RTName: "relation1",
					MappingRules: &interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							ID: "view1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "id"},
								TargetProp: interfaces.SimpleProperty{Name: "view_id"},
							},
						},
						TargetMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "view_target_id"},
								TargetProp: interfaces.SimpleProperty{Name: "target_id"},
							},
						},
					},
				},
			}

			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(interfaces.ViewData{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			conditions, viewDataMap, err := service.buildIndirectBatchConditions(ctx, query, currentLevelObjects, edge, true)
			So(err, ShouldNotBeNil)
			So(conditions, ShouldBeNil)
			So(viewDataMap, ShouldBeNil)
		})

		Convey("成功 - 反向映射构建条件", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"target_id": "456", // 反向映射时，batchGetViewData 使用 TargetMappingRules，需要 target_id 字段
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					RTName: "relation1",
					MappingRules: &interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							ID: "view1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "id"},
								TargetProp: interfaces.SimpleProperty{Name: "view_id"},
							},
						},
						TargetMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "view_target_id"},
								TargetProp: interfaces.SimpleProperty{Name: "target_id"},
							},
						},
					},
				},
			}

			viewData := interfaces.ViewData{
				Datas: []map[string]any{
					{
						"view_id":        "123",
						"view_target_id": "456",
					},
				},
			}

			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).DoAndReturn(func(ctx context.Context, viewID string, viewQuery interfaces.ViewQuery) (interfaces.ViewData, error) {
				// 验证查询条件：反向映射时使用 TargetMappingRules，查询 view_target_id = "456"
				So(viewQuery.Filters, ShouldNotBeNil)
				return viewData, nil
			})

			conditions, viewDataMap, err := service.buildIndirectBatchConditions(ctx, query, currentLevelObjects, edge, false)
			So(err, ShouldBeNil)
			So(conditions, ShouldNotBeNil)
			So(len(conditions), ShouldBeGreaterThan, 0)
			So(viewDataMap, ShouldNotBeNil)
			So(len(viewDataMap), ShouldBeGreaterThan, 0)
			// 验证视图数据映射：反向映射时使用 SourceMappingRules 构建条件，从视图数据中提取 view_id
			So(len(viewDataMap["obj1"]), ShouldEqual, 1)
		})

		Convey("成功 - 空视图数据", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					RTName: "relation1",
					MappingRules: &interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							ID: "view1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "id"},
								TargetProp: interfaces.SimpleProperty{Name: "view_id"},
							},
						},
						TargetMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "view_target_id"},
								TargetProp: interfaces.SimpleProperty{Name: "target_id"},
							},
						},
					},
				},
			}

			viewData := interfaces.ViewData{
				Datas: []map[string]any{},
			}

			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(viewData, nil)

			conditions, viewDataMap, err := service.buildIndirectBatchConditions(ctx, query, currentLevelObjects, edge, true)
			So(err, ShouldBeNil)
			So(len(conditions), ShouldEqual, 0)
			So(len(viewDataMap), ShouldEqual, 0)
		})

		Convey("成功 - 单个映射规则且多个值使用in操作", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
				{
					ObjectID: "obj2",
					ObjectData: map[string]any{
						"id": "456",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					RTName: "relation1",
					MappingRules: &interfaces.InDirectMapping{
						BackingDataSource: &interfaces.ResourceInfo{
							ID: "view1",
						},
						SourceMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "id"},
								TargetProp: interfaces.SimpleProperty{Name: "view_id"},
							},
						},
						TargetMappingRules: []interfaces.Mapping{
							{
								SourceProp: interfaces.SimpleProperty{Name: "view_target_id"},
								TargetProp: interfaces.SimpleProperty{Name: "target_id"},
							},
						},
					},
				},
			}

			viewData := interfaces.ViewData{
				Datas: []map[string]any{
					{"view_id": "123", "view_target_id": "789"},
					{"view_id": "456", "view_target_id": "790"},
				},
			}

			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(viewData, nil)

			conditions, viewDataMap, err := service.buildIndirectBatchConditions(ctx, query, currentLevelObjects, edge, true)
			So(err, ShouldBeNil)
			So(conditions, ShouldNotBeNil)
			So(len(conditions), ShouldEqual, 1)
			So(conditions[0].Operation, ShouldEqual, "in")
			So(viewDataMap, ShouldNotBeNil)
		})
	})
}

func Test_knowledgeNetworkService_batchGetViewData(t *testing.T) {
	Convey("Test knowledgeNetworkService batchGetViewData", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		uAccess := omock.NewMockUniqueryAccess(mockCtrl)

		logics.UA = uAccess

		service := &knowledgeNetworkService{
			appSetting: appSetting,
			uAccess:    uAccess,
		}

		ctx := context.Background()
		knID := "kn1"

		Convey("成功 - 批量获取视图数据", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					RTName: "relation1",
				},
			}

			mappingRules := &interfaces.InDirectMapping{
				BackingDataSource: &interfaces.ResourceInfo{
					ID: "view1",
				},
				SourceMappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "id"},
						TargetProp: interfaces.SimpleProperty{Name: "view_id"},
					},
				},
				TargetMappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "view_target_id"},
						TargetProp: interfaces.SimpleProperty{Name: "target_id"},
					},
				},
			}

			viewData := interfaces.ViewData{
				Datas: []map[string]any{
					{
						"view_id":        "123",
						"view_target_id": "456",
					},
				},
			}

			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(viewData, nil)

			result, err := service.batchGetViewData(ctx, query, edge, currentLevelObjects, mappingRules, true)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result), ShouldBeGreaterThan, 0)
		})

		Convey("失败 - 获取视图数据错误", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					RTName: "relation1",
				},
			}

			mappingRules := &interfaces.InDirectMapping{
				BackingDataSource: &interfaces.ResourceInfo{
					ID: "view1",
				},
				SourceMappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "id"},
						TargetProp: interfaces.SimpleProperty{Name: "view_id"},
					},
				},
			}

			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(interfaces.ViewData{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.batchGetViewData(ctx, query, edge, currentLevelObjects, mappingRules, true)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功 - 批量处理多个对象", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			// 创建超过批次大小的对象列表（批次大小为50）
			currentLevelObjects := make([]interfaces.LevelObject, 60)
			for i := 0; i < 60; i++ {
				currentLevelObjects[i] = interfaces.LevelObject{
					ObjectID: fmt.Sprintf("obj%d", i),
					ObjectData: map[string]any{
						"id": fmt.Sprintf("%d", i),
					},
				}
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					RTName: "relation1",
				},
			}

			mappingRules := &interfaces.InDirectMapping{
				BackingDataSource: &interfaces.ResourceInfo{
					ID: "view1",
				},
				SourceMappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "id"},
						TargetProp: interfaces.SimpleProperty{Name: "view_id"},
					},
				},
				TargetMappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "view_target_id"},
						TargetProp: interfaces.SimpleProperty{Name: "target_id"},
					},
				},
			}

			viewData := interfaces.ViewData{
				Datas: []map[string]any{
					{"view_id": "0", "view_target_id": "100"},
					{"view_id": "1", "view_target_id": "101"},
				},
			}

			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(viewData, nil).Times(2)

			result, err := service.batchGetViewData(ctx, query, edge, currentLevelObjects, mappingRules, true)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("成功 - 单个映射规则使用in操作", func() {
			query := &interfaces.SubGraphQueryBaseOnSource{
				KNID: knID,
			}

			currentLevelObjects := []interfaces.LevelObject{
				{
					ObjectID: "obj1",
					ObjectData: map[string]any{
						"id": "123",
					},
				},
				{
					ObjectID: "obj2",
					ObjectData: map[string]any{
						"id": "456",
					},
				},
			}

			edge := &interfaces.TypeEdge{
				RelationType: interfaces.RelationType{
					RTName: "relation1",
				},
			}

			mappingRules := &interfaces.InDirectMapping{
				BackingDataSource: &interfaces.ResourceInfo{
					ID: "view1",
				},
				SourceMappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "id"},
						TargetProp: interfaces.SimpleProperty{Name: "view_id"},
					},
				},
			}

			viewData := interfaces.ViewData{
				Datas: []map[string]any{
					{"view_id": "123"},
					{"view_id": "456"},
				},
			}

			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).DoAndReturn(func(ctx context.Context, viewID string, query interfaces.ViewQuery) (interfaces.ViewData, error) {
				So(query.Filters, ShouldNotBeNil)
				So(query.Filters.Operation, ShouldEqual, "in")
				return viewData, nil
			})

			result, err := service.batchGetViewData(ctx, query, edge, currentLevelObjects, mappingRules, true)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})
	})
}

func Test_knowledgeNetworkService_mapViewDataToObjects(t *testing.T) {
	Convey("Test knowledgeNetworkService mapViewDataToObjects", t, func() {
		service := &knowledgeNetworkService{}

		Convey("成功 - 映射视图数据到对象", func() {
			viewData := []map[string]any{
				{
					"view_id":        "123",
					"view_target_id": "456",
				},
			}

			batchConditions := []*cond.CondCfg{
				{
					Name:      "view_id",
					Operation: "==",
					ValueOptCfg: cond.ValueOptCfg{
						Value: "123",
					},
				},
			}

			objectMapping := map[int]string{
				0: "obj1",
			}

			mappingRules := &interfaces.InDirectMapping{
				SourceMappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "id"},
						TargetProp: interfaces.SimpleProperty{Name: "view_id"},
					},
				},
			}

			result := make(map[string][]map[string]any)

			service.mapViewDataToObjects(viewData, batchConditions, objectMapping, mappingRules, true, result)
			So(len(result), ShouldBeGreaterThan, 0)
		})

		Convey("成功 - 空视图数据", func() {
			viewData := []map[string]any{}

			batchConditions := []*cond.CondCfg{}
			objectMapping := map[int]string{}
			mappingRules := &interfaces.InDirectMapping{}

			result := make(map[string][]map[string]any)

			service.mapViewDataToObjects(viewData, batchConditions, objectMapping, mappingRules, true, result)
			So(len(result), ShouldEqual, 0)
		})

		Convey("成功 - 条件索引越界", func() {
			viewData := []map[string]any{
				{
					"view_id":        "123",
					"view_target_id": "456",
				},
			}

			batchConditions := []*cond.CondCfg{
				{
					Name:      "view_id",
					Operation: "==",
					ValueOptCfg: cond.ValueOptCfg{
						Value: "123",
					},
				},
			}

			objectMapping := map[int]string{
				0: "obj1",
				5: "obj2", // 索引5超出batchConditions长度
			}

			mappingRules := &interfaces.InDirectMapping{
				SourceMappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "id"},
						TargetProp: interfaces.SimpleProperty{Name: "view_id"},
					},
				},
			}

			result := make(map[string][]map[string]any)

			service.mapViewDataToObjects(viewData, batchConditions, objectMapping, mappingRules, true, result)
			So(len(result), ShouldBeGreaterThan, 0)
		})

		Convey("成功 - 反向映射规则", func() {
			viewData := []map[string]any{
				{
					"id":             "123", // 反向映射时，CheckViewDataMatchesCondition 使用 SourceProp.Name，即 "id"
					"view_id":        "123",
					"view_target_id": "456",
				},
			}

			batchConditions := []*cond.CondCfg{
				{
					Name:      "id", // 条件字段名应该对应 SourceProp.Name
					Operation: "==",
					ValueOptCfg: cond.ValueOptCfg{
						Value: "123",
					},
				},
			}

			objectMapping := map[int]string{
				0: "obj1",
			}

			mappingRules := &interfaces.InDirectMapping{
				TargetMappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "id"}, // 反向映射时使用 SourceProp.Name
						TargetProp: interfaces.SimpleProperty{Name: "view_id"},
					},
				},
			}

			result := make(map[string][]map[string]any)

			service.mapViewDataToObjects(viewData, batchConditions, objectMapping, mappingRules, false, result)
			So(len(result), ShouldBeGreaterThan, 0)
			So(len(result["obj1"]), ShouldEqual, 1)
		})
	})
}
