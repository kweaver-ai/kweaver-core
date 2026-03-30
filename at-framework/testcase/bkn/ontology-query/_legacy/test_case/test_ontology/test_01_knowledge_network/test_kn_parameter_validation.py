import pytest
import allure
from src.api.ontology.kn_network_api import KnowledgeNetworkApi
from src.common.read_data import data_reader
from test_case.ontology_base_test import BaseTest

validation_data = data_reader.read_yaml(
    "ontology/knowledge_network/kn_parameter_validation.yaml"
)


@allure.feature("业务知识网络")
class TestKnowledgeNetworkParameterValidation(BaseTest):
    api = KnowledgeNetworkApi()

    @allure.story("创建业务知识网络-参数校验")
    @allure.title("{case_name}")
    @pytest.mark.parametrize("test_data", validation_data)
    def test_kn_parameter_validation(self, test_data):
        """参数校验测试用例：验证必填参数、参数格式、参数类型等"""
        test_data = test_data[0]
        case_name = test_data.get("case_name")
        kn_id = test_data.get("kn_id")
        kn_name = test_data.get("kn_name")
        test_type = test_data.get("test_type")
        exclude_fields = test_data.get("exclude_fields", [])
        expected_status = test_data.get("expected_status", 201)
        should_exist_in_db = test_data.get("should_exist_in_db", True)
        note = test_data.get("note", "")

        with allure.step(f"执行参数校验测试: {case_name}"):
            if note:
                allure.attach(note, name="测试说明")

            tags = test_data.get("tags")
            comment = test_data.get("comment")

            resp = self.api.create_kn(
                kn_name=kn_name,
                kn_id=kn_id,
                tags=tags,
                comment=comment,
                exclude_fields=exclude_fields,
            )

        self.assert_response_code(resp, expected_status)

        if test_type == "positive":
            self.assert_response_structure(resp, ["id"])

            if should_exist_in_db:
                created_id = resp.json().get("id")
                self.assert_db_record_exists(
                    table=self.kn_table,
                    params=(created_id,),
                )
        else:
            self.assert_response_structure(resp, ["error_code", "description"])

            if not should_exist_in_db:
                self.assert_db_record_not_exists(
                    table=self.kn_table,
                    params=(kn_id,),
                )
