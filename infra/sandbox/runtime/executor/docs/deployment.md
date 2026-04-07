# 部署指南

本文档介绍如何在不同环境中部署 Sandbox Executor。

## 目录

- [Docker 部署](#docker-部署)
- [Docker Compose 部署](#docker-compose-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [生产环境建议](#生产环境建议)
- [监控和维护](#监控和维护)

---

## Docker 部署

### 前置要求

Bubblewrap 需要宿主机启用**非特权用户命名空间**支持。

**Ubuntu/Debian**:
```bash
# 启用非特权用户命名空间
sudo sysctl -w kernel.unprivileged_userns_clone=1

# 持久化配置
echo "kernel.unprivileged_userns_clone=1" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

**CentOS/RHEL**:
```bash
# 设置用户命名空间最大数量
echo "user.max_user_namespaces=10000" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

**验证配置**:
```bash
# 检查当前设置
sysctl kernel.unprivileged_userns_clone  # Ubuntu/Debian
sysctl user.max_user_namespaces          # CentOS/RHEL

# 测试用户命名空间
unshare -U -m echo "User namespaces work"
```

### 构建镜像

```bash
cd runtime/executor
docker build -t sandbox-executor:v1.0 .
```

### 运行容器

**推荐方式（安全配置）**:
```bash
docker run -d \
  --name sandbox-executor \
  --security-opt seccomp=default \
  --security-opt no-new-privileges \
  --cap-drop ALL \
  -p 8080:8080 \
  -e CONTROL_PLANE_URL=http://host.docker.internal:8000 \
  -e WORKSPACE_PATH=/workspace \
  -e LOG_LEVEL=INFO \
  -v $(pwd)/workspace:/workspace \
  sandbox-executor:v1.0
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `--security-opt seccomp=default` | 启用用户命名空间的 seccomp 配置 |
| `--security-opt no-new-privileges` | 禁止权限提升 |
| `--cap-drop ALL` | 移除所有 Linux capabilities |
| `-p 8080:8080` | 端口映射 |
| `-e CONTROL_PLANE_URL` | Control Plane 地址 |
| `-v $(pwd)/workspace:/workspace` | 工作目录挂载 |

**⚠️ 安全警告**：
- **不要使用** `--privileged` 标志，这会完全破坏容器隔离
- Bubblewrap 通过用户命名空间工作，不需要特权模式
- 如果遇到权限错误，请检查宿主机的用户命名空间配置

### 验证部署

```bash
# 检查容器状态
docker ps | grep sandbox-executor

# 查看日志
docker logs -f sandbox-executor

# 健康检查
curl http://localhost:8080/health

# 进入容器调试
docker exec -it sandbox-executor /bin/bash
```

### 停止和删除

```bash
# 停止容器
docker stop sandbox-executor

# 删除容器
docker rm sandbox-executor

# 删除镜像
docker rmi sandbox-executor:v1.0
```

---

## Docker Compose 部署

### 前置要求

确保宿主机已启用用户命名空间（参见上文 Docker 部署的前置要求）。

### Compose 配置

```yaml
version: '3.8'

services:
  executor:
    build:
      context: .
      dockerfile: Dockerfile
    image: sandbox-executor:v1.0
    container_name: sandbox-executor
    cap_drop:
      - ALL
    security_opt:
      - seccomp=default
      - no-new-privileges
    ports:
      - "8080:8080"
    environment:
      - CONTROL_PLANE_URL=http://control-plane:8000
      - WORKSPACE_PATH=/workspace
      - LOG_LEVEL=INFO
    volumes:
      - ./workspace:/workspace
    networks:
      - sandbox-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  control-plane:
    image: sandbox-control-plane:v1.0
    container_name: sandbox-control-plane
    ports:
      - "8000:8000"
    networks:
      - sandbox-network
    restart: unless-stopped

networks:
  sandbox-network:
    driver: bridge
```

### 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f executor

# 查看状态
docker-compose ps

# 停止服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v
```

### 扩展服务

```bash
# 扩展到 3 个实例
docker-compose up -d --scale executor=3

# 注意：需要配置负载均衡器
```

---

## Kubernetes 部署

### 前置要求

确保 Kubernetes 节点已启用用户命名空间：

```bash
# 在所有节点上执行
sudo sysctl -w kernel.unprivileged_userns_clone=1  # Ubuntu/Debian
# 或
echo "user.max_user_namespaces=10000" | sudo tee -a /etc/sysctl.conf  # CentOS/RHEL
sudo sysctl -p
```

### 部署配置

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sandbox-executor
  labels:
    app: sandbox-executor
spec:
  replicas: 3
  selector:
    matchLabels:
      app: sandbox-executor
  template:
    metadata:
      labels:
        app: sandbox-executor
    spec:
      containers:
      - name: executor
        image: sandbox-executor:v1.0
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: CONTROL_PLANE_URL
          value: "http://control-plane-service:8000"
        - name: WORKSPACE_PATH
          value: "/workspace"
        - name: LOG_LEVEL
          value: "INFO"
        securityContext:
          runAsUser: 1000
          runAsGroup: 1000
          allowPrivilegeEscalation: false
          capabilities:
            drop:
              - ALL
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: workspace
          mountPath: /workspace
      volumes:
      - name: workspace
        emptyDir: {}
```

### 服务配置

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: sandbox-executor-service
spec:
  selector:
    app: sandbox-executor
  ports:
  - name: http
    protocol: TCP
    port: 8080
    targetPort: 8080
  type: ClusterIP
```

### 部署到 Kubernetes

```bash
# 创建部署
kubectl apply -f deployment.yaml

# 创建服务
kubectl apply -f service.yaml

# 查看状态
kubectl get pods -l app=sandbox-executor

# 查看日志
kubectl logs -f deployment/sandbox-executor

# 扩展副本
kubectl scale deployment sandbox-executor --replicas=5

# 删除部署
kubectl delete deployment sandbox-executor
```

### 水平自动扩缩容 (HPA)

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: sandbox-executor-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: sandbox-executor
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

```bash
# 创建 HPA
kubectl apply -f hpa.yaml

# 查看 HPA 状态
kubectl get hpa
```

---

## 生产环境建议

### 资源配置

| 环境 | CPU | 内存 | 实例数 |
|------|-----|------|--------|
| 开发 | 0.5 | 512MB | 1 |
| 测试 | 1 | 1GB | 2 |
| 生产 | 2+ | 2GB+ | 3+ |

### 安全配置

```yaml
# 使用 secrets 管理敏感信息
apiVersion: v1
kind: Secret
metadata:
  name: executor-secrets
type: Opaque
stringData:
  INTERNAL_API_TOKEN: "your-secret-token"
```

### 日志配置

```yaml
env:
- name: LOG_LEVEL
  value: "INFO"
- name: LOG_FORMAT
  value: "json"

# 使用日志收集器
volumeMounts:
- name: logs
  mountPath: /var/log/executor
volumes:
- name: logs
  emptyDir: {}
```

### 监控配置

```yaml
# Prometheus 监控
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

---

## 监控和维护

### 健康检查

```bash
# 基础健康检查
curl http://localhost:8080/health

# 服务信息
curl http://localhost:8080/info

# 检查活跃执行数
curl http://localhost:8080/info | jq '.active_executions'
```

### 日志管理

```bash
# Docker 日志
docker logs -f sandbox-executor

# Kubernetes 日志
kubectl logs -f deployment/sandbox-executor

# 持久化日志
docker run -v $(pwd)/logs:/var/log/executor sandbox-executor:v1.0
```

### 工作区清理

```bash
# 进入容器清理
docker exec sandbox-executor rm -rf /workspace/*

# 定期清理（cron）
0 2 * * * docker exec sandbox-executor find /workspace -type f -mtime +7 -delete
```

### 性能监控

```python
# 使用 Prometheus 指标
from prometheus_client import Counter, Histogram

execution_counter = Counter('executions_total', 'Total executions')
execution_duration = Histogram('execution_duration_seconds', 'Execution duration')

# 在代码中使用
execution_counter.inc()
with execution_duration.time():
    # 执行代码
    pass
```

---

## 故障恢复

### 自动重启

```yaml
# Docker Compose
restart: unless-stopped

# Kubernetes
restartPolicy: Always
```

### 备份和恢复

```bash
# 备份工作区
docker exec sandbox-executor tar -czf /tmp/workspace-backup.tar.gz /workspace
docker cp sandbox-executor:/tmp/workspace-backup.tar.gz ./backup/

# 恢复工作区
docker cp ./backup/workspace-backup.tar.gz sandbox-executor:/tmp/
docker exec sandbox-executor tar -xzf /tmp/workspace-backup.tar.gz -C /
```

---

## 相关文档

- [用户命名空间配置](user-namespace-guide.md) - 双层隔离原理和配置
- [快速开始](quick-start.md) - 本地开发部署
- [配置说明](configuration.md) - 环境变量和配置
- [故障排查](troubleshooting.md) - 常见问题解决
