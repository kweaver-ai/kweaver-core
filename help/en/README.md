# KWeaver Core Documentation

KWeaver Core is a **backend-only** platform. Use the CLI, SDKs, or HTTP APIs to operate each subsystem.

## Getting started

- [Prerequisites](installation/prerequisites.md) — hardware, OS, network, and tooling
- [Deploy](installation/deploy.md) — one-click install with `deploy.sh`, including post-install checks
- [Quick start](quick-start.md) — end-to-end path from deploy to first BKN and agent actions

## Modules

| Doc | Description |
| --- | --- |
| [Data Source Management](datasource.md) | Database connections, table discovery, CSV import, lifecycle |
| [Model Management](model.md) | LLM, Embedding, and Reranker registration and management |
| [BKN Engine](bkn.md) | Business Knowledge Network — object types, relations, actions, instances |
| [VEGA Engine](vega.md) | Data virtualization — connections, models, views, unified query |
| [Context Loader](context-loader.md) | Agent context assembly from ontology and data |
| [Execution Factory](execution-factory.md) | Tools, operators, and skills for agents |
| [Dataflow](dataflow.md) | Pipeline orchestration and automation |
| [Decision Agent](decision-agent.md) | Goal-oriented agents, runtime, and observability |
| [Trace AI](trace-ai.md) | Traces, metrics, and evidence-chain style observability |
| [Info Security Fabric](isf.md) | Identity, permissions, policies, and audit (when enabled) |

## Community

<img src="../qrcode.png" width="200" alt="KWeaver community QR code" />

For deployment details beyond this help set, see the [Deployment guide](../../deploy/README.md). For the `kweaver` CLI and language SDKs, see [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk).
