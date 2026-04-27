def get_model_restful_api_document(llm_id):
    api = {
        "openapi": "3.0.2",
        "info": {
            "title": "大语言模型服务",
            "version": "1.0.0",
            "description": "大语言模型服务RESTful API文档"
        },
        "paths": {
            "/api/model-factory/v1/llm-used/{}".format(llm_id): {
                "post": {
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/serviceReq"
                                },
                                "examples": {
                                    "request": {
                                        "value": {
                                            "ai_system": "",
                                            "ai_user": "",
                                            "ai_assistant": "",
                                            "ai_history": [
                                                {
                                                    "role": "ai",
                                                    "message": ""
                                                },
                                                {
                                                    "role": "human",
                                                    "message": ""
                                                }
                                            ],
                                            "top_p": 1,
                                            "temperature": 1,
                                            "max_token": 16,
                                            "frequency_penalty": 1,
                                            "presence_penalty": 1
                                        }
                                    }
                                }
                            }
                        },
                        "required": False
                    },
                    "tags": [
                        "LLM Service"
                    ],
                    "parameters": [
                        {
                            "$ref": "#/components/parameters/ServiceUserTokenColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserTimeStampColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserAppKeyColumn"
                        }
                    ],
                    "responses": {
                        "200": {
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/LLMResp"
                                    },
                                    "examples": {
                                        "resp": {
                                            "value": {
                                                "res": "这是调用大预言模型的返回结果"
                                            }
                                        }
                                    }
                                }
                            },
                            "description": "ok"
                        },
                        "500": {
                            "$ref": "#/components/responses/RespStaSE500"
                        }
                    },
                    "summary": "大语言模型接口"
                }
            }
        },
        "components": {
            "schemas": {
                "serviceReq": {
                    "description": "使用服务请求体",
                    "type": "object",
                    "properties": {
                        "ai_system": {
                            "description": "系统角色",
                            "type": "string"
                        },
                        "ai_user": {
                            "description": "用户角色",
                            "type": "string"
                        },
                        "ai_assistant": {
                            "description": "助手角色",
                            "type": "string"
                        },
                        "ai_history": {
                            "description": "历史对话",
                            "type": "array",
                            "items": {}
                        },
                        "top_p": {
                            "description": "核采样",
                            "type": "number"
                        },
                        "temperature": {
                            "description": "随机性",
                            "type": "number"
                        },
                        "max_token": {
                            "description": "单次回复限制",
                            "type": "integer"
                        },
                        "frequency_penalty": {
                            "description": "频率惩罚度",
                            "type": "number"
                        },
                        "presence_penalty": {
                            "description": "话题新鲜度",
                            "type": "number"
                        }
                    }
                },
                "Error": {
                    "description": "接口调用错误信息基类，具体错误情况可查看字段ErrorDetails",
                    "required": [
                        "code",
                        "solution",
                        "description",
                        "detail",
                        "link"
                    ],
                    "type": "object",
                    "properties": {
                        "description": {
                            "description": "导致此错误的原因",
                            "type": "string"
                        },
                        "code": {
                            "description": "业务错误码，'.' 字符前为微服务名， '.' 字符后为具体错误码",
                            "type": "string"
                        },
                        "solution": {
                            "description": "解决错误办法",
                            "type": "string"
                        },
                        "link": {
                            "description": "错误链接",
                            "type": "string"
                        },
                        "detail": {
                            "description": "错误详情",
                            "type": "array",
                            "items": {
                                "type": "object",
                                "properties": {
                                    "details": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    }
                },
                "LLMResp": {
                    "description": "",
                    "required": [
                        "res"
                    ],
                    "type": "object",
                    "properties": {
                        "res": {
                            "description": "调用大模型结果",
                            "type": "object",
                            "properties": {
                                "time": {
                                    "description": "大模型生成结果的耗时",
                                    "type": "string"
                                },
                                "token_len": {
                                    "description": "大模型返回结果的token总数",
                                    "type": "integer"
                                },
                                "data": {
                                    "description": "调用提示词的返回结果",
                                    "type": "string"
                                }
                            }
                        }
                    }
                }
            },
            "responses": {
                "RespStaSE500": {
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/Error"
                            },
                            "examples": {
                                "Err500": {
                                    "$ref": "#/components/examples/Status500"
                                }
                            }
                        }
                    },
                    "description": "Bad Request"
                }
            },
            "parameters": {
                "ServiceUserTokenColumn": {
                    "name": "appid",
                    "description": "appid: 用户登录AnyDATA 通过 /manager/v1/appid 获取的AnyDATA账户的APPID，用于数据科学家和开发者调用或者使用AnyDATA服务时的唯一标识",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": True
                },
                "ServiceUserTimeStampColumn": {
                    "name": "timestamp",
                    "description": "timestamp: 此时间戳的接受范围以AD的系统时间为基准，接受客户端的时间在AD的系统有效期为30分钟，误差时间15分钟",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                },
                "ServiceUserAppKeyColumn": {
                    "name": "appkey",
                    "description": "appkey: 客户端使用特定的算法通过appid、时间戳和请求参数这三个参数生成的校验字符串，用于后端校验请求的合法性",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                }
            },
            "examples": {
                "Status500": {
                    "summary": "请求参数错误",
                    "description": "以下原因均会导致该错误\n- 当前大模型不存在\n- 参数错误\n",
                    "value": {
                        "description": "Parma Error",
                        "code": "LLMUsed.ParameterError",
                        "detail": "",
                        "solution": "",
                        "link": ""
                    }
                }
            }
        },
        "tags": [
            {
                "name": "LLM Service",
                "description": "大语言模型服务API"
            }
        ]
    }
    return api


