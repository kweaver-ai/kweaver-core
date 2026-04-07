package actions

import (
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

const templateNotExists = 404003032

const (
	anyshareFolder            = "@anyshare/folder"
	anyshareFolderCopy        = "copy"
	anyshareFolderRemove      = "remove"
	anyshareFolderMove        = "move"
	anyshareFolderRename      = "rename"
	anyshareFolderTag         = "addtag"
	anyshareFolderCreate      = "create"
	anyshareFolderGetpath     = "getpath"
	anyshareFolderSetTemplate = "settemplate"
	anyshareFolderSetPerm     = "perm"
	anyshareFolderStat        = "stat"
)

// AnyShareDirCopy 文件夹复制操作
type AnyShareDirCopy struct {
	DocID      string `json:"docid"`
	DocName    string `json:"name"`
	DestParent string `json:"destparent"`
	OnDup      int    `json:"ondup"`
}

// Name 操作名称
func (a *AnyShareDirCopy) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFolder, anyshareFolderCopy)
}

// Run 操作方法
func (a *AnyShareDirCopy) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	input := params.(*AnyShareDirCopy)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	doc, err := ctx.NewASDoc().CopyDir(ctx.Context(), input.DocID, input.DestParent, input.OnDup, token.Token, token.LoginIP)
	if err != nil {
		return nil, err
	}
	data, err := handleDocCopy(ctx, doc["docid"], token)
	if err != nil {
		return nil, err
	}
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, err
}

// ParameterNew 初始化参数
func (a *AnyShareDirCopy) ParameterNew() interface{} {
	return &AnyShareDirCopy{}
}

// AnyShareDirMove 文件夹移动操作
type AnyShareDirMove struct {
	DocID      string `json:"docid"`
	DocName    string `json:"name"`
	DestParent string `json:"destparent"`
	OnDup      int    `json:"ondup"`
}

// Name 操作名称
func (a *AnyShareDirMove) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFolder, anyshareFolderMove)
}

// Run 操作方法
func (a *AnyShareDirMove) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDirMove)
	// 移动文件
	doc, err := ctx.NewASDoc().MoveDir(ctx.Context(), input.DocID, input.DestParent, input.OnDup, token.Token, token.LoginIP)
	if err != nil {
		return doc, err
	}
	attr, _, err := getDocInfo(ctx.Context(), doc["docid"], token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, attr)
	ctx.Trace(ctx.Context(), "run end")
	return attr, err
}

// ParameterNew 初始化参数
func (a *AnyShareDirMove) ParameterNew() interface{} {
	return &AnyShareDirMove{}
}

// AnyShareDirRemove 文件夹删除操作
type AnyShareDirRemove struct {
	DocID string `json:"docid"`
}

// Name 操作名称
func (a *AnyShareDirRemove) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFolder, anyshareFolderRemove)
}

// Run 操作方法
func (a *AnyShareDirRemove) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDirRemove)
	data, err := handleDocRemove(ctx, input.DocID, token)
	if err != nil {
		return nil, err
	}
	err = ctx.NewASDoc().DeleteDir(ctx.Context(), input.DocID, token.Token, token.LoginIP)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, err
}

// ParameterNew 初始化参数
func (a *AnyShareDirRemove) ParameterNew() interface{} {
	return &AnyShareDirRemove{}
}

// AnyShareDirRename 文件夹重命名操作
type AnyShareDirRename struct {
	DocID   string `json:"docid"`
	DocName string `json:"name"`
	OnDup   int    `json:"ondup"`
}

// Name 操作名称
func (a *AnyShareDirRename) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFolder, anyshareFolderRename)
}

// Run 操作方法
func (a *AnyShareDirRename) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDirRename)
	doc, err := ctx.NewASDoc().RenameDir(ctx.Context(), input.DocID, input.DocName, input.OnDup, token.Token, token.LoginIP)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return doc, err
	}

	data, err := handleDocRename(ctx, input.DocID, token)
	if err != nil {
		return nil, err
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)

	ctx.Trace(ctx.Context(), "run end")
	return data, err
}

