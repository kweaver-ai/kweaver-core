# kn_search 接口实现设计文档

## 1. 需求概述

### 1.1 目标
在 agent-retrieval 服务中代理 data-retrieval 服务的 kn_search 接口，统一对外提供服务，便于 Agent/大模型调用。

### 1.2 接口信息
- **路由**: `POST /kn/kn_search`
- **目标接口**: `POST http://data-retrieval:9100/tools/kn_search`
- **功能**: 基于知识网络的智能检索工具，支持传入完整的问题或一个或多个关键词，能够检索问题或关键词的属性信息和上下文信息

## 2. 技术架构设计

### 2.1 整体架构
采用项目现有的分层架构：
```
HTTP请求 → driveradapters (处理器层)
         → logics (业务逻辑层)
         → drivenadapters (外部服务调用层)
         → data-retrieval服务
```

### 2.2 设计理念

**透传设计原则**：
- 对于底层接口的复杂数据结构（如 `retrieval_config`、`object_types`、`nodes` 等），使用 `any` 类型进行透传
- 避免在当前服务中明确定义底层接口的数据结构，减少维护成本
- 当底层接口结构变动时，无需修改当前服务的代码，提高系统的灵活性
- 请求体和响应体直接透传，保持数据的完整性和准确性

**参数透传**：
- Header 参数（`x-account-id`、`x-account-type`）透传给目标服务
- Body 参数完全透传，包括所有可选配置项
- 响应结构完全透传，保持原始数据格式

### 2.3 模块划分

#### 2.3.1 driveradapters 层
- **文件**: `server/driveradapters/knsearch/index.go`
- **职责**:
  - HTTP 请求参数绑定和校验
  - 调用业务逻辑层
  - 响应封装
- **处理器**: `KnSearchHandler`

#### 2.3.2 logics 层
- **文件**: `server/logics/knsearch/index.go`
- **职责**:
  - 业务逻辑处理
  - 错误信息包装（大模型可理解）
- **服务**: `KnSearchService`

#### 2.3.3 drivenadapters 层
- **文件**: `server/drivenadapters/data_retrieval.go` (扩展现有文件)
- **职责**:
  - 调用 data-retrieval 服务的 kn_search 接口
  - HTTP 请求/响应处理

#### 2.3.4 interfaces 层
- **文件**: `server/interfaces/drivenadapters.go` (扩展现有文件)
- **职责**:
  - 定义 kn_search 相关的接口和数据结构

## 3. 数据结构设计

### 3.1 请求结构

```go
// KnSearchReq kn_search 请求
type KnSearchReq struct {
    // Header 参数
    XAccountID   string `header:"x-account-id"`
    XAccountType string `header:"x-account-type"`

    // Body 参数 - 使用 any 避免明确定义复杂结构
    // 对应 data-retrieval 接口的完整请求结构
    Query              string  `json:"query" binding:"required"`
    KnIDs              any     `json:"kn_ids" binding:"required"`
    SessionID          *string `json:"session_id,omitempty"`
    AdditionalContext  *string `json:"additional_context,omitempty"`
    RetrievalConfig    any     `json:"retrieval_config,omitempty"`
    OnlySchema         *bool   `json:"only_schema,omitempty"`
}
```

**设计说明**：
- 使用 `any` 类型定义 `KnIDs` 和 `RetrievalConfig`，避免明确定义底层接口的复杂结构
- 底层接口结构变动时，无需修改当前服务的代码
- 请求体直接透传给 data-retrieval 服务
- 必需参数（`query`、`kn_ids`）进行校验，可选参数透传

### 3.2 响应结构

```go
// KnSearchResp kn_search 响应
type KnSearchResp struct {
    // 使用 any 直接返回底层接口的原始结构
    // 对应 data-retrieval 接口的完整响应结构
    ObjectTypes  any     `json:"object_types,omitempty"`
    RelationTypes any    `json:"relation_types,omitempty"`
    ActionTypes   any    `json:"action_types,omitempty"`
    Nodes         any    `json:"nodes,omitempty"`
    Message       *string `json:"message,omitempty"`
}
```

**设计说明**：
- 使用 `any` 类型定义响应字段，直接返回底层接口的原始结构
- 避免在当前服务中明确定义 `ObjectType`、`RelationType`、`ActionType`、`Node` 等复杂结构
- 底层接口响应结构变动时，无需修改当前服务的代码
- 响应体直接透传给调用方，保持数据完整性

## 4. 接口设计

### 4.1 DataRetrieval 接口扩展

在 `server/interfaces/drivenadapters.go` 中扩展 `DataRetrieval` 接口：

