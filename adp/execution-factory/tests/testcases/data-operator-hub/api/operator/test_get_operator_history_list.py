# -*- coding:UTF-8 -*-
"""
获取算子历史版本列表接口测试

测试目标：
    验证获取算子历史版本列表的功能，包括正常获取和异常场景处理。

测试覆盖：
    1. 正常场景：算子存在且有多个历史版本，获取成功
    2. 异常场景：算子不存在，获取失败

说明：
    算子历史版本是指算子在不同状态（已发布、已下架、编辑中）下的所有版本记录。
    只有已发布（published）或已下架（offline）状态的版本才会出现在历史版本列表中。
"""

import pytest
import allure
import string
import random
import uuid

from lib.operator import Operator
from common.get_content import GetContent

operator_id = ""

@allure.feature("算子注册与管理接口测试：获取算子历史版本列表")
class TestGetOperatorHistoryList:
    """
    获取算子历史版本列表测试类
    
    说明：
        测试获取算子的所有历史版本列表功能。历史版本包括已发布和已下架状态的版本。
    """
    
    client = Operator()

    def _edit_and_publish_operator(self, operator_id, filepath, summary_suffix, description_suffix, Headers):
        """
        辅助方法：编辑算子并发布，生成新版本
        
        参数：
            operator_id: 算子ID
            filepath: OpenAPI文件路径
            summary_suffix: summary字段的后缀（用于区分不同版本）
            description_suffix: description字段的后缀
            Headers: 请求头
        
        返回：
            bool: 是否成功
        """
        api_data = GetContent(filepath).yamlfile()
        
        # 修改 OpenAPI 数据，生成不同版本的内容
        for path_key in list(api_data.get("paths", {}).keys()):
            for method_key in list(api_data["paths"][path_key].keys()):
                if "summary" in api_data["paths"][path_key][method_key]:
                    api_data["paths"][path_key][method_key]["summary"] = f"test_edit_{summary_suffix}"
                if "description" in api_data["paths"][path_key][method_key]:
                    api_data["paths"][path_key][method_key]["description"] = f"test edit {description_suffix}"
        
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "operator_id": operator_id,
            "metadata_type": "openapi",
            "data": api_data,
            "name": name
        }
        
        re = self.client.EditOperator(data, Headers)
        # 如果编辑失败，尝试只传递 name（降级处理）
        if re[0] != 200:
            print(f"警告: 编辑算子失败，状态码: {re[0]}, 尝试只传递 name")
            data = {
                "operator_id": operator_id,
                "name": name
            }
            re = self.client.EditOperator(data, Headers)
        
        if re[0] != 200:
            return False
        
        # 如果编辑后状态为 editing，需要发布
        if re[1].get("status") == "editing":
            data = [{
                "operator_id": operator_id,
                "status": "published"
            }]
            result = self.client.UpdateOperatorStatus(data, Headers)
            if result[0] != 200:
                print(f"警告: 发布算子失败，状态码: {result[0]}, 继续执行测试")
                return False
        
        return True

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        """
        测试前置准备
        
        功能：
            创建一个算子并生成4个历史版本：
            1. 初始版本：注册后直接发布
            2. 第二个版本：编辑已发布算子后发布
            3. 第三个版本：下架后编辑并发布
            4. 第四个版本：再次编辑并发布
        
        说明：
            - 通过多次编辑和发布操作，生成多个历史版本
            - 每个版本都有不同的内容（通过修改summary和description区分）
            - 下架操作会创建一个历史版本记录（offline状态）
            - 最终生成4个历史版本用于测试
        """
        global operator_id

        # 步骤1：注册算子并直接发布（生成第1个版本）
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True  # 直接发布，生成第1个版本
        }
        re = self.client.RegisterOperator(data, Headers)
        assert re[0] == 200, f"注册算子失败，状态码: {re[0]}, 响应: {re}"
        assert len(re[1]) > 0, "注册算子返回空列表"
        operator_id = re[1][0]["operator_id"]

        # 步骤2：编辑已发布算子并发布（生成第2个版本）
        # 场景：已发布状态 -> 编辑 -> 发布
        success = self._edit_and_publish_operator(
            operator_id, 
            filepath, 
            "1234567", 
            "1234567", 
            Headers
        )
        assert success, "编辑并发布算子失败，无法生成第2个版本"

        # 步骤3：下架后编辑，再发布，再编辑生成新版本（生成第3个版本）
        # 场景：已发布 -> 下架 -> 编辑 -> 发布 -> 编辑
        # 下架算子
        data = [{
            "operator_id": operator_id,
            "status": "offline"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200, f"下架算子失败，状态码: {result[0]}"
        
        # 下架后编辑（生成新版本，状态为unpublish）
        success = self._edit_and_publish_operator(
            operator_id, 
            filepath, 
            "1234567_dfhgkjh", 
            "1234567 xfhg", 
            Headers
        )
        assert success, "下架后编辑算子失败，无法生成第3个版本"
        
        # 发布该版本
        data = [{
            "operator_id": operator_id,
            "status": "published"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200, f"发布算子失败，状态码: {result[0]}"
        
        # 再次编辑生成新版本（第3个版本的编辑状态）
        success = self._edit_and_publish_operator(
            operator_id, 
            filepath, 
            "hgfjhgj7", 
            "ghkjkj", 
            Headers
        )
        assert success, "发布后再次编辑算子失败"
        
    @allure.title("获取算子历史版本列表 - 算子存在且有多个版本，获取成功")
    def test_get_operator_history_list_01(self, Headers):
        """
        测试用例1：正常场景 - 获取算子历史版本列表
        
        测试场景：
            - 算子存在且有4个历史版本（已发布或已下架状态）
            - 调用获取历史版本列表接口
        
        验证点：
            - 接口返回200状态码
            - 返回的数据是列表类型
            - 列表包含4个版本
            - 所有版本的operator_id都匹配
        
        说明：
            历史版本列表只包含已发布（published）或已下架（offline）状态的版本，
            编辑中（editing）或未发布（unpublish）状态的版本不会出现在历史版本列表中。
            setup中通过以下操作生成了4个历史版本：
            1. 注册并直接发布（版本1 - published）
            2. 编辑并发布（版本2 - published，然后下架变为offline）
            3. 下架后编辑并发布（版本3 - published）
            4. 再次编辑并发布（版本4 - published）
        """
        global operator_id
        
        # 查询历史版本列表
        result = self.client.GetOperatorHistoryList(operator_id, Headers)
        
        # 验证接口调用成功
        assert result[0] == 200, f"获取历史版本列表失败，状态码: {result[0]}, 响应: {result}"
        
        # 验证返回数据类型
        assert isinstance(result[1], list), f"返回数据应该是列表类型，实际: {type(result[1])}"
        
        # 验证版本数量（应该包含4个历史版本）
        assert len(result[1]) == 4, f"应该返回4个历史版本，实际: {len(result[1])}"
        
        # 验证所有版本的operator_id都匹配
        assert all(item["operator_id"] == operator_id for item in result[1]), \
            "所有历史版本的operator_id应该一致"

    @allure.title("获取算子历史版本列表 - 算子不存在，获取失败")
    def test_get_operator_history_list_02(self, Headers):
        """
        测试用例2：异常场景 - 算子不存在
        
        测试场景：
            - 使用不存在的算子ID（随机UUID）
            - 调用获取历史版本列表接口
        
        验证点：
            - 接口返回404状态码（Not Found）
        
        说明：
            当算子ID不存在时，应该返回404错误，表示资源未找到。
        """
        # 使用随机UUID作为不存在的算子ID
        fake_operator_id = str(uuid.uuid4())
        result = self.client.GetOperatorHistoryList(fake_operator_id, Headers)
        
        # 验证返回404错误
        assert result[0] == 404, f"不存在的算子ID应该返回404，实际: {result[0]}, 响应: {result}"
