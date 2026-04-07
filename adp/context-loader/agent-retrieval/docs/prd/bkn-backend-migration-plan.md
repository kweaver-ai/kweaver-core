## 背景

`context-loader/agent-retrieval` 目前通过名为 `ontology-manager` 的后端服务获取知识网络、本体对象类、关系类、行动类等信息。  
在 BKN 体系中，该服务已经统一为 `bkn-backend`，实际运行时：

- **服务名（host）**：由 `ontology-manager-svc` 调整为 `bkn-backend-svc`
- **API 前缀**：由 `/api/ontology-manager` 调整为 `/api/bkn-backend`

目标是：在 **不修改任何既有业务/PRD/设计文档内容** 的前提下，将 `agent-retrieval` 内部所有与 `ontology-manager` 相关的代码和配置，统一迁移到 `bkn-backend`，避免运行时仍依赖旧服务。

---

## 范围说明

本方案仅覆盖 `context-loader/agent-retrieval` 模块内：

- **包含**：
  - Go 代码中访问本体服务的客户端实现
  - Helm 配置（values / ConfigMap 模板）
  - 本地运行配置（`agent-retrieval.yaml`）
  - 开发辅助 HTTP 示例
- **不包含**：
  - 任意 PRD / 设计 / 需求文档内容修改
  - 其它子项目（如 `bkn-backend` 自身、`ontology-query` 等）的代码或文档
  - BKN 概念层术语（例如 “BKN Engine = ontology-manager + ontology-query”）的重写

---

## 影响面梳理（仅 agent-retrieval）

### 1. 运行时代码

- 文件：`context-loader/agent-retrieval/server/drivenadapters/ontology_manager.go`
- 当前行为：
  - 使用配置对象 `conf.OntologyManager` 构造基础 URL：
    ```go
    baseURL: conf.OntologyManager.BuildURL("/api/ontology-manager")
    ```
  - 后续所有下游调用（知识网络、对象类、关系类、行动类、ontology jobs）都基于 `baseURL` 拼接 `/in/v1/knowledge-networks/...` 等路径。
- 问题：
  - 当集群仅部署 `bkn-backend`，且网关只暴露 `/api/bkn-backend` 时，这些调用会直接 404 或指向已经下线的服务。

### 2. 部署配置（Helm）

- 文件：`context-loader/agent-retrieval/helm/agent-retrieval/values.yaml`
  - 片段：
    ```yaml
    depServices:
      ontology-manager:
        privateHost: ontology-manager-svc
        privatePort: 13014
        privateProtocol: http
    ```

- 文件：`context-loader/agent-retrieval/helm/agent-retrieval/templates/configmap.yaml`
  - 将上述值注入到运行时配置：
    ```yaml
    ontology_manager:
      private_protocol: {{ index .Values "depServices" "ontology-manager" "privateProtocol" | quote }}
      private_host: {{ index .Values "depServices" "ontology-manager" "privateHost" | quote }}
      private_port: {{ index .Values "depServices" "ontology-manager" "privatePort" }}
    ```

### 3. 本地/非 Helm 配置

- 文件：`context-loader/agent-retrieval/server/infra/config/agent-retrieval.yaml`
  - 片段：
    ```yaml
    ontology_manager:
      private_protocol: "http"
      private_host: "ontology-manager-svc.anyshare"
      private_port: 13014
    ```

### 4. 开发辅助 HTTP 示例

- 文件：`context-loader/agent-retrieval/server/tests/http/ontology_manager.http`
  - 示例 URL：
    ```text
    http://ontology-manager-svc.anyshare:13014/api/ontology-manager/in/v1/knowledge-networks/...
    ```

> 以上 4 个维度为**实际运行时和开发依赖**，本方案仅修改这些位置。  
> 所有 `docs/prd/...` 以及其它文档中的 `ontology-manager` 内容**保持不变**。

---

## 调整目标

1. **所有运行时调用**统一通过 `bkn-backend` 访问本体相关接口：
   - host 指向：`bkn-backend-svc`（或 `bkn-backend-svc.anyshare`）
   - API 前缀：`/api/bkn-backend`
2. `agent-retrieval` 的配置与代码中，不再有“仍然调用旧的 ontology-manager 服务”的实际行为。
3. 文档内容保持现状，仅在代码与配置层面体现服务已经是 `bkn-backend`。

---

## 详细调整方案

### 步骤 1：更新本体客户端 API 前缀（代码）

