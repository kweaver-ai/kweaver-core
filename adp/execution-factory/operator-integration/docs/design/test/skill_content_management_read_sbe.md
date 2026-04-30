# Skill 内容管理读取 — SBE（Specification By Example）

## 范围

- `SkillManagementReader` 接口的 3 个方法：`GetManagementContent` / `ReadManagementFile` / `DownloadManagementSkill`
- 辅助函数：`detectSkillFileType` / `buildArchiveFromFiles` / `updateFrontmatterNameDesc`
- FR-5：content 注册时 SKILL.md 写入 OSS + skill_file_index（`parseRegisterReq` content 分支）
- FR-6：元数据编辑后 OSS SKILL.md frontmatter 同步重写（`UpdateSkillMetadata` + `rewriteSkillMDFrontmatter`）

---

## 1. GetManagementContent

**Endpoint:** `GET /v1/skills/{skill_id}/management/content[?response_mode=url|content|auto]`
**Interface:** `GetManagementContent(ctx, req) → (resp, err)`
**Req header:** `X-Business-Domain`, `User-ID`(public) | **Req uri:** `skill_id` | **Req query:** `response_mode`(缺省`url`)
**Logic source:** `skillRepository.SelectSkillByID` + `skillFileIndex.SelectSkillFileByPath`(SKILL.md) + `assetStore.GetDownloadURL` / `assetStore.Download`

| ID | 场景 | 前置条件 | 请求参数 | 期望响应 | 验证点 |
|----|------|---------|---------|---------|--------|
| E1 | zip 注册 — 默认(url模式)返回 URL | FileManifest 含 scripts/main.py、SKILL.md 等多条；skill_file_index 存在 SKILL.md 记录 | BusinessDomainID:bd-1 SkillID:skill-1 response_mode:""(缺省) | file_type:zip url:https://...(非空) content:"(omitempty) files:[...] | url 非空；content 不存在或为空；files 含全部 manifest |
| E2 | zip 注册 + content 模式 — 返回 OSS 下载的内联正文 | 同 E1；且 OSS 中 SKILL.md 可下载 | 同上 + response_mode:content | file_type:zip url:https://... content:"# SKILL.md body" | content 为 OSS 中 SKILL.md 全文；url 仍非空 |
| E3 | content 注册 + url 模式 — 只返回 URL | SkillContent 非空；skill_file_index 有 SKILL.md 记录 | BusinessDomainID:bd-1 SkillID:skill-content-1 response_mode:url | file_type:content url:https://... content:""(omitempty) | url 非空；content 为空 |
| E4 | content 注册 + content 模式 — 返回 DB 正文 | SkillContent 非空；skill_file_index 可有可无 | 同上 + response_mode:content | file_type:content content:"body text" url:https://...(若有 OSS) | content 等于 SkillContent 原文 |
| E5 | content 注册 + auto 模式 — 向后兼容 | SkillContent 非空；skill_file_index 有 SKILL.md 记录 | 同上 + response_mode:auto | file_type:content url:https://... content:"body text" | content 非空(同旧行为) |
| E6 | 已删除 Skill — 返回 404 | IsDeleted:true | BusinessDomainID:bd-1 SkillID:skill-deleted-1 | HTTP 404 | 不调用 fileRepo/assetStore |
| E7 | 不存在 Skill — 返回 404 | SelectSkillByID 返回 nil,nil | BusinessDomainID:bd-1 SkillID:skill-nonexistent | HTTP 404 | 同上 |
| E8 | 公有 API + 有权限 — 正常返回 | IsPublicAPIFromCtx:true；OperationCheckAny(view,modify)返回true | 同上 + UserID:user-view | 正常 200 响应 | AuthService.GetAccessor + OperationCheckAny 被调用 |
| E9 | 公有 API + 无权限 — 返回 403 | IsPublicAPIFromCtx:true；OperationCheckAny 返回 false | 同上 + UserID:no-perm | HTTP 403 | 不继续查询 fileRepo/assetStore |
| E10 | 内部 API — 跳过权限校验 | IsPublicAPIFromCtx:false | BusinessDomainID:bd-1 SkillID:skill-1 response_mode:url | 正常返回 | AuthService 不被调用 |

