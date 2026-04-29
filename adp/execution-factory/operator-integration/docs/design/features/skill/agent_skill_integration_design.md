# 🏗️ Design Doc: 执行工厂 Skill 接入与管理

> 状态: In Review
> 负责人: 待确认
> Reviewers: 待确认
> 关联 PRD: ../../product/prd/agent_skill接入与管理-prd.md

---

# 📌 1. 概述（Overview）

## 1.1 背景

- 当前现状：
  - 执行工厂已经具备 `operator`、`toolbox`、`mcp` 等资源域的统一接入与治理能力，但 Skill 资源此前缺少正式实现设计文档。

- 业务 / 技术背景：
  - Skill 是执行工厂新增资源类型，用于承载 Agent 可消费的能力包。
  - 首版目标是建立 Skill 的正式资源模型、存储模型、接口分层、权限与业务域接入方式，以及渐进式加载能力。

---

## 1.2 目标

- 建立 Skill 的首版正式实现设计，覆盖注册、查询、内容读取、文件读取、下载、删除和市场能力边界。
- 统一 Skill 的资源命名与接口语义，对外使用 `skill_content` 与 `/content`，不保留 `instructions`、`guide` 旧语义。
- 统一 Skill 业务状态语义，与现有资源域保持一致：`unpublish / published / offline`。
- 将 Skill 纳入统一权限体系与业务域体系，资源类型固定为 `skill`。
- 通过“列表/详情轻量化 -> 内容读取 -> 文件读取”的模式落实渐进式加载。
- 明确当前代码的已实现能力、部分实现能力和未实现能力，作为后续开发和评审基线。

---

## 1.3 非目标（Out of Scope）

- 不支持 Skill 在线编辑、在线回写和版本回滚。
- 不支持 Skill 自动推荐、自动绑定和依赖求解。
- 不支持 Skill 沙箱执行。

---

## 1.4 术语说明（Optional）

| 术语 | 说明 |
|------|------|
| Skill | Agent 可消费的能力包，包含 `SKILL.md` 与附属文件 |
| Skill Content | `SKILL.md` 内容对象，对外通过下载地址暴露 |
| Skill 文件 | Skill 内部模板、脚本、参考资料、配置等附属文件 |
| 管理接口 | 面向 Skill 提供方和资源维护方的接口 |
| 市场接口 | 面向 Skill 发现与选择场景的接口 |
| 读取接口 | 面向 `skill_content` 与文件渐进式读取的接口 |
| `f_is_deleted=1` | 删除补偿中间态，属于内部删除流程标记，对外统一视为不存在 |
| AuthService | 平台统一权限服务 |
| BusinessDomainService | 平台统一业务域资源治理服务 |

---

# 🏗️ 2. 整体设计（HLD）

> 本章节关注系统“怎么搭建”，不涉及具体实现细节

---

## 🌍 2.1 系统上下文（C4 - Level 1）

### 参与者
- 用户：Skill 提供方、Agent 配置方、执行工厂运行链路
- 外部系统：统一权限服务、业务域服务、对象存储、MariaDB
- 第三方服务：`oss-gateway-backend`

### 系统关系

    [Skill 提供方 / Agent 配置方 / 运行链路]
        → [执行工厂 Skill 资源域]
        → [MariaDB / Object Storage / AuthService / BusinessDomainService]

---

## 🧱 2.2 容器架构（C4 - Level 2）

| 容器 | 技术栈 | 职责 |
|------|--------|------|
| API Service | Go + Gin | 暴露 Skill 管理接口、读取接口与市场接口 |
| Core Service | Go | 实现注册、查询、删除、下载、内容读取、文件读取、市场逻辑 |
| Parser | Go | 解析 `SKILL.md` frontmatter、正文和 ZIP 内容 |
| Asset Store | Go + `oss-gateway-backend` | 管理 Skill 对象上传、下载、删除与下载地址获取 |
| Storage | MariaDB + Object Storage | 保存 Skill 主记录、文件索引和文件内容 |

---

### 容器交互

    Client → API Service → Skill Core → Repository / Parser / Asset Store
    Skill Core → AuthService / BusinessDomainService
    Repository → MariaDB
    Asset Store → Object Storage

