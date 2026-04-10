package dagmodel

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
)

// ToDagModel 转换为DagModel
func ToDagModel(dag *entity.Dag, isupdate bool) *DagModel {
	if isupdate {
		dag.Update()
	} else {
		dag.Initial()
	}
	id, _ := strconv.ParseUint(dag.ID, 10, 64)
	dagVarBytes, _ := json.Marshal(dag.Vars)
	tasksBytes, _ := json.Marshal(dag.Tasks)
	stepsBytes, _ := json.Marshal(dag.Steps)
	shortcutsBytes, _ := json.Marshal(dag.Shortcuts)
	accessorBytes, _ := json.Marshal(dag.Accessors)
	appInfoBytes, _ := json.Marshal(dag.AppInfo)
	emailsBytes, _ := json.Marshal(dag.Emails)
	triggerConfigBytes, _ := json.Marshal(dag.TriggerConfig)
	subIDsBytes, _ := json.Marshal(dag.SubIDs)
	outputsBytes, _ := json.Marshal(dag.OutPuts)
	instructionsBytes, _ := json.Marshal(dag.Instructions)
	incValuesBytes, _ := json.Marshal(dag.IncValues)

	return &DagModel{
		ID:            id,
		CreatedAt:     dag.CreatedAt,
		UpdatedAt:     dag.UpdatedAt,
		UserID:        dag.UserID,
		Name:          dag.Name,
		Desc:          dag.Desc,
		Trigger:       string(dag.Trigger),
		Cron:          dag.Cron,
		Vars:          string(dagVarBytes),
		Status:        string(dag.Status),
		Tasks:         string(tasksBytes),
		Steps:         string(stepsBytes),
		Description:   dag.Description,
		Shortcuts:     string(shortcutsBytes),
		Accessors:     string(accessorBytes),
		Type:          dag.Type,
		PolicyType:    dag.PolicyType,
		AppInfo:       string(appInfoBytes),
		Priority:      dag.Priority,
		Emails:        string(emailsBytes),
		Template:      dag.Template,
		Published:     dag.Published,
		TriggerConfig: string(triggerConfigBytes),
		SubIDs:        string(subIDsBytes),
		ExecMode:      dag.ExecMode,
		Category:      dag.Category,
		OutPuts:       string(outputsBytes),
		Instructions:  string(instructionsBytes),
		OperatorID:    dag.OperatorID,
		IncValues:     string(incValuesBytes),
		Version:       dag.Version.ToString(),
		VersionID:     dag.VersionID,
		ModifyBy:      dag.ModifyBy,
		IsDebug:       dag.IsDebug,
		DeBugID:       dag.DeBugID,
		BizDomainID:   dag.BizDomainID,
	}
}

// ToDagInstanceModel 转换为DagInstanceModel
func ToDagInstanceModel(dagIns *entity.DagInstance, isupdate bool) *DagInstanceModel {
	if isupdate {
		dagIns.Update()
	} else {
		dagIns.Initial()
	}

	id, _ := strconv.ParseUint(dagIns.ID, 10, 64)
	dagInsID, _ := strconv.ParseUint(dagIns.DagID, 10, 64)
	varsBytes, _ := json.Marshal(dagIns.Vars)
	keywordsBytes, _ := json.Marshal(dagIns.Keywords)
	sharedataBytes, _ := json.Marshal(dagIns.ShareData)
	sharedataextBytes, _ := json.Marshal(dagIns.ShareDataExt)
	cmdBytes, _ := json.Marshal(dagIns.Cmd)
	appInfoBytes, _ := json.Marshal(dagIns.AppInfo)
	dumpextbytes, _ := json.Marshal(dagIns.DumpExt)
	callchainBytes, _ := json.Marshal(dagIns.CallChain)
	hasCmd := dagIns.Cmd != nil && !reflect.DeepEqual(*dagIns.Cmd, entity.Command{})
	batchRunID := ""
	if val, ok := dagIns.Vars["batch_run_id"]; ok {
		batchRunID = val.Value
	}

	return &DagInstanceModel{
		ID:               id,
		CreatedAt:        dagIns.CreatedAt,
		UpdatedAt:        dagIns.UpdatedAt,
		DagID:            dagInsID,
		Trigger:          string(dagIns.Trigger),
		Worker:           dagIns.Worker,
		Source:           dagIns.Source,
		Vars:             string(varsBytes),
		Keywords:         string(keywordsBytes),
		EventPersistence: int(dagIns.EventPersistence),
		EventOssPath:     dagIns.EventOssPath,
		ShareData:        string(sharedataBytes),
		ShareDataExt:     string(sharedataextBytes),
		Status:           string(dagIns.Status),
		Reason:           dagIns.Reason,
		Cmd:              string(cmdBytes),
		HasCmd:           hasCmd,
		BatchRunID:       batchRunID,
		UserID:           dagIns.UserID,
		EndedAt:          dagIns.EndedAt,
		DagType:          dagIns.DagType,
		PolicyType:       dagIns.PolicyType,
		AppInfo:          string(appInfoBytes),
		Priority:         dagIns.Priority,
		Mode:             int(dagIns.Mode),
		Dump:             dagIns.Dump,
		DumpExt:          string(dumpextbytes),
		SuccessCallback:  dagIns.SuccessCallback,
		ErrorCallback:    dagIns.ErrorCallback,
		CallChain:        string(callchainBytes),
		ResumeData:       dagIns.ResumeData,
		ResumeStatus:     string(dagIns.ResumeStatus),
		Version:          dagIns.Version.ToString(),
		VersionID:        dagIns.VersionID,
		BizDomainID:      dagIns.BizDomainID,
	}
}

