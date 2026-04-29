# 04 · Skill Routing Loop — 业务知识网络驱动的 Skill 治理

> [English](./README.md)

> 3 个物料触发同样的库存告警，Decision Agent 给出 3 条不同处置路径——
> 每条都能在业务知识网络里找到依据。

## 故事

续作 03 那位采购工程师：她现在看到每张告警单上已经写好了处置方案。3 个物料、
3 条不同路径，**没改一行 prompt**。BKN 里的 `applicable_skill` 关系是 Skill 路由的
**唯一真相源**——Agent 只能从 `find_skills` 返回的候选集里挑，没有别的杠杆。

## 这个 example 展示什么

5 个组件协同跑通一个完整闭环：

| 组件 | 职责 |
|---|---|
| **execution-factory** | 注册 / 版本化 3 个 Skill 包 |
| **业务知识网络（BKN）** | 通过 `applicable_skill` 关系把 Skill 绑到物料 |
| **Vega** | 把 BKN ObjectType 映射到 MySQL 表（读多写少） |
| **context-loader (`find_skills`)** | 按物料实例召回适用的 Skill |
| **Decision Agent** | 读 KN 证据 → 选 Skill → 执行 → 审计 |

## 前置条件

- `kweaver` CLI ≥ 0.7.1
- 启用了 Decision Agent + execution-factory + Vega 的 KWeaver 平台
  （先 `kweaver auth login <平台地址> [--insecure]`）
- **平台能访问到**的 MySQL（不是你笔记本上的），且账号有 CREATE/INSERT/SELECT/UPDATE 权限
- `python3`（依赖 Flask + mysql-connector-python，
  `pip install -r tool_backend/requirements.txt`）
- 平台模型工厂里注册的 LLM 模型（用
  `kweaver call /api/mf-model-manager/v1/llm/list` 拿 model_id）

## 快速开始

```bash
cd examples/05-skill-routing-loop
cp env.sample .env
vim .env                                    # 填 PLATFORM_HOST、LLM_ID、DB_*
pip install -r tool_backend/requirements.txt
./run.sh                                    # 端到端约 5 分钟
./run.sh --bonus                            # 跑 Bonus 段
```

> **并发注意：** 请不要同时运行两个 `./run.sh` 实例。脚本使用固定的 `KN_ID`
> （`ex05_skill_routing`），第二个实例的清理逻辑会把第一个实例的 KN 一起删掉。

## 你会看到什么

| 物料 | KN 证据 | DA 选 | 原因 |
|---|---|---|---|
| MAT-001 | 绑定 `substitute_swap`；SUB-001A/B 有库存 | substitute_swap | Python 算法打分挑替代料 → 调 MES |
| MAT-002 | 绑定 `supplier_expedite`；SUP-2 capability=expedite | supplier_expedite | 供应商能加急 → POST 供应商门户 |
| MAT-003 | 只绑定 `standard_replenish` | standard_replenish | 默认路径 → 走 ERP 下单 |

## Bonus — 改业务 → KN 重建 → AI 跟着变

`./run.sh --bonus` 会调 mock 业务系统的 admin 端点，把 MAT-002 的绑定 Skill
从 `supplier_expedite` 改成 `standard_replenish`（直接 UPDATE
`materials.bound_skill_id`，由 `applicable_skill` 的 direct-mapping FK 决定边），
然后触发一次 `kweaver bkn build` 重新物化 `applicable_skill` 边，再让 Agent
重新处理 MAT-002。Decision Agent 下一次 `find_skills` 拿到的就是新候选集，
自动切到 `standard_replenish`——**没改 prompt、没重新部署任何服务**。

> **为什么需要重建：** `applicable_skill` 是关系（relation），它的边在
> KN build 时物化进图，不是 live-mapped。ObjectType 的 data property
> （比如 `supplier.capability`）是 Vega 直接 live-read MySQL 的，但
> 关系边需要重新 build 才会刷新。把重建作为显式步骤写进脚本，
> 让"业务变更 → KN 同步 → AI 跟随"的反馈环对学习者完整可见。

## 原理细节

完整设计文档：[`docs/superpowers/specs/2026-04-27-skill-routing-loop-example-design.md`](../../docs/superpowers/specs/2026-04-27-skill-routing-loop-example-design.md)

包括：
- BKN schema 和 `applicable_skill` 的 direct-mapping FK
- 为什么 MCP server 注册时必须带 `X-Kn-ID` header
- 为什么 agent `mode` 必须是 `"react"`（默认模式不挂载工具）
- MCP / Skill 清理的三态机协议

## Troubleshooting

如果在 chat trace 里看到 `builtin_skill_load returned 404`，这是无害的。
Decision Agent 实际是从 `find_skills` 返回的 MCP 元数据和 `SKILL.md` 描述里
挑出正确的 skill；显式的 `builtin_skill_load` 查询只是 best-effort，跑一遍
失败不影响整体流程。

## Cleanup

脚本退出时（成功 / 失败）自动清理所有资源：KN、MCP、Skills、Agent、Datasource、
mock backend 进程。
