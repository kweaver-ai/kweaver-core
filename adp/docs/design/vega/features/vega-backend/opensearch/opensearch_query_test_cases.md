# OpenSearch查询接口测试用例

## 接口信息
- **接口路径**: `POST /api/vega-backend/v1/resources/:id/data`
- **请求方法**: POST
- **功能描述**: 查询OpenSearch类型资源的数据

## 请求参数
```json
{
  "offset": 0,
  "limit": 10,
  "sort": [
    {
      "field": "timestamp",
      "direction": "desc"
    }
  ],
  "filter_condition": {
    "operation": "and",
    "sub_conditions": [
      {
        "field": "status",
        "operation": "==",
        "value": "active"
      }
    ]
  },
  "output_fields": ["id", "name", "status", "timestamp"],
  "need_total": true
}
```

## 测试用例

### 1. 条件查询

```json
{
    "filter_condition": {
        "field": "description",
        "operation": "contain",
        "value": [
            "张",
            "李"
        ]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl

{
  "from": 0,
  "query": {
    "bool": {
      "minimum_should_match": 1,
      "should": [
        {
          "match": {
            "description": "张三"
          }
        },
        {
          "match": {
            "description": "李四"
          }
        }
      ]
    }
  },
  "size": 10
}
```




### 2. 排序功能测试

#### 2.1 单字段升序排序
**测试目的**: 验证单字段升序排序是否正常工作

**请求示例**:
```json
{
  "limit": 10,
  "sort": [
    {
      "field": "name",
      "direction": "asc"
    }
  ]
}
```

**预期结果**: 返回按name字段升序排列的前10条记录


#### 2.2 单字段降序排序
**测试目的**: 验证单字段降序排序是否正常工作

**请求示例**:
```json
{
  "limit": 10,
  "sort": [
    {
      "field": "timestamp",
      "direction": "desc"
    }
  ]
}
```

**预期结果**: 返回按timestamp字段降序排列的前10条记录


#### 2.3 多字段排序
**测试目的**: 验证多字段排序是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description",
        "operation": "contain",
        "value": [
            "张三",
            "李四",
            "王五"
        ]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true,
    "sort": [
        {
            "field": "create_time",
            "direction": "asc"
        }
    ]
}

# dsl
{
  "from": 0,
  "query": {
    "bool": {
      "minimum_should_match": 1,
      "should": [
        {
          "match": {
            "description": "张三"
          }
        },
        {
          "match": {
            "description": "李四"
          }
        },
        {
          "match": {
            "description": "王五"
          }
        }
      ]
    }
  },
  "size": 10,
  "sort": [
    {
      "create_time": {
        "order": "asc"
      }
    }
  ]
}

```

**预期结果**: 返回先按status升序，再按timestamp降序排列的前10条记录


### 3. 过滤条件测试

#### 3.1 等于条件 (==)
**测试目的**: 验证等于条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "age",
        "operation": "==",
        "value": 10
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#dsl
{
  "from": 0,
  "query": {
    "term": {
      "age": 10
    }
  },
  "size": 10
}
#result
{
    "entries": [
        {
            "age": 10,
            "_id": "lljz_5wB9J2t5Jnx9d3Y",
            "_score": 1,
            "description": "王五",
            "create_time": "2026-03-18T10:00:00Z"
        }
    ],
    "total_count": 1
}
```




#### 3.2 不等于条件 (!=)
**测试目的**: 验证不等于条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "age",
        "operation": "!=",
        "value": 10
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#dsl
 {
  "from": 0,
  "query": {
    "bool": {
      "must_not": {
        "term": {
          "age": 10
        }
      }
    }
  },
  "size": 10
}
#result
{
    "entries": [
        {
            "create_time": "2026-03-17T10:00:00Z",
            "_id": "kliM-pwB9J2t5Jnxxt2j",
            "_score": 0,
            "description": "张三"
        },
        {
            "_id": "k1jS_5wB9J2t5JnxYN3D",
            "_score": 0,
            "description": "李四",
            "create_time": "2026-03-17T10:00:00Z"
        },
        {
            "_score": 0,
            "description": "李四3",
            "create_time": "2026-03-17T10:00:00Z",
            "_id": "lFjS_5wB9J2t5JnxnN2s"
        },
        {
            "description": "王五",
            "create_time": "2026-03-18T10:00:00Z",
            "_id": "lVjj_5wB9J2t5Jnxrt3C",
            "_score": 0
        }
    ],
    "total_count": 4
}
```




#### 3.3 大于条件 (>)
**测试目的**: 验证大于条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "age",
        "operation": ">",
        "value": 2
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#dsl
 {
  "from": 0,
  "query": {
    "range": {
      "age": {
        "gt": 2
      }
    }
  },
  "size": 10
}
#result
{
    "entries": [
        {
            "create_time": "2026-03-18T10:00:00Z",
            "age": 10,
            "_id": "lljz_5wB9J2t5Jnx9d3Y",
            "_score": 1,
            "description": "王五"
        }
    ],
    "total_count": 1
}
```