- 文件：`server/drivenadapters/ontology_manager.go`
- 变更点：
  - 将基础 URL 前缀从：
    ```go
    baseURL: conf.OntologyManager.BuildURL("/api/ontology-manager")
    ```
  - 修改为：
    ```go
    baseURL: conf.OntologyManager.BuildURL("/api/bkn-backend")
    ```
- 说明：
  - 保持 `conf.OntologyManager` 配置结构不变，仅修改其用于构造 URL 时的路径前缀。
  - 假定 `bkn-backend` 在 `/api/bkn-backend` 路径下已经提供与原 `/api/ontology-manager` 等价的 `/in/v1/knowledge-networks/...` 路由。

### 步骤 2：更新集群部署的服务 host（Helm）

- 文件：`helm/agent-retrieval/values.yaml`
  - 将：
    ```yaml
    depServices:
      ontology-manager:
        privateHost: ontology-manager-svc
        privatePort: 13014
        privateProtocol: http
    ```
  - 修改为：
    ```yaml
    depServices:
      ontology-manager:
        privateHost: bkn-backend-svc
        privatePort: 13014
        privateProtocol: http
    ```
- 文件：`helm/agent-retrieval/templates/configmap.yaml`
  - 不修改字段名与模板结构，继续使用 `ontology_manager` 段和 `depServices.ontology-manager` key，只依赖 `values.yaml` 中的 host 变化：
    ```yaml
    ontology_manager:
      private_host: {{ index .Values "depServices" "ontology-manager" "privateHost" | quote }}
    ```
- 结果：
  - 渲染后的运行时配置中，`ontology_manager.private_host = bkn-backend-svc`。

### 步骤 3：更新本地运行配置的服务 host

- 文件：`server/infra/config/agent-retrieval.yaml`
  - 将：
    ```yaml
    ontology_manager:
      private_protocol: "http"
      private_host: "ontology-manager-svc.anyshare"
      private_port: 13014
    ```
  - 修改为：
    ```yaml
    ontology_manager:
      private_protocol: "http"
      private_host: "bkn-backend-svc.anyshare"
      private_port: 13014
    ```
- 说明：
  - 保持 `ontology_manager` 这个 YAML 段名不变，仅调整其中的 host 值。
  - 本地开发或直接用该文件启动服务时，将自动指向新的 `bkn-backend` 服务。

### 步骤 4：更新 HTTP 测试示例

- 文件：`server/tests/http/ontology_manager.http`
  - 将 URL 中的 host 和 path 从：
    ```text
    http://ontology-manager-svc.anyshare:13014/api/ontology-manager/in/v1/knowledge-networks/...
    ```
  - 修改为：
    ```text
    http://bkn-backend-svc.anyshare:13014/api/bkn-backend/in/v1/knowledge-networks/...
    ```
- 说明：
  - 仅影响开发调试示例，不影响运行时。更新后可用于直接验证新链路。

---

## 验证方案

### 1. 单服务验证（本地或测试环境）

1. 启动或部署更新后的 `bkn-backend`，确认其路由为：
   - `/api/bkn-backend/in/v1/knowledge-networks/...`
2. 使用更新后的配置启动 `agent-retrieval`：
   - 本地：使用更新后的 `server/infra/config/agent-retrieval.yaml`
   - 集群：使用更新的 Helm `values.yaml` 和模板渲染
3. 通过 `server/tests/http/ontology_manager.http` 或上游接口触发一次典型调用（例如获取对象类定义）。
4. 预期结果：
   - 日志中请求 URL 指向 `bkn-backend-svc` + `/api/bkn-backend/...`
   - 不再出现访问 `ontology-manager-svc` 或 `/api/ontology-manager` 的记录。
   - 接口返回成功的业务响应。

### 2. 集成链路验证

1. 在集群环境中部署：
   - 更新后的 `bkn-backend`
   - 更新后的 `agent-retrieval`
2. 从上游（如 Context Loader 工具、Agent 调用链）触发完整使用场景：
   - 逻辑属性解析
   - 基于对象类/关系类的知识网络查询
3. 预期结果：
   - 上游请求成功返回，无 404 / service not found。
   - 监控与日志中，所有本体相关调用都指向 `bkn-backend-svc` 和 `/api/bkn-backend`。

---

## 命名与结构重构方案（一次性完成）

在完成前述运行时切换（host + path）后，本方案进一步一次性完成代码层面的命名重构，使得 `agent-retrieval` 内部不再出现以 `ontology-manager` 命名的类型/方法/文件，统一为 `bkn-backend` 语义。

### 1. 客户端文件与类型重命名

- 文件重命名：
  - `server/drivenadapters/ontology_manager.go` → `server/drivenadapters/bkn_backend.go`
