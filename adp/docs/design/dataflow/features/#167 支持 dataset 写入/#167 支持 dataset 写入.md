# DataSet 写入节点及 Dataflow-Dataset-BKN CLI 优化

## 概述

本文档描述如何打通 Dataflow → Dataset → BKN 的完整流程，使用户能够通过 CLI 命令快速创建和关联这三类资源。

### 目标

1. 新增 `DatasetWriteDocs` 节点，支持 dataflow 向指定 dataset 写入文档
2. 扩展 kweaver CLI，支持通过模板创建 dataset、bkn、dataflow
3. 提供内置模板系统，简化用户操作

### 服务架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         kweaver CLI                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ templates   │  │create-dataset│ │ create-bkn  │   create     │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  ┌──────┐    │
│         │                │                │         │      │    │
│         ▼                ▼                ▼         │      │    │
│  ┌─────────────────────────────────────────────┐   │      │    │
│  │              Template Loader                 │   │      │    │
│  │  - 加载内置模板 / 文件路径                     │   │      │    │
│  │  - 解析 manifest.json                        │   │      │    │
│  │  - 参数合并与校验                             │   │      │    │
│  │  - 占位符替换                                 │   │      │    │
│  └─────────────────────────────────────────────┘   │      │    │
└─────────────────────────────────────────────────────┼──────┼────┘
                                                      │      │
                              ┌───────────────────────┼──────┼────┐
                              │                       │      │    │
                              ▼                       ▼      ▼    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                           Backend Services                           │
│                                                                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────┐  │
│  │ vega-backend    │  │ontology-manager │  │  dataflow-backend   │  │
│  │ /resources      │  │/knowledge-      │  │  /dataflows         │  │
│  │ (Dataset CRUD)  │  │ networks (BKN)  │  │  (Dataflow CRUD)    │  │
│  └────────┬────────┘  └────────┬────────┘  └──────────┬──────────┘  │
│           │                    │                      │             │
│           ▼                    ▼                      ▼             │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      OpenSearch                              │   │
│  │  - content_index (切片向量)                                   │   │
│  │  - content_element (文档元素)                                 │   │
│  │  - content_document (文件元信息)                              │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 一、DatasetWriteDocs 节点设计

### 1.1 节点定义

| 属性 | 值 |
|------|-----|
| 节点名称 | `DatasetWriteDocs` |
| 操作符 | `@dataset/write-docs` |
| 功能 | 向指定 dataset 写入文档数据 |

### 1.2 输入参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `dataset_id` | string | 是 | 目标 dataset ID |
| `documents` | array/object | 是 | 待写入的文档数据，支持数组或单个对象 |

### 1.3 文档结构

documents 参数支持的结构与 `OpenSearchBulkUpsert` 解析方式一致：

```json
[
  {
    "id": "doc-1",
    "field1": "value1",
    "field2": "value2"
  },
  {
    "id": "doc-2",
    "field1": "value3",
    "field2": "value4"
  }
]
```

### 1.4 节点配置示例

```json
{
  "id": "1001",
  "title": "写入文档到dataset",
  "operator": "@dataset/write-docs",
  "parameters": {
    "dataset_id": "{{__0.dataset_id}}",
    "documents": "{{__1.chunks}}"
  }
}
```

### 1.5 实现参考

参考 `opensearch.go` 中的 `OpenSearchBulkUpsert` 实现：

- 输入数据解析方式保持一致
- 调用 `/api/vega-backend/v1/resources/dataset/${dataset_id}/docs` 写入数据
- 支持异步写入和批量写入

---

## 二、CLI 命令设计

### 2.1 命令结构

```
kweaver dataflow
├── templates                    # 列出所有可用模板
├── create-dataset               # 创建 dataset
├── create-bkn                   # 创建 bkn
└── create                       # 创建 dataflow
```

### 2.2 命令详解

#### 2.2.1 templates - 列出模板

```bash
kweaver dataflow templates
```

输出示例：
```
Dataset Templates:
  - document          文档元信息数据集
  - document-content  文档切片及向量数据集
  - document-element  文档元素数据集

BKN Templates:
  - document          文档知识网络

Dataflow Templates:
  - unstructured      非结构化文档处理流程
```

