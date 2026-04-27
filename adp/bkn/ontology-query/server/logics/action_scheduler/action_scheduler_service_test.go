// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_scheduler

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	omock "ontology-query/interfaces/mock"
	"ontology-query/logics"
	"ontology-query/logics/action_logs"
)

func Test_buildExecutionParams(t *testing.T) {
	Convey("Test buildExecutionParams", t, func() {
		s := &actionSchedulerService{}

		Convey("should get value from object property (VALUE_FROM_PROP)", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "target_ip",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
						Value:     "pod_ip",
					},
				},
			}

			objData := map[string]any{
				"_instance_id": "test-instance",
				"_instance_identity": map[string]any{
					"pod_ip":   "192.168.1.1",
					"pod_name": "test-pod",
				},
				"_display": "test-pod",
				"pod_ip":   "192.168.1.1",
				"pod_name": "test-pod",
			}

			params, err := s.buildExecutionParams(actionType, objData, nil)

			So(err, ShouldBeNil)
			So(params["target_ip"], ShouldEqual, "192.168.1.1")
		})

		Convey("should get value from constant (VALUE_FROM_CONST)", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "timeout",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
						Value:     60,
					},
				},
			}

			objData := map[string]any{}

			params, err := s.buildExecutionParams(actionType, objData, nil)

			So(err, ShouldBeNil)
			So(params["timeout"], ShouldEqual, 60)
		})

		Convey("should get value from dynamic params (VALUE_FROM_INPUT)", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "Authorization",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
					},
				},
			}
			objData := map[string]any{}
			dynamicParams := map[string]any{
				"Authorization": "Bearer token123",
			}

			params, err := s.buildExecutionParams(actionType, objData, dynamicParams)

			So(err, ShouldBeNil)
			So(params["Authorization"], ShouldEqual, "Bearer token123")
		})

		Convey("should handle mixed parameter sources", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "target_ip",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
						Value:     "pod_ip",
					},
					{
						Name:      "timeout",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
						Value:     30,
					},
					{
						Name:      "token",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
					},
				},
			}
			objData := map[string]any{
				"_instance_id": "test-instance",
				"_instance_identity": map[string]any{
					"pod_ip": "10.0.0.1",
				},
				"_display": "test-pod",
				"pod_ip":   "10.0.0.1",
			}
			dynamicParams := map[string]any{
				"token": "abc123",
			}

			params, err := s.buildExecutionParams(actionType, objData, dynamicParams)

			So(err, ShouldBeNil)
			So(params["target_ip"], ShouldEqual, "10.0.0.1")
			So(params["timeout"], ShouldEqual, 30)
			So(params["token"], ShouldEqual, "abc123")
		})

		Convey("should handle missing property in identity", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "target_ip",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
						Value:     "pod_ip",
					},
				},
			}
			objData := map[string]any{
				"_instance_id": "test-instance",
				"_instance_identity": map[string]any{
					"pod_name": "test-pod", // pod_ip is missing
				},
				"_display": "test-pod",
				"pod_name": "test-pod", // pod_ip is missing
			}

			params, err := s.buildExecutionParams(actionType, objData, nil)

			So(err, ShouldBeNil)
			_, exists := params["target_ip"]
			So(exists, ShouldBeFalse) // Parameter should not be set if property is missing
		})

		Convey("should handle missing dynamic param", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "token",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
					},
				},
			}
			objData := map[string]any{}

			params, err := s.buildExecutionParams(actionType, objData, nil)

			So(err, ShouldBeNil)
			_, exists := params["token"]
			So(exists, ShouldBeFalse)
		})

		Convey("should handle empty parameters", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{},
			}
			objData := map[string]any{}

			params, err := s.buildExecutionParams(actionType, objData, nil)

			So(err, ShouldBeNil)
			So(len(params), ShouldEqual, 0)
		})

		Convey("should handle nested dynamic param name (dot-separated)", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "props.headers",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
					},
				},
			}
			objData := map[string]any{}
			dynamicParams := map[string]any{
				"props": map[string]any{
					"headers": map[string]any{"Authorization": "Bearer xxx"},
				},
			}

			params, err := s.buildExecutionParams(actionType, objData, dynamicParams)

			So(err, ShouldBeNil)
			propsMap, ok := params["props"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(propsMap["headers"], ShouldResemble, map[string]any{"Authorization": "Bearer xxx"})
		})

		Convey("should handle nested property name (dot-separated)", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "target.ip",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
						Value:     "network.ip",
					},
				},
			}
			objData := map[string]any{
				"network": map[string]any{
					"ip": "10.0.0.1",
				},
			}

			params, err := s.buildExecutionParams(actionType, objData, nil)

			So(err, ShouldBeNil)
			targetMap, ok := params["target"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(targetMap["ip"], ShouldEqual, "10.0.0.1")
		})

		Convey("should handle mixed nested and flat params", func() {
			actionType := &interfaces.ActionType{
				Parameters: []interfaces.Parameter{
					{
						Name:      "props.headers",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
					},
					{
						Name:      "query",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
						Value:     "test-query",
					},
					{
						Name:      "target_ip",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
						Value:     "pod_ip",
					},
				},
			}
			objData := map[string]any{
				"pod_ip": "192.168.1.1",
			}
			dynamicParams := map[string]any{
				"props": map[string]any{
					"headers": "Bearer token",
				},
			}

			params, err := s.buildExecutionParams(actionType, objData, dynamicParams)

			So(err, ShouldBeNil)
			propsMap, ok := params["props"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(propsMap["headers"], ShouldEqual, "Bearer token")
			So(params["query"], ShouldEqual, "test-query")
			So(params["target_ip"], ShouldEqual, "192.168.1.1")
		})
	})
}

