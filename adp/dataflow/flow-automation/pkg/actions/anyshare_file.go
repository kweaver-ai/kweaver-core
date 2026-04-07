package actions

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

const (
	// anyshareFile 文件操作
	anyshareFile = "@anyshare/file"
	// anyshareFileCopy 复制文件
	anyshareFileCopy = "copy"
	// anyshareFileRemove 删除文件
	anyshareFileRemove = "remove"
	// anyshareFileMove 移动文件
	anyshareFileMove = "move"
	// anyshareFileRename 重命名文件
	anyshareFileRename = "rename"
	// anyshareFileTag 文件添加标签
	anyshareFileTag = "addtag"
	// anyshareFileGetpath 获取文件路径
	anyshareFileGetpath = "getpath"
	// anyshareMatchContent 匹配文件内容
	anyshareFileMatchContent = "matchcontent"
	// anyshareFileSetCsfLevel 设置文件密级
	anyshareFileSetCsfLevel = "setcsflevel"
	// anyshareFileSetTemplate 设置文件编目
	anyshareFileSetTemplate = "settemplate"
	// anyshareFileGetPage 获取Word/PDF文件页数
	anyshareFileGetPage = "getpage"
	// anyshareFileSetPerm 设置文件权限
	anyshareFileSetPerm = "perm"
	// anyshareFileCreate 新建文件
	anyshareFileCreate = "create"
	// anyshareFileEdit 更新文件
	anyshareFileEdit = "edit"
	// anyshareExcelFileEdit 编辑excel文件
	anyshareExcelFileEdit = "editexcel"
	// anyshareDocxFileEdit 编辑docx文件
	anyshareDocxFileEdit = "editdocx"
	// 获取文件信息
	anyshareFileStat = "stat"
)

// MaxFileSize 最大能处理的文件大小
var MaxFileSize int64 = 100 * 1024 * 1024

// AnyShareFileCopy 文件复制操作
type AnyShareFileCopy struct {
	DocID      string `json:"docid"`
	DocName    string `json:"name"`
	DestParent string `json:"destparent"`
	OnDup      int    `json:"ondup"`
}

// Name 操作名称
func (a *AnyShareFileCopy) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileCopy)
}

// Run 操作方法
func (a *AnyShareFileCopy) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareFileCopy)
	doc, err := ctx.NewASDoc().CopyFile(ctx.Context(), input.DocID, input.DestParent, input.OnDup, token.Token, token.LoginIP)
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
func (a *AnyShareFileCopy) ParameterNew() interface{} {
	return &AnyShareFileCopy{}
}

// AnyShareFileMove 文件移动操作
type AnyShareFileMove struct {
	DocID      string `json:"docid"`
	DocName    string `json:"name"`
	DestParent string `json:"destparent"`
	OnDup      int    `json:"ondup"`
}

// Name 操作名称
func (a *AnyShareFileMove) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileMove)
}

// Run 操作方法
func (a *AnyShareFileMove) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareFileMove)
	// 移动文件
	doc, err := ctx.NewASDoc().MoveFile(ctx.Context(), input.DocID, input.DestParent, input.OnDup, token.Token, token.LoginIP)
	if err != nil {
		return doc, err
	}

	data, err := handleDocMove(ctx, doc["docid"], token)
	if err != nil {
		return nil, err
	}
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, err
}

// ParameterNew 初始化参数
func (a *AnyShareFileMove) ParameterNew() interface{} {
	return &AnyShareFileMove{}
}

// AnyShareFileRemove 文件删除操作
type AnyShareFileRemove struct {
	DocID string `json:"docid"`
}

// Name 操作名称
func (a *AnyShareFileRemove) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileRemove)
}

// Run 操作方法
func (a *AnyShareFileRemove) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareFileRemove)
	data, err := handleDocRemove(ctx, input.DocID, token)
	if err != nil {
		return nil, err
	}
	err = ctx.NewASDoc().DeleteFile(ctx.Context(), input.DocID, token.Token, token.LoginIP)

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
func (a *AnyShareFileRemove) ParameterNew() interface{} {
	return &AnyShareFileRemove{}
}

// AnyShareFileRename 文件重命名操作
type AnyShareFileRename struct {
	DocID   string `json:"docid"`
	DocName string `json:"name"`
	OnDup   int    `json:"ondup"`
}

