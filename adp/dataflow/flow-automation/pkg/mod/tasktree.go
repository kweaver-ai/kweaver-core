package mod

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/log"
)

const (
	virtualTaskRootID = "_virtual_root"
)

// TaskInfoGetter
type TaskInfoGetter interface {
	GetDepend() []string
	GetID() string
	GetGraphID() string
	GetStatus() entity.TaskInstanceStatus
}

// MapTaskInsToGetter
func MapTaskInsToGetter(taskIns []*entity.TaskInstance) (ret []TaskInfoGetter) {
	for i := range taskIns {
		ret = append(ret, taskIns[i])
	}
	return
}

// MapTasksToGetter
func MapTasksToGetter(taskIns []entity.Task) (ret []TaskInfoGetter) {
	for i := range taskIns {
		ret = append(ret, &taskIns[i])
	}
	return
}

// // MapMockTasksToGetter
// func MapMockTasksToGetter(taskIns []*MockTaskInfoGetter) (ret []TaskInfoGetter) {
// 	for i := range taskIns {
// 		ret = append(ret, taskIns[i])
// 	}
// 	return
// }

// MustBuildRootNode
func MustBuildRootNode(tasks []TaskInfoGetter) *TaskNode {
	root, err := BuildRootNode(tasks)
	if err != nil {
		panic(fmt.Errorf("build tasks failed: %s", err))
	}
	return root
}

// BuildRootNode
func BuildRootNode(tasks []TaskInfoGetter) (*TaskNode, error) {
	root := &TaskNode{
		TaskInsID: virtualTaskRootID,
		Status:    entity.TaskInstanceStatusSuccess,
	}
	m, err := buildGraphNodeMap(tasks)
	if err != nil {
		return nil, err
	}

	processed := make(map[string]bool)
	for i := len(tasks) - 1; i >= 0; i-- {
		id := tasks[i].GetGraphID()
		if processed[id] {
			continue
		}
		processed[id] = true

		if len(tasks[i].GetDepend()) == 0 {
			n := m[id]
			n.AppendParent(root)
			root.children = append(root.children, n)
		}

		if len(tasks[i].GetDepend()) > 0 {
			for _, dependId := range tasks[i].GetDepend() {
				parent, ok := m[dependId]
				if !ok {
					// Check if it's a self-dependency (which shouldn't happen but good to be safe)
					if dependId == id {
						continue
					}
					// If strict mode, return error. But for now maybe just warn?
					// The original code returned error. Let's keep it strict for dependencies.
					return nil, fmt.Errorf("does not find task[%s] depend: %s", id, dependId)
				}
				parent.AppendChild(m[id])
				m[id].AppendParent(parent)
			}
		}
	}

	if len(root.children) == 0 {
		return nil, errors.New("here is no start nodes")
	}

	if cycleStart := root.HasCycle(); cycleStart != nil {
		return nil, fmt.Errorf("dag has cycle at: %s", cycleStart.TaskInsID)
	}

	return root, nil
}

func buildGraphNodeMap(tasks []TaskInfoGetter) (map[string]*TaskNode, error) {
	m := map[string]*TaskNode{}
	for i := range tasks {
		if _, ok := m[tasks[i].GetGraphID()]; ok {
			log.Warnf("task id is repeat, id: %s", tasks[i].GetGraphID())
		}
		m[tasks[i].GetGraphID()] = NewTaskNodeFromGetter(tasks[i])
	}
	return m, nil
}

// TaskTree
type TaskTree struct {
	DagIns *entity.DagInstance
	Root   *TaskNode
}

// NewTaskNodeFromGetter
func NewTaskNodeFromGetter(instance TaskInfoGetter) *TaskNode {
	return &TaskNode{
		TaskInsID: instance.GetID(),
		Status:    instance.GetStatus(),
	}
}

// TaskNode
type TaskNode struct {
	TaskInsID string
	Status    entity.TaskInstanceStatus

	children []*TaskNode
	parents  []*TaskNode
}

