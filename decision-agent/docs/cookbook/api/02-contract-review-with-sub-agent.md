# 用 REST API 创建带 Sub-Agent 的合同审核流程

> - **难度**：⭐⭐ 进阶
> - **耗时**：约 25 分钟
> - **涉及模块**：`agent-factory REST`、`sub-agent`、`incremental-streaming`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个通过 REST API 管理的主 Agent，它调用已发布 Sub-Agent 完成合同风险抽取，并通过增量流式返回执行过程。

## 2. Prerequisites（前置条件）

- 已准备可用模型。
- 已能运行 `docs/user_manual/examples/api`。
- 已创建并发布 Sub-Agent，拿到 `<sub-agent-key>`。

## 3. Steps（操作步骤）

### 3.1 创建 Sub-Agent

```bash
AGENT_NAME=contract-risk-extractor make -C docs/user_manual/examples/api/agents create
export SUB_AGENT_ID="$(jq -r '.id' docs/user_manual/examples/api/.tmp/create-response.json)"
AGENT_ID="$SUB_AGENT_ID" make -C docs/user_manual/examples/api/agents publish
```

### 3.2 创建主 Agent 并引用 Sub-Agent

主 Agent 配置中增加：

```jsonc
{
  "skills": {
    "agents": [
      {
        "agent_key": "<sub-agent-key>",
        "agent_version": "latest",
        "agent_timeout": 60
      }
    ]
  }
}
```

### 3.3 使用增量流式调用主 Agent

```bash
AGENT_ID=<main-agent-id> AGENT_KEY=<main-agent-key> \
  make -C docs/user_manual/examples/api chat-incremental-stream
```

增量流式的每个 `data:` 都是 patch：

```jsonc
{ "seq_id": 3, "key": ["message", "content"], "content": "新增文本", "action": "append" }
```

客户端按 `seq_id` 顺序消费，根据 `key` 定位响应对象路径，根据 `action` 合并：

| action | 用法 |
| --- | --- |
| `append` | 字符串前缀增长时追加 `content`。 |
| `upsert` | 新字段或结构变化时写入/替换 `content`。 |
| `remove` | 删除 `key` 指向的字段。 |
| `end` | `key: []` 表示流结束。 |

## 4. Expected output（期望输出）

> **判定成功的依据**：流中出现 `append` 或 `upsert` patch，最终文本包含合同风险条款和建议。

```jsonc
{
  "seq_id": 10,
  "key": ["progress"],
  "action": "upsert",
  "content": [{ "skill_info": { "type": "agent" }, "status": "success" }]
}
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| `agent_key is required` | Sub-Agent 配置缺少 `agent_key` | 先发布 Sub-Agent，并写入发布后的 `agent_key`。 |
| 客户端文本重复 | 把 `append` 当完整快照处理 | 增量流式只追加 patch 内容，不要整帧覆盖文本。 |
| 看不到 Sub-Agent 进度 | 主 Agent 指令未触发技能调用 | 在 system prompt 或 Dolphin 中明确调用 Sub-Agent。 |

## 6. See also（延伸阅读）

- 场景说明：[带 Sub-Agent 的合同审核流程](../scenario/02-contract-review-with-sub-agent.md)
- 用户手册：[增量流式](../../user_manual/api/incremental-streaming.md)
- API 示例：[chat/incremental-stream.sh](../../user_manual/examples/api/chat/incremental-stream.sh)

