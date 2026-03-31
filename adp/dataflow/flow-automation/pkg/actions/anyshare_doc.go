package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	errs "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

type AnyShareDocRename struct {
	DocID      string `json:"docid"`
	DocName    string `json:"name"`
	DestParent string `json:"destparent"`
	OnDup      int    `json:"ondup"`
}

func (*AnyShareDocRename) Name() string {
	return common.OpAnyShareDocRename
}

func (*AnyShareDocRename) ParameterNew() interface{} {
	return &AnyShareDocRename{}
}

// Run implements entity.Action.
func (*AnyShareDocRename) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	input := params.(*AnyShareDocRename)
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

type AnyShareDocRemove struct {
	DocID string `json:"docid"`
}

// Name implements entity.Action.
func (*AnyShareDocRemove) Name() string {
	return common.OpAnyShareDocRemove
}

func (*AnyShareDocRemove) ParameterNew() interface{} {
	return &AnyShareDocRemove{}
}

// Run implements entity.Action.
func (*AnyShareDocRemove) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDocRemove)
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

type AnyShareDocTag struct {
	DocID string      `json:"docid"`
	Tags  interface{} `json:"tags"`
}

// Name implements entity.Action.
func (*AnyShareDocTag) Name() string {
	return common.OpAnyShareDocAddTag
}

func (*AnyShareDocTag) ParameterNew() interface{} {
	return &AnyShareDocTag{}
}

// Run implements entity.Action.
func (*AnyShareDocTag) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDocTag)
	curID := utils.GetDocCurID(input.DocID)
	tags := getTags(input.Tags)

	var res interface{}
	if len(tags) > 0 {

		res, err = ctx.NewASDoc().SetTag(ctx.Context(), curID, tags, token.Token, token.LoginIP)

		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}
	}

	ctx.Trace(ctx.Context(), "run end")
	return res, err
}

type AnyShareDocSetCsfLevel struct {
	DocID    string      `json:"docid"`
	CsfLevel interface{} `json:"csf_level"`
}

type CsfLevelInfo struct {
	CsfInfo  CsfInfo `json:"csfinfo"`
	CsfLevel int     `json:"csflevel"`
}

type CsfInfo struct {
	Scope         string      `json:"scope"`
	Screason      string      `json:"screason"`
	Secrecyperiod interface{} `json:"secrecyperiod"`
}

// Name implements entity.Action.
func (*AnyShareDocSetCsfLevel) Name() string {
	return common.OpAnyShareDocSetCsfLevel
}

func (*AnyShareDocSetCsfLevel) ParameterNew() interface{} {
	return &AnyShareDocSetCsfLevel{}
}

// Run implements entity.Action.
func (*AnyShareDocSetCsfLevel) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDocSetCsfLevel)

	efast := ctx.NewASDoc()

	doc, err := efast.GetDocMsg(ctx.Context(), input.DocID)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	var result float64

	var csflevel int
	var csfinfo *CsfInfo = nil

	if val, ok := input.CsfLevel.(int); ok {
		csflevel = val
	} else if val, ok := input.CsfLevel.(CsfLevelInfo); ok {
		csflevel = val.CsfLevel
		csfinfo = &val.CsfInfo
	} else if val, ok := input.CsfLevel.(string); ok {
		if val == "" {
			return map[string]interface{}{}, nil
		}
		csflevel, err = strconv.Atoi(val)
		if err != nil {
			var obj CsfLevelInfo
			err = json.Unmarshal([]byte(val), &obj)
			if err != nil {
				return nil, err
			}
			csflevel = obj.CsfLevel
			csfinfo = &obj.CsfInfo
		}
	} else if val, ok := input.CsfLevel.(map[string]interface{}); ok {
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
		csfinfo = &obj.CsfInfo
	}

	if doc.Size == -1 {
		objectID := utils.GetDocCurID(input.DocID)
		err = efast.SetDirAttributes(ctx.Context(), objectID, []string{"csflevel"}, map[string]interface{}{
			"csflevel": csflevel,
		}, token.Token, token.LoginIP)

		if err != nil {
			result = float64(csflevel)
		}
	} else {
		result, err = efast.SetCsfLevel(ctx.Context(), input.DocID, csflevel, token.Token, token.LoginIP)
	}

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	if csfinfo != nil && csflevel > 5 && doc.Size != -1 {
		err = efast.SetCsfInfo(ctx.Context(), input.DocID, csfinfo.Scope, csfinfo.Screason, fmt.Sprintf("%v", csfinfo.Secrecyperiod), token.Token, token.LoginIP)

		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}
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

