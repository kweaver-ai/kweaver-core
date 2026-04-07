// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package worker

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/mohae/deepcopy"

	"bkn-backend/common"
	cond "bkn-backend/common/condition"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
)

type VectorProperty struct {
	Name           string
	VectorField    string
	Model          *interfaces.SmallModel
	AllVectorResps []*cond.VectorResp
	Err            error
}

type ObjectTypeTask struct {
	appSetting *common.AppSetting
	dva        interfaces.DataViewAccess
	mfa        interfaces.ModelFactoryAccess
	ja         interfaces.JobAccess
	osa        interfaces.OpenSearchAccess

	ViewDataLimit    int
	JobMaxRetryTimes int
	taskInfo         *interfaces.TaskInfo
	objectType       *interfaces.ObjectType
	objectTypeStatus *interfaces.ObjectTypeStatus

	propertyMapping  map[string]*interfaces.Field
	vectorProperties []*VectorProperty
	totalCount       int64
	currentCount     int64

	incField *interfaces.Field
}

func NewObjectTypeTask(appSetting *common.AppSetting, taskInfo *interfaces.TaskInfo,
	objectType *interfaces.ObjectType) *ObjectTypeTask {

	return &ObjectTypeTask{
		appSetting: appSetting,
		dva:        logics.DVA,
		mfa:        logics.MFA,
		ja:         logics.JA,
		osa:        logics.OSA,

		ViewDataLimit:    appSetting.ServerSetting.ViewDataLimit,
		JobMaxRetryTimes: appSetting.ServerSetting.JobMaxRetryTimes,
		taskInfo:         taskInfo,
		objectType:       objectType,
		objectTypeStatus: &interfaces.ObjectTypeStatus{},

		propertyMapping:  make(map[string]*interfaces.Field),
		vectorProperties: make([]*VectorProperty, 0),
		totalCount:       0,
		currentCount:     0,

		incField: nil,
	}
}

func (ott *ObjectTypeTask) GetTaskInfo() *interfaces.TaskInfo {
	return ott.taskInfo
}

