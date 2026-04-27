input_block = {
    "fields": [
        {"name": "query", "type": "string"},
        {"name": "history", "type": "list"},
    ],
    "augment": {
        "enable": True,
        "data_source": {
            "kg": [
                {"kg_id": "3", "kg_name": "人物关系图谱", "fields": ["case", "lawyer"]}
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
    "id": "retriever1",  # 新建时不传，后续保存和返回时都有
    "name": "Retriever_Block",
    "type": "retriever_block",
    "input": [{"name": "query", "type": "string", "from": "input"}],
    "data_source": {
        "kg": [
            {
                "kg_id": "3",
                "kg_name": "",
                "fields": ["lawyer", "case", "court", "customer"],
                "output_fields": ["lawyer", "case"],
            }
        ],
        "doc": [
            # {
            #     "ds_id": 1,
            #     "ds_name": "部门文档库",  # 保存时不传，返回时返回
            #     "fields": [
            #         {
            #             "name": "AnyDATA研发线",
            #             "path": "部门文档库1/AnyDATA研发线",
            #             "source": "gns:#CBBB3180731847DA9CE55F262C7CD3D8/AEC0E4D9BD224763BC5BEF8D72D5866D"
            #         }
            #     ]
            # }
        ],
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
        "max_tokens": 2000,
    },
    "knowledge_augment": {
        "enable": True,
        "data_source": {
            "kg": [
                {"kg_id": "3", "kg_name": "人物关系图谱", "fields": ["case", "lawyer"]}
            ]
        },
    },
}
logic_block_llm_1 = {
    "id": "llm_1",
    "name": "LLM_Block1",
    "type": "llm_block",
    "mode": "common",
    "system_prompt": "你是一个擅长图谱问答的专家，能够根据用户问题及图谱信息分析推理答案，也能够借助外部工具获得额外的辅助信息。不知道答案时，请回复“抱歉，您的问题我暂时无法处理。”,并给出相应的建议。",
    "tools": [
        {
            "tool_id": "0",
            "tool_name": "NL2NGQL",
            "tool_description": "将用户问题的自然语言转为图数据库查询语句，并返回查询结果。",
            "tool_box_id": "",
            "tool_box_name": "",
            "config": {
                "kg_id": "3",
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
                "schema_linking_res": {
                    "name": "retriever_output.data",
                    "from": "Retriever_Block",
                },
            },
            "tool_use_description": "将用户问题的自然语言转为图数据库查询语句，并返回查询结果。若图谱中存在用户问题需要的信息，但现有召回结果中没有，须进一步查询图数据库。",
        }
    ],
    "user_prompt": "现在我给你图谱召回的结果，请你先判断图谱召回的结果中是否有问题答案的具体信息，如果有，组织答案后返回，如果没有，必须调用工具回答问题。\n用户的问题 query = {{query}}\n图谱召回的结果 schema_linking_res = {{retriever_output.data}}\n注意concept中property是一些与用户问题相关的属性值，并不同时存在于同一实体中。\n\n{{1_tool_use}}{{2_format_constraint}}开始！\n\nQuestion： {{query}}",
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
        {"name": "retriever_output.data", "from": "Retriever_Block"},
    ],
    "output": [{"name": "llm_output1", "type": "string"}],
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
logic_block_llm_2 = {
    "id": "llm_2",
    "name": "LLM_Block2",
    "type": "llm_block",
    "mode": "common",
    "system_prompt": "你是一位专业的的 AI 助手。请根据要求回答用户问题，如用户问题不明确，请返回相应提示。不知道答案时，请回复“抱歉，您的问题我暂时无法处理。”,并给出相应的建议。",
    "tools": [],
    "user_prompt": "根据参考文档回答问题。\n参考文档：\n'''{{retriever_output.text}}'''\n问题：\n'''{{query}}？'''\n回答需满足要求：\n1、第一步先回答引用参考文档的ID。\n2、然后请一步一步的思考来回答问题。\n3、只能根据相关的检索结果或者知识回答，禁止编造。\n4、如果没有相关结果，请回答“都不相关，我不知道”。\n5、这个问题对我至关重要，请认真作答。\n6、尽量用参考文档的原文。",
    "user_prompt_variables": [
        {"name": "query", "from": "input"},
        {"name": "retriever_output.text", "from": "Retriever_Block"},
    ],
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
logic_block_llm_3 = {
    "id": "llm_3",
    "name": "LLM_Block3",
    "type": "llm_block",
    "mode": "common",
    "system_prompt": "你是一位专业的的 AI 助手。请根据要求回答用户问题，如用户问题不明确，请返回相应提示。不知道答案时，请回复“抱歉，您的问题我暂时无法处理。”,并给出相应的建议。",
    "tools": [],
    "user_prompt": "给出问题{{query}}，并给你两个可参考的答案，第一个答案：{{llm_output1}}， 第二个答案{{llm_output2}}，请你对问题做出回答",
    "user_prompt_variables": [
        {"name": "query", "from": "input"},
        {"name": "llm_output1", "from": "LLM_Block1"},
        {"name": "llm_output2", "from": "LLM_Block2"},
    ],
    "output": [{"name": "llm_output3", "type": "string"}],
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
    "logic_block": [
        logic_block_retriever,
        logic_block_llm_1,
        logic_block_llm_2,
        logic_block_llm_3,
    ],
    "output": {
        "answer": [{"name": "output", "type": "string", "from": "output"}],
        "block_answer": [
            {"name": "query", "type": "string", "from": "input"},
            # {
            #     "name": "retriever_output",
            #     "type": "object",
            #     "from": "Retriever_Block"  # 召回块名称
            # },
            {"name": "llm_output1", "type": "string", "from": "LLM_Block1"},
            {"name": "llm_output2", "type": "string", "from": "LLM_Block2"},
            {"name": "llm_output3", "type": "string", "from": "LLM_Block3"},
        ],
    },
    "version": "3.0.1.2",  # 新建时不传，后续保存和返回时都有
}
