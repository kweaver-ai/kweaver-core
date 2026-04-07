# 用户命名空间配置指南

本文档详细解释 Bubblewrap 如何通过用户命名空间实现非特权容器隔离，以及如何正确配置宿主机以支持双层隔离架构。

## 目录

- [什么是双层隔离](#什么是双层隔离)
- [用户命名空间原理](#用户命名空间原理)
- [宿主机配置](#宿主机配置)
- [验证配置](#验证配置)
- [常见问题](#常见问题)
- [安全架构详解](#安全架构详解)

---

## 什么是双层隔离

Sandbox Executor 实现了双层隔离架构来确保安全执行不受信任的代码：

```
┌─────────────────────────────────────────────────────────────┐
│ 第一层：Docker 容器隔离                                       │
├─────────────────────────────────────────────────────────────┤
│ • 非特权用户 (UID:GID=1000:1000)                              │
│ • CapDrop=ALL（移除所有 Linux capabilities）                  │
│ • no-new-privileges（禁止权限提升）                           │
│ • seccomp=default（系统调用过滤）                             │
│ • 独立的网络命名空间                                          │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ 第二层：Bubblewrap 进程隔离                                   │
├─────────────────────────────────────────────────────────────┤
│ • 通过用户命名空间创建隔离环境                                │
│ • 只读文件系统绑定（除了 /workspace）                         │
│ • 独立的 PID/NET/MNT 命名空间                                │
│ • 无网络访问（--unshare-net）                                │
│ • 清空环境变量（--clearenv）                                 │
└─────────────────────────────────────────────────────────────┘
```

### 为什么需要两层隔离？

1. **防御深度**：即使一层被攻破，另一层仍能提供保护
2. **最小权限原则**：每层只提供执行所需的最小权限
3. **不同的威胁模型**：
   - Docker 层：防止容器逃逸到宿主机
   - bwrap 层：防止恶意代码逃逸到容器

---

## 用户命名空间原理

### 什么是用户命名空间？

用户命名空间（User Namespace）是 Linux 内核 3.8+ 引入的一项特性，允许**非特权用户**创建隔离的 UID/GID 空间。

### 关键特性

```bash
# 在宿主机上（root）
$ id
uid=0(root) gid=0(root)

# 在用户命名空间内（非特权）
$ unshare -U -m
$ id
uid=0(root) gid=0(root)  # 映射到容器内的 root
```

**神奇之处**：
- 容器内的 `uid=0(root)` 在宿主机看来是一个**非特权用户**
- 容器内认为自己有 root 权限，可以执行需要 root 的操作（如创建命名空间）
- 但在宿主机层面，所有操作都以非特权用户身份执行

### Bubblewrap 如何使用用户命名空间

Bubblewrap 的核心工作流程：

```python
# Bubblewrap 执行流程（简化）
1. bwrap 通过 unshare() 系统调用创建用户命名空间
2. 在新的命名空间内，bwrap 映射 UID 0 到非特权 UID
3. 由于在命名空间内 UID=0，可以创建其他命名空间（PID/NET/MNT）
4. 创建隔离的文件系统视图
5. 执行用户代码
```

**关键点**：步骤 3 需要的"权限"只在用户命名空间内有效，不影响宿主机。

---

## 宿主机配置

### Linux 内核要求

- **最低内核版本**: 3.8+（用户命名空间支持）
- **推荐内核版本**: 4.19+（更完善的用户命名空间支持）

检查内核版本：
```bash
uname -r
```

### 各发行版配置

#### Ubuntu / Debian

Ubuntu 默认禁用非特权用户命名空间（Ubuntu 23.04+ 开始默认启用）。

```bash
# 检查当前配置
sysctl kernel.unprivileged_userns_clone

# 临时启用（重启后失效）
sudo sysctl -w kernel.unprivileged_userns_clone=1

# 持久化配置
echo "kernel.unprivileged_userns_clone=1" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

**Ubuntu 版本说明**：
- Ubuntu 23.04+: 默认已启用，无需配置
- Ubuntu 22.04 及以下: 需要手动启用
- Docker Desktop (Ubuntu): 通常已自动配置

#### CentOS / RHEL / Fedora

```bash
# 检查当前配置
sysctl user.max_user_namespaces

# 设置最大命名空间数量
echo "user.max_user_namespaces=10000" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

**CentOS 8+/RHEL 8+**: 默认已启用，通常无需配置。

#### Arch Linux

```bash
# 检查当前配置
sysctl kernel.unprivileged_userns_clone

# 启用（如果未启用）
sudo sysctl -w kernel.unprivileged_userns_clone=1
echo "kernel.unprivileged_userns_clone=1" | sudo tee -a /etc/sysctl.conf
```

#### Docker Desktop (macOS/Windows)

Docker Desktop 使用 Linux 虚拟机，通常已默认启用用户命名空间。

验证：
```bash
# 进入 Docker Desktop 虚拟机
docker run -it --rm alpine sysctl kernel.unprivileged_userns_clone
# 输出应为: kernel.unprivileged_userns_clone = 1
```

### Kubernetes 节点配置

在**所有 Kubernetes 节点**上执行：

```bash
# Ubuntu/Debian
cat <<EOF | sudo tee /etc/sysctl.d/99-sandbox-userns.conf
kernel.unprivileged_userns_clone = 1
EOF
sudo sysctl -p /etc/sysctl.d/99-sandbox-userns.conf

# CentOS/RHEL
cat <<EOF | sudo tee /etc/sysctl.d/99-sandbox-userns.conf
user.max_user_namespaces = 10000
EOF
sudo sysctl -p /etc/sysctl.d/99-sandbox-userns.conf
```

验证所有节点：
```bash
kubectl get nodes -o name | xargs -I {} kubectl debug {} -- sysctl kernel.unprivileged_userns_clone
```

---

## 验证配置

### 1. 验证用户命名空间支持

```bash
# 测试创建用户命名空间
unshare -U -m --map-root-user echo "User namespaces work"

# 如果成功，无错误输出
# 如果失败，显示 "unshare: unshare failed: Operation not permitted"
```

### 2. 验证 Bubblewrap 功能

```bash
# 测试 Bubblewrap 基本功能
bwrap --ro-bind /usr /usr --unshare-all /bin/echo "Bubblewrap works"

# 测试用户命名空间隔离
bwrap --ro-bind /usr /usr --unshare-all --unshare-user /bin/id
# 应该显示: uid=0(root) gid=0(root) groups=0(root)
```

### 3. 在容器内验证

```bash
# 启动测试容器
docker run --rm -it \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  --security-opt seccomp=default \
  ubuntu:22.04 bash

# 在容器内测试
apt-get update && apt-get install -y bubblewrap
bwrap --ro-bind /usr /usr --unshare-all /bin/echo "Success!"
```

### 4. 验证 Sandbox Executor

```bash
# 启动 executor
docker run -d \
  --name test-executor \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  --security-opt seccomp=default \
  -p 8080:8080 \
  -e DISABLE_BWRAP=false \
  sandbox-executor:latest

# 检查日志
docker logs test-executor

# 应该看到类似输出：
# [INFO] Using BubblewrapRunner for Linux isolation
# [INFO] Bubblewrap verified, version=bwrap 1.x.x
```

---

## 常见问题

### Q1: 为什么不使用 `--privileged`？

**A**: `--privileged` 会赋予容器**完整的宿主机访问权限**，完全破坏 Docker 的隔离边界：

```yaml
# ❌ 错误做法
docker run --privileged ...
# 容器可以：
# - 访问所有设备（/dev/sda, /dev/mem...）
# - 修改宿主机网络配置
# - 逃逸到宿主机
# - 绕过所有 seccomp/apparmor 限制
```

使用用户命名空间才是正确做法：
```yaml
# ✅ 正确做法
docker run \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  --security-opt seccomp=default \
  ...
```

### Q2: 用户命名空间有性能开销吗？

**A**: 开销极小（<1%）。

用户命名空间通过内核级别的 UID/GID 映射表工作，性能影响几乎可以忽略。

基准测试：
```bash
# 不使用命名空间
time for i in {1..1000}; do unshare -m echo "test"; done
# real: 0m2.3s

# 使用命名空间
time for i in {1..1000}; do unshare -U -m echo "test"; done
# real: 0m2.4s  (~4% 差异，可接受)
```

### Q3: macOS 上如何使用？

**A**: macOS 不支持用户命名空间，Executor 会自动降级到 Seatbelt (sandbox-exec)：

```python
# 代码自动检测
if platform.system() == "Linux":
    runner = BubblewrapRunner()  # 使用 bwrap
elif platform.system() == "Darwin":
    runner = MacSeatbeltRunner()  # 使用 sandbox-exec
```

Seatbelt 是 macOS 原生的沙箱机制，提供类似的隔离能力。

### Q4: 容器内如何检查命名空间支持？

**A**: 执行以下检查：

```bash
# 在容器内
# 1. 检查是否在用户命名空间内
readlink /proc/self/ns/user
# 输出: user:[4026531837] (如果在命名空间内)

# 2. 检查 UID 映射
cat /proc/self/uid_map
# 输出类似: 0 1000 1 (容器内 UID 0 映射到宿主机 UID 1000)

# 3. 测试创建新命名空间
unshare -U true && echo "Success" || echo "Failed"
```

### Q5: Kubernetes 上如何配置？

**A**: 使用安全上下文，**不要**使用 `privileged: true`：

```yaml
# ✅ 正确配置
securityContext:
  runAsUser: 1000
  runAsGroup: 1000
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL

# ❌ 错误配置
securityContext:
  privileged: true  # 不要这样做！
```

### Q6: 如何在禁用用户命名空间的环境中使用？

**A**: 有两个选择：

1. **临时禁用 Bubblewrap**（不推荐，失去一层隔离）：
```bash
export DISABLE_BWRAP=true
```

2. **使用 gVisor/Firecracker** 等更强的隔离方案：
```yaml
# gVisor 示例
runtimeClassName: gvisor
```

---

## 安全架构详解

### 当前实现的隔离层级

```
┌──────────────────────────────────────────────────────────┐
│ 宿主机 (Host)                                            │
│ Kernel: Linux 4.19+                                      │
│ Security: SELinux/AppArmor + seccomp                     │
└──────────────────────────────────────────────────────────┘
              │
              │ Docker 容器边界
              ▼
┌──────────────────────────────────────────────────────────┐
│ Docker 容器 (sandbox-executor)                           │
│ ├─ User: UID:GID=1000:1000 (非特权用户)                  │
│ ├─ Capabilities: ALL dropped                             │
│ ├─ Security Options:                                     │
│ │   ├─ no-new-privileges (禁止权限提升)                  │
│ │   └─ seccomp=default (系统调用过滤)                    │
│ └─ Network: bridge/veth (隔离网络栈)                      │
└──────────────────────────────────────────────────────────┘
              │
              │ Bubblewrap 进程边界
              ▼
┌──────────────────────────────────────────────────────────┐
│ Bubblewrap 沙箱进程                                       │
│ ├─ User Namespace: UID/GID remapping                     │
│ ├─ PID Namespace: 独立进程树                             │
│ ├─ MNT Namespace: 独立文件系统视图                       │
│ ├─ NET Namespace: 网络隔离 (--unshare-net)               │
│ ├─ UTS Namespace: 独立 hostname                          │
│ ├─ IPC Namespace: 独立 IPC                               │
│ ├─ Filesystem: 只读绑定（除了 /workspace）               │
│ └─ Environment: 清空环境变量                             │
└──────────────────────────────────────────────────────────┘
              │
              │ 用户代码执行
              ▼
┌──────────────────────────────────────────────────────────┐
│ 执行的用户代码                                            │
│ - 在隔离的命名空间内运行                                  │
│ - 无法访问容器外部文件系统                               │
│ - 无法访问网络                                            │
│ - 无法创建新进程（PID namespace 隔离）                    │
└──────────────────────────────────────────────────────────┘
```

### 安全特性对比

| 特性 | Docker 层 | Bubblewrap 层 | 说明 |
|------|-----------|---------------|------|
| 用户隔离 | ✅ UID 1000 | ✅ UID remapping | 双重 UID 隔离 |
| 文件系统隔离 | ✅ 容器根文件系统 | ✅ 只读绑定 | bwrap 额外限制可写区域 |
| 网络隔离 | ✅ bridge/veth | ✅ --unshare-net | bwrap 完全禁止网络访问 |
| 进程隔离 | ✅ PID namespace | ✅ 独立 PID tree | bwrap 防止进程间通信 |
| 权限控制 | ✅ CapDrop=ALL | ✅ UID 0 in NS | 双重权限限制 |
| 系统调用过滤 | ✅ seccomp | ⚠️ 继承容器 | bwrap 依赖容器层过滤 |

### 逃逸难度评估

```
完全隔离 ←━━━━━━━━━━━━━━━━━━━━━━━━━━━━━→ 无隔离

攻击者需要突破：

1. Docker 容器边界（难度：高）
   ├─ 非特权用户运行
   ├─ 所有 capabilities 移除
   ├─ seccomp 系统调用过滤
   └─ AppArmor/SELinux 策略

2. Bubblewrap 进程边界（难度：高）
   ├─ 用户命名空间 UID 映射
   ├─ 多个命名空间隔离（PID/NET/MNT...）
   ├─ 只读文件系统绑定
   └─ 网络完全隔离

总逃逸难度：极高（需要两个独立漏洞）
```

---

## 相关文档

- [部署指南](deployment.md) - 部署和配置
- [故障排查](troubleshooting.md) - 常见问题解决
- [配置说明](configuration.md) - 环境变量和配置
- [架构文档](architecture.md) - 系统架构设计