def get_prompt_restful_api_document(prompt_id, var_dict, prompt):
    api = {
        "openapi": "3.0.2",
        "info": {
            "title": "提示词服务",
            "version": "1.0.0",
            "description": "提示词服务RESTful API文档"
        },
        "paths": {
            "/api/model-factory/v1/prompt/{}/used".format(prompt_id): {
                "post": {
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/serviceReq"
                                },
                                "examples": {
                                    "request": {
                                        "value": {
                                            "inputs": var_dict,
                                            "history_dia": [
                                                {
                                                    "role": "ai",
                                                    "message": ""
                                                },
                                                {
                                                    "role": "human",
                                                    "message": ""
                                                }
                                            ]
                                        }
                                    }
                                }
                            }
                        },
                        "required": True
                    },
                    "tags": [
                        "Prompt Service"
                    ],
                    "parameters": [
                        {
                            "$ref": "#/components/parameters/ServiceUserTokenColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserTimeStampColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserAppKeyColumn"
                        }
                    ],
                    "responses": {
                        "200": {
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/PromptResp"
                                    },
                                    "examples": {
                                        "resp": {
                                            "value": {
                                                "res": {
                                                    "time": 1.4746403948576325,
                                                    "token_len": 1,
                                                    "data": "hello"
                                                }
                                            }
                                        }
                                    }
                                }
                            },
                            "description": "ok"
                        },
                        "500": {
                            "$ref": "#/components/responses/RespStaSE500"
                        }
                    },
                    "summary": "提示词工程接口",
                    "description": "您当前使用的提示词为：{}".format(prompt)
                }
            }
        },
        "components": {
            "schemas": {
                "serviceReq": {
                    "description": "使用服务请求体",
                    "type": "object",
                    "properties": {
                        "inputs": {
                            "description": "变量值",
                            "type": "object",
                            "items": {}
                        },
                        "history_dia": {
                            "description": "历史对话",
                            "type": "array",
                            "items": {}
                        }
                    }
                },
                "Error": {
                    "description": "接口调用错误信息基类，具体错误情况可查看字段ErrorDetails",
                    "required": [
                        "code",
                        "solution",
                        "description",
                        "detail",
                        "link"
                    ],
                    "type": "object",
                    "properties": {
                        "description": {
                            "description": "导致此错误的原因",
                            "type": "string"
                        },
                        "code": {
                            "description": "业务错误码，'.' 字符前为微服务名， '.' 字符后为具体错误码",
                            "type": "string"
                        },
                        "solution": {
                            "description": "解决错误办法",
                            "type": "string"
                        },
                        "link": {
                            "description": "错误链接",
                            "type": "string"
                        },
                        "detail": {
                            "description": "错误详情",
                            "type": "array",
                            "items": {
                                "type": "object",
                                "properties": {
                                    "details": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    }
                },
                "PromptResp": {
                    "description": "提示词服务响应结构体",
                    "required": [
                        "res"
                    ],
                    "type": "object",
                    "properties": {
                        "res": {
                            "description": "结果",
                            "type": "object",
                            "properties": {
                                "time": {
                                    "description": "大模型生成结果的耗时",
                                    "type": "string"
                                },
                                "token_len": {
                                    "description": "大模型返回结果的token总数",
                                    "type": "integer"
                                },
                                "data": {
                                    "description": "调用提示词的返回结果",
                                    "type": "string"
                                }
                            }
                        }
                    }
                }
            },
            "responses": {
                "RespStaSE500": {
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/Error"
                            },
                            "examples": {
                                "Err400": {
                                    "$ref": "#/components/examples/Status500"
                                }
                            }
                        }
                    },
                    "description": "Bad Request"
                }
            },
            "parameters": {
                "ServiceUserTokenColumn": {
                    "name": "appid",
                    "description": "appid: 用户登录AnyDATA 通过 /manager/v1/appid 获取的AnyDATA账户的APPID，用于数据科学家和开发者调用或者使用AnyDATA服务时的唯一标识",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": True
                },
                "ServiceUserTimeStampColumn": {
                    "name": "timestamp",
                    "description": "timestamp: 此时间戳的接受范围以AD的系统时间为基准，接受客户端的时间在AD的系统有效期为30分钟，误差时间15分钟",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                },
                "ServiceUserAppKeyColumn": {
                    "name": "appkey",
                    "description": "appkey: 客户端使用特定的算法通过appid、时间戳和请求参数这三个参数生成的校验字符串，用于后端校验请求的合法性",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                }
            },
            "examples": {
                "Status500": {
                    "summary": "请求参数错误",
                    "description": "以下原因均会导致该错误\n- 提示词变量输入异常\n- 历史信息字数超过限制\n- 当前prompt未发布\n",
                    "value": {
                        "description": "Parma Error",
                        "code": "PromptUsed.ParameterError",
                        "detail": "",
                        "solution": "",
                        "link": ""
                    }
                }
            }
        },
        "tags": [
            {
                "name": "Prompt Service",
                "description": "提示词服务API"
            }
        ]
    }
    return api


