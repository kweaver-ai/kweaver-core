# KWeaver Core 0.7.0 Release Notes

---

## Version Overview

### KWeaver Core 0.7.0 Is Now Generally Available

**Metric Semantic Fusion, Streaming Build Ready, Multi-Dimensional Agent Capability Leap**

KWeaver Core 0.7.0 is centered on a **dual-direction upgrade to the BKN Engine**. At the semantic modeling layer, the Metric type officially becomes the fourth first-class citizen of the Business Knowledge Network, standing alongside Object Types, Relation Types, and Action Types to express business entities and quantitative logic. Relation Mapping gains a new `FilteredCrossJoinMapping` type, which supports applying filter conditions on the Cartesian product of two data sources to precisely select matching instance pairs — covering complex multi-table join scenarios. The BKN Specification SDK is simultaneously upgraded to v0.1.3, with a new Exchange email recovery domain knowledge network template added. This release also completes VEGA's support for PostgreSQL streaming index builds; with full, incremental, and streaming modes now all covered. The Logic View engine gains multi-table JOIN, UNION view types, and custom SQL capability. The Execution Factory completes the Skill lifecycle management loop, supporting full-package updates, version control, and rollback. Context Loader consolidates its tools and adds Metric type concept recall, providing unified semantic search across Object, Relation, Action, and Metric entities. Decision Agent introduces a React Agent running mode. The overall system also includes multiple targeted improvements to API error semantics, concurrency stability, and deployment/operations experience.

---

## KWeaver Core 0.7.0 — Core Highlights

**1. Dual-Direction BKN Engine Upgrade: Metric Semantic Modeling & Filtered Cross-Join Mapping**

The BKN Engine achieves two major breakthroughs. **Metric Modeling**: The Metric type officially becomes the fourth first-class citizen in BKN, alongside Object Types, Relation Types, and Action Types, supporting calculation formulas, grouping dimensions, time dimensions, and multiple query modes (instant, trend, period-over-period, and proportion). Context Loader simultaneously adds Metric semantic recall, completing the full chain of "metric modeling → semantic recall → agent Q&A." **Relation Mapping**: The new `FilteredCrossJoinMapping` type supports applying filter conditions on the Cartesian product of two data sources to precisely select matching instance pairs, covering complex multi-table join scenarios such as material–inventory–order conditional cross-matching. Together with existing `DirectMappingRule` / `InDirectMappingRule`, all three types form a complete mapping system.

**2. PostgreSQL Streaming Index Build Now Live — Real-Time Data Ingestion**

VEGA adds PostgreSQL streaming index builds, complementing the existing full and incremental build modes with a real-time streaming mode: row-level changes (INSERT / UPDATE / DELETE) in the database are instantly synchronized to the search index via the PG change-listening mechanism, with no need for scheduled full refreshes. VEGA now fully covers all three build modes for PG data sources — batch (full, incremental) + streaming — meeting real-time analytics scenarios with strict data freshness requirements.

**3. Execution Factory — Full Skill Lifecycle Management with Version Control & Rollback**

The Execution Factory upgrades Skill management from "register-and-execute" to a complete lifecycle: supporting metadata editing (draft/published state separation), full-package version updates, historical version queries, and one-click rollback for already-published Skills, along with Skill Dataset rebuild compensation and incremental build. Organizations can iterate Skills without interrupting live services, with version management ensuring every change is traceable and revertible.

**4. Decision Agent Adds React Running Mode — Enhanced Complex Task Reasoning**

Decision Agent introduces a React Agent running mode (`agent_mode: react`), adding an independent `ReactConfig` (supporting conversation history toggle and LLM cache control) on top of the existing Dolphin mode. React mode provides stronger context continuity in multi-step reasoning and longer tool-call-chain scenarios; combined with the new configurable built-in Skill registration rules, agent execution efficiency for complex decision tasks is significantly improved.

**5. Logic View Engine Extended: Multi-Table JOIN / UNION & Custom SQL**

