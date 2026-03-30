# coding: utf-8
"""
关系类型修改测试用例
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
    "ontology/relation_type/relation_update_postive.yaml"
)
negative_data = data_reader.read_yaml(
    "ontology/relation_type/relation_update_negative.yaml"
)


@allure.feature("知识网络关系类管理")
class TestRelationUpdate(BaseTest):
    """关系类型修改相关测试"""

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

    @allure.story("更新业务知识网络关系类-正常场景")
    @allure.title("{relation_data[case_name]}")
    @pytest.mark.parametrize("relation_data", positive_data)
    def test_update_relation_success(self, relation_data, test_kn_id_use):
        """测试更新关系类型的正向场景"""
        # 创建关系类
        resp = self.api.create_relation_type(
            kn_id=test_kn_id_use,
            id="",
            name=relation_data["relation_c_name"],
            source_object_type_id="object1",
            target_object_type_id="object2",
            type="direct",
            mapping_rules=self.mapping_rule,
        )
        self.assert_response_code(resp, 201)
        relation_id = json.loads(resp.text)[0].get("id")
        # 更新关系类
        resp = self.api.update_relation_type(
            kn_id=test_kn_id_use,
            id=relation_id,
            name=relation_data["relation_u_name"],
            source_object_type_id=relation_data.get("source_object_type_id", "object1"),
            target_object_type_id=relation_data.get("target_object_type_id", "object2"),
            type=relation_data.get("type", "direct"),
            mapping_rules=relation_data.get("mapping_rules", self.mapping_rule),
            tags=relation_data.get("tags", None),
            comment=relation_data.get("comment", None),
        )

        # 验证响应状态码
        self.assert_response_code(resp, 204)

        # 验证数据库记录存在
        self.assert_db_record_exists(
            self.relation_table,
            condition="f_id = %s AND f_kn_id = %s",
            params=(relation_id, test_kn_id_use),
        )

    @allure.story("更新业务知识网络关系类-参数校验异常场景")
    @allure.title("{relation_data[case_name]}")
    @pytest.mark.parametrize("relation_data", negative_data)
    def test_update_relation_negative(self, relation_data, test_kn_id_use):
        """测试更新关系类型的负向场景"""
        # 创建关系类
        resp = self.api.create_relation_type(
            kn_id=test_kn_id_use,
            id=relation_data.get("relation_id", ""),
            name=relation_data["relation_c_name"],
            source_object_type_id="object1",
            target_object_type_id="object2",
            type="direct",
            mapping_rules=self.mapping_rule,
        )
        self.assert_response_code(resp, 201)
        relation_id = json.loads(resp.text)[0].get("id")
        # 更新关系类（使用无效参数）
        resp = self.api.update_relation_type(
            kn_id=test_kn_id_use,
            id=relation_id,
            name=relation_data["relation_u_name"],
            source_object_type_id=relation_data.get("source_object_type_id", "object1"),
            target_object_type_id=relation_data.get("target_object_type_id", "object2"),
            type=relation_data.get("type", "direct"),
            mapping_rules=relation_data.get("mapping_rules", self.mapping_rule),
            tags=relation_data.get("tags", None),
            comment=relation_data.get("comment", None),
            module_type=relation_data.get("module_type", None),
            exclude_fields=relation_data.get("exclude_fields", None),
        )

        # 验证响应状态码
        self.assert_response_code(resp, relation_data["expected_status"])
