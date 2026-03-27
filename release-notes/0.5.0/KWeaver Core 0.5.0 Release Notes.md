# KWeaver Core 0.5.0 Release Notes

---

## Overview

KWeaver Core 0.5.0 is a major release that significantly accelerates platform capabilities. This release marks the **official launch of KWeaver SDK**, opening the full platform capabilities to developers for the first time. **TraceAI** establishes a complete end-to-end TraceID mechanism with agent execution trace collection and query now generally available. **Context Loader** achieves ≥30% token compression and unified action recall management via BKN. Both **VEGA** and **Dataflow** reach important milestones in data indexing and data lake decoupling. Additionally, BKN, Execution Factory, Context Loader, and Dataflow all complete architectural decoupling from business domains and ISF, further improving platform modularity and deployment flexibility.

---

## Highlights

**1. KWeaver SDK — Official Release, Full Platform Access**

KWeaver SDK is officially released with dual-language support in TypeScript and Python, covering CLI tools, SDK libraries, and companion Skills. Developers can access Business Knowledge Networks, Agent conversations, data management, Context Loader, and other core platform capabilities through a unified client. Companion Skills enable AI agents to call platform APIs directly with automatic parameter resolution — no manual input required.

**2. TraceAI — End-to-End Observability Now Available**

TraceAI v0.1.1 is officially released, built on the OpenTelemetry standard. A complete end-to-end TraceID mechanism is established for unified collection and management of agent execution traces. A unified trace resource query API is provided, supporting retrieval by TraceID along with accessed BKN resources, enabling AI to automatically analyze execution evidence chains.

**3. Context Loader — ≥30% Token Compression & Unified Action Management**

Context Loader return content structure is fully optimized to reduce redundant information, achieving an average token compression rate of ≥30% and agent execution accuracy of ≥90%. The action recall mechanism is upgraded architecturally — now routing through BKN action drivers instead of direct Execution Factory/MCP connections, laying the foundation for unified risk control across the full action lifecycle.

**4. VEGA — Complete Data Build Capability**

VEGA adds Resource data build capabilities with support for full and incremental index construction and custom views. Catalog is extended with PostgreSQL data source support and scan task management, completing the end-to-end data build capability loop in the virtualization layer.

**5. Dataflow — Content Data Lake Decoupling**

Dataflow decouples from the content data lake by introducing the DFS file protocol. External files can now directly trigger data processing pipelines, significantly reducing deployment dependencies and operational overhead.

---

## Detailed Changes

### 【KWeaver SDK】

KWeaver SDK is the official developer integration toolkit for the KWeaver platform. This is its first official release, providing dual-language implementations in TypeScript and Python — covering CLI tools, SDK libraries, and companion Skills — to give developers and AI agents full access to core platform capabilities.

**1. Dual-Language SDK & CLI**

Both TypeScript and Python implementations are released simultaneously:

- **TypeScript SDK**: `npm install @kweaver-ai/kweaver-sdk` (requires Node.js 22+), provides `KWeaverClient` with streaming Agent conversation and interactive TUI support; CLI: `npm install -g @kweaver-ai/kweaver-sdk`
- **Python SDK**: `pip install kweaver-sdk` (requires Python 3.10+), provides `from kweaver import KWeaverClient`; CLI: `pip install kweaver-sdk[cli]`
- **Unified Authentication**: Supports both OAuth2 browser login (`kweaver auth login`) and client credentials, covering interactive and headless use cases

**2. Core Platform API Coverage**

The SDK wraps the main API capabilities of the KWeaver platform, accessible through a unified client:

- **Business Knowledge Networks**: BKN list query, instance search, subgraph query, action invocation
- **Agent Conversation**: Single Q&A and streaming conversation, with Session management
- **Data Management**: Data source management, data views, Dataflow pipeline triggering, CSV import
- **Observability**: VEGA Catalog query, TraceAI trace data retrieval
- **Context Loader**: Semantic search and context assembly

**3. Companion Skills**

Companion Skills are released alongside the SDK, enabling AI agents to invoke KWeaver platform capabilities directly — with automatic technical parameter resolution and no manual user input required.

---

### 【BKN Engine】

Version 0.5.0 adds a "Risk" domain extension to the BKN semantic modeling framework and completes architectural decoupling between BKN and business domains, enhancing both semantic expressiveness and deployment flexibility.

**1. Risk Semantic Modeling**

"Risk" semantic modeling is now supported in BKN Business Knowledge Networks. Users can define and manage risk-related business semantic structures within a knowledge network. This capability provides standardized knowledge support for agents in scenarios such as risk identification, compliance review, and anomaly detection, expanding BKN's business semantic coverage.

**2. Relaxed Relation Type Naming Constraints**

Within the same Business Knowledge Network, relation type names are no longer required to be unique — duplicate relation type names are now permitted. This change accommodates real-world business modeling requirements and removes constraints caused by naming conflicts.

