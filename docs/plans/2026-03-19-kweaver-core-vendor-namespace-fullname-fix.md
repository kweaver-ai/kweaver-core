# KWeaver Core Vendor Chart 改造示例

## 文档目的

### Chart改造完成情况

根据 `/code/DIP-295272626-200326-2058-7262.pdf` 中 “Chart改造完成情况” 表格整理如下：

| 模块 | Chart 名称 | Github 地址 | 改造已完成 | 备注 |
| --- | --- | --- | --- | --- |
| ISF | hydra | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/hydra/chart/hydra |  | 分支待推送 |
| ISF | sharemgnt-single | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/ShareMgnt/chart/sharemgnt-single |  | 分支待推送 |
| ISF | user-management | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/UserManagement/chart/user-management |  | 分支待推送 |
| ISF | sharemgnt | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/ShareMgnt/chart/sharemgnt |  | 分支待推送 |
| ISF | authentication | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/Authentication/chart/authentication |  | 分支待推送 |
| ISF | policy-management | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/PolicyManagement/chart/policy-management |  | 分支待推送 |
| ISF | audit-log | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/Auditlog/chart/audit-log |  | 分支待推送 |
| ISF | eacp | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/ShareServer/chart/eacp |  | 分支待推送 |
| ISF | isfwebthrift | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/ISFWeb/chart/isfwebthrift |  | 分支待推送 |
| ISF | isfweb | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/ISFWeb/chart/isfweb |  | 分支待推送 |
| ISF | authorization | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/Authorization/chart/authorization |  | 分支待推送 |
| ISF | ingress-informationsecurityfabric | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/Ingress_Proton_IdentifyAndAuthentication/ingress-informationsecurityfabric |  | 分支待推送 |
| ISF | oauth2-ui | https://github.com/kweaver-ai/isf/tree/feat/support-umbrella-chart/Authentication/oauth2-ui/chart/oauth2-ui |  | 分支待推送 |
| AgentOperator | agent-operator-integration | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/execution-factory/operator-integration/helm/agent-operator-integration |  | 分支已推送 |
| AgentOperator | operator-web | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/execution-factory/operator-web |  | 分支已推送 |
| AgentOperator | agent-retrieval | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/context-loader/agent-retrieval/helm/agent-retrieval |  | 分支已推送 |
| AgentOperator | data-retrieval |  |  | 移到了 DIP |
| flow automation | flow-web | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/dataflow/dia-flow-web/charts/flow-web |  | 分支已推送 |
| flow automation | dataflow | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/dataflow/charts/dataflow |  | 分支已推送 |
| flow automation | coderunner | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/dataflow/charts/coderunner |  | 分支已推送 |
| DataAgent | agent-backend | https://github.com/kweaver-ai/decision-agent/tree/feat/support-umbrella-chart/agent-backend/deploy/helm/agent-backend |  | 分支待推送 |
| DataAgent | agent-web | https://github.com/kweaver-ai/decision-agent/tree/feat/support-umbrella-chart/agent-web/charts/agent-web |  | 分支待推送 |
| Ontology | bkn-backend | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/bkn/bkn-backend/helm/bkn-backend |  | 原  ontology-manager；分支已推送 |
| Ontology | ontology-query | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/bkn/ontology-query/helm/ontology-query |  | 分支已推送 |
| Ontology | vega-backend | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/vega/vega-backend/helm/vega-backend |  | 分支已推送 |
| Ontology | vega-web | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/web/helm/vega-web |  | 分支已推送 |
| Ontology | data-connection | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/vega/data-connection/helm/data-connection |  | 分支已推送 |
| Ontology | vega-gateway | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/vega/vega-gateway/helm/vega-gateway |  | 分支已推送 |
| Ontology | vega-gateway-pro | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/vega/vega-gateway-pro/helm/vega-gateway-pro |  | 分支已推送 |
| Ontology | mdl-data-model | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/vega/mdl-data-model/helm/mdl-data-model |  | 分支已推送 |
| Ontology | mdl-uniquery | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/vega/mdl-uniquery/helm/mdl-uniquery |  | 分支已推送 |
| Ontology | mdl-data-model-job | https://github.com/kweaver-ai/adp/tree/feat/support-umbrella-chart/vega/mdl-data-model-job/helm/mdl-data-model-job |  | 分支已推送 |
| Sandbox | sandbox | https://github.com/kweaver-ai/sandbox/tree/feat/support-umbrella-chart/deploy/helm/sandbox |  | 分支待推送 |
| Studio | deploy-web | https://github.com/kweaver-ai/studio/tree/feat/support-umbrella-chart/deploy-web/charts/deploy-web |  | 分支待推送 |
| Studio | studio-web | https://github.com/kweaver-ai/studio/tree/feat/support-umbrella-chart/studio-web/charts/studio-web |  | 分支待推送 |
| Studio | business-system-frontend | https://github.com/kweaver-ai/studio/tree/feat/support-umbrella-chart/frontend/applications/business-system/charts/business-system |  | 分支待推送 |
| Studio | business-system-service | https://github.com/kweaver-ai/studio/tree/feat/support-umbrella-chart/business-system-backend/helm/business-system-service |  | 分支待推送 |
| Studio | mf-model-manager-nginx | https://github.com/kweaver-ai/studio/tree/feat/support-umbrella-chart/frontend/applications/model-manager/charts/model-manager |  | 分支待推送 |
| Studio | mf-model-manager | https://github.com/kweaver-ai/studio/tree/feat/support-umbrella-chart/mf-model-manager/charts |  | 分支待推送 |
| Studio | mf-model-api | https://github.com/kweaver-ai/studio/tree/feat/support-umbrella-chart/mf-model-api/charts |  | 分支待推送 |

