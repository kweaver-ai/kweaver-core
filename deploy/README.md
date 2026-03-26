# KWeaver Deploy

[中文](README.zh.md) | English

One-click deployment of the KWeaver AI platform to a single-node Kubernetes cluster.

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](../LICENSE.txt)

## 🚀 Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/kweaver-ai/kweaver.git
cd kweaver/deploy

# 2. Edit the config file (optional; skip to use defaults)
# vim conf/config.yaml

# 3. Install KWeaver Core (includes ISF by default)
bash ./deploy.sh kweaver-core install

# 3'. Install KWeaver DIP (automatically installs missing dependencies)
# bash ./deploy.sh kweaver-dip install
```

After deployment, open `https://<node-ip>/studio`. Username: `admin`, initial password: `eisoo.com`.

## 📋 Prerequisites

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

```bash
# Recommended install paths
./deploy.sh kweaver-core install
# Install KWeaver Core; ISF is installed by default

./deploy.sh kweaver-core install --enable-isf=false
# Install KWeaver Core without ISF

./deploy.sh kweaver-dip install
# Install KWeaver DIP; if K8s, data services, ISF, or KWeaver Core are missing, they will be installed automatically

./deploy.sh core install
# Same as above; `core` is an alias of `kweaver-core`

./deploy.sh dip install
# Same as above; `dip` is an alias of `kweaver-dip`

# KWeaver Core examples
./deploy.sh kweaver-core install --config=/root/.kweaver-ai/config.yaml
# Use a specific config file

./deploy.sh kweaver-core install --helm_repo=https://acr.aishu.cn/chartrepo/public --version=0.4.0
# Install a specific version from a specific Helm repo

# Optional commands
./deploy.sh isf install
./deploy.sh config generate
./deploy.sh k8s install
./deploy.sh storage install
./deploy.sh mariadb install
./deploy.sh redis install
./deploy.sh kafka install
./deploy.sh zookeeper install
./deploy.sh opensearch install
./deploy.sh ingress-nginx install

# Status and uninstall
./deploy.sh isf status
./deploy.sh kweaver-core status
./deploy.sh kweaver-dip status
./deploy.sh kweaver uninstall
./deploy.sh kweaver-core uninstall
./deploy.sh isf uninstall
./deploy.sh kweaver-dip uninstall
./deploy.sh k8s reset

# Help
./deploy.sh
```

### Verify deployment

```bash
# Cluster status
kubectl get nodes
kubectl get pods -A

# Service status
./deploy.sh kweaver status
```

## ⚙️ Configuration

Config file: `conf/config.yaml`

Key settings:

```yaml
namespace: kweaver          # Namespace
image:
  registry: swr.cn-east-3.myhuaweicloud.com/kweaver-ai  # Image registry

depServices:
  rds:
    source_type: internal   # internal=embedded MariaDB, external=external DB
    host: 'mariadb.resource.svc.cluster.local'
    user: 'adp'
    password: ''            # Auto-generated
```

### Use an external database

If you use an external database:

1. Change `source_type` to `external`
2. Configure external DB connection settings
3. Manually run the SQL initialization scripts under `scripts/sql/`

## 📁 Project Structure

```
deploy/
├── deploy.sh           # Main entry script
├── conf/
│   ├── config.yaml              # Deployment config
│   ├── kube-flannel.yml         # Flannel network config
│   └── local-path-storage.yaml  # Local storage config
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

```bash
# Full uninstall
./deploy.sh full reset         # Uninstall everything (apps + infrastructure)

# Layered uninstall
./deploy.sh kweaver uninstall  # Uninstall application services only
./deploy.sh infra reset        # Uninstall infrastructure only

# Uninstall a single component
./deploy.sh mariadb uninstall
./deploy.sh k8s reset
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

### View component logs

```bash
kubectl logs -n <namespace> <pod-name>
```

## 📄 License

[Apache License 2.0](../LICENSE.txt)
