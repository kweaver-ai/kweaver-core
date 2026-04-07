package common

const (
	// 触发器类型

	// MannualTrigger 手动触发
	MannualTrigger = "@trigger/manual"
	// EventTrigger 事件触发
	EventTrigger = "@trigger/event"
	// CronTrigger 定时触发
	CronTrigger = "@trigger/cron"
	// CronWeekTrigger 定时每周触发
	CronWeekTrigger = "@trigger/cron/week"
	// CronMonthTrigger 定时每月触发
	CronMonthTrigger = "@trigger/cron/month"
	// CronCustomTrigger 定时自定义触发
	CronCustomTrigger = "@trigger/cron/custom"
	// WebhookTrigger 回调触发
	WebhookTrigger = "@trigger/webhook"
	// FormTrigger 表单触发
	FormTrigger = "@trigger/form"
	// AnyshareFileUploadTrigger 文件上传时触发
	AnyshareFileUploadTrigger = "@anyshare-trigger/upload-file"
	// AnyshareFileMoveTrigger 文件移动时触发
	AnyshareFileMoveTrigger = "@anyshare-trigger/move-file"
	// AnyshareFileCopyTrigger 文件复制时触发
	AnyshareFileCopyTrigger = "@anyshare-trigger/copy-file"
	// AnyshareFileRemoveTrigger 文件删除时触发
	AnyshareFileRemoveTrigger = "@anyshare-trigger/remove-file"
	// AnyshareFolderCreateTrigger 文件夹创建时触发
	AnyshareFolderCreateTrigger = "@anyshare-trigger/create-folder"
	// AnyshareFolderMoveTrigger 文件夹移动时触发
	AnyshareFolderMoveTrigger = "@anyshare-trigger/move-folder"
	// AnyshareFolderCopyTrigger 文件夹复制时触发
	AnyshareFolderCopyTrigger = "@anyshare-trigger/copy-folder"
	// AnyshareFolderRemoveTrigger 文件夹删除时触发
	AnyshareFolderRemoveTrigger = "@anyshare-trigger/remove-folder"
	// AnyshareFileRenameTrigger 文件重命名时触发
	AnyshareFileRenameTrigger = "@anyshare-trigger/rename-file"
	// AnyshareFileReversionTrigger 文件恢复版本时触发
	AnyshareFileReversionTrigger = "@anyshare-trigger/reversion-file"
	// AnyshareFileDeleteTrigger 文件彻底删除时触发
	AnyshareFileDeleteTrigger = "@anyshare-trigger/delete-file"
	// AnyshareFileRestoreTrigger 文件从回收站恢复时触发
	AnyshareFileRestoreTrigger = "@anyshare-trigger/restore-file"
	// AnyshareUserDeleteTrigger 删除用户时触发
	AnyshareUserDeleteTrigger = "@anyshare-trigger/delete-user"
	// AnyshareUserFreezeTrigger 冻结用户时触发
	AnyshareUserFreezeTrigger = "@anyshare-trigger/freeze-user"
	// AnyshareUserCreateTrigger 创建用户时触发
	AnyshareUserCreateTrigger = "@anyshare-trigger/create-user"
	// AnyshareOrgNameModifyTrigger 用户、部门、联系人名称变更
	AnyshareOrgNameModifyTrigger = "@anyshare-trigger/modify-org-name"
	// AnyshareDeptDeleteTrigger 删除部门时触发
	AnyshareDeptDeleteTrigger = "@anyshare-trigger/delete-dept"
	// AnyshareDeptCreateTrigger 创建部门时触发
	AnyshareDeptCreateTrigger = "@anyshare-trigger/create-dept"
	// AnyshareUserMovedTrigger 移动用户时触发
	AnyshareUserMovedTrigger = "@anyshare-trigger/move-user"
	// AnyshareDeptMovedTrigger 移动部门时触发
	AnyshareDeptMovedTrigger = "@anyshare-trigger/move-dept"
	// AnyshareUserAddDeptTrigger 添加用户到部门时触发
	AnyshareUserAddDeptTrigger = "@anyshare-trigger/add-user-to-dept"
	// AnyshareUserRemoveDeptTrigger 从部门移除用户时触发
	AnyshareUserRemoveDeptTrigger = "@anyshare-trigger/remove-user-from-dept"
	// AnyshareUserChangeTrigger 用户信息变更
	AnyshareUserChangeTrigger = "@anyshare-trigger/change-user"
	// AnyshareTagTreeCreateTrigger 官方标签树创建时触发
	AnyshareTagTreeCreateTrigger = "@anyshare-trigger/create-tag-tree"
	// AnyshareTagTreeAddedTrigger 增加标签时触发
	AnyshareTagTreeAddedTrigger = "@anyshare-trigger/add-tag-tree"
	// AnyshareTagTreeEditedTrigger 编辑标签时触发
	AnyshareTagTreeEditedTrigger = "@anyshare-trigger/edit-tag-tree"
	// AnyshareTagTreeDeletedTrigger 删除标签时触发
	AnyshareTagTreeDeletedTrigger = "@anyshare-trigger/delete-tag-tree"

	// AnyshareFileVersionUpdateTrigger 文档版本更新
	AnyshareFileVersionUpdateTrigger = "@anyshare-trigger/file-version-update"
	// AnyshareFilePathUpdateTrigger 文档路径更新
	AnyshareFilePathUpdateTrigger = "@anyshare-trigger/file-path-update"
	// AnyshareFileVersionDeleteTrigger 文档版本删除
	AnyshareFileVersionDeleteTrigger = "@anyshare-trigger/file-version-delete"
	// AnyshareUserUpdateDeptTrigger 用户所属部门信息更新
	AnyshareUserUpdateDeptTrigger = "@anyshare-trigger/user-update-dept"

	// dataflow trigger

	// DataflowDocTrigger 文档数据流触发
	DataflowDocTrigger = "@trigger/dataflow-doc"
	// DataflowUserTrigger 用户数据流触发
	DataflowUserTrigger = "@trigger/dataflow-user"
	// DataflowDeptTrigger 部门数据流触发
	DataflowDeptTrigger = "@trigger/dataflow-dept"
	// DataflowTagTrigger 标签数据流触发
	DataflowTagTrigger = "@trigger/dataflow-tag"
	// 通过算子获取数据
	OperatorTrigger = "@trigger/operator"

	// 数据源类型

	// AnyshareDataListFiles 列举文件
	AnyshareDataListFiles = "@anyshare-data/list-files"
	// AnyshareDataListFolders 列举文件夹
	AnyshareDataListFolders = "@anyshare-data/list-folders"
	// AnyshareDataSpecifyFiles 指定文件
	AnyshareDataSpecifyFiles = "@anyshare-data/specify-files"
	// AnyshareDataSpecifyFolders 指定文件夹
	AnyshareDataSpecifyFolders = "@anyshare-data/specify-folders"
	// SecurityPolicyTrigger 安全策略流程触发
	SecurityPolicyTrigger = "@trigger/security-policy"

	// AnyshareDataDepartment AnyShare部门组织
	AnyshareDataDepartment = "@anyshare-data/dept-tree"

	// AnyshareDataUser AnyShare用户
	AnyshareDataUser = "@anyshare-data/user"

	// AnyshareDataTagTree AnyShare标签树
	AnyshareDataTagTree = "@anyshare-data/tag-tree"

	// S3DataListObjects S3列表对象
	S3DataListObjects = "@s3/list-objects"

	// LocalDataListFiles 本地文件列表操作符
	LocalDataListFiles = "@local/list-files"

	// 数据视图触发流程配置数据源
	MDLDataViewTrigger = "@trigger/dataview"
)

