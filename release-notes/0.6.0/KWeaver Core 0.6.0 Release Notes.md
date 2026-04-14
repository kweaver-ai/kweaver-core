# KWeaver Core 0.6.0 Release Notes

---

## Overview

### KWeaver Core 0.6.0 is Now Available

**Enterprise Decision Intelligence Ecosystem — End-to-End Skill Pipeline**

KWeaver Core 0.6.0 is centered on **end-to-end Skill pipeline integration**, connecting Skill capabilities across every platform layer: from business knowledge network modeling and Context Loader semantic retrieval, through the Execution Factory unified execution layer, to Decision Agent perception and invocation — forming a complete closed loop. This release also delivers seamless integration of VEGA with the AnyShare enterprise knowledge base, bringing mainstream enterprise unstructured data sources under unified business knowledge network management. The sandbox runtime gains archive upload with auto-extraction and shell script execution capabilities, greatly expanding agent code execution flexibility. Deployment memory footprint has been reduced to 24 GB. `kweaver-eval` ships its first Acceptance module, covering 104 test cases across 6 core modules with a dual-scoring framework (deterministic assertions + Agent Judge), marking a new phase of systematic quality governance for KWeaver Core.

---

## Highlights

**1. End-to-End Skill Pipeline — Skills Integrated Across All Platform Layers**

0.6.0 completes the full Skill pipeline across KWeaver Core: BKN supports modeling Skill object types within knowledge networks; Context Loader adds the `find_skills` tool for semantic Skill candidate retrieval within knowledge network boundaries; the Execution Factory adds a Skill execution endpoint with dual Dataset writes; Decision Agent gains native Skill loading from the Execution Factory; and KWeaver SDK fully integrates with the Execution Factory Skill management module. Skills are no longer isolated functional units — they are governable, retrievable, and executable agent capabilities deeply integrated with business knowledge networks.

**2. AnyShare Enterprise Knowledge Base Integration — Unified Unstructured Data Source Management**

VEGA completes Catalog integration with AnyShare enterprise knowledge bases, supporting Discover commands to scan and enumerate accessible knowledge bases (up to 1,000 entries per page), and modeling AnyShare document data into the business knowledge network via BKN object types. Three authentication methods are supported — application account, permanent token, and SSO single sign-on — enabling enterprise-grade permission-aware queries. The Resource layer adds real-time streaming incremental builds and custom view support, continuously enhancing data virtualization capabilities.

**3. Sandbox Runtime Upgrade — Expanded Agent Execution Boundaries**

Sandbox 0.3.2 adds archive upload with auto-extraction, supporting ZIP packages uploaded to Session Workspaces with automatic extraction to specified paths, including built-in path traversal protection. Shell script execution is now supported with a `working_directory` parameter, covering chained commands, relative path execution, and other complex scenarios. Together, these capabilities give agents complete Skill package deployment and script execution abilities in code execution contexts.

**4. KWeaver SDK — Skill Management and Dataflow CLI, Significantly Improved Developer Experience**

KWeaver SDK now fully integrates the Execution Factory Skill management module, adding the `kweaver skill` command family covering the complete Skill lifecycle: registration, installation, lookup, read, status query, and content download. The new `kweaver dataflow` command family supports uploading local files or triggering Dataflow unstructured data processing pipelines via remote URLs, enabling the SDK to cover the full end-to-end chain of "file upload → document parsing → knowledge network construction → agent Q&A". The new `kweaver explore` command launches a local browser SPA with four views: platform status overview, BKN browser, Decision Agent streaming chat (with Trace), and VEGA data preview. Multi-account profile management is also introduced.

**5. Deployment Memory Reduced to 24 GB — Lower Enterprise Adoption Barrier**

0.6.0 reduces full-stack deployment memory (ADP + ISF + Core) to under 24 GB, significantly lower than previous versions. Combined with the `kweaver-eval` first Acceptance module — 104 test cases across 6 core modules using a dual-scoring framework — this release provides a strong engineering foundation for ongoing platform quality governance.

---

## Detailed Feature Notes

### [VEGA Data Virtualization]

0.6.0 delivers significant capability expansions in both data source connectivity and Resource construction. AnyShare enterprise knowledge base integration is the most critical VEGA breakthrough in this release.