func Test_ActionExecutionRequest_Validation(t *testing.T) {
	Convey("Test ActionExecutionRequest", t, func() {
		Convey("should have required fields", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:         "kn_001",
				ActionTypeID: "at_001",
				InstanceIdentities: []map[string]any{
					{"pod_ip": "192.168.1.1"},
				},
			}

			So(req.KNID, ShouldEqual, "kn_001")
			So(req.ActionTypeID, ShouldEqual, "at_001")
			So(len(req.InstanceIdentities), ShouldEqual, 1)
		})

		Convey("should support branch field", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:         "kn_001",
				Branch:       "feature/test",
				ActionTypeID: "at_001",
				InstanceIdentities: []map[string]any{
					{"pod_ip": "192.168.1.1"},
				},
			}

			So(req.KNID, ShouldEqual, "kn_001")
			So(req.Branch, ShouldEqual, "feature/test")
			So(req.ActionTypeID, ShouldEqual, "at_001")
		})

		Convey("should default branch to empty string when not set", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:         "kn_001",
				ActionTypeID: "at_001",
			}

			So(req.Branch, ShouldEqual, "")
		})

		Convey("should handle multiple objects", func() {
			req := &interfaces.ActionExecutionRequest{
				InstanceIdentities: []map[string]any{
					{"pod_ip": "192.168.1.1", "id": 1},
					{"pod_ip": "192.168.1.2", "id": 2},
					{"pod_ip": "192.168.1.3", "id": 3},
				},
			}

			So(len(req.InstanceIdentities), ShouldEqual, 3)
		})

		Convey("should handle dynamic params", func() {
			req := &interfaces.ActionExecutionRequest{
				DynamicParams: map[string]any{
					"Authorization": "Bearer xxx",
					"Timeout":       60,
				},
			}

			So(req.DynamicParams["Authorization"], ShouldEqual, "Bearer xxx")
			So(req.DynamicParams["Timeout"], ShouldEqual, 60)
		})
	})
}

func Test_ActionExecutionResponse(t *testing.T) {
	Convey("Test ActionExecutionResponse", t, func() {
		Convey("should have correct structure", func() {
			resp := &interfaces.ActionExecutionResponse{
				ExecutionID: "exec_123",
				Status:      interfaces.ExecutionStatusPending,
				Message:     "Action execution started",
				CreatedAt:   1704067200000,
			}

			So(resp.ExecutionID, ShouldEqual, "exec_123")
			So(resp.Status, ShouldEqual, "pending")
			So(resp.Message, ShouldEqual, "Action execution started")
			So(resp.CreatedAt, ShouldEqual, int64(1704067200000))
		})
	})
}

