# 流程并行执行逻辑设计文档

## 1. 概述

### 1.1 背景

当前流程执行系统基于DAG（有向无环图）模型，任务通过`DependOn`字段定义依赖关系。系统已经具备多worker并发执行能力，Executor层可以并行执行不同任务。

**核心问题：缺乏并行分支节点的定义机制**

当前系统虽然支持条件分支节点（`@control/flow/branches`）和循环节点（`@control/flow/loop`），但**缺乏并行分支节点**来显式定义并行执行。这导致：

1. **无法显式定义并行执行**：虽然可以通过显式的`dependOn`字段让多个任务依赖同一个前置任务来实现并行，但这种方式不够直观，且容易出错
2. **默认串行依赖**：在`buildTasks`阶段，如果没有显式定义`dependOn`，每个step默认依赖前一个step（`logics/mgnt/mgnt.go:3196-3197`），导致即使可以并行的任务也被强制串行化
3. **缺乏并行语义**：没有类似`@control/flow/branches`的`@control/flow/parallel`节点来明确表达"这些分支应该并行执行"的意图

因此，需要引入`@control/flow/parallel`并行分支节点，提供显式的并行执行定义机制，让用户可以清晰地表达并行执行意图。

### 1.2 目标
- 支持通过`@control/flow/parallel`并行分支节点定义并行执行
- 支持同一DAG实例中无依赖关系的任务并行执行
- 保持依赖关系的正确性，确保任务按依赖顺序执行
- 优化资源利用率，提高流程执行效率
- 保持系统稳定性，避免引入并发安全问题

### 1.3 范围
本文档涵盖：
- 并行分支节点（`@control/flow/parallel`）的流程定义规范
- Steps到Tasks的转换逻辑（包括并行分支节点的展开）
- 并行执行的实现方案
- 并发安全保障措施
- 性能优化方案

## 2. 交互设计示意图

下图展示了并行执行的交互设计示意图，直观地说明了并行分支节点的执行流程和任务依赖关系：

![并行执行交互设计示意图](./parallel.jpg)

该示意图展示了：
- 并行分支节点的定义结构
- 多个分支的并行执行关系
- 分支内任务的顺序执行
- 并行分支完成后的汇聚点

## 3. 当前架构分析

### 3.1 现有执行流程

```
DAG实例启动
    ↓
Parser解析任务树（TaskTree）
    ↓
找出可执行任务（GetExecutableTaskIds）
    ↓
任务路由到Executor（通过murmur3 hash，同一DAG实例路由到同一worker）
    ↓
任务执行
    ↓
任务完成，通知Parser（EntryTaskIns）
    ↓
找出下一个可执行任务（GetNextTaskIds）
    ↓
循环执行
```

### 3.2 当前并发模型

**Parser层：**
- 多个worker goroutine（`workerNumber`个）
- 使用murmur3 hash将同一DAG实例的任务路由到同一worker（`pkg/mod/parser.go:765-770`）
- **目的**：避免同一DAG实例的并发写入冲突，特别是ShareData的并发访问

**Executor层：**
- 按优先级分类的worker队列（Highest, High, Medium, Low, Lowest）
- 每个优先级有多个worker goroutine
- 任务通过优先级队列分发到对应的worker
- **已支持并行执行**：不同任务可以在不同的executor worker中并行执行

### 3.3 关键组件

**TaskTree（任务树）：**
- 位置：`pkg/mod/tasktree.go`
- 功能：维护任务依赖关系，管理任务状态
- 核心方法：
  - `GetExecutableTaskIds()`：获取当前可执行的任务列表
  - `GetNextTaskIds()`：获取指定任务完成后的下一个可执行任务
  - `CanBeExecuted()`：检查任务是否满足依赖条件

**buildTasks函数：**
- 位置：`logics/mgnt/mgnt.go:3173`
- 功能：将Steps转换为Tasks，建立依赖关系
- 当前支持：
  - 条件分支节点：`@control/flow/branches`（line 3206-3218）
  - 循环节点：`@control/flow/loop`（line 3219-3229）
  - 普通步骤：默认依赖前一个步骤（line 3196-3197）

