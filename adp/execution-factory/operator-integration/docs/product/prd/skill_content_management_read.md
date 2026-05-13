# PRD: Skill 内容管理读取能力

## 文档信息

- Status: Draft
- Owner: @chenshu-zhao
- Last Updated: 2026-04-29

## 关联设计

- Issue: [execution-factory Skill 内容管理能力完善 #302](https://github.com/kweaver-ai/kweaver-core/issues/302)
- Design: [Skill 内容管理读取设计](../../design/features/skill_content_management_read.md)

---

## 1. 背景

### 业务现状

执行工厂目前支持 Skill 的完整生命周期管理（注册、编辑、发布、下架、删除），但**内容读取能力**主要集中在"已发布 Skill 的使用场景"上。当前：

- `GetSkillContent` 和 `ReadSkillFile` 只读取 `skill_release`（已发布快照）中的内容
- `GetSkillDetail` 只返回结构化元数据（名称、描述、版本、状态），不包含 SKILL.md 正文或文件清单
- 管理端前端缺少面向内容的读接口，难以实现"内容预览、文件浏览、在线查看"等功能

### 存在问题

1. **管理态内容不可读**：管理端无法查看正在编辑中的 SKILL.md 原文、文件列表和文件内容
2. **content 注册类型的读能力缺失**：`content` 方式注册的 Skill 只在 DB 中存有 SKILL.md 正文，OSS 中无对应文件，`GetSkillContent` 会返回 404
3. **状态语义不清晰**：不同状态下（unpublish / editing / published / offline），"当前管理内容"与"已发布内容"的边界未在 API 层区分
4. **两种注册方式的读体验不统一**：`content` 和 `zip` 注册的 Skill 在管理端读取路径不一致

### 触发原因

Skill 功能从"注册即用"演进到"持续维护、多人协作、多环境管理"，需要管理端、SDK、CLI 具备一致的**管理态内容读取能力**，以支撑以下业务场景：

- 管理端 UI 展示 Skill 当前编辑态的 SKILL.md 原文
- 管理端 UI 展示 Skill 包内的文件清单并支持文件预览
- SDK/CLI 通过统一 API 拉取管理态内容用于离线编辑和审计

---

## 2. 目标

### 业务目标

- 管理端能查看任意状态 Skill 的**当前编辑态** SKILL.md 内容
- 管理端能查看 Skill 包的**文件清单**并读取指定文件
- 管理端能**下载**当前编辑态的完整 Skill 包（含全部文件资产）
- 已有发布态读接口不受影响，保持向后兼容

### 产品目标

- 建立"管理态读"与"发布态读"两条清晰的读路径
- 明确各状态下管理端读的语义
- 统一 `content` 注册与 `zip` 注册的管理端读体验
- 权限校验与现有体系一致：管理态走"查看/修改权限"，发布态走"执行/公开访问权限"

---

## 3. 非目标

- 管理端前端 UI 开发（此 PRD 只覆盖后端 API）
- 单文件粒度的增量补丁编辑
- 历史版本之间的内容差异对比
- 跨 Skill 批量内容读取
- SDK/CLI 端包装实现
- 旧版本 OSS 资产的自动回收

---

## 4. 术语定义

| 术语 | 定义 |
|------|------|
| 管理态内容 | `skill_repository` 表中当前编辑中的草稿内容 |
| 发布态内容 | `skill_release` 表中最后一次发布时的快照内容 |
| content 注册 | 通过提交纯文本 SKILL.md 注册 Skill，本次改造后补齐 OSS 文件资产 |
| zip 注册 | 通过上传 ZIP 包注册 Skill，文件存储在 OSS，元数据在 `skill_file_index` |
| SKILL.md | Skill 定义文件，含 YAML frontmatter 和 Markdown body |
| file_manifest | `skill_repository` 表中的 JSON 字段，记录注册时的文件摘要列表 |

---

## 5. 状态与读语义矩阵

管理端读取"当前编辑态内容"，不受发布状态影响。不同业务状态下读接口的行为：

| Skill 状态 | 管理端读到的内容 | 发布端（现有）读到的内容 |
|-----------|----------------|----------------------|
| **unpublish**（未发布） | repository 当前内容 ✅ | 404（无 release） |
| **editing**（编辑中） | repository 当前内容 ✅ | 上一次 release 快照 |
| **published**（已发布，无编辑） | repository 当前内容（与 release 一致）| release 快照 |
| **offline**（已下架） | repository 当前内容 ✅ | 404（release 已删除） |
| **deleted**（已删除） | 404，明确错误码 | 404 |

**核心原则：**
- 管理端读路径始终返回 `skill_repository` 中的当前数据
- 发布端读路径（现有接口）保持不变，始终返回 `skill_release` 中的数据
- 两条路径互不干扰

---

## 6. 用户与场景

### 6.1 用户角色

| 角色 | 描述 |
|------|------|
| Skill 作者 | 编写和维护 SKILL.md、管理文件资产 |
| 管理端用户 | 通过 UI 查看、浏览、审阅 Skill 内容 |
| SDK/CLI 用户 | 通过 API 拉取管理态内容用于本地编辑或审计 |

### 6.2 用户故事

- 作为 Skill 作者，我希望在管理端查看当前编辑态的 SKILL.md 内容，从而确认我的修改已正确保存。
- 作为 Skill 作者，我希望在管理端浏览 Skill 包内的所有文件，从而了解当前版本包含哪些资产。
- 作为管理端用户，我希望预览任意指定文件的内容（如 Python 脚本、配置文件），从而在不下载的情况下快速审阅。
- 作为管理端用户，我希望在 Skill 处于 unpublish / editing / offline 状态下仍能查看/下载其内容，从而在不上架的情况下进行协作。
- 作为 SDK 集成方，我希望能统一获取管理态内容，无论 Skill 是 content 还是 zip 注册的。

### 6.3 使用场景

- **场景 1**：管理端 Skill 详情页展示 SKILL.md 原文 + 元数据，进入"内容预览"模式
- **场景 2**：管理端 Skill 文件列表页展示包内所有文件（文件名、类型、大小、MIME），支持按路径搜索
- **场景 3**：管理端点击文件，发起"文件读取"请求，获取可预览的 URL 或文件内容
- **场景 4**：管理端下载当前编辑态的完整 Skill 包（ZIP 格式）
- **场景 5**：SDK 调用 `kweaver skill content` 获取管理态 SKILL.md

---

## 7. 需求范围

### ✅ In Scope

- 管理态 SKILL.md 内容读取
- 管理态文件清单查询
- 管理态指定文件内容读取（返回预签名 URL）
- content 注册类型的管理端读能力补齐
- 状态语义在 reader 层的明确区分（管理态 vs 发布态）
- 权限、业务域隔离、删除态处理、路径安全校验
- 向后兼容：现有发布态读接口不受影响
- OpenAPI 文档更新

### ❌ Out of Scope

- 单文件粒度的增量补丁编辑
- 历史版本差异对比接口
- 跨 Skill 批量内容读取
- SDK/CLI 包装实现
- 管理前端 UI 开发

---

## 8. 功能需求

### 8.1 功能结构

```
Skill 内容管理读取
├── 管理态读接口（新增/扩展）
│   ├── GetSkillRepositoryContent    — 读取当前编辑态 SKILL.md
│   ├── GetSkillRepositoryFiles      — 获取当前编辑态文件清单
│   ├── ReadSkillRepositoryFile      — 读取当前编辑态指定文件
│   └── DownloadSkillRepository      — 下载当前编辑态完整包（已有，需确认）
├── 发布态读接口（现有，不变）
│   ├── GetSkillContent
│   ├── ReadSkillFile
│   ├── DownloadSkill
│   └── GetSkillReleaseHistory
└── 公共能力
    ├── normalizeZipPath 路径校验（复用）
    └── 权限 + 业务域隔离（复用）
```

### 8.2 API 设计

#### 方案：独立端点（独立路由 + 独立 Logic 接口）

管理态读接口使用独立的 `/management/` 路径前缀，与现有发布态接口完全解耦。handler 层、logic 层、数据源均为独立实现，不存在参数分流。

```
# 发布态读接口（现有，不变）
GET  /skills/{skill_id}/content                  → SkillReader.GetSkillContent
POST /skills/{skill_id}/files/read               → SkillReader.ReadSkillFile
GET  /skills/{skill_id}/download                 → SkillRegistry.DownloadSkill
GET  /skills/{skill_id}/history                  → SkillReader.GetSkillReleaseHistory

# 管理态读接口（新增，独立路由）
GET  /skills/{skill_id}/management/content       → SkillManagementReader.GetManagementContent
POST /skills/{skill_id}/management/files/read    → SkillManagementReader.ReadManagementFile
GET  /skills/{skill_id}/management/download      → SkillManagementReader.DownloadManagementSkill
```

**决策理由：**
- 已有服务**零影响**：现有调用不需要改任何 URL 或传任何额外参数
- 路由自描述：看路径就知道是读管理态还是发布态
- 权限链分离：路由层直接挂不同的 middleware，不存在 handler 内部分流
- Logic 层隔离：`SkillReader` 不改一行代码，新增 `SkillManagementReader` 作为第三个独立接口
- OpenAPI 文档完全独立，无交叉描述

---

#### 【FR-1】管理态 SKILL.md 内容读取

| 字段 | 值 |
|------|-----|
| 接口 | `GET /skills/{skill_id}/management/content` |
| 权限 | 公共 API：检查 `view` / `modify` 权限 |
| | 内部 API：无权限校验（已有机制） |

**请求：**
```
GET /api/agent-operator-integration/v1/skills/{skill_id}/management/content
Authorization: Bearer <token>
x-business-domain: <bd_id>
```

**响应：**

```json
{
  "skill_id": "uuid",
  "name": "skill-name",
  "description": "skill description",
  "version": "v1.0.0",
  "status": "editing",
  "source": "custom",
  "file_type": "zip",
  "url": "https://oss-presigned-url/skills/SKILL.md",
  "files": [
    {
      "rel_path": "scripts/main.py",
      "file_type": "script",
      "size": 1024,
      "mime_type": "text/x-python"
    }
  ]
}
```

**响应字段说明：**
| 字段 | 说明 |
|------|------|
| `url` | 始终返回 SKILL.md 的 OSS 预签名 URL（content 注册补齐 OSS 后也支持） |
| `files` | file_manifest 反序列化结果，content 注册可能为空数组 `[]` |

**业务规则：**
- 总是从 `skill_repository` 表读取
- `url` 指向的 OSS 文件中的 name/desc 与 DB 中一致（通过元数据编辑时同步重写 OSS 保证）
- Skill 已删除（`is_deleted=true`）返回 404
- 不存在 release 记录不影响管理端读取

**边界条件：**
| 场景 | 行为 |
|------|------|
| content 注册（补齐后），status=unpublish | 返回 OSS presigned URL，files=[] |
| zip 注册，status=published（有编辑） | 返回当前编辑态的 SKILL.md 和文件清单（可能与 release 不同） |
| zip 注册，status=published（无编辑） | 返回 repository 内容（与 release 内容一致） |
| skill 已删除 (`is_deleted=true`) | 返回 404 |

**异常处理：**
- Skill 存在但无 `file_type` 信息 → 默认按 content 处理
- SKILL.md OSS 记录缺失 → 返回错误

---

#### 【FR-2】管理态文件清单查询

| 字段 | 值 |
|------|-----|
| 接口 | `GET /skills/{skill_id}/management/files` |
| 权限 | 公共 API：检查 `view` / `modify` 权限 |

**请求：**
```
GET /api/agent-operator-integration/v1/skills/{skill_id}/management/files
```

**响应：**
```json
{
  "skill_id": "uuid",
  "version": "v1.0.0",
  "file_type": "zip",
  "total_files": 3,
  "files": [
    { "rel_path": "SKILL.md", "file_type": "reference", "size": 2048, "mime_type": "text/markdown" },
    { "rel_path": "scripts/main.py", "file_type": "script", "size": 4096, "mime_type": "text/x-python" },
    { "rel_path": "config.yaml", "file_type": "config", "size": 512, "mime_type": "application/yaml" }
  ]
}
```

**业务规则：**
- 从 `skill_repository.file_manifest` 反序列化文件列表
- content 注册：`file_type` = "content"，`total_files` = 0，`files` = 空数组
- 不存在 file_manifest 时返回空数组而非 null

---

#### 【FR-3】管理态指定文件读取

| 字段 | 值 |
|------|-----|
| 接口 | `POST /skills/{skill_id}/management/files/read` |
| 权限 | 公共 API：检查 `view` / `modify` 权限 |

**请求：**
```
POST /api/agent-operator-integration/v1/skills/{skill_id}/management/files/read
{
  "rel_path": "scripts/main.py"
}
```

**响应：**
```json
{
  "skill_id": "uuid",
  "rel_path": "scripts/main.py",
  "url": "https://oss-presigned-url/skills/file",
  "mime_type": "text/x-python",
  "file_type": "script",
  "size": 4096
}
```

**业务规则：**
- 从 `skill_file_index` 表查询，以 `skill_repository.version` 为版本键（而非 release version）
- 路径安全校验复用 `normalizeZipPath`
- 所有注册类型统一走 OSS 预签名 URL（content 注册补齐 OSS 后也会在 file_index 中有 SKILL.md 记录）

**边界条件：**
| 场景 | 行为 |
|------|------|
| 路径存在 | 返回 OSS 预签名 URL |
| 路径不存在 | 返回 404 |
| 路径穿越攻击 (`../../etc/passwd`) | `normalizeZipPath` 拦截，返回 400 |
| Skill 版本更新后读旧路径 | 以 repository 的当前 version 为准 |

---

#### 【FR-4】管理态内容下载

| 字段 | 值 |
|------|-----|
| 接口 | `GET /skills/{skill_id}/management/download` |
| 权限 | 公共 API：检查 `view` / `modify` 权限 |

**说明：**
从 `skill_repository` 和 `skill_file_index` 构建当前编辑态的完整 zip 包。已有 `DownloadSkill`（读 repository）行为不变，新增此端点用于语义明确。

对于 `content` 注册类型：构建一个仅含 SKILL.md 的 ZIP 包返回。

---

#### 【FR-5】content 注册的 OSS 补齐

**问题：** content 注册时 SKILL.md 仅存于 `f_skill_content` 字段（markdown body，不含 frontmatter），未写入 OSS，导致所有基于 OSS 的读链路无法工作。

**方案：** content 注册（及包更新）时，将原始请求的完整 SKILL.md（含 frontmatter）写入 OSS + `skill_file_index`，建立一条 `rel_path=SKILL.md` 的索引记录。

**改动范围：**
- parser 层：content 分支生成 `file_manifest` 和 `skillAsset`，asset 的 Content 为原始请求的 SKILL.md 全文
- registry 层：content 注册也走 `persistSkillAssets`（文件上传 + file_index 写入）
- `f_skill_content` 字段仅存 frontmatter 之后的 body 部分（保持现有行为）

**效果：**

| 维度 | 改造前 | 改造后 |
|------|--------|--------|
| OSS | 无任何文件 | 存有原始完整 SKILL.md |
| `skill_file_index` | 无记录 | 有 `rel_path=SKILL.md` 记录 |
| `file_manifest` | null | 有 SKILL.md 的文件摘要 |
| `f_skill_content` | body 文本 | body 文本（不变） |

**存量处理：**
- 新注册 content：注册时自动写入 OSS ✅
- 存量 content：惰性补齐——`GetManagementContent` 检测到 file_manifest 为空且 `skill_content` 非空时，自动补齐后再返回；或提供后台任务批量补齐

---

#### 【FR-6】元数据编辑时同步重写 OSS SKILL.md

**问题：** `UpdateSkillMetadata` 只更新了 DB 中的结构化字段（`name`、`description`），OSS 中原始 SKILL.md 的 frontmatter 仍是旧值，导致所有基于 OSS 的读路径返回的 SKILL.md 内容与实际元数据不一致。

**方案：** `UpdateSkillMetadata` 成功提交事务后，异步（或同步）重写 OSS 中当前版本 SKILL.md 的 frontmatter。

**交互流程：**
```
UpdateSkillMetadata(newName, newDesc)
  ├── 1. 查询 skill_repository（已有）
  ├── 2. 权限校验（已有）
  ├── 3. 重名校验（已有）
  ├── 4. 更新 DB：repo.name = newName, repo.desc = newDesc（已有）
  ├── 5. 提交事务（已有）
  └── 6. × 新增：重写 OSS SKILL.md frontmatter
        ├── a. 从 OSS 下载当前 SKILL.md
        ├── b. 解析 YAML frontmatter
        ├── c. 替换 name/desc 为最新值（其他自定义字段保持不动）
        ├── d. 重新上传到同一 OSS 路径（覆盖）
        └── e. 失败只记录日志，不阻塞主流程
```

**详细逻辑：**

```python
# OSS SKILL.md frontmatter rewrite 伪代码
oss_content = download_from_oss(skill_id, version, "SKILL.md")
parts = oss_content.split("---", 2)  # ['', '<frontmatter>', '\n<body>']

frontmatter = yaml.load(parts[1])
frontmatter["name"] = new_name        # 替换
frontmatter["description"] = new_desc # 替换
# 其余自定义字段（如 icon, tags, version 等）保持不动

parts[1] = yaml.dump(frontmatter)
new_content = "---".join(parts)

upload_to_oss(skill_id, version, "SKILL.md", new_content)
```

**重要：只替换 `name` 和 `description` 两个字段，YAML frontmatter 中所有其他自定义字段（如 icon、tags、author_email 等）保持不动。**

**影响范围：**

| 读路径 | 改造前 | 改造后 |
|--------|--------|--------|
| `GetManagementContent(url)` | name/desc 可能是旧的 | name/desc 是新的 ✅ |
| `ReadManagementFile(SKILL.md)` | name/desc 可能是旧的 | name/desc 是新的 ✅ |
| `DownloadManagementSkill` | ZIP 中 SKILL.md 的 name/desc 可能是旧的 | ZIP 中 name/desc 是新的 ✅ |
| 发布态 `GetSkillContent(url)` | name/desc 可能是旧的（editing 状态时） | name/desc 是新的 ✅ |

**补充：`UpdateSkillPackage` 时是否需要？**

不需要。包更新时用户上传了新的 SKILL.md，frontmatter 中的 name/desc 就是用户提交的值，不存在不一致问题。

---

### 8.3 接口对比总结

| 接口 | 路径 | 数据源 | 权限（公共 API） | 变更类型 |
|------|------|--------|----------------|---------|
| `GetSkillContent` | `GET .../content` | `skill_release` | execute / view / public_access | 已有，不变 |
| `GetManagementContent` | `GET .../management/content` | `skill_repository` | view / modify | 新增 |
| `ReadSkillFile` | `POST .../files/read` | `skill_file_index`(release version) | execute / view / public_access | 已有，不变 |
| `ReadManagementFile` | `POST .../management/files/read` | `skill_file_index`(repo version) | view / modify | 新增 |
| `DownloadSkill` | `GET .../download` | `skill_repository` + `skill_file_index` | view / public_access | 已有，不变 |
| `DownloadManagementSkill` | `GET .../management/download` | `skill_repository` + `skill_file_index` | view / modify | 新增 |
| `GetSkillReleaseHistory` | `GET .../history` | `skill_release_history` | execute / view | 已有，不变 |

### 8.4 数据一致性说明

管理态读路径的数据一致性由以下两个新增机制保证：

| 机制 | 触发时机 | 保证的数据一致性 |
|------|---------|----------------|
| **FR-5: content 注册 OSS 补齐** | 注册/包更新 | content 注册的 Skill 也有 OSS 文件资产，读路径统一 |
| **FR-6: OSS SKILL.md 重写** | 元数据编辑成功后 | OSS 中的 SKILL.md frontmatter 的 name/desc 始终与 DB 一致 |

---

## 9. 非功能需求

### 9.1 性能
- 管理端内容读取接口 P99 响应时间 < 500ms（OSS 预签名 URL 生成无外部网络开销）
- 文件读取接口 P99 < 1s（涉及 OSS 查询）

### 9.2 安全
- 路径遍历防护复用现有 `normalizeZipPath` 机制
- 管理态读接口需要更严格的权限校验（`view` 或 `modify`）
- 删除态 Skill 不允许任何读操作

### 9.3 可观测性
- 所有新增接口支持 tracing（`o11y.StartInternalSpan`）
- 日志记录 `skill_id`、`user_id`、`bd_id`
- 区分管理读 vs 发布读的 metrics（通过路由前缀区分）

---

## 10. 验收标准

### 正向场景

1. **管理端读取编辑态内容**
   - Given 一个 `editing` 状态的 Skill（zip 注册），When 调用 `GET .../management/content`，Then 返回当前 `skill_repository` 中的 SKILL.md 预签名 URL 和文件清单

2. **content 注册的管理端可读（FR-5）**
   - Given 一个刚注册的 content 类型 Skill，When 调用 `GET .../management/content`，Then 返回 SKILL.md 的 OSS 预签名 URL（不再返回 null）

3. **文件清单可查**
   - Given 一个 zip 注册的 Skill，When 调用 `GET .../management/files`，Then 返回 `file_manifest` 中的完整文件列表

4. **指定文件可读**
   - Given 一个 zip 注册的 Skill，When 调用 `POST .../management/files/read` 且传有效 `rel_path`，Then 返回该文件的 OSS 预签名 URL

5. **content 注册读 SKILL.md 文件（FR-5）**
   - Given 一个 content 注册的 Skill，When 调用 `POST .../management/files/read` 且 `rel_path=SKILL.md`，Then 返回 OSS 预签名 URL（不再需要 DB 重构）

6. **删除态不可读**
   - Given 一个已删除 Skill（`is_deleted=true`），When 调用任意管理读接口，Then 返回 404

7. **路径安全校验**
   - Given 一个 zip 注册的 Skill，When 调用管理态文件读取接口且 `rel_path` 包含 `../`，Then 返回 400 错误

8. **元数据编辑后 OSS 内容同步更新（FR-6）**
   - Given 一个已存在的 Skill，When 调用 `UpdateSkillMetadata` 更新 name/desc 后，Then 调用 `GET .../management/content` 返回的 `url` 指向的 OSS 文件中，frontmatter 的 name/desc 与刚更新的值一致

9. **元数据编辑不影响自定义 YAML 字段**
   - Given 一个 Skill 的 SKILL.md 中存在自定义字段 `icon: foo`，When 调用 `UpdateSkillMetadata` 更新 name 后，Then OSS 中 SKILL.md 的 `icon` 字段仍然是 `foo`

### 向后兼容

10. **发布态读完全不变**
    - Given 一个已发布 Skill，When 调用 `GET .../content`，Then 返回结果与现有行为完全一致（URL 不变，参数不变，响应体不变）

11. **管理读不影响发布态**
    - Given 一个 `editing` 态的 Skill，When 多次调用管理读接口，Then `skill_release` 数据不受影响

12. **现有路由无参数污染**
    - Given 现有调用方，When 继续使用原路径调用发布态接口，Then 不需要传任何新增参数

---

## 11. 失败条件

- 未区分管理态和发布态的读路径，导致 SDK/CLI 读到错误的版本
- content 注册的 Skill 在管理端仍无法读取 SKILL.md 内容
- 权限校验不当导致未授权用户读到草稿内容
- 现有发布态接口因改造行为异常（回归缺陷）

---

## 12. 风险与待确认项

### 风险
- content 注册 OSS 补齐会增加注册流程的写入延迟（一次 OSS 上传 + file_index 写入）
- 存量 content 注册的 Skill 补齐 OSS 记录前管理端读可能不可用（惰性补齐策略可缓解）
- OSS SKILL.md 重写失败只记录日志，不阻塞元数据编辑主流程——需关注重写失败后的告警和补偿

### 待确认项
- [ ] 存量 content 注册的 SKILL.md 补齐策略：惰性补齐（读时触发）还是后台批量任务
- [ ] SLA 目标（响应时间、可用性）

---

## 13. 依赖

- 内部依赖：权限服务（已就绪）、业务域服务（已就绪）、OSS 网关（已就绪）
- 文档依赖：[API 设计文档](../../design/features/skill_content_management_read.md) 待编写
- OpenAPI 规范：`docs/apis/api_public/skill.yaml` 和 `docs/apis/api_private/skill.yaml` 待更新
