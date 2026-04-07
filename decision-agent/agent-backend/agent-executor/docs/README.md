# Agent-Executor 文档目录

本目录包含 Agent-Executor 项目的所有技术文档，按模块分类组织。

## 📁 目录结构

```
docs/
├── README.md                    # 本文件
├── .gitignore                   # Git 忽略配置
├── makefile                     # 文档构建工具
├── changelogs/                  # 变更日志
│
├── architecture/                # 架构设计文档
│   ├── redocly-guide.md        # API 文档生成指南
│   ├── TTFT_FEATURE.md         # TTFT 功能文档
│   └── error-handling/         # 错误处理系统 🆕
│       ├── README.md           # 错误和异常系统完整文档
│       ├── MIGRATION_GUIDE.md  # 迁移指南
│       └── QUICK_REFERENCE.md  # 快速参考
│
├── configuration/               # 配置管理文档
│   └── config_v2_implementation.md  # 配置管理 V2 实现
│
├── logging/                     # 日志相关文档
│   └── struct_logger_usage.md  # 结构化日志使用指南
│
├── middleware/                  # 中间件相关文档
│   └── starlette_cancel_scope_fix.md  # Starlette Cancel Scope 修复
│
├── session/                     # 会话管理文档
│   ├── IMPLEMENTATION_SUMMARY.md           # Session 缓存实现总结
│   ├── redis_cache_implementation_summary.md  # Redis 缓存实现
│   └── conversation_session_analysis.md    # 会话分析文档
│
├── refactoring/                 # 重构相关文档
│   └── skill_vo_refactoring.md # Skills 字段 VO 重构
│
└── troubleshooting/             # 故障排查文档
    └── keyboard_interrupt_issue.md  # KeyboardInterrupt 问题说明
```

## 📚 文档分类说明

### 🏗️ Architecture（架构设计）
包含系统架构、设计模式、API 设计等相关文档。

**文档列表：**
- `redocly-guide.md` - API 文档生成和管理指南
- `TTFT_FEATURE.md` - TTFT（Time To First Token）功能文档
- **error-handling/** - 🆕 错误和异常系统文档
  - `README.md` ⭐ - 错误和异常系统完整文档
  - `MIGRATION_GUIDE.md` - 从旧系统迁移指南
  - `QUICK_REFERENCE.md` 🔥 - 快速参考手册

### ⚙️ Configuration（配置管理）
包含配置系统的设计、实现和使用文档。

**文档列表：**
- `config_v2_implementation.md` - 配置管理 V2 版本实现文档

### 📝 Logging（日志系统）
包含日志系统的设计、使用和最佳实践。

**文档列表：**
- `struct_logger_usage.md` - 结构化日志完整使用指南

### 🔄 Middleware（中间件）
包含 HTTP 中间件的实现、问题修复和最佳实践。

**文档列表：**
- `starlette_cancel_scope_fix.md` - Starlette Cancel Scope 异常修复文档

### 💾 Session（会话管理）
包含会话缓存、Redis 缓存等相关实现文档。

**文档列表：**
- `IMPLEMENTATION_SUMMARY.md` - Conversation Session 缓存机制实现总结
- `redis_cache_implementation_summary.md` - Redis 缓存实现总结
- `conversation_session_analysis.md` - 会话系统分析文档

### 🔧 Refactoring（重构记录）
包含代码重构的设计、实现和影响分析。

**文档列表：**
- `skill_vo_refactoring.md` - Skills 字段 VO 重构总结

### 🐛 Troubleshooting（故障排查）
包含常见问题、故障排查和解决方案。

**文档列表：**
- `keyboard_interrupt_issue.md` - KeyboardInterrupt 警告问题说明

## 📖 如何使用本文档

### 快速查找

1. **按功能模块查找**：根据你关注的功能模块（如日志、配置、会话等）进入对应目录
2. **按问题类型查找**：遇到问题时，先查看 `troubleshooting/` 目录
3. **按时间顺序查找**：查看 `changelogs/` 目录了解最新变更

### 文档规范

所有文档应遵循以下规范：

1. **Markdown 格式**：使用标准 Markdown 语法
2. **清晰的标题层级**：使用 `#`、`##`、`###` 等标题层级
3. **代码示例**：提供清晰的代码示例和注释
4. **更新日期**：在文档末尾注明最后更新日期
5. **相关链接**：提供相关文档、Issue、PR 的链接

### 贡献文档

添加新文档时，请：

1. 选择合适的分类目录
2. 使用描述性的文件名（小写字母，下划线分隔）
3. 在本 README 中更新文档列表
4. 确保文档包含必要的章节（概述、实现、示例、测试等）

## 🔗 相关资源

### 项目仓库
- **主仓库**：[AISHUDevOps/DIP/_git/agent-executor](https://dev.azure.com/AISHUDevOps/DIP/_git/agent-executor)

### 外部文档
- [FastAPI 官方文档](https://fastapi.tiangolo.com/)
- [Starlette 官方文档](https://www.starlette.io/)
- [Structlog 官方文档](https://www.structlog.org/)
- [Redis 官方文档](https://redis.io/documentation)

### 测试文件
- **单元测试**：`/test/` 目录
- **中间件测试**：`/test/router_test/test_middleware_compatibility.py`

## 📅 文档维护

### 最后更新
- **日期**：2025-11-02
- **维护者**：Agent-Executor 开发团队

### 更新记录
- 2025-11-02：🆕 添加错误和异常系统文档（error-handling/）
- 2025-10-30：添加 TTFT 功能文档
- 2025-10-28：创建文档目录结构，按模块分类整理文档
- 2025-10-28：添加 Starlette Cancel Scope 修复文档
- 2025-10-27：添加 KeyboardInterrupt 问题说明文档
- 2025-10-20：添加 Session 缓存实现文档

## 💡 提示

- 📌 **重要文档**标记为 ⭐
- 🆕 **最新文档**标记为 🆕
- 🔥 **常用文档**标记为 🔥

如有文档相关问题或建议，请联系开发团队。
