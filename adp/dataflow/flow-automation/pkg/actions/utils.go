package actions

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	normalizeutil "github.com/kweaver-ai/adp/autoflow/flow-automation/utils/normalize"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskSleepTime 任务运行间隔时间
const TaskSleepTime = 1

func getTriggerVars(ctx entity.ExecuteContext) map[string]interface{} {
	id, idOk := ctx.GetVar("id")
	if !idOk {
		id = ""
	}
	newID, newIDOk := ctx.GetVar("new_id")
	if !newIDOk {
		newID = ""
	}
	name, nameOK := ctx.GetVar("name")
	if !nameOK {
		name = ""
	}
	path, pathOK := ctx.GetVar("path")
	if !pathOK {
		path = ""
	}
	newPath, newPathOk := ctx.GetVar("new_path")
	if !newPathOk {
		newPath = ""
	}
	var size int64
	s, sizeOK := ctx.GetVar("size")
	if !sizeOK {
		size = 0
	} else {
		var err error
		size, err = strconv.ParseInt(s.(string), 10, 64)
		if err != nil {
			size = 0
		}
	}

	res := map[string]interface{}{
		"id":          id,
		"name":        name,
		"path":        path,
		"size":        size,
		"new_id":      newID,
		"new_path":    newPath,
		"item_id":     utils.GetDocCurID(id.(string)),
		"new_item_id": utils.GetDocCurID(newID.(string)),
	}

	accessorID, _ := ctx.GetVar("operator_id")
	accessorName, _ := ctx.GetVar("operator_name")
	accessorType, _ := ctx.GetVar("operator_type")
	bytes, _ := json.Marshal(map[string]interface{}{
		"id":   accessorID,
		"name": accessorName,
		"type": accessorType,
	})
	res["accessor"] = string(bytes)

	return res
}

func getDocInfo(ctx context.Context, docID, token, ip string, isapp bool, asdoc drivenadapters.Efast) (doc map[string]interface{}, attr *drivenadapters.DocAttr, err error) {
	attr, err = asdoc.GetDocMsg(ctx, docID)

	if err != nil {
		return
	}

	doc = map[string]interface{}{
		"id":          attr.ID,
		"docid":       attr.ID,
		"creator":     attr.Creator,
		"creator_id":  attr.CreatorID,
		"create_time": TimeToISOString(attr.CreateTime, TimeUnitMicrosecond),
		"modified":    TimeToISOString(attr.Modified, TimeUnitMicrosecond),
		"editor":      attr.Editor,
		"editor_id":   attr.EditorID,
		"name":        attr.Name,
		"path":        attr.Path,
		"size":        attr.Size,
		"rev":         attr.Rev,
		"csflevel":    attr.CsfLevel,
	}
	if isapp {
		return doc, attr, nil
	}
	res, err := asdoc.CheckPerm(ctx, docID, "display", token, ip)
	if err != nil {
		return
	}

	if res != 0 {
		err = errors.NewIError(errors.NoPermission, "", map[string]interface{}{
			"info": "has no perm to get doc metadata",
			"doc": map[string]string{
				"docid":   docID,
				"docname": attr.Name,
			},
		})
		return
	}

	return doc, attr, err
}

func getTags(inputTags interface{}) []string {
	tags := []string{}

	switch s := inputTags.(type) {
	case string:

		err := json.Unmarshal([]byte(s), &tags)
		if err == nil {
			return tags
		}

		tagMap := make(map[int]string, 0)
		err = json.Unmarshal([]byte(s), &tagMap)
		if err != nil {
			tags = append(tags, s)
			break
		}
		for _, tag := range tagMap {
			tags = append(tags, tag)
		}
	case []string:
		tags = append(tags, s...)
	case primitive.A:
		for i := range s {
			if ss, ok := s[i].(string); ok {
				tags = append(tags, ss)
			}
		}
	case []interface{}:
		for i := range s {
			if ss, ok := s[i].(string); ok {
				tags = append(tags, ss)
			}
		}
	}

	return tags
}

func getDocPath(order string, depth int, doc map[string]interface{}) string {
	var path string
	arrs := strings.Split(doc["path"].(string), "/")
	paths := arrs[0 : len(arrs)-1]

	if depth == -1 || depth > len(paths) {
		path = strings.Join(paths, "/")
	} else {
		if order == common.DESC {
			path = strings.Join(paths[0:depth], "/")
		} else if order == common.ASC {
			path = strings.Join(paths[len(paths)-depth:], "/")
		}
	}

	return path
}

