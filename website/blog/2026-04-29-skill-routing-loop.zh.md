---
slug: skill-routing-loop
title: "3 张告警，3 条路径，0 行 Prompt 修改：KWeaver 是怎么让 Agent 跟得上业务变化的"
authors: [kweaver]
tags: [kweaver, bkn, agent, architecture]
---


## 一、一个再普通不过的周一早上

9:15。供应链调度员小李刚泡好咖啡，工位上的告警面板就弹出 3 张 critical 单子：

| 物料 | 当前库存 / 安全水位 |
|---|---|
| MAT-001 电池芯 | 40 / 100 |
| MAT-002 电源模块 | 30 / 120 |
| MAT-003 连接器 | 15 / 80 |

她扫一眼就知道这三张单子的处置方式从来都不一样：

- MAT-001 在仓库里有合格替代料 SUB-001A、SUB-001B——走「替代料切换」，让 MES 切生产；
- MAT-002 的供应商 SUP-2 有「加急」能力——走「催供应商加急」，调供应商门户的加急 API；
- MAT-003 没有替代、供应商也加不了急——只能走「标准补货」，从 ERP 下采购单。

如果让 AI 帮她处置，AI 怎么知道每张单子"对的姿势"？更尖锐一点：**当业务侧明天决定 MAT-002 改走标准补货，AI 怎么自动跟上？**

<!-- truncate -->

## 二、行业最常见的做法，以及它的天花板

绝大多数 agent 项目今天的答案是：**把规则写进 prompt**。

```
你是供应链助手。
当物料有替代料且充足时调 substitute_swap，
当供应商支持加急时调 supplier_expedite，
否则调 standard_replenish……
```

这个写法在 demo 里跑得动。问题在生产线上才显现：

- 一年后 prompt 长到 5000 字，没人完整读过；
- 加一个 Skill 要改 3 处条件分支、跑回归、上线一次；
- 业务侧想把 MAT-002 改去标准补货，要找 IT、找算法、改 prompt、回归、发版——一个本该是 SKU 维护的动作，变成了一次小型 release。

更根本的问题：**Skill 路由本来是业务治理问题，被错误地塞进了 prompt 工程**。

## 三、KWeaver 的答案：让产品自己消化这件事

看 KWeaver 写出来的方案是怎样的。完整 demo 在 `examples/05-skill-routing-loop/`，一条命令端到端跑通：

```bash
cp env.sample .env  # 填平台地址 / LLM ID / DB
./run.sh --bonus    # 5 分钟跑完
```

跑完之后小李会看到：3 张告警，AI 给出 3 条不同处置路径；脚本还会断言每条路径，并检查 mock 业务后端确实收到了 MES、供应商门户、ERP 三类调用。**整个工程目录里没有一行 if/else 描述「该选哪个 Skill」**。所有路由判断都被外包给了 KWeaver 的三件设施联动：

- **Skill** 是企业能力的一阶资产：3 个 Skill 包（标准补货 / 替代料切换 / 供应商加急）先注册到 execution-factory，拿到真实 Skill ID，有版本、有 SKILL.md 契约、能 Python 脚本也能 HTTP API；
- **BKN（业务知识网络）** 把 MySQL 业务表映射成图，并通过一条 `applicable_skill` 关系把 Skill 绑定到物料——`materials.bound_skill_id` 改成新的真实 Skill ID，这条边就跟着变；
- **Agent** 用 `find_skills` 查 BKN 拿到候选集，再用 `builtin_skill_load` / `builtin_skill_execute_script` 落到 Skill 执行——它**不知道也不需要知道**业务上"哪个物料该用哪个 Skill"。

而 system_prompt 里**没有任何业务路由规则**：

> 你是供应链处置助手。`find_skills` 返回的就是该物料当前适用的 Skill 集合，你只能从这个集合中选择并执行……
> 1. 调 find_skills 拿候选；
> 2. 查 BKN 取证据；
> 3. 选一个 Skill 加载契约；
> 4. 执行；
> 5. 输出决策依据。

