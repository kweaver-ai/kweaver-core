# Examples & Templates (English)

## `sample-agent.import.json`

Minimal “analysis assistant” Agent import template (no Dolphin). `config.skills.tools` is pre-bound to **all tools in the Context Loader toolbox** (the `tool_box_id` and each `tool_id` match `adp/context-loader/agent-retrieval/docs/release/toolset/context_loader_toolset.adp` in this repo — 9 tools in total).

After importing, open Studio, replace the embedded **`REPLACE_WITH_YOUR_LLM_NAME`** with your actual model name, bind a business knowledge network, then publish.

If your environment imports the toolbox via impex and the **UUIDs differ from the repo ADP**, re-select the tools in Studio, or replace `tool_box_id` / `tool_id` based on the response from `kweaver call '/api/agent-operator-integration/v1/tool-box/list?...'`.

Endpoint: `POST /api/agent-factory/v3/agent-inout/import` (`multipart`: `file` + `import_type`).

The end-to-end supply-chain demo can still use `deploy/auto_cofig/agent.json`. The legacy one-shot helper `deploy/auto_cofig/auto_config.sh` is being phased out and is not extended here.
