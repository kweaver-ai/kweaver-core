# KWeaver Core/DIP Umbrella Chart 实施方案

**目标：** 交付 `kweaver-core`、`kweaver-dip` 两个 Helm 父 chart，兼容三种部署场景，实现一条命令安装产品包，不依赖 shell 部署脚本；其中 `kweaver-dip` 条件依赖 `kweaver-core`。

**三种部署场景：**

| 场景 | 前置条件 | 安装方式 |
|------|----------|----------|
| **A. 在线标准** | 客户已有 K8s + 兼容数据服务（MariaDB/Redis/…） | `helm install kweaver-dip oci://ghcr.io/kweaver-ai/kweaver-dip` |
| **B. 裸机 + Proton** | 客户只有裸机 | Proton 部署基础设施 → `helm install kweaver-dip` |
| **C. 离线环境** | 无外网访问 | Proton 离线版 + `kweaver-dip` 离线版（OCI → 私有 Harbor） |

**架构：** 通过 OCI Registry（`ghcr.io`）发布 chart，产品级 chart（core/dip）与组件级子 chart 通过不同的 OCI 路径自然隔离。父 chart 源码放在 `kweaver/deploy/charts`，其中 `kweaver-dip` 在依赖中显式引入 `kweaver-core`（条件依赖），并在打包时将依赖 vendor 到发布包。基础设施（K8s、数据服务）由独立的 `proton` 项目管理，不纳入 umbrella chart。

**技术栈：** Helm 3（OCI native）、Kubernetes、GitHub Actions、GitHub Container Registry（`ghcr.io`）、YAML anchors/aliases、Chart `dependencies + condition`。

---

## 1. 验收基线与边界

### 1.1 验收口径

- 用户在编辑配置后，可一条命令安装 `kweaver-core` 或 `kweaver-dip`。
- 验收路径不使用 `deploy.sh` 或任何自定义 shell 包装器。
- 需支持非项目成员按文档独立完成部署。

### 1.2 本方案边界

- 包含：
  - `kweaver-core`、`kweaver-dip` 的 chart 设计、发布、索引隔离、分支包策略。
  - 单一 values 文件设计与模块开关。
  - 现有子 chart 兼容 global 变量场景。
- 不包含：
  - 基础设施（MariaDB/Redis/Kafka/OpenSearch/MongoDB 等）生命周期自动化。

## 2. 目标状态设计

### 2.1 部署架构与职责边界

```text
┌─────────────────────────────────────────────────────────┐
│                    用户环境                               │
│                                                         │
│  ┌────────────────────────────────────────────────────┐ │
│  │              kweaver-dip (Helm Chart)              │ │
│  │   Studio · ModelFactory · 业务子系统 · ...           │ │
│  └───────────────────────┬────────────────────────────┘ │
│                          │ 依赖                         │
│  ┌───────────────────────▼────────────────────────────┐ │
│  │              kweaver-core (Helm Chart)             │ │
│  │  ISF · AgentOperator · DataAgent · Ontology · ...  │ │
│  └───────────────────────┬────────────────────────────┘ │
│                          │ 运行于 / 连接至              │
│  ┌───────────────────────▼────────────────────────────┐ │
│  │                  基础设施层                        │ │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐│ │
│  │  │  Kubernetes  │ │   MariaDB    │ │  Redis/Kafka ││ │
│  │  │  (必需)      │ │   (必需)     │ │  OpenSearch… ││ │
│  │  └──────────────┘ └──────────────┘ └──────────────┘│ │
│  │       ▲ 场景A: 客户已有                            │ │
│  │       ▲ 场景B/C: 由 Proton 自动化部署              │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

**职责划分：**
- **Proton 项目**（`/code/kweaver/proton`）：负责 K8s 集群初始化 + 数据服务部署（MariaDB/Redis/Kafka/OpenSearch/MongoDB/ZooKeeper/Ingress）。在场景 B/C 下使用。
- **Umbrella Chart**（`kweaver-core` / `kweaver-dip`）：仅包含业务服务，不包含基础设施。假定 K8s + 数据服务已就绪。
- **deploy/ 目录**：仅保留 umbrella chart 源码 + values 配置 + 部署文档。原有的 `deploy.sh`、`scripts/`、`sql/` 不再作为发布路径（由 Proton 替代）。

### 2.2 通过 OCI Registry 路径隔离

使用 GitHub Container Registry（`ghcr.io`）发布 Helm chart，通过 OCI 路径自然隔离产品级与组件级：

```text
ghcr.io/kweaver-ai/
  kweaver-core:<ver>           # 产品级（对外）
  kweaver-dip:<ver>            # 产品级（对外）
  charts/
    agent-backend:<ver>        # 组件级（内部依赖）
    isfweb:<ver>
    deploy-web:<ver>
    ...
```

说明：
- OCI 模式不需要 `index.yaml`、不需要 `gh-pages` 分支。
- 产品级 chart 直接在 `ghcr.io/kweaver-ai/` 根路径，组件级子 chart 在 `ghcr.io/kweaver-ai/charts/` 子路径。
- Preview 包通过 OCI tag 区分：`kweaver-core:0.0.0-pr.123.abc1234`。

### 2.3 用户视角

- 对外文档只暴露产品级 OCI 路径：
  - `helm install kweaver-core oci://ghcr.io/kweaver-ai/kweaver-core --version <ver>`
  - `helm install kweaver-dip oci://ghcr.io/kweaver-ai/kweaver-dip --version <ver>`
- 用户无需 `helm repo add`，OCI 原生支持直接拉取。
- 子 chart 路径（`ghcr.io/kweaver-ai/charts/`）不在对外文档中出现。

## 3. 服务归属（待确认）

### 3.1 `kweaver-core`（40 个子 chart）

- ISF（17）：
  - `hydra`、`sharemgnt-single`、`user-management`、`sharemgnt`、`authentication`、`policy-management`、`audit-log`、`eacp`、`thirdparty-message-plugin`、`isfwebthrift`、`message`、`isfweb`、`authorization`、`news-feed`、`ingress-informationsecurityfabric`、`eacp-single`、`oauth2-ui`
- AgentOperator（4）：
  - `agent-operator-integration`、`operator-web`、`agent-retrieval`、`data-retrieval`
- DataAgent（9）：
  - `agent-backend`、`agent-web`、`agent-app`、`agent-executor`、`agent-factory`、`agent-memory`、`coderunner`、`dataflow`、`agent-operator-app`