VEGA Logic Views gain two new view types — multi-table JOIN (left join, etc.) and multi-table UNION — defined through configuration nodes that declare data sources, join keys, and output fields, allowing multi-table query modeling without writing SQL. A custom SQL view mode is also added, supporting full SQL statement definitions. All three view types are exposed through VEGA's unified query interface and seamlessly support MariaDB, PostgreSQL, and OpenSearch data sources.

---

## Detailed Feature Descriptions

### 【BKN Engine】

Version 0.7.0 completes significant upgrades in two directions — Metric modeling and Relation Mapping — while also improving boundary validation and error semantics.

**1. BKN Metric Model: Business-Semantic Metric Definition & Querying**

BKN adds Metric Type, enabling definition of quantitative metrics with business semantics directly within the knowledge network:

- **Metric Structure Modeling**: A metric is fully described by `scope` (the associated Object Type), `calculation formula` (filter conditions + aggregation method + grouping fields), `time dimension`, and `analysis dimensions`, expressing business quantitative logic completely. Supports the `atomic` metric type; composite metric types will be added in future iterations.
- **Multi-Mode Querying**: Metrics support instant queries (`instant=true`, retrieves current aggregate value), trend queries (`instant=false` + time range, returns time-series data at calendar steps such as day/month/year), period-over-period analysis (`type=parallel`, calculates growth value and growth rate with configurable offsets), and proportion analysis (`type=proportion`, returns percentage breakdown per analysis dimension).
- **Semantic Retrieval**: Metrics support semantic concept retrieval — vector indexes are generated from ID, name, and description, enabling natural-language-based metric discovery.
- **Independent Concept Status**: Metrics exist as an independent concept type in BKN, without direct associations with Object Types or other concept types, but can reference a specific Object Type via `scope` to define the statistical subject.

**2. Condition Configuration Refactor & Validation Improvements**

Refactors the BKN Condition configuration structure, separating `ActionCondCfg` from `CondCfg` into an independent struct, and standardizing the field name from `Name` to `Field` for clearer semantics:

- Improves Condition validation for Object Type logical attribute binding resources in strict mode (`strict_mode`).
- Auto-completes system parameters for Metric Type attributes, reducing manual configuration overhead.
- Optimizes Relation Type validation logic with new non-strict mode support.
- Updates Action source type valid values from `tool/map` to `tool/mcp` to align with the current tooling system.

**3. FilteredCrossJoinMapping: Filtered Cross-Join Relation Mapping**

BKN Relation Types gain a new `FilteredCrossJoinMapping` mapping type, supporting the application of filter conditions on the Cartesian product of two data sources to precisely select instance pairs that satisfy the join condition:

- Resolves the limitation of prior relation mappings, which could only express simple field equality joins, by allowing inline declaration of complex multi-field filter logic within Relation Types.
- Applicable to "select valid associations from the combination of Resource A and Resource B based on business rules" scenarios, such as conditional cross-matching of material, inventory, and order records.
- Listed alongside existing `DirectMappingRule` / `InDirectMappingRule`; together, the three mapping types cover simple mapping, indirect mapping, and filtered cross-join mapping.

**4. API Error Semantics Standardization**

Standardizes the HTTP status code for resource-not-found scenarios to 404 (previously 403), covering all resource types including knowledge networks, concept groups, action types, relation types, and object types. The `action-type execute` API now returns an explicit 400 error when required `input` parameters are missing (previously silently discarded), improving diagnosability along agent error paths.

---

### 【VEGA Data Virtualization】

Version 0.7.0 delivers significant capability extensions in three areas: data build modes, Logic View capabilities, and data source integration.

**1. PostgreSQL Streaming Index Build**

VEGA adds PostgreSQL streaming index builds, completing three build modes for PG data sources:

- **Full Build**: Synchronizes the entire PG table data to the search index in a single pass.
- **Incremental Build**: Syncs only new or changed data since the last build, based on an incremental field.
- **Streaming Build**: Captures row-level changes (INSERT / UPDATE / DELETE) in real time via the PG change-listening mechanism, synchronizing data changes to the index immediately — meeting millisecond-level data freshness requirements.

