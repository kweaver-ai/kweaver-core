# OSS 网关文件上传下载使用指南

本文档详细说明如何使用 OSS 网关服务进行文件上传、下载和删除操作。

---

## 目录

- [核心概念](#核心概念)
- [单文件上传](#单文件上传)
  - [PUT 方式上传](#put-方式上传)
  - [POST 表单上传](#post-表单上传)
- [分片上传（大文件）](#分片上传大文件)
- [文件下载](#文件下载)
- [文件删除](#文件删除)
- [常见问题](#常见问题)

---

## 核心概念

### 预签名 URL 机制

OSS 网关使用预签名 URL 机制，工作流程如下：

```
┌────────┐   ① 请求预签名URL    ┌──────────┐
│ 客户端 │ ──────────────────> │ OSS 网关  │
│        │   ② 返回预签名URL    │          │
│        │ <────────────────── │          │
└────┬───┘                     └──────────┘
     │
     │ ③ 使用预签名URL直连对象存储
     ↓
┌─────────────┐
│ 对象存储服务 │
│  (OSS/OBS)  │
└─────────────┘
```

**优势：**
- 减轻网关负载（不代理文件数据）
- 提高传输速度（客户端直连对象存储）
- 安全可控（URL 有过期时间和权限限制）

---

## 单文件上传

### 适用场景

- 文件大小 < 100MB
- 需要快速上传
- 一次性上传完整文件

### PUT 方式上传

PUT 方式是最简单直接的上传方式，适合大多数场景。

#### 步骤 1：获取上传 URL

```bash
curl -X GET "http://localhost:8080/api/v1/upload/A1B2C3D4E5F6G7H8/test%2Ffile.txt?request_method=PUT"
```

**参数说明：**
- `A1B2C3D4E5F6G7H8`: 存储 ID（替换为实际值）
- `test%2Ffile.txt`: 对象 Key（`test/file.txt` URL 编码后），该参数可以理解为在bucket中的存储路径，文件名file.txt无需和真实的文件名一样，该名称类似重命名效果，使用时建议使用uuid随机生成，不要直接使用本地文件名，避免冲突
- `request_method=PUT`: 指定使用 PUT 方式

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "method": "PUT",
    "url": "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?Expires=1234567890&OSSAccessKeyId=xxx&Signature=xxx",
    "headers": {
      "Content-Type": "application/octet-stream"
    }
  }
}
```

#### 步骤 2：使用预签名 URL 上传文件

```bash
curl -X PUT \
  "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?Expires=1234567890&OSSAccessKeyId=xxx&Signature=xxx" \
  --data-binary @/path/to/local/file.txt
```

**注意：** 不要添加额外的 Header（如 Content-Type），除非在生成预签名 URL 时已明确指定，否则会导致签名验证失败。

**成功响应：**
- HTTP 状态码：200 或 204
- 响应头包含 `ETag`

#### 完整示例：上传图片

```bash
# 1. 获取上传 URL
RESPONSE=$(curl -s -X GET \
  "http://localhost:8080/api/v1/upload/A1B2C3D4E5F6G7H8/images%2Flogo.png?request_method=PUT")

# 2. 提取预签名 URL
UPLOAD_URL=$(echo $RESPONSE | jq -r '.data.url')

# 3. 上传文件
curl -X PUT "$UPLOAD_URL" \
  --data-binary @logo.png

echo "上传成功！"
```

---

### POST 表单上传

POST 表单方式适合浏览器环境或需要额外表单字段的场景。

#### 步骤 1：获取上传 URL 和表单字段

```bash
curl -X GET "http://localhost:8080/api/v1/upload/A1B2C3D4E5F6G7H8/test/file.txt?request_method=POST"
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "method": "POST",
    "url": "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/",
    "headers": {},
    "form_field": {
      "OSSAccessKeyId": "LTAI5xxx",
      "policy": "eyJleHBpcmF0aW9uIjoi...",
      "Signature": "abc123...",
      "key": "test/file.txt",
      "x-oss-signature-version": "OSS2",
      "x-oss-algorithm": "OSS2-HMAC-SHA256",
      "x-oss-credential": "LTAI5xxx/20240101/cn-hangzhou/oss/aliyun_v2_request",
      "x-oss-date": "20240101T000000Z"
    }
  }
}
```

#### 步骤 2：使用表单上传文件

```bash
curl -X POST "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/" \
  -F "OSSAccessKeyId=LTAI5xxx" \
  -F "policy=eyJleHBpcmF0aW9uIjoi..." \
  -F "Signature=abc123..." \
  -F "key=test/file.txt" \
  -F "x-oss-signature-version=OSS2" \
  -F "x-oss-algorithm=OSS2-HMAC-SHA256" \
  -F "x-oss-credential=LTAI5xxx/20240101/cn-hangzhou/oss/aliyun_v2_request" \
  -F "x-oss-date=20240101T000000Z" \
  -F "file=@/path/to/local/file.txt"
```

**注意：**
- 所有 `form_field` 中的字段都必须包含
- `file` 字段必须放在最后

#### 完整脚本示例

```bash
#!/bin/bash

# 配置
GATEWAY_URL="http://localhost:8080/api/v1"
STORAGE_ID="A1B2C3D4E5F6G7H8"
OBJECT_KEY="test/file.txt"
LOCAL_FILE="/path/to/local/file.txt"

# URL 编码对象 Key
ENCODED_KEY=$(echo -n "$OBJECT_KEY" | jq -sRr @uri)

# 1. 获取上传信息
echo "获取上传 URL..."
RESPONSE=$(curl -s -X GET \
  "$GATEWAY_URL/upload/$STORAGE_ID/$ENCODED_KEY?request_method=POST")

# 2. 提取字段
URL=$(echo $RESPONSE | jq -r '.data.url')
OSS_ACCESS_KEY_ID=$(echo $RESPONSE | jq -r '.data.form_field.OSSAccessKeyId')
POLICY=$(echo $RESPONSE | jq -r '.data.form_field.policy')
SIGNATURE=$(echo $RESPONSE | jq -r '.data.form_field.Signature')
KEY=$(echo $RESPONSE | jq -r '.data.form_field.key')
SIG_VERSION=$(echo $RESPONSE | jq -r '.data.form_field."x-oss-signature-version"')
ALGORITHM=$(echo $RESPONSE | jq -r '.data.form_field."x-oss-algorithm"')
CREDENTIAL=$(echo $RESPONSE | jq -r '.data.form_field."x-oss-credential"')
DATE=$(echo $RESPONSE | jq -r '.data.form_field."x-oss-date"')

# 3. 上传文件
echo "上传文件..."
curl -X POST "$URL" \
  -F "OSSAccessKeyId=$OSS_ACCESS_KEY_ID" \
  -F "policy=$POLICY" \
  -F "Signature=$SIGNATURE" \
  -F "key=$KEY" \
  -F "x-oss-signature-version=$SIG_VERSION" \
  -F "x-oss-algorithm=$ALGORITHM" \
  -F "x-oss-credential=$CREDENTIAL" \
  -F "x-oss-date=$DATE" \
  -F "file=@$LOCAL_FILE"

echo "上传完成！"
```

---

## 分片上传（大文件）

### 适用场景

- 文件大小 > 100MB
- 需要断点续传
- 网络不稳定，需要分片重传
- 超大文件（最大支持 50TB）

### 工作流程

```
初始化分片上传 → 获取分片上传URL → 上传各分片 → 完成分片上传 → 执行合并
```

### 完整示例：上传 200MB 视频

假设要上传文件：`video.mp4`，大小 200MB

#### 步骤 1：初始化分片上传

```bash
# 文件大小：200MB = 209715200 字节
FILE_SIZE=209715200

curl -X GET \
  "http://localhost:8080/api/v1/initmultiupload/A1B2C3D4E5F6G7H8/videos%2Fvideo.mp4?size=$FILE_SIZE"
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "upload_id": "0004B9894A22E5B1F2A2A778BC7D4C19",
    "part_size": 5242880,
    "key": "videos/video.mp4"
  }
}
```

**说明：**
- `upload_id`: 上传会话 ID（务必保存）
- `part_size`: 建议分片大小（5MB）
- 分片数量：200MB ÷ 5MB = 40 个分片

#### 步骤 2：分割文件

**Linux/Mac 分片：**
```bash
# 分割成 5MB 的分片
split -b 5M video.mp4 video.mp4.part
```

**Windows PowerShell 分片：**
```powershell
$file = "video.mp4"
$chunkSize = 5242880  # 5MB
$stream = [System.IO.File]::OpenRead($file)
$buffer = New-Object byte[] $chunkSize
$partNumber = 1

while (($bytesRead = $stream.Read($buffer, 0, $chunkSize)) -gt 0) {
    $output = [System.IO.File]::Create("video.mp4.part$partNumber")
    $output.Write($buffer, 0, $bytesRead)
    $output.Close()
    $partNumber++
}
$stream.Close()
```

#### 步骤 3：获取所有分片的上传 URL

```bash
UPLOAD_ID="0004B9894A22E5B1F2A2A778BC7D4C19"

curl -X POST \
  "http://localhost:8080/api/v1/uploadpart/A1B2C3D4E5F6G7H8/videos%2Fvideo.mp4" \
  -H "Content-Type: application/json" \
  -d '{
    "upload_id": "'"$UPLOAD_ID"'",
    "part_id": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40]
  }'
