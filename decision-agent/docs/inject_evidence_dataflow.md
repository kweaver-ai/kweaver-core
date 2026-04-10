# Inject Evidence 完整数据流

## 1. 概述

Inject Evidence 功能的目标是：
1. 从工具调用结果中提取实体（如人名、地名、组织名等）
2. 在 LLM 生成包含这些实体引用的文本时，自动标注实体位置
3. 在响应中返回 `_evidence` 元数据，包含实体名称和位置索引

## 2. 完整数据流

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Agent 执行流程                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐      │
│  │  run_dolphin()   │────▶│  process_arun_  │────▶│  create_        │      │
│  │                 │     │  loop()         │     │  evidence_      │      │
│  │                 │     │                 │     │  injection_     │      │
│  │                 │     │                 │     │  stream()       │      │
│  └─────────────────┘     └─────────────────┘     └─────────────────┘      │
│           │                       │                       │                  │
│           ▼                       ▼                       ▼                  │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │                         每个 item (agent.arun() 迭代)                   │ │
│  │  ┌─────────────────────────────────────────────────────────────────┐ │ │
│  │  │  item: {                                                      │ │ │
│  │  │    "answer": "LLM生成的文本或工具结果",                         │ │ │
│  │  │    "thinking": "...",                                          │ │ │
│  │  │    "_progress": [                                              │ │ │
│  │  │      {                                                        │ │ │
│  │  │        "stage": "llm" | "skill" | "assign",                   │ │ │
│  │  │        "answer": "该阶段的输出",                               │ │ │
│  │  │        "skill_info": {...}                                     │ │ │
│  │  │      },                                                       │ │ │
│  │  │      ...                                                      │ │ │
│  │  │    ],                                                         │ │ │
│  │  │    ...其他字段                                                │ │ │
│  │  │  }                                                           │ │ │
│  │  └─────────────────────────────────────────────────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                          证据提取流程 (步骤 1)                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  在 _check_and_prepare_evidence() 中:                                       │
│                                                                             │
│  1. 检查 item["_progress"] 是否有 stage="skill" 的条目                       │
│                                                                             │
│  2. 提取工具结果:                                                          │
│     - 从 skill stage 的 answer 字段获取工具返回结果                           │
│     - 例如: {"name": "简单对话助手", "profile": "...", ...}                  │
│                                                                             │
│  3. 调用 evidence_prepare() 提取实体:                                        │
│     输入: 工具调用结果 dict                                                   │
│     处理:                                                                   │
│       ┌─────────────────────────────────────────────────────────────────┐   │
│       │  evidence_prepare()                                              │   │
│       │  ├─► 检查是否已有 _evidence 结构                                 │   │
│       │  ├─► LLM 提取: 调用 LLM 从工具结果中提取实体                        │   │
│       │  ├─► 规则回退: 从 nodes 字段直接提取                              │   │
│       │  ├─► 去重: 按 object_type_name 去重，保留分数最高的                │   │
│       │  └─► 分配 local_id: e1, e2, e3...                               │   │
│       └─────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│     输出:                                                                  │
│     {                                                                      │
│       "evidence_id": "ev_uuid",                                           │
│       "evidences": [                                                      │
│         {                                                                │
│           "local_id": "e1",                                             │
│           "object_type_name": "简单对话助手",                             │
│           "aliases": ["对话助手"],                                        │
│           "object_type_id": "ot_agent",                                  │
│           "score": 0.95                                                  │
│         },                                                               │
│         ...                                                              │
│       ],                                                                 │
│       "summary": "简单对话助手等 3 个证据"                                │
│     }                                                                    │
│                                                                             │
│  4. 存储到 EvidenceStore:                                                   │
│     - 生成 evidence_store_key (如 "ev_abc123")                            │
│     - 调用 store.add(evidence_store_key, evidences)                       │
│     - EvidenceStore 使用 LRU 缓存 + TTL 过期                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                          证据注入流程 (步骤 2)                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  在 create_evidence_injection_stream() 中:                                   │
│                                                                             │
│  1. 检查 output 是否包含 evidence_store_key                                  │
│                                                                             │
│  2. 从 EvidenceStore 获取证据:                                               │
│     evidences = store.get(evidence_store_key)                                │
│                                                                             │
│  3. 创建 EvidenceInjectProcessor:                                            │
│     processor = EvidenceInjectProcessor(                                     │
│       evidences=evidences,                                                  │
│       enable_alias_match=True,                                             │
│       min_sentence_length=10                                               │
│     )                                                                      │
│                                                                             │
│     processor 内部构建匹配索引:                                              │
│     _match_index = {                                                       │
│       "简单对话助手": [("e1", 0)],                                         │
│       "对话助手": [("e1", 1)],  # 别名，优先级较低                          │
│       "获取agent详情": [("e2", 0)]                                         │
│     }                                                                      │
│                                                                             │
│  4. 对每个 item 进行处理:                                                    │
│                                                                             │
│     a. 提取实际文本:                                                        │
│        - 从 item["answer"] 获取 answer                                     │
│        - 如果 answer 是 dict，提取 answer["answer"]                         │
│        - 检查文本长度是否 > 20                                              │
│                                                                             │
│     b. 调用 processor._annotate_text() 标注实体:                              │
│        ┌─────────────────────────────────────────────────────────────────┐   │
│        │  _annotate_text("简单对话助手是一个基于...")                     │   │
│        │  ├─► _match_annotations(): 匹配所有实体                          │   │
│        │  │    返回: [("e1", 0, 6), ("e2", 20, 28), ...]               │   │
│        │  ├─► _resolve_overlaps(): 解决重叠冲突                           │   │
│        │  │    保留更长/更精确的匹配                                       │   │
│        │  └─► 构建 evidence_meta:                                         │   │
│        │       {                                                          │   │
│        │         "e1": {                                                │   │
│        │           "object_type_name": "简单对话助手",                     │   │
│        │           "positions": [[0, 6]]                                  │   │
│        │         },                                                      │   │
│        │         "e2": {...}                                            │   │
│        │       }                                                         │   │
│        └─────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│     c. 将 _evidence 添加到 item["_progress"]:                                │
│        - 找到最新的 progress 条目 (应该是当前 LLM stage)                    │
│        - 添加: item["_progress"][-1]["_evidence"] = evidence_meta             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                          最终响应结构                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  {                                                                          │
│    "message": {                                                             │
│      "content": {                                                           │
│        "final_answer": {                                                    │
│          "answer": {"text": "最终答案..."}                                  │
│        },                                                                   │
│        "middle_answer": {                                                   │
│          "progress": [                                                     │
│            {                                                               │
│              "stage": "llm",                                              │
│              "answer": "第一次LLM输出...",                                  │
│              "status": "completed"                                        │
│            },                                                              │
│            {                                                               │
│              "stage": "skill",                                            │
│              "skill_info": {"name": "获取agent详情"},                       │
│              "answer": {...工具结果...},                                   │
│              "status": "completed"                                        │
│            },                                                              │
│            {                                                               │
│              "stage": "llm",                                              │
│              "answer": "简单对话助手是一个基于简单对话模板创建的...",        │
│              "status": "completed",                                        │
│              "_evidence": {                                                │
│                "e1": {                                                  │
│                  "object_type_name": "简单对话助手",                       │
│                  "positions": [[0, 6]]                                   │
│                },                                                       │
│                "e2": {                                                  │
│                  "object_type_name": "获取agent详情",                      │
│                  "positions": [[20, 28]]                                 │
│                }                                                        │
│              }                                                           │
│            }                                                               │
│          ]                                                                 │
│        }                                                                   │
│      }                                                                     │
│    }                                                                       │
│  }                                                                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

