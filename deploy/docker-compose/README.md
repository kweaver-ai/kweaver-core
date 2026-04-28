# KWeaver Core — Docker Compose

This directory mirrors the **`deploy.sh kweaver-core install --minimum`** Helm release set from
[`deploy/release-manifests/0.7.0/kweaver-core.yaml`](../release-manifests/0.7.0/kweaver-core.yaml)
on a single Docker host. `--minimum` only adds Helm `--set auth.enabled=false businessDomain.enabled=false`;
the release set itself is the same. Optional ISF (`enabledIf: auth.enabled`) is **not** installed.

The local source of truth for which services run, which charts they map to, and which image
tag/registry they use is [`compose-manifest.yaml`](./compose-manifest.yaml). `compose.sh`,
`setup.sh`, and `tools/extract-helm-templates.py` all read from it; **edit the manifest, not
those scripts**.

The mandatory sanity check is **`./setup.sh`** (renders templates, validates manifest ↔
`docker-compose.yml`, runs `docker compose config`). It does **not** pull images.

## What is included (32 `docker compose` services)

- **Infra (6):** `mariadb`, `redis`, `zookeeper`, `kafka`, `opensearch`, `minio`.
- **Job (1):** `kweaver-core-data-migrator` (one-shot; others wait for `service_completed_successfully`).
- **KWeaver app (24):** `bkn-backend`, `mf-model-manager`, `mf-model-api`, `ontology-query`,
  `vega-backend`, `vega-gateway`, `vega-gateway-pro`, `data-connection`, `mdl-data-model`,
  `mdl-uniquery`, `mdl-data-model-job`, `agent-operator-integration`, `agent-retrieval`,
  `agent-backend`, `dataflow`, `flow-stream-data-pipeline`, `coderunner`, `dataflowtools`,
  `doc-convert-gotenberg`, `doc-convert-tika`, `sandbox`, `oss-gateway-backend`,
  `otelcol-contrib`, `agent-observability`.
- **Entry (1):** `nginx` (starts with the app phase; upstreams are KWeaver services).

Notes:

- `dataflow` chart contributes two Compose services (`dataflow` for flow-automation +
  ecron-management; `flow-stream-data-pipeline` for the SDP container).
- `doc-convert` is split into `doc-convert-gotenberg` + `doc-convert-tika`
  (Tika joins gotenberg via `network_mode: service:doc-convert-gotenberg`).
- `coderunner` chart contributes `coderunner` + `dataflowtools` (sidecar image).
- `sandbox` runs control-plane only; the chart’s `sandbox-template-python-basic` image is
  referenced through generated configs (env), not as a Compose service.
- Some KWeaver services need MongoDB / SASL / IAM that are not provisioned by this Compose
  stack; they will start but their dataflow paths can fail at runtime. Use this stack for
  configuration smoke tests and core-line UI/API checks first.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and **Docker Compose v2** (`docker compose`).
  v2.17+ recommended (v2.20+ preferred).
- **~16 GB RAM** typical for the full set (OpenSearch + Kafka + MariaDB + 24 app containers).
- **Image registry:** Default images use Huawei SWR under
  `swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip`. Keep `.env` as
  `IMAGE_REGISTRY=swr.cn-east-3.myhuaweicloud.com/kweaver-ai` and `DIP_NAMESPACE=dip`.
  `agent-observability` is published at the registry root (no `dip/` segment) — that is
  recorded in `compose-manifest.yaml` so generated `image:` paths stay correct.
- **Python 3 + PyYAML** on the host (used by the manifest tool and template renderer).

## One-time setup

```bash
cd deploy/docker-compose
chmod +x ./setup.sh ./compose.sh
./setup.sh
```

`setup.sh` will:

1. Copy `.env.example` → `.env` (gitignored) if `.env` is missing.
2. Resolve passwords for `MARIADB_ROOT_PASSWORD`, `MARIADB_PASSWORD`, `MINIO_ROOT_PASSWORD`
   (CLI > env > shared `-p` / `PASSWORD` > `.env` > prompt > error).
3. Default `SANDBOX_DATABASE_URL` from `MARIADB_*` when unset (sandbox uses `mariadb` + the
   `sandbox` schema created by `configs/mariadb/init.sql`).
4. Run `python3 tools/manifest.py check-compose` to ensure `docker-compose.yml` services
   match `compose-manifest.yaml`.
5. Backfill any missing `*_VERSION` (tagEnv) value in `.env` from `compose-manifest.yaml`
   and run `python3 tools/manifest.py check-env .env` to warn if your `.env` pinned a
   tag that no longer matches the manifest. `docker-compose.yml` references each tag as
   `${X_VERSION:?...}`, so a missing version short-circuits `docker compose` with a clear
   error pointing back to `setup.sh`.
6. Run `tools/render_compose_configs.py`: substitute `configs/kweaver/**/*.tmpl` →
   `configs/generated/...` and merge per-service env files (`dataflow/flow-automation.env`,
   `coderunner/coderunner.env`, `coderunner/dataflowtools.env`, `sandbox/sandbox.env`).
