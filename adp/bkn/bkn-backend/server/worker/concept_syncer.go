// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package worker

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/mitchellh/mapstructure"

	"bkn-backend/common"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
)

var (
	cSyncerOnce sync.Once
	cSyncer     *ConceptSyncer
)

type ConceptSyncer struct {
	appSetting *common.AppSetting
	ata        interfaces.ActionTypeAccess
	cga        interfaces.ConceptGroupAccess
	mfa        interfaces.ModelFactoryAccess
	kna        interfaces.KNAccess
	vba        interfaces.VegaBackendAccess
	ota        interfaces.ObjectTypeAccess
	rta        interfaces.RelationTypeAccess
	riskTypeA  interfaces.RiskTypeAccess
	ma         interfaces.MetricAccess
}

func NewConceptSyncer(appSetting *common.AppSetting) *ConceptSyncer {
	cSyncerOnce.Do(func() {
		cSyncer = &ConceptSyncer{
			appSetting: appSetting,
			ata:        logics.ATA,
			mfa:        logics.MFA,
			kna:        logics.KNA,
			cga:        logics.CGA,
			vba:        logics.VBA,
			ota:        logics.OTA,
			rta:        logics.RTA,
			riskTypeA:  logics.RiskTypeAccess,
			ma:         logics.MA,
		}
	})
	return cSyncer
}

// KNDetailInfo 知识网络详情信息结构
type KNDetailInfo struct {
	NetworkInfo   map[string]any `json:"network_info"`
	ObjectTypes   []SimpleItem   `json:"object_types"`
	RelationTypes []SimpleItem   `json:"relation_types"`
	ActionTypes   []SimpleItem   `json:"action_types"`
	ConceptGroups []SimpleItem   `json:"concept_groups"`
}

// SimpleItem 简化项结构，仅保留id、name、tag、comment字段
type SimpleItem struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Tags    []string `json:"tags"`
	Comment string   `json:"comment"`

	// for relation types
	SourceObjectTypeName string `json:"source_object_type_name,omitempty"`
	TargetObjectTypeName string `json:"target_object_type_name,omitempty"`

	// for action types
	ObjectTypeName string `json:"object_type_name,omitempty"`
}

// GeneratorTicker 生成业务知识网络详情定时任务
func (cs *ConceptSyncer) Start() {
	for {
		err := cs.handleKNs()
		if err != nil {
			logger.Errorf("[handleKNs] Failed: %v", err)
		}
		time.Sleep(5 * time.Minute)
	}
}

// handleKNs 处理业务知识网络详情 todo：补充 对象类、关系类、行动类的detail，并且要更新概念索引
func (cs *ConceptSyncer) handleKNs() error {
	defer func() {
		if rerr := recover(); rerr != nil {
			logger.Errorf("[handleKNs] Failed: %v", rerr)
			debug.PrintStack()
			return
		}
	}()

	logger.Debug("[handleKNs] Start")

	ctx := context.Background()

	knsInDB, err := cs.kna.GetAllKNs(ctx)
	if err != nil {
		logger.Errorf("Failed to list knowledge networks: %v", err)
		return err
	}

	knsInDataset, err := cs.getAllKNsFromDataset(ctx)
	if err != nil {
		logger.Errorf("Failed to list knowledge networks in dataset: %v", err)
		return err
	}

	for _, knInDB := range knsInDB {
		need_update := false
		knInDataset, exist := knsInDataset[knInDB.KNID]
		if !exist {
			need_update = true
		} else if knInDB.UpdateTime != knInDataset.UpdateTime {
			need_update = true
		}

		err := cs.handleKnowledgeNetwork(ctx, knInDB, need_update)
		if err != nil {
			logger.Errorf("Failed to handle knowledge network %s (%s %s): %v", knInDB.KNName, knInDB.KNID, knInDB.Branch, err)
			continue
		}
	}

	logger.Info("handle KNs completed")
	return nil
}

