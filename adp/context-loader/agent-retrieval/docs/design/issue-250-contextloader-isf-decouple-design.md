# Design Doc: Context-Loader 与 ISF 解耦

> 状态: Draft  
> 负责人: @criller  
> Reviewers: 待确认  
> 关联 PRD: ../prd/issue-250-contextloader-isf-decouple-prd.md  

---

# 1. 概述（Overview）

## 1.1 背景

- 当前现状：
  - Context-Loader (agent-retrieval) 公网路由 `/api/agent-retrieval/v1/*` 通过 `middlewareIntrospectVerify` 中间件调用 Hydra Admin 的 `/oauth2/introspect` 接口进行 Token 内省校验。
  - `NewRestPublicHandler()` 初始化时通过 `drivenadapters.NewHydra()` 创建 Hydra HTTP 客户端，读取 `config.OAuth.BuildAdminURL()` 获取 Hydra Admin 地址。
  - 代码中存在完整的 `UserManagement` 客户端实现（`drivenadapters/user_management.go`）及接口定义，但业务逻辑层和路由层从未调用过任何 `UserManagement` 方法。

- 存在问题：
  - 无 ISF 环境时，`middlewareIntrospectVerify` 会因 Hydra 不可达而导致所有公网请求返回错误，无法进行开发调试。
  - 未使用的 `UserManagement` 代码增加维护成本，配置中 `user_management` 块可能在轻量部署时造成困惑。

- 业务 / 技术背景：
  - 本设计遵循 ISSUE-177 总体逻辑设计中的 `AUTH_ENABLED` 环境变量 + 工厂模式 + Noop 实现的统一方案。
  - Context-Loader 不使用 `authorization-private`（权限服务），因此不需要 PermissionService 抽象。

---

## 1.2 目标

- 通过 `AUTH_ENABLED` 环境变量一键控制 Hydra Token 内省的开关。
- `AUTH_ENABLED=false` 时，公网中间件跳过 Token 校验，直接构造空的 `AccountAuthContext` 放行请求。
- 移除未使用的 `UserManagement` 客户端代码、接口、Mock、测试和相关配置。
- Helm Chart 新增 `auth.enabled` 配置项并注入环境变量。

---

## 1.3 非目标（Out of Scope）

- 不引入 PermissionService 抽象（Context-Loader 无此需求）。
- 不改造内网路由 `/api/agent-retrieval/in/v1/*` 的 Header 认证中间件。
- 不引入运行时热切换 `AUTH_ENABLED` 的能力。
- 不改造 MCP 协议或其他微服务。

---

## 1.4 术语说明

| 术语 | 说明 |
|------|------|
| ISF | 信息安全框架（Information Security Framework），提供 Hydra、Authorization、User-Management 等基础服务 |
| Hydra Admin | ISF 中的 OAuth2 Token 内省服务，用于验证 Bearer Token 有效性 |
| AUTH_ENABLED | 环境变量，控制认证功能开关，`true` 启用（默认），`false` 禁用 |
| Noop | No-operation 实现，满足接口契约但不执行任何实际操作 |
| 公网路由 | `/api/agent-retrieval/v1/*`，经过 Hydra Token 校验 |
| 内网路由 | `/api/agent-retrieval/in/v1/*`，通过 HTTP Header 传递身份信息，不依赖 ISF |

---

# 2. 整体设计（HLD）

## 2.1 系统上下文（C4 - Level 1）

### 参与者
- 用户：Agent 开发者、上游 Agent 服务
- 外部系统：BKN Backend、Ontology Query、Data Retrieval、Agent App、MF-Model API
- ISF 服务（可选）：Hydra Admin

### 系统关系

```
[Agent / 调用方]
    │
    ├── (公网) Bearer Token ──→ [Context-Loader] ──→ [Hydra Admin] (仅 AUTH_ENABLED=true)
    │                                │
    └── (内网) Header 身份 ──→ [Context-Loader]
                                     │
                                     ├──→ [BKN Backend]
                                     ├──→ [Ontology Query]
                                     ├──→ [Data Retrieval]
                                     ├──→ [Agent App]
                                     └──→ [MF-Model API]
```

AUTH_ENABLED=false 时，Hydra Admin 连线断开，Context-Loader 不建立与 Hydra 的连接。

---

## 2.2 容器架构（C4 - Level 2）

| 容器 | 技术栈 | 职责 |
|------|--------|------|
| agent-retrieval | Go / Gin | Context-Loader 主服务，提供知识检索、MCP 代理、行动召回等 API |
| Hydra Admin | ISF 组件 | OAuth2 Token 内省（AUTH_ENABLED=true 时使用） |

本次改造仅涉及 agent-retrieval 容器内部变更。

---

