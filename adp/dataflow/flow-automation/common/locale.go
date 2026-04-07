package common

import (
	"fmt"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

var (
	// Languages 支持的语言
	Languages = [3]string{"zh_cn", "zh_tw", "en_us"}
)

// GetLogBody 获取日志信息
func GetLogBody(opt string, detailParams, extParams []interface{}) (detail, extMsg string) {
	lang := strings.ToLower(utils.GetLanguage())

	body, ok := EacpLogMap[opt][lang]
	if !ok {
		body = EacpLogMap[opt][Languages[0]]
	}
	detail = fmt.Sprintf(body[0], detailParams...)
	extMsg = fmt.Sprintf(body[1], extParams...)
	return
}

// GetEmailSubject 获取邮件主题国际化
func GetEmailSubject(opt string, params ...interface{}) string {
	lang := strings.ToLower(utils.GetLanguage())

	body, ok := EmailSubjectMap[opt][lang]
	if !ok {
		body = EmailSubjectMap[opt][Languages[0]]
	}

	return fmt.Sprintf(body, params...)
}

// GetDocTriggerLogBody 文档事件触发流程时日志
func GetDocTriggerLogBody(tp, name string, msg *DocMsg) (detail, extMsg string) {
	lang := strings.ToLower(utils.GetLanguage())
	logs := map[string][]string{
		Languages[0]: {fmt.Sprintf("触发 自动任务\"<%v>\" 成功", name)},
		Languages[1]: {fmt.Sprintf("觸發 自動任務\"<%v>\" 成功", name)},
		Languages[2]: {fmt.Sprintf("Trigger automatic task\"<%v>\" succeeded", name)},
	}
	switch tp {
	case AnyshareFileCopyTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 复制文件\"%v\"到\"%v\"", msg.Path, msg.NewPath))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 複製文件\"%v\"到\"%v\"", msg.Path, msg.NewPath))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: copy file \"%v\" to \"%v\"", msg.Path, msg.NewPath))
	case AnyshareFileMoveTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 移动文件\"%v\"到\"%v\"", msg.Path, msg.NewPath))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 移動文件\"%v\"到\"%v\"", msg.Path, msg.NewPath))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: move file \"%v\" to \"%v\"", msg.Path, msg.NewPath))
	case AnyshareFileUploadTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 上传文件\"%v\"", msg.Path))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 上傳文件\"%v\"", msg.Path))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: upload file \"%v\"", msg.Path))
	case AnyshareFileRemoveTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 删除文件\"%v\"", msg.Path))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 刪除文件\"%v\"", msg.Path))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: delete file \"%v\"", msg.Path))
	case AnyshareFolderCreateTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 创建文件夹\"%v\"", msg.Path))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 創建文件夾\"%v\"", msg.Path))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: create folder \"%v\"", msg.Path))
	case AnyshareFolderMoveTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 移动文件夹\"%v\"到\"%v\"", msg.Path, msg.NewPath))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 移動文件夾\"%v\"到\"%v\"", msg.Path, msg.NewPath))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: move file \"%v\" to \"%v\"", msg.Path, msg.NewPath))
	case AnyshareFolderRemoveTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 删除文件夹\"%v\"", msg.Path))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 删除文件夹\"%v\"", msg.Path))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: delete folder \"%v\"", msg.Path))
	case AnyshareFolderCopyTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 复制文件夹\"%v\"到\"%v\"", msg.DocID, msg.NewID))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 複製文件夾\"%v\"到\"%v\"", msg.DocID, msg.NewID))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: copy folder \"%v\" to \"%v\"", msg.DocID, msg.NewID))
	case AnyshareFileReversionTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 还原文件\"%v\"版本为\"%v\"", msg.DocID, msg.Rev))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 還原文件\"%v\"版本為\"%v\"", msg.DocID, msg.Rev))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: restore version of file \"%v\" to \"%v\"", msg.DocID, msg.Rev))
	case AnyshareFileRestoreTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 从回收站还原文件\"%v\"", msg.Path))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 從恢復站文件\"%v\"", msg.Path))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: restore files \"%v\" from Recycle Bin", msg.Path))
	case AnyshareFileDeleteTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 彻底删除文件\"%v\"", msg.Path))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 徹底刪除文件\"%v\"", msg.Path))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: delete file completely \"%v\"", msg.Path))
	case AnyshareFileRenameTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 重命名文件\"%v\"为\"%v\"", msg.Path, msg.NewPath))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 重新命名文件\"%v\"为\"%v\"", msg.Path, msg.NewPath))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: Rename file \"%v\" as \"%v\"", msg.Path, msg.NewPath))
	default:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: %s, 文件ID: %s, 文件路径: %s, 文件名: %s", tp, msg.DocID, msg.Path, msg.DocName))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: %s, 文件ID: %s, 文件路徑: %s, 文件名稱: %s", tp, msg.DocID, msg.Path, msg.DocName))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: %s, file ID: %s, file path: %s, file name: %s", tp, msg.DocID, msg.Path, msg.DocName))
	}

	body, ok := logs[lang]
	if !ok {
		body = logs[Languages[0]]
	}

	return body[0], body[1]
}

