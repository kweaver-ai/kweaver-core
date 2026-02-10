# KWeaver

中文 | [English](README.md)

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE.txt)

KWeaver 是一个构建、发布、运行决策智能型 AI 应用的开源生态。此生态采用本体作为业务知识网络的核心方法，以 DIP 为核心平台，旨在提供弹性、敏捷、可靠的企业级决策智能，进一步释放每一员的生产力。

DIP 平台包括 ADP、Decision Agent、DIP Studio、AI Store 等关键子系统。

## 📚 快速链接

- 🤝 [贡献指南](rules/CONTRIBUTING.zh.md) - 项目贡献指南
- 🚀 [发布规范](rules/RELEASE.zh.md) - 版本管理与发布流程
- 📄 [许可证](LICENSE.txt) - Apache License 2.0
- 🐛 [报告 Bug](https://github.com/kweaver-ai/kweaver/issues) - 报告问题或 Bug
- 💡 [功能建议](https://github.com/kweaver-ai/kweaver/issues) - 提出新功能建议

## 平台架构

```text
┌─────────────────────────────────────────────┐
│               DIP 平台                       │
│  ┌───────────────────────────────────────┐  │
│  │             AI Store                  │  │
│  ├───────────────────────────────────────┤  │
│  │            DIP Studio                 │  │
│  ├───────────────────────────────────────┤  │
│  │          Decision Agent               │  │
│  ├───────────────────────────────────────┤  │
│  │               ADP                     │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

### 核心子系统

| 子项目 | 描述 | 仓库地址 |
| --- | --- | --- |
| **AI Store** | AI 应用与组件市场 | [kweaver-ai/ai-store](https://github.com/kweaver-ai/ai-store) |
| **Studio** | DIP Studio - 可视化开发与管理界面 | [kweaver-ai/studio](https://github.com/kweaver-ai/studio) |
| **Decision Agent** | 决策智能体 | [kweaver-ai/decision-agent](https://github.com/kweaver-ai/decision-agent) |
| **ADP** | 智能数据平台 - 核心开发框架，包含本体引擎、ContextLoader 和 VEGA 数据虚拟化引擎 | [kweaver-ai/adp](https://github.com/kweaver-ai/adp) |
| **Operator Hub** | 算子平台，负责算子管理与编排 | [kweaver-ai/operator-hub](https://github.com/kweaver-ai/operator-hub) |
| **Sandbox** | 沙箱运行环境 | [kweaver-ai/sandbox](https://github.com/kweaver-ai/sandbox) |

## 贡献指南

我们欢迎贡献！请查看我们的[贡献指南](rules/CONTRIBUTING.zh.md)了解如何为项目做出贡献。

快速开始：

1. Fork 代码库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

本项目采用 Apache License 2.0 许可证。详情请参阅 [LICENSE](LICENSE.txt) 文件。

## 支持与联系

- **贡献指南**: [贡献指南](rules/CONTRIBUTING.zh.md)
- **问题反馈**: [GitHub Issues](https://github.com/kweaver-ai/kweaver/issues)
- **许可证**: [Apache License 2.0](LICENSE.txt)

---

后续更多组件开源，敬请期待！
