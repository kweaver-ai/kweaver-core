package skill

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestSkillManagementReader(t *testing.T) {
	Convey("SkillManagementReader", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		validRepo := &model.SkillRepositoryDB{
			SkillID:      "skill-1",
			Version:      "v1",
			Name:         "test-skill",
			Description:  "test description",
			Status:       interfaces.BizStatusEditing.String(),
			Source:       "custom",
			FileManifest: `[{"rel_path":"scripts/main.py","file_type":"script","size":1024,"mime_type":"text/x-python"}]`,
		}

		Convey("GetManagementContent returns SKILL.md URL and files for zip registration", func() {
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return validRepo, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
				Logger:                logger.DefaultLogger(),
			}

			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-1", "v1", SkillMD).
				Return(&model.SkillFileIndexDB{
					SkillID:      "skill-1",
					SkillVersion: "v1",
					RelPath:      SkillMD,
					StorageKey:   testBuildObjectKey("skill-1", "v1", SkillMD),
				}, nil)
			mockAssetStore.EXPECT().GetDownloadURL(gomock.Any(), gomock.Any()).
				Return("https://download/skill-1/SKILL.md", nil)

			resp, err := reader.GetManagementContent(context.Background(), &interfaces.GetManagementContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-1",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.SkillID, ShouldEqual, "skill-1")
			So(resp.Name, ShouldEqual, "test-skill")
			So(resp.URL, ShouldEqual, "https://download/skill-1/SKILL.md")
			So(resp.FileType, ShouldEqual, "zip")
			So(len(resp.Files), ShouldEqual, 1)
			So(resp.Files[0].RelPath, ShouldEqual, "scripts/main.py")
			So(resp.Files[0].FileType, ShouldEqual, "script")
			So(resp.Content, ShouldEqual, "") // url(default) mode: content is empty
		})

		Convey("GetManagementContent returns empty URL for content registration without OSS file", func() {
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-2", "v1", SkillMD).
				Return(nil, nil)
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return &model.SkillRepositoryDB{
							SkillID:      "skill-2",
							Version:      "v1",
							Name:         "content-skill",
							Description:  "content desc",
							Status:       interfaces.BizStatusUnpublish.String(),
							Source:       "custom",
							SkillContent: "some body text",
						}, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mocks.NewMockskillAssetStore(ctrl),
				AuthService:           mocks.NewMockIAuthorizationService(ctrl),
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			resp, err := reader.GetManagementContent(context.Background(), &interfaces.GetManagementContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-2",
				ResponseMode:     "content",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.URL, ShouldEqual, "")
			So(resp.Content, ShouldEqual, "some body text")
			So(resp.FileType, ShouldEqual, "content")
			So(resp.Files, ShouldNotBeNil)
			So(len(resp.Files), ShouldEqual, 0)
		})

		Convey("GetManagementContent returns 404 for deleted skill", func() {
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return &model.SkillRepositoryDB{
							SkillID:   "skill-deleted",
							IsDeleted: true,
						}, nil
					},
				},
				fileRepo:              mocks.NewMockISkillFileIndex(ctrl),
				assetStore:            mocks.NewMockskillAssetStore(ctrl),
				AuthService:           mocks.NewMockIAuthorizationService(ctrl),
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			resp, err := reader.GetManagementContent(context.Background(), &interfaces.GetManagementContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-deleted",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*errors.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("GetManagementContent checks view permission for public API", func() {
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return validRepo, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				AuthService:           mockAuthService,
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			ctx := common.SetPublicAPIToCtx(context.Background(), true)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-view").
				Return(&interfaces.AuthAccessor{ID: "user-view"}, nil)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), "skill-1",
				interfaces.AuthResourceTypeSkill,
				interfaces.AuthOperationTypeView, interfaces.AuthOperationTypeModify).
				Return(true, nil)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-1", "v1", SkillMD).
				Return(&model.SkillFileIndexDB{
					SkillID: "skill-1", SkillVersion: "v1", RelPath: SkillMD,
					StorageKey: testBuildObjectKey("skill-1", "v1", SkillMD),
				}, nil)
			mockAssetStore.EXPECT().GetDownloadURL(gomock.Any(), gomock.Any()).
				Return("https://download/skill-1/SKILL.md", nil)

			resp, err := reader.GetManagementContent(ctx, &interfaces.GetManagementContentReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-view",
				SkillID:          "skill-1",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.URL, ShouldEqual, "https://download/skill-1/SKILL.md")
		})

		Convey("GetManagementContent returns 403 for unauthorized public API", func() {
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return validRepo, nil
					},
				},
				fileRepo:              mocks.NewMockISkillFileIndex(ctrl),
				assetStore:            mocks.NewMockskillAssetStore(ctrl),
				AuthService:           mockAuthService,
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			ctx := common.SetPublicAPIToCtx(context.Background(), true)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "no-perm-user").
				Return(&interfaces.AuthAccessor{ID: "no-perm-user"}, nil)
			mockAuthService.EXPECT().OperationCheckAny(gomock.Any(), gomock.Any(), "skill-1",
				interfaces.AuthResourceTypeSkill,
				interfaces.AuthOperationTypeView, interfaces.AuthOperationTypeModify).
				Return(false, nil)

			resp, err := reader.GetManagementContent(ctx, &interfaces.GetManagementContentReq{
				BusinessDomainID: "bd-1",
				UserID:           "no-perm-user",
				SkillID:          "skill-1",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*errors.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("GetManagementContent returns 404 for non-existent skill", func() {
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return nil, nil
					},
				},
				fileRepo:              mocks.NewMockISkillFileIndex(ctrl),
				assetStore:            mocks.NewMockskillAssetStore(ctrl),
				AuthService:           mocks.NewMockIAuthorizationService(ctrl),
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			resp, err := reader.GetManagementContent(context.Background(), &interfaces.GetManagementContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-nonexistent",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*errors.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("GetManagementContent internal API skips auth check", func() {
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return validRepo, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				AuthService:           mocks.NewMockIAuthorizationService(ctrl),
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-1", "v1", SkillMD).
				Return(&model.SkillFileIndexDB{
					SkillID: "skill-1", SkillVersion: "v1", RelPath: SkillMD,
					StorageKey: testBuildObjectKey("skill-1", "v1", SkillMD),
				}, nil)
			mockAssetStore.EXPECT().GetDownloadURL(gomock.Any(), gomock.Any()).
				Return("https://download/skill-1/SKILL.md", nil)

			resp, err := reader.GetManagementContent(context.Background(), &interfaces.GetManagementContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-1",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.URL, ShouldEqual, "https://download/skill-1/SKILL.md")
		})

		Convey("GetManagementContent zip registration + content mode returns OSS content", func() {
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return validRepo, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				AuthService:           mockAuthService,
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-1", "v1", SkillMD).
				Return(&model.SkillFileIndexDB{
					SkillID: "skill-1", SkillVersion: "v1", RelPath: SkillMD,
					StorageKey: testBuildObjectKey("skill-1", "v1", SkillMD),
				}, nil)
			mockAssetStore.EXPECT().Download(gomock.Any(), gomock.Any()).
				Return([]byte("# SKILL.md from OSS"), nil)

			resp, err := reader.GetManagementContent(context.Background(), &interfaces.GetManagementContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-1",
				ResponseMode:     "content",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.URL, ShouldEqual, "") // content mode: url is empty
			So(resp.Content, ShouldEqual, "# SKILL.md from OSS")
			So(resp.FileType, ShouldEqual, "zip")
		})

		Convey("GetManagementContent content registration + content mode returns DB content", func() {
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-2", "v1", SkillMD).
				Return(nil, nil)
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return &model.SkillRepositoryDB{
							SkillID:      "skill-2",
							Version:      "v1",
							Name:         "content-skill",
							Description:  "content desc",
							Status:       interfaces.BizStatusUnpublish.String(),
							Source:       "custom",
							SkillContent: "some body text",
						}, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mocks.NewMockskillAssetStore(ctrl),
				AuthService:           mocks.NewMockIAuthorizationService(ctrl),
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			resp, err := reader.GetManagementContent(context.Background(), &interfaces.GetManagementContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-2",
				ResponseMode:     "content",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.URL, ShouldEqual, "")
			So(resp.Content, ShouldEqual, "some body text")
			So(resp.FileType, ShouldEqual, "content")
		})

		Convey("GetManagementContent content registration + url mode returns empty content", func() {
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-2", "v1", SkillMD).
				Return(nil, nil)
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return &model.SkillRepositoryDB{
							SkillID:      "skill-2",
							Version:      "v1",
							Name:         "content-skill",
							Description:  "content desc",
							Status:       interfaces.BizStatusUnpublish.String(),
							Source:       "custom",
							SkillContent: "some body text",
						}, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mocks.NewMockskillAssetStore(ctrl),
				AuthService:           mocks.NewMockIAuthorizationService(ctrl),
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			resp, err := reader.GetManagementContent(context.Background(), &interfaces.GetManagementContentReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-2",
				ResponseMode:     "url",
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.URL, ShouldEqual, "")
			So(resp.Content, ShouldEqual, "") // url mode: content is empty
			So(resp.FileType, ShouldEqual, "content")
		})

		Convey("ReadManagementFile returns presigned URL for valid path", func() {
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return validRepo, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mockAssetStore,
				AuthService:           mocks.NewMockIAuthorizationService(ctrl),
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-1", "v1", "scripts/main.py").
				Return(&model.SkillFileIndexDB{
					SkillID: "skill-1", SkillVersion: "v1",
					RelPath: "scripts/main.py", FileType: "script", MimeType: "text/x-python", Size: 1024,
					StorageKey: testBuildObjectKey("skill-1", "v1", "scripts/main.py"),
				}, nil)
			mockAssetStore.EXPECT().GetDownloadURL(gomock.Any(), gomock.Any()).
				Return("https://download/skill-1/scripts/main.py", nil)

			resp, err := reader.ReadManagementFile(context.Background(), &interfaces.ReadManagementFileReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-1",
				RelPath:          "scripts/main.py",
			})

			So(err, ShouldBeNil)
			So(resp.URL, ShouldEqual, "https://download/skill-1/scripts/main.py")
			So(resp.MimeType, ShouldEqual, "text/x-python")
			So(resp.FileType, ShouldEqual, "script")
			So(resp.Size, ShouldEqual, 1024)
		})

		Convey("ReadManagementFile rejects path traversal", func() {
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return validRepo, nil
					},
				},
				fileRepo:              mocks.NewMockISkillFileIndex(ctrl),
				assetStore:            mocks.NewMockskillAssetStore(ctrl),
				AuthService:           mocks.NewMockIAuthorizationService(ctrl),
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			resp, err := reader.ReadManagementFile(context.Background(), &interfaces.ReadManagementFileReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-1",
				RelPath:          "../../etc/passwd",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*errors.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("ReadManagementFile returns 404 for non-existent file", func() {
			mockFileRepo := mocks.NewMockISkillFileIndex(ctrl)
			reader := &skillManagementReader{
				skillRepo: &stubSkillRepo{
					selectByID: func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
						return validRepo, nil
					},
				},
				fileRepo:              mockFileRepo,
				assetStore:            mocks.NewMockskillAssetStore(ctrl),
				AuthService:           mocks.NewMockIAuthorizationService(ctrl),
				BusinessDomainService: mocks.NewMockIBusinessDomainService(ctrl),
				Logger:                logger.DefaultLogger(),
			}

			mockFileRepo.EXPECT().SelectSkillFileByPath(gomock.Any(), gomock.Nil(), "skill-1", "v1", "missing.py").
				Return(nil, nil)

			resp, err := reader.ReadManagementFile(context.Background(), &interfaces.ReadManagementFileReq{
				BusinessDomainID: "bd-1",
				SkillID:          "skill-1",
				RelPath:          "missing.py",
			})

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*errors.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})
	})
}

