# Context Loader

中文文档请见 [README-zh.md](README-zh.md)。

## Module Overview

Agent Retrieval is the retrieval service component of the context-loader in ADP. It provides knowledge-network semantic search, action recall, ontology query proxying, logic property resolution, and MCP (Model Context Protocol) server capabilities. It integrates with ontology-query, ontology-manager, operator integration, and other services to support agent-driven knowledge retrieval and tool invocation.

## Core Features

Core features are six tools and the Context Loader MCP server:

1. **kn_search**: Knowledge network search (concept- and instance-level semantic retrieval)
2. **kn_schema_search**: Knowledge network schema search
3. **query_object_instance**: Query object-type instances in a knowledge network
4. **query_instance_subgraph**: Query subgraph around instances with configurable depth and relations
5. **get_action_info**: Get action metadata and schema for operator/agent tooling
6. **get_logic_properties_values**: Resolve logic properties and get property values
7. **Context Loader MCP Server**: Exposes the six tools above to MCP clients (e.g. Cursor, Claude Desktop)

## Technical Architecture

### Tech Stack
- **Language**: Go 1.24
- **Web framework**: Gin 1.11
- **Observability**: OpenTelemetry
- **Cache**: Redis (optional, sentinel mode supported)
- **Validation**: go-playground/validator

### Project Structure
```
server/
├── driveradapters/     # HTTP/MCP adapters
│   ├── knactionrecall/
│   ├── knlogicpropertyresolver/
│   ├── knontologyjob/
│   ├── knqueryobjectinstance/
│   ├── knquerysubgraph/
│   ├── knretrieval/
│   ├── knsearch/
│   ├── mcp/            # MCP server & schemas
│   └── mcpproxy/
├── drivenadapters/     # Data/backend clients
│   ├── ontology_query.go
│   ├── ontology_manager.go
│   ├── operator_integration.go
│   ├── data_retrieval.go
│   ├── agent_app.go
│   └── ...
├── interfaces/         # Port interfaces
├── logics/             # Business logic
│   ├── knactionrecall/
│   ├── knlogicpropertyresolver/
│   ├── knquerysubgraph/
│   ├── knrerank/
│   ├── knretrieval/
│   └── knsearch/
├── infra/              # Config, logger, telemetry, i18n
│   ├── config/
│   ├── logger/
│   ├── telemetry/
│   └── localize/
└── main.go
```

## Quick Start

### Requirements
- Go 1.24+
- Access to ontology-query, ontology-manager, and other configured backends (or use ktctl to connect to a remote cluster)

### Local Development

1. **Clone**
   ```bash
   git clone <repo-url>
   cd context-loader/agent-retrieval
   ```

2. **Config**
   Edit `server/infra/config/agent-retrieval.yaml` (and optionally `agent-retrieval-secret.yaml`) with service hosts and ports. For local runs against a remote cluster, use ktctl and then `make build` so configs are copied to `/sysvol/config` and `/sysvol/secret`.

3. **Dependencies**
   ```bash
   go mod download
   ```

4. **Run**
   ```bash
   # Option A: with ktctl and make (recommended for dev)
   make build

   # Option B: run from server dir (config from default paths)
   cd server && go run main.go
   ```
   Service listens on `http://0.0.0.0:30779` by default.

### Make Targets
- `make help` – List targets
- `make lint` – Run golangci-lint
- `make test` – Unit tests
- `make test-cover` – Tests with coverage
- `make generate-mock` – Generate mocks
- `make helm-template` – Render Helm templates
- `make preview` – Preview API docs (from `docs/`)

### Docker

Build from the **repository root** (context-loader), so that the Dockerfile can copy the agent-retrieval subtree:

```bash
# From context-loader root
docker build -t agent-retrieval:latest -f agent-retrieval/docker/Dockerfile .
```

Run:

```bash
docker run -d -p 30779:30779 --name agent-retrieval agent-retrieval:latest
```

### Kubernetes / Helm

From `agent-retrieval` directory:

```bash
helm install agent-retrieval ./helm/agent-retrieval/
# Or render only:
helm template agent-retrieval ./helm/agent-retrieval -n <namespace> -f ./helm/agent-retrieval/values.yaml
```

## Configuration

Main config file: `server/infra/config/agent-retrieval.yaml`.

Key sections:
- **project**: `host`, `port` (default 30779), `language`, `logger_level`, `debug`
- **ontology_query**, **ontology_manager**, **data_retrieval**, **operator_integration**: backend service URLs/ports
- **oauth**: Hydra OAuth (for public API)
- **redis**: optional Redis (e.g. sentinel) for cache
- **concept_search_config**, **deploy_agent**, **rerank_llm**: feature-specific settings

Secrets and sensitive values: `server/infra/config/agent-retrieval-secret.yaml` (excluded from version control as needed).

## Monitoring and Operations

- **Ready**: `GET /health/ready`
- **Liveness**: `GET /health/alive`
- **Logging**: Structured logging; level and output configurable via config.
- **Tracing**: OpenTelemetry; enable and configure in `observability` config.

## Development

- Follow Go standard style and project conventions.
- Clean architecture: keep driver/driven adapters, interfaces, and logics clearly separated.
- Run `make lint` and `make test` (or `make test-cover`) before submitting.

## Version

See `VERSION` in the project root.

## Support

- **Team**: AISHU ADP
- **Docs**: See `docs/` (PRD, APIs, releases).
