# Dataflow

## 概述

**Dataflow** 编排 AI 数据平台上的**流水线**、**定时任务**、**自动化**与**代码执行器**，连接采集、转换以及向智能体或下游系统的交付。

典型 Ingress 前缀（见 `adp/dataflow` 下 Helm 配置）：

| 前缀 | 作用 |
| --- | --- |
| `/api/automation/v1` | 自动化与调度 |
| `/api/flow-stream-data-pipeline/v1` | 流式 / 流水线控制 |
| `/api/coderunner` | 沙箱或托管代码执行入口 |

**相关模块：** [VEGA 引擎](vega.md)、[Execution Factory](execution-factory.md)、[BKN 引擎](bkn.md)。

## 使用方式

```bash
export KWEAVER_BASE="https://<访问地址>"
export TOKEN="<bearer-token>"
```

### CLI

```bash
kweaver dataflow --help
# 列出流程、触发运行或查看历史（取决于 CLI 版本）
# kweaver dataflow list
# kweaver dataflow run <flow-id>
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")
# flows = client.dataflow.list_flows()
# run = client.dataflow.trigger(flow_id="...", inputs={})
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });
// const flows = await client.dataflow.listFlows();
```

### curl

```bash
curl -sk "$KWEAVER_BASE/api/automation/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# 示例：流水线或流程 API
curl -sk -X GET "$KWEAVER_BASE/api/flow-stream-data-pipeline/v1/flows" \
  -H "Authorization: Bearer $TOKEN"
```