**1. AnyShare Knowledge Base Integration (Catalog)**

VEGA Catalog now supports AnyShare enterprise knowledge base connectivity. The Discover command automatically scans the list of knowledge bases accessible to the current account (up to 1,000 entries per page) and retrieves complete Resource metadata for each (name, ID, type, creator, modifier, DNS address, field definitions, etc.).

Three authentication methods are supported to accommodate different enterprise deployment scenarios:
- **Application Account**: AppID + Secret, suitable for integrated deployments
- **Permanent Token**: suitable for high-privilege operations environments
- **SSO Single Sign-On**: integrated with the KWeaver account system; the current user's permissions are passed through to AnyShare, enabling user-level data permission isolation

The current version prioritizes AnyShare knowledge library type; document libraries and other types will be covered in subsequent iterations.

**2. Real-Time Streaming Incremental Resource Builds**

Resource now supports real-time streaming incremental build mode, enabling incremental data sync based on incremental fields in addition to full builds. Compared to full builds, incremental builds process only data added or changed since the last build, significantly reducing data update costs for time-sensitive scenarios.

**3. Custom Resource Views**

Custom views can now be defined for Resources, allowing flexible configuration of field mappings and vector index structures to meet different business retrieval needs, further enhancing VEGA's flexibility at the data virtualization layer.

**4. Large Dataset Query Stability Optimization**

Targeted performance optimization for Catalog and Resource list interfaces under large dataset conditions, improving system stability for large-scale knowledge base integration scenarios.

---

### [BKN Engine]

0.6.0 completes BKN integration with VEGA Resources, allowing object types in business knowledge networks to be directly bound to VEGA data resources, connecting the business semantic modeling layer with the data virtualization layer.

**1. BKN Integration with VEGA Resources**

BKN object types can now be bound to VEGA Resource data sources (supporting AnyShare knowledge bases, MySQL, Dataset, and more), enabling unified querying and retrieval of data from multiple heterogeneous data sources through the business knowledge network. Once bound, real data in the corresponding Resource can be retrieved via BKN's `ontology_query` interface, closing the loop between semantic modeling and data access within the same knowledge network.

Two object creation modes are supported:
- **Strict Mode (`strict_mode=true`)**: validates the existence of the associated Resource when creating an object type; returns an error if not found
- **Non-strict Mode (`strict_mode=false`)**: skips dependency existence checks and creates the object type directly, suitable for phased modeling workflows

**2. Stability Fixes**

Fixed an error when updating object types or action types without an associated branch; fixed abnormal list ordering after binding a Resource to an object type, improving stability of business knowledge network editing operations.

---

### [Context Loader]

0.6.0 adds the `find_skills` tool, bringing Skill candidate discovery into the Context Loader toolset and closing the loop on Skills at the semantic retrieval layer.

**1. find_skills: Skill Candidate Discovery within Knowledge Network Boundaries**

The new `find_skills` tool supports discovering Skill candidates within a specified `kn_id` and business context boundary, completing coverage of the Context Loader `find_*` semantic tool family.

Key features:
- **Dual Retrieval Modes**: supports object-type-level retrieval (find Skills under a specified object type) and instance-level retrieval (find Skills associated with a specific instance); network-level retrieval will be released in a future version
- **Minimal Metadata Response**: returns only `skill_id`, `name`, and `description`, minimizing token consumption
- **Basic Contract Validation**: automatically validates that the `skills` ObjectType exists in the knowledge network and contains at least `skill_id` and `name` data properties at the entry point; returns a clear error if not satisfied, preventing invalid retrievals

> Note: `find_skills` is a candidate resource discovery tool and does not replace `kn_search` / `query_object_instance`. `object_type_id` is a required parameter. Skills must already be registered and bound in the knowledge network for retrieval results to be reliable.

**2. Unified Response Format Support**

A new `response_format` parameter supports `json` and `toon` formats:
- HTTP interfaces default to `json` for backward compatibility
- MCP Tools default to `toon` compressed format to reduce token consumption
- Error responses always use JSON format

The `x-account-id` and `x-account-type` parameter conventions across Context Loader tools have also been unified to simplify caller integration.

---

### [Execution Factory]

