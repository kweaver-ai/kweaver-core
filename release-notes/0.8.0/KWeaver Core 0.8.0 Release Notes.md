# KWeaver Core 0.8.0 Release Notes

---

## Overview

### KWeaver Core 0.8.0 Now Available

**Out-of-the-box toolchain · Full-stack observability · Multi-language runtime · Single-node deployment**

KWeaver Core 0.8.0 centers on **developer experience** as its core theme, delivering major advances across three areas: toolchain delivery, observability infrastructure, and sandbox runtime. The Context Loader toolset is now built into the platform as the default toolbox — registration happens automatically on deployment, eliminating the need to manually import it in the Execution Factory. The `search_schema` tool gains a `concept_groups` parameter, allowing semantic recall to be scoped precisely to designated BKN concept groups, resolving boundary confusion when multiple business domains share a single knowledge network. The BKN engine upgrades ActionType semantics to an **Action Intent** declaration, and extends Impact Object Type to a multi-target **Impact Contract**, giving agents a richer model of the business side-effects of any action before execution. On the observability front, `bkn-backend`, `ontology-query`, and `vega-backend` complete full OpenTelemetry trace instrumentation — cross-service call spans are now correlatable in OpenSearch, laying the groundwork for TraceAI's evidence-chain capability. The sandbox runtime advances to v0.4.0 with a new multi-language composite template environment and official Go language support. VEGA ships production-grade rate limiting for concurrent data queries, alongside a comprehensive API restructuring of BuildTask, DiscoverTask, Dataset, and DiscoverSchedule — unifying design conventions and eliminating legacy inconsistencies. On the deployment front, this release introduces single-node deployment via K3s (Linux) and Docker + Kind (macOS), together with a redesigned installation flow featuring pre-flight check scripts and a post-install onboarding wizard, significantly reducing the cost of first-time deployment and ongoing operations.

---

## KWeaver Core 0.8.0 Highlights

**1. Context Loader — Out-of-the-Box with Concept Group Scoped Recall**

The Context Loader toolset is now a built-in platform default, automatically registered at startup without any manual import step in the Execution Factory. The `search_schema` tool gains the `concept_groups` parameter, enabling recall to be bounded to a specified BKN concept group — preventing semantic boundary bleed when multiple business scenarios share the same knowledge network.

**2. Full-Stack Observability Foundation: Three Core Services on OpenTelemetry**

`bkn-backend`, `ontology-query`, and `vega-backend` complete full OpenTelemetry trace instrumentation. HTTP, OAuth, and OpenSearch access layers are migrated to OTel-compatible clients. Spans from multiple microservices involved in a single data query can now be correlated in OpenSearch, establishing a foundation for production diagnostics. TraceAI agent-side call-chain visualization will continue to evolve in future releases.

**3. Deeper BKN Action Semantics: Action Intent and Impact Contract**

The ActionType field is promoted to **Action Intent**, clarifying its role as a declaration of business action intent rather than an execution directive. Impact Object Type is extended to **Impact Contract**, allowing a single action to declare effects on multiple object types and operation kinds — enabling agents to fully understand business side-effects before committing to an action.

**4. Sandbox Multi-Language Runtime Upgrade with Official Go Support**

The sandbox runtime advances to v0.4.0, introducing a multi-language composite template environment with official Go support. Python dependency installation gains a configurable timeout to prevent long-running installs from blocking sessions. A silent failure bug in the s3fs startup script that corrupted working directories is also resolved.

**5. VEGA Production-Grade Rate Limiting: Two-Layer Protection for Data Sources**

VEGA Resource data queries now enforce a complete rate-limiting mechanism with independent global and per-Catalog concurrency limits, returning standard `429 Too Many Requests` responses, fundamentally preventing data source overload and resource contention under concurrent query loads.

**6. Single-Node Deployment — One Command to Install Locally**

KWeaver Core adds a lightweight single-node deployment path using K3s (Linux) and Docker + Kind (macOS). A single command installs the full platform on a local machine or standalone server, suitable for development, evaluation, and resource-constrained production environments. Pre-flight check scripts and a post-install onboarding wizard are included to reduce trial-and-error during initial setup and routine maintenance.

---

## Detailed Release Notes

### 【Context Loader】