---

## 🧩 2.3 组件设计（C4 - Level 3）

### Skill Core 组件

| 组件 | 职责 |
|------|------|
| SkillRegistry | 注册、删除、下载、管理列表、管理详情、市场列表、市场详情 |
| SkillReader | `SKILL.md` 下载地址获取、单文件下载地址获取 |
| SkillParser | `content` / `zip` 注册内容解析 |
| SkillAssetStore | 对象上传、下载、删除、下载地址获取 |
| SkillRepository(DB) | Skill 主表访问 |
| SkillFileIndex(DB) | Skill 文件索引访问 |
| Governance Integration | 在 `skillRegistry` 中直接复用 `AuthService` 与 `BusinessDomainService` |

---

## 🔄 2.4 数据流（Data Flow）

### 主流程

    注册请求 → 解析 SKILL.md → 写主表 → 写文件索引 → 写对象存储 → 业务域绑定 → 列表发现 → 内容读取 → 文件读取 / 下载

### 子流程（可选）

    删除请求 → CheckDeletePermission → 主记录置 f_is_deleted=1 → 同步删除对象 / 索引 / 主记录 → 解绑业务域 → 删除权限策略

---

## ⚖️ 2.5 关键设计决策（Design Decisions）

| 决策 | 说明 |
|------|------|
| Skill 独立为正式资源类型 | 已新增 `AuthResourceTypeSkill`，避免复用或伪装成其他资源类型 |
| 管理 / 市场 / 读取三层分离 | 列表、详情、市场、内容读取、文件读取的调用方和返回负载不同，必须分层 |
| 内容与文件统一为下载地址 | `/content` 与 `/files/read` 均返回下载地址，统一读取接口形式 |
| 文件索引持久化对象引用 | 文件索引通过 `storage_id + storage_key` 精确定位对象，避免默认存储切换后历史对象漂移 |
| 业务状态与删除流程状态分离 | `status` 仅表达 `unpublish/published/offline`，删除流程通过 `f_is_deleted` 表达 |
| 删除成功后物理删除主记录 | 首版不保留 `deleted` 终态，`f_is_deleted=1` 仅表示待删除/删除中 |
| 对外状态类型统一为 `BizStatus` | 接口契约统一使用 `interfaces.BizStatus`；持久化模型当前以字符串承载同语义状态值，避免包循环依赖 |
| 权限和业务域沿用平台模式接入 | 不为 Skill 单独发明治理层，直接在 `skillRegistry` 上复用统一服务 |
| 下载复用已登记文件索引重建 ZIP | 首版下载按“主记录 + 文件索引 + 对象存储”重建包，不保留原始 ZIP 原封不动回传 |
| 文件读取不做文件级 ACL | Skill 附件可读性仅由 `execute` 权限决定，文件索引只承担定位职责 |

---

## 🚀 2.6 部署架构（Deployment）

- 部署环境：K8s
- 拓扑结构：Skill 作为执行工厂服务内的一个资源域实现，不单独拆服务
- 扩展策略：服务实例水平扩展；MariaDB 复用平台现有部署能力；对象存储通过 `oss-gateway-backend` 接入，优先使用显式 `storage_id`，为空时自动解析默认存储。

---

## 🔐 2.7 非功能设计

### 性能
- 管理列表和详情仅返回轻量元数据，不返回 `skill_content` 与文件清单
- 市场列表同样只返回轻量摘要，且仅允许 `published` 进入市场结果
- `skill_content` 与文件读取均先命中对象引用，再返回下载地址
- 文件读取按 `(skill_id, rel_path)` 精确命中索引，并使用 `storage_id + storage_key` 生成下载地址

### 可用性
- 删除通过 `f_is_deleted=1` 作为补偿标记
- `f_is_deleted=1` 对管理、市场、读取接口均不可见
- 当前不再保留本地存储回退，`oss-gateway-backend` 初始化失败将直接暴露

### 安全
- 注册、管理查询、删除、市场查询已接入统一权限/业务域骨架
- 文件路径标准化，禁止路径穿越
- 文件读取校验路径合法性与 `content_sha256`
- 当前读取接口已接入权限校验；业务域可见性在设计上要求统一接入，后续如需进一步收紧可在读取链继续补强

