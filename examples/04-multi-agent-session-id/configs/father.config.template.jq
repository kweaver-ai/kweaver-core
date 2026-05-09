# Inputs (--arg):
#   $son1_key, $son1_version
#   $son2_key, $son2_version
#
# Father uses Dolphin orchestration mode (is_dolphin_mode=1).
# The dolphin DSL invokes son agents via @<agent_name>(input=value) syntax —
# the platform's executor binds them based on skills.agents and routes the
# named arguments into each son's input.fields.

. as $base
| $base
| .is_dolphin_mode = 1
| .dolphin = (
    "@exp_son1(session_id=$session_id, query=$query) -> res_1\n" +
    "@exp_son2(session_id=$session_id, query=$query) -> res_2\n" +
    "{\"res_1\": $res_1, \"res_2\": $res_2} -> answer\n"
  )
| .pre_dolphin = []
| .post_dolphin = []
| .system_prompt = ""
| .input.fields = (
    (.input.fields // []) + [{
      "name": "session_id",
      "type": "string",
      "desc": "外部传入的会话标识，需透传给子 agent"
    }]
  )
| .skills = (
    (.skills // {}) | .agents = [
      {
        "agent_key":                         $son1_key,
        "agent_version":                     $son1_version,
        "agent_timeout":                     300,
        "intervention":                      false,
        "intervention_confirmation_message": "",
        "agent_input": [
          { "enable": true, "map_type": "auto", "input_name": "session_id", "input_type": "string", "input_desc": "透传 session_id" },
          { "enable": true, "map_type": "auto", "input_name": "query",      "input_type": "string", "input_desc": "透传 query" }
        ],
        "data_source_config": { "type": "self_configured", "specific_inherit": "" },
        "llm_config":         { "type": "self_configured" }
      },
      {
        "agent_key":                         $son2_key,
        "agent_version":                     $son2_version,
        "agent_timeout":                     300,
        "intervention":                      false,
        "intervention_confirmation_message": "",
        "agent_input": [
          { "enable": true, "map_type": "auto", "input_name": "session_id", "input_type": "string", "input_desc": "透传 session_id" },
          { "enable": true, "map_type": "auto", "input_name": "query",      "input_type": "string", "input_desc": "透传 query" }
        ],
        "data_source_config": { "type": "self_configured", "specific_inherit": "" },
        "llm_config":         { "type": "self_configured" }
      }
    ] | .tools = [] | .mcps = [] | .skills = []
  )
| .output.variables.answer_var = "answer"
| .llms = (.llms // [] | map(.llm_config.max_tokens = 2048))
| .conversation_history_config = {
    "strategy":           "count",
    "count_params":       { "count_limit": 1 },
    "time_window_params": null,
    "token_limit_params": null
  }
