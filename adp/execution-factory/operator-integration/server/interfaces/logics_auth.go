package interfaces

import (
	"context"
)

// AccountAuthContext 账户认证上下文
type AccountAuthContext struct {
	// AccountID 账户唯一标识符
	AccountID string `json:"account_id"`
	// AccountType 账户类型
	AccountType AccessorType `json:"account_type"`
	// Token信息
	TokenInfo *TokenInfo `json:"token_info"`
}

// AuthOperationType 操作类型
//
//go:generate mockgen -source=logics_auth.go -destination=../mocks/auth.go -package=mocks
type AuthOperationType string

const (
	AuthOperationTypeCreate       AuthOperationType = "create"        // 新建
	AuthOperationTypeModify       AuthOperationType = "modify"        // 编辑
	AuthOperationTypeDelete       AuthOperationType = "delete"        // 删除
	AuthOperationTypeView         AuthOperationType = "view"          // 查看
	AuthOperationTypePublish      AuthOperationType = "publish"       // 发布
	AuthOperationTypeUnpublish    AuthOperationType = "unpublish"     // 下架
	AuthOperationTypeAuthorize    AuthOperationType = "authorize"     // 权限管理
	AuthOperationTypePublicAccess AuthOperationType = "public_access" // 公共访问
	AuthOperationTypeExecute      AuthOperationType = "execute"       // 执行
)

var (
	// 所有者权限
	OwnerPolicyList = []AuthOperationType{
		AuthOperationTypeCreate,
		AuthOperationTypeModify,
		AuthOperationTypeDelete,
		AuthOperationTypeView,
		AuthOperationTypePublish,
		AuthOperationTypeUnpublish,
		AuthOperationTypeAuthorize,
		AuthOperationTypePublicAccess,
		AuthOperationTypeExecute,
	}
)

// AuthResourceType 资源类型
type AuthResourceType string

// 支持的资源类型
const (
	AuthResourceTypeToolBox  AuthResourceType = "tool_box" // 工具箱
	AuthResourceTypeMCP      AuthResourceType = "mcp"      // MCP
	AuthResourceTypeOperator AuthResourceType = "operator" // 算子
	AuthResourceTypeSkill    AuthResourceType = "skill"    // Skill
)

func (a AuthResourceType) String() string {
	return string(a)
}

// ResourceID 资源ID类型别名
type ResourceID = string

// 特殊资源ID常量
const (
	ResourceIDAll = "*" // 表示所有资源
)

// QueryOption 查询选项函数类型
type QueryOption[T any, PT PtrBizIdentifiable[T]] func() ([]PT, error)

// ResourceListFunc 获取有权限资源ID列表的函数类型
type ResourceListFunc func() ([]string, error)

// IAuthorizationService Authorization Service接口
type IAuthorizationService interface {
	// CheckCreatePermission 检查新建权限
	CheckCreatePermission(ctx context.Context, accessor *AuthAccessor, resourceType AuthResourceType) error
	// CheckViewPermission 检查查看权限
	CheckViewPermission(ctx context.Context, accessor *AuthAccessor, resourceID string, resourceType AuthResourceType) error
	// CheckModifyPermission 检查编辑权限
	CheckModifyPermission(ctx context.Context, accessor *AuthAccessor, resourceID string, resourceType AuthResourceType) error
	// CheckDeletePermission 检查删除权限
	CheckDeletePermission(ctx context.Context, accessor *AuthAccessor, resourceID string, resourceType AuthResourceType) error
	// CheckPublishPermission 检查发布权限
	CheckPublishPermission(ctx context.Context, accessor *AuthAccessor, resourceID string, resourceType AuthResourceType) error
	// CheckUnpublishPermission 检查下架权限
	CheckUnpublishPermission(ctx context.Context, accessor *AuthAccessor, resourceID string, resourceType AuthResourceType) error
	// CheckAuthorizePermission 检查权限管理权限
	CheckAuthorizePermission(ctx context.Context, accessor *AuthAccessor, resourceID string, resourceType AuthResourceType) error
	// CheckPublicAccessPermission 检查公共访问权限
	CheckPublicAccessPermission(ctx context.Context, accessor *AuthAccessor, resourceID string, resourceType AuthResourceType) error
	// CheckExecutePermission 检查执行权限
	CheckExecutePermission(ctx context.Context, accessor *AuthAccessor, resourceID string, resourceType AuthResourceType) error
	// MultiCheckOperationPermission 多操作权限检查
	MultiCheckOperationPermission(ctx context.Context, accessor *AuthAccessor, resourceID string, resourceType AuthResourceType, operations ...AuthOperationType) error

	// CreateOwnerPolicy 创建owner权限
	CreateOwnerPolicy(ctx context.Context, accessor *AuthAccessor, authResource *AuthResource) error
	// CreateIntCompPolicyForAllUsers 创建内部组件权限策略，作用于所有用户
	CreateIntCompPolicyForAllUsers(ctx context.Context, authResource *AuthResource) error

	// ResourceFilterIDs 资源过滤
	ResourceFilterIDs(ctx context.Context, accessor *AuthAccessor, resourceIDS []string, resourceType AuthResourceType, operations ...AuthOperationType) ([]string, error)
	// ResourceListIDs 资源列举
	ResourceListIDs(ctx context.Context, accessor *AuthAccessor, resourceType AuthResourceType, operations ...AuthOperationType) ([]string, error)

	// OperationCheckAll AND关系：需要满足所有操作权限
	OperationCheckAll(ctx context.Context, accessor *AuthAccessor, resourceID string, resourceType AuthResourceType, operations ...AuthOperationType) (bool, error)
	// OperationCheckAny OR关系，只需满足任意一个操作权限
	OperationCheckAny(ctx context.Context, accessor *AuthAccessor, resourceID string, resourceType AuthResourceType, operations ...AuthOperationType) (bool, error)
	// CreatePolicy 创建策略
	CreatePolicy(ctx context.Context, accessor *AuthAccessor, authResource *AuthResource, allow []AuthOperationType, deny []AuthOperationType) error
	// DeletePolicy 删除策略
	DeletePolicy(ctx context.Context, resourceIDs []string, resourceType AuthResourceType) error
	// NotifyResourceChange 资源名称变更消息通知
	NotifyResourceChange(ctx context.Context, authResource *AuthResource) error

	// GetAccessor 获取访问者信息
	GetAccessor(ctx context.Context, userID string) (*AuthAccessor, error)
}
