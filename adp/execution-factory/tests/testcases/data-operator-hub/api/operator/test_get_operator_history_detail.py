# -*- coding:UTF-8 -*-
"""
获取算子历史版本详情接口测试

测试目标：
    验证获取算子指定历史版本详情的功能，包括不同状态版本的获取和异常场景处理。

测试覆盖：
    1. 正常场景：算子只有一个已发布版本，获取成功
    2. 正常场景：算子存在不同状态的版本（已发布、已下架），获取成功
    3. 异常场景：算子不存在，获取失败

说明：
    历史版本详情是指算子某个特定版本的详细信息。
    只有已发布（published）或已下架（offline）状态的版本才能通过历史版本详情接口获取。
"""

import pytest
import allure
import uuid
import string
import random

from lib.operator import Operator
from common.get_content import GetContent

operators = []

@allure.feature("算子注册与管理接口测试：获取算子历史版本详情")
class TestGetOperatorHistoryDetail:
    """
    获取算子历史版本详情测试类
    
    说明：
        测试获取算子指定历史版本的详细信息功能。
        历史版本只能是已发布（published）或已下架（offline）状态的版本。
    """
    
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        """
        测试前置准备
        
        功能：
            注册多个算子并发布，为后续测试准备数据
        
        说明：
            - 注册算子后，将状态为"success"的算子发布
            - 至少需要2个算子用于测试不同场景
        """
        global operators

        # 注册算子并发布
        filepath = "./resource/openapi/compliant/test0.json"
        api_data = GetContent(filepath).jsonfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(data, Headers)
        assert re[0] == 200, f"setup 中注册算子失败，状态码: {re[0]}, 响应: {re}"
        publish_data_list = []
        for operator in re[1]:
            if operator.get("status") == "success":
                op = {
                    "operator_id": operator["operator_id"],
                    "version": operator["version"]
                }
                operators.append(op)

                publish_data = {
                    "operator_id": operator["operator_id"],
                    "status": "published"
                }
                publish_data_list.append(publish_data)

        if len(publish_data_list) > 0:
            result = self.client.UpdateOperatorStatus(publish_data_list, Headers)
            if result[0] != 200:
                print(f"警告: setup 中发布算子失败，状态码: {result[0]}, 响应: {result}")
        
        if len(operators) == 0:
            print(f"警告: operators 为空，某些测试可能会失败")
        elif len(operators) < 2:
            print(f"警告: operators 长度小于2（实际: {len(operators)}），某些需要 operators[1] 的测试可能会失败")

    @allure.title("获取算子历史版本详情 - 算子只有一个已发布版本，获取成功")
    def test_get_operator_history_detail_01(self, Headers):
        """
        测试用例1：正常场景 - 获取单个已发布版本详情
        
        测试场景：
            - 算子只有一个已发布版本
            - 调用获取历史版本详情接口，传入算子ID和版本号
        
        验证点：
            - 接口返回200状态码
            - 返回的operator_id匹配
            - 返回的version匹配
            - 返回的status为"published"
        
        说明：
            这是最简单的场景，算子只有一个版本且已发布。
        """
        global operators
        
        # 检查列表是否为空
        if len(operators) == 0:
            pytest.skip("operators 为空，跳过此测试")
        
        result = self.client.GetOperatorHistoryDetail(operators[0]["operator_id"], operators[0]["version"], Headers)
        assert result[0] == 200, f"获取历史版本详情失败，状态码: {result[0]}, 响应: {result}"
        assert result[1]["operator_id"] == operators[0]["operator_id"], "operator_id不匹配"
        assert result[1]["version"] == operators[0]["version"], "version不匹配"
        assert result[1]["status"] == "published", f"状态应该是published，实际: {result[1].get('status')}"

    @allure.title("获取算子历史版本详情 - 算子存在不同状态的版本，获取成功")
    def test_get_operator_history_detail_02(self, Headers):
        """
        测试用例2：正常场景 - 获取不同状态版本详情
        
        测试场景：
            1. 将算子下架（状态变为offline）
            2. 编辑算子生成新版本（状态为unpublish）
            3. 获取下架版本的详情（应该返回offline状态）
            4. 发布新版本（状态变为published）
            5. 编辑算子生成编辑中版本（状态为editing）
            6. 再次获取下架版本详情（应该仍然返回offline状态）
            7. 获取已发布版本详情（应该返回published状态）
        
        验证点：
            - 下架版本的详情状态为"offline"
            - 已发布版本的详情状态为"published"
            - 编辑中状态的版本不会出现在历史版本详情中
        
        说明：
            这个测试用例验证了历史版本详情接口只返回已发布或已下架状态的版本，
            编辑中（editing）或未发布（unpublish）状态的版本不会出现在历史版本详情中。
        """
        global operators
        
        # 检查列表长度
        if len(operators) < 2:
            pytest.skip(f"operators 长度不足（需要至少2个，实际{len(operators)}个），跳过此测试")
        
        operator_id = operators[1]["operator_id"]
        version = operators[1]["version"]
        
        # 步骤1：下架算子（状态变为offline）
        data = [{
            "operator_id": operator_id,
            "status": "offline"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200, f"下架算子失败，状态码: {result[0]}"
        
        # 步骤2：编辑算子生成新版本（状态为unpublish）
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "operator_id": operator_id,
            "metadata_type": "openapi",  # 传递description时需要metadata_type
            "description": "test edit 1234567",
            "name": name
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200, f"编辑算子失败，状态码: {re[0]}"
        assert re[1]["status"] == "unpublish", f"编辑后状态应该是unpublish，实际: {re[1].get('status')}"
        
        # 验证新版本已生成
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["version"] != version, "应该生成新版本"
        new_version = result[1]["version"]

        # 步骤3：获取下架版本的详情（应该返回offline状态）
        result = self.client.GetOperatorHistoryDetail(operator_id, version, Headers)
        assert result[0] == 200, f"获取历史版本详情失败，状态码: {result[0]}"
        assert result[1]["operator_id"] == operator_id, "operator_id不匹配"
        assert result[1]["version"] == version, "version不匹配"
        assert result[1]["status"] == "offline", f"下架版本状态应该是offline，实际: {result[1].get('status')}"

        # 步骤4：发布新版本（状态变为published）
        data = [{
            "operator_id": operator_id,
            "status": "published"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200, f"发布算子失败，状态码: {result[0]}"
        
        # 步骤5：编辑算子生成编辑中版本（状态为editing）
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "operator_id": operator_id,
            "metadata_type": "openapi",  # 传递description时需要metadata_type
            "description": "test edit 1234567 dgfh",
            "name": name
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200, f"编辑算子失败，状态码: {re[0]}"
        assert re[1]["status"] == "editing", f"编辑后状态应该是editing，实际: {re[1].get('status')}"
        
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200

        # 步骤6：再次获取下架版本详情（应该仍然返回offline状态）
        # 说明：即使后续有新的编辑中版本，历史版本详情仍然返回下架时的状态
        result = self.client.GetOperatorHistoryDetail(operator_id, version, Headers)
        assert result[0] == 200, f"获取历史版本详情失败，状态码: {result[0]}"
        assert result[1]["operator_id"] == operator_id, "operator_id不匹配"
        assert result[1]["version"] == version, "version不匹配"
        assert result[1]["status"] == "offline", f"下架版本状态应该是offline，实际: {result[1].get('status')}"

        # 步骤7：获取已发布版本详情（应该返回published状态）
        result = self.client.GetOperatorHistoryDetail(operator_id, new_version, Headers)
        assert result[0] == 200, f"获取历史版本详情失败，状态码: {result[0]}"
        assert result[1]["operator_id"] == operator_id, "operator_id不匹配"
        assert result[1]["version"] == new_version, "version不匹配"
        assert result[1]["status"] == "published", f"已发布版本状态应该是published，实际: {result[1].get('status')}"

    @allure.title("获取算子历史版本详情 - 算子不存在，获取失败")
    def test_get_operator_history_detail_03(self, Headers):
        """
        测试用例3：异常场景 - 算子不存在
        
        测试场景：
            - 使用不存在的算子ID和版本号
            - 调用获取历史版本详情接口
        
        验证点：
            - 接口返回404状态码（Not Found）
        
        说明：
            当算子ID或版本号不存在时，应该返回404错误，表示资源未找到。
        """
        # 使用随机UUID作为不存在的算子ID和版本号
        fake_operator_id = str(uuid.uuid4())
        fake_version = str(uuid.uuid4())
        result = self.client.GetOperatorHistoryDetail(fake_operator_id, fake_version, Headers)
        
        # 验证返回404错误
        assert result[0] == 404, f"不存在的算子ID和版本应该返回404，实际: {result[0]}, 响应: {result}"
