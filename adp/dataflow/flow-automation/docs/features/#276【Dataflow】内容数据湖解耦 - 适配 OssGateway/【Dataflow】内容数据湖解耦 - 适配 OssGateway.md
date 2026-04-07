# #276 【Dataflow】内容数据湖解耦 - 适配 OssGateway

---

## 一、需求分析

### 1.1 需求背景

#### 需求信息

| 字段     | 内容                                         |
| -------- | -------------------------------------------- |
| 需求号   | 276                                          |
| 类型     | Feature                                      |
| 标题     | 【Dataflow】内容数据湖解耦 - 适配 OssGateway |
| 需求来源 | 产品规划                                     |

#### 需求场景

Dataflow 的非结构化处理能力源于内容数据湖。随着 Dataflow 演进为 ADP 的数据处理基础设施，Dataflow 亟需与内容数据湖解耦，以满足多种用户场景的非结构化数据处理要求。

OssGateway 作为内容数据湖非结构化数据存储的基础设施，在 Dataflow 中具有重要作用，主要应用于以下场景：

**流程数据归档**

Dataflow 将运行成功的流程数据通过 OssGateway 归档到对象存储，降低数据库压力。

**任务缓存**

为避免重复处理相同的文档，文档解析等节点支持将处理结果上传到 OssGateway 并记录任务哈希，实现处理结果的缓存和复用。

**内容处理节点中间数据流转**

ContentPipeline 相关节点依赖 OssGateway 实现中间数据流转。

**非结构化数据触发节点上传 S3 文档**

上传本地文档到 S3 并运行数据处理流程。

#### 用户期望

本次需求目标包括：

1. **支持多种对象存储后端**
   - 内容数据湖 OssGateway 服务
   - S3 兼容存储（过度方案）
   - ADP OssGateway 兼容服务

2. **保持向后兼容**
   - 兼容现有 Edition 配置
   - 平滑迁移到新的配置方式

3. **灵活的配置管理**
   - 通过 `storageBackend` 配置选择存储后端
   - 支持运行时动态选择

4. **完整的存储操作支持**
   - 文件上传
   - 文件下载
   - 文件删除
   - 获取文件元数据
   - 生成预签名下载链接

---

### 1.2 用户故事

| 角色         | 痛点（Why）                            | 活动（What）              | 价值（Value）              |
| ------------ | -------------------------------------- | ------------------------- | -------------------------- |
| ADP 部署者   | 需要 ADP OssGateway 而非内容数据湖服务 | 配置 storageBackend 参数  | 灵活选择存储后端           |
| 开发者       | 无需 OssGateway Manager 服务           | 使用 S3 存储后端          | 简化开发环境部署           |
| 系统运维人员 | 需要管理多个存储后端                   | 配置不同的存储后端方案    | 灵活管理存储资源           |
| 应用开发者   | 希望代码在不同存储后端间保持一致       | 使用统一的 OSS 接口       | 降低开发和维护成本         |
| 数据流设计者 | 需要在数据流中使用对象存储             | 使用统一接口上传/下载文件 | 支持数据处理流程的文件操作 |

---

### 1.3 现状分析

#### 过度方案

在过度版本中，由于 OssGateway 的缺失，Dataflow 实现了基于 S3 的 OssGateway 兼容方案 [808344-OSS 适配 s3 存储](../808344-OSS%20适配%20s3%20存储/808344-OSS%20适配%20s3%20存储.md)。支持基于应用版本（社区版、商业版）采用不同的对象存储后端方案。

**Edition 配置方式**

| Edition 值 | 存储后端           | 说明                          |
| ---------- | ------------------ | ----------------------------- |
| community  | S3 存储            | 社区版，直接连接 S3 兼容存储  |
| commercial | OssGateway Manager | 商业版，通过 OSS Gateway 服务 |

#### 过度方案的局限性

1. 仅支持两种固定后端，无法扩展
2. 与版本绑定，缺乏灵活性
3. 无法支持 ADP OssGateway 兼容服务

---

### 1.4 目标

ADP 实现了 OssGateway 兼容服务（以下称为 ADP OssGateway），废弃基于 S3 的过度方案。

1. **对象存储后端支持可配置**
   - ossgateway: 内容数据湖 OssGateway 服务
   - s3: 过度方案实现的 S3 存储
   - ossgateway_adp: ADP OssGateway 兼容服务

### 1.5 限制

对象存储后端属于基础设施，由部署参数决定，不支持兼容执行中、已完成的流程数据。

对于有兼容需求的客户，可提供迁移方案（不在本需求范围内实现）。

---

## 二、业务功能设计

### 2.1 概念与术语

