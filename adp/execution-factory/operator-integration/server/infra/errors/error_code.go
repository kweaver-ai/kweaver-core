// Package errors 定义错误码
// @file errors_code.go
// @description: 定义错误码
package errors

// ErrorCode 错误码
type ErrorCode string

func (e ErrorCode) String() string {
	return string(e)
}

// 算子拓展错误码定义
const (
	ErrExtOperatorExists           ErrorCode = "OperatorExists"           // 算子已存在
	ErrExtOperatorRegisterFailed   ErrorCode = "OperatorRegisterFailed"   // 算子注册失败
	ErrExtOperatorDirectPublishErr ErrorCode = "OperatorDirectPublishErr" // 算子直接发布失败
	ErrExtCategoryTypeInvalid      ErrorCode = "CategoryTypeInvalid"      // 无效的算子类型
	ErrExtOperatorUnparsed         ErrorCode = "OperatorUnparsed"         // 未解析到有效的算子
	ErrExtOperatorNotFound         ErrorCode = "OperatorNotFound"         // 算子不存在
	ErrExtOperatorMetadataNotFound ErrorCode = "OperatorMetadataNotFound" // 算子元数据不存在
	ErrExtOperatorUnSupportUpgrade ErrorCode = "OperatorUnSupportUpgrade" // 当前算子不支持升级
	ErrExtOperatorDeleteForbidden  ErrorCode = "OperatorDeleteForbidden"  // 当前算子不允许删除
	ErrExtOperatorUnSupportEdit    ErrorCode = "OperatorUnSupportEdit"    // 当前算子不支持编辑
	ErrExtOperatorEditFailed       ErrorCode = "OperatorEditFailed"       // 算子编辑失败
	ErrExtOperatorImportLimit      ErrorCode = "OperatorImportLimit"      // 单次导入算子数量限制
	ErrExtOperatorNameEmpty        ErrorCode = "OperatorNameEmpty"        // 算子名称不能为空
	ErrExtOperatorNameTooLong      ErrorCode = "OperatorNameTooLong"      // 算子名称长度不能超过%d个字符
	ErrExtOperatorDescEmpty        ErrorCode = "OperatorDescEmpty"        // 算子描述不能为空
	ErrExtOperatorDescTooLong      ErrorCode = "OperatorDescTooLong"      // 算子描述长度不能超过%d个字符
	ErrExtOperatorImportDataLimit  ErrorCode = "OperatorImportDataLimit"  // 导入算子数据超出限制
	ErrExtOperatorExistsSameName   ErrorCode = "OperatorExistsSameName"   // 算子“%s”已存在
	ErrExtOperatorEditLimit        ErrorCode = "OperatorEditLimit"        // 仅允许单个算子编辑
	ErrExtOperatorNotAvailable     ErrorCode = "OperatorNotAvailable"     // 算子不可用
	ErrExtOnlySyncModeDebug        ErrorCode = "OnlySyncModeDebug"        // 仅支持同步模式调试
	ErrExtOperatorStatusInvalid    ErrorCode = "OperatorStatusInvalid"    // 算子状态无效
	ErrExtOperatorAsyncDataSource  ErrorCode = "OperatorAsyncDataSource"  // 异步算子不支持添加为数据源算子
	ErrExtOperatorNotExistInFile   ErrorCode = "OperatorNotExistInFile"   // 您上传的文件未包含已存在的算子
)