**3. BKN Decoupled from Business Domains and ISF**

The architectural decoupling of BKN from business domains and ISF is complete. BKN core services no longer have a hard dependency on business domain configuration, supporting more flexible independent deployment and cross-domain reuse. Also fixed: an incorrect permission check during data view preview when authentication is disabled.

---

### 【TraceAI】

Version 0.5.0 marks the official release of TraceAI v0.1.1, building a complete observability infrastructure on the OpenTelemetry standard and bringing agent observability to full coverage.

**1. End-to-End TraceID Mechanism**

A complete end-to-end TraceID mechanism is established for unified collection and management of agent execution traces. Each Agent execution generates a unique TraceID that links all invocation steps, tool usages, and BKN resource access records throughout the execution chain, providing a data foundation for execution traceability and analysis.

**2. OpenTelemetry-Based Observability Infrastructure**

TraceAI uses the OpenTelemetry standard for collection and storage:

- **Collection pipeline**: LLM applications/Agents report trace data via OTLP → OpenTelemetry Collector (deployed via Kubernetes Helm Chart) → OpenSearch persistent storage
- **Query service**: The `agent-observability` service provides a REST API supporting raw DSL queries and conversation-level trace retrieval, with a default return limit of 1,000 records
- **Agent-Factory integration**: Decision Agent conversation messages are linked to Agent execution chains, enabling full execution record traceability from the conversation dimension
- **Deployment support**: Docker images, Helm Charts, and GitHub Actions workflows are provided, covering the full containerized deployment lifecycle

**3. Unified Trace Resource Query API**

A unified trace resource query API is provided, supporting:

- **Full execution trace by TraceID**: Returns all step records from an Agent execution, including tool calls, intermediate reasoning, and final results
- **Accessed BKN resource query**: Retrieves the Business Knowledge Network resources accessed during the execution
- **Execution evidence chain analysis**: Trace data is structured to meet AI auto-analysis requirements, enabling downstream agents to read and generate interpretable execution evidence chain reports

---

### 【VEGA Data Virtualization】

Version 0.5.0 delivers significant capability expansions in both Catalog metadata management and Resource data build.

**1. Catalog Support for PostgreSQL and Scan Task Management**

VEGA Catalog now supports metadata collection and management for PostgreSQL data sources. Users can scan and register PostgreSQL schema structures and field information through the unified Catalog management interface. Scan task management is also added, supporting creation, viewing, and status tracking of metadata scan tasks to improve governance operability and visibility.

**2. Resource Data Build (Index Construction)**

New Resource data build capabilities support vector index construction for data sources within the VEGA virtualization layer:

- **Full build**: Executes a complete index construction over the entire data source
- **Incremental build**: Based on a configured incremental field, syncs only data added or changed since the last build, reducing build cost

MySQL data sources are supported in this release; PostgreSQL support is planned.

**3. Custom Views for Resources**

Custom view definition is now supported for Resources, allowing users to flexibly configure field mappings and vector index structures to meet the retrieval requirements of different business scenarios.

**4. VEGA Decoupled from ISF**

The architectural decoupling of VEGA from ISF is complete. VEGA data processing no longer depends on ISF services, reducing cross-service coupling and improving VEGA's suitability for standalone deployment scenarios.

---

### 【Dataflow】

Version 0.5.0 completes the full decoupling from the content data lake and introduces a unified file protocol to open up external data source ingestion.

**1. Content Data Lake Decoupling**

Dataflow is fully decoupled from the content data lake. By introducing the DFS (Distributed File Service) file protocol, the data transfer intermediate layer within Dataflow is now unified on DFS file references. Data passing between nodes is no longer bound to a specific storage service, significantly simplifying Dataflow's deployment dependencies and architectural complexity.

**2. DFS File Protocol and External Data Source Ingestion**

A unified DFS file protocol is established across all Dataflow processing nodes:

- **Local file trigger**: Supports uploading local files directly via API to trigger Dataflow pipelines; files automatically enter the unstructured processing chain (parsing, chunking, vectorization, etc.)
- **External URL ingestion**: Supports passing external data source URLs; any external data source can be connected by implementing the corresponding connector
- **Full node compatibility**: All Dataflow processing nodes have been updated to support the DFS file protocol; intermediate data conversion is now unified on DFS file information flow
- **Underlying implementation**: The file subsystem is implemented based on OssGateway; intermediate temporary data storage uses OSS Gateway and no longer depends on the built-in content data lake storage service

**3. SQL Write Node Enhancement**

A new "Truncate and Append" write mode is added to the Dataflow SQL write node, while the existing "Append" and "Overwrite" modes remain unchanged. API parameter validation logic is also optimized to improve write operation reliability and usability.

**4. Dataflow Decoupled from Business Domains and ISF**

