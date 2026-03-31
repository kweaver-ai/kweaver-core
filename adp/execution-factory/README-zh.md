# Execution Factory

中文 | [English](README.md)

Execution Factory 是 KWeaver 生态的一部分。如果您喜欢这个项目，欢迎也给 KWeaver 项目点个 ⭐！

KWeaver 是一个构建、发布、运行决策智能型 AI 应用的开源生态。此生态采用本体作为业务知识网络的核心方法，以 DIP 为核心平台，旨在提供弹性、敏捷、可靠的企业级决策智能，进一步释放每一员的生产力。

DIP 平台包括 ADP、Decision Agent、DIP Studio、AI Store 等关键子系统。

## 📚 快速链接

- 🤝 [贡献指南](CONTRIBUTING.zh.md) - 项目贡献指南
- 📄 [许可证](LICENSE.txt) - Apache License 2.0
- 🐛 [报告 Bug](https://github.com/kweaver-ai/operator-hub/issues) - 报告问题或 Bug
- 💡 [功能建议](https://github.com/kweaver-ai/operator-hub/issues) - 提出新功能建议

## Execution Factory 定义

Execution Factory 是一个开源的算子与工具管理平台，旨在连接大语言模型（LLMs）与实际业务能力。通过支持 [Model Context Protocol (MCP)](https://modelcontextprotocol.io/)，它提供了一套标准化的机制来注册、管理和执行各种算子（Operators）及工具（Tools），帮助开发者快速构建强大的 AI Agent 应用。

## 核心组件

本项目包含两个主要组件：

### 1. Operator Integration (`operator-integration`)
核心集成服务平台，负责算子和工具的全生命周期管理。
- **算子管理**：支持算子的注册、版本控制、发布与下架。
- **工具箱（Toolbox）**：支持将多个工具组合成工具箱，便于统一管理和调用。
- **MCP 支持**：作为 MCP Server，向 LLM 提供标准化的工具调用接口。
- **多协议适配**：支持 HTTP、SSE 等多种通信协议。
- **权限控制**：内置基于策略的访问控制机制。

### 2. Operator App (`operator-app`)
应用端运行时与示例实现。
- 提供了一个轻量级的算子执行环境。
- 展示了如何集成和使用 Execution Factory 的核心能力。
- 包含 MCP 客户端与服务端的交互示例。

## 特性

- **标准化接口**：基于 MCP 协议，实现模型与工具的解耦。
- **灵活扩展**：支持多种编程语言编写的算子（如 Go, Python 等）。
- **可观测性**：集成了 OpenTelemetry，提供全链路追踪能力。
- **高性能**：基于 Go 语言开发，具备高并发处理能力。

## 快速开始

### 前置要求
- Go 1.24+
- MySQL / MariaDB / Dameng DB
- Redis

### 编译与运行

#### 运行 Operator Integration
```bash
cd operator-integration
# 安装依赖
go mod tidy
# 编译
go build -o operator-integration server/main.go
# 运行 (需配置相应的配置文件)
./operator-integration
```

#### 运行 Operator App
```bash
cd operator-app
# 安装依赖
go mod tidy
# 编译
go build -o operator-app server/main.go
# 运行
./operator-app
```

## 贡献

欢迎提交 Pull Request 或 Issue！

## 许可证

[Apache-2.0](LICENSE)