// 工具箱拓展错误码定义
const (
	ErrExtToolBoxNotFound                 ErrorCode = "ToolBoxNotFound"                 // 工具箱不存在
	ErrExtToolBoxNameExists               ErrorCode = "ToolBoxNameExists"               // 工具箱名称已存在
	ErrExtToolBoxCategoryTypeInvalid      ErrorCode = "ToolBoxCategoryTypeInvalid"      // 无效的工具箱类型
	ErrExtToolExists                      ErrorCode = "ToolExists"                      // 工具已存在
	ErrExtMetadataNotFound                ErrorCode = "MetadataNotFound"                // 元数据不存在
	ErrExtToolNotFound                    ErrorCode = "ToolNotFound"                    // 工具不存在
	ErrExtToolNotAvailable                ErrorCode = "ToolNotAvailable"                // 工具不可用
	ErrExtToolConvertOnlySupportSync      ErrorCode = "ToolConvertOnlySupportSync"      // 仅支持同步算子转换为工具
	ErrExtToolConvertOnlySupportAPI       ErrorCode = "ToolConvertOnlySupportAPI"       // 仅支持API算子转换为工具
	ErrExtToolBoxStatusInvalid            ErrorCode = "ToolBoxStatusInvalid"            // 工具箱状态无效
	ErrExtToolBoxNameEmpty                ErrorCode = "ToolBoxNameEmpty"                // 工具名称不能为空
	ErrExtToolBoxNameLimit                ErrorCode = "ToolBoxNameLimit"                // 工具名称长度不能超过%d个字符
	ErrExtToolBoxDescLimit                ErrorCode = "ToolBoxDescLimit"                // 工具描述长度不能超过%d个字符
	ErrExtToolNameEmpty                   ErrorCode = "ToolNameEmpty"                   // 工具名称不能为空
	ErrExtToolNameLimit                   ErrorCode = "ToolNameLimit"                   // 工具名称长度不能超过%d个字符
	ErrExtToolDescLimit                   ErrorCode = "ToolDescLimit"                   // 工具描述长度不能超过%d个字符
	ErrExtInternalToolBoxVersion          ErrorCode = "InternalToolBoxVersion"          // 内部工具箱版本号格式错误
	ErrExtToolNameDuplicate               ErrorCode = "ToolNameDuplicate"               // 工具名称重复
	ErrExtToolOperatorNotAllowEdit        ErrorCode = "ToolOperatorNotAllowEdit"        // 算子工具不允许编辑元数据
	ErrExtToolDescEmpty                   ErrorCode = "ToolDescEmpty"                   // 工具描述不能为空
	ErrExtToolBoxDescEmpty                ErrorCode = "ToolBoxDescEmpty"                // 工具箱描述不能为空
	ErrExtToolNotExistInFile              ErrorCode = "ToolNotExistInFile"              // 您上传的文件未包含已存在的工具
	ErrExtToolConvertMetadataTypeNotMatch ErrorCode = "ToolConvertMetadataTypeNotMatch" // 算子元数据类型与工具不匹配
	ErrExtToolTypeMismatch                ErrorCode = "ToolTypeMismatch"                // 工具类型与工具箱类型不匹配
	ErrExtToolRefOperatorNotFound         ErrorCode = "ToolRefOperatorNotFound"         // 工具“%s”不可启用，依赖的算子已被删除，请重新配置
)

// MCP拓展错误码定义
const (
	ErrExtMCPModeNotSupported      ErrorCode = "MCPModeNotSupported"      // MCP模式不支持
	ErrExtMCPExists                ErrorCode = "MCPExists"                // MCP已存在
	ErrExtMCPNotFound              ErrorCode = "MCPNotFound"              // MCP不存在
	ErrExtMCPStatusInvalid         ErrorCode = "MCPStatusInvalid"         // MCP状态无效
	ErrExtMCPNameEmpty             ErrorCode = "MCPNameEmpty"             // MCP名称不能为空
	ErrExtMCPNameLimit             ErrorCode = "MCPNameLimit"             // MCP名称长度不能超过%d个字符
	ErrExtMCPUnSupportEdit         ErrorCode = "MCPUnSupportEdit"         // MCP不支持编辑
	ErrExtMCPUnSupportDelete       ErrorCode = "MCPUnSupportDelete"       // 当前MCP不允许删除
	ErrExtMCPParseFailed           ErrorCode = "MCPParseFailed"           // MCP解析失败
	ErrExtMCPServerNotAccessible   ErrorCode = "MCPServerNotAccessible"   // MCP Server 无法访问
	ErrExtMCPListToolsFailed       ErrorCode = "MCPListToolsFailed"       // 无法获取当前MCP服务下的工具列表
	ErrExtMCPCallToolFailed        ErrorCode = "MCPCallToolFailed"        // 调用MCP工具失败
	ErrExtMCPDescLimit             ErrorCode = "MCPDescLimit"             // MCP描述长度不能超过%d个字符
	ErrExtMCPToolMaxCount          ErrorCode = "MCPToolMaxCount"          // MCP工具数量不能超过%d个
	ErrExtMCPToolNameDuplicate     ErrorCode = "MCPToolNameDuplicate"     // MCP工具名称重复
	ErrExtMCPInstanceAlreadyExists ErrorCode = "MCPInstanceAlreadyExists" // MCP实例已存在
	ErrExtMCPInstanceNotFound      ErrorCode = "MCPInstanceNotFound"      // MCP实例不存在
)

