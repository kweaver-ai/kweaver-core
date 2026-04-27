# ADP (AI Data Platform)

[中文](README.zh.md) | English

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE.txt)

**[ADP (AI Data Platform)](https://github.com/kweaver-ai/adp)** is part of the KWeaver ecosystem. If you like it, please also star⭐ the **[KWeaver](https://github.com/kweaver-ai/kweaver-core)** project as well.

**[KWeaver](https://github.com/kweaver-ai/kweaver-core)** is an open-source ecosystem for building, deploying, and running decision intelligence AI applications. This ecosystem adopts ontology as the core methodology for business knowledge networks, with DIP as the core platform, aiming to provide elastic, agile, and reliable enterprise-grade decision intelligence to further unleash everyone's productivity.

The DIP platform includes key subsystems such as ADP, Decision Agent, DIP Studio, and AI Store.

## 📚 Quick Links

- 🤝 [Contributing](../rules/CONTRIBUTING.md) - Guidelines for contributing to the project
- 📄 [License](LICENSE.txt) - Apache License 2.0
- 🐛 [Report Bug](https://github.com/kweaver-ai/adp/issues) - Report a bug or issue
- 💡 [Request Feature](https://github.com/kweaver-ai/adp/issues) - Suggest a new feature

## Platform Definition

ADP is an intelligent data platform that bridges the gap between heterogeneous data sources and AI agents. It abstracts data complexity through business knowledge networks (Ontology), provides unified data access (VEGA), and orchestrates data processing through DataFlow pipelines.

## Key Components

### 1. Ontology Engine
The Ontology Engine is a distributed business knowledge network management system. It allows enterprises to model their business world digitally.
- **Multi-dimensional Modeling**: Define object types, relation types, and action types to map real-world entities.
- **Visual Configuration**: Intuitive interface for ontology management.
- **Intelligent Query**: Supports complex multi-hop relationship path queries and semantic search.

### 2. ContextLoader
ContextLoader is responsible for constructing high-quality context for AI agents.
- **Precise Recall**: Retrieves information based on ontology concepts rather than just keyword matching.
- **Dynamic Assembly**: Assembles context fragments based on the current task needs and user permissions.
- **On-Demand Loading**: Loads only the necessary data to prevent context window overflow.

### 3. VEGA Data Virtualization
VEGA provides a unified SQL interface for heterogeneous data sources, decoupling applications from underlying database implementations.
- **Single Access Point**: Connect to MariaDB, DM8, REST APIs, and more through a single interface.
- **Cross-Source Query**: Join data across different databases seamlessly.
- **Standardized Semantics**: Ensures consistent data definitions across all applications.

### 4. DataFlow
DataFlow is a visual pipeline orchestration engine designed for data processing.
- **Data Pipeline Design**: Build and orchestrate data processing flows through a visual interface.
- **Code Execution**: Supports sandboxed Python code execution for complex data transformation tasks.
- **Real-time Processing**: Supports scheduled and event-driven data processing pipelines.

## Technical Goals

- **Unified Semantics**: Decouple business logic from code by defining it in the Ontology, allowing for global reuse across agents.
- **Data Agility**: Virtualize data access to avoid hard-coded integrations and enable rapid adaptation to data source changes.
- **Observable Process**: Make all agent actions, data flows, and decisions traceable and auditable.
- **Secure Execution**: Enforce granular permission controls and validation at every step of the data flow.

## Architecture

```text
┌───────────────────────────────────────────────────────────────────┐
│                           ADP Platform                            │
│                                                                   │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐           │
│  │   DataFlow   │◄──┤ ContextLoader│◄──┤ Ontology Eng.│           │
│  │(Data Pipeline)│   │  (Assembly)  │   │  (Modeling)  │           │
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

## Quick Start

### Prerequisites

- **Go**: 1.23+ (for Ontology)
- **Java**: JDK 1.8+ (for VEGA)
- **Node.js**: 18+ (for Web Console)
- **Database**: MariaDB 11.4+ or DM8
- **Search Engine**: OpenSearch 2.x

### Build & Run Setup

1.  **Clone the Repository**
    ```bash
    git clone https://github.com/kweaver-ai/adp.git
    cd adp
    ```

2.  **Initialize Database**
    Run the SQL initialization scripts located in the `sql/` directory to set up your database schema.

3.  **Build Modules**

    *   **BKN (Go)**:
        Refer to [bkn/README.md](bkn/README.md) for detailed instructions.
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
        Refer to [dataflow/README.md](dataflow/README.md) for detailed instructions.

    *   **Web Console (Node.js)**:
        ```bash
        cd web
        npm install
        npm run dev
        ```

## Contributing

We welcome contributions! Please see our [Contributing Guide](../rules/CONTRIBUTING.md) for details on how to contribute to this project.

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE.txt) file for details.

## Support & Contact

- **Issues**: [GitHub Issues](https://github.com/kweaver-ai/adp/issues)
- **Contributing**: [Contributing Guide](../rules/CONTRIBUTING.md)