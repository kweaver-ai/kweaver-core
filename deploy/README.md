# KWeaver Deploy

[中文](README.zh.md) | English

One-click deployment of the KWeaver AI platform to a single-node Kubernetes cluster.

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](../LICENSE.txt)

## 🚀 Quick Start

```bash
# Option 1: Quick install from release package (recommended)
# Installs latest release automatically
curl -fsSL https://raw.githubusercontent.com/kweaver-ai/kweaver/main/install.sh | sh

# Install specific version
curl -fsSL https://raw.githubusercontent.com/kweaver-ai/kweaver/main/install.sh | sh -s -- --version v1.0.0

# Option 2: Clone the repository
# 1. Clone the repository
git clone https://github.com/kweaver-ai/kweaver.git
cd kweaver/deploy

# 2. Edit the config file (optional; skip to use defaults)
# vim conf/config.yaml

# 3. Deploy all components (installs the latest version by default)
# Note: deploy.sh requires root privileges
sudo bash ./deploy.sh full init
```

For first-time installation, especially on cloud VMs or slower networks, the full deployment process may take a while. In some environments it can take more than an hour to complete.

After deployment, open `https://<node-ip>/studio`. Username: `admin`, initial password: `eisoo.com`.

## 📋 Prerequisites

### Permissions

**`./deploy.sh` requires root privileges** to install system packages, configure Kubernetes, and manage cluster resources. Run with `sudo` or as root:

```bash
sudo bash ./deploy.sh full init
# or
sudo bash ./deploy.sh kweaver init
```

### System requirements

| Item | Minimum | Recommended |
| --- | --- | --- |
| OS | CentOS 7/8+, RHEL 8 | CentOS 7 |
| CPU | 16 cores | 24 cores |
| Memory | 48 GB | 64 GB |
| Disk | 200 GB | 500 GB |

### Host prerequisites (required)

```bash
# 1. Disable firewall
systemctl stop firewalld && systemctl disable firewalld

# 2. Disable swap
swapoff -a && sed -i '/ swap / s/^/#/' /etc/fstab

# 3. Disable SELinux (optional; the script may handle this)
setenforce 0

# 4. Manually install container-selinux
```

### Network requirements

The deployment scripts need access to the following domains:

| Domain | Purpose |
| --- | --- |
| `mirrors.aliyun.com` | RPM package mirrors |
| `mirrors.tuna.tsinghua.edu.cn` | TUNA `containerd.io` RPM mirror |
| `registry.aliyuncs.com` | Kubernetes component images |
| `swr.cn-east-3.myhuaweicloud.com` | Application images registry |
| `repo.huaweicloud.com` | Helm binary download |
| `kweaver-ai.github.io` | KWeaver Helm chart repository |

## 📦 Components

### Infrastructure

- **Kubernetes** v1.28 (single-node)
- **containerd** v1.6+
- **Flannel CNI** v0.25.5
- **ingress-nginx** v1.14.1

### Data services

- **MariaDB** v11.4.7
- **MongoDB** v4.4.30
- **Redis** v7.4.6 (Sentinel)
- **Kafka** v3.9.0
- **OpenSearch** v2.19.4
- **ZooKeeper** v3.9.3

## 🔧 Usage

### Deployment commands

**Note:** All `deploy.sh` commands require root privileges. Use `sudo` or run as root.

```bash
# Full one-click deployment (recommended)
sudo bash ./deploy.sh full init     # Infrastructure + KWeaver application services

# Layered deployment
sudo bash ./deploy.sh infra init    # Infrastructure only: K8s + data services
sudo bash ./deploy.sh kweaver init  # Application services only: ISF/Studio/Ontology, etc.

# Deploy a single infrastructure component
sudo bash ./deploy.sh k8s init         # Kubernetes cluster
sudo bash ./deploy.sh mariadb init     # MariaDB
sudo bash ./deploy.sh mongodb init     # MongoDB
sudo bash ./deploy.sh redis init       # Redis
sudo bash ./deploy.sh kafka init       # Kafka
sudo bash ./deploy.sh opensearch init  # OpenSearch

# Deploy a single application service
sudo bash ./deploy.sh isf init         # ISF service
sudo bash ./deploy.sh studio init      # Studio service

# Specify Helm repo and version
sudo bash ./deploy.sh kweaver init --helm_repo=https://kweaver-ai.github.io/helm-repo/ --version=0.1.0

# Multiple version types are supported
sudo bash ./deploy.sh kweaver init --version=0.1.0              # Stable release
sudo bash ./deploy.sh kweaver init --version=0.0.0-feature-xxx  # Branch/dev build
sudo bash ./deploy.sh kweaver init                              # Latest

# Help (no root required)
./deploy.sh --help
```

### Verify deployment

```bash
# Cluster status
kubectl get nodes
kubectl get pods -A

# Service status (no root required)
./deploy.sh kweaver status
```

## ⚙️ Configuration

