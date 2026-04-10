package drivenadapters

import (
	"context"
	"net/http"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestVegaBackendClient(t *testing.T) {
	Convey("VegaBackendClient", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		httpClient := mocks.NewMockHTTPClient(ctrl)
		client := &vegaBackendClient{
			baseURL:    "http://vega-backend:9898/api/vega-backend",
			logger:     logger.DefaultLogger(),
			httpClient: httpClient,
		}
		ctx := common.SetAccountAuthContextToCtx(context.Background(), &interfaces.AccountAuthContext{
			AccountID:   "acc-1",
			AccountType: interfaces.AccessorTypeUser,
		})
		headers := map[string]string{
			"Content-Type":  "application/json",
			"x-account-id":   "acc-1",
			"x-account-type": "user",
		}

		Convey("creates catalog with fixed fields", func() {
			req := &interfaces.VegaCatalogRequest{
				ID:          "kweaver_execution_factory_catalog",
				Name:        "kweaver_execution_factory_catalog",
				Tags:        []string{"execution-factory", "索引"},
				Description: "执行工厂的逻辑命名空间",
			}
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), "http://vega-backend:9898/api/vega-backend/v1/catalogs", headers, req).
				Return(http.StatusCreated, []byte(`{"id":"kweaver_execution_factory_catalog","name":"kweaver_execution_factory_catalog"}`), nil)

			resp, err := client.CreateCatalog(ctx, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.ID, ShouldEqual, "kweaver_execution_factory_catalog")
		})

		Convey("gets resource from entries response", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), "http://vega-backend:9898/api/vega-backend/v1/resources/kweaver_execution_factory_skill_dataset", gomock.Nil(), headers).
				Return(http.StatusOK, []byte(`{"entries":[{"id":"kweaver_execution_factory_skill_dataset","catalog_id":"kweaver_execution_factory_catalog"}]}`), nil)

			resp, err := client.GetResourceByID(ctx, "kweaver_execution_factory_skill_dataset")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.ID, ShouldEqual, "kweaver_execution_factory_skill_dataset")
		})

		Convey("writes dataset documents", func() {
			docs := []map[string]any{
				{"_id": "skill-1", "skill_id": "skill-1", "name": "demo"},
			}
			httpClient.EXPECT().PostNoUnmarshal(gomock.Any(), "http://vega-backend:9898/api/vega-backend/v1/resources/dataset/kweaver_execution_factory_skill_dataset/docs", headers, docs).
				Return(http.StatusCreated, []byte(`{}`), nil)

			err := client.WriteDatasetDocuments(ctx, "kweaver_execution_factory_skill_dataset", docs)
			So(err, ShouldBeNil)
		})

		Convey("deletes dataset document by id", func() {
			httpClient.EXPECT().DeleteNoUnmarshal(gomock.Any(), "http://vega-backend:9898/api/vega-backend/v1/resources/dataset/kweaver_execution_factory_skill_dataset/docs/skill-1", headers).
				Return(http.StatusNoContent, []byte{}, nil)

			err := client.DeleteDatasetDocumentByID(ctx, "kweaver_execution_factory_skill_dataset", "skill-1")
			So(err, ShouldBeNil)
		})
	})
}