def get_embedding_restful_api_document(model_name):
    api = {
        "openapi": "3.0.2",
        "info": {
            "title": "embedding小模型",
            "version": "1.0.0",
            "description": "embedding小模型 RESTful API 文档"
        },
        "paths": {
            "/api/model-factory/v1/small_model_run": {
                "post": {
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/serviceReq"
                                },
                                "examples": {
                                    "request": {
                                        "value": {
                                            "model_name": model_name,
                                            "param_data": {
                                                "texts": ["test"]
                                            }
                                        }
                                    }
                                }
                            }
                        },
                        "required": True
                    },
                    "parameters": [
                        {
                            "$ref": "#/components/parameters/ServiceUserTokenColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserTimeStampColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserAppKeyColumn"
                        }
                    ],
                    "responses": {
                        "200": {
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/EmbeddingResp"
                                    },
                                    "examples": {
                                        "resp": {
                                            "value": [
                                                [
                                                    1.1843433380126953,
                                                    0.7108592987060547,
                                                    -0.11932545900344849,
                                                    0.15900762379169464
                                                ]
                                            ]
                                        }
                                    }
                                }
                            },
                            "description": "ok"
                        },
                        "500": {
                            "$ref": "#/components/responses/RespStaSE500"
                        }
                    },
                    "summary": "embedding推理接口"
                }
            }
        },
        "components": {
            "schemas": {
                "serviceReq": {
                    "description": "使用服务请求体",
                    "type": "object",
                    "properties": {
                        "model_name": {
                            "description": "模型名称",
                            "type": "string"
                        },
                        "param_data": {
                            "description": "推理参数",
                            "type": "object",
                            "properties": {
                                "texts": {
                                    "description": "需向量化的文本",
                                    "type": "array",
                                    "items": {}
                                }
                            }
                        }
                    }
                },
                "Error": {
                    "description": "接口调用错误信息基类，具体错误情况可查看字段detail",
                    "required": [
                        "code",
                        "solution",
                        "description",
                        "detail",
                        "link"
                    ],
                    "type": "object",
                    "properties": {
                        "description": {
                            "description": "导致此错误的原因",
                            "type": "string"
                        },
                        "code": {
                            "description": "业务错误码，'.' 字符前为微服务名， '.' 字符后为具体错误码",
                            "type": "string"
                        },
                        "solution": {
                            "description": "解决错误办法",
                            "type": "string"
                        },
                        "link": {
                            "description": "错误链接",
                            "type": "string"
                        },
                        "detail": {
                            "description": "错误详情",
                            "type": "array",
                            "items": {
                                "type": "object",
                                "properties": {
                                    "details": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    }
                },
                "EmbeddingResp": {
                    "description": "Embedding响应结构体",
                    "required": [
                        "res"
                    ],
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {}
                    }
                }
            },
            "responses": {
                "RespStaSE500": {
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/Error"
                            },
                            "examples": {
                                "Err400": {
                                    "$ref": "#/components/examples/Status500"
                                }
                            }
                        }
                    },
                    "description": "Bad Request"
                }
            },
            "parameters": {
                "ServiceUserTokenColumn": {
                    "name": "appid",
                    "description": "appid: 用户登录AnyDATA 通过 /manager/v1/appid 获取的AnyDATA账户的APPID，用于数据科学家和开发者调用或者使用AnyDATA服务时的唯一标识",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": True
                },
                "ServiceUserTimeStampColumn": {
                    "name": "timestamp",
                    "description": "timestamp: 此时间戳的接受范围以AD的系统时间为基准，接受客户端的时间在AD的系统有效期为30分钟，误差时间15分钟",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                },
                "ServiceUserAppKeyColumn": {
                    "name": "appkey",
                    "description": "appkey: 客户端使用特定的算法通过appid、时间戳和请求参数这三个参数生成的校验字符串，用于后端校验请求的合法性",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                }
            },
            "examples": {
                "Status500": {
                    "summary": "请求参数错误",
                    "value": {
                        "code": "ModelFactory.SmallModelRouter.SmallModelRun.ParameterError",
                        "description": "缺少参数 model_name",
                        "detail": "缺少参数 model_name",
                        "solution": "请检查填写的参数是否正确。",
                        "link": ""
                    }
                }
            }
        }
    }
    return api


