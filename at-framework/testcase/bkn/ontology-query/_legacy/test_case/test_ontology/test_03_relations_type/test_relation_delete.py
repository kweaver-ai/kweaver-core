# coding: utf-8
"""
关系类型创建测试用例
"""
import pytest
import allure
import json
from src.api.ontology.kn_relation_type_api import RelationTypeApi
from src.api.ontology.kn_object_type_api import ObjectTypeApi
from src.common.business_util import random_data_properties
from test_case.ontology_base_test import BaseTest


@allure.feature("知识网络关系类管理")
class TestRelationTypeDelete(BaseTest):
    """关系类型创建相关测试"""

    api = RelationTypeApi()
    object_api = ObjectTypeApi()
    mapping_rule = [
        {
            "source_property": {
                "name": "test_string",
            },
            "target_property": {
                "name": "test_string",
            },
        }
    ]

    @classmethod
    @pytest.fixture(autouse=True, scope="class")
    def setup_class(cls, test_kn_id_use):
        """类级别前置方法，使用test_kn_id fixture"""
        cls.test_kn_id_use = test_kn_id_use
        # 创建对象类用来测试关系类型创建
        cls.object_api.create_object_types(
            kn_id=cls.test_kn_id_use,
            object_id="object1",
            name="Object1",
            data_source="",
            data_properties=random_data_properties(static=True),
        )
        cls.object_api.create_object_types(
            kn_id=cls.test_kn_id_use,
            object_id="object2",
            name="Object2",
            data_source="",
            data_properties=random_data_properties(static=True),
        )
        cls.object_api.create_object_types(
            kn_id=cls.test_kn_id_use,
            object_id="object3",
            name="Object3",
            data_source="",
            data_properties=random_data_properties(static=True),
        )

    @allure.story("删除关系类")
    @allure.title("删除关系类")
    def test_delete_relation_type(self, test_kn_id_use):
        """测试删除关系类型"""
        resp = self.api.create_relation_type(
            kn_id=test_kn_id_use,
            id="",
            name="test_delete",
            source_object_type_id="object1",
            target_object_type_id="object2",
            type="direct",
            mapping_rules=self.mapping_rule,
        )
        self.assert_response_code(resp, 201)
        relation_id = json.loads(resp.text)[0].get("id")
        # 删除
        resp = self.api.delete_relation_types(test_kn_id_use, relation_id)
        self.assert_response_code(resp, 204)

        self.assert_db_record_not_exists(
            table=self.relation_table,
            condition="f_id = %s and f_kn_id = %s",
            params=(relation_id, test_kn_id_use),
        )
