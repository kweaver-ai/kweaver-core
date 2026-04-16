# Vega Logic View logic_definition 设计

## 1. 设计背景
在低代码数据建模场景中，需要一套 DSL（领域特定语言）来描述数据从源表到最终输出的流转逻辑。Vega 采用“**配置极简化，运行时标准化**”的设计哲学，支持 Join、Union、SQL、Resource 等多种算子。

## 2. 核心架构设计

### 2.1 节点通用结构
所有逻辑节点遵循统一的 JSON 结构，将“转换参数”与“输出协议”分离：
- **`id` / `type` / `name`**: 节点基础信息。
- **`inputs`**: 来源节点 ID 列表，定义了图的拓扑结构。
- **`config`**: **私有配置**。存放算子特有的执行参数（如 Join 条件、SQL 语句）。
- **`output_fields`**: **公开协议**。定义该节点向外输出哪些字段，支持多态缩写。

### 2.2 output_fields 的五种形态
为了平衡用户操作的便捷性与逻辑的严谨性，`output_fields` 数组支持以下格式：
1. **通配符模式**：`["*"]` —— 全量透传上游字段，或由后端自动推断(SQL节点)。
2. **投影模式**：`["field_a", "field_b"]` —— 字符串数组，仅选择字段，原样输出。
3. **映射模式 (Join)**：`{"name": "target", "from": "src", "from_node": "node_a"}` —— 处理字段重命名及冲突。
4. **对齐模式 (Union)**：`{"name": "target", "from": [{"from": "f1", "from_node": "node_a"}, {"from": "f2", "from_node": "node_b"}]}` —— 按索引顺序对齐多个输入源。
5. **定义模式 (SQL)**：`{"name": "target", "type": "string"}` —— 显式定义字段属性。

---

## 3. 场景配置示例

### 场景 A：数据关联 (Join)
通过 `from` 属性实现跨节点的字段精准引用。
```json
{
  "id": "node_join_001",
  "type": "join",
  "inputs": ["node_source_A", "node_source_B"],
  "config": {
    "join_type": "left",
    "join_on": [{ "left_field": "a_id", "operator": "=", "right_field": "b_id" }]
  },
  "output_fields": [
    { "name": "user_name", "from": "name", "from_node": "node_source_A" },
    { "name": "order_price", "from": "price", "from_node": "node_source_B" }
  ]
}
```

### 场景 B：数据合并 (Union)
通过 `sources` 数组实现多源字段的索引位对齐。
```json
{
  "id": "node_union_001",
  "type": "union",
  "inputs": ["node_A", "node_B"],
  "config": { "union_type": "all" },
  "output_fields": [
    {
      "name": "total_qty",
      "from": [
        { "from": "qty_a", "from_node": "node_A" },
        { "from": "qty_b", "from_node": "node_B" }
      ]
    }
  ]
}
```

### 场景 C：高级 SQL 与自动推断
当 SQL 节点配置为 `["*"]` 时，后端将触发字段自动推断逻辑。
```json
{
  "id": "node_sql_001",
  "type": "sql",
  "inputs": ["node_A"],
  "config": { "sql": "SELECT *, price * 0.9 AS discount_price FROM {{.node_A}}" },
  "output_fields": ["*"]
}
```

---

## 4. 运行时元数据回写 (Runtime)

### 4.1 回写机制
后端在执行查询前，会遍历所有节点并生成 `runtime_output_fields`。此字段包含完整的元数据信息，结构参考：

```Go
type Property struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	DisplayName  string            `json:"display_name"`
	OriginalName string            `json:"original_name"`
	Description  string            `json:"description"`
	Features     []PropertyFeature `json:"features,omitempty"`
}
```

### 4.2 收益分析
- 前端友好：下游节点配置时，可直接读取上游的 `runtime_output_fields` 作为下拉选项。
- SQL 准确：生成 SQL 时不再需要递归寻找物理表，直接根据运行时定义的 `from`（字段映射或对齐）字段生成别名。

---

## 5. 接口交互与运行时逻辑

### 5.1 元数据推断流程
后端在实际执行查询或提供预览前，会进行**元数据补全**：
1. **标准化**：将所有缩写形式（如 `["*"]`）转换为对象数组格式。
2. **递归溯源**：沿着 `inputs` 链路向上查找字段的原始 `type`、`description` 和 `original_name`。
3. **回写运行时字段**：将推断结果写入 `runtime_output_fields`。

