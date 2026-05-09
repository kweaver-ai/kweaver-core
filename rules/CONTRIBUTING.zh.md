# 贡献指南

中文 | [English](CONTRIBUTING.md)

感谢你对 KWeaver 项目的兴趣！我们欢迎所有形式的贡献，包括修复 Bug、提出新特性、编写文档、回答问题等。

请在提交贡献前阅读本文，确保流程一致、提交规范统一。

---

## 🏗 仓库结构

KWeaver Core 是一个 **monorepo**（[`kweaver-ai/kweaver-core`](https://github.com/kweaver-ai/kweaver-core)），平台后端各模块统一存放于此。请根据要修改的组件，进入对应目录：

| 模块 | 路径 | 描述 |
| --- | --- | --- |
| **AI Data Platform（ADP）** | [`adp/`](../adp) | 本体引擎（`adp/bkn`）、Context Loader（`adp/context-loader`）、Dataflow（`adp/dataflow`）、Execution Factory（`adp/execution-factory`）、VEGA 数据虚拟化（`adp/vega`） |
| **Decision Agent** | [`decision-agent/`](../decision-agent) | `agent-backend/` 下的 Agent Executor / Factory / Memory |
| **Trace AI** | [`trace-ai/`](../trace-ai) | Agent 可观测与 OpenTelemetry Collector Chart |
| **Infra** | [`infra/`](../infra) | `mf-model-manager`（模型注册）、`oss-gateway-backend`、`sandbox` 运行时 |
| **BKN 示例** | [`bkn/`](../bkn) | 业务知识网络示例（如 `smart_home_supply_chain`） |
| **示例** | [`examples/`](../examples) | 端到端 CLI 示例（数据库 / CSV / Action） |
| **文档与站点** | [`help/`](../help)、[`website/`](../website) | 中英双语产品文档与对外 docs 站点 |
| **部署** | [`deploy/`](../deploy) | Kubernetes 一键 `deploy.sh`（K8s + KWeaver Core Charts） |

与本后端配套的对外 CLI / SDK 在独立仓库维护：

| 仓库 | 用途 |
| --- | --- |
| [`kweaver-ai/kweaver-sdk`](https://github.com/kweaver-ai/kweaver-sdk) | `kweaver` CLI + TypeScript / Python SDK + Agent Skills（`kweaver-core`、`create-bkn`） |
| [`kweaver-ai/kweaver-admin`](https://github.com/kweaver-ai/kweaver-admin) | `kweaver-admin` 平台管理员 CLI + Agent Skill |

> **说明**：每个模块都有自己的 README（通常还包括 `AGENTS.md` / `CLAUDE.md`），里面给出本模块语言对应的构建、测试与开发循环命令，请在动手前先看。

---

## 📖 DOCUMENTATION（文档放置规范）

按**读者**与**范围**选择目录：

1. **[`docs/`](../docs)（仓库根）** — 系统级架构、跨子系统的整体设计、影响整个 Core 平台的技术决策。
2. **各子模块下的 `docs/`** — 仅属于该模块的设计与技术决策（例如 [`adp/bkn/docs/`](../adp/bkn/docs)、各 `decision-agent` 服务下的设计文档目录等）。
3. **[`help/{en,zh}/manual/`](../help/zh/manual)** — 面向用户/运维的**使用手册与参考**（按产品子域分文件；`install.md` / `quick-start.md` 与 `manual/` 同级放在 `help/{语言}/` 下）。
4. **[`help/{en,zh}/cookbook/`](../help/zh/cookbook)** — **Cookbook**：可复制的场景化操作步骤。新写一篇请直接复制 [`_TEMPLATE.md`](../help/zh/cookbook/_TEMPLATE.md)，参考已写好的示例 [`cookbook_example.md`](../help/zh/cookbook/cookbook_example.md)，并在 [`cookbook/README.md`](../help/zh/cookbook/README.md) 索引表加一行（英文在 `help/en/cookbook/`）。

面向最终用户的 Help 正文：**英文**写在 `help/en/`，**中文**写在 `help/zh/`。

---

## 🧩 贡献方式类型

你可以通过以下方式参与：

- 🐛 **报告 Bug**: 帮助我们识别和修复问题
- 🌟 **提出新特性**: 建议新功能或改进
- 📚 **改进文档**: 完善文档、示例或教程
- 🔧 **修复 Bug**: 为现有问题提交补丁
- 🚀 **实现新功能**: 构建新功能
- 🧪 **补充测试**: 提高测试覆盖率
- 🎨 **优化代码结构**: 重构代码，提高可维护性

---

## 🗂 Issue 规范（Bug & Feature）

### 1. Bug 报告格式

请在提交 Bug 时提供以下信息：

- **版本号 / 环境**：
  - KWeaver Core 版本（`git describe --tags` 或 `VERSION` 文件，如 `v0.6.0`）
  - 受影响的模块（如 `adp/bkn`、`decision-agent/agent-backend/agent-executor`、`infra/sandbox`）
  - 运行时（Java / Go / Python / Node 及版本，如 JDK 17、Go 1.23、Python 3.11）
  - 操作系统（Linux 发行版与内核 / macOS / Windows）
  - 集群形态（单机 K3s / kubeadm / 托管 K8s）以及安装方式（`deploy.sh kweaver-core install [--minimum]`）
  - 涉及的存储/中间件（MariaDB / DM8、OpenSearch、Redis、Kafka）

- **复现步骤**: 清晰、逐步的复现说明

- **期望结果 vs 实际结果**: 应该发生什么 vs 实际发生了什么

- **错误日志 / 截图**: 包含相关的错误消息、堆栈跟踪或截图

- **最小复现代码（MRC）**: 能够演示问题的最小代码示例

**Bug 报告模板示例：**

```markdown
**环境:**
- KWeaver Core: v0.6.0
- 模块: adp/bkn
- 运行时: JDK 17
- 操作系统: Linux Ubuntu 22.04
- 集群: 单机 K3s（deploy.sh kweaver-core install）
- 数据库: MariaDB 11.4

**复现步骤:**
1. 启动服务
2. 执行操作
3. 发生错误

**期望行为:**
操作应该成功完成

**实际行为:**
错误: "unexpected error"

**错误日志:**
[在此粘贴错误日志]
```

### 2. Feature 申请格式

请在 Issue 中描述：

- **背景 / 用途**: 为什么需要这个功能？它解决了什么问题？

- **功能期望**: 详细描述提议的功能

- **API 草案**（如适用）: 提议的 API 更改或新端点

- **潜在影响**: 对现有功能的潜在影响（向后兼容性）

- **实现方向**（可选）: 关于如何实现的建议

> **提示**：所有大的 Feature 需要先开 Issue 讨论，通过后再提 PR。

**Feature 申请模板示例：**

```markdown
**背景:**
目前，用户在更新后需要手动刷新知识网络。
此功能将自动化刷新过程。

**功能描述:**
添加自动刷新机制，当底层数据更改时更新知识网络。

**提议的 API:**
POST /api/v1/networks/{id}/auto-refresh
{
  "enabled": true,
  "interval": 300
}

**向后兼容性:**
这是一个新功能，不影响现有功能。
```

---

## 🔀 Pull Request（PR）流程

### 1. Fork 本仓库

Fork 本仓库到你的 GitHub 账户。

### 2. 创建新分支

从 `main`（或适当的基础分支）创建新分支：

```bash
git checkout -b feature/my-feature
# 或
git checkout -b fix/bug-description
# 或（模块作用域 + 关联 Issue：第二段常以「编号-简述」开头）
git checkout -b feature/agent-web/123-add-login
git checkout -b fix/bkn-backend/456-query-timeout
```

**分支命名规范：**

分支名称会在每次 Pull Request 时由 CI 自动校验（与 `.github/workflows/lint-branch-name.yml` 一致）。

**格式（类型前缀之后最多两段路径）**：

- `<类型>/<描述>`，例如 `feature/add-oauth-support`
- `<类型>/<Issue编号>-<描述>`（关联 Issue 时常用），例如 `fix/123-memory-leak`
- `<类型>/<模块>/<描述>`（可选，用于按子模块或目录划分），其中**描述**里可含 Issue 编号，例如 `feature/agent-web/123-add-login`、`fix/bkn-backend/456-query-timeout`

**不要**使用类型前缀之后**三层及以上**路径，例如 `feature/foo/bar/baz` 会校验失败（CI 要求类型后至多 2 个 path 段）。

**示例（含模块与 Issue）**：

| 分支名 | 说明 |
| --- | --- |
| `feature/studio/789-export-pipeline` | 功能 + 子模块 `studio`，Issue `#789` |
| `fix/ontology-query/404-id-not-exist` | 修复 + 子模块 `ontology-query`，Issue `#404` |
| `docs/rules/120-contributing-branch` | 文档 + 目录 `rules`，Issue `#120` |

| 分支类型 | 命名格式 | 说明 | 示例 |
| --- | --- | --- | --- |
| 功能分支 | `feature/*` 或 `feat/*` | 新功能开发 | `feature/add-oauth-support` |
| 修复分支 | `fix/*` | Bug 修复 | `fix/123-memory-leak-in-loader` |
| 紧急修复 | `hotfix/*` | 紧急生产修复 | `hotfix/critical-auth-bypass` |
| 文档分支 | `docs/*` | 文档更改 | `docs/update-api-reference` |
| 重构分支 | `refactor/*` | 代码重构 | `refactor/simplify-auth-flow` |
| 测试分支 | `test/*` | 添加或更新测试 | `test/add-unit-tests-for-loader` |
| 杂务分支 | `chore/*` | 维护任务 | `chore/upgrade-dependencies` |
| CI 分支 | `ci/*` | CI/CD 配置更改 | `ci/add-branch-name-lint` |
| 性能分支 | `perf/*` | 性能优化 | `perf/optimize-query-execution` |
| 构建分支 | `build/*` | 构建系统或依赖 | `build/update-go-module` |
| 样式分支 | `style/*` | 代码样式 / 格式化 | `style/fix-linter-warnings` |
| 回滚分支 | `revert/*` | 回滚之前的更改 | `revert/rollback-auth-change` |
| 发布分支 | `release/x.y.z`（或带预发布后缀） | 发布准备 | `release/1.2.0`、`release/1.2.0-rc.1` |

规则：
- 若关联 Issue，常见写法是 `<类型>/<N>-<描述>`（如 `fix/123-memory-leak`）；也可在「模块/描述」两段式里带上编号（如 `feature/agent-web/123-add-login`）
- 每一段（`/` 之间的部分）须为小写字母或数字开头，其余字符为连字符（`-`）、点（`.`）或下划线（`_`）等（与 CI 一致）
- 分支名必须以有效的类型前缀开头，后跟 `/`
- Bot 分支（`dependabot/*`、`renovate/*`）自动豁免

> **说明**：分支策略、版本规则和发布流程请参阅 [发布规范](RELEASE.zh.md)。

### 3. 进行更改

- 编写清晰、可维护的代码
- 遵循项目的代码结构和架构模式
- 添加适当的注释和文档
- 添加标准文件头（参见下方 [源代码文件头规范](#-源代码文件头规范)）

### 4. 编写测试

- 为新功能添加单元测试
- 确保现有测试仍然通过
- 争取良好的测试覆盖率
- 使用模块本身使用的测试框架。各模块 README / `AGENTS.md` 给出了规范命令，例如：
  - Java（Maven）模块：`mvn test`
  - Go 模块：`go test ./...`
  - Python 模块：`pytest`
  - Node / TypeScript 模块：`npm test`

### 5. 更新文档

- 如果你的更改影响面向用户的功能，请更新相关文档
- 如果修改了端点，请更新 API 文档
- 如果引入新功能，请添加示例
- 如适用，更新 CHANGELOG.md

#### README 规范

更新 README 文件时，请遵循以下规范：

- **默认语言**: `README.md` 应为英文（默认）
- **中文版本**: 中文文档应在 `README.zh.md` 中
- **保持同步**: 如果更新了 `README.md`，请同时更新 `README.zh.md`
- **结构一致**: 保持英文和中文版本的结构一致
- **链接更新**: 更新每个 README 文件顶部的语言切换链接：
  - 英文版: `[中文](README.zh.md) | English`
  - 中文版: `[中文](README.zh.md) | [English](README.md)`

**README 结构示例：**

```markdown
# 项目名称

[中文](README.zh.md) | [English](README.md)

[![License](...)](LICENSE.txt)
[![Go Version](...)](...)

简要描述...

## 📚 快速链接

- 文档、贡献指南等链接

## 主要内容

...
```

### 6. 提交更改

编写清晰、描述性的提交消息：

```bash
git commit -m "feat: 为知识网络添加自动刷新功能

- 添加自动刷新配置端点
- 实现后台刷新工作器
- 添加刷新功能的测试

Closes #123"
```

**提交消息格式：**

遵循 [Conventional Commits](https://www.conventionalcommits.org/zh-hans/) 规范：

| 类型 | 说明 |
| --- | --- |
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 仅文档更改 |
| `style` | 代码样式更改（格式化等） |
| `refactor` | 代码重构（非 feat/fix） |
| `perf` | 性能优化 |
| `test` | 添加或更新测试 |
| `build` | 构建系统或外部依赖 |
| `ci` | CI 配置更改 |
| `chore` | 其他维护任务 |
| `revert` | 回滚提交 |

> **说明**：详细的 Commit 规范和版本规则请参阅 [发布规范](RELEASE.zh.md)。

### 7. 保持分支与主分支同步

由于本项目要求线性历史，请在推送前将你的分支 rebase 到最新的 `main` 分支：

```bash
# 确保你在你的功能分支上
git checkout feature/my-feature

# 确保所有更改都已提交
git status  # 检查是否有未提交的更改

# 如果有未提交的更改，请先提交：
# git add .
# git commit -m "你的提交消息"

# 方式 1: 如果已配置 upstream，从 upstream 获取并 rebase
# git fetch upstream
# git rebase upstream/main

# 方式 2: 从 origin 获取最新更改并 rebase 到 origin/main
git fetch origin
git rebase origin/main

# 如果有冲突，解决后继续：
# 1. 修复冲突文件
# 2. git add <已解决的文件>
# 3. git rebase --continue

# 如果想中止 rebase：
# git rebase --abort

# 强制推送（rebase 后必需）
git push origin feature/my-feature --force-with-lease
```

> **注意**:
>
> - 使用 `--force-with-lease` 而不是 `--force`，以避免覆盖其他人的工作。
> - 确保在 rebase 前你在你的功能分支上。
> - 如果你想跟踪上游仓库，可以添加：`git remote add upstream https://github.com/kweaver-ai/kweaver-core.git`

### 8. 推送到你的 Fork

```bash
git push origin feature/my-feature
```

### 9. 创建 Pull Request

1. 转到 GitHub 上的原始仓库
1. 点击 "New Pull Request"
1. 选择你的 Fork 和分支
1. 填写 PR 模板，包括：
   - 更改描述
   - 相关 Issue 编号（如适用）
   - 测试说明
   - 截图（如果是 UI 更改）

**PR 检查清单：**

- [ ] 已完成自我审查
- [ ] 为复杂代码添加了注释
- [ ] 文档已更新
- [ ] 测试已添加/更新
- [ ] 所有测试通过
- [ ] 更改向后兼容（或提供了迁移指南）

---

## 📋 代码审查流程

1. **自动化检查**: PR 将通过 CI/CD 流水线进行检查
   - 单元测试
   - 构建验证

1. **审查**: 维护者将审查你的 PR
   - 及时处理审查意见
   - 进行请求的更改
   - 保持讨论建设性

1. **批准**: 一旦批准，维护者将合并你的 PR
   - PR 将使用 squash merge 或 rebase merge 合并，以保持线性历史
   - 请在请求审查前确保你的分支是最新的

---

## ⚙️ CI Workflow 规范

所有 GitHub Actions workflow 文件位于 `.github/workflows/`。GitHub **不支持**子目录——嵌套文件夹中的文件会被静默忽略。

### 文件命名规范

使用**分类前缀**对 workflow 文件进行分组，使相关文件在列表中自然排列在一起：

| 前缀 | 用途 | 示例 |
| --- | --- | --- |
| `lint-` | 代码 / Commit / 分支校验 | `lint-branch-name.yml`、`lint-commit.yml`、`lint-workflow-files.yml` |
| `ci-` | PR / push 上的构建、测试、类型检查、集成 | `ci-backend.yml`、`ci-website.yml` |
| `release-` | 构建与发布 | `release-agent-observability.yml` |
| `deploy-` | 部署任务 | `deploy-pages.yml` |
| `security-` | 供应链 / 应用安全扫描 | `security-codeql.yml`、`security-dependency-review.yml` |
| `automation-` | 仓库机器人、定时家务流 | `automation-stale.yml`、`automation-labeler.yml` |
| `reusable-` | 仅被其它 workflow 调用的入口（`on.workflow_call`） | `reusable-ci-go.yml` |

规则：
- 文件名必须使用小写 kebab-case，扩展名统一为 `.yml`
- 始终使用分类前缀，便于分组和查找
- 如果 workflow 内部引用了自身文件路径（如 `on.push.paths`），重命名时**必须**同时更新文件名和内部路径引用

### 可复用 Workflow 与 Composite Action

当 workflow 数量增多或出现共享逻辑时：
- **可复用 workflow** 放在 `.github/workflows/` 中，通过 `uses: ./.github/workflows/xxx.yml` 调用
- **Composite Action** 放在 `.github/actions/<name>/action.yml` 中，用于 step 级别的复用

---

## 📝 源代码文件头规范

本节定义了 **kweaver.ai** 开源项目中使用的标准源代码文件头。

目标是确保：

- 明确的版权归属
- 明确的许可证（Apache License 2.0）
- 一致且可读的文件文档

> **说明**：我们使用 "The kweaver.ai Authors" 而不是个人作者名。
> Git 历史记录已经追踪了所有贡献者，这种方式更易于维护。

### 标准文件头（Go / C / Java）

所有核心源文件使用以下文件头：

```go
// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.
```

### 各语言变体

#### Python

```python
# Copyright The kweaver.ai Authors.
#
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file in the project root for details.
```

#### JavaScript / TypeScript

```ts
/**
 * Copyright The kweaver.ai Authors.
 *
 * Licensed under the Apache License, Version 2.0.
 * See the LICENSE file in the project root for details.
 */
```

#### Shell

```bash
#!/usr/bin/env bash
# Copyright The kweaver.ai Authors.
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file in the project root for details.
```

#### HTML / XML

```html
<!--
  Copyright The kweaver.ai Authors.
  Licensed under the Apache License, Version 2.0.
  See the LICENSE file in the project root for details.
-->
```

### 派生或 Fork 的文件（可选）

如果文件最初来自其他项目，可以在许可证头后添加来源说明（仅用于关键文件）：

```go
// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.
//
// This file is derived from [original-project](https://github.com/org/repo)
```

这是可选的，但建议添加以保持透明度和社区信任。

### 适用范围

文件头**推荐**用于：

- 核心逻辑和业务代码
- 公共 API 和接口
- 库和 SDK
- CLI 工具和实用程序

文件头**可选**用于：

- 单元测试和测试夹具
- 示例和演示
- 生成的文件（protobuf、OpenAPI 等）
- 配置文件（YAML、JSON、TOML）
- 文档文件（Markdown 等）

### 为什么不写个人作者名？

遵循主流开源项目（Kubernetes、TensorFlow 等）的做法：

- **Git 历史**已经提供了所有贡献者的完整准确记录
- 个人作者列表**难以维护**，容易过时
- 使用 "The kweaver.ai Authors" 确保所有文件的**一致归属**
- 贡献者通过项目的 **CONTRIBUTORS** 文件和 git log 获得认可

### 许可证要求

所有仓库**必须**包含一个 `LICENSE` 文件，其中包含 Apache License 2.0 的完整文本。

### 指导原则

> 如果一个文件预计会被复用、fork 或长期维护，它就值得拥有一个清晰明确的文件头。

---

## 🏗 开发环境设置

### 环境要求

KWeaver Core 是多语言项目，**只需安装你要修改的模块所需的工具链**：

- **Git**（必备）
- **Java**（JDK 17+）+ Maven —— ADP 与 decision-agent 大多数后端模块
- **Go**（1.23+）—— `infra/oss-gateway-backend`、若干 CLI 与小型服务
- **Python**（3.11+）—— `infra/mf-model-manager` 等模型/数据组件
- **Node.js** —— kweaver 相关 CLI 需 **22+**（以 [npm 上 `kweaver-sdk`](https://www.npmjs.com/package/@kweaver-ai/kweaver-sdk) 的 `engines` 为准）；`website/` 以各包 `package.json` 的 `engines` 为准（如 >= 20）
- **Docker** + 一套 Kubernetes（单机 K3s / kubeadm / Docker Desktop）—— 端到端验证
- **MariaDB 11.4+**（或 DM8）、**OpenSearch 2.x**、**Redis**、**Kafka** —— 仅在调试相关服务时需要，通常由 `deploy.sh` 提供

各模块 `README.md` / `AGENTS.md` 中列出了该模块的精确依赖、构建命令与开发循环，**动手前请先看**。

### 本地开发

1. **克隆你 Fork 的 `kweaver-core`：**

   ```bash
   git clone https://github.com/YOUR_USERNAME/kweaver-core.git
   cd kweaver-core
   ```

2. **添加上游远程仓库：**

   ```bash
   git remote add upstream https://github.com/kweaver-ai/kweaver-core.git
   ```

3. **进入要修改的模块，按其 README 操作。** 例如：

   ```bash
   # Java 模块（Maven）
   cd adp/bkn/bkn-backend && mvn -DskipTests package

   # Go 模块
   cd infra/oss-gateway-backend && make build && make test

   # Python 模块
   cd infra/mf-model-manager && pip install -r requirements.txt && pytest

   # 文档站点
   cd website && npm install && npm run start
   ```

4. **（可选）拉起完整集群** 做端到端测试，使用 `deploy/`（详见 [`help/zh/install.md`](../help/zh/install.md)）：

   ```bash
   cd deploy
   ./deploy.sh kweaver-core install --minimum    # 快速体验
   # 完整安装：./deploy.sh kweaver-core install
   ```

---

## 🐛 报告安全问题

**请不要通过公共 GitHub Issues 报告安全漏洞。**

请通过 GitHub 内置的安全公告私密通道报告：

- [报告漏洞 — kweaver-ai/kweaver-core](https://github.com/kweaver-ai/kweaver-core/security/advisories/new)

我们会确认收到并与你协同修复。请在报告中包含：复现步骤、受影响版本（`git describe --tags`）以及你观察到的影响。

---

## ❓ 获取帮助

- **文档**: 查看 [README](README.zh.md) 和模块特定文档
- **Issues**: 在创建新 Issue 之前搜索现有 Issues
- **讨论**: 使用 GitHub Discussions 提问和讨论想法

---

## 📜 许可证

通过向 KWeaver 贡献，你同意你的贡献将在 Apache License 2.0 下许可。

---

感谢你为 KWeaver 做出贡献！🎉
