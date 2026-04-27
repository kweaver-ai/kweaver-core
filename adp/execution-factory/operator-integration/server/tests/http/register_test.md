# 接口测试数据

## 注册接口
`POST http://{{host}}:{{port}}/api/agent-operator-integration/v1/operator/register`

**请求参数**
```bash
curl -i -k -XPOST "https://192.168.124.91/api/agent-operator-integration/v1/operator/register" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_Qe9iyXVsca4hQi6vp4tWCXKA7hGuLaxM1lMImL-CFtc.j5mYi45-IOxD-_VAsEhPP1Ktmr8TVRHD4tlBEYMnes8" \
--data-binary @- <<EOF
{
  "data": $(cat /root/go/src/github.com/kweaver-ai/operator-hub/operator-integration/server/tests/file/json/auth.json | jq -Rs .),
  "operator_metadata_type": "openapi"
}
EOF
```
**响应结果**
```json
[
    {
        "status": "success",
        "operator_id": "ee304a62-22bc-427c-a894-4998b8b2c9e5",
        "version": "a79e684f-1d58-40cd-8ce7-2d44499c661a",
        "error": null
    },
    {
        "status": "success",
        "operator_id": "369171c3-668f-4cbe-bbe2-4ecfc2396488",
        "version": "14b4c189-352b-4f4c-95c9-1b85d29bbfa6",
        "error": null
    },
    {
        "status": "success",
        "operator_id": "bbaa2b73-1725-48c3-a911-c737ad867dbe",
        "version": "d130665f-178d-42f8-a81b-8e6c5d2e4380",
        "error": null
    },
    {
        "status": "success",
        "operator_id": "b2101424-3696-4abc-99db-dc5b250b9779",
        "version": "07746507-1868-49ff-89b7-94a12ccb7ab5",
        "error": null
    },
    {
        "status": "success",
        "operator_id": "0ffa5f82-296a-4842-8252-932c3f6e2c40",
        "version": "4430f26c-8610-4d33-85f6-62abed2a86f5",
        "error": null
    },
    {
        "status": "success",
        "operator_id": "6c9a660e-7992-4425-bd7d-b2e75b2de10b",
        "version": "51bec819-111a-4457-89d4-52c41fea59cc",
        "error": null
    },
    {
        "status": "success",
        "operator_id": "b30dfe23-ed1f-471b-b389-2dc282dc7d4c",
        "version": "36d9b384-e361-4dc4-b13c-62be029b6eb9",
        "error": null
    },
    {
        "status": "success",
        "operator_id": "227377b3-1e09-4a8d-a71c-bf947fe50084",
        "version": "e9e5b74a-eccc-42f5-a7ad-de13fd99320c",
        "error": null
    },
    {
        "status": "success",
        "operator_id": "620259c7-cb83-4a63-a9bb-7b435e8e50e2",
        "version": "0b79213c-c308-46cd-90c1-a815358a1f52",
        "error": null
    },
    {
        "status": "success",
        "operator_id": "02f92592-b816-49cb-badb-bd33f29b953a",
        "version": "be762c5e-149c-44b4-a1fe-e5326b14232e",
        "error": null
    },
    {
        "status": "success",
        "operator_id": "3ca07e44-cf0c-4bb9-a802-447ffff29478",
        "version": "c5277e88-3aeb-4aca-acf8-fbe403d54265",
        "error": null
    }
]
```

**请求参数**
> 传参为空时
```bash
curl -i -k -XPOST "https://192.168.124.91/api/agent-operator-integration/v1/operator/register" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_X-gPx9PMuVBFZ0eQCjCy4kykR1SRBG2DrH9JdX68wEM.A1fcM8rOhBt-kCxW4zmLidIWAbOA_g8xuffdF4jZi_Y" \
-d '{}'
```

**响应结果**
```json
// HTTP/2 400
{
    "code": "Public.BadRequest",
    "description": "参数错误",
    "solution": "无",
    "link": "无",
    "details": {
        "error": "data is empty",
        "data": ""
    }
}
```


## 算子分类
`GET http://{{host}}:{{port}}/api/agent-operator-integration/v1/operator/category`

**请求参数**
```bash
curl -i -k -XGET "https://192.168.124.91/api/agent-operator-integration/v1/operator/category" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_91jcXEKqR-diHPNszMaKwgfnbVNhm1FEFzyVvnfJLww.Bz0mZwRY6ni5i3skNOHcBqnu56OlMJgjlkzVF1t322g"
```
**响应结果**

```json
// HTTP/2 200
[
    {
        "category_type": "other_category",
        "name": "其他"
    },
    {
        "category_type": "data_process",
        "name": "数据处理"
    },
    {
        "category_type": "data_transform",
        "name": "数据转换"
    },
    {
        "category_type": "data_store",
        "name": "数据存储"
    },
    {
        "category_type": "data_analysis",
        "name": "数据分析"
    },
    {
        "category_type": "data_query",
        "name": "数据查询"
    },
    {
        "category_type": "data_extract",
        "name": "数据提取"
    },
    {
        "category_type": "data_split",
        "name": "数据分割"
    },
    {
        "category_type": "model_train",
        "name": "模型训练"
    }
]
```



## 获取接口信息
`GET http://{{host}}:{{port}}/api/agent-operator-integration/v1/operator/info/{operator_id}`