type AnyShareDocSetTemplate struct {
	DocID string      `json:"docid"`
	Tpls  interface{} `json:"templates"`
}

// Name implements entity.Action.
func (*AnyShareDocSetTemplate) Name() string {
	return common.OpAnyShareDocSetTemplate
}

func (*AnyShareDocSetTemplate) ParameterNew() interface{} {
	return &AnyShareDocSetTemplate{}
}

// Run implements entity.Action.
func (*AnyShareDocSetTemplate) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDocSetTemplate)

	res, err := setTemplate(ctx, input, token.Token, token.LoginIP)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	ctx.Trace(ctx.Context(), "run end")
	return res, err
}

func setTemplate(ctx entity.ExecuteContext, input *AnyShareDocSetTemplate, token, ip string) (map[string]map[string]interface{}, error) { //nolint
	curID := utils.GetDocCurID(input.DocID)
	preTpls, err := ctx.NewASDoc().GetTemplates(ctx.Context(), curID, token, ip)
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	var res = make(map[string]map[string]interface{})
	var latestErr error

	var templates = make(map[string]map[string]interface{}, 0)

	if t, ok := input.Tpls.(map[string]interface{}); ok {
		for k, v := range t {
			if val, ok := v.(map[string]interface{}); ok {
				templates[k] = val
			} else if val, ok := v.(string); ok {
				var m = make(map[string]interface{})
				err = json.Unmarshal([]byte(val), &m)
				if err != nil {
					return nil, err
				}
				templates[k] = m
			} else {
				return nil, errors.New("invalid template")
			}
		}
	} else if t, ok := input.Tpls.(string); ok {
		if t == "" {
			return res, nil
		}
		err = json.Unmarshal([]byte(t), &templates)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("invalid template")
	}

	for key, tpl := range templates {
		curTpl := make(map[string]interface{})
		tplStrct, terr := ctx.NewASDoc().GetTemplateStruct(ctx.Context(), key, token, ip)
		if terr != nil {
			httpError, ok := terr.(errs.ExHTTPError)
			if ok && httpError.Status == http.StatusNotFound {
				continue
			}
			ctx.Trace(ctx.Context(), terr.Error())
			return nil, terr
		}
		tplFields := tplStrct.Fields
		fieldsMap := make(map[string]drivenadapters.TemplateField)
		for _, v := range tplFields {
			fieldsMap[v.Key] = v
		}

		for k := range fieldsMap {
			val, ok := tpl[k]
			if ok {
				curTpl[k] = parseTmpValue(k, val, fieldsMap)
			} else {
				curTpl[k] = parseTmpValue(k, "", fieldsMap)
			}
		}

		preTpl, exist := preTpls[key]
		if exist {
			for tplKey, tplValue := range preTpl {
				if strings.Contains(tplKey, "field_") {
					curKey, ok := curTpl[tplKey]
					if !ok {
						continue
					}
					if curKey == "" || curKey == nil {
						curTpl[tplKey] = tplValue
					}
				}
			}
		}

		_, err = ctx.NewASDoc().SetTemplate(ctx.Context(), curID, key, curTpl, token, ip)

		if err != nil {
			parsedError, _err := errs.ExHTTPErrorParser(err)
			if _err != nil {
				latestErr = err
				continue
			}
			if parsedError["code"] != float64(templateNotExists) {
				latestErr = err
			}
			continue
		}

		res[key] = curTpl
	}

	return res, latestErr
}

