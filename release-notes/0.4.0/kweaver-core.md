# KWeaver Core 0.4.0 Release Notes

KWeaver Core 0.4.0 is a significant milestone release representing a deep optimization of the KWeaver platform architecture. This release targets "engineering architecture upgrade, significant performance improvement, and comprehensive quality hardening" as its core objectives, achieving major breakthroughs in BKN engine capability expansion, VEGA data virtualization self-developed engine upgrade, and end-to-end observability establishment — while completing an overall simplification of the platform engineering architecture: memory footprint reduced to 48GB, microservice count reduced by 12, with significant improvements in accuracy and latency metrics for both structured and unstructured scenarios.

---

## Highlights

**1. BKN Engine Fully Expanded — Knowledge Network Construction Capability Elevated**

Added Markdown format document import and management support, completed a Dataset-based concept index system reconstruction, corrected Action's object creation query logic, and added the ability to build indexes for specified object classes. The BKN engine has been officially migrated from ontology-manager to bkn-backend, completing service architecture standardization with comprehensive improvements in knowledge network construction efficiency and accuracy.

**2. VEGA Data Virtualization Self-Developed Engine Online — Multi-Source Metadata Collection Expanded**

The new VEGA engine adopts a self-developed data query framework, achieving database table scanning and native query capabilities for MariaDB — eliminating third-party dependencies. Metadata collection for MariaDB, Oracle, MySQL, and OpenSearch has been expanded simultaneously, with support for cross-datasource JOIN queries, significantly improving the flexibility and coverage of the data virtualization layer.

**3. TraceAI End-to-End Observability System Established**

Completed TracingCollector data collection specifications, covering complete collection and TracingID association for Agent conversations, Session, Run levels, and tool call chains; simultaneously advancing BKN, Execution Factory, and Dataflow tracing instrumentation to establish a unified cross-component observability data standard.

**4. Architecture Engineering Continuously Streamlined — Platform Evolving Toward Lightweight**

Fully removed MongoDB dependency, refactored the eacp-single service, and introduced session-level Python dependency management in the sandbox runtime. Memory footprint reduced to 48GB, microservice count reduced by 12, with the architecture continuously evolving toward a lightweight design.

---

## 1. BKN Engine

BKN (Business Knowledge Network) is a Markdown-based declarative modeling language for defining objects, relationships, and actions in a business knowledge network. Version 0.4.0 completes the BKN engine service renaming (ontology-manager → bkn-backend), and achieves important extensions in document import, index construction, and action execution capabilities.

**1. Markdown Format Document Import and Management Support**

This version adds import and management capabilities for Markdown (`.md`) format documents, fully implementing the document management entry point as required by the product layer. Users can upload, parse, and bind Markdown documents to the knowledge network through a unified interface, incorporating unstructured business documents into the BKN knowledge source system and expanding the document coverage of the knowledge network.

**2. Dataset-Based Concept Index System Reconstruction**

Completed a comprehensive reconstruction of the concept index construction logic, tightly binding the index generation process to Dataset data collections, while adding support for users to specify particular object classes for index construction — improving the granularity and efficiency of index building.

**3. Action Query Logic Correction**

Corrected a query logic issue in the Action's object creation process, enhancing action execution capability in scenarios without supporting materials, ensuring accuracy and completeness of action query results, and providing a more reliable data foundation for automated business process execution.

**4. bkn-backend Service Standardization Migration**

Completed the official renaming and migration of the ontology-manager service to bkn-backend, maintaining complete backward compatibility for existing API interfaces, updating error type naming conventions, and completing the reconstruction of related documentation and error packages — laying an architectural foundation for subsequent BKN engine capability expansion.

---

## 2. VEGA Data Virtualization

The VEGA data virtualization layer is responsible for unified access, querying, and virtualized management of multi-source data. Version 0.4.0 completes the launch of the self-developed query engine and significantly expands data source metadata collection coverage.

**1. New VEGA Engine: Self-Developed Data Querying and MariaDB Table Scanning**

The new VEGA engine is online, adopting a self-developed data query framework to replace the original third-party dependencies, achieving database table structure scanning and native data querying capabilities for MariaDB databases. The new engine shows significant improvements in query performance and flexibility, laying the foundation for subsequent multi-database adaptation and expansion.