def get_reranker_restful_api_document(model_name):
    api = {
        "openapi": "3.0.2",
        "info": {
            "title": "reranker小模型",
            "version": "1.0.0",
            "description": "reranker小模型 RESTful API 文档"
        },
        "paths": {
            "/api/model-factory/v1/small_model_run": {
                "post": {
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/serviceReq"
                                },
                                "examples": {
                                    "request": {
                                        "value": {
                                            "model_name": model_name,
                                            "param_data": {
                                                "slices": ["中国有56个民族", "中国", "美国"],
                                                "query": "中国有多少民族"
                                            }
                                        }
                                    }
                                }
                            }
                        },
                        "required": True
                    },
                    "parameters": [
                        {
                            "$ref": "#/components/parameters/ServiceUserTokenColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserTimeStampColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserAppKeyColumn"
                        }
                    ],
                    "responses": {
                        "200": {
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/RerankerResp"
                                    },
                                    "examples": {
                                        "resp": {
                                            "value": [
                                                1.1843433380126953,
                                                0.7108592987060547,
                                                -0.11932545900344849,
                                                0.15900762379169464
                                            ]
                                        }
                                    }
                                }
                            },
                            "description": "ok"
                        },
                        "500": {
                            "$ref": "#/components/responses/RespStaSE500"
                        }
                    },
                    "summary": "reranker推理接口"
                }
            }
        },
        "components": {
            "schemas": {
                "serviceReq": {
                    "description": "使用服务请求体",
                    "type": "object",
                    "properties": {
                        "model_name": {
                            "description": "模型名称",
                            "type": "string"
                        },
                        "param_data": {
                            "description": "推理参数",
                            "type": "object",
                            "properties": {
                                "slices": {
                                    "description": "需计算相似度得分的一组文本",
                                    "type": "array",
                                    "items": {}
                                },
                                "query": {
                                    "description": "计算相似度得分的问题",
                                    "type": "string"
                                }
                            }
                        }
                    }
                },
                "Error": {
                    "description": "接口调用错误信息基类，具体错误情况可查看字段detail",
                    "required": [
                        "code",
                        "solution",
                        "description",
                        "detail",
                        "link"
                    ],
                    "type": "object",
                    "properties": {
                        "description": {
                            "description": "导致此错误的原因",
                            "type": "string"
                        },
                        "code": {
                            "description": "业务错误码，'.' 字符前为微服务名， '.' 字符后为具体错误码",
                            "type": "string"
                        },
                        "solution": {
                            "description": "解决错误办法",
                            "type": "string"
                        },
                        "link": {
                            "description": "错误链接",
                            "type": "string"
                        },
                        "detail": {
                            "description": "错误详情",
                            "type": "array",
                            "items": {
                                "type": "object",
                                "properties": {
                                    "details": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    }
                },
                "RerankerResp": {
                    "description": "Reranker响应结构体",
                    "required": [
                        "res"
                    ],
                    "type": "array",
                    "items": {
                        "type": "number"
                    }
                }
            },
            "responses": {
                "RespStaSE500": {
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/Error"
                            },
                            "examples": {
                                "Err400": {
                                    "$ref": "#/components/examples/Status500"
                                }
                            }
                        }
                    },
                    "description": "Bad Request"
                }
            },
            "parameters": {
                "ServiceUserTokenColumn": {
                    "name": "appid",
                    "description": "appid: 用户登录AnyDATA 通过 /manager/v1/appid 获取的AnyDATA账户的APPID，用于数据科学家和开发者调用或者使用AnyDATA服务时的唯一标识",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": True
                },
                "ServiceUserTimeStampColumn": {
                    "name": "timestamp",
                    "description": "timestamp: 此时间戳的接受范围以AD的系统时间为基准，接受客户端的时间在AD的系统有效期为30分钟，误差时间15分钟",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                },
                "ServiceUserAppKeyColumn": {
                    "name": "appkey",
                    "description": "appkey: 客户端使用特定的算法通过appid、时间戳和请求参数这三个参数生成的校验字符串，用于后端校验请求的合法性",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                }
            },
            "examples": {
                "Status500": {
                    "summary": "请求参数错误",
                    "value": {
                        "code": "ModelFactory.SmallModelRouter.SmallModelRun.ParameterError",
                        "description": "缺少参数 model_name",
                        "detail": "缺少参数 model_name",
                        "solution": "请检查填写的参数是否正确。",
                        "link": ""
                    }
                }
            }
        }
    }
    return api


