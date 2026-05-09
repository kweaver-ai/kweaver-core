// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	cond "bkn-backend/common/condition"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

func TestNewObjectTypeTask(t *testing.T) {
	Convey("Test NewObjectTypeTask", t, func() {
		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				ViewDataLimit:    1000,
				JobMaxRetryTimes: 3,
			},
		}

		taskInfo := &interfaces.TaskInfo{
			ID:          "task1",
			ConceptID:   "ot1",
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
		}

		objectType := &interfaces.ObjectType{
			ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
				OTID: "ot1",
			},
		}

		task := NewObjectTypeTask(appSetting, taskInfo, objectType)

		Convey("Should create task with correct settings", func() {
			So(task, ShouldNotBeNil)
			So(task.ViewDataLimit, ShouldEqual, 1000)
			So(task.JobMaxRetryTimes, ShouldEqual, 3)
			So(task.taskInfo, ShouldEqual, taskInfo)
			So(task.objectType, ShouldEqual, objectType)
		})
	})
}

func TestObjectTypeTask_GetTaskInfo(t *testing.T) {
	Convey("Test GetTaskInfo", t, func() {
		taskInfo := &interfaces.TaskInfo{
			ID: "task1",
		}

		task := &ObjectTypeTask{
			taskInfo: taskInfo,
		}

		Convey("Should return task info", func() {
			info := task.GetTaskInfo()
			So(info, ShouldEqual, taskInfo)
		})
	})
}