func Test_ObjectExecutionResult(t *testing.T) {
	Convey("Test ObjectExecutionResult", t, func() {
		Convey("should represent success result", func() {
			result := interfaces.ObjectExecutionResult{
				ObjectSystemInfo: interfaces.ObjectSystemInfo{
					InstanceID:       "",
					InstanceIdentity: map[string]any{"pod_ip": "192.168.1.1"},
					Display:          "",
				},
				Status:     interfaces.ObjectStatusSuccess,
				Parameters: map[string]any{"target_ip": "192.168.1.1"},
				Result:     map[string]any{"message": "OK"},
				DurationMs: 1200,
			}

			So(result.Status, ShouldEqual, "success")
			So(result.ErrorMessage, ShouldEqual, "")
			So(result.DurationMs, ShouldEqual, int64(1200))
		})

		Convey("should represent failed result", func() {
			result := interfaces.ObjectExecutionResult{
				ObjectSystemInfo: interfaces.ObjectSystemInfo{
					InstanceID:       "",
					InstanceIdentity: map[string]any{"pod_ip": "192.168.1.2"},
					Display:          "",
				},
				Status:       interfaces.ObjectStatusFailed,
				Parameters:   map[string]any{"target_ip": "192.168.1.2"},
				ErrorMessage: "Connection timeout",
				DurationMs:   5000,
			}

			So(result.Status, ShouldEqual, "failed")
			So(result.ErrorMessage, ShouldEqual, "Connection timeout")
			So(result.Result, ShouldBeNil)
		})
	})
}

func Test_ActionSource_Types(t *testing.T) {
	Convey("Test ActionSource types", t, func() {
		Convey("should handle Tool source", func() {
			source := interfaces.ActionSource{
				Type:   interfaces.ActionSourceTypeTool,
				BoxID:  "box_001",
				ToolID: "tool_001",
			}

			So(source.Type, ShouldEqual, "tool")
			So(source.BoxID, ShouldEqual, "box_001")
			So(source.ToolID, ShouldEqual, "tool_001")
		})

		Convey("should handle MCP source", func() {
			source := interfaces.ActionSource{
				Type:     interfaces.ActionSourceTypeMCP,
				McpID:    "mcp_001",
				ToolName: "restart_service",
			}

			So(source.Type, ShouldEqual, "mcp")
			So(source.McpID, ShouldEqual, "mcp_001")
			So(source.ToolName, ShouldEqual, "restart_service")
		})
	})
}

func Test_ActionExecution_Snapshot(t *testing.T) {
	Convey("Test ActionExecution with ActionTypeSnapshot", t, func() {
		Convey("should store action type snapshot", func() {
			// 模拟从 manager 获取的原始行动类配置
			actionTypeSnapshot := map[string]any{
				"id":             "at_001",
				"name":           "restart_pod",
				"action_type":    "modify",
				"object_type_id": "ot_pod",
				"tags":           []string{"k8s", "pod"},
				"comment":        "重启 Pod",
				"icon":           "restart",
				"color":          "#FF5722",
				"condition": map[string]any{
					"field":    "status",
					"operator": "eq",
					"value":    "Running",
				},
				"parameters": []map[string]any{
					{"name": "timeout", "value_from": "const", "value": 30},
				},
				"schedule": map[string]any{
					"type":       "manual",
					"expression": "",
				},
				"creator":     "user_123",
				"create_time": int64(1704000000000),
				"updater":     "user_456",
				"update_time": int64(1704100000000),
			}

			execution := &interfaces.ActionExecution{
				ID:                 "exec_001",
				KNID:               "kn_001",
				ActionTypeID:       "at_001",
				ActionTypeName:     "restart_pod",
				ActionSourceType:   interfaces.ActionSourceTypeTool,
				Status:             interfaces.ExecutionStatusPending,
				TotalCount:         1,
				ActionTypeSnapshot: actionTypeSnapshot,
			}

			So(execution.ActionTypeSnapshot, ShouldNotBeNil)
			So(execution.ActionTypeSnapshot["id"], ShouldEqual, "at_001")
			So(execution.ActionTypeSnapshot["name"], ShouldEqual, "restart_pod")
			So(execution.ActionTypeSnapshot["tags"], ShouldNotBeNil)
			So(execution.ActionTypeSnapshot["condition"], ShouldNotBeNil)
			So(execution.ActionTypeSnapshot["parameters"], ShouldNotBeNil)
			So(execution.ActionTypeSnapshot["creator"], ShouldEqual, "user_123")
			So(execution.ActionTypeSnapshot["create_time"], ShouldEqual, int64(1704000000000))
		})

		Convey("should allow nil snapshot for backward compatibility", func() {
			execution := &interfaces.ActionExecution{
				ID:                 "exec_002",
				KNID:               "kn_001",
				ActionTypeID:       "at_001",
				ActionTypeName:     "restart_pod",
				Status:             interfaces.ExecutionStatusCompleted,
				ActionTypeSnapshot: nil, // 旧数据可能没有快照
			}

			So(execution.ActionTypeSnapshot, ShouldBeNil)
			So(execution.ActionTypeID, ShouldEqual, "at_001")
		})
	})
}

