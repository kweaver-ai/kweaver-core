# Context-Loader 与 ISF 解耦

关联 Issue: #250  
负责人: @criller  
状态: Draft  

---

## 1. 背景（Why）

- 当前问题是什么？
  - Context-Loader (agent-retrieval) 的公网路由 `/api/agent-retrieval/v1/*` 通过 `middlewareIntrospectVerify` 中间件硬依赖 ISF 平台的 Hydra Admin 服务进行 OAuth2 Token 内省校验。
  - 这导致微服务无法脱离 ISF 平台进行独立部署、本地开发调试和单元/集成测试。
  - 此外，代码中存在已实现但从未被业务逻辑调用的 UserManagement 客户端代码及相关配置，属于无效依赖，增加维护成本和理解负担。
- 为什么要做？
  - 需要让 Context-Loader 在无 ISF 环境下也能正常启动和运行，降低开发环境搭建成本。
  - 需要清理未使用的死代码，保持代码库整洁。
- 来源
  - GitHub Issue `#250`，属于 ISSUE-177 "Vega/Ontology 系列微服务与 ISF 解耦"的子任务，在 Context-Loader 上落地。

---

## 2. 目标（Goal）

- 本次要解决什么问题？
  - 通过 `AUTH_ENABLED` 环境变量一键切换 Context-Loader 的认证功能开关。
  - 禁用认证时跳过 Hydra Token 内省校验，使公网路由可在无 ISF 环境下正常访问。
  - 移除从未被业务逻辑使用的 UserManagement 客户端代码、接口定义、Mock 和相关配置项。
- 成功的标准是什么？
  - Given `AUTH_ENABLED=true`（或未设置），When 调用公网 API，Then 行为与改造前完全一致，Hydra Token 校验正常执行。
  - Given `AUTH_ENABLED=false`，When 调用公网 API 不携带 Token，Then 请求可正常到达业务 handler 并返回正确响应。
  - Given `AUTH_ENABLED=false`，When 服务启动，Then 不连接 Hydra Admin 服务，启动无报错。
  - Given 代码改造完成，When 搜索 UserManagement 相关代码，Then 仅在 git 历史中可见，当前代码库中已完全移除。

---

## 3. 用户/使用场景（Who & When）

- 谁会用？
  - 开发者：本地开发调试时使用 `AUTH_ENABLED=false` 快速启动服务。
  - 测试工程师：CI 流水线和单元测试中无需部署 ISF 栈。
  - 运维人员：轻量化交付场景下通过 Helm values 关闭认证，精简部署依赖。
- 在什么场景使用？
  - 本地开发环境无 ISF 服务时，设置 `AUTH_ENABLED=false` 即可启动 Context-Loader 并调用全部 API 进行功能验证。
  - 编写和执行自动化测试时，无需模拟或部署 Hydra 服务。
  - 在不需要多租户或权限管控的场景中精简部署，减少对 Hydra 等 ISF 组件的资源消耗。

---

## 4. 功能需求（What）

### 4.1 AUTH_ENABLED 环境变量控制

- 描述：新增 `AUTH_ENABLED` 环境变量，控制认证功能的全局开关。
- 输入：环境变量 `AUTH_ENABLED`，值为 `"true"` / `"false"` / `"0"` / `"1"` 或未设置。
- 输出：布尔值，决定是否启用 Hydra Token 内省校验。
- 规则：
  - 默认值为 `true`（安全优先），环境变量未设置时认证启用。
  - 仅当显式设为 `"false"` 或 `"0"` 时返回 `false`。
  - 该值在服务启动时读取一次，运行期间不变。

### 4.2 公网路由认证中间件适配

- 描述：`AUTH_ENABLED=false` 时，公网路由跳过 Hydra Token 内省，直接放行请求。
- 输入：HTTP 请求（可能不携带 Token）。
- 输出：请求上下文中设置匿名用户身份的 `AccountAuthContext`（`AccountID` 为空，`AccountType` 为 `anonymous`），业务 handler 正常执行。
- 规则：
  - `AUTH_ENABLED=true` 时行为与改造前完全一致。
  - `AUTH_ENABLED=false` 时不实例化 Hydra HTTP 客户端，不连接 Hydra Admin 服务。
  - 禁用时使用 Noop 实现替代 Hydra，直接返回 `Active=true` 的匿名用户 Token 信息，中间件据此构造匿名身份上下文。
  - 禁用时接口参数和返回结构保持不变，仅跳过安全校验。
  - 内网路由 `/api/agent-retrieval/in/v1/*` 不受影响（它本来就不使用 Hydra）。