// Name 操作名称
func (a *AnyShareFileRename) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileRename)
}

// Run 操作方法
func (a *AnyShareFileRename) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareFileRename)

	data, err := handleDocRename(ctx, input.DocID, token)
	if err != nil {
		return nil, err
	}

	name, ok := data["name"].(string)
	if !ok {
		name = input.DocName
	}
	ext := path.Ext(name)

	doc, err := ctx.NewASDoc().RenameFile(ctx.Context(), input.DocID, fmt.Sprintf("%s%s", input.DocName, ext), input.OnDup, token.Token, token.LoginIP)
	newName := doc.Name
	if newName == "" {
		newName = fmt.Sprintf("%s%s", input.DocName, ext)
	}
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return doc, err
	}

	data["name"] = newName
	parts := strings.Split(data["path"].(string), "/")
	data["path"] = strings.Join(append(parts[:len(parts)-1], newName), "/")

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)

	ctx.Trace(ctx.Context(), "run end")
	return data, err
}

// ParameterNew 初始化参数
func (a *AnyShareFileRename) ParameterNew() interface{} {
	return &AnyShareFileRename{}
}

// AnyShareFilePath 文件获取路径操作
type AnyShareFilePath struct {
	DocID  string `json:"docid"`
	Order  string `json:"order"`
	Depth  int    `json:"depth"`
	Custom int    `json:"custom"`
}

// Name 操作名称
func (a *AnyShareFilePath) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileGetpath)
}

// Run 操作方法
func (a *AnyShareFilePath) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareFilePath)
	doc, _, err := getDocInfo(ctx.Context(), input.DocID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return doc, err
	}
	var docpath = getDocPath(input.Order, input.Depth, doc)

	data := map[string]string{"docid": doc["id"].(string), "path": docpath}
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)

	ctx.Trace(ctx.Context(), "run end")
	return data, err
}

// ParameterNew 初始化参数
func (a *AnyShareFilePath) ParameterNew() interface{} {
	return &AnyShareFilePath{}
}

// AnyShareFileTag 文件打标签操作
type AnyShareFileTag struct {
	DocID string      `json:"docid"`
	Tags  interface{} `json:"tags"`
}

// Name 操作名称
func (a *AnyShareFileTag) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileTag)
}

// Run 操作方法
func (a *AnyShareFileTag) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareFileTag)
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
func (a *AnyShareFileTag) ParameterNew() interface{} {
	return &AnyShareFileTag{}
}

// AnyShareFileSetCsfLevel 文件打标签操作
type AnyShareFileSetCsfLevel struct {
	DocID    string `json:"docid"`
	CsfLevel int    `json:"csf_level"`
}

// Name 操作名称
func (a *AnyShareFileSetCsfLevel) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileSetCsfLevel)
}

// Run 操作方法
func (a *AnyShareFileSetCsfLevel) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareFileSetCsfLevel)

	result, err := ctx.NewASDoc().SetCsfLevel(ctx.Context(), input.DocID, input.CsfLevel, token.Token, token.LoginIP)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	var res = map[string]interface{}{
		"docid":     input.DocID,
		"csf_level": input.CsfLevel,
		"result":    result,
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, res)

	ctx.Trace(ctx.Context(), "run end")
	return res, err
}

// ParameterNew 初始化参数
func (a *AnyShareFileSetCsfLevel) ParameterNew() interface{} {
	return &AnyShareFileSetCsfLevel{}
}

// AnyShareFileMatchContent 文件打标签操作
type AnyShareFileMatchContent struct {
	DocID     string `json:"docid"`
	MatchType string `json:"matchtype"`
	Keyword   string `json:"keyword"`
	Reg       string `json:"reg"`
}

// Name 操作名称
func (a *AnyShareFileMatchContent) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileMatchContent)
}