这份文档不是按“执行步骤”来写，而是按“问题类别”来归纳这次 `kweaver-core` umbrella vendor chart 改造中实际做过的修改，方便其他团队参考类似场景下应该改哪里、为什么要这样改。

本次改造的核心目标有两个：

- 让真实发布产物形式的 `kweaver-core` umbrella chart 可以安装，不再出现固定 namespace 泄漏。
- 让多个子 chart 组合安装时，不再出现 Deployment、Service、ConfigMap、Secret、Ingress、Hook Job 等对象重名冲突。

补充说明：

- 本文中的“vendor chart”指的是 [subCharts](/code/kweaver/kweaver/deploy/charts/subCharts) 下展开后的依赖 chart。
- 文中同时包含“通用修复”和“仅用于测试验证的临时改动”两类内容，后者已单独标明。

---

## 问题类别一：chart 默认 `namespace` 值导致资源落到错误命名空间

### 为什么要这么改

很多原始 chart 里都带有类似 `namespace: dip`、`namespace: anyshare` 的默认值，并且模板里直接使用 `.Values.namespace` 渲染 `metadata.namespace`，或者用它去拼接同 namespace 下的服务 FQDN。

这在“单 chart 独立部署”时问题不明显，但放进 umbrella 里以后，所有资源本来都应该跟随安装命令的 release namespace。如果模板仍然使用 `.Values.namespace`，就会出现：

- release 安装在 `kweaver-vendorX`，但部分资源仍然落到 `dip` / `anyshare`
- 同一个 umbrella release 被拆散到多个 namespace
- 同 namespace 服务地址拼接错误，运行时访问落到错误 namespace

所以这类场景统一改成使用 `.Release.Namespace`，而不是继续信任 chart 自带的 `namespace` 默认值。

### 所有 vendor chart（批量处理）

#### 修改范围

对 [subCharts](/code/kweaver/kweaver/deploy/charts/subCharts) 下所有模板中出现的以下场景进行了统一替换：

- `metadata.namespace`
- `subjects[].namespace`
- 其他资源对象上的 namespace 字段
- 同 namespace 集群内服务 FQDN 拼接

#### 代表性 chart 与文件

以下只是代表性文件，实际是一次批量修复：

- [agent-app/templates/ingress.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-app/templates/ingress.yaml)
- [agent-factory/templates/ingress.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/ingress.yaml)
- [agent-operator-app/templates/ingress.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-operator-app/templates/ingress.yaml)
- [agent-operator-integration/templates/ingress.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-operator-integration/templates/ingress.yaml)
- [agent-retrieval/templates/ingress.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-retrieval/templates/ingress.yaml)
- [data-connection/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/data-connection/templates/deployment.yaml)
- [dataflow/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/dataflow/templates/deployment.yaml)
- [sandbox/templates/control-plane-deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/sandbox/templates/control-plane-deployment.yaml)
- [vega-gateway-pro/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/vega-gateway-pro/templates/deployment.yaml)

#### 实际改动

- 将模板中的 `.Values.namespace` 改为 `.Release.Namespace`
- 将通过 `.Values.namespace` 拼接的服务 DNS 名称改为 `.Release.Namespace`

#### 结果

重新渲染后，不再出现因为 chart 默认值而把资源落到 `dip` / `anyshare` 的情况。

---

## 问题类别二：多个子 chart 共用 `.Release.Name`，导致主资源对象重名

### 为什么要这么改

umbrella chart 下，所有 subchart 看到的 `.Release.Name` 都是同一个值，例如 `kweaver-core`。

如果多个 chart 都写成：

```yaml
metadata:
  name: {{ .Release.Name }}
```

那么它们最终会同时生成：

- `Deployment/kweaver-core`
- `Service/kweaver-core`
- `ConfigMap/kweaver-core-yaml`
- `Secret/kweaver-core`

这些资源在同一个 namespace 下必然冲突。

这类问题的修复原则是：

- 不再直接使用父 chart 的 `.Release.Name`
- 改用 chart 自己稳定且有区分度的名字
- 改名后，所有内部引用一起调整

### agent-app

#### 涉及文件

- [agent-app/templates/service.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-app/templates/service.yaml)
- [agent-app/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-app/templates/deployment.yaml)
- [agent-app/templates/configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-app/templates/configmap.yaml)
- [agent-app/templates/secret.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-app/templates/secret.yaml)

#### 改动内容

- `Service` 名称从 `{{ .Release.Name }}` 改为 `{{ .Values.moduleName }}`
- `Deployment` 名称、selector、pod label 从 `{{ .Release.Name }}` 改为 `{{ .Values.moduleName }}`
- `ConfigMap` 从 `{{ .Release.Name }}-yaml` 改为 `{{ .Values.moduleName }}-yaml`
- `Secret` 从 `{{ .Release.Name }}-secret-yaml` 改为 `{{ .Values.moduleName }}-secret-yaml`
- 同步修改 `configMapKeyRef.name`、volume 名称、`secretName`

#### 改动原因

否则 `agent-app` 会和其他同样使用 `.Release.Name` 命名的 chart 一起生成 `kweaver-core`、`kweaver-core-yaml`、`kweaver-core-secret-yaml`，直接冲突。

### agent-operator-app

#### 涉及文件