### 5.2 运行时字段结构 (Property)
补全后的每个字段对象将包含：
- `name`: 内部标识名。
- `display_name`: 前端显示的标签。
- `type`: 数据类型（string, decimal, integer, etc.）。
- `description`: 业务说明。
- `features`: 字段特征（如精确匹配、全文检索、向量特征）。

### 5.3 前端交互建议
- **Resource/Output 节点**：提供 Checkbox 列表，用户操作产生 `["a", "b"]`。
- **Join/Union 节点**：提供 Mapping 表格，用户操作产生 `from` 对象（单一映射或对齐数组）。
- **SQL 节点**：点击“解析”按钮，后端回写 `runtime_output_fields`，前端同步更新预览。

---

## 6. 设计优势
1. **统一性**：一套 Struct 处理所有算子，代码复用率高。
2. **健壮性**：`from` 路径引用彻底解决了 SQL 生成时的 `Ambiguous column` 错误。
3. **灵活性**：支持 `["*"]` 极大地减少了大型表建模时的配置工作量。
4. **AI 友好**：结构化的 DSL 极大降低了大模型（Agent）生成错误配置的概率。

---

## 7. Other

<details>
<summary>完整 logic_defintition 示例</summary>

### 7.1 join
```json
{
    "id": "test_join",
    "catalog_id": "logicview",
    "name": "复合视图join",
    "tags": [],
    "description": "两表join",
    "category": "logicview",
    "logic_definition": [
        {
            "id": "node_output",
            "name": "输出节点",
            "type": "output",
            "inputs": [
                "node3"
            ],
            "config": {},
            "output_fields": [
                "*"
            ]
        },
        {
            "id": "node3",
            "name": "学生表和选课信息表关联",
            "type": "join",
            "inputs": [
                "node1",
                "node2"
            ],
            "config": {
                "join_on": [
                    {
                        "left_field": "student_id",
                        "operator": "=",
                        "right_field": "student_id"
                    }
                ],
                "join_type": "left"
            },
            "output_fields": [
                {
                    "name": "student_id",
                    "from": "student_id",
                    "from_node": "node1"
                },
                {
                    "name": "student_no",
                    "from": "student_no",
                    "from_node": "node1"
                },
                {
                    "name": "student_name",
                    "from": "student_name",
                    "from_node": "node1"
                },
                {
                    "name": "gender",
                    "from": "gender",
                    "from_node": "node1"
                },
                {
                    "name": "age",
                    "from": "age",
                    "from_node": "node1"
                },
                {
                    "name": "class_name",
                    "from": "class_name",
                    "from_node": "node1"
                },
                {
                    "name": "create_time",
                    "from": "create_time",
                    "from_node": "node1"
                },
                {
                    "name": "sc_id",
                    "from": "sc_id",
                    "from_node": "node2"
                },
                {
                    "name": "student_id_student_course",
                    "from": "student_id",
                    "from_node": "node2"
                },
                {
                    "name": "course_id",
                    "from": "course_id",
                    "from_node": "node2"
                },
                {
                    "name": "score",
                    "from": "score",
                    "from_node": "node2"
                },
                {
                    "name": "select_time",
                    "from": "select_time",
                    "from_node": "node2"
                }
            ]
        },
        {
            "id": "node2",
            "name": "学生选课信息表",
            "type": "resource",
            "inputs": [],
            "config": {
                "resource_id": "d7fg89jkqlq4ronf0epg",
                "distinct": false,
                "filters": {}
            },
            "output_fields": [
                "*"
            ]
        },
        {
            "id": "node1",
            "name": "学生信息表",
            "type": "resource",
            "inputs": [],
            "config": {
                "distinct": false,
                "filters": {},
                "resource_id": "d7fg89jkqlq4ronf0ep0"
            },
            "output_fields": [
                "*"
            ]
        }
    ]
}
```

