# 创建合同摘要 Agent

> - **难度**：⭐ 入门
> - **耗时**：约 15 分钟
> - **涉及模块**：`agent`、`chat`、`conversation`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个可以接收合同文本并输出摘要、风险点和建议动作的普通 Agent；它不依赖 Sub-Agent，适合作为最小接入样例。

## 2. Prerequisites（前置条件）

- 本地 Decision Agent 服务已启动，或已准备远程 `KWEAVER_BASE_URL`。
- 已准备可用模型：`KWEAVER_LLM_ID` 和 `KWEAVER_LLM_NAME`。
- 已选择接入方式：REST API、CLI 或 TypeScript SDK。

## 3. Steps（操作步骤）

### 3.1 创建 Agent 配置

配置侧只需要输入字段、输出字段和默认模型：

```jsonc
{
  "input": { "fields": [{ "name": "query", "type": "string" }] },
  "output": { "default_format": "markdown", "variables": { "answer_var": "answer" } },
  "llms": [{ "is_default": true, "llm_config": { "id": "<llm-id>", "name": "<llm-name>" } }]
}
```

### 3.2 发布并执行

创建后发布到广场，再用发布后的 `agent_key` 发起对话。第二轮对话传入上一轮返回的 `conversation_id`。

```bash
# API、CLI、SDK 版本见“延伸阅读”中的具体示例。
```

### 3.3 管理对话

根据 `agent_key` 查询对话列表，根据 `conversation_id` 查询详情；如果需要可以更新标题或标记已读。

## 4. Expected output（期望输出）

> **判定成功的依据**：第一次调用返回 `conversation_id`，第二次调用携带同一个 `conversation_id` 后能继续上下文。

```jsonc
{
  "conversation_id": "01...",
  "message": {
    "content": "合同摘要..."
  }
}
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| `llms is empty` 或模型相关错误 | 未配置默认模型 | 设置 `KWEAVER_LLM_ID`、`KWEAVER_LLM_NAME` 后重新创建。 |
| 对话返回 404 | 使用了未发布版本或错误 `agent_key` | 先发布 Agent，再用发布信息中的 `agent_key` 调用。 |
| 第二轮没有上下文 | 没有传 `conversation_id` | 从第一轮响应中提取并传回。 |

## 6. See also（延伸阅读）

- API Recipe：[api/01-contract-summary-basic.md](../api/01-contract-summary-basic.md)
- CLI Recipe：[cli/01-contract-summary-basic.md](../cli/01-contract-summary-basic.md)
- SDK Recipe：[sdk/typescript/01-contract-summary-basic.md](../sdk/typescript/01-contract-summary-basic.md)
- 用户手册：[对话与会话](../../user_manual/api/chat-and-conversation.md)