---

## 2. ReadManagementFile

**Endpoint:** `POST /v1/skills/{skill_id}/management/files/read`
**Interface:** `ReadManagementFile(ctx, req) → (resp, err)`
**Req header:** `X-Business-Domain`, `User-ID`(public) | **Req uri:** `skill_id` | **Req body:** `{"rel_path": "..."}`
**Logic source:** `skillRepository.SelectSkillByID` + `normalizeZipPath` + `skillFileIndex.SelectSkillFileByPath` + `assetStore.GetDownloadURL`

| ID | 场景 | 前置条件 | 请求参数 | 期望响应 | 验证点 |
|----|------|---------|---------|---------|--------|
| E11 | 有效文件路径 — 返回 presigned URL 和元信息 | zip 注册；`file_index` 存在 `scripts/main.py` 记录 `{size:1024, mime_type:"text/x-python", file_type:"script"}` | `SkillID: "skill-1"` `RelPath: "scripts/main.py"` | `url: "https://..."` `mime_type: "text/x-python"` `file_type: "script"` `size: 1024` | 元信息与 file_index 一致 |
| E12 | 路径穿越 — 返回 400 | 任意 Skill | `RelPath: "../../etc/passwd"` | HTTP 400 `"invalid skill file path"` | normalizeZipPath 先拦截 |
| E13 | 文件不存在 — 返回 404 | zip 注册；`file_index` 无 `missing.py` 记录 | `RelPath: "missing.py"` | HTTP 404 | 查询 file_index 后返回 `nil` |
| E14 | 公有 API + 无权限 — 返回 403 | `IsPublicAPIFromCtx: true`；无 view/modify 权限 | `BusinessDomainID: "bd-1"` `UserID:"no-perm"` `SkillID:"skill-1"` `RelPath:"scripts/main.py"` | HTTP 403 | 同 E7 权限校验模式 |

---

## 3. DownloadManagementSkill

**Endpoint:** `GET /v1/skills/{skill_id}/management/download`
**Interface:** `DownloadManagementSkill(ctx, req) → (resp, err)`
**Req header:** `X-Business-Domain`, `User-ID`(public) | **Req uri:** `skill_id`
**Logic source:** `skillRepository.SelectSkillByID` + `skillFileIndex.SelectSkillFileBySkillID` + `buildArchiveFromFiles` (OSS Download → ZIP)

| ID | 场景 | 前置条件 | 请求参数 | 期望响应 | 验证点 |
|----|------|---------|---------|---------|--------|
| E15 | zip 注册 — 下载完整 ZIP | `file_index` 含 SKILL.md、scripts/main.py 等多条；OSS 各文件均可下载 | `SkillID: "skill-1"` | ZIP 二进制；文件名 `"{skill_name}.zip"`；ZIP 内路径与 `RelPath` 一致 | 包含所有文件 |
| E16 | content 注册 — 下载仅含 SKILL.md 的 ZIP | `file_index` 仅含 SKILL.md；OSS 可下载 | 同上 | ZIP 二进制；仅含 SKILL.md | ZIP 内仅 1 个文件 |
| E17 | OSS 下载部分失败 — 返回错误 | `file_index` 有记录；某文件 OSS `Download` 返回错误 | 同上 | 返回 error；不返回 ZIP | buildArchiveFromFiles 失败冒泡 |

---

## 4. 辅助函数

### 4.1 detectSkillFileType

**Signature:** `detectSkillFileType(skill *SkillRepositoryDB) string`
**Source:** `mgmt_reader.go`

| ID | 场景 | 输入 (skill 关键字段) | 预期输出 | 逻辑说明 |
|----|------|---------------------|---------|---------|
| E18 | 多条 manifest 记录 → "zip" | `FileManifest: [{"rel_path":"SKILL.md"},{"rel_path":"scripts/main.py"}]` | `"zip"` | 多条 → zip |
| E19 | 空 manifest → "content" | `FileManifest: ""`, `SkillContent: "body text"` | `"content"` | manifest 为空 → content |
| E20 | 仅 SKILL.md + 非空 Content → "content" (FR-5 场景) | `FileManifest: [{"rel_path":"SKILL.md"}]`, `SkillContent: "body text"` | `"content"` | 仅 1 条且为 SKILL.md + content 非空 → content |
| E21 | 全空 → "content" | `FileManifest: ""`, `SkillContent: ""` | `"content"` | 兜底默认 content |