| 中文           | 英文                  | 定义                                                    |
| -------------- | --------------------- | ------------------------------------------------------- |
| 对象存储后端   | Storage Backend       | 对象存储的实现方式，支持 OssGateway、S3、ADP OssGateway |
| OssGateway     | OSS Gateway           | 内容数据湖的对象存储网关服务                            |
| ADP OssGateway | ADP OSS Gateway       | ADP 实现的 OssGateway 兼容服务                          |
| S3 后端        | S3 Backend            | 基于 AWS S3 SDK 实现的对象存储后端                      |
| Edition        | Edition               | 系统版本标识，区分社区版和商业版                        |
| 存储后端配置   | StorageBackend Config | 指定使用的对象存储后端类型                              |

---

### 2.2 业务用例

#### 用例名称

**配置对象存储后端**

#### 用例说明

| 项目     | 描述                                 |
| -------- | ------------------------------------ |
| 参与者   | 系统运维人员、ADP 部署者             |
| 前置条件 | 已部署系统；对象存储服务可用         |
| 后置条件 | 存储后端配置完成并可正常执行文件操作 |

---

### 2.3 业务功能定义

#### 存储后端配置

##### StorageBackend 字段

在 `Server` 配置中新增 `StorageBackend` 字段，用于指定对象存储后端：

| 字段名         | 类型   | 可选值                                 | 描述         |
| -------------- | ------ | -------------------------------------- | ------------ |
| StorageBackend | string | `ossgateway` / `s3` / `ossgateway_adp` | 存储后端类型 |

##### 后端选择逻辑

根据 `StorageBackend` 字段选择对应的 OSS 网关实现：

| 值             | 实现类        | 说明                               |
| -------------- | ------------- | ---------------------------------- |
| ossgateway     | ossGatetway   | 内容数据湖 OssGateway Manager 服务 |
| s3             | ossGatetWayS3 | 基于 S3 SDK 的实现                 |
| ossgateway_adp | ossGateWayAdp | ADP OssGateway 兼容服务            |

##### 兼容性处理

为保持向后兼容，当 `StorageBackend` 未配置时，使用 `Edition` 字段推断：

| Edition    | 推断的 StorageBackend |
| ---------- | --------------------- |
| community  | s3                    |
| commercial | ossgateway            |

---

#### 存储后端实现

##### OssGateway（商业版）

通过 OSS Gateway Manager 服务提供对象存储功能：

- 依赖 OSS Gateway Manager 服务
- 支持多租户
- 支持权限控制

##### S3 后端（社区版）

直接连接 S3 兼容存储：

- 直接连接 S3 兼容存储（AWS S3、MinIO、Ceph RGW 等）
- 无需依赖 OSS Gateway Manager 服务
- 支持标准 S3 协议
- 配置详见 [808344-OSS 适配 s3 存储](../808344-OSS%20适配%20s3%20存储/808344-OSS%20适配%20s3%20存储.md)

##### ADP OssGateway（新增）

ADP 实现的 OssGateway 兼容服务：

- 独立于内容数据湖
- 提供 OssGateway 兼容接口
- 支持标准对象存储操作

---

### 2.4 业务流程

#### 2.4.1 OSS 网关初始化流程

```
开始
  ↓
调用 NewOssGateWay()
  ↓
读取配置中的 StorageBackend 字段
  ↓
StorageBackend 已配置？
  ├─ 是 → 根据 StorageBackend 选择实现
  │        ├─ "ossgateway" → 创建 ossGatetway 实例
  │        ├─ "s3"         → 创建 ossGatetWayS3 实例
  │        └─ "ossgateway_adp" → 创建 ossGateWayAdp 实例
  │
  └─ 否 → 读取 Edition 字段
           ↓
         Edition == "community"?
           ├─ 是 → 创建 ossGatetWayS3 实例
           └─ 否 → 创建 ossGatetway 实例
  ↓
返回对应的 OSS 网关实例
  ↓
结束
```

#### 2.4.2 文件上传流程

```
开始
  ↓
接收上传请求
  ↓
调用 OssGateWay.UploadFile()
  ↓
根据后端类型执行上传
  ├─ ossgateway     → 调用 OSS Gateway Manager API
  ├─ s3             → 调用 S3 PutObject API
  └─ ossgateway_adp → 调用 ADP OssGateway API
  ↓
返回上传结果
  ↓
结束
```

#### 2.4.3 文件下载流程

```
开始
  ↓
接收下载请求
  ↓
调用 OssGateWay.DownloadFile()
  ↓
根据后端类型执行下载
  ├─ ossgateway     → 调用 OSS Gateway Manager API
  ├─ s3             → 调用 S3 GetObject API
  └─ ossgateway_adp → 调用 ADP OssGateway API
  ↓
返回文件内容
  ↓
结束
```