```

**响应示例（部分）：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "authrequest": {
      "1": {
        "method": "PUT",
        "url": "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/videos/video.mp4?partNumber=1&uploadId=xxx...",
        "headers": {
          "Content-Type": "application/octet-stream"
        }
      },
      "2": {
        "method": "PUT",
        "url": "https://...",
        "headers": {
          "Content-Type": "application/octet-stream"
        }
      }
    }
  }
}
```

#### 步骤 4：上传每个分片并保存 ETag

```bash
#!/bin/bash

UPLOAD_ID="0004B9894A22E5B1F2A2A778BC7D4C19"
TOTAL_PARTS=40

# 创建 JSON 文件存储 ETags
echo "{" > etags.json

# 上传每个分片
for i in $(seq 1 $TOTAL_PARTS); do
  echo "上传分片 $i/$TOTAL_PARTS ..."
  
  # 获取该分片的上传 URL
  PART_URL=$(curl -s -X POST \
    "http://localhost:8080/api/v1/uploadpart/A1B2C3D4E5F6G7H8/videos%2Fvideo.mp4" \
    -H "Content-Type: application/json" \
    -d "{\"upload_id\": \"$UPLOAD_ID\", \"part_id\": [$i]}" \
    | jq -r ".data.authrequest[\"$i\"].url")
  
  # 上传分片并提取 ETag
  ETAG=$(curl -s -i -X PUT "$PART_URL" \
    -H "Content-Type: application/octet-stream" \
    --data-binary "@video.mp4.part$i" \
    | grep -i "ETag:" | awk '{print $2}' | tr -d '\r\n')
  
  # 保存 ETag
  echo "  \"$i\": $ETAG," >> etags.json
  
  echo "分片 $i 上传完成，ETag: $ETAG"
done

# 移除最后一个逗号并关闭 JSON
sed -i '$ s/,$//' etags.json
echo "}" >> etags.json

echo "所有分片上传完成！"
```

