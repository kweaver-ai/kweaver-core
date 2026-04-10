# Changelog
## 0.6.0

### Features & Improvements

- Add skill type support in `agent-factory`
  - Add `SkillSkill` value object with validation
  - Support skill configuration in agent create/detail/update handlers and run service
- Add TraceAI evidence header support in `agent-executor`
  - Introduce `enable_traceai_evidence` feature flag in `FeaturesConfig`
  - Inject `X-TraceAi-Enable-Evidence` header into API tool proxy requests
- Add v0.6.0 database migrations: skill-related tables and `agent-memory` history table (dm8 and mariadb)

### Refactoring & Cleanup

- Refactor publish request validation in `agent-factory`: use constructor semantics, validate and trim request fields, return 400 with explicit error messages instead of 500
- Reorganize database migration file structure, removing redundant `pre` directory paths
- Remove unused `ENABLE_EVIDENCE_EXTRACTION` env var and dict copy from deployment; update `agent-executor` dependencies and optimize Dockerfile

## 0.5.3

### Refactoring & Cleanup

- Simplify API chat route in `agent-factory`: get `agent_key` from request body instead of URL path, update route to `/api/agent-factory/v1/api/chat/completion`, and sync Swagger docs and `agent-web` API document components
- Add automated OpenAPI 3 doc generation from handler code annotations in `agent-factory`, with Scalar-based runtime doc server (`/scalar`), dynamic server URL rewriting, and embedded static assets (JSON, YAML, HTML, favicon)

### Testing

- Add unit tests for API chat handler and API doc resume-chat flow

## 0.5.2

### Bug Fixes

- Fix logger method calls in agent-executor dolphin paths
  - Replace `warning` with `warn` when dolphin trace listener creation fails
  - Replace `warning` with `warn` when run-agent options fall back to auto-generated conversation id

## 0.5.1

### Bug Fixes

- Fix long-term memory prompt assembly to prevent errors when memory retrieval is enabled
  - Update memory injection to append retrieved memories into `_history` as a system message instead of mutating the wrong history variable
  - Serialize retrieved memory results before injection so downstream prompt rendering can consume them consistently

### Observability

- Improve OpenTelemetry failure diagnostics in `agent-executor`
  - Replace direct stdout prints in trace wrapper with structured logger output
  - Include full traceback details when trace provider initialization fails

## 0.5.0

### Features & Improvements

- Add OpenTelemetry upgrades across `agent-factory` and `agent-executor`
  - Initialize the new OTel pipeline in `agent-factory` and expose service name, version, environment, OTLP endpoint, trace sampling, and log level through configuration
  - Standardize GenAI trace attributes for agent ID, user ID, conversation ID, and operation name so request spans and downstream spans can be correlated consistently
  - Update `agent-executor` trace configuration to use the unified OTLP endpoint and environment-based sampling settings
- Add evidence extraction support for API tools in `agent-executor`
  - Allow API tool responses to extract `nodes` data into `_evidence` payloads
  - Add the `ENABLE_EVIDENCE_EXTRACTION` switch to control the behavior at runtime
- Expand deployment and operations controls
  - Add Helm and config support for sandbox platform toggles and improved business-domain toggle rendering
  - Add permission-check bypass support for internal agent-app middleware when `disable_pms_check` is enabled

### Documentation

- Add generated API documentation for `agent-factory` and `agent-executor`
- Add Helm component metadata for the new OTel configuration surface

### Frontend (agent-web)

- Bug fix: support list-type interrupt parameters during conversation interruption.

## 0.4.4

### Frontend (agent-web)

- Bug fix: fix the issue where knowledge entries could not be added in Agent configuration.
- Bug fix: fix the issue where metrics could not be added in Agent configuration.

## 0.4.3

### Features & Improvements

- Add DisableBizDomain feature
  - Add DisableBizDomain configuration option in SwitchFields
  - Implement IsBizDomainDisabled() helper method
  - Update agent config services to support business domain disable
  - Modify personal space and published services accordingly
  - Update OAuth middleware to handle disabled business domain
  - Add configuration examples and documentation updates
  - Deploy Helm chart configuration updates

### Testing

- Add comprehensive unit test coverage for disable business domain functionality
  - Add disable_biz_domain_test.go test file
  - Update test cases for related services
  - Enhance middleware test coverage

### Bug Fixes
- Bug fix: Gracefully handle unavailable tools during skill initialization instead of failing the entire `dolphin_run` request
  - Update `get_tool_info` to log tool availability errors and return an empty result instead of raising an exception
  - Remove unavailable tools from `skills.tools` so remaining tools can continue to load and execute
  - Set the sandbox execute-sync OpenAPI `session_id` path parameter default to `sess-agent-default`
  - Add unit tests covering unavailable tool filtering and the downgraded tool info error path

