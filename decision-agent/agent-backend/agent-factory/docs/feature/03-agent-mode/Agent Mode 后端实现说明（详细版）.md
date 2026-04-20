# Agent Mode 后端实现说明（详细版）

## 1. 文档目的与范围

本文面向后端开发、架构设计、联调负责人，用于完整说明本次 `Agent Mode` 在 `agent-factory` 中的实现方式，包括：

1. 配置模型与枚举的新增/调整
2. 请求校验与模式归一化逻辑
3. 新增的 React Agent 专用接口
4. 详情接口返回逻辑调整
5. OpenAPI 与文档产物同步链路
6. 已覆盖的测试与验证方式

本文只覆盖 **agent-factory 后端**，不包含 `agent-web` 本地试验性实现，也不讨论最终前端交互方案。

---

## 2. 变更概述

本次改动的目标是将 Agent 配置中的“模式”从分散表达统一为 `config.mode`，同时为 `react` 模式提供清晰的专属配置与创建入口。

### 核心目标
- **降低使用门槛**：通过 ReAct 模式减少用户对 Dolphin 语句的依赖
- **统一模式表达**：用 `config.mode` 作为唯一的主模式字段
- **保持向后兼容**：原有的角色指令模式仍可正常使用
- **预留扩展空间**：为未来可能引入的其他模式做好架构准备

最终形成的后端能力包括：

1. `config.mode` 作为 Agent 模式主字段
2. 支持 `default` / `dolphin` / `react` 三种模式
3. 对外统一使用 `react_config`
4. 新增 `POST /api/agent-factory/v3/agent/react`
5. 详情接口返回时补齐空 `mode`
6. OpenAPI 产物与 `docs/api/README.md` 同步更新

---

## 3. 数据模型调整

### 3.1 AgentMode 枚举

文件：

```text
src/domain/enum/cdaenum/agent_mode.go
```

定义如下：

```go
const (
	AgentModeDefault AgentMode = "default"
	AgentModeDolphin AgentMode = "dolphin"
	AgentModeReact   AgentMode = "react"
)
```

枚举校验通过 `EnumCheck()` 完成，只允许上述 3 个值。

---

### 3.2 Config 结构体新增/调整

文件：

```text
src/domain/valueobject/daconfvalobj/config.go
```

关键字段如下：

| 字段 | 说明 |
| --- | --- |
| `Mode` | Agent 主模式字段 |
| `IsDolphinMode` | Dolphin 模式标记，仍参与当前后端逻辑判断 |
| `IsUseToolIDInDolphin` | Dolphin 模式下是否使用 tool id |
| `PlanMode` | 任务规划模式配置 |
| `ReactConfig` | ReAct 模式专属配置 |
| `ConversationHistoryConfig` | 会话历史配置 |

当前结构中的模式相关字段并存关系如下：

1. `Mode` 是新的主模式字段
2. `IsDolphinMode` 仍被保留，用于 Dolphin 逻辑判断与旧数据归一化
3. `ReactConfig` 是 `react` 模式专属配置
4. 旧字段 `non_dolphin_mode_config` 已移除，不再作为请求字段解析

---

### 3.3 ReactConfig 结构

文件：

```text
src/domain/valueobject/daconfvalobj/react_config.go
```

结构如下：

```go
type ReactConfig struct {
	DisableHistoryInAConversation bool `json:"disable_history_in_a_conversation"`
	DisableLLMCache               bool `json:"disable_llm_cache"`
}
```

当前 `ValObjCheck()` 为空实现，即该对象本身暂时没有复杂校验，仅由 `mode` 约束其使用位置。

---

## 4. 模式归一化与校验逻辑

### 4.1 总入口

模式相关校验的主入口位于：

```text
src/domain/valueobject/daconfvalobj/config.go
func (p *Config) ValObjCheckWithCtx(...)
```

其处理顺序中，与 Agent Mode 最相关的步骤如下：

1. 调用 `normalizeMode()` 对模式字段做归一化
2. 执行 `p.Mode.EnumCheck()`
3. 执行 `checkAboutDolphin()`
4. 校验 `plan_mode`
5. 校验 `react_config`
6. 校验 `conversation_history_config`