// 算子分类拓展错误码定义
const (
	ErrExtCategoryNameEmpty ErrorCode = "CategoryNameEmpty" // 算子分类名称不能为空
	ErrExtCategoryNameLimit ErrorCode = "CategoryNameLimit" // 算子分类名称长度不能超过%d个字符
	ErrExtCategoryNotFound  ErrorCode = "CategoryNotFound"  // 算子分类不存在
	ErrExtCategoryNameExist ErrorCode = "CategoryNameExist" // 算子分类名称已存在
)

// Skill拓展错误码定义
const (
	// 当前Agent Skill不允许删除
	ErrExtSkillUnSupportDelete ErrorCode = "SkillUnSupportDelete" // 当前Agent Skill不允许删除
	// 技能状态无效
	ErrExtSkillStatusInvalid ErrorCode = "SkillStatusInvalid" // 技能状态无效
	// 技能名称重复
	ErrExtSkillNameDuplicate ErrorCode = "SkillNameDuplicate" // 技能名称重复
	// 技能分类不存在
	ErrExtSkillCategoryNotFound ErrorCode = "SkillCategoryNotFound" // 技能分类不存在
)

// 代理模块错误码定义
const (
	// 请求转发失败，请检查是否可用，或稍后重试
	ErrExtProxyForwardFailed ErrorCode = "ProxyForwardFailed"
)

// common拓展错误码定义
const (
	ErrExtCommonOperationForbidden                ErrorCode = "CommonOperationForbidden"                // 没有操作权限
	ErrExtCommonAddForbidden                      ErrorCode = "CommonAddForbidden"                      // 没有新建权限
	ErrExtCommonEditForbidden                     ErrorCode = "CommonEditForbidden"                     // 没有编辑权限
	ErrExtCommonDeleteForbidden                   ErrorCode = "CommonDeleteForbidden"                   // 没有删除权限
	ErrExtCommonPublishForbidden                  ErrorCode = "CommonPublishForbidden"                  // 没有发布权限
	ErrExtCommonUnpublishForbidden                ErrorCode = "CommonUnpublishForbidden"                // 没有下架权限
	ErrExtCommonPermissionForbidden               ErrorCode = "CommonPermissionForbidden"               // 没有权限管理权限
	ErrExtCommonPublicAccessForbidden             ErrorCode = "CommonPublicAccessForbidden"             // 没有公共访问权限
	ErrExtCommonUseForbidden                      ErrorCode = "CommonUseForbidden"                      // 没有使用权限
	ErrExtCommonViewForbidden                     ErrorCode = "CommonViewForbidden"                     // 没有查看权限
	ErrExtCommonUserNotFound                      ErrorCode = "CommonUserNotFound"                      // 用户不存在
	ErrExtCommonAnonymousUserNotAllowed           ErrorCode = "CommonAnonymousUserNotAllowed"           // 匿名用户不允许访问
	ErrExtCommonDepartmentOrGroupOrRoleNotAllowed ErrorCode = "CommonDepartmentOrGroupOrRoleNotAllowed" // 部门/用户组/角色账户不允许访问
	ErrExtCommonInvalidAccessorType               ErrorCode = "CommonInvalidAccessorType"               // 无效账户类型
)

// 通用错误码定义
const (
	ErrExtCommonNameInvalid                 ErrorCode = "CommonNameInvalid"                 // 仅支持字母、数字、下划线
	ErrExtCommonResourceIDConflict          ErrorCode = "CommonResourceIDConflict"          // 资源ID冲突
	ErrExtCommonInternalComponentNotAllowed ErrorCode = "CommonInternalComponentNotAllowed" // 内置组件不允许导入导出
	ErrExtCommonImportDataEmpty             ErrorCode = "CommonImportDataEmpty"             // 导入数据为空
	ErrExtCommonNameExists                  ErrorCode = "CommonNameExists"                  // 此名称已被占用，请重新命名。
	ErrExtCommonNoMatchedMethodPath         ErrorCode = "CommonNoMatchedMethodPath"         // 未匹配到对应的API方法路
	ErrExtCommonCodeNotFound                ErrorCode = "CommonCodeNotFound"                // 调试模式下，代码不能为空
	ErrExtCommonMetadataTypeConflict        ErrorCode = "CommonMetadataTypeConflict"        // 元数据类型冲突
)

