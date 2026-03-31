package common

const (
	// MannualTriggerOpt 手动触发器
	MannualTriggerOpt = "@trigger/manual"
	// AnyshareFileCopyOpt 复制文件
	AnyshareFileCopyOpt = "@anyshare/file/copy"
	// AnyshareFileMoveOpt 移动文件
	AnyshareFileMoveOpt = "@anyshare/file/move"
	// AnyshareFileRemoveOpt 移除文件
	AnyshareFileRemoveOpt = "@anyshare/file/remove"
	// AnyshareFileRenameOpt 重命名文件
	AnyshareFileRenameOpt = "@anyshare/file/rename"
	// AnyshareFileAddTagsOpt 文件添加标签
	AnyshareFileAddTagsOpt = "@anyshare/file/addtag"
	// AnyshareFileGetPathOpt 获取文件路径
	AnyshareFileGetPathOpt = "@anyshare/file/getpath"
	// AnyshareFileMatchContentOpt 匹配文件内容
	AnyshareFileMatchContentOpt = "@anyshare/file/matchcontent"
	// AnyshareFileSetCsfLevelOpt 设置文件密级
	AnyshareFileSetCsfLevelOpt = "@anyshare/file/setcsflevel"
	// AnyshareFileSetTemplate 给文件添加编目
	AnyshareFileSetTemplate = "@anyshare/file/settemplate"
	// AnyshareFileGetPage 给文件添加编目
	AnyshareFileGetPage = "@anyshare/file/getpage"
	// AnyshareFileCreateOpt 新建文件
	AnyshareFileCreateOpt = "@anyshare/file/create"
	// AnyshareFileEditOpt 更新文件
	AnyshareFileEditOpt = "@anyshare/file/edit"
	// AnyshareExcelFileEditOpt 编辑excel文件
	AnyshareExcelFileEditOpt = "@anyshare/file/editexcel"
	// AnyshareDocxFileEditOpt 编辑docx文件
	AnyshareDocxFileEditOpt = "@anyshare/file/editdocx"
	// AnyshareFileOCROpt 文件OCR识别
	AnyshareFileOCROpt = "@anyshare/ocr/general"
	// AnyshareEleInvoiceOpt 电子发票识别
	AnyshareEleInvoiceOpt = "@anyshare/ocr/eleinvoice"
	// AnyshareIDCardOpt 身份证识别
	AnyshareIDCardOpt = "@anyshare/ocr/idcard"
	// AnyshareFileSetPermOpt 配置权限
	AnyshareFileSetPermOpt = "@anyshare/file/perm"

	// AnyshareFolderCopyOpt 复制文件夹
	AnyshareFolderCopyOpt = "@anyshare/folder/copy"
	// AnyshareFolderMoveOpt 移动文件夹
	AnyshareFolderMoveOpt = "@anyshare/folder/move"
	// AnyshareFolderRemoveOpt 移除文件夹
	AnyshareFolderRemoveOpt = "@anyshare/folder/remove"
	// AnyshareFolderRenameOpt 重命名文件夹
	AnyshareFolderRenameOpt = "@anyshare/folder/rename"
	// AnyshareFloderCreateOpt 创建文件夹
	AnyshareFloderCreateOpt = "@anyshare/folder/create"
	// AnyshareFolderAddTagsOpt 文件夹添加标签
	AnyshareFolderAddTagsOpt = "@anyshare/folder/addtag"
	// AnyshareFolderGetPathOpt 获取文件夹路径
	AnyshareFolderGetPathOpt = "@anyshare/folder/getpath"
	// AnyshareFolderSetTemplate 给文件夹添加编目
	AnyshareFolderSetTemplate = "@anyshare/folder/settemplate"
	// AnyshareFolderSetPermOpt 配置权限
	AnyshareFolderSetPermOpt = "@anyshare/folder/perm"
	// AnyshareDocLibQuotaScaleOpt 配额空间扩容
	AnyshareDocLibQuotaScaleOpt = "@anyshare/doclib/quota-scale"

	// InternalTextSplitOpt 文本分割
	InternalTextSplitOpt = "@internal/text/split"
	// InternalTextJoinOpt 文本合并
	InternalTextJoinOpt = "@internal/text/join"
	// InternalTextMatchOpt 文本匹配
	InternalTextMatchOpt = "@internal/text/match"
	// InternalReturnOpt 节点结束
	InternalReturnOpt = "@internal/return"

	// InternalDefineOpt 变量初始化
	InternalDefineOpt = "@internal/define"
	// InternalAssignOpt 变量赋值
	InternalAssignOpt = "@internal/assign"

	// InternalToolPy3Opt python3
	InternalToolPy3Opt = "@internal/tool/py3"

	// InternalTimeNow 获取当前时间
	InternalTimeNow = "@internal/time/now"

	// InternalTimeRelative 获取相对时间
	InternalTimeRelative = "@internal/time/relative"

	// WorkflowApproval 异步操作测试
	WorkflowApproval = "@workflow/approval"

	// IntelliinfoTranfer 智能数据转换
	IntelliinfoTranfer = "@intelliinfo/transfer"

	// AudioTransfer 音频转换
	AudioTransfer = "@audio/transfer"
	// DocInfoEntityExtract 文本实体信息提取
	DocInfoEntityExtract = "@docinfo/entity/extract"

	OpSandboxExecute = "@sandbox/execute"

	// Loop 循环
	Loop = "@control/flow/loop"

	// ControlFlowParallel 并行分支
	ControlFlowParallel = "@control/flow/parallel"

	// BranchOpt 分支
	BranchOpt = "@control/flow/branches"
	// BranchCmpStringEqOpt 字符串比较相等
	BranchCmpStringEqOpt = "@internal/cmp/string-eq"
	// BranchCmpStringNeqOpt 字符串比较不相等
	BranchCmpStringNeqOpt = "@internal/cmp/string-neq"
	// BranchCmpStringContainsOpt 字符串比较包含
	BranchCmpStringContainsOpt = "@internal/cmp/string-contains"
	// BranchCmpStringNotContainsOpt 字符串比较不包含
	BranchCmpStringNotContainsOpt = "@internal/cmp/string-not-contains"
	// BranchCmpStringStartWithOpt 字符串比较以某字符串开头
	BranchCmpStringStartWithOpt = "@internal/cmp/string-start-with"
	// BranchCmpStringEndWithOpt 字符串比较以某字符串结尾
	BranchCmpStringEndWithOpt = "@internal/cmp/string-end-with"
	// BranchCmpStringEmpty 字符串为空
	BranchCmpStringEmpty = "@internal/cmp/string-empty"
	// BranchCmpStringNotEmpty 字符串不为空
	BranchCmpStringNotEmpty = "@internal/cmp/string-not-empty"
	// BranchCmpStringMatch 字符串匹配正则表达式
	BranchCmpStringMatch = "@internal/cmp/string-match"
	// BranchCmpNumberEqOpt 数字比较相等
	BranchCmpNumberEqOpt = "@internal/cmp/number-eq"
	// BranchCmpNumberNeqOpt 数字比较不相等
	BranchCmpNumberNeqOpt = "@internal/cmp/number-neq"
	// BranchCmpNumberLtOpt 数字比较小于
	BranchCmpNumberLtOpt = "@internal/cmp/number-lt"
	// BranchCmpNumberLteOpt 数字比较小于等于
	BranchCmpNumberLteOpt = "@internal/cmp/number-lte"
	// BranchCmpNumberGtOpt 数字比较大于
	BranchCmpNumberGtOpt = "@internal/cmp/number-gt"
	// BranchCmpNumberGteOpt 数字比较大于等于
	BranchCmpNumberGteOpt = "@internal/cmp/number-gte"
	// BranchCmpdateEqOpt 日期比较等于
	BranchCmpdateEqOpt = "@internal/cmp/date-eq"
	// BranchCmpdateNeqOpt 日期比较不等于
	BranchCmpdateNeqOpt = "@internal/cmp/date-neq"
	// BranchCmpdateEarlierThanOpt 日期比较早于当前
	BranchCmpdateEarlierThanOpt = "@internal/cmp/date-earlier-than"
	// BranchCmpdateLaterThanOpt 日期比较晚于当前
	BranchCmpdateLaterThanOpt = "@internal/cmp/date-later-than"
	// BranchWorkflowEq 审核结果等于
	BranchWorkflowEq = "@workflow/cmp/approval-eq"
	// BranchWorkflowNeq 审核结果等于
	BranchWorkflowNeq = "@workflow/cmp/approval-neq"

	// DatabaseWriteOpt MySQL数据库写入操作
	DatabaseWriteOpt = "@internal/database/write"
	// SubFlowCallWorkflowOpt 工作流子流程调用操作
	SubFlowCallWorkflowOpt = "@subflow/call/workflow"
	// SubFlowCallWorkflowOpt 数据流子流程调用操作
	SubFlowCallDataflowOpt = "@subflow/call/dataflow"
)