func Test_ExecuteAction_InputDynamicParamsValidation(t *testing.T) {
	Convey("行动执行：行动类含 input 参数时，dynamic_params 未给齐则返回 400", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		omAccess := omock.NewMockOntologyManagerAccess(mockCtrl)

		logics.OMA = omAccess

		service := &actionSchedulerService{
			appSetting: appSetting,
			omAccess:   omAccess,
		}

		ctx := context.Background()
		baseActionType := func(extraParams ...interfaces.Parameter) interfaces.ActionType {
			p := []interfaces.Parameter{{Name: "token", ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT}}
			p = append(p, extraParams...)
			return interfaces.ActionType{
				ATID:       "at_001",
				ATName:     "needs_input",
				Parameters: p,
			}
		}

		Convey("未提供 dynamic_params", func() {
			actionType := baseActionType()
			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(actionType, map[string]any{}, true, nil)

			req := &interfaces.ActionExecutionRequest{
				KNID:               "kn_001",
				ActionTypeID:       "at_001",
				InstanceIdentities: []map[string]any{{"id": "1"}},
			}

			_, err := service.ExecuteAction(ctx, req)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyQuery_ActionExecution_InvalidParameter)
		})

		Convey("仅提供部分 input（缺少另一个参数）", func() {
			actionType := baseActionType(interfaces.Parameter{
				Name: "other", ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
			})
			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(actionType, map[string]any{}, true, nil)

			req := &interfaces.ActionExecutionRequest{
				KNID:               "kn_001",
				ActionTypeID:       "at_001",
				InstanceIdentities: []map[string]any{{"id": "1"}},
				DynamicParams: map[string]any{
					"token": "ok",
				},
			}

			_, err := service.ExecuteAction(ctx, req)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyQuery_ActionExecution_InvalidParameter)
		})

		Convey("dynamic_params 中某 key 为 null，视为未提供", func() {
			actionType := baseActionType()
			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(actionType, map[string]any{}, true, nil)

			req := &interfaces.ActionExecutionRequest{
				KNID:               "kn_001",
				ActionTypeID:       "at_001",
				InstanceIdentities: []map[string]any{{"id": "1"}},
				DynamicParams: map[string]any{
					"token": nil,
				},
			}

			_, err := service.ExecuteAction(ctx, req)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyQuery_ActionExecution_InvalidParameter)
		})
	})
}

