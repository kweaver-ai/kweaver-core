# worldcup-bkn-vega（Vega Catalog · 离线 BKN 模版）

本目录是本示例 **唯一的** 检入库内 BKN 树：对象类数据源为 **`resource | {{*_RES_ID}} | wc_<stem>`**（占位符对齐 Vega **table Resource** UUID），**不经 mdl Dataview**。

详见 **[../WORKFLOW-BRANCH-VEGA.zh.md](../WORKFLOW-BRANCH-VEGA.zh.md)**。`network.bkn` 默认 **`id: worldcup_vega_catalog_bkn`**，与 `./run.sh` 临时创建的 KN 区分。

## 校验

占位符尚未替换时可先通过结构校验：

```bash
TMPDIR="$(pwd)/.tmp" mkdir -p .tmp && TMPDIR="$(pwd)/.tmp" \
  kweaver bkn validate ./worldcup-bkn-vega
```

若本机 `$TMPDIR` 不可写会导致 `mkdtemp` 失败，可如上指定可写目录。

## 填 Resource UUID

从 `kweaver vega catalog resources <catalog-id> --category table` 得到每张 `wc_*` 表的 **resource id** 后，将各 OT 末尾表格中的 **`{{FOO_RES_ID}}`** 整块替换为实际 UUID（与 `WORLD_CUP_DATASET_STEMS` 对齐的 `wc_<stem>` 名称列仍保留，便于人工核对）。

也可使用示例脚本：`../scripts/render_worldcup_bkn_vega_placeholders.py`（见 **`WORKFLOW-BRANCH-VEGA.zh.md`**）。
