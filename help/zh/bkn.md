# BKN 引擎

## 概述

**业务知识网络（BKN）** 是 KWeaver Core 的语义层，用**对象类型**、**关系类型**、**动作类型**描述领域，并存储**实例**与**关系**，为智能体与分析提供统一本体。

典型服务（Ingress 前缀随版本可能调整）：

| 前缀 | 作用 |
| --- | --- |
| `/api/bkn-backend/v1` | BKN 后台 API（管理与运行时对接） |
| `/api/ontology-manager/v1` | 本体 / Schema 管理 |
| `/api/ontology-query/v1` | 面向图与本体的查询 |

**相关模块：** [VEGA 引擎](vega.md)（视图背后的数据）、[Context Loader](context-loader.md)（基于本体的上下文）、[Decision Agent](decision-agent.md)（运行时消费 BKN）。

## 使用方式

HTTP 示例中请设置基址与令牌：

```bash
export KWEAVER_BASE="https://<访问地址>"
export TOKEN="<bearer-token>"   # 最小化安装关闭认证时可省略或调整
```

具体路径与请求体以集群 OpenAPI 或 `kweaver` CLI 为准。

### CLI

```bash
kweaver bkn list
kweaver bkn --help
# 从文件推送或同步 BKN 定义（见 kweaver-sdk / create-bkn skill）
# kweaver bkn push ./my-network.bkn
```

### Python SDK

从 [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk) 安装（包名与 API 以发布版本为准）。

```python
# 示例 — 请按实际 SDK 版本替换客户端 API
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")
# client.auth.login(...)
networks = client.bkn.list_networks()
print(networks)
```

### TypeScript SDK

```typescript
// 示例 — 请按实际 SDK 版本替换客户端 API
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });
// await client.auth.login(...);
const networks = await client.bkn.listNetworks();
console.log(networks);
```

### curl

```bash
# 示例：探测 ontology-manager（路径请对照 OpenAPI）
curl -sk "$KWEAVER_BASE/api/ontology-manager/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# 示例：ontology-query（替换为文档中的真实查询接口）
curl -sk -X POST "$KWEAVER_BASE/api/ontology-query/v1/query" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```