#### 3.4 大于等于条件 (>=)
**测试目的**: 验证大于等于条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "age",
        "operation": ">",
        "value": 2
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#dsl
 {
  "from": 0,
  "query": {
    "range": {
      "age": {
        "gt": 2
      }
    }
  },
  "size": 10
}
#result
{
    "entries": [
        {
            "create_time": "2026-03-18T10:00:00Z",
            "age": 10,
            "_id": "lljz_5wB9J2t5Jnx9d3Y",
            "_score": 1,
            "description": "王五"
        }
    ],
    "total_count": 1
}
```


#### 3.5 小于条件 (<)
**测试目的**: 验证小于条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "age",
        "operation": "<",
        "value": 20
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#dsl
{
    "filter_condition": {
        "field": "age",
        "operation": "<",
        "value": 20
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}

#result
{
    "entries": [
        {
            "age": 10,
            "_id": "lljz_5wB9J2t5Jnx9d3Y",
            "_score": 1,
            "description": "王五",
            "create_time": "2026-03-18T10:00:00Z"
        }
    ],
    "total_count": 1
}
```




#### 3.6 小于等于条件 (<=)
**测试目的**: 验证小于等于条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "age",
        "operation": "<",
        "value": 20
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#dsl
{
    "filter_condition": {
        "field": "age",
        "operation": "<",
        "value": 20
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}

#result
{
    "entries": [
        {
            "age": 10,
            "_id": "lljz_5wB9J2t5Jnx9d3Y",
            "_score": 1,
            "description": "王五",
            "create_time": "2026-03-18T10:00:00Z"
        }
    ],
    "total_count": 1
}
```


#### 3.7 IN条件
**测试目的**: 验证IN条件是否正常工作

**请求示例**:

```json
# 如果 description 是text 分词字段，terms 只能匹配分词后的关键词。

{
    "filter_condition": {
        "field": "description.keyword",
        "operation": "in",
        "value": ["王天"]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#dsl
 {
  "from": 0,
  "query": {
    "terms": {
      "description.keyword": [
        "王天"
      ]
    }
  },
  "size": 10
}

#result

{
    "entries": [
        {
            "age": 10,
            "_id": "l1jlA50B9J2t5Jnxht3R",
            "_score": 1,
            "description": "王天",
            "create_time": "2026-04-18T10:00:00Z"
        }
    ],
    "total_count": 1
}
```


#### 3.8 NOT IN条件
**测试目的**: 验证NOT IN条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description.keyword",
        "operation": "not_in",
        "value": ["赵六"]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "bool": {
      "must_not": {
        "terms": {
          "description.keyword": [
            "赵六"
          ]
        }
      }
    }
  },
  "size": 10
}
#result
{
    "entries": [
        {
            "_id": "kliM-pwB9J2t5Jnxxt2j",
            "_score": 0,
            "description": "张三",
            "create_time": "2026-03-17T10:00:00Z"
        },
        {
            "create_time": "2026-03-17T10:00:00Z",
            "_id": "k1jS_5wB9J2t5JnxYN3D",
            "_score": 0,
            "description": "李四"
        },
        {
            "description": "李四3",
            "create_time": "2026-03-17T10:00:00Z",
            "_id": "lFjS_5wB9J2t5JnxnN2s",
            "_score": 0
        },
        {
            "description": "王五",
            "create_time": "2026-03-18T10:00:00Z",
            "_id": "lVjj_5wB9J2t5Jnxrt3C",
            "_score": 0
        },
        {
            "age": 10,
            "_id": "lljz_5wB9J2t5Jnx9d3Y",
            "_score": 0,
            "description": "王五",
            "create_time": "2026-03-18T10:00:00Z"
        },
        {
            "description": "王天",
            "create_time": "2026-04-18T10:00:00Z",
            "age": 10,
            "_id": "l1jlA50B9J2t5Jnxht3R",
            "_score": 0
        }
    ],
    "total_count": 6
}
```


#### 3.9 LIKE条件
**测试目的**: 验证LIKE条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description.keyword",
        "operation": "like",
        "value": "%赵六%"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}

# dsl

{
  "from": 0,
  "query": {
    "regexp": {
      "description.keyword": ".*赵六.*"
    }
  },
  "size": 10
}
# result
{
    "entries": [
        {
            "_score": 1,
            "description": "赵六",
            "create_time": "2025-04-18T10:00:00Z",
            "age": 10,
            "_id": "mFgGCZ0B9J2t5Jnx493b"
        }
    ],
    "total_count": 1
}
```


#### 3.10 NOT LIKE条件
**测试目的**: 验证NOT LIKE条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description.keyword",
        "operation": "not_like",
        "value": "%赵六%"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}

# dsl

 {
  "from": 0,
  "query": {
    "bool": {
      "must_not": {
        "regexp": {
          "description.keyword": ".*赵六.*"
        }
      }
    }
  },
  "size": 10
}

# result

{
    "total_count": 6,
    "entries": [
        {
            "description": "张三",
            "create_time": "2026-03-17T10:00:00Z",
            "_id": "kliM-pwB9J2t5Jnxxt2j",
            "_score": 0
        },
        {
            "_id": "k1jS_5wB9J2t5JnxYN3D",
            "_score": 0,
            "description": "李四",
            "create_time": "2026-03-17T10:00:00Z"
        },
        {
            "create_time": "2026-03-17T10:00:00Z",
            "_id": "lFjS_5wB9J2t5JnxnN2s",
            "_score": 0,
            "description": "李四3"
        },
        {
            "description": "王五",
            "create_time": "2026-03-18T10:00:00Z",
            "_id": "lVjj_5wB9J2t5Jnxrt3C",
            "_score": 0
        },
        {
            "age": 10,
            "_id": "lljz_5wB9J2t5Jnx9d3Y",
            "_score": 0,
            "description": "王五",
            "create_time": "2026-03-18T10:00:00Z"
        },
        {
            "create_time": "2026-04-18T10:00:00Z",
            "age": 10,
            "_id": "l1jlA50B9J2t5Jnxht3R",
            "_score": 0,
            "description": "王天"
        }
    ]
}

```



#### 3.11 CONTAIN条件

**测试目的**: 验证CONTAIN条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description.keyword",
        "operation": "contain",
        "value": [
            "好好学习-false"
        ]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "bool": {
      "minimum_should_match": 1,
      "should": [
        {
          "term": {
            "description.keyword": "好好学习-false"
          }
        }
      ]
    }
  },
  "size": 10
}
# result
{
    "total_count": 1,
    "entries": [
        {
            "description": "好好学习-false",
            "create_time": "2025-04-18T10:00:00Z",
            "age": 12,
            "is_enable": false,
            "_id": "m1jqCZ0B9J2t5Jnxrt2t",
            "_score": 1.3862942
        }
    ]
}
```


#### 3.12 NOT CONTAIN条件
**测试目的**: 验证NOT CONTAIN条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description.keyword",
        "operation": "not_contain",
        "value": [
            "王五"
        ]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "bool": {
      "must_not": [
        {
          "term": {
            "description.keyword": "王五"
          }
        }
      ]
    }
  },
  "size": 10
}
```