**2. Multi-Source Metadata Collection Expansion**

Added metadata collection capabilities for four data source types: MariaDB, Oracle, MySQL, and OpenSearch. Users can obtain table structure, field information, and index metadata for these data sources through a unified metadata management interface, improving the platform's coverage of enterprise diverse data infrastructure.

**3. Cross-Datasource JOIN Query Capability**

Corrected the custom data query framework, officially opening cross-datasource JOIN query capability, supporting associated queries of data from multiple data sources at the VEGA virtualization layer, greatly expanding the flexibility of data analysis scenarios.

---

## 3. TraceAI

TraceAI is KWeaver platform's end-to-end observability component, responsible for data collection, chain tracing, and behavioral auditing of Agent execution processes. Version 0.4.0 completes core collection specification establishment and cross-component instrumentation advancement.

**1. TracingCollector Collection Specification Establishment**

Completed a comprehensive data collection specification design for TracingCollector, defining clear collection standards for Agent-related chains:

- **Conversation-level collection**: Complete recording of Agent conversation content, associating Session and Run level context data
- **Execution step collection**: Recording each Step executed by the Agent, including tool call Steps and their response content
- **Chain association**: Cross-component association calls for execution chain data based on TracingID

**2. Cross-Component Tracing Instrumentation Advancement**

Simultaneously advancing the BKN engine, Execution Factory, and Dataflow to improve their tracing chain data instrumentation according to TracingCollector specifications, establishing a unified observability data standard and providing a complete data foundation for subsequent Agent behavior analysis and troubleshooting.

---

## 4. Dataflow

Dataflow is KWeaver platform's data stream processing engine, responsible for document parsing, data flow transfer, and multi-node orchestration. Version 0.4.0 completes two important architectural improvements.

**1. Platform-Wide MongoDB Dependency Removal**

Completed the removal of MongoDB dependencies from the Dataflow layer and across the entire system — no MongoDB dependency components remain in the platform architecture. This transformation significantly simplifies system deployment and operations complexity, reduces maintenance costs from technology stack diversity, and further streamlines the overall platform architecture.

**2. Document Parsing Node Adapted to MinerU Official API**

The document parsing node has completed adaptation to the MinerU official API, replacing the original local parsing solution. The MinerU official API delivers significant improvements in parsing accuracy for complex documents such as PDFs and mixed text-image layouts, providing higher quality document parsing results for knowledge network construction.

---

## 5. Context Loader

Context Loader is the context loading and management component for KWeaver platform agents, responsible for providing precise and efficient knowledge context retrieval and injection for Agents. Version 0.4.0 completes important optimizations in interface specifications and compression capabilities.

**1. Interface Form Optimization — Improving Agent Accuracy and Efficiency**

Comprehensively optimized the interface call form of Context Loader, streamlining the interface call chain, reducing redundant data transmission, and improving agent response speed and retrieval accuracy during the context retrieval phase.

**2. TOON Compression Capability Added**

Added context compression capability in TOON format, supporting compressed data transmission over both HTTP and MCP protocols. Through context compression, Token consumption for context injection is effectively reduced while maintaining knowledge integrity, improving agent response efficiency in large-scale knowledge network scenarios.

---

## 6. Execution Factory

The Execution Factory is the core of KWeaver platform's function computation and sandbox execution scheduling, responsible for managing interactions between user-defined functions and sandbox runtimes. Version 0.4.0 completes the closed-loop development of function dependency library management capabilities.

**1. Function Dependency Library Installation Support**

Completed full support for function sandbox dependency libraries. Functions defined by users in the Execution Factory can declare and install required Python dependency packages through a unified mechanism. The system automatically handles dependency resolution, installation, and version management, eliminating dependency missing issues during function execution.

**2. Function Editor Supports Dependency Package Configuration**

The operator platform function editor adds a dependency package configuration interface, allowing users to directly configure the dependency package list required by functions within the editor and view installation status in real time, achieving an integrated experience of dependency management and function development.

---

## 7. ISF Information Security Fabric

