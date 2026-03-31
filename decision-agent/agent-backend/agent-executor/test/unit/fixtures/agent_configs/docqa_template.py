input_block = {
    "fields": [
        {"name": "query", "type": "string"},
        {"name": "history", "type": "list"},
    ],
    "augment": {
        "enable": True,
        "data_source": {
            "kg": [
                {"kg_id": "1", "kg_name": "人物关系图谱", "fields": ["custom_subject"]}
            ]
        },
    },
    "rewrite": {
        "enable": True,
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
        # "advanced_config": {
        #     "history_epochs": 1,
        #     "do_rewrite_query_check": True,
        #     "rewrite_qeury_at_origin_query_top_line": 2,
        #     "rewrite_qeury_at_origin_query_bottom_line": 0.5,
        #     "rewrite_qeury_at_origin_query_jaccard_bottom_line": 0.3
        # }
    },
}
logic_block_retriever = {
    "id": "1",
    "name": "Retriever_Block",
    "type": "retriever_block",
    "input": [{"name": "query", "type": "string", "from": "input"}],
    "data_source": {
        "doc": [
            {
                "ds_id": 1,
                "ds_name": "部门文档库",
                "fields": [
                    {
                        "name": "AnyDATA研发线",
                        "path": "部门文档库1/AnyDATA研发线",
                        "source": "gns://CBBB3180731847DA9CE55F262C7CD3D8/AEC0E4D9BD224763BC5BEF8D72D5866D",
                    }
                ],
            }
        ]
    },
    "output": [{"name": "retriever_output", "type": "object"}],
    "llm_config": {
        "id": "1780110534704762881",
        "name": "Qwen2-72B-Chat",
        "temperature": 0,
        "top_p": 0.95,
        "top_k": 1,
        "frequency_penalty": 0,
        "presence_penalty": 0,
        "max_tokens": 500,
    },
    "knowledge_augment": {
        "enable": True,
        "data_source": {
            "kg": [
                {"kg_id": "4", "kg_name": "人物关系图谱", "fields": ["custom_subject"]}
            ]
        },
    },
}
logic_block_llm_simple_chat = {
    "id": "2",
    "name": "LLM_Block",
    "type": "llm_block",
    "mode": "common",
    "system_prompt": "你是一位专业的的 AI 助手。请根据要求回答用户问题，如用户问题不明确，请返回相应提示。不知道答案时，请回复“抱歉，您的问题我暂时无法处理。”,并给出相应的建议。",
    "tools": [],
    "user_prompt": "根据参考文档回答问题。\n参考文档：\n'''{{retriever_output}}'''\n问题：\n'''{{query}}？'''\n回答需满足要求：\n1、第一步先回答引用参考文档的ID。\n2、然后请一步一步的思考来回答问题。\n3、只能根据相关的检索结果或者知识回答，禁止编造。\n4、如果没有相关结果，请回答“都不相关，我不知道”。\n5、这个问题对我至关重要，请认真作答。\n6、尽量用参考文档的原文。",
    "user_prompt_variables": [
        {"name": "query", "from": "input"},
        {"name": "retriever_output", "from": "Retriever_Block"},
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
        "max_tokens": 500,
    },
}
agent_config = {
    "input": input_block,
    "logic_block": [logic_block_retriever, logic_block_llm_simple_chat],
    "output": {
        "answer": [{"name": "output", "type": "string", "from": "output"}],
        "block_answer": [
            {"name": "query", "type": "string", "from": "input"},
            # {
            #     "name": "retriever_output",
            #     "type": "object",
            #     "from": "Retriever_Block"  # 召回块名称
            # },
            {"name": "llm_output", "type": "string", "from": "LLM_Block"},
        ],
    },
    "version": "3.0.1.2",  # 新建时不传，后续保存和返回时都有
}
