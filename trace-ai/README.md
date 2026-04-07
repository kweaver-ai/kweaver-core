# Tracing AI

English | [中文](README.zh.md)

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE.txt)

Tracing AI is a verification and observability framework for LLM applications and agent systems. It is built to turn opaque AI execution into inspectable, attributable, and production-ready workflows through end-to-end tracing, structured correlation, and queryable evidence.

This repository currently contains two core building blocks:

- `agent-observability`: a Go-based trace query service for searching agent traces from OpenSearch
- `otelcol-contribute-chart`: a Helm chart for deploying OpenTelemetry Collector Contrib with OTLP ingestion and OpenSearch export

## Why Tracing AI

Traditional application tracing tells you whether a request failed. AI-native systems require more: what the model saw, which tool it called, what knowledge it retrieved, and why a result should be trusted.

Tracing AI is designed to support that transition:

- End-to-end execution trace observability across prompts, model calls, tool invocations, retrieval, and intermediate reasoning-related spans
- Evidence tracing that links outputs to data sources, knowledge items, and execution context
- Timeline-style replay and trace inspection for locating latency bottlenecks and bad cases
- A foundation for closed-loop optimization from failed production traces to evaluation cases
- A path toward automated root cause analysis for multi-agent systems

## Core Capabilities

### Available in this repository today

- OpenTelemetry-based ingestion path through Collector deployment on Kubernetes
- OTLP trace and log receiving via OpenTelemetry Collector Contrib
- OpenSearch export pipeline for collected telemetry
- Trace query service for raw DSL search and conversation-based lookup
- Swagger documentation, Docker image build, Helm chart packaging, and GitHub Actions release workflows

### Target capabilities of Tracing AI

The following capabilities describe the product direction of Tracing AI. Some are partially enabled by the current architecture, while others are planned on top of the existing foundation:

- Full execution trajectory capture, including model input/output, tool calls, retrieval, and reasoning steps
- Evidence-grounded decision analysis with source-level attribution
- Visual timeline replay for complex agent runs
- Closed-loop workflow from failed traces to reusable evaluation cases
- Intelligent RCA using causal analysis methods in multi-agent environments

## Architecture

Tracing AI follows OpenTelemetry conventions so telemetry data can be collected in a standard, non-intrusive way.

- `Trace`: the full lifecycle of one AI interaction
- `Span`: one operation within the trace, such as a model call, retrieval, or tool execution
- `Collector`: receives OTLP data, batches and routes telemetry, then exports it to storage
- `Query Service`: exposes APIs to search and inspect stored traces
- `Storage`: currently OpenSearch-oriented in this repository; the broader architecture can evolve toward other high-scale AI data stores

Current repository architecture:

```text
LLM App / Agent
  -> OTLP
OpenTelemetry Collector
  -> OpenSearch
agent-observability
  -> trace query APIs / Swagger
```

## Repository Layout

```text
.
|-- agent-observability/
|   |-- main.go
|   |-- Dockerfile
|   |-- Makefile
|   |-- charts/agent-observability/
|   `-- docs/
|-- otelcol-contribute-chart/
|   |-- charts/otelcol-contrib/
|   `-- scripts/
`-- .github/workflows/
```

## Components

### agent-observability

`agent-observability` is the trace query service in this repository. It currently provides:

- `POST /api/v1/traces/_search`: proxy raw OpenSearch DSL to the configured trace index
- `GET /api/v1/traces/by-conversation?conversation_id=...`: search traces by conversation ID
- Swagger endpoints under `/swagger/`

Local development:

```bash
cd agent-observability
make test
make gen-swag
make docker-build
```

Run with Helm:

```bash
helm upgrade --install agent-observability agent-observability/charts/agent-observability \
  --set image.repository=swr.cn-east-3.myhuaweicloud.com/kweaver-ai/agent-observability \
  --set image.tag=0.1.0 \
  --set opensearch.endpoint=http://opensearch-read.resource.svc.cluster.local:9200 \
  --set opensearch.auth.enabled=false \
  -n observability --create-namespace
```

### otelcol-contribute-chart

`otelcol-contribute-chart` packages OpenTelemetry Collector Contrib for Kubernetes deployment and provides:

- Deployment-based Collector installation
- OTLP gRPC and HTTP receivers
- OpenSearch exporter with optional basic auth
- GHCR chart packaging workflow

Quick validation:

```bash
helm lint otelcol-contribute-chart/charts/otelcol-contrib
helm template otelcol-contrib otelcol-contribute-chart/charts/otelcol-contrib
```

Install example:

```bash
helm upgrade --install otelcol-contrib otelcol-contribute-chart/charts/otelcol-contrib \
  -n observability \
  --create-namespace \
  --set opensearchExporter.http.endpoint=http://opensearch-read.resource.svc.cluster.local:9200
```

Create and publish a multi-arch collector manifest:

```bash
docker buildx imagetools create \
  -t swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/opentelemetry-collector-contrib:0.148.0 \
  swr.cn-north-4.myhuaweicloud.com/ddn-k8s/ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-contrib:0.148.0 \
  swr.cn-north-4.myhuaweicloud.com/ddn-k8s/ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-contrib:0.148.0-linuxarm64
```

Verify the manifest platforms:

```bash
docker buildx imagetools inspect \
  swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/opentelemetry-collector-contrib:0.148.0
```

## Quick Start

1. Deploy `otelcol-contrib` to receive OTLP telemetry and export it to OpenSearch.
2. Send telemetry from your LLM app or agent runtime through OTLP.
3. Deploy `agent-observability`.
4. Query traces through:

```text
POST /api/v1/traces/_search
GET  /api/v1/traces/by-conversation
GET  /swagger/index.html
```

## Business Value

- Improve trust in AI systems with traceable evidence
- Reduce uncertainty by inspecting critical execution nodes instead of only final outputs
- Shorten debugging cycles with searchable production traces
- Build evaluation datasets from real-world execution behavior rather than intuition alone

## Related Docs

- `agent-observability/README.md`
- `otelcol-contribute-chart/README.md`
- `agent-observability/docs/design/agent-tracing-system-design.md`
- `agent-observability/docs/prd/agent-tracing-system-prd.md`

## License

Apache 2.0. See `LICENSE.txt`.
