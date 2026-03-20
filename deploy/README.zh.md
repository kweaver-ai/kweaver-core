# KWeaver Deploy

中文 | [English](README.md)

一键部署 KWeaver AI 平台到单节点 Kubernetes 集群。

[License](../LICENSE.txt)

## 🚀 Quick Start

```bash
# 方式 1：从 Release 包快速安装（推荐）
# 自动安装最新版本
curl -fsSL https://raw.githubusercontent.com/kweaver-ai/kweaver/main/install.sh | sh

# 安装指定版本
curl -fsSL https://raw.githubusercontent.com/kweaver-ai/kweaver/main/install.sh | sh -s -- --version v1.0.0

# 方式 2：克隆仓库
# 1. 克隆仓库
git clone https://github.com/kweaver-ai/kweaver.git
cd kweaver/deploy

# 2. 编辑配置文件（可选，使用默认配置可跳过）
# vim conf/config.yaml

# 3. 一键部署所有组件，默认安装最新版本
# 注意：deploy.sh 需要 root 权限
sudo bash ./deploy.sh full init
```

首次安装时，尤其是在云主机或网络较慢的环境中，完整部署过程可能需要较长时间。在某些环境下，完成安装可能会超过 1 小时。

部署完成后，访问 `https://<节点IP>/studio` 即可使用,账号admin，初始密码eisoo.com

## 📋 Prerequisites

### 权限要求

**`./deploy.sh` 需要 root 权限**，用于安装系统软件包、配置 Kubernetes 和管理集群资源。请使用 `sudo` 或以 root 用户运行：

```bash
sudo bash ./deploy.sh full init
# 或
sudo bash ./deploy.sh kweaver init
```

### 系统要求


| 项目  | 最低配置                | 推荐配置     |
| --- | ------------------- | -------- |
| OS  | CentOS 7/8+, RHEL 8 | CentOS 7 |
| CPU | 16 核                | 24 核     |
| 内存  | 48 GB               | 64 GB    |
| 磁盘  | 200 GB              | 500 GB   |


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


| 域名                                | 用途                      |
| --------------------------------- | ----------------------- |
| `mirrors.aliyun.com`              | RPM 软件包源                |
| `mirrors.tuna.tsinghua.edu.cn`    | 清华大学containerd.io RPM源  |
| `registry.aliyuncs.com`           | Kubernetes 组件镜像         |
| `swr.cn-east-3.myhuaweicloud.com` | 应用镜像仓库                  |
| `repo.huaweicloud.com`            | Helm 二进制文件              |
| `kweaver-ai.github.io`            | Kweaver 服务Helm Chart 仓库 |


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

**注意：** 所有 `deploy.sh` 命令都需要 root 权限。请使用 `sudo` 或以 root 用户运行。

```bash
# 完整一键部署（推荐）
sudo bash ./deploy.sh full init     # 基础设施 + KWeaver 应用服务

# 分层部署
sudo bash ./deploy.sh infra init    # 仅基础设施：K8s + 数据服务
sudo bash ./deploy.sh kweaver init  # 仅应用服务：ISF/Studio/Ontology 等

# 部署单个基础设施组件
sudo bash ./deploy.sh k8s init      # Kubernetes 集群
sudo bash ./deploy.sh mariadb init  # MariaDB
sudo bash ./deploy.sh mongodb init  # MongoDB
sudo bash ./deploy.sh redis init    # Redis
sudo bash ./deploy.sh kafka init    # Kafka
sudo bash ./deploy.sh opensearch init  # OpenSearch

# 部署单个应用服务
sudo bash ./deploy.sh isf init      # ISF 服务
sudo bash ./deploy.sh studio init   # Studio 服务

# 指定 Helm 仓库和版本
sudo bash ./deploy.sh kweaver init --helm_repo=https://kweaver-ai.github.io/helm-repo/ --version=0.1.0

# 支持多种版本类型
sudo bash ./deploy.sh kweaver init --version=0.1.0                    # 稳定版
sudo bash ./deploy.sh kweaver init --version=0.0.0-feature-xxx        # 分支/开发版
sudo bash ./deploy.sh kweaver init                                     # 最新版

# 查看帮助（无需 root）
./deploy.sh --help
```

### 验证部署

```bash
# 检查集群状态
kubectl get nodes
kubectl get pods -A

# 检查服务状态（无需 root）
./deploy.sh kweaver status
```