func (ott *ObjectTypeTask) HandleObjectTypeTask(ctx context.Context, jobInfo *interfaces.JobInfo,
	taskInfo *interfaces.TaskInfo, objectType *interfaces.ObjectType) error {

	startTime := time.Now()
	logger.Infof("开始处理 object type %s", objectType.OTID)

	dataSource := objectType.DataSource
	if dataSource.Type != "data_view" {
		logger.Warnf("data source type %s is not data_view", dataSource.Type)
		return nil
	}

	for _, property := range objectType.DataProperties {
		// 未配置映射或者映射的字段名为空
		if property.MappedField == nil || property.MappedField.Name == "" {
			continue
		}
		_, ok := interfaces.KN_INDEX_PROP_TYPE_MAPPING[property.Type]
		if !ok {
			logger.Errorf("Unknown property type %s", property.Type)
			continue
		}

		ott.propertyMapping[property.Name] = property.MappedField
		if property.Type == "varchar" || property.Type == "string" || property.Type == "text" {
			if property.IndexConfig != nil && property.IndexConfig.VectorConfig.Enabled {

				model, err := ott.mfa.GetModelByID(ctx, property.IndexConfig.VectorConfig.ModelID)
				if err != nil {
					return err
				}
				if model == nil {
					return fmt.Errorf("failed to get small model by id '%s'", property.IndexConfig.VectorConfig.ModelID)
				}

				ott.vectorProperties = append(ott.vectorProperties, &VectorProperty{
					Name:           property.Name,
					VectorField:    "_vector_" + property.Name,
					Model:          model,
					AllVectorResps: make([]*cond.VectorResp, 0),
				})
			}
		}
	}

	// 判断主键是否有映射
	if len(ott.objectType.PrimaryKeys) == 0 {
		return fmt.Errorf("ott.objectType is None or len of PrimaryKeys is 0, taskInfo:%v", ott.taskInfo)
	}
	for _, pk := range objectType.PrimaryKeys {
		if _, exist := ott.propertyMapping[pk]; !exist {
			return fmt.Errorf("primary key %s unmapped", pk)
		}
	}

	incFieldName := ""
	var incFieldValue any
	if jobInfo.JobType == interfaces.JobTypeIncremental {
		if objectType.IncrementalKey != "" {
			if field, ok := ott.propertyMapping[objectType.IncrementalKey]; ok {
				ott.objectTypeStatus.IncrementalKey = objectType.IncrementalKey
				ott.incField = field
				incFieldName = field.Name

				if objectType.Status.IndexAvailable && objectType.Status.IncrementalKey == objectType.IncrementalKey {
					ott.objectTypeStatus.IncrementalValue = objectType.Status.IncrementalValue
					ott.objectTypeStatus.Index = objectType.Status.Index

					switch field.Type {
					case "integer":
						incFieldValue, _ = strconv.ParseInt(ott.objectTypeStatus.IncrementalValue, 10, 64)
					case "datetime", "timestamp":
						incFieldValue = ott.objectTypeStatus.IncrementalValue
					default:
						err := fmt.Errorf("unsupported incremental key type %s", field.Type)
						return err
					}
				}
			}
		}
	}
	if ott.objectTypeStatus.Index == "" {
		ott.objectTypeStatus.Index = ott.generateTaskIndexName(jobInfo.KNID, jobInfo.Branch, objectType.OTID, taskInfo.ID)
		err := ott.handlerIndex(ctx, ott.objectTypeStatus.Index, objectType)
		if err != nil {
			return err
		}
	}

	dataView, err := ott.dva.GetDataViewByID(ctx, dataSource.ID)
	if err != nil {
		return err
	}
	if dataView == nil {
		return fmt.Errorf("object type %s's data view %s not found", objectType.OTName, dataSource.ID)
	}

	currentStartTime := time.Now()
	viewQueryResult, err := ott.dva.GetDataStart(ctx, dataView.ViewID, incFieldName,
		incFieldValue, ott.ViewDataLimit)
	if err != nil {
		logger.Errorf("从 %s 读取第一批数据失败: %s", dataView.ViewName, err.Error())
		return err
	}

	ott.totalCount = viewQueryResult.TotalCount
	logger.Infof("从 %s 读取第一批数据, 总条数：%d, 当前条数：%d, 耗时：%dms, searchAfter: %v, 进度：%d/%d",
		dataView.ViewID, ott.totalCount, len(viewQueryResult.Entries), time.Since(currentStartTime).Milliseconds(),
		viewQueryResult.SearchAfter, ott.currentCount, ott.totalCount)

	stateInfo := interfaces.TaskStateInfo{
		Index:    ott.objectTypeStatus.Index,
		DocCount: viewQueryResult.TotalCount,
	}
	err = ott.ja.UpdateTaskState(ctx, taskInfo.ID, stateInfo)
	if err != nil {
		logger.Errorf("更新 task %s 状态失败: %s", taskInfo.ID, err.Error())
		return err
	}

	err = ott.handlerIndexData(ctx, viewQueryResult)
	if err != nil {
		logger.Errorf("从视图 %s 读取第一批数据后写入第一批索引数据失败, 失败原因: %s",
			dataView.ViewName, err.Error())
		return err
	}
	ott.currentCount += int64(len(viewQueryResult.Entries))

	logger.Infof("从 %s 读取第一批数据并处理完成, 总条数：%d, 当前条数：%d, 耗时：%dms, searchAfter: %v, 进度：%d/%d",
		dataView.ViewID, ott.totalCount, len(viewQueryResult.Entries), time.Since(currentStartTime).Milliseconds(),
		viewQueryResult.SearchAfter, ott.currentCount, ott.totalCount)

	if ott.incField != nil && len(viewQueryResult.Entries) > 0 {
		lastEntry := viewQueryResult.Entries[len(viewQueryResult.Entries)-1]
		if value, ok := lastEntry[ott.incField.Name]; ok {
			switch ott.incField.Type {
			case "integer":
				ott.objectTypeStatus.IncrementalValue = strconv.FormatInt(value.(int64), 10)
			case "datetime", "timestamp":
				ott.objectTypeStatus.IncrementalValue = value.(string)
			default:
				ott.objectTypeStatus.IncrementalValue = fmt.Sprintf("%v", value)
			}
		}
	}

	for len(viewQueryResult.SearchAfter) > 0 {
		currentStartTime := time.Now()
		viewQueryResult, err = ott.dva.GetDataNext(ctx, dataView.ViewID,
			viewQueryResult.SearchAfter, ott.ViewDataLimit)
		if err != nil {
			logger.Errorf("从 %s 分批读取数据失败: %s", dataView.ViewName, err.Error())
			return err
		}
		logger.Infof("从 %s 分批读取数据, 总条数：%d, 当前条数：%d, 耗时：%dms, searchAfter: %v, 进度：%d/%d",
			dataView.ViewName, ott.totalCount, len(viewQueryResult.Entries), time.Since(currentStartTime).Milliseconds(),
			viewQueryResult.SearchAfter, ott.currentCount, ott.totalCount)

		err = ott.handlerIndexData(ctx, viewQueryResult)
		if err != nil {
			logger.Errorf("从视图 %s 分批读取数据后写入索引数据失败, 总条数：%d, 当前条数：%d, 失败原因: %s",
				dataView.ViewName, ott.totalCount, len(viewQueryResult.Entries), err.Error())
			return err
		}
		ott.currentCount += int64(len(viewQueryResult.Entries))

		logger.Infof("从 %s 分批读取数据并处理完成, 总条数：%d, 当前条数：%d, 耗时：%dms, searchAfter: %v, 进度：%d/%d",
			dataView.ViewName, ott.totalCount, len(viewQueryResult.Entries), time.Since(currentStartTime).Milliseconds(),
			viewQueryResult.SearchAfter, ott.currentCount, ott.totalCount)

		if ott.incField != nil && len(viewQueryResult.Entries) > 0 {
			lastEntry := viewQueryResult.Entries[len(viewQueryResult.Entries)-1]
			if value, ok := lastEntry[ott.incField.Name]; ok {
				switch ott.incField.Type {
				case "integer":
					ott.objectTypeStatus.IncrementalValue = strconv.FormatInt(value.(int64), 10)
				case "datetime", "timestamp":
					ott.objectTypeStatus.IncrementalValue = value.(string)
				default:
					ott.objectTypeStatus.IncrementalValue = fmt.Sprintf("%v", value)
				}
			}
		}
	}

	err = ott.osa.Refresh(ctx, ott.objectTypeStatus.Index)
	if err != nil {
		logger.Errorf("Refresh err:%v", err)
		return err
	}

	indexStats, err := ott.osa.GetIndexStats(ctx, ott.objectTypeStatus.Index)
	if err != nil {
		logger.Errorf("GetIndexStats err:%v", err)
		return err
	}
	ott.objectTypeStatus.IndexAvailable = true
	ott.objectTypeStatus.DocCount = indexStats.DocCount
	ott.objectTypeStatus.StorageSize = indexStats.StorageSize
	ott.objectTypeStatus.UpdateTime = time.Now().UnixMilli()

	if jobInfo.JobType == interfaces.JobTypeFull && ott.totalCount != indexStats.DocCount {
		logger.Warnf("处理 object type %s 完成, 但 totalCount %d != indexStats.DocCount %d", ott.objectType.OTID, ott.totalCount, indexStats.DocCount)
	}

	logger.Infof("处理 object type %s 完成, 总条数：%d, 当前条数：%d, 耗时：%dms",
		objectType.OTID, ott.totalCount, ott.currentCount, time.Since(startTime).Milliseconds())
	return nil
}

