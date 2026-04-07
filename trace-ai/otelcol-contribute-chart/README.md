这是一个用于部署 OpenTelemetry Collector Contrib 的 Helm Chart，当前默认能力如下：

- 以 **Deployment 模式**在 Kubernetes 部署 Collector；
- 内置 **OTLP In / OpenSearch Out**（接收 OTLP、写入 OpenSearch）；
- 通过 GitHub Actions 将 chart 以 OCI 形式发布到 **GHCR**。

---

## 1. 项目信息

Chart 名称：`otelcol-contrib`

推荐 GHCR 包名（OCI Chart）：

```text
ghcr.io/<github_org_or_user>/charts/otelcol-contrib
```

当前版本：

- Chart version: `0.1.0`
- App version: `0.145.0`

---

## 2. 项目结构

```text
.
├── charts/
│   └── otelcol-contrib/
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
│           ├── _helpers.tpl
│           ├── configmap.yaml
│           ├── deployment.yaml
│           ├── service.yaml
│           ├── serviceaccount.yaml
│           └── NOTES.txt
├── .github/
│   └── workflows/
│       ├── chart-ci.yaml
│       └── chart-release.yaml
└── scripts/
    └── create-nodeport-service.sh
```

---

## 3. 当前能力

### Deployment 模式

`values.yaml` 中通过 `mode: deployment` 控制，当前仅支持 deployment。

### OTLP In / OpenSearch Out

默认配置：

- Receiver: `otlp`（gRPC: `4317` / HTTP: `4318`）
- Exporter: `opensearch`
- Pipelines: `traces` / `logs`

可通过 `values.yaml` 的 `opensearchExporter` 段配置 OpenSearch exporter，也可直接覆盖 `config` 段自定义 collector 配置。

默认 collector 镜像使用国内镜像源：

```text
swr.cn-north-4.myhuaweicloud.com/ddn-k8s/ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-contrib:0.145.0
```

默认会渲染以下 exporter 结构：

```yaml
exporters:
  opensearch:
    http:
      endpoint: http://opensearch-cluster-master.observability.svc.cluster.local:9200
      tls:
        insecure: true
    dataset: default
    namespace: namespace
    bulk_action: create
    mapping:
      mode: ss4o
```

如果开启 `opensearchExporter.auth.enabled=true`，chart 会自动增加 `basicauth/client` extension，并通过 `opensearch.http.auth.authenticator` 将其绑定到 exporter。

注意：官方 `opensearchexporter` 当前支持 `traces` 和 `logs`，默认不为 `metrics` 创建 OpenSearch 导出 pipeline。

---

## 4. Pipeline 设计

### CI（`.github/workflows/chart-ci.yaml`）

触发：PR、push 到 `main`。

步骤：

1. 安装 Helm
2. `helm lint charts/otelcol-contrib`
3. `helm template` 渲染校验

### 发布（`.github/workflows/chart-release.yaml`）

触发：

- 手动触发（`workflow_dispatch`）
- 推送 tag：`chart-v*`

步骤：

1. 登录 GHCR（使用 `GITHUB_TOKEN`）
2. `helm package` 生成 `.tgz`
3. `helm push` 到 `oci://ghcr.io/<owner>/charts`

---

## 5. 发布 Chart

### 步骤 1：确认仓库权限

仓库 Actions 需要 `packages: write`（workflow 已配置）。

### 步骤 2：打 tag 并推送

```bash
git tag chart-v<version>
git push origin chart-v<version>
```

### 步骤 3：验证 chart

发布成功后可拉取：

```bash
helm pull oci://ghcr.io/<github_org_or_user>/charts/otelcol-contrib --version <version>
```

---

## 6. 本地快速验证

```bash
helm lint charts/otelcol-contrib
helm template otelcol-contrib charts/otelcol-contrib
```

安装示例：

```bash
helm upgrade --install otelcol-contrib charts/otelcol-contrib -n observability --create-namespace
```

如果你在本地开发阶段使用 Docker 启动的 OpenSearch，可覆盖为宿主机地址，例如：