> Note: Streaming build requires PG Logical Replication to be enabled on the server side. If not enabled, build tasks will return an error prompt. Please verify your database configuration before enabling this mode.

**2. Logic View Extended: New JOIN, UNION, and Custom SQL Types**

Logic Views expand from basic queries to support three complete view types:

- **JOIN View**: Supports multi-table left joins and other join operations. Declare source Resources, join keys, and output fields through configuration nodes; the system automatically translates the configuration into the corresponding SQL. Supports MariaDB and PostgreSQL data sources.
- **UNION View**: Supports multi-table UNION queries, merging data from different sources with the same structure into a unified view. Supports mixed MariaDB and OpenSearch data sources.
- **Custom SQL View**: Supports full SQL statement definitions using template syntax (`{{.nodeId}}`) to reference Resources declared in the configuration, balancing flexibility and safety.

All Logic Views are exposed through VEGA's unified query interface (`/resources/{id}/data`), with support for pagination, sorting, condition filtering, and `limit` parameters.

**3. AnyShare Document Library Integration & Filter Condition Support**

Building on the AnyShare Knowledge Base integration in 0.6.0, VEGA Catalog now adds support for the **AnyShare Document Library** type in 0.7.0, enabling enterprise document data (Word, PDF, and other file formats) to be managed within the Business Knowledge Network. A new `cond` condition filter parameter is also added, supporting content filtering within document libraries/knowledge bases at the Catalog discovery stage based on business conditions, reducing unnecessary data ingestion.

**4. Enhanced Resource Data Query Capabilities**

The Resource Data query interface (`/resources/{id}/data`) adds aggregate analysis support:

- **Aggregation / Grouping**: Supports `GROUP BY` grouped statistics.
- **Having Filtering**: Supports secondary condition filtering on aggregation results.
- **Query Parameter Fix**: Fixes an issue where the `limit` parameter did not take effect in some scenarios, ensuring pagination behavior matches expectations.

---

### 【Execution Factory】

Version 0.7.0 completes the full Skill lifecycle management loop, upgrading from simple register-and-execute capability to a comprehensive lifecycle management system that supports version control, metadata iteration, and data compensation.

**1. Skill Full-Package Update & Metadata Editing**

Supports two update paths for published Skills:

- **Metadata Editing**: Modify a Skill's name, description, category, and other metadata without replacing the Skill execution package. Changes generate a draft version (edit state) that coexists with the currently published version, without affecting live calls.
- **Full-Package Update**: Re-upload the complete Skill package to replace the existing implementation with a new version. The update also enters the draft state first and is published to production upon confirmation.

Both update paths take effect through an explicit publish operation, ensuring stability of the live version.

**2. Version Management & Historical Rollback**

Each publish operation records a version history entry, supporting the following version management operations:

- **Version History Query**: View the complete version change history for a specified Skill.
- **Historical Version Restore to Draft**: Restore any historical version to the draft box as a candidate for the next release.
- **Direct Historical Version Publish**: Publish a historical version directly as the live version, enabling rapid rollback.

Only one official published version is maintained at any time. The Dataset always stores index data corresponding to the latest live version.

**3. Skill Dataset Rebuild Compensation & Incremental Build**

Adds a Skill Dataset rebuild compensation mechanism: after a Skill version update, the system automatically triggers a rebuild or incremental update task for the associated Dataset, ensuring Dataset index data always stays consistent with the currently published version. Incremental builds process only the portions changed since the last build, reducing the performance overhead of full rebuilds.

---

### 【Context Loader】

Version 0.7.0 completes tool consolidation and capability expansion, incorporating Metric types into the unified semantic recall system and further reducing agent invocation costs.

**1. search_schema Tool Merge: Unified Recall Endpoint for Four Concept Types**

Merges the original `kn_schema_search` and `kn_search` tools into a unified `search_schema` tool, covering unified semantic recall for all four BKN concept types: Object Types, Relation Types, Action Types, and **Metric Types**:

