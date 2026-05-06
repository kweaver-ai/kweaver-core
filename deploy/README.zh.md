# KWeaver Core Deploy

中文 | [English](README.md)

一键将 **KWeaver Core** 部署到单节点 Kubernetes 集群。

这个 `deploy` 目录提供脚本安装 KWeaver Core 及其依赖，包括 Kubernetes、基础设施服务和数据服务。

**平台说明：** **Linux** 是推荐且文档最完整的目标环境（`preflight.sh`、k3s/kubeadm、数据服务等）。**macOS** 仅作**本机开发/验证**可选方案（Docker + kind + `dev/mac.sh`），详见 **[Mac 安装（开发向）](dev/README.md)**，**不能**替代 Linux 上的生产安装。

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](../LICENSE.txt)

## Linux：默认 `k8s`（kubeadm）与可选 k3s

**`KUBE_DISTRO` 默认为 `k8s`**（包管理安装 Kubernetes + 单节点 **kubeadm**）。**k3s** 为可选的更轻的单节点栈。若已用 `deploy.sh k3s install` 装好集群，后续的 **`preflight.sh`** / **`kweaver-core`** 请保持 distro 一致：在子模块名前加 **`--distro=k3s`**，或 **`export KUBE_DISTRO=k3s`**，否则 `preflight` 可能报 k3s 与 kubeadm 路径不一致，bootstrap 行为也容易对不上。

### kubeadm / `KUBE_DISTRO=k8s`（默认）

单节点 kubeadm 流程为 **`bash ./deploy.sh k8s install`**（`deploy/scripts/services/k8s.sh`）。若 `kubectl` 已可用，`ensure_k8s` 会跳过重复安装；随后 **`ensure_platform_prerequisites`** 会安装随平台一起交付的 **data-services**（MariaDB、Redis、Kafka、ZooKeeper、OpenSearch 等），再装 Core。**macOS kind** 不写宿主机 kubeadm：**`KWEAVER_SKIP_PLATFORM_BOOTSTRAP` 下，`kweaver-core install` 会先跑与 `data-services install` 相同的 Helm 数据层**，见下文 macOS。历史写法 **`kubeadm`** 仍可作为 **`k8s`** 的别名。

**`deploy.sh` 全局参数**（`--distro`、`-y`、`--force-upgrade`、`--config` 等）必须写在**子模块名之前**。正确：`bash ./deploy.sh --distro=k3s kweaver-core install --minimum`。错误：`bash ./deploy.sh kweaver-core install --minimum --distro=k3s`（末尾的 `--distro` 不会按全局参数解析）。不想改命令顺序时可用：`export KUBE_DISTRO=k3s` 再执行 `bash ./deploy.sh kweaver-core install --minimum`。

```bash
bash ./deploy.sh k8s install
bash ./deploy.sh kweaver-core install --minimum
```

### k3s（可选 — 轻量单节点）

使用官方 k3s 安装脚本（禁用 Traefik；仍会安装 **ingress-nginx** 以保持与现有 chart/accessAddress 一致）。可通过 `K3S_INSTALL_URL`、`INSTALL_K3S_VERSION`、`INSTALL_K3S_MIRROR` 等环境变量切换镜像或版本。

```bash
cd kweaver-core/deploy

bash ./deploy.sh k3s install

# 与 k3s 对齐 distro，供 preflight 与平台 bootstrap 使用：
bash ./deploy.sh --distro=k3s kweaver-core install --minimum
# 或：export KUBE_DISTRO=k3s && bash ./deploy.sh kweaver-core install --minimum
```

查看状态：`bash ./deploy.sh k3s status`；卸载：`bash ./deploy.sh k3s uninstall`。

### macOS（可选 — 本机 kind 开发）

**仅供 Mac 上做验证；正式安装请以本文 Linux 章节为准。** 本机用 **kind** 起 Kubernetes，不在 Mac 上跑 `preflight.sh` / `k3s install`。**`mac.sh` 设置 `KWEAVER_SKIP_PLATFORM_BOOTSTRAP`**。**`kweaver-core install` 会先执行 `ensure_data_services`**（与单独跑 `data-services install` 一致：MariaDB、Redis、Kafka、Zookeeper、OpenSearch）；**`mac.sh` 默认 `AUTO_INSTALL_INGRESS_NGINX=false`**，避免重复装 ingress。需要跳过自带数据层时使用 **`KWEAVER_SKIP_DATA_SERVICES_BUNDLE=true`**（高级用法 / 外接中间件）。仍可单独执行 **`data-services install`** 只做数据层或刷新。**Apple Silicon：** kind 节点为 **arm64**；**步骤见 [dev/README.md](dev/README.md)。**

