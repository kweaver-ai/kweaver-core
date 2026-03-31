package v3agentconfigsvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/skillvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentConfigSvc_DetailPmsCheck_PrivateAPI(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	po := &dapo.DataAgentPo{
		ID:        "agent-1",
		Name:      "Test Agent",
		CreatedBy: "user-123",
	}

	err := svc.detailPmsCheck(ctx, po, true, "user-456")

	assert.NoError(t, err)
}

func TestDataAgentConfigSvc_DetailPmsCheck_PublicAPI_NoPermission(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	po := &dapo.DataAgentPo{
		ID:        "agent-1",
		Name:      "Test Agent",
		CreatedBy: "user-123",
	}

	// isOwnerOrHasBuiltInAgentMgmtPermission will return error when user doesn't have permission
	err := svc.detailPmsCheck(ctx, po, false, "user-456")

	// This will fail with error since the user is not the owner and doesn't have built-in agent mgmt permission
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_IsOwnerOrBuiltIn_BuiltInHasPermission(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	builtInYes := cdaenum.BuiltInYes
	po := &dapo.DataAgentPo{
		CreatedBy: "other",
		IsBuiltIn: &builtInYes,
	}

	// uid != createdBy, isBuiltIn=true, hasPermission=true → no error
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

	err := svc.isOwnerOrHasBuiltInAgentMgmtPermission(context.Background(), po, "user-1")
	assert.NoError(t, err)
}

func TestDataAgentConfigSvc_IsOwnerOrBuiltIn_BuiltInNoPermission(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	builtInYes := cdaenum.BuiltInYes
	po := &dapo.DataAgentPo{
		CreatedBy: "other",
		IsBuiltIn: &builtInYes,
	}

	// uid != createdBy, isBuiltIn=true, hasPermission=false → 403
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	err := svc.isOwnerOrHasBuiltInAgentMgmtPermission(context.Background(), po, "user-1")
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_IsOwnerOrBuiltIn_BuiltInPermissionError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	builtInYes := cdaenum.BuiltInYes
	po := &dapo.DataAgentPo{
		CreatedBy: "other",
		IsBuiltIn: &builtInYes,
	}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("pms err"))

	err := svc.isOwnerOrHasBuiltInAgentMgmtPermission(context.Background(), po, "user-1")
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_MarkSkillAgentPmsForDetail_NilSkill(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	config := &daconfvalobj.Config{
		Skill: &skillvalobj.Skill{},
	}
	eo := &daconfeo.DataAgent{
		Config: config,
	}

	err := svc.markSkillAgentPmsForDetail(ctx, eo, "user-123")

	assert.NoError(t, err)
}

func TestDataAgentConfigSvc_MarkSkillAgentPmsForDetail_EmptySkillAgents(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	config := &daconfvalobj.Config{
		Skill: &skillvalobj.Skill{
			Agents: []*skillvalobj.SkillAgent{},
		},
	}
	eo := &daconfeo.DataAgent{
		Config: config,
	}

	err := svc.markSkillAgentPmsForDetail(ctx, eo, "user-123")

	assert.NoError(t, err)
}

func TestDataAgentConfigSvc_DetailPmsCheck_SameUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	po := &dapo.DataAgentPo{
		ID:        "agent-1",
		Name:      "Test Agent",
		CreatedBy: "user-123",
	}

	err := svc.detailPmsCheck(ctx, po, false, "user-123")

	assert.NoError(t, err)
}

