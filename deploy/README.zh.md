# KWeaver Deploy

中文 | [English](README.md)

一键部署 KWeaver AI 平台到单节点 Kubernetes 集群。

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](../LICENSE.txt)

## 🚀 Quick Start

```bash
# 1. 克隆仓库
git clone https://github.com/kweaver-ai/kweaver.git
cd kweaver/deploy

# 2. 编辑配置文件（可选，使用默认配置可跳过）
# vim conf/config.yaml

# 3. 安装 KWeaver Core（默认包含 ISF）
bash ./deploy.sh kweaver-core install

# 3'. 先预下载 chart 到 deploy/.tmp/charts，再显式从该目录安装
# bash ./deploy.sh kweaver-core download
# bash ./deploy.sh kweaver-core install --charts_dir=./.tmp/charts

# 3'. 安装 KWeaver DIP（会自动补齐依赖）
bash ./deploy.sh kweaver-dip install
```

部署完成后，访问 `https://<节点IP>/studio` 即可使用,账号admin，初始密码eisoo.com

## 📋 Prerequisites

### 系统要求

| 项目 | 最低配置 | 推荐配置 |
|------|---------|---------|
| OS | CentOS 7/8+, RHEL 8 | CentOS 7 |
| CPU | 16 核 | 24 核 |
| 内存 | 48 GB | 64 GB |
| 磁盘 | 200 GB | 500 GB |

### 前置条件（必须）

```bash
# 1. 关闭防火墙
systemctl stop firewalld && systemctl disable firewalld

# 2. 关闭 Swap
swapoff -a && sed -i '/ swap / s/^/#/' /etc/fstab

# 3. 关闭 SELinux（可选，脚本会自动处理）
setenforce 0

# 4. 手动安装  containerd.io
dnf install containerd.io
```

### 网络要求

部署脚本需要访问以下域名：

| 域名 | 用途 |
|------|------|
| `mirrors.aliyun.com` | RPM 软件包源 |
| `mirrors.tuna.tsinghua.edu.cn` | 清华大学containerd.io RPM源  |
| `registry.aliyuncs.com` | Kubernetes 组件镜像 |
| `swr.cn-east-3.myhuaweicloud.com` | 应用镜像仓库 |
| `repo.huaweicloud.com` | Helm 二进制文件 |
| `kweaver-ai.github.io` | Kweaver 服务Helm Chart 仓库 |

## 📦 Components

### 基础设施

- **Kubernetes** v1.28 (单节点)
- **containerd** v1.6+
- **Flannel CNI** v0.25.5
- **ingress-nginx** v1.14.1

### 数据服务

- **MariaDB** v11.4.7
- **MongoDB** v4.4.30
- **Redis** v7.4.6 (Sentinel)
- **Kafka** v3.9.0
- **OpenSearch** v2.19.4
- **ZooKeeper** v3.9.3

## 🔧 Usage

### 部署命令

```bash
# 推荐安装方式
./deploy.sh kweaver-core install
# 安装 KWeaver Core，默认会安装 ISF

./deploy.sh kweaver-core install --enable-isf=false
# 安装 KWeaver Core，但不安装 ISF

./deploy.sh kweaver-dip install
# 安装 KWeaver DIP；如果 K8s、数据服务、ISF 或 KWeaver Core 缺失，会自动补齐依赖

./deploy.sh kweaver-core download
# 预下载/更新 KWeaver Core chart 到 deploy/.tmp/charts；默认会同时处理 ISF chart

./deploy.sh kweaver-core download --charts_dir=/path/to/charts
# 预下载/更新 KWeaver Core chart 到指定本地目录

./deploy.sh kweaver-core download --enable-isf=false
# 仅下载 Core chart，不下载 ISF chart

./deploy.sh kweaver-core install --charts_dir=./.tmp/charts
# 从预下载的本地 chart 安装 KWeaver Core

./deploy.sh kweaver-dip download
# 预下载/更新 DIP + Core + ISF chart 到 deploy/.tmp/charts

./deploy.sh kweaver-dip download --charts_dir=/path/to/charts
# 预下载/更新 DIP + Core + ISF chart 到指定本地目录

./deploy.sh kweaver-dip install --charts_dir=./.tmp/charts
# 从预下载的本地 chart 安装 KWeaver DIP

./deploy.sh isf download --force-refresh
# 强制重新下载 ISF chart 到 deploy/.tmp/charts

./deploy.sh isf download --charts_dir=/path/to/charts
# 预下载/更新 ISF chart 到指定本地目录

./deploy.sh isf install --charts_dir=./.tmp/charts
# 从预下载的本地 chart 安装 ISF

./deploy.sh core install
# 同上，core 是 kweaver-core 的别名

./deploy.sh dip install
# 同上，dip 是 kweaver-dip 的别名

# kweaver-core 核心用法
./deploy.sh kweaver-core install --config=/root/.kweaver-ai/config.yaml
# 指定配置文件

./deploy.sh kweaver-core install --helm_repo=https://acr.aishu.cn/chartrepo/public --version=0.4.0
# 从指定 Helm 仓库安装指定版本

./deploy.sh kweaver-core download --helm_repo=https://acr.aishu.cn/chartrepo/public --version=0.4.0
# 从指定 Helm 仓库预下载指定版本 chart


# 其他可选命令
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

# 状态与卸载
./deploy.sh isf status
./deploy.sh kweaver-core status
./deploy.sh kweaver-dip status
./deploy.sh kweaver uninstall
./deploy.sh kweaver-core uninstall
./deploy.sh isf uninstall
./deploy.sh kweaver-dip uninstall
./deploy.sh k8s reset

# 查看帮助
./deploy.sh
```

