# KWeaver Core Deploy

[中文](README.zh.md) | English

One-click deployment of **KWeaver Core** onto a single-node Kubernetes cluster.

This `deploy` directory provides scripts to install KWeaver Core along with its dependencies including Kubernetes, infrastructure services, and data services.

**Platforms:** **Linux** is the recommended and fully documented install target (`preflight.sh`, k3s or kubeadm, data services). **macOS** is **optional** for **local development only** (Docker + kind + `dev/mac.sh`); see **[Mac install (dev)](dev/README.md)** — not a substitute for Linux production installs.

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](../LICENSE.txt)

## Linux: default `k8s` (kubeadm) vs optional k3s

**`KUBE_DISTRO` defaults to `k8s`** (package-manager Kubernetes + **kubeadm** on a single node). **k3s** is an optional lighter single-node stack. If you install k3s with `deploy.sh k3s install`, keep **`preflight.sh`** / **`kweaver-core`** aligned by passing **`--distro=k3s`** **before** the module (or **`export KUBE_DISTRO=k3s`**) — otherwise `preflight` may flag a k3s/kubeadm mismatch and installs can disagree on bootstrap.

### kubeadm / `KUBE_DISTRO=k8s` (default)

Single-node kubeadm flow is **`bash ./deploy.sh k8s install`** (`deploy/scripts/services/k8s.sh`). Product modules reuse an existing cluster when `kubectl` already works (`ensure_k8s` skips reinstall), then **`ensure_platform_prerequisites`** installs the bundled **data-services** layer (MariaDB, Redis, Kafka, ZooKeeper, OpenSearch, …) before Core. **macOS kind** skips host kubeadm but **`kweaver-core install` still runs `ensure_data_services` first** (see macOS section). Legacy **`kubeadm`** is still accepted as an alias for **`k8s`**.

**`deploy.sh` global flags** (`--distro`, `-y`, `--force-upgrade`, `--config`, …) must appear **before** the module name. Correct: `bash ./deploy.sh --distro=k3s kweaver-core install --minimum`. Wrong: `bash ./deploy.sh kweaver-core install --minimum --distro=k3s` (that `--distro` is not read as a global option). Equivalent without moving flags: `export KUBE_DISTRO=k3s` then `bash ./deploy.sh kweaver-core install --minimum`.

```bash
bash ./deploy.sh k8s install
bash ./deploy.sh kweaver-core install --minimum
```

### k3s (optional — lightweight single-node)

Uses the upstream k3s installer (Traefik disabled; this stack still installs **ingress-nginx** for a consistent chart/accessAddress setup). Override `K3S_INSTALL_URL`, `INSTALL_K3S_VERSION`, or `INSTALL_K3S_MIRROR` if you need a mirror or air-gapped tuning.

```bash
cd kweaver-core/deploy

bash ./deploy.sh k3s install

# Align distro with k3s for preflight and platform bootstrap:
bash ./deploy.sh --distro=k3s kweaver-core install --minimum
# or: export KUBE_DISTRO=k3s && bash ./deploy.sh kweaver-core install --minimum
```

Check status: `bash ./deploy.sh k3s status` — remove: `bash ./deploy.sh k3s uninstall`.

### macOS (optional — local dev with kind)

**Use this only for Mac validation; for real installs use Linux above.** Local Kubernetes via **kind** — no `preflight.sh` / `k3s` on the Mac host. **`mac.sh` sets `KWEAVER_SKIP_PLATFORM_BOOTSTRAP`** (no host k3s/kubeadm bootstrap). **`kweaver-core install` now runs `ensure_data_services` first** — same Helm layer as **`data-services install`** (MariaDB, Redis, Kafka, ZooKeeper, OpenSearch); **`mac.sh` defaults `AUTO_INSTALL_INGRESS_NGINX=false`** so kind’s existing ingress is not duplicated. Set **`KWEAVER_SKIP_DATA_SERVICES_BUNDLE=true`** to skip bundled data installs (advanced / external infra). **`data-services install`** alone remains useful to pre-stage or refresh the data layer. **Apple Silicon:** kind nodes are **arm64**; use arm64/multi-arch images (see `dev/conf/mac-config.yaml`). **Step order:** [dev/README.md](dev/README.md).