0.8.0 closes the delivery loop for the Context Loader toolchain, landing three capabilities simultaneously: toolset built-in, on-demand concept group scoped recall, and an out-of-the-box experience.

**1. Context Loader Toolset Built Into the Platform**

The Context Loader toolset is now a platform-native default toolbox, registered automatically on deployment — no manual installation or configuration required:

- **Auto-registration**: The platform checks and syncs built-in tool dependencies at startup, ensuring the toolset always aligns with the current version contract
- **No manual import**: Eliminates the need to manually import the toolbox in the Execution Factory, removing the post-fresh-install problem where `kn_*` tools are unavailable; agents still need to configure tool parameters as needed
- **Versioned delivery**: The toolbox embeds a contract version (currently 0.8.0); platform upgrades automatically trigger toolset updates to prevent version drift

**2. search_schema Supports Concept Group Scoped Recall**

The `search_schema` tool adds a `search_scope.concept_groups` parameter to scope Schema recall to specific BKN concept groups:

- **Group-scoped recall**: When `concept_groups` is specified, recall of object types, relation types, action types, and metric types is all bounded to that group — different business scenarios can maintain precise semantic boundaries within a shared knowledge network
- **All-type coverage**: `concept_groups` composes orthogonally with existing `include_object_types`, `include_relation_types`, `include_action_types`, and `include_metric_types` flags for fine-grained recall control
- **Backward compatible**: When `concept_groups` is not supplied, the original full-network recall behavior is preserved; existing agents require no changes
- **Note**: Metric concept group filtering depends on BKN metrics-side support. In this release, metric recall carries the concept group scope parameter; actual filtering behavior is governed by the BKN side

**3. CLI Help Experience Improvements**

The Context Loader CLI help system is redesigned in the style of GitHub / Docker CLI conventions:

- **Per-scenario documentation**: All 11 core tool scenarios include dedicated usage notes, parameter descriptions, and recommended workflow hints
- **Tool grouping**: Commands are organized into functional families — schema, query, instance, action, concept_group — reducing cognitive overhead
- **Example-driven**: Every tool includes complete `example` entries with common parameter combinations, ready to copy and run

---

### 【BKN Engine】

0.8.0 delivers important upgrades across action semantic modeling, metrics capability, and the SDK, along with several reliability fixes.

**1. ActionType Upgraded to Action Intent**

The ActionType field is semantically upgraded to **Action Intent**, making its role explicit as a declaration of business action intent rather than a constraint on execution:

- **Semantic clarity**: Action Intent declares only "what this action intends to do" (create / update / delete, etc.); the actual execution logic is determined by the bound Action Source — the two concerns are cleanly separated
- **Value set preserved**: The supported value set remains identical to the original ActionType; existing configurations require no migration and can be extended as business needs evolve

**2. Impact Object Type Upgraded to Impact Contract**

Impact Object Type is upgraded to **Impact Contract**, supporting multiple impact targets declared on a single action:

- **Multi-target declaration**: A single action can declare effects on multiple object types; each contract entry records the affected object type, operation kind, and impact description — fully expressing the business side-effects of an action
- **Decision support**: Before executing an action, an agent can consult Impact Contracts to assess the object scope involved, providing more readable decision context in risk-sensitive scenarios
- **Non-binding declaration**: Impact Contracts are declarative information and do not constrain actual tool execution behavior, keeping the modeling barrier low

**3. Metric Types Support Concept Group Filtered Search**

The `SearchMetrics` API gains concept group filter support, consistent with `SearchObjectTypes` and similar interfaces:

- Metric types can be retrieved within a specified concept group scope, avoiding metric semantic confusion in multi-scenario knowledge networks
- Returns a 404 error (rather than 500) when all specified concept groups are not found, improving error diagnosability

**4. BKN Paradigm Metric Import / Export**

The BKN paradigm completes import/export support for Metric Definitions:

- **tar export**: BKN tar export packages now include MetricDefinition, enabling metric definitions to migrate and distribute alongside the knowledge network
- **tar import**: Import automatically parses and restores MetricDefinition, with batch import and integration test coverage
- **Example updates**: Standard examples such as `k8s-network/metrics` are updated to include metric definition samples

**5. OpenTelemetry Trace Instrumentation Across Three Core Services**

