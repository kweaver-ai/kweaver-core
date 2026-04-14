# 验证安装

`deploy.sh kweaver-core install` 完成后，请确认集群与 API 可用。

## Kubernetes

```bash
kubectl get nodes
kubectl get pods -A
```

等待核心命名空间中关键工作负载为 `Running` / `Ready`。

## 部署脚本状态

```bash
cd kweaver-core/deploy
./deploy.sh kweaver-core status
```

## CLI 访问

从 [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk) 安装 CLI 后登录并列出 BKN：

```bash
kweaver auth login https://<访问地址> -k
kweaver bkn list
```

将 `<访问地址>` 替换为 `--access_address` 或安装程序提示的节点地址。

## HTTP 探测

若已知公开路由（路径随网关与版本可能不同），可用 curl：

```bash
export KWEAVER_BASE="https://<访问地址>"
curl -sk "$KWEAVER_BASE/health" || true
```

子系统具体路径请以环境中的 OpenAPI 或 Ingress 规则为准。

## 故障排查

见 [deploy/README.zh.md — 故障排查](../../../deploy/README.zh.md)。

## 下一步

- [快速开始](../quick-start.md)
- 从 [文档索引](../README.md) 进入各模块说明