// ParameterNew 初始化参数
func (a *AnyShareDirRename) ParameterNew() interface{} {
	return &AnyShareDirRename{}
}

// AnyShareDirCreate 文件夹新建操作
type AnyShareDirCreate struct {
	DocID   string `json:"docid"`
	DocName string `json:"name"`
	OnDup   int    `json:"ondup"`
}

// Name 操作名称
func (a *AnyShareDirCreate) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFolder, anyshareFolderCreate)
}

// Run 操作方法
func (a *AnyShareDirCreate) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDirCreate)
	doc, err := ctx.NewASDoc().CreateDir(ctx.Context(), input.DocID, input.DocName, input.OnDup, token.Token, token.LoginIP)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return doc, err
	}

	attr, err := ctx.NewASDoc().ConvertPath(ctx.Context(), doc.DocID, token.Token, token.LoginIP)
	if err != nil {
		return nil, err
	}

	id := ctx.GetTaskID()
	data := map[string]interface{}{
		"docid":       doc.DocID,
		"name":        doc.Name,
		"editor":      doc.Editor,
		"modified":    TimeToISOString(doc.Modified, TimeUnitMicrosecond),
		"creator":     doc.Creator,
		"create_time": TimeToISOString(doc.CreateTime, TimeUnitMicrosecond),
		"path":        attr.Path,
	}
	ctx.ShareData().Set(id, data)

	ctx.Trace(ctx.Context(), "run end")

	result := map[string]interface{}{
		"docid":       doc.DocID,
		"name":        doc.Name,
		"creator":     doc.Creator,
		"create_time": TimeToISOString(doc.CreateTime, TimeUnitMicrosecond),
	}
	return result, err
}

// ParameterNew 初始化参数
func (a *AnyShareDirCreate) ParameterNew() interface{} {
	return &AnyShareDirCreate{}
}

// AnyShareDirTag 文件夹打标签操作
type AnyShareDirTag struct {
	DocID string      `json:"docid"`
	Tags  interface{} `json:"tags"`
}

// Name 操作名称
func (a *AnyShareDirTag) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFolder, anyshareFolderTag)
}

// Run 操作方法
func (a *AnyShareDirTag) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDirTag)
	curID := utils.GetDocCurID(input.DocID)
	tags := getTags(input.Tags)

	res, err := ctx.NewASDoc().SetTag(ctx.Context(), curID, tags, token.Token, token.LoginIP)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	ctx.Trace(ctx.Context(), "run end")
	return res, err
}

// ParameterNew 初始化参数
func (a *AnyShareDirTag) ParameterNew() interface{} {
	return &AnyShareDirTag{}
}

// AnyShareDirPath 文件获取路径操作
type AnyShareDirPath struct {
	DocID  string `json:"docid"`
	Order  string `json:"order"`
	Depth  int    `json:"depth"`
	Custom int    `json:"custom"`
}

// Name 操作名称
func (a *AnyShareDirPath) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFolder, anyshareFolderGetpath)
}

// Run 操作方法
func (a *AnyShareDirPath) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDirPath)
	doc, _, err := getDocInfo(ctx.Context(), input.DocID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return doc, err
	}
	var path = getDocPath(input.Order, input.Depth, doc)

	id := ctx.GetTaskID()
	data := map[string]string{"docid": doc["id"].(string), "path": path}
	ctx.ShareData().Set(id, data)

	ctx.Trace(ctx.Context(), "run end")
	return data, err
}

// ParameterNew 初始化参数
func (a *AnyShareDirPath) ParameterNew() interface{} {
	return &AnyShareDirPath{}
}

// AnyShareDirSetTemplate 文件夹设置编目操作
type AnyShareDirSetTemplate struct {
	DocID string                            `json:"docid"`
	Tpls  map[string]map[string]interface{} `json:"templates"`
}

// Name 操作名称
func (a *AnyShareDirSetTemplate) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFolder, anyshareFolderSetTemplate)
}