**请求参数**

```bash
curl -i -k -XGET https://192.168.124.91/api/agent-operator-integration/v1/operator/info/3ca07e44-cf0c-4bb9-a802-447ffff29478\?version\=c5277e88-3aeb-4aca-acf8-fbe403d54265 \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_PvaeuCVaWFP8zKH3xoLjOvL6WkhKuyn2319ErjyzqVM.AlDDGU-bHSVaufAMdh670xGryQuIk8hIT84MvTIZE_I"
```

**响应结果**
- 成功：200
```json
{
    "name": "客户端账号密码认证",
    "operator_id": "3ca07e44-cf0c-4bb9-a802-447ffff29478",
    "version": "c5277e88-3aeb-4aca-acf8-fbe403d54265",
    "status": "published",
    "metadata_type": "openapi",
    "metadata": {
        "id": 0,
        "version": "c5277e88-3aeb-4aca-acf8-fbe403d54265",
        "summary": "客户端账号密码认证",
        "description": "",
        "server_url": "http://host:9080/api",
        "path": "/authentication/v1/client-account-auth",
        "method": "POST",
        "create_time": 1744805378583105542,
        "update_time": 1744805378583105542,
        "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
        "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
        "is_deleted": false,
        "api_spec": {
            "parameters": [],
            "request_body": {
                "description": "",
                "content": {
                    "application/json": {
                        "$ref": "#/components/schemas/ClientLoginReq"
                    }
                }
            },
            "responses": [
                {
                    "status_code": "401",
                    "description": "未授权",
                    "content": {
                        "application/json": {
                            "$ref": "#/components/schemas/Err"
                        }
                    }
                },
                {
                    "status_code": "500",
                    "description": "服务器错误",
                    "content": {
                        "application/json": {
                            "$ref": "#/components/schemas/Err"
                        }
                    }
                },
                {
                    "status_code": "200",
                    "description": "调用成功",
                    "content": {
                        "application/json": {
                            "$ref": "#/components/schemas/ClientLoginRes"
                        }
                    }
                },
                {
                    "status_code": "400",
                    "description": "非法请求",
                    "content": {
                        "application/json": {
                            "$ref": "#/components/schemas/Err"
                        }
                    }
                }
            ],
            "schemas": {
                "ClientLoginReq": {
                    "type": "object",
                    "required": [
                        "account",
                        "password",
                        "method"
                    ],
                    "properties": {
                        "option": {
                            "$ref": "#/components/schemas/ClientLoginOption"
                        },
                        "password": {
                            "type": "string",
                            "description": "明文"
                        },
                        "account": {
                            "type": "string",
                            "description": "用户登录账号"
                        },
                        "method": {
                            "description": "方法",
                            "enum": [
                                "GET"
                            ],
                            "type": "string"
                        }
                    }
                },
                "ClientLoginOption": {
                    "properties": {
                        "uuid": {
                            "description": "验证码唯一标识",
                            "type": "string"
                        },
                        "vcode": {
                            "type": "string",
                            "description": "验证码字符串/动态密码otp"
                        },
                        "vcodeType": {
                            "type": "integer",
                            "description": "验证码类型"
                        }
                    },
                    "type": "object",
                    "description": "用户登录附带信息"
                },
                "Err": {
                    "type": "object",
                    "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                    "required": [
                        "code",
                        "message",
                        "cause"
                    ],
                    "properties": {
                        "message": {
                            "type": "string",
                            "description": "业务错误信息，与code一一对应。"
                        },
                        "cause": {
                            "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。",
                            "type": "string"
                        },
                        "code": {
                            "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n",
                            "type": "integer",
                            "format": "int64"
                        },
                        "detail": {
                            "type": "object",
                            "description": "错误辅助信息"
                        }
                    }
                },
                "ClientLoginRes": {
                    "type": "object",
                    "required": [
                        "user_id"
                    ],
                    "properties": {
                        "user_id": {
                            "type": "string",
                            "description": "用户ID"
                        }
                    }
                }
            },
            "callbacks": null,
            "security": null,
            "tags": [
                "登录认证"
            ],
            "external_docs": null
        }
    },
    "extend_info": null,
    "operator_info": {
        "operator_type": "basic",
        "execution_mode": "sync",
        "category": "other_category"
    },
    "operator_execute_control": {
        "timeout": 3000,
        "retry_policy": {
            "max_attempts": 3,
            "initial_delay": 1000,
            "backoff_factor": 2,
            "max_delay": 6000,
            "retry_conditions": {
                "status_code": null,
                "error_codes": null
            }
        }
    },
    "create_user": "admin",
    "create_time": 1744805378583725512,
    "update_user": "admin",
    "update_time": 1744805378583725512
}
```

- 失败：404
```json
{
    "code": "Public.NotFound",
    "description": "对象不存在",
    "solution": "无",
    "link": "无",
    "details": {
        "operator_id": "2f024ebb-031e-4fd1-993c-b121b03a64d6",
        "version": "9720478d-3346-4f36-be2a-69376707f3b4"
    }
}
```

## 分页接口
`GET http://{{host}}:{{port}}/api/agent-operator-integration/v1/operator/info/list`

> **需要增加支持根据{name等}排序** : 默认根据更新时间