`bkn-backend`, `ontology-query`, and `vega-backend` complete full OpenTelemetry trace instrumentation, laying the data foundation for ongoing TraceAI evidence-chain capability:

- OTel span tracing and error log instrumentation across driver adapters, driven adapters, and logic layers
- HTTP / OAuth / Hydra / OpenSearch / vega-backend access layers migrated to OTel-compatible clients
- Success-branch span status and failure-branch structured error logs added throughout business flows
- Multi-microservice spans from a single data query (e.g., the 24 spans across vega-backend → bkn-backend → ontology-query) are now correlatable in OpenSearch

**6. BKN SDK Upgrades**

kweaver-sdk delivers comprehensive BKN metrics support and several important improvements to the knowledge network construction workflow:

- **bkn metric CLI**: New `bkn metric list` / `bkn metric query` / `bkn metric dry-run` commands enable programmatic BKN metric data querying; `BknMetricsResource` and `MetricQueryResource` API client wrappers are complete in both TypeScript and Python
- **create-from-ds migrated to VEGA Catalog**: Table metadata scanning in `bkn create-from-ds` and `create-from-csv` is migrated from the legacy data-connection API to the VEGA Catalog API (`/api/vega-backend/v1/catalogs/*`); a VEGA catalog id is now required; CLI prompts and help documentation are updated accordingly
- **Transactional creation**: `create-from-ds` / `create-from-csv` now batch-POST ObjectTypes with backend transaction semantics — all-or-nothing; any failure automatically rolls back the created knowledge network, eliminating orphaned data; client-side pre-validation of names is performed before the CSV import phase to surface violations early
- **PK auto-detection improvements**: New `--pk-map` flag for manually specifying primary key columns; auto-detect now fails fast on ambiguity instead of silently falling back to the first column (which could cause up to ~99.8% data loss on low-cardinality columns); schema-declared PRIMARY KEY constraints take precedence over sample-based heuristics

**7. Bug Fixes**

- Fixed a null pointer exception in Indirect Relation Type when `BackingDataSource` is empty, preventing service crashes under certain configurations
- `action-type execute` now returns an explicit error when input parameters are missing, rather than silently discarding them — improving diagnosability when auth parameters are absent in tool call chains
- `ontology-query`: fixed a parse failure on resource calendar-bucket queries that return date strings in metric trend queries

---

### 【TraceAI】

0.8.0 completes the foundational instrumentation for TraceAI's evidence-chain capability, with three core services fully instrumented.

**1. Call-Chain Instrumentation Across Three Core Services**

BKN Engine, VEGA, and Ontology Query complete OTel-based trace instrumentation and reporting — the data collection foundation for TraceAI call-chain analysis is in place:

- Enable call-chain tracing for the relevant services by setting `traceenable` in configuration
- Service name, version, and other OTel standard attributes are centrally configured in each microservice's Helm ConfigMap
- Collected trace data can be queried and correlated via OpenSearch

**2. Current Stage**

TraceAI evidence-chain capability is under active development:

- **Completed**: Core service call-chain data collection and reporting; trace data is queryable in OpenSearch
- **In progress**: Agent-side call-chain visualization and query via the TraceAI agent-observability service — to be delivered in a future release

---

### 【VEGA Data Virtualization】

0.8.0 delivers on three fronts: production-grade rate limiting, a comprehensive API restructuring across core resource domains (BuildTask / DiscoverTask / Dataset / DiscoverSchedule), and extensible custom field support for Catalog and Resource.

**1. Resource Data Query Rate Limiting**

VEGA Resource data queries now enforce a complete concurrent rate-limiting mechanism:

- **Global concurrency limit**: Caps the total concurrent queries across the VEGA service to prevent self-overload; exceeding the global limit returns `429 Too Many Requests` (`ErrGlobalLimitExceeded`)
- **Per-Catalog concurrency limit**: Each data source (Catalog) has an independent maximum concurrent query count, protecting downstream databases and services from burst traffic from a single Catalog; violations return `ErrCatalogLimitExceeded`
- **Independent dual-layer enforcement**: Both limits are evaluated independently and in parallel; a query must satisfy both to proceed
- **Dynamic configuration**: Both global and per-Catalog limits are configurable at runtime without service restart