**ShareData：**
- 位置：`pkg/entity/dag.go`
- 功能：存储DAG实例的共享数据，供各个任务访问
- **当前状态**：无并发控制机制（无锁）

## 4. 并行分支流程定义规范

### 4.1 并行分支节点定义

通过`@control/flow/parallel`节点定义并行执行，结构类似条件分支节点：

```json
{
  "id": "1",
  "title": "并行处理节点",
  "operator": "@control/flow/parallel",
  "branches": [
    {
      "steps": [
        {"id": "1-1", "operator": "...", "parameters": {...}}
      ]
    },
    {
      "steps": [
        {"id": "2-1", "operator": "...", "parameters": {...}}
      ]
    },
    {
      "steps": [
        {"id": "3-1", "operator": "...", "parameters": {...}},
        {"id": "3-2", "operator": "...", "parameters": {...}}
      ]
    }
  ]
}
```

**关键特性：**
- `operator`：固定为 `@control/flow/parallel`
- `branches`：数组，每个元素是一个分支
- 每个分支包含 `steps` 数组，定义该分支的执行步骤
- 所有分支的第一个step会并行执行
- 分支内的steps按顺序执行

### 4.2 完整流程示例

```json
{
  "title": "并行处理示例流程",
  "description": "使用并行分支节点定义并行执行",
  "status": "normal",
  "steps": [
    {
      "id": "0",
      "title": "准备数据",
      "operator": "@anyshare-trigger/upload-file",
      "parameters": {
        "docid": "gns://9A8C9277947D4898A350427C768A194C/B06A19BD068B40E788360ED54FA3CB7D",
        "inherit": true
      }
    },
    {
      "id": "1",
      "title": "并行处理节点",
      "operator": "@control/flow/parallel",
      "branches": [
        {
          "steps": [
            {
              "id": "1-1",
              "title": "处理文件A",
              "operator": "@anyshare/file/copy",
              "parameters": {
                "docid": "{{__0.id}}",
                "destparent": "gns://9A8C9277947D4898A350427C768A194C/pathA",
                "ondup": 2
              }
            }
          ]
        },
        {
          "steps": [
            {
              "id": "2-1",
              "title": "处理文件B",
              "operator": "@anyshare/file/copy",
              "parameters": {
                "docid": "{{__0.id}}",
                "destparent": "gns://9A8C9277947D4898A350427C768A194C/pathB",
                "ondup": 2
              }
            }
          ]
        },
        {
          "steps": [
            {
              "id": "3-1",
              "title": "处理文件C",
              "operator": "@anyshare/file/copy",
              "parameters": {
                "docid": "{{__0.id}}",
                "destparent": "gns://9A8C9277947D4898A350427C768A194C/pathC",
                "ondup": 2
              }
            },
            {
              "id": "3-2",
              "title": "处理文件C的后续操作",
              "operator": "@anyshare/file/process",
              "parameters": {
                "docid": "{{__3-1.new_docid}}"
              }
            }
          ]
        }
      ]
    },
    {
      "id": "4",
      "title": "合并结果",
      "operator": "@anyshare/file/merge",
      "parameters": {
        "docid_a": "{{__1-1.new_docid}}",
        "docid_b": "{{__2-1.new_docid}}",
        "docid_c": "{{__3-2.new_docid}}",
        "destparent": "gns://9A8C9277947D4898A350427C768A194C/result"
      }
    }
  ]
}
```

**执行流程分析：**

1. **Step 0**：准备数据，无依赖，首先执行
2. **Step 1**：并行分支节点`@control/flow/parallel`，包含3个分支：
   - **Branch 1**：处理文件A（Step 1-1）
   - **Branch 2**：处理文件B（Step 2-1）
   - **Branch 3**：处理文件C（Step 3-1）和后续操作（Step 3-2）
   - 三个分支**并行执行**，每个分支内的steps按顺序执行
3. **Step 4**：合并结果，依赖所有并行分支的结果