def get_spr_restful_api_document(model_name):
    api = {
        "openapi": "3.0.2",
        "info": {
            "title": "spr小模型",
            "version": "1.0.0",
            "description": "spr小模型 RESTful API 文档"
        },
        "paths": {
            "/api/model-factory/v1/small_model_run": {
                "post": {
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/serviceReq"
                                },
                                "examples": {
                                    "request": {
                                        "value": {
                                            "model_name": model_name,
                                            "param_data": [
                                                {
                                                    "query_text": "中国国石油天然气股份有限公司天津销售分公司中生加油站的简介是什么",
                                                    "entity_dict_list": [{"企业": "中国国石油天然气股份有限公司天津销售分公司中生加油站"}],
                                                    "property_list": ["简介"],
                                                    "relation_list": [""]
                                                },
                                                {
                                                    "query_text": "同辉佳视（北京）信息技术股份有限公司的股东教育程度是什么",
                                                    "entity_dict_list": [{"企业": "同辉佳视（北京）信息技术股份有限公司"}],
                                                    "property_list": ["教育程度"],
                                                    "relation_list": ["股东"]
                                                }
                                            ]
                                        }
                                    }
                                }
                            }
                        },
                        "required": True
                    },
                    "parameters": [
                        {
                            "$ref": "#/components/parameters/ServiceUserTokenColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserTimeStampColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserAppKeyColumn"
                        }
                    ],
                    "responses": {
                        "200": {
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/SprResp"
                                    },
                                    "examples": {
                                        "resp": {
                                            "value": [
                                                "entity_property",
                                                "entity_relation_pro"
                                            ]
                                        }
                                    }
                                }
                            },
                            "description": "ok"
                        },
                        "500": {
                            "$ref": "#/components/responses/RespStaSE500"
                        }
                    },
                    "summary": "spr推理接口"
                }
            }
        },
        "components": {
            "schemas": {
                "serviceReq": {
                    "description": "使用服务请求体",
                    "type": "object",
                    "properties": {
                        "model_name": {
                            "description": "模型名称",
                            "type": "string"
                        },
                        "param_data": {
                            "description": "推理参数",
                            "type": "array",
                            "items": {
                                "type": "object",
                                "properties": {
                                    "query_text": {
                                        "description": "查询文本",
                                        "type": "string"
                                    },
                                    "entity_dict_list": {
                                        "description": "实体字典列表",
                                        "type": "array",
                                        "items": {
                                            "type": "object"
                                        }
                                    },
                                    "property_list": {
                                        "description": "属性列表",
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
                                    },
                                    "relation_list": {
                                        "description": "关系列表",
                                        "type": "array",
                                        "items": {
                                            "type": "string"
                                        }
                                    }
                                }
                            }
                        }
                    }
                },
                "Error": {
                    "description": "接口调用错误信息基类，具体错误情况可查看字段detail",
                    "required": [
                        "code",
                        "solution",
                        "description",
                        "detail",
                        "link"
                    ],
                    "type": "object",
                    "properties": {
                        "description": {
                            "description": "导致此错误的原因",
                            "type": "string"
                        },
                        "code": {
                            "description": "业务错误码，'.' 字符前为微服务名， '.' 字符后为具体错误码",
                            "type": "string"
                        },
                        "solution": {
                            "description": "解决错误办法",
                            "type": "string"
                        },
                        "link": {
                            "description": "错误链接",
                            "type": "string"
                        },
                        "detail": {
                            "description": "错误详情",
                            "type": "array",
                            "items": {
                                "type": "object",
                                "properties": {
                                    "details": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    }
                },
                "SprResp": {
                    "description": "Spr响应结构体",
                    "required": [
                        "res"
                    ],
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            },
            "responses": {
                "RespStaSE500": {
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/Error"
                            },
                            "examples": {
                                "Err400": {
                                    "$ref": "#/components/examples/Status500"
                                }
                            }
                        }
                    },
                    "description": "Bad Request"
                }
            },
            "parameters": {
                "ServiceUserTokenColumn": {
                    "name": "appid",
                    "description": "appid: 用户登录AnyDATA 通过 /manager/v1/appid 获取的AnyDATA账户的APPID，用于数据科学家和开发者调用或者使用AnyDATA服务时的唯一标识",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": True
                },
                "ServiceUserTimeStampColumn": {
                    "name": "timestamp",
                    "description": "timestamp: 此时间戳的接受范围以AD的系统时间为基准，接受客户端的时间在AD的系统有效期为30分钟，误差时间15分钟",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                },
                "ServiceUserAppKeyColumn": {
                    "name": "appkey",
                    "description": "appkey: 客户端使用特定的算法通过appid、时间戳和请求参数这三个参数生成的校验字符串，用于后端校验请求的合法性",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                }
            },
            "examples": {
                "Status500": {
                    "summary": "请求参数错误",
                    "value": {
                        "code": "ModelFactory.SmallModelRouter.SmallModelRun.ParameterError",
                        "description": "缺少参数 model_name",
                        "detail": "缺少参数 model_name",
                        "solution": "请检查填写的参数是否正确。",
                        "link": ""
                    }
                }
            }
        }
    }
    return api