ISF (Information Security Fabric) is responsible for permission management, identity authentication, and access control systems of the KWeaver platform. Version 0.4.0 completes service architecture reconstruction and authorization management capability expansion.

**1. Refactored eacp-single Service Removal**

Completed the removal of the eacp-single service through architectural refactoring, replacing the original single service model with a Kubernetes-based multi-replica concurrent execution control solution. The new solution has stronger horizontal scaling capability and higher execution concurrency, eliminating single-point bottlenecks and significantly improving platform stability and throughput in high-concurrency scenarios.

**2. Authorization Management Supports Resource Creator Condition Configuration**

Authorization management adds access control configuration capabilities at the resource creator dimension: creators of resources can manage operations on their created resources without requiring additional administrator authorization. This capability further refines the platform's permission granularity, making permission policies more aligned with ownership management requirements in actual business scenarios.

---

## 8. Sandbox Runtime

The Sandbox Runtime provides a secure isolated code execution environment for the KWeaver platform, supporting Python function execution, file workspace management, and multi-runtime scheduling. Version 0.4.0 (corresponding to Sandbox v0.3.0) adds session-level Python dependency management capabilities.

**1. Session-Level Python Dependency Management**

Added complete session-level Python dependency configuration and management capabilities:

- **Dependency declaration and installation**: Supports declaring required dependency packages when creating a session, with background tasks triggering initial dependency synchronization installation
- **Manual installation**: Supports on-demand manual triggering of dependency installation, flexibly responding to runtime environment changes
- **Status tracking**: Session responses include details of installed dependencies and error information, with the frontend session management interface displaying dependency installation operations and status in sync

**2. Runtime Dependency Synchronization and Isolation**

Ensures complete synchronization of dependency states between runtime executors and the control plane, automatically handling dependency conflicts and environment isolation, enhancing sandbox compatibility and stability in complex deployment environments.

**3. Database and Version Upgrade Support**

Added automatic database upgrade capability, automatically completing database structure version alignment on service startup, supporting smooth upgrades from historical versions, reducing operations costs and upgrade risks.

---

## 9. Decision Agent

Decision Agent is the core agent component of the KWeaver platform for business decision scenarios, providing multi-step reasoning, tool invocation, and conversation management capabilities. Version 0.4.0 completes permission handling enhancement and API routing fixes.

**1. agent-executor Log Directory Permission Handling Enhancement**

Enhanced agent-executor operational stability across various deployment environments, improved log directory permission handling and error recovery mechanisms, ensuring complete recording of agent execution logs.

**2. agent router v1 Registration Missing Fix**

Fixed the issue of historical version API endpoints being unavailable, restoring complete API routing functionality, and ensuring normal integration and access for existing business systems.

---

## More Resources

**1. Product Installation Packages and Technical Documentation**

**AI Data Platform (ADP):**
- **GitHub Release**: [https://github.com/kweaver-ai/adp/tree/release/0.4.0](https://github.com/kweaver-ai/adp/tree/release/0.4.0)
- **Technical Documentation**: [https://github.com/kweaver-ai/adp/blob/main/README.md](https://github.com/kweaver-ai/adp/blob/main/README.md)

**Decision Agent:**
- **GitHub Release**: [https://github.com/kweaver-ai/decision-agent/tree/feature/20260312](https://github.com/kweaver-ai/decision-agent/tree/feature/20260312)
- **Changelog**: [https://github.com/kweaver-ai/decision-agent/blob/feature/20260312/changelog.zh.md](https://github.com/kweaver-ai/decision-agent/blob/feature/20260312/changelog.zh.md)

**Sandbox:**
- **GitHub Release**: [https://github.com/kweaver-ai/sandbox/tree/feature/20260312](https://github.com/kweaver-ai/sandbox/tree/feature/20260312)
- **Changelog**: [https://github.com/kweaver-ai/sandbox/blob/feature/20260312/CHANGELOG_ZH.md](https://github.com/kweaver-ai/sandbox/blob/feature/20260312/CHANGELOG_ZH.md)

**2. Product Release Materials**

- **Release Date**: 2026-03-14
- **Version Target Document**: KWeaver Core 0.4.0 Key Planning

---

**KWeaver — Intelligent Data Platform and Decision Engine for AI Agents**
