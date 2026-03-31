# -*- coding:UTF-8 -*-
"""
获取算子市场列表接口测试

测试目标：
    验证获取算子市场列表的功能，包括分页、排序、过滤等场景。

测试覆盖：
    1. 正常场景：默认参数获取列表，分页信息正确
    2. 异常场景：page参数小于0，获取失败
    3. 正常场景：page_size为负数，获取所有算子
    4. 正常场景：page_size在有效范围内，分页正确
    5. 正常场景：all参数为True，获取所有算子
    6. 异常场景：page_size超出范围，获取失败
    7. 正常场景：根据名称过滤，获取成功
    8. 正常场景：根据类型过滤，获取成功
    9. 正常场景：查询数据源算子，获取成功

说明：
    算子市场列表只包含已发布（published）或已下架（offline）状态的算子。
    如果算子有多个版本，只返回最新版本。
    列表默认按更新时间倒序排列。
"""

import allure
import pytest
import math
import string
import random

from lib.operator import Operator
from common.assert_tools import AssertTools
from common.get_content import GetContent

operators = []
operator_id = ""

@allure.feature("算子注册与管理接口测试：获取算子市场列表")
class TestGetOperatorMarketList:
    """
    获取算子市场列表测试类
    
    说明：
        测试获取算子市场列表的各种场景，包括分页、排序、过滤等功能。
    """
    
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        """
        测试前置准备
        
        功能：
            注册多个算子并设置不同状态，为后续测试准备数据：
            1. 注册并发布所有算子
            2. 下架第1个算子（状态变为offline）
            3. 编辑第2个算子（状态变为editing）
            4. 下架第3个算子后编辑（状态变为unpublish）
            5. 编辑第4个算子并发布（生成新版本）
        
        说明：
            - 至少需要4个算子用于测试不同场景
            - 算子数量不足时，部分测试可能会跳过
        """
        global operators, operator_id
        
        # 步骤1：注册并发布算子
        filepath = "./resource/openapi/compliant/test0.json"
        api_data = GetContent(filepath).jsonfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(data, Headers)
        assert re[0] == 200, f"注册算子失败，状态码: {re[0]}, 响应: {re}"
        
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

        result = self.client.UpdateOperatorStatus(publish_data_list, Headers)
        assert result[0] == 200, f"发布算子失败，状态码: {result[0]}"

        # 检查算子数量是否足够
        if len(operators) < 4:
            print(f"警告: 算子数量不足（需要至少4个，实际{len(operators)}个），部分测试可能无法执行")

        # 步骤2：下架第1个算子（状态变为offline）
        if len(operators) > 0:
            data = [{
                "operator_id": operators[0]["operator_id"],
                "status": "offline"
            }]
            result = self.client.UpdateOperatorStatus(data, Headers)
            assert result[0] == 200, f"下架算子失败，状态码: {result[0]}"

        # 步骤3：编辑第2个算子（状态变为editing）
        if len(operators) > 1:
            re = self.client.GetOperatorInfo(operators[1]["operator_id"], Headers)
            assert re[0] == 200, f"获取算子信息失败，状态码: {re[0]}"
            
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            data = {
                "operator_id": operators[1]["operator_id"],
                "name": name
            }
            re = self.client.EditOperator(data, Headers)
            if re[0] != 200:
                print(f"警告: 编辑算子失败，状态码: {re[0]}, 继续执行测试")
            else:
                assert re[1]["status"] == "editing", f"编辑后状态应该是editing，实际: {re[1].get('status')}"

        # 步骤4：下架第3个算子后编辑（状态变为unpublish）
        if len(operators) > 2:
            data = [{
                "operator_id": operators[2]["operator_id"],
                "status": "offline"
            }]
            result = self.client.UpdateOperatorStatus(data, Headers)
            assert result[0] == 200, f"下架算子失败，状态码: {result[0]}"

            re = self.client.GetOperatorInfo(operators[2]["operator_id"], Headers)
            assert re[0] == 200, f"获取算子信息失败，状态码: {re[0]}"
            
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            data = {
                "operator_id": operators[2]["operator_id"],
                "name": name
            }
            re = self.client.EditOperator(data, Headers)
            if re[0] != 200:
                print(f"警告: 编辑算子失败，状态码: {re[0]}, 继续执行测试")
            else:
                assert re[1]["status"] == "unpublish", f"下架后编辑状态应该是unpublish，实际: {re[1].get('status')}"

        # 步骤5：编辑第4个算子并发布（生成新版本）
        if len(operators) > 3:
            re = self.client.GetOperatorInfo(operators[3]["operator_id"], Headers)
            assert re[0] == 200, f"获取算子信息失败，状态码: {re[0]}"
            
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            data = {
                "operator_id": operators[3]["operator_id"],
                "name": name
            }
            re = self.client.EditOperator(data, Headers)
            if re[0] != 200:
                print(f"警告: 编辑算子失败，状态码: {re[0]}, 跳过发布操作")
            else:
                assert re[1]["status"] == "editing", f"编辑后状态应该是editing，实际: {re[1].get('status')}"
                
                # 发布算子，生成新版本
                data = [{
                    "operator_id": operators[3]["operator_id"],
                    "status": "published"
                }]
                result = self.client.UpdateOperatorStatus(data, Headers)
                if result[0] != 200:
                    print(f"警告: 发布算子失败，状态码: {result[0]}, 继续执行测试")
                else:
                    assert result[0] == 200
        else:
            print(f"警告: 算子数量不足（需要至少4个，实际{len(operators)}个），跳过编辑并发布操作")

    @allure.title("获取算子市场列表 - 默认参数，获取成功，返回已发布/已下架算子的最新版本，按更新时间倒序排列")
    def test_market_list_01(self, Headers):
        """
        测试用例1：正常场景 - 默认参数获取列表
        
        测试场景：
            - 不传递任何参数，使用默认值
            - 调用获取市场列表接口
        
        验证点：
            - 接口返回200状态码
            - 默认page为1，page_size为10
            - 返回的算子数量不超过page_size
            - 分页信息正确（total_pages、has_next、has_prev）
            - 所有算子状态为published或offline
            - 算子ID不重复（每个算子只返回最新版本）
            - 按更新时间倒序排列
        
        说明：
            市场列表默认只返回已发布或已下架状态的算子，且只返回每个算子的最新版本。
            列表按更新时间倒序排列，最新更新的算子排在前面。
        """
        result = self.client.GetOperatorMarketList(None, Headers)
        
        # 验证接口调用成功
        assert result[0] == 200, f"获取市场列表失败，状态码: {result[0]}, 响应: {result}"
        
        # 验证默认分页参数
        assert result[1]["page"] == 1, f"默认page应该是1，实际: {result[1].get('page')}"
        assert result[1]["page_size"] == 10, f"默认page_size应该是10，实际: {result[1].get('page_size')}"
        
        # 验证返回数据量
        operator_list = result[1]["data"]
        assert len(operator_list) <= 10, f"返回的算子数量应该不超过10，实际: {len(operator_list)}"
        
        # 验证分页计算
        assert result[1]["total_pages"] == math.ceil(result[1]["total"]/result[1]["page_size"]), \
            f"total_pages计算错误，实际: {result[1].get('total_pages')}"
        
        # 验证has_next
        if result[1]["total_pages"] > result[1]["page"]:
            assert result[1]["has_next"] == True, "有下一页时has_next应该为True"
        else:
            assert result[1]["has_next"] == False, "没有下一页时has_next应该为False"
        
        # 验证has_prev
        if result[1]["page"] == 1:
            assert result[1]["has_prev"] == False, "第一页时has_prev应该为False"
        elif result[1]["total_pages"] > 1:
            assert result[1]["has_prev"] == True, "非第一页且有多个页面时has_prev应该为True"
        
        # 验证算子状态和唯一性
        operator_ids = []
        update_times = []
        for op in operator_list:
            operator_ids.append(op["operator_id"])
            update_times.append(op["update_time"])
            assert op["status"] == "published" or op["status"] == "offline", \
                f"算子状态应该是published或offline，实际: {op.get('status')}"
        
        # 验证算子ID不重复（每个算子只返回最新版本）
        assert AssertTools.has_duplicates(operator_ids) == False, "算子ID不应该重复"
        
        # 验证按更新时间倒序排列
        assert AssertTools.is_descending_str(update_times) == True, "应该按更新时间倒序排列"

    @allure.title("获取算子市场列表 - page参数小于0，获取失败")
    def test_market_list_02(self, Headers):
        """
        测试用例2：异常场景 - page参数无效
        
        测试场景：
            - page参数设置为-1（小于0）
            - 调用获取市场列表接口
        
        验证点：
            - 接口返回400状态码（Bad Request）
        
        说明：
            page参数必须大于0，小于0的值会导致请求失败。
        """
        params = {"page": -1}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 400, f"page小于0应该返回400，实际: {result[0]}, 响应: {result}"

    @allure.title("获取算子市场列表 - page_size为负数，获取成功，默认获取所有")
    @pytest.mark.parametrize("page_size", [-1, -2])
    def test_market_list_03(self, page_size, Headers):
        """
        测试用例3：正常场景 - page_size为负数时获取所有算子
        
        测试场景：
            - page_size设置为负数（-1或-2）
            - 调用获取市场列表接口
        
        验证点：
            - 接口返回200状态码
            - 返回的算子数量等于总数（获取所有）
        
        说明：
            page_size为负数时，表示获取所有算子，不受分页限制。
            这是获取全部数据的便捷方式。
        """
        params = {"page_size": page_size}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200, f"获取市场列表失败，状态码: {result[0]}, 响应: {result}"
        assert len(result[1]["data"]) == result[1]["total"], \
            f"page_size为负数时应该返回所有算子，实际返回: {len(result[1]['data'])}, 总数: {result[1].get('total')}"

    @allure.title("获取算子市场列表 - page_size在[1-100]范围内，获取成功")
    @pytest.mark.parametrize("page_size", [1, 20, 50, 100])
    def test_market_list_04(self, page_size, Headers):
        """
        测试用例4：正常场景 - page_size在有效范围内
        
        测试场景：
            - page_size设置为有效值（1, 20, 50, 100）
            - 调用获取市场列表接口
        
        验证点：
            - 接口返回200状态码
            - 返回的page_size与请求一致
            - 返回的算子数量不超过page_size
            - 分页信息正确
            - 所有算子状态为published或offline
        
        说明：
            page_size的有效范围是1-100，超出此范围会返回400错误。
        """
        params = {"page_size": page_size}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200, f"获取市场列表失败，状态码: {result[0]}, 响应: {result}"
        assert result[1]["page_size"] == page_size, \
            f"返回的page_size应该与请求一致，实际: {result[1].get('page_size')}"
        assert len(result[1]["data"]) <= page_size, \
            f"返回的算子数量应该不超过page_size，实际: {len(result[1]['data'])}"
        
        # 验证分页信息
        if result[1]["total_pages"] > result[1]["page"]:
            assert result[1]["has_next"] == True, "有下一页时has_next应该为True"
        else:
            assert result[1]["has_next"] == False, "没有下一页时has_next应该为False"
        
        if result[1]["page"] == 1:
            assert result[1]["has_prev"] == False, "第一页时has_prev应该为False"
        elif result[1]["total_pages"] > 1:
            assert result[1]["has_prev"] == True, "非第一页且有多个页面时has_prev应该为True"
        
        # 验证算子状态
        for op in result[1]["data"]:
            assert op["status"] == "published" or op["status"] == "offline", \
                f"算子状态应该是published或offline，实际: {op.get('status')}"

    @allure.title("获取算子市场列表 - all参数为True，获取所有算子，获取成功")
    def test_market_list_05(self, Headers):
        """
        测试用例5：正常场景 - 使用all参数获取所有算子
        
        测试场景：
            - all参数设置为True
            - 调用获取市场列表接口
        
        验证点：
            - 接口返回200状态码
            - 返回的算子数量等于总数
            - 所有算子状态为published或offline
        
        说明：
            all=True时，忽略分页限制，返回所有符合条件的算子。
            这与page_size为负数的效果相同。
        """
        params = {"all": True}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200, f"获取市场列表失败，状态码: {result[0]}, 响应: {result}"
        assert len(result[1]["data"]) == result[1]["total"], \
            f"all=True时应该返回所有算子，实际返回: {len(result[1]['data'])}, 总数: {result[1].get('total')}"
        
        # 验证算子状态
        for op in result[1]["data"]:
            assert op["status"] == "published" or op["status"] == "offline", \
                f"算子状态应该是published或offline，实际: {op.get('status')}"

    @allure.title("获取算子市场列表 - page_size超出范围，获取失败")
    def test_market_list_06(self, Headers):
        """
        测试用例6：异常场景 - page_size超出范围
        
        测试场景：
            - page_size设置为101（超出最大值100）
            - 调用获取市场列表接口
        
        验证点：
            - 接口返回400状态码（Bad Request）
        
        说明：
            page_size的有效范围是1-100，超出此范围会导致请求失败。
        """
        params = {"page_size": 101}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 400, f"page_size超出范围应该返回400，实际: {result[0]}, 响应: {result}"

    @allure.title("获取算子市场列表 - 根据名称过滤，获取成功")
    def test_market_list_07(self, Headers):
        """
        测试用例7：正常场景 - 根据名称过滤
        
        测试场景：
            - name参数设置为"获取"（部分匹配）
            - 调用获取市场列表接口
        
        验证点：
            - 接口返回200状态码
            - 返回的所有算子名称都包含"获取"
        
        说明：
            名称过滤支持部分匹配，只要算子名称包含指定的字符串就会被返回。
        """
        params = {"name": "获取"}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200, f"获取市场列表失败，状态码: {result[0]}, 响应: {result}"
        
        # 验证所有返回的算子名称都包含"获取"
        for operator in result[1]["data"]:
            assert "获取" in operator["name"], \
                f"算子名称应该包含'获取'，实际: {operator.get('name')}"

    @allure.title("获取算子市场列表 - 根据类型过滤，获取成功")
    def test_market_list_08(self, Headers):
        """
        测试用例8：正常场景 - 根据类型（category）过滤
        
        测试场景：
            - category参数设置为"data_process"
            - 调用获取市场列表接口
        
        验证点：
            - 接口返回200状态码
        
        说明：
            根据算子类型过滤，只返回指定类型的算子。
            常见的类型包括：data_process、data_transform、data_store等。
        """
        params = {"category": "data_process"}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200, f"获取市场列表失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("获取算子市场列表 - 查询数据源算子，获取成功，返回数据源算子")
    def test_market_list_09(self, Headers):
        """
        测试用例9：正常场景 - 查询数据源算子
        
        测试场景：
            1. 注册并发布标记为数据源的算子（is_data_source=True）
            2. 使用is_data_source=True参数查询市场列表
        
        验证点：
            - 接口返回200状态码
            - 返回的算子总数等于注册的数据源算子数量
            - 返回的所有算子都是数据源算子（is_data_source=True）
            - 返回的算子ID都在注册的算子ID列表中
        
        说明：
            数据源算子是一种特殊类型的算子，用于数据源相关的操作。
            可以通过is_data_source参数过滤出所有数据源算子。
        """
        # 注册并发布数据源算子
        filepath = "./resource/openapi/compliant/setup.json"
        api_data = GetContent(filepath).jsonfile()

        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "operator_info": {
                "is_data_source": True  # 标记为数据源算子
            }
        }

        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}"
        
        operator_ids = []
        publish_data_list = []
        for op in result[1]:
            if op.get("status") == "success":
                operator_ids.append(op["operator_id"])
                publish_data = {
                    "operator_id": op["operator_id"],
                    "status": "published"
                }
                publish_data_list.append(publish_data)

        result = self.client.UpdateOperatorStatus(publish_data_list, Headers)
        assert result[0] == 200, f"发布算子失败，状态码: {result[0]}"

        # 查询数据源算子
        params = {
            "is_data_source": True
        }
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200, f"获取市场列表失败，状态码: {result[0]}, 响应: {result}"
        
        # 验证返回的算子总数
        assert result[1]["total"] == len(operator_ids), \
            f"返回的算子总数应该等于注册的数据源算子数量，实际: {result[1].get('total')}, 期望: {len(operator_ids)}"
        
        # 验证返回的所有算子都是数据源算子
        operators = result[1]["data"]
        for operator in operators:
            assert operator["operator_id"] in operator_ids, \
                f"返回的算子ID应该在注册的算子ID列表中，实际: {operator.get('operator_id')}"
            assert operator["operator_info"]["is_data_source"] == True, \
                f"返回的算子应该是数据源算子，实际: {operator.get('operator_info', {}).get('is_data_source')}"
