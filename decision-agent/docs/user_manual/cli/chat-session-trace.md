# CLI 对话与 Trace

请先完成 [CLI 安装和运行](./install-and-run.md)，并准备 `AGENT_ID`。完整可运行脚本见 [../examples/cli](../examples/cli/README.md)。

## `kweaver agent chat`

用途：与 Agent 对话。省略 `-m/--message` 时进入交互式模式；带 `-m` 时发送单轮非交互消息。

```bash
test -n "$AGENT_ID"
export AGENT_VERSION="${AGENT_VERSION:-v0}"

kweaver agent chat "$AGENT_ID" \
  --version "$AGENT_VERSION" \
  -m "请用一句话介绍你自己" \
  --no-stream \
  --verbose | tee /tmp/kweaver-agent-chat.json
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。CLI 会先解析对应版本的 `agent_key`。 |
| `-m, --message <text>` | 否 | 无 | 单轮消息；省略时进入交互式对话。 |
| `--version <value>` / `-version <value>` | 否 | `v0` | 用于解析 `agent_key` 的 Agent 版本。 |
| `--conversation-id <id>` | 否 | 无 | 继续已有对话。 |
| `-cid <id>` | 否 | 无 | `--conversation-id` 短别名。 |
| `--conversation_id <id>` | 否 | 无 | 兼容别名。 |
| `-conversation-id <id>` | 否 | 无 | 兼容别名。 |
| `-conversation_id <id>` | 否 | 无 | 兼容别名。 |
| `--session-id <id>` | 否 | 无 | 历史兼容别名，含义同 `conversation_id`。 |
| `--stream` | 否 | 交互式默认开启 | 启用流式输出。 |
| `--no-stream` | 否 | 非交互默认关闭流式 | 禁用流式输出。 |
| `--tui` | 否 | 关闭 | 启动实验性 OpenTUI 界面，交互式使用。 |
| `--verbose, -v, --debug` | 否 | 关闭 | 输出请求细节或调试信息。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |

背后 API：CLI 先通过广场详情接口解析 `agent_key`，再调用对话接口。

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/agent-market/agent/$AGENT_ID/version/$AGENT_VERSION?is_visit=true"

curl -X POST "$KWEAVER_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/chat/completion" \
  -H "Content-Type: application/json" \
  -d '{"query":"请用一句话介绍你自己","conversation_id":"","stream":false}'
```

对应示例：

```bash
AGENT_ID=<agent-id> make -C docs/user_manual/examples/cli chat-examples
```

## 继续对话

从上一步输出或 `sessions` 命令中取得 `conversation_id`：

```bash
: "${CONVERSATION_ID:?请先从上一步 chat 输出或 sessions 结果导出 CONVERSATION_ID}"

kweaver agent chat "$AGENT_ID" \
  --version "$AGENT_VERSION" \
  --conversation-id "$CONVERSATION_ID" \
  -m "请继续补充一个使用建议" \
  --no-stream
```

## 流式输出

```bash
kweaver agent chat "$AGENT_ID" \
  --version "$AGENT_VERSION" \
  -m "请分三点说明你的能力" \
  --stream
```

## 交互式对话

```bash
kweaver agent chat "$AGENT_ID" \
  --version "$AGENT_VERSION"
```

退出方式：输入 `exit`、`quit` 或 `q`。

## `kweaver agent sessions`

用途：查询 Agent 的对话列表。`sessions` 是 CLI 的历史命令名，这里查询的是对话列表，不是底层运行期 session。

```bash
kweaver agent sessions "$AGENT_ID" \
  --limit 10 \
  --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。CLI 会先解析 `agent_key`。 |
| `--limit <n>` | 否 | `30` | 返回对话数。小于 1 或非法值会回退为 `30`。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--compact` | 否 | 关闭 | 紧凑 JSON 输出。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation?page=1&size=10" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`AGENT_ID=<agent-id> make -C docs/user_manual/examples/cli chat-examples`。

## `kweaver agent history`

用途：查询某个对话的消息历史。

```bash
test -n "$CONVERSATION_ID"

kweaver agent history "$AGENT_ID" "$CONVERSATION_ID" \
  --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 条件必填 | 无 | 推荐传入。用于解析 `agent_key`。 |
| `<conversation_id>` | 是 | 无 | 对话 ID。也兼容只传 `conversation_id` 的历史形式。 |
| `--limit <n>` | 否 | `30` | 兼容参数；当前消息详情接口不依赖该参数。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--compact` | 否 | 关闭 | 紧凑 JSON 输出。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v1/app/$AGENT_KEY/conversation/$CONVERSATION_ID" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`AGENT_ID=<agent-id> CONVERSATION_ID=<conversation-id> make -C docs/user_manual/examples/cli chat-examples`。

## `kweaver agent trace`

用途：查看一次对话中的执行链路、工具调用与中间进度。这个能力依赖当前环境暴露的 trace 路由；如果后端未开放对应路由（或 trace-ai 服务不可用），CLI 会返回 404，此时可先使用 `history` 查看对话消息。

```bash
test -n "$CONVERSATION_ID"

kweaver agent trace "$AGENT_ID" "$CONVERSATION_ID" \
  --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |
| `<conversation_id>` | 是 | 无 | 对话 ID。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--compact` | 否 | 关闭 | 紧凑 JSON 输出。 |

背后 API：

```bash
curl -X POST "$KWEAVER_BASE_URL/api/agent-factory/v1/observability/agent/$AGENT_ID/conversation/$CONVERSATION_ID/session" \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"'"$AGENT_ID"'","start_time":1,"end_time":<now_plus_one_day_ms>,"page":1,"size":50}'
```

对应示例：`AGENT_ID=<agent-id> CONVERSATION_ID=<conversation-id> make -C docs/user_manual/examples/cli chat-examples`。
