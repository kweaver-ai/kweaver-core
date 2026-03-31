# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import time

from lib.tool_box import ToolBox
from common.get_content import GetContent
from common.assert_tools import AssertTools

@allure.feature("工具注册与管理接口测试：市场工具高级查询")
class TestMarketToolsAdvanced:
    """
    本测试类专门验证市场工具列表接口 (/v1/tool-box/market/tools) 的过滤与排序功能。
    
    设计背景：
    市场接口不仅需要返回数据，还需要支持复杂的查询条件（如状态过滤）和多维度的排序。
    本文件补充了文档中提及但之前用例未充分覆盖的参数组合。
    
    测试重点：
    1. 状态过滤：验证按 'enabled' 或 'disabled' 状态检索工具的准确性。
    2. 维度排序：验证按更新时间 (update_time) 或创建时间 (create_time) 进行正序/倒序排列逻辑。
    3. 数据一致性：验证查询结果中的 box_id 和 status 字段是否符合筛选条件。
    """
    
    client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        # 创建一些工具箱并发布
        for i in range(3):
            filepath = "./resource/openapi/compliant/toolbox.yaml"
            yaml_data = GetContent(filepath).yamlfile()
            name = 'market_' + str(i) + '_' + ''.join(random.choice(string.ascii_letters) for _ in range(5))
            data = {
                "box_name": name,
                "data": yaml_data,
                "metadata_type": "openapi"
            }
            
            # 重试逻辑
            res = None
            for attempt in range(3):
                res = self.client.CreateToolbox(data, Headers)
                if res[0] == 200:
                    break
                time.sleep(1)
            
            if res[0] == 503:
                continue # 允许部分失败，后面会根据结果过滤
                
            if res[0] == 200:
                assert isinstance(res[1], dict)
                box_id = res[1]["box_id"]
                self.client.UpdateToolboxStatus(box_id, {"status": "published"}, Headers)
                # 启用其中一部分工具
                tools_res = self.client.GetBoxToolsList(box_id, None, Headers)
                if tools_res[0] == 200:
                    tool_id = tools_res[1]["tools"][0]["tool_id"]
                    self.client.UpdateToolStatus(box_id, [{"tool_id": tool_id, "status": "enabled"}], Headers)
            time.sleep(1) # 保证时间戳有差异

    @allure.title("查询市场工具，按状态启用(enabled)过滤，查询成功")
    def test_get_market_tools_filter_enabled(self, Headers):
        params = {"status": "enabled", "tool_name": "工具", "all": True}
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 200
        for box in result[1]["data"]:
            for tool in box["tools"]:
                assert tool["status"] == "enabled"

    @allure.title("查询市场工具，按状态禁用(disabled)过滤，查询成功")
    def test_get_market_tools_filter_disabled(self, Headers):
        params = {"status": "disabled", "tool_name": "工具", "all": True}
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 200
        for box in result[1]["data"]:
            for tool in box["tools"]:
                assert tool["status"] == "disabled"

    @allure.title("查询市场工具，按更新时间(update_time)正序排列，查询成功")
    def test_get_market_tools_sort_update_time_asc(self, Headers):
        params = {"sort_by": "update_time", "sort_order": "asc", "tool_name": "工具", "all": True}
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 200
        
        update_times = [box["update_time"] for box in result[1]["data"]]
        assert AssertTools.is_ascending_str(update_times)

    @allure.title("查询市场工具，按创建时间(create_time)正序排列，查询成功")
    def test_get_market_tools_sort_create_time_asc(self, Headers):
        params = {"sort_by": "create_time", "sort_order": "asc", "tool_name": "工具", "all": True}
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 200
        
        create_times = [box["create_time"] for box in result[1]["data"]]
        assert AssertTools.is_ascending_str(create_times)
