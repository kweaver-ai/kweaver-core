agent_config = {
    "input": {
        "fields": [{"name": "query", "type": "string", "from": "input"}],
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
            "name": "function",
            "type": "function_block",
            "function_id": "1856271184971304960",
            "function_code": "def main(main_arg1, main_arg2):\n    return main_arg1 + main_arg2",
            "function_inputs": [
                {
                    "key": "main_arg1",
                    "value": {"from": "input", "name": "query", "type": "string"},
                },
                {
                    "key": "main_arg2",
                    "value": {"from": "input", "name": "query", "type": "string"},
                },
            ],
            "output": [
                {"from": "function", "name": "function_answer", "type": "object"}
            ],
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
