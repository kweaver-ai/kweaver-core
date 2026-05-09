# 开发规范

中文 | [English](README.md)

本目录包含 KWeaver 项目的开发规范。如果你是新加入的成员，请按以下顺序阅读。

## 阅读顺序

| 序号 | 文档 | 你能学到 |
|------|------|----------|
| 1 | [贡献指南](CONTRIBUTING.zh.md) | 如何 Fork、提 PR、分支命名、代码风格 |
| 2 | [团队协作流程](WORKFLOW.zh.md) | Issue → 设计文档 → 分支 → PR → 合并流程 |
| 3 | [架构规范](ARCHITECTURE.zh.md) | 系统分层、模块边界、依赖规则 |
| 4 | [研发规范](DEVELOPMENT.zh.md) | API 设计、HTTP 语义、错误处理、认证 |
| 5 | [测试规范](TESTING.zh.md) | 测试分层、Makefile 约定、Agent-First 测试 |
| 6 | [发布规范](RELEASE.zh.md) | 版本号、分支策略、发布流程 |

## 快速参考

- **分支命名**：`feature/`、`fix/`、`refactor/`、`docs/`、`ci/`、`chore/` — 详见[贡献指南](CONTRIBUTING.zh.md)
- **提交格式**：Conventional Commits（`type(scope): subject`）— 详见[贡献指南](CONTRIBUTING.zh.md)
- **文档划分**：仓库根 [`docs/`](../docs)、各模块自有 `docs/`、[`help/`](../help) 下 **`manual/`**（手册）与 **`cookbook/`**（场景化 Cookbook；模版见各语言下 `cookbook/` 目录）— 详见 [贡献指南](CONTRIBUTING.zh.md) 中的 **DOCUMENTATION（文档放置规范）** 一节
- **设计文档**：位于各模块的 `docs/design/` 目录 — 详见[团队协作流程](WORKFLOW.zh.md)
- **API 检查清单**：错误格式、分页、状态码 — 详见[研发规范](DEVELOPMENT.zh.md)