func handleDocCopy(ctx entity.ExecuteContext, docID string, token *entity.Token) (map[string]interface{}, error) {
	newDocInfo, _, err := getDocInfo(ctx.Context(), docID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"new_docid":   docID,
		"name":        newDocInfo["name"],
		"new_path":    newDocInfo["path"],
		"size":        newDocInfo["size"],
		"create_time": newDocInfo["create_time"],
		"creator":     newDocInfo["creator"],
	}

	return data, err
}

func handleDocRemove(ctx entity.ExecuteContext, docID string, token *entity.Token) (map[string]interface{}, error) {
	attr, _, err := getDocInfo(ctx.Context(), docID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"docid": docID,
		"name":  attr["name"],
		"path":  attr["path"],
	}

	return data, err
}

func handleDocMove(ctx entity.ExecuteContext, docID string, token *entity.Token) (map[string]interface{}, error) {
	attr, _, err := getDocInfo(ctx.Context(), docID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"docid":       docID,
		"name":        attr["name"],
		"path":        attr["path"],
		"size":        attr["size"],
		"create_time": attr["create_time"],
		"creator":     attr["creator"],
	}

	return data, err
}

func handleDocRename(ctx entity.ExecuteContext, docID string, token *entity.Token) (map[string]interface{}, error) {
	attr, _, err := getDocInfo(ctx.Context(), docID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"docid":       docID,
		"name":        attr["name"],
		"path":        attr["path"],
		"size":        attr["size"],
		"create_time": attr["create_time"],
		"creator":     attr["creator"],
	}

	return data, err
}

func handleSetTemplate(ctx entity.ExecuteContext, input *AnyShareSetTemplateParam, token, ip string) (map[string]map[string]interface{}, error) { //nolint
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
				return nil, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{})
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
		return nil, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{})
	}

	for key, tpl := range templates {
		curTpl := make(map[string]interface{})
		tplStrct, terr := ctx.NewASDoc().GetTemplateStruct(ctx.Context(), key, token, ip)
		if terr != nil {
			httpError, ok := terr.(errors.ExHTTPError)
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
			parsedError, _err := errors.ExHTTPErrorParser(err)
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

func parseTmpValue(k string, v interface{}, fieldsMap map[string]drivenadapters.TemplateField) interface{} {
	if v == nil {
		return nil
	}
	field, ok := fieldsMap[k]
	if !ok {
		return ""
	}

	switch field.Type {
	case "int":
		num, err := strconv.Atoi(fmt.Sprintf("%v", v))
		if err != nil {
			return 0
		}
		return num
	default:
		return v
	}
}

func getFileStream(URL string, timeout time.Duration) (*[]byte, error) {
	client := drivenadapters.NewRawHTTPClient()
	client.Timeout = timeout
	resp, err := client.Get(URL)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			commonLog.NewLogger().Errorln(closeErr)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respCode := resp.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		err = errors.ExHTTPError{
			Body:   string(body),
			Status: respCode,
		}
		return nil, err
	}

	return &body, nil
}

// convertTimeStringToMsTimestamp 将时间字符串转换为微秒级时间戳
func convertTimeStringToMsTimestamp(timeStr string) (int64, error) {
	return utils.ConvertTimeStringToMsTimestamp(timeStr)
}

func checkOCRAvailable(ctx entity.ExecuteContext, token *entity.Token) error {
	appstoreAdapter := drivenadapters.NewAppStore()
	userManagement := drivenadapters.NewUserManagement()
	userDetail, err := userManagement.GetUserInfoByType(token.UserID, utils.IfNot(token.IsApp, common.APP.ToString(), common.User.ToString()))
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return err
	}
	if !utils.IsAdminRole(userDetail.Roles) {
		whitelistStatus, err := appstoreAdapter.GetWhiteListStatus(ctx.Context(), "action_ocr", token.Token)
		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return err
		}
		if whitelistStatus["enable"] == false {
			return errors.NewIError(errors.Forbidden, "", nil)
		}
	}

	return nil
}

