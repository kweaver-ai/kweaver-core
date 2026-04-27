# Debug 返回结构改进文档

## 1. 背景与问题

### 1.1 当前问题

在调试 `kn-logic-property-resolver` 接口时，存在以下问题：

1. **debug 信息仅在请求成功时返回**
   - 当前实现中，debug 信息只在 `Step 6` 构建响应时返回
   - 如果在 `Step 4` 生成 dynamic_params 时失败或缺参，不会返回 debug 信息
   - 导致调试时无法查看传递给 agent 的参数

2. **debug 信息不完整**
   - 缺少每个属性对应的类型（metric/operator）
   - 缺少传递给 agent 的请求参数
   - 无法追踪 agent 调用的完整过程

3. **调试困难**
   - 需要查看日志和代码才能了解传递给 agent 的参数
   - 无法直观地看到 agent 的输入输出
   - 参数校验失败时无法查看生成的参数

### 1.2 需求

1. **调整 debug 返回结构**
   - 除了返回 `dynamic_params`，还应返回：
     - 每个属性对应的类型
     - agent 的请求参数

2. **改进 debug 返回时机**
   - 如果设置 `"return_debug": true`，只要请求 agent 成功（即使返回 `_error`），都需要返回 debug 信息
   - 参数校验失败时也应返回 debug 信息

## 2. 设计方案

### 2.1 新的 Debug 数据结构

```go
// ResolveDebugInfo Debug 信息（改进版）
type ResolveDebugInfo struct {
    // Agent 生成的所有动态参数
    DynamicParams map[string]any `json:"dynamic_params,omitempty"`

    // Agent 调用信息（按 property 分组）
    AgentInfo map[string]*AgentInfo `json:"agent_info,omitempty"`

    // 服务器时间戳
    NowMs int64 `json:"now_ms,omitempty"`

    // 警告信息
    Warnings []string `json:"warnings,omitempty"`

    // 追踪 ID
    TraceID string `json:"trace_id,omitempty"`
}

// AgentInfo Agent 调用信息
type AgentInfo struct {
    // 属性类型
    PropertyType string `json:"property_type"` // "metric" 或 "operator"

    // Agent 请求参数（直接存储 Agent 请求结构）
    Request AgentRequestDebugInfo `json:"request,omitempty"`

    // Agent 响应信息
    Response *AgentResponseDebugInfo `json:"response,omitempty"`
}

// AgentRequestDebugInfo Agent 请求调试信息
// 直接存储 Agent 的请求结构：MetricDynamicParamsGeneratorReq 或 OperatorDynamicParamsGeneratorReq
type AgentRequestDebugInfo any

// AgentResponseDebugInfo Agent 响应调试信息
// 直接存储 Agent 的响应信息
type AgentResponseDebugInfo struct {
    // Agent 成功响应：动态参数
    DynamicParams map[string]any `json:"dynamic_params,omitempty"`

    // Agent 失败响应：错误信息（对应 _error 字段）
    Error string `json:"_error,omitempty"`
}
```

### 2.2 Debug 返回示例

#### 2.2.1 成功场景

```json
{
  "datas": [
    {
      "company_id": "company_000001",
      "approved_drug_count": 15,
      "business_health_score": 85.5
    }
  ],
  "debug": {
    "dynamic_params": {
      "approved_drug_count": {
        "instant": true,
        "start": 1766656234484,
        "end": 1766656234484
      },
      "business_health_score": {
        "instant": false,
        "start": 1766656234484,
        "end": 1766656234484,
        "step": "month"
      }
    },
    "agent_info": {
      "approved_drug_count": {
        "property_type": "metric",
        "request": {
          "logic_property": {
            "name": "approved_drug_count",
            "type": "metric",
            "parameters": [
              {
                "name": "instant",
                "type": "boolean",
                "value_from": "input",
                "if_system_generate": true
              },
              {
                "name": "start",
                "type": "integer",
                "value_from": "input",
                "if_system_generate": true
              },
              {
                "name": "end",
                "type": "integer",
                "value_from": "input",
                "if_system_generate": true
              }
            ]
          },
          "query": "查询企业已批准药品数量",
          "unique_identities": [
            {
              "company_id": "company_000001"
            }
          ],
          "additional_context": "",
          "now_ms": 1766656238129,
          "timezone": ""
        },
        "response": {
          "dynamic_params": {
            "instant": true,
            "start": 1766656234484,
            "end": 1766656234484
          }
        }
      },
      "business_health_score": {
        "property_type": "metric",
        "request": {
          "logic_property": {
            "name": "business_health_score",
            "type": "metric",
            "parameters": [
              {
                "name": "instant",
                "type": "boolean",
                "value_from": "input",
                "if_system_generate": true
              },
              {
                "name": "start",
                "type": "integer",
                "value_from": "input",
                "if_system_generate": true
              },
              {
                "name": "end",
                "type": "integer",
                "value_from": "input",
                "if_system_generate": true
              },
              {
                "name": "step",
                "type": "string",
                "value_from": "input",
                "if_system_generate": true
              }
            ]
          },
          "query": "查询企业健康度",
          "unique_identities": [
            {
              "company_id": "company_000001"
            }
          ],
          "additional_context": "",
          "now_ms": 1766656238129,
          "timezone": ""
        },
        "response": {
          "dynamic_params": {
            "instant": false,
            "start": 1766656234484,
            "end": 1766656234484,
            "step": "month"
          }
        }
      }
    },
    "now_ms": 1766656238129,
    "trace_id": "trace-123456"
  }
}
```

