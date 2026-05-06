---
title: Example 04 — Skill Routing Loop Design
date: 2026-04-27
status: draft
---

# Example 04: Skill Routing Loop · KN-driven Skill 召回 + Decision Agent 端到端治理

## 一句话定位

展示 **KWeaver 端到端 Skill 治理闭环**——业务真相变化（业务系统 → Vega → BKN）自动驱动 Decision Agent 的决策行为变化，全程**无 prompt 修改、无 agent 重新部署**。

## 核心价值主张

行业现状：Agent Skill 普遍"裸奔"——无版本、无作用域、无召回控制、无执行审计。

KWeaver 把 Skill 治理拆成**五件事**，分别由五个组件分工：

| 治理维度 | 组件 |
|---|---|
| 注册 / 版本 / 执行 sandbox | execution-factory |
| 绑定（Skill 与业务对象关系） | 业务知识网络（BKN） |
| 业务真相投影（数据来源） | Vega（logic view） |
| 召回（按上下文 scope 收窄） | context-loader（`find_skills`） |
| 编排 / 选择 / 执行 / 审计 | Decision Agent |

本 example 用一个最小完整闭环把五件事跑通。

## Story 叙事线（example 主线）

> 续作 03 的采购工程师故事：03 里知识网络帮她在早 8:00 拿到了告警。本篇里，**告警的处置方案也由系统自动选**——3 个不同物料触发同样的 critical 告警，Decision Agent 借助业务知识网络召回各自适用的 Skill，给出 3 条不同处置路径，全程可解释、可追溯。
>
> Bonus：在业务库里改一行供应商 capability，下一次同物料告警的处置路径自动换轨——展示"业务变 → AI 跟着变"的最小闭环。

---

## 架构概览

```
                ┌──── 业务系统 / 业务库 (MySQL) ────┐
                │  suppliers / materials / mat_skill │
                └──────────────┬─────────────────────┘
                               │  Vega logic view (映射)
                               ▼
                ┌──── BKN: spike_skill_routing ────┐
                │   ObjectTypes:                   │
                │     • material  (sku PK)         │
                │     • supplier (supplier_id PK)  │
                │     • skills    (skill_id PK)    │
                │   RelationTypes:                 │
                │     • applicable_skill (M→S)     │
                │     • supplied_by (M→Supplier)   │
                └──────────────┬─────────────────────┘
                               │  context-loader: find_skills
                               ▼
                ┌──── Decision Agent ──────────────┐
                │   • find_skills MCP 召回         │
                │   • builtin_skill_load 读 SKILL.md │
                │   • LLM 选 → execute_script       │
                │   • Action: 调 mock 业务系统      │
                │   • Audit log                    │
                └──────────────┬─────────────────────┘
                               │  Skill 包来源
                ┌──────────────▼────────────────────┐
                │   execution-factory              │
                │     • standard_replenish (无 py) │
                │     • substitute_swap (★ Python) │
                │     • supplier_expedite (无 py)  │
                └──────────────────────────────────┘
```

**说明**：
- BKN **不存数据**，通过 Vega logic view 映射到业务库表（`suppliers` / `materials` / `mat_skill`）
- find_skills MCP 接入 Decision Agent 的方式：作为 `skills.mcps[]` 配置项，引用 `mcp_server_id`（指向 context-loader 的 MCP endpoint）
- Skill 包由 execution-factory 注册/版本化/执行，DA 通过 `builtin_skill_load` / `builtin_skill_execute_script` 调用

---

## 文件结构

```
examples/04-skill-routing-loop/
├── README.md / README.zh.md
├── env.sample
├── run.sh                                  # 端到端脚本（与 03 风格一致）
├── data/
│   ├── materials.csv                       # 含 bound_skill_id FK 字段
│   ├── suppliers.csv
│   └── skills.csv                          # 3 个 Skill 实例（绑定 KN）
├── skills/                                 # 本地 Skill 包源
│   ├── standard_replenish/
│   │   └── SKILL.md
│   ├── substitute_swap/
│   │   ├── SKILL.md
│   │   └── pick_substitute.py             # ★ 多准则打分 Python
│   └── supplier_expedite/
│       └── SKILL.md
├── tool_backend/
│   └── server.py                           # mock 业务系统 (沿用 03 风格)
├── bkn/                                    # bkn push 输入目录
│   ├── network.bkn
│   ├── object_types/
│   │   ├── material.bkn
│   │   └── skills.bkn
│   └── relation_types/
│       └── applicable_skill.bkn
└── agent.json                              # Decision Agent 配置（含 find_skills MCP）
```