### 默认分页
**请求参数**
```bash
curl -i -k -XGET "https://192.168.124.91/api/agent-operator-integration/v1/operator/info/list?page=2&page_size=1" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_GNCb0Bu5OcRejQ-WqPrjDFYnFugeuZAqsb5_7TRUDNk.pUg0iSrnbPLfLFk-Mza3X0w4s78i6bk_Ux9JxnGz2Do"
```
**响应结果**
```json
{
    "total": 11,
    "page": 2,
    "page_size": 1,
    "total_pages": 11,
    "has_next": true,
    "has_prev": true,
    "data": [
        {
            "name": "记录审计日志",
            "operator_id": "02f92592-b816-49cb-badb-bd33f29b953a",
            "version": "be762c5e-149c-44b4-a1fe-e5326b14232e",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "be762c5e-149c-44b4-a1fe-e5326b14232e",
                "summary": "记录审计日志",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/audit-log",
                "method": "POST",
                "create_time": 1744805378580585039,
                "update_time": 1744805378580585039,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {
                            "application/json": {
                                "$ref": "#/components/schemas/AuditLogInfo"
                            }
                        }
                    },
                    "responses": [
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "204",
                            "description": "无内容",
                            "content": {}
                        },
                        {
                            "status_code": "400",
                            "description": "非法请求",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "Err": {
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "detail": {
                                    "description": "错误辅助信息",
                                    "type": "object"
                                },
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                },
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "type": "integer",
                                    "format": "int64",
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n"
                                }
                            },
                            "type": "object"
                        },
                        "AuditLogInfo": {
                            "required": [
                                "topic",
                                "message"
                            ],
                            "properties": {
                                "topic": {
                                    "type": "string",
                                    "description": "日志类型"
                                },
                                "message": {
                                    "description": "日志内容",
                                    "type": "object"
                                }
                            },
                            "type": "object"
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "审计日志"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378580935005,
            "update_user": "admin",
            "update_time": 1744805378580935005
        }
    ]
}
```
### 不分页获取全部
```bash
curl -i -k -XGET "http://127.0.0.1:9000/api/agent-operator-integration/v1/operator/info/list?&page_size=-1" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_LxxgXnB-oH2fIXVn5qBckvj9CANwKbC8cINiuyVc4ag.dIbpZEMwFNO0t2-_fgC94SsPoasmgtYPlTB7Z_Uuec8"
```

### 根据状态分页查询
```bash
curl -i -k -XGET "https://192.168.124.91/api/agent-operator-integration/v1/operator/info/list?page=1&status=published" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_nhfx1GZn1WV3CyPpHCZdF7Ac6UZejJ3Sfcde8h0JdOw.yDwDK-AyFghoPkKbMkgHGrf16IUTj0Fk7eclyKWVGUU"
```

