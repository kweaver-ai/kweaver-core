# 构建

## 镜像构建

默认只构建随项目 `VERSION` 发布的最终 executor/template 镜像：

```bash
cd images
./build.sh
```

默认输出：

- `sandbox-template-python-basic:<VERSION>`
- `sandbox-template-multi-language:<VERSION>`

稳定运行时 base 镜像不随每次业务版本发布重建。只有 Python、Go、Bash、系统包等基础依赖变化时才构建：

```bash
cd images
./build.sh --build-bases
```

base 镜像：

- `sandbox-python-executor-base:python3.11-v1`
- `sandbox-multi-executor-base:go1.25-python3.11-v1`

构建 base 镜像后可以同步推送到华为云 SWR。脚本会额外执行一次 `docker buildx build`，按默认平台 `linux/amd64,linux/arm64` 生成 OCI archive，然后使用 `skopeo copy --all` 推送到 SWR：

```bash
cd images
./build.sh --build-bases --push-swr-bases \
  --swr-registry swr.cn-east-3.myhuaweicloud.com \
  --swr-namespace kweaver-ai/sandbox \
  --swr-creds '<username>:<password>' \
  --swr-dest-tls-verify false
```

由于 Docker 默认 buildx builder 可能使用 `docker` driver，而该 driver 不支持 `--output type=oci`，脚本会自动创建并使用 `docker-container` driver 的 `sandbox-swr-builder`。如需复用已有 builder，可以指定：

```bash
cd images
./build.sh --build-bases --push-swr-bases \
  --swr-buildx-builder my-buildx-builder \
  --swr-registry swr.cn-east-3.myhuaweicloud.com \
  --swr-namespace kweaver-ai/sandbox \
  --swr-creds '<username>:<password>'
```

默认推送目标：

- `docker://<SWR_REGISTRY>/<SWR_NAMESPACE>/sandbox-python-executor-base:python3.11-v1`
- `docker://<SWR_REGISTRY>/<SWR_NAMESPACE>/sandbox-multi-executor-base:go1.25-python3.11-v1`

multi-language base 的多架构构建会通过 SWR 上的 Python base 作为父镜像，因此推荐用 `--build-bases --push-swr-bases` 一次性构建并推送两个 base。如果只推送 multi-language base，需要先确保对应的 Python base 已存在于 SWR。

SWR 多架构构建使用 `docker buildx build --platform <platforms>`。Dockerfile 中的 `TARGETARCH` 不需要在脚本里手动设置，Buildx 会针对 `linux/amd64`、`linux/arm64` 等目标平台分别注入 `amd64`、`arm64`。

可指定目标平台和 OCI archive 输出目录：

```bash
cd images
./build.sh --build-bases --push-swr-bases \
  --swr-registry swr.cn-east-3.myhuaweicloud.com \
  --swr-namespace kweaver-ai/sandbox \
  --swr-creds '<username>:<password>' \
  --swr-platforms linux/amd64,linux/arm64 \
  --swr-oci-output-dir /tmp/sandbox-base-oci-images
```

如果 SWR 仓库名需要和本地镜像名不同，可以显式指定：

```bash
cd images
./build.sh --build-bases --push-swr-bases \
  --swr-registry swr.cn-east-3.myhuaweicloud.com \
  --swr-namespace kweaver-ai/sandbox \
  --swr-creds '<username>:<password>' \
  --swr-python-base-repository python \
  --swr-multi-base-repository multi-language
```

只构建多语言最终镜像：

```bash
cd images
./build.sh --templates "multi-language"
```

可选镜像源：

```bash
cd images
USE_MIRROR=true ./build.sh
```

构建 multi-language base 时，`--use-mirror` 也会把 Go tarball 下载源从 `https://go.dev/dl` 切到 `https://mirrors.ustc.edu.cn/golang`，用于规避 `go.dev` TLS 连接中断或下载超时：

```bash
cd images
./build.sh --build-multi-base --use-mirror
```

构建 Python base 时，`--use-mirror` 会把 Docker Hub 基础镜像 `python:3.11-slim` 切到 `docker.m.daocloud.io/library/python:3.11-slim`。如果当前网络仍无法访问该镜像源，可以指定内部镜像源：

```bash
cd images
./build.sh --build-python-base --use-mirror \
  --python-image-mirror <registry>/library/python:3.11-slim
```

也可以显式指定 Go 下载源：

```bash
cd images
./build.sh --build-multi-base \
  --go-download-base https://mirrors.ustc.edu.cn/golang
```

常用参数：

```bash
# 指定最终镜像 tag；不指定时读取 VERSION 文件
./build.sh --template-tag 2.1.4

# 指定稳定 base tag
./build.sh --build-multi-base --multi-base-tag go1.25-python3.11-v2

# 指定本地 docker build 的单架构平台；SWR 多架构推送请使用 --swr-platforms
./build.sh --build-multi-base --docker-build-platform linux/arm64
```

## 前端构建

```bash
cd sandbox_web
npm run build
```