- Ontology（9）：
  - `ontology-manager`、`ontology-query`、`vega-web`、`data-connection`、`vega-gateway`、`vega-gateway-pro`、`mdl-data-model`、`mdl-uniquery`、`mdl-data-model-job`
- SandboxRuntime（1）：
  - `sandbox`

### 3.2 `kweaver-dip`（7 个子 chart + 强依赖 `kweaver-core`）

- 业务子 chart：
  - `deploy-web`、`studio-web`、`business-system-frontend`、`business-system-service`、`mf-model-manager-nginx`、`mf-model-manager`、`mf-model-api`
- 父包强依赖：
  - `kweaver-core`

### 3.3 基础服务排除项（由 Proton 管理）

- `mariadb`、`redis`、`mongodb`、`kafka`、`opensearch`、`zookeeper`、`ingress-nginx`
- 不纳入 core/dip 父 chart，由 Proton 项目（`/code/kweaver/proton`）独立管理。
- 场景 A 用户自行提供兼容数据服务；场景 B/C 由 Proton 自动部署。

## 4. 仓库职责与目录

### 4.1 三仓库职责划分

| 仓库 | 当前职责 | 目标职责 |
|------|----------|----------|
| **子服务仓库**（N 个） | 构建 `.tgz` → 手动提交到 `helm-repo/packages/` | 构建 `.tgz` → CI 自动 `helm push` 到 OCI |
| **kweaver 主仓库** | `deploy.sh` + 20 个 shell 脚本 + SQL | umbrella chart 源码 + values + 部署文档 |
| **helm-repo** | `.tgz` 中转站 + gh-pages 索引生成 | **归档只读**（历史 packages 保留，不再接受新提交） |
| **proton** | — | K8s 初始化 + 数据服务部署（场景 B/C） |

### 4.2 kweaver 主仓库 `deploy/` 目录结构

```text
deploy/
├── charts/
│   ├── kweaver-core/                # umbrella chart 源码
│   │   ├── Chart.yaml
│   │   ├── Chart.lock
│   │   ├── values.yaml
│   │   ├── values.schema.json
│   │   └── templates/
│   │       ├── db-init-job.yaml
│   │       └── tests/
│   │           └── test-connection.yaml
│   └── kweaver-dip/                 # umbrella chart 源码
│       ├── Chart.yaml
│       ├── Chart.lock
│       ├── values.yaml
│       ├── values.schema.json
│       └── templates/
│           ├── db-init-job.yaml
│           └── tests/
│               └── test-connection.yaml
├── conf/
│   └── products-values.yaml         # 统一 values 文件
├── helmfile.yaml                    # 声明式部署（Phase 2）
├── argocd/
│   └── appset.yaml                  # GitOps 示例（Phase 2）
├── README.md                        # 更新为 OCI 安装命令
└── README.zh.md
```

说明：
- `deploy/charts/` 仅保留 `kweaver-core`、`kweaver-dip` 两个 umbrella chart 源码。
- 原有的 `deploy.sh`、`scripts/`、`scripts/sql/`、`auto_cofig/`、`charts/*.tgz`（基础设施包）不再作为发布路径，由 Proton 项目替代。
- `deploy/conf/` 仅保留 `products-values.yaml`，原有的 `config.yaml`、`kube-flannel.yml` 等基础设施配置由 Proton 管理。

### 4.3 helm-repo 仓库归档

`helm-repo` 仓库标记为**归档只读**：

- 保留历史 `packages/` 供旧版本兼容使用。
- `README.md` 更新为指向 OCI Registry 的迁移说明。
- 停止接受新 `.tgz` 提交。
- GitHub Actions workflow 可选保留（用于历史版本的 gh-pages 索引），但不作为新版本发布路径。


## 5. Umbrella Chart 设计细则

### Task 2：定义依赖与模块开关

**Files:**
- Modify: `deploy/charts/kweaver-core/Chart.yaml`
- Modify: `deploy/charts/kweaver-dip/Chart.yaml`
- Modify: `deploy/charts/kweaver-core/values.yaml`
- Modify: `deploy/charts/kweaver-dip/values.yaml`

**Step 1: 依赖来源规则**

`kweaver-core/Chart.yaml` 中业务依赖统一用：

```yaml
dependencies:
  - name: agent-backend
    version: "0.3.4"
    repository: "oci://ghcr.io/kweaver-ai/charts"
    condition: modules.dataagent.enabled
```

`kweaver-dip/Chart.yaml` 中需显式加入 `kweaver-core` 条件依赖：

```yaml
dependencies:
  - name: kweaver-core
    version: "0.1.0"
    repository: "oci://ghcr.io/kweaver-ai"
    condition: kweaver-core.enabled      # 默认 true，已有 core 的用户可设 false
  - name: deploy-web
    version: "0.3.0"
    repository: "oci://ghcr.io/kweaver-ai/charts"
    condition: modules.studio.enabled
```

说明：
- `kweaver-core` 作为 `kweaver-dip` 的依赖，通过 `condition: kweaver-core.enabled` 控制，**默认 `true`**（新用户一条命令装全套）。
- 已单独安装 `kweaver-core` 的用户，设 `--set kweaver-core.enabled=false` 避免资源重复部署。
- 此模式对齐 `kube-prometheus-stack` 管理内嵌 Grafana 的做法：默认内嵌，允许外部替代。
- 业务子 chart 继续从 `components` 索引解析，避免对子 chart 使用本地散落路径。
- `condition` 挂到模块级开关，不让用户逐个 chart 管控。

**Step 2: 按模块设置 `condition` 键**

建议模块开关：
- `modules.isf.enabled`
- `modules.agentoperator.enabled`
- `modules.dataagent.enabled`
- `modules.ontology.enabled`
- `modules.sandboxruntime.enabled`
- `modules.modelfactory.enabled`

**Step 3: 默认策略**

- `kweaver-core/values.yaml`: core 相关模块默认 `true`
- `kweaver-dip/values.yaml`: `kweaver-core.enabled: true`（默认连带安装 core），studio 相关默认 `true`
- `kweaver-dip` 安装时默认会连带安装 `kweaver-core`；已有 core 的环境可设 `kweaver-core.enabled=false` 跳过。
- 用户可基于资源情况关闭可选模块并保持一条命令安装。

**三种典型安装场景：**

| 场景 | 命令 | 行为 |
|------|------|------|
| 新用户一键装全套 | `helm install kweaver-dip ...` | core + dip 一起安装（默认） |
| 已有 core，补装 dip | `helm install kweaver-dip --set kweaver-core.enabled=false` | 只装 dip 的 7 个子 chart |
| 仅安装 core | `helm install kweaver-core ...` | 独立安装 core |

