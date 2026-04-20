# Sandbox Platform Helm Chart

This Helm chart deploys the Sandbox Platform - a secure code execution platform with isolated container environments for AI agent applications.

## Introduction

The Sandbox Platform consists of:
- **Control Plane**: FastAPI-based management service handling API requests, scheduling, session management
- **Web Console**: React-based web application for visual management of sessions, templates, and executions
- **MariaDB**: Database for session, execution, and template storage
- **MinIO**: S3-compatible object storage for workspace files

## Prerequisites

- Kubernetes 1.24+
- Helm 3.0+
- PV provisioner support in the underlying infrastructure

## Installing the Chart

### Install with default values

```bash
helm install sandbox ./sandbox --namespace sandbox-system --create-namespace
```

### Install with custom values file

```bash
helm install sandbox ./sandbox --namespace sandbox-system --create-namespace -f custom-values.yaml
```

### Install with specific parameters

```bash
helm install sandbox ./sandbox \
  --namespace sandbox-system \
  --create-namespace \
  --set controlPlane.replicaCount=3 \
  --set controlPlane.resources.limits.cpu=1
```

## Uninstalling the Chart

```bash
helm uninstall sandbox --namespace sandbox-system
```

## Configuration

The following table lists the configurable parameters of the Sandbox chart and their default values.

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `namespace` | Namespace for all resources | `sandbox-system` |
| `image.registry` | Shared image registry used by chart images | `""` |
| `imagePullPolicy` | Global image pull policy | `IfNotPresent` |
| `image.controlPlane.repository` | Control Plane image repository | `sandbox-control-plane` |
| `image.controlPlane.tag` | Control Plane image tag | `latest` |
| `image.defaultTemplate.repository` | Default sandbox template image repository | `sandbox-template-python-basic` |
| `image.defaultTemplate.tag` | Default sandbox template image tag | `v1.0.0` |
| `image.web.repository` | Web Console image repository | `sandbox-web` |
| `image.web.tag` | Web Console image tag | `latest` |
| `image.mariadb.repository` | MariaDB image repository | `mariadb` |
| `image.mariadb.tag` | MariaDB image tag | `11.2` |
| `image.minio.repository` | MinIO image repository | `quay.io/minio/minio` |
| `image.minio.tag` | MinIO image tag | `latest` |
| `image.busybox.repository` | BusyBox image repository for init containers | `busybox` |
| `image.busybox.tag` | BusyBox image tag for init containers | `1.36` |

### Control Plane Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `controlPlane.enabled` | Enable Control Plane deployment | `true` |
| `controlPlane.replicaCount` | Number of replicas | `1` |
| `controlPlane.service.type` | Kubernetes service type | `ClusterIP` |
| `controlPlane.service.port` | Service port | `8000` |
| `controlPlane.resources.requests.cpu` | CPU request | `200m` |
| `controlPlane.resources.requests.memory` | Memory request | `256Mi` |
| `controlPlane.resources.limits.cpu` | CPU limit | `500m` |
| `controlPlane.resources.limits.memory` | Memory limit | `512Mi` |

### Web Console Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `web.enabled` | Enable Web Console deployment | `true` |
| `web.replicaCount` | Number of replicas | `1` |
| `web.service.port` | Service port | `80` |
| `web.env.VITE_API_BASE_URL` | API base URL for web console | `http://sandbox-control-plane.sandbox-system.svc.cluster.local:8000` |

### MariaDB Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `mariadb.enabled` | Enable MariaDB deployment | `true` |
| `mariadb.replicaCount` | Number of replicas | `1` |
| `mariadb.auth.rootPassword` | MariaDB root password | `password` |
| `mariadb.auth.database` | Database name | `sandbox` |
| `mariadb.persistence.enabled` | Enable persistence | `true` |
| `mariadb.persistence.size` | Persistent volume size | `10Gi` |

### MinIO Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `minio.enabled` | Enable MinIO deployment | `true` |
| `minio.replicaCount` | Number of replicas | `1` |
| `minio.auth.rootUser` | MinIO root user | `minioadmin` |
| `minio.auth.rootPassword` | MinIO root password | `minioadmin` |
| `minio.defaultBuckets` | Default buckets to create | `["sandbox-workspace"]` |
| `minio.persistence.enabled` | Enable persistence | `true` |
| `minio.persistence.size` | Persistent volume size | `10Gi` |

## Accessing the Services

After installation, you can access the services using port-forwarding:

```bash
# Control Plane API
kubectl port-forward svc/sandbox-control-plane 8000:8000 -n sandbox-system

# Web Console
kubectl port-forward svc/sandbox-web 1101:80 -n sandbox-system

# MinIO Console
kubectl port-forward svc/minio 9001:9001 -n sandbox-system
```

Or use the provided port-forward script:

```bash
cd ../../scripts
./port-forward.sh start --all --background
```

## Upgrading

### Upgrade the chart

```bash
helm upgrade sandbox ./sandbox --namespace sandbox-system
```

### Upgrade with new values

```bash
helm upgrade sandbox ./sandbox \
  --namespace sandbox-system \
  --set controlPlane.replicaCount=3 \
  --reuse-values
```

## Rollback

### Rollback to a previous revision

```bash
helm rollback sandbox <revision> --namespace sandbox-system
```

## Development

### Template generation

Generate the Kubernetes manifests without installing:

```bash
make template
# or
helm template sandbox ./sandbox > helm-template-gen.yaml
```

### Linting

Lint the chart for syntax errors:

```bash
make lint
# or
helm lint ./sandbox
```

### Dry-run installation

Test the installation without actually deploying:

```bash
make test
# or
helm install sandbox ./sandbox --namespace sandbox-system --dry-run --debug
```

## Custom Values Examples

### Production deployment

```yaml
# production-values.yaml
controlPlane:
  replicaCount: 3
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 2
      memory: 2Gi

web:
  replicaCount: 2

mariadb:
  persistence:
    size: 50Gi

minio:
  persistence:
    size: 100Gi
```

Install with production values:

```bash
helm install sandbox ./sandbox \
  --namespace sandbox-system \
  --create-namespace \
  -f production-values.yaml
```

### External database and storage

```yaml
# external-values.yaml
mariadb:
  enabled: false

minio:
  enabled: false

controlPlane:
  env:
    DATABASE_URL: "mysql+aiomysql://user:pass@external-mariadb:3306/sandbox"
  s3:
    endpointUrl: "https://s3.amazonaws.com"
    bucket: "my-sandbox-bucket"

secret:
  s3:
    accessKeyId: "YOUR_ACCESS_KEY"
    secretAccessKey: "YOUR_SECRET_KEY"
```

## Troubleshooting

### Check pod status

```bash
kubectl get pods -n sandbox-system
```

### View logs

```bash
# Control Plane logs
kubectl logs -f deployment/sandbox-control-plane -n sandbox-system

# Web Console logs
kubectl logs -f deployment/sandbox-web -n sandbox-system

# MariaDB logs
kubectl logs -f deployment/mariadb -n sandbox-system

# MinIO logs
kubectl logs -f deployment/minio -n sandbox-system
```

### Check events

```bash
kubectl get events -n sandbox-system --sort-by='.lastTimestamp'
```

## License

Copyright © 2025