const (
	// AnyshareManualTrigger 手动触发器schema路径
	AnyshareManualTrigger = "trigger/manualtrigger.json"
	// AnyshareEventTrigger 事件触发器schema路径
	AnyshareEventTrigger = "trigger/eventtrigger.json"
	// AnyshareCronTrigger 定时触发器schema路径
	AnyshareCronTrigger = "trigger/crontrigger.json"
	// AnyshareCopyOrMove 文件、文件夹复制移动操作schema路径
	AnyshareCopyOrMove = "anyshare/copyormove.json"
	// AnyshareRemove 文件、文件夹移除schema路径
	AnyshareRemove = "anyshare/remove.json"
	// AnyshareRename 文件、文件夹重命名schema路径
	AnyshareRename = "anyshare/rename.json"
	// AnyshareAddTags 文件、文件夹添加标签schenma路径
	AnyshareAddTags = "anyshare/addtags.json"
	// AnyshareGetPath 获取文件、文件夹操作schema路径
	AnyshareGetPath = "anyshare/getpath.json"
	// AnyshareMatchContent 匹配文件内容schema路径
	AnyshareMatchContent = "anyshare/matchcontent.json"
	// AnyshareSetCsfLevel 设置文件密级schema路径
	AnyshareSetCsfLevel = "anyshare/setcsflevel.json"
	// AnyshareSetTemplate 设置文件、文件夹编目schema路径
	AnyshareSetTemplate = "anyshare/settemplate.json"
	// AnyshareGetDocMeta 获取文档页数
	AnyshareGetDocMeta = "anyshare/getdocmeta.json"
	// AnyshareFloderCreate 创建文件夹schema路径
	AnyshareFloderCreate = "anyshare/createfolder.json"
	// InternalTextSplit 文本分割schema路径
	InternalTextSplit = "utils/textsplit.json"
	// InternalTextJoin 文本合并schema路径
	InternalTextJoin = "utils/textjoin.json"
	// InternalTextMatch 文本匹配schema路径
	InternalTextMatch = "utils/textmatch.json"
	// ChooseOneFolder 选择指定文件夹列举文件或文件夹
	ChooseOneFolder = "datasource/chooseonefolder.json"
	// ChooseMultiFolder 选择多个文件或文件夹
	ChooseMultiFolder = "datasource/choosemultifolder.json"
	// InternalToolPy python代码执行节点
	InternalToolPy = "tools/py.json"
	// FileOCR ocr识别
	FileOCR = "tools/ocr.json"
	// AudioTransferSchema 音频转文字schema路径
	AudioTransferSchema = "/tools/audiotransfer.json"
	// FilePermSet 文件权限申请schema路径
	FilePermSet = "anyshare/setperm.json"
	// Workflow 审核节点schema路径
	Workflow = "anyshare/workflow.json"
	// SecurityPolicyTriggerSchema 安全策略触发器节点schema路径
	SecurityPolicyTriggerSchema = "trigger/security-policy-trigger.json"
	// InternalReturnSchema 提前结束流程节点schema路径
	InternalReturnSchema = "utils/return.json"
	// InternalTimeRelativeSchema 获取相对时间节点schema路径
	InternalTimeRelativeSchema = "utils/timerelative.json"
	// AnyshareFileCreate 新建文件schema路径
	AnyshareFileCreate = "anyshare/createfile.json"
	// AnyshareFileEdit 更新文件schema路径
	AnyshareFileEdit = "anyshare/editfile.json"
	// AnyshareExcelFileEdit 编辑excel文件schema路径
	AnyshareExcelFileEdit = "anyshare/editexcel.json"
	// AnyshareDocxFileEdit 编辑docx文件schema路径
	AnyshareDocxFileEdit = "anyshare/editdocx.json"
	// DocInfoEntityExtractSchema 自定义模型提取schema路径
	DocInfoEntityExtractSchema = "tools/extract.json"
	// AnyshareDocLibQuotaScaleOptSchema 配额空间扩容schema路径
	AnyshareDocLibQuotaScaleOptSchema = "anyshare/quatoscale.json"
	// AnyshareDocLibQuotaScaleOptSchema 配额空间扩容schema路径
	ComboOperatorsOptSchema = "dataflow/operators.json"
	// LoopSchema 循环
	LoopSchema = "utils/loop.json"
)

