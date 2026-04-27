# 团队研发管理流程

中文 | [English](WORKFLOW.md)

本文档定义了 KWeaver 团队的研发协作规范，覆盖 Issue 管理、Feature 追踪、设计文档管理，以及团队通知流程。每个规则都有明确的操作步骤和文件路径，确保可落地执行。

---

## 📋 目录

- [Issue 管理](#-issue-管理)
- [Feature 追踪：Issue → Branch → 设计文档](#-feature-追踪issue--branch--设计文档)
- [设计文档规范](#-设计文档规范)
- [PR 与合并流程](#-pr-与合并流程)
- [邮件通知流程](#-邮件通知流程)

---

## 🗂 Issue 管理

### Issue 类型与 Label

| 类型 | Label | 是否需要设计文档 | 说明 |
| --- | --- | --- | --- |
| Bug 报告 | `type: bug` | 否（复杂 bug 可选） | 功能异常、错误行为 |
| Feature 申请 | `type: feature` | **必须** | 新功能或功能增强 |
| 任务 | `type: task` | 推荐 | 工程任务、调研、重构 |
| 文档 | `type: docs` | 否 | 文档缺失或改进 |

### 优先级标签

| 优先级 | Label | 响应时效 |
| --- | --- | --- |
| 紧急 | `priority: critical` | 24 小时内响应，当前 Sprint 必须完成 |
| 高 | `priority: high` | 当前 Sprint 必须完成 |
| 中 | `priority: medium` | 近 1–2 个 Sprint 内完成 |
| 低 | `priority: low` | 未来排期，无强制时限 |

### Issue 生命周期

```text
Open → Triaged → In Progress → In Review → Done
         │                                    │
         └─────────── Closed (Won't Fix) ─────┘
```

| 状态 | 操作说明 |
| --- | --- |
| `Open` | 已创建，待分配 |
| `Triaged` | 已评估优先级、已分配 Assignee 和 Milestone |
| `In Progress` | 已创建分支和设计文档，正在开发（**需在 Issue 中更新追踪信息**，见下节） |
| `In Review` | PR 已提交，待 Code Review |
| `Done` | PR 已合并，Issue 自动或手动关闭 |
| `Closed (Won't Fix)` | 决定不做，评论中说明原因 |

### Issue 分配规范

- Issue 创建后 **2 个工作日** 内完成 Triage（分配 Assignee、Priority、Milestone）
- 跨模块 Issue 需 `@` 对应模块负责人并说明期望
- 超过 30 天无进展的 Issue 需重新评估或关闭

---

## 🔗 Feature 追踪：Issue → Branch → 设计文档

这是本规范的核心。每一个 Feature（`type: feature`）从 Issue 到代码合并，必须完整经过以下步骤，并在 **Issue 评论** 中维护追踪信息。

### 完整流程

```text
1. 创建 Issue
      │
      ▼
2. Triage：分配 Assignee、Priority、Milestone
      │
      ▼
3. 创建设计文档（docs/design/{module}/features/{issue-id}-{desc}.md）
      │
      ▼
4. 创建分支（feature/{issue-id}-{desc}）
      │
      ▼
5. 在 Issue 中更新追踪评论（branch + 设计文档链接）
      │
      ▼
6. 开发 + 更新设计文档
      │
      ▼
7. 提交 PR（关联 Issue，关联设计文档）
      │
      ▼
8. Code Review（含文档审查）
      │
      ▼
9. 合并 → 更新设计文档状态为 Implemented
```

### 步骤 3：创建设计文档

**文件路径规则：**

```
docs/design/{module}/features/{issue-id}-{short-desc}.md
```

| 占位符 | 说明 | 示例 |
| --- | --- | --- |
| `{module}` | 所属模块名，与代码目录保持一致 | `auth`、`knowledge-graph`、`data-agent` |
| `{issue-id}` | GitHub Issue 编号（不含 `#`） | `123` |
| `{short-desc}` | Issue 标题的短横线小写形式，不超过 5 个词 | `add-oauth-support` |

**示例：**

```
docs/design/auth/features/123-add-oauth-support.md
docs/design/knowledge-graph/features/456-batch-import-nodes.md
docs/design/data-agent/features/789-streaming-response.md
```

**目录结构：**

```
docs/
└── design/
    ├── auth/
    │   └── features/
    │       └── 123-add-oauth-support.md
    ├── knowledge-graph/
    │   └── features/
    │       └── 456-batch-import-nodes.md
    └── data-agent/
        ├── features/
        │   └── 789-streaming-response.md
        └── adr/                          ← 架构决策记录（见下节）
            └── 0001-use-opensearch.md
```

### 步骤 4：创建分支

**分支命名规则：**

```
feature/{issue-id}-{short-desc}
```

分支名中的 `{issue-id}` 和 `{short-desc}` 必须与设计文档文件名完全一致：

```bash
# 示例
git checkout -b feature/123-add-oauth-support
git checkout -b feature/456-batch-import-nodes
```

### 步骤 5：在 Issue 中更新追踪信息

开始开发后，必须在 Issue 中发一条评论，记录追踪信息（**每个 Issue 只需一条，后续更新此评论**）：

```markdown
## 📌 开发追踪

| 项目 | 内容 |
| --- | --- |
| **分支** | `feature/123-add-oauth-support` |
| **设计文档** | [docs/design/auth/features/123-add-oauth-support.md](../docs/design/auth/features/123-add-oauth-support.md) |
| **状态** | In Progress |
| **负责人** | @username |
| **预计完成** | YYYY-MM-DD |
```

> **说明**：PR 合并后将此评论的状态更新为 `Done`，并补充 PR 链接。

---

## 📄 设计文档规范

### 文档 Frontmatter（元信息头）

每份设计文档必须以 YAML frontmatter 开头，记录关键元信息：

```markdown
---
issue: "#123"
branch: "feature/123-add-oauth-support"
module: "auth"
status: "draft"          # draft | in-review | approved | implemented
author: "@username"
created: "2026-03-16"
pr: ""                   # PR 合并后填写，如 "#456"
---
```

| 字段 | 必须 | 说明 |
| --- | --- | --- |
| `issue` | 是 | 关联的 GitHub Issue 编号 |
| `branch` | 是 | 对应的开发分支名 |
| `module` | 是 | 所属模块 |
| `status` | 是 | 文档/功能状态，见下表 |
| `author` | 是 | 主要负责人 |
| `created` | 是 | 文档创建日期 |
| `pr` | 否 | PR 合并后填写 |

**Status 取值：**

| 值 | 说明 |
| --- | --- |
| `draft` | 设计中，尚未评审 |
| `in-review` | 已提交 PR，正在 Review |
| `approved` | 已评审通过，可以实施 |
| `implemented` | 已合并，功能已上线 |

### 设计文档模板

```markdown
---
issue: "#{issue-id}"
branch: "feature/{issue-id}-{short-desc}"
module: "{module}"
status: "draft"
author: "@username"
created: "YYYY-MM-DD"
pr: ""
---

# Feature #{issue-id}: {功能标题}

## 背景与目标

描述功能背景、用户痛点，以及本次开发要达成的目标。

## 方案设计

### 概要

简要描述整体方案。

### API 变更（如有）

```http
POST /api/v1/{endpoint}
Content-Type: application/json

{
  "field": "value"
}
```

响应：

```json
{
  "id": "xxx",
  "status": "ok"
}
```

### 数据库变更（如有）

描述新增或修改的表/字段，并提供迁移脚本路径。

### 关键流程

用文字或伪代码描述核心逻辑流程。

## 验收标准

- [ ] 功能条件 1
- [ ] 功能条件 2
- [ ] 测试覆盖率满足要求

## 测试策略

描述需要补充的单元测试、集成测试或 AT 用例。

## 影响分析

- **向后兼容性**：是/否，说明
- **依赖变更**：是/否，说明
- **性能影响**：说明（如无影响可略）

## 参考资料

- 相关 Issue/PR 链接
- 相关文档链接
```

### Bug 复杂分析文档（可选）

对于复杂 Bug（如涉及多模块、需要根因分析），可在以下路径创建分析文档：

```
docs/design/{module}/bugs/{issue-id}-{short-desc}.md
```

模板相对简单，包含：问题描述、根因分析、修复方案、验证方式。

### ADR（架构决策记录）

对于涉及重大设计决策的 Feature，在设计文档中说明决策后，可选择同步到 ADR 存档：

**路径：** `docs/design/{module}/adr/NNNN-{short-title}.md`

**文件命名：** 序号从 `0001` 开始，按模块独立计数。

```markdown
---
number: "0001"
module: "{module}"
status: "accepted"       # proposed | accepted | deprecated | superseded
date: "YYYY-MM-DD"
related-issue: "#123"
---

# ADR-{module}-0001: {决策标题}

## 背景

描述做出此决策的背景与约束条件。

## 决策

我们决定采用 ___。

## 原因

- 原因 1
- 原因 2

## 后果

**正面影响：**
- ...

**负面影响（权衡）：**
- ...

## 替代方案

描述考虑过但未采用的方案及原因。
```

---

## 🔀 PR 与合并流程

### PR 描述模板

PR 描述必须包含以下内容，以完成 Issue → Branch → 设计文档的三向关联：

```markdown
## 变更描述

简要描述本次变更内容。

## 关联信息

| 项目 | 内容 |
| --- | --- |
| **Issue** | Closes #123 |
| **设计文档** | [docs/design/auth/features/123-add-oauth-support.md](../docs/design/auth/features/123-add-oauth-support.md) |
| **分支** | `feature/123-add-oauth-support` |

## 变更类型

- [ ] Bug 修复
- [ ] 新功能
- [ ] 文档更新
- [ ] 重构

## 测试说明

描述如何验证本次变更（测试命令、手动验证步骤等）。

## 合并前检查清单

- [ ] 设计文档已更新（status 改为 `in-review`，`pr` 字段已填写）
- [ ] CHANGELOG.md 已更新（在 `[Unreleased]` 节）
- [ ] API 文档已更新（如有 API 变更）
- [ ] 测试已添加/更新，本地测试通过
- [ ] 无破坏性变更，或已在 CHANGELOG 中标注
```

### Code Review 文档检查项

Reviewer 在 Code Review 时需确认：

- [ ] 设计文档与实际实现一致
- [ ] API 文档已同步更新
- [ ] CHANGELOG.md 已记录面向用户的变更
- [ ] 设计文档 frontmatter 的 `pr` 字段和 `status` 已更新

### 合并后操作

PR 合并后由合并者或 Assignee 完成：

1. 将设计文档的 `status` 更新为 `implemented`
2. 更新 Issue 追踪评论，将状态改为 `Done` 并补充 PR 链接

---

## 📧 邮件通知流程

### 通知触发时机

| 事件 | 收件人 | 触发方式 |
| --- | --- | --- |
| 新正式版本发布 | 全体成员 + 外部订阅者 | CI 自动 / 手动 |
| RC 版本发布 | 内部测试团队 | CI 自动 / 手动 |
| 代码冻结开始 | 全体研发成员 | 手动（Release Manager） |
| Critical Bug 修复 | 受影响模块负责人 + 测试团队 | 手动 |
| Sprint 开始 | 本 Sprint 参与成员 | 手动（项目负责人） |

### 邮件模板

#### Release 发布通知

```
主题：[KWeaver] vX.Y.Z 正式发布

KWeaver vX.Y.Z 已正式发布！

## 主要变更

### ✨ 新增
- 功能描述（#Issue编号，设计文档链接）

### 🐛 修复
- 修复描述（#Issue编号）

### ⚠️ Breaking Changes（如有）
- 描述变更及迁移方式

## 下载
- GitHub Releases: <链接>
- Docker: docker pull kweaver/kweaver:vX.Y.Z

## 完整变更日志
<CHANGELOG 链接>
```

#### RC 版本通知

```
主题：[KWeaver] vX.Y.Z-rc.N 测试版发布，邀请测试反馈

KWeaver vX.Y.Z-rc.N 已发布，请测试团队进行验证。

## 测试范围
- 变更列表（关联设计文档链接）

## 已知问题
- （如有）

## 反馈方式
在 GitHub Issue 中提交问题，标注 milestone: vX.Y.Z

## 测试截止
YYYY-MM-DD
```

#### 代码冻结通知

```
主题：[KWeaver] vX.Y.Z 代码冻结通知

vX.Y.Z release 分支已创建，代码冻结自 YYYY-MM-DD 开始。

## 冻结规则
- ✅ 允许：Bug 修复、文档更新、版本号更新
- ❌ 禁止：新功能、重构、依赖升级

## 冻结期间合并申请
如需合并非 Bug 修复内容，请回复此邮件并说明原因，由 Release Manager 审批。

## 预计正式发布
YYYY-MM-DD
```

### 发送规范

| 规范项 | 要求 |
| --- | --- |
| 发送方式 | 团队邮件列表，重要 Release 可由 CI 自动触发 |
| 主题前缀 | 统一使用 `[KWeaver]`，便于过滤和归档 |
| 语言 | 中文为主，重要 Release 可附英文版 |
| 抄送 | **所有邮件必须抄送测试负责人**；Release 通知同时抄送项目负责人 |
| 附件 | 禁止附件，统一使用链接引用文档或制品 |

---

## 📚 相关资源

- [贡献指南](CONTRIBUTING.zh.md)
- [发布规范](RELEASE.zh.md)
- [测试规范](TESTING.zh.md)
- [架构规范](ARCHITECTURE.zh.md)

---

*最后更新：2026-03-16*
