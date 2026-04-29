# multi-agent-session-id

**证明 `session_id` 可以从用户请求穿透 father agent → son agents → SKILL，全链路零丢失。**

---

## 它证明了什么

KWeaver 平台支持通过 `custom_querys.session_id` 向 agent 注入自定义会话标识。这个 demo 验证：
该字段不仅由 father agent 接收到，还能被 father 的 dolphin DSL 显式传递给下游的 son agents，最终被 son 调用的 SKILL (`exp_session_echo`) 原文回显出来。

整个链路没有任何硬编码的 session_id 值——SKILL 必须从运行时的 `input.session_id` 读取并原文输出。
两条独立的 shell 断言（见下方"校验做了什么"）从平台返回的响应 JSON 中机械地检查这一点，
不依赖 LLM 描述，也不依赖人工肉眼核对。

---

## 架构

```
user
  │  POST /api/agent-factory/v1/app/<father_key>/chat/completion
  │  body: { agent_id, agent_key, agent_version, query,
  │          custom_querys: { session_id: "DEMO-XXX" }, stream: false }
  ▼
exp_father  (is_dolphin_mode=1)
  dolphin DSL:
    @exp_son1(session_id=$session_id, query=$query) -> res_1
    @exp_son2(session_id=$session_id, query=$query) -> res_2
    {"res_1": $res_1, "res_2": $res_2} -> answer
  │
  ├─► exp_son1  (is_dolphin_mode=1)
  │     dolphin DSL:
  │       "[input.session_id=" + $session_id + "] " + $query -> q
  │       /explore/(history=true) <prompt>\n$q\n -> answer
  │     skills.skills = [{ skill_id: exp_session_echo }]
  │     skills.tools  = [list_skills_v2 with X-Authorization mapping]
  │     │
  │     ├── list_skills_v2()       → 返回可用 skill 列表
  │     ├── builtin_skill_load()   → 加载 exp_session_echo 内容
  │     └── LLM 输出: "[exp_session_echo] RECEIVED session_id=DEMO-XXX from exp_son1"
  │
  └─► exp_son2  (配置与 exp_son1 完全相同，仅 name/profile 不同)
        同样链路 → "[exp_session_echo] RECEIVED session_id=DEMO-XXX from exp_son2"

响应路径:
  .message.content.final_answer.answer_type_other.{res_1,res_2}.answer.answer
  .message.content.final_answer.answer_type_other.{res_1,res_2}.answer.input_message
```

---

## 怎么跑

### 前提

- `kweaver` CLI 已通过 `kweaver auth login` 认证到 `admin` 平台（`https://115.190.186.186`）
- 环境中有 `jq`、`bash`、`curl`

### 三条命令

```bash
# 默认：检测现有产物（已存在则复用），调用 father，跑两条断言
./run.sh

# 指定自定义 session_id（默认随机生成 DEMO-2026-XXXXXX）
./run.sh --session-id MY-CUSTOM-ID

# 清理：在平台上 unpublish + 删除三个 agent 和 skill（不可逆，慎用）
./run.sh --cleanup
```

> **警告**：`--cleanup` 会永久删除平台上的 `exp_father`、`exp_son1`、`exp_son2`、`exp_session_echo`。
> 这些产物目前被保留供人工验证，请勿随意清理。

### 成功时的输出尾部

```
[exp] session_id=DEMO-2026-XXXXXX
[exp] skill_id=30537b46-b510-4964-b235-eb529394d00e (reused if existed)
[exp] agent exp_son1 already exists (01KQA4JH5QNBJBD42KN43458FW), reusing as-is
[exp] agent exp_son2 already exists (01KQA4MD1K463BDK6F10WGBRT5), reusing as-is
[exp] agent exp_father already exists (01KQA4NZXQFQP6EJMXSGH2S8DM), reusing as-is
[exp] invoking father with custom_querys.session_id=DEMO-2026-XXXXXX...
[exp] conversation_id=<conv_id>, elapsed=33s
[exp] son1 answer (first 200 chars):
[exp_session_echo] RECEIVED session_id=DEMO-2026-XXXXXX from exp_son1 ...
[exp] son2 answer (first 200 chars):
[exp_session_echo] RECEIVED session_id=DEMO-2026-XXXXXX from exp_son2 ...
[exp] assert_literal: exp_son1 — PASS
[exp] assert_literal: exp_son2 — PASS
[exp] assert_propagation: exp_son1 input contains [input.session_id=DEMO-2026-XXXXXX] — PASS
[exp] assert_propagation: exp_son2 input contains [input.session_id=DEMO-2026-XXXXXX] — PASS
[exp] ALL ASSERTIONS PASSED. session_id=DEMO-2026-XXXXXX, conversation=<conv_id>
[exp] (artifacts kept on platform; run with --cleanup to remove)
```

---

## 校验做了什么

校验逻辑在 `lib/verify.sh`，两条断言都从 `/tmp/exp_run_resp.json` 读取平台响应。

### 断言 1：`assert_literal_in_answer`

检查每个 son 的文本回答中是否包含：

```
[exp_session_echo] RECEIVED session_id=<sid> from <son_name>
```

这一行由 SKILL（`exp_session_echo`）严格按照 `SKILL.md` 格式输出。
如果平台没有把 `session_id` 传进 son 的上下文，SKILL 就无法知道它——
这条断言排除了 LLM 猜测或硬编码的可能。

### 断言 2：`assert_session_id_in_son_input`