**生成的 etags.json：**
```json
{
  "1": "\"5eb63bbbe01eeed093cb22bb8f5acdc3\"",
  "2": "\"7fc93cbbe01eeed093cb22bb8f5acdc4\"",
  "3": "\"9ab13cbbe01eeed093cb22bb8f5acdc5\"",
  ...
  "40": "\"1234567890abcdef1234567890abcdef\""
}
```

#### 步骤 5：请求完成分片上传

```bash
UPLOAD_ID="0004B9894A22E5B1F2A2A778BC7D4C19"

curl -X POST \
  "http://localhost:8080/api/v1/completeupload/A1B2C3D4E5F6G7H8/videos%2Fvideo.mp4?upload_id=$UPLOAD_ID" \
  -H "Content-Type: application/json" \
  -d @etags.json
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "method": "POST",
    "url": "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/videos/video.mp4?uploadId=xxx",
    "headers": {
      "Content-Type": "application/xml"
    },
    "request_body": "<CompleteMultipartUpload><Part><PartNumber>1</PartNumber><ETag>\"5eb63bbbe01eeed093cb22bb8f5acdc3\"</ETag></Part>...</CompleteMultipartUpload>"
  }
}
```

#### 步骤 6：执行文件合并

```bash
# 提取 URL 和请求体
COMPLETE_URL=$(echo $RESPONSE | jq -r '.data.url')
REQUEST_BODY=$(echo $RESPONSE | jq -r '.data.request_body')

# 执行合并
curl -X POST "$COMPLETE_URL" \
  -H "Content-Type: application/xml" \
  -d "$REQUEST_BODY"
```