- 类型与构造函数重命名：
  - 结构体：
    - `type ontologyManagerAccess struct { ... }` → `type bknBackendAccess struct { ... }`
  - 构造函数：
    - `func NewOntologyManagerAccess() interfaces.OntologyManagerAccess`  
      → `func NewBknBackendAccess() interfaces.BknBackendAccess`
  - 方法接收者：
    - 所有以 `func (oma *ontologyManagerAccess) ...` 定义的方法，统一改为 `func (b *bknBackendAccess) ...`（或类似短名），方法名本身可复用（如 `GetKnowledgeNetworkDetail` 等保持不变）。

> 为避免大面积编译错误，推荐按照顺序执行：先在 `interfaces` 中引入新接口类型（见下一节），然后逐步替换实现与调用。

### 2. 接口定义与依赖注入重命名

- 在 `server/interfaces` 包中：
  - 将现有 `OntologyManagerAccess` 接口重命名为 `BknBackendAccess`：
    - 接口方法列表保持不变，只改接口名。
  - 若需要兼容旧调用方，可短期保留：
    ```go
    type OntologyManagerAccess = BknBackendAccess
    ```
    作为 type alias，后续可在一次清理中移除。

- 在所有依赖注入 / 使用点：
  - 例如 handler、service、use case 中的字段：
    - `omAccess interfaces.OntologyManagerAccess` → `bknAccess interfaces.BknBackendAccess`
  - 构造函数中：
    - `NewOntologyManagerAccess()` → `NewBknBackendAccess()`

### 3. 配置结构与字段重命名（代码层面）

- 在配置定义处（示意）：
  - 结构体：
    - `type OntologyManagerConfig struct { ... }` → `type BknBackendConfig struct { ... }`
  - 顶层配置：
    - `Config.OntologyManager BknBackendConfig` → `Config.BknBackend BknBackendConfig`

- 在配置加载逻辑中（YAML/环境变量映射）：
  - 代码内部使用 `BknBackend` 命名，例如：
    ```go
    type Config struct {
        BknBackend BknBackendConfig `yaml:"bkn_backend" mapstructure:"bkn_backend"`
    }
    ```
  - 为兼容现有 YAML 中的 `ontology_manager` 段，可以在加载时增加兼容逻辑：
    - 优先从 `bkn_backend` 读取；
    - 若不存在，则回退到 `ontology_manager`，并记录一条 deprecation 日志。

> 配置文件本身（YAML key）是否立即从 `ontology_manager` 改为 `bkn_backend`，可以根据团队发布节奏决定；本方案建议代码层面先统一为 `BknBackend`，YAML key 支持双读一段时间，再通过后续版本移除旧 key。

### 4. 变量命名与内部引用统一

- 在客户端实现与调用处，将所有与本体服务相关的变量从 `om` / `ontologyManager` 命名统一为 `bkn` / `bknBackend`，例如：
  - `omAccess` → `bknAccess`
  - `omCfg` → `bknCfg`
- 仅更改变量名与类型名，不改变逻辑分支与请求参数字段。

### 5. 验证与回归

- 保持前文“验证方案”步骤不变，额外增加：
  - 确认所有引用 `NewOntologyManagerAccess` / `OntologyManagerAccess` / `OntologyManagerConfig` 的代码均已替换为 `BknBackend` 版本；
  - 使用 `rg "ontology-manager" context-loader/agent-retrieval/server` 等命令确认 server 代码中不再出现以 `ontology-manager` 作为类型/变量/方法名的标识符（仅保留在必要的字符串、日志、或兼容代码中）。

---

## 不在本次范围内的事项（明确说明）

- **不改动任何业务/PRD/设计文档内容**：
  - 包括 `docs/prd/feature-799460/*.md`、`docs/prd/feature-799460/99-依赖接口/ontology-manager-object-type.json` 等。
  - 文档中出现的 `ontology-manager` 术语、架构图、错误码说明均保持原样，作为历史背景与概念描述。

---

## 总结

- 本方案在完成运行时切换（host + API 前缀）基础上，一次性完成 `agent-retrieval` 内部围绕本体服务的命名与结构重构，将所有代码层面依赖统一为 `bkn-backend` 语义。
- 通过类型/接口/配置/变量的系统重命名，降低长期维护成本，避免团队成员继续以 `ontology-manager` 名称理解当前的 BKN Backend。
- 所有业务/PRD/设计文档保持不变，仅作为背景资料存在；代码与配置则以 `bkn-backend` 为唯一权威命名。