const (
	OpAnyShareDocPrefix            = "@anyshare/doc/"
	OpAnyShareDocRename            = OpAnyShareDocPrefix + "rename"
	OpAnyShareDocRemove            = OpAnyShareDocPrefix + "remove"
	OpAnyShareDocAddTag            = OpAnyShareDocPrefix + "addtag"
	OpAnyShareDocSetCsfLevel       = OpAnyShareDocPrefix + "setcsflevel"
	OpAnyShareDocSetTemplate       = OpAnyShareDocPrefix + "settemplate"
	OpAnyShareDocSetPerm           = OpAnyShareDocPrefix + "perm"
	OpAnyShareDocGetPath           = OpAnyShareDocPrefix + "getpath"
	OpAnyShareDocSetSpaceQuota     = OpAnyShareDocPrefix + "setspacequota"
	OpAnyShareDocSetAllowSuffixDoc = OpAnyShareDocPrefix + "setallowsuffixdoc"
)

const (
	OpCognitiveAssistantCustomPrompt  = "@cognitive-assistant/custom-prompt"
	OpCognitiveAssistantDocSummarize  = "@cognitive-assistant/doc-summarize"
	OpCognitiveAssistantMeetSummarize = "@cognitive-assistant/meet-summarize"
)

const (
	OpAnyShareSelectedFileTrigger   = "@trigger/selected-file"
	OpAnyShareSelectedFolderTrigger = "@trigger/selected-folder"
	OpAnyShareFileGetByName         = "@anyshare/file/get-file-by-name"
)