// Run 操作方法
func (a *AnyShareFileMatchContent) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) { //nolint
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareFileMatchContent)
	id := ctx.GetTaskID()
	var result = map[string]interface{}{
		"docid":      input.DocID,
		"is_match":   false,
		"match_nums": 0,
	}

	res, err := ctx.NewASDoc().InnerOSDownload(ctx.Context(), input.DocID, "")
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}
	// 文件大小超过100M，跳过处理
	if res.Size > MaxFileSize {
		err = ierrors.NewIError(ierrors.FileSizeExceed, "", map[string]interface{}{
			"doc": map[string]interface{}{
				"docid": input.DocID,
				"limit": MaxFileSize,
			},
		})
		ctx.Trace(ctx.Context(), err.Error())
		return result, err
	}

	client := drivenadapters.NewRawHTTPClient()
	client.Timeout = 60 * time.Second
	resp, err := client.Get(res.URL)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			traceLog.WithContext(ctx.Context()).Errorln(closeErr)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respCode := resp.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		err = ierrors.ExHTTPError{
			Body:   string(body),
			Status: respCode,
		}
		return nil, err
	}

	if len(body) == 0 {
		result["match_nums"] = 0
		result["is_match"] = false
		ctx.ShareData().Set(id, result)
		ctx.Trace(ctx.Context(), "run end")
		return result, err
	}

	tikaClient := drivenadapters.NewTika()
	parsedCon, err := tikaClient.ParseContent(ctx.Context(), res.Name, &body)
	if err != nil {
		return nil, err
	}

	var isFastTextAnalisysOK = true
	err = tikaClient.CheckFastTextAnalysys(ctx.Context())

	if err != nil {
		// fate-text-analysis服务不可用时，自行进行验证
		isFastTextAnalisysOK = false
	}

	if !isFastTextAnalisysOK && input.MatchType == "KEYWORD" {
		count := strings.Count(string(*parsedCon), input.Keyword)
		result["is_match"] = false
		if count > 0 {
			result["is_match"] = true
		}

		result["match_nums"] = count

		ctx.ShareData().Set(id, result)

		ctx.Trace(ctx.Context(), "run end")
		return result, nil
	}

	if !isFastTextAnalisysOK && input.MatchType == "REG" {
		reg, _ := regexp.Compile(input.Reg)
		macthed := reg.FindAllString(string(*parsedCon), -1)

		result["match_nums"] = len(macthed)
		result["is_match"] = false
		if len(macthed) > 0 {
			result["is_match"] = true
		}

		ctx.ShareData().Set(id, result)

		ctx.Trace(ctx.Context(), "run end")
		return result, nil
	}

	var methodMap = map[string]string{
		"KEYWORD": "KWD",
		"REG":     "REG",
	}
	var tpl = make(map[string]interface{})

	tpl["name"] = input.MatchType
	tpl["expression"] = input.Reg
	tpl["auxiliary_words"] = []string{input.Keyword}
	method, ok := methodMap[input.MatchType]
	if ok {
		tpl["method"] = method
	} else {
		tpl["method"] = "KWD"
	}

	matchedRes, err := tikaClient.MatchContent(ctx.Context(), parsedCon, tpl)

	if err != nil {
		return nil, err
	}

	if matchedRes.HasPrivateInfo {
		results, ok := matchedRes.Results[input.MatchType]
		if ok {
			result["is_match"] = true
			result["match_nums"] = len(results.Info)
		}
	}
	ctx.ShareData().Set(id, result)

	ctx.Trace(ctx.Context(), "run end")
	return result, err
}

// ParameterNew 初始化参数
func (a *AnyShareFileMatchContent) ParameterNew() interface{} {
	return &AnyShareFileMatchContent{}
}

// AnyShareFileSetTemplate 文件夹设置编目操作
type AnyShareFileSetTemplate struct {
}

// AnyShareSetTemplateParam 文件夹设置编目操作参数
type AnyShareSetTemplateParam struct {
	DocID string      `json:"docid"`
	Tpls  interface{} `json:"templates"`
}

// Name 操作名称
func (a *AnyShareFileSetTemplate) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileSetTemplate)
}

// Run 操作方法
func (a *AnyShareFileSetTemplate) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
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
func (a *AnyShareFileSetTemplate) ParameterNew() interface{} {
	return &AnyShareSetTemplateParam{}
}

// AnyShareFileGetPage 获取Word/PDF页数
type AnyShareFileGetPage struct {
}

// AnyShareFileGetPageParam 文件夹设置编目操作参数
type AnyShareFileGetPageParam struct {
	DocID string `json:"docid"`
}

// Name 操作名称
func (a *AnyShareFileGetPage) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileGetPage)
}