**成功响应：**
- HTTP 状态码：200
- 文件上传完成

### 完整自动化脚本

```bash
#!/bin/bash

# ============ 配置 ============
GATEWAY_URL="http://localhost:8080/api/v1"
STORAGE_ID="A1B2C3D4E5F6G7H8"
LOCAL_FILE="video.mp4"
OBJECT_KEY="videos/video.mp4"
PART_SIZE=5242880  # 5MB

# ============ 函数定义 ============
url_encode() {
  echo -n "$1" | jq -sRr @uri
}

# ============ 1. 初始化分片上传 ============
echo "========== 初始化分片上传 =========="
FILE_SIZE=$(stat -f%z "$LOCAL_FILE" 2>/dev/null || stat -c%s "$LOCAL_FILE")
ENCODED_KEY=$(url_encode "$OBJECT_KEY")

INIT_RESPONSE=$(curl -s -X GET \
  "$GATEWAY_URL/initmultiupload/$STORAGE_ID/$ENCODED_KEY?size=$FILE_SIZE")

UPLOAD_ID=$(echo $INIT_RESPONSE | jq -r '.data.upload_id')
SUGGESTED_PART_SIZE=$(echo $INIT_RESPONSE | jq -r '.data.part_size')

echo "Upload ID: $UPLOAD_ID"
echo "建议分片大小: $SUGGESTED_PART_SIZE 字节"

# 计算分片数量
TOTAL_PARTS=$(( ($FILE_SIZE + $SUGGESTED_PART_SIZE - 1) / $SUGGESTED_PART_SIZE ))
echo "总分片数: $TOTAL_PARTS"

# ============ 2. 分割文件 ============
echo ""
echo "========== 分割文件 =========="
split -b $SUGGESTED_PART_SIZE "$LOCAL_FILE" "${LOCAL_FILE}.part"
echo "文件分割完成"

# ============ 3. 上传所有分片 ============
echo ""
echo "========== 上传分片 =========="

# 生成分片 ID 数组
PART_IDS=$(seq -s, 1 $TOTAL_PARTS)

# 获取所有分片的上传 URL
PARTS_RESPONSE=$(curl -s -X POST \
  "$GATEWAY_URL/uploadpart/$STORAGE_ID/$ENCODED_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"upload_id\": \"$UPLOAD_ID\", \"part_id\": [$PART_IDS]}")

# 上传每个分片
echo "{" > etags.json
PART_FILES=(${LOCAL_FILE}.part*)

for i in $(seq 1 $TOTAL_PARTS); do
  PART_FILE="${PART_FILES[$((i-1))]}"
  
  echo "上传分片 $i/$TOTAL_PARTS ($PART_FILE)..."
  
  # 提取该分片的 URL
  PART_URL=$(echo $PARTS_RESPONSE | jq -r ".data.authrequest[\"$i\"].url")
  
  # 上传分片
  UPLOAD_RESULT=$(curl -s -i -X PUT "$PART_URL" \
    -H "Content-Type: application/octet-stream" \
    --data-binary "@$PART_FILE")
  
  # 提取 ETag
  ETAG=$(echo "$UPLOAD_RESULT" | grep -i "ETag:" | awk '{print $2}' | tr -d '\r\n')
  
  # 保存 ETag
  if [ $i -lt $TOTAL_PARTS ]; then
    echo "  \"$i\": $ETAG," >> etags.json
  else
    echo "  \"$i\": $ETAG" >> etags.json
  fi
  
  echo "  ✓ 分片 $i 上传完成 (ETag: $ETAG)"
done

echo "}" >> etags.json
echo "所有分片上传完成！"

# ============ 4. 完成分片上传 ============
echo ""
echo "========== 完成分片上传 =========="

COMPLETE_RESPONSE=$(curl -s -X POST \
  "$GATEWAY_URL/completeupload/$STORAGE_ID/$ENCODED_KEY?upload_id=$UPLOAD_ID" \
  -H "Content-Type: application/json" \
  -d @etags.json)

COMPLETE_URL=$(echo $COMPLETE_RESPONSE | jq -r '.data.url')
REQUEST_BODY=$(echo $COMPLETE_RESPONSE | jq -r '.data.request_body')

# ============ 5. 执行合并 ============
echo "执行文件合并..."

MERGE_RESULT=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$COMPLETE_URL" \
  -H "Content-Type: application/xml" \
  -d "$REQUEST_BODY")

HTTP_STATUS=$(echo "$MERGE_RESULT" | grep "HTTP_STATUS:" | cut -d: -f2)

if [ "$HTTP_STATUS" = "200" ]; then
  echo "✓ 文件上传成功！"
else
  echo "✗ 文件上传失败，HTTP 状态码: $HTTP_STATUS"
  exit 1
fi

# ============ 6. 清理临时文件 ============
echo ""
echo "========== 清理临时文件 =========="
rm -f ${LOCAL_FILE}.part*
rm -f etags.json
echo "清理完成"

echo ""
echo "========== 完成 =========="
echo "文件已成功上传到: $OBJECT_KEY"
```

