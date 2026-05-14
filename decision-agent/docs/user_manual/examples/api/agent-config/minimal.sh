#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：生成创建 Agent 所需的最小 config。
# 输出：api/.tmp/agent-config-minimal.json
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用

# 使用 jq 创建 JSON 对象。
jq -n \
  --arg llm_id "${KWEAVER_LLM_ID:-<llm-id>}" \
  --arg llm_name "${KWEAVER_LLM_NAME:-<llm-name>}" \
  '{
    input: { fields: [{ name: "query", type: "string" }] },
    output: { default_format: "markdown", variables: { answer_var: "answer" } },
    llms: [{
      is_default: true,
      llm_config: {
        id: $llm_id,
        name: $llm_name,
        model_type: "llm",
        temperature: 1,
        top_p: 1,
        top_k: 1,
        max_tokens: 1000,
        retrieval_max_tokens: 32
      }
    }]
  }' > "$API_TMP_DIR/agent-config-minimal.json"  # 生成最小配置文件

jq . "$API_TMP_DIR/agent-config-minimal.json"  # 美化输出配置文件