// ToTaskInstanceModel 转换为TaskInstanceModel
func ToTaskInstanceModel(taskIns *entity.TaskInstance, isupdate bool) *TaskInstanceModel {
	if isupdate {
		taskIns.Update()
	} else {
		taskIns.Initial()
	}

	id, _ := strconv.ParseUint(taskIns.ID, 10, 64)
	dagInsID, _ := strconv.ParseUint(taskIns.DagInsID, 10, 64)

	return &TaskInstanceModel{
		ID:             id,
		CreatedAt:      taskIns.CreatedAt,
		UpdatedAt:      taskIns.UpdatedAt,
		TaskID:         taskIns.TaskID,
		DagInsID:       dagInsID,
		Name:           taskIns.Name,
		DependOn:       marshalToString(taskIns.DependOn),
		ActionName:     taskIns.ActionName,
		TimeoutSecs:    taskIns.TimeoutSecs,
		Params:         marshalToString(taskIns.Params),
		Traces:         marshalToString(taskIns.Traces),
		Status:         string(taskIns.Status),
		Reason:         marshalToString(taskIns.Reason),
		PreChecks:      marshalToString(taskIns.PreChecks),
		Results:        marshalToString(taskIns.Results),
		Steps:          marshalToString(taskIns.Steps),
		LastModifiedAt: taskIns.LastModifiedAt,
		RenderedParams: marshalToString(taskIns.RenderedParams),
		Hash:           taskIns.Hash,
		Settings:       marshalToString(taskIns.Settings),
		MetaData:       marshalToString(taskIns.MetaData),
	}
}

// ToDagVersionModel 转换为DagVersionModel
func ToDagVersionModel(dagVersion *entity.DagVersion) *DagVersionModel {
	dagVersion.Initial()
	id, _ := strconv.ParseUint(dagVersion.ID, 10, 64)

	return &DagVersionModel{
		ID:        id,
		CreatedAt: dagVersion.CreatedAt,
		UpdatedAt: dagVersion.UpdatedAt,
		DagID:     dagVersion.DagID,
		UserID:    dagVersion.UserID,
		Version:   dagVersion.Version.ToString(),
		VersionID: dagVersion.VersionID,
		ChangeLog: dagVersion.ChangeLog,
		Config:    string(dagVersion.Config),
		SortTime:  dagVersion.SortTime,
	}
}

// ToInboxModel 转换为InboxModel
func ToInboxModel(inbox *entity.InBox) *InBoxModel {
	inbox.Initial()
	id, _ := strconv.ParseUint(inbox.ID, 10, 64)
	msg, _ := jsoniter.MarshalToString(inbox.Msg)
	dags, _ := jsoniter.MarshalToString(inbox.Dags)

	return &InBoxModel{
		ID:        id,
		CreatedAt: inbox.CreatedAt,
		UpdatedAt: inbox.UpdatedAt,
		Msg:       msg,
		Topic:     inbox.Topic,
		DocID:     inbox.DocID,
		Dags:      dags,
	}
}

// ToEntity 将src转换为dest
func ToEntity(src, dest interface{}) error {
	return copyFields(src, dest)
}

func marshalToString(val interface{}) string {
	if val == nil {
		return ""
	}
	bytes, err := json.Marshal(val)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func copyFields(src interface{}, dest interface{}) error {
	srcVal := reflect.ValueOf(src)
	destVal := reflect.ValueOf(dest)

	if srcVal.Kind() != reflect.Ptr || destVal.Kind() != reflect.Ptr {
		return fmt.Errorf("both src and dest must be pointers")
	}

	srcVal = srcVal.Elem()
	destVal = destVal.Elem()

	if srcVal.Kind() != reflect.Struct || destVal.Kind() != reflect.Struct {
		return fmt.Errorf("both src and dest must be structs")
	}

	srcType := srcVal.Type()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		srcFieldType := srcType.Field(i)

		destField := destVal.FieldByName(srcFieldType.Name)
		if !destField.IsValid() || !destField.CanSet() {
			continue
		}

		if converted, ok := convertFieldValue(srcField, destField.Type()); ok {
			destField.Set(converted)
		}
	}

	return nil
}