```go
// DataRetrieval 数据检索接口
type DataRetrieval interface {
    KnowledgeRerank(ctx context.Context, req *KnowledgeRerankReq) (results []*ConceptResult, err error)
    // KnSearch 知识网络检索
    KnSearch(ctx context.Context, req *KnSearchReq) (resp *KnSearchResp, err error)
}
```

**接口设计说明**：
- 请求和响应结构体中的复杂字段（如 `KnIDs`、`RetrievalConfig`、`ObjectTypes`、`Nodes` 等）使用 `any` 类型
- 这样可以避免在当前服务中明确定义底层接口的复杂数据结构
- 当底层接口结构变动时，无需修改当前服务的代码
- 请求体和响应体直接透传，保持数据的完整性和灵活性

### 4.2 IKnSearchService 接口定义

在 `server/interfaces/interface.go` 中添加：

```go
// IKnSearchService kn_search 服务接口
type IKnSearchService interface {
    // KnSearch 知识网络检索
    KnSearch(ctx context.Context, req *KnSearchReq) (resp *KnSearchResp, err error)
}
```

## 5. 实现细节

### 5.1 driveradapters 层实现

**文件**: `server/driveradapters/knsearch/index.go`

```go
package knsearch

import (
    "net/http"
    "sync"

    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/config"
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/errors"
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/rest"
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/interfaces"
    logicskn "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/logics/knsearch"
    "github.com/creasty/defaults"
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)

// KnSearchHandler kn_search 处理器
type KnSearchHandler interface {
    KnSearch(c *gin.Context)
}

type knSearchHandler struct {
    Logger          interfaces.Logger
    KnSearchService interfaces.IKnSearchService
}

var (
    ksOnce    sync.Once
    ksHandler KnSearchHandler
)

// NewKnSearchHandler 新建 KnSearchHandler
func NewKnSearchHandler() KnSearchHandler {
    ksOnce.Do(func() {
        conf := config.NewConfigLoader()
        ksHandler = &knSearchHandler{
            Logger:          conf.GetLogger(),
            KnSearchService: logicskn.NewKnSearchService(),
        }
    })
    return ksHandler
}

// KnSearch 知识网络检索
func (h *knSearchHandler) KnSearch(c *gin.Context) {
    var err error
    req := &interfaces.KnSearchReq{}

    // 绑定 Header
    if err = c.ShouldBindHeader(req); err != nil {
        err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
        rest.ReplyError(c, err)
        return
    }

    // 绑定 JSON Body
    if err = c.ShouldBindJSON(req); err != nil {
        err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
        rest.ReplyError(c, err)
        return
    }

    // 设置默认值
    if err = defaults.Set(req); err != nil {
        err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
        rest.ReplyError(c, err)
        return
    }

    // 参数校验
    err = validator.New().Struct(req)
    if err != nil {
        rest.ReplyError(c, err)
        return
    }

    // 调用业务逻辑
    resp, err := h.KnSearchService.KnSearch(c.Request.Context(), req)
    if err != nil {
        h.Logger.Errorf("[KnSearchHandler#KnSearch] KnSearch failed, err: %v", err)
        rest.ReplyError(c, err)
        return
    }

    // 返回成功响应
    rest.ReplyOK(c, http.StatusOK, resp)
}
```

### 5.2 logics 层实现

**文件**: `server/logics/knsearch/index.go`

```go
package knsearch

import (
    "context"
    "sync"

    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/drivenadapters"
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/infra/config"
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/interfaces"
)

// KnSearchService kn_search 服务
type KnSearchService interface {
    KnSearch(ctx context.Context, req *interfaces.KnSearchReq) (resp *interfaces.KnSearchResp, err error)
}

type knSearchService struct {
    Logger        interfaces.Logger
    DataRetrieval interfaces.DataRetrieval
}

var (
    ksServiceOnce sync.Once
    ksService     KnSearchService
)

// NewKnSearchService 新建 KnSearchService
func NewKnSearchService() KnSearchService {
    ksServiceOnce.Do(func() {
        conf := config.NewConfigLoader()
        ksService = &knSearchService{
            Logger:        conf.GetLogger(),
            DataRetrieval: drivenadapters.NewDataRetrievalClient(),
        }
    })
    return ksService
}

// KnSearch 知识网络检索
func (s *knSearchService) KnSearch(ctx context.Context, req *interfaces.KnSearchReq) (resp *interfaces.KnSearchResp, err error) {
    // 调用 drivenadapters 层进行检索
    resp, err = s.DataRetrieval.KnSearch(ctx, req)
    return
}
```

### 5.3 drivenadapters 层实现

