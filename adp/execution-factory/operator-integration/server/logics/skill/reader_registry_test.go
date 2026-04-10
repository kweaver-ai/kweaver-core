package skill

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/sandbox"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

type fakeSessionPool struct {
	acquireFunc func(ctx context.Context) (string, error)
	releaseFunc func(sessionID string)
}

func (f *fakeSessionPool) ExecuteCode(ctx context.Context, req *interfaces.ExecuteCodeReq) (*interfaces.ExecuteCodeResp, error) {
	return nil, nil
}

func (f *fakeSessionPool) GetDependencies(ctx context.Context) (*sandbox.DependenciesInfo, error) {
	return nil, nil
}

func (f *fakeSessionPool) AcquireSession(ctx context.Context) (string, error) {
	return f.acquireFunc(ctx)
}

func (f *fakeSessionPool) ReleaseSession(sessionID string) {
	if f.releaseFunc != nil {
		f.releaseFunc(sessionID)
	}
}

func TestSkillReaderAndRegistry(t *testing.T) {
	Convey("SkillReader and SkillRegistry", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("GetSkillContent returns skill download url and manifest", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			reader := &skillReader{
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				Logger:                logger.DefaultLogger(),
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-1").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-1",
				Version:      "v1",
				Status:       interfaces.BizStatusPublished.String(),
				SkillContent: "demo guide",
				FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference","size":5,"mime_type":"text/markdown"}]`,
			}, nil)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-1", "v1", SkillMD).Return(&model.SkillFileIndexDB{
				SkillID:      "skill-1",
				SkillVersion: "v1",
				RelPath:      SkillMD,
				StorageKey:   testBuildObjectKey("skill-1", "v1", SkillMD),
			}, nil)
			mockAssetStore.EXPECT().GetDownloadURL(gomock.Any(), &interfaces.OssObject{
				StorageKey: testBuildObjectKey("skill-1", "v1", SkillMD),
			}).Return("https://download/skill-1/SKILL.md", nil)

			resp, err := reader.GetSkillContent(context.Background(), &interfaces.GetSkillContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-1",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.URL, ShouldEqual, "https://download/skill-1/SKILL.md")
			So(len(resp.Files), ShouldEqual, 1)
			So(resp.Files[0].RelPath, ShouldEqual, "refs/guide.md")
			So(resp.Files[0].MimeType, ShouldEqual, "text/markdown")
		})

		Convey("ReadSkillFile checks execute permission before reading file", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			reader := &skillReader{
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				Logger:                logger.DefaultLogger(),
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-2").Return(&model.SkillRepositoryDB{
				SkillID: "skill-2", Status: interfaces.BizStatusPublished.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), "skill-2", interfaces.AuthResourceTypeSkill,
				interfaces.AuthOperationTypeExecute, interfaces.AuthOperationTypePublicAccess, interfaces.AuthOperationTypeView).Return(false, errors.New("execute forbidden"))

				ctx := common.SetPublicAPIToCtx(context.Background(), true)
				resp, err := reader.ReadSkillFile(ctx, &interfaces.ReadSkillFileReq{
					BusinessDomainID: "bd-1",
					SkillID:          "skill-2",
					RelPath:          "refs/secret.md",
				})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "execute forbidden")
		})

		Convey("ReadSkillFile returns file download url", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			reader := &skillReader{
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				Logger:                logger.DefaultLogger(),
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-3").Return(&model.SkillRepositoryDB{
				SkillID: "skill-3", Status: interfaces.BizStatusPublished.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), "skill-3", interfaces.AuthResourceTypeSkill,
				interfaces.AuthOperationTypeExecute, interfaces.AuthOperationTypePublicAccess, interfaces.AuthOperationTypeView).Return(true, nil)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-3", gomock.Any(), "refs/guide.md").Return(&model.SkillFileIndexDB{
				SkillID:       "skill-3",
				RelPath:       "refs/guide.md",
				StorageID:     "storage-1",
				StorageKey:    "/tmp/f1",
				ContentSHA256: checksumSHA256([]byte("original")),
			}, nil)
			mockAssetStore.EXPECT().GetDownloadURL(gomock.Any(), &interfaces.OssObject{
				StorageID:  "storage-1",
				StorageKey: "/tmp/f1",
			}).Return("https://download/f1", nil)

				ctx := common.SetPublicAPIToCtx(context.Background(), true)
				resp, err := reader.ReadSkillFile(ctx, &interfaces.ReadSkillFileReq{
					BusinessDomainID: "bd-1",
					SkillID:          "skill-3",
					RelPath:          "refs/guide.md",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.URL, ShouldEqual, "https://download/f1")
		})

		Convey("DeleteSkill rejects invalid status", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-4", interfaces.AuthResourceTypeSkill).Return(nil)
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-4").Return(&model.SkillRepositoryDB{
				SkillID: "skill-4", Status: interfaces.BizStatusPublished.String(), IsDeleted: true,
			}, nil)

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-4",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("DeleteSkill ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				dbTx:        mockDBTx,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-5").Return(&model.SkillRepositoryDB{
				SkillID: "skill-5", Status: interfaces.BizStatusOffline.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-5", interfaces.AuthResourceTypeSkill).Return(nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(nil, errors.New("tx unavailable"))

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-5",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get tx failed")
		})

		Convey("RegisterSkill checks create permission before registration", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				parser:                newSkillParser(),
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				dbTx:                  mockDBTx,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				Logger:                logger.DefaultLogger(),
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeSkill).Return(errors.New("create forbidden"))

			resp, err := registry.RegisterSkill(context.Background(), &interfaces.RegisterSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				FileType:         "content",
				File:             json.RawMessage(validSkillMarkdown()),
				Source:           "unit-test",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create forbidden")
		})

		Convey("RegisterSkill associates business domain after registration succeeds", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				parser:                newSkillParser(),
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				dbTx:                  mockDBTx,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				Logger:                logger.DefaultLogger(),
			}
			tx, cleanup := beginTestTx(t)
			defer cleanup()

			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeSkill).Return(nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(tx, nil)
			mockSkillRepo.EXPECT().InsertSkill(gomock.Any(), tx, gomock.Any()).Return("skill-registered", nil)
			mockBusinessDomainService.EXPECT().AssociateResource(gomock.Any(), "bd-1", "skill-registered", interfaces.AuthResourceTypeSkill).Return(nil)
			mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			resp, err := registry.RegisterSkill(context.Background(), &interfaces.RegisterSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				FileType:         "content",
				File:             json.RawMessage(validSkillMarkdown()),
				Source:           "unit-test",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-registered")
			So(resp.Status, ShouldEqual, interfaces.BizStatusUnpublish)
		})

		Convey("UpdateSkillStatus publishes skill after permission and duplicate-name checks", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-publish").Return(&model.SkillRepositoryDB{
				SkillID: "skill-publish", Name: "demo-skill", Status: interfaces.BizStatusUnpublish.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), "skill-publish", interfaces.AuthResourceTypeSkill).Return(nil)
			mockSkillRepo.EXPECT().SelectSkillByName(gomock.Any(), gomock.Nil(), "demo-skill", []string{interfaces.BizStatusPublished.String()}).Return(false, nil, nil)
			mockSkillRepo.EXPECT().UpdateSkillStatus(gomock.Any(), gomock.Nil(), "skill-publish", interfaces.BizStatusPublished.String(), "user-1").Return(nil)

			resp, err := registry.UpdateSkillStatus(context.Background(), &interfaces.UpdateSkillStatusReq{
				UserID:  "user-1",
				SkillID: "skill-publish",
				Status:  interfaces.BizStatusPublished,
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-publish")
			So(resp.Status, ShouldEqual, interfaces.BizStatusPublished)
		})

		Convey("QuerySkillList omits instructions and files from list payload", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			mockUserMgnt := mocks.NewMockUserManagement(ctrl)
			mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				BusinessDomainService: mockBusinessDomainService,
				UserMgnt:              mockUserMgnt,
				CategoryManager:       mockCategoryManager,
				Logger:                logger.DefaultLogger(),
			}
			mockBusinessDomainService.EXPECT().BatchResourceList(gomock.Any(), []string{"bd-1"}, interfaces.AuthResourceTypeSkill).Return(map[string]string{"skill-6": "bd-1"}, nil)
			mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).Return(int64(1), nil)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).Return([]*model.SkillRepositoryDB{
				{
					SkillID:      "skill-6",
					Name:         "demo-skill",
					Description:  "demo-desc",
					SkillContent: "full skill markdown",
					FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference"}]`,
					Status:       interfaces.BizStatusPublished.String(),
				},
			}, nil)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").AnyTimes()

			ctx := common.SetBusinessDomainToCtx(context.Background(), "bd-1")
			resp, err := registry.QuerySkillList(ctx, &interfaces.QuerySkillListReq{
				BusinessDomainID: "bd-1",
				CommonPageParams: interfaces.CommonPageParams{Page: 1, PageSize: 10},
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp.Data), ShouldEqual, 1)
			raw, marshalErr := json.Marshal(resp.Data[0])
			So(marshalErr, ShouldBeNil)
			So(string(raw), ShouldNotContainSubstring, "instructions")
			So(string(raw), ShouldNotContainSubstring, "files")
			So(string(raw), ShouldNotContainSubstring, "owner_id")
			So(string(raw), ShouldNotContainSubstring, "owner_type")
		})

		Convey("QuerySkillList ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			mockUserMgnt := mocks.NewMockUserManagement(ctrl)
			mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				BusinessDomainService: mockBusinessDomainService,
				UserMgnt:              mockUserMgnt,
				CategoryManager:       mockCategoryManager,
				Logger:                logger.DefaultLogger(),
			}
			mockBusinessDomainService.EXPECT().BatchResourceList(gomock.Any(), []string{"bd-1"}, interfaces.AuthResourceTypeSkill).Return(map[string]string{"skill-6b": "bd-1"}, nil)
			mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).DoAndReturn(
				func(_ context.Context, _ interface{}, filter map[string]interface{}) (int64, error) {
					_, exists := filter["owner_id"]
					So(exists, ShouldBeFalse)
					return int64(1), nil
				},
			)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).DoAndReturn(
				func(_ context.Context, _ interface{}, filter map[string]interface{}, _ interface{}, _ interface{}) ([]*model.SkillRepositoryDB, error) {
					_, exists := filter["owner_id"]
					So(exists, ShouldBeFalse)
					return []*model.SkillRepositoryDB{
						{SkillID: "skill-6b", Name: "demo-skill", Status: interfaces.BizStatusUnpublish.String()},
					}, nil
				},
			)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").AnyTimes()

			ctx := common.SetBusinessDomainToCtx(context.Background(), "bd-1")
			resp, err := registry.QuerySkillList(ctx, &interfaces.QuerySkillListReq{
				BusinessDomainID: "bd-1",
				CommonPageParams: interfaces.CommonPageParams{Page: 1, PageSize: 10},
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp.Data), ShouldEqual, 1)
			So(resp.Data[0].SkillID, ShouldEqual, "skill-6b")
		})

		Convey("GetSkillDetail omits instructions and files from detail payload", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockUserMgnt := mocks.NewMockUserManagement(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				UserMgnt:    mockUserMgnt,
				Logger:      logger.DefaultLogger(),
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-7").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-7",
				Name:         "demo-skill",
				Description:  "demo-desc",
				SkillContent: "full skill markdown",
				FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference"}]`,
				Status:       interfaces.BizStatusPublished.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckViewPermission(gomock.Any(), gomock.Any(), "skill-7", interfaces.AuthResourceTypeSkill).Return(nil)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)

			resp, err := registry.GetSkillDetail(context.Background(), &interfaces.GetSkillDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-7",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			raw, marshalErr := json.Marshal(resp)
			So(marshalErr, ShouldBeNil)
			So(string(raw), ShouldNotContainSubstring, "instructions")
			So(string(raw), ShouldNotContainSubstring, "files")
			So(string(raw), ShouldNotContainSubstring, "owner_id")
			So(string(raw), ShouldNotContainSubstring, "owner_type")
		})

		Convey("GetSkillDetail ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockUserMgnt := mocks.NewMockUserManagement(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				UserMgnt:    mockUserMgnt,
				Logger:      logger.DefaultLogger(),
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-7b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-7b", Status: interfaces.BizStatusOffline.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckViewPermission(gomock.Any(), gomock.Any(), "skill-7b", interfaces.AuthResourceTypeSkill).Return(nil)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)

			resp, err := registry.GetSkillDetail(context.Background(), &interfaces.GetSkillDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-7b",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-7b")
		})

		Convey("QuerySkillList defaults to deleting status filter when status is empty", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			mockUserMgnt := mocks.NewMockUserManagement(ctrl)
			mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				BusinessDomainService: mockBusinessDomainService,
				UserMgnt:              mockUserMgnt,
				CategoryManager:       mockCategoryManager,
				Logger:                logger.DefaultLogger(),
			}
			mockBusinessDomainService.EXPECT().BatchResourceList(gomock.Any(), []string{"bd-1"}, interfaces.AuthResourceTypeSkill).Return(map[string]string{"skill-11": "bd-1"}, nil)
			mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).Return(int64(1), nil)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).Return([]*model.SkillRepositoryDB{
				{SkillID: "skill-11", Name: "hiding", Status: interfaces.BizStatusPublished.String(), IsDeleted: true},
			}, nil)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").AnyTimes()

			ctx := common.SetBusinessDomainToCtx(context.Background(), "bd-1")
			resp, err := registry.QuerySkillList(ctx, &interfaces.QuerySkillListReq{
				BusinessDomainID: "bd-1",
				CommonPageParams: interfaces.CommonPageParams{Page: 1, PageSize: 10},
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp.Data), ShouldEqual, 1)
			So(resp.Data[0].SkillID, ShouldEqual, "skill-11")
		})

		Convey("GetSkillDetail hides deleting skills", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckViewPermission(gomock.Any(), gomock.Any(), "skill-12", interfaces.AuthResourceTypeSkill).Return(nil)
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-12").Return(&model.SkillRepositoryDB{
				SkillID: "skill-12", Status: interfaces.BizStatusPublished.String(), IsDeleted: true,
			}, nil)

			resp, err := registry.GetSkillDetail(context.Background(), &interfaces.GetSkillDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-12",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("GetSkillDetail checks view permission before returning detail", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo, AuthService: mockAuthService, Logger: logger.DefaultLogger()}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckViewPermission(gomock.Any(), gomock.Any(), "skill-12b", interfaces.AuthResourceTypeSkill).Return(errors.New("view forbidden"))

			resp, err := registry.GetSkillDetail(context.Background(), &interfaces.GetSkillDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-12b",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "view forbidden")
		})

		Convey("QuerySkillList filters non-viewable skills by auth service", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			mockUserMgnt := mocks.NewMockUserManagement(ctrl)
			mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				UserMgnt:              mockUserMgnt,
				CategoryManager:       mockCategoryManager,
				Logger:                logger.DefaultLogger(),
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().ResourceListIDs(gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeSkill, interfaces.AuthOperationTypeView).Return([]string{"skill-12c"}, nil)
			mockBusinessDomainService.EXPECT().BatchResourceList(gomock.Any(), []string{"bd-1"}, interfaces.AuthResourceTypeSkill).Return(map[string]string{
				"skill-12c": "bd-1",
			}, nil)
			mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).Return(int64(1), nil)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).Return([]*model.SkillRepositoryDB{
				{SkillID: "skill-12c", Name: "visible", IsDeleted: true},
			}, nil)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").AnyTimes()

			ctx := common.SetPublicAPIToCtx(context.Background(), true)
			ctx = common.SetBusinessDomainToCtx(ctx, "bd-1")
			resp, err := registry.QuerySkillList(ctx, &interfaces.QuerySkillListReq{
				BusinessDomainID: "bd-1",
				CommonPageParams: interfaces.CommonPageParams{Page: 1, PageSize: 10},
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(len(resp.Data), ShouldEqual, 1)
			So(resp.Data[0].SkillID, ShouldEqual, "skill-12c")
		})

		Convey("GetSkillContent hides deleting skills", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			reader := &skillReader{skillRepo: mockSkillRepo, Logger: logger.DefaultLogger()}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-13").Return(&model.SkillRepositoryDB{
				SkillID: "skill-13", Status: interfaces.BizStatusPublished.String(), IsDeleted: true,
			}, nil)

			resp, err := reader.GetSkillContent(context.Background(), &interfaces.GetSkillContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-13",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("GetSkillContent ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			reader := &skillReader{
				skillRepo:   mockSkillRepo,
				fileRepo:    mockFileRepo,
				assetStore:  mockAssetStore,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-13b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-13b", Version: "v1", Status: interfaces.BizStatusPublished.String(), SkillContent: "demo guide",
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().OperationCheckAny(
				gomock.Any(),
				gomock.Any(),
				"skill-13b",
				interfaces.AuthResourceTypeSkill,
				interfaces.AuthOperationTypeExecute,
				interfaces.AuthOperationTypePublicAccess,
				interfaces.AuthOperationTypeView,
			).Return(true, nil)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-13b", "v1", SkillMD).Return(&model.SkillFileIndexDB{
				SkillID:      "skill-13b",
				SkillVersion: "v1",
				RelPath:      SkillMD,
				StorageKey:   testBuildObjectKey("skill-13b", "v1", SkillMD),
			}, nil)
			mockAssetStore.EXPECT().GetDownloadURL(gomock.Any(), &interfaces.OssObject{
				StorageKey: testBuildObjectKey("skill-13b", "v1", SkillMD),
			}).Return("https://download/skill-13b/SKILL.md", nil)

				ctx := common.SetPublicAPIToCtx(context.Background(), true)
				resp, err := reader.GetSkillContent(ctx, &interfaces.GetSkillContentReq{
					BusinessDomainID: "bd-1",
					SkillID:          "skill-13b",
				})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-13b")
			So(resp.URL, ShouldEqual, "https://download/skill-13b/SKILL.md")
		})

		Convey("ReadSkillFile hides deleting skills", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			reader := &skillReader{skillRepo: mockSkillRepo, Logger: logger.DefaultLogger()}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-14").Return(&model.SkillRepositoryDB{
				SkillID: "skill-14", Status: interfaces.BizStatusPublished.String(), IsDeleted: true,
			}, nil)

			resp, err := reader.ReadSkillFile(context.Background(), &interfaces.ReadSkillFileReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-14",
				RelPath:          "refs/guide.md",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("ReadSkillFile ignores owner and business domain direct comparison", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			reader := &skillReader{
				skillRepo:   mockSkillRepo,
				fileRepo:    mockFileRepo,
				assetStore:  mockAssetStore,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-14b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-14b", Status: interfaces.BizStatusPublished.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().OperationCheckAny(
				gomock.Any(),
				gomock.Any(),
				"skill-14b",
				interfaces.AuthResourceTypeSkill,
				interfaces.AuthOperationTypeExecute,
				interfaces.AuthOperationTypePublicAccess,
				interfaces.AuthOperationTypeView,
			).Return(true, nil)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-14b", gomock.Any(), "refs/guide.md").Return(&model.SkillFileIndexDB{
				SkillID: "skill-14b", RelPath: "refs/guide.md", StorageID: "storage-14b", StorageKey: "/tmp/f14b", ContentSHA256: checksumSHA256([]byte("ok")),
			}, nil)
			mockAssetStore.EXPECT().GetDownloadURL(gomock.Any(), &interfaces.OssObject{
				StorageID:  "storage-14b",
				StorageKey: "/tmp/f14b",
			}).Return("https://download/f14b", nil)

				ctx := common.SetPublicAPIToCtx(context.Background(), true)
				resp, err := reader.ReadSkillFile(ctx, &interfaces.ReadSkillFileReq{
					BusinessDomainID: "bd-1",
					SkillID:          "skill-14b",
					RelPath:          "refs/guide.md",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-14b")
			So(resp.URL, ShouldEqual, "https://download/f14b")
			raw, marshalErr := json.Marshal(resp)
			So(marshalErr, ShouldBeNil)
			So(string(raw), ShouldNotContainSubstring, "access_level")
		})

			Convey("QuerySkillMarketList filters by public access and business domain visibility", func() {
				mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
				mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
				mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
				mockUserMgnt := mocks.NewMockUserManagement(ctrl)
				mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
				registry := &skillRegistry{
					skillRepo:             mockSkillRepo,
					AuthService:           mockAuthService,
					BusinessDomainService: mockBusinessDomainService,
					UserMgnt:              mockUserMgnt,
					CategoryManager:       mockCategoryManager,
					Logger:                logger.DefaultLogger(),
				}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().ResourceListIDs(gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeSkill, interfaces.AuthOperationTypePublicAccess).Return([]string{"skill-m1", "skill-m2", "skill-m3"}, nil)
			mockBusinessDomainService.EXPECT().BatchResourceList(gomock.Any(), []string{"bd-1"}, interfaces.AuthResourceTypeSkill).Return(map[string]string{
				"skill-m1": "bd-1",
			}, nil)
			mockSkillRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).DoAndReturn(
				func(_ context.Context, _ interface{}, filter map[string]interface{}) (int64, error) {
					So(filter["status"], ShouldHaveSameTypeAs, "")
					So(filter["status"], ShouldEqual, interfaces.BizStatusPublished.String())
					return int64(1), nil
				},
			)
			mockSkillRepo.EXPECT().SelectSkillListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).DoAndReturn(
				func(_ context.Context, _ interface{}, filter map[string]interface{}, _ interface{}, _ interface{}) ([]*model.SkillRepositoryDB, error) {
					So(filter["status"], ShouldHaveSameTypeAs, "")
					So(filter["status"], ShouldEqual, interfaces.BizStatusPublished.String())
					return []*model.SkillRepositoryDB{
						{SkillID: "skill-m1", Name: "visible", Status: interfaces.BizStatusPublished.String()},
					}, nil
				},
				)
				mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)
				mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").AnyTimes()

				ctx := common.SetPublicAPIToCtx(context.Background(), true)
				ctx = common.SetBusinessDomainToCtx(ctx, "bd-1")
			resp, err := registry.QuerySkillMarketList(ctx, &interfaces.QuerySkillMarketListReq{
				BusinessDomainID: "bd-1",
				CommonPageParams: interfaces.CommonPageParams{Page: 1, PageSize: 10},
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.TotalCount, ShouldEqual, 1)
			So(len(resp.Data), ShouldEqual, 1)
			So(resp.Data[0].SkillID, ShouldEqual, "skill-m1")
		})

		Convey("GetSkillMarketDetail checks public access and business domain visibility", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockUserMgnt := mocks.NewMockUserManagement(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				UserMgnt:    mockUserMgnt,
				Logger:      logger.DefaultLogger(),
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-m-detail").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-m-detail",
				Name:         "market-visible",
				Description:  "demo-desc",
				SkillContent: "full skill markdown",
				FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference"}]`,
				Status:       interfaces.BizStatusPublished.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckPublicAccessPermission(gomock.Any(), gomock.Any(), "skill-m-detail", interfaces.AuthResourceTypeSkill).Return(nil)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)

			resp, err := registry.GetSkillMarketDetail(context.Background(), &interfaces.GetSkillMarketDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-m-detail",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			raw, marshalErr := json.Marshal(resp)
			So(marshalErr, ShouldBeNil)
			So(string(raw), ShouldNotContainSubstring, "instructions")
			So(string(raw), ShouldNotContainSubstring, "files")
			So(resp.SkillID, ShouldEqual, "skill-m-detail")
		})

		Convey("GetSkillMarketDetail hides deleting skills", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo, AuthService: mockAuthService, Logger: logger.DefaultLogger()}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckPublicAccessPermission(gomock.Any(), gomock.Any(), "skill-m-deleting", interfaces.AuthResourceTypeSkill).Return(nil)
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-m-deleting").Return(&model.SkillRepositoryDB{
				SkillID: "skill-m-deleting", Status: interfaces.BizStatusPublished.String(), IsDeleted: true,
			}, nil)

			resp, err := registry.GetSkillMarketDetail(context.Background(), &interfaces.GetSkillMarketDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-m-deleting",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("GetSkillMarketDetail hides non-published skills", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{skillRepo: mockSkillRepo, AuthService: mockAuthService, Logger: logger.DefaultLogger()}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckPublicAccessPermission(gomock.Any(), gomock.Any(), "skill-m-unpublish", interfaces.AuthResourceTypeSkill).Return(nil)
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-m-unpublish").Return(&model.SkillRepositoryDB{
				SkillID: "skill-m-unpublish", Status: interfaces.BizStatusUnpublish.String(),
			}, nil)

			resp, err := registry.GetSkillMarketDetail(context.Background(), &interfaces.GetSkillMarketDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-m-unpublish",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("DeleteSkill marks deleting before cleanup and hard deletes repository on success", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				dbTx:                  mockDBTx,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				Logger:                logger.DefaultLogger(),
			}

				tx, cleanup := beginTestTx(t)
				defer cleanup()

				mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-8").Return(&model.SkillRepositoryDB{
					SkillID: "skill-8", Status: interfaces.BizStatusOffline.String(),
				}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-8", interfaces.AuthResourceTypeSkill).Return(nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(tx, nil)
			mockSkillRepo.EXPECT().UpdateSkillDeleted(gomock.Any(), tx, "skill-8", true, "user-1").Return(nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), tx, "skill-8", gomock.Any()).Return([]*model.SkillFileIndexDB{
				{SkillID: "skill-8", StorageKey: "/tmp/object-1"},
			}, nil)
			mockAssetStore.EXPECT().Delete(gomock.Any(), &interfaces.OssObject{StorageKey: "/tmp/object-1"}).Return(nil)
			mockFileRepo.EXPECT().DeleteSkillFileBySkillID(gomock.Any(), tx, "skill-8", gomock.Any()).Return(nil)
			mockSkillRepo.EXPECT().DeleteSkillByID(gomock.Any(), tx, "skill-8").Return(nil)
				mockBusinessDomainService.EXPECT().DisassociateResource(gomock.Any(), "bd-1", "skill-8", interfaces.AuthResourceTypeSkill).Return(nil)
			mockAuthService.EXPECT().DeletePolicy(gomock.Any(), []string{"skill-8"}, interfaces.AuthResourceTypeSkill).Return(nil)

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-8",
			})

			So(err, ShouldBeNil)
		})

			Convey("DeleteSkill keeps deleting status when asset cleanup fails", func() {
				mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
				mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
				mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				dbTx:                  mockDBTx,
				AuthService:           mockAuthService,
					BusinessDomainService: mockBusinessDomainService,
					Logger:                logger.DefaultLogger(),
				}

				db, sqlMock, err := sqlmock.New()
				So(err, ShouldBeNil)
				sqlMock.ExpectBegin()
				tx, err := db.Begin()
				So(err, ShouldBeNil)
				sqlMock.ExpectRollback()
				sqlMock.ExpectClose()
				defer func() {
					So(db.Close(), ShouldBeNil)
					So(sqlMock.ExpectationsWereMet(), ShouldBeNil)
				}()

				mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-9").Return(&model.SkillRepositoryDB{
					SkillID: "skill-9", Status: interfaces.BizStatusOffline.String(),
				}, nil)
				mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
				mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-9", interfaces.AuthResourceTypeSkill).Return(nil)
				mockDBTx.EXPECT().GetTx(gomock.Any()).Return(tx, nil)
			mockSkillRepo.EXPECT().UpdateSkillDeleted(gomock.Any(), tx, "skill-9", true, "user-1").Return(nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), tx, "skill-9", gomock.Any()).Return([]*model.SkillFileIndexDB{
				{SkillID: "skill-9", StorageKey: "/tmp/object-2"},
			}, nil)
			mockAssetStore.EXPECT().Delete(gomock.Any(), &interfaces.OssObject{StorageKey: "/tmp/object-2"}).Return(errors.New("delete failed"))

				err = registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
					BusinessDomainID: "bd-1",
					UserID:           "user-1",
					SkillID:          "skill-9",
				})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "delete failed")
		})

		Convey("DeleteSkill checks delete permission before cleanup", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}

				mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
				mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-9b", interfaces.AuthResourceTypeSkill).Return(errors.New("delete forbidden"))

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-9b",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "delete forbidden")
		})

		Convey("DeleteSkill follows common deletable status rule", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-9c").Return(&model.SkillRepositoryDB{
				SkillID: "skill-9c", Status: interfaces.BizStatusPublished.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-9c", interfaces.AuthResourceTypeSkill).Return(nil)

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-9c",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "can not be deleted")
		})

		Convey("DownloadSkill validates visibility and builds zip with skill content and files", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				fileRepo:    mockFileRepo,
				assetStore:  mockAssetStore,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-zip-1").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-zip-1",
				Name:         "demo-skill",
				SkillContent: "Use this skill carefully.",
				Status:       interfaces.BizStatusPublished.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), "skill-zip-1", interfaces.AuthResourceTypeSkill,
				interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess).Return(true, nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-zip-1", gomock.Any()).Return([]*model.SkillFileIndexDB{
				{SkillID: "skill-zip-1", RelPath: "refs/guide.md", StorageKey: "obj-1"},
			}, nil)
			mockAssetStore.EXPECT().Download(gomock.Any(), &interfaces.OssObject{StorageKey: "obj-1"}).Return([]byte("guide body"), nil)

			resp, err := registry.DownloadSkill(context.Background(), &interfaces.DownloadSkillReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-zip-1",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.FileName, ShouldEqual, "demo-skill.zip")

			zipReader, zipErr := zip.NewReader(bytes.NewReader(resp.Content), int64(len(resp.Content)))
			So(zipErr, ShouldBeNil)
			entries := map[string]string{}
			for _, file := range zipReader.File {
				rc, openErr := file.Open()
				So(openErr, ShouldBeNil)
				body, readErr := io.ReadAll(rc)
				So(readErr, ShouldBeNil)
				_ = rc.Close()
				entries[file.Name] = string(body)
			}
			So(entries["SKILL.md"], ShouldContainSubstring, "Use this skill carefully.")
			So(entries["refs/guide.md"], ShouldEqual, "guide body")
		})
	})
}