检查每个 son 在响应中的 `input_message` 字段是否包含：

```
[input.session_id=<sid>]
```

这一段前缀是 son 的 dolphin DSL 主动拼接的：

```
"[input.session_id=" + $session_id + "] " + $query -> q
```

`input_message` 是平台记录的"son 实际收到的 query 字符串"，不经过 LLM 加工。
该断言证明平台确实把 father 传来的 `session_id` 路由进了 son 的执行上下文，
而不是 LLM 在回答中自行编造。

---

## 现网产物（保留状态）

以下产物已部署在 `https://115.190.186.186`（business domain: `bd_public`），**请勿删除**：

| 类型   | 名称               | ID                                       | 额外信息                       |
|--------|--------------------|------------------------------------------|--------------------------------|
| SKILL  | `exp_session_echo` | `30537b46-b510-4964-b235-eb529394d00e`   | status: published              |
| Agent  | `exp_son1`         | `01KQA4JH5QNBJBD42KN43458FW`            | key: `01KQA4JH5QNBJBD42KN2K1S5JK`, v1 |
| Agent  | `exp_son2`         | `01KQA4MD1K463BDK6F10WGBRT5`            | key: `01KQA4MD1K463BDK6F10CZFMM1`, v1 |
| Agent  | `exp_father`       | `01KQA4NZXQFQP6EJMXSGH2S8DM`            | key: `01KQA4NZXQFQP6EJMXSD36ZBXD`, v1 |

Web UI：`https://115.190.186.186/dip-hub/studio/digital-human`
（在 agent 列表中按名称搜索 `exp_father` / `exp_son1` / `exp_son2`）

---

## 看证据在哪

运行 `./run.sh` 后：

**完整响应 JSON**

```bash
cat /tmp/exp_run_resp.json | jq .
```

两个 son 的回答分别在：

```bash
jq '.message.content.final_answer.answer_type_other.res_1.answer.answer' /tmp/exp_run_resp.json
jq '.message.content.final_answer.answer_type_other.res_2.answer.answer' /tmp/exp_run_resp.json
```

son 收到的实际 input（含 `[input.session_id=...]` 前缀）：

```bash
jq '.message.content.final_answer.answer_type_other.res_1.answer.input_message' /tmp/exp_run_resp.json
```

**查看平台上 agent 的当前配置**

```bash
kweaver agent get 01KQA4NZXQFQP6EJMXSGH2S8DM --verbose   # exp_father
kweaver agent get 01KQA4JH5QNBJBD42KN43458FW  --verbose   # exp_son1
```

> 注意：`kweaver agent trace <agent_id> <conv_id>` 当前返回 HTTP 500（平台 Trace API 故障），
> 不推荐使用。所有需要的证据都在 chat completion 的响应 JSON 中。

---

## 平台行为踩坑记录

以下是构建过程中发现的平台非直觉行为，记录供后续排查参考：

- **`custom_querys.session_id` 是唯一有效路径**：自定义输入字段必须放在请求体的 `custom_querys` 对象内；直接放在顶层的 body 字段不会被路由进 agent 的 `input.*`。`kweaver agent chat` 命令不支持传 `custom_querys`，必须用 `kweaver call -d <body>`。

- **`is_dolphin_mode=0` 会忽略 `pre_dolphin`**：在纯 prompt 模式下，平台会用 system_prompt 自动合成 dolphin 程序，忽略用户设置的 `pre_dolphin` 数组。要让 `session_id`（或任何自定义输入字段）出现在 LLM 的上下文里，必须用 `is_dolphin_mode=1` 加显式 dolphin DSL。

- **`list_skills_v2` 必须配置 X-Authorization 映射**：工具箱 `tmp_skill_discovery_R1` 的 `list_skills_v2` 工具（id `51382ef3-b35b-44a6-8a53-c670cbf53f10`）在 agent 调用时会返回 401，除非 agent 配置中为该工具的 `tool_input` 添加了 `X-Authorization` → `header.authorization` 的 `map_type=var` 映射。

- **`kweaver agent get` 默认返回精简字段**：不带 `--verbose` 时只返回 `{id, name, description, status, kn_ids}`。要看完整的 `.config`（包括 dolphin DSL、skills 等），必须加 `--verbose`。

- **Trace API 当前不可用**：`kweaver agent trace <agent_id> <conv_id>` 返回 HTTP 500（Uniquery DataView 错误）。用 chat completion 响应 JSON 替代——所有验证所需数据都在里面。

- **LLM id 路径是 `.llms[0].llm_config.id`**：平台的 base agent 模板中，LLM 配置的 id 存在 `.llms[0].llm_config.id`，而非 `.llms[0].id`（后者为空）。jq 模板和 `common.sh` 均按前者取值。

---

## 前提

- `kweaver` CLI 已通过 `kweaver auth login` 认证到 `admin` 平台
- `jq` >= 1.6
- `bash` >= 4
- `curl`（由 `kweaver call` 内部使用）

---

## 设计文档

[`docs/superpowers/plans/2026-04-28-exp-multi-agent-session-id.md`](../../docs/superpowers/plans/2026-04-28-exp-multi-agent-session-id.md)

注意：计划文档大体准确，但部分章节（特别是 SKILL 调用机制、`list_skills_v2` 的鉴权要求、dolphin DSL 的 `is_dolphin_mode` 约束）在实际构建中有经验性修正。
**README 和 git commit history 是最终权威来源。**