### 4.2 buildArchiveFromFiles

**Signature:** `buildArchiveFromFiles(ctx, assetStore, skill, files) → (skill, zipName, content, err)`

| ID | 场景 | 前置条件 | 输入 | 期望输出 | 验证点 |
|----|------|---------|------|---------|--------|
| E39 | 正常构建 ZIP | OSS Download 返回每个文件的内容 | `files: [{RelPath:"SKILL.md",StorageKey:"key1"},{RelPath:"scripts/main.py",StorageKey:"key2"}]` | `zipName: "test-skill.zip"` `content: []byte(ZIP)` | ZIP 内路径与 RelPath 一致 |
| E40 | OSS 下载部分失败 — 返回错误 | 某文件 OSS Download 返回 error | 同上（其中一个文件下载失败） | `err != nil`, `content == nil` | 第一个失败即返回 |

### 4.3 updateFrontmatterNameDesc

**Signature:** `updateFrontmatterNameDesc(rawMD, newName, newDesc string) (string, error)`
**Source:** `registry.go` (FR-6)

| ID | 场景 | 输入 (rawMD frontmatter + body) | newName / newDesc | 预期输出 | 验证点 |
|----|------|--------------------------------|-------------------|---------|--------|
| E29 | 只替换 name/description，保留自定义字段 | `name: original\ndescription: original desc\nicon: robot\ntags: [ai, nlp]\n---\n# Body` | `"new-name"`, `"new desc"` | `name: "new-name"`, `description: "new desc"`, `icon: robot`, `tags: [ai, nlp]`, body 不变 | 自定义字段保留 |
| E30 | 空字符串不替换 | `name: original\ndescription: original desc` | `""`, `""` | name/description 保持不变 | 原样返回 |
| E31 | 无效格式返回错误 | 纯文本（无 `---` 分隔符） | `"any"`, `"any"` | error: `"missing frontmatter"` | 错误提示 |

---

## 5. FR-5：Content 注册 OSS 补齐（parser.go）

**Function:** `parseRegisterReq` content 分支
**Effect:** content 注册时从请求体的原始 SKILL.md 构建 `asset` + `file_summary`，经 `persistSkillAssets` 写入 OSS + `skill_file_index`

| ID | 场景 | 前置条件 | 输入 | 中间产物 | 最终状态（经 persist） |
|----|------|---------|------|---------|---------------------|
| E22 | content 注册返回 asset 和 file_summary | `FileType: "content"`；请求体为合法 SKILL.md（含 frontmatter + body） | `req.File: []byte("---\nname: my-skill\n...\n---\nbody text")` | `files: [{RelPath:"SKILL.md",FileType:"reference",MimeType:"text/markdown"}]`; `assets: [{RelPath:"SKILL.md",Content:[]byte(raw)}]` | files 含 1 条 SKILL.md；assets 含原始全文 |
| E23 | asset 被 persistSkillAssets 持久化到 OSS + file_index | 承接 E20 的 assets 返回值；调用方 `RegisterSkill`/`UpdateSkillPackage` 调用 `persistSkillAssets` | E20 的 `assets` | — | OSS 中 `{bucket}/{skill_id}/{version}/SKILL.md` 存在；`skill_file_index` 写入 SKILL.md 记录 |

---

## 6. FR-6：元数据编辑时 OSS SKILL.md 同步重写（registry.go）

**Trigger:** `UpdateSkillMetadata` 事务提交成功后，检测 name/desc 变更 → `rewriteSkillMDFrontmatter`
**Rewrite function:** `rewriteSkillMDFrontmatter(skillID, version, newName, newDesc)` → 查 file_index → OSS Download → updateFrontmatterNameDesc → OSS Upload
**Non-blocking:** 重写失败仅记日志，不阻塞 `UpdateSkillMetadata` 的主流程返回

