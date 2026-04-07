---
title: DIP算子服务私有接口协议 v1.0.0
language_tabs:
  - http: HTTP
language_clients:
  - http: ""
toc_footers: []
includes: []
search: true
highlight_theme: darkula
headingLevel: 2

---

<!-- Generator: Widdershins v4.0.1 -->

<h1 id="dip-">DIP算子服务私有接口协议 v1.0.0</h1>

> Scroll down for code samples, example requests and responses. Select a language for code samples from the tabs above or the mobile navigation menu.

DIP算子服务私有接口协议RESTful API集合,用于集群内部服务调用, 无需接口认证信息

Base URLs:

* <a href="http://{{host}}:{{port}}">http://{{host}}:{{port}}</a>

    * **host** - agent-operator-integration, 地址默认为agent-operator-integration 命名空间为 anyshare Default: agent-operator-integration

    * **port** - 默认端口9000 Default: 9000

<h1 id="dip--default">Default</h1>

## 注册算子

> Code samples

```http
POST http://{{host}}:{{port}}/api/agent-operator-integration/internal-v1/operator/register HTTP/1.1

Content-Type: application/json
Accept: application/json

```

`POST /api/agent-operator-integration/internal-v1/operator/register`

注册算子

> Body parameter

```json
{
  "data": "string",
  "operator_metadata_type": "openapi",
  "operator_info": {
    "operator_type": "basic",
    "execution_mode": "sync",
    "category": "other_category",
    "category_name": "string",
    "source": "system"
  },
  "operator_execute_control": {
    "timeout": 0,
    "retry_policy": {
      "max_attempts": 3,
      "initial_delay": 1000,
      "max_delay": 6000,
      "backoff_factor": 2,
      "retry_conditions": {
        "status_code": [
          0
        ],
        "error_codes": [
          "string"
        ]
      }
    }
  },
  "extend_info": {},
  "direct_publish": true,
  "user_token": "string"
}
```

<h3 id="注册算子-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[OperatorRegisterReq](#schemaoperatorregisterreq)|false|none|

> Example responses

> 200 Response

```json
[
  {
    "status": "success",
    "operator_id": "string",
    "version": "string",
    "error": {
      "code": "string",
      "description": "string",
      "detail": {},
      "solution": "string",
      "link": "string"
    }
  }
]
```