**使用方法：**
```bash
chmod +x multipart_upload.sh
./multipart_upload.sh
```

---

## 文件下载

### 步骤 1：获取下载 URL

```bash
curl -X GET \
  "http://localhost:8080/api/v1/download/A1B2C3D4E5F6G7H8/videos%2Fvideo.mp4?save_name=downloaded_video.mp4&expires=7200"
```

**参数说明：**
- `save_name`: 下载时的文件名（可选）
- `expires`: URL 有效期，7200秒 = 2小时（可选）

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "method": "GET",
    "url": "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/videos/video.mp4?response-content-disposition=attachment%3B%20filename%3Ddownloaded_video.mp4&...",
    "headers": {}
  }
}
```

### 步骤 2：下载文件

```bash
# 提取下载 URL
DOWNLOAD_URL=$(echo $RESPONSE | jq -r '.data.url')

# 下载文件
curl -o "local_video.mp4" "$DOWNLOAD_URL"
```

### 完整下载脚本

```bash
#!/bin/bash

# 配置
GATEWAY_URL="http://localhost:8080/api/v1"
STORAGE_ID="A1B2C3D4E5F6G7H8"
OBJECT_KEY="videos/video.mp4"
SAVE_NAME="downloaded_video.mp4"
LOCAL_FILE="local_video.mp4"

# URL 编码
ENCODED_KEY=$(echo -n "$OBJECT_KEY" | jq -sRr @uri)
ENCODED_SAVE_NAME=$(echo -n "$SAVE_NAME" | jq -sRr @uri)

# 获取下载 URL
echo "获取下载 URL..."
RESPONSE=$(curl -s -X GET \
  "$GATEWAY_URL/download/$STORAGE_ID/$ENCODED_KEY?save_name=$ENCODED_SAVE_NAME&expires=7200")

# 提取 URL
DOWNLOAD_URL=$(echo $RESPONSE | jq -r '.data.url')

if [ "$DOWNLOAD_URL" = "null" ]; then
  echo "错误：获取下载 URL 失败"
  echo $RESPONSE | jq .
  exit 1
fi

# 下载文件
echo "开始下载..."
curl -# -o "$LOCAL_FILE" "$DOWNLOAD_URL"

echo "下载完成！文件保存为: $LOCAL_FILE"
```

### 在浏览器中下载

直接在浏览器地址栏粘贴预签名 URL：
```
https://my-bucket.oss-cn-hangzhou.aliyuncs.com/videos/video.mp4?...
```

浏览器会自动下载文件。

---

## 文件删除

### 步骤 1：获取删除 URL

```bash
curl -X GET \
  "http://localhost:8080/api/v1/delete/A1B2C3D4E5F6G7H8/test%2Ffile.txt?expires=3600"
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "method": "DELETE",
    "url": "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?...",
    "headers": {}
  }
}
```

### 步骤 2：执行删除操作

```bash
# 提取删除 URL
DELETE_URL=$(echo $RESPONSE | jq -r '.data.url')

# 删除文件
curl -X DELETE "$DELETE_URL"
```

**成功响应：**
- HTTP 状态码：204 (No Content)

### 完整删除脚本

```bash
#!/bin/bash