---

### 4.2 normalizeMode() 逻辑

文件：

```text
src/domain/valueobject/daconfvalobj/config.go
```

逻辑摘要：

### 4.2.1 `mode == dolphin`

当 `mode == "dolphin"` 时：

1. 自动将 `is_dolphin_mode` 对齐为启用

### 4.2.2 `mode == default` 或 `mode == react`

当 `mode` 为 `default` 或 `react` 时：

1. 如果当前不是 dolphin，则将 `is_dolphin_mode` 对齐为关闭

### 4.2.3 `mode == ""`

当 `mode` 为空时：

1. 调用 `GetMode()` 推导实际模式
2. 如果推导结果为 `dolphin`，则将 `is_dolphin_mode` 对齐为启用

这一分支的主要作用是兼容历史数据或旧存量数据中 `mode` 为空的情况。

---

### 4.3 GetMode() 逻辑

`GetMode()` 的行为如下：

1. `Config == nil` 时，返回 `default`
2. `Mode` 为空且 `is_dolphin_mode == true` 时，返回 `dolphin`
3. `Mode` 为空且非 dolphin 时，返回 `default`
4. `Mode` 非空时，直接返回 `Mode`

这意味着：

1. 存储中的空 `mode` 不会直接暴露为“未知模式”
2. 旧数据在读取和校验链路中仍可得到一个稳定模式值

---

### 4.4 Dolphin 相关校验

逻辑位于：

```text
src/domain/valueobject/daconfvalobj/config.go
func (p *Config) checkAboutDolphin() error
```

关键规则：

1. `is_dolphin_mode` 必须是合法枚举值
2. 当 `mode != ""` 且 `mode != dolphin` 且 `is_dolphin_mode == true` 时，报错：

```text
[Config]: mode conflicts with is_dolphin_mode
```

3. 当 `is_dolphin_mode == true` 时，`pre_dolphin` / `post_dolphin` / `dolphin` 至少有一个有内容，否则报错

这保证了：

1. 非 Dolphin 模式不会与 `is_dolphin_mode` 形成自相矛盾
2. Dolphin 模式在配置层必须具备最基本的执行内容

---

### 4.5 PlanMode 与 ReactConfig 校验

同样位于：

```text
src/domain/valueobject/daconfvalobj/config.go
```

### 4.5.1 PlanMode 规则

当 `plan_mode.is_enabled == true` 且 `is_dolphin_mode == true` 时，返回错误：

```text
[Config]: plan_mode is invalid when is_dolphin_mode is true
```

即：

1. `plan_mode` 允许用于 `default`
2. `plan_mode` 允许用于 `react`
3. `plan_mode` 不允许与 Dolphin 模式同时启用

### 4.5.2 ReactConfig 规则

当 `react_config != nil` 时：

1. 如果 `mode != react`，返回错误：

```text
[Config]: react_config is only allowed when mode is react
```

2. 如果 `mode == react`，则继续执行 `ReactConfig.ValObjCheck()`

因此，`react_config` 是一个严格绑定 `react` 模式的专属配置对象。

---

## 5. 架构设计考虑

### 5.1 扩展性设计

当前的 AgentMode 枚举和 Config 结构体设计充分考虑了扩展性：

1. **枚举扩展**：新增模式只需在 `AgentMode` 枚举中添加常量
2. **配置隔离**：每种模式都有专属的配置字段（如 `ReactConfig`），避免字段冲突
3. **校验框架**：`ValObjCheckWithCtx` 提供了统一的校验入口，便于添加新模式校验

### 5.2 为未来新模式预留说明

虽然当前主要聚焦于 ReAct 模式，但架构设计已为未来可能引入的 Prompt、Plan 等模式预留了空间。新增模式时只需：

1. 在枚举中添加新模式
2. 定义专属配置结构体
3. 在校验逻辑中添加相应规则
4. 更新 Swagger 模型

这种设计确保了系统的可扩展性和维护性。

---