- **search_scope Parameter**: Flexibly controls the recall scope via four toggles — `include_object_types`, `include_relation_types`, `include_action_types`, `include_metric_types` — with all types enabled by default.
- **Compact Response Mode**: Supports the `schema_brief` parameter; when enabled, returns a compact concept summary format to reduce token consumption.
- **Response Format**: HTTP interface defaults to JSON; MCP Tool defaults to `toon` compressed format. The original two tools remain in the tool set but are not exposed externally via MCP.

**2. Metric Type Concept Recall**

Via the `search_schema` tool, agents can now discover metrics defined in the knowledge network through natural language queries:

- Returns the metric's `id`, `name`, `comment` (business description), and `metric_type`.
- Supports semantic vector matching based on metric name and description.
- Supports mixed recall: queries with semantics like "concept count" can simultaneously return Object Types and related metrics, with cross-scoring to determine the final ranking.

**3. Zombie Dependency Cleanup**

Removes the `data_retrieval` zombie dependency in `agent-retrieval` that was bypassed by a hardcoded feature flag. Cleans up invalidated configuration entries in the codebase, ensuring the tool recall path behavior is consistent with code declarations.

---

### 【Decision Agent】

Version 0.7.0 completes important upgrades in two dimensions — running modes and configurability — while also simplifying the observability system to improve operational maintainability.

**1. React Agent Running Mode**

Adds the `agent_mode: react` running mode, creating React Agent configurations via the `/v3/agent/react` endpoint:

- **ReactConfig**: An independent React mode configuration item supporting `disable_history_in_a_conversation` (disables in-conversation history) and `disable_llm_cache` (disables LLM caching).
- **AgentMode Enum**: Adds `default / dolphin / react` three-mode enum, reserving extension points for future mode expansion.
- React mode provides stronger context continuity in multi-step tool-call-chain scenarios and is suitable for complex decision tasks that require precise tracking of reasoning steps.

**2. Built-in Skill Registration Rules Made Configurable**

Supports controlling the registration behavior and invocation rules for built-in Skills via configuration files (`skill_enabled` configuration item), flexibly adapting to the varying scope of built-in capability exposure needs across different deployment environments.

**3. Agent Template Copy Fix**

Fixes an issue where `published_at` and `published_by` fields were not cleared when copying an Agent template, ensuring that newly copied templates are always initialized in an unpublished state, avoiding status confusion.

**4. OAuth Bearer Forwarding Fix**

Fixes an issue where `agent-executor`'s OpenAPI Tool did not forward OAuth Bearer Tokens, ensuring that OAuth-protected downstream services registered through Toolbox can be properly called by Decision Agent.

**5. Observability System Simplification**

Migrates the observability implementation from custom O11Y tracing to standard OpenTelemetry (OTLP) tracing. Removes the outdated `observability handler` and related redundant components, unifies the telemetry configuration structure across services (`agent-factory`, `agent-executor`, `agent-memory`), and reduces operational complexity. Also adds LLM message logging (`LLMMessageLoggingConfig`) for capturing complete LLM input/output messages during development and debugging.

**6. Deprecated Configuration Cleanup**

Removes the following deprecated service configurations: `kn_data_query`, `kn_knowledge_data`, `data_connection`, `search_engine`, `ecosearch`, `ecoindex_public`, and others. Removes the `disable_biz_domain_init` configuration item to simplify business domain initialization logic.

---

### 【Sandbox Runtime】

Version 0.7.0 adds Control Plane Takeover capability to the Sandbox, significantly improving upgrade stability in Kubernetes environments.

**1. Control Plane Takeover**

Adds Control Plane takeover capability for existing session Pods on startup:

- After a Control Plane restart or upgrade, automatically scans and takes over existing session Pods still running in the Kubernetes cluster, restoring the binding between sessions and executors.
- Uses the Pod Owner Reference mechanism to achieve affinity binding between executor Pods and Control Plane instances, preventing orphaned Pods from continuously consuming resources.
- Handles same-name Pod rebuild conflicts: waits for Pods in the `Terminating` state to be fully deleted before creating new Pods, eliminating race conditions during startup state synchronization.

