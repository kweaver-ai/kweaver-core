# 提交监听服务：基于 GitHub 的技术实现

本文档说明如何**对接 GitHub 开源仓库**，实现《测试架构设计_基于Commit与多智能体》中的「提交监听服务」：在每次 push 后产出**变更路径、变更类型、diff、受影响 API、scope、suggested_suites、need_add_cases** 等，供下游智能体与执行器使用。

---

## 一、目标输出（与架构对齐）

提交监听服务需产出的**标准化输出**（建议 JSON），供智能体 A/B/C 消费：

```json
{
  "repo": "owner/repo",
  "branch": "main",
  "commit_sha": "abc123...",
  "commit_message": "feat: add datasource test API",
  "commit_url": "https://github.com/owner/repo/commit/abc123",
  "changed_files": [
    { "path": "vega/data-connection/src/controller/DatasourceController.java", "status": "modified" },
    { "path": "vega/data-connection/src/service/DatasourceService.java", "status": "modified" }
  ],
  "change_summary": {
    "added": ["path/a.go"],
    "modified": ["vega/data-connection/..."],
    "removed": []
  },
  "scopes": ["vega-data-connection"],
  "scope_tags": ["regression", "data-connection"],
  "suggested_suites": ["测试数据源", "新增数据源", "查询数据源"],
  "affected_api_names": ["测试数据源连接", "新增数据源"],
  "affected_api_paths": ["/api/data-connection/v1/datasource/test", "/api/data-connection/v1/datasource"],
  "need_add_cases": true,
  "diffs": [{ "path": "...", "patch": "..." }]
}
```

- **changed_files**：变更路径 + 变更类型（added / modified / removed），对应架构中的「变更路径、变更类型」。
- **diffs**：可选，每个文件的 patch，供智能体做细粒度分析。
- **scopes / suggested_suites**：通过本仓库的 **path_scope_mapping** 对变更路径匹配得到。
- **affected_api_names / affected_api_paths**：通过「路径→API」映射或解析 OpenAPI/代码得到（见下文）。

---

## 二、实现方式概览