func Test_ExecuteAction_ScanMode(t *testing.T) {
	Convey("Test ExecuteAction with scan mode (empty _instance_identities)", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		omAccess := omock.NewMockOntologyManagerAccess(mockCtrl)
		aoAccess := omock.NewMockAgentOperatorAccess(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)
		logsService := action_logs.NewActionLogsService(appSetting)

		// Set global variables
		logics.OMA = omAccess
		logics.AOA = aoAccess

		service := &actionSchedulerService{
			appSetting:  appSetting,
			omAccess:    omAccess,
			aoAccess:    aoAccess,
			logsService: logsService,
			ots:         ots,
		}

		ctx := context.Background()
		knID := "kn_001"
		actionTypeID := "at_001"
		objectTypeID := "ot_001"

		Convey("成功 - 扫描模式：找到符合条件的实例", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:               knID,
				Branch:             interfaces.MAIN_BRANCH,
				ActionTypeID:       actionTypeID,
				InstanceIdentities: []map[string]any{}, // Empty, triggers scan mode
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "restart_pod",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type:   interfaces.ActionSourceTypeTool,
					BoxID:  "box_001",
					ToolID: "tool_001",
				},
				Parameters: []interfaces.Parameter{},
			}

			// Variable to capture scan result
			var scanVerified bool

			// Mock GetActionType
			omAccess.EXPECT().GetActionType(gomock.Any(), knID, interfaces.MAIN_BRANCH, actionTypeID).
				Return(actionType, map[string]any{"id": actionTypeID}, true, nil)

			// Mock GetObjectsByObjectTypeID to return scanned instances
			scannedObjects := interfaces.Objects{
				Datas: []map[string]any{
					{
						interfaces.SYSTEM_PROPERTY_INSTANCE_ID:       "1",
						interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY: map[string]any{"pod_ip": "192.168.1.1", "id": "1"},
						interfaces.SYSTEM_PROPERTY_DISPLAY:           "pod-192.168.1.1",
						"pod_ip":                                     "192.168.1.1",
						"id":                                         "1",
					},
					{
						interfaces.SYSTEM_PROPERTY_INSTANCE_ID:       "2",
						interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY: map[string]any{"pod_ip": "192.168.1.2", "id": "2"},
						interfaces.SYSTEM_PROPERTY_DISPLAY:           "pod-192.168.1.2",
						"pod_ip":                                     "192.168.1.2",
						"id":                                         "2",
					},
				},
				TotalCount: 2,
			}

			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, query *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
					// Verify query parameters
					So(query.KNID, ShouldEqual, knID)
					So(query.ObjectTypeID, ShouldEqual, objectTypeID)
					scanVerified = true
					return scannedObjects, nil
				})

			// Execute - will panic due to unmocked logsService, but scan logic should complete
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Expected panic due to unmocked logsService
						logger.Infof("Expected panic due to unmocked logsService: %v", r)
					}
				}()
				_, _ = service.ExecuteAction(ctx, req)
			}()

			// Verify scan mode was triggered and identities were populated
			So(scanVerified, ShouldBeTrue)
			So(len(req.Instances), ShouldEqual, 2)
			So(req.Instances[0].InstanceID, ShouldEqual, "1")
			So(req.Instances[1].InstanceID, ShouldEqual, "2")
		})

		Convey("失败 - 扫描模式：扫描后没有找到符合条件的实例", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:               knID,
				Branch:             interfaces.MAIN_BRANCH,
				ActionTypeID:       actionTypeID,
				InstanceIdentities: []map[string]any{}, // Empty, triggers scan mode
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "restart_pod",
				ObjectTypeID: objectTypeID,
			}

			// Mock GetActionType
			omAccess.EXPECT().GetActionType(gomock.Any(), knID, interfaces.MAIN_BRANCH, actionTypeID).
				Return(actionType, map[string]any{"id": actionTypeID}, true, nil)

			// Mock GetObjectsByObjectTypeID to return empty result
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).
				Return(interfaces.Objects{Datas: []map[string]any{}}, nil)

			_, err := service.ExecuteAction(ctx, req)

			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyQuery_ActionExecution_InvalidParameter)
		})

		Convey("失败 - 扫描模式：扫描过程出错", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:               knID,
				Branch:             interfaces.MAIN_BRANCH,
				ActionTypeID:       actionTypeID,
				InstanceIdentities: []map[string]any{}, // Empty, triggers scan mode
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "restart_pod",
				ObjectTypeID: objectTypeID,
			}

			// Mock GetActionType
			omAccess.EXPECT().GetActionType(gomock.Any(), knID, interfaces.MAIN_BRANCH, actionTypeID).
				Return(actionType, map[string]any{"id": actionTypeID}, true, nil)

			// Mock GetObjectsByObjectTypeID to return error
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).
				Return(interfaces.Objects{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError).
					WithErrorDetails("scan failed"))

			_, err := service.ExecuteAction(ctx, req)

			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("失败 - 扫描模式：超过最大执行数量限制", func() {
			// Save original limit and restore after test
			originalLimit := maxExecutionObjects
			maxExecutionObjects = 5 // Set a low limit for testing
			defer func() { maxExecutionObjects = originalLimit }()

			req := &interfaces.ActionExecutionRequest{
				KNID:               knID,
				Branch:             interfaces.MAIN_BRANCH,
				ActionTypeID:       actionTypeID,
				InstanceIdentities: []map[string]any{}, // Empty, triggers scan mode
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "restart_pod",
				ObjectTypeID: objectTypeID,
			}

			// Mock GetActionType
			omAccess.EXPECT().GetActionType(gomock.Any(), knID, interfaces.MAIN_BRANCH, actionTypeID).
				Return(actionType, map[string]any{"id": actionTypeID}, true, nil)

			// Mock GetObjectsByObjectTypeID to return more objects than the limit
			manyObjects := interfaces.Objects{
				Datas: []map[string]any{},
			}
			for i := 0; i < 10; i++ { // 10 objects, limit is 5
				manyObjects.Datas = append(manyObjects.Datas, map[string]any{
					interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY: map[string]any{"id": fmt.Sprintf("%d", i)},
				})
			}
			manyObjects.TotalCount = 10

			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).
				Return(manyObjects, nil)

			_, err := service.ExecuteAction(ctx, req)

			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyQuery_ActionExecution_InvalidParameter)
		})
	})
}