// stubSkillRepo provides a minimal ISkillRepository stub for testing
type stubSkillRepo struct {
	model.ISkillRepository
	selectByID func(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error)
}

func (s *stubSkillRepo) SelectSkillByID(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
	if s.selectByID != nil {
		return s.selectByID(ctx, tx, skillID)
	}
	return nil, nil
}

func TestDetectSkillFileType(t *testing.T) {
	Convey("detectSkillFileType", t, func() {
		Convey("zip registration with multiple files in manifest", func() {
			ft := detectSkillFileType(&model.SkillRepositoryDB{
				FileManifest: `[{"rel_path":"SKILL.md"},{"rel_path":"scripts/main.py"}]`,
			})
			So(ft, ShouldEqual, "zip")
		})
		Convey("content registration without manifest", func() {
			ft := detectSkillFileType(&model.SkillRepositoryDB{
				SkillContent: "body text",
			})
			So(ft, ShouldEqual, "content")
		})
		Convey("content registration with SKILL.md-only manifest (FR-5)", func() {
			ft := detectSkillFileType(&model.SkillRepositoryDB{
				SkillContent: "body text",
				FileManifest: `[{"rel_path":"SKILL.md","file_type":"reference","size":42,"mime_type":"text/markdown"}]`,
			})
			So(ft, ShouldEqual, "content")
		})
		Convey("no manifest and no content defaults to content", func() {
			ft := detectSkillFileType(&model.SkillRepositoryDB{})
			So(ft, ShouldEqual, "content")
		})
	})
}

