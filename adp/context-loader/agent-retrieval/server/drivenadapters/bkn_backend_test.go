// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
)

// TestSearchObjectTypes_Success 测试 SearchObjectTypes 成功场景
func TestSearchObjectTypes_Success(t *testing.T) {
	convey.Convey("TestSearchObjectTypes_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []byte(`{"object_types": []}`), nil)

		resp, err := client.SearchObjectTypes(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestSearchObjectTypes_HTTPError 测试 SearchObjectTypes HTTP 错误
func TestSearchObjectTypes_HTTPError(t *testing.T) {
	convey.Convey("TestSearchObjectTypes_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.SearchObjectTypes(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestSearchObjectTypes_NotFound 测试 SearchObjectTypes 404 错误
func TestSearchObjectTypes_NotFound(t *testing.T) {
	convey.Convey("TestSearchObjectTypes_NotFound", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock 404 响应
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(404, nil, nil)

		_, err := client.SearchObjectTypes(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestGetObjectTypeDetail_Success 测试 GetObjectTypeDetail 成功场景
func TestGetObjectTypeDetail_Success(t *testing.T) {
	convey.Convey("TestGetObjectTypeDetail_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []byte(`{"entries": [{"id": "ot-001", "name": "测试对象类"}]}`), nil)

		resp, err := client.GetObjectTypeDetail(ctx, "kn-001", []string{"ot-001"}, true)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(len(resp), convey.ShouldEqual, 1)
	})
}

// TestGetObjectTypeDetail_HTTPError 测试 GetObjectTypeDetail HTTP 错误
func TestGetObjectTypeDetail_HTTPError(t *testing.T) {
	convey.Convey("TestGetObjectTypeDetail_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.GetObjectTypeDetail(ctx, "kn-001", []string{"ot-001"}, true)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestSearchRelationTypes_Success 测试 SearchRelationTypes 成功场景
func TestSearchRelationTypes_Success(t *testing.T) {
	convey.Convey("TestSearchRelationTypes_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []byte(`{"relation_types": []}`), nil)

		resp, err := client.SearchRelationTypes(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestSearchRelationTypes_HTTPError 测试 SearchRelationTypes HTTP 错误
func TestSearchRelationTypes_HTTPError(t *testing.T) {
	convey.Convey("TestSearchRelationTypes_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.SearchRelationTypes(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestSearchActionTypes_Success 测试 SearchActionTypes 成功场景
func TestSearchActionTypes_Success(t *testing.T) {
	convey.Convey("TestSearchActionTypes_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []byte(`{"action_types": []}`), nil)

		resp, err := client.SearchActionTypes(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestGetActionTypeDetail_Success 测试 GetActionTypeDetail 成功场景
func TestGetActionTypeDetail_Success(t *testing.T) {
	convey.Convey("TestGetActionTypeDetail_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []byte(`[{"id": "at-001", "name": "测试行动类"}]`), nil)

		resp, err := client.GetActionTypeDetail(ctx, "kn-001", []string{"at-001"}, true)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(len(resp), convey.ShouldEqual, 1)
	})
}

// TestGetActionTypeDetail_HTTPError 测试 GetActionTypeDetail HTTP 错误
func TestGetActionTypeDetail_HTTPError(t *testing.T) {
	convey.Convey("TestGetActionTypeDetail_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.GetActionTypeDetail(ctx, "kn-001", []string{"at-001"}, true)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestSearchMetricTypes_Success 测试 SearchMetricTypes 成功场景
func TestSearchMetricTypes_Success(t *testing.T) {
	convey.Convey("TestSearchMetricTypes_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		respBody := []byte(`{
			"entries": [
				{
					"id": "m_001",
					"name": "cpu_usage",
					"comment": "CPU usage metric",
					"unit_type": "percent",
					"unit": "%",
					"metric_type": "atomic",
					"scope_type": "object_type",
					"scope_ref": "pod",
					"time_dimension": {
						"name": "timestamp"
					},
					"calculation_formula": {
						"op": "avg"
					},
					"analysis_dimensions": [
						{
							"name": "cluster"
						}
					]
				}
			],
			"total_count": 1
		}`)
		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, respBody, nil)

		resp, err := client.SearchMetricTypes(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(resp.TotalCount, convey.ShouldEqual, 1)
		convey.So(len(resp.Entries), convey.ShouldEqual, 1)
		convey.So(resp.Entries[0].ID, convey.ShouldEqual, "m_001")
		convey.So(resp.Entries[0].Name, convey.ShouldEqual, "cpu_usage")
		convey.So(resp.Entries[0].ScopeType, convey.ShouldEqual, "object_type")
		convey.So(resp.Entries[0].ScopeRef, convey.ShouldEqual, "pod")
	})
}

// TestSearchMetricTypes_HTTPError 测试 SearchMetricTypes HTTP 错误
func TestSearchMetricTypes_HTTPError(t *testing.T) {
	convey.Convey("TestSearchMetricTypes_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.SearchMetricTypes(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestSearchMetricTypes_NotFound 测试 SearchMetricTypes 404 错误
func TestSearchMetricTypes_NotFound(t *testing.T) {
	convey.Convey("TestSearchMetricTypes_NotFound", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &bknBackendAccess{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/bkn-backend",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryConceptsReq{
			KnID: "kn-001",
		}

		mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(404, nil, nil)

		_, err := client.SearchMetricTypes(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestBknBackendAccess_NotFoundWithBodyUsesBknError(t *testing.T) {
	testCases := []struct {
		name       string
		expectHTTP func(mockHTTPClient *mocks.MockHTTPClient, respBody []byte)
		call       func(ctx context.Context, client *bknBackendAccess) error
	}{
		{
			name: "GetKnowledgeNetworkDetail",
			expectHTTP: func(mockHTTPClient *mocks.MockHTTPClient, respBody []byte) {
				mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(http.StatusNotFound, respBody, nil)
			},
			call: func(ctx context.Context, client *bknBackendAccess) error {
				_, err := client.GetKnowledgeNetworkDetail(ctx, "kn-001")
				return err
			},
		},
		{
			name: "SearchObjectTypes",
			expectHTTP: func(mockHTTPClient *mocks.MockHTTPClient, respBody []byte) {
				mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(http.StatusNotFound, respBody, nil)
			},
			call: func(ctx context.Context, client *bknBackendAccess) error {
				_, err := client.SearchObjectTypes(ctx, &interfaces.QueryConceptsReq{KnID: "kn-001"})
				return err
			},
		},
		{
			name: "GetObjectTypeDetail",
			expectHTTP: func(mockHTTPClient *mocks.MockHTTPClient, respBody []byte) {
				mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(http.StatusNotFound, respBody, nil)
			},
			call: func(ctx context.Context, client *bknBackendAccess) error {
				_, err := client.GetObjectTypeDetail(ctx, "kn-001", []string{"ot-001"}, true)
				return err
			},
		},
		{
			name: "SearchRelationTypes",
			expectHTTP: func(mockHTTPClient *mocks.MockHTTPClient, respBody []byte) {
				mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(http.StatusNotFound, respBody, nil)
			},
			call: func(ctx context.Context, client *bknBackendAccess) error {
				_, err := client.SearchRelationTypes(ctx, &interfaces.QueryConceptsReq{KnID: "kn-001"})
				return err
			},
		},
		{
			name: "GetRelationTypeDetail",
			expectHTTP: func(mockHTTPClient *mocks.MockHTTPClient, respBody []byte) {
				mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(http.StatusNotFound, respBody, nil)
			},
			call: func(ctx context.Context, client *bknBackendAccess) error {
				_, err := client.GetRelationTypeDetail(ctx, "kn-001", []string{"rt-001"}, true)
				return err
			},
		},
		{
			name: "SearchActionTypes",
			expectHTTP: func(mockHTTPClient *mocks.MockHTTPClient, respBody []byte) {
				mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(http.StatusNotFound, respBody, nil)
			},
			call: func(ctx context.Context, client *bknBackendAccess) error {
				_, err := client.SearchActionTypes(ctx, &interfaces.QueryConceptsReq{KnID: "kn-001"})
				return err
			},
		},
		{
			name: "SearchMetricTypes",
			expectHTTP: func(mockHTTPClient *mocks.MockHTTPClient, respBody []byte) {
				mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(http.StatusNotFound, respBody, nil)
			},
			call: func(ctx context.Context, client *bknBackendAccess) error {
				_, err := client.SearchMetricTypes(ctx, &interfaces.QueryConceptsReq{KnID: "kn-001"})
				return err
			},
		},
		{
			name: "GetActionTypeDetail",
			expectHTTP: func(mockHTTPClient *mocks.MockHTTPClient, respBody []byte) {
				mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(http.StatusNotFound, respBody, nil)
			},
			call: func(ctx context.Context, client *bknBackendAccess) error {
				_, err := client.GetActionTypeDetail(ctx, "kn-001", []string{"at-001"}, true)
				return err
			},
		},
		{
			name: "CreateFullBuildOntologyJob",
			expectHTTP: func(mockHTTPClient *mocks.MockHTTPClient, respBody []byte) {
				mockHTTPClient.EXPECT().PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(http.StatusNotFound, respBody, nil)
			},
			call: func(ctx context.Context, client *bknBackendAccess) error {
				_, err := client.CreateFullBuildOntologyJob(ctx, "kn-001", &interfaces.CreateFullBuildOntologyJobReq{Name: "full-build"})
				return err
			},
		},
		{
			name: "ListOntologyJobs",
			expectHTTP: func(mockHTTPClient *mocks.MockHTTPClient, respBody []byte) {
				mockHTTPClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(http.StatusNotFound, respBody, nil)
			},
			call: func(ctx context.Context, client *bknBackendAccess) error {
				_, err := client.ListOntologyJobs(ctx, "kn-001", &interfaces.ListOntologyJobsReq{Limit: 10})
				return err
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		convey.Convey(testCase.name, t, func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := mocks.NewMockLogger(ctrl)
			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

			mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
			mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			client := &bknBackendAccess{
				logger:     mockLogger,
				baseURL:    "http://localhost:8080/api/bkn-backend",
				httpClient: mockHTTPClient,
			}

			respBody := []byte(`{
				"error_code": "BknBackend.KnowledgeNetwork.NotFound",
				"description": "knowledge network not found",
				"solution": "check kn_id",
				"error_link": "https://example.com/bkn-error",
				"error_details": {"kn_id": "kn-001"}
			}`)
			testCase.expectHTTP(mockHTTPClient, respBody)

			err := testCase.call(context.Background(), client)
			convey.So(err, convey.ShouldNotBeNil)

			httpErr, ok := err.(*infraErr.HTTPError)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(httpErr.HTTPCode, convey.ShouldEqual, http.StatusNotFound)
			convey.So(httpErr.Code, convey.ShouldEqual, "BknBackend.KnowledgeNetwork.NotFound")
			convey.So(httpErr.Description, convey.ShouldEqual, "knowledge network not found")
			convey.So(httpErr.Solution, convey.ShouldEqual, "check kn_id")
			convey.So(httpErr.ErrorLink, convey.ShouldEqual, "https://example.com/bkn-error")
			convey.So(httpErr.ErrorDetails, convey.ShouldResemble, map[string]interface{}{"kn_id": "kn-001"})
		})
	}
}
