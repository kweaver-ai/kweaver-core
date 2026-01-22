# KWeaver Deploy

一键部署 KWeaver AI 平台到单节点 Kubernetes 集群。

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

## 🚀 Quick Start

```bash
# 1. 克隆仓库
git clone https://github.com/kweaver-ai/kweaver.git
cd kweaver/deploy

# 2. 编辑配置文件（可选，使用默认配置可跳过）
# vim conf/config.yaml

# 3. 一键部署所有组件
bash ./deploy.sh full init
```

部署完成后，访问 `https://<节点IP>` 即可使用。

## 📋 Prerequisites

### 系统要求

| 项目 | 最低配置 | 推荐配置 |
|------|---------|---------|
| OS | CentOS 7/8+, RHEL 8 | CentOS 8 |
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

# 4. 手动安装  container-selinux
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
# 完整一键部署（推荐）
./deploy.sh full init     # 基础设施 + KWeaver 应用服务

# 分层部署
./deploy.sh infra init    # 仅基础设施：K8s + 数据服务
./deploy.sh kweaver init  # 仅应用服务：ISF/Studio/Ontology 等

# 部署单个基础设施组件
./deploy.sh k8s init      # Kubernetes 集群
./deploy.sh mariadb init  # MariaDB
./deploy.sh mongodb init  # MongoDB
./deploy.sh redis init    # Redis
./deploy.sh kafka init    # Kafka
./deploy.sh opensearch init  # OpenSearch

# 部署单个应用服务
./deploy.sh isf init      # ISF 服务
./deploy.sh studio init   # Studio 服务

# 指定 Helm 仓库和版本
./deploy.sh kweaver init --helm_repo=https://kweaver-ai.github.io/helm-repo/ --version=0.1.0

# 支持多种版本类型
./deploy.sh kweaver init --version=0.1.0                    # 稳定版
./deploy.sh kweaver init --version=0.0.0-feature-xxx        # 分支/开发版
./deploy.sh kweaver init                                     # 最新版

# 查看帮助
./deploy.sh --help
```

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

### 查看组件日志

```bash
kubectl logs -n <namespace> <pod-name>
```

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

## 📄 License

[Apache License 2.0](LICENSE)