## 3. 关键数据结构

### 3.1 工具调用结果 (来自 skill stage)
```json
{
  "id": "01KN704X6EQ4FTK6YVF2SM5ERG",
  "name": "简单对话助手",
  "profile": "基于简单对话模板创建",
  "config": {...},
  "status": "unpublished",
  ...
}
```

### 3.2 提取的证据 (来自 evidence_prepare)
```json
{
  "local_id": "e1",
  "object_type_name": "简单对话助手",
  "aliases": ["对话助手"],
  "object_type_id": "ot_agent",
  "score": 0.95
}
```

### 3.3 _evidence 元数据 (注入到 progress)
```json
{
  "e1": {
    "object_type_name": "简单对话助手",
    "positions": [[0, 6], [50, 56]]
  },
  "e2": {
    "object_type_name": "获取agent详情",
    "positions": [[20, 28]]
  }
}
```

## 4. 发现的问题与待确认项

### ✅ 问题 1: item 和 output 的引用关系 (已确认)

**分析**:
```python
# process_arun_loop 中
item["_progress"] = [...]
item_value = {key: get_dolphin_var_value(value) for key, value in item.items()}
output = {"answer": item_value, ...}
yield output

# create_evidence_injection_stream 中
async for item in original_stream:  # item = output
    item["_progress"][-1]["_evidence"] = {...}
```

**结论**: 如果 `_progress` 是普通列表（非 Dolphin Var），`get_dolphin_var_value` 返回原对象引用，所以修改有效。

**需要日志验证**: 确认 `item_value["_progress"] is item["_progress"]`

### ⚠️ 问题 2: _progress 修改时机 (待确认)

**流程**:
1. process_arun_loop: 过滤掉 `stage="assign"` 的 progress 条目
2. create_evidence_injection_stream: 在最新 progress 条目添加 `_evidence`