7. Run `docker compose config` to validate the final stack.

### Password rule

Use only `[A-Za-z0-9_-]` — values are written to `.env` and embedded in generated configs
where required.

### Secret templates

`configs/kweaver/**/secret-*/*.tmpl` use `__KEY__` placeholders (e.g. `__MARIADB_PASSWORD__`,
`__REDIS_PASSWORD__`, `__KAFKA_PASSWORD__`) that the renderer fills from `.env`. Chart
fixtures like `xxxxxx` / `root` / `minioadmin` are intentionally **not** present in the
templates; if they reappear after a re-extract, `tools/extract-helm-templates.py` knows how
to map known secret keys back to placeholders (see `SECRET_KEY_TO_PLACEHOLDER`).

## Bringing the stack up

```bash
./compose.sh infra up
./compose.sh app up
```

`infra up` starts public dependency images. `app up` then starts the migrator, all 24
KWeaver services from the manifest, and `nginx`. If an SWR application image pull fails,
infra can stay running while you fix the image path or registry login.

```bash
./compose.sh app pull   # pre-pull application images
./compose.sh all up     # infra + app
./compose.sh all down   # stop everything
```

`kweaver-core-data-migrator` completes once; the rest wait for it.

### Vega only (mirrors `deploy.sh` grouping)

Start **infra**, **migrator**, then only the Vega services (`vega-backend`,
`data-connection`, `vega-gateway`, `vega-gateway-pro`). This does **not** start `nginx` —
compose `nginx.depends_on` covers the full app set; use `./compose.sh app up` or
`./compose.sh all up` if you need `http://localhost:8080`.

```bash
./compose.sh vega up
```

Other actions: `./compose.sh vega pull|down|restart|status|logs`.

## Entry points

| What                       | URL / port                                                              |
|----------------------------|-------------------------------------------------------------------------|
| APIs via nginx             | `http://<ACCESS_HOST>:<KWEAVER_HTTP_PORT>` (default `http://localhost:8080`) |
| Health                     | `http://localhost:8080/healthz`                                         |
| Sandbox control plane (HTTP) | `http://localhost:${SANDBOX_HTTP_PORT:-8001}` (the chart’s ingress is bypassed in Compose) |

`nginx` proxies the standard prefixes (`/api/bkn-backend/`, `/api/agent-factory/`,
`/api/automation/`, `/api/coderunner/`, `/api/oss-gateway/`,
`/api/agent-observability/`, `/api/sandbox/`, etc.).

### Smoke checks (local)

```bash
curl -sS http://localhost:8080/healthz                                   # ok
curl -sI http://localhost:8080/api/bkn-backend/v1/nonexistent            # 401/404 = routed
docker compose logs bkn-backend 2>&1 | head -50
```

## Limitations vs Kubernetes / Helm

- No ingress TLS, no multi-replica HA, no Helm hooks; migrator is a one-shot Compose service.
- **Auth / IAM:** charts reference `authorization-private`, `hydra-admin`, etc. Extracted
  YAML may rewrite hosts to `nginx` so DNS resolves; without real IAM, those calls fail.
  Wherever `AUTH_ENABLED` / `BUSINESS_DOMAIN_ENABLED` exist they are forced to `false`.
- **MongoDB / SASL Kafka:** the dataflow chart expects MongoDB and SASL-authenticated Kafka.
  Compose runs plaintext Kafka and no MongoDB; flow-automation may degrade or fail at
  runtime in those code paths.
- **Redis:** infra ships a single-node Redis; chart configs that ask for Sentinel are
  rewritten via placeholders — review your service’s rendered config if Sentinel features
  matter.

## Developer: refresh templates from Helm

```bash
# 1. Unpack charts to /tmp/kc-charts-unpacked (one folder per <chart>-<version>).
# 2. Re-extract — chart list comes from compose-manifest.yaml, so version bumps propagate.
python3 deploy/docker-compose/tools/extract-helm-templates.py
```

After extraction, reconcile any manual Compose-only edits (e.g. `mf-model-*/cm-kw-yaml.env.tmpl`,
`dm_svc.conf.tmpl`, the `__KEY__` placeholders mentioned above).

### Manifest-driven helpers

```bash
python3 tools/manifest.py services --phase=app    # service list (used by compose.sh)
python3 tools/manifest.py charts                  # <chart-folder>\t<out_sub> (used by extract)
python3 tools/manifest.py env-defaults            # tagEnv defaults (used by setup.sh)
python3 tools/manifest.py check-compose           # diff manifest vs docker-compose.yml
python3 tools/manifest.py check-env [path]        # diff manifest tagEnv vs an .env file
```

## Remote lab (e.g. Ubuntu VM)

After syncing this directory to the host:

```bash
cd deploy/docker-compose
./setup.sh -p YOUR_PASSWORD -y
sudo docker compose up -d   # if daemon requires sudo
```

Verify `curl http://127.0.0.1:8080/healthz` and backend routing as above.

---

For Kubernetes deployment, see [../README.md](../README.md).