- Bug fix: Fix the issue where conversation status is not updated to failed when agent-executor process is killed or other exceptions occur
  - Modify chat_process.go file to set messageChanClosed = true when errChan is closed or receives EOF error
  - Ensure conversation status is correctly updated to failed when agent-executor process is killed or other exceptions occur
  - Avoid UI showing status as running and error "conversation_id not found" when reopening historical conversation

- Bug fix: Fix the issue where assistant message content is empty after user stops the conversation
  - Modify HandleStopChan function to update session's temporary message content to database when user clicks stop button
  - Ensure stopped conversation content can be displayed normally in conversation list

## 0.4.2

### Frontend (agent-web)

- Bug fix: remove legacy business knowledge network checks from the Agent configuration page.
- Bug fix: remove document-related checks from the Agent configuration page.

## 0.4.1

### Features & Improvements

- Add conversation history configuration in `agent-factory`
  - Support configuring conversation history retention strategy via API
  - Implement count-based history control (default strategy)
  - Reserve configuration options for time window and token usage strategies
  - Support no-history conversation mode
  - Add `conversationHistoryConfig` structure and related enum types
  - Add history count limit constant (range 1-1000)

### Bug Fixes

- Remove duplicate parameter validation: caller already validates `historyLimit` parameter, subsequent functions no longer repeat validation
- Correct count limit range description: updated to 1-1000

### Testing

- Add unit tests for conversation history configuration validation logic
- Add test cases for conversation history retrieval edge cases

### Documentation

- Add conversation history strategy configuration design document

## 0.4.0

### Features & Improvements

- Enhance log directory creation with proper permission handling in `agent-executor`
  - Implement directory permission handling with 0o755 mode
  - Add write permission check for LOG_DIR with appropriate error handling
  - Add unit tests for new permission checking logic
- Fix missing agent router v1 registration in `agent-executor`
  - Import and register agent_router_v1 alongside existing v2 router
  - Restore complete API routing functionality for v1 endpoints

## 0.3.6

### Features & Improvements

- Add agent lookup by ID or key in `agent-factory` and update related chat / API doc flows
- Increase API tool timeout buffer in `agent-executor` to improve long-running request stability

### Bug Fixes

- Fix permission checks for agent-key based access by resolving the actual agent ID before validation
- Update sandbox session ID to use the default `sess-agent-default` format required by code execution

### Refactoring & Cleanup

- Update square service interfaces and related service logic to reuse the new ID-or-key lookup path
- Improve `api_doc.go` step annotations for clearer code flow

### Testing

- Add unit tests for API tool timeout handling and agent lookup by ID or key
- Update related `agent-factory` service tests to cover the new lookup behavior and permission flow

### Chores

- Bump project version to 0.3.6

## 0.3.5

### Features & Improvements

- Add CI/CD workflows and code quality improvements for decision-agent
- Add pre-commit hooks and optimize code formatting
- Enhance docker-compose deployment with configurable ports and automation
- Add comprehensive unit tests for domain services and infrastructure
- Add comprehensive unit tests for agent-factory
- Add configuration file and update gitignore for agent-executor

### Bug Fixes

- Fix EnsureSandboxSession to delete failed sessions before recreating
- Clear InterruptInfo from Ext when resuming interrupted chat
- Fix route matching errors when agent-web runs independently
- Fix static resource loading errors when agent-web runs independently

### Refactoring & Cleanup

- Fix temporary zone session ID and remove old file upload component in agent-web
- Remove unused imports and optimize code structure in agent-executor
- Remove deprecated space-related modules
- Refactor unit tests to avoid environment variable race conditions
- Rename helm templates and remove unused configmap
- Replace agent-factory HTTP access layer with internal service calls
- Improve env loading and add air hot-reload support
- Remove unused DelInternationalPath function
- Refactor agent-web Dockerfile to support Docker Compose

### Testing

- Add comprehensive unit tests for DTO and Infrastructure layers
- Add comprehensive unit tests for infrastructure and handler layers
- Add comprehensive unit tests for domain services and value objects
- Add unit tests for agentrunsvc and related services
- Add unit tests for agentconfigsvc, agentrunsvc and releasesvc
- Add comprehensive test cases for agent import/export functionality
- Add unit tests for tool_v2 API tool package modules
- Add comprehensive test coverage for agent factory configuration
- Add tests for dependencies modules and boot module
- Add tests for exception handlers and router middleware modules
- Expand tests for tool_requests and json modules

### Documentation

- Add code quality guidelines and linting instructions
- Add comprehensive documentation for Claude Code guidance (CLAUDE.md)

### Chores

- Add pre-commit configuration and update development workflow for agent-executor
- Add coverage configuration to exclude test and build files
- Update configuration files for production deployment
- Update PyInstaller spec with setuptools dependencies

## 0.3.4

### Bug Fixes

- Fix Agent conversation error when memory function is enabled
- Add dbutilsx dependency and refactor memory configuration parsing
- Update numpy minimum version to 1.23.5 for compatibility
- Fix memory retrieval reranking error when relevance_score is None

