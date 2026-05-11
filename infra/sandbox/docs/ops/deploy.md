# 部署

## 本地部署

```bash
docker-compose -f deploy/docker-compose/docker-compose.yml up -d
```

docker-compose 支持通过 `TEMPLATE_IMAGE_TAG` 生成 Control Plane 初始化默认模板时使用的镜像地址：

- `DEFAULT_TEMPLATE_IMAGE`
- `DEFAULT_MULTI_LANGUAGE_TEMPLATE_IMAGE`

未设置 `TEMPLATE_IMAGE_TAG` 时，Control Plane 会读取 `/app/VERSION` 并使用该值作为模板镜像 tag。如果部署分支构建产物，需要传入完整的分支镜像 tag：

```bash
TEMPLATE_IMAGE_TAG=0.4.0-feature-sandbox-20260512.git-4188ba2-opensource \
docker-compose -f deploy/docker-compose/docker-compose.yml up -d --force-recreate control-plane
```

展开后的默认模板镜像地址示例：

```text
swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/sandbox-template-python-basic:0.4.0-feature-sandbox-20260512.git-4188ba2-opensource
swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/sandbox-template-multi-language:0.4.0-feature-sandbox-20260512.git-4188ba2-opensource
```

也可以直接设置完整镜像地址：

```bash
DEFAULT_TEMPLATE_IMAGE=<python-basic-image> \
DEFAULT_MULTI_LANGUAGE_TEMPLATE_IMAGE=<multi-language-image> \
docker-compose -f deploy/docker-compose/docker-compose.yml up -d --force-recreate control-plane
```

已运行的 control-plane 容器不会自动读取新的 compose 环境变量，修改 tag 后需要重建该容器。

## Kubernetes

主要部署资产：

- `deploy/manifests/`
- `deploy/helm/sandbox/`

补充背景可参考 [历史文档：监控与部署](./monitoring-and-deployment.md)。