#### 3.13 RANGE条件
**测试目的**: 验证RANGE条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "age",
        "operation": "range",
        "value": [
            1,
            10
        ]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "range": {
      "age": {
        "gte": 1,
        "lte": 10
      }
    }
  },
  "size": 10
}
```


#### 3.14 OUT RANGE条件
**测试目的**: 验证OUT RANGE条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "age",
        "operation": "out_range",
        "value": [
            1,
            10
        ]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl

 {
  "from": 0,
  "query": {
    "bool": {
      "minimum_should_match": 1,
      "should": [
        {
          "range": {
            "age": {
              "lt": 1
            }
          }
        },
        {
          "range": {
            "age": {
              "gt": 10
            }
          }
        }
      ]
    }
  },
  "size": 10
}
```


#### 3.15 EXIST条件
**测试目的**: 验证EXIST条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "is_enable",
        "operation": "exist"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#dsl
{
  "from": 0,
  "query": {
    "exists": {
      "field": "is_enable"
    }
  },
  "size": 10
}
```


#### 3.16 NOT EXIST条件
**测试目的**: 验证NOT EXIST条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "age",
        "operation": "not_exist"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#
 {
  "from": 0,
  "query": {
    "bool": {
      "must_not": {
        "exists": {
          "field": "age"
        }
      }
    }
  },
  "size": 10
}
```


