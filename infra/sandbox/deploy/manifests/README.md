# Kubernetes 部署指南

本目录包含 Sandbox Platform 的 Kubernetes 原生部署 YAML 文件。

Sandbox Platform 提供三种部署方式：
1. **Docker Compose** - 适合本地开发环境
2. **Helm Chart** (推荐生产环境) - 使用 Helm 包管理器进行部署，支持参数化配置
3. **K8s 原生 YAML** (本目录) - 直接使用 kubectl 应用 YAML 文件

## 部署方式对比

| 特性 | Helm Chart | 静态 YAML |
|------|-----------|----------|
| 配置灵活性 | 参数化配置，支持 values 覆盖 | 需要手动编辑 YAML |
| 版本管理 | 支持版本回滚、升级 | 手动管理 |
| 依赖管理 | 自动处理依赖关系 | 手动按顺序部署 |
| 生产环境 | 推荐使用 | 适合简单场景 |

## 方式一：Helm Chart 部署 (推荐)

### 前置要求

- Kubernetes 1.24+ (或 Minikube / K3s / Kind 等本地 K8s 环境)
- Helm 3.0+
- kubectl CLI
- Docker 镜像已构建并推送到镜像仓库

### 快速开始

```bash
# 进入 helm chart 目录
cd deploy/helm

# 安装 Helm Chart
make install
# 或使用 helm 命令
helm install sandbox ./sandbox --namespace sandbox-system --create-namespace

# 使用自定义配置
helm install sandbox ./sandbox \
  --namespace sandbox-system \
  --create-namespace \
  --set controlPlane.replicaCount=3 \
  --set web.replicaCount=2

# 查看部署状态
helm status sandbox --namespace sandbox-system

# 卸载
make uninstall
# 或
helm uninstall sandbox --namespace sandbox-system
```

### 配置说明

主要配置参数：

| 参数 | 描述 | 默认值 |
|------|------|--------|
| `controlPlane.replicaCount` | Control Plane 副本数 | `1` |
| `controlPlane.image.repository` | Control Plane 镜像仓库 | `sandbox-control-plane` |
| `controlPlane.image.tag` | Control Plane 镜像标签 | `latest` |
| `web.replicaCount` | Web Console 副本数 | `1` |
| `web.image.repository` | Web Console 镜像仓库 | `sandbox-web` |
| `mariadb.enabled` | 是否部署内置 MariaDB | `true` |
| `minio.enabled` | 是否部署内置 MinIO | `true` |

完整配置请参考 [deploy/helm/sandbox/README.md](../helm/sandbox/README.md)

### 常用命令

```bash
# 模板生成（调试用）
make template
# 或
helm template sandbox ./sandbox > helm-template-gen.yaml

# Lint 检查
make lint
# 或
helm lint ./sandbox

# 升级部署
make upgrade
# 或
helm upgrade sandbox ./sandbox --namespace sandbox-system

# 回滚
helm rollback sandbox <revision> --namespace sandbox-system
```

---

## 方式二：静态 YAML 文件部署

### 前置要求

- Kubernetes 1.24+ (或 Minikube / K3s / Kind 等本地 K8s 环境)
- kubectl CLI
- Docker 镜像已构建并推送到镜像仓库

## 快速开始

### 1. 构建镜像

```bash
# 在 sandbox_control_plane 目录下
docker build -t sandbox-control-plane:latest .
```

### 2. 加载镜像到本地 K8s（Minikube/K3s/Kind）

```bash
# Minikube
minikube image load sandbox-control-plane:latest

# Kind
kind load docker-image sandbox-control-plane:latest

# K3s（直接使用 Docker 镜像）
# 无需额外操作
```

### 3. 部署到 Kubernetes

```bash
# 按顺序部署所有资源
kubectl apply -f deploy/manifests/00-namespace.yaml
kubectl apply -f deploy/manifests/01-configmap.yaml
kubectl apply -f deploy/manifests/02-secret.yaml
kubectl apply -f deploy/manifests/03-serviceaccount.yaml
kubectl apply -f deploy/manifests/04-role.yaml
kubectl apply -f deploy/manifests/09-runtime-namespace.yaml
kubectl apply -f deploy/manifests/08-mariadb-deployment.yaml
kubectl apply -f deploy/manifests/07-minio-deployment.yaml
kubectl apply -f deploy/manifests/05-control-plane-deployment.yaml
kubectl apply -f deploy/manifests/11-sandbox-web-deployment.yaml
kubectl apply -f deploy/manifests/06-hpa.yaml

# 或一次性部署所有
kubectl apply -f deploy/manifests/
```

### 4. 验证部署

```bash
# 检查所有 Pod 状态
kubectl get pods -n sandbox-system

# 检查服务状态
kubectl get svc -n sandbox-system

# 检查 Control Plane 日志
kubectl logs -f deployment/sandbox-control-plane -n sandbox-system

# 检查 Web Console 日志
kubectl logs -f deployment/sandbox-web -n sandbox-system

# 端口转发到本地（使用 port-forward.sh 脚本）
cd ../../scripts
./port-forward.sh start --all --background

# 或单独端口转发
kubectl port-forward svc/sandbox-control-plane 8000:8000 -n sandbox-system
kubectl port-forward svc/sandbox-web 1101:80 -n sandbox-system
kubectl port-forward svc/minio 9001:9001 -n sandbox-system
```

## 配置说明

### 命名空间

- `sandbox-system`: Control Plane 管理命名空间
- `sandbox-runtime`: Executor Pod 运行命名空间

