# KWeaver Core — Docker Compose

本目录在单机 Docker 上对齐 **`deploy.sh kweaver-core install --minimum`** 在
Helm 上的 release 集合（来自
[`deploy/release-manifests/0.7.0/kweaver-core.yaml`](../release-manifests/0.7.0/kweaver-core.yaml)）。
`--minimum` 仅为 Helm 增加 `--set auth.enabled=false businessDomain.enabled=false`，
release 集合本身不变。可选的 ISF（`enabledIf: auth.enabled`）**不**安装。

「哪些服务、对应哪个 chart、用什么镜像/tag/registry」的本地唯一来源是
[`compose-manifest.yaml`](./compose-manifest.yaml)。`compose.sh`、`setup.sh` 和
`tools/extract-helm-templates.py` 都从中读取；**修改请改 manifest，不要改这些脚本**。

必跑的检查是 **`./setup.sh`**（渲染模板、校验 manifest ↔ `docker-compose.yml`、
执行 `docker compose config`）；它**不会**拉取镜像。

## 包含内容（32 个 `docker compose` 服务）

- **基础设施（6）：** `mariadb`、`redis`、`zookeeper`、`kafka`、`opensearch`、`minio`。
- **任务（1）：** `kweaver-core-data-migrator`（一次性；其他服务等待其
  `service_completed_successfully`）。
- **KWeaver 应用（24）：** `bkn-backend`、`mf-model-manager`、`mf-model-api`、
  `ontology-query`、`vega-backend`、`vega-gateway`、`vega-gateway-pro`、
  `data-connection`、`mdl-data-model`、`mdl-uniquery`、`mdl-data-model-job`、
  `agent-operator-integration`、`agent-retrieval`、`agent-backend`、`dataflow`、
  `flow-stream-data-pipeline`、`coderunner`、`dataflowtools`、`doc-convert-gotenberg`、
  `doc-convert-tika`、`sandbox`、`oss-gateway-backend`、`otelcol-contrib`、
  `agent-observability`。
- **入口（1）：** `nginx`（随 app 阶段启动；上游为 KWeaver 服务）。

说明：

- `dataflow` chart 拆成两个 Compose 服务：`dataflow`（flow-automation +
  ecron-management）和 `flow-stream-data-pipeline`（SDP 容器）。
- `doc-convert` 拆成 `doc-convert-gotenberg` + `doc-convert-tika`，
  Tika 通过 `network_mode: service:doc-convert-gotenberg` 共享网络。
- `coderunner` chart 拆成 `coderunner` + `dataflowtools`（sidecar 镜像）。
- `sandbox` 仅运行 control-plane；chart 中的 `sandbox-template-python-basic`
  仅在生成的配置里被引用（环境变量），不作为 Compose 服务。
- 部分 KWeaver 服务在 K8s 下还依赖 MongoDB / SASL / IAM 等本 Compose 栈未提供的组件；
  容器可以起来，但相关业务路径可能在运行时失败。本栈优先用于配置冒烟与核心 UI/API
  联通验证。

## 前置条件