// Run 操作方法
func (a *AnyShareFileGetPage) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareFileGetPageParam)
	id := ctx.GetTaskID()

	var result = map[string]interface{}{
		"docid":     input.DocID,
		"page_nums": 0,
	}

	res, err := ctx.NewASDoc().InnerOSDownload(ctx.Context(), input.DocID, "")
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}
	// 文件大小超过100M，跳过处理
	if res.Size > MaxFileSize {
		err = ierrors.NewIError(ierrors.FileSizeExceed, "", map[string]interface{}{
			"doc": map[string]interface{}{
				"docid": input.DocID,
				"limit": MaxFileSize,
			},
		})
		ctx.Trace(ctx.Context(), err.Error())
		return result, err
	}

	client := drivenadapters.NewRawHTTPClient()
	client.Timeout = 60 * time.Second
	resp, err := client.Get(res.URL)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			traceLog.WithContext(ctx.Context()).Errorln(closeErr)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respCode := resp.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		err = ierrors.ExHTTPError{
			Body:   string(body),
			Status: respCode,
		}
		return nil, err
	}

	if len(body) == 0 {
		ctx.ShareData().Set(id, result)
		ctx.Trace(ctx.Context(), "run end")
		return result, err
	}

	tikaClient := drivenadapters.NewTika()
	docMetadata, err := tikaClient.ParseMetadata(ctx.Context(), res.Name, &body)
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	if docMetadata.Page == "" {
		err = ierrors.NewIError(ierrors.NotContainPageData, "", map[string]interface{}{
			"doc": map[string]interface{}{
				"docid": input.DocID,
			},
		})
		ctx.Trace(ctx.Context(), err.Error())
		return result, err
	}

	num, err := strconv.Atoi(docMetadata.Page)
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return result, err
	}

	result["page_nums"] = num

	ctx.ShareData().Set(id, result)

	ctx.Trace(ctx.Context(), "run end")
	return result, nil
}

// ParameterNew 初始化参数
func (a *AnyShareFileGetPage) ParameterNew() interface{} {
	return &AnyShareFileGetPageParam{}
}

// AnyshareFileSetPerm 权限配置节点参数
type AnyshareFileSetPerm struct {
	DocID         string    `json:"docid"`
	ConfigInherit bool      `json:"config_inherit"`
	Inherit       bool      `json:"inherit"`
	Perminfos     PermInfos `json:"perminfos"`
	AppID         string    `json:"appid"`
	AppPwd        string    `json:"apppwd"`
}

// Name 操作名称
func (a *AnyshareFileSetPerm) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileSetPerm)
}

// Run 操作方法
func (a *AnyshareFileSetPerm) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyshareFileSetPerm)
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
	return result, err
}

// ParameterNew 初始化参数
func (a *AnyshareFileSetPerm) ParameterNew() interface{} {
	return &AnyshareFileSetPerm{}
}

// AnyshareCreateFile 新建文件节点参数
type AnyshareCreateFile struct {
	Type       string      `json:"type"`
	DocName    string      `json:"name"`
	DocID      string      `json:"docid"`
	OnDup      int         `json:"ondup"`
	NewType    string      `json:"new_type"`
	SourceType string      `json:"source_type"`
	Content    interface{} `json:"content"`
}

// Name 操作名称
func (a *AnyshareCreateFile) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileCreate)
}

// Run run
func (a *AnyshareCreateFile) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	newCtx = context.WithValue(newCtx, common.Authorization, token.Token)

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyshareCreateFile)
	id := ctx.GetTaskID()

	a.SourceType = utils.IfNot(a.SourceType == "", "full_text", a.SourceType)

	codeRunnerAdapter := drivenadapters.NewCodeRunner()
	reqParms := drivenadapters.CreateFileReq{
		FileType:   input.Type,
		Name:       input.DocName,
		Docid:      input.DocID,
		Ondup:      input.OnDup,
		NewType:    input.NewType,
		SourceType: input.SourceType,
		Content:    input.Content,
	}
	docID, err := codeRunnerAdapter.CreateFile(ctx.Context(), reqParms)
	if err != nil {
		return nil, err
	}

	doc, err := ctx.NewASDoc().GetDocMsg(ctx.Context(), docID)
	if err != nil {
		return nil, err
	}
	var result = map[string]interface{}{}
	result["docid"] = docID
	result["name"] = doc.Name
	result["path"] = doc.Path
	result["create_time"] = TimeToISOString(doc.CreateTime, TimeUnitMicrosecond)
	result["creator"] = doc.Creator
	result["modify_time"] = TimeToISOString(doc.Modified, TimeUnitMicrosecond)
	result["editor"] = doc.Editor

	ctx.ShareData().Set(id, result)

	ctx.Trace(ctx.Context(), "run end")
	return result, err
}

