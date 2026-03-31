# -*- coding:UTF-8 -*-
"""
导出接口测试

测试目标：
    验证资源导出功能，包括算子、工具箱、MCP的导出，不同状态的资源导出，
    权限校验，资源依赖关系处理，异常场景处理等。

测试覆盖：
    1. 正常场景：导出未发布算子，导出成功
    2. 正常场景：导出已发布算子，导出成功
    3. 正常场景：导出已下架算子，导出成功
    4. 正常场景：导出工具箱，导出成功
    5. 正常场景：导出包含从算子转换为工具的工具箱，导出成功
    6. 正常场景：导出自定义mcp，导出成功
    7. 正常场景：导出从工具箱转换的mcp，且mcp下的工具来自不同工具箱，导出成功
    8. 正常场景：导出从工具箱转换的mcp，且工具箱中包含从算子导入的工具，导出成功
    9. 异常场景：导出mcp，mcp中的部分工具不存在，导出失败
    10. 异常场景：导出内置组件，导出失败
    11. 异常场景：导出的组件id与导出类型不匹配，导出失败
    12. 异常场景：导出的组件不存在，导出失败
    13. 异常场景：对算子/工具箱/MCP无查看权限，导出失败
    14. 异常场景：对工具箱引用的算子无查看权限，导出失败
    15. 异常场景：对MCP引用的工具无查看权限，导出失败
    16. 异常场景：对MCP引用的算子转换成的工具有查看权限，但对算子本身无查看权限，导出失败
    17. 异常场景：导出工具箱，工具箱中引用的算子不存在，导出失败

说明：
    导出功能：
    - 导出算子、工具箱、MCP等资源为JSON格式文件
    - 导出的文件可以用于后续的导入操作
    - 导出时会包含资源的所有信息，包括状态、版本、依赖关系等
    
    资源状态：
    - 可以导出不同状态的资源（unpublish/published/offline）
    - 导出时会保持资源的状态信息
    
    权限要求：
    - 导出资源需要对该资源有查看权限
    - 如果资源引用了其他资源（如工具箱引用算子，MCP引用工具），需要对引用的资源也有查看权限
    - 内置组件（builtin）不允许导出
    
    资源依赖关系：
    - 工具箱可以包含从算子转换而来的工具
    - MCP可以包含来自不同工具箱的工具
    - MCP可以包含从算子转换而来的工具（通过工具箱）
    - 导出时需要确保所有依赖的资源都存在且有查看权限
    
    异常场景：
    - 资源不存在：返回404
    - 资源类型不匹配：返回404
    - 无查看权限：返回403
    - 内置组件：返回403
    - 依赖资源不存在：返回404
    - 依赖资源无查看权限：返回403
"""

import allure
import pytest
import uuid

from common.file_process import FileProcess

from lib.operator import Operator
from lib.tool_box import ToolBox
from lib.impex import Impex
from lib.permission import Perm