0.6.0 adds a Skill execution endpoint, bringing Skill execution under unified Execution Factory management with support for dual Dataset writes of execution results.

**1. Skill Execution Endpoint and Dataset Dual-Write**

The Execution Factory adds a dedicated Skill execution endpoint, supporting invocation of registered Skills through a unified interface and synchronous writes of execution results to Dataset, enabling persistent storage of Skill output data. Dataset data can subsequently be queried by BKN object type bindings, forming a complete data loop of "execute → store → retrieve".

**2. Stability Fixes**

Fixed a failure in composite operator creation; fixed a nil pointer dereference panic when streaming forwarding is missing a ResponseWriter; optimized service startup logic to avoid index initialization blocking startup, switching to a background retry mechanism to improve service availability.

---

### [Decision Agent]

0.6.0 completes Decision Agent's full integration with Execution Factory Skill capabilities, giving the agent the complete ability to perceive, load, and invoke Skills from the knowledge network.

**1. Native Skill Loading in Agents**

`agent-factory` adds Skill type support, with full Skill configuration support across agent create, detail view, update handlers, and run services. Agents can now read and load available Skills from the Execution Factory at runtime, enabling native Skill perception and invocation.

Related database migrations are complete (v0.6.0 Skill-related tables and `agent-memory` history tables, covering DM8 and MariaDB).

**2. TraceAI Evidence Chain Support**

`agent-executor` adds TraceAI Evidence request header support, introducing an `enable_traceai_evidence` feature flag (configured in `FeaturesConfig`). When enabled, `X-TraceAi-Enable-Evidence` is automatically injected into API tool proxy requests, enabling Evidence data collection across all call chains beyond Decision Agent for unified full-chain evidence extraction.

**3. Publish Request Validation Improvement**

Refactored `agent-factory` publish request validation to use constructor semantics for field validation and sanitization. Validation failures now return clear 400 errors instead of 500, improving API error diagnosability.

---

### [Sandbox Runtime]

0.6.0 corresponds to Sandbox 0.3.2, adding archive upload and Shell execution capabilities that significantly expand agent flexibility and applicability in code execution scenarios.

**1. Archive Upload and Auto-Extraction**

Adds Session Workspace archive upload capability, supporting ZIP archives uploaded to specified Workspace paths with automatic extraction:

- **Overwrite Control**: configurable conflict behavior; upload responses include overwrite statistics and extraction result metadata
- **Path Safety**: built-in path safety validation rejects illegal entries and path traversal content (Path Traversal protection)
- **Skill Package Deployment**: combined with Skill execution capabilities, supports one-shot upload and deployment of entire Skill dependency packages to the Sandbox environment

**2. Shell Script Execution**

Adds `language=shell` execution mode in `execute` and `execute-sync` endpoints:

- **Working Directory Control**: optional `working_directory` parameter specifies the script execution working directory
- **Command Normalization**: automatically normalizes incorrectly prefixed `bash/sh` commands
- **Scenario Coverage**: supports chained commands, relative path execution, and other complex Shell script scenarios with complete end-to-end test coverage

**3. Helm Chart Image Configuration**

Unified Sandbox Helm Chart image configuration to the top-level `image` values structure, enabling offline packaging tools to fully extract all images from Chart values, simplifying offline installation for private deployments.

---

### [KWeaver SDK]

0.6.0 completes KWeaver SDK integration with the Execution Factory Skill management module, and adds `kweaver explore` interactive exploration and multi-account management capabilities to further enhance developer experience.

**1. Skill Management Commands**

SDK CLI adds the `kweaver skill` command family, covering the complete Skill lifecycle from registration to execution:

- `skill list`: list currently available Skills
- `skill market`: browse the Skill marketplace
- `skill get`: retrieve configuration details, input/output parameter definitions for a specified Skill
- `skill register`: register a Skill to the Execution Factory under unified execution management
- `skill install`: trigger Skill dependency installation and runtime environment initialization
- `skill status`: query Skill installation/runtime status
- `skill content` / `read-file`: read Skill content and associated files
- `skill download`: download the Skill package

**2. kweaver explore: Local Visual Platform Exploration**