#### 3.17 EMPTY条件
**测试目的**: 验证EMPTY条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description",
        "operation": "empty"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#dsl
 {
  "from": 0,
  "query": {
    "bool": {
      "minimum_should_match": 1,
      "should": [
        {
          "term": {
            "description": ""
          }
        },
        {
          "bool": {
            "must_not": {
              "exists": {
                "field": "description"
              }
            }
          }
        }
      ]
    }
  },
  "size": 10
}
```

**预期结果**: 返回description字段为空的前10条记录


#### 3.18 NOT EMPTY条件
**测试目的**: 验证NOT EMPTY条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description",
        "operation": "not_empty"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "bool": {
      "must": [
        {
          "exists": {
            "field": "description"
          }
        },
        {
          "bool": {
            "must_not": {
              "term": {
                "description": ""
              }
            }
          }
        }
      ]
    }
  },
  "size": 10
}
```

**预期结果**: 返回description字段不为空的前10条记录


#### 3.19 REGEX条件
**测试目的**: 验证REGEX条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description.keyword",
        "operation": "regex",
        "value": ".*赵六.*"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "regexp": {
      "description.keyword": ".*赵六.*"
    }
  },
  "size": 10
}

# result

{
    "entries": [
        {
            "_id": "mFgGCZ0B9J2t5Jnx493b",
            "_score": 1,
            "description": "赵六",
            "create_time": "2025-04-18T10:00:00Z",
            "age": 10
        }
    ],
    "total_count": 1
}
```


#### 3.20 MATCH条件
**测试目的**: 验证MATCH条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description",
        "operation": "match",
        "value": "六"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "match": {
      "description": "六"
    }
  },
  "size": 10
}

#result
{
    "entries": [
        {
            "_id": "mFgGCZ0B9J2t5Jnx493b",
            "_score": 1.7209104,
            "description": "赵六",
            "create_time": "2025-04-18T10:00:00Z",
            "age": 10
        }
    ],
    "total_count": 1
}
```




#### 3.21 MATCH_PHRASE条件
**测试目的**: 验证MATCH_PHRASE条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description",
        "operation": "match_phrase",
        "value": "好好学习天天向上"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}

# dsl

 {
  "from": 0,
  "query": {
    "match_phrase": {
      "description": "好好学习天天向上"
    }
  },
  "size": 10
}

# result

{
    "entries": [
        {
            "create_time": "2025-04-18T10:00:00Z",
            "age": 10,
            "_id": "mVgtCZ0B9J2t5Jnxut0L",
            "_score": 7.6983833,
            "description": "好好学习天天向上"
        }
    ],
    "total_count": 1
}
```




#### 3.22 PREFIX条件
**测试目的**: 验证PREFIX条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description.keyword",
        "operation": "prefix",
        "value": "好好学习"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "prefix": {
      "description.keyword": "好好学习"
    }
  },
  "size": 10
}
# result
{
    "entries": [
        {
            "description": "好好学习天天向上",
            "create_time": "2025-04-18T10:00:00Z",
            "age": 10,
            "_id": "mVgtCZ0B9J2t5Jnxut0L",
            "_score": 1
        }
    ],
    "total_count": 1
}
```