// ParameterNew 初始化参数
func (a *AnyshareCreateFile) ParameterNew() interface{} {
	return &AnyshareCreateFile{}
}

// AnyshareDocxFileUpdate 更新文件节点参数
type AnyshareFileUpdate struct {
	Type       string      `json:"type"`
	DocID      string      `json:"docid"`
	NewType    string      `json:"new_type"`
	InsertType string      `json:"insert_type"`
	InsertPos  int         `json:"insert_pos"`
	Content    interface{} `json:"content"`
}

// Name 操作名称
func (a *AnyshareFileUpdate) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileEdit)
}

// Run run
func (a *AnyshareFileUpdate) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	newCtx = context.WithValue(newCtx, common.Authorization, token.Token)

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyshareFileUpdate)
	id := ctx.GetTaskID()

	codeRunnerAdapter := drivenadapters.NewCodeRunner()
	reqParams := drivenadapters.UpdateFileReq{
		FileType:   input.Type,
		DocID:      input.DocID,
		InsertType: input.InsertType,
		Content:    input.Content,
	}

	if input.Type == "xlsx" {
		reqParams.NewType = input.NewType
		reqParams.InsertPos = input.InsertPos
	}

	docID, err := codeRunnerAdapter.UpdateFile(ctx.Context(), reqParams)
	if err != nil {
		return nil, err
	}

	doc, err := ctx.NewASDoc().GetDocMsg(ctx.Context(), docID)
	if err != nil {
		return nil, err
	}
	var result = map[string]interface{}{}
	result["docid"] = docID
	result["name"] = doc.Name
	result["path"] = doc.Path
	result["create_time"] = TimeToISOString(doc.CreateTime, TimeUnitMicrosecond)
	result["creator"] = doc.Creator
	result["modify_time"] = TimeToISOString(doc.Modified, TimeUnitMicrosecond)
	result["editor"] = doc.Editor

	ctx.ShareData().Set(id, result)

	ctx.Trace(ctx.Context(), "run end")

	return result, nil
}

// ParameterNew 初始化参数
func (a *AnyshareFileUpdate) ParameterNew() interface{} {
	return &AnyshareFileUpdate{}
}

// AnyshareExcelFileUpdate 更新excel文件节点参数
type AnyshareExcelFileUpdate struct {
	DocID      string      `json:"docid"`
	NewType    string      `json:"new_type"`
	InsertType string      `json:"insert_type"`
	InsertPos  int         `json:"insert_pos"`
	Content    interface{} `json:"content"`
}

// Name 操作名称
func (a *AnyshareExcelFileUpdate) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareExcelFileEdit)
}

// Run run
func (a *AnyshareExcelFileUpdate) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	newCtx = context.WithValue(newCtx, common.Authorization, token.Token)

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyshareExcelFileUpdate)
	id := ctx.GetTaskID()

	codeRunnerAdapter := drivenadapters.NewCodeRunner()
	reqParams := drivenadapters.UpdateFileReq{
		FileType:   "xlsx",
		DocID:      input.DocID,
		NewType:    input.NewType,
		InsertType: input.InsertType,
		InsertPos:  input.InsertPos,
		Content:    input.Content,
	}
	docID, err := codeRunnerAdapter.UpdateFile(ctx.Context(), reqParams)
	if err != nil {
		return nil, err
	}

	doc, err := ctx.NewASDoc().GetDocMsg(ctx.Context(), docID)
	if err != nil {
		return nil, err
	}
	var result = map[string]interface{}{}
	result["docid"] = docID
	result["name"] = doc.Name
	result["path"] = doc.Path
	result["create_time"] = TimeToISOString(doc.CreateTime, TimeUnitMicrosecond)
	result["creator"] = doc.Creator
	result["modify_time"] = TimeToISOString(doc.Modified, TimeUnitMicrosecond)
	result["editor"] = doc.Editor

	ctx.ShareData().Set(id, result)

	ctx.Trace(ctx.Context(), "run end")

	return result, nil
}