The new `kweaver explore` command launches a local browser SPA with four interactive views: platform status aggregation overview, BKN knowledge network browser (object types / relation types / instances / subgraphs), Decision Agent streaming chat (with real-time execution Trace), and VEGA Catalog and data preview. Developers can quickly validate platform data and agent behavior without opening the full platform UI.

**3. Dataflow CLI — End-to-End Unstructured Data Pipeline**

The new `kweaver dataflow` command family supports operating Dataflow document processing pipelines directly from the SDK:

- `dataflow list`: list all Dataflow DAGs
- `dataflow run <dagId> --file <path>`: upload a local file (PDF, Word, etc.) to directly trigger unstructured data processing
- `dataflow run <dagId> --url <url> --name <filename>`: trigger via remote URL without local upload
- `dataflow runs <dagId>`: view run history for a specified DAG, with `--since` time filtering
- `dataflow logs <dagId> <instanceId>`: view step-by-step execution logs; `--detail` expands full input/output payloads per step

Combined with Dataflow document parsing nodes, Execution Factory Skill dual-write to Dataset, BKN object type binding, and Context Loader semantic retrieval, the SDK now covers the complete end-to-end pipeline of "file upload → document parsing → knowledge network construction → agent Q&A".

**4. Multi-Account Management and Auth Enhancements**

Multi-account profile management support is added, allowing developers to manage login sessions for multiple KWeaver instances on a single machine:

- `--alias` parameter names login accounts; `auth use` switches between accounts quickly
- Global `--user` flag enables per-command credential override without changing the global account
- `--port` and `--redirect-uri` support custom OAuth2 callback addresses for complex network environments
- Stale client auto-recovery reduces re-authentication frequency

---

### [TraceAI]

0.6.0 corresponds to TraceAI 0.2.2, completing Helm Chart image version management optimization to ensure packaged Charts inherit the version numbers resolved by the release pipeline, keeping image tags strongly consistent with actual release images.

---

### [kweaver-eval]

0.6.0 ships the first version of the `kweaver-eval` Acceptance module, providing independent acceptance testing for the full KWeaver Core stack. Coverage spans **6 core modules** — **Agent, BKN, VEGA, Data Sources (DS), Dataview, and Context Loader** — with **104 test cases**, of which **79 pass (76%)**.

**1. Test Case Coverage Overview**

| Module | Cases | Passed | Known Issues |
|--------|-------|--------|--------------|
| Agent (Decision Agent) | 33 | 33 | 0 |
| BKN Engine | 26 | 23 | 3 |
| VEGA + DS + Dataview | 27 | 19 | 6 |
| Context Loader | 3 | 3 | 0 |
| Dataflow | 14 | TBD | — |
| Token Refresh | 1 | 1 | 0 |

The Agent module covers CRUD lifecycle, single-turn/multi-turn/streaming conversations, conversational robustness (concurrent sessions, special character input, long messages, cross-turn context retention), and error paths — 33 cases, all passing. The Dataflow module's 14 test cases were activated on the 0.6.0 release date with the `kweaver dataflow` CLI.

**2. Dual-Scoring Framework**

Each test case produces two scoring dimensions:

- **Deterministic Assertions**: validates exit codes, JSON structure, and field values; applied to all cases
- **Agent Judge Semantic Evaluation**: uses the Claude API to score semantic correctness of outputs, supporting CRITICAL / HIGH / MEDIUM / LOW severity classification; selectively enabled per case

**3. Cross-Run Issue Tracking and Upstream Defect Feedback**

A built-in cross-run issue tracking mechanism (`feedback.json` persistence) automatically identifies defects that appear across consecutive runs and escalates them for human follow-up. Acceptance testing has already surfaced and tracked **20+ upstream defects** (including adp#427, adp#428, adp#442, adp#445, adp#447, adp#448, and others), establishing an engineering quality feedback loop between kweaver-eval and upstream repositories.

---

## Release Resources

### GitHub Repositories

**KWeaver Core**
- GitHub Release: https://github.com/kweaver-ai/kweaver-core/tree/release/0.6.0

**KWeaver SDK**
- GitHub: https://github.com/kweaver-ai/kweaver-sdk

**kweaver-eval**
- GitHub: https://github.com/kweaver-ai/kweaver-eval

---