### 4.3 Helm Chart 配置支持

- 描述：Helm `values.yaml` 新增 `auth.enabled` 配置项，通过 `deployment.yaml` 注入 `AUTH_ENABLED` 环境变量。
- 输入：Helm values 中 `auth.enabled` 的值。
- 输出：容器环境变量 `AUTH_ENABLED`。
- 规则：
  - 默认值为 `true`，向后兼容。
  - `AUTH_ENABLED=false` 时，`depServices.hydra` 配置块可省略。

### 4.4 移除未使用的 UserManagement 代码

- 描述：清理从未被业务逻辑调用的 UserManagement 相关代码和配置。
- 规则：
  - 移除 `server/drivenadapters/user_management.go` 及其测试文件。
  - 移除 `server/interfaces/drivenadapters.go` 中 `UserManagement` 接口定义及相关类型（`UserInfo`、`AppInfo`、`ErrorResponse`）。
  - 移除 `server/mocks/drivenadapters.go` 中 `MockUserManagement` 相关代码。
  - 移除 `server/infra/config/config.go` 中 `UserMgnt` 配置字段。
  - 移除 `server/infra/config/agent-retrieval.yaml` 中 `user_management` 配置块。
  - 移除 `helm/agent-retrieval/values.yaml` 中 `depServices.user-management` 配置块。
  - 移除 `helm/agent-retrieval/templates/configmap.yaml` 中 `user_management` 渲染部分。

---

## 5. 非功能需求（Non-Functional）

- 性能：无新增开销。`AUTH_ENABLED=false` 时减少一次 HTTP 调用（Hydra introspect），公网路由延迟降低。
- 可用性：无状态服务，可用性由 K8s 平台保证。
- 安全性：
  - 默认 `AUTH_ENABLED=true`，安全优先。
  - 文档和日志中明确提示禁用认证的风险，仅建议在开发/测试/轻量化交付场景使用。
- 兼容性：
  - 未设置 `AUTH_ENABLED` 时行为与改造前完全一致（向后兼容）。
  - API 接口签名不变，认证禁用对调用方透明。

---

## 6. 验收标准（DoD - Definition of Done）

- [ ] `AUTH_ENABLED` 未设置或设为 `true` 时，公网 API 行为与改造前一致，Hydra Token 校验正常工作
- [ ] `AUTH_ENABLED=false` 时，公网 API 无需 Token 即可正常调用并获取正确响应
- [ ] `AUTH_ENABLED=false` 时，服务启动不连接 Hydra Admin，启动日志包含明确的禁用提示
- [ ] 内网 API (`/api/agent-retrieval/in/v1/*`) 在两种模式下行为均不受影响
- [ ] Helm Chart `values.yaml` 新增 `auth.enabled` 配置项，默认值为 `true`
- [ ] Helm `deployment.yaml` 正确注入 `AUTH_ENABLED` 环境变量
- [ ] UserManagement 相关代码（实现、接口、Mock、测试）和配置（YAML、Helm）已完全移除
- [ ] 移除后项目编译通过，现有测试全部通过
- [ ] 有覆盖 `AUTH_ENABLED=true` 和 `AUTH_ENABLED=false` 两种路径的单元测试

---

## 7. 不做范围（Out of Scope）

- 不引入 PermissionService 抽象（Context-Loader 不使用 authorization-private 服务）
- 不改造内网路由的认证逻辑（内网 Header 认证不依赖 ISF）
- 不改造 MCP 协议或 MCP Handler 本身
- 不引入运行时动态切换 `AUTH_ENABLED` 的能力（重启生效即可）
- 不在本期处理其他微服务的 ISF 解耦

---

## 8. 风险 & 依赖

- 外部依赖：
  - 遵循 ISSUE-177 总体设计方案中的 `AUTH_ENABLED` 环境变量命名约定和行为规范。
- 潜在风险：
  - `AUTH_ENABLED=false` 在生产环境被误设会导致公网 API 完全无认证保护。缓解措施：默认值为 `true`，启动日志中输出醒目的安全警告。
  - 移除 UserManagement 代码后，若后续有新功能需要调用用户管理服务，需重新实现。缓解措施：git 历史中保留完整代码，可快速恢复。

---

## 9. 参考资料

- [GitHub Issue #250 Context-Loader与ISF解耦](https://github.com/kweaver-ai/adp/issues/250)
- [GitHub Issue #177 总体逻辑设计文档](docs/prd/[1]+ISSUE-177+逻辑设计.doc)
- ISSUE-177 设计模式：`AUTH_ENABLED` 环境变量 + 工厂模式 + Noop 实现
