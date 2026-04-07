package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	normalizeutil "github.com/kweaver-ai/adp/autoflow/flow-automation/utils/normalize"
)

const conflictCode = 409

// AuditMsg 审核消息体
type AuditMsg struct {
	Process  AuditProcessMsg        `json:"process"`  // 审核流程信息
	Data     map[string]interface{} `json:"data"`     // 需审核的数据
	Workflow AuditWorkflowMsg       `json:"workflow"` // workflow需要做显示用的信息
}

// AuditProcessMsg 审核信息process
type AuditProcessMsg struct {
	AuditType       string `json:"audit_type"`        // 审核申请类型
	ApplyID         string `json:"apply_id"`          // 审核申请id
	ConflictApplyID string `json:"conflict_apply_id"` // 上一个申请id
	UserID          string `json:"user_id"`           // 申请人id
	UserName        string `json:"user_name"`         // 申请人显示名
	ProcDefKey      string `json:"proc_def_key"`      // 审核流程key
}

// AuditWorkflowMsg 审核信息workflow
type AuditWorkflowMsg struct {
	TopCsf       int64             `json:"top_csf"`       // 审核内容和申请人的最高密级
	MsgForEmail  []string          `json:"msg_for_email"` // 审核邮件所需展示的key及顺序
	MsgForLog    []string          `json:"msg_for_log"`   // 审核日志所需展示的key及顺序
	Content      map[string]string `json:"content"`       // 邮件、日志所需展示的内容
	AbstractInfo map[string]string `json:"abstract_info"` // 审核界面摘要显示的图标和文字
}

// AuditWorkflowCancelMsg 审核取消消息
type AuditWorkflowCancelMsg struct {
	ApplyIDs []string          `json:"apply_ids"`
	Cause    map[string]string `json:"cause"`
}

// WorkflowAsyncTask python代码执行
type WorkflowAsyncTask struct {
	WorkflowID string                   `json:"workflow"`
	UserID     string                   `json:"user_id"`
	UserName   string                   `json:"user_name"`
	Contents   []map[string]interface{} `json:"contents"`
	Title      string                   `json:"title"`
}

type ProcApprovalMsg struct {
	PID     string `json:"pid"`
	ApplyID string `json:"apply_id"`
	GroupID string `json:"group_id"`
}

// Name 操作名称
func (a *WorkflowAsyncTask) Name() string {
	return common.WorkflowApproval
}

type PermValue struct {
	Allow []string `json:"allow"`
	Deny  []string `json:"deny"`
}

// MemberInfo user或department 信息
type MemberInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// RunBefore 操作方法
func (a *WorkflowAsyncTask) RunBefore(_ entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	return entity.TaskInstanceStatusBlocked, nil
}