Config file: `conf/config.yaml`

Key settings:

```yaml
namespace: kweaver          # Namespace
image:
  registry: swr.cn-east-3.myhuaweicloud.com/kweaver-ai  # Image registry

accessAddress:
  host: <public-ip-or-domain>  # Required on cloud VMs
  port: 443
  scheme: https
  path: /

depServices:
  rds:
    source_type: internal   # internal=embedded MariaDB, external=external DB
    host: 'mariadb.resource.svc.cluster.local'
    user: 'adp'
    password: ''            # Auto-generated
```

If you deploy on a cloud VM, you must set `accessAddress.host` in `conf/config.yaml` to the public IP or public domain used for external access. Using an internal address may cause access failures after installation.

**Important:** For cloud VM deployments, you must also open ports **80** and **443** in your cloud provider's security group/firewall settings to allow external access. The ingress-nginx controller listens on these ports to serve the KWeaver platform.

### Use an external database

If you use an external database:

1. Change `source_type` to `external`
2. Configure external DB connection settings
3. Manually run the SQL initialization scripts under `scripts/sql/`

### Scenario-based auto configuration

The `auto_cofig` directory provides scripts to quickly set up a demo scenario (e.g. supply chain) after deployment. The workflow revolves around a central config file `config.env`:

1. **`setup_tem_db.sh`** reads database credentials from `conf/config.yaml`, creates the demo database in MariaDB, imports sample data (`dump-tem.sql`), and **writes back** the connection info (`DS_HOST`, `DS_USERNAME`, `DS_PASSWORD`) into `config.env`.
2. **`auto_config.sh`** reads `config.env` for authentication and data source settings, then creates data sources, imports knowledge networks, agents, data flows, etc.

**Prerequisites:**

1. Log in to the system console (`https://<node-ip>/deploy`, default: `admin/eisoo.com`)
2. Create a test user in **Information Security Management → Unified Identity Authentication → Accounts → Users**
3. Add the test user to roles: Data Administrator, AI Administrator, Application Administrator in **Roles & Access Policies → Role Management**
4. Log in to Studio (`https://<node-ip>/studio`) with the test user (default password: `123456`) and change the password when prompted

**Usage:**

```bash
cd deploy/auto_cofig

# Step 1: Prepare demo database (auto-fills config.env with DB credentials)
chmod +x setup_tem_db.sh auto_config.sh
./setup_tem_db.sh

# Step 2: Import scenario configuration
./auto_config.sh agent.json 供应链业务知识网络.json dataflow.json

# Step 3: Import toolboxes
./auto_config.sh --step 7 contextloader工具集_020.adp
./auto_config.sh --step 7 基础结构化数据分析工具箱2.adp
```

For detailed usage instructions, see `auto_cofig/README.md`.

## 📁 Project Structure

```
deploy/
├── deploy.sh           # Main entry script
├── conf/
│   ├── config.yaml              # Deployment config
│   ├── kube-flannel.yml         # Flannel network config
│   └── local-path-storage.yaml  # Local storage config
├── auto_cofig/                  # Scenario-based auto configuration
│   ├── auto_config.sh           # Auto configuration script
│   ├── config.env               # Configuration template
│   ├── README.md                 # Usage instructions
│   └── *.json, *.adp            # Example scenario files
└── scripts/
    ├── lib/
    │   └── common.sh            # Common utilities
    ├── services/                # Component installation scripts
    │   ├── k8s.sh
    │   ├── mariadb.sh
    │   ├── mongodb.sh
    │   └── ...
    └── sql/                     # SQL init scripts
        ├── isf/
        ├── studio/
        └── ...
```

## 🗑️ Uninstall

**Note:** Uninstall commands also require root privileges.

```bash
# Full uninstall
sudo bash ./deploy.sh full reset         # Uninstall everything (apps + infrastructure)

# Layered uninstall
sudo bash ./deploy.sh kweaver uninstall  # Uninstall application services only
sudo bash ./deploy.sh infra reset        # Uninstall infrastructure only

# Uninstall a single component
sudo bash ./deploy.sh mariadb uninstall
sudo bash ./deploy.sh k8s reset
```

## 🔍 Troubleshooting

### CoreDNS not ready

```bash
# Check whether firewall is disabled
systemctl status firewalld

# Restart CoreDNS
kubectl -n kube-system delete pod -l k8s-app=kube-dns
```

### Pods fail to pull images

```bash
# Check network connectivity
curl -I https://swr.cn-east-3.myhuaweicloud.com

# Check containerd config
cat /etc/containerd/config.toml
```

### Helm install timeout (context deadline exceeded)

If you see errors like `UPGRADE FAILED: context deadline exceeded` when installing services (especially ISF services like `sharemgnt-single`), the default timeout may be too short for slow networks or large images.

**Solution:** Increase timeout via environment variables before running deploy:

