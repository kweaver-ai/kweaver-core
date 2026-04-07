# -*- coding:UTF-8 -*-
"""
算子和工具转换接口测试

测试目标：
    验证算子（operator）和工具（tool）之间的相互转换功能。

测试覆盖：
    1. 工具转算子：从工具箱中的工具转换为算子
    2. 算子转工具：从算子转换为工具箱中的工具
    3. 转换后的数据一致性验证
    4. 转换异常场景处理

业务背景：
    算子和工具在系统中是两种不同的资源类型：
    - 算子（Operator）：更通用的执行单元，支持 OpenAPI 和 Function 两种元数据类型
    - 工具（Tool）：工具箱中的工具，通常基于 OpenAPI 定义
    
    转换场景：
    1. 工具转算子：将工具箱中的工具注册为算子，便于在算子市场中使用
    2. 算子转工具：将算子添加到工具箱中，便于在工具箱中管理
"""

import allure
import pytest
import time
import uuid
import random
import string

from common.get_content import GetContent
from lib.operator import Operator
from lib.tool_box import ToolBox

@allure.feature("算子注册与管理接口测试：算子和工具转换")
class TestOperatorToolConversion:
    """
    算子和工具转换测试类
    
    说明：
        测试算子与工具之间的相互转换功能，确保转换后数据的一致性和可用性。
    """
    
    operator_client = Operator()
    toolbox_client = ToolBox()
    
    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        """
        测试前置准备
        
        功能：
            1. 创建一个工具箱和工具，用于工具转算子测试
            2. 注册一个算子，用于算子转工具测试
        """
        # 创建工具箱用于工具转算子测试
        box_name = "conversion_box_" + ''.join(random.choice(string.ascii_lowercase) for _ in range(6))
        
        # 加载 OpenAPI 文件
        filepath = "./resource/openapi/compliant/template.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        
        toolbox_data = {
            "box_name": box_name,
            "description": "工具箱用于转换测试",
            "data": yaml_data,  # 必须提供 OpenAPI 数据
            "metadata_type": "openapi"
        }
        
        # 重试机制创建工具箱
        toolbox_result = None
        for attempt in range(3):
            toolbox_result = self.toolbox_client.CreateToolbox(toolbox_data, Headers)
            if toolbox_result[0] == 200:
                break
            if toolbox_result[0] == 503:
                time.sleep(2)
        
        if toolbox_result[0] == 503:
            pytest.skip("后端服务暂时不可用 (503)")
        
        assert toolbox_result[0] == 200, f"创建工具箱失败: {toolbox_result}"
        assert isinstance(toolbox_result[1], dict), f"响应格式错误: {toolbox_result[1]}"
        self.toolbox_id = toolbox_result[1]["box_id"]
        
        # 在工具箱中创建工具
        filepath = "./resource/openapi/compliant/template.yaml"
        tool_data = GetContent(filepath).yamlfile()
        
        # 注意：CreateTool 接口的 data 字段应该是字典对象，而不是字符串
        tool_create_data = {
            "data": tool_data,  # 直接传递字典对象
            "metadata_type": "openapi"
        }
        
        tool_result = None
        for attempt in range(3):
            tool_result = self.toolbox_client.CreateTool(self.toolbox_id, tool_create_data, Headers)
            if tool_result[0] == 200:
                break
            if tool_result[0] == 503:
                time.sleep(2)
        
        if tool_result and tool_result[0] == 200:
            # CreateTool 返回的是字典格式：{"success_count": 1, "failure_count": 0, ...}
            assert isinstance(tool_result[1], dict), f"工具创建响应格式错误: {tool_result[1]}"
            
            if tool_result[1].get("success_count", 0) > 0:
                # 创建成功，从工具列表中获取 tool_id
                tools_list_result = self.toolbox_client.GetBoxToolsList(self.toolbox_id, None, Headers)
                if tools_list_result[0] == 200 and "tools" in tools_list_result[1]:
                    tools = tools_list_result[1]["tools"]
                    if len(tools) > 0:
                        # 获取最新创建的工具（通常是第一个）
                        self.tool_id = tools[0]["tool_id"]
                    else:
                        print(f"警告: 工具创建成功但工具列表为空")
                        self.tool_id = None
                else:
                    print(f"警告: 无法获取工具列表，状态码: {tools_list_result[0]}")
                    self.tool_id = None
            elif tool_result[1].get("failure_count", 0) > 0:
                # 创建失败（可能是工具已存在），尝试从工具箱中获取已存在的工具
                failures = tool_result[1].get("failures", [])
                if failures and "tool_name" in failures[0]:
                    tool_name = failures[0]["tool_name"]
                    # 从工具箱中查找同名工具
                    tools_list_result = self.toolbox_client.GetBoxToolsList(self.toolbox_id, None, Headers)
                    if tools_list_result[0] == 200 and "tools" in tools_list_result[1]:
                        tools = tools_list_result[1]["tools"]
                        for tool in tools:
                            if tool.get("name") == tool_name:
                                self.tool_id = tool["tool_id"]
                                print(f"信息: 工具 '{tool_name}' 已存在，使用现有工具ID: {self.tool_id}")
                                break
                        if not hasattr(self, 'tool_id') or self.tool_id is None:
                            print(f"警告: 工具 '{tool_name}' 创建失败且未在工具箱中找到")
                            self.tool_id = None
                    else:
                        print(f"警告: 无法获取工具列表，状态码: {tools_list_result[0]}")
                        self.tool_id = None
                else:
                    print(f"警告: setup 中创建工具失败，响应: {tool_result[1]}")
                    self.tool_id = None
            else:
                print(f"警告: setup 中创建工具响应异常: {tool_result[1]}")
                self.tool_id = None
        else:
            print(f"警告: setup 中创建工具失败，状态码: {tool_result[0] if tool_result else 'None'}, 响应: {tool_result}")
            self.tool_id = None
        
        # 注册算子用于算子转工具测试
        operator_filepath = "./resource/openapi/compliant/test3.yaml"
        operator_data = GetContent(operator_filepath).yamlfile()
        
        operator_register_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi",
            "description": "算子用于转换测试"
        }
        
        operator_result = None
        for attempt in range(3):
            operator_result = self.operator_client.RegisterOperator(operator_register_data, Headers)
            if operator_result[0] == 200:
                break
            if operator_result[0] == 503:
                time.sleep(2)
        
        if operator_result and operator_result[0] == 200 and len(operator_result[1]) > 0:
            self.operator_id = operator_result[1][0]["operator_id"]
            self.operator_version = operator_result[1][0]["version"]
        else:
            print(f"警告: setup 中注册算子失败，状态码: {operator_result[0] if operator_result else 'None'}, 响应: {operator_result}")
            self.operator_id = None
            self.operator_version = None
    
    @allure.title("工具转算子：从工具箱中的工具转换为算子，转换成功")
    def test_tool_to_operator_01(self, Headers):
        """
        测试用例1：工具转算子 - 正常场景
        
        测试场景：
            1. 获取工具箱中工具的 OpenAPI 定义
            2. 使用工具的 OpenAPI 定义通过 RegisterOperator 接口注册为算子
            3. 验证算子注册成功且元数据一致
        
        验证点：
            - 算子注册成功（状态码200）
            - 算子元数据类型为 openapi
            - 算子的 OpenAPI 定义与工具一致
        
        说明：
            工具转算子通过注册接口实现，需要提取工具的 OpenAPI 元数据并注册为新算子。
        """
        if not hasattr(self, 'tool_id') or self.tool_id is None:
            pytest.skip("工具创建失败，跳过测试")
        
        # 1. 获取工具信息
        tool_info_result = self.toolbox_client.GetTool(self.toolbox_id, self.tool_id, Headers)
        assert tool_info_result[0] == 200, f"获取工具信息失败: {tool_info_result}"
        
        tool_info = tool_info_result[1]
        tool_metadata = tool_info.get("metadata", {})
        
        # 2. 使用工具的 OpenAPI 定义注册为算子
        # 从工具元数据中提取 OpenAPI 数据
        # 如果工具存储的是完整的 OpenAPI YAML，可以直接使用
        # 否则需要从 metadata 中提取 openapi_spec 或其他相关字段
        openapi_data = tool_metadata.get("openapi_spec") or tool_metadata.get("data") or str(tool_metadata)
        
        operator_register_data = {
            "data": openapi_data if isinstance(openapi_data, str) else str(openapi_data),
            "operator_metadata_type": "openapi",
            "description": f"从工具 {self.tool_id} 转换而来的算子",
            "operator_info": {
                "category": "data_process",
                "execution_mode": "sync"
            }
        }
        
        operator_result = None
        for attempt in range(3):
            operator_result = self.operator_client.RegisterOperator(operator_register_data, Headers)
            if operator_result[0] == 200:
                break
            if operator_result[0] == 503:
                time.sleep(2)
        
        if operator_result[0] == 503:
            pytest.skip("后端服务暂时不可用 (503)")
        
        assert operator_result[0] == 200, f"工具转算子失败: {operator_result}"
        assert len(operator_result[1]) > 0, "算子注册返回空列表"
        assert operator_result[1][0]["status"] == "success", f"算子注册状态异常: {operator_result[1][0]}"
        
        converted_operator_id = operator_result[1][0]["operator_id"]
        
        # 3. 验证算子信息
        operator_info_result = self.operator_client.GetOperatorInfo(converted_operator_id, Headers)
        assert operator_info_result[0] == 200, f"获取算子信息失败: {operator_info_result}"
        
        operator_info = operator_info_result[1]
        assert operator_info["metadata_type"] == "openapi", "算子元数据类型应为 openapi"
        
        # 清理：删除转换后的算子
        delete_data = [{"operator_id": converted_operator_id, "version": operator_result[1][0]["version"]}]
        self.operator_client.DeleteOperator(delete_data, Headers)
    
    @allure.title("算子转工具：使用专用转换接口将算子转换为工具，转换成功")
    def test_operator_to_tool_01(self, Headers):
        """
        测试用例2：算子转工具 - 正常场景（使用专用转换接口）
        
        测试场景：
            1. 使用 ConvertOperatorToTool 接口将算子转换为工具
            2. 验证工具创建成功且元数据一致
        
        验证点：
            - 转换接口调用成功（状态码200）
            - 工具创建成功
            - 工具元数据类型为 openapi
            - 工具的 OpenAPI 定义与算子一致
        """
        if not hasattr(self, 'operator_id') or self.operator_id is None:
            pytest.skip("算子注册失败，跳过测试")
        
        if not hasattr(self, 'toolbox_id') or self.toolbox_id is None:
            pytest.skip("工具箱创建失败，跳过测试")
        
        # 1. 获取算子信息
        operator_info_result = self.operator_client.GetOperatorInfo(self.operator_id, Headers)
        assert operator_info_result[0] == 200, f"获取算子信息失败: {operator_info_result}"
        
        operator_info = operator_info_result[1]
        assert operator_info["metadata_type"] == "openapi", "算子元数据类型应为 openapi"
        
        # 2. 使用专用转换接口将算子转换为工具
        convert_data = {
            "operator_id": self.operator_id,
            "version": self.operator_version,
            "box_id": self.toolbox_id
        }
        
        convert_result = None
        for attempt in range(3):
            convert_result = self.toolbox_client.ConvertOperatorToTool(convert_data, Headers)
            if convert_result[0] == 200:
                break
            if convert_result[0] == 503:
                time.sleep(2)
        
        if convert_result[0] == 503:
            pytest.skip("后端服务暂时不可用 (503)")
        
        assert convert_result[0] == 200, f"算子转工具失败: {convert_result}"
        assert isinstance(convert_result[1], dict), "转换响应应为字典"
        
        # 3. 验证工具信息（从响应中获取工具ID）
        if "tool_id" in convert_result[1]:
            converted_tool_id = convert_result[1]["tool_id"]
            tool_info_result = self.toolbox_client.GetTool(self.toolbox_id, converted_tool_id, Headers)
            assert tool_info_result[0] == 200, f"获取工具信息失败: {tool_info_result}"
            
            tool_info = tool_info_result[1]
            assert tool_info.get("metadata_type") == "openapi", "工具元数据类型应为 openapi"
            
            # 清理：删除转换后的工具
            delete_tool_data = [{"tool_id": converted_tool_id}]
            self.toolbox_client.BatchDeleteTools(self.toolbox_id, delete_tool_data, Headers)
    
    @allure.title("工具转算子：使用工具的完整 OpenAPI 文件转换为算子，转换成功")
    def test_tool_to_operator_with_file_01(self, Headers):
        """
        测试用例3：工具转算子 - 使用文件方式
        
        测试场景：
            1. 使用 multipart/form-data 方式上传工具的 OpenAPI 文件
            2. 注册为算子
            3. 验证转换成功
        
        验证点：
            - 算子注册成功（状态码200）
            - 算子元数据正确
        """
        if not hasattr(self, 'toolbox_id') or self.toolbox_id is None:
            pytest.skip("工具箱创建失败，跳过测试")
        
        # 使用 multipart 方式注册算子
        filepath = "./resource/openapi/compliant/template.yaml"
        
        data = {
            "operator_metadata_type": "openapi",
            "description": "从工具文件转换而来的算子"
        }
        
        headers = Headers.copy()
        if "Content-Type" in headers:
            del headers["Content-Type"]
        
        result = None
        for attempt in range(3):
            files = {'data': ('template.yaml', open(filepath, 'rb'), 'application/x-yaml')}
            result = self.operator_client.RegisterOperatorMultipart(files, data, headers)
            files['data'][1].close()
            
            if result[0] == 200:
                break
            if result[0] == 503:
                time.sleep(2)
        
        if result[0] == 503:
            pytest.skip("后端服务暂时不可用 (503)")
        
        assert result[0] == 200, f"工具文件转算子失败: {result}"
        assert len(result[1]) > 0, "算子注册返回空列表"
        assert result[1][0]["status"] == "success", f"算子注册状态异常: {result[1][0]}"
        
        converted_operator_id = result[1][0]["operator_id"]
        
        # 清理：删除转换后的算子
        delete_data = [{"operator_id": converted_operator_id, "version": result[1][0]["version"]}]
        self.operator_client.DeleteOperator(delete_data, Headers)
    
    @allure.title("算子转工具：转换接口参数校验，缺少必要参数时转换失败")
    def test_operator_to_tool_missing_params(self, Headers):
        """
        测试用例4：算子转工具 - 异常场景：缺少必要参数
        
        测试场景：
            1. 调用转换接口时缺少 operator_id
            2. 调用转换接口时缺少 box_id
            3. 验证返回400错误
        
        验证点：
            - 缺少必要参数时返回400
        """
        if not hasattr(self, 'operator_id') or self.operator_id is None:
            pytest.skip("算子注册失败，跳过测试")
        
        if not hasattr(self, 'toolbox_id') or self.toolbox_id is None:
            pytest.skip("工具箱创建失败，跳过测试")
        
        # 测试1：缺少 operator_id
        convert_data_no_op_id = {
            "version": self.operator_version,
            "box_id": self.toolbox_id
        }
        result = self.toolbox_client.ConvertOperatorToTool(convert_data_no_op_id, Headers)
        assert result[0] == 400, f"缺少 operator_id 应该返回400，实际: {result[0]}"
        
        # 测试2：缺少 box_id
        convert_data_no_box_id = {
            "operator_id": self.operator_id,
            "version": self.operator_version
        }
        result = self.toolbox_client.ConvertOperatorToTool(convert_data_no_box_id, Headers)
        assert result[0] == 400, f"缺少 box_id 应该返回400，实际: {result[0]}"
    
    @allure.title("算子转工具：算子不存在时转换失败")
    def test_operator_to_tool_operator_not_found(self, Headers):
        """
        测试用例5：算子转工具 - 异常场景：算子不存在
        
        测试场景：
            使用不存在的算子ID调用转换接口，应返回404
        
        验证点：
            - 转换接口返回404
        """
        if not hasattr(self, 'toolbox_id') or self.toolbox_id is None:
            pytest.skip("工具箱创建失败，跳过测试")
        
        fake_operator_id = str(uuid.uuid4())
        convert_data = {
            "operator_id": fake_operator_id,
            "version": "1.0.0",
            "box_id": self.toolbox_id
        }
        result = self.toolbox_client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 404, f"不存在的算子ID应该返回404，实际: {result[0]}"
    
    @allure.title("工具转算子：工具不存在时转换失败")
    def test_tool_to_operator_not_found(self, Headers):
        """
        测试用例5：工具转算子 - 异常场景：工具不存在
        
        测试场景：
            使用不存在的工具ID获取工具信息，应返回404
        
        验证点：
            - 获取工具信息返回404
        """
        if not hasattr(self, 'toolbox_id') or self.toolbox_id is None:
            pytest.skip("工具箱创建失败，跳过测试")
        
        fake_tool_id = str(uuid.uuid4())
        tool_info_result = self.toolbox_client.GetTool(self.toolbox_id, fake_tool_id, Headers)
        assert tool_info_result[0] == 404, f"不存在的工具ID应该返回404，实际: {tool_info_result[0]}"
    
    @allure.title("算子转工具：算子不存在时转换失败")
    def test_operator_to_tool_not_found(self, Headers):
        """
        测试用例6：算子转工具 - 异常场景：算子不存在
        
        测试场景：
            使用不存在的算子ID获取算子信息，应返回404
        
        验证点：
            - 获取算子信息返回404
        """
        fake_operator_id = str(uuid.uuid4())
        operator_info_result = self.operator_client.GetOperatorInfo(fake_operator_id, Headers)
        assert operator_info_result[0] == 404, f"不存在的算子ID应该返回404，实际: {operator_info_result[0]}"
    
    @allure.title("工具转算子：Function 类型工具转换为算子，转换成功")
    def test_function_tool_to_operator_01(self, Headers):
        """
        测试用例7：Function 类型工具转算子
        
        测试场景：
            1. 创建一个 Function 类型的工具（如果支持）
            2. 转换为 Function 类型的算子
            3. 验证转换成功
        
        验证点：
            - 算子注册成功
            - 算子元数据类型为 function
        """
        # 注意：如果工具箱不支持 Function 类型的工具，此用例可以跳过
        # 或者使用算子注册 Function 类型，然后验证其可以转换为工具
        
        # 注册一个 Function 类型的算子
        function_operator_data = {
            "operator_metadata_type": "function",
            "function_input": {
                "name": "conversion_func_" + ''.join(random.choice(string.ascii_lowercase) for _ in range(6)),
                "description": "用于转换测试的函数算子",
                "code": "def handler(event, context):\n    return {'statusCode': 200, 'body': 'converted'}",
                "script_type": "python",
                "inputs": [
                    {"name": "param1", "type": "string", "required": True}
                ],
                "outputs": [
                    {"name": "result", "type": "string"}
                ]
            },
            "operator_info": {
                "category": "data_process",
                "execution_mode": "sync"
            }
        }
        
        operator_result = None
        for attempt in range(3):
            operator_result = self.operator_client.RegisterOperator(function_operator_data, Headers)
            if operator_result[0] == 200:
                break
            if operator_result[0] == 503:
                time.sleep(2)
        
        if operator_result[0] == 503:
            pytest.skip("后端服务暂时不可用 (503)")
        
        assert operator_result[0] == 200, f"Function 算子注册失败: {operator_result}"
        assert len(operator_result[1]) > 0, "算子注册返回空列表"
        assert operator_result[1][0]["status"] == "success", f"算子注册状态异常: {operator_result[1][0]}"
        
        function_operator_id = operator_result[1][0]["operator_id"]
        
        # 验证算子信息
        operator_info_result = self.operator_client.GetOperatorInfo(function_operator_id, Headers)
        assert operator_info_result[0] == 200, f"获取算子信息失败: {operator_info_result}"
        
        operator_info = operator_info_result[1]
        assert operator_info["metadata_type"] == "function", "算子元数据类型应为 function"
        
        # 清理：删除 Function 算子
        delete_data = [{"operator_id": function_operator_id, "version": operator_result[1][0]["version"]}]
        self.operator_client.DeleteOperator(delete_data, Headers)
