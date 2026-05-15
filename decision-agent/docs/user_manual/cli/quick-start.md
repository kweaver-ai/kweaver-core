# CLI 快速开始

本模块用 CLI 跑通创建、详情、发布、对话。请先完成 [CLI 安装和运行](./install-and-run.md)，并准备 `AGENT_CONFIG`。

## 1. 创建 Agent

```bash
export AGENT_NAME="doc_cli_agent_$(date +%Y%m%d%H%M%S)"

kweaver agent create \
  --name "$AGENT_NAME" \
  --profile "Decision Agent CLI user guide local example" \
  --product-key dip \
  --config "$AGENT_CONFIG" \
  --pretty | tee /tmp/kweaver-cli-agent-create.json

export AGENT_ID="$(jq -r '.id' /tmp/kweaver-cli-agent-create.json)"
echo "$AGENT_ID"
```

## 2. 获取详情

```bash
test -n "$AGENT_ID"

kweaver agent get "$AGENT_ID" \
  --pretty | tee /tmp/kweaver-cli-agent-detail.json
```

## 3. 发布

```bash
test -n "$AGENT_ID"

kweaver agent publish "$AGENT_ID" \
  | tee /tmp/kweaver-cli-agent-publish.json
```

## 4. 非流式对话

发布后按版本对话。如果还没有发布版本，可先使用 `--version v0` 测试未发布版本。

```bash
test -n "$AGENT_ID"
export AGENT_VERSION="${AGENT_VERSION:-v0}"

kweaver agent chat "$AGENT_ID" \
  --version "$AGENT_VERSION" \
  -m "请用一句话介绍你自己" \
  --no-stream \
  --verbose | tee /tmp/kweaver-cli-chat.json
```

## 5. 对话列表

```bash
kweaver agent sessions "$AGENT_ID" \
  --limit 10 \
  --pretty
```

`sessions` 是 CLI 的历史命令名，这里查询的是当前 Agent 的对话列表。

## 6. 清理

```bash
test -n "$AGENT_ID"

kweaver agent unpublish "$AGENT_ID" || true

kweaver agent delete "$AGENT_ID" -y
```
