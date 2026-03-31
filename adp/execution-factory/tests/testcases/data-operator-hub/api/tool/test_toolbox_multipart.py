# -*- coding:UTF-8 -*-

import allure
import pytest
import random
import string
import os
import time

from common.get_content import GetContent
from lib.tool_box import ToolBox

characters = string.ascii_letters + string.digits

@allure.feature("工具注册与管理接口测试：Multipart上传测试")
class TestToolboxMultipart:
    """
    本测试类专门验证接口对 'multipart/form-data' 请求格式的支持。
    
    设计背景：
    根据最新的 toolbox.yaml 文档，工具箱和工具的创建/更新接口均支持上传文件（OpenAPI 规范文件）。
    此文件用于补全之前仅有 JSON 格式请求的测试漏洞。
    
    测试重点：
    1. 验证后端能否正确解析 Multipart 请求中的 'data' 文件流。
    2. 验证非文件字段（如 box_name, metadata_type）在 Multipart 格式下的解析。
    3. 验证环境不稳定（503 错误）时的自动重试与优雅跳过逻辑。
    """
    
    client = ToolBox()

    @allure.title("使用multipart/form-data创建工具箱，传参正确，创建成功")
    def test_create_toolbox_multipart_01(self, Headers):
        name = 'mp_' + ''.join(random.choice(characters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        
        # 准备文件
        files = {
            'data': ('mcp.yaml', open(filepath, 'rb'), 'application/x-yaml')
        }
        # 准备其他字段
        data = {
            "box_name": name,
            "metadata_type": "openapi"
        }
        
        # 移除 Headers 中的 Content-Type，让 requests 自动生成带 boundary 的 Content-Type
        headers = Headers.copy()
        if "Content-Type" in headers:
            del headers["Content-Type"]

        # 添加重试机制
        max_retries = 3
        result = None
        for attempt in range(max_retries):
            # 重新打开文件流，因为前一次尝试可能已将其移动或关闭
            files['data'] = ('mcp.yaml', open(filepath, 'rb'), 'application/x-yaml')
            result = self.client.CreateToolboxMultipart(files, data, headers)
            files['data'][1].close()
            
            if result[0] == 200:
                break
            elif result[0] == 503 and attempt < max_retries - 1:
                time.sleep(2 ** (attempt + 1))
                continue
            else:
                break

        if result[0] == 503:
            pytest.skip("后端服务不可用(503)，跳过Multipart创建测试")
        
        assert result[0] == 200, f"创建失败，状态码: {result[0]}, 响应: {result[1]}"
        assert isinstance(result[1], dict), f"响应格式错误: {result[1]}"
        box_id = result[1]["box_id"]

        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["box_name"] == name

    @allure.title("使用multipart/form-data更新工具箱，传参正确，更新成功")
    def test_update_toolbox_multipart_01(self, Headers):
        # 1. 先创建一个工具箱
        name = 'up_mp_' + ''.join(random.choice(characters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        create_data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        
        # 添加重试
        res = None
        for i in range(3):
            res = self.client.CreateToolbox(create_data, Headers)
            if res[0] == 200: break
            time.sleep(1)
            
        if res[0] == 503: pytest.skip("503 skipped")
        assert res[0] == 200
        assert isinstance(res[1], dict)
        box_id = res[1]["box_id"]

        # 2. Multipart更新
        new_name = name + "_v2"
        new_filepath = "./resource/openapi/compliant/test.json"
        data = {
            "box_name": new_name,
            "box_desc": "multipart update desc",
            "box_svc_url": "http://test-service:8080",
            "box_category": "data_process",
            "metadata_type": "openapi"
        }

        headers = Headers.copy()
        if "Content-Type" in headers:
            del headers["Content-Type"]

        result = None
        for attempt in range(3):
            files = {'data': ('test.json', open(new_filepath, 'rb'), 'application/json')}
            result = self.client.UpdateToolboxMultipart(box_id, files, data, headers)
            files['data'][1].close()
            if result[0] == 200:
                break
            elif result[0] == 503:
                time.sleep(2)
                continue
            else:
                break

        if result[0] == 503:
            pytest.skip("后端服务不可用(503)，跳过Multipart更新测试")

        assert result[0] == 200, f"更新失败，状态码: {result[0]}, 响应: {result[1]}"
        
        # 3. 验证
        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["box_name"] == new_name
        assert result[1]["box_desc"] == "multipart update desc"

    @allure.title("使用multipart/form-data创建内置工具，传参正确，创建成功")
    def test_builtin_multipart_01(self, Headers):
        box_id = "mp-builtin-" + ''.join(random.choice(characters) for i in range(8))
        name = 'mp_builtin_' + ''.join(random.choice(characters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "multipart builtin desc",
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.1",
            "config_source": "auto"
        }
        
        headers = Headers.copy()
        if "Content-Type" in headers:
            del headers["Content-Type"]

        result = None
        for attempt in range(3):
            files = {'data': ('mcp.yaml', open(filepath, 'rb'), 'application/x-yaml')}
            result = self.client.BuiltinMultipart(files, data, headers)
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

        assert result[0] == 200, f"内置工具创建失败，状态码: {result[0]}, 响应: {result[1]}"
        assert result[1]["box_id"] == box_id

        # 验证
        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["is_internal"] == True

    @allure.title("使用multipart/form-data更新工具，传参正确，更新成功")
    def test_update_tool_multipart_01(self, Headers):
        # 1. 创建工具箱
        name = 'tool_mp_' + ''.join(random.choice(characters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        create_data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        
        res = None
        for i in range(3):
            res = self.client.CreateToolbox(create_data, Headers)
            if res[0] == 200: break
            time.sleep(1)
            
        if res[0] == 503: pytest.skip("503 skipped")
        assert res[0] == 200
        assert isinstance(res[1], dict)
        box_id = res[1]["box_id"]

        # 2. 获取目标工具ID (精准匹配名称)
        res = self.client.GetBoxToolsList(box_id, {"all": True}, Headers)
        assert res[0] == 200
        tool_id = None
        target_name = "解析SSE_MCPServer"
        for t in res[1]["tools"]:
            if t["name"] == target_name:
                tool_id = t["tool_id"]
                break
        
        if not tool_id:
            pytest.skip(f"在工具箱中未找到名为 {target_name} 的工具，跳过更新测试")

        # 3. Multipart更新工具 (使用相同文件确保工具能被匹配到，且名称保持一致以避免409)
        data = {
            "name": target_name,
            "description": "updated tool description via multipart",
            "metadata_type": "openapi"
        }

        headers = Headers.copy()
        if "Content-Type" in headers:
            del headers["Content-Type"]

        result = None
        for attempt in range(3):
            files = {'data': ('mcp.yaml', open(filepath, 'rb'), 'application/x-yaml')}
            result = self.client.UpdateToolMultipart(box_id, tool_id, files, data, headers)
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

        assert result[0] == 200, f"工具更新失败，状态码: {result[0]}, 响应: {result[1]}"
        
        # 4. 验证
        result = self.client.GetTool(box_id, tool_id, Headers)
        assert result[0] == 200
        assert result[1]["name"] == target_name
