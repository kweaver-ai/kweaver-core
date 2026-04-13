# KWeaver Core Deploy

中文 | [English](README.md)

一键将 **KWeaver Core** 部署到单节点 Kubernetes 集群。

这个 `deploy` 目录提供脚本安装 KWeaver Core 及其依赖，包括 Kubernetes、基础设施服务和数据服务。

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](../LICENSE.txt)

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

# 2. 安装 KWeaver Core
# 脚本会交互式提示输入访问地址，并自动检测 API Server 地址。
bash ./deploy.sh kweaver-core install

# 或显式指定地址（跳过交互提示）：
#   --access_address       客户端访问 KWeaver 服务的地址（可以是 IP 或域名）
#   --api_server_address   K8s API Server 绑定的本机网卡 IP（必须是真实的网卡地址）
bash ./deploy.sh kweaver-core install \
  --access_address=<你的IP> \
  --api_server_address=<你的IP>

# （可选）自定义 ingress 端口（默认 80/443）：
export INGRESS_NGINX_HTTP_PORT=8080
export INGRESS_NGINX_HTTPS_PORT=8443
```

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

### Kubernetes apt 源 404（Ubuntu/Debian）

如果 `apt update` 报错，提示旧的 `packages.cloud.google.com` 仓库 404：

```text
Err:7 https://packages.cloud.google.com/apt kubernetes-xenial Release
  404  Not Found
```

旧版 Google 托管 apt 源已废弃，需要迁移到 `pkgs.k8s.io`：

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

### 查看组件日志

```bash
kubectl logs -n <namespace> <pod-name>
```

## 📄 License

[Apache License 2.0](../LICENSE)
