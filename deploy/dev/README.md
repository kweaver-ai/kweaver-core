# macOS dev path (`dev/mac.sh`, kind)

**Audience:** Optional for **macOS developers** doing quick validation. **Production deployments and all primary documentation assume Linux** — start with [`deploy/README.md`](../README.md) and [`help/en/install.md`](../../help/en/install.md) / [`help/zh/install.md`](../../help/zh/install.md).

English | [中文](#中文说明)

Local Kubernetes with **kind** plus the same Helm charts as Linux `deploy.sh`. No host `preflight` / k3s / kubeadm.

### Repository (clone first)

Scripts and vendored manifests live in the repo tree — **`mac.sh` is not a standalone installer.** Clone **[kweaver-ai/kweaver-core](https://github.com/kweaver-ai/kweaver-core)** (check out the branch you deploy from, e.g. `feature/deploy/k3s-module`), then **`cd`** into **`deploy/`** before any command below:

```bash
git clone https://github.com/kweaver-ai/kweaver-core.git
cd kweaver-core/deploy   # always run bash ./dev/mac.sh ... from this directory
```

Same layout applies if your product tarball extracts to a **`kweaver-core/`** root with a **`deploy/`** subdirectory.

### Architecture (Apple Silicon / arm64)

On **Apple Silicon** Macs, kind nodes are **linux/arm64** by default. Charts pull from `image.registry` in your [`dev/conf/mac-config.yaml`](conf/mac-config.yaml) (copy from [`mac-config.yaml.example`](conf/mac-config.yaml.example)); those images must be **arm64-capable** (multi-arch manifest or an arm64 tag). If a registry only ships **amd64**, pods often fail with *exec format error*. Intel Macs still get **amd64** kind nodes unless you force another platform.

### Access URL (HTTP and automatic host)

- **HTTP vs HTTPS:** HTTPS uses TLS to encrypt traffic and verify the server identity; HTTP is unencrypted. On a trusted LAN, HTTP avoids dealing with local TLS certs and is typical for dev. Browsers may still show “Not secure” for HTTP — expected.
- **Automatic IP:** Your `mac-config.yaml` uses `accessAddress.scheme: http` and may **omit** `host` (see the example file). On `kweaver-core install`, the flow detects your LAN IP (on macOS, usually the default-route interface) and writes it into values so other devices on the network can open the UI. Set `accessAddress.host` yourself (for example `localhost`) if you want same-machine-only URLs.

## Order of operations

Run from the **`deploy/`** directory (`cd deploy` in this repo). Invoke **`mac.sh` with bash** (e.g. `bash ./dev/mac.sh ...`). **`kweaver-core` / `core`:** the wrapper **defaults to `--minimum`** (smaller chart set; skips ISF in manifest terms). Pass **`--full`** for the full manifest profile (adds ISF download/install when the manifest enables it).

| Step | Command | Required? |
|------|---------|-----------|
| 1 | `bash ./dev/mac.sh doctor` | Recommended |
| 2 | `bash ./dev/mac.sh doctor --fix` (or `-y doctor --fix`) | If something is missing |
| 3 | `bash ./dev/mac.sh cluster up` | **Yes** before install |
| 4 | `bash ./dev/mac.sh data-services install` | Optional — only to install/refresh **data layer alone**; **`kweaver-core install` invokes the same bundled install first** (`KWEAVER_SKIP_DATA_SERVICES_BUNDLE=true` skips it). |
| 5 | `bash ./dev/mac.sh kweaver-core download` | Optional (local chart cache; **minimum** by default) |
| 6 | `bash ./dev/mac.sh kweaver-core install` | **Yes** — deploy Core (**`--minimum` implied**); runs bundled data-services beforehand unless skipped |
| 7 | `bash ./dev/mac.sh onboard` | Optional (models/BKN; needs `kweaver` CLI; add `-y` to skip prompts) |

Optional (same `deploy.sh` Helm paths as Linux; you need a working cluster + values that match your dependencies): `bash ./dev/mac.sh isf install|download|uninstall|status`, `bash ./dev/mac.sh etrino install|...` (Vega stack; **`vega` is an alias of `etrino`**). ISF may require DB/config beyond the minimal mac sample—see Linux `deploy.sh` help and your `CONFIG_YAML_PATH`.

**Minimal path:** `cluster up` → `kweaver-core install` (wrapper implies `--minimum` and runs **data-services** first). If you skip that bundle (`KWEAVER_SKIP_DATA_SERVICES_BUNDLE=true`), you must provide reachable DB/Kafka/etc. yourself or run **`data-services install`** beforehand.

**Pause to save resources (keep the cluster):** Quit **Docker Desktop**. Kind uses Docker, so that stops the cluster without `kind delete`. Open Docker again when you want to keep working.

If Docker should stay up, stop only the kind **node** container(s) (same effect for the cluster; not `kind delete`):

```bash
CLUSTER="${KIND_CLUSTER_NAME:-kweaver-dev}"
docker stop $(docker ps -q --filter "label=io.x-k8s.kind.cluster=${CLUSTER}")
```

Resume: `docker start $(docker ps -aq --filter "label=io.x-k8s.kind.cluster=${CLUSTER}")` (reuse the same `CLUSTER`).

**Teardown (delete the cluster):** Optionally `bash ./dev/mac.sh data-services uninstall` (tear down MariaDB/Redis/Kafka/ZK/OpenSearch Helm releases; keeps kind), then `bash ./dev/mac.sh cluster down` (runs `kind delete cluster`; destroys the cluster).

Config: copy [`dev/conf/mac-config.yaml.example`](conf/mac-config.yaml.example) to **`dev/conf/mac-config.yaml`** (one-time). The real **`mac-config.yaml` is gitignored** so generated passwords are not committed; adjust `accessAddress` and registry as needed.  
`kweaver-dip` is not wired in `mac.sh` (use Linux `deploy.sh`).

See also: top-of-file comments in [`mac.sh`](mac.sh), `bash ./dev/mac.sh -h`.

### Recommended sizing & known gotchas

- **Resources (Docker Desktop / colima)**: give the VM **≥ 10 CPU**, **≥ 14 GB**, **60 GB disk** for a comfortable `--minimum` install. Less is risky:
  - `doc-convert` alone requests 1.5 CPU; 6 CPU schedulers fail with `Insufficient cpu` on 7+ Pending pods, 8 CPU still leaves no headroom.
  - `--memory 12` (GB) actually allocates **11.66 GiB** to the VM (GB→GiB conversion), which is **below the 12 GiB doctor threshold** — set `--memory 14` to clear it. Example: `colima start --cpu 10 --memory 14 --disk 60`.
  - **Avoid resizing mid-install.** Stopping/starting the VM after Helm releases are deployed has triggered the Redis ACL bug below.

- **Proxy env pollution before `cluster up`**: kind containers inherit `HTTP_PROXY` / `HTTPS_PROXY` from your shell. If they point at `127.0.0.1:<port>` or at a proxy that is not actually running, image pulls inside the kind node fail with `proxyconnect tcp: connect: connection refused`, and `curl http://localhost/...` from the host returns 502 (curl goes through the dead proxy). Always run before any `mac.sh` step:
  ```bash
  unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY all_proxy ALL_PROXY
  ```
  When verifying ingress later, also pass `--noproxy '*'` to curl as a belt-and-braces.

- **Redis pod stuck in `CrashLoopBackOff` with `WRONGPASS invalid username-password pair`** (typically after a VM/node restart, or after a Redis `helm upgrade`): the `proton-redis` image's `/config-init.sh` hard-codes the `monitor-user` password hash and only `sed`-replaces existing ACL lines, while the `sentinel`/`exporter` sidecars run `ACL SETUSER` + `ACL SAVE` at runtime and overwrite the on-disk ACL with hashes that no longer match the Secret. The on-disk file does not self-heal. Recover with:
  ```bash
  bash ./deploy.sh redis fix-acl
  ```
  This deletes `/data/conf/{users,sentinel-users}.acl` from the PVC and the Pod, so the init container re-enters its "if file does not exist" branch and copies fresh ACL files (with correct hashes) from the ConfigMap. Pods that crashed waiting on Redis (e.g. `agent-operator-integration`, `coderunner`) recover on their next backoff, or `kubectl delete pod` them to speed up. Equivalent manual recipe if you cannot use `deploy.sh`:
  ```bash
  kubectl exec -n resource redis-proton-redis-0 -c redis -- \
    rm -f /data/conf/users.acl /data/conf/sentinel-users.acl
  kubectl delete pod -n resource redis-proton-redis-0
  ```

- **`onboard --config` requires `=`**: use `bash ./dev/mac.sh -y onboard --config=conf/models.yaml`. The space form `--config conf/models.yaml` is rejected by the wrapped `onboard.sh` and produces `Unknown: --config`.

- **kind images don't show up in Docker Desktop's "Images" tab**: kind nodes run their own `containerd` inside the node container, separate from the host Docker engine. Kweaver application images live there, not in Docker Desktop's image store. They still consume Docker Desktop's disk budget (~15–25 GB for the full stack). Inspect / preload via:
  ```bash
  docker exec kweaver-dev-control-plane crictl images        # list images inside the kind node
  kind load docker-image <img:tag> --name kweaver-dev        # push a host-built image into kind
  ```

- **`mac.sh isf install` switches the stack to HTTPS automatically**: ISF (hydra/oauth2) requires HTTPS issuers, so the install path will (1) flip `mac-config.yaml` `accessAddress` to `https/443`, (2) generate a self-signed TLS cert + Secret `kweaver-ingress-tls`, (3) `helm upgrade` any already-installed `kweaver-core` releases so they pick up the new https `accessAddress`, then (4) install ISF and patch its ingress with TLS. Total time ~10 min on a fresh install. Browsers will warn on the self-signed cert — accept once. To stay on HTTP, just don't install ISF (`--minimum` already disables `auth.enabled`).

- **Quick verify after install** (proxy unset, Core pods Ready):
  ```bash
  curl --noproxy '*' http://localhost/                       # → 200, Sandbox Control Plane JSON
  curl --noproxy '*' http://localhost/api/bkn-backend/v1     # → 404 (path exists, not a handler)
  ```
  The historical `curl http://<lan-ip>/api/v1/health` printed by the installer does not match any ingress route — use the paths above (or any documented service path under `/api/...`) instead.

### Troubleshooting

- **`failed to connect to the docker API` / `docker.sock: no such file` when running `cluster up`:** the Docker **CLI** is installed but the **engine** is not running. Open **Docker Desktop**, wait until it is fully started, run `docker info` to confirm, then retry `cluster up`. `doctor` also checks engine reachability. **`doctor --fix` does not start Docker** (Homebrew only installs the CLI/cask); if everything else is already installed, just start Desktop and re-run `doctor`.

- **`kweaver-core-data-migrator` / pre-install job `BackoffLimitExceeded`:** ensure the **data layer** is up (normally automatic with **`kweaver-core install`**; otherwise run **`bash ./dev/mac.sh data-services install`**). Ensure **`depServices.rds`** points at in-cluster MariaDB after install (`mac-config` loopback placeholders may be updated when MariaDB is installed). Remove a failed release if Helm left it pending: `helm uninstall kweaver-core-data-migrator -n <namespace>` then re-run `kweaver-core install`.

### Onboard and `kweaver-admin` (full ISF)

`bash ./dev/mac.sh onboard` runs **`onboard.sh`** with **`CONFIG_YAML_PATH`** (usually `dev/conf/mac-config.yaml`). On a **full** install with ISF, the script calls **`kweaver-admin auth login`** with HTTP `-u`/`-p`. When the backend returns **401001017** (factory-default password blocked for HTTP `/oauth2/signin`), **if stdin and stdout are both a terminal**, **`onboard.sh` prompts** for the recovery method: **[Enter]** runs **`kweaver-admin auth change-password`** (CLI, default); **`o`** or **`oauth`** runs browser OAuth (**`auth login`** without `-u`/`-p`). After successful change-password, it asks once for the new password and retries HTTP **`auth login`**. **`onboard.sh` runtime hints stay English.**

| Approach | Typical command |
|----------|-----------------|
| OAuth / browser | `kweaver-admin auth login https://<access-address> -k` *(omit `-u` and `-p`)* |
| HTTP change-password | `kweaver-admin auth change-password https://<access-address> -u admin -k` *(URL required)* |
| Non-interactive | `kweaver-admin auth login … -p '<initial>' --new-password '<new>' -k`; then `export ONBOARD_DEFAULT_KWEAVER_PASSWORD` before `onboard -y` |

**Always pass the platform base URL** on the CLI. If you omit it, `kweaver-admin` uses the saved **active profile** from `kweaver-admin auth list`, which may be a different cluster — **not** a Helm `accessAddress` misread.

Details: [`help/en/install.md`](../../help/en/install.md) · [`help/zh/install.md`](../../help/zh/install.md) (administrator tool / 401001017).


## 中文说明

与**文档开头的英文小节**结构与内容对应；中英为同一流程的两套叙述。

**读者：**可与文首英文 **Audience** 相同：**macOS 开发者**作快速验证；**生产环境与主文档以 Linux 为准** —— [`deploy/README.zh.md`](../README.zh.md)、[`help/zh/install.md`](../../help/zh/install.md)。

本机 **kind** 起 Kubernetes，与 Linux `deploy.sh` 使用同一套 Helm Chart；宿主机不跑 **`preflight` / k3s / kubeadm**。

### 克隆仓库（先做好）

脚本与清单在仓库目录内；**`mac.sh` 不能脱离仓库单独使用**。请先 **[clone kweaver-ai/kweaver-core](https://github.com/kweaver-ai/kweaver-core)**，并切换到实际部署所用分支（如 `feature/deploy/k3s-module`），然后 **`cd` 进入 `deploy/`**：

```bash
git clone https://github.com/kweaver-ai/kweaver-core.git
cd kweaver-core/deploy   # 在此目录执行 bash ./dev/mac.sh ...（与 deploy.sh 同层）
```

从产品包解压时，路径中须有 **`deploy/`** 目录，布局与上文一致即可。

### 架构（Apple Silicon / arm64）

**Apple Silicon** 上 kind 节点默认为 **linux/arm64**。镜像来自 [`dev/conf/mac-config.yaml`](conf/mac-config.yaml) 的 **`image.registry`**（由 [`mac-config.yaml.example`](conf/mac-config.yaml.example) 复制）；须 **arm64 可用**（多架构 manifest 或 arm64 标签）。仅 **amd64** 的镜像易导致 *exec format error*。Intel Mac 上节点多为 **amd64**（除非另行指定平台）。

### 访问地址（HTTP 与自动 host）

- **HTTP 与 HTTPS：**HTTPS 加密并校验服务端；HTTP 不加密，开发环境常见；浏览器对 HTTP 的「不安全」提示属预期。
- **自动 IP：**`mac-config.yaml` 常用 **`accessAddress.scheme: http`**，且可**省略 `host`**（见示例）。**`kweaver-core install`** 会探测本机局域网 IP 并写入 values。若仅本机访问，可设 **`accessAddress.host: localhost`**。

### 操作流程

在 **`deploy/`** 下执行；用 **bash** 调用（如 `bash ./dev/mac.sh ...`）。**`kweaver-core` / `core`** 封装**默认带 `--minimum`**；全量依赖加 **`--full`**。

| 步骤 | 命令 | 是否必需？ |
|------|------|------------|
| 1 | `bash ./dev/mac.sh doctor` | 建议 |
| 2 | `bash ./dev/mac.sh doctor --fix`（或 `-y doctor --fix`） | 缺工具时 |
| 3 | `bash ./dev/mac.sh cluster up` | **安装前必须** |
| 4 | `bash ./dev/mac.sh data-services install` | **可选** — 仅单独装/刷新数据层；**`kweaver-core install` 会先跑同一套捆绑安装**（`KWEAVER_SKIP_DATA_SERVICES_BUNDLE=true` 可跳过） |
| 5 | `bash ./dev/mac.sh kweaver-core download` | **可选**（本地 chart 缓存；默认 **minimum** profile） |
| 6 | `bash ./dev/mac.sh kweaver-core install` | **必须** — 部署 Core（**默认 `--minimum`**）；默认**先装捆绑 data-services** |
| 7 | `bash ./dev/mac.sh onboard` | **可选**（需 `kweaver` CLI；`-y` 少交互） |

其它与 Linux **`deploy.sh`** 相同（须集群就绪、[`CONFIG_YAML_PATH`](conf/mac-config.yaml) 等与安装一致）：`bash ./dev/mac.sh isf install|download|uninstall|status`，`bash ./dev/mac.sh etrino …`（Vega；**`vega`** 为 **`etrino`** 别名）。ISF 对 DB/配置要求常更高。**未接入 `mac.sh`：**`kweaver-dip`。

**最短路径：**`cluster up` → `kweaver-core install`（**`--minimum` + 先装数据层**）。若 **`KWEAVER_SKIP_DATA_SERVICES_BUNDLE=true`**，须自备 DB/Kafka 等可达实例，或先执行 **`data-services install`**。

**暂歇省资源（不删集群）：**退出 **Docker Desktop** 即可。kind 依赖 Docker，等于停掉本地集群，**不是** `cluster down`（不会 `kind delete`）。再用时重新打开 Docker。

若 **Docker 要一直开着**，可以只停 kind **节点**容器（效果同样是集群不可用，未执行 `kind delete`）：

```bash
CLUSTER="${KIND_CLUSTER_NAME:-kweaver-dev}"
docker stop $(docker ps -q --filter "label=io.x-k8s.kind.cluster=${CLUSTER}")
```

恢复：再执行 `docker start $(docker ps -aq --filter "label=io.x-k8s.kind.cluster=${CLUSTER}")`（`CLUSTER` 同上）。

**卸载（删除集群）：**可先 **`bash ./dev/mac.sh data-services uninstall`**（卸数据层 Helm，保留 kind），再 **`bash ./dev/mac.sh cluster down`**（执行 `kind delete cluster`，销毁集群）。

**配置：**[`mac-config.yaml.example`](conf/mac-config.yaml.example) → **`mac-config.yaml`**；**已被 .gitignore**，避免口令入库；按需调整 **`accessAddress`**、**`image.registry`**。

另见：[`mac.sh`](mac.sh) 顶部注释、`bash ./dev/mac.sh -h`。

### 推荐资源 & 已知坑

- **资源（Docker Desktop / colima）**：建议给虚拟机 **≥ 10 CPU**、**≥ 14 GB 内存**、**60 GB 磁盘**才能稳跑 `--minimum`。给少了的风险：
  - 仅 `doc-convert` 一个 pod 就 request 1.5 CPU；6 CPU 时 7+ 个 pod `Insufficient cpu` 调度失败，8 CPU 也几乎打满。
  - `--memory 12`（GB）实际分配 **11.66 GiB**（GB→GiB 换算损失），**低于 doctor 的 12 GiB 阈值**会报内存不足；建议直接 `--memory 14`。例：`colima start --cpu 10 --memory 14 --disk 60`。
  - **不要安装中途 resize**：Helm 部署完后停/起 VM 会触发下面的 Redis ACL bug。

- **代理污染**（`cluster up` 之前）：kind 容器会继承宿主 `HTTP_PROXY` / `HTTPS_PROXY`。若代理地址是 `127.0.0.1:<port>` 或实际未启动，kind 节点拉镜像会报 `proxyconnect tcp: connect: connection refused`，宿主上 `curl http://localhost/...` 也会返 502（curl 走了死代理）。每次 `mac.sh` 之前先：
  ```bash
  unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY all_proxy ALL_PROXY
  ```
  后面验证 ingress 时也建议加 `--noproxy '*'` 给 curl 兜底。

- **Redis pod `CrashLoopBackOff` 且报 `WRONGPASS invalid username-password pair`**（通常发生在 VM/节点重启后，或 Redis `helm upgrade` 之后）：`proton-redis` 镜像内的 `/config-init.sh` 给 `monitor-user` 写死了一个固定 SHA256 且 else 分支只 `sed` 替换不补行；同时 `sentinel`/`exporter` sidecar 在运行期会跑 `ACL SETUSER` + `ACL SAVE`，把盘上 ACL 写成跟 Secret 不一致的 hash。盘上文件不会自愈，用一条命令修复：
  ```bash
  bash ./deploy.sh redis fix-acl
  ```
  脚本会删掉 PVC 上的 `/data/conf/{users,sentinel-users}.acl` 并删 Pod，让 init 容器走 "if not exists" 分支从 ConfigMap 重新拷正确 hash 的 ACL。依赖 Redis 的应用（`agent-operator-integration`、`coderunner` 等）会在下一次 backoff 后自愈，或 `kubectl delete pod` 加速。如果不方便用 `deploy.sh`，对应的手动等价命令：
  ```bash
  kubectl exec -n resource redis-proton-redis-0 -c redis -- \
    rm -f /data/conf/users.acl /data/conf/sentinel-users.acl
  kubectl delete pod -n resource redis-proton-redis-0
  ```

- **`onboard --config` 必须用 `=`**：`bash ./dev/mac.sh -y onboard --config=conf/models.yaml`。空格形式 `--config conf/...` 会被底层 `onboard.sh` 拒绝并报 `Unknown: --config`。

- **kind 镜像在 Docker Desktop 的 "Images" 面板看不到**：kind 节点在节点容器内独立跑了一份 `containerd`，跟宿主 Docker 不共享存储。kweaver 应用镜像都在那里，不在 Docker Desktop 的镜像列表里——但仍然占 Docker Desktop 的磁盘配额（全栈 ~15–25 GB）。查看 / 预加载：
  ```bash
  docker exec kweaver-dev-control-plane crictl images        # 列出 kind 节点内的镜像
  kind load docker-image <img:tag> --name kweaver-dev        # 把宿主已有镜像推进 kind
  ```

- **`mac.sh isf install` 会自动把整套切到 HTTPS**：ISF（hydra/oauth2）的 issuer 必须是 https，所以 install 流程会自动：(1) 把 `mac-config.yaml` 的 `accessAddress` 改成 `https/443`，(2) 用 openssl 生 self-signed 证书并落到 Secret `kweaver-ingress-tls`，(3) 对已装的 `kweaver-core` release 做 `helm upgrade` 让它们读到新 https `accessAddress`，(4) 装 ISF 并给 ingress patch TLS。全新场景大约 10 min。浏览器会提示自签证书风险，确认一次即可。**不装 ISF 的话不用动**——`--minimum` 默认就关了 `auth.enabled`。

- **装完后的快速验证**（代理已 unset、Core pod 全 Ready 后）：
  ```bash
  curl --noproxy '*' http://localhost/                       # → 200，Sandbox Control Plane JSON
  curl --noproxy '*' http://localhost/api/bkn-backend/v1     # → 404（路径在，无对应处理函数）
  ```
  安装器历史输出里的 `curl http://<局域网 IP>/api/v1/health` 没有任何 ingress 路由匹配，请改用上面这些路径或文档化的 `/api/...` 业务路径。

### 故障排除

- **`cluster up` 报 Docker API / `docker.sock`：**多为 **CLI 已装但引擎未起**。请先启动 **Docker Desktop**，`docker info` 通过后重试。**`doctor --fix`** 不会拉起守护进程。
- **`kweaver-core-data-migrator` / Job `BackoffLimitExceeded`：**确认数据层就绪（一般由 **`kweaver-core install` 自动安装**；否则 **`data-services install`**）。确认 **`depServices.rds`** 指向集群内 MariaDB；必要时 `helm uninstall kweaver-core-data-migrator -n <namespace>` 后再装 Core。

### Onboard 与 `kweaver-admin`（全量 ISF）

**`bash ./dev/mac.sh onboard`** 调用 **`onboard.sh`**（`CONFIG_YAML_PATH` 多为 **`dev/conf/mac-config.yaml`**）。**全量 + ISF** 时会用 HTTP **`-u`/`-p`** 执行 **`kweaver-admin auth login`**。若返回 **401001017**，在 **标准输入与标准输出均为终端** 时，**脚本会先询问方式**：（**默认回车**）执行 **`auth change-password`（CLI）**；输入 **`o` / `oauth`** 则用浏览器 OAuth（无 `-u`/`-p` 的 **`auth login`**）。CLI 改密成功后会再问一次新密码并重试 HTTP 登录。无 TTY（如纯 **`onboard -y` 流水线**）时只打印英文备选说明。**终端里脚本提示仍为英文。**

| 方式 | 命令要点 |
|------|----------|
| OAuth / 浏览器 | `kweaver-admin auth login https://<访问地址> -k`，**不要**加 `-u`/`-p` |
| HTTP 改密 | `kweaver-admin auth change-password https://<访问地址> -u admin -k`，**必须写 URL** |
| 非交互 | `kweaver-admin auth login … -p '<初始>' --new-password '<新>' -k`；`onboard -y` 前 **`export ONBOARD_DEFAULT_KWEAVER_PASSWORD`** |

**务必在命令中写出平台访问基址**；省略时 CLI 使用 **`kweaver-admin auth list`** 里当前激活会话的地址，易连到其它环境，**并非** Helm 误读 `accessAddress`。

详见 [`help/zh/install.md`](../../help/zh/install.md)、[`help/en/install.md`](../../help/en/install.md)（完整安装后的 `kweaver-admin` / 401001017）。
