{
  "toolbox": {
    "configs": [
      {
        "box_id": "e521d454-4a0b-4dc9-8a28-d0986de1cef9",
        "box_name": "contextloader工具集",
        "box_desc": "ContextLoader 标准内置工具集；契约版本: 0.8.0",
        "box_svc_url": "http://agent-retrieval:30779",
        "status": "published",
        "category_type": "data_query",
        "category_name": "数据查询",
        "is_internal": false,
        "source": "custom",
        "tools": [
          {
            "tool_id": "05275bb1-46e2-4727-9c6f-97d9ea0af94b",
            "name": "get_logic_properties_values",
            "description": "根据 query 生成 dynamic_params，批量查询指定对象的逻辑属性值。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "bd55b7ef-8804-4690-af9d-b9996bf5d3ea",
              "summary": "get_logic_properties_values",
              "description": "根据 query 生成 dynamic_params，批量查询指定对象的逻辑属性值。",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/logic-property-resolver",
              "method": "POST",
              "create_time": 1776920840666727400,
              "update_time": 1776920840666727400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
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
                      "enum": [
                        "user",
                        "app",
                        "anonymous"
                      ],
                      "type": "string"
                    }
                  },
                  {
                    "name": "response_format",
                    "in": "query",
                    "description": "响应格式：json 或 toon，默认 json",
                    "required": false,
                    "schema": {
                      "default": "json",
                      "enum": [
                        "json",
                        "toon"
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
                        "示例": {
                          "value": {
                            "_instance_identities": [
                              {
                                "company_id": "company_000001"
                              }
                            ],
                            "kn_id": "kn_medical",
                            "ot_id": "company",
                            "properties": [
                              "approved_drug_count",
                              "business_health_score"
                            ],
                            "query": "最近一年这些药企的药品上市数量和健康度"
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
                    "ResolveLogicPropertiesResponse": {
                      "oneOf": [
                        {
                          "$ref": "#/components/schemas/ObjectPropertiesValuesResponse"
                        },
                        {
                          "$ref": "#/components/schemas/MissingParamsError"
                        }
                      ],
                      "description": "成功返回 datas；缺参时返回 error_code、missing（含 hint）"
                    },
                    "ObjectPropertiesValuesResponse": {
                      "properties": {
                        "debug": {
                          "$ref": "#/components/schemas/ResolveDebugInfo"
                        },
                        "datas": {
                          "items": {
                            "type": "object"
                          },
                          "type": "array",
                          "description": "与 _instance_identities 顺序对齐，每项含主键和请求的 properties"
                        }
                      },
                      "type": "object",
                      "required": [
                        "datas"
                      ]
                    },
                    "ResolveDebugInfo": {
                      "properties": {
                        "warnings": {
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        },
                        "dynamic_params": {
                          "type": "object"
                        },
                        "now_ms": {
                          "type": "integer"
                        },
                        "trace_id": {
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "MissingParamsError": {
                      "type": "object",
                      "properties": {
                        "message": {
                          "type": "string"
                        },
                        "missing": {
                          "items": {
                            "type": "object",
                            "properties": {
                              "params": {
                                "type": "array",
                                "items": {
                                  "type": "object",
                                  "properties": {
                                    "name": {
                                      "type": "string"
                                    },
                                    "hint": {
                                      "type": "string"
                                    }
                                  }
                                }
                              },
                              "property": {
                                "type": "string"
                              }
                            }
                          },
                          "type": "array"
                        },
                        "error_code": {
                          "type": "string"
                        }
                      }
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
                    "ResolveLogicPropertiesRequest": {
                      "type": "object",
                      "required": [
                        "kn_id",
                        "ot_id",
                        "query",
                        "_instance_identities",
                        "properties"
                      ],
                      "properties": {
                        "query": {
                          "type": "string",
                          "description": "用户查询，需含时间（如\"最近一年\"）、统计维度、业务上下文，用于生成 dynamic_params"
                        },
                        "_instance_identities": {
                          "items": {
                            "type": "object"
                          },
                          "type": "array",
                          "description": "对象实例标识数组。**必须从上游提取，不可臆造。** 流程：先调 query_object_instance 或 query_instance_subgraph → 从每个对象的 _instance_identity 字段取值 → 按原顺序组成数组传入。"
                        },
                        "additional_context": {
                          "description": "可选。补充上下文，如 timezone、instant、step、对象属性等，帮助生成 dynamic_params。",
                          "type": "string"
                        },
                        "kn_id": {
                          "description": "知识网络ID。例 kn_medical",
                          "type": "string"
                        },
                        "options": {
                          "$ref": "#/components/schemas/ResolveOptions"
                        },
                        "ot_id": {
                          "description": "对象类ID。例 company、drug",
                          "type": "string"
                        },
                        "properties": {
                          "description": "逻辑属性名列表（metric/operator）。自动生成 dynamic_params 并查询。",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        }
                      }
                    },
                    "ResolveOptions": {
                      "properties": {
                        "return_debug": {
                          "description": "是否返回 debug（dynamic_params、warnings 等）。默认 false",
                          "type": "boolean"
                        }
                      },
                      "type": "object",
                      "description": "【可选配置】控制接口行为的高级选项\n"
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
            "create_time": 1776920840668983300,
            "update_time": 1776920840668983300,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "bd55b7ef-8804-4690-af9d-b9996bf5d3ea",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "52b35175-cee3-41ea-91c0-1d70e8371f9c",
            "name": "search_schema",
            "description": "统一的 Schema 探索入口。根据 query 返回相关 object_types、relation_types、action_types。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "c8371e6b-7d72-47a0-a64c-b3c1bd2f0018",
              "summary": "search_schema",
              "description": "统一的 Schema 探索入口。适用于还不确定应该继续调用哪个\n`query_*`、`find_*`、`get_*` 工具时，先通过该接口探索相关 schema 概念。\n",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/search_schema",
              "method": "POST",
              "create_time": 1776920840666727400,
              "update_time": 1776920840666727400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
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
                      "enum": [
                        "user",
                        "app",
                        "anonymous"
                      ],
                      "type": "string"
                    }
                  },
                  {
                    "name": "response_format",
                    "in": "query",
                    "description": "响应格式：json 或 toon，默认 json",
                    "required": false,
                    "schema": {
                      "default": "json",
                      "enum": [
                        "json",
                        "toon"
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
                        "$ref": "#/components/schemas/SearchSchemaRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "成功返回 Schema 探索结果",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/SearchSchemaResponse"
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
                    "ErrorResponse": {
                      "properties": {
                        "description": {
                          "type": "string"
                        },
                        "details": {
                          "description": "错误详情"
                        },
                        "link": {
                          "type": "string"
                        },
                        "solution": {
                          "type": "string"
                        },
                        "code": {
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "SearchSchemaRequest": {
                      "type": "object",
                      "required": [
                        "query",
                        "kn_id"
                      ],
                      "properties": {
                        "schema_brief": {
                          "default": false,
                          "type": "boolean",
                          "description": "是否返回精简 Schema，默认返回相对完整的 Schema"
                        },
                        "search_scope": {
                          "$ref": "#/components/schemas/SearchScope"
                        },
                        "enable_rerank": {
                          "description": "是否启用关系类型 Rerank",
                          "default": true,
                          "type": "boolean"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "知识网络ID。HTTP 接口通过 request body 传入。"
                        },
                        "max_concepts": {
                          "type": "integer",
                          "description": "Schema 候选规模上限",
                          "default": 10
                        },
                        "query": {
                          "type": "string",
                          "description": "用户查询问题或关键词"
                        }
                      }
                    },
                    "SearchScope": {
                      "type": "object",
                      "description": "Schema 探索范围。至少需要开启一种概念类型；不传时默认四类全开。\n`concept_groups` 用于按 BKN 概念分组限定 Schema 召回范围，只作用于概念层发现，不作为实例数据过滤条件。\n`search_scope` 仅约束响应输出，不阻断系统内部使用相关线索辅助召回。\n四个资源开关不能同时为 `false`；若同时为 `false`，接口返回参数错误。\n",
                      "properties": {
                        "concept_groups": {
                          "type": "array",
                          "description": "BKN 概念分组 ID 列表，用于限定 object_types、relation_types、action_types 与 metric_types 的 Schema 召回范围；不传或为空数组表示不限定分组。分组语义由 BKN 完成（ContextLoader 直接调用 BKN 分组搜索接口并把列表透传下去）。当传入的分组在该知识网络中不存在时，BKN 当前会返回 5xx 错误（如 'BknBackend.ObjectType.InternalError' 含 'all concept group not found ...'），本工具会直接向上透传该错误而不是返回空结果，调用方据此区分'分组不存在'与'分组合法但范围内无概念'。",
                          "items": {
                            "type": "string"
                          }
                        },
                        "include_metric_types": {
                          "description": "是否包含指标类",
                          "default": true,
                          "type": "boolean"
                        },
                        "include_object_types": {
                          "default": true,
                          "type": "boolean",
                          "description": "是否包含对象类"
                        },
                        "include_relation_types": {
                          "type": "boolean",
                          "description": "是否包含关系类",
                          "default": true
                        },
                        "include_action_types": {
                          "type": "boolean",
                          "description": "是否包含动作类",
                          "default": true
                        }
                      }
                    },
                    "SearchSchemaResponse": {
                      "properties": {
                        "relation_types": {
                          "description": "关系类型列表",
                          "items": {
                            "type": "object"
                          },
                          "type": "array"
                        },
                        "action_types": {
                          "items": {
                            "type": "object"
                          },
                          "type": "array",
                          "description": "动作类型列表"
                        },
                        "metric_types": {
                          "items": {
                            "type": "object"
                          },
                          "type": "array",
                          "description": "指标类型列表"
                        },
                        "object_types": {
                          "description": "对象类型列表",
                          "items": {
                            "type": "object"
                          },
                          "type": "array"
                        }
                      },
                      "type": "object"
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "SchemaSearch"
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
            "create_time": 1776920840668983300,
            "update_time": 1776920840668983300,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "c8371e6b-7d72-47a0-a64c-b3c1bd2f0018",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "5467284c-f25a-4665-8f05-a320dd6e1ec3",
            "name": "get_kn_index_build_status",
            "description": "查询最新50个构建任务的整体状态（按创建时间倒排）。如果所有任务都已完成则返回completed，如果有任务正在运行则返回running",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "168f0ebd-ef5a-49aa-b2f1-dce0850733f8",
              "summary": "get_kn_index_build_status",
              "description": "查询最新50个构建任务的整体状态（按创建时间倒排）。如果所有任务都已完成则返回completed，如果有任务正在运行则返回running",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/full_ontology_building_status",
              "method": "GET",
              "create_time": 1776920840666727400,
              "update_time": 1776920840666727400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
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
                    "ErrorResponse": {
                      "type": "object",
                      "properties": {
                        "code": {
                          "type": "string",
                          "description": "错误码"
                        },
                        "description": {
                          "type": "string",
                          "description": "错误描述"
                        },
                        "detail": {
                          "description": "错误详情",
                          "type": "object"
                        },
                        "link": {
                          "description": "错误链接",
                          "type": "string"
                        },
                        "solution": {
                          "description": "解决方案",
                          "type": "string"
                        }
                      }
                    },
                    "BuildStatusSimpleResponse": {
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
                      "description": "构建状态响应",
                      "required": [
                        "kn_id",
                        "state",
                        "state_detail"
                      ]
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
            "create_time": 1776920840668983300,
            "update_time": 1776920840668983300,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "168f0ebd-ef5a-49aa-b2f1-dce0850733f8",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "3c78c9da-0bd5-48b4-b23b-960972d4d2af",
            "name": "create_kn_index_build_job",
            "description": "创建一个全量构建业务知识网络的任务",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "28462a3e-575c-4ada-b939-dc840d6fcb5f",
              "summary": "create_kn_index_build_job",
              "description": "创建一个全量构建业务知识网络的任务",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/full_build_ontology",
              "method": "POST",
              "create_time": 1776920840666727400,
              "update_time": 1776920840666727400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
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
                    "ErrorResponse": {
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
                          "description": "解决方案",
                          "type": "string"
                        },
                        "code": {
                          "type": "string",
                          "description": "错误码"
                        },
                        "description": {
                          "description": "错误描述",
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "CreateJobRequest": {
                      "required": [
                        "kn_id",
                        "name"
                      ],
                      "properties": {
                        "name": {
                          "description": "任务名称",
                          "type": "string"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "业务知识网络ID"
                        }
                      },
                      "type": "object"
                    },
                    "CreateJobResponse": {
                      "properties": {
                        "id": {
                          "description": "任务ID",
                          "type": "string"
                        }
                      },
                      "type": "object"
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
            "create_time": 1776920840668983300,
            "update_time": 1776920840668983300,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "28462a3e-575c-4ada-b939-dc840d6fcb5f",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "8542b7c2-f82a-4c1e-ab8c-83e2e73ccfdf",
            "name": "query_instance_subgraph",
            "description": "基于预定义的关系路径查询知识图谱中的对象子图。支持多条路径查询，每条路径返回独立子图。对象以map形式返回，支持过滤条件和排序。query_type需设为\"relation_path\"。\r\n",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "859c997b-3558-4feb-a60d-432681994dff",
              "summary": "query_instance_subgraph",
              "description": "基于预定义的关系路径查询知识图谱中的对象子图。支持多条路径查询，每条路径返回独立子图。对象以map形式返回，支持过滤条件和排序。query_type需设为\"relation_path\"。\n",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/query_instance_subgraph",
              "method": "POST",
              "create_time": 1776920840666727400,
              "update_time": 1776920840666727400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
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
                  },
                  {
                    "name": "include_logic_params",
                    "in": "query",
                    "description": "包含逻辑属性的计算参数，默认false，返回结果不包含逻辑属性的字段和值",
                    "required": false,
                    "schema": {
                      "type": "boolean"
                    }
                  },
                  {
                    "name": "response_format",
                    "in": "query",
                    "description": "响应格式：json 或 toon，默认 json",
                    "required": false,
                    "schema": {
                      "default": "json",
                      "enum": [
                        "json",
                        "toon"
                      ],
                      "type": "string"
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
                      "required": [
                        "relations",
                        "length"
                      ],
                      "properties": {
                        "relations": {
                          "items": {
                            "$ref": "#/components/schemas/Relation"
                          },
                          "type": "array",
                          "description": "路径的边集合，沿着路径顺序出现的边"
                        },
                        "length": {
                          "type": "integer",
                          "description": "当前路径的长度"
                        }
                      },
                      "type": "object",
                      "description": "对象的关系路径"
                    },
                    "SubGraphQueryBaseOnTypePath": {
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
                      "type": "object",
                      "description": "查询请求的顶层结构。用于基于关系类路径查询对象子图。relation_type_paths数组中可以包含多条不同的关系路径，系统会同时查询并返回所有路径的结果。每条路径必须符合严格的顺序和方向要求。"
                    },
                    "ObjectTypeOnPath": {
                      "properties": {
                        "sort": {
                          "description": "对当前对象类的排序字段",
                          "items": {
                            "$ref": "#/components/schemas/Sort"
                          },
                          "type": "array"
                        },
                        "condition": {
                          "$ref": "#/components/schemas/Condition"
                        },
                        "id": {
                          "type": "string",
                          "description": "对象类id"
                        },
                        "limit": {
                          "type": "integer",
                          "description": "对象类获取对象数量的限制"
                        }
                      },
                      "type": "object",
                      "description": "路径中的对象类信息",
                      "required": [
                        "id",
                        "condition",
                        "limit"
                      ]
                    },
                    "ObjectSubGraphResponse": {
                      "properties": {
                        "objects": {
                          "type": "object",
                          "description": "子图中的对象map，格式为：\n{\n  \"对象ID1\": {ObjectInfoInSubgraph对象1},\n  \"对象ID2\": {ObjectInfoInSubgraph对象2}\n}\n其中key是ObjectInfoInSubgraph中的id属性，value是完整的ObjectInfoInSubgraph对象。\n动态数据字段，其值可以是基本类型、MetricProperty或OperatorProperty\n"
                        },
                        "relation_paths": {
                          "description": "对象的关系路径集合",
                          "items": {
                            "$ref": "#/components/schemas/RelationPath"
                          },
                          "type": "array"
                        },
                        "search_after": {
                          "description": "表示返回的最后一个起点类对象的排序值，获取这个用于下一次 search_after 分页",
                          "items": {},
                          "type": "array"
                        },
                        "total_count": {
                          "description": "起点对象类的总条数",
                          "type": "integer"
                        }
                      },
                      "type": "object",
                      "description": "对象子图",
                      "required": [
                        "objects",
                        "relation_paths",
                        "total_count",
                        "search_after"
                      ]
                    },
                    "Condition": {
                      "properties": {
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
                        },
                        "sub_conditions": {
                          "items": {
                            "$ref": "#/components/schemas/Condition"
                          },
                          "type": "array",
                          "description": "子过滤条件数组，用于逻辑操作符(and/or)的组合查询"
                        },
                        "value": {
                          "description": "字段值，格式根据操作符类型而定：\n- 比较操作符: 单个值\n- 范围查询: [min, max]数组\n- 集合操作: 值数组\n- 向量搜索: 特定格式数组\n\n**必须与 `value_from: \"const\"` 同时使用**\n",
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
                              "type": "array",
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
                              }
                            }
                          ]
                        },
                        "value_from": {
                          "type": "string",
                          "description": "字段值来源。\n\n**重要：** 当前仅支持 \"const\"（常量值），且必须与 `value` 字段同时使用\n",
                          "enum": [
                            "const"
                          ]
                        }
                      },
                      "type": "object",
                      "description": "过滤条件结构，用于构建对象实例的查询筛选条件。\n\n**重要规则：**\n- `value_from` 和 `value` 必须同时使用，不能单独使用\n- `value_from` 当前仅支持 \"const\"（常量值）\n- 当使用 `value_from: \"const\"` 时，必须同时提供 `value` 字段\n",
                      "required": [
                        "operation"
                      ]
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
                          "description": "关系类id",
                          "type": "string"
                        },
                        "relation_type_name": {
                          "description": "关系类名称",
                          "type": "string"
                        },
                        "source_object_id": {
                          "type": "string",
                          "description": "起点对象id"
                        },
                        "target_object_id": {
                          "description": "终点对象id",
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "TypeEdge": {
                      "type": "object",
                      "description": "路径中的边信息。**方向和顺序极其重要**！通过关系类id确定边，通过路径的起点对象类id和终点对象类id来确定当前路径的方向为正向还是反向，与关系类的起终点一致为正向，相反则为反向。每个TypeEdge必须与路径中的前后对象类型严格对应，这直接影响查询结果的正确性。",
                      "required": [
                        "relation_type_id",
                        "source_object_type_id",
                        "target_object_type_id"
                      ],
                      "properties": {
                        "source_object_type_id": {
                          "type": "string",
                          "description": "路径的起点对象类id"
                        },
                        "target_object_type_id": {
                          "description": "路径的终点对象类id",
                          "type": "string"
                        },
                        "relation_type_id": {
                          "description": "关系类id",
                          "type": "string"
                        }
                      }
                    },
                    "PathEntries": {
                      "required": [
                        "entries"
                      ],
                      "properties": {
                        "entries": {
                          "type": "array",
                          "description": "路径子图",
                          "items": {
                            "$ref": "#/components/schemas/ObjectSubGraphResponse"
                          }
                        }
                      },
                      "type": "object",
                      "description": "路径子图返回体"
                    },
                    "Sort": {
                      "properties": {
                        "direction": {
                          "type": "string",
                          "description": "排序方向",
                          "enum": [
                            "desc",
                            "asc"
                          ]
                        },
                        "field": {
                          "type": "string",
                          "description": "排序字段"
                        }
                      },
                      "type": "object",
                      "description": "排序字段",
                      "required": [
                        "field",
                        "direction"
                      ]
                    },
                    "RelationTypePath": {
                      "required": [
                        "relation_types",
                        "object_types"
                      ],
                      "properties": {
                        "object_types": {
                          "items": {
                            "$ref": "#/components/schemas/ObjectTypeOnPath"
                          },
                          "type": "array",
                          "description": "路径中的对象类集合，**顺序必须严格**与路径中节点出现顺序保持一致。对于n跳路径，object_types数组长度应为n+1，且必须按照source_object_type → 中间节点 → target_object_type的顺序排列。如果某个节点没有过滤条件或者排序或者限制数量，也必须保留其id字段以确保顺序正确。"
                        },
                        "relation_types": {
                          "items": {
                            "$ref": "#/components/schemas/TypeEdge"
                          },
                          "type": "array",
                          "description": "路径的边集合，**顺序必须严格**按照路径中关系出现的顺序排列。对于n跳路径，relation_types数组长度应为n，且必须与object_types数组中的对象类型严格对应：第i个relation_type的source_object_type_id必须等于object_types数组中第i个对象的id，target_object_type_id必须等于object_types数组中第i+1个对象的id。"
                        },
                        "limit": {
                          "description": "当前路径返回的路径数量的限制。",
                          "type": "integer"
                        }
                      },
                      "type": "object",
                      "description": "基于路径获取对象子图。**这是查询的核心结构**！用于定义完整的关系路径查询模板，包括路径中的所有对象类型和关系类型。object_types和relation_types数组的顺序**必须严格对应**，共同构成一个完整的关系路径。"
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
            "create_time": 1776920840668983300,
            "update_time": 1776920840668983300,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "859c997b-3558-4feb-a60d-432681994dff",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "f46fa5df-f371-447f-8451-2d2f34cc78e9",
            "name": "query_object_instance",
            "description": "根据单个对象类查询对象实例，该接口基于业务知识网络语义检索接口返回的对象类定义，查询具体的对象实例数据。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "3c8de1d0-4943-4460-a600-dc7e979a7573",
              "summary": "query_object_instance",
              "description": "根据单个对象类查询对象实例，该接口基于业务知识网络语义检索接口返回的对象类定义，查询具体的对象实例数据。",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/query_object_instance",
              "method": "POST",
              "create_time": 1776920840666727400,
              "update_time": 1776920840666727400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
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
                    "name": "include_logic_params",
                    "in": "query",
                    "description": "包含逻辑属性的计算参数，默认false，返回结果不包含逻辑属性的字段和值",
                    "required": false,
                    "schema": {
                      "type": "boolean"
                    }
                  },
                  {
                    "name": "response_format",
                    "in": "query",
                    "description": "响应格式：json 或 toon，默认 json",
                    "required": false,
                    "schema": {
                      "default": "json",
                      "enum": [
                        "json",
                        "toon"
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
                    "LogicSource": {
                      "properties": {
                        "id": {
                          "type": "string",
                          "description": "数据来源ID"
                        },
                        "name": {
                          "type": "string",
                          "description": "名称。查看详情时返回。"
                        },
                        "type": {
                          "type": "string",
                          "description": "数据来源类型",
                          "enum": [
                            "metric",
                            "operator"
                          ]
                        }
                      },
                      "type": "object",
                      "description": "数据来源",
                      "required": [
                        "type",
                        "id"
                      ]
                    },
                    "FirstQueryWithSearchAfter": {
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
                          "type": "integer",
                          "description": "返回的数量，默认值 10。范围 1-100",
                          "default": 10
                        },
                        "need_total": {
                          "description": "是否需要总数，默认false",
                          "type": "boolean"
                        },
                        "properties": {
                          "items": {
                            "type": "string"
                          },
                          "type": "array",
                          "description": "指定返回的对象属性字段列表，默认返回所有属性。"
                        }
                      },
                      "type": "object",
                      "description": "分页查询的第一次查询请求"
                    },
                    "Parameter4Metric": {
                      "description": "逻辑参数",
                      "required": [
                        "name",
                        "value_from",
                        "operation"
                      ],
                      "properties": {
                        "name": {
                          "type": "string",
                          "description": "参数名称"
                        },
                        "operation": {
                          "description": "操作符。映射指标模型的属性时，此字段必须",
                          "enum": [
                            "in",
                            "=",
                            "!=",
                            ">",
                            ">=",
                            "<",
                            "<="
                          ],
                          "type": "string"
                        },
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
                        }
                      },
                      "type": "object"
                    },
                    "ObjectTypeDetail": {
                      "type": "object",
                      "description": "对象类信息",
                      "properties": {
                        "comment": {
                          "type": "string",
                          "description": "备注（可以为空）"
                        },
                        "create_time": {
                          "format": "int64",
                          "description": "创建时间",
                          "type": "integer"
                        },
                        "display_key": {
                          "description": "对象实例的显示属性",
                          "type": "string"
                        },
                        "data_properties": {
                          "type": "array",
                          "description": "数据属性",
                          "items": {
                            "$ref": "#/components/schemas/DataProperty"
                          }
                        },
                        "primary_keys": {
                          "description": "主键",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        },
                        "kn_id": {
                          "description": "业务知识网络id",
                          "type": "string"
                        },
                        "icon": {
                          "description": "图标",
                          "type": "string"
                        },
                        "tags": {
                          "description": "标签。 （可以为空）",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        },
                        "color": {
                          "description": "颜色",
                          "type": "string"
                        },
                        "detail": {
                          "description": "说明书。按需返回，若指定了include_detail=true，则返回，否则不返回。列表查询时不返回此字段",
                          "type": "string"
                        },
                        "name": {
                          "type": "string",
                          "description": "对象类名称"
                        },
                        "concept_groups": {
                          "type": "array",
                          "description": "概念分组id",
                          "items": {
                            "$ref": "#/components/schemas/ConceptGroup"
                          }
                        },
                        "data_source": {
                          "$ref": "#/components/schemas/DataSource"
                        },
                        "logic_properties": {
                          "description": "逻辑属性",
                          "items": {
                            "$ref": "#/components/schemas/LogicProperty"
                          },
                          "type": "array"
                        },
                        "update_time": {
                          "type": "integer",
                          "format": "int64",
                          "description": "最近一次更新时间"
                        },
                        "branch": {
                          "type": "string",
                          "description": "分支ID"
                        },
                        "creator": {
                          "type": "string",
                          "description": "创建人ID"
                        },
                        "updater": {
                          "description": "最近一次修改人",
                          "type": "string"
                        },
                        "id": {
                          "description": "对象类ID",
                          "type": "string"
                        },
                        "module_type": {
                          "description": "模块类型",
                          "enum": [
                            "object_type"
                          ],
                          "type": "string"
                        }
                      }
                    },
                    "Parameter4Operator": {
                      "required": [
                        "name",
                        "value_from"
                      ],
                      "properties": {
                        "value_from": {
                          "type": "string",
                          "description": "值来源",
                          "enum": [
                            "property",
                            "input"
                          ]
                        },
                        "name": {
                          "type": "string",
                          "description": "参数名称"
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
                          "description": "参数值。value_from=property时，填入的是对象类的数据属性名称；value_from=input时，不设置此字段",
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "description": "逻辑参数"
                    },
                    "Sort": {
                      "type": "object",
                      "description": "排序字段",
                      "required": [
                        "field",
                        "direction"
                      ],
                      "properties": {
                        "direction": {
                          "type": "string",
                          "description": "排序方向",
                          "enum": [
                            "desc",
                            "asc"
                          ]
                        },
                        "field": {
                          "description": "排序字段",
                          "type": "string"
                        }
                      }
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
                          "type": "string",
                          "description": "属性描述"
                        },
                        "display_name": {
                          "type": "string",
                          "description": "属性显示名"
                        },
                        "fulltext_config": {
                          "$ref": "#/components/schemas/FulltextConfig"
                        },
                        "index": {
                          "description": "是否开启索引，默认是true",
                          "type": "boolean"
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
                    "FulltextConfig": {
                      "description": "全文索引的配置",
                      "required": [
                        "analyzer",
                        "field_keyword"
                      ],
                      "properties": {
                        "field_keyword": {
                          "type": "boolean",
                          "description": "是否保留原始字符串，保留原始字符串可用于精确匹配。默认是false"
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
                      "type": "object"
                    },
                    "ObjectDataResponse": {
                      "type": "object",
                      "description": "节点（对象类）信息",
                      "required": [
                        "groups",
                        "type",
                        "datas",
                        "search_after"
                      ],
                      "properties": {
                        "datas": {
                          "description": "对象实例数据。动态数据字段，其值可以是基本类型、MetricProperty或OperatorProperty",
                          "items": {
                            "type": "object"
                          },
                          "type": "array"
                        },
                        "object_type": {
                          "$ref": "#/components/schemas/ObjectTypeDetail"
                        },
                        "search_after": {
                          "type": "array",
                          "description": "表示返回的最后一个文档的排序值，获取这个用于下一次 search_after 分页",
                          "items": {}
                        },
                        "total_count": {
                          "description": "总条数",
                          "type": "integer"
                        }
                      }
                    },
                    "ConceptGroup": {
                      "description": "概念分组",
                      "required": [
                        "id",
                        "name"
                      ],
                      "properties": {
                        "name": {
                          "description": "概念分组名称",
                          "type": "string"
                        },
                        "id": {
                          "description": "概念分组ID",
                          "type": "string"
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
                        "field": {
                          "description": "字段名称，也即对象类的属性名称",
                          "type": "string"
                        },
                        "operation": {
                          "description": "查询条件操作符。\n**注意：** 虽然这里列出了所有可能的操作符，但每个对象类实际支持的操作符列表以对象类定义中的 `condition_operations` 字段为准。\n",
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
                          "type": "string"
                        },
                        "sub_conditions": {
                          "description": "子过滤条件数组，用于逻辑操作符(and/or)的组合查询",
                          "items": {
                            "$ref": "#/components/schemas/Condition"
                          },
                          "type": "array"
                        },
                        "value": {
                          "description": "字段值，格式根据操作符类型而定：\n- 比较操作符: 单个值\n- 范围查询: [min, max]数组\n- 集合操作: 值数组\n- 向量搜索: 特定格式数组\n\n**必须与 `value_from: \"const\"` 同时使用**\n"
                        },
                        "value_from": {
                          "description": "字段值来源。\n\n**重要：** 当前仅支持 \"const\"（常量值），且必须与 `value` 字段同时使用\n",
                          "enum": [
                            "const"
                          ],
                          "type": "string"
                        }
                      }
                    },
                    "ViewField": {
                      "description": "视图字段信息",
                      "required": [
                        "name"
                      ],
                      "properties": {
                        "name": {
                          "description": "字段名称",
                          "type": "string"
                        },
                        "type": {
                          "description": "视图字段类型，查看时有此字段",
                          "type": "string"
                        },
                        "display_name": {
                          "description": "字段显示名.查看时有此字段",
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "LogicProperty": {
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
                          "items": {
                            "$ref": "#/components/schemas/Parameter"
                          },
                          "type": "array",
                          "description": "逻辑所需的参数"
                        },
                        "type": {
                          "type": "string",
                          "description": "属性数据类型。除了视图的字段类型之外，还有 metric、objective、event、trace、log、operator"
                        },
                        "comment": {
                          "description": "属性描述",
                          "type": "string"
                        },
                        "data_source": {
                          "$ref": "#/components/schemas/LogicSource"
                        },
                        "display_name": {
                          "type": "string",
                          "description": "属性显示名"
                        },
                        "index": {
                          "type": "boolean",
                          "description": "是否开启索引，默认是true"
                        }
                      },
                      "type": "object",
                      "description": "逻辑属性"
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
            "create_time": 1776920840668983300,
            "update_time": 1776920840668983300,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "3c8de1d0-4943-4460-a600-dc7e979a7573",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "00929c5f-3375-4ddc-9fb0-c48a24707f39",
            "name": "get_action_info",
            "description": "根据对象实例标识召回关联行动，返回 _dynamic_tools。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "7eb88316-d921-4361-ad08-2c1c812e5448",
              "summary": "get_action_info",
              "description": "根据对象实例标识召回关联行动，返回 _dynamic_tools。",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/get_action_info",
              "method": "POST",
              "create_time": 1776920840666727400,
              "update_time": 1776920840666727400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
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
                      "examples": {
                        "multi_instance_example": {
                          "summary": "多对象实例示例",
                          "value": {
                            "_instance_identities": [
                              {
                                "disease_id": "disease_000001"
                              },
                              {
                                "disease_id": "disease_000002"
                              }
                            ],
                            "at_id": "generate_treatment_plan",
                            "kn_id": "kn_medical"
                          }
                        },
                        "single_instance_example": {
                          "summary": "单对象实例示例",
                          "value": {
                            "_instance_identities": [
                              {
                                "disease_id": "disease_000001"
                              }
                            ],
                            "at_id": "generate_treatment_plan",
                            "kn_id": "kn_medical"
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
                    "status_code": "200",
                    "description": "成功返回动态工具列表",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ActionRecallResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "400",
                    "description": "请求参数错误",
                    "content": {
                      "application/json": {
                        "examples": {
                          "invalid_request": {
                            "value": {
                              "code": "INVALID_REQUEST",
                              "description": "_instance_identities 格式错误"
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
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "ErrorResponse": {
                      "type": "object",
                      "properties": {
                        "description": {
                          "type": "string"
                        },
                        "code": {
                          "type": "string"
                        }
                      }
                    },
                    "ActionRecallRequest": {
                      "required": [
                        "kn_id",
                        "at_id"
                      ],
                      "properties": {
                        "at_id": {
                          "type": "string",
                          "description": "行动类ID（从 Schema 获取）"
                        },
                        "kn_id": {
                          "description": "知识网络ID",
                          "type": "string"
                        },
                        "_instance_identities": {
                          "description": "对象实例标识列表（可选）。每个元素为主键键值对，必须从 query_object_instance 或 query_instance_subgraph 返回的 _instance_identity 字段提取，不可臆造。",
                          "items": {
                            "type": "object"
                          },
                          "type": "array"
                        }
                      },
                      "type": "object"
                    },
                    "ActionRecallResponse": {
                      "type": "object",
                      "required": [
                        "_dynamic_tools"
                      ],
                      "properties": {
                        "_dynamic_tools": {
                          "description": "Function Call 格式的工具列表",
                          "items": {
                            "properties": {
                              "parameters": {
                                "type": "object"
                              },
                              "api_url": {
                                "type": "string"
                              },
                              "description": {
                                "type": "string"
                              },
                              "fixed_params": {
                                "type": "object"
                              },
                              "name": {
                                "type": "string"
                              }
                            },
                            "type": "object"
                          },
                          "type": "array"
                        },
                        "headers": {
                          "type": "object"
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
            "create_time": 1776920840668983300,
            "update_time": 1776920840668983300,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "7eb88316-d921-4361-ad08-2c1c812e5448",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "74498bbd-2bdf-4db3-a4da-a90133c7fb65",
            "name": "find_skills",
            "description": "基于业务上下文召回 Skill 候选列表。\r\n\r\n**召回模式选择规则（自动判断，无需显式指定）：**\r\n\r\n| 请求参数 | 召回模式 |\r\n|---------|---------|\r\n| 仅 kn_id | 网络级（Mode 1） |\r\n| kn_id + object_type_id | 对象类级（Mode 2） |\r\n| kn_id + object_type_id + instance_identities | 实例级（Mode 3） |\r\n\r\n**网络级特殊规则：** 仅当 skills ObjectType 与任意其他 ObjectType 都不存在 RelationType 时才返回结果，\r\n否则返回空列表（需要更精确的 object_type_id 才能召回）。\r\n\r\n**skill_query 说明：** 传入时会对 skills 实例的 name/description 字段追加文本过滤条件（支持 knn/match/like，\r\n优先使用已构建的向量/索引能力），若 BKN 中 skills ObjectType 不存在或元数据获取失败则返回 502。\r\n",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "69c7eb67-6e11-4df0-bd8f-293c0f4c92bc",
              "summary": "find_skills",
              "description": "基于业务上下文召回 Skill 候选列表。\n\n**召回模式选择规则（自动判断，无需显式指定）：**\n\n| 请求参数 | 召回模式 |\n|---------|---------|\n| 仅 kn_id | 网络级（Mode 1） |\n| kn_id + object_type_id | 对象类级（Mode 2） |\n| kn_id + object_type_id + instance_identities | 实例级（Mode 3） |\n\n**网络级特殊规则：** 仅当 skills ObjectType 与任意其他 ObjectType 都不存在 RelationType 时才返回结果，\n否则返回空列表（需要更精确的 object_type_id 才能召回）。\n\n**skill_query 说明：** 传入时会对 skills 实例的 name/description 字段追加文本过滤条件（支持 knn/match/like，\n优先使用已构建的向量/索引能力），若 BKN 中 skills ObjectType 不存在或元数据获取失败则返回 502。\n",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/find_skills",
              "method": "POST",
              "create_time": 1776920840666727400,
              "update_time": 1776920840666727400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户 ID，用于内部服务调用时传递账户信息",
                    "required": false,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型：user（用户）、app（应用）、anonymous（匿名）",
                    "required": false,
                    "schema": {
                      "enum": [
                        "user",
                        "app",
                        "anonymous"
                      ],
                      "type": "string"
                    }
                  },
                  {
                    "name": "response_format",
                    "in": "query",
                    "description": "响应格式：json 或 toon，默认 json",
                    "required": false,
                    "schema": {
                      "default": "json",
                      "enum": [
                        "json",
                        "toon"
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
                        "instance_level": {
                          "summary": "实例级召回（kn_id + object_type_id + instance_identities）",
                          "value": {
                            "instance_identities": [
                              {
                                "contract_id": "C-2024-001"
                              }
                            ],
                            "kn_id": "kn_legal",
                            "object_type_id": "contract",
                            "top_k": 10
                          }
                        },
                        "network_level": {
                          "summary": "网络级召回（仅 kn_id）",
                          "value": {
                            "kn_id": "kn_legal",
                            "top_k": 10
                          }
                        },
                        "object_type_level": {
                          "summary": "对象类级召回（kn_id + object_type_id）",
                          "value": {
                            "kn_id": "kn_legal",
                            "object_type_id": "contract",
                            "top_k": 10
                          }
                        },
                        "with_skill_query": {
                          "summary": "带语义过滤的实例级召回",
                          "value": {
                            "instance_identities": [
                              {
                                "contract_id": "C-2024-001"
                              }
                            ],
                            "kn_id": "kn_legal",
                            "object_type_id": "contract",
                            "skill_query": "合同审查",
                            "top_k": 5
                          }
                        }
                      },
                      "schema": {
                        "$ref": "#/components/schemas/FindSkillsRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "成功返回 Skill 候选列表（可能为空列表）",
                    "content": {
                      "application/json": {
                        "examples": {
                          "empty_result_network_scope": {
                            "summary": "无匹配 Skill（网络级但 skills 有关联）",
                            "value": {
                              "entries": [],
                              "message": "当前知识网络中的 Skill 不是全局生效，网络级范围未返回结果。请补充 object_type_id；如已定位到实例，再补充 instance_identities 后重试。"
                            }
                          },
                          "empty_result_no_binding": {
                            "summary": "无匹配 Skill（对象类未配置 Skill 绑定）",
                            "value": {
                              "entries": [],
                              "message": "当前对象类未配置 Skill 绑定关系，无法在该范围内召回 Skill。请确认该对象类是否已绑定 Skill。"
                            }
                          },
                          "success_with_results": {
                            "summary": "返回 Skill 候选",
                            "value": {
                              "entries": [
                                {
                                  "description": "对合同条款进行全面审查，识别风险点",
                                  "name": "合同审查",
                                  "skill_id": "skill_contract_review"
                                },
                                {
                                  "description": "提取合同中的关键条款和义务",
                                  "name": "关键条款提取",
                                  "skill_id": "skill_clause_extract"
                                }
                              ]
                            }
                          }
                        },
                        "schema": {
                          "$ref": "#/components/schemas/FindSkillsResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "400",
                    "description": "请求参数错误",
                    "content": {
                      "application/json": {
                        "examples": {
                          "instance_without_object_type": {
                            "value": {
                              "code": "INVALID_REQUEST",
                              "description": "instance_identities 不为空时 object_type_id 也必须提供"
                            }
                          },
                          "missing_kn_id": {
                            "value": {
                              "code": "INVALID_REQUEST",
                              "description": "kn_id 为必填字段"
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
                    "description": "上游服务不可用（BKN 或 ontology-query 异常）",
                    "content": {
                      "application/json": {
                        "examples": {
                          "skills_ot_not_found": {
                            "value": {
                              "code": "BAD_GATEWAY",
                              "description": "skill_query requires skills ObjectType (id=skills) but none found in kn_id=kn_legal"
                            }
                          }
                        },
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "FindSkillsRequest": {
                      "required": [
                        "kn_id",
                        "object_type_id"
                      ],
                      "properties": {
                        "skill_query": {
                          "description": "可选的 Skill 语义过滤词，对 skills 实例的 name/description 字段追加文本过滤。\nBKN 已构建向量时使用 knn，已构建全文索引时使用 match，否则使用 like。\n若 skills ObjectType 元数据获取失败则返回 502。\n",
                          "type": "string"
                        },
                        "top_k": {
                          "description": "最多返回的 Skill 数量，默认 10，最大 20",
                          "default": 10,
                          "type": "integer"
                        },
                        "instance_identities": {
                          "type": "array",
                          "description": "对象实例标识列表。每个元素为主键键值对，必须从 query_object_instance 或\nquery_instance_subgraph 返回的 _instance_identity 字段提取，不可臆造。\n传入时 object_type_id 也必须提供，否则返回 400。\n",
                          "items": {
                            "type": "object"
                          }
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "知识网络 ID"
                        },
                        "object_type_id": {
                          "description": "业务对象类型 ID（从 search_schema 返回的概念结果获取）。\n不传时为网络级召回；传入时为对象类级或实例级召回。\n",
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "FindSkillsResponse": {
                      "required": [
                        "entries"
                      ],
                      "properties": {
                        "entries": {
                          "items": {
                            "$ref": "#/components/schemas/SkillItem"
                          },
                          "type": "array",
                          "description": "Skill 候选列表，按命中层级优先级（实例级 > 对象类级 > 网络级）和相关性分数降序排列。\n同优先级同分数时按 skill_id 字典序排列，保证顺序稳定。\n无匹配时返回空数组。\n"
                        },
                        "message": {
                          "type": "string",
                          "description": "空结果说明信息。仅当 entries 为空且接口返回 200 时出现。\n用于向调用方（Agent）解释当前为什么没有结果，以及下一步建议。\nentries 非空时不返回此字段。支持多语言（由请求 X-Language 头决定）。\n"
                        }
                      },
                      "type": "object"
                    },
                    "SkillItem": {
                      "required": [
                        "skill_id",
                        "name"
                      ],
                      "properties": {
                        "skill_id": {
                          "description": "Skill 唯一标识，由 execution-factory 写入 BKN skills ObjectType 时确定",
                          "type": "string"
                        },
                        "description": {
                          "type": "string",
                          "description": "Skill 功能描述（可选，BKN 中无此属性时不返回）"
                        },
                        "name": {
                          "description": "Skill 名称",
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "ErrorResponse": {
                      "properties": {
                        "code": {
                          "type": "string",
                          "description": "错误码"
                        },
                        "description": {
                          "type": "string",
                          "description": "错误详情"
                        }
                      },
                      "type": "object"
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "skill-recall"
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
            "create_time": 1776920840668983300,
            "update_time": 1776920840668983300,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "69c7eb67-6e11-4df0-bd8f-293c0f4c92bc",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          }
        ],
        "create_time": 1776920840665934300,
        "update_time": 1776920840665934300,
        "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
        "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
        "metadata_type": "openapi"
      }
    ]
  }
}
