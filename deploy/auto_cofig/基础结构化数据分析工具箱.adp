{
  "toolbox": {
    "configs": [
      {
        "box_id": "2e8f2ef8-3464-45c9-82b4-85ba27b787ce",
        "box_name": "基础结构化数据分析工具箱",
        "box_desc": "支持对结构话数据进行处理的工具箱，工具包含: \r\n1. json2plot\r\n2. execute_code_legacy\r\n3. create_file_legacy\r\n",
        "box_svc_url": "http://data-retrieval:9100",
        "status": "published",
        "category_type": "other_category",
        "category_name": "未分类",
        "is_internal": false,
        "source": "custom",
        "tools": [
          {
            "tool_id": "50c954b6-23d2-4753-a100-4a90131af8ce",
            "name": "json2plot",
            "description": "根据绘图参数生成用于前端展示的 JSON 对象。如果包含此工具，则优先使用此工具绘图\n\n调用方法是 `json2plot(title, chart_type, group_by, data_field, tool_result_cache_key)`\n\n**注意：**\n- 你拿到结果后, 不需要给用户展示这个 JSON 对象, 前端会自动画图",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "7363dbf9-8d75-4b32-8084-e406ce7c8479",
              "summary": "json2plot",
              "description": "根据绘图参数生成用于前端展示的 JSON 对象。如果包含此工具，则优先使用此工具绘图\n\n调用方法是 `json2plot(title, chart_type, group_by, data_field, tool_result_cache_key)`\n\n**注意：**\n- 你拿到结果后, 不需要给用户展示这个 JSON 对象, 前端会自动画图",
              "server_url": "http://data-retrieval:9100",
              "path": "/tools/json2plot",
              "method": "POST",
              "create_time": 1775201265557508400,
              "update_time": 1775201265557508400,
              "create_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
              "update_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
              "api_spec": {
                "parameters": [
                  {
                    "name": "stream",
                    "in": "query",
                    "description": "是否流式返回",
                    "required": false,
                    "schema": {
                      "default": false,
                      "type": "boolean"
                    }
                  },
                  {
                    "name": "mode",
                    "in": "query",
                    "description": "请求模式",
                    "required": false,
                    "schema": {
                      "default": "http",
                      "enum": [
                        "http",
                        "sse"
                      ],
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "schema": {
                        "example": {
                          "chart_type": "Line",
                          "data": [],
                          "data_field": "营收收入指标",
                          "group_by": [
                            "报告时间(按年)"
                          ],
                          "session_id": "123",
                          "session_type": "in_memory",
                          "title": "2024年1月1日到2024年1月3日，每天的销售额",
                          "tool_result_cache_key": ""
                        },
                        "properties": {
                          "chart_type": {
                            "description": "图表的类型, 输出仅支持三种: Pie, Line, Column, 环形图也属于 Pie",
                            "enum": [
                              "Pie",
                              "Line",
                              "Column"
                            ],
                            "type": "string"
                          },
                          "data": {
                            "description": "用于作图的 JSON 数据，与 tool_result_cache_key 参数不能同时设置。如果 tool_result_cache_key 为空，则使用此参数。数据格式为对象数组，每个对象表示一条数据记录",
                            "items": {
                              "additionalProperties": {
                                "type": [
                                  "string",
                                  "number"
                                ]
                              },
                              "type": "object"
                            },
                            "type": "array"
                          },
                          "data_field": {
                            "description": "数据字段，注意设置的 group_by 和 data_field 必须和数据匹配，不要自己生成，如果数据中没有，可以询问用户",
                            "type": "string"
                          },
                          "group_by": {
                            "description": "\n分组字段列表，支持多个字段，如果有时间字段，请放在第一位。另外:\n- 对于折线图, group_by 可能有1~2个值, 第一个是 x 轴, 第二个字段是 分组, data_field 是 y 轴\n- 对于柱状图, group_by 可能有1~3个值, 第一个是 x 轴, 第二个字段是 堆叠, 第三个字段是 分组, data_field 是 y 轴\n- 对于饼图, group_by 只有一个值, 即 colorField, data_field 是 angleField\n",
                            "items": {
                              "type": "string"
                            },
                            "type": "array"
                          },
                          "session_id": {
                            "description": "会话ID，用于标识和管理会话状态，同一会话ID可以共享历史数据和缓存",
                            "type": "string"
                          },
                          "session_type": {
                            "default": "redis",
                            "description": "会话类型，in_memory 表示内存存储（临时），redis 表示 Redis 存储（持久化）",
                            "enum": [
                              "in_memory",
                              "redis"
                            ],
                            "type": "string"
                          },
                          "timeout": {
                            "default": 30,
                            "description": "请求超时时间（秒），超过此时间未完成则返回超时错误，默认30秒",
                            "type": "number"
                          },
                          "title": {
                            "description": "和数据的 title 保持一致, 是一个字符串, **不是dict**",
                            "type": "string"
                          },
                          "tool_result_cache_key": {
                            "description": "text2metric 或 text2sql工具缓存 key, 其他工具的结果没有意义，key 是一个字符串, 与 data 不能同时设置",
                            "type": "string"
                          }
                        },
                        "required": [
                          "title",
                          "chart_type",
                          "group_by",
                          "data_field"
                        ],
                        "type": "object"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "Successful operation",
                    "content": {
                      "application/json": {
                        "example": {
                          "output": {
                            "chart_config": {
                              "chart_type": "Column",
                              "groupField": "",
                              "seriesField": "报告类型",
                              "xField": "报告时间(按年)",
                              "yField": "营收收入指标"
                            },
                            "config": {
                              "angleField": "",
                              "chart_type": "Column",
                              "colorField": "",
                              "xField": "报告时间(按年)",
                              "yField": "营收收入指标"
                            },
                            "data_sample": [
                              {
                                "报告时间(按年)": "2015",
                                "报告类型": "一季报",
                                "营收收入指标": 12312312
                              }
                            ],
                            "result_cache_key": "CACHE_KEY",
                            "title": "2024年1月1日到2024年1月3日，每天的销售额"
                          }
                        },
                        "schema": {
                          "properties": {
                            "chart_config": {
                              "description": "详细图表配置，包含完整的图表渲染参数，如 xField（X轴字段）、yField（Y轴字段）、seriesField（系列字段）、groupField（分组字段）等",
                              "type": "object"
                            },
                            "config": {
                              "description": "基础图表配置，包含图表类型、坐标轴字段等基础信息",
                              "type": "object"
                            },
                            "data_sample": {
                              "description": "数据样例，仅返回第一条数据用于预览，完整数据需通过 result_cache_key 从缓存获取",
                              "items": {
                                "type": "object"
                              },
                              "type": "array"
                            },
                            "result_cache_key": {
                              "description": "结果缓存键，用于从缓存中获取完整数据，前端可通过此键获取完整图表数据",
                              "type": "string"
                            },
                            "title": {
                              "description": "图表标题，用于前端展示",
                              "type": "string"
                            }
                          },
                          "type": "object"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {}
                },
                "callbacks": null,
                "security": null,
                "tags": null,
                "external_docs": null
              }
            },
            "use_rule": "",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1775201265580683500,
            "update_time": 1775201265580683500,
            "create_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
            "update_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
            "extend_info": {},
            "resource_object": "tool",
            "source_id": "7363dbf9-8d75-4b32-8084-e406ce7c8479",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "8a3907af-cfbb-4c38-a560-c4acc6db2264",
            "name": "create_file_legacy",
            "description": "在沙箱环境中创建新文件，支持文本内容或从缓存中获取内容",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "7feaf9ee-fa2f-4000-b239-c7535a3f7e67",
              "summary": "create_file_legacy",
              "description": "在沙箱环境中创建新文件，支持文本内容或从缓存中获取内容",
              "server_url": "http://data-retrieval:9100",
              "path": "/tools/create_file_legacy",
              "method": "POST",
              "create_time": 1775201265557508400,
              "update_time": 1775201265557508400,
              "create_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
              "update_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
              "api_spec": {
                "parameters": [
                  {
                    "name": "stream",
                    "in": "query",
                    "description": "是否流式返回",
                    "required": false,
                    "schema": {
                      "default": false,
                      "type": "boolean"
                    }
                  },
                  {
                    "name": "mode",
                    "in": "query",
                    "description": "请求模式",
                    "required": false,
                    "schema": {
                      "default": "http",
                      "enum": [
                        "http",
                        "sse"
                      ],
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "examples": {
                        "create_from_cache": {
                          "description": "使用缓存中的数据创建文件",
                          "summary": "从缓存创建文件",
                          "value": {
                            "filename": "data.json",
                            "result_cache_key": "cached_data_123",
                            "server_url": "http://localhost:8080",
                            "session_id": "test_session_123"
                          }
                        },
                        "create_python_file": {
                          "description": "创建包含 Python 代码的文件",
                          "summary": "创建 Python 文件",
                          "value": {
                            "content": "def fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)\n\n# 计算前10个斐波那契数\nfor i in range(10):\n    print(f'F({i}) = {fibonacci(i)}')",
                            "filename": "fibonacci.py",
                            "server_url": "http://localhost:8080",
                            "session_id": "test_session_123"
                          }
                        }
                      },
                      "schema": {
                        "properties": {
                          "content": {
                            "description": "文件内容, 如果 result_cache_key 参数不为空，则无需设置该参数",
                            "type": "string"
                          },
                          "filename": {
                            "description": "要创建的文件名",
                            "type": "string"
                          },
                          "result_cache_key": {
                            "description": "之前工具的结果缓存key，可以用于将结果写入到文件中，有此参数则无需设置 content 参数",
                            "type": "string"
                          },
                          "server_url": {
                            "default": "http://sandbox-control-plane:8000",
                            "description": "可选，沙箱服务器URL，默认使用配置文件中的 SANDBOX_URL",
                            "type": "string"
                          },
                          "session_id": {
                            "description": "沙箱会话ID",
                            "type": "string"
                          },
                          "session_type": {
                            "description": "会话类型, 可选值为: redis, in_memory, 默认值为 redis",
                            "type": "string"
                          },
                          "timeout": {
                            "default": 120,
                            "description": "超时时间",
                            "type": "number"
                          },
                          "title": {
                            "description": "对于当前操作的简单描述，便于用户理解",
                            "type": "string"
                          }
                        },
                        "required": [
                          "filename"
                        ],
                        "type": "object"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "Successful operation",
                    "content": {
                      "application/json": {
                        "schema": {
                          "properties": {
                            "message": {
                              "description": "操作状态消息",
                              "type": "string"
                            },
                            "result": {
                              "description": "操作结果, 包含标准输出、标准错误输出、返回值",
                              "type": "object"
                            }
                          },
                          "type": "object"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "400",
                    "description": "Bad request",
                    "content": {
                      "application/json": {
                        "schema": {
                          "properties": {
                            "detail": {
                              "description": "详细错误信息",
                              "type": "string"
                            },
                            "error": {
                              "description": "错误信息",
                              "type": "string"
                            }
                          },
                          "type": "object"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {}
                },
                "callbacks": null,
                "security": null,
                "tags": null,
                "external_docs": null
              }
            },
            "use_rule": "",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1775201265580683500,
            "update_time": 1775201265580683500,
            "create_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
            "update_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
            "extend_info": {},
            "resource_object": "tool",
            "source_id": "7feaf9ee-fa2f-4000-b239-c7535a3f7e67",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "f2cf3bfc-87a4-454d-9932-ab8c13a52e62",
            "name": "execute_code_legacy",
            "description": "在沙箱环境中执行 Python 代码，支持 pandas 等数据分析库，注意沙箱环境是受限环境，没有网络连接，不能使用 pip 安装第三方库。运行代码时，需要通过 print 输出结果，或者设置输出变量 output_params 参数，返回结果",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "9604a797-2614-406a-91d3-b8435777a764",
              "summary": "execute_code_legacy",
              "description": "在沙箱环境中执行 Python 代码，支持 pandas 等数据分析库，注意沙箱环境是受限环境，没有网络连接，不能使用 pip 安装第三方库。运行代码时，需要通过 print 输出结果，或者设置输出变量 output_params 参数，返回结果",
              "server_url": "http://data-retrieval:9100",
              "path": "/tools/execute_code_legacy",
              "method": "POST",
              "create_time": 1775201265557508400,
              "update_time": 1775201265557508400,
              "create_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
              "update_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
              "api_spec": {
                "parameters": [
                  {
                    "name": "stream",
                    "in": "query",
                    "description": "是否流式返回",
                    "required": false,
                    "schema": {
                      "default": false,
                      "type": "boolean"
                    }
                  },
                  {
                    "name": "mode",
                    "in": "query",
                    "description": "请求模式",
                    "required": false,
                    "schema": {
                      "default": "http",
                      "enum": [
                        "http",
                        "sse"
                      ],
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "examples": {
                        "basic_execution": {
                          "description": "执行简单的 Python 代码",
                          "summary": "基础代码执行",
                          "value": {
                            "content": "print('Hello World')\nx = 10\ny = 20\nresult = x + y\nprint(f'{x} + {y} = {result}')",
                            "filename": "hello.py",
                            "output_params": [
                              "result"
                            ],
                            "server_url": "http://localhost:8080",
                            "session_id": "test_session_123"
                          }
                        },
                        "data_analysis": {
                          "description": "使用 pandas 进行数据分析",
                          "summary": "数据分析示例",
                          "value": {
                            "content": "import pandas as pd\nimport numpy as np\n\n# 创建示例数据\ndata = {\n    'name': ['Alice', 'Bob', 'Charlie'],\n    'age': [25, 30, 35],\n    'salary': [50000, 60000, 70000]\n}\ndf = pd.DataFrame(data)\n\n# 计算统计信息\nstats = {\n    'mean_age': df['age'].mean(),\n    'mean_salary': df['salary'].mean(),\n    'total_records': len(df)\n}\n\nprint('数据统计:')\nfor key, value in stats.items():\n    print(f'{key}: {value}')\n\nresult = stats",
                            "filename": "data_analysis.py",
                            "output_params": [
                              "result",
                              "df"
                            ],
                            "server_url": "http://localhost:8080",
                            "session_id": "test_session_123"
                          }
                        }
                      },
                      "schema": {
                        "properties": {
                          "args": {
                            "description": "代码执行参数",
                            "items": {
                              "type": "string"
                            },
                            "type": "array"
                          },
                          "content": {
                            "description": "要执行的 Python 代码内容",
                            "type": "string"
                          },
                          "filename": {
                            "description": "文件名，用于指定代码文件的名称",
                            "type": "string"
                          },
                          "output_params": {
                            "description": "输出参数列表，用于指定要返回的变量名",
                            "items": {
                              "type": "string"
                            },
                            "type": "array"
                          },
                          "server_url": {
                            "default": "http://sandbox-control-plane:8000",
                            "description": "可选，沙箱服务器URL，默认使用配置文件中的 SANDBOX_URL",
                            "type": "string"
                          },
                          "session_id": {
                            "description": "沙箱会话ID",
                            "type": "string"
                          },
                          "timeout": {
                            "default": 120,
                            "description": "超时时间",
                            "type": "number"
                          },
                          "title": {
                            "description": "对于当前操作的简单描述，便于用户理解",
                            "type": "string"
                          }
                        },
                        "required": [
                          "content"
                        ],
                        "type": "object"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "400",
                    "description": "Bad request",
                    "content": {
                      "application/json": {
                        "schema": {
                          "properties": {
                            "detail": {
                              "description": "详细错误信息",
                              "type": "string"
                            },
                            "error": {
                              "description": "错误信息",
                              "type": "string"
                            }
                          },
                          "type": "object"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "200",
                    "description": "Successful operation",
                    "content": {
                      "application/json": {
                        "schema": {
                          "properties": {
                            "message": {
                              "description": "操作状态消息",
                              "type": "string"
                            },
                            "result": {
                              "description": "操作结果, 包含标准输出、标准错误输出、返回值",
                              "type": "object"
                            }
                          },
                          "type": "object"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {}
                },
                "callbacks": null,
                "security": null,
                "tags": null,
                "external_docs": null
              }
            },
            "use_rule": "",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1775201265580683500,
            "update_time": 1775201265580683500,
            "create_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
            "update_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
            "extend_info": {},
            "resource_object": "tool",
            "source_id": "9604a797-2614-406a-91d3-b8435777a764",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          }
        ],
        "create_time": 1775201265554708000,
        "update_time": 1775201670499196700,
        "create_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
        "update_user": "18dfa6e0-29d7-11f1-af4a-0e1a0427dbe7",
        "metadata_type": "openapi"
      }
    ]
  }
}