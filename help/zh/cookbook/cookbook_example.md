# 从 CSV 一键建知识网络

> 写新 Recipe 请复制 [`_TEMPLATE.md`](./_TEMPLATE.md)；本文是首个示例，演示模版各段如何填。

> - **难度**：⭐ 入门
> - **耗时**：约 10 分钟
> - **涉及模块**：`bkn`、`datasource`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个名为 `supply-kn` 的知识网络（KN），每张 CSV 自动成为一个对象类（OT），可用 `bkn object-type query` 查询、`bkn search` 语义检索 —— 全程一条命令，无需手写 schema。

## 2. Prerequisites（前置条件）

- 已通过 `kweaver auth login <平台地址>` 登录。
- 业务域选择正确：`kweaver config show`；不对就 `kweaver config set-bd <uuid>`。
- 准备一个 KWeaver 可访问的 **数据源**（CSV 会先入到该数据源做中间存储）。
- 本地 CSV 文件（首行表头，UTF-8）。下文以两份为例：`物料.csv`、`库存.csv`，均含 `material_code`、`material_name` 两列。

## 3. Steps（操作步骤）

### 3.1 选/建数据源

先看现有数据源，从中挑一个：

```bash
kweaver ds list
```

如果没有合适的，连接一个新的（示例为 MySQL）：

```bash
kweaver ds connect mysql db.example.com 3306 erp \
  --account root --password pass123
# → 返回 ds_id
```

> 选好后记下 **`<ds_id>`**。下面把它当成已知量。

### 3.2 一键从 CSV 建 KN

```bash
kweaver bkn create-from-csv <ds_id> \
  --files "物料.csv,库存.csv" \
  --name "supply-kn" \
  --table-prefix sc_
# → 自动完成：CSV 入库 → 创建 dataview → 创建 OT → 构建索引
# → 返回 kn_id
```

参数速查：

| 参数 | 是否必填 | 说明 |
| --- | --- | --- |
| `<ds_id>` | 是 | 用于落 CSV 的数据源 ID |
| `--files` | 是 | 逗号分隔或 glob，例如 `"*.csv"` |
| `--name` | 是 | 知识网络名 |
| `--table-prefix` | 否 | 表名前缀，避免和已有表冲突 |
| `--build` / `--no-build` | 否 | 默认 `--build`；`--no-build` 跳过构建 |
| `--timeout` | 否 | 构建等待超时秒数（默认 300） |

<details>
<summary>等价的两步路径（想自定义主键 / 显示键时用）</summary>

```bash
kweaver ds import-csv <ds_id> --files "*.csv" --table-prefix sc_
kweaver bkn create-from-ds <ds_id> --name "supply-kn" --build
```

分步路径下可在 `bkn object-type create` 时用 `--primary-key` / `--display-key` 显式指定字段。

</details>

### 3.3 验证 KN 可用

```bash
# 列对象类，确认每张 CSV 都生成了一个 OT
kweaver bkn object-type list <kn_id>

# 抽样查询（限制 limit，避免大宽表 JSON 截断）
kweaver bkn object-type query <kn_id> <ot_id> '{"limit":5}'

# 语义检索
kweaver bkn search <kn_id> "物料"
```

## 4. Expected output（期望输出）

> **判定成功的依据**：`object-type query` 返回的 `total > 0`，且 `datas[0]` 包含你导入的 CSV 列；`bkn search` 返回非空 `concepts`。

`object-type query` 应返回类似：

```jsonc
{
  "total": 1234,
  "datas": [
    {
      "_instance_identity": "...",
      "material_code": "M-001",
      "material_name": "螺丝",
      // ... 其它列
    }
  ]
}
```

`bkn search` 的 `concepts` 列表非空说明检索通道正常。

## 5. Troubleshooting（常见问题）

> 「现象」列写**用户能直接看到的具体输出 / 报错**，便于复制搜索。

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| `401 Unauthorized` 或返回体含 `oauth info is not active` | token 过期 | `kweaver auth login <平台地址>` |
| `kweaver bkn object-type list <kn_id>` 输出 `[]` | CSV 路径错 / glob 没匹配到 | 确认 `--files` 路径，必要时改用绝对路径 |
| `object-type query` 响应中 `total = 0` | 构建未完成或映射错 | `kweaver bkn stats <kn_id>` 看 `doc_count`；必要时 `kweaver bkn build <kn_id> --wait --timeout 600` 重建 |
| `ds import-csv` 报 `table already exists` | 同名表已存在 | 首批加 `--recreate`：`kweaver ds import-csv <ds_id> --files "*.csv" --recreate` |
| 自动选出的主键不是业务唯一键 | 启发式无法识别 | 走分步路径，`kweaver bkn object-type create` 显式 `--primary-key` / `--display-key` |
| `bkn search` 返回 `HTTP 500` | 视图不支持全文检索 | 把查询 `condition` 从 `match` 改为 `like` |

## 6. See also（延伸阅读）

- 参考：[BKN 引擎](../manual/bkn.md) · [数据源管理](../manual/datasource.md) · [快速开始](../quick-start.md)
- 完整示例项目：仓库内 [`examples/02-csv-to-kn/`](../../../examples/02-csv-to-kn/)
- Agent 导入模板：[`../examples/sample-agent.import.json`](../examples/sample-agent.import.json)（可在 KN 建好后绑定使用）
