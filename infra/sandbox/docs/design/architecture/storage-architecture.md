# 存储架构 - MinIO + s3fs

## 架构概述

系统使用 MinIO (S3-compatible 对象存储) 和 s3fs (FUSE 文件系统) 来实现 workspace 文件存储。

### 数据流

```
用户上传文件 → Control Plane → S3 API 写入 MinIO
                                      ↓
                               s3fs-fuse 挂载 ─┘
                                              │
                                    ┌────────▼────────┐
                                    │  Executor Pod   │
                                    │  /workspace/    │
                                    └─────────────────┘
```

## s3fs 挂载方案

### 工作原理

s3fs-fuse 允许将 S3 bucket 直接挂载为本地目录：

```bash
s3fs bucket-name /workspace -o url=http://minio:9000 \
    -o use_path_request_style -o allow_other
```

### Kubernetes 实现

在 Executor Pod 中，s3fs 通过启动脚本在容器内运行：

1. 挂载整个 S3 bucket 到 `/mnt/s3-root`
2. 使用 `mount --bind` 将 session 目录覆盖到 `/workspace`
3. 启动 executor 进程

```bash
# 启动脚本片段
s3fs {bucket} /mnt/s3-root \
    -o url={minio_url} \
    -o use_path_request_style \
    -o allow_other \
    -o uid=1000 \
    -o gid=1000 \
    -o passwd_file=/etc/s3fs-passwd/s3fs-passwd &

# 等待挂载完成
sleep 2

# 创建 session 目录
SESSION_PATH="/mnt/s3-root/sessions/{session_id}"
mkdir -p "$SESSION_PATH"

# 使用 bind mount 覆盖 /workspace
mount --bind "$SESSION_PATH" /workspace

# 启动 executor
exec gosu sandbox python -m executor.interfaces.http.rest
```

### 路径映射

| MinIO 路径 | s3fs 挂载路径 | /workspace 最终路径 |
|------------|---------------|-------------------|
| `sessions/{id}/test.py` | `/mnt/s3-root/sessions/{id}/test.py` | `/workspace/test.py` ✅ |

## 配置

### Secret 配置

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: s3fs-passwd
  namespace: sandbox-system
type: Opaque
stringData:
  s3fs-passwd: |
    minioadmin:minioadmin
```

### 环境变量

```bash
# MinIO 配置
S3_ENDPOINT_URL=http://minio.sandbox-system.svc.cluster.local:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_BUCKET=sandbox-workspace
S3_REGION=us-east-1
```

## 优势

### 简化性
- 只需一个组件: s3fs-fuse
- 只需配置 MinIO 连接信息
- 不需要额外的元数据数据库

### 可维护性
- 调试简单: 文件上传后直接在 MinIO 查看
- 网络友好: s3fs-fuse 在各大发行版包管理器中可用
- 版本稳定

### 可靠性
- 无单点: 不依赖额外的元数据存储
- 一致性: MinIO 是最终真理来源
- 可恢复: MinIO 数据可直接访问

## 安全配置

### 容器安全上下文

s3fs 需要 privileged 模式进行 FUSE 挂载：

```yaml
securityContext:
  privileged: true  # s3fs 需要 FUSE 挂载
  runAsUser: 0      # root 用户进行挂载
  runAsGroup: 0
```

### 进程切换

挂载完成后，使用 gosu 切换到 sandbox 用户运行 executor：

```bash
exec gosu sandbox python -m executor.interfaces.http.rest
```

## 性能考虑

### 优化建议

1. **文件缓存**: 使用 `--use_cache` 选项
2. **并发控制**: 限制并发文件操作数量
3. **监控**: 添加文件操作延迟监控

## 故障排查

### 常见问题

1. **权限问题**: 确保 uid=1000, gid=1000 与 sandbox 用户匹配
2. **挂载超时**: 增加 sleep 时间等待挂载完成
3. **文件不可见**: 检查 SESSION_PATH 路径是否正确

### 调试命令

```bash
# 检查 s3fs 挂载
kubectl exec -it <pod> -- df -h | grep s3

# 检查 /workspace 内容
kubectl exec -it <pod> -- ls -la /workspace/

# 检查 s3fs 进程
kubectl exec -it <pod> -- ps aux | grep s3fs
```