---

## BKN Schema（最小化版）

### ObjectTypes（3 个）

| ObjectType | id（必须字面量） | Properties | 实例 |
|---|---|---|---|
| Material | `material` | `sku` (PK) · `name` · `current_stock` · `safety_stock` · `material_risk` · `bound_skill_id` (FK→skills) · `supplier_id` (FK→supplier) | 5 条（3 critical + 2 substitute） |
| Supplier | `supplier` | `supplier_id` (PK) · `name` · `capability` ∈ {`normal`, `expedite`} | 3 条 |
| Skills | `skills`（**hard-coded**，平台 `SkillsObjectTypeID` 配置） | `skill_id` (PK) · `name` · `description` | 3 条 |

### RelationTypes（2 个）

| 关系 | 端点 | Type | Mapping | 用途 |
|---|---|---|---|---|
| `applicable_skill` | Material → Skills | `direct` | `bound_skill_id` → `skill_id` | find_skills 召回核心边 |
| `supplied_by` | Material → Supplier | `direct` | `supplier_id` → `supplier_id` | DA 读 supplier capability 作为推理证据 |

> **DA 推理 vs find_skills 召回的分工**：find_skills 只看 `applicable_skill` 边返回候选 Skill 集合（MAT-002 召回到 [supplier_expedite, standard_replenish] 两个候选）；**DA 拿到候选后通过 query KN 读 MAT-002 的 supplier.capability，决定走哪一个**。这条分工让 Bonus 段的"改业务变 AI 跟着变"成立——capability 改了，DA 推理结果跟着变，无需重新绑定。

> **关键约束（来自 spike P1-1）**：Skills ObjectType 的 id 必须字面量为 `"skills"`，否则 `find_skills` 报 `[BknBackendAccess] /object-types/skills not found`。
> **关键约束（来自 spike P1-2）**：RelationType.Endpoint.Type 必须用 `direct`（FK 直连），写 `data_view` 会触发 platform 502 nginx。

### 数据示例（CSV）

**materials.csv**：
```
sku,name,current_stock,safety_stock,material_risk,supplier_id,bound_skill_id
MAT-001,Battery Cell,40,100,critical,SUP-1,substitute_swap
MAT-002,Power Module,30,120,critical,SUP-2,supplier_expedite
MAT-003,Connector,15,80,critical,SUP-3,standard_replenish
SUB-001A,Battery Cell Substitute,200,50,normal,SUP-1,
SUB-001B,Battery Cell Alt,80,40,normal,SUP-1,
```

> 5 条物料中，3 条是 primary（MAT-*），各绑一种 Skill + 一个供应商；2 条是 substitute（SUB-*），不绑 Skill。
>
> MAT-002 binds to `supplier_expedite`，但 SUP-2 的 `capability` 决定 DA 实际是否能走加急——这是 Bonus 段演示"改 capability 让 DA 换轨"的触点。

**skills.csv**（实例）：
```
skill_id,name,description
standard_replenish,标准补货,Default procurement order via ERP
substitute_swap,替代料切换,Pick best substitute via Python scoring then call MES
supplier_expedite,供应商加急,Send expedite request to supplier portal
```

**suppliers.csv**：
```
supplier_id,name,capability
SUP-1,Acme Corp,normal
SUP-2,Bolt Industries,expedite
SUP-3,Cell Source,normal
```

> Bonus 段会把 SUP-2 的 capability 改为 `normal`，演示 MAT-002 处置路径自动从 `supplier_expedite` 切到 `standard_replenish`（找不到加急路径，fall back 到默认）。

