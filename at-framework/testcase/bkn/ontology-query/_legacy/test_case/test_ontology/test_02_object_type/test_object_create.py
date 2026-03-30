# coding: utf-8
"""
对象类型创建测试用例
"""
import pytest
import allure
from src.api.ontology.kn_object_type_api import ObjectTypeApi
from src.api.vega.data_views_api import DataViewsApi
from src.common.business_util import random_data_properties, error_properties
from src.common.read_data import data_reader
from test_case.ontology_base_test import BaseTest

positive_data = data_reader.read_yaml("ontology/object_type/object_create_postive.yaml")
negative_data = data_reader.read_yaml(
    "ontology/object_type/object_create_negative.yaml"
)
negative_properties_data = data_reader.read_yaml(
    "ontology/object_type/object_create_negative_by_properties.yaml"
)


@allure.feature("知识网络对象类型管理")
class TestObjectTypeCreate(BaseTest):
    """对象类型创建相关测试"""

    api = ObjectTypeApi()
    view_api = DataViewsApi()

    @allure.story("创建对象类-参数校验")
    @allure.title("{object_data[case_name]}")
    @pytest.mark.parametrize("object_data", positive_data)
    def test_create_object_type_positive(self, object_data, test_kn_id_use):

        resp = self.api.create_object_types(
            kn_id=test_kn_id_use,
            object_id=object_data["object_id"],
            name=object_data["object_name"],
            data_source="",
            tags=object_data.get("tags", None),
            comment=object_data.get("comment", None),
            branch=object_data.get("branch", None),
            data_properties=random_data_properties(
                object_data["data_properties"], static=object_data.get("static", False)
            ),
            primary_keys=object_data.get("primary_key", None),
            incremental_key=object_data.get("incremental_key", None),
            display_key=object_data.get("display_key", None),
            exclude_fields=object_data.get("exclude_fields", None),
        )
        # 验证响应状态码和结构
        self.assert_response_code(resp, 201)
        self.assert_response_structure_list(resp, ["id"])
        # 验证数据库记录存在
        self.assert_db_record_exists(
            self.object_table,
            params=object_data["object_name"],
            condition="f_name = %s",
        )

    @allure.story("创建对象类-参数校验-异常场景")
    @allure.title("{object_data[case_name]}")
    @pytest.mark.parametrize("object_data", negative_data)
    def test_create_object_type_negative(self, object_data, test_kn_id_use):
        if object_data.get("note") == "重复ID—name":
            resp = self.api.create_object_types(
                kn_id=test_kn_id_use,
                object_id=object_data["object_id"],
                name=object_data["object_name"],
                data_source="",
                data_properties=random_data_properties(),
            )
        resp = self.api.create_object_types(
            kn_id=test_kn_id_use,
            object_id=object_data["object_id"],
            name=object_data["object_name"],
            data_source="",
            tags=object_data.get("tags", None),
            comment=object_data.get("comment", None),
            data_properties=random_data_properties(
                object_data["data_properties"], static=object_data.get("static", False)
            ),
            primary_keys=object_data.get("primary_key", None),
            incremental_key=object_data.get("incremental_key", None),
            display_key=object_data.get("display_key", None),
            exclude_fields=object_data.get("exclude_fields", None),
        )
        # 验证响应状态码和结构
        self.assert_response_code(resp, object_data["expected_status"])
        # 验证数据库记录不存在
        if not object_data.get("note") == "name为空字符串应该失败":

            self.assert_db_record_not_exists(
                self.object_table, params=object_data.get("object_name", "test")
            )

    @allure.story("创建对象类-参数校验-数据属性-异常场景")
    @allure.title("{object_data[case_name]}")
    @pytest.mark.parametrize("object_data", negative_properties_data)
    def test_create_object_type_negative_by_properties(
        self, object_data, test_kn_id_use
    ):
        resp = self.api.create_object_types(
            kn_id=test_kn_id_use,
            object_id=object_data["object_id"],
            name=object_data["object_name"],
            data_source="",
            data_properties=error_properties(object_data["data_properties"]),
            primary_keys=object_data.get("primary_key", None),
            incremental_key=object_data.get("incremental_key", None),
            display_key=object_data.get("display_key", None),
            exclude_fields=object_data.get("exclude_fields", None),
        )
        # 验证响应状态码和结构
        self.assert_response_code(resp, object_data["expected_status"])
        # 验证数据库记录不存在
        self.assert_db_record_not_exists(
            self.object_table, params=object_data["object_name"]
        )
