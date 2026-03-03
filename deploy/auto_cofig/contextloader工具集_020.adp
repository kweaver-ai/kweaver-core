{
  "toolbox": {
    "configs": [
      {
        "box_id": "408b5319-eefc-445e-9b15-7d332f0706ee",
        "box_name": "contextloader工具集_020",
        "box_desc": "context_loader工具集_020",
        "box_svc_url": "http://agent-retrieval:30779",
        "status": "published",
        "category_type": "other_category",
        "category_name": "未分类",
        "is_internal": false,
        "source": "custom",
        "tools": [
          {
            "tool_id": "cf6753b9-fba2-4c66-9cbe-27659087882b",
            "name": "get_logic_property_values",
            "description": "## 功能说明\r\n解析并查询知识网络中对象的逻辑属性值（包括 metric 指标和 operator 算子属性）。\r\n\r\n## 典型使用场景\r\n- 查询企业在指定时间段的药品上市数量（metric 指标）\r\n- 计算企业的健康度评分（operator 算子）\r\n- 批量查询多个对象的多个逻辑属性值\r\n",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "56270d68-d9d5-4995-852f-4e4d76aad80b",
              "summary": "逻辑属性解析接口",
              "description": "## 功能说明\n解析并查询知识网络中对象的逻辑属性值（包括 metric 指标和 operator 算子属性）。\n\n## 典型使用场景\n- 查询企业在指定时间段的药品上市数量（metric 指标）\n- 计算企业的健康度评分（operator 算子）\n- 批量查询多个对象的多个逻辑属性值\n",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/logic-property-resolver",
              "method": "POST",
              "create_time": 1770082202223080700,
              "update_time": 1770082202223080700,
              "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "用户ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "examples": {
                        "示例": {
                          "value": {
                            "additional_context": "用途：给 LLM/Agent 补充生成 dynamic_params 所需的上下文（自由文本或 JSON 均可）。\n建议至少包含：\n1) 对象实例信息：company_id=company_000001，registered_capital=2000000\n2) 约束/筛选：registered_capital > 1000000\n3) 时间信息：now_ms=1762996342241，timezone=Asia/Shanghai；若为趋势查询，建议注明 step=month；若为即时查询，注明 instant=true\n",
                            "kn_id": "kn_medical",
                            "options": {
                              "max_repair_rounds": 1,
                              "return_debug": false
                            },
                            "ot_id": "company",
                            "properties": [
                              "approved_drug_count",
                              "business_health_score"
                            ],
                            "query": "注册资金大于100万人民币的企业有哪些？想知道这些药企的药品上市情况，并且这些企业的健康度怎么样？",
                            "unique_identities": [
                              {
                                "company_id": "company_000001"
                              }
                            ]
                          }
                        }
                      },
                      "schema": {
                        "$ref": "#/components/schemas/ResolveLogicPropertiesRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "ok",
                    "content": {
                      "application/json": {
                        "examples": {
                          "成功示例": {
                            "value": {
                              "datas": [
                                {
                                  "approved_drug_count": {
                                    "datas": [
                                      {
                                        "labels": {},
                                        "times": [
                                          1762996342241
                                        ],
                                        "values": [
                                          12
                                        ]
                                      }
                                    ],
                                    "model": {
                                      "unit": "times",
                                      "unit_type": "countUnit"
                                    }
                                  },
                                  "business_health_score": {
                                    "level": "B",
                                    "score": 0.82
                                  },
                                  "company_id": "company_000001"
                                }
                              ]
                            }
                          },
                          "缺参示例": {
                            "value": {
                              "error_code": "MISSING_INPUT_PARAMS",
                              "message": "dynamic_params 缺少必需的 input 参数",
                              "missing": [
                                {
                                  "params": [
                                    {
                                      "hint": "在 additional_context 中补充明确时间范围（start/end 或 “最近半年”），或在 query 中明确时间范围",
                                      "name": "start",
                                      "type": "INTEGER"
                                    }
                                  ],
                                  "property": "approved_drug_count"
                                }
                              ],
                              "trace_id": "3f5d6c1c-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                            }
                          }
                        },
                        "schema": {
                          "$ref": "#/components/schemas/ResolveLogicPropertiesResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "400",
                    "description": "bad request",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/Error"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "500",
                    "description": "internal error",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/Error"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "ObjectPropertiesValuesResponse": {
                      "required": [
                        "datas"
                      ],
                      "properties": {
                        "datas": {
                          "type": "array",
                          "description": "对象实例数据数组（与 unique_identities 对齐，字段动态包含请求的 properties）",
                          "items": {
                            "$ref": "#/components/schemas/ObjectPropertyValue"
                          }
                        },
                        "debug": {
                          "$ref": "#/components/schemas/ResolveDebugInfo"
                        }
                      },
                      "type": "object",
                      "description": "成功响应，包含查询结果和可选的调试信息"
                    },
                    "ResolveOptions": {
                      "type": "object",
                      "description": "【可选配置】控制接口行为的高级选项\n",
                      "properties": {
                        "max_concurrency": {
                          "description": "当查询多个 properties 时，并发调用 Agent 生成参数的最大并发数。\n\n**说明**:\n- 较小值（1-3）：降低服务压力，适合 Agent 响应较慢的情况\n- 推荐值（4-6）：平衡性能和资源消耗\n- 较大值（7-10）：加快处理速度，适合对响应时间敏感的场景\n\n**注意**: 过大的并发数可能导致 Agent 服务过载\n",
                          "default": 4,
                          "type": "integer",
                          "format": "int32"
                        },
                        "max_repair_rounds": {
                          "default": 1,
                          "type": "integer",
                          "format": "int32",
                          "description": "当 Agent 返回的 JSON 格式不合法时，最大尝试修复的轮次。\n\n**说明**:\n- 0 = 不进行修复，直接报错\n- 1 = 尝试修复一次（推荐，默认值）\n- 2+ = 多次修复尝试（可能增加延迟）\n\n**适用场景**: Agent 输出质量不稳定时可适当增加\n"
                        },
                        "return_debug": {
                          "description": "是否返回调试信息。设置为 true 时，响应中会包含 debug 字段，内容包括：\n- dynamic_params: Agent 生成的所有动态参数\n- now_ms: 服务器时间戳\n- warnings: 警告信息（如有）\n\n**用途**: 调试参数生成逻辑、验证 Agent 行为\n",
                          "default": false,
                          "type": "boolean"
                        }
                      }
                    },
                    "UniqueIdentity": {
                      "type": "object",
                      "description": "对象唯一标识（主键字段 map）"
                    },
                    "MissingParamsError": {
                      "properties": {
                        "message": {
                          "description": "错误信息",
                          "type": "string"
                        },
                        "missing": {
                          "description": "缺参清单（按 property 分组）",
                          "items": {
                            "$ref": "#/components/schemas/MissingPropertyParams"
                          },
                          "type": "array"
                        },
                        "trace_id": {
                          "type": "string",
                          "description": "请求追踪ID"
                        },
                        "error_code": {
                          "type": "string",
                          "description": "错误码，固定为 MISSING_INPUT_PARAMS"
                        }
                      },
                      "type": "object",
                      "description": "缺参错误响应。根据 hint 在 additional_context 或 query 中补充信息后重试",
                      "required": [
                        "error_code",
                        "message",
                        "missing",
                        "trace_id"
                      ]
                    },
                    "ResolveLogicPropertiesRequest": {
                      "type": "object",
                      "required": [
                        "kn_id",
                        "ot_id",
                        "query",
                        "unique_identities",
                        "properties"
                      ],
                      "properties": {
                        "additional_context": {
                          "description": "【可选】补充上下文信息，帮助 Agent 更准确地生成 dynamic_params。\n\n**何时需要提供**:\n- 需要传递对象实例的关键属性值\n- 需要指定特殊的查询模式（如即时查询 vs 趋势查询）\n\n**推荐格式**: 结构化的 JSON 字符串（也支持自由文本）\n\n**建议包含的信息**:\n1. 时间信息：timezone（时区）\n2. 对象属性：对象实例的关键属性值（如 company_name、registered_capital）\n3. 查询模式：instant（是否即时查询）、step（趋势查询的步长：day/week/month/quarter/year）\n4. 筛选条件：业务约束条件（如 registered_capital > 1000000）\n\n**示例**:\n```json\n{\n  \"timezone\": \"Asia/Shanghai\",\n  \"instant\": true,\n  \"object_context\": {\n    \"company_id\": \"company_000001\",\n    \"company_name\": \"广西金秀圣堂药业有限责任公司\"\n  }\n}\n```\n\n**注意**: 避免传递大段无关文本，保持信息精准和结构化。\n",
                          "type": "string"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "业务知识网络ID。例如：kn_medical（医疗领域）、kn_finance（金融领域）\n"
                        },
                        "options": {
                          "$ref": "#/components/schemas/ResolveOptions"
                        },
                        "ot_id": {
                          "type": "string",
                          "description": "对象类ID。指定要查询的对象类型。例如：company（企业）、drug（药品）、person（人物）\n"
                        },
                        "properties": {
                          "description": "需要查询的逻辑属性名称列表。支持两种类型：\n- **metric**: 指标属性（如药品上市数量、销售额趋势等）\n- **operator**: 算子属性（如健康度评分、风险等级等）\n\n本接口会自动为每个属性生成 dynamic_params 并查询其值。\n",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        },
                        "query": {
                          "type": "string",
                          "description": "用户原始查询问题。用于 Agent 理解查询意图，生成逻辑属性所需的动态参数。\n\n**重要**: query 的内容会直接影响参数生成质量，应包含：\n- 时间信息（如\"最近一年\"、\"2023年\"）\n- 统计维度（如\"总数\"、\"趋势\"、\"分布\"）\n- 业务上下文（帮助 Agent 理解查询目的）\n"
                        },
                        "unique_identities": {
                          "description": "要查询的对象实例主键数组（支持批量查询）。\n\n每个元素是一个键值对对象，包含该对象类的主键字段。\n例如：[{\"company_id\": \"company_000001\"}, {\"company_id\": \"company_000002\"}]\n",
                          "items": {
                            "$ref": "#/components/schemas/UniqueIdentity"
                          },
                          "type": "array"
                        }
                      }
                    },
                    "ObjectPropertyValue": {
                      "type": "object",
                      "description": "单个对象实例的属性值（字段动态，包含主键和请求的 properties）"
                    },
                    "ResolveDebugInfo": {
                      "properties": {
                        "now_ms": {
                          "type": "integer",
                          "format": "int64",
                          "description": "服务器时间戳（毫秒）"
                        },
                        "trace_id": {
                          "type": "string",
                          "description": "请求追踪ID"
                        },
                        "warnings": {
                          "type": "array",
                          "description": "警告信息",
                          "items": {
                            "type": "string"
                          }
                        },
                        "dynamic_params": {
                          "type": "object",
                          "description": "Agent 生成的动态参数"
                        }
                      },
                      "type": "object",
                      "description": "调试信息（仅当 return_debug=true 时返回）"
                    },
                    "ResolveLogicPropertiesResponse": {
                      "description": "成功返回 datas 数组；缺参时返回 error_code 和 missing 清单",
                      "oneOf": [
                        {
                          "$ref": "#/components/schemas/ObjectPropertiesValuesResponse"
                        },
                        {
                          "$ref": "#/components/schemas/MissingParamsError"
                        }
                      ]
                    },
                    "Error": {
                      "type": "object",
                      "properties": {
                        "error_code": {
                          "type": "string"
                        },
                        "message": {
                          "type": "string"
                        }
                      }
                    },
                    "MissingPropertyParams": {
                      "type": "object",
                      "description": "单个属性的缺参信息",
                      "required": [
                        "property",
                        "params"
                      ],
                      "properties": {
                        "params": {
                          "type": "array",
                          "description": "该属性缺失的参数列表",
                          "items": {
                            "$ref": "#/components/schemas/MissingParam"
                          }
                        },
                        "property": {
                          "type": "string",
                          "description": "逻辑属性名称"
                        }
                      }
                    },
                    "MissingParam": {
                      "type": "object",
                      "description": "缺失参数的详情（名称、类型、补充建议）",
                      "required": [
                        "name",
                        "type"
                      ],
                      "properties": {
                        "type": {
                          "type": "string",
                          "description": "参数类型（INTEGER/BOOLEAN/STRING/array/object）"
                        },
                        "hint": {
                          "type": "string",
                          "description": "补参建议（如何在 additional_context 或 query 中补充）"
                        },
                        "name": {
                          "type": "string",
                          "description": "参数名（如 start, end, instant, step）"
                        }
                      }
                    }
                  }
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
            "create_time": 1770082202229131000,
            "update_time": 1770082202229131000,
            "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "extend_info": {},
            "resource_object": "tool",
            "source_id": "56270d68-d9d5-4995-852f-4e4d76aad80b",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "fcfaee61-7055-4847-bb45-60e5e22b02b0",
            "name": "kn_schema_search",
            "description": "该接口专注于业务知识网络的Schema信息语义检索，通过接收用户的自然语言查询，在指定的业务知识网络中召回与查询最相关的概念定义信息。主要用于获取业务知识网络中的对象类定义（包含属性、主键、操作符等）、关系类定义（包含源对象类、目标对象类等关联信息）以及动作类定义。支持关键词+向量混合检索模式，可限定搜索范围（对象类、关系类、动作类），并提供查询理解结果与概念相关性评分，帮助智能体或应用系统高效获取业务知识网络的结构化Schema定义数据。\r\n",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "5bc066eb-3182-4cae-a681-74c88cb440a1",
              "summary": "kn_schema_search",
              "description": "基于用户查询意图，返回业务知识网络中相关的概念信息",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/semantic-search",
              "method": "POST",
              "create_time": 1770082202223080700,
              "update_time": 1770082202223080700,
              "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "用户ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "schema": {
                        "$ref": "#/components/schemas/SemanticSearchRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "成功返回相关概念信息",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/SemanticSearchResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "400",
                    "description": "参数错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "500",
                    "description": "服务器内部错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "QueryStrategy": {
                      "type": "object",
                      "properties": {
                        "strategy_type": {
                          "type": "string",
                          "description": "策略类型",
                          "enum": [
                            "concept_get",
                            "concept_discovery",
                            "object_instance_discovery"
                          ]
                        },
                        "filter": {
                          "$ref": "#/components/schemas/ConceptFilter"
                        }
                      }
                    },
                    "QueryIntent": {
                      "type": "object",
                      "properties": {
                        "confidence": {
                          "description": "置信度",
                          "type": "number",
                          "format": "float"
                        },
                        "query_segment": {
                          "description": "对应的查询片段",
                          "type": "string"
                        },
                        "reasoning": {
                          "type": "string",
                          "description": "简要识别推理过程"
                        },
                        "related_concepts": {
                          "items": {
                            "$ref": "#/components/schemas/RelatedConcept"
                          },
                          "type": "array"
                        },
                        "requires_reasoning": {
                          "type": "boolean",
                          "description": "是否需要进一步推理",
                          "default": false
                        }
                      }
                    },
                    "RelatedConcept": {
                      "type": "object",
                      "properties": {
                        "concept_id": {
                          "type": "string",
                          "description": "概念类ID"
                        },
                        "concept_name": {
                          "type": "string",
                          "description": "概念类名称"
                        },
                        "concept_type": {
                          "description": "概念类型",
                          "enum": [
                            "object_type",
                            "relation_type",
                            "action_type"
                          ],
                          "type": "string"
                        }
                      }
                    },
                    "DataProperty": {
                      "type": "object",
                      "description": "数据属性结构定义",
                      "properties": {
                        "condition_operations": {
                          "type": "array",
                          "description": "该数据属性支持的查询条件操作符列表。\n",
                          "items": {
                            "type": "string",
                            "enum": [
                              "==",
                              "!=",
                              ">",
                              "<",
                              ">=",
                              "<=",
                              "in",
                              "not_in",
                              "like",
                              "not_like",
                              "range",
                              "out_range",
                              "exist",
                              "not_exist",
                              "regex",
                              "match",
                              "knn"
                            ]
                          }
                        },
                        "display_name": {
                          "type": "string",
                          "description": "属性显示名称"
                        },
                        "mapped_field": {
                          "description": "视图字段信息"
                        },
                        "name": {
                          "description": "属性名称",
                          "type": "string"
                        },
                        "type": {
                          "type": "string",
                          "description": "属性数据类型"
                        },
                        "comment": {
                          "type": "string",
                          "description": "备注"
                        }
                      }
                    },
                    "Concept": {
                      "properties": {
                        "rerank_score": {
                          "format": "float",
                          "description": "重排序得分",
                          "type": "number"
                        },
                        "samples": {
                          "description": "实例样本列表",
                          "items": {
                            "type": "object"
                          },
                          "type": "array"
                        },
                        "concept_detail": {
                          "oneOf": [
                            {
                              "$ref": "#/components/schemas/ObjectTypeDetail"
                            },
                            {
                              "$ref": "#/components/schemas/RelationTypeDetail"
                            },
                            {
                              "$ref": "#/components/schemas/ActionTypeDetail"
                            }
                          ],
                          "description": "概念类详情，根据concept_type返回不同结构：\n- 当concept_type为\"object_type\"时，返回ObjectTypeDetail结构，包含对象类的完整信息\n- 当concept_type为\"relation_type\"时，返回RelationTypeDetail结构，包含关系类的完整信息\n- 当concept_type为\"action_type\"时，返回ActionTypeDetail结构，包含行动类的完整信息\n"
                        },
                        "concept_id": {
                          "type": "string",
                          "description": "概念类ID"
                        },
                        "concept_name": {
                          "type": "string",
                          "description": "概念类名称"
                        },
                        "concept_type": {
                          "enum": [
                            "object_type",
                            "relation_type",
                            "action_type"
                          ],
                          "type": "string",
                          "description": "概念类型"
                        },
                        "intent_score": {
                          "format": "float",
                          "description": "意图得分",
                          "type": "number"
                        },
                        "match_score": {
                          "type": "number",
                          "format": "float",
                          "description": "匹配得分"
                        }
                      },
                      "type": "object"
                    },
                    "ObjectTypeDetail": {
                      "type": "object",
                      "description": "对象类概念详情",
                      "properties": {
                        "data_source": {
                          "$ref": "#/components/schemas/ResourceInfo"
                        },
                        "comment": {
                          "type": "string",
                          "description": "备注"
                        },
                        "logic_properties": {
                          "type": "array",
                          "description": "逻辑属性",
                          "items": {
                            "type": "object"
                          }
                        },
                        "tags": {
                          "items": {
                            "type": "string"
                          },
                          "type": "array",
                          "description": "标签"
                        },
                        "module_type": {
                          "type": "string",
                          "description": "模块类型"
                        },
                        "primary_keys": {
                          "items": {
                            "type": "string"
                          },
                          "type": "array",
                          "description": "主键字段"
                        },
                        "data_properties": {
                          "type": "array",
                          "description": "数据属性",
                          "items": {
                            "$ref": "#/components/schemas/DataProperty"
                          }
                        },
                        "name": {
                          "type": "string",
                          "description": "对象名称"
                        },
                        "id": {
                          "type": "string",
                          "description": "对象id"
                        },
                        "_score": {
                          "format": "float",
                          "description": "分数",
                          "type": "number"
                        }
                      }
                    },
                    "ActionTypeDetail": {
                      "type": "object",
                      "description": "行动类概念详情",
                      "properties": {
                        "comment": {
                          "description": "备注",
                          "type": "string"
                        },
                        "id": {
                          "type": "string",
                          "description": "行动类ID"
                        },
                        "module_type": {
                          "type": "string",
                          "description": "模块类型"
                        },
                        "name": {
                          "type": "string",
                          "description": "行动类名称"
                        },
                        "object_type_id": {
                          "description": "行动类所绑定的对象类ID",
                          "type": "string"
                        },
                        "tags": {
                          "type": "array",
                          "description": "标签",
                          "items": {
                            "type": "string"
                          }
                        },
                        "_score": {
                          "format": "float",
                          "description": "分数",
                          "type": "number"
                        }
                      }
                    },
                    "QueryUnderstanding": {
                      "properties": {
                        "origin_query": {
                          "description": "用户原始查询",
                          "type": "string"
                        },
                        "processed_query": {
                          "type": "string",
                          "description": "LLM处理后的标准化查询"
                        },
                        "query_strategy": {
                          "type": "array",
                          "items": {
                            "$ref": "#/components/schemas/QueryStrategy"
                          }
                        },
                        "intent": {
                          "type": "array",
                          "items": {
                            "$ref": "#/components/schemas/QueryIntent"
                          }
                        }
                      },
                      "type": "object"
                    },
                    "Condition": {
                      "properties": {
                        "operation": {
                          "type": "string",
                          "description": "操作类型"
                        },
                        "value": {
                          "type": "string",
                          "description": "值"
                        },
                        "field": {
                          "type": "string",
                          "description": "字段名称"
                        }
                      },
                      "type": "object"
                    },
                    "ConceptFilter": {
                      "type": "object",
                      "properties": {
                        "conditions": {
                          "type": "array",
                          "items": {
                            "$ref": "#/components/schemas/Condition"
                          }
                        },
                        "concept_type": {
                          "type": "string",
                          "description": "概念类型",
                          "enum": [
                            "object_type",
                            "relation_type",
                            "action_type"
                          ]
                        }
                      }
                    },
                    "RelationTypeDetail": {
                      "properties": {
                        "tags": {
                          "description": "标签",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        },
                        "type": {
                          "type": "string",
                          "description": "关系类型"
                        },
                        "target_object_type_id": {
                          "description": "目标对象类ID",
                          "type": "string"
                        },
                        "name": {
                          "type": "string",
                          "description": "关系类名称"
                        },
                        "comment": {
                          "type": "string",
                          "description": "备注"
                        },
                        "source_object_type_id": {
                          "description": "起点对象类ID",
                          "type": "string"
                        },
                        "_score": {
                          "description": "分数",
                          "type": "number",
                          "format": "float"
                        },
                        "id": {
                          "type": "string",
                          "description": "关系类id"
                        },
                        "module_type": {
                          "type": "string",
                          "description": "模块类型"
                        }
                      },
                      "type": "object",
                      "description": "关系类概念详情"
                    },
                    "SemanticSearchResponse": {
                      "properties": {
                        "query_understanding": {
                          "$ref": "#/components/schemas/QueryUnderstanding"
                        },
                        "concepts": {
                          "type": "array",
                          "items": {
                            "$ref": "#/components/schemas/Concept"
                          }
                        }
                      },
                      "type": "object"
                    },
                    "ResourceInfo": {
                      "properties": {
                        "name": {
                          "description": "视图名称",
                          "type": "string"
                        },
                        "type": {
                          "type": "string",
                          "description": "数据来源类型"
                        },
                        "id": {
                          "description": "数据视图id",
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "description": "数据来源信息"
                    },
                    "SearchScope": {
                      "type": "object",
                      "properties": {
                        "include_relation_types": {
                          "type": "boolean",
                          "description": "是否包含关系类"
                        },
                        "concept_groups": {
                          "type": "array",
                          "description": "限定的概念分组",
                          "items": {
                            "type": "string"
                          }
                        },
                        "include_action_types": {
                          "description": "是否包含行作类",
                          "type": "boolean"
                        },
                        "include_object_types": {
                          "type": "boolean",
                          "description": "是否包含对象类"
                        }
                      }
                    },
                    "SemanticSearchRequest": {
                      "type": "object",
                      "required": [
                        "query",
                        "kn_id"
                      ],
                      "properties": {
                        "previous_queries": {
                          "description": "之前的查询历史",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        },
                        "query": {
                          "description": "用户自然语言查询",
                          "type": "string"
                        },
                        "rerank_action": {
                          "enum": [
                            "default",
                            "vector",
                            "llm"
                          ],
                          "type": "string",
                          "description": "重排动作",
                          "default": "default"
                        },
                        "return_query_understanding": {
                          "default": false,
                          "type": "boolean",
                          "description": "是否返回查询理解信息"
                        },
                        "search_scope": {
                          "$ref": "#/components/schemas/SearchScope"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "业务知识网络ID"
                        },
                        "max_concepts": {
                          "description": "最大返回概念数量",
                          "default": 10,
                          "type": "integer"
                        }
                      }
                    },
                    "ErrorResponse": {
                      "type": "object",
                      "properties": {
                        "solution": {
                          "type": "string",
                          "description": "解决方案"
                        },
                        "code": {
                          "type": "string",
                          "description": "错误码"
                        },
                        "description": {
                          "type": "string",
                          "description": "错误描述"
                        },
                        "detail": {
                          "type": "object",
                          "description": "错误详情"
                        },
                        "link": {
                          "type": "string",
                          "description": "错误链接"
                        }
                      }
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "SemanticSearch"
                ],
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
            "create_time": 1770082202229131000,
            "update_time": 1770082202229131000,
            "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "extend_info": {},
            "resource_object": "tool",
            "source_id": "5bc066eb-3182-4cae-a681-74c88cb440a1",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "636c77ee-70ad-46a3-b68b-78b0ed2eb7c8",
            "name": "get_action_info",
            "description": "行动信息召回，根据对象的唯一标识，召回与该对象关联的行动信息。用于实现工具的动态调用。\r\n",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "7eae029a-85a7-4f9a-af1c-a323beeb8c81",
              "summary": "行动信息召回接口",
              "description": "根据对象的唯一标识，召回与该对象关联的行动信息。\n\n**重要说明：**\n1. 只接收 unique_identity（单个对象标识）\n",
              "server_url": "http://agent-retrieval-svc:8000",
              "path": "/api/agent-retrieval/in/v1/kn/get_action_info",
              "method": "POST",
              "create_time": 1770082202223080700,
              "update_time": 1770082202223080700,
              "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户 ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    },
                    "example": "bdb78b62-6c48-11f0-af96-fa8dcc0a06b2"
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型",
                    "required": true,
                    "schema": {
                      "enum": [
                        "user",
                        "system"
                      ],
                      "type": "string"
                    },
                    "example": "user"
                  }
                ],
                "request_body": {
                  "description": "对象的唯一标识",
                  "content": {
                    "application/json": {
                      "examples": {
                        "disease_example": {
                          "summary": "疾病对象示例",
                          "value": {
                            "at_id": "generate_treatment_plan",
                            "kn_id": "kn_medical",
                            "unique_identity": {
                              "disease_id": "disease_000001"
                            }
                          }
                        }
                      },
                      "schema": {
                        "$ref": "#/components/schemas/ActionRecallRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "400",
                    "description": "请求参数错误",
                    "content": {
                      "application/json": {
                        "examples": {
                          "invalid_request": {
                            "summary": "参数格式错误",
                            "value": {
                              "code": "INVALID_REQUEST",
                              "description": "unique_identity 格式错误",
                              "detail": {
                                "error": "unique_identity must be an object with at least one property"
                              },
                              "solution": "请确保 unique_identity 是一个包含至少一个属性的 JSON 对象"
                            }
                          },
                          "unsupported_action_type": {
                            "summary": "不支持的行动源类型",
                            "value": {
                              "code": "UNSUPPORTED_ACTION_TYPE",
                              "description": "当前仅支持 type=tool 的行动源，MCP 类型将在下个版本支持",
                              "detail": {
                                "action_source": {
                                  "box_id": "xxx-xxx-xxx",
                                  "tool_id": "yyy-yyy-yyy",
                                  "type": "mcp"
                                }
                              },
                              "solution": "请使用 type=tool 的行动源，或等待下个版本的 MCP 支持"
                            }
                          }
                        },
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "500",
                    "description": "服务器内部错误",
                    "content": {
                      "application/json": {
                        "examples": {
                          "schema_conversion_error": {
                            "summary": "Schema 转换失败",
                            "value": {
                              "code": "SCHEMA_CONVERSION_ERROR",
                              "description": "无法转换工具的 OpenAPI Schema 为 Function Call Schema",
                              "detail": {
                                "error": "Failed to resolve $ref",
                                "tool_id": "xxx-xxx-xxx"
                              },
                              "solution": "请检查工具定义的 OpenAPI Schema 是否符合规范，特别是 $ref 引用是否正确"
                            }
                          }
                        },
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "502",
                    "description": "上游服务不可用",
                    "content": {
                      "application/json": {
                        "examples": {
                          "ontology_query_error": {
                            "summary": "行动查询服务不可用",
                            "value": {
                              "code": "SERVICE_UNAVAILABLE",
                              "description": "行动查询接口调用失败",
                              "detail": {
                                "error": "Connection timeout",
                                "service": "ontology-query"
                              },
                              "solution": "请检查 ontology-query 服务是否正常运行，或稍后重试"
                            }
                          },
                          "toolbox_error": {
                            "summary": "工具箱服务不可用",
                            "value": {
                              "code": "SERVICE_UNAVAILABLE",
                              "description": "工具详情接口调用失败",
                              "detail": {
                                "error": "404 Not Found",
                                "service": "agent-operator-integration"
                              },
                              "solution": "请检查 agent-operator-integration 服务是否正常运行，以及工具是否存在"
                            }
                          }
                        },
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "200",
                    "description": "成功返回动态工具列表",
                    "content": {
                      "application/json": {
                        "examples": {
                          "empty_result": {
                            "summary": "无可用行动（空结果）",
                            "value": {
                              "_dynamic_tools": []
                            }
                          },
                          "success_example": {
                            "summary": "成功返回示例",
                            "value": {
                              "_dynamic_tools": [
                                {
                                  "api_url": "http://agent-operator-integration:9000/api/agent-operator-integration/internal-v1/tool-box/e59f35f8-65f3-46bb-b5f1-a428380b3edc/proxy/7acb8add-ecd0-45cf-bd4d-57674c3c69b0",
                                  "description": "该接口基于业务知识网络语义检索接口返回的对象类定义，查询具体的对象实例数据。",
                                  "fixed_params": {
                                    "body": {},
                                    "header": {
                                      "X-HTTP-Method-Override": "GET",
                                      "x-account-id": "test",
                                      "x-account-type": "user"
                                    },
                                    "path": {},
                                    "query": {}
                                  },
                                  "name": "根据单个对象类查询对象实例",
                                  "original_schema": {
                                    "components": {},
                                    "parameters": [],
                                    "request_body": {}
                                  },
                                  "parameters": {
                                    "properties": {
                                      "condition": {
                                        "description": "过滤条件",
                                        "type": "object"
                                      },
                                      "kn_id": {
                                        "description": "业务知识网络ID",
                                        "type": "string"
                                      },
                                      "limit": {
                                        "description": "返回的数量，默认值 10。范围 1-10000",
                                        "type": "integer"
                                      },
                                      "ot_id": {
                                        "description": "对象类ID",
                                        "type": "string"
                                      }
                                    },
                                    "required": [
                                      "kn_id",
                                      "ot_id"
                                    ],
                                    "type": "object"
                                  }
                                }
                              ],
                              "headers": {
                                "x-account-id": "test",
                                "x-account-type": "user"
                              }
                            }
                          }
                        },
                        "schema": {
                          "$ref": "#/components/schemas/ActionRecallResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "ActionRecallRequest": {
                      "properties": {
                        "at_id": {
                          "type": "string",
                          "description": "行动类型ID"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "业务知识网络ID"
                        },
                        "unique_identity": {
                          "type": "object",
                          "description": "对象的唯一标识，键值对形式。\n- key: 主键属性名\n- value: 属性值（支持 string, number, boolean 类型）\n\n**注意：** 当前版本仅支持单个对象标识（unique_identity），不支持多个对象（unique_identities）\n"
                        }
                      },
                      "type": "object",
                      "required": [
                        "kn_id",
                        "at_id",
                        "unique_identity"
                      ]
                    },
                    "ErrorResponse": {
                      "type": "object",
                      "properties": {
                        "detail": {
                          "type": "object",
                          "description": "错误详情"
                        },
                        "link": {
                          "type": "string",
                          "description": "错误链接"
                        },
                        "solution": {
                          "type": "string",
                          "description": "解决方案"
                        },
                        "code": {
                          "type": "string",
                          "description": "错误码"
                        },
                        "description": {
                          "description": "错误描述",
                          "type": "string"
                        }
                      }
                    },
                    "ActionRecallResponse": {
                      "properties": {
                        "_dynamic_tools": {
                          "type": "array",
                          "description": "动态工具列表，每个元素代表一个可用的行动工具",
                          "items": {
                            "$ref": "#/components/schemas/DynamicTool"
                          }
                        },
                        "headers": {
                          "type": "object",
                          "description": "HTTP Header 参数"
                        }
                      },
                      "type": "object",
                      "required": [
                        "_dynamic_tools"
                      ]
                    },
                    "DynamicTool": {
                      "properties": {
                        "fixed_params": {
                          "$ref": "#/components/schemas/FixedParams"
                        },
                        "name": {
                          "type": "string",
                          "description": "工具名称，来自工具详情的 name 字段"
                        },
                        "original_schema": {
                          "description": "工具的原始 OpenAPI 定义（api_spec），保留完整的工具元数据。\n\n**用途：**\n- 供调用方判断参数位置（path/query/header/body）\n- 供调用方组装实际的 HTTP 请求\n- 提供完整的 schema 定义（包括 $ref 引用）\n",
                          "properties": {
                            "responses": {
                              "type": "array",
                              "description": "响应定义",
                              "items": {
                                "type": "object"
                              }
                            },
                            "components": {
                              "type": "object",
                              "description": "组件定义（包括 schemas）"
                            },
                            "parameters": {
                              "description": "API 参数定义",
                              "items": {
                                "type": "object"
                              },
                              "type": "array"
                            },
                            "request_body": {
                              "type": "object",
                              "description": "请求体定义"
                            }
                          },
                          "type": "object"
                        },
                        "parameters": {
                          "properties": {
                            "type": {
                              "enum": [
                                "object"
                              ],
                              "type": "string"
                            },
                            "properties": {
                              "type": "object",
                              "description": "参数定义，key 为参数名，value 为参数 schema"
                            },
                            "required": {
                              "description": "必填参数列表",
                              "items": {
                                "type": "string"
                              },
                              "type": "array"
                            }
                          },
                          "type": "object",
                          "description": "符合 OpenAI Function Call 规范的参数 Schema。\n\n**转换规则：**\n- 扁平化所有 path/query/header/body 参数到一个 properties 对象\n- 不保留参数位置信息（in: path/query/header/body）\n- 合并所有位置的 required 参数到顶层 required 数组\n- 基本的 $ref 解析（MVP 版本不处理复杂嵌套）\n"
                        },
                        "api_call_strategy": {
                          "type": "string",
                          "description": "结果处理策略， 固定值为 kn_action_recall"
                        },
                        "api_url": {
                          "format": "uri",
                          "description": "工具执行的请求URL，用于实际调用工具。\n",
                          "type": "string"
                        },
                        "description": {
                          "type": "string",
                          "description": "工具描述，来自工具详情的 description 字段"
                        }
                      },
                      "type": "object",
                      "required": [
                        "name",
                        "description",
                        "parameters",
                        "api_url",
                        "original_schema",
                        "fixed_params"
                      ]
                    },
                    "FixedParams": {
                      "type": "object",
                      "description": "已实例化的固定参数，来自行动查询的 actions[0].parameters。\n\n**重要说明：**\n- 这些参数已经根据对象实例进行了实例化\n- 调用方需要将这些参数与 LLM 生成的动态参数合并\n- 参数位置（header/path/query/body）的判断可以参考 original_schema\n\n**参数来源：**\n- 行动查询接口返回的 actions[0].parameters\n- 不包括 dynamic_params（由 LLM 动态生成）\n",
                      "required": [
                        "header",
                        "path",
                        "query",
                        "body"
                      ],
                      "properties": {
                        "header": {
                          "type": "object",
                          "description": "HTTP Header 参数"
                        },
                        "path": {
                          "type": "object",
                          "description": "URL Path 参数"
                        },
                        "query": {
                          "type": "object",
                          "description": "URL Query 参数"
                        },
                        "body": {
                          "type": "object",
                          "description": "Request Body 参数"
                        }
                      }
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "action-recall"
                ],
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
            "create_time": 1770082202229131000,
            "update_time": 1770082202229131000,
            "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "extend_info": {},
            "resource_object": "tool",
            "source_id": "7eae029a-85a7-4f9a-af1c-a323beeb8c81",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "a666b98e-9278-401f-b4e5-e491bcaed3f8",
            "name": "query_instance_subgraph",
            "description": "基于预定义的关系路径查询知识图谱中的对象子图。支持多条路径查询，每条路径返回独立子图。对象以map形式返回，支持过滤条件和排序。query_type需设为\"relation_path\"。\r\n",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "1cad418b-8a73-445c-bf09-7168508e4d7b",
              "summary": "query_instance_subgraph",
              "description": "基于预定义的关系路径查询知识图谱中的对象子图。支持多条路径查询，每条路径返回独立子图。对象以map形式返回，支持过滤条件和排序。query_type需设为\"relation_path\"。\n",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/query_instance_subgraph",
              "method": "POST",
              "create_time": 1770082202223080700,
              "update_time": 1770082202223080700,
              "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "kn_id",
                    "in": "query",
                    "description": "业务知识网络ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "include_logic_params",
                    "in": "query",
                    "description": "包含逻辑属性的计算参数，默认false，返回结果不包含逻辑属性的字段和值",
                    "required": false,
                    "schema": {
                      "type": "boolean"
                    }
                  }
                ],
                "request_body": {
                  "description": "子图查询请求体",
                  "content": {
                    "application/json": {
                      "schema": {
                        "$ref": "#/components/schemas/SubGraphQueryBaseOnTypePath"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "对象子图查询响应体",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/PathEntries"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "RelationPath": {
                      "properties": {
                        "length": {
                          "description": "当前路径的长度",
                          "type": "integer"
                        },
                        "relations": {
                          "type": "array",
                          "description": "路径的边集合，沿着路径顺序出现的边",
                          "items": {
                            "$ref": "#/components/schemas/Relation"
                          }
                        }
                      },
                      "type": "object",
                      "description": "对象的关系路径",
                      "required": [
                        "relations",
                        "length"
                      ]
                    },
                    "RelationTypePath": {
                      "type": "object",
                      "description": "基于路径获取对象子图。**这是查询的核心结构**！用于定义完整的关系路径查询模板，包括路径中的所有对象类型和关系类型。object_types和relation_types数组的顺序**必须严格对应**，共同构成一个完整的关系路径。",
                      "required": [
                        "relation_types",
                        "object_types"
                      ],
                      "properties": {
                        "limit": {
                          "type": "integer",
                          "description": "当前路径返回的路径数量的限制。"
                        },
                        "object_types": {
                          "type": "array",
                          "description": "路径中的对象类集合，**顺序必须严格**与路径中节点出现顺序保持一致。对于n跳路径，object_types数组长度应为n+1，且必须按照source_object_type → 中间节点 → target_object_type的顺序排列。如果某个节点没有过滤条件或者排序或者限制数量，也必须保留其id字段以确保顺序正确。",
                          "items": {
                            "$ref": "#/components/schemas/ObjectTypeOnPath"
                          }
                        },
                        "relation_types": {
                          "items": {
                            "$ref": "#/components/schemas/TypeEdge"
                          },
                          "type": "array",
                          "description": "路径的边集合，**顺序必须严格**按照路径中关系出现的顺序排列。对于n跳路径，relation_types数组长度应为n，且必须与object_types数组中的对象类型严格对应：第i个relation_type的source_object_type_id必须等于object_types数组中第i个对象的id，target_object_type_id必须等于object_types数组中第i+1个对象的id。"
                        }
                      }
                    },
                    "PathEntries": {
                      "type": "object",
                      "description": "路径子图返回体",
                      "required": [
                        "entries"
                      ],
                      "properties": {
                        "entries": {
                          "description": "路径子图",
                          "items": {
                            "$ref": "#/components/schemas/ObjectSubGraphResponse"
                          },
                          "type": "array"
                        }
                      }
                    },
                    "Relation": {
                      "description": "一度关系（边）",
                      "required": [
                        "relation_type_id",
                        "relation_type_name",
                        "source_object_id",
                        "target_object_id"
                      ],
                      "properties": {
                        "relation_type_id": {
                          "type": "string",
                          "description": "关系类id"
                        },
                        "relation_type_name": {
                          "type": "string",
                          "description": "关系类名称"
                        },
                        "source_object_id": {
                          "description": "起点对象id",
                          "type": "string"
                        },
                        "target_object_id": {
                          "type": "string",
                          "description": "终点对象id"
                        }
                      },
                      "type": "object"
                    },
                    "Condition": {
                      "type": "object",
                      "description": "过滤条件结构，用于构建对象实例的查询筛选条件。\n\n**重要规则：**\n- `value_from` 和 `value` 必须同时使用，不能单独使用\n- `value_from` 当前仅支持 \"const\"（常量值）\n- 当使用 `value_from: \"const\"` 时，必须同时提供 `value` 字段\n",
                      "required": [
                        "operation"
                      ],
                      "properties": {
                        "sub_conditions": {
                          "description": "子过滤条件数组，用于逻辑操作符(and/or)的组合查询",
                          "items": {
                            "$ref": "#/components/schemas/Condition"
                          },
                          "type": "array"
                        },
                        "value": {
                          "oneOf": [
                            {
                              "type": "string"
                            },
                            {
                              "type": "number"
                            },
                            {
                              "type": "boolean"
                            },
                            {
                              "items": {
                                "oneOf": [
                                  {
                                    "type": "string"
                                  },
                                  {
                                    "type": "number"
                                  },
                                  {
                                    "type": "boolean"
                                  }
                                ]
                              },
                              "type": "array"
                            }
                          ],
                          "description": "字段值，格式根据操作符类型而定：\n- 比较操作符: 单个值\n- 范围查询: [min, max]数组\n- 集合操作: 值数组\n- 向量搜索: 特定格式数组\n\n**必须与 `value_from: \"const\"` 同时使用**\n"
                        },
                        "value_from": {
                          "description": "字段值来源。\n\n**重要：** 当前仅支持 \"const\"（常量值），且必须与 `value` 字段同时使用\n",
                          "enum": [
                            "const"
                          ],
                          "type": "string"
                        },
                        "field": {
                          "type": "string",
                          "description": "字段名称，也即对象类的属性名称"
                        },
                        "operation": {
                          "type": "string",
                          "description": "查询条件操作符。**注意：** 虽然这里列出了所有可能的操作符，但每个对象类实际支持的操作符列表以对象类定义中的 `condition_operations` 字段为准。",
                          "enum": [
                            "and",
                            "or",
                            "==",
                            "!=",
                            ">",
                            ">=",
                            "<",
                            "<=",
                            "in",
                            "not_in",
                            "like",
                            "not_like",
                            "exist",
                            "not_exist",
                            "match"
                          ]
                        }
                      }
                    },
                    "SubGraphQueryBaseOnTypePath": {
                      "description": "查询请求的顶层结构。用于基于关系类路径查询对象子图。relation_type_paths数组中可以包含多条不同的关系路径，系统会同时查询并返回所有路径的结果。每条路径必须符合严格的顺序和方向要求。",
                      "required": [
                        "relation_type_paths"
                      ],
                      "properties": {
                        "relation_type_paths": {
                          "type": "array",
                          "description": "关系类路径集合,数组中可以包含多条不同的关系路径，系统会同时查询并返回所有路径的结果。每条路径必须符合严格的顺序和方向要求。",
                          "items": {
                            "$ref": "#/components/schemas/RelationTypePath"
                          }
                        }
                      },
                      "type": "object"
                    },
                    "Sort": {
                      "type": "object",
                      "description": "排序字段",
                      "required": [
                        "field",
                        "direction"
                      ],
                      "properties": {
                        "field": {
                          "description": "排序字段",
                          "type": "string"
                        },
                        "direction": {
                          "enum": [
                            "desc",
                            "asc"
                          ],
                          "type": "string",
                          "description": "排序方向"
                        }
                      }
                    },
                    "TypeEdge": {
                      "required": [
                        "relation_type_id",
                        "source_object_type_id",
                        "target_object_type_id"
                      ],
                      "properties": {
                        "target_object_type_id": {
                          "type": "string",
                          "description": "路径的终点对象类id"
                        },
                        "relation_type_id": {
                          "type": "string",
                          "description": "关系类id"
                        },
                        "source_object_type_id": {
                          "type": "string",
                          "description": "路径的起点对象类id"
                        }
                      },
                      "type": "object",
                      "description": "路径中的边信息。**方向和顺序极其重要**！通过关系类id确定边，通过路径的起点对象类id和终点对象类id来确定当前路径的方向为正向还是反向，与关系类的起终点一致为正向，相反则为反向。每个TypeEdge必须与路径中的前后对象类型严格对应，这直接影响查询结果的正确性。"
                    },
                    "ObjectSubGraphResponse": {
                      "type": "object",
                      "description": "对象子图",
                      "required": [
                        "objects",
                        "relation_paths",
                        "total_count",
                        "search_after"
                      ],
                      "properties": {
                        "search_after": {
                          "type": "array",
                          "description": "表示返回的最后一个起点类对象的排序值，获取这个用于下一次 search_after 分页",
                          "items": {}
                        },
                        "total_count": {
                          "type": "integer",
                          "description": "起点对象类的总条数"
                        },
                        "objects": {
                          "description": "子图中的对象map，格式为：\n{\n  \"对象ID1\": {ObjectInfoInSubgraph对象1},\n  \"对象ID2\": {ObjectInfoInSubgraph对象2}\n}\n其中key是ObjectInfoInSubgraph中的id属性，value是完整的ObjectInfoInSubgraph对象。\n动态数据字段，其值可以是基本类型、MetricProperty或OperatorProperty\n",
                          "type": "object"
                        },
                        "relation_paths": {
                          "items": {
                            "$ref": "#/components/schemas/RelationPath"
                          },
                          "type": "array",
                          "description": "对象的关系路径集合"
                        }
                      }
                    },
                    "ObjectTypeOnPath": {
                      "properties": {
                        "limit": {
                          "description": "对象类获取对象数量的限制",
                          "type": "integer"
                        },
                        "sort": {
                          "type": "array",
                          "description": "对当前对象类的排序字段",
                          "items": {
                            "$ref": "#/components/schemas/Sort"
                          }
                        },
                        "condition": {
                          "$ref": "#/components/schemas/Condition"
                        },
                        "id": {
                          "description": "对象类id",
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "description": "路径中的对象类信息",
                      "required": [
                        "id",
                        "condition",
                        "limit"
                      ]
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": null,
                "external_docs": null
              }
            },
            "use_rule": "# 关系路径查询模板文档（无properties版本）\r\n\r\n## 核心规则\r\n\r\n### 1. 路径方向规则\r\n- ✅ **支持**：单向路径，所有箭头方向一致\r\n  - 正向：A → B → C → D\r\n  - 反向：D → C → B → A\r\n- ❌ **不支持**：双向路径或混合方向\r\n  - 不支持：A → B ← C（箭头方向不一致）\r\n  - 不支持：A ← B → C（箭头方向不一致）\r\n\r\n### 2. 数组顺序规则\r\n- `object_types` 数组的顺序必须与查询路径方向完全一致\r\n- `relation_types` 数组的顺序必须与 `object_types` 一一对应\r\n- 第 i 个 `relation_types[i]` 连接 `object_types[i]` 和 `object_types[i+1]`\r\n\r\n### 3. 关系方向规则\r\n- 知识网络的关系本身没有方向概念\r\n- 关系的 source 和 target 由查询方向决定\r\n- `relation_types[i].source_object_type_id` = `object_types[i].id`\r\n- `relation_types[i].target_object_type_id` = `object_types[i+1].id`\r\n\r\n---\r\n\r\n## 模板结构\r\n\r\n### 基础结构\r\n```yaml\r\nrelation_type_paths:\r\n  - object_types: []      # 对象类型数组，顺序必须与路径方向一致，必填项，不能为空\r\n    relation_types: []    # 关系类型数组，与object_types一一对应，必填项，不能为空\r\n    limit: 200            # 结果数量限制（默认200；后端上限10000；工具最终会截断返回，详见返回meta）\r\n```\r\n\r\n---\r\n\r\n## 模板示例\r\n\r\n### 模板1：两节点路径（A → B）\r\n\r\n#### 正向查询：从A查B\r\n```yaml\r\nrelation_type_paths:\r\n  - object_types:\r\n      - id: \"A\"\r\n        condition:          # 可选：查询条件\r\n          operation: \"and\"\r\n          sub_conditions:\r\n            - field: \"field_name\"\r\n              operation: \"==\"\r\n              value: \"value\"\r\n      - id: \"B\"\r\n    relation_types:\r\n      - relation_type_id: \"A_to_B\"\r\n        source_object_type_id: \"A\"        # 对应 object_types[0]\r\n        target_object_type_id: \"B\"        # 对应 object_types[1]\r\n    limit: 200\r\n```\r\n\r\n#### 反向查询：从B查A\r\n```yaml\r\nrelation_type_paths:\r\n  - object_types:\r\n      - id: \"B\"\r\n        condition:\r\n          operation: \"and\"\r\n          sub_conditions:\r\n            - field: \"field_name\"\r\n              operation: \"==\"\r\n              value: \"value\"\r\n      - id: \"A\"\r\n    relation_types:\r\n      - relation_type_id: \"A_to_B\"       # 关系ID不变\r\n        source_object_type_id: \"B\"       # 反转：对应 object_types[0]\r\n        target_object_type_id: \"A\"       # 反转：对应 object_types[1]\r\n    limit: 200\r\n```\r\n\r\n---\r\n\r\n### 模板2：三节点路径（A → B → C）\r\n\r\n#### 正向查询：从A查C（经过B）\r\n```yaml\r\nrelation_type_paths:\r\n  - object_types:\r\n      - id: \"A\"\r\n        condition:\r\n          operation: \"and\"\r\n          sub_conditions:\r\n            - field: \"field_name\"\r\n              operation: \"==\"\r\n              value: \"value\"\r\n      - id: \"B\"\r\n      - id: \"C\"\r\n    relation_types:\r\n      # 第一个关系：连接 object_types[0] 和 object_types[1]\r\n      - relation_type_id: \"A_to_B\"\r\n        source_object_type_id: \"A\"       # object_types[0]\r\n        target_object_type_id: \"B\"       # object_types[1]\r\n      # 第二个关系：连接 object_types[1] 和 object_types[2]\r\n      - relation_type_id: \"B_to_C\"\r\n        source_object_type_id: \"B\"       # object_types[1]\r\n        target_object_type_id: \"C\"       # object_types[2]\r\n    limit: 200\r\n```\r\n\r\n#### 反向查询：从C查A（经过B）\r\n```yaml\r\nrelation_type_paths:\r\n  - object_types:\r\n      - id: \"C\"\r\n        condition:\r\n          operation: \"and\"\r\n          sub_conditions:\r\n            - field: \"field_name\"\r\n              operation: \"==\"\r\n              value: \"value\"\r\n      - id: \"B\"\r\n      - id: \"A\"\r\n    relation_types:\r\n      # 第一个关系：连接 object_types[0] 和 object_types[1]\r\n      - relation_type_id: \"B_to_C\"       # 关系ID不变\r\n        source_object_type_id: \"C\"       # 反转：object_types[0]\r\n        target_object_type_id: \"B\"       # 反转：object_types[1]\r\n      # 第二个关系：连接 object_types[1] 和 object_types[2]\r\n      - relation_type_id: \"A_to_B\"       # 关系ID不变\r\n        source_object_type_id: \"B\"       # 反转：object_types[1]\r\n        target_object_type_id: \"A\"       # 反转：object_types[2]\r\n    limit: 200\r\n```\r\n\r\n---\r\n\r\n### 模板3：四节点路径（A → B → C → D）\r\n\r\n#### 正向查询：从A查D（经过B和C）\r\n```yaml\r\nrelation_type_paths:\r\n  - object_types:\r\n      - id: \"A\"\r\n        condition:\r\n          operation: \"and\"\r\n          sub_conditions:\r\n            - field: \"field_name\"\r\n              operation: \"==\"\r\n              value: \"value\"\r\n      - id: \"B\"\r\n      - id: \"C\"\r\n      - id: \"D\"\r\n    relation_types:\r\n      # 关系1：A → B\r\n      - relation_type_id: \"A_to_B\"\r\n        source_object_type_id: \"A\"       # object_types[0]\r\n        target_object_type_id: \"B\"       # object_types[1]\r\n      # 关系2：B → C\r\n      - relation_type_id: \"B_to_C\"\r\n        source_object_type_id: \"B\"       # object_types[1]\r\n        target_object_type_id: \"C\"       # object_types[2]\r\n      # 关系3：C → D\r\n      - relation_type_id: \"C_to_D\"\r\n        source_object_type_id: \"C\"       # object_types[2]\r\n        target_object_type_id: \"D\"       # object_types[3]\r\n    limit: 200\r\n```\r\n\r\n#### 反向查询：从D查A（经过C和B）\r\n```yaml\r\nrelation_type_paths:\r\n  - object_types:\r\n      - id: \"D\"\r\n        condition:\r\n          operation: \"and\"\r\n          sub_conditions:\r\n            - field: \"field_name\"\r\n              operation: \"==\"\r\n              value: \"value\"\r\n      - id: \"C\"\r\n      - id: \"B\"\r\n      - id: \"A\"\r\n    relation_types:\r\n      # 关系1：D ← C（原C_to_D的反向）\r\n      - relation_type_id: \"C_to_D\"\r\n        source_object_type_id: \"D\"       # 反转：object_types[0]\r\n        target_object_type_id: \"C\"       # 反转：object_types[1]\r\n      # 关系2：C ← B（原B_to_C的反向）\r\n      - relation_type_id: \"B_to_C\"\r\n        source_object_type_id: \"C\"       # 反转：object_types[1]\r\n        target_object_type_id: \"B\"       # 反转：object_types[2]\r\n      # 关系3：B ← A（原A_to_B的反向）\r\n      - relation_type_id: \"A_to_B\"\r\n        source_object_type_id: \"B\"       # 反转：object_types[2]\r\n        target_object_type_id: \"A\"       # 反转：object_types[3]\r\n    limit: 200\r\n```\r\n\r\n---\r\n\r\n## 实际应用示例\r\n\r\n### 示例1：张三的学校有哪些专业（三节点正向）\r\n\r\n**问题**：张三的学校有哪些专业\r\n\r\n**路径**：person → school → major\r\n\r\n```yaml\r\nrelation_type_paths:\r\n  - object_types:\r\n      - id: \"person\"\r\n        condition:\r\n          operation: \"and\"\r\n          sub_conditions:\r\n            - field: \"name\"\r\n              operation: \"==\"\r\n              value: \"张三\"\r\n      - id: \"school\"\r\n      - id: \"major\"\r\n    relation_types:\r\n      - relation_type_id: \"person_belongs_to_school\"\r\n        source_object_type_id: \"person\"\r\n        target_object_type_id: \"school\"\r\n      - relation_type_id: \"school_has_major\"\r\n        source_object_type_id: \"school\"\r\n        target_object_type_id: \"major\"\r\n    limit: 200\r\n```\r\n\r\n### 示例2：计算机专业在哪些学校开设，这些学校有哪些学生叫张三（三节点反向）\r\n\r\n**问题**：计算机专业在哪些学校开设，这些学校有哪些学生叫张三\r\n\r\n**路径**：major → school → person\r\n\r\n```yaml\r\nrelation_type_paths:\r\n  - object_types:\r\n      - id: \"major\"\r\n        condition:\r\n          operation: \"and\"\r\n          sub_conditions:\r\n            - field: \"major_name\"\r\n              operation: \"==\"\r\n              value: \"计算机\"\r\n      - id: \"school\"\r\n      - id: \"person\"\r\n        condition:\r\n          operation: \"and\"\r\n          sub_conditions:\r\n            - field: \"name\"\r\n              operation: \"==\"\r\n              value: \"张三\"\r\n    relation_types:\r\n      - relation_type_id: \"school_has_major\"\r\n        source_object_type_id: \"major\"\r\n        target_object_type_id: \"school\"\r\n      - relation_type_id: \"person_belongs_to_school\"\r\n        source_object_type_id: \"school\"\r\n        target_object_type_id: \"person\"\r\n    limit: 200\r\n```\r\n\r\n---\r\n\r\n## 关键检查点\r\n\r\n### ✅ 正确配置检查清单\r\n\r\n1. **路径方向一致性**\r\n   - [ ] 所有箭头方向是否一致（要么全部正向，要么全部反向）\r\n   - [ ] 不存在混合方向（如 A → B ← C）\r\n\r\n2. **数组顺序一致性**\r\n   - [ ] `object_types` 顺序是否与路径方向一致\r\n   - [ ] `relation_types` 数量是否等于 `object_types.length - 1`\r\n   - [ ] `relation_types[i]` 是否连接 `object_types[i]` 和 `object_types[i+1]`\r\n\r\n3. **关系配置正确性**\r\n   - [ ] `relation_types[i].source_object_type_id` 是否等于 `object_types[i].id`\r\n   - [ ] `relation_types[i].target_object_type_id` 是否等于 `object_types[i+1].id`\r\n\r\n4. **查询条件位置**\r\n   - [ ] 查询条件是否放在路径起点的对象类型上\r\n   - [ ] 正向查询：条件在第一个 `object_types[0]`\r\n   - [ ] 反向查询：条件在第一个 `object_types[0]`（反向后的起点）\r\n\r\n---\r\n\r\n## 常见错误示例\r\n\r\n### ❌ 错误1：object_types 顺序与路径方向不一致\r\n\r\n```yaml\r\n# 错误：想查 person → school → major，但顺序错了\r\nobject_types:\r\n  - id: \"person\"\r\n  - id: \"major\"      # ❌ 错误：应该是 school\r\n  - id: \"school\"     # ❌ 错误：应该是 major\r\n```\r\n\r\n### ❌ 错误2：relation_types 的 source/target 与 object_types 不对应\r\n\r\n```yaml\r\n# 错误：relation_types[0] 应该连接 object_types[0] 和 object_types[1]\r\nobject_types:\r\n  - id: \"person\"\r\n  - id: \"school\"\r\nrelation_types:\r\n  - relation_type_id: \"person_belongs_to_school\"\r\n    source_object_type_id: \"school\"     # ❌ 错误：应该是 \"person\"\r\n    target_object_type_id: \"person\"     # ❌ 错误：应该是 \"school\"\r\n```\r\n\r\n### ❌ 错误3：混合方向（不支持）\r\n\r\n```yaml\r\n# ❌ 不支持：A → B ← C 这种格式\r\n# 如果需要这种查询，需要拆分成两个独立的查询路径\r\n```\r\n\r\n---\r\n\r\n## 总结\r\n\r\n1. **路径必须是单向的**：A → B → C → D 或 D → C → B → A\r\n2. **数组顺序必须一致**：`object_types` 和 `relation_types` 的顺序要一一对应\r\n3. **关系方向由查询决定**：关系的 source/target 根据查询方向设置\r\n4. **查询条件在起点**：条件放在路径起点的对象类型上\r\n\r\n",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1770082202229131000,
            "update_time": 1770082202229131000,
            "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "extend_info": {},
            "resource_object": "tool",
            "source_id": "1cad418b-8a73-445c-bf09-7168508e4d7b",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "4b66ad66-4277-4d68-b847-c302c94265c9",
            "name": "query_object_instance",
            "description": "根据单个对象类查询对象实例，该接口基于业务知识网络语义检索接口返回的对象类定义，查询具体的对象实例数据。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "d27d3834-e0d0-4da1-b3c1-13a3aa40fbbb",
              "summary": "query_object_instance",
              "description": "根据单个对象类查询对象实例，该接口基于业务知识网络语义检索接口返回的对象类定义，查询具体的对象实例数据。",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/query_object_instance",
              "method": "POST",
              "create_time": 1770082202223080700,
              "update_time": 1770082202223080700,
              "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "kn_id",
                    "in": "query",
                    "description": "业务知识网络ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "ot_id",
                    "in": "query",
                    "description": "对象类ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "include_type_info",
                    "in": "query",
                    "description": "是否包含对象类信息, 默认false，不包含",
                    "required": false,
                    "schema": {
                      "type": "boolean"
                    }
                  },
                  {
                    "name": "include_logic_params",
                    "in": "query",
                    "description": "包含逻辑属性的计算参数，默认false，返回结果不包含逻辑属性的字段和值",
                    "required": false,
                    "schema": {
                      "type": "boolean"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "schema": {
                        "$ref": "#/components/schemas/FirstQueryWithSearchAfter"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "ok",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ObjectDataResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "FirstQueryWithSearchAfter": {
                      "required": [
                        "limit"
                      ],
                      "properties": {
                        "sort": {
                          "description": "排序字段，默认使用 @timestamp排序，排序方向为 desc",
                          "items": {
                            "$ref": "#/components/schemas/Sort"
                          },
                          "type": "array"
                        },
                        "condition": {
                          "$ref": "#/components/schemas/Condition"
                        },
                        "limit": {
                          "description": "返回的数量，默认值 10。范围 1-100",
                          "type": "integer"
                        },
                        "need_total": {
                          "type": "boolean",
                          "description": "是否需要总数，默认false"
                        },
                        "properties": {
                          "description": "指定返回的对象属性字段列表，默认返回所有属性。",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        }
                      },
                      "type": "object",
                      "description": "分页查询的第一次查询请求"
                    },
                    "Condition": {
                      "description": "过滤条件结构，用于构建对象实例的查询筛选条件。\n\n**重要规则：**\n- `value_from` 和 `value` 必须同时使用，不能单独使用\n- `value_from` 当前仅支持 \"const\"（常量值）\n- 当使用 `value_from: \"const\"` 时，必须同时提供 `value` 字段\n",
                      "required": [
                        "operation"
                      ],
                      "properties": {
                        "sub_conditions": {
                          "type": "array",
                          "description": "子过滤条件数组，用于逻辑操作符(and/or)的组合查询",
                          "items": {
                            "$ref": "#/components/schemas/Condition"
                          }
                        },
                        "value": {
                          "description": "字段值，格式根据操作符类型而定：\n- 比较操作符: 单个值\n- 范围查询: [min, max]数组\n- 集合操作: 值数组\n- 向量搜索: 特定格式数组\n\n**必须与 `value_from: \"const\"` 同时使用**\n"
                        },
                        "value_from": {
                          "type": "string",
                          "description": "字段值来源。\n\n**重要：** 当前仅支持 \"const\"（常量值），且必须与 `value` 字段同时使用\n",
                          "enum": [
                            "const"
                          ]
                        },
                        "field": {
                          "type": "string",
                          "description": "字段名称，也即对象类的属性名称"
                        },
                        "operation": {
                          "enum": [
                            "and",
                            "or",
                            "==",
                            "!=",
                            ">",
                            ">=",
                            "<",
                            "<=",
                            "in",
                            "not_in",
                            "like",
                            "not_like",
                            "exist",
                            "not_exist",
                            "match"
                          ],
                          "type": "string",
                          "description": "查询条件操作符。\n**注意：** 虽然这里列出了所有可能的操作符，但每个对象类实际支持的操作符列表以对象类定义中的 `condition_operations` 字段为准。\n"
                        }
                      },
                      "type": "object"
                    },
                    "ObjectTypeDetail": {
                      "type": "object",
                      "description": "对象类信息",
                      "properties": {
                        "comment": {
                          "description": "备注（可以为空）",
                          "type": "string"
                        },
                        "update_time": {
                          "description": "最近一次更新时间",
                          "type": "integer",
                          "format": "int64"
                        },
                        "name": {
                          "type": "string",
                          "description": "对象类名称"
                        },
                        "icon": {
                          "type": "string",
                          "description": "图标"
                        },
                        "id": {
                          "description": "对象类ID",
                          "type": "string"
                        },
                        "creator": {
                          "type": "string",
                          "description": "创建人ID"
                        },
                        "display_key": {
                          "type": "string",
                          "description": "对象实例的显示属性"
                        },
                        "updater": {
                          "type": "string",
                          "description": "最近一次修改人"
                        },
                        "module_type": {
                          "description": "模块类型",
                          "enum": [
                            "object_type"
                          ],
                          "type": "string"
                        },
                        "detail": {
                          "type": "string",
                          "description": "说明书。按需返回，若指定了include_detail=true，则返回，否则不返回。列表查询时不返回此字段"
                        },
                        "logic_properties": {
                          "description": "逻辑属性",
                          "items": {
                            "$ref": "#/components/schemas/LogicProperty"
                          },
                          "type": "array"
                        },
                        "data_properties": {
                          "items": {
                            "$ref": "#/components/schemas/DataProperty"
                          },
                          "type": "array",
                          "description": "数据属性"
                        },
                        "data_source": {
                          "$ref": "#/components/schemas/DataSource"
                        },
                        "branch": {
                          "type": "string",
                          "description": "分支ID"
                        },
                        "color": {
                          "type": "string",
                          "description": "颜色"
                        },
                        "create_time": {
                          "description": "创建时间",
                          "type": "integer",
                          "format": "int64"
                        },
                        "tags": {
                          "type": "array",
                          "description": "标签。 （可以为空）",
                          "items": {
                            "type": "string"
                          }
                        },
                        "primary_keys": {
                          "type": "array",
                          "description": "主键",
                          "items": {
                            "type": "string"
                          }
                        },
                        "concept_groups": {
                          "items": {
                            "$ref": "#/components/schemas/ConceptGroup"
                          },
                          "type": "array",
                          "description": "概念分组id"
                        },
                        "kn_id": {
                          "description": "业务知识网络id",
                          "type": "string"
                        }
                      }
                    },
                    "Parameter4Metric": {
                      "description": "逻辑参数",
                      "required": [
                        "name",
                        "value_from",
                        "operation"
                      ],
                      "properties": {
                        "value": {
                          "type": "string",
                          "description": "参数值。value_from=property时，填入的是对象类的数据属性名称；value_from=input时，不设置此字段"
                        },
                        "value_from": {
                          "enum": [
                            "property",
                            "input"
                          ],
                          "type": "string",
                          "description": "值来源"
                        },
                        "name": {
                          "type": "string",
                          "description": "参数名称"
                        },
                        "operation": {
                          "type": "string",
                          "description": "操作符。映射指标模型的属性时，此字段必须",
                          "enum": [
                            "in",
                            "=",
                            "!=",
                            ">",
                            ">=",
                            "<",
                            "<="
                          ]
                        }
                      },
                      "type": "object"
                    },
                    "ViewField": {
                      "type": "object",
                      "description": "视图字段信息",
                      "required": [
                        "name"
                      ],
                      "properties": {
                        "name": {
                          "type": "string",
                          "description": "字段名称"
                        },
                        "type": {
                          "description": "视图字段类型，查看时有此字段",
                          "type": "string"
                        },
                        "display_name": {
                          "type": "string",
                          "description": "字段显示名.查看时有此字段"
                        }
                      }
                    },
                    "LogicProperty": {
                      "description": "逻辑属性",
                      "required": [
                        "name",
                        "data_source",
                        "parameters"
                      ],
                      "properties": {
                        "name": {
                          "description": "属性名称。只能包含小写英文字母、数字、下划线（_）、连字符（-），且不能以下划线和连字符开头",
                          "type": "string"
                        },
                        "parameters": {
                          "type": "array",
                          "description": "逻辑所需的参数",
                          "items": {
                            "$ref": "#/components/schemas/Parameter"
                          }
                        },
                        "type": {
                          "type": "string",
                          "description": "属性数据类型。除了视图的字段类型之外，还有 metric、objective、event、trace、log、operator"
                        },
                        "comment": {
                          "type": "string",
                          "description": "属性描述"
                        },
                        "data_source": {
                          "$ref": "#/components/schemas/LogicSource"
                        },
                        "display_name": {
                          "description": "属性显示名",
                          "type": "string"
                        },
                        "index": {
                          "type": "boolean",
                          "description": "是否开启索引，默认是true"
                        }
                      },
                      "type": "object"
                    },
                    "FulltextConfig": {
                      "required": [
                        "analyzer",
                        "field_keyword"
                      ],
                      "properties": {
                        "field_keyword": {
                          "description": "是否保留原始字符串，保留原始字符串可用于精确匹配。默认是false",
                          "type": "boolean"
                        },
                        "analyzer": {
                          "type": "string",
                          "description": "分词器",
                          "enum": [
                            "standard",
                            "ik_max_word"
                          ]
                        }
                      },
                      "type": "object",
                      "description": "全文索引的配置"
                    },
                    "ConceptGroup": {
                      "required": [
                        "id",
                        "name"
                      ],
                      "properties": {
                        "id": {
                          "type": "string",
                          "description": "概念分组ID"
                        },
                        "name": {
                          "type": "string",
                          "description": "概念分组名称"
                        }
                      },
                      "type": "object",
                      "description": "概念分组"
                    },
                    "VectorConfig": {
                      "type": "object",
                      "description": "向量索引的配置",
                      "required": [
                        "dimension"
                      ],
                      "properties": {
                        "dimension": {
                          "type": "integer",
                          "description": "向量维度"
                        }
                      }
                    },
                    "Parameter": {
                      "description": "逻辑/指标参数",
                      "oneOf": [
                        {
                          "$ref": "#/components/schemas/Parameter4Operator"
                        },
                        {
                          "$ref": "#/components/schemas/Parameter4Metric"
                        }
                      ],
                      "type": "object"
                    },
                    "Parameter4Operator": {
                      "description": "逻辑参数",
                      "required": [
                        "name",
                        "value_from"
                      ],
                      "properties": {
                        "name": {
                          "description": "参数名称",
                          "type": "string"
                        },
                        "source": {
                          "description": "参数来源",
                          "type": "string"
                        },
                        "type": {
                          "type": "string",
                          "description": "参数类型"
                        },
                        "value": {
                          "type": "string",
                          "description": "参数值。value_from=property时，填入的是对象类的数据属性名称；value_from=input时，不设置此字段"
                        },
                        "value_from": {
                          "type": "string",
                          "description": "值来源",
                          "enum": [
                            "property",
                            "input"
                          ]
                        }
                      },
                      "type": "object"
                    },
                    "DataSource": {
                      "required": [
                        "type",
                        "id"
                      ],
                      "properties": {
                        "name": {
                          "type": "string",
                          "description": "名称。查看详情时返回。"
                        },
                        "type": {
                          "type": "string",
                          "description": "数据来源类型为数据视图",
                          "enum": [
                            "data_view"
                          ]
                        },
                        "id": {
                          "description": "数据视图ID",
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "description": "数据来源"
                    },
                    "Sort": {
                      "type": "object",
                      "description": "排序字段",
                      "required": [
                        "field",
                        "direction"
                      ],
                      "properties": {
                        "field": {
                          "type": "string",
                          "description": "排序字段"
                        },
                        "direction": {
                          "enum": [
                            "desc",
                            "asc"
                          ],
                          "type": "string",
                          "description": "排序方向"
                        }
                      }
                    },
                    "DataProperty": {
                      "type": "object",
                      "description": "数据属性",
                      "required": [
                        "name",
                        "display_name",
                        "type",
                        "comment",
                        "mapped_field",
                        "index",
                        "fulltext_config",
                        "vector_config"
                      ],
                      "properties": {
                        "display_name": {
                          "type": "string",
                          "description": "属性显示名"
                        },
                        "fulltext_config": {
                          "$ref": "#/components/schemas/FulltextConfig"
                        },
                        "index": {
                          "type": "boolean",
                          "description": "是否开启索引，默认是true"
                        },
                        "mapped_field": {
                          "$ref": "#/components/schemas/ViewField"
                        },
                        "name": {
                          "type": "string",
                          "description": "属性名称。只能包含小写英文字母、数字、下划线（_）、连字符（-），且不能以下划线和连字符开头"
                        },
                        "type": {
                          "description": "属性数据类型。除了视图的字段类型之外，还有 metric、objective、event、trace、log、operator",
                          "type": "string"
                        },
                        "vector_config": {
                          "$ref": "#/components/schemas/VectorConfig"
                        },
                        "comment": {
                          "description": "属性描述",
                          "type": "string"
                        }
                      }
                    },
                    "LogicSource": {
                      "description": "数据来源",
                      "required": [
                        "type",
                        "id"
                      ],
                      "properties": {
                        "type": {
                          "enum": [
                            "metric",
                            "operator"
                          ],
                          "type": "string",
                          "description": "数据来源类型"
                        },
                        "id": {
                          "description": "数据来源ID",
                          "type": "string"
                        },
                        "name": {
                          "description": "名称。查看详情时返回。",
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "ObjectDataResponse": {
                      "properties": {
                        "object_type": {
                          "$ref": "#/components/schemas/ObjectTypeDetail"
                        },
                        "search_after": {
                          "items": {},
                          "type": "array",
                          "description": "表示返回的最后一个文档的排序值，获取这个用于下一次 search_after 分页"
                        },
                        "total_count": {
                          "type": "integer",
                          "description": "总条数"
                        },
                        "datas": {
                          "type": "array",
                          "description": "对象实例数据。动态数据字段，其值可以是基本类型、MetricProperty或OperatorProperty",
                          "items": {
                            "type": "object"
                          }
                        }
                      },
                      "type": "object",
                      "description": "节点（对象类）信息",
                      "required": [
                        "groups",
                        "type",
                        "datas",
                        "search_after"
                      ]
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": null,
                "external_docs": null
              }
            },
            "use_rule": "建议：如果涉及的模糊匹配，若没有明确指定根据前缀还是后缀进行匹配，建议搜索词前后都追加%",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1770082202229131000,
            "update_time": 1770082202229131000,
            "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "extend_info": {},
            "resource_object": "tool",
            "source_id": "d27d3834-e0d0-4da1-b3c1-13a3aa40fbbb",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "fe3f1f26-ff56-4f62-a5c7-a68d1e8ac2d7",
            "name": "kn_search",
            "description": "基于知识网络的智能检索工具，支持传入完整的问题或一个或多个关键词，能够检索问题或关键词的属性信息和上下文信息。\r\n支持概念召回、语义实例召回、多轮对话等功能。\r\n",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "b1d8e6f9-3af4-47f5-bf0e-2e74ca6c7f66",
              "summary": "kn_search",
              "description": "基于知识网络的智能检索工具，支持传入完整的问题或一个或多个关键词，能够检索问题或关键词的属性信息和上下文信息。\n支持概念召回、语义实例召回、多轮对话等功能。\n",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/kn_search",
              "method": "POST",
              "create_time": 1770082202223080700,
              "update_time": 1770082202223080700,
              "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户ID，用于内部服务调用时传递账户信息",
                    "required": false,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型：user(用户), app(应用), anonymous(匿名)",
                    "required": false,
                    "schema": {
                      "default": "user",
                      "enum": [
                        "user",
                        "app",
                        "anonymous"
                      ],
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "kn_search 请求体",
                  "content": {
                    "application/json": {
                      "schema": {
                        "$ref": "#/components/schemas/KnSearchRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "成功返回检索结果",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/KnSearchResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "400",
                    "description": "参数错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "500",
                    "description": "服务器内部错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "LogicProperty": {
                      "type": "object",
                      "properties": {
                        "display_name": {
                          "type": "string",
                          "description": "属性显示名称"
                        },
                        "name": {
                          "type": "string",
                          "description": "属性名称"
                        }
                      }
                    },
                    "ActionType": {
                      "type": "object",
                      "description": "操作类型信息。精简模式（schema_brief=True）下仅包含：id, name, action_type, object_type_id, object_type_name, comment, tags, kn_id",
                      "properties": {
                        "comment": {
                          "type": "string",
                          "description": "注释说明"
                        },
                        "id": {
                          "type": "string",
                          "description": "操作类型ID"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "知识网络ID"
                        },
                        "name": {
                          "type": "string",
                          "description": "操作类型名称"
                        },
                        "object_type_id": {
                          "type": "string",
                          "description": "对象类型ID"
                        },
                        "object_type_name": {
                          "type": "string",
                          "description": "对象类型名称"
                        },
                        "tags": {
                          "description": "标签列表",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        },
                        "action_type": {
                          "type": "string",
                          "description": "操作类型（如：add, modify等）"
                        }
                      }
                    },
                    "KnSearchRequest": {
                      "type": "object",
                      "required": [
                        "query",
                        "kn_id"
                      ],
                      "properties": {
                        "retrieval_config": {
                          "type": "object",
                          "description": "召回配置参数，用于控制不同类型的召回场景（概念召回、语义实例召回、属性过滤）。如果不提供，将使用系统默认配置。",
                          "properties": {
                            "concept_retrieval": {
                              "$ref": "#/components/schemas/ConceptRetrievalConfig"
                            }
                          }
                        },
                        "session_id": {
                          "type": "string",
                          "description": "会话ID，用于维护多轮对话存储的历史召回记录"
                        },
                        "additional_context": {
                          "type": "string",
                          "description": "额外的上下文信息，用于二次检索时提供更精确的检索信息"
                        },
                        "enable_rerank": {
                          "type": "boolean",
                          "description": "是否启用重排序。如果为true，则启用重排序。",
                          "default": true
                        },
                        "kn_id": {
                          "description": "指定的知识网络ID，必须传递",
                          "type": "string"
                        },
                        "only_schema": {
                          "type": "boolean",
                          "description": "是否只召回概念（schema），不召回语义实例。如果为True，则只返回object_types、relation_types和action_types，不返回nodes。",
                          "default": false
                        },
                        "query": {
                          "type": "string",
                          "description": "用户查询问题或关键词，多个关键词之间用空格隔开"
                        }
                      }
                    },
                    "ErrorResponse": {
                      "properties": {
                        "error": {
                          "type": "string",
                          "description": "错误信息"
                        },
                        "message": {
                          "description": "错误详情",
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "ConceptRetrievalConfig": {
                      "type": "object",
                      "description": "概念召回/概念流程配置参数（原最外层参数已收敛到此处）",
                      "properties": {
                        "schema_brief": {
                          "default": true,
                          "type": "boolean",
                          "description": "概念召回时是否返回精简schema。True仅返回必要字段（概念ID/名称/关系source&target），不返回大字段。"
                        },
                        "return_union": {
                          "default": false,
                          "type": "boolean",
                          "description": "概念召回多轮检索时是否返回并集。True返回所有轮次并集；False仅返回当前轮次增量（默认False）。"
                        }
                      }
                    },
                    "Node": {
                      "description": "节点数据，至少包含 object_type_id、<object_type_id>_name、unique_identities",
                      "properties": {
                        "unique_identities": {
                          "type": "object",
                          "description": "对象的唯一标识信息"
                        },
                        "object_type_id": {
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "DataProperty": {
                      "type": "object",
                      "properties": {
                        "comment": {
                          "type": "string",
                          "description": "属性描述（非精简模式）"
                        },
                        "display_name": {
                          "description": "属性显示名称",
                          "type": "string"
                        },
                        "name": {
                          "type": "string",
                          "description": "属性名称"
                        }
                      }
                    },
                    "RelationType": {
                      "type": "object",
                      "required": [
                        "concept_id",
                        "concept_name",
                        "source_object_type_id",
                        "target_object_type_id"
                      ],
                      "properties": {
                        "target_object_type_id": {
                          "description": "目标对象类型ID",
                          "type": "string"
                        },
                        "concept_id": {
                          "description": "概念ID",
                          "type": "string"
                        },
                        "concept_name": {
                          "type": "string",
                          "description": "概念名称"
                        },
                        "concept_type": {
                          "type": "string",
                          "description": "概念类型: relation_type"
                        },
                        "source_object_type_id": {
                          "type": "string",
                          "description": "源对象类型ID"
                        }
                      }
                    },
                    "KnSearchResponse": {
                      "type": "object",
                      "description": "检索结果，返回object_types/relation_types/action_types，并返回语义实例nodes/message。\n多轮时由concept_retrieval.return_union控制 nodes 的并集/增量。\n",
                      "properties": {
                        "nodes": {
                          "type": "array",
                          "description": "语义实例召回结果（当不提供conditions且召回到实例时返回），与条件召回节点风格对齐的扁平列表。\n每个节点至少包含 object_type_id、<object_type_id>_name、unique_identities\n",
                          "items": {
                            "$ref": "#/components/schemas/Node"
                          }
                        },
                        "object_types": {
                          "items": {
                            "$ref": "#/components/schemas/ObjectType"
                          },
                          "type": "array",
                          "description": "对象类型列表（概念召回时返回）。\n当schema_brief=True时，仅包含：concept_id, concept_name, comment, data_properties（仅name和display_name）, logic_properties（仅name和display_name）, sample_data（当include_sample_data=True时）。\n当schema_brief=False时，包含完整字段（包括primary_keys, display_key, sample_data等）\n"
                        },
                        "relation_types": {
                          "type": "array",
                          "description": "关系类型列表（概念召回时返回）。\n精简模式和完整模式均包含：concept_id, concept_name, source_object_type_id, target_object_type_id\n",
                          "items": {
                            "$ref": "#/components/schemas/RelationType"
                          }
                        },
                        "action_types": {
                          "description": "操作类型列表（概念召回时返回）。\n当schema_brief=True时，每个action_type仅包含以下字段：id, name, action_type, object_type_id, object_type_name, comment, tags, kn_id\n",
                          "items": {
                            "$ref": "#/components/schemas/ActionType"
                          },
                          "type": "array"
                        },
                        "message": {
                          "type": "string",
                          "description": "提示信息（例如未召回到实例数据时返回原因说明）"
                        }
                      }
                    },
                    "ObjectType": {
                      "type": "object",
                      "required": [
                        "concept_id",
                        "concept_name"
                      ],
                      "properties": {
                        "display_key": {
                          "type": "string",
                          "description": "显示字段名（用于获取instance_name）。仅当schema_brief=False时返回"
                        },
                        "sample_data": {
                          "description": "样例数据（当include_sample_data=True时返回，无论schema_brief是否为True）",
                          "type": "object"
                        },
                        "logic_properties": {
                          "description": "逻辑属性列表（指标等）。精简模式下仅包含name和display_name字段（数量不截断）",
                          "items": {
                            "$ref": "#/components/schemas/LogicProperty"
                          },
                          "type": "array"
                        },
                        "concept_type": {
                          "type": "string",
                          "description": "概念类型: object_type"
                        },
                        "data_properties": {
                          "items": {
                            "$ref": "#/components/schemas/DataProperty"
                          },
                          "type": "array",
                          "description": "对象属性列表。精简模式下仅包含name和display_name字段（数量不截断）"
                        },
                        "comment": {
                          "description": "概念描述",
                          "type": "string"
                        },
                        "concept_name": {
                          "description": "概念名称",
                          "type": "string"
                        },
                        "concept_id": {
                          "description": "概念ID",
                          "type": "string"
                        },
                        "primary_keys": {
                          "description": "主键字段列表（支持多个主键）。仅当schema_brief=False时返回",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        }
                      }
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "kn-search"
                ],
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
            "create_time": 1770082202229131000,
            "update_time": 1770082202229131000,
            "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "extend_info": {},
            "resource_object": "tool",
            "source_id": "b1d8e6f9-3af4-47f5-bf0e-2e74ca6c7f66",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "b06c7316-9359-4aff-a425-4db8f43f2837",
            "name": "创建业务知识网络构建任务",
            "description": "创建一个全量构建业务知识网络的任务",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "c3076f12-6503-4c34-bd4b-3982757fa63f",
              "summary": "创建业务知识网络构建任务",
              "description": "创建一个全量构建业务知识网络的任务",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/full_build_ontology",
              "method": "POST",
              "create_time": 1770082202223080700,
              "update_time": 1770082202223080700,
              "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户ID，用于内部服务调用时传递账户信息",
                    "required": false,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型：user(用户), app(应用), anonymous(匿名)",
                    "required": false,
                    "schema": {
                      "default": "user",
                      "enum": [
                        "user",
                        "app",
                        "anonymous"
                      ],
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "example": {
                        "kn_id": "kn_1234567890",
                        "name": "全量构建任务"
                      },
                      "schema": {
                        "$ref": "#/components/schemas/CreateJobRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "201",
                    "description": "创建成功",
                    "content": {
                      "application/json": {
                        "example": {
                          "id": "job_1234567890"
                        },
                        "schema": {
                          "$ref": "#/components/schemas/CreateJobResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "400",
                    "description": "参数错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "401",
                    "description": "未授权",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "500",
                    "description": "服务器内部错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "CreateJobRequest": {
                      "properties": {
                        "kn_id": {
                          "type": "string",
                          "description": "业务知识网络ID"
                        },
                        "name": {
                          "type": "string",
                          "description": "任务名称"
                        }
                      },
                      "type": "object",
                      "required": [
                        "kn_id",
                        "name"
                      ]
                    },
                    "CreateJobResponse": {
                      "properties": {
                        "id": {
                          "type": "string",
                          "description": "任务ID"
                        }
                      },
                      "type": "object"
                    },
                    "ErrorResponse": {
                      "type": "object",
                      "properties": {
                        "detail": {
                          "type": "object",
                          "description": "错误详情"
                        },
                        "link": {
                          "type": "string",
                          "description": "错误链接"
                        },
                        "solution": {
                          "type": "string",
                          "description": "解决方案"
                        },
                        "code": {
                          "type": "string",
                          "description": "错误码"
                        },
                        "description": {
                          "type": "string",
                          "description": "错误描述"
                        }
                      }
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "OntologyJob"
                ],
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
            "create_time": 1770082202229131000,
            "update_time": 1770114146324083500,
            "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "c3076f12-6503-4c34-bd4b-3982757fa63f",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "139e027f-af32-4d5c-9d8b-4fa48b185d96",
            "name": "获取业务知识网络构建状态",
            "description": "查询最新50个构建任务的整体状态（按创建时间倒排）。如果所有任务都已完成则返回completed，如果有任务正在运行则返回running",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "03a34bdf-3cfd-4814-bb70-47ec38f9855f",
              "summary": "获取业务知识网络构建状态",
              "description": "查询最新50个构建任务的整体状态（按创建时间倒排）。如果所有任务都已完成则返回completed，如果有任务正在运行则返回running",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/full_ontology_building_status",
              "method": "GET",
              "create_time": 1770082202223080700,
              "update_time": 1770082202223080700,
              "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户ID，用于内部服务调用时传递账户信息",
                    "required": false,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型：user(用户), app(应用), anonymous(匿名)",
                    "required": false,
                    "schema": {
                      "default": "user",
                      "enum": [
                        "user",
                        "app",
                        "anonymous"
                      ],
                      "type": "string"
                    }
                  },
                  {
                    "name": "kn_id",
                    "in": "query",
                    "description": "业务知识网络ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {},
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "400",
                    "description": "参数错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "401",
                    "description": "未授权",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "500",
                    "description": "服务器内部错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "200",
                    "description": "成功返回构建状态",
                    "content": {
                      "application/json": {
                        "example": {
                          "kn_id": "d5levlh818p1vl2slp60",
                          "state": "completed",
                          "state_detail": "All latest 50 jobs are completed"
                        },
                        "schema": {
                          "$ref": "#/components/schemas/BuildStatusSimpleResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "BuildStatusSimpleResponse": {
                      "required": [
                        "kn_id",
                        "state",
                        "state_detail"
                      ],
                      "properties": {
                        "state_detail": {
                          "type": "string",
                          "description": "状态详情"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "业务知识网络ID"
                        },
                        "state": {
                          "description": "构建状态（running表示有任务正在运行，completed表示所有任务都已完成）",
                          "enum": [
                            "running",
                            "completed"
                          ],
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "description": "构建状态响应"
                    },
                    "ErrorResponse": {
                      "type": "object",
                      "properties": {
                        "detail": {
                          "description": "错误详情",
                          "type": "object"
                        },
                        "link": {
                          "description": "错误链接",
                          "type": "string"
                        },
                        "solution": {
                          "type": "string",
                          "description": "解决方案"
                        },
                        "code": {
                          "type": "string",
                          "description": "错误码"
                        },
                        "description": {
                          "type": "string",
                          "description": "错误描述"
                        }
                      }
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "OntologyJob"
                ],
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
            "create_time": 1770082202229131000,
            "update_time": 1770114146992381000,
            "create_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "update_user": "0ae82800-6f60-11f0-b0dc-36fa540cff80",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "03a34bdf-3cfd-4814-bb70-47ec38f9855f",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          }
        ],
        "create_time": 1767507418690066000,
        "update_time": 1770102127780611000,
        "create_user": "4c20aa70-6f67-11f0-b0dc-36fa540cff80",
        "update_user": "48adab0e-6f67-11f0-b0dc-36fa540cff80",
        "metadata_type": "openapi"
      }
    ]
  }
}