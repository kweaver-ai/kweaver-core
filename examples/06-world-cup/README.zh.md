# 06 · 世界杯数据 → Vega Catalog → BKN → Agent 问答

> 将公开的 [Fjelstul World Cup Database](https://github.com/jfjelstul/worldcup) 落成 MySQL **`wc_*` 表**，由单脚本 **`./run.sh`** 串起 **Vega 扫描 → BKN 推送 → 索引构建 → Vega-SQL 工具注册 → Agent 创建**，得到一个可直接对话的世界杯分析 Agent。

[English](./README.md)

## 主路径

```
                       ┌─ 1) 下载 CSV     从 jfjelstul/worldcup 下载 27 个 CSV（已有则跳过）
                       │
                       ├─ 2) 导入 MySQL   kweaver ds connect + ds import-csv → wc_* 表
                       │                 （预建 wc_matches / wc_team_appearances 为 VARCHAR(255)
                       │                   规避 MySQL Error 1118 行宽超限）
                       │
                       ├─ 3) Vega 扫描    vega catalog create + discover --wait
                       │
   ./run.sh  ─────────►├─ 4) 渲染 BKN     map vega Resources → render worldcup-bkn.tar
                       │
                       ├─ 5) Push BKN +   bkn validate + push（幂等），
                       │   建索引          再为 27 个 vega resource 建 OpenSearch dataset
                       │                  （7 张实体表追加向量 embedding）
                       │
                       ├─ 6) 上传工具箱   kweaver toolbox create + tool upload <OpenAPI>
                       │                 （注册 `vega_sql_execute`，让 agent 跑原生 SQL）
                       │
                       └─ 7) 创建 Agent   agent create --config + bind KN + publish
                                          （重跑按 name 复用同名 agent）
```

仓库内 checked-in 资产：

- **`worldcup-bkn.tar`** — 离线 BKN 模板（27 个对象类、29 条 `rel_*` 关系）打包成 tar；每个 OT 末行带 **`resource | {{*_RES_ID}}`** 占位；`network.bkn` 的 `id` 为 **`worldcup_vega_catalog_bkn`**。`run.sh` 渲染前会解包到 `.tmp/worldcup-bkn/`。
- **`agent-worldcup.config.json`** — Agent 配置模板（Context Loader 工具箱 + system prompt）；`run.sh` 运行时把 `data_source.knowledge_network[0].knowledge_network_id` 和 `vega_sql_execute` 的 tool/box id 注入进去。
- **`vega_sql_execute.openapi.json`** — SQL-execute 工具的 OpenAPI 3.0 描述。step 6 通过 `kweaver tool upload`（OpenAPI 解析器路径）注册，避开 0.7.0 `kweaver toolbox import` 把 `api_spec` 写为 null 的 bug。

## 数据来源与许可

CSV 来自 Joshua C. Fjelstul 的 **The Fjelstul World Cup Database**（[仓库](https://github.com/jfjelstul/worldcup)）。

- **© 2023 Joshua C. Fjelstul, Ph.D.**
- 许可：**CC-BY-SA 4.0** — [许可全文](https://creativecommons.org/licenses/by-sa/4.0/legalcode)

再分发衍生数据或教程时需保留署名并保持同许可。

**锁定版本：** `.env` 中设置 `WORLDCUP_REF`（默认 `master`，会随上游变更）。

## 首次启用清单（First-time setup）

`run.sh` 只自动化上述 7 步。换一台机器 + 一个新集群，下面这 6 件平台级事必须先手动做完一次：

1. **安装 KWeaver Core 平台**（k8s + bkn-backend + ontology-query + vega-backend + agent-* + mf-model-* + opensearch + minio + mariadb）。用仓库根目录的 `deploy/onboard.sh`，详见 [deploy/README.zh.md](../../deploy/README.zh.md)。**建议 `0.8.0+`**（修了 `_score` resource-path bug 与 toolbox import 写入 bug）。
2. **CLI 登录**：`kweaver auth login https://<你的平台地址>`（凭据写入 `~/.kweaver/`）。
3. **注册 LLM 模型**（agent chat 必需）：
   ```bash
   kweaver model llm add --body-file <llm.json>   # 参考 `kweaver model llm --template`
   ```
   把返回的 `model_id` 填到 `.env` 的 `AGENT_LLM_ID`。
4. **注册 embedding 模型**（向量索引需要；不注册会自动降级为关键字索引）：
   ```bash
   kweaver model small add --name text-embedding-v4-cn \
     --type embedding --batch-size 10 --max-tokens 512 --embedding-dim 1024 \
     --model-config-file <emb.json>
   ```
   `.env` 里的 `EMBEDDING_MODEL_NAME` 运行时被解析成 `model_id`（默认 `text-embedding-v4-cn`）。
5. **把 BKN 默认 embedding 写进 ConfigMap**（KN 级语义检索路径用到；本示例不强依赖）：
   ```bash
   sudo bash deploy/onboard.sh --enable-bkn-search \
     --bkn-embedding-name=text-embedding-v4-cn
   ```
6. **MySQL 准备**：建 `worldcup` 库 + 一个能从 kweaver 平台 pod 网络可达的账号。填 `.env` 的 `DB_HOST / DB_PORT / DB_NAME / DB_USER / DB_PASS`。

以上做完，`./run.sh` 就是一键 + 幂等。

## 单次执行的前置（per-shell）

```bash
npm install -g @kweaver-ai/kweaver-sdk
kweaver auth login https://<你的平台地址>
# CLI：用 Node SDK 的 `kweaver`，避免 `which kweaver` 落到无效的 /usr/local/bin/kweaver
# MySQL：平台 DS 与 Vega 连接器均需能访问
# curl + jq + python3 + （可选）本地 `mysql` 客户端用于宽表预建
```

## 快速开始

```bash
cd examples/06-world-cup
cp env.sample .env
vim .env   # 至少填 DB_*、AGENT_LLM_ID

# 一键跑七步，全部幂等
./run.sh
```

`./run.sh --help` 列出全部 flags。常用：

| 命令 | 作用 |
|------|------|
| `./run.sh` | 完整跑 1→7 |
| `./run.sh --dry-run` | 只打印计划，不调 API |
| `./run.sh --from 3` | 从 Vega 扫描起重跑（CSV 已在 MySQL） |
| `./run.sh --only 5` | 只跑 step 5（push BKN + 建索引） |
| `./run.sh --only 7` | 只跑 agent 创建/更新 |
| `./run.sh --no-publish` | Agent 留在私人空间，不发布 |
| `./run.sh --no-reuse` | 总是创建新 Agent，不复用同名 |

## Agent 的 3 个工具

跑完 step 7 后，agent 装了 3 个互补的只读工具，LLM 按问题自动挑：

| 工具 | 来自工具箱 | 什么时候用 |
|---|---|---|
| **`search_schema`** | 平台内置（`contextloader工具集_070`） | 探索对象类 / 关系；拿某 wc_* resource 的 `data_source.id`。 |
| **`query_object_instance`** | 平台内置 | 单对象类的等值 / `in` / 范围条件查询（如 `family_name='Sun' AND given_name='Wen'`）。读 step 5 建的 OpenSearch dataset，**等值查询很快**。 |
| **`vega_sql_execute`** | step 6 用 `vega_sql_execute.openapi.json` 注册 | 原生 MySQL SQL —— `ORDER BY` / `GROUP BY` / 多表 `JOIN` / `COUNT(*)` 都靠它。表名用 `{{<resource_id>}}` 占位，`resource_id` 通过 `search_schema` 拿。 |

LLM 典型链路：`search_schema` → 拿 OT + `data_source.id` → 等值过滤走 `query_object_instance`、聚合/排序走 `vega_sql_execute`。

## 实际能问什么

跑完 `./run.sh` 后开聊：

```bash
kweaver agent chat <AGENT_ID> -m '<你的问题>' --stream
```

`<AGENT_ID>` 在 step 7 完成后会被写回 `.env`。

### Q1 · 梅西在世界杯获过哪些奖
**问题**：`梅西获得过哪些奖项？`

**Agent 链路**：`search_schema(award)` → `query_object_instance(award_winners, family_name='Messi' AND given_name='Lionel')` → 返回：

- 2014 巴西世界杯 — **金球奖**
- 2022 卡塔尔世界杯 — **金球奖** + **银靴奖**

### Q2 · 孙雯每届女足世界杯进球数 + 历史排名
**问题**：`孙雯每届女足世界杯进球数？历史射手榜排第几？`

**Agent 链路**：`vega_sql_execute(SELECT tournament_name, COUNT(*) FROM goals WHERE family_name='Sun' AND given_name='Wen' GROUP BY tournament_name)` + 再来一条 SQL 算榜单 → 返回：

| 届次 | 进球 |
|---|---|
| 1991 女足 | 1 |
| 1995 女足 | 2 |
| **1999 女足** | **7**（同届还包揽金球 + 金靴） |
| 2003 女足 | 1 |
| **总计** | **11**（女足世界杯历史并列第 5） |

### Q3 · 近三届男足世界杯冠军
**问题**：`近三届男足世界杯冠军？`

**Agent 链路**：`vega_sql_execute(SELECT year, host_country, winner FROM tournaments WHERE tournament_name LIKE '%Men%' ORDER BY CAST(year AS UNSIGNED) DESC LIMIT 3)` → 返回：

- 2022 卡塔尔 → 阿根廷
- 2018 俄罗斯 → 法国
- 2014 巴西 → 德国

每个数字都从 step 2 导入的 27 张 `wc_*` 表实时拿 —— **零幻觉**。

## 27 个数据集（分组）

1. **基础实体** — `tournaments`、`confederations`、`teams`、`players`、`managers`、`referees`、`stadiums`、`matches`、`awards`
2. **赛事级映射** — `qualified_teams`、`squads`、`manager_appointments`、`referee_appointments`
3. **场次出场** — `team_appearances`、`player_appearances`、`manager_appearances`、`referee_appearances`
4. **场内事件** — `goals`、`penalty_kicks`、`bookings`、`substitutions`
5. **积分榜与奖项结果** — `host_countries`、`tournament_stages`、`groups`、`group_standings`、`tournament_standings`、`award_winners`

## 故障排查

| 现象 | 处理 |
|------|------|
| step 1 下载失败 | 检查网络，确认 `WORLDCUP_REF` 指向含 `data-csv/` 的版本 |
| `kweaver auth` 401 | `kweaver auth login` 再来一次；`kweaver config show` 核对业务域 |
| `import-csv` 触发 MySQL **Error 1118** | step 2 已用本地 `mysql` CLI 预建 `wc_matches` / `wc_team_appearances`（VARCHAR(255)），未装 mysql client 时需手动建表或放宽列类型 |
| Vega `discover` 失败 | 设 `VEGA_CATALOG_ID` 后 `./run.sh --from 4` |
| Resource 少于 27 张表 | connector `databases` 或 discover 不全；调整 `VEGA_MYSQL_DATABASES` 后重跑 step 3 |
| Step 5 某些大表被 skip | 0.7.0 平台预期行为 — `vega-backend` 对 `>2× batch_size` 的表（8 张事件表）有 batch-sync 游标 bug，会卡死循环。agent 对这些表会 fallback 到 `vega_sql_execute`。0.8.0+ 已修。 |
| Step 5 提示 `embedding model … not registered` | 要么注册一个（见首次启用清单），要么设 `DO_INDEX=0` / `EMBEDDING_MODEL_NAME=`（空）让脚本走纯关键字索引 |
| `agent create` 报 LLM 不存在 | 在 `.env` 设 `AGENT_LLM_ID=<model_id>`（`kweaver model llm list` 查一下） |

## 与示例 02 的差异

| | 02-csv-to-kn | 06-world-cup |
|---|--------------|--------------|
| 数据 | 仓库内置 3 个小 CSV | upstream 27 份 CSV（CC-BY-SA，运行时下载） |
| 建网方式 | `create-from-csv` | **MySQL + Vega Resource** + **`worldcup-bkn.tar` push** |
| 工具 | 无 | OpenAPI 注册 `vega_sql_execute`，agent 可跑原生 SQL |
| 入口 | 多脚本 | 单脚本 `./run.sh`（七步可拆，全部幂等） |

## 清理

`./run.sh` 不会自动删除数据源、MySQL 表、Vega catalog、KN、Toolbox 或 Agent；不用时在 Studio / CLI 自行清理。
