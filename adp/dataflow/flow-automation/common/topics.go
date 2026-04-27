// Package common Topics定义
package common

var (
	// 文档操作相关topic

	// TopicFolderCreate 文件夹创建消息
	TopicFolderCreate = "core.folder.create"
	// TopicFolderRemove 文件夹删除消息
	TopicFolderRemove = "core.folder.remove"
	// TopicFolderMove 文件夹移动消息
	TopicFolderMove = "core.folder.move"
	// TopicFolderCopy 文件夹复制消息
	TopicFolderCopy = "core.folder.copy"
	// TopicFolderRestore 文件夹恢复消息
	TopicFolderRestore = "core.folder.restore"
	// TopicFileUpload 文件上传消息
	TopicFileUpload = "core.file.dupload"
	// TopicFileCreate 文件创建消息
	TopicFileCreate = "core.file.create"
	// TopicFileMove 文件移动消息
	TopicFileMove = "core.file.move"
	// TopicFileRemove 文件删除消息
	TopicFileRemove = "core.file.remove"
	// TopicFileCopy 文件复制消息
	TopicFileCopy = "core.file.copy.version"
	// TopicFileEdit 文件编辑消息
	TopicFileEdit = "core.file.edit"
	// TopicFileRename 文件重命名消息
	TopicFileRename = "core.file.rename"
	// TopicFileReversion 文件恢复版本
	TopicFileReversion = "core.file.restore.reversion"
	// TopicFileDelete 文件彻底删除
	TopicFileDelete = "core.file.delete"
	// TopicFileRestore 文件从回收站还原
	TopicFileRestore = "core.file.restore"

	// 人员变更相关事件topic

	// TopicUserDelete 删除用户
	TopicUserDelete = "core.user.delete"
	// TopicUserFreeze 冻结用户
	TopicUserFreeze = "core.user.freeze"
	// TopicUserCreate 创建用户
	TopicUserCreate = "core.user_management.user.created"
	// TopicNameModify 用户、部门、联系人名称变更
	TopicNameModify = "core.org.name.modify"

	// 组织变更事件

	// TopicDeptDelete 删除部门
	TopicDeptDelete = "core.dept.delete"
	// TopicDeptCreate 创建部门
	TopicDeptCreate = "core.user_management.dept.created"
	// TopicDeptMoved 移动部门
	TopicDeptMoved = "user_management.dept.moved"

	// 人员组织关系变更

	// TopicUserMoved 移动用户
	TopicUserMoved = "user_management.user.moved"
	// TopicUserAddDept 添加用户到部门
	TopicUserAddDept = "user_management.department.user.added"
	// TopicUserRemoveDept 从部门移除用户
	TopicUserRemoveDept = "user_management.department.user.removed"

	// 官方标签变更

	// TopicTagTreeCreated 创建标签树
	TopicTagTreeCreated = "metadata.tag_tree.created"
	// TopicTagTreeAdded 在标签树中添加标签
	TopicTagTreeAdded = "metadata.tag_tree.tag.added"
	// TopicTagTreeEdited 修改标签树中的标签
	TopicTagTreeEdited = "metadata.tag_tree.tag.edited"
	// TopicTagTreeDeleted 删除标签树中的标签
	TopicTagTreeDeleted = "metadata.tag_tree.tag.deleted"

	// 审核相关topic

	// TopicWorkflowApply 发起审核操作topic
	TopicWorkflowApply = "workflow.audit.apply"
	// TopicWorkflowResult 订阅审核结果topic
	TopicWorkflowResult = "workflow.audit.result.automation"
	// TopicWorkflowAuditor 订阅匹配到审核员
	TopicWorkflowAuditor = "workflow.audit.auditor.automation"
	// TopicWorkflowUpdate 更新Workflow内容
	TopicWorkflowUpdate = "workflow.audit.update"

	TopicContentPipelineFulltextResult         = "pipe.task.full_text"
	TopicContentPipelineDocFormatConvertResult = "pipe.task.doc_format_convert"
	TopicAsyncTaskResult                       = "automation.async_task.result"

	TopicWorkflowSecurityPolicyPermResult              = "workflow.audit.result.security_policy_perm"
	TopicWorkflowSecurityPolicyPermAuditor             = "workflow.audit.auditor.security_policy_perm"
	TopicWorkflowSecurityPolicyUploadResult            = "workflow.audit.result.security_policy_upload"
	TopicWorkflowSecurityPolicyUploadAuditor           = "workflow.audit.auditor.security_policy_upload"
	TopicWorkflowSecurityPolicyDeleteResult            = "workflow.audit.result.security_policy_delete"
	TopicWorkflowSecurityPolicyDeleteAuditor           = "workflow.audit.auditor.security_policy_delete"
	TopicWorkflowSecurityPolicyRenameResult            = "workflow.audit.result.security_policy_rename"
	TopicWorkflowSecurityPolicyRenameAuditor           = "workflow.audit.auditor.security_policy_rename"
	TopicWorkflowSecurityPolicyCopyResult              = "workflow.audit.result.security_policy_copy"
	TopicWorkflowSecurityPolicyCopyAuditor             = "workflow.audit.auditor.security_policy_copy"
	TopicWorkflowSecurityPolicyMoveResult              = "workflow.audit.result.security_policy_move"
	TopicWorkflowSecurityPolicyMoveAuditor             = "workflow.audit.auditor.security_policy_move"
	TopicWorkflowSecurityPolicyFolderPropertiesResult  = "workflow.audit.result.security_policy_modify_folder_property"
	TopicWorkflowSecurityPolicyFolderPropertiesAuditor = "workflow.audit.auditor.security_policy_modify_folder_property"

	// 撤销审核
	TopicWorkflowCancel = "workflow.audit.cancel"

	// 安全策略运行结果 topic
	TopicSecurityPolicyProcResult = "default.as.automation.security_policy_proc_result"

	// 安全策略发起审核 topic
	TopicSecurityPolicyProcApproval = "default.as.automation.security_policy_proc_approval"

	TopicSecurityPolicyProcCreate = "default.as.automation.security_policy_proc_create"
	TopicSecurityPolicyProcStatus = "default.as.automation.security_policy_proc_status"
	TopicSecurityPolicyFlowDelete = "default.as.automation.security_policy_flow_delete"
	// TopicAuditLog 审计日志 topic
	TopicAuditLog = "as.audit_log.log_operation"
	// TopicDIPFlowAuditLog 审计日志 topic
	TopicDIPFlowAuditLog = "isf.audit_log.log"
	// TopicARLog ar日志上报 topic
	TopicARLog = "as.audit_log.operation_log.content_automation"

	// 知识图谱提取结果 topic
	TopicDocSetGraphInfoResult = "docset.graph_info.data"

	TopicDatahubIndexNotify = "default.as.datahubcentral.indextrace.notify"

	// TopicFlowDelete 流程删除
	TopicFlowDelete = "default.as.automation.flow_delete"

	// TopicFlowActive 流程状态激活
	TopicFlowActive = "default.as.automation.flow_active"

	// TopicFlowSuspended 流程状态休眠
	TopicFlowSuspended = "default.as.automation.flow_suspended"

	// TopicAuthorizationNameModify 策略名称变更
	TopicAuthorizationNameModify = "authorization.resource.name.modify"

	// TopicOperatorDelete 算子删除
	TopicOperatorDelete = "agent_operator_integration.operator.delete"

	// TopicFlowFileDownloadResult 文件下载完成结果
	TopicFlowFileDownloadResult = "automation.flow_file.download.result"
)

