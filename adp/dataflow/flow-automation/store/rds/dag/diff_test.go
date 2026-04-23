package dagmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffDagVars(t *testing.T) {
	t.Run("all insert", func(t *testing.T) {
		existing := []existingDagVar{}
		newVars := []*DagVarModel{
			{VarName: "var1", DefaultValue: "val1", VarType: "string"},
			{VarName: "var2", DefaultValue: "val2", VarType: "int"},
		}
		result := diffDagVars(existing, newVars)
		assert.Len(t, result.toInsert, 2)
		assert.Len(t, result.toUpdate, 0)
		assert.Len(t, result.toDelete, 0)
	})

	t.Run("all delete", func(t *testing.T) {
		existing := []existingDagVar{
			{VarName: "var1", DefaultValue: "val1"},
			{VarName: "var2", DefaultValue: "val2"},
		}
		newVars := []*DagVarModel{}
		result := diffDagVars(existing, newVars)
		assert.Len(t, result.toInsert, 0)
		assert.Len(t, result.toUpdate, 0)
		assert.Len(t, result.toDelete, 2)
	})

	t.Run("update needed", func(t *testing.T) {
		existing := []existingDagVar{
			{VarName: "var1", DefaultValue: "old_val", VarType: "string"},
		}
		newVars := []*DagVarModel{
			{VarName: "var1", DefaultValue: "new_val", VarType: "string"},
		}
		result := diffDagVars(existing, newVars)
		assert.Len(t, result.toInsert, 0)
		assert.Len(t, result.toUpdate, 1)
		assert.Len(t, result.toDelete, 0)
	})

	t.Run("no change", func(t *testing.T) {
		existing := []existingDagVar{
			{VarName: "var1", DefaultValue: "val1", VarType: "string", Description: "desc"},
		}
		newVars := []*DagVarModel{
			{VarName: "var1", DefaultValue: "val1", VarType: "string", Description: "desc"},
		}
		result := diffDagVars(existing, newVars)
		assert.Len(t, result.toInsert, 0)
		assert.Len(t, result.toUpdate, 0)
		assert.Len(t, result.toDelete, 0)
	})

	t.Run("mixed operations", func(t *testing.T) {
		existing := []existingDagVar{
			{VarName: "var1", DefaultValue: "val1"},
			{VarName: "var2", DefaultValue: "old_val"},
		}
		newVars := []*DagVarModel{
			{VarName: "var2", DefaultValue: "new_val"},
			{VarName: "var3", DefaultValue: "val3"},
		}
		result := diffDagVars(existing, newVars)
		assert.Len(t, result.toInsert, 1)
		assert.Len(t, result.toUpdate, 1)
		assert.Len(t, result.toDelete, 1)
	})
}

func TestDiffDagSteps(t *testing.T) {
	t.Run("mixed operations", func(t *testing.T) {
		existing := []existingDagStep{
			{ID: 1, Operator: "op1", SourceID: "src1", HasDatasource: false},
			{ID: 2, Operator: "op2", SourceID: "src2", HasDatasource: false},
		}
		newSteps := []*DagStepModel{
			{Operator: "op2", SourceID: "src2", HasDatasource: true},
			{Operator: "op3", SourceID: "src3", HasDatasource: false},
		}
		result := diffDagSteps(existing, newSteps)
		assert.Len(t, result.toInsert, 1)
		assert.Len(t, result.toUpdate, 1)
		assert.Len(t, result.toDelete, 1)
	})
}

func TestDiffDagAccessors(t *testing.T) {
	t.Run("insert and delete", func(t *testing.T) {
		existing := []existingDagAccessor{
			{ID: 1, AccessorID: "acc1"},
			{ID: 2, AccessorID: "acc2"},
		}
		newAccessors := []*DagAccessorModel{
			{AccessorID: "acc2"},
			{AccessorID: "acc3"},
		}
		result := diffDagAccessors(existing, newAccessors)
		assert.Len(t, result.toInsert, 1)
		assert.Len(t, result.toDelete, 1)
	})
}