// GetUserInfoTriggerLogBody 用户信息事件触发流程时日志
func GetUserInfoTriggerLogBody(tp, name string, msg *UserInfoMsg) (detail, extMsg string) {
	lang := strings.ToLower(utils.GetLanguage())
	logs := map[string][]string{
		Languages[0]: {fmt.Sprintf("触发 自动任务\"<%v>\" 成功", name)},
		Languages[1]: {fmt.Sprintf("觸發 自動任務\"<%v>\" 成功", name)},
		Languages[2]: {fmt.Sprintf("Trigger automatic task\"<%v>\" succeeded", name)},
	}
	switch tp {
	case AnyshareUserCreateTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 创建用户\"%v\", 详细信息: \"%v\"", msg.Name, msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 創建用戶\"%v\", 詳細資訊: \"%v\"", msg.Name, msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: create user \"%v\", details: \"%v\"", msg.Name, msg))
	case AnyshareUserDeleteTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 删除用户\"%v\", 详细信息: \"%v\"", msg.Name, msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 刪除用戶\"%v\", 詳細資訊: \"%v\"", msg.Name, msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: delete user \"%v\", details: \"%v\"", msg.Name, msg))
	case AnyshareOrgNameModifyTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 用户、部门、联系人名称变更为\"%v\", 详细信息: \"%v\"", msg.Name, msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 使用者、部門、聯絡人名稱變更為\"%v\", 詳細資訊: \"%v\"", msg.Name, msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: the user, department, and contact names are changed to \"%v\", details: \"%v\"", msg.Name, msg))
	case AnyshareDeptDeleteTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 删除部门\"%v\", 详细信息: \"%v\"", msg.Name, msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 刪除部門\"%v\", 詳細資訊: \"%v\"", msg.Name, msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: delete department \"%v\", details: \"%v\"", msg.Name, msg))
	case AnyshareDeptCreateTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 创建部门\"%v\", 详细信息: \"%v\"", msg.Name, msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 創建部門\"%v\", 詳細資訊: \"%v\"", msg.Name, msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: create department \"%v\", details: \"%v\"", msg.Name, msg))
	case AnyshareUserMovedTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 移动用户\"%v\", 详细信息: \"%v\"", msg.Name, msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 行動用戶\"%v\", 詳細資訊: \"%v\"", msg.Name, msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: move user \"%v\", details: \"%v\"", msg.Name, msg))
	case AnyshareUserAddDeptTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 添加用户到部门, 详细信息: \"%v\"", msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 新增用戶到部門, 詳細資訊: \"%v\"", msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: add user to department, details: \"%v\"", msg))
	case AnyshareUserRemoveDeptTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 从部门移除用户, 详细信息: \"%v\"", msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 从部门删除用户, 詳細資訊: \"%v\"", msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: remove user from department, details: \"%v\"", msg))
	case AnyshareUserChangeTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 用户信息变更, 详细信息: \"%v\"", msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 使用者資訊變更, 詳細資訊: \"%v\"", msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: user information changes, details: \"%v\"", msg))
	default:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: %s, 详细信息 %v", tp, msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: %s, 詳細資訊: \"%v\"", tp, msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: %s, details: \"%v\"", tp, msg))
	}

	body, ok := logs[lang]
	if !ok {
		body = logs[Languages[0]]
	}

	return body[0], body[1]
}