`--json` 输出：
```json
{
  "dataset": [
    { "name": "document", "description": "文档元信息数据集" },
    { "name": "document-content", "description": "文档切片及向量数据集" },
    { "name": "document-element", "description": "文档元素数据集" }
  ],
  "bkn": [
    { "name": "document", "description": "文档知识网络" }
  ],
  "dataflow": [
    { "name": "unstructured", "description": "非结构化文档处理流程" }
  ]
}
```

#### 2.2.2 create-dataset - 创建 dataset

```bash
kweaver dataflow create-dataset --template <template> --set "key=value" [--json]
```

| 参数 | 说明 |
|------|------|
| `--template` | 模板名称（内置）或文件路径 |
| `--set` | 设置参数值，可多次使用 |
| `--json` | JSON 格式输出 |

示例：
```bash
# 使用内置模板
kweaver dataflow create-dataset --template document --set "name=my-docs"

# 使用文件路径
kweaver dataflow create-dataset --template ./my-template.json --set "name=my-docs" --json
```

输出：
```
dataset created: id=abc123
```

`--json` 输出：
```json
{
  "success": true,
  "id": "abc123",
  "name": "my-docs"
}
```

#### 2.2.3 create-bkn - 创建 bkn

```bash
kweaver dataflow create-bkn --template <template> --set "key=value" [--json]
```

示例：
```bash
kweaver dataflow create-bkn --template document \
  --set "name=my-bkn" \
  --set "document_content_id=dataset-id-123" \
  --set "document_id=dataset-id-456" \
  --set "document_element_id=dataset-id-789"
```

#### 2.2.4 create - 创建 dataflow

```bash
kweaver dataflow create --template <template> --set "key=value" [--json]
```

示例：
```bash
kweaver dataflow create --template unstructured \
  --set "name=my-flow" \
  --json
```

### 2.3 --set 语法规则

- 格式：`--set "key=value"`
- 支持多次使用：`--set "key1=value1" --set "key2=value2"`
- 仅支持顶层参数，不支持嵌套路径
- 值始终作为字符串处理，类型转换由模板 manifest 定义

### 2.4 输出格式

**默认模式：** 打印资源 ID

```
<type> created: id=xxx
```

**--json 模式：**

```json
{
  "success": true,
  "id": "xxx",
  "name": "xxx"
}
```

---

## 三、模板文件结构

### 3.1 目录结构

```
kweaver-sdk/packages/typescript/templates/
├── dataset/
│   ├── document/
│   │   ├── template.json      # 数据集定义模板
│   │   └── manifest.json      # 元数据与参数定义
│   ├── document-content/
│   │   ├── template.json
│   │   └── manifest.json
│   └── document-element/
│       ├── template.json
│       └── manifest.json
├── bkn/
│   └── document/
│       ├── template.json      # BKN 定义模板
│       └── manifest.json
└── dataflow/
    └── unstructured/
        ├── template.json      # Dataflow 定义模板
        └── manifest.json
```

### 3.2 template.json 示例

```json
{
  "catalog_id": "{{catalog_id}}",
  "name": "{{name}}",
  "category": "dataset",
  "status": "active",
  "description": "文档元信息数据集",
  "source_identifier": "dataflow_document",
  "schema_definition": [
    { "name": "id", "type": "keyword" },
    { "name": "document_id", "type": "keyword" },
    { "name": "doc_name", "type": "text" }
  ]
}
```

### 3.3 manifest.json 结构

```json
{
  "name": "document",
  "type": "dataset",
  "description": "文档元信息数据集",
  "arguments": [
    {
      "name": "name",
      "required": true,
      "description": "数据集名称",
      "type": "string",
      "path": "$.name"
    },
    {
      "name": "catalog_id",
      "required": false,
      "default": "adp_bkn_catalog",
      "description": "所属目录ID",
      "type": "string",
      "path": "$.catalog_id"
    }
  ]
}
```

### 3.4 arguments 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 参数名称，对应 --set 的 key |
| `required` | boolean | 是 | 是否必填 |
| `description` | string | 是 | 参数描述 |
| `type` | string | 是 | 参数类型：string, integer, boolean, array |
| `default` | any | 否 | 默认值（仅 required=false 时有效） |
| `path` | string | 是 | JSONPath，定位模板中的替换位置 |

### 3.5 模板加载规则