type AnyShareDocSetPerm struct {
	DocID         string      `json:"docid"`
	Inherit       bool        `json:"inherit"`
	Perminfos     PermInfos   `json:"perminfos"`
	AccessorPerms interface{} `json:"asAccessorPerms"`
	Type          string      `json:"type"`
}

type AccessorPerms struct {
	Inherit   bool      `json:"inherit"`
	Perminfos PermInfos `json:"perminfos"`
}

// Name implements entity.Action.
func (*AnyShareDocSetPerm) Name() string {
	return common.OpAnyShareDocSetPerm
}

func (*AnyShareDocSetPerm) ParameterNew() interface{} {
	return &AnyShareDocSetPerm{}
}

// Run implements entity.Action.
func (*AnyShareDocSetPerm) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	taskIns := ctx.GetTaskInstance()
	if taskIns == nil {
		return nil, fmt.Errorf("get taskinstance failed")
	}

	input := params.(*AnyShareDocSetPerm)
	id := ctx.GetTaskID()

	inherit := input.Inherit
	perminfos := input.Perminfos

	if input.Type == "asAccessorPerms" {

		switch input.AccessorPerms.(type) {
		case string:
			{
				if input.AccessorPerms == "" {
					return map[string]interface{}{}, nil
				}
				var accessorPerms AccessorPerms
				err = json.Unmarshal([]byte(input.AccessorPerms.(string)), &accessorPerms)
				if err != nil {
					return nil, err
				}
				inherit = accessorPerms.Inherit
				perminfos = accessorPerms.Perminfos
			}

		case map[string]interface{}:
			{
				bytes, err := json.Marshal(input.AccessorPerms)

				if err != nil {
					return nil, err
				}

				var accessorPerms AccessorPerms
				err = json.Unmarshal(bytes, &accessorPerms)
				if err != nil {
					return nil, err
				}
				inherit = accessorPerms.Inherit
				perminfos = accessorPerms.Perminfos
			}

		case AccessorPerms:
			{
				accessorPerms := input.AccessorPerms.(AccessorPerms)
				inherit = accessorPerms.Inherit
				perminfos = accessorPerms.Perminfos
			}

		default:
			{
				return nil, errors.New("invalid type")
			}
		}
	}

	docShare := drivenadapters.NewDocShare()
	efast := ctx.NewASDoc()

	permInfoMap := make(PermInfoMap)

	// 申请权限先获取当前权限
	if taskIns.RelatedDagInstance.DagType == common.DagTypeSecurityPolicy && taskIns.RelatedDagInstance.PolicyType == common.SecurityPolicyPerm {
		currentPerms, err := docShare.GetPerm(ctx.Context(), input.DocID, token.Token)

		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}
		// 权限申请不修改继承权限
		inherit = currentPerms.Inherit

		for _, permInfo := range currentPerms.Perminfos {
			if permInfo.InheritDocID == "" {
				config, err := permInfo.ToPermConfig()
				if err != nil {
					return nil, err
				}
				permInfoMap[permInfo.AccessorID] = *config
			}
		}
	}

	err = perminfos.Build(permInfoMap)
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	var res float64

	// 上传审核记录权限
	if taskIns.RelatedDagInstance.DagType == common.DagTypeSecurityPolicy && taskIns.RelatedDagInstance.PolicyType == common.SecurityPolicyUpload {

		perms := make([]drivenadapters.UploadProcessPerm, 0)

		for _, permInfo := range permInfoMap {

			// 解析时间戳
			endtime := utils.TimeParse(permInfo.ExpiresAt)

			perms = append(perms, drivenadapters.UploadProcessPerm{
				Accessor: drivenadapters.Accessor{
					ID:   permInfo.Accessor.ID,
					Type: permInfo.Accessor.Type,
				},
				Allow:   permInfo.Allow,
				Deny:    permInfo.Deny,
				Endtime: endtime,
			})
		}

		err = drivenadapters.NewEfast().SetUploadProcessPerm(ctx.Context(), taskIns.RelatedDagInstance.ID, input.DocID, perms)

		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}
	} else {

		kcmc := drivenadapters.NewKcmc()
		permInfoParams := make([]drivenadapters.PermConfig, 0)

		for _, permInfo := range permInfoMap {
			permInfoParams = append(permInfoParams, permInfo)
		}

		docAttr, err := efast.GetDocMsg(ctx.Context(), input.DocID)

		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}

		var isArticle bool

		if docAttr.DocLibType == common.CustomDocLib && docAttr.CustomType != nil {
			isArticle, err = kcmc.IsArticleProxyDocLibSubtype(ctx.Context(), docAttr.CustomType.ID)

			if err != nil {
				return nil, err
			}
		}

		if isArticle {
			article, err := kcmc.GetArticleByProxyDirID(ctx.Context(), docAttr.DocID)

			if err != nil {
				return nil, err
			}

			kcPermItems := make([]*drivenadapters.KcPermItem, 0)

			for _, permInfo := range permInfoParams {
				// 解析时间戳
				endtime := utils.TimeParse(permInfo.ExpiresAt)

				allowMap := make(map[string]bool)

				for _, p := range permInfo.Allow {
					allowMap[p] = true
				}

				if _, ok := allowMap["delete"]; ok {
					kcPermItems = append(kcPermItems, &drivenadapters.KcPermItem{
						EndTime:    endtime,
						Kind:       "admin",
						ObjectID:   permInfo.Accessor.ID,
						ObjectType: permInfo.Accessor.Type,
					})
					continue
				}

				if _, ok := allowMap["modify"]; ok {
					kcPermItems = append(kcPermItems, &drivenadapters.KcPermItem{
						EndTime:    endtime,
						Kind:       "write",
						ObjectID:   permInfo.Accessor.ID,
						ObjectType: permInfo.Accessor.Type,
					})
					continue
				}

				if _, ok := allowMap["download"]; ok {
					kcPermItems = append(kcPermItems, &drivenadapters.KcPermItem{
						EndTime:    endtime,
						Kind:       "download",
						ObjectID:   permInfo.Accessor.ID,
						ObjectType: permInfo.Accessor.Type,
					})
					continue
				}

				if _, ok := allowMap["preview"]; ok {
					kcPermItems = append(kcPermItems, &drivenadapters.KcPermItem{
						EndTime:    endtime,
						Kind:       "read",
						ObjectID:   permInfo.Accessor.ID,
						ObjectType: permInfo.Accessor.Type,
					})
					continue
				}

				if _, ok := allowMap["display"]; ok {
					kcPermItems = append(kcPermItems, &drivenadapters.KcPermItem{
						EndTime:    endtime,
						Kind:       "view",
						ObjectID:   permInfo.Accessor.ID,
						ObjectType: permInfo.Accessor.Type,
					})
					continue
				}
			}

			res, err = kcmc.SetPerm(ctx.Context(), drivenadapters.KcPerm{
				Link:     "",
				SAID:     strconv.FormatUint(article.ArticleID, 10),
				PowerArr: kcPermItems,
			}, token.Token)

			if err != nil {
				ctx.Trace(ctx.Context(), err.Error())
				return nil, err
			}
		} else {
			permConfig := &drivenadapters.DocPermConfig{
				Configs:     permInfoParams,
				Inherit:     inherit,
				SendMessage: true,
			}
			_, err = docShare.SetDocPerm2(ctx.Context(), permConfig, token.Token, input.DocID)
			if err != nil {
				ctx.Trace(ctx.Context(), err.Error())
				return nil, err
			}
			res = 1
		}
	}

	var result = map[string]interface{}{
		"result": res,
	}

	ctx.ShareData().Set(id, result)

	ctx.Trace(ctx.Context(), "run end")
	return result, err
}