**执行时间线：**
```
Step 0 ────────── (准备数据)
        ↓
Step 1 (@control/flow/parallel)
  ├─ Branch 1: Step 1-1 ─┐
  ├─ Branch 2: Step 2-1 ─┤ 并行执行
  └─ Branch 3: Step 3-1 → Step 3-2 ─┘
        ↓
Step 4 ────────── (合并结果)
```

## 5. Steps到Tasks的转换逻辑

### 5.1 转换规则

系统会将`steps`转换为`tasks`，并建立依赖关系。对于并行分支节点`@control/flow/parallel`，系统会将其展开为多个并行执行的tasks。

**转换规则：**
1. **普通step**：转换为一个task，依赖前一个task（如果存在）
2. **条件分支节点** `@control/flow/branches`：生成一个分支控制task，每个分支递归调用buildTasks
3. **循环节点** `@control/flow/loop`：生成一个循环控制task，包含循环体的steps
4. **并行分支节点** `@control/flow/parallel`：
   - 并行分支节点本身**不生成task**
   - 每个分支下的第一个step依赖并行分支节点的前置step
   - 同一并行分支节点内的不同分支的第一个step可以并行执行
   - 分支内的后续steps按顺序执行，依赖分支内的前一个step
   - 并行分支节点后的step依赖所有分支的最后一个step

### 5.2 转换示例

**Steps定义：**
```json
{
  "steps": [
    {
      "id": "0",
      "operator": "@anyshare-trigger/upload-file",
      "parameters": {"docid": "..."}
    },
    {
      "id": "1",
      "operator": "@control/flow/parallel",
      "branches": [
        {
          "steps": [
            {"id": "1-1", "operator": "@anyshare/file/copy", "parameters": {"docid": "{{__0.id}}"}}
          ]
        },
        {
          "steps": [
            {"id": "2-1", "operator": "@anyshare/file/copy", "parameters": {"docid": "{{__0.id}}"}}
          ]
        },
        {
          "steps": [
            {"id": "3-1", "operator": "@anyshare/file/copy", "parameters": {"docid": "{{__0.id}}"}},
            {"id": "3-2", "operator": "@anyshare/file/process", "parameters": {"docid": "{{__3-1.new_docid}}"}}
          ]
        }
      ]
    },
    {
      "id": "4",
      "operator": "@anyshare/file/merge",
      "parameters": {
        "docid_a": "{{__1-1.new_docid}}",
        "docid_b": "{{__2-1.new_docid}}",
        "docid_c": "{{__3-2.new_docid}}"
      }
    }
  ]
}
```

**转换为Tasks：**
```json
{
  "tasks": [
    {"id": "0", "dependOn": []},
    {"id": "1-1", "dependOn": ["0"]},
    {"id": "2-1", "dependOn": ["0"]},
    {"id": "3-1", "dependOn": ["0"]},
    {"id": "3-2", "dependOn": ["3-1"]},
    {"id": "4", "dependOn": ["1-1", "2-1", "3-2"]}
  ]
}
```

**转换说明：**
- Step 0：转换为task "0"，无依赖
- Step 1（并行分支节点）：不生成task，但展开其分支：
  - Branch 1的Step 1-1：转换为task "1-1"，依赖Step 0
  - Branch 2的Step 2-1：转换为task "2-1"，依赖Step 0（与1-1并行）
  - Branch 3的Step 3-1：转换为task "3-1"，依赖Step 0（与前两个并行）
  - Branch 3的Step 3-2：转换为task "3-2"，依赖分支内的Step 3-1
- Step 4：转换为task "4"，依赖所有并行分支的最后一步（1-1, 2-1, 3-2）

**并行执行关系：**
- Tasks "1-1", "2-1", "3-1" 可以并行执行（都依赖Step 0）
- Task "3-2" 需要等待 "3-1" 完成（分支内顺序执行）
- Task "4" 需要等待所有并行分支完成（"1-1", "2-1", "3-2"）

### 5.3 Task ID生成规则

为了确保Task ID的唯一性和可追溯性，并行分支节点的Task ID生成规则如下：

