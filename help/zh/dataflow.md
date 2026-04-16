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

### CLI

#### 列出流程

```bash
# 列出所有已注册的 DAG 流程
kweaver dataflow list
```

输出包含每个 DAG 的 ID、名称、调度状态与最近运行时间。

#### 触发运行

```bash
# 使用本地配置文件触发运行
kweaver dataflow run dag_etl_daily --file ./params/etl-config.json

# 使用远程 URL 上的配置文件触发运行
kweaver dataflow run dag_etl_daily \
  --url https://storage.example.com/configs/etl-prod.json \
  --name etl-prod.json
```

`--file` 接受本地 JSON 文件路径，文件内容作为运行参数传入 DAG。`--url` 与 `--name` 用于从远程下载配置文件作为运行输入，`--name` 指定保存的文件名。

#### 查询运行历史

```bash
# 查询最近所有运行记录（自动分页获取全部）
kweaver dataflow runs dag_etl_daily

# 只查看某一天之后的运行记录
kweaver dataflow runs dag_etl_daily --since 2025-01-15
```

`--since` 接受 `YYYY-MM-DD` 格式的日期，仅返回该日期之后触发的运行记录。CLI 自动翻页获取所有匹配记录，无需手动处理分页参数。

输出示例：

```
ID                    状态       开始时间                  耗时
run_20250115_001      success   2025-01-15T08:00:00Z      12m34s
run_20250115_002      failed    2025-01-15T12:00:00Z      3m21s
run_20250116_001      running   2025-01-16T08:00:00Z      --
```

#### 查看运行日志

```bash
# 查看运行摘要（各步骤的状态与耗时总览）
kweaver dataflow logs dag_etl_daily run_20250115_001

# 查看详细日志（每个步骤的完整标准输出与错误输出）
kweaver dataflow logs dag_etl_daily run_20250115_001 --detail
```

**默认摘要模式**输出每个步骤的名称、状态（success / failed / running / skipped）和耗时：

```
步骤                   状态       耗时
extract_orders         success   2m10s
transform_normalize    success   5m44s
load_to_warehouse      success   4m40s
notify_downstream      success   0m12s
```

**`--detail` 模式**在摘要之后追加每个步骤的完整日志输出，包含标准输出、标准错误及退出码，用于排查失败原因。

#### 端到端流程

```bash
# 1. 查看可用流程
kweaver dataflow list

# 2. 触发 ETL 流程
kweaver dataflow run dag_etl_daily --file ./params/etl-config.json

# 3. 查看运行记录
kweaver dataflow runs dag_etl_daily --since 2025-01-15

# 4. 检查失败运行的详细日志
kweaver dataflow logs dag_etl_daily run_20250115_002 --detail
```

---

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")

flows = client.dataflow.list_flows()
for flow in flows["data"]:
    print(flow["dag_id"], flow["name"], flow["schedule"], flow["status"])

run = client.dataflow.trigger(
    dag_id="dag_etl_daily",
    params={"date": "2025-01-15", "mode": "full"}
)
print(f"运行 ID: {run['run_id']}, 状态: {run['status']}")

runs = client.dataflow.list_runs("dag_etl_daily", since="2025-01-15")
for r in runs["data"]:
    print(r["run_id"], r["status"], r["start_time"], r["duration"])

logs = client.dataflow.get_logs("dag_etl_daily", "run_20250115_001", detail=True)
for step in logs["steps"]:
    print(f"[{step['status']}] {step['name']} ({step['duration']})")
    if step["status"] == "failed":
        print(f"  错误: {step['stderr']}")

run_with_url = client.dataflow.trigger(
    dag_id="dag_etl_daily",
    file_url="https://storage.example.com/configs/etl-prod.json",
    file_name="etl-prod.json"
)
print(f"运行 ID: {run_with_url['run_id']}")
```

---

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });

const flows = await client.dataflow.listFlows();
flows.data.forEach((flow) =>
  console.log(flow.dagId, flow.name, flow.schedule, flow.status),
);

const run = await client.dataflow.trigger({
  dagId: 'dag_etl_daily',
  params: { date: '2025-01-15', mode: 'full' },
});
console.log('运行 ID:', run.runId, '状态:', run.status);

const runs = await client.dataflow.listRuns('dag_etl_daily', {
  since: '2025-01-15',
});
runs.data.forEach((r) =>
  console.log(r.runId, r.status, r.startTime, r.duration),
);

const logs = await client.dataflow.getLogs('dag_etl_daily', 'run_20250115_001', {
  detail: true,
});
logs.steps.forEach((step) => {
  console.log(`[${step.status}] ${step.name} (${step.duration})`);
  if (step.status === 'failed') {
    console.log('  错误:', step.stderr);
  }
});

const runWithUrl = await client.dataflow.trigger({
  dagId: 'dag_etl_daily',
  fileUrl: 'https://storage.example.com/configs/etl-prod.json',
  fileName: 'etl-prod.json',
});
console.log('运行 ID:', runWithUrl.runId);
```

---

### curl

```bash
# 列出所有 DAG 流程
curl -sk "https://<访问地址>/api/flow-stream-data-pipeline/v1/flows" \
  -H "Authorization: Bearer $(kweaver token)"

# 触发运行
curl -sk -X POST "https://<访问地址>/api/flow-stream-data-pipeline/v1/flows/dag_etl_daily/runs" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "params": {"date": "2025-01-15", "mode": "full"}
  }'

# 使用远程配置文件触发运行
curl -sk -X POST "https://<访问地址>/api/flow-stream-data-pipeline/v1/flows/dag_etl_daily/runs" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "file_url": "https://storage.example.com/configs/etl-prod.json",
    "file_name": "etl-prod.json"
  }'

# 查询运行历史
curl -sk "https://<访问地址>/api/flow-stream-data-pipeline/v1/flows/dag_etl_daily/runs?since=2025-01-15&limit=50" \
  -H "Authorization: Bearer $(kweaver token)"

# 获取运行日志（摘要）
curl -sk "https://<访问地址>/api/flow-stream-data-pipeline/v1/flows/dag_etl_daily/runs/run_20250115_001/logs" \
  -H "Authorization: Bearer $(kweaver token)"

# 获取运行日志（详细）
curl -sk "https://<访问地址>/api/flow-stream-data-pipeline/v1/flows/dag_etl_daily/runs/run_20250115_001/logs?detail=true" \
  -H "Authorization: Bearer $(kweaver token)"

# 代码执行器调用
curl -sk -X POST "https://<访问地址>/api/coderunner/execute" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "language": "python",
    "code": "import pandas as pd\ndf = pd.read_csv('\\''data.csv'\\'')\nprint(df.describe())",
    "timeout": 30
  }'
```
