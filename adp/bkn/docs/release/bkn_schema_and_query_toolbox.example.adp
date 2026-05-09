{
  "toolbox": {
    "configs": [
      {
        "box_id": "a0000001-0000-4000-8000-000000000001",
        "box_name": "bkn_schema_and_query_reads",
        "box_desc": "BKN 元数据读取（对象类 / 关系类 / 行动类）+ ontology-query 实例查询与行动执行。**示例导出**：占位 URL / UUID / `box_id`，导入前替换为环境与网关路由；对齐内部 `/in/v1` 时须成对传递 `x-account-id` / `x-account-type`；业务域用 `x-business-domain`。\n\n**路由前缀**：建模与 Schema 同源双前缀——`/api/bkn-backend/v1` 与 `/api/ontology-manager/v1`（及对应 `/in/v1`）挂载同一 Handler；任选其一或与网关对齐即可。实例与执行走 **ontology-query**，非 bkn-backend。",
        "box_svc_url": "http://gateway-or-bkn-backend-placeholder:8080",
        "status": "draft",
        "category_type": "other_category",
        "category_name": "未分类",
        "is_internal": true,
        "source": "custom",
        "metadata_type": "openapi",
        "tools": [
          {
            "tool_id": "b0000001-0001-4000-8000-000000000011",
            "name": "bkn_list_object_types",
            "description": "列出指定知识网络下的对象类定义（GET）。需 `kn_id`。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "b0000001-0001-4100-8000-000000000011",
              "summary": "bkn_list_object_types",
              "description": "GET `/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/object-types`（等价于 `/api/ontology-manager/in/v1/...`）。支持 offset/limit 等查询参数以服务端为准。",
              "server_url": "http://bkn-backend-placeholder:13014",
              "path": "/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/object-types",
              "method": "GET",
              "api_spec": {
                "parameters": [
                  {
                    "name": "kn_id",
                    "in": "path",
                    "description": "知识网络 UUID。",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-business-domain",
                    "in": "header",
                    "description": "业务域；与 CLI `config`/环境变量同源。",
                    "required": false,
                    "schema": { "type": "string", "default": "bd_public" }
                  },
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "内部链路账号 ID；与 `x-account-type` 成对。",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "内部链路账号类型（如 user / app）。",
                    "required": true,
                    "schema": { "type": "string" }
                  }
                ],
                "request_body": { "description": "", "content": {}, "required": false },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "对象类列表。",
                    "content": { "application/json": { "schema": { "type": "object" } } }
                  }
                ],
                "components": { "schemas": {} }
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
            "resource_object": "tool",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "b0000001-0001-4000-8000-000000000012",
            "name": "bkn_get_object_types",
            "description": "按 path 中的 `ot_ids` 批量获取对象类定义（GET；路由名含 ids，实际常见为单个 id）。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "b0000001-0001-4100-8000-000000000012",
              "summary": "bkn_get_object_types",
              "description": "GET `/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/object-types/{ot_ids}`。",
              "server_url": "http://bkn-backend-placeholder:13014",
              "path": "/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/object-types/{ot_ids}",
              "method": "GET",
              "api_spec": {
                "parameters": [
                  {
                    "name": "kn_id",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "ot_ids",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-business-domain",
                    "in": "header",
                    "required": false,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  }
                ],
                "request_body": { "description": "", "content": {}, "required": false },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "对象类定义。",
                    "content": { "application/json": { "schema": { "type": "object" } } }
                  }
                ],
                "components": { "schemas": {} }
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
            "resource_object": "tool",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "b0000001-0001-4000-8000-000000000013",
            "name": "bkn_list_relation_types",
            "description": "列出关系类定义（GET）。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "b0000001-0001-4100-8000-000000000013",
              "summary": "bkn_list_relation_types",
              "description": "GET `/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/relation-types`。",
              "server_url": "http://bkn-backend-placeholder:13014",
              "path": "/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/relation-types",
              "method": "GET",
              "api_spec": {
                "parameters": [
                  {
                    "name": "kn_id",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-business-domain",
                    "in": "header",
                    "required": false,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  }
                ],
                "request_body": { "description": "", "content": {}, "required": false },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "关系类列表。",
                    "content": { "application/json": { "schema": { "type": "object" } } }
                  }
                ],
                "components": { "schemas": {} }
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
            "resource_object": "tool",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "b0000001-0001-4000-8000-000000000014",
            "name": "bkn_get_relation_types",
            "description": "按 `rt_ids` 获取关系类定义（GET）。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "b0000001-0001-4100-8000-000000000014",
              "summary": "bkn_get_relation_types",
              "description": "GET `/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/relation-types/{rt_ids}`。",
              "server_url": "http://bkn-backend-placeholder:13014",
              "path": "/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/relation-types/{rt_ids}",
              "method": "GET",
              "api_spec": {
                "parameters": [
                  {
                    "name": "kn_id",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "rt_ids",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-business-domain",
                    "in": "header",
                    "required": false,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  }
                ],
                "request_body": { "description": "", "content": {}, "required": false },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "关系类定义。",
                    "content": { "application/json": { "schema": { "type": "object" } } }
                  }
                ],
                "components": { "schemas": {} }
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
            "resource_object": "tool",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "b0000001-0001-4000-8000-000000000015",
            "name": "bkn_list_action_types",
            "description": "列出行动类（Action type）定义（GET）。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "b0000001-0001-4100-8000-000000000015",
              "summary": "bkn_list_action_types",
              "description": "GET `/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/action-types`。CLI `action-type query` 常见为对列表结果做条件过滤，非单独 GraphQL 式查询端点。",
              "server_url": "http://bkn-backend-placeholder:13014",
              "path": "/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/action-types",
              "method": "GET",
              "api_spec": {
                "parameters": [
                  {
                    "name": "kn_id",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-business-domain",
                    "in": "header",
                    "required": false,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  }
                ],
                "request_body": { "description": "", "content": {}, "required": false },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "行动类列表。",
                    "content": { "application/json": { "schema": { "type": "object" } } }
                  }
                ],
                "components": { "schemas": {} }
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
            "resource_object": "tool",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "b0000001-0001-4000-8000-000000000016",
            "name": "bkn_get_action_types",
            "description": "按 `at_ids` 获取行动类定义（GET）。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "b0000001-0001-4100-8000-000000000016",
              "summary": "bkn_get_action_types",
              "description": "GET `/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/action-types/{at_ids}`。",
              "server_url": "http://bkn-backend-placeholder:13014",
              "path": "/api/bkn-backend/in/v1/knowledge-networks/{kn_id}/action-types/{at_ids}",
              "method": "GET",
              "api_spec": {
                "parameters": [
                  {
                    "name": "kn_id",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "at_ids",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-business-domain",
                    "in": "header",
                    "required": false,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  }
                ],
                "request_body": { "description": "", "content": {}, "required": false },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "行动类定义。",
                    "content": { "application/json": { "schema": { "type": "object" } } }
                  }
                ],
                "components": { "schemas": {} }
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
            "resource_object": "tool",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "b0000001-0001-4000-8000-000000000021",
            "name": "ontology_query_objects_in_object_type",
            "description": "按条件分页查询某对象类下的实例（POST body：`conditions` / `logic` / `limit` / `search_after` 等，以服务端为准）。对应 CLI `bkn object-type query`。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "b0000001-0001-4100-8000-000000000021",
              "summary": "ontology_query_objects_in_object_type",
              "description": "POST `/api/ontology-query/in/v1/knowledge-networks/{kn_id}/object-types/{ot_id}`（外部等价 `/api/ontology-query/v1/...`）。**须** `Content-Type: application/json`。",
              "server_url": "http://ontology-query-placeholder:13018",
              "path": "/api/ontology-query/in/v1/knowledge-networks/{kn_id}/object-types/{ot_id}",
              "method": "POST",
              "api_spec": {
                "parameters": [
                  {
                    "name": "kn_id",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "ot_id",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-business-domain",
                    "in": "header",
                    "required": false,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  }
                ],
                "request_body": {
                  "description": "查询 DSL",
                  "required": true,
                  "content": {
                    "application/json": {
                      "schema": {
                        "type": "object",
                        "properties": {
                          "conditions": { "type": "array" },
                          "logic": { "type": "string" },
                          "limit": { "type": "integer" },
                          "search_after": { "type": "array", "items": {} }
                        }
                      }
                    }
                  }
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "实例分页结果。",
                    "content": { "application/json": { "schema": { "type": "object" } } }
                  }
                ],
                "components": { "schemas": {} }
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
            "resource_object": "tool",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          },
          {
            "tool_id": "b0000001-0001-4000-8000-000000000022",
            "name": "ontology_query_execute_action_type",
            "description": "执行行动类（有副作用；执行前需业务确认）。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "b0000001-0001-4100-8000-000000000022",
              "summary": "ontology_query_execute_action_type",
              "description": "POST `/api/ontology-query/in/v1/knowledge-networks/{kn_id}/action-types/{at_id}/execute`。请求体字段名以线上部署为准（常见为 `input` 或 `params`）。",
              "server_url": "http://ontology-query-placeholder:13018",
              "path": "/api/ontology-query/in/v1/knowledge-networks/{kn_id}/action-types/{at_id}/execute",
              "method": "POST",
              "api_spec": {
                "parameters": [
                  {
                    "name": "kn_id",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "at_id",
                    "in": "path",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-business-domain",
                    "in": "header",
                    "required": false,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "required": true,
                    "schema": { "type": "string" }
                  }
                ],
                "request_body": {
                  "description": "执行入参",
                  "required": true,
                  "content": {
                    "application/json": {
                      "schema": {
                        "type": "object",
                        "properties": {
                          "input": { "type": "object" },
                          "params": { "type": "object" }
                        }
                      }
                    }
                  }
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "执行受理或结果。",
                    "content": { "application/json": { "schema": { "type": "object" } } }
                  }
                ],
                "components": { "schemas": {} }
              }
            },
            "use_rule": "执行前向用户确认副作用。",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "resource_object": "tool",
            "source_type": "openapi",
            "script_type": "",
            "code": "",
            "dependencies": [],
            "dependencies_url": ""
          }
        ]
      }
    ]
  }
}
