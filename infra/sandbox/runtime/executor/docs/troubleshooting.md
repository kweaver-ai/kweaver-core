# 故障排查

本文档提供 Sandbox Executor 常见问题的诊断和解决方案。

## 目录

- [常见问题](#常见问题)
- [调试技巧](#调试技巧)
- [日志分析](#日志分析)
- [性能问题](#性能问题)
- [获取帮助](#获取帮助)

---

## 常见问题

### 1. Bubblewrap 权限错误

**错误信息**:
```
bwrap: No permissions to create new namespace
bwrap: Creating new namespace failed: Operation not permitted
```

**原因**: 宿主机未启用非特权用户命名空间支持。

**解决方案**:

**不要使用 `--privileged`**，这会完全破坏容器安全隔离！

**正确做法 - 在宿主机启用用户命名空间**:

Ubuntu/Debian:
```bash
# 临时启用
sudo sysctl -w kernel.unprivileged_userns_clone=1

# 持久化配置
echo "kernel.unprivileged_userns_clone=1" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

CentOS/RHEL:
```bash
# 设置用户命名空间最大数量
echo "user.max_user_namespaces=10000" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

验证配置:
```bash
# 测试用户命名空间
unshare -U -m echo "User namespaces work"

# 检查系统配置
sysctl kernel.unprivileged_userns_clone  # Ubuntu/Debian
sysctl user.max_user_namespaces          # CentOS/RHEL
```

**临时替代方案（不推荐用于生产）**:
```bash
# 禁用 Bubblewrap，降级到普通 subprocess 模式
export DISABLE_BWRAP=true
```

**安全说明**:
- Bubblewrap 通过用户命名空间工作，**不需要** `--privileged`
- 使用 `--privileged` 会让容器获得宿主机的完全访问权限
- 正确配置后，容器可以在非特权模式下运行，保持双层隔离

---

### 2. 代码执行超时

**错误信息**:
```
asyncio.TimeoutError: Execution timeout
```

**原因**: 代码执行时间超过配置的超时限制。

**解决方案**:

1. 增加超时参数:
```bash
curl -X POST http://localhost:8080/execute \
  -d '{
    "timeout": 600,
    ...
  }'
```

2. 检查代码是否有死循环:
```python
# 检查是否有以下模式
while True:
    # 可能的无限循环
    pass
```

3. 使用 `__timeout` 参数:
```python
event = {
    "data": "...",
    "__timeout": 600  # 10 分钟
}
```

---

### 3. 容器间网络不通

**错误信息**:
```
Failed to report result: Connection refused
Failed to callback Control Plane: HTTPConnectionError
```

**原因**: Executor 无法访问 Control Plane 服务。

**解决方案**:

1. 检查 `CONTROL_PLANE_URL` 配置:
```bash
# Docker 容器内访问宿主机
CONTROL_PLANE_URL=http://host.docker.internal:8000

# Docker Compose 内部网络
CONTROL_PLANE_URL=http://control-plane:8000

# Kubernetes 服务发现
CONTROL_PLANE_URL=http://control-plane-service:8000
```

2. 验证网络连通性:
```bash
# 进入容器测试
docker exec -it sandbox-executor curl http://control-plane:8000/health

# 检查 DNS 解析
docker exec -it sandbox-executor nslookup control-plane
```

3. 使用 Docker 自定义网络:
```yaml
networks:
  sandbox-network:
    driver: bridge
```

---

### 4. macOS 上 Bubblewrap 不可用

**错误信息**:
```
RuntimeError: Bubblewrap (bwrap) is not installed
```

**原因**: macOS 不支持 Bubblewrap，应该使用 Seatbelt (sandbox-exec)。

**解决方案**:

Executor 会自动检测并切换到 macOS Seatbelt。如果没有自动切换：

```python
# 手动指定隔离适配器
from executor.infrastructure.isolation.macseatbelt import SeatbeltRunner

runner = SeatbeltRunner(workspace_path="/workspace")
```

验证 Seatbelt 可用性:
```bash
which sandbox-exec
# 应该输出: /usr/bin/sandbox-exec
```

---

### 5. 模块导入错误

**错误信息**:
```
ModuleNotFoundError: No module named 'executor'
```

**原因**: `PYTHONPATH` 未正确设置。

**解决方案**:

```bash
# 设置 PYTHONPATH
export PYTHONPATH="$(cd .. && pwd)"

# 或在启动时指定
PYTHONPATH=/path/to/runtime python3 -m executor.interfaces.http.rest
```

---

### 6. 工作区权限错误

**错误信息**:
```
PermissionError: [Errno 13] Permission denied: '/workspace'
```

**原因**: 工作目录权限不足。

**解决方案**:

```bash
# 修改工作目录权限
chmod 755 /workspace

# 或使用其他可写目录
export WORKSPACE_PATH=/tmp/sandbox_workspace
mkdir -p /tmp/sandbox_workspace
chmod 777 /tmp/sandbox_workspace
```

---

### 7. 内存不足

**错误信息**:
```
MemoryError: Cannot allocate memory
Killed
```

**原因**: 执行代码超出内存限制。

**解决方案**:

1. 增加内存限制:
```bash
export MAX_MEMORY_MB=1024
```

2. 优化代码内存使用:
```python
# 使用生成器而非列表
def process_large_data():
    for item in large_dataset:
        yield process(item)

# 及时释放内存
del large_list
import gc
gc.collect()
```

3. Docker 增加容器内存:
```bash
docker run -m 2g sandbox-executor:v1.0
```

---

## 调试技巧

### 启用调试日志

```bash
export LOG_LEVEL=DEBUG
python3 -m executor.interfaces.http.rest
```

### 使用 Python 调试器

```python
import pdb; pdb.set_trace()

# 或使用 ipdb（推荐）
import ipdb; ipdb.set_trace()
```

### 进入容器调试

```bash
# Docker
docker exec -it sandbox-executor /bin/bash

# Kubernetes
kubectl exec -it deployment/sandbox-executor -- /bin/bash

# 在容器内检查
ps aux | grep python
ls -la /workspace
cat /proc/1/status
```

### 测试隔离适配器

```bash
# 测试 Bubblewrap
bwrap --ro-bind /usr /usr --unshare-all /bin/echo "Hello"

# 测试 Seatbelt (macOS)
sandbox-exec /bin/echo "Hello"
```

---

## 日志分析

### 查看日志

```bash
# Docker
docker logs -f sandbox-executor

# Docker Compose
docker-compose logs -f executor

# Kubernetes
kubectl logs -f deployment/sandbox-executor

# 持久化日志
tail -f /var/log/executor/executor.log
```

### 日志级别

| 级别 | 说明 | 使用场景 |
|------|------|----------|
| DEBUG | 详细调试信息 | 开发调试 |
| INFO | 一般信息 | 正常运行 |
| WARNING | 警告信息 | 需要注意的情况 |
| ERROR | 错误信息 | 错误和异常 |

### 常见日志模式

**成功执行**:
```
[INFO] Starting execution: exec_001
[INFO] Execution completed: exec_001 (duration: 1234ms)
```

**超时错误**:
```
[ERROR] Execution timeout: exec_001 (timeout: 300s)
[INFO] Process killed: exec_001
```

**隔离错误**:
```
[ERROR] Isolation setup failed: bwrap not available
[INFO] Falling back to Seatbelt runner
```

---

## 性能问题

### 高 CPU 使用

**诊断**:
```bash
# 查看 CPU 使用
docker stats sandbox-executor

# 或
top -p $(pgrep -f executor.interfaces.http.rest)
```

**解决方案**:
1. 减少并发执行数
2. 增加 `MAX_CONCURRENT_EXECUTIONS`
3. 优化执行代码

### 高内存使用

**诊断**:
```bash
# 查看内存使用
docker stats sandbox-executor

# 或
ps aux | grep python
```

**解决方案**:
1. 定期清理工作区
2. 减少 `MAX_MEMORY_MB`
3. 使用内存分析工具:
```bash
pip install memory_profiler
python -m memory_profiler executor/interfaces/http/rest.py
```

### 执行慢

**诊断**:
```bash
# 使用 time 测量执行时间
time curl http://localhost:8080/execute -d '{...}'
```

**解决方案**:
1. 检查隔离适配器版本
2. 使用更快的存储 (SSD)
3. 启用异步并发:
```bash
# 验证异步是否工作
for i in {1..5}; do
  curl http://localhost:8080/execute -d '{...}' &
done
wait
# 如果异步正常，总时间应该接近单个请求时间
```

---

## 获取帮助

### 检查版本

```bash
# 检查 Executor 版本
curl http://localhost:8080/info

# 检查 Bubblewrap 版本
bwrap --version

# 检查 Python 版本
python --version
```

### 收集诊断信息

```bash
# 系统信息
uname -a
docker version
kubectl version

# 服务状态
curl http://localhost:8080/health
curl http://localhost:8080/info

# 日志（最近 100 行）
docker logs --tail 100 sandbox-executor
```

### 报告问题

在报告问题时，请包含：

1. 完整的错误信息
2. Executor 版本 (`/info` 端点输出)
3. 操作系统和版本
4. 最小的可复现代码
5. 相关日志

```bash
# 生成问题报告
cat << 'EOF' > bug_report.txt
**Version**: $(curl -s http://localhost:8080/info | jq -r '.version')
**OS**: $(uname -a)
**Error**:
\`\`\`
$(curl -X POST http://localhost:8080/execute -d '{...}' 2>&1)
\`\`\`

**Logs**:
\`\`\`
$(docker logs --tail 50 sandbox-executor)
\`\`\`
EOF
```

---

## 相关文档

- [用户命名空间配置](user-namespace-guide.md) - 双层隔离原理和用户命名空间配置
- [配置说明](configuration.md) - 环境变量配置
- [部署指南](deployment.md) - 部署相关问题
- [开发指南](development.md) - 调试技巧
