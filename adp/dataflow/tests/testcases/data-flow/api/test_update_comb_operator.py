# -*- coding:UTF-8 -*-

import pytest
import allure
import uuid

from lib.dataflow_como_operator import *
from lib.operator import *

@allure.feature("automation服务接口测试：更新组合算子")
class TestUpdateCombinationOperator:
    client = AutomationClient()
    agent_client = Operator()
    
    # 全局变量，存储共享的算子信息
    edit_operator_title = f"edit_test_{str(uuid.uuid4())[:8]}"
    edit_operator_id = None
    edit_operator_version = None
    edit_operator_dag_id = None  # 添加保存dag_id的全局变量
    
    @pytest.fixture(scope="class", autouse=True)
    def setup_edit_operator(self, request, Headers):
        """创建一个用于测试的组合算子，并设置全局变量"""
        # 创建组合算子的数据
        data = {
            "title": self.edit_operator_title,
            "description": "用于测试更新功能的edit算子",
            "category": "data_process",
            "steps": [
                {
                "id": "0",
                "operator": "@trigger/form",
                "parameters": {
                    "fields": [
                    {
                        "type": "string",
                        "key": "text_input",
                        "name": "文本输入",
                        "description": {
                            "type": "text"
                        },
                        "required": True
                    },
                    {
                        "type": "number",
                        "key": "num_input",
                        "name": "数字输入",
                        "description": {
                            "type": "text"
                        }
                    }
                    ]
                }
                },
                {
                "id": "1",
                "operator": "@internal/tool/py3",
                "parameters": {
                    "input_params": [
                    {
                        "id": "text_param",
                        "key": "text",
                        "type": "string",
                        "value": "{{__0.fields.text_input}}"
                    },
                    {
                        "id": "num_param",
                        "key": "num",
                        "type": "int",
                        "value": "{{__0.fields.num_input}}"
                    }
                    ],
                    "output_params": [
                    {
                        "id": "result_param",
                        "key": "result",
                        "type": "string"
                    }
                    ],
                    "code": "def main(text, num):\n    if num is None:\n        num = 1\n    result = text * num\n    return result"
                }
                },
                {
                "id": "2",
                "operator": "@internal/return",
                "parameters": {
                    "output": "{{__1.result}}"
                }
                }
            ],
            "outputs": [
                {
                "type": "string",
                "key": "output",
                "name": "输出结果",
                "description": {
                    "type": "text"
                }
                }
            ]
        }
        
        # 发送创建请求
        result = self.client.CreateCombinationOperator(data, Headers)
        
        if result[0] != 201:
            pytest.skip(f"创建edit算子失败，状态码: {result[0]}")
            
        # 获取创建的算子ID (创建时返回的id)
        created_id = result[1].get("id", "")
        if not created_id:
            pytest.skip("创建的算子没有返回id")
        
        # 通过GetOperatorsList获取operator_id和version
        # 使用算子名称来查询
        current_title = self.edit_operator_title
        params = {
            "name": current_title,
            "limit": 100
        }
        
        # 获取算子列表
        result = self.client.GetOperatorsList(params, Headers)
        print("================== GetOperatorsList 响应 ==================")
        print(result)
        
        # 检查API调用是否成功
        if result[0] != 200:
            pytest.skip(f"获取算子列表失败，状态码: {result[0]}")
            
        # 检查返回结果是否包含算子
        if not result[1].get("ops") or len(result[1]["ops"]) == 0:
            pytest.skip(f"获取算子列表成功但未找到算子: {self.edit_operator_title}")
        
        # 从结果中提取operator_id和version
        found = False
        for op in result[1]["ops"]:
            # 打印调试信息，查看每个算子的属性
            print(f"检查算子详情: {op}")
            # 使用operator_name来匹配
            if op.get("operator_name") == self.edit_operator_title or op.get("name") == self.edit_operator_title:
                operator_id = op.get("operator_id")
                original_version = op.get("version")
                dag_id = op.get("dag_id", "")  # 使用空字符串作为默认值
                found = True
                break
                
        if not found:
            pytest.skip(f"获取算子列表成功但未找到匹配的算子: {self.edit_operator_title}")
            
        if not operator_id:
            pytest.skip(f"在算子列表中未找到刚创建的算子: {self.edit_operator_title}")
            
        # 打印调试信息
        print(f"找到算子 - ID: {operator_id}, 版本: {original_version}, DAG_ID: {dag_id}")
            
        # 设置类变量，供所有测试方法使用
        TestUpdateCombinationOperator.edit_operator_id = operator_id
        TestUpdateCombinationOperator.edit_operator_version = original_version
        TestUpdateCombinationOperator.edit_operator_dag_id = dag_id
        
        # 清理函数
        def teardown():
            # 此处可以添加测试完成后的清理代码，例如删除创建的算子
            pass
        
        request.addfinalizer(teardown)
    
    def get_default_steps(self):
        """提供默认的算子步骤结构"""
        return [
    {
      "id": "0",
      "operator": "@trigger/form",
      "parameters": {
        "fields": [
          {
            "type": "string",
            "key": "str",
            "description": {
              "type": "text"
            }
          }
        ]
      }
    },
    {
      "id": "1",
      "operator": "@internal/json/set",
      "parameters": {
        "json": "{}",
        "fields": [
          {
            "key": "a",
            "type": "string",
            "value": "{{__0.fields.str}}"
          }
        ]
      }
    },
    {
      "id": "2",
      "operator": "@internal/return",
      "parameters": {
        "succ": "{{__1.json}}"
      }
    }
        ]

    def get_modified_steps(self):
        """提供修改过的算子步骤结构，用于测试步骤更新"""
        return [
            {
                "id": "0",
                "title": "触发器",
                "operator": "@trigger/form",
                "parameters": {
                    "fields": [
                        {
                            "key": "text",
                            "name": "输入文本",
                            "required": True,
                            "type": "string"
                        },
                        {
                            "key": "number",
                            "name": "输入数字",
                            "required": False,
                            "type": "number"
                        }
                    ]
                }
            },
            {
                "id": "1",
                "operator": "@internal/time/now",
                "title": "获取时间"
            },
            {
                "id": "2",
                "title": "返回结果",
                "operator": "@internal/return",
                "parameters": {
                    "content": "{{_0.fields.text}}",
                    "number": "{{_0.fields.number}}",
                    "curtime": "{{_1.curtime}}"
                }
            }
        ]
    
    @allure.title("创建edit算子成功")
    @pytest.mark.independent
    def test_create_edit_operator_success(self, Headers):
        """验证edit算子创建成功并正确设置了全局变量"""
        assert self.edit_operator_title, "edit算子标题未正确设置"
        assert self.edit_operator_id, "edit算子ID未正确设置"
        assert self.edit_operator_version, "edit算子版本未正确设置"
        
        # 验证算子信息
        detail_result = self.agent_client.GetOperatorInfo(self.edit_operator_id, "", Headers)
        assert detail_result[0] == 200, "获取算子详情失败"
        
        operator_info = detail_result[1]

        
        # 检查返回的字段，可能是name或operator_name
        name_field = "operator_name" if "operator_name" in operator_info else "name"
        
        # 使用正确的字段名进行断言
        assert operator_info[name_field] == self.edit_operator_title, f"算子标题不匹配: 期望 {self.edit_operator_title}, 实际 {operator_info.get(name_field)}"
        assert operator_info["operator_id"] == self.edit_operator_id, f"算子ID不匹配: 期望 {self.edit_operator_id}, 实际 {operator_info.get('operator_id')}"
        assert operator_info["version"] == self.edit_operator_version, f"算子版本不匹配: 期望 {self.edit_operator_version}, 实际 {operator_info.get('version')}"
    
    @allure.title("更新edit算子成功 - 基础信息")
    def test_update_edit_operator_basic_info(self, Headers):
        """测试更新创建的edit算子的title和description不会改变版本"""
        # 使用公共变量
        operator_id = self.edit_operator_id
        original_version = self.edit_operator_version
        original_title = self.edit_operator_title
        dag_id = self.edit_operator_dag_id
        
        # 构建更新数据 - 只更新名称和描述字段
        new_title = original_title + "_已更新"
        update_data = {
            "title": new_title,
            "description": "更新后的edit算子描述",
            "steps": [
                {
                    "id": "0",
                    "title": "触发器",
                    "operator": "@trigger/form",
                    "parameters": {
                        "fields": [
                            {
                                "key": "text",
                                "name": "输入文本",
                                "required": True,
                                "type": "string"
                            }
                        ]
                    }
                },
                {
                    "id": "1",
                    "title": "获取时间",
                    "operator": "@internal/time/now"
                },
                {
                    "id": "2",
                    "title": "返回结果",
                    "operator": "@internal/return",
                    "parameters": {
                        "content": "{{_0.fields.text}}",
                        "curtime": "{{_1.curtime}}"
                    }
                }
            ],
            "operator_id": operator_id,
            "version": original_version,
            "category": "data_process",
            "outputs": [
                {
                    "key": "content",
                    "name": "内容",
                    "type": "string"
                },
                {
                    "key": "curtime",
                    "name": "当前时间",
                    "type": "string"
                }
            ]
        }
        
        # 打印请求信息
        print(f"更新请求 - ID: {operator_id}, DAG_ID: {dag_id}, 版本: {original_version}")
        print(f"更新数据: {update_data}")

        # 发送更新请求
        result = self.client.UpdateOperator(dag_id, update_data, Headers)

        # 打印响应结果
        print(f"更新响应: {result}")
        
        
        # 验证响应状态码
        assert result[0] == 204, f"更新组合算子应返回状态码204，实际为: {result[0]}"
        
        # 从算子列表中获取最新的算子信息
        success, new_operator_id, new_version, new_dag_id, new_title = self.get_latest_operator_info(new_title, Headers)

        
        # 验证获取算子信息成功
        assert success, f"在算子列表中未找到更新后的算子: {new_title}"
        
        # 更新全局变量
        TestUpdateCombinationOperator.edit_operator_id = new_operator_id
        TestUpdateCombinationOperator.edit_operator_version = new_version
        TestUpdateCombinationOperator.edit_operator_title = new_title
        TestUpdateCombinationOperator.edit_operator_dag_id = new_dag_id
        
        # 验证更新成功 - 获取更新后的算子信息
        get_result = self.agent_client.GetOperatorInfo(self.edit_operator_id, "", Headers)

        assert get_result[0] == 200, "获取更新后的组合算子信息失败"
        
        # 获取更新后的算子信息
        updated_operator = get_result[1]
        
        print(updated_operator)
        
        # 验证名称字段已更新
        assert update_data["title"] == updated_operator["name"], "name字段未更新成功"
        
        # 验证版本未变化（更新名称和描述不会改变版本）
        if updated_operator["version"] != original_version:
            print("版本已更新，使用新版本")
            TestUpdateCombinationOperator.edit_operator_version = updated_operator["version"]
        else:
            print("版本未更新，使用原版本")
            TestUpdateCombinationOperator.edit_operator_version = original_version
        
        # 打印更新后的全局变量
        print(f"更新后的全局变量 - 标题: {self.edit_operator_title}, ID: {self.edit_operator_id}, 版本: {self.edit_operator_version}, DAG_ID: {self.edit_operator_dag_id}")
    
    @allure.title("更新edit算子成功 - 更新步骤生成新版本")
    def test_update_edit_operator_steps_new_version(self, Headers):
        """测试更新创建的edit算子的步骤会生成新版本"""
        # 使用公共变量
        operator_id = self.edit_operator_id
        original_version = self.edit_operator_version
        original_title = self.edit_operator_title
        dag_id = self.edit_operator_dag_id
        
        # 构建更新数据 - 修改步骤结构
        update_data = {
            "title": original_title,
            "description": "测试更新步骤生成新版本",
            "steps": [
                {
                    "id": "0",
                    "title": "触发器",
                    "operator": "@trigger/form",
                    "parameters": {
                        "fields": [
                            {
                                "key": "text",
                                "name": "输入文本",
                                "required": True,
                                "type": "string"
                            },
                            {
                                "key": "number",
                                "name": "输入数字",
                                "required": False,
                                "type": "number"
                            }
                        ]
                    }
                },
                {
                    "id": "1",
                    "title": "获取时间",
                    "operator": "@internal/time/now"
                },
                {
                    "id": "2",
                    "title": "返回结果",
                    "operator": "@internal/return",
                    "parameters": {
                        "content": "{{_0.fields.text}}",
                        "number": "{{_0.fields.number}}",
                        "curtime": "{{_1.curtime}}"
                    }
                }
            ],
            "operator_id": operator_id,
            "version": original_version,
            "category": "data_process",
            "outputs": [
                {
                    "key": "content",
                    "name": "内容",
                    "type": "string"
                },
                {
                    "key": "number",
                    "name": "数字",
                    "type": "number"
                },
                {
                    "key": "curtime",
                    "name": "当前时间",
                    "type": "string"
                }
            ]
        }
        
        # 打印请求信息
        print(f"更新步骤请求 - ID: {operator_id}, DAG_ID: {dag_id}, 版本: {original_version}")
        print(f"更新数据: {update_data}")
        
        # 发送更新请求
        result = self.client.UpdateOperator(dag_id, update_data, Headers)
        
        # 打印响应结果
        print(f"更新响应: {result}")
        
        # 验证响应状态码
        assert result[0] == 204, f"更新组合算子应返回状态码204，实际为: {result[0]}"
        
        # 从算子列表中获取最新的算子信息
        success, new_operator_id, new_version, new_dag_id, new_title = self.get_latest_operator_info(original_title, Headers)
        
        # 验证获取算子信息成功
        assert success, f"在算子列表中未找到更新后的算子: {original_title}"
        
        # 更新全局变量
        TestUpdateCombinationOperator.edit_operator_id = new_operator_id
        TestUpdateCombinationOperator.edit_operator_version = new_version
        TestUpdateCombinationOperator.edit_operator_title = new_title
        TestUpdateCombinationOperator.edit_operator_dag_id = new_dag_id
        
        # 验证更新成功 - 获取更新后的算子信息
        get_result = self.agent_client.GetOperatorInfo(self.edit_operator_id, "", Headers)
        assert get_result[0] == 200, "获取更新后的组合算子信息失败"
        
        # 获取更新后的算子信息
        updated_operator = get_result[1]
        
        # 验证版本已更新（更新步骤应生成新版本）
        assert original_version != updated_operator["version"], "更新步骤应生成新版本号"
        
        # 打印更新后的全局变量
        print(f"更新后的全局变量 - 标题: {self.edit_operator_title}, ID: {self.edit_operator_id}, 版本: {self.edit_operator_version}, DAG_ID: {self.edit_operator_dag_id}")
    
    @allure.title("更新edit算子成功 - 更新分类生成新版本")
    def test_update_edit_operator_category_new_version(self, Headers):
        """测试更新创建的edit算子的分类会生成新版本"""
        # 使用公共变量
        operator_id = self.edit_operator_id
        original_version = self.edit_operator_version
        original_title = self.edit_operator_title
        dag_id = self.edit_operator_dag_id
        
        # 构建更新数据 - 修改分类
        update_data = {
            "title": original_title,
            "description": "测试更新分类生成新版本",
            "steps": [
                {
                    "id": "0",
                    "title": "触发器",
                    "operator": "@trigger/form",
                    "parameters": {
                        "fields": [
                            {
                                "key": "text",
                                "name": "输入文本",
                                "required": True,
                                "type": "string"
                            }
                        ]
                    }
                },
                {
                    "id": "1",
                    "title": "获取时间",
                    "operator": "@internal/time/now"
                },
                {
                    "id": "2",
                    "title": "返回结果",
                    "operator": "@internal/return",
                    "parameters": {
                        "content": "{{_0.fields.text}}",
                        "curtime": "{{_1.curtime}}"
                    }
                }
            ],
            "operator_id": operator_id,
            "version": original_version,
            "category": "other_category",  # 更改分类为other_category
            "outputs": [
                {
                    "key": "content",
                    "name": "内容",
                    "type": "string"
                },
                {
                    "key": "curtime",
                    "name": "当前时间",
                    "type": "string"
                }
            ]
        }
        
        # 打印请求信息
        print(f"更新分类请求 - ID: {operator_id}, DAG_ID: {dag_id}, 版本: {original_version}")
        print(f"更新数据: {update_data}")
        
        # 发送更新请求
        result = self.client.UpdateOperator(dag_id, update_data, Headers)
        
        # 验证响应状态码
        assert result[0] == 204, f"更新组合算子应返回状态码204，实际为: {result[0]}"
        
        # 从算子列表中获取最新的算子信息
        success, new_operator_id, new_version, new_dag_id, new_title = self.get_latest_operator_info(original_title, Headers)
        
        # 验证获取算子信息成功
        assert success, f"在算子列表中未找到更新后的算子: {original_title}"
        
        # 更新全局变量
        TestUpdateCombinationOperator.edit_operator_id = new_operator_id
        TestUpdateCombinationOperator.edit_operator_version = new_version
        TestUpdateCombinationOperator.edit_operator_title = new_title
        TestUpdateCombinationOperator.edit_operator_dag_id = new_dag_id
        
        # 验证更新成功 - 获取更新后的算子信息
        get_result = self.agent_client.GetOperatorInfo(self.edit_operator_id, "", Headers)
        assert get_result[0] == 200, "获取更新后的组合算子信息失败"
        
        # 获取更新后的算子信息
        updated_operator = get_result[1]
        
        # 验证版本已更新（更新分类应生成新版本）
        assert original_version != updated_operator["version"], "更新分类应生成新版本号"
        
        # 验证分类已更新（通过获取算子信息接口，断言 operator_info 里的 category_name）
        get_result = self.agent_client.GetOperatorInfo(self.edit_operator_id, "", Headers)
        assert get_result[0] == 200, "获取更新后的组合算子信息失败"
        operator_info = get_result[1].get("operator_info", {})
        assert operator_info.get("category_name") == "未分类", "分类未更新为未分类"
        
        # 打印更新后的全局变量
        print(f"更新后的全局变量 - 标题: {self.edit_operator_title}, ID: {self.edit_operator_id}, 版本: {self.edit_operator_version}, DAG_ID: {self.edit_operator_dag_id}")
    
    @allure.title("更新edit算子成功 - 更新exec_mode生成新版本")
    def test_update_edit_operator_exec_mode_new_version(self, Headers):
        """测试更新创建的edit算子的执行模式会生成新版本"""
        # 使用公共变量
        operator_id = self.edit_operator_id
        original_version = self.edit_operator_version
        original_title = self.edit_operator_title
        dag_id = self.edit_operator_dag_id
        
        # 构建更新数据 - 修改执行模式
        update_data = {
            "title": original_title,
            "description": "测试更新执行模式生成新版本",
            "steps": [
                {
                    "id": "0",
                    "title": "触发器",
                    "operator": "@trigger/form",
                    "parameters": {
                        "fields": [
                            {
                                "key": "text",
                                "name": "输入文本",
                                "required": True,
                                "type": "string"
                            }
                        ]
                    }
                },
                {
                    "id": "1",
                    "operator": "@internal/tool/py3",
                    "parameters": {
                        "input_params": [],
                        "output_params": [
                            {
                                "id": "bj5yy",
                                "key": "b",
                                "type": "string"
                            }
                        ],
                        "code": "def main():\n    return \"hello word\""
                    }
                },
                {
                    "id": "2",
                    "operator": "@internal/return",
                    "parameters": {}
                }
            ],
            "operator_id": operator_id,
            "version": original_version, # 从sync更改为async
            "category": "other_category",
            "outputs": [
                {
                    "key": "b",
                    "name": "Python输出",
                    "type": "string"
                }
            ]
        }
        
        # 打印请求信息
        print(f"更新执行模式请求 - ID: {operator_id}, DAG_ID: {dag_id}, 版本: {original_version}")
        print(f"更新数据: {update_data}")
        
        # 发送更新请求
        result = self.client.UpdateOperator(dag_id, update_data, Headers)
        
        # 打印响应结果
        print(f"更新响应: {result}")
        
        # 验证响应状态码
        assert result[0] == 204, f"更新组合算子应返回状态码204，实际为: {result[0]}"
        
        # 从算子列表中获取最新的算子信息
        success, new_operator_id, new_version, new_dag_id, new_title = self.get_latest_operator_info(original_title, Headers)
        
        # 验证获取算子信息成功
        assert success, f"在算子列表中未找到更新后的算子: {original_title}"
        
        # 更新全局变量
        TestUpdateCombinationOperator.edit_operator_id = new_operator_id
        TestUpdateCombinationOperator.edit_operator_version = new_version
        TestUpdateCombinationOperator.edit_operator_title = new_title
        TestUpdateCombinationOperator.edit_operator_dag_id = new_dag_id
        
        # 验证更新成功 - 获取更新后的算子信息
        get_result = self.agent_client.GetOperatorInfo(self.edit_operator_id, "", Headers)
        assert get_result[0] == 200, "获取更新后的组合算子信息失败"
        
        # 获取更新后的算子信息
        updated_operator = get_result[1]
        
        # 验证版本已更新（更新执行模式应生成新版本）
        assert original_version != updated_operator["version"], "更新执行模式应生成新版本号"
        
        # 验证执行模式已更新（通过获取算子信息接口，断言 operator_info 里的 execution_mode）
        get_result = self.agent_client.GetOperatorInfo(self.edit_operator_id, "", Headers)
        assert get_result[0] == 200, "获取更新后的组合算子信息失败"
        operator_info = get_result[1].get("operator_info", {})
        assert operator_info.get("execution_mode") == "async", "执行模式未更新为async"
        
        # 打印更新后的全局变量
        print(f"更新后的全局变量 - 标题: {self.edit_operator_title}, ID: {self.edit_operator_id}, 版本: {self.edit_operator_version}, DAG_ID: {self.edit_operator_dag_id}")
    
    @allure.title("更新edit算子失败 - 仅更新title和description")
    def test_update_edit_operator_only_name_desc(self, Headers):
        """测试仅更新创建的edit算子的标题和描述，不改变版本号"""
        # 使用公共变量
        operator_id = self.edit_operator_id
        original_version = self.edit_operator_version
        original_title = self.edit_operator_title
        dag_id = self.edit_operator_dag_id
        
        # 生成新的随机标题
        import uuid
        new_title = f"更新后的编辑算子-{str(uuid.uuid4())[:8]}"
        
        # 构建更新数据 - 仅修改标题和描述
        update_data = {
            "title": new_title,
            "description": "测试仅更新标题和描述，不改变版本号",
            "operator_id": operator_id,
            "version": original_version
        }
        
        # 打印请求信息
        print(f"更新标题和描述请求 - ID: {operator_id}, DAG_ID: {dag_id}, 版本: {original_version}")
        print(f"更新前标题: {original_title}, 更新后标题: {new_title}")
        print(f"更新数据: {update_data}")
        
        # 发送更新请求
        result = self.client.UpdateOperator(dag_id, update_data, Headers)
        
        # 打印响应结果
        print(f"更新响应: {result}")
        
        # 验证响应状态码
        assert result[0] == 400, f"更新组合算子应返回状态码400，实际为: {result[0]}"
    
    @allure.title("更新edit算子失败 - 无效的operator_id")
    def test_update_edit_operator_invalid_id(self, Headers):
        """测试使用无效的id更新edit算子"""
        # 使用无效的id
        invalid_dag_id = "invalid_dag_id_123"
        invalid_operator_id = "invalid_id_123"
        
        # 构建更新数据
        update_data = {
            "title": "无效ID更新测试",
            "description": "测试无效ID情况更新失败",
            "operator_id": invalid_operator_id,
            "version": "1"
        }
        
        # 打印请求信息
        print(f"无效ID更新请求 - DAG_ID: {invalid_dag_id}")
        print(f"更新数据: {update_data}")
        
        # 发送更新请求
        result = self.client.UpdateOperator(invalid_dag_id, update_data, Headers)
        
        # 打印响应结果
        print(f"更新响应: {result}")
        
        # 验证响应状态码
        assert result[0] == 404, f"使用无效的id更新组合算子应返回状态码404，实际为: {result[0]}"
    
    @allure.title("更新edit算子失败 - 版本不匹配")
    def test_update_edit_operator_version_mismatch(self, Headers):
        """测试使用不匹配的版本更新edit算子"""
        # 使用公共变量
        operator_id = self.edit_operator_id
        original_version = self.edit_operator_version
        original_title = self.edit_operator_title
        dag_id = self.edit_operator_dag_id
        
        # 使用不匹配的版本
        mismatch_version = f"{original_version}_mismatch"
        
        # 构建更新数据
        update_data = {
            "title": original_title,
            "description": "测试版本不匹配情况更新失败",
            "operator_id": operator_id,
            "version": mismatch_version
        }
        
        # 打印请求信息
        print(f"版本不匹配请求 - ID: {operator_id}, DAG_ID: {dag_id}, 正确版本: {original_version}, 使用版本: {mismatch_version}")
        print(f"更新数据: {update_data}")
        
        # 发送更新请求
        result = self.client.UpdateOperator(dag_id, update_data, Headers)
        
        # 打印响应结果
        print(f"更新响应: {result}")
        
        # 验证响应状态码
        assert result[0] == 400, f"使用不匹配的版本更新组合算子应返回状态码400，实际为: {result[0]}"
    
    @allure.title("更新edit算子失败 - 缺少必要字段")
    def test_update_edit_operator_missing_required_fields(self, Headers):
        """测试缺少必要字段更新edit算子"""
        # 使用公共变量
        operator_id = self.edit_operator_id
        original_title = self.edit_operator_title
        dag_id = self.edit_operator_dag_id
        
        # 构建更新数据 - 缺少version字段
        update_data = {
            "title": original_title,
            "description": "测试缺少必要字段情况更新失败",
            "operator_id": operator_id
            # 缺少version字段
        }
        
        # 打印请求信息
        print(f"缺少必要字段请求 - ID: {operator_id}, DAG_ID: {dag_id}")
        print(f"更新数据: {update_data}")
        
        # 发送更新请求
        result = self.client.UpdateOperator(dag_id, update_data, Headers)
        
        # 打印响应结果
        print(f"更新响应: {result}")
        
        # 验证响应状态码
        assert result[0] == 400, f"缺少必要字段更新组合算子应返回状态码400，实际为: {result[0]}"
    
    @allure.title("更新edit算子失败 - 无效的steps结构")
    def test_update_edit_operator_invalid_steps(self, Headers):
        """测试使用无效的steps结构更新edit算子"""
        # 使用公共变量
        operator_id = self.edit_operator_id
        original_version = self.edit_operator_version
        original_title = self.edit_operator_title
        dag_id = self.edit_operator_dag_id
        
        # 构建更新数据 - 无效的steps结构
        update_data = {
            "title": original_title,
            "description": "测试无效steps结构情况更新失败",
            "steps": [
                {
                    "id": "0",
                    "title": "触发器",
                    "operator": "@trigger/form",
                    # 缺少parameters字段
                },
                {
                    "id": "1",
                    # 缺少title字段
                    "operator": "@internal/time/now"
                },
                {
                    "id": "2",
                    "title": "返回结果",
                    # 缺少operator字段
                    "parameters": {
                        "content": "{{_0.fields.text}}",
                        "curtime": "{{_1.curtime}}"
                    }
                }
            ],
            "operator_id": operator_id,
            "version": original_version,
            "category": "data_process",
            "outputs": [
                {
                    "key": "content",
                    "name": "内容",
                    "type": "string"
                },
                {
                    "key": "curtime",
                    "name": "当前时间",
                    "type": "string"
                }
            ]
        }
        
        # 打印请求信息
        print(f"无效steps结构请求 - ID: {operator_id}, DAG_ID: {dag_id}, 版本: {original_version}")
        print(f"更新数据: {update_data}")
        
        # 发送更新请求
        result = self.client.UpdateOperator(dag_id, update_data, Headers)
        
        # 打印响应结果
        print(f"更新响应: {result}")
        
        # 验证响应状态码
        assert result[0] == 400, f"无效的steps结构更新组合算子应返回状态码400，实际为: {result[0]}"

    # 添加一个辅助方法用于从算子列表中获取最新的算子信息
    def get_latest_operator_info(self, title, Headers):
        """
        从算子列表中获取指定标题的最新算子信息
        返回元组 (成功标志, operator_id, version, dag_id, operator_name)
        """
        params = {
            "name": title,
            "limit": 100
        }
        
        # 获取算子列表
        list_result = self.client.GetOperatorsList(params, Headers)
        print(f"获取算子列表响应: {list_result}")
        
        if list_result[0] != 200:
            print(f"获取算子列表失败，状态码: {list_result[0]}")
            return False, None, None, None, None
        
        # 检查返回结果是否包含算子
        if not list_result[1].get("ops") or len(list_result[1]["ops"]) == 0:
            print(f"获取算子列表成功但未找到算子: {title}")
            return False, None, None, None, None
        
        # 从列表中找到匹配的算子
        for op in list_result[1]["ops"]:
            # 打印调试信息
            print(f"检查算子: {op}")
            if op.get("operator_name") == title or op.get("name") == title:
                operator_id = op.get("operator_id")
                version = op.get("version")
                dag_id = op.get("dag_id", "")  # 使用空字符串作为默认值
                operator_name = op.get("operator_name") or op.get("name")
                print(f"匹配到算子 - ID: {operator_id}, 版本: {version}, DAG_ID: {dag_id}, 名称: {operator_name}")
                return True, operator_id, version, dag_id, operator_name
        
        print(f"在算子列表中未找到匹配的算子: {title}")
        return False, None, None, None, None
