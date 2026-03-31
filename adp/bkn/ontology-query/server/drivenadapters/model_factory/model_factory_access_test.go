// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package model_factory

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	"ontology-query/interfaces"
)

// newTestModelFactoryAccess 创建用于测试的 modelFactoryAccess，允许注入 mock HTTP 客户端
func newTestModelFactoryAccess(appSetting *common.AppSetting, httpClient rest.HTTPClient) *modelFactoryAccess {
	return &modelFactoryAccess{
		appSetting:   appSetting,
		httpClient:   httpClient,
		mfManagerUrl: appSetting.ModelFactoryManagerUrl,
		mfAPIUrl:     appSetting.ModelFactoryAPIUrl,
	}
}

func Test_NewModelFactoryAccess(t *testing.T) {
	Convey("Test NewModelFactoryAccess", t, func() {
		appSetting := &common.AppSetting{
			ModelFactoryManagerUrl: "http://test-mf-manager",
			ModelFactoryAPIUrl:     "http://test-mf-api",
		}

		Convey("成功 - 创建单例实例", func() {
			// 重置单例
			mfAccessOnce = sync.Once{}
			mfAccess = nil

			access1 := NewModelFactoryAccess(appSetting)
			access2 := NewModelFactoryAccess(appSetting)

			So(access1, ShouldNotBeNil)
			So(access2, ShouldNotBeNil)
			So(access1, ShouldEqual, access2) // 应该是同一个实例
		})
	})
}

func Test_modelFactoryAccess_GetVector(t *testing.T) {
	Convey("Test modelFactoryAccess GetVector", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ModelFactoryAPIUrl: "http://test-mf-api",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		mfa := newTestModelFactoryAccess(appSetting, mockHTTPClient)

		ctx := context.Background()
		model := &interfaces.SmallModel{
			ModelID:   "model1",
			MaxTokens: 100,
			BatchSize: 2,
		}
		words := []string{"word1", "word2", "word3"}

		Convey("成功 - 获取向量", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			vectorResp1 := cond.VectorResp{
				Vector: []float32{0.1, 0.2, 0.3},
				Index:  0,
			}
			vectorResp2 := cond.VectorResp{
				Vector: []float32{0.4, 0.5, 0.6},
				Index:  1,
			}

			response1 := struct {
				Data []cond.VectorResp `json:"data"`
			}{
				Data: []cond.VectorResp{vectorResp1, vectorResp2},
			}
			responseBytes1, _ := sonic.Marshal(response1)

			// 第二次批量请求
			vectorResp3 := cond.VectorResp{
				Vector: []float32{0.7, 0.8, 0.9},
				Index:  0,
			}
			response2 := struct {
				Data []cond.VectorResp `json:"data"`
			}{
				Data: []cond.VectorResp{vectorResp3},
			}
			responseBytes2, _ := sonic.Marshal(response2)

			// expectedURL := "http://test-mf-api/small-model/embeddings"
			// expectedHeaders := map[string]string{
			// 	"Content-Type":                      "application/json",
			// 	interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
			// 	interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
			// }

			// 第一次批量请求 (words[0:2])
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes1, nil)

			// 第二次批量请求 (words[2:3])
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes2, nil)

			result, err := mfa.GetVector(ctx, model, words)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 3)
		})

		Convey("失败 - model 为 nil", func() {
			result, err := mfa.GetVector(ctx, nil, words)

			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("成功 - words 为空", func() {
			result, err := mfa.GetVector(ctx, model, []string{})

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
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

			result, err := mfa.GetVector(ctx, model, words)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("失败 - HTTP 状态码非 200", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, []byte("error"), nil)

			result, err := mfa.GetVector(ctx, model, words)

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

			result, err := mfa.GetVector(ctx, model, words)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("失败 - 向量数量不匹配", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			// 返回的向量数量少于输入
			response := struct {
				Data []cond.VectorResp `json:"data"`
			}{
				Data: []cond.VectorResp{
					{Vector: []float32{0.1, 0.2}, Index: 0},
				},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := mfa.GetVector(ctx, model, words[:2])

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功 - 文本长度超过 MaxTokens 被截断", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			longWord := ""
			for i := 0; i < 200; i++ {
				longWord += "a"
			}
			words := []string{longWord}

			vectorResp := cond.VectorResp{
				Vector: []float32{0.1, 0.2, 0.3},
				Index:  0,
			}
			response := struct {
				Data []cond.VectorResp `json:"data"`
			}{
				Data: []cond.VectorResp{vectorResp},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := mfa.GetVector(ctx, model, words)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
		})
	})
}

func Test_modelFactoryAccess_GetModelByID(t *testing.T) {
	Convey("Test modelFactoryAccess GetModelByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ModelFactoryManagerUrl: "http://test-mf-manager",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		mfa := newTestModelFactoryAccess(appSetting, mockHTTPClient)

		ctx := context.Background()
		modelID := "model1"

		Convey("成功 - 获取模型信息", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			smallModel := interfaces.SmallModel{
				ModelID:      modelID,
				ModelName:    "test-model",
				ModelType:    interfaces.SMALL_MODEL_TYPE_EMBEDDING,
				EmbeddingDim: 384,
				BatchSize:    10,
				MaxTokens:    512,
			}
			responseBytes, _ := sonic.Marshal(smallModel)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := mfa.GetModelByID(ctx, modelID)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.ModelID, ShouldEqual, modelID)
			So(result.ModelName, ShouldEqual, "test-model")
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

			result, err := mfa.GetModelByID(ctx, modelID)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功 - 模型不存在 (404)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusNotFound, nil, nil)

			result, err := mfa.GetModelByID(ctx, modelID)

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("失败 - HTTP 状态码非 200", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, []byte("error"), nil)

			result, err := mfa.GetModelByID(ctx, modelID)

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
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, invalidJSON, nil)

			result, err := mfa.GetModelByID(ctx, modelID)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}