#### 2.2.2 缺参场景（Agent返回_error）

##### 2.2.2.1 开启 debug 时的返回（推荐）

**特殊处理**：当 `return_debug=true` 时，缺参场景不抛出错误，而是返回正常响应，错误信息放在 `debug` 字段中。

```json
{
  "datas": [],
  "debug": {
    "dynamic_params": {
      "approved_drug_count": {
        "instant": true,
        "start": 1766656234484
      }
    },
    "agent_info": {
      "approved_drug_count": {
        "property_type": "metric",
        "request": {
          "logic_property": {
            "name": "approved_drug_count",
            "type": "metric",
            "parameters": [
              {
                "name": "instant",
                "type": "boolean",
                "value_from": "input",
                "if_system_generate": true
              },
              {
                "name": "start",
                "type": "integer",
                "value_from": "input",
                "if_system_generate": true
              },
              {
                "name": "end",
                "type": "integer",
                "value_from": "input",
                "if_system_generate": true
              }
            ]
          },
          "query": "查询企业已批准药品数量",
          "unique_identities": [
            {
              "company_id": "company_000001"
            }
          ],
          "additional_context": "",
          "now_ms": 1766656238129,
          "timezone": ""
        },
        "response": {
          "_error": "missing approved_drug_count: end | ask: 请提供查询结束时间"
        }
      }
    },
    "now_ms": 1766656238129,
    "trace_id": "trace-123456"
  }
}
```

**说明**：
- 返回正常的响应结构（包含 `datas` 和 `debug` 字段）
- 错误信息放在 `debug.agent_info.{property_name}.response._error` 字段中（直接对应 Agent 返回的 `_error` 字段）
- `request` 字段直接存储 Agent 请求结构（`MetricDynamicParamsGeneratorReq` 或 `OperatorDynamicParamsGeneratorReq`）
- `response` 字段直接存储 Agent 响应信息（成功时包含 `dynamic_params`，失败时包含 `_error`）
- `datas` 字段为空数组（因为没有成功的数据）
- 这样可以避免错误信息被统一框架包装到 `details` 字段中，便于调试

##### 2.2.2.2 未开启 debug 时的返回（现有行为）

```json
{
  "error_code": "MISSING_INPUT_PARAMS",
  "message": "dynamic_params 缺少必需的 input 参数",
  "missing": [
    {
      "property": "approved_drug_count",
      "params": [
        {
          "name": "end",
          "type": "integer",
          "hint": "请提供查询结束时间"
        }
      ]
    }
  ],
  "trace_id": "trace-123456"
}
```

**说明**：
- 保持现有的错误返回行为
- 错误信息会被统一框架包装到 `details` 字段中
- 向后兼容，不影响现有调用方

```json
{
  "error_code": "MISSING_INPUT_PARAMS",
  "message": "dynamic_params 缺少必需的 input 参数",
  "missing": [
    {
      "property": "approved_drug_count",
      "params": [
        {
          "name": "_error",
          "type": "error",
          "hint": "validate params failed: metric property approved_drug_count: instant=true cannot have 'step' field"
        }
      ]
    }
  ],
  "trace_id": "trace-123456",
  "debug": {
    "dynamic_params": {
      "approved_drug_count": {
        "instant": true,
        "start": 1766656234484,
        "end": 1766656234484,
        "step": "day"
      }
    },
    "agent_info": {
      "approved_drug_count": {
        "property_type": "metric",
        "request": {
          "logic_property": {
            "name": "approved_drug_count",
            "type": "metric",
            "parameters": [
              {
                "name": "instant",
                "type": "boolean",
                "value_from": "input",
                "if_system_generate": true
              },
              {
                "name": "start",
                "type": "integer",
                "value_from": "input",
                "if_system_generate": true
              },
              {
                "name": "end",
                "type": "integer",
                "value_from": "input",
                "if_system_generate": true
              }
            ]
          },
          "query": "查询企业已批准药品数量",
          "unique_identities": [
            {
              "company_id": "company_000001"
            }
          ],
          "additional_context": "",
          "now_ms": 1766656238129,
          "timezone": ""
        },
        "response": {
          "dynamic_params": {
            "instant": true,
            "start": 1766656234484,
            "end": 1766656234484,
            "step": "day"
          }
        }
      }
    },
    "now_ms": 1766656238129,
    "trace_id": "trace-123456"
  }
}
```

