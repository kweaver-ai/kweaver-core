# 06 · 世界杯数据 → 知识网络 → Agent 问答

> 将公开的 [Fjelstul World Cup Database](https://github.com/jfjelstul/worldcup) 导入为可查询的 KWeaver 知识网络，再用 Agent 跨越比赛、进球与球员等表做数据分析类提问。

[English](./README.md)

## 要解决什么问题

历届世界杯信息分散在大量相关表里：赛事、比赛、阵容、进球、奖项等。想回答「哪位球员累计进球最多？」「东道主最好成绩如何？」通常要反复做表连接或手工透视。

本示例一次性把 **27 份 CSV** 建成一个知识网络（`kweaver bkn create-from-csv`），抽样查询 `matches` / `players` / `goals` 三类对象，可选地用 Context Loader 做 **schema 语义检索**，最后用 **Agent 对话**（把 schema 摘要 + 抽样行放进提示词）。

## 示例在做什么

```
本地下载的 CSV（download.sh）
       │
       ▼
MySQL 中间库  ◀── kweaver ds connect
       │
       ▼
知识网络     ◀── kweaver bkn create-from-csv（27 张表 → OT + 索引构建）
       │
       ├── object-type 抽样查询
       ├── context-loader kn-search（仅 schema，可选）
       └── agent chat（带 schema + 数据片段）
```

0. **下载** upstream 的 27 个 CSV（不入库，`data/` 被 gitignore）。
1. **连接** 平台可访问的 MySQL。
2. **导入并构建** 单个 KN（`--table-prefix wc_`）。
3. **列出** OT，对 `matches`、`players`、`goals` **各查 5 条**示例。
4. 若部署支持，用 Context Loader 做 **schema 检索**。
5. 使用已有 Agent（或 `kweaver agent list` 里的第一个）做一次 **对话**。

### 数据来源与许可

CSV 来自 Joshua C. Fjelstul 的 **The Fjelstul World Cup Database**（[仓库](https://github.com/jfjelstul/worldcup)）。

- **© 2023 Joshua C. Fjelstul, Ph.D.**
- 许可：**CC-BY-SA 4.0** — [许可全文](https://creativecommons.org/licenses/by-sa/4.0/legalcode)

再分发衍生数据或教程时需保留署名并保持同许可；本 README 已对脚本与教程用途做必要署名说明。

**锁定数据版本：**在 `.env` 中设置 `WORLDCUP_REF`（见 `env.sample`）为分支名、tag 或 commit SHA，用于 Raw 下载 URL。默认 `master` 会随上游变更而变。

### 27 个数据集（分组）

与上游文档一致：

1. **基础实体** — `tournaments`、`confederations`、`teams`、`players`、`managers`、`referees`、`stadiums`、`matches`、`awards`
2. **赛事级映射** — `qualified_teams`、`squads`、`manager_appointments`、`referee_appointments`
3. **场次出场** — `team_appearances`、`player_appearances`、`manager_appearances`、`referee_appearances`
4. **场内事件** — `goals`、`penalty_kicks`、`bookings`、`substitutions`
5. **积分榜与奖项结果** — `host_countries`、`tournament_stages`、`groups`、`group_standings`、`tournament_standings`、`award_winners`

## 前置条件

```bash
npm install -g @kweaver-ai/kweaver-sdk
kweaver auth login https://<你的平台地址>
# MySQL：平台能连上即可，空库亦可，导入会建表
# curl：用于 download.sh
```

## 快速开始

```bash
cd examples/06-world-cup
./download.sh
cp env.sample .env
vim .env   # 填写 DB_HOST / DB_NAME / DB_USER / DB_PASS，可选 AGENT_ID、WORLDCUP_REF
./run.sh
```

> **安全：**`.env` 与 `data/` 已加入 gitignore，勿提交凭据与下载数据。

### Agent 准备

脚本**不会**自动创建 Agent。可以：

- 在 `.env` 里设置 `AGENT_ID`，或  
- 依赖 `kweaver agent list` 取第一个；

或在 Studio 导入类似 [`help/zh/examples/sample-agent.import.json`](../../help/zh/examples/sample-agent.import.json)，配置模型并绑定 KN 后，把 `AGENT_ID` 写入 `.env` 再跑 `./run.sh`。

## 关键命令

| 命令 | 作用 |
|------|------|
| `./download.sh` | 从 `jfjelstul/worldcup` 拉取 `data-csv/*.csv` |
| `kweaver ds connect mysql …` | 注册中间数据源 |
| `kweaver bkn create-from-csv … --files 'data/*.csv' --table-prefix wc_ --build` | 一次导入全部 CSV 并建 KN |
| `kweaver bkn object-type list \| query` | 浏览 OT 与实例 |
| `kweaver context-loader kn-search '…' --kn-id … --only-schema` | 可选的 schema 语义检索 |
| `kweaver agent chat <id> -m '…'` | Agent 对话（脚本会把 schema / 抽样行塞进提示词） |

## 与示例 02 的差异

| | 02-csv-to-kn | 06-world-cup |
|---|--------------|--------------|
| 数据 | 仓库内置 3 个小 CSV | upstream 27 份 CSV（CC-BY-SA，运行时下载） |
| 表前缀 | 默认 | `wc_` 避免与已有表冲突 |
| 流程 | 含子图遍历、导出 | 聚焦 list/query、可选 CL、Agent |
| 构建超时 | 300s | 600s（表多、数据量大） |

## 故障排查

| 现象 | 处理 |
|------|------|
| `download.sh` 中 `curl` 失败 | 检查网络；确认 `WORLDCUP_REF` 对应分支/tag 上存在 `data-csv/`。 |
| `401` / OAuth | 重新 `kweaver auth login`；检查业务域 `kweaver config show`。 |
| `create-from-csv` 失败或超时 | 临时加大 `run.sh` 里 `--timeout`；或改两步：`ds import-csv` + `bkn create-from-ds --build`。 |
| 导入后 OT 数量不足 27 | 看平台侧错误；可先 `--no-build` 再修复表后 `kweaver bkn build`。 |
| `object-type query` 的 `total = 0` | 等待构建；`kweaver bkn stats` / `kweaver bkn build --wait`。 |
| Context Loader 步骤被跳过 | 当前部署可能未启用；脚本会继续。 |
| 无可用 Agent | 导入或创建 Agent，设置 `AGENT_ID`。 |

## 清理

`run.sh` 在退出时通过 `trap` **删除本次创建的 KN 与数据源**；**不会**删除 Agent。