// DocMsg doc消息体
type DocMsg struct {
	DocID    string   `json:"id"`
	DocName  string   `json:"name"`
	Rev      string   `json:"rev"`
	OssID    string   `json:"oss_id"`
	Path     string   `json:"path"`
	Size     int64    `json:"size"`
	NewID    string   `json:"new_id"`
	NewPath  string   `json:"new_path"`
	Cover    bool     `json:"cover"`
	Operator Operator `json:"operator"`
}

// Operator 操作人
type Operator struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// UserInfoMsg 用户信息消息体
type UserInfoMsg struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	NewName     string   `json:"new_name,omitempty"`
	Type        string   `json:"type,omitempty"`
	OldDeptPath string   `json:"old_dept_path,omitempty"`
	NewDeptPath string   `json:"new_dept_path,omitempty"`
	DeptPaths   []string `json:"dept_paths,omitempty"`

	Email            string `json:"email,omitempty"`
	Tags             string `json:"tags,omitempty"`
	IsExpert         bool   `json:"is_expert,omitempty"`
	VerificationInfo string `json:"verification_info,omitempty"`
	University       string `json:"university,omitempty"`
	Contact          string `json:"contact,omitempty"`
	Position         string `json:"position,omitempty"`
	WorkAt           string `json:"work_at,omitempty"`
	IsDelete         int    `json:"is_delete,omitempty"`
	Professional     string `json:"professional,omitempty"`
	Status           int    `json:"status,omitempty"`
}

// TagInfo 标签信息
type TagInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Version  int    `json:"version"`
	ParentID string `json:"parent_id"`
}