func handleOCR(ctx entity.ExecuteContext, token *entity.Token, docID, recType, scene, uri string, taskType int64) (map[string]interface{}, error) {
	id := ctx.GetTaskID()
	ocrType := common.NewConfig().T4th.Type
	var result = map[string]interface{}{
		"results": nil,
	}

	_, docInfo, err := getDocInfo(ctx.Context(), docID, token.Token, token.LoginIP, token.IsApp, ctx.NewASDoc())
	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	payload := map[string]interface{}{
		"docid":    docID,
		"rec_type": recType,
		"scene":    scene,
	}

	if ocrType == "fileReader" {
		payload["task_type"] = taskType
		payload["uri"] = uri
	}

	ocrRes, err := ctx.NewRepo().RecognizeText(ctx.Context(), payload, docInfo)
	if err != nil {
		// 超过指定的文件大小跳过处理
		if errors.Is(err, errors.FileSizeExceed) {
			ctx.ShareData().Set(id, result)
		} else {
			ctx.Trace(ctx.Context(), err.Error())
		}
		return nil, err
	}
	jdata, _ := json.Marshal(ocrRes)
	if recType == "general" {
		result["results"] = string(jdata)
	} else {
		ocrRes["results"] = string(jdata)
		result = ocrRes
	}
	ctx.ShareData().Set(id, result)

	return result, nil
}

type AnyShareRelevanceParams struct {
	DocID     string      `json:"docid"`
	Relevance interface{} `json:"relevance"`
}

func handleAddRelevance(ctx context.Context, params *AnyShareRelevanceParams, token string, ip string) error {
	var err error

	docids := []string{params.DocID}

	if s, ok := params.Relevance.([]string); ok {
		docids = append(docids, s...)
	} else if str, ok := params.Relevance.(string); ok {
		docids = append(docids, str)
	} else if arr, ok := normalizeutil.AsSlice(params.Relevance); ok {
		for _, item := range arr {
			if docid, ok := item.(string); ok {
				docids = append(docids, docid)
			}
		}
	}

	docids = utils.RemoveRepByMap(docids)

	if len(docids) < 2 {
		err = fmt.Errorf("relevance is empty")
		return errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{
			"error":  err.Error(),
			"docids": docids,
		})
	}

	efast := drivenadapters.NewEfast()

	metadatas, err := efast.BatchGetMetadata(ctx, docids, []string{"types,sizes"})

	if err != nil {
		return err
	}

	var itemMap = make(map[string]drivenadapters.Object)

	for _, metadata := range metadatas {
		docid := metadata["id"].(string)
		itemMap[docid] = drivenadapters.Object{
			ID:   GetObjectIDFromDocID(docid),
			Type: metadata["type"].(string),
		}
	}

	var relevanceParams drivenadapters.RelevanceParams

	item, ok := itemMap[params.DocID]

	if !ok {
		return errors.NewIError(errors.FileNotFound, "", map[string]interface{}{"docid": params.DocID})
	}

	relevanceParams.Item = item

	for _, docid := range docids[1:] {
		item, ok := itemMap[docid]
		if !ok {
			return errors.NewIError(errors.FileNotFound, "", map[string]interface{}{"docid": params.DocID})
		}

		relevanceParams.Relevance = append(relevanceParams.Relevance, drivenadapters.RelevanceItem{
			Details: "",
			Item:    item,
		})
	}

	err = efast.AddRelevance(ctx, relevanceParams, token, ip)

	return err
}

func GetObjectIDFromDocID(docid string) string {
	return docid[len(docid)-32:]
}

func ParseInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		isInt64, err := IsInt64DataType(v)
		if err != nil || !isInt64 {
			return 0
		}
		return int(v)
	default:
		return 0
	}
}

// IsInt64DataType 判断float64是否为int64类型
func IsInt64DataType(num float64) (bool, error) {
	valueStr := fmt.Sprintf("%.10f", num)
	f, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return false, err
	}
	if f == float64(int64(f)) {
		return true, nil
	}
	return false, nil
}