## 2.3 组件设计（C4 - Level 3）

### agent-retrieval 受影响组件

| 组件 | 职责 | 改造内容 |
|------|------|----------|
| `infra/config` | 配置加载 | 新增 `GetAuthEnabled()` 函数；移除 `UserMgnt` 字段 |
| `drivenadapters/hydra.go` | Hydra 客户端 | 新增 `noopHydra` 实现；`NewHydra()` 改为工厂模式 |
| `driveradapters/middleware.go` | 公网认证中间件 | 无需修改，中间件仍调用 `Hydra.Introspect()`，Noop 实现直接返回 |
| `driveradapters/rest_public_handler.go` | 公网路由注册 | 无需修改，仍调用 `drivenadapters.NewHydra()` |
| `drivenadapters/user_management.go` | UserManagement 客户端 | 整文件删除 |
| `interfaces/drivenadapters.go` | 接口定义 | 移除 `UserManagement` 接口及相关类型 |
| `mocks/drivenadapters.go` | Mock | 重新生成，移除 `MockUserManagement` |
| Helm Chart | 部署配置 | 新增 `auth.enabled`；移除 `depServices.user-management` |

---

## 2.4 数据流（Data Flow）

### AUTH_ENABLED=true（默认）主流程

```
HTTP 请求 (Bearer Token)
  → middlewareIntrospectVerify
    → hydra.Introspect(token)
      → POST hydra-admin:4445/admin/oauth2/introspect
    ← TokenInfo{Active: true, VisitorID: "user-001", ...}
  → 设置 AccountAuthContext 到 context
  → 业务 Handler
  → HTTP 响应
```

### AUTH_ENABLED=false 主流程

```
HTTP 请求 (无 Token)
  → middlewareIntrospectVerify
    → noopHydra.Introspect("")
    ← TokenInfo{Active: true} (空身份)
  → 设置空 AccountAuthContext 到 context
  → 业务 Handler
  → HTTP 响应
```

### 内网路由（不受影响）

```
HTTP 请求 (Header: x-account-id, x-account-type)
  → middlewareHeaderAuthContext
    → 从 Header 构造 AccountAuthContext
  → 业务 Handler
  → HTTP 响应
```

---

## 2.5 关键设计决策（Design Decisions）

| 决策 | 说明 |
|------|------|
| 使用 Noop 实现而非条件分支 | 遵循 ISSUE-177 统一方案，在 `NewHydra()` 工厂中决定返回哪个实现，中间件代码无需感知开关状态，符合开闭原则 |
| 环境变量而非配置文件控制 | 环境变量可通过 Helm values / Docker env 灵活注入，不需要修改配置 YAML 文件内容 |
| 默认 AUTH_ENABLED=true | 安全优先原则，防止因遗漏配置导致生产环境无认证保护 |
| 直接删除 UserManagement 而非保留 Noop | UserManagement 从未被业务代码调用，保留会产生维护负担；git 历史可随时恢复 |
| noopHydra 返回 Active=true 的空 TokenInfo | 确保中间件不会因 Active=false 而返回 401，同时下游业务逻辑可正常运行 |

---

## 2.6 部署架构（Deployment）

- 部署环境：K8s（与现有部署方式一致）
- 变更点：Helm values 新增 `auth.enabled`，deployment.yaml 新增 `AUTH_ENABLED` 环境变量注入
- 向后兼容：未设置 `auth.enabled` 时默认为 `true`，升级后行为不变

---

## 2.7 非功能设计

### 性能
- `AUTH_ENABLED=false` 时减少一次 Hydra HTTP 调用，公网请求延迟降低数毫秒

### 可用性
- 无状态服务，可用性由 K8s 平台保证

### 安全
- 默认启用认证，安全优先
- `AUTH_ENABLED=false` 时启动日志输出警告信息

### 可观测性
- 启动时记录 `AUTH_ENABLED` 的实际值到日志
- `AUTH_ENABLED=false` 时输出 WARN 级别日志：`ISF authentication disabled via AUTH_ENABLED env`

---

# 3. 详细设计（LLD）

## 3.1 配置层变更

### 3.1.1 新增 GetAuthEnabled()

文件：`server/infra/config/config.go`

```go
func GetAuthEnabled() bool {
	envVal := os.Getenv("AUTH_ENABLED")
	return envVal != "false" && envVal != "0"
}
```

该函数作为包级公开函数，不绑定到 Config 结构体，因为它在 Hydra 工厂初始化时需要被调用，且仅读取环境变量。

### 3.1.2 移除 UserMgnt 字段

文件：`server/infra/config/config.go`

从 `Config` 结构体中移除：
```go
// 移除此行
UserMgnt PrivateBaseConfig `yaml:"user_management"`
```