## 6. 对外接口调整

### 6.1 普通创建接口保持不变

路径：

```text
POST /api/agent-factory/v3/agent
```

实现文件：

```text
src/driveradapter/api/httphandler/agentconfighandler/create.go
```

该接口仍然是统一的创建入口，底层服务逻辑没有因为新增 `react` 而拆分。

---

### 6.2 新增 React Agent 专用创建接口

路径：

```text
POST /api/agent-factory/v3/agent/react
```

注册位置：

```text
src/driveradapter/api/httphandler/agentconfighandler/define.go
```

实现文件：

```text
src/driveradapter/api/httphandler/agentconfighandler/create_react.go
```

其核心实现非常轻量：

```go
func (h *daConfHTTPHandler) CreateReact(c *gin.Context) {
	h.create(c, validateCreateReactReq)
}
```

也就是说：

1. HTTP Handler 层新增了一个语义化路由
2. Service 层仍复用普通创建逻辑
3. 通过 `extraCheck` 注入一层额外校验

---

### 6.3 React 创建接口的额外校验

校验函数位于：

```text
src/driveradapter/api/httphandler/agentconfighandler/create.go
func validateCreateReactReq(...)
```

校验规则：

1. 如果 `req.Config == nil`，不在此层报错，继续走后续通用校验
2. 如果 `req.Config.Mode != react`，直接返回：

```text
config.mode must be "react"
```

对应行为：

1. 该错误发生在 HTTP Handler 层
2. 响应码为 `400`
3. 后续不会进入 Service 创建逻辑

这样做的目的：

1. 将“React 专用路由”的语义约束放在最外层
2. 避免 Service 层新增一套重复方法
3. 保持普通创建和 React 创建共用同一套主体流程

---

### 6.4 私有路由说明

`RegPubRouter()` 中新增了：

```text
/agent/react
```

当前 `RegPriRouter()` 中没有同步新增 `/agent/react` 私有路由。

这意味着本次新增的是公开路由组下的 React 专用创建入口，私有创建仍沿用原 `/agent`。

---

## 7. 详情接口返回逻辑调整

### 7.1 调整原因

为保证详情接口对前端/调用方返回的 `config.mode` 稳定可用，当前详情响应会在返回前补齐空 `mode`。

---

### 7.2 实现位置

文件：

```text
src/driveradapter/api/rdto/agent_config/agentconfigresp/detail.go
```

当前 `DetailRes.Config` 已改回：

```go
Config *daconfvalobj.Config `json:"config"`
```

不再使用单独的 `ConfigForShow` 结构。

---

### 7.3 处理方式

`LoadFromEo()` 的逻辑如下：

1. 先将 `eo` 拷贝到响应对象
2. 如果 `eo.Config == nil`，则响应中的 `Config` 直接为 `nil`
3. 否则额外复制一份 `Config` 作为响应副本
4. 如果副本中的 `Mode == ""`，执行：

```go
respConfig.Mode = respConfig.GetMode()
```

5. 将该副本赋值给响应对象

这意味着：

1. 详情响应会补齐空 `mode`
2. 不会修改数据库读取出来的原始 `eo.Config`
3. 返回结构仍然是统一的 `Config`

---

## 8. Swagger / OpenAPI 模型同步

### 8.1 Swagger 模型文件

文件：

```text
src/driveradapter/api/rdto/swagger/agent_config_models.go
```

当前 `AgentConfigConfig` 已包含：

1. `mode`
2. `is_dolphin_mode`
3. `is_use_tool_id_in_dolphin`
4. `plan_mode`
5. `react_config`
6. `conversation_history_config`

这保证了：

1. 创建接口文档
2. 更新接口文档
3. 详情接口文档

在 OpenAPI 层都能反映当前模式设计。

---

### 8.2 新增路由的 Swagger 注释

`CreateReact` 的注释中明确说明：

1. 这是“创建 react agent”接口
2. 请求体与普通创建接口一致
3. `config.mode` 必须为 `react`

OpenAPI 生成后，`/v3/agent/react` 与 `/v3/agent` 使用相同 schema。

