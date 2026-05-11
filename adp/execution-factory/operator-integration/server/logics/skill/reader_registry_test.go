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
	"github.com/agiledragon/gomonkey/v2"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
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

type stubSkillReleaseRepo struct {
	selectBySkillID func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error)
	insert          func(ctx context.Context, tx *sql.Tx, release *model.SkillReleaseDB) error
	updateBySkillID func(ctx context.Context, tx *sql.Tx, release *model.SkillReleaseDB) error
}

func (s *stubSkillReleaseRepo) Insert(ctx context.Context, tx *sql.Tx, release *model.SkillReleaseDB) error {
	if s.insert != nil {
		return s.insert(ctx, tx, release)
	}
	return nil
}

func (s *stubSkillReleaseRepo) UpdateBySkillID(ctx context.Context, tx *sql.Tx, release *model.SkillReleaseDB) error {
	if s.updateBySkillID != nil {
		return s.updateBySkillID(ctx, tx, release)
	}
	return nil
}

func (s *stubSkillReleaseRepo) SelectBySkillID(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
	if s.selectBySkillID != nil {
		return s.selectBySkillID(ctx, tx, skillID)
	}
	return nil, nil
}

func (s *stubSkillReleaseRepo) SelectListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{},
	sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) ([]*model.SkillReleaseDB, error) {
	return nil, nil
}

func (s *stubSkillReleaseRepo) CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (int64, error) {
	return 0, nil
}

func (s *stubSkillReleaseRepo) DeleteBySkillID(ctx context.Context, tx *sql.Tx, skillID string) error {
	return nil
}

type stubSkillReleaseHistoryRepo struct {
	insert                    func(ctx context.Context, tx *sql.Tx, history *model.SkillReleaseHistoryDB) error
	selectBySkillID           func(ctx context.Context, tx *sql.Tx, skillID string) ([]*model.SkillReleaseHistoryDB, error)
	selectBySkillIDAndVersion func(ctx context.Context, tx *sql.Tx, skillID, version string) (*model.SkillReleaseHistoryDB, error)
	deleteByID                func(ctx context.Context, tx *sql.Tx, id int64) error
	deleteBySkillID           func(ctx context.Context, tx *sql.Tx, skillID string) error
}

func (s *stubSkillReleaseHistoryRepo) Insert(ctx context.Context, tx *sql.Tx, history *model.SkillReleaseHistoryDB) error {
	if s.insert != nil {
		return s.insert(ctx, tx, history)
	}
	return nil
}

func (s *stubSkillReleaseHistoryRepo) SelectBySkillID(ctx context.Context, tx *sql.Tx, skillID string) ([]*model.SkillReleaseHistoryDB, error) {
	if s.selectBySkillID != nil {
		return s.selectBySkillID(ctx, tx, skillID)
	}
	return nil, nil
}

func (s *stubSkillReleaseHistoryRepo) SelectBySkillIDAndVersion(ctx context.Context, tx *sql.Tx, skillID, version string) (*model.SkillReleaseHistoryDB, error) {
	if s.selectBySkillIDAndVersion != nil {
		return s.selectBySkillIDAndVersion(ctx, tx, skillID, version)
	}
	return nil, nil
}

func (s *stubSkillReleaseHistoryRepo) DeleteByID(ctx context.Context, tx *sql.Tx, id int64) error {
	if s.deleteByID != nil {
		return s.deleteByID(ctx, tx, id)
	}
	return nil
}