type TreeStatus string

const (
	TreeStatusRunning TreeStatus = "running"
	TreeStatusSuccess TreeStatus = "success"
	TreeStatusFailed  TreeStatus = "failed"
	TreeStatusBlocked TreeStatus = "blocked"
)

// HasCycle
func (t *TaskNode) HasCycle() (cycleStart *TaskNode) {
	visited, incomplete := map[string]struct{}{}, map[string]*TaskNode{}
	waitQueue := []*TaskNode{t}
	bfsCheckCycle(waitQueue, visited, incomplete)
	if len(incomplete) > 0 {
		for k := range incomplete {
			return incomplete[k]
		}
	}
	return
}

func bfsCheckCycle(waitQueue []*TaskNode, visited map[string]struct{}, incomplete map[string]*TaskNode) {
	queueLen := len(waitQueue)
	if queueLen == 0 {
		return
	}

	isParentCompleted := func(node *TaskNode) bool {
		for _, p := range node.parents {
			if _, ok := visited[p.TaskInsID]; !ok {
				return false
			}
		}
		return true
	}

	for i := 0; i < queueLen; i++ {
		cur := waitQueue[i]
		if !isParentCompleted(cur) {
			incomplete[cur.TaskInsID] = cur
			continue
		}
		visited[cur.TaskInsID] = struct{}{}
		delete(incomplete, cur.TaskInsID)
		for _, c := range cur.children {
			waitQueue = append(waitQueue, c)
		}
	}
	waitQueue = waitQueue[queueLen:]
	bfsCheckCycle(waitQueue, visited, incomplete)
	return
}

// ComputeStatus
func (t *TaskNode) ComputeStatus() (status TreeStatus, srcTaskInsId string) {
	walkNode(t, func(node *TaskNode) bool {
		switch node.Status {
		case entity.TaskInstanceStatusFailed, entity.TaskInstanceStatusCanceled:
			status = TreeStatusFailed
			srcTaskInsId = node.TaskInsID
			return true
		case entity.TaskInstanceStatusBlocked:
			status = TreeStatusBlocked
			srcTaskInsId = node.TaskInsID
			return true
		case entity.TaskInstanceStatusSuccess, entity.TaskInstanceStatusSkipped:
			return true
		default:
			status = TreeStatusRunning
			srcTaskInsId = node.TaskInsID
			return false
		}
	}, false)
	if srcTaskInsId != "" {
		return
	}
	return TreeStatusSuccess, ""
}

func walkNode(root *TaskNode, walkFunc func(node *TaskNode) bool, walkChildrenIgnoreStatus bool) {
	dfsWalk(root, walkFunc, walkChildrenIgnoreStatus)
}

func dfsWalk(
	root *TaskNode,
	walkFunc func(node *TaskNode) bool,
	walkChildrenIgnoreStatus bool) bool {

	if root.TaskInsID != virtualTaskRootID {
		if !walkFunc(root) {
			return false
		}
	}

	// we cannot execute children, but should execute brother nodes
	if !walkChildrenIgnoreStatus && !root.CanExecuteChild() {
		return true
	}
	for _, c := range root.children {
		// if children's parent is not just root, we must check it
		if len(c.parents) > 1 && !c.CanBeExecuted() {
			continue
		}

		if !dfsWalk(c, walkFunc, walkChildrenIgnoreStatus) {
			return false
		}
	}
	return true
}

// AppendChild
func (t *TaskNode) AppendChild(task *TaskNode) {
	t.children = append(t.children, task)
}

// AppendParent
func (t *TaskNode) AppendParent(task *TaskNode) {
	t.parents = append(t.parents, task)
}

// CanExecuteChild
func (t *TaskNode) CanExecuteChild() bool {
	return t.Status == entity.TaskInstanceStatusSuccess || t.Status == entity.TaskInstanceStatusSkipped
	// return t.Status == entity.TaskInstanceStatusSuccess || t.Status == entity.TaskInstanceStatusSkipped
}