**文件**: `server/drivenadapters/data_retrieval.go` (扩展现有文件)

```go
const (
    knowledgeRerank = "/tools/knowledge_rerank" // 知识重排接口
    knSearch       = "/tools/kn_search"         // kn_search 接口
)

// KnSearch 知识网络检索
func (dr *dataRetrievalClient) KnSearch(ctx context.Context, req *interfaces.KnSearchReq) (resp *interfaces.KnSearchResp, err error) {
    url := fmt.Sprintf("%s%s", dr.baseURL, knSearch)

    // 构建请求头 - 透传 Header 参数
    header := map[string]string{
        rest.ContentTypeKey: rest.ContentTypeJSON,
    }
    if req.XAccountID != "" {
        header["x-account-id"] = req.XAccountID
    }
    if req.XAccountType != "" {
        header["x-account-type"] = req.XAccountType
    }

    // 构建请求体 - 直接透传所有 Body 参数
    body := map[string]any{
        "query": req.Query,
        "kn_ids": req.KnIDs,
    }
    if req.SessionID != nil {
        body["session_id"] = *req.SessionID
    }
    if req.AdditionalContext != nil {
        body["additional_context"] = *req.AdditionalContext
    }
    if req.RetrievalConfig != nil {
        body["retrieval_config"] = req.RetrievalConfig
    }
    if req.OnlySchema != nil {
        body["only_schema"] = *req.OnlySchema
    }

    // 记录请求日志
    bodyJSON, _ := json.Marshal(body)
    dr.logger.WithContext(ctx).Debugf("[DataRetrieval#KnSearch] URL: %s", url)
    dr.logger.WithContext(ctx).Debugf("[DataRetrieval#KnSearch] Request Body: %s", string(bodyJSON))

    // 发送请求
    _, respBody, err := dr.httpClient.Post(ctx, url, header, body)
    if err != nil {
        dr.logger.WithContext(ctx).Errorf("[DataRetrieval#KnSearch] Request failed, err: %v", err)
        return nil, err
    }

    // 解析响应 - 直接解析到 any
    resp = &interfaces.KnSearchResp{}
    resultByt := utils.ObjectToByte(respBody)
    err = json.Unmarshal(resultByt, resp)
    if err != nil {
        dr.logger.WithContext(ctx).Errorf("[DataRetrieval#KnSearch] Unmarshal failed, body: %s, err: %v", string(resultByt), err)
        err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("解析 kn_search 响应失败: %v", err))
        return nil, err
    }

    // 记录响应日志
    respJSON, _ := json.Marshal(resp)
    dr.logger.WithContext(ctx).Debugf("[DataRetrieval#KnSearch] Response: %s", string(respJSON))

    return resp, nil
}
```

**实现说明**：
- Header 参数（`x-account-id`、`x-account-type`）透传给目标服务
- Body 参数完全透传，包括所有可选配置项
- `KnIDs` 和 `RetrievalConfig` 使用 `any` 类型，直接透传给底层接口
- 响应直接解析到 `any`，保持原始结构

### 5.4 路由注册

**文件**: `server/driveradapters/rest_private_handler.go`

修改现有文件，添加路由注册：

```go
package driveradapters

import (
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knactionrecall"
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knlogicpropertyresolver"
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knqueryobjectinstance"
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knquerysubgraph"
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knretrieval"
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/driveradapters/knsearch"  // 新增
    "devops.aishu.cn/AISHUDevOps/DIP/_git/agent-retrieval/server/interfaces"
    "github.com/gin-gonic/gin"
)

type restPrivateHandler struct {
    KnRetrievalHandler             knretrieval.KnRetrievalHandler
    KnLogicPropertyResolverHandler knlogicpropertyresolver.KnLogicPropertyResolverHandler
    KnActionRecallHandler          knactionrecall.KnActionRecallHandler
    KnQueryObjectInstanceHandler   knqueryobjectinstance.KnQueryObjectInstanceHandler
    KnQuerySubgraphHandler         knquerysubgraph.KnQuerySubgraphHandler
    KnSearchHandler                knsearch.KnSearchHandler  // 新增
    Logger                         interfaces.Logger
}

// NewRestPrivateHandler 创建restHandler实例
func NewRestPrivateHandler(logger interfaces.Logger) interfaces.HTTPRouterInterface {
    return &restPrivateHandler{
        KnRetrievalHandler:             knretrieval.NewKnRetrievalHandler(),
        KnLogicPropertyResolverHandler: knlogicpropertyresolver.NewKnLogicPropertyResolverHandler(),
        KnActionRecallHandler:          knactionrecall.NewKnActionRecallHandler(),
        KnQueryObjectInstanceHandler:   knqueryobjectinstance.NewKnQueryObjectInstanceHandler(),
        KnQuerySubgraphHandler:         knquerysubgraph.NewKnQuerySubgraphHandler(),
        KnSearchHandler:                knsearch.NewKnSearchHandler(),  // 新增
        Logger:                         logger,
    }
}

// RegisterRouter 注册路由
func (r *restPrivateHandler) RegisterRouter(engine *gin.RouterGroup) {
    mws := []gin.HandlerFunc{}
    mws = append(mws, middlewareRequestLog(r.Logger), middlewareTrace, middlewareHeaderAuthContext())
    engine.Use(mws...)

    engine.POST("/kn/semantic-search", r.KnRetrievalHandler.SemanticSearch)
    engine.POST("/kn/logic-property-resolver", r.KnLogicPropertyResolverHandler.ResolveLogicProperties)
    engine.POST("/kn/get_action_info", r.KnActionRecallHandler.GetActionInfo)
    engine.POST("/kn/query_object_instance", r.KnQueryObjectInstanceHandler.QueryObjectInstance)
    engine.POST("/kn/query_instance_subgraph/:kn_id", r.KnQuerySubgraphHandler.QueryInstanceSubgraph)
    engine.POST("/kn/kn_search", r.KnSearchHandler.KnSearch)  // 新增
}
```