## 6. 单一 values 文件策略（无脚本）

### Task 3：建立 `products-values.yaml`

**Files:**
- Create: `deploy/conf/products-values.yaml`

**设计决策：** 使用 Helm 原生 `global` 块向所有子 chart 透传通用配置，而非 YAML 锚点。详见 Section 17 的架构决策说明——锚点方案因生态工具不兼容（ArgoCD/Rancher UI 会展开锚点）而放弃。

**Step 1: 定义 `global` 通用配置块**

```yaml
global:
  image:
    registry: "swr.cn-east-3.myhuaweicloud.com/kweaver-ai"
  imagePullSecrets: []
  replicaCount: 1
  mode: Community
  storageClass: "local-path"
  ingressClassName: "nginx"
  env:
    language: en_US.UTF-8
    timezone: Asia/Shanghai
  accessAddress:
    host: 10.4.175.152
    port: 443
    scheme: https
    path: /
  depServices:
    rds:
      host: 'mariadb.resource.svc.cluster.local'
      port: 3306
      user: 'adp'
      password: ''
      database: 'adp'
      type: MariaDB
    redis:
      connectType: sentinel
      connectInfo:
        sentinelHost: proton-redis-sentinel.resource.svc.cluster.local
        sentinelPort: 26379
        password: ''
    mq:
      mqType: kafka
      mqHost: 'kafka.resource.svc.cluster.local'
      mqPort: 9092
    opensearch:
      host: 'opensearch-cluster-master.resource.svc.cluster.local'
      port: 9200
      protocol: https
    mongodb:
      host: 'mongodb.resource.svc.cluster.local'
      port: 28000
```

说明：
- Helm 会自动将 `global` 块透传给所有子 chart，无需逻一对每个子 chart 复制配置。
- 子 chart 模板通过 Section 17 的“Global + Templates 降级”模式读取：优先取 `global.xxx`，若为空则回退到局部 `.Values.xxx`。

**Step 2: 暴露模块开关**

```yaml
modules:
  isf: { enabled: true }
  agentoperator: { enabled: true }
  dataagent: { enabled: true }
  ontology: { enabled: true }
  sandboxruntime: { enabled: true }
  modelfactory: { enabled: true }
```

**Step 3: 特定子 chart 覆盖（可选）**

Helm 的 values 传递规则：

| 层级 | 传递方式 | 优先级 |
|------|----------|--------|
| `global.*` | Helm **自动透传**给所有子 chart，子 chart 通过 `.Values.global.*` 访问 | 最低（被子 chart 局部值覆盖） |
| `<subchart-name>.*` | Helm 将该键下的所有字段**注入**为子 chart 的 `.Values.*` | 中等 |
| `modules.*` | **不透传**，仅父 chart 自己读取，用于 dependency condition 判断 | — |

在 `products-values.yaml` 中，如果某个子 chart 需要与全局不同的配置，直接在**顶层以子 chart 名称为键**写入覆盖值：

```yaml
# products-values.yaml

global:
  image:
    registry: "swr.cn-east-3.myhuaweicloud.com/kweaver-ai"
  replicaCount: 1
  depServices:
    rds:
      host: 'mariadb.resource.svc.cluster.local'
      # ...

modules:
  isf: { enabled: true }
  # ...

# --- 以下为可选的子 chart 级覆盖 ---
# 可以覆盖子 chart values.yaml 中的任意字段，不限于 global 中定义的键。

# 例 1：agent-backend 需要 3 副本（覆盖 global.replicaCount=1）
agent-backend:
  replicaCount: 3

# 例 2：isfweb 使用不同的镜像源（覆盖 global.image.registry）
isfweb:
  image:
    registry: "harbor.internal/kweaver"

# 例 3：dataflow 连接专用数据库（覆盖 global.depServices.rds）
dataflow:
  depServices:
    rds:
      host: 'dataflow-dedicated-db.svc.cluster.local'
      database: 'dataflow'

# 例 4：覆盖子 chart 的任意字段（resources、nodeSelector、extraEnv 等）
# 这些字段不在 global 中，但存在于子 chart 自己的 values.yaml 中
agent-backend:
  resources:
    limits:
      cpu: "4"
      memory: 8Gi
    requests:
      cpu: "1"
      memory: 2Gi
  nodeSelector:
    gpu: "true"
  tolerations:
    - key: "gpu"
      operator: "Exists"
      effect: "NoSchedule"
  extraEnv:
    - name: LOG_LEVEL
      value: debug
```

> **说明：** `<subchart-name>.*` 注入机制可以覆盖子 chart `values.yaml` 中的**任意字段**，不仅限于 `global` 中定义的键。只要子 chart 的 `values.yaml` 中声明了该字段（如 `resources`、`nodeSelector`、`tolerations`、`extraEnv`、`serviceAccount` 等），父 chart 就可以通过这个键进行覆盖。Helm 会对这些值做 deep merge，将父 chart 指定的值与子 chart 的默认值合并。

**优先级解析顺序（子 chart 模板视角）：**

```text
子 chart 模板中的 .Values.replicaCount 解析优先级：

1. --set agent-backend.replicaCount=5         ← 命令行最高优先
2. products-values.yaml 中 agent-backend.replicaCount: 3  ← 子 chart 键覆盖
3. 子 chart 自己的 values.yaml 中 replicaCount: 1   ← 子 chart 默认值

子 chart 模板中的 .Values.global.depServices.rds 解析优先级：

1. --set global.depServices.rds.host=xxx      ← 命令行最高优先
2. products-values.yaml 中 global.depServices.rds.host    ← values 文件
3. 父 chart values.yaml 中 global.depServices.rds.host   ← 父 chart 默认值
```

> **设计要点：** `global` 是 Helm 内置的魔术键，自动向下透传。而 `<subchart-name>` 键也是 Helm 内置的注入机制——父 chart values 中 `agent-backend.replicaCount: 3` 会被 Helm 自动注入为 `agent-backend` 子 chart 的 `.Values.replicaCount`。这两个机制是 Helm 原生能力，无需额外代码。

**Step 4: 嵌套依赖下的子 chart 配置覆盖**

当 `kweaver-dip` 内嵌 `kweaver-core` 时，存在三层嵌套：`kweaver-dip → kweaver-core → 子 chart（如 hydra）`。