// GetTagTreeTriggerLogBody 官方标签树事件触发流程时日志
func GetTagTreeTriggerLogBody(tp, name string, msg interface{}) (detail, extMsg string) {
	lang := strings.ToLower(utils.GetLanguage())
	logs := map[string][]string{
		Languages[0]: {fmt.Sprintf("触发 自动任务\"<%v>\" 成功", name)},
		Languages[1]: {fmt.Sprintf("觸發 自動任務\"<%v>\" 成功", name)},
		Languages[2]: {fmt.Sprintf("Trigger automatic task\"<%v>\" succeeded", name)},
	}
	switch tp {
	case AnyshareTagTreeCreateTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 创建官方标签树, 详细信息: \"%v\"", msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 創建官方標籤樹, 詳細資訊: \"%v\"", msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: create official tag tree, details: \"%v\"", msg))
	case AnyshareTagTreeAddedTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 增加标签, 详细信息: \"%v\"", msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 增加標籤, 詳細資訊: \"%v\"", msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: add tags, details: \"%v\"", msg))
	case AnyshareTagTreeEditedTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 编辑标签, 详细信息: \"%v\"", msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 編輯標籤, 詳細資訊: \"%v\"", msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: edit tags, details: \"%v\"", msg))
	case AnyshareTagTreeDeletedTrigger:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: 删除标签, 详细信息: \"%v\"", msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: 刪除標籤, 詳細資訊: \"%v\"", msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: delete tags, details: \"%v\"", msg))
	default:
		logs[Languages[0]] = append(logs[Languages[0]], fmt.Sprintf("触发条件: %s, 详细信息 %v", tp, msg))
		logs[Languages[1]] = append(logs[Languages[1]], fmt.Sprintf("觸發條件: %s, 詳細資訊: \"%v\"", tp, msg))
		logs[Languages[2]] = append(logs[Languages[2]], fmt.Sprintf("Triggering conditions: %s, details: \"%v\"", tp, msg))
	}

	body, ok := logs[lang]
	if !ok {
		body = logs[Languages[0]]
	}

	return body[0], body[1]
}

