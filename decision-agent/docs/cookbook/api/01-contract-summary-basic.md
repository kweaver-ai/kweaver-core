# 用 REST API 创建合同摘要 Agent

> - **难度**：⭐ 入门
> - **耗时**：约 15 分钟
> - **涉及模块**：`agent-factory REST`、`chat`、`conversation`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个通过 REST API 创建、发布并调用的合同摘要 Agent，同时掌握如何用 `conversation_id` 继续多轮对话。

## 2. Prerequisites（前置条件）

- `AF_BASE_URL` 指向 Decision Agent 服务。
- no-auth 本地环境设置 `AF_TOKEN=__NO_AUTH__` 或 `KWEAVER_NO_AUTH=1`。
- 已设置 `KWEAVER_LLM_ID` 和 `KWEAVER_LLM_NAME`。

## 3. Steps（操作步骤）

### 3.1 生成最小配置

```bash
make -C docs/user_manual/examples/api agent-config-minimal
```

### 3.2 创建、发布、删除完整流程

```bash
make -C docs/user_manual/examples/api flow
```

### 3.3 调用并继续对话

```bash
AGENT_ID=<agent-id> AGENT_KEY=<agent-key> make -C docs/user_manual/examples/api chat-non-stream
CONVERSATION_ID=<conversation-id> AGENT_KEY=<agent-key> make -C docs/user_manual/examples/api conversations
```

## 4. Expected output（期望输出）

> **判定成功的依据**：创建响应包含 `id`，发布响应成功；对话响应包含 `conversation_id`。

```jsonc
{
  "id": "01...",
  "version": "v0"
}
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| `Please set KWEAVER_LLM_ID` | 未设置模型 | 设置 `KWEAVER_LLM_ID` 和 `KWEAVER_LLM_NAME`。 |
| `401 Unauthorized` | 鉴权模式和 token 不匹配 | no-auth 设置 `KWEAVER_NO_AUTH=1`；鉴权环境设置 `AF_TOKEN`。 |
| `CONVERSATION_ID` 为空 | 还没有执行第一轮对话 | 先运行 `chat-non-stream` 并从响应中取值。 |

## 6. See also（延伸阅读）

- 场景说明：[创建合同摘要 Agent](../scenario/01-contract-summary-basic.md)
- 示例目录：[API examples](../../user_manual/examples/api/README.md)
- 用户手册：[Agent 生命周期 API](../../user_manual/api/agent-lifecycle.md)