Helm 的值注入规则按**嵌套键路径**逍归传递：

```yaml
# products-values.yaml — 通过 kweaver-dip 安装时自定义 core 的子 chart

global:
  depServices:
    rds:
      host: 'mariadb.resource.svc.cluster.local'
  # global 自动逍归透传到所有层级，无需嵌套键

# 自定义 hydra（core 的子 chart）— 需要嵌套键路径
kweaver-core:
  hydra:
    replicaCount: 2
    resources:
      limits:
        cpu: "2"
        memory: 4Gi

  # 自定义 isfweb（core 的子 chart）
  isfweb:
    replicaCount: 3

  # 自定义 agent-backend（core 的子 chart）
  agent-backend:
    resources:
      limits:
        cpu: "4"
        memory: 8Gi
    nodeSelector:
      gpu: "true"

# 自定义 deploy-web（dip 的直接子 chart）— 无需嵌套
deploy-web:
  replicaCount: 2
```

**不同安装方式下的键路径对比：**

| 安装方式 | 自定义 hydra 的键 | 命令行 `--set` |
|----------|-----------------|------------------|
| 直接装 core | `hydra.replicaCount: 2` | `--set hydra.replicaCount=2` |
| 通过 dip 装全套 | `kweaver-core.hydra.replicaCount: 2` | `--set kweaver-core.hydra.replicaCount=2` |

> **关键区别：** `global.*` 无论嵌套多深都自动穿透所有层级（无需嵌套键），而 `<subchart-name>.*` 只注入到直接子 chart，不自动透传。所以通过 dip 自定义 core 的子 chart 时必须用 `kweaver-core.<subchart>.*` 路径。

Expected: 用户只改这一个 values 文件即可安装 core 或 dip。

## 7. 发布流水线

### Task 4：子 chart 自动打包推送到 OCI

**Files:**
- Modify: 各子项目仓库 CI（外部仓库，本文仅定义接口）

**Step 1: 子项目构建并自动推送到 OCI**

规范：
- 版本号使用稳定 semver 或 semver+build 后缀（如 `0.3.0-abc.1`）。
- 成功构建后自动推送到 OCI 组件路径。

Run in CI:
```bash
helm package ./charts/<subchart>
helm push <subchart>-<ver>.tgz oci://ghcr.io/kweaver-ai/charts
```

**Step 2: 路径隔离保证**

- 子 chart 推送到 `oci://ghcr.io/kweaver-ai/charts/<subchart>:<ver>`（组件级路径）。
- 产品级 chart（core/dip）推送到 `oci://ghcr.io/kweaver-ai/<name>:<ver>`（根路径）。
- 用户文档只引用根路径，不暴露 `charts/` 子路径。

### Task 5：子服务仓库 OCI CI 模板

**Files:**
- Create: 各子服务仓库 `.github/workflows/chart-release.yml`

**Step 1: 标准化子服务 CI（每个子服务仓库添加）**

```yaml
# .github/workflows/chart-release.yml
name: Release Chart to OCI
on:
  push:
    tags: ['v*']
jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      packages: write
    steps:
      - uses: actions/checkout@v4
      - uses: azure/setup-helm@v4

      - name: Login to GHCR
        run: echo "${{ secrets.GITHUB_TOKEN }}" | helm registry login ghcr.io -u ${{ github.actor }} --password-stdin

      - name: Package & Push to OCI
        run: |
          helm package ./charts/<subchart-name>
          helm push <subchart-name>-*.tgz oci://ghcr.io/kweaver-ai/charts
```

**Step 2: 迁移策略**

- 逐个子服务仓库添加此 CI，无需一次全部改造。
- 添加 CI 后，该子服务不再需要手动提交 `.tgz` 到 `helm-repo`。
- 全部子服务迁移完成后，`helm-repo` 正式归档。

Expected: 子服务 Git tag 触发 → CI 自动构建 → OCI 推送完成（无人工中转）。

### Task 6：kweaver 仓库 Umbrella Release CI

**Files:**
- Create: `/code/kweaver/kweaver/.github/workflows/umbrella-release.yml`

**Step 1: Umbrella Chart 发布 Workflow**

```yaml
# .github/workflows/umbrella-release.yml
name: Release Umbrella Charts
on:
  push:
    tags: ['v*']
    paths:
      - 'deploy/charts/kweaver-core/**'
      - 'deploy/charts/kweaver-dip/**'
  workflow_dispatch: {}

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      packages: write
    steps:
      - uses: actions/checkout@v4
      - uses: azure/setup-helm@v4

      - name: Login to GHCR
        run: echo "${{ secrets.GITHUB_TOKEN }}" | helm registry login ghcr.io -u ${{ github.actor }} --password-stdin

      - name: Build & Push kweaver-core
        run: |
          helm dependency update deploy/charts/kweaver-core
          helm package deploy/charts/kweaver-core -d /tmp/pkg
          helm push /tmp/pkg/kweaver-core-*.tgz oci://ghcr.io/kweaver-ai

      - name: Build & Push kweaver-dip
        run: |
          helm dependency update deploy/charts/kweaver-dip
          helm package deploy/charts/kweaver-dip -d /tmp/pkg
          helm push /tmp/pkg/kweaver-dip-*.tgz oci://ghcr.io/kweaver-ai
```

说明：
- Helm 3.8+ 原生支持从 OCI 拉取依赖，`helm dependency update` 会自动解析 `oci://` repository。
- `kweaver-core` 先打包推送，再打包推送 `kweaver-dip`（顺序打包保证依赖正确）。
- 触发条件：Git tag `v*` 或 chart 源码变更。

**Step 2: 本地手动构建（开发调试用）**

```bash
helm dependency update /code/kweaver/kweaver/deploy/charts/kweaver-core
helm dependency update /code/kweaver/kweaver/deploy/charts/kweaver-dip

helm package /code/kweaver/kweaver/deploy/charts/kweaver-core -d /tmp/pkg
helm push /tmp/pkg/kweaver-core-*.tgz oci://ghcr.io/kweaver-ai

helm package /code/kweaver/kweaver/deploy/charts/kweaver-dip -d /tmp/pkg
helm push /tmp/pkg/kweaver-dip-*.tgz oci://ghcr.io/kweaver-ai
```

Expected: OCI 推送后用户可立即拉取最新版本。

## 8. Feature/分支包策略

### Task 7：建立 preview 通道（可选但建议）

**Files:**
- Modify: 各子服务仓库 CI 及 kweaver 仓库 umbrella release CI

