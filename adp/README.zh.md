# ADP (人工智能数据平台)

[中文](README.zh.md) | [English](README.md)

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE.txt)

**[ADP (智能数据平台)](https://github.com/kweaver-ai/adp)** 是 KWeaver 生态系统的一部分。如果你喜欢这个项目，请同时也为 **[KWeaver](https://github.com/kweaver-ai/kweaver)** 项目点亮星标⭐。

**[KWeaver](https://github.com/kweaver-ai/kweaver)** 是一个用于构建、部署和运行决策智能 AI 应用的开源生态系统。该生态系统采用本体作为业务知识网络的核心方法论，以 DIP 为核心平台，旨在提供弹性、敏捷、可靠的企业级决策智能，进一步释放每个人的生产力。

DIP 平台包含 ADP、Decision Agent、DIP Studio 和 AI Store 等关键子系统。

## 📚 快速链接

- 🤝 [贡献指南](../rules/CONTRIBUTING.zh.md) - 项目贡献指引
- 📄 [许可证](LICENSE.txt) - Apache License 2.0
- 🐛 [报告 Bug](https://github.com/kweaver-ai/adp/issues) - 报告错误或问题
- 💡 [请求功能](https://github.com/kweaver-ai/adp/issues) - 建议新功能

## 平台定义

ADP 是一个智能数据平台，旨在弥合异构数据源与 AI Agent 之间的鸿沟。它通过业务知识网络（本体）抽象数据复杂性，提供统一的数据访问（VEGA），并通过数据流编排（DataFlow）处理和转换数据。

## 核心组件

### 1. 本体引擎 (Ontology Engine)
本体引擎是一个分布式的业务知识网络管理系统，允许企业对业务世界进行数字化建模。
- **多维建模**：定义对象类型、关系类型和动作类型，以映射现实世界的实体。
- **可视化配置**：直观的本体管理界面。
- **智能查询**：支持复杂的多跳关系路径查询和语义搜索。

### 2. 上下文加载器 (ContextLoader)
ContextLoader 负责为 AI Agent 构建高质量的上下文。
- **精准召回**：基于本体概念而非简单的关键词匹配来检索信息。
- **动态组装**：根据当前任务需求和用户权限组装上下文片段。
- **按需加载**：仅加载必要的数据，防止上下文窗口溢出。

### 3. VEGA 数据虚拟化 (VEGA Data Virtualization)
VEGA 为异构数据源提供统一的 SQL 接口，将应用程序与底层数据库实现解耦。
- **单一访问点**：通过统一接口连接 MariaDB、DM8、REST API 等多种数据源。
- **跨源查询**：无缝连接不同数据库中的数据进行查询。
- **标准化语义**：确保所有应用程序的数据定义一致。

### 4. 数据流编排 (DataFlow)
DataFlow 是一个专为数据处理设计的可视化流程编排引擎。
- **数据管道设计**：通过可视化界面构建和编排数据处理流程。
- **代码执行**：支持沙箱化 Python 代码执行，处理复杂的数据转换任务。
- **实时处理**：支持定时调度和事件驱动的数据处理流程。

## 技术目标

- **统一语义**：通过在本体中定义业务逻辑，实现代码与业务逻辑的解耦，允许跨 Agent 全局复用。
- **数据敏捷**：虚拟化数据访问，避免硬编码集成，能够快速适应数据源的变化。
- **过程可观测**：使所有 Agent 的动作、数据流向和决策过程都可追踪和审计。
- **安全执行**：在数据流转的每一步强制执行细粒度的权限控制和验证。

## 架构

```text
┌───────────────────────────────────────────────────────────────────┐
│                           ADP Platform                            │
│                                                                   │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐           │
│  │   DataFlow   │◄──┤ ContextLoader│◄──┤ Ontology Eng.│           │
│  │   (数据编排)  │   │    (组装)    │   │    (建模)    │           │
│  └──────┬───────┘   └──────┬───────┘   └──────┬───────┘           │
│         │                  │                  │                   │
│         ▼                  ▼                  ▼                   │
│  ┌──────────────────────────────────────────────────────┐         │
│  │             VEGA Data Virtualization Engine          │         │
│  └─────────────────────────┬────────────────────────────┘         │
│                            │                                      │
│                            ▼                                      │
│  ┌────────────┐     ┌────────────┐     ┌────────────┐             │
│  │  MariaDB   │     │    DM8     │     │ ExternalAPI│             │
│  └────────────┘     └────────────┘     └────────────┘             │
└───────────────────────────────────────────────────────────────────┘
```

## 快速开始

### 前置要求

- **Go**: 1.23+ (用于 Ontology)
- **Java**: JDK 1.8+ (用于 VEGA)
- **Node.js**: 18+ (用于 Web Console)
- **数据库**: MariaDB 11.4+ 或 DM8
- **搜索引擎**: OpenSearch 2.x

### 构建与运行设置

1.  **克隆仓库**
    ```bash
    git clone https://github.com/kweaver-ai/adp.git
    cd adp
    ```

2.  **初始化数据库**
    运行 `sql/` 目录下的 SQL 初始化脚本以设置数据库 Schema。

3.  **构建模块**

    *   **BKN 引擎 (Go)**:
        详细说明请参考 [bkn/README.zh.md](bkn/README.zh.md)。
        ```bash
        cd bkn/bkn-backend
        go run main.go
        ```

    *   **VEGA (Java)**:
        ```bash
        cd vega
        mvn clean install
        ```

    *   **DataFlow**:
        详细说明请参考 [dataflow/README.zh.md](dataflow/README.zh.md)。

    *   **Web 控制台 (Node.js)**:
        ```bash
        cd web
        npm install
        npm run dev
        ```

## 贡献

我们欢迎贡献！有关如何为本项目做出贡献的详细信息，请参阅我们的 [贡献指南](../rules/CONTRIBUTING.zh.md)。

## 许可证

本项目采用 Apache License 2.0 许可证。详情请参阅 [LICENSE](LICENSE.txt) 文件。

## 支持与联系

- **Issues**: [GitHub Issues](https://github.com/kweaver-ai/adp/issues)
- **贡献**: [贡献指南](../rules/CONTRIBUTING.zh.md)