- [agent-operator-app/templates/service.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-operator-app/templates/service.yaml)
- [agent-operator-app/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-operator-app/templates/deployment.yaml)
- [agent-operator-app/templates/configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-operator-app/templates/configmap.yaml)
- [agent-operator-app/templates/secret.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-operator-app/templates/secret.yaml)

#### 改动内容

- `Service`、`Deployment`、`ConfigMap`、`Secret` 的主名称由 `.Release.Name` 改为 `moduleName`
- 同步修改 selector、volume、`secretName`、`configMap.name`

#### 改动原因

否则会和 `agent-operator-integration`、`agent-retrieval` 等 chart 一起生成 `kweaver-core` 这类对象。

### agent-operator-integration

#### 涉及文件

- [agent-operator-integration/templates/service.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-operator-integration/templates/service.yaml)
- [agent-operator-integration/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-operator-integration/templates/deployment.yaml)
- [agent-operator-integration/templates/configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-operator-integration/templates/configmap.yaml)
- [agent-operator-integration/templates/secret.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-operator-integration/templates/secret.yaml)

#### 改动内容

- `Service`、`Deployment`、`ConfigMap`、`Secret` 统一改为 `moduleName`
- 同步调整内部 volume 与 selector 引用

#### 改动原因

避免继续生成 `Service/Deployment/ConfigMap/Secret kweaver-core`。

### agent-retrieval

#### 涉及文件

- [agent-retrieval/templates/service.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-retrieval/templates/service.yaml)
- [agent-retrieval/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-retrieval/templates/deployment.yaml)
- [agent-retrieval/templates/configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-retrieval/templates/configmap.yaml)
- [agent-retrieval/templates/secret.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-retrieval/templates/secret.yaml)

#### 改动内容

- `Service`、`Deployment`、`ConfigMap`、`Secret` 统一改为 `moduleName`
- 同步调整 selector、volume、secret 与 configmap 引用

#### 改动原因

否则和其他多个 chart 一样，都会在 umbrella 下撞成 `kweaver-core`。

### data-retrieval

#### 涉及文件

- [data-retrieval/templates/service.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/data-retrieval/templates/service.yaml)
- [data-retrieval/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/data-retrieval/templates/deployment.yaml)
- [data-retrieval/templates/configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/data-retrieval/templates/configmap.yaml)

#### 改动内容

- 将 `Service`、`Deployment`、`ConfigMap` 名称从 `.Release.Name` 统一调整为 `.Chart.Name`
- 同步调整 `Deployment` 中所有 `configMapKeyRef.name`

#### 改动原因

`data-retrieval` 没有像其他 chart 那样明显可复用的 `moduleName` 体系，因此这里直接采用 chart 名称做稳定唯一标识。

---

## 问题类别三：legacy 聚合 chart 与拆分 chart 同时创建同名固定资源

### 为什么要这么改

这类冲突和“共享 `.Release.Name`”不是同一个问题。

这里的问题是：某些 chart 本身就在创建固定名字的资源，而 umbrella 里又同时引入了拆分后的独立 chart，导致双方在同一个 namespace 下创建同名对象。

最典型的就是：

- `agent-backend`
- `agent-factory`
- `agent-app`
- `agent-executor`
- `agent-memory`

在 umbrella 里是同时存在的，但 `agent-backend` 这个 legacy 聚合 chart 又会额外创建：

- `agent-app`
- `agent-factory`
- `agent-executor`
- `agent-memory`
- `agent-executor-yaml`
- `agent-memory-yaml`
- `agent-factory-ingress`

而拆分后的独立 chart 也会创建同名对象，于是直接冲突。

这类问题的处理原则是：

- 保留拆分后独立 chart 的 canonical 名字
- 将 legacy 聚合 chart 里那套重复资源整体加前缀，避免与独立 chart 争抢同名对象

### agent-factory

#### 涉及文件

- [agent-factory/templates/service.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/service.yaml)
- [agent-factory/templates/service-executor.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/service-executor.yaml)
- [agent-factory/templates/service-memory.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/service-memory.yaml)
- [agent-factory/templates/configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/configmap.yaml)
- [agent-factory/templates/configmap-executor.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/configmap-executor.yaml)
- [agent-factory/templates/configmap-memory.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/configmap-memory.yaml)
- [agent-factory/templates/configmap-executor-v2.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/configmap-executor-v2.yaml)
- [agent-factory/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/deployment.yaml)
- [agent-factory/templates/dependant-configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/dependant-configmap.yaml)
- [agent-factory/templates/secret.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/secret.yaml)
- [agent-factory/templates/ingress.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-factory/templates/ingress.yaml)

#### 改动内容

- 原先依赖 `.Release.Name` 的主资源，统一改为基于 `{{ .Values.moduleName }}` 命名
- 在当前 values 下，主 `Service` / `Deployment` 渲染结果为 `agent-factory`
- 附属 `Service` 渲染结果为：
  - `{{ .Values.moduleName }}-executor` -> `agent-factory-executor`
  - `{{ .Values.moduleName }}-memory` -> `agent-factory-memory`
- 配套 `ConfigMap` / `Secret` / 依赖配置对象渲染结果为：
  - `{{ .Values.moduleName }}-yaml` -> `agent-factory-yaml`
  - `{{ .Values.moduleName }}-executor-yaml` -> `agent-factory-executor-yaml`
  - `{{ .Values.moduleName }}-memory-yaml` -> `agent-factory-memory-yaml`
  - `{{ .Values.moduleName }}-executor-yaml-v2` -> `agent-factory-executor-yaml-v2`
  - `{{ .Values.moduleName }}-dependant-configmap` -> `agent-factory-dependant-configmap`
  - `{{ .Values.moduleName }}-secret-yaml` -> `agent-factory-secret-yaml`
