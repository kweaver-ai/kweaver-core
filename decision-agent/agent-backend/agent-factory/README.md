# Agent Factory

English | [中文](README.zh.md)

Agent Factory is a Go service in Decision Agent responsible for Agent configuration, publishing, API exposure, and runtime orchestration metadata management. This document focuses on the current API Chat integration guide, as well as integration instructions related to OpenAPI / DIP pages.

## Quick Access

- External API documentation: [docs/api/README.md](./docs/api/README.md)
- Static OpenAPI Scalar HTML: [docs/api/agent-factory.html](./docs/api/agent-factory.html)
- Static OpenAPI Redoc HTML: [docs/api/agent-factory-redoc.html](./docs/api/agent-factory-redoc.html)
- Static OpenAPI JSON: [docs/api/agent-factory.json](./docs/api/agent-factory.json)
- Static OpenAPI YAML: [docs/api/agent-factory.yaml](./docs/api/agent-factory.yaml)
- API Chat entry code: [src/driveradapter/api/httphandler/agenthandler/api_chat.go](./src/driveradapter/api/httphandler/agenthandler/api_chat.go)
- Unified Chat service: [src/domain/service/agentrunsvc/chat.go](./src/domain/service/agentrunsvc/chat.go)

Common commands:

```bash
make gen-api-docs
make validate-api-docs
make view-swag
make view-redoc
make goTest
make ciLint
make ciLintFix
make fmt
```

## Project Structure

Agent Factory is a Go service adopting DDD + Hexagonal Architecture.

Core directories:
- `main.go`
  - Process entry point; initializes OpenTelemetry and starts HTTP service
- `conf/`
  - Runtime configuration and local example configuration, including database / Redis / port / Swagger switches, etc.
- `src/boot/`
  - Startup initialization logic for configuration, DB, Redis, logging, permissions, auditing, etc.
- `src/domain/`
  - Domain layer: entities, services, aggregates, value objects, enums, and core business logic
- `src/port/`
  - Driver / driven interface definitions, used to keep domain logic decoupled from adapters
- `src/drivenadapter/`
  - Infrastructure adapters, such as database access, external HTTP access
- `src/driveradapter/api/`
  - HTTP handlers and request / response DTOs
- `src/infra/server/httpserver/`
  - HTTP service assembly, route registration, middleware mounting, and graceful shutdown lifecycle
- `cmd/openapi-docs/` and `internal/openapidoc/`
  - OpenAPI generation, comparison, validation, and static document rendering
- `docs/`
  - Integration documentation, screenshots, and external OpenAPI artifacts

From a routing perspective, the service mainly has two types of interfaces:
- Management side interfaces
  - Mainly V3 handlers, responsible for agent configuration, publishing, marketplace, products, categories, etc.
- Run side interfaces
  - Mainly V1 handlers, responsible for chat, API Chat, conversation, session, and other runtime capabilities

## How to Run the Project

Agent Factory actually has two practical ways to run.

### Running in the Project Directory

Minimal direct run process:

1. Prepare local configuration
   - Start from `conf/agent-factory.example.yaml`
   - Focus on:
     - `project.port` (example config uses `30777`)
     - `db.*`
     - `redis.*`
     - `agent_executor.*`
     - `enable_swagger`
     - Local development switches, such as `auth_enable`, `switch_fields.disable_biz_domain`, and mock-related configurations
2. Ensure dependencies are reachable
   - MySQL
   - Redis
   - Agent Executor
3. Start the service

Direct startup command can be:

```bash
go run ./
```

Since `src/boot/init.go` will initialize DB, Redis, logging, permissions, business domains, and optional auditing at startup, the direct run method will only succeed when related configurations and dependencies are ready. Some local development switches (such as `auth_enable`, `switch_fields.*`, etc.) can help you skip certain dependencies.

## Development and Validation

Code quality:

```bash
make fmt
make ciLint
make ciLintFix
```

Testing:

```bash
make goTest
make ut
make utExclude
```

OpenAPI documentation:

```bash
make gen-api-docs
make validate-api-docs
make compare-api-docs
```

Additional notes:
- The Swagger host declared in `main.go` is `localhost:30777`, and the example port in `conf/agent-factory.example.yaml` is also `30777`
- Runtime Scalar UI entry: `http://127.0.0.1:30777/swagger/index.html`
- Runtime Redoc UI entry: `http://127.0.0.1:30777/redoc/index.html`
- Run side public interface base path is `/api/agent-factory/v1`
- Run side internal interface base path is `/api/agent-factory/internal/v1`

## API Chat Integration

For complete API Chat integration instructions, please check the documents in the repository `docs/` root directory:

- English: [docs/api-chat-integration.md](./docs/api-chat-integration.md)
- 中文: [docs/api-chat-integration.zh.md](./docs/api-chat-integration.zh.md)

This document includes:
- Differences between `agent_id` / `agent_key` / `app_key`
- Routing differences between regular Chat and API Chat
- Real page request examples
- Request / response structure explanations
- Complete snapshot stream and incremental stream explanations
- Process state / tool call structure explanations
- API documentation entry points