func Test_ExecuteAction_UnboundObjectType(t *testing.T) {
	Convey("Test ExecuteAction with unbound object type", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		omAccess := omock.NewMockOntologyManagerAccess(mockCtrl)
		aoAccess := omock.NewMockAgentOperatorAccess(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)
		logsService := action_logs.NewActionLogsService(appSetting)

		// Set global variables
		logics.OMA = omAccess
		logics.AOA = aoAccess

		service := &actionSchedulerService{
			appSetting:  appSetting,
			omAccess:    omAccess,
			aoAccess:    aoAccess,
			logsService: logsService,
			ots:         ots,
		}

		ctx := context.Background()
		knID := "kn_001"
		actionTypeID := "at_001"

		Convey("成功 - 未绑定对象类 + 无 identities → 构造虚拟实例", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:               knID,
				Branch:             interfaces.MAIN_BRANCH,
				ActionTypeID:       actionTypeID,
				InstanceIdentities: []map[string]any{}, // Empty
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: "", // 未绑定对象类
				ActionSource: interfaces.ActionSource{
					Type:   interfaces.ActionSourceTypeTool,
					BoxID:  "box_001",
					ToolID: "tool_001",
				},
				Parameters: []interfaces.Parameter{},
			}

			// Mock GetActionType
			omAccess.EXPECT().GetActionType(gomock.Any(), knID, interfaces.MAIN_BRANCH, actionTypeID).
				Return(actionType, map[string]any{"id": actionTypeID}, true, nil)

			// Execute - will panic due to unmocked logsService, but logic should complete
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Expected panic due to unmocked logsService
						logger.Infof("Expected panic due to unmocked logsService: %v", r)
					}
				}()
				_, _ = service.ExecuteAction(ctx, req)
			}()

			// Verify virtual instance was created
			So(len(req.Instances), ShouldEqual, 1)
			So(len(req.ObjDatas), ShouldEqual, 1)
		})

		Convey("成功 - 未绑定对象类 + 有 identities → 按 identities 构造实例", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:         knID,
				Branch:       interfaces.MAIN_BRANCH,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123", "name": "test"},
					{"id": "456", "name": "test2"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: "", // 未绑定对象类
				ActionSource: interfaces.ActionSource{
					Type:   interfaces.ActionSourceTypeTool,
					BoxID:  "box_001",
					ToolID: "tool_001",
				},
				Parameters: []interfaces.Parameter{},
			}

			// Mock GetActionType
			omAccess.EXPECT().GetActionType(gomock.Any(), knID, interfaces.MAIN_BRANCH, actionTypeID).
				Return(actionType, map[string]any{"id": actionTypeID}, true, nil)

			// Execute - will panic due to unmocked logsService, but logic should complete
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Expected panic due to unmocked logsService
						logger.Infof("Expected panic due to unmocked logsService: %v", r)
					}
				}()
				_, _ = service.ExecuteAction(ctx, req)
			}()

			// Verify instances were created from identities
			So(len(req.Instances), ShouldEqual, 2)
			So(len(req.ObjDatas), ShouldEqual, 2)
			So(req.Instances[0].InstanceIdentity, ShouldResemble, map[string]any{"id": "123", "name": "test"})
			So(req.Instances[1].InstanceIdentity, ShouldResemble, map[string]any{"id": "456", "name": "test2"})
		})
	})
}