```bash
# For ISF services (default: 600s helm, 900s command)
export ISF_HELM_TIMEOUT=900s
export ISF_COMMAND_TIMEOUT=1200
sudo bash ./deploy.sh isf init

# For all services (global override)
export HELM_INSTALL_TIMEOUT=600s
export HELM_COMMAND_TIMEOUT=900
sudo bash ./deploy.sh kweaver init

# For Kafka (already has longer default: 1800s)
export KAFKA_HELM_TIMEOUT=2400s
sudo bash ./deploy.sh kafka init
```

**Diagnosis:**
```bash
# Check Pod status and events
kubectl get pods -n <namespace> | grep <service-name>
kubectl describe pod -n <namespace> <pod-name>
kubectl get events -n <namespace> --sort-by='.lastTimestamp' | tail -20
```

### Pod CrashLoopBackOff

If a Pod is in `CrashLoopBackOff` state (restarting repeatedly), the container is crashing after startup. Common causes include:

1. **Configuration errors** (missing/invalid config values)
2. **Database connection failures** (database not ready or wrong credentials)
3. **Missing dependencies** (other services not available)
4. **Resource limits** (memory/CPU constraints)
5. **Application errors** (check application logs)

**Diagnosis:**
```bash
# 1. Check Pod status and describe (shows events and container status)
kubectl get pod <pod-name> -n <namespace> -o wide
kubectl describe pod <pod-name> -n <namespace>

# 2. Check current container logs
kubectl logs <pod-name> -n <namespace> --tail=100

# 3. Check previous container logs (from crashed container)
kubectl logs <pod-name> -n <namespace> --previous --tail=100

# 4. Check recent events in namespace
kubectl get events -n <namespace> --sort-by='.lastTimestamp' | tail -30

# 5. Verify dependencies are ready (e.g., database)
kubectl get pods -n <namespace> | grep -E 'mariadb|mongodb|mysql'
kubectl get svc -n <namespace> | grep -E 'mariadb|mongodb|mysql'

# 6. Check Helm release status
helm status <release-name> -n <namespace>
```

**Common fixes:**
- **Database not ready**: Wait for database pods to be `Running` before installing dependent services
- **Wrong database credentials**: Check `config.yaml` database connection settings
- **Missing environment variables**: Verify Helm values and config.yaml
- **Resource constraints**: Check `kubectl describe pod` for `OOMKilled` or resource limit errors

### Kubernetes apt source 404 (Ubuntu/Debian)

If `apt update` fails with a 404 for the legacy `packages.cloud.google.com` repository:

```
Err:7 https://packages.cloud.google.com/apt kubernetes-xenial Release
  404  Not Found
```

The old Google-hosted apt repository has been deprecated. Migrate to the new `pkgs.k8s.io` source:

```bash
# Remove old source and key
sudo apt-mark unhold kubeadm kubelet kubectl || true
sudo apt remove -y kubeadm kubelet kubectl
sudo rm -f /etc/apt/sources.list.d/kubernetes.list
sudo rm -f /etc/apt/keyrings/kubernetes-apt-keyring.gpg

# Create keyrings directory
sudo mkdir -p /etc/apt/keyrings

# Add new pkgs.k8s.io source (v1.28 to match KWeaver's requirement)
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.28/deb/Release.key \
  | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.28/deb/ /' \
  | sudo tee /etc/apt/sources.list.d/kubernetes.list

# Reinstall and pin
sudo apt update
sudo apt install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

### Cannot access from external network

If you cannot access `https://<node-ip>/studio` from external networks after deployment:

1. **Check security group/firewall settings** in your cloud provider console:
   - Open port **80** (HTTP) and **443** (HTTPS) for inbound traffic
   - Protocol: TCP
   - Source: `0.0.0.0/0` (or your specific IP range)
   - Action: Allow

2. **Verify service is running**:
   ```bash
   kubectl get pods -n kweaver
   kubectl get ingress -n kweaver
   ```

3. **Check if ports are listening**:
   ```bash
   sudo ss -tlnp | grep -E ':80 |:443 '
   ```

4. **Test local access**:
   ```bash
   curl -I http://localhost/studio
   ```

If local access works but external access fails, it's likely a security group/firewall configuration issue.

### View component logs

```bash
kubectl logs -n <namespace> <pod-name>
```

### Helm uninstall hook failure

If a service installation or upgrade fails because a `post-delete` hook job is stuck or exits with an error, you can bypass the hook and clean up the leftover resources manually:

```bash
# Bypass the failing post-delete hook
helm uninstall <release-name> -n <namespace> --no-hooks

# Check for leftover job/pod resources
kubectl get job,pod -n <namespace> | grep <release-name>

# Remove leftovers if they still exist
kubectl delete job -n <namespace> <release-name>-post-delete-job --ignore-not-found
kubectl delete pod -n <namespace> -l job-name=<release-name>-post-delete-job --ignore-not-found
```

After cleanup, retry the installation.

## 📄 License

[Apache License 2.0](../LICENSE.txt)

