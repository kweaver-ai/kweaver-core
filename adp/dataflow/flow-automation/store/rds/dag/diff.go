package dagmodel

// existingDagVar 现有变量（用于 diff）
type existingDagVar struct {
	VarName      string
	DefaultValue string
	VarType      string
	Description  string
}

// existingDagStep 现有步骤（用于 diff）
type existingDagStep struct {
	ID            uint64
	Operator      string
	SourceID      string
	HasDatasource bool
}

// existingDagAccessor 现有访问者（用于 diff）
type existingDagAccessor struct {
	ID         uint64
	AccessorID string
}

// dagVarsDiff diff 计算结果
type dagVarsDiff struct {
	toInsert []*DagVarModel
	toUpdate []*DagVarModel
	toDelete []string // var_name 列表
}

// dagStepsDiff diff 计算结果
type dagStepsDiff struct {
	toInsert []*DagStepModel
	toUpdate []*DagStepModel
	toDelete []uint64 // id 列表
}

// dagAccessorsDiff diff 计算结果
type dagAccessorsDiff struct {
	toInsert []*DagAccessorModel
	toDelete []uint64 // id 列表
}

// diffDagVars 计算变量差异
func diffDagVars(existing []existingDagVar, newVars []*DagVarModel) *dagVarsDiff {
	result := &dagVarsDiff{
		toInsert: make([]*DagVarModel, 0),
		toUpdate: make([]*DagVarModel, 0),
		toDelete: make([]string, 0),
	}

	existingMap := make(map[string]existingDagVar)
	for _, v := range existing {
		existingMap[v.VarName] = v
	}

	newVarNames := make(map[string]bool)
	for _, newVar := range newVars {
		newVarNames[newVar.VarName] = true
		if existing, ok := existingMap[newVar.VarName]; ok {
			if existing.DefaultValue != newVar.DefaultValue ||
				existing.VarType != newVar.VarType ||
				existing.Description != newVar.Description {
				result.toUpdate = append(result.toUpdate, newVar)
			}
		} else {
			result.toInsert = append(result.toInsert, newVar)
		}
	}

	for _, v := range existing {
		if !newVarNames[v.VarName] {
			result.toDelete = append(result.toDelete, v.VarName)
		}
	}

	return result
}

// stepKey 生成步骤的唯一标识
func stepKey(operator, sourceID string) string {
	return operator + ":" + sourceID
}

// diffDagSteps 计算步骤差异
// 注意：此函数会修改 newSteps 中需要 UPDATE 的元素的 ID 字段
func diffDagSteps(existing []existingDagStep, newSteps []*DagStepModel) *dagStepsDiff {
	result := &dagStepsDiff{
		toInsert: make([]*DagStepModel, 0),
		toUpdate: make([]*DagStepModel, 0),
		toDelete: make([]uint64, 0),
	}

	existingMap := make(map[string]existingDagStep)
	for _, s := range existing {
		key := stepKey(s.Operator, s.SourceID)
		existingMap[key] = s
	}

	newKeys := make(map[string]bool)
	for _, newStep := range newSteps {
		key := stepKey(newStep.Operator, newStep.SourceID)
		newKeys[key] = true

		if existing, ok := existingMap[key]; ok {
			if existing.HasDatasource != newStep.HasDatasource {
				newStep.ID = existing.ID
				result.toUpdate = append(result.toUpdate, newStep)
			}
		} else {
			result.toInsert = append(result.toInsert, newStep)
		}
	}

	for _, s := range existing {
		key := stepKey(s.Operator, s.SourceID)
		if !newKeys[key] {
			result.toDelete = append(result.toDelete, s.ID)
		}
	}

	return result
}

// diffDagAccessors 计算访问者差异
func diffDagAccessors(existing []existingDagAccessor, newAccessors []*DagAccessorModel) *dagAccessorsDiff {
	result := &dagAccessorsDiff{
		toInsert: make([]*DagAccessorModel, 0),
		toDelete: make([]uint64, 0),
	}

	existingMap := make(map[string]existingDagAccessor)
	for _, a := range existing {
		existingMap[a.AccessorID] = a
	}

	newAccessorIDs := make(map[string]bool)
	for _, newAcc := range newAccessors {
		newAccessorIDs[newAcc.AccessorID] = true
		if _, ok := existingMap[newAcc.AccessorID]; !ok {
			result.toInsert = append(result.toInsert, newAcc)
		}
	}

	for _, a := range existing {
		if !newAccessorIDs[a.AccessorID] {
			result.toDelete = append(result.toDelete, a.ID)
		}
	}

	return result
}