### 可观测性
- 设计要求：应记录注册失败、删除补偿、下载失败、文件校验失败
- 当前实现状态：已有测试覆盖和错误返回，但未看到独立 metrics / tracing / 审计日志落地代码，属于未实现项

---

# 🔧 3. 详细设计（LLD）

> 本章节关注“如何实现”，开发可直接参考

---

## 🌐 3.1 API 设计

### Skill 管理接口

**Endpoint:** `POST /api/agent-operator-integration/v1/skills`

**Request:**

```json
{
  "file_type": "zip",
  "file": "multipart 或原始内容",
  "source": "upload_zip",
  "extend_info": {
    "tag": "demo"
  }
}
```

**Response:**

```json
{
  "skill_id": "skill-xxx",
  "name": "demo-skill",
  "description": "demo",
  "version": "1.0.0",
  "status": "unpublish",
  "files": [
    "templates/reply.md"
  ]
}
```

实现状态：已实现。

### Skill 管理列表接口

**Endpoint:** `GET /api/agent-operator-integration/v1/skills`

**Request:**

```json
{
  "page": 1,
  "page_size": 10,
  "name": "demo",
  "status": "unpublish",
  "source": "upload_zip"
}
```

**Response:**

```json
{
  "total_count": 1,
  "page": 1,
  "page_size": 10,
  "data": [
    {
      "skill_id": "skill-xxx",
      "name": "demo-skill",
      "description": "demo",
      "version": "1.0.0",
      "status": "unpublish"
    }
  ]
}
```

实现状态：已实现。

### Skill 管理详情接口

**Endpoint:** `GET /api/agent-operator-integration/v1/skills/{skill_id}`

**Request:**

```json
{
  "skill_id": "skill-xxx"
}
```

**Response:**

```json
{
  "skill_id": "skill-xxx",
  "name": "demo-skill",
  "description": "demo",
  "version": "1.0.0",
  "status": "published",
  "source": "upload_zip",
  "extend_info": {},
  "dependencies": {}
}
```

实现状态：已实现。

### Skill 内容接口

**Endpoint:** `GET /api/agent-operator-integration/v1/skills/{skill_id}/content`

**Request:**

```json
{
  "skill_id": "skill-xxx"
}
```

**Response:**

```json
{
  "skill_id": "skill-xxx",
  "url": "https://oss-gateway/download/skill-xxx/SKILL.md",
  "status": "published",
  "files": [
    {
      "rel_path": "templates/reply.md"
    }
  ]
}
```

实现状态：已实现，当前返回下载地址而非正文。

### Skill 文件读取接口

**Endpoint:** `POST /api/agent-operator-integration/v1/skills/{skill_id}/files/read`

**Request:**

```json
{
  "rel_path": "templates/reply.md"
}
```

**Response:**

```json
{
  "skill_id": "skill-xxx",
  "rel_path": "templates/reply.md",
  "url": "https://oss-gateway/download/templates/reply.md",
  "mime_type": "text/markdown"
}
```

实现状态：已实现，当前返回下载地址而非文件正文。

### Skill 下载接口

**Endpoint:** `GET /api/agent-operator-integration/v1/skills/{skill_id}/download`

**Request:**

```json
{
  "skill_id": "skill-xxx"
}
```

**Response:**

```json
{
  "response_type": "application/zip",
  "file_name": "demo-skill.zip"
}
```

实现状态：已实现。

### Skill 市场接口

**Endpoint:** `GET /api/agent-operator-integration/v1/skills/market`

**Request:**

```json
{
  "page": 1,
  "page_size": 10,
  "name": "demo"
}
```

**Response:**

```json
{
  "total_count": 1,
  "data": [
    {
      "skill_id": "skill-xxx",
      "name": "demo-skill",
      "description": "demo"
    }
  ]
}
```

实现状态：部分实现。
说明：逻辑层、公开 HTTP 层和 API 文档均已补齐。

---

## 🗂️ 3.2 数据模型

### SkillRepository

