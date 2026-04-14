# 部署

KWeaver Core 通过本仓库 `deploy/` 目录下的 `deploy.sh` 安装。

## 克隆并进入 deploy 目录

```bash
git clone https://github.com/kweaver-ai/kweaver-core.git
cd kweaver-core/deploy
chmod +x deploy.sh
```

## 安装 KWeaver Core

### 最小化安装（首次体验推荐）

跳过部分可选模块（如认证、业务域），资源占用更小：

```bash
./deploy.sh kweaver-core install --minimum
```

等价写法：

```bash
./deploy.sh kweaver-core install --set auth.enabled=false --set businessDomain.enabled=false
```

### 完整安装

包含认证与业务域等相关组件：

```bash
./deploy.sh kweaver-core install
```

脚本可能交互式询问 **访问地址**，并自动探测 **API Server 地址**。

### 非交互安装

```bash
./deploy.sh kweaver-core install \
  --access_address=<你的IP或域名> \
  --api_server_address=<K8s API 绑定的网卡 IP>
```

- `--access_address` — 客户端访问 KWeaver（Ingress）所用的地址
- `--api_server_address` — Kubernetes API Server 绑定的真实网卡 IP

### 自定义 Ingress 端口（可选）

```bash
export INGRESS_NGINX_HTTP_PORT=8080
export INGRESS_NGINX_HTTPS_PORT=8443
./deploy.sh kweaver-core install
```

## 常用命令

```bash
./deploy.sh kweaver-core status
./deploy.sh kweaver-core uninstall
./deploy.sh --help
```

## 安装内容概览

1. 单节点 Kubernetes（如需要）、存储、Ingress
2. 数据服务：MariaDB、Redis、Kafka、ZooKeeper、OpenSearch（以发布清单为准）
3. KWeaver Core 应用 Helm Chart

卸载与集群重置见 [deploy/README.zh.md](../../../deploy/README.zh.md)。

## 下一步

- [验证安装](verify.md)
- [快速开始](../quick-start.md)