```json
{
  "error_code": "MISSING_INPUT_PARAMS",
  "message": "dynamic_params 缺少必需的 input 参数",
  "missing": [
    {
      "property": "approved_drug_count",
      "params": [
        {
          "name": "end",
          "type": "integer",
          "hint": "请提供查询结束时间"
        }
      ]
    }
  ],
  "trace_id": "trace-123456",
  "debug": {
    "dynamic_params": {
      "approved_drug_count": {
        "instant": true,
        "start": 1766656234484
      }
    },
    "agent_info": {
      "approved_drug_count": {
        "property_type": "metric",
        "request": {
          "logic_property": {
            "name": "approved_drug_count",
            "type": "metric",
            "parameters": [
              {
                "name": "instant",
                "type": "boolean",
                "value_from": "input",
                "if_system_generate": true
              },
              {
                "name": "start",
                "type": "integer",
                "value_from": "input",
                "if_system_generate": true
              },
              {
                "name": "end",
                "type": "integer",
                "value_from": "input",
                "if_system_generate": true
              }
            ]
          },
          "query": "查询企业已批准药品数量",
          "unique_identities": [
            {
              "company_id": "company_000001"
            }
          ],
          "additional_context": "",
          "now_ms": 1766656238129,
          "timezone": ""
        },
        "response": {
          "_error": "缺少必需参数：end"
        }
      }
    },
    "now_ms": 1766656238129,
    "trace_id": "trace-123456"
  }
}
```

### 3. 实现方案

### 3.1 修改数据结构

#### 3.1.1 修改 `ResolveDebugInfo` 结构

位置：`server/interfaces/kn_logic_property_resolver.go`

```go
// ResolveDebugInfo Debug 信息（改进版）
type ResolveDebugInfo struct {
    // Agent 生成的所有动态参数
    DynamicParams map[string]any `json:"dynamic_params,omitempty"`

    // Agent 信息（按属性分组，包含类型、请求、响应）
    AgentInfo map[string]*AgentInfo `json:"agent_info,omitempty"`

    // 服务器时间戳
    NowMs int64 `json:"now_ms,omitempty"`

    // 警告信息
    Warnings []string `json:"warnings,omitempty"`

    // 追踪 ID
    TraceID string `json:"trace_id,omitempty"`
}

// AgentInfo Agent调试信息（按属性分组）
type AgentInfo struct {
    // 属性类型
    PropertyType string `json:"property_type"`

    // Agent 请求参数（直接存储 Agent 请求结构）
    Request AgentRequestDebugInfo `json:"request,omitempty"`

    // Agent 响应信息
    Response *AgentResponseDebugInfo `json:"response,omitempty"`
}

// AgentRequestDebugInfo Agent 请求调试信息
// 直接存储 Agent 的请求结构：MetricDynamicParamsGeneratorReq 或 OperatorDynamicParamsGeneratorReq
type AgentRequestDebugInfo any

// AgentResponseDebugInfo Agent 响应调试信息
// 直接存储 Agent 的响应信息
type AgentResponseDebugInfo struct {
    // Agent 成功响应：动态参数
    DynamicParams map[string]any `json:"dynamic_params,omitempty"`

    // Agent 失败响应：错误信息（对应 _error 字段）
    Error string `json:"_error,omitempty"`
}

// MissingParamsError 缺参错误响应（改进版）
type MissingParamsError struct {
    ErrorCode string `json:"error_code"`
    Message   string `json:"message"`

    // 直接返回Agent生成的原始错误消息
    ErrorMsg string `json:"error_msg,omitempty"`

    // Debug信息序列化为字符串，减少token消耗
    DebugStr string `json:"debug_str,omitempty"`

    TraceID string `json:"trace_id"`
}
```

### 3.2 修改业务逻辑

#### 3.2.1 修改 `ResolveLogicProperties` 方法

位置：`server/logics/knlogicpropertyresolver/index.go`

**关键改动**：

1. **在开始处理时初始化 debug 信息收集器**
   ```go
   // 初始化 debug 信息收集器
   var debugCollector *DebugCollector
   if req.Options.ReturnDebug {
       debugCollector = NewDebugCollector()
       debugCollector.SetTraceID(s.getTraceID(ctx))
   }
   ```

2. **在 `generateDynamicParams` 中收集 debug 信息**
   ```go
   // Step 4: 生成 dynamic_params
   dynamicParams, missingParams, err := s.generateDynamicParams(ctx, req, logicPropertiesDef, debugCollector)
   ```

3. **特殊处理：开启 debug 时的缺参场景**
   ```go
   // 如果有缺参，根据是否开启 debug 决定处理方式
   if len(missingParams) > 0 {
       s.logger.WithContext(ctx).Warnf("[Step 2] ⚠️ 存在缺参: %d 个属性", len(missingParams))

       // 特殊处理：如果开启了 debug，返回正常响应，错误信息放在 debug 中
       if req.Options.ReturnDebug {
           s.logger.WithContext(ctx).Infof("[Step 2] 🔍 Debug模式：缺参场景返回正常响应，错误信息放在 debug 中")

           // 构建正常响应，datas 为空数组
           debugInfo := debugCollector.BuildDebugInfo()
           return &interfaces.ResolveLogicPropertiesResponse{
               Datas: []map[string]any{},  // 空数组，因为没有成功的数据
               Debug: debugInfo,
           }, nil
       }

       // 未开启 debug：保持现有行为，抛出错误
       missingError := s.buildMissingParamsError(ctx, missingParams)
       return nil, missingError
   }
   ```

