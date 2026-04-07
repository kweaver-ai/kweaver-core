package auth

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

// GetAccessor 获取访问者信息
func (s *authServiceImpl) GetAccessor(ctx context.Context, userID string) (*interfaces.AuthAccessor, error) {
	// 从上下文中读取账户认证上下文
	authContext, ok := common.GetAccountAuthContextFromCtx(ctx)
	if !ok {
		authContext = &interfaces.AccountAuthContext{
			AccountID:   userID,
			AccountType: interfaces.AccessorTypeUser, // 默认用户类型
		}
	}

	// 内部接口仅允许实名用户访问
	if authContext.AccountID == "" {
		return nil, infraerrors.NewHTTPError(ctx, http.StatusNotFound, infraerrors.ErrExtCommonUserNotFound, "userID is empty")
	}
	accessor := &interfaces.AuthAccessor{
		ID: authContext.AccountID,
	}
	switch authContext.AccountType {
	case interfaces.AccessorTypeUser, interfaces.AccessorTypeAnonymous:
		// 实名用户, 匿名用户
		userInfos, err := s.userManagement.GetUsersInfo(ctx, []string{authContext.AccountID}, []string{interfaces.DisplayName})
		if err != nil {
			return nil, err
		}
		if len(userInfos) == 0 {
			return nil, infraerrors.NewHTTPError(ctx, http.StatusNotFound, infraerrors.ErrExtCommonUserNotFound, nil)
		}
		accessor.Type = interfaces.AccessorTypeUser
		accessor.Name = userInfos[0].DisplayName
	case interfaces.AccessorTypeApp:
		// 应用账户
		appInfo, err := s.userManagement.GetAppInfo(ctx, authContext.AccountID)
		if err != nil {
			return nil, err
		}
		accessor.Type = interfaces.AccessorTypeApp
		accessor.Name = appInfo.Name
	case interfaces.AccessorTypeDepartment, interfaces.AccessorTypeGroup, interfaces.AccessorTypeRole:
		return nil, infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonDepartmentOrGroupOrRoleNotAllowed,
			"department, group or role account not allowed")
	default:
		return nil, infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonInvalidAccessorType, "invalid accessor type")
	}
	return accessor, nil
}
