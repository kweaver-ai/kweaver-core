# -*- coding:UTF-8 -*-

import allure
import uuid
import pytest

from common.get_content import GetContent
from lib.operator import Operator

operator_list = []

@allure.feature("算子注册与管理接口测试：编辑算子")
class TestEditOperator:
    """仅在编辑metadata时生成新版本"""
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global operator_list

        filepath = "./resource/openapi/compliant/relations.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        operators = re[1]

        for operator in operators:
            if operator["status"] == "success":
                op = {
                    "operator_id": operator["operator_id"],
                    "version": operator["version"]
                }
                operator_list.append(op)
                update_data = [
                    {
                        "operator_id": operator["operator_id"],
                        "status": "published"
                    }
                ] 
                result = self.client.UpdateOperatorStatus(update_data, Headers)
                assert result[0] == 200

    @allure.title("修改已发布算子名称，编辑成功，不生成新版本，状态为已发布编辑中")
    def test_edit_operator_01(self, Headers):
        global operator_list
        
        data = {
            "operator_id": operator_list[0]["operator_id"],
            "name": "test_edit"
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["operator_id"] == operator_list[0]["operator_id"]
        assert re[1]["status"] == "editing"
        assert re[1]["version"] == operator_list[0]["version"]

        # 再次编辑
        data = {
            "operator_id": operator_list[0]["operator_id"],
            "extend_info": {
                "option": "test"
            }
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["operator_id"] == operator_list[0]["operator_id"]
        assert re[1]["status"] == "editing"
        assert re[1]["version"] == operator_list[0]["version"]

        update_data = [
            {
                "operator_id": operator_list[0]["operator_id"],
                "status": "published"
            }
        ]
        re = self.client.UpdateOperatorStatus(update_data, Headers)     # 发布算子
        assert re[0] == 200

    @allure.title("算子名称超过50字符，编辑失败")
    def test_edit_operator_02(self, Headers):
        global operator_list
        data = {
            "name": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
            "operator_id": operator_list[0]["operator_id"]
        }
        re = self.client.EditOperator(data, Headers)    # name有51个字符
        assert re[0] == 400

    @allure.title("算子不存在，编辑失败")
    def test_edit_operator_03(self, Headers):
        global versions
        id = uuid.uuid4()
        data = {
            "operator_id": str(id)
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 404

    @allure.title("修改算子描述，编辑成功，生成新版本")
    def test_edit_operator_04(self, Headers):
        global operator_list
        
        re = self.client.GetOperatorInfo(operator_list[0]["operator_id"], Headers)
        assert re[0] == 200
        data = {
            "operator_id": operator_list[0]["operator_id"],
            "description": "test edit 1234567"
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["operator_id"] == operator_list[0]["operator_id"]
        assert re[1]["status"] == "editing"
        assert re[1]["version"] != operator_list[0]["version"]

    @allure.title("编辑已发布算子信息，编辑成功，无新版本生成")
    def test_edit_operator_05(self, Headers):
        global operator_list
        global versions
        data = {
            "operator_id": operator_list[1]["operator_id"],
            "operator_info": {
                "operator_type": "composite",
                "execution_mode": "async",
                "category": "data_processing"
            }
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "editing"
        assert re[1]["version"] == operator_list[1]["version"]

    @allure.title("编辑已下架算子，编辑成功，编辑后状态为unpublish")
    def test_edit_operator_06(self, Headers):
        filepath = "./resource/openapi/compliant/edit-test2.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        operators = re[1]
        update_data = [
            {
                "operator_id": operators[0]["operator_id"],
                "status": "offline"
            }
        ]
        re = self.client.UpdateOperatorStatus(update_data, Headers)     # 下架算子
        assert re[0] == 200

        data = {
            "operator_id": operators[0]["operator_id"]
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "unpublish"
        assert re[1]["version"] == operators[0]["version"]

    @allure.title("编辑未发布算子，编辑成功，无新版本生成")
    def test_edit_operator_07(self, Headers):
        filepath = "./resource/openapi/compliant/test3.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        operators = re[1]
        data = {
            "name": "edit_test",
            "operator_id": operators[0]["operator_id"]
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "unpublish"
        assert re[1]["version"] == operators[0]["version"]

    @allure.title("算子名称包含特殊字符，编辑失败")
    @pytest.mark.parametrize("name", ["invalid name","name~","name@","name`","name#","name$","name%","name^","name^","name&", 
                                    "name*","name()","name-","name+","name=","name[]","name{}","name|","name\\","name:",
                                    "name;","name'","name,","name.","name?","name/","name<","name>","name；","name“","name：",
                                    "name’","name【】","name《","name》","name？","name·","name、","name，","name。"])   
    def test_edit_operator_08(self, name, Headers):
        filepath = "./resource/openapi/compliant/edit-test1.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        operators = re[1]
        data = {
            "name": name,
            "operator_id": operators[0]["operator_id"]
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 400

    @allure.title("算子描述超过255个字符，编辑失败")
    def test_edit_operator_09(self, Headers):
        filepath = "./resource/openapi/compliant/edit-test2.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        operator_data = re[1]
        
        re = self.client.GetOperatorInfo(operator_data[0]["operator_id"], Headers)
        assert re[0] == 200
        data = {
            "operator_id": operator_data[0]["operator_id"],
            "description": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 400

    @allure.title("编辑算子执行控制，编辑成功")
    def test_edit_operator_10(self, Headers):
        filepath = "./resource/openapi/compliant/test3.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        operators = re[1]
        data = {
            "operator_id": operators[0]["operator_id"],
            "operator_execute_control": {
                "timeout": 90000,
                "retry_policy": {
                    "max_attempts": 5,
                    "initial_delay": 100,
                    "max_delay": 9000,
                    "backoff_factor": 3,
                    "retry_conditions": {
                        "status_code": [
                            500,
                            501
                        ],
                        "error_codes": [
                            "500",
                            "501"
                        ]
                    }
                }
            }
            
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "unpublish"
        assert re[1]["version"] == operators[0]["version"]

    @allure.title("编辑算子扩展信息，编辑成功")
    def test_edit_operator_11(self, Headers):
        filepath = "./resource/openapi/compliant/template.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        operators = re[1]
        data = {
            "operator_id": operators[0]["operator_id"],
            "extend_info": {
                "custom_info": {
                    "key1": "value1",
                    "key2": "value2"
                }
            }
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "unpublish"
        assert re[1]["version"] == operators[0]["version"]

    @allure.title("编辑算子信息，执行模式为异步，标识为数据源算子，编辑失败")
    def test_edit_operator_12(self, Headers):
        global operator_list
        global versions
        data = {
            "operator_id": operator_list[1]["operator_id"],
            "operator_info": {
                "operator_type": "composite",
                "execution_mode": "async",
                "category": "data_processing",
                "is_data_source": True
            }
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 400

    @allure.title("编辑算子信息，执行模式为同步，标识为数据源算子，编辑成功")
    def test_edit_operator_13(self, Headers):
        global operator_list
        global versions
        data = {
            "operator_id": operator_list[1]["operator_id"],
            "operator_info": {
                "operator_type": "composite",
                "execution_mode": "sync",
                "category": "data_processing",
                "is_data_source": True
            }
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200

        result = self.client.GetOperatorInfo(operator_list[1]["operator_id"], Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_list[1]["operator_id"]
        operator_info = result[1]["operator_info"]
        assert operator_info["operator_type"] == "composite"
        assert operator_info["execution_mode"] == "sync"
        assert operator_info["category"] == "data_processing"
        assert operator_info["is_data_source"] == True

    @allure.title("更新算子元数据，metadata_type为其他类型，编辑失败")
    def test_edit_operator_14(self, Headers):
        global operator_list
        data = {
            "operator_id": operator_list[0]["operator_id"],
            "metadata_type": "operator"
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 400

    @allure.title("更新算子元数据，未匹配到当前算子，编辑失败")
    def test_edit_operator_15(self, Headers):
        global operator_list
        filepath = "./resource/openapi/compliant/template.yaml"
        operator_data = GetContent(filepath).yamlfile()
        data = {
            "operator_id": operator_list[0]["operator_id"],
            "metadata_type": "openapi",
            "data": operator_data
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 404

    @allure.title("更新算子元数据，openapi中包含多个算子，可匹配到当前算子，编辑成功，生成新版本")
    def test_edit_operator_16(self, Headers):
        global operator_list
        filepath = "./resource/openapi/compliant/relations.yaml"
        operator_data = GetContent(filepath).yamlfile()
        data = {
            "operator_id": operator_list[0]["operator_id"],
            "metadata_type": "openapi",
            "data": operator_data
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["operator_id"] == operator_list[0]["operator_id"]
        assert re[1]["status"] == "editing"
        assert re[1]["version"] != operator_list[0]["version"]

    @allure.title("更新算子元数据，openapi中仅包含一个算子，且匹配到当前算子，编辑成功，生成新版本")
    def test_edit_operator_17(self, Headers):
        filepath = "./resource/openapi/compliant/template.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        operators = re[1]
        data = {
            "operator_id": operators[0]["operator_id"],
            "metadata_type": "openapi",
            "data": operator_data
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["operator_id"] == operators[0]["operator_id"]
        assert re[1]["status"] == "editing"
        assert re[1]["version"] != operators[0]["version"]
