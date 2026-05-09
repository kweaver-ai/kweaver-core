# 🏗️ Design Doc: 执行工厂支持 ContextLoader 依赖工具包内部导入

> 状态: Draft  
> 负责人: 待确认  
> Reviewers: 待确认  
> 关联设计: ../../../../context-loader/agent-retrieval/docs/design/issue-243-contextloader-tool-deps-auto-import-design.md  

---

# 📌 1. 概述（Overview）

## 1.1 背景

- 当前现状：
  - 执行工厂已有公共 `impex/import/:type` 导入接口，可消费 `.adp` 组件快照。
  - 执行工厂已有 `tool-box/intcomp` 内置工具注册接口，但该接口只支持 `openapi/function` 元数据，不支持 `.adp`。
  - `context-loader` 希望在服务启动时自动同步依赖工具包 `execution_factory_tools.adp`，不再依赖人工导入。

- 存在问题：
  - 公共 `impex/import/:type` 主要服务于用户显式导入，不适合作为服务间自动同步的内部入口。
  - 若将 `.adp` 直接接入 `tool-box/intcomp`，会把“元数据注册”和“配置包导入”两类语义混在一起。
  - 现有公共人工导入接口没有 `ImportConfig` 相关配置，无法复用一套复杂的来源仲裁模型。

- 业务 / 技术背景：
  - `.adp` 本质是 `ComponentImpexConfigModel` 序列化结果，天然属于导入语义。
  - 当前交付包与执行工厂中的 toolbox 均具备可识别的版本信息，因此可以直接基于版本做更新判定。

---

## 1.2 目标

- 为 `context-loader` 提供一条面向内部服务的依赖工具包导入接口。
- 复用现有 `.adp` 导入逻辑，不新增第二套 toolbox 导入实现。
- 在内部导入接口中增加“当前系统版本 < 待导入版本才更新”的判定逻辑。
- 明确该接口与公共 `impex import`、`tool-box/intcomp` 的职责边界。

---

## 1.3 非目标（Out of Scope）

- 不修改公共 `impex/import/:type` 的用户接口契约。
- 不扩展 `CreateInternalToolBox` 以支持 `.adp`。
- 不引入新的托管/脱管/保护锁仲裁状态。

---

## 1.4 术语说明（Optional）

| 术语 | 说明 |
|------|------|
| 内部导入接口 | 面向内部服务的 `.adp` 自动同步入口 |
| 公共导入接口 | 现有 `/impex/import/:type`，面向用户显式导入 |
| 内置工具注册 | 现有 `/tool-box/intcomp`，面向 openapi/function 元数据注册 |
| 当前系统版本 | 执行工厂中已存在资源的版本 |
| 待导入版本 | 当前 `.adp` 中携带的版本 |

---

# 🏗️ 2. 整体设计（HLD）

## 🌍 2.1 系统上下文（C4 - Level 1）

### 参与者
- 调用方：`context-loader/agent-retrieval`
- 系统：执行工厂 `operator-integration`

### 系统关系

    ContextLoader
      → OperatorIntegration Internal Impex API
      → ImpexManager / Toolbox Query / Toolbox Import

---

## 🧩 2.2 组件设计（C4 - Level 3）

| 组件 | 职责 |
|------|------|
| Internal Impex Route | 提供内部 `.adp` 导入入口，底层复用公共导入逻辑 |
| ComponentImpexManager | 复用现有 `impex` 解析与事务导入逻辑 |
| Toolbox Query Service | 查询当前资源是否存在以及当前版本 |
| Toolbox Import Service | 在需要更新时复用既有 toolbox/operator 依赖导入逻辑 |

---

## ⚖️ 2.3 关键设计决策（Design Decisions）

| 决策 | 说明 |
|------|------|
| 新增内部导入接口，不改公共导入接口 | 将服务间自动同步与用户人工导入隔离 |
| 不改造 `tool-box/intcomp` 支持 `.adp` | 保持接口职责单一 |
| 复用现有 `impex import` | 降低实现风险，避免重复解析和重复事务逻辑 |
| 用版本驱动单向升级替代复杂仲裁模型 | 既兼容人工导入，又不要求公共导入写额外状态 |

---

# 🔧 3. 详细设计（LLD）

## 🌐 3.1 接口设计

**Endpoint:** `POST /api/agent-operator-integration/internal-v1/impex/intcomp/import/:type`

**当前支持：**

- `:type = toolbox`

**请求字段：**

| 字段 | 说明 |
|------|------|
| `data` | `.adp` 文件内容 |
| `mode` | 默认 `upsert` |
| `package_version` | 由 `context-loader` 从镜像内固定 `VERSION` 文件读取后传入 |

**返回状态：**

- `imported`
- `updated`
- `skipped`

---

## 🧠 3.2 接口处理流程

### 主流程

1. 解析 `multipart/form-data`
2. 校验 `type`
3. 读取并反序列化 `.adp`
4. 解析包内 toolbox 的待导入版本
5. 查询执行工厂当前已有资源及其当前版本
6. 若 `current_version < package_version`，则复用现有 `impex import` 执行导入
7. 若 `current_version >= package_version`，则直接返回 `skipped`

说明：

- `execution_factory_tools.adp` 路径不由调用方配置，属于 `context-loader` 镜像固定交付内容
- `package_version` 也不由部署配置决定，而是由 `context-loader` 启动时读取镜像内 `VERSION` 文件并传入

### 复用点

- `.adp` 解析：`ComponentImpexConfigModel`
- toolbox 导入：`toolbox.Import()`
- 事务：`importConfigWithTx()`

---

## 🔁 3.3 版本判定规则

### 规则

对包内每个 toolbox 资源按以下规则处理：

1. 不存在：导入，返回 `imported`
2. 已存在且 `current_version < package_version`：更新，返回 `updated`
3. 已存在且 `current_version == package_version`：跳过，返回 `skipped`
4. 已存在且 `current_version > package_version`：跳过，返回 `skipped`

### 设计原因

- 不要求公共人工导入接口感知额外仲裁状态
- 能避免自动任务覆盖更高版本的人工导入结果
- 规则简单，易于测试、运维和解释

---

## 🚫 3.4 与其他接口的边界

### 与公共 `impex/import/:type` 的边界

- 公共接口：
  - 面向用户显式导入
  - 按用户权限和业务域进行导入
  - 不承担服务间自动同步语义

- 内部接口：
  - 面向服务间自动同步
  - 接收待导入版本
  - 返回更明确的自动同步状态

### 与 `tool-box/intcomp` 的边界

- `tool-box/intcomp`：
  - 面向单个内置工具箱元数据注册
  - 输入是 `openapi/function`
  - 核心职责是“注册 / 升级元数据”

- 内部导入接口：
  - 面向 `.adp` 快照导入
  - 输入是组件配置包
  - 核心职责是“导入 / 版本驱动更新依赖包”

---

## 🧪 3.5 测试要点

- `.adp` 文件解析成功与失败
- `type=toolbox` 的正常导入
- 首次导入返回 `imported`
- 新版本自动更新返回 `updated`
- 同版本重复导入返回 `skipped`
- 当前系统版本更高时返回 `skipped`

---

# ✅ 4. 验收标准（DoD）

- [ ] 执行工厂新增内部 `.adp` 导入接口
- [ ] 内部接口底层复用现有 `impex import` 和 toolbox 导入逻辑
- [ ] 自动同步场景支持 `imported / updated / skipped` 状态
- [ ] `tool-box/intcomp` 不被改造成 `.adp` 入口
- [ ] 当前系统版本大于或等于待导入版本时，内部导入接口直接跳过
