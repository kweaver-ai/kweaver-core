# VEGA 引擎

## 概述

**VEGA** 提供跨异构数据源的**数据虚拟化**：**数据连接（Catalog）**、**资源发现**、**连接器类型**与**数据视图**（含原子视图与组合视图）。智能体与应用通过统一的类 SQL 访问面查询，而无需为每个数据源单独适配。

典型 Ingress 前缀：

| 前缀 | 作用 |
| --- | --- |
| `/api/vega-backend/v1` | VEGA 后台 — 连接、元数据、查询执行 |

**相关模块：** [BKN 引擎](bkn.md)、[Context Loader](context-loader.md)、[Dataflow](dataflow.md)（数据落地与转换流程）。

## 使用方式

```bash
export KWEAVER_BASE="https://<访问地址>"
export TOKEN="<bearer-token>"
```

---

### CLI

#### 平台健康与统计

```bash
# 检查 VEGA 引擎整体健康状态
kweaver vega health

# 查看平台统计摘要（Catalog 数、资源数、连接器类型数）
kweaver vega stats

# 深度巡检：检查所有 Catalog 连接状态并输出报告
kweaver vega inspect
```

#### Catalog 管理

```bash
# 列出所有 Catalog，按状态过滤
kweaver vega catalog list --status healthy --limit 20
kweaver vega catalog list --status degraded

# 获取单个 Catalog 详情
kweaver vega catalog get cat_pg001

# 批量检查 Catalog 健康状态
kweaver vega catalog health cat_pg001 cat_mysql002 cat_es003
kweaver vega catalog health --all

# 测试 Catalog 连接是否可达
kweaver vega catalog test-connection cat_pg001

# 触发元数据发现并等待完成
kweaver vega catalog discover cat_pg001 --wait

# 列出 Catalog 下的资源（表、视图等）
kweaver vega catalog resources cat_pg001 --category table
```

#### 资源管理

```bash
# 列出资源，按 Catalog 与类别过滤
kweaver vega resource list --catalog-id cat_pg001 --category table --limit 50

# 获取资源详情（字段、类型、统计信息）
kweaver vega resource get res_orders_001

# 预览资源数据（默认 10 行）
kweaver vega resource preview res_orders_001 --limit 20

# 对资源执行数据查询（POST .../resources/:id/data，结构化 body：过滤、排序、分页）
kweaver vega resource query res_orders_001 \
  -d '{"limit":10,"offset":0,"need_total":true}'
```

#### 结构化查询与 SQL 查询（vega-backend）

以下两条命令都走 **`vega-backend`**，**不依赖** `vega-calculate-coordinator`（Trino）。适合在仅安装 KWeaver Core、已配置 MySQL/PostgreSQL Catalog 的场景下查数。

**结构化查询** — `POST /api/vega-backend/v1/query/execute`

```bash
kweaver vega query execute -d '<json>'
```

请求体要点：`tables`（必填，`resource_id` + 可选 `alias`）、`joins`（同 Catalog 内多表）、`output_fields`、`filter_condition`、`sort`、`offset` / `limit`（`limit` 最大 10000）、`need_total`。首页分页时 `query_id` 可不传；翻页需带上次返回的 `query_id`。JOIN 的 `on` 条件里 **`left_field` / `right_field` 须与 `kweaver vega resource get` 返回的 `schema_definition[].name` 一致**（逻辑名，非数据库原始列名若不同则需以 schema 为准）。**所有表必须属于同一 Catalog**，否则返回 501。

`filter_condition` 常用 `operation`：`==`/`eq`、`!=`/`not_eq`、`>`/`gt`、`>=`/`gte`、`<`/`lt`、`<=`/`lte`、`in`/`not_in`、`like`/`not_like`（仅当该字段在 schema 中为 string 类型）、`range`、`null`/`not_null`；逻辑组合用 `and`/`or` 嵌套 `sub_conditions`。叶子条件通常含 `field`、`operation`、`value`、`value_from`（常量填 `"const"`）。

单表示例：

```bash
kweaver vega query execute -d '{"tables":[{"resource_id":"res_mysql_supplier"}],"limit":5,"need_total":true}'
```

两表 JOIN 示例（请替换为真实 `resource_id` 与字段名）：

```bash
kweaver vega query execute -d '{
  "tables": [
    {"resource_id":"res_a","alias":"a"},
    {"resource_id":"res_b","alias":"b"}
  ],
  "joins":[{"type":"inner","left_table_alias":"a","right_table_alias":"b","on":[{"left_field":"fk_id","right_field":"id"}]}],
  "output_fields":["a.name","b.amount"],
  "limit":10
}'
```

**直连 SQL** — `POST /api/vega-backend/v1/resources/query`

```bash
kweaver vega sql -d '<json>'
kweaver vega sql --help
```

请求体必填：`query`（SQL 字符串，或 OpenSearch 的 DSL 对象）、`resource_type`（`mysql` | `mariadb` | `postgresql` | `opensearch`）。可选：`stream_size`（100–10000）、`query_timeout`（秒，1–3600）、`query_id`。

