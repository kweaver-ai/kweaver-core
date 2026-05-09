# 🔁 Dataflow

## 📖 Overview

**Dataflow** orchestrates **pipelines**, **scheduled jobs**, **automation**, and **code runners** across the AI Data Platform. It connects ingestion, transformation, and hand-off to agents or downstream systems.

Typical ingress prefixes (see Helm values in `adp/dataflow`):

| Prefix | Role |
| --- | --- |
| `/api/automation/v1` | Automation and scheduling |
| `/api/flow-stream-data-pipeline/v1` | Stream / pipeline control |
| `/api/coderunner` | Sandboxed or managed code execution hooks |

**Related modules:** [VEGA Engine](vega.md), [Execution Factory](execution-factory.md), [BKN Engine](bkn.md).

## CLI

### Listing Dataflows

```bash
# List all registered dataflows (DAGs)
kweaver dataflow list

# Verbose output with run status and schedule info
kweaver dataflow list -v
```

### Triggering a Run

```bash
# Trigger a run by DAG ID with a local file as input
kweaver dataflow run <dag_id> --file ./data/report.pdf

# Trigger a run with a URL-based input and a custom run name
kweaver dataflow run <dag_id> --url https://storage.example.com/data.csv --name "daily-import-2026-04-14"

# Trigger with both file and parameters
kweaver dataflow run <dag_id> --file ./data/input.json --name "batch-run"
```

### Viewing Run History

```bash
# List all runs for a dataflow
kweaver dataflow runs <dag_id>

# Filter runs since a date (supports natural date parsing)
kweaver dataflow runs <dag_id> --since 2026-04-01
kweaver dataflow runs <dag_id> --since "last week"
kweaver dataflow runs <dag_id> --since "3 days ago"

# Limit results
kweaver dataflow runs <dag_id> --since 2026-04-01 --limit 10
```

### Viewing Run Logs

```bash
# Get logs for a specific run instance
kweaver dataflow logs <dag_id> <instance_id>

# Detailed logs (include input/output for each step)
kweaver dataflow logs <dag_id> <instance_id> --detail

# Show only failed steps
kweaver dataflow logs <dag_id> <instance_id> --status failed
```

### End-to-End Example

```bash
# 1. List available dataflows
kweaver dataflow list

# 2. Trigger a document processing pipeline with a local PDF
kweaver dataflow run dag-doc-ingest --file ./contracts/new-agreement.pdf --name "contract-import"
# → instance_id: run-abc123

# 3. Monitor the run
kweaver dataflow runs dag-doc-ingest --since today

# 4. View detailed logs once complete
kweaver dataflow logs dag-doc-ingest run-abc123 --detail

# 5. Verify data landed in VEGA/BKN
kweaver dataview find --name "contracts"
kweaver bkn search kn-legal "new agreement"
```

---

## Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")

# List dataflows
flows = client.dataflow.list_flows()
for f in flows:
    print(f["dag_id"], f["name"], f["schedule"], f["status"])

# Trigger a run with a file
run = client.dataflow.trigger(
    dag_id="dag-doc-ingest",
    file_path="./contracts/new-agreement.pdf",
    name="contract-import",
)
print("instance_id:", run["instance_id"])

# Trigger a run with a URL
run = client.dataflow.trigger(
    dag_id="dag-csv-loader",
    url="https://storage.example.com/data.csv",
    name="daily-import",
)

# List runs with date filter
runs = client.dataflow.list_runs(
    dag_id="dag-doc-ingest",
    since="2026-04-01",
    limit=20,
)
for r in runs:
    print(r["instance_id"], r["status"], r["started_at"], r["duration_ms"])

# Get logs for a run
logs = client.dataflow.get_logs(
    dag_id="dag-doc-ingest",
    instance_id="run-abc123",
    detail=True,
)
for step in logs["steps"]:
    print(step["name"], step["status"], step["duration_ms"])
    if step.get("error"):
        print("  error:", step["error"])

# Get a single run's status
status = client.dataflow.get_run("dag-doc-ingest", "run-abc123")
print(status["status"], status["progress"])
```

---

## TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });

// List dataflows
const flows = await client.dataflow.listFlows();
flows.forEach((f) => console.log(f.dagId, f.name, f.schedule, f.status));

// Trigger a run with a file
const run = await client.dataflow.trigger({
  dagId: 'dag-doc-ingest',
  filePath: './contracts/new-agreement.pdf',
  name: 'contract-import',
});
console.log('instance_id:', run.instanceId);

// Trigger a run with a URL
const urlRun = await client.dataflow.trigger({
  dagId: 'dag-csv-loader',
  url: 'https://storage.example.com/data.csv',
  name: 'daily-import',
});

// List runs since a date
const runs = await client.dataflow.listRuns({
  dagId: 'dag-doc-ingest',
  since: '2026-04-01',
  limit: 20,
});
runs.forEach((r) =>
  console.log(r.instanceId, r.status, r.startedAt, r.durationMs),
);

// Get detailed logs
const logs = await client.dataflow.getLogs({
  dagId: 'dag-doc-ingest',
  instanceId: 'run-abc123',
  detail: true,
});
logs.steps.forEach((step) => {
  console.log(step.name, step.status, step.durationMs);
  if (step.error) console.log('  error:', step.error);
});
```

---

## curl

```bash
# List all dataflows
curl -sk "https://<access-address>/api/automation/v1/flows" \
  -H "Authorization: Bearer $(kweaver token)"

# Trigger a run (JSON input)
curl -sk -X POST "https://<access-address>/api/automation/v1/flows/dag-doc-ingest/runs" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "contract-import",
    "url": "https://storage.example.com/contracts/new-agreement.pdf"
  }'

# Trigger a run with file upload (multipart)
curl -sk -X POST "https://<access-address>/api/automation/v1/flows/dag-doc-ingest/runs" \
  -H "Authorization: Bearer $(kweaver token)" \
  -F "name=contract-import" \
  -F "file=@./contracts/new-agreement.pdf"

# List runs for a dataflow (with since filter)
curl -sk "https://<access-address>/api/automation/v1/flows/dag-doc-ingest/runs?since=2026-04-01&limit=20" \
  -H "Authorization: Bearer $(kweaver token)"

# Get a specific run's status
curl -sk "https://<access-address>/api/automation/v1/flows/dag-doc-ingest/runs/run-abc123" \
  -H "Authorization: Bearer $(kweaver token)"

# Get detailed logs for a run
curl -sk "https://<access-address>/api/automation/v1/flows/dag-doc-ingest/runs/run-abc123/logs?detail=true" \
  -H "Authorization: Bearer $(kweaver token)"

# Get stream/pipeline status
curl -sk "https://<access-address>/api/flow-stream-data-pipeline/v1/pipelines" \
  -H "Authorization: Bearer $(kweaver token)"

# Health check
curl -sk "https://<access-address>/api/automation/v1/health" \
  -H "Authorization: Bearer $(kweaver token)"
```