整段 prompt 里有一个值得反复读的事实：它从头到尾**没有任何「业务条件 → Skill」的分支**。没有"使用 substitute_swap 当 X"，没有"遇到 Y 调 supplier_expedite"。这条 prompt 是 **route-agnostic 的**——

- 加一个新 Skill：注册到 execution-factory，绑给某些物料——**不动 prompt**；
- 退役一个 Skill：解绑、下线——**不动 prompt**；
- 把 MAT-002 从「催供应商」改成「标准补货」：改一行 SQL——**不动 prompt**。

prompt 里那句"`find_skills` 返回的就是真相源"不是"针对供应链场景调出来的话术"，而是一条**领域无关的治理不变式**：它定义了 Agent 和 KN 之间的权责契约——**Agent 负责怎么思考，KN 负责能做什么**。这条契约一次写好，整个供应链处置场景所有的 Skill 增删改、所有的物料绑定调整，都在 prompt 之外完成。

这才是把 Skill 路由从「prompt 工程问题」搬到「数据治理问题」之后，prompt 真正该长的样子。

## 四、Demo 实录：3 张告警，3 条路径

3 张告警丢给同一个 Decision Agent。每一条都走同样的动作序列：**拿候选 → 取证 → 选定 → 加载契约 → 执行**——只是因为 BKN 给的候选集不同，最后落到的 Skill 也就不同。

**MAT-001 → substitute_swap**

> 候选集只有 substitute_swap，agent 加载它的 SKILL.md 契约后，按 compat_score / cost_delta / lead_time 评分挑出 SUB-001A，POST 到 mock MES `/mes/swap`。

**MAT-002 → supplier_expedite**

> 候选集只有 supplier_expedite。Agent 在图里取证发现 SUP-2 capability=expedite，加载契约后按 `sku` / `supplier_id` / `sla_hours` 调供应商门户 `/supplier/expedite`。

**MAT-003 → standard_replenish**

> 候选集只有 standard_replenish——没替代、没加急。直接加载契约、按 sku/qty 调 ERP `/procurement/order`。

> 实现细节：拿候选用的是 context-loader 暴露的 MCP 工具 `find_skills`；加载契约和执行脚本用的是 KWeaver 的两个内置工具 `builtin_skill_load` / `builtin_skill_execute_script`。这三件事是 agent.json 里挂上去的标准能力，不是为这个 demo 写的。

值得反复看的是这种对称性：**Agent 的"行为"在三条路径上完全一样**（同一个 system_prompt、同一套工具调用顺序），变化的只是候选集里的那个 Skill ID。BKN 的边一改，加载的契约就变、执行的脚本就变、落到的业务后端也跟着变。Agent 自始至终不需要懂"哪个物料该怎么处置"——它只负责按契约执行 KN 给的那一个候选。

3 条不同路径，**没改一行 prompt**——差别全部来自 BKN 里的 `applicable_skill` 边。脚本最后会看到类似：

```text
✓ verified route=substitute_swap and mock action
✓ verified route=supplier_expedite and mock action
✓ verified route=standard_replenish and mock action
✓ mock backend observed MES, supplier, and ERP calls
```

## 五、Bonus：当业务规则变了，AI 怎么自动跟上？

真正的考验。业务侧决定：MAT-002 不再走"催供应商"，改走"标准补货"。

```bash
# 1. 业务系统改一行 SQL（这里写入的是 standard_replenish 注册后的真实 Skill ID）
UPDATE materials SET bound_skill_id='<standard_replenish_skill_id>' WHERE sku='MAT-002';

# 2. 刷新 KWeaver 这边的资源快照（这一步的工程含义见下一节）
kweaver bkn build ex05_skill_routing --wait

# 3. 重新触发告警
kweaver agent chat <agent-id> -m "Material MAT-002 hit critical stock level again..."
```

agent 这次的推理日志：

> 供应商 SUP-2 具有加急能力，但对应的 supplier_expedite 技能**不在候选集中**——按规则我必须从 find_skills 返回的候选里选。
> 选定技能：standard_replenish（标准补货）。

