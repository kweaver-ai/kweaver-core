# 1. 架构设计


> **文档导航**: [返回架构总览](./overview.md)


## 1. 架构设计
### 1.1 整体架构

系统采用管理中心（Control Plane）与容器调度器（Container Scheduler）分离的云原生架构，支持 Docker 和 Kubernetes 两种部署模式。
核心设计原则：

- 控制平面无状态，支持水平扩展
- 容器调度器池化管理，动态伸缩
- 协议驱动的解耦设计
- 多层安全隔离
- 异步高并发处理


### 1.2 C4 架构模型
#### C4 Level 1: 系统上下文
```mermaid
graph TB
    subgraph External["外部系统"]
        上层服务["Data Agent/Operator Platform"]
        Dev["开发者<br/>(通过 SDK/API)"]
        K8s["Kubernetes 集群"]
        Docker["Docker Engine"]
    end

    subgraph SandboxPlatform["Python 沙箱平台"]
        ControlPlane["管理中心<br/>(Control Plane)"]
        ContainerScheduler["Container Scheduler 模块"]
    end

    上层服务 -->|执行代码请求| ControlPlane
    Dev -->|SDK/API 调用| ControlPlane
    ContainerScheduler -->|直接调用| K8s
    ContainerScheduler -->|直接访问| Docker
    ContainerScheduler -->|上报结果| ControlPlane

    style SandboxPlatform fill:#e1f5ff
    style External fill:#fff4e6
```
外部交互：

- DataAgent/Operator Platform 系统通过 RESTful API 发起代码执行请求
- 开发者通过 Python SDK 集成沙箱能力
- 依赖 Kubernetes/Docker 提供容器基础设施


#### C4 Level 2: 容器视图
```mermaid
graph TB
    subgraph ControlPlane["管理中心 (Control Plane)"]
        API["API Gateway<br/>(FastAPI)"]
        Scheduler["调度器<br/>(Scheduler)"]
        SessionMgr["会话管理器<br/>(Session Manager)"]
        TemplateMgr["模板管理器<br/>(Template Manager)"]
        Monitor["监控探针<br/>(Health Probe)"]
        ResultStore["结果存储<br/>(Result Store)"]
        SessionCleanup["会话清理服务<br/>(Session Cleanup)"]
    end

    subgraph ContainerScheduler["Container Scheduler 模块"]
        DockerRuntime["Docker Scheduler"]
        K8sRuntime["K8s Scheduler"]
    end

    subgraph Sandbox["沙箱实例"]
        Container["容器<br/>(Docker/Pod)"]
        BubbleWrap["Bubblewrap<br/>(进程隔离)"]
        Executor["执行器<br/>(Code Executor)"]
    end

    subgraph Storage["存储层"]
        MariaDB["MariaDB<br/>(会话状态/模板)"]
        S3["MinIO<br/>(S3-compatible 对象存储)"]
        Etcd["Etcd<br/>(配置中心)"]
    end

    API --> Scheduler
    API --> SessionMgr
    API --> TemplateMgr
    Scheduler --> Monitor
    Scheduler --> DockerRuntime
    Scheduler --> K8sRuntime
    SessionMgr --> MariaDB
    TemplateMgr --> MariaDB
    Monitor --> DockerRuntime
    Monitor --> K8sRuntime
    ResultStore --> S3
    SessionCleanup --> SessionMgr

    DockerRuntime --> Container
    K8sRuntime --> Container
    Container --> BubbleWrap
    BubbleWrap --> Executor
    Executor -->|上报结果| API

    TemplateMgr --> Etcd

    style ControlPlane fill:#bbdefb
    style ContainerScheduler fill:#c8e6c9
    style Sandbox fill:#fff9c4
    style Storage fill:#f8bbd0
```
关键容器：

- API Gateway: 统一入口，基于 FastAPI 实现
- 调度器: 智能任务分发和资源调度
- 会话管理器: 会话生命周期管理
- Container Scheduler: Docker/K8s 运行时实例管理
- 存储层：
  - MariaDB（会话状态/模板/执行记录）
  - MinIO（S3-compatible 对象存储，workspace 文件）
  - Etcd（配置中心）

**存储架构说明**：
- Control Plane 通过 S3 API 将文件写入 MinIO 的 /sessions/{session_id}/ 路径
- Executor Pod 使用 s3fs 在容器内挂载 S3 bucket 的 session 子目录到 /workspace
- 执行时生成的文件通过 S3 API 直接写入 MinIO
- MariaDB 存储 stdout、stderr、执行状态和文件列表（artifacts）
- 下载文件时通过文件 API 直接从 MinIO 获取

> 详细存储架构请参考 [10. MinIO-Only 存储架构](10-minio-only-architecture.md)

