{
  "toolbox": {
    "configs": [
      {
        "box_id": "1a98d9e8-cfa6-4150-9c54-2ad8445d31a5",
        "box_name": "执行工厂工具集",
        "box_desc": "这是执行工厂提供的工具集",
        "box_svc_url": "http://agent-operator-integration:9000/api/agent-operator-integration/internal-v1",
        "status": "published",
        "category_type": "other_category",
        "category_name": "未分类",
        "is_internal": true,
        "source": "custom",
        "tools": [
          {
            "tool_id": "598b4027-47a8-47d4-ae77-e1efb08478f9",
            "name": "get_operator_schema",
            "description": "获取算子市场指定算子详情",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "22f58434-3b5d-4122-8cb9-aae1e2452161",
              "summary": "获取算子市场指定算子详情",
              "description": "获取算子市场指定算子详情",
              "server_url": "http://agent-operator-integration:9000/api/agent-operator-integration/internal-v1",
              "path": "/operator/market/{operator_id}",
              "method": "GET",
              "create_time": 1770799431798123300,
              "update_time": 1770799431798123300,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "api_spec": {
                "parameters": [
                  {
                    "name": "operator_id",
                    "in": "path",
                    "description": "算子ID",
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
                    "description": "成功",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/OperatorDataInfo"
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
                          "$ref": "#/components/schemas/Error"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "404",
                    "description": "算子不存在",
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
                    "description": "内部错误",
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
                    "OperatorDataInfo": {
                      "type": "object",
                      "description": "算子信息",
                      "properties": {
                        "update_user": {
                          "type": "string",
                          "description": "更新用户"
                        },
                        "name": {
                          "description": "算子名称",
                          "type": "string"
                        },
                        "business_domain_id": {
                          "description": "业务域Id",
                          "type": "string"
                        },
                        "version": {
                          "type": "string",
                          "description": "算子版本"
                        },
                        "create_user": {
                          "type": "string",
                          "description": "创建用户"
                        },
                        "metadata": {
                          "type": "object",
                          "description": "算子元数据",
                          "properties": {
                            "openapi": {
                              "$ref": "#/components/schemas/OperatorDataInfoOpenAPI"
                            }
                          }
                        },
                        "operator_info": {
                          "$ref": "#/components/schemas/OperatorInfo"
                        },
                        "operator_execute_control": {
                          "$ref": "#/components/schemas/OperatorExecuteControl"
                        },
                        "status": {
                          "description": "算子状态",
                          "enum": [
                            "unpublish",
                            "published",
                            "offline",
                            "editing"
                          ],
                          "type": "string"
                        },
                        "metadata_type": {
                          "enum": [
                            "openapi"
                          ],
                          "type": "string",
                          "description": "算子元数据类型"
                        },
                        "is_internal": {
                          "type": "boolean",
                          "description": "是否内部算子"
                        },
                        "release_user": {
                          "description": "发布用户",
                          "type": "string"
                        },
                        "update_time": {
                          "description": "更新时间",
                          "type": "integer"
                        },
                        "create_time": {
                          "description": "创建时间",
                          "type": "integer"
                        },
                        "extend_info": {
                          "type": "object",
                          "description": "扩展信息"
                        },
                        "operator_id": {
                          "type": "string",
                          "description": "算子ID"
                        },
                        "tag": {
                          "description": "语义版本号",
                          "type": "string"
                        },
                        "release_time": {
                          "type": "integer",
                          "description": "发布时间"
                        }
                      }
                    },
                    "OperatorExecuteControl": {
                      "description": "算子执行控制",
                      "properties": {
                        "retry_policy": {
                          "type": "object",
                          "description": "重试策略",
                          "properties": {
                            "initial_delay": {
                              "description": "初始延迟时间(毫秒)",
                              "default": 1000,
                              "type": "integer"
                            },
                            "max_attempts": {
                              "description": "最大重试次数",
                              "default": 3,
                              "type": "integer"
                            },
                            "max_delay": {
                              "type": "integer",
                              "description": "最大延迟时间(毫秒)",
                              "default": 6000
                            },
                            "retry_conditions": {
                              "description": "重试条件",
                              "properties": {
                                "error_codes": {
                                  "type": "array",
                                  "description": "异常类型",
                                  "items": {
                                    "type": "string"
                                  }
                                },
                                "status_code": {
                                  "type": "array",
                                  "description": "HTTP状态码",
                                  "items": {
                                    "type": "integer"
                                  }
                                }
                              },
                              "type": "object"
                            },
                            "backoff_factor": {
                              "type": "integer",
                              "description": "退避因子",
                              "default": 2
                            }
                          }
                        },
                        "timeout": {
                          "type": "integer",
                          "description": "超时时间（毫秒）"
                        }
                      },
                      "type": "object"
                    },
                    "OperatorInfo": {
                      "type": "object",
                      "description": "算子信息",
                      "properties": {
                        "source": {
                          "description": "算子来源, 默认为unknown, 可选值为system/unknown, 如果是system不允许修改",
                          "default": "unknown",
                          "enum": [
                            "system",
                            "unknown"
                          ],
                          "type": "string"
                        },
                        "category": {
                          "type": "string",
                          "description": "算子分类 默认为other_category",
                          "default": "other_category",
                          "enum": [
                            "other_category",
                            "data_process",
                            "data_transform",
                            "data_store",
                            "data_analysis",
                            "data_query",
                            "data_extract",
                            "data_split",
                            "model_train"
                          ]
                        },
                        "category_name": {
                          "description": "算子分类名称",
                          "type": "string"
                        },
                        "execution_mode": {
                          "enum": [
                            "sync",
                            "async"
                          ],
                          "type": "string",
                          "description": "执行模式：sync: 同步执行, async: 异步执行, 默认为sync",
                          "default": "sync"
                        },
                        "is_data_source": {
                          "type": "boolean",
                          "description": "是否为数据源算子",
                          "default": false
                        },
                        "operator_type": {
                          "description": "算子类型： basic: 基础算子, composite: 组合算子, 默认为basic",
                          "default": "basic",
                          "enum": [
                            "basic",
                            "composite"
                          ],
                          "type": "string"
                        }
                      }
                    },
                    "OperatorDataInfoOpenAPI": {
                      "type": "object",
                      "description": "openapi类型元数据",
                      "properties": {
                        "description": {
                          "description": "描述",
                          "type": "string"
                        },
                        "method": {
                          "type": "string",
                          "description": "方法"
                        },
                        "path": {
                          "type": "string",
                          "description": "路径"
                        },
                        "server_url": {
                          "type": "array",
                          "items": {
                            "type": "string",
                            "description": "服务器地址"
                          }
                        },
                        "summary": {
                          "type": "string",
                          "description": "摘要"
                        },
                        "api_spec": {
                          "type": "object",
                          "properties": {
                            "schemas": {
                              "type": "object"
                            },
                            "security": {
                              "type": "array",
                              "items": {
                                "type": "object",
                                "properties": {
                                  "securityScheme": {
                                    "type": "string",
                                    "enum": [
                                      "apiKey",
                                      "http",
                                      "oauth2"
                                    ]
                                  }
                                }
                              }
                            },
                            "parameters": {
                              "items": {
                                "type": "object",
                                "properties": {
                                  "description": {
                                    "type": "string",
                                    "description": "参数描述"
                                  },
                                  "in": {
                                    "type": "string",
                                    "description": "参数位置，如path/query/header/cookie"
                                  },
                                  "name": {
                                    "description": "参数名称",
                                    "type": "string"
                                  },
                                  "required": {
                                    "type": "boolean",
                                    "default": false
                                  },
                                  "schema": {
                                    "$ref": "#/components/schemas/ParameterSchema"
                                  }
                                }
                              },
                              "type": "array"
                            },
                            "request_body": {
                              "properties": {
                                "required": {
                                  "type": "boolean",
                                  "default": false
                                },
                                "content": {
                                  "type": "object"
                                },
                                "description": {
                                  "description": "请求体描述",
                                  "type": "string"
                                }
                              },
                              "type": "object"
                            },
                            "responses": {
                              "type": "object"
                            }
                          }
                        }
                      }
                    },
                    "ParameterSchema": {
                      "properties": {
                        "format": {
                          "type": "string",
                          "enum": [
                            "int32",
                            "int64",
                            "float",
                            "double",
                            "byte"
                          ]
                        },
                        "type": {
                          "type": "string",
                          "enum": [
                            "string",
                            "number",
                            "integer",
                            "boolean",
                            "array"
                          ]
                        },
                        "example": {
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "Error": {
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
                          "description": "错误解决方案",
                          "type": "string"
                        },
                        "code": {
                          "description": "错误码",
                          "type": "string"
                        },
                        "description": {
                          "type": "string",
                          "description": "错误信息"
                        }
                      },
                      "type": "object",
                      "description": "错误信息"
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "算子市场"
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
            "create_time": 1770799431799465500,
            "update_time": 1770799431799465500,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": {},
            "resource_object": "tool",
            "source_id": "22f58434-3b5d-4122-8cb9-aae1e2452161",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          }
        ],
        "create_time": 1770799431797154300,
        "update_time": 1770865284962307800,
        "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
        "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
        "metadata_type": "openapi"
      }
    ]
  }
}