// Run 操作方法
func (a *WorkflowAsyncTask) Run(ctx entity.ExecuteContext, params interface{}, _ *entity.Token) (interface{}, error) {
	var err error
	taskIns := ctx.GetTaskInstance()
	if taskIns == nil {
		return nil, fmt.Errorf("get taskinstance failed")
	}
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	userMgntAdapters := drivenadapters.NewUserManagement()
	docshareAdapters := drivenadapters.NewDocShare()
	operatorID, _ := ctx.GetVar("operator_id")
	operatorName, _ := ctx.GetVar("operator_name")

	topCsfLevel := 0
	input := params.(*WorkflowAsyncTask)
	input.UserID = operatorID.(string)
	input.UserName = operatorName.(string)
	content := input.Contents

	auditMsgData := make(map[string]interface{})
	auditMsgData["content"] = content

	docids := make([]string, 0)

	auditType := common.AuditType

	dagType, exists := ctx.GetVar("dag_type")

	abstractInfo := map[string]string{
		"text": input.Title,
	}

	msgForEmail := []string{}
	msgForLog := []string{}
	msgContent := map[string]string{}

	var permMap = map[string]string{
		"display":  common.GetLocale("display"),
		"preview":  common.GetLocale("preview"),
		"cache":    common.GetLocale("cache"),
		"download": common.GetLocale("download"),
		"create":   common.GetLocale("create"),
		"modify":   common.GetLocale("modify"),
		"delete":   common.GetLocale("delete"),
	}

	var showMsgForEmail = true
	var isSecurityPolicy = false
	var isSecurityPolicyForFolder = false

	log := traceLog.WithContext(ctx.Context())
	efast := ctx.NewASDoc()

	source, sourceExist := ctx.ShareData().Get("__source")
	if sourceExist {
		auditMsgData["source"] = source
	}

	kcmc := drivenadapters.NewKcmc()

	if exists && dagType == common.DagTypeSecurityPolicy {

		if !sourceExist {
			return nil, errors.New("missing source")
		}

		isSecurityPolicy = true
		policyType, exists := ctx.GetVar("policy_type")

		if !exists {
			return nil, errors.New("missing policy_type")
		}

		auditType = common.SecurityPolicyAuditPrefix + policyType.(string)

		// 新建策略和修改文件夹属性策略邮件不显示内容
		if policyType == "upload" || policyType == "modify_folder_property" {
			showMsgForEmail = false
		}

		var sourceDocID string
		var article *drivenadapters.Article

		if m, ok := source.(map[string]interface{}); ok {
			isSecurityPolicyForFolder = m["type"] == "folder"
			if isSecurityPolicyForFolder {
				sourceDocID = m["id"].(string)
				docAttrs, err := ctx.NewASDoc().GetDocMsg(ctx.Context(), sourceDocID)
				if err != nil {
					log.Warnf("[async.run] GetDocMsg err, detail: %s", err.Error())
					return nil, errors.New("get doc info failed")
				}

				if docAttrs.DocLibType == common.CustomDocLib && docAttrs.CustomType != nil {
					result, err := kcmc.IsArticleProxyDocLibSubtype(ctx.Context(), docAttrs.CustomType.ID)

					if err != nil {
						return nil, err
					}

					if result {
						article, err = kcmc.GetArticleByProxyDirID(ctx.Context(), sourceDocID)

						if err != nil {
							log.Warnf("[async.run] GetArticleByProxyDirID err, detail: %s", err.Error())
							return nil, errors.New("get article failsed")
						}
					}
				}
			}

			if m["type"] == "file" || m["type"] == "folder" {
				if article != nil {

					title := article.Title

					switch article.Source {
					case drivenadapters.ArticleSourceHtml:
						abstractInfo["icon"] = "article"
					case drivenadapters.ArticleSourceAs:
						abstractInfo["icon"] = "file"
						title = article.Title + article.Suffix
					case drivenadapters.ArticleSourceAutoSheet:
						abstractInfo["icon"] = "autosheet"
					case drivenadapters.ArticleSourceGroup:
						abstractInfo["icon"] = "group"
					default:
						abstractInfo["icon"] = "file"
					}

					abstractInfo["text"] = title

					docids = append(docids, sourceDocID)

					auditMsgData["article"] = map[string]interface{}{
						"title":      title,
						"article_id": strconv.FormatUint(article.ArticleID, 10),
					}
					msgContent["doc_name"] = fmt.Sprintf("%s:%v", common.GetLocale("doc_name"), title)
					msgContent["doc_id"] = fmt.Sprintf("%s:%v", common.GetLocale("doc_id"), article.ArticleID)
					msgForLog = append(msgForLog, "doc_name", "doc_id")
				} else {
					abstractInfo["icon"] = m["type"].(string)
					n, ok := m["name"].(string)
					if ok {
						abstractInfo["text"] = n
					}

					maxCsfLevel, err := efast.GetMaxCsfLevel(ctx.Context(), m["id"].(string))

					if err != nil {
						log.Warnf("[async.run] GetMaxCsfLevel err, detail: %s", err.Error())
					} else {
						if maxCsfLevel > topCsfLevel {
							topCsfLevel = maxCsfLevel
						}

						docids = append(docids, m["id"].(string))

						msgContent["doc_name"] = fmt.Sprintf("%s:%v", common.GetLocale("doc_name"), m["name"])
						msgContent["doc_id"] = fmt.Sprintf("%s:%v", common.GetLocale("doc_id"), utils.GetDocCurID(m["id"].(string)))
						msgContent["doc_path"] = fmt.Sprintf("%s:%v", common.GetLocale("doc_path"), m["path"])
						msgForLog = append(msgForLog, "doc_name", "doc_id", "doc_path")
					}
				}
			}
		}
	}

	for index, item := range content {

		key := fmt.Sprintf("content_%v", index)

		if item["type"] == "asFile" || item["type"] == "asFolder" || item["type"] == "asDoc" {
			docID := item["value"].(string)
			attr, err := efast.GetDocMsg(ctx.Context(), docID)
			if err != nil {
				log.Warnf("[async.run] GetDocMsg err, detail: %s", err.Error())
				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
				msgForEmail = append(msgForEmail, key)
				msgForLog = append(msgForLog, key)
				continue
			}

			if attr.Size == -1 {
				maxCsfLevel, err := efast.GetMaxCsfLevel(ctx.Context(), docID)
				if err != nil {
					log.Warnf("[async.run] GetMaxCsfLevel err, detail: %s", err.Error())
					continue
				}
				if maxCsfLevel > topCsfLevel {
					topCsfLevel = maxCsfLevel
				}
			}

			docids = append(docids, docID)

			if attr.CsfLevel > float64(topCsfLevel) {
				topCsfLevel = int(attr.CsfLevel)
			}

			isKcArticle := false

			if attr.DocLibType == common.CustomDocLib && attr.CustomType != nil {
				isKcArticle, err = kcmc.IsArticleProxyDocLibSubtype(ctx.Context(), attr.CustomType.ID)

				if err != nil {
					log.Warnf("[async.run] GetArticleByProxyDirID err, detail: %s", err.Error())
					msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
					msgForEmail = append(msgForEmail, key)
					msgForLog = append(msgForLog, key)
					continue
				}
			}

			if isKcArticle {
				article, err := kcmc.GetArticleByProxyDirID(ctx.Context(), attr.DocID)

				if err != nil {
					log.Warnf("[async.run] GetArticleByProxyDirID err, detail: %s", err.Error())
					msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
					msgForEmail = append(msgForEmail, key)
					msgForLog = append(msgForLog, key)
					continue
				}

				title := article.Title

				if article.Source == drivenadapters.ArticleSourceAs {
					title = article.Title + article.Suffix
				}

				item["name"] = title
				item["value"] = strconv.FormatUint(article.ArticleID, 10)
				item["type"] = "article"
				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], title)
				msgForEmail = append(msgForEmail, key)
				msgForLog = append(msgForLog, key)
			} else {
				item["name"] = attr.Name
				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], attr.Name)
				msgForEmail = append(msgForEmail, key)
				msgForLog = append(msgForLog, key)
			}

		} else if item["type"] == "multipleFiles" {
			var names, docIDs []string
			_docIDs := item["value"].(string)
			err = json.Unmarshal([]byte(_docIDs), &docIDs)
			if err != nil {
				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
				msgForEmail = append(msgForEmail, key)
				msgForLog = append(msgForLog, key)
				continue
			}

			for _, docID := range docIDs {
				attr, err := efast.GetDocMsg(ctx.Context(), docID)
				if err != nil {
					log.Warnf("[async.run] GetDocMsg err, detail: %s", err.Error())
					continue
				}

				if attr.Size == -1 {
					maxCsfLevel, err := efast.GetMaxCsfLevel(ctx.Context(), docID)
					if err != nil {
						log.Warnf("[async.run] GetMaxCsfLevel err, detail: %s", err.Error())
						continue
					}
					if maxCsfLevel > topCsfLevel {
						topCsfLevel = maxCsfLevel
					}
				}

				docids = append(docids, docID)
				names = append(names, attr.Name)
				if attr.CsfLevel > float64(topCsfLevel) {
					topCsfLevel = int(attr.CsfLevel)
				}
			}

			if len(names) == 0 {
				names = append(names, docIDs...)
			}
			msgContent[key] = fmt.Sprintf("%v:%v", item["title"], strings.Join(names, ","))
			msgForEmail = append(msgForEmail, key)
			msgForLog = append(msgForLog, key)

		} else if item["type"] == "asTags" {

			if value, ok := item["value"].(string); ok {
				var tags []string
				err := json.Unmarshal([]byte(value), &tags)

				if err != nil {
					msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
				} else {
					msgContent[key] = fmt.Sprintf("%v:%v", item["title"], strings.Join(tags, ", "))
				}
			} else {
				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
			}

			msgForLog = append(msgForLog, key)
		} else if item["type"] == "asLevel" {

			var csflevel int

			if val, ok := item["value"].(int); ok {
				csflevel = val
			} else if val, ok := item["value"].(CsfLevelInfo); ok {
				csflevel = val.CsfLevel
			} else if val, ok := item["value"].(string); ok {
				if val != "" {
					csflevel, err = strconv.Atoi(val)
					if err != nil {
						var obj CsfLevelInfo
						err = json.Unmarshal([]byte(val), &obj)
						if err != nil {
							return nil, err
						}
						csflevel = obj.CsfLevel
					}
				}
			} else if val, ok := item["value"].(map[string]interface{}); ok {
				str, err := jsoniter.Marshal(val)
				if err != nil {
					return nil, err
				}
				var obj CsfLevelInfo
				err = json.Unmarshal(str, &obj)
				if err != nil {
					return nil, err
				}
				csflevel = obj.CsfLevel
			}

			if csflevel > topCsfLevel {
				topCsfLevel = csflevel
			}

			msgContent[key] = fmt.Sprintf("%v:%v", item["title"], "")
			msgForLog = append(msgForLog, key)
		} else if item["type"] == "asMetadata" {
			msgContent[key] = fmt.Sprintf("%v:%v", item["title"], "")
			msgForLog = append(msgForLog, key)
		} else if item["type"] == "asPerm" {

			var perm PermValue
			var bytes []byte

			if value, ok := item["value"].(string); ok {
				bytes = []byte(value)
			} else {
				bytes, _ = json.Marshal(item["value"])
			}
			err := json.Unmarshal(bytes, &perm)

			if err != nil {
				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
			} else {
				var permStr string
				if len(perm.Deny) == 7 {
					permStr = common.GetLocale("deny_all")
				} else {
					var allowSegments []string
					for _, item := range perm.Allow {
						if s, ok := permMap[item]; ok {
							allowSegments = append(allowSegments, s)
						} else {
							allowSegments = append(allowSegments, item)
						}
					}

					permStr = strings.Join(allowSegments, ", ")

					if len(perm.Deny) > 0 {
						var denySegments []string
						for _, item := range perm.Deny {
							if s, ok := permMap[item]; ok {
								denySegments = append(denySegments, s)
							} else {
								denySegments = append(denySegments, item)
							}
						}

						permStr = fmt.Sprintf("%s (%s %s)", permStr, common.GetLocale("deny"), strings.Join(denySegments, ", "))
					}
				}

				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], permStr)
			}

			msgForEmail = append(msgForEmail, key)
			msgForLog = append(msgForLog, key)

		} else if item["type"] == "asAccessorPerms" {
			// 安全策略日志不显示此项
			if !isSecurityPolicy {
				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], "")
				msgForLog = append(msgForLog, key)
			}
		} else if item["type"] == "datetime" {

			var timestamp int64 = -1

			if t, ok := item["value"].(int64); ok {
				timestamp = t
			} else if t, ok := item["value"].(string); ok {
				if t != "" {
					val, err := convertTimeStringToMsTimestamp(t)
					if err == nil {
						timestamp = val
					}
				}
			}

			item["value"] = timestamp

			if timestamp == -1 {
				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], common.GetLocale("forever"))
			} else {
				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], time.Unix(timestamp/1e6, 0).Format("2006-01-02 15:04:05"))
			}

			msgForEmail = append(msgForEmail, key)
			msgForLog = append(msgForLog, key)
		} else if item["type"] == "asSpaceQuota" {
			// 非安全策略或是安全策略对象为文件夹显示此项
			if !isSecurityPolicy || isSecurityPolicyForFolder {
				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
				msgForEmail = append(msgForEmail, key)
				msgForLog = append(msgForLog, key)
			}
		} else if item["type"] == "asAllowSuffixDoc" {
			// 非安全策略或是安全策略对象为文件夹显示此项
			if !isSecurityPolicy || isSecurityPolicyForFolder {
				msgContent[key] = fmt.Sprintf("%v:%v", item["title"], "")
				msgForLog = append(msgForLog, key)
			}
		} else if item["type"] == "asDepartments" || item["type"] == "asUsers" {
			var members []MemberInfo
			var names []string
			_ = json.Unmarshal([]byte(fmt.Sprintf("%v", item["value"])), &members)
			for _, member := range members {
				names = append(names, member.Name)
			}

			msgContent[key] = fmt.Sprintf("%v:%v", item["title"], strings.TrimSpace(strings.Join(names, ", ")))
			msgForEmail = append(msgForEmail, key)
			msgForLog = append(msgForLog, key)
		} else {
			msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
			msgForEmail = append(msgForEmail, key)
			msgForLog = append(msgForLog, key)
		}
	}

	fmt.Printf("msgContent: %v\n", msgContent)

	applyID := taskIns.ID

	gid, err := userMgntAdapters.CreateInternalGroup()
	if err != nil {
		log.Warnf("[asyncDocAudit] CreateInternalGroup failed, err: %s", err.Error())
		return nil, err
	}

	res := map[string]interface{}{
		"title":    input.Title,
		"group_id": gid,
	}

	for k, v := range input.Contents {
		res[fmt.Sprintf("contents_%v", k)] = v
	}

	executeMethods := ctx.NewExecuteMethods()

	go func() {
		time.Sleep(100 * time.Microsecond)

		var workflowMsg = AuditWorkflowMsg{
			TopCsf:       int64(topCsfLevel),
			MsgForEmail:  msgForEmail,
			MsgForLog:    msgForLog,
			Content:      msgContent,
			AbstractInfo: abstractInfo,
		}
		dag, err := executeMethods.GetDag(ctx.Context(), taskIns.RelatedDagInstance.DagID, taskIns.RelatedDagInstance.VersionID)
		if err != nil {
			log.Warnf("[asyncDocAudit] GetDag failed, detail: %s", err.Error())
			auditMsgData["automation_flow_name"] = input.Title
		} else {
			auditMsgData["automation_flow_name"] = dag.Name
		}

		auditMsgData["dagid"] = taskIns.RelatedDagInstance.DagID

		if !showMsgForEmail {
			workflowMsg.MsgForEmail = []string{}
		}

		auditMsg := AuditMsg{
			Process: AuditProcessMsg{
				AuditType:       auditType,
				ApplyID:         applyID,
				ConflictApplyID: "",
				UserID:          input.UserID,
				UserName:        input.UserName,
				ProcDefKey:      input.WorkflowID,
			},
			Data:     auditMsgData,
			Workflow: workflowMsg,
		}

		msg, _ := jsoniter.Marshal(auditMsg)
		err = executeMethods.Publish(common.TopicWorkflowApply, msg)

		if err != nil {
			log.Warnf("[asyncDocAudit] Publish failed, auditMsg: %v, err: %v", auditMsg.Process.ApplyID, err.Error())
		}
		log.Infof("[asyncDocAudit] asyncDocAudit-auditMsg: %v", auditMsg.Process.ApplyID)
	}()

	if len(docids) == 0 {
		ctx.Trace(ctx.Context(), "run end")
		return res, nil
	}

	configs := make([]map[string]interface{}, 0)
	configs = append(configs, map[string]interface{}{
		"accessor": map[string]string{
			"id":   gid,
			"type": "internal_group",
		},
		"allow": []string{
			"preview",
			"download",
			"display",
		},
	})

	for _, docid := range docids {
		code, cerr := docshareAdapters.SetDocPerm(ctx.Context(), docid, configs)
		if cerr != nil {
			if code == conflictCode {
				continue
			}
			log.Warnf("[HandleAuditorsMacth] SetDocPerm failed, err: %s", err.Error())
			return nil, err
		}
	}

	// 发送流程进入审核消息
	procApprovalMsg := ProcApprovalMsg{
		PID:     taskIns.DagInsID,
		ApplyID: applyID,
		GroupID: gid,
	}

	procApprovalMsgBytes, _ := jsoniter.Marshal(procApprovalMsg)
	err = executeMethods.Publish(common.TopicSecurityPolicyProcApproval, procApprovalMsgBytes)

	if err != nil {
		log.Warnf("[asyncDocAudit] Publish failed, procApprovalMsg: %v, err: %v", procApprovalMsg.PID, err.Error())
	}

	ctx.Trace(ctx.Context(), "run end")
	return res, nil
}

