package releasesvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
)

// --- GetPublishInfo 缺失分支补全 ---

// TestGetPublishInfo_GetByReleaseIDError 分类关联查询出错
func TestGetPublishInfo_GetByReleaseIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockCategoryRelRepo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:                service.NewSvcBase(),
		agentConfigRepo:        mockAgentRepo,
		releaseRepo:            mockReleaseRepo,
		releaseCategoryRelRepo: mockCategoryRelRepo,
	}

	ctx := context.Background()
	mockAgentRepo.EXPECT().ExistsByID(ctx, "agent-1").Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(ctx, "agent-1").Return(&dapo.ReleasePO{
		ID:      "rel-1",
		AgentID: "agent-1",
	}, nil)
	mockCategoryRelRepo.EXPECT().GetByReleaseID(ctx, "rel-1").Return(nil, errors.New("category rel db error"))

	_, err := svc.GetPublishInfo(ctx, "agent-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get category relations failed")
}

// TestGetPublishInfo_CategoryNameMapError 分类名称映射查询出错
func TestGetPublishInfo_CategoryNameMapError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockCategoryRelRepo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
	mockCategoryRepo := idbaccessmock.NewMockICategoryRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:                service.NewSvcBase(),
		agentConfigRepo:        mockAgentRepo,
		releaseRepo:            mockReleaseRepo,
		releaseCategoryRelRepo: mockCategoryRelRepo,
		categoryRepo:           mockCategoryRepo,
	}

	ctx := context.Background()
	mockAgentRepo.EXPECT().ExistsByID(ctx, "agent-2").Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(ctx, "agent-2").Return(&dapo.ReleasePO{
		ID:      "rel-2",
		AgentID: "agent-2",
	}, nil)
	mockCategoryRelRepo.EXPECT().GetByReleaseID(ctx, "rel-2").Return([]*dapo.ReleaseCategoryRelPO{
		{CategoryID: "cat-1"},
	}, nil)
	mockCategoryRepo.EXPECT().GetIDNameMap(ctx, []string{"cat-1"}).Return(nil, errors.New("category name db error"))

	_, err := svc.GetPublishInfo(ctx, "agent-2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get category name map failed")
}

// TestGetPublishInfo_WithCategoriesSuccess 含分类信息，无权限控制，成功
func TestGetPublishInfo_WithCategoriesSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockCategoryRelRepo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
	mockCategoryRepo := idbaccessmock.NewMockICategoryRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:                service.NewSvcBase(),
		agentConfigRepo:        mockAgentRepo,
		releaseRepo:            mockReleaseRepo,
		releaseCategoryRelRepo: mockCategoryRelRepo,
		categoryRepo:           mockCategoryRepo,
	}

	ctx := context.Background()
	isPmsCtrl := 0

	mockAgentRepo.EXPECT().ExistsByID(ctx, "agent-3").Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(ctx, "agent-3").Return(&dapo.ReleasePO{
		ID:        "rel-3",
		AgentID:   "agent-3",
		AgentDesc: "带分类的智能体",
		IsPmsCtrl: &isPmsCtrl,
	}, nil)
	mockCategoryRelRepo.EXPECT().GetByReleaseID(ctx, "rel-3").Return([]*dapo.ReleaseCategoryRelPO{
		{CategoryID: "cat-1"},
		{CategoryID: "cat-2"},
	}, nil)
	mockCategoryRepo.EXPECT().GetIDNameMap(ctx, []string{"cat-1", "cat-2"}).Return(map[string]string{
		"cat-1": "分类一",
		"cat-2": "分类二",
	}, nil)

	resp, err := svc.GetPublishInfo(ctx, "agent-3")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "带分类的智能体", resp.Description)
	assert.Len(t, resp.Categories, 2)
	assert.Nil(t, resp.PmsControl)
}