### 3.1.3 移除 YAML 中的 user_management 块

文件：`server/infra/config/agent-retrieval.yaml`

移除：
```yaml
# 移除此块
user_management:
  private_host: "user-management-private.anyshare"
  private_port: 30980
  private_protocol: "http"
```

---

## 3.2 Hydra 工厂模式改造

文件：`server/drivenadapters/hydra.go`

### NewHydra() 工厂

```go
func NewHydra() interfaces.Hydra {
	once.Do(func() {
		if !config.GetAuthEnabled() {
			logger := config.NewConfigLoader().GetLogger()
			logger.Warn("ISF authentication disabled via AUTH_ENABLED env, using noop hydra")
			h = &noopHydra{}
		} else {
			conf := config.NewConfigLoader()
			h = &hydra{
				adminAddress: conf.OAuth.BuildAdminURL(),
				logger:       conf.GetLogger(),
				httpClient:   rest.NewHTTPClient(),
			}
		}
	})
	return h
}
```

### noopHydra 实现

```go
type noopHydra struct{}

func (n *noopHydra) Introspect(_ context.Context, _ string) (*interfaces.TokenInfo, error) {
	return &interfaces.TokenInfo{
		Active:     true,
		VisitorTyp: interfaces.Anonymous,
	}, nil
}
```

Noop 实现返回 `Active: true` 的 `TokenInfo`，`VisitorTyp` 设为 `Anonymous`，确保：
- 中间件不会因 `Active=false` 返回 401
- `VisitorID` 为空字符串，下游代码处理此情况
- `VisitorTyp` 为 `Anonymous`，与未认证语义一致

---

## 3.3 UserManagement 代码移除

### 删除文件

| 文件 | 说明 |
|------|------|
| `server/drivenadapters/user_management.go` | 客户端实现 |
| `server/drivenadapters/user_management_test.go` | 单元测试 |

### 修改文件

| 文件 | 变更 |
|------|------|
| `server/interfaces/drivenadapters.go` | 移除 `UserManagement` 接口、`UserInfo`、`AppInfo`、`ErrorResponse` 类型、`DisplayName` 常量 |
| `server/mocks/drivenadapters.go` | 重新执行 `go generate` 以移除 `MockUserManagement` |
| `server/infra/config/config.go` | 移除 `UserMgnt` 字段 |
| `server/infra/config/agent-retrieval.yaml` | 移除 `user_management` 配置块 |
| `helm/agent-retrieval/values.yaml` | 移除 `depServices.user-management` 块 |
| `helm/agent-retrieval/templates/configmap.yaml` | 移除 `user_management` 渲染段 |

---

## 3.4 Helm Chart 变更

### values.yaml 新增

```yaml
# 认证开关
auth:
  enabled: true
```

### values.yaml 移除

```yaml
# 移除此块
depServices:
  user-management:
    privateHost: user-management-private
    privatePort: 30980
    privateProtocol: http
```

### deployment.yaml 新增环境变量

在 `env:` 段中追加：

```yaml
{{- if hasKey .Values "auth" }}
  - name: AUTH_ENABLED
    value: {{ .Values.auth.enabled | quote }}
{{- end }}
```

### configmap.yaml 移除

移除 `user_management` 段（第 34-37 行）：

```yaml
# 移除此段
user_management:
  private_protocol: {{ index .Values "depServices" "user-management" "privateProtocol" | quote }}
  private_host: {{ index .Values "depServices" "user-management" "privateHost" | quote }}
  private_port: {{ index .Values "depServices" "user-management" "privatePort" }}
```

---

## 3.5 核心流程（详细）

### 服务启动流程

1. `main.go` → `config.NewConfigLoader()`：加载 YAML + 环境变量覆盖
2. `NewRestPublicHandler()` → `drivenadapters.NewHydra()`
3. `NewHydra()` 内部调用 `config.GetAuthEnabled()`
   - `true` → 创建 `hydra{}` 实例，读取 `config.OAuth.BuildAdminURL()` 构建 Hydra Admin 地址
   - `false` → 创建 `noopHydra{}` 实例，输出 WARN 日志
4. 公网路由注册 `middlewareIntrospectVerify(hydra)`，中间件持有的 Hydra 实例已确定
5. 服务启动完毕，进入请求处理循环

### 公网请求处理流程（AUTH_ENABLED=false）

1. 收到 HTTP 请求
2. `middlewareIntrospectVerify` 调用 `getToken(c)` 获取 Token（可能为空字符串）
3. 调用 `noopHydra.Introspect(ctx, "")` → 直接返回 `TokenInfo{Active: true, VisitorTyp: Anonymous}`
4. 中间件构造 `AccountAuthContext{AccountID: "", AccountType: "anonymous"}`
5. 设置到 context，`c.Next()` 进入业务 Handler
6. 业务 Handler 正常执行并返回响应