# 配置
GATEWAY_URL="http://localhost:8080/api/v1"
STORAGE_ID="A1B2C3D4E5F6G7H8"
OBJECT_KEY="test/file.txt"

# URL 编码
ENCODED_KEY=$(echo -n "$OBJECT_KEY" | jq -sRr @uri)

# 获取删除 URL
echo "获取删除 URL..."
RESPONSE=$(curl -s -X GET \
  "$GATEWAY_URL/delete/$STORAGE_ID/$ENCODED_KEY")

# 提取 URL
DELETE_URL=$(echo $RESPONSE | jq -r '.data.url')

if [ "$DELETE_URL" = "null" ]; then
  echo "错误：获取删除 URL 失败"
  echo $RESPONSE | jq .
  exit 1
fi

# 确认删除
read -p "确定要删除文件 '$OBJECT_KEY' 吗？(y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "取消删除"
  exit 0
fi

# 执行删除
echo "删除文件..."
HTTP_STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X DELETE "$DELETE_URL")

if [ "$HTTP_STATUS" = "204" ]; then
  echo "✓ 文件删除成功！"
else
  echo "✗ 文件删除失败，HTTP 状态码: $HTTP_STATUS"
  exit 1
fi
```

---

## 常见问题

### Q1: URL 编码问题

**问题：** 对象 Key 包含特殊字符，请求失败

**解决方案：**

路径中的特殊字符必须进行 URL 编码：

| 字符 | 编码    | 示例                          |
|------|---------|-------------------------------|
| `/`  | `%2F`   | `folder/file.txt` → `folder%2Ffile.txt` |
| 空格 | `%20`   | `my file.txt` → `my%20file.txt` |
| `+`  | `%2B`   | `file+1.txt` → `file%2B1.txt` |
| `?`  | `%3F`   | `file?.txt` → `file%3F.txt` |
| `&`  | `%26`   | `file&data.txt` → `file%26data.txt` |

**Shell 脚本编码：**
```bash
# 使用 jq 进行 URL 编码
ENCODED=$(echo -n "folder/file.txt" | jq -sRr @uri)
# 输出: folder%2Ffile.txt
```

---

### Q2: ETag 格式错误

**问题：** 完成分片上传时提示 ETag 格式错误

**解决方案：**

ETag 必须包含双引号：

```json
// ✗ 错误
{
  "1": "5eb63bbbe01eeed093cb22bb8f5acdc3"
}

// ✓ 正确
{
  "1": "\"5eb63bbbe01eeed093cb22bb8f5acdc3\""
}
```

**从响应头提取 ETag：**
```bash
# 方法 1: 使用 grep 和 awk
ETAG=$(curl -s -i -X PUT "$URL" --data-binary @file \
  | grep -i "ETag:" | awk '{print $2}' | tr -d '\r\n')

# 方法 2: 使用 sed
ETAG=$(curl -s -i -X PUT "$URL" --data-binary @file \
  | sed -n 's/.*ETag: *\(.*\).*/\1/p' | tr -d '\r\n')
```

---

### Q3: 签名过期

**问题：** 使用预签名 URL 时提示签名过期

**解决方案：**

1. **增加过期时间：**
```bash
# 设置 2 小时过期
curl "http://localhost:8080/api/v1/upload/xxx/file.txt?expires=7200"
```

2. **检查时间同步：**
```bash
# Linux/Mac
ntpdate -u time.nist.gov

# Windows
w32tm /resync
```

3. **重新获取 URL：**
预签名 URL 过期后需要重新获取，无法续期。

---

### Q4: 分片上传中断

**问题：** 上传过程中网络中断，部分分片已上传

**解决方案：**

1. 使用相同的 `upload_id` 重新上传失败的分片
2. 已上传成功的分片无需重传（对象存储会保留）
3. 所有分片上传完成后再调用完成上传接口

**断点续传示例：**
```bash
# 假设分片 5, 8, 12 上传失败
FAILED_PARTS=(5 8 12)

for part in "${FAILED_PARTS[@]}"; do
  echo "重新上传分片 $part..."
  
  # 获取上传 URL
  PART_URL=$(get_part_url $part)
  
  # 重新上传
  curl -X PUT "$PART_URL" \
    -H "Content-Type: application/octet-stream" \
    --data-binary "@video.mp4.part$part"