// EacpLogMap 日志国际化
var EacpLogMap = map[string]map[string][]string{
	"createTask": {
		Languages[0]: []string{"新建 自动任务\"<%v>\" 成功", ""},
		Languages[1]: []string{"新建 自動任務\"<%v>\" 成功", ""},
		Languages[2]: []string{"Create automatic task\"<%v>\" succeeded", ""},
	},
	"updateTask": {
		Languages[0]: []string{"编辑 自动任务\"<%v>\" 成功", ""},
		Languages[1]: []string{"編輯 自動任務\"<%v>\" 成功", ""},
		Languages[2]: []string{"Update automatic task\"<%v>\" succeeded", ""},
	},
	"deleteTask": {
		Languages[0]: []string{"删除 自动任务\"<%v>\" 成功", ""},
		Languages[1]: []string{"刪除 自動任務\"<%v>\" 成功", ""},
		Languages[2]: []string{"Delete automatic task\"<%v>\"", ""},
	},
	TriggerTaskManually: {
		Languages[0]: []string{"触发 自动任务\"<%v>\" 成功", "触发条件: 手动触发"},
		Languages[1]: []string{"觸發 自動任務\"<%v>\" 成功", "觸發條件: 手動觸發"},
		Languages[2]: []string{"Trigger automatic task\"<%v>\" succeeded", "Triggering conditions: manual trigger"},
	},
	TriggerTaskCron: {
		Languages[0]: []string{"触发 自动任务\"<%v>\" 成功", "触发条件: 定时触发"},
		Languages[1]: []string{"觸發 自動任務\"<%v>\" 成功", "觸發條件: 定時觸發"},
		Languages[2]: []string{"Trigger automatic task\"<%v>\" succeeded", "Triggering conditions: timing trigger"},
	},
	CompleteTaskWithSuccess: {
		Languages[0]: []string{"执行 自动任务\"<%v>\" 完成", "执行结果: 成功"},
		Languages[1]: []string{"執行 自動任務\"<%v>\" 完成", "執行結果：成功"},
		Languages[2]: []string{"Run automatic task\"<%v>\" completed", "Execution result: success"},
	},
	CompleteTaskWithFailed: {
		Languages[0]: []string{"执行 自动任务\"<%v>\" 完成", "执行结果: 失败"},
		Languages[1]: []string{"執行 自動任務\"<%v>\" 完成", "執行結果：失敗"},
		Languages[2]: []string{"Run automatic task\"<%v>\" completed", "Execution result: failed"},
	},
	"CreateSecurityPolicyFlow": {
		Languages[0]: []string{"创建 执行规则\"<%v>\" 成功", ""},
		Languages[1]: []string{"創建 执行规则\"<%v>\" 成功", ""},
		Languages[2]: []string{"Create execution rule\"<%v>\" succeeded", ""},
	},
	"UpdateSecurityPolicyFlow": {
		Languages[0]: []string{"编辑 执行规则\"<%v>\" 成功", ""},
		Languages[1]: []string{"編輯 执行规则\"<%v>\" 成功", ""},
		Languages[2]: []string{"Update execution rule\"<%v>\" succeeded", ""},
	},
	"DeleteSecurityPolicyFlow": {
		Languages[0]: []string{"删除 执行规则\"<%v>\" 成功", ""},
		Languages[1]: []string{"刪除 执行规则\"<%v>\" 成功", ""},
		Languages[2]: []string{"Delete execution rule\"<%v>\"", ""},
	},
	"RunSecurityPolicyFlow": {
		Languages[0]: []string{"触发 执行规则\"<%v>\" 成功", ""},
		Languages[1]: []string{"觸發 执行规则\"<%v>\" 成功", ""},
		Languages[2]: []string{"Trigger execution rule\"<%v>\" succeeded", ""},
	},
	"RunSecurityPolicyFlowSuccess": {
		Languages[0]: []string{"执行 执行规则\"<%v>\" 完成", "执行结果: 成功"},
		Languages[1]: []string{"執行 执行规则\"<%v>\" 完成", "執行結果：成功"},
		Languages[2]: []string{"Run execution rule \"<%v>\" completed", "Execution result: success"},
	},
	"RunSecurityPolicyFlowFailed": {
		Languages[0]: []string{"执行 执行规则\"<%v>\" 完成", "执行结果: 失败"},
		Languages[1]: []string{"執行 执行规则\"<%v>\" 完成", "執行結果：失敗"},
		Languages[2]: []string{"Run execution rule \"<%v>\" completed", "Execution result: failed"},
	},
	"createCustomCapabily": {
		Languages[0]: []string{"新建 新建自定义能力\"<%v>\" 成功", ""},
		Languages[1]: []string{"建立 新建自訂能力\"<%v>\" 成功", ""},
		Languages[2]: []string{"Custom AI \"<%v>\" has been created successfully", ""},
	},
	"updateCustomCapabily": {
		Languages[0]: []string{"编辑 新建自定义能力\"<%v>\" 成功", ""},
		Languages[1]: []string{"編輯 新建自訂能力\"<%v>\" 成功", ""},
		Languages[2]: []string{"Custom AI \"<%v>\" has been edited successfully", ""},
	},
	"deleteCustomCapabily": {
		Languages[0]: []string{"删除  新建自定义能力\"<%v>\" 成功", ""},
		Languages[1]: []string{"删除  新建自訂能力\"<%v>\" 成功", ""},
		Languages[2]: []string{"Custom AI \"<%v>\" has been deleted successfully", ""},
	},
	"createCustomExecutor": {
		Languages[0]: []string{"新建 自定义节点\"<%v>\" 成功", "可用范围: %v; 状态: %v; 自定义动作: %v"},
		Languages[1]: []string{"新建 自訂節點\"<%v>\" 成功", "可用範圍: %v; 狀態: %v; 自訂動作: %v"},
		Languages[2]: []string{"The custom node \"<%v>\" is created successfully", "Available scope: %v; Status: %v; Custom action: %v"},
	},
	"updateCustomExecutor": {
		Languages[0]: []string{"编辑 自定义节点\"<%v>\" 成功", "可用范围: %v; 状态: %v; 自定义动作: %v"},
		Languages[1]: []string{"编辑 自定义节点\"<%v>\" 成功", "可用範圍: %v; 狀態: %v; 自訂動作: %v"},
		Languages[2]: []string{"The custom node \"<%v>\" is edited successfully", "Available scope: %v; Status: %v; Custom action: %v"},
	},
	"deleteCustomExecutor": {
		Languages[0]: []string{"删除 自定义节点\"<%v>\" 成功", "状态: %v"},
		Languages[1]: []string{"刪除 自訂節點\"<%v>\" 成功", "狀態: %v"},
		Languages[2]: []string{"The custom node \"<%v>\" is deleted successfully", "Status: %v"},
	},
	"createCustomExecutorAction": {
		Languages[0]: []string{"新建 自定义动作\"<%v>\" 成功", ""},
		Languages[1]: []string{"新建 自訂動作\"<%v>\" 成功", ""},
		Languages[2]: []string{"The custom action \"<%v>\" is created successfully", ""},
	},
	"updateCustomExecutorAction": {
		Languages[0]: []string{"编辑 自定义动作\"<%v>\" 成功", ""},
		Languages[1]: []string{"編輯 自訂動作\"<%v>\" 成功", ""},
		Languages[2]: []string{"The custom action \"<%v>\" is edited successfully", ""},
	},
	"deleteCustomExecutorAction": {
		Languages[0]: []string{"删除 自定义动作\"<%v>\" 成功", ""},
		Languages[1]: []string{"刪除 自訂動作\"<%v>\" 成功", ""},
		Languages[2]: []string{"The custom action \"<%v>\" is deleted successfully", ""},
	},
	"startRunDag": {
		Languages[0]: []string{"开始执行 自动任务\"<%v>\" 成功", ""},
		Languages[1]: []string{"開始執行 自動任務\"<%v>\" 成功", ""},
		Languages[2]: []string{"Started to execute automatic task\"<%v>\" succeeded", ""},
	},
	"cancelRunningInstance": {
		Languages[0]: []string{"取消执行 自动任务\"<%v>\" 成功", "执行结果: 取消"},
		Languages[1]: []string{"取消執行 自動任務\"<%v>\" 成功", "執行結果：取消"},
		Languages[2]: []string{"Canceled the execution of automatic task\"<%v>\"", "Execution result: canceled"},
	},
}