---

## 三、技术设计

### 3.1 架构设计

#### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                     业务层                                   │
│  (数据流、工作流、算子、任务缓存等)                          │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      │ 调用统一接口
                      ↓
┌─────────────────────────────────────────────────────────────┐
│                  OSS Gateway 接口层                         │
│  OssGateWay Interface                                       │
│  - UploadFile                                               │
│  - DownloadFile                                             │
│  - DeleteFile                                               │
│  - GetDownloadURL                                           │
│  - GetObjectMeta                                            │
│  - GetAvaildOSS                                             │
└─────────────────────┬───────────────────────────────────────┘
                      │
          ┌───────────┼───────────┐
          │           │           │
          ↓           ↓           ↓
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  OssGateway     │ │   S3 后端       │ │  ADP OssGateway  │
│  (商业版)       │ │  (社区版)       │ │  (新增)          │
│  ossGatetway    │ │  ossGatetWayS3  │ │  ossGateWayAdp   │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         │ HTTP API          │ AWS SDK           │ HTTP API
         ↓                   ↓                   ↓
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  OssGateway     │ │   S3 兼容存储   │ │  ADP OssGateway  │
│  Manager 服务   │ │  (AWS/MinIO)    │ │  服务            │
└─────────────────┘ └─────────────────┘ └─────────────────┘
```

#### 模块划分

| 模块名称       | 文件路径                           | 职责                          |
| -------------- | ---------------------------------- | ----------------------------- |
| OSS 网关接口   | `drivenadapters/ossgateway.go`     | OssGateWay 接口定义           |
| 商业版网关实现 | `drivenadapters/ossgateway.go`     | 商业版 OSS Gateway 实现       |
| S3 网关实现    | `drivenadapters/ossgateway_s3.go`  | 社区版 S3 网关实现            |
| ADP 网关实现   | `drivenadapters/ossgateway_adp.go` | ADP OssGateway 兼容实现       |
| S3 连接管理    | `libs/go/s3/conn.go`               | S3 连接配置和客户端初始化     |
| S3 配置管理    | `libs/go/s3/s3.go`                 | S3 配置加载和连接管理         |
| 配置定义       | `common/config.go`                 | StorageBackend 字段和配置定义 |

---

### 3.2 数据结构设计

#### StorageBackend 类型

```go
type StorageBackend string

const (
    StorageBackendOssGateway    StorageBackend = "ossgateway"
    StorageBackendS3            StorageBackend = "s3"
    StorageBackendOssGatewayAdp StorageBackend = "ossgateway_adp"
)
```

#### 配置结构

```go
type Server struct {
    // ... 其他配置
    Edition        string         `yaml:"edition"`         // 版本标识（向后兼容）
    StorageBackend StorageBackend `yaml:"storageBackend"` // 存储后端配置
    // ... 其他配置
}
```

#### ADP OssGateway 配置

```go
type AdpOssConfig struct {
    Endpoint   string `yaml:"endpoint"`    // ADP OssGateway 服务地址
    BucketName string `yaml:"bucketName"`  // 存储桶名称
    // ... 其他配置
}
```

---

### 3.3 接口设计

#### OssGateWay 接口（保持不变）

```go
type OssGateWay interface {
    UploadFile(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader, size int64) error
    SimpleUpload(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader) error
    DownloadFile(ctx context.Context, ossID, key string, internalRequest bool, opts ...OssOpt) ([]byte, error)
    DownloadFile2Local(ctx context.Context, ossID, key string, internalRequest bool, filePath string, opts ...OssOpt) (int64, error)
    DeleteFile(ctx context.Context, ossID, key string, internalRequest bool) error
    GetDownloadURL(ctx context.Context, ossID, key string, expires int64, internalRequest bool, opts ...OssOpt) (string, error)
    GetObjectMeta(ctx context.Context, ossID, key string, internalRequest bool, opts ...OssOpt) (int64, error)
    GetAvaildOSS(ctx context.Context) (string, error)
    NewReader(ossID string, ossKey string, opts ...OssOpt) *Reader
}
```

#### ossGateWayAdp 实现

```go
type ossGateWayAdp struct {
    client    *AdpOssClient
    config    *AdpOssConfig
}