**Step 1: 版本规范与获取方式**

- 稳定版本 tag：`0.1.0`、`0.2.0`（纯 semver）

针对 Preview（分支/PR）包，区分两类仓库的获取策略：

**A. 子服务仓库（组件级）**
为了简化 CI，推荐使用 `0.0.0` 作为固定的基础版本前缀。
- PR 构建：`0.0.0-pr.<PR号>.<短SHA>` (例如: `0.0.0-pr.123.abc1234`)
- 分支构建：`0.0.0-<分支名>.<短SHA>` (例如: `0.0.0-feature-foo.abc1234`)

*CI 提取示例（子服务）：*
```yaml
      - name: Calculate Subchart Version
        id: calc_version
        run: |
          SHA=$(git rev-parse --short HEAD)
          if [[ "${{ github.event_name }}" == "pull_request" ]]; then
            echo "version=0.0.0-pr.${{ github.event.pull_request.number }}.${SHA}" >> $GITHUB_OUTPUT
          else
            SAFE_BRANCH=$(echo "${{ github.ref_name }}" | tr '/' '-')
            echo "version=0.0.0-${SAFE_BRANCH}.${SHA}" >> $GITHUB_OUTPUT
          fi
```

**B. KWeaver 主仓库（Umbrella 产品级）**
推荐从当前的 `Chart.yaml` 中动态提取 `<chart-version>` 作为基础前缀，保持产品语义连贯。
- PR 构建：`<chart-version>-pr.<PR号>.<短SHA>` (例如: `0.1.0-pr.123.abc1234`)
- 分支构建：`<chart-version>-<分支名>.<短SHA>` (例如: `0.1.0-feature-foo.abc1234`)

*CI 提取示例（Umbrella）：*
```yaml
      - name: Calculate Umbrella Version
        id: calc_version
        run: |
          CHART_VER=$(yq e '.version' deploy/charts/kweaver-core/Chart.yaml)
          SHA=$(git rev-parse --short HEAD)
          if [[ "${{ github.event_name }}" == "pull_request" ]]; then
            echo "version=${CHART_VER}-pr.${{ github.event.pull_request.number }}.${SHA}" >> $GITHUB_OUTPUT
          else
            SAFE_BRANCH=$(echo "${{ github.ref_name }}" | tr '/' '-')
            echo "version=${CHART_VER}-${SAFE_BRANCH}.${SHA}" >> $GITHUB_OUTPUT
          fi
```

**Step 2: OCI Tag 路由（无需索引分流）**

OCI 模式下，稳定版本与 preview 版本通过 **tag 语义** 自然隔离，无需单独的索引目录：

- 稳定版本 tag：`0.1.0`、`0.2.0`（纯 semver）
- Preview 版本 tag：`0.0.0-pr.123.abc1234`、`0.0.0-feature.xxx.def5678`（prerelease semver）

**Step 3: 部署使用**

- 联调环境（指定 preview tag）：
  ```bash
  helm install kweaver-core oci://ghcr.io/kweaver-ai/kweaver-core --version 0.0.0-pr.123.abc1234
  ```
- 生产环境只允许稳定通道（纯 semver tag）。

说明：
- OCI 天然支持多 tag 共存于同一路径，不需要 `kweaver-preview` 等额外目录。
- 子仓库分支名与 tag 名不需要一致，关联依据是版本号（PR 号 + SHA）。
- CI 可通过正则过滤，仅允许纯 semver tag 出现在生产文档中。

## 9. 数据库初始化 Hook 设计

### Task 9：为 core/dip 添加数据库 Migration Job

**Files:**
- Create: `deploy/charts/kweaver-core/templates/db-init-job.yaml`
- Create: `deploy/charts/kweaver-dip/templates/db-init-job.yaml`

**Step 1: Core 数据库初始化**

使用 Helm pre-install/pre-upgrade hook，在业务 Pod 启动前完成 schema migration：

```yaml
# kweaver-core/templates/db-init-job.yaml
{{- if .Values.dbInit.enabled }}
{{- $rds := .Values.global.depServices.rds }}
{{- $registry := .Values.global.image.registry | default "" }}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Release.Name }}-db-init
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-weight": "-5"
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
spec:
  backoffLimit: {{ .Values.dbInit.backoffLimit | default 3 }}
  template:
    spec:
      restartPolicy: Never
      initContainers:
        - name: wait-for-db
          image: busybox:1.36
          command: ['sh', '-c']
          args:
            - |
              until nc -z {{ $rds.host }} {{ $rds.port }}; do echo waiting for db; sleep 5; done
      containers:
        - name: migrate
          image: "{{ if $registry }}{{ $registry }}/{{ end }}{{ .Values.dbInit.image.repository }}:{{ .Values.dbInit.image.tag }}"
          env:
            - name: DB_HOST
              value: {{ $rds.host | quote }}
            - name: DB_PORT
              value: {{ $rds.port | quote }}
            - name: DB_USER
              value: {{ $rds.user | quote }}
            - name: DB_PASSWORD
              value: {{ $rds.password | quote }}
            - name: DB_NAME
              value: {{ $rds.database | default "adp" | quote }}
            - name: MIGRATION_TARGET
              value: "core"
{{- end }}
```

**Step 2: DIP 数据库初始化**

DIP 使用相同模式，但 `hook-weight` 更大（在 core 之后执行）：

```yaml
# kweaver-dip/templates/db-init-job.yaml
# 与 core 相同结构，但 hook-weight 更大（core=-5 先执行，dip=-3 后执行）
metadata:
  name: {{ .Release.Name }}-db-init
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-weight": "-3"
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
```

**Step 3: 执行顺序保证**

```text
hook-weight -10  →  check-db-ready（可选：验证数据库可达）
hook-weight -5   →  core-db-init（core schema migration）
hook-weight -3   →  dip-db-init（dip schema migration）
hook-weight 0    →  业务 Pod 正常启动
```

说明：
- `backoffLimit: 3` 提供重试能力，避免因数据库短暂不可用导致安装失败。
- `before-hook-creation` 策略确保重复安装/升级时先清理旧 Job。
- 升级时也会执行（`pre-upgrade`），适用于 schema migration 场景。

---

## 10. OCI Registry 架构说明

### Task 10：OCI 发布架构

**背景：** 本方案从 Day 1 采用 OCI Registry（`ghcr.io`）作为唯一发布通道。Helm 3.8+ 原生支持 OCI artifact，无需 `index.yaml` 或 `gh-pages` 分支。