| ID | 场景 | 前置条件 | 输入/变更 | 预期行为 | 验证点 |
|----|------|---------|----------|---------|--------|
| E24 | name 变更 — OSS 重写 | `skill.Name: "old-name"`, `req.Name: "new-name"`, `req.Description` 与 DB 相同；`file_index` 有 SKILL.md 记录 | `needsRewrite: true` | OSS 中 SKILL.md 的 frontmatter `name` → "new-name"；`description` 和其他字段不变 | rewriteSkillMDFrontmatter 被调用 |
| E25 | description 变更 — OSS 重写 | `skill.Description: "old desc"`, `req.Description: "new desc"`, `req.Name` 与 DB 相同 | `needsRewrite: true` | OSS 中 `description` → "new desc"；`name` 不变 | 同上 |
| E26 | name 和 description 均未变更 — 不触发 | `req.Name == skill.Name`, `req.Description == skill.Description` | `needsRewrite: false` | `rewriteSkillMDFrontmatter` 不被调用 | 跳过重写 |
| E27 | SKILL.md 不在 OSS — 日志记录，不阻塞 | `file_index` 无 SKILL.md 记录；`needsRewrite: true` | 重写流程第一步失败 | `rewriteSkillMDFrontmatter` 返回 error → 日志记录；`UpdateSkillMetadata` 返回成功 | 主流程不返回该 error |
| E28 | OSS 下载失败 — 日志记录，不阻塞 | SKILL.md 在 `file_index` 中；OSS `Download` 返回 error | 重写流程第二步失败 | 同 E25 — 日志记录，主流程返回成功 | 同上 |

---

## 7. parseRegisterReq（解析器）

**Signature:** `parseRegisterReq(ctx, req) → (skill, files, assets, err)`
**Logic source:** `parser.go`
**FileType branch:** "content" / "zip" / else → error

| ID | 场景 | 前置条件 | 输入 | 期望输出 | 验证点 |
|----|------|---------|------|---------|--------|
| E32 | content 合法 — 正确解析 | 含有效 frontmatter 的 content 请求 | `FileType: "content"` `File: []byte("---\nname: my-skill\nversion: 1.0.0\n---\nbody")` | `skill.Name: "my-skill"` `skill.Version: "{uuid}"`(非原始 version) `skill.SkillContent: "body"` `skill.Status: "unpublish"` | Version 被重写为 UUID |
| E33 | zip 合法 — 正确解析 | zip 内含 SKILL.md + 其他文件 | `FileType: "zip"` `File: []byte(zip)` zip 中含 `SKILL.md`, `scripts/main.py` | `skill.Name: frontmatter.name` `files: [SKILL.md, scripts/main.py]` `assets: [file contents]` | files 和 assets 数量与 zip 内文件一致 |
| E34 | zip 缺少 SKILL.md — 返回错误 | zip 不含 SKILL.md | 同上（无 SKILL.md 的 zip） | error: `"SKILL.md not found"` | 早期校验拦截 |
| E35 | zip 含路径穿越文件 — 返回错误 | zip 含 `../secret.txt` | 同上（含穿越路径的 zip） | error: `"invalid skill file path"` | normalizeZipPath 拦截 |
| E36 | 不支持的文件类型 — 返回错误 | `FileType: "exe"` | `FileType: "exe"` | error: `"unsupported file type"` | switch default 分支 |
| E37 | content 缺少 `---` 分隔符 — 返回错误 | 不含 frontmatter 分隔符 | `File: []byte("plain text without frontmatter")` | error: `"missing frontmatter"` | parseSkillContent 解析失败 |
| E38 | frontmatter 缺少必填字段 — 返回错误 | 只含 description 不含 name | `File: []byte("---\ndescription: test\n---\nbody")` | validator error: `"invalid skill frontmatter"` | 必填字段校验 |

---

## 用例汇总