// Run 操作方法
func (a *AnyShareDirSetTemplate) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareSetTemplateParam)

	res, err := handleSetTemplate(ctx, input, token.Token, token.LoginIP)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	ctx.Trace(ctx.Context(), "run end")
	return res, err
}

// ParameterNew 初始化参数
func (a *AnyShareDirSetTemplate) ParameterNew() interface{} {
	return &AnyShareSetTemplateParam{}
}

type AnyShareFolderRelevance struct {
}

func (a *AnyShareFolderRelevance) Name() string {
	return common.OpAnyShareFolderRelevance
}

func (a *AnyShareFolderRelevance) ParameterNew() interface{} {
	return &AnyShareRelevanceParams{}
}

func (a *AnyShareFolderRelevance) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {

	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	input := params.(*AnyShareRelevanceParams)

	err = handleAddRelevance(ctx.Context(), input, token.Token, token.LoginIP)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
	}

	ctx.Trace(ctx.Context(), "run end")
	return nil, err
}

// AnyshareFolderSetPerm 权限配置节点参数
type AnyshareFolderSetPerm struct {
	DocID         string    `json:"docid"`
	ConfigInherit bool      `json:"config_inherit"`
	Inherit       bool      `json:"inherit"`
	Perminfos     PermInfos `json:"perminfos"`
	AppID         string    `json:"appid"`
	AppPwd        string    `json:"apppwd"`
}

// Name 操作名称
func (a *AnyshareFolderSetPerm) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFolder, anyshareFolderSetPerm)
}

// Run 操作方法
func (a *AnyshareFolderSetPerm) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyshareFolderSetPerm)
	id := ctx.GetTaskID()

	tokenStr := token.Token
	if input.AppID != "" && input.AppPwd != "" {
		appToken, _, err := drivenadapters.NewAppTokenMgnt(input.AppID, input.AppPwd, false).GetAppToken("")
		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}
		tokenStr = appToken
	}

	docShare := drivenadapters.NewDocShare()
	currentPerms, err := docShare.GetPerm(ctx.Context(), input.DocID, token.Token)
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	permInfoMap := make(PermInfoMap)
	for _, permInfo := range currentPerms.Perminfos {
		if permInfo.InheritDocID == "" {
			config, err := permInfo.ToPermConfig()
			if err != nil {
				return nil, err
			}
			permInfoMap[permInfo.AccessorID] = *config
		}
	}

	if !input.ConfigInherit {
		input.Inherit = currentPerms.Inherit
	}

	permInfoParams := make([]drivenadapters.PermConfig, 0)
	err = input.Perminfos.Build(permInfoMap)
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	for _, permInfo := range permInfoMap {
		permInfoParams = append(permInfoParams, permInfo)
	}

	permConfig := &drivenadapters.DocPermConfig{
		Configs:     permInfoParams,
		Inherit:     input.Inherit,
		SendMessage: true,
	}
	_, err = docShare.SetDocPerm2(ctx.Context(), permConfig, tokenStr, input.DocID)
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	var result = map[string]interface{}{
		"result": 1,
	}

	ctx.ShareData().Set(id, result)

	ctx.Trace(ctx.Context(), "run end")
	return nil, err
}

// ParameterNew 初始化参数
func (a *AnyshareFolderSetPerm) ParameterNew() interface{} {
	return &AnyshareFolderSetPerm{}
}

type AnyShareDirStat struct {
	DocID string `json:"docid"`
}

// Name 操作名称
func (a *AnyShareDirStat) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFolder, anyshareFolderStat)
}

// Run 操作方法
func (a *AnyShareDirStat) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDirStat)
	doc, _, err := getDocInfo(ctx.Context(), input.DocID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return doc, err
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, doc)

	ctx.Trace(ctx.Context(), "run end")
	return doc, err
}

// ParameterNew 初始化参数
func (a *AnyShareDirStat) ParameterNew() interface{} {
	return &AnyShareDirStat{}
}