```bash
cd deploy   # repository deploy/ directory
bash ./dev/mac.sh doctor
# optional: install missing tools via Homebrew — bash ./dev/mac.sh doctor --fix (or -y doctor --fix to skip confirm)
bash ./dev/mac.sh cluster up
bash ./dev/mac.sh kweaver-core install --minimum   # implies --minimum; bundled data-services first (same as data-services install)
# optional: bash ./dev/mac.sh data-services install   # only if you want the data layer without Core, or to refresh it
# optional: bash ./dev/mac.sh kweaver-core download
# optional: bash ./dev/mac.sh onboard
# add leading -y for non-interactive (deploy.sh / onboard)
```

Config defaults: `dev/conf/mac-config.yaml`. `kweaver-dip` is not wired in `mac.sh` (use Linux `deploy.sh`); `isf` / `etrino` (`vega`) are delegated to `deploy.sh` — see [dev/README.md](dev/README.md).

## 🚀 Quick Start

### Host prerequisites

Run install commands as `root` or through `sudo`.

```bash
# 1. Disable firewall
systemctl stop firewalld && systemctl disable firewalld

# 2. Disable swap
swapoff -a && sed -i '/ swap / s/^/#/' /etc/fstab

# 3. Set SELinux to permissive if needed
setenforce 0

# 4. Install containerd.io
dnf install containerd.io
```

### Install KWeaver Core

```bash
# 1. Clone the repository
git clone https://github.com/kweaver-ai/kweaver-core.git
cd kweaver-core/deploy

# 2. (Recommended) Pre-install host check / fix
sudo bash ./preflight.sh                # check-only (default)
sudo bash ./preflight.sh --fix          # check + interactive fixes
sudo bash ./preflight.sh --fix -y       # auto-approve every fix
sudo bash ./preflight.sh --list-fixes   # preview which fixes would run, no changes
sudo bash ./preflight.sh --help         # all flags (--role, --skip, --report, --output=json, …)
# Default checks match k8s/kubeadm; for single-node k3s use: sudo bash ./preflight.sh --distro=k3s
# (same as env KUBE_DISTRO=k3s — shared with deploy.sh)

# 3. Install KWeaver Core
# Minimum installation — recommended for first-time experience
bash ./deploy.sh kweaver-core install --minimum
# Default is kubeadm (k8s). For single-node k3s instead (--distro must be BEFORE kweaver-core):
# bash ./deploy.sh --distro=k3s kweaver-core install --minimum
# or: export KUBE_DISTRO=k3s && bash ./deploy.sh kweaver-core install --minimum
# Equivalent to:
# bash ./deploy.sh kweaver-core install --set auth.enabled=false --set businessDomain.enabled=false

# Full installation (includes auth & business-domain modules)
bash ./deploy.sh kweaver-core install

# The script will interactively prompt for the access address and auto-detect the API server address.

# Or specify addresses explicitly (skips interactive prompts):
#   --access_address       Address for clients to reach KWeaver services (can be IP or domain)
#   --api_server_address   IP bound to a local network interface for K8s API server (must be a real NIC IP)
bash ./deploy.sh kweaver-core install \
  --access_address=<your-ip> \
  --api_server_address=<your-ip>

# (Optional) Customize ingress ports (default 80/443):
export INGRESS_NGINX_HTTP_PORT=8080
export INGRESS_NGINX_HTTPS_PORT=8443

# 4. (Recommended) Post-install bootstrap
#    Registers an LLM + embedding (skips when already there), patches the BKN ConfigMap
#    only when the default actually changes, and on a full ISF install creates the business
#    user `test`, assigns every role from `kweaver-admin role list`, switches `kweaver` to
#    that user, and imports the Context Loader toolset.
sudo bash ./onboard.sh        # interactive
sudo bash ./onboard.sh -y     # non-interactive (uses defaults)
sudo bash ./onboard.sh --help # all flags (--config=models.yaml, --enable-bkn-search, --skip-context-loader, …)
```

