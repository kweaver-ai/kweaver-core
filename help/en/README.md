# KWeaver Core documentation

KWeaver Core is a **backend-only** platform. Use the CLI, SDKs, or HTTP APIs to operate each subsystem.

---

## Getting started

**Deploy:** **Linux** is recommended for full installs. **macOS** (optional): local validation with kind — [`deploy/dev/README.md`](../../deploy/dev/README.md) ([中文](../../deploy/dev/README.zh.md)).

| Doc | Description |
| --- | --- |
| [Install and deploy](install.md) | Prerequisites, `deploy.sh` install, and post-install checks |
| [Quick start](quick-start.md) | End-to-end path from deploy to first BKN and agent actions |

---

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
| [Platform admin tool](install.md#-administrator-tool-after-a-full-install-kweaver-admin) | `kweaver-admin` — users / orgs / roles / models / audit (after a full install) |

---

## Community

<img src="../qrcode.png" width="200" alt="KWeaver community QR code" />

> End users install the CLI with `npm install -g @kweaver-ai/kweaver-sdk`; platform administrators additionally install `npm install -g @kweaver-ai/kweaver-admin`. For cluster operations beyond this help set, follow the deployment guide bundled with your release.