| ID | API / 函数 | 场景 | 核心断言 | 优先级 |
|----|-----------|------|---------|--------|
| E1 | GetManagementContent | zip 注册正常 | url 非空 + files 完整 | P0 |
| E2 | GetManagementContent | content 注册无 OSS | content 内联 + url 空 | P0 |
| E3 | GetManagementContent | content 注册有 OSS | url + content 均非空 | P0 |
| E4 | GetManagementContent | 已删除 | 404 | P0 |
| E5 | GetManagementContent | 不存在 | 404 | P0 |
| E6 | GetManagementContent | 公有 API 有权限 | 200 | P0 |
| E7 | GetManagementContent | 公有 API 无权限 | 403 | P0 |
| E8 | GetManagementContent | 内部 API | 跳过 auth | P1 |
| E9 | ReadManagementFile | 有效路径 | presigned URL + 元信息 | P0 |
| E10 | ReadManagementFile | 路径穿越 | 400 | P0 |
| E11 | ReadManagementFile | 文件不存在 | 404 | P0 |
| E12 | ReadManagementFile | 公有 API 无权限 | 403 | P0 |
| E13 | DownloadManagementSkill | zip 注册 | ZIP 包含全部文件 | P0 |
| E14 | DownloadManagementSkill | content 注册 | ZIP 仅含 SKILL.md | P1 |
| E15 | DownloadManagementSkill | OSS 下载失败 | 返回 error | P1 |
| E16 | detectSkillFileType | 多文件 manifest | "zip" | P0 |
| E17 | detectSkillFileType | 空 manifest | "content" | P0 |
| E18 | detectSkillFileType | 仅 SKILL.md manifest + content | "content" | P0 |
| E19 | detectSkillFileType | 全空 | "content" | P0 |
| E20 | FR-5: parseRegisterReq content 分支 | content 注册返回 assets | 1 条 SKILL.md asset | P0 |
| E21 | FR-5: persistSkillAssets | content 持久化 | OSS + file_index | P0 |
| E22 | FR-6: rewriteSkillMDFrontmatter | name 变更 | OSS 重写 | P0 |
| E23 | FR-6: rewriteSkillMDFrontmatter | desc 变更 | OSS 重写 | P0 |
| E24 | FR-6: rewriteSkillMDFrontmatter | 无变更 | 不触发 | P0 |
| E25 | FR-6: rewriteSkillMDFrontmatter | SKILL.md 不在 OSS | 日志不阻塞 | P0 |
| E26 | FR-6: rewriteSkillMDFrontmatter | OSS 下载失败 | 日志不阻塞 | P1 |
| E27 | updateFrontmatterNameDesc | 保留自定义字段 | 仅替换 name/desc | P0 |
| E28 | updateFrontmatterNameDesc | 空字符串 | 不替换 | P1 |
| E29 | updateFrontmatterNameDesc | 无效格式 | 错误"missing frontmatter" | P1 |
| E30 | parseRegisterReq | 有效 content | 正确解析 | P0 |
| E31 | parseRegisterReq | 有效 zip | 正确解析 | P0 |
| E32 | parseRegisterReq | zip 缺 SKILL.md | 错误 | P0 |
| E33 | parseRegisterReq | zip 路径穿越 | 错误 | P0 |
| E34 | parseRegisterReq | 不支持的类型 | 错误 | P0 |
| E35 | parseSkillContent | 缺少分隔符 | 错误"missing frontmatter" | P0 |
| E36 | parseSkillContent | 少必填字段 | validator 错误 | P0 |
| E37 | buildArchiveFromFiles | 正常构建 | ZIP 内路径正确 | P1 |
| E38 | buildArchiveFromFiles | 部分下载失败 | 返回 error | P1 |

---

## 当前代码覆盖状态

| 功能 | 已覆盖 | 缺失 |
|------|--------|------|
| GetManagementContent | E1, E2, E4, E5, E6, E7 | E3, E8 |
| ReadManagementFile | E9, E10, E11 | E12 |
| DownloadManagementSkill | E13 (依赖 buildArchiveFromFiles) | E14, E15 |
| detectSkillFileType | E16, E17, E18, E19 | — |
| FR-5: parser content 分支 | E20 (parser_test.go) | E21 (集成验证) |
| FR-6: rewrite | 无 | E22–E29 |
| parseRegisterReq | E30, E32, E33, E31 | E34, E35, E36 |
| buildArchiveFromFiles | E37 | E38 |
| updateFrontmatterNameDesc | 无 | E27, E28, E29 |

> 注：E21 是集成测试场景，需要真实 DB 和 OSS。单元测试中用 mock 验证 `persistSkillAssets` 被正确调用即可。
