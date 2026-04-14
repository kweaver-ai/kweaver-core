# Context Loader

## 概述

**Context Loader**（含 **agent-retrieval** 等服务）为 Decision Agent 组装**高质量上下文**：面向本体的召回、排序，以及从 BKN 与数据面的按需加载。它位于原始数据/VEGA 与智能体运行时之间。

典型 Ingress 前缀：

| 前缀 | 作用 |
| --- | --- |
| `/api/agent-retrieval/v1` | 检索与上下文组装 API |

**相关模块：** [BKN 引擎](bkn.md)、[VEGA 引擎](vega.md)、[Decision Agent](decision-agent.md)。

## 使用方式

```bash
export KWEAVER_BASE="https://<访问地址>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver --help
# 新版本 CLI 可能在独立分组下提供 context / retrieval 子命令
# kweaver context --help
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")
# 示例：为任务 / 对话轮次构建上下文
# ctx = client.context_loader.build(
#     agent_id="...",
#     query="用户问题",
#     bkn_id="...",
# )
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });
// const ctx = await client.contextLoader.build({ ... });
```

### curl

```bash
curl -sk "$KWEAVER_BASE/api/agent-retrieval/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# 示例检索请求 — 请求体请按 OpenAPI 填写
curl -sk -X POST "$KWEAVER_BASE/api/agent-retrieval/v1/retrieve" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query":"示例","limit":10}'
```
