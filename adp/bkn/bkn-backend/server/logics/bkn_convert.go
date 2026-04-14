// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"bkn-backend/interfaces"
)

func ToADPNetWork(bknNetwork *bknsdk.BknNetwork) *interfaces.KN {
	return &interfaces.KN{
		KNID:   bknNetwork.ID,
		KNName: bknNetwork.Name,
		CommonInfo: interfaces.CommonInfo{
			Tags:          bknNetwork.Tags,
			Comment:       bknNetwork.Description,
			BKNRawContent: bknNetwork.RawContent,
		},

		SkillContent: bknNetwork.SkillContent,

		Branch:         bknNetwork.Branch,
		BusinessDomain: bknNetwork.BusinessDomain,
	}
}

func ToBKNNetWork(kn *interfaces.KN) *bknsdk.BknNetwork {
	return &bknsdk.BknNetwork{
		BknNetworkFrontmatter: bknsdk.BknNetworkFrontmatter{
			Type:        interfaces.MODULE_TYPE_KN,
			ID:          kn.KNID,
			Name:        kn.KNName,
			Tags:        kn.Tags,
			Summary:     bknsdk.ExtractSummary(kn.Comment),
			Description: kn.Comment,

			Branch:         kn.Branch,
			BusinessDomain: kn.BusinessDomain,
		},
		RawContent:   kn.BKNRawContent,
		SkillContent: kn.SkillContent,
	}
}

// ToADPObjectType 将 BKN ObjectType 转换为 ADP ObjectType
func ToADPObjectType(knID string, branch string, bknObj *bknsdk.BknObjectType) *interfaces.ObjectType {
	adpObj := &interfaces.ObjectType{
		ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
			OTID:           bknObj.ID,
			OTName:         bknObj.Name,
			PrimaryKeys:    bknObj.PrimaryKeys,
			DisplayKey:     bknObj.DisplayKey,
			IncrementalKey: bknObj.IncrementalKey,
		},
		CommonInfo: interfaces.CommonInfo{
			Tags:    bknObj.Tags,
			Comment: bknObj.Description,
		},
		KNID:   knID,
		Branch: branch,
	}

	// 转换 DataSource
	if bknObj.DataSource != nil {
		adpObj.DataSource = &interfaces.ResourceInfo{
			ID:   bknObj.DataSource.ID,
			Type: bknObj.DataSource.Type,
			Name: bknObj.DataSource.Name,
		}
	}

	// 转换 DataProperties
	for _, dp := range bknObj.DataProperties {
		adpDP := &interfaces.DataProperty{
			Name:        dp.Name,
			DisplayName: dp.DisplayName,
			Type:        dp.Type,
			Comment:     dp.Description,
		}

		if dp.MappedField != "" {
			adpDP.MappedField = &interfaces.Field{
				Name: dp.MappedField,
			}
		}

		adpObj.DataProperties = append(adpObj.DataProperties, adpDP)
	}

	// 转换 LogicProperties
	for _, lp := range bknObj.LogicProperties {
		adpLP := &interfaces.LogicProperty{
			Name:        lp.Name,
			DisplayName: lp.DisplayName,
			Type:        lp.Type,
			Comment:     lp.Description,
		}

		if lp.DataSource != nil {
			adpLP.DataSource = &interfaces.ResourceInfo{
				ID:   lp.DataSource.ID,
				Type: lp.DataSource.Type,
				Name: lp.DataSource.Name,
			}
		}

		for _, param := range lp.Parameters {
			adpLP.Parameters = append(adpLP.Parameters, interfaces.Parameter{
				Name:        param.Name,
				Type:        param.Type,
				Source:      param.Source,
				Operation:   param.Operation,
				ValueFrom:   param.ValueFrom,
				Value:       param.Value,
				IfSystemGen: &param.IfSystemGen,
				Comment:     &param.Description,
			})
		}

		for _, dim := range lp.AnalysisDims {
			adpLP.AnalysisDims = append(adpLP.AnalysisDims, interfaces.Field{
				Name:        dim.Name,
				Type:        dim.Type,
				DisplayName: dim.DisplayName,
			})
		}

		adpObj.LogicProperties = append(adpObj.LogicProperties, adpLP)
	}

	return adpObj
}

