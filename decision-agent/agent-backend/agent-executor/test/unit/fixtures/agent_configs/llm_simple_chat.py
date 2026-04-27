input_block = {
    "fields": [
        {"name": "query", "type": "string"},
        {"name": "content", "type": "string"},
    ],
    "augment": {"enable": False},
    "rewrite": {"enable": False},
}
logic_block_llm_simple_chat = {
    "id": "2",
    "name": "LLM_Block",
    "type": "llm_block",
    "mode": "common",
    "system_prompt": "你是一位专业的的 AI 助手。请根据要求回答用户问题，如用户问题不明确，请返回相应提示。不知道答案时，请回复“抱歉，您的问题我暂时无法处理。”,并给出相应的建议。",
    "tools": [],
    "user_prompt": "根据参考文档回答问题。\n参考文档：\n'''{{content}}'''\n问题：\n'''{{query}}？'''\n回答需满足要求：\n1、第一步先回答引用参考文档的ID。\n2、然后请一步一步的思考来回答问题。\n3、只能根据相关的检索结果或者知识回答，禁止编造。\n4、如果没有相关结果，请回答“都不相关，我不知道”。\n5、这个问题对我至关重要，请认真作答。\n6、尽量用参考文档的原文。",
    "user_prompt_variables": [
        {"name": "query", "from": "input", "type": "string"},
        {"name": "content", "from": "input", "type": "string"},
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
    "logic_block": [logic_block_llm_simple_chat],
    "output": {
        "answer": [{"name": "output", "type": "string", "from": "output"}],
        "block_answer": [
            {"name": "query", "type": "string", "from": "input"},
            {"name": "llm_output", "type": "string", "from": "LLM_Block"},
        ],
    },
    "version": "3.0.1.2",  # 新建时不传，后续保存和返回时都有
}
