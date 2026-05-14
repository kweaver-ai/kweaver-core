#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

# 目标：生成包含知识网络、长期记忆和相关问题开关的复杂 config。
# 知识网络只填写 knowledge_network_id。
# 输出：api/.tmp/agent-config-advanced.json
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/_common.sh"  # 加载公共函数库
require_jq  # 检查 jq 命令是否可用

# 使用 jq 创建 JSON 对象。
jq -n \
  --arg llm_id "${KWEAVER_LLM_ID:-<llm-id>}" \
  --arg llm_name "${KWEAVER_LLM_NAME:-<llm-name>}" \
  --arg kn_id "${KN_ID:-<knowledge-network-id>}" \
  '{
    input: {
      fields: [
        { name: "query", type: "string" },
        { name: "history", type: "array" }
      ]
    },
    output: { default_format: "markdown", variables: { answer_var: "answer" } },
    llms: [{
      is_default: true,
      llm_config: {
        id: $llm_id,
        name: $llm_name,
        model_type: "llm",
        temperature: 0.7,
        top_p: 0.9,
        top_k: 40,
        max_tokens: 2000,
        retrieval_max_tokens: 128
      }
    }],
    data_source: {
      knowledge_network: [
        { knowledge_network_id: $kn_id }
      ]
    },
    memory: { is_enabled: false },
    related_question: { is_enabled: false },
    plan_mode: { is_enabled: false }
  }' > "$API_TMP_DIR/agent-config-advanced.json"  # 生成高级配置文件

jq . "$API_TMP_DIR/agent-config-advanced.json"  # 美化输出配置文件