---

## Skill 包设计

### Skill 1: `standard_replenish`（无 Python）

**SKILL.md**：
```markdown
---
name: standard_replenish
description: 标准补货流程——调 ERP 下采购订单
version: 0.1.0
---

# 标准补货 Skill

当物料触发 critical 库存告警，且无替代料 / 供应商不可加急时使用。

## 调用契约

输入：material_sku
动作：HTTP POST → mock_backend `/procurement/order` { sku, qty }
输出：PO 编号
```

### Skill 2: `substitute_swap`（★ Python 算法）

**SKILL.md**：
```markdown
---
name: substitute_swap
description: 替代料切换——多准则打分选最优替代料并通知 MES
version: 0.1.0
---

# 替代料切换 Skill

当物料有可用替代料时，跑 pick_substitute.py 多准则打分（库存/兼容度/成本/lead time）选最优。

## 调用契约

输入：material_sku
脚本：pick_substitute.py material_sku=<sku>
脚本动作：
  1. 从环境变量读 KN_ID
  2. 查 KN 找该物料的所有替代料候选
  3. 多准则打分（4 维 weighted）
  4. 调 mock_backend `/mes/swap` { from_sku, to_sku, qty }
  5. 输出 { chosen_sku, scores: [...] }
```

**pick_substitute.py**（核心算法骨架）：
```python
import sys
import os
import json
import urllib.request

KN_ID = os.environ["KN_ID"]
PLATFORM_URL = os.environ["KWEAVER_URL"]
TOOL_BACKEND = os.environ["TOOL_BACKEND_URL"]

def get_substitutes(material_sku):
    # 调 platform 的 query-instance-subgraph 找 has_substitute 关系
    ...

def score(sub):
    # weighted: stock(0.4) + compat(0.3) + cost_delta(0.2) + lead_time(0.1)
    ...

def main():
    sku = parse_arg("material_sku")
    candidates = get_substitutes(sku)
    scored = [(c, score(c)) for c in candidates]
    chosen = max(scored, key=lambda x: x[1])
    # 调 MES
    requests.post(f"{TOOL_BACKEND}/mes/swap", json={...})
    print(json.dumps({"chosen_sku": chosen[0]["sku"], "scores": scored}))
```

> **重要**：Python 脚本需要查 KN 找替代料——这一步通过 `query_instance_subgraph` MCP（context-loader 暴露）调用。spec 阶段先标 `[需 plan 阶段验证 platform 是否支持 Skill 内调 MCP]`。
> **简化备选**：如果 Skill 执行 sandbox 内调 MCP 路径不通，则 substitute candidates 直接从 stdin/参数传进来（DA 在调 Skill 之前先调 query），但这削弱"Python 自带智能"叙事。

### Skill 3: `supplier_expedite`（无 Python）

**SKILL.md**：
```markdown
---
name: supplier_expedite
description: 供应商加急——向供应商门户发出加急请求
version: 0.1.0
---

# 供应商加急 Skill

当物料供应商 capability 为 expedite 时使用。

## 调用契约

输入：material_sku
动作：HTTP POST → mock_backend `/supplier/expedite` { sku, qty, sla_hours: 36 }
```

### Skill 包打包

每个 Skill 目录打 zip（**SKILL.md 必须在 zip 根**，spike P1-B 验证）：

```bash
for s in skills/*/; do
  name=$(basename "$s")
  (cd "$s" && zip -qr "../${name}.zip" .)
done
```

---

## Agent 配置（agent.json）

**spike 已端到端验证**：以下是真实跑通的配置（mode=react + MCP X-Kn-ID 注入）。

`agent.json` 内容（**仅 config 部分**——`name` 和 `profile` 通过 CLI flag 传，**不能放进 JSON**）：

