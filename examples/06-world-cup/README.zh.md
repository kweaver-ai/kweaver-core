# 06 · 世界杯数据 → Vega Catalog → BKN → Agent 问答

> 将公开的 [Fjelstul World Cup Database](https://github.com/jfjelstul/worldcup) 落成 MySQL **`wc_*` 表**，由单脚本 **`./run.sh`** 串起 **Vega 扫描 → BKN 渲染/推送 → Agent 创建**，得到一个可直接对话的世界杯分析 Agent。

[English](./README.md)

## 主路径

```
                       ┌─ 1) 下载 CSV     从 jfjelstul/worldcup 下载 27 个 CSV（已有则跳过）
                       │
                       ├─ 2) 导入 MySQL   kweaver ds connect + ds import-csv → wc_* 表
                       │
                       ├─ 3) Vega 扫描    vega catalog create + discover --wait
                       │
   ./run.sh  ─────────►├─ 4) 渲染 BKN    map Resources → render worldcup-bkn
                       │
                       ├─ 5) Push BKN    bkn validate + push
                       │
                       ├─ 6) 上传工具箱  kweaver toolbox import + publish（按 box_name 幂等）
                       │
                       └─ 7) 创建 Agent  agent create --config + bind KN + publish
```

仓库内 checked-in 资产：
- **`worldcup-bkn.tar`** — 离线 BKN 模板（27 个对象类、29 条 `rel_*` 关系）打包成 tar；每个 OT 末行带 **`resource | {{*_RES_ID}}`** 占位；`network.bkn` 的 `id` 为 **`worldcup_vega_catalog_bkn`**。`run.sh` 渲染前会解包到 `.tmp/worldcup-bkn/`。
- **`agent-worldcup.config.json`** — Agent 配置模板（Context Loader 工具箱 + system prompt）；`run.sh` 运行时把 `data_source.knowledge_network[0].knowledge_network_id` 替换为实际 KN id。
- **`bkn_schema_and_query_toolbox.adp`** — 工具箱 ADP（Schema 读 + ontology 实例查询/Action 执行）；已对齐 43.131.23.35 k8s 部署，集群内直连 ClusterIP 服务 `bkn-backend-svc:13014` / `ontology-query-svc:13018`，路径走 `/in/v1` 内部链路；其它集群按 `server_url` / 路径 / 账号头适配。

## 数据来源与许可

CSV 来自 Joshua C. Fjelstul 的 **The Fjelstul World Cup Database**（[仓库](https://github.com/jfjelstul/worldcup)）。
- **© 2023 Joshua C. Fjelstul, Ph.D.**
- 许可：**CC-BY-SA 4.0** — [许可全文](https://creativecommons.org/licenses/by-sa/4.0/legalcode)

再分发衍生数据或教程时需保留署名并保持同许可。

**锁定版本：**`.env` 中设置 `WORLDCUP_REF`（默认 `master`，会随上游变更）。

## 前置条件

```bash
npm install -g @kweaver-ai/kweaver-sdk
kweaver auth login https://<你的平台地址>
# CLI：用 Node SDK 的 `kweaver`，避免 `which kweaver` 落到无效的 /usr/local/bin/kweaver
# MySQL：平台 DS 与 Vega 连接器均需能访问
# curl + jq + python3
```

## 快速开始

```bash
cd examples/06-world-cup
cp env.sample .env
vim .env   # 至少填 DB_*、VEGA_CATALOG_NAME

# 一键跑七步（默认 DO_TOOLBOX=1；如已有同名工具箱会自动复用并跳过 publish）
./run.sh
```

`./run.sh --help` 列出全部 flags。常用：

| 命令 | 作用 |
|------|------|
| `./run.sh` | 完整跑 1→7 |
| `./run.sh --dry-run` | 只打印计划，不调 API |
| `./run.sh --from 3` | 从 Vega 扫描起重跑（CSV 已在 MySQL） |
| `./run.sh --only 7` | 只执行 Agent 创建 |
| `./run.sh --no-publish` | Agent 留在私人空间，不发布 |

## 27 个数据集（分组）

1. **基础实体** — `tournaments`、`confederations`、`teams`、`players`、`managers`、`referees`、`stadiums`、`matches`、`awards`
2. **赛事级映射** — `qualified_teams`、`squads`、`manager_appointments`、`referee_appointments`
3. **场次出场** — `team_appearances`、`player_appearances`、`manager_appearances`、`referee_appearances`
4. **场内事件** — `goals`、`penalty_kicks`、`bookings`、`substitutions`
5. **积分榜与奖项结果** — `host_countries`、`tournament_stages`、`groups`、`group_standings`、`tournament_standings`、`award_winners`

## 故障排查

| 现象 | 处理 |
|------|------|
| `download.sh` 失败 | 检查网络，确认 `WORLDCUP_REF` 指向含 `data-csv/` 的版本 |
| `kweaver auth` 401 | `kweaver auth login` 再来一次；`kweaver config show` 核对业务域 |
| `import-csv` 触发 MySQL **Error 1118** | step 2 已用 `mysql` CLI 预建 `wc_matches` / `wc_team_appearances`（VARCHAR(255)），未装 mysql client 时需手动建表或放宽列类型 |
| `discover` 失败 | 见 [`report/troubleshooting.md`](./report/troubleshooting.md)；可设 `VEGA_CATALOG_ID` 后 `./run.sh --from 2` |
| Resource 少于 27 张表 | connector `databases` 或 discover 不全；调整 `VEGA_MYSQL_DATABASES` 后重跑 step 1 |
| `bkn build` 报 `NoneConceptType` | 全 `resource` KN 常见；保持 `DO_BKN_BUILD=0`，由 Vega/工具查询 |
| `agent create` 报 LLM 不存在 | 在 `.env` 设 `AGENT_LLM_ID=<model_id>`（`kweaver model llm list` 查一下） |

## 与示例 02 的差异

| | 02-csv-to-kn | 06-world-cup |
|---|--------------|--------------|
| 数据 | 仓库内置 3 个小 CSV | upstream 27 份 CSV（CC-BY-SA，运行时下载） |
| 建网方式 | `create-from-csv` | **MySQL + Vega Resource** + **`worldcup-bkn` push** |
| 入口 | 多脚本 | 单脚本 `./run.sh`（七步可拆） |

## 清理

`./run.sh` 不会自动删除数据源、MySQL 表、Vega catalog、KN 或 Agent；不用时在 Studio / CLI 自行清理。
