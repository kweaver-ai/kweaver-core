// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package agent_operator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"ontology-query/common"
	"ontology-query/interfaces"
)

// newTestAgentOperatorAccess 创建用于测试的 agentOperatorAccess，允许注入 mock HTTP 客户端
func newTestAgentOperatorAccess(appSetting *common.AppSetting, httpClient rest.HTTPClient) *agentOperatorAccess {
	return &agentOperatorAccess{
		appSetting:       appSetting,
		agentOperatorUrl: appSetting.AgentOperatorUrl,
		httpClient:       httpClient,
	}
}

func Test_NewAgentOperatorAccess(t *testing.T) {
	Convey("Test NewAgentOperatorAccess", t, func() {
		appSetting := &common.AppSetting{
			AgentOperatorUrl: "http://test-ao",
		}

		Convey("成功 - 创建单例实例", func() {
			// 重置单例
			aoAccessOnce = sync.Once{}
			aoAccess = nil

			access1 := NewAgentOperatorAccess(appSetting)
			access2 := NewAgentOperatorAccess(appSetting)

			So(access1, ShouldNotBeNil)
			So(access2, ShouldNotBeNil)
			So(access1, ShouldEqual, access2) // 应该是同一个实例
		})
	})
}

func Test_agentOperatorAccess_GetAgentOperatorByID(t *testing.T) {
	Convey("Test agentOperatorAccess GetAgentOperatorByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			AgentOperatorUrl: "http://test-ao",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		aoa := newTestAgentOperatorAccess(appSetting, mockHTTPClient)

		ctx := context.Background()
		operatorID := "op1"

		Convey("成功 - 获取算子信息", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			operatorInfo := interfaces.AgentOperator{
				OperatorId: operatorID,
				Name:       "test-operator",
				Version:    "1.0.0",
				Status:     "active",
				OperatorInfo: interfaces.OperatorInfo{
					OperatorType:  "tool",
					ExecutionMode: "sync",
				},
			}
			responseBytes, _ := json.Marshal(operatorInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := aoa.GetAgentOperatorByID(ctx, operatorID)

			So(err, ShouldBeNil)
			So(result.OperatorId, ShouldEqual, operatorID)
			So(result.Name, ShouldEqual, "test-operator")
		})

		Convey("失败 - HTTP 请求错误", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, fmt.Errorf("http request failed"))

			result, err := aoa.GetAgentOperatorByID(ctx, operatorID)

			So(err, ShouldNotBeNil)
			So(result.OperatorId, ShouldEqual, "")
		})

		Convey("失败 - HTTP 状态码非 200", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			baseError := rest.BaseError{
				ErrorCode:   "ERROR_CODE",
				Description: "Error description",
			}
			errorBytes, _ := json.Marshal(baseError)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, errorBytes, nil)

			result, err := aoa.GetAgentOperatorByID(ctx, operatorID)

			So(err, ShouldNotBeNil)
			So(result.OperatorId, ShouldEqual, "")
		})

		Convey("失败 - HTTP 状态码非 200 且解析 BaseError 失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, invalidJSON, nil)

			result, err := aoa.GetAgentOperatorByID(ctx, operatorID)

			So(err, ShouldNotBeNil)
			So(result.OperatorId, ShouldEqual, "")
		})

		Convey("失败 - 响应体为空", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, nil, nil)

			result, err := aoa.GetAgentOperatorByID(ctx, operatorID)

			So(err, ShouldNotBeNil)
			So(result.OperatorId, ShouldEqual, "")
		})

		Convey("失败 - 解析响应失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, invalidJSON, nil)

			result, err := aoa.GetAgentOperatorByID(ctx, operatorID)

			So(err, ShouldNotBeNil)
			So(result.OperatorId, ShouldEqual, "")
		})

		Convey("成功 - 无账户信息", func() {
			operatorInfo := interfaces.AgentOperator{
				OperatorId: operatorID,
				Name:       "test-operator",
			}
			responseBytes, _ := json.Marshal(operatorInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := aoa.GetAgentOperatorByID(ctx, operatorID)

			So(err, ShouldBeNil)
			So(result.OperatorId, ShouldEqual, operatorID)
		})
	})
}