**响应结果**
- 成功：200
```json
{
    "total": 11,
    "page": 1,
    "page_size": 10,
    "total_pages": 2,
    "has_next": true,
    "has_prev": false,
    "data": [
        {
            "name": "客户端账号密码认证",
            "operator_id": "3ca07e44-cf0c-4bb9-a802-447ffff29478",
            "version": "c5277e88-3aeb-4aca-acf8-fbe403d54265",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "c5277e88-3aeb-4aca-acf8-fbe403d54265",
                "summary": "客户端账号密码认证",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/client-account-auth",
                "method": "POST",
                "create_time": 1744805378583105542,
                "update_time": 1744805378583105542,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {
                            "application/json": {
                                "$ref": "#/components/schemas/ClientLoginReq"
                            }
                        }
                    },
                    "responses": [
                        {
                            "status_code": "401",
                            "description": "未授权",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "200",
                            "description": "调用成功",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/ClientLoginRes"
                                }
                            }
                        },
                        {
                            "status_code": "400",
                            "description": "非法请求",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "ClientLoginReq": {
                            "required": [
                                "account",
                                "password",
                                "method"
                            ],
                            "properties": {
                                "password": {
                                    "type": "string",
                                    "description": "明文"
                                },
                                "account": {
                                    "type": "string",
                                    "description": "用户登录账号"
                                },
                                "method": {
                                    "type": "string",
                                    "description": "方法",
                                    "enum": [
                                        "GET"
                                    ]
                                },
                                "option": {
                                    "$ref": "#/components/schemas/ClientLoginOption"
                                }
                            },
                            "type": "object"
                        },
                        "ClientLoginOption": {
                            "properties": {
                                "uuid": {
                                    "type": "string",
                                    "description": "验证码唯一标识"
                                },
                                "vcode": {
                                    "type": "string",
                                    "description": "验证码字符串/动态密码otp"
                                },
                                "vcodeType": {
                                    "type": "integer",
                                    "description": "验证码类型"
                                }
                            },
                            "type": "object",
                            "description": "用户登录附带信息"
                        },
                        "Err": {
                            "properties": {
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                },
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n",
                                    "type": "integer",
                                    "format": "int64"
                                },
                                "detail": {
                                    "type": "object",
                                    "description": "错误辅助信息"
                                }
                            },
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ]
                        },
                        "ClientLoginRes": {
                            "type": "object",
                            "required": [
                                "user_id"
                            ],
                            "properties": {
                                "user_id": {
                                    "type": "string",
                                    "description": "用户ID"
                                }
                            }
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "登录认证"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378583725512,
            "update_user": "admin",
            "update_time": 1744805378583725512
        },
        {
            "name": "记录审计日志",
            "operator_id": "02f92592-b816-49cb-badb-bd33f29b953a",
            "version": "be762c5e-149c-44b4-a1fe-e5326b14232e",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "be762c5e-149c-44b4-a1fe-e5326b14232e",
                "summary": "记录审计日志",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/audit-log",
                "method": "POST",
                "create_time": 1744805378580585039,
                "update_time": 1744805378580585039,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {
                            "application/json": {
                                "$ref": "#/components/schemas/AuditLogInfo"
                            }
                        }
                    },
                    "responses": [
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "204",
                            "description": "无内容",
                            "content": {}
                        },
                        {
                            "status_code": "400",
                            "description": "非法请求",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "Err": {
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                },
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "type": "integer",
                                    "format": "int64",
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n"
                                },
                                "detail": {
                                    "description": "错误辅助信息",
                                    "type": "object"
                                }
                            },
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail"
                        },
                        "AuditLogInfo": {
                            "type": "object",
                            "required": [
                                "topic",
                                "message"
                            ],
                            "properties": {
                                "topic": {
                                    "type": "string",
                                    "description": "日志类型"
                                },
                                "message": {
                                    "description": "日志内容",
                                    "type": "object"
                                }
                            }
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "审计日志"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378580935005,
            "update_user": "admin",
            "update_time": 1744805378580935005
        },
        {
            "name": "配置应用账户获取任意用户访问令牌的权限",
            "operator_id": "620259c7-cb83-4a63-a9bb-7b435e8e50e2",
            "version": "0b79213c-c308-46cd-90c1-a815358a1f52",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "0b79213c-c308-46cd-90c1-a815358a1f52",
                "summary": "配置应用账户获取任意用户访问令牌的权限",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/access-token-perm/app/{app_id}",
                "method": "PUT",
                "create_time": 1744805378578023156,
                "update_time": 1744805378578023156,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {}
                    },
                    "responses": [
                        {
                            "status_code": "204",
                            "description": "调用接口成功",
                            "content": {}
                        },
                        {
                            "status_code": "400",
                            "description": "非法请求",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "404",
                            "description": "资源不存在",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "Err": {
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                },
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "type": "integer",
                                    "format": "int64",
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n"
                                },
                                "detail": {
                                    "description": "错误辅助信息",
                                    "type": "object"
                                }
                            },
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail"
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "访问令牌权限管理"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378578422872,
            "update_user": "admin",
            "update_time": 1744805378578422872
        },
        {
            "name": "删除应用账户获取任意用户访问令牌的权限",
            "operator_id": "227377b3-1e09-4a8d-a71c-bf947fe50084",
            "version": "e9e5b74a-eccc-42f5-a7ad-de13fd99320c",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "e9e5b74a-eccc-42f5-a7ad-de13fd99320c",
                "summary": "删除应用账户获取任意用户访问令牌的权限",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/access-token-perm/app/{app_id}",
                "method": "DELETE",
                "create_time": 1744805378575262300,
                "update_time": 1744805378575262300,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {}
                    },
                    "responses": [
                        {
                            "status_code": "400",
                            "description": "非法请求",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "404",
                            "description": "资源不存在",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "204",
                            "description": "调用接口成功",
                            "content": {}
                        }
                    ],
                    "schemas": {
                        "Err": {
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                },
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n",
                                    "type": "integer",
                                    "format": "int64"
                                },
                                "detail": {
                                    "type": "object",
                                    "description": "错误辅助信息"
                                }
                            }
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "访问令牌权限管理"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378575835033,
            "update_user": "admin",
            "update_time": 1744805378575835033
        },
        {
            "name": "获取所有具备获取任意用户访问令牌权限的应用账户",
            "operator_id": "b30dfe23-ed1f-471b-b389-2dc282dc7d4c",
            "version": "36d9b384-e361-4dc4-b13c-62be029b6eb9",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "36d9b384-e361-4dc4-b13c-62be029b6eb9",
                "summary": "获取所有具备获取任意用户访问令牌权限的应用账户",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/access-token-perm/app",
                "method": "GET",
                "create_time": 1744805378572955442,
                "update_time": 1744805378572955442,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {}
                    },
                    "responses": [
                        {
                            "status_code": "200",
                            "description": "接口调用成功",
                            "content": {
                                "application/json": {}
                            }
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "Err": {
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "type": "integer",
                                    "format": "int64",
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n"
                                },
                                "detail": {
                                    "description": "错误辅助信息",
                                    "type": "object"
                                },
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                }
                            }
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "访问令牌权限管理"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378573300796,
            "update_user": "admin",
            "update_time": 1744805378573300796
        },
        {
            "name": "Webhook",
            "operator_id": "6c9a660e-7992-4425-bd7d-b2e75b2de10b",
            "version": "51bec819-111a-4457-89d4-52c41fea59cc",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "51bec819-111a-4457-89d4-52c41fea59cc",
                "summary": "Webhook",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/token-hook",
                "method": "POST",
                "create_time": 1744805378570187810,
                "update_time": 1744805378570187810,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {
                            "application/json": {
                                "$ref": "#/components/schemas/SessionBody"
                            }
                        }
                    },
                    "responses": [
                        {
                            "status_code": "200",
                            "description": "接口调用成功",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/TokenHookClaims"
                                }
                            }
                        },
                        {
                            "status_code": "204",
                            "description": "调用接口成功",
                            "content": {}
                        },
                        {
                            "status_code": "400",
                            "description": "非法请求",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "SessionBody": {
                            "properties": {
                                "session": {
                                    "type": "object",
                                    "properties": {
                                        "allowed_top_level_claims": {
                                            "type": "array",
                                            "items": {}
                                        },
                                        "client_id": {
                                            "type": "string"
                                        },
                                        "consent_challenge": {
                                            "type": "string"
                                        },
                                        "exclude_not_before_claim": {
                                            "type": "boolean"
                                        },
                                        "extra": {
                                            "type": "object"
                                        },
                                        "id_token": {
                                            "properties": {
                                                "headers": {
                                                    "type": "object",
                                                    "properties": {
                                                        "extra": {
                                                            "type": "object"
                                                        }
                                                    }
                                                },
                                                "id_token_claims": {
                                                    "properties": {
                                                        "jti": {
                                                            "type": "string"
                                                        },
                                                        "sub": {
                                                            "type": "string"
                                                        },
                                                        "nonce": {
                                                            "type": "string"
                                                        },
                                                        "iss": {
                                                            "type": "string"
                                                        },
                                                        "amr": {},
                                                        "at_hash": {
                                                            "type": "string"
                                                        },
                                                        "c_hash": {
                                                            "type": "string"
                                                        },
                                                        "ext": {
                                                            "type": "object"
                                                        },
                                                        "acr": {
                                                            "type": "string"
                                                        },
                                                        "aud": {
                                                            "type": "array",
                                                            "items": {
                                                                "type": "string"
                                                            }
                                                        }
                                                    },
                                                    "type": "object"
                                                },
                                                "subject": {
                                                    "type": "string"
                                                },
                                                "username": {
                                                    "type": "string"
                                                }
                                            },
                                            "type": "object"
                                        }
                                    }
                                },
                                "request": {
                                    "type": "object",
                                    "properties": {
                                        "grant_types": {
                                            "type": "array",
                                            "items": {
                                                "type": "string"
                                            }
                                        },
                                        "granted_audience": {
                                            "type": "array",
                                            "items": {
                                                "type": "string"
                                            }
                                        },
                                        "granted_scopes": {
                                            "type": "array",
                                            "items": {
                                                "type": "string"
                                            }
                                        },
                                        "payload": {
                                            "properties": {
                                                "assertion": {
                                                    "items": {
                                                        "type": "string"
                                                    },
                                                    "type": "array",
                                                    "description": "只有在jwt-bearer授权类型时，assertion为必需字段"
                                                }
                                            },
                                            "type": "object"
                                        },
                                        "client_id": {
                                            "type": "string"
                                        }
                                    }
                                }
                            },
                            "type": "object"
                        },
                        "TokenHookClaims": {
                            "type": "object",
                            "properties": {
                                "session": {
                                    "type": "object",
                                    "properties": {
                                        "access_token": {
                                            "type": "object"
                                        },
                                        "id_token": {
                                            "type": "object"
                                        }
                                    }
                                }
                            }
                        },
                        "Err": {
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "type": "integer",
                                    "format": "int64",
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n"
                                },
                                "detail": {
                                    "type": "object",
                                    "description": "错误辅助信息"
                                },
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                }
                            }
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "Webhook"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378570656174,
            "update_user": "admin",
            "update_time": 1744805378570656174
        },
        {
            "name": "保存session记录",
            "operator_id": "0ffa5f82-296a-4842-8252-932c3f6e2c40",
            "version": "4430f26c-8610-4d33-85f6-62abed2a86f5",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "4430f26c-8610-4d33-85f6-62abed2a86f5",
                "summary": "保存session记录",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/session/{id}",
                "method": "PUT",
                "create_time": 1744805378567181311,
                "update_time": 1744805378567181311,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {
                            "application/json": {
                                "$ref": "#/components/schemas/PutSessionReq"
                            }
                        }
                    },
                    "responses": [
                        {
                            "status_code": "201",
                            "description": "接口调用成功",
                            "content": {}
                        },
                        {
                            "status_code": "400",
                            "description": "非法请求",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "404",
                            "description": "资源不存在",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "PutSessionReq": {
                            "properties": {
                                "remember_for": {
                                    "description": "设置应记住的的认证时间（以秒为单位）。如果设置为0，将在浏览器会话期间（使用cookie）记住该授权。",
                                    "type": "integer",
                                    "format": "int64"
                                },
                                "subject": {
                                    "type": "string",
                                    "description": "经过身份验证的用户id"
                                },
                                "client_id": {
                                    "type": "string",
                                    "description": "客户端ID"
                                },
                                "context": {
                                    "$ref": "#/components/schemas/Context"
                                }
                            },
                            "type": "object",
                            "required": [
                                "subject",
                                "client_id",
                                "remember_for",
                                "context"
                            ]
                        },
                        "Context": {
                            "properties": {
                                "property1": {
                                    "type": "object"
                                },
                                "property2": {
                                    "type": "object"
                                }
                            },
                            "type": "object",
                            "description": "context是一个可选对象，可以保存任意数据。在“context”字段下获取授权请求时，数据将可用。这在登录和授权端点共享数据的情况下很有用。"
                        },
                        "Err": {
                            "properties": {
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                },
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n",
                                    "type": "integer",
                                    "format": "int64"
                                },
                                "detail": {
                                    "type": "object",
                                    "description": "错误辅助信息"
                                }
                            },
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ]
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "session管理"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378567584515,
            "update_user": "admin",
            "update_time": 1744805378567584515
        },
        {
            "name": "获取session记录",
            "operator_id": "b2101424-3696-4abc-99db-dc5b250b9779",
            "version": "07746507-1868-49ff-89b7-94a12ccb7ab5",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "07746507-1868-49ff-89b7-94a12ccb7ab5",
                "summary": "获取session记录",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/session/{id}",
                "method": "GET",
                "create_time": 1744805378564398471,
                "update_time": 1744805378564398471,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {}
                    },
                    "responses": [
                        {
                            "status_code": "200",
                            "description": "接口调用成功",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/GetSessionRes"
                                }
                            }
                        },
                        {
                            "status_code": "404",
                            "description": "资源不存在",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "GetSessionRes": {
                            "type": "object",
                            "required": [
                                "subject",
                                "client_id",
                                "session_id",
                                "context"
                            ],
                            "properties": {
                                "context": {
                                    "$ref": "#/components/schemas/Context"
                                },
                                "session_id": {
                                    "description": "session ID"
                                },
                                "subject": {
                                    "description": "经过身份验证的用户ID"
                                },
                                "client_id": {
                                    "description": "客户端ID"
                                }
                            }
                        },
                        "Context": {
                            "type": "object",
                            "description": "context是一个可选对象，可以保存任意数据。在“context”字段下获取授权请求时，数据将可用。这在登录和授权端点共享数据的情况下很有用。",
                            "properties": {
                                "property1": {
                                    "type": "object"
                                },
                                "property2": {
                                    "type": "object"
                                }
                            }
                        },
                        "Err": {
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                },
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "format": "int64",
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n",
                                    "type": "integer"
                                },
                                "detail": {
                                    "type": "object",
                                    "description": "错误辅助信息"
                                }
                            },
                            "type": "object"
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "session管理"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378564783896,
            "update_user": "admin",
            "update_time": 1744805378564783896
        },
        {
            "name": "删除session记录",
            "operator_id": "bbaa2b73-1725-48c3-a911-c737ad867dbe",
            "version": "d130665f-178d-42f8-a81b-8e6c5d2e4380",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "d130665f-178d-42f8-a81b-8e6c5d2e4380",
                "summary": "删除session记录",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/session/{id}",
                "method": "DELETE",
                "create_time": 1744805378561941565,
                "update_time": 1744805378561941565,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {}
                    },
                    "responses": [
                        {
                            "status_code": "204",
                            "description": "无内容",
                            "content": {}
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "Err": {
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "type": "integer",
                                    "format": "int64",
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n"
                                },
                                "detail": {
                                    "description": "错误辅助信息",
                                    "type": "object"
                                },
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                }
                            }
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "session管理"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378562310839,
            "update_user": "admin",
            "update_time": 1744805378562310839
        },
        {
            "name": "设置认证配置",
            "operator_id": "369171c3-668f-4cbe-bbe2-4ecfc2396488",
            "version": "14b4c189-352b-4f4c-95c9-1b85d29bbfa6",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "14b4c189-352b-4f4c-95c9-1b85d29bbfa6",
                "summary": "设置认证配置",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/config/{fields}",
                "method": "PUT",
                "create_time": 1744805378559204272,
                "update_time": 1744805378559204272,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {
                            "application/json": {
                                "$ref": "#/components/schemas/Configs"
                            }
                        }
                    },
                    "responses": [
                        {
                            "status_code": "204",
                            "description": "无内容",
                            "content": {}
                        },
                        {
                            "status_code": "400",
                            "description": "非法请求",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "Configs": {
                            "type": "object",
                            "properties": {
                                "anonymous_sms_expiration": {
                                    "type": "integer",
                                    "format": "int64",
                                    "description": "匿名登录短信验证码过期时间，单位分钟"
                                },
                                "remember_for": {
                                    "description": "设置应记住的的认证时间（以秒为单位）。值为非负整数，如果设置为0，将在浏览器会话期间（使用cookie）记住该授权。",
                                    "type": "integer",
                                    "format": "int64"
                                },
                                "remember_visible": {
                                    "type": "boolean",
                                    "description": "记住登录状态按钮是否可见，默认为true。"
                                }
                            }
                        },
                        "Err": {
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "detail": {
                                    "type": "object",
                                    "description": "错误辅助信息"
                                },
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                },
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n",
                                    "type": "integer",
                                    "format": "int64"
                                }
                            }
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "配置管理"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378559656429,
            "update_user": "admin",
            "update_time": 1744805378559656429
        }
    ]
}
```
- 失败：400
```json
{
    "code": "Public.BadRequest",
    "description": "参数错误",
    "solution": "无",
    "link": "无",
    "details": {
        "status": "ssss",
        "error": "invalid status"
    }
}
```