func TestDataAgentConfigSvc_MarkSkillAgentPmsForDetail_WithSkillAgents_PubedRepoError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPubedAgentRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		pubedAgentRepo: mockPubedAgentRepo,
	}

	ctx := context.Background()
	config := &daconfvalobj.Config{
		Skill: &skillvalobj.Skill{
			Agents: []*skillvalobj.SkillAgent{
				{AgentKey: "agent-key-1"},
			},
		},
	}
	eo := &daconfeo.DataAgent{Config: config}

	mockPubedAgentRepo.EXPECT().GetPubedPoMapByXx(gomock.Any(), gomock.Any()).Return(nil, errors.New("repo err"))

	err := svc.markSkillAgentPmsForDetail(ctx, eo, "user-123")
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_MarkSkillAgentPmsForDetail_CheckUseAgentPmsError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPubedAgentRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		pubedAgentRepo: mockPubedAgentRepo,
		authZHttp:      mockAuthZHttp,
	}

	ctx := context.Background()
	keyA := "agent-key-a"
	skillAgentA := &skillvalobj.SkillAgent{AgentKey: keyA}
	config := &daconfvalobj.Config{
		Skill: &skillvalobj.Skill{
			Agents: []*skillvalobj.SkillAgent{skillAgentA},
		},
	}
	eo := &daconfeo.DataAgent{Config: config}

	pubedRet := padbret.NewGetPaPoMapByXxRet()
	pubedRet.JoinPosKey2PoMap[keyA] = &dapo.PublishedJoinPo{DataAgentPo: dapo.DataAgentPo{Key: keyA, ID: "id-a"}}

	mockPubedAgentRepo.EXPECT().GetPubedPoMapByXx(gomock.Any(), gomock.Any()).Return(pubedRet, nil)
	mockAuthZHttp.EXPECT().FilterCanUseAgentIDMap(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("authz err"))

	err := svc.markSkillAgentPmsForDetail(ctx, eo, "user-123")
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_MarkSkillAgentPmsForDetail_PmsCtrlHasPms(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPubedAgentRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		pubedAgentRepo: mockPubedAgentRepo,
		authZHttp:      mockAuthZHttp,
	}

	ctx := context.Background()
	keyPms := "agent-pms-ctrl"
	keyNoPms := "agent-no-pms-ctrl"
	skillAgentPms := &skillvalobj.SkillAgent{AgentKey: keyPms}
	skillAgentNoPms := &skillvalobj.SkillAgent{AgentKey: keyNoPms}
	config := &daconfvalobj.Config{
		Skill: &skillvalobj.Skill{
			Agents: []*skillvalobj.SkillAgent{skillAgentPms, skillAgentNoPms},
		},
	}
	eo := &daconfeo.DataAgent{Config: config}

	pubedRet := padbret.NewGetPaPoMapByXxRet()
	// IsPmsCtrl=1 for keyPms → filtered via FilterCanUseAgentIDMap
	pubedRet.JoinPosKey2PoMap[keyPms] = &dapo.PublishedJoinPo{
		DataAgentPo:   dapo.DataAgentPo{Key: keyPms, ID: "id-pms"},
		ReleasePartPo: dapo.ReleasePartPo{IsPmsCtrl: 1},
	}
	// IsPmsCtrl=0 for keyNoPms → directly in hasPmsMap
	pubedRet.JoinPosKey2PoMap[keyNoPms] = &dapo.PublishedJoinPo{
		DataAgentPo: dapo.DataAgentPo{Key: keyNoPms, ID: "id-no-pms"},
	}

	mockPubedAgentRepo.EXPECT().GetPubedPoMapByXx(gomock.Any(), gomock.Any()).Return(pubedRet, nil)
	// keyPms has permission (id-pms in filteredMap), keyNoPms has permission (not PmsCtrl)
	mockAuthZHttp.EXPECT().FilterCanUseAgentIDMap(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(map[string]struct{}{"id-pms": {}}, nil)

	err := svc.markSkillAgentPmsForDetail(ctx, eo, "user-123")
	assert.NoError(t, err)
	assert.Equal(t, skillvalobj.CurrentPmsCheckStatusSuccess, skillAgentPms.CurrentPmsCheckStatus)
	assert.Equal(t, skillvalobj.CurrentPmsCheckStatusSuccess, skillAgentNoPms.CurrentPmsCheckStatus)
}

func TestDataAgentConfigSvc_MarkSkillAgentPmsForDetail_PmsCtrlNoPms(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPubedAgentRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		pubedAgentRepo: mockPubedAgentRepo,
		authZHttp:      mockAuthZHttp,
	}

	ctx := context.Background()
	keyPms := "agent-pms-ctrl-nopms"
	skillAgent := &skillvalobj.SkillAgent{AgentKey: keyPms}
	config := &daconfvalobj.Config{
		Skill: &skillvalobj.Skill{
			Agents: []*skillvalobj.SkillAgent{skillAgent},
		},
	}
	eo := &daconfeo.DataAgent{Config: config}

	pubedRet := padbret.NewGetPaPoMapByXxRet()
	pubedRet.JoinPosKey2PoMap[keyPms] = &dapo.PublishedJoinPo{
		DataAgentPo:   dapo.DataAgentPo{Key: keyPms, ID: "id-pms"},
		ReleasePartPo: dapo.ReleasePartPo{IsPmsCtrl: 1},
	}

	mockPubedAgentRepo.EXPECT().GetPubedPoMapByXx(gomock.Any(), gomock.Any()).Return(pubedRet, nil)
	// keyPms NOT in filteredMap → CurrentPmsCheckStatusFailed (covers line 70-72)
	mockAuthZHttp.EXPECT().FilterCanUseAgentIDMap(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(map[string]struct{}{}, nil)

	err := svc.markSkillAgentPmsForDetail(ctx, eo, "user-123")
	assert.NoError(t, err)
	assert.True(t, skillAgent.CurrentIsExistsAndPublished)
	assert.Equal(t, skillvalobj.CurrentPmsCheckStatusFailed, skillAgent.CurrentPmsCheckStatus)
}

