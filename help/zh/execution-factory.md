# Execution Factory

## 概述

**Execution Factory** 负责注册与执行智能体可调用的**算子**、**工具**与**技能**，在策略控制下将 LLM 规划落地为具体业务动作（HTTP 工具、代码执行、MCP 等）。

典型 Ingress 前缀：

| 前缀 | 作用 |
| --- | --- |
| `/api/agent-operator-integration/v1` | 算子集成与工具执行面 |

**相关模块：** [Decision Agent](decision-agent.md)、[Dataflow](dataflow.md)（自动化与 code-runner 路径）。

## 使用方式

```bash
export KWEAVER_BASE="https://<访问地址>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver execution --help
# 或按 CLI 版本使用 tool / operator 子命令
# kweaver tool list
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")
# tools = client.execution_factory.list_tools()
# result = client.execution_factory.invoke(tool_id="...", input={})
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });
// const tools = await client.executionFactory.listTools();
```

### curl

```bash
curl -sk "$KWEAVER_BASE/api/agent-operator-integration/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# 示例：列出或调用算子 — 与 OpenAPI 对齐
curl -sk -X GET "$KWEAVER_BASE/api/agent-operator-integration/v1/operators" \
  -H "Authorization: Bearer $TOKEN"
```