### 根据operator_id分页查询

**请求参数**

```bash
curl -i -k -XGET "https://192.168.124.91/api/agent-operator-integration/v1/operator/info/list?page=1&page_size=10&operator_id=3ca07e44-cf0c-4bb9-a802-447ffff29478" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_nhfx1GZn1WV3CyPpHCZdF7Ac6UZejJ3Sfcde8h0JdOw.yDwDK-AyFghoPkKbMkgHGrf16IUTj0Fk7eclyKWVGUU"
```

**响应结果**
- 查询到结果
```json
{
    "total": 1,
    "page": 1,
    "page_size": 10,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false,
    "data": [
        {
            "name": "客户端账号密码认证",
            "operator_id": "3ca07e44-cf0c-4bb9-a802-447ffff29478",
            "version": "c5277e88-3aeb-4aca-acf8-fbe403d54265",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "c5277e88-3aeb-4aca-acf8-fbe403d54265",
                "summary": "客户端账号密码认证",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/client-account-auth",
                "method": "POST",
                "create_time": 1744805378583105542,
                "update_time": 1744805378583105542,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {
                            "application/json": {
                                "$ref": "#/components/schemas/ClientLoginReq"
                            }
                        }
                    },
                    "responses": [
                        {
                            "status_code": "401",
                            "description": "未授权",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "200",
                            "description": "调用成功",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/ClientLoginRes"
                                }
                            }
                        },
                        {
                            "status_code": "400",
                            "description": "非法请求",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "ClientLoginReq": {
                            "type": "object",
                            "required": [
                                "account",
                                "password",
                                "method"
                            ],
                            "properties": {
                                "method": {
                                    "type": "string",
                                    "description": "方法",
                                    "enum": [
                                        "GET"
                                    ]
                                },
                                "option": {
                                    "$ref": "#/components/schemas/ClientLoginOption"
                                },
                                "password": {
                                    "type": "string",
                                    "description": "明文"
                                },
                                "account": {
                                    "type": "string",
                                    "description": "用户登录账号"
                                }
                            }
                        },
                        "ClientLoginOption": {
                            "description": "用户登录附带信息",
                            "properties": {
                                "vcodeType": {
                                    "type": "integer",
                                    "description": "验证码类型"
                                },
                                "uuid": {
                                    "type": "string",
                                    "description": "验证码唯一标识"
                                },
                                "vcode": {
                                    "type": "string",
                                    "description": "验证码字符串/动态密码otp"
                                }
                            },
                            "type": "object"
                        },
                        "Err": {
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                },
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "format": "int64",
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n",
                                    "type": "integer"
                                },
                                "detail": {
                                    "type": "object",
                                    "description": "错误辅助信息"
                                }
                            }
                        },
                        "ClientLoginRes": {
                            "properties": {
                                "user_id": {
                                    "type": "string",
                                    "description": "用户ID"
                                }
                            },
                            "type": "object",
                            "required": [
                                "user_id"
                            ]
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "登录认证"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378583725512,
            "update_user": "admin",
            "update_time": 1744805378583725512
        }
    ]
}
```
- 无结果
```json
{
    "total": 0,
    "page": 1,
    "page_size": 10,
    "total_pages": 0,
    "has_next": false,
    "has_prev": false,
    "data": []
}
```