type AnyShareDocGetPath struct {
	DocID  string `json:"docid"`
	Order  string `json:"order"`
	Depth  int    `json:"depth"`
	Custom int    `json:"custom"`
}

func (*AnyShareDocGetPath) Name() string {
	return common.OpAnyShareDocGetPath
}

func (*AnyShareDocGetPath) ParameterNew() interface{} {
	return &AnyShareDocGetPath{}
}

func (*AnyShareDocGetPath) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*AnyShareDocGetPath)
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

type AnyShareDocSetSpaceQuota struct {
	DocID string      `json:"docid"`
	Quota interface{} `json:"quota"`
}

func (*AnyShareDocSetSpaceQuota) Name() string {
	return common.OpAnyShareDocSetSpaceQuota
}

func (*AnyShareDocSetSpaceQuota) ParameterNew() interface{} {
	return &AnyShareDocSetSpaceQuota{}
}

func (*AnyShareDocSetSpaceQuota) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	input := params.(*AnyShareDocSetSpaceQuota)

	efast := ctx.NewASDoc()

	objectID := utils.GetDocCurID(input.DocID)

	quota := 0

	switch input.Quota.(type) {
	case string:
		quota, err = strconv.Atoi(input.Quota.(string))

		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}
	case int:
		quota = input.Quota.(int)
	case int32:
		quota = int(input.Quota.(int32))
	case int64:
		quota = int(input.Quota.(int64))
	case float32:
		quota = int(input.Quota.(float32))
	case float64:
		quota = int(input.Quota.(float64))
	default:
		ctx.Trace(ctx.Context(), "invalid quota type")
		return nil, errors.New("invalid quota type")
	}

	if objectID == "" {
		ctx.Trace(ctx.Context(), "invalid docid")
		return nil, errors.New("invalid docid")
	}

	err = efast.SetDirAttributes(ctx.Context(), objectID, []string{"space_quota"}, map[string]interface{}{
		"space_quota": quota,
	}, token.Token, token.LoginIP)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	ctx.Trace(ctx.Context(), "run end")
	return nil, nil
}