```bash
cd deploy   # 仓库的 deploy/ 目录
bash ./dev/mac.sh doctor
# 可选：用 Homebrew 补全缺失工具 — bash ./dev/mac.sh doctor --fix（或 -y doctor --fix 跳过确认）
bash ./dev/mac.sh cluster up
bash ./dev/mac.sh kweaver-core install --minimum   # 默认带 --minimum；前置自动装 data-services（与 data-services install 相同）
# 可选：bash ./dev/mac.sh data-services install   # 仅数据层 / 刷新
# 可选：bash ./dev/mac.sh kweaver-core download
# 可选：bash ./dev/mac.sh onboard；需非交互时在命令前加 -y
```

默认配置：`dev/conf/mac-config.yaml`。`kweaver-dip` 未在 `mac.sh` 接入（请用 Linux `deploy.sh`）；`isf` / `etrino`（`vega`）会转调 `deploy.sh` —— 见 [dev/README.md](dev/README.md)。

## 🚀 Quick Start

### 主机前置条件

安装命令需要以 `root` 用户执行，或通过 `sudo` 执行。

```bash
# 1. 关闭防火墙
systemctl stop firewalld && systemctl disable firewalld

# 2. 关闭 Swap
swapoff -a && sed -i '/ swap / s/^/#/' /etc/fstab

# 3. 调整 SELinux（脚本可处理，但建议预先设为宽松）
setenforce 0

# 4. 安装 containerd.io
dnf install containerd.io
```

### 安装 KWeaver Core

```bash
# 1. 克隆仓库
git clone https://github.com/kweaver-ai/kweaver-core.git
cd kweaver-core/deploy

# 2.（推荐）装机前体检 / 修复
sudo bash ./preflight.sh                # 仅检查（默认）
sudo bash ./preflight.sh --fix          # 检查 + 交互修复
sudo bash ./preflight.sh --fix -y       # 全部自动确认修复
sudo bash ./preflight.sh --list-fixes   # 预览将会执行哪些修复，不改任何东西
sudo bash ./preflight.sh --help         # 全部参数（--role、--skip、--report、--output=json 等）
# 默认体检对齐 k8s/kubeadm；走单节点 k3s 时用：sudo bash ./preflight.sh --distro=k3s
#（与 deploy 共用环境变量 KUBE_DISTRO=k3s）

# 3. 安装 KWeaver Core
# 最小化安装 — 首次体验推荐
bash ./deploy.sh kweaver-core install --minimum
# 默认走 kubeadm（k8s）。若改用单节点 k3s（--distro 须写在 kweaver-core 之前）：
# bash ./deploy.sh --distro=k3s kweaver-core install --minimum
# 或：export KUBE_DISTRO=k3s && bash ./deploy.sh kweaver-core install --minimum
# 等价于:
# bash ./deploy.sh kweaver-core install --set auth.enabled=false --set businessDomain.enabled=false

# 完整安装（包含 auth 和 business-domain 模块）
bash ./deploy.sh kweaver-core install

# 脚本会交互式提示输入访问地址，并自动检测 API Server 地址。

# 或显式指定地址（跳过交互提示）：
#   --access_address       客户端访问 KWeaver 服务的地址（可以是 IP 或域名）
#   --api_server_address   K8s API Server 绑定的本机网卡 IP（必须是真实的网卡地址）
bash ./deploy.sh kweaver-core install \
  --access_address=<你的IP> \
  --api_server_address=<你的IP>

# （可选）自定义 ingress 端口（默认 80/443）：
export INGRESS_NGINX_HTTP_PORT=8080
export INGRESS_NGINX_HTTPS_PORT=8443

# 4.（推荐）安装后引导
#    注册 LLM + embedding（已有则跳过）；只有当默认 embedding 实际变更时才会 patch BKN ConfigMap；
#    在 ISF 全量下还会创建业务用户 `test`、把 `kweaver-admin role list` 中所有角色都挂上、
#    切换 `kweaver` 到该用户身份，并导入 Context Loader 工具集。
sudo bash ./onboard.sh        # 交互模式
sudo bash ./onboard.sh -y     # 非交互模式（按默认）
sudo bash ./onboard.sh --help # 全部参数（--config=models.yaml、--enable-bkn-search、--skip-context-loader 等）
```

