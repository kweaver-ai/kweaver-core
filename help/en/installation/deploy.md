# Deploy

KWeaver Core is installed with the `deploy.sh` script under the `deploy/` directory in this repository.

## Clone and enter deploy directory

```bash
git clone https://github.com/kweaver-ai/kweaver-core.git
cd kweaver-core/deploy
chmod +x deploy.sh
```

## Install KWeaver Core

### Minimum install (recommended for first try)

Skips some optional modules (e.g. auth / business domain) for a lighter footprint:

```bash
./deploy.sh kweaver-core install --minimum
```

Equivalent flags:

```bash
./deploy.sh kweaver-core install --set auth.enabled=false --set businessDomain.enabled=false
```

### Full install

Includes auth and business-domain related components:

```bash
./deploy.sh kweaver-core install
```

The script may prompt for **access address** and detect **API server address** automatically.

### Non-interactive install

```bash
./deploy.sh kweaver-core install \
  --access_address=<your-ip-or-domain> \
  --api_server_address=<nic-ip-for-k8s-api>
```

- `--access_address` — URL or IP clients use to reach KWeaver (ingress)
- `--api_server_address` — real NIC IP bound for the Kubernetes API server

### Custom ingress ports (optional)

```bash
export INGRESS_NGINX_HTTP_PORT=8080
export INGRESS_NGINX_HTTPS_PORT=8443
./deploy.sh kweaver-core install
```

## Useful commands

```bash
./deploy.sh kweaver-core status
./deploy.sh kweaver-core uninstall
./deploy.sh --help
```

## What gets installed

1. Single-node Kubernetes (if needed), storage, ingress
2. Data services: MariaDB, Redis, Kafka, ZooKeeper, OpenSearch (as defined by release manifests)
3. KWeaver Core application Helm charts

For uninstall and cluster reset, see [deploy/README.md](../../../deploy/README.md).

## After install (check cluster and API)

When `deploy.sh kweaver-core install` finishes, confirm the cluster and that you can reach the platform.

### Kubernetes

```bash
kubectl get nodes
kubectl get pods -A
```

Wait until core namespaces show `Running` / `Ready` for critical workloads.

### Deploy script status

```bash
cd kweaver-core/deploy
./deploy.sh kweaver-core status
```

### CLI

Install the CLI from [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk):

```bash
npm install -g @kweaver-ai/kweaver-sdk
# or run without global install: npx kweaver --help
```

Then log in and verify:

```bash
kweaver auth login https://<access-address> -k
kweaver bkn list
```

Use the same host as `--access_address` or the node IP from the installer.

### HTTP (optional)

```bash
curl -sk "https://<access-address>/health" || true
```

Exact paths depend on your ingress and version; use OpenAPI from your environment for subsystem routes.

### Optional: Etrino (dataview `--sql`)

On a **KWeaver Core–only** install, `kweaver dataview query <id>` without `--sql` usually works (paging against the view definition). **Ad-hoc SQL** via `kweaver dataview query --sql "..."` requires **`vega-calculate-coordinator`** in the cluster. That comes from the **Etrino** Helm stack: **`vega-hdfs`**, **`vega-calculate`** (includes the coordinator), and **`vega-metadata`**.

On a cluster where Core is already running, use the new `etrino` subcommand in `deploy.sh`:

```bash
cd kweaver-core/deploy
./deploy.sh etrino install
# Check status
./deploy.sh etrino status
# Uninstall
./deploy.sh etrino uninstall
```

If needed, add `--config=/path/to/config.yaml`. Under the hood this still runs the Etrino installer logic: adds the Helm repo alias **`myrepo`** (`https://kweaver-ai.github.io/helm-repo/`), labels nodes, prepares HDFS directories, and installs **`vega-hdfs` → `vega-calculate` → `vega-metadata`**. Ensure nodes have disk and resources; image registries in the chart defaults may differ from Core, so override `image.registry` / values when needed.

**If you install DIP anyway:** `./deploy.sh kweaver-dip install` runs the same Etrino installation flow after DIP charts, so you do not need to run it twice.

### Configure Models (required for semantic search and Agent)

KWeaver does not include pre-configured models by default. To use **semantic search** (`kweaver bkn search`) or **Decision Agent**, you must register an LLM and an Embedding model first.

For full details, see [Model Management](../model.md). Minimal registration example:

```bash
# Register an LLM (DeepSeek example)
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "deepseek-chat",
  "model_series": "deepseek",
  "max_model_len": 8192,
  "model_config": {
    "api_key": "<your-api-key>",
    "api_model": "deepseek-chat",
    "api_url": "https://api.deepseek.com/chat/completions"
  }
}'

# Register an Embedding model
kweaver call /api/mf-model-manager/v1/small-model/add -d '{
  "model_name": "bge-m3",
  "model_type": "embedding",
  "model_config": {
    "api_url": "https://api.siliconflow.cn/v1/embeddings",
    "api_model": "BAAI/bge-m3",
    "api_key": "<your-api-key>"
  },
  "batch_size": 32,
  "max_tokens": 512,
  "embedding_dim": 1024
}'

# Verify
kweaver call '/api/mf-model-manager/v1/llm/list?page=1&size=50'
kweaver call '/api/mf-model-manager/v1/small-model/list?page=1&size=50'
```

Enabling BKN semantic search also requires a ConfigMap change — see [Model Management — Enable BKN Semantic Search](../model.md#enable-bkn-semantic-search).

### Troubleshooting

See [deploy/README.md — Troubleshooting](../../../deploy/README.md).

## Next steps

- [Quick start](../quick-start.md)
