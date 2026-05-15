# 创建带 Sub-Agent 的合同审核流程

> - **难度**：⭐⭐ 进阶
> - **耗时**：约 25 分钟
> - **涉及模块**：`agent`、`sub-agent`、`chat`、`incremental-streaming`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个主合同审核 Agent，它会把“抽取风险条款”委托给已发布的 Sub-Agent，再汇总为可读审核意见。

## 2. Prerequisites（前置条件）

- 已准备可用模型：`KWEAVER_LLM_ID` 和 `KWEAVER_LLM_NAME`。
- 已理解发布语义：Sub-Agent 必须先发布，主 Agent 才能通过 `agent_key` 引用它。
- 若使用 API 版本，建议使用增量流式观察主 Agent 调用 Sub-Agent 的过程。

## 3. Steps（操作步骤）

### 3.1 创建并发布 Sub-Agent

Sub-Agent 专注识别合同风险条款，输出结构化风险项。

```jsonc
{
  "name": "contract-risk-extractor",
  "config": {
    "input": { "fields": [{ "name": "query", "type": "string" }] },
    "output": { "default_format": "json", "variables": { "answer_var": "answer" } }
  }
}
```

### 3.2 创建主 Agent 并集成 Sub-Agent

主 Agent 的 `skills.agents[]` 引用 Sub-Agent 的 `agent_key`：

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

### 3.3 调用主 Agent 并验证 Sub-Agent 调用

在流式 progress 或最终响应中检查是否出现 Sub-Agent 相关执行步骤。

## 4. Expected output（期望输出）

> **判定成功的依据**：主 Agent 返回中包含风险条款、原因和建议动作，并能在进度信息中看到子 Agent 调用痕迹。

```jsonc
{
  "message": {
    "content": "发现 3 条高风险条款..."
  },
  "progress": [
    { "skill_info": { "type": "agent" }, "status": "success" }
  ]
}
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| Sub-Agent 调用失败 | 子 Agent 未发布或 `agent_key` 错误 | 先发布 Sub-Agent，并重新写入主 Agent 配置。 |
| 主 Agent 只直接回答 | Dolphin/角色指令没有引导调用子 Agent | 在主 Agent 指令中明确“先调用风险抽取 Agent”。 |
| 流式内容难以合并 | 使用了增量流式但客户端未实现 patch 合并 | 参考 API Recipe 的增量流式算法。 |

## 6. See also（延伸阅读）

- API Recipe：[api/02-contract-review-with-sub-agent.md](../api/02-contract-review-with-sub-agent.md)
- CLI Recipe：[cli/02-contract-review-with-sub-agent.md](../cli/02-contract-review-with-sub-agent.md)
- SDK Recipe：[sdk/typescript/02-contract-review-with-sub-agent.md](../sdk/typescript/02-contract-review-with-sub-agent.md)
- 用户手册：[增量流式](../../user_manual/api/incremental-streaming.md)

