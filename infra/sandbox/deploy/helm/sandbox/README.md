# Sandbox Core Component Helm Chart

This chart deploys Sandbox as a Kweaver Core component.

Use this chart when Sandbox is installed by, or packaged into, Kweaver Core. It follows the Core component values format and expects shared platform services, especially the database service, to be provided by Core.

For a self-contained Sandbox deployment with MariaDB and the web console, use `deploy/helm/sandbox_standalone`.

## Components

- **Control Plane**: FastAPI management service for sessions, templates, scheduling, and storage access.
- **MinIO**: S3-compatible workspace object storage when enabled by this chart.
- **Default template metadata**: Initializes `python-basic` and `multi-language` template records through Control Plane environment variables.

This chart does not deploy `sandbox_web` or MariaDB.

## Prerequisites

- Kubernetes 1.24+
- Helm 3.0+
- Kweaver Core deployment environment
- Core-provided database service through `depServices.rds`

## Installing

From the repository root:

```bash
helm install sandbox deploy/helm/sandbox --namespace anyshare --create-namespace
```

With an override file:

```bash
helm install sandbox deploy/helm/sandbox \
  --namespace anyshare \
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
      repository: dip/sandbox-template-python-basic
      tag: ""
    multiLanguage:
      repository: dip/sandbox-template-multi-language
      tag: ""
```

An empty `tag` means the Control Plane reads `/app/VERSION` at startup and uses that value as the template image tag. Set the tag only when deploying branch or custom template images.

The deprecated `image.defaultTemplate` value is still accepted as a compatibility fallback for `python-basic`.

## Key Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `namespace` | Kubernetes namespace | `anyshare` |
| `imagePullPolicy` | Executor Pod image pull policy | `IfNotPresent` |
| `imagePullSecrets` | Executor Pod image pull secret names | `[]` |
| `image.registry` | Shared image registry prefix | `acr.aishu.cn` |
| `image.controlPlane.repository` | Control Plane image repository | `dip/sandbox-control-plane` |
| `image.controlPlane.tag` | Control Plane image tag | `latest` |
| `image.defaultTemplates.pythonBasic.repository` | Python template image repository | `dip/sandbox-template-python-basic` |
| `image.defaultTemplates.pythonBasic.tag` | Python template image tag; empty uses `/app/VERSION` | `""` |
| `image.defaultTemplates.multiLanguage.repository` | Multi-language template image repository | `dip/sandbox-template-multi-language` |
| `image.defaultTemplates.multiLanguage.tag` | Multi-language template image tag; empty uses `/app/VERSION` | `""` |
| `controlPlane.env.ENVIRONMENT` | Control Plane environment | `staging` |
| `depServices.rds` | Core-provided database service configuration | enabled by values |

## Rendering

```bash
helm template sandbox deploy/helm/sandbox
```

To verify manual template versions:

```bash
helm template sandbox deploy/helm/sandbox \
  --show-only templates/configmap.yaml \
  --set image.defaultTemplates.pythonBasic.tag=dev-python \
  --set image.defaultTemplates.multiLanguage.tag=dev-multi
```

## Related Chart

See `deploy/helm/README.md` for the comparison between `sandbox` and `sandbox_standalone`.