| 方式 | 适用场景 | 要点 |
|------|----------|------|
| **A. GitHub Webhook + 自建服务** | 监听在你们自己的服务器/内网，仓库在 GitHub | 在 GitHub 配置 Webhook，push 时调你的 HTTP 服务；服务里调 GitHub API 拿 commit 详情并做 path 匹配 |
| **B. GitHub Actions** | 希望全部在 GitHub 上跑、无需公网回调地址 | 在**被监听的研发仓库**里加 workflow：`on: push`，用 `GITHUB_TOKEN` 调 Commits API 取文件列表，再跑脚本做 path_scope_mapping、输出 JSON（artifact 或触发下游） |
| **C. 内网部署（无法访问 GitHub）** | 开发/测试服务器在内网，无法访问 GitHub | 见 [八、内网部署方案](#八内网部署方案) |

两种方式 A、B **后续逻辑一致**：拿到 commit + 变更文件列表后，用同一套「路径匹配 + 可选 API 推断」产出上述 JSON。下面以 B（Actions）为主、A 简要说明。

---

## 三、方式 A：GitHub Webhook + 自建服务

### 3.1 在 GitHub 仓库配置 Webhook

1. 仓库 **Settings → Webhooks → Add webhook**。
2. **Payload URL**：你的服务公网地址，例如 `https://your-server.com/webhooks/github/push`。
3. **Content type**：`application/json`。
4. **Secret**：填一个随机字符串，用于请求签名校验（见 3.3）。
5. **Which events**：仅 **Just the push event**（或根据需要选 Pushes）。
6. 保存后，每次 push 会向该 URL 发 POST。

**多分支、多人同时提交**时的处理方式见 [3.4 多分支并发 push 的处理](#34-多分支并发-push-的处理)。

### 3.2 Webhook 请求体（Push Event）里有什么

Push 的 payload 示例（只列关键字段）：

```json
{
  "ref": "refs/heads/main",
  "repository": { "full_name": "owner/repo", "clone_url": "..." },
  "commits": [
    {
      "id": "abc123...",
      "message": "feat: add api",
      "added": [],
      "removed": [],
      "modified": ["vega/data-connection/Controller.java"]
    }
  ]
}
```

注意：**部分情况下** push 里的 `commits[].added/removed/modified` 可能不完整或为空（例如 merge、大 push）。更稳妥的做法是：用 **commits[].id**（或 `after` 的 commit sha）再调 **GitHub API** 取单次 commit 的完整文件列表与 diff。

### 3.3 自建服务要做的事

1. **接收 POST**  
   - 校验 `X-Hub-Signature-256`（HMAC-SHA256(secret, body)），避免伪造。  
   - 解析 `ref`、`repository.full_name`、`commits`（或 `head_commit`）。

2. **取 commit 详情（文件列表 + 可选 diff）**  
   - 使用 GitHub **REST API**（需 Token）：  
     - **推荐**：用 **Compare API** `GET /repos/{owner}/{repo}/compare/{base}...{head}`，其中 `base` 为该 push 的 `before`、`head` 为 `after`，得到**整次 push 的汇总 diff**（一次 push 含多 commit 时也是这一段区间内所有变更的并集）。响应里 `files[]` 含 `filename`、`status`、`patch`。  
     - 或单 commit：`GET /repos/{owner}/{repo}/commits/{ref}`，仅当只关心「最后一个 commit」时使用；多 commit 时用 compare 更符合「这一 push 改了啥」的语义。  
   - 同一分支多次提交时的策略见 [3.5 同一分支多次提交的处理](#35-同一分支多次提交的处理)。

3. **路径匹配与产出**  
   - 用变更的 **filename** 列表，对 **path_scope_mapping.yaml** 做匹配（见第四节）。  
   - 得到 `scopes`、`scope_tags`、`suggested_suites`；再根据「路径→API」映射或 OpenAPI 解析得到 `affected_api_names/paths`；根据变更类型与 scope 决定 `need_add_cases`。  
   - 组装成上节的 JSON，**下发给流水线**（例如发到消息队列、调 CI API、或写文件供后续步骤读）。

4. **Token**  
   - 调用 GitHub API 时使用 **Personal Access Token (PAT)** 或 **GitHub App** 安装 token，权限至少：`contents: read`（公开库只读即可）。

### 3.4 多分支并发 push 的处理

当**不同分支上多人同时提交**时，GitHub 会按每次 push **分别发送 webhook**：每个请求的 payload 里都带 **`ref`**（如 `refs/heads/main`、`refs/heads/feature/xxx`），表示「这次 push 发生在哪个分支」。因此会**并发**收到多条 webhook。

**处理原则：**

| 要点 | 说明 |
|------|------|
| **按分支、按 push 独立处理** | 每条 webhook 对应「某分支的某次 push」，单独跑一轮：取该 push 的 commit/diff → 路径匹配 → 产出 JSON → 触发下游（智能体/执行）。**分支之间互不依赖**，不需要把多分支「合并」成一次再处理。 |
| **并发能力** | 服务端应支持**并发**处理多条 webhook（异步处理、队列 + 多 worker，或快速 200 后后台任务）。同一时刻 main 上一条 push、feature 上一条 push，可并行跑两套流水线，输出两份 listener 结果（各自带 `branch`、`commit_sha`）。 |
| **幂等与去重** | 用 `(repository.full_name, ref, after)`（或 `head_commit.id`）作为**幂等键**：同一 push 若因 GitHub 重试被多次投递，只处理一次，后续相同键直接返回 200 或跳过。不同分支、不同 commit 的 push 键不同，可并行。 |
| **只监听部分分支（可选）** | 若只关心合入主分支（如 `main`/`master`）后的分析，可在服务端**过滤 ref**：仅当 `ref == "refs/heads/main"`（或配置的白名单）时才做「取 commit → 匹配 → 下游」；其他 ref 收到后直接返回 200，不继续处理，减少噪音和资源。 |

**下游衔接建议：**

- **测试执行**：每个 push 对应自己的测试任务（参数带 `branch` + `commit_sha`），报告与流水线按「分支 + sha」区分即可。
- **用例仓库更新**：若智能体会改测试用例仓库（如 suites/*.yaml），需约定是「仅 main 的 push 触发写回」还是「每分支各自 PR」；多分支同时改同一模块时，由 Git 合并策略与 Code Owner 处理冲突，监听服务只负责按 push 正确产出上下文。

**GitHub Actions 场景**：每个 push 会触发一次 workflow run，**不同分支的 push 会触发多个 run**，天然按分支独立。若只想在主分支 push 时跑监听，在 workflow 里写 `on: push: branches: [main]` 即可；不写则所有分支 push 都会触发。

### 3.5 同一分支多次提交的处理

分两种情形：**一次 push 里包含多个 commit**、**同一分支上先后多次 push**。

| 情形 | 行为 | 推荐处理方式 |
|------|------|--------------|
| **一次 push 里含多个 commit** | 例如本地连续 3 个 commit 后一次 `git push`，GitHub 只发**一条** webhook，payload 里 `commits[]` 有多个、`after` 为最新 sha。 | 用 **Compare API** 取 `before...after`（即该 push 的「从旧头到新头」）得到**整次 push 的汇总 diff**，即这段区间内所有变更文件的并集。**按「一次 push = 一份 listener 输出」**处理即可，无需按每个 commit 拆开；智能体/下游看到的是「这一 push 总共改了哪些」，足够做 scope 与用例分析。若确有需求按每个 commit 单独分析，可遍历 `commits[]` 对每个调 `compare(commit^..commit)` 产出多份结果，一般不必。 |
| **同一分支上先后多次 push** | 例如先 push 一次（main@abc），过一会儿再 push（main@def）。**每次 push 都会触发一条 webhook**：第一次 `ref=main, after=abc`，第二次 `ref=main, after=def`。 | **每次 push 独立处理**：第一次产出「从上一状态到 abc」的变更并跑下游，第二次产出「从 abc 到 def」的变更并跑下游。两套流水线、两份结果，用 `(branch, commit_sha)` 区分即可。**可选策略**：若希望「同一分支只对最新 push 跑流水线」（避免短时间多次 push 造成排队），可在处理前查当前分支 tip：若 `after` 已不是该分支最新 sha，则跳过本次；或下游用「同一 branch 新任务取消同分支未完成任务」的队列策略。 |

**小结**：同一分支多次提交时，**一次 push 对应一份汇总 diff 与一份监听输出**；同一分支多次 push 则对应多份，按 push 粒度独立处理即可，必要时再叠加「只跑最新」的策略。

### 3.6 智能体分析过程中又有新提交且与上一次冲突的处理

**场景**：流水线已基于某次 push（如 main@abc）启动智能体分析（补用例、判维度、筛 case 等）；在分析或写回过程中，**同一分支又产生了新 push**（main@def），且与上一次在**变更文件或写回目标**上存在重叠（例如都改了同一接口、或都会改同一份 suites/*.yaml），形成「基于旧提交的产出」与「新提交」的冲突。

**冲突的两类含义：**

| 类型 | 说明 |
|------|------|
| **上下文过期** | 智能体当前分析基于 commit abc 的 diff/文件内容；但分支 tip 已是 def，abc 与 def 改了相同或相关文件。此时基于 abc 的补用例或维度判断可能**不再贴合当前代码**，若仍写回或按 abc 跑测试，可能无效或误导。 |
| **写回冲突** | 智能体要把对「用例仓库」的修改（如 suites/*.yaml）推回或提 PR；但此时另一条流水线（基于 def）或人工已修改了同一文件，导致 **Git 合并冲突**（merge conflict）或后写覆盖前写。 |

**推荐策略（可组合）：**

| 策略 | 做法 | 适用 |
|------|------|------|
| **写回前校验分支 tip** | 在智能体**写回用例仓库或触发执行前**，调 GitHub API 查询该分支当前 tip（如 `GET .../git/ref/heads/main`）。若 tip 不等于本次流水线所带的 `commit_sha`，说明已有新提交，**本 run 视为过期**：不写回、不触发测试，可选触发一条「基于新 tip」的流水线。 | 避免基于过期上下文写回或跑测；减少与「新提交」的冲突。 |
| **同一分支串行 / 新 push 取消旧 run** | 约定**同一分支同一时刻只跑一条「监听→智能体→写回」流水线**。新 push 到达时：取消该分支上仍在进行的 run（如 GitHub Actions `concurrency: cancel-in-progress` 同 branch），只保留「基于最新 push」的 run。这样不会出现「分析 abc 的同时 def 也在跑」的两条并行写回。 | 从源头减少「分析过程中又来新提交」的交叉。 |
| **写回走 PR + 以最新为 base** | 智能体不直接 push 到用例仓库 main，而是推送到 `auto/{branch}-{sha}` 并提 PR。若本 run 因 tip 已变而被判定过期，可**不合并该 PR** 或关闭；新 push 对应新 PR，新 PR 的 base 为当前 main（已含或未含前一次结果）。合并时若有冲突，按常规 PR 流程解决（人工或 bot rebase）。 | 写回冲突通过 Git/PR 流程暴露并解决，避免静默覆盖。 |
| **冲突时的合并策略** | 若允许「两次 run 都写完」且都提了 PR：后合并的 PR 需在最新 base 上 rebase，若冲突则人工或策略（如「同一 suite 文件取新 run 的版本」）解决。 | 不取消旧 run 时的兜底。 |

**建议组合**：**写回前校验 tip + 同一分支新 push 取消旧 run**。即：同一分支只保留「最新 push」对应的那条流水线；若某 run 在写回前发现 tip 已不是自己的 `commit_sha`，则直接放弃写回并退出。这样「智能体分析过程中又有新提交且与上一次冲突」时，旧 run 会被取消或主动放弃，只保留基于最新提交的一条链路。

**GitHub Actions 示例（同一分支取消未完成 run）**：

```yaml
concurrency:
  group: commit-listener-${{ github.ref }}   # 同分支共用一个 group
  cancel-in-progress: true                   # 同分支新 push 取消未完成的 run
```

---

## 四、方式 B：GitHub Actions（推荐用于开源仓库）

无需公网回调，一切在 GitHub 上完成：在**被监听的研发仓库**里加一个 workflow，在 push 时用 GitHub 提供的 `GITHUB_TOKEN` 调 Commits API，再跑脚本做路径匹配并输出 JSON。

### 4.1 在研发仓库添加 workflow

在**研发仓库**（被监听的 GitHub 仓库）根目录创建 `.github/workflows/commit-listener.yml`（或与现有 CI 合并）：

```yaml
name: Commit Listener

on:
  push:
    branches: [main, master]   # 仅这些分支 push 时触发；多分支同时 push 会各自触发一次 run，互不干扰

# 同一分支新 push 时取消未完成的 run，避免「分析过程中又来新提交」的冲突（见 3.6 节）
concurrency:
  group: commit-listener-${{ github.ref }}
  cancel-in-progress: true

jobs:
  collect-changes:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repo
        uses: actions/checkout@v4
        with:
          fetch-depth: 2   # 至少 2 以便和前一 commit 对比

      - name: Get changed files (API)
        id: changes
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          # 使用 GitHub API 获取本次 push 涉及的文件（单 commit 或 compare）
          if [ "${{ github.event.before }}" = "0000000000000000000000000000000000000000" ]; then
            # 首次 push，只有 after
            REF="${{ github.sha }}"
          else
            REF="${{ github.event.before }}...${{ github.sha }}"
          fi
          curl -sS -H "Authorization: token $GH_TOKEN" \
            "https://api.github.com/repos/${{ github.repository }}/compare/$REF" \
            > compare.json
          # 提取 files: [{"filename":"path","status":"added|modified|removed"}, ...]
          jq '[.files[] | {path: .filename, status: .status}]' compare.json > changed_files.json
          jq -r '.files[] | "\(.filename) \(.status)"' compare.json > changed_list.txt
          echo "changed_files<<EOF" >> $GITHUB_OUTPUT
          cat changed_files.json >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Checkout test framework (path_scope_mapping)
        run: |
          git clone --depth 1 https://github.com/YOUR_ORG/DIP_AT.git dip_at || true
          # 若 path_scope_mapping 在研发仓库内，可改为本 repo 路径
        # 若映射表就在本仓库，可删掉 clone，直接读本 repo 下的路径

      - name: Run scope mapping
        id: scope
        run: |
          pip install pyyaml
          python scripts/commit_scope_mapping.py \
            --changed-list changed_list.txt \
            --mapping dip_at/testcase/vega/_config/path_scope_mapping.yaml \
            --repo "${{ github.repository }}" \
            --sha "${{ github.sha }}" \
            --message "${{ github.event.head_commit.message }}" \
            --output listener_output.json
        # 若映射在本仓库：--mapping path/to/path_scope_mapping.yaml

      - name: Upload listener output
        uses: actions/upload-artifact@v4
        with:
          name: commit-listener-output
          path: listener_output.json
```

- **compare API**：`GET /repos/{owner}/{repo}/compare/{base}...{head}` 返回两次 commit 之间的 diff，其中 `files[]` 含 `filename`、`status`（added/modified/removed），可选带 `patch`。  
- **path_scope_mapping**：若 DIP_AT 在另一仓库，可 clone 后读其 `testcase/vega/_config/path_scope_mapping.yaml`；若映射表放在研发仓库内，则不必 clone，直接 `--mapping` 指到本仓库路径即可。

### 4.2 路径匹配脚本（与 path_scope_mapping 对齐）

本仓库已提供脚本 **`scripts/commit_scope_mapping.py`**：读 `changed_list.txt`（每行 `path status`）和 `path_scope_mapping.yaml`，输出架构所需的 JSON。用法示例：

```bash
python scripts/commit_scope_mapping.py \
  --changed-list changed_list.txt \
  --mapping testcase/vega/_config/path_scope_mapping.yaml \
  --repo "${{ github.repository }}" --sha "${{ github.sha }}" \
  --message "${{ github.event.head_commit.message }}" \
  --output listener_output.json
```

可选：若配置了 `--path-to-api path_to_api_mapping.yaml`，可产出 `affected_api_names` / `affected_api_paths`（见下）。  
匹配逻辑与《测试框架架构_全自动与智能体规范》第七节一致：

1. **按路径匹配 scope**  
   - 对每个变更文件路径，与各 subsystem 的 `path_patterns` 做**通配匹配**（如 `vega/data-connection/**` 用 `fnmatch` 或 `pathlib.Path.match`）。  
   - 注意：GitHub 返回的 path 是仓库根相对路径，无前导 `/`；path_scope_mapping 里也是相对模式，例如 `vega/data-connection/**`。  
   - 合并所有命中的 subsystem 的 `id` → `scopes`，`scope_tags`、`suggested_suites` 取并集。

2. **need_add_cases**  
   - 若存在 `added` 且命中某 subsystem，或变更类型为新增接口相关文件，可置为 `true`；否则可按策略置为 `false`。

3. **affected_api_names / affected_api_paths**  
   - **方式 1（推荐先做）**：在模块 _config 下增加 **path_to_api_mapping.yaml**（或合并进 path_scope_mapping），例如：
     ```yaml
     path_to_api:
       - path_pattern: "**/data-connection/**/DatasourceController.java"
         api_path: /api/data-connection/v1/datasource
         api_name: 新增数据源
       - path_pattern: "**/datasource/test*"
         api_path: /api/data-connection/v1/datasource/test
         api_name: 测试数据源连接
     ```
     对每个变更 path 做 pattern 匹配，得到 api_path / api_name 列表。  
   - **方式 2**：若仓库内有 OpenAPI 描述（如 `openapi.yaml`），可解析「本次改动的文件是否属于某 operation」来推断 affected_api_paths，再通过 apis.yaml 的 name↔path 得到 api_name。

4. **diffs**  
   - 若需要 patch：在调用 Compare API 时，响应里 `files[].patch` 已包含；在脚本里把 `compare.json` 中的 `files[].filename` 与 `files[].patch` 写入输出 JSON 的 `diffs` 即可。

输出写入 `listener_output.json`，格式与第一节的 JSON 一致，供后续 job（或另一 workflow）读取：例如触发智能体流水线、或本 workflow 内继续「根据 listener_output 调用 get_cases + pytest」。

### 4.3 与下游衔接

- **Artifact**：`listener_output.json` 已作为 artifact 上传，后续 job 可通过 `download-artifact` 读取。  
- **触发测试流水线**：  
  - 同一 repo：在 `collect-changes` 后加 job，`needs: collect-changes`，读 artifact，用 `affected_api_names`/`scopes`/`suggested_suites` 调用 DIP_AT 的 get_cases 或设置环境变量再跑 pytest。  
  - 跨 repo（例如测试框架在另一仓库）：可用 `repository_dispatch` 或 `workflow_dispatch` 带 `listener_output` 的 JSON 触发测试仓库的 workflow；或通过 GitHub API 触发外部 CI。

---

## 五、路径匹配与 path_scope_mapping 的对应关系

- **path_scope_mapping.yaml** 中 `path_patterns` 是**仓库相对路径**的 glob，例如：
  - `vega/data-connection/**`
  - `vega/vega-backend/**/data-connection*`
- GitHub 返回的 `filename` 也是仓库根相对路径，例如 `vega/data-connection/src/controller/X.java`。
- 匹配时建议：对每个 `filename`，若存在某个 subsystem 的某个 `path_pattern` 满足 `fnmatch(filename, pattern)`（或把 pattern 转为 regex），则该 subsystem 命中，收集其 `id`、`scope_tags`、`suggested_suites`。

这样得到的 `scopes`、`scope_tags`、`suggested_suites` 与现有 DIP_AT 的 get_cases(scope=...) 完全一致。

---

## 六、安全与权限（GitHub）

- **Webhook**：务必校验 `X-Hub-Signature-256`，Secret 与 GitHub 配置一致。  
- **Token**：  
  - Actions 内用 `GITHUB_TOKEN` 即可读当前 repo 的 commit/compare。  
  - 若需读**其他私有仓库**（如 DIP_AT 或研发私有库），需在 repo 的 Settings → Secrets 配置 PAT 或 GitHub App 安装 token，并赋予 `contents: read`。  
- **开源仓库**：若研发仓库公开，仅读 commit 与 compare 不需要额外 token 权限；若 DIP_AT 公开，clone 时用 HTTPS 即可。

---

## 七、智能体代码上下文：diff 与是否提供完整仓库

智能体需要**理解本次提交改了什么**才能做下一步分析（补用例、判测试维度、筛 case）。仅给 **diff（patch）** 有时不够：缺少改动行的完整上下文（如新增的接口在哪个类、参数校验范围等）。是否要提供**完整代码仓库**，建议如下。

### 7.1 结论：不必默认提供完整仓库，推荐「diff + 变更文件完整内容」

| 方式 | 智能体能做的 | 成本/风险 | 建议 |
|------|--------------|-----------|------|
| **仅 diff** | 知道改了什么行，但缺少前后文，难以推断「新接口定义、参数约束」等 | 体积小、实现简单 | 适合只做 scope/路径匹配；做用例与维度分析时容易信息不足 |
| **diff + 变更文件完整内容（commit 后版本）** | 既能看改动片段，又能看每个变更文件的**当前完整内容**，便于理解新接口、新参数、新校验 | 仅多传变更的 N 个文件，体积可控 | **推荐**：满足大多数「理解提交 → 补用例 / 判维度」场景 |
| **diff + 变更文件 + 关联文件** | 在上一档基础上，再附带接口定义（如 OpenAPI）、DTO、常量类等 | 需定义「关联」规则（同目录、同模块、或从 import 解析） | 改动涉及多文件、契约变更时可选 |
| **完整代码仓库** | 可做任意跨文件、跨模块分析 | 仓库大时 token/传输/安全压力大；智能体也未必需要全量 | **按需**：仅当确实要做深度影响分析、重构分析时再提供 |

因此：**默认不提供完整仓库**；优先为智能体提供 **diff + 变更文件的完整内容（commit 后版本）**，必要时再扩展「关联文件」或只读仓库访问。

### 7.2 实现方式：在监听输出中附带「变更文件完整内容」

在提交监听的输出 JSON 中，除 `diffs` 外，增加**变更文件在 commit 后的完整内容**，供智能体一次性使用：

- **方式 1：内联在 JSON**  
  - 增加字段 `changed_file_contents`：`[{ "path": "vega/.../X.java", "content": "完整文件内容（commit 后）" }]`。  
  - 适合变更文件少、单文件不大的情况；否则 JSON 会很大。

- **方式 2：单独「智能体上下文包」artifact（推荐）**  
  - 提交监听步骤中：在拿到 compare 结果后，用 GitHub API **按 path + ref 拉取每个变更文件的原始内容**（`GET /repos/{owner}/{repo}/contents/{path}?ref={sha}`），写入一个目录或 zip（如 `agent_context/`：`diffs.json`、`files/ path/to/file.java`）。  
  - 将该目录或 zip 作为 **artifact** 上传（如 `commit-context-for-agent`）。  
  - 智能体侧：下载该 artifact，先读 `diffs.json` 与文件列表，再按需读 `files/` 下完整文件内容做「整理理解」与下一步分析。  
  - 优点：不撑大主 JSON；大文件、二进制可选择性不放入 artifact。

- **方式 3：按需拉取 URL**  
  - 在 listener 输出中为每个变更文件提供 **内容 URL**，例如：  
    `"content_url": "https://api.github.com/repos/{owner}/{repo}/contents/{path}?ref={sha}"`  
  - 智能体在需要时再请求该 URL（需 Token）。适合智能体运行环境可直接访问 GitHub API 的场景。

**GitHub Actions 示例（方式 2）**：在现有「Get changed files」步骤后增加一步，用 `gh api` 或 `curl` 对 `changed_files.json` 中每个 path 调 Contents API，将响应中的 `content`（base64 解码）写入 `agent_context/files/<path>`，同时把 `diffs` 写入 `agent_context/diffs.json`；最后将 `agent_context` 打包为 zip 并 upload-artifact。

### 7.3 何时才提供「完整仓库」给智能体

在以下情况可考虑提供**完整仓库**（或更大范围代码）：

- 需要做**跨模块 / 跨目录**的影响分析（如某公共库改动影响哪些调用方）。  
- 智能体需要**静态分析**（如 AST、调用图），仅变更文件不够。  
- 明确要求「在完整代码库上做检索 / 问答」。

提供方式可选：

- **只读 clone**：在流水线里 `git clone --depth 1 <repo> && git checkout <sha>`，将仓库目录打包为 artifact 或挂载到智能体可访问的路径；或给出 **clone_url + ref**，由智能体所在环境自行 clone。  
- **稀疏检出**：若仓库很大，可只检出本次变更涉及的目录（如 `git sparse-checkout set vega/data-connection`），再打包或暴露给智能体，减少体积。

仍建议**默认不把完整仓库作为必选输入**，而是「diff + 变更文件完整内容」为主，完整仓库按需开启。

### 7.4 与智能体输入的衔接

在《测试架构设计_基于Commit与多智能体》中，智能体 A/B 的输入包含「commit 上下文」。建议在架构中明确：

- **commit 上下文** 至少包含：commit_message、changed_files、diffs、scope/suggested_suites、affected_api_*。  
- **可选但推荐**：**变更文件的完整内容**（commit 后版本），由提交监听以 artifact 或内联方式提供；智能体先对 diff 与完整内容做「整理理解」，再执行用例增删改或测试维度分析。  
- **完整仓库**：仅在约定场景下提供，不作为默认契约。

---

## 八、内网部署方案（无法访问 GitHub）

当**开发/测试服务器在公司内网，无法访问 GitHub** 时，A、B 两种方式均不可直接使用（A 需内网服务调 GitHub API；B 的 Actions 跑在 GitHub，无法直接触发内网流水线）。可采用以下方案。

### 8.1 方案对比

| 方案 | 前置条件 | 复杂度 | 推荐度 |
|------|----------|--------|--------|
| **C1. Git 镜像 + 内网 Webhook** | 有内网 Git（GitLab/Gitee/自建） | 中 | ⭐⭐⭐ 推荐 |
| **C2. Webhook 中继 + 内网轮询** | 可部署一台有公网 URL 的中继 | 中 | ⭐⭐ |
| **C3. 定时拉取 + 本地 diff** | 内网有代理可访问 GitHub，或可定期拷入数据 | 低 | ⭐ |

### 8.2 方案 C1：Git 镜像 + 内网 Webhook（推荐）

**思路**：将 GitHub 仓库**镜像**到内网 Git，内网 Git 的 push webhook 触发监听服务；监听服务基于**本地 git diff** 计算变更，无需调 GitHub API。

```
GitHub (push) → 镜像同步任务（有外网机器）→ 内网 Git (push)
                                                    ↓
                                            内网 Webhook 触发
                                                    ↓
                                            监听服务（内网）
                                            用 git diff 计算变更
                                            产出 listener_output.json
```

**步骤**：

1. **配置镜像同步**  
   - 在一台**可访问 GitHub** 的机器（跳板机、CI 外网节点、开发本机等）上，配置定时或事件触发的镜像任务：
     - 方式 a：`git clone --mirror` + 定期 `git fetch` + `git push` 到内网 Git 仓库
     - 方式 b：内网 Git（GitLab/Gitee）的「镜像仓库」功能，从 GitHub 拉取
   - 镜像目标：内网 Git 的**研发仓库**（与 GitHub 对应）

2. **内网 Git 配置 Webhook**  
   - 在内网 Git 仓库设置 Webhook：`push` 事件 → 内网监听服务 URL（如 `http://内网服务器:port/webhooks/push`）

3. **内网监听服务**  
   - 接收 Webhook，从 payload 解析 `ref`、`before`、`after`（commit sha）
   - **不调 GitHub API**，改为在**内网仓库 clone** 上执行：
     - `git fetch` 或已是最新（镜像刚 push）
     - `git diff --name-status {before} {after}` 得到 changed_files
     - `git diff {before} {after}` 或 `git show` 得到 patch
   - 用 `scripts/commit_scope_mapping.py` 或等价逻辑做 path_scope_mapping 匹配，产出 listener_output.json
   - 将 JSON 写入共享目录、或投递到内网 MQ、或触发内网 CI，供智能体流水线消费

4. **path_scope_mapping / path_to_api**  
   - 放在内网监听服务可读路径（如测试框架 clone 或配置中心），与现有逻辑一致

**优点**：内网完全自闭环，不依赖外网；Webhook 实时性好。**缺点**：需维护镜像同步与内网 Git。

### 8.3 方案 C2：Webhook 中继 + 内网轮询

**思路**：在**有公网 URL** 的机器（云服务器、DMZ）部署中继，接收 GitHub Webhook；中继将 payload 存库或写文件；内网服务**轮询**中继拉取新事件，再调 GitHub API 取 commit 详情。

**限制**：内网服务需能访问中继；若内网完全无法访问中继，则需中继**主动推送**到内网（如内网 MQ 有外网可写入口、或 VPN 反向通道），实现较复杂。

**简化版**：中继接收 Webhook 后，直接调 GitHub API 完成 path 匹配，产出 listener_output.json，写入对象存储（如 S3、MinIO）或共享盘；内网服务轮询该存储拉取新 JSON。此时内网只需能访问该存储，无需访问 GitHub。

### 8.4 方案 C3：定时拉取 + 本地 diff

**适用**：内网有 **HTTP 代理可访问 GitHub**，或可定期从外网拷入数据。

- **有代理**：内网监听服务通过代理调 GitHub API，实现与方式 A 相同；Webhook 需 GitHub 能访问内网（端口映射/公网 IP）或通过中继转发。
- **无代理、可拷入**：外网机器定时拉取 GitHub 最新 commit，将 `listener_output.json` 或原始 diff 拷入内网共享目录；内网服务轮询该目录，发现新文件则触发流水线。实时性差，但实现最简单。

### 8.5 华为云打通内外网方案

若使用**华为云**，可利用其 VPC、NAT、OBS、CodeArts 等能力打通内外网，实现「GitHub Webhook → 华为云处理 → 内网消费」。

| 方案 | 架构 | 适用 |
|------|------|------|
| **D1. 华为云 ECS 中继 + OBS** | GitHub Webhook → 华为云 ECS（EIP）→ 调 API 产 JSON → 写 OBS；内网通过 VPN/专线接入华为云 VPC，轮询 OBS | 内网可接入华为云 VPC |
| **D2. CodeArts 镜像 + 流水线** | GitHub → CodeArts 镜像仓库；CodeArts push 触发流水线，流水线产 JSON；写 OBS 或回调内网 | 已用/可迁 CodeArts |
| **D3. 华为云全栈** | 监听 + 智能体 + 执行器均部署在华为云；内网开发服务器通过 VPN 访问华为云资源 | 测试流水线可上云 |

**D1 详细流程**：

```
GitHub (push) → Webhook → 华为云 ECS（绑定 EIP，可访问 GitHub API）
                              ↓
                         调 Compare API，path 匹配
                              ↓
                         listener_output.json 写入 OBS
                              ↓
内网开发服务器 ← VPN/专线 ← 华为云 VPC ← OBS（内网 endpoint）
     ↓
轮询 OBS 或 OBS 事件通知，拉取新 JSON，触发智能体流水线
```

**前置条件**：内网与华为云 VPC 通过 **VPN 连接** 或 **专线** 打通；华为云 OBS 使用**内网 endpoint**（如 `obs.cn-north-4.myhuaweicloud.com` 内网域名），内网 ECS/服务器可访问。

**D2 CodeArts 镜像**：CodeArts 代码托管支持「导入外部仓库」或定时同步；镜像后，CodeArts 的 push 触发 CodeArts 流水线；流水线内可调 GitHub API（若需）或直接用 `git diff`，产出 JSON 写入 OBS，供内网轮询。

### 8.6 小结（内网部署）

| 场景 | 推荐方案 |
|------|----------|
| 有内网 Git（GitLab/Gitee） | **C1**：镜像 + 内网 Webhook，监听服务用 `git diff` 本地计算 |
| 有公网中继、内网可访问中继或存储 | **C2**：中继收 Webhook、产 JSON，内网轮询 |
| **使用华为云，内网可 VPN/专线接入** | **D1**：华为云 ECS 中继 + OBS，内网轮询 OBS |
| **使用华为云 CodeArts** | **D2**：CodeArts 镜像 + 流水线产 JSON → OBS |
| 仅能定期拷入数据 | **C3**：定时拉取 + 轮询共享目录 |

---

## 九、小结

| 步骤 | 实现要点 |
|------|----------|
| **拿到变更文件** | Webhook：用 push 的 commit sha 调 `GET /repos/.../commits/{ref}` 或 compare；Actions：用 compare API 取 `base...head` 的 `files[]`。 |
| **变更类型** | `files[].status` 即 added / modified / removed。 |
| **diff** | `files[].patch`（单 commit 或 compare 均有，大文件可能无）。 |
| **scope / suggested_suites** | 用 path_scope_mapping 的 path_patterns 对 filename 做 glob 匹配。 |
| **affected_api** | 建议先做 path_to_api 映射表；可选再通过 OpenAPI 解析增强。 |
| **智能体代码上下文** | **不必默认提供完整仓库**；推荐提供 **diff + 变更文件完整内容（commit 后）**（如 artifact 包），供智能体整理理解提交后再做用例与维度分析；完整仓库仅按需提供。 |
| **输出** | 统一 JSON + 可选「智能体上下文包」artifact，供智能体 A/B/C 与执行器使用。 |
| **内网部署** | 开发服务器无法访问 GitHub 时，见 **第八节**：Git 镜像 + 内网 Webhook（推荐）、Webhook 中继、或定时拉取。 |

对接 GitHub 开源仓库时，**优先用 GitHub Actions + Compare API** 即可在无公网服务的前提下实现完整「提交监听」能力；**为智能体提供 diff + 变更文件完整内容**即可支撑「理解新提交 → 下一步分析」，无需默认提供完整代码仓库。**内网无法访问 GitHub** 时，采用第八节的 Git 镜像 + 内网 Webhook 方案，监听服务用本地 `git diff` 计算变更，无需调 GitHub API。
