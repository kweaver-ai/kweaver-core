# 快速开始

以下步骤假设 KWeaver Core 已按 [部署](installation/deploy.md) 文档完成安装及文中的安装后检查。

---

## 准备工作

无论选择哪种操作方式，以下准备工作需要在终端中执行一次。

### 安装 CLI

```bash
npm install -g @kweaver-ai/kweaver-sdk
```

需要 Node.js 22+。也可用 `npx kweaver --help` 免安装试用。

### 登录平台

```bash
kweaver auth login <平台地址> -k
```

- `<平台地址>` 是部署完成后 `deploy.sh` 输出的访问地址。
- `-k` 用于自签名证书；正式证书可省略。

登录成功后确认当前配置：

```bash
kweaver config show
```

> 其它登录方式（无浏览器、CI 自动化、Playwright 等）和业务域切换见 [认证参考](installation/auth-reference.md)。

Core 安装完成后，`deploy.sh` 会尝试导入 Context Loader 工具集 ADP（可用环境变量关闭或覆盖路径，见仓库 `deploy/deploy.sh` 说明）。若要确认是否已在平台注册工具箱，可执行：

```bash
kweaver call '/api/agent-operator-integration/v1/tool-box/list?name=contextloader&page=1&page_size=50' -bd bd_public --pretty
```

（与 `kweaver context-loader tools` 不同：前者为 Operator 工具箱列表，后者为 MCP 工具列表。）

轻量「分析助手」Agent 导入 JSON 模板见 [`sample-agent.import.json`](../examples/sample-agent.import.json)。

### 配置模型（按需）

| 能力 | 需要的模型 | 不配会怎样 |
|------|-----------|-----------|
| 数据源接入、知识网络创建、条件查询 | 无 | 正常使用 |
| 语义搜索 | Embedding（向量小模型） | 搜索报错，条件查询仍可用 |
| Agent 对话 | LLM（大语言模型） | 创建成功但对话报错 |