**OCI 路径规划：**

```text
oci://ghcr.io/kweaver-ai/kweaver-core:<ver>        ← 产品级（对外文档暴露）
oci://ghcr.io/kweaver-ai/kweaver-dip:<ver>         ← 产品级（对外文档暴露）
oci://ghcr.io/kweaver-ai/charts/<subchart>:<ver>   ← 组件级（仅作为依赖引用）
```

**OCI 优势（相比传统 GitHub Pages + index.yaml）：**
- 不需要 `index.yaml` 和 `gh-pages` 分支维护
- 不需要 `helm repo add`，用户直接 `helm install oci://...`
- 与容器镜像同仓管理（chart + image 在同一 registry）
- 天然支持 public/private 权限控制
- Preview 包与稳定包通过 tag 语义隔离，无需额外目录
- 现成 GitHub Actions 支持：`appany/helm-oci-chart-releaser`

**离线/内网场景：**

对于无法访问 `ghcr.io` 的环境，可将 OCI chart 导入私有 registry：

```bash
# 从 ghcr.io 拉取
helm pull oci://ghcr.io/kweaver-ai/kweaver-core --version 0.1.0

# 推送到内网 Harbor
helm push kweaver-core-0.1.0.tgz oci://harbor.internal/kweaver
```

**Files:**
- 已在 Task 5（子服务 OCI CI）和 Task 6（kweaver umbrella release CI）中覆盖。

---

## 11. 多部署路径支持（Helmfile / GitOps）

### Task 11：提供声明式与 GitOps 部署路径

**背景：** 除了原生 `helm install`，业界主流项目还提供 Helmfile 和 Argo CD/Flux 集成示例，覆盖从个人开发者到企业级用户。

**Step 1: Helmfile 声明式部署（推荐 DevOps 团队使用）**

```yaml
# deploy/helmfile.yaml
repositories:
  - name: kweaver
    url: "ghcr.io/kweaver-ai"
    oci: true

releases:
  - name: kweaver-core
    namespace: kweaver
    chart: oci://ghcr.io/kweaver-ai/kweaver-core
    version: 0.1.0
    values:
      - conf/products-values.yaml
    wait: true
    timeout: 1800

  - name: kweaver-dip
    namespace: kweaver
    chart: oci://ghcr.io/kweaver-ai/kweaver-dip
    version: 0.1.0
    needs:
      - kweaver/kweaver-core          # 声明式依赖排序
    values:
      - conf/products-values.yaml
      - kweaver-core:
          enabled: false               # core 已由上面的 release 安装
    wait: true
    timeout: 1800
```

用户体验：
```bash
# 一条命令安装全部（core + dip 作为独立 release）
helmfile -f deploy/helmfile.yaml apply

# 只升级 dip
helmfile -f deploy/helmfile.yaml apply -l name=kweaver-dip
```

**Step 2: Argo CD ApplicationSet 示例（推荐 GitOps 团队使用）**

```yaml
# deploy/argocd/appset.yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: kweaver
  namespace: argocd
spec:
  generators:
    - list:
        elements:
          - name: kweaver-core
            chart: kweaver-core
            repoURL: "ghcr.io/kweaver-ai"
            version: "0.1.0"
            syncWave: "1"
          - name: kweaver-dip
            chart: kweaver-dip
            repoURL: "ghcr.io/kweaver-ai"
            version: "0.1.0"
            syncWave: "2"
  template:
    metadata:
      name: '{{name}}'
      annotations:
        argocd.argoproj.io/sync-wave: '{{syncWave}}'
    spec:
      project: default
      source:
        repoURL: '{{repoURL}}'
        chart: '{{chart}}'
        targetRevision: '{{version}}'
        helm:
          valueFiles:
            - values-prod.yaml
      destination:
        server: https://kubernetes.default.svc
        namespace: kweaver
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
```

**Step 3: 部署路径总览**

| 用户类型 | 推荐方式 | 命令/工具 |
|----------|----------|-----------|
| 新手 / 快速评估 | Helm CLI | `helm install kweaver-dip ...` |
| DevOps 团队 | Helmfile | `helmfile apply` |
| GitOps / 企业级 | Argo CD / Flux | `git push` → 自动同步 |

**Files:**
- Create: `deploy/helmfile.yaml`
- Create: `deploy/argocd/appset.yaml`（示例）

---

## 12. Schema 校验与 Helm Test

### Task 12：增强部署可靠性

**Step 1: values.schema.json（参数校验）**

为 `kweaver-core` 和 `kweaver-dip` 编写 JSON Schema，在 `helm install` 前自动校验参数：

```json
{
  "$schema": "https://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "modules": {
      "type": "object",
      "properties": {
        "isf":            { "type": "object", "properties": { "enabled": { "type": "boolean" } }, "required": ["enabled"] },
        "agentoperator":  { "type": "object", "properties": { "enabled": { "type": "boolean" } }, "required": ["enabled"] },
        "dataagent":      { "type": "object", "properties": { "enabled": { "type": "boolean" } }, "required": ["enabled"] },
        "ontology":       { "type": "object", "properties": { "enabled": { "type": "boolean" } }, "required": ["enabled"] },
        "sandboxruntime": { "type": "object", "properties": { "enabled": { "type": "boolean" } }, "required": ["enabled"] }
      }
    },
    "depServices": {
      "type": "object",
      "required": ["rds", "redis", "mq", "opensearch"],
      "properties": {
        "rds": {
          "type": "object",
          "required": ["host", "port", "user", "password"],
          "properties": {
            "host": { "type": "string" },
            "port": { "type": "integer", "minimum": 1, "maximum": 65535 }
          }
        }
      }
    }
  }
}
```

**Step 2: Helm Test Hook（部署后验证）**

```yaml
# kweaver-core/templates/tests/test-connection.yaml
apiVersion: v1
kind: Pod
metadata:
  name: {{ .Release.Name }}-test-connection
  annotations:
    "helm.sh/hook": test
spec:
  restartPolicy: Never
  containers:
    - name: test
      image: busybox:1.36
      command: ['sh', '-c']
      args:
        - |
          echo "=== KWeaver Core Connectivity Test ==="
          # 检查关键服务端口
          nc -z {{ .Release.Name }}-isfweb 80 && echo "isfweb: OK" || echo "isfweb: FAIL"
          nc -z {{ .Release.Name }}-agent-backend 80 && echo "agent-backend: OK" || echo "agent-backend: FAIL"
          echo "=== Test Complete ==="
```