// ToBKNObjectType 将 ADP ObjectType 转换为 BKN ObjectType
func ToBKNObjectType(adpObj *interfaces.ObjectType) *bknsdk.BknObjectType {
	bknObj := &bknsdk.BknObjectType{
		BknObjectTypeFrontmatter: bknsdk.BknObjectTypeFrontmatter{
			Type:        interfaces.MODULE_TYPE_OBJECT_TYPE,
			ID:          adpObj.OTID,
			Name:        adpObj.OTName,
			Tags:        adpObj.Tags,
			Summary:     bknsdk.ExtractSummary(adpObj.Comment),
			Description: adpObj.Comment,
		},
		PrimaryKeys:    adpObj.PrimaryKeys,
		DisplayKey:     adpObj.DisplayKey,
		IncrementalKey: adpObj.IncrementalKey,
	}

	// 转换 DataSource
	if adpObj.DataSource != nil {
		bknObj.DataSource = &bknsdk.ResourceInfo{
			ID:   adpObj.DataSource.ID,
			Type: adpObj.DataSource.Type,
			Name: adpObj.DataSource.Name,
		}
	}

	// 转换 DataProperties
	for _, adpDP := range adpObj.DataProperties {
		dp := &bknsdk.DataProperty{
			Name:        adpDP.Name,
			DisplayName: adpDP.DisplayName,
			Type:        adpDP.Type,
			Description: adpDP.Comment,
		}

		if adpDP.MappedField != nil {
			dp.MappedField = adpDP.MappedField.Name
		}

		bknObj.DataProperties = append(bknObj.DataProperties, dp)
	}

	// 转换 LogicProperties
	for _, adpLP := range adpObj.LogicProperties {
		lp := &bknsdk.LogicProperty{
			Name:        adpLP.Name,
			DisplayName: adpLP.DisplayName,
			Type:        adpLP.Type,
			Description: adpLP.Comment,
		}

		if adpLP.DataSource != nil {
			lp.DataSource = &bknsdk.ResourceInfo{
				ID:   adpLP.DataSource.ID,
				Type: adpLP.DataSource.Type,
				Name: adpLP.DataSource.Name,
			}
		}

		for _, adpParam := range adpLP.Parameters {
			param := bknsdk.Parameter{
				Name:        adpParam.Name,
				Type:        adpParam.Type,
				Source:      adpParam.Source,
				Operation:   adpParam.Operation,
				ValueFrom:   adpParam.ValueFrom,
				Value:       adpParam.Value,
				Description: "",
			}
			if adpParam.Comment != nil {
				param.Description = *adpParam.Comment
			}
			if adpParam.IfSystemGen != nil {
				param.IfSystemGen = *adpParam.IfSystemGen
			}
			lp.Parameters = append(lp.Parameters, param)
		}

		for _, adpDim := range adpLP.AnalysisDims {
			lp.AnalysisDims = append(lp.AnalysisDims, bknsdk.Field{
				Name:        adpDim.Name,
				Type:        adpDim.Type,
				DisplayName: adpDim.DisplayName,
			})
		}

		bknObj.LogicProperties = append(bknObj.LogicProperties, lp)
	}

	return bknObj
}