注意 agent 自己说出来的这句话：「我看到了 SUP-2 能加急，但治理结果不让我走那条路。」**有边界、可治理、能解释**——这才是企业级 Agent 应有的姿态。

prompt 没动、agent 没改、Skill 包没重新部署。整条变化只发生在 MySQL 一行 UPDATE 和一次资源快照刷新。

## 六、为什么这里要 build——以及为什么这不是平台限制

读到这里有个会让人皱眉的细节：Bonus 段里"业务改 SQL"和"AI 跟随"之间多了一步 `kweaver bkn build`。看上去像"业务一改就要重建图"——如果真是这样，前面讲的"治理资源"听上去就有点虚。

先把误会拆掉：**`kweaver bkn build` 不是在重建图**。

`applicable_skill` 这种 direct-mapping 关系，**本身就是查询时按底层资源的当前快照实时计算的**——图的拓扑、关系语义、Schema 都不需要也不会被"重建"。剩下的唯一问题是：**底层资源的"当前快照"什么时候被同步过来？**

这件事是 KWeaver 数据层（Vega）的职责，它给了两档同步模式：

| 模式 | 同步方式 | 业务变更到 AI 看到的延迟 | 基础设施要求 |
|---|---|---|---|
| **batch** | `kweaver bkn build` 时全量读一次快照 | 直到下一次 build | 数据库可达即可 |
| **streaming** | Debezium CDC → Kafka → 资源秒级订阅 | ~秒级 | Kafka + Kafka Connect + binlog ROW 模式 |

example-05 选了 batch——因为这是个 5 分钟跑通的 demo，不该让读者先搭一套 Kafka 才能体验 KN 驱动的 Skill 路由。**batch 的代价就是要人工触发一次 `kweaver bkn build` 来刷新快照**，所以 Bonus 段才会有那一行命令。它只是 batch 同步模式的"手动 tick"，跟图本身的可治理性没有任何关系。

把同样这个 example 跑在生产环境，把 dataview 切到 streaming：

- 业务系统执行 `UPDATE materials SET bound_skill_id=...`
- Debezium 捕获 binlog 变更 → Kafka → Vega 资源在秒级内更新
- 下一次 `find_skills` 拿到的就是新候选集
- **没有 build 步骤，没有人工干预**

反馈环时延从「人为触发的 build」变成「业务系统的事务延迟 + 几秒 CDC」——这才是 production 部署该有的样子。

所以这不是平台限制，而是 **demo 在"零基建复杂度"和"秒级反馈环"之间，自己选了前者**。换个说法：**KWeaver 给的是一档可调旋钮，不是固定档位**——POC 用 batch，一台 MySQL 就够；上生产切 streaming，秒级反馈环。**开发者写的代码（Skill 包、BKN 模板、agent.json）一行不用动**，切的只是 dataview 的同步模式。

## 七、端到端：从 Skill 到 BKN 到 context-loader 到 Agent

值得反复看的，不只是"prompt 里没写 Skill 名"这一个点，而是这条 demo 让你在 5 分钟之内看清 KWeaver 整条治理链路是怎么咬合的。从一次告警进来到 AI 给出处置，4 件设施按顺序在做这些事——每一件只回答一个问题，互相之间通过几个干净的协议啮合。

**1. Skill：企业能力的一阶资产清单。** 3 个 Skill 包（标准补货 / 替代料切换 / 供应商加急）注册到 execution-factory，每一个都有 SKILL.md 契约、独立版本号，能 Python 也能 HTTP API。这一层回答的是「系统能做什么」——不是写在 prompt 里的几句话，而是和数据库里的表、API 网关上的路由一样的可治理资源，能被增删改查、版本化、上下线。

**2. BKN：业务真相到 AI 可读图状态的桥。** `material` / `supplier` / `skills` 三个 ObjectType 把 MySQL 业务表映射成图；`applicable_skill` 这条 direct-mapping 关系把 Skill 绑定到具体物料——`materials.bound_skill_id` 改成新的 Skill ID，这条边就跟着变。这一层回答的是「哪个物料该用哪个能力」，把 Skill 路由从 prompt 工程问题彻底搬到了数据治理问题。业务系统的 SQL UPDATE 和 AI 看到的图状态，第一次有了清晰的因果链。