**2. Catalog / Resource Extension Fields**

Catalog and Resource gain entity-level extension KV storage (`t_entity_extension`), enabling custom business attributes beyond standard fields:

- Extension fields support create, query, and update operations with complete interface and validation coverage
- Database migration scripts align with both MariaDB and DM8

**3. API Restructuring**

Core resource domains — BuildTask, DiscoverTask, Dataset, and DiscoverSchedule — undergo a comprehensive API redesign to unify conventions and eliminate legacy path inconsistencies and dual-key redundancy:

- **BuildTask top-level resource**: Migrated from the nested path `/resources/buildtask/...` to top-level `/build-tasks` with single-key addressing; state transitions replaced by explicit action endpoints `POST .../start` and `POST .../stop`; full CRUD endpoint set added
- **DiscoverTask API consolidation**: Endpoints converged to read-only + cleanup (list / get / delete); new atomic batch delete added; field naming normalized (`scheduled_id` → `schedule_id`)
- **Dataset & DiscoverSchedule consolidation**: Dataset document APIs unified under `/resources/{id}/data`; DiscoverSchedule promoted from a nested path to top-level `/discover-schedules`, removing dual-key paths; legacy `SQL` naming renamed to `Raw` to eliminate semantic confusion
- **Full OpenAPI coverage**: All resource domains — Catalog, Resource, Dataset, BuildTask, DiscoverTask, Query/Discover/TestConnection — now have independent OpenAPI 3.1 documentation; multiple discrepancies between documentation and actual behavior are resolved

**4. Bug Fixes**

- Fixed Logic View related bugs and added SQL validation
- Fixed a parse failure on resource calendar-bucket metric trend queries returning date strings
- Removed the redundant unified query endpoint `/api/vega-backend/v1/query/execute`, consolidating API entry points

---

### 【Execution Factory】

0.8.0 adds management-side Skill content inspection, establishes automated tool naming validation, and delivers full Skill lifecycle CLI and SDK coverage via kweaver-sdk.

**1. Skill Management-Side Content Inspection**

Building on the existing Skill lifecycle management capabilities (version control, full-package update, history rollback), management-side Skill content read access is now available:

- **SKILL.md content read**: Retrieve the complete SKILL.md description and internal file structure of a Skill through the management API — review Skill implementation details without downloading the full package
- **Response mode toggle**: Supports both `url` mode (returns an OSS pre-signed download link) and `content` mode (returns file content inline)
- **OSS auto-sync**: SKILL.md is automatically synced to OSS on Skill registration or update, ensuring management-side reads always reflect the latest content
- **Publish-time name uniqueness check**: Name uniqueness validation now fires only at publish time rather than during editing, reducing friction in iterative Skill workflows

**2. Tool Name Validation**

To improve tool call success rates with large language models (especially the DeepSeek family), upfront tool naming validation is now enforced:

- **Registration-time validation**: Tool names are checked at registration, import, edit, and publish for compliance with DeepSeek tool naming rules (e.g., restrictions on Chinese characters and other special characters)
- **Warning output**: Non-compliant names trigger clear warnings, prompting correction before runtime and preventing silent tool call failures
- **Coverage**: Validation applies to API, Agent, and MCP tool registration paths

**3. Built-In Toolbox Auto-Registration and Upgrade**

Supporting the Context Loader built-in toolset, the Execution Factory completes the ADP-format auto-registration and upgrade flow for built-in toolboxes, ensuring the built-in toolset version stays in sync with the runtime after platform upgrades.

**4. Skill SDK and CLI — Full Lifecycle Management**

kweaver-sdk delivers CLI and SDK coverage for Skill editing and version history management, enabling programmatic management of the full Skill lifecycle:

- **Metadata and package editing**: `skill update-metadata` updates name, description, category, and other metadata; `skill update-package` replaces the content package (SKILL.md or ZIP)
- **Status management**: `skill status` / `skill set-status` for inspecting and toggling publish status; `skill delete` for removing a Skill
- **Version history management**: `skill history` lists historical versions; `skill republish` restores a historical version to draft; `skill publish-history` directly publishes a specified historical version
- **Draft content access**: `skill management-content` / `skill management-read-file` / `skill management-download` cover draft content index, individual file read, and full ZIP download
- Corresponding `SkillsResource` method wrappers are complete in both Python and TypeScript SDKs, covering all new commands