The architectural decoupling of Dataflow from business domains and ISF is complete. Dataflow data processing no longer depends on business domain or ISF services, reducing cross-service coupling and improving Dataflow's suitability for standalone deployment scenarios.

---

### 【Context Loader】

Version 0.5.0 delivers major upgrades in both compression efficiency and action capability architecture.

**1. Return Content Structure Optimization — ≥30% Token Compression**

Context Loader return content structure is comprehensively optimized:

- **Redundancy reduction**: Redundant fields and repeated content are removed from responses, reducing unnecessary token consumption
- **Compressed format support**: TOON compression format output is added; both HTTP and MCP protocols support compressed transfer. MCP connections default to TOON format; HTTP connections default to JSON format for backward compatibility
- **Quantified results**: Average token compression rate ≥30%, agent execution accuracy ≥90%, reduced trial-and-error cost, fewer reasoning rounds

**2. Action Recall Unified via BKN Action Driver**

A significant architectural upgrade to the action recall mechanism, refactoring the `get_action_info` recall logic:

- **Unified ingestion path**: Action recall tools previously connected directly to Execution Factory/MCP proxy execution interfaces; this has been fully migrated to route through the BKN action driver
- **Unified execution management**: Recalled action tools are dispatched and executed through the action driver; action trigger records are visible within the Business Knowledge Network
- **Risk control foundation**: The new architecture establishes the groundwork for implementing unified risk controls at the action driver layer (execution permissions, operation approval, anomaly interception)

**3. Context Loader Decoupled from Business Domains and ISF**

The architectural decoupling of Context Loader from business domains and ISF is complete. Context retrieval and injection processes no longer depend on business domain or ISF permission services, improving standalone deployment flexibility and request chain performance.

---

### 【Execution Factory】

Version 0.5.0 adds Agent Skill ingestion capability, extending the platform's execution capabilities further into the Skill ecosystem.

**1. Agent Skill Ingestion Support**

The Execution Factory now supports ingestion, registration, and invocation of Agent Skills:

- **Skill registration**: Agent Skills can be registered with the Execution Factory and managed uniformly within the execution capability framework
- **Skill configuration**: Provides basic Skill configuration capabilities, supporting parameter definition and runtime environment configuration
- **Skill invocation**: Registered Skills can be invoked by Agents through the standard Execution Factory interface without additional adaptation

This capability provides core execution support for Skill-based deployments in DIP scenarios.

**2. Execution Factory Decoupled from Business Domains and ISF**

The architectural decoupling of Execution Factory from business domains and ISF is complete. Tool scheduling and function execution no longer have a hard dependency on business domain or ISF configuration, supporting more flexible multi-domain and cross-domain deployment scenarios. Also fixed: a bug where Context Loader MCP failed to retrieve the tool list.

---

### 【Decision Agent】

Version 0.5.0 focuses on stability fixes and TraceAI integration.

**1. Stability Fixes**

Multiple stability issues reported from version 0.4.0 have been resolved:

- **Long-term memory fix**: Fixed an error that occurred when using agents with long-term memory enabled; long-term memory now functions correctly in new deployment environments
- **Conversation history fix**: Fixed an issue where previously answered questions were not displayed when re-entering from conversation history after an interruption, restoring full visibility of historical conversations
- **Dolphin scheduling fix**: Fixed intermittent Dolphin scheduling errors during agent usage, improving stability in multi-agent collaboration scenarios
- **Conversation context loss fix**: Fixed an issue where historical context information was lost during agent conversations

**2. TraceAI Knowledge Evidence Chain Integration**

Deep integration between Decision Agent and TraceAI is complete:

- **Conversation message linked to execution chain**: Agent-Factory conversation messages are linked to Agent execution chains, enabling full execution record traceability from the conversation dimension
- **Knowledge evidence chain data extraction**: Context Loader tool responses now include extraction of object-type data, providing TraceAI with structured knowledge source information
- **Default business domain configuration**: Added support for a default business domain configuration to prevent API errors when no business domain ID is provided, lowering the initialization configuration barrier

**3. Decision Agent Decoupled from Business Domains and ISF**

The architectural decoupling of Decision Agent from business domains and ISF is complete. Decision Agent no longer has a hard dependency on business domain or ISF configuration, supporting more flexible multi-domain and cross-domain deployment scenarios.

---

## Resources

### 1. GitHub Packages and Documentation

**KWeaver SDK**
- GitHub: https://github.com/kweaver-ai/kweaver-sdk

**AI Data Platform**
- GitHub Release: https://github.com/kweaver-ai/adp/tree/release/0.5.0

**Decision Agent**
- GitHub Release: https://github.com/kweaver-ai/decision-agent/tree/release/0.5.0

**TraceAI**
- GitHub: https://github.com/kweaver-ai/tracing-ai

---
