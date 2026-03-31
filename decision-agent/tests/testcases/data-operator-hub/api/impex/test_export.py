# -*- coding:UTF-8 -*-

import allure
import pytest
import uuid

from common.file_process import FileProcess

from lib.operator import Operator
from lib.tool_box import ToolBox
from lib.impex import Impex
from lib.permission import Perm


@allure.feature("算子平台测试：导入导出测试")
class TestExport:
    impex_client = Impex()
    operator_client = Operator()
    toolbox_client = ToolBox()
    file_client = FileProcess()
    perm_client = Perm()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, PrepareData):
        TestExport.operator_ids = PrepareData[0] # 前4个为基础算子，第1个为未发布状态，第2为下架状态，第3、4个为已发布状态；第5个为内置算子
        TestExport.toolbox_ids = PrepareData[1] # 前30个工具箱中的所有工具均为本地导入；第31个工具箱包含一个从算子导入的工具；第32个为内置工具箱
        TestExport.mcp_ids = PrepareData[2] # 第1个为自定义mcp；第2个为工具转换为的mcp，每个mcp存在30个工具，分别来自不同的工具箱；第3个为工具转换为的mcp，mcp下的工具为从算子导入的工具；第4个为内置mcp
        TestExport.t1_operator_id = PrepareData[3]
        TestExport.t1_toolbox_id = PrepareData[4]
        TestExport.t1_mcp_id =  PrepareData[5]
        TestExport.t2_headers = PrepareData[6]
        TestExport.user_t2 = PrepareData[7]
        TestExport.operator_id = PrepareData[8] # 转换成工具的算子id
        TestExport.operator_version = PrepareData[9] # 转换成工具的算子版本

    @allure.title("导出未发布算子，导出成功")
    def test_export_01(self, Headers):
        id = TestExport.operator_ids[0]
        result = self.impex_client.export("operator", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/operator_unpublish.json")

    @allure.title("导出已发布算子，导出成功")
    def test_export_02(self, Headers):
        id = TestExport.operator_ids[2]
        result = self.impex_client.export("operator", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/operator_published.json")

    @allure.title("导出已下架算子，导出成功")
    def test_export_03(self, Headers):
        id = TestExport.operator_ids[1]
        result = self.impex_client.export("operator", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/operator_offline.json")

    @allure.title("导出工具箱，导出成功")
    def test_export_04(self, Headers):
        id = TestExport.toolbox_ids[0]
        result = self.impex_client.export("toolbox", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/toolbox_published.json")

    @allure.title("导出包含从算子转换为工具的工具箱，导出成功")
    def test_export_05(self, Headers):
        id = TestExport.toolbox_ids[30]
        result = self.impex_client.export("toolbox", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/toolbox_op_imported.json")

    @allure.title("导出自定义mcp，导出成功")
    def test_export_06(self, Headers):
        id = TestExport.mcp_ids[0]
        result = self.impex_client.export("mcp", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/mcp_custom.json")

    @allure.title("导出从工具箱转换的mcp，且mcp下的工具来自不同工具箱，导出成功")
    def test_export_07(self, Headers):
        id = TestExport.mcp_ids[1]
        result = self.impex_client.export("mcp", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/mcp_tool_imported.json")

    @allure.title("导出从工具箱转换的mcp，且工具箱中包含从算子导入的工具，导出成功")
    def test_export_08(self, Headers):
        id = TestExport.mcp_ids[2]
        result = self.impex_client.export("mcp", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/mcp_op_imported.json")

    @allure.title("导出mcp，mcp中的部分工具不存在，导出失败")
    def test_export_09(self, Headers):
        # 删除工具箱
        box_id = TestExport.toolbox_ids[1]
        result = self.toolbox_client.UpdateToolboxStatus(box_id, {"status": "offline"}, Headers) # 下架工具箱
        assert result[0] == 200
        result = self.toolbox_client.DeleteToolbox(box_id, Headers)
        assert result[0] == 200
        #导出mcp
        id = TestExport.mcp_ids[1]
        result = self.impex_client.export("mcp", id, Headers)
        assert result[0] == 404

    @allure.title("导出内置组件，导出失败")
    def test_export_10(self, Headers):
        result = self.impex_client.export("operator", TestExport.operator_ids[4], Headers)
        assert result[0] == 403
        result = self.impex_client.export("toolbox", TestExport.toolbox_ids[31], Headers)
        assert result[0] == 403
        result = self.impex_client.export("mcp", TestExport.mcp_ids[3], Headers)
        assert result[0] == 403

    @allure.title("导出的组件id与导出类型不匹配，导出失败")
    def test_export_11(self, Headers):
        id = TestExport.toolbox_ids[0]
        result = self.impex_client.export("mcp", id, Headers)
        assert result[0] == 404

    @allure.title("导出的组件不存在，导出失败") 
    def test_export_12(self, Headers):
        id = str(uuid.uuid4())
        result = self.impex_client.export("operator", id, Headers)
        assert result[0] == 404

    @allure.title("对算子/工具箱/MCP无查看权限，导出失败")
    def test_export_13(self):
        result = self.impex_client.export("operator", TestExport.t1_operator_id, TestExport.t2_headers)
        assert result[0] == 403
        result = self.impex_client.export("toolbox", TestExport.t1_toolbox_id, TestExport.t2_headers)
        assert result[0] == 403
        result = self.impex_client.export("mcp", TestExport.t1_mcp_id, TestExport.t2_headers)
        assert result[0] == 403

    @allure.title("对工具箱引用的算子无查看权限，导出失败")
    def test_export_14(self, Headers):
        # 授权工具箱的查看权限给t2，无引用算子的查看权限
        data = [
            {
                "accessor": {"id": TestExport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": TestExport.toolbox_ids[30], "type": "tool_box", "name": "工具箱权限"},
                "operation": {"allow": [{"id": "view"}], "deny": []}
            }
        ]
        result = self.perm_client.SetPerm(data, Headers)
        assert "20" in str(result[0])
        # 导出
        result = self.impex_client.export("toolbox", TestExport.toolbox_ids[30], TestExport.t2_headers)
        assert result[0] == 403

    @allure.title("对MCP引用的工具无查看权限，导出失败")
    def test_export_15(self, Headers):
        # 授权工具箱的查看权限给t2，无引用工具的查看权限
        data = [
            {
                "accessor": {"id": TestExport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": TestExport.mcp_ids[1], "type": "mcp", "name": "mcp权限"},
                "operation": {"allow": [{"id": "view"}], "deny": []}
            }
        ]
        result = self.perm_client.SetPerm(data, Headers)
        assert "20" in str(result[0])
        # 导出
        result = self.impex_client.export("mcp", TestExport.mcp_ids[1], TestExport.t2_headers)
        assert result[0] == 403

    @allure.title("对MCP引用的算子转换成的工具有查看权限，但对算子本身无查看权限，导出失败")
    def test_export_16(self, Headers):
        # 授权工具箱的查看权限给t2，无引用工具的查看权限
        data = [
            {
                "accessor": {"id": TestExport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": TestExport.mcp_ids[2], "type": "mcp", "name": "mcp权限"},
                "operation": {"allow": [{"id": "view"}], "deny": []}
            },
            {
                "accessor": {"id": TestExport.user_t2, "name": "t2", "type": "user"},
                "resource": {"id": TestExport.toolbox_ids[30], "type": "tool_box", "name": "工具箱权限"},
                "operation": {"allow": [{"id": "view"}], "deny": []}
            }
        ]
        result = self.perm_client.SetPerm(data, Headers)
        assert "20" in str(result[0])
        # 导出
        result = self.impex_client.export("mcp", TestExport.mcp_ids[2], TestExport.t2_headers)
        assert result[0] == 403

    @allure.title("导出工具箱，工具箱中引用的算子不存在，导出失败")
    def test_export_17(self, Headers):
        data = [
            {
                "operator_id": TestExport.operator_id,
                "status": "offline"
            }
        ]
        result = self.operator_client.UpdateOperatorStatus(data, Headers)  # 下架算子
        assert result[0] == 200  
        del_data = [
            {
                "operator_id": TestExport.operator_id,
                "version": TestExport.operator_version
            }
        ]
        result = self.operator_client.DeleteOperator(del_data, Headers) # 删除算子
        assert result[0] == 200
        #导出工具箱
        id = TestExport.toolbox_ids[30]
        result = self.impex_client.export("toolbox", id, Headers)
        assert result[0] == 404
    
