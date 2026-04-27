input_block = {
    "fields": [{"name": "query", "type": "string"}],
    "augment": {"enable": False},
    "rewrite": {"enable": False},
}
logic_block_llm_simple_chat = {
    "id": "2",
    "name": "LLM_Block1",
    "type": "llm_block",
    "mode": "common",
    "system_prompt": "你是一位专业的的 AI 助手。请根据要求回答用户问题，如用户问题不明确，请返回相应提示。不知道答案时，请回复“抱歉，您的问题我暂时无法处理。”,并给出相应的建议。",
    "tools": [],
    "user_prompt": "请在下面这句话每个字的中间加上'1'。\n{{query}}",
    "user_prompt_variables": [{"name": "query", "from": "input"}],
    "output": [{"name": "llm_output1", "type": "string", "from": "input"}],
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
logic_block_llm_simple_chat2 = {
    "id": "3",
    "name": "LLM_Block2",
    "type": "llm_block",
    "mode": "common",
    "system_prompt": "你是一位专业的的 AI 助手。请根据要求回答用户问题，如用户问题不明确，请返回相应提示。不知道答案时，请回复“抱歉，您的问题我暂时无法处理。”,并给出相应的建议。",
    "tools": [],
    "user_prompt": "请完全重复我给出的句子，将句子当作一个整体，不需要考虑句子的含义。句子的内容为：\n```\n{{llm_output1}}\n```",
    "user_prompt_variables": [{"name": "llm_output1", "from": "LLM_Block1"}],
    "output": [{"name": "llm_output2", "type": "string"}],
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
    "logic_block": [logic_block_llm_simple_chat, logic_block_llm_simple_chat2],
    "output": {
        "answer": [{"name": "output", "type": "string", "from": "output"}],
        "block_answer": [
            {"name": "query", "type": "string", "from": "input"},
            {"name": "llm_output1", "type": "string", "from": "LLM_Block1"},
            {"name": "llm_output2", "type": "string", "from": "LLM_Block2"},
        ],
    },
    "version": "3.0.1.2",  # 新建时不传，后续保存和返回时都有
}