- `Ingress` 保持为 `{{ .Values.moduleName }}-ingress`，当前渲染结果为 `agent-factory-ingress`
- `Deployment` 中的 `configMapKeyRef.name`、volume 名称、`secretName`、backend service 名称同步跟随 `moduleName` 体系调整

#### 改动原因

`agent-factory` 是拆分后的独立 chart，应该保留一套稳定、可识别的主命名体系，作为 umbrella 里的 canonical 资源集合。

### agent-backend

#### 涉及文件

- [agent-backend/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-backend/templates/deployment.yaml)
- [agent-backend/templates/service-app.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-backend/templates/service-app.yaml)
- [agent-backend/templates/service-factory.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-backend/templates/service-factory.yaml)
- [agent-backend/templates/service-executor.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-backend/templates/service-executor.yaml)
- [agent-backend/templates/service-memory.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-backend/templates/service-memory.yaml)
- [agent-backend/templates/configmap-factory.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-backend/templates/configmap-factory.yaml)
- [agent-backend/templates/configmap-executor.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-backend/templates/configmap-executor.yaml)
- [agent-backend/templates/configmap-memory.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-backend/templates/configmap-memory.yaml)
- [agent-backend/templates/secret.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-backend/templates/secret.yaml)
- [agent-backend/templates/ingress.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-backend/templates/ingress.yaml)

#### 改动内容

- 为 legacy 聚合资源整体增加 `agent-backend-` 前缀
- `Service` 改为：
  - `agent-backend-app`
  - `agent-backend-factory`
  - `agent-backend-executor`
  - `agent-backend-memory`
- `ConfigMap` 改为：
  - `agent-backend-factory-yaml`
  - `agent-backend-executor-yaml`
  - `agent-backend-memory-yaml`
- `Secret` 改为 `agent-backend-factory-secret-yaml`
- `Ingress` 改为 `agent-backend-factory-ingress`
- `Deployment` 名称与 pod label 改为 `agent-backend-factory`
- `Deployment` 中大部分 volume、`configMapKeyRef`、selector、service backend 已同步改名为 `agent-backend-*`
- 这里也暴露了一个很典型的检查点：对象名改完后，还要继续确认所有引用是否完全闭合。当前模板里 secret volume 的 `secretName` 仍保留旧值 `agent-factory-secret-yaml`，说明这类改造不能只看对象创建成功，还要继续检查引用链

#### 改动原因

`agent-backend` 本质上保留了老的一体化资源模型，如果继续和拆分 chart 使用相同固定名字，就一定会和独立 chart 正面冲突。这里采用统一前缀化，是为了在不删除 legacy 功能的前提下先让 umbrella 能组合安装。

---

## 问题类别四：资源改名后，引用这些资源的 chart 也必须一起改

### 为什么要这么改

有些 chart 自己的 Service / Deployment 名字没变，但它们引用的 ConfigMap、挂载卷、辅助配置文件名原先仍然依赖 `.Release.Name`。

如果只改对象名，不改引用，结果通常是：

- 模板能渲染
- Helm 能创建对象
- Pod 在运行时因为找不到 ConfigMap / Secret 而起不来

这类问题必须把“资源名”和“引用名”作为一组一起改。

### agent-executor

#### 涉及文件

- [agent-executor/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-executor/templates/deployment.yaml)
- [agent-executor/templates/configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-executor/templates/configmap.yaml)
- [agent-executor/templates/configmap-v2.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-executor/templates/configmap-v2.yaml)

#### 改动内容

- `ConfigMap` 从 `{{ .Release.Name }}-yaml` 改为 `{{ .Values.image.name }}-yaml`
- `ConfigMap v2` 从 `{{ .Release.Name }}-yaml-v2` 改为 `{{ .Values.image.name }}-yaml-v2`
- `Deployment` 中对应的 `configMapKeyRef.name`
- `Deployment` 中 `rds-conf` 与 `yaml-v2` volume 名称及其引用

#### 改动原因

`agent-executor` 本身已经用 `image.name` 做 Service / Deployment 的主命名，因此它内部引用也应该和 `image.name` 对齐，不能继续残留 `.Release.Name`。

### agent-memory

#### 涉及文件

- [agent-memory/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-memory/templates/deployment.yaml)
- [agent-memory/templates/configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-memory/templates/configmap.yaml)

#### 改动内容

- `ConfigMap` 从 `{{ .Release.Name }}-yaml` 改为 `{{ .Values.image.name }}-yaml`
- `Deployment` 中对应的 `configMapKeyRef.name`
- `Deployment` 中 `rds-conf` volume 名称与其引用

#### 改动原因

和 `agent-executor` 一样，这个 chart 的主资源命名本来就是 `image.name`，内部依赖必须与之保持一致。

---

## 问题类别五：多个 Web chart 的 Hook Job 名称相同

### 为什么要这么改

`agent-web`、`operator-web`、`vega-web` 三个 chart 都定义了：

- `post-install-job`
- `post-delete-job`

原始模板里这些 Job 直接使用 `.Release.Name` 命名，放进 umbrella 后就都会变成：

- `kweaver-core-post-install-job`
- `kweaver-core-post-delete-job`

这会导致 Hook Job 在安装或删除阶段直接冲突。

这类问题的修复原则是：

- 保留 Hook 逻辑不变
- 只把 Job 名字改成 chart 级唯一名