func TestDataAgentConfigSvc_MarkSkillAgentPmsForDetail_HasPmsAndNoPms(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPubedAgentRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		pubedAgentRepo: mockPubedAgentRepo,
		authZHttp:      mockAuthZHttp,
	}

	ctx := context.Background()
	keyHasPms := "agent-key-has-pms"
	keyNoPms := "agent-key-no-pms"

	skillAgentHas := &skillvalobj.SkillAgent{AgentKey: keyHasPms}
	skillAgentNo := &skillvalobj.SkillAgent{AgentKey: keyNoPms}
	config := &daconfvalobj.Config{
		Skill: &skillvalobj.Skill{
			Agents: []*skillvalobj.SkillAgent{skillAgentHas, skillAgentNo},
		},
	}
	eo := &daconfeo.DataAgent{Config: config}

	// Both agents are in pubedMap; IsPmsCtrl=false → keys go directly into hasPmsMap
	pubedRet := padbret.NewGetPaPoMapByXxRet()
	pubedRet.JoinPosKey2PoMap[keyHasPms] = &dapo.PublishedJoinPo{DataAgentPo: dapo.DataAgentPo{Key: keyHasPms, ID: "id-has"}}
	pubedRet.JoinPosKey2PoMap[keyNoPms] = &dapo.PublishedJoinPo{DataAgentPo: dapo.DataAgentPo{Key: keyNoPms, ID: "id-no"}}

	mockPubedAgentRepo.EXPECT().GetPubedPoMapByXx(gomock.Any(), gomock.Any()).Return(pubedRet, nil)
	// IsPmsCtrl=false (zero value) → no need for FilterCanUseAgentIDMap with agent ids
	// checkUseAgentPms calls FilterCanUseAgentIDMap with empty agentIds slice
	mockAuthZHttp.EXPECT().FilterCanUseAgentIDMap(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(map[string]struct{}{}, nil)

	err := svc.markSkillAgentPmsForDetail(ctx, eo, "user-123")
	assert.NoError(t, err)
	// Both are in pubedMap → CurrentIsExistsAndPublished = true
	assert.True(t, skillAgentHas.CurrentIsExistsAndPublished)
	assert.True(t, skillAgentNo.CurrentIsExistsAndPublished)
	// Both keys in hasPmsMap (IsPmsCtrl=false → direct insert) → status success
	assert.Equal(t, skillvalobj.CurrentPmsCheckStatusSuccess, skillAgentHas.CurrentPmsCheckStatus)
	assert.Equal(t, skillvalobj.CurrentPmsCheckStatusSuccess, skillAgentNo.CurrentPmsCheckStatus)
}

func TestDataAgentConfigSvc_MarkSkillAgentPmsForDetail_AgentNotInPubedMap(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPubedAgentRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		pubedAgentRepo: mockPubedAgentRepo,
		authZHttp:      mockAuthZHttp,
	}

	ctx := context.Background()
	agentKey := "agent-key-not-published"
	skillAgent := &skillvalobj.SkillAgent{AgentKey: agentKey}
	config := &daconfvalobj.Config{
		Skill: &skillvalobj.Skill{
			Agents: []*skillvalobj.SkillAgent{skillAgent},
		},
	}
	eo := &daconfeo.DataAgent{Config: config}

	// pubedMap is empty → checkUseAgentPms gets empty map → FilterCanUseAgentIDMap called with empty slice
	pubedRet := padbret.NewGetPaPoMapByXxRet()
	mockPubedAgentRepo.EXPECT().GetPubedPoMapByXx(gomock.Any(), gomock.Any()).Return(pubedRet, nil)
	mockAuthZHttp.EXPECT().FilterCanUseAgentIDMap(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]struct{}{}, nil).AnyTimes()

	err := svc.markSkillAgentPmsForDetail(ctx, eo, "user-123")
	assert.NoError(t, err)
	// agent not in pubedMap → CurrentIsExistsAndPublished = false
	assert.False(t, skillAgent.CurrentIsExistsAndPublished)
}
