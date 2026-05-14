# 用 CLI 创建带 Sub-Agent 的合同审核流程

> - **难度**：⭐⭐ 进阶
> - **耗时**：约 25 分钟
> - **涉及模块**：`kweaver agent`、`sub-agent`、`config`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个通过 CLI 创建的主 Agent，它在配置中引用已发布 Sub-Agent。

## 2. Prerequisites（前置条件）

- 已能运行 `make -C docs/user_manual/examples/cli flow`。
- 已准备子 Agent 配置文件和主 Agent 配置文件。
- 已发布子 Agent，并取得 `SUB_AGENT_KEY`。

## 3. Steps（操作步骤）

### 3.1 创建并发布 Sub-Agent

```bash
kweaver agent create \
  --name "contract-risk-extractor" \
  --profile "抽取合同风险条款" \
  --product-key dip \
  --config ./sub-agent-config.json \
  --pretty

kweaver agent publish <sub-agent-id>
```

### 3.2 创建主 Agent

主 Agent 的配置文件包含：

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

```bash
kweaver agent create \
  --name "contract-review-main" \
  --profile "调用 Sub-Agent 进行合同审核" \
  --product-key dip \
  --config ./main-agent-config.json \
  --pretty
```

### 3.3 调用主 Agent

```bash
kweaver agent chat <main-agent-id> \
  --message "请审核这段合同文本：<contract text>" \
  --no-stream \
  --verbose
```

## 4. Expected output（期望输出）

> **判定成功的依据**：主 Agent 返回包含风险条款、原因和建议动作的审核结果。

```text
发现以下风险条款：...
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| 主 Agent 未调用 Sub-Agent | 指令或 Dolphin 未描述调用策略 | 在主 Agent 指令中明确“先调用合同风险抽取 Agent”。 |
| `agent_key is required` | 主配置未写 `skills.agents[].agent_key` | 发布子 Agent 后写入 `agent_key`。 |
| 需要观察增量流式 | CLI 当前没有独立 inc_stream 开关 | 使用 API Recipe 中的增量流式版本。 |

## 6. See also（延伸阅读）

- 场景说明：[带 Sub-Agent 的合同审核流程](../scenario/02-contract-review-with-sub-agent.md)
- API Recipe：[api/02-contract-review-with-sub-agent.md](../api/02-contract-review-with-sub-agent.md)
- CLI 示例：[examples/cli/agents](../../user_manual/examples/cli/agents/)

