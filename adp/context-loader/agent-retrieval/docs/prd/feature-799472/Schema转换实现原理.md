# Schema 转换实现原理说明

## 📋 概述

本文档说明如何将 **OpenAPI Schema** 转换为 **OpenAI Function Call Schema**，包括分层结构实现、$ref 解析和循环引用防护。

---

## 🎯 核心目标

将 OpenAPI 格式的工具定义转换为 LLM 可以理解的格式，同时：
1. **保持参数位置信息**（header/path/query/body）
2. **支持 $ref 引用解析**
3. **防止循环引用导致的无限递归**

---

## 📊 转换流程概览

```
┌─────────────────────────────────┐
│  输入：OpenAPI Schema            │
│  - parameters (path/query/header)│
│  - request_body (body)           │
│  - components.schemas ($ref)     │
└──────────────┬──────────────────┘
               │
               ↓
┌─────────────────────────────────┐
│  步骤1：初始化分层结构            │
│  创建 header/path/query/body     │
│  四个"容器"                      │
└──────────────┬──────────────────┘
               │
               ↓
┌─────────────────────────────────┐
│  步骤2：处理 parameters          │
│  遍历每个参数：                  │
│  1. 读取参数位置 (in: path/...) │
│  2. 解析 schema（可能含 $ref）  │
│  3. 放入对应位置的容器           │
└──────────────┬──────────────────┘
               │
               ↓
┌─────────────────────────────────┐
│  步骤3：处理 request_body        │
│  1. 解析 body schema（可能含 $ref）│
│  2. 展开 properties             │
│  3. 放入 body 容器              │
└──────────────┬──────────────────┘
               │
               ↓
┌─────────────────────────────────┐
│  步骤4：设置必填项               │
│  为每个位置设置 required 数组    │
└──────────────┬──────────────────┘
               │
               ↓
┌─────────────────────────────────┐
│  输出：OpenAI Function Call Schema│
│  {                               │
│    "type": "object",             │
│    "properties": {               │
│      "header": {...},            │
│      "path": {...},              │
│      "query": {...},             │
│      "body": {...}               │
│    }                              │
│  }                                │
└─────────────────────────────────┘
```

---

## 🏗️ 一、分层结构实现原理

### 1.1 为什么需要分层？

**问题：** OpenAPI 中参数有位置概念（path/query/header/body），但扁平化会丢失这个信息。

**解决方案：** 在顶层创建四个对象，分别对应四个位置。

### 1.2 实现方式

```go
// 初始化四个"容器"
properties := map[string]interface{}{
    "header": {
        "type": "object",
        "description": "HTTP Header 参数",
        "properties": {}  // 空的容器，等待填充
    },
    "path": {...},
    "query": {...},
    "body": {...}
}
```

### 1.3 参数分类过程

```
遍历 OpenAPI parameters 数组
    ↓
读取每个参数的 "in" 字段（位置信息）
    ↓
根据位置，放入对应的容器：
    - in: "path"  → 放入 properties["path"]
    - in: "query" → 放入 properties["query"]
    - in: "header" → 放入 properties["header"]
```

**示例：**

```json
// OpenAPI 输入
{
  "parameters": [
    {"name": "kn_id", "in": "path", "schema": {"type": "string"}},
    {"name": "limit", "in": "query", "schema": {"type": "integer"}}
  ]
}

// 转换后输出
{
  "type": "object",
  "properties": {
    "path": {
      "type": "object",
      "properties": {
        "kn_id": {"type": "string"}
      }
    },
    "query": {
      "type": "object",
      "properties": {
        "limit": {"type": "integer"}
      }
    }
  }
}
```

---

## 🔗 二、$ref 引用解析原理

### 2.1 什么是 $ref？

OpenAPI 中，为了复用定义，可以使用 `$ref` 引用其他地方的 schema：

```json
{
  "request_body": {
    "content": {
      "application/json": {
        "schema": {
          "$ref": "#/components/schemas/QueryCondition"  // 引用定义
        }
      }
    }
  },
  "components": {
    "schemas": {
      "QueryCondition": {  // 实际定义在这里
        "type": "object",
        "properties": {
          "field": {"type": "string"},
          "value": {"type": "string"}
        }
      }
    }
  }
}
```

### 2.2 解析流程