**2. Session Pod Resource Configuration Optimization**

In K8s scheduling mode, the CPU and memory Requests for session Pods are adjusted to zero (Limits are retained), reducing Pod scheduling failures caused by per-session resource reservations and improving scheduling success rates in high-density session scenarios.

---

### 【ISF Information Security Fabric】

Version 0.7.0 completes ISF service consolidation and capability extension. The `sharemgnt-single` standalone service is removed, with its functionality unified into the main ISF service, reducing the overall number of deployed services and memory footprint.

**1. Personal Access Token (PAT) Support**

The Authentication module adds Personal Access Token (PAT) capability, supporting PAT creation, querying, and revocation in the ISF management interface:

- Each account can create up to 10 PATs (configurable).
- Supports both permanent validity and specified expiration times.
- Token lifecycle is managed uniformly through Hydra OAuth2, deeply integrated with the platform authentication system.
- Database migration supports MariaDB, DM8, and KDB9.

**2. Bulk User Import / Export**

UserManagement adds bulk user management capability, supporting batch account creation via Excel files:

- **Bulk Import**: Import users via a standard Excel template, supporting fields such as name, department (multi-level, separated by `/`), phone number, validity period, storage quota, initial password, and user security level. Red fields are required; black fields are optional.
- **Bulk Export**: Export the current user list to Excel, supporting asynchronous background task execution so that large-volume operations do not block the foreground.
- Internationalization support: Simplified Chinese and Traditional Chinese instructions are embedded in the template, with no additional documentation needed.

---

### 【BKN Specification SDK】

The BKN Specification SDK is an independently released open-specification toolkit for programmatically parsing, building, and serializing BKN Business Knowledge Network specification documents. Version 0.7.0 delivers stable releases, completing an underlying refactor and expanding the example library.

**1. Parser & Serializer Underlying Refactor**

- **Parser Refactor**: Introduces `extractSectionsWithDesc` / `buildDescription` to handle document structure uniformly. Summaries are now automatically extracted from the first sentence of the description content, eliminating the need for a separate `summary` field declaration.
- **Serializer Optimization**: Extracts a common `encodeYAMLBlock` function, removes redundant fields (`sub_conds`, `value_from`), and standardizes table formatting conventions.

**2. Exchange Email Recovery Domain Template & Example Library Expansion**

Adds an Exchange email recovery domain knowledge network template, covering complete object types including data domains, Exchange servers, backup time points, recovery tasks, and recovery jobs. Includes two risk action types — backup time point validation failure and production data overwrite — along with complete scripts, providing a ready-to-use starting template for enterprise email backup and recovery scenarios. Existing examples (k8s-network, supplychain-hd, mock_system) are updated to support the `filtered_cross_join` relation type and the new format conventions.

---

### 【KWeaver SDK】

Version 0.7.0 delivers multiple capability extensions to the KWeaver SDK, focusing on completing the Toolbox/Tool command set, improving the Action Type execution experience, and aligning Python SDK authentication.

**1. Toolbox / Tool Commands Now Available**

Adds the `kweaver toolbox` and `kweaver tool` command families, covering complete lifecycle management for toolboxes and tools:

- `kweaver toolbox create / list / publish / unpublish / delete`: Toolbox creation, publishing, and unpublishing management.
- `kweaver tool upload / list / enable / disable`: OpenAPI Spec upload, tool enabling/disabling.

Combined with the new `kweaver call -F <key>=@<file>` multipart file upload support, this fully replaces the original manual `curl + call` workflow in `examples/03-action-lifecycle`.

**2. kweaver tool execute / debug**

Adds `kweaver tool execute` and `kweaver tool debug` subcommands, encapsulating the header / query / body / timeout enum structures required for Toolbox proxy calls:

