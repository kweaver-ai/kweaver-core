# 🚢 Ship and deploy

This page covers **prerequisites**, **install steps**, and **post-install checks** for KWeaver Core.

> Use the `deploy.sh` script under the `deploy/` directory from your product bundle or build tree.

---

## 🧱 Prerequisites

Prepare the host, network, and client tooling before you deploy.

### Host requirements

> Run installation as `root` or with `sudo`.

| Item | Minimum | Recommended |
| --- | --- | --- |
| OS | CentOS 8+, openEuler 23+ | CentOS 8+ |
| CPU | 16 cores | 16 cores |
| Memory | 48 GB | 64 GB |
| Disk | 200 GB | 500 GB |

### Host preparation (typical Linux)

```bash
# 1. Disable firewall (or open required ports per your policy)
systemctl stop firewalld && systemctl disable firewalld

# 2. Disable swap
swapoff -a && sed -i '/ swap / s/^/#/' /etc/fstab

# 3. SELinux permissive if needed
setenforce 0

# 4. Install container runtime (example: containerd)
# dnf install containerd.io   # adjust for your distro
```

> Exact steps depend on your OS; follow the deployment guide shipped with your release.

### Network access

The deploy scripts may need outbound access to mirrors and registries, for example:

| Domain | Purpose |
| --- | --- |
| `mirrors.aliyun.com` | RPM mirrors |
| `mirrors.tuna.tsinghua.edu.cn` | containerd RPM mirror |
| `registry.aliyuncs.com` | Kubernetes images |
| `swr.cn-east-3.myhuaweicloud.com` | KWeaver images |
| `repo.huaweicloud.com` | Helm binary |
| `kweaver-ai.github.io` | Helm chart repo |

### Client tooling (after deploy)

On your workstation (with network access to the cluster):

- **kubectl** — optional but useful for health checks
- **kweaver CLI** — install via npm package `@kweaver-ai/kweaver-sdk`

```bash
npm install -g @kweaver-ai/kweaver-sdk
# or: npx kweaver --help
```

- **curl** — for raw HTTP API calls

---

## 📥 Enter the deploy directory

From your extracted bundle or repo:

```bash
cd deploy
chmod +x deploy.sh
```

> Adjust the path if your layout differs.

---

## 🚀 Install KWeaver Core

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

> The script may prompt for **access address** and detect **API server address** automatically.

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

### Useful commands

```bash
./deploy.sh kweaver-core status
./deploy.sh kweaver-core uninstall
./deploy.sh --help
```

### What gets installed

1. Single-node Kubernetes (if needed), storage, ingress
2. Data services: MariaDB, Redis, Kafka, ZooKeeper, OpenSearch (as defined by release manifests)
3. KWeaver Core application Helm charts

> For uninstall and cluster reset, follow the operations guide bundled with your release.

---

## 🛡️ Administrator tool after a full install (kweaver-admin)

