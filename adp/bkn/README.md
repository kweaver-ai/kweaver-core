# Ontology Engine

[中文](README.zh.md) | English

## Project Overview

The Ontology Engine is a distributed business knowledge network management system developed in Go, providing ontology modeling, data management, and intelligent query capabilities. The system adopts a microservices architecture, divided into ontology management and ontology query modules, supporting the construction, storage, and querying of large-scale knowledge networks.

As a core component of the KWeaver AI platform, the Ontology Engine focuses on building enterprise-level business knowledge networks, enabling business knowledge modeling, storage, querying, and application. The system follows clean architecture principles and SOLID design, offering excellent scalability and maintainability.

### Core Features

#### Ontology Modeling & Management
- **Multi-dimensional Ontology Definition**: Supports complete definition and management of object types, relation types, and action types
- **Branch Management**: Supports branch development and merging of ontology models
- **Visual Configuration**: Provides intuitive visual configuration interface for ontology models

#### Knowledge Network Construction
- **Multi-domain Knowledge Networks**: Supports building cross-domain complex knowledge networks
- **Semantic Relationship Management**: Supports defining and managing complex semantic relationships
- **Topology Analysis**: Provides topology structure analysis and optimization for knowledge networks
- **Auto-sync**: Supports automatic synchronization between ontology models and data sources

#### Intelligent Query Engine
- **Complex Relationship Queries**: Supports multi-hop relationship path queries and subgraph queries
- **Semantic Search**: Intelligent semantic search based on vector similarity
- **Pattern Matching**: Supports intelligent pattern matching queries for graphs
- **Performance Optimization**: High-performance query engine based on OpenSearch

#### Data Integration & Application
- **VEGA Virtualization**: Integrates multiple data sources through the VEGA virtualization engine
- **Data Sync**: Supports real-time synchronization between ontology data and business data
- **Job Scheduling**: Supports complex background job scheduling and execution
- **Permission Management**: Role-based fine-grained permission control

#### System Features
- **Microservices Architecture**: Microservices-based design supporting horizontal scaling
- **High Availability**: Supports distributed deployment and fault recovery
- **Monitoring Integration**: Integrated with OpenTelemetry for full-link monitoring
- **Internationalization Support**: Multi-language support and localization configuration

## System Architecture

### Module Structure

```text
adp/
└── bkn/
    ├── bkn-backend/     # Ontology Management Module
    ├── ontology-query/       # Ontology Query Module
    ├── README.md             # Project documentation
    └── README.zh.md          # Chinese documentation
```

### BKN Manager Module (bkn-backend)

Responsible for creating, editing, and managing bkn models. Main features include:

- **Knowledge Network Management**: Build and manage business knowledge networks
- **Object Type Management**: Define and manage object types in knowledge networks
- **Relation Type Management**: Define and manage relation types in knowledge networks
- **Action Type Management**: Define executable operations and actions
- **Job Scheduling**: Background task and job management

### Ontology Query Module (ontology-query)

Provides efficient knowledge graph query services. Main features include:

- **Model Query**: Query and browse ontology models
- **Graph Query**: Complex relationship path queries
- **Semantic Search**: Semantic-based intelligent search
- **Data Retrieval**: Multi-dimensional data filtering and retrieval

## Quick Start

### Prerequisites

- **Go**: 1.24.0 or higher
- **Database**: MariaDB 11.4+ or DM8 (for data storage)
- **Search Engine**: OpenSearch 2.x (for search and indexing)
- **Dependency Services**: Requires other KWeaver platform services
- **Docker**: Optional, for containerized deployment
- **Kubernetes**: Optional, for cluster deployment

### Local Development

#### 1. Clone the repository

```bash
git clone https://github.com/kweaver-ai/adp.git
cd adp/ontology
```

#### 2. Configure environment

Each module has its own configuration file that needs to be configured according to the actual environment:

```yaml
# BKN Manager configuration
bkn-backend/server/config/bkn-backend-config.yaml

# Ontology Query configuration
ontology-query/server/config/ontology-query-config.yaml
```

**Key configuration items**:
- Database connection information (host, port, user, password)
- OpenSearch connection information
- Dependency service addresses (user-management, data-model, etc.)
- Service port configuration

#### 3. Initialize database

```bash
# Execute database initialization scripts
# MariaDB
mysql -u root -p < bkn-backend/migrations/mariadb/6.0.0/pre/init.sql

# DM8
disql SYSDBA/SYSDBA@localhost:5236 < bkn-backend/migrations/dm8/6.0.0/pre/init.sql
```

#### 4. Run the BKN Manager module

```bash
cd bkn-backend/server
go mod download
go run main.go
```

The service will start at `http://localhost:13014`

**Health Check**:
```bash
curl http://localhost:13014/health
```

#### 5. Run the Ontology Query module

```bash
cd ../ontology-query/server
go mod download
go run main.go
```

The service will start at `http://localhost:13018`

**Health Check**:
```bash
curl http://localhost:13018/health
```