func testBuildObjectKey(skillID, version, relPath string) string {
	return filepath.ToSlash(filepath.Join(interfaces.OSSGatewayPrefix, "skill", skillID, version, relPath))
}

func beginTestTx(t *testing.T) (*sql.Tx, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error = %v", err)
	}
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		_ = db.Close()
		t.Fatalf("db.Begin error = %v", err)
	}
	mock.ExpectCommit()
	mock.ExpectClose()

	return tx, func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close error = %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	}
}

func TestRegisterSkillPersistsSkillVersionForZipAssets(t *testing.T) {
	Convey("RegisterSkill writes skill version into file indices for zip uploads", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db, sqlMock, err := sqlmock.New()
		So(err, ShouldBeNil)
		defer db.Close()

		sqlMock.ExpectBegin()
		tx, err := db.Begin()
		So(err, ShouldBeNil)
		sqlMock.ExpectCommit()

		mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
		mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
		mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
		mockDBTx := mocks.NewMockDBTx(ctrl)
		mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
		mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
		registry := &skillRegistry{
			parser:                newSkillParser(),
			skillRepo:             mockSkillRepo,
			fileRepo:              mockFileRepo,
			assetStore:            mockAssetStore,
			dbTx:                  mockDBTx,
			AuthService:           mockAuthService,
			BusinessDomainService: mockBusinessDomainService,
			Logger:                logger.DefaultLogger(),
		}

		mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
		mockAuthService.EXPECT().CheckCreatePermission(gomock.Any(), gomock.Any(), interfaces.AuthResourceTypeSkill).Return(nil)
		mockDBTx.EXPECT().GetTx(gomock.Any()).Return(tx, nil)
		mockSkillRepo.EXPECT().InsertSkill(gomock.Any(), tx, gomock.Any()).DoAndReturn(
			func(_ context.Context, _ *sql.Tx, skill *model.SkillRepositoryDB) (string, error) {
				So(skill.Version, ShouldNotBeBlank)
				return "skill-versioned", nil
			},
		)
		mockAssetStore.EXPECT().Upload(gomock.Any(), "skill-versioned", gomock.Any(), "SKILL.md", []byte(validSkillMarkdown())).Return(
			&interfaces.OssObject{StorageID: "storage-skill-md", StorageKey: "object-skill-md"},
			checksumSHA256([]byte(validSkillMarkdown())),
			nil,
		)
		mockAssetStore.EXPECT().Upload(gomock.Any(), "skill-versioned", gomock.Any(), "refs/guide.md", []byte("guide")).Return(
			&interfaces.OssObject{StorageID: "storage-guide", StorageKey: "object-guide"},
			checksumSHA256([]byte("guide")),
			nil,
		)
		mockFileRepo.EXPECT().BatchInsertSkillFiles(gomock.Any(), tx, gomock.Any()).DoAndReturn(
			func(_ context.Context, _ *sql.Tx, files []*model.SkillFileIndexDB) error {
				So(files, ShouldHaveLength, 2)
				So(files[0].SkillVersion, ShouldNotBeBlank)
				So(files[1].SkillVersion, ShouldEqual, files[0].SkillVersion)
				So(files[0].SkillID, ShouldEqual, "skill-versioned")
				So(files[1].SkillID, ShouldEqual, "skill-versioned")
				return nil
			},
		)
		mockBusinessDomainService.EXPECT().AssociateResource(gomock.Any(), "bd-1", "skill-versioned", interfaces.AuthResourceTypeSkill).Return(nil)
		mockAuthService.EXPECT().CreateOwnerPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		resp, err := registry.RegisterSkill(context.Background(), &interfaces.RegisterSkillReq{
			BusinessDomainID: "bd-1",
			UserID:           "user-1",
			FileType:         "zip",
			File: buildZip(t, map[string]string{
				"SKILL.md":      validSkillMarkdown(),
				"refs/guide.md": "guide",
			}),
			Source: "unit-test",
		})

		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(resp.SkillID, ShouldEqual, "skill-versioned")
		So(resp.Version, ShouldNotBeBlank)
		So(sqlMock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestExecuteSkillUploadsBeforeShellExecution(t *testing.T) {
	Convey("ExecuteSkill uploads archive before executing shell", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
		mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
		mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
		mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
		mockSandbox := mocks.NewMockSandBoxControlPlane(ctrl)
		callOrder := []string{}
		sessionPool := &fakeSessionPool{
			acquireFunc: func(ctx context.Context) (string, error) {
				callOrder = append(callOrder, "acquire")
				return "sess_aoi_0", nil
			},
			releaseFunc: func(sessionID string) {
				callOrder = append(callOrder, "release")
				So(sessionID, ShouldEqual, "sess_aoi_0")
			},
		}
		registry := &skillRegistry{
			skillRepo:     mockSkillRepo,
			fileRepo:      mockFileRepo,
			assetStore:    mockAssetStore,
			sandboxClient: mockSandbox,
			sessionPool:   sessionPool,
			AuthService:   mockAuthService,
			Logger:        logger.DefaultLogger(),
		}

		mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
		mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), "skill-exec-1", interfaces.AuthResourceTypeSkill,
			interfaces.AuthOperationTypeExecute, interfaces.AuthOperationTypePublicAccess).Return(true, nil)
		mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-exec-1").Return(&model.SkillRepositoryDB{
			SkillID:      "skill-exec-1",
			Name:         "demo-skill",
			Description:  "demo desc",
			Version:      "v1",
			SkillContent: "run this skill",
		}, nil)
		mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-exec-1", "v1").Return([]*model.SkillFileIndexDB{
			{
				SkillID:    "skill-exec-1",
				RelPath:    "refs/guide.md",
				StorageKey: "obj-1",
			},
		}, nil)
		mockAssetStore.EXPECT().Download(gomock.Any(), &interfaces.OssObject{StorageKey: "obj-1"}).Return([]byte("guide body"), nil)

		mockSandbox.EXPECT().UploadSkillArchive(gomock.Any(), "sess_aoi_0", gomock.Any()).DoAndReturn(
			func(_ context.Context, sessionID string, req *interfaces.UploadSkillArchiveReq) (*interfaces.UploadSkillArchiveResp, error) {
				callOrder = append(callOrder, "upload")
				So(sessionID, ShouldEqual, "sess_aoi_0")
				So(req.WorkDir, ShouldEqual, "skills/skill-exec-1")
				So(req.FileName, ShouldEqual, "demo-skill.zip")

				zr, zipErr := zip.NewReader(bytes.NewReader(req.Content), int64(len(req.Content)))
				So(zipErr, ShouldBeNil)
				entries := map[string]string{}
				for _, f := range zr.File {
					rc, openErr := f.Open()
					So(openErr, ShouldBeNil)
					body, readErr := io.ReadAll(rc)
					So(readErr, ShouldBeNil)
					So(rc.Close(), ShouldBeNil)
					entries[f.Name] = string(body)
				}
				So(entries["SKILL.md"], ShouldContainSubstring, "name: demo-skill")
				So(entries["refs/guide.md"], ShouldEqual, "guide body")
				return &interfaces.UploadSkillArchiveResp{
					SessionID:    sessionID,
					WorkDir:      "skills/skill-exec-1",
					FileName:     req.FileName,
					UploadedPath: "skills/skill-exec-1",
					Mocked:       true,
				}, nil
			},
		)
		mockSandbox.EXPECT().ExecuteShell(gomock.Any(), "sess_aoi_0", gomock.Any()).DoAndReturn(
			func(_ context.Context, sessionID string, req *interfaces.ExecuteShellReq) (*interfaces.ExecuteShellResp, error) {
				callOrder = append(callOrder, "exec")
				So(sessionID, ShouldEqual, "sess_aoi_0")
				So(req.WorkDir, ShouldEqual, "skills/skill-exec-1")
				So(req.Command, ShouldEqual, "bash run.sh")
				So(req.Timeout, ShouldEqual, 15)
				return &interfaces.ExecuteShellResp{
					SessionID:     sessionID,
					WorkDir:       req.WorkDir,
					Command:       req.Command,
					ExitCode:      0,
					Stdout:        "ok",
					ExecutionTime: 8,
					Mocked:        true,
				}, nil
			},
		)

		resp, err := registry.ExecuteSkill(context.Background(), &interfaces.ExecuteSkillReq{
			BusinessDomainID: "bd-1",
			UserID:           "user-1",
			SkillID:          "skill-exec-1",
			EntryShell:       "bash run.sh",
			Timeout:          15,
		})

		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(resp.SessionID, ShouldEqual, "sess_aoi_0")
		So(resp.WorkDir, ShouldEqual, "skills/skill-exec-1")
		So(resp.UploadedPath, ShouldEqual, "skills/skill-exec-1")
		So(resp.Command, ShouldEqual, "bash run.sh")
		So(resp.Stdout, ShouldEqual, "ok")
		So(callOrder, ShouldResemble, []string{"acquire", "upload", "exec", "release"})
	})
}
