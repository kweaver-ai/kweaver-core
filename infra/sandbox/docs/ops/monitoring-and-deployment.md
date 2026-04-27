# 8. 监控与可观测性 + 9. 部署方案


> **文档导航**: [返回运维文档](./monitoring.md)


## 8. 监控与可观测性

### 8.1 指标定义

**系统指标：**
- `sandbox_sessions_total`: 会话总数
- `sandbox_sessions_active`: 活跃会话数
- `sandbox_executions_total`: 执行总数
- `sandbox_execution_duration_seconds`: 执行耗时
- `sandbox_warm_pool_size`: 预热池大小
- `sandbox_runtime_cpu_usage`: 运行时 CPU 使用率
- `sandbox_runtime_memory_usage`: 运行时内存使用率

**业务指标：**
- `sandbox_cold_start_duration`: 冷启动耗时
- `sandbox_warm_start_duration`: 热启动耗时
- `sandbox_failure_rate`: 失败率
- `sandbox_timeout_rate`: 超时率

### 8.2 监控集成

**日志结构化：**
```python
import structlog

logger = structlog.get_logger()

logger.info(
    "session_created",
    session_id=session.id,
    template_id=session.template_id,
    runtime_node=session.runtime_node,
    duration_ms=100
)
```

## 9. 部署方案

### 9.1 Docker Compose 部署（开发/小规模）
```yaml
version: '3.8'

services:
  control-plane:
    build: ./control-plane
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=mysql+aiomysql://sandbox:sandbox_pass@mariadb:3306/sandbox
      - S3_ENDPOINT=http://minio:9000
      - RUNTIME_MODE=docker
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    depends_on:
      - mariadb
      - minio

  mariadb:
    image: mariadb:11.2
    ports:
      - "3306:3306"
    environment:
      - MYSQL_ROOT_PASSWORD=root_password
      - MYSQL_DATABASE=sandbox
      - MYSQL_USER=sandbox
      - MYSQL_PASSWORD=sandbox_pass
    volumes:
      - mariadb_data:/var/lib/mysql
      - ./init.sql:/docker-entrypoint-initdb.d/01-init.sql:ro
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
      - --max-connections=500
      - --innodb-buffer-pool-size=256M

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=admin
      - MINIO_ROOT_PASSWORD=password
    volumes:
      - minio_data:/data

volumes:
  mariadb_data:
  minio_data:
```

数据库初始化脚本 `init.sql`:
```sql
-- 创建数据库
CREATE DATABASE IF NOT EXISTS sandbox CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE sandbox;

-- 会话表
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(64) PRIMARY KEY,
    template_id VARCHAR(64) NOT NULL,
    status ENUM('creating', 'running', 'completed', 'failed', 'timeout', 'terminated') NOT NULL,
    runtime_type ENUM('docker', 'kubernetes') NOT NULL,
    runtime_node VARCHAR(128),           -- 当前运行的节点（可为空，支持会话迁移）
    container_id VARCHAR(128),           -- 当前容器 ID
    pod_name VARCHAR(128),               -- 当前 Pod 名称
    workspace_path VARCHAR(256),         -- S3 路径：s3://bucket/sessions/{session_id}/
    resources_cpu VARCHAR(16),
    resources_memory VARCHAR(16),
    resources_disk VARCHAR(16),
    env_vars JSON,
    timeout INT NOT NULL DEFAULT 300,
    last_activity_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- 最后活动时间（用于自动清理）
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    INDEX idx_status (status),
    INDEX idx_template (template_id),
    INDEX idx_created (created_at),
    INDEX idx_runtime_node (runtime_node),  -- 支持节点故障时查询会话
    INDEX idx_last_activity (last_activity_at)  -- 支持自动清理查询
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 执行记录表
CREATE TABLE IF NOT EXISTS executions (
    id VARCHAR(64) PRIMARY KEY,
    session_id VARCHAR(64) NOT NULL,
    code TEXT NOT NULL,
    language VARCHAR(16) NOT NULL,
    status ENUM('pending', 'running', 'completed', 'failed', 'timeout', 'crashed') NOT NULL,
    stdout MEDIUMTEXT,
    stderr MEDIUMTEXT,
    exit_code INT,
    execution_time FLOAT,
    artifacts JSON,  -- Artifact 对象数组: [{"path": "...", "size": 123, "mime_type": "...", ...}]
    retry_count INT DEFAULT 0,  -- 重试次数
    last_heartbeat_at TIMESTAMP NULL,  -- 最后心跳时间
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    INDEX idx_session (session_id),
    INDEX idx_status (status),
    INDEX idx_created (created_at),
    INDEX idx_last_heartbeat (last_heartbeat_at)  -- 支持心跳超时检测
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 模板表
CREATE TABLE IF NOT EXISTS templates (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE,
    image VARCHAR(256) NOT NULL,
    base_image VARCHAR(256),
    pre_installed_packages JSON,
    default_resources_cpu VARCHAR(16),
    default_resources_memory VARCHAR(16),
    default_resources_disk VARCHAR(16),
    security_context JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入默认模板
INSERT INTO templates (id, name, image, base_image, default_resources_cpu, default_resources_memory, default_resources_disk) VALUES
('python-basic', 'Python Basic', 'sandbox-python:3.11-basic', 'python:3.11-slim', '1', '512Mi', '1Gi'),
('python-datascience', 'Python Data Science', 'sandbox-python:3.11-datascience', 'python:3.11-slim', '2', '2Gi', '5Gi'),
('nodejs-basic', 'Node.js Basic', 'sandbox-nodejs:20-basic', 'node:20-alpine', '1', '512Mi', '1Gi')
ON DUPLICATE KEY UPDATE name=VALUES(name);
```