验收：
```bash
helm test kweaver-core -n kweaver
helm test kweaver-dip -n kweaver
```

**Step 3: 回滚策略**

```bash
# 查看历史版本
helm history kweaver-core -n kweaver

# 回滚到上一个版本
helm rollback kweaver-core 1 -n kweaver

# 回滚 dip（不影响 core）
helm rollback kweaver-dip 1 -n kweaver
```

说明：因 core 和 dip 为独立 Helm release（即使 dip 内嵌 core 默认开启），各自有独立的 revision 历史，回滚互不影响。

**Files:**
- Create: `deploy/charts/kweaver-core/values.schema.json`
- Create: `deploy/charts/kweaver-dip/values.schema.json`
- Create: `deploy/charts/kweaver-core/templates/tests/test-connection.yaml`
- Create: `deploy/charts/kweaver-dip/templates/tests/test-connection.yaml`

---

## 13. 用户安装与验收命令

### Task 8：提供最终对外命令

**Files:**
- Modify: `deploy/README.md`
- Modify: `deploy/README.zh.md`

---

### 场景 A：在线标准（客户已有 K8s + 数据服务）

前置条件：K8s 集群就绪，MariaDB/Redis/Kafka/OpenSearch 等数据服务已部署并可连接。

**Step A1: 只装 core**

```bash
helm upgrade --install kweaver-core oci://ghcr.io/kweaver-ai/kweaver-core \
  --version 0.1.0 \
  -n kweaver --create-namespace \
  -f products-values.yaml \
  --wait --timeout 30m
```

**Step A2: 一条命令装全套（dip 含 core）**

```bash
helm upgrade --install kweaver-dip oci://ghcr.io/kweaver-ai/kweaver-dip \
  --version 0.1.0 \
  -n kweaver --create-namespace \
  -f products-values.yaml \
  --wait --timeout 30m
```

说明：默认 `kweaver-core.enabled=true`，一条命令安装 core + dip 全部业务服务。

**Step A3: 已有 core 补装 dip**

```bash
helm upgrade --install kweaver-dip oci://ghcr.io/kweaver-ai/kweaver-dip \
  --version 0.1.0 \
  -n kweaver \
  -f products-values.yaml \
  --set kweaver-core.enabled=false \
  --wait --timeout 30m
```

---

### 场景 B：裸机 + Proton（客户只有主机）

前置条件：客户只有一台或多台 Linux 主机，无 K8s 和数据服务。

**Step B1: 使用 Proton 部署基础设施**

```bash
# 安装 proton-cli
curl -sfL https://get.proton.kweaver.ai | sh

# 启动 Web 向导，初始化 K8s + 数据服务
proton-cli serve
# → 浏览器访问 http://<host>:3000 完成 K8s 集群初始化和数据服务部署
```

**Step B2: 使用 Helm 安装业务服务**

```bash
helm upgrade --install kweaver-dip oci://ghcr.io/kweaver-ai/kweaver-dip \
  --version 0.1.0 \
  -n kweaver --create-namespace \
  -f products-values.yaml \
  --wait --timeout 30m
```

---

### 场景 C：离线环境

前置条件：无外网访问，需要离线安装包。

**Step C1: 在联网机器上准备离线包**

```bash
# 拉取 Helm chart
helm pull oci://ghcr.io/kweaver-ai/kweaver-dip --version 0.1.0

# 拉取容器镜像（使用镜像列表脚本）
# 镜像列表由 CI 自动生成，包含所有子 chart 的 image 引用
docker save -o kweaver-images.tar $(cat image-list.txt | tr '\n' ' ')
```

**Step C2: 传输到离线环境并部署**

```bash
# 加载镜像到内网 registry 或本地 containerd
docker load -i kweaver-images.tar
# 或推送到私有 Harbor
for img in $(cat image-list.txt); do
  docker tag "$img" harbor.internal/kweaver/"$(basename $img)"
  docker push harbor.internal/kweaver/"$(basename $img)"
done

# 推送 chart 到私有 Harbor
helm push kweaver-dip-0.1.0.tgz oci://harbor.internal/kweaver

# 使用私有 registry 安装
helm upgrade --install kweaver-dip oci://harbor.internal/kweaver/kweaver-dip \
  --version 0.1.0 \
  -n kweaver --create-namespace \
  -f products-values.yaml \
  --set global.image.registry=harbor.internal/kweaver \
  --wait --timeout 30m
```

说明：场景 C 中基础设施由 Proton 离线版部署，Helm chart 和容器镜像通过私有 Harbor 分发。

---

### 验收校验（适用于全部场景）

```bash
# Pod 状态验证
kubectl get pods -n kweaver

# Helm release 验证
helm list -n kweaver

# 自动化连通性测试
helm test kweaver-core -n kweaver
helm test kweaver-dip -n kweaver
```

Expected:
- `helm list -n kweaver` 返回 `kweaver-core` 和/或 `kweaver-dip` release
- 模块按 values 开关启停，部署成功。
- `helm test` 通过所有连通性检查。

## 14. 风险与控制

- 风险：stable/preview 包混入。
  - 控制：OCI tag 语义隔离（纯 semver vs prerelease），CI 中正则过滤。
- 风险：子 chart 发布后 core/dip 未及时更新。
  - 控制：依赖更新流程自动化（机器人 PR 或定时 job）。
- 风险：用户误引用组件级 OCI 路径。
  - 控制：对外文档只提供产品级路径 `oci://ghcr.io/kweaver-ai/kweaver-core`，不暴露 `charts/` 子路径。
- 风险：先装 `kweaver-core` 再装 `kweaver-dip` 可能出现依赖重复部署。
  - 控制：`kweaver-core.enabled` 条件开关（默认 true），已有 core 的用户设 false；文档和 README 中明确说明。
- 风险：feature 包无限增长。
  - 控制：OCI tag 保留策略（GitHub Actions 定期清理 prerelease tag，如 14/30 天）。
- 风险：36 个子 chart 耦合导致 umbrella chart 发版阻塞。
  - 控制：建立版本兼容性矩阵，CI 自动化测试各子 chart 版本组合；提供 Helmfile 独立 release 路径作为备选。
- 风险：数据库初始化失败导致全量安装回滚。
  - 控制：hook Job 设 `backoffLimit: 3` 重试；initContainer 等待数据库就绪后再执行 migration。
- 风险：缺少参数校验导致部署失败排查困难。
  - 控制：`values.schema.json` 在 `helm install` 前自动校验参数类型和必填项。
