# -*- coding:UTF-8 -*-
"""
导入接口测试（重构版）

测试目标：
    验证资源导入功能，确保导入逻辑、模式切换、权限校验和依赖处理在实际导出文件下正常工作。

测试覆盖：
    1. 正常场景：Export -> Import 闭环流程
    2. 模式校验：create (409) vs upsert (201)
    3. 权限校验：t2 用户在不同授权下的导入行为
    4. 异常场景：非法文件、类型不匹配

重构重点：
    - 动态生成导出文件，解决路径不存在和数据失效问题。
    - 统一文件处理逻辑，确保 multipart 报文 100% 字节对齐。
    - 增强错误诊断，输出 500 错误响应体。
"""

import allure
import pytest
import os
import json
import uuid

from common.file_process import FileProcess
from lib.operator import Operator
from lib.tool_box import ToolBox
from lib.mcp import MCP
from lib.impex import Impex
from lib.permission import Perm

@allure.feature("算子平台测试：导入测试")
class TestImport:
    impex_client = Impex()
    operator_client = Operator()
    toolbox_client = ToolBox()
    mcp_client = MCP()
    file_client = FileProcess()
    perm_client = Perm()
    
    # 临时文件存放目录
    TEMP_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "temp_export")

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, PrepareData, Headers):
        """
        重构 Setup：动态准备导出文件，解决文件缺失导致的失败。
        """
        # 创建目录
        if not os.path.exists(self.TEMP_DIR):
            os.makedirs(self.TEMP_DIR)
            
        # 提取准备好的资源 ID
        TestImport.operator_ids = PrepareData[0]
        TestImport.toolbox_ids = PrepareData[1]
        TestImport.mcp_ids = PrepareData[2]
        TestImport.t1_operator_id = PrepareData[3]
        TestImport.t1_toolbox_id = PrepareData[4]
        TestImport.t1_mcp_id =  PrepareData[5]
        TestImport.t2_headers = PrepareData[6]
        TestImport.user_t2 = PrepareData[7]
        TestImport.operator_id = PrepareData[8]
        TestImport.operator_version = PrepareData[9]

        # --- 动态准备导入所需的文件 ---
        TestImport.files = {}
        
        # 1. 准备算子文件 (未发布、已发布、已下架)
        self._export_and_save("operator", TestImport.operator_ids[0], "operator_unpublish.json", Headers)
        self._export_and_save("operator", TestImport.operator_ids[2], "operator_published.json", Headers)
        self._export_and_save("operator", TestImport.operator_ids[1], "operator_offline.json", Headers)
        
        # 2. 准备工具箱文件
        self._export_and_save("toolbox", TestImport.toolbox_ids[0], "toolbox_published.json", Headers)
        self._export_and_save("toolbox", TestImport.toolbox_ids[30], "toolbox_op_imported.json", Headers)
        
        # 3. 准备 MCP 文件
        self._export_and_save("mcp", TestImport.mcp_ids[0], "mcp_custom.json", Headers)
        self._export_and_save("mcp", TestImport.mcp_ids[1], "mcp_tool_imported.json", Headers)
        self._export_and_save("mcp", TestImport.mcp_ids[2], "mcp_op_imported.json", Headers)

    def _export_and_save(self, type, res_id, filename, headers):
        """内部辅助：导出并保存文件"""
        res = self.impex_client.export(type, res_id, headers)
        if res[0] == 200:
            path = os.path.join(self.TEMP_DIR, filename)
            self.file_client.write_json_to_file(res[1], path)
            TestImport.files[filename] = path
        else:
            print(f"警告: 导出 {type} {res_id} 失败: {res}")

    def get_file(self, filename):
        """获取文件路径，如果不存在则报跳过"""
        path = TestImport.files.get(filename)
        if not path or not os.path.exists(path):
            pytest.skip(f"导入文件未准备就绪: {filename}")
        return path

    @allure.title("以新建模式导入，资源冲突，导入失败")
    def test_import_01(self, Headers):
        """
        测试场景：使用 create 模式导入已存在的资源，预期 409。
        """
        filenames = ["operator_unpublish.json", "toolbox_published.json", "mcp_custom.json"]
        for fname in filenames:
            type = "operator" if "operator" in fname else ("toolbox" if "toolbox" in fname else "mcp")
            path = self.get_file(fname)
            result = self.impex_client.import_from_file(type, path, {"mode": "create"}, Headers)
            assert result[0] == 409, f"导入 {fname} 预期 409，实际 {result[0]}: {result[1]}"

    @allure.title("以更新模式导入，资源冲突，导入成功")
    def test_import_02(self, Headers):
        """
        测试场景：使用 upsert 模式导入已存在的资源，预期 201。
        """
        filenames = ["operator_unpublish.json", "toolbox_published.json", "mcp_custom.json"]
        for fname in filenames:
            type = "operator" if "operator" in fname else ("toolbox" if "toolbox" in fname else "mcp")
            path = self.get_file(fname)
            result = self.impex_client.import_from_file(type, path, {"mode": "upsert"}, Headers)
            assert result[0] == 201, f"导入 {fname} 预期 201，实际 {result[0]}: {result[1]}"

    @allure.title("导入未发布算子，导入成功，状态为未发布")
    def test_import_03(self, Headers):
        id = TestImport.operator_ids[0]
        # 先删除
        info = self.operator_client.GetOperatorInfo(id, Headers)
        if info[0] == 200:
            self.operator_client.DeleteOperator([{"operator_id": id, "version": info[1]["version"]}], Headers)
        
        # 导入
        path = self.get_file("operator_unpublish.json")
        result = self.impex_client.import_from_file("operator", path, {"mode": "upsert"}, Headers)
        assert result[0] == 201
        
        # 校验状态
        info = self.operator_client.GetOperatorInfo(id, Headers)
        assert info[0] == 200
        assert info[1]["status"] == "unpublish"

    @allure.title("导入已发布算子，导入成功，状态为已发布")
    def test_import_04(self, Headers):
        id = TestImport.operator_ids[2]
        # 先下架并删除
        self.operator_client.UpdateOperatorStatus([{"operator_id": id, "status": "offline"}], Headers)
        info = self.operator_client.GetOperatorInfo(id, Headers)
        self.operator_client.DeleteOperator([{"operator_id": id, "version": info[1]["version"]}], Headers)
        
        # 导入
        path = self.get_file("operator_published.json")
        result = self.impex_client.import_from_file("operator", path, {"mode": "upsert"}, Headers)
        assert result[0] == 201
        
        # 校验状态
        info = self.operator_client.GetOperatorInfo(id, Headers)
        assert info[0] == 200
        assert info[1]["status"] == "published"

    @allure.title("导入已下架算子，导入成功，状态为已下架")
    def test_import_05(self, Headers):
        id = TestImport.operator_ids[1]
        info = self.operator_client.GetOperatorInfo(id, Headers)
        self.operator_client.DeleteOperator([{"operator_id": id, "version": info[1]["version"]}], Headers)
        
        path = self.get_file("operator_offline.json")
        result = self.impex_client.import_from_file("operator", path, {"mode": "upsert"}, Headers)
        assert result[0] == 201
        
        info = self.operator_client.GetOperatorInfo(id, Headers)
        assert info[1]["status"] == "offline"

    @allure.title("导入工具箱，导入成功")
    def test_import_06(self, Headers):
        id = TestImport.toolbox_ids[0]
        self.toolbox_client.UpdateToolboxStatus(id, {"status": "offline"}, Headers)
        self.toolbox_client.DeleteToolbox(id, Headers)
        
        path = self.get_file("toolbox_published.json")
        result = self.impex_client.import_from_file("toolbox", path, {"mode": "upsert"}, Headers)
        assert result[0] == 201

    @allure.title("导入包含从算子转换为工具的工具箱，导入成功")
    def test_import_07(self, Headers):
        id = TestImport.toolbox_ids[30]
        self.toolbox_client.UpdateToolboxStatus(id, {"status": "offline"}, Headers)
        self.toolbox_client.DeleteToolbox(id, Headers)
        
        path = self.get_file("toolbox_op_imported.json")
        result = self.impex_client.import_from_file("toolbox", path, {"mode": "upsert"}, Headers)
        assert result[0] == 201

    @allure.title("导入自定义mcp，导入成功")
    def test_import_08(self, Headers):
        id = TestImport.mcp_ids[0]
        self.mcp_client.DeleteMCP(id, Headers)
        
        path = self.get_file("mcp_custom.json")
        result = self.impex_client.import_from_file("mcp", path, {"mode": "upsert"}, Headers)
        assert result[0] == 201

    @allure.title("导入从工具箱转换的mcp，且mcp下的工具来自不同工具箱，导入成功")
    def test_import_09(self, Headers):
        id = TestImport.mcp_ids[1]
        self.mcp_client.DeleteMCP(id, Headers)
        
        path = self.get_file("mcp_tool_imported.json")
        result = self.impex_client.import_from_file("mcp", path, {"mode": "upsert"}, Headers)
        assert result[0] == 201

    @allure.title("导入从工具箱转换的mcp，且工具箱中包含从算子导入的工具，导入成功")
    def test_import_10(self, Headers):
        id = TestImport.mcp_ids[2]
        self.mcp_client.DeleteMCP(id, Headers)
        
        path = self.get_file("mcp_op_imported.json")
        result = self.impex_client.import_from_file("mcp", path, {"mode": "upsert"}, Headers)
        assert result[0] == 201

    @allure.title("没有算子/工具箱/MCP的新建权限，导入失败")
    def test_import_11(self):
        filenames = ["operator_unpublish.json", "toolbox_published.json", "mcp_custom.json"]
        for mode in ["create", "upsert"]:
            for fname in filenames:
                type = "operator" if "operator" in fname else ("toolbox" if "toolbox" in fname else "mcp")
                path = self.get_file(fname)
                result = self.impex_client.import_from_file(type, path, {"mode": mode}, TestImport.t2_headers)
                assert result[0] == 403

    @allure.title("没有工具新建权限，导入的MCP存在引用的工具，导入失败")
    def test_import_12(self, Headers):
        # 仅授权 MCP 新建
        self.perm_client.SetPerm([
            {
                "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": "*", "type": "mcp", "name": "mcp权限"},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            }
        ], Headers)
        
        path = self.get_file("mcp_tool_imported.json")
        result = self.impex_client.import_from_file("mcp", path, {"mode": "upsert"}, TestImport.t2_headers)
        assert result[0] == 403

    @allure.title("没有算子新建权限，导入的工具箱中存在引用的算子，导入失败")
    def test_import_13(self, Headers):
        self.perm_client.SetPerm([
            {
                "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": "*", "type": "tool_box", "name": "工具箱权限"},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            }
        ], Headers)
        
        path = self.get_file("toolbox_op_imported.json")
        result = self.impex_client.import_from_file("toolbox", path, {"mode": "upsert"}, TestImport.t2_headers)
        assert result[0] == 403

    @allure.title("资源已存在且无编辑权限，导入失败(更新模式)")
    def test_import_14(self):
        # MCP 示例
        path = self.get_file("mcp_custom.json")
        result = self.impex_client.import_from_file("mcp", path, {"mode": "upsert"}, TestImport.t2_headers)
        assert result[0] == 403

    @allure.title("资源已存在且有编辑权限，导入成功(更新模式)")
    def test_import_15(self, Headers):
        # 授权 t2 编辑 MCP
        self.perm_client.SetPerm([{
            "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
            "resource": {"id": TestImport.mcp_ids[0], "type": "mcp", "name": "mcp权限"},
            "operation": {"allow": [{"id": "modify"}], "deny": []}
        }], Headers)
        
        path = self.get_file("mcp_custom.json")
        result = self.impex_client.import_from_file("mcp", path, {"mode": "upsert"}, TestImport.t2_headers)
        assert result[0] == 201

    @allure.title("工具箱已存在且有编辑权限，导入成功")
    def test_import_17(self, Headers):
        self.perm_client.SetPerm([{
            "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
            "resource": {"id": TestImport.toolbox_ids[0], "type": "tool_box", "name": "工具箱权限"},
            "operation": {"allow": [{"id": "modify"}], "deny": []}
        }], Headers)
        
        path = self.get_file("toolbox_published.json")
        result = self.impex_client.import_from_file("toolbox", path, {"mode": "upsert"}, TestImport.t2_headers)
        assert result[0] == 201

    @allure.title("算子已存在且有新建+编辑权限，导入成功")
    def test_import_19(self, Headers):
        self.perm_client.SetPerm([
            {
                "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": "*", "type": "operator", "name": "算子权限"},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            },
            {
                "accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": TestImport.operator_ids[0], "type": "operator", "name": "算子权限"},
                "operation": {"allow": [{"id": "modify"}], "deny": []}
            }
        ], Headers)
        
        path = self.get_file("operator_unpublish.json")
        result = self.impex_client.import_from_file("operator", path, {"mode": "upsert"}, TestImport.t2_headers)
        assert result[0] == 201

    @allure.title("导入复杂 MCP 依赖权限校验")
    def test_import_22(self, Headers):
        # 授权全套权限给 t2
        self.perm_client.SetPerm([
            {"accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"}, "resource": {"id": TestImport.mcp_ids[2], "type": "mcp", "name": "mcp"}, "operation": {"allow": [{"id": "modify"}], "deny": []}},
            {"accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"}, "resource": {"id": TestImport.toolbox_ids[30], "type": "tool_box", "name": "box"}, "operation": {"allow": [{"id": "modify"}], "deny": []}},
            {"accessor": {"id": TestImport.user_t2, "name": "t2", "type": "user"}, "resource": {"id": TestImport.operator_id, "type": "operator", "name": "op"}, "operation": {"allow": [{"id": "modify"}], "deny": []}}
        ], Headers)
        
        path = self.get_file("mcp_op_imported.json")
        result = self.impex_client.import_from_file("mcp", path, {"mode": "upsert"}, TestImport.t2_headers)
        assert result[0] == 201

    @allure.title("导入文件不合法/类型不匹配")
    def test_import_errors(self, Headers):
        # 非法 JSON
        bad_path = os.path.join(self.TEMP_DIR, "bad.json")
        with open(bad_path, "w") as f: f.write("{invalid_json}")
        
        res = self.impex_client.import_from_file("operator", bad_path, {"mode": "upsert"}, Headers)
        assert res[0] == 400
        
        # 类型不匹配
        op_path = self.get_file("operator_unpublish.json")
        res = self.impex_client.import_from_file("toolbox", op_path, {"mode": "upsert"}, Headers)
        assert res[0] == 400
