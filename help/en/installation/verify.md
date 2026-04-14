# Verify installation

After `deploy.sh kweaver-core install` completes, confirm the cluster and APIs.

## Kubernetes

```bash
kubectl get nodes
kubectl get pods -A
```

Wait until core namespaces show `Running` / `Ready` for critical workloads.

## Deploy script status

```bash
cd kweaver-core/deploy
./deploy.sh kweaver-core status
```

## CLI access

Install the CLI from [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk), then log in and list BKNs:

```bash
kweaver auth login https://<access-address> -k
kweaver bkn list
```

Replace `<access-address>` with the value you used for `--access_address` or the node IP shown by the installer.

## HTTP sanity check

If you know a public route (paths vary by gateway and version), you can probe with curl:

```bash
export KWEAVER_BASE="https://<access-address>"
curl -sk "$KWEAVER_BASE/health" || true
```

Use OpenAPI or ingress rules from your environment for subsystem-specific paths.

## Troubleshooting

See [deploy/README.md — Troubleshooting](../../../deploy/README.md).

## Next steps

- [Quick start](../quick-start.md)
- Module docs from the [documentation index](../README.md)