// ToADPRelationType 将 BKN RelationType 转换为 ADP RelationType
func ToADPRelationType(knID string, branch string, bknRel *bknsdk.BknRelationType) *interfaces.RelationType {
	relType := &interfaces.RelationType{
		RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{
			RTID:   bknRel.ID,
			RTName: bknRel.Name,
		},
		CommonInfo: interfaces.CommonInfo{
			Tags:    bknRel.Tags,
			Comment: bknRel.Description,
		},
		KNID:   knID,
		Branch: branch,
	}

	// 转换 Endpoint
	relType.SourceObjectTypeID = bknRel.Endpoint.Source
	relType.TargetObjectTypeID = bknRel.Endpoint.Target
	relType.Type = bknRel.Endpoint.Type

	// 转换 MappingRules
	if bknRel.MappingRules != nil {
		switch rules := bknRel.MappingRules.(type) {
		case bknsdk.DirectMappingRule:
			var mappings []interfaces.Mapping
			for _, rule := range rules {
				mappings = append(mappings, interfaces.Mapping{
					SourceProp: interfaces.SimpleProperty{
						Name: rule.SourceProperty,
					},
					TargetProp: interfaces.SimpleProperty{
						Name: rule.TargetProperty,
					},
				})
			}
			relType.MappingRules = mappings

		case *bknsdk.InDirectMappingRule:
			indirect := &interfaces.InDirectMapping{
				BackingDataSource: &interfaces.ResourceInfo{
					Type: rules.BackingDataSource.Type,
					ID:   rules.BackingDataSource.ID,
					Name: rules.BackingDataSource.Name,
				},
			}
			for _, rule := range rules.SourceMappingRules {
				indirect.SourceMappingRules = append(indirect.SourceMappingRules, interfaces.Mapping{
					SourceProp: interfaces.SimpleProperty{Name: rule.SourceProperty},
					TargetProp: interfaces.SimpleProperty{Name: rule.TargetProperty},
				})
			}
			for _, rule := range rules.TargetMappingRules {
				indirect.TargetMappingRules = append(indirect.TargetMappingRules, interfaces.Mapping{
					SourceProp: interfaces.SimpleProperty{Name: rule.SourceProperty},
					TargetProp: interfaces.SimpleProperty{Name: rule.TargetProperty},
				})
			}
			relType.MappingRules = indirect

		default:
			logger.Errorf("Unknown mappingRules type: %T", rules)
		}
	}

	return relType
}

// ToBKNRelationType 将 ADP RelationType 转换为 BKN RelationType
func ToBKNRelationType(adpRel *interfaces.RelationType) *bknsdk.BknRelationType {
	bknRel := &bknsdk.BknRelationType{
		BknRelationTypeFrontmatter: bknsdk.BknRelationTypeFrontmatter{
			Type:        interfaces.MODULE_TYPE_RELATION_TYPE,
			ID:          adpRel.RTID,
			Name:        adpRel.RTName,
			Tags:        adpRel.Tags,
			Summary:     bknsdk.ExtractSummary(adpRel.Comment),
			Description: adpRel.Comment,
		},
		Endpoint: bknsdk.Endpoint{
			Source: adpRel.SourceObjectTypeID,
			Target: adpRel.TargetObjectTypeID,
			Type:   adpRel.Type,
		},
	}

	// 转换 MappingRules
	if adpRel.MappingRules != nil {
		switch rules := adpRel.MappingRules.(type) {
		case []interfaces.Mapping:
			var mappingRules bknsdk.DirectMappingRule
			for _, rule := range rules {
				mappingRules = append(mappingRules, bknsdk.MappingRule{
					SourceProperty: rule.SourceProp.Name,
					TargetProperty: rule.TargetProp.Name,
				})
			}
			bknRel.MappingRules = mappingRules

		case *interfaces.InDirectMapping:
			indirectRules := &bknsdk.InDirectMappingRule{
				BackingDataSource: &bknsdk.ResourceInfo{
					Type: rules.BackingDataSource.Type,
					ID:   rules.BackingDataSource.ID,
					Name: rules.BackingDataSource.Name,
				},
			}
			for _, rule := range rules.SourceMappingRules {
				indirectRules.SourceMappingRules = append(indirectRules.SourceMappingRules, bknsdk.MappingRule{
					SourceProperty: rule.SourceProp.Name,
					TargetProperty: rule.TargetProp.Name,
				})
			}
			for _, rule := range rules.TargetMappingRules {
				indirectRules.TargetMappingRules = append(indirectRules.TargetMappingRules, bknsdk.MappingRule{
					SourceProperty: rule.SourceProp.Name,
					TargetProperty: rule.TargetProp.Name,
				})
			}
			bknRel.MappingRules = indirectRules

		default:
			logger.Errorf("Unknown mappingRules type: %T", rules)
		}
	}

	return bknRel
}