## ⚙️ Configuration

配置文件：`conf/config.yaml`

关键配置项：

```yaml
namespace: kweaver          # 部署命名空间
image:
  registry: swr.cn-east-3.myhuaweicloud.com/kweaver-ai  # 镜像仓库

accessAddress:
  host: <公网 IP 或域名>  # 云主机部署时必须设置
  port: 443
  scheme: https
  path: /

depServices:
  rds:
    source_type: internal   # internal=内置MariaDB, external=外部数据库
    host: 'mariadb.resource.svc.cluster.local'
    user: 'adp'
    password: ''            # 自动生成
```

如果部署在云主机上，务必在 `conf/config.yaml` 中将 `accessAddress.host` 设置为对外访问使用的公网 IP 或公网域名。若使用内网地址，安装完成后可能会出现访问失败。

**重要提示：** 在云主机上部署时，还需要在云服务商的安全组/防火墙中开放 **80** 和 **443** 端口，以允许外部访问。ingress-nginx 控制器会监听这些端口来提供 KWeaver 平台服务。

### 使用外部数据库

如果使用外部数据库，需要：

1. 将 `source_type` 改为 `external`
2. 配置外部数据库连接信息
3. 手动执行 SQL 初始化脚本（位于 `scripts/sql/` 目录）

### 场景化自动配置

`auto_cofig` 目录提供了部署后快速搭建演示场景（如供应链）的脚本。整个流程围绕统一配置文件 `config.env`：

1. **`setup_tem_db.sh`** 从 `conf/config.yaml` 读取数据库凭据，在集群内的 MariaDB 中创建演示数据库并导入示例数据（`dump-tem.sql`），完成后将连接信息（`DS_HOST`、`DS_USERNAME`、`DS_PASSWORD`）**回写**到 `config.env`。
2. **`auto_config.sh`** 读取 `config.env` 中的认证信息和数据源配置，依次创建数据源、导入知识网络、智能体、数据流等。

**前置条件：**

1. 登录系统工作台（`https://<节点IP>/deploy`，默认账号/密码：`admin/eisoo.com`）
2. 在 **信息安全管理 → 统一身份认证 → 账户 → 用户** 中新建用户 `test`
3. 在 **角色与访问策略 → 角色管理** 中，将 `test` 添加到数据管理员、AI管理员、应用管理员角色
4. 访问 Studio（`https://<节点IP>/studio`），使用 `test` 登录（默认密码：`123456`），按提示修改密码

**使用方法：**

```bash
cd deploy/auto_cofig

# 步骤 1：准备示例数据库（自动将数据库连接信息写入 config.env）
chmod +x setup_tem_db.sh auto_config.sh
./setup_tem_db.sh

# 步骤 2：导入场景配置
./auto_config.sh agent.json 供应链业务知识网络.json dataflow.json

# 步骤 3：导入工具集
./auto_config.sh --step 7 contextloader工具集_020.adp
./auto_config.sh --step 7 基础结构化数据分析工具箱2.adp
```

详细使用说明请参考 `auto_cofig/README.md`。

## 📁 Project Structure

```
deploy/
├── deploy.sh           # 主入口脚本
├── conf/
│   ├── config.yaml         # 部署配置文件
│   ├── kube-flannel.yml    # Flannel 网络配置
│   └── local-path-storage.yaml  # 本地存储配置
├── auto_cofig/             # 场景化自动配置
│   ├── auto_config.sh      # 自动配置脚本
│   ├── config.env          # 配置模板
│   ├── README.md           # 使用说明
│   └── *.json, *.adp       # 示例场景文件
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

**注意：** 卸载命令也需要 root 权限。

```bash
# 完整卸载
sudo bash ./deploy.sh full reset     # 卸载全部（应用服务 + 基础设施）

# 分层卸载
sudo bash ./deploy.sh kweaver uninstall  # 仅卸载应用服务
sudo bash ./deploy.sh infra reset        # 仅卸载基础设施

# 卸载单个组件
sudo bash ./deploy.sh mariadb uninstall
sudo bash ./deploy.sh k8s reset
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

### Helm 安装超时（context deadline exceeded）

如果安装服务（特别是 ISF 服务如 `sharemgnt-single`）时出现 `UPGRADE FAILED: context deadline exceeded` 错误，默认超时时间可能对慢速网络或大镜像来说太短。