**说明**：
- **开启 debug 时**：缺参场景不抛出错误，返回正常响应，`datas` 为空数组，错误信息放在 `debug` 字段中
- **未开启 debug 时**：保持现有行为，抛出 `MissingParamsError` 错误
- 这样可以避免错误信息被统一框架包装到 `details` 字段中，便于调试

#### 3.2.2 修改 `generateDynamicParams` 方法

**关键改动**：

1. **添加 `debugCollector` 参数**
   ```go
   func (s *knLogicPropertyResolverService) generateDynamicParams(
       ctx context.Context,
       req *interfaces.ResolveLogicPropertiesRequest,
       logicPropertiesDef map[string]*interfaces.LogicPropertyDef,
       debugCollector *DebugCollector,
   ) (dynamicParams map[string]interface{}, missingParams []interfaces.MissingPropertyParams, err error)
   ```

2. **在调用 `generateSinglePropertyParams` 时传递 debugCollector**
   ```go
   // 生成单个 property 的 dynamic_params
   params, missing, err := s.generateSinglePropertyParams(ctx, req, t.Name, t.Property, debugCollector)
   ```

3. **收集 property 类型信息**
   ```go
   // 收集 property 类型信息
   if debugCollector != nil {
       debugCollector.AddPropertyType(t.Name, t.Property.Type)
   }
   ```

#### 3.2.3 修改 `generateSinglePropertyParams` 方法

**关键改动**：

1. **添加 `debugCollector` 参数**
   ```go
   func (s *knLogicPropertyResolverService) generateSinglePropertyParams(
       ctx context.Context,
       req *interfaces.ResolveLogicPropertiesRequest,
       propertyName string,
       property *interfaces.LogicPropertyDef,
       debugCollector *DebugCollector,
   ) (dynamicParams map[string]interface{}, missingParams *interfaces.MissingPropertyParams, err error)
   ```

2. **在调用 Agent 前记录请求信息**
   ```go
   // 记录 Agent 请求信息
   if debugCollector != nil {
       debugCollector.RecordAgentRequest(propertyName, property, req)
   }
   ```

3. **在调用 Agent 后记录响应信息**
   ```go
   // 记录 Agent 响应信息
   if debugCollector != nil {
       if missingParams != nil {
           debugCollector.RecordAgentResponseMissingParams(propertyName, missingParams)
       } else if dynamicParams != nil {
           debugCollector.RecordAgentResponseSuccess(propertyName, dynamicParams)
       }
   }
   ```

4. **在调用 Agent 前记录请求信息**
   ```go
   // 记录 Agent 请求信息（在 generateMetricParams 或 generateOperatorParams 中）
   if debugCollector != nil {
       debugCollector.RecordMetricAgentRequest(propertyName, agentReq)
       // 或
       debugCollector.RecordOperatorAgentRequest(propertyName, agentReq)
   }
   ```

### 3.3 新增 DebugCollector 工具类

创建新文件：`server/logics/knlogicpropertyresolver/debug_collector.go`