```
遇到 $ref: "#/components/schemas/QueryCondition"
    ↓
步骤1：解析路径
    - 提取 schema 名称：QueryCondition
    - 路径格式：#/components/schemas/{name}
    ↓
步骤2：查找定义
    - 在 apiSpec["components"]["schemas"] 中查找
    - 找到 QueryCondition 的定义
    ↓
步骤3：递归解析
    - 如果定义中还有 $ref，继续递归
    - 直到所有 $ref 都解析完成
    ↓
步骤4：展开 properties
    - 将解析后的 properties 合并到目标位置
```

### 2.3 代码实现逻辑

```go
resolveDollarRef(refPath, apiSpec) {
    // 1. 解析路径：提取 schema 名称
    schemaName = extractSchemaName(refPath)  // "QueryCondition"

    // 2. 查找定义
    schema = apiSpec["components"]["schemas"][schemaName]

    // 3. 递归解析（可能包含嵌套 $ref）
    resolvedSchema = resolveSchema(schema, apiSpec)

    return resolvedSchema
}
```

**示例：**

```json
// 输入：包含 $ref
{
  "schema": {
    "$ref": "#/components/schemas/QueryCondition"
  }
}

// 解析后：展开为实际定义
{
  "type": "object",
  "properties": {
    "field": {"type": "string"},
    "value": {"type": "string"}
  }
}
```

---

## 🔄 三、循环引用防护原理（深度限制剪枝策略）

### 3.1 什么是循环引用？

当两个或多个 schema 相互引用时，形成循环：

```json
{
  "components": {
    "schemas": {
      "Node": {
        "type": "object",
        "description": "节点对象",
        "properties": {
          "id": {"type": "string"},
          "children": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/Node"  // 循环引用自身
            }
          }
        }
      }
    }
  }
}
```

### 3.2 深度限制剪枝策略

**核心思想：**
1. 设定最大解析深度（如 2-3 层）
2. 在深度范围内正常展开 properties
3. 超过最大深度后剪枝：保留类型和原始描述，移除 properties
4. 不添加循环引用说明（节省 token）

### 3.3 防护机制流程

```
开始解析 $ref: "#/components/schemas/Node" (深度: 0)
    ↓
步骤1：检查深度
    currentDepth (0) < MaxSchemaDepth (2) ✅
    ↓
步骤2：标记为已访问
    visitedRefs["#/components/schemas/Node"] = true
    ↓
步骤3：解析 Node，发现 properties
    - id: string ✅
    - children: array，items 引用 Node
    ↓
开始解析 children.items (深度: 1)
    ↓
步骤1：检查深度
    currentDepth (1) < MaxSchemaDepth (2) ✅
    ↓
步骤2：解析 $ref: "#/components/schemas/Node" (深度: 1)
    ↓
步骤3：检查是否已访问
    visitedRefs["#/components/schemas/Node"] = true ⚠️
    ↓
检测到循环引用，但深度未达到上限
    ↓
继续展开（深度: 1 → 2）
    ↓
开始解析 properties (深度: 2)
    ↓
步骤1：检查深度
    currentDepth (2) >= MaxSchemaDepth (2) ❌
    ↓
达到最大深度，执行剪枝
    ↓
返回剪枝后的 schema：
    {
      "type": "object",
      "description": "节点对象"  // 保留原始描述
      // 不包含 properties（避免继续递归）
    }
```

### 3.4 实现细节

```go
const (
    MaxSchemaDepth = 2  // 最大解析深度，可配置
)

// 访问记录表
visitedRefs := make(map[string]bool)

// 递归解析 schema，支持深度控制
resolveSchema(schema, apiSpec, visitedRefs, currentDepth) {
    // 1. 检查是否达到最大深度
    if currentDepth >= MaxSchemaDepth {
        // 超过深度，执行剪枝
        return pruneSchema(schema)
    }

    // 2. 处理 $ref 引用
    if schema 有 $ref {
        refPath = schema["$ref"]

        // 检查循环引用
        if visitedRefs[refPath] == true {
            // 检测到循环，但深度未达到上限
            // 继续展开（在深度范围内）
            return resolveDollarRef(refPath, apiSpec, visitedRefs, currentDepth + 1)
        }

        // 标记为已访问
        visitedRefs[refPath] = true
        defer delete(visitedRefs, refPath)  // 递归返回时清理

        // 解析 $ref（深度 +1）
        return resolveDollarRef(refPath, apiSpec, visitedRefs, currentDepth + 1)
    }

    // 3. 处理 properties（深度不变，同一层级）
    if schema 有 properties {
        for each property {
            resolved = resolveSchema(property, apiSpec, visitedRefs, currentDepth)
            // 注意：解析 properties 时深度不变
        }
    }

    return schema
}

// 剪枝函数：保留类型和描述，移除 properties
pruneSchema(schema) {
    result = {
        "type": schema.type,  // 保留类型
    }

    // 保留原始 description（如果存在，不修改）
    if schema.description != nil {
        result["description"] = schema.description
    }

    // 如果是 array，保留 items 结构但不展开 properties
    if schema.type == "array" && schema.items != nil {
        result["items"] = pruneSchema(schema.items)  // 递归剪枝 items
    }

    // 不包含 properties（避免继续递归）
    return result
}
```

