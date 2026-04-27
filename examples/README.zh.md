# KWeaver 示例

[English](./README.md)

通过 CLI 演示 KWeaver 核心能力的端到端示例。

| 示例 | 故事 | 展示内容 |
|------|------|---------|
| [01-db-to-qa](./01-db-to-qa/) | *供应链分析师不再等 DBA 写 SQL — 数据库直接用自然语言回答问题* | MySQL → 知识网络 → 语义搜索 → Agent 对话 |
| [02-csv-to-kn](./02-csv-to-kn/) | *HR 总监散落的表格变成了可以遍历和查询的知识网络* | CSV → 知识网络 → 子图遍历 → Agent 问答 |
| [03-action-lifecycle](./03-action-lifecycle/) | *采购员 8 点到岗，今天的库存预警清单已经生成好了 — 知识网络在夜里自己完成了* | CSV → 知识网络 → 行动 → 调度 → 审计日志 |

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

脚本退出时（无论成功或失败）自动删除所有创建的资源（数据源、知识网络、行动等）。
