# OSS 网关 API 文档

## 基础信息

**基础 URL:** `http://localhost:8080/api/v1`

**协议:** HTTP/HTTPS

**请求格式:** JSON

**响应格式:** JSON

---

## 响应规范

所有 API 响应遵循统一格式：

### 成功响应（普通接口）
```json
{
  "data": { ... }
}
```

### 成功响应（列表接口）
```json
{
  "count": 10,
  "data": [ ... ]
}
```

### 错误响应
```json
{
  "code": "400031101",
  "message": "Invalid parameter",
  "description": "The parameter 'storage_id' is invalid",
  "solution": "Please provide a valid storage_id",
  "cause": "storage_id is required"
}
```

**说明：**
- 成功响应使用 `res: 0` 标识成功，直接在 `data` 字段中返回数据
- 列表接口使用 `count` 字段表示总数，`data` 字段返回列表数据
- 错误响应包含详细的错误信息，包括错误码、消息、描述、解决方案和原因

### 错误码表

| 错误码        | HTTP 状态码 | 描述              |
|--------------|------------|-------------------|
| 400031100    | 400        | 错误的请求          |
| 400031101    | 400        | 无效的参数          |
| 400031102    | 400        | 无效的文件大小       |
| 400031107    | 400        | 存储名称已存在       |
| 400031108    | 400        | 存储配置已存在       |
| 400031109    | 400        | 键数量过多          |
| 400031110    | 400        | 无效的供应商类型     |
| 400031112    | 400        | 默认存储已存在       |
| 404031100    | 404        | 未找到             |
| 404031101    | 404        | 存储配置未找到       |
| 500031100    | 500        | 内部服务器错误       |
| 500031101    | 500        | 连接失败            |
| 503031100    | 503        | 服务未就绪          |

---

## 存储管理 API

### 1. 获取存储列表

获取所有存储配置。

**接口地址:** `GET /storages`

**查询参数:**

| 参数名       | 类型    | 必填 | 说明                                       |
|-------------|---------|------|-------------------------------------------|
| vendor_type | string  | 否   | 厂商类型过滤 (OSS/OBS/ECEPH)               |
| enabled     | boolean | 否   | 启用状态过滤 (true/false)                  |
| is_default  | boolean | 否   | 默认存储过滤 (true/false)                  |
| page        | int     | 否   | 页码，从1开始，默认1                        |
| size        | int     | 否   | 每页大小，默认10，最大1000                  |
| order       | string  | 否   | 排序方向 (asc/desc)，默认 desc             |
| rule        | string  | 否   | 排序字段 (create_time/update_time/storage_name)，默认 update_time |
| name        | string  | 否   | 存储名称模糊搜索                            |

**示例请求：**

1. 获取所有存储：`GET /storages`
2. 获取启用的存储：`GET /storages?enabled=true`
3. 获取默认存储：`GET /storages?is_default=true`
4. 获取启用的默认存储：`GET /storages?enabled=true&is_default=true`
5. 按名称搜索：`GET /storages?name=阿里云`

