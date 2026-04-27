# Flow Automation

[中文](README.zh.md) | English

Flow Automation is a core service responsible for orchestrating and executing automated data pipelines. It manages the complete lifecycle of data flow executions, integrates with various infrastructure components, and provides comprehensive APIs for data pipeline management.

## Core Architecture

This project follows a Hexagonal Architecture (Ports and Adapters), separating core business logic from external interfaces:

```
┌─────────────────────────────────────────────────────────┐
│                    Driver Adapters                      │
│          (HTTP API, Message Queue Consumers)            │
│  ┌───────────────────────────────────────────────────┐  │
│  │              Core Business Logic                  │  │
│  │         (logics/, module/, pkg/)                  │  │
│  └───────────────────────────────────────────────────┘  │
│                   Driven Adapters                       │
│        (Database, External Services, Cache)             │
└─────────────────────────────────────────────────────────┘
```

### Directory Structure

- **`logics/`**: Core business logic layer, containing data pipeline management, executors, triggers, etc.
- **`driveradapters/`**: Driver adapters (inbound), handling HTTP requests, message queue consumption, etc.
- **`drivenadapters/`**: Driven adapters (outbound), interacting with databases, external services, cache, etc.
- **`module/`**: Domain modules, including initialization and configuration
- **`pkg/`**: Shared utility packages and entity definitions

## Key Features

### 1. Data Pipeline Management (DAG Management)
- **Create & Edit**: Support for creating, updating, and deleting data pipelines (DAGs)
- **Version Control**: Data pipeline version management
- **Execution Management**: Start, cancel, and pause pipeline instances
- **Debug Mode**: Support for single-step and full debugging
- **Batch Operations**: Batch query and operations on data pipelines

### 2. Custom Executors
- **Executor Management**: Create, update, and delete custom executors
- **Action Management**: Add, modify, and delete actions for executors
- **Access Control**: User permission-based executor access control
- **Agent Import**: Support for importing agents from app store

### 3. Trigger System
- **Scheduled Triggers**: Cron-based scheduled task triggers
- **Event Triggers**: Support for external event-driven data pipeline execution
- **Manual Triggers**: Support for user-initiated data pipeline execution
- **Callback Handling**: Process asynchronous task callbacks

### 4. Authentication & Authorization
- **Token Authentication**: OAuth2 authentication via Hydra
- **Permission Verification**: Fine-grained resource access control
- **Business Domain Isolation**: Multi-tenant business domain isolation
- **Minimal Dependency Mode**: Supports disabling ISF dependencies through `auth.enabled` / `SERVER_AUTH_ENABLED`; in this mode the service can run without Hydra, Authentication, Authorization, and User Management, but permission and business-domain isolation are bypassed

### 5. Security Policies
- **Access Control**: API access security policies
- **Data Isolation**: Tenant data isolation
- **Audit Logging**: Operation audit trails

### 6. Observability
- **Metrics Monitoring**: Prometheus metrics exposure
- **Distributed Tracing**: OpenTelemetry distributed tracing
- **Log Management**: Structured logging and streaming logs
- **Health Checks**: Service health status monitoring

### 7. Data Flow Management
- **Data Connections**: Manage data source connection configurations
- **Data Transformation**: Support for data flow transformation and processing

### 8. Operator Management
- **Operator Registration**: Register and manage data processing operators
- **Operator Execution**: Execute various data processing operators

## Tech Stack

- **Language**: Go 1.24.0
- **Web Framework**: Gin
- **Databases**: 
  - MongoDB (data pipeline definitions and instance storage)
  - MariaDB/MySQL (relational data)
  - Redis (caching and distributed locks)
- **Message Queue**: Kafka
- **Configuration**: Viper, godotenv
- **Observability**: OpenTelemetry, Prometheus
- **Container Orchestration**: Kubernetes (Helm Charts)

## Prerequisites

- Go 1.24+
- Docker & Docker Compose (for local infrastructure)

## Getting Started

For detailed local development setup instructions, including running infrastructure and mock services, please refer to [LOCAL_DEVELOPMENT.md](LOCAL_DEVELOPMENT.md).

### Quick Run

1. **Start Dependencies**:
   ```bash
   docker-compose up -d
   ```

2. **Run Application**:
   ```bash
   go run main.go
   ```