---

### 【Decision Agent】

Decision Agent v0.7.1 ships in the 0.8.0 release cycle, centered on three themes: tool naming compliance, Skill built-in tool registration refactoring, and user documentation.

**1. Tool Name Validation and DeepSeek Compatibility**

To improve tool call success rates on DeepSeek and similar models, built-in tool names are brought into compliance and registration-time validation is established:

- **Existing built-in tool renaming**: Non-compliant built-in tool names are standardized (e.g., `获取agent详情` renamed to `get_agent_detail`); deprecated tools for retired endpoints are removed
- **Registration-time validation**: Tool names are validated at registration to contain only `a-z`, `A-Z`, `0-9`, `_`, `-`, with a maximum length of 64 characters; non-compliant names are logged as warnings to prevent silent runtime failures
- **Dolphin dependency upgrade**: `kweaver-dolphin` upgraded to v0.7.6, resolving DeepSeek v4 compatibility issues

**2. Skill Built-In Tool Registration Refactoring**

The service boundary and assembly mechanism for Skill built-in tools is refactored:

- **Responsibility decoupling**: Skill API capabilities are switched to the Agent Operator Integration service, correcting an incorrect downstream service dependency
- **Explicit registration path**: Skill contract tools are moved into the agent core logic layer; generic tool assembly and platform built-in tool injection are decoupled
- **Interface completeness**: New Skill HTTP endpoints, request/response models, OpenAPI tool definitions, and `skill_tools.json` initialization configuration are added

**3. React Mode Full-Flow CLI Enhancements**

Building on the React Agent mode introduced in 0.7.0, 0.8.0 completes full-flow CLI configuration enhancements for React mode:

- **Config file support**: `agent_mode: react` can now be declared in a configuration file, eliminating the need to pass it on the command line every time — reducing friction for complex configuration scenarios
- **`disable_history_in_a_conversation`**: Explicitly supported in config files; disables in-conversation history for independent task scenarios that don't require context, reducing token consumption
- **Config validation**: A validated Agent `Config` value object is introduced, adding complete structural validation for Agent configuration and Dolphin/React mode; invalid config now returns clear error messages; Agent Mode documentation and repository Agent config guides are updated accordingly

**4. User Documentation and Examples**

- **User manual**: A complete Decision Agent user manual covering API, CLI, concepts, TypeScript SDK guide, and an aggregated reference document, with ready-to-run API, CLI, and SDK examples
- **Cookbook**: Integration Cookbook scenarios covering contract summarization, Sub-Agent contract review, human-in-the-loop interruption/termination, and more
- **Example reorganization**: API / CLI / TypeScript SDK examples and Makefile targets split into capability-scoped directories; shared environment setup and state handling added for easier cross-example workflow reuse

---

### 【Sandbox Runtime】

0.8.0 upgrades the sandbox runtime to v0.4.0, delivering a multi-language composite template environment and official Go support, alongside improvements to image layering, file security, and dependency installation reliability.

**1. Multi-Language Composite Template Environment**

A new built-in `multi-language` template supports Python, Go, and Bash in a composite execution environment — multiple languages can be mixed within a single session:

- **Go 1.25.2 runtime**: Go 1.25.2 is bundled in the multi-language runtime base; Go build and module caches are configured at `/workspace/.cache`; the `go` command is available in both Bubblewrap and subprocess execution paths
- **Per-template Helm overrides**: New `image.defaultTemplates.pythonBasic` and `image.defaultTemplates.multiLanguage` Helm values allow independent image version overrides for the two built-in templates at deploy time, supporting air-gapped and customized deployments
- **Default template auto-fallback**: create-session requests that omit `template_id` automatically use `DEFAULT_TEMPLATE_ID`, removing the need to explicitly specify a template

**2. Stable Runtime Base Image Layering**

The image build architecture is refactored to decouple runtime dependencies from executor application code:

- **Stable base layer**: New Python and multi-language runtime base images contain only system dependencies and language runtimes — they do not rebuild when application code changes, significantly reducing image layer churn
- **Versioned executor layer**: Final executor/template image tags follow the project `VERSION`; the heavyweight runtime layer stays stable while only the lightweight application layer is updated on upgrades
- **Shared executor Dockerfile**: Per-template Dockerfiles are replaced by a single shared executor template Dockerfile, simplifying image maintenance