### 3.5 关键点

**深度控制：**
- ✅ 每次解析 `$ref` 时，深度 +1
- ✅ 解析 `properties` 中的属性时，深度不变（同一层级）
- ✅ 达到最大深度时，停止展开 properties

**循环引用检测：**
- ✅ 使用 `visitedRefs` map 记录已访问的 $ref
- ✅ 检测到循环时，检查当前深度
- ✅ 如果深度未达到上限，继续展开（允许有限深度的循环展开）

**剪枝策略：**
- ✅ 保留类型信息（`type: object/array`）
- ✅ 保留原始 `description`（不修改，不添加说明）
- ✅ 保留 `items` 结构（如果是 array）
- ❌ 移除 `properties`（避免继续递归）
- ❌ 不添加循环引用说明（节省 token）

### 3.6 示例对比

**输入（循环引用）：**
```json
{
  "Node": {
    "type": "object",
    "description": "节点对象",
    "properties": {
      "id": {"type": "string"},
      "children": {
        "type": "array",
        "description": "子节点列表",
        "items": {
          "$ref": "#/components/schemas/Node"
        }
      }
    }
  }
}
```

**输出（MaxSchemaDepth = 2）：**
```json
{
  "type": "object",
  "description": "节点对象",  // 保留原始描述
  "properties": {
    "id": {"type": "string"},
    "children": {
      "type": "array",
      "description": "子节点列表",  // 保留原始描述
      "items": {
        "type": "object",
        "description": "节点对象"  // 保留原始描述
        // 超过深度，不包含 properties（避免继续递归）
      }
    }
  }
}
```

**关键点：**
- ✅ 在深度范围内（0-2 层），正常展开 properties
- ✅ 超过深度后，保留类型和描述，移除 properties
- ✅ 不添加循环引用说明（节省 token）

---

## 📝 四、完整示例

### 4.1 示例1：普通 $ref 引用（无循环）

**输入（OpenAPI）：**
```json
{
  "parameters": [
    {
      "name": "kn_id",
      "in": "path",
      "required": true,
      "schema": {"type": "string"}
    },
    {
      "name": "limit",
      "in": "query",
      "schema": {"type": "integer"}
    }
  ],
  "request_body": {
    "content": {
      "application/json": {
        "schema": {
          "$ref": "#/components/schemas/QueryCondition"
        }
      }
    }
  },
  "components": {
    "schemas": {
      "QueryCondition": {
        "type": "object",
        "properties": {
          "field": {"type": "string"},
          "value": {"type": "string"}
        },
        "required": ["field"]
      }
    }
  }
}
```

**转换过程：**

**步骤1：初始化容器**
```json
{
  "header": {"type": "object", "properties": {}},
  "path": {"type": "object", "properties": {}},
  "query": {"type": "object", "properties": {}},
  "body": {"type": "object", "properties": {}}
}
```

**步骤2：处理 parameters**
- `kn_id` (in: path) → 放入 `path.properties`
- `limit` (in: query) → 放入 `query.properties`

**步骤3：处理 request_body**
- 解析 `$ref: "#/components/schemas/QueryCondition"`（深度: 0 → 1）
- 从 `components.schemas` 中找到定义
- 展开 properties 到 `body.properties`（深度: 1，未达到上限）

**步骤4：设置必填项**
- `path.required = ["kn_id"]`
- `body.required = ["field"]`