### 根据名字模糊查询
**请求参数**

```bash
curl -i -k -XGET "https://192.168.124.91/api/agent-operator-integration/v1/operator/info/list?page=1&page_size=10&name=删除" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_BzEXSNR5UmvjGG0MdKjc-qPatMWlIlVpaq1PsnsMt9k.NcdHjKJp1L5R54jC7DURlbaOt1XT4mdurIROIA6h74g"
```
**响应结果**
```json
{
    "total": 2,
    "page": 1,
    "page_size": 10,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false,
    "data": [
        {
            "name": "删除应用账户获取任意用户访问令牌的权限",
            "operator_id": "227377b3-1e09-4a8d-a71c-bf947fe50084",
            "version": "e9e5b74a-eccc-42f5-a7ad-de13fd99320c",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "e9e5b74a-eccc-42f5-a7ad-de13fd99320c",
                "summary": "删除应用账户获取任意用户访问令牌的权限",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/access-token-perm/app/{app_id}",
                "method": "DELETE",
                "create_time": 1744805378575262300,
                "update_time": 1744805378575262300,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {}
                    },
                    "responses": [
                        {
                            "status_code": "400",
                            "description": "非法请求",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "404",
                            "description": "资源不存在",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "204",
                            "description": "调用接口成功",
                            "content": {}
                        }
                    ],
                    "schemas": {
                        "Err": {
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail",
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "code": {
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n",
                                    "type": "integer",
                                    "format": "int64"
                                },
                                "detail": {
                                    "type": "object",
                                    "description": "错误辅助信息"
                                },
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                },
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                }
                            }
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "访问令牌权限管理"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378575835033,
            "update_user": "admin",
            "update_time": 1744805378575835033
        },
        {
            "name": "删除session记录",
            "operator_id": "bbaa2b73-1725-48c3-a911-c737ad867dbe",
            "version": "d130665f-178d-42f8-a81b-8e6c5d2e4380",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "d130665f-178d-42f8-a81b-8e6c5d2e4380",
                "summary": "删除session记录",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/session/{id}",
                "method": "DELETE",
                "create_time": 1744805378561941565,
                "update_time": 1744805378561941565,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {}
                    },
                    "responses": [
                        {
                            "status_code": "204",
                            "description": "无内容",
                            "content": {}
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "Err": {
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "type": "integer",
                                    "format": "int64",
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n"
                                },
                                "detail": {
                                    "type": "object",
                                    "description": "错误辅助信息"
                                },
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                }
                            },
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail"
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "session管理"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378562310839,
            "update_user": "admin",
            "update_time": 1744805378562310839
        }
    ]
}
```
### 根据分类查询

