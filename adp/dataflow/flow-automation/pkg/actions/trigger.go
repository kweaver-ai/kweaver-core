package actions

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/store"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

func patchDagInstanceSource(ctx entity.ExecuteContext, data map[string]interface{}, triggerName string) {
	if len(data) == 0 {
		return
	}

	taskIns := ctx.GetTaskInstance()
	if taskIns == nil {
		return
	}

	b, err := json.Marshal(data)

	if err != nil {
		log.Warnf("[%s] Marshal source err %s", triggerName, err.Error())
		return
	}

	if err := ctx.NewExecuteMethods().PatchDagIns(ctx.Context(), &entity.DagInstance{
		BaseInfo: taskIns.RelatedDagInstance.BaseInfo,
		Source:   string(b),
	}); err != nil {
		log.Warnf("[%s] PatchDagIns err %s", triggerName, err.Error())
	}
}

// CronTrigger cron trigger
type CronTrigger struct {
}

// CronTriggerParam cron trigger param
type CronTriggerParam struct {
	Cron string `json:"cron"`
}

// Name 操作名称
func (a *CronTrigger) Name() string {
	return common.CronTrigger
}

// Run 操作方法
func (a *CronTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	return triggerManual(ctx, params, token, a.Name())
}

// ParameterNew 初始化参数
func (a *CronTrigger) ParameterNew() interface{} {
	return &CronTriggerParam{}
}

// CronWeekTrigger 定时触发
type CronWeekTrigger struct {
}

// Name 操作名称
func (a *CronWeekTrigger) Name() string {
	return common.CronWeekTrigger
}

// Run 操作方法
func (a *CronWeekTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	return triggerManual(ctx, params, token, a.Name())
}

// ParameterNew 初始化参数
func (a *CronWeekTrigger) ParameterNew() interface{} {
	return &CronTriggerParam{}
}

// CronMonthTrigger 定时触发
type CronMonthTrigger struct {
}

// Name 操作名称
func (a *CronMonthTrigger) Name() string {
	return common.CronMonthTrigger
}

// Run 操作方法
func (a *CronMonthTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	return triggerManual(ctx, params, token, a.Name())
}

// ParameterNew 初始化参数
func (a *CronMonthTrigger) ParameterNew() interface{} {
	return &CronTriggerParam{}
}

// CronCustomTrigger 定时触发
type CronCustomTrigger struct {
}

// Name 操作名称
func (a *CronCustomTrigger) Name() string {
	return common.CronCustomTrigger
}

// Run 操作方法
func (a *CronCustomTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	return triggerManual(ctx, params, token, a.Name())
}

// ParameterNew 初始化参数
func (a *CronCustomTrigger) ParameterNew() interface{} {
	return &CronTriggerParam{}
}

// ManualTrigger 手动触发器
type ManualTrigger struct {
}

// ManualTriggerParam 手动触发器参数
type ManualTriggerParam struct {
}

// Name 操作名称
func (a *ManualTrigger) Name() string {
	return common.MannualTrigger
}

// Run 操作方法 手动触发
func (a *ManualTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	return triggerManual(ctx, params, token, a.Name())
}

// ParameterNew new parameter
func (a *ManualTrigger) ParameterNew() interface{} {
	return &ManualTriggerParam{}
}