```json
{
  "input": { "fields": [{ "name": "user_input", "type": "string", "desc": "" }] },
  "output": { "default_format": "markdown" },
  "mode": "react",
  "system_prompt": "你是供应链处置助手。当收到物料告警时：调用 find_skills 工具，参数 object_type_id='material'，instance_identities=[{sku: <提到的物料SKU>}]，召回该物料适用的 Skill 候选集。再用 query_object_instance / query_instance_subgraph 查 supplier 当前 capability、material 当前 stock。基于这些证据选 1 个 Skill，调 builtin_skill_load 读它的 SKILL.md 确认契约，必要时 builtin_skill_execute_script 跑 Python。输出处置结果 + 决策依据。",
  "data_source": { "kn_id": "<kn_id>" },
  "skills": {
    "tools": [],
    "agents": [],
    "mcps": [{ "mcp_server_id": "<mcp-id>" }],
    "skills": [
      { "skill_id": "standard_replenish" },
      { "skill_id": "substitute_swap" },
      { "skill_id": "supplier_expedite" }
    ]
  },
  "llms": [{
    "is_default": true,
    "llm_config": {
      "id": "<llm-id-from-env>",
      "name": "<llm-name>",
      "max_tokens": 4096,
      "temperature": 0.7,
      "top_p": 1,
      "top_k": 1
    }
  }],
  "react_config": {
    "disable_history_in_a_conversation": false,
    "disable_llm_cache": false
  },
  "memory": { "is_enabled": false },
  "related_question": { "is_enabled": false },
  "plan_mode": { "is_enabled": false }
}
```

**创建命令**（注意 name 只能字母数字下划线、不能数字开头、不能含 dash）：
```bash
kweaver agent create \
  --name "ex04_skill_routing" \
  --profile "KN-driven skill routing demo" \
  --config /tmp/agent.json
# → {"id":"<agent-id>","version":"v0"}

kweaver agent publish <agent-id>
```

**关键约束（spike 实证）**：
1. **`mode` 必须设为 `react` 或 `dolphin`**——默认值（不设）走的是 v1 路径，**完全不加载 MCP 工具**。
2. **`skills.skills[].skill_id` 必须等于 execution-factory 注册的 skill 包 id**——否则 LLM 调 builtin_skill_load 时找不到。**含义**：example 要么先注册 skill 包拿 UUID 再写进 skills.csv，要么用人类可读 ID（如 `standard_replenish`）作为 skill 包注册时的指定 ID。后者更优雅，spec 默认走这条。
3. **MCP server 的 headers 必须含 `X-Kn-ID`**——否则 find_skills 报 kn_id required。
4. agent name 约束：`[a-zA-Z_][a-zA-Z0-9_]*`，不允许 dash。

---

## MCP Server 注册（context-loader）

spike P1 关键发现：62 平台 MCP 注册表初始为空，example 必须自己注册一次。

**关键：必须把 `X-Kn-ID` 写进 MCP 的 headers**——否则 LLM 调 find_skills 时 context-loader 会报"kn_id is required"。**这是因为当前 `mcp_tool.py` 不会把 agent 的 `data_source.kn_id` 自动传到 MCP 调用里**（见 SDK/平台 issue P1-E）。

```bash
kweaver call /api/agent-operator-integration/v1/mcp/ -X POST \
  -H "Content-Type: application/json" \
  -H "x-business-domain: bd_public" \
  -d "{
    \"mode\": \"stream\",
    \"url\": \"https://${PLATFORM_HOST}/api/agent-retrieval/v1/mcp\",
    \"name\": \"ex04_context_loader\",
    \"description\": \"Context-loader MCP for find_skills\",
    \"creation_type\": \"custom\",
    \"headers\": { \"X-Kn-ID\": \"${KN_ID}\" }
  }"
# → { "mcp_id": "<uuid>", "status": "unpublish" }

kweaver call /api/agent-operator-integration/v1/mcp/<mcp_id>/status -X POST \
  -H "x-business-domain: bd_public" \
  -d '{"status":"published"}'
```

---

## run.sh 流程设计

