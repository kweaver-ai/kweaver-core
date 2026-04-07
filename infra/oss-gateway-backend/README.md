# OSS Gateway Backend

A unified object storage gateway service supporting multiple cloud providers (Alibaba Cloud OSS, Huawei Cloud OBS, Ceph S3).

## Features

- **Multi-vendor Support**: Alibaba Cloud OSS, Huawei Cloud OBS, Ceph S3
- **Unified API**: Single API for all object storage operations
- **Presigned URLs**: Generate presigned URLs for direct client access
- **Multipart Upload**: Support for large file uploads with multipart
- **Database Compatibility**: MariaDB/MySQL and DM8 support
- **Redis Cache**: Distributed cache for multi-instance deployment
- **High Performance**: Redis cache provides 10-100x performance improvement

## Architecture

```
┌─────────────────┐
│   Client/Frontend    │
└────────┬────────┘
         │ ① Request presigned URL
         ▼
┌─────────────────────────┐
│   Golang OSS Gateway    │
│  ├── Storage Management  │
│  ├── URL Generation      │
│  └── Vendor Adapters     │
└─────────┬───────────────┘
          │ ② Return presigned URL
          ▼
┌─────────────────┐
│   Client/Frontend    │ ③ Direct upload/download
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│          Object Storage Services         │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │Aliyun OSS│ │Huawei OBS│ │  Ceph   │   │
│  └─────────┘ └─────────┘ └─────────┘   │
└─────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21+
- MariaDB/MySQL or DM8 database
- Redis 6.0+ (for multi-instance deployment)
- Object storage account (OSS/OBS/Ceph)

### Installation

1. Clone the repository
```bash
cd oss-gateway-backend
```

2. Copy environment file
```bash
cp .env.example .env
```

3. Edit `.env` file with your configuration

4. Install dependencies
```bash
go mod download
```

5. Initialize database
```bash
go run main.go initdb
```

6. Start server
```bash
go run main.go server
```

## Configuration

配置方式优先级：**环境变量 > local.go > 默认值**

### 方式 1：修改本地配置文件（推荐用于开发）

编辑 `internal/config/local.go` 文件中的 `GetLocalConfig()` 函数：

```go
func GetLocalConfig() *LocalConfig {
    return &LocalConfig{
        RedisClusterMode: "standalone",
        RedisHost:        "localhost",
        RedisPort:        "6379",
        // ... 修改其他配置
    }
}
```

### 方式 2：使用环境变量

复制 `.env.example` 为 `.env.debug` 并修改：

```bash
cp .env.example .env.debug
```

### 方式 3：直接设置环境变量

```bash
export REDISCLUSTERMODE=standalone
export REDISHOST=localhost
export REDISPORT=6379
go run main.go server
```

### 📋 完整配置说明

详见 [环境变量配置文档](./ENVIRONMENT_VARIABLES.md)，包含：
- Redis 三种模式配置（单机/主从/哨兵）
- K8s ConfigMap/Secret 示例
- 与 Python 项目的环境变量兼容性说明

### 主要配置项

#### Server Configuration
- `PORT`: Server port (default: 8080)
- `NAME`: Service name (default: oss-gateway)

#### Database Configuration
- `RDSHOST`: 数据库地址
- `RDSPORT`: 数据库端口
- `RDSUSER`: 数据库用户名
- `RDSPASS`: 数据库密码
- `RDSDBNAME`: 数据库名称
- `DB_TYPE`: Database type (MYSQL, DM8, KDB9)

#### Redis Configuration（环境变量命名与 Python 项目统一）

**模式选择：**
- `REDISCLUSTERMODE`: Redis 模式 (standalone/master-slave/sentinel)

**单机模式 (standalone):**
- `REDISHOST`: Redis 地址
- `REDISPORT`: Redis 端口
- `REDISUSER`: Redis 用户名（可选）
- `REDISPASS`: Redis 密码（可选）
- `REDIS_DB`: Redis 数据库编号

**主从模式 (master-slave):**
- `REDISREADHOST`, `REDISREADPORT`, `REDISREADUSER`, `REDISREADPASS`: 读节点配置
- `REDISWRITEHOST`, `REDISWRITEPORT`, `REDISWRITEUSER`, `REDISWRITEPASS`: 写节点配置

**哨兵模式 (sentinel):**
- `REDIS_SENTINEL_ADDRS`: 哨兵地址列表（逗号分隔）
- `SENTINELMASTER`: 主节点名称
- `SENTINELUSER`: 哨兵用户名（可选）
- `SENTINELPASS`: 哨兵密码（可选）

> **注意**: 数据库和 Redis 环境变量命名已与 `mf-model-api` Python 项目保持一致，方便 K8s 统一管理。

#### OSS Configuration
- `OSS_DEFAULT_VALID_SECONDS`: Default URL expiration time (default: 3600)
- `OSS_MIN_PART_SIZE`: Minimum part size for multipart upload (default: 5MB)
- `OSS_MAX_PART_SIZE`: Maximum part size for multipart upload (default: 5GB)

## Multi-Instance Deployment

For production multi-instance deployment with Redis:

1. **Redis 支持三种模式**：
   - **单机模式 (standalone)**: 适用于开发和小规模生产
   - **主从模式 (master-slave)**: 支持读写分离
   - **哨兵模式 (sentinel)**: 高可用自动故障转移

2. **配置 Redis 连接**（环境变量命名与 Python 项目统一）:
   ```bash
   # 单机模式
   REDISCLUSTERMODE=standalone
   REDISHOST=your-redis-host
   REDISPORT=6379
   REDISPASS=your-redis-password
   
   # 主从模式
   REDISCLUSTERMODE=master-slave
   REDISREADHOST=redis-slave-host
   REDISWRITEHOST=redis-master-host
   
   # 哨兵模式
   REDISCLUSTERMODE=sentinel
   REDIS_SENTINEL_ADDRS=sentinel1:26379,sentinel2:26379
   SENTINELMASTER=mymaster
   ```

3. All instances share the same Redis and database
4. Cache hit rate >95% for high-concurrency scenarios

详细配置说明见 [ENVIRONMENT_VARIABLES.md](./ENVIRONMENT_VARIABLES.md)  
缓存架构说明见 [REDIS_CACHE_ARCHITECTURE.md](./REDIS_CACHE_ARCHITECTURE.md)

## API Documentation

### Health Check
- `GET /health/ready` - Readiness probe
- `GET /health/alive` - Liveness probe

### Storage Management
- `GET /api/v1/storages` - List all storages
- `GET /api/v1/storages/:id` - Get storage by ID
- `POST /api/v1/storages` - Create new storage
- `PUT /api/v1/storages/:id` - Update storage
- `DELETE /api/v1/storages/:id` - Delete storage
- `POST /api/v1/storages/:id/check` - Check storage connection

### Object Operations
- `GET /api/v1/head/:storageId/:key` - Get object metadata URL
- `POST /api/v1/head/:storageId` - Batch get object metadata URLs
- `GET /api/v1/upload/:storageId/:key` - Get upload URL
- `GET /api/v1/download/:storageId/:key` - Get download URL
- `GET /api/v1/delete/:storageId/:key` - Get delete URL

### Multipart Upload
- `GET /api/v1/initmultiupload/:storageId/:key` - Initialize multipart upload
- `POST /api/v1/uploadpart/:storageId/:key` - Get part upload URLs
- `POST /api/v1/completeupload/:storageId/:key` - Complete multipart upload

## Database Schema

### storage_config
Stores object storage configuration.

### multipart_upload_task
Tracks multipart upload tasks.

### operation_log
Logs all operations for audit.

## Development

### Project Structure

```
oss-gateway-backend/
├── cmd/
│   ├── server/         # Server entry point
│   └── initdb/         # Database initialization
├── internal/
│   ├── config/         # Configuration management
│   ├── database/       # Database connection
│   ├── handler/        # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   ├── model/          # Data models
│   ├── repository/     # Data access layer
│   ├── router/         # Route configuration
│   └── service/        # Business logic
├── pkg/
│   ├── adapter/        # Storage adapters
│   ├── crypto/         # Encryption utilities
│   ├── response/       # Response utilities
│   └── utils/          # Common utilities
└── migrations/         # Database migrations
    ├── mariadb/
    └── dm8/
```

## License

Copyright © 2024
