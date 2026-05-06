# Sandbox Helm Charts

`deploy/helm` contains two charts for different deployment scenarios. They are intentionally separate because Kweaver Core component packaging and standalone Sandbox deployment have different ownership boundaries and values formats.

## Chart Comparison

| Chart | Use Case | Includes | Database | Web Console | Values Format |
|-------|----------|----------|----------|-------------|---------------|
| `sandbox` | Kweaver Core component deployment | Control Plane, MinIO, RBAC, config, secrets | Uses Core-provided `depServices.rds` | Not included | Follows Kweaver Core component conventions |
| `sandbox_standalone` | Independent Sandbox deployment for development, demos, or non-Core environments | Control Plane, Web Console, MariaDB, MinIO, RBAC, config, secrets | Deploys internal MariaDB by default | Included | Standalone chart values, not constrained by Core |

## Which Chart Should I Use?

Use `deploy/helm/sandbox` when Sandbox is packaged as a Kweaver Core component. This chart assumes the platform already provides shared infrastructure such as the database service. It should stay compatible with Core deployment conventions.

Use `deploy/helm/sandbox_standalone` when you want a self-contained Sandbox stack. This chart is better for local development, feature verification, demos, and environments that do not install Sandbox through Kweaver Core.

## Default Template Images

Both charts initialize two default templates in `t_sandbox_template`:

- `python-basic`
- `multi-language`

Both charts expose the template image versions through the same values shape:

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

During rendering, these values become `DEFAULT_TEMPLATE_IMAGE` and `DEFAULT_MULTI_LANGUAGE_TEMPLATE_IMAGE` for the Control Plane. The Control Plane seed logic uses those environment variables before falling back to the image tag derived from `VERSION`.

## Examples

Render the Kweaver Core component chart:

```bash
helm template sandbox deploy/helm/sandbox
```

Render the standalone chart:

```bash
helm template sandbox deploy/helm/sandbox_standalone
```

Override default template versions during deployment:

```bash
helm upgrade --install sandbox deploy/helm/sandbox \
  --set image.defaultTemplates.pythonBasic.tag=dev-python \
  --set image.defaultTemplates.multiLanguage.tag=dev-multi
```

For standalone deployment, replace the chart path with `deploy/helm/sandbox_standalone`.
