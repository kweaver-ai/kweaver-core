agent_config = {
    "input": {
        "fields": [{"name": "query", "type": "string", "from": "input"}],
        "augment": {
            "enable": False,
            "data_source": {
                "kg": [
                    {
                        "knw_id": "1",
                        "kg_id": "16",
                        "fields": [
                            "tag",
                            "district",
                            "business",
                            "industry",
                            "person",
                            "orgnization",
                        ],
                        "output_fields": None,
                        "fields_alias": [
                            "标签 (tag)",
                            "地区 (district)",
                            "专业领域 (business)",
                            "行业 (industry)",
                            "人员 (person)",
                            "组织 (orgnization)",
                        ],
                    }
                ]
            },
        },
        "rewrite": {
            "enable": False,
            "llm_config": {
                "id": "1827877371131334656",
                "name": "Benchmark-Qwen2-72B-Chat",
                "temperature": 0.01,
                "top_p": 0.5,
                "top_k": 2,
                "frequency_penalty": 0.5,
                "presence_penalty": 0.5,
                "max_tokens": 5000,
            },
        },
    },
    "logic_block": [
        {
            "id": "1",
            "name": "Retrievers_Block",
            "type": "retriever_block",
            "input": [{"name": "query", "type": "string", "from": "input"}],
            "data_source": {
                "kg": [
                    {
                        "kg_id": "16",
                        "knw_id": "1",
                        "fields": [
                            "tag",
                            "business",
                            "district",
                            "person",
                            "industry",
                            "orgnization",
                        ],
                        "fields_alias": [
                            "标签 (tag)",
                            "专业领域 (business)",
                            "地区 (district)",
                            "人员 (person)",
                            "行业 (industry)",
                            "组织 (orgnization)",
                        ],
                        "output_fields": [
                            # "tag",
                            # "business",
                            # "district",
                            "person",
                            # "industry",
                            # "orgnization"
                        ],
                    }
                ]
            },
            "output": [
                {"name": "retrival_res", "type": "object", "from": "Retrievers_Block"}
            ],
            "llm_config": {
                "id": "1827877371131334656",
                "name": "Benchmark-Qwen2-72B-Chat",
                "temperature": 0.01,
                "top_p": 0.5,
                "top_k": 2,
                "frequency_penalty": 0.5,
                "presence_penalty": 0.5,
                "max_tokens": 5000,
            },
            "knowledge_augment": {
                "enable": True,
                "data_source": {
                    "kg": [
                        {
                            "knw_id": "1",
                            "kg_id": "16",
                            "fields": [
                                "tag",
                                "business",
                                "district",
                                "industry",
                                "person",
                                "orgnization",
                            ],
                            "output_fields": None,
                            "fields_alias": [
                                "标签 (tag)",
                                "专业领域 (business)",
                                "地区 (district)",
                                "行业 (industry)",
                                "人员 (person)",
                                "组织 (orgnization)",
                            ],
                        }
                    ]
                },
            },
        },
    ],
    "output": {
        "answer": [{"name": "output", "type": "string", "from": "output"}],
        "block_answer": [
            # {
            #     "name": "query",
            #     "type": "string",
            #     "from": "input"
            # },
            # # {
            # #     "name": "retrival_res",
            # #     "type": "string",
            # #     "from": "Retrievers_Block"
            # #
            # # },
            # {
            #     "name": "llm_ans1",
            #     "type": "string",
            #     "from": "LLM_Block1"
            # }
        ],
    },
    "version": "3.0.1.2",
}
logic_block_llm_1 = {
    "id": "2",
    "name": "LLM_Block1",
    "type": "llm_block",
    "mode": "common",
    "system_prompt": "你是一个擅长图谱问答的专家，能够根据用户问题及图谱信息分析推理答案，也能够借助外部工具获得额外的辅助信息。不知道答案时，请回复“抱歉，您的问题我暂时无法处理。”,并给出相应的建议。",
    "tools": [
        {
            "tool_id": "0",
            "tool_box_id": "built-in",
            "tool_use_description": "将用户问题的自然语言转为图数据库查询的NGQL查询语句，并返回查询结果",
            "config": {
                "llm_config": {
                    "temperature": 0.01,
                    "top_k": 2,
                    "top_p": 0.5,
                    "frequency_penalty": 0.5,
                    "id": "1827877371131334656",
                    "max_tokens": 5000,
                    "name": "Benchmark-Qwen2-72B-Chat",
                    "presence_penalty": 0.5,
                },
                "query": {"name": "query", "from": "input"},
                "schema_linking_res": {
                    "name": "retrival_res.data",
                    "from": "Retrievers_Block",
                },
                "kg_id": "16",
            },
        }
    ],
    "user_prompt": "现在我给你图谱召回的结果，并请调用工具回答问题。\n用户问题为：{{query}}\n{{1_tool_use}}{{2_format_constraint}}",
    "user_prompt_variables": [
        {"name": "1_tool_use", "value": "请调用以下工具回答用户问题："},
        {
            "name": "2_format_constraint",
            "value": "如果调用工具，请使用以下格式：\n  \nQuestion: 你必须回答的问题\nThought: 思考你需要做什么以及调用哪个工具可以找到答案\nAction: 你选择使用的工具名称，工具名称必须从 [{tool_names}] 中选择。不需要调用工具时，为null\nAction Input: 工具输入参数，以json形式返回，不使用工具时为null\nObservation: 调用工具后得到的结果。你不用返回这项。\n ... (Thought/Action/Action Input/Observation的流程可能需要重复多次才能解决问题)\n  \n当已经满足用户要求时，请使用以下格式：\nThought: 我已经知道最终答案了\nFinal Answer: 用户问题的最终答案\n还没有满足用户要求时，不要返回最终答案\n",
        },
        {"name": "query", "from": "input"},
        # {
        #     "name": "retrival_res.data",
        #     "from": "Retrievers_Block"
        # }
    ],
    "output": [{"name": "llm_ans1", "type": "string"}],
    "llm_config": {
        "id": "1827877371131334656",
        "name": "Benchmark-Qwen2-72B-Chat",
        "temperature": 0.1,
        "top_p": 0.5,
        "top_k": 2,
        "frequency_penalty": 0.5,
        "presence_penalty": 0.5,
        "max_tokens": 5000,
    },
}
agent_config["logic_block"].append(logic_block_llm_1)
