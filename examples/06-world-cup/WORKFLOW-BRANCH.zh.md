# 分支流程：CSV → 入库 → 扫描数据源 → 导出 BKN → push → Agent 问答

对应 `run.sh` 的「一键 CSV → KN」路径，本分支强调 **把本地 CSV 经平台写入 MySQL**（`ds import-csv`）、**扫描数据源**、**拉出 `.bkn` 文件再校验/推送**，并可选 **绑定新建 Agent**。

这与「只是把 CSV rsync 到某台服务器」是两回事：**表与数据必须由 `import-csv` 写进你连上的 `DB_*`，平台才能扫描并建模。**

[English summary in README.md → Branch workflow](./README.md)

## 步骤与命令对照

| # | 你要做的事 | 实际操作 |
|---|------------|----------|
| 1 | **将 CSV 导入中间库（MySQL）** | `./download.sh` 拉齐 `data/` 后，脚本会先（默认）瘦身 `matches.csv` / `team_appearances.csv`（`kweaver ds import-csv` 每列等价 `VARCHAR(512)`，列数多时触发 MySQL **Error 1118**）；再 `ds connect` → **`ds import-csv`**。若要保留上游全列可自行建表或使用 `SLIM_WIDE_CSV_FOR_MYSQL=0`（可能仍失败）。 |
| 2 | 扫描数据源 | `kweaver ds tables <ds_id>`（导入前后各扫一次可看 `wc_*` 是否就绪）。 |
| 3 | （由扫描结果）得到 BKN 文件 | 脚本用 `数据库名.wc_<csv 主文件名>` 拼出全表列表，执行 `kweaver bkn create-from-ds` 建 KN，再 **`kweaver bkn pull <kn_id> <目录>`** 拉成 **本地 `.bkn` 树**。若你要 **完全手写** BKN，请在 `ds tables` 之后改用 **create-bkn 技能**，再 `validate` / `push`。 |
| 4 | 导入 BKN（校验并推送定义） | `kweaver bkn validate <目录>` → `kweaver bkn push <目录>`。可用 `DO_BKN_PUSH=false` 只校验不推送。 |
| 5 | 基于 BKN 创建 Agent | `kweaver agent create --llm-id …` → `kweaver agent update <id> --knowledge-network-id <kn_id>` → `kweaver agent publish <id>`（`AGENT_LLM_ID` 见 `env.sample`）。 |
| 6 | 问答 | `kweaver agent chat <id> -m "…"`（脚本里用 `BRANCH_CHAT_QUESTION` 打样）。 |

## 疑难：wc_matches / wc_team_appearances 建表失败（Error 1118）

日志若出现 **`Row size too large`**：`import-csv` 为每一列生成的目标类型过宽（当前 SDK 等价每列 `VARCHAR(512)`），这两张 CSV 列数多时会超过单行上限。默认 **`scripts/slim_wide_tables_for_mysql_import.py`** 会在 `run.sh` / `run-branch-bkn.sh` 灌库前去掉可经其它表恢复的冗余文字列（首次会备份为 `matches.csv.wide_backup` 等）。设 `SLIM_WIDE_CSV_FOR_MYSQL=0` 可跳过（需自担风险）。

若需要把 `./data/` **同步到远端目录**（rsync）、与数据库无关运维场景，可**单独**运行 `./upload-data.sh`，并在 `.env` 配置 `REMOTE_USER` / `REMOTE_HOST` / `REMOTE_DIR`。**`run-branch-bkn.sh` 不会调用它。**

## 一键脚本

```bash
cd examples/06-world-cup
cp env.sample .env && vim .env    # DB_*、AGENT_*、KN_NAME_BRANCH、BKN_EXPORT_DIR …
./download.sh
./run-branch-bkn.sh
```

- `TEARDOWN_BRANCH=1 ./run-branch-bkn.sh` — 脚本退出时 **自动删除** 本次新建的 KN 与数据源（默认 **不删**，便于你连 Agent 调试）。
- `./run-branch-bkn.sh --dry-run` — 只打印计划不做 API 调用。
- 平台若要求 `create-from-ds` 的 `--tables` 形态与当前 `DB_NAME.wc_*` 不一致，请用 `kweaver ds tables "$DS_ID"` 对照后改 `run-branch-bkn.sh` 内的 `build_qualified_tables()` 或手写列表。

## 与主 `run.sh` 的差异

| | `run.sh` | `run-branch-bkn.sh` |
|--|----------|----------------------|
| 建网 | `create-from-csv` | `import-csv` + `create-from-ds` + `bkn pull` |
| 产出 | 仅平台侧 KN | 本地 `bkn-export/…` 一组 `.bkn` 文件 |
| 结束 | 默认 `trap` 删 KN/DS | 默认 **保留** KN/DS |