def get_info_extract_restful_api_document(model_name):
    api = {
        "openapi": "3.0.2",
        "info": {
            "title": "info_extract小模型",
            "version": "1.0.0",
            "description": "info_extract小模型 RESTful API 文档"
        },
        "paths": {
            "/api/model-factory/v1/small_model_run": {
                "post": {
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/serviceReq"
                                },
                                "examples": {
                                    "request": {
                                        "value": {
                                            "model_name": model_name,
                                            "param_data": {
                                                "schema": ["人物", "time"],
                                                "texts": [
                                                    "2月8日上午北京冬奥会自由式滑雪女子大跳台决赛中中国选手谷爱凌以188.25分获得金牌！",
                                                    "In 1997, Steve was excited to become the CEO of Apple."
                                                ]
                                            }
                                        }
                                    }
                                }
                            }
                        },
                        "required": True
                    },
                    "parameters": [
                        {
                            "$ref": "#/components/parameters/ServiceUserTokenColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserTimeStampColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserAppKeyColumn"
                        }
                    ],
                    "responses": {
                        "200": {
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/InfoExtractResp"
                                    },
                                    "examples": {
                                        "resp": {
                                            "value": {
                                                "res": [
                                                    {
                                                        "人物": [
                                                            {
                                                                "text": "谷爱凌",
                                                                "start": 28,
                                                                "end": 31,
                                                                "probability": 0.9972447470570458
                                                            }
                                                        ],
                                                        "time": [
                                                            {
                                                                "text": "2月8日上午",
                                                                "start": 0,
                                                                "end": 6,
                                                                "probability": 0.9955962371222924
                                                            }
                                                        ]
                                                    },
                                                    {
                                                        "人物": [
                                                            {
                                                                "text": "Steve",
                                                                "start": 9,
                                                                "end": 14,
                                                                "probability": 0.9998301338845295
                                                            }
                                                        ]
                                                    }
                                                ]
                                            }
                                        }
                                    }
                                }
                            },
                            "description": "ok"
                        },
                        "500": {
                            "$ref": "#/components/responses/RespStaSE500"
                        }
                    },
                    "summary": "info_extract推理接口"
                }
            }
        },
        "components": {
            "schemas": {
                "serviceReq": {
                    "description": "使用服务请求体",
                    "type": "object",
                    "properties": {
                        "model_name": {
                            "description": "模型名称",
                            "type": "string"
                        },
                        "param_data": {
                            "description": "推理参数",
                            "type": "object",
                            "properties": {
                                "schema": {
                                    "description": "抽取schema",
                                    "type": "array",
                                    "items": {
                                        "type": "string"
                                    }
                                },
                                "texts": {
                                    "description": "要抽取的文本内容",
                                    "type": "array",
                                    "items": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    }
                },
                "Error": {
                    "description": "接口调用错误信息基类，具体错误情况可查看字段detail",
                    "required": [
                        "code",
                        "solution",
                        "description",
                        "detail",
                        "link"
                    ],
                    "type": "object",
                    "properties": {
                        "description": {
                            "description": "导致此错误的原因",
                            "type": "string"
                        },
                        "code": {
                            "description": "业务错误码，'.' 字符前为微服务名， '.' 字符后为具体错误码",
                            "type": "string"
                        },
                        "solution": {
                            "description": "解决错误办法",
                            "type": "string"
                        },
                        "link": {
                            "description": "错误链接",
                            "type": "string"
                        },
                        "detail": {
                            "description": "错误详情",
                            "type": "array",
                            "items": {
                                "type": "object",
                                "properties": {
                                    "details": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    }
                },
                "InfoExtractResp": {
                    "description": "info_extract响应结构体",
                    "type": "object",
                    "properties": {
                        "res": {
                            "type": "array",
                            "items": {
                                "type": "object"
                            }
                        }
                    }
                }
            },
            "responses": {
                "RespStaSE500": {
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/Error"
                            },
                            "examples": {
                                "Err400": {
                                    "$ref": "#/components/examples/Status500"
                                }
                            }
                        }
                    },
                    "description": "Bad Request"
                }
            },
            "parameters": {
                "ServiceUserTokenColumn": {
                    "name": "appid",
                    "description": "appid: 用户登录AnyDATA 通过 /manager/v1/appid 获取的AnyDATA账户的APPID，用于数据科学家和开发者调用或者使用AnyDATA服务时的唯一标识",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": True
                },
                "ServiceUserTimeStampColumn": {
                    "name": "timestamp",
                    "description": "timestamp: 此时间戳的接受范围以AD的系统时间为基准，接受客户端的时间在AD的系统有效期为30分钟，误差时间15分钟",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                },
                "ServiceUserAppKeyColumn": {
                    "name": "appkey",
                    "description": "appkey: 客户端使用特定的算法通过appid、时间戳和请求参数这三个参数生成的校验字符串，用于后端校验请求的合法性",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                }
            },
            "examples": {
                "Status500": {
                    "summary": "请求参数错误",
                    "value": {
                        "code": "ModelFactory.SmallModelRouter.SmallModelRun.ParameterError",
                        "description": "缺少参数 model_name",
                        "detail": "缺少参数 model_name",
                        "solution": "请检查填写的参数是否正确。",
                        "link": ""
                    }
                }
            }
        }
    }
    return api


def get_audio_restful_api_document(model_name):
    api = {
        "openapi": "3.0.2",
        "info": {
            "title": "audio小模型",
            "version": "1.0.0",
            "description": "audio小模型 RESTful API 文档"
        },
        "paths": {
            "/api/model-factory/v1/small_model_run": {
                "post": {
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/serviceReq"
                                },
                                "examples": {
                                    "request": {
                                        "value": {
                                            "model_name": model_name,
                                            "param_data": {
                                                "file_url": "https://test.com:443",
                                                "file_size": "方案运营策略会议.mp4",
                                                "file_name": "70337768"
                                            }
                                        }
                                    }
                                }
                            }
                        },
                        "required": True
                    },
                    "parameters": [
                        {
                            "$ref": "#/components/parameters/ServiceUserTokenColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserTimeStampColumn"
                        },
                        {
                            "$ref": "#/components/parameters/ServiceUserAppKeyColumn"
                        }
                    ],
                    "responses": {
                        "200": {
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/AudioResp"
                                    },
                                    "examples": {
                                        "resp": {
                                            "value": {
                                                "res": {
                                                    "id": "62d58c31bf034c1b96e26eaa3785dac9"
                                                }
                                            }
                                        }
                                    }
                                }
                            },
                            "description": "ok"
                        },
                        "500": {
                            "$ref": "#/components/responses/RespStaSE500"
                        }
                    },
                    "summary": "info_extract推理接口"
                }
            }
        },
        "components": {
            "schemas": {
                "serviceReq": {
                    "description": "使用服务请求体",
                    "type": "object",
                    "properties": {
                        "model_name": {
                            "description": "模型名称",
                            "type": "string"
                        },
                        "param_data": {
                            "description": "推理参数",
                            "type": "object",
                            "properties": {
                                "file_url": {
                                    "description": "音频文件下载地址",
                                    "type": "string"
                                },
                                "file_size": {
                                    "description": "音频文件大小",
                                    "type": "string"
                                },
                                "file_name": {
                                    "description": "音频文件名称",
                                    "type": "string"
                                }
                            }
                        }
                    }
                },
                "Error": {
                    "description": "接口调用错误信息基类，具体错误情况可查看字段detail",
                    "required": [
                        "code",
                        "solution",
                        "description",
                        "detail",
                        "link"
                    ],
                    "type": "object",
                    "properties": {
                        "description": {
                            "description": "导致此错误的原因",
                            "type": "string"
                        },
                        "code": {
                            "description": "业务错误码，'.' 字符前为微服务名， '.' 字符后为具体错误码",
                            "type": "string"
                        },
                        "solution": {
                            "description": "解决错误办法",
                            "type": "string"
                        },
                        "link": {
                            "description": "错误链接",
                            "type": "string"
                        },
                        "detail": {
                            "description": "错误详情",
                            "type": "array",
                            "items": {
                                "type": "object",
                                "properties": {
                                    "details": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    }
                },
                "AudioResp": {
                    "description": "info_extract响应结构体",
                    "type": "object",
                    "properties": {
                        "id": {
                            "type": "string"
                        }
                    }
                }
            },
            "responses": {
                "RespStaSE500": {
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/Error"
                            },
                            "examples": {
                                "Err400": {
                                    "$ref": "#/components/examples/Status500"
                                }
                            }
                        }
                    },
                    "description": "Bad Request"
                }
            },
            "parameters": {
                "ServiceUserTokenColumn": {
                    "name": "appid",
                    "description": "appid: 用户登录AnyDATA 通过 /manager/v1/appid 获取的AnyDATA账户的APPID，用于数据科学家和开发者调用或者使用AnyDATA服务时的唯一标识",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": True
                },
                "ServiceUserTimeStampColumn": {
                    "name": "timestamp",
                    "description": "timestamp: 此时间戳的接受范围以AD的系统时间为基准，接受客户端的时间在AD的系统有效期为30分钟，误差时间15分钟",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                },
                "ServiceUserAppKeyColumn": {
                    "name": "appkey",
                    "description": "appkey: 客户端使用特定的算法通过appid、时间戳和请求参数这三个参数生成的校验字符串，用于后端校验请求的合法性",
                    "schema": {
                        "type": "string"
                    },
                    "in": "header",
                    "required": False
                }
            },
            "examples": {
                "Status500": {
                    "summary": "请求参数错误",
                    "value": {
                        "code": "ModelFactory.SmallModelRouter.SmallModelRun.ParameterError",
                        "description": "缺少参数 model_name",
                        "detail": "缺少参数 model_name",
                        "solution": "请检查填写的参数是否正确。",
                        "link": ""
                    }
                }
            }
        }
    }
    return api