#### 3.23 NOT PREFIX条件
**测试目的**: 验证NOT PREFIX条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description.keyword",
        "operation": "not_prefix",
        "value": "王天"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "bool": {
      "must_not": {
        "prefix": {
          "description.keyword": "王天"
        }
      }
    }
  },
  "size": 10
}
```

**预期结果**: 返回name字段不以"test"开头的前10条记录


#### 3.24 NULL条件
**测试目的**: 验证NULL条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "description.keyword",
        "operation": "null",
        "value": "age"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "bool": {
      "must_not": {
        "exists": {
          "field": "description.keyword"
        }
      }
    }
  },
  "size": 10
}
```


#### 3.25 NOT NULL条件
**测试目的**: 验证NOT NULL条件是否正常工作

**请求示例**:
```json
 {
  "from": 0,
  "query": {
    "exists": {
      "field": "description.keyword"
    }
  },
  "size": 10
}
```

**预期结果**: 返回deleted_at字段不为null的前10条记录


#### 3.26 TRUE条件
**测试目的**: 验证TRUE条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "is_enable",
        "operation": "true"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#dsl
 {
  "from": 0,
  "query": {
    "term": {
      "is_enable": true
    }
  },
  "size": 10
}
#result
{
    "entries": [
        {
            "is_enable": true,
            "_id": "mljgCZ0B9J2t5Jnx1N3c",
            "_score": 0.2876821,
            "description": "好好学习",
            "create_time": "2025-04-18T10:00:00Z",
            "age": 1
        }
    ],
    "total_count": 1
}
```




#### 3.27 FALSE条件
**测试目的**: 验证FALSE条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "is_enable",
        "operation": "false"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
#dsl
 {
  "from": 0,
  "query": {
    "term": {
      "is_enable": false
    }
  },
  "size": 10
}
#result
{
    "entries": [
        {
            "_score": 0.6931471,
            "description": "好好学习-false",
            "create_time": "2025-04-18T10:00:00Z",
            "age": 12,
            "is_enable": false,
            "_id": "m1jqCZ0B9J2t5Jnxrt2t"
        }
    ],
    "total_count": 1
}
```


#### 3.28 BEFORE条件
**测试目的**: 验证BEFORE条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "create_time",
        "operation": "before",
        "value": [5,"2026-03-17T10:00:00Z"]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "range": {
      "create_time": {
        "lt": "2026-03-17T05:00:00Z"
      }
    }
  },
  "size": 10
}
```

**预期结果**: 返回created_at字段早于2023-01-01的前10条记录


#### 3.29 CURRENT条件
**测试目的**: 验证CURRENT条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "create_time",
        "operation": "current",
        "value": "month"
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}

# dsl

 {
  "from": 0,
  "query": {
    "range": {
      "create_time": {
        "gte": "2026-03-01T00:00:00+08:00",
        "lt": "2026-04-01T00:00:00+08:00"
      }
    }
  },
  "size": 10
}
```

**预期结果**: 返回created_at字段为当前日期的前10条记录


#### 3.30 BETWEEN条件
**测试目的**: 验证BETWEEN条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "field": "create_time",
        "operation": "between",
        "value": ["2026-03-17T10:00:00Z","2026-03-17T10:00:00Z"]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl
 {
  "from": 0,
  "query": {
    "range": {
      "create_time": {
        "gte": "2026-03-17T10:00:00Z",
        "lte": "2026-03-17T10:00:00Z"
      }
    }
  },
  "size": 10
}
#result
{
    "entries": [
        {
            "description": "张三",
            "create_time": "2026-03-17T10:00:00Z",
            "_id": "kliM-pwB9J2t5Jnxxt2j",
            "_score": 1
        },
        {
            "_id": "k1jS_5wB9J2t5JnxYN3D",
            "_score": 1,
            "description": "李四",
            "create_time": "2026-03-17T10:00:00Z"
        },
        {
            "_id": "lFjS_5wB9J2t5JnxnN2s",
            "_score": 1,
            "description": "李四3",
            "create_time": "2026-03-17T10:00:00Z"
        }
    ],
    "total_count": 3
}
```

**预期结果**: 返回created_at字段在2023年内的前10条记录


### 4. 组合条件测试

#### 4.1 AND组合条件
**测试目的**: 验证AND组合条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "operation": "and",
        "sub_conditions": [
            {
                "field": "is_enable",
                "operation": "true"
            },
            {
                "field": "description",
                "operation": "match_phrase",
                "value": "好好学习"
            }
        ]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}

# dsl

 {
  "from": 0,
  "query": {
    "bool": {
      "must": [
        {
          "term": {
            "is_enable": true
          }
        },
        {
          "match_phrase": {
            "description": "好好学习"
          }
        }
      ]
    }
  },
  "size": 10
}
```




