# BKN 引擎

## 概述

**业务知识网络（BKN）** 是 KWeaver Core 的语义层，用**对象类型**、**关系类型**、**动作类型**描述领域，并存储**实例**与**关系**，为智能体与分析提供统一本体。

**相关模块：** [VEGA 引擎](vega.md)（视图背后的数据）、[Context Loader](context-loader.md)（基于本体的上下文）、[Decision Agent](decision-agent.md)（运行时消费 BKN）。

**运维提示：** 语义搜索由 **bkn-backend** 与 **ontology-query** 协同完成；需在 **模型工厂** 注册 Embedding，并在两侧配置与注册名一致的默认小模型（`model_name`）。步骤与排障见 [模型管理 — 启用 BKN 语义搜索](model.md#启用-bkn-语义搜索)。

---

## BKN 语言

**BKN (Business Knowledge Network)** 是一种基于 Markdown 的声明式建模语言，用于定义业务知识网络中的对象、关系和行动。BKN 只负责描述模型结构与语义，不包含执行逻辑。

完整语言规范见 [bkn-specification](https://github.com/kweaver-ai/bkn-specification)。

### 核心概念

| 概念 | 说明 |
|------|------|
| **knowledge_network** | 业务知识网络的整体集合，是顶层容器 |
| **object_type** | 业务对象类型（如 Pod、客户、订单），定义属性与数据来源 |
| **relation_type** | 连接两个对象类型的关系（如 "Pod 属于 Node"、"客户下单"） |
| **action_type** | 对对象执行的操作定义，可绑定 tool 或 MCP |
| **risk_type** | 对行动和对象的执行风险进行结构化建模 |
| **concept_group** | 将相关对象类型组织在一起的分组 |

### 文件格式

BKN 文件使用 `.bkn` 扩展名，UTF-8 编码。每个文件由两部分组成：

1. **YAML Frontmatter** — 文件元数据（`---` 包裹）
2. **Markdown Body** — 用标准 Markdown 表格和标题描述定义内容

```markdown
---
type: object_type
id: pod
name: Pod实例
tags: [容器, Kubernetes]
---

## ObjectType: Pod实例

Kubernetes 中的最小部署单元。

### Data Properties

| Name | Display Name | Type | Description | Mapped Field |
|------|--------------|------|-------------|--------------|
| id | ID | integer | 主键 | id |
| pod_name | Pod名称 | string | Pod名称 | pod_name |
| pod_status | Pod状态 | string | Running/Pending/Failed | pod_status |
| pod_node_name | 所在节点 | string | Pod所在节点名称 | pod_node_name |

### Keys

Primary Keys: id
Display Key: pod_name

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | pod_info_view | pod_info_view |
```

标题层级固定：`#` 网络标题、`##` 类型定义（`ObjectType:` / `RelationType:` / `ActionType:`）、`###` 类型内 section（Data Properties、Keys、Endpoint 等）。

### 目录结构

每个对象/关系/行动/风险独立一个文件，按类型放入子目录：

```
my-network/
├── network.bkn              # 网络根文件 (type: knowledge_network)
├── SKILL.md                 # Agent 入口（可选，agentskills.io 标准）
├── object_types/
│   ├── customer.bkn
│   └── order.bkn
├── relation_types/
│   └── customer_places_order.bkn
├── action_types/
│   └── check_order_status.bkn
├── concept_groups/
│   └── ecommerce.bkn
└── data/                    # 可选，CSV 实例数据
    └── customers.csv
```

### 更新模型

BKN 采用**无 patch 的更新模型**：

- **新增/修改** — 编辑 `.bkn` 文件并导入，按 `(network, type, id)` 执行 upsert
- **删除** — 通过 SDK/CLI 的 delete API 显式执行，不通过 BKN 文件表达

### SDK

解析、校验与转换 BKN 文件的官方 SDK：

| 语言 | 包 | 安装 |
|------|-----|------|
| Python | [PyPI](https://pypi.org/project/kweaver-bkn/) | `pip install kweaver-bkn` |
| TypeScript | [npm](https://www.npmjs.com/package/@kweaver-ai/bkn) | `npm install @kweaver-ai/bkn` |
| Golang | — | `go get github.com/kweaver-ai/bkn-specification/sdk/golang` |

---

## 用 Agent 创建 BKN

AI 编码 Agent（如 Cursor、Claude Code、Codex）可以通过 [create-bkn](https://github.com/kweaver-ai/bkn-specification/tree/main/.cursor/skills/create-bkn) 技能自动生成符合规范的 BKN 目录。

### 安装 Skill

Cursor 用户将 `create-bkn` Skill 目录放到 `~/.cursor/skills/` 或项目 `.cursor/skills/` 下。其他 Agent 环境参照各自的 Skill 加载方式。

同时需安装 [kweaver-core](https://github.com/kweaver-ai/kweaver-sdk) Skill 以获得 CLI 验证和推送能力。

### 用自然语言描述业务域

向 Agent 描述你的业务领域即可。例如：

> 帮我建一个供应链知识网络，包含"物料"、"仓库"、"库存"三个对象。物料和仓库之间通过库存关联。
> 需要一个"库存盘点"动作，绑定到库存对象上。

Agent 会自动：
1. 读取 BKN 规范，确认语法规则
2. 生成 `network.bkn` 根文件和按类型分目录的 `.bkn` 文件
3. 生成 `SKILL.md` 索引文件（Agent 可读的网络导航）
4. 交叉检查 ID 引用、标题层级、必填字段

### 校验与推送

BKN 文件生成后，使用 CLI 校验并推送到平台：

```bash
# 校验：检查格式、引用完整性
kweaver bkn validate ./my-network/

# 推送到平台（创建或更新知识网络）
kweaver bkn push ./my-network/

# 从平台拉取已有网络到本地
kweaver bkn pull <kn_id> ./export-dir/
```

### 完整流程示例

```bash
# 1. 在 Agent 中生成 BKN 目录（交互式）
#    → Agent 生成 ./supply-chain/ 目录

# 2. 校验
kweaver bkn validate ./supply-chain/

# 3. 推送到平台
kweaver bkn push ./supply-chain/

# 4. 查看已创建的知识网络
kweaver bkn list

# 5. 构建索引
kweaver bkn build <kn_id> --wait

# 6. 语义搜索验证
kweaver bkn search <kn_id> "库存不足的物料"
```

---

## 从数据源创建

除了用 Agent 编写 BKN 文件，也可以直接从已有数据源自动生成知识网络：

```bash
# 从数据库数据源创建，自动发现表结构并构建索引
kweaver bkn create-from-ds <ds_id> \
  --name "销售网络" \
  --tables orders,customers,products \
  --build --timeout 300

# 从 CSV 文件批量创建
kweaver bkn create-from-csv <ds_id> \
  --files "./data/*.csv" \
  --name "财务分析网络" \
  --build
```

---

## CLI

### 知识网络管理

```bash
kweaver bkn list --name "客户" --tag crm --sort update_time --direction desc --limit 50 -v
kweaver bkn get kn_abc123 --stats
kweaver bkn get kn_abc123 --export
kweaver bkn pull kn_abc123 ./my-network
```

### 构建与推送

```bash
kweaver bkn build kn_abc123 --wait --timeout 300
kweaver bkn validate ./my-network
kweaver bkn push ./my-network --branch main
```

### 对象类型 CRUD

```bash
kweaver bkn object-type list kn_abc123
kweaver bkn object-type get kn_abc123 ot_customer

kweaver bkn object-type create kn_abc123 \
  --name "客户" \
  --dataview-id dv_001 \
  --primary-key customer_id \
  --display-key customer_name

kweaver bkn object-type update kn_abc123 ot_customer \
  --add-property '{"name":"phone","type":"string","display_name":"联系电话"}' \
  --update-property '{"name":"email","display_name":"电子邮箱"}' \
  --remove-property legacy_field \
  --tags "核心,CRM"

kweaver bkn object-type delete kn_abc123 ot_customer
```

### 对象类型实例查询

```bash
# 等值查询
kweaver bkn object-type query kn_abc123 ot_customer \
  '{"conditions":[{"field":"status","op":"==","value":"active"}],"limit":20}'

# 模糊查询
kweaver bkn object-type query kn_abc123 ot_customer \
  '{"conditions":[{"field":"name","op":"like","value":"%张%"}],"limit":10}'

# 枚举查询（IN）
kweaver bkn object-type query kn_abc123 ot_customer \
  '{"conditions":[{"field":"region","op":"in","value":["华东","华北"]}]}'

# 组合条件查询（AND/OR）
kweaver bkn object-type query kn_abc123 ot_customer \
  '{"logic":"and","conditions":[{"field":"status","op":"==","value":"active"},{"field":"region","op":"in","value":["华东"]}]}'

# search_after 分页
kweaver bkn object-type query kn_abc123 ot_customer \
  '{"conditions":[],"limit":20,"search_after":["2024-01-15T10:30:00Z","cust_500"]}'
```

### 关系类型

```bash
kweaver bkn relation-type list kn_abc123
kweaver bkn relation-type get kn_abc123 rt_purchase
kweaver bkn relation-type create kn_abc123 \
  --name "购买" \
  --source-type ot_customer \
  --target-type ot_product
kweaver bkn relation-type delete kn_abc123 rt_purchase
```

### 语义搜索

```bash
kweaver bkn search kn_abc123 "近三个月高价值客户"
```

### 动作类型与执行

```bash
kweaver bkn action-type list kn_abc123
kweaver bkn action-type query kn_abc123 at_send_email
kweaver bkn action-type execute kn_abc123 at_send_email \
  --params '{"to":"user@example.com","subject":"提醒","body":"您好"}'

kweaver bkn action-log list kn_abc123 --limit 20
kweaver bkn action-log get kn_abc123 log_789
kweaver bkn action-log cancel kn_abc123 log_789
kweaver bkn action-execution get kn_abc123 exec_456
```

### 端到端流程

```bash
# 1. 连接数据源
kweaver ds connect --type postgresql \
  --host db.example.com --port 5432 \
  --database sales --user admin --password secret \
  --name "销售数据库"

# 2. 从数据源创建知识网络
kweaver bkn create-from-ds ds_001 --name "销售网络" --build --timeout 300

# 3. 查看生成的对象类型
kweaver bkn object-type list kn_abc123

# 4. 查询客户实例
kweaver bkn object-type query kn_abc123 ot_customer \
  '{"conditions":[{"field":"status","op":"==","value":"active"}],"limit":5}'

# 5. 语义搜索
kweaver bkn search kn_abc123 "近一年采购额超过100万的客户"
```

---

## Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")

networks = client.bkn.list_networks(name="客户", tag="crm", sort="update_time", direction="desc", limit=50)
for kn in networks["data"]:
    print(kn["id"], kn["name"])

kn = client.bkn.get("kn_abc123", stats=True)
print(f"对象类型数: {kn['stats']['object_type_count']}")

ot_list = client.bkn.object_type.list("kn_abc123")
for ot in ot_list["data"]:
    print(ot["id"], ot["name"])

results = client.bkn.object_type.query("kn_abc123", "ot_customer", {
    "conditions": [{"field": "status", "op": "==", "value": "active"}],
    "limit": 20
})
for row in results["data"]:
    print(row["customer_name"], row["region"])

search_results = client.bkn.search("kn_abc123", query="高价值客户")
for item in search_results["data"]:
    print(item["score"], item["object_type"], item["display_name"])
```

---

## TypeScript SDK

> 完整可运行示例见 [kweaver-sdk/examples](https://github.com/kweaver-ai/kweaver-sdk/tree/main/examples)

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = await KWeaverClient.connect();

const knList = await client.knowledgeNetworks.list({ limit: 50 });
for (const kn of knList) {
  console.log(`${kn.name} (${kn.id})`);
}

const knId = knList[0].id;
const detail = await client.knowledgeNetworks.get(knId, { include_statistics: true });

const objectTypes = await client.knowledgeNetworks.listObjectTypes(knId);
for (const ot of objectTypes) {
  console.log(`${ot.name} (${ot.id}) — ${ot.properties?.length ?? 0} 个属性`);
}

const relationTypes = await client.knowledgeNetworks.listRelationTypes(knId);
for (const rt of relationTypes) {
  console.log(`${rt.source_object_type?.name} —[${rt.name}]→ ${rt.target_object_type?.name}`);
}

const actionTypes = await client.knowledgeNetworks.listActionTypes(knId);

const otId = objectTypes[0].id;
const instances = await client.bkn.queryInstances(knId, otId, {
  page: 1,
  limit: 20,
});
console.log(instances.datas);

const identity = instances.datas[0]._instance_identity;
const properties = await client.bkn.queryProperties(knId, otId, { identity });

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

const result = await client.bkn.semanticSearch(knId, '高价值客户');
for (const concept of result.concepts ?? []) {
  console.log(`${concept.concept_name} (score: ${concept.intent_score})`);
}

const atId = actionTypes[0].id;
const actionDetail = await client.bkn.queryAction(knId, atId, {});
const logs = await client.bkn.listActionLogs(knId, { atId, limit: 5 });

const buildStatus = await client.knowledgeNetworks.buildAndWait(knId, {
  timeout: 300_000,
  interval: 5_000,
});
```

---

## curl

```bash
# 列出知识网络
curl -sk "https://<访问地址>/api/ontology-manager/v1/knowledge-networks?name=客户&sort=update_time&direction=desc&limit=50" \
  -H "Authorization: Bearer $(kweaver token)"

# 获取知识网络详情
curl -sk "https://<访问地址>/api/ontology-manager/v1/knowledge-networks/kn_abc123" \
  -H "Authorization: Bearer $(kweaver token)"

# 列出对象类型
curl -sk "https://<访问地址>/api/ontology-manager/v1/knowledge-networks/kn_abc123/object-types" \
  -H "Authorization: Bearer $(kweaver token)"

# 查询对象类型实例
curl -sk -X POST "https://<访问地址>/api/ontology-query/v1/query" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "kn_id": "kn_abc123",
    "object_type_id": "ot_customer",
    "conditions": [
      {"field": "status", "op": "==", "value": "active"},
      {"field": "region", "op": "in", "value": ["华东","华北"]}
    ],
    "logic": "and",
    "limit": 20
  }'

# 语义搜索
curl -sk -X POST "https://<访问地址>/api/ontology-query/v1/search" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "kn_id": "kn_abc123",
    "query": "近三个月高价值客户",
    "limit": 10
  }'

# 创建对象类型
curl -sk -X POST "https://<访问地址>/api/ontology-manager/v1/knowledge-networks/kn_abc123/object-types" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "客户",
    "dataview_id": "dv_001",
    "primary_key": "customer_id",
    "display_key": "customer_name",
    "properties": [
      {"name": "customer_id", "type": "string"},
      {"name": "customer_name", "type": "string"},
      {"name": "region", "type": "string"},
      {"name": "status", "type": "string"}
    ]
  }'

# 执行动作
curl -sk -X POST "https://<访问地址>/api/bkn-backend/v1/knowledge-networks/kn_abc123/actions/at_send_email/execute" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "params": {"to": "user@example.com", "subject": "提醒", "body": "您好"}
  }'
```
