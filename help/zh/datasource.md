# 数据源管理

## 概述

**数据源管理（DS）** 负责外部数据库的**连接注册**、**表结构发现**、**CSV 导入**与**生命周期维护**。它是构建知识网络（BKN）的前置步骤——先将数据库接入平台，再通过 `bkn create-from-ds` 或 `bkn create-from-csv` 把表结构转化为知识网络。

典型 Ingress 前缀：

| 前缀 | 作用 |
| --- | --- |
| `/api/builder/v1` | 数据源连接、发现与管理 |

**相关模块：** [BKN 引擎](bkn.md)（从数据源创建知识网络）、[VEGA 引擎](vega.md)（数据虚拟化与查询）、[Dataflow](dataflow.md)（数据流转与加工）。

## 支持的数据库类型

mysql、postgresql、sqlserver、oracle、clickhouse、hive、opensearch、elasticsearch 等。使用 `kweaver vega connector-type list` 可查看当前平台已安装的连接器类型。

### CLI

#### 连接数据源

```bash
# 连接 MySQL
kweaver ds connect mysql db.example.com 3306 erp \
  --account root --password pass123
# → 返回 ds_id，例如 ds-abc123

# 连接 PostgreSQL（指定 schema 和自定义名称）
kweaver ds connect postgresql pg.example.com 5432 analytics \
  --account reader --password pass456 \
  --schema public --name "分析库"
```

参数顺序：`<数据库类型> <主机> <端口> <数据库名>`，`--account` 和 `--password` 为连接凭据。

#### 列出与查看数据源

```bash
# 列出所有数据源
kweaver ds list

# 按关键词搜索
kweaver ds list --keyword "erp"

# 按类型过滤
kweaver ds list --type mysql

# 获取单个数据源详情
kweaver ds get ds-abc123
```

#### 查看表结构

```bash
# 列出数据源下的所有表
kweaver ds tables ds-abc123

# 按关键词搜索表
kweaver ds tables ds-abc123 --keyword "order"
```

#### 导入 CSV

将本地 CSV 文件写入已有数据源（数据库），再用于创建知识网络。

```bash
# 导入多个 CSV 文件
kweaver ds import-csv ds-abc123 --files "物料.csv,库存.csv"

# 使用 glob 匹配并添加表前缀
kweaver ds import-csv ds-abc123 --files "*.csv" --table-prefix sc_

# 覆盖重建已有同名表（列结构变更后）
kweaver ds import-csv ds-abc123 --files "物料.csv" --recreate

# 大文件调整批量写入行数
kweaver ds import-csv ds-abc123 --files "大表.csv" --batch-size 1000
```

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `datasource_id` | 是 | — | 目标数据源 ID |
| `--files` | 是 | — | CSV 文件路径，逗号分隔或 glob |
| `--table-prefix` | 否 | `""` | 生成的表名前缀 |
| `--batch-size` | 否 | 500 | 每批写入行数（1–10000） |
| `--recreate` | 否 | off | 首批发 overwrite，覆盖重建表 |

#### 删除数据源

```bash
# 删除（需确认）
kweaver ds delete ds-abc123

# 跳过确认
kweaver ds delete ds-abc123 --yes
```

#### 端到端流程

```bash
# 1. 连接数据源
kweaver ds connect mysql db.example.com 3306 erp \
  --account root --password pass123
# → ds-abc123

# 2. 查看表结构
kweaver ds tables ds-abc123

# 3. 从数据源创建知识网络
kweaver bkn create-from-ds ds-abc123 \
  --name "erp-供应链" \
  --tables "orders,products,customers" \
  --build --timeout 600

# 4. 验证知识网络
kweaver bkn object-type list <kn_id>
kweaver bkn search <kn_id> "超期订单"
```

---

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")

ds_list = client.ds.list()
for ds in ds_list["data"]:
    print(ds["id"], ds["name"], ds["type"], ds["status"])

ds_list_filtered = client.ds.list(keyword="erp", type="mysql")

detail = client.ds.get("ds-abc123")
print(f"主机: {detail['host']}, 数据库: {detail['database']}, 状态: {detail['status']}")

new_ds = client.ds.connect(
    type="mysql",
    host="db.example.com",
    port=3306,
    database="erp",
    account="root",
    password="pass123",
)
print(f"数据源 ID: {new_ds['id']}")

tables = client.ds.tables("ds-abc123")
for t in tables["data"]:
    print(t["name"], t["columns"])

tables_filtered = client.ds.tables("ds-abc123", keyword="order")

import_result = client.ds.import_csv(
    datasource_id="ds-abc123",
    files=["物料.csv", "库存.csv"],
    table_prefix="sc_",
    batch_size=500,
)
print(f"导入表: {import_result['tables']}")

client.ds.delete("ds-abc123")
```

---

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });

const dsList = await client.ds.list();
dsList.data.forEach((ds) => console.log(ds.id, ds.name, ds.type, ds.status));

const dsFiltered = await client.ds.list({ keyword: 'erp', type: 'mysql' });

const detail = await client.ds.get('ds-abc123');
console.log('主机:', detail.host, '数据库:', detail.database, '状态:', detail.status);

const newDs = await client.ds.connect({
  type: 'mysql',
  host: 'db.example.com',
  port: 3306,
  database: 'erp',
  account: 'root',
  password: 'pass123',
});
console.log('数据源 ID:', newDs.id);

const tables = await client.ds.tables('ds-abc123');
tables.data.forEach((t) => console.log(t.name, t.columns));

const importResult = await client.ds.importCsv({
  datasourceId: 'ds-abc123',
  files: ['物料.csv', '库存.csv'],
  tablePrefix: 'sc_',
  batchSize: 500,
});
console.log('导入表:', importResult.tables);

await client.ds.delete('ds-abc123');
```

---

### curl

```bash
# 列出数据源
curl -sk "https://<访问地址>/api/builder/v1/datasources?page=1&size=20" \
  -H "Authorization: Bearer $(kweaver token)"

# 按类型过滤
curl -sk "https://<访问地址>/api/builder/v1/datasources?type=mysql&page=1&size=20" \
  -H "Authorization: Bearer $(kweaver token)"

# 获取数据源详情
curl -sk "https://<访问地址>/api/builder/v1/datasources/ds-abc123" \
  -H "Authorization: Bearer $(kweaver token)"

# 连接新数据源
curl -sk -X POST "https://<访问地址>/api/builder/v1/datasources" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "db.example.com",
    "port": 3306,
    "database": "erp",
    "account": "root",
    "password": "pass123"
  }'

# 列出表结构
curl -sk "https://<访问地址>/api/builder/v1/datasources/ds-abc123/tables" \
  -H "Authorization: Bearer $(kweaver token)"

# 删除数据源
curl -sk -X DELETE "https://<访问地址>/api/builder/v1/datasources/ds-abc123" \
  -H "Authorization: Bearer $(kweaver token)"
```