#### 4.2 OR组合条件
**测试目的**: 验证OR组合条件是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "operation": "or",
        "sub_conditions": [
            {
                "field": "is_enable",
                "operation": "true"
            },
            {
                "field": "description",
                "operation": "match_phrase",
                "value": "好好学习天天向上"
            }
        ]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true
}
# dsl

 {
  "from": 0,
  "query": {
    "bool": {
      "minimum_should_match": 1,
      "should": [
        {
          "term": {
            "is_enable": true
          }
        },
        {
          "match_phrase": {
            "description": "好好学习天天向上"
          }
        }
      ]
    }
  },
  "size": 10
}
```


#### 4.3 嵌套组合条件
**测试目的**: 验证嵌套组合条件是否正常工作

**请求示例**:
```json
{
    "limit": 10,
    "filter_condition": {
        "operation": "and",
        "sub_conditions": [
            {
                "field": "is_enable",
                "operation": "true"
            },
            {
                "operation": "or",
                "sub_conditions": [
                    {
                        "field": "is_enable",
                        "operation": "true"
                    },
                    {
                        "field": "description",
                        "operation": "match_phrase",
                        "value": "好好学习天天向上"
                    }
                ]
            }
        ]
    }
}
# dsl
 {
  "from": 0,
  "query": {
    "bool": {
      "must": [
        {
          "term": {
            "is_enable": true
          }
        },
        {
          "bool": {
            "minimum_should_match": 1,
            "should": [
              {
                "term": {
                  "is_enable": true
                }
              },
              {
                "match_phrase": {
                  "description": "好好学习天天向上"
                }
              }
            ]
          }
        }
      ]
    }
  },
  "size": 10
}
```


### 5. 输出字段测试

#### 5.1 指定输出字段
**测试目的**: 验证指定输出字段是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "operation": "and",
        "sub_conditions": [
            {
                "field": "is_enable",
                "operation": "true"
            },
            {
                "field": "description",
                "operation": "match_phrase",
                "value": "好好学习"
            }
        ]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true,
    "output_fields": ["description"]
}
# dsl
 {
  "_source": [
    "description"
  ],
  "from": 0,
  "query": {
    "bool": {
      "must": [
        {
          "term": {
            "is_enable": true
          }
        },
        {
          "match_phrase": {
            "description": "好好学习"
          }
        }
      ]
    }
  },
  "size": 10
}
# result
{
    "entries": [
        {
            "description": "好好学习",
            "_id": "mljgCZ0B9J2t5Jnx1N3c",
            "_score": 4.8486786
        }
    ],
    "total_count": 1
}
```


#### 5.2 包含特殊字段
**测试目的**: 验证包含特殊字段(_score)的输出是否正常工作

**请求示例**:
```json
{
    "filter_condition": {
        "operation": "and",
        "sub_conditions": [
            {
                "field": "is_enable",
                "operation": "true"
            },
            {
                "field": "description",
                "operation": "match_phrase",
                "value": "好好学习"
            }
        ]
    },
    "limit": 10,
    "offset": 0,
    "need_total": true,
    "output_fields": ["description","_score"]
}
# dsl
{
  "_source": [
    "description"
  ],
  "from": 0,
  "query": {
    "bool": {
      "must": [
        {
          "term": {
            "is_enable": true
          }
        },
        {
          "match_phrase": {
            "description": "好好学习"
          }
        }
      ]
    }
  },
  "size": 10,
  "track_scores": true
}
#result
{
    "entries": [
        {
            "description": "好好学习",
            "_id": "mljgCZ0B9J2t5Jnx1N3c",
            "_score": 4.8486786
        }
    ],
    "total_count": 1
}
```