## 6. 错误处理设计

### 6.1 错误分类

1. **参数错误** (400)
   - 缺少必需参数（`query`、`kn_ids`）
   - 参数格式错误
   - 参数值不合法

2. **权限错误** (403)
   - 无权访问该知识网络
   - 账户信息无效

3. **数据不存在** (404)
   - 知识网络不存在
   - 查询无结果

4. **服务错误** (500)
   - data-retrieval 服务不可用
   - 网络超时
   - 数据解析失败

### 6.2 错误信息包装

错误信息需要包装成大模型可理解的内容：

```go
// 示例：参数错误
"查询参数错误：缺少必需的查询内容（query）或知识网络ID列表（kn_ids）"

// 示例：服务错误
"知识网络检索服务暂时不可用，请稍后重试。如果问题持续存在，请联系技术支持。"

// 示例：数据不存在
"未找到匹配的知识网络数据，请检查查询内容或知识网络配置是否正确。"
```

## 7. 日志设计

### 7.1 日志级别

- **DEBUG**: 记录完整的请求和响应数据（用于问题排查）
- **INFO**: 记录关键操作节点
- **WARN**: 记录可恢复的错误
- **ERROR**: 记录需要关注的错误

### 7.2 日志内容

```go
// 请求日志
dr.logger.WithContext(ctx).Debugf("[DataRetrieval#KnSearch] URL: %s", url)
dr.logger.WithContext(ctx).Debugf("[DataRetrieval#KnSearch] Request Body: %s", string(bodyJSON))

// 响应日志
dr.logger.WithContext(ctx).Debugf("[DataRetrieval#KnSearch] Response: %s", string(respJSON))

// 错误日志
h.Logger.Errorf("[KnSearchHandler#KnSearch] KnSearch failed, err: %v", err)
dr.logger.WithContext(ctx).Errorf("[DataRetrieval#KnSearch] Request failed, err: %v", err)
```

## 8. 测试计划

### 8.1 单元测试

1. **driveradapters 层测试**
   - 参数绑定测试
   - 参数校验测试（必需参数、可选参数）
   - 错误处理测试

2. **logics 层测试**
   - 业务逻辑测试
   - 错误包装测试

3. **drivenadapters 层测试**
   - HTTP 请求构建测试
   - Header 透传测试
   - Body 透传测试
   - 响应解析测试
   - 错误处理测试

### 8.2 集成测试

1. **端到端测试**
   - 正常流程测试（仅必需参数）
   - 正常流程测试（包含可选参数）
   - 异常流程测试（参数错误、服务不可用等）
   - 边界条件测试（空查询、空知识网络列表等）

2. **性能测试**
   - 并发请求测试
   - 大数据量测试
   - 响应时间测试

### 8.3 测试用例示例

```go
// 测试用例1：仅必需参数
{
    "query": "测试查询",
    "kn_ids": [{"knowledge_network_id": "kn_001"}]
}

// 测试用例2：包含所有可选参数
{
    "query": "测试查询",
    "kn_ids": [{"knowledge_network_id": "kn_001"}],
    "session_id": "session_123",
    "additional_context": "额外上下文",
    "retrieval_config": {
        "concept_retrieval": {
            "top_k": 10,
            "skip_llm": true
        }
    },
    "only_schema": false
}

// 测试用例3：参数错误 - 缺少 query
{
    "kn_ids": [{"knowledge_network_id": "kn_001"}]
}

// 测试用例4：参数错误 - 缺少 kn_ids
{
    "query": "测试查询"
}
```