**解决方案：** 在运行部署前通过环境变量增加超时时间：

```bash
# ISF 服务专用（默认：helm 600s，command 900s）
export ISF_HELM_TIMEOUT=900s
export ISF_COMMAND_TIMEOUT=1200
sudo bash ./deploy.sh isf init

# 全局覆盖（所有服务）
export HELM_INSTALL_TIMEOUT=600s
export HELM_COMMAND_TIMEOUT=900
sudo bash ./deploy.sh kweaver init

# Kafka（已有更长默认值：1800s）
export KAFKA_HELM_TIMEOUT=2400s
sudo bash ./deploy.sh kafka init
```

**诊断：**
```bash
# 检查 Pod 状态和事件
kubectl get pods -n <namespace> | grep <service-name>
kubectl describe pod -n <namespace> <pod-name>
kubectl get events -n <namespace> --sort-by='.lastTimestamp' | tail -20
```

### Pod CrashLoopBackOff（容器崩溃循环）

如果 Pod 处于 `CrashLoopBackOff` 状态（不断重启），说明容器启动后持续崩溃。常见原因包括：

1. **配置错误**（缺少或无效的配置值）
2. **数据库连接失败**（数据库未就绪或凭证错误）
3. **缺少依赖服务**（其他服务不可用）
4. **资源限制**（内存/CPU 约束）
5. **应用错误**（检查应用日志）

**诊断：**
```bash
# 1. 检查 Pod 状态和详细信息（显示事件和容器状态）
kubectl get pod <pod-name> -n <namespace> -o wide
kubectl describe pod <pod-name> -n <namespace>

# 2. 查看当前容器日志
kubectl logs <pod-name> -n <namespace> --tail=100

# 3. 查看上一个容器的日志（崩溃前的容器）
kubectl logs <pod-name> -n <namespace> --previous --tail=100

# 4. 查看命名空间最近的事件
kubectl get events -n <namespace> --sort-by='.lastTimestamp' | tail -30

# 5. 验证依赖服务是否就绪（如数据库）
kubectl get pods -n <namespace> | grep -E 'mariadb|mongodb|mysql'
kubectl get svc -n <namespace> | grep -E 'mariadb|mongodb|mysql'

# 6. 检查 Helm Release 状态
helm status <release-name> -n <namespace>
```

**常见修复方法：**
- **数据库未就绪**：等待数据库 Pod 进入 `Running` 状态后再安装依赖服务
- **数据库凭证错误**：检查 `config.yaml` 中的数据库连接配置
- **缺少环境变量**：验证 Helm values 和 config.yaml
- **资源限制**：检查 `kubectl describe pod` 中是否有 `OOMKilled` 或资源限制错误

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

# 创建 keyrings 目录
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

### 无法从外部网络访问

如果部署完成后无法从外部网络访问 `https://<节点IP>/studio`：

1. **检查云服务商的安全组/防火墙设置**：
   - 开放 **80**（HTTP）和 **443**（HTTPS）端口的入站流量
   - 协议：TCP
   - 源：`0.0.0.0/0`（或你的特定 IP 范围）
   - 动作：允许

2. **验证服务是否运行**：
   ```bash
   kubectl get pods -n kweaver
   kubectl get ingress -n kweaver
   ```

3. **检查端口是否在监听**：
   ```bash
   sudo ss -tlnp | grep -E ':80 |:443 '
   ```

4. **测试本地访问**：
   ```bash
   curl -I http://localhost/studio
   ```

如果本地访问正常但外部访问失败，通常是安全组/防火墙配置问题。

### 查看组件日志

```bash
kubectl logs -n <namespace> <pod-name>
```

### Helm 卸载 Hook 失败

如果某个服务在安装或升级过程中因为 `post-delete` hook Job 卡住或执行失败而报错，可以先绕过该 hook，并手动清理残留资源：

```bash
# 跳过失败的 post-delete hook
helm uninstall <release-name> -n <namespace> --no-hooks

# 检查是否还有残留 Job / Pod
kubectl get job,pod -n <namespace> | grep <release-name>

# 如仍有残留则手动删除
kubectl delete job -n <namespace> <release-name>-post-delete-job --ignore-not-found
kubectl delete pod -n <namespace> -l job-name=<release-name>-post-delete-job --ignore-not-found
```

清理完成后，再重新执行安装。

## 📄 License

[Apache License 2.0](../LICENSE.txt)