func TestObjectTypeTask_HandleObjectTypeTask(t *testing.T) {
	Convey("Test HandleObjectTypeTask", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				ViewDataLimit:    100,
				JobMaxRetryTimes: 3,
			},
		}

		dva := bmock.NewMockDataViewAccess(mockCtrl)
		vba := bmock.NewMockVegaBackendAccess(mockCtrl)
		mfa := bmock.NewMockModelFactoryAccess(mockCtrl)
		ja := bmock.NewMockJobAccess(mockCtrl)
		osa := bmock.NewMockOpenSearchAccess(mockCtrl)

		jobInfo := &interfaces.JobInfo{
			ID:         "job1",
			KNID:       "kn1",
			Branch:     "main",
			JobType:    interfaces.JobTypeFull,
			CreateTime: time.Now().UnixMilli(),
		}

		taskInfo := &interfaces.TaskInfo{
			ID:          "task1",
			JobID:       "job1",
			ConceptID:   "ot1",
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
		}

		objectType := &interfaces.ObjectType{
			ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
				OTID:        "ot1",
				OTName:      "object_type1",
				PrimaryKeys: []string{"pk1"},
				DataSource: &interfaces.ResourceInfo{
					Type: "data_view",
					ID:   "dv1",
				},
				DataProperties: []*interfaces.DataProperty{
					{
						Name: "pk1",
						Type: "string",
						MappedField: &interfaces.Field{
							Name: "field1",
							Type: "string",
						},
					},
				},
			},
		}

		task := NewObjectTypeTask(appSetting, taskInfo, objectType)
		task.dva = dva
		task.vba = vba
		task.mfa = mfa
		task.ja = ja
		task.osa = osa

		Convey("Success handling object type task", func() {
			// 确保 objectType.Status 不为 nil（如果需要增量更新）
			objectType.Status = &interfaces.ObjectTypeStatus{}

			dataView := &interfaces.DataView{
				ViewID:   "dv1",
				ViewName: "test_view",
			}

			viewQueryResult := &interfaces.ViewQueryResult{
				TotalCount:  10,
				SearchAfter: nil,
				Entries: []map[string]any{
					{
						"field1": "value1",
					},
				},
			}

			// HandleObjectTypeTask 会先构建 propertyMapping，然后检查索引
			osa.EXPECT().IndexExists(ctx, gomock.Any()).Return(false, nil)
			osa.EXPECT().CreateIndex(ctx, gomock.Any(), gomock.Any()).Return(nil)

			dva.EXPECT().GetDataViewByID(ctx, "dv1").Return(dataView, nil)
			dva.EXPECT().GetDataStart(ctx, "dv1", "", nil, 100).Return(viewQueryResult, nil)
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(nil)
			osa.EXPECT().BulkInsertData(ctx, gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().Refresh(ctx, gomock.Any()).Return(nil)
			osa.EXPECT().GetIndexStats(ctx, gomock.Any()).Return(&interfaces.IndexStats{
				DocCount:    10,
				StorageSize: 1024,
			}, nil)

			err := task.HandleObjectTypeTask(ctx, jobInfo, taskInfo, objectType)
			So(err, ShouldBeNil)
		})

		Convey("Resource data source skips index build and returns nil", func() {
			resObjectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:        "ot1",
					OTName:      "object_type1",
					PrimaryKeys: []string{"pk1"},
					DataSource: &interfaces.ResourceInfo{
						Type: interfaces.DATA_SOURCE_TYPE_RESOURCE,
						ID:   "res1",
					},
					DataProperties: []*interfaces.DataProperty{
						{
							Name: "pk1",
							Type: "string",
							MappedField: &interfaces.Field{
								Name: "field1",
								Type: "string",
							},
						},
					},
				},
			}
			resTask := NewObjectTypeTask(appSetting, taskInfo, resObjectType)
			resTask.dva = dva
			resTask.vba = vba
			resTask.mfa = mfa
			resTask.ja = ja
			resTask.osa = osa

			err := resTask.HandleObjectTypeTask(ctx, jobInfo, taskInfo, resObjectType)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid data source type", func() {
			invalidObjectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					DataSource: &interfaces.ResourceInfo{
						Type: "invalid",
						ID:   "dv1",
					},
				},
			}
			invalidTask := NewObjectTypeTask(appSetting, taskInfo, invalidObjectType)
			invalidTask.dva = dva
			invalidTask.mfa = mfa
			invalidTask.ja = ja
			invalidTask.osa = osa

			err := invalidTask.HandleObjectTypeTask(ctx, jobInfo, taskInfo, invalidObjectType)
			So(err, ShouldBeNil) // Returns nil but logs warning
		})

		Convey("Failed with no primary keys", func() {
			noPKObjectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:        "ot1",
					PrimaryKeys: []string{},
					DataSource: &interfaces.ResourceInfo{
						Type: "data_view",
						ID:   "dv1",
					},
				},
			}
			noPKTask := NewObjectTypeTask(appSetting, taskInfo, noPKObjectType)
			noPKTask.dva = dva
			noPKTask.mfa = mfa
			noPKTask.ja = ja
			noPKTask.osa = osa

			err := noPKTask.HandleObjectTypeTask(ctx, jobInfo, taskInfo, noPKObjectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to get data view", func() {
			osa.EXPECT().IndexExists(ctx, gomock.Any()).Return(false, nil)
			osa.EXPECT().CreateIndex(ctx, gomock.Any(), gomock.Any()).Return(nil)
			dva.EXPECT().GetDataViewByID(ctx, "dv1").Return(nil, errors.New("db error"))

			err := task.HandleObjectTypeTask(ctx, jobInfo, taskInfo, objectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when primary key unmapped", func() {
			invalidObjectType := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:        "ot1",
					PrimaryKeys: []string{"pk1", "pk2"},
					DataSource: &interfaces.ResourceInfo{
						Type: "data_view",
						ID:   "dv1",
					},
					DataProperties: []*interfaces.DataProperty{
						{
							Name: "pk1",
							Type: "string",
							MappedField: &interfaces.Field{
								Name: "field1",
								Type: "string",
							},
						},
						// pk2 没有映射
					},
				},
			}
			invalidTask := NewObjectTypeTask(appSetting, taskInfo, invalidObjectType)
			invalidTask.dva = dva
			invalidTask.mfa = mfa
			invalidTask.ja = ja
			invalidTask.osa = osa

			err := invalidTask.HandleObjectTypeTask(ctx, jobInfo, taskInfo, invalidObjectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetModelByID returns error", func() {
			objectTypeWithVector := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:        "ot1",
					PrimaryKeys: []string{"pk1"},
					DataSource: &interfaces.ResourceInfo{
						Type: "data_view",
						ID:   "dv1",
					},
					DataProperties: []*interfaces.DataProperty{
						{
							Name: "pk1",
							Type: "string",
							MappedField: &interfaces.Field{
								Name: "field1",
								Type: "string",
							},
							IndexConfig: &interfaces.IndexConfig{
								VectorConfig: interfaces.VectorConfig{
									Enabled: true,
									ModelID: "model1",
								},
							},
						},
					},
				},
			}
			vectorTask := NewObjectTypeTask(appSetting, taskInfo, objectTypeWithVector)
			vectorTask.dva = dva
			vectorTask.mfa = mfa
			vectorTask.ja = ja
			vectorTask.osa = osa

			osa.EXPECT().IndexExists(gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
			osa.EXPECT().CreateIndex(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			mfa.EXPECT().GetModelByID(gomock.Any(), gomock.Any()).Return(nil, errors.New("model error")).AnyTimes()

			err := vectorTask.HandleObjectTypeTask(ctx, jobInfo, taskInfo, objectTypeWithVector)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetModelByID returns nil", func() {
			objectTypeWithVector := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:        "ot1",
					PrimaryKeys: []string{"pk1"},
					DataSource: &interfaces.ResourceInfo{
						Type: "data_view",
						ID:   "dv1",
					},
					DataProperties: []*interfaces.DataProperty{
						{
							Name: "pk1",
							Type: "string",
							MappedField: &interfaces.Field{
								Name: "field1",
								Type: "string",
							},
							IndexConfig: &interfaces.IndexConfig{
								VectorConfig: interfaces.VectorConfig{
									Enabled: true,
									ModelID: "model1",
								},
							},
						},
					},
				},
			}
			vectorTask := NewObjectTypeTask(appSetting, taskInfo, objectTypeWithVector)
			vectorTask.dva = dva
			vectorTask.mfa = mfa
			vectorTask.ja = ja
			vectorTask.osa = osa

			osa.EXPECT().IndexExists(gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
			osa.EXPECT().CreateIndex(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			mfa.EXPECT().GetModelByID(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

			err := vectorTask.HandleObjectTypeTask(ctx, jobInfo, taskInfo, objectTypeWithVector)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetDataStart returns error", func() {
			dataView := &interfaces.DataView{
				ViewID:   "dv1",
				ViewName: "test_view",
			}

			osa.EXPECT().IndexExists(ctx, gomock.Any()).Return(false, nil)
			osa.EXPECT().CreateIndex(ctx, gomock.Any(), gomock.Any()).Return(nil)
			dva.EXPECT().GetDataViewByID(ctx, "dv1").Return(dataView, nil)
			dva.EXPECT().GetDataStart(ctx, "dv1", "", nil, 100).Return(nil, errors.New("data view error"))

			err := task.HandleObjectTypeTask(ctx, jobInfo, taskInfo, objectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when UpdateTaskState returns error", func() {
			dataView := &interfaces.DataView{
				ViewID:   "dv1",
				ViewName: "test_view",
			}

			viewQueryResult := &interfaces.ViewQueryResult{
				TotalCount:  10,
				SearchAfter: nil,
				Entries: []map[string]any{
					{
						"field1": "value1",
					},
				},
			}

			osa.EXPECT().IndexExists(ctx, gomock.Any()).Return(false, nil)
			osa.EXPECT().CreateIndex(ctx, gomock.Any(), gomock.Any()).Return(nil)
			dva.EXPECT().GetDataViewByID(ctx, "dv1").Return(dataView, nil)
			dva.EXPECT().GetDataStart(ctx, "dv1", "", nil, 100).Return(viewQueryResult, nil)
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(errors.New("db error"))

			err := task.HandleObjectTypeTask(ctx, jobInfo, taskInfo, objectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when handlerIndexData returns error", func() {
			dataView := &interfaces.DataView{
				ViewID:   "dv1",
				ViewName: "test_view",
			}

			viewQueryResult := &interfaces.ViewQueryResult{
				TotalCount:  10,
				SearchAfter: nil,
				Entries: []map[string]any{
					{
						"field1": "value1",
					},
				},
			}

			osa.EXPECT().IndexExists(ctx, gomock.Any()).Return(false, nil)
			osa.EXPECT().CreateIndex(ctx, gomock.Any(), gomock.Any()).Return(nil)
			dva.EXPECT().GetDataViewByID(ctx, "dv1").Return(dataView, nil)
			dva.EXPECT().GetDataStart(ctx, "dv1", "", nil, 100).Return(viewQueryResult, nil)
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(nil)
			osa.EXPECT().BulkInsertData(ctx, gomock.Any(), gomock.Any()).Return(errors.New("opensearch error"))

			err := task.HandleObjectTypeTask(ctx, jobInfo, taskInfo, objectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetDataNext returns error", func() {
			dataView := &interfaces.DataView{
				ViewID:   "dv1",
				ViewName: "test_view",
			}

			viewQueryResult := &interfaces.ViewQueryResult{
				TotalCount:  10,
				SearchAfter: []interface{}{"value1"},
				Entries: []map[string]any{
					{
						"field1": "value1",
					},
				},
			}

			osa.EXPECT().IndexExists(ctx, gomock.Any()).Return(false, nil)
			osa.EXPECT().CreateIndex(ctx, gomock.Any(), gomock.Any()).Return(nil)
			dva.EXPECT().GetDataViewByID(ctx, "dv1").Return(dataView, nil)
			dva.EXPECT().GetDataStart(ctx, "dv1", "", nil, 100).Return(viewQueryResult, nil)
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(nil)
			osa.EXPECT().BulkInsertData(ctx, gomock.Any(), gomock.Any()).Return(nil)
			dva.EXPECT().GetDataNext(ctx, "dv1", viewQueryResult.SearchAfter, 100).Return(nil, errors.New("data view error"))

			err := task.HandleObjectTypeTask(ctx, jobInfo, taskInfo, objectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when Refresh returns error", func() {
			dataView := &interfaces.DataView{
				ViewID:   "dv1",
				ViewName: "test_view",
			}

			viewQueryResult := &interfaces.ViewQueryResult{
				TotalCount:  10,
				SearchAfter: nil,
				Entries: []map[string]any{
					{
						"field1": "value1",
					},
				},
			}

			osa.EXPECT().IndexExists(ctx, gomock.Any()).Return(false, nil)
			osa.EXPECT().CreateIndex(ctx, gomock.Any(), gomock.Any()).Return(nil)
			dva.EXPECT().GetDataViewByID(ctx, "dv1").Return(dataView, nil)
			dva.EXPECT().GetDataStart(ctx, "dv1", "", nil, 100).Return(viewQueryResult, nil)
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(nil)
			osa.EXPECT().BulkInsertData(ctx, gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().Refresh(ctx, gomock.Any()).Return(errors.New("opensearch error"))

			err := task.HandleObjectTypeTask(ctx, jobInfo, taskInfo, objectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed when GetIndexStats returns error", func() {
			dataView := &interfaces.DataView{
				ViewID:   "dv1",
				ViewName: "test_view",
			}

			viewQueryResult := &interfaces.ViewQueryResult{
				TotalCount:  10,
				SearchAfter: nil,
				Entries: []map[string]any{
					{
						"field1": "value1",
					},
				},
			}

			osa.EXPECT().IndexExists(ctx, gomock.Any()).Return(false, nil)
			osa.EXPECT().CreateIndex(ctx, gomock.Any(), gomock.Any()).Return(nil)
			dva.EXPECT().GetDataViewByID(ctx, "dv1").Return(dataView, nil)
			dva.EXPECT().GetDataStart(ctx, "dv1", "", nil, 100).Return(viewQueryResult, nil)
			ja.EXPECT().UpdateTaskState(ctx, "task1", gomock.Any()).Return(nil)
			osa.EXPECT().BulkInsertData(ctx, gomock.Any(), gomock.Any()).Return(nil)
			osa.EXPECT().Refresh(ctx, gomock.Any()).Return(nil)
			osa.EXPECT().GetIndexStats(ctx, gomock.Any()).Return(nil, errors.New("opensearch error"))

			err := task.HandleObjectTypeTask(ctx, jobInfo, taskInfo, objectType)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestObjectTypeTask_handlerIndex(t *testing.T) {
	Convey("Test handlerIndex", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osa := bmock.NewMockOpenSearchAccess(mockCtrl)

		task := &ObjectTypeTask{
			osa: osa,
			objectType: &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					DataProperties: []*interfaces.DataProperty{
						{
							Name: "prop1",
							Type: "string",
						},
					},
				},
			},
			vectorProperties: []*VectorProperty{},
		}

		Convey("Success creating new index", func() {
			osa.EXPECT().IndexExists(ctx, "test_index").Return(false, nil)
			osa.EXPECT().CreateIndex(ctx, "test_index", gomock.Any()).Return(nil)

			err := task.handlerIndex(ctx, "test_index", task.objectType)
			So(err, ShouldBeNil)
		})

		Convey("Success deleting existing index and creating new one", func() {
			osa.EXPECT().IndexExists(ctx, "test_index").Return(true, nil)
			osa.EXPECT().DeleteIndex(ctx, "test_index").Return(nil)
			osa.EXPECT().CreateIndex(ctx, "test_index", gomock.Any()).Return(nil)

			err := task.handlerIndex(ctx, "test_index", task.objectType)
			So(err, ShouldBeNil)
		})

		Convey("Failed to check index existence", func() {
			osa.EXPECT().IndexExists(ctx, "test_index").Return(false, errors.New("opensearch error"))

			err := task.handlerIndex(ctx, "test_index", task.objectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to delete existing index", func() {
			osa.EXPECT().IndexExists(ctx, "test_index").Return(true, nil)
			osa.EXPECT().DeleteIndex(ctx, "test_index").Return(errors.New("opensearch error"))

			err := task.handlerIndex(ctx, "test_index", task.objectType)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to create index", func() {
			osa.EXPECT().IndexExists(ctx, "test_index").Return(false, nil)
			osa.EXPECT().CreateIndex(ctx, "test_index", gomock.Any()).Return(errors.New("opensearch error"))

			err := task.handlerIndex(ctx, "test_index", task.objectType)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestObjectTypeTask_handlerIndexData(t *testing.T) {
	Convey("Test handlerIndexData", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		osa := bmock.NewMockOpenSearchAccess(mockCtrl)

		task := &ObjectTypeTask{
			osa: osa,
			objectTypeStatus: &interfaces.ObjectTypeStatus{
				Index: "test_index",
			},
			propertyMapping: map[string]*interfaces.Field{
				"prop1": {
					Name: "field1",
					Type: "string",
				},
			},
			objectType: &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					PrimaryKeys: []string{"prop1"},
				},
			},
			vectorProperties: []*VectorProperty{},
		}

		Convey("Success handling index data", func() {
			viewQueryResult := &interfaces.ViewQueryResult{
				Entries: []map[string]any{
					{
						"field1": "value1",
					},
				},
			}

			osa.EXPECT().BulkInsertData(ctx, "test_index", gomock.Any()).Return(nil)

			err := task.handlerIndexData(ctx, viewQueryResult)
			So(err, ShouldBeNil)
		})

		Convey("Success with empty entries", func() {
			viewQueryResult := &interfaces.ViewQueryResult{
				Entries: []map[string]any{},
			}

			err := task.handlerIndexData(ctx, viewQueryResult)
			So(err, ShouldBeNil)
		})

		Convey("Failed to bulk insert data", func() {
			viewQueryResult := &interfaces.ViewQueryResult{
				Entries: []map[string]any{
					{
						"field1": "value1",
					},
				},
			}

			osa.EXPECT().BulkInsertData(ctx, "test_index", gomock.Any()).Return(errors.New("opensearch error"))

			err := task.handlerIndexData(ctx, viewQueryResult)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with vector properties", func() {
			mfa := bmock.NewMockModelFactoryAccess(mockCtrl)
			task.mfa = mfa
			task.JobMaxRetryTimes = 3

			viewQueryResult := &interfaces.ViewQueryResult{
				Entries: []map[string]any{
					{
						"field1": "value1",
					},
				},
			}

			property := &VectorProperty{
				Name:           "prop1",
				VectorField:    "_vector_prop1",
				Model:          &interfaces.SmallModel{ModelID: "model1"},
				AllVectorResps: make([]*cond.VectorResp, 0),
			}
			task.vectorProperties = []*VectorProperty{property}

			vectorResps := []*cond.VectorResp{
				{Vector: []float32{0.1, 0.2}},
			}

			mfa.EXPECT().GetVector(ctx, property.Model, []string{"value1"}).Return(vectorResps, nil)
			osa.EXPECT().BulkInsertData(ctx, "test_index", gomock.Any()).Return(nil)

			err := task.handlerIndexData(ctx, viewQueryResult)
			So(err, ShouldBeNil)
		})

		Convey("Failed when handlerVector returns error after retries", func() {
			mfa := bmock.NewMockModelFactoryAccess(mockCtrl)
			task.mfa = mfa
			task.JobMaxRetryTimes = 3

			viewQueryResult := &interfaces.ViewQueryResult{
				Entries: []map[string]any{
					{
						"field1": "value1",
					},
				},
			}

			property := &VectorProperty{
				Name:           "prop1",
				VectorField:    "_vector_prop1",
				Model:          &interfaces.SmallModel{ModelID: "model1"},
				AllVectorResps: make([]*cond.VectorResp, 0),
			}
			task.vectorProperties = []*VectorProperty{property}

			mfa.EXPECT().GetVector(ctx, property.Model, []string{"value1"}).Return(nil, errors.New("vector error")).Times(3)

			err := task.handlerIndexData(ctx, viewQueryResult)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestObjectTypeTask_handlerVector(t *testing.T) {
	Convey("Test handlerVector", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mfa := bmock.NewMockModelFactoryAccess(mockCtrl)

		task := &ObjectTypeTask{
			mfa: mfa,
		}

		property := &VectorProperty{
			Name:           "prop1",
			VectorField:    "_vector_prop1",
			Model:          &interfaces.SmallModel{ModelID: "model1"},
			AllVectorResps: make([]*cond.VectorResp, 0),
		}

		Convey("Success handling vector", func() {
			newEntries := []any{
				map[string]any{
					"prop1": "text1",
				},
				map[string]any{
					"prop1": "text2",
				},
			}

			vectorResps := []*cond.VectorResp{
				{Vector: []float32{0.1, 0.2}},
				{Vector: []float32{0.3, 0.4}},
			}

			mfa.EXPECT().GetVector(ctx, property.Model, []string{"text1", "text2"}).Return(vectorResps, nil)

			err := task.handlerVector(ctx, property, newEntries)
			So(err, ShouldBeNil)
			So(len(property.AllVectorResps), ShouldEqual, 2)
		})

		Convey("Success with nil values", func() {
			newEntries := []any{
				map[string]any{
					"prop1": "text1",
				},
				map[string]any{
					"prop1": nil,
				},
			}

			vectorResps := []*cond.VectorResp{
				{Vector: []float32{0.1, 0.2}},
			}

			mfa.EXPECT().GetVector(ctx, property.Model, []string{"text1"}).Return(vectorResps, nil)

			err := task.handlerVector(ctx, property, newEntries)
			So(err, ShouldBeNil)
		})

		Convey("Failed to get vector", func() {
			newEntries := []any{
				map[string]any{
					"prop1": "text1",
				},
			}

			mfa.EXPECT().GetVector(ctx, property.Model, []string{"text1"}).Return(nil, errors.New("model error"))

			err := task.handlerVector(ctx, property, newEntries)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestObjectTypeTask_GetObjectID(t *testing.T) {
	Convey("Test GetObjectID", t, func() {
		task := &ObjectTypeTask{
			objectType: &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					PrimaryKeys: []string{"pk1", "pk2"},
				},
			},
			propertyMapping: map[string]*interfaces.Field{
				"pk1": {
					Name: "field1",
					Type: "string",
				},
				"pk2": {
					Name: "field2",
					Type: "string",
				},
			},
		}

		Convey("Success getting object ID", func() {
			objectData := map[string]any{
				"field1": "value1",
				"field2": "value2",
			}

			objectID := task.GetObjectID(objectData)
			So(objectID, ShouldNotBeEmpty)
			So(len(objectID), ShouldEqual, 32) // MD5 hash length
		})

		Convey("Success with missing field values", func() {
			objectData := map[string]any{
				"field1": "value1",
			}

			objectID := task.GetObjectID(objectData)
			So(objectID, ShouldNotBeEmpty)
		})
	})
}

func TestObjectTypeTask_generateTaskIndexName(t *testing.T) {
	Convey("Test generateTaskIndexName", t, func() {
		task := &ObjectTypeTask{}

		Convey("Should generate correct index name", func() {
			indexName := task.generateTaskIndexName("kn1", "main", "ot1", "task1")
			So(indexName, ShouldEqual, "adp-kn_ot_index-kn1-main-ot1-task1")
		})
	})
}