The application will start two services:
- Public API Server: `http://localhost:8082`
- Private API Server: `http://localhost:8083`

## Configuration

The application uses environment variables for configuration:

- **`.env`**: Default configuration (committed to Git)
- **`.env.local`**: Local development overrides (gitignored)

Key configuration items:
- `API_SERVER_PORT`: Public API port (default 8082)
- `API_SERVER_PRIVATE_PORT`: Private API port (default 8083)
- `DB_TYPE` (`dbType`/`db_type`): Selects the workflow metadata storage backend (`mongodb` or `mysql`).
  - `mongodb`: Uses MongoDB to store DAG definitions, execution instances, and other document-style workflow data.
  - `mysql`: Uses the MySQL/MariaDB-based repository implementation for workflow metadata after relational storage deployment or migration.
    If the service previously ran with MongoDB storage, migrate the historical workflow data before switching to `mysql`, for example by running `python3 migrations/mariadb/0.4.0/pre/02-mongodb_to_mysql_migration.py`.
- `SERVER_AUTH_ENABLED`: Controls whether auth-related ISF dependencies are enabled for `flow-automation`.
  - `true` (default): Uses real Hydra, Authentication, Authorization, and User Management services.
  - `false`: Uses built-in mock adapters so the service can run in a minimal-dependency mode without ISF.
  - In Helm deployments, this is managed by `auth.enabled`.
  - This mode is intended for local development, demos, or minimal-dependency deployments, and it bypasses the original permission and business-domain isolation behavior.
- `MONGODB_HOST`: MongoDB address
- `REDIS_HOST`: Redis address
- `KAFKA_BROKERS`: Kafka cluster addresses

For detailed design and scope of this change, see [#273 Dataflow decoupling ISF](docs/features/%23273%20Dataflow%E8%A7%A3%E8%80%A6ISF/Dataflow%E8%A7%A3%E8%80%A6ISF.md).

## Project Structure

```text
.
├── common/             # Common utilities and constants
├── conf/               # Configuration files
├── docs/               # Documentation
├── drivenadapters/     # Outbound adapters (DB, external APIs)
├── driveradapters/     # Inbound adapters (HTTP, Event Consumers)
│   ├── admin/          # Admin interfaces
│   ├── alarm/          # Alarm interfaces
│   ├── auth/           # Authentication interfaces
│   ├── executor/       # Executor interfaces
│   ├── mgnt/           # Data pipeline management interfaces
│   ├── trigger/        # Trigger interfaces
│   └── ...
├── helm/               # Helm deployment charts
├── logics/             # Business logic
│   ├── mgnt/           # Data pipeline management logic
│   ├── executor/       # Executor logic
│   ├── cronjob/        # Scheduled task logic
│   └── ...
├── mock-server/        # Mock services for local development
├── module/             # Domain modules
├── pkg/                # Shared packages
│   ├── actions/        # Action definitions
│   ├── entity/         # Entity definitions
│   └── ...
├── resource/           # Static resources
├── schema/             # Database schemas
├── scripts/            # Helper scripts
├── store/              # Data storage interfaces
└── utils/              # General utilities
```

## API Documentation

Main API endpoints:

### Data Pipeline Management
- `POST /api/automation/v1/dags` - Create data pipeline
- `PUT /api/automation/v1/dags/:id` - Update data pipeline
- `GET /api/automation/v1/dags/:id` - Get data pipeline details
- `DELETE /api/automation/v1/dags/:id` - Delete data pipeline
- `GET /api/automation/v1/dags` - List data pipelines

### Data Pipeline Execution
- `POST /api/automation/v1/dags/:id/run` - Execute data pipeline
- `POST /api/automation/v1/dags/:id/cancel` - Cancel execution
- `GET /api/automation/v1/dags/:id/instances` - Query execution instances

### Executor Management
- `POST /api/automation/v1/executors` - Create executor
- `PUT /api/automation/v1/executors/:id` - Update executor
- `GET /api/automation/v1/executors` - List executors

## Deployment

The project uses Helm for Kubernetes deployment:

```bash
cd helm/flow-automation
helm install flow-automation . -f values.yaml
```

## Development Guide

1. **Code Style**: Follow Go standard code conventions, use `golangci-lint` for code checking
2. **Testing**: Use `go test` to run unit tests
3. **Mock Generation**: Use `go.uber.org/mock` to generate test mocks