**输出（OpenAI Function Call）：**
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "object",
      "description": "URL Path 参数",
      "properties": {
        "kn_id": {"type": "string"}
      },
      "required": ["kn_id"]
    },
    "query": {
      "type": "object",
      "description": "URL Query 参数",
      "properties": {
        "limit": {"type": "integer"}
      }
    },
    "body": {
      "type": "object",
      "description": "Request Body 参数",
      "properties": {
        "field": {"type": "string"},
        "value": {"type": "string"}
      },
      "required": ["field"]
    }
  }
}
```

---

### 4.2 示例2：循环引用（深度限制剪枝）

**输入（OpenAPI，包含循环引用）：**
```json
{
  "request_body": {
    "content": {
      "application/json": {
        "schema": {
          "$ref": "#/components/schemas/Node"
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Node": {
        "type": "object",
        "description": "节点对象",
        "properties": {
          "id": {"type": "string"},
          "name": {"type": "string"},
          "children": {
            "type": "array",
            "description": "子节点列表",
            "items": {
              "$ref": "#/components/schemas/Node"  // 循环引用
            }
          }
        },
        "required": ["id"]
      }
    }
  }
}
```

**转换过程（MaxSchemaDepth = 2）：**

**步骤1：初始化容器**
```json
{
  "body": {"type": "object", "properties": {}}
}
```

**步骤2：处理 request_body**
- 解析 `$ref: "#/components/schemas/Node"`（深度: 0 → 1）
- 从 `components.schemas` 中找到 Node 定义
- 展开 properties：
  - `id: string` ✅
  - `name: string` ✅
  - `children: array`，解析 items（深度: 1 → 2）

**步骤3：处理 children.items（循环引用）**
- 解析 `$ref: "#/components/schemas/Node"`（深度: 2）
- 检测到循环引用（visitedRefs["Node"] = true）
- 检查深度：`currentDepth (2) >= MaxSchemaDepth (2)` ✅
- **执行剪枝**：保留类型和描述，移除 properties

**输出（OpenAI Function Call）：**
```json
{
  "type": "object",
  "properties": {
    "body": {
      "type": "object",
      "description": "Request Body 参数",
      "properties": {
        "id": {
          "type": "string"
        },
        "name": {
          "type": "string"
        },
        "children": {
          "type": "array",
          "description": "子节点列表",  // 保留原始描述
          "items": {
            "type": "object",
            "description": "节点对象"  // 保留原始描述
            // 超过深度，不包含 properties（避免继续递归）
          }
        }
      },
      "required": ["id"]
    }
  }
}
```

**关键点：**
- ✅ 在深度范围内（0-2 层），正常展开 properties
- ✅ 超过深度后，保留类型和原始描述，移除 properties
- ✅ 不添加循环引用说明（节省 token）

### 4.3 输出（OpenAI Function Call）

```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "object",
      "description": "URL Path 参数",
      "properties": {
        "kn_id": {
          "type": "string"
        }
      },
      "required": ["kn_id"]
    },
    "query": {
      "type": "object",
      "description": "URL Query 参数",
      "properties": {
        "limit": {
          "type": "integer"
        }
      }
    },
    "body": {
      "type": "object",
      "description": "Request Body 参数",
      "properties": {
        "field": {
          "type": "string"
        },
        "value": {
          "type": "string"
        }
      },
      "required": ["field"]
    }
  }
}
```

---

## 🔍 五、关键函数说明

### 5.1 `convertSchemaToFunctionCall`
**作用：** 主转换函数，协调整个转换流程

**流程：**
1. 初始化四个位置的容器
2. 处理 parameters（path/query/header）
3. 处理 request_body（body）
4. 设置各位置的 required
5. 清理空容器

### 5.2 `resolveSchema`
**作用：** 递归解析 schema，处理 $ref 和嵌套结构，支持深度控制和循环引用检测

**特点：**
- 支持 $ref 解析（递归深度控制）
- 支持循环引用检测（深度限制剪枝）
- 支持嵌套 properties 解析
- 支持深度控制（MaxSchemaDepth）

**参数：**
- `schema`: 要解析的 schema
- `apiSpec`: 完整的 OpenAPI 定义
- `visitedRefs`: 已访问的 $ref 记录表
- `currentDepth`: 当前递归深度

**流程：**
1. 检查是否达到最大深度 → 如果达到，执行剪枝
2. 检查是否有 $ref → 如果有，解析 $ref（深度 +1）
3. 检查循环引用 → 如果检测到，但深度未达到上限，继续展开
4. 处理 properties → 递归处理每个属性（深度不变）

### 5.3 `resolveDollarRef`
**作用：** 解析 $ref 路径，从 components.schemas 中查找定义

**支持的格式：**
- `#/components/schemas/SchemaName`