<h3 id="注册算子-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|[OperatorRegisterResp](#schemaoperatorregisterresp)|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|参数错误|[Error](#schemaerror)|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|无权限|[Error](#schemaerror)|
|409|[Conflict](https://tools.ietf.org/html/rfc7231#section-6.5.8)|算子已存在|[Error](#schemaerror)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|内部错误|[Error](#schemaerror)|

<aside class="success">
This operation does not require authentication
</aside>

## 更新算子信息

> Code samples

```http
POST http://{{host}}:{{port}}/api/agent-operator-integration/internal-v1/operator/info/update HTTP/1.1

Content-Type: application/json
Accept: application/json

```

`POST /api/agent-operator-integration/internal-v1/operator/info/update`

更新算子信息

> Body parameter

```json
{
  "operator_id": "string",
  "version": "string",
  "data": "string",
  "operator_metadata_type": "openapi",
  "operator_info": {
    "operator_type": "basic",
    "execution_mode": "sync",
    "category": "other_category",
    "category_name": "string",
    "source": "system"
  },
  "operator_execute_control": {
    "timeout": 0,
    "retry_policy": {
      "max_attempts": 3,
      "initial_delay": 1000,
      "max_delay": 6000,
      "backoff_factor": 2,
      "retry_conditions": {
        "status_code": [
          0
        ],
        "error_codes": [
          "string"
        ]
      }
    }
  },
  "extend_info": {},
  "direct_publish": true,
  "user_token": "string"
}
```

<h3 id="更新算子信息-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|object|false|none|

> Example responses

> 200 Response

```json
[
  {
    "status": "success",
    "operator_id": "string",
    "version": "string",
    "error": {
      "code": "string",
      "description": "string",
      "detail": {},
      "solution": "string",
      "link": "string"
    }
  }
]
```

<h3 id="更新算子信息-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|[OperatorRegisterResp](#schemaoperatorregisterresp)|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|参数错误|[Error](#schemaerror)|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|无权限|[Error](#schemaerror)|
|409|[Conflict](https://tools.ietf.org/html/rfc7231#section-6.5.8)|算子已存在|[Error](#schemaerror)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|内部错误|[Error](#schemaerror)|

<aside class="success">
This operation does not require authentication
</aside>

## 获取算子列表

> Code samples

```http
GET http://{{host}}:{{port}}/api/agent-operator-integration/internal-v1/operator/info/list HTTP/1.1

Accept: application/json

```

`GET /api/agent-operator-integration/internal-v1/operator/info/list`

获取算子列表

<h3 id="获取算子列表-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|page|query|integer|false|页码，从1开始，默认为1|
|page_size|query|integer|false|每页数量，默认为10，最大为100，当page_size为-1时，返回所有算子信息|
|sort_by|query|string|false|排序字段，支持按创建时间、更新时间、名称排序|
|sort_order|query|string|false|排序规则|
|operator_id|query|string|false|根据算子ID查询，返回所有版本的算子信息，默认根据更新时间排序，desc|
|user_id|query|string|false|根据创建用户查询,默认根据更新时间排序，desc|
|name|query|string|false|根据算子名称查询，返回所有版本的算子信息，默认根据更新时间排序，desc|
|version|query|string|false|根据算子版本查询，返回指定版本的算子信息，支持模糊查询，默认根据更新时间排序，desc|
|status|query|string|false|根据算子状态查询，返回所有版本的算子信息，默认根据更新时间排序，desc|
|category|query|string|false|根据算子分类查询，返回所有版本的算子信息|

#### Enumerated Values

|Parameter|Value|
|---|---|
|sort_by|create_time|
|sort_by|update_time|
|sort_by|name|
|sort_order|asc|
|sort_order|desc|
|status|unpublish|
|status|published|
|status|offline|
|category|other_category|
|category|data_process|
|category|data_transform|
|category|data_store|
|category|data_analysis|
|category|data_query|
|category|data_extract|
|category|data_split|
|category|model_train|

> Example responses

> 200 Response

```json
{
  "total": 0,
  "page": 0,
  "page_size": 0,
  "total_pages": 0,
  "has_next": true,
  "has_prev": true,
  "data": [
    {
      "name": "string",
      "operator_id": "string",
      "version": "string",
      "status": "unpublish",
      "metadata_type": "openapi",
      "metadata": {
        "openapi": {
          "summary": "string",
          "path": "string",
          "method": "string",
          "description": "string",
          "server_url": [
            "string"
          ],
          "api_spec": {
            "parameters": [
              {
                "name": "string",
                "in": "string",
                "description": "string",
                "required": false,
                "schema": {
                  "type": "string",
                  "format": "int32",
                  "example": "string"
                }
              }
            ],
            "request_body": {
              "description": "string",
              "required": false,
              "content": {
                "property1": {
                  "schema": {
                    "field1": "sample_value",
                    "field2": 123
                  },
                  "example": {
                    "field1": "sample_value",
                    "field2": 123
                  }
                },
                "property2": {
                  "schema": {
                    "field1": "sample_value",
                    "field2": 123
                  },
                  "example": {
                    "field1": "sample_value",
                    "field2": 123
                  }
                }
              }
            },
            "responses": {
              "property1": {
                "description": "string",
                "content": {
                  "property1": {
                    "schema": {},
                    "example": {}
                  },
                  "property2": {
                    "schema": {},
                    "example": {}
                  }
                }
              },
              "property2": {
                "description": "string",
                "content": {
                  "property1": {
                    "schema": {},
                    "example": {}
                  },
                  "property2": {
                    "schema": {},
                    "example": {}
                  }
                }
              }
            },
            "schemas": {
              "property1": {
                "type": "string",
                "format": "int32",
                "example": "string"
              },
              "property2": {
                "type": "string",
                "format": "int32",
                "example": "string"
              }
            },
            "security": [
              {
                "securityScheme": "apiKey"
              }
            ]
          }
        }
      },
      "operator_info": {
        "operator_type": "basic",
        "execution_mode": "sync",
        "category": "other_category",
        "category_name": "string",
        "source": "system"
      },
      "operator_execute_control": {
        "timeout": 0,
        "retry_policy": {
          "max_attempts": 3,
          "initial_delay": 1000,
          "max_delay": 6000,
          "backoff_factor": 2,
          "retry_conditions": {
            "status_code": [
              0
            ],
            "error_codes": [
              "string"
            ]
          }
        }
      },
      "extend_info": {},
      "create_time": 0,
      "update_time": 0,
      "create_user": "string",
      "update_user": "string"
    }
  ]
}
```

<h3 id="获取算子列表-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|[OperatorDataInfoList](#schemaoperatordatainfolist)|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|参数错误|[Error](#schemaerror)|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|算子不存在|[Error](#schemaerror)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|内部错误|[Error](#schemaerror)|

<aside class="success">
This operation does not require authentication
</aside>

## 获取算子信息

> Code samples

```http
GET http://{{host}}:{{port}}/api/agent-operator-integration/internal-v1/operator/info/{operator_id} HTTP/1.1

Accept: application/json

```

`GET /api/agent-operator-integration/internal-v1/operator/info/{operator_id}`

获取算子信息

<h3 id="获取算子信息-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|operator_id|path|string|true|算子ID|
|version|query|string|false|算子版本|

> Example responses

> 200 Response

```json
{
  "name": "string",
  "operator_id": "string",
  "version": "string",
  "status": "unpublish",
  "metadata_type": "openapi",
  "metadata": {
    "openapi": {
      "summary": "string",
      "path": "string",
      "method": "string",
      "description": "string",
      "server_url": [
        "string"
      ],
      "api_spec": {
        "parameters": [
          {
            "name": "string",
            "in": "string",
            "description": "string",
            "required": false,
            "schema": {
              "type": "string",
              "format": "int32",
              "example": "string"
            }
          }
        ],
        "request_body": {
          "description": "string",
          "required": false,
          "content": {
            "property1": {
              "schema": {
                "field1": "sample_value",
                "field2": 123
              },
              "example": {
                "field1": "sample_value",
                "field2": 123
              }
            },
            "property2": {
              "schema": {
                "field1": "sample_value",
                "field2": 123
              },
              "example": {
                "field1": "sample_value",
                "field2": 123
              }
            }
          }
        },
        "responses": {
          "property1": {
            "description": "string",
            "content": {
              "property1": {
                "schema": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                },
                "example": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                }
              },
              "property2": {
                "schema": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                },
                "example": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                }
              }
            }
          },
          "property2": {
            "description": "string",
            "content": {
              "property1": {
                "schema": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                },
                "example": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                }
              },
              "property2": {
                "schema": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                },
                "example": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                }
              }
            }
          }
        },
        "schemas": {
          "property1": {
            "type": "string",
            "format": "int32",
            "example": "string"
          },
          "property2": {
            "type": "string",
            "format": "int32",
            "example": "string"
          }
        },
        "security": [
          {
            "securityScheme": "apiKey"
          }
        ]
      }
    }
  },
  "operator_info": {
    "operator_type": "basic",
    "execution_mode": "sync",
    "category": "other_category",
    "category_name": "string",
    "source": "system"
  },
  "operator_execute_control": {
    "timeout": 0,
    "retry_policy": {
      "max_attempts": 3,
      "initial_delay": 1000,
      "max_delay": 6000,
      "backoff_factor": 2,
      "retry_conditions": {
        "status_code": [
          0
        ],
        "error_codes": [
          "string"
        ]
      }
    }
  },
  "extend_info": {},
  "create_time": 0,
  "update_time": 0,
  "create_user": "string",
  "update_user": "string"
}
```

<h3 id="获取算子信息-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|[OperatorDataInfo](#schemaoperatordatainfo)|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|参数错误|[Error](#schemaerror)|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|算子不存在|[Error](#schemaerror)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|内部错误|[Error](#schemaerror)|

<aside class="success">
This operation does not require authentication
</aside>

## 获取算子分类

> Code samples

```http
GET http://{{host}}:{{port}}/api/agent-operator-integration/internal-v1/operator/category HTTP/1.1

Accept: application/json
Authorization: string

```

`GET /api/agent-operator-integration/internal-v1/operator/category`

获取算子分类列表

<h3 id="获取算子分类-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|Authorization|header|string|true|Bearer token|

> Example responses

> 200 Response

```json
[
  {
    "category_type": "other_category",
    "name": "string"
  }
]
```

<h3 id="获取算子分类-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|Inline|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|资源不存在|[Error](#schemaerror)|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|内部错误|[Error](#schemaerror)|

<h3 id="获取算子分类-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» category_type|string|false|none|分类类型|
|» name|string|false|none|分类名称，支持国际化|

#### Enumerated Values

|Property|Value|
|---|---|
|category_type|other_category|
|category_type|data_process|
|category_type|data_transform|
|category_type|data_store|
|category_type|data_analysis|
|category_type|data_query|
|category_type|data_extract|
|category_type|data_split|
|category_type|model_train|

<aside class="success">
This operation does not require authentication
</aside>

# Schemas

<h2 id="tocS_Error">Error</h2>
<!-- backwards compatibility -->
<a id="schemaerror"></a>
<a id="schema_Error"></a>
<a id="tocSerror"></a>
<a id="tocserror"></a>

```json
{
  "code": "string",
  "description": "string",
  "detail": {},
  "solution": "string",
  "link": "string"
}

```

错误信息

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|code|string|false|none|错误码|
|description|string|false|none|错误信息|
|detail|object|false|none|错误详情|
|solution|string|false|none|错误解决方案|
|link|string|false|none|错误链接|

<h2 id="tocS_OperatorStatusItem">OperatorStatusItem</h2>
<!-- backwards compatibility -->
<a id="schemaoperatorstatusitem"></a>
<a id="schema_OperatorStatusItem"></a>
<a id="tocSoperatorstatusitem"></a>
<a id="tocsoperatorstatusitem"></a>

```json
{
  "operator_id": "op-123456",
  "version": "1.0.0",
  "status": "published"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|operator_id|string|true|none|算子ID|
|version|string|true|none|算子版本|
|status|string|true|none|算子状态|

#### Enumerated Values

|Property|Value|
|---|---|
|status|unpublish|
|status|published|
|status|offline|

<h2 id="tocS_OperatorDataInfoList">OperatorDataInfoList</h2>
<!-- backwards compatibility -->
<a id="schemaoperatordatainfolist"></a>
<a id="schema_OperatorDataInfoList"></a>
<a id="tocSoperatordatainfolist"></a>
<a id="tocsoperatordatainfolist"></a>

```json
{
  "total": 0,
  "page": 0,
  "page_size": 0,
  "total_pages": 0,
  "has_next": true,
  "has_prev": true,
  "data": [
    {
      "name": "string",
      "operator_id": "string",
      "version": "string",
      "status": "unpublish",
      "metadata_type": "openapi",
      "metadata": {
        "openapi": {
          "summary": "string",
          "path": "string",
          "method": "string",
          "description": "string",
          "server_url": [
            "string"
          ],
          "api_spec": {
            "parameters": [
              {
                "name": "string",
                "in": "string",
                "description": "string",
                "required": false,
                "schema": {
                  "type": "string",
                  "format": "int32",
                  "example": "string"
                }
              }
            ],
            "request_body": {
              "description": "string",
              "required": false,
              "content": {
                "property1": {
                  "schema": {
                    "field1": "sample_value",
                    "field2": 123
                  },
                  "example": {
                    "field1": "sample_value",
                    "field2": 123
                  }
                },
                "property2": {
                  "schema": {
                    "field1": "sample_value",
                    "field2": 123
                  },
                  "example": {
                    "field1": "sample_value",
                    "field2": 123
                  }
                }
              }
            },
            "responses": {
              "property1": {
                "description": "string",
                "content": {
                  "property1": {
                    "schema": {},
                    "example": {}
                  },
                  "property2": {
                    "schema": {},
                    "example": {}
                  }
                }
              },
              "property2": {
                "description": "string",
                "content": {
                  "property1": {
                    "schema": {},
                    "example": {}
                  },
                  "property2": {
                    "schema": {},
                    "example": {}
                  }
                }
              }
            },
            "schemas": {
              "property1": {
                "type": "string",
                "format": "int32",
                "example": "string"
              },
              "property2": {
                "type": "string",
                "format": "int32",
                "example": "string"
              }
            },
            "security": [
              {
                "securityScheme": "apiKey"
              }
            ]
          }
        }
      },
      "operator_info": {
        "operator_type": "basic",
        "execution_mode": "sync",
        "category": "other_category",
        "category_name": "string",
        "source": "system"
      },
      "operator_execute_control": {
        "timeout": 0,
        "retry_policy": {
          "max_attempts": 3,
          "initial_delay": 1000,
          "max_delay": 6000,
          "backoff_factor": 2,
          "retry_conditions": {
            "status_code": [
              0
            ],
            "error_codes": [
              "string"
            ]
          }
        }
      },
      "extend_info": {},
      "create_time": 0,
      "update_time": 0,
      "create_user": "string",
      "update_user": "string"
    }
  ]
}

```

算子信息列表

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|total|integer|false|none|总记录数|
|page|integer|false|none|当前页码|
|page_size|integer|false|none|每页记录数|
|total_pages|integer|false|none|总页数|
|has_next|boolean|false|none|是否有下一页|
|has_prev|boolean|false|none|是否有上一页|
|data|[[OperatorDataInfo](#schemaoperatordatainfo)]|false|none|[算子信息]|

<h2 id="tocS_OperatorDataInfoOpenAPI">OperatorDataInfoOpenAPI</h2>
<!-- backwards compatibility -->
<a id="schemaoperatordatainfoopenapi"></a>
<a id="schema_OperatorDataInfoOpenAPI"></a>
<a id="tocSoperatordatainfoopenapi"></a>
<a id="tocsoperatordatainfoopenapi"></a>

```json
{
  "summary": "string",
  "path": "string",
  "method": "string",
  "description": "string",
  "server_url": [
    "string"
  ],
  "api_spec": {
    "parameters": [
      {
        "name": "string",
        "in": "string",
        "description": "string",
        "required": false,
        "schema": {
          "type": "string",
          "format": "int32",
          "example": "string"
        }
      }
    ],
    "request_body": {
      "description": "string",
      "required": false,
      "content": {
        "property1": {
          "schema": {
            "field1": "sample_value",
            "field2": 123
          },
          "example": {
            "field1": "sample_value",
            "field2": 123
          }
        },
        "property2": {
          "schema": {
            "field1": "sample_value",
            "field2": 123
          },
          "example": {
            "field1": "sample_value",
            "field2": 123
          }
        }
      }
    },
    "responses": {
      "property1": {
        "description": "string",
        "content": {
          "property1": {
            "schema": {
              "code": 200,
              "data": {
                "result": "success"
              }
            },
            "example": {
              "code": 200,
              "data": {
                "result": "success"
              }
            }
          },
          "property2": {
            "schema": {
              "code": 200,
              "data": {
                "result": "success"
              }
            },
            "example": {
              "code": 200,
              "data": {
                "result": "success"
              }
            }
          }
        }
      },
      "property2": {
        "description": "string",
        "content": {
          "property1": {
            "schema": {
              "code": 200,
              "data": {
                "result": "success"
              }
            },
            "example": {
              "code": 200,
              "data": {
                "result": "success"
              }
            }
          },
          "property2": {
            "schema": {
              "code": 200,
              "data": {
                "result": "success"
              }
            },
            "example": {
              "code": 200,
              "data": {
                "result": "success"
              }
            }
          }
        }
      }
    },
    "schemas": {
      "property1": {
        "type": "string",
        "format": "int32",
        "example": "string"
      },
      "property2": {
        "type": "string",
        "format": "int32",
        "example": "string"
      }
    },
    "security": [
      {
        "securityScheme": "apiKey"
      }
    ]
  }
}

```

openapi类型元数据

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|summary|string|false|none|摘要|
|path|string|false|none|路径|
|method|string|false|none|方法|
|description|string|false|none|描述|
|server_url|[string]|false|none|none|
|api_spec|object|false|none|none|
|» parameters|[object]|false|none|none|
|»» name|string|false|none|参数名称|
|»» in|string|false|none|参数位置，如path/query/header/cookie|
|»» description|string|false|none|参数描述|
|»» required|boolean|false|none|none|
|»» schema|[ParameterSchema](#schemaparameterschema)|false|none|none|
|» request_body|object|false|none|none|
|»» description|string|false|none|请求体描述|
|»» required|boolean|false|none|none|
|»» content|object|false|none|none|
|»»» **additionalProperties**|object|false|none|none|
|»»»» schema|[ExampleRequest](#schemaexamplerequest)|false|none|none|
|»»»» example|[ExampleRequest](#schemaexamplerequest)|false|none|none|
|» responses|object|false|none|none|
|»» **additionalProperties**|object|false|none|none|
|»»» description|string|false|none|响应描述|
|»»» content|object|false|none|none|
|»»»» **additionalProperties**|object|false|none|none|
|»»»»» schema|[ExampleResponse](#schemaexampleresponse)|false|none|none|
|»»»»» example|[ExampleResponse](#schemaexampleresponse)|false|none|none|
|» schemas|object|false|none|none|
|»» **additionalProperties**|[ParameterSchema](#schemaparameterschema)|false|none|none|
|» security|[object]|false|none|none|
|»» securityScheme|string|false|none|none|

#### Enumerated Values

|Property|Value|
|---|---|
|securityScheme|apiKey|
|securityScheme|http|
|securityScheme|oauth2|

<h2 id="tocS_ParameterSchema">ParameterSchema</h2>
<!-- backwards compatibility -->
<a id="schemaparameterschema"></a>
<a id="schema_ParameterSchema"></a>
<a id="tocSparameterschema"></a>
<a id="tocsparameterschema"></a>

```json
{
  "type": "string",
  "format": "int32",
  "example": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|type|string|false|none|none|
|format|string|false|none|none|
|example|string|false|none|none|

#### Enumerated Values

|Property|Value|
|---|---|
|type|string|
|type|number|
|type|integer|
|type|boolean|
|type|array|
|format|int32|
|format|int64|
|format|float|
|format|double|
|format|byte|

<h2 id="tocS_ExampleRequest">ExampleRequest</h2>
<!-- backwards compatibility -->
<a id="schemaexamplerequest"></a>
<a id="schema_ExampleRequest"></a>
<a id="tocSexamplerequest"></a>
<a id="tocsexamplerequest"></a>

```json
{
  "field1": "sample_value",
  "field2": 123
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|field1|string|false|none|none|
|field2|integer(int32)|false|none|none|

<h2 id="tocS_ExampleResponse">ExampleResponse</h2>
<!-- backwards compatibility -->
<a id="schemaexampleresponse"></a>
<a id="schema_ExampleResponse"></a>
<a id="tocSexampleresponse"></a>
<a id="tocsexampleresponse"></a>

```json
{
  "code": 200,
  "data": {
    "result": "success"
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|code|integer(int32)|false|none|none|
|data|object|false|none|none|
|» result|string|false|none|none|

<h2 id="tocS_OperatorDataInfo">OperatorDataInfo</h2>
<!-- backwards compatibility -->
<a id="schemaoperatordatainfo"></a>
<a id="schema_OperatorDataInfo"></a>
<a id="tocSoperatordatainfo"></a>
<a id="tocsoperatordatainfo"></a>

```json
{
  "name": "string",
  "operator_id": "string",
  "version": "string",
  "status": "unpublish",
  "metadata_type": "openapi",
  "metadata": {
    "openapi": {
      "summary": "string",
      "path": "string",
      "method": "string",
      "description": "string",
      "server_url": [
        "string"
      ],
      "api_spec": {
        "parameters": [
          {
            "name": "string",
            "in": "string",
            "description": "string",
            "required": false,
            "schema": {
              "type": "string",
              "format": "int32",
              "example": "string"
            }
          }
        ],
        "request_body": {
          "description": "string",
          "required": false,
          "content": {
            "property1": {
              "schema": {
                "field1": "sample_value",
                "field2": 123
              },
              "example": {
                "field1": "sample_value",
                "field2": 123
              }
            },
            "property2": {
              "schema": {
                "field1": "sample_value",
                "field2": 123
              },
              "example": {
                "field1": "sample_value",
                "field2": 123
              }
            }
          }
        },
        "responses": {
          "property1": {
            "description": "string",
            "content": {
              "property1": {
                "schema": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                },
                "example": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                }
              },
              "property2": {
                "schema": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                },
                "example": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                }
              }
            }
          },
          "property2": {
            "description": "string",
            "content": {
              "property1": {
                "schema": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                },
                "example": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                }
              },
              "property2": {
                "schema": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                },
                "example": {
                  "code": 200,
                  "data": {
                    "result": "success"
                  }
                }
              }
            }
          }
        },
        "schemas": {
          "property1": {
            "type": "string",
            "format": "int32",
            "example": "string"
          },
          "property2": {
            "type": "string",
            "format": "int32",
            "example": "string"
          }
        },
        "security": [
          {
            "securityScheme": "apiKey"
          }
        ]
      }
    }
  },
  "operator_info": {
    "operator_type": "basic",
    "execution_mode": "sync",
    "category": "other_category",
    "category_name": "string",
    "source": "system"
  },
  "operator_execute_control": {
    "timeout": 0,
    "retry_policy": {
      "max_attempts": 3,
      "initial_delay": 1000,
      "max_delay": 6000,
      "backoff_factor": 2,
      "retry_conditions": {
        "status_code": [
          0
        ],
        "error_codes": [
          "string"
        ]
      }
    }
  },
  "extend_info": {},
  "create_time": 0,
  "update_time": 0,
  "create_user": "string",
  "update_user": "string"
}

```

算子信息

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|name|string|false|none|算子名称|
|operator_id|string|false|none|算子ID|
|version|string|false|none|算子版本|
|status|string|false|none|算子状态|
|metadata_type|string|false|none|算子元数据类型|
|metadata|object|false|none|算子元数据|
|» openapi|[OperatorDataInfoOpenAPI](#schemaoperatordatainfoopenapi)|false|none|openapi类型元数据|
|operator_info|[OperatorInfo](#schemaoperatorinfo)|false|none|算子信息|
|operator_execute_control|[OperatorExecuteControl](#schemaoperatorexecutecontrol)|false|none|算子执行控制|
|extend_info|object|false|none|扩展信息|
|create_time|integer|false|none|创建时间|
|update_time|integer|false|none|更新时间|
|create_user|string|false|none|创建用户|
|update_user|string|false|none|更新用户|

#### Enumerated Values

|Property|Value|
|---|---|
|status|unpublish|
|status|published|
|status|offline|
|metadata_type|openapi|

<h2 id="tocS_OperatorRegisterReq">OperatorRegisterReq</h2>
<!-- backwards compatibility -->
<a id="schemaoperatorregisterreq"></a>
<a id="schema_OperatorRegisterReq"></a>
<a id="tocSoperatorregisterreq"></a>
<a id="tocsoperatorregisterreq"></a>

```json
{
  "data": "string",
  "operator_metadata_type": "openapi",
  "operator_info": {
    "operator_type": "basic",
    "execution_mode": "sync",
    "category": "other_category",
    "category_name": "string",
    "source": "system"
  },
  "operator_execute_control": {
    "timeout": 0,
    "retry_policy": {
      "max_attempts": 3,
      "initial_delay": 1000,
      "max_delay": 6000,
      "backoff_factor": 2,
      "retry_conditions": {
        "status_code": [
          0
        ],
        "error_codes": [
          "string"
        ]
      }
    }
  },
  "extend_info": {},
  "direct_publish": true,
  "user_token": "string"
}

```

算子注册请求

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|data|string|true|none|算子数据|
|operator_metadata_type|string|true|none|算子元数据类型|
|operator_info|[OperatorInfo](#schemaoperatorinfo)|false|none|算子信息|
|operator_execute_control|[OperatorExecuteControl](#schemaoperatorexecutecontrol)|false|none|算子执行控制|
|extend_info|object|false|none|算子拓展信息|
|direct_publish|boolean|false|none|是否直接发布算子，只允许单个算子直接发布|
|user_token|string|false|none|用户token，内部接口传参，用于认证用户身份|

#### Enumerated Values

|Property|Value|
|---|---|
|operator_metadata_type|openapi|

<h2 id="tocS_OperatorRegisterResp">OperatorRegisterResp</h2>
<!-- backwards compatibility -->
<a id="schemaoperatorregisterresp"></a>
<a id="schema_OperatorRegisterResp"></a>
<a id="tocSoperatorregisterresp"></a>
<a id="tocsoperatorregisterresp"></a>

```json
[
  {
    "status": "success",
    "operator_id": "string",
    "version": "string",
    "error": {
      "code": "string",
      "description": "string",
      "detail": {},
      "solution": "string",
      "link": "string"
    }
  }
]

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|status|string|true|none|算子状态|
|operator_id|string|false|none|算子ID|
|version|string|false|none|算子版本|
|error|[Error](#schemaerror)|false|none|错误信息|

#### Enumerated Values

|Property|Value|
|---|---|
|status|success|
|status|failed|

<h2 id="tocS_OperatorInfo">OperatorInfo</h2>
<!-- backwards compatibility -->
<a id="schemaoperatorinfo"></a>
<a id="schema_OperatorInfo"></a>
<a id="tocSoperatorinfo"></a>
<a id="tocsoperatorinfo"></a>

```json
{
  "operator_type": "basic",
  "execution_mode": "sync",
  "category": "other_category",
  "category_name": "string",
  "source": "system"
}

```

算子信息

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|operator_type|string|false|none|算子类型： basic: 基础算子, composite: 组合算子, 默认为basic|
|execution_mode|string|false|none|执行模式：sync: 同步执行, async: 异步执行, 默认为sync|
|category|string|false|none|算子分类 默认为other_category|
|category_name|string|false|none|算子分类名称|
|source|string|false|none|算子来源, 默认为unknown, 可选值为system/unknown, 如果是system不允许修改|

#### Enumerated Values

|Property|Value|
|---|---|
|operator_type|basic|
|operator_type|composite|
|execution_mode|sync|
|execution_mode|async|
|category|other_category|
|category|data_process|
|category|data_transform|
|category|data_store|
|category|data_analysis|
|category|data_query|
|category|data_extract|
|category|data_split|
|category|model_train|
|source|system|
|source|unknown|

<h2 id="tocS_OperatorExecuteControl">OperatorExecuteControl</h2>
<!-- backwards compatibility -->
<a id="schemaoperatorexecutecontrol"></a>
<a id="schema_OperatorExecuteControl"></a>
<a id="tocSoperatorexecutecontrol"></a>
<a id="tocsoperatorexecutecontrol"></a>

```json
{
  "timeout": 0,
  "retry_policy": {
    "max_attempts": 3,
    "initial_delay": 1000,
    "max_delay": 6000,
    "backoff_factor": 2,
    "retry_conditions": {
      "status_code": [
        0
      ],
      "error_codes": [
        "string"
      ]
    }
  }
}

```

算子执行控制

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|timeout|integer|false|none|超时时间（毫秒）|
|retry_policy|object|false|none|重试策略|
|» max_attempts|integer|false|none|最大重试次数|
|» initial_delay|integer|false|none|初始延迟时间(毫秒)|
|» max_delay|integer|false|none|最大延迟时间(毫秒)|
|» backoff_factor|integer|false|none|退避因子|
|» retry_conditions|object|false|none|重试条件|
|»» status_code|[integer]|false|none|HTTP状态码|
|»» error_codes|[string]|false|none|异常类型|

