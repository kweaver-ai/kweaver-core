# 05 · Skill Routing Loop — 业务知识网络驱动的 Skill 治理

> [English](./README.md)

> 3 个物料触发同样的库存告警，Decision Agent 给出 3 条不同处置路径——
> 每条都能在业务知识网络里找到依据。

## 故事

续作 03 那位采购工程师：她现在看到每张告警单上已经写好了处置方案。3 个物料、
3 条不同路径，**没改一行 prompt**。BKN 里的 `applicable_skill` 关系是 Skill 路由的
**唯一真相源**——Agent 只能从 `find_skills` 返回的候选集里挑，没有别的杠杆。

## 这个 example 展示什么

5 个组件协同跑通一个可验收闭环：

| 组件 | 职责 |
|---|---|
| **execution-factory** | 注册 / 版本化 3 个 Skill 包 |
| **业务知识网络（BKN）** | 通过 `applicable_skill` 关系把 Skill 绑到物料 |
| **Vega** | 把 BKN ObjectType 映射到 MySQL 表（读多写少） |
| **context-loader (`find_skills`)** | 按物料实例召回适用的 Skill |
| **Decision Agent** | 读 KN 证据 → 选 Skill → 输出可校验决策 |
| **run.sh 验收器** | 断言 Agent 选路 → 调 mock 业务端点 → 检查调用日志 |

## 前置条件

- `kweaver` CLI ≥ 0.7.1
- 启用了 Decision Agent + execution-factory + Vega 的 KWeaver 平台
  （先 `kweaver auth login <平台地址> [--insecure]`）
- **平台能访问到**的 MySQL（不是你笔记本上的），且账号有 CREATE/INSERT/SELECT/UPDATE 权限
- `python3`（依赖 Flask + mysql-connector-python，
  `pip install -r tool_backend/requirements.txt`）
- 平台模型工厂里注册的 LLM 模型（用
  `kweaver call /api/mf-model-manager/v1/llm/list` 拿 model_id）

简单自检平台组件是否就绪：

```bash
kweaver auth whoami                                      # 是否登录
kweaver call /api/mf-model-manager/v1/llm/list | head    # 模型工厂可达？
kweaver call /api/agent-operator-integration/v1/mcp/     # execution-factory 可达？
```

## 快速开始

```bash
cd examples/05-skill-routing-loop
cp env.sample .env
vim .env                                    # 填 PLATFORM_HOST、LLM_ID、DB_*
pip install -r tool_backend/requirements.txt
./run.sh                                    # 端到端约 5 分钟
./run.sh --bonus                            # 跑 Bonus 段，并做可校验验收
```

> **并发注意：** 请不要同时运行两个 `./run.sh` 实例。脚本使用固定的 `KN_ID`
> （`ex05_skill_routing`）以及固定的 Skill 名（`standard_replenish` /
> `substitute_swap` / `supplier_expedite`）；第二个实例会在 Skill 注册阶段
> 直接撞上，并且任一实例的清理逻辑会把另一个实例的 KN 一起删掉。

## 你会看到什么

| 物料 | KN 证据 | DA 选 | 原因 |
|---|---|---|---|
| MAT-001 | 绑定 `substitute_swap`；SUB-001A/B 有库存 | substitute_swap | Python 算法打分挑替代料 → 调 MES |
| MAT-002 | 绑定 `supplier_expedite`；SUP-2 capability=expedite | supplier_expedite | 供应商能加急 → POST 供应商门户 |
| MAT-003 | 只绑定 `standard_replenish` | standard_replenish | 默认路径 → 走 ERP 下单 |

脚本会把每次 Agent 输出保存到 `.chat-<SKU>.log`，并检查输出里包含期望的
Skill 名。随后脚本会调用本机 mock 业务后端完成可观察动作，并检查
`.tool_backend.log` 里出现：

```text
[mes/swap]
[supplier/expedite]
[procurement]
```

看到 `✓ mock backend observed MES, supplier, and ERP calls` 说明三条业务动作
都已经打到 mock 后端。

如果你希望 `builtin_skill_execute_script` 在平台执行沙箱里也真正打到 mock 后端，
把 `.env` 里的 `TOOL_BACKEND_PUBLIC_URL` 设置成平台/沙箱可访问的地址
（例如一台内网可达机器的 `http://<host>:8765`）。默认
`http://127.0.0.1:8765` 只保证本机验收器可访问；平台沙箱里的
`127.0.0.1` 不是你的笔记本。

## Bonus — 改业务 → KN 重建 → AI 跟着变

`./run.sh --bonus` 会调 mock 业务系统的 admin 端点，把 MAT-002 的绑定 Skill
从 `supplier_expedite` 改成新注册的 `standard_replenish` Skill ID（直接 UPDATE
`materials.bound_skill_id`，由 `applicable_skill` 的 direct-mapping FK 决定边），
然后触发一次 `kweaver bkn build` 刷新底层 Vega 资源快照，再让 Agent 重新处理
MAT-002。Decision Agent 下一次 `find_skills` 拿到的就是新候选集，
自动切到 `standard_replenish`——**没改 prompt、没重新部署任何服务**。

> **为什么需要重建——以及为什么这不是平台限制：** 这个 example 用的是 Vega 的
> **batch 模式** dataview，图查询读的是 build 时拍下的资源快照。像
> `applicable_skill` 这样的 direct-mapping 关系在每次查询时实时计算——但底下
> 的数据是快照，MySQL UPDATE 要到下一次 build 才会反映出来。Vega 也支持
> **streaming 模式** 资源（基于 Debezium CDC + Kafka），业务变更秒级生效、
> 无需手工 rebuild——那才是生产路径。这里用 batch 是为了让 demo 只靠一个
> MySQL 跑通，不引 Kafka / Debezium / 额外基础设施依赖。

## 原理细节

完整设计文档：[`docs/superpowers/specs/2026-04-27-skill-routing-loop-example-design.md`](../../docs/superpowers/specs/2026-04-27-skill-routing-loop-example-design.md)

包括：
- BKN schema 和 `applicable_skill` 的 direct-mapping FK
- 为什么 MCP server 注册时必须带 `X-Kn-ID` header
- 为什么 agent `mode` 必须是 `"react"`（默认模式不挂载工具）
- 为什么脚本先注册 Skill，再用真实 Skill ID 渲染 CSV 和 agent config
- MCP / Skill 清理的三态机协议

## Troubleshooting

如果在 chat trace 里看到 `builtin_skill_load returned 404`，说明 BKN 里的
`skills.skill_id` 或 agent config 里的 `skills[].skill_id` 没有对齐到
execution-factory 注册返回的真实 Skill ID。当前脚本会先注册 Skill，再用真实 ID
渲染 CSV 和 agent config；正常情况下不应再出现这个错误。

## Cleanup

脚本退出时（成功 / 失败）自动清理所有资源：KN、MCP、Skills、Agent、Datasource、
mock backend 进程。
