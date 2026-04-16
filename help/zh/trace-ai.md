# Trace AI

## 概述

**Trace AI** 提供**全链路可观测**：接收 OTLP 链路、导出到检索后端，并通过 **agent-observability** 服务查询与智能体及平台活动关联的 Span。每一次智能体对话都会生成完整的执行链路，记录从用户输入到最终回复的每一步决策与操作。

典型 Ingress 前缀：

| 前缀 | 作用 |
| --- | --- |
| `/api/agent-observability/v1` | 链路查询与可观测 API |

**相关模块：** [Decision Agent](decision-agent.md)，以及平台日志、指标管道（如 feed-ingester）。

## Trace 数据结构

一次智能体对话产生的 Trace 由多个 **Span** 组成，形成父子树形结构：

```
conversation (根 Span)
├── planning                     → LLM 规划阶段
│   ├── context_retrieval        → 上下文检索
│   │   ├── kn_search            → 知识网络搜索
│   │   └── instance_query       → 实例查询
│   └── llm_call                 → 模型调用
├── tool_execution               → 工具执行阶段
│   ├── tool:web_search          → Web 搜索工具
│   └── tool:code_runner         → 代码执行工具
├── synthesis                    → 结果合成
│   └── llm_call                 → 模型调用（生成回复）
└── response                     → 最终响应
```

### Span 关键字段

| 字段 | 说明 |
| --- | --- |
| `span_id` | Span 唯一标识 |
| `parent_span_id` | 父 Span ID（根 Span 为空） |
| `operation` | 操作名称（如 `planning`、`tool:web_search`） |
| `start_time` / `end_time` | 起止时间戳（ISO 8601） |
| `duration_ms` | 耗时（毫秒） |
| `status` | 状态：`ok`、`error`、`timeout` |
| `attributes` | 键值属性（如 `llm.model`、`tool.name`、`query`） |
| `events` | 事件列表（如 `token_usage`、`error`） |

### 证据链分析

Trace 数据支持以下分析场景：

- **性能瓶颈定位**：通过 `duration_ms` 找到耗时最长的 Span
- **错误溯源**：从 `status: error` 的 Span 追溯到根因
- **Token 用量审计**：通过 `events` 中的 `token_usage` 事件统计模型调用成本
- **决策路径回放**：按父子关系重建智能体的完整思考与执行路径
- **工具调用审计**：检查每个工具调用的输入参数与返回结果

### CLI

`kweaver agent trace` 需要两个参数：**智能体 ID**、**会话 ID**（与 `kweaver agent sessions <agent_id>` 或对话返回一致）。

#### 查看对话链路

```bash
# 格式化输出 — 以树形结构展示 Span 层级、耗时与状态
kweaver agent trace agt_001 conv_20250115_001 --pretty
```

`--pretty` 输出示例：

```
Trace: conv_20250115_001
总耗时: 3,245ms | Span 数: 12 | 状态: ok

conversation (3,245ms) ✓
├── planning (1,820ms) ✓
│   ├── context_retrieval (1,200ms) ✓
│   │   ├── kn_search (450ms) ✓
│   │   │   query: "近期高价值客户"
│   │   │   results: 8
│   │   └── instance_query (720ms) ✓
│   │       object_type: ot_customer
│   │       conditions: status == active
│   │       results: 23
│   └── llm_call (580ms) ✓
│       model: gpt-4o
│       tokens: 1,240 (prompt) + 320 (completion)
├── tool_execution (890ms) ✓
│   └── tool:code_runner (890ms) ✓
│       code: "df.groupby('region').sum()"
│       exit_code: 0
├── synthesis (480ms) ✓
│   └── llm_call (480ms) ✓
│       model: gpt-4o
│       tokens: 2,100 (prompt) + 560 (completion)
└── response (55ms) ✓
    content_length: 1,240 chars
```

```bash
# 紧凑输出 — 适合管道处理与日志聚合
kweaver agent trace agt_001 conv_20250115_001 --compact
```

`--compact` 输出为单行 JSON 数组，每个元素一个 Span：

```json
[
  {"span_id":"sp_001","parent":"","op":"conversation","duration_ms":3245,"status":"ok"},
  {"span_id":"sp_002","parent":"sp_001","op":"planning","duration_ms":1820,"status":"ok"},
  {"span_id":"sp_003","parent":"sp_002","op":"context_retrieval","duration_ms":1200,"status":"ok"}
]
```

#### 解读 Trace 输出

**时序分析**：比较同层级 Span 的 `duration_ms`，定位瓶颈。上例中 `context_retrieval` 占 `planning` 的 66%，可考虑优化知识网络索引。

**父子关系**：`parent_span_id` 为空表示根 Span。每个子 Span 的时间区间应在父 Span 内。多个子 Span 可能并行执行（时间区间重叠）或顺序执行。