SQL 中可使用占位符 `{{.<资源ID>}}` 或 `{{<资源ID>}}`（资源 ID 为 Vega `resource_id`），后端替换为该资源的物理表标识。无占位符时也可写**原生 SQL**（仍需 `resource_type`），表名需符合目标库语法。

占位符示例：

```bash
kweaver vega sql -d '{
  "resource_type":"mysql",
  "query":"SELECT * FROM {{.res_mysql_supplier}} LIMIT 5"
}'
```

原生 SQL 示例（表名按实际库填写）：

```bash
kweaver vega sql -d '{
  "resource_type":"mysql",
  "query":"SELECT 1 AS one"
}'
```

**三种查询方式对照**

| 方式 | 入口 | 依赖 | 典型用途 |
|------|------|------|----------|
| 结构化查询 | `kweaver vega query execute` | vega-backend | 同 Catalog 多表 JOIN、统一 filter DSL |
| 直连 SQL | `kweaver vega sql` | vega-backend | 复杂 SQL、聚合、占位符引用资源 |
| 单资源数据 API | `kweaver vega resource query <id> -d {...}` | vega-backend | 单表过滤、sort、`search_after` 分页 |
| Dataview + `--sql` | `kweaver dataview query ... --sql` | mdl-uniquery + **Trino**（Etrino） | 跨源/复杂 SQL 经计算集群（需单独安装 Etrino） |

SDK：TypeScript `client.vega.executeQuery(body)` 为结构化查询，`client.vega.sqlQuery(body)` 为 `resources/query`；Python `client.vega.query.execute(...)` 为结构化查询（参数见 SDK），`client.vega.query.sql_query(body)` 为 `resources/query`。

#### 连接器类型

```bash
# 列出所有支持的连接器类型（PostgreSQL、MySQL、Elasticsearch 等）
kweaver vega connector-type list

# 获取连接器类型详情（支持参数、版本等）
kweaver vega connector-type get postgresql
```

#### 数据视图（Dataview）

```bash
# 列出所有数据视图
kweaver dataview list

# 按名称查找数据视图，精确匹配并等待就绪
kweaver dataview find --name "客户订单视图" --exact --wait

# 获取数据视图详情
kweaver dataview get dv_001

# 对数据视图执行 SQL 查询
kweaver dataview query dv_001 \
  --sql "SELECT customer_name, order_count FROM customer_orders WHERE region = '华东' LIMIT 20"
```

**自定义 SQL（`--sql`）与 Etrino**：不带 `--sql` 时，`dataview query` 使用视图内建定义，走直连数据源；`--sql` 会经 `vega-gateway-pro` 调用 **`vega-calculate-coordinator`**（Hetu/Presto 系引擎），该组件不在 KWeaver Core 默认清单中，需部署 **Etrino 相关 Chart**：`vega-hdfs`、`vega-calculate`（内含 coordinator）、`vega-metadata`。**不必安装 DIP**：在 `deploy` 目录执行 `./deploy.sh etrino install` 即可单独安装 Etrino；仅极简化场景可只 `helm install kweaver/vega-calculate` 并自行对齐依赖与镜像。**复杂 SQL 请使用 catalog.`"schema"."table"` 全限定名。** 步骤见 [部署文档](installation/deploy.md) 中的「可选：Etrino」。

#### 端到端流程

```bash
# 1. 查看支持的连接器
kweaver vega connector-type list

# 2. 确认 Catalog 健康
kweaver vega catalog health cat_pg001

# 3. 发现 Catalog 下的表
kweaver vega catalog discover cat_pg001 --wait
kweaver vega catalog resources cat_pg001 --category table

# 4. 预览数据
kweaver vega resource preview res_orders_001 --limit 5

# 5. 通过数据视图查询
kweaver dataview query dv_001 \
  --sql "SELECT * FROM customer_orders WHERE amount > 10000 ORDER BY amount DESC LIMIT 10"
```

---

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")

health = client.vega.health()
print(f"状态: {health['status']}")

stats = client.vega.stats()
print(f"Catalog 数: {stats['catalog_count']}, 资源数: {stats['resource_count']}")

catalogs = client.vega.catalog.list(status="healthy", limit=20)
for cat in catalogs["data"]:
    print(cat["id"], cat["name"], cat["status"])

cat_detail = client.vega.catalog.get("cat_pg001")
print(f"连接器: {cat_detail['connector_type']}, 状态: {cat_detail['status']}")

test = client.vega.catalog.test_connection("cat_pg001")
print(f"连接测试: {'成功' if test['reachable'] else '失败'}")

resources = client.vega.resource.list(catalog_id="cat_pg001", category="table", limit=50)
for res in resources["data"]:
    print(res["id"], res["name"], res["row_count"])

preview = client.vega.resource.preview("res_orders_001", limit=5)
for row in preview["rows"]:
    print(row)

result = client.vega.resources.data("res_orders_001", body={"limit": 10, "offset": 0, "need_total": True})

# 结构化查询（vega-backend /query/execute）；tables 可为 resource_id 字符串或 {"resource_id","alias"} 字典
q = client.vega.query.execute(tables=["res_orders_001"], limit=5, need_total=True)
print(q)