// 验证器错误码定义
const (
	ErrExtCodeValidationRequired ErrorCode = "ValidationRequired" // 必填项
	ErrExtCodeValidationFormat   ErrorCode = "ValidationFormat"   // 格式错误
	ErrExtCodeValidationRange    ErrorCode = "ValidationRange"    // 范围错误
	ErrExtCodeValidationEnum     ErrorCode = "ValidationEnum"     // 枚举错误
)

// openapi 错误码
const (
	// 加载阶段错误
	ErrExtOpenAPISyntaxInvalid ErrorCode = "OpenAPISyntaxInvalid" // 文件格式不正确，请检查是否符合OpenAPI 3.0规范

	// 验证阶段错误 - 支持参数的错误消息
	ErrExtOpenAPIInvalidPath                ErrorCode = "OpenAPIInvalidPath"                // API路径定义缺失或格式错误，请检查路径定义是否正确
	ErrExtOpenAPIInvalidParameterRequired   ErrorCode = "OpenAPIInvalidParameterRequired"   // 参数“%s”缺少必需字段，请检查是否有缺失参数
	ErrExtOpenAPIInvalidParameterSchema     ErrorCode = "OpenAPIInvalidParameterSchema"     // 参数“%s”Schema定义错误，请检查参数定义是否正确
	ErrExtOpenAPIInvalidParameterDefinition ErrorCode = "OpenAPIInvalidParameterDefinition" // 参数“%s”定义错误，请检查参数定义是否正确
	ErrExtOpenAPIInvalidParameterValue      ErrorCode = "OpenAPIInvalidParameterValue"      // Parameter校验错误，请查看错误详情
	ErrExtOpenAPIInvalidResponseRequired    ErrorCode = "OpenAPIInvalidResponseRequired"    // 响应“%s”缺少必需字段，请检查是否有缺失响应字段
	ErrExtOpenAPIInvalidResponseDefinition  ErrorCode = "OpenAPIInvalidResponseDefinition"  // 响应“%s”定义错误，请检查响应定义是否正确
	ErrExtOpenAPIInvalidResponseSchema      ErrorCode = "OpenAPIInvalidResponseSchema"      // 响应Schema定义错误，请查看错误详情
	ErrExtOpenAPIInvalidSchemaRef           ErrorCode = "OpenAPIInvalidSchemaRef"           // Schema“%s”引用错误，请检查$ref定义是否正确
	ErrExtOpenAPIInvalidSchemaType          ErrorCode = "OpenAPIInvalidSchemaType"          // Schema类型“%s”定义错误，请检查类型定义是否正确
	ErrExtOpenAPIInvalidSchemaValue         ErrorCode = "OpenAPIInvalidSchemaValue"         // Schema定义错误，请检查值定义是否正确
	ErrExtOpenAPIInvalidSpecification       ErrorCode = "OpenAPIInvalidSpecification"       // OpenAPI规范验证失败，请检查完整性
	ErrExtOpenAPIInvalidURLFormat           ErrorCode = "OpenAPIInvalidURLFormat"           // URL格式错误，请检查URL是否符合规范
	ErrExtOpenAPIInvalidComponent           ErrorCode = "OpenAPIInvalidComponent"           // 组件定义错误，请检查组件定义是否正确

	// 通用验证错误
	ErrExtOpenAPIInvalidSpecificationRequired  ErrorCode = "OpenAPIInvalidSpecificationRequired"  // 缺少必需字段“%s”，请检查是否有缺失字段
	ErrExtOpenAPIInvalidSpecificationMissing   ErrorCode = "OpenAPIInvalidSpecificationMissing"   // 字段“%s”缺失，请检查是否有缺失字段
	ErrExtOpenAPIInvalidSpecificationInvalid   ErrorCode = "OpenAPIInvalidSpecificationInvalid"   // 字段“%s”值无效，请检查是否有无效值
	ErrExtOpenAPIInvalidSpecificationDuplicate ErrorCode = "OpenAPIInvalidSpecificationDuplicate" // 字段“%s”重复，请检查是否有重复字段
	ErrExtOpenAPIInvalidSpecificationOperation ErrorCode = "OpenAPIInvalidSpecificationOperation" // 操作“%s”失败，请检查是否有其他错误

	// 自定义验证
	// Summary不允许为空
	ErrExtOpenAPIInvalidSpecificationSummaryEmpty ErrorCode = "OpenAPIInvalidSpecificationSummaryEmpty" // 该“%s”Summary为空, 请补充完整

	// 函数校验
	ErrExtFunctionNoHandlerFound                     ErrorCode = "FunctionNoHandlerFound"                     // 未检测到入口函数handler(event)，请检查函数代码是否正确
	ErrExtFunctionInvalidParameterType               ErrorCode = "FunctionInvalidParameterType"               // 参数“%s”类型无效: %s, 必须是 string, number, boolean, array, object 之一
	ErrExtFunctionInvalidParameterSubParameters      ErrorCode = "FunctionInvalidParameterSubParameters"      // 参数“%s”类型为 %s, 不支持 sub_parameters 字段, 只有 array 和 object 类型才能有子参数
	ErrExtFunctionInvalidParameterSubParametersCount ErrorCode = "FunctionInvalidParameterSubParametersCount" // 参数“%s”是 array 类型, sub_parameters 必须只包含一个元素来定义数组项的结构, 当前有 %d 个元素
)