```go
package knlogicpropertyresolver

import (
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/interfaces"
)

// DebugCollector Debug 信息收集器
type DebugCollector struct {
    propertyTypes  map[string]string
    agentRequests  map[string]interfaces.AgentRequestDebugInfo
    agentResponses map[string]*interfaces.AgentResponseDebugInfo
    dynamicParams  map[string]any
    nowMs          int64
    traceID        string
    warnings       []string
}

// NewDebugCollector 创建 Debug 信息收集器
func NewDebugCollector() *DebugCollector {
    return &DebugCollector{
        propertyTypes:  make(map[string]string),
        agentRequests:  make(map[string]interfaces.AgentRequestDebugInfo),
        agentResponses: make(map[string]*interfaces.AgentResponseDebugInfo),
        dynamicParams:  make(map[string]any),
        warnings:       make([]string, 0),
    }
}

// AddPropertyType 添加属性类型
func (dc *DebugCollector) AddPropertyType(propertyName string, propertyType string) {
    dc.propertyTypes[propertyName] = propertyType
}

// RecordMetricAgentRequest 记录 Metric Agent 请求（直接存储 Agent 请求结构）
func (dc *DebugCollector) RecordMetricAgentRequest(
    propertyName string,
    agentReq *interfaces.MetricDynamicParamsGeneratorReq,
) {
    // 直接存储 Agent 请求结构
    dc.agentRequests[propertyName] = agentReq
}

// RecordOperatorAgentRequest 记录 Operator Agent 请求（直接存储 Agent 请求结构）
func (dc *DebugCollector) RecordOperatorAgentRequest(
    propertyName string,
    agentReq *interfaces.OperatorDynamicParamsGeneratorReq,
) {
    // 直接存储 Agent 请求结构
    dc.agentRequests[propertyName] = agentReq
}

// RecordAgentResponseSuccess 记录 Agent 成功响应（直接存储 Agent 响应）
func (dc *DebugCollector) RecordAgentResponseSuccess(propertyName string, dynamicParams map[string]any) {
    // 直接存储 Agent 响应：成功时返回 dynamicParams
    dc.agentResponses[propertyName] = &interfaces.AgentResponseDebugInfo{
        DynamicParams: dynamicParams,
    }

    // 同时收集到 dynamic_params
    dc.dynamicParams[propertyName] = dynamicParams
}

// RecordAgentResponseMissingParams 记录 Agent 缺参响应（直接存储 Agent 响应）
func (dc *DebugCollector) RecordAgentResponseMissingParams(
    propertyName string,
    missingParams *interfaces.MissingPropertyParams,
) {
    errorMsg := ""
    if missingParams != nil {
        errorMsg = missingParams.ErrorMsg
    }

    // 直接存储 Agent 响应：失败时返回 _error 字段
    dc.agentResponses[propertyName] = &interfaces.AgentResponseDebugInfo{
        Error: errorMsg,
    }
}

// RecordAgentResponseError 记录 Agent 错误响应（直接存储 Agent 响应）
func (dc *DebugCollector) RecordAgentResponseError(propertyName string, errorMsg string) {
    // 直接存储 Agent 响应：失败时返回 _error 字段
    dc.agentResponses[propertyName] = &interfaces.AgentResponseDebugInfo{
        Error: errorMsg,
    }
}

// SetNowMs 设置当前时间戳
func (dc *DebugCollector) SetNowMs(nowMs int64) {
    dc.nowMs = nowMs
}

// SetTraceID 设置追踪 ID
func (dc *DebugCollector) SetTraceID(traceID string) {
    dc.traceID = traceID
}

// AddWarning 添加警告信息
func (dc *DebugCollector) AddWarning(warning string) {
    dc.warnings = append(dc.warnings, warning)
}

// BuildDebugInfo 构建最终的 Debug 信息
func (dc *DebugCollector) BuildDebugInfo() *interfaces.ResolveDebugInfo {
    // 将 property_types、agent_requests、agent_responses 合并为 agent_info
    agentInfo := make(map[string]*interfaces.AgentInfo)
    for propertyName, propertyType := range dc.propertyTypes {
        agentInfo[propertyName] = &interfaces.AgentInfo{
            PropertyType: propertyType,
            Request:      dc.agentRequests[propertyName],
            Response:     dc.agentResponses[propertyName],
        }
    }

    return &interfaces.ResolveDebugInfo{
        DynamicParams: dc.dynamicParams,
        AgentInfo:     agentInfo,
        NowMs:         dc.nowMs,
        Warnings:      dc.warnings,
        TraceID:       dc.traceID,
    }
}
```

### 3.4 修改错误处理逻辑

#### 3.4.1 缺参场景的 debug 模式差异化处理

**问题**：当前缺参场景无论是否开启 debug 都会抛出错误，导致错误信息被统一框架包装到 `details` 字段中，不便于调试。

**解决方案**：根据是否开启 debug 模式，采用不同的处理方式：

| 场景 | 处理方式 | 返回结构 |
|------|---------|---------|
| **开启 debug** | 不抛出错误，返回正常响应 | `{datas: [], debug: {...}}` |
| **未开启 debug** | 抛出 `MissingParamsError` 错误 | `{error_code, message, missing, trace_id}` |

**优势**：
1. 避免错误信息被包装到 `details` 字段中
2. 开启 debug 时，开发者可以在 `debug.agent_info.{property_name}.response.error` 中看到清晰的错误信息
3. 未开启 debug 时保持现有行为，向后兼容

#### 3.4.2 修改业务逻辑中的缺参处理

**位置**：`server/logics/knlogicpropertyresolver/index.go`

**关键改动**：

```go
// 如果有缺参，根据是否开启 debug 决定处理方式
if len(missingParams) > 0 {
    s.logger.WithContext(ctx).Warnf("[Step 2] ⚠️ 存在缺参: %d 个属性", len(missingParams))

    // 特殊处理：如果开启了 debug，返回正常响应，错误信息放在 debug 中
    if req.Options.ReturnDebug {
        debugInfo := debugCollector.BuildDebugInfo()
        return &interfaces.ResolveLogicPropertiesResponse{
            Datas: []map[string]any{},  // 空数组
            Debug: debugInfo,
        }, nil
    }

    // 未开启 debug：保持现有行为，抛出错误
    missingError := s.buildMissingParamsError(ctx, missingParams)
    return nil, missingError
}
```

