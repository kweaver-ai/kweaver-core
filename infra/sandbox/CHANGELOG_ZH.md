# 更新日志

本分支 (`feature/803264`) 中新增的所有功能和特性记录如下。

## [0.3.3]

### 🚀 新功能

- **K8s Session 的 Control Plane 接管恢复**
  - 新增 control plane 启动接管流程，可在 control plane 重启或升级后识别由上一代 control-plane pod 创建的活跃 K8s session pod
  - 新增 control plane Pod 身份注入能力（`POD_NAME`、`POD_UID`）以及 executor Pod owner reference 绑定能力，使重建后的 session pod 能重新归属到当前 control plane
  - 新增升级接管场景下对中断执行任务的处理逻辑，当必须重建 executor pod 时，会将进行中的 execution 标记为失败

### 🐛 问题修复

- **Control Plane 重启后的 Session 恢复稳定性**
  - 修复启动阶段状态同步恢复 session 时无法通过真实模板仓储解析模板的问题，避免因模板解析失败导致恢复链路直接失败
  - 修复同名 Pod 重建时与旧 `Terminating` Pod 冲突的问题，在重试创建前会等待旧 Pod 从 K8s API 中彻底删除
  - 修复 takeover recovery 成功后仍沿用旧 `last_activity_at` 导致被空闲清理任务立即回收的问题，恢复成功后会刷新会话活动时间

- **K8s Session 调度资源请求**
  - 调整 K8s 沙箱 session Pod 的 CPU 和内存 request 为 0，同时保持运行时 limits 不变
  - 降低资源紧张集群中由于每个 session 预留 request 导致的 Pod 调度与启动失败概率

### 🔧 改进

- 为启动时状态同步补充 direct session/execution repository 支持，使接管逻辑可分页扫描全部活跃 session，并在不依赖请求级仓储注入的情况下持久化中断 execution 状态
- 补充 K8s takeover、owner reference 识别、旧 Pod 重建冲突、依赖注入 wiring 以及恢复后活动时间刷新的单元测试覆盖

### 📚 文档

- 新增 control-plane 与 executor 生命周期绑定在重启、升级场景下的 PRD 和设计文档
- 新增 `sandbox_local` Helm Chart，补充本地部署模板、组件元数据、RBAC 配置与运维说明，便于本地打包和环境初始化

---

*发布于 2026-04-20*

## [0.3.2]

### 🚀 新功能

- **压缩包上传与自动解压**
  - 为 session workspace 新增 ZIP 压缩包上传能力
  - 新增将压缩包自动解压到指定 workspace 相对路径的能力
  - 上传响应中增加覆盖控制、冲突跳过统计和解压结果元数据
  - 增加压缩包路径安全校验，拒绝非法条目与路径穿越内容

- **Shell 执行与工作目录支持**
  - 在 control plane、executor 和 Web UI 中新增 `language=shell` 支持
  - 为 `execute` 与 `execute-sync` 请求新增可选 `working_directory`
  - 增加对误写 `bash/sh` 前缀命令的 shell 规范化处理
  - 补充 shell 相对路径执行与链式命令场景的端到端测试覆盖

### 📚 文档

- 新增 session archive upload 和 shell execution 的 PRD 与设计文档
- 更新 sandbox OpenAPI 文档，补充压缩包上传与 shell 执行请求变更

### 🔧 改进

- **Helm Chart 镜像提取兼容性**
  - 将 sandbox Helm Chart 镜像配置统一规范到顶层 `image` values 结构
  - 将 control plane、web、MariaDB、MinIO、默认模板和 BusyBox 镜像引用改为使用 `image.<name>.repository` 与 `image.<name>.tag`
  - 移除 Web initContainer 中硬编码的 BusyBox 镜像，使离线打包工具可从 chart values 中提取该镜像
  - 同步更新 Helm README、本地安装覆盖参数与组件镜像元数据，保持与新镜像结构一致

---

*发布于 2026-04-07*

## [0.3.1]

### 🚀 新功能

- **数据库名称升级处理**
  - 将 control plane 默认数据库名从 `adp` 调整为 `kweaver`
  - 新增启动时升级逻辑，可检测旧库 `adp` 并将其中表迁移到 `kweaver`
  - 将运行期数据库连接统一规范到 `kweaver`，避免服务启动时重新创建旧库

### 🔧 改进

- 补充旧数据库重命名流程与运行期数据库 URL 规范化的单元测试覆盖

---

*发布于 2026-04-02*

## [0.3.0]

### 🚀 新功能

