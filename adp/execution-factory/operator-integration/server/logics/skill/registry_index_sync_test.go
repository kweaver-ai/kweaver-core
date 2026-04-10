package skill

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestSkillRegistryIndexSync(t *testing.T) {
	Convey("SkillRegistry index sync", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("publish updates status then upserts index", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				indexSync:   mockIndexSync,
				Logger:      logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-publish").Return(&model.SkillRepositoryDB{
				SkillID: "skill-publish", Name: "demo-skill", Status: interfaces.BizStatusUnpublish.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), "skill-publish", interfaces.AuthResourceTypeSkill).Return(nil)
			mockSkillRepo.EXPECT().SelectSkillByName(gomock.Any(), gomock.Nil(), "demo-skill", []string{interfaces.BizStatusPublished.String()}).Return(false, nil, nil)
			mockSkillRepo.EXPECT().UpdateSkillStatus(gomock.Any(), gomock.Nil(), "skill-publish", interfaces.BizStatusPublished.String(), "user-1").Return(nil)
			mockIndexSync.EXPECT().UpsertSkill(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, skill *model.SkillRepositoryDB) error {
				So(skill.SkillID, ShouldEqual, "skill-publish")
				return nil
			})

			resp, err := registry.UpdateSkillStatus(context.Background(), &interfaces.UpdateSkillStatusReq{
				UserID:  "user-1",
				SkillID: "skill-publish",
				Status:  interfaces.BizStatusPublished,
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Status, ShouldEqual, interfaces.BizStatusPublished)
		})

		Convey("publish still succeeds when deferred index sync fails", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				indexSync:   mockIndexSync,
				Logger:      logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-publish-fail").Return(&model.SkillRepositoryDB{
				SkillID: "skill-publish-fail", Name: "demo-skill", Status: interfaces.BizStatusUnpublish.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckPublishPermission(gomock.Any(), gomock.Any(), "skill-publish-fail", interfaces.AuthResourceTypeSkill).Return(nil)
			mockSkillRepo.EXPECT().SelectSkillByName(gomock.Any(), gomock.Nil(), "demo-skill", []string{interfaces.BizStatusPublished.String()}).Return(false, nil, nil)
			mockSkillRepo.EXPECT().UpdateSkillStatus(gomock.Any(), gomock.Nil(), "skill-publish-fail", interfaces.BizStatusPublished.String(), "user-1").Return(nil)
			mockIndexSync.EXPECT().UpsertSkill(gomock.Any(), gomock.Any()).Return(errors.New("vega write failed"))

			resp, err := registry.UpdateSkillStatus(context.Background(), &interfaces.UpdateSkillStatusReq{
				UserID:  "user-1",
				SkillID: "skill-publish-fail",
				Status:  interfaces.BizStatusPublished,
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Status, ShouldEqual, interfaces.BizStatusPublished)
		})

		Convey("offline updates status then deletes index", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				indexSync:   mockIndexSync,
				Logger:      logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-offline").Return(&model.SkillRepositoryDB{
				SkillID: "skill-offline", Status: interfaces.BizStatusPublished.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckUnpublishPermission(gomock.Any(), gomock.Any(), "skill-offline", interfaces.AuthResourceTypeSkill).Return(nil)
			mockSkillRepo.EXPECT().UpdateSkillStatus(gomock.Any(), gomock.Nil(), "skill-offline", interfaces.BizStatusOffline.String(), "user-1").Return(nil)
			mockIndexSync.EXPECT().DeleteSkill(gomock.Any(), "skill-offline").Return(nil)

			resp, err := registry.UpdateSkillStatus(context.Background(), &interfaces.UpdateSkillStatusReq{
				UserID:  "user-1",
				SkillID: "skill-offline",
				Status:  interfaces.BizStatusOffline,
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Status, ShouldEqual, interfaces.BizStatusOffline)
		})

		Convey("offline still succeeds when deferred index delete fails", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				AuthService: mockAuthService,
				indexSync:   mockIndexSync,
				Logger:      logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-offline-fail").Return(&model.SkillRepositoryDB{
				SkillID: "skill-offline-fail", Status: interfaces.BizStatusPublished.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckUnpublishPermission(gomock.Any(), gomock.Any(), "skill-offline-fail", interfaces.AuthResourceTypeSkill).Return(nil)
			mockSkillRepo.EXPECT().UpdateSkillStatus(gomock.Any(), gomock.Nil(), "skill-offline-fail", interfaces.BizStatusOffline.String(), "user-1").Return(nil)
			mockIndexSync.EXPECT().DeleteSkill(gomock.Any(), "skill-offline-fail").Return(errors.New("vega delete failed"))

			resp, err := registry.UpdateSkillStatus(context.Background(), &interfaces.UpdateSkillStatusReq{
				UserID:  "user-1",
				SkillID: "skill-offline-fail",
				Status:  interfaces.BizStatusOffline,
			})

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Status, ShouldEqual, interfaces.BizStatusOffline)
		})

		Convey("delete stops before cleanup when index delete fails", func() {
			mockSkillRepo := mocks.NewMockISkillRepository(ctrl)
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockDBTx := mocks.NewMockDBTx(ctrl)
			mockIndexSync := mocks.NewMockSkillIndexSyncService(ctrl)
			registry := &skillRegistry{
				skillRepo:   mockSkillRepo,
				dbTx:        mockDBTx,
				AuthService: mockAuthService,
				indexSync:   mockIndexSync,
				Logger:      logger.DefaultLogger(),
			}

			mockSkillRepo.EXPECT().SelectSkillByID(gomock.Any(), gomock.Nil(), "skill-delete-fail").Return(&model.SkillRepositoryDB{
				SkillID: "skill-delete-fail", Status: interfaces.BizStatusOffline.String(),
			}, nil)
			mockAuthService.EXPECT().GetAccessor(gomock.Any(), "user-1").Return(&interfaces.AuthAccessor{ID: "user-1"}, nil)
			mockAuthService.EXPECT().CheckDeletePermission(gomock.Any(), gomock.Any(), "skill-delete-fail", interfaces.AuthResourceTypeSkill).Return(nil)
			mockIndexSync.EXPECT().DeleteSkill(gomock.Any(), "skill-delete-fail").Return(errors.New("vega delete failed"))

			err := registry.DeleteSkill(context.Background(), &interfaces.DeleteSkillReq{
				BusinessDomainID: "bd-1",
				UserID:           "user-1",
				SkillID:          "skill-delete-fail",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "vega delete failed")
		})
	})
}