**说明**：
- 当 `req.Options.ReturnDebug == true` 时，不抛出错误，而是返回一个正常的响应对象
- `Datas` 字段为空数组 `[]map[string]any{}`
- `Debug` 字段包含完整的 debug 信息，包括 Agent 返回的错误信息
- 当 `req.Options.ReturnDebug == false` 时，保持现有行为，调用 `buildMissingParamsError` 抛出错误

#### 3.4.3 保持现有的错误处理逻辑

**位置**：`server/logics/knlogicpropertyresolver/index.go`

`buildMissingParamsError` 方法保持不变：

```go
func (s *knLogicPropertyResolverService) buildMissingParamsError(
    ctx context.Context,
    missingParams []interfaces.MissingPropertyParams,
) error {
    missingError := &interfaces.MissingParamsError{
        ErrorCode: "MISSING_INPUT_PARAMS",
        Message:   "dynamic_params 缺少必需的 input 参数",
        Missing:   missingParams,
        TraceID:   s.getTraceID(ctx),
    }

    return errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("%+v", missingError))
}
```

**说明**：
- 该方法仅在未开启 debug 时调用
- 保持现有的错误结构，包含 `Missing` 字段
- 不需要修改 HTTP Handler 层的错误处理逻辑

#### 3.4.4 Agent 调用逻辑保持不变

**位置**：`server/drivenadapters/agent_app.go`

Agent 调用逻辑保持不变，因为：
- Agent 返回的错误信息已经通过 `DebugCollector.RecordAgentResponseMissingParams` 方法记录到 debug 信息中
- 开启 debug 时，错误信息会通过 `debug.agent_info.{property_name}.response.error` 字段返回
- 无需修改 Agent 调用逻辑


```
// 调用Agent生成动态参数
answer, err := a.agentClient.CallAgent(ctx, agentReq)
if err != nil {
    return nil, nil, fmt.Errorf("调用Agent失败: %w", err)
}

// 解析Agent返回结果
result, err := parseResultFromAgentV1Answer(answer)
if err != nil {
    return nil, nil, fmt.Errorf("解析Agent结果失败: %w", err)
}

// 检查是否有错误
if result.Error != "" {
    // 直接返回Agent生成的错误消息，无需再次解析
    return nil, nil, fmt.Errorf("Agent返回错误: %s", result.Error)
}

// 检查是否有缺参
if result.Missing != nil && len(result.Missing) > 0 {
    // 直接返回Agent生成的缺参错误消息，无需再次解析
    errorMsg := result.Missing[0].Message // 或者使用其他方式获取错误消息
    return nil, &interfaces.MissingPropertyParams{
        PropertyName: result.Missing[0].PropertyName,
        Message:       errorMsg,
    }, nil
}
```

## 8. 简化方案（基于实际实施）

### 8.1 方案背景

在实际实施过程中，发现原始设计过于复杂，特别是 `MissingParams` 的解析逻辑。经过分析，决定采用简化方案：

1. **简化 MissingPropertyParams 结构**
   - 不再解析具体的参数信息（`Params []MissingParam`）
   - 直接返回 Agent 生成的错误消息（`ErrorMsg string`）
   - 移除 `MissingParam` 结构定义

2. **简化 parseMetricMissingParamsFromError 方法**
   - 不再解析 Agent 返回的 `_error` 字段中的具体参数信息
   - 直接返回 Agent 生成的原始错误消息
   - 只保留 property 名称提取逻辑

3. **简化错误处理流程**
   - 直接返回 Agent 生成的错误消息，无需再次解析
   - 减少代码复杂度，提高可维护性

### 8.2 数据结构修改

#### 8.2.1 修改 MissingPropertyParams 结构

**修改前**：
```go
// MissingPropertyParams 缺参信息
type MissingPropertyParams struct {
    Property string        `json:"property"`
    Params   []MissingParam `json:"params,omitempty"`
}

// MissingParam 缺失的参数
type MissingParam struct {
    Name string `json:"name"`
    Type string `json:"type"`
    Hint string `json:"hint,omitempty"`
}
```

**修改后**：
```go
// MissingPropertyParams 缺参信息
type MissingPropertyParams struct {
    Property  string `json:"property"`
    ErrorMsg  string `json:"error_msg,omitempty"`
}
```

**修改说明**：
- 移除 `Params []MissingParam` 字段
- 新增 `ErrorMsg string` 字段，直接存储 Agent 生成的错误消息
- 移除 `MissingParam` 结构定义

#### 8.2.2 修改 MissingParamsError 结构

**修改前**：
```go
// MissingParamsError 缺参错误
type MissingParamsError struct {
    Missing []MissingPropertyParams `json:"missing"`
    Debug   *ResolveDebugInfo       `json:"debug,omitempty"`
}
```

**修改后**：
```go
// MissingParamsError 缺参错误
type MissingParamsError struct {
    Missing []MissingPropertyParams `json:"missing"`
    Debug   *ResolveDebugInfo       `json:"debug,omitempty"`
    TraceID string                  `json:"trace_id,omitempty"`
}
```

**修改说明**：
- 新增 `TraceID` 字段，用于追踪请求
- `TraceID` 字段暂时设置为空字符串，添加 TODO 注释说明需要从 context 中提取