func (ott *ObjectTypeTask) handlerIndex(ctx context.Context, index string, objectType *interfaces.ObjectType) error {
	logger.Debugf("handlerIndex: %v", index)

	exists, err := ott.osa.IndexExists(ctx, index)
	if err != nil {
		logger.Errorf("CheckKNConceptIndexExists err:%v", err)
		return err
	}
	if exists {
		err = ott.osa.DeleteIndex(ctx, index)
		if err != nil {
			logger.Errorf("DeleteKNConceptIndex err:%v", err)
			return err
		}
	}

	propertiesMap := map[string]any{}
	for _, property := range objectType.DataProperties {
		propConfig, ok := interfaces.KN_INDEX_PROP_TYPE_MAPPING[property.Type]
		if !ok {
			logger.Errorf("Unknown property type %s", property.Type)
			continue
		}
		propConfig = deepcopy.Copy(propConfig)

		if property.IndexConfig != nil {
			switch property.Type {
			case "string", "varchar", "keyword":
				if property.IndexConfig.KeywordConfig.Enabled {
					propConfig.(map[string]any)["ignore_above"] = property.IndexConfig.KeywordConfig.IgnoreAboveLen
				}
				if property.IndexConfig.FulltextConfig.Enabled {
					textPropConfig := deepcopy.Copy(interfaces.KN_INDEX_PROP_TYPE_MAPPING["text"])
					textPropConfig.(map[string]any)["analyzer"] = property.IndexConfig.FulltextConfig.Analyzer
					propConfig.(map[string]any)["fields"] = map[string]any{
						"text": textPropConfig,
					}
				}
			case "text":
				if property.IndexConfig.FulltextConfig.Enabled {
					propConfig.(map[string]any)["analyzer"] = property.IndexConfig.FulltextConfig.Analyzer
				}
				if property.IndexConfig.KeywordConfig.Enabled {
					keywordPropConfig := deepcopy.Copy(interfaces.KN_INDEX_PROP_TYPE_MAPPING["keyword"])
					keywordPropConfig.(map[string]any)["ignore_above"] = property.IndexConfig.KeywordConfig.IgnoreAboveLen
					propConfig.(map[string]any)["fields"] = map[string]any{
						"keyword": keywordPropConfig,
					}
				}
			}
		}

		propertiesMap[property.Name] = propConfig
	}

	for _, prop := range ott.vectorProperties {
		propVectoConfig := deepcopy.Copy(interfaces.KN_INDEX_PROP_TYPE_MAPPING["vector"])
		propVectoConfig.(map[string]any)["dimension"] = prop.Model.EmbeddingDim
		propertiesMap[prop.VectorField] = propVectoConfig
	}

	indexBody := map[string]any{
		"settings": interfaces.KN_INDEX_SETTINGS,
		"mappings": map[string]any{
			"dynamic_templates": interfaces.KN_INDEX_DYNAMIC_TEMPLATES,
			"properties":        propertiesMap,
		},
	}
	err = ott.osa.CreateIndex(ctx, index, indexBody)
	if err != nil {
		logger.Errorf("CreateKNConceptIndex err:%v", err)
		return err
	}
	return nil
}