// handleKnowledgeNetwork 处理单个知识网络
func (cs *ConceptSyncer) handleKnowledgeNetwork(ctx context.Context, kn *interfaces.KN, need_update bool) error {
	logger.Debugf("Handle knowledge network: %s (%s %s), %s", kn.KNName, kn.KNID, kn.Branch)

	// 获取对象类型列表
	objectTypes, ot_need_update, err := cs.handleObjectTypes(ctx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("Failed to handle object types %s %s: %v", kn.KNID, kn.Branch, err)
		return err
	}

	// 获取关系类型列表
	relationTypes, rt_need_update, err := cs.handleRelationTypes(ctx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("Failed to handle relation types %s %s: %v", kn.KNID, kn.Branch, err)
		return err
	}

	// 获取行动类型列表
	actionTypes, at_need_update, err := cs.handleActionTypes(ctx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("Failed to handle action types %s %s: %v", kn.KNID, kn.Branch, err)
		return err
	}

	conceptGroups, cg_need_update, err := cs.handleConceptGroups(ctx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("Failed to handle concept groups %s %s: %v", kn.KNID, kn.Branch, err)
		return err
	}

	riskTypes, rtRisk_need_update, err := cs.handleRiskTypes(ctx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("Failed to handle risk types %s %s: %v", kn.KNID, kn.Branch, err)
		return err
	}

	_, metric_need_update, err := cs.handleMetrics(ctx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("Failed to handle metrics %s %s: %v", kn.KNID, kn.Branch, err)
		return err
	}

	if !need_update && !ot_need_update && !rt_need_update && !at_need_update && !cg_need_update && !rtRisk_need_update && !metric_need_update {
		logger.Debugf("Knowledge network %s (%s %s) does not need update", kn.KNName, kn.KNID, kn.Branch)
		return nil
	}

	kn.ObjectTypes = objectTypes
	kn.RelationTypes = relationTypes
	kn.ActionTypes = actionTypes
	kn.ConceptGroups = conceptGroups
	kn.RiskTypes = riskTypes

	bknNetwork := logics.ToBKNNetWork(kn)
	kn.BKNRawContent = bknsdk.SerializeBknNetwork(bknNetwork)

	// 更新知识网络详情
	err = cs.kna.UpdateKNDetail(ctx, kn.KNID, kn.Branch, kn.BKNRawContent)
	if err != nil {
		logger.Errorf("Failed to update KN detail for %s (%s %s): %v", kn.KNName, kn.KNID, kn.Branch, err)
		return err
	}

	err = cs.insertDatasetDataForKN(ctx, kn)
	if err != nil {
		logger.Errorf("Failed to insert dataset data for KN %s (%s %s): %v", kn.KNName, kn.KNID, kn.Branch, err)
		return err
	}

	logger.Debugf("Generated KN detail for %s (%s %s): %s", kn.KNName, kn.KNID, kn.Branch, kn.BKNRawContent)
	return nil
}

// handleObjectTypes 获取知识网络的对象类型
func (cs *ConceptSyncer) handleObjectTypes(ctx context.Context, knID string,
	branch string) ([]*interfaces.ObjectType, bool, error) {

	logger.Debugf("Handle object types for knowledge network %s %s", knID, branch)
	objectTypesInDB, err := cs.ota.GetAllObjectTypesByKnID(ctx, knID, branch)
	if err != nil {
		return []*interfaces.ObjectType{}, false, err
	}

	objectTypesInDataset, err := cs.getAllObjectTypesFromDatasetByKnID(ctx, knID, branch)
	if err != nil {
		return []*interfaces.ObjectType{}, false, err
	}

	need_update := false
	add_list := []*interfaces.ObjectType{}
	for _, otInDB := range objectTypesInDB {
		otInDataset, exist := objectTypesInDataset[otInDB.OTID]
		if !exist {
			add_list = append(add_list, otInDB)
		} else if otInDB.UpdateTime != otInDataset.UpdateTime {
			add_list = append(add_list, otInDB)
		}
		// todo: DB里没有，dataset里有的，需要删除dataset里的数据？
	}
	if len(add_list) > 0 {
		logger.Debugf("Need add (%d) object types to dataset", len(add_list))
		need_update = true
	}

	err = cs.insertDatasetDataForObjectTypes(ctx, add_list)
	if err != nil {
		return []*interfaces.ObjectType{}, false, err
	}

	arrObjectTypes := make([]*interfaces.ObjectType, 0, len(objectTypesInDB))
	for _, otInDB := range objectTypesInDB {
		arrObjectTypes = append(arrObjectTypes, otInDB)
	}

	logger.Debugf("Handle object types for knowledge network %s %s done", knID, branch)
	return arrObjectTypes, need_update, nil
}

// handleRelationTypes 获取知识网络的关系类型
func (cs *ConceptSyncer) handleRelationTypes(ctx context.Context, knID string,
	branch string) ([]*interfaces.RelationType, bool, error) {
	logger.Debugf("Handle relation types for knowledge network %s %s", knID, branch)
	relationTypesInDB, err := cs.rta.GetAllRelationTypesByKnID(ctx, knID, branch)
	if err != nil {
		return []*interfaces.RelationType{}, false, err
	}

	relationTypesInDataset, err := cs.getAllRelationTypesFromDatasetByKnID(ctx, knID, branch)
	if err != nil {
		return []*interfaces.RelationType{}, false, err
	}

	need_update := false
	add_list := []*interfaces.RelationType{}
	for _, rtInDB := range relationTypesInDB {
		rtInDataset, exist := relationTypesInDataset[rtInDB.RTID]
		if !exist {
			add_list = append(add_list, rtInDB)
		} else if rtInDB.UpdateTime != rtInDataset.UpdateTime {
			add_list = append(add_list, rtInDB)
		}
	}
	if len(add_list) > 0 {
		logger.Debugf("Need add (%d) relation types to dataset", len(add_list))
		need_update = true
	}

	err = cs.insertDatasetDataForRelationTypes(ctx, add_list)
	if err != nil {
		return []*interfaces.RelationType{}, false, err
	}

	arrRelationTypes := make([]*interfaces.RelationType, 0, len(relationTypesInDB))
	for _, rtInDB := range relationTypesInDB {
		arrRelationTypes = append(arrRelationTypes, rtInDB)
	}

	logger.Debugf("Handle relation types for knowledge network %s %s done", knID, branch)
	return arrRelationTypes, need_update, nil
}