After a full install (with `auth.enabled=true` and `businessDomain.enabled=true`), platform-level operations — **users, organizations, roles, models, audit** — are managed via the standalone npm CLI [`@kweaver-ai/kweaver-admin`](https://github.com/kweaver-ai/kweaver-admin). It is complementary to the `kweaver` CLI from `kweaver-sdk`:

| CLI | Audience | Scope |
| --- | --- | --- |
| `kweaver` (`@kweaver-ai/kweaver-sdk`) | End users / Agents | BKN, Decision Agent, Action, Skill, query |
| `kweaver-admin` (`@kweaver-ai/kweaver-admin`) | Platform administrators | Users, organizations, roles, models, audit, raw HTTP |

**When to install:** after a full install (`./deploy.sh kweaver-core install` without `--minimum`). **On a `--minimum` install most `kweaver-admin` commands return 401 / 404 — that is expected, the relevant services are not deployed.**

**Backend services it talks to (from the kweaver-admin architecture doc):** `user-management` / `deploy-manager` / `deploy-auth` / `eacp` / `mf-model-manager` / OAuth2 (Hydra) — exactly the set enabled by a full install.

### 📥 Install

Requires **Node.js 18+**. Credentials are stored under `~/.kweaver-admin/platforms/`, isolated from `~/.kweaver/`.

```bash
npm install -g @kweaver-ai/kweaver-admin
kweaver-admin --version
kweaver-admin --help
```

### 🔑 Login

```bash
# Browser OAuth2 (skip TLS for self-signed certs)
kweaver-admin auth login https://<access-address> -k

# Username/password (CI / headless)
kweaver-admin auth login https://<access-address> -u <user> -p <password> -k

# Or via environment variables (CI / headless)
export KWEAVER_BASE_URL=https://<access-address>
export KWEAVER_ADMIN_TOKEN=<bearer-token>   # preferred; falls back to KWEAVER_TOKEN

# Inspect session and identity
kweaver-admin auth status
kweaver-admin auth whoami
kweaver-admin auth list
```

> The token stores of `kweaver-admin` and `kweaver` are independent — both can coexist on the same machine for separate admin / user identities.

### 🧰 Common admin tasks

#### Organizations (departments)

```bash
kweaver-admin org tree                # tree view of departments
kweaver-admin org list                # paginated list
kweaver-admin org create              # create a department
kweaver-admin org members <orgId>     # list members
```

#### Users

```bash
kweaver-admin user list
kweaver-admin user create --login alice            # default password 123456, forced change at first sign-in
kweaver-admin user reset-password -u alice         # admin reset
kweaver-admin user roles <userId>
kweaver-admin user assign-role <userId> <roleId>
kweaver-admin user revoke-role <userId> <roleId>
```

#### Roles

```bash
kweaver-admin role list
kweaver-admin role get <roleId>
kweaver-admin role add-member <roleId> -u alice
kweaver-admin role remove-member <roleId> -u alice
```

#### Models (LLM / Embedding)

```bash
kweaver-admin llm list
kweaver-admin llm add
kweaver-admin llm test <modelId>

kweaver-admin small-model list
kweaver-admin small-model add
kweaver-admin small-model test <modelId>
```

> Equivalent to invoking `kweaver call /api/mf-model-manager/...` (see [Model management](model.md)); for day-to-day admin work the `kweaver-admin llm` / `small-model` subcommands offer better validation and output.

#### Audit

```bash
kweaver-admin audit list \
  --user alice --start 2026-04-01 --end 2026-04-30
```

#### Raw HTTP (with auth header)

```bash
kweaver-admin call /api/user-management/v1/management/users -X GET
kweaver-admin --json call /api/eacp/v1/... -X POST -d '{"...":"..."}'
```

### ⚠️ Things you must know

- **New users created via `user create` always start with the platform default password `123456`** and are forced to change it at first sign-in. This is documented upstream behavior of the ISF user store (`Usrm_AddUser` thrift does not accept a password parameter). Hand the account to the user over a secure channel; for lost-password rotation use `kweaver-admin user reset-password`.
- **Separation-of-duties built-in accounts** — `system / admin / security / audit` must not be casually modified; operators should use **individual accounts** rather than the shared `admin` for traceable audit logs.
- **First-login forced password change (error `401001017`)**: when `kweaver-admin auth login` hits this code, on a TTY the CLI guides you to set a new password and retries the login; in non-TTY contexts pass `--new-password '<new>'` to do it in one shot (same flow as the `kweaver` CLI; see also [`kweaver-admin/docs/SECURITY.md`](https://github.com/kweaver-ai/kweaver-admin/blob/main/docs/SECURITY.md)).
- **TLS:** `-k` / `--insecure` (or env var `KWEAVER_TLS_INSECURE=1`) is for development / self-signed certs only — never use in production.
- **Capabilities not exposed by the Web console — CLI is the primary path:** department writes (`Usrm_AddDepartment` / `Usrm_EditDepartment`), user updates (`Usrm_EditUser` fallback), user-role lookup (`role list` + `role members` fallback), etc. See [`kweaver-admin/docs/SECURITY.md`](https://github.com/kweaver-ai/kweaver-admin/blob/main/docs/SECURITY.md).

### 📖 Further reading

- [`kweaver-admin` repository README](https://github.com/kweaver-ai/kweaver-admin)
- [`ARCHITECTURE.md`](https://github.com/kweaver-ai/kweaver-admin/blob/main/ARCHITECTURE.md) — command tree and backend API mapping
- [`docs/SECURITY.md`](https://github.com/kweaver-ai/kweaver-admin/blob/main/docs/SECURITY.md) — tokens, TLS, audit and fallback routes

---

## ✅ After install (check cluster and API)

When `deploy.sh kweaver-core install` finishes, confirm the cluster and that you can reach the platform.

### Kubernetes

```bash
kubectl get nodes
kubectl get pods -A
```

> Wait until core namespaces show `Running` / `Ready` for critical workloads.

### Deploy script status

```bash
./deploy.sh kweaver-core status
```

### CLI

```bash
kweaver auth login https://<access-address> -k
kweaver bkn list
```

> Use the same host as `--access_address` or the node IP from the installer. Omit `-k` when using a trusted TLS certificate.

### HTTP (optional)

```bash
curl -sk "https://<access-address>/health" || true
```

> Exact paths depend on your ingress and version; use OpenAPI from your environment for subsystem routes.

---

## 🧮 Optional: Etrino (dataview `--sql`)

On a **KWeaver Core–only** install, `kweaver dataview query <id>` without `--sql` usually works (paging against the view definition).

> **Ad-hoc SQL** via `kweaver dataview query --sql "..."` requires **`vega-calculate-coordinator`** in the cluster. That comes from the **Etrino** Helm stack: **`vega-hdfs`**, **`vega-calculate`** (includes the coordinator), and **`vega-metadata`**.

On a cluster where Core is already running:

```bash
./deploy.sh etrino install
./deploy.sh etrino status
./deploy.sh etrino uninstall
```

> Add `--config=/path/to/config.yaml` if needed. The flow adds the Helm repo alias **`myrepo`** (`https://kweaver-ai.github.io/helm-repo/`), labels nodes, prepares HDFS directories, and installs **`vega-hdfs` → `vega-calculate` → `vega-metadata`**. Ensure nodes have disk and resources; override `image.registry` / values when chart defaults do not match your registry.

> **If you install DIP anyway:** `./deploy.sh kweaver-dip install` runs the same Etrino installation flow after DIP charts, so you do not need to run it twice.

---

## 🧠 Configure models

KWeaver does not include pre-configured models by default. To use **semantic search** (`kweaver bkn search`) or **Decision Agent**, register an LLM and an Embedding model first.

> Full details: [Model management](model.md). Minimal registration example:

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

> Enabling BKN semantic search also requires a ConfigMap change — see [Model management — Enable BKN semantic search](model.md#enable-bkn-semantic-search).

---

## 🔧 Troubleshooting

Use `kubectl` logs and deploy script output; follow the operations guide bundled with your release for a full checklist.

---

## 📖 Next steps

| Goal | Doc |
| --- | --- |
| First commands and BKN | [Quick start](quick-start.md) |
| Model registration | [Model management](model.md) |
| Enable BKN semantic search | [Enable BKN semantic search](model.md#enable-bkn-semantic-search) |
