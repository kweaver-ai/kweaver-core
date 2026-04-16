# Prerequisites

Before deploying KWeaver Core, prepare the host, network, and client tooling.

## Host requirements

Run installation as `root` or with `sudo`.

| Item | Minimum | Recommended |
| --- | --- | --- |
| OS | CentOS 8+, openEuler 23+ | CentOS 8+ |
| CPU | 16 cores | 16 cores |
| Memory | 48 GB | 64 GB |
| Disk | 200 GB | 500 GB |

## Host preparation (typical Linux)

```bash
# 1. Disable firewall (or open required ports per your policy)
systemctl stop firewalld && systemctl disable firewalld

# 2. Disable swap
swapoff -a && sed -i '/ swap / s/^/#/' /etc/fstab

# 3. SELinux permissive if needed
setenforce 0

# 4. Install container runtime (example: containerd)
# dnf install containerd.io   # adjust for your distro
```

Exact steps depend on your OS; see [deploy/README.md](../../../deploy/README.md) for the supported flow.

## Network access

The deploy scripts may need outbound access to mirrors and registries, for example:

| Domain | Purpose |
| --- | --- |
| `mirrors.aliyun.com` | RPM mirrors |
| `mirrors.tuna.tsinghua.edu.cn` | containerd RPM mirror |
| `registry.aliyuncs.com` | Kubernetes images |
| `swr.cn-east-3.myhuaweicloud.com` | KWeaver images |
| `repo.huaweicloud.com` | Helm binary |
| `kweaver-ai.github.io` | Helm chart repo |

## Client tooling (after deploy)

On your workstation (any OS with network access to the cluster):

- **kubectl** — optional but useful for health checks
- **kweaver CLI** — from [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk)

```bash
npm install -g @kweaver-ai/kweaver-sdk
# or: npx kweaver --help
```

- **curl** — for raw HTTP API calls

## Next steps

- [Deploy](deploy.md) — install and post-install verification