**参数：**
- `refPath`: $ref 路径（如 `#/components/schemas/Node`）
- `apiSpec`: 完整的 OpenAPI 定义
- `visitedRefs`: 已访问的 $ref 记录表
- `currentDepth`: 当前递归深度

**流程：**
1. 解析路径 → 提取 schema 名称
2. 查找定义 → 从 `components.schemas` 中查找
3. 递归解析 → 调用 `resolveSchema`（深度 +1）

### 5.4 `pruneSchema`
**作用：** 剪枝函数，当达到最大深度时，保留类型和描述，移除 properties

**输入：** 原始 schema（可能包含 properties）

**输出：** 剪枝后的 schema（不包含 properties）

**处理逻辑：**
1. 保留 `type` 字段
2. 保留原始 `description`（如果存在，不修改）
3. 如果是 `array`，递归剪枝 `items`
4. 移除 `properties`（避免继续递归）

**示例：**
```go
// 输入
{
  "type": "object",
  "description": "节点对象",
  "properties": {...}
}

// 输出
{
  "type": "object",
  "description": "节点对象"
  // 不包含 properties
}
```

### 5.5 `buildPropertyDefinition`
**作用：** 构建单个属性的定义（type, description, enum 等）

### 5.6 `mergeSchemaProperties`
**作用：** 将解析后的 schema properties 合并到目标位置

---

## ⚠️ 六、边界情况处理

### 6.1 空参数
如果某个位置没有参数，该位置会被移除（除非所有位置都为空，则保留 body 作为空对象）

### 6.2 $ref 解析失败
- 记录警告日志
- 返回占位符 schema，不中断流程

### 6.3 循环引用（深度限制剪枝）

**处理策略：**
1. 检测到循环引用时，检查当前深度
2. 如果深度未达到上限，继续展开（允许有限深度的循环展开）
3. 如果深度达到上限，执行剪枝：
   - 保留类型信息（`type: object/array`）
   - 保留原始 `description`（不修改）
   - 移除 `properties`（避免继续递归）
4. 不添加循环引用说明（节省 token）

**示例：**
```json
// 输入：循环引用
{
  "Node": {
    "properties": {
      "children": {
        "type": "array",
        "items": {"$ref": "#/components/schemas/Node"}
      }
    }
  }
}

// 输出（MaxSchemaDepth = 2）
{
  "type": "object",
  "description": "节点对象",  // 保留原始描述
  "properties": {
    "children": {
      "type": "array",
      "items": {
        "type": "object",
        "description": "节点对象"
        // 超过深度，不包含 properties
      }
    }
  }
}
```

### 6.4 缺少 type 字段
- 默认使用 `"string"` 类型
- 保证 schema 的完整性

---

## 📚 总结

**核心思想：**
1. **分层结构**：保持参数位置信息（header/path/query/body），便于 LLM 理解
2. **$ref 解析**：递归展开引用，获取完整定义（支持深度控制）
3. **循环防护**：深度限制剪枝策略，在深度范围内展开，超过深度后剪枝

**深度限制剪枝策略：**
- ✅ 设定最大解析深度（2-3 层，可配置）
- ✅ 在深度范围内正常展开 properties
- ✅ 超过深度后：保留类型和原始描述，移除 properties
- ✅ 不添加循环引用说明（节省 token，避免干扰 LLM）

**优势：**
- ✅ **语义清晰**：LLM 能明确知道参数应该放在哪里
- ✅ **信息完整**：在深度范围内保留完整结构
- ✅ **节省 token**：不添加冗长的循环引用说明
- ✅ **避免干扰**：不修改原始 description
- ✅ **类型明确**：LLM 能明确知道参数类型（object/array）
- ✅ **安全可靠**：防止循环引用导致的无限递归

**配置参数：**
```go
const (
    MaxSchemaDepth = 2  // 最大解析深度，建议 2-3 层
)
```

---

**文档版本：** v2.0
**最后更新：** 2025-12-23
**策略更新：** 采用深度限制剪枝策略处理循环引用

> 📖 **相关文档：**
> - [循环引用场景分析](./循环引用场景分析.md) - 详细梳理各种循环引用场景
> - [循环引用处理策略评估](./循环引用处理策略评估.md) - 策略评估和决策过程

