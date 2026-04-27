# agent-observability

基于参考项目 `agent-factory` 的六边形架构实现的 Agent Trace 查询服务。

当前提供：
- Trace 原始 DSL 查询接口：`POST /api/agent-observability/v1/traces/_search`
- Conversation 维度包装查询接口：`GET /api/agent-observability/v1/traces/by-conversation?conversation_id=...`
- OpenSearch 查询客户端
- Swagger 文档生成
- Docker 镜像构建
- Helm Deployment Chart
- GitHub Actions 构建与发布流水线

## Development

本地测试：

```bash
make test
```

生成 Swagger 文档：

```bash
make gen-swag
```

查看 Swagger 文档地址：

```bash
make view-swag
```

服务启动后可访问：

```text
http://localhost:8080/api/agent-observability/v1/swagger/swagger.json
http://localhost:8080/api/agent-observability/v1/swagger/swagger.yaml
```

## Docker

本地构建镜像：

```bash
make docker-build
```

默认镜像名：

```text
swr.cn-east-3.myhuaweicloud.com/kweaver-ai/agent-observability:local
```

也可以覆盖：

```bash
make docker-build IMAGE=swr.cn-east-3.myhuaweicloud.com/kweaver-ai/agent-observability:v0.1.1
```

## Helm

Chart 目录：

```text
charts/agent-observability
```

本地校验：

```bash
make helm-lint
```

打包：

```bash
make helm-package
```

安装示例：

```bash
helm upgrade --install agent-observability charts/agent-observability \
  --set image.repository=swr.cn-east-3.myhuaweicloud.com/kweaver-ai/agent-observability \
  --set image.tag=0.1.1 \
  --set opensearch.endpoint=http://opensearch-cluster-master:9200 \
  --set opensearch.auth.enabled=false \
  -n observability --create-namespace
```

启用 OpenSearch Basic Auth：

```bash
helm upgrade --install agent-observability charts/agent-observability \
  --set image.repository=swr.cn-east-3.myhuaweicloud.com/kweaver-ai/agent-observability \
  --set image.tag=0.1.1 \
  --set opensearch.endpoint=http://opensearch-cluster-master:9200 \
  --set opensearch.auth.enabled=true \
  --set opensearch.auth.username=your-username \
  --set opensearch.auth.password=your-password \
  -n observability --create-namespace
```

## CI/CD

GitHub Actions 工作流位于：

```text
.github/workflows/release-agent-observability.yml
```

分为三个阶段：
- `test-and-lint`：执行 `go test ./...` 和 `golangci-lint`
- `build-and-push-image`：构建并推送 `linux/amd64`、`linux/arm64` 镜像到 SWR
- `package-and-push-chart`：打包 Helm chart 并推送到 `ghcr.io`

当前默认镜像仓库：

```text
swr.cn-east-3.myhuaweicloud.com/kweaver-ai/agent-observability
```

需要配置的 GitHub Secrets：
- `SWR_USERNAME`
- `SWR_PASSWORD`

Chart 会推送到：

```text
ghcr.io/<github-owner>/charts
```