**请求参数**

```bash
curl -i -k -XGET "https://192.168.124.91/api/agent-operator-integration/v1/operator/info/list?page=1&page_size=1&category=other_category" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_BzEXSNR5UmvjGG0MdKjc-qPatMWlIlVpaq1PsnsMt9k.NcdHjKJp1L5R54jC7DURlbaOt1XT4mdurIROIA6h74g"
```

**响应结果**

```json
{
    "total": 11,
    "page": 1,
    "page_size": 1,
    "total_pages": 11,
    "has_next": true,
    "has_prev": false,
    "data": [
        {
            "name": "客户端账号密码认证",
            "operator_id": "3ca07e44-cf0c-4bb9-a802-447ffff29478",
            "version": "c5277e88-3aeb-4aca-acf8-fbe403d54265",
            "status": "published",
            "metadata_type": "openapi",
            "metadata": {
                "id": 0,
                "version": "c5277e88-3aeb-4aca-acf8-fbe403d54265",
                "summary": "客户端账号密码认证",
                "description": "",
                "server_url": "http://host:9080/api",
                "path": "/authentication/v1/client-account-auth",
                "method": "POST",
                "create_time": 1744805378583105542,
                "update_time": 1744805378583105542,
                "create_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "update_user": "266c6a42-6131-4d62-8f39-853e7093701c",
                "is_deleted": false,
                "api_spec": {
                    "parameters": [],
                    "request_body": {
                        "description": "",
                        "content": {
                            "application/json": {
                                "$ref": "#/components/schemas/ClientLoginReq"
                            }
                        }
                    },
                    "responses": [
                        {
                            "status_code": "401",
                            "description": "未授权",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "500",
                            "description": "服务器错误",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        },
                        {
                            "status_code": "200",
                            "description": "调用成功",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/ClientLoginRes"
                                }
                            }
                        },
                        {
                            "status_code": "400",
                            "description": "非法请求",
                            "content": {
                                "application/json": {
                                    "$ref": "#/components/schemas/Err"
                                }
                            }
                        }
                    ],
                    "schemas": {
                        "ClientLoginReq": {
                            "type": "object",
                            "required": [
                                "account",
                                "password",
                                "method"
                            ],
                            "properties": {
                                "method": {
                                    "type": "string",
                                    "description": "方法",
                                    "enum": [
                                        "GET"
                                    ]
                                },
                                "option": {
                                    "$ref": "#/components/schemas/ClientLoginOption"
                                },
                                "password": {
                                    "type": "string",
                                    "description": "明文"
                                },
                                "account": {
                                    "description": "用户登录账号",
                                    "type": "string"
                                }
                            }
                        },
                        "ClientLoginOption": {
                            "type": "object",
                            "description": "用户登录附带信息",
                            "properties": {
                                "uuid": {
                                    "type": "string",
                                    "description": "验证码唯一标识"
                                },
                                "vcode": {
                                    "type": "string",
                                    "description": "验证码字符串/动态密码otp"
                                },
                                "vcodeType": {
                                    "type": "integer",
                                    "description": "验证码类型"
                                }
                            }
                        },
                        "Err": {
                            "required": [
                                "code",
                                "message",
                                "cause"
                            ],
                            "properties": {
                                "message": {
                                    "type": "string",
                                    "description": "业务错误信息，与code一一对应。"
                                },
                                "cause": {
                                    "type": "string",
                                    "description": "导致此错误的原因，也可用于给接口调用者提示解决此错误的办法，相同的code可对应不同的cause。"
                                },
                                "code": {
                                    "description": "业务错误码，前三位为 HTTP标准状态码，中间三位为系统内全局唯一的微服务错误码标识号，后三位为自定义状态码。\n",
                                    "type": "integer",
                                    "format": "int64"
                                },
                                "detail": {
                                    "type": "object",
                                    "description": "错误辅助信息"
                                }
                            },
                            "type": "object",
                            "description": "接口调用错误信息结构基类，具体错误情况可查看新增的字段detail"
                        },
                        "ClientLoginRes": {
                            "properties": {
                                "user_id": {
                                    "type": "string",
                                    "description": "用户ID"
                                }
                            },
                            "type": "object",
                            "required": [
                                "user_id"
                            ]
                        }
                    },
                    "callbacks": null,
                    "security": null,
                    "tags": [
                        "登录认证"
                    ],
                    "external_docs": null
                }
            },
            "extend_info": null,
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync",
                "category": "other_category"
            },
            "operator_execute_control": {
                "timeout": 3000,
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_delay": 1000,
                    "backoff_factor": 2,
                    "max_delay": 6000,
                    "retry_conditions": {
                        "status_code": null,
                        "error_codes": null
                    }
                }
            },
            "create_user": "admin",
            "create_time": 1744805378583725512,
            "update_user": "admin",
            "update_time": 1744805378583725512
        }
    ]
}
```
### 根据数据来源
[x] 待补充