// 业务域 错误码
const (
	// 业务域id必传
	ErrExtBusinessDomainIDRequired ErrorCode = "BusinessDomainIDRequired" // 请求头中缺少业务域ID，请确保在请求中包含 x-business-domain 头部参数
	// BusinessDomainForbidden 业务域权限不足
	ErrExtBusinessDomainForbidden ErrorCode = "BusinessDomainForbidden" // 您没有权限访问该业务域，请联系管理员获取权限
	// BusinessDomainResourceConflict 资源已关联到业务域
	ErrExtBusinessDomainResourceConflict ErrorCode = "BusinessDomainResourceConflict" // 该资源已关联到业务域，无需重复关联
)

// 依赖服务 错误码定义
const (
	// 沙箱函数运行报错
	ErrExtSandboxRuntimeExecuteCodeFailed ErrorCode = "SandboxRuntimeExecuteCodeFailed" // 沙箱函数运行报错，请检查代码是否正确
	ErrExtDebugParamsInvalid              ErrorCode = "DebugParamsInvalid"              // 调试传参错误，必须为JSON格式
	ErrExtFunctionAIGenerateFailed        ErrorCode = "FunctionAIGenerateFailed"        // AI生成失败，请检查默认模型是否正常
	ErrExtFunctionAIGenerateModelFailed   ErrorCode = "FunctionAIGenerateModelFailed"   // 模型生成内容异常，请检查默认模型是否可用，或前往设置配置有效的模型
	// 依赖沙箱服务异常
	ErrExtSandboxControlPlaneFailed ErrorCode = "SandboxControlPlaneFailed" // 依赖沙箱服务异常，请查看错误详情
	// Pypi源不可用
	ErrExtPypiRepoUnavailable ErrorCode = "PypiRepoUnavailable" // PyPI仓库不可用，请检查网络连接或稍后重试
	// Pypi源解析器错误
	ErrExtPypiParserFailed ErrorCode = "PypiParserFailed" // 版本联想功能需镜像源支持 JSON API (如 PyPI 官方、清华源等)，请检查镜像源配置是否正确
	// 请求OSS网关失败
	ErrExtOSSGatewayFailed ErrorCode = "OSSGatewayFailed" // 请求OSS网关失败，请检查网关配置是否正确
	// 默认OSS网关存储不存在
	ErrExtOSSGatewayDefaultStorageNotFound ErrorCode = "OSSGatewayDefaultStorageNotFound" // 默认OSS网关存储不存在，请检查网关配置是否正确
)

const (
	NoneErrorLink = "None"
)