func NewOssGateWayAdp(config *AdpOssConfig) *ossGateWayAdp
func (g *ossGateWayAdp) UploadFile(ctx context.Context, ossID, key string, internalRequest bool, file io.Reader, size int64) error
func (g *ossGateWayAdp) DownloadFile(ctx context.Context, ossID, key string, internalRequest bool, opts ...OssOpt) ([]byte, error)
func (g *ossGateWayAdp) DeleteFile(ctx context.Context, ossID, key string, internalRequest bool) error
func (g *ossGateWayAdp) GetDownloadURL(ctx context.Context, ossID, key string, expires int64, internalRequest bool, opts ...OssOpt) (string, error)
func (g *ossGateWayAdp) GetObjectMeta(ctx context.Context, ossID, key string, internalRequest bool, opts ...OssOpt) (int64, error)
func (g *ossGateWayAdp) GetAvaildOSS(ctx context.Context) (string, error)
```

#### 网关初始化函数

```go
func NewOssGateWay(cfg *config.Server) OssGateWay {
    // 优先使用 StorageBackend 配置
    if cfg.StorageBackend != "" {
        switch cfg.StorageBackend {
        case StorageBackendOssGateway:
            return newOssGateWay(cfg)
        case StorageBackendS3:
            return newOssGateWayS3(cfg)
        case StorageBackendOssGatewayAdp:
            return newOssGateWayAdp(cfg)
        default:
            log.Warnf("unknown storage backend: %s, using edition fallback", cfg.StorageBackend)
        }
    }

    // 向后兼容：使用 Edition 字段推断
    if cfg.Edition == "community" {
        return newOssGateWayS3(cfg)
    }
    return newOssGateWay(cfg)
}
```

---

## 四、部署说明

### 4.1 前置条件

- 已部署对象存储服务（OssGateway Manager / S3 兼容存储 / ADP OssGateway）
- 已配置相关凭证和连接信息

### 4.2 配置步骤

#### 使用 OssGateway（商业版）

```yaml
flowAutomation:
  service:
    storageBackend: "ossgateway"
    # OssGateway 相关配置
    ossgateway:
      endpoint: "oss-gateway-manager.example.com"
```

#### 使用 S3 存储（社区版）

```yaml
flowAutomation:
  service:
    storageBackend: "s3"
  s3:
    default: "default"
    connections:
      - name: default
        endpoint: s3.amazonaws.com
        region: us-east-1
        accessKeyID: "your-access-key-id"
        secretAccessKey: "your-secret-access-key"
        bucketName: "dataflow"
```

#### 使用 ADP OssGateway

```yaml
flowAutomation:
  service:
    storageBackend: "ossgateway_adp"
  adpOss:
    endpoint: "adp-oss.example.com"
    bucketName: "dataflow"
```

#### 向后兼容配置

```yaml
# 不指定 storageBackend，使用 edition 推断
flowAutomation:
  service:
    edition: "community"  # 将使用 s3 后端
```

### 4.3 配置参数说明

| 参数名         | 类型   | 必填 | 描述                                       |
| -------------- | ------ | ---- | ------------------------------------------ |
| storageBackend | string | 否   | 存储后端类型，不配置时使用 edition 推断    |
| edition        | string | 否   | 版本标识（community/commercial），向后兼容 |
| s3.*           | object | 否   | S3 配置，详见 [808344-OSS 适配 s3 存储]    |
| adpOss.*       | object | 否   | ADP OssGateway 配置                        |
| ossgateway.*   | object | 否   | OssGateway Manager 配置                    |

---

## 五、附录

### 5.1 相关文件清单

| 文件路径                                                   | 说明                             |
| ---------------------------------------------------------- | -------------------------------- |
| `drivenadapters/ossgateway.go`                             | OSS Gateway 接口定义和商业版实现 |
| `drivenadapters/ossgateway_s3.go`                          | 社区版 S3 网关实现               |
| `drivenadapters/ossgateway_adp.go`                         | ADP OssGateway 兼容实现          |
| `libs/go/s3/conn.go`                                       | S3 连接配置和客户端实现          |
| `libs/go/s3/s3.go`                                         | S3 配置管理和连接池              |
| `common/config.go`                                         | StorageBackend 字段和配置定义    |
| `charts/dataflow/values.yaml`                              | Helm Chart 配置文件              |
| `charts/dataflow/templates/flow-automation/configmap.yaml` | ConfigMap 模板                   |

### 5.2 相关需求

| 需求号 | 标题                         | 关系     |
| ------ | ---------------------------- | -------- |
| 808344 | 【Autoflow】OSS 适配 s3 存储 | 过度方案 |

### 5.3 参考资料

- [808344-OSS 适配 s3 存储](../808344-OSS%20适配%20s3%20存储/808344-OSS%20适配%20s3%20存储.md)
- [AWS SDK for Go v2 文档](https://aws.github.io/aws-sdk-go-v2/docs/)
- [MinIO 文档](https://docs.min.io/)