**3. Session Dependency Install Timeout**

Manual incremental session dependency install requests gain an `install_timeout` parameter, propagated through to the executor session-config sync call:

- Prevents large Skill dependency packages from being truncated by the executor client's default timeout during installation
- File upload size validation is now driven by configuration rather than hardcoded at 100 MB, allowing adjustment per deployment requirements

**4. File Upload Security**

ZIP archive extraction gains multiple security safeguards:

- Dual limits on file count and total decompressed size prevent zip bomb attacks
- Symbolic link entries are rejected during extraction, reducing the risk of unsafe archive processing

**5. Bug Fixes**

- Fixed an issue where the default seed template image address was hardcoded to `v1.0.0` or `latest`; it now follows `VERSION`, `TEMPLATE_IMAGE_TAG`, or `PROJECT_VERSION`, ensuring template images align automatically after version upgrades

---

### 【Deployment & Infrastructure】

**1. KWeaver Core Deployment Flow Improvements**

The KWeaver Core deployment flow is redesigned to reduce the risk of errors:

- Pre-flight check script to detect environment issues early and assist with remediation when needed
- Post-install onboarding wizard script covering basic initialization and setup finalization
- Installation flow and documentation updated in sync, reducing trial-and-error for both first-time deployments and routine operations

**2. KWeaver Core Single-Node Deployment**

KWeaver Core adds a lightweight single-node deployment path based on K3s (Linux) and Docker + Kind (macOS):

- One command installs the complete KWeaver Core platform on a local machine or standalone server — suitable for development, evaluation, and resource-constrained production environments
- One-command environment check, repair, and initialization configuration
- Detailed installation guide: [https://github.com/kweaver-ai/kweaver-core/tree/main/deploy](https://github.com/kweaver-ai/kweaver-core/tree/main/deploy)

**3. OSS Gateway Adds Volcengine TOS Support**

The OSS Gateway adds support for Volcengine TOS (Tinder Object Storage), expanding coverage of mainstream domestic object storage providers.

---

### 【Dataflow】

0.8.0 advances Dataflow with a version upgrade and a new Dataset write node.

**1. Dataset Write Node (`@dataset/write-docs`)**

Dataflow pipelines gain a `@dataset/write-docs` write node, enabling pipeline outputs to be written directly to a VEGA Dataset — closing the last mile of data flow into the knowledge network:

- A Dataset write node can be configured in a Dataflow pipeline to persist upstream node output to a designated Dataset
- Integrates with existing VEGA Resource/Dataset APIs, reusing established data auditing and versioning mechanisms

---

### 【ISF — Information Security Fabric】

0.8.0 delivers several improvements in identity security and deployment dependencies.

**1. Concurrent Multi-Device Login Control**

New security controls for concurrent logins from the same account:

- The same account is prevented from being logged in simultaneously across multiple terminals or browsers, preventing concurrent credential abuse
- After login, the previous login details (including IP address and timestamp) are displayed, helping users quickly identify suspicious access

**2. Auth and Org-Sync Plugin Decoupled from Object Storage**

The authentication and org-structure sync plugin no longer depends on object storage (OSS), lowering the installation barrier and simplifying service dependency configuration in security-sensitive environments.

---

### 【Example Library Updates】

0.8.0 adds two new complete end-to-end examples:

**Example 04: Multi-Agent session_id Propagation (Multi-Agent Custom-Input Propagation)**

Demonstrates passing `session_id` across a multi-agent call chain via custom inputs, enabling shared conversation context across agents — covering a complete session_id end-to-end scenario.

**Example 05: KN-Driven Skill Routing**

Demonstrates how an agent uses BKN knowledge network semantics to dynamically route to the appropriate Skill, with a ready-to-run end-to-end demo and accompanying blog post.

---

## Release Resources

### 1. Installation Packages and Technical Documentation

**KWeaver Core**
- GitHub Release: https://github.com/kweaver-ai/kweaver-core/tree/release/0.8.0

**KWeaver SDK**
- GitHub: https://github.com/kweaver-ai/kweaver-sdk


---
