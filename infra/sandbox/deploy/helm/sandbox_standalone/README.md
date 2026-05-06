# Sandbox Standalone Helm Chart

This chart deploys Sandbox as a self-contained application.

Use this chart for development, demos, integration testing, or environments where Sandbox should be deployed independently from Kweaver Core. It owns its supporting services and is not constrained by the Core component values format.

For Kweaver Core component packaging, use `deploy/helm/sandbox`.

## Components

- **Control Plane**: FastAPI management service for sessions, templates, scheduling, and storage access.
- **Web Console**: React application for visual session, template, and execution management.
- **MariaDB**: Internal database for Sandbox data.
- **MinIO**: S3-compatible workspace object storage.
- **Default template metadata**: Initializes `python-basic` and `multi-language` template records.

## Prerequisites

- Kubernetes 1.24+
- Helm 3.0+
- PV provisioner support for MariaDB and MinIO persistence

## Installing

From the repository root:

```bash
helm install sandbox deploy/helm/sandbox_standalone \
  --namespace sandbox-system \
  --create-namespace
```

With an override file:

```bash
helm install sandbox deploy/helm/sandbox_standalone \
  --namespace sandbox-system \
  --create-namespace \
  -f custom-values.yaml
```

## Default Template Images

The Control Plane seeds two default template rows into `t_sandbox_template`:

- `python-basic`
- `multi-language`

The chart exposes both image tags independently:

```yaml
image:
  defaultTemplates:
    pythonBasic:
      repository: sandbox-template-python-basic
      tag: latest
    multiLanguage:
      repository: sandbox-template-multi-language
      tag: latest
```

The deprecated `image.defaultTemplate` value is still accepted as a compatibility fallback for `python-basic`.

## Key Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `namespace` | Kubernetes namespace | `sandbox-system` |
| `image.registry` | Shared image registry prefix | `""` |
| `image.controlPlane.repository` | Control Plane image repository | `sandbox-control-plane` |
| `image.web.repository` | Web Console image repository | `sandbox-web` |
| `image.mariadb.repository` | MariaDB image repository | `mariadb` |
| `image.minio.repository` | MinIO image repository | `quay.io/minio/minio` |
| `image.defaultTemplates.pythonBasic.tag` | Python template image tag | `latest` |
| `image.defaultTemplates.multiLanguage.tag` | Multi-language template image tag | `latest` |
| `mariadb.enabled` | Deploy internal MariaDB | `true` |
| `web.enabled` | Deploy Sandbox Web Console | `true` |
| `minio.enabled` | Deploy internal MinIO | `true` |

## Access

```bash
kubectl port-forward svc/sandbox-control-plane 8000:8000 -n sandbox-system
kubectl port-forward svc/sandbox-web 1101:80 -n sandbox-system
kubectl port-forward svc/minio 9001:9001 -n sandbox-system
```

## Rendering

```bash
helm template sandbox deploy/helm/sandbox_standalone
```

To verify manual template versions:

```bash
helm template sandbox deploy/helm/sandbox_standalone \
  --show-only templates/configmap.yaml \
  --set image.defaultTemplates.pythonBasic.tag=dev-python \
  --set image.defaultTemplates.multiLanguage.tag=dev-multi
```

## Related Chart

See `deploy/helm/README.md` for the comparison between `sandbox` and `sandbox_standalone`.