### agent-web

#### 涉及文件

- [agent-web/templates/post-install-job.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-web/templates/post-install-job.yaml)
- [agent-web/templates/post-delete-job.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/agent-web/templates/post-delete-job.yaml)

#### 改动内容

- Job 名称从依赖 `.Release.Name` 改为依赖 `{{ .Values.image.name }}`
- 当前 values 下：
  - `post-install` Job 为 `agent-web-post-install-job`
  - `post-delete` Job 为 `agent-web-post-delete-job`
- Pod template name 同步改为 `{{ .Values.image.name }}-post-install` / `{{ .Values.image.name }}-post-delete`

### operator-web

#### 涉及文件

- [operator-web/templates/post-install-job.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/operator-web/templates/post-install-job.yaml)
- [operator-web/templates/post-delete-job.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/operator-web/templates/post-delete-job.yaml)

#### 改动内容

- Job 名称从依赖 `.Release.Name` 改为依赖 `{{ .Values.image.name }}`
- 当前 values 下：
  - `post-install` Job 为 `operator-web-post-install-job`
  - `post-delete` Job 为 `operator-web-post-delete-job`
- Pod template name 同步改为 `{{ .Values.image.name }}-post-install` / `{{ .Values.image.name }}-post-delete`

### vega-web

#### 涉及文件

- [vega-web/templates/post-install-job.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/vega-web/templates/post-install-job.yaml)
- [vega-web/templates/post-delete-job.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/vega-web/templates/post-delete-job.yaml)

#### 改动内容

- Job 名称从依赖 `.Release.Name` 改为依赖 `{{ .Values.moduleName }}`
- 当前 values 下：
  - `post-install` Job 为 `vega-web-post-install-job`
  - `post-delete` Job 为 `vega-web-post-delete-job`
- Pod template name 同步改为 `{{ .Values.moduleName }}-post-install` / `{{ .Values.moduleName }}-post-delete`

#### 改动原因

Hook Job 冲突不会体现在常规 Deployment / Service 层面，但会直接让 Helm 在 hook 执行阶段报资源已存在错误，因此也必须纳入命名修复范围。

---

## 问题类别六：通过 `_helpers` 和模板预处理兼容 `mergedGlobalValues` 透传变量

### 为什么要这么改

很多子 chart 原先只读取自己 values 里的局部变量，例如：

- `.Values.image.registry`
- `.Values.depServices`
- `.Values.env`
- `.Values.replicaCount`

一旦放到 umbrella 下面，父 chart 更合理的做法通常是统一从 `mergedGlobalValues` 下透传，例如：

- `mergedGlobalValues.image.registry`
- `mergedGlobalValues.depServices`
- `mergedGlobalValues.env`
- `mergedGlobalValues.replicaCount`

如果子 chart 不兼容这套写法，就会出现几个问题：

- 父 chart 已经传了 `mergedGlobalValues.*`，但子 chart 仍然读不到
- 不同子 chart 对同一类参数要重复配置多份
- 镜像仓库、依赖服务地址、语言时区、副本数无法统一收敛

所以这一类改造的目标，是在不破坏子 chart 原有局部 values 用法的前提下，让它们同时兼容 `mergedGlobalValues` 透传。

统一处理原则是：

- 在模板开头先取出 `mergedGlobalValues` 变量
- 对 map 类型配置用 `mergeOverwrite` 做“局部值 + 全局值”合并
- 对标量值用 `coalesce` 做“优先全局、回退本地”选择
- 对镜像地址用 `_helpers.tpl` 统一封装，避免每个模板自己拼

### vega-web

#### 涉及文件

- [vega-web/templates/_helpers.tpl](/code/kweaver/kweaver/deploy/charts/subCharts/vega-web/templates/_helpers.tpl)
- [vega-web/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/vega-web/templates/deployment.yaml)
- [vega-web/templates/ingress.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/vega-web/templates/ingress.yaml)

#### 改动内容

- 在 [_helpers.tpl](/code/kweaver/kweaver/deploy/charts/subCharts/vega-web/templates/_helpers.tpl) 中增加 `mergedGlobalValues.image` 等通用 helper 函数
- helper 内使用：
  - `mergedGlobalValues.imageRegistry`: 优先读取 `global.image.registry`，未配置时再回退到 chart 自己的 `image.registry`
  - `mergedGlobalValues.replicaCount`: 使用 `hasKey` 检查 global 值，优先全局、回退本地
  - `mergedGlobalValues.env`: 使用 `mergeOverwrite` 深度合并全局和本地环境变量
  - `mergedGlobalValues.depServices`: 使用 `mergeOverwrite` 深度合并依赖服务配置
- 在 [deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/vega-web/templates/deployment.yaml) 中：
  - 用 `{{ include "mergedGlobalValues.image" . }}` 统一生成镜像地址
  - 用 `{{ include "mergedGlobalValues.replicaCount" . }}` 设置副本数
  - 在文件开头定义 `{{- $env := include "mergedGlobalValues.env" . | fromYaml -}}`
- 在 [ingress.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/vega-web/templates/ingress.yaml) 中：
  - 在文件开头定义 `{{- $depServices := include "mergedGlobalValues.depServices" . | fromYaml -}}`
  - 用 `{{ include "mergedGlobalValues.ingressClassName" . }}` 设置 ingress class

#### 改动原因

`vega-web` 既依赖镜像仓库统一覆盖，也依赖 `env`、`ingress class` 这类公共配置。如果不把读取逻辑收敛到 helper 和模板前置变量里，父 umbrella 很难统一透传。

