package drivenadapters

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestGetEmbeddingModel(t *testing.T) {
	Convey("GetEmbeddingModel", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		httpClient := mocks.NewMockHTTPClient(ctrl)
		manager := &mfModelManager{
			baseURL:    "http://mf-model-manage:9898/api/private/mf-model-manage",
			logger:     logger.DefaultLogger(),
			httpClient: httpClient,
		}
		ctx := common.SetAccountAuthContextToCtx(context.Background(), &interfaces.AccountAuthContext{
			AccountID:   "acc-1",
			AccountType: interfaces.AccessorTypeUser,
		})

		Convey("parses embedding model from res list", func() {
			httpClient.EXPECT().Get(gomock.Any(), "http://mf-model-manage:9898/api/private/mf-model-manage/v1/small-model/list",
				gomock.Any(), map[string]string{
					"Content-Type":  "application/json",
					"x-account-id":   "acc-1",
					"x-account-type": "user",
				}).
				DoAndReturn(func(_ context.Context, _ string, query url.Values, _ map[string]string) (int, any, error) {
					So(query.Get("model_name"), ShouldEqual, "embedding")
					So(query.Get("model_type"), ShouldEqual, "embedding")
					return 200, map[string]any{
						"res": []map[string]any{
							{
								"model_id":      "model-1",
								"model_name":    "embedding",
								"model_type":    "embedding",
								"embedding_dim": 768,
								"batch_size":    16,
								"max_tokens":    2048,
							},
						},
					}, nil
				})

			resp, err := manager.GetEmbeddingModel(ctx, "embedding", interfaces.SmallModelTypeEmbedding)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.ModelID, ShouldEqual, "model-1")
			So(resp.EmbeddingDim, ShouldEqual, 768)
		})

		Convey("returns not found when response has no models", func() {
			httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(200, map[string]any{"res": []map[string]any{}}, nil)

			resp, err := manager.GetEmbeddingModel(ctx, "embedding", interfaces.SmallModelTypeEmbedding)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestEmbeddings(t *testing.T) {
	Convey("Embeddings", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		httpClient := mocks.NewMockHTTPClient(ctrl)
		client := &mfModelAPIClient{
			baseURL:    "http://mf-model-api:9898/api/private/mf-model-api",
			logger:     logger.DefaultLogger(),
			httpClient: httpClient,
		}
		ctx := common.SetAccountAuthContextToCtx(context.Background(), &interfaces.AccountAuthContext{
			AccountID:   "acc-1",
			AccountType: interfaces.AccessorTypeUser,
		})

		httpClient.EXPECT().Post(gomock.Any(), "http://mf-model-api:9898/api/private/mf-model-api/v1/small-model/embeddings",
			map[string]string{
				"x-account-id":   "acc-1",
				"x-account-type": "user",
			}, &interfaces.EmbeddingReq{
				Model: "embedding",
				Input: []string{"name\ndesc"},
			}).
			Return(200, map[string]any{
				"data": []map[string]any{
					{
						"object":    "embedding",
						"embedding": []float32{0.1, 0.2},
						"index":     0,
					},
				},
			}, nil)

		resp, err := client.Embeddings(ctx, &interfaces.EmbeddingReq{
			Model: "embedding",
			Input: []string{"name\ndesc"},
		})
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(len(resp.Data), ShouldEqual, 1)
		So(fmt.Sprintf("%.1f", resp.Data[0].Embedding[0]), ShouldEqual, "0.1")
	})
}