func TestBuildArchiveFromFiles(t *testing.T) {
	Convey("buildArchiveFromFiles", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("builds ZIP from files successfully", func() {
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			skill := &model.SkillRepositoryDB{
				SkillID: "skill-1",
				Name:    "test-skill",
				Version: "v1",
			}
			files := []*model.SkillFileIndexDB{
				{
					RelPath:    "SKILL.md",
					StorageKey: "key/skill-1/v1/SKILL.md",
				},
			}

			mockAssetStore.EXPECT().Download(gomock.Any(), gomock.Any()).
				Return([]byte("skill content"), nil)

			resultSkill, zipName, content, err := buildArchiveFromFiles(context.Background(), mockAssetStore, skill, files)
			So(err, ShouldBeNil)
			So(resultSkill, ShouldEqual, skill)
			So(zipName, ShouldEqual, "test-skill.zip")
			So(len(content), ShouldBeGreaterThan, 0)
		})

		Convey("returns error when OSS download fails", func() {
			mockAssetStore := mocks.NewMockskillAssetStore(ctrl)
			skill := &model.SkillRepositoryDB{
				SkillID: "skill-1", Name: "test-skill", Version: "v1",
			}
			files := []*model.SkillFileIndexDB{
				{RelPath: "SKILL.md", StorageKey: "key1"},
				{RelPath: "scripts/main.py", StorageKey: "key2"},
			}

			mockAssetStore.EXPECT().Download(gomock.Any(), gomock.Any()).
				Return(nil, fmt.Errorf("oss download failed"))

			resultSkill, zipName, content, err := buildArchiveFromFiles(context.Background(), mockAssetStore, skill, files)
			So(err, ShouldNotBeNil)
			So(resultSkill, ShouldBeNil)
			So(zipName, ShouldEqual, "")
			So(content, ShouldBeNil)
		})
	})
}