// handleActionTypes 获取知识网络的行动类型
func (cs *ConceptSyncer) handleActionTypes(ctx context.Context, knID string,
	branch string) ([]*interfaces.ActionType, bool, error) {
	logger.Debugf("Handle action types for knowledge network %s %s", knID, branch)
	actionTypesInDB, err := cs.ata.GetAllActionTypesByKnID(ctx, knID, branch)
	if err != nil {
		return []*interfaces.ActionType{}, false, err
	}

	actionTypesInDataset, err := cs.getAllActionTypesFromDatasetByKnID(ctx, knID, branch)
	if err != nil {
		return []*interfaces.ActionType{}, false, err
	}

	need_update := false
	add_list := []*interfaces.ActionType{}
	for _, atInDB := range actionTypesInDB {
		atInDataset, exist := actionTypesInDataset[atInDB.ATID]
		if !exist {
			add_list = append(add_list, atInDB)
		} else if atInDB.UpdateTime != atInDataset.UpdateTime {
			add_list = append(add_list, atInDB)
		}
	}
	if len(add_list) > 0 {
		logger.Debugf("Need add (%d) action types to dataset", len(add_list))
		need_update = true
	}

	err = cs.insertDatasetDataForActionTypes(ctx, add_list)
	if err != nil {
		return []*interfaces.ActionType{}, false, err
	}

	arrActionTypes := make([]*interfaces.ActionType, 0, len(actionTypesInDB))
	for _, atInDB := range actionTypesInDB {
		arrActionTypes = append(arrActionTypes, atInDB)
	}

	logger.Debugf("Handle action types for knowledge network %s %s done", knID, branch)
	return arrActionTypes, need_update, nil
}

// handleRiskTypes 获取知识网络的风险类
func (cs *ConceptSyncer) handleRiskTypes(ctx context.Context, knID string, branch string) ([]*interfaces.RiskType, bool, error) {
	if cs.riskTypeA == nil {
		return []*interfaces.RiskType{}, false, nil
	}
	logger.Debugf("Handle risk types for knowledge network %s %s", knID, branch)
	riskTypesInDB, err := cs.riskTypeA.GetAllRiskTypesByKnID(ctx, knID, branch)
	if err != nil {
		return []*interfaces.RiskType{}, false, err
	}

	riskTypesInDataset, err := cs.getAllRiskTypesFromDatasetByKnID(ctx, knID, branch)
	if err != nil {
		return []*interfaces.RiskType{}, false, err
	}

	need_update := false
	add_list := []*interfaces.RiskType{}
	for _, rtInDB := range riskTypesInDB {
		rtInDataset, exist := riskTypesInDataset[rtInDB.RTID]
		if !exist {
			add_list = append(add_list, rtInDB)
		} else if rtInDB.UpdateTime != rtInDataset.UpdateTime {
			add_list = append(add_list, rtInDB)
		}
	}
	if len(add_list) > 0 {
		logger.Debugf("Need add (%d) risk types to dataset", len(add_list))
		need_update = true
	}

	err = cs.insertDatasetDataForRiskTypes(ctx, add_list)
	if err != nil {
		return []*interfaces.RiskType{}, false, err
	}

	arrRiskTypes := append([]*interfaces.RiskType{}, riskTypesInDB...)

	logger.Debugf("Handle risk types for knowledge network %s %s done", knID, branch)
	return arrRiskTypes, need_update, nil
}

// handleMetrics 同步指标到概念索引（与 handleObjectTypes 等一致：DB 与 dataset 按 id+update_time 比较）。
func (cs *ConceptSyncer) handleMetrics(ctx context.Context, knID, branch string) ([]*interfaces.MetricDefinition, bool, error) {
	if cs.ma == nil {
		return nil, false, nil
	}

	logger.Debugf("Handle metrics for knowledge network %s %s", knID, branch)
	metricsInDB, err := cs.ma.ListMetrics(ctx, interfaces.MetricsListQueryParams{
		KNID:   knID,
		Branch: branch,
	})
	if err != nil {
		return nil, false, err
	}

	metricsInDataset, err := cs.getAllMetricsFromDatasetByKnID(ctx, knID, branch)
	if err != nil {
		return nil, false, err
	}

	need_update := false
	addList := []*interfaces.MetricDefinition{}
	for _, mInDB := range metricsInDB {
		mInDataset, exist := metricsInDataset[mInDB.ID]
		if !exist {
			addList = append(addList, mInDB)
		} else if mInDB.UpdateTime != mInDataset.UpdateTime {
			addList = append(addList, mInDB)
		}
	}
	if len(addList) > 0 {
		logger.Debugf("Need add (%d) metrics to dataset", len(addList))
		need_update = true
	}

	err = cs.insertDatasetDataForMetrics(ctx, addList)
	if err != nil {
		return nil, false, err
	}

	out := make([]*interfaces.MetricDefinition, 0, len(metricsInDB))
	out = append(out, metricsInDB...)

	logger.Debugf("Handle metrics for knowledge network %s %s done", knID, branch)
	return out, need_update, nil
}

