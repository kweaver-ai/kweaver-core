# macOS dev path (`dev/mac.sh`, kind)

**Audience:** Optional for **macOS developers** doing quick validation. **Production deployments and all primary documentation assume Linux** — start with [`deploy/README.md`](../README.md) and [`help/en/install.md`](../../help/en/install.md) / [`help/zh/install.md`](../../help/zh/install.md).

English | [中文](#中文说明)

Local Kubernetes with **kind** plus the same Helm charts as Linux `deploy.sh`. No host `preflight` / k3s / kubeadm.

### Architecture (Apple Silicon / arm64)

On **Apple Silicon** Macs, kind nodes are **linux/arm64** by default. Charts pull from `image.registry` in [`dev/conf/mac-config.yaml`](conf/mac-config.yaml); those images must be **arm64-capable** (multi-arch manifest or an arm64 tag). If a registry only ships **amd64**, pods often fail with *exec format error*. Intel Macs still get **amd64** kind nodes unless you force another platform.

## Order of operations

Run from the **`deploy/`** directory (`cd deploy` in this repo).

| Step | Command | Required? |
|------|---------|-----------|
| 1 | `bash ./dev/mac.sh doctor` | Recommended |
| 2 | `bash ./dev/mac.sh doctor --fix` (or `-y doctor --fix`) | If something is missing |
| 3 | `bash ./dev/mac.sh cluster up` | **Yes** before install |
| 4 | `bash ./dev/mac.sh kweaver-core download` | Optional (local chart cache only) |
| 5 | `bash ./dev/mac.sh kweaver-core install --minimum` | **Yes** to deploy Core |
| 6 | `bash ./dev/mac.sh onboard` | Optional (models/BKN; needs `kweaver` CLI; add `-y` to skip prompts) |

Optional (same `deploy.sh` Helm paths as Linux; you need a working cluster + values that match your dependencies): `bash ./dev/mac.sh isf install|download|uninstall|status`, `bash ./dev/mac.sh etrino install|...` (Vega stack; **`vega` is an alias of `etrino`**). ISF may require DB/config beyond the minimal mac sample—see Linux `deploy.sh` help and your `CONFIG_YAML_PATH`.

**Minimal path:** `cluster up` → `kweaver-core install --minimum`.

**Teardown:** `bash ./dev/mac.sh cluster down`.

Default values file: [`dev/conf/mac-config.yaml`](conf/mac-config.yaml).  
`kweaver-dip` is not wired in `mac.sh` (use Linux `deploy.sh`).

See also: top-of-file comments in [`mac.sh`](mac.sh), `bash ./dev/mac.sh -h`.

---

## 中文说明

在 **`deploy/`** 目录下执行（例如 `cd deploy`）。

**推荐顺序：**

1. `bash ./dev/mac.sh doctor` — 检查环境（可选但建议）。  
2. `bash ./dev/mac.sh doctor --fix` — 缺工具时用 Homebrew 安装（可加 `-y` 跳过确认）。  
3. `bash ./dev/mac.sh cluster up` — **必须先有集群**（kind + ingress）。  
4. `bash ./dev/mac.sh kweaver-core download` — **可选**，只下载 chart 到本机，不装集群。  
5. `bash ./dev/mac.sh kweaver-core install --minimum` — **真正往集群里装 Core**（需要少交互时在命令前加 `-y`）。  
6. `bash ./dev/mac.sh onboard` — **可选**（可加 `-y` 跳过交互）。

**可选：**`bash ./dev/mac.sh isf ...`、`bash ./dev/mac.sh etrino ...`（Vega 三套件；**`vega` 为 `etrino` 别名**）。ISF 通常对数据库/配置有更多要求，请结合 Linux 侧 `deploy.sh` 与你的 `CONFIG_YAML_PATH`。**未接入：**`kweaver-dip`。

**最短路径：**`cluster up` → `kweaver-core install --minimum`。  
**删除本机 kind：**`bash ./dev/mac.sh cluster down`。

**架构：**Apple Silicon 上 kind 节点一般为 **linux/arm64**，镜像需支持 arm64/多架构（见上节及 `dev/conf/mac-config.yaml` 中的 `image.registry`）。仅 amd64 的镜像在 arm64 节点上常会 *exec format error*。

更多参数见 `bash ./dev/mac.sh -h` 及脚本头部注释。