// ToADPActionType 将 BKN ActionType 转换为 ADP ActionType
func ToADPActionType(knID string, branch string, bknAction *bknsdk.BknActionType) *interfaces.ActionType {
	adpAction := &interfaces.ActionType{
		ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{
			ATID:         bknAction.ID,
			ATName:       bknAction.Name,
			ActionType:   bknAction.ActionType,
			ObjectTypeID: bknAction.BoundObject,
		},
		CommonInfo: interfaces.CommonInfo{
			Tags:    bknAction.Tags,
			Comment: bknAction.Description,
		},
		KNID:   knID,
		Branch: branch,
	}

	// 转换 Affect
	if bknAction.AffectObject != nil {
		adpAction.Affect = &interfaces.ActionAffect{
			ObjectTypeID: bknAction.AffectObject.ObjectType,
			Comment:      bknAction.AffectObject.Description,
		}
	}

	// 转换 Condition
	if bknAction.TriggerCondition != nil {
		adpAction.Condition = toADPCondCfg(bknAction.TriggerCondition)
	}

	// 转换 ActionSource
	if bknAction.ActionSource != nil {
		adpAction.ActionSource = interfaces.ActionSource{
			Type:     bknAction.ActionSource.Type,
			BoxID:    bknAction.ActionSource.BoxID,
			ToolID:   bknAction.ActionSource.ToolID,
			McpID:    bknAction.ActionSource.McpID,
			ToolName: bknAction.ActionSource.ToolName,
		}
	}

	// 转换 Parameters
	for _, param := range bknAction.Parameters {
		adpAction.Parameters = append(adpAction.Parameters, interfaces.Parameter{
			Name:        param.Name,
			Type:        param.Type,
			Source:      param.Source,
			Operation:   param.Operation,
			ValueFrom:   param.ValueFrom,
			Value:       param.Value,
			IfSystemGen: &param.IfSystemGen,
			Comment:     &param.Description,
		})
	}

	// 转换 Schedule
	if bknAction.Schedule != nil {
		adpAction.Schedule = interfaces.Schedule{
			Type:       bknAction.Schedule.Type,
			Expression: bknAction.Schedule.Expression,
		}
	}

	return adpAction
}

// ToBKNActionType 将 ADP ActionType 转换为 BKN ActionType
func ToBKNActionType(adpAction *interfaces.ActionType) *bknsdk.BknActionType {
	bknAction := &bknsdk.BknActionType{
		BknActionTypeFrontmatter: bknsdk.BknActionTypeFrontmatter{
			Type:        interfaces.MODULE_TYPE_ACTION_TYPE,
			ID:          adpAction.ATID,
			Name:        adpAction.ATName,
			Tags:        adpAction.Tags,
			Summary:     bknsdk.ExtractSummary(adpAction.Comment),
			Description: adpAction.Comment,
			ActionType:  adpAction.ActionType,
		},
		BoundObject: adpAction.ObjectTypeID,
	}

	if adpAction.Affect != nil {
		bknAction.AffectObject = &bknsdk.ActionAffect{
			ObjectType:  adpAction.Affect.ObjectTypeID,
			Description: adpAction.Affect.Comment,
		}
	}
	// 转换 Condition
	if adpAction.Condition != nil {
		bknAction.TriggerCondition = toBKNCondCfg(adpAction.Condition)
	}

	// 转换 ActionSource
	if adpAction.ActionSource.Type != "" {
		bknAction.ActionSource = &bknsdk.ActionSource{
			Type:     adpAction.ActionSource.Type,
			BoxID:    adpAction.ActionSource.BoxID,
			ToolID:   adpAction.ActionSource.ToolID,
			McpID:    adpAction.ActionSource.McpID,
			ToolName: adpAction.ActionSource.ToolName,
		}
	}

	// 转换 Parameters
	for _, adpParam := range adpAction.Parameters {
		param := bknsdk.Parameter{
			Name:      adpParam.Name,
			Type:      adpParam.Type,
			Source:    adpParam.Source,
			Operation: adpParam.Operation,
			ValueFrom: adpParam.ValueFrom,
			Value:     adpParam.Value,
		}
		if adpParam.Comment != nil {
			param.Description = *adpParam.Comment
		}
		if adpParam.IfSystemGen != nil {
			param.IfSystemGen = *adpParam.IfSystemGen
		}
		bknAction.Parameters = append(bknAction.Parameters, param)
	}

	// 转换 Schedule
	if adpAction.Schedule.Type != "" {
		bknAction.Schedule = &bknsdk.Schedule{
			Type:       adpAction.Schedule.Type,
			Expression: adpAction.Schedule.Expression,
		}
	}

	return bknAction
}