```bash
helm upgrade --install otelcol-contrib charts/otelcol-contrib \
  -n observability \
  --create-namespace \
  --set opensearchExporter.http.endpoint=http://192.168.139.3:9200
```

覆盖 OpenSearch endpoint 示例：

```bash
helm upgrade --install otelcol-contrib charts/otelcol-contrib \
  -n observability \
  --create-namespace \
  --set opensearchExporter.http.endpoint=http://opensearch-cluster-master.observability.svc.cluster.local:9200
```

开启 Basic Auth 示例：

```bash
helm upgrade --install otelcol-contrib charts/otelcol-contrib \
  -n observability \
  --create-namespace \
  --set opensearchExporter.auth.enabled=true \
  --set opensearchExporter.auth.username=admin \
  --set opensearchExporter.auth.password=admin
```

## 7. 使用 telemetrygen 生成测试 Trace

本地开发阶段可以使用 `telemetrygen` 生成测试 trace，验证以下链路是否打通：

- `telemetrygen -> otelcol-contrib -> OpenSearch`

这一节参考 OpenTelemetry Collector 官方 Quick Start 中关于 `telemetrygen` 的安装和发送方式，并结合当前 chart 的部署方式调整为 NodePort 验证流程：

- 官方文档：<https://opentelemetry.io/docs/collector/quick-start/>

### 7.1 安装 telemetrygen

官方 Quick Start 要求本机安装 Go，并设置 `GOBIN`。如果 `GOBIN` 未设置，可先执行：

```bash
export GOBIN=${GOBIN:-$(go env GOPATH)/bin}
```

安装 `telemetrygen`：

```bash
go install github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen@latest
```

验证是否安装成功：

```bash
$GOBIN/telemetrygen traces --help
```

### 7.2 暴露 Collector NodePort

如果当前 chart 安装的是默认 `ClusterIP` Service，可先创建一个 NodePort Service 供本机访问：

```bash
./scripts/create-nodeport-service.sh otelcol-contrib observability
```

默认端口：

- OTLP gRPC: `30417`
- OTLP HTTP: `30418`

假设 Kubernetes 节点 IP 为 `192.168.139.66`，则 OTLP gRPC 地址为：

```text
192.168.139.66:30417
```

### 7.3 发送测试 Trace

使用 OTLP/gRPC 发送 3 条 trace：

```bash
$GOBIN/telemetrygen traces \
  --otlp-insecure \
  --otlp-endpoint 192.168.139.66:30417 \
  --traces 3
```

如果命令执行成功，通常会看到类似输出：

```text
INFO traces/worker.go:180 traces generated {"worker": 0, "traces": 3}
```

### 7.4 查看 Collector 日志

查看 collector 是否已接收到 trace：

```bash
kubectl -n observability logs deploy/otelcol-contrib -f
```

如果启用了 `debugExporter.enabled=true`，日志中会直接打印 span 内容，便于确认 receiver 已收到数据。

### 7.5 查看 OpenSearch 是否已写入

当前 chart 默认将 trace 写入 OpenSearch `ss4o` 风格索引。可通过以下命令验证：

```bash
curl 'http://<opensearch-host>:9200/_cat/indices?v'
```

如果写入成功，通常会看到类似索引：

```text
ss4o_traces-default-namespace
```

也可以进一步查询文档：

```bash
curl 'http://<opensearch-host>:9200/ss4o_traces-default-namespace/_search?pretty'
```

### 7.6 常见问题

- `rpc error: code = Unavailable desc = no children to pick from`
  这通常不是 `telemetrygen` 参数错误，而是 collector 的 exporter 目标不可达，需检查 OpenSearch endpoint 是否可访问。

- 修改 Helm values 后日志仍显示旧 exporter
  当前 chart 默认不会因 ConfigMap 更新自动触发 Pod 滚动重启。修改配置后请手动执行：

```bash
kubectl -n observability rollout restart deploy/otelcol-contrib
kubectl -n observability rollout status deploy/otelcol-contrib
```
