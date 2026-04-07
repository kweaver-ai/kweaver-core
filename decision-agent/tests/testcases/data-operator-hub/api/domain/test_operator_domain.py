# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import uuid

from common.get_content import GetContent
from lib.operator import Operator

@allure.feature("算子平台业务域测试：算子业务域测试")
class TestOperatorDomain:
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, TestDomainData, DomainPrepare):
        domain_list, headers, operator_list, _, _ = TestDomainData
        user_list = DomainPrepare[0]
        TestOperatorDomain.domain_list = domain_list
        TestOperatorDomain.pub_headers = headers[0]
        TestOperatorDomain.a_headers = headers[1]
        TestOperatorDomain.b_headers = headers[2]
        TestOperatorDomain.operator_list = operator_list  # 保存完整的operator_list
        TestOperatorDomain.pub_op_list = operator_list[0:1]  # 公共域自定义算子
        TestOperatorDomain.A_op_list = operator_list[2:3]  # A域自定义算子
        TestOperatorDomain.B_op_list = operator_list[4:5]  # B域自定义算子
        TestOperatorDomain.user_list = user_list

    @allure.title("有新建权限，在公共业务域和自身所属业务域新建算子，新建成功")
    def test_operator_domain_01(self):
        headers1 = TestOperatorDomain.a_headers.copy()
        headers1["x-business-domain"] = "bd_public"
        for headers in [headers1, TestOperatorDomain.a_headers]:
            filepath = "./resource/openapi/compliant/relations.yaml"
            yaml_data = GetContent(filepath).yamlfile()
            data = {
                "data": str(yaml_data),
                "operator_metadata_type": "openapi"
            }
            result = self.client.RegisterOperator(data, headers)
            assert result[0] == 200

    @pytest.mark.skip
    @allure.title("有新建权限，在其他业务域新建算子，新建失败")
    def test_operator_domain_02(self):
        # 用户a1尝试在业务域B创建算子，应该失败
        headers = TestOperatorDomain.a_headers.copy()
        headers["x-business-domain"] = TestOperatorDomain.domain_list[1]  # 业务域B
        filepath = "./resource/openapi/compliant/relations.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, headers)
        assert result[0] == 200
        for op in result[1]:
            assert op["status"] == "failed"

    @allure.title("获取算子列表，在公共业务域，仅能获取到公共业务域的算子")
    def test_operator_domain_03(self):
        # 在公共业务域获取算子列表
        params = {"all": "true"}
        result = self.client.GetOperatorList(params, TestOperatorDomain.pub_headers)
        assert result[0] == 200
        operator_ids = [op["operator_id"] for op in result[1]["data"]]
        # 应该包含公共域的算子
        for op_id in TestOperatorDomain.pub_op_list:
            assert op_id in operator_ids
        # 不应该包含A域或B域的算子
        for op_id in TestOperatorDomain.A_op_list:
            assert op_id not in operator_ids
        for op_id in TestOperatorDomain.B_op_list:
            assert op_id not in operator_ids

    @allure.title("获取算子列表，在自身所属业务域，仅能获取到本业务域的算子")
    def test_operator_domain_04(self):
        # 在业务域A获取算子列表
        params = {"all": "true"}
        result = self.client.GetOperatorList(params, TestOperatorDomain.a_headers)
        assert result[0] == 200
        operator_ids = [op["operator_id"] for op in result[1]["data"]]
        # 应该包含A域的算子
        for op_id in TestOperatorDomain.A_op_list:
            assert op_id in operator_ids
        # 不应该包含公共域或B域的算子
        for op_id in TestOperatorDomain.pub_op_list:
            assert op_id not in operator_ids
        for op_id in TestOperatorDomain.B_op_list:
            assert op_id not in operator_ids

    @pytest.mark.skip
    @allure.title("获取算子列表，在其他业务域，无法获取到算子")
    def test_operator_domain_05(self):
        # 用户a1尝试在业务域B获取算子列表，应该失败或返回空列表
        headers = TestOperatorDomain.a_headers.copy()
        headers["x-business-domain"] = TestOperatorDomain.domain_list[1]  # 业务域B
        params = {"all": "true"}
        result = self.client.GetOperatorList(params, headers)
        assert result[0] == 403

    @allure.title("获取算子市场列表，在公共业务域，仅能获取到公共业务域的算子")
    def test_operator_domain_06(self):
        # 在公共业务域获取算子市场列表
        params = {"all": "true"}
        result = self.client.GetOperatorMarketList(params, TestOperatorDomain.pub_headers)
        assert result[0] == 200
        operator_ids = [op["operator_id"] for op in result[1]["data"]]
        # 应该包含公共域已发布的算子
        for op_id in TestOperatorDomain.pub_op_list:
            assert op_id in operator_ids
        # 不应该包含A域或B域的算子
        for op_id in TestOperatorDomain.A_op_list:
            assert op_id not in operator_ids
        for op_id in TestOperatorDomain.B_op_list:
            assert op_id not in operator_ids

    @allure.title("获取算子市场列表，在自身所属业务域，仅能获取到本业务域的算子")
    def test_operator_domain_07(self):
        # 在业务域A获取算子市场列表
        params = {"all": "true"}
        result = self.client.GetOperatorMarketList(params, TestOperatorDomain.a_headers)
        assert result[0] == 200
        operator_ids = [op["operator_id"] for op in result[1]["data"]]
        # 应该包含A域的已发布算子
        for op_id in TestOperatorDomain.A_op_list:
            assert op_id in operator_ids
        # 不应该包含公共域或B域的算子
        for op_id in TestOperatorDomain.pub_op_list:
            assert op_id not in operator_ids
        for op_id in TestOperatorDomain.B_op_list:
            assert op_id not in operator_ids

    @pytest.mark.skip
    @allure.title("获取算子市场列表，在其他业务域，无法获取到算子")
    def test_operator_domain_08(self):
        # 用户a1尝试在业务域B获取算子市场列表，应该失败或返回空列表
        headers = TestOperatorDomain.a_headers.copy()
        headers["x-business-domain"] = TestOperatorDomain.domain_list[1]  # 业务域B
        params = {"all": "true"}
        result = self.client.GetOperatorMarketList(params, headers)
        assert result[0] == 403

    @allure.title("获取算子服务市场列表，可一次性获取到自身所属业务域和公共业务域的算子")
    def test_mcp_domain_09(self):
        headers = TestOperatorDomain.a_headers.copy()
        a_domain = TestOperatorDomain.domain_list[0]
        domain_list = f"bd_public,{a_domain}"
        headers["x-business-domain"] = domain_list
        # 获取自身所属业务域和公共业务域的算子市场列表
        params = {"all": "true"}
        result = self.client.GetOperatorMarketList(params, headers)
        assert result[0] == 200
        operator_ids = [op["operator_id"] for op in result[1]["data"]]
        # 应该包含A域和公共域的已发布算子
        for operator_id in TestOperatorDomain.A_op_list:
            assert operator_id in operator_ids   
        for operator_id in TestOperatorDomain.pub_op_list:
            assert operator_id in operator_ids
        # 应该包含B域的算子
        for operator_id in TestOperatorDomain.B_op_list:
            assert operator_id not in operator_ids

    @allure.title("有删除权限，删除公共业务域的算子，删除成功")
    def test_operator_domain_10(self):
        # 用户a1删除公共域的算子，应该成功
        headers = TestOperatorDomain.a_headers.copy()
        headers["x-business-domain"] = "bd_public"
        operator_id = TestOperatorDomain.pub_op_list[0]
        # 获取算子版本信息
        result = self.client.GetOperatorInfo(operator_id, headers)
        assert result[0] == 200
        version = result[1]["version"]
        # 先下架
        update_data = [{"operator_id": operator_id, "status": "offline"}]
        result = self.client.UpdateOperatorStatus(update_data, headers)
        assert result[0] == 200
        # 再删除
        del_data = [{"operator_id": operator_id, "version": version}]
        result = self.client.DeleteOperator(del_data, headers)
        assert result[0] == 200

    @allure.title("有删除权限，删除自身所属业务域的算子，删除成功")
    def test_operator_domain_11(self):
        # 用户a1删除A域的算子，应该成功
        operator_id = TestOperatorDomain.A_op_list[0]
        # 获取算子版本信息
        result = self.client.GetOperatorInfo(operator_id, TestOperatorDomain.a_headers)
        assert result[0] == 200
        version = result[1]["version"]
        # 先下架
        update_data = [{"operator_id": operator_id, "status": "offline"}]
        result = self.client.UpdateOperatorStatus(update_data, TestOperatorDomain.a_headers)
        assert result[0] == 200
        # 再删除
        del_data = [{"operator_id": operator_id, "version": version}]
        result = self.client.DeleteOperator(del_data, TestOperatorDomain.a_headers)
        assert result[0] == 200

    @pytest.mark.skip
    @allure.title("有删除权限，删除其他业务域的算子，删除失败")
    def test_operator_domain_12(self):
        # 用户a1尝试删除B域的算子，应该失败（403）
        headers = TestOperatorDomain.a_headers.copy()
        headers["x-business-domain"] = TestOperatorDomain.domain_list[1]  # 业务域B
        operator_id = TestOperatorDomain.B_op_list[0]
        # 获取算子版本信息
        result = self.client.GetOperatorInfo(operator_id, TestOperatorDomain.b_headers)
        assert result[0] == 200
        version = result[1]["version"]
        # 先下架
        update_data = [{"operator_id": operator_id, "status": "offline"}]
        result = self.client.UpdateOperatorStatus(update_data, TestOperatorDomain.b_headers)
        assert result[0] == 200
        # 再删除
        del_data = [{"operator_id": operator_id, "version": version}]
        result = self.client.DeleteOperator(del_data, headers)
        assert result[0] == 403

    @allure.title("有新建权限，在公共业务域和自身所属业务域注册内置算子，注册成功")
    def test_operator_domain_13(self):
        # 在公共业务域和自身所属业务域注册内置算子
        for domain in ["bd_public", TestOperatorDomain.domain_list[0]]:
            headers = {
                "x-account-id": TestOperatorDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": domain
            }
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            operator_id = str(uuid.uuid4())
            filepath = "./resource/openapi/compliant/test3.yaml"
            api_data = GetContent(filepath).yamlfile()
            payload = {
                "operator_id": operator_id,
                "name": name,
                "data": api_data,
                "metadata_type": "openapi",
                "operator_type": "basic",
                "execution_mode": "sync",
                "source": "intenal",
                "config_source": "auto",
                "config_version": "1.0.0"
            }
            result = self.client.InternalBuiltinOperator(payload, headers)
            assert result[0] == 200

    @pytest.mark.skip
    @allure.title("有新建权限，在其他业务域注册内置算子，注册失败")
    def test_operator_domain_14(self):
        # 用户a1尝试在业务域B注册内置算子，应该失败
        headers = {
                "x-account-id": TestOperatorDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": TestOperatorDomain.domain_list[1]  # 业务域B
        }
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        operator_id = str(uuid.uuid4())
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        payload = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        result = self.client.InternalBuiltinOperator(payload, headers)
        assert result[0] == 403

    @allure.title("有更新权限，更新公共业务域的内置算子，更新成功")
    def test_operator_domain_15(self):
        # 用户a1更新公共域的内置算子，应该成功
        # 从operator_list中获取公共域的内置算子（索引1）
        pub_builtin_op_id = TestOperatorDomain.operator_list[1]  # 公共域内置算子
        headers = {
                "x-account-id": TestOperatorDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": "bd_public"
        }
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        payload = {
            "operator_id": pub_builtin_op_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        result = self.client.InternalBuiltinOperator(payload, headers)
        assert result[0] == 200

    @allure.title("有更新权限，更新自身所属业务域的内置算子，更新成功")
    def test_operator_domain_16(self):
        # 用户a1更新A域的内置算子，应该成功
        # 从operator_list中获取A域的内置算子（索引3）
        a_builtin_op_id = TestOperatorDomain.operator_list[3]  # A域内置算子
        headers = {
                "x-account-id": TestOperatorDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": TestOperatorDomain.domain_list[0]
        }
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        payload = {
            "operator_id": a_builtin_op_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        result = self.client.InternalBuiltinOperator(payload, headers)
        assert result[0] == 200

    # 仅新建、查看、删除校验业务域，更新时不做校验
    @pytest.mark.skip
    @allure.title("有更新权限，更新其他业务域的内置算子，更新失败")
    def test_operator_domain_17(self):
        # 用户a1尝试更新B域的内置算子，应该失败
        # 从operator_list中获取B域的内置算子（索引5）
        b_builtin_op_id = TestOperatorDomain.operator_list[5]  # B域内置算子
        headers = {
                "x-account-id": TestOperatorDomain.user_list[0],
                "x-account-type": "user",
                "x-business-domain": TestOperatorDomain.domain_list[1]
        }
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        payload = {
            "operator_id": b_builtin_op_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        result = self.client.InternalBuiltinOperator(payload, headers)
        assert result[0] == 403