var triggerMap = map[string][]string{
	TopicFileCopy:       {AnyshareFileCopyTrigger, AnyshareFileVersionUpdateTrigger},
	TopicFileMove:       {AnyshareFileMoveTrigger, AnyshareFilePathUpdateTrigger},
	TopicFileUpload:     {AnyshareFileUploadTrigger, AnyshareFileVersionUpdateTrigger},
	TopicFileCreate:     {AnyshareFileUploadTrigger, AnyshareFileVersionUpdateTrigger},
	TopicFileEdit:       {AnyshareFileUploadTrigger, AnyshareFileVersionUpdateTrigger},
	TopicFileRemove:     {AnyshareFileRemoveTrigger, AnyshareFileVersionDeleteTrigger},
	TopicFolderCreate:   {AnyshareFolderCreateTrigger},
	TopicFolderMove:     {AnyshareFolderMoveTrigger},
	TopicFolderRemove:   {AnyshareFolderRemoveTrigger},
	TopicFolderCopy:     {AnyshareFolderCopyTrigger},
	TopicFileReversion:  {AnyshareFileReversionTrigger, AnyshareFileVersionUpdateTrigger},
	TopicFileRename:     {AnyshareFileRenameTrigger, AnyshareFilePathUpdateTrigger},
	TopicFileDelete:     {AnyshareFileDeleteTrigger, AnyshareFileVersionDeleteTrigger},
	TopicFileRestore:    {AnyshareFileRestoreTrigger, AnyshareFileVersionUpdateTrigger},
	TopicUserDelete:     {AnyshareUserDeleteTrigger},
	TopicUserFreeze:     {AnyshareUserFreezeTrigger},
	TopicUserCreate:     {AnyshareUserCreateTrigger},
	TopicNameModify:     {AnyshareOrgNameModifyTrigger},
	TopicDeptDelete:     {AnyshareDeptDeleteTrigger},
	TopicDeptCreate:     {AnyshareDeptCreateTrigger},
	TopicUserMoved:      {AnyshareUserMovedTrigger, AnyshareUserUpdateDeptTrigger, AnyshareUserChangeTrigger},
	TopicDeptMoved:      {AnyshareDeptMovedTrigger},
	TopicUserAddDept:    {AnyshareUserAddDeptTrigger, AnyshareUserUpdateDeptTrigger, AnyshareUserChangeTrigger},
	TopicUserRemoveDept: {AnyshareUserRemoveDeptTrigger, AnyshareUserUpdateDeptTrigger, AnyshareUserChangeTrigger},
	TopicTagTreeCreated: {AnyshareTagTreeCreateTrigger},
	TopicTagTreeAdded:   {AnyshareTagTreeAddedTrigger},
	TopicTagTreeEdited:  {AnyshareTagTreeEditedTrigger},
	TopicTagTreeDeleted: {AnyshareTagTreeDeletedTrigger},
}

// GetTriggerTypeFromTopic 从topic转换任务触发类型
func GetTriggerTypeFromTopic(topic string) []string {
	triggerType, ok := triggerMap[topic]
	if !ok {
		return []string{}
	}

	return triggerType
}
