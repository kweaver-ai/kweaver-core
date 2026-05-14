# 用 CLI 查看带人工干预的合同审核流程

> - **难度**：⭐⭐⭐ 专家
> - **耗时**：约 35 分钟
> - **涉及模块**：`kweaver agent`、`intervention`、`REST fallback`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个 CLI 视角的人工干预流程说明：配置和普通执行用 `kweaver`，当前 CLI 未覆盖的终止、续连、对话更新能力明确回落到 REST API。

## 2. Prerequisites（前置条件）

- 已完成 Sub-Agent 场景。
- 主 Agent 配置中已写入 `intervention: true`。
- 已了解 CLI 当前只覆盖 `chat`、`sessions`、`history`、`trace` 等命令。

## 3. Steps（操作步骤）

### 3.1 创建带干预配置的主 Agent

```bash
kweaver agent create \
  --name "contract-review-human-check" \
  --profile "带人工干预的合同审核 Agent" \
  --product-key dip \
  --config ./main-agent-intervention-config.json \
  --pretty
```

### 3.2 执行并查看对话

```bash
kweaver agent chat <main-agent-id> \
  --message "请审核这段合同文本：<contract text>" \
  --no-stream \
  --verbose

kweaver agent sessions <main-agent-id> --limit 10 --pretty
kweaver agent history <main-agent-id> <conversation-id> --pretty
```

### 3.3 使用 REST API 补齐终止和恢复

当前 CLI 未暴露独立的 terminate / resume 命令，请使用 API 示例：

```bash
AGENT_KEY=<main-agent-key> CONVERSATION_ID=<conversation-id> \
  make -C docs/user_manual/examples/api chat-terminate
```

## 4. Expected output（期望输出）

> **判定成功的依据**：CLI 能创建、发布、调用和查看对话；终止和恢复能力通过 REST API 示例验证。

```text
To continue this conversation, rerun the command with --conversation-id:
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| 找不到 terminate 命令 | 当前 CLI 未暴露该能力 | 使用 REST API Recipe。 |
| history 无结果 | 没有传正确 `conversation_id` | 从 chat 输出或 sessions 中获取。 |
| trace 404 | trace 服务或路由未开启 | 先用 `history` 查看消息详情。 |

## 6. See also（延伸阅读）

- 场景说明：[人工干预和终止恢复流程](../scenario/03-contract-review-intervention-termination.md)
- API Recipe：[api/03-contract-review-intervention-termination.md](../api/03-contract-review-intervention-termination.md)
- CLI 示例：[examples/cli/chat](../../user_manual/examples/cli/chat/)

