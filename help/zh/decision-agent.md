# Decision Agent

## 概述

**Decision Agent** 是面向目标的智能体：规划、检索上下文、在策略下调用工具并基于反馈迭代。核心服务包括 **agent-factory**（编排 API）、**agent-executor**、**记忆**与检索集成等。

典型 Ingress 前缀：

| 前缀 | 作用 |
| --- | --- |
| `/api/agent-factory/v1` | 工厂 API（模板、智能体、运行） |
| `/api/agent-factory/v2`、`/api/agent-factory/v3` | 版本化工厂接口 |
| `/api/agent-app/v1` | 面向应用的智能体 API |

**相关模块：** [Context Loader](context-loader.md)、[Execution Factory](execution-factory.md)、[BKN 引擎](bkn.md)、[Trace AI](trace-ai.md)。

## 使用方式

```bash
export KWEAVER_BASE="https://<访问地址>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver agent --help
# 列出智能体、从模板创建、对话或执行任务（取决于 CLI 版本）
# kweaver agent list
# kweaver agent chat <agent-id> --message "你好"
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")
# agents = client.agent.list()
# reply = client.agent.chat(agent_id="...", message="下一步怎么做")
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });
// const agents = await client.agent.list();
```

### curl

```bash
curl -sk "$KWEAVER_BASE/api/agent-factory/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# 示例：列出智能体或模板
curl -sk -X GET "$KWEAVER_BASE/api/agent-factory/v1/agents" \
  -H "Authorization: Bearer $TOKEN"
```