// CanBeExecuted check whether task could be executed
func (t *TaskNode) CanBeExecuted() bool {
	if len(t.parents) == 0 {
		return true
	}

	for _, p := range t.parents {
		if !p.CanExecuteChild() {
			return false
		}
	}
	return true
}

// GetExecutableTaskIds is unique task id map
func (t *TaskNode) GetExecutableTaskIds() (executables []string) {
	walkNode(t, func(node *TaskNode) bool {
		if node.Executable() {
			executables = append(executables, node.TaskInsID)
		}
		return true
	}, false)
	return
}

// GetNextTaskIds
func (t *TaskNode) GetNextTaskIds(completedOrRetryTask *entity.TaskInstance) (executable []string, find bool) {
	// 特殊处理循环任务
	if completedOrRetryTask.ActionName == common.Loop {
		// 检查失败原因
		if completedOrRetryTask.Status == entity.TaskInstanceStatusFailed {
			log.Errorf("Loop task [%s] failed with reason: %v",
				completedOrRetryTask.ID, completedOrRetryTask.Reason)
			if completedOrRetryTask.Results != nil {
				log.Errorf("Loop task [%s] results: %+v",
					completedOrRetryTask.ID, completedOrRetryTask.Results)
			}
		} else if completedOrRetryTask.Status == entity.TaskInstanceStatusCanceled {
			log.Errorf("Loop task [%s] was canceled with reason: %v",
				completedOrRetryTask.ID, completedOrRetryTask.Reason)
		}

		// 如果循环任务执行成功，找出它的后续节点执行
		if completedOrRetryTask.Status == entity.TaskInstanceStatusSuccess || completedOrRetryTask.Status == entity.TaskInstanceStatusSkipped {
			// 先查找循环节点自身，更新其状态
			loopNodeFound := false
			walkNode(t, func(node *TaskNode) bool {
				if completedOrRetryTask.ID == node.TaskInsID {
					loopNodeFound = true
					find = true

					// 记录节点状态变化
					node.Status = completedOrRetryTask.Status

					return false
				}
				return true
			}, false)

			if !loopNodeFound {
				log.Warnf("Loop node [%s] not found in task tree during first pass, trying to rebuild tree",
					completedOrRetryTask.ID)
			}

			// 查找根节点下的所有节点，找到循环节点和它的下一个兄弟节点
			walkNode(t, func(node *TaskNode) bool {
				// 找到循环节点
				if completedOrRetryTask.ID == node.TaskInsID {
					find = true

					// 记录节点状态变化
					node.Status = completedOrRetryTask.Status

					// 先查看是否有子节点需要执行
					hasExecutableChildren := false
					for i := range node.children {
						canExec := node.children[i].Executable()

						if canExec {
							executable = append(executable, node.children[i].TaskInsID)
							hasExecutableChildren = true
						}
					}

					// 如果没有可执行的子节点，查找父节点的下一个子节点（兄弟节点）
					if !hasExecutableChildren && len(node.parents) > 0 {
						for _, parent := range node.parents {
							if parent.TaskInsID == virtualTaskRootID {
								// 对于根节点的直接子节点，单独处理

								// 查找当前节点在父节点的位置
								currentNodeIdx := -1
								for i, sibling := range parent.children {
									if sibling.TaskInsID == node.TaskInsID {
										currentNodeIdx = i
										break
									}
								}

								if currentNodeIdx == -1 {
									log.Warnf("Loop task [%s] not found in root's children", node.TaskInsID)
									continue
								}

								for i := range parent.children {
									if i == currentNodeIdx {
										continue // 跳过自己
									}

									// 查找下一个兄弟节点
									if i > currentNodeIdx {
										if i+1 < len(parent.children) {
											nextSibling := parent.children[i]
											canExec := nextSibling.Executable()

											if canExec {
												executable = append(executable, nextSibling.TaskInsID)
											} else {
												// 如果是Init状态，检查它的依赖
												if nextSibling.Status == entity.TaskInstanceStatusInit {
													canBeExec := nextSibling.CanBeExecuted()

													if canBeExec {
														executable = append(executable, nextSibling.TaskInsID)
													} else {
														// 检查父节点是否允许执行子节点
														for range nextSibling.parents {
															// 检查父节点状态
														}
													}
												}
											}
										}
										break // 只查找第一个后续节点
									}
								}
							} else {
								// 对于非根节点的子节点，查找父节点的所有子节点中当前节点的位置

								// 查找当前节点在父节点的位置
								currentNodeIdx := -1
								for i, sibling := range parent.children {
									if sibling.TaskInsID == node.TaskInsID {
										currentNodeIdx = i
										break
									}
								}

								if currentNodeIdx == -1 {
									log.Warnf("Loop task [%s] not found in parent's [%s] children",
										node.TaskInsID, parent.TaskInsID)
									continue
								}

								for i := range parent.children {
									if i == currentNodeIdx {
										continue // 跳过自己
									}

									// 查找下一个兄弟节点
									if i > currentNodeIdx {
										if i+1 < len(parent.children) {
											nextSibling := parent.children[i]
											canExec := nextSibling.Executable()

											if canExec {
												executable = append(executable, nextSibling.TaskInsID)
											} else {
												// 如果下一个节点不可执行，但是Init状态，直接标记为可执行
												if nextSibling.Status == entity.TaskInstanceStatusInit {
													executable = append(executable, nextSibling.TaskInsID)
												}
											}
										}
										break // 只查找第一个后续节点
									}
								}
							}
						}
					}

					// 找到了节点，停止遍历
					return false
				}
				// 继续遍历其他节点
				return true
			}, true) // 允许遍历所有节点，不受状态限制

			if find {
				return executable, find
			} else {
				log.Warnf("Loop task [%s] not found in task tree during search",
					completedOrRetryTask.ID)
			}
		} else if completedOrRetryTask.Status == entity.TaskInstanceStatusInit {
			// 如果循环任务刚初始化，将它标记为可执行
			executable = append(executable, completedOrRetryTask.ID)
			find = true
			return executable, find
		} else if completedOrRetryTask.Status == entity.TaskInstanceStatusFailed ||
			completedOrRetryTask.Status == entity.TaskInstanceStatusCanceled {
			// 循环任务失败或取消了，但我们允许它重试
			executable = append(executable, completedOrRetryTask.ID)
			find = true
			return executable, find
		}

		// 如果循环任务正在进行中，返回找到但没有后续任务
		find = true
		return []string{}, true
	}

	walkNode(t, func(node *TaskNode) bool {
		if completedOrRetryTask.ID == node.TaskInsID {
			find = true
			node.Status = completedOrRetryTask.Status

			if node.Status == entity.TaskInstanceStatusInit {
				executable = append(executable, node.TaskInsID)
				return false
			}

			if !node.CanExecuteChild() {
				return false
			}
			for i := range node.children {
				if node.children[i].Executable() {
					executable = append(executable, node.children[i].TaskInsID)
				}
			}
			return false
		}
		return true
	}, false)

	return
}

// Executable
func (t *TaskNode) Executable() bool {
	// 循环节点总是可以执行，允许从失败状态恢复
	if t.Status == entity.TaskInstanceStatusInit ||
		t.Status == entity.TaskInstanceStatusRetrying ||
		t.Status == entity.TaskInstanceStatusEnding ||
		(t.Status == entity.TaskInstanceStatusFailed && strings.HasPrefix(t.TaskInsID, "loop_")) {

		if len(t.parents) == 0 {
			return true
		}

		for i := range t.parents {
			if !t.parents[i].CanExecuteChild() {
				// 对于循环节点，即使父节点不允许执行子节点，也可以执行
				// 这是为了允许循环节点在失败后仍然可以继续执行
				if strings.HasPrefix(t.TaskInsID, "loop_") {
					continue
				}
				return false
			}
		}
		return true
	}
	return false
}