// handleConceptGroups 获取知识网络的概念组
func (cs *ConceptSyncer) handleConceptGroups(ctx context.Context, knID string,
	branch string) ([]*interfaces.ConceptGroup, bool, error) {

	logger.Debugf("Handle concept groups for knowledge network %s %s", knID, branch)
	conceptGroupsInDB, err := cs.cga.GetAllConceptGroupsByKnID(ctx, knID, branch)
	if err != nil {
		return []*interfaces.ConceptGroup{}, false, err
	}

	conceptGroupsInDataset, err := cs.getAllConceptGroupsFromDatasetByKnID(ctx, knID, branch)
	if err != nil {
		return []*interfaces.ConceptGroup{}, false, err
	}

	need_update := false
	add_list := []*interfaces.ConceptGroup{}
	for _, cgInDB := range conceptGroupsInDB {
		cgInDataset, exist := conceptGroupsInDataset[cgInDB.CGID]
		if !exist {
			add_list = append(add_list, cgInDB)
		} else if cgInDB.UpdateTime != cgInDataset.UpdateTime {
			add_list = append(add_list, cgInDB)
		}
	}
	if len(add_list) > 0 {
		logger.Debugf("Need add (%d) concept groups to dataset", len(add_list))
		need_update = true
	}

	err = cs.insertDatasetDataForConceptGroups(ctx, add_list)
	if err != nil {
		return []*interfaces.ConceptGroup{}, false, err
	}

	// 简化为仅保留id、name、tag、comment字段
	arrConceptGroups := make([]*interfaces.ConceptGroup, 0, len(conceptGroupsInDB))
	for _, cgInDB := range conceptGroupsInDB {
		arrConceptGroups = append(arrConceptGroups, cgInDB)
	}

	logger.Debugf("Handle concept groups for knowledge network %s %s done", knID, branch)
	return arrConceptGroups, need_update, nil
}

func (cs *ConceptSyncer) insertDatasetDataForKN(ctx context.Context, kn *interfaces.KN) error {
	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{kn.KNName}
		words = append(words, kn.Tags...)
		words = append(words, kn.Comment, kn.BKNRawContent)
		word := strings.Join(words, "\n")

		defaultModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, defaultModel, []string{word})
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}

		kn.Vector = vectors[0].Vector
	}

	docid := interfaces.GenerateConceptDocuemtnID(kn.KNID, interfaces.MODULE_TYPE_KN, kn.KNID, kn.Branch)
	kn.ModuleType = interfaces.MODULE_TYPE_KN

	// Convert to map for dataset
	docBytes, err := sonic.Marshal(kn)
	if err != nil {
		logger.Errorf("Failed to marshal KN: %s", err.Error())
		return err
	}

	var doc map[string]any
	if err := sonic.Unmarshal(docBytes, &doc); err != nil {
		logger.Errorf("Failed to unmarshal KN: %s", err.Error())
		return err
	}

	// Set document ID
	doc["_id"] = docid

	err = cs.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, []map[string]any{doc})
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		return err
	}

	return nil
}

func (cs *ConceptSyncer) insertDatasetDataForObjectTypes(ctx context.Context, objectTypes []*interfaces.ObjectType) error {
	if len(objectTypes) == 0 {
		return nil
	}

	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, objectType := range objectTypes {
			arr := []string{objectType.OTName}
			arr = append(arr, objectType.Tags...)
			arr = append(arr, objectType.Comment, objectType.BKNRawContent)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}

		if len(vectors) != len(objectTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(objectTypes), len(vectors))
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(objectTypes), len(vectors))
		}

		for i, objectType := range objectTypes {
			objectType.Vector = vectors[i].Vector
		}
	}

	documents := make([]map[string]any, 0, len(objectTypes))
	for _, objectType := range objectTypes {
		docid := interfaces.GenerateConceptDocuemtnID(objectType.KNID, interfaces.MODULE_TYPE_OBJECT_TYPE,
			objectType.OTID, objectType.Branch)
		objectType.ModuleType = interfaces.MODULE_TYPE_OBJECT_TYPE

		// Convert to map for dataset
		docBytes, err := sonic.Marshal(objectType)
		if err != nil {
			logger.Errorf("Failed to marshal ObjectType: %s", err.Error())
			return err
		}

		var doc map[string]any
		if err := sonic.Unmarshal(docBytes, &doc); err != nil {
			logger.Errorf("Failed to unmarshal ObjectType: %s", err.Error())
			return err
		}

		// Serialize logic_properties[].parameters to JSON string
		if logicProps, ok := doc["logic_properties"].([]any); ok {
			for _, lp := range logicProps {
				if lpMap, ok := lp.(map[string]any); ok {
					if params, exists := lpMap["parameters"]; exists {
						paramsBytes, err := sonic.Marshal(params)
						if err != nil {
							logger.Errorf("Failed to marshal logic_properties parameters: %s", err.Error())
							return err
						}
						lpMap["parameters"] = string(paramsBytes)
					}
				}
			}
		}

		// Set document ID
		doc["_id"] = docid
		documents = append(documents, doc)
	}

	err := cs.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, documents)
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		return err
	}

	return nil
}