**响应示例:**
```json
{
  "count": 1,
  "data": [
    {
      "storage_id": "A1B2C3D4E5F6G7H8",
      "storage_name": "阿里云存储1",
      "vendor_type": "OSS",
      "endpoint": "https://oss-cn-hangzhou.aliyuncs.com",
      "bucket_name": "my-bucket",
      "region": "cn-hangzhou",
      "is_default": true,
      "is_enabled": true,
      "internal_endpoint": "",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### 2. 获取单个存储配置

根据 ID 获取单个存储配置。

**接口地址:** `GET /storages/:id`

**路径参数:**

| 参数名 | 类型   | 必填 | 说明     |
|--------|--------|------|----------|
| id     | string | 是   | 存储 ID  |

**响应示例:**
```json
{
  "data": {
    "storage_id": "A1B2C3D4E5F6G7H8",
    "storage_name": "阿里云存储1",
    "vendor_type": "OSS",
    "endpoint": "https://oss-cn-hangzhou.aliyuncs.com",
    "bucket_name": "my-bucket",
    "region": "cn-hangzhou",
    "is_default": true,
    "is_enabled": true,
    "internal_endpoint": "",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 3. 创建存储配置

创建新的存储配置。

**接口地址:** `POST /storages`

**请求体:**
```json
{
  "storage_name": "阿里云存储1",
  "vendor_type": "OSS",
  "endpoint": "https://oss-cn-hangzhou.aliyuncs.com",
  "bucket_name": "my-bucket",
  "access_key_id": "LTAI5xxx",
  "access_key_secret": "your-secret",
  "region": "cn-hangzhou",
  "is_default": false,
  "internal_endpoint": ""
}
```

**字段说明:**

| 字段名            | 类型    | 必填 | 说明                                        |
|------------------|---------|------|-------------------------------------------|
| storage_name     | string  | 是   | 存储配置的显示名称                                 |
| vendor_type      | string  | 是   | 厂商类型 (OSS/OBS/ECEPH)                      |
| endpoint         | string  | 是   | 服务端点 URL (必须以 http:// 或 https:// 开头)      |
| bucket_name      | string  | 是   | 存储桶名称                                     |
| access_key_id    | string  | 是   | 访问密钥 ID,对于私有化部署的ECEPH，该字段对应用户的账户名         |
| access_key_secret| string  | 是   | 访问密钥                                      |
| region           | string  | 条件 | 区域标识符 (OSS/OBS必填，ECEPH可选)                 |
| is_default       | boolean | 否   | 是否设为默认存储（全局只能有一个默认存储，如果系统已存在其他默认存储，创建会失败） |
| internal_endpoint| string  | 否   | 内网访问端点                                    |

**默认存储规则：**
- 系统全局只允许存在一个默认存储
- 如果创建时设置 `is_default: true`，但系统已存在其他默认存储，则创建失败
- 错误响应示例：
```json
{
  "code": "400031112",
  "message": "Default storage already exists",
  "description": "A default storage already exists: 阿里云存储1",
  "solution": "Please disable the current default storage before setting a new one"
}
```

**响应示例:**
```json
{
  "data": {
    "id": "2035957924035170304",
    "status": "ok"
  }
}
```

---

### 4. 更新存储配置

更新现有的存储配置。

**接口地址:** `PUT /storages/:id`

**路径参数:**

| 参数名 | 类型   | 必填 | 说明     |
|--------|--------|------|----------|
| id     | string | 是   | 存储 ID  |

**请求体:** (所有字段可选)
```json
{
  "storage_name": "新名称",
  "is_enabled": true,
  "is_default": false
}
```

**注意事项：**
- 如果要将存储设置为默认 (`is_default: true`)，但系统中已存在其他默认存储，更新会失败
- 必须先将其他存储的 `is_default` 设为 `false`，才能设置新的默认存储
- 错误响应示例（当已存在默认存储时）：
```json
{
  "code": "400031112",
  "message": "Default storage already exists",
  "description": "A default storage already exists: 阿里云存储1",
  "solution": "Please disable the current default storage before setting a new one"
}
```

**响应示例:**
```json
{
  "data": {
    "id": "2035957924035170304",
    "status": "ok"
  }
}
```

---

### 5. 删除存储配置

删除指定的存储配置。

**接口地址:** `DELETE /storages/:id`

**路径参数:**

| 参数名 | 类型   | 必填 | 说明     |
|--------|--------|------|----------|
| id     | string | 是   | 存储 ID  |

**响应示例:**
```json
{
  "data": {
    "id": "2034170764420321280",
    "status": "ok"
  }
}
```

---

### 6. 检查存储连接

测试与对象存储服务的连接。

**接口地址:** `POST /storages/:id/check`

**路径参数:**

| 参数名 | 类型   | 必填 | 说明     |
|--------|--------|------|----------|
| id     | string | 是   | 存储 ID  |

**响应示例:**
```json
{
  "data": {
    "connected": true
  }
}
```

---

## 对象元数据 API

### 7. 获取对象元数据 URL

获取对象 HEAD 请求的预签名 URL。

**接口地址:** `GET /head/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名           | 类型    | 必填 | 说明                          |
|-----------------|---------|------|------------------------------|
| internal_request| boolean | 否   | 是否使用内网端点，默认 false    |
| expires         | int64   | 否   | URL 过期时间（秒），默认 3600  |

**响应示例:**
```json
{
  "data": {
    "method": "HEAD",
    "url": "https://oss-yexiaoyan.oss-cn-shanghai.aliyuncs.com/test/test.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=LTAI5tJ3NTKj3oGmF7ag8x7u%2F20260323%2Fcn-shanghai%2Fs3%2Faws4_request&X-Amz-Date=20260323T060812Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=f60194d088658be26b1d88d132f7acdbfff1f7d29c6c4ea0535f29bf8bdac6e6",
    "headers": {},
    "form_field": null,
    "body": ""
  }
}
```

---

### 8. 批量获取对象元数据 URL

批量获取多个对象的 HEAD 请求预签名 URL。

**接口地址:** `POST /head/:storageId`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明     |
|-----------|--------|------|----------|
| storageId | string | 是   | 存储 ID  |

**请求体:**
```json
{
  "keys": [
    "path/to/file1.txt",
    "path/to/file2.txt",
    "path/to/file3.txt"
  ],
  "internal_request": false,
  "expires": 3600
}
```

**字段说明:**

| 字段名           | 类型    | 必填 | 说明                                |
|-----------------|---------|------|-------------------------------------|
| keys            | array   | 是   | 对象 Key 列表，最多 100 个           |
| internal_request| boolean | 否   | 是否使用内网端点，默认 false          |
| expires         | int64   | 否   | URL 过期时间（秒），默认 3600        |

**响应示例:**
```json
{
  "data": {
    "test/test.txt": {
      "method": "HEAD",
      "url": "https://oss-yexiaoyan.oss-cn-shanghai.aliyuncs.com/test/test.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=LTAI5tJ3NTKj3oGmF7ag8x7u%2F20260323%2Fcn-shanghai%2Fs3%2Faws4_request&X-Amz-Date=20260323T060850Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=cfe029b1b7a1a72ea524f7c95b4d7656bae547954a663deb3c5f84fc8bc86b42",
      "headers": {},
      "form_field": null,
      "body": ""
    }
  }
}
```

---

## 单文件上传 API

### 9. 获取单文件上传 URL

获取单文件上传的预签名 URL。

**接口地址:** `GET /upload/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名           | 类型    | 必填 | 说明                    |
|-----------------|---------|------|-----------------------|
| request_method  | string  | 否   | 上传方式 (POST/PUT)，默认 PUT |
| expires         | int64   | 否   | URL 过期时间（秒）           |
| internal_request| boolean | 否   | 是否使用内网端点，默认 false     |

**响应示例 (PUT):**
```json
{
  "data": {
    "method": "PUT",
    "url": "https://oss-yexiaoyan.oss-cn-shanghai.aliyuncs.com/test/file.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=LTAI5tJ3NTKj3oGmF7ag8x7u%2F20260323%2Fcn-shanghai%2Fs3%2Faws4_request&X-Amz-Date=20260323T060933Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=779ed198ce467ba615dd14753a41385993e4e3b8140f7f4dc72a47b203045c0d",
    "headers": {},
    "form_field": null,
    "body": ""
  }
}
```

**响应示例 (POST):**
```json
{
  "data": {
    "method": "POST",
    "url": "https://oss-yexiaoyan.oss-cn-shanghai.aliyuncs.com/",
    "headers": {},
    "form_field": {
      "bucket": "oss-yexiaoyan",
      "key": "test/file.txt",
      "policy": "eyJleHBpcmF0aW9uIjoiMjAyNi0wMy0yM1QwNzowOTo0OS4yNTFaIiwiY29uZGl0aW9ucyI6W1siZXEiLCIkYnVja2V0Iiwib3NzLXlleGlhb3lhbiJdLFsiZXEiLCIka2V5IiwidGVzdC9maWxlLnR4dCJdLFsiZXEiLCIkeC1hbXotZGF0ZSIsIjIwMjYwMzIzVDA2MDk0OVoiXSxbImVxIiwiJHgtYW16LWFsZ29yaXRobSIsIkFXUzQtSE1BQy1TSEEyNTYiXSxbImVxIiwiJHgtYW16LWNyZWRlbnRpYWwiLCJMVEFJNXRKM05US2ozb0dtRjdhZzh4N3UvMjAyNjAzMjMvY24tc2hhbmdoYWkvczMvYXdzNF9yZXF1ZXN0Il1dfQ==",
      "x-amz-algorithm": "AWS4-HMAC-SHA256",
      "x-amz-credential": "LTAI5tJ3NTKj3oGmF7ag8x7u/20260323/cn-shanghai/s3/aws4_request",
      "x-amz-date": "20260323T060949Z",
      "x-amz-signature": "bba105b3d61a3e59f1d8ab37843d80c9cd81adae7081f13ddb11903bf0c3cdad"
    },
    "body": ""
  }
}
```

---

## 分片上传 API

### 10. 初始化分片上传

初始化分片上传任务。

**接口地址:** `GET /initmultiupload/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名 | 类型  | 必填 | 说明                |
|--------|-------|------|---------------------|
| size   | int64 | 是   | 文件总大小（字节）    |

**响应示例:**
```json
{
  "data": {
    "upload_id": "7BA9BC4C2C264A8A8C45C21338ED1513",
    "part_size": 5242880,
    "key": "videos/test123.mp4"
  }
}
```

**字段说明:**

| 字段名     | 类型   | 说明                          |
|-----------|--------|------------------------------|
| upload_id | string | 上传会话 ID                   |
| part_size | int64  | 建议的分片大小（字节）          |
| key       | string | 对象 Key                     |

---

### 11. 获取分片上传 URL

获取各个分片的上传 URL。

**接口地址:** `POST /uploadpart/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**请求体:**
```json
{
  "upload_id": "0004B9894A22E5B1F2A2A778BC7D4C19",
  "part_id": [1, 2, 3, 4, 5],
  "internal_request": false
}
```

**字段说明:**

| 字段名           | 类型    | 必填 | 说明                              |
|-----------------|---------|------|----------------------------------|
| upload_id       | string  | 是   | 初始化时返回的 upload_id           |
| part_id         | array   | 是   | 分片编号列表 (1-10000)             |
| internal_request| boolean | 否   | 是否使用内网端点，默认 false        |

**响应示例:**
```json
{
  "data": {
    "authrequest": {
      "1": {
        "method": "PUT",
        "url": "https://oss-yexiaoyan.oss-cn-shanghai.aliyuncs.com/videos/test123.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=LTAI5tJ3NTKj3oGmF7ag8x7u%2F20260323%2Fcn-shanghai%2Fs3%2Faws4_request&X-Amz-Date=20260323T061116Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&partNumber=1&uploadId=0004B9894A22E5B1F2A2A778BC7D4C19&X-Amz-Signature=e38428782ba1cffa5cd310bb7e9eff57c0395e0457570550f919e99c6891983d",
        "headers": {},
        "form_field": null,
        "body": ""
      },
      "2": {
        "method": "PUT",
        "url": "https://oss-yexiaoyan.oss-cn-shanghai.aliyuncs.com/videos/test123.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=LTAI5tJ3NTKj3oGmF7ag8x7u%2F20260323%2Fcn-shanghai%2Fs3%2Faws4_request&X-Amz-Date=20260323T061116Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&partNumber=2&uploadId=0004B9894A22E5B1F2A2A778BC7D4C19&X-Amz-Signature=f948caf1f0f91dfd741920cebf4712a78118da5c8be492287fda68471a3d657d",
        "headers": {},
        "form_field": null,
        "body": ""
      },
      "3": {
        "method": "PUT",
        "url": "https://oss-yexiaoyan.oss-cn-shanghai.aliyuncs.com/videos/test123.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=LTAI5tJ3NTKj3oGmF7ag8x7u%2F20260323%2Fcn-shanghai%2Fs3%2Faws4_request&X-Amz-Date=20260323T061116Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&partNumber=3&uploadId=0004B9894A22E5B1F2A2A778BC7D4C19&X-Amz-Signature=ca23c7dc47082e003f9c2ed93df5f177678014a2fd39c820d6da026a9322517a",
        "headers": {},
        "form_field": null,
        "body": ""
      },
      "4": {
        "method": "PUT",
        "url": "https://oss-yexiaoyan.oss-cn-shanghai.aliyuncs.com/videos/test123.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=LTAI5tJ3NTKj3oGmF7ag8x7u%2F20260323%2Fcn-shanghai%2Fs3%2Faws4_request&X-Amz-Date=20260323T061116Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&partNumber=4&uploadId=0004B9894A22E5B1F2A2A778BC7D4C19&X-Amz-Signature=f01e902a16052d74e6620ae99d2f56e8ce50731484b8e4a67054c8fa8cb34e47",
        "headers": {},
        "form_field": null,
        "body": ""
      },
      "5": {
        "method": "PUT",
        "url": "https://oss-yexiaoyan.oss-cn-shanghai.aliyuncs.com/videos/test123.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=LTAI5tJ3NTKj3oGmF7ag8x7u%2F20260323%2Fcn-shanghai%2Fs3%2Faws4_request&X-Amz-Date=20260323T061116Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&partNumber=5&uploadId=0004B9894A22E5B1F2A2A778BC7D4C19&X-Amz-Signature=e56e1b29862bdfd40d8306f55a5081051ee0a4e7e73c831f610c877256cea003",
        "headers": {},
        "form_field": null,
        "body": ""
      }
    }
  }
}
```

---

### 12. 完成分片上传

完成分片上传并合并文件。

**接口地址:** `POST /completeupload/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名     | 类型   | 必填 | 说明        |
|-----------|--------|------|-------------|
| upload_id | string | 是   | 上传 ID     |

**请求体:**
```json
{
  "1": "\"5eb63bbbe01eeed093cb22bb8f5acdc3\"",
  "2": "\"5eb63bbbe01eeed093cb22bb8f5acdc4\"",
  "3": "\"5eb63bbbe01eeed093cb22bb8f5acdc5\""
}
```

**注意:** Key 为分片编号（字符串），Value 为对应分片的 ETag（必须包含双引号）。

**响应示例:**
```json
{
  "data": {
    "method": "POST",
    "url": "https://...?uploadId=xxx",
    "headers": {
      "Content-Type": "application/xml"
    },
    "request_body": "<CompleteMultipartUpload><Part><PartNumber>1</PartNumber><ETag>\"etag1\"</ETag></Part>...</CompleteMultipartUpload>"
  }
}
```

---

## 下载 API

### 13. 获取下载 URL

获取文件下载的预签名 URL。

**接口地址:** `GET /download/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名           | 类型    | 必填 | 说明                              |
|-----------------|---------|------|----------------------------------|
| expires         | int64   | 否   | URL 过期时间（秒）                 |
| save_name       | string  | 否   | 下载时的文件名 (需 URL 编码)        |
| internal_request| boolean | 否   | 是否使用内网端点，默认 false        |

**响应示例:**
```json
{
  "data": {
    "method": "GET",
    "url": "https://obs.cn-east-3.myhuaweicloud.com/obs-as7/test/file.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=IEWQ43FVSDUADX40BFNI%2F20260318%2Fcn-east-3%2Fs3%2Faws4_request&X-Amz-Date=20260318T034412Z&X-Amz-Expires=7200&X-Amz-SignedHeaders=host&response-content-disposition=attachment%3B%20filename%2A%3Dutf-8%27%27test.txt&X-Amz-Signature=27f54fbc559da514c1fa4eca2c139dcd8bfc52247cb738c1aaaeb8d5ddc70f35",
    "headers": {},
    "form_field": null,
    "body": ""
  }
}
```

---

## 删除 API

### 14. 获取删除 URL

获取文件删除的预签名 URL。

**接口地址:** `GET /delete/:storageId/:key`

**路径参数:**

| 参数名     | 类型   | 必填 | 说明                      |
|-----------|--------|------|--------------------------|
| storageId | string | 是   | 存储 ID                   |
| key       | string | 是   | 对象 Key (需 URL 编码)     |

**查询参数:**

| 参数名           | 类型    | 必填 | 说明                          |
|-----------------|---------|------|------------------------------|
| expires         | int64   | 否   | URL 过期时间（秒）             |
| internal_request| boolean | 否   | 是否使用内网端点，默认 false    |

**响应示例:**
```json
{
  "data": {
    "method": "DELETE",
    "url": "https://oss-yexiaoyan.oss-cn-shanghai.aliyuncs.com/test/file.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=LTAI5tJ3NTKj3oGmF7ag8x7u%2F20260323%2Fcn-shanghai%2Fs3%2Faws4_request&X-Amz-Date=20260323T061618Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=5c082c9581b206ceaa4e4957b9e28d7fe8723d49c3145b58fb86e72a18a883ed",
    "headers": {},
    "form_field": null,
    "body": ""
  }
}
```

---

## 健康检查 API

### 15. 就绪检查

检查服务是否准备好接收流量（Kubernetes Readiness Probe）。

**接口地址:** `GET /health/ready`

**响应示例（就绪）:**
```json
{
  "checks": {
    "database": "ok",
    "redis": "ok (standalone)"
  },
  "res": 0,
  "status": "ok"
}
```

**响应示例（未就绪）:**
```json
{
  "code": "503031100",
  "message": "Service not ready",
  "description": "One or more services are not ready",
  "checks": {
    "database": "failed",
    "redis": "failed"
  }
}
```

---

### 16. 存活检查

检查服务进程是否正常运行（Kubernetes Liveness Probe）。

**接口地址:** `GET /health/alive`

**响应示例:**
```json
{
  "res": 0,
  "status": "ok"
}
```

---

## 注意事项

### 默认存储

- 系统全局只能有**一个**默认存储
- **创建存储**：如果设置 `is_default: true`，但系统中已存在其他默认存储，创建会失败（错误码：400031112）
- **更新存储**：如果要将存储设置为默认（`is_default: true`），但系统中已存在其他默认存储，更新会失败（错误码：400031112）
- **安全机制**：必须先手动将现有默认存储的 `is_default` 设为 `false`，才能设置新的默认存储
- **查询默认存储**：
  - 方法1：`GET /storages?is_default=true` - 直接筛选默认存储
  - 方法2：`GET /storages?enabled=true&is_default=true` - 筛选启用的默认存储
  - 返回结果中获取 `storage_id` 用于后续文件操作
- **注意**：目前所有文件操作接口（上传、下载、删除等）都需要显式指定 `storage_id` 参数，不支持自动使用默认存储

### URL 编码

对象 Key 中的特殊字符需要进行 URL 编码：
- `/` → `%2F`
- ` ` (空格) → `%20`
- `+` → `%2B`
- `?` → `%3F`
- `&` → `%26`
- `=` → `%3D`

### ETag 格式

完成分片上传时，ETag 必须包含双引号：
```json
{
  "1": "\"5eb63bbbe01eeed093cb22bb8f5acdc3\"",
  "2": "\"5eb63bbbe01eeed093cb22bb8f5acdc4\""
}
```

### 文件大小限制

- 单文件上传：最大 5GB（取决于厂商限制）
- 分片上传：推荐用于 > 100MB 的文件
- 最大文件大小：50TB（通过分片上传）

### URL 过期时间

- 默认：3600 秒（1 小时）
- 最小：1 秒
- 最大：604800 秒（7 天）

### 内网与外网

使用 `internal_request=true` 的场景：
- 网关和对象存储在同一 VPC 内
- 需要更快的上传/下载速度
- 希望节省带宽成本

---