### ontology-query

#### 涉及文件

- [ontology-query/templates/_helpers.tpl](/code/kweaver/kweaver/deploy/charts/subCharts/ontology-query/templates/_helpers.tpl)
- [ontology-query/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/ontology-query/templates/deployment.yaml)
- [ontology-query/templates/configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/ontology-query/templates/configmap.yaml)
- [ontology-query/templates/ingress.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/ontology-query/templates/ingress.yaml)

#### 改动内容

- 在 `_helpers.tpl` 中增加 `mergedGlobalValues.image` 等通用 helper 函数
- `deployment.yaml` 中：
  - 使用 `{{ include "mergedGlobalValues.image" . }}` 设置镜像
  - 使用 `{{ include "mergedGlobalValues.replicaCount" . }}` 设置副本数
  - 在文件开头定义 `{{- $env := include "mergedGlobalValues.env" . | fromYaml -}}`
- `configmap.yaml` 与 `ingress.yaml` 中：
  - 在文件开头定义 `{{- $depServices := include "mergedGlobalValues.depServices" . | fromYaml -}}`
  - 使用 `{{ include "mergedGlobalValues.ingressClassName" . }}` 设置 ingress class

#### 改动原因

`ontology-query` 同时依赖镜像仓库、环境变量和多个依赖服务地址。它如果只认本地 values，就无法在 umbrella 层做到统一透传。

### mdl-data-model

#### 涉及文件

- [mdl-data-model/templates/_helpers.tpl](/code/kweaver/kweaver/deploy/charts/subCharts/mdl-data-model/templates/_helpers.tpl)
- [mdl-data-model/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/mdl-data-model/templates/deployment.yaml)
- [mdl-data-model/templates/configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/mdl-data-model/templates/configmap.yaml)
- [mdl-data-model/templates/ingress.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/mdl-data-model/templates/ingress.yaml)

#### 改动内容

- 在 `_helpers.tpl` 中增加 `mergedGlobalValues.image` 等通用 helper 函数
- 在 `deployment.yaml` 中：
  - 使用 `{{ include "mergedGlobalValues.image" . }}` 设置镜像
  - 使用 `{{ include "mergedGlobalValues.replicaCount" . }}` 设置副本数
  - 在文件开头定义 `{{- $env := include "mergedGlobalValues.env" . | fromYaml -}}`
  - 在文件开头定义 `{{- $depServices := include "mergedGlobalValues.depServices" . | fromYaml -}}`
- 在 `configmap.yaml` 与 `ingress.yaml` 中：
  - 统一从 `$depServices` 变量读取依赖服务配置

#### 改动原因

`mdl-data-model` 需要大量依赖服务参数。把 `depServices` 合并逻辑放到模板入口后，父 chart 才能通过一份 `global.depServices` 去驱动它。

### mdl-data-model-job

#### 涉及文件

- [mdl-data-model-job/templates/_helpers.tpl](/code/kweaver/kweaver/deploy/charts/subCharts/mdl-data-model-job/templates/_helpers.tpl)
- [mdl-data-model-job/templates/deployment.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/mdl-data-model-job/templates/deployment.yaml)
- [mdl-data-model-job/templates/configmap.yaml](/code/kweaver/kweaver/deploy/charts/subCharts/mdl-data-model-job/templates/configmap.yaml)

#### 改动内容

- 在 `_helpers.tpl` 中增加 `mergedGlobalValues.image` 等通用 helper 函数
- 在 `deployment.yaml` 中：
  - 使用 `{{ include "mergedGlobalValues.image" . }}` 设置镜像
  - 使用 `{{ include "mergedGlobalValues.replicaCount" . }}` 设置副本数
  - 在文件开头定义 `{{- $env := include "mergedGlobalValues.env" . | fromYaml -}}`
  - 在文件开头定义 `{{- $depServices := include "mergedGlobalValues.depServices" . | fromYaml -}}`
- 在 `configmap.yaml` 中从 `$depServices` 变量读取依赖服务配置

#### 改动原因

`mdl-data-model-job` 和 `mdl-data-model` 属于同一类问题：如果不兼容 `global.*`，父 chart 就要重复维护一套 job 专用配置，透传链路会变得零散且难维护。

---

## 验证结论

### 本地验证

执行以下命令后通过：

- `helm dependency build /code/kweaver/kweaver/deploy/charts/kweaver-core`
- `helm lint /code/kweaver/kweaver/deploy/charts/kweaver-core -f /code/kweaver/kweaver/deploy/conf/products-values.yaml -f /tmp/kweaver-core-remote-values.yaml`
- `helm template kweaver-core /code/kweaver/kweaver/deploy/charts/kweaver-core -f /code/kweaver/kweaver/deploy/conf/products-values.yaml -f /tmp/kweaver-core-remote-values.yaml`
- `helm package /code/kweaver/kweaver/deploy/charts/kweaver-core -d /tmp/kweaver-core-package`

验证结果：

- 重复 `(kind, namespace, name)` 为 `0`
- 不再出现固定 `dip` / `anyshare` namespace 漏出

### 远端验证

在测试机安装 vendor bundle 后，已确认：

- 原先的 namespace 冲突不再出现
- 原先的重复资源名冲突不再出现
- `helm get manifest` 中仍能看到预期透传值，例如：
  - `mariadb-proton-mariadb.resource.svc.cluster.local`
  - `redis-proton-redis-sentinel.resource.svc.cluster.local`
  - `131.186.7.78`

当前剩余问题已经转为运行态问题，而不是 chart 组合问题，例如：