> **为什么要 `sudo`？** `onboard.sh` 会读 `$HOME/.kweaver-ai/config.yaml`（由 `sudo deploy.sh` 写到 `/root/.kweaver-ai/` 下）并把 `kweaver` 认证 token 写到 `$HOME/.kweaver`。不加 `sudo` 会回退到仓库内模板 `deploy/conf/config.yaml`，可能解析出和安装时不一致的 access URL。**macOS 开发路径**（`bash ./dev/mac.sh onboard`）**不需要** `sudo`。脚本启动时也会打印这条提示；可用 `ONBOARD_SUDO_HINT_DISABLED=1` 关闭。

> 完整的 preflight / onboard 流程、ISF 双 CLI 鉴权与 Mermaid 流程图见 [help/zh/install.md — Post-install：`onboard.sh`](../help/zh/install.md#post-installonboardsh安装后引导)。

> **`onboard.sh` 终端输出为英文**；ISF HTTP **401001017** 且 **stdin/stdout 为 TTY** 时脚本会**先询问**：（**默认回车**）**`auth change-password`**；（**o / oauth**）浏览器 **`auth login` -k**。说明见 [`dev/README.md`](../dev/README.md)。产品文档 [`help/zh/install.md`](../help/zh/install.md)、[`help/en/install.md`](../help/en/install.md)。

## 📋 Prerequisites

### 系统要求

| 项目 | 最低配置 | 推荐配置 |
| --- | --- | --- |
| OS | CentOS 8+, OpenEuler 23+ | CentOS 8+ |
| CPU | 16 核 | 16 核 |
| 内存 | 48 GB | 64 GB |
| 磁盘 | 200 GB | 500 GB |

### 网络要求

部署脚本需要访问以下域名：

| 域名 | 用途 |
| --- | --- |
| `mirrors.aliyun.com` | RPM 软件包源 |
| `mirrors.tuna.tsinghua.edu.cn` | `containerd.io` RPM 源 |
| `registry.aliyuncs.com` | Kubernetes 组件镜像 |
| `swr.cn-east-3.myhuaweicloud.com` | KWeaver 应用镜像仓库 |
| `repo.huaweicloud.com` | Helm 二进制下载 |
| `kweaver-ai.github.io` | KWeaver Helm Chart 仓库 |
| `rancher-mirror.rancher.cn` | k3s 安装脚本/二进制（k3s 快速路径；可用 `K3S_INSTALL_URL` 覆盖） |

## 📦 部署模型

`kweaver-core` 是这个 `deploy` 目录里的产品入口，安装链路如下：

1. 安装或补齐单节点 Kubernetes、local-path storage、ingress-nginx。
2. 安装或补齐数据服务：MariaDB、Redis、Kafka、ZooKeeper、OpenSearch。
3. 部署 KWeaver Core 应用层 chart。

Core 应用层包括数据服务管理、应用部署和任务编排相关的 chart。



## 🔧 Usage

### 推荐命令

```bash
# 安装 KWeaver Core（推荐入口）
./deploy.sh kweaver-core install

# 查看 Core 状态
./deploy.sh kweaver-core status

# 卸载 Core
./deploy.sh kweaver-core uninstall

# 集群与 Pod 状态
kubectl get nodes
kubectl get pods -A
```

## 📁 Project Structure

```text
deploy/
├── deploy.sh                 # 主入口脚本
├── conf/                     # 内置配置与静态清单
├── release-manifests/        # 按版本组织的发布物料
├── scripts/
│   ├── lib/                  # 公共函数
│   ├── services/             # 各产品与依赖服务安装脚本
│   └── sql/                  # 按版本组织的 SQL 初始化脚本
└── .tmp/charts/              # download 命令生成的本地 chart 缓存
```

## 🗑️ Uninstall

`bash deploy.sh kweaver-core uninstall` 只卸载 Core 应用层。

```bash
# 1. 卸载 Core 应用层
./deploy.sh kweaver-core uninstall

```
`bash deploy.sh k8s reset` 重置 Kubernetes 集群，包括数据服务和core。

```bash
# 重置 Kubernetes 集群
./deploy.sh k8s reset
```

## 🔍 Troubleshooting

### CoreDNS 不就绪

```bash
# 检查防火墙是否关闭
systemctl status firewalld

# 重启 CoreDNS
kubectl -n kube-system delete pod -l k8s-app=kube-dns
```

### Pod 拉取镜像失败

```bash
# 检查网络连通性
curl -I https://swr.cn-east-3.myhuaweicloud.com

# 检查 containerd 配置
cat /etc/containerd/config.toml
```

### Kubernetes apt / yum 源缺失或 404

`preflight.sh --check-only` 在**严格模式**（默认）下会报：

```text
[FAIL] Deprecated Kubernetes apt source detected (packages.cloud.google.com) ...
[FAIL] apt has no install candidate for kubeadm — Kubernetes apt source missing or unreachable.
[FAIL] dnf/yum has no install candidate for kubeadm — Kubernetes yum repo missing or unreachable.
```

**推荐修复（一条命令搞定）：**

```bash
sudo bash deploy/preflight.sh --fix --fix-allow=k8s-pkgs-repo
# 也可以一次性把 containerd / helm / Node 等全准备好：
sudo bash deploy/preflight.sh --fix -y
```

`preflight --fix → k8s-pkgs-repo`（旧文档中的 `k8s-apt-source` 仍为 `--fix-allow` 别名）同时覆盖**两种**情况：

- 检测到旧的 `packages.cloud.google.com` 源 → 自动迁移到 `pkgs.k8s.io`。
- 完全没配置 K8s 源 → 直接写入 `/etc/apt/sources.list.d/kubernetes.list`（或 `/etc/yum.repos.d/kubernetes.repo`），指向 `pkgs.k8s.io/core:/stable:/<vX.Y>/deb|rpm/`。

可用 `PREFLIGHT_K8S_APT_MINOR=v1.28` 锁定特定 minor 版本（默认从已安装的 `kubeadm` 推断，回退 `v1.28`）。

**手动备选（Ubuntu/Debian）：**

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

**手动备选（RHEL/CentOS/openEuler）：**

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

sudo dnf install -y --disableexcludes=kubernetes kubeadm kubelet kubectl   # 或者 yum
```

### `containerd` 装不上（没有 `containerd.io` 候选）

原版 Ubuntu 默认不带 Docker CE 源，preflight 会报：

```text
[FAIL] apt has no install candidate for containerd.io OR containerd.
[FAIL] containerd not found ...
```

`preflight --fix → containerd-install` 现在会先试 `containerd.io`（Docker CE 源），失败时**自动回退**到发行版自带的 `containerd` 包：

```bash
sudo bash deploy/preflight.sh --fix --fix-allow=containerd-install
```

如果两者都失败，说明 apt/yum 源本身有问题——先把 `apt-get update` / `dnf repolist` 修好。

### 严格模式与 `--lenient`

`preflight.sh` 默认开启**严格模式**（`PREFLIGHT_STRICT=true`）。下面这些「会阻塞 install 且 `--fix` 能搞定」的项会报 `[FAIL]`（导致 `--check-only` 以退出码 `1` 退出），不再是 `[WARN]`：

- `swap`、`net.ipv4.ip_forward`、`br_netfilter` / `overlay` 内核模块、`bridge-nf-call-*`
- `vm.max_map_count`、`fs.inotify.*`、`ulimit -n soft`
- `containerd` 未安装 / socket 缺失、`kubectl`、`helm`、`overlay` 文件系统
- `apt-get update` 失败、`dnf/yum repolist` 失败、kubeadm / containerd 没有安装候选

如果你**确实**接受风险（比如只是 lab 上的小机器跑个体验），可以临时降回 `[WARN]`：

```bash
sudo bash deploy/preflight.sh --check-only --lenient
# 等价于 PREFLIGHT_STRICT=false PREFLIGHT_STRICT_SOURCES=false sudo bash deploy/preflight.sh
```

### 查看组件日志

```bash
kubectl logs -n <namespace> <pod-name>
```

## 📄 License

[Apache License 2.0](../LICENSE)