const (
	OpAnyShareFileRelevance   = "@anyshare/file/relevance"
	OpAnyShareFolderRelevance = "@anyshare/folder/relevance"
)

const (
	OpJsonGet      = "@internal/json/get"
	OpJsonSet      = "@internal/json/set"
	OpJsonTemplate = "@internal/json/template"
	OpJsonParse    = "@internal/json/parse"
)

const (
	OpArrayFilter = "@internal/array/filter"
)

const (
	OpContentFullText = "@content/fulltext"
	OpContentAbstract = "@content/abstract"
	OpContentEntity   = "@content/entity"
)

const (
	OpLLMChatCompletion = "@llm/chat/completion"
	OpLLmEmbedding      = "@llm/embedding"
	OpLLMReranker       = "@llm/reranker"
)

const (
	OpAnyDataCallAgent = "@anydata/call-agent"
)

const (
	OpEcoconfigReindex = "@ecoconfig/reindex"
)

const (
	ComboOperatorPrefix = "@operator/"
	// CustomOperatorPrefix 自定义算子前缀
	CustomOperatorPrefix = "@custom/"

	TriggerOperatorPrefix = "@trigger/operator/"
)

const (
	OpOpenSearchBulkUpsert = "@opensearch/bulk-upsert"
)

const (
	OpContentPipelineFullText         = "@contentpipeline/full_text"
	OpContentPipelineDocFormatConvert = "@contentpipeline/doc_format_convert"
	OpContentFileParse                = "@content/file_parse"
	OpOCRNew                          = "@anyshare/ocr/new"
)