func (cs *ConceptSyncer) insertDatasetDataForActionTypes(ctx context.Context, actionTypes []*interfaces.ActionType) error {
	if len(actionTypes) == 0 {
		return nil
	}

	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, actionType := range actionTypes {
			arr := []string{actionType.ATName}
			arr = append(arr, actionType.Tags...)
			arr = append(arr, actionType.Comment, actionType.BKNRawContent)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}

		if len(vectors) != len(actionTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(actionTypes), len(vectors))
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(actionTypes), len(vectors))
		}

		for i, actionType := range actionTypes {
			actionType.Vector = vectors[i].Vector
		}
	}

	documents := make([]map[string]any, 0, len(actionTypes))
	for _, actionType := range actionTypes {
		docid := interfaces.GenerateConceptDocuemtnID(actionType.KNID, interfaces.MODULE_TYPE_ACTION_TYPE,
			actionType.ATID, actionType.Branch)
		actionType.ModuleType = interfaces.MODULE_TYPE_ACTION_TYPE

		// Convert to map for dataset
		docBytes, err := sonic.Marshal(actionType)
		if err != nil {
			logger.Errorf("Failed to marshal ActionType: %s", err.Error())
			return err
		}

		var doc map[string]any
		if err := sonic.Unmarshal(docBytes, &doc); err != nil {
			logger.Errorf("Failed to unmarshal ActionType: %s", err.Error())
			return err
		}

		// Serialize parameters to JSON string
		if params, exists := doc["parameters"]; exists {
			paramsBytes, err := sonic.Marshal(params)
			if err != nil {
				logger.Errorf("Failed to marshal action_type parameters: %s", err.Error())
				return err
			}
			doc["parameters"] = string(paramsBytes)
		}

		// Serialize condition to JSON string
		if cond, exists := doc["condition"]; exists && cond != nil {
			condBytes, err := sonic.Marshal(cond)
			if err != nil {
				logger.Errorf("Failed to marshal action_type condition: %s", err.Error())
				return err
			}
			doc["condition"] = string(condBytes)
		}

		// Set document ID
		doc["_id"] = docid
		documents = append(documents, doc)
	}

	err := cs.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, documents)
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		return err
	}

	return nil
}

func (cs *ConceptSyncer) insertDatasetDataForRelationTypes(ctx context.Context, relationTypes []*interfaces.RelationType) error {
	if len(relationTypes) == 0 {
		return nil
	}

	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, relationType := range relationTypes {
			arr := []string{relationType.RTName}
			arr = append(arr, relationType.Tags...)
			arr = append(arr, relationType.Comment, relationType.BKNRawContent)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}

		if len(vectors) != len(relationTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(relationTypes), len(vectors))
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(relationTypes), len(vectors))
		}

		for i, relationType := range relationTypes {
			relationType.Vector = vectors[i].Vector
		}
	}

	documents := make([]map[string]any, 0, len(relationTypes))
	for _, relationType := range relationTypes {
		docid := interfaces.GenerateConceptDocuemtnID(relationType.KNID, interfaces.MODULE_TYPE_RELATION_TYPE,
			relationType.RTID, relationType.Branch)
		relationType.ModuleType = interfaces.MODULE_TYPE_RELATION_TYPE

		// Convert to map for dataset
		docBytes, err := sonic.Marshal(relationType)
		if err != nil {
			logger.Errorf("Failed to marshal RelationType: %s", err.Error())
			return err
		}

		var doc map[string]any
		if err := sonic.Unmarshal(docBytes, &doc); err != nil {
			logger.Errorf("Failed to unmarshal RelationType: %s", err.Error())
			return err
		}

		// Set document ID
		doc["_id"] = docid
		documents = append(documents, doc)
	}

	err := cs.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, documents)
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		return err
	}

	return nil
}

