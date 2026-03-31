# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import uuid

from lib.mcp import MCP
from lib.mcp_internal import InternalMCP

@allure.feature("算子平台业务域测试：MCP业务域测试")
class TestMCPDomain:
    client = MCP()
    client1 = InternalMCP()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, TestDomainData, DomainPrepare):
        domain_list, headers, _, _, mcp_list = TestDomainData
        user_list = DomainPrepare[0]
        TestMCPDomain.domain_list = domain_list
        TestMCPDomain.pub_headers = headers[0]
        TestMCPDomain.a_headers = headers[1]
        TestMCPDomain.b_headers = headers[2]
        TestMCPDomain.mcp_list = mcp_list  # 保存完整的mcp_list
        TestMCPDomain.pub_mcp_list = mcp_list[0:1]  # 公共域自定义mcp
        TestMCPDomain.A_mcp_list = mcp_list[2:3]  # A域自定义mcp
        TestMCPDomain.B_mcp_list = mcp_list[4:5]  # B域自定义mcp
        TestMCPDomain.user_list = user_list

    @allure.title("有新建权限，在公共业务域和自身所属业务域新建mcp，新建成功")
    def test_mcp_domain_01(self):
        headers1 = TestMCPDomain.a_headers.copy()
        headers1["x-business-domain"] = "bd_public"
        for headers in [headers1, TestMCPDomain.a_headers]:
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            create_data = {
                "name": name,
                "description": "test mcp server",
                "mode": "sse",
                "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
                "source": "custom",
                "category": "data_analysis"
            }
            result = self.client.RegisterMCP(create_data, headers)
            assert result[0] == 200

    @pytest.mark.skip
    @allure.title("有新建权限，在其他业务域新建mcp，新建失败")
    def test_mcp_domain_02(self):
        # 用户a1尝试在业务域B创建mcp，应该失败
        headers = TestMCPDomain.a_headers.copy()
        headers["x-business-domain"] = TestMCPDomain.domain_list[1]  # 业务域B
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        create_data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "source": "custom",
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(create_data, headers)
        assert result[0] == 403

    @allure.title("获取mcp列表，在公共业务域，仅能获取到公共业务域的mcp")
    def test_mcp_domain_03(self):
        # 在公共业务域获取mcp列表
        params = {"all": "true"}
        result = self.client.GetMCPList(params, TestMCPDomain.pub_headers)
        assert result[0] == 200
        mcp_ids = [mcp["mcp_id"] for mcp in result[1]["data"]]
        # 应该包含公共域的mcp
        for mcp_id in TestMCPDomain.pub_mcp_list:
            assert mcp_id in mcp_ids
        # 不应该包含A域或B域的mcp
        for mcp_id in TestMCPDomain.A_mcp_list:
            assert mcp_id not in mcp_ids
        for mcp_id in TestMCPDomain.B_mcp_list:
            assert mcp_id not in mcp_ids

    @allure.title("获取mcp列表，在自身所属业务域，仅能获取到本业务域的mcp")
    def test_mcp_domain_04(self):
        # 在业务域A获取mcp列表
        params = {"all": "true"}
        result = self.client.GetMCPList(params, TestMCPDomain.a_headers)
        assert result[0] == 200
        mcp_ids = [mcp["mcp_id"] for mcp in result[1]["data"]]
        # 应该包含A域的mcp
        for mcp_id in TestMCPDomain.A_mcp_list:
            assert mcp_id in mcp_ids
        # 不应该包含公共域或B域的mcp
        for mcp_id in TestMCPDomain.pub_mcp_list:
            assert mcp_id not in mcp_ids
        for mcp_id in TestMCPDomain.B_mcp_list:
            assert mcp_id not in mcp_ids

    @pytest.mark.skip
    @allure.title("获取mcp列表，在其他业务域，无法获取到mcp")
    def test_mcp_domain_05(self):
        # 用户a1尝试在业务域B获取mcp列表，应该失败或返回空列表
        headers = TestMCPDomain.a_headers.copy()
        headers["x-business-domain"] = TestMCPDomain.domain_list[1]  # 业务域B
        params = {"all": "true"}
        result = self.client.GetMCPList(params, headers)
        assert result[0] == 403

    @allure.title("获取mcp服务市场列表，在公共业务域，仅能获取到公共业务域的mcp")
    def test_mcp_domain_06(self):
        # 在公共业务域获取mcp市场列表
        params = {"all": "true"}
        result = self.client.GetMCPMarketList(params, TestMCPDomain.pub_headers)
        assert result[0] == 200
        mcp_ids = [mcp["mcp_id"] for mcp in result[1]["data"]]
        # 应该包含公共域已发布的mcp
        for mcp_id in TestMCPDomain.pub_mcp_list:
            assert mcp_id in mcp_ids
        # 不应该包含A域或B域的mcp
        for mcp_id in TestMCPDomain.A_mcp_list:
            assert mcp_id not in mcp_ids
        for mcp_id in TestMCPDomain.B_mcp_list:
            assert mcp_id not in mcp_ids

    @allure.title("获取mcp服务市场列表，在自身所属业务域，仅能获取到本业务域的mcp")
    def test_mcp_domain_07(self):
        # 在业务域A获取mcp市场列表
        params = {"all": "true"}
        result = self.client.GetMCPMarketList(params, TestMCPDomain.a_headers)
        assert result[0] == 200
        mcp_ids = [mcp["mcp_id"] for mcp in result[1]["data"]]
        # 应该包含A域的已发布mcp
        for mcp_id in TestMCPDomain.A_mcp_list:
            assert mcp_id in mcp_ids
        # 不应该包含公共域或B域的mcp
        for mcp_id in TestMCPDomain.pub_mcp_list:
            assert mcp_id not in mcp_ids
        for mcp_id in TestMCPDomain.B_mcp_list:
            assert mcp_id not in mcp_ids

    @pytest.mark.skip
    @allure.title("获取mcp服务市场列表，在其他业务域，无法获取到mcp")
    def test_mcp_domain_08(self):
        # 用户a1尝试在业务域B获取mcp市场列表，应该失败或返回空列表
        headers = TestMCPDomain.a_headers.copy()
        headers["x-business-domain"] = TestMCPDomain.domain_list[1]  # 业务域B
        params = {"all": "true"}
        result = self.client.GetMCPMarketList(params, headers)
        assert result[0] == 403

    @allure.title("获取mcp服务市场列表，可一次性获取到自身所属业务域和公共业务域的mcp资源")
    def test_mcp_domain_09(self):
        headers = TestMCPDomain.a_headers.copy()
        a_domain = TestMCPDomain.domain_list[0]
        domain_list = f"bd_public,{a_domain}"
        headers["x-business-domain"] = domain_list
        # 获取自身所属业务域和公共业务域的mcp市场列表
        params = {"all": "true"}
        result = self.client.GetMCPMarketList(params, headers)
        assert result[0] == 200
        mcp_ids = [mcp["mcp_id"] for mcp in result[1]["data"]]
        # 应该包含A域和公共域的已发布mcp
        for mcp_id in TestMCPDomain.A_mcp_list:
            assert mcp_id in mcp_ids   
        for mcp_id in TestMCPDomain.pub_mcp_list:
            assert mcp_id in mcp_ids
        # 应该包含B域的mcp
        for mcp_id in TestMCPDomain.B_mcp_list:
            assert mcp_id not in mcp_ids

    @allure.title("有删除权限，删除公共业务域的mcp，删除成功")
    def test_mcp_domain_10(self):
        # 用户a1删除公共域的mcp，应该成功
        headers = TestMCPDomain.a_headers.copy()
        headers["x-business-domain"] = "bd_public"
        mcp_id = TestMCPDomain.pub_mcp_list[0]
        # 先下架
        result = self.client.MCPReleaseAction(mcp_id, {"status": "offline"}, headers)
        assert result[0] == 200
        # 再删除
        result = self.client.DeleteMCP(mcp_id, headers)
        assert result[0] == 200

    @allure.title("有删除权限，删除自身所属业务域的mcp，删除成功")
    def test_mcp_domain_11(self):
        # 用户a1删除A域的mcp，应该成功
        mcp_id = TestMCPDomain.A_mcp_list[0]
        # 先下架
        result = self.client.MCPReleaseAction(mcp_id, {"status": "offline"}, TestMCPDomain.a_headers)
        assert result[0] == 200
        # 再删除
        result = self.client.DeleteMCP(mcp_id, TestMCPDomain.a_headers)
        assert result[0] == 200

    @pytest.mark.skip
    @allure.title("有删除权限，删除其他业务域的mcp，删除失败")
    def test_mcp_domain_12(self):
        # 用户a1尝试删除B域的mcp，应该失败（403）
        headers = TestMCPDomain.a_headers.copy()
        headers["x-business-domain"] = TestMCPDomain.domain_list[1]  # 业务域B
        mcp_id = TestMCPDomain.B_mcp_list[0]
        # 先下架
        result = self.client.MCPReleaseAction(mcp_id, {"status": "offline"}, TestMCPDomain.b_headers)
        assert result[0] == 200
        # 再删除
        result = self.client.DeleteMCP(mcp_id, headers)
        assert result[0] == 403

    @allure.title("有新建权限，在公共业务域和自身所属业务域注册内置mcp，注册成功")
    def test_mcp_domain_13(self):
        # 在公共业务域和自身所属业务域注册内置mcp
        for domain in ["bd_public", TestMCPDomain.domain_list[0]]:
            headers = {
                "x-account-id": TestMCPDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": domain
            }
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            mcp_id = str(uuid.uuid4())
            payload = {
                "mcp_id": mcp_id,
                "name": name,
                "description": "test builtin mcp server",
                "mode": "sse",
                "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
                "command": "ls",
                "args": ["-l"],
                "headers": {"Content-Type": "application/json"},
                "env": {},
                "source": "intenal",
                "protected_flag": False,
                "config_version": "1.0.0",
                "config_source": "auto"
            }
            result = self.client1.Register(payload, headers)
            assert result[0] == 200

    @pytest.mark.skip
    @allure.title("有新建权限，在其他业务域注册内置mcp，注册失败")
    def test_mcp_domain_14(self):
        # 用户a1尝试在业务域B注册内置mcp，应该失败
        headers = {
                "x-account-id": TestMCPDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": TestMCPDomain.domain_list[1]  # 业务域B
        }
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        mcp_id = str(uuid.uuid4())
        payload = {
            "mcp_id": mcp_id,
            "name": name,
            "description": "test builtin mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "command": "ls",
            "args": ["-l"],
            "headers": {"Content-Type": "application/json"},
            "env": {},
            "source": "intenal",
            "protected_flag": False,
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client1.Register(payload, headers)
        assert result[0] == 403

    @allure.title("有删除权限，注销公共业务域的内置mcp，注销成功")
    def test_mcp_domain_15(self):
        # 用户a1注销公共域的内置mcp，应该成功
        # 从mcp_list中获取公共域的内置mcp（索引1）
        pub_builtin_mcp_id = TestMCPDomain.mcp_list[1]  # 公共域内置mcp
        headers = {
                "x-account-id": TestMCPDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": "bd_public"
        }
        result = self.client1.UnRegister(pub_builtin_mcp_id, {}, headers)
        assert result[0] == 200

    @allure.title("有删除权限，注销自身所属业务域的内置mcp，注销成功")
    def test_mcp_domain_16(self):
        # 用户a1注销A域的内置mcp，应该成功
        # 从mcp_list中获取A域的内置mcp（索引3）
        a_builtin_mcp_id = TestMCPDomain.mcp_list[3]  # A域内置mcp
        headers = {
                "x-account-id": TestMCPDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": TestMCPDomain.domain_list[0]
        }
        result = self.client1.UnRegister(a_builtin_mcp_id, {}, headers)
        assert result[0] == 200

    @pytest.mark.skip
    @allure.title("有删除权限，注销其他业务域的内置mcp，注销失败")
    def test_mcp_domain_17(self):
        # 用户a1尝试注销B域的内置mcp，应该失败
        # 从mcp_list中获取B域的内置mcp（索引5）
        b_builtin_mcp_id = TestMCPDomain.mcp_list[5]  # B域内置mcp
        headers = {
                "x-account-id": TestMCPDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": TestMCPDomain.domain_list[1]
        }
        result = self.client1.UnRegister(b_builtin_mcp_id, {}, headers)
        assert result[0] == 403