done
```

---

### Q5: 文件大小限制

**问题：** 上传文件时提示文件过大

**解决方案：**

| 上传方式   | 建议场景          | 最大文件大小 |
|-----------|-------------------|-------------|
| PUT 上传  | < 100MB           | 5GB         |
| POST 上传 | < 100MB           | 5GB         |
| 分片上传   | > 100MB           | 50TB        |

**自动选择上传方式：**
```bash
FILE_SIZE=$(stat -c%s "$FILE")

if [ $FILE_SIZE -lt 104857600 ]; then
  # < 100MB: 使用单文件上传
  echo "使用单文件上传"
  use_simple_upload
else
  # >= 100MB: 使用分片上传
  echo "使用分片上传"
  use_multipart_upload
fi
```

---

### Q6: 性能优化

#### 并行上传分片

```bash
#!/bin/bash

# 并行上传分片（最多 5 个并发）
MAX_PARALLEL=5
COUNTER=0

for i in $(seq 1 $TOTAL_PARTS); do
  (
    upload_part $i
  ) &
  
  COUNTER=$((COUNTER + 1))
  
  # 控制并发数
  if [ $COUNTER -ge $MAX_PARALLEL ]; then
    wait -n  # 等待任意一个后台任务完成
    COUNTER=$((COUNTER - 1))
  fi
done

wait  # 等待所有后台任务完成
```

#### 使用内网传输

如果网关和对象存储在同一 VPC：

```bash
# 添加 internal_request=true 参数
curl "http://localhost:8080/api/v1/upload/xxx/file.txt?internal_request=true"
```

优势：
- 更快的传输速度
- 节省带宽成本
- 更稳定的网络连接

---

## 性能建议

### 1. 选择合适的上传方式

| 文件大小      | 推荐方式   | 原因                  |
|--------------|-----------|----------------------|
| < 10MB       | PUT       | 简单快速              |
| 10MB - 100MB | PUT       | 适中，易于实现         |
| 100MB - 1GB  | 分片上传   | 支持断点续传          |
| > 1GB        | 分片上传   | 必须使用，提高成功率   |

### 2. 优化分片大小

```bash
# 根据文件大小动态计算分片大小
calculate_part_size() {
  local file_size=$1
  local min_part_size=5242880      # 5MB
  local max_part_size=5368709120   # 5GB
  local max_parts=10000
  
  local part_size=$((file_size / max_parts))
  
  if [ $part_size -lt $min_part_size ]; then
    part_size=$min_part_size
  elif [ $part_size -gt $max_part_size ]; then
    part_size=$max_part_size
  fi
  
  echo $part_size
}
```

### 3. 并发控制

- 小文件（< 100MB）：无需并发
- 中等文件（100MB - 1GB）：5-10 个并发
- 大文件（> 1GB）：10-20 个并发

### 4. 网络优化

- 使用内网端点（如果可用）
- 选择同区域的存储服务
- 启用 HTTP/2（如果对象存储支持）

---

## 安全建议

### 1. 使用 HTTPS

生产环境必须使用 HTTPS：
```bash
GATEWAY_URL="https://gateway.example.com/api/v1"
```

### 2. 设置合理的过期时间

| 操作类型 | 建议过期时间 | 说明                |
|---------|-------------|---------------------|
| 上传     | 1-24 小时    | 根据文件大小调整     |
| 下载     | 5-60 分钟    | 临时访问             |
| 分享     | 1-7 天       | 长期分享             |
| 删除     | 5-15 分钟    | 立即操作             |

### 3. 访问控制

在网关层实施认证：
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "$GATEWAY_URL/upload/..."
```

### 4. 日志记录

记录所有操作：
```bash
LOG_FILE="upload_$(date +%Y%m%d_%H%M%S).log"

{
  echo "$(date): 开始上传 $FILE"
  upload_file "$FILE"
  echo "$(date): 上传完成"
} | tee -a "$LOG_FILE"
```

---

## 相关文档

- [API 文档](./API_DOCUMENTATION.md) - 完整的 API 接口定义
- [设计文档](./golang-oss-gateway-design.md) - 系统架构和技术设计
