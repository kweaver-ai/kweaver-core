# Inputs (--arg):
#   $agent_name              e.g. "exp_son1"
#   $skill_id                exp_session_echo skill_id
#
# Discovered platform behavior (validated 2026-04-28 against MySQL Recovery Agent):
#  - is_dolphin_mode=0 + system_prompt-only path: executor synthesizes its own
#    dolphin program and IGNORES pre_dolphin → no way to inject session_id.
#  - is_dolphin_mode=1 + custom dolphin field: full control. Use /explore/ block
#    with the registered list_skills_v2 / builtin_skill_load tools to find and
#    invoke exp_session_echo skill.
#  - skills.tools[0] (list_skills_v2) MUST configure X-Authorization tool_input
#    with map_type:var → header.authorization, otherwise proxy returns 401
#    "token is invalid".

. as $base
| $base
| .is_dolphin_mode = 1
| .dolphin = (
    # Inject session_id (which is in input.fields but not visible to LLM via $query)
    # into the prompt by string-concatenation.
    "\"[input.session_id=\" + $session_id + \"] \" + $query -> q\n" +
    "\n" +
    "/explore/(history=true)\n" +
    "你是 " + $agent_name + "。\n" +
    "用户消息开头会带 [input.session_id=XXX] 的方括号块；XXX 就是当前 session_id。\n" +
    "任务：\n" +
    "1) 调用 list_skills_v2 工具拿到全部已发布 skill 列表。\n" +
    "2) 找到 name 严格等于 exp_session_echo 的那条，拿它的 skill_id。\n" +
    "3) 调用 builtin_skill_load(skill_id) 读取 SKILL.md。\n" +
    "4) 严格按 SKILL.md 的指令输出第一行：[exp_session_echo] RECEIVED session_id=<XXX 的字面值> from " + $agent_name + "\n" +
    "之后可以补一句确认。绝不修改 session_id 的值，绝不翻译。\n" +
    "\n" +
    "用户消息：\n" +
    "$q\n" +
    "-> answer\n"
  )
| .pre_dolphin = []
| .post_dolphin = []
| .system_prompt = ""
| .input.fields = (
    (.input.fields // []) + [{
      "name": "session_id",
      "type": "string",
      "desc": "外部传入的会话标识，用于全链路透传校验"
    }]
  )
| .skills = (
    (.skills // {}) | .tools = [{
      "tool_box_id":  "f182d75a-7fd2-421a-a8e0-7064d75e39af",
      "tool_id":      "51382ef3-b35b-44a6-8a53-c670cbf53f10",
      "tool_timeout": 300,
      "tool_input": [
        { "input_name": "page",              "input_type": "integer", "map_type": "auto", "map_value": "",                     "enable": false },
        { "input_name": "page_size",         "input_type": "integer", "map_type": "auto", "map_value": "",                     "enable": false },
        { "input_name": "all",               "input_type": "boolean", "map_type": "auto", "map_value": "",                     "enable": false },
        { "input_name": "X-Authorization",   "input_type": "string",  "map_type": "var",  "map_value": "header.authorization", "enable": true,
          "input_desc": "OAuth Bearer token, mapped from header.authorization to satisfy proxy auth" },
        { "input_name": "x-account-id",      "input_type": "string",  "map_type": "auto", "map_value": "",                     "enable": false },
        { "input_name": "x-account-type",    "input_type": "string",  "map_type": "auto", "map_value": "",                     "enable": false },
        { "input_name": "x-business-domain", "input_type": "string",  "map_type": "auto", "map_value": "",                     "enable": false }
      ],
      "intervention":                      false,
      "intervention_confirmation_message": "",
      "result_process_strategies":         []
    }]
    | .agents = []
    | .mcps   = []
    | .skills = [{ "skill_id": $skill_id }]
  )
| .llms = (.llms // [] | map(.llm_config.max_tokens = 2048))
| .conversation_history_config = {
    "strategy":           "count",
    "count_params":       { "count_limit": 1 },
    "time_window_params": null,
    "token_limit_params": null
  }
