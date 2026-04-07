# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import uuid

from common.get_content import GetContent
from common.get_token import GetToken
from lib.tool_box import ToolBox
from lib.operator import Operator
from lib.permission import Perm


@allure.feature("算子平台权限测试：工具箱权限测试")
class TestToolboxPerm:
    client = ToolBox()
    client_operator = Operator()
    perm_client = Perm()
    a_headers = {}
    b_headers = {}
    c_headers = {}
    d_headers = {}
    box_ids = []
    user_list = []

    @pytest.fixture(scope="class", autouse=True)
    # def setup(self, Headers):
    def setup(self, Headers, PermPrepare):
        # 获取用户token
        configfile = "./config/env.ini"
        file = GetContent(configfile)
        config = file.config()
        host = config["server"]["host"]
        a_token = GetToken(host=host).get_token(host, "a", "111111")
        TestToolboxPerm.a_headers = {
            "Authorization": f"Bearer {a_token[1]}"
        }

        b_token = GetToken(host=host).get_token(host, "b", "111111")
        TestToolboxPerm.b_headers = {
            "Authorization": f"Bearer {b_token[1]}"
        }

        c_token = GetToken(host=host).get_token(host, "c", "111111")
        TestToolboxPerm.c_headers = {
            "Authorization": f"Bearer {c_token[1]}"
        }

        d_token = GetToken(host=host).get_token(host, "d", "111111")
        TestToolboxPerm.d_headers = {
            "Authorization": f"Bearer {d_token[1]}"
        }

        # AI管理员创建工具箱
        for i in range(5):
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            filepath = "./resource/openapi/compliant/mcp.yaml"
            yaml_data = GetContent(filepath).yamlfile()

            data = {
                "box_name": name,
                "data": yaml_data,
                "metadata_type": "openapi"
            }
            result = self.client.CreateToolbox(data, Headers)
            assert result[0] == 200
            box_id = result[1]["box_id"]
            TestToolboxPerm.box_ids.append(box_id)

        '''
        为用户a设置算子资源的新建权限
        为用户a设置工具箱资源的新建权限
        为用户b设置工具箱的查看、编辑、删除权限
        为用户c设置工具箱的发布、下架、公共访问和使用权限
        为用户d设置工具箱的新建、编辑、发布权限
        '''
        TestToolboxPerm.user_list = PermPrepare[1]
        # TestToolboxPerm.user_list = ['f86d7374-b3a7-11f0-a157-c6666f116050', 'f87fe90a-b3a7-11f0-bb3f-c6666f116050', 'f894c348-b3a7-11f0-97e9-c6666f116050', 'f8a6ed84-b3a7-11f0-9bf1-c6666f116050']
        perm_data = [
            {
                "accessor": {"id": TestToolboxPerm.user_list[0], "name": "a", "type": "user"},
                "resource": {"id": "*", "type": "tool_box", "name": ""},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            },
            {
                "accessor": {"id": TestToolboxPerm.user_list[3], "name": "d", "type": "user"},
                "resource": {"id": "*", "type": "tool_box", "name": ""},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            },
            {
                "accessor": {"id": TestToolboxPerm.user_list[0], "name": "a", "type": "user"},
                "resource": {"id": "*", "type": "operator", "name": ""},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            }
        ]
        for id in TestToolboxPerm.box_ids:
            b_data = {
                "accessor": {"id": TestToolboxPerm.user_list[1], "name": "b", "type": "user"},
                "resource": {"id": id, "type": "tool_box", "name": "查看编辑删除权限"},
                "operation": {"allow": [{"id": "view"}, {"id": "modify"}, {"id": "delete"}], "deny": []}
            }
            c_data = {
                "accessor": {"id": TestToolboxPerm.user_list[2], "name": "c", "type": "user"},
                "resource": {"id": id, "type": "tool_box", "name": "发布下架公开访问使用权限"},
                "operation": { "allow": [{"id": "publish"}, {"id": "unpublish"}, {"id": "public_access"}, {"id": "execute"}],  "deny": []}
            }
            d_data = {
                "accessor": {"id": TestToolboxPerm.user_list[3], "name": "d", "type": "user"},
                "resource": {"id": id, "type": "tool_box", "name": "发布编辑用权限"},
                "operation": {"allow": [{"id": "publish"}, {"id": "modify"}], "deny": []}
            }
            perm_data.append(b_data)
            perm_data.append(c_data)
            perm_data.append(d_data)
        result = self.perm_client.SetPerm(perm_data, Headers)
        assert "20" in str(result[0])

    @allure.title("有新建权限，新建工具箱，新建成功，创建者和AI管理员可对该工具箱进行所有操作")
    def test_toolbox_perm_01(self, Headers):
        # 新建工具箱
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/relations.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        create_data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(create_data, TestToolboxPerm.a_headers)
        assert result[0] == 200
        box_id = result[1]["box_id"]
        for header in [TestToolboxPerm.a_headers, Headers]:
            # 更新工具箱
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            update_data = {
                "box_name": name,
                "box_desc": "test toolbox update description 22",
                "box_svc_url": "http://127.0.0.1",
                "box_icon": "icon-color-tool-FADB14",
                "box_category": "data_process"
            }
            result = self.client.UpdateToolbox(box_id, update_data, header)
            assert result[0] == 200
            # 获取工具箱详情
            result = self.client.GetToolbox(box_id, header)
            assert result[0] == 200
            assert result[1]["box_id"] == box_id
            # 获取工具箱列表
            result = self.client.GetToolboxList({"all": True}, header)
            assert result[0] == 200
            assert box_id in str(result[1]["data"])
            # 发布工具箱
            result = self.client.UpdateToolboxStatus(box_id, {"status": "published"}, header)
            assert result[0] == 200
            # 下架工具箱
            result = self.client.UpdateToolboxStatus(box_id, {"status": "offline"}, header)
            assert result[0] == 200
            # 获取工具箱内工具列表
            result = self.client.GetBoxToolsList(box_id, None, header)
            assert result[0] == 200
            tool_list = result[1]["tools"]
            # 创建工具
            filepath = "./resource/openapi/compliant/tool.json"
            data = GetContent(filepath).jsonfile()
            tool_data = {"data": data, "metadata_type": "openapi"}
            result = self.client.CreateTool(box_id, tool_data, header)
            assert result[0] == 200
            tool_id = result[1]["success_ids"][0]
            # 更新工具
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            update_data = {
                "name": name,
                "description": "test tool update description"
            }
            result = self.client.UpdateTool(box_id, tool_id, update_data, header)
            assert result[0] == 200
            # 获取工具详情
            result = self.client.GetTool(box_id, tool_id, header)
            assert result[0] == 200
            assert result[1]["tool_id"] == tool_id
            # 批量删除工具
            tool_ids = [tool_id, tool_list[0]["tool_id"]]
            result = self.client.BatchDeleteTools(box_id, {"tool_ids": tool_ids}, header)
            assert result[0] == 200
            # 更新工具状态
            update_data = [{"tool_id": tool_list[1]["tool_id"], "status": "enabled"},
                {"tool_id": tool_list[2]["tool_id"], "status": "enabled"}]
            result = self.client.UpdateToolStatus(box_id, update_data, header)
            assert result[0] == 200
            # 工具调试
            debug_data = {"header": Headers, "path": {"box_id": box_id}}
            result = self.client.DebugTool(box_id, tool_list[1]["tool_id"], debug_data, header)
            assert result[0] == 200
            # 工具执行
            result = self.client.ProxyTool(box_id, tool_list[1]["tool_id"], debug_data, header)
            assert result[0] == 200
            # 删除工具箱
            result = self.client.DeleteToolbox(box_id, header)
            assert result[0] == 200
            # 创建
            result = self.client.CreateToolbox(create_data, TestToolboxPerm.a_headers)
            assert result[0] == 200
            box_id = result[1]["box_id"]

    @allure.title("新建工具箱，无权限新建失败，有权限新建成功")
    def test_toolbox_perm_02(self):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        # b、c都无新建权限
        for header in [TestToolboxPerm.b_headers, TestToolboxPerm.c_headers]:
            result = self.client.CreateToolbox(data, header)
            assert result[0] == 403
        # d有新建权限
        result = self.client.CreateToolbox(data, TestToolboxPerm.d_headers)
        assert result[0] == 200

    @allure.title("更新工具箱，无编辑权限编辑失败，有编辑权限编辑成功")
    def test_toolbox_perm_03(self):
        box_id = TestToolboxPerm.box_ids[0]
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        update_data = {
            "box_name": name,
            "box_desc": "test toolbox update description",
            "box_svc_url": "http://127.0.0.1",
            "box_icon": "icon-color-tool-FADB14",
            "box_category": "data_process"
        }
        # c无编辑权限
        result = self.client.UpdateToolbox(box_id, update_data, TestToolboxPerm.c_headers)
        assert result[0] == 403
        # b有编辑权限
        result = self.client.UpdateToolbox(box_id, update_data, TestToolboxPerm.b_headers)
        assert result[0] == 200

    @allure.title("获取工具箱详情，无查看权限获取失败，有查看权限获取成功")
    def test_toolbox_perm_04(self):
        box_id = TestToolboxPerm.box_ids[0]
        # d无查看权限
        result = self.client.GetToolbox(box_id, TestToolboxPerm.d_headers)
        assert result[0] == 403
        # b有查看权限
        result = self.client.GetToolbox(box_id, TestToolboxPerm.b_headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id

    @allure.title("发布工具箱，无发布权限发布失败，有发布权限发布成功")
    def test_toolbox_perm_05(self):
        box_id = TestToolboxPerm.box_ids[0]
        update_data = {"status": "published"}
        # b无发布权限
        result = self.client.UpdateToolboxStatus(box_id, update_data, TestToolboxPerm.b_headers)
        assert result[0] == 403
        # c有发布权限
        result = self.client.UpdateToolboxStatus(box_id, update_data, TestToolboxPerm.c_headers)
        assert result[0] == 200

    @allure.title("下架工具箱，无下架权限下架失败，有下架权限下架成功")
    def test_toolbox_perm_06(self):
        box_id = TestToolboxPerm.box_ids[1]
        # 先发布工具箱
        update_data = {"status": "published"}
        result = self.client.UpdateToolboxStatus(box_id, update_data, TestToolboxPerm.c_headers)
        assert result[0] == 200

        update_data = {"status": "offline"}
        # b无下架权限
        result = self.client.UpdateToolboxStatus(box_id, update_data, TestToolboxPerm.b_headers)
        assert result[0] == 403
        # c有下架权限
        result = self.client.UpdateToolboxStatus(box_id, update_data, TestToolboxPerm.c_headers)
        assert result[0] == 200

    @allure.title("获取工具箱列表，不能获取到无查看权限的工具箱，可以获取到有查看权限的工具箱")
    def test_toolbox_perm_08(self):
        params = {"all": True}
        # d对a创建的工具箱（setup中）无查看权限
        result = self.client.GetToolboxList(params, TestToolboxPerm.d_headers)
        assert result[0] == 200
        for box_id in TestToolboxPerm.box_ids:
            assert box_id not in str(result[1]["data"])
        # b对a创建的工具箱（setup中）有查看权限
        result = self.client.GetToolboxList(params, TestToolboxPerm.b_headers)
        assert result[0] == 200
        for box_id in TestToolboxPerm.box_ids:
            assert box_id in str(result[1]["data"])

    @allure.title("批量获取市场中工具箱信息，无公开访问权限获取失败，有公开访问权限获取成功")
    def test_toolbox_perm_09(self):
        box_id = TestToolboxPerm.box_ids[0]
        # d无公开访问权限
        result = self.client.GetMarketDetail(box_id, "name,description", TestToolboxPerm.d_headers)
        assert result[0] == 200
        assert result[1] == []
        # c有公开访问权限
        result = self.client.GetMarketDetail(box_id, "name,description", TestToolboxPerm.c_headers)
        assert result[0] == 200
        assert result[1][0]["box_id"] == box_id

    @allure.title("获取市场中工具箱详情，无公开访问权限获取失败，有公开访问权限获取成功")
    def test_toolbox_perm_10(self):
        box_id = TestToolboxPerm.box_ids[0]
        # d无公开访问权限
        result = self.client.GetMarketToolbox(box_id, TestToolboxPerm.d_headers)
        assert result[0] == 403
        # c有公开访问权限
        result = self.client.GetMarketToolbox(box_id, TestToolboxPerm.c_headers)
        assert result[0] == 200

    @allure.title("获取市场中工具箱列表，无法获取到无公开访问权限的工具箱，可以获取到有公开访问权限的工具箱")
    def test_toolbox_perm_11(self):
        # 用户a新建工具箱并发布
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, TestToolboxPerm.a_headers)
        assert result[0] == 200
        box_id = result[1]["box_id"]
        result = self.client.UpdateToolboxStatus(box_id, {"status": "published"}, TestToolboxPerm.a_headers)
        assert result[0] == 200
        # 给d配置公开访问权限
        perm_data = [{
                "accessor": {"id": TestToolboxPerm.user_list[3], "name": "d", "type": "user"},
                "resource": {"id": box_id, "type": "tool_box", "name": "公开访问权限"},
                "operation": {"allow": [{"id": "public_access"}], "deny": []}
            }]
        result = self.perm_client.SetPerm(perm_data, TestToolboxPerm.a_headers)
        assert "20" in str(result[0])
        # d对a在setup中创建的工具箱无公开访问权限，对a在本用例中创建的工具箱有公开访问权限
        result = self.client.GetMarketToolboxList({"all": True}, TestToolboxPerm.d_headers)
        assert result[0] == 200
        assert box_id in str(result[1]["data"])
        assert TestToolboxPerm.box_ids[0] not in str(result[1]["data"])
        # c对a在setup中创建的工具箱有公开访问权限，对a在本用例中创建的工具箱无公开访问权限
        result = self.client.GetMarketToolboxList({"all": True}, TestToolboxPerm.c_headers)
        assert result[0] == 200
        assert box_id not in str(result[1]["data"])
        assert TestToolboxPerm.box_ids[0] in str(result[1]["data"])

    @allure.title("新建工具，对所属工具箱无编辑权限新建失败，对所属工具箱有编辑权限新建成功")
    def test_toolbox_perm_12(self):
        box_id = TestToolboxPerm.box_ids[2]
        filepath = "./resource/openapi/compliant/tool.json"
        data = GetContent(filepath).jsonfile()
        tool_data = {"data": data, "metadata_type": "openapi"}
        # c无编辑权限
        result = self.client.CreateTool(box_id, tool_data, TestToolboxPerm.c_headers)
        assert result[0] == 403
        # b有编辑权限
        result = self.client.CreateTool(box_id, tool_data, TestToolboxPerm.b_headers)
        assert result[0] == 200

    @allure.title("编辑工具，对所属工具箱无编辑权限编辑失败，对所属工具箱有编辑权限编辑成功")
    def test_toolbox_perm_13(self):
        box_id = TestToolboxPerm.box_ids[2]
        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, None, TestToolboxPerm.b_headers)
        assert result[0] == 200
        tool_id = result[1]["tools"][0]["tool_id"]

        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        update_data = {
            "name": name,
            "description": "test tool update description"
        }
        # c无编辑权限
        result = self.client.UpdateTool(box_id, tool_id, update_data, TestToolboxPerm.c_headers)
        assert result[0] == 403
        # b有编辑权限
        result = self.client.UpdateTool(box_id, tool_id, update_data, TestToolboxPerm.b_headers)
        assert result[0] == 200

    @allure.title("获取工具，对所属工具箱无查看和公开访问权限获取失败，对所属工具箱有查看或公开访问权限获取成功")
    def test_toolbox_perm_14(self):
        box_id = TestToolboxPerm.box_ids[0]
        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, None, TestToolboxPerm.b_headers)
        assert result[0] == 200
        tool_id = result[1]["tools"][0]["tool_id"]

        # d无查看和公开访问权限
        result = self.client.GetTool(box_id, tool_id, TestToolboxPerm.d_headers)
        assert result[0] == 403
        # b有查看权限
        result = self.client.GetTool(box_id, tool_id, TestToolboxPerm.b_headers)
        assert result[0] == 200
        # c有公开访问权限
        result = self.client.GetTool(box_id, tool_id, TestToolboxPerm.c_headers)
        assert result[0] == 200

    @allure.title("批量删除工具，对所属工具箱无编辑权限删除失败，对所属工具箱有编辑权限删除成功")
    def test_toolbox_perm_15(self):
        box_id = TestToolboxPerm.box_ids[2]
        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, None, TestToolboxPerm.b_headers)
        assert result[0] == 200
        tool_ids = [result[1]["tools"][0]["tool_id"], result[1]["tools"][1]["tool_id"]]

        delete_data = {"tool_ids": tool_ids}
        # c无编辑权限
        result = self.client.BatchDeleteTools(box_id, delete_data, TestToolboxPerm.c_headers)
        assert result[0] == 403
        # b有编辑权限
        result = self.client.BatchDeleteTools(box_id, delete_data, TestToolboxPerm.b_headers)
        assert result[0] == 200

    @allure.title("获取工具箱内工具列表，对所属工具箱无查看和公开访问权限获取失败，对所属工具箱有查看或公开访问权限获取成功")
    def test_toolbox_perm_16(self):
        box_id = TestToolboxPerm.box_ids[0]
        # d无查看和公开访问权限
        result = self.client.GetBoxToolsList(box_id, None, TestToolboxPerm.d_headers)
        assert result[0] == 403
        # b有查看权限
        result = self.client.GetBoxToolsList(box_id, None, TestToolboxPerm.b_headers)
        assert result[0] == 200
        # c有公开访问权限
        result = self.client.GetBoxToolsList(box_id, None, TestToolboxPerm.c_headers)
        assert result[0] == 200

    @allure.title("更新工具状态，对所属工具箱无编辑权限更新失败，对所属工具箱有编辑权限更新成功")
    def test_toolbox_perm_17(self):
        box_id = TestToolboxPerm.box_ids[2]
        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, None, TestToolboxPerm.b_headers)
        assert result[0] == 200
        tool_id = result[1]["tools"][0]["tool_id"]

        update_data = [{"tool_id": tool_id, "status": "enabled"}]
        # c无编辑权限
        result = self.client.UpdateToolStatus(box_id, update_data, TestToolboxPerm.c_headers)
        assert result[0] == 403
        # b有编辑权限
        result = self.client.UpdateToolStatus(box_id, update_data, TestToolboxPerm.b_headers)
        assert result[0] == 200

    @allure.title("调试工具，对所属工具箱无使用权限调试失败，对所属工具箱有使用权限调试成功")
    def test_toolbox_perm_18(self):
        box_id = TestToolboxPerm.box_ids[3]
        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, None, TestToolboxPerm.b_headers)
        assert result[0] == 200
        tool_id = result[1]["tools"][0]["tool_id"]

        debug_data = {"header": TestToolboxPerm.b_headers, "path": {"box_id": box_id}}
        # b无使用权限
        result = self.client.DebugTool(box_id, tool_id, debug_data, TestToolboxPerm.b_headers)
        assert result[0] == 403
        # c有使用权限
        result = self.client.DebugTool(box_id, tool_id, debug_data, TestToolboxPerm.c_headers)
        assert result[0] == 200

    @allure.title("工具执行，对所属工具箱无使用权限执行失败，对所属工具箱有使用权限执行成功")
    def test_toolbox_perm_19(self):
        box_id = TestToolboxPerm.box_ids[3]
        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, None, TestToolboxPerm.b_headers)
        assert result[0] == 200
        tool_id = result[1]["tools"][0]["tool_id"]
        # 启用工具
        update_data = [{"tool_id": tool_id, "status": "enabled"}]
        result = self.client.UpdateToolStatus(box_id, update_data, TestToolboxPerm.b_headers)
        assert result[0] == 200

        debug_data = {"header": TestToolboxPerm.b_headers, "path": {"box_id": box_id}}
        # b无使用权限
        result = self.client.ProxyTool(box_id, tool_id, debug_data, TestToolboxPerm.b_headers)
        assert result[0] == 403
        # c有使用权限
        result = self.client.ProxyTool(box_id, tool_id, debug_data, TestToolboxPerm.c_headers)
        assert result[0] == 200

    @allure.title("算子转换成工具，对所属工具箱无编辑权限，或对算子无公开访问权限，或对算子无使用权限转换失败")
    def test_toolbox_perm_20(self):
        box_id = TestToolboxPerm.box_ids[4]
        # 用户a注册算子
        filepath = "./resource/openapi/compliant/edit-test1.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        create_result = self.client_operator.RegisterOperator(data, TestToolboxPerm.a_headers)
        assert create_result[0] == 200
        operator_id = create_result[1][0]["operator_id"]
        # 发布算子
        update_data = [{"operator_id": operator_id, "status": "editing"}]
        result = self.client_operator.UpdateOperatorStatus(update_data, TestToolboxPerm.a_headers)
        assert result[0] == 200
        # 给用户b配置该算子的公开访问权限，给用户d配置该算子的使用权限
        perm_data = [{
                "accessor": {"id": TestToolboxPerm.user_list[1], "name": "b", "type": "user"},
                "resource": {"id": operator_id, "type": "operator", "name": "公开访问权限"},
                "operation": {"allow": [{"id": "public_access"}], "deny": []}},
            {
                "accessor": {"id": TestToolboxPerm.user_list[3], "name": "d", "type": "user"},
                "resource": {"id": operator_id, "type": "operator", "name": "使用权限"},
                "operation": {"allow": [{"id": "execute"}], "deny": []}
            }]
        result = self.perm_client.SetPerm(perm_data, TestToolboxPerm.a_headers)
        assert "20" in str(result[0])

        convert_data = {
            "box_id": box_id,
            "operator_id": operator_id
        }
        # c对工具箱无编辑权限
        result = self.client.ConvertOperatorToTool(convert_data, TestToolboxPerm.c_headers)
        assert result[0] == 403
        # d对工具箱有编辑权限，对算子有使用权限但对无公开访问权限
        result = self.client.ConvertOperatorToTool(convert_data, TestToolboxPerm.d_headers)
        assert result[0] == 403
        # b对工具箱有编辑权限，对算子有公开访问权限但对无使用权限
        result = self.client.ConvertOperatorToTool(convert_data, TestToolboxPerm.b_headers)
        assert result[0] == 403

    @allure.title("算子转换成工具，对所属工具箱有编辑权限，对算子有查看和使用权限转换成功")
    def test_toolbox_perm_21(self):
        box_id = TestToolboxPerm.box_ids[4]
        # 用户a注册算子
        filepath = "./resource/openapi/compliant/edit-test2.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        create_result = self.client_operator.RegisterOperator(data, TestToolboxPerm.a_headers)
        assert create_result[0] == 200
        operator_id = create_result[1][0]["operator_id"]
        # 发布算子
        update_data = [{"operator_id": operator_id, "status": "editing"}]
        result = self.client_operator.UpdateOperatorStatus(update_data, TestToolboxPerm.a_headers)
        assert result[0] == 200
        # 给用户b配置该算子的公开访问和使用权限，给用户d配置该算子的使用权限
        perm_data = [{
                "accessor": {"id": TestToolboxPerm.user_list[1], "name": "b", "type": "user"},
                "resource": {"id": operator_id, "type": "operator", "name": "公开访问权限"},
                "operation": {"allow": [{"id": "public_access"}, {"id": "execute"}], "deny": []}
            }]
        result = self.perm_client.SetPerm(perm_data, TestToolboxPerm.a_headers)
        assert "20" in str(result[0])
        # 获取算子列表
        result = self.client_operator.GetOperatorList({"all": True}, TestToolboxPerm.a_headers)
        assert result[0] == 200
        operator_id = result[1]["data"][0]["operator_id"]

        convert_data = {
            "box_id": box_id,
            "operator_id": operator_id
        }
        # b对工具箱有编辑权限，对算子有公开访问和使用权限
        result = self.client.ConvertOperatorToTool(convert_data, TestToolboxPerm.b_headers)
        assert result[0] == 200

    @allure.title("删除工具箱，无删除权限删除失败，有删除权限删除成功")
    def test_toolbox_perm_22(self):
        # 创建新工具箱用于删除测试
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, TestToolboxPerm.a_headers)
        assert result[0] == 200
        box_id = result[1]["box_id"]
        # 给用户c配置该工具箱的删除权限
        perm_data = [{
                "accessor": {"id": TestToolboxPerm.user_list[2], "name": "b", "type": "user"},
                "resource": {"id": box_id, "type": "tool_box", "name": "公开访问权限"},
                "operation": {"allow": [{"id": "delete"}], "deny": []}
            }]
        result = self.perm_client.SetPerm(perm_data, TestToolboxPerm.a_headers)
        assert "20" in str(result[0])
        # b无删除权限
        result = self.client.DeleteToolbox(box_id, TestToolboxPerm.b_headers)
        assert result[0] == 403
        # c有删除权限
        result = self.client.DeleteToolbox(box_id, TestToolboxPerm.c_headers)
        assert result[0] == 200

    @allure.title("创建内置工具箱（外部接口），无新建创建失败，有新建创建成功")
    def test_toolbox_perm_23(self):
        box_id = str(uuid.uuid4())
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        # c无新建权限
        result = self.client.Builtin(data, TestToolboxPerm.c_headers)
        assert result[0] == 403
        # a有新建权限
        result = self.client.Builtin(data, TestToolboxPerm.a_headers)
        assert result[0] == 200

    @allure.title("编辑内置工具箱（外部接口），无编辑权限编辑失败，有编辑权限编辑成功")
    def test_toolbox_perm_24(self):
        box_id = str(uuid.uuid4())
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        # 用户a新建
        result = self.client.Builtin(data, TestToolboxPerm.a_headers)
        assert result[0] == 200
        # 给用户b配置编辑权限
        perm_data = [{
                "accessor": {"id": TestToolboxPerm.user_list[1], "name": "b", "type": "user"},
                "resource": {"id": box_id, "type": "tool_box", "name": "编辑权限"},
                "operation": {"allow": [{"id": "modify"}], "deny": []}
            }]
        result = self.perm_client.SetPerm(perm_data, TestToolboxPerm.a_headers)
        assert "20" in str(result[0])
        # c无编辑权限
        result = self.client.Builtin(data, TestToolboxPerm.c_headers)
        assert result[0] == 403
        # b有编辑权限
        result = self.client.Builtin(data, TestToolboxPerm.b_headers)
        assert result[0] == 200

    @allure.title("新建内置工具箱，AI管理员可对该内置工具箱进行所有操作，普通用户可公开访问和使用该内置工具箱")
    def test_toolbox_perm_25(self, Headers):
        # AI管理员创建内置工具箱
        box_id = str(uuid.uuid4())
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 200
        box_id = result[1]["box_id"]

        # AI管理员可进行所有操作
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        update_data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test builtin toolbox update",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Builtin(update_data, Headers)    # 编辑工具箱
        assert result[0] == 200

        result = self.client.GetToolbox(box_id, Headers)      # 查看工具箱详情
        assert result[0] == 200

        result = self.client.GetBoxToolsList(box_id, None, Headers)   # 获取工具箱内工具列表
        assert result[0] == 200

        result = self.client.GetMarketToolbox(box_id, Headers)  # 公开访问
        assert result[0] == 200

        # 给用户b配置权限
        perm_data = [{
                "accessor": {"id": TestToolboxPerm.user_list[1], "name": "b", "type": "user"},
                "resource": {"id": box_id, "type": "tool_box", "name": "配置权限"},
                "operation": {"allow": [{"id": "view"}], "deny": []}
            }]
        result = self.perm_client.SetPerm(perm_data, Headers)
        assert "20" in str(result[0])

        # 普通用户不可编辑
        result = self.client.Builtin(update_data, TestToolboxPerm.c_headers)
        assert result[0] == 403

        # 普通用户可公开访问
        result = self.client.GetMarketToolbox(box_id, TestToolboxPerm.c_headers)
        assert result[0] == 200

        # 普通用户可使用工具
        result = self.client.GetBoxToolsList(box_id, None, TestToolboxPerm.c_headers)
        assert result[0] == 200
        if result[1]["tools"]:
            tool_id = result[1]["tools"][0]["tool_id"]
            debug_data = {"header": TestToolboxPerm.c_headers, "path": {"box_id": box_id}}
            result = self.client.DebugTool(box_id, tool_id, debug_data, TestToolboxPerm.c_headers)
            assert result[0] == 200

    @allure.title("查询工具，可查询到有公开访问权限的工具，无法查询到无公开访问权限的工具")
    def test_toolbox_perm_26(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/tool.json"
        api_data = GetContent(filepath).jsonfile()

        data = {
            "box_name": name,
            "data": api_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 200
        box_id = result[1]["box_id"]
        result = self.client.UpdateToolboxStatus(box_id, {"status": "published"}, Headers)
        assert result[0] == 200
        # 给用户b配置公开访问权限
        perm_data = [{
                "accessor": {"id": TestToolboxPerm.user_list[1], "name": "b", "type": "user"},
                "resource": {"id": box_id, "type": "tool_box", "name": "公开访问权限"},
                "operation": {"allow": [{"id": "public_access"}], "deny": []}
            }]
        result = self.perm_client.SetPerm(perm_data, Headers)
        assert "20" in str(result[0])
        box_id = TestToolboxPerm.box_ids[0]
        # d无公开访问权限
        result = self.client.GetMarketToolsList({"tool_name": "用户"}, TestToolboxPerm.d_headers)
        assert result[0] == 200
        assert result[1]["data"] == []
        # b有公开访问权限
        result = self.client.GetMarketToolsList({"tool_name": "用户"}, TestToolboxPerm.b_headers)
        assert result[0] == 200
        for box in result[1]["data"]:
            for tool in box["tools"]:
                assert "用户" in tool["name"]
