# VEGA 引擎

## 概述

**VEGA** 提供跨异构数据源的**数据虚拟化**：**数据连接**、**模型**与**视图**（含原子视图与组合视图）。智能体与应用通过统一的类 SQL 访问面查询，而无需为每个数据源单独适配。

典型 Ingress 前缀：

| 前缀 | 作用 |
| --- | --- |
| `/api/vega-backend/v1` | VEGA 后台 — 连接、元数据、查询执行 |

**相关模块：** [BKN 引擎](bkn.md)、[Context Loader](context-loader.md)、[Dataflow](dataflow.md)（数据落地与转换流程）。

## 使用方式

```bash
export KWEAVER_BASE="https://<访问地址>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver vega --help
# 列出连接、目录或执行即席查询（取决于 CLI 版本）
# kweaver vega connection list
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")
# 示例 — 方法名以 SDK 发布为准
# connections = client.vega.list_connections()
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });
// const connections = await client.vega.listConnections();
```

### curl

```bash
curl -sk "$KWEAVER_BASE/api/vega-backend/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# 示例：对已注册视图执行 SQL（请求体为示意）
curl -sk -X POST "$KWEAVER_BASE/api/vega-backend/v1/query" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"sql":"SELECT 1"}'
```