**不配模型也能走完数据源接入、知识网络创建和条件查询。** 语义搜索和 Agent 对话分别需要 Embedding 和 LLM；配置方法见 [模型管理](model.md)，注册 Embedding 后还需 [启用 BKN 语义搜索](model.md#启用-bkn-语义搜索)。

---

## 通过 AI 编程助手（推荐）

如果你使用 Claude Code、Codex、Cursor 等 AI 编程助手，可以用自然语言完成所有操作——不需要记任何命令和参数。

### 安装 KWeaver Skill

```bash
npx skills add https://github.com/kweaver-ai/kweaver-sdk \
  --skill kweaver-core --skill create-bkn
```

> 安装过程会弹出交互式选择器让你选择目标 AI 编程助手。如需跳过交互直接安装到所有已知助手，可加 `-y` 标志。

安装后，在 AI 编程助手中即可通过自然语言或 `/kweaver-core` 斜杠命令操作平台。

### 完整对话示例

以下是一个从零到 Agent 对话的完整流程，每一步都是在 AI 编程助手中输入的自然语言：

**接入数据源：**

> 帮我连接 MySQL 数据库，地址 db.example.com，端口 3306，库名 erp，用户 root，密码 pass123

**创建知识网络：**

> 把刚才的数据源中的 orders、products、customers 三张表创建为知识网络，名称叫"erp-供应链"

**查询数据：**

> 查一下 orders 里状态为 overdue 的前 10 条记录

**语义搜索**（需要 Embedding）：

> 在知识网络里搜索"超期订单"

**创建 Agent 并对话**（需要 LLM）：

> 创建一个 Agent，名字叫"供应链助手"，绑定刚才的知识网络，然后发布

> 跟供应链助手对话：本月有多少超期订单？

**追踪推理过程：**

> 查看刚才对话的 trace

AI 助手会自动调用 `kweaver` CLI 完成全部操作，包括数据源发现、表结构解析、主键选择、对象类型创建、Agent 发布等。遇到问题时助手也会自动排查和提示。

---

## 通过 CLI

### 场景：5 分钟内完成首次数据查询

**故事线**：你刚部署好 KWeaver Core，手头有一台 MySQL 数据库装着 ERP 数据。你的目标是把数据库变成一个知识网络，然后查询数据。

#### 第 1 步：接入数据源

```bash
kweaver ds connect mysql db.example.com 3306 erp \
  --account root --password pass123
# → 返回 ds_id，例如 ds-abc123
```

参数说明：`mysql` 为数据源类型（支持 mysql / postgresql / hive 等），后跟 **主机**、**端口**、**数据库名**，`--account` 和 `--password` 为连接凭据。

查看已有数据源和表结构：

```bash
kweaver ds list
kweaver ds tables ds-abc123
```

#### 第 2 步：创建知识网络

```bash
kweaver bkn create-from-ds ds-abc123 \
  --name "erp-供应链" \
  --tables "erp.orders,erp.products,erp.customers" \
  --build --timeout 600
```

> **表名格式**：`--tables` 需要使用 `数据库名.表名` 的全限定格式（与 `kweaver ds tables` 输出一致）。裸表名会导致 `No tables available` 错误。

这一条命令完成了：自动发现表结构 → 创建对象类型 → 映射字段。如果对象类型是 resource-backed（直接映射数据源表），`--build` 会自动跳过（不需要构建索引，数据直接从源表实时查询）；只有需要独立索引的对象类型才会执行构建。

> **注意**：`create-from-ds` 会自动选择主键（primary key）和显示键（display key）。如果源表没有明确的主键，自动选择可能不理想（如选择 `status` 字段），导致相同主键值的记录被合并。建议后续通过 `kweaver bkn object-type update` 手动指定正确的主键。

验证：

```bash
kweaver bkn object-type list <kn_id>
# 输出：orders (ot-1)、products (ot-2)、customers (ot-3)
```

#### 第 3 步：查询数据

条件查询：

```bash
kweaver bkn object-type query <kn_id> ot-1 \
  '{"limit":10,"condition":{"field":"status","operation":"==","value":"overdue"}}'
```

语义搜索（需要 Embedding 模型并 [启用 BKN 语义搜索](model.md#启用-bkn-语义搜索)）：

```bash
kweaver bkn search <kn_id> "超期订单"
```

> 即使没有 Embedding 或未启用语义搜索，上面的条件查询仍然可用。

---

### 场景：创建 Agent 并对话

**故事线**：知识网络建好了，你希望给业务团队一个自然语言接口 — 不用写 SQL，直接问问题就能得到回答。

> **前置条件**：Agent 需要 LLM；配置见 [模型管理](model.md)。语义能力还需 Embedding 并 [启用 BKN 语义搜索](model.md#启用-bkn-语义搜索)。

```bash
# 查看已注册的 LLM（获取 llm_id）
kweaver call '/api/mf-model-manager/v1/llm/list?page=1&size=50'

# 查看可用模板（--minimum 安装可能为空）
kweaver agent template-list

# 直接创建 Agent（指定 --llm-id）
kweaver agent create \
  --name "供应链助手" \
  --profile "回答供应链相关问题" \
  --llm-id <llm_id>

# 如有模板，可用模板配置创建
kweaver agent template-get <template_id> --save-config /tmp/config.json
kweaver agent create \
  --name "供应链助手" \
  --profile "回答供应链相关问题" \
  --config /tmp/config-*.json

# 绑定知识网络
kweaver agent update <agent_id> --knowledge-network-id <kn_id>

# 发布后才能对话
kweaver agent publish <agent_id>

# 单轮对话
kweaver agent chat <agent_id> -m "本月有多少超期订单？"

# 交互式多轮对话
kweaver agent chat <agent_id>
# > 哪些供应商交货最慢？
# > 给出改进建议
```

---

### 场景：追踪推理过程（Trace AI）

**故事线**：Agent 给出的回答看起来不太对，你想知道它到底查了哪些数据、调了哪些工具、每一步花了多少时间。

> **注意**：Trace 功能依赖完整的后端服务（包括 Uniquery/DataView 等组件）。仅 Core 最小部署时，Trace 接口可能返回 500 错误；此时需确认相关服务已正常运行。

```bash
# 查看会话列表
kweaver agent sessions <agent_id>

# 获取完整 trace（须同时传入智能体 ID 与会话 ID）
kweaver agent trace <agent_id> <conversation_id> --pretty
```

Trace 返回按时间排列的 Span 树，展示：
- Agent 的思考与规划过程
- 调用了哪些工具（BKN 查询、VEGA SQL、外部 API）
- 每步的输入、输出与耗时
- Context Loader 组装了哪些上下文

```
[HTTP 请求] → [意图识别] → [BKN 查询] → [SQL 执行] → [答案生成]
      ↓            ↓            ↓            ↓            ↓
   用户问题     "查超期订单"   条件过滤      3条结果      "本月有3笔..."
   已接收       识别完成       ot: orders   从 VEGA      合成回答
```

---

### 场景：从 CSV 文件构建知识网络

**故事线**：你没有数据库，只有几份 CSV 报表。

```bash
# 先找一个可用的数据源（CSV 需要一个中间存储）
kweaver ds list

# 导入 CSV 到数据源
kweaver ds import-csv <ds_id> --files "物料.csv,库存.csv" --table-prefix sc_

# 一键创建知识网络
kweaver bkn create-from-csv <ds_id> \
  --files "物料.csv,库存.csv" \
  --name "供应链报表" --build

# 验证
kweaver bkn search <kn_id> "库存为零"
```

---

### 场景：VEGA 数据视图与 SQL 查询

**故事线**：你想直接对底层数据执行 SQL，而不是通过知识网络。

```bash
# 平台健康检查
kweaver vega inspect

# 列出 catalog
kweaver vega catalog list

# 查看某个 catalog 下的资源
kweaver vega catalog resources <catalog_id> --category table

# 查找数据视图
kweaver dataview find --name "supplier_entity"

# 查询数据视图（默认使用视图定义）
kweaver dataview query <view_id> --limit 10

# 自定义 SQL 查询（需使用 catalog."schema"."table" 全限定名）
kweaver dataview query <view_id> --sql "SELECT supplier_name, city FROM <catalog>.\"supply_chain\".\"supplier_entity\" LIMIT 10"

# 全限定名请以 dataview 为准（勿手写猜 catalog）：
# kweaver dataview get <view_id> → 使用响应 JSON 字段 meta_table_name（与 vega catalog id + 源库 schema/表名 一致）
```

其中 `<catalog>` 须替换为该数据源在 **Vega** 中注册得到的 **catalog id**（见 `kweaver vega catalog list`），**不要**用视图逻辑名或裸表名代替；`"supply_chain"`、`"supplier_entity"` 分别对应源库中的 database/schema 与物理表名。**可靠做法**：`kweaver dataview get <view_id>` 取响应中的 **`meta_table_name`** 字段，在 SQL 中原样引用；`sql_str`、`fields` 含义见 [VEGA](vega.md)「数据视图」中的字段表。

仅 **Core** 部署时，`dataview query` 不带 `--sql` 可做分页、选列等结构化查询；**`--sql` 复杂自定义 SQL** 需要 **`vega-calculate-coordinator`**，由 **Etrino** 套件提供（`vega-hdfs`、`vega-calculate`、`vega-metadata`）。在 `deploy` 目录执行 `./deploy.sh etrino install` 即可。详见 [部署文档](installation/deploy.md) 与 [VEGA](vega.md)。

---

### 场景：Dataflow 流程编排

**故事线**：你有一个文档处理流水线，需要上传 PDF 触发解析。

```bash
# 列出流程
kweaver dataflow list

# 上传文件触发运行
kweaver dataflow run <dag_id> --file ./contract.pdf

# 查看今天的运行记录
kweaver dataflow runs <dag_id> --since 2026-04-14

# 查看执行日志（含输入输出）
kweaver dataflow logs <dag_id> <instance_id> --detail
```

---

## 通过 TypeScript SDK

如果你更习惯编程方式，以下 TypeScript 代码实现与上面 CLI 完全相同的流程。

> 完整示例见 [kweaver-sdk/examples](https://github.com/kweaver-ai/kweaver-sdk/tree/main/examples)。

### 最简方式（Simple API — 3 行代码）

```typescript
import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

kweaver.configure({ config: true }); // 自动读取 ~/.kweaver/ 凭据

const knList = await kweaver.bkns({ limit: 10 });
console.log(`找到 ${knList.length} 个知识网络`);

const result = await kweaver.search('超期订单', { bknId: knList[0].id, maxConcepts: 5 });
for (const c of result.concepts ?? []) {
  console.log(`${c.concept_name} (score: ${c.intent_score})`);
}
```

### 完整方式（Client API — 更多控制）

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

// 从 ~/.kweaver/ 自动读取凭据（kweaver auth login 写入的）
const client = await KWeaverClient.connect();
```

### 发现知识网络

```typescript
const knList = await client.knowledgeNetworks.list({ limit: 10 });
for (const kn of knList) {
  console.log(`${kn.name} (${kn.id})`);
}
```

### 浏览 Schema：对象类型、关系、动作

```typescript
const knId = knList[0].id;

const objectTypes = await client.knowledgeNetworks.listObjectTypes(knId);
for (const ot of objectTypes) {
  console.log(`${ot.name} — ${ot.properties?.length ?? 0} 个属性`);
}

const relationTypes = await client.knowledgeNetworks.listRelationTypes(knId);
for (const rt of relationTypes) {
  console.log(`${rt.source_object_type?.name} —[${rt.name}]→ ${rt.target_object_type?.name}`);
}

const actionTypes = await client.knowledgeNetworks.listActionTypes(knId);
```

### 查询实例与子图遍历

```typescript
const otId = objectTypes[0].id;

// 条件查询
const instances = await client.bkn.queryInstances(knId, otId, {
  limit: 5,
  condition: { field: 'status', operation: '==', value: 'overdue' },
});
console.log(instances.datas);

// 子图遍历（沿关系类型展开）
const rt = relationTypes[0];
const subgraph = await client.bkn.querySubgraph(knId, {
  relation_type_paths: [{
    relation_types: [{
      relation_type_id: rt.id,
      source_object_type_id: rt.source_object_type?.id,
      target_object_type_id: rt.target_object_type?.id,
    }],
  }],
  limit: 5,
});
```

### 语义搜索

> 需已注册 Embedding 并完成 [启用 BKN 语义搜索](model.md#启用-bkn-语义搜索)。

```typescript
const result = await client.bkn.semanticSearch(knId, '超期订单');
for (const concept of result.concepts ?? []) {
  console.log(`${concept.concept_name} (score: ${concept.intent_score})`);
}
```

### Context Loader（MCP 分层检索）

```typescript
const { baseUrl } = client.base();
const mcpUrl = `${baseUrl}/api/agent-retrieval/v1/mcp`;
const cl = client.contextLoader(mcpUrl, knId);

// Layer 1：Schema 搜索
const schema = await cl.schemaSearch({ query: '订单', max_concepts: 5 });

// Layer 2：实例查询
const mcpInstances = await cl.queryInstances({ ot_id: otId, limit: 5 });
```

### Agent 对话

```typescript
// 列出 Agent
const agents = await client.agents.list({ limit: 10 });

// 单轮对话
const reply = await client.agents.chat(agentId, '本月有多少超期订单？');
console.log(reply.text);

// 查看推理链路
for (const step of reply.progress ?? []) {
  console.log(`[${step.skill_info?.type}] ${step.skill_info?.name} → ${step.status}`);
}

// 流式对话（实时输出）
let prevLen = 0;
await client.agents.stream(agentId, '哪些供应商交货最慢？', {
  onTextDelta: (fullText) => {
    process.stdout.write(fullText.slice(prevLen));
    prevLen = fullText.length;
  },
  onProgress: (progress) => {
    for (const p of progress) {
      console.log(`[progress] ${p.skill_info?.name} → ${p.status}`);
    }
  },
});

// 会话历史
const sessions = await client.conversations.list(agentId, { limit: 5 });
const messages = await client.conversations.listMessages(conversationId, { limit: 20 });
```

---

## 接下来读什么

| 目标 | 文档 |
| --- | --- |
| 完整 BKN 操作（Schema、条件查询、Action） | [bkn.md](bkn.md) |
| 模型注册、测试与管理 | [model.md](model.md) |
| 集群中启用语义搜索（ConfigMap） | [启用 BKN 语义搜索](model.md#启用-bkn-语义搜索) |
| 数据虚拟化与 Catalog 管理 | [vega.md](vega.md) |
| Agent 全生命周期 | [decision-agent.md](decision-agent.md) |
| 流程编排详细 | [dataflow.md](dataflow.md) |
| MCP 分层检索 | [context-loader.md](context-loader.md) |
| 工具与技能管理 | [execution-factory.md](execution-factory.md) |
| 链路追踪与证据链 | [trace-ai.md](trace-ai.md) |
| 认证与安全治理 | [isf.md](isf.md) |

SDK 完整示例代码见 [kweaver-sdk/examples](https://github.com/kweaver-ai/kweaver-sdk/tree/main/examples)。
