package mgnt

import (
	"encoding/json"
	"testing"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
)

// TestBuildTaskInstanceFromEvents_ParallelExecution 测试并行执行时的事件乱序处理
func TestBuildTaskInstanceFromEvents_ParallelExecution(t *testing.T) {
	// 模拟真实的并行执行事件序列
	events := []*entity.DagInstanceEvent{
		// Task 0: trigger
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "0", Operator: "@trigger/dataflow-doc", Status: "running", Timestamp: 1769762005630182},
		{Type: rds.DagInstanceEventTypeVariable, Name: "__0", Data: map[string]any{"name": "转成pdf.pdf"}, Timestamp: 1769762005688897},
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "0", Operator: "@trigger/dataflow-doc", Status: "success", Timestamp: 1769762005694425},
		{Type: rds.DagInstanceEventTypeTrace, TaskID: "0", Name: "__0_trace", Data: map[string]any{"attempts": 0, "duration": 71, "max_retry": 3}, Timestamp: 1769762005696705},

		// Task 1006: split
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "1006", Operator: "@internal/text/split", Status: "running", Timestamp: 1769762005703635},
		{Type: rds.DagInstanceEventTypeVariable, Name: "__1006", Data: map[string]any{"slices": "{\"0\":\"123\"}"}, Timestamp: 1769762005705726},
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "1006", Operator: "@internal/text/split", Status: "success", Timestamp: 1769762005709578},
		{Type: rds.DagInstanceEventTypeTrace, TaskID: "1006", Name: "__1006_trace", Data: map[string]any{"attempts": 0, "duration": 7, "max_retry": 3}, Timestamp: 1769762005711287},

		// 并行分支: Task 1010 和 1009 的事件交错出现
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "1010", Operator: "@internal/tool/py3", Status: "blocked", Timestamp: 1769762005719755},
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "1009", Operator: "@internal/tool/py3", Status: "blocked", Timestamp: 1769762005720916},
		{Type: rds.DagInstanceEventTypeTrace, TaskID: "1010", Name: "__1010_trace", Data: map[string]any{"attempts": 0, "duration": 12, "max_retry": 3}, Timestamp: 1769762005730100},
		{Type: rds.DagInstanceEventTypeTrace, TaskID: "1009", Name: "__1009_trace", Data: map[string]any{"attempts": 0, "duration": 17, "max_retry": 3}, Timestamp: 1769762005737922},
		{Type: rds.DagInstanceEventTypeVariable, Name: "__1010", Data: map[string]any{"b": 1}, Timestamp: 1769762205875170},
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "1010", Operator: "@internal/tool/py3", Status: "success", Timestamp: 1769762205876861},
		{Type: rds.DagInstanceEventTypeVariable, Name: "__1009", Data: map[string]any{"a": 1}, Timestamp: 1769762205881930},
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "1009", Operator: "@internal/tool/py3", Status: "success", Timestamp: 1769762205883099},

		// Task 1008: join
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "1008", Operator: "@internal/text/join", Status: "running", Timestamp: 1769762207604973},
		{Type: rds.DagInstanceEventTypeVariable, Name: "__1008", Data: map[string]any{"text": "11"}, Timestamp: 1769762207606522},
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "1008", Operator: "@internal/text/join", Status: "success", Timestamp: 1769762207609781},
		{Type: rds.DagInstanceEventTypeTrace, TaskID: "1008", Name: "__1008_trace", Data: map[string]any{"attempts": 0, "duration": 6, "max_retry": 3}, Timestamp: 1769762207611207},
	}

	dagIns := &entity.DagInstance{
		BaseInfo: entity.BaseInfo{ID: "test-dag-ins"},
	}

	dag := &entity.Dag{
		Steps: []entity.Step{
			{ID: "0", Title: "Trigger", Operator: "@trigger/dataflow-doc"},
			{ID: "1006", Title: "Split Text", Operator: "@internal/text/split"},
			{ID: "1009", Title: "Python Script A", Operator: "@internal/tool/py3"},
			{ID: "1010", Title: "Python Script B", Operator: "@internal/tool/py3"},
			{ID: "1008", Title: "Join Text", Operator: "@internal/text/join"},
		},
	}

	// 执行重建
	tasks := buildTaskInstanceFromEvents(events, dagIns, dag)

	// 验证结果
	if len(tasks) != 5 {
		t.Fatalf("应该生成5个任务实例, 实际: %d", len(tasks))
	}

	// 验证任务顺序(按首次出现顺序)
	expectedOrder := []string{"0", "1006", "1010", "1009", "1008"}
	for i, task := range tasks {
		if task.TaskID != expectedOrder[i] {
			t.Errorf("任务顺序错误: 位置 %d 期望 %s, 实际 %s", i, expectedOrder[i], task.TaskID)
		}
	}

	// 验证并行任务的状态都正确更新
	task1009 := tasks[3]
	task1010 := tasks[2]

	if task1009.TaskID != "1009" {
		t.Errorf("Task1009 ID 错误: 期望 1009, 实际 %s", task1009.TaskID)
	}
	if task1009.Status != entity.TaskInstanceStatusSuccess {
		t.Errorf("Task 1009 状态错误: 期望 success, 实际 %s", task1009.Status)
	}
	if task1009.Results == nil {
		t.Error("Task 1009 应该有结果")
	}
	if results, ok := task1009.Results.(map[string]any); ok {
		if a, ok := results["a"].(int); !ok || a != 1 {
			t.Errorf("Task 1009 的结果应该包含 a=1, 实际: %v", results["a"])
		}
	}

	if task1010.TaskID != "1010" {
		t.Errorf("Task1010 ID 错误: 期望 1010, 实际 %s", task1010.TaskID)
	}
	if task1010.Status != entity.TaskInstanceStatusSuccess {
		t.Errorf("Task 1010 状态错误: 期望 success, 实际 %s", task1010.Status)
	}
	if task1010.Results == nil {
		t.Error("Task 1010 应该有结果")
	}
	if results, ok := task1010.Results.(map[string]any); ok {
		if b, ok := results["b"].(int); !ok || b != 1 {
			t.Errorf("Task 1010 的结果应该包含 b=1, 实际: %v", results["b"])
		}
	}

	// 验证 Trace 信息正确关联
	if task1009.MetaData == nil {
		t.Error("Task 1009 应该有 MetaData")
	} else if task1009.MetaData.Duration != 17 {
		t.Errorf("Task 1009 的 duration 应该是 17, 实际: %d", task1009.MetaData.Duration)
	}

	if task1010.MetaData == nil {
		t.Error("Task 1010 应该有 MetaData")
	} else if task1010.MetaData.Duration != 12 {
		t.Errorf("Task 1010 的 duration 应该是 12, 实际: %d", task1010.MetaData.Duration)
	}

	// 打印调试信息
	t.Logf("生成的任务实例:")
	for i, task := range tasks {
		resultsJSON, _ := json.MarshalIndent(task.Results, "", "  ")
		t.Logf("[%d] TaskID: %s, Status: %s, Operator: %s, Results: %s",
			i, task.TaskID, task.Status, task.ActionName, string(resultsJSON))
	}
}