| 字段 | 类型 | 说明 |
|------|------|------|
| skill_id | string | Skill 业务主键 |
| name | string | Skill 名称 |
| description | string | Skill 描述 |
| skill_content | text | `SKILL.md` 正文 |
| version | string | 版本号 |
| status | enum | `unpublish/published/offline` |
| source | string | 来源 |
| extend_info | json string | 扩展元数据 |
| dependencies | json string | 依赖声明 |
| file_manifest | json string | 精简文件摘要 |
| create_user | string | 创建人 |
| update_user | string | 更新人 |
| create_time | int64 | 创建时间纳秒 |
| update_time | int64 | 更新时间纳秒 |
| delete_time | int64 | 删除时间纳秒 |
| delete_user | string | 删除人 |

实现状态：已实现。

### SkillFileIndex

| 字段 | 类型 | 说明 |
|------|------|------|
| skill_id | string | Skill 主键 |
| rel_path | string | 标准化相对路径 |
| path_hash | string | 相对路径 MD5 |
| storage_id | string | 对象所属存储 ID |
| storage_key | string | 对象存储 key |
| file_type | string | 文件分类 |
| content_sha256 | string | 文件内容 SHA-256 |
| mime_type | string | 文件 MIME |
| size | int64 | 文件大小 |
| create_time | int64 | 创建时间纳秒 |
| update_time | int64 | 更新时间纳秒 |

实现状态：已实现。

### SkillSpec（解析视图）

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | frontmatter 必填字段 |
| description | string | frontmatter 必填字段 |
| version | string | 可选版本，默认补 `1.0.0` |
| dependencies | map | 依赖声明 |
| metadata | map | 扩展元数据 |

实现状态：已实现。

---

## 💾 3.3 存储设计

- 存储类型：MariaDB + 对象存储
- 数据分布：
  - `t_skill_repository` 保存 Skill 主记录和 `skill_content`
  - `t_skill_file_index` 保存附件索引和对象引用
  - `oss-gateway-backend` 管理对象访问，执行工厂仅持久化 `storage_id + storage_key`
- 索引设计：
  - `t_skill_repository` 通过 `f_skill_id` 精确查询
  - 列表查询按 `name/source/create_user/status` 过滤
  - `t_skill_file_index` 通过 `(skill_id, rel_path)` 精确读取
  - `path_hash` 仅作为辅助字段，不替代路径校验

实现状态：
- 主表字段和索引：已实现
- 文件索引表：已实现
- `oss-gateway-backend` 正式接入：已实现
- 流式大文件下载：未实现，当前下载为内存组包

---

## 🔁 3.4 核心流程（详细）

### Skill 注册流程

1. API 层绑定 header、form/json/multipart 参数。
2. `skillRegistry.RegisterSkill` 调用 `AuthService.GetAccessor`。
3. 调用 `CheckCreatePermission(ctx, accessor, skill)`。
4. `skillParser` 根据 `file_type` 解析 `content` 或 `zip`。
5. `parseSkillContent` 提取 frontmatter，构建主记录，正文写入 `SkillContent`。
6. 若为 ZIP，解析 `SKILL.md` 之外的文件，生成 `SkillFileSummary` 与 `skillAsset`。
7. 开启 DB 事务，写入主表。
8. 状态保持为 `unpublish`。
9. 将附件内容上传到 `oss-gateway-backend` 所在对象存储，并写入 `storage_id + storage_key` 文件索引。
10. 调用 `BusinessDomainService.AssociateResource` 绑定业务域。
11. 返回注册响应。

异常流：
- `SKILL.md` 不合法、缺失 frontmatter、路径非法时直接失败。
- 附件写入或索引写入失败时整体失败。

实现状态：已实现。
未实现项：注册成功后未看到 `CreateOwnerPolicy` 调用，若平台要求 owner 策略自动创建，需要补充实现。

### Skill 删除流程

1. 查询 Skill 主记录。
2. 校验业务状态可删除。
3. 调用 `GetAccessor` 和 `CheckDeletePermission`。
4. 开启 DB 事务并将 `f_is_deleted` 置为 `1`。
   对应仓储接口：`UpdateSkillDeleted(...)`
