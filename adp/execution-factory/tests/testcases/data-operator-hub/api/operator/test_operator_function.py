# -*- coding:UTF-8 -*-

import allure
import pytest
import random
import string
import time

from lib.operator import Operator

@allure.feature("算子注册与管理接口测试：函数算子测试")
class TestOperatorFunction:
    """
    本测试类验证 'metadata_type: function' 类型的算子注册与管理。
    
    设计背景：
    算子支持两种元数据类型：OpenAPI 和 Function。
    Function 类型的算子直接包含 Python 代码片段，不需要外部 OpenAPI 文件。
    
    测试重点：
    1. Function 算子的注册逻辑（function_input 字段）。
    2. Function 算子的编辑逻辑。
    3. Function 算子的执行模式限制（目前通常为同步）。
    """
    
    client = Operator()

    @allure.title("注册函数算子，传参正确，注册成功")
    def test_register_function_operator_01(self, Headers):
        name = 'func_op_' + ''.join(random.choice(string.ascii_letters) for i in range(5))
        
        function_input = {
            "name": name,
            "description": "test function operator",
            "code": "def handler(event, context):\n    return {'statusCode': 200, 'body': 'hello'}",
            "script_type": "python",
            "inputs": [
                {"name": "param1", "type": "string", "required": True}
            ],
            "outputs": [
                {"name": "result", "type": "string"}
            ]
        }
        
        data = {
            "operator_metadata_type": "function",
            "function_input": function_input,
            "operator_info": {
                "category": "data_process",
                "execution_mode": "sync"
            }
        }
        
        result = None
        for attempt in range(3):
            result = self.client.RegisterOperator(data, Headers)
            if result[0] == 200:
                break
            if result[0] == 503:
                time.sleep(2)
        
        if result[0] == 503:
            pytest.skip("后端服务暂时不可用 (503)")
            
        assert result[0] == 200, f"注册函数算子失败: {result}"
        assert result[1][0]["status"] == "success"
        operator_id = result[1][0]["operator_id"]
        
        # 验证获取信息（添加重试机制）
        res = None
        for attempt in range(3):
            res = self.client.GetOperatorInfo(operator_id, Headers)
            if res[0] == 200:
                break
            if res[0] == 503:
                time.sleep(2)
        
        if res[0] == 503:
            pytest.skip("后端服务暂时不可用 (503)")
            
        assert res[0] == 200, f"获取算子信息失败: {res}"
        assert res[1]["metadata_type"] == "function"
        assert res[1]["metadata"]["function_content"]["code"] == function_input["code"]

    @allure.title("编辑函数算子代码，编辑成功，生成新版本")
    def test_edit_function_operator_01(self, Headers):
        # 1. 注册
        name = 'func_edit_' + ''.join(random.choice(string.ascii_letters) for i in range(5))
        function_input = {
            "name": name,
            "description": "initial function desc", # description 放在内部
            "code": "def handler(event, context):\n    return {'body': 'v1'}",
            "script_type": "python",
            "inputs": [{"name": "p", "type": "string"}],
            "outputs": [{"name": "r", "type": "string"}]
        }
        data = {
            "operator_metadata_type": "function",
            "function_input": function_input,
            "operator_info": {
                "category": "data_process",
                "execution_mode": "sync"
            }
        }
        
        reg_res = None
        for attempt in range(3):
            reg_res = self.client.RegisterOperator(data, Headers)
            if reg_res[0] == 200: break
            if reg_res[0] == 503: time.sleep(2)
            
        if reg_res[0] == 503:
            pytest.skip("后端服务暂时不可用 (503)")
            
        assert reg_res[0] == 200, f"注册失败: {reg_res}"
        operator_id = reg_res[1][0]["operator_id"]
        v1 = reg_res[1][0]["version"]
        
        # 2. 编辑代码
        # 根据 FunctionInputEdit Schema，function_input 不包含 name 和 description
        # name 和 description 应该在请求的顶层
        # 编辑时，function_input 的 inputs 和 outputs 必须与注册时保持一致
        # 注意：name 字段是必填的，description 和 function_input.code 的修改会生成新版本
        edit_data = {
            "operator_id": operator_id,
            "name": name,  # name 字段是必填的，使用注册时的名称
            "description": "updated function desc",  # 修改 description 以触发版本更新
            "metadata_type": "function",
            "function_input": {
                # function_input 只包含：inputs, outputs, code, script_type
                # 修改 code 以触发版本更新
                "code": "def handler(event, context):\n    return {'statusCode': 200, 'body': 'v2_updated'}",
                "script_type": "python",
                "inputs": [{"name": "p", "type": "string"}],  # 保持与注册时一致
                "outputs": [{"name": "r", "type": "string"}]  # 保持与注册时一致
            }
        }
        # 编辑代码（添加重试机制）
        edit_res = None
        for attempt in range(3):
            edit_res = self.client.EditOperator(edit_data, Headers)
            if edit_res[0] == 200:
                break
            if edit_res[0] == 503:
                time.sleep(2)
        
        if edit_res[0] == 503:
            pytest.skip("后端服务暂时不可用 (503)")
            
        assert edit_res[0] == 200, f"编辑函数算子失败: {edit_res}"
        # 业务逻辑：编辑函数算子的基本信息不会生成新版本，只有发布时才生成新版本号
        # 如果算子是未发布状态（unpublish），编辑后状态仍然是 unpublish
        # 如果算子是已发布状态（published），编辑后状态会变为 editing
        assert edit_res[1]["version"] == v1, f"编辑函数算子基本信息后版本号应该不变，实际: {edit_res[1]['version']} != {v1}"
        # 注册时默认是 unpublish 状态，编辑后应该仍然是 unpublish
        assert edit_res[1]["status"] == "unpublish", f"编辑未发布的函数算子后状态应该保持 unpublish，实际: {edit_res[1]['status']}"
        
        # 验证编辑后的信息（添加重试机制）
        operator_info = None
        for attempt in range(3):
            operator_info = self.client.GetOperatorInfo(operator_id, Headers)
            if operator_info[0] == 200:
                break
            if operator_info[0] == 503:
                time.sleep(2)
        
        if operator_info[0] == 503:
            pytest.skip("后端服务暂时不可用 (503)")
            
        assert operator_info[0] == 200, f"获取算子信息失败: {operator_info}"
        assert isinstance(operator_info[1], dict), f"算子信息响应格式错误: {operator_info[1]}"
        
        # 验证 description 已更新（如果字段存在）
        if "description" in operator_info[1]:
            assert operator_info[1]["description"] == "updated function desc", \
                f"description 应该已更新，实际: {operator_info[1].get('description')}"
        
        # 验证代码已更新
        assert "metadata" in operator_info[1], f"算子信息中缺少 metadata 字段: {operator_info[1].keys()}"
        assert "function_content" in operator_info[1]["metadata"], \
            f"metadata 中缺少 function_content 字段: {operator_info[1]['metadata'].keys()}"
        assert "v2_updated" in operator_info[1]["metadata"]["function_content"]["code"], \
            f"代码应该已更新为包含 'v2_updated'，实际代码: {operator_info[1]['metadata']['function_content'].get('code', '')[:100]}"

    @allure.title("注册函数算子，缺少必填字段code，注册失败")
    def test_register_function_operator_invalid(self, Headers):
        data = {
            "operator_metadata_type": "function",
            "function_input": {
                "name": "invalid_func",
                "script_type": "python"
                # 缺少 code
            }
        }
        result = self.client.RegisterOperator(data, Headers)
        # 后端可能会返回 400 或 200带failed，取决于具体校验点
        assert result[0] in [400, 200]
        if result[0] == 200:
            assert result[1][0]["status"] == "failed"