### Docker Deployment

#### Build images

```bash
# Build BKN Manager module
cd bkn-backend
docker build -t bkn-backend:latest -f docker/Dockerfile .

# Build Ontology Query module  
cd ../ontology-query
docker build -t ontology-query:latest -f docker/Dockerfile .
```

#### Run containers

```bash
# Run BKN Manager module
docker run -d -p 13014:13014 --name bkn-backend bkn-backend:latest

# Run Ontology Query module
docker run -d -p 13018:13018 --name ontology-query ontology-query:latest
```

### Kubernetes Deployment

The project provides Helm charts for Kubernetes deployment:

```bash
# Deploy BKN Manager module
helm3 install bkn-backend bkn-backend/helm/bkn-backend/

# Deploy Ontology Query module
helm3 install ontology-query ontology-query/helm/ontology-query/
```

## API Documentation

The system provides complete RESTful API documentation supporting OpenAPI 3.0 specification:

### Ontology Manager APIs

- **Knowledge Network API**: [Knowledge Network API](bkn-backend/api_doc/bkn-backend-network.html)
  - Supports creation, query, update, and deletion of knowledge networks

- **Object Type API**: [Object Type API](bkn-backend/api_doc/bkn-backend-object-type.html)
  - Supports definition and management of object types

- **Relation Type API**: [Relation Type API](bkn-backend/api_doc/bkn-backend-relation-type.html)
  - Supports definition and management of relation types
  - Provides directionality and multiplicity configuration

- **Action Type API**: [Action Type API](bkn-backend/api_doc/bkn-backend-action-type.html)
  - Supports definition and management of action types

- **Job Management API**: [Job Management API](bkn-backend/api_doc/bkn-backend-job-api.html)
  - Supports creation and scheduling of background jobs

### Ontology Query APIs

- **Query Service API**: [Query Service API](ontology-query/api_doc/ontology-query.html)
  - Supports complex relationship path queries
  - Provides semantic search and pattern matching
  - Supports multi-dimensional data filtering and retrieval
  - Provides high-performance pagination queries

### API Access Methods

**Local Development Environment**:
```
# BKN Manager API
http://localhost:13014/api/bkn-backend/v1/

# Ontology Query API
http://localhost:13018/api/ontology-query/v1/
```

**Production Environment**:
```
# BKN Manager API
https://your-domain.com/api/bkn-backend/v1/

# Ontology Query API
https://your-domain.com/api/ontology-query/v1/
```

## Database Support

The system supports multiple databases:

- **MariaDB**: Primary data storage
- **DM8**: DM8 database support
- **OpenSearch**: Search engine and data analysis

Database migration scripts are located at:

- `bkn-backend/migrations/`
- `ontology-query/migrations/`

## Monitoring & Logging

- **Logging System**: Integrated structured logging with multi-level log recording
- **Distributed Tracing**: OpenTelemetry-based distributed tracing
- **Health Checks**: Health check endpoints provided

## Development Guide

### Code Structure

The project follows clean architecture principles and SOLID design, with a clear and maintainable code structure:

```text
server/
├── common/              # Common configuration, utility functions, and constants
├── config/              # Configuration files and configuration loading
├── drivenadapters/      # Data access layer (driven adapters)
├── driveradapters/      # Interface adapter layer (driver adapters)
├── errors/              # Error definitions and handling mechanisms
├── interfaces/          # Interface definitions and abstractions
├── locale/              # Internationalization support and multi-language configuration
├── logics/              # Business logic layer
├── main.go              # Application entry point and startup logic
├── version/             # Version information and build details
└── worker/              # Background tasks and job executors
```

### Development Standards

1. **Clean Architecture**: Follow clean architecture principles, separating concerns
2. **Interface Isolation**: Clearly define interfaces and implementations, depend on abstractions rather than concrete implementations
3. **Error Handling**: Unified error handling mechanism using custom error types
4. **Logging Standards**: Structured logging using Zap logging library
5. **Test Coverage**: Unit tests and integration tests, aiming for high test coverage
6. **Code Style**: Follow Go official code style, format code with go fmt
7. **Comment Standards**: Clear comments including function comments and complex logic explanations
8. **Version Control**: Follow Git Flow workflow, use semantic versioning

### Testing Guide

```bash
# Run unit tests
cd bkn-backend/server
go test ./... -v

# Run integration tests
cd bkn-backend/server
go test ./... -v -tags=integration

# Generate test coverage report
cd bkn-backend/server
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Contributing

Refer to the KWeaver project content

## Version History

- **v0.1.0**: Current version, based on Go 1.24

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](../../LICENSE.txt) file for details.

## Support & Contact

- **Technical Support**: AISHU ADP R&D Team
- **Documentation Updates**: Continuously updated
- **Issue Reporting**: Submit through internal system

---

**Note**: This is an enterprise-level internal project. Code and documentation may contain specific business logic and configurations. Please adjust according to your actual environment.