---

### 8.3 文档产物同步

本次变更已同步到以下文档链路：

1. `docs/api/README.md`
2. `docs/api/agent-factory.json`
3. `docs/api/agent-factory.yaml`
4. `docs/api/agent-factory.html`
5. `docs/api/agent-factory-redoc.html`
6. `src/infra/server/apidocs/assets/*`
7. `cmd/openapi-docs/generated/swagger/*`

`docs/api/README.md` 中当前已经补充：

1. `v3 Agent Config` 模式说明
2. `react` 模式使用 `react_config`
3. React Agent 专用创建接口说明

---

## 9. 错误场景说明

以下是当前与 Agent Mode 直接相关的典型错误场景。

### 9.1 React 专用创建接口传入非 react 模式

请求：

```json
{
  "config": {
    "mode": "default"
  }
}
```

接口：

```text
POST /api/agent-factory/v3/agent/react
```

结果：

1. Handler 层返回 `400`
2. 错误信息：

```text
config.mode must be "react"
```

### 9.2 非 react 模式携带 react_config

请求：

```json
{
  "config": {
    "mode": "default",
    "react_config": {
      "disable_llm_cache": true
    }
  }
}
```

结果：

1. 配置校验失败
2. 错误信息：

```text
[Config]: react_config is only allowed when mode is react
```

### 9.3 非 dolphin 模式与 is_dolphin_mode 冲突

请求：

```json
{
  "config": {
    "mode": "react",
    "is_dolphin_mode": 1
  }
}
```

结果：

1. 配置校验失败
2. 错误信息：

```text
[Config]: mode conflicts with is_dolphin_mode
```

### 9.4 Dolphin 模式启用 plan_mode

结果：

```text
[Config]: plan_mode is invalid when is_dolphin_mode is true
```

---

## 10. 测试与验证

本次 Agent Mode 相关实现已经补充或覆盖以下测试点：

1. `CreateReact` 在 `mode != react` 时返回 `400`
2. `CreateReact` 正常创建成功
3. Swagger / OpenAPI 中 `/agent/react` 使用与普通创建一致的 schema
4. 详情接口返回中 `react_config` 正常序列化
5. 详情接口在 `mode` 为空时返回补齐后的 `mode`
6. `react` 枚举值与 `react_config` 字段在 swagger 模型中正常存在

本地验证命令包括：

```bash
go test ./src/driveradapter/api/httphandler/agentconfighandler ./src/infra/server/apidocs
make gen-api-docs
make validate-api-docs
```

如果只关注详情返回链路，也可以执行：

```bash
go test ./src/driveradapter/api/rdto/agent_config/agentconfigresp ./src/driveradapter/api/rdto/swagger ./src/domain/valueobject/daconfvalobj
```

---

## 11. 对前端/产品对接时建议明确的点

虽然本文聚焦后端实现，但在进入前后端联调前，建议明确以下对接结论：

1. 页面层是否以 `mode` 作为唯一模式切换依据
2. 是否使用普通创建接口还是 React 专用创建接口作为主要入口
3. `react_config` 的默认值是否由前端显式下发，还是依赖后端零值
4. `default` / `react` / `dolphin` 三种模式下的字段显隐规则
5. 是否需要在产品文档中明确 `plan_mode` 与 Dolphin 模式互斥

---

## 12. 结论

本次 Agent Mode 改动在后端层面完成了四件关键事情：

1. 用 `config.mode` 统一了 Agent 模式表达
2. 用 `react_config` 明确了 ReAct 模式的专属配置边界
3. 用 `/v3/agent/react` 提供了一个语义清晰、但不增加 Service 复杂度的专用创建入口
4. 通过架构设计为未来扩展其他模式预留了空间

同时，详情返回和 OpenAPI 文档链路也已同步到位，已经具备与前端、产品进行正式方案对接的基础。

### 技术亮点
- **渐进式改进**：在保持兼容性的同时引入新模式
- **清晰的边界**：每种模式都有明确的配置和使用边界
- **良好的扩展性**：为未来功能扩展做好了架构准备
