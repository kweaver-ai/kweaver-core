# -*- coding:UTF-8 -*-

import allure
import pytest
import random
import string
import os
import time
import uuid

from common.get_content import GetContent
from lib.operator import Operator

characters = string.ascii_letters + string.digits

@allure.feature("算子注册与管理接口测试：Multipart上传测试")
class TestOperatorMultipart:
    """
    本测试类专门验证算子接口对 'multipart/form-data' 请求格式的支持。
    
    设计背景：
    根据最新的 operator.yaml 文档，算子编辑接口明确支持 multipart/form-data。
    注册接口虽然文档未详尽列出 multipart，但考虑到包含二进制 data 字段，
    本用例一并验证其对 multipart 的兼容性，确保测试覆盖度。
    
    测试重点：
    1. 验证后端能否正确解析 Multipart 请求中的 'data' 文件流。
    2. 验证非文件字段（如 operator_id, metadata_type）在 Multipart 格式下的解析。
    3. 验证环境不稳定时的自动重试逻辑。
    """
    
    client = Operator()

    @allure.title("使用multipart/form-data编辑算子，传参正确，编辑成功")
    def test_edit_operator_multipart_01(self, Headers):
        # 1. 先注册一个算子
        filepath = "./resource/openapi/compliant/test1.json"
        api_data = GetContent(filepath).jsonfile()
        reg_data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        
        re = None
        for i in range(3):
            re = self.client.RegisterOperator(reg_data, Headers)
            if re[0] == 200: break
            time.sleep(1)
            
        if re[0] != 200:
            pytest.skip(f"注册算子失败，跳过编辑测试: {re}")
            
        operator_id = re[1][0]["operator_id"]
        original_version = re[1][0]["version"]

        # 2. Multipart 编辑 (使用相同文件确保路径匹配，只修改描述)
        new_filepath = "./resource/openapi/compliant/test1.json"
        data = {
            "operator_id": operator_id,
            "name": "mp_edit_name_" + ''.join(random.choice(characters) for i in range(5)),
            "description": "multipart edit desc",
            "metadata_type": "openapi"
        }

        # 准备文件
        files = {
            'data': ('test.json', open(new_filepath, 'rb'), 'application/json')
        }

        # 移除 Headers 中的 Content-Type，让 requests 自动生成带 boundary 的 Content-Type
        headers = Headers.copy()
        if "Content-Type" in headers:
            del headers["Content-Type"]

        result = None
        for attempt in range(3):
            # 重新打开文件流
            files['data'] = ('test.json', open(new_filepath, 'rb'), 'application/json')
            result = self.client.EditOperatorMultipart(files, data, headers)
            files['data'][1].close()
            
            if result[0] == 200:
                break
            elif result[0] == 503:
                time.sleep(2)
                continue
            else:
                break

        if result[0] == 503:
            pytest.skip("后端服务不可用(503)，跳过Multipart编辑测试")
            
        assert result[0] == 200, f"编辑失败，状态码: {result[0]}, 响应: {result[1]}"
        assert result[1]["operator_id"] == operator_id
        assert result[1]["status"] == "editing"
        # 修改了 data，应该生成新版本
        assert result[1]["version"] != original_version

    @allure.title("使用multipart/form-data注册内置算子，传参正确，注册成功")
    def test_builtin_operator_multipart_01(self, Headers):
        operator_id = str(uuid.uuid4())
        name = 'mp_builtin_op_' + ''.join(random.choice(characters) for i in range(8))
        # 使用 template.yaml 确保只包含一个算子定义
        filepath = "./resource/openapi/compliant/template.yaml"
        
        data = {
            "operator_id": operator_id,
            "name": name,
            "metadata_type": "openapi",
            "description": "builtin multipart desc",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "internal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        
        headers = Headers.copy()
        if "Content-Type" in headers:
            del headers["Content-Type"]

        result = None
        for attempt in range(3):
            files = {'data': ('template.yaml', open(filepath, 'rb'), 'application/x-yaml')}
            result = self.client.RegisterBuiltinOperatorMultipart(files, data, headers)
            files['data'][1].close()
            if result[0] == 200:
                break
            elif result[0] == 503:
                time.sleep(2)
                continue
            else:
                break

        if result[0] == 503:
            pytest.skip("503 skipped")

        assert result[0] == 200, f"内置算子注册失败，状态码: {result[0]}, 响应: {result[1]}"
        assert result[1]["operator_id"] == operator_id
        assert result[1]["status"] == "success"
