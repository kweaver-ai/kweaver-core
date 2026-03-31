package apierr

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/kweaver-ai/decision-agent/agent-factory/locale"
)

// 公共错误码, 服务内所有模块均可使用
const (
	// 400
	AgentFactory_InvalidParameter_RequestBody = "AgentFactory.InvalidParameter.RequestBody"
	// 401
	AgentFactory_InvalidRequestHeader_Authorization = "AgentFactory.InvalidRequestHeader.Authorization"

	// 403
	AgentFactory_Forbidden_FilterField = "AgentFactory.Forbidden.FilterField"

	// 406
	AgentFactory_InvalidRequestHeader_ContentType = "AgentFactory.InvalidRequestHeader.ContentType"

	// 500
	AgentFactory_InternalError = "AgentFactory.InternalError"
)

// release agent error code
const (
	AgentFactory_Release_InternalError_PublishFailed = "AgentFactory.Release.InternalError.PublishFailed"

	//	not exists
	ReleaseNotFound = "AgentFactory.Release.NotFound"

	// 因为配置错误，无法发布
	PublishFailedByConfigError = "AgentFactory.Release.BadRequest.PublishFailedByConfigError"
)

// release agent history error code
const (
	ReleaseHistoryNotFound = "AgentFactory.ReleaseHistory.NotFound"
)

// data agent config error code
const (
	DataAgentConfigNameExists = "AgentFactory.DataAgentConfig.Conflict.NameExists" // data agent 名称已存在
	DataAgentConfigNotFound   = "AgentFactory.DataAgentConfig.NotFound"            // data agent 不存在

	DataAgentConfigForbiddenNotOwner = "AgentFactory.DataAgentConfig.Forbidden.NotOwner" // 访问者不是创建者，无法操作

	DataAgentConfigPublishedCannotBeDeleted = "AgentFactory.DataAgentConfig.Conflict.Published.CannotBeDeleted" // data agent 已发布，无法删除

)

// custom space error code
const (
	CustomSpaceNameExists = "AgentFactory.CustomSpace.Conflict.NameExists" // 自定义空间名称已存在

	CustomSpaceKeyExists = "AgentFactory.CustomSpace.Conflict.KeyExists" // 自定义空间Key已存在

	CustomSpaceNotFound = "AgentFactory.CustomSpace.NotFound" // 自定义空间不存在
)

// custom space member error code
const (
	CustomSpaceMemberNotFound = "AgentFactory.CustomSpaceMember.NotFound" // 自定义空间成员不存在

	CustomSpaceMemberNotBelongsToSpace = "AgentFactory.CustomSpaceMember.NotBelongsToSpace" // 自定义空间成员不属于该空间
)

// custom space resource error code
const (
	CustomSpaceResourceNotFound = "AgentFactory.CustomSpaceResource.NotFound" // 自定义空间资源不存在

	CustomSpaceResourceNotBelongsToSpace = "AgentFactory.CustomSpaceResource.NotBelongsToSpace" // 自定义空间资源不属于该空间
)

// product error code
const (
	ProductNameExists = "AgentFactory.Product.Conflict.NameExists" // 产品名称已存在

	ProductKeyExists = "AgentFactory.Product.Conflict.KeyExists" // 产品Key已存在
	ProductNotFound  = "AgentFactory.Product.NotFound"           // 产品不存在

)

// category error code
const (
	CategoryNotFound = "AgentFactory.Category.NotFound" // 分类不存在
)

// agent tpl error code
const (
	AgentTplNameExists = "AgentFactory.AgentTpl.Conflict.NameExists" // agent tpl 名称已存在
	AgentTplKeyExists  = "AgentFactory.AgentTpl.Conflict.KeyExists"  // agent tpl Key已存在
	AgentTplNotFound   = "AgentFactory.AgentTpl.NotFound"            // agent tpl 不存在

	AgentTplForbiddenNotOwner = "AgentFactory.AgentTpl.Forbidden.NotOwner" // 访问者不是创建者，无法操作

	AgentTplIsUnpublished = "AgentFactory.AgentTpl.IsUnpublished" // agent tpl 是未发布状态

	AgentTplPublishedCannotBeDeleted = "AgentFactory.AgentTpl.Conflict.Published.CannotBeDeleted" // agent tpl 已发布，无法删除
)

const (
	AiAutogenError = "AgentFactory.AiAutogenError" // AI自动生成错误
)

// published agent tpl error code
const (
	PublishedTplNotFound = "AgentFactory.PublishedTpl.NotFound" // published agent tpl 不存在
)

// 权限通用错误码
const (
	AgentFactoryPermissionForbidden = "AgentFactory.Permission.Forbidden"
)

// 导入、导出
const (
	AgentFactoryInoutParseFileFailed = "AgentFactory.Inout.ParseFileFailed"
	// 超出单次导入最大数量
	AgentFactoryInoutMaxSizeExceeded = "AgentFactory.Inout.MaxSizeExceeded"
)

var errCodeList = []string{
	// ---公共错误码---
	// 400
	AgentFactory_InvalidParameter_RequestBody,
	AgentFactory_InvalidRequestHeader_Authorization,
	AgentFactory_Forbidden_FilterField,
	AgentFactory_InvalidRequestHeader_ContentType,
	AgentFactory_InternalError,

	// release error code
	AgentFactory_Release_InternalError_PublishFailed,
	ReleaseNotFound,
	PublishFailedByConfigError,

	// release history error code
	ReleaseHistoryNotFound,

	// data agent config error code
	DataAgentConfigNameExists,
	DataAgentConfigNotFound,
	DataAgentConfigForbiddenNotOwner,
	DataAgentConfigPublishedCannotBeDeleted,
	capierr.DataAgentConfigLlmRequired,
	capierr.DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize,

	// custom space error code
	CustomSpaceNameExists,
	CustomSpaceKeyExists,
	CustomSpaceNotFound,

	// custom space member error code
	CustomSpaceMemberNotFound,

	// custom space resource error code
	CustomSpaceResourceNotFound,

	// product error code
	ProductNameExists,
	ProductKeyExists,
	ProductNotFound,

	// category error code
	CategoryNotFound,

	// agent tpl error code
	AgentTplNameExists,
	AgentTplKeyExists,
	AgentTplNotFound,
	AgentTplForbiddenNotOwner,
	AgentTplIsUnpublished,
	AgentTplPublishedCannotBeDeleted,

	// ai autogen error code
	AiAutogenError,

	// published agent tpl error code
	PublishedTplNotFound,

	// permission error code
	AgentFactoryPermissionForbidden,

	// import error code
	AgentFactoryInoutParseFileFailed,
	AgentFactoryInoutMaxSizeExceeded,
}

func init() {
	locale.Register()
	rest.Register(errCodeList)
	// 注册 APP 相关错误码（来自合并的 agent-app）
	rest.Register(GetAppErrorCodeList())
}