# 直连 SQL（vega-backend /resources/query）
sql_rows = client.vega.query.sql_query({
    "resource_type": "mysql",
    "query": "SELECT 1 AS one",
})
print(sql_rows)

connectors = client.vega.connector_type.list()
for ct in connectors["data"]:
    print(ct["id"], ct["name"], ct["version"])

dv_list = client.dataview.list()
for dv in dv_list["data"]:
    print(dv["id"], dv["name"])

dv_result = client.dataview.query("dv_001", sql="SELECT * FROM customer_orders LIMIT 5")
for row in dv_result["data"]:
    print(row)
```

---

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });

const health = await client.vega.health();
console.log('状态:', health.status);

const stats = await client.vega.stats();
console.log('Catalog 数:', stats.catalogCount, '资源数:', stats.resourceCount);

const catalogs = await client.vega.catalog.list({ status: 'healthy', limit: 20 });
catalogs.data.forEach((cat) => console.log(cat.id, cat.name, cat.status));

const detail = await client.vega.catalog.get('cat_pg001');
console.log('连接器:', detail.connectorType, '状态:', detail.status);

const test = await client.vega.catalog.testConnection('cat_pg001');
console.log('连接测试:', test.reachable ? '成功' : '失败');

const resources = await client.vega.resource.list({
  catalogId: 'cat_pg001', category: 'table', limit: 50,
});
resources.data.forEach((res) => console.log(res.id, res.name, res.rowCount));

const preview = await client.vega.resource.preview('res_orders_001', { limit: 5 });
preview.rows.forEach((row) => console.log(row));

const resData = await client.vega.queryResourceData('res_orders_001', JSON.stringify({
  limit: 10, offset: 0, need_total: true,
}));
console.log(resData);

const structured = await client.vega.executeQuery(JSON.stringify({
  tables: [{ resource_id: 'res_orders_001' }],
  limit: 5,
  need_total: true,
}));
console.log(structured);

const sqlResp = await client.vega.sqlQuery(JSON.stringify({
  resource_type: 'mysql',
  query: 'SELECT 1 AS one',
}));
console.log(sqlResp);

const dvList = await client.dataview.list();
dvList.data.forEach((dv) => console.log(dv.id, dv.name));

const dvResult = await client.dataview.query('dv_001', {
  sql: "SELECT * FROM customer_orders WHERE region = '华东' LIMIT 5",
});
dvResult.data.forEach((row) => console.log(row));
```

---

### curl

```bash
# 健康检查
curl -sk "$KWEAVER_BASE/api/vega-backend/v1/health" \
  -H "Authorization: Bearer $TOKEN"

# 列出 Catalog
curl -sk "$KWEAVER_BASE/api/vega-backend/v1/catalogs?status=healthy&limit=20" \
  -H "Authorization: Bearer $TOKEN"

# 获取 Catalog 详情
curl -sk "$KWEAVER_BASE/api/vega-backend/v1/catalogs/cat_pg001" \
  -H "Authorization: Bearer $TOKEN"

# 测试连接
curl -sk -X POST "$KWEAVER_BASE/api/vega-backend/v1/catalogs/cat_pg001/test-connection" \
  -H "Authorization: Bearer $TOKEN"

# 触发元数据发现
curl -sk -X POST "$KWEAVER_BASE/api/vega-backend/v1/catalogs/cat_pg001/discover" \
  -H "Authorization: Bearer $TOKEN"

# 列出 Catalog 下的资源
curl -sk "$KWEAVER_BASE/api/vega-backend/v1/catalogs/cat_pg001/resources?category=table" \
  -H "Authorization: Bearer $TOKEN"

# 预览资源数据
curl -sk "$KWEAVER_BASE/api/vega-backend/v1/resources/res_orders_001/preview?limit=10" \
  -H "Authorization: Bearer $TOKEN"

# 查询资源数据（结构化 body，与 CLI resource query 一致）
curl -sk -X POST "$KWEAVER_BASE/api/vega-backend/v1/resources/res_orders_001/data" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "x-http-method-override: GET" \
  -d '{"limit":10,"offset":0,"need_total":true}'

# 结构化查询 query/execute
curl -sk -X POST "$KWEAVER_BASE/api/vega-backend/v1/query/execute" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tables":[{"resource_id":"res_orders_001"}],"limit":5,"need_total":true}'

# 直连 SQL resources/query
curl -sk -X POST "$KWEAVER_BASE/api/vega-backend/v1/resources/query" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"resource_type":"mysql","query":"SELECT 1 AS one"}'

# 列出连接器类型
curl -sk "$KWEAVER_BASE/api/vega-backend/v1/connector-types" \
  -H "Authorization: Bearer $TOKEN"

# 列出数据视图
curl -sk "$KWEAVER_BASE/api/vega-backend/v1/dataviews" \
  -H "Authorization: Bearer $TOKEN"

# 对数据视图执行 SQL 查询
curl -sk -X POST "$KWEAVER_BASE/api/vega-backend/v1/dataviews/dv_001/query" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sql": "SELECT customer_name, order_count FROM customer_orders WHERE region = '\''华东'\'' LIMIT 20"
  }'
```