**潜在问题**: 
- 如果被过滤掉的是最后一个条目怎么办？
- 需要确保注入时 _progress 非空

### ⚠️ 问题 3: answer 字段的类型混淆

**代码逻辑**:
```python
answer = item.get("answer", {})        # item["answer"] = item_value (所有字段)
actual_text = answer.get("answer", "")  # item_value["answer"] (LLM 文本)
```

**分析**: item["answer"] 包含所有字段（query, header, _progress, answer...），
而不是单纯的 LLM 文本。

**实际文本位置**: `item["answer"]["answer"]` 是 LLM 生成的文本

**疑问**: 为什么有两层 answer？这是 Dolphin 框架的设计还是我们的处理？

### ⚠️ 问题 4: 工具结果提取时机

**当前逻辑**: 在每次迭代都检查 `_progress` 中的 `stage="skill"` 条目

**问题**: 
- 同一个工具可能被多次提取（如果它在多次迭代的 _progress 中）
- 应该只提取一次工具结果

### ⚠️ 问题 5: evidence_store_key 传递链路

**当前链路**: 
```
_check_and_prepare_evidence → 返回 evidence_store_key
process_arun_loop → output["evidence_store_key"] = evidence_store_key
create_evidence_injection_stream → 从 item.get("evidence_store_key") 获取
```

**验证点**: 确保每一步都正确传递和读取 evidence_store_key

## 5. 需要添加的验证日志

### 在 process_arun_loop 中添加:
```python
# 在创建 item_value 后
StandLogger.info_log(
    f"[process_arun_loop] item_value['_progress'] is item['_progress']: "
    f"{item_value.get('_progress') is item.get('_progress')}"
)

# 在设置 evidence_store_key 后
StandLogger.info_log(
    f"[process_arun_loop] Set evidence_store_key: {current_evidence_key}, "
    f"output['evidence_store_key'] = {output.get('evidence_store_key')}"
)
```

### 在 create_evidence_injection_stream 中添加:
```python
# 在修改 _progress 后
StandLogger.info_log(
    f"[EvidenceInject] After injection, item['_progress'] length: {len(item.get('_progress', []))}, "
    f"last entry has _evidence: {'_evidence' in item.get('_progress', [])[-1] if item.get('_progress') else 'N/A'}"
)

# 检查 answer 字段结构
StandLogger.info_log(
    f"[EvidenceInject] answer structure: keys={list(answer.keys()) if isinstance(answer, dict) else type(answer)}, "
    f"has 'answer' key: {'answer' in answer if isinstance(answer, dict) else 'N/A'}"
)
```

## 6. 边界情况测试清单

- [ ] 工具调用在对话中间发生（不是第一次迭代）
- [ ] 多个工具连续调用
- [ ] LLM 输出没有引用任何实体
- [ ] 工具结果为空或格式异常
- [ ] 同一个实体在文本中出现多次
- [ ] 实体名称是其他实体的子串
- [ ] 超过 evidence_store_max_size 的缓存淘汰
- [ ] 超过 evidence_store_ttl_seconds 的过期
- [ ] 工具结果没有 nodes 字段（规则回退）
- [ ] LLM 提取超时或失败

## 7. 性能考虑

- **LLM 提取**: 每次工具调用都调用 LLM，可能影响性能
- **缓存策略**: evidence_store_key 在整个对话中保持，避免重复提取
- **匹配性能**: 对每个 LLM 输出都进行全文匹配，O(n*m) 复杂度
- **Progress 数组**: 随着对话变长，_progress 数组会增长

## 8. evidence_prepare 调用条件

详见: [evidence_prepare_call_conditions.md](evidence_prepare_call_conditions.md)

**简要**: 只有满足以下所有条件才会调用:
1. ✅ `enable_evidence_injection = true`
2. ✅ item 包含 `_progress` 数组
3. ✅ `_progress` 中有 `stage="skill"` 条目
4. ✅ 工具结果是 dict 类型

## 9. 关键文件

| 文件 | 作用 |
|------|------|
| `run_dolphin.py` | Agent 执行入口，创建流式处理管道 |
| `interrupt_utils.py` | process_arun_loop, _check_and_prepare_evidence |
| `evidence_inject_processor.py` | create_evidence_injection_stream, EvidenceInjectProcessor |
| `evidence_prepare.py` | 从工具结果提取实体 |
| `evidence_store.py` | 证据存储 (LRU + TTL) |

## 5. 配置项

```yaml
features:
  enable_evidence_injection: true
  llm_extraction_timeout: 300
  llm_extraction_model: "Tome-pro"
  enable_alias_match: true
  min_sentence_length: 10
  evidence_store_max_size: 1000
  evidence_store_ttl_seconds: 3600
```