// ParameterNew 初始化参数
func (a *AnyshareExcelFileUpdate) ParameterNew() interface{} {
	return &AnyshareExcelFileUpdate{}
}

// AnyshareDocxFileUpdate 更新docx文件节点参数
type AnyshareDocxFileUpdate struct {
	DocID      string      `json:"docid"`
	InsertType string      `json:"insert_type"`
	Content    interface{} `json:"content"`
}

// Name 操作名称
func (a *AnyshareDocxFileUpdate) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareDocxFileEdit)
}

// Run run
func (a *AnyshareDocxFileUpdate) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	newCtx = context.WithValue(newCtx, common.Authorization, token.Token)

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyshareDocxFileUpdate)
	id := ctx.GetTaskID()

	codeRunnerAdapter := drivenadapters.NewCodeRunner()
	reqParams := drivenadapters.UpdateFileReq{
		FileType:   "docx",
		DocID:      input.DocID,
		InsertType: input.InsertType,
		Content:    input.Content,
	}
	docID, err := codeRunnerAdapter.UpdateFile(ctx.Context(), reqParams)
	if err != nil {
		return nil, err
	}

	doc, err := ctx.NewASDoc().GetDocMsg(ctx.Context(), docID)
	if err != nil {
		return nil, err
	}
	var result = map[string]interface{}{}
	result["docid"] = docID
	result["name"] = doc.Name
	result["path"] = doc.Path
	result["create_time"] = TimeToISOString(doc.CreateTime, TimeUnitMicrosecond)
	result["creator"] = doc.Creator
	result["modify_time"] = TimeToISOString(doc.Modified, TimeUnitMicrosecond)
	result["editor"] = doc.Editor

	ctx.ShareData().Set(id, result)

	ctx.Trace(ctx.Context(), "run end")

	return result, nil
}

// ParameterNew 初始化参数
func (a *AnyshareDocxFileUpdate) ParameterNew() interface{} {
	return &AnyshareDocxFileUpdate{}
}

type AnyShareFileRelevance struct {
}

func (a *AnyShareFileRelevance) Name() string {
	return common.OpAnyShareFileRelevance
}

func (a *AnyShareFileRelevance) ParameterNew() interface{} {
	return &AnyShareRelevanceParams{}
}

func (a *AnyShareFileRelevance) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {

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

type AnyShareFileGetByName struct {
	DocID    string `json:"docid"`
	Filename string `json:"name"`
}

func (a *AnyShareFileGetByName) Name() string {
	return common.OpAnyShareFileGetByName
}

func (a *AnyShareFileGetByName) ParameterNew() interface{} {
	return &AnyShareFileGetByName{}
}

func (a *AnyShareFileGetByName) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {

	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	input := params.(*AnyShareFileGetByName)

	efast := ctx.NewASDoc()

	files, _, err := efast.ListDir(ctx.Context(), input.DocID, token.Token, token.LoginIP)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	var result = map[string]interface{}{
		"docid":       "",
		"name":        "",
		"path":        "",
		"create_time": 0,
		"creator":     "",
		"modify_time": 0,
		"editor":      "",
	}

	for _, item := range files {
		if file, ok := item.(map[string]interface{}); ok {
			if file["name"] == input.Filename {
				docID := file["docid"].(string)
				doc, err := ctx.NewASDoc().GetDocMsg(ctx.Context(), docID)
				if err != nil {
					return nil, err
				}
				result["docid"] = docID
				result["name"] = doc.Name
				result["path"] = doc.Path
				result["create_time"] = TimeToISOString(doc.CreateTime, TimeUnitMicrosecond)
				result["creator"] = doc.Creator
				result["modify_time"] = TimeToISOString(doc.Modified, TimeUnitMicrosecond)
				result["editor"] = doc.Editor
				break
			}
		}
	}

	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, result)
	ctx.Trace(ctx.Context(), "run end")

	return result, nil
}

type AnyShareFileStat struct {
	DocID string `json:"docid"`
}

func (a *AnyShareFileStat) Name() string {
	return fmt.Sprintf("%s/%s", anyshareFile, anyshareFileStat)
}

// Run 操作方法
func (a *AnyShareFileStat) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareFilePath)
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
func (a *AnyShareFileStat) ParameterNew() interface{} {
	return &AnyShareFileStat{}
}