### ConfigMap 配置项

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `ENVIRONMENT` | 运行环境 | production |
| `DATABASE_URL` | 数据库连接字符串 | - |
| `DEFAULT_TIMEOUT` | 默认超时时间（秒） | 300 |
| `MAX_TIMEOUT` | 最大超时时间（秒） | 3600 |
| `DISABLE_BWRAP` | 禁用 Bubblewrap | true |

### Secret 配置项

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `S3_ENDPOINT_URL` | MinIO/S3 端点 | - |
| `S3_ACCESS_KEY_ID` | S3 访问密钥 ID | - |
| `S3_SECRET_ACCESS_KEY` | S3 访问密钥 | - |
| `S3_BUCKET` | S3 Bucket 名称 | sandbox-workspace |

## 扩缩容

### 手动扩缩容

```bash
# 扩展到 3 个副本
kubectl scale deployment/sandbox-control-plane --replicas=3 -n sandbox-system
```

### 自动扩缩容（HPA）

Control Plane 配置了 HPA，会根据 CPU 和内存利用率自动扩缩容：

- 最小副本数：2
- 最大副本数：10
- 目标 CPU 利用率：70%
- 目标内存利用率：80%

```bash
# 查看 HPA 状态
kubectl get hpa -n sandbox-system
```

## 故障排查

### Pod 无法启动

```bash
# 查看 Pod 状态
kubectl describe pod <pod-name> -n sandbox-system

# 查看 Pod 日志
kubectl logs <pod-name> -n sandbox-system

# 查看之前容器的日志
kubectl logs <pod-name> -n sandbox-system --previous
```

### 数据库连接问题

```bash
# 检查 MariaDB Pod
kubectl get pods -n sandbox-system -l app=mariadb

# 测试数据库连接
kubectl exec -it deployment/mariadb -n sandbox-system -- mariadb -u root -ppassword -e "SELECT 1"
```

### MinIO 存储问题

```bash
# 检查 MinIO Pod
kubectl get pods -n sandbox-system -l app=minio

# 访问 MinIO Console
kubectl port-forward svc/minio 9001:9001 -n sandbox-system
# 浏览器访问 http://localhost:9001
# 用户名/密码: minioadmin/minioadmin
```

### Web Console 访问问题

```bash
# 检查 Web Console Pod
kubectl get pods -n sandbox-system -l app=sandbox-web

# 检查 Web Console 日志
kubectl logs -f deployment/sandbox-web -n sandbox-system

# 确保能访问 Control Plane
kubectl exec -it deployment/sandbox-web -n sandbox-system -- wget -O- http://sandbox-control-plane:8000/api/v1/health
```

## 卸载

```bash
# 删除所有资源
kubectl delete -f deploy/manifests/

# 或单独删除命名空间
kubectl delete namespace sandbox-system
kubectl delete namespace sandbox-runtime
```

## 本地开发（Minikube/K3s）

### Minikube

```bash
# 启动 Minikube
minikube start --cpus=4 --memory=8192 --disk-size=50g

# 启用 Ingress（可选）
minikube addons enable ingress

# 部署应用
kubectl apply -f deploy/manifests/

# 访问应用
minikube tunnel  # 在另一个终端运行
```

### K3s

```bash
# 安装 K3s
curl -sfL https://get.k3s.io | sh -

# 获取 K3s kubeconfig
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

# 部署应用
kubectl apply -f deploy/manifests/
```

### Kind

```bash
# 创建 Kind 集群
kind create cluster --name sandbox

# 加载镜像
kind load docker-image sandbox-control-plane:latest --name sandbox

# 部署应用
kubectl apply -f deploy/manifests/

# 端口转发
kubectl port-forward svc/sandbox-control-plane 8000:8000 -n sandbox-system
```

## 生产环境注意事项

1. **镜像仓库**: 将镜像推送到私有镜像仓库，修改 Deployment 中的镜像地址
2. **持久化存储**: 配置合适的 StorageClass（如 Longhorn、Ceph RBD）
3. **证书管理**: 使用 cert-manager 管理 TLS 证书
4. **监控集成**: 添加 Prometheus ServiceMonitor 和 Grafana Dashboard
5. **日志聚合**: 集成 ELK 或 Loki 收集日志
6. **备份策略**: 定期备份 MariaDB 数据

## 架构说明

```
┌─────────────────────────────────────────────────────────────┐
│                    sandbox-system 命名空间                    │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │          Control Plane Deployment (2+ replicas)         │ │
│  │                                                         │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │ │
│  │  │  API Gateway │  │   Scheduler  │  │Session Mgr   │  │ │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │          Web Console Deployment (1+ replica)            │ │
│  │                                                         │ │
│  │  ┌──────────────────────────────────────────────────┐  │ │
│  │  │      Nginx + React SPA (port 80)                 │  │ │
│  │  └──────────────────────────────────────────────────┘  │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐                         │
│  │    MinIO     │  │   MariaDB    │                         │
│  │  (S3 Storage)│  │  (Database)  │                         │
│  └──────────────┘  └──────────────┘                         │
└─────────────────────────────────────────────────────────────┘
                            ↓
                    ┌───────────────────────────────────┐
                    │     sandbox-runtime 命名空间        │
                    │                                   │
                    │  ┌────────────┐  ┌────────────┐  │
                    │  │ Executor-1 │  │ Executor-2 │  │ ... (N)
                    │  │  (Session)  │  │  (Session)  │  │
                    │  └────────────┘  └────────────┘  │
                    │                                   │
                    └───────────────────────────────────┘
```
