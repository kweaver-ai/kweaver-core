import pytest
import allure
import time
from src.api.ontology.kn_network_api import KnowledgeNetworkApi
from src.common.read_data import data_reader
from test_case.ontology_base_test import BaseTest

validation_data = data_reader.read_yaml(
    "ontology/knowledge_network/kn_update_parameter_validation.yaml"
)


@allure.feature("业务知识网络")
class TestKnowledgeNetworkUpdateParameterValidation(BaseTest):
    api = KnowledgeNetworkApi()

    @allure.story("更新业务知识网络-参数校验")
    @allure.title("{case_name}")
    @pytest.mark.parametrize("test_data", validation_data)
    def test_kn_update_parameter_validation(self, test_data):
        """参数校验测试用例：验证更新接口的必填参数、参数格式、参数类型等"""
        test_data = test_data[0]
        case_name = test_data.get("case_name")
        kn_c_name = test_data.get("kn_c_name")
        kn_u_name = test_data.get("kn_u_name")
        kn_id = test_data.get("kn_id")
        test_type = test_data.get("test_type")
        exclude_fields = test_data.get("exclude_fields", [])
        expected_status = test_data.get("expected_status", 204)
        note = test_data.get("note", "")
        tags = test_data.get("tags")
        comment = test_data.get("comment")

        with allure.step("步骤1: 创建业务知识网络"):
            create_resp = self.api.create_kn(kn_name=kn_c_name, kn_id=kn_id)
            self.assert_response_code(create_resp, 201)

        with allure.step(f"步骤2: 执行参数校验测试 - {case_name}"):
            if note:
                allure.attach(note, name="测试说明")

            update_resp = self.api.update_kn(
                kn_name=kn_u_name,
                kn_id=kn_id,
                tags=tags,
                comment=comment,
                exclude_fields=exclude_fields,
            )

        with allure.step(f"步骤3: 验证HTTP响应状态码: 期望 {expected_status}"):
            self.assert_response_code(update_resp, expected_status)

        if test_type == "positive":
            with allure.step("步骤4: 验证数据库记录已更新"):
                self.assert_db_record_exists(
                    table=self.kn_table,
                    condition="f_name = %s",
                    params=(kn_u_name,),
                )
        else:
            with allure.step("步骤4: 验证数据库记录未被修改"):
                self.assert_db_record_exists(
                    table=self.kn_table,
                    condition="f_name = %s",
                    params=(kn_c_name,),
                )

        with allure.step("步骤5: 清理测试数据"):
            self.api.delete_kn(kn_id)
