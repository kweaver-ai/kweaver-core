# 开发环境准备

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- Python 3.11+
- Node.js 与 npm
- `uv`

## 初始化

```bash
docker-compose -f deploy/docker-compose/docker-compose.yml up -d
```

如需使用分支构建的模板镜像，启动或重建 control-plane 时指定完整 tag：

```bash
TEMPLATE_IMAGE_TAG=0.4.0-feature-sandbox-20260512.git-4188ba2-opensource \
docker-compose -f deploy/docker-compose/docker-compose.yml up -d --force-recreate control-plane
```

docker-compose 会将该 tag 注入到默认模板镜像地址：

- `swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/sandbox-template-python-basic:<TEMPLATE_IMAGE_TAG>`
- `swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/sandbox-template-multi-language:<TEMPLATE_IMAGE_TAG>`

```bash
cd sandbox_control_plane
uv sync
```

```bash
cd sandbox_web
npm install
```

如需构建基础镜像，先执行 [构建文档](./build.md)。
