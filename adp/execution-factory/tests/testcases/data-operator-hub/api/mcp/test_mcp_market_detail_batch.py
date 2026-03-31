 # -*- coding:UTF-8 -*-

import allure
import uuid
import string
import random
import pytest

from lib.mcp import MCP

mcp_ids = []

@allure.feature("MCP服务市场接口测试：批量获取MCP服务市场详情")
class TestBatchMarketDetail:
    """批量获取MCP服务市场详情屏蔽headers、args、env信息"""
    client = MCP()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global mcp_ids
        # 创建MCP Server
        for i in range(8):
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            data = {
                "name": name,
                "description": "test mcp server",
                "mode": "sse",
                "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
                "headers": {
                    "Content-Type": "application/json"
                },
                "source": "custom",
                "category": "other_category",
                "command": "ls",
                "args": ["a"],
                "env": {
                    "name": "test"
                }
            }
            result = self.client.RegisterMCP(data, Headers)
            assert result[0] == 200
            mcp_id = result[1]["mcp_id"]
            mcp_ids.append(mcp_id)

        for i in range(6):
        # 发布MCP服务
            release_data = {
                    "status": "published"
                }
            result = self.client.MCPReleaseAction(mcp_ids[i], release_data, Headers)
            assert result[0] == 200

    @allure.title("批量获取MCP服务市场详情，获取单个mcp的某个字段信息，获取成功")
    def test_batch_market_detail_01(self, Headers):
        global mcp_ids
        result = self.client.BatchGetMCPMarketDetail(mcp_ids[0], "name", Headers)
        assert result[0] == 200
        assert len(result[1]) == 1
        assert result[1][0]["mcp_id"] == mcp_ids[0]
        assert "name" in result[1][0]

    @allure.title("批量获取MCP服务市场详情，获取多个mcp的多个字段信息，获取成功")
    def test_batch_market_detail_02(self, Headers):
        global mcp_ids
        ids = ",".join(mcp_ids[:5])
        fields = "name,description,source,category,mode"
        result = self.client.BatchGetMCPMarketDetail(ids, fields, Headers)
        assert result[0] == 200
        assert len(result[1]) == 5
        for mcp in result[1]:
            assert mcp["mcp_id"] in mcp_ids[:5]
            assert "name" in mcp
            assert "description" in mcp
            assert "source" in mcp
            assert "category" in mcp
            assert "mode" in mcp

    @allure.title("批量获取MCP服务市场详情，mcp_id不存在，返回空列表")
    def test_batch_market_detail_03(self, Headers):
        mcp_id = str(uuid.uuid4())
        fields = "name,description,category"
        result = self.client.BatchGetMCPMarketDetail(mcp_id, fields, Headers)
        assert result[0] == 200
        assert result[1] == []

    @allure.title("批量获取MCP服务市场详情，部分mcp_id不存在，仅返回存在的mcp信息")
    def test_batch_market_detail_04(self, Headers):
        global mcp_ids
        mcp_id = str(uuid.uuid4())
        ids = ",".join(mcp_ids[:5]) + "," + mcp_id
        fields = "name,description,category"
        result = self.client.BatchGetMCPMarketDetail(ids, fields, Headers)
        assert result[0] == 200
        assert len(result[1]) == 5
        for mcp in result[1]:
            assert mcp["mcp_id"] in mcp_ids[:5]
            assert "name" in mcp
            assert "description" in mcp
            assert "category" in mcp

    @allure.title("批量获取MCP服务市场详情，服务未发布，返回空列表")
    def test_batch_market_detail_05(self, Headers):
        global mcp_ids
        ids = ",".join(mcp_ids[6:])
        fields = "name,description"
        result = self.client.BatchGetMCPMarketDetail(ids, fields, Headers)
        assert result[0] == 200
        assert result[1] == []

    @allure.title("批量获取MCP服务市场详情，部分服务未发布或已下架，仅返回已发布的mcp信息")
    def test_batch_market_detail_06(self, Headers):
        global mcp_ids
        release_data = {
                    "status": "offline"
                }
        result = self.client.MCPReleaseAction(mcp_ids[5], release_data, Headers)
        assert result[0] == 200
        ids = ",".join(mcp_ids)
        fields = "name,description"
        result = self.client.BatchGetMCPMarketDetail(ids, fields, Headers)
        assert result[0] == 200
        assert len(result[1]) == 5
        for mcp in result[1]:
            assert mcp["mcp_id"] in mcp_ids[:5]
            assert "name" in mcp
            assert "description" in mcp

    @allure.title("批量获取MCP服务市场详情，fields为无效字段，获取失败")
    def test_batch_market_detail_07(self, Headers):
        global mcp_ids
        result = self.client.BatchGetMCPMarketDetail(mcp_ids[0], "id", Headers)
        assert result[0] == 400

    @allure.title("批量获取MCP服务市场详情，fields有部分无效字段，获取失败")
    def test_batch_market_detail_08(self, Headers):
        global mcp_ids
        result = self.client.BatchGetMCPMarketDetail(mcp_ids[0], "name,id", Headers)
        assert result[0] == 400
