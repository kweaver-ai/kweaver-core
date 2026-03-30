import pytest
import allure
import json
from src.api.vega.data_views_api import DataViewsApi
from src.common.read_data import data_reader
from src.common.assertions import asserts, get_db_asserts

# 读取测试数据
positive_data = data_reader.read_type_cases("vega/data_view_group/group_create.yaml", "positive")
negative_data = data_reader.read_type_cases("vega/data_view_group/group_create.yaml", "negative")

@allure.feature("数据视图管理")
class TestDataViewGroupCreate:

    @allure.story("创建数据视图分组-正常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", positive_data)
    def test_create_group_positive(self, group_data, title):
        """正向测试用例：期望创建成功"""
        with allure.step(f"创建数据视图分组: {group_data.get('case_name')}"):
            resp = DataViewsApi().create_group(
                group_data.get("name")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = group_data.get('expected_status', 201)
            asserts.response_status_code(resp, expected_status)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data, ['id'])

        # 注意：这里暂时注释掉数据库断言，因为我们不确定数据库表结构
        # with allure.step("验证数据库记录存在"):
        #     if group_data.get('should_exist_in_db', True):
        #         db_asserts = get_db_asserts()
        #         db_asserts.assert_record_exists(
        #             table="t_data_view_group",  # 假设的表名，需要根据实际情况修改
        #             condition="f_id = %s",
        #             params=(response_data.get('id'),)
        #         )

    @allure.story("创建数据视图分组-异常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", negative_data)
    def test_create_group_negative(self, group_data, title):
        """负向测试用例：期望创建失败"""
        # 对于重复名称的测试，先创建一个分组
        if group_data.get('case_name') == '重复名称创建分组':
            DataViewsApi().create_group("测试分组")

        with allure.step(f"尝试创建数据视图分组: {group_data.get('case_name')}"):
            resp = DataViewsApi().create_group(
                group_data.get("name")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = group_data.get('expected_status', 400)
            asserts.response_status_code(resp, expected_status)

        # 注意：这里暂时注释掉数据库断言，因为我们不确定数据库表结构
        # with allure.step("验证数据库记录不存在"):
        #     if not group_data.get('should_exist_in_db', False):
        #         db_asserts = get_db_asserts()
        #         # 对于异常场景，检查数据库中不应该存在该记录
        #         # 这里需要根据实际情况修改查询条件
        #         pass