## 9. 部署计划

### 9.1 部署步骤

1. 代码审查
2. 合并到开发分支
3. 运行单元测试
4. 运行集成测试
5. 部署到测试环境
6. 功能验证
7. 部署到生产环境

### 9.2 回滚计划

如果部署后出现问题，可以通过以下方式回滚：
1. 回滚代码到上一个稳定版本
2. 重新部署
3. 验证功能正常

## 10. 未来扩展

### 10.1 MCP Server 封装

未来可以将此接口封装成 MCP Server 对外提供服务：

```go
// MCP Server 接口定义
type MCPTool struct {
    Name        string
    Description string
    InputSchema map[string]any
    Handler     func(ctx context.Context, input map[string]any) (map[string]any, error)
}

// kn_search MCP Tool
var KnSearchTool = MCPTool{
    Name:        "kn_search",
    Description: "基于知识网络的智能检索工具，支持传入完整的问题或一个或多个关键词，能够检索问题或关键词的属性信息和上下文信息",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "query": map[string]any{
                "type":        "string",
                "description": "用户查询问题或关键词，多个关键词之间用空格隔开",
            },
            "kn_ids": map[string]any{
                "type":        "array",
                "description": "指定的知识网络配置列表",
            },
            "session_id": map[string]any{
                "type":        "string",
                "description": "会话ID，用于维护多轮对话存储的历史召回记录",
            },
            "additional_context": map[string]any{
                "type":        "string",
                "description": "额外的上下文信息",
            },
            "retrieval_config": map[string]any{
                "type":        "object",
                "description": "召回配置参数",
            },
            "only_schema": map[string]any{
                "type":        "boolean",
                "description": "是否只召回概念（schema），不召回语义实例",
            },
        },
        "required": []string{"query", "kn_ids"},
    },
    Handler: handleKnSearch,
}
```

### 10.2 性能优化

1. **缓存策略**
   - 对频繁查询的结果进行缓存
   - 设置合理的缓存过期时间
   - 考虑基于 `query` 和 `kn_ids` 的缓存键

2. **并发优化**
   - 支持批量查询
   - 并行处理多个知识网络的查询

3. **数据压缩**
   - 对大数据量响应进行压缩
   - 减少网络传输时间

### 10.3 功能扩展

1. **查询优化**
   - 支持查询结果过滤
   - 支持结果排序
   - 支持分页查询

2. **监控告警**
   - 添加接口调用监控
   - 添加性能指标监控
   - 添加错误率告警

## 11. 风险评估

### 11.1 技术风险

| 风险项 | 影响 | 概率 | 应对措施 |
|--------|------|------|----------|
| data-retrieval 服务不稳定 | 高 | 中 | 增加重试机制和熔断机制 |
| 大数据量查询超时 | 中 | 高 | 增加查询超时控制和分页支持 |
| 数据结构变更 | 中 | 低 | 使用透传设计，降低影响范围 |
| 复杂配置参数透传错误 | 中 | 中 | 增加参数校验和日志记录 |

### 11.2 业务风险

| 风险项 | 影响 | 概率 | 应对措施 |
|--------|------|------|----------|
| 错误信息不够友好 | 中 | 中 | 持续优化错误提示信息 |
| 性能不满足需求 | 高 | 低 | 进行性能测试和优化 |
| 接口调用频率过高 | 中 | 中 | 添加限流机制 |

## 12. 总结

本设计文档详细描述了 kn_search 接口的实现方案，包括：

1. **架构设计**：采用分层架构，职责清晰
2. **数据结构**：使用透传设计，避免定义复杂结构
3. **接口设计**：定义了各层之间的接口
4. **实现细节**：提供了各层的具体实现代码
5. **错误处理**：设计了完善的错误处理机制和错误信息包装
6. **日志设计**：定义了详细的日志记录方案
7. **测试计划**：制定了完整的测试方案
8. **部署计划**：规划了部署和回滚方案
9. **未来扩展**：考虑了 MCP Server 封装、性能优化和功能扩展
10. **风险评估**：识别了潜在风险并制定了应对措施

该设计方案遵循了项目现有的架构模式，采用透传设计原则，具有良好的可维护性和可扩展性。当底层 data-retrieval 服务的接口结构发生变化时，当前服务的代码修改成本最低。