// LocaleMap 审核相关国际化资源
var LocaleMap = map[string]map[string]string{
	"doc_name": {
		Languages[0]: "文档名称",
		Languages[1]: "文件名稱",
		Languages[2]: "File name",
	},
	"doc_id": {
		Languages[0]: "文件唯一标识",
		Languages[1]: "檔案唯一標識",
		Languages[2]: "Doc ID",
	},
	"doc_path": {
		Languages[0]: "文档路径",
		Languages[1]: "文件路徑",
		Languages[2]: "Document Location",
	},
	"forever": {
		Languages[0]: "永久有效",
		Languages[1]: "永久有效",
		Languages[2]: "Never",
	},
	"deny_all": {
		Languages[0]: "拒绝访问",
		Languages[1]: "拒絕存取",
		Languages[2]: "No Access",
	},
	"deny": {
		Languages[0]: "拒绝",
		Languages[1]: "拒絕",
		Languages[2]: "Deny",
	},
	"display": {
		Languages[0]: "显示",
		Languages[1]: "顯示",
		Languages[2]: "Display",
	},
	"preview": {
		Languages[0]: "预览",
		Languages[1]: "預覽",
		Languages[2]: "Preview",
	},
	"cache": {
		Languages[0]: "缓存",
		Languages[1]: "快取",
		Languages[2]: "Cache",
	},
	"download": {
		Languages[0]: "下载",
		Languages[1]: "下載",
		Languages[2]: "Download",
	},
	"create": {
		Languages[0]: "创建",
		Languages[1]: "新增",
		Languages[2]: "Create",
	},
	"modify": {
		Languages[0]: "修改",
		Languages[1]: "修改",
		Languages[2]: "Modify",
	},
	"delete": {
		Languages[0]: "删除",
		Languages[1]: "刪除",
		Languages[2]: "Delete",
	},
	"enabled": {
		Languages[0]: "启用中",
		Languages[1]: "啟用中",
		Languages[2]: "Enabled",
	},
	"disabled": {
		Languages[0]: "已停用",
		Languages[1]: "已停用",
		Languages[2]: "Disabled",
	},
}

// GetLocale 获取国际化资源
func GetLocale(key string) string {
	lang := strings.ToLower(utils.GetLanguage())

	body, ok := LocaleMap[key][lang]
	if !ok {
		body = LocaleMap[key][Languages[0]]
	}
	return body
}

// EmailSubjectMap 邮件主题多语言资源
var EmailSubjectMap = map[string]map[string]string{
	"notifyToExecutor": {
		Languages[0]: "工作流程运行失败提醒",
		Languages[1]: "工作流程運行失敗提醒",
		Languages[2]: "Reminder of flow failure",
	},
}