```
1. 判断 --template 参数类型：
   - 不含路径分隔符 → 内置模板
   - 含路径分隔符 → 文件路径

2. 内置模板定位：
   templates/<type>/<name>/template.json
   templates/<type>/<name>/manifest.json

3. 文件路径：
   直接加载指定文件
   manifest 文件在同目录下查找 manifest.json
```

---

## 四、执行流程

### 4.1 总体流程图

```
┌────────────────────────────────────────────────────────────────────────┐
│                          CLI 执行流程                                   │
└────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │   1. 解析 --template 参数      │
                    │   - 内置模板 or 文件路径        │
                    └───────────────┬───────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │   2. 加载模板文件              │
                    │   - template.json             │
                    │   - manifest.json (可选)       │
                    └───────────────┬───────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │   3. 解析 arguments           │
                    │   - 用户参数 (--set)          │
                    │   - 默认值 (manifest)         │
                    └───────────────┬───────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │   4. 校验必填参数              │
                    │   缺失 → 报错退出              │
                    └───────────────┬───────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │   5. 占位符替换                │
                    │   按 path 定位并替换           │
                    └───────────────┬───────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │   6. 调用后端 API              │
                    │   - create-dataset → vega     │
                    │   - create-bkn → ontology     │
                    │   - create → dataflow         │
                    └───────────────┬───────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │   7. 输出结果                  │
                    │   - 默认: 打印 id              │
                    │   - --json: 完整 JSON          │
                    └───────────────────────────────┘
```

### 4.2 create-dataset 执行流程

```
1. 解析 --template 参数，定位模板
   - 内置模板：templates/dataset/<name>/template.json
   - 文件路径：直接加载

2. 加载 manifest.json（如果存在），解析 arguments

3. 合并参数：
   - 从 --set 获取用户参数
   - 从 manifest.default 获取默认值
   - 检查必填参数是否全部提供

4. 执行占位符替换：
   - 遍历 arguments
   - 按 path（JSONPath）定位模板字段
   - 替换 {{argument_name}} 为实际值

5. 调用 API 创建 dataset：
   POST /api/vega-backend/v1/resources
   Body: 替换后的 template.json

6. 输出结果：
   - 默认：打印 dataset id
   - --json：输出完整响应 JSON
```

### 4.3 create-bkn 执行流程

```
1. 解析模板，加载 manifest

2. 合并参数，检查必填项

3. 执行占位符替换

4. 调用 API 创建 BKN：
   POST /api/ontology-manager/v1/knowledge-networks?validate_dependency=false
   Headers: { "x-business-domain": "bd_public" }
   Body: 替换后的 template.json

5. 输出结果
```

### 4.4 create (dataflow) 执行流程

```
1. 解析模板，加载 manifest

2. 合并参数，检查必填项

3. 执行占位符替换

4. 调用 API 创建 dataflow：
   POST /api/dataflow-backend/v1/dataflows
   Body: 替换后的 template.json

5. 输出结果
```

### 4.5 完整使用流程

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     用户完整操作流程                                     │
└─────────────────────────────────────────────────────────────────────────┘

Step 1: 创建 3 个 Dataset
┌─────────────────────────────────────────────────────────────────────────┐
│ $ kweaver dataflow create-dataset --template document \                 │
│     --set "name=my-document" --json                                     │
│ > {"success": true, "id": "ds-document-001"}                            │
│                                                                         │
│ $ kweaver dataflow create-dataset --template document-content \         │
│     --set "name=my-document-content" --json                             │
│ > {"success": true, "id": "ds-content-002"}                             │
│                                                                         │
│ $ kweaver dataflow create-dataset --template document-element \         │
│     --set "name=my-document-element" --json                             │
│ > {"success": true, "id": "ds-element-003"}                             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
Step 2: 创建 BKN（关联 Dataset）
┌─────────────────────────────────────────────────────────────────────────┐
│ $ kweaver dataflow create-bkn --template document \                     │
│     --set "name=my-bkn" \                                               │
│     --set "document_content_id=ds-content-002" \                        │
│     --set "document_id=ds-document-001" \                               │
│     --set "document_element_id=ds-element-003" \                        │
│     --json                                                              │
│ > {"success": true, "id": "bkn-001"}                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
Step 3: 创建 Dataflow（关联 Dataset）
┌─────────────────────────────────────────────────────────────────────────┐
│ $ kweaver dataflow create --template unstructured \                     │
│     --set "name=my-flow" \                                              │
│     --set "content_dataset_id=ds-content-002" \                         │
│     --set "document_dataset_id=ds-document-001" \                       │
│     --set "element_dataset_id=ds-element-003" \                         │
│     --json                                                              │
│ > {"success": true, "id": "flow-001"}                                   │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 五、错误处理