func (cs *ConceptSyncer) insertDatasetDataForConceptGroups(ctx context.Context, conceptGroups []*interfaces.ConceptGroup) error {
	if len(conceptGroups) == 0 {
		return nil
	}

	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, conceptGroup := range conceptGroups {
			arr := []string{conceptGroup.CGName}
			arr = append(arr, conceptGroup.Tags...)
			arr = append(arr, conceptGroup.Comment, conceptGroup.BKNRawContent)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}

		if len(vectors) != len(conceptGroups) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(conceptGroups), len(vectors))
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(conceptGroups), len(vectors))
		}

		for i, conceptGroup := range conceptGroups {
			conceptGroup.Vector = vectors[i].Vector
		}
	}

	documents := make([]map[string]any, 0, len(conceptGroups))
	for _, conceptGroup := range conceptGroups {
		docid := interfaces.GenerateConceptDocuemtnID(conceptGroup.KNID, interfaces.MODULE_TYPE_CONCEPT_GROUP,
			conceptGroup.CGID, conceptGroup.Branch)
		conceptGroup.ModuleType = interfaces.MODULE_TYPE_CONCEPT_GROUP

		// Convert to map for dataset
		docBytes, err := sonic.Marshal(conceptGroup)
		if err != nil {
			logger.Errorf("Failed to marshal ConceptGroup: %s", err.Error())
			return err
		}

		var doc map[string]any
		if err := sonic.Unmarshal(docBytes, &doc); err != nil {
			logger.Errorf("Failed to unmarshal ConceptGroup: %s", err.Error())
			return err
		}

		// Set document ID
		doc["_id"] = docid
		documents = append(documents, doc)
	}

	err := cs.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, documents)
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		return err
	}

	return nil
}

func (cs *ConceptSyncer) insertDatasetDataForRiskTypes(ctx context.Context, riskTypes []*interfaces.RiskType) error {
	if len(riskTypes) == 0 {
		return nil
	}

	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, riskType := range riskTypes {
			arr := []string{riskType.RTName}
			arr = append(arr, riskType.Tags...)
			arr = append(arr, riskType.Comment, riskType.BKNRawContent)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}

		if len(vectors) != len(riskTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(riskTypes), len(vectors))
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(riskTypes), len(vectors))
		}

		for i, riskType := range riskTypes {
			riskType.Vector = vectors[i].Vector
		}
	}

	documents := make([]map[string]any, 0, len(riskTypes))
	for _, riskType := range riskTypes {
		docid := interfaces.GenerateConceptDocuemtnID(riskType.KNID, interfaces.MODULE_TYPE_RISK_TYPE,
			riskType.RTID, riskType.Branch)
		riskType.ModuleType = interfaces.MODULE_TYPE_RISK_TYPE

		// Convert to map for dataset
		docBytes, err := sonic.Marshal(riskType)
		if err != nil {
			logger.Errorf("Failed to marshal RiskType: %s", err.Error())
			return err
		}

		var doc map[string]any
		if err := sonic.Unmarshal(docBytes, &doc); err != nil {
			logger.Errorf("Failed to unmarshal RiskType: %s", err.Error())
			return err
		}

		// Set document ID
		doc["_id"] = docid
		documents = append(documents, doc)
	}

	err := cs.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, documents)
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		return err
	}

	return nil
}

func (cs *ConceptSyncer) insertDatasetDataForMetrics(ctx context.Context, metrics []*interfaces.MetricDefinition) error {
	if len(metrics) == 0 {
		return nil
	}

	if cs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := make([]string, 0, len(metrics))
		for _, m := range metrics {
			arr := []string{m.Name}
			arr = append(arr, m.Tags...)
			arr = append(arr, m.Comment, m.BKNRawContent)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := cs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			return err
		}
		vectors, err := cs.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			return err
		}
		if len(vectors) != len(metrics) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(metrics), len(vectors))
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(metrics), len(vectors))
		}
		for i := range metrics {
			metrics[i].Vector = vectors[i].Vector
		}
	}

	documents := make([]map[string]any, 0, len(metrics))
	for _, def := range metrics {
		docid := interfaces.GenerateConceptDocuemtnID(def.KnID, interfaces.MODULE_TYPE_METRIC, def.ID, def.Branch)
		def.ModuleType = interfaces.MODULE_TYPE_METRIC

		docBytes, err := sonic.Marshal(def)
		if err != nil {
			logger.Errorf("Failed to marshal MetricDefinition: %s", err.Error())
			return err
		}
		var doc map[string]any
		if err := sonic.Unmarshal(docBytes, &doc); err != nil {
			logger.Errorf("Failed to unmarshal MetricDefinition: %s", err.Error())
			return err
		}
		doc["_id"] = docid
		documents = append(documents, doc)
	}

	if err := cs.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, documents); err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		return err
	}
	return nil
}

func (cs *ConceptSyncer) getAllKNsFromDataset(ctx context.Context) (map[string]*interfaces.KN, error) {
	filterCondition := map[string]any{
		"field":      "module_type",
		"operation":  "==",
		"value":      interfaces.MODULE_TYPE_KN,
		"value_from": "const",
	}

	params := &interfaces.ResourceDataQueryParams{
		FilterCondition: filterCondition,
		Offset:          0,
		Limit:           10000,
		NeedTotal:       false,
	}
	response, err := cs.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
	if err != nil {
		return map[string]*interfaces.KN{}, err
	}

	kns := map[string]*interfaces.KN{}
	for _, entry := range response.Entries {
		kn := interfaces.KN{}
		err := mapstructure.Decode(entry, &kn)
		if err != nil {
			return map[string]*interfaces.KN{}, err
		}

		kns[kn.KNID] = &kn
	}

	return kns, nil
}