func (ott *ObjectTypeTask) handlerIndexData(ctx context.Context, viewQueryResult *interfaces.ViewQueryResult) error {
	if len(viewQueryResult.Entries) == 0 {
		return nil
	}

	newEntries := make([]any, 0, len(viewQueryResult.Entries))
	for _, entry := range viewQueryResult.Entries {
		newEntry := map[string]any{}
		// propertyMapping 是属性到视图字段的对应
		for k, v := range ott.propertyMapping {
			if entry[v.Name] != nil { // 数据不为空，才写入opensearch
				newEntry[k] = entry[v.Name]
			}
		}
		//  handler __ID
		objectID := ott.GetObjectID(entry)
		newEntry[interfaces.OBJECT_ID] = objectID

		newEntries = append(newEntries, newEntry)
	}

	if len(ott.vectorProperties) > 0 {
		var wg sync.WaitGroup
		for _, property := range ott.vectorProperties {
			wg.Add(1)
			go func(prop *VectorProperty) {
				defer wg.Done()
				for retry := 0; retry < ott.JobMaxRetryTimes; retry++ {
					err := ott.handlerVector(ctx, prop, newEntries)
					if err != nil {
						logger.Errorf("handlerVector err:%v, retry times %d", err, retry)
						prop.Err = err
					} else {
						prop.Err = nil
						break
					}
				}
			}(property)
		}
		wg.Wait()

		for _, property := range ott.vectorProperties {
			if property.Err != nil {
				return property.Err
			}
			for i, entry := range newEntries {
				if property.AllVectorResps[i] != nil {
					entry.(map[string]any)[property.VectorField] = property.AllVectorResps[i].Vector
				}
			}
		}
	}

	// todo 分批 block 100m
	err := ott.osa.BulkInsertData(ctx, ott.objectTypeStatus.Index, newEntries)
	if err != nil {
		return err
	}
	return nil
}

func (ott *ObjectTypeTask) handlerVector(ctx context.Context, property *VectorProperty, newEntries []any) error {
	words := make([]string, 0, len(newEntries))
	validIdxs := make([]int, 0, len(newEntries))
	for idx, entry := range newEntries {
		if str, exist := entry.(map[string]any)[property.Name]; exist && str != nil {
			word := str.(string)
			words = append(words, word)
			validIdxs = append(validIdxs, idx)
		}
	}

	allVectorResps := make([]*cond.VectorResp, len(newEntries))
	vectorResps, err := ott.mfa.GetVector(ctx, property.Model, words)
	if err != nil {
		return err
	}
	for i, idx := range validIdxs {
		allVectorResps[idx] = vectorResps[i]
	}

	property.AllVectorResps = allVectorResps
	return nil
}

// 从对象数据中提取对象ID
func (ott *ObjectTypeTask) GetObjectID(objectData map[string]any) string {
	idStr := ""
	// 使用主键构建对象ID
	var idParts []string
	for _, pk := range ott.objectType.PrimaryKeys {
		if value, exists := objectData[ott.propertyMapping[pk].Name]; exists {
			idParts = append(idParts, fmt.Sprintf("%v", value))
		} else {
			idParts = append(idParts, "__NULL__")
		}
	}

	idStr = strings.Join(idParts, "-")

	// id: md5(主键值-主键值-...)
	md5Hasher := md5.New()
	md5Hasher.Write([]byte(idStr))
	hashed := md5Hasher.Sum(nil)

	return hex.EncodeToString(hashed)
}

func (ott *ObjectTypeTask) generateTaskIndexName(knID string, branch string, otID string, task_id string) string {
	// adp-kn_<kn_id>_<object_type_id>_<task_id>
	return fmt.Sprintf("adp-kn_ot_index-%s-%s-%s-%s", knID, branch, otID, task_id)
}
