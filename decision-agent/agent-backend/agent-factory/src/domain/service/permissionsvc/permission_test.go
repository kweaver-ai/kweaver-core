package permissionsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewPermissionService_WithAllDependencies(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dto := &NewPermissionSvcDto{
		SvcBase:               service.NewSvcBase(),
		AgentConfigRepo:       idbaccessmock.NewMockIDataAgentConfigRepo(ctrl),
		ReleaseRepo:           idbaccessmock.NewMockIReleaseRepo(ctrl),
		ReleasePermissionRepo: idbaccessmock.NewMockIReleasePermissionRepo(ctrl),
		UmHttp:                httpaccmock.NewMockUmHttpAcc(ctrl),
		AuthZHttp:             authzaccmock.NewMockAuthZHttpAcc(ctrl),
	}

	svc := NewPermissionService(dto)

	assert.NotNil(t, svc)
	assert.IsType(t, &permissionSvc{}, svc)
}

func TestNewPermissionService_WithMinimalDependencies(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dto := &NewPermissionSvcDto{
		SvcBase:               service.NewSvcBase(),
		AgentConfigRepo:       nil,
		ReleaseRepo:           nil,
		ReleasePermissionRepo: nil,
		UmHttp:                nil,
		AuthZHttp:             nil,
	}

	svc := NewPermissionService(dto)

	assert.NotNil(t, svc)
}

func TestBuildAgentOperationItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		op            cdapmsenum.Operator
		expectedNil   bool
		expectedID    string
		checkNames    bool
		checkScope    bool
		expectedScope []string
	}{
		{
			name:          "AgentPublish operation",
			op:            cdapmsenum.AgentPublish,
			expectedNil:   false,
			expectedID:    string(cdapmsenum.AgentPublish),
			checkNames:    true,
			checkScope:    true,
			expectedScope: []string{"type"},
		},
		{
			name:          "AgentUnpublish operation",
			op:            cdapmsenum.AgentUnpublish,
			expectedNil:   false,
			expectedID:    string(cdapmsenum.AgentUnpublish),
			checkNames:    true,
			checkScope:    true,
			expectedScope: []string{"type"},
		},
		{
			name:          "AgentUse operation",
			op:            cdapmsenum.AgentUse,
			expectedNil:   false,
			expectedID:    string(cdapmsenum.AgentUse),
			checkNames:    true,
			checkScope:    true,
			expectedScope: []string{"type", "instance"},
		},
		{
			name:          "AgentCreateSystemAgent operation",
			op:            cdapmsenum.AgentCreateSystemAgent,
			expectedNil:   false,
			expectedID:    string(cdapmsenum.AgentCreateSystemAgent),
			checkNames:    true,
			checkScope:    true,
			expectedScope: []string{"type"},
		},
		{
			name:        "Unknown operation returns nil",
			op:          cdapmsenum.Operator("unknown_operation"),
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := buildAgentOperationItem(tt.op)

			if tt.expectedNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedID, result.ID)

				if tt.checkNames {
					assert.NotNil(t, result.Name)
					assert.Len(t, result.Name, 3) // Should have 3 language variants

					// Check that all required languages are present
					languages := make(map[string]bool)
					for _, name := range result.Name {
						languages[name.Language] = true
					}

					assert.True(t, languages["zh-cn"])
					assert.True(t, languages["en-us"])
					assert.True(t, languages["zh-tw"])
				}

				if tt.checkScope {
					assert.Equal(t, tt.expectedScope, result.Scope)
				}
			}
		})
	}
}

func TestBuildAgentOperationItem_AllAgentOperations(t *testing.T) {
	t.Parallel()

	// Test all agent operations have proper mappings
	allOps := cdapmsenum.GetAllAgentOperator()

	for _, op := range allOps {
		t.Run(string(op), func(t *testing.T) {
			t.Parallel()

			result := buildAgentOperationItem(op)
			assert.NotNil(t, result, "Operation %s should have a valid mapping", op)
			assert.Equal(t, string(op), result.ID)
			assert.NotNil(t, result.Name)
			assert.NotEmpty(t, result.Scope)
		})
	}
}

func TestBuildAgentOperationItem_AllPublishVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		op       cdapmsenum.Operator
		scope    []string
		hasNames bool
	}{
		{
			name:     "AgentPublishToBeSkillAgent",
			op:       cdapmsenum.AgentPublishToBeSkillAgent,
			scope:    []string{"type"},
			hasNames: true,
		},
		{
			name:     "AgentPublishToBeWebSdkAgent",
			op:       cdapmsenum.AgentPublishToBeWebSdkAgent,
			scope:    []string{"type"},
			hasNames: true,
		},
		{
			name:     "AgentPublishToBeApiAgent",
			op:       cdapmsenum.AgentPublishToBeApiAgent,
			scope:    []string{"type"},
			hasNames: true,
		},
		{
			name:     "AgentPublishToBeDataFlowAgent",
			op:       cdapmsenum.AgentPublishToBeDataFlowAgent,
			scope:    []string{"type"},
			hasNames: true,
		},
		{
			name:     "AgentUnpublishOtherUserAgent",
			op:       cdapmsenum.AgentUnpublishOtherUserAgent,
			scope:    []string{"type"},
			hasNames: true,
		},
		{
			name:     "AgentBuiltInAgentMgmt",
			op:       cdapmsenum.AgentBuiltInAgentMgmt,
			scope:    []string{"type"},
			hasNames: true,
		},
		{
			name:     "AgentSeeTrajectoryAnalysis",
			op:       cdapmsenum.AgentSeeTrajectoryAnalysis,
			scope:    []string{"type"},
			hasNames: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := buildAgentOperationItem(tt.op)
			assert.NotNil(t, result)
			assert.Equal(t, string(tt.op), result.ID)
			assert.Equal(t, tt.scope, result.Scope)

			if tt.hasNames {
				assert.NotNil(t, result.Name)
				assert.Len(t, result.Name, 3)
			}
		})
	}
}