### 8.3 方法实现修改

#### 8.3.1 简化 parseMetricMissingParamsFromError 方法

**修改前**：
```go
func (a *AgentApp) parseMetricMissingParamsFromError(
    ctx context.Context,
    propertyName string,
    agentAnswer *AgentV1Answer,
) (*interfaces.MissingPropertyParams, error) {
    // 提取 _error 字段
    errorMsg := agentAnswer.Error
    if errorMsg == "" {
        return nil, errors.New("agent returned _error field is empty")
    }

    // 解析错误消息，提取参数信息
    // ... 复杂的解析逻辑
    // 构建 Params 数组
    params := make([]interfaces.MissingParam, 0)
    // ... 参数构建逻辑

    return &interfaces.MissingPropertyParams{
        Property: propertyName,
        Params:   params,
    }, nil
}
```

**修改后**：
```go
func (a *AgentApp) parseMetricMissingParamsFromError(
    ctx context.Context,
    propertyName string,
    agentAnswer *AgentV1Answer,
) (*interfaces.MissingPropertyParams, error) {
    // 提取 _error 字段
    errorMsg := agentAnswer.Error
    if errorMsg == "" {
        return nil, errors.New("agent returned _error field is empty")
    }

    // 直接返回 Agent 生成的错误消息，不再解析具体参数信息
    return &interfaces.MissingPropertyParams{
        Property: propertyName,
        ErrorMsg: errorMsg,
    }, nil
}
```

**修改说明**：
- 移除复杂的参数解析逻辑
- 直接返回 Agent 生成的原始错误消息
- 只保留 property 名称提取逻辑

#### 8.3.2 简化 buildMissingParamsError 方法

**修改前**：
```go
func (s *knLogicPropertyResolverService) buildMissingParamsError(
    ctx context.Context,
    missingParams []interfaces.MissingPropertyParams,
    debugInfo *interfaces.ResolveDebugInfo,
) error {
    return &interfaces.MissingParamsError{
        Missing: missingParams,
        Debug:   debugInfo,
    }
}
```

**修改后**：
```go
func (s *knLogicPropertyResolverService) buildMissingParamsError(
    ctx context.Context,
    errorMsg string,
    debugInfo *interfaces.ResolveDebugInfo,
) error {
    return &interfaces.MissingParamsError{
        Missing: []interfaces.MissingPropertyParams{
            {
                ErrorMsg: errorMsg,
            },
        },
        Debug:   debugInfo,
        TraceID: "", // TODO: 从 context 中提取 trace_id
    }
}
```

**修改说明**：
- 参数从 `missingParams []interfaces.MissingPropertyParams` 改为 `errorMsg string`
- 直接构建包含 `ErrorMsg` 的 `MissingParamsError`
- 添加 `TraceID` 字段，暂时设置为空字符串

#### 8.3.3 修改 generateDynamicParams 方法

**修改前**：
```go
func (s *knLogicPropertyResolverService) generateDynamicParams(
    ctx context.Context,
    req *interfaces.ResolveLogicPropertiesRequest,
    logicPropertiesDef map[string]*interfaces.LogicPropertyDef,
    debugCollector *DebugCollector,
) (dynamicParams map[string]interface{}, missingParams []interfaces.MissingPropertyParams, err error) {
    // ... 生成逻辑
    if len(missingParams) > 0 {
        return nil, nil, nil
    }
    return dynamicParams, nil, nil
}
```

**修改后**：
```go
func (s *knLogicPropertyResolverService) generateDynamicParams(
    ctx context.Context,
    req *interfaces.ResolveLogicPropertiesRequest,
    logicPropertiesDef map[string]*interfaces.LogicPropertyDef,
    debugCollector *DebugCollector,
) (dynamicParams map[string]interface{}, errorMsg string, err error) {
    // ... 生成逻辑
    if errorMsg != "" {
        return nil, errorMsg, nil
    }
    return dynamicParams, "", nil
}
```

**修改说明**：
- 返回值从 `missingParams []interfaces.MissingPropertyParams` 改为 `errorMsg string`
- 错误检查逻辑从 `len(missingParams) > 0` 改为 `errorMsg != ""`

#### 8.3.4 修改 ResolveLogicProperties 方法

**修改前**：
```go
func (s *knLogicPropertyResolverService) ResolveLogicProperties(
    ctx context.Context,
    req *interfaces.ResolveLogicPropertiesRequest,
) (*interfaces.ResolveLogicPropertiesResponse, error) {
    // ... 前置逻辑
    dynamicParams, missingParams, err := s.generateDynamicParams(ctx, req, logicPropertiesDef, debugCollector)
    if err != nil {
        return nil, err
    }
    if len(missingParams) > 0 {
        return nil, s.buildMissingParamsError(ctx, missingParams, debugInfo)
    }
    // ... 后续逻辑
}
```

