# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import uuid

from common.get_content import GetContent
from lib.tool_box import ToolBox

@allure.feature("算子平台业务域测试：工具箱业务域测试")
class TestToolboxDomain:
    client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, TestDomainData, DomainPrepare):
        domain_list, headers, _, box_list, _ = TestDomainData
        user_list = DomainPrepare[0]
        TestToolboxDomain.domain_list = domain_list
        TestToolboxDomain.pub_headers = headers[0]
        TestToolboxDomain.a_headers = headers[1]
        TestToolboxDomain.b_headers = headers[2]
        TestToolboxDomain.box_list = box_list  # 保存完整的box_list
        TestToolboxDomain.pub_box_list = box_list[0:1]  # 公共域自定义工具箱
        TestToolboxDomain.A_box_list = box_list[2:3]  # A域自定义工具箱
        TestToolboxDomain.B_box_list = box_list[4:5]  # B域自定义工具箱
        TestToolboxDomain.user_list = user_list

    @allure.title("有新建权限，在公共业务域和自身所属业务域新建工具箱，新建成功")
    def test_toolbox_domain_01(self):
        headers1 = TestToolboxDomain.a_headers.copy()
        headers1["x-business-domain"] = "bd_public"
        for headers in [headers1, TestToolboxDomain.a_headers]:
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            filepath = "./resource/openapi/compliant/mcp.yaml"
            yaml_data = GetContent(filepath).yamlfile()
            data = {
                "box_name": name,
                "data": yaml_data,
                "metadata_type": "openapi"
            }
            result = self.client.CreateToolbox(data, headers)
            assert result[0] == 200

    @pytest.mark.skip
    @allure.title("有新建权限，在其他业务域新建工具箱，新建失败")
    def test_toolbox_domain_02(self):
        # 用户a1尝试在业务域B创建工具箱，应该失败
        headers = TestToolboxDomain.a_headers.copy()
        headers["x-business-domain"] = TestToolboxDomain.domain_list[1]  # 业务域B
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, headers)
        assert result[0] == 403

    @allure.title("获取工具箱列表，在公共业务域，仅能获取到公共业务域的工具箱")
    def test_toolbox_domain_03(self):
        # 在公共业务域获取工具箱列表
        params = {"all": "true"}
        result = self.client.GetToolboxList(params, TestToolboxDomain.pub_headers)
        assert result[0] == 200
        box_ids = [box["box_id"] for box in result[1]["data"]]
        # 应该包含公共域的工具箱
        for box_id in TestToolboxDomain.pub_box_list:
            assert box_id in box_ids
        # 不应该包含A域或B域的工具箱
        for box_id in TestToolboxDomain.A_box_list:
            assert box_id not in box_ids
        for box_id in TestToolboxDomain.B_box_list:
            assert box_id not in box_ids

    @allure.title("获取工具箱列表，在自身所属业务域，仅能获取到本业务域的工具箱")
    def test_toolbox_domain_04(self):
        # 在业务域A获取工具箱列表
        params = {"all": "true"}
        result = self.client.GetToolboxList(params, TestToolboxDomain.a_headers)
        assert result[0] == 200
        box_ids = [box["box_id"] for box in result[1]["data"]]
        # 应该包含A域的工具箱
        for box_id in TestToolboxDomain.A_box_list:
            assert box_id in box_ids
        # 不应该包含公共域或B域的工具箱
        for box_id in TestToolboxDomain.pub_box_list:
            assert box_id not in box_ids
        for box_id in TestToolboxDomain.B_box_list:
            assert box_id not in box_ids

    @pytest.mark.skip
    @allure.title("获取工具箱列表，在其他业务域，无法获取到工具箱")
    def test_toolbox_domain_05(self):
        # 用户a1尝试在业务域B获取工具箱列表，应该失败或返回空列表
        headers = TestToolboxDomain.a_headers.copy()
        headers["x-business-domain"] = TestToolboxDomain.domain_list[1]  # 业务域B
        params = {"all": "true"}
        result = self.client.GetToolboxList(params, headers)
        assert result[0] == 403

    @allure.title("获取工具箱市场列表，在公共业务域，仅能获取到公共业务域的工具箱")
    def test_toolbox_domain_06(self):
        # 在公共业务域获取工具箱市场列表
        params = {"all": "true"}
        result = self.client.GetMarketToolboxList(params, TestToolboxDomain.pub_headers)
        assert result[0] == 200
        box_ids = [box["box_id"] for box in result[1]["data"]]
        # 应该包含公共域已发布的工具箱
        for box_id in TestToolboxDomain.pub_box_list:
            assert box_id in box_ids
        # 不应该包含A域或B域的工具箱
        for box_id in TestToolboxDomain.A_box_list:
            assert box_id not in box_ids
        for box_id in TestToolboxDomain.B_box_list:
            assert box_id not in box_ids

    @allure.title("获取工具箱市场列表，在自身所属业务域，仅能获取到本业务域的工具箱")
    def test_toolbox_domain_07(self):
        # 在业务域A获取工具箱市场列表
        params = {"all": "true"}
        result = self.client.GetMarketToolboxList(params, TestToolboxDomain.a_headers)
        assert result[0] == 200
        box_ids = [box["box_id"] for box in result[1]["data"]]
        # 应该包含A域的已发布工具箱
        for box_id in TestToolboxDomain.A_box_list:
            assert box_id in box_ids
        # 不应该包含公共域或B域的工具箱
        for box_id in TestToolboxDomain.pub_box_list:
            assert box_id not in box_ids
        for box_id in TestToolboxDomain.B_box_list:
            assert box_id not in box_ids

    @pytest.mark.skip
    @allure.title("获取工具箱市场列表，在其他业务域，无法获取到工具箱")
    def test_toolbox_domain_08(self):
        # 用户a1尝试在业务域B获取工具箱市场列表，应该失败或返回空列表
        headers = TestToolboxDomain.a_headers.copy()
        headers["x-business-domain"] = TestToolboxDomain.domain_list[1]  # 业务域B
        params = {"all": "true"}
        result = self.client.GetMarketToolboxList(params, headers)
        assert result[0] == 403

    @allure.title("获取工具箱服务市场列表，可一次性获取到自身所属业务域和公共业务域的工具箱")
    def test_mcp_domain_09(self):
        headers = TestToolboxDomain.a_headers.copy()
        a_domain = TestToolboxDomain.domain_list[0]
        domain_list = f"bd_public,{a_domain}"
        headers["x-business-domain"] = domain_list
        # 获取自身所属业务域和公共业务域的工具箱市场列表
        params = {"all": "true"}
        result = self.client.GetMarketToolboxList(params, headers)
        assert result[0] == 200
        box_ids = [box["box_id"] for box in result[1]["data"]]
        # 应该包含A域和公共域的已发布工具箱
        for box_id in TestToolboxDomain.A_box_list:
            assert box_id in box_ids   
        for box_id in TestToolboxDomain.pub_box_list:
            assert box_id in box_ids
        # 应该包含B域的工具箱
        for box_id in TestToolboxDomain.B_box_list:
            assert box_id not in box_ids

    @allure.title("有删除权限，删除公共业务域的工具箱，删除成功")
    def test_toolbox_domain_10(self):
        # 用户a1删除公共域的工具箱，应该成功
        headers = TestToolboxDomain.a_headers.copy()
        headers["x-business-domain"] = "bd_public"
        box_id = TestToolboxDomain.pub_box_list[0]
        # 先下架
        result = self.client.UpdateToolboxStatus(box_id, {"status": "offline"}, headers)
        assert result[0] == 200
        # 再删除
        result = self.client.DeleteToolbox(box_id, headers)
        assert result[0] == 200

    @allure.title("有删除权限，删除自身所属业务域的工具箱，删除成功")
    def test_toolbox_domain_11(self):
        # 用户a1删除A域的工具箱，应该成功
        box_id = TestToolboxDomain.A_box_list[0]
        # 先下架
        result = self.client.UpdateToolboxStatus(box_id, {"status": "offline"}, TestToolboxDomain.a_headers)
        assert result[0] == 200
        # 再删除
        result = self.client.DeleteToolbox(box_id, TestToolboxDomain.a_headers)
        assert result[0] == 200

    @pytest.mark.skip
    @allure.title("有删除权限，删除其他业务域的工具箱，删除失败")
    def test_toolbox_domain_12(self):
        # 用户a1尝试删除B域的工具箱，应该失败（403）
        headers = TestToolboxDomain.a_headers.copy()
        headers["x-business-domain"] = TestToolboxDomain.domain_list[1]  # 业务域B
        box_id = TestToolboxDomain.B_box_list[0]
        # 先下架
        result = self.client.UpdateToolboxStatus(box_id, {"status": "offline"}, TestToolboxDomain.b_headers)
        assert result[0] == 200
        # 再删除
        result = self.client.DeleteToolbox(box_id, headers)
        assert result[0] == 403

    @allure.title("有新建权限，在公共业务域和自身所属业务域注册内置工具箱，注册成功")
    def test_toolbox_domain_13(self):
        # 在公共业务域和自身所属业务域注册内置工具箱
        for domain in ["bd_public", TestToolboxDomain.domain_list[0]]:
            headers = {
                "x-account-id": TestToolboxDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": domain
            }
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            box_id = str(uuid.uuid4())
            filepath = "./resource/openapi/compliant/mcp.yaml"
            yaml_data = GetContent(filepath).yamlfile()
            payload = {
                "box_id": box_id,
                "box_name": name,
                "box_desc": "test description",
                "data": yaml_data,
                "metadata_type": "openapi",
                "source": "internal",
                "config_version": "1.0.0",
                "config_source": "auto"
            }
            result = self.client.InternalBuiltin(payload, headers)
            assert result[0] == 200

    @pytest.mark.skip
    @allure.title("有新建权限，在其他业务域注册内置工具箱，注册失败")
    def test_toolbox_domain_14(self):
        # 用户a1尝试在业务域B注册内置工具箱，应该失败
        headers = {
                "x-account-id": TestToolboxDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": TestToolboxDomain.domain_list[1]  # 业务域B
        }
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        box_id = str(uuid.uuid4())
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        payload = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.InternalBuiltin(payload, headers)
        assert result[0] == 403

    @allure.title("有更新权限，更新公共业务域的内置工具箱，更新成功")
    def test_toolbox_domain_15(self):
        # 用户a1更新公共域的内置工具箱，应该成功
        # 从box_list中获取公共域的内置工具箱（索引1）
        pub_builtin_box_id = TestToolboxDomain.box_list[1]  # 公共域内置工具箱
        headers = {
                "x-account-id": TestToolboxDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": "bd_public"
        }
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        payload = {
            "box_id": pub_builtin_box_id,
            "box_name": name,
            "box_desc": "updated description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.InternalBuiltin(payload, headers)
        assert result[0] == 200

    @allure.title("有更新权限，更新自身所属业务域的内置工具箱，更新成功")
    def test_toolbox_domain_16(self):
        # 用户a1更新A域的内置工具箱，应该成功
        # 从box_list中获取A域的内置工具箱（索引3）
        a_builtin_box_id = TestToolboxDomain.box_list[3]  # A域内置工具箱
        headers = {
                "x-account-id": TestToolboxDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": TestToolboxDomain.domain_list[0]
        }
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        payload = {
            "box_id": a_builtin_box_id,
            "box_name": name,
            "box_desc": "updated description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.InternalBuiltin(payload, headers)
        assert result[0] == 200

    # 仅新建、查看、删除校验业务域，更新时不做校验
    @pytest.mark.skip
    @allure.title("有更新权限，更新其他业务域的内置工具箱，更新失败")
    def test_toolbox_domain_17(self):
        # 用户a1尝试更新B域的内置工具箱，应该失败
        # 从box_list中获取B域的内置工具箱（索引5）
        b_builtin_box_id = TestToolboxDomain.box_list[5]  # B域内置工具箱
        headers = {
                "x-account-id": TestToolboxDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": TestToolboxDomain.domain_list[1]
        }
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        payload = {
            "box_id": b_builtin_box_id,
            "box_name": name,
            "box_desc": "updated description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.InternalBuiltin(payload, headers)
        assert result[0] == 403

    @allure.title("(内部接口)注册内置工具箱，不传业务域id，注册成功，默认在公共域")
    def test_toolbox_domain_18(self):
        # 注册内置工具箱，不传业务域id
        headers = {
                "x-account-id": TestToolboxDomain.user_list[0],
                "x-account-type": "user"
        }
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        box_id = str(uuid.uuid4())
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        payload = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.InternalBuiltin(payload, headers)
        assert result[0] == 200
        # 公共域下可获取到该工具箱
        params = {"all": "true"}
        result = self.client.GetToolboxList(params, TestToolboxDomain.pub_headers)
        assert result[0] == 200
        box_ids = [box["box_id"] for box in result[1]["data"]]
        assert box_id in box_ids