---

## 3.6 错误处理

| 场景 | 处理方式 |
|------|----------|
| AUTH_ENABLED=false 但上游仍传了 Token | noopHydra 忽略 Token，直接返回空身份，不报错 |
| AUTH_ENABLED=true 且 Hydra 不可达 | 行为与改造前一致，返回 HTTP 500/502 错误 |
| AUTH_ENABLED 值为非法字符串（如 "yes"） | `GetAuthEnabled()` 返回 true（安全优先），认证启用 |

---

## 3.7 配置设计

| 配置项 | 来源 | 默认值 | 说明 |
|--------|------|--------|------|
| AUTH_ENABLED | 环境变量 | true | 认证开关，`false` 或 `0` 时禁用 |
| auth.enabled | Helm values | true | 通过 deployment.yaml 注入为 AUTH_ENABLED 环境变量 |

---

## 3.8 可观测性实现

- logging：
  - 启动时输出 `AUTH_ENABLED` 值：`logger.Infof("AUTH_ENABLED=%v", config.GetAuthEnabled())`
  - `AUTH_ENABLED=false` 时输出 WARN：`ISF authentication disabled via AUTH_ENABLED env, using noop hydra`

- tracing：
  - `noopHydra.Introspect` 不生成任何 span，减少无意义的 trace 数据

- metrics：
  - 不涉及

---

# 4. 风险与权衡（Risks & Trade-offs）

| 风险 | 影响 | 解决方案 |
|------|------|----------|
| 生产环境误设 AUTH_ENABLED=false | 公网 API 完全无认证保护 | 默认值为 true；启动日志输出醒目 WARN；文档强调仅限开发/测试使用 |
| 移除 UserManagement 后未来需要重新引入 | 需重新实现客户端代码 | git 历史保留完整实现；接口和实现模式在 ISSUE-177 文档中有详细说明 |
| noopHydra 返回空身份导致下游逻辑异常 | 部分业务逻辑依赖 AccountID 做数据过滤 | AccountID 为空时，下游应返回无过滤的全量数据；需验证各 handler 对空 AccountID 的兼容性 |

---

# 5. 测试策略（Testing Strategy）

- 单元测试：
  - `config.GetAuthEnabled()` 的各种输入组合测试（未设置、"true"、"false"、"0"、"1"、非法值）
  - `noopHydra.Introspect()` 返回值验证
  - `NewHydra()` 工厂在不同 AUTH_ENABLED 值下返回正确实现类型
- 集成测试：
  - AUTH_ENABLED=false 时公网 API 无 Token 调用成功
  - AUTH_ENABLED=true 时无 Token 调用返回 401
- 编译验证：
  - 移除 UserManagement 后 `go build ./...` 通过
  - `go generate ./...` 重新生成 Mock 后 `go test ./...` 通过

---

# 6. 发布与回滚（Release Plan）

### 发布步骤
1. 合入代码变更（Hydra 工厂模式 + UserManagement 移除 + Helm 变更）
2. CI 验证编译和测试通过
3. 构建新镜像
4. 部署时 Helm values 无需特殊配置（auth.enabled 默认 true，向后兼容）

### 回滚方案
- Helm rollback 到上一版本即可恢复
- 回滚后 `AUTH_ENABLED` 环境变量会被忽略（旧版代码不读取该变量）

---

# 7. 附录（Appendix）

## 变更文件清单

| 操作 | 文件 |
|------|------|
| 修改 | `server/infra/config/config.go` |
| 修改 | `server/infra/config/agent-retrieval.yaml` |
| 修改 | `server/drivenadapters/hydra.go` |
| 修改 | `server/interfaces/drivenadapters.go` |
| 修改 | `server/mocks/drivenadapters.go`（重新生成） |
| 修改 | `helm/agent-retrieval/values.yaml` |
| 修改 | `helm/agent-retrieval/templates/deployment.yaml` |
| 修改 | `helm/agent-retrieval/templates/configmap.yaml` |
| 删除 | `server/drivenadapters/user_management.go` |
| 删除 | `server/drivenadapters/user_management_test.go` |

共涉及 10 个文件（8 修改 + 2 删除）。

## 相关文档
- PRD: ../prd/issue-250-contextloader-isf-decouple-prd.md
- 总体逻辑设计: ../prd/[1]+ISSUE-177+逻辑设计.doc
- [GitHub Issue #250](https://github.com/kweaver-ai/adp/issues/250)

## 参考资料
- [ISSUE-177 设计模式：AUTH_ENABLED + 工厂 + Noop](../prd/[1]+ISSUE-177+逻辑设计.doc)