func (cs *ConceptSyncer) getAllObjectTypesFromDatasetByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.ObjectType, error) {

	filterCondition := map[string]any{
		"operation": "and",
		"sub_conditions": []map[string]any{
			{
				"field":      "kn_id",
				"operation":  "==",
				"value":      knID,
				"value_from": "const",
			},
			{
				"field":      "branch",
				"operation":  "==",
				"value":      branch,
				"value_from": "const",
			},
			{
				"field":      "module_type",
				"operation":  "==",
				"value":      interfaces.MODULE_TYPE_OBJECT_TYPE,
				"value_from": "const",
			},
		},
	}

	params := &interfaces.ResourceDataQueryParams{
		FilterCondition: filterCondition,
		Offset:          0,
		Limit:           10000,
		NeedTotal:       false,
	}
	response, err := cs.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
	if err != nil {
		return map[string]*interfaces.ObjectType{}, err
	}

	objectTypes := map[string]*interfaces.ObjectType{}
	for _, entry := range response.Entries {
		// Deserialize logic_properties[].parameters from JSON string
		if logicProps, ok := entry["logic_properties"].([]any); ok {
			for _, lp := range logicProps {
				if lpMap, ok := lp.(map[string]any); ok {
					if paramsStr, exists := lpMap["parameters"]; exists {
						if paramsStrStr, ok := paramsStr.(string); ok {
							var params []interfaces.Parameter
							if err := sonic.Unmarshal([]byte(paramsStrStr), &params); err != nil {
								logger.Errorf("Failed to unmarshal logic_properties parameters: %s", err.Error())
								return map[string]*interfaces.ObjectType{}, err
							}
							lpMap["parameters"] = params
						}
					}
				}
			}
		}

		objectType := interfaces.ObjectType{}
		err := mapstructure.Decode(entry, &objectType)
		if err != nil {
			return map[string]*interfaces.ObjectType{}, err
		}

		objectTypes[objectType.OTID] = &objectType
	}

	return objectTypes, nil
}

func (cs *ConceptSyncer) getAllRelationTypesFromDatasetByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.RelationType, error) {

	filterCondition := map[string]any{
		"operation": "and",
		"sub_conditions": []map[string]any{
			{
				"field":      "kn_id",
				"operation":  "==",
				"value":      knID,
				"value_from": "const",
			},
			{
				"field":      "branch",
				"operation":  "==",
				"value":      branch,
				"value_from": "const",
			},
			{
				"field":      "module_type",
				"operation":  "==",
				"value":      interfaces.MODULE_TYPE_RELATION_TYPE,
				"value_from": "const",
			},
		},
	}

	params := &interfaces.ResourceDataQueryParams{
		FilterCondition: filterCondition,
		Offset:          0,
		Limit:           10000,
		NeedTotal:       false,
	}
	response, err := cs.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
	if err != nil {
		return map[string]*interfaces.RelationType{}, err
	}

	relationTypes := map[string]*interfaces.RelationType{}
	for _, entry := range response.Entries {
		relationType := interfaces.RelationType{}
		err = mapstructure.Decode(entry, &relationType)
		if err != nil {
			return map[string]*interfaces.RelationType{}, err
		}

		relationTypes[relationType.RTID] = &relationType
	}

	return relationTypes, nil
}

func (cs *ConceptSyncer) getAllActionTypesFromDatasetByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.ActionType, error) {

	filterCondition := map[string]any{
		"operation": "and",
		"sub_conditions": []map[string]any{
			{
				"field":      "kn_id",
				"operation":  "==",
				"value":      knID,
				"value_from": "const",
			},
			{
				"field":      "branch",
				"operation":  "==",
				"value":      branch,
				"value_from": "const",
			},
			{
				"field":      "module_type",
				"operation":  "==",
				"value":      interfaces.MODULE_TYPE_ACTION_TYPE,
				"value_from": "const",
			},
		},
	}

	params := &interfaces.ResourceDataQueryParams{
		FilterCondition: filterCondition,
		Offset:          0,
		Limit:           10000,
		NeedTotal:       false,
	}
	response, err := cs.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
	if err != nil {
		return map[string]*interfaces.ActionType{}, err
	}

	actionTypes := map[string]*interfaces.ActionType{}
	for _, entry := range response.Entries {
		// Deserialize condition from JSON string
		if condStr, exists := entry["condition"]; exists {
			if condStrStr, ok := condStr.(string); ok && condStrStr != "" {
				var condCfg interfaces.ActionCondCfg
				if err := sonic.Unmarshal([]byte(condStrStr), &condCfg); err != nil {
					logger.Errorf("Failed to unmarshal action_type condition: %s", err.Error())
					return map[string]*interfaces.ActionType{}, err
				}
				entry["condition"] = &condCfg
			} else if condStr == nil {
				entry["condition"] = nil
			}
		}

		// Deserialize parameters from JSON string
		if paramsStr, exists := entry["parameters"]; exists {
			if paramsStrStr, ok := paramsStr.(string); ok {
				var params []interfaces.Parameter
				if err := sonic.Unmarshal([]byte(paramsStrStr), &params); err != nil {
					logger.Errorf("Failed to unmarshal action_type parameters: %s", err.Error())
					return map[string]*interfaces.ActionType{}, err
				}
				entry["parameters"] = params
			}
		}

		actionType := interfaces.ActionType{}
		err = mapstructure.Decode(entry, &actionType)
		if err != nil {
			return map[string]*interfaces.ActionType{}, err
		}

		actionTypes[actionType.ATID] = &actionType
	}

	return actionTypes, nil
}