**修改后**：
```go
func (s *knLogicPropertyResolverService) ResolveLogicProperties(
    ctx context.Context,
    req *interfaces.ResolveLogicPropertiesRequest,
) (*interfaces.ResolveLogicPropertiesResponse, error) {
    // ... 前置逻辑
    dynamicParams, errorMsg, err := s.generateDynamicParams(ctx, req, logicPropertiesDef, debugCollector)
    if err != nil {
        return nil, err
    }
    if errorMsg != "" {
        return nil, s.buildMissingParamsError(ctx, errorMsg, debugInfo)
    }
    // ... 后续逻辑
}
```

**修改说明**：
- 接收 `errorMsg` 变量而非 `missingParams`
- 错误检查逻辑从 `len(missingParams) > 0` 改为 `errorMsg != ""`
- 将 `errorMsg` 作为参数传入 `buildMissingParamsError` 方法

### 8.4 修改文件清单

1. **server/interfaces/kn_logic_property_resolver.go**
   - 修改 `MissingPropertyParams` 结构（移除 `Params` 字段，新增 `ErrorMsg` 字段）
   - 修改 `MissingParamsError` 结构（新增 `TraceID` 字段）
   - 移除 `MissingParam` 结构定义

2. **server/drivenadapters/agent_app.go**
   - 简化 `parseMetricMissingParamsFromError` 方法（直接返回 Agent 错误消息）
   - 简化 `parseOperatorMissingParamsFromError` 方法（直接返回 Agent 错误消息）

3. **server/logics/knlogicpropertyresolver/index.go**
   - 修改 `buildMissingParamsError` 方法（接收 `errorMsg` 参数，添加 `TraceID` 字段）
   - 修改 `generateDynamicParams` 方法（返回 `errorMsg` 而非 `missingParams`）
   - 修改 `ResolveLogicProperties` 方法（使用 `errorMsg` 进行错误检查）

### 8.5 简化方案的优势

1. **降低复杂度**
   - 移除复杂的参数解析逻辑
   - 减少代码量，提高可维护性

2. **提高可靠性**
   - 直接返回 Agent 生成的错误消息，避免解析错误
   - 减少因解析逻辑导致的 bug

3. **简化调试**
   - Agent 生成的错误消息直接返回给用户
   - 用户可以直接看到 Agent 的原始错误信息

4. **向后兼容**
   - 新增字段使用 `omitempty` 标签
   - 不会影响现有调用方

### 8.6 后续优化建议

1. **从 context 中提取 TraceID**
   - 实现 `getTraceID` 方法
   - 从 context 中提取 trace_id
   - 更新 `buildMissingParamsError` 方法

2. **统一错误消息格式**
   - 确保 Agent 生成的错误消息格式统一
   - 便于前端解析和展示

3. **添加错误消息国际化支持**
   - 支持多语言错误消息
   - 提高用户体验

4. **添加错误消息分类**
   - 区分不同类型的错误（参数缺失、参数校验失败、Agent 调用失败等）
   - 便于错误处理和展示

## 9. 总结

本设计方案通过以下方式解决了当前的调试问题：

1. **改进 debug 返回结构**
   - 添加了 `AgentInfo` 字段（合并 PropertyTypes、AgentRequests、AgentResponses）
   - `AgentRequestDebugInfo` 直接存储 Agent 请求结构（`MetricDynamicParamsGeneratorReq` 或 `OperatorDynamicParamsGeneratorReq`）
   - `AgentResponseDebugInfo` 直接存储 Agent 响应信息（成功时包含 `dynamic_params`，失败时包含 `_error`）
   - 移除了额外的包装字段（如 `Success`、`ValidationPassed` 等），保持与 Agent 请求/响应的一致性
   - 提供了完整的 agent 调用过程追踪
   - 结构更简洁，组织更清晰

2. **改进 debug 返回时机**
   - 只要 agent 请求成功，就返回 debug 信息
   - 参数校验失败时也返回 debug 信息

3. **提供 DebugCollector 工具类**
   - 统一管理 debug 信息收集
   - 在 Agent 请求构建后直接记录请求信息（`RecordMetricAgentRequest`、`RecordOperatorAgentRequest`）
   - 直接存储 Agent 的请求和响应结构，不进行额外包装
   - 简化业务逻辑代码

4. **支持错误响应中的 debug 信息**
   - 修改了 `MissingParamsError` 结构
   - 添加了 `error_msg` 字段（直接返回 Agent 生成的原始错误消息）
   - 添加了 `debug` 字段（仅在开启 debug 时返回）
   - 未开启 debug 时不构建 debug 信息，减少不必要的计算开销

5. **优化缺参场景的错误处理（简化方案）**
   - 直接返回 Agent 生成的错误消息，无需再次解析
   - 简化 `MissingPropertyParams` 结构，移除复杂的参数解析逻辑
   - 降低代码复杂度，提高可维护性
   - 保留了排查所需的所有信息

通过这些改进，用户可以更方便地调试 `kn-logic-property-resolver` 接口，无需查看日志和代码即可了解传递给 agent 的参数和 agent 的响应，同时减少了 token 消耗和代码复杂度。