var DataSourceActionMap = map[string]string{
	AnyshareDataListFiles:      ChooseOneFolder,
	AnyshareDataListFolders:    ChooseOneFolder,
	AnyshareDataSpecifyFiles:   ChooseMultiFolder,
	AnyshareDataSpecifyFolders: ChooseMultiFolder,
	AnyshareDataDepartment:     "", // TODO 补全json schema
	AnyshareDataUser:           "",
	AnyshareDataTagTree:        "",
}

var CmpActionMap = map[string]string{
	BranchCmpStringEqOpt:          "",
	BranchCmpStringNeqOpt:         "",
	BranchCmpStringContainsOpt:    "",
	BranchCmpStringNotContainsOpt: "",
	BranchCmpStringStartWithOpt:   "",
	BranchCmpStringEndWithOpt:     "",
	BranchCmpStringEmpty:          "",
	BranchCmpStringNotEmpty:       "",
	BranchCmpStringMatch:          "",
	BranchCmpNumberEqOpt:          "",
	BranchCmpNumberNeqOpt:         "",
	BranchCmpNumberLtOpt:          "",
	BranchCmpNumberLteOpt:         "",
	BranchCmpNumberGtOpt:          "",
	BranchCmpNumberGteOpt:         "",
	BranchCmpdateEqOpt:            "",
	BranchCmpdateNeqOpt:           "",
	BranchCmpdateEarlierThanOpt:   "",
	BranchCmpdateLaterThanOpt:     "",
	BranchWorkflowEq:              "",
	BranchWorkflowNeq:             "",
}