- **会话级 Python 依赖管理**
  - 新增会话级依赖配置与安装状态跟踪
  - 创建会话时支持后台触发初始依赖同步
  - 新增同步接口 `POST /api/v1/sessions/{session_id}/dependencies/install`
  - 会话响应中增加已安装依赖明细与错误信息

- **运行时执行器依赖同步**
  - 新增执行器侧会话配置同步服务，用于完整对齐依赖状态
  - 重新安装依赖前自动清理隔离依赖目录
  - 增加 uv/virtualenv 场景下的 pip Python 可执行文件探测
  - 改进 bwrap、subprocess 和 macOS seatbelt 隔离后端兼容性

- **会话管理界面**
  - 会话列表页新增依赖安装操作与状态展示
  - 前端新增手动安装依赖所需的 API 类型与 hooks
  - 会话详情中可查看依赖安装进度与失败信息

- **数据库与升级支持**
  - 新增 `0.3.0` 对应的 MariaDB 和 DM8 初始化 SQL
  - 新增服务启动时的数据库 schema 自动补齐能力，便于老版本升级

### 🔧 改进

- 统一 REST OpenAPI 文档为 `docs/api/rest/sandbox-openapi.json`
- 扩展会话 DTO、持久化模型与 API 响应中的依赖元数据
- 补充初始同步、手动安装和依赖执行链路的集成测试与单元测试

### 📚 文档

- 新增会话 Python 依赖管理的设计文档与 PRD
- 按架构、开发、运维、产品维度重组仓库文档结构
- 为同步执行接口补充独立 OpenAPI 描述

---

*发布于 2026-03-11*

## [0.2.1]

### 🐛 问题修复

- **K8s 调度器容器恢复能力**
  - 将 Pod `restartPolicy` 从 `Never` 改为 `Always`
  - 确保容器退出后（包括退出码 0）自动重启
  - 修复 s3fs 挂载断开导致 runtime 永久不可用的问题

### 🚀 新功能

- **心跳服务可靠性**
  - 改进心跳服务，增加更好的错误处理
  - 添加心跳功能的完整测试覆盖

- **状态同步服务**
  - 控制平面 URL 可通过环境变量配置
  - 添加状态同步服务的设置初始化

- **回调客户端**
  - 添加 JSON 清理功能，处理非标准浮点值（NaN、Infinity）
  - 确保回调响应的正确 JSON 序列化

- **运行时执行器**
  - 命令执行改为异步模式，提升性能
  - Dockerfile 中使用 uv 进行依赖安装

- **Helm Chart 改进**
  - 添加模板镜像的备用镜像仓库支持
  - 在 ConfigMap 和部署中添加 CONTROL_PLANE_URL
  - 切换到阿里云 PyPI 镜像加速依赖安装

- **会话管理**
  - 添加硬删除功能，支持级联删除
  - 增加 ID 字段长度
  - 空闲会话设置为永不过期（可配置）

- **模板管理**
  - 添加模板 ID 验证
  - 添加默认超时配置
  - 添加模板名称更新功能

- **MCP 服务器**
  - 添加 MCP 服务器实现，支持同步代码执行

### 🔧 改进

- 更新 MariaDB 数据库架构定义
- 更新 API 文档至 OpenAPI 3.1.0 规范
- 添加 uv.lock 确保依赖版本可重现

---

*发布于 2025-03-05*

## [0.2.0]

### 🚀 新功能

#### 存储与工作空间
- **S3 存储集成与 MinIO**
  - S3 兼容的对象存储后端
  - 直接文件上传/下载 API
  - 工作空间路径管理 (`s3://bucket/sessions/{id}/`)
  - 多格式文件支持

- **s3fs 工作空间挂载 (Kubernetes)**
  - 通过 s3fs FUSE 实现容器级 S3 存储桶挂载
  - 将会话目录绑定挂载到 `/workspace`
  - 无需额外的元数据库
  - 生产环境就绪，支持多节点 K8s 集群

- **Docker 卷挂载**
  - 本地开发环境卷挂载
  - 工作空间文件持久化
  - 无缝 S3 集成

#### 会话管理
- **会话列表 API**
  - `GET /api/v1/sessions` 支持过滤
  - 过滤条件：`status`、`template_id`、`created_after`、`created_before`
  - 使用 `limit` 和 `offset` 参数进行分页
  - 通过数据库索引优化查询性能

- **会话清理服务**
  - 自动清理空闲会话
  - 可配置的空闲阈值 (`IDLE_THRESHOLD_MINUTES`，默认：30 分钟)
  - 最大生命周期强制执行 (`MAX_LIFETIME_HOURS`，默认：6 小时)
  - 后台任务，可配置执行间隔 (`CLEANUP_INTERVAL_SECONDS`，默认：300 秒)
  - 设置为 `-1` 可禁用清理功能

