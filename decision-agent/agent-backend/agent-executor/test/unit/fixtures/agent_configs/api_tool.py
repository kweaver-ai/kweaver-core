input_block = {
    "fields": [{"name": "query", "type": "string"}],
    "augment": {"enable": False},
    "rewrite": {"enable": False},
}
logic_block_llm_react = {
    "id": "2",
    "name": "LLM_Block",
    "type": "llm_block",
    "mode": "common",
    "system_prompt": "你是一个擅长图谱问答的专家，能够根据用户问题及图谱信息分析推理答案，也能够借助外部工具获得额外的辅助信息。不知道答案时，请回复“抱歉，您的问题我暂时无法处理。”,并给出相应的建议。",
    "tools": [
        {
            "tool_id": "1",
            "tool_name": "fulltext_search",
            "tool_description": "在图谱中搜索相关实体",
            "tool_box_id": "1",
            "tool_box_name": "",
            "tool_use_description": "在图谱中搜索相关实体",
        }
    ],
    "user_prompt": "{{1_tool_use}}{{2_format_constraint}}开始！\n\nQuestion： {{query}}",
    "user_prompt_variables": [
        {"name": "query", "from": "input"},
        {
            "name": "1_tool_use",
            "value": "你可以不使用工具直接回答用户问题，也可以调用以下工具回答用户问题：",
        },
        {
            "name": "2_format_constraint",
            "value": "如果调用工具，请使用以下格式：\\n\\nQuestion: 你必须回答的问题\\nThought: 思考你需要做什么以及调用哪个工具可以找到答案\\nAction: 你选择使用的工具名称，工具名称必须从 [{tool_names}] 中选择。不需要调用工具时，为null\\nAction Input: 工具输入参数，不使用工具时为null\\nObservation: 调用工具后得到的结果\\n... (Thought/Action/Action Input/Observation的流程可能需要重复多次才能解决问题)\\n\\n当已经满足用户要求时，请使用以下格式：\\nThought: 我已经知道最终答案了\\nFinal Answer: 用户问题的最终答案\n",
        },
    ],
    "output": [{"name": "llm_output", "type": "string"}],
    "llm_config": {
        "id": "1780110534704762881",
        "name": "Qwen2-72B-Chat",
        "temperature": 0,
        "top_p": 0.95,
        "top_k": 1,
        "frequency_penalty": 0,
        "presence_penalty": 0,
        "max_tokens": 2000,
    },
}
agent_config = {
    "input": input_block,
    "logic_block": [logic_block_llm_react],
    "output": {
        "answer": [{"name": "output", "type": "string", "from": "output"}],
        "block_answer": [
            {"name": "query", "type": "string", "from": "input"},
            {"name": "llm_output", "type": "string", "from": "LLM_Block"},
        ],
    },
    "version": "3.0.1.2",  # 新建时不传，后续保存和返回时都有
}