func Test_agentOperatorAccess_ExecuteOperator(t *testing.T) {
	Convey("Test agentOperatorAccess ExecuteOperator", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			AgentOperatorUrl: "http://test-ao",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		aoa := newTestAgentOperatorAccess(appSetting, mockHTTPClient)

		ctx := context.Background()
		operatorID := "op1"
		execRequest := interfaces.OperatorExecutionRequest{
			Header:  map[string]any{"key": "value"},
			Body:    map[string]any{"param": "value"},
			Query:   map[string]any{"q": "test"},
			Path:    map[string]any{"id": "123"},
			Timeout: 30,
		}

		Convey("成功 - 执行算子 (状态码200)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			operatorResult := operatorExecuteResult{
				StatusCode: http.StatusOK,
				Headers:    map[string]any{"Content-Type": "application/json"},
				Body:       map[string]any{"result": "success"},
				DurationMs: 100,
			}
			responseBytes, _ := json.Marshal(operatorResult)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := aoa.ExecuteOperator(ctx, operatorID, execRequest)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			body, ok := result.(map[string]any)
			So(ok, ShouldBeTrue)
			So(body["result"], ShouldEqual, "success")
		})

		Convey("成功 - 执行算子 (状态码201)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			operatorResult := operatorExecuteResult{
				StatusCode: http.StatusCreated,
				Headers:    map[string]any{},
				Body:       map[string]any{"result": "created"},
				DurationMs: 50,
			}
			responseBytes, _ := json.Marshal(operatorResult)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := aoa.ExecuteOperator(ctx, operatorID, execRequest)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})

		Convey("失败 - HTTP 请求错误", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, fmt.Errorf("http request failed"))

			result, err := aoa.ExecuteOperator(ctx, operatorID, execRequest)

			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, operatorExecuteResult{})
		})

		Convey("失败 - HTTP 状态码非 200", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			opError := OperatorError{
				Code:        "ERROR_CODE",
				Description: "Error description",
				Detail:      "Error detail",
			}
			errorBytes, _ := json.Marshal(opError)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, errorBytes, nil)

			result, err := aoa.ExecuteOperator(ctx, operatorID, execRequest)

			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, operatorExecuteResult{})
		})

		Convey("失败 - HTTP 状态码非 200 且解析 OperatorError 失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, invalidJSON, nil)

			result, err := aoa.ExecuteOperator(ctx, operatorID, execRequest)

			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, operatorExecuteResult{})
		})

		Convey("失败 - 响应体为空", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, nil, nil)

			result, err := aoa.ExecuteOperator(ctx, operatorID, execRequest)

			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, operatorExecuteResult{})
		})

		Convey("失败 - 算子执行失败 (状态码400)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			operatorResult := operatorExecuteResult{
				StatusCode: http.StatusBadRequest,
				Headers:    map[string]any{},
				Body:       map[string]any{"error": "bad request"},
				Error:      "execution failed",
				DurationMs: 50,
			}
			responseBytes, _ := json.Marshal(operatorResult)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := aoa.ExecuteOperator(ctx, operatorID, execRequest)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("失败 - 算子执行失败 (状态码300)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			operatorResult := operatorExecuteResult{
				StatusCode: http.StatusMultipleChoices,
				Headers:    map[string]any{},
				Body:       map[string]any{"error": "multiple choices"},
				DurationMs: 50,
			}
			responseBytes, _ := json.Marshal(operatorResult)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := aoa.ExecuteOperator(ctx, operatorID, execRequest)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("失败 - 算子执行失败 (状态码小于100)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			operatorResult := operatorExecuteResult{
				StatusCode: 99,
				Headers:    map[string]any{},
				Body:       map[string]any{"error": "invalid status"},
				DurationMs: 50,
			}
			responseBytes, _ := json.Marshal(operatorResult)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := aoa.ExecuteOperator(ctx, operatorID, execRequest)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("失败 - 解析响应失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, invalidJSON, nil)

			result, err := aoa.ExecuteOperator(ctx, operatorID, execRequest)

			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, operatorExecuteResult{})
		})
	})
}
