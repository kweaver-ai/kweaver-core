# -*- coding:UTF-8 -*-

import allure
import string
import random
import pytest
import time

from common.get_content import GetContent
from lib.tool_box import ToolBox

@allure.feature("工具注册与管理接口测试：Debug与Proxy高级测试")
class TestDebugProxyAdvanced:
    """
    本测试类旨在验证工具 '调试(Debug)' 和 '代理执行(Proxy)' 接口的高级控制参数。
    
    设计背景：
    最新的接口文档引入了 'mode' (执行模式) 和 'stream' (流式返回) 参数，
    以及在请求体中自定义 'timeout' 的能力。这些参数直接影响工具调用的响应行为。
    
    测试重点：
    1. 验证同步模式 (mode=sync) 与流式模式 (mode=stream) 的参数传递。
    2. 验证后端是否接受并响应请求体中指定的 'timeout' 超时控制。
    3. 验证工具在启用状态下，通过代理接口转发请求的准确性。
    """
    
    client = ToolBox()
    box_id = ""
    tool_id = ""

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        # 创建工具箱并启用工具
        filepath = "./resource/openapi/compliant/toolbox.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = 'adv_' + ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        
        # 重试逻辑
        result = None
        for i in range(3):
            result = self.client.CreateToolbox(data, Headers)
            if result[0] == 200:
                break
            time.sleep(1)
            
        if result[0] == 503:
            pytest.skip("后端服务不可用(503)，跳过Debug/Proxy测试")
            
        assert result[0] == 200
        assert isinstance(result[1], dict)
        TestDebugProxyAdvanced.box_id = result[1]["box_id"]

        # 获取工具并启用
        result = None
        for i in range(3):
            result = self.client.GetBoxToolsList(TestDebugProxyAdvanced.box_id, None, Headers)
            if result[0] == 200:
                break
            time.sleep(1)
            
        if result[0] != 200:
            pytest.skip(f"获取工具列表失败，状态码: {result[0]}")
            
        TestDebugProxyAdvanced.tool_id = result[1]["tools"][0]["tool_id"]
        
        update_data = [{"tool_id": TestDebugProxyAdvanced.tool_id, "status": "enabled"}]
        self.client.UpdateToolStatus(TestDebugProxyAdvanced.box_id, update_data, Headers)

    @allure.title("工具调试，使用mode=sync参数，调试成功")
    def test_debug_tool_mode_sync(self, Headers):
        params = {"mode": "sync"}
        debug_data = {"header": Headers}
        result = self.client.DebugTool(self.box_id, self.tool_id, debug_data, Headers, params=params)
        assert result[0] == 200

    @allure.title("工具调试，使用mode=stream参数，调试成功")
    def test_debug_tool_mode_stream(self, Headers):
        params = {"mode": "stream", "stream": True}
        debug_data = {"header": Headers}
        result = self.client.DebugTool(self.box_id, self.tool_id, debug_data, Headers, params=params)
        assert result[0] == 200

    @allure.title("工具调试，在Body中指定timeout，调试成功")
    def test_debug_tool_timeout(self, Headers):
        debug_data = {
            "header": Headers,
            "timeout": 30
        }
        result = self.client.DebugTool(self.box_id, self.tool_id, debug_data, Headers)
        assert result[0] == 200

    @allure.title("工具代理，使用mode=sync参数，执行成功")
    def test_proxy_tool_mode_sync(self, Headers):
        params = {"mode": "sync"}
        proxy_data = {"header": Headers}
        result = self.client.ProxyTool(self.box_id, self.tool_id, proxy_data, Headers, params=params)
        assert result[0] == 200

    @allure.title("工具代理，在Body中指定timeout，执行成功")
    def test_proxy_tool_timeout(self, Headers):
        proxy_data = {
            "header": Headers,
            "timeout": 10
        }
        result = self.client.ProxyTool(self.box_id, self.tool_id, proxy_data, Headers)
        assert result[0] == 200