- Automatically injects the current logged-in Bearer Token into the envelope, eliminating manual authentication header handling.
- Resolves Authorization header forwarding restrictions in the `agent-operator-integration` service.
- Python / TypeScript SDKs simultaneously add `ToolboxesResource.execute()` / `.debug()` methods.

**3. BKN Action-Type Execution Experience Upgrade**

Adds the `kweaver bkn action-type inputs <kn> <at>` command, which lists all parameters with `value_from=="input"` and their type-aware starter templates for direct use with `--dynamic-params`. The `bkn action-type execute` command adds a flag-form mode (`--dynamic-params / --instance / --trigger-type`), eliminating the need to manually assemble JSON envelopes, while maintaining backward compatibility with the original positional argument form.

**4. Python SDK Authentication Full Alignment**

The Python SDK completes authentication mechanism alignment with the TypeScript CLI:

- Adds `HttpSigninAuth`, implementing browser-free HTTP login via RSA password encryption + OAuth2 redirect chain, covering edge cases such as `200+JSON` redirects.
- The `~/.kweaver` storage format (`displayName`, `logoutRedirectUri`, ISO-Z timestamps) is byte-level aligned with the TS CLI; `kweaver auth list/whoami` can correctly recognize sessions written by Python.
- Adds `change_password` method, supporting forced initial password changes (HTTP 401 code `401001017` automatically triggers guidance).

**5. kweaver agent skill add / remove / list**

Adds an Agent Skill membership management command family, supporting direct Skill configuration for Agents via CLI, replacing the original workflow of manually editing Agent JSON config:

- `kweaver agent skill add <agent_id> <skill_id>`: Add a Skill binding to an Agent.
- `kweaver agent skill remove <agent_id> <skill_id>`: Remove a Skill binding.
- `kweaver agent skill list <agent_id>`: View the list of Skills currently bound to an Agent.

---

### 【Deployment & Infrastructure】

**1. kweaver-admin: Platform Admin CLI Now Available**

`kweaver-admin` is a standalone command-line tool for platform administrators. In environments with ISF (Information Security Fabric) installed, it supports complete user, role, and model management operations via CLI, without requiring login to the Web console.

```bash
npm install -g @kweaver-ai/kweaver-admin
kweaver-admin auth login https://your-platform.example/
```

Core capabilities cover the following management domains:

- **User Management** (`user`): List, query, create, update, and delete users; admin password reset; view assigned roles for a user.
- **Role Management** (`role`): List roles and members; assign/revoke roles for users.
- **Organization / Department Management** (`org`): List, tree view, create, update, and delete departments; view department members.
- **Large Model Management** (`llm`): Add, edit, delete, and test large model configurations.
- **Small Model Management** (`small-model`): Add, edit, delete, and test small model configurations.
- **Audit Query** (`audit`): Query login audit events by user and time range.
- **Raw HTTP Calls** (`call` / `curl`): Arbitrary API calls with authentication headers included.

`kweaver-admin` also ships as an Agent Skill, installable via `npx skills add` into Agent workflows that support skill loading, enabling AI assistants to perform platform management operations using natural language:

```bash
npx skills add https://github.com/kweaver-ai/kweaver-admin --skill kweaver-admin
```

**2. Stability Fixes**

- Fixes deadlock issues in DAG variable creation (`CreateDagVars`) and index refresh (`refreshDagIndexes`), eliminating probabilistic blocking under concurrent operations.
- Fixes data insertion exceptions in `vega-backend` caused by a redundant unique index (`uk_catalog_source_identifier`).
- Fixes an issue where model factory application accounts could not delete small models.

---

## Release Information

### 1. GitHub Packages & Technical Documentation

**KWeaver Core**
- GitHub Release: https://github.com/kweaver-ai/kweaver-core/tree/release/0.7.0

**KWeaver SDK**
- GitHub: https://github.com/kweaver-ai/kweaver-sdk

**kweaver-admin**
- GitHub: https://github.com/kweaver-ai/kweaver-admin
- npm: `npm install -g @kweaver-ai/kweaver-admin`

---