### 5.1 错误场景与处理

| 场景 | 错误码 | 处理方式 |
|------|--------|----------|
| 模板不存在 | `TEMPLATE_NOT_FOUND` | 报错退出，提示模板路径 |
| manifest.json 缺失 | - | 视为无参数模板，直接使用 template.json |
| 必填参数缺失 | `MISSING_ARGUMENT` | 报错退出，列出缺失参数 |
| 参数类型不匹配 | `INVALID_ARGUMENT_TYPE` | 报错退出，提示期望类型 |
| API 调用失败 | `API_ERROR` | 输出错误信息，退出码非零 |
| JSONPath 定位失败 | `PATH_NOT_FOUND` | 报错退出，提示无效路径 |

### 5.2 错误输出格式

**普通模式：**
```
Error: Missing required argument: name
Usage: kweaver dataflow create-dataset --template document --set "name=xxx"

Run 'kweaver dataflow create-dataset --help' for more information.
```

**--json 模式：**
```json
{
  "success": false,
  "error": {
    "code": "MISSING_ARGUMENT",
    "message": "Missing required argument: name",
    "details": {
      "missing": ["name"]
    }
  }
}
```

---

## 六、代码实现位置

### 6.1 kweaver-core（节点实现）

```
kweaver-core/adp/dataflow/flow-automation/pkg/actions/
└── dataset.go        # DatasetWriteDocs 节点实现（新增）
```

参考 `opensearch.go` 中的 `OpenSearchBulkUpsert` 实现。

### 6.2 kweaver-sdk（CLI + 模板）

```
kweaver-sdk/packages/typescript/
├── src/
│   └── cli/
│       └── dataflow/
│           ├── index.ts           # dataflow 子命令入口
│           ├── templates.ts       # templates 命令实现
│           ├── create-dataset.ts  # create-dataset 命令
│           ├── create-bkn.ts      # create-bkn 命令
│           ├── create.ts          # create 命令
│           └── template-loader.ts # 模板加载与占位符替换
└── templates/
    ├── dataset/
    │   ├── document/
    │   │   ├── template.json
    │   │   └── manifest.json
    │   ├── document-content/
    │   │   ├── template.json
    │   │   └── manifest.json
    │   └── document-element/
    │       ├── template.json
    │       └── manifest.json
    ├── bkn/
    │   └── document/
    │       ├── template.json
    │       └── manifest.json
    └── dataflow/
        └── unstructured/
            ├── template.json
            └── manifest.json
```

---

## 七、设计总结

### 7.1 核心变更

| 模块 | 变更内容 |
|------|----------|
| kweaver-core | 新增 `DatasetWriteDocs` 节点，支持向指定 dataset 写入文档 |
| kweaver-sdk CLI | 新增 `dataflow` 子命令，包含 `templates`、`create-dataset`、`create-bkn`、`create` |
| kweaver-sdk templates | 新增内置模板目录，包含 dataset/bkn/dataflow 模板及 manifest |

### 7.2 API 端点汇总

| 操作 | 端点 | 方法 |
|------|------|------|
| 创建 Dataset | `/api/vega-backend/v1/resources` | POST |
| 写入 Dataset 文档 | `/api/vega-backend/v1/resources/dataset/${id}/docs` | POST |
| 创建 BKN | `/api/ontology-manager/v1/knowledge-networks?validate_dependency=false` | POST |
| 创建 Dataflow | `/api/dataflow-backend/v1/dataflows` | POST |

### 7.3 依赖关系

```
┌─────────────┐
│  Dataset x3 │ (document, document-content, document-element)
└──────┬──────┘
       │
       ├──────────────────────┐
       │                      │
       ▼                      ▼
┌─────────────┐      ┌─────────────┐
│     BKN     │      │  Dataflow   │
│ (关联Dataset)│      │ (关联Dataset)│
└─────────────┘      └─────────────┘
```