```
Step 0   检查 .env 必填变量（DB_*, PLATFORM_HOST, LLM_ID）
Step 1   ds connect mysql → 拿 ds_id
Step 2   ds import-csv → 上传 3 张 CSV 表
Step 3   bkn create-from-csv（短暂建一次 KN，仅为获取 dataview ID）→ 拿 dataview ID → 删 KN（保留 dataview）
Step 4   生成 bkn/object_types/*.bkn 模板填入 dataview ID
Step 5   bkn push bkn/ → 拿 kn_id (= "spike_skill_routing")
Step 6   bkn build --wait（同步构建）
Step 7   起 mock 业务 backend（python3 tool_backend/server.py &）
Step 8   for s in skills/*/; do zip + skill register → published
Step 9   注册 context-loader MCP server → published
Step 10  jq 填充 agent.json 模板（mcp_id + skill_ids + llm_id + kn_id）
Step 11  agent create --config → 拿 agent_id
Step 12  for sku in MAT-001 MAT-002 MAT-003; do
           agent chat <id> -m "Material $sku hit critical, decide and act"
         done
Step 13  agent history + audit log → 打印 trace 给读者
Step 14  Bonus（可选 --bonus 参数触发）：
         mysql UPDATE suppliers SET capability='normal' WHERE supplier_id='SUP-2'
         agent chat <id> -m "Material MAT-002 hit critical"  # 验证换轨
Step 15  cleanup（trap EXIT）
```

### Cleanup 协议（按 spike 学到的三态机）

```bash
cleanup() {
    # 顺序很关键，平台依赖关系
    [ -n "$AGENT_ID" ] && kweaver agent delete "$AGENT_ID" -y
    [ -n "$MCP_ID" ] && {
        kweaver call /api/.../mcp/$MCP_ID/status -X POST -d '{"status":"offline"}'
        kweaver call /api/.../mcp/$MCP_ID -X DELETE
    }
    for sid in $SKILL_IDS; do
        kweaver skill status "$sid" offline
        echo y | kweaver skill delete "$sid"
    done
    [ -n "$KN_ID" ] && kweaver bkn delete "$KN_ID" -y
    # 注：dataview 不会随 KN 删除（spike P1-6），example 通过 ds 删除间接清理
    [ -n "$DS_ID" ] && kweaver ds delete "$DS_ID" -y
    kill %1 2>/dev/null  # mock backend
}
trap cleanup EXIT
```

---

## 已验证事实清单（spike 阶段，含 spike B 端到端验证）

| # | 事实 | 验证命令 |
|---|---|---|
| 1 | `bkn push <dir>` 接受 schema-only，无需预存 dataview | `kweaver bkn push /tmp/spike-bkn` |
| 2 | Skills ObjectType id 必须字面量 `"skills"` | find_skills 调用证伪反推 |
| 3 | `RelationType.Type=direct` + FK 字段映射可被 find_skills 识别 | spike v4 完整跑通 |
| 4 | `find_skills` instance-scope 精准召回单个 skill | MAT-001 → spike_skill_swap (1 entry) |
| 5 | `find_skills` object_type-scope 召回全部 bound skills | object_type=material → 2 entries |
| 6 | Skill 包 zip 必须 SKILL.md 在根 | `zip -qr ../foo.zip .`（cd 进目录后） |
| 7 | Skill 三态机：unpublish → published → offline → DELETE | spike 完整跑通 |
| 8 | MCP server 注册 body schema | `{mode, url, name, creation_type, headers}` 已验证 |
| 9 | MCP proxy 能列 6 个 context-loader 工具（含 find_skills） | spike 已验 |
| 10 | Agent skills config schema | Go 源码 `skill.go` 已读出，spec 给完整模板 |
| 11 | **agent create + publish + chat 端到端通**（spike B）| `kweaver agent create/publish/chat` |
| 12 | **mode=react 触发 run_dolphin 路径**，MCP 工具被注入到 LLM | LLM 实际调用 find_skills |
| 13 | **MCP server 的 X-Kn-ID header 决定 find_skills 知道查哪个 KN** | spike B 反复证伪 |
| 14 | **LLM 真的会基于 KN 召回结果做证据驱动决策** | LLM 输出引用了 SUP-2 的 capability、bound_skill_id 等 |
| 15 | LLM 召回到 skill_id 后调 builtin_skill_load——若 skill 未注册则报 not found | spike B 复现 |

