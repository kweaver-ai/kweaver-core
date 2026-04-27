# 沙箱平台设计总览

当前设计文档按“架构 / 模块 / 需求设计 / ADR / 历史归档”组织。

## 文档导航

### 架构

- [系统上下文](./system-context.md)
- [逻辑架构与流程](./logical-architecture.md)
- [安全与性能](./security-and-performance.md)
- [存储架构](./storage-architecture.md)

### 模块

- [Control Plane](../modules/control-plane.md)
- [Container Scheduler](../modules/container-scheduler.md)
- [Executor](../modules/executor.md)
- [Python Dependencies](../modules/python-dependencies.md)

### 需求设计

- [Session Python 依赖管理](../features/session-python-dependency-management.md)
- [MCP Server Implementation](../features/mcp-server-implementation.md)

### 其他

- [ADR 模板](../decisions/template.md)
- [历史归档：sandbox-design-v2.1](../archive/sandbox-design-v2.1.md)
- [返回文档中心](../../README.md)

## 🏗️ 架构概览

系统采用**管理中心（Control Plane）与容器调度器（Container Scheduler）分离**的云原生架构，支持 Docker 和 Kubernetes 两种部署模式。

### 核心设计原则

1. **控制平面无状态**，支持水平扩展
2. **容器调度器池化管理**，动态伸缩
3. **协议驱动的解耦设计**
4. **多层安全隔离**（容器 + Bubblewrap）
5. **异步高并发处理**

### 双层隔离架构

```
┌─────────────────────────────────────────┐
│ 宿主机 (Host)                            │
│  ├─ Docker Engine / Kubernetes          │
│  └─ 运行时管理器                         │
└─────────────────────────────────────────┘
           ↓ 创建容器
┌─────────────────────────────────────────┐
│ 容器 (Container) - 第一层隔离            │
│  ├─ 网络隔离 (NetworkMode=none)          │
│  ├─ 资源限制 (CPU/Memory/PID)            │
│  └─ 非特权用户 (sandbox:sandbox)         │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │ 执行器进程 (Executor)               │ │
│  │  - 监听管理中心的执行请求           │ │
│  │  - 调用 bwrap 启动用户代码          │ │
│  └────────────────────────────────────┘ │
│           ↓ 调用 bwrap                   │
│  ┌────────────────────────────────────┐ │
│  │ Bubblewrap 沙箱 - 第二层隔离       │ │
│  │  ├─ 新的命名空间 (PID/NET/MNT...)  │ │
│  │  ├─ 只读文件系统                    │ │
│  │  └─ seccomp 系统调用过滤            │ │
│  │                                     │ │
│  │  ┌──────────────────────────────┐  │ │
│  │  │ 用户代码进程                  │  │ │
│  │  └──────────────────────────────┘  │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

---

## 📦 项目结构

```
sandbox/
├── deploy/                   # Deployment configurations
│   ├── k8s/                  # Kubernetes manifests
│   ├── docker-compose/       # Docker Compose
│   └── helm/                 # Helm charts
│
├── sandbox_control_plane/    # FastAPI control plane
│   ├── src/
│   │   ├── application/      # Business logic
│   │   ├── domain/           # Domain models
│   │   ├── infrastructure/   # External dependencies
│   │   ├── interfaces/       # REST API
│   │   └── shared/           # Shared utilities
│   └── tests/                # Tests
│
├── sandbox_web/              # React web console
├── runtime/executor/          # Sandbox executor daemon
├── scripts/                  # Utility scripts
└── docs/                     # Documentation (this directory)
```

---

## 🔗 相关资源

- **项目主目录**: `.`
- **Control Plane 代码**: `sandbox_control_plane/src/`
- **K8s 配置文件**: `deploy/manifests/`
- **Helm Chart**: `deploy/helm/sandbox/`

---

## 📝 文档版本

- **版本**: V2.1
- **最后更新**: 2025-01-21
- **状态**: 已拆分重构

---

选择上方章节开始阅读，或返回 [文档中心](../../README.md)。
