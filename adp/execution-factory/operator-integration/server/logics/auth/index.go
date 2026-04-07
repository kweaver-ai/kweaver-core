package auth

import (
	"context"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/mq"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

type authServiceImpl struct {
	logger         interfaces.Logger
	authorization  interfaces.Authorization
	mqClient       mq.MQClient
	userManagement interfaces.UserManagement
}

type noopAuthService struct{}

func NewAuthServiceImpl() interfaces.IAuthorizationService {
	if !config.GetAuthEnabled() {
		return &noopAuthService{}
	}
	return &authServiceImpl{
		logger:         config.NewConfigLoader().GetLogger(),
		authorization:  drivenadapters.NewAuthorization(),
		mqClient:       mq.NewMQClient(),
		userManagement: drivenadapters.NewUserManagementClient(),
	}
}

func (n *noopAuthService) CheckCreatePermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopAuthService) CheckViewPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopAuthService) CheckModifyPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopAuthService) CheckDeletePermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopAuthService) CheckPublishPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopAuthService) CheckUnpublishPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopAuthService) CheckAuthorizePermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopAuthService) CheckPublicAccessPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopAuthService) CheckExecutePermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopAuthService) MultiCheckOperationPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType, operations ...interfaces.AuthOperationType) error {
	return nil
}

func (n *noopAuthService) CreateOwnerPolicy(ctx context.Context, accessor *interfaces.AuthAccessor, authResource *interfaces.AuthResource) error {
	return nil
}

func (n *noopAuthService) CreateIntCompPolicyForAllUsers(ctx context.Context, authResource *interfaces.AuthResource) error {
	return nil
}

func (n *noopAuthService) ResourceFilterIDs(ctx context.Context, accessor *interfaces.AuthAccessor, resourceIDS []string, resourceType interfaces.AuthResourceType, operations ...interfaces.AuthOperationType) ([]string, error) {
	return resourceIDS, nil
}

func (n *noopAuthService) ResourceListIDs(ctx context.Context, accessor *interfaces.AuthAccessor, resourceType interfaces.AuthResourceType, operations ...interfaces.AuthOperationType) ([]string, error) {
	return []string{interfaces.ResourceIDAll}, nil
}

func (n *noopAuthService) OperationCheckAll(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType, operations ...interfaces.AuthOperationType) (bool, error) {
	return true, nil
}

func (n *noopAuthService) OperationCheckAny(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType, operations ...interfaces.AuthOperationType) (bool, error) {
	return true, nil
}

func (n *noopAuthService) CreatePolicy(ctx context.Context, accessor *interfaces.AuthAccessor, authResource *interfaces.AuthResource, allow []interfaces.AuthOperationType, deny []interfaces.AuthOperationType) error {
	return nil
}

func (n *noopAuthService) DeletePolicy(ctx context.Context, resourceIDs []string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopAuthService) NotifyResourceChange(ctx context.Context, authResource *interfaces.AuthResource) error {
	return nil
}

func (n *noopAuthService) GetAccessor(ctx context.Context, userID string) (*interfaces.AuthAccessor, error) {
	authContext, ok := common.GetAccountAuthContextFromCtx(ctx)
	if !ok {
		authContext = &interfaces.AccountAuthContext{
			AccountID:   userID,
			AccountType: interfaces.AccessorTypeUser,
		}
	}
	if authContext.AccountID == "" {
		authContext.AccountID = userID
	}
	accessor := &interfaces.AuthAccessor{
		ID:   authContext.AccountID,
		Type: authContext.AccountType,
		Name: authContext.AccountID,
	}
	if accessor.ID == "" {
		accessor.ID = interfaces.UnknownUser
		accessor.Type = interfaces.AccessorTypeAnonymous
	}
	return accessor, nil
}
