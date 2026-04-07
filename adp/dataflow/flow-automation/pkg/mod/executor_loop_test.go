package mod

import (
	"testing"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/stretchr/testify/assert"
)

func TestRenderParamsV2_LoopIndexInjection(t *testing.T) {
	// Setup
	executor := NewDefExecutor(0)

	dagIns := &entity.DagInstance{
		BaseInfo: entity.BaseInfo{ID: "test_dag_ins"},
		ShareData: &entity.ShareData{
			Dict: make(map[string]interface{}),
		},
	}
	dagIns.ShareData.DagInstance = dagIns

	// Task with current_iteration and template using index
	taskIns := &entity.TaskInstance{
		BaseInfo:   entity.BaseInfo{ID: "test_task"},
		TaskID:     "test_task_id",
		ActionName: "common",
		Params: map[string]interface{}{
			"current_iteration": 5,
			"test_index":        "{{index}}",
			"test_loop_index":   "{{__loop_index}}",
		},
		RelatedDagInstance: dagIns,
	}

	// Execute
	err := executor.renderParamsV2(taskIns)
	assert.NoError(t, err)

	// Verify
	params := taskIns.GetParams()
	// Index is injected into Env, not returned in Params unless explicitly requested.
	// So we don't check params["index"].

	// Check if templates were rendered correctly
	assert.Equal(t, 5, params["test_index"], "Template {{index}} should be rendered to 5")
	assert.Equal(t, 5, params["test_loop_index"], "Template {{__loop_index}} should be rendered to 5")
}
