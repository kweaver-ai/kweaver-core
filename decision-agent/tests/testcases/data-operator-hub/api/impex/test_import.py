# -*- coding:UTF-8 -*-

import allure
import pytest
import os

from common.file_process import FileProcess

from lib.operator import Operator
from lib.tool_box import ToolBox
from lib.mcp import MCP
from lib.impex import Impex
from lib.permission import Perm

@allure.feature("算子平台测试：导入导入测试")
class TestImport:
    impex_client = Impex()
    operator_client = Operator()
    toolbox_client = ToolBox()
    mcp_client = MCP()
    file_client = FileProcess()
    perm_client = Perm()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, PrepareData):
        TestImport.operator_ids = PrepareData[0] # 前4个为基础算子，第1个为未发布状态，第2为下架状态，第3、4个为已发布状态；第5个为内置算子
        TestImport.toolbox_ids = PrepareData[1] # 前30个工具箱中的所有工具均为本地导入；第31个工具箱包含一个从算子导入的工具；第32个为内置工具箱
        TestImport.mcp_ids = PrepareData[2] # 第1个为自定义mcp；第2个为工具转换为的mcp，每个mcp存在30个工具，分别来自不同的工具箱；第3个为工具转换为的mcp，mcp下的工具为从算子导入的工具；第4个为内置mcp
        TestImport.t1_operator_id = PrepareData[3]
        TestImport.t1_toolbox_id = PrepareData[4]
        TestImport.t1_mcp_id =  PrepareData[5]
        TestImport.t2_headers = PrepareData[6]
        TestImport.user_t2 = PrepareData[7]
        TestImport.operator_id = PrepareData[8] # 转换成工具的算子id
        TestImport.operator_version = PrepareData[9] # 转换成工具的算子版本

    @allure.title("以新建模式导入，资源冲突，导入失败")
    def test_import_01(self, Headers):
        filenames = [
            "./testcases/api/data-operator-hub/impex/export/operator_unpublish.json",
            "./testcases/api/data-operator-hub/impex/export/toolbox_published.json",
            "./testcases/api/data-operator-hub/impex/export/mcp_custom.json"
        ]
        for file in filenames:
            if "operator" in file:
                type = "operator"
            if "toolbox" in file:
                type = "toolbox"
            if "mcp" in file:
                type = "mcp"
            files = {"data": (os.path.basename(file), open(file, "rb"))}
            result = self.impex_client.importation(type, files, {"mode":"create"}, Headers)
            assert result[0] == 409

    @allure.title("以更新模式导入，资源冲突，导入成功")
    def test_import_02(self, Headers):
        filenames = [
            "./testcases/api/data-operator-hub/impex/export/operator_unpublish.json",
            "./testcases/api/data-operator-hub/impex/export/toolbox_published.json",
            "./testcases/api/data-operator-hub/impex/export/mcp_custom.json"
        ]
        for file in filenames:
            if "operator" in file:
                type = "operator"
            if "toolbox" in file:
                type = "toolbox"
            if "mcp" in file:
                type = "mcp"
            files = {"data": (os.path.basename(file), open(file, "rb"))}
            result = self.impex_client.importation(type, files, {"mode":"upsert"}, Headers)
            assert result[0] == 201
    
    @allure.title("导入未发布算子，导入成功，状态为未发布")
    def test_import_03(self, Headers):
        id = TestImport.operator_ids[0]
        result = self.operator_client.GetOperatorInfo(id, Headers) # 获取算子信息
        assert result[0] == 200
        version = result[1]["version"]
        del_data = [
            {
                "operator_id": id,
                "version": version
            }
        ]
        result = self.operator_client.DeleteOperator(del_data, Headers) # 删除算子
        assert result[0] == 200
        result = self.operator_client.GetOperatorInfo(id, Headers) # 获取算子信息
        assert result[0] == 404
        file = "./testcases/api/data-operator-hub/impex/export/operator_unpublish.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("operator", files, {"mode":"upsert"}, Headers)   # 导入算子
        assert result[0] == 201
        result = self.operator_client.GetOperatorInfo(id, Headers) # 获取算子信息
        assert result[0] == 200
        assert result[1]["status"] == "unpublish"

    @allure.title("导入已发布算子，导入成功，状态为已发布")
    def test_import_04(self, Headers):
        id = TestImport.operator_ids[2]
        data = [
            {
                "operator_id": id,
                "status": "offline"
            }
        ]
        result = self.operator_client.UpdateOperatorStatus(data, Headers)  # 下架算子
        assert result[0] == 200  
        result = self.operator_client.GetOperatorInfo(id, Headers) # 获取算子信息
        assert result[0] == 200
        version = result[1]["version"]
        del_data = [
            {
                "operator_id": id,
                "version": version
            }
        ]
        result = self.operator_client.DeleteOperator(del_data, Headers) # 删除算子
        assert result[0] == 200
        result = self.operator_client.GetOperatorInfo(id, Headers) # 获取算子信息
        assert result[0] == 404
        file = "./testcases/api/data-operator-hub/impex/export/operator_published.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("operator", files, {"mode":"upsert"}, Headers)   # 导入算子
        assert result[0] == 201
        result = self.operator_client.GetOperatorInfo(id, Headers) # 获取算子信息
        assert result[0] == 200
        assert result[1]["status"] == "published"

    @allure.title("导入已下架算子，导入成功，状态为已下架")
    def test_import_05(self, Headers):
        id = TestImport.operator_ids[1]
        result = self.operator_client.GetOperatorInfo(id, Headers) # 获取算子信息
        assert result[0] == 200
        version = result[1]["version"]
        del_data = [
            {
                "operator_id": id,
                "version": version
            }
        ]
        result = self.operator_client.DeleteOperator(del_data, Headers) # 删除算子
        assert result[0] == 200
        result = self.operator_client.GetOperatorInfo(id, Headers) # 获取算子信息
        assert result[0] == 404
        file = "./testcases/api/data-operator-hub/impex/export/operator_offline.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("operator", files, {"mode":"upsert"}, Headers)   # 导入算子
        assert result[0] == 201
        result = self.operator_client.GetOperatorInfo(id, Headers) # 获取算子信息
        assert result[0] == 200
        assert result[1]["status"] == "offline"

    @allure.title("导入工具箱，导入成功")
    def test_import_06(self, Headers):
        # 先删除对应工具箱并确认删除
        id = TestImport.toolbox_ids[0]
        result = self.toolbox_client.UpdateToolboxStatus(id, {"status": "offline"}, Headers)
        assert result[0] == 200
        result = self.toolbox_client.DeleteToolbox(id, Headers)
        assert result[0] == 200
        result = self.toolbox_client.GetToolbox(id, Headers)
        assert result[0] == 400
        # 导入
        file = "./testcases/api/data-operator-hub/impex/export/toolbox_published.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("toolbox", files, {"mode":"upsert"}, Headers)
        assert result[0] == 201

    @allure.title("导入包含从算子转换为工具的工具箱，导入成功")
    def test_import_07(self, Headers):
        # 先删除对应工具箱并确认删除（对应导出为toolbox_ids[30]）
        id = TestImport.toolbox_ids[30]
        result = self.toolbox_client.UpdateToolboxStatus(id, {"status": "offline"}, Headers)
        assert result[0] == 200
        result = self.toolbox_client.DeleteToolbox(id, Headers)
        assert result[0] == 200
        result = self.toolbox_client.GetToolbox(id, Headers)
        assert result[0] == 400
        # 导入
        file = "./testcases/api/data-operator-hub/impex/export/toolbox_op_imported.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("toolbox", files, {"mode":"upsert"}, Headers)
        assert result[0] == 201

    @allure.title("导入自定义mcp，导入成功")
    def test_import_08(self, Headers):
        # 先删除对应mcp并确认删除（对应导出为mcp_ids[0]）
        id = TestImport.mcp_ids[0]
        result = self.mcp_client.DeleteMCP(id, Headers)
        assert result[0] == 200
        result = self.mcp_client.GetMCPDetail(id, Headers)
        assert result[0] == 404
        # 导入
        file = "./testcases/api/data-operator-hub/impex/export/mcp_custom.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("mcp", files, {"mode":"upsert"}, Headers)
        assert result[0] == 201

    @allure.title("导入从工具箱转换的mcp，且mcp下的工具来自不同工具箱，导入成功")
    def test_import_09(self, Headers):
        # 先删除对应mcp并确认删除（对应导出为mcp_ids[1]）
        id = TestImport.mcp_ids[1]
        result = self.mcp_client.DeleteMCP(id, Headers)
        assert result[0] == 200
        result = self.mcp_client.GetMCPDetail(id, Headers)
        assert result[0] == 404
        # 导入
        file = "./testcases/api/data-operator-hub/impex/export/mcp_tool_imported.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("mcp", files, {"mode":"upsert"}, Headers)
        assert result[0] == 201

    @allure.title("导入从工具箱转换的mcp，且工具箱中包含从算子导入的工具，导入成功")
    def test_import_10(self, Headers):
        # 先删除对应mcp并确认删除（对应导出为mcp_ids[2]）
        id = TestImport.mcp_ids[2]
        result = self.mcp_client.DeleteMCP(id, Headers)
        assert result[0] == 200
        result = self.mcp_client.GetMCPDetail(id, Headers)
        assert result[0] == 404
        # 导入
        file = "./testcases/api/data-operator-hub/impex/export/mcp_op_imported.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("mcp", files, {"mode":"upsert"}, Headers)
        assert result[0] == 201

    @allure.title("没有算子/工具箱/MCP的新建权限，导入失败(创建模式/更新模式)")
    def test_import_11(self):
        filenames = [
            "./testcases/api/data-operator-hub/impex/export/operator_unpublish.json",
            "./testcases/api/data-operator-hub/impex/export/toolbox_published.json",
            "./testcases/api/data-operator-hub/impex/export/mcp_custom.json"
        ]
        for mode in ["create", "upsert"]:
            for file in filenames:
                if "operator" in file:
                    type = "operator"
                if "toolbox" in file:
                    type = "toolbox"
                if "mcp" in file:
                    type = "mcp"
                files = {"data": (os.path.basename(file), open(file, "rb"))}
                result = self.impex_client.importation(type, files, {"mode":mode}, TestImport.t2_headers)
                assert result[0] == 403

    @allure.title("没有工具新建权限，导入的MCP存在引用的工具，导入失败(创建模式/更新模式)")
    def test_import_12(self, Headers):
        data = [
            {
                "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": "*", "type": "mcp", "name": "mcp权限"},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            }
        ]
        result = self.perm_client.SetPerm(data, Headers)
        assert "20" in str(result[0])
        file  = "./testcases/api/data-operator-hub/impex/export/mcp_tool_imported.json"
        for mode in ["create", "upsert"]:
            files = {"data": (os.path.basename(file), open(file, "rb"))}
            result = self.impex_client.importation("mcp", files, {"mode":mode}, TestImport.t2_headers)
            assert result[0] == 403

    @allure.title("没有算子新建权限，导入的工具箱中存在引用的算子，导入失败(创建模式/更新模式)")
    def test_import_13(self, Headers):
        data = [
            {
                "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": "*", "type": "tool_box", "name": "工具箱权限"},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            }
        ]
        result = self.perm_client.SetPerm(data, Headers)
        assert "20" in str(result[0])
        file  = "./testcases/api/data-operator-hub/impex/export/toolbox_op_imported.json"
        for mode in ["create", "upsert"]:
            files = {"data": (os.path.basename(file), open(file, "rb"))}
            result = self.impex_client.importation("toolbox", files, {"mode":mode}, TestImport.t2_headers)
            assert result[0] == 403

    @allure.title("mcp已存在，且对已存在的mcp无编辑权限，导入失败(更新模式)")
    def test_import_14(self):
        file = "./testcases/api/data-operator-hub/impex/export/mcp_custom.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("mcp", files, {"mode":"upsert"}, TestImport.t2_headers)
        assert result[0] == 403

    @allure.title("mcp已存在，且对已存在的mcp有编辑权限，导入成功(更新模式)")
    def test_import_15(self, Headers):
        # 授权t2编辑该mcp
        perm = [{
            "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
            "resource": {"id": TestImport.mcp_ids[0], "type": "mcp", "name": "mcp权限"},
            "operation": {"allow": [{"id": "modify"}], "deny": []}
        }]
        result = self.perm_client.SetPerm(perm, Headers)
        assert "20" in str(result[0])
        # upsert导入
        file = "./testcases/api/data-operator-hub/impex/export/mcp_custom.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("mcp", files, {"mode":"upsert"}, TestImport.t2_headers)
        assert result[0] == 201

    @allure.title("工具箱已存在，且对已存在的工具箱无编辑权限，导入失败(更新模式)")
    def test_import_16(self):
        file = "./testcases/api/data-operator-hub/impex/export/toolbox_published.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("toolbox", files, {"mode":"upsert"}, TestImport.t2_headers)
        assert result[0] == 403

    @allure.title("工具箱已存在，且对已存在的工具箱有编辑权限，导入成功(更新模式)")
    def test_import_17(self, Headers):
        # 授权t2编辑该工具箱
        perm = [{
            "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
            "resource": {"id": TestImport.toolbox_ids[0], "type": "tool_box", "name": "工具箱权限"},
            "operation": {"allow": [{"id": "modify"}], "deny": []}
        }]
        result = self.perm_client.SetPerm(perm, Headers)
        assert "20" in str(result[0])
        # upsert导入
        file = "./testcases/api/data-operator-hub/impex/export/toolbox_published.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("toolbox", files, {"mode":"upsert"}, TestImport.t2_headers)
        assert result[0] == 201

    @allure.title("算子已存在，且对已存在的算子无编辑权限，导入失败(更新模式)")
    def test_import_18(self):
        file = "./testcases/api/data-operator-hub/impex/export/operator_unpublish.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("operator", files, {"mode":"upsert"}, TestImport.t2_headers)
        assert result[0] == 403

    @allure.title("算子已存在，有算子新建权限且对已存在的算子有编辑权限，导入成功(更新模式)")
    def test_import_19(self, Headers):
        # 授权t2编辑该算子
        perm = [{
            "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
            "resource": {"id": "*", "type": "operator", "name": "算子权限"},
            "operation": {"allow": [{"id": "create"}], "deny": []}
        },
        {
            "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
            "resource": {"id": TestImport.operator_ids[0], "type": "operator", "name": "算子权限"},
            "operation": {"allow": [{"id": "modify"}], "deny": []}
        }]
        result = self.perm_client.SetPerm(perm, Headers)
        assert "20" in str(result[0])
        # upsert导入
        file = "./testcases/api/data-operator-hub/impex/export/operator_unpublish.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("operator", files, {"mode":"upsert"}, TestImport.t2_headers)
        assert result[0] == 201

    @allure.title("已存在从工具导入的mcp，对该mcp有编辑权限，但对工具箱无编辑权限，导入失败(更新模式)")
    def test_import_20(self, Headers):
        # 授权t2仅编辑mcp
        perm = [{
            "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
            "resource": {"id": TestImport.mcp_ids[1], "type": "mcp", "name": "mcp权限"},
            "operation": {"allow": [{"id": "modify"}], "deny": []}
        }]
        result = self.perm_client.SetPerm(perm, Headers)
        assert "20" in str(result[0])
        # upsert导入该mcp（工具来自不同工具箱）
        file = "./testcases/api/data-operator-hub/impex/export/mcp_tool_imported.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("mcp", files, {"mode":"upsert"}, TestImport.t2_headers)
        assert result[0] == 403

    @allure.title("已存在从工具导入的mcp，对该mcp及工具箱有编辑权限，但对由算子导入的工具无编辑权限，导入失败(更新模式)")
    def test_import_21(self, Headers):
        # 授权t2编辑mcp与所有工具箱
        perm = [
            {
                "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": TestImport.mcp_ids[2], "type": "mcp", "name": "mcp权限"},
                "operation": {"allow": [{"id": "modify"}], "deny": []}
            },
            {
                "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": TestImport.toolbox_ids[30], "type": "tool_box", "name": "工具箱权限"},
                "operation": {"allow": [{"id": "modify"}], "deny": []}
            }
        ]
        result = self.perm_client.SetPerm(perm, Headers)
        assert "20" in str(result[0])
        # upsert导入（该mcp引用的工具来自算子导入，未授权算子编辑）
        file = "./testcases/api/data-operator-hub/impex/export/mcp_op_imported.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("mcp", files, {"mode":"upsert"}, TestImport.t2_headers)
        assert result[0] == 403

    @allure.title("已存在从工具导入的mcp，对该mcp、工具箱及算子导入的工具有编辑权限，导入成功(更新模式)")
    def test_import_22(self, Headers):
        # 授权t2编辑mcp、所有工具箱与所有算子
        perm = [
            {
                "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": TestImport.mcp_ids[2], "type": "mcp", "name": "mcp权限"},
                "operation": {"allow": [{"id": "modify"}], "deny": []}
            },
            {
                "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": TestImport.toolbox_ids[30], "type": "tool_box", "name": "工具箱权限"},
                "operation": {"allow": [{"id": "modify"}], "deny": []}
            },
            {
                "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": TestImport.operator_id, "type": "operator", "name": "算子权限"},
                "operation": {"allow": [{"id": "modify"}], "deny": []}
            }
        ]
        result = self.perm_client.SetPerm(perm, Headers)
        assert "20" in str(result[0])
        # upsert导入
        file = "./testcases/api/data-operator-hub/impex/export/mcp_op_imported.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("mcp", files, {"mode":"upsert"}, TestImport.t2_headers)
        assert result[0] == 201

    @allure.title("导入文件不合法，导入失败")
    def test_import_23(self, Headers):
        file = "./resource/openapi/compliant/tool.json"
        for mode in ["create", "upsert"]:
            for type in ["operator", "toolbox", "mcp"]:
                files = {"data": (os.path.basename(file), open(file, "rb"))}
                result = self.impex_client.importation(type, files, {"mode":mode}, Headers)   # 导入算子
                assert result[0] == 400

    @allure.title("导入类型与文件不匹配，导入失败")
    def test_import_24(self, Headers):
        file = "./testcases/api/data-operator-hub/impex/export/operator_published.json"
        for type in ["toolbox", "mcp"]:
            files = {"data": (os.path.basename(file), open(file, "rb"))}
            result = self.impex_client.importation(type, files, {"mode":"upsert"}, Headers)
            assert result[0] == 400
        file = "./testcases/api/data-operator-hub/impex/export/toolbox_published.json"
        for type in ["operator", "mcp"]:
            files = {"data": (os.path.basename(file), open(file, "rb"))}
            result = self.impex_client.importation(type, files, {"mode":"upsert"}, Headers)
            assert result[0] == 400
        file = "./testcases/api/data-operator-hub/impex/export/mcp_custom.json"
        for type in ["toolbox", "operator"]:
            files = {"data": (os.path.basename(file), open(file, "rb"))}
            result = self.impex_client.importation(type, files, {"mode":"upsert"}, Headers)
            assert result[0] == 400
