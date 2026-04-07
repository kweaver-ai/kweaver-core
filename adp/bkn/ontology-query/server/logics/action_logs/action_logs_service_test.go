// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_logs

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"ontology-query/interfaces"
)

func Test_structToMap(t *testing.T) {
	Convey("Test structToMap", t, func() {
		Convey("should convert struct to map", func() {
			exec := &interfaces.ActionExecution{
				ID:             "exec_123",
				KNID:           "kn_001",
				ActionTypeID:   "at_001",
				ActionTypeName: "restart_pod",
				Status:         interfaces.ExecutionStatusPending,
				TotalCount:     2,
			}

			result := structToMap(exec)

			So(result["id"], ShouldEqual, "exec_123")
			So(result["kn_id"], ShouldEqual, "kn_001")
			So(result["action_type_id"], ShouldEqual, "at_001")
			So(result["action_type_name"], ShouldEqual, "restart_pod")
			So(result["status"], ShouldEqual, "pending")
			So(result["total_count"], ShouldEqual, float64(2)) // JSON numbers are float64
		})

		Convey("should handle empty struct", func() {
			exec := &interfaces.ActionExecution{}
			result := structToMap(exec)

			So(result, ShouldNotBeNil)
			So(result["id"], ShouldEqual, "")
		})
	})
}

func Test_mapToActionExecution(t *testing.T) {
	Convey("Test mapToActionExecution", t, func() {
		Convey("should convert map to ActionExecution", func() {
			m := map[string]any{
				"id":               "exec_123",
				"kn_id":            "kn_001",
				"action_type_id":   "at_001",
				"action_type_name": "restart_pod",
				"status":           "completed",
				"total_count":      float64(2),
				"success_count":    float64(1),
				"failed_count":     float64(1),
			}

			exec, err := mapToActionExecution(m)

			So(err, ShouldBeNil)
			So(exec.ID, ShouldEqual, "exec_123")
			So(exec.KNID, ShouldEqual, "kn_001")
			So(exec.ActionTypeID, ShouldEqual, "at_001")
			So(exec.ActionTypeName, ShouldEqual, "restart_pod")
			So(exec.Status, ShouldEqual, "completed")
			So(exec.TotalCount, ShouldEqual, 2)
			So(exec.SuccessCount, ShouldEqual, 1)
			So(exec.FailedCount, ShouldEqual, 1)
		})

		Convey("should handle results array", func() {
			m := map[string]any{
				"id":     "exec_123",
				"status": "completed",
				"results": []any{
					map[string]any{
						"_instance_identity": map[string]any{"pod_ip": "192.168.1.1"},
						"status":             "success",
						"duration_ms":        float64(1200),
					},
				},
			}

			exec, err := mapToActionExecution(m)

			So(err, ShouldBeNil)
			So(len(exec.Results), ShouldEqual, 1)
			So(exec.Results[0].Status, ShouldEqual, "success")
			So(exec.Results[0].DurationMs, ShouldEqual, 1200)
		})

		Convey("should handle empty map", func() {
			m := map[string]any{}

			exec, err := mapToActionExecution(m)

			So(err, ShouldBeNil)
			So(exec, ShouldNotBeNil)
			So(exec.ID, ShouldEqual, "")
		})
	})
}

func Test_GetActionExecutionIndex(t *testing.T) {
	Convey("Test GetActionExecutionIndex", t, func() {
		Convey("should return correct index name", func() {
			index := interfaces.GetActionExecutionIndex("kn_001")
			So(index, ShouldEqual, "ontology_action_executions_kn_001")
		})

		Convey("should handle empty kn_id", func() {
			index := interfaces.GetActionExecutionIndex("")
			So(index, ShouldEqual, "ontology_action_executions_")
		})
	})
}

func Test_ActionLogQuery_Defaults(t *testing.T) {
	Convey("Test ActionLogQuery defaults", t, func() {
		Convey("should have default values", func() {
			query := &interfaces.ActionLogQuery{
				KNID: "kn_001",
			}

			So(query.Limit, ShouldEqual, 0) // Will be set to 20 in service
			So(query.NeedTotal, ShouldEqual, false)
			So(query.SearchAfter, ShouldBeNil)
		})
	})
}

func Test_ActionExecution_Status_Constants(t *testing.T) {
	Convey("Test execution status constants", t, func() {
		Convey("should have correct status values", func() {
			So(interfaces.ExecutionStatusPending, ShouldEqual, "pending")
			So(interfaces.ExecutionStatusRunning, ShouldEqual, "running")
			So(interfaces.ExecutionStatusCompleted, ShouldEqual, "completed")
			So(interfaces.ExecutionStatusFailed, ShouldEqual, "failed")
		})

		Convey("should have correct object status values", func() {
			So(interfaces.ObjectStatusPending, ShouldEqual, "pending")
			So(interfaces.ObjectStatusSuccess, ShouldEqual, "success")
			So(interfaces.ObjectStatusFailed, ShouldEqual, "failed")
		})

		Convey("should have correct trigger type values", func() {
			So(interfaces.TriggerTypeManual, ShouldEqual, "manual")
			So(interfaces.TriggerTypeScheduled, ShouldEqual, "scheduled")
		})

		Convey("should have correct action source type values", func() {
			So(interfaces.ActionSourceTypeTool, ShouldEqual, "tool")
			So(interfaces.ActionSourceTypeMCP, ShouldEqual, "mcp")
		})
	})
}

// // Mock OpenSearch access for integration tests
// type mockOpenSearchAccess struct {
// 	data       map[string]map[string]any
// 	indexExist map[string]bool
// }

// func newMockOpenSearchAccess() *mockOpenSearchAccess {
// 	return &mockOpenSearchAccess{
// 		data:       make(map[string]map[string]any),
// 		indexExist: make(map[string]bool),
// 	}
// }

// func (m *mockOpenSearchAccess) IndexExists(ctx context.Context, indexName string) (bool, error) {
// 	return m.indexExist[indexName], nil
// }

// func (m *mockOpenSearchAccess) CreateIndex(ctx context.Context, indexName string, body any) error {
// 	m.indexExist[indexName] = true
// 	return nil
// }

// func (m *mockOpenSearchAccess) InsertData(ctx context.Context, indexName, id string, data any) error {
// 	if m.data[indexName] == nil {
// 		m.data[indexName] = make(map[string]any)
// 	}
// 	m.data[indexName][id] = data
// 	return nil
// }

// func (m *mockOpenSearchAccess) SearchData(ctx context.Context, indexName string, query any) ([]interfaces.Hit, error) {
// 	return nil, nil
// }

// func (m *mockOpenSearchAccess) Count(ctx context.Context, indexName string, query any) ([]byte, error) {
// 	return []byte(`{"count": 0}`), nil
// }

// func (m *mockOpenSearchAccess) DeleteByQuery(ctx context.Context, indexName string, query any) error {
// 	return nil
// }
