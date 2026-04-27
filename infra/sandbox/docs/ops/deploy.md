# 部署

## 本地部署

```bash
docker-compose -f deploy/docker-compose/docker-compose.yml up -d
```

## Kubernetes

主要部署资产：

- `deploy/manifests/`
- `deploy/helm/sandbox/`

补充背景可参考 [历史文档：监控与部署](./monitoring-and-deployment.md)。
