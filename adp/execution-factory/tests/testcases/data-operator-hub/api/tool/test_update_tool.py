# -*- coding:UTF-8 -*-

import allure
import string
import random
import uuid
import pytest
import time

from common.get_content import GetContent
from lib.tool_box import ToolBox
from lib.operator import Operator

box_id = ""
tools_id = []

@allure.feature("工具注册与管理接口测试：更新工具")
class TestUpdateTool:
    
    client = ToolBox()
    client1 = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_id
        global tools_id

        # 创建工具箱，添加重试机制处理503错误
        filepath = "./resource/openapi/compliant/test.json"
        json_data = GetContent(filepath).jsonfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": json_data,
            "metadata_type": "openapi"
        }
        
        # 重试创建工具箱，最多重试3次
        max_retries = 3
        result = None
        for attempt in range(max_retries):
            result = self.client.CreateToolbox(data, Headers)
            if result[0] == 200:
                break
            elif result[0] == 503 and attempt < max_retries - 1:
                # 503错误时等待后重试
                wait_time = min(2 ** attempt, 5)  # 最多等待5秒
                time.sleep(wait_time)
                continue
            else:
                # 其他错误或最后一次重试失败
                break
        
        # 如果创建失败，根据错误类型处理
        if result[0] == 503:
            # 503服务不可用，跳过setup
            pytest.skip(f"后端服务不可用(503)，无法创建工具箱进行测试。响应: {result[1]}")
        elif result[0] != 200:
            # 其他错误，断言失败
            assert False, f"创建工具箱失败，状态码: {result[0]}, 响应: {result[1]}"
        
        # 确保result[1]是字典类型
        assert isinstance(result[1], dict), f"响应格式错误，期望字典，实际: {type(result[1])}, 内容: {result[1]}"
        box_id = result[1]["box_id"]

        # 获取工具箱内工具列表，添加重试机制
        result = None
        for attempt in range(max_retries):
            result = self.client.GetBoxToolsList(box_id, None, Headers)
            if result[0] == 200:
                break
            elif result[0] == 503 and attempt < max_retries - 1:
                wait_time = min(2 ** attempt, 5)
                time.sleep(wait_time)
                continue
            else:
                break
        
        if result[0] == 503:
            pytest.skip(f"后端服务不可用(503)，无法获取工具列表。响应: {result[1]}")
        elif result[0] != 200:
            assert False, f"获取工具列表失败，状态码: {result[0]}, 响应: {result[1]}"
        
        assert isinstance(result[1], dict), f"响应格式错误，期望字典，实际: {type(result[1])}"
        tools = result[1].get("tools", [])
        for tool in tools:
            tools_id.append(tool["tool_id"])

    @allure.title("更新工具，传参正确，更新成功")
    def test_update_tool_01(self, Headers):
        global box_id
        global tools_id

        # 更新工具
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        update_data = {
            "name": name,
            "description": "test tool update description",
            "metadata_type": "openapi",  # 添加必填字段
            "use_rule": "quis labore ipsum",
            "extend_info": {},
            "global_parameters": {
                "in": "query",
                "name": "www",
                "type": "string",
                "value": "pariatur est eu ex sed",
                "required": True,
                "description": "test desctiption"
            },
            "quota_control": {
                "quota_type": "ip",
                "quota_value": 1000,
                "time_window": {
                    "unit": "second",
                    "value": 1
                },
                "burst_capacity": 100,
                "overage_policy": "queue"
            }
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 200, f"更新工具失败，状态码: {result[0]}, 响应: {result[1]}, 请求数据: {update_data}"

    @allure.title("更新工具，工具箱不存在，更新失败")
    def test_update_tool_02(self, Headers):
        global tools_id

        box_id = str(uuid.uuid4())
        update_data = {
            "name": "test_update_tool",
            "description": "test tool update description"
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400

    @allure.title("更新工具，工具不存在，更新失败")
    def test_update_tool_03(self, Headers):
        global box_id

        tool_id = str(uuid.uuid4())
        update_data = {
            "name": "test_tool_update",
            "description": "test tool update description"
        }
        result = self.client.UpdateTool(box_id, tool_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("更新工具，工具名称已存在，更新失败")
    def test_update_tool_04(self, Headers):
        global box_id
        global tools_id

        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        update_data = {
            "name": name,
            "description": "test tool update description",
            "metadata_type": "openapi"  # 添加必填字段
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 200, f"第一次更新工具失败，状态码: {result[0]}, 响应: {result[1]}, 请求数据: {update_data}"

        result = self.client.UpdateTool(box_id, tools_id[1], update_data, Headers)
        assert result[0] == 409, f"第二次更新工具（期望409冲突）失败，状态码: {result[0]}, 响应: {result[1]}" 

    @allure.title("更新工具，必填参数name不传，更新失败")
    def test_update_tool_05(self, Headers):
        global box_id
        global tools_id

        update_data = {
            "description": "test tool update description"
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400        

    @allure.title("更新工具，必填参数description不传，更新失败")
    def test_update_tool_06(self, Headers):
        global box_id
        global tools_id

        update_data = {
            "name": "test_tool_update_name"
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400  

    @allure.title("更新工具，参数位置不在支持范围内，更新失败")
    def test_update_tool_07(self, Headers):
        global box_id
        global tools_id

        update_data = {
            "name": "test_tool_update_name",
            "description": "test tool update description",
            "global_parameters": {
                "in": "session",
                "name": "www",
                "type": "string",
                "value": "pariatur est eu ex sed",
                "required": True,
                "description": "test desctiption"
            }
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400  

    @allure.title("更新工具，参数类型不在支持范围内，更新失败")
    def test_update_tool_08(self, Headers):
        global box_id
        global tools_id

        update_data = {
            "name": "test_tool_update_name",
            "description": "test tool update description",
            "global_parameters": {
                "in": "header",
                "name": "www",
                "type": "json",
                "value": "pariatur est eu ex sed",
                "required": True,
                "description": "test desctiption"
            }
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400

    @allure.title("更新通过算子转换成的工具，更新成功")
    def test_update_tool_09(self, Headers):
        global box_id
        global tools_id

        # 注册算子
        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()

        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }

        result = self.client1.RegisterOperator(data, Headers)
        assert result[0] == 200
        operators = result[1]
        operator_id = operators[0]["operator_id"]
        operator_version = operators[0]["version"]

        # 转换算子为工具
        convert_data = {
            "box_id": box_id,
            "operator_id": operator_id,
            "operator_version": operator_version
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 200
        tool_id = result[1]["tool_id"]
        tools_id.append(tool_id)

        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        update_data = {
            "name": name,
            "description": "test tool update description",
            "metadata_type": "openapi"  # 添加必填字段
        }
        result = self.client.UpdateTool(box_id, tool_id, update_data, Headers)
        assert result[0] == 200, f"更新通过算子转换的工具失败，状态码: {result[0]}, 响应: {result[1]}, 请求数据: {update_data}, tool_id: {tool_id}"

        # 验证更新结果
        result = self.client.GetTool(box_id, tool_id, Headers)
        assert result[0] == 200, f"获取工具信息失败，状态码: {result[0]}, 响应: {result[1]}"
        assert isinstance(result[1], dict), f"响应格式错误，期望字典，实际: {type(result[1])}"
        assert result[1]["tool_id"] == tool_id
        assert result[1]["name"] == name
        assert result[1]["description"] == "test tool update description"

        # 算子不变
        result = self.client1.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["name"] != name

    @allure.title("更新工具元数据，metadata_type为其他类型，更新失败")
    def test_update_tool_10(self, Headers):
        global box_id
        global tools_id

        update_data = {
            "name": "test_tool_update_name_10",
            "description": "test tool update description",
            "metadata_type": "tool"
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400

    @allure.title("更新算子转换成的工具的元数据，更新失败")
    def test_update_tool_11(self, Headers):
        global box_id
        global tools_id

        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()
        update_data = {
            "name": "test_tool_update_name_11",
            "description": "test tool update description",
            "metadata_type": "openapi",
            "data": api_data
        }
        result = self.client.UpdateTool(box_id, tools_id[-1], update_data, Headers)
        assert result[0] == 405

    @allure.title("更新工具元数据，未匹配到当前工具，编辑失败")
    def test_update_tool_12(self, Headers):
        global box_id
        global tools_id

        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()
        update_data = {
            "name": "test_tool_update_name_12",
            "description": "test tool update description",
            "metadata_type": "openapi",
            "data": api_data
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 404

    @allure.title("更新工具元数据，openapi中包含多个工具，可匹配到当前工具，更新成功")
    def test_update_tool_13(self, Headers):
        global box_id
        global tools_id

        filepath = "./resource/openapi/compliant/test.json"
        api_data = GetContent(filepath).jsonfile()
        update_data = {
            "name": "test_tool_update_name_13",
            "description": "test tool update description",
            "metadata_type": "openapi",
            "data": api_data
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 200

    @allure.title("更新工具元数据，openapi中仅包含一个工具，可匹配到当前工具，更新成功")
    def test_update_tool_14(self, Headers):
        global box_id

        filepath = "./resource/openapi/compliant/tool.json"
        data = GetContent(filepath).jsonfile()
        tool_data = {
            "metadata_type": "openapi",
            "data": data
        }
        result = self.client.CreateTool(box_id, tool_data, Headers)
        assert result[0] == 200
        tool_id = result[1]["success_ids"][0]
        update_data = {
            "name": "test_tool_update_name_14",
            "description": "test tool update description",
            "metadata_type": "openapi",
            "data": data
        }
        result = self.client.UpdateTool(box_id, tool_id, update_data, Headers)
        assert result[0] == 200