#### 部署架构
```mermaid
graph TB
    subgraph Internet["🌐 互联网"]
        User["👤 开发者/Agent系统"]
    end
    
    subgraph K8sCluster["☸️ Kubernetes 集群"]
        
        subgraph IngressLayer["入口层"]
            Ingress["Ingress Controller<br/>Nginx/Traefik"]
            LB["Load Balancer<br/>L4/L7"]
        end
        
        subgraph ControlPlaneNS["📦 Namespace: sandbox-system"]
            subgraph ControlPlaneDeployment["Deployment: control-plane"]
                CP1["Pod: control-plane-1<br/>├─ API Gateway<br/>├─ Scheduler<br/>├─ Session Manager<br/>└─ Health Probe"]
                CP2["Pod: control-plane-2<br/>├─ API Gateway<br/>├─ Scheduler<br/>├─ Session Manager<br/>└─ Health Probe"]
                CP3["Pod: control-plane-3<br/>├─ API Gateway<br/>├─ Scheduler<br/>├─ Session Manager<br/>└─ Health Probe"]
            end
            
            CPService["Service: control-plane-svc<br/>Type: ClusterIP<br/>Port: 8000"]
            
            HPA["HPA<br/>Min: 3, Max: 10<br/>CPU Target: 70%"]
        end
        
        subgraph RuntimeNS["🔒 Namespace: sandbox-runtime"]

            subgraph ActiveSandboxGroup["活跃沙箱"]
                SB1["Pod: sandbox-abc123<br/>├─ Session: abc123<br/>├─ Status: Executing<br/>└─ CPU: 0.8, Mem: 400Mi"]
                SB2["Pod: sandbox-def456<br/>├─ Session: def456<br/>├─ Status: Idle<br/>└─ CPU: 0.1, Mem: 200Mi"]
                SB3["Pod: sandbox-xyz789<br/>├─ Session: xyz789<br/>├─ Status: Executing<br/>└─ CPU: 1.0, Mem: 512Mi"]
            end

            NetworkPolicy["NetworkPolicy<br/>- 禁止 Pod 间通信<br/>- 仅允许访问管理中心<br/>- 可选白名单外部访问"]
        end
        
        subgraph DataLayer["💾 数据层 - Namespace: data"]

            subgraph MariaDBCluster["StatefulSet: MariaDB Cluster"]
                DB1["mariadb-0<br/>Role: Primary<br/>Sandbox DB"]
                DB2["mariadb-1<br/>Role: Replica"]
                DB3["mariadb-2<br/>Role: Replica"]
            end

            subgraph EtcdCluster["StatefulSet: Etcd Cluster"]
                Etcd1["etcd-0"]
                Etcd2["etcd-1"]
                Etcd3["etcd-2"]
            end

            subgraph MinIOCluster["MinIO Cluster"]
                MinIO1["minio-0<br/>S3 Workspace Storage"]
                MinIO2["minio-1"]
                MinIO3["minio-2"]
                MinIO4["minio-3"]
            end

            MariaDBService["Service: mariadb-svc<br/>Port: 3306"]
            EtcdService["Service: etcd-svc"]
            MinIOService["Service: minio-svc<br/>Port: 9000/9001"]
        end
        
    end
    
    subgraph ExternalServices["☁️ 外部服务"]
        Registry["Container Registry<br/>- Docker Hub<br/>- Harbor<br/>- 私有镜像仓库"]
    end
    
    User -->|"HTTPS<br/>TLS 1.3"| Ingress
    Ingress --> LB
    LB --> CPService
    CPService --> CP1
    CPService --> CP2
    CPService --> CP3
    
    HPA -.->|监控副本数| ControlPlaneDeployment
    
    CP1 -->|"查询/写入<br/>会话状态 / 元数据"| MariaDBService
    CP2 -->|"查询/写入<br/>会话状态 / 元数据"| MariaDBService
    CP3 -->|"查询/写入<br/>会话状态 / 元数据"| MariaDBService
    
    MariaDBService --> DB1
    DB1 -.->|主从复制| DB2
    DB1 -.->|主从复制| DB3
    
    CP1 -->|"读取配置<br/>模板信息"| EtcdService
    CP2 -->|"读取配置<br/>模板信息"| EtcdService
    EtcdService --> Etcd1
    EtcdService --> Etcd2
    EtcdService --> Etcd3

    MinIOService --> MinIO1
    MinIOService -.->|"数据复制"| MinIO2
    MinIOService -.->|"数据复制"| MinIO3
    MinIOService -.->|"数据复制"| MinIO4

    CP1 -.->|"K8s API<br/>创建 Pod"| SB1
    CP2 -.->|"K8s API<br/>创建 Pod"| SB2
    CP3 -.->|"K8s API<br/>创建 Pod"| SB3
    
    NetworkPolicy -.->|限制| SB1
    NetworkPolicy -.->|限制| SB2
    NetworkPolicy -.->|限制| SB3
    
    SB1 -->|"上报结果<br/>S3 API"| MinIOService
    SB2 -->|"上报结果<br/>S3 API"| MinIOService
    SB3 -->|"上报结果<br/>S3 API"| MinIOService

    CP1 -.->|"写入会话状态<br/>S3 API"| MariaDBService
    CP2 -.->|"写入会话状态<br/>S3 API"| MariaDBService
    CP3 -.->|"写入会话状态<br/>S3 API"| MariaDBService

    CP1 -.->|"拉取镜像"| Registry
    Warm1 -.->|"基础镜像"| Registry
    SB1 -.->|"用户镜像"| Registry
    
    
    classDef ingressStyle fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef controlStyle fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef runtimeStyle fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px
    classDef dataStyle fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef externalStyle fill:#f1f8e9,stroke:#33691e,stroke-width:2px

    class Ingress,LB ingressStyle
    class CP1,CP2,CP3,CPService,HPA controlStyle
    class SB1,SB2,SB3,NetworkPolicy runtimeStyle
    class DB1,DB2,DB3,Etcd1,Etcd2,Etcd3,MinIO1,MinIO2,MinIO3,MinIO4,MariaDBService,EtcdService,MinIOService dataStyle
    class Registry externalStyle


```