// TestGetPublishInfo_WithPermissionsGetError 权限控制开启但 repo 查询出错
func TestGetPublishInfo_WithPermissionsGetError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockCategoryRelRepo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)

	isPmsCtrl := 1
	svc := &releaseSvc{
		SvcBase:                service.NewSvcBase(),
		agentConfigRepo:        mockAgentRepo,
		releaseRepo:            mockReleaseRepo,
		releaseCategoryRelRepo: mockCategoryRelRepo,
		releasePermissionRepo:  mockPermRepo,
	}

	ctx := context.Background()
	mockAgentRepo.EXPECT().ExistsByID(ctx, "agent-4").Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(ctx, "agent-4").Return(&dapo.ReleasePO{
		ID:        "rel-4",
		AgentID:   "agent-4",
		IsPmsCtrl: &isPmsCtrl,
	}, nil)
	mockCategoryRelRepo.EXPECT().GetByReleaseID(ctx, "rel-4").Return([]*dapo.ReleaseCategoryRelPO{}, nil)
	mockPermRepo.EXPECT().GetByReleaseID(ctx, "rel-4").Return(nil, errors.New("perm db error"))

	_, err := svc.GetPublishInfo(ctx, "agent-4")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get permissions failed")
}

// TestGetPublishInfo_WithPermissionsSuccess 权限控制开启，成功获取（本地开发模式下）
func TestGetPublishInfo_WithPermissionsSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockCategoryRelRepo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	isPmsCtrl := 1
	svc := &releaseSvc{
		SvcBase:                service.NewSvcBase(),
		agentConfigRepo:        mockAgentRepo,
		releaseRepo:            mockReleaseRepo,
		releaseCategoryRelRepo: mockCategoryRelRepo,
		releasePermissionRepo:  mockPermRepo,
		umHttp:                 mockUm,
	}

	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck
	mockAgentRepo.EXPECT().ExistsByID(ctx, "agent-5").Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(ctx, "agent-5").Return(&dapo.ReleasePO{
		ID:        "rel-5",
		AgentID:   "agent-5",
		IsPmsCtrl: &isPmsCtrl,
	}, nil)
	mockCategoryRelRepo.EXPECT().GetByReleaseID(ctx, "rel-5").Return([]*dapo.ReleaseCategoryRelPO{}, nil)
	// 返回有权限数据
	mockPermRepo.EXPECT().GetByReleaseID(ctx, "rel-5").Return([]*dapo.ReleasePermissionPO{
		{ObjectId: "user-1", ObjectType: cenum.PmsTargetObjTypeUser},
		{ObjectId: "role-1", ObjectType: cenum.PmsTargetObjTypeRole},
		{ObjectId: "group-1", ObjectType: cenum.PmsTargetObjTypeUserGroup},
		{ObjectId: "dept-1", ObjectType: cenum.PmsTargetObjTypeDep},
		{ObjectId: "app-1", ObjectType: cenum.PmsTargetObjTypeAppAccount},
	}, nil)

	// 非所有环境都处于 local 模式，所以可能调用 GetOsnNames
	osnMap := umtypes.NewOsnInfoMapS()
	osnMap.UserNameMap["user-1"] = "User One"
	mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).Return(osnMap, nil).AnyTimes()

	resp, err := svc.GetPublishInfo(ctx, "agent-5")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.PmsControl)
}

// TestGetPublishInfo_WithPermissions_EmptyList 有权限控制标志但权限为空时，不调用 genPmsControlResp
func TestGetPublishInfo_WithPermissions_EmptyList(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockCategoryRelRepo := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)

	isPmsCtrl := 1
	svc := &releaseSvc{
		SvcBase:                service.NewSvcBase(),
		agentConfigRepo:        mockAgentRepo,
		releaseRepo:            mockReleaseRepo,
		releaseCategoryRelRepo: mockCategoryRelRepo,
		releasePermissionRepo:  mockPermRepo,
	}

	ctx := context.Background()
	mockAgentRepo.EXPECT().ExistsByID(ctx, "agent-6").Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(ctx, "agent-6").Return(&dapo.ReleasePO{
		ID:        "rel-6",
		AgentID:   "agent-6",
		IsPmsCtrl: &isPmsCtrl,
	}, nil)
	mockCategoryRelRepo.EXPECT().GetByReleaseID(ctx, "rel-6").Return([]*dapo.ReleaseCategoryRelPO{}, nil)
	// 权限列表为空
	mockPermRepo.EXPECT().GetByReleaseID(ctx, "rel-6").Return([]*dapo.ReleasePermissionPO{}, nil)

	resp, err := svc.GetPublishInfo(ctx, "agent-6")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	// PmsControl 已通过 NewPmsControlResp 初始化，但为空列表
	assert.NotNil(t, resp.PmsControl)
}

