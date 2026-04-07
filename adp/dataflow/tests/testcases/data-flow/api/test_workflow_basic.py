# -*- coding:UTF-8 -*-

import pytest
import allure
import uuid
import json
import time

from lib.dataflow_como_operator import AutomationClient


@allure.feature("automation服务接口测试：工作流操作")
class TestWorkflowOperations:
    client = AutomationClient()

    @allure.title("创建工作流--json节点")
    def test_create_workflow_json_node(self, Headers):
        """测试创建工作流，使用json节点"""
        # 准备测试数据
        unique_title = f"json_三个_{str(uuid.uuid4())[:8]}"
        data = {
            "title": unique_title,
            "description": "",
            "status": "normal",
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/manual"
                },
                {
                    "id": "1",
                    "title": "",
                    "operator": "@internal/json/template",
                    "parameters": {
                        "json": "{\n  \"name\": \"张三\",\n  \"age\": 28,\n  \"hobbies\": [\"阅读\", \"游泳\", \"编程\"],\n  \"address\": {\n    \"city\": \"北京\",\n    \"postalCode\": \"100000\"\n  }\n}",
                        "template": "用户信息：\n姓名：{{.name}}\n年龄：{{.age}}岁\n城市：{{.address.city}}\n爱好：{{index .hobbies 0}}、{{index .hobbies 1}} 和 {{index .hobbies 2}}。"
                    }
                },
                {
                    "id": "1001",
                    "title": "",
                    "operator": "@internal/json/get",
                    "parameters": {
                        "fields": [
                            {
                                "key": "name",
                                "type": "string",
                                "value": ""
                            },
                            {
                                "key": "age",
                                "type": "number",
                                "value": ""
                            },
                            {
                                "key": "hobbies",
                                "type": "array",
                                "value": ""
                            },
                            {
                                "key": "address",
                                "type": "object",
                                "value": ""
                            }
                        ],
                        "json": "{\n  \"name\": \"张三\",\n  \"age\": 28,\n  \"hobbies\": [\"阅读\", \"游泳\", \"编程\"],\n  \"address\": {\n    \"city\": \"北京\",\n    \"postalCode\": \"100000\"\n  }\n}"
                    }
                },
                {
                    "id": "1002",
                    "title": "",
                    "operator": "@internal/json/set",
                    "parameters": {
                        "fields": [
                            {
                                "key": "aa",
                                "type": "string",
                                "value": "{{__1001.fields._0}}"
                            },
                            {
                                "key": "bb",
                                "type": "string",
                                "value": "{{__1001.fields._1}}"
                            }
                        ],
                        "json": "{}"
                    }
                }
            ],
            "accessors": []
        }

        result = self.client.CreateDag(data, Headers)
        
        # 验证响应状态码
        assert result[0] == 201, f"创建工作流应返回状态码201，实际为: {result[0]}"
        
        # 验证响应数据结构
        assert "id" in result[1], "响应应包含id字段"
        
        # 保存dag_id供后续测试使用
        dag_id = result[1]["id"]
        allure.attach(str(dag_id), name="创建的DAG ID")
        
        return dag_id

    @allure.title("运行工作流")
    def test_run_workflow(self, Headers):
        """测试运行工作流"""
        # 先创建一个工作流
        unique_title = f"json_运行测试_{str(uuid.uuid4())[:8]}"
        create_data = {
            "title": unique_title,
            "description": "用于运行测试的工作流",
            "status": "normal",
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/manual"
                },
                {
                    "id": "1",
                    "title": "",
                    "operator": "@internal/json/template",
                    "parameters": {
                        "json": "{\"test\": \"data\"}",
                        "template": "测试数据：{{.test}}"
                    }
                }
            ],
            "accessors": []
        }
        
        create_result = self.client.CreateDag(create_data, Headers)
        assert create_result[0] == 201, "创建工作流失败"
        dag_id = create_result[1]["id"]
        
        # 运行工作流
        run_data = {}
        result = self.client.RunInstance(dag_id, run_data, Headers)
        
        # 验证响应状态码
        assert result[0] in [200, 201, 202], f"运行工作流应返回状态码200/201/202，实际为: {result[0]}"
        
        # 等待一段时间让工作流执行
        time.sleep(2)
        
        return dag_id

    @allure.title("查看运行日志")
    def test_get_workflow_results(self, Headers):
        """测试查看工作流运行日志"""
        # 先创建并运行一个工作流
        unique_title = f"json_日志测试_{str(uuid.uuid4())[:8]}"
        create_data = {
            "title": unique_title,
            "description": "用于日志测试的工作流",
            "status": "normal",
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/manual"
                },
                {
                    "id": "1",
                    "operator": "@internal/time/now"
                }
            ],
            "create_by": "direct"
        }
        
        create_result = self.client.CreateDag(create_data, Headers)
        assert create_result[0] == 201, "创建工作流失败"
        dag_id = create_result[1]["id"]
        
        # 运行工作流
        run_result = self.client.RunInstance(dag_id, {}, Headers)
        assert run_result[0] in [200, 201, 202], "运行工作流失败"
        
        # 等待执行完成
        time.sleep(3)
        
        # 查看运行日志
        params = {}
        result = self.client.GetDagResultsV2(dag_id, params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"查看运行日志应返回状态码200，实际为: {result[0]}"
        
        # 验证响应数据结构
        assert "results" in result[1], "响应应包含results字段"
        assert isinstance(result[1]["results"], list), "results应为数组类型"
        
        # 如果有结果，保存第一个result_id
        if len(result[1]["results"]) > 0:
            run_id = result[1]["results"][0]["id"]
            allure.attach(str(run_id), name="运行记录ID")
            return dag_id, run_id
        
        return dag_id, None



    @allure.title("列表查询 - 数据流列表")
    def test_list_dags_v2(self, Headers):
        """测试查询数据流列表（v2接口）"""
        params = {
            "page": 0,
            "limit": 50,
            "sortby": "updated_at",
            "order": "desc",
            "type": "data-flow"
        }
        
        result = self.client.ListDagsV2(params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"列表查询应返回状态码200，实际为: {result[0]}"
        
        # 验证响应数据结构
        assert "dags" in result[1] or "results" in result[1], "响应应包含dags或results字段"
        
        # 验证分页参数
        if "page" in result[1]:
            assert result[1]["page"] == params["page"], "返回的page应与请求参数一致"
        if "limit" in result[1]:
            assert result[1]["limit"] == params["limit"], "返回的limit应与请求参数一致"

    @allure.title("我的流程_列表查询")
    def test_list_my_dags(self, Headers):
        """测试查询我的工作流列表"""
        params = {
            "page": 0,
            "limit": 50,
            "sortby": "updated_at",
            "order": "desc"
        }
        
        result = self.client.ListMyDags(params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"我的流程列表查询应返回状态码200，实际为: {result[0]}"
        
        # 验证响应数据结构
        assert "dags" in result[1] or "results" in result[1], "响应应包含dags或results字段"

    @allure.title("分配给我的_列表查询")
    def test_list_shared_dags(self, Headers):
        """测试查询分配给我的工作流列表"""
        params = {
            "page": 0,
            "limit": 50,
            "sortBy": "updated_at",
            "order": "desc"
        }
        
        result = self.client.ListSharedDags(params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"分配给我的列表查询应返回状态码200，实际为: {result[0]}"
        
        # 验证响应数据结构
        assert "dags" in result[1] or "results" in result[1], "响应应包含dags或results字段"

    @allure.title("获取数据流详情")
    def test_get_dag_detail(self, Headers):
        """测试获取数据流详情"""
        # 先创建一个工作流
        unique_title = f"json_详情测试_{str(uuid.uuid4())[:8]}"
        create_data = {
            "title": unique_title,
            "description": "",
            "status": "normal",
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/manual"
                },
                {
                    "id": "1",
                    "operator": "@internal/time/now"
                }
            ],
            "create_by": "direct"
        }
        
        create_result = self.client.CreateDag(create_data, Headers)
        assert create_result[0] == 201, "创建工作流失败"
        dag_id = create_result[1]["id"]
        
        # 获取数据流详情
        result = self.client.GetDagDetail(dag_id, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"获取数据流详情应返回状态码200，实际为: {result[0]}"
        
        # 验证响应数据结构
        assert "id" in result[1], "响应应包含id字段"
        assert result[1]["id"] == dag_id, "返回的id应与请求的dag_id一致"
        assert "title" in result[1], "响应应包含title字段"
        assert result[1]["title"] == unique_title, "返回的title应与创建时一致"

    @allure.title("删除工作流")
    def test_delete_data_flow(self, Headers):
        """测试删除数据流"""
        # 先创建一个工作流
        unique_title = f"json_删除测试_{str(uuid.uuid4())[:8]}"
        create_data = {
            "title": unique_title,
            "description": "用于删除测试的工作流",
            "status": "normal",
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/manual"
                },
                {
                    "id": "1",
                    "operator": "@internal/time/now"
                }
            ],
            "create_by": "direct"
        }
        
        create_result = self.client.CreateDag(create_data, Headers)
        assert create_result[0] == 201, "创建工作流失败"
        dag_id = create_result[1]["id"]
        
        # 删除数据流
        result = self.client.DeleteDataFlow(dag_id, Headers)
        
        # 验证响应状态码（删除操作通常返回204 No Content或200）
        assert result[0] in [200, 204], f"删除数据流应返回状态码200或204，实际为: {result[0]}"
        
        # 验证删除后无法再获取
        get_result = self.client.GetDagDetail(dag_id, Headers)
        assert get_result[0] in [404, 400], f"删除后获取应返回404或400，实际为: {get_result[0]}"

    @allure.title("完整工作流生命周期测试")
    def test_workflow_lifecycle(self, Headers):
        """测试完整的工作流生命周期：创建->运行->查看日志->查看详情->删除"""
        # 1. 创建工作流
        unique_title = f"json_生命周期测试_{str(uuid.uuid4())[:8]}"
        create_data = {
            "title": unique_title,
            "description": "完整生命周期测试",
            "status": "normal",
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/manual"
                },
                {
                    "id": "1",
                    "title": "",
                    "operator": "@internal/json/template",
                    "parameters": {
                        "json": "{\"message\": \"test\"}",
                        "template": "消息：{{.message}}"
                    }
                }
            ],
            "accessors": []
        }
        
        create_result = self.client.CreateDag(create_data, Headers)
        assert create_result[0] == 201, "创建工作流失败"
        dag_id = create_result[1]["id"]
        allure.attach(str(dag_id), name="创建的DAG ID")
        
        # 2. 获取工作流详情
        detail_result = self.client.GetDagDetail(dag_id, Headers)
        assert detail_result[0] == 200, "获取工作流详情失败"
        assert detail_result[1]["id"] == dag_id, "获取的ID应与创建的一致"
        
        # 3. 运行工作流
        run_result = self.client.RunInstance(dag_id, {}, Headers)
        assert run_result[0] in [200, 201, 202], "运行工作流失败"
        
        # 4. 等待执行完成
        time.sleep(3)
        
        # 5. 查看运行日志
        results_result = self.client.GetDagResultsV2(dag_id, {}, Headers)
        assert results_result[0] == 200, "查看运行日志失败"
        assert "results" in results_result[1], "响应应包含results字段"
        
        # 6. 如果有运行记录，查看详情
        if len(results_result[1]["results"]) > 0:
            run_id = results_result[1]["results"][0]["id"]
            detail_result = self.client.GetDagResultDetail(dag_id, run_id, Headers)
            assert detail_result[0] == 200, "查看日志详情失败"
        
        # 7. 删除工作流
        delete_result = self.client.DeleteDataFlow(dag_id, Headers)
        assert delete_result[0] in [200, 204], "删除工作流失败"
        
        # 8. 验证删除成功
        get_after_delete = self.client.GetDagDetail(dag_id, Headers)
        assert get_after_delete[0] in [404, 400], "删除后应无法获取工作流"

    @allure.title("创建工作流--变量节点")
    def test_create_workflow_variable_node(self, Headers):
        """测试创建工作流，使用变量节点"""
        unique_title = f"变量_{str(uuid.uuid4())[:8]}"
        data = {
            "title": unique_title,
            "description": "",
            "status": "normal",
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/manual"
                },
                {
                    "id": "1",
                    "operator": "@internal/define",
                    "parameters": {
                        "type": "string",
                        "value": "a"
                    },
                    "title": ""
                },
                {
                    "id": "1008",
                    "operator": "@internal/define",
                    "parameters": {
                        "type": "number",
                        "value": 6
                    }
                },
                {
                    "id": "1001",
                    "title": "",
                    "operator": "@control/flow/branches",
                    "branches": [
                        {
                            "id": "1002",
                            "conditions": [],
                            "steps": [
                                {
                                    "id": "1003",
                                    "title": "",
                                    "operator": "@internal/assign",
                                    "parameters": {
                                        "target": "{{__1.value}}",
                                        "value": "b"
                                    }
                                }
                            ]
                        },
                        {
                            "id": "1004",
                            "conditions": [],
                            "steps": [
                                {
                                    "id": "1005",
                                    "operator": "@internal/assign",
                                    "parameters": {
                                        "target": "{{__1008.value}}",
                                        "value": 88
                                    },
                                    "title": ""
                                }
                            ]
                        }
                    ]
                }
            ],
            "accessors": []
        }

        result = self.client.CreateDag(data, Headers)
        
        assert result[0] == 201, f"创建工作流应返回状态码201，实际为: {result[0]}"
        assert "id" in result[1], "响应应包含id字段"
        
        dag_id = result[1]["id"]
        allure.attach(str(dag_id), name="创建的DAG ID")
        return dag_id

    @allure.title("创建工作流--文本节点")
    def test_create_workflow_text_node(self, Headers):
        """测试创建工作流，使用文本节点和表单触发"""
        unique_title = f"文本_三个_{str(uuid.uuid4())[:8]}"
        data = {
            "title": unique_title,
            "description": "",
            "status": "normal",
            "accessors": [
                {
                    "id": "{{departmentId}}",
                    "name": "组织结构",
                    "type": "department"
                }
            ],
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/form",
                    "parameters": {
                        "fields": [
                            {
                                "key": "ZHmeLdTgSBbsQHuS",
                                "type": "string",
                                "name": "aa",
                                "description": {},
                                "default": "default"
                            }
                        ]
                    },
                    "title": ""
                },
                {
                    "id": "1",
                    "title": "",
                    "operator": "@internal/text/match",
                    "parameters": {
                        "matchtype": "NUMBER",
                        "text": "订单12345,金额678.90,数量100,状态1,优先级5"
                    }
                },
                {
                    "id": "1001",
                    "title": "",
                    "operator": "@internal/text/split",
                    "parameters": {
                        "separator": ";",
                        "text": "SKU001;SKU002;SKU003;SKU004;SKU005"
                    }
                },
                {
                    "id": "1002",
                    "title": "",
                    "operator": "@internal/text/join",
                    "parameters": {
                        "custom": "-",
                        "separator": "custom",
                        "texts": [
                            "需求分析",
                            "测试验收"
                        ]
                    }
                }
            ],
            "create_by": "local"
        }

        result = self.client.CreateDag(data, Headers)
        
        assert result[0] == 201, f"创建工作流应返回状态码201，实际为: {result[0]}"
        assert "id" in result[1], "响应应包含id字段"
        
        dag_id = result[1]["id"]
        allure.attach(str(dag_id), name="创建的DAG ID")
        return dag_id

    @allure.title("运行工作流--表单触发")
    def test_run_workflow_form(self, Headers):
        """测试运行表单触发的工作流"""
        # 先创建一个表单触发的工作流
        dag_id = self.test_create_workflow_text_node(Headers)
        
        # 运行工作流（表单触发）
        run_data = {
            "data": {
                "ZHmeLdTgSBbsQHuS": "default"
            }
        }
        result = self.client.RunInstanceForm(dag_id, run_data, Headers)
        
        assert result[0] in [200, 201, 202], f"运行工作流应返回状态码200/201/202，实际为: {result[0]}"
        
        time.sleep(2)
        return dag_id

    @allure.title("创建数据流--python节点")
    def test_create_dataflow_python_node(self, Headers):
        """测试创建数据流，使用python节点"""
        unique_title = f"python节点_{str(uuid.uuid4())[:8]}"
        data = {
            "title": unique_title,
            "description": "",
            "status": "normal",
            "steps": [
                {
                    "id": "0",
                    "title": "",
                    "operator": "@trigger/manual"
                },
                {
                    "id": "1001",
                    "title": "",
                    "operator": "@internal/tool/py3",
                    "parameters": {
                        "code": """def main(input_string, input_number, input_array, input_object):
    # 1. 处理string类型
    str_result = ""
    try:
        str_result = str(input_string)
        processed_string = "处理后的字符串: " + str_result.upper()
    except:
        processed_string = "无法处理字符串"
    
    # 2. 处理number类型
    number_result = 0
    try:
        number_result = float(input_number)
        processed_number = number_result * 2 + 10
    except:
        processed_number = "输入不是有效数字"
    
    # 3. 处理array类型
    arr = []
    try:
        if input_array[0]:
            arr = []
            for item in input_array:
                arr.append(item)
    except:
        arr = [str(input_array)]
    
    # 计算数组统计信息
    array_sum = 0
    array_avg = 0
    array_length = len(arr)
    
    count_num = 0
    total = 0
    for item in arr:
        try:
            val = float(item)
            total = total + val
            count_num = count_num + 1
        except:
            continue
    
    if count_num > 0:
        array_sum = total
        array_avg = total / count_num
    
    # 创建处理后的数组
    if count_num > 0:
        new_arr = []
        for item in arr:
            new_arr.append(item)
        new_arr.append(6)
        new_arr.append(7)
        new_arr.append(8)
        processed_array = new_arr
    else:
        new_arr = []
        for item in arr:
            new_arr.append(item)
        new_arr.append("新增1")
        new_arr.append("新增2")
        processed_array = new_arr
    
    # 4. 处理object类型
    obj_original = {}
    obj_processed = {}
    
    try:
        obj_keys = []
        for key in input_object:
            obj_keys.append(key)
        
        if len(obj_keys) > 0:
            obj_original = input_object
            obj_processed = {}
            for key in obj_keys:
                obj_processed[key] = input_object[key]
            obj_processed["processed"] = True
        else:
            obj_original = {"value": input_object}
            obj_processed = {"value": input_object, "processed": True}
    except:
        obj_original = {"value": input_object}
        obj_processed = {"value": input_object, "processed": True}
    
    # 返回结果
    result = {}
    result["string_original"] = str_result
    result["string_processed"] = processed_string
    result["number_original"] = input_number
    result["number_processed"] = processed_number
    result["array_original"] = arr
    result["array_processed"] = processed_array
    result["array_stats"] = {
        "sum": array_sum,
        "average": array_avg,
        "length": array_length,
        "numeric_count": count_num
    }
    result["object_original"] = obj_original
    result["object_processed"] = obj_processed
    result["summary"] = {
        "string_length": len(str_result),
        "array_length": array_length,
        "result": "处理完成"
    }
    
    return result""",
                        "input_params": [
                            {
                                "id": "zc852",
                                "key": "input_string",
                                "type": "string",
                                "value": "Hello"
                            },
                            {
                                "id": "2bjaw",
                                "key": "input_number",
                                "type": "int",
                                "value": "42"
                            },
                            {
                                "id": "didvh",
                                "key": "input_array",
                                "type": "array",
                                "value": " [1, 2, 3, 4, 5]"
                            },
                            {
                                "id": "kql12",
                                "key": "input_object",
                                "type": "object",
                                "value": "{\"name\": \"张三\", \"age\": 25, \"city\": \"北京\"}"
                            }
                        ],
                        "output_params": [
                            {
                                "id": "eyz6c",
                                "key": "result",
                                "type": "object"
                            }
                        ]
                    }
                },
                {
                    "id": "1002",
                    "title": "",
                    "operator": "@internal/json/get",
                    "parameters": {
                        "fields": [
                            {
                                "key": "shuzu",
                                "type": "array",
                                "value": "{{__1001.result.array_original}}"
                            },
                            {
                                "key": "duixiang",
                                "type": "object",
                                "value": "{{__1001.result.array_stats}}"
                            },
                            {
                                "key": "str",
                                "type": "string",
                                "value": "{{__1001.result.string_original}}"
                            },
                            {
                                "key": "num",
                                "type": "number",
                                "value": "{{__1001.result.summary.array_length}}"
                            }
                        ],
                        "json": "{{__1001.result}}"
                    }
                }
            ],
            "accessors": []
        }

        result = self.client.CreateDag(data, Headers)
        
        assert result[0] == 201, f"创建数据流应返回状态码201，实际为: {result[0]}"
        assert "id" in result[1], "响应应包含id字段"
        
        dag_id = result[1]["id"]
        allure.attach(str(dag_id), name="创建的DAG ID")
        return dag_id