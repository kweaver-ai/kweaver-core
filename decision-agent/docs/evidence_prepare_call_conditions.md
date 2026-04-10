# evidence_prepare 调用条件分析

## 完整条件链

要成功调用 `evidence_prepare`，必须满足以下**所有**条件：

### 1. 功能开关启用 ✅ 必须

```yaml
Config.features.enable_evidence_injection = true
```

**检查位置**: `_check_and_prepare_evidence()` 第 66 行

**失败日志**: `"[_check_and_prepare_evidence] END: evidence injection disabled"`

---

### 2. item 是字典类型 ✅ 必须

`item` 必须是 `dict` 类型，这是 `agent.arun()` 返回的格式。

**检查位置**: `_check_and_prepare_evidence()` 第 74 行

**失败日志**: `"[_check_and_prepare_evidence] END: item is not dict"`

---

### 3. item 包含 _progress 数组 ✅ 必须

```python
item.get("_progress") != None and len(item["_progress"]) > 0
```

**检查位置**: `_check_and_prepare_evidence()` 第 83-91 行

**失败日志**: `"[_check_and_prepare_evidence] END: no progress entries"`

---

### 4. _progress 中有 stage="skill" 的条目 ✅ 必须

```python
any(p.get("stage") == "skill" for p in item["_progress"])
```

**检查位置**: `_check_and_prepare_evidence()` 第 95-108 行

**失败日志**: `"[_check_and_prepare_evidence] END: no valid tool results"`

---

### 5. skill 条目包含有效的 tool_name 和 answer ✅ 必须

```python
skill_info.get("name") != ""  # 工具名不为空
progress_item.get("answer") is not None  # 有返回结果
```

**检查位置**: `_check_and_prepare_evidence()` 第 99-104 行

**失败日志**: 不会单独失败，只是该工具被跳过

---

### 6. 工具结果是 dict 类型 ✅ 必须

```python
isinstance(result, dict) == True
```

**检查位置**: `_check_and_prepare_evidence()` 第 147 行

**失败日志**: `"[_check_and_prepare_evidence] SKIP: result is not dict, type=xxx"`

---

## 调用流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                    每次_agent.arun()迭代                         │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │  item 是 dict?          │ ───→ NO ──► 返回
                └───────────────────────┘
                            │ YES
                            ▼
                ┌───────────────────────┐
                │  有 _progress?          │ ───→ NO ──► 返回
                └───────────────────────┘
                            │ YES
                            ▼
                ┌───────────────────────┐
                │  有 stage="skill"?      │ ───→ NO ──► 返回
                └───────────────────────┘
                            │ YES
                            ▼
            ┌───────────────────────────────────────┐
            │  遍历 skill 条目，提取工具结果           │
            │  tool_name != ""                      │
            │  answer is not None                    │
            └───────────────────────────────────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │  有工具结果?            │ ───→ NO ──► 返回
                └───────────────────────┘
                            │ YES
                            ▼
            ┌───────────────────────────────────────┐
            │  对每个工具结果:                         │
            │  ┌─────────────────────────────────┐  │
            │  │  result 是 dict?                  │  │
            │  └─────────────────────────────────┘  │
            │             │ NO                     │
            │             ▼                        │
            │         返回 (跳过该工具)              │
            │             │ YES                    │
            │             ▼                        │
            │  ┌─────────────────────────────────┐  │
            │  │  await evidence_prepare()        │  │
            │  └─────────────────────────────────┘  │
            └───────────────────────────────────────┘
                            │
                            ▼
                    存储到 EvidenceStore
```

## 实际执行示例

### 场景 1: 正常工具调用（✅ 会调用）

```
对话开始
  ↓
迭代 1: item["_progress"] = [{"stage": "llm", "answer": "让我查询一下"}]
  → 没有 skill stage → 不调用 evidence_prepare
  
  ↓
迭代 2: item["_progress"] = [
    {"stage": "llm", "answer": "让我查询一下"},
    {"stage": "skill", "skill_info": {"name": "获取agent详情"}, "answer": {...}}
  ]
  → 有 skill stage → 提取工具结果 → result 是 dict
  → ✅ 调用 evidence_prepare(获取agent详情的结果)
```

### 场景 2: 没有工具调用（❌ 不会调用）

```
迭代 1: item["_progress"] = [{"stage": "llm", "answer": "直接回答"}]
  → 没有 skill stage → 不调用
```

### 场景 3: 工具结果不是 dict（❌ 不会调用）

```
迭代 2: item["_progress"] = [
    {"stage": "skill", "skill_info": {"name": "工具名"}, "answer": "字符串结果"}
  ]
  → skill stage 存在 → 但 answer 不是 dict
  → 跳过，不调用 evidence_prepare
```

## 常见不调用的情况

| 情况 | 原因 | 日志标识 |
|------|------|----------|
| 功能未开启 | `enable_evidence_injection = false` | "evidence injection disabled" |
| 纯 LLM 对话 | 没有调用任何工具 | "no valid tool results" |
| 工具返回格式错误 | answer 不是 dict 类型 | "SKIP: result is not dict" |
| 第一次迭代 | LLM 还在生成问题，还没调用工具 | "no valid tool results" |

## 验证证据提取是否执行的日志

在日志中搜索以下关键标识：

```bash
# 确认功能已启用
grep "enable_evidence_injection" log

# 确认找到工具结果
grep "Tool call results extracted" log

# 确认进入 try 块
grep "=== BEFORE TRY BLOCK ===" log

# 确认调用函数
grep "About to call evidence_prepare" log

# 确认提取成功
grep "evidence_prepare returned" log
```