**3. context-loader：图与 LLM 之间的运行时翻译层。** Agent 调 `find_skills(object_type=material, instance=MAT-002)` 时，是 context-loader 在背后查 BKN、按 `applicable_skill` 关系召回候选、把结果以 MCP 工具结果的形式还给 Agent；同样地，`query_object_instance` / `query_instance_subgraph` 让 Agent 能继续在图里取证（SUP-2 的 capability、其他替代料的库存）。这一层回答的是「图里的事实怎么交到 LLM 手里」——它把图状态翻译成 Agent 能调用的语义工具，让 Agent 不需要懂 SQL、不需要懂 Schema，就能拿到当下这一刻该物料的真相。

**4. Agent：有边界的编排者。** Decision Agent 的 system_prompt 里**不出现任何业务路由分支**，只规定流程：拿候选 → 查证据 → 选 Skill → 加载契约 → 执行 → 给依据。它不知道也不需要知道业务里"哪个物料该用哪个 Skill"——那是 BKN 的事；它也不直接读 MySQL——那是 Vega + context-loader 的事。它只在 context-loader 给的候选集里挑，再用 `builtin_skill_load` / `builtin_skill_execute_script` 把 Skill 落地。这一层回答的是「该怎么思考」，并且**只回答这一个问题**。

四件设施咬合的接口也很克制：
- 业务系统 ↔ BKN：Vega 的 direct-mapping（batch / streaming 两种同步模式）
- BKN ↔ context-loader：dataview + 关系定义
- context-loader ↔ Agent：MCP server + `X-Kn-ID` header
- Skill ↔ Agent：`builtin_skill_load` / `builtin_skill_execute_script` 内置工具

**用户在这个 example 里只需要写：3 个 Skill 包 + 5 张 BKN 模板 + 一份 `agent.json` + 一份业务数据 CSV。** `run.sh` 会先注册 Skill，再把真实 Skill ID 渲染回 CSV 和 agent config，避免 `builtin_skill_load` 找不到包。剩下的复杂性——Skill 的注册 / 版本化 / 上下线、业务表到图的双向桥、Agent 与 KN / Skill 之间的协议、mode=react 的工具挂载、Decision 链的可观察、清理回滚——全部由产品来扛。

这才是好用的 AI 平台该有的样子：**让简单的事简单，让复杂的事至少可控。**

## 八、一句话主张

Agent 的能力边界不该写在 prompt 里。它应该被治理成一阶资源，和数据库里的字段、流程图里的节点、图谱里的关系一样可以被增删改查、版本化、审计、回滚。

KWeaver 的回答是这条四件套的链路：让 **Skill 成为可治理的能力资产**，让 **BKN 成为业务真相到 AI 可读图状态的桥**，让 **context-loader 成为图与 LLM 之间的运行时翻译层**，让 **Agent 成为有边界的编排者**——四者通过几个干净的协议啮合，开发者只需要写自己业务里那点真正独特的逻辑。

剩下的复杂性，产品来扛。

---

**自己跑一遍**

```bash
git clone https://github.com/kweaver-ai/kweaver-core
cd kweaver-core/examples/05-skill-routing-loop
cp env.sample .env  # 填 PLATFORM_HOST、LLM_ID、DB_*
./run.sh --bonus    # 端到端约 5 分钟
```

如果希望平台执行沙箱里的 `builtin_skill_execute_script` 也直接打到 mock 后端，请把 `.env` 里的 `TOOL_BACKEND_PUBLIC_URL` 设置成平台/沙箱可访问的地址。默认 `http://127.0.0.1:8765` 只保证本机验收器可访问；平台沙箱里的 `127.0.0.1` 不是你的笔记本。

完整代码、BKN 模板、Skill 包、agent.json，都在 `examples/05-skill-routing-loop/`。
