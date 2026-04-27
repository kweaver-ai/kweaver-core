# Dataflow

Dataflow is a comprehensive data processing platform that enables users to build, orchestrate, and execute automated data pipelines through visual pipeline design, code execution, and data transformation capabilities.

[中文文档](README_zh.md)

## Overview

Dataflow provides a complete solution for enterprise data processing needs. Whether you need to process large volumes of data, transform and analyze information, or integrate complex data sources, Dataflow provides the tools and services to accomplish your goals efficiently.

## Core Capabilities

Build and execute automated data pipelines with visual pipeline design, code execution, and data transformation capabilities.

**Key Features:**
- Visual pipeline designer for data flows
- Sandboxed Python code execution
- Data transformation and analysis
- Document processing (Word, Excel, PDF)
- OCR and text extraction
- Scheduled and event-driven execution
- Real-time data streaming

**Use Cases:**
- ETL (Extract, Transform, Load) pipelines
- Data quality validation and cleansing
- Automated report generation
- Document processing and analysis
- Image and text recognition pipelines

## Architecture

Dataflow is built as a microservices architecture with the following components:

```
┌─────────────────────────────────────────────────────────┐
│                  Frontend Layer                         │
│  - dia-flow-web: Data flow visual designer              │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│               Data Processing Services                  │
│  - flow-automation: Data flow orchestration             │
│  - coderunner: Sandboxed code execution                 │
│  - flow-stream-data-pipeline: Real-time streaming       │
│  - ecron: Scheduled task management                     │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│                 Shared Libraries                        │
│  - ide-go-lib: Common Go libraries                      │
└─────────────────────────────────────────────────────────┘
```

## Services Overview

### Data Processing Services

#### [flow-automation](flow-automation/)
Core data flow orchestration service that manages the complete lifecycle of data pipeline executions.
- **Language**: Go
- **Framework**: Gin
- **Key Features**: DAG management, executor management, trigger system, data connections

#### [coderunner](coderunner/)
Sandboxed Python code execution service for running custom data processing logic.
- **Language**: Python 3.9
- **Key Features**: RestrictedPython execution, package management, document processing, OCR

#### [flow-stream-data-pipeline](flow-stream-data-pipeline/)
Real-time data streaming pipeline service.
- **Key Features**: Stream processing, real-time data transformation

#### [ecron](ecron/)
Distributed cron job scheduling and execution service.
- **Language**: Go
- **Key Features**: Cron-based scheduling, immediate execution, task monitoring, multi-node support

### Frontend Applications

#### [dia-flow-web](dia-flow-web/)
Visual designer for building data processing flows.
- **Technology**: Modern web framework
- **Features**: Drag-and-drop pipeline design, node configuration, execution monitoring

### Shared Libraries

#### [ide-go-lib](ide-go-lib/)
Common Go libraries shared across Go-based services.

## Tech Stack

### Backend Services
- **Go**: flow-automation, ecron, ide-go-lib
- **Python**: coderunner, flow-stream-data-pipeline

### Frameworks & Libraries
- **Go**: Gin, MongoDB, Redis, Kafka
- **Python**: Tornado, RestrictedPython, pandas, SQLAlchemy

### Infrastructure
- **Databases**: MongoDB, MySQL/MariaDB, Redis
- **Message Queues**: Kafka, NSQ
- **Container Orchestration**: Kubernetes (Helm)
- **Authentication**: OAuth2

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Kubernetes cluster (for production deployment)
- MongoDB, MySQL/MariaDB, Redis
- Kafka or NSQ message queue

### Quick Start with Docker Compose

```bash
# Clone the repository
git clone <repository-url>
cd dataflow

# Start all services
docker-compose up -d

# Access the application
# Data Flow Designer: http://localhost:3000
```

### Individual Service Setup

Each service can be run independently. Refer to the README in each service directory for specific setup instructions:

- [flow-automation/README.md](flow-automation/README.md)
- [coderunner/README.md](coderunner/README.md)
- [ecron/README.md](ecron/README.md)

## Deployment

### Kubernetes Deployment

Each service includes Helm charts for Kubernetes deployment:

```bash
# Deploy flow-automation
cd flow-automation/helm
helm install flow-automation . -f values.yaml

# Deploy coderunner
cd coderunner/helm
helm install coderunner . -f values.yaml

# Deploy ecron
cd ecron/helm
helm install ecron . -f values.yaml
```

### Configuration

Each service uses environment variables or configuration files. Key configuration areas:

- **Database connections**: MongoDB, MySQL, Redis
- **Message queue**: Kafka/NSQ endpoints
- **Authentication**: OAuth2 service endpoints
- **Service discovery**: Internal service URLs

## Integration

### External System Integration

- **OAuth2 Authentication**: Integrate with external identity providers
- **Message Queues**: Connect to Kafka/NSQ for event-driven architectures
- **REST APIs**: All services expose RESTful APIs for integration
- **Webhooks**: Configure webhooks for event notifications

## Development

### Project Structure

```
dataflow/
├── flow-automation/       # Data flow orchestration (Go)
├── coderunner/           # Code execution service (Python)
├── ecron/                # Scheduled tasks (Go)
├── flow-stream-data-pipeline/  # Streaming pipeline (Python)
├── dia-flow-web/         # Data flow UI
└── ide-go-lib/          # Shared Go libraries
```

### Contributing

1. Choose the service you want to contribute to
2. Follow the development guide in the service's README
3. Write tests for your changes
4. Submit a pull request

### Code Style

- **Go**: Follow Go standard conventions, use `golangci-lint`
- **Python**: Follow PEP 8, use `black` and `pylint`

## Documentation

- [Flow Automation Documentation](flow-automation/README.md)
- [CodeRunner Documentation](coderunner/README.md)
- [ECron Documentation](ecron/README.md)

## Use Case Examples

### Example 1: Automated Data Processing Pipeline

1. Design a data flow in dia-flow-web
2. Configure data source connections
3. Add transformation nodes with custom Python code
4. Set up scheduled execution via ecron
5. Monitor execution in flow-automation dashboard

### Example 2: Document Processing Pipeline

1. Create a data flow for document ingestion
2. Add OCR and text extraction nodes
3. Configure data transformation and validation
4. Set up automated report generation
5. Monitor processing results

### Example 3: Real-time Data Streaming

1. Configure data sources for streaming
2. Design transformation pipeline
3. Set up real-time processing rules
4. Monitor streaming data flow
5. Export processed results

## Monitoring and Observability

- **Health Checks**: All services expose health endpoints
- **Metrics**: Prometheus metrics for monitoring
- **Logging**: Structured logging across all services
- **Tracing**: Distributed tracing support

## Security

- **Authentication**: OAuth2-based authentication
- **Authorization**: Role-based access control
- **Code Execution**: Sandboxed execution environment
- **Data Isolation**: Multi-tenant data isolation
- **Audit Trails**: Comprehensive audit logging

## Support

For questions and support:
- Check service-specific README files
- Review API documentation
- Contact the development team