func (s *stubSkillReleaseHistoryRepo) DeleteBySkillID(ctx context.Context, tx *sql.Tx, skillID string) error {
	if s.deleteBySkillID != nil {
		return s.deleteBySkillID(ctx, tx, skillID)
	}
	return nil
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
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			reader := &skillReader{
				releaseRepo: &stubSkillReleaseRepo{
					selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
						return &model.SkillReleaseDB{
							SkillID:      "skill-1",
							Version:      "v1",
							Status:       interfaces.BizStatusPublished.String(),
							SkillContent: "demo guide",
							FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference","size":5,"mime_type":"text/markdown"}]`,
						}, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				Logger:                logger.DefaultLogger(),
			}
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
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			reader := &skillReader{
				releaseRepo: &stubSkillReleaseRepo{
					selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
						return &model.SkillReleaseDB{
							SkillID: "skill-2", Version: "v1", Status: interfaces.BizStatusPublished.String(),
						}, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				Logger:                logger.DefaultLogger(),
			}
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
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			reader := &skillReader{
				releaseRepo: &stubSkillReleaseRepo{
					selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
						return &model.SkillReleaseDB{
							SkillID: "skill-3", Version: "v1", Status: interfaces.BizStatusPublished.String(),
						}, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				Logger:                logger.DefaultLogger(),
			}
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
			mockAssetStore.EXPECT().Upload(gomock.Any(), "skill-registered", gomock.Any(), "SKILL.md", gomock.Any()).
				Return(&interfaces.OssObject{StorageID: "s1", StorageKey: "k1"}, "checksum", nil)
			mockFileRepo.EXPECT().BatchInsertSkillFiles(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
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
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			tx := &sql.Tx{}
			defer patchTxMethods()()
			registry := &skillRegistry{
				skillRepo:          mockSkillRepo,
				releaseRepo:        &stubSkillReleaseRepo{},
				releaseHistoryRepo: &stubSkillReleaseHistoryRepo{},
				dbTx:               mockDBTx,
				AuthService:        mockAuthService,
				Logger:             logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-publish").Return(&model.SkillRepositoryDB{
				SkillID: "skill-publish", Name: "demo-skill", Status: interfaces.BizStatusUnpublish.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), "skill-publish", interfaces.AuthResourceTypeSkill).Return(nil)
			mockSkillRepo.EXPECT().SelectSkillByName(gomock.Any(), gomock.Nil(), "demo-skill", []string{interfaces.BizStatusPublished.String()}).Return(false, nil, nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(tx, nil)
			mockSkillRepo.EXPECT().UpdateSkillStatus(gomock.Any(), tx, "skill-publish", interfaces.BizStatusPublished.String(), "user-1").Return(nil)

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

		Convey("UpdateSkillMetadata moves published draft back to editing without changing version", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
			tx := &sql.Tx{}
			defer patchTxMethods()()
			registry := &skillRegistry{
				skillRepo:       mockSkillRepo,
				fileRepo:        mockFileRepo,
				assetStore:      mockAssetStore,
				dbTx:            mockDBTx,
				AuthService:     mockAuthService,
				CategoryManager: mockCategoryManager,
				Logger:          logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-meta-1").Return(&model.SkillRepositoryDB{
				SkillID: "skill-meta-1", Name: "old-name", Description: "old-desc", Version: "v1",
				Status: interfaces.BizStatusPublished.String(), Category: interfaces.CategoryTypeOther.String(), Source: "custom",
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), "skill-meta-1", interfaces.AuthResourceTypeSkill).Return(nil)
			mockCategoryManager.EXPECT().CheckCategory(interfaces.CategoryTypeOther).Return(true)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(tx, nil)
			mockSkillRepo.EXPECT().UpdateSkill(gomock.Any(), tx, gomock.Any()).DoAndReturn(
				func(_ context.Context, _ *sql.Tx, skill *model.SkillRepositoryDB) error {
					So(skill.SkillID, ShouldEqual, "skill-meta-1")
					So(skill.Name, ShouldEqual, "old-name")
					So(skill.Description, ShouldEqual, "new-desc")
					So(skill.Version, ShouldEqual, "v1")
					So(skill.Status, ShouldEqual, interfaces.BizStatusEditing.String())
					So(skill.ExtendInfo, ShouldEqual, `{"foo":"bar"}`)
					return nil
				},
			)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-meta-1", "v1", SkillMD).
				Return(nil, nil)

			resp, err := registry.UpdateSkillMetadata(context.Background(), &interfaces.UpdateSkillMetadataReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-meta-1",
				Name:             "old-name",
				Description:      "new-desc",
				Category:         interfaces.CategoryTypeOther,
				Source:           "custom",
				ExtendInfo:       json.RawMessage(`{"foo":"bar"}`),
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-meta-1")
			So(resp.Version, ShouldEqual, "v1")
			So(resp.Status, ShouldEqual, interfaces.BizStatusEditing)
		})

		Convey("UpdateSkillPackage creates a new draft version and file indices", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			tx := &sql.Tx{}
			defer patchTxMethods()()
			registry := &skillRegistry{
				parser:      newSkillParser(),
				skillRepo:   mockSkillRepo,
				fileRepo:    mockFileRepo,
				assetStore:  mockAssetStore,
				dbTx:        mockDBTx,
				AuthService: mockAuthService,
				indexSync:   mockIndexSync,
				Logger:      logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-pkg-1").Return(&model.SkillRepositoryDB{
				SkillID: "skill-pkg-1", Name: "demo-skill", Description: "old-desc", Version: "old-v",
				Status: interfaces.BizStatusPublished.String(), Category: interfaces.CategoryTypeOther.String(), Source: "custom",
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), "skill-pkg-1", interfaces.AuthResourceTypeSkill).Return(nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(tx, nil)
			mockSkillRepo.EXPECT().UpdateSkill(gomock.Any(), tx, gomock.Any()).DoAndReturn(
				func(_ context.Context, _ *sql.Tx, skill *model.SkillRepositoryDB) error {
					So(skill.SkillID, ShouldEqual, "skill-pkg-1")
					So(skill.Version, ShouldNotEqual, "old-v")
					So(skill.Status, ShouldEqual, interfaces.BizStatusEditing.String())
					So(skill.FileManifest, ShouldContainSubstring, "refs/guide.md")
					return nil
				},
			)
			mockAssetStore.EXPECT().Upload(gomock.Any(), "skill-pkg-1", gomock.Any(), "SKILL.md", []byte(validSkillMarkdown())).Return(
				&interfaces.OssObject{StorageID: "storage-skill-md", StorageKey: "object-skill-md"},
				checksumSHA256([]byte(validSkillMarkdown())),
				nil,
			)
			mockAssetStore.EXPECT().Upload(gomock.Any(), "skill-pkg-1", gomock.Any(), "refs/guide.md", []byte("guide")).Return(
				&interfaces.OssObject{StorageID: "storage-guide", StorageKey: "object-guide"},
				checksumSHA256([]byte("guide")),
				nil,
			)
			mockFileRepo.EXPECT().BatchInsertSkillFiles(gomock.Any(), tx, gomock.Any()).DoAndReturn(
				func(_ context.Context, _ *sql.Tx, files []*model.SkillFileIndexDB) error {
					So(files, ShouldHaveLength, 2)
					So(files[0].SkillVersion, ShouldNotBeBlank)
					So(files[1].SkillVersion, ShouldEqual, files[0].SkillVersion)
					return nil
				},
			)
			mockIndexSync.EXPECT().UpdateSkill(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, skill *model.SkillRepositoryDB) error {
				So(skill.SkillID, ShouldEqual, "skill-pkg-1")
				So(skill.Version, ShouldNotEqual, "old-v")
				So(skill.Status, ShouldEqual, interfaces.BizStatusEditing.String())
				return nil
			})

			resp, err := registry.UpdateSkillPackage(context.Background(), &interfaces.UpdateSkillPackageReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-pkg-1",
				FileType:         "zip",
				File: buildZip(t, map[string]string{
					"SKILL.md":      validSkillMarkdown(),
					"refs/guide.md": "guide",
				}),
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-pkg-1")
			So(resp.Version, ShouldNotEqual, "old-v")
			So(resp.Status, ShouldEqual, interfaces.BizStatusEditing)
		})

		Convey("RepublishSkillHistory restores a historical release into draft editing state", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			tx := &sql.Tx{}
			defer patchTxMethods()()
			registry := &skillRegistry{
				skillRepo: mockSkillRepo,
				releaseHistoryRepo: &stubSkillReleaseHistoryRepo{
					selectBySkillIDAndVersion: func(ctx context.Context, tx *sql.Tx, skillID, version string) (*model.SkillReleaseHistoryDB, error) {
						return &model.SkillReleaseHistoryDB{
							SkillID:      "skill-hist-1",
							Version:      "hist-v1",
							SkillRelease: `{"skill_id":"skill-hist-1","name":"hist-name","description":"hist-desc","skill_content":"hist content","version":"hist-v1","category":"other_category","status":"published","source":"custom","extend_info":"{\"foo\":\"bar\"}","dependencies":"{\"pkg\":\"1.0\"}","file_manifest":"[{\"rel_path\":\"refs/guide.md\"}]","create_user":"creator","create_time":1,"update_user":"publisher","update_time":2}`,
						}, nil
					},
				},
				dbTx:        mockDBTx,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-hist-1").Return(&model.SkillRepositoryDB{
				SkillID: "skill-hist-1", Name: "draft-name", Description: "draft-desc", Version: "draft-v2",
				Status: interfaces.BizStatusPublished.String(), Category: interfaces.CategoryTypeOther.String(), Source: "custom",
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), "skill-hist-1", interfaces.AuthResourceTypeSkill).Return(nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(tx, nil)
			mockSkillRepo.EXPECT().UpdateSkill(gomock.Any(), tx, gomock.Any()).DoAndReturn(
				func(_ context.Context, _ *sql.Tx, skill *model.SkillRepositoryDB) error {
					So(skill.SkillID, ShouldEqual, "skill-hist-1")
					So(skill.Name, ShouldEqual, "hist-name")
					So(skill.Description, ShouldEqual, "hist-desc")
					So(skill.SkillContent, ShouldEqual, "hist content")
					So(skill.Version, ShouldEqual, "hist-v1")
					So(skill.Status, ShouldEqual, interfaces.BizStatusEditing.String())
					So(skill.FileManifest, ShouldContainSubstring, "refs/guide.md")
					return nil
				},
			)

			resp, err := registry.RepublishSkillHistory(context.Background(), &interfaces.RepublishSkillHistoryReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-hist-1",
				Version:          "hist-v1",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-hist-1")
			So(resp.Version, ShouldEqual, "hist-v1")
			So(resp.Status, ShouldEqual, interfaces.BizStatusEditing)
		})

		Convey("PublishSkillHistory republishes a historical version and syncs index after commit", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			callOrder := []string{}
			tx := &sql.Tx{}
			rollbackPatch := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error { return nil })
			defer rollbackPatch.Reset()
			commitPatch := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
				callOrder = append(callOrder, "commit")
				return nil
			})
			defer commitPatch.Reset()
			registry := &skillRegistry{
				skillRepo: mockSkillRepo,
				releaseRepo: &stubSkillReleaseRepo{
					selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
						return nil, nil
					},
					insert: func(ctx context.Context, tx *sql.Tx, release *model.SkillReleaseDB) error {
						callOrder = append(callOrder, "insert-release")
						So(release.Version, ShouldEqual, "hist-v1")
						So(release.Name, ShouldEqual, "hist-name")
						return nil
					},
				},
				releaseHistoryRepo: &stubSkillReleaseHistoryRepo{
					selectBySkillIDAndVersion: func(ctx context.Context, tx *sql.Tx, skillID, version string) (*model.SkillReleaseHistoryDB, error) {
						return &model.SkillReleaseHistoryDB{
							SkillID:      "skill-publish-h1",
							Version:      "hist-v1",
							SkillRelease: `{"skill_id":"skill-publish-h1","name":"hist-name","description":"hist-desc","skill_content":"hist content","version":"hist-v1","category":"other_category","status":"published","source":"custom","extend_info":"{\"foo\":\"bar\"}","dependencies":"{\"pkg\":\"1.0\"}","file_manifest":"[{\"rel_path\":\"refs/guide.md\"}]","create_user":"creator","create_time":1,"update_user":"publisher","update_time":2}`,
						}, nil
					},
				},
				dbTx:        mockDBTx,
				AuthService: mockAuthService,
				indexSync:   mockIndexSync,
				Logger:      logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-publish-h1").Return(&model.SkillRepositoryDB{
				SkillID: "skill-publish-h1", Name: "draft-name", Description: "draft-desc", Version: "draft-v2",
				Status: interfaces.BizStatusEditing.String(), Category: interfaces.CategoryTypeOther.String(), Source: "custom",
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckModifyPermission(gomock.Any(), gomock.Any(), "skill-publish-h1", interfaces.AuthResourceTypeSkill).Return(nil)
			mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), "skill-publish-h1", interfaces.AuthResourceTypeSkill).Return(nil)
			mockSkillRepo.EXPECT().SelectSkillByName(gomock.Any(), gomock.Nil(), "hist-name", []string{interfaces.BizStatusPublished.String()}).Return(false, nil, nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(tx, nil)
			mockSkillRepo.EXPECT().UpdateSkill(gomock.Any(), tx, gomock.Any()).DoAndReturn(
				func(_ context.Context, _ *sql.Tx, skill *model.SkillRepositoryDB) error {
					callOrder = append(callOrder, "update-skill")
					So(skill.Version, ShouldEqual, "hist-v1")
					So(skill.Status, ShouldEqual, interfaces.BizStatusPublished.String())
					return nil
				},
			)
			mockIndexSync.EXPECT().UpsertSkill(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, skill *model.SkillRepositoryDB) error {
				callOrder = append(callOrder, "index-sync")
				So(skill.Version, ShouldEqual, "hist-v1")
				return nil
			})

			resp, err := registry.PublishSkillHistory(context.Background(), &interfaces.PublishSkillHistoryReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-publish-h1",
				Version:          "hist-v1",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-publish-h1")
			So(resp.Version, ShouldEqual, "hist-v1")
			So(resp.Status, ShouldEqual, interfaces.BizStatusPublished)
			So(callOrder, ShouldResemble, []string{"update-skill", "insert-release", "index-sync", "commit"})
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
			mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
			registry := &skillRegistry{
				skillRepo:       mockSkillRepo,
				AuthService:     mockAuthService,
				UserMgnt:        mockUserMgnt,
				CategoryManager: mockCategoryManager,
				Logger:          logger.DefaultLogger(),
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
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").AnyTimes()

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
			mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
			registry := &skillRegistry{
				skillRepo:       mockSkillRepo,
				AuthService:     mockAuthService,
				UserMgnt:        mockUserMgnt,
				CategoryManager: mockCategoryManager,
				Logger:          logger.DefaultLogger(),
			}
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-7b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-7b", Status: interfaces.BizStatusOffline.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckViewPermission(gomock.Any(), gomock.Any(), "skill-7b", interfaces.AuthResourceTypeSkill).Return(nil)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").AnyTimes()

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
			reader := &skillReader{releaseRepo: &stubSkillReleaseRepo{}, Logger: logger.DefaultLogger()}

			resp, err := reader.GetSkillContent(context.Background(), &interfaces.GetSkillContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-13",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("GetSkillContent ignores owner and business domain direct comparison", func() {
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			reader := &skillReader{
				releaseRepo: &stubSkillReleaseRepo{
					selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
						return &model.SkillReleaseDB{
							SkillID: "skill-13b", Version: "v1", Status: interfaces.BizStatusPublished.String(), SkillContent: "demo guide",
						}, nil
					},
				},
				fileRepo:    mockFileRepo,
				assetStore:  mockAssetStore,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}
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
			reader := &skillReader{releaseRepo: &stubSkillReleaseRepo{}, Logger: logger.DefaultLogger()}

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
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			reader := &skillReader{
				releaseRepo: &stubSkillReleaseRepo{
					selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
						return &model.SkillReleaseDB{
							SkillID: "skill-14b", Version: "v1", Status: interfaces.BizStatusPublished.String(),
						}, nil
					},
				},
				fileRepo:    mockFileRepo,
				assetStore:  mockAssetStore,
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}
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

		Convey("GetSkillReleaseHistory returns published history summaries", func() {
			reader := &skillReader{
				releaseHistoryRepo: &stubSkillReleaseHistoryRepo{
					selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) ([]*model.SkillReleaseHistoryDB, error) {
						return []*model.SkillReleaseHistoryDB{
							{
								SkillID:      "skill-h1",
								Version:      "v2",
								ReleaseDesc:  "stable release",
								SkillRelease: `{"skill_id":"skill-h1","name":"demo","description":"desc","version":"v2","status":"published","category":"other_category","source":"custom","release_user":"publisher","release_time":123456789,"create_user":"creator","create_time":123,"update_user":"publisher","update_time":456}`,
							},
						}, nil
					},
				},
				Logger: logger.DefaultLogger(),
			}

			resp, err := reader.GetSkillReleaseHistory(context.Background(), &interfaces.GetSkillReleaseHistoryReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-h1",
			})

			So(err, ShouldBeNil)
			So(len(resp), ShouldEqual, 1)
			So(resp[0].SkillID, ShouldEqual, "skill-h1")
			So(resp[0].Version, ShouldEqual, "v2")
			So(resp[0].Name, ShouldEqual, "demo")
			So(resp[0].ReleaseDesc, ShouldEqual, "stable release")
			So(resp[0].ReleaseUser, ShouldEqual, "publisher")
		})

		Convey("QuerySkillMarketList filters by public access and business domain visibility", func() {
			mockReleaseRepo := mocks.NewMockISkillReleaseDB(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			mockUserMgnt := mocks.NewMockUserManagement(ctrl)
			mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
			registry := &skillRegistry{
				releaseRepo:           mockReleaseRepo,
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
			mockReleaseRepo.EXPECT().CountByWhereClause(gomock.Any(), gomock.Nil(), gomock.Any()).DoAndReturn(
				func(_ context.Context, _ interface{}, filter map[string]interface{}) (int64, error) {
					So(filter["status"], ShouldHaveSameTypeAs, "")
					So(filter["status"], ShouldEqual, interfaces.BizStatusPublished.String())
					return int64(1), nil
				},
			)
			mockReleaseRepo.EXPECT().SelectListPage(gomock.Any(), gomock.Nil(), gomock.Any(), gomock.Any(), gomock.Nil()).DoAndReturn(
				func(_ context.Context, _ interface{}, filter map[string]interface{}, _ interface{}, _ interface{}) ([]*model.SkillReleaseDB, error) {
					So(filter["status"], ShouldHaveSameTypeAs, "")
					So(filter["status"], ShouldEqual, interfaces.BizStatusPublished.String())
					return []*model.SkillReleaseDB{
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
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockUserMgnt := mocks.NewMockUserManagement(ctrl)
			mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
			registry := &skillRegistry{
				releaseRepo: &stubSkillReleaseRepo{
					selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
						return &model.SkillReleaseDB{
							SkillID:      "skill-m-detail",
							Name:         "market-visible",
							Description:  "demo-desc",
							SkillContent: "full skill markdown",
							FileManifest: `[{"rel_path":"refs/guide.md","file_type":"reference"}]`,
							Status:       interfaces.BizStatusPublished.String(),
							CreateUser:   "creator",
							UpdateUser:   "updater",
						}, nil
					},
				},
				AuthService:     mockAuthService,
				UserMgnt:        mockUserMgnt,
				CategoryManager: mockCategoryManager,
				Logger:          logger.DefaultLogger(),
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckPublicAccessPermission(gomock.Any(), gomock.Any(), "skill-m-detail", interfaces.AuthResourceTypeSkill).Return(nil)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").AnyTimes()

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

		Convey("GetSkillMarketDetail prefers published release snapshot", func() {
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockUserMgnt := mocks.NewMockUserManagement(ctrl)
			mockCategoryManager := mocks.NewMockCategoryManager(ctrl)
			registry := &skillRegistry{
				releaseRepo: &stubSkillReleaseRepo{
					selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
						return &model.SkillReleaseDB{
							SkillID:      "skill-release-detail",
							Name:         "published-name",
							Description:  "published-desc",
							SkillContent: "published content",
							Version:      "release-v1",
							Status:       interfaces.BizStatusPublished.String(),
							CreateUser:   "publisher",
							UpdateUser:   "publisher",
						}, nil
					},
				},
				AuthService:     mockAuthService,
				UserMgnt:        mockUserMgnt,
				CategoryManager: mockCategoryManager,
				Logger:          logger.DefaultLogger(),
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckPublicAccessPermission(gomock.Any(), gomock.Any(), "skill-release-detail", interfaces.AuthResourceTypeSkill).Return(nil)
			mockUserMgnt.EXPECT().GetUsersName(gomock.Any(), gomock.Any()).Return(map[string]string{
				"publisher": "Publisher",
			}, nil)
			mockCategoryManager.EXPECT().GetCategoryName(gomock.Any(), gomock.Any()).Return("").AnyTimes()

			resp, err := registry.GetSkillMarketDetail(context.Background(), &interfaces.GetSkillMarketDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-release-detail",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Name, ShouldEqual, "published-name")
			So(resp.Description, ShouldEqual, "published-desc")
			So(resp.Version, ShouldEqual, "release-v1")
		})

		Convey("GetSkillMarketDetail hides deleting skills", func() {
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				releaseRepo: &stubSkillReleaseRepo{},
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckPublicAccessPermission(gomock.Any(), gomock.Any(), "skill-m-deleting", interfaces.AuthResourceTypeSkill).Return(nil)

			resp, err := registry.GetSkillMarketDetail(context.Background(), &interfaces.GetSkillMarketDetailReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-m-deleting",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "skill not found")
		})

		Convey("GetSkillMarketDetail hides non-published skills", func() {
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			registry := &skillRegistry{
				releaseRepo: &stubSkillReleaseRepo{},
				AuthService: mockAuthService,
				Logger:      logger.DefaultLogger(),
			}
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().CheckPublicAccessPermission(gomock.Any(), gomock.Any(), "skill-m-unpublish", interfaces.AuthResourceTypeSkill).Return(nil)

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
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			releaseHistoryRepo := &stubSkillReleaseHistoryRepo{}
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				releaseHistoryRepo:    releaseHistoryRepo,
				assetStore:            mockAssetStore,
				dbTx:                  mockDBTx,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				indexSync:             mockIndexSync,
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
			releaseHistoryRepo.deleteBySkillID = func(ctx context.Context, txArg *sql.Tx, skillID string) error {
				So(txArg, ShouldEqual, tx)
				So(skillID, ShouldEqual, "skill-8")
				return nil
			}
			mockSkillRepo.EXPECT().DeleteSkillByID(gomock.Any(), tx, "skill-8").Return(nil)
			mockBusinessDomainService.EXPECT().DisassociateResource(gomock.Any(), "bd-1", "skill-8", interfaces.AuthResourceTypeSkill).Return(nil)
			mockAuthService.EXPECT().DeletePolicy(gomock.Any(), []string{"skill-8"}, interfaces.AuthResourceTypeSkill).Return(nil)
			mockIndexSync.EXPECT().DeleteSkill(gomock.Any(), "skill-8").Return(nil)

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-8",
			})

			So(err, ShouldBeNil)
		})

		Convey("DeleteSkill keeps succeeding when dataset cleanup fails", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			releaseHistoryRepo := &stubSkillReleaseHistoryRepo{}
			registry := &skillRegistry{
				skillRepo:             mockSkillRepo,
				fileRepo:              mockFileRepo,
				releaseHistoryRepo:    releaseHistoryRepo,
				assetStore:            mockAssetStore,
				dbTx:                  mockDBTx,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				indexSync:             mockIndexSync,
				Logger:                logger.DefaultLogger(),
			}

			tx, cleanup := beginTestTx(t)
			defer cleanup()

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-8b").Return(&model.SkillRepositoryDB{
				SkillID: "skill-8b", Status: interfaces.BizStatusOffline.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-8b", interfaces.AuthResourceTypeSkill).Return(nil)
			mockDBTx.EXPECT().GetTx(gomock.Any()).Return(tx, nil)
			mockSkillRepo.EXPECT().UpdateSkillDeleted(gomock.Any(), tx, "skill-8b", true, "user-1").Return(nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), tx, "skill-8b", gomock.Any()).Return(nil, nil)
			mockFileRepo.EXPECT().DeleteSkillFileBySkillID(gomock.Any(), tx, "skill-8b", gomock.Any()).Return(nil)
			releaseHistoryRepo.deleteBySkillID = func(ctx context.Context, txArg *sql.Tx, skillID string) error {
				So(txArg, ShouldEqual, tx)
				So(skillID, ShouldEqual, "skill-8b")
				return nil
			}
			mockSkillRepo.EXPECT().DeleteSkillByID(gomock.Any(), tx, "skill-8b").Return(nil)
			mockBusinessDomainService.EXPECT().DisassociateResource(gomock.Any(), "bd-1", "skill-8b", interfaces.AuthResourceTypeSkill).Return(nil)
			mockAuthService.EXPECT().DeletePolicy(gomock.Any(), []string{"skill-8b"}, interfaces.AuthResourceTypeSkill).Return(nil)
			mockIndexSync.EXPECT().DeleteSkill(gomock.Any(), "skill-8b").Return(errors.New("dataset delete failed"))

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-8b",
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
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), "skill-zip-1", interfaces.AuthResourceTypeSkill,
				interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess).Return(true, nil)
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-zip-1").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-zip-1",
				Name:         "demo-skill",
				SkillContent: "Use this skill carefully.",
				Version:      "v1",
				Status:       interfaces.BizStatusPublished.String(),
			}, nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-zip-1", "v1").Return([]*model.SkillFileIndexDB{
				{SkillID: "skill-zip-1", RelPath: SkillMD, StorageKey: "obj-skill-md"},
				{SkillID: "skill-zip-1", RelPath: "refs/guide.md", StorageKey: "obj-1"},
			}, nil)
			mockAssetStore.EXPECT().Download(gomock.Any(), &interfaces.OssObject{StorageKey: "obj-skill-md"}).Return([]byte(validSkillMarkdown()), nil)
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
			So(entries["SKILL.md"], ShouldContainSubstring, "name: demo-skill")
			So(entries["refs/guide.md"], ShouldEqual, "guide body")
		})

		Convey("DownloadSkill uses current skill snapshot without duplicating stored SKILL.md", func() {
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
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), "skill-zip-dup", interfaces.AuthResourceTypeSkill,
				interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess).Return(true, nil)
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-zip-dup").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-zip-dup",
				Name:         "demo-skill",
				Description:  "dup-desc",
				SkillContent: "Primary skill body.",
				Version:      "v1",
				Status:       interfaces.BizStatusPublished.String(),
			}, nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-zip-dup", "v1").Return([]*model.SkillFileIndexDB{
				{SkillID: "skill-zip-dup", RelPath: SkillMD, StorageKey: "obj-skill-md"},
				{SkillID: "skill-zip-dup", RelPath: "refs/guide.md", StorageKey: "obj-guide"},
			}, nil)
			mockAssetStore.EXPECT().Download(gomock.Any(), &interfaces.OssObject{StorageKey: "obj-skill-md"}).Return([]byte(validSkillMarkdown()), nil)
			mockAssetStore.EXPECT().Download(gomock.Any(), &interfaces.OssObject{StorageKey: "obj-guide"}).Return([]byte("guide body"), nil)

			resp, err := registry.DownloadSkill(context.Background(), &interfaces.DownloadSkillReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-zip-dup",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)

			zipReader, zipErr := zip.NewReader(bytes.NewReader(resp.Content), int64(len(resp.Content)))
			So(zipErr, ShouldBeNil)

			skillMDCount := 0
			entries := map[string]string{}
			for _, file := range zipReader.File {
				if file.Name == SkillMD {
					skillMDCount++
				}
				rc, openErr := file.Open()
				So(openErr, ShouldBeNil)
				body, readErr := io.ReadAll(rc)
				So(readErr, ShouldBeNil)
				_ = rc.Close()
				entries[file.Name] = string(body)
			}
			So(skillMDCount, ShouldEqual, 1)
			So(entries[SkillMD], ShouldContainSubstring, "name: demo-skill")
			So(entries["refs/guide.md"], ShouldEqual, "guide body")
		})

		Convey("DownloadSkill returns archive even when stored SKILL.md is missing", func() {
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
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), "skill-zip-missing", interfaces.AuthResourceTypeSkill,
				interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess).Return(true, nil)
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-zip-missing").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-zip-missing",
				Name:         "demo-skill",
				Description:  "missing-desc",
				SkillContent: "Primary skill body.",
				Version:      "v1",
				Status:       interfaces.BizStatusPublished.String(),
			}, nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-zip-missing", "v1").Return([]*model.SkillFileIndexDB{
				{SkillID: "skill-zip-missing", RelPath: "refs/guide.md", StorageKey: "obj-guide"},
			}, nil)
			mockAssetStore.EXPECT().Download(gomock.Any(), &interfaces.OssObject{StorageKey: "obj-guide"}).Return([]byte("guide body"), nil)

			resp, err := registry.DownloadSkill(context.Background(), &interfaces.DownloadSkillReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-zip-missing",
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
			So(entries[SkillMD], ShouldEqual, "")
			So(entries["refs/guide.md"], ShouldEqual, "guide body")
		})

		Convey("DownloadSkill builds archive from current skill snapshot version", func() {
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

			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "").Return(&interfaces.AuthAccessor{ID: "viewer"}, nil)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), "skill-release-zip", interfaces.AuthResourceTypeSkill,
				interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess).Return(true, nil)
			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-release-zip").Return(&model.SkillRepositoryDB{
				SkillID:      "skill-release-zip",
				Name:         "released-skill",
				Description:  "published-desc",
				SkillContent: "Published body.",
				Version:      "current-v3",
				Status:       interfaces.BizStatusPublished.String(),
			}, nil)
			mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-release-zip", "current-v3").Return([]*model.SkillFileIndexDB{
				{SkillID: "skill-release-zip", SkillVersion: "current-v3", RelPath: SkillMD, StorageKey: "release-skill-md"},
				{SkillID: "skill-release-zip", SkillVersion: "current-v3", RelPath: "refs/guide.md", StorageKey: "release-obj-1"},
			}, nil)
			mockAssetStore.EXPECT().Download(gomock.Any(), &interfaces.OssObject{StorageKey: "release-skill-md"}).Return([]byte(validSkillMarkdown()), nil)
			mockAssetStore.EXPECT().Download(gomock.Any(), &interfaces.OssObject{StorageKey: "release-obj-1"}).Return([]byte("released guide"), nil)

			resp, err := registry.DownloadSkill(context.Background(), &interfaces.DownloadSkillReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-release-zip",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.FileName, ShouldEqual, "released-skill.zip")

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
			So(entries["SKILL.md"], ShouldContainSubstring, "name: demo-skill")
			So(entries["refs/guide.md"], ShouldEqual, "released guide")
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

func patchTxMethods() func() {
	rollbackPatch := gomonkey.ApplyFunc((*sql.Tx).Rollback, func(*sql.Tx) error {
		return nil
	})
	commitPatch := gomonkey.ApplyFunc((*sql.Tx).Commit, func(*sql.Tx) error {
		return nil
	})
	return func() {
		rollbackPatch.Reset()
		commitPatch.Reset()
	}
}

func TestPublishSkillSnapshotKeepsNewest10HistoryVersions(t *testing.T) {
	Convey("publishSkillSnapshot replaces the same skill version with the latest snapshot", t, func() {
		histories := []*model.SkillReleaseHistoryDB{
			{ID: 1, SkillID: "skill-history", Version: "v1"},
		}
		insertedCount := 0
		deletedIDs := make([]int64, 0, 1)
		registry := &skillRegistry{
			releaseRepo: &stubSkillReleaseRepo{
				selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
					return &model.SkillReleaseDB{SkillID: skillID, Version: "v1"}, nil
				},
			},
			releaseHistoryRepo: &stubSkillReleaseHistoryRepo{
				insert: func(ctx context.Context, tx *sql.Tx, history *model.SkillReleaseHistoryDB) error {
					insertedCount++
					history.ID = 2
					histories = append([]*model.SkillReleaseHistoryDB{history}, histories...)
					return nil
				},
				selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) ([]*model.SkillReleaseHistoryDB, error) {
					return histories, nil
				},
				selectBySkillIDAndVersion: func(ctx context.Context, tx *sql.Tx, skillID, version string) (*model.SkillReleaseHistoryDB, error) {
					for _, history := range histories {
						if history.SkillID == skillID && history.Version == version {
							return history, nil
						}
					}
					return nil, nil
				},
				deleteByID: func(ctx context.Context, tx *sql.Tx, id int64) error {
					deletedIDs = append(deletedIDs, id)
					filtered := make([]*model.SkillReleaseHistoryDB, 0, len(histories))
					for _, history := range histories {
						if history.ID != id {
							filtered = append(filtered, history)
						}
					}
					histories = filtered
					return nil
				},
			},
		}

		err := registry.publishSkillSnapshot(context.Background(), nil, &model.SkillRepositoryDB{
			SkillID: "skill-history",
			Name:    "demo-skill",
			Version: "v1",
		}, "user-1")

		So(err, ShouldBeNil)
		So(insertedCount, ShouldEqual, 1)
		So(deletedIDs, ShouldResemble, []int64{1})
		So(len(histories), ShouldEqual, 1)
		So(histories[0].ID, ShouldEqual, 2)
		So(histories[0].Version, ShouldEqual, "v1")
	})

	Convey("publishSkillSnapshot replaces an existing version and keeps 10 histories", t, func() {
		histories := []*model.SkillReleaseHistoryDB{
			{ID: 10, SkillID: "skill-history", Version: "v10"},
			{ID: 9, SkillID: "skill-history", Version: "v9"},
			{ID: 8, SkillID: "skill-history", Version: "v8"},
			{ID: 7, SkillID: "skill-history", Version: "v7"},
			{ID: 6, SkillID: "skill-history", Version: "v6"},
			{ID: 5, SkillID: "skill-history", Version: "v5"},
			{ID: 4, SkillID: "skill-history", Version: "v4"},
			{ID: 3, SkillID: "skill-history", Version: "v3"},
			{ID: 2, SkillID: "skill-history", Version: "v2"},
			{ID: 1, SkillID: "skill-history", Version: "v1"},
		}
		insertedCount := 0
		deletedIDs := make([]int64, 0, 1)
		registry := &skillRegistry{
			releaseRepo: &stubSkillReleaseRepo{
				selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
					return &model.SkillReleaseDB{SkillID: skillID, Version: "v10"}, nil
				},
			},
			releaseHistoryRepo: &stubSkillReleaseHistoryRepo{
				insert: func(ctx context.Context, tx *sql.Tx, history *model.SkillReleaseHistoryDB) error {
					insertedCount++
					history.ID = 11
					histories = append([]*model.SkillReleaseHistoryDB{history}, histories...)
					return nil
				},
				selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) ([]*model.SkillReleaseHistoryDB, error) {
					return histories, nil
				},
				selectBySkillIDAndVersion: func(ctx context.Context, tx *sql.Tx, skillID, version string) (*model.SkillReleaseHistoryDB, error) {
					for _, history := range histories {
						if history.SkillID == skillID && history.Version == version {
							return history, nil
						}
					}
					return nil, nil
				},
				deleteByID: func(ctx context.Context, tx *sql.Tx, id int64) error {
					deletedIDs = append(deletedIDs, id)
					filtered := make([]*model.SkillReleaseHistoryDB, 0, len(histories))
					for _, history := range histories {
						if history.ID != id {
							filtered = append(filtered, history)
						}
					}
					histories = filtered
					return nil
				},
			},
		}

		err := registry.publishSkillSnapshot(context.Background(), nil, &model.SkillRepositoryDB{
			SkillID: "skill-history",
			Name:    "demo-skill",
			Version: "v10",
		}, "user-1")

		So(err, ShouldBeNil)
		So(insertedCount, ShouldEqual, 1)
		So(deletedIDs, ShouldResemble, []int64{10})
		So(len(histories), ShouldEqual, 10)
		So(histories[0].ID, ShouldEqual, 11)
		So(histories[0].Version, ShouldEqual, "v10")
		for _, history := range histories {
			So(history.ID, ShouldNotEqual, 10)
		}
	})

	Convey("publishSkillSnapshot keeps the newest 10 distinct skill versions", t, func() {
		histories := []*model.SkillReleaseHistoryDB{
			{ID: 10, SkillID: "skill-history", Version: "v10"},
			{ID: 9, SkillID: "skill-history", Version: "v9"},
			{ID: 8, SkillID: "skill-history", Version: "v8"},
			{ID: 7, SkillID: "skill-history", Version: "v7"},
			{ID: 6, SkillID: "skill-history", Version: "v6"},
			{ID: 5, SkillID: "skill-history", Version: "v5"},
			{ID: 4, SkillID: "skill-history", Version: "v4"},
			{ID: 3, SkillID: "skill-history", Version: "v3"},
			{ID: 2, SkillID: "skill-history", Version: "v2"},
			{ID: 1, SkillID: "skill-history", Version: "v1"},
		}
		deletedIDs := make([]int64, 0, 1)
		registry := &skillRegistry{
			releaseRepo: &stubSkillReleaseRepo{
				selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillReleaseDB, error) {
					return &model.SkillReleaseDB{SkillID: skillID, Version: "v10"}, nil
				},
			},
			releaseHistoryRepo: &stubSkillReleaseHistoryRepo{
				insert: func(ctx context.Context, tx *sql.Tx, history *model.SkillReleaseHistoryDB) error {
					history.ID = 11
					histories = append([]*model.SkillReleaseHistoryDB{history}, histories...)
					return nil
				},
				selectBySkillID: func(ctx context.Context, tx *sql.Tx, skillID string) ([]*model.SkillReleaseHistoryDB, error) {
					return histories, nil
				},
				selectBySkillIDAndVersion: func(ctx context.Context, tx *sql.Tx, skillID, version string) (*model.SkillReleaseHistoryDB, error) {
					for _, history := range histories {
						if history.SkillID == skillID && history.Version == version {
							return history, nil
						}
					}
					return nil, nil
				},
				deleteByID: func(ctx context.Context, tx *sql.Tx, id int64) error {
					deletedIDs = append(deletedIDs, id)
					filtered := make([]*model.SkillReleaseHistoryDB, 0, len(histories))
					for _, history := range histories {
						if history.ID != id {
							filtered = append(filtered, history)
						}
					}
					histories = filtered
					return nil
				},
			},
		}

		err := registry.publishSkillSnapshot(context.Background(), nil, &model.SkillRepositoryDB{
			SkillID: "skill-history",
			Name:    "demo-skill",
			Version: "v11",
		}, "user-1")

		So(err, ShouldBeNil)
		So(len(histories), ShouldEqual, 10)
		So(histories[0].ID, ShouldEqual, 11)
		So(histories[0].Version, ShouldEqual, "v11")
		So(deletedIDs, ShouldResemble, []int64{1})
		for _, history := range histories {
			So(history.Version, ShouldNotEqual, "v1")
		}
	})
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
			Status:       interfaces.BizStatusPublished.String(),
		}, nil)
		mockFileRepo.EXPECT().SelectSkillFileBySkillID(gomock.Any(), gomock.Nil(), "skill-exec-1", "v1").Return([]*model.SkillFileIndexDB{
			{
				SkillID:    "skill-exec-1",
				RelPath:    SkillMD,
				StorageKey: "obj-skill-md",
			},
			{
				SkillID:    "skill-exec-1",
				RelPath:    "refs/guide.md",
				StorageKey: "obj-1",
			},
		}, nil)
		mockAssetStore.EXPECT().Download(gomock.Any(), &interfaces.OssObject{StorageKey: "obj-skill-md"}).Return([]byte(validSkillMarkdown()), nil)
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
