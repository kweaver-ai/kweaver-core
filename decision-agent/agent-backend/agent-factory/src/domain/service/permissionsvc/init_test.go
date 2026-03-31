package permissionsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func init() {
	// Initialize global config for tests
	if global.GConfig == nil {
		global.GConfig = &conf.Config{
			SwitchFields: conf.NewSwitchFields(),
		}
	}
}

func TestPermissionSvc_InitPermission_GrantAgentUsePmsForAppAdminError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZHttp,
		umHttp:    mockUmHttp,
	}

	ctx := context.Background()
	httpErr := errors.New("http request failed")

	mockAuthZHttp.EXPECT().GrantAgentUsePmsForAppAdmin(ctx).Return(httpErr)

	err := svc.InitPermission(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "grant agent use pms for app admin failed")
}

func TestPermissionSvc_InitPermission_GrantMgmtPmsForAppAdminError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZHttp,
		umHttp:    mockUmHttp,
	}

	ctx := context.Background()
	httpErr := errors.New("http request failed")

	mockAuthZHttp.EXPECT().GrantAgentUsePmsForAppAdmin(ctx).Return(nil)
	mockAuthZHttp.EXPECT().GrantMgmtPmsForAppAdmin(ctx).Return(httpErr)

	err := svc.InitPermission(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "grant agent use pms for app admin failed")
}

func TestPermissionSvc_InitPermission_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:   service.NewSvcBase(),
		authZHttp: mockAuthZHttp,
		umHttp:    mockUmHttp,
	}

	ctx := context.Background()

	mockAuthZHttp.EXPECT().GrantAgentUsePmsForAppAdmin(ctx).Return(nil)
	mockAuthZHttp.EXPECT().GrantMgmtPmsForAppAdmin(ctx).Return(nil)

	err := svc.InitPermission(ctx)

	assert.NoError(t, err)
}
