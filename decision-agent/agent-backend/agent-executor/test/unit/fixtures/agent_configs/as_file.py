agent_config = {
    "input": {
        "fields": [
            {"name": "query", "type": "string", "from": "input"},
            {"name": "doc", "type": "file", "from": "input"},
        ],
        "augment": {"enable": False, "data_source": {}},
        "rewrite": {
            "enable": False,
            "llm_config": {
                "id": "1828243657887715328",
                "name": "qwen2-7b",
                "temperature": 1,
                "top_p": 1,
                "top_k": 1,
                "frequency_penalty": 0,
                "presence_penalty": 0,
                "max_tokens": 1000,
                "icon": "",
            },
        },
    },
    "logic_block": [
        {
            "id": "1",
            "name": "LLM_Block",
            "mode": "common",
            "type": "llm_block",
            "system_prompt": "你是一个经验丰富、专业权威的文档专家，能够运用深厚的专业知识和丰富的实践经验，高效精准地解决各类文档相关的复杂问题。",
            "user_prompt": "{{query}} {{doc}}",
            "user_prompt_variables": [
                {"name": "query", "type": "string", "from": "input"},
                {"name": "doc", "type": "file", "from": "input"},
            ],
            "tools": [],
            "output": [{"name": "answer", "type": "string", "from": "LLM_Block"}],
            "llm_config": {
                "id": "1825866319950647296",
                "name": "Qwen2-72B-Chat",
                "temperature": 1,
                "top_p": 1,
                "top_k": 1,
                "frequency_penalty": 0,
                "presence_penalty": 0,
                "max_tokens": 1000,
                "icon": "",
            },
        }
    ],
    "output": {
        "answer": [{"name": "回答", "type": "string", "from": "output"}],
        "block_answer": [],
    },
    "version": "",
    "kg_ontology": {},
    "AsToken": "",
}