5. 在当前请求内同步查询该 Skill 的文件索引。
6. 逐个删除对象存储单文件。
7. 删除文件索引记录。
8. 删除主表记录。
9. 调用 `DisassociateResource` 解绑业务域。
10. 调用 `DeletePolicy` 删除权限策略。

异常流：
- 任意对象删除失败时，主表保留 `f_is_deleted=1`，后续需由独立补偿任务继续重试。
- 成功后主表物理删除，不保留 `deleted` 终态。

实现状态：部分实现。
未实现项：独立补偿任务未实现；审计日志未实现。

### Skill 管理查询流程

1. 列表查询先 `GetAccessor`。
2. DB 侧按名称、来源、状态、创建人等条件查询。
3. 逻辑层剔除 `f_is_deleted=1` 的记录，并仅对外暴露符合当前接口语义的业务状态。
4. 调用 `ResourceFilterIDs(..., view)` 过滤可见 Skill。
5. 返回 `SkillSummary`。
6. 详情查询先查单条记录，再做 `CheckViewPermission`。

实现状态：已实现。
注意：列表总数当前基于 DB count 计算，权限过滤后可能与最终 `data` 长度存在短暂差异。

### Skill 市场查询流程

1. 获取访问者 `Accessor`。
2. 调用 `ResourceListIDs(..., public_access)`。
3. 调用 `BusinessDomainService.BatchResourceList` 获取业务域资源映射。
4. 查询候选 Skill 列表。
5. 逻辑层过滤掉 `f_is_deleted=1`、无公共访问权限、无业务域映射、非 `published` 的 Skill。
6. 市场详情通过 `CheckPublicAccessPermission` 且状态为 `published` 后返回。

实现状态：已实现。
说明：逻辑层、公开 HTTP 层和 API 文档均已补齐。

### Skill 内容读取流程

1. `GetSkillContent` 查询主记录。
2. 若不存在或 `f_is_deleted=1`，返回未找到。
3. 基于 `SKILL.md` 对象引用返回下载地址和 `file_manifest`。

实现状态：已实现。
说明：当前已接入 `GetAccessor + OperationCheckAny(execute/public_access/view)`，并在 `f_is_deleted=1` 时统一返回未找到。

### Skill 文件读取流程

1. 查询主记录，要求 `f_is_deleted=0`；附件读取权限当前通过 `OperationCheckAny(execute/public_access/view)` 判定。
2. 标准化 `rel_path`。
3. 按 `(skill_id, rel_path)` 查询文件索引。
4. 不再依赖文件级访问控制字段，附件读取权限当前通过 `OperationCheckAny(execute/public_access/view)` 判定。
5. 根据索引中的 `storage_id + storage_key` 生成下载地址。
6. 返回下载地址和元数据。

实现状态：已实现。
说明：当前已接入 `GetAccessor + OperationCheckAny(execute/public_access/view)`，并保留路径校验、索引校验和 SHA-256 校验。

### Skill ZIP 下载流程

1. 查询主记录，`f_is_deleted=1` 不可下载。
2. 调用 `GetAccessor` 和 `CheckViewPermission`。
3. 查询文件索引列表。
4. 在内存中创建 ZIP。
5. 重新生成 `SKILL.md`：
   - frontmatter：`name/description/version`
   - body：`skill_content`
6. 逐个读取对象存储文件并写入 ZIP。
7. 返回 zip 二进制和下载文件名。

实现状态：已实现。
注意：当前下载包不是原始上传 ZIP 的原封重放，而是按主数据和附件索引重建。

---

## 🧠 3.5 关键逻辑设计

### `SKILL.md` 解析逻辑
- 通过 `---` 分隔 frontmatter 与正文。
- frontmatter 必填 `name`、`description`。
- `version` 缺省默认 `1.0.0`。
- `metadata` 映射到 `extend_info`。
- 正文写入 `SkillContent`。

### ZIP 路径标准化逻辑
- 使用 `filepath.Clean`
- 统一转换为 `/`
- 去除前导 `./` 和 `/`
- 拒绝空路径、`.`、`..`、路径穿越

