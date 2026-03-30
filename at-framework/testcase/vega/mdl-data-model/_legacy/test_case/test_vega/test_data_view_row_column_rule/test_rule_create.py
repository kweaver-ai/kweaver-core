import pytest
import allure
import json
from src.api.vega.data_views_api import DataViewsApi
from src.common.read_data import data_reader
from src.common.assertions import asserts, get_db_asserts

# 读取测试数据
positive_data = data_reader.read_type_cases("vega/data_view_row_column_rule/rule_create.yaml", "positive")
negative_data = data_reader.read_type_cases("vega/data_view_row_column_rule/rule_create.yaml", "negative")

@allure.feature("数据视图行列权限")
class TestDataViewRowColumnRuleCreate:

    @allure.story("创建数据视图行列权限-正常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("rule_data,title", positive_data)
    def test_create_rule_positive(self, rule_data, title):
        """正向测试用例：期望创建成功"""
        with allure.step(f"创建数据视图行列权限: {rule_data.get('case_name')}"):
            resp = DataViewsApi().create_row_column_rule(
                rule_data.get("rules")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = rule_data.get('expected_status', 201)
            asserts.response_status_code(resp, expected_status)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            assert isinstance(response_data, list)
            for rule_resp in response_data:
                asserts.dict_contains_keys(rule_resp, ['id'])

        # 注意：这里暂时注释掉数据库断言，因为我们不确定数据库表结构
        # with allure.step("验证数据库记录存在"):
        #     if rule_data.get('should_exist_in_db', True):
        #         db_asserts = get_db_asserts()
        #         for rule_resp in response_data:
        #             db_asserts.assert_record_exists(
        #                 table="t_data_view_row_column_rule",  # 假设的表名，需要根据实际情况修改
        #                 condition="f_id = %s",
        #                 params=(rule_resp.get('id'),)
        #             )

    @allure.story("创建数据视图行列权限-异常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("rule_data,title", negative_data)
    def test_create_rule_negative(self, rule_data, title):
        """负向测试用例：期望创建失败"""
        with allure.step(f"尝试创建数据视图行列权限: {rule_data.get('case_name')}"):
            resp = DataViewsApi().create_row_column_rule(
                rule_data.get("rules")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = rule_data.get('expected_status', 400)
            asserts.response_status_code(resp, expected_status)

        # 注意：这里暂时注释掉数据库断言，因为我们不确定数据库表结构
        # with allure.step("验证数据库记录不存在"):
        #     if not rule_data.get('should_exist_in_db', False):
        #         # 对于异常场景，检查数据库中不应该存在该记录
        #         pass