// ToADPRiskType 将 BKN RiskType 转换为 ADP RiskType
func ToADPRiskType(knID string, branch string, bknRisk *bknsdk.BknRiskType) *interfaces.RiskType {
	adpRisk := &interfaces.RiskType{
		RTID:   bknRisk.ID,
		RTName: bknRisk.Name,
		CommonInfo: interfaces.CommonInfo{
			Tags:          bknRisk.Tags,
			Comment:       bknRisk.Description,
			BKNRawContent: bknRisk.RawContent,
		},
		KNID:   knID,
		Branch: branch,
	}

	return adpRisk
}

// ToBKNRiskType 将 ADP RiskType 转换为 BKN RiskType
func ToBKNRiskType(adpRisk *interfaces.RiskType) *bknsdk.BknRiskType {
	bknRisk := &bknsdk.BknRiskType{
		BknRiskTypeFrontmatter: bknsdk.BknRiskTypeFrontmatter{
			Type:        interfaces.MODULE_TYPE_RISK_TYPE,
			ID:          adpRisk.RTID,
			Name:        adpRisk.RTName,
			Tags:        adpRisk.Tags,
			Summary:     bknsdk.ExtractSummary(adpRisk.Comment),
			Description: adpRisk.Comment,
		},
	}

	return bknRisk
}

// ToADPConceptGroup 将 BKN ConceptGroup 转换为 ADP ConceptGroup
func ToADPConceptGroup(knID string, branch string, bknCG *bknsdk.BknConceptGroup) *interfaces.ConceptGroup {
	adpCG := &interfaces.ConceptGroup{
		CGID:   bknCG.ID,
		CGName: bknCG.Name,
		CommonInfo: interfaces.CommonInfo{
			Tags:    bknCG.Tags,
			Comment: bknCG.Description,
		},
		KNID:          knID,
		Branch:        branch,
		ObjectTypeIDs: bknCG.ObjectTypes,
	}

	return adpCG
}

// ToBKNConceptGroup 将 ADP ConceptGroup 转换为 BKN ConceptGroup
func ToBKNConceptGroup(adpCG *interfaces.ConceptGroup) *bknsdk.BknConceptGroup {
	bknCG := &bknsdk.BknConceptGroup{
		BknConceptGroupFrontmatter: bknsdk.BknConceptGroupFrontmatter{
			Type:        interfaces.MODULE_TYPE_CONCEPT_GROUP,
			ID:          adpCG.CGID,
			Name:        adpCG.CGName,
			Tags:        adpCG.Tags,
			Summary:     bknsdk.ExtractSummary(adpCG.Comment),
			Description: adpCG.Comment,
		},
		ObjectTypes: adpCG.ObjectTypeIDs,
	}

	return bknCG
}

// toADPCondCfg 将 BKN CondCfg 转换为 ADP CondCfg
func toADPCondCfg(bknCond *bknsdk.CondCfg) *interfaces.CondCfg {
	if bknCond == nil {
		return nil
	}

	adpCond := &interfaces.CondCfg{
		ObjectTypeID: bknCond.ObjectTypeID,
		Field:        bknCond.Field,
		Operation:    bknCond.Operation,
		ValueOptCfg: interfaces.ValueOptCfg{
			ValueFrom: bknCond.ValueFrom,
			Value:     bknCond.Value,
		},
	}

	for _, subCond := range bknCond.SubConds {
		adpCond.SubConds = append(adpCond.SubConds, toADPCondCfg(subCond))
	}

	return adpCond
}

// toBKNCondCfg 将 ADP CondCfg 转换为 BKN CondCfg
func toBKNCondCfg(adpCond *interfaces.CondCfg) *bknsdk.CondCfg {
	if adpCond == nil {
		return nil
	}

	bknCond := &bknsdk.CondCfg{
		ObjectTypeID: adpCond.ObjectTypeID,
		Field:        adpCond.Field,
		Operation:    adpCond.Operation,
		ValueFrom:    adpCond.ValueFrom,
		Value:        adpCond.Value,
	}

	for _, subCond := range adpCond.SubConds {
		bknCond.SubConds = append(bknCond.SubConds, toBKNCondCfg(subCond))
	}

	return bknCond
}