- 部分镜像地址仍有历史值问题
- 测试机节点资源 / 状态限制
- 部分业务容器自身启动失败

换句话说，这次文档中归纳的 namespace / fullname / 重名资源修复，已经足以支撑 umbrella 包从“安装阶段失败”推进到“实际工作负载创建阶段”。

---

## 改动归类建议

### 建议回提到各子 chart 仓库的通用修复

- 所有 `.Values.namespace` -> `.Release.Namespace` 的修复
- 所有因为共享 `.Release.Name` 导致的 fullname / resource name 冲突修复
- 所有通过 `_helpers.tpl` 增加的 `mergedGlobalValues.*` 通用 helper 函数
- 所有模板文件中使用 `{{ include "mergedGlobalValues.xxx" . }}` 替代直接使用 `.Values.xxx`
- 所有对象改名后配套的 selector、volume、`configMapKeyRef`、`secretName`、Ingress backend、Hook Job 名称同步修复

这类修改本质上是在修复 chart 的可组合性问题。只要 chart 未来还可能被 umbrella 方式复用，这些改动就应该沉淀回子 chart 上游。

---

## Chart 改造完成情况

### ISF 相关 Chart (17个)

| Chart 名称 | GitHub 地址 | 是否改造完成 | 备注 |
|-----------|------------|------------|------|
| hydra | https://github.com/kweaver-ai/hydra | ✅ 已完成 | 追加 mergedGlobalValues helper |
| sharemgnt-single | https://github.com/kweaver-ai/sharemgnt-single | ✅ 已完成 | 新建 mergedGlobalValues helper |
| user-management | https://github.com/kweaver-ai/user-management | ✅ 已完成 | 新建 mergedGlobalValues helper |
| sharemgnt | https://github.com/kweaver-ai/sharemgnt | ✅ 已完成 | 新建 mergedGlobalValues helper |
| authentication | https://github.com/kweaver-ai/authentication | ✅ 已完成 | 新建 mergedGlobalValues helper |
| policy-management | https://github.com/kweaver-ai/policy-management | ✅ 已完成 | 追加 mergedGlobalValues helper |
| audit-log | https://github.com/kweaver-ai/audit-log | ✅ 已完成 | 追加 mergedGlobalValues helper |
| eacp | https://github.com/kweaver-ai/eacp | ✅ 已完成 | 新建 mergedGlobalValues helper |
| thirdparty-message-plugin | https://github.com/kweaver-ai/thirdparty-message-plugin | ✅ 已完成 | 追加 mergedGlobalValues helper |
| isfwebthrift | https://github.com/kweaver-ai/isfwebthrift | ✅ 已完成 | 追加 mergedGlobalValues helper |
| message | https://github.com/kweaver-ai/message | ✅ 已完成 | 新建 mergedGlobalValues helper |
| isfweb | https://github.com/kweaver-ai/isfweb | ✅ 已完成 | 追加 mergedGlobalValues helper |
| authorization | https://github.com/kweaver-ai/authorization | ✅ 已完成 | 新建 mergedGlobalValues helper |
| news-feed | https://github.com/kweaver-ai/news-feed | ✅ 已完成 | 新建 mergedGlobalValues helper |
| ingress-informationsecurityfabric | https://github.com/kweaver-ai/ingress-informationsecurityfabric | ✅ 已完成 | 新建 mergedGlobalValues helper |
| eacp-single | https://github.com/kweaver-ai/eacp-single | ✅ 已完成 | 新建 mergedGlobalValues helper |
| oauth2-ui | https://github.com/kweaver-ai/oauth2-ui | ✅ 已完成 | 新建 mergedGlobalValues helper |

### AgentOperator 相关 Chart (4个)

| Chart 名称 | GitHub 地址 | 是否改造完成 | 备注 |
|-----------|------------|------------|------|
| agent-operator-integration | https://github.com/kweaver-ai/agent-operator-integration | ✅ 已完成 | 新建 mergedGlobalValues helper |
| operator-web | https://github.com/kweaver-ai/operator-web | ✅ 已完成 | 新建 mergedGlobalValues helper |
| agent-retrieval | https://github.com/kweaver-ai/agent-retrieval | ✅ 已完成 | 新建 mergedGlobalValues helper |
| data-retrieval | https://github.com/kweaver-ai/data-retrieval | ✅ 已完成 | 新建 mergedGlobalValues helper |

### DataAgent 相关 Chart (9个)

| Chart 名称 | GitHub 地址 | 是否改造完成 | 备注 |
|-----------|------------|------------|------|
| agent-backend | https://github.com/kweaver-ai/agent-backend | ✅ 已完成 | 新建 mergedGlobalValues helper |
| agent-web | https://github.com/kweaver-ai/agent-web | ✅ 已完成 | 新建 mergedGlobalValues helper |
| agent-app | https://github.com/kweaver-ai/agent-app | ✅ 已完成 | 新建 mergedGlobalValues helper |
| agent-executor | https://github.com/kweaver-ai/agent-executor | ✅ 已完成 | 新建 mergedGlobalValues helper |
| agent-factory | https://github.com/kweaver-ai/agent-factory | ✅ 已完成 | 新建 mergedGlobalValues helper |
| agent-memory | https://github.com/kweaver-ai/agent-memory | ✅ 已完成 | 新建 mergedGlobalValues helper |
| coderunner | https://github.com/kweaver-ai/coderunner | ✅ 已完成 | 追加 mergedGlobalValues helper |
| dataflow | https://github.com/kweaver-ai/dataflow | ✅ 已完成 | 追加 mergedGlobalValues helper |
| agent-operator-app | https://github.com/kweaver-ai/agent-operator-app | ✅ 已完成 | 新建 mergedGlobalValues helper |