### Chart 预下载与缓存

- 共享缓存目录默认是 `deploy/.tmp/charts`
- `download` 如果检测不到 `helm`，会先自动安装 `helm`
- `download` 默认增量刷新，不会每次全量重下
- 如果指定 `--version`，脚本只检查该版本是否已存在；不存在才下载
- 如果不指定 `--version`，脚本会比较 Helm repo 最新版本和本地缓存的最新版本，仅在 repo 更新时下载
- `kweaver-core download` 默认会连同 ISF 一起下载；可用 `--enable-isf=false` 关闭
- `kweaver-dip download` 会自动下载 DIP、KWeaver Core、ISF 的完整依赖 chart
- 只有 `download` 会创建或更新默认共享缓存目录 `deploy/.tmp/charts`
- `install` 不会自动读取 `deploy/.tmp/charts`；如果要使用预下载的本地 `.tgz`，请显式传入 `--charts_dir=<目录>`

### 验证部署

```bash
# 检查集群状态
kubectl get nodes
kubectl get pods -A

# 检查服务状态
./deploy.sh kweaver status
```

## ⚙️ Configuration

配置文件：`conf/config.yaml`

关键配置项：

```yaml
namespace: kweaver          # 部署命名空间
image:
  registry: swr.cn-east-3.myhuaweicloud.com/kweaver-ai  # 镜像仓库

depServices:
  rds:
    source_type: internal   # internal=内置MariaDB, external=外部数据库
    host: 'mariadb.resource.svc.cluster.local'
    user: 'adp'
    password: ''            # 自动生成
```

### 使用外部数据库

如果使用外部数据库，需要：

1. 将 `source_type` 改为 `external`
2. 配置外部数据库连接信息
3. 手动执行 SQL 初始化脚本（位于 `scripts/sql/` 目录）

## 📁 Project Structure

```
deploy/
├── deploy.sh           # 主入口脚本
├── conf/
│   ├── config.yaml         # 部署配置文件
│   ├── kube-flannel.yml    # Flannel 网络配置
│   └── local-path-storage.yaml  # 本地存储配置
└── scripts/
    ├── lib/
    │   └── common.sh       # 公共函数库
    ├── services/           # 各组件安装脚本
    │   ├── k8s.sh
    │   ├── mariadb.sh
    │   ├── mongodb.sh
    │   └── ...
    └── sql/                # SQL 初始化脚本
        ├── isf/
        ├── studio/
        └── ...
```

## 🗑️ Uninstall

```bash
# 完整卸载
./deploy.sh full reset     # 卸载全部（应用服务 + 基础设施）

# 分层卸载
./deploy.sh kweaver uninstall  # 仅卸载应用服务
./deploy.sh infra reset        # 仅卸载基础设施

# 卸载单个组件
./deploy.sh mariadb uninstall
./deploy.sh k8s reset
```

## 🔍 Troubleshooting

### CoreDNS 不就绪

```bash
# 检查防火墙是否关闭
systemctl status firewalld

# 手动重启 CoreDNS
kubectl -n kube-system delete pod -l k8s-app=kube-dns
```

### Pod 拉取镜像失败

```bash
# 检查网络连通性
curl -I https://swr.cn-east-3.myhuaweicloud.com

# 检查 containerd 配置
cat /etc/containerd/config.toml
```

### Kubernetes apt 源 404（Ubuntu/Debian）

如果 `apt update` 报错，提示旧的 `packages.cloud.google.com` 仓库 404：

```
Err:7 https://packages.cloud.google.com/apt kubernetes-xenial Release
  404  Not Found
```

旧版 Google 托管的 apt 源已废弃，需要迁移到新的 `pkgs.k8s.io` 源：

```bash
# 移除旧源和密钥
sudo apt-mark unhold kubeadm kubelet kubectl || true
sudo apt remove -y kubeadm kubelet kubectl
sudo rm -f /etc/apt/sources.list.d/kubernetes.list
sudo rm -f /etc/apt/keyrings/kubernetes-apt-keyring.gpg
sudo mkdir -p /etc/apt/keyrings

# 添加新的 pkgs.k8s.io 源（v1.28，与 KWeaver 要求一致）
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.28/deb/Release.key \
  | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.28/deb/ /' \
  | sudo tee /etc/apt/sources.list.d/kubernetes.list

# 重新安装并锁定版本
sudo apt update
sudo apt install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

### 查看组件日志

```bash
kubectl logs -n <namespace> <pod-name>
```

## 📄 License

[Apache License 2.0](../LICENSE.txt)