func IsInt(val interface{}) bool {
	switch reflect.TypeOf(val).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

type TimeUnit string

const (
	TimeUnitMicrosecond TimeUnit = "µs"
	TimeUnitMillisecond TimeUnit = "ms"
	TimeUnitSecond      TimeUnit = "s"
)

func TimeToISOString[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](timestamp T, unit TimeUnit) string {
	var timeInSeconds float64
	switch unit {
	case TimeUnitMicrosecond:
		timeInSeconds = float64(timestamp) / 1e6
	case TimeUnitMillisecond:
		timeInSeconds = float64(timestamp) / 1e3
	case TimeUnitSecond:
		timeInSeconds = float64(timestamp)
	}

	return time.Unix(int64(timeInSeconds), 0).Format(time.RFC3339)
}

type PermInfos []PermInfo

// PermInfo 权限信息
type PermInfo struct {
	Accessor interface{} `json:"accessor"`
	Perm     interface{} `json:"perm"`
	EndTime  interface{} `json:"endtime"`
}

// Accessor 访问者信息
type Accessor struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Perm 权限配置
type Perm struct {
	Allow []string `json:"allow"`
	Deny  []string `json:"deny"`
}

type PermInfoMap map[string]drivenadapters.PermConfig

func (ps PermInfos) Build(permInfoMap PermInfoMap) error {
	var err error
	for _, permInfo := range ps {
		// 永久有效，固定格式日期
		var endTime = "1970-01-01T08:00:00+08:00"
		if t, ok := permInfo.EndTime.(string); ok {
			if t != "" {
				endTime, err = utils.ConvertToRFC3339(t)
				if err != nil {
					return err
				}
			}
		}

		var accessor Accessor
		switch parsedAccessor := permInfo.Accessor.(type) {
		case string:
			err = json.Unmarshal([]byte(parsedAccessor), &accessor)
			if err != nil {
				return err
			}
		case map[string]interface{}:
			accessor = Accessor{
				Type: parsedAccessor["type"].(string),
				ID:   parsedAccessor["id"].(string),
			}
		}
		var perm Perm
		switch parsedPerm := permInfo.Perm.(type) {
		case string:
			err = json.Unmarshal([]byte(parsedPerm), &perm)
			if perm.Allow == nil {
				perm.Allow = []string{}
			}
			if perm.Deny == nil {
				perm.Deny = []string{}
			}
			if err != nil {
				return err
			}
		case map[string]interface{}:
			parsedAllow := []string{}
			parsedDeny := []string{}
			if allow, ok := normalizeutil.AsSlice(parsedPerm["allow"]); ok {
				for _, item := range allow {
					if str, ok := item.(string); ok {
						parsedAllow = append(parsedAllow, str)
					}
				}
			}

			if deny, ok := normalizeutil.AsSlice(parsedPerm["deny"]); ok {
				for _, item := range deny {
					if str, ok := item.(string); ok {
						parsedDeny = append(parsedDeny, str)
					}
				}
			}

			perm = Perm{
				Allow: parsedAllow,
				Deny:  parsedDeny,
			}
		}

		if p, ok := permInfoMap[accessor.ID]; ok {

			allowMap := make(map[string]bool)
			denyMap := make(map[string]bool)

			for _, item := range p.Deny {
				denyMap[item] = true
			}

			for _, item := range p.Allow {
				allowMap[item] = true
			}

			for _, item := range perm.Deny {
				denyMap[item] = true
			}

			for _, item := range perm.Allow {
				allowMap[item] = true
			}

			if len(allowMap) > 0 {
				allowMap["display"] = true
			}

			if _, allowModify := allowMap["modify"]; allowModify {

				_, allowPreview := allowMap["preview"]
				_, allowDownload := allowMap["download"]

				if !allowPreview && !allowDownload {
					allowMap["download"] = true
					allowMap["preview"] = true
				}
			}

			allow := make([]string, 0)
			deny := make([]string, 0)

			for k, v := range allowMap {
				if v {
					allow = append(allow, k)
					delete(denyMap, k)
				}
			}

			for k, v := range denyMap {
				if v {
					deny = append(deny, k)
				}
			}

			permInfoMap[accessor.ID] = drivenadapters.PermConfig{
				Accessor: drivenadapters.Accessor{
					ID:   accessor.ID,
					Type: accessor.Type,
				},
				Allow:     allow,
				Deny:      deny,
				ExpiresAt: endTime,
			}
		} else {
			permInfoMap[accessor.ID] = drivenadapters.PermConfig{
				Accessor: drivenadapters.Accessor{
					ID:   accessor.ID,
					Type: accessor.Type,
				},
				Allow:     perm.Allow,
				Deny:      perm.Deny,
				ExpiresAt: endTime,
			}
		}
	}

	return nil
}

// 获取路径
// . => []
// a => ["a"]
// a.b => ["a", "b"]
// .a.b => ["a", "b"]
// .a[0] => ["a", 0]
// .a.b["c"] => ["a", "b", "c"]
// .a.b.["c"] => ["a", "b", "c"]
// .a."b.c".d => ["a", "b.c", "d"]
func keyToPath(key string) []interface{} {
	key = strings.TrimSpace(key)
	key = strings.Trim(key, ".")

	var path []interface{}
	var i int
	for i < len(key) {
		if key[i] == ' ' {
			i++
			continue
		}

		if key[i] == '.' {
			i++
			continue
		}

		if key[i] == '[' {
			j := i + 1
			for j < len(key) && key[j] != ']' {
				j++
			}
			if j >= len(key) {
				break
			}
			indexStr := key[i+1 : j]
			i = j + 1
			if num, err := strconv.Atoi(indexStr); err == nil {
				path = append(path, num)
			} else {
				if strings.HasPrefix(indexStr, `"`) && strings.HasSuffix(indexStr, `"`) {
					indexStr = indexStr[1 : len(indexStr)-1]
				}
				path = append(path, indexStr)
			}
			continue
		}

		if key[i] == '"' {
			j := i + 1
			for j < len(key) && key[j] != '"' {
				j++
			}
			if j >= len(key) {
				break
			}
			str := key[i+1 : j]
			path = append(path, str)
			i = j + 1
			continue
		}

		j := i
		for j < len(key) && key[j] != '.' && key[j] != '[' && key[j] != '"' && key[j] != ' ' {
			j++
		}
		str := key[i:j]
		path = append(path, str)
		i = j
	}
	return path
}

func lookupJson(obj interface{}, path []interface{}, defaultValue interface{}) (value interface{}) {

	if len(path) == 0 {
		return obj
	}

	if obj == nil {
		return defaultValue
	}

	key := fmt.Sprintf("%v", path[0])
	switch obj := obj.(type) {
	case map[string]interface{}:
		if v, ok := obj[key]; ok {
			return lookupJson(v, path[1:], defaultValue)
		} else {
			return defaultValue
		}
	case []interface{}:
		index, err := strconv.Atoi(key)
		if err != nil {
			return defaultValue
		}
		if index < 0 || index >= len(obj) {
			return defaultValue
		}
		return lookupJson(obj[index], path[1:], defaultValue)
	default:
		return defaultValue
	}
}

func setJson(obj interface{}, paths []interface{}, value interface{}) interface{} {
	if len(paths) == 0 {
		return value
	}
	obj = deepCopy(obj)
	switch t := obj.(type) {
	case map[string]interface{}:
		var key string
		path := paths[0]
		switch path := path.(type) {
		case string:
			key = path
		default:
			key = fmt.Sprintf("%v", path)
		}
		if len(paths) == 1 {
			t[key] = value
			return obj
		}
		t[key] = setJson(t[key], paths[1:], value)
		return obj

	case []interface{}:
		var index int
		path := paths[0]
		switch path := path.(type) {
		case int:
			index = path
		default:
			s := fmt.Sprintf("%v", path)
			var err error
			index, err = strconv.Atoi(s)
			if err != nil {
				return obj
			}
		}

		if index >= len(t) {
			t = append(t, make([]interface{}, index-len(t)+1)...)
		}

		if len(paths) == 1 {
			t[index] = value
			return t
		}

		t[index] = setJson(t[index], paths[1:], value)
		return obj

	default:
		if len(paths) == 0 {
			return value
		}

		var newStructure interface{}
		path := paths[0]
		switch path.(type) {
		case int:
			newStructure = make([]interface{}, 0)
		default:
			newStructure = make(map[string]interface{})
		}
		return setJson(newStructure, paths, value)
	}
}

func deepCopy(obj interface{}) interface{} {
	b, err := json.Marshal(obj)
	if err != nil {
		return obj
	}
	var copy interface{}
	err = json.Unmarshal(b, &copy)
	if err != nil {
		return obj
	}
	return copy
}

func hash(s string) string {
	data := []byte(s)
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// FileDownloadInfo 文件下载信息
type FileDownloadInfo struct {
	URL      string
	Filename string
	Size     int64
	DocAttr  *drivenadapters.DocAttr // 文件属性信息（可选）
}

// GetFileDownloadInfo 获取文件下载信息
func GetFileDownloadInfo(ctx context.Context, execCtx entity.ExecuteContext, docID, version string) (*FileDownloadInfo, error) {
	if common.IsDFSURI(docID) {
		return getDFSFileDownloadInfo(ctx, execCtx, docID)
	}
	return getEfastFileDownloadInfo(ctx, docID, version)
}

// GetFileDownloadInfoWithDocAttr 获取文件下载信息和文件属性
func GetFileDownloadInfoWithDocAttr(ctx context.Context, execCtx entity.ExecuteContext, docID, version string) (*FileDownloadInfo, error) {
	if common.IsDFSURI(docID) {
		return getDFSFileDownloadInfoWithDocAttr(ctx, execCtx, docID)
	}
	return getEfastFileDownloadInfoWithDocAttr(ctx, docID, version, execCtx.NewASDoc())
}

// getDFSFileDownloadInfo 获取DFS URI文件的下载信息
func getDFSFileDownloadInfo(ctx context.Context, execCtx entity.ExecuteContext, docID string) (*FileDownloadInfo, error) {
	dc := execCtx.NewRepo().DocumentConverter()
	f, err := dc.ResolveFlowFile(ctx, docID)
	if err != nil {
		return nil, err
	}

	og := drivenadapters.NewOssGatewayBackend()
	url, err := og.GetDownloadURL(ctx, f.Storage.OssID, f.Storage.ObjectKey, 3600, true)
	if err != nil {
		return nil, err
	}

	return &FileDownloadInfo{
		URL:      url,
		Filename: f.File.Name,
		Size:     int64(f.Storage.Size),
	}, nil
}

// getDFSFileDownloadInfoWithDocAttr 获取DFS URI文件的下载信息和属性
func getDFSFileDownloadInfoWithDocAttr(ctx context.Context, execCtx entity.ExecuteContext, docID string) (*FileDownloadInfo, error) {
	dc := execCtx.NewRepo().DocumentConverter()
	f, err := dc.ResolveFlowFile(ctx, docID)
	if err != nil {
		return nil, err
	}

	og := drivenadapters.NewOssGatewayBackend()
	url, err := og.GetDownloadURL(ctx, f.Storage.OssID, f.Storage.ObjectKey, 3600, true)
	if err != nil {
		return nil, err
	}

	return &FileDownloadInfo{
		URL:      url,
		Filename: f.File.Name,
		Size:     int64(f.Storage.Size),
		DocAttr: &drivenadapters.DocAttr{
			DocID: docID,
			Name:  f.File.Name,
			Size:  float64(f.Storage.Size),
		},
	}, nil
}

// getEfastFileDownloadInfo 获取普通文档ID的下载信息
func getEfastFileDownloadInfo(ctx context.Context, docID, version string) (*FileDownloadInfo, error) {
	efast := drivenadapters.NewEfast()
	downInfo, err := efast.InnerOSDownload(ctx, docID, version)
	if err != nil {
		return nil, err
	}

	return &FileDownloadInfo{
		URL:      downInfo.URL,
		Filename: downInfo.Name,
		Size:     int64(downInfo.Size),
	}, nil
}

// getEfastFileDownloadInfoWithDocAttr 获取普通文档ID的下载信息和属性
func getEfastFileDownloadInfoWithDocAttr(ctx context.Context, docID, version string, asdoc drivenadapters.Efast) (*FileDownloadInfo, error) {
	efast := drivenadapters.NewEfast()
	downInfo, err := efast.InnerOSDownload(ctx, docID, version)
	if err != nil {
		return nil, err
	}

	docAttr, err := asdoc.GetDocMsg(ctx, docID)
	if err != nil {
		return nil, err
	}

	return &FileDownloadInfo{
		URL:      downInfo.URL,
		Filename: downInfo.Name,
		Size:     int64(downInfo.Size),
		DocAttr:  docAttr,
	}, nil
}