### Ontology 相关 Chart (9个)

| Chart 名称 | GitHub 地址 | 是否改造完成 | 备注 |
|-----------|------------|------------|------|
| ontology-manager | https://github.com/kweaver-ai/ontology-manager | ✅ 已完成 | 追加 mergedGlobalValues helper |
| ontology-query | https://github.com/kweaver-ai/ontology-query | ✅ 已完成 | 追加 mergedGlobalValues helper |
| vega-web | https://github.com/kweaver-ai/vega-web | ✅ 已完成 | 追加 mergedGlobalValues helper |
| data-connection | https://github.com/kweaver-ai/data-connection | ✅ 已完成 | 新建 mergedGlobalValues helper |
| vega-gateway | https://github.com/kweaver-ai/vega-gateway | ✅ 已完成 | 新建 mergedGlobalValues helper |
| vega-gateway-pro | https://github.com/kweaver-ai/vega-gateway-pro | ✅ 已完成 | 追加 mergedGlobalValues helper |
| mdl-data-model | https://github.com/kweaver-ai/mdl-data-model | ✅ 已完成 | 追加 mergedGlobalValues helper |
| mdl-uniquery | https://github.com/kweaver-ai/mdl-uniquery | ✅ 已完成 | 追加 mergedGlobalValues helper |
| mdl-data-model-job | https://github.com/kweaver-ai/mdl-data-model-job | ✅ 已完成 | 追加 mergedGlobalValues helper |

### 其他 Chart (1个)

| Chart 名称 | GitHub 地址 | 是否改造完成 | 备注 |
|-----------|------------|------------|------|
| sandbox | https://github.com/kweaver-ai/sandbox | ✅ 已完成 | 追加 mergedGlobalValues helper |

### 改造统计

- **总计**: 40 个 Chart
- **已完成**: 40 个 (100%)
- **新建 helper**: 24 个
- **追加 helper**: 16 个

### 改造内容

所有 Chart 已完成以下改造：

1. ✅ 添加通用 helper 函数 `mergedGlobalValues.*`（11个函数）
2. ✅ 更新 deployment.yaml 使用通用 helper（35个）
3. ✅ 更新 configmap.yaml 使用通用 helper（30个）
4. ✅ 更新 ingress.yaml 使用通用 helper（3个）
5. ✅ 更新 secret.yaml 使用通用 helper（7个）
6. ✅ 清理所有内联 global 逻辑冲突（51个文件）
7. ✅ 替换所有 `$imageRegistry` 为通用 helper

### 测试验证

所有 Chart 已通过以下测试：

```bash
# 单个 Chart 测试
helm template test ./subCharts/[chart-name] \
  --set global.image.registry=test.registry.com \
  --set global.replicaCount=5

# 结果: ✅ replicas: 5, image: test.registry.com/...
```

### 通用 Helper 函数列表

| 函数名 | 功能 | 优先级 |
|-------|------|--------|
| `mergedGlobalValues.imageRegistry` | 镜像仓库地址 | Global 优先 |
| `mergedGlobalValues.replicaCount` | 副本数 | Global 优先 |
| `mergedGlobalValues.env` | 环境变量 | 深度合并 |
| `mergedGlobalValues.depServices` | 依赖服务配置 | 深度合并 |
| `mergedGlobalValues.imagePullSecrets` | 镜像拉取密钥 | Local 优先 |
| `mergedGlobalValues.namespace` | 命名空间 | Global 优先 |
| `mergedGlobalValues.mode` | 运行模式 | Global 优先 |
| `mergedGlobalValues.accessAddress` | 访问地址 | 深度合并 |
| `mergedGlobalValues.ingressClassName` | Ingress 类名 | Global 优先 |
| `mergedGlobalValues.flowAutomation` | 流程自动化配置 | 深度合并 |
| `mergedGlobalValues.image` | 完整镜像路径 | 组合函数 |

### Helper 函数示例

```yaml
{{- define "mergedGlobalValues.replicaCount" -}}
{{- $global := .Values.global | default dict -}}
{{- if hasKey $global "replicaCount" -}}
{{- $global.replicaCount -}}
{{- else -}}
{{- .Values.replicaCount -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.env" -}}
{{- $globalEnv := (.Values.global | default dict).env | default dict -}}
{{- if $globalEnv -}}
{{- toYaml (mergeOverwrite (deepCopy .Values.env) $globalEnv) -}}
{{- else -}}
{{- toYaml .Values.env -}}
{{- end -}}
{{- end -}}

{{- define "mergedGlobalValues.image" -}}
{{- $registry := include "mergedGlobalValues.imageRegistry" . -}}
{{- printf "%s/%s:%s" $registry .Values.image.repository .Values.image.tag -}}
{{- end -}}
```

### 使用示例

```yaml
# deployment.yaml
{{- $env := include "mergedGlobalValues.env" . | fromYaml -}}
{{- $depServices := include "mergedGlobalValues.depServices" . | fromYaml -}}
apiVersion: apps/v1
kind: Deployment
spec:
  replicas: {{ include "mergedGlobalValues.replicaCount" . }}
  template:
    spec:
      containers:
      - name: myapp
        image: {{ include "mergedGlobalValues.image" . }}
        env:
        - name: REDIS_HOST
          value: {{ $depServices.redis.connectInfo.host }}
```
