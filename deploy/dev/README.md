# macOS dev path (`dev/mac.sh`, kind)

**Audience:** Optional for **macOS developers** doing quick validation. **Production deployments and all primary documentation assume Linux** — start with [`deploy/README.md`](../README.md) and [`help/en/install.md`](../../help/en/install.md) / [`help/zh/install.md`](../../help/zh/install.md).

English | [中文](README.zh.md)

Local Kubernetes with **kind** plus the same Helm charts as Linux `deploy.sh`. No host `preflight` / k3s / kubeadm.

> ⚠️ **Do not use `sudo` on macOS.** Run every command with plain `bash`. Docker Desktop / `kind` / `$HOME` (config + `kweaver` token) all belong to your user; `sudo` redirects them to `/var/root` and breaks the flow. `deploy.sh` already short-circuits root checks on `Darwin`, so `sudo` adds nothing.

### Repository (clone first)

Scripts and vendored manifests live in the repo tree — **`mac.sh` is not a standalone installer.** Clone **[kweaver-ai/kweaver-core](https://github.com/kweaver-ai/kweaver-core)** (check out the branch you deploy from), then **`cd`** into **`deploy/`** before any command below:

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
  - **8 GB** Docker RAM is usually **too small** for this stack (symptoms: high pod **RESTARTS**, long-lived **`0/1 Running`**, or thrashing after the VM sleeps). Treat **~10 GB as a practical floor** to even try; **`mac.sh doctor` warns below ~12 GiB** (`MAC_DOCTOR_MIN_MEM_GB`); use **14–16 GB** for stable headroom as below.
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