## 0.3.3

### Frontend (agent-web)

- Fix filter status abnormal bug on My Agent template page

## 0.3.2

### Bug Fixes

- Fix known issues

## 0.3.1

### Frontend (agent-web)

- Add hide support for temporary area in Agent conversation interface
- Fix temporary area display issue in Agent conversation interface
- Add i18n support for "Decision Agent" text across pages
- Fix missing x-business-domain header in Agent API page debug requests
- Remove unused KNSpaceTree, DocTree, ContentDataTree components

## 0.3.0

### Features & Improvements

- Add Sandbox Platform integration for code execution and file management
- Add PyCodeGenerate agent with sandbox execution tools
- Add Swagger documentation support for agent-factory
- Add SelectedFiles field to debug endpoint
- Add development control switches and new configuration options for Helm Chart
- Update kweaver-dolphin dependency to v0.2.4

### Bug Fixes

- Fix tool execution status display error when user skips interrupted tool steps
- Fix interrupted JSON value type rendering error
- Fix file display issue in debug mode
- Fix debug mode file parameter passing

### Refactoring & Cleanup

- Remove EcoIndex and DataHubCentral configurations
- Remove deprecated doc_qa and graph_qa tools
- Remove pandas dependency from agent-executor
- Remove deprecated service classes and prompt utilities
- Remove unused data access objects and old dataset tables
- Simplify configuration and remove unused stop words file config
- Hide Agent Trajectory Analysis Module
- Remove batch-check-index-status interface call in frontend

### Testing

- Add comprehensive unit tests for agent-memory module
- Add unit tests for config loader and utility functions

### Frontend (agent-web)

- Support skipped status display for interrupted tool execution
- Adapt Agent conversation for sandbox file upload
- Remove file-related configurations from Agent config interface
- Remove file deletion and preview functions from debug mode

## 0.2.3

### Bug Fixes

- Fix Agent white screen issue when running in role instruction mode

## 0.2.2

### Bug Fixes

- Fix agent interrupt parameter passing in frontend
- Fix conversation interface white screen issue
- Fix configuration type dropdown selection failure
- Fix template creation agent 404 error
- Fix agent-memory permission error and improve observability

### Features & Improvements

- Add tool interrupt resume support via unified Run API
- Make TelemetrySDK optional dependency in agent-executor
- Optimize message extension structure and add status handling
- Simplify interrupt handling and type conversion
- Optimize chat resume with unified DTO types and interrupt recovery

### Frontend (agent-web)

- Support standalone operation without micro-frontend
- Streamline interrupt chat interface to only pass changed parameters
- Remove redundant changelog files

## 0.2.1

### Bug Fixes

- Fix agent-web installation blocking issue
- Fix agent retrieval functionality (#37, #38)

### Infrastructure

- Rename Helm Chart from agent-factory to agent-backend
- Remove compiled artifacts from tests/tools to reduce repository size

### Documentation

- Update changelog for recent changes

## 0.2.0

### Architecture & Deployment

- Unified multi-service Docker architecture with supervisor process management
- Helm Chart configuration fixes for agent-factory deployment
- Add missing service configurations (agent_executor, efast, docset, ecoconfig, uniquery)
- Fix volumeMounts to use subPath for precise file mounting
- Update securityContext runAsUser/runAsGroup to 1001
- Enable GOPROXY support for Docker build optimization
- Enable mq-sdk and telemetrysdk-python dependencies

### Agent Interrupt & Resume

- Add agent interrupt and resume functionality
- Custom ToolInterruptException for tool interrupt handling
- Fix progress handling for interrupted sessions
- Frontend adaptation for interrupt operations

### Agent Executor

- Move agent-executor module to agent-backend directory
- Add backward compatibility aliases for PascalCase function names
- Fix parameter handling in memory handler
- Refactor tool interrupt handling and DTO naming

### Agent Factory

- Add agent-factory-v2 complete implementation with DDD architecture
- Restructure httpserver module with legacy path configuration support
- Add streaming response logging and improve request logging
- Enable keep_legacy_app_path configuration

### Frontend (agent-web)

- Agent streaming API supports agent_run_id parameter
- Tool configuration with confirmation prompt support
- Fix MCP tree node expansion bug when adding skills
- Fix YAML syntax errors in deployment files
- Menu registration updates

### Code Quality & Refactoring

- Remove agent-go-common-pkg external dependency
- Migrate DolphinLanguageSDK imports to new dolphin package structure
- Remove deprecated function error classes
- Simplify Dockerfile with unified copy command
- Add opencode workflow for automated code review
- Remove compiled artifacts from tests/tools/fetch-log/build to reduce repository size
- Update .gitignore to exclude build artifacts and log files

### Data Retrieval

- Add Jupyter Gateway runner for code execution
- Add code runner utilities (exec_runner, ipython_runner)
- Enhance DIP services integration
- Add MCP test utilities and examples
- Add text-to-DIP metric tools and prompts
