# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import uuid

from common.get_content import GetContent
from common.get_token import GetToken
from lib.operator import Operator
from lib.permission import Perm
from lib.operator_internal import InternalOperator

@allure.feature("算子平台权限测试：算子权限测试")
class TestOperatorPerm:
    client = Operator()
    perm_client = Perm()
    a_headers = {}
    b_headers = {}
    c_headers = {}
    d_headers = {}
    operators = []
    user_list = []

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers, PermPrepare):
        # 获取用户token
        configfile = "./config/env.ini"
        file = GetContent(configfile)
        config = file.config()
        host = config["server"]["host"]
        a_token = GetToken(host=host).get_token(host, "a", "111111")
        TestOperatorPerm.a_headers = {
            "Authorization": f"Bearer {a_token[1]}"
        }
        
        b_token = GetToken(host=host).get_token(host, "b", "111111")
        TestOperatorPerm.b_headers = {
            "Authorization": f"Bearer {b_token[1]}"
        }

        c_token = GetToken(host=host).get_token(host, "c", "111111")
        TestOperatorPerm.c_headers = {
            "Authorization": f"Bearer {c_token[1]}"
        }

        d_token = GetToken(host=host).get_token(host, "d", "111111")
        TestOperatorPerm.d_headers = {
            "Authorization": f"Bearer {d_token[1]}"
        }

        # AI管理员新建算子
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        for op in result[1]:
            if op["status"] == "success":
                operator = {
                    "operator_id": op["operator_id"],
                    "version": op["version"]
                }
                TestOperatorPerm.operators.append(operator)

        '''
        为用户a设置算子资源的新建权限
        为用户b设置算子的查看、编辑、删除权限
        为用户c设置算子的发布、下架、公共访问和使用权限
        为用户d设置算子的新建、编辑、发布权限
        '''
        TestOperatorPerm.user_list = PermPrepare[1]
        # TestOperatorPerm.user_list = ['caad2ad2-56ec-11f0-bce8-8269137aaf40', 'cace09a0-56ec-11f0-8f88-8269137aaf40', 'caec81f0-56ec-11f0-9591-8269137aaf40', 'cb2e864a-56ec-11f0-b88f-8269137aaf40']
        perm_data = [
            {
                "accessor": {"id": TestOperatorPerm.user_list[0], "name": "a", "type": "user"},
                "resource": {"id": "*", "type": "operator", "name": ""},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            },
            {
                "accessor": {"id": TestOperatorPerm.user_list[3], "name": "d", "type": "user"},
                "resource": {"id": "*", "type": "operator", "name": ""},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            }
        ]
        for operator in TestOperatorPerm.operators:
            b_data = {
                "accessor": {"id": TestOperatorPerm.user_list[1], "name": "b", "type": "user"},
                "resource": {"id": operator["operator_id"], "type": "operator", "name": "查看编辑删除权限"},
                "operation": {"allow": [{"id": "view"}, {"id": "modify"}, {"id": "delete"}], "deny": []}
            }
            c_data = {
                "accessor": {"id": TestOperatorPerm.user_list[2], "name": "c", "type": "user"},
                "resource": {"id": operator["operator_id"], "type": "operator", "name": "发布下架公开访问使用权限"},
                "operation": { "allow": [{"id": "publish"}, {"id": "unpublish"}, {"id": "public_access"}, {"id": "execute"}],  "deny": []}
            }
            d_data = {
                "accessor": {"id": TestOperatorPerm.user_list[3], "name": "d", "type": "user"},
                "resource": {"id": operator["operator_id"], "type": "operator", "name": "发布编辑权限"},
                "operation": {"allow": [{"id": "modify"}, {"id": "publish"}], "deny": []}
            }
            perm_data.append(b_data)
            perm_data.append(c_data)
            perm_data.append(d_data)
        result = self.perm_client.SetPerm(perm_data, Headers)
        assert "20" in str(result[0])
    
    @allure.title("有新建权限，新建算子不发布，新建成功，创建者和AI管理员可对该算子进行所有操作")
    def test_operator_perm_01(self, Headers):
        # 新建
        filepath = "./resource/openapi/compliant/relations.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        create_result = self.client.RegisterOperator(data, TestOperatorPerm.a_headers)
        assert create_result[0] == 200
        operator_id = create_result[1][0]["operator_id"]
        headers = [Headers, TestOperatorPerm.a_headers]
        for header in headers:
            # 更新算子信息
            filepath = "./resource/openapi/compliant/update-test1.yaml"
            api_data = GetContent(filepath).yamlfile()
            data = {
                "operator_id": operator_id,
                "data": str(api_data),
                "operator_metadata_type": "openapi"
            }
            result = self.client.UpdateOperatorInfo(data, header)
            assert result[0] == 200
            # 获取算子列表
            data = {"all": "true"}
            result = self.client.GetOperatorList(data, header)
            assert result[0] == 200
            assert operator_id in [op["operator_id"] for op in result[1]["data"]]
            # 编辑算子
            data = { "operator_id": operator_id, "name": "test_edit" }
            re = self.client.EditOperator(data, header)
            assert re[0] == 200
            # 获取算子信息
            result = self.client.GetOperatorInfo(operator_id, header)
            assert result[0] == 200
            assert result[1]["operator_id"] == operator_id
            # 删除算子
            del_data = [{"operator_id": operator_id, "version": create_result[1][0]["version"]}]
            result = self.client.DeleteOperator(del_data, header)
            assert result[0] == 200
            operator_id = create_result[1][1]["operator_id"]
            # 发布算子
            update_data = [{"operator_id": create_result[1][2]["operator_id"], "status": "published"}] 
            result = self.client.UpdateOperatorStatus(update_data, header)
            assert result[0] == 200
            # 下架算子
            update_data = [{"operator_id": create_result[1][2]["operator_id"], "status": "offline"}] 
            result = self.client.UpdateOperatorStatus(update_data, header)
            assert result[0] == 200
            # 调试算子
            debug_data = {
                "operator_id": create_result[1][2]["operator_id"],  "version": create_result[1][2]["version"],  "header": header,
                "path": { "id": operator_id }
            }
            result = self.client.OperatorDebug(debug_data, header)
            assert result[0] == 200
            # 发布算子
            update_data = [{"operator_id": create_result[1][3]["operator_id"], "status": "published"}] 
            result = self.client.UpdateOperatorStatus(update_data, header)
            assert result[0] == 200
            # 发布后编辑生成历史版本
            re = self.client.GetOperatorInfo(create_result[1][3]["operator_id"], header)
            assert re[0] == 200
            data = {
                "operator_id": create_result[1][3]["operator_id"],
                "description": "test edit 1234567"
            }
            re = self.client.EditOperator(data, header)
            assert re[0] == 200
            # 获取算子历史版本详情
            result = self.client.GetOperatorHistoryDetail(create_result[1][3]["operator_id"], create_result[1][3]["version"], header)
            assert result[0] == 200
            # 获取历史版本列表
            result = self.client.GetOperatorHistoryList(create_result[1][3]["operator_id"], header)
            assert result[0] == 200
            assert create_result[1][3]["version"] in [op["version"] for op in result[1]]
            # 获取算子市场列表
            result = self.client.GetOperatorMarketList({"all": True}, header)
            assert result[0] == 200
            assert create_result[1][3]["operator_id"] in [op["operator_id"] for op in result[1]["data"]]
            # 获取算子市场指定算子详情
            result = self.client.GetOperatorMarketDetail(create_result[1][3]["operator_id"], header)
            assert result[0] == 200
    
    @allure.title("有新建权限，新建算子并发布，新建成功，创建者和AI管理员可对该算子进行所有操作")
    def test_operator_perm_02(self, Headers):
        filepath = "./resource/openapi/compliant/template.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        create_result = self.client.RegisterOperator(data, TestOperatorPerm.a_headers)
        assert create_result[0] == 200
        operator_id = create_result[1][0]["operator_id"]
        # 创建者和AI管理员都可进行操作
        for header in [Headers, TestOperatorPerm.a_headers]:
            # 更新算子信息
            filepath = "./resource/openapi/compliant/update-test4.yaml"
            api_data = GetContent(filepath).yamlfile()
            data = {
                "operator_id": operator_id,
                "data": str(api_data),
                "operator_metadata_type": "openapi"
            }
            result = self.client.UpdateOperatorInfo(data, header)
            assert result[0] == 200
            # 获取算子列表
            data = {"all": "true"}
            result = self.client.GetOperatorList(data, header)
            assert result[0] == 200
            assert operator_id in [op["operator_id"] for op in result[1]["data"]]
            # 编辑算子
            data = { "operator_id": operator_id, "name": "test_edit" }
            re = self.client.EditOperator(data, header)
            assert re[0] == 200
            # 获取算子信息
            result = self.client.GetOperatorInfo(operator_id, header)
            assert result[0] == 200
            assert result[1]["operator_id"] == operator_id

    @allure.title("无新建权限，新建算子不发布新建失败，新建算子并发布新建失败")
    def test_operator_perm_03(self):
        filepath = "./resource/openapi/compliant/template.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        # b、c都无新建权限
        for header in [TestOperatorPerm.b_headers, TestOperatorPerm.c_headers]:
            result = self.client.RegisterOperator(data, header)
            assert result[0] == 403
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        # b、c都无新建权限
        for header in [TestOperatorPerm.b_headers, TestOperatorPerm.c_headers]:
            result = self.client.RegisterOperator(data, header)
            assert result[0] == 403

    @allure.title("更新算子信息，更新后不发布，无编辑权限编辑失败，有编辑权限编辑成功")
    def test_operator_perm_04(self):
        operator_id = TestOperatorPerm.operators[0]["operator_id"]
        filepath = "./resource/openapi/compliant/update-test4.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "operator_id": operator_id,
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        # c无编辑权限
        result = self.client.EditOperator(data, TestOperatorPerm.c_headers)
        assert result[0] == 403
        # b有编辑权限
        result = self.client.EditOperator(data, TestOperatorPerm.b_headers)
        assert result[0] == 200

    @allure.title("更新算子信息，更新后发布，无编辑或发布权限编辑失败，有编辑和发布权限编辑成功")
    def test_operator_perm_05(self):
        operator_id = TestOperatorPerm.operators[0]["operator_id"]
        filepath = "./resource/openapi/compliant/update-test4.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "operator_id": operator_id,
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        # c有发布无编辑权限
        result = self.client.UpdateOperatorInfo(data, TestOperatorPerm.c_headers)
        assert result[0] == 403
        # assert result[1][0]["status"] == "failed"
        # b有编辑无发布权限
        result = self.client.UpdateOperatorInfo(data, TestOperatorPerm.b_headers)
        assert result[0] == 403
        # assert result[1][0]["status"] == "failed"
        # d有编辑和发布权限
        result = self.client.UpdateOperatorInfo(data, TestOperatorPerm.d_headers)
        assert result[0] == 200
        assert result[1][0]["status"] == "success"        

    @allure.title("获取算子列表，无法获取到无查看权限的算子，可获取到有查看权限的算子")
    def test_operator_perm_06(self):
        operator = TestOperatorPerm.operators[0]
        # d无查看权限
        result = self.client.GetOperatorList({}, TestOperatorPerm.d_headers)
        assert operator["operator_id"] not in [op["operator_id"] for op in result[1]["data"]]
        # b有查看权限
        result = self.client.GetOperatorList({}, TestOperatorPerm.b_headers)
        assert operator["operator_id"] in [op["operator_id"] for op in result[1]["data"]]

    @allure.title("编辑算子，无编辑权限编辑失败，有编辑权限编辑成功")
    def test_operator_perm_07(self):
        operator = TestOperatorPerm.operators[0]
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {"operator_id": operator["operator_id"], "name": name}
        # d有编辑权限
        result = self.client.EditOperator(data, TestOperatorPerm.d_headers)
        assert result[0] == 200
        # c无编辑权限
        result = self.client.EditOperator(data, TestOperatorPerm.c_headers)
        assert result[0] == 403

    @allure.title("获取算子信息，无查看权限获取失败，有查看权限编辑成功")
    def test_operator_perm_08(self):
        operator = TestOperatorPerm.operators[0]
        # d无查看权限
        result = self.client.GetOperatorInfo(operator["operator_id"], TestOperatorPerm.d_headers)
        assert result[0] == 403
        # b有查看权限
        result = self.client.GetOperatorInfo(operator["operator_id"], TestOperatorPerm.b_headers)
        assert result[0] == 200

    @allure.title("删除未发布算子，无删除权限删除失败，有删除权限删除成功")
    def test_operator_perm_09(self):
        operator = TestOperatorPerm.operators[6]
        del_data = [{"operator_id": operator["operator_id"], "version": operator["version"]}]
        # c无删除权限
        result = self.client.DeleteOperator(del_data, TestOperatorPerm.c_headers)
        assert result[0] == 403
        # b有删除权限
        result = self.client.DeleteOperator(del_data, TestOperatorPerm.b_headers)
        assert result[0] == 200

    @allure.title("发布算子，无发布权限发布失败，有发布权限发布成功")
    def test_operator_perm_10(self):
        operator = TestOperatorPerm.operators[1]
        update_data = [{"operator_id": operator["operator_id"], "status": "published"}]
        # b无发布权限
        result = self.client.UpdateOperatorStatus(update_data, TestOperatorPerm.b_headers)
        assert result[0] == 403
        # c有发布权限
        result = self.client.UpdateOperatorStatus(update_data, TestOperatorPerm.c_headers)
        assert result[0] == 200

    @allure.title("下架算子，无下架权限下架失败，有下架权限下架成功")
    def test_operator_perm_11(self):
        operator = TestOperatorPerm.operators[2]
        # c有发布权限
        update_data = [{"operator_id": operator["operator_id"], "status": "published"}]
        result = self.client.UpdateOperatorStatus(update_data, TestOperatorPerm.c_headers)
        assert result[0] == 200

        update_data = [{"operator_id": operator["operator_id"], "status": "offline"}]
        # b无下架权限
        result = self.client.UpdateOperatorStatus(update_data, TestOperatorPerm.b_headers)
        assert result[0] == 403
        # c有下架权限
        result = self.client.UpdateOperatorStatus(update_data, TestOperatorPerm.c_headers)
        assert result[0] == 200

    @allure.title("更新算子状态到editing状态，无编辑权限更新失败，有编辑权限更新成功")
    def test_operator_perm_12(self):
        operator = TestOperatorPerm.operators[3]
        # c有发布权限
        update_data = [{"operator_id": operator["operator_id"], "status": "published"}]
        result = self.client.UpdateOperatorStatus(update_data, TestOperatorPerm.c_headers)
        assert result[0] == 200

        update_data = [{"operator_id": operator["operator_id"], "status": "editing"}]
        # c无编辑权限
        result = self.client.UpdateOperatorStatus(update_data, TestOperatorPerm.c_headers)
        assert result[0] == 403
        # b有编辑权限
        result = self.client.UpdateOperatorStatus(update_data, TestOperatorPerm.b_headers)
        assert result[0] == 200

    @allure.title("调试算子，无使用权限调试失败，有使用权限调试成功")
    def test_operator_perm_13(self):
        operator = TestOperatorPerm.operators[4]
        debug_data = {
            "operator_id": operator["operator_id"],
            "version": operator["version"],
            "header": TestOperatorPerm.b_headers,
            "path": {"id": operator["operator_id"]}
        }
        # b无使用权限
        result = self.client.OperatorDebug(debug_data, TestOperatorPerm.b_headers)
        assert result[0] == 403
        # c有使用权限
        result = self.client.OperatorDebug(debug_data, TestOperatorPerm.c_headers)
        assert result[0] == 200

    @allure.title("获取算子历史版本详情，无查看和公开访问权限获取失败，有查看或公开访问权限获取成功")
    def test_operator_perm_14(self):
        operator = TestOperatorPerm.operators[5]
        # 发布算子
        update_data = [{"operator_id": operator["operator_id"], "status": "published"}] 
        result = self.client.UpdateOperatorStatus(update_data, TestOperatorPerm.c_headers)
        assert result[0] == 200
        # 发布后编辑生成历史版本
        re = self.client.GetOperatorInfo(operator["operator_id"], TestOperatorPerm.b_headers)
        assert re[0] == 200
        data = {
            "operator_id": operator["operator_id"],
            "description": "test edit 1234567"
        }
        re = self.client.EditOperator(data, TestOperatorPerm.b_headers)
        assert re[0] == 200
        # d无查看和公开访问权限
        result = self.client.GetOperatorHistoryDetail(operator["operator_id"], operator["version"], TestOperatorPerm.d_headers)
        assert result[0] == 403
        # b有查看权限
        result = self.client.GetOperatorHistoryDetail(operator["operator_id"], operator["version"], TestOperatorPerm.b_headers)
        assert result[0] == 200
        # c有公开访问权限
        result = self.client.GetOperatorHistoryDetail(operator["operator_id"], operator["version"], TestOperatorPerm.c_headers)
        assert result[0] == 200

    @allure.title("获取算子版本列表，无查看和公开访问权限获取失败，有查看或公开访问权限获取成功")
    def test_operator_perm_15(self):
        operator = TestOperatorPerm.operators[5]
        # d无查看和公开访问权限
        result = self.client.GetOperatorHistoryList(operator["operator_id"], TestOperatorPerm.d_headers)
        assert result[0] == 403
        # b有查看权限
        result = self.client.GetOperatorHistoryList(operator["operator_id"], TestOperatorPerm.b_headers)
        assert result[0] == 200
        # c有公开访问权限
        result = self.client.GetOperatorHistoryList(operator["operator_id"], TestOperatorPerm.b_headers)
        assert result[0] == 200

    @allure.title("获取算子市场列表，无法获取到无公开访问权限的算子，可获取到有公开访问权限的算子")
    def test_operator_perm_16(self):
        operator = TestOperatorPerm.operators[5]
        filepath = "./resource/openapi/compliant/test3.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        create_result = self.client.RegisterOperator(data, TestOperatorPerm.a_headers)
        assert create_result[0] == 200
        operator_id = create_result[1][0]["operator_id"]
        # c对AI管理员创建的算子（setup中）有公开访问权限，对用户a创建的算子无公开访问权限
        result = self.client.GetOperatorMarketList({"all": True}, TestOperatorPerm.c_headers)
        assert result[0] == 200
        assert operator["operator_id"] in [op["operator_id"] for op in result[1]["data"]]
        assert operator_id not in [op["operator_id"] for op in result[1]["data"]]
        # d对AI管理员创建的算子（setup中）无公开访问权限
        result = self.client.GetOperatorMarketList({"all": True}, TestOperatorPerm.d_headers)
        assert result[0] == 200
        assert operator["operator_id"] not in [op["operator_id"] for op in result[1]["data"]]

    @allure.title("获取算子市场指定算子详情，无公开访问权限获取失败，有公开访问权限获取成功")
    def test_operator_perm_17(self):
        operator = TestOperatorPerm.operators[5]
        # d无公开访问权限
        result = self.client.GetOperatorMarketDetail(operator["operator_id"], TestOperatorPerm.d_headers)
        assert result[0] == 403
        # c有公开访问权限
        result = self.client.GetOperatorMarketDetail(operator["operator_id"], TestOperatorPerm.c_headers)
        assert result[0] == 200

    @allure.title("注册内置算子（外部接口），无新建权限注册失败，有新建权限注册成功")
    def test_operator_perm_18(self):
        operator_id = str(uuid.uuid4())
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
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
        # b无新建权限
        result = self.client.RegisterBuiltinOperator(data, TestOperatorPerm.b_headers)
        assert result[0] == 403
        # d有新建权限
        result = self.client.RegisterBuiltinOperator(data, TestOperatorPerm.d_headers)
        assert result[0] == 200

    @allure.title("编辑内置算子（外部接口），无编辑权限编辑失败，有编辑权限编辑成功")
    def test_operator_perm_19(self, Headers):
        operator_id = str(uuid.uuid4())
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
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
        # 新建
        result = self.client.RegisterBuiltinOperator(data, TestOperatorPerm.a_headers)
        assert result[0] == 200
        # 给用户b配置编辑权限
        perm_data = [{
                "accessor": {"id": TestOperatorPerm.user_list[1], "name": "b", "type": "user"},
                "resource": {"id": operator_id, "type": "operator", "name": "编辑权限"},
                "operation": {"allow": [{"id": "modify"}], "deny": []}
            }]
        result = self.perm_client.SetPerm(perm_data, TestOperatorPerm.a_headers)
        assert "20" in str(result[0])
        # c无编辑权限
        result = self.client.RegisterBuiltinOperator(data, TestOperatorPerm.c_headers)
        assert result[0] == 403
        # b有编辑权限
        result = self.client.RegisterBuiltinOperator(data, TestOperatorPerm.b_headers)
        assert result[0] == 200

    @allure.title("新建内置算子，AI管理员可对该内置算子进行所有操作，普通用户可公开访问和使用该内置算子")
    def test_operator_perm_20(self, Headers):
        operator_id = str(uuid.uuid4())
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
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
        # AI管理员新建内置算子
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 200
        version = result[1]["version"]
        # 普通用户可公开访问和使用
        result = self.client.GetOperatorMarketDetail(operator_id, TestOperatorPerm.b_headers)
        assert result[0] == 200
        debug_data = {
            "operator_id": operator_id,
            "version": version,
            "header": TestOperatorPerm.b_headers,
            "path": {"id": operator_id}
        }
        result = self.client.OperatorDebug(debug_data, TestOperatorPerm.b_headers)
        assert result[0] == 200
        # 普通用户不可编辑
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        update_data = {
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
        result = self.client.RegisterBuiltinOperator(update_data, TestOperatorPerm.b_headers)
        assert result[0] == 403
        # AI管理员给用户b配置权限
        perm_data = [{
                "accessor": {"id": TestOperatorPerm.user_list[2], "name": "c", "type": "user"},
                "resource": {"id": operator_id, "type": "operator", "name": "配置权限"},
                "operation": {"allow": [{"id": "view"}], "deny": []}
            }]
        result = self.perm_client.SetPerm(perm_data, Headers)
        assert "20" in str(result[0])
        # AI管理员可编辑可下架
        result = self.client.RegisterBuiltinOperator(update_data, Headers)
        assert result[0] == 200
        update_data = [{"operator_id": operator_id, "status": "offline"}] 
        result = self.client.UpdateOperatorStatus(update_data, Headers) 
        assert result[0] == 200

    @allure.title("代理执行已发布算子，无使用权限执行失败，有使用权限执行成功")
    def test_operator_perm_21(self):
        internal_client = InternalOperator()
        operator = TestOperatorPerm.operators[5]
        proxy_data = {
            "header": TestOperatorPerm.b_headers,
            "body": {"id": operator["operator_id"], "action": "publish"}
        }
        b_headers = {
            "x-account-id": TestOperatorPerm.user_list[1]
        }
        c_headers = {
            "x-account-id": TestOperatorPerm.user_list[2]
        }
        # b无使用权限
        result = internal_client.ProxyOperator(operator["operator_id"], proxy_data, b_headers)
        assert result[0] == 403
        # c有使用权限
        result = internal_client.ProxyOperator(operator["operator_id"], proxy_data, c_headers)
        assert result[0] == 200
    