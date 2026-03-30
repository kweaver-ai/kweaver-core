import pytest
import allure
from src.api.ontology.kn_object_type_api import ObjectTypeApi
from src.common.business_util import random_data_properties
from test_case.ontology_base_test import BaseTest


@allure.feature("知识网络对象类型管理")
class TestObjectTypeDelete(BaseTest):
    """对象类型删除相关测试"""

    api = ObjectTypeApi()

    @allure.story("删除建对象类")
    @allure.title("删除建对象类")
    def test_delete_object_type(self, test_kn_id_use):
        self.api.create_object_types(
            kn_id=test_kn_id_use,
            object_id="test_delete",
            name="test_delete",
            data_source="",
            data_properties=random_data_properties(),
        )

        resp = self.api.delete_objects(test_kn_id_use, "test_delete")

        self.assert_response_code(resp, 204)

        self.assert_db_record_not_exists(
            table="t_object_type", condition="f_name = %s", params=("test_delete",)
        )
