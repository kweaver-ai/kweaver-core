package skill

import (
	"context"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestSkillIndexSync(t *testing.T) {
	Convey("SkillIndexSync", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("EnsureDataset creates catalog and resource when absent", func() {
			var createdCatalog *interfaces.VegaCatalogRequest
			var createdResource *interfaces.VegaResourceRequest
			mockModelManager := mocks.NewMockMFModelManager(ctrl)
			mockModelAPI := mocks.NewMockMFModelAPIClient(ctrl)
			mockVegaClient := mocks.NewMockVegaBackendClient(ctrl)
			syncer := &skillIndexSync{
				modelManager: mockModelManager,
				modelAPI:     mockModelAPI,
				vegaClient:   mockVegaClient,
				logger:       logger.DefaultLogger(),
			}
			mockVegaClient.EXPECT().GetCatalogByID(gomock.Any(), executionFactoryCatalogID).Return(nil, nil)
			mockVegaClient.EXPECT().CreateCatalog(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *interfaces.VegaCatalogRequest) (*interfaces.VegaCatalog, error) {
				createdCatalog = req
				return &interfaces.VegaCatalog{ID: req.ID}, nil
			})
			mockVegaClient.EXPECT().GetResourceByID(gomock.Any(), executionFactorySkillDataset).Return(nil, nil)
			mockModelManager.EXPECT().GetEmbeddingModel(gomock.Any(), interfaces.SmallModelTypeEmbedding, interfaces.SmallModelTypeEmbedding).
				Return(&interfaces.EmbeddingModel{EmbeddingDim: 768}, nil)
			mockVegaClient.EXPECT().CreateResource(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *interfaces.VegaResourceRequest) (*interfaces.VegaResource, error) {
				createdResource = req
				return &interfaces.VegaResource{ID: req.ID}, nil
			})

			err := syncer.Init(context.Background())
			So(err, ShouldBeNil)
			So(createdCatalog, ShouldNotBeNil)
			So(createdCatalog.ID, ShouldEqual, executionFactoryCatalogID)
			So(createdResource, ShouldNotBeNil)
			So(createdResource.ID, ShouldEqual, executionFactorySkillDataset)
			So(createdResource.Status, ShouldEqual, executionFactoryDatasetStatus)
			So(len(createdResource.SchemaDefinition), ShouldEqual, 10)
			var nameProperty interfaces.VegaProperty
			var descriptionProperty interfaces.VegaProperty
			for _, property := range createdResource.SchemaDefinition {
				switch property.Name {
				case "name":
					nameProperty = property
				case "description":
					descriptionProperty = property
				}
			}
			So(nameProperty.Name, ShouldEqual, "name")
			So(descriptionProperty.Name, ShouldEqual, "description")
			So(len(nameProperty.Features), ShouldEqual, 2)
			So(len(descriptionProperty.Features), ShouldEqual, 2)
			So(nameProperty.Features[0].Name, ShouldEqual, "keyword_name")
			So(nameProperty.Features[0].FeatureType, ShouldEqual, "keyword")
			So(nameProperty.Features[0].Config["ignore_above"], ShouldEqual, 1024)
			So(nameProperty.Features[1].Name, ShouldEqual, "fulltext_name")
			So(nameProperty.Features[1].FeatureType, ShouldEqual, "fulltext")
			So(descriptionProperty.Features[0].Name, ShouldEqual, "keyword_description")
			So(descriptionProperty.Features[0].FeatureType, ShouldEqual, "keyword")
			So(descriptionProperty.Features[0].Config["ignore_above"], ShouldEqual, 1024)
			So(descriptionProperty.Features[1].Name, ShouldEqual, "fulltext_description")
			So(descriptionProperty.Features[1].FeatureType, ShouldEqual, "fulltext")
		})

		Convey("EnsureDataset succeeds without embedding lookup when resource already exists", func() {
			mockModelManager := mocks.NewMockMFModelManager(ctrl)
			mockModelAPI := mocks.NewMockMFModelAPIClient(ctrl)
			mockVegaClient := mocks.NewMockVegaBackendClient(ctrl)
			syncer := &skillIndexSync{
				modelManager: mockModelManager,
				modelAPI:     mockModelAPI,
				vegaClient:   mockVegaClient,
				logger:       logger.DefaultLogger(),
			}
			mockVegaClient.EXPECT().GetCatalogByID(gomock.Any(), executionFactoryCatalogID).Return(&interfaces.VegaCatalog{ID: executionFactoryCatalogID}, nil)
			mockVegaClient.EXPECT().GetResourceByID(gomock.Any(), executionFactorySkillDataset).Return(&interfaces.VegaResource{ID: executionFactorySkillDataset}, nil)

			err := syncer.Init(context.Background())
			So(err, ShouldBeNil)
			So(syncer.isInitialized(), ShouldBeTrue)
		})

		Convey("UpsertSkill writes complete document with _id and vector", func() {
			var writtenDocs []map[string]any
			mockModelAPI := mocks.NewMockMFModelAPIClient(ctrl)
			mockVegaClient := mocks.NewMockVegaBackendClient(ctrl)
			syncer := &skillIndexSync{
				modelAPI:    mockModelAPI,
				vegaClient:  mockVegaClient,
				logger:      logger.DefaultLogger(),
				initialized: true,
			}
			mockModelAPI.EXPECT().Embeddings(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *interfaces.EmbeddingReq) (*interfaces.EmbeddingResp, error) {
				So(req.Model, ShouldEqual, interfaces.SmallModelTypeEmbedding)
				So(req.Input, ShouldResemble, []string{"demo\ndesc"})
				return &interfaces.EmbeddingResp{
					Data: []interfaces.EmbeddingData{{Embedding: []float32{0.1, 0.2}}},
				}, nil
			})
			mockVegaClient.EXPECT().WriteDatasetDocuments(gomock.Any(), executionFactorySkillDataset, gomock.Any()).
				DoAndReturn(func(ctx context.Context, datasetID string, documents []map[string]any) error {
					So(datasetID, ShouldEqual, executionFactorySkillDataset)
					writtenDocs = documents
					return nil
				})

			err := syncer.UpsertSkill(context.Background(), &model.SkillRepositoryDB{
				SkillID:     "skill-1",
				Name:        "demo",
				Description: "desc",
				Version:     "1.0.0",
				Category:    "general",
				CreateUser:  "u1",
				CreateTime:  100,
				UpdateUser:  "u2",
				UpdateTime:  200,
			})
			So(err, ShouldBeNil)
			So(len(writtenDocs), ShouldEqual, 1)
			So(writtenDocs[0]["_id"], ShouldEqual, "skill-1")
			So(writtenDocs[0]["skill_id"], ShouldEqual, "skill-1")
			So(writtenDocs[0]["name"], ShouldEqual, "demo")
			So(writtenDocs[0]["description"], ShouldEqual, "desc")
			So(writtenDocs[0]["version"], ShouldEqual, "1.0.0")
			So(writtenDocs[0]["category"], ShouldEqual, "general")
			So(writtenDocs[0]["_vector"], ShouldResemble, []float32{0.1, 0.2})
		})

		Convey("DeleteSkill deletes dataset document by skill id", func() {
			mockVegaClient := mocks.NewMockVegaBackendClient(ctrl)
			syncer := &skillIndexSync{
				vegaClient:  mockVegaClient,
				logger:      logger.DefaultLogger(),
				initialized: true,
			}
			mockVegaClient.EXPECT().DeleteDatasetDocumentByID(gomock.Any(), executionFactorySkillDataset, "skill-1").Return(nil)

			err := syncer.DeleteSkill(context.Background(), "skill-1")
			So(err, ShouldBeNil)
		})

		Convey("UpsertSkill fails when embedding result is empty", func() {
			mockModelAPI := mocks.NewMockMFModelAPIClient(ctrl)
			mockVegaClient := mocks.NewMockVegaBackendClient(ctrl)
			syncer := &skillIndexSync{
				modelAPI:    mockModelAPI,
				vegaClient:  mockVegaClient,
				logger:      logger.DefaultLogger(),
				initialized: true,
			}
			mockModelAPI.EXPECT().Embeddings(gomock.Any(), gomock.Any()).Return(&interfaces.EmbeddingResp{}, nil)

			err := syncer.UpsertSkill(context.Background(), &model.SkillRepositoryDB{
				SkillID: "skill-1",
				Name:    "demo",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "embedding result is empty")
		})
	})
}