**错误追踪**：当某个 Span 的 `status` 为 `error` 时，检查其 `events` 列表中的 `error` 事件，包含 `message` 与 `stack_trace`。从失败的叶子节点向根回溯，可还原完整的错误传播路径。

---

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient()  # 需先 kweaver auth login；构造方式以 kweaver-sdk 为准

trace = client.agent.trace("agt_001", "conv_20250115_001")
print(f"Trace ID: {trace['trace_id']}")
print(f"总耗时: {trace['duration_ms']}ms")
print(f"Span 数: {len(trace['spans'])}")
print(f"状态: {trace['status']}")

def print_tree(spans, parent_id="", depth=0):
    children = [s for s in spans if s.get("parent_span_id", "") == parent_id]
    for span in children:
        status_icon = "✓" if span["status"] == "ok" else "✗"
        indent = "  " * depth
        print(f"{indent}{span['operation']} ({span['duration_ms']}ms) {status_icon}")
        for key, val in span.get("attributes", {}).items():
            print(f"{indent}  {key}: {val}")
        print_tree(spans, span["span_id"], depth + 1)

print_tree(trace["spans"])

error_spans = [s for s in trace["spans"] if s["status"] == "error"]
for span in error_spans:
    print(f"错误 Span: {span['operation']}")
    for event in span.get("events", []):
        if event["name"] == "error":
            print(f"  消息: {event['attributes']['message']}")
            print(f"  堆栈: {event['attributes'].get('stack_trace', 'N/A')}")

llm_spans = [s for s in trace["spans"] if s["operation"] == "llm_call"]
total_prompt = sum(s["attributes"].get("prompt_tokens", 0) for s in llm_spans)
total_completion = sum(s["attributes"].get("completion_tokens", 0) for s in llm_spans)
print(f"Token 用量 — Prompt: {total_prompt}, Completion: {total_completion}, 总计: {total_prompt + total_completion}")

sorted_spans = sorted(trace["spans"], key=lambda s: s["duration_ms"], reverse=True)
print("耗时 Top 5:")
for span in sorted_spans[:5]:
    print(f"  {span['operation']}: {span['duration_ms']}ms")
```

---

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = await KWeaverClient.connect();

const trace = await client.agent.trace('agt_001', 'conv_20250115_001');
console.log(`Trace ID: ${trace.traceId}`);
console.log(`总耗时: ${trace.durationMs}ms`);
console.log(`Span 数: ${trace.spans.length}`);
console.log(`状态: ${trace.status}`);

function printTree(spans: typeof trace.spans, parentId = '', depth = 0) {
  const children = spans.filter((s) => (s.parentSpanId ?? '') === parentId);
  for (const span of children) {
    const icon = span.status === 'ok' ? '✓' : '✗';
    const indent = '  '.repeat(depth);
    console.log(`${indent}${span.operation} (${span.durationMs}ms) ${icon}`);
    for (const [key, val] of Object.entries(span.attributes ?? {})) {
      console.log(`${indent}  ${key}: ${val}`);
    }
    printTree(spans, span.spanId, depth + 1);
  }
}
printTree(trace.spans);

const errorSpans = trace.spans.filter((s) => s.status === 'error');
for (const span of errorSpans) {
  console.log(`错误 Span: ${span.operation}`);
  for (const event of span.events ?? []) {
    if (event.name === 'error') {
      console.log(`  消息: ${event.attributes.message}`);
    }
  }
}

const llmSpans = trace.spans.filter((s) => s.operation === 'llm_call');
const totalPrompt = llmSpans.reduce((sum, s) => sum + (s.attributes?.promptTokens ?? 0), 0);
const totalCompletion = llmSpans.reduce((sum, s) => sum + (s.attributes?.completionTokens ?? 0), 0);
console.log(`Token 用量 — Prompt: ${totalPrompt}, Completion: ${totalCompletion}`);

const sorted = [...trace.spans].sort((a, b) => b.durationMs - a.durationMs);
console.log('耗时 Top 5:');
sorted.slice(0, 5).forEach((s) => console.log(`  ${s.operation}: ${s.durationMs}ms`));
```

---

### curl

```bash
# 获取对话 Trace
curl -sk "https://<访问地址>/api/agent-observability/v1/traces/conv_20250115_001" \
  -H "Authorization: Bearer $(kweaver token)"

# 按条件搜索 Trace
curl -sk -X POST "https://<访问地址>/api/agent-observability/v1/traces/search" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agt_001",
    "start_time": "2025-01-15T00:00:00Z",
    "end_time": "2025-01-16T00:00:00Z",
    "status": "error",
    "limit": 20
  }'

# 获取特定 Span 详情
curl -sk "https://<访问地址>/api/agent-observability/v1/spans/sp_003" \
  -H "Authorization: Bearer $(kweaver token)"

# 获取 Trace 统计摘要（耗时分布、错误率等）
curl -sk "https://<访问地址>/api/agent-observability/v1/traces/conv_20250115_001/summary" \
  -H "Authorization: Bearer $(kweaver token)"
```
