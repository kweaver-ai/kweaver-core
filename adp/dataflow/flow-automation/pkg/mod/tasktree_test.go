package mod

import (
	"testing"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/stretchr/testify/assert"
)

type mockTaskInfo struct {
	id       string
	dependOn []string
	status   entity.TaskInstanceStatus
}

func (m *mockTaskInfo) GetID() string {
	return m.id
}

func (m *mockTaskInfo) GetGraphID() string {
	return m.id
}

func (m *mockTaskInfo) GetDepend() []string {
	return m.dependOn
}

func (m *mockTaskInfo) GetStatus() entity.TaskInstanceStatus {
	return m.status
}

func TestBuildRootNode_DuplicateTasks(t *testing.T) {
	// Create two tasks with the same ID but different status (to simulate race update)
	task1 := &mockTaskInfo{
		id:     "task1",
		status: entity.TaskInstanceStatusInit,
	}
	task2 := &mockTaskInfo{
		id:     "task1", // Duplicate ID
		status: entity.TaskInstanceStatusSuccess,
	}

	// Order matters if we want the "last" one to win.
	// In the implemented logic, we iterate backwards and keep the first one we find (which is the last in the slice).
	tasks := []TaskInfoGetter{task1, task2}

	// This should succeed now
	root, err := BuildRootNode(tasks)
	assert.NoError(t, err)
	assert.NotNil(t, root)

	// Verify that the task in the tree has the status of the last task (Success)
	// BuildRootNode returns a virtual root. We need to find "task1" status from the graph.
	found := false
	walkNode(root, func(node *TaskNode) bool {
		if node.TaskInsID == "task1" {
			found = true
			assert.Equal(t, entity.TaskInstanceStatusSuccess, node.Status, "should use status from the last duplicate task")
			return false
		}
		return true
	}, true)
	assert.True(t, found, "task1 not found in tree")
}