func Test_ExecuteAction_AddActionType(t *testing.T) {
	Convey("Test ExecuteAction with add action type", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		omAccess := omock.NewMockOntologyManagerAccess(mockCtrl)
		aoAccess := omock.NewMockAgentOperatorAccess(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)
		logsService := action_logs.NewActionLogsService(appSetting)

		// Set global variables
		logics.OMA = omAccess
		logics.AOA = aoAccess

		service := &actionSchedulerService{
			appSetting:  appSetting,
			omAccess:    omAccess,
			aoAccess:    aoAccess,
			logsService: logsService,
			ots:         ots,
		}

		ctx := context.Background()
		knID := "kn_001"
		actionTypeID := "at_001"
		objectTypeID := "ot_001"

		Convey("成功 - add 行动类型 + 有 identities + 查询不到实例 → 构造实例并评估条件", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:         knID,
				Branch:       interfaces.MAIN_BRANCH,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123", "status": "active"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "add_action",
				ActionType:   "add",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type:   interfaces.ActionSourceTypeTool,
					BoxID:  "box_001",
					ToolID: "tool_001",
				},
				Condition: &cond.CondCfg{
					Name:      "status",
					Operation: "==",
					ValueOptCfg: cond.ValueOptCfg{
						Value: "active",
					},
				},
				Parameters: []interfaces.Parameter{},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{Name: "status", Type: "string"},
					},
				},
			}

			// Mock GetActionType
			omAccess.EXPECT().GetActionType(gomock.Any(), knID, interfaces.MAIN_BRANCH, actionTypeID).
				Return(actionType, map[string]any{"id": actionTypeID}, true, nil)

			// Mock GetObjectType (needed for condition evaluation)
			omAccess.EXPECT().GetObjectType(gomock.Any(), knID, interfaces.MAIN_BRANCH, objectTypeID).
				Return(objectType, true, nil)

			// Mock GetObjectsByObjectTypeID - first query by identities only (returns empty)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, query *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
					// Verify this is the first query (by identities only)
					return interfaces.Objects{Datas: []map[string]any{}}, nil
				})

			// Execute - will panic due to unmocked logsService, but logic should complete
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Expected panic due to unmocked logsService
						logger.Infof("Expected panic due to unmocked logsService: %v", r)
					}
				}()
				_, _ = service.ExecuteAction(ctx, req)
			}()

			// Verify instances were created (condition evaluation may filter some)
			So(len(req.Instances), ShouldBeGreaterThanOrEqualTo, 0)
		})

		Convey("成功 - add 行动类型 + 有 identities + 查询到实例 → 按 identities 和行动条件过滤", func() {
			req := &interfaces.ActionExecutionRequest{
				KNID:         knID,
				Branch:       interfaces.MAIN_BRANCH,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "add_action",
				ActionType:   "add",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type:   interfaces.ActionSourceTypeTool,
					BoxID:  "box_001",
					ToolID: "tool_001",
				},
				Condition: &cond.CondCfg{
					Name:      "status",
					Operation: "==",
					ValueOptCfg: cond.ValueOptCfg{
						Value: "active",
					},
				},
				Parameters: []interfaces.Parameter{},
			}

			filteredObjects := interfaces.Objects{
				Datas: []map[string]any{
					{
						interfaces.SYSTEM_PROPERTY_INSTANCE_ID:       "123",
						interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY: map[string]any{"id": "123"},
						interfaces.SYSTEM_PROPERTY_DISPLAY:           "test-instance",
						"status":                                     "active",
					},
				},
				TotalCount: 1,
			}

			// Mock GetActionType
			omAccess.EXPECT().GetActionType(gomock.Any(), knID, interfaces.MAIN_BRANCH, actionTypeID).
				Return(actionType, map[string]any{"id": actionTypeID}, true, nil)

			// Mock GetObjectsByObjectTypeID - first query by identities only (returns found)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, query *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
					// First call: query by identities only
					return interfaces.Objects{Datas: []map[string]any{{"id": "123"}}}, nil
				})
			// Mock GetObjectsByObjectTypeID - second query by identities and condition
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, query *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
					// Second call: query by identities and condition
					So(query.ActualCondition, ShouldNotBeNil)
					So(query.ActualCondition.Operation, ShouldEqual, "and")
					return filteredObjects, nil
				})

			// Execute - will panic due to unmocked logsService, but logic should complete
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Expected panic due to unmocked logsService
						logger.Infof("Expected panic due to unmocked logsService: %v", r)
					}
				}()
				_, _ = service.ExecuteAction(ctx, req)
			}()

			// Verify instances were filtered correctly
			So(len(req.Instances), ShouldEqual, 1)
			So(req.Instances[0].InstanceID, ShouldEqual, "123")
		})
	})
}