## 待 plan 阶段补 spike 的 unknowns（已收窄）

A、B 已在 spike B 中关闭。剩余：

| # | 问题 | 影响 | 缓解 |
|---|---|---|---|
| C | Skill 内 Python 脚本能否调 MCP 反查 KN（pick_substitute.py 找替代料） | 中 | 备选：DA 在调 Skill 前先 query，把候选传进来 |
| D | Vega logic view 是否真在 BKN 读路径上"实时"（业务库改了下一秒生效） | 低 | plan 阶段验一次；如果有缓存延迟，README 注明 |
| E | LLM 凭证（llm-id）从哪拿？现 .env 中提供 | ✅ | 已确认目标平台上有 deepseek-v3.2，model_id 由开发者填到本地 `.env` 中 |
| F | `kweaver skill register --content-file` vs `--zip-file` | ✅ | 只用 --zip-file，run.sh 自动打包 |
| G | skills.csv 里 skill_id 怎么和 execution-factory 注册的 skill 包对齐 | 中 | run.sh 先注册 skill 包（拿到平台 UUID），再用 UUID 写 skills.csv 后再 import |

---

## Bonus 段落设计

README 末尾独立段落，约 250 字 + 1 段 shell。

```bash
# Bonus: 验证"业务变 → AI 跟着变"
# 模拟李工在业务系统里改 SUP-2 的 capability
mysql -h $DB_HOST -u $DB_USER -p$DB_PASS $DB_NAME \
  -e "UPDATE suppliers SET capability='normal' WHERE supplier_id='SUP-2'"

# 重新触发 MAT-002 告警
kweaver agent chat $AGENT_ID -m "Material MAT-002 hit critical, decide and act"
# 预期变化：
# • find_skills 召回结果不变（still [supplier_expedite, standard_replenish]）
# • 但 DA 推理时调 query_instance_subgraph 读到 SUP-2.capability='normal'
# • DA 决定不走 supplier_expedite，落到 standard_replenish
# • BKN 通过 Vega 实时读业务库——无缓存延迟
```

> **架构含义**：Skill 候选由 BKN 关系决定（设计时确定，相对稳定），决策由 DA 推理 + KN 状态决定（运行时确定，跟着业务变）。两层分工让"业务变"和"AI 行为变"既解耦又对齐。
>
> **不演反向闭环**：Decision Agent action 写回业务库 → 下次召回看到新状态走不同路径——那是后续 example（feedback loop）的主题。

---

# Part B · 配套公众号文章叙事大纲

## 文章定位

- **标题（推荐）**：《业务真相即 AI 行为——Skill 治理的端到端实录》
- **备选**：
  - 《当一次 AI 决策能追溯到一次 ERP 字段修改》
  - 《让 AI 跟着业务跑——而不是反过来》
- **篇幅**：5500-6500 字
- **配图**：4 张（架构总图 / 五维治理表 / find_skills 召回反差表 / 5 跳 lineage 截图）
- **目标受众**：技术决策者（CTO / 技术总监）+ AI infra 实操者
- **节奏曲线**：§0 高 → §4 平稳推进 → §5 再起 → §7 升华全文最高 → §8 收尾

## 9 节结构

### § 0 · 导语（240 字）

场景钩子：4/16 17:42 李工在供应商管理系统改 SUP-3 capability；4/17 8:03 MAT-003 处置单换轨。承接句："要让'理所当然'成立，底层要做对至少五件事。"

### § 1 · 那 30 秒走过哪几跳（380 字）

拆 4 跳链路：业务库 → Vega → BKN → find_skills → DA。点睛：业务和 AI 之间没有"翻译环节"。

### § 2 · 换别的框架李工得通知谁加班（600 字）

行业对比表（4 行：OpenAI function calling / 通用 MCP / LangChain tools / 内嵌业务规则的 RAG agent —— **删 RAG 那行，保留 4 行**）。克制评论。