### 7.2 union
```json
{
    "id": "test_union",
    "catalog_id": "logicview",
    "name": "复合视图union",
    "tags": [],
    "description": "三表union",
    "category": "logicview",
    "logic_definition": [
        {
            "id": "node_output",
            "name": "输出节点",
            "type": "output",
            "inputs": [
                "node4"
            ],
            "config": {},
            "output_fields": [
                "*"
            ]
        },
        {
            "id": "node4",
            "name": "大陆办公室员工表、香港办公室员工表和临时外包人员表",
            "type": "union",
            "inputs": [
                "node1",
                "node2",
                "node3"
            ],
            "config": {
                "union_type": "all"
            },
            "output_fields": [
                {
                    "name": "emp_id",
                    "from": [
                        {
                            "from": "emp_id",
                            "from_node": "node1"
                        },
                        {
                            "from": "emp_id",
                            "from_node": "node2"
                        },
                        {
                            "from": "id",
                            "from_node": "node3"
                        }
                    ]
                },
                {
                    "name": "name",
                    "from": [
                        {
                            "from": "name",
                            "from_node": "node1"
                        },
                        {
                            "from": "name",
                            "from_node": "node2"
                        },
                        {
                            "from": "full_name",
                            "from_node": "node3"
                        }
                    ]
                },
                {
                    "name": "role",
                    "from": [
                        {
                            "from": "role",
                            "from_node": "node1"
                        },
                        {
                            "from": "role",
                            "from_node": "node2"
                        },
                        {
                            "from": "job_title",
                            "from_node": "node3"
                        }
                    ]
                }
            ]
        },
        {
            "id": "node3",
            "name": "临时外包人员表",
            "type": "resource",
            "inputs": [],
            "config": {
                "resource_id": "d7g3sojkqlq7jfnadmdg",
                "distinct": false,
                "filters": {}
            }
        },
        {
            "id": "node2",
            "name": "香港办公室员工表",
            "type": "resource",
            "inputs": [],
            "config": {
                "resource_id": "d7g3sojkqlq7jfnadmd0",
                "distinct": false,
                "filters": {}
            },
            "output_fields": [
                "*"
            ]
        },
        {
            "id": "node1",
            "name": "大陆办公室员工表",
            "type": "resource",
            "inputs": [],
            "config": {
                "distinct": false,
                "filters": {},
                "resource_id": "d7g3sojkqlq7jfnadmcg"
            },
            "output_fields": [
                "*"
            ]
        }
    ]
}
```

### 7.3 sql
```json
{
    "logic_definition": [
        {
            "id": "node_UgVju",
            "name": "erp_real_time_inventory",
            "type": "resource",
            "inputs": [],
            "config": {
                "resource_id": "2017573348090867713",
                "filters": {},
                "distinct": false
            },
            "output_fields": [
                "aux_attr",
                "available_base_qty",
                "available_inventory_qty",
                "base_qty",
                "base_uom",
                "batch_master",
                "batch_no",
                "bin_location",
                "conv_ratio",
                "conv_ratio_available",
                "conv_ratio_reserved",
                "custodian",
                "custodian_type",
                "expiry_date",
                "inbound_date",
                "inventory_org",
                "inventory_qty",
                "inventory_uom",
                "manufacture_date",
                "material_code",
                "material_name",
                "owner",
                "owner_type",
                "purchase_qty",
                "purchase_qty_available",
                "purchase_qty_reserved",
                "purchase_uom",
                "reserved_base_qty",
                "reserved_inventory_qty",
                "seq_no",
                "spec_model",
                "stock_status",
                "stock_type",
                "total_col",
                "warehouse"
            ]
        },
        {
            "id": "node_CRfXL",
            "name": "erp_material",
            "type": "resource",
            "inputs": [],
            "config": {
                "resource_id": "2017573348468355073",
                "filters": {
                    "value": "库存商品-产成品",
                    "operation": "!=",
                    "value_from": "const",
                    "field": "group_name"
                },
                "distinct": false
            },
            "output_fields": [
                "baseunit_name",
                "baseunit_number",
                "group_name",
                "group_type",
                "huid_productline_name",
                "material_code",
                "material_name",
                "material_standard_price",
                "materialattr",
                "modelnum",
                "product_fixedleadtime",
                "product_status",
                "purchase_fixedleadtime",
                "purchase_huid_batchqty",
                "purchase_huid_minlotsize"
            ]
        },
        {
            "id": "node_DQTew",
            "name": "SQL",
            "type": "sql",
            "inputs": [
                "node_UgVju",
                "node_CRfXL"
            ],
            "config": {
                "sql": "SELECT *\nFROM {{.node_UgVju}} eri\nWHERE NOT EXISTS (\n    SELECT 1\n    FROM {{.node_CRfXL}} emf\n    WHERE emf.material_code = eri.material_code\n)\n"
            },
            "output_fields": [
                "*"
            ]
        },
        {
            "id": "node-output",
            "name": "输出视图",
            "type": "output",
            "inputs": [
                "node_DQTew"
            ],
            "config": {},
            "output_fields": [
                "*"
            ]
        }
    ]
}
```

</details>

