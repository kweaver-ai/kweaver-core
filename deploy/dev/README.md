# macOS dev path (`dev/mac.sh`, kind)

**Audience:** Optional for **macOS developers** doing quick validation. **Production deployments and all primary documentation assume Linux** — start with [`deploy/README.md`](../README.md) and [`help/en/install.md`](../../help/en/install.md) / [`help/zh/install.md`](../../help/zh/install.md).

English | [中文](#中文说明)

Local Kubernetes with **kind** plus the same Helm charts as Linux `deploy.sh`. No host `preflight` / k3s / kubeadm.

### Architecture (Apple Silicon / arm64)

On **Apple Silicon** Macs, kind nodes are **linux/arm64** by default. Charts pull from `image.registry` in [`dev/conf/mac-config.yaml`](conf/mac-config.yaml); those images must be **arm64-capable** (multi-arch manifest or an arm64 tag). If a registry only ships **amd64**, pods often fail with *exec format error*. Intel Macs still get **amd64** kind nodes unless you force another platform.

### Access URL (HTTP and automatic host)

- **HTTP vs HTTPS:** HTTPS uses TLS to encrypt traffic and verify the server identity; HTTP is unencrypted. On a trusted LAN, HTTP avoids dealing with local TLS certs and is typical for dev. Browsers may still show “Not secure” for HTTP — expected.
- **Automatic IP:** [`mac-config.yaml`](conf/mac-config.yaml) uses `accessAddress.scheme: http` and **omits** `host`. On `kweaver-core install`, the flow detects your LAN IP (on macOS, usually the default-route interface) and writes it into values so other devices on the network can open the UI. Set `accessAddress.host` yourself (for example `localhost`) if you want same-machine-only URLs.

## Order of operations

Run from the **`deploy/`** directory (`cd deploy` in this repo). Invoke **`mac.sh` with bash** (e.g. `bash ./dev/mac.sh ...`). **`kweaver-core` / `core`:** the wrapper **defaults to `--minimum`** (smaller chart set; skips ISF in manifest terms). Pass **`--full`** for the full manifest profile (adds ISF download/install when the manifest enables it).

| Step | Command | Required? |
|------|---------|-----------|
| 1 | `bash ./dev/mac.sh doctor` | Recommended |
| 2 | `bash ./dev/mac.sh doctor --fix` (or `-y doctor --fix`) | If something is missing |
| 3 | `bash ./dev/mac.sh cluster up` | **Yes** before install |
| 4 | `bash ./dev/mac.sh kweaver-core download` | Optional (local chart cache; **minimum** by default) |
| 5 | `bash ./dev/mac.sh kweaver-core install` | **Yes** to deploy Core (**`--minimum` implied**; same as explicit `--minimum` or add `--full` to opt out) |
| 6 | `bash ./dev/mac.sh onboard` | Optional (models/BKN; needs `kweaver` CLI; add `-y` to skip prompts) |

Optional (same `deploy.sh` Helm paths as Linux; you need a working cluster + values that match your dependencies): `bash ./dev/mac.sh isf install|download|uninstall|status`, `bash ./dev/mac.sh etrino install|...` (Vega stack; **`vega` is an alias of `etrino`**). ISF may require DB/config beyond the minimal mac sample—see Linux `deploy.sh` help and your `CONFIG_YAML_PATH`.

**Minimal path:** `cluster up` → `kweaver-core install` (mac wrapper implies `--minimum`).

**Teardown:** `bash ./dev/mac.sh cluster down`.

Default values file: [`dev/conf/mac-config.yaml`](conf/mac-config.yaml).  
`kweaver-dip` is not wired in `mac.sh` (use Linux `deploy.sh`).

See also: top-of-file comments in [`mac.sh`](mac.sh), `bash ./dev/mac.sh -h`.

### Troubleshooting

- **`failed to connect to the docker API` / `docker.sock: no such file` when running `cluster up`:** the Docker **CLI** is installed but the **engine** is not running. Open **Docker Desktop**, wait until it is fully started, run `docker info` to confirm, then retry `cluster up`. `doctor` also checks engine reachability. **`doctor --fix` does not start Docker** (Homebrew only installs the CLI/cask); if everything else is already installed, just start Desktop and re-run `doctor`.

- **`ingress-nginx` wait failed / controller not ready:** ensure network can pull `registry.k8s.io` images (corporate proxy may block). Run `kubectl -n ingress-nginx get pods` and `kubectl -n ingress-nginx describe pod -l app.kubernetes.io/component=controller`. If ingress was half-installed, `kubectl delete ns ingress-nginx` then `bash ./dev/mac.sh cluster up` again (or `cluster down` and recreate kind).

---

## 中文说明

在 **`deploy/`** 目录下执行（例如 `cd deploy`）。请用 **bash** 调用：`bash ./dev/mac.sh ...`。**`kweaver-core` / `core`** 默认会加上 **`--minimum`**（精简 chart、按 manifest 跳过 ISF）；需要完整依赖时用 **`--full`**。

**推荐顺序：**

1. `bash ./dev/mac.sh doctor` — 检查环境（可选但建议）。  
2. `bash ./dev/mac.sh doctor --fix` — 缺工具时用 Homebrew 安装（可加 `-y` 跳过确认）。  
3. `bash ./dev/mac.sh cluster up` — **必须先有集群**（kind + ingress）。  
4. `bash ./dev/mac.sh kweaver-core download` — **可选**，只下载 chart 到本机，不装集群。  
5. `bash ./dev/mac.sh kweaver-core install` — **往集群装 Core**（封装默认 **`--minimum`**；完整依赖加 **`--full`**；少交互可在命令前加 `-y`）。  
6. `bash ./dev/mac.sh onboard` — **可选**（可加 `-y` 跳过交互）。

**可选：**`bash ./dev/mac.sh isf ...`、`bash ./dev/mac.sh etrino ...`（Vega 三套件；**`vega` 为 `etrino` 别名**）。ISF 通常对数据库/配置有更多要求，请结合 Linux 侧 `deploy.sh` 与你的 `CONFIG_YAML_PATH`。**未接入：**`kweaver-dip`。

**最短路径：**`cluster up` → `kweaver-core install`（mac 封装默认带 **`--minimum`**，等价于以前的 `install --minimum`；要完整版型请加 **`--full`**）。  
**删除本机 kind：**`bash ./dev/mac.sh cluster down`。

**架构：**Apple Silicon 上 kind 节点一般为 **linux/arm64**，镜像需支持 arm64/多架构（见上节及 `dev/conf/mac-config.yaml` 中的 `image.registry`）。仅 amd64 的镜像在 arm64 节点上常会 *exec format error*。

**访问地址：**默认 **HTTP**（`accessAddress.scheme: http`，端口与 `mac-config` 一致）。**不写 `host`** 时，安装流程会**自动探测本机局域网 IP**（macOS 上多为默认路由网卡），便于同网段其它机器访问；若只要本机访问，可在 `mac-config` 里显式设 `accessAddress.host: localhost`。**HTTPS** 需 TLS 证书，适合生产或需要加密/校验域名的场景；本地开发常用 HTTP 以减少证书折腾。

**故障：**若 `cluster up` 报错无法连接 `docker.sock`，请先**打开 Docker Desktop** 并等其启动完成，执行 `docker info` 确认后再试；`doctor` 会检查 Docker 引擎是否可用。**`doctor --fix` 不会启动 Docker 守护进程**（只能按需用 Homebrew 装 CLI）；若其它工具已齐，只需启动 Desktop 后再执行 `doctor`。

更多参数见 `bash ./dev/mac.sh -h` 及脚本头部注释。
