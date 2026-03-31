# -*- coding:UTF-8 -*-
"""
获取算子列表接口测试

测试目标：
    验证获取算子列表的功能，包括分页、排序、过滤等场景。

测试覆盖：
    1. 分页功能：默认分页、自定义分页、负数获取所有
    2. 排序功能：按时间/名称排序、升序/降序
    3. 过滤功能：按状态、类型、名称、用户、分类等过滤
    4. 组合过滤：多个条件组合查询
    5. 异常场景：参数无效、状态无效、类型无效

说明：
    获取算子列表接口支持丰富的查询功能，包括分页、排序、过滤等。
    如果算子有多个版本，只返回最新版本。
    列表默认按更新时间倒序排列。
"""

import pytest
import allure
import json
import math
import random
import string
import uuid

from common.get_content import GetContent
from common.assert_tools import AssertTools
from lib.operator import Operator

count = 0
operator_list = []
operator_ids = []
names = []
versions = []
status = ["unpublish", "published", "offline"]
categorys = ["other_category", "data_process", "data_transform", "data_store", "data_analysis", "data_query", "data_extract", "data_split", "model_train"]

@allure.feature("算子注册与管理接口测试：获取算子列表")
class TestGetOperatorList:
    """
    获取算子列表测试类
    
    说明：
        测试获取算子列表的各种场景，包括分页、排序、过滤等功能。
    """

    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        """
        测试前置准备
        
        功能：
            注册多个算子并设置不同状态，为后续测试准备数据：
            1. 分10次注册算子，每次使用不同的分类
            2. 发布70%的算子（状态为published）
            3. 下架20%的算子（状态为offline）
            4. 剩余10%保持未发布状态（状态为unpublish）
        
        说明：
            - 至少需要40+个算子用于测试各种场景
            - 算子分布在不同的分类中
            - 算子有不同的状态，用于测试状态过滤
        """
        global count
        global operator_list
        global operator_ids
        global names
        global versions
        global categorys

        # 步骤1：分10次注册算子，每次使用不同的分类
        for i in range(10):
            current_category = categorys[i % len(categorys)]
            filepath = "./resource/openapi/compliant/setup.json"
            json_data = GetContent(filepath).jsonfile()
            # 修改每个路径下的summary字段避免重名
            for path in json_data["paths"]:
                for method in json_data["paths"][path]:
                    if "summary" in json_data["paths"][path][method]:
                        json_data["paths"][path][method]["summary"] = "test_summary_" + ''.join(random.choice(string.ascii_letters) for i in range(8))
            req_data = {
                "data": str(json_data),
                "operator_metadata_type": "openapi",
                "operator_info": {
                    "category": current_category
                }
            }
            re = self.client.RegisterOperator(req_data, Headers)
            assert re[0] == 200
            operators = re[1]

            # 处理每个算子
            for operator in operators:
                if operator["status"] == "success":
                    count = count + 1
                    operator_id = operator["operator_id"]
                    version = operator["version"]
                    operator_ids.append(operator_id)
                    versions.append(version)

        # 步骤2：发布70%的算子（状态为published）
        update_data1 = []
        for i in range(int(count*0.7)):            
            update_data = {
                "operator_id": operator_ids[i],
                "version": versions[i],
                "status": "published"
            }
            update_data1.append(update_data)

        # 步骤3：下架20%的算子（状态为offline）
        update_data2 = []
        for i in range(int(count*0.2)):            
            update_data = {
                "operator_id": operator_ids[i],
                "version": versions[i],
                "status": "offline"
            }
            update_data2.append(update_data)

        re = self.client.UpdateOperatorStatus(update_data1, Headers)
        assert re[0] == 200

        re = self.client.UpdateOperatorStatus(update_data2, Headers)
        assert re[0] == 200

        # 获取算子信息并存储
        for i in range(count):
            re = self.client.GetOperatorInfo(operator_ids[i], Headers)
            assert re[0] == 200
            operator_list.append(re[1])
            names.append(re[1]["name"]) 
        
        print(f"\n算子总数: {count}")
        print(f"已发布算子数量: {len([op for op in operator_list if op['status'] == 'published'])}")
        print(f"已下架算子数量: {len([op for op in operator_list if op['status'] == 'offline'])}")
        print(f"未发布算子数量: {len([op for op in operator_list if op['status'] == 'unpublish'])}")
    
    @allure.title("获取算子列表 - 默认参数，每页10个，按照更新时间降序排列")
    def test_get_operator_list_00(self, Headers):
        """
        测试用例0：正常场景 - 默认参数获取列表
        
        测试场景：
            - 不传递任何参数，使用默认值
            - 调用获取算子列表接口
        
        验证点：
            - 接口返回200状态码
            - 返回总数等于注册的算子数量
            - 默认每页返回10个算子
            - 按更新时间倒序排列
        
        说明：
            默认情况下，列表按更新时间倒序排列，每页10个算子。
        """
        result = self.client.GetOperatorList(None, Headers)
        assert result[0] == 200, f"获取算子列表失败，状态码: {result[0]}, 响应: {result}"
        assert result[1]["total"] == count, \
            f"返回总数应该等于注册的算子数量，实际: {result[1].get('total')}, 期望: {count}"
        assert len(result[1]["data"]) == 10, \
            f"默认每页应该返回10个算子，实际: {len(result[1]['data'])}"

        # 验证按更新时间倒序排列
        update_times = []
        operators = result[1]["data"]
        for operator in operators:
            update_times.append(operator["update_time"])

        assert AssertTools.is_descending_str(update_times) == True, "应该按更新时间倒序排列"
    
    @allure.title("获取算子列表 - page参数小于0，获取失败")
    def test_get_operator_list_01(self, Headers):
        """
        测试用例1：异常场景 - page参数无效
        
        测试场景：
            - page参数设置为-1（小于0）
            - 调用获取算子列表接口
        
        验证点：
            - 接口返回400状态码（Bad Request）
        
        说明：
            page参数必须大于0，小于0的值会导致请求失败。
        """
        data = {
            "page": -1
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 400, f"page小于0应该返回400，实际: {result[0]}, 响应: {result}"

    @allure.title("获取算子列表 - page_size为负数，获取成功，返回所有算子")
    @pytest.mark.parametrize("page_size", [-1, -2])
    def test_get_operator_list_02(self, page_size, Headers):
        """
        测试用例2：正常场景 - page_size为负数时获取所有算子
        
        测试场景：
            - page_size设置为负数（-1或-2）
            - 调用获取算子列表接口
        
        验证点：
            - 接口返回200状态码
            - 返回的算子数量等于总数（获取所有）
        
        说明：
            page_size为负数时，表示获取所有算子，不受分页限制。
        """
        global count

        data = {
            "page_size": page_size
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200, f"获取算子列表失败，状态码: {result[0]}, 响应: {result}"
        assert result[1]["total"] == count, \
            f"返回总数应该等于注册的算子数量，实际: {result[1].get('total')}, 期望: {count}"
        assert len(result[1]["data"]) == count, \
            f"page_size为负数时应该返回所有算子，实际返回: {len(result[1]['data'])}, 总数: {count}"

    @allure.title("获取算子列表 - page和page_size为0，获取成功，采用默认值")
    def test_get_operator_list_03(self, Headers):
        """
        测试用例3：正常场景 - page和page_size为0时使用默认值
        
        测试场景：
            - page和page_size都设置为0
            - 调用获取算子列表接口
        
        验证点：
            - 接口返回200状态码
            - page自动调整为1（默认值）
            - page_size自动调整为10（默认值）
            - 返回10个算子
        
        说明：
            page和page_size为0时，系统会自动使用默认值（page=1, page_size=10）。
        """
        data = {
            "page": 0,
            "page_size": 0
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200, f"获取算子列表失败，状态码: {result[0]}, 响应: {result}"
        assert result[1]["page"] == 1, f"page为0时应该自动调整为1，实际: {result[1].get('page')}"
        assert result[1]["page_size"] == 10, f"page_size为0时应该自动调整为10，实际: {result[1].get('page_size')}"
        assert len(result[1]["data"]) == 10, f"应该返回10个算子，实际: {len(result[1]['data'])}"
        assert result[1]["total"] == count, \
            f"返回总数应该等于注册的算子数量，实际: {result[1].get('total')}, 期望: {count}"

    @allure.title("获取算子列表 - page_size大于100，获取失败")
    def test_get_operator_list_04(self, Headers):
        """
        测试用例4：异常场景 - page_size超出范围
        
        测试场景：
            - page_size设置为101（超出最大值100）
            - 调用获取算子列表接口
        
        验证点：
            - 接口返回400状态码（Bad Request）
        
        说明：
            page_size的有效范围是1-100，超出此范围会导致请求失败。
        """
        data = {
            "page_size": 101
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 400, f"page_size超出范围应该返回400，实际: {result[0]}, 响应: {result}"
        
    @allure.title("获取算子列表 - 按照时间或名称排序，获取成功")
    @pytest.mark.parametrize("sort_by", ["create_time", "update_time", "name"])
    def test_get_operator_list_05(self, sort_by, Headers):
        """
        测试用例5：正常场景 - 按不同字段排序
        
        测试场景：
            - sort_by设置为create_time、update_time或name
            - 调用获取算子列表接口
        
        验证点：
            - 接口返回200状态码
        
        说明：
            支持按创建时间、更新时间或名称排序。
        """
        data = {
            "sort_by": sort_by
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200, f"获取算子列表失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("获取算子列表 - 按照升序或降序排序，获取成功")
    @pytest.mark.parametrize("sort_order", ["asc", "desc"])
    def test_get_operator_list_06(self, sort_order, Headers):
        """
        测试用例6：正常场景 - 按升序或降序排序
        
        测试场景：
            - sort_order设置为asc（升序）或desc（降序）
            - 调用获取算子列表接口
        
        验证点：
            - 接口返回200状态码
            - 列表按指定顺序排列（升序或降序）
        
        说明：
            支持升序（asc）和降序（desc）两种排序方式。
            默认排序方式为降序（desc）。
        """
        data = {
            "sort_order": sort_order
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200, f"获取算子列表失败，状态码: {result[0]}, 响应: {result}"
        
        # 验证排序顺序
        update_times = [] 
        for operator in result[1]["data"]:
            update_times.append(operator["update_time"])
            
        if sort_order == "desc":
            assert AssertTools.is_descending_str(update_times) == True, "降序排列时应该按时间倒序"
        if sort_order == "asc":
            assert AssertTools.is_ascending_str(update_times) == True, "升序排列时应该按时间正序"

    @allure.title("从第N-1页开始，获取算子列表，上一页和下一页标记均为true")
    def test_get_operator_list_07(self, Headers):
        global count

        page_size = int(count/3)
        total_pages = math.ceil(count/page_size)
        data = {
            "page": total_pages-1,
            "page_size": page_size
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        assert len(result[1]["data"]) == page_size
        assert result[1]["total"] == count
        assert result[1]["has_next"] == True
        assert result[1]["has_prev"] == True
        assert result[1]["total_pages"] == total_pages

    @allure.title("从第N页开始，获取算子列表，有上一页无下一页")
    def test_get_operator_list_08(self, Headers):
        global count

        page_size = int(count/3)
        total_pages = math.ceil(count/page_size)
        # print(total_pages, page_size)
        data = {
            "page": total_pages,
            "page_size": page_size
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        assert len(result[1]["data"]) <= page_size
        # print(len(result[1]["data"]))
        assert result[1]["total"] == count
        assert result[1]["has_next"] == False
        assert result[1]["has_prev"] == True
        assert result[1]["total_pages"] == total_pages

    @allure.title("根据operator_type获取算子列表，返回符合条件的算子的最新版本")
    def test_get_operator_list_09(self, Headers):
        global operator_list

        data = {
            "operator_type": "basic"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        operators = result[1]["data"]
        for operator in operators:
            assert operator["operator_info"]["operator_type"] == "basic"

    @allure.title("根据name获取算子列表，返回符合条件的算子的最新版本")
    def test_get_operator_list_10(self, Headers):
        global operator_list

        # 检查列表长度
        if len(operator_list) < 41:
            pytest.skip(f"operator_list 长度不足（需要至少41个，实际{len(operator_list)}个），跳过此测试")

        # 编辑算子，修改operator_info（修改category）
        # 注意：使用application/json格式时，operator_info应该是对象，而不是JSON字符串
        # Request.post使用json=data会自动序列化整个请求体，所以operator_info直接传递对象即可
        # 注意：当传递metadata_type时，如果原算子description为空，需要传递description字段
        # 先获取算子信息，确保有description
        operator_info = self.client.GetOperatorInfo(operator_list[40]["operator_id"], Headers)
        assert operator_info[0] == 200, f"获取算子信息失败，状态码: {operator_info[0]}"
        current_description = operator_info[1].get("description", "test edit description")
        
        data = {
            "operator_id": operator_list[40]["operator_id"],
            "metadata_type": "openapi",  # 传递operator_info时需要metadata_type
            "description": current_description,  # 确保description不为空
            "operator_info": {
                "category": "data_process"
            }
        }
        re = self.client.EditOperator(data, Headers)    # 编辑已发布算子
        assert re[0] == 200, f"编辑算子失败，状态码: {re[0]}, 响应: {re}"
        unpublish_version = re[1]["version"]

        data = {
            "name": operator_list[40]["name"]
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        assert len(result[1]["data"]) == 1
        operator = result[1]["data"][0]   # 获取到最新版本
        assert operator["name"] == operator_list[40]["name"]
        assert operator["version"] == unpublish_version
    
    @allure.title("根据用户获取算子列表，返回符合条件的算子")
    def test_get_operator_list_11(self, UserHeaders, Headers):
        global versions

        user_id = UserHeaders["x-account-id"]
        data = {
            "create_user": user_id
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        operators = result[1]["data"]
        for operator in operators:
            assert operator["create_user"] == "A0"
    
    @allure.title("根据status获取算子列表，返回符合条件的算子的最新版本")
    def test_get_operator_list_12(self, Headers):
        global status

        for item in status:
            data = {
                "status": item
            }
            result = self.client.GetOperatorList(data, Headers)
            assert result[0] == 200
            operators = result[1]["data"]
            operator_ids = []
            for operator in operators:
                assert operator["status"] == item
                operator_ids.append(operator["operator_id"])

            assert AssertTools.has_duplicates(operator_ids) == False
    
    @allure.title("根据category获取算子列表，返回符合条件的算子的最新版本")
    def test_get_operator_list_13(self, Headers):
        global categorys
        
        for category in categorys:
            data = {
                "category": category
            }
            result = self.client.GetOperatorList(data, Headers)
            assert result[0] == 200
            operator_ids = []
            operators = result[1]["data"]
            for operator in operators:
                assert operator["operator_info"]["category"] == category
                operator_ids.append(operator["operator_id"])

            assert AssertTools.has_duplicates(operator_ids) == False

    @allure.title("根据create_user+name获取算子列表，返回符合条件的算子的最新版本，无符合条件的算子时返回空列表")
    def test_get_operator_list_14(self, UserHeaders, Headers):
        # 不符合条件，返回空
        global operator_ids
        global operator_list

        user_id = UserHeaders["x-account-id"]
        data = {
            "create_user": user_id,
            "name": "test test"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        assert result[1]["data"] == []

        # 符合条件，返回算子列表
        name = operator_list[0]["name"]
        data = {
            "create_user": user_id,
            "name": name
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        operators = result[1]["data"]
        for operator in operators:
            assert operator["create_user"] == "A0"
            assert operator["name"] == name

    @allure.title("根据create_user+name+operator_type获取算子列表，返回符合条件的算子的最新版本，无符合条件的算子时返回空列表")
    def test_get_operator_list_15(self, UserHeaders, Headers):
        # 不符合条件，返回空
        global operator_ids
        global operator_list

        user_id = UserHeaders["x-account-id"]
        data = {
            "create_user": user_id,
            "name": "test test",
            "operator_type": "composite"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        assert result[1]["data"] == []

        # 符合条件，返回算子列表
        name = operator_list[0]["name"]
        data = {
            "create_user": user_id,
            "name": name,
            "operator_type": "basic"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        operators = result[1]["data"]
        for operator in operators:
            assert operator["create_user"] == "A0"
            assert operator["name"] == name
            assert operator["operator_info"]["operator_type"] == "basic"

    @allure.title("根据create_user+name+operator_type+status获取算子列表，返回符合条件的算子的最新版本，无符合条件的算子时返回空列表")
    def test_get_operator_list_16(self, UserHeaders, Headers):
        # 不符合条件，返回空
        global operator_ids
        global operator_list

        user_id = UserHeaders["x-account-id"]
        data = {
            "create_user": user_id,
            "name": "test test",
            "operator_type": "composite",
            "status": "published"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        assert result[1]["data"] == []

        # 符合条件，返回算子列表
        name = operator_list[0]["name"]
        status = operator_list[0]["status"]
        data = {
            "create_user": user_id,
            "name": name,
            "operator_type": "basic",
            "status": status
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        operators = result[1]["data"]
        for operator in operators:
            assert operator["create_user"] == "A0"
            assert operator["name"] == name
            assert operator["operator_info"]["operator_type"] == "basic"
            assert operator["status"] == status
   
    @allure.title("根据create_user+name+operator_type+status+category获取算子列表，返回符合条件的算子的最新版本，无符合条件的算子时返回空列表")
    def test_get_operator_list_17(self, UserHeaders, Headers):
        # 不符合条件，返回空
        global operator_ids
        global operator_list

        user_id = UserHeaders["x-account-id"]
        data = {
            "create_user": user_id,
            "name": "test test",
            "operator_type": "basic",
            "status": "published",
            "category": "data_extract"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        assert result[1]["data"] == []

        # 符合条件，返回算子列表
        name = operator_list[0]["name"]
        status = operator_list[0]["status"]
        category = operator_list[0]["operator_info"]["category"]
        data = {
            "create_user": user_id,
            "name": name,
            "operator_type": "basic",
            "status": status,
            "category": category
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        operators = result[1]["data"]
        for operator in operators:
            assert operator["create_user"] == "A0"
            assert operator["name"] == name
            assert operator["operator_info"]["operator_type"] == "basic"
            assert operator["status"] == status
            assert operator["operator_info"]["category"] == category

    @allure.title("算子状态无效，获取算子列表，获取失败")
    def test_get_operator_list_18(self, Headers):
        data = {
            "status": "unknown"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 400

    @allure.title("算子类型无效，获取算子列表，获取失败")
    def test_get_operator_list_19(self, Headers):
        data = {
            "category": "unknown"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 400

    @allure.title("all为True，获取算子列表成功，返回所有算子的最新版本")
    def test_get_operator_list_20(self, Headers):
        global count

        data = {
            "all": "true"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        assert result[1]["total"] == count
        assert len(result[1]["data"]) == count
        operator_ids = []
        operators = result[1]["data"]
        for operator in operators:
            operator_ids.append(operator["operator_id"])

        assert AssertTools.has_duplicates(operator_ids) == False

    @allure.title("查询数据源算子，获取算子列表成功，返回数据源算子")
    def test_get_operator_list_21(self, Headers):
        filepath = "./resource/openapi/compliant/test0.json"
        api_data = GetContent(filepath).jsonfile()

        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "operator_info": {
                "is_data_source": True
            }
        }

        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        operator_ids = []
        for op in result[1]:
            if op["status"] == "success":
                operator_ids.append(op["operator_id"])

        data = {
            "is_data_source": True
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        assert result[1]["total"] == len(operator_ids)
        operators = result[1]["data"]
        for operator in operators:
            assert operator["operator_id"] in operator_ids
            assert operator["operator_info"]["is_data_source"] == True
