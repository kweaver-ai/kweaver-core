# 贡献指南

中文 | [English](CONTRIBUTING.md)

感谢你对 KWeaver 项目的兴趣！我们欢迎所有形式的贡献，包括修复 Bug、提出新特性、编写文档、回答问题等。

请在提交贡献前阅读本文，确保流程一致、提交规范统一。

---

## 🏗 子项目

KWeaver 是一个由多个子项目组成的开源生态。请根据你想贡献的组件，导航到相应的仓库：

| 子项目 | 描述 | 仓库地址 |
| --- | --- | --- |
| **AI Store** | AI 应用与组件市场 | [kweaver-ai/dip](https://github.com/kweaver-ai/dip)  |
| **Studio** | DIP Studio - 可视化开发与管理界面 | [kweaver-ai/studio](https://github.com/kweaver-ai/studio) |
| **Decision Agent** | 决策智能体 | [kweaver-ai/data-agent](https://github.com/kweaver-ai/data-agent) |
| **ADP** | AI Data Platform（智能数据平台）- 包含本体引擎、ContextLoader 和 VEGA 数据虚拟化引擎 | [kweaver-ai/adp](https://github.com/kweaver-ai/adp) |
| **Operator Hub** | 算子平台 - 负责算子管理与编排 | [kweaver-ai/operator-hub](https://github.com/kweaver-ai/operator-hub) |
| **Sandbox** | 沙箱运行环境 | [kweaver-ai/sandbox](https://github.com/kweaver-ai/sandbox) |

> **说明**：每个子项目都有自己的 README 和贡献指南。请参阅具体仓库获取详细的设置和开发说明。

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
  - Go 版本（如 Go 1.23.0）
  - 操作系统（Windows/Linux/macOS）
  - 数据库版本（MariaDB 11.4+ / DM8）
  - OpenSearch 版本（如适用）
  - 受影响的模块（如 ADP、Decision Agent、DIP Studio）

- **复现步骤**: 清晰、逐步的复现说明

- **期望结果 vs 实际结果**: 应该发生什么 vs 实际发生了什么

- **错误日志 / 截图**: 包含相关的错误消息、堆栈跟踪或截图

- **最小复现代码（MRC）**: 能够演示问题的最小代码示例

**Bug 报告模板示例：**

```markdown
**环境:**
- Go: 1.23.0
- 操作系统: Linux Ubuntu 22.04
- 模块: ADP
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
```

**分支命名规范：**

- `feature/` - 新功能
- `fix/` - Bug 修复
- `docs/` - 文档更改
- `refactor/` - 代码重构
- `test/` - 添加或更新测试

### 3. 进行更改

- 编写清晰、可维护的代码
- 遵循项目的代码结构和架构模式
- 添加适当的注释和文档
- 添加标准文件头（参见下方 [源代码文件头规范](#-源代码文件头规范)）

### 4. 编写测试

- 为新功能添加单元测试
- 确保现有测试仍然通过
- 争取良好的测试覆盖率

```bash
# 运行测试
go test ./...

# 运行测试并查看覆盖率
go test -cover ./...
```

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

遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

- `feat:` - 新功能
- `fix:` - Bug 修复
- `docs:` - 仅文档更改
- `style:` - 代码样式更改（格式化等）
- `refactor:` - 代码重构
- `test:` - 添加或更新测试
- `chore:` - 维护任务

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
> - 如果你想跟踪上游仓库，可以添加：`git remote add upstream https://github.com/kweaver-ai/kweaver.git`

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

- Go 1.23.0 或更高版本
- MariaDB 11.4+ 或 DM8
- OpenSearch 2.x（可选，用于完整功能）
- Git

### 本地开发

1. **克隆你的 Fork：**

```bash
git clone https://github.com/YOUR_USERNAME/kweaver.git
cd kweaver
```

1. **添加上游远程仓库：**

```bash
git remote add upstream https://github.com/kweaver-ai/kweaver.git
```

1. **设置开发环境：**

```bash
# 导航到你要工作的模块
cd <module-directory>/server

# 下载依赖
go mod download

# 运行服务
go run main.go
```

1. **运行测试：**

```bash
go test ./...
```

---

## 🐛 报告安全问题

**请不要通过公共 GitHub Issues 报告安全漏洞。**

相反，请通过以下方式报告：

- 邮箱: [安全联系邮箱]
- 内部安全报告系统

我们将确认收到并与你合作解决问题。

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