func TestCheckByPmsPlatform_EmptyAccessorIDs(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agentPo := &dapo.DataAgentPo{
		ID:        "agent-123",
		CreatedBy: "owner-user",
	}

	svc := &permissionSvc{
		SvcBase: service.NewSvcBase(),
	}

	hasPms, err := svc.checkByPmsPlatform(context.Background(), agentPo, "", "")

	assert.False(t, hasPms)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user id or app account id cannot be all empty")
}

func TestCheckByPmsPlatform_WithUserID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZ := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	agentPo := &dapo.DataAgentPo{
		ID:        "agent-123",
		CreatedBy: "owner-user",
	}

	mockAuthZ.EXPECT().SingleAgentUseCheck(gomock.Any(), "user-123", gomock.Any(), "agent-123").Return(true, nil)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZ,
	}

	hasPms, err := svc.checkByPmsPlatform(context.Background(), agentPo, "user-123", "")

	assert.True(t, hasPms)
	assert.NoError(t, err)
}

func TestCheckByPmsPlatform_WithAppAccountID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZ := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	agentPo := &dapo.DataAgentPo{
		ID:        "agent-123",
		CreatedBy: "owner-user",
	}

	mockAuthZ.EXPECT().SingleAgentUseCheck(gomock.Any(), "app-123", gomock.Any(), "agent-123").Return(false, nil)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZ,
	}

	hasPms, err := svc.checkByPmsPlatform(context.Background(), agentPo, "", "app-123")

	assert.False(t, hasPms)
	assert.NoError(t, err)
}

func TestCheckByPmsPlatform_AuthZError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZ := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	agentPo := &dapo.DataAgentPo{
		ID:        "agent-123",
		CreatedBy: "owner-user",
	}

	expectedErr := errors.New("authZ service error")
	mockAuthZ.EXPECT().SingleAgentUseCheck(gomock.Any(), "user-123", gomock.Any(), "agent-123").Return(false, expectedErr)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZ,
	}

	hasPms, err := svc.checkByPmsPlatform(context.Background(), agentPo, "user-123", "")

	assert.False(t, hasPms)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "[checkByPmsPlatform][SingleAgentUseCheck]")
}

func TestCheckUserPms_OwnerHasPermission(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agentPo := &dapo.DataAgentPo{
		ID:        "agent-123",
		CreatedBy: "user-123",
	}

	svc := &permissionSvc{
		SvcBase: service.NewSvcBase(),
	}

	// Owner has permission immediately, no DB call needed
	hasPms, err := svc.checkUserPms(context.Background(), agentPo, "user-123", "")

	assert.True(t, hasPms)
	assert.NoError(t, err)
}

func TestCheckUserPms_EmptyAccessorIDs(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)

	agentPo := &dapo.DataAgentPo{
		ID:        "agent-123",
		CreatedBy: "owner-user",
	}

	// Since user is not the owner, GetByAgentID will be called
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "agent-123").Return(nil, nil)

	svc := &permissionSvc{
		SvcBase:     service.NewSvcBase(),
		releaseRepo: mockReleaseRepo,
	}

	// Empty IDs will be checked in checkByPmsPlatform
	hasPms, err := svc.checkUserPms(context.Background(), agentPo, "", "")

	assert.False(t, hasPms)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user id or app account id cannot be all empty")
}

func TestCheckUserPms_NonOwnerWithPublishedAgentNoPermissionControl(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)

	agentPo := &dapo.DataAgentPo{
		ID:        "agent-123",
		CreatedBy: "owner-user",
	}

	isPmsCtrl := 0
	releasePo := &dapo.ReleasePO{
		ID:        "release-123",
		AgentID:   "agent-123",
		IsPmsCtrl: &isPmsCtrl,
	}

	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "agent-123").Return(releasePo, nil)

	svc := &permissionSvc{
		SvcBase:     service.NewSvcBase(),
		releaseRepo: mockReleaseRepo,
	}

	hasPms, err := svc.checkUserPms(context.Background(), agentPo, "other-user", "")

	assert.True(t, hasPms)
	assert.NoError(t, err)
}

func TestCheckUserPms_ReleaseRepositoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)

	agentPo := &dapo.DataAgentPo{
		ID:        "agent-123",
		CreatedBy: "owner-user",
	}

	expectedErr := errors.New("release database error")
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "agent-123").Return(nil, expectedErr)

	svc := &permissionSvc{
		SvcBase:     service.NewSvcBase(),
		releaseRepo: mockReleaseRepo,
	}

	hasPms, err := svc.checkUserPms(context.Background(), agentPo, "other-user", "")

	assert.False(t, hasPms)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "[checkUserPms][GetByAgentID]")
}

func TestCheckUserPms_NoReleaseWithPermissionCheck(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockAuthZ := authzaccmock.NewMockAuthZHttpAcc(ctrl)

	agentPo := &dapo.DataAgentPo{
		ID:        "agent-123",
		CreatedBy: "owner-user",
	}

	// No release found (nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "agent-123").Return(nil, nil)
	// But user has permission through authZ
	mockAuthZ.EXPECT().SingleAgentUseCheck(gomock.Any(), "other-user", gomock.Any(), "agent-123").Return(true, nil)

	svc := &permissionSvc{
		SvcBase:     service.NewSvcBase(),
		releaseRepo: mockReleaseRepo,
		authZHttp:   mockAuthZ,
	}

	hasPms, err := svc.checkUserPms(context.Background(), agentPo, "other-user", "")

	assert.True(t, hasPms)
	assert.NoError(t, err)
}