```
格式：{parallel_step_id}_b{branch_index}_s{step_index}

示例：
- 并行节点ID为 "1"，第1个分支的第1个step：1_b0_s0
- 并行节点ID为 "1"，第2个分支的第1个step：1_b1_s0
- 并行节点ID为 "1"，第3个分支的第2个step：1_b2_s1
```

**说明：**
- `parallel_step_id`：并行分支节点的ID
- `branch_index`：分支索引，从0开始
- `step_index`：分支内step的索引，从0开始

## 6. 实现方案

### 6.1 buildTasks函数改动

在`logics/mgnt/mgnt.go`的`buildTasks`函数中添加并行分支节点的处理逻辑：

```go
func (m *mgnt) buildTasks(triggerStep *entity.Step, steps []entity.Step, tasks *[]entity.Task, bch *entity.Branch, stepList *[]map[string]interface{}, inheritChecks *entity.PreChecks, branchsID *string) {
    // ... 前置逻辑 ...
    
    for _, step := range steps {
        dependOn := []string{}
        if len(*tasks) != 0 {
            dependOn = append(dependOn, (*tasks)[len(*tasks)-1].ID)
        }
        
        // ... stepMap处理 ...
        
        if step.Operator == actions.ControlFlowBranches {
            // 条件分支节点处理（已存在）
            // ...
        } else if step.Operator == common.Loop {
            // 循环节点处理（已存在）
            // ...
        } else if step.Operator == "@control/flow/parallel" {
            // 并行分支节点处理（新增）
            parallelBranchLastTasks := []string{}
            
            for branchIndex, branch := range step.Branches {
                branchLastTaskID := ""
                
                for stepIndex, branchStep := range branch.Steps {
                    // 生成Task ID
                    stepTaskID := fmt.Sprintf("%s_b%d_s%d", step.ID, branchIndex, stepIndex)
                    
                    // 确定依赖关系
                    stepDependOn := dependOn  // 第一个step依赖并行节点的前置step
                    if stepIndex > 0 {
                        // 分支内的后续step依赖前一个step
                        prevStepTaskID := fmt.Sprintf("%s_b%d_s%d", step.ID, branchIndex, stepIndex-1)
                        stepDependOn = []string{prevStepTaskID}
                    }
                    
                    // 创建Task
                    task := entity.Task{
                        ID:         stepTaskID,
                        ActionName: branchStep.Operator,
                        Name:       branchStep.Title,
                        DependOn:   stepDependOn,
                        Params:     branchStep.Parameters,
                        PreChecks:  prechecks,
                        Settings:   branchStep.Settings,
                    }
                    
                    // 设置超时时间
                    if branchStep.Settings == nil {
                        task.TimeoutSecs = m.taskTimeoutConfig.GetTimeout(branchStep.Operator)
                    } else {
                        task.TimeoutSecs = branchStep.Settings.TimeOut.Delay
                    }
                    task.TimeoutSecs += 60  // 看门狗时间窗口
                    
                    *tasks = append(*tasks, task)
                    branchLastTaskID = stepTaskID
                }
                
                parallelBranchLastTasks = append(parallelBranchLastTasks, branchLastTaskID)
            }
            
            // 更新dependOn，后续step依赖所有分支的最后一个task
            dependOn = parallelBranchLastTasks
            
        } else {
            // 普通step的处理逻辑（已存在）
            // ...
        }
    }
}
```

### 6.2 并行执行机制

**当前架构已支持并行执行：**

1. **TaskTree识别并行任务**：
   - `GetExecutableTaskIds()` 方法会返回所有满足依赖条件的任务
   - 对于并行分支节点展开后的tasks，如果它们都依赖同一个前置task，且该前置task已完成，则它们都会被识别为可执行任务

2. **Parser层路由**：
   - 虽然同一DAG实例的任务通过murmur3 hash路由到同一worker
   - 但该worker会将所有可执行任务推送到Executor

3. **Executor层并行执行**：
   - Executor层有多个worker goroutine
   - 不同的任务可以在不同的executor worker中并行执行
   - 这是**真正的并行执行**