- [Docker](https://docs.docker.com/get-docker/) 与 **Docker Compose v2**（`docker compose`）。
  建议 v2.17+，更推荐 v2.20+。
- 启全量大约需要 **约 16 GB 内存**（OpenSearch + Kafka + MariaDB + 24 个应用容器）。
- **镜像仓库：** 默认镜像位于华为云 SWR
  `swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip`。请保持 `.env` 中
  `IMAGE_REGISTRY=swr.cn-east-3.myhuaweicloud.com/kweaver-ai`、`DIP_NAMESPACE=dip`。
  `agent-observability` 发布在 registry 根（不带 `dip/` 段）——这条已记录在
  `compose-manifest.yaml` 中，生成的 `image:` 路径会保持正确。
- 主机需要 **Python 3 + PyYAML**（manifest 工具与模板渲染会用到）。

## 一次性设置

```bash
cd deploy/docker-compose
chmod +x ./setup.sh ./compose.sh
./setup.sh
```

`setup.sh` 会：

1. 若 `.env` 不存在，从 `.env.example` 复制（`.env` 已被 git 忽略）。
2. 解析 `MARIADB_ROOT_PASSWORD`、`MARIADB_PASSWORD`、`MINIO_ROOT_PASSWORD`
   （优先级：命令行 > 环境变量 > 共享 `-p` / `PASSWORD` > `.env` > 交互输入 > 报错）。
3. 若未设置 `SANDBOX_DATABASE_URL`，从 `MARIADB_*` 推导（sandbox 使用 `mariadb` +
   `configs/mariadb/init.sql` 中创建的 `sandbox` 库）。
4. 调用 `python3 tools/manifest.py check-compose`，确保 `docker-compose.yml`
   的服务集合与 `compose-manifest.yaml` 一致。
5. 用 `compose-manifest.yaml` 中的 `tagEnv` 默认值，回填 `.env` 缺失的
   `*_VERSION`，并执行 `python3 tools/manifest.py check-env .env` 在你本地 pin
   的 tag 与 manifest 不一致时给出警告。`docker-compose.yml` 中所有 tag 都写为
   `${X_VERSION:?...}`，缺失会让 `docker compose` 直接报错并指向 `setup.sh`。
6. 运行 `tools/render_compose_configs.py`：将 `configs/kweaver/**/*.tmpl`
   渲染到 `configs/generated/...`，并合并各服务的 env 文件
   （`dataflow/flow-automation.env`、`coderunner/coderunner.env`、
   `coderunner/dataflowtools.env`、`sandbox/sandbox.env`）。
7. 运行 `docker compose config` 校验最终栈。

### 密码规则

只能使用 `[A-Za-z0-9_-]`。这些值会写入 `.env`，并在需要时嵌入生成配置。

### Secret 模板

`configs/kweaver/**/secret-*/*.tmpl` 使用 `__KEY__` 占位符（如
`__MARIADB_PASSWORD__`、`__REDIS_PASSWORD__`、`__KAFKA_PASSWORD__`），
渲染时由 `.env` 注入。chart 自带的 `xxxxxx` / `root` / `minioadmin` 等占位值
**不**会留在模板里；如果重新抽取 chart 后再次出现，
`tools/extract-helm-templates.py` 会通过 `SECRET_KEY_TO_PLACEHOLDER` 自动还原为
占位符。

## 启动服务

```bash
./compose.sh infra up
./compose.sh app up
```

`infra up` 启动公共依赖镜像；`app up` 接着启动 migrator、manifest 中所有 24 个
KWeaver 服务以及 `nginx`。如果某个 SWR 应用镜像拉取失败，基础设施可继续运行，
你可以同时修正镜像路径或 registry 登录。

```bash
./compose.sh app pull   # 预拉取应用镜像
./compose.sh all up     # infra + app
./compose.sh all down   # 全部停止
```

`kweaver-core-data-migrator` 完成一次后，其他服务才会启动。

### 仅 Vega（与 `deploy.sh` 中分组方式对应）

先启动 **基础设施**、**migrator**，再仅启动 Vega 系（`vega-backend`、
`data-connection`、`vega-gateway`、`vega-gateway-pro`）。**不会**启动 `nginx` —
compose 中 `nginx.depends_on` 覆盖整套应用。如需 `http://localhost:8080` 入口，
请使用 `./compose.sh app up` 或 `./compose.sh all up`。

```bash
./compose.sh vega up
```

其他子命令：`./compose.sh vega pull|down|restart|status|logs`。

## 入口

| 内容                       | URL / 端口                                                              |
|----------------------------|-------------------------------------------------------------------------|
| nginx 代理 API             | `http://<ACCESS_HOST>:<KWEAVER_HTTP_PORT>`，默认 `http://localhost:8080` |
| 健康检查                   | `http://localhost:8080/healthz`                                         |
| Sandbox 控制面（HTTP）     | `http://localhost:${SANDBOX_HTTP_PORT:-8001}`（Compose 中绕过 chart 的 ingress） |

`nginx` 代理常见前缀（`/api/bkn-backend/`、`/api/agent-factory/`、
`/api/automation/`、`/api/coderunner/`、`/api/oss-gateway/`、
`/api/agent-observability/`、`/api/sandbox/` 等）。

### 本地冒烟检查

```bash
curl -sS http://localhost:8080/healthz                                   # ok
curl -sI http://localhost:8080/api/bkn-backend/v1/nonexistent            # 401/404 = 路由通
docker compose logs bkn-backend 2>&1 | head -50
```

## 与 Kubernetes / Helm 的差异

- 没有 ingress TLS、多副本高可用或 Helm hooks；migrator 建模为一次性 Compose 服务。
- **Auth / IAM：** chart 引用 `authorization-private`、`hydra-admin` 等。抽取后的
  YAML 可能将这些 host 改写为 `nginx` 以让 DNS 能解析；没有真实 IAM 时这些调用会失败。
  能配置 `AUTH_ENABLED` / `BUSINESS_DOMAIN_ENABLED` 的服务都被强制为 `false`。
- **MongoDB / SASL Kafka：** dataflow chart 期望 MongoDB 与带 SASL 鉴权的 Kafka。
  Compose 仅提供明文 Kafka，不带 MongoDB；flow-automation 在相关代码路径上可能降级
  或失败。
- **Redis：** 基础设施中只跑单机 Redis；chart 中要求 Sentinel 的字段通过占位符改写，
  如果你依赖 Sentinel 特性，请检查渲染后的配置。

## 开发者：从 Helm 刷新模板

```bash
# 1. 将 chart 解包到 /tmp/kc-charts-unpacked（每个 chart 一个 <chart>-<version> 目录）。
# 2. 重新抽取 — chart 列表来自 compose-manifest.yaml，版本升级会自动同步。
python3 deploy/docker-compose/tools/extract-helm-templates.py
```

抽取后请核对手工修改（如 `mf-model-*/cm-kw-yaml.env.tmpl`、`dm_svc.conf.tmpl`，
以及上文提到的 `__KEY__` 占位符）。

### Manifest 工具

```bash
python3 tools/manifest.py services --phase=app    # 服务列表（compose.sh 使用）
python3 tools/manifest.py charts                  # <chart-folder>\t<out_sub>（extract 使用）
python3 tools/manifest.py env-defaults            # tagEnv 默认值（setup.sh 使用）
python3 tools/manifest.py check-compose           # manifest 与 docker-compose.yml 对账
python3 tools/manifest.py check-env [path]        # manifest 与 .env 中的 tagEnv 对账
```

## 远程实验环境（例如 Ubuntu VM）

将本目录同步到目标主机后：

```bash
cd deploy/docker-compose
./setup.sh -p YOUR_PASSWORD -y
sudo docker compose up -d   # 如果 Docker daemon 需要 sudo
```

按上面的冒烟检查验证 `curl http://127.0.0.1:8080/healthz` 与后端路由。

---

Kubernetes 部署请参考 [../README.zh.md](../README.zh.md)。