- **状态同步服务**
  - 启动时与运行时容器同步
  - 孤立会话恢复
  - 基于容器健康状态的自动状态修正
  - 健康检查集成

#### Kubernetes 支持
- **Helm Chart 部署**
  - 完整的生产环境部署 Helm Chart
  - 可配置服务：控制平面、Web 控制台、MariaDB、MinIO
  - RBAC、ServiceAccount 和网络策略
  - 基于配置文件的环境差异化配置

- **Kubernetes 调度器**
  - 完整的 K8s 运行时支持
  - Pod 创建和生命周期管理
  - 通过 s3fs 挂载 S3 工作空间
  - 支持临时和持久化会话模式

- **原生 K8s 清单文件**
  - 独立的 K8s 部署 YAML 清单
  - Namespace、ConfigMap、Secret、ServiceAccount、Role 定义
  - MariaDB 和 MinIO 部署配置
  - s3fs 密码密钥管理

#### 运行时执行器
- **Python 依赖安装**
  - 自动从工作空间安装 `requirements.txt`
  - 依赖安装到本地文件系统（与 `/workspace` 隔离）
  - 执行前依赖准备
  - 支持自定义包索引（镜像源）

- **六边形架构**
  - 清晰的分层：领域层、应用层、基础设施层、接口层
  - 外部依赖使用端口-适配器模式
  - 通过依赖注入提升可测试性
  - 执行器端口：IExecutorPort、ICallbackPort、IIsolationPort、IArtifactScannerPort、IHeartbeatPort、ILifecyclePort

- **增强的执行模型**
  - 返回值存储和检索
  - 指标收集（CPU、内存、执行时间）
  - 错误信息捕获
  - 依赖安装状态跟踪

#### 开发工具
- **Docker Compose 部署**
  - 完整的开发环境
  - 一键部署：`docker-compose up -d`
  - 运行时节点注册
  - 健康检查集成

- **构建系统**
  - UV 包管理器集成
  - 可配置的基础镜像参数
  - 支持镜像源加速（国内下载优化）
  - 多阶段 Docker 构建

### 📚 文档

- **架构文档**
  - 完整的系统架构概览
  - 控制平面设计和组件
  - 存储架构（MinIO + s3fs）
  - Kubernetes 部署指南

- **API 文档**
  - RESTful API 端点参考
  - 请求/响应模式
  - 认证和安全

- **技术规范**
  - Python 依赖安装规范
  - S3 工作空间挂载架构
  - Kubernetes 运行时设计
  - 容器调度器架构

- **项目结构**
  - PROJECT_STRUCTURE.md 包含六边形架构详细信息
  - 使用 Mermaid 更新的架构图
  - 服务访问文档

### 🎯 核心能力

| 能力 | 描述 |
|------------|-------------|
| **多运行时** | Docker 和 Kubernetes 运行时支持 |
| **S3 存储** | MinIO 配合 s3fs 挂载实现工作空间持久化 |
| **会话生命周期** | 创建、执行、监控、清理 |
| **依赖管理** | 自动 Python 包安装 |
| **健康监控** | 容器健康检查和状态同步 |
| **生产就绪** | K8s Helm Chart、本地 Docker Compose |

### 📦 配置

#### 新增环境变量

```bash
# S3 存储
S3_BUCKET=sandbox-workspace
S3_REGION=us-east-1
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_ENDPOINT_URL=http://minio:9000

# 会话清理
IDLE_THRESHOLD_MINUTES=30      # 设置为 -1 禁用空闲清理
MAX_LIFETIME_HOURS=6           # 设置为 -1 禁用生命周期限制
CLEANUP_INTERVAL_SECONDS=300

# Kubernetes
KUBERNETES_NAMESPACE=sandbox-runtime
KUBECONFIG=/path/to/kubeconfig

# 执行器
CONTROL_PLANE_URL=http://control-plane:8000
EXECUTOR_PORT=8080
DISABLE_BWRAP=true
```

### 🔜 服务访问

| 服务 | URL | 凭据 |
|---------|-----|-------------|
| **API 文档** | http://localhost:8000/docs | - |
| **控制平面 API** | http://localhost:8000/api/v1 | - |
| **Web 控制台** | http://localhost:1101 | - |
| **MinIO 控制台** | http://localhost:9001 | minioadmin/minioadmin |

---

*发布于 2025-02-05*
