# 可选流程：Vega Catalog（物理库直连）→ Resource → BKN OT（不经 Dataview）

适用于：**MySQL 已建好 `wc_*` 表和数据**（例如 **`./download.sh` + `./run.sh`**、`kweaver ds import-csv` 或手工灌库）。若 **不想走 mdl Dataview**，改为在 **Vega 注册 Catalog**，Discovery 产出 **table Resource**，再通过 **`resource | <uuid>`** 绑定 BKN OT；等价 CLI：**`kweaver bkn object-type create … --dataview-id <vega_resource_id>`** ——参数名仍为 `--dataview-id`，此处填 **Vega Resource UUID**（语义 `data_source.type = resource`，与 mdl Dataview UUID 不同）。详见 **`kweaver vega`** 与平台文档。

绑定 Resource 的对象类通常 **走 Vega 实时查询**，一般不执行离线 **`bkn build`**（与 **`create-from-csv`** 一步建 KN 的策略不同）。

## 本仓库的离线模版：`worldcup-bkn-vega/`

每个 OT 末行：**`resource | {{*_RES_ID}} | wc_<stem>`**；占位符可与 **`scripts/render_worldcup_bkn_vega_placeholders.py --mapping mappings.json`** 批量替换生成 `.rendered-bkn-vega/`。**`network.bkn`** 默认 **`id: worldcup_vega_catalog_bkn`**。

占位符未替换也可做结构校验：**`TMPDIR="$(pwd)/.tmp" kweaver bkn validate ./worldcup-bkn-vega`**（见 [`worldcup-bkn-vega/README.zh.md`](./worldcup-bkn-vega/README.zh.md)）。

[English](./WORKFLOW-BRANCH-VEGA.md)

## 一键：推送 BKN（自动填 Resource UUID）

在完成 **Catalog + `discover --wait`** 后（参见 **`./run-branch-vega.sh`**），可运行 **`./run-vega-bkn-push.sh`**：

1. `vega catalog resources` 列出 table Resource，按 **`source_identifier`**（`worldcup.wc_<stem>`）自动生成 **27** 条占位符映射。
2. 渲染 **`worldcup-bkn-vega`** → **`.rendered-bkn-vega/`**
3. **`kweaver bkn validate`** → **`kweaver bkn push`**，`kn_id` 为 **`network.bkn`** 里的 **`worldcup_vega_catalog_bkn`**

**索引进阶**：本平台对 **全部为 `resource` 数据源的对象类**一般不派发离线 **`bkn build`** 任务；若 **`DO_BKN_BUILD=1`**，可能收到 **`NoneConceptType` / 「no indexable object types」**，属预期；线上查询可走 **Vega / Resource**，与 **`create-from-csv` + `bkn build`** 的离线索引路径不同。

## 前置

- CLI 登录与业务域：`kweaver auth login …`，`kweaver config show`。
- Vega 后端能访问的 **MySQL 地址**（`host` 多为集群内可达 IP/DNS，不是本机专有地址）。
- **Discover**：`kweaver vega catalog discover <id> --wait` 依赖环境具备文档所述条件（常见部署问题见[`report/troubleshooting.md`](./report/troubleshooting.md)）；失败时可按下方「手动 Resource」兜底。

## 步骤与命令

### 1. 注册 Vega Catalog（物理连接）

```bash
kweaver vega catalog create \
  --name "<name>" \
  --connector-type mysql \
  --connector-config '{"host":"…","port":3306,"username":"…","password":"…","databases":["<db-name>"]}' \
  --pretty
```

- 连接器字段以 `kweaver vega connector-type get mysql` 为准。
- `databases` 为 Vega 扫描范围；与本示例一致时通常为含 `wc_*` 的库名。

记下返回的 **`catalog_id`**，或：`kweaver vega catalog list` 按名称查找。

### 2. Discovery → 暴露 table Resource

```bash
kweaver vega catalog discover <catalog-id> --wait
```

若自动发现不可用，可对每张物理表手动建 Resource：

```bash
kweaver vega resource create \
  --catalog-id <catalog-id> \
  --name "<logical-name>" \
  --category table \
  -d '{"source_identifier":"<db>.<table>"}'
```

`source_identifier` 形态以部署文档为准。

### 3. 查看 Resource ID

```bash
kweaver vega catalog resources <catalog-id> --category table [--limit N]
# 或
kweaver vega resource list --catalog-id <catalog-id> --category table
```

`resource list` / `catalog resources` 返回的 **`id`** 即后续绑定 OT 用的 UUID。

列元数据：**`resource get` 不一定带列**，可用 `kweaver vega resource query <resource-id> -d '{"limit":1}'` 抽样推断主键 / 展示列。

### 4. BKN：先建空 KN，再按表建 OT

```bash
kweaver bkn create --name "<kn-name>"
# 记下 kn_id（或 kn list 按名称查）

kweaver bkn object-type create <kn_id> \
  --name "<ot-name>" \
  --dataview-id <vega_resource_id> \
  --primary-key <pk> \
  --display-key <dk>
```

- **`--dataview-id`**：此处填入 **Vega Resource** 的 id（CLI 沿用该参数名；**不要**填入 mdl **Dataview** UUID）。
- `ot-name`：建议与本示例表语义一致（如 `matches`、`teams`）。
- **`pk` / `dk`**：必须与库表一致；上游 `wc_*` 的常见情况需对照实际 DDL（本仓库不提供自动推断）。
- **27 张 `wc_<stem>` 表**：stem 列表见 `scripts/worldcup_dataset_stems.inc.sh`；每张表对应一次 `object-type create`（或之后在本地 `.bkn` 批量维护再 `validate`/`push`）。

关系类型（relation types）仍可后续在 Studio 或通过 `.bkn` 补充；纯 Catalog→OT **只覆盖「对象类接表」**，不自动生成全图关系。

### 5. 校验与推送定义

**推荐使用本仓库 [`worldcup-bkn-vega/`](./worldcup-bkn-vega)**：填写 **`{{*_RES_ID}}`**（或 **`scripts/render_worldcup_bkn_vega_placeholders.py`** 生成 `.rendered-bkn-vega/`），再：

```bash
mkdir -p .tmp
TMPDIR="$(pwd)/.tmp" kweaver bkn validate ./worldcup-bkn-vega
# 使用渲染副本时替换路径为 ./.rendered-bkn-vega

kweaver bkn push ./worldcup-bkn-vega
```

若仅在平台用 `object-type create` / Studio，可用 `kweaver bkn pull <kn_id> <目录>` 拉回对齐版本库。

**数据源类型须一致：**本模版使用 **`resource`**；请勿把 Vega Resource UUID 填入 **`data_view | …`** 块，否则会校验失败。

### 6. Agent

与主流程相同：绑定 `knowledge_network_id`，配置 Context Loader 工具等；参见 [README.zh.md](./README.zh.md)。

## 一键辅助（可选）

[`./run-branch-vega.sh`](./run-branch-vega.sh) 可按 `.env` 中的 **共用 `DB_*` + Vega 专有变量** 尝试：**创建 Catalog → `discover --wait`**；**不会**自动为 27 张表批量建 OT（因各表 PK/DK 与 Resource 映射需你对照）。

- `./run-branch-vega.sh --dry-run`：只打印计划。
- 仍需人工：`vega catalog resources` / `bkn object-type create` 循环或使用自有脚本拼接。