func (a *WorkflowAsyncTask) Cancel(taskIns entity.TaskInstance, executeMethods entity.ExecuteMethods) error {
	applyId := taskIns.ID

	cancelMsg := AuditWorkflowCancelMsg{
		ApplyIDs: []string{applyId},
		// TODO 添加审核取消原因
		Cause: map[string]string{
			"zh-cn": "撤销申请",
			"zh-tw": "撤銷申請",
			"en-us": "Withdraw Application",
		},
	}
	msg, _ := jsoniter.Marshal(cancelMsg)

	err := executeMethods.Publish(common.TopicWorkflowCancel, msg)

	log.Infof("[asyncDocAudit] cancel applyId: %v", applyId)

	taskResult, ok := normalizeutil.AsMap(taskIns.Results)

	if ok {
		groupID, ok := taskResult["group_id"].(string)
		if ok && groupID != "" {
			usermgnt := drivenadapters.NewUserManagement()
			err := usermgnt.DeleteInternalGroup([]string{groupID})
			if err != nil {
				log.Warnf("[asyncDocAudit] DeleteInternalGroup, groupID: %s, err: %s", groupID, err.Error())
			}
		}
	}

	return err
}

// ParameterNew 初始化参数
func (a *WorkflowAsyncTask) ParameterNew() interface{} {
	return &WorkflowAsyncTask{}
}