### 文件访问控制逻辑
- 文件索引只承担路径定位、类型标识和内容校验职责
- 附件可读性仅由 Skill `execute` 权限决定
- 文件摘要是否出现在内容接口中，仅由内容接口返回模型决定

### 状态机逻辑
- `unpublish`：资源已存在但未进入发布态
- `published`：可进入市场并可被正常发现
- `offline`：资源下线，不进入市场
- `f_is_deleted=1`：删除补偿中间态，对外不可见

### 权限与业务域治理逻辑
- 资源类型固定为 `AuthResourceTypeSkill`
- 注册：`CheckCreatePermission + AssociateResource`
- 管理查询：`ResourceListIDs(..., view)`
- 管理详情：`CheckViewPermission`
- 内容读取 / 附件读取：`OperationCheckAny(execute/public_access/view)`
- 删除：`CheckDeletePermission + DisassociateResource + DeletePolicy`
- 市场列表：当前实现经 `querySkillListPage` 走 `ResourceListIDs(..., public_access) + BatchResourceList`，仅返回 `published`
- 市场详情：`CheckPublicAccessPermission`，且仅返回 `published`

当前实现缺口：
- 下载接口当前满足 `execute/public_access/view` 任一权限即可访问，未定义独立 `download` 操作

---

## ❗ 3.6 错误处理

- 参数错误：
  - `SKILL.md not found in zip`
  - `invalid SKILL.md format`
  - `invalid skill file path`
- 业务错误：
  - `skill not found`
  - `skill can not be deleted in status`
  - `skill file access denied`
- 一致性错误：
  - 对象存储读取失败
  - 文件校验和不匹配
  - 删除过程中对象残留

处理策略：
- `f_is_deleted=1` 统一对外视为不存在
- 下载和文件读取遇到对象缺失时整体失败，不返回部分内容
- 当前未实现统一错误码枚举和独立补偿任务，需要后续补足

---

## ⚙️ 3.7 配置设计

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `oss_gateway.private_base.base_url` | 无 | `oss-gateway-backend` 私网地址 |
| `oss_gateway.storage_id` | 无 | 显式指定的存储 ID，优先使用 |
| `oss_gateway.refresh_interval` | `5m` | 默认存储刷新周期 |

实现状态：
- 正式配置结构：已实现
- 默认存储自动发现：已实现
- 本地文件存储回退：未实现

---

## 📊 3.8 可观测性实现

- tracing：
  - 设计要求：注册、删除、下载、内容读取、文件读取、市场查询均应埋点
  - 当前状态：未看到 Skill 独立 tracing 实现，未实现

- metrics：
  - 设计要求：注册成功率、读取成功率、下载失败次数、删除补偿次数、对象残留次数
  - 当前状态：未看到 Skill 独立 metrics 实现，未实现

- logging：
  - 当前实现依赖通用错误返回和基础日志能力
  - 设计要求：应明确记录权限失败、业务域过滤失败、对象存储异常、校验和异常
  - 当前状态：部分实现

---

## 实现状态汇总

| 能力 | 状态 | 说明 |
|------|------|------|
| Skill 注册 | 已实现 | 支持 `zip` / `content` |
| 管理列表 / 管理详情 | 已实现 | 已接入管理侧权限过滤 |
| `skill_content` 读取 | 已实现 | 已公开暴露，当前返回下载地址，权限为 `execute/public_access/view` 任一满足 |
| 文件读取 | 已实现 | 已返回下载地址，权限为 `execute/public_access/view` 任一满足 |
| ZIP 下载 | 已实现 | 管理侧路由已暴露 |
| 删除补偿态 | 部分实现 | 删除入口已切换为 `f_is_deleted`，当前仍为同步删除，独立补偿任务未实现 |
| 市场列表 / 市场详情逻辑 | 已实现 | 逻辑层、handler、route、API 文档均已暴露 |
| 权限治理骨架 | 已实现 | 已复用 `AuthService` |
| 业务域治理骨架 | 已实现 | 已复用 `BusinessDomainService` |
| 运行时绑定接口 | 未实现 | 仅有接口，无实现代码 |
| 补偿任务 / 审计日志 / 指标 | 未实现 | 需后续补足 |