type DocSuffix struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Suffix []string `json:"suffix"`
}

type AnyShareDocSetAllowSuffixDoc struct {
	DocID          string      `json:"docid"`
	AllowSuffixDoc interface{} `json:"allow_suffix_doc"`
}

func (*AnyShareDocSetAllowSuffixDoc) Name() string {
	return common.OpAnyShareDocSetAllowSuffixDoc
}

func (*AnyShareDocSetAllowSuffixDoc) ParameterNew() interface{} {
	return &AnyShareDocSetAllowSuffixDoc{}
}

func (*AnyShareDocSetAllowSuffixDoc) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	input := params.(*AnyShareDocSetAllowSuffixDoc)
	efast := ctx.NewASDoc()
	objectID := utils.GetDocCurID(input.DocID)

	if objectID == "" {
		ctx.Trace(ctx.Context(), "invalid docid")
		return nil, errors.New("invalid docid")
	}

	var s []byte

	switch input.AllowSuffixDoc.(type) {
	case string:
		s = []byte(input.AllowSuffixDoc.(string))
	default:
		s, err = json.Marshal(input.AllowSuffixDoc)

		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}
	}

	var allowSuffixDoc []DocSuffix

	if err = json.Unmarshal(s, &allowSuffixDoc); err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	err = efast.SetDirAttributes(ctx.Context(), objectID, []string{"allow_suffix_doc"}, map[string]interface{}{
		"allow_suffix_doc": allowSuffixDoc,
	}, token.Token, token.LoginIP)
	return nil, err
}

var (
	_ entity.Action = (*AnyShareDocRename)(nil)
	_ entity.Action = (*AnyShareDocRemove)(nil)
	_ entity.Action = (*AnyShareDocTag)(nil)
	_ entity.Action = (*AnyShareDocSetCsfLevel)(nil)
	_ entity.Action = (*AnyShareDocSetTemplate)(nil)
	_ entity.Action = (*AnyShareDocSetPerm)(nil)
	_ entity.Action = (*AnyShareDocGetPath)(nil)
	_ entity.Action = (*AnyShareDocSetSpaceQuota)(nil)
	_ entity.Action = (*AnyShareDocSetAllowSuffixDoc)(nil)
)
