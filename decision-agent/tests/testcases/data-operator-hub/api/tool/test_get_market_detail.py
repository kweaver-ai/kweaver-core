# -*- coding:UTF-8 -*-

import allure
import string
import random
import pytest
import uuid

from lib.tool_box import ToolBox
from common.get_content import GetContent

box_ids = []

@allure.feature("工具注册与管理接口测试：获取工具箱市场信息")
class TestGetMarketDetail:
    
    client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_ids
        for i in range(10):
            filepath = "./resource/openapi/compliant/test3.yaml"
            yaml_data = GetContent(filepath).yamlfile()
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            
            data = {
                "box_name": name,
                "data": yaml_data,
                "metadata_type": "openapi"
            }
            result = self.client.CreateToolbox(data, Headers)
            assert result[0] == 200
            box_id = result[1]["box_id"]
            box_ids.append(box_id)

        for i in range(6):
            update_data = {
                "status": "published"
            }
            result = self.client.UpdateToolboxStatus(box_ids[i], update_data, Headers)
            assert result[0] == 200

    @allure.title("批量获取工具箱服务市场详情，获取单个工具箱的某个字段信息，获取成功")
    def test_get_market_detail_01(self, Headers):
        global box_ids
        result = self.client.GetMarketDetail(box_ids[0], "box_name", Headers)
        assert result[0] == 200
        assert len(result[1]) == 1
        assert result[1][0]["box_id"] == box_ids[0]
        assert "box_name" in result[1][0]      

    @allure.title("批量获取工具箱服务市场详情，获取多个工具箱的多个字段信息，获取成功")
    def test_get_market_detail_02(self, Headers):
        global box_ids
        box_id = ",".join(box_ids[:5])
        fields = "box_name,box_desc,box_svc_url,status,category_type,category_name,is_internal,source,create_user,update_user,release_user,tools"
        result = self.client.GetMarketDetail(box_id, fields, Headers)
        assert result[0] == 200
        assert len(result[1]) == 5
        for box in result[1]:
            assert box["box_id"] in box_ids[:5]
            assert "box_name" in box
            assert "box_desc" in box
            assert "box_svc_url" in box
            assert "status" in box
            assert "category_type" in box
            assert "category_name" in box
            assert "is_internal" in box
            assert "source" in box
            assert "create_user" in box
            assert "update_user" in box
            assert "release_user" in box
            assert "tools" in box

    @allure.title("批量获取工具箱服务市场详情，工具箱不存在，返回空列表")
    def test_get_market_detail_03(self, Headers):
        box_id = str(uuid.uuid4())
        fields = "box_name,box_desc"
        result = self.client.GetMarketDetail(box_id, fields, Headers)
        assert result[0] == 200
        assert result[1] == []

    @allure.title("批量获取工具箱服务市场详情，部分工具箱不存在，仅返回存在的工具箱信息")
    def test_get_market_detail_04(self, Headers):
        global box_ids
        non_existent_id = str(uuid.uuid4())
        box_id = ",".join(box_ids[:3]) + "," + non_existent_id
        fields = "box_name,box_desc"
        result = self.client.GetMarketDetail(box_id, fields, Headers)
        assert result[0] == 200
        assert len(result[1]) == 3
        for box in result[1]:
            assert box["box_id"] in box_ids[:3]
            assert "box_name" in box
            assert "box_desc" in box

    @allure.title("批量获取工具箱服务市场详情，服务未发布，返回空列表")
    def test_get_market_detail_05(self, Headers):
        global box_ids
        
        fields = "box_name,box_desc"
        result = self.client.GetMarketDetail(box_ids[7], fields, Headers)
        assert result[0] == 200
        assert result[1] == []

    @allure.title("批量获取工具箱服务市场详情，部分服务未发布，仅返回已发布的工具箱信息")
    def test_get_market_detail_06(self, Headers):
        global box_ids
        
        # 组合已发布和未发布的工具箱ID
        box_id = ",".join(box_ids[:2]) + "," + box_ids[7]
        fields = "box_name,box_desc"
        result = self.client.GetMarketDetail(box_id, fields, Headers)
        assert result[0] == 200
        assert len(result[1]) == 2
        for box in result[1]:
            assert box["box_id"] in box_ids[:2]
            assert "box_name" in box
            assert "box_desc" in box

    @allure.title("批量获取工具箱服务市场详情，fields为无效字段，获取成功")
    def test_get_market_detail_07(self, Headers):
        global box_ids
        result = self.client.GetMarketDetail(box_ids[0], "invalid_field", Headers)
        assert result[0] == 200
        for box in result[1]:
            assert box["box_id"] == box_ids[0]
            assert "invalid_field" not in box

    @allure.title("批量获取工具箱服务市场详情，fields有部分无效字段，获取成功")
    def test_get_market_detail_08(self, Headers):
        global box_ids
        fields = "box_name,invalid_field,box_desc"
        result = self.client.GetMarketDetail(box_ids[0], fields, Headers)
        assert result[0] == 200
        for box in result[1]:
            assert box["box_id"] == box_ids[0]
            assert "box_name" in box
            assert "invalid_field" not in box
            assert "box_desc" in box
    