// TestBuildTaskInstanceFromEvents_SequentialExecution 测试顺序执行的向后兼容性
func TestBuildTaskInstanceFromEvents_SequentialExecution(t *testing.T) {
	events := []*entity.DagInstanceEvent{
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "1", Operator: "op1", Status: "running", Timestamp: 1000},
		{Type: rds.DagInstanceEventTypeVariable, Name: "__1", Data: map[string]any{"result": "a"}, Timestamp: 2000},
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "1", Operator: "op1", Status: "success", Timestamp: 3000},
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "2", Operator: "op2", Status: "running", Timestamp: 4000},
		{Type: rds.DagInstanceEventTypeVariable, Name: "__2", Data: map[string]any{"result": "b"}, Timestamp: 5000},
		{Type: rds.DagInstanceEventTypeTaskStatus, TaskID: "2", Operator: "op2", Status: "success", Timestamp: 6000},
	}

	dagIns := &entity.DagInstance{
		BaseInfo: entity.BaseInfo{ID: "test-dag-ins"},
	}

	dag := &entity.Dag{
		Steps: []entity.Step{
			{ID: "1", Title: "Task 1", Operator: "op1"},
			{ID: "2", Title: "Task 2", Operator: "op2"},
		},
	}

	tasks := buildTaskInstanceFromEvents(events, dagIns, dag)

	if len(tasks) != 2 {
		t.Fatalf("应该生成2个任务实例, 实际: %d", len(tasks))
	}
	if tasks[0].TaskID != "1" {
		t.Errorf("Task 0 ID 错误: 期望 1, 实际 %s", tasks[0].TaskID)
	}
	if tasks[1].TaskID != "2" {
		t.Errorf("Task 1 ID 错误: 期望 2, 实际 %s", tasks[1].TaskID)
	}
	if tasks[0].Status != entity.TaskInstanceStatusSuccess {
		t.Errorf("Task 0 状态错误: 期望 success, 实际 %s", tasks[0].Status)
	}
	if tasks[1].Status != entity.TaskInstanceStatusSuccess {
		t.Errorf("Task 1 状态错误: 期望 success, 实际 %s", tasks[1].Status)
	}
}
