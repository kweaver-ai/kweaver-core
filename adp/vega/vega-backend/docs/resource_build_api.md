# 资源构建 API 文档

本文档描述了 VEGA Manager 中资源构建相关的 API 接口，包括构建任务的创建、查询、停止等操作。

## 1. 构建任务类型

VEGA Manager 支持以下类型的构建任务：

| 任务类型 | 描述 |
|---------|------|
| `batch` | 批处理构建任务，用于处理资源的初始构建 |
| `streaming` | 流式构建任务，用于处理实时数据流 |

## 2. 构建任务状态

构建任务有以下状态：

| 状态 | 描述 |
|------|------|
| `init` | 初始状态 |
| `running` | 任务正在执行 |
| `completed` | 任务执行完成 |
| `failed` | 任务执行失败 |
| `stopping` | 任务正在停止 |
| `stopped` | 任务已停止 |
| `failed` | 任务执行失败 |

## 3. API 接口

### 3.1 创建构建任务

**请求路径**：`POST /api/vega-backend/v1/resources/buildtask/{resource_id}`

**请求参数**：

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `resource_id` | string | 是 | 资源 ID |
| `task_type` | string | 是 | 任务类型，可选值：`batch`、`streaming` |
| `embedding_fields` | string | 否 | 嵌入字段，多个字段用逗号分隔 |
| `embedding_model` | string | 否 | 嵌入模型名称 |
| `model_dimensions` | int | 否 | 嵌入模型维度 |
| `build_key_fields` | string | 是 | 构建键字段，`batch`类型任务用于分批读取数据时排序，`streaming`类型任务用于源数据修改时的唯一标识 |

**请求示例**：

```json
{
  "task_type": "batch",
  "embedding_fields": "name,description",
  "embedding_model": "embedding",
  "model_dimensions": 1536
}
```

**响应示例**：

```json
{
  "task_id": "d7k489labmakp06gtus0"
}
```

### 3.2 查询构建任务信息

**请求路径**：`GET /api/vega-backend/v1/resources/buildtask/{resource_id}/{task_id}`

**请求参数**：

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `resource_id` | string | 是 | 资源 ID |
| `task_id` | string | 是 | 任务 ID |

**响应示例**：

```json
{
  "id": "d7k3rn5abmakp06gturg",
  "resource_id": "user-info-index",
  "status": "completed",
  "mode": "batch",
  "total_count": 4,
  "synced_count": 4,
  "vectorized_count": 0,
  "synced_mark": "{\"id\":\"ddd\"}",
  "creator": {
    "id": "",
    "type": ""
  },
  "create_time": 1776827868026,
  "updater": {
    "id": "",
    "type": ""
  },
  "update_time": 1776827919638,
  "embedding_fields": "hobby",
  "build_key_fields": "id",
  "embedding_model": "embedding",
  "model_dimensions": 1536
}
```

### 3.3 修改构建任务状态

**请求路径**：`PUT /api/vega-backend/v1/resources/buildtask/{resource_id}/{task_id}/status`

**请求参数**：

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `resource_id` | string | 是 | 资源 ID |
| `task_id` | string | 是 | 任务 ID |
| `status` | string | 是 | 任务状态，可选值：`running`、`stopped` |
| `execute_type` | string | 是 | 执行类型，可选值：`full`、`incremental`，默认值为 `incremental`，只对 `batch` 任务类型生效 |

**请求示例**：

```json
{
  "status": "running",
  "execute_type": "full"
}
```

### 3.4 查询资源的构建任务列表

**请求路径**：`GET /api/vega-backend/v1/resources/buildtask`

**响应示例**：

```json
{
  "total_count": 1,
  "entries": [
    {
      "id": "d7k489labmakp06gtus0",
      "resource_id": "user-info-index",
      "status": "completed",
      "mode": "batch",
      "total_count": 4,
      "synced_count": 4,
      "vectorized_count": 0,
      "synced_mark": "{\"id\":\"ddd\"}",
      "creator": {
        "id": "",
        "type": ""
      },
      "create_time": 1776829478576,
      "updater": {
        "id": "",
        "type": ""
      },
      "update_time": 1776829817964,
      "embedding_fields": "hobby",
      "build_key_fields": "id",
      "embedding_model": "embedding",
      "model_dimensions": 1536
    }
  ]
}
```

### 3.5 删除构建任务

**请求路径**：`DELETE /api/vega-backend/v1/resources/buildtask/{task_ids}`

**请求参数**：

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `task_ids` | string | 是 | 任务 ID 列表，多个 ID 用逗号分隔 |

## 5. 错误处理

| 错误码 | 描述 | 解决方案 |
|--------|------|----------|
| `VegaBackend.BuildTask.InternalError.CreateFailed` | 创建构建任务失败 | 请联系管理员 |
| `VegaBackend.BuildTask.InternalError.GetFailed` | 获取构建任务失败 | 请联系管理员 |
| `VegaBackend.BuildTask.Exist` | 构建任务已存在 | 请检查资源是否已存在构建任务 |
| `VegaBackend.BuildTask.Running` | 构建任务正在运行 | 请等待任务完成或停止后再操作 |
| `VegaBackend.BuildTask.InvalidStatus` | 无效的构建任务状态 | 请检查状态值 |
| `VegaBackend.BuildTask.InvalidExecuteType` | 无效的构建任务执行类型 | 请检查执行类型值 |
| `VegaBackend.BuildTask.InternalError.UpdateFailed` | 更新构建任务失败 | 请联系管理员 |
| `VegaBackend.BuildTask.InternalError.DeleteFailed` | 删除构建任务失败 | 请联系管理员 |
