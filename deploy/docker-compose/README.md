# KWeaver Core — Docker Compose (minimum stack)

This directory provides a **minimal** local or lab stack for KWeaver Core (no ISF, no auth / business-domain bundles by default), similar in spirit to `./deploy.sh kweaver-core install --minimum` but using **Docker Compose** instead of Kubernetes.

**Important:** This document does **not** require you to successfully run the full stack on your laptop; real bring-up happens in your environment. The **mandatory** sanity check is **`docker compose config`** (no image pull, no containers).

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/) (Compose v2 plugin: `docker compose`).
- Enough disk and RAM for MariaDB, Kafka, OpenSearch, Redis, MinIO, and many application containers (typical dev machine: 16 GB+ RAM recommended).
- **Network access** to the image registry in `.env` (`IMAGE_REGISTRY`). Change it if you mirror images elsewhere.

## One-time setup

```bash
cd deploy/docker-compose
chmod +x ./setup.sh
./setup.sh
```

`setup.sh` will:

1. Copy `.env.example` → `.env` (gitignored) if `.env` is missing.
2. Resolve passwords for `MARIADB_ROOT_PASSWORD`, `MARIADB_PASSWORD`, `MINIO_ROOT_PASSWORD`. For each, the value is taken from (in order):
   - matching per-field CLI flag (e.g. `--mariadb-password=…`)
   - matching per-field env var (e.g. `MARIADB_PASSWORD=…`)
   - shared CLI flag `-p / --password=…` or shared env var `PASSWORD=…` (one value applied to all three fields)
   - current value in `.env`
   - interactive prompt (TTY only; `Enter` keeps the current value)
   - `.env.example` default (with a warning if it is still the placeholder)
3. Always rewrite `SANDBOX_DATABASE_URL` so its password matches `MARIADB_PASSWORD`.
4. Render `configs/kweaver/config.yaml.template` → `configs/generated/config.yaml`.
5. Run `docker compose config` as an offline sanity check.

### Examples

Interactive prompt (default — first asks for one shared password, then per-field if you skip):

```bash
./setup.sh
```

Use the **same** password for MariaDB root + MariaDB user + MinIO root in one shot:

```bash
./setup.sh --password='OnePw_2026' --non-interactive
# or
PASSWORD='OnePw_2026' ./setup.sh --non-interactive
```

Pass per-field passwords on the command line (good for automation / CI):

```bash
./setup.sh \
  --mariadb-root-password='RootPw_2026' \
  --mariadb-password='AdpPw_2026' \
  --minio-root-password='MinioPw_2026' \
  --non-interactive
```

Pass via environment variables:

```bash
MARIADB_ROOT_PASSWORD=RootPw_2026 \
MARIADB_PASSWORD=AdpPw_2026 \
MINIO_ROOT_PASSWORD=MinioPw_2026 \
./setup.sh --non-interactive
```

Or just edit `.env` by hand and re-run `./setup.sh --non-interactive`.

> Password tip: stick to `[A-Za-z0-9_-]` so the value can be embedded in `SANDBOX_DATABASE_URL` (a Python DSN) without URL-encoding.

## Bringing the stack up (optional)

When you are ready to run containers in **your** environment:

```bash
docker compose up -d
```

Expect the **data migrator** job to run once (`kweaver-core-data-migrator`); core services should start after migrations. Check logs with `docker compose logs -f <service>`.

**Stopping:**

```bash
docker compose down
```

## Entry points

| What | URL / port |
|------|------------|
| KWeaver API (via nginx) | `http://<ACCESS_HOST>:<KWEAVER_HTTP_PORT>` — default `http://localhost:8080` |
| Sandbox control plane | **Not** behind nginx; map `8000` (container) → **`SANDBOX_HTTP_PORT`** on the host (default **8001**) — `http://localhost:8001` |

Set `ACCESS_HOST`, `KWEAVER_HTTP_PORT`, and `SANDBOX_HTTP_PORT` in `.env` to avoid port clashes.

## What is included

- **Infra:** MariaDB, Redis (no password on the internal network — matches generated config), Kafka (Zookeeper + single broker, PLAINTEXT), OpenSearch (security plugin disabled), MinIO.
- **Core microservices** as in the [0.7.0 release manifest](../release-manifests/0.7.0/kweaver-core.yaml), with **version overrides** from `.env.example` where charts differ.
- **Nginx** as a single HTTP entry, proxying path prefixes to backend container ports (derived from Helm defaults in this repo where available).
- **Optional:** `otelcol-contrib` and observability sidecars as in the manifests; OpenTelemetry collector routes are not fully wired in this minimal compose.

## Limitations vs Kubernetes / Helm

- No ingress TLS, no multi-replica HA, no Helm hooks — only what Compose can express.
- Some images expect specific `CONFIG_FILE` paths and volume layouts; the compose file uses a shared bind mount to `configs/generated/config.yaml` **where the chart does so**; if a service fails at startup, compare with its chart `values.yaml` in `../charts/` and adjust `environment` / `volumes`.
- **Sandbox** uses the Docker socket and host-mapped MinIO; review security before using outside a trusted dev machine.

## Offline / CI validation (no full stack)

Recommended checks (no container run required beyond what `config` does internally):

```bash
docker compose config
```

Optional: add the same command to CI (e.g. GitHub Actions) so every PR validates Compose syntax and variable substitution.

---

For production-style deployment on Kubernetes, see [../README.md](../README.md).
