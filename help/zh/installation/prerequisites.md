# 环境要求

部署 KWeaver Core 前，请准备好主机、网络与客户端工具。

## 主机要求

安装过程需以 `root` 或 `sudo` 执行。

| 项 | 最低 | 推荐 |
| --- | --- | --- |
| 操作系统 | CentOS 8+、openEuler 23+ | CentOS 8+ |
| CPU | 16 核 | 16 核 |
| 内存 | 48 GB | 64 GB |
| 磁盘 | 200 GB | 500 GB |

## 主机准备（典型 Linux）

```bash
# 1. 关闭防火墙（或按策略放行端口）
systemctl stop firewalld && systemctl disable firewalld

# 2. 关闭 swap
swapoff -a && sed -i '/ swap / s/^/#/' /etc/fstab

# 3. 按需将 SELinux 设为 permissive
setenforce 0

# 4. 安装容器运行时（示例：containerd）
# dnf install containerd.io   # 按发行版调整
```

具体步骤因发行版而异，完整流程见 [deploy/README.zh.md](../../../deploy/README.zh.md)。

## 网络访问

部署脚本可能需要访问外网镜像与仓库，例如：

| 域名 | 用途 |
| --- | --- |
| `mirrors.aliyun.com` | RPM 镜像 |
| `mirrors.tuna.tsinghua.edu.cn` | containerd RPM 镜像 |
| `registry.aliyuncs.com` | Kubernetes 镜像 |
| `swr.cn-east-3.myhuaweicloud.com` | KWeaver 镜像 |
| `repo.huaweicloud.com` | Helm 二进制 |
| `kweaver-ai.github.io` | Helm Chart 仓库 |

## 部署后的客户端工具

在能访问集群的工作机上：

- **kubectl** — 可选，用于健康检查
- **kweaver CLI** — 来自 [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk)

```bash
npm install -g @kweaver-ai/kweaver-sdk
# 或: npx kweaver --help
```

- **curl** — 直接调用 HTTP API

## 下一步

- [部署](deploy.md) — 安装与安装后检查