func triggerManual(ctx entity.ExecuteContext, params interface{}, token *entity.Token, actionName string) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	usermgntAdaper := drivenadapters.NewUserManagement()
	ecotagAdapter := drivenadapters.NewEcoTag()
	kcAdapter := drivenadapters.NewKcmc()

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	_id, ok := ctx.GetVar("id")

	if !ok {
		if _id, ok = ctx.GetVar("docid"); !ok {
			ctx.Trace(ctx.Context(), "run end")
			return nil, nil
		}
	}

	idStr, _ := _id.(string)

	var data = make(map[string]interface{})

	defer func() {
		patchDagInstanceSource(ctx, data, actionName)
	}()

	sourceType, _ := ctx.GetVar("source_type")
	if sourceType == "" || sourceType == "doc" {
		data["id"] = idStr
		data["docid"] = idStr
		data["item_id"] = utils.GetDocCurID(idStr)
		data["_type"] = "file"

		attr, doc, err := getDocInfo(ctx.Context(), idStr, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
		if err == nil {
			data = attr
			data["modify_time"] = attr["modified"]
			data["item_id"] = utils.GetDocCurID(doc.DocID)
			if doc.Size == -1 {
				data["_type"] = "folder"
			} else {
				data["_type"] = "file"
			}
		}

	} else if sourceType == "dept" {
		data["id"] = idStr
		data["_type"] = "department"
		departInfo, err := usermgntAdaper.GetDepartmentInfo(idStr)
		if err == nil {
			data["id"] = departInfo.DepartmentID
			data["name"] = departInfo.Name

			if len(departInfo.ParentDeps) > 0 {
				parentDep := departInfo.ParentDeps[len(departInfo.ParentDeps)-1]
				data["parent_id"] = parentDep.ID
			}

			parent := make([]string, 0)
			for _, dep := range departInfo.ParentDeps {
				parent = append(parent, dep.Name)
			}
			data["parent"] = strings.Join(parent, "/")

			mails, err := usermgntAdaper.GetDeptMailList([]string{idStr})
			if err != nil {
				traceLog.WithContext(ctx.Context()).Warnln(err)
				return data, nil
			}

			if len(mails) > 0 {
				data["email"] = mails[0]
			}
		}
	} else if sourceType == common.User.ToString() {
		userID := idStr
		data["id"] = userID
		data["_type"] = "user"
		userInfo, err := usermgntAdaper.GetUserInfo(userID)
		if err == nil {
			data["name"] = userInfo.UserName
			roles, _ := json.Marshal(userInfo.Roles)
			data["role"] = string(roles)
			data["csflevel"] = userInfo.CsfLevel
			data["email"] = userInfo.Email
			data["contact"] = userInfo.Telephone
			parentIDs, _ := json.Marshal(userInfo.ParentDepIDs)
			data["parent_ids"] = string(parentIDs)
			data["status"] = "enabled"
			if !userInfo.Enabled {
				data["status"] = "disabled"
			}
		}

		kcUserInfo, err := kcAdapter.GetUserEntity(ctx.Context(), userID, token.Token)
		if err == nil {
			kcUserInfoData := kcUserInfo.Data
			data["email"] = kcUserInfoData.Email
			data["contact"] = kcUserInfoData.Contact
			data["tags"] = kcUserInfoData.Tags
			data["is_expert"] = kcUserInfoData.IsExpert
			data["verification_info"] = kcUserInfoData.VerificationInfo
			data["university"] = kcUserInfoData.University
			data["position"] = kcUserInfoData.Position
			var targetLen = 12
			var parsed = make([]string, 0)
			jerr := json.Unmarshal([]byte(kcUserInfoData.WorkAt), &parsed)
			if jerr != nil {
				data["work_at"] = kcUserInfoData.WorkAt
			}
			if len(parsed) > 0 {
				code := parsed[len(parsed)-1]
				paddingString := fmt.Sprintf("%s%0*d", code, targetLen-len(code), 0)
				data["work_at"] = paddingString
			}
			data["professional"] = kcUserInfoData.Professional
		}
	} else if sourceType == "tagtree" {
		tagID := idStr
		data["id"] = tagID
		data["_type"] = "tag"
		tagInfos, err := ecotagAdapter.GetTags(ctx.Context(), map[string][]string{"id": {tagID}})
		if err == nil && len(tagInfos) > 0 {
			tagInfo := tagInfos[0]
			data["id"] = tagInfo.ID
			data["path"] = tagInfo.Path
			data["version"] = tagInfo.Version
			data["name"] = tagInfo.Name
			if strings.Contains(tagInfo.Path, "/") {
				parentPath := strings.TrimSuffix(tagInfo.Path, "/"+tagInfo.Name)
				parentTagInfos, err := ecotagAdapter.GetTags(ctx.Context(), map[string][]string{"path": {parentPath}})
				if err != nil {
					traceLog.WithContext(ctx.Context()).Warnln(err)
					return data, nil
				}
				data["parent_id"] = parentTagInfos[0].ID
			}
		}
	}

	datasourceid, ok := ctx.GetVar("datasourceid")

	if ok {
		id, idOk := datasourceid.(string)
		if idOk {
			ctx.ShareData().Set(id, data)
		}
	} else {
		id := ctx.GetTaskID()
		ctx.ShareData().Set(id, data)
	}
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// FormTrigger 手动触发器
type FormTrigger struct {
}

type FormTriggerParamField struct {
	Name     string `json:"name"`
	Key      string `json:"key"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Default  any    `json:"default"`
}

// FormTriggerParam 手动触发器参数
type FormTriggerParam struct {
	Fields []*FormTriggerParamField `json:"fields"`
}

// Name 操作名称
func (a *FormTrigger) Name() string {
	return common.FormTrigger
}

// Run 操作方法 表单触发
func (a *FormTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	input, ok := params.(*FormTriggerParam)
	if !ok {
		return nil, fmt.Errorf("invalid parameter type")
	}

	attr := make(map[string]interface{})
	attr["_type"] = "form"
	fields := make(map[string]interface{})
	finalFields := make(map[string]any)
	accessor := Accessor{}
	defer func() {
		patchDagInstanceSource(ctx, attr, a.Name())
	}()

	isVM := false
	runArgs := make(map[string]any)

	ctx.IterateVars(func(key, val string) (stop bool) {
		if key == "operator_id" {
			accessor.ID = val
		} else if key == "operator_name" {
			accessor.Name = val
		} else if key == "operator_type" {
			accessor.Type = val
		} else if key == "run_mode" {
			isVM = true
		} else if key == "run_args" {
			_ = json.Unmarshal([]byte(val), &runArgs)
		} else {
			fields[key] = val
		}
		return false
	})

	if isVM {
		for k, v := range runArgs {
			fields[k] = v
		}
	}

	for _, f := range input.Fields {
		if v, ok := fields[f.Key]; ok {
			finalFields[f.Key] = v
			continue
		}

		if f.Default != nil {
			finalFields[f.Key] = f.Default
			continue
		}

		if f.Required {
			title := f.Key
			if f.Name != "" {
				title = f.Name
			}
			return nil, fmt.Errorf("%s is required", title)
		}
	}

	accessorBytes, _ := json.Marshal(accessor)

	attr["fields"] = finalFields
	attr["accessor"] = string(accessorBytes)

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, attr)
	ctx.Trace(ctx.Context(), "run end")
	return attr, nil
}

// ParameterNew new parameter
func (a *FormTrigger) ParameterNew() interface{} {
	return &FormTriggerParam{}
}

// AnyshareFileUploadTrigger 上传文件触发
type AnyshareFileUploadTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFileUploadTrigger) Name() string {
	return common.AnyshareFileUploadTrigger
}

// Run 操作方法 上传文件触发
func (a *AnyshareFileUploadTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	data := getTriggerVars(ctx)
	data["_type"] = "file"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	docID := data["id"].(string)
	attr, doc, err := getDocInfo(ctx.Context(), docID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data["creator"] = doc.Creator
	data["creator_id"] = doc.CreatorID
	data["create_time"] = attr["create_time"]
	data["modify_time"] = attr["modified"]
	data["rev"] = doc.Rev
	data["path"] = doc.Path
	data["csflevel"] = doc.CsfLevel
	data["editor"] = doc.Editor
	data["editor_id"] = doc.EditorID
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFileUploadTrigger) ParameterNew() interface{} {
	return &AnyshareFileUploadTrigger{}
}

// AnyshareFileCopyTrigger 复制文件触发
type AnyshareFileCopyTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFileCopyTrigger) Name() string {
	return common.AnyshareFileCopyTrigger
}

// Run 操作方法 复制文件触发
func (a *AnyshareFileCopyTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := getTriggerVars(ctx)
	data["_type"] = "file"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	newDocID := data["new_id"].(string)
	attr, doc, err := getDocInfo(ctx.Context(), newDocID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data["creator"] = doc.Creator
	data["creator_id"] = doc.CreatorID
	data["create_time"] = attr["create_time"]
	data["modify_time"] = attr["modified"]
	data["rev"] = doc.Rev
	data["path"] = doc.Path
	data["csflevel"] = doc.CsfLevel
	data["editor"] = doc.Editor
	data["editor_id"] = doc.EditorID
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFileCopyTrigger) ParameterNew() interface{} {
	return &AnyshareFileCopyTrigger{}
}

// AnyshareFileMoveTrigger 移动文件触发
type AnyshareFileMoveTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFileMoveTrigger) Name() string {
	return common.AnyshareFileMoveTrigger
}

// Run 操作方法 移动文件触发
func (a *AnyshareFileMoveTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := getTriggerVars(ctx)
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	paths := strings.Split(data["new_path"].(string), "/")
	data["_type"] = "file"
	data["name"] = paths[len(paths)-1]
	data["path"] = data["new_path"]
	data["id"] = data["new_id"]
	attr, doc, err := getDocInfo(ctx.Context(), data["id"].(string), token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data["creator"] = doc.Creator
	data["creator_id"] = doc.CreatorID
	data["create_time"] = attr["create_time"]
	data["modify_time"] = attr["modified"]
	data["rev"] = doc.Rev
	data["path"] = doc.Path
	data["csflevel"] = doc.CsfLevel
	data["editor"] = doc.Editor
	data["editor_id"] = doc.EditorID
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFileMoveTrigger) ParameterNew() interface{} {
	return &AnyshareFileMoveTrigger{}
}

// AnyshareFileRemoveTrigger 删除文件触发
type AnyshareFileRemoveTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFileRemoveTrigger) Name() string {
	return common.AnyshareFileRemoveTrigger
}

// Run 操作方法 删除文件触发
func (a *AnyshareFileRemoveTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := getTriggerVars(ctx)
	data["_type"] = "file"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFileRemoveTrigger) ParameterNew() interface{} {
	return &AnyshareFileRemoveTrigger{}
}

// AnyshareFolderCreateTrigger 创建目录触发
type AnyshareFolderCreateTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFolderCreateTrigger) Name() string {
	return common.AnyshareFolderCreateTrigger
}

// Run 操作方法 创建目录触发
func (a *AnyshareFolderCreateTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := getTriggerVars(ctx)
	data["_type"] = "folder"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	docID := data["id"].(string)
	attr, doc, err := getDocInfo(ctx.Context(), docID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data["creator"] = doc.Creator
	data["create_time"] = attr["create_time"]
	data["editor"] = doc.Editor
	data["editor_id"] = doc.EditorID
	data["modify_time"] = attr["modified"]
	data["path"] = doc.Path
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFolderCreateTrigger) ParameterNew() interface{} {
	return &AnyshareFolderCreateTrigger{}
}

// AnyshareFolderMoveTrigger 移动目录触发
type AnyshareFolderMoveTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFolderMoveTrigger) Name() string {
	return common.AnyshareFolderMoveTrigger
}

// Run 操作方法 移动目录触发
func (a *AnyshareFolderMoveTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := getTriggerVars(ctx)
	data["_type"] = "folder"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	paths := strings.Split(data["new_path"].(string), "/")
	data["name"] = paths[len(paths)-1]
	data["path"] = data["new_path"]
	data["id"] = data["new_id"]
	attr, doc, err := getDocInfo(ctx.Context(), data["id"].(string), token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data["creator"] = doc.Creator
	data["create_time"] = attr["create_time"]
	data["editor"] = doc.Editor
	data["editor_id"] = doc.EditorID
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFolderMoveTrigger) ParameterNew() interface{} {
	return &AnyshareFolderMoveTrigger{}
}

// AnyshareFolderCopyTrigger 复制目录触发
type AnyshareFolderCopyTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFolderCopyTrigger) Name() string {
	return common.AnyshareFolderCopyTrigger
}

// Run 操作方法 复制目录触发
func (a *AnyshareFolderCopyTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := getTriggerVars(ctx)
	data["_type"] = "folder"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	newDocID := data["new_id"].(string)
	attr, newDoc, err := getDocInfo(ctx.Context(), newDocID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data["new_path"] = newDoc.Path
	data["name"] = newDoc.Name
	data["creator"] = newDoc.Creator
	data["create_time"] = attr["create_time"]
	data["modify_time"] = attr["modified"]
	data["path"] = newDoc.Path
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFolderCopyTrigger) ParameterNew() interface{} {
	return &AnyshareFolderCopyTrigger{}
}

// AnyshareFolderRemoveTrigger 删除目录触发
type AnyshareFolderRemoveTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFolderRemoveTrigger) Name() string {
	return common.AnyshareFolderRemoveTrigger
}

// Run 操作方法 删除目录触发
func (a *AnyshareFolderRemoveTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := getTriggerVars(ctx)
	data["_type"] = "folder"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFolderRemoveTrigger) ParameterNew() interface{} {
	return &AnyshareFolderRemoveTrigger{}
}

// AnyshareFileReversionTrigger 还原文件版本触发
type AnyshareFileReversionTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFileReversionTrigger) Name() string {
	return common.AnyshareFileReversionTrigger
}

// Run 操作方法 还原文件版本触发
func (a *AnyshareFileReversionTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	data := getTriggerVars(ctx)
	data["_type"] = "file"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	docID := data["id"].(string)
	attr, doc, err := getDocInfo(ctx.Context(), docID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data["creator"] = doc.Creator
	data["creator_id"] = doc.CreatorID
	data["create_time"] = attr["create_time"]
	data["modify_time"] = attr["modified"]
	data["rev"] = doc.Rev
	data["path"] = doc.Path
	data["csflevel"] = doc.CsfLevel
	data["editor"] = doc.Editor
	data["editor_id"] = doc.EditorID
	data["size"] = doc.Size
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFileReversionTrigger) ParameterNew() interface{} {
	return &AnyshareFileReversionTrigger{}
}

// AnyshareFileRestoreTrigger 从回收站还原文件触发
type AnyshareFileRestoreTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFileRestoreTrigger) Name() string {
	return common.AnyshareFileRestoreTrigger
}

// Run 操作方法 从回收站还原文件触发
func (a *AnyshareFileRestoreTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	data := getTriggerVars(ctx)
	data["_type"] = "file"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	docID := data["id"].(string)
	attr, doc, err := getDocInfo(ctx.Context(), docID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data["creator"] = doc.Creator
	data["creator_id"] = doc.CreatorID
	data["create_time"] = attr["create_time"]
	data["modify_time"] = attr["modified"]
	data["rev"] = doc.Rev
	data["path"] = doc.Path
	data["csflevel"] = doc.CsfLevel
	data["name"] = doc.Name
	data["size"] = doc.Size
	data["editor"] = doc.Editor
	data["editor_id"] = doc.EditorID
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFileRestoreTrigger) ParameterNew() interface{} {
	return &AnyshareFileRestoreTrigger{}
}

// AnyshareFileRenameTrigger 重命名触发
type AnyshareFileRenameTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFileRenameTrigger) Name() string {
	return common.AnyshareFileRenameTrigger
}

// Run 操作方法 重命名文件触发
func (a *AnyshareFileRenameTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := getTriggerVars(ctx)
	data["_type"] = "file"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	data["path"] = data["new_path"]
	attr, doc, err := getDocInfo(ctx.Context(), data["id"].(string), token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data["creator"] = doc.Creator
	data["creator_id"] = doc.CreatorID
	data["create_time"] = attr["create_time"]
	data["modify_time"] = attr["modified"]
	data["rev"] = doc.Rev
	data["csflevel"] = doc.CsfLevel
	data["name"] = doc.Name
	data["size"] = doc.Size
	data["editor"] = doc.Editor
	data["editor_id"] = doc.EditorID
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFileRenameTrigger) ParameterNew() interface{} {
	return &AnyshareFileRenameTrigger{}
}

// AnyshareFileDeleteTrigger 删除文件触发
type AnyshareFileDeleteTrigger struct {
	DocID string
}

// Name 操作名称
func (a *AnyshareFileDeleteTrigger) Name() string {
	return common.AnyshareFileDeleteTrigger
}

// Run 操作方法 彻底删除文件触发
func (a *AnyshareFileDeleteTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := getTriggerVars(ctx)
	data["_type"] = "file"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFileDeleteTrigger) ParameterNew() interface{} {
	return &AnyshareFileDeleteTrigger{}
}

// AnyshareUserCreateTrigger 创建用户时触发
type AnyshareUserCreateTrigger struct {
}

// Name 操作名称
func (a *AnyshareUserCreateTrigger) Name() string {
	return common.AnyshareUserCreateTrigger
}

// Run 操作方法 用户创建时触发
func (a *AnyshareUserCreateTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	usermgntAdaper := drivenadapters.NewUserManagement()

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "user"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	_id, _ := ctx.GetVar("id")
	userID := _id.(string)
	userInfo, err := usermgntAdaper.GetUserInfo(userID)
	if err != nil {
		traceLog.WithContext(ctx.Context()).Warnln(err)
		return nil, err
	}
	data["id"] = userID
	data["name"] = userInfo.UserName
	data["role"] = userInfo.Roles
	data["csflevel"] = userInfo.CsfLevel
	data["email"] = userInfo.Email
	data["contact"] = userInfo.Telephone
	parentIDs, _ := json.Marshal(userInfo.ParentDepIDs)
	data["parent_ids"] = string(parentIDs)
	data["status"] = "enabled"
	if !userInfo.Enabled {
		data["status"] = "disabled"
	}
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareUserCreateTrigger) ParameterNew() interface{} {
	return &AnyshareUserCreateTrigger{}
}

// AnyshareUserDeleteTrigger 删除用户时触发
type AnyshareUserDeleteTrigger struct {
}

// Name 操作名称
func (a *AnyshareUserDeleteTrigger) Name() string {
	return common.AnyshareUserDeleteTrigger
}

// Run 操作方法 删除用户时触发
func (a *AnyshareUserDeleteTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "user"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	_id, _ := ctx.GetVar("id")
	userID := _id.(string)
	data["id"] = userID
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareUserDeleteTrigger) ParameterNew() interface{} {
	return &AnyshareUserDeleteTrigger{}
}

// AnyshareUserFreezeTrigger 冻结用户时触发
type AnyshareUserFreezeTrigger struct {
}

// Name 操作名称
func (a *AnyshareUserFreezeTrigger) Name() string {
	return common.AnyshareUserFreezeTrigger
}

// Run 操作方法 冻结用户时触发
func (a *AnyshareUserFreezeTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "user"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	_id, _ := ctx.GetVar("id")
	userID := _id.(string)
	data["id"] = userID
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareUserFreezeTrigger) ParameterNew() interface{} {
	return &AnyshareUserFreezeTrigger{}
}

// AnyshareOrgNameModifyTrigger 组织改名时触发
type AnyshareOrgNameModifyTrigger struct {
}

// Name 操作名称
func (a *AnyshareOrgNameModifyTrigger) Name() string {
	return common.AnyshareOrgNameModifyTrigger
}

// Run 操作方法 组织改名时触发
func (a *AnyshareOrgNameModifyTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "department"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	_id, _ := ctx.GetVar("id")
	userID := _id.(string)
	_newName, _ := ctx.GetVar("new_name")
	newName := _newName.(string)
	_dtype, _ := ctx.GetVar("type")
	dtype := _dtype.(string)
	data["id"] = userID
	data["new_name"] = newName
	data["type"] = dtype

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareOrgNameModifyTrigger) ParameterNew() interface{} {
	return &AnyshareOrgNameModifyTrigger{}
}

// AnyshareUserMovedTrigger 用户移动时触发
type AnyshareUserMovedTrigger struct {
}

// Name 操作名称
func (a *AnyshareUserMovedTrigger) Name() string {
	return common.AnyshareUserMovedTrigger
}

// Run 操作方法 用户移动时触发
func (a *AnyshareUserMovedTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "user"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	_id, _ := ctx.GetVar("id")
	userID := _id.(string)
	_newDeptPath, _ := ctx.GetVar("new_dept_path")
	newDeptPath := _newDeptPath.(string)
	_oldDeptPath, _ := ctx.GetVar("old_dept_path")
	oldDeptPath := _oldDeptPath.(string)
	data["id"] = userID
	data["old_dept_path"] = utils.GetCurDeptID(oldDeptPath)
	data["new_dept_path"] = utils.GetCurDeptID(newDeptPath)

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareUserMovedTrigger) ParameterNew() interface{} {
	return &AnyshareUserMovedTrigger{}
}

// AnyshareDeptMovedTrigger 用户移动时触发
type AnyshareDeptMovedTrigger struct {
}

// Name 操作名称
func (a *AnyshareDeptMovedTrigger) Name() string {
	return common.AnyshareDeptMovedTrigger
}

// Run 操作方法 部门移动时触发
func (a *AnyshareDeptMovedTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	usermgntAdaper := drivenadapters.NewUserManagement()

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "department"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	_id, _ := ctx.GetVar("id")
	ID := _id.(string)
	data["id"] = ID

	departInfo, err := usermgntAdaper.GetDepartmentInfo(ID)
	if err != nil {
		traceLog.WithContext(ctx.Context()).Warnln(err)
		return data, nil
	}

	data["name"] = departInfo.Name
	if len(departInfo.ParentDeps) > 0 {
		parentDep := departInfo.ParentDeps[len(departInfo.ParentDeps)-1]
		data["parent_id"] = parentDep.ID
	}

	parent := make([]string, 0)
	for _, dep := range departInfo.ParentDeps {
		parent = append(parent, dep.Name)
	}
	data["parent"] = strings.Join(parent, "/")

	mails, err := usermgntAdaper.GetDeptMailList([]string{ID})
	if err != nil {
		traceLog.WithContext(ctx.Context()).Warnln(err)
		return data, nil
	}

	if len(mails) > 0 {
		data["email"] = mails[0]
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareDeptMovedTrigger) ParameterNew() interface{} {
	return &AnyshareDeptMovedTrigger{}
}

// AnyshareUserAddDeptTrigger 添加用户到部门时触发
type AnyshareUserAddDeptTrigger struct {
}

// Name 操作名称
func (a *AnyshareUserAddDeptTrigger) Name() string {
	return common.AnyshareUserAddDeptTrigger
}

// Run 操作方法 添加用户到部门时触发
func (a *AnyshareUserAddDeptTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "user"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	deps := make([]map[string]interface{}, 0)
	_id, _ := ctx.GetVar("id")
	userID := _id.(string)
	_deptPaths, _ := ctx.GetVar("dept_paths")
	deptPaths := _deptPaths.(string)
	var paths []string
	err = json.Unmarshal([]byte(deptPaths), &paths)
	if err != nil {
		traceLog.WithContext(ctx.Context()).Warnln(err)
	}

	for _, path := range paths {
		curPath := utils.GetCurDeptID(path)
		deps = append(deps, map[string]interface{}{"id": userID, "new_dept_path": curPath})
	}
	res, _ := json.Marshal(deps)
	data["data"] = string(res)

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareUserAddDeptTrigger) ParameterNew() interface{} {
	return &AnyshareUserAddDeptTrigger{}
}

// AnyshareUserRemoveDeptTrigger 从部门中移除用户时触发
type AnyshareUserRemoveDeptTrigger struct {
}

// Name 操作名称
func (a *AnyshareUserRemoveDeptTrigger) Name() string {
	return common.AnyshareUserRemoveDeptTrigger
}

// Run 操作方法 从部门中移除用户时触发
func (a *AnyshareUserRemoveDeptTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "user"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	deps := make([]map[string]interface{}, 0)
	_id, _ := ctx.GetVar("id")
	userID := _id.(string)
	_deptPaths, _ := ctx.GetVar("dept_paths")
	deptPaths := _deptPaths.(string)
	var paths []string
	err = json.Unmarshal([]byte(deptPaths), &paths)
	if err != nil {
		traceLog.WithContext(ctx.Context()).Warnln(err)
	}
	for _, path := range paths {
		curPath := utils.GetCurDeptID(path)
		deps = append(deps, map[string]interface{}{"id": userID, "old_dept_path": curPath})
	}
	res, _ := json.Marshal(deps)
	data["data"] = string(res)

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareUserRemoveDeptTrigger) ParameterNew() interface{} {
	return &AnyshareUserRemoveDeptTrigger{}
}

// AnyshareDeptCreateTrigger 创建部门时触发
type AnyshareDeptCreateTrigger struct {
}

// Name 操作名称
func (a *AnyshareDeptCreateTrigger) Name() string {
	return common.AnyshareDeptCreateTrigger
}

// Run 操作方法 创建部门时触发
func (a *AnyshareDeptCreateTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	usermgntAdaper := drivenadapters.NewUserManagement()

	data := make(map[string]interface{})
	data["_type"] = "department"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	_id, _ := ctx.GetVar("id")
	ID := _id.(string)
	_name, _ := ctx.GetVar("name")
	name := _name.(string)
	data["id"] = ID
	data["name"] = name

	departInfo, err := usermgntAdaper.GetDepartmentInfo(ID)
	if err != nil {
		traceLog.WithContext(ctx.Context()).Warnln(err)
		return data, nil
	}

	if len(departInfo.ParentDeps) > 0 {
		parentDep := departInfo.ParentDeps[len(departInfo.ParentDeps)-1]
		data["parent_id"] = parentDep.ID
	}
	parent := make([]string, 0)
	for _, dep := range departInfo.ParentDeps {
		parent = append(parent, dep.Name)
	}
	data["parent"] = strings.Join(parent, "/")

	mails, err := usermgntAdaper.GetDeptMailList([]string{ID})
	if err != nil {
		traceLog.WithContext(ctx.Context()).Warnln(err)
		return data, nil
	}

	if len(mails) > 0 {
		data["email"] = mails[0]
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareDeptCreateTrigger) ParameterNew() interface{} {
	return &AnyshareDeptCreateTrigger{}
}

// AnyshareDeptDeleteTrigger 删除部门时触发
type AnyshareDeptDeleteTrigger struct {
}

// Name 操作名称
func (a *AnyshareDeptDeleteTrigger) Name() string {
	return common.AnyshareDeptDeleteTrigger
}

// Run 操作方法 删除部门时触发
func (a *AnyshareDeptDeleteTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "department"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	_id, _ := ctx.GetVar("id")
	ID := _id.(string)
	data["id"] = ID

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareDeptDeleteTrigger) ParameterNew() interface{} {
	return &AnyshareDeptDeleteTrigger{}
}

// AnyshareUserChangeTrigger 用户变化时触发
type AnyshareUserChangeTrigger struct {
}

// Name 操作名称
func (a *AnyshareUserChangeTrigger) Name() string {
	return common.AnyshareUserChangeTrigger
}

// Run 操作方法 KC用户变化时触发
func (a *AnyshareUserChangeTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "user"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	ctx.IterateVars(func(key, val string) (stop bool) {
		data[key] = val
		return false
	})

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareUserChangeTrigger) ParameterNew() interface{} {
	return &AnyshareUserChangeTrigger{}
}

// AnyshareTagTreeCreateTrigger 添加标签树时触发
type AnyshareTagTreeCreateTrigger struct {
}

// Name 操作名称
func (a *AnyshareTagTreeCreateTrigger) Name() string {
	return common.AnyshareTagTreeCreateTrigger
}

// Run 操作方法 添加标签树时触发
func (a *AnyshareTagTreeCreateTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "tag"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	ctx.IterateVars(func(key, val string) (stop bool) {
		data[key] = val
		return false
	})

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareTagTreeCreateTrigger) ParameterNew() interface{} {
	return &AnyshareTagTreeCreateTrigger{}
}

// AnyshareTagTreeAddedTrigger 增加标签时触发
type AnyshareTagTreeAddedTrigger struct {
}

// Name 操作名称
func (a *AnyshareTagTreeAddedTrigger) Name() string {
	return common.AnyshareTagTreeAddedTrigger
}

// Run 操作方法 增加标签时触发
func (a *AnyshareTagTreeAddedTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "tag"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	ctx.IterateVars(func(key, val string) (stop bool) {
		data[key] = val
		return false
	})

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareTagTreeAddedTrigger) ParameterNew() interface{} {
	return &AnyshareTagTreeAddedTrigger{}
}

// AnyshareTagTreeEditedTrigger 编辑标签时触发
type AnyshareTagTreeEditedTrigger struct {
}

// Name 操作名称
func (a *AnyshareTagTreeEditedTrigger) Name() string {
	return common.AnyshareTagTreeEditedTrigger
}

// Run 操作方法 编辑标签时触发
func (a *AnyshareTagTreeEditedTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "tag"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	ctx.IterateVars(func(key, val string) (stop bool) {
		data[key] = val
		return false
	})

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareTagTreeEditedTrigger) ParameterNew() interface{} {
	return &AnyshareTagTreeEditedTrigger{}
}

// AnyshareTagTreeDeletedTrigger 编辑标签时触发
type AnyshareTagTreeDeletedTrigger struct {
}

// Name 操作名称
func (a *AnyshareTagTreeDeletedTrigger) Name() string {
	return common.AnyshareTagTreeDeletedTrigger
}

// Run 操作方法 删除标签时触发
func (a *AnyshareTagTreeDeletedTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	data := make(map[string]interface{})
	data["_type"] = "tag"
	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()
	ctx.IterateVars(func(key, val string) (stop bool) {
		data[key] = val
		return false
	})

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *AnyshareTagTreeDeletedTrigger) ParameterNew() interface{} {
	return &AnyshareTagTreeDeletedTrigger{}
}

type AnyShareSelectedFileTrigger struct{}

func (a AnyShareSelectedFileTrigger) Name() string {
	return common.OpAnyShareSelectedFileTrigger
}

func (a AnyShareSelectedFileTrigger) ParameterNew() interface{} {
	return &AnyShareSelectedFileTrigger{}
}

func (a AnyShareSelectedFileTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	attr := make(map[string]interface{})
	fields := make(map[string]interface{})
	accessor := Accessor{}
	var source map[string]interface{}
	attr["_type"] = "file"
	defer func() {
		patchDagInstanceSource(ctx, attr, a.Name())
	}()
	ctx.IterateVars(func(key, val string) (stop bool) {
		if key == "operator_id" {
			accessor.ID = val
		} else if key == "operator_name" {
			accessor.Name = val
		} else if key == "operator_type" {
			accessor.Type = val
		} else if key == "source" {
			json.Unmarshal([]byte(val), &source)
		} else {
			fields[key] = val
		}
		return false
	})

	if name, ok := source["name"]; ok {
		attr["name"] = name
	}

	accessorBytes, _ := json.Marshal(accessor)

	attr["fields"] = fields
	attr["accessor"] = string(accessorBytes)
	attr["source"] = source

	id := ctx.GetTaskID()
	ctx.ShareData().Set("source", source)
	ctx.ShareData().Set(id, attr)
	ctx.Trace(ctx.Context(), "run end")
	return attr, nil
}

type AnyShareSelectedFolderTrigger struct{}

func (a AnyShareSelectedFolderTrigger) Name() string {
	return common.OpAnyShareSelectedFolderTrigger
}

func (a AnyShareSelectedFolderTrigger) ParameterNew() interface{} {
	return &AnyShareSelectedFolderTrigger{}
}

func (a AnyShareSelectedFolderTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	attr := make(map[string]interface{})
	fields := make(map[string]interface{})
	accessor := Accessor{}
	var source map[string]interface{}
	attr["_type"] = "folder"
	defer func() {
		patchDagInstanceSource(ctx, attr, a.Name())
	}()
	ctx.IterateVars(func(key, val string) (stop bool) {
		if key == "operator_id" {
			accessor.ID = val
		} else if key == "operator_name" {
			accessor.Name = val
		} else if key == "operator_type" {
			accessor.Type = val
		} else if key == "source" {
			json.Unmarshal([]byte(val), &source)
		} else {
			fields[key] = val
		}
		return false
	})

	accessorBytes, _ := json.Marshal(accessor)

	attr["fields"] = fields
	attr["accessor"] = string(accessorBytes)
	attr["source"] = source
	if name, ok := source["name"]; ok {
		attr["name"] = name
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set("source", source)
	ctx.ShareData().Set(id, attr)
	ctx.Trace(ctx.Context(), "run end")
	return attr, nil
}

// DataFlowDocTrigger 手动触发器
type DataFlowDocTrigger struct {
}

// DataFlowDocTriggerParam 手动触发器参数
type DataFlowDocTriggerParam struct {
}

// Name 操作名称
func (a *DataFlowDocTrigger) Name() string {
	return common.DataflowDocTrigger
}

// Run 操作方法 手动触发
func (a *DataFlowDocTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	// 如果包含bucket参数，说明是S3数据源触发
	if _, ok := ctx.GetVar("bucket"); ok {
		var err error
		newCtx, span := trace.StartInternalSpan(ctx.Context())
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		ctx.SetContext(newCtx)

		ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

		var data = make(map[string]interface{})
		data["_type"] = "s3"

		defer func() {
			patchDagInstanceSource(ctx, data, a.Name())
		}()

		// 从上下文变量中获取S3对象信息
		id, _ := ctx.GetVar("id")
		bucket, _ := ctx.GetVar("bucket")
		key, _ := ctx.GetVar("key")
		size, _ := ctx.GetVar("size")
		lastModified, _ := ctx.GetVar("last_modified")
		etag, _ := ctx.GetVar("etag")
		downloadURL, _ := ctx.GetVar("download_url")

		if idStr, ok := id.(string); ok && idStr != "" {
			data["id"] = idStr
			data["object_id"] = idStr
			data["doc_id"] = idStr
		}
		if bucketStr, ok := bucket.(string); ok && bucketStr != "" {
			data["bucket"] = bucketStr
		}
		if keyStr, ok := key.(string); ok && keyStr != "" {
			data["key"] = keyStr
			data["path"] = keyStr
			// 从key中提取文件名
			parts := strings.Split(keyStr, "/")
			if len(parts) > 0 {
				data["name"] = parts[len(parts)-1]
			}
		}
		if sizeStr, ok := size.(string); ok && sizeStr != "" {
			data["size"] = sizeStr
		}
		if lastModifiedStr, ok := lastModified.(string); ok && lastModifiedStr != "" {
			data["modify_time"] = lastModifiedStr
			data["create_time"] = lastModifiedStr

		}
		if etagStr, ok := etag.(string); ok && etagStr != "" {
			data["etag"] = etagStr
			data["md5"] = etagStr // ETag通常是MD5值
			data["rev"] = etagStr
		}
		if downloadURLStr, ok := downloadURL.(string); ok && downloadURLStr != "" {
			data["download_url"] = downloadURLStr
		}

		// 将数据存储到共享数据中
		datasourceid, ok := ctx.GetVar("datasourceid")
		if ok {
			id, idOk := datasourceid.(string)
			if idOk {
				ctx.ShareData().Set(id, data)
			}
		} else {
			id := ctx.GetTaskID()
			ctx.ShareData().Set(id, data)
		}

		ctx.Trace(ctx.Context(), "run end")
		return data, nil
	}

	// 检查是否为 Dataflow 文件子系统触发 (source_from 为 local、remote、form)
	if sourceFrom, ok := ctx.GetVar("source_from"); ok {
		if sourceFromStr, ok := sourceFrom.(string); ok && (sourceFromStr == "local" || sourceFromStr == "remote" || sourceFromStr == "form") {
			return a.runDataflowFileTrigger(ctx, params, token)
		}
	}

	return triggerManual(ctx, params, token, a.Name())
}

// runDataflowFileTrigger 处理 Dataflow 文件子系统触发
func (a *DataFlowDocTrigger) runDataflowFileTrigger(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	log := traceLog.WithContext(ctx.Context())
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	var data = make(map[string]interface{})
	data["_type"] = "file"
	data["source_type"] = "doc"

	defer func() {
		patchDagInstanceSource(ctx, data, a.Name())
	}()

	// 从上下文变量中获取文件信息
	id, _ := ctx.GetVar("id")
	userid, _ := ctx.GetVar("userid")
	operatorID, _ := ctx.GetVar("operator_id")
	operatorName, _ := ctx.GetVar("operator_name")
	operatorType, _ := ctx.GetVar("operator_type")
	downloadURL, _ := ctx.GetVar("download_url")

	if idStr, ok := id.(string); ok && idStr != "" {
		data["id"] = idStr
		data["docid"] = idStr
		data["doc_id"] = idStr
	}
	if useridStr, ok := userid.(string); ok && useridStr != "" {
		data["userid"] = useridStr
	}
	if operatorIDStr, ok := operatorID.(string); ok && operatorIDStr != "" {
		data["operator_id"] = operatorIDStr
	}
	if operatorNameStr, ok := operatorName.(string); ok && operatorNameStr != "" {
		data["operator_name"] = operatorNameStr
	}
	if operatorTypeStr, ok := operatorType.(string); ok && operatorTypeStr != "" {
		data["operator_type"] = operatorTypeStr
	}
	if downloadURLStr, ok := downloadURL.(string); ok && downloadURLStr != "" {
		data["download_url"] = downloadURLStr
	}

	// 检查文件状态，决定是否需要阻塞
	taskIns := ctx.GetTaskInstance()
	fileIDStr, _ := id.(string)
	if fileIDStr != "" && common.IsDFSURI(fileIDStr) {
		fileID, parseErr := common.ParseDFSURI(fileIDStr)
		if parseErr == nil {
			// 查询文件状态
			flowFile, queryErr := rds.GetFlowFileDao().GetByID(ctx.Context(), fileID)
			if queryErr == nil && flowFile != nil {
				// 添加文件名
				if flowFile.Name != "" {
					data["name"] = flowFile.Name
				}

				if flowFile.Status == rds.FlowFileStatusPending {
					// 文件未就绪，需要阻塞等待
					log.Infof("[DataFlowDocTrigger] File %d is pending, creating task_resume record", fileID)

					// 创建 task_resume 记录
					now := time.Now().Unix()
					taskResume := &rds.FlowTaskResume{
						ID:             store.NextID(),
						TaskInstanceID: taskIns.ID,
						DagInstanceID:  taskIns.RelatedDagInstance.ID,
						ResourceType:   "file",
						ResourceID:     fileID,
						CreatedAt:      now,
						UpdatedAt:      now,
					}

					if insertErr := rds.GetFlowTaskResumeDao().Insert(ctx.Context(), taskResume); insertErr != nil {
						log.Warnf("[DataFlowDocTrigger] Insert task_resume err: %s", insertErr.Error())
					}

					// 设置状态为 Blocked
					statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
					ctx.ShareData().Set(statusKey, entity.TaskInstanceStatusBlocked)

					// 设置文件状态信息，供后续恢复使用
					data["status"] = "pending"
				} else if flowFile.Status == rds.FlowFileStatusReady {
					// 文件已就绪
					data["status"] = "ready"

					// 如果有 storage 信息，补充相关字段
					if flowFile.StorageID > 0 {
						storage, storageErr := rds.GetFlowStorageDao().GetByID(ctx.Context(), flowFile.StorageID)
						if storageErr == nil && storage != nil {
							data["size"] = storage.Size
							if storage.ContentType != "" {
								data["content_type"] = storage.ContentType
							}
							if storage.Etag != "" {
								data["etag"] = storage.Etag
								data["md5"] = storage.Etag
							}
						}
					}
				} else if flowFile.Status == rds.FlowFileStatusInvalid {
					// 文件无效
					data["status"] = "invalid"
				}
			}
		}
	}

	// 将数据存储到共享数据中
	datasourceid, ok := ctx.GetVar("datasourceid")
	if ok {
		dsid, dsidOk := datasourceid.(string)
		if dsidOk {
			ctx.ShareData().Set(dsid, data)
		}
	} else {
		tid := ctx.GetTaskID()
		ctx.ShareData().Set(tid, data)
	}

	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// RunAfter 执行后检查，决定是否阻塞
func (a *DataFlowDocTrigger) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	taskIns := ctx.GetTaskInstance()
	statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
	status, ok := ctx.ShareData().Get(statusKey)
	if ok && status == entity.TaskInstanceStatusBlocked {
		return entity.TaskInstanceStatusBlocked, nil
	}

	return entity.TaskInstanceStatusSuccess, nil
}

// ParameterNew new parameter
func (a *DataFlowDocTrigger) ParameterNew() interface{} {
	return &DataFlowDocTriggerParam{}
}

type DataFlowUserTrigger struct {
}

// DataFlowDocTriggerParam 手动触发器参数
type DataFlowUserTriggerParam struct {
}

// Name 操作名称
func (a *DataFlowUserTrigger) Name() string {
	return common.DataflowUserTrigger
}

// Run 操作方法 手动触发
func (a *DataFlowUserTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	return triggerManual(ctx, params, token, a.Name())
}

// ParameterNew new parameter
func (a *DataFlowUserTrigger) ParameterNew() interface{} {
	return &DataFlowUserTriggerParam{}
}

type DataFlowDeptTrigger struct {
}

type DataFlowDeptTriggerParam struct {
}

func (a *DataFlowDeptTrigger) Name() string {
	return common.DataflowDeptTrigger
}

func (a *DataFlowDeptTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	return triggerManual(ctx, params, token, a.Name())
}

func (a *DataFlowDeptTrigger) ParameterNew() interface{} {
	return &DataFlowDeptTriggerParam{}
}

type DataFlowTagTrigger struct {
}

type DataFlowTagTriggerParam struct {
}

func (a *DataFlowTagTrigger) Name() string {
	return common.DataflowTagTrigger
}

func (a *DataFlowTagTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	return triggerManual(ctx, params, token, a.Name())
}

func (a *DataFlowTagTrigger) ParameterNew() interface{} {
	return &DataFlowTagTriggerParam{}
}

type TriggerOperator struct {
	Operator string                `json:"operator"`
	Version  string                `json:"version"`
	Request  *ComboOperatorRequest `json:"request"`
}

func (a *TriggerOperator) Name() string {
	return common.OperatorTrigger
}

func (a *TriggerOperator) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	log := traceLog.WithContext(ctx.Context())

	input := params.(*TriggerOperator)
	taskIns := ctx.GetTaskInstance()

	if !strings.HasPrefix(input.Operator, a.Name()) {
		err = fmt.Errorf("invalid operator %s", input.Operator)
		log.Warnf("[TriggerOperator] err %s, taskId %s", err.Error(), taskIns.ID)
		return nil, err
	}

	operatorID := strings.TrimPrefix(input.Operator, common.TriggerOperatorPrefix)

	userInfo := &drivenadapters.UserInfo{}
	if token != nil {
		userInfo.UserID = token.UserID
		userInfo.AccountType = common.User.ToString()
		if token.IsApp {
			userInfo.AccountType = common.APP.ToString()
		}
	}

	agentOperatorIntegration := drivenadapters.NewAgentOperatorIntegration()
	operator, err := agentOperatorIntegration.GetOperatorInfo(ctx.Context(), operatorID, input.Version, ctx.GetTaskInstance().RelatedDagInstance.BizDomainID, userInfo)

	if err != nil {
		log.Warnf("[TriggerOperator] GetOperatorInfo err %s, taskId %s", err.Error(), taskIns.ID)
		return nil, err
	}

	if !operator.OperatorInfo.IsDataSource {
		log.Warnf("[TriggerOperator] operator is not datasource, id %s", operatorID)
		return nil, fmt.Errorf("operator is not datasource, id %s", operatorID)
	}

	if operator.OperatorInfo.ExecutionMode == "async" {
		log.Warnf("[TriggerOperator] async operator is not supported, operator_id: %s", operatorID)
		return nil, fmt.Errorf("async operator is not supported, operator_id: %s", operatorID)
	}

	data, err := callOperator(
		ctx.Context(),
		taskIns,
		token,
		operator,
		input.Request.Parameters,
		input.Request.Body,
		false,
	)

	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{"data": data}
	result["_type"] = "operator"
	result["name"] = operator.Name
	defer func() {
		patchDagInstanceSource(ctx, result, a.Name())
	}()
	ctx.ShareData().Set(ctx.GetTaskID(), result)
	return result, nil
}

func (a *TriggerOperator) ParameterNew() interface{} {
	return &TriggerOperator{}
}