func convertFieldValue(srcField reflect.Value, destType reflect.Type) (reflect.Value, bool) {
	if !srcField.IsValid() {
		return reflect.Value{}, false
	}

	// 类型完全一致直接赋值，但 string -> interface{} 需要优先尝试 JSON 解析。
	if srcField.Type().AssignableTo(destType) &&
		!(isStringKind(srcField) && destType.Kind() == reflect.Interface) {
		return srcField, true
	}

	// 指针转换通过递归复用同一套规则，避免分支重复。
	if srcField.Kind() == reflect.Ptr {
		if srcField.IsNil() {
			return reflect.Value{}, false
		}

		if destType.Kind() == reflect.Ptr {
			converted, ok := convertFieldValue(srcField.Elem(), destType.Elem())
			if !ok {
				return reflect.Value{}, false
			}

			newVal := reflect.New(destType.Elem())
			newVal.Elem().Set(converted)
			return newVal, true
		}

		return convertFieldValue(srcField.Elem(), destType)
	}

	if destType.Kind() == reflect.Ptr {
		converted, ok := convertFieldValue(srcField, destType.Elem())
		if !ok {
			return reflect.Value{}, false
		}

		newVal := reflect.New(destType.Elem())
		newVal.Elem().Set(converted)
		return newVal, true
	}

	// 数值 -> 字符串（含自定义字符串类型）
	if isNumeric(srcField) && destType.Kind() == reflect.String {
		strVal := numericToString(srcField)
		return reflect.ValueOf(strVal).Convert(destType), true
	}

	// 字符串 -> 数值
	if isStringKind(srcField) && isNumericKind(destType.Kind()) {
		if converted, ok := stringToNumeric(srcField.String(), destType); ok {
			return converted, true
		}
	}

	// 字符串 -> 复杂类型 / interface{} 的 JSON 反序列化
	if isStringKind(srcField) && srcField.String() != "" {
		if converted, ok := unmarshalJSONStringToType(srcField.String(), destType); ok {
			return converted, true
		}
	}

	// 底层可转换类型（string ↔ MyStr, int ↔ MyInt 等）
	if safeConvertible(srcField.Type(), destType) {
		return srcField.Convert(destType), true
	}

	return reflect.Value{}, false
}

func unmarshalJSONStringToType(raw string, destType reflect.Type) (reflect.Value, bool) {
	if destType.Kind() == reflect.Interface {
		var parsed interface{}
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return reflect.Value{}, false
		}
		if parsed == nil {
			return reflect.Zero(destType), true
		}
		return reflect.ValueOf(parsed), true
	}

	if destType.Kind() == reflect.Ptr {
		destInstancePtr := reflect.New(destType.Elem())
		if err := json.Unmarshal([]byte(raw), destInstancePtr.Interface()); err != nil {
			return reflect.Value{}, false
		}
		return destInstancePtr, true
	}

	destInstancePtr := reflect.New(destType)
	if err := json.Unmarshal([]byte(raw), destInstancePtr.Interface()); err != nil {
		return reflect.Value{}, false
	}
	return destInstancePtr.Elem(), true
}

// isNumeric 判断是否为数值类型（包括自定义数值类型如 type MyInt int）
func isNumeric(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// isStringKind 判断底层是否为字符串类型（包括 type MyStr string）
func isStringKind(v reflect.Value) bool {
	return v.Kind() == reflect.String
}

// numericToString 将数值转为字符串
func numericToString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	}
	return fmt.Sprintf("%v", v.Interface())
}

// stringToNumeric 将字符串转为目标数值类型
func stringToNumeric(s string, targetType reflect.Type) (reflect.Value, bool) {
	// 获取底层 Kind
	kind := targetType.Kind()

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		val := reflect.New(targetType).Elem()
		val.SetInt(n)
		return val, true

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		val := reflect.New(targetType).Elem()
		val.SetUint(n)
		return val, true

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		val := reflect.New(targetType).Elem()
		val.SetFloat(n)
		return val, true

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return reflect.Value{}, false
		}
		val := reflect.New(targetType).Elem()
		val.SetBool(b)
		return val, true
	}

	return reflect.Value{}, false
}

// safeConvertible 安全的类型转换判断，排除整数→字符串的 Unicode 码点误转换
func safeConvertible(srcType, destType reflect.Type) bool {
	if !srcType.ConvertibleTo(destType) {
		return false
	}

	srcKind := srcType.Kind()
	destKind := destType.Kind()

	// 排除: 整数 → 字符串（Go 会把整数当 rune 转换，不是我们要的行为）
	if isIntKind(srcKind) && destKind == reflect.String {
		return false
	}
	// 排除: 字符串 → 整数（Convert 本身也不支持，但以防万一）
	if srcKind == reflect.String && isIntKind(destKind) {
		return false
	}

	return true
}

func isIntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func isNumericKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func updateJSONMapString(raw, key string, val any) (string, error) {
	m := map[string]any{}
	if raw != "" {
		if err := jsoniter.UnmarshalFromString(raw, &m); err != nil {
			return "", err
		}
	}
	m[key] = val
	bytes, err := jsoniter.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func parseUint64Slice(ids []string) []uint64 {
	var res []uint64
	for _, id := range ids {
		v, _ := strconv.ParseUint(id, 10, 64)
		res = append(res, v)
	}
	return res
}