@allure.feature("算子平台测试：导出测试")
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
        """
        测试用例1：正常场景 - 导出未发布算子
        
        测试场景：
            - 导出一个未发布状态的算子
            - 验证导出成功并保存为JSON文件
        
        验证点：
            - 导出接口返回200状态码
            - 导出数据成功保存到文件
        
        说明：
            可以导出未发布状态的算子。
            导出的文件可以用于后续的导入操作。
        """
        id = TestExport.operator_ids[0]
        result = self.impex_client.export("operator", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/operator_unpublish.json")

    @allure.title("导出已发布算子，导出成功")
    def test_export_02(self, Headers):
        """
        测试用例2：正常场景 - 导出已发布算子
        
        测试场景：
            - 导出一个已发布状态的算子
            - 验证导出成功并保存为JSON文件
        
        验证点：
            - 导出接口返回200状态码
            - 导出数据成功保存到文件
        
        说明：
            可以导出已发布状态的算子。
            导出的文件会包含算子的发布状态信息。
        """
        id = TestExport.operator_ids[2]
        result = self.impex_client.export("operator", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/operator_published.json")

    @allure.title("导出已下架算子，导出成功")
    def test_export_03(self, Headers):
        """
        测试用例3：正常场景 - 导出已下架算子
        
        测试场景：
            - 导出一个已下架状态的算子
            - 验证导出成功并保存为JSON文件
        
        验证点：
            - 导出接口返回200状态码
            - 导出数据成功保存到文件
        
        说明：
            可以导出已下架状态的算子。
            导出的文件会包含算子的下架状态信息。
        """
        id = TestExport.operator_ids[1]
        result = self.impex_client.export("operator", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/operator_offline.json")

    @allure.title("导出工具箱，导出成功")
    def test_export_04(self, Headers):
        """
        测试用例4：正常场景 - 导出工具箱
        
        测试场景：
            - 导出一个工具箱
            - 验证导出成功并保存为JSON文件
        
        验证点：
            - 导出接口返回200状态码
            - 导出数据成功保存到文件
        
        说明：
            可以导出工具箱，导出时会包含工具箱中的所有工具。
            导出的文件可以用于后续的导入操作。
        """
        id = TestExport.toolbox_ids[0]
        result = self.impex_client.export("toolbox", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/toolbox_published.json")

    @allure.title("导出包含从算子转换为工具的工具箱，导出成功")
    def test_export_05(self, Headers):
        """
        测试用例5：正常场景 - 导出包含从算子转换为工具的工具箱
        
        测试场景：
            - 导出一个包含从算子转换为工具的工具箱
            - 验证导出成功并保存为JSON文件
        
        验证点：
            - 导出接口返回200状态码
            - 导出数据成功保存到文件
        
        说明：
            可以导出包含从算子转换而来的工具的工具箱。
            导出时会包含工具的来源信息（算子ID和版本）。
        """
        id = TestExport.toolbox_ids[30]
        result = self.impex_client.export("toolbox", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/toolbox_op_imported.json")

    @allure.title("导出自定义mcp，导出成功")
    def test_export_06(self, Headers):
        """
        测试用例6：正常场景 - 导出自定义MCP
        
        测试场景：
            - 导出一个自定义MCP
            - 验证导出成功并保存为JSON文件
        
        验证点：
            - 导出接口返回200状态码
            - 导出数据成功保存到文件
        
        说明：
            可以导出自定义MCP（用户手动创建的MCP）。
            导出的文件可以用于后续的导入操作。
        """
        id = TestExport.mcp_ids[0]
        result = self.impex_client.export("mcp", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/mcp_custom.json")

    @allure.title("导出从工具箱转换的mcp，且mcp下的工具来自不同工具箱，导出成功")
    def test_export_07(self, Headers):
        """
        测试用例7：正常场景 - 导出从工具箱转换的MCP（工具来自不同工具箱）
        
        测试场景：
            - 导出一个从工具箱转换的MCP，该MCP下的工具来自不同工具箱
            - 验证导出成功并保存为JSON文件
        
        验证点：
            - 导出接口返回200状态码
            - 导出数据成功保存到文件
        
        说明：
            可以导出从工具箱转换的MCP。
            导出时会包含工具的来源信息（工具箱ID）。
        """
        id = TestExport.mcp_ids[1]
        result = self.impex_client.export("mcp", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/mcp_tool_imported.json")

    @allure.title("导出从工具箱转换的mcp，且工具箱中包含从算子导入的工具，导出成功")
    def test_export_08(self, Headers):
        """
        测试用例8：正常场景 - 导出从工具箱转换的MCP（工具箱包含从算子导入的工具）
        
        测试场景：
            - 导出一个从工具箱转换的MCP，该工具箱包含从算子导入的工具
            - 验证导出成功并保存为JSON文件
        
        验证点：
            - 导出接口返回200状态码
            - 导出数据成功保存到文件
        
        说明：
            可以导出包含复杂依赖关系的MCP。
            导出时会包含完整的依赖链信息（MCP -> 工具箱 -> 算子）。
        """
        id = TestExport.mcp_ids[2]
        result = self.impex_client.export("mcp", id, Headers)
        assert result[0] == 200
        export_data = result[1]
        self.file_client.write_json_to_file(export_data, "./testcases/api/data-operator-hub/impex/export/mcp_op_imported.json")

    @allure.title("导出mcp，mcp中的部分工具不存在，导出失败")
    def test_export_09(self, Headers):
        """
        测试用例9：异常场景 - MCP引用的工具不存在导致导出失败
        
        测试场景：
            1. 删除MCP引用的一个工具箱
            2. 尝试导出该MCP
            3. 验证导出失败
        
        验证点：
            - 导出接口返回404状态码（Not Found）
        
        说明：
            如果MCP引用的工具不存在，导出会失败。
            导出时需要确保所有依赖的资源都存在。
        """
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
        """
        测试用例10：异常场景 - 导出内置组件失败
        
        测试场景：
            - 尝试导出内置算子、内置工具箱、内置MCP
            - 验证导出失败
        
        验证点：
            - 导出接口返回403状态码（Forbidden）
        
        说明：
            内置组件（builtin）不允许导出。
            这是系统保护机制，防止误操作影响系统稳定性。
        """
        result = self.impex_client.export("operator", TestExport.operator_ids[4], Headers)
        assert result[0] == 403
        result = self.impex_client.export("toolbox", TestExport.toolbox_ids[31], Headers)
        assert result[0] == 403
        result = self.impex_client.export("mcp", TestExport.mcp_ids[3], Headers)
        assert result[0] == 403

    @allure.title("导出的组件id与导出类型不匹配，导出失败")
    def test_export_11(self, Headers):
        """
        测试用例11：异常场景 - 组件ID与导出类型不匹配
        
        测试场景：
            - 使用工具箱ID尝试导出为MCP类型
            - 验证导出失败
        
        验证点：
            - 导出接口返回404状态码（Not Found）
        
        说明：
            导出类型必须与组件ID对应的资源类型匹配。
            使用不匹配的类型会导致找不到资源，返回404错误。
        """
        id = TestExport.toolbox_ids[0]
        result = self.impex_client.export("mcp", id, Headers)
        assert result[0] == 404

    @allure.title("导出的组件不存在，导出失败") 
    def test_export_12(self, Headers):
        """
        测试用例12：异常场景 - 导出的组件不存在
        
        测试场景：
            - 使用不存在的组件ID尝试导出
            - 验证导出失败
        
        验证点：
            - 导出接口返回404状态码（Not Found）
        
        说明：
            如果组件不存在，导出会失败。
            需要确保组件ID正确且资源存在。
        """
        id = str(uuid.uuid4())
        result = self.impex_client.export("operator", id, Headers)
        assert result[0] == 404

    @allure.title("对算子/工具箱/MCP无查看权限，导出失败")
    def test_export_13(self):
        """
        测试用例13：异常场景 - 无查看权限导致导出失败
        
        测试场景：
            - 使用无查看权限的用户（t2）尝试导出算子、工具箱、MCP
            - 验证导出失败
        
        验证点：
            - 导出接口返回403状态码（Forbidden）
        
        说明：
            导出资源需要对该资源有查看权限。
            如果无查看权限，导出会失败。
        """
        result = self.impex_client.export("operator", TestExport.t1_operator_id, TestExport.t2_headers)
        assert result[0] == 403
        result = self.impex_client.export("toolbox", TestExport.t1_toolbox_id, TestExport.t2_headers)
        assert result[0] == 403
        result = self.impex_client.export("mcp", TestExport.t1_mcp_id, TestExport.t2_headers)
        assert result[0] == 403

    @allure.title("对工具箱引用的算子无查看权限，导出失败")
    def test_export_14(self, Headers):
        """
        测试用例14：异常场景 - 无算子查看权限导致工具箱导出失败
        
        测试场景：
            1. 授权用户t2对工具箱有查看权限，但对工具箱引用的算子无查看权限
            2. 尝试导出该工具箱
            3. 验证导出失败
        
        验证点：
            - 导出接口返回403状态码（Forbidden）
        
        说明：
            如果工具箱包含从算子转换而来的工具，导出工具箱时需要对算子也有查看权限。
            即使工具箱本身有查看权限，如果算子无查看权限，导出也会失败。
        """
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
        """
        测试用例15：异常场景 - 无工具查看权限导致MCP导出失败
        
        测试场景：
            1. 授权用户t2对MCP有查看权限，但对MCP引用的工具无查看权限
            2. 尝试导出该MCP
            3. 验证导出失败
        
        验证点：
            - 导出接口返回403状态码（Forbidden）
        
        说明：
            如果MCP引用了工具，导出MCP时需要对工具也有查看权限。
            即使MCP本身有查看权限，如果工具无查看权限，导出也会失败。
        """
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
        """
        测试用例16：异常场景 - 无算子查看权限导致MCP导出失败
        
        测试场景：
            1. 授权用户t2对MCP和工具箱有查看权限，但对由算子转换成的工具（算子）无查看权限
            2. 尝试导出该MCP（工具箱包含从算子导入的工具）
            3. 验证导出失败
        
        验证点：
            - 导出接口返回403状态码（Forbidden）
        
        说明：
            如果MCP引用的工具来自算子转换，导出MCP时需要对算子也有查看权限。
            即使MCP和工具箱都有查看权限，如果算子无查看权限，导出也会失败。
        """
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
        """
        测试用例17：异常场景 - 引用的算子不存在导致工具箱导出失败
        
        测试场景：
            1. 下架并删除工具箱引用的算子
            2. 尝试导出该工具箱
            3. 验证导出失败
        
        验证点：
            - 导出接口返回404状态码（Not Found）
        
        说明：
            如果工具箱包含从算子转换而来的工具，且算子不存在，导出会失败。
            导出时需要确保所有依赖的资源都存在。
        """
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
    