> **Why `sudo`?** `onboard.sh` reads `$HOME/.kweaver-ai/config.yaml` (written by `sudo deploy.sh` into `/root/.kweaver-ai/`) and writes the `kweaver` auth token to `$HOME/.kweaver`. Running it without `sudo` falls back to the in-repo template `deploy/conf/config.yaml` and may resolve a different access URL. **macOS dev path** (`bash ./dev/mac.sh onboard`) does **not** need `sudo`. The script also prints this hint at startup; silence with `ONBOARD_SUDO_HINT_DISABLED=1`.

> Full preflight / onboard flow, ISF dual-CLI auth and Mermaid diagrams: see [help/en/install.md — Post-install: `onboard.sh`](../help/en/install.md#post-install-onboardsh).

> **`onboard.sh` runtime messages** are **English**; on ISF HTTP **401001017**, a **TTY** prompts (**Enter**=`auth change-password` default; **`o`**=OAuth browser). Chinese + English context: [`dev/README.md`](../dev/README.md#onboard-and-kweaver-admin-full-isf); product docs [`help/zh/install.md`](../help/zh/install.md) / [`help/en/install.md`](../help/en/install.md).

## 📋 Prerequisites

### System requirements

| Item | Minimum | Recommended |
| --- | --- | --- |
| OS | CentOS 8+, OpenEuler 23+ | CentOS 8+ |
| CPU | 16 cores | 16 cores |
| Memory | 48 GB | 64 GB |
| Disk | 200 GB | 500 GB |

### Network requirements

The deployment scripts need access to these domains:

| Domain | Purpose |
| --- | --- |
| `mirrors.aliyun.com` | RPM package mirrors |
| `mirrors.tuna.tsinghua.edu.cn` | `containerd.io` RPM mirror |
| `registry.aliyuncs.com` | Kubernetes component images |
| `swr.cn-east-3.myhuaweicloud.com` | KWeaver application image registry |
| `repo.huaweicloud.com` | Helm binary download |
| `kweaver-ai.github.io` | KWeaver Helm chart repository |
| `rancher-mirror.rancher.cn` | k3s install script / binary (k3s quickstart path; override with `K3S_INSTALL_URL`) |

## 📦 Deployment Model

`kweaver-core` is the product-level entrypoint in this repository. The install flow is:

1. Install or repair single-node Kubernetes, local-path storage, and ingress-nginx.
2. Install or repair data services: MariaDB, Redis, Kafka, ZooKeeper, and OpenSearch.
3. Deploy the KWeaver Core application charts.

The Core application layer includes charts for data services management, application deployment, and task orchestration.

## 🔧 Usage

### Recommended commands

```bash
# Install KWeaver Core
./deploy.sh kweaver-core install

# Show Core status
./deploy.sh kweaver-core status

# Uninstall Core
./deploy.sh kweaver-core uninstall

# Cluster and Pod status
kubectl get nodes
kubectl get pods -A
```

## 📁 Project Structure

```text
deploy/
├── deploy.sh                 # Main entry script
├── conf/                     # Bundled config and static manifests
├── release-manifests/        # Versioned release bill of materials
├── scripts/
│   ├── lib/                  # Common helper functions
│   ├── services/             # Product and dependency install scripts
│   └── sql/                  # Versioned SQL initialization scripts
└── .tmp/charts/              # Local chart cache generated by download
```

## 🗑️ Uninstall

`bash deploy.sh kweaver-core uninstall` removes only the Core application layer.

```bash
# Remove the Core application layer
./deploy.sh kweaver-core uninstall
```

`bash deploy.sh k8s reset` resets the Kubernetes cluster, including data services and core.

```bash
# Reset Kubernetes cluster
./deploy.sh k8s reset
```

## 🔍 Troubleshooting

### CoreDNS is not ready

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

### Kubernetes apt / yum source missing or 404

`preflight.sh --check-only` reports one of these in **strict mode** (default):

```text
[FAIL] Deprecated Kubernetes apt source detected (packages.cloud.google.com) ...
[FAIL] apt has no install candidate for kubeadm — Kubernetes apt source missing or unreachable.
[FAIL] dnf/yum has no install candidate for kubeadm — Kubernetes yum repo missing or unreachable.
```

**Recommended fix (one command):**

```bash
sudo bash deploy/preflight.sh --fix --fix-allow=k8s-pkgs-repo
# or, to also pre-stage containerd / helm / Node, etc.:
sudo bash deploy/preflight.sh --fix -y
```

`preflight --fix → k8s-pkgs-repo` covers **both** scenarios (legacy alias `k8s-apt-source` still works in `--fix-allow`):

- Legacy `packages.cloud.google.com` source present → migrate it to `pkgs.k8s.io`.
- No K8s source configured at all → write `/etc/apt/sources.list.d/kubernetes.list` (or `/etc/yum.repos.d/kubernetes.repo`) pointing to `pkgs.k8s.io/core:/stable:/<vX.Y>/deb|rpm/`.

Pin a specific minor with `PREFLIGHT_K8S_APT_MINOR=v1.28` (default: detected from installed `kubeadm`, falls back to `v1.28`).

**Manual fallback (Ubuntu/Debian):**

```bash
sudo apt-mark unhold kubeadm kubelet kubectl || true
sudo apt remove -y kubeadm kubelet kubectl
sudo rm -f /etc/apt/sources.list.d/kubernetes.list
sudo rm -f /etc/apt/keyrings/kubernetes-apt-keyring.gpg
sudo mkdir -p /etc/apt/keyrings

curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.28/deb/Release.key \
  | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.28/deb/ /' \
  | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt update
sudo apt install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

**Manual fallback (RHEL/CentOS/openEuler):**

```bash
sudo tee /etc/yum.repos.d/kubernetes.repo > /dev/null <<'EOF'
[kubernetes]
name=Kubernetes
baseurl=https://pkgs.k8s.io/core:/stable:/v1.28/rpm/
enabled=1
gpgcheck=1
gpgkey=https://pkgs.k8s.io/core:/stable:/v1.28/rpm/repodata/repomd.xml.key
exclude=kubelet kubeadm kubectl cri-tools kubernetes-cni
EOF

sudo dnf install -y --disableexcludes=kubernetes kubeadm kubelet kubectl   # or: yum
```

### `containerd` cannot be installed (no `containerd.io` candidate)

On stock Ubuntu (no Docker CE repo) `preflight.sh` reports:

```text
[FAIL] apt has no install candidate for containerd.io OR containerd.
[FAIL] containerd not found ...
```

`preflight --fix → containerd-install` now tries `containerd.io` first (Docker CE repo) and falls back to the distro `containerd` package automatically:

```bash
sudo bash deploy/preflight.sh --fix --fix-allow=containerd-install
```

If both fail, the apt/yum source itself is broken — fix `apt-get update` / `dnf repolist` first.

### Strict mode and `--lenient`

`preflight.sh` defaults to **strict mode** (`PREFLIGHT_STRICT=true`). Items that block install AND are auto-fixable by `--fix` are reported as `[FAIL]` (so `--check-only` exits `1`), not `[WARN]`:

- `swap`, `net.ipv4.ip_forward`, `br_netfilter` / `overlay` modules, `bridge-nf-call-*`
- `vm.max_map_count`, `fs.inotify.*`, `ulimit -n soft`
- `containerd` not found / socket missing, `kubectl`, `helm`, `overlay` fs
- broken `apt-get update`, `dnf/yum repolist` failures, missing kubeadm/containerd install candidate

To downgrade these back to `[WARN]` (e.g. on a low-spec lab box where you accept the risk):

```bash
sudo bash deploy/preflight.sh --check-only --lenient
# equivalent to: PREFLIGHT_STRICT=false PREFLIGHT_STRICT_SOURCES=false sudo bash deploy/preflight.sh
```

### View component logs

```bash
kubectl logs -n <namespace> <pod-name>
```

## 📄 License

[Apache License 2.0](../LICENSE)
