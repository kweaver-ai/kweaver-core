# Deploy

KWeaver Core is installed with the `deploy.sh` script under the `deploy/` directory in this repository.

## Clone and enter deploy directory

```bash
git clone https://github.com/kweaver-ai/kweaver-core.git
cd kweaver-core/deploy
chmod +x deploy.sh
```

## Install KWeaver Core

### Minimum install (recommended for first try)

Skips some optional modules (e.g. auth / business domain) for a lighter footprint:

```bash
./deploy.sh kweaver-core install --minimum
```

Equivalent flags:

```bash
./deploy.sh kweaver-core install --set auth.enabled=false --set businessDomain.enabled=false
```

### Full install

Includes auth and business-domain related components:

```bash
./deploy.sh kweaver-core install
```

The script may prompt for **access address** and detect **API server address** automatically.

### Non-interactive install

```bash
./deploy.sh kweaver-core install \
  --access_address=<your-ip-or-domain> \
  --api_server_address=<nic-ip-for-k8s-api>
```

- `--access_address` — URL or IP clients use to reach KWeaver (ingress)
- `--api_server_address` — real NIC IP bound for the Kubernetes API server

### Custom ingress ports (optional)

```bash
export INGRESS_NGINX_HTTP_PORT=8080
export INGRESS_NGINX_HTTPS_PORT=8443
./deploy.sh kweaver-core install
```

## Useful commands

```bash
./deploy.sh kweaver-core status
./deploy.sh kweaver-core uninstall
./deploy.sh --help
```

## What gets installed

1. Single-node Kubernetes (if needed), storage, ingress
2. Data services: MariaDB, Redis, Kafka, ZooKeeper, OpenSearch (as defined by release manifests)
3. KWeaver Core application Helm charts

For uninstall and cluster reset, see [deploy/README.md](../../../deploy/README.md).

## Next steps

- [Verify installation](verify.md)
- [Quick start](../quick-start.md)