// --- genPmsControlResp 缺失分支：GetOsnNames 失败 ---

// TestGenPmsControlResp_LocalDev_AllTypes 非本地环境调用 GetOsnNames 出错
func TestGenPmsControlResp_LocalDev_AllTypes(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &releaseSvc{
		SvcBase: service.NewSvcBase(),
		umHttp:  mockUm,
	}

	// 构建混合类型权限列表
	permissions := []*dapo.ReleasePermissionPO{
		{ObjectId: "u1", ObjectType: cenum.PmsTargetObjTypeUser},
		{ObjectId: "r1", ObjectType: cenum.PmsTargetObjTypeRole},
		{ObjectId: "g1", ObjectType: cenum.PmsTargetObjTypeUserGroup},
		{ObjectId: "d1", ObjectType: cenum.PmsTargetObjTypeDep},
		{ObjectId: "a1", ObjectType: cenum.PmsTargetObjTypeAppAccount},
	}

	// 支持两种环境分支
	osnMap := umtypes.NewOsnInfoMapS()
	osnMap.UserNameMap["u1"] = "u1_name"
	mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).Return(osnMap, nil).AnyTimes()

	// 本地开发模式：_name 后缀生成或通过 mock 返回
	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck
	resp, err := svc.genPmsControlResp(ctx, permissions)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	// 每种类型各一个
	assert.Len(t, resp.Users, 1)
	assert.Equal(t, "u1", resp.Users[0].UserID)
	assert.Equal(t, "u1_name", resp.Users[0].Username)
	assert.Len(t, resp.Roles, 1)
	assert.Len(t, resp.UserGroups, 1)
	assert.Len(t, resp.Departments, 1)
	assert.Len(t, resp.AppAccounts, 1)
}

// TestGenPmsControlResp_GetOsnNames_Success 非本地环境，GetOsnNames 成功，用户名不在 map → unknownUser
func TestGenPmsControlResp_GetOsnNames_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &releaseSvc{
		SvcBase: service.NewSvcBase(),
		umHttp:  mockUm,
	}

	permissions := []*dapo.ReleasePermissionPO{
		{ObjectId: "u-missing", ObjectType: cenum.PmsTargetObjTypeUser},
		{ObjectId: "u-found", ObjectType: cenum.PmsTargetObjTypeUser},
	}

	// 模拟 GetOsnNames 返回，只有 u-found 有名字
	osnMap := umtypes.NewOsnInfoMapS()
	osnMap.UserNameMap["u-found"] = "Found User"
	mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).Return(osnMap, nil)

	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck
	resp, err := svc.genPmsControlResp(ctx, permissions)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Users, 2)
	// u-found 有名字
	var foundUsername, missingUsername string

	for _, u := range resp.Users {
		if u.UserID == "u-found" {
			foundUsername = u.Username
		} else {
			missingUsername = u.Username
		}
	}

	assert.Equal(t, "Found User", foundUsername)
	// u-missing 不在 map，应该使用 unknownUserName（i18n）
	assert.NotEmpty(t, missingUsername)
}

// TestGenPmsControlResp_GetOsnNames_Error GetOsnNames 返回错误
func TestGenPmsControlResp_GetOsnNames_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &releaseSvc{
		SvcBase: service.NewSvcBase(),
		umHttp:  mockUm,
	}

	permissions := []*dapo.ReleasePermissionPO{
		{ObjectId: "u1", ObjectType: cenum.PmsTargetObjTypeUser},
	}

	mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).Return(nil, errors.New("osn service error"))

	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck
	_, err := svc.genPmsControlResp(ctx, permissions)
	assert.Error(t, err)
}
