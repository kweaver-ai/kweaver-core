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

```bash
cd sandbox_control_plane
uv sync
```

```bash
cd sandbox_web
npm install
```

如需构建基础镜像，先执行 [构建文档](./build.md)。