### § 3 · Skill 治理的五件事（550 字）

主视觉：五维表。KWeaver 五组件分工。

### § 4 · 端到端 demo · 五组件接力（1700 字）

倒序展开（从 8:03 那一刻开始往前回溯）：
- § 4.1 8:03 那一刻 DA 在做什么（350 字）
- § 4.2 顺到 context-loader（400 字）
- § 4.3 再顺到 BKN（400 字）
- § 4.4 再顺到 Vega（250 字 · 带过）
- § 4.5 最后到 execution-factory（300 字）

### § 5 · 回到 4/17 那张处置单 + 反向闭环带过（550 字）

主体 400 字：5 跳 lineage 完整展示。
末尾 150 字：承认完整闭环还有反方向（AI 行动 → 业务跟着变），不展开，留后续例子。

### § 6 · 治理的终极意义（400 字）

升华：controllability 是底线，evolvability 是终极价值。AI 团队从业务变更循环里被解耦。

### § 7 · 收尾 / CTA（200 字）

example 链接 + 一句行动召唤。

## Demo 与文章的对照

| 文章章节 | 对应 example 文件 / 步骤 |
|---|---|
| § 0-1 导语 | run.sh 整体 + Bonus |
| § 4.1 DA | agent.json + Step 11-12 |
| § 4.2 context-loader | MCP 注册 + find_skills 调用 |
| § 4.3 BKN | bkn/ 目录 + Step 5-6 |
| § 4.4 Vega | dataview ID 获取（Step 3-4） |
| § 4.5 execution-factory | skills/ 目录 + Step 8 |
| § 5 lineage | agent history + audit log |
| Bonus | run.sh --bonus 参数 |

---

# Part C · Spike 已知 issue 引用

平台 / SDK 严重级问题已由用户单独 thread 跟踪。example 设计中**显式吸收**的约束：

| Issue | 设计如何吸收 |
|---|---|
| 平台 P1-1（OT id 字面量） | bkn push 路径，frontmatter 控制 id="skills" |
| 平台 P1-2（502 on data_view） | RelationType 用 direct，不用 data_view |
| 平台 P1-3（mcp/skill 状态机） | cleanup 函数三步走 |
| **平台/SDK P1-E（新发现）**：`mcp_tool.py` 不把 agent.data_source.kn_id 自动注入到 MCP 调用 header | 在注册 MCP server 时把 `X-Kn-ID` 写进 headers（绑死 KN）。**这是工程缺陷**：单个 MCP server 只能服务一个 KN，需要每个 KN 各注册一次。建议平台修：在调 MCP proxy 时自动从 agent context 取 kn_id 注入 |
| **平台/SDK P1-F（新发现）**：agent create 不设 `mode` 时进入 v1 路径，MCP 工具完全不加载，且 chat 不报错只是 LLM 没工具 | `mode: "react"` 强制必填；README 警告 |
| **SDK P1-G（新发现）**：agent name 只接受 `[a-zA-Z_][a-zA-Z0-9_]*`，CLI 不预校验 | run.sh 命名规则约束 |
| SDK P1-A（CLI 缺 find-skills 子命令） | 用 `kweaver call` 调 raw HTTP；README 注明 |
| SDK P1-B（--content-file 不可用） | 只用 --zip-file；run.sh 自动打包 |
| SDK P1-C（create-from-csv 无回滚） | run.sh 用 `set -euo pipefail` + 显式 cleanup 收尾 |
| SDK P1-D（ds connect 不去重） | 用时间戳后缀避免冲突 |

---

## Open Questions（需要 Plan 阶段决策）

1. agent.json 中的 `system_prompt` 用中文还是英文？影响 LLM 召回行为稳定性。
2. mock 业务 backend 暴露端口冲突如何避免？（03 用了固定端口，可能冲突）
3. Bonus 段是否纳入主线 trace 打印？还是仅 README 文字说明 + 用户手动跑？
4. 文章发布前必须把 § 4.1 的 trace 截图拍出来——需要在 plan 阶段安排"补 e2e 跑通"任务。
