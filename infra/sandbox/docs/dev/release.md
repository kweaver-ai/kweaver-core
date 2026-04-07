# 发布

当前仓库的发布相关资产主要位于：

- `deploy/docker-compose/`
- `deploy/manifests/`
- `deploy/helm/sandbox/`
- `images/`

发布前至少完成：

1. 构建并验证镜像。
2. 执行各子项目测试。
3. 检查 OpenAPI 与实现是否一致。
4. 更新受影响的 PRD、设计和运维文档。
