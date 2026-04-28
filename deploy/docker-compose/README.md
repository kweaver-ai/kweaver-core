# KWeaver Core — Docker Compose (B1 demo subset)

This directory provides a **lean demo** stack: **infrastructure + one-shot DB migrator + 11 KWeaver business services** (ontology/query, data model, model factory, Vega, data-connection). It is **not** the full `deploy.sh kweaver-core --minimum` Helm surface (no agent-*, dataflow, coderunner, doc-convert, sandbox, oss-gateway, or bundled otel/observability).

Service definitions, ports, volumes, and `configs/kweaver/**` templates are aligned with charts in **[kweaver-ai/helm-repo](https://github.com/kweaver-ai/helm-repo)**. Re-run `tools/extract-helm-templates.py` after unpacking charts in `/tmp/kc-charts-unpacked` if you refresh templates; then reconcile any **manual** Compose tweaks (e.g. `mf-model-*/cm-kw-yaml.env.tmpl` host wiring).

The **mandatory** sanity check is **`./setup.sh`** (renders templates + runs `docker compose config`). That does **not** pull images.

## What is included (19 `docker compose` services)

| Layer | Services |
|-------|----------|
| **Infra (7)** | `mariadb`, `redis`, `zookeeper`, `kafka`, `opensearch`, `minio`, `nginx` |
| **Job (1)** | `kweaver-core-data-migrator` (one-shot; others wait for `service_completed_successfully`) |
| **KWeaver (11)** | `bkn-backend`, `ontology-query`, `mdl-data-model`, `mdl-uniquery`, `mdl-data-model-job`, `mf-model-manager`, `mf-model-api`, `vega-backend`, `vega-gateway`, `vega-gateway-pro`, `data-connection` |

**Not included:** `agent-backend`, `agent-operator-integration`, `agent-retrieval`, `agent-observability`, `dataflow`, `flow-stream-data-pipeline`, `coderunner`, `dataflowtools`, `doc-convert-*`, `sandbox`, `oss-gateway-backend`, `otelcol-contrib`. Generated configs may still mention these hosts in comments or deps; calls routed to `nginx` for DNS will **502** unless you add those services.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and **Docker Compose v2** (`docker compose`). **v2.17+ recommended** (v2.20+ preferred).
- **~10–12 GB RAM** typical for this subset (OpenSearch + Kafka + MariaDB).
- **Image registry:** Default images use Huawei SWR under `swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip`. Keep `.env` as `IMAGE_REGISTRY=swr.cn-east-3.myhuaweicloud.com/kweaver-ai` and `DIP_NAMESPACE=dip`. If pull fails (`You may not login yet`, etc.), run `docker compose config --images`, confirm paths include `/dip/`, then `docker login swr.cn-east-3.myhuaweicloud.com` if your org requires it. **Public** images (MariaDB, Redis, Kafka, Zookeeper, Nginx, OpenSearch, MinIO) do not need SWR.

## One-time setup

```bash
cd deploy/docker-compose
chmod +x ./setup.sh
./setup.sh
```

`setup.sh` will:

1. Copy `.env.example` → `.env` (gitignored) if `.env` is missing.
2. Resolve passwords for `MARIADB_ROOT_PASSWORD`, `MARIADB_PASSWORD`, `MINIO_ROOT_PASSWORD` (CLI > env > shared `-p` / `PASSWORD` > `.env` > prompt > error).
3. Run `tools/render_compose_configs.py`: substitute `configs/kweaver/**/*.tmpl` → `configs/generated/...` (includes `cm-kw-yaml.env.tmpl` → `cm-kw-yaml.env` for `mf-model-*`).

### Password rule

Use only `[A-Za-z0-9_-]` — values are written to `.env` and embedded in generated configs where required.

## Bringing the stack up (optional)

```bash
docker compose up -d
```

Expect `kweaver-core-data-migrator` to **complete once**; application services start after it.

**Stopping:**

```bash
docker compose down
```

### Smoke checks (local)

```bash
curl -sS http://localhost:8080/healthz    # expect: ok
curl -sI http://localhost:8080/api/bkn-backend/v1/nonexistent  # routing to bkn-backend (401/404 acceptable)
docker compose logs bkn-backend 2>&1 | head -50
```

## Infra-only smoke (no KWeaver images)

```bash
docker compose up -d mariadb redis zookeeper kafka opensearch minio
```

(Optional: add `nginx` if you mount only `configs/nginx/default.conf` — run `./setup.sh` first if anything references `configs/generated/`.)

## Entry points

| What | URL / port |
|------|------------|
| APIs via nginx | `http://<ACCESS_HOST>:<KWEAVER_HTTP_PORT>` — default `http://localhost:8080` |
| Health | `http://localhost:8080/healthz` |

## Limitations vs Kubernetes / Helm

- No ingress TLS, no multi-replica HA, no Helm hooks; migrator is a one-shot Compose service.
- **Auth / IAM:** charts reference `authorization-private`, `hydra-admin`, etc. Extracted YAML may rewrite hosts to `nginx` so DNS resolves; without real IAM, those calls fail. `AUTH_ENABLED=false` is set where the compose file exposes it.
- **Redis:** `mf-model-*` env uses a **standalone** Redis (`REDISCLUSTERMODE=false`, port `6379`), not Sentinel as in some charts.

## Developer: refresh templates from Helm

```bash
# Unpack charts to /tmp/kc-charts-unpacked (see tools/extract-helm-templates.py header)
python3 deploy/docker-compose/tools/extract-helm-templates.py
```

Then reconcile manual edits (especially `mf-model-*/cm-kw-yaml.env.tmpl` and `dm_svc.conf.tmpl` for Compose).

## Remote lab (e.g. Ubuntu VM)

After syncing this directory to the host:

```bash
cd deploy/docker-compose
./setup.sh -p YOUR_PASSWORD -y
sudo docker compose up -d   # if daemon requires sudo
```

Verify `curl http://127.0.0.1:8080/healthz` and backend routing as above. (Automated SSH from CI/agents may be blocked by firewall.)

---

For Kubernetes deployment, see [../README.md](../README.md).