**执行流程：**
```
Step 0完成 → Parser Worker
    ↓
GetNextTaskIds() 返回 ["1-1", "2-1", "3-1"]
    ↓
Parser Worker将3个任务推送到Executor
    ↓
Executor的3个worker并行执行这3个任务
    ↓
任务完成后，继续执行后续任务
```

### 6.3 并发安全保障

**关键问题：ShareData并发访问**

当前ShareData没有并发控制机制，但由于murmur3 hash绑定，同一DAG实例的任务在同一Parser Worker中处理，避免了并发写入冲突。

**保持现状的原因：**
1. **murmur3 hash绑定保证了ShareData的串行访问**：虽然任务在Executor层并行执行，但ShareData的更新都在同一Parser Worker中进行
2. **避免引入复杂的并发控制**：不需要为ShareData添加锁机制
3. **系统稳定性优先**：保持现有架构的稳定性

**未来优化方向：**
如果需要进一步提升并行度，可以考虑：
1. 实现`SafeShareData`，添加`sync.RWMutex`并发控制
2. 移除murmur3 hash绑定，允许同一DAG实例的任务分发到不同Parser Worker
3. 实现更细粒度的负载均衡

## 7. 性能优化

### 7.1 并行度优化

**当前并行度：**
- Executor层已支持多worker并行执行
- 并行分支节点展开后的tasks可以在Executor层并行执行
- 并行度受限于Executor worker数量和任务优先级

**优化方向：**
1. **动态调整Executor worker数量**：根据系统负载动态调整
2. **优先级优化**：为并行任务设置合适的优先级
3. **资源隔离**：为不同类型的任务分配不同的资源池

### 7.2 性能监控

**关键指标：**
1. **并行度指标**
   - 平均并行任务数
   - 最大并行任务数
   - 并行度分布

2. **性能指标**
   - 任务执行时间
   - DAG实例总执行时间
   - 吞吐量（任务数/秒）

3. **资源指标**
   - CPU使用率
   - 内存使用率
   - Goroutine数量

## 8. 测试方案

### 8.1 单元测试

**测试用例：**
1. **并行分支节点转换测试**
   - 验证并行分支节点正确展开为tasks
   - 验证Task ID生成规则正确
   - 验证依赖关系正确建立

2. **依赖关系测试**
   - 验证并行分支的第一个step依赖前置step
   - 验证分支内steps按顺序依赖
   - 验证后续step依赖所有分支的最后一个step

3. **嵌套并行分支节点测试**
   - 验证并行分支节点内嵌套并行分支节点的正确性

### 8.2 集成测试

**测试场景：**
1. **简单并行执行**
   - 3个并行分支，每个分支1个step
   - 验证3个任务并行执行

2. **复杂并行执行**
   - 3个并行分支，每个分支多个steps
   - 验证分支内顺序执行，分支间并行执行

3. **并行+条件分支**
   - 并行分支节点后接条件分支节点
   - 验证执行逻辑正确

4. **并行+循环**
   - 并行分支节点内包含循环节点
   - 验证执行逻辑正确

### 8.3 性能测试

**测试指标：**
1. **并行加速比**：对比串行执行和并行执行的时间
2. **资源消耗**：CPU、内存、Goroutine数量
3. **吞吐量**：单位时间内完成的任务数

## 9. 风险评估

### 9.1 高风险点

1. **Task ID冲突**（严重性：高）
   - 风险：并行分支节点的Task ID可能与现有Task ID冲突
   - 缓解措施：使用特定的ID生成规则，添加唯一性检查

2. **依赖关系错误**（严重性：高）
   - 风险：并行分支节点的依赖关系处理不当，导致执行顺序错误
   - 缓解措施：充分的单元测试和集成测试

### 9.2 中风险点

1. **性能下降**（严重性：中）
   - 风险：并行分支节点的处理逻辑可能导致性能下降
   - 缓解措施：性能测试，优化算法

2. **嵌套并行分支节点**（严重性：中）
   - 风险：嵌套并行分支节点可能导致复杂度增加
   - 缓解措施：添加嵌套测试用例，限制嵌套深度