### 9.2 Kubernetes 部署（生产环境）

**MariaDB 部署：**
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mariadb
  namespace: sandbox-system
spec:
  serviceName: mariadb
  replicas: 1
  selector:
    matchLabels:
      app: mariadb
  template:
    metadata:
      labels:
        app: mariadb
    spec:
      containers:
      - name: mariadb
        image: mariadb:11.2
        ports:
        - containerPort: 3306
          name: mysql
        env:
        - name: MYSQL_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mariadb-secret
              key: root-password
        - name: MYSQL_DATABASE
          value: sandbox
        - name: MYSQL_USER
          value: sandbox
        - name: MYSQL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mariadb-secret
              key: user-password
        volumeMounts:
        - name: mariadb-storage
          mountPath: /var/lib/mysql
        - name: init-scripts
          mountPath: /docker-entrypoint-initdb.d
        resources:
          requests:
            cpu: "500m"
            memory: "1Gi"
          limits:
            cpu: "2"
            memory: "4Gi"
        livenessProbe:
          exec:
            command:
            - mysqladmin
            - ping
            - -h
            - localhost
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - mysql
            - -h
            - localhost
            - -u
            - sandbox
            - -p${MYSQL_PASSWORD}
            - -e
            - SELECT 1
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: init-scripts
        configMap:
          name: mariadb-init-scripts
  volumeClaimTemplates:
  - metadata:
      name: mariadb-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 50Gi
---
apiVersion: v1
kind: Service
metadata:
  name: mariadb
  namespace: sandbox-system
spec:
  selector:
    app: mariadb
  ports:
  - port: 3306
    targetPort: 3306
  clusterIP: None
---
apiVersion: v1
kind: Secret
metadata:
  name: mariadb-secret
  namespace: sandbox-system
type: Opaque
data:
  root-password: cm9vdF9wYXNzd29yZF9jaGFuZ2VfbWU=  # Base64 encoded
  user-password: c2FuZGJveF9wYXNzd29yZF9jaGFuZ2VfbWU=
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: mariadb-init-scripts
  namespace: sandbox-system
data:
  01-init.sql: |
    CREATE DATABASE IF NOT EXISTS sandbox CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
    USE sandbox;

    CREATE TABLE IF NOT EXISTS sessions (
        id VARCHAR(64) PRIMARY KEY,
        template_id VARCHAR(64) NOT NULL,
        status ENUM('creating', 'running', 'completed', 'failed', 'timeout', 'terminated') NOT NULL,
        runtime_type ENUM('docker', 'kubernetes') NOT NULL,
        runtime_node VARCHAR(128),           -- 当前运行的节点（可为空，支持会话迁移）
        container_id VARCHAR(128),           -- 当前容器 ID
        pod_name VARCHAR(128),               -- 当前 Pod 名称
        workspace_path VARCHAR(256),         -- S3 路径：s3://bucket/sessions/{session_id}/
        resources_cpu VARCHAR(16),
        resources_memory VARCHAR(16),
        resources_disk VARCHAR(16),
        env_vars JSON,
        timeout INT NOT NULL DEFAULT 300,
        last_activity_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- 最后活动时间（用于自动清理）
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        completed_at TIMESTAMP NULL,
        INDEX idx_status (status),
        INDEX idx_template (template_id),
        INDEX idx_created (created_at),
        INDEX idx_runtime_node (runtime_node),  -- 支持节点故障时查询会话
        INDEX idx_last_activity (last_activity_at)  -- 支持自动清理查询
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**管理中心部署：**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sandbox-control-plane
  namespace: sandbox-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: control-plane
  template:
    metadata:
      labels:
        app: control-plane
    spec:
      serviceAccountName: sandbox-controller
      initContainers:
      - name: wait-for-mariadb
        image: mariadb:11.2
        command:
        - sh
        - -c
        - |
          until mysql -h mariadb -u sandbox -p${MYSQL_PASSWORD} -e "SELECT 1" 2>/dev/null; do
            echo "Waiting for MariaDB..."
            sleep 2
          done
        env:
        - name: MYSQL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mariadb-secret
              key: user-password
      containers:
      - name: control-plane
        image: sandbox-control-plane:v1.0.0
        ports:
        - containerPort: 8000
        env:
        - name: DATABASE_URL
          value: "mysql+aiomysql://sandbox:$(MYSQL_PASSWORD)@mariadb:3306/sandbox"
        - name: MYSQL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mariadb-secret
              key: user-password
        - name: RUNTIME_MODE
          value: "kubernetes"
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "2"
            memory: "2Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: control-plane-service
  namespace: sandbox-system
spec:
  selector:
    app: control-plane
  ports:
  - port: 80
    targetPort: 8000
  type: LoadBalancer
```

**RBAC 配置：**
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sandbox-controller
  namespace: sandbox-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: sandbox-controller-role
rules:
- apiGroups: ["sandbox.ai"]
  resources: ["sandboxes", "sandboxtemplates"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: [""]
  resources: ["pods", "pods/log"]
  verbs: ["get", "list", "watch", "create", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: sandbox-controller-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: sandbox-controller-role
subjects:
- kind: ServiceAccount
  name: sandbox-controller
  namespace: sandbox-system
```

**HPA 自动扩缩容：**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: control-plane-hpa
  namespace: sandbox-system
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: sandbox-control-plane
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
