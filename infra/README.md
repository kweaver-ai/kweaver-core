# Infra

[中文](README.zh.md) | English

Infrastructure services that provide foundational capabilities for the KWeaver platform.

## Services

| Service | Description | Language | Version |
|---------|-------------|----------|---------|
| [sandbox](sandbox/) | Cloud-native secure code execution platform for AI agent applications | Python | see `sandbox/VERSION` |
| [oss-gateway-backend](oss-gateway-backend/) | Unified object storage gateway (Alibaba Cloud OSS, Huawei Cloud OBS, Ceph S3) | Go | — |
| [mf-model-manager](mf-model-manager/) | Model Factory — model lifecycle management and Kafka consumer | Python | see `mf-model-manager/VERSION` |
| [mf-model-api](mf-model-api/) | Model Factory — LLM/SLM API gateway | Python | — |

## Architecture

```
infra/
├── sandbox/                  # Sandbox Control Plane + Web + Runtime
│   ├── sandbox_control_plane/  # Control plane API server
│   ├── sandbox_web/            # Web UI
│   ├── runtime/executor/       # Code execution daemon
│   ├── deploy/helm/            # Helm chart
│   └── docs/                   # Design docs & PRDs
├── oss-gateway-backend/      # Object storage gateway
│   ├── internal/               # Business logic
│   ├── pkg/                    # Shared packages
│   ├── charts/                 # Helm chart
│   └── migrations/             # DB migrations
├── mf-model-manager/         # Model lifecycle manager
│   ├── app/                    # Application code
│   ├── charts/                 # Helm chart
│   └── migrations/             # DB migrations
└── mf-model-api/             # Model API service
    ├── app/                    # Application code
    └── charts/                 # Helm chart
```

## Development

Each service can be built and deployed independently. See individual service READMEs for setup instructions:

- **sandbox**: Full documentation in [sandbox/README.md](sandbox/README.md)
- **oss-gateway-backend**: API docs in [oss-gateway-backend/README.md](oss-gateway-backend/README.md)
- **mf-model-manager**: Setup in [mf-model-manager/README.md](mf-model-manager/README.md)
- **mf-model-api**: Setup in [mf-model-api/README.md](mf-model-api/README.md)
