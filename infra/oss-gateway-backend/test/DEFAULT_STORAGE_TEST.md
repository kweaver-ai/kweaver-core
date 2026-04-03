# 默认存储互斥测试

## 功能说明

系统全局只允许存在一个默认存储配置。当创建或更新存储时，如果设置 `is_default: true`，但系统中已存在其他默认存储，操作会失败。

## 测试场景

### 场景1：创建第一个默认存储（成功）

**请求：**
```bash
POST /api/v1/storages
Content-Type: application/json

{
  "storage_name": "阿里云存储1",
  "vendor_type": "OSS",
  "endpoint": "https://oss-cn-hangzhou.aliyuncs.com",
  "bucket_name": "bucket1",
  "access_key_id": "LTAI5xxx",
  "access_key_secret": "secret1",
  "region": "cn-hangzhou",
  "is_default": true
}
```

**预期响应：**
```json
{
  "res": 0,
  "data": {
    "id": "A1B2C3D4E5F6G7H8",
    "status": "ok"
  }
}
```

---

### 场景2：创建第二个默认存储（失败）

**前提条件：** 系统中已存在默认存储"阿里云存储1"

**请求：**
```bash
POST /api/v1/storages
Content-Type: application/json

{
  "storage_name": "华为云存储",
  "vendor_type": "OBS",
  "endpoint": "https://obs.cn-north-4.myhuaweicloud.com",
  "bucket_name": "bucket2",
  "access_key_id": "ACCESS_KEY",
  "access_key_secret": "secret2",
  "region": "cn-north-4",
  "is_default": true
}
```

**预期响应：**
```json
{
  "code": "400031112",
  "message": "Default storage already exists",
  "description": "A default storage already exists: 阿里云存储1",
  "solution": "Please disable the current default storage before setting a new one"
}
```

**HTTP 状态码：** 400

---

### 场景3：更新已有存储为默认（失败）

**前提条件：** 
- 系统中已存在默认存储"阿里云存储1"（storage_id: A1B2C3D4）
- 存在另一个非默认存储"华为云存储"（storage_id: B2C3D4E5）

**请求：**
```bash
PUT /api/v1/storages/B2C3D4E5
Content-Type: application/json

{
  "is_default": true
}
```

**预期响应：**
```json
{
  "code": "400031112",
  "message": "Default storage already exists",
  "description": "A default storage already exists: 阿里云存储1",
  "solution": "Please disable the current default storage before setting a new one"
}
```

**HTTP 状态码：** 400

---

### 场景4：先禁用现有默认存储，再设置新的（成功）

**步骤1：** 禁用当前默认存储

```bash
PUT /api/v1/storages/A1B2C3D4
Content-Type: application/json

{
  "is_default": false
}
```

**预期响应：**
```json
{
  "res": 0,
  "data": {
    "status": "ok",
    "id": "A1B2C3D4"
  }
}
```

**步骤2：** 设置新的默认存储

```bash
PUT /api/v1/storages/B2C3D4E5
Content-Type: application/json

{
  "is_default": true
}
```

**预期响应：**
```json
{
  "res": 0,
  "data": {
    "status": "ok",
    "id": "B2C3D4E5"
  }
}
```

---

### 场景5：创建非默认存储（成功）

**前提条件：** 系统中已存在默认存储

**请求：**
```bash
POST /api/v1/storages
Content-Type: application/json

{
  "storage_name": "测试存储",
  "vendor_type": "OSS",
  "endpoint": "https://oss-cn-beijing.aliyuncs.com",
  "bucket_name": "bucket3",
  "access_key_id": "LTAI5yyy",
  "access_key_secret": "secret3",
  "region": "cn-beijing",
  "is_default": false
}
```

**预期响应：**
```json
{
  "res": 0,
  "data": {
    "id": "C3D4E5F6G7H8I9J0",
    "status": "ok"
  }
}
```

---

## 验证方法

### 1. 通过 List 接口验证

**查询所有存储：**
```bash
GET /api/v1/storages
```

**只查询默认存储：**
```bash
GET /api/v1/storages?is_default=true
```

**查询启用的默认存储：**
```bash
GET /api/v1/storages?enabled=true&is_default=true
```

**检查点：**
- 使用 `is_default=true` 参数时，返回的列表中最多只有一个存储
- 该存储的 `is_default: true`

### 2. 通过数据库直接查询

```sql
SELECT f_storage_id, f_storage_name, f_is_default 
FROM t_storage_config 
WHERE f_is_default = 1;
```

**检查点：**
- 查询结果应该最多返回 1 条记录

---

## 注意事项

1. **安全设计**：修改默认存储是危险操作，因此采用显式两步操作：
   - 步骤1：先将现有默认存储设为非默认
   - 步骤2：再设置新的默认存储

2. **错误码**：400031112 - Default storage already exists

3. **兼容性**：如果系统中没有默认存储，可以直接创建或更新为默认存储