- 风险：内网/离线环境无法访问 `ghcr.io`。
  - 控制：提供 `helm pull` + `helm push` 到私有 Harbor 的离线部署文档。



## 17. 附录：子 Chart 适配 Umbrella 改造指南

为了让各个独立的子服务 Chart 能够无缝作为 `kweaver-core` 或 `kweaver-dip` 的依赖运行，各子服务仓库的 Chart 需要进行以下适配改造：

### 1. 命名空间（Namespace）统一

**背景：** 在独立部署时，子 chart 经常在 `values.yaml` 中硬编码 `namespace`（如 `anyshare`、`dip`），并在 templates 中通过 `{{ .Values.namespace }}` 渲染。在 Umbrella 模式下，所有的组件都必须部署到用户执行 `helm install` 时指定的同一个命名空间（即 `-n kweaver`）。

**改造：** 
移除所有资源文件（`metadata.namespace`）中对 `{{ .Values.namespace }}` 的引用，**统一使用 Helm 内置变量 `{{ .Release.Namespace }}`**。

*Before:*
```yaml
metadata:
  name: isfweb
  namespace: {{ .Values.namespace }}
```

*After:*
```yaml
metadata:
  name: isfweb
  namespace: {{ .Release.Namespace }}
```
*(注：如果移除 `namespace` 字段，Helm 默认也会将其部署到 Release 的命名空间，但显式写 `{{ .Release.Namespace }}` 更严谨。)*

### 2. 双场景兼容：利用 Templates 降级读取 Global 变量

**背景与架构决策：** 
由于 KWeaver 的各个子 chart 需要同时满足两种部署形态：
1. **商业独立版本**：各子组件独立部署，各自的 `values.yaml` 中包含完整的局部配置（如 `depServices`，`replicaCount` 等），且不会传递 `global` 变量。
2. **开源 Umbrella 版本**：将 40+ 个组件作为 `kweaver-core` 或 `kweaver-dip` 的子 chart。为了提升 UI 部署的体验（无需用户重复配置 40 次密码等）并适配业界主流的生态工具，在父 chart 的 `values.yaml` 中统一提供 **`global`** 配置块进行向下透传。

**为什么放弃其他 Values 传递方案（如 YAML 锚点）？**
在设计之初，我们曾考虑过以下方案，但最终由于无法契合开源/商业交付生态而放弃：
1. **纯 YAML 锚点（YAML Anchors `&` 与 `<<`）**：
   - *方案与优势*：在顶层用锚点将 `depServices` 映射给所有子 chart，子 chart 可实现“零代码改造”。
   - *致命缺陷（生态工具不兼容）*：主流的云原生 UI 部署面板（如 Rancher, KubeSphere, ArgoCD 参数覆盖）在读取 chart 时，会将 YAML 解析为 JSON 树，这会导致所有锚点被瞬间**完全展开（Flatten）**。用户会在可视化界面上看到 40 份一模一样的数据库配置表单；同时该结构无法与 `values.schema.json` 配合做严格校验。
2. **引入 Helmfile 等编排工具**：
   - *缺陷*：引入了新的 CLI 依赖，增加了开源用户的学习成本，打破了“纯 Helm 原生 (`helm install`) 一键部署”的最佳体验。
3. **Kustomize 后置渲染（Post-renderer）**：
   - *缺陷*：依靠底层强行替换 k8s manifest（如 ConfigMap/Secret），面对 40 多个组件不同的字段结构，Patch 规则极其脆弱且难以维护。

**当前方案（Global + Templates 降级）的核心优势：**
- **UI 体验极佳**：由于通用配置全部收敛在顶层 `global` 中，用户在前端面板（如 ArgoCD UI）只需填写一次数据库密码，即可自动应用到所有组件。
- **100% Helm 原生**：零外部工具依赖，符合开源界（如 Bitnami, Prometheus 社区）处理巨型 Umbrella Chart 的最佳实践。
- **平滑双轨并行**：商业部署（无 global）继续读取原有 `values`，开源部署（有 global）实现全局统一接管。

**改造方案（改模板，不改 Values 结构）：**
由于 Helm 的限制，我们无法在子 chart 的 `values.yaml` 里直接引用 `global` 变量。必须在各子 chart 的**模板文件（Templates）**层面做兼容性修改：优先读取 `global`，若为空则降级（Fallback）读取局部变量。

> **注意：** 此降级模式仅适用于**子 chart 模板**（需同时支持独立部署和 Umbrella 部署）。Umbrella chart 自身的模板（如 `db-init-job.yaml`）可直接读取 `.Values.global.depServices`，因为它们仅在 Umbrella 模式下运行。

**改造示例（最佳实践）：**
建议在每个子 chart 的 `templates/_helpers.tpl` 中统一定义兼容函数，然后替换 deployment 等文件中的直接引用。

*1. 在 `_helpers.tpl` 中添加：*
```gotemplate
{{- /* 优先获取全局 registry，不存在则使用局部的 */ -}}
{{- define "kweaver.imageRegistry" -}}
{{- $globalImage := ( .Values.global | default dict ).image | default dict -}}
{{- coalesce $globalImage.registry .Values.image.registry -}}
{{- end -}}

{{- /* 优先获取全局 depServices 并与局部合并，避免全局遗漏某些字段 */ -}}
{{- define "kweaver.depServices" -}}
{{- $globalDeps := ( .Values.global | default dict ).depServices | default dict -}}
{{- if $globalDeps -}}
{{- toYaml (mergeOverwrite (deepCopy .Values.depServices) $globalDeps) -}}
{{- else -}}
{{- toYaml .Values.depServices -}}
{{- end -}}
{{- end -}}
```

*2. 在 `deployment.yaml` 等资源中调用：*
```gotemplate
      containers:
        - name: app
          # 使用 helper 渲染 registry
          image: "{{ include "kweaver.imageRegistry" . }}/{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          env:
            # 使用 helper 渲染 depServices，并反序列化为字典
            {{- $deps := include "kweaver.depServices" . | fromYaml }}
            - name: DB_HOST
              value: {{ $deps.rds.host | quote }}
```

**结论：** 
通过这样的改造，**商业版独立部署**时没有 `global` 输入，系统完美回退到现有的 `values`，没有任何影响；而在**开源 Umbrella 部署**时，用户只需在最外层填写一次 `global.depServices`，所有子图将自动继承这些全局通用配置，大幅提升了可视化部署面板中的操作体验和安全性。
