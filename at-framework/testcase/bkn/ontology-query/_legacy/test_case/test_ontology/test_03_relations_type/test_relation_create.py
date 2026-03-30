# coding: utf-8
"""
关系类型创建测试用例
"""
import pytest
import allure
import json
from src.api.ontology.kn_relation_type_api import RelationTypeApi
from src.api.ontology.kn_object_type_api import ObjectTypeApi
from src.api.ontology.kn_network_api import KnowledgeNetworkApi
from src.common.business_util import random_data_properties
from src.common.read_data import data_reader
from src.common.logger import logger
from test_case.ontology_base_test import BaseTest

# 读取测试数据
positive_data = data_reader.read_yaml(
    "ontology/relation_type/relation_create_postive.yaml"
)
negative_data = data_reader.read_yaml(
    "ontology/relation_type/relation_create_negative.yaml"
)


@allure.feature("知识网络关系类管理")
class TestRelationCreate(BaseTest):
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
        yield
        # 清理业务知识网络```
        KnowledgeNetworkApi().delete_kn(cls.test_kn_id_use)

    @allure.story("创建关系类-参数校验")
    @allure.title("{relation_data[case_name]}")
    @pytest.mark.parametrize("relation_data", positive_data)
    def test_create_relation_type_positive(self, relation_data, test_kn_id_use):
        """测试创建关系类型的正向场景"""

        resp = self.api.create_relation_type(
            kn_id=test_kn_id_use,
            id=relation_data["id"],
            name=relation_data["name"],
            source_object_type_id=relation_data.get("source_object_type_id", "object1"),
            target_object_type_id=relation_data.get("target_object_type_id", "object2"),
            type=relation_data.get("type", "direct"),
            mapping_rules=relation_data.get("mapping_rules", self.mapping_rule),
            tags=relation_data.get("tags", None),
            comment=relation_data.get("comment", None),
        )

        # 验证响应状态码和结构
        self.assert_response_code(resp, 201)
        self.assert_response_structure_list(resp, ["id"])

        # 验证数据库记录存在
        self.assert_db_record_exists(
            self.relation_table,
            condition="f_id = %s AND f_kn_id = %s",
            params=(
                json.loads(resp.text)[0].get("id"),
                test_kn_id_use,
            ),
        )

    @allure.story("创建关系类-参数校验-异常场景")
    @allure.title("{relation_data[case_name]}")
    @pytest.mark.parametrize("relation_data", negative_data)
    def test_create_relation_type_negative(self, relation_data, test_kn_id_use):
        """测试创建关系类型的负向场景"""

        resp = self.api.create_relation_type(
            kn_id=test_kn_id_use,
            id=relation_data["id"],
            name=relation_data["name"],
            source_object_type_id=relation_data.get("source_object_type_id", "object1"),
            target_object_type_id=relation_data.get("target_object_type_id", "object2"),
            type=relation_data.get("type", "direct"),
            mapping_rules=relation_data.get("mapping_rules", self.mapping_rule),
            tags=relation_data.get("tags", None),
            comment=relation_data.get("comment", None),
            module_type=relation_data.get("module_type", None),
            exclude_fields=relation_data.get("exclude_fields", None),
        )

        # 验证响应状态码和结构
        self.assert_response_code(resp, relation_data.get("expected_status"))
        # self.assert_response_structure_list(resp, ["id"])

        # 验证数据库记录不存在
        self.assert_db_record_not_exists(
            self.relation_table,
            condition="f_id = %s AND f_kn_id = %s",
            params=(relation_data["id"], test_kn_id_use),
        )
