# KWeaver 示例

[English](./README.md)

通过 CLI 演示 KWeaver 核心能力的端到端示例。

| 示例 | 故事 | 展示内容 |
|------|------|---------|
| [01-db-to-qa](./01-db-to-qa/) | *供应链分析师不再等 DBA 写 SQL — 数据库直接用自然语言回答问题* | MySQL → 知识网络 → 语义搜索 → Agent 对话 |
| [02-csv-to-kn](./02-csv-to-kn/) | *HR 总监散落的表格变成了可以遍历和查询的知识网络* | CSV → 知识网络 → 子图遍历 → Agent 问答 |
| [03-action-lifecycle](./03-action-lifecycle/) | *采购员 8 点到岗，今天的库存预警清单已经生成好了 — 知识网络在夜里自己完成了* | CSV → 知识网络 → 行动 → 调度 → 审计日志 |
| [04-multi-agent-session-id](./04-multi-agent-session-id/) | *平台特性巡检：自定义入参完整地从父 agent 透传到子 agent 再到 SKILL，每一步都有据可查* | Dolphin 编排 → 多 agent → 自定义入参 → SKILL 调用 |
| [05-skill-routing-loop](./05-skill-routing-loop/) | *3 个物料、3 条 critical 告警、3 条不同处置路径——每条都能在知识网络里找到依据* | MySQL → BKN (经 Vega) → find_skills → Decision Agent → Skill → Action |
| [06-world-cup](./06-world-cup/) | *分析师把 27 份公开世界杯 CSV 丢进平台，让 Agent 回答「谁是历史射手王」——不写定制 ETL* | 公开 CSV（CC-BY-SA）→ 知识网络 → Agent 问答 |

## 快速开始

每个示例独立运行。进入目录，复制 `env.sample` 为 `.env`，填写连接信息，执行脚本：

```bash
cd 01-db-to-qa
cp env.sample .env
vim .env        # 填写 DB_HOST、DB_USER、DB_PASS 等
./run.sh
```

> **安全提示：** `.env` 文件已被 gitignore 排除。请勿将含有真实凭据的 `.env` 提交到版本控制。
> 每个 `env.sample` 包含占位值和注释说明，帮助你了解每个变量的用途。

所有示例需要：
- KWeaver CLI：`npm install -g @kweaver-ai/kweaver-sdk`
- 平台登录：`kweaver auth login https://<your-platform-url>`

各示例的详细前置条件见对应 README。

## 清理

多数脚本退出时（无论成功或失败）自动删除所有创建的资源（数据源、知识网络、行动等）。

例外：`04-multi-agent-session-id` 在跑成功后**默认保留** SKILL 与三个 agent，便于在 Web UI 检视；传 `--cleanup` 即可清理。