// ActionMap action map
var ActionMap = map[string]string{
	// trigger opt
	CronTrigger:                   AnyshareCronTrigger,
	CronWeekTrigger:               AnyshareCronTrigger,
	CronMonthTrigger:              AnyshareCronTrigger,
	CronCustomTrigger:             AnyshareCronTrigger,
	WebhookTrigger:                AnyshareEventTrigger,
	MannualTriggerOpt:             AnyshareManualTrigger,
	FormTrigger:                   "",
	AnyshareFileUploadTrigger:     AnyshareEventTrigger,
	AnyshareFileMoveTrigger:       AnyshareEventTrigger,
	AnyshareFileCopyTrigger:       AnyshareEventTrigger,
	AnyshareFileRemoveTrigger:     AnyshareEventTrigger,
	AnyshareFolderCreateTrigger:   AnyshareEventTrigger,
	AnyshareFolderMoveTrigger:     AnyshareEventTrigger,
	AnyshareFolderCopyTrigger:     AnyshareEventTrigger,
	AnyshareFolderRemoveTrigger:   AnyshareEventTrigger,
	AnyshareFileReversionTrigger:  AnyshareEventTrigger,
	AnyshareFileRenameTrigger:     AnyshareEventTrigger,
	AnyshareFileDeleteTrigger:     AnyshareEventTrigger,
	AnyshareFileRestoreTrigger:    AnyshareEventTrigger,
	AnyshareUserDeleteTrigger:     "",
	AnyshareUserFreezeTrigger:     "",
	AnyshareUserCreateTrigger:     "",
	AnyshareOrgNameModifyTrigger:  "",
	AnyshareDeptDeleteTrigger:     "",
	AnyshareDeptCreateTrigger:     "",
	AnyshareUserMovedTrigger:      "",
	AnyshareDeptMovedTrigger:      "",
	AnyshareUserAddDeptTrigger:    "",
	AnyshareUserRemoveDeptTrigger: "",
	AnyshareUserChangeTrigger:     "",
	AnyshareTagTreeAddedTrigger:   "",
	AnyshareTagTreeCreateTrigger:  "",
	AnyshareTagTreeEditedTrigger:  "",
	AnyshareTagTreeDeletedTrigger: "",
	SecurityPolicyTrigger:         SecurityPolicyTriggerSchema,
	DataflowDocTrigger:            "",
	DataflowUserTrigger:           "",
	DataflowDeptTrigger:           "",
	DataflowTagTrigger:            "",
	TriggerOperatorPrefix:         "",
	MDLDataViewTrigger:            "trigger/dataviewtrigger.json",

	// file operator
	AnyshareFileCopyOpt:         AnyshareCopyOrMove,
	AnyshareFileMoveOpt:         AnyshareCopyOrMove,
	AnyshareFileRemoveOpt:       AnyshareRemove,
	AnyshareFileRenameOpt:       AnyshareRename,
	AnyshareFileAddTagsOpt:      AnyshareAddTags,
	AnyshareFileGetPathOpt:      AnyshareGetPath,
	AnyshareFileMatchContentOpt: AnyshareMatchContent,
	AnyshareFileSetCsfLevelOpt:  AnyshareSetCsfLevel,
	AnyshareFileSetTemplate:     AnyshareSetTemplate,
	AnyshareFileGetPage:         AnyshareGetDocMeta,
	AnyshareFileOCROpt:          FileOCR,
	AnyshareEleInvoiceOpt:       FileOCR,
	AnyshareIDCardOpt:           FileOCR,
	AnyshareFileSetPermOpt:      FilePermSet,
	AnyshareFileCreateOpt:       AnyshareFileCreate,
	AnyshareFileEditOpt:         AnyshareFileEdit,
	AnyshareExcelFileEditOpt:    AnyshareExcelFileEdit,
	AnyshareDocxFileEditOpt:     AnyshareDocxFileEdit,

	// folder operator
	AnyshareFolderCopyOpt:       AnyshareCopyOrMove,
	AnyshareFolderMoveOpt:       AnyshareCopyOrMove,
	AnyshareFolderRemoveOpt:     AnyshareRemove,
	AnyshareFolderRenameOpt:     AnyshareRename,
	AnyshareFloderCreateOpt:     AnyshareFloderCreate,
	AnyshareFolderAddTagsOpt:    AnyshareAddTags,
	AnyshareFolderGetPathOpt:    AnyshareGetPath,
	AnyshareFolderSetTemplate:   AnyshareSetTemplate,
	AnyshareFolderSetPermOpt:    FilePermSet,
	AnyshareDocLibQuotaScaleOpt: AnyshareDocLibQuotaScaleOptSchema,

	// utils
	InternalTextSplitOpt: InternalTextSplit,
	InternalTextJoinOpt:  InternalTextJoin,
	InternalTextMatchOpt: InternalTextMatch,
	InternalReturnOpt:    InternalReturnSchema,
	InternalTimeNow:      "",
	InternalTimeRelative: InternalTimeRelativeSchema,

	// json
	OpJsonGet:      "",
	OpJsonSet:      "",
	OpJsonTemplate: "",
	OpJsonParse:    "",

	// dataSource
	AnyshareDataListFiles:      ChooseOneFolder,
	AnyshareDataListFolders:    ChooseOneFolder,
	AnyshareDataSpecifyFiles:   ChooseMultiFolder,
	AnyshareDataSpecifyFolders: ChooseMultiFolder,
	AnyshareDataDepartment:     "", // TODO 补全json schema
	AnyshareDataUser:           "",
	AnyshareDataTagTree:        "",

	// tools
	InternalToolPy3Opt:   InternalToolPy,
	WorkflowApproval:     Workflow,
	IntelliinfoTranfer:   "", // TODO 补全json schema
	AudioTransfer:        AudioTransferSchema,
	DocInfoEntityExtract: DocInfoEntityExtractSchema,
	OpSandboxExecute:     "tools/sandboxexecute.json",

	// flow
	Loop:                          LoopSchema,
	ControlFlowParallel:           "", // 并行分支节点不需要schema验证
	BranchOpt:                     "",
	BranchCmpStringEqOpt:          "",
	BranchCmpStringNeqOpt:         "",
	BranchCmpStringContainsOpt:    "",
	BranchCmpStringNotContainsOpt: "",
	BranchCmpStringStartWithOpt:   "",
	BranchCmpStringEndWithOpt:     "",
	BranchCmpStringEmpty:          "",
	BranchCmpStringNotEmpty:       "",
	BranchCmpStringMatch:          "",
	BranchCmpNumberEqOpt:          "",
	BranchCmpNumberNeqOpt:         "",
	BranchCmpNumberLtOpt:          "",
	BranchCmpNumberLteOpt:         "",
	BranchCmpNumberGtOpt:          "",
	BranchCmpNumberGteOpt:         "",
	BranchCmpdateEqOpt:            "",
	BranchCmpdateNeqOpt:           "",
	BranchCmpdateEarlierThanOpt:   "",
	BranchCmpdateLaterThanOpt:     "",
	BranchWorkflowEq:              "",
	BranchWorkflowNeq:             "",

	// 安全策略

	OpAnyShareDocAddTag:            "",
	OpAnyShareDocRemove:            "",
	OpAnyShareDocRename:            "",
	OpAnyShareDocSetCsfLevel:       "",
	OpAnyShareDocSetPerm:           "",
	OpAnyShareDocSetTemplate:       "",
	OpAnyShareDocSetAllowSuffixDoc: "",
	OpAnyShareDocSetSpaceQuota:     "",

	OpCognitiveAssistantCustomPrompt:  "cognitive-assistant/custom-prompt.json",
	OpCognitiveAssistantDocSummarize:  "cognitive-assistant/custom-prompt.json",
	OpCognitiveAssistantMeetSummarize: "cognitive-assistant/custom-prompt.json",
	OpAnyShareSelectedFileTrigger:     "",
	OpAnyShareSelectedFolderTrigger:   "",
	OpAnyShareFileRelevance:           "",
	OpAnyShareFolderRelevance:         "",
	OpAnyShareFileGetByName:           "",

	OpContentAbstract: "",
	OpContentFullText: "",
	OpContentEntity:   "",

	OpLLMChatCompletion: "llm/chat-completion.json",
	OpLLMReranker:       "llm/reranker.json",
	OpLLmEmbedding:      "llm/embedding.json",

	OpAnyDataCallAgent: "",
	OpEcoconfigReindex: "",

	ComboOperatorPrefix:    ComboOperatorsOptSchema,
	InternalAssignOpt:      "utils/assign.json",
	InternalDefineOpt:      "utils/define.json",
	OpOpenSearchBulkUpsert: "opensearch/bulkupsert.json",
	DatabaseWriteOpt:       "database/write.json", // MySQL数据库写入操作

	OpContentPipelineFullText:         "tools/fulltext.json",
	OpContentPipelineDocFormatConvert: "tools/docformatconvert.json",
	OpOCRNew:                          "tools/ocrnew.json",
	OpContentFileParse:                "tools/fileparse.json",
	SubFlowCallWorkflowOpt:            "tools/workflowcall.json",
	SubFlowCallDataflowOpt:            "tools/dataflowcall.json",
}