func (cs *ConceptSyncer) getAllRiskTypesFromDatasetByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.RiskType, error) {

	filterCondition := map[string]any{
		"operation": "and",
		"sub_conditions": []map[string]any{
			{
				"field":      "kn_id",
				"operation":  "==",
				"value":      knID,
				"value_from": "const",
			},
			{
				"field":      "branch",
				"operation":  "==",
				"value":      branch,
				"value_from": "const",
			},
			{
				"field":      "module_type",
				"operation":  "==",
				"value":      interfaces.MODULE_TYPE_RISK_TYPE,
				"value_from": "const",
			},
		},
	}

	params := &interfaces.ResourceDataQueryParams{
		FilterCondition: filterCondition,
		Offset:          0,
		Limit:           10000,
		NeedTotal:       false,
	}
	response, err := cs.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
	if err != nil {
		return map[string]*interfaces.RiskType{}, err
	}

	riskTypes := map[string]*interfaces.RiskType{}
	for _, entry := range response.Entries {
		riskType := interfaces.RiskType{}
		err := mapstructure.Decode(entry, &riskType)
		if err != nil {
			return map[string]*interfaces.RiskType{}, err
		}

		riskTypes[riskType.RTID] = &riskType
	}

	return riskTypes, nil
}

func (cs *ConceptSyncer) getAllConceptGroupsFromDatasetByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.ConceptGroup, error) {

	filterCondition := map[string]any{
		"operation": "and",
		"sub_conditions": []map[string]any{
			{
				"field":      "kn_id",
				"operation":  "==",
				"value":      knID,
				"value_from": "const",
			},
			{
				"field":      "branch",
				"operation":  "==",
				"value":      branch,
				"value_from": "const",
			},
			{
				"field":      "module_type",
				"operation":  "==",
				"value":      interfaces.MODULE_TYPE_CONCEPT_GROUP,
				"value_from": "const",
			},
		},
	}

	params := &interfaces.ResourceDataQueryParams{
		FilterCondition: filterCondition,
		Offset:          0,
		Limit:           10000,
		NeedTotal:       false,
	}
	response, err := cs.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
	if err != nil {
		return map[string]*interfaces.ConceptGroup{}, err
	}

	conceptGroups := map[string]*interfaces.ConceptGroup{}
	for _, entry := range response.Entries {
		conceptGroup := interfaces.ConceptGroup{}
		err := mapstructure.Decode(entry, &conceptGroup)
		if err != nil {
			return map[string]*interfaces.ConceptGroup{}, err
		}

		conceptGroups[conceptGroup.CGID] = &conceptGroup
	}

	return conceptGroups, nil
}

func (cs *ConceptSyncer) getAllMetricsFromDatasetByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.MetricDefinition, error) {

	filterCondition := map[string]any{
		"operation": "and",
		"sub_conditions": []map[string]any{
			{
				"field":      "kn_id",
				"operation":  "==",
				"value":      knID,
				"value_from": "const",
			},
			{
				"field":      "branch",
				"operation":  "==",
				"value":      branch,
				"value_from": "const",
			},
			{
				"field":      "module_type",
				"operation":  "==",
				"value":      interfaces.MODULE_TYPE_METRIC,
				"value_from": "const",
			},
		},
	}

	params := &interfaces.ResourceDataQueryParams{
		FilterCondition: filterCondition,
		Offset:          0,
		Limit:           10000,
		NeedTotal:       false,
	}
	response, err := cs.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
	if err != nil {
		return map[string]*interfaces.MetricDefinition{}, err
	}

	metrics := map[string]*interfaces.MetricDefinition{}
	for _, entry := range response.Entries {
		md := interfaces.MetricDefinition{}
		if err := mapstructure.WeakDecode(entry, &md); err != nil {
			return map[string]*interfaces.MetricDefinition{}, err
		}
		mcopy := md
		metrics[mcopy.ID] = &mcopy
	}

	return metrics, nil
}
