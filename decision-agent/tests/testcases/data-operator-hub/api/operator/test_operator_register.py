# -*- coding:UTF-8 -*-

import allure

from common.get_content import GetContent
from lib.operator import Operator


@allure.feature("算子注册与管理接口测试：算子注册")
class TestOperatorRegister:
    '''
    批量注册算子，不允许直接发布
    未发布算子不校验是否重复
    '''
    client = Operator()

    @allure.title("注册算子，direct_publish默认为false，注册成功，算子状态为未发布，默认执行模式为同步，为非数据源算子")
    def test_register_operator_01(self, Headers):
        filepath = "./resource/openapi/compliant/test0.json"
        api_data = GetContent(filepath).jsonfile()

        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }

        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        operators = result[1]
        for operator in operators:
            assert operator["status"] == "success"
            id = operator["operator_id"]

            data = {
                "page_size": -1,
                "status": "unpublish"
            }
            result = self.client.GetOperatorList(data, Headers)
            unpublished_ops = result[1]["data"]
            for unpublished_op in unpublished_ops:
                if unpublished_op["operator_id"] == id:
                    assert unpublished_op["status"] == "unpublish"
                    assert unpublished_op["operator_info"]["execution_mode"] == "sync"
                    assert unpublished_op["operator_info"]["is_data_source"] == False

    @allure.title("注册单个算子，direct_publish为true，注册成功，算子状态为已发布")
    def test_register_operator_02(self, Headers):
        filepath = "./resource/openapi/compliant/test1.json"
        api_data = GetContent(filepath).jsonfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        operators = result[1]
        for operator in operators:
            assert operator["status"] == "success"
            id = operator["operator_id"]
            result = self.client.GetOperatorInfo(id, Headers)
            assert result[1]["status"] == "published"

    @allure.title("注册多个算子，direct_publish为true，注册失败")
    def test_register_operator_03(self, Headers):
        filepath = "./resource/openapi/compliant/test0.json"
        api_data = GetContent(filepath).jsonfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 400

    @allure.title("符合算子导入规范，数据格式为yaml，注册算子，注册成功")
    def test_register_operator_04(self, Headers):
        filepath = "./resource/openapi/compliant/test2.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        for item in result[1]:
            assert item["status"] == "success"

    @allure.title("存在同名的已发布算子时，注册同名算子失败")
    def test_register_operator_05(self, Headers):
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }

        result = self.client.RegisterOperator(data, Headers)    # 注册算子，直接发布
        assert result[0] == 200
        assert result[1][0]["status"] == "success"

        result = self.client.RegisterOperator(data, Headers)    # 再次注册，算子已被发布
        # print(result)
        assert result[0] == 200
        assert result[1][0]["status"] == "failed"

    @allure.title("注册算子，算子名称超过50个字符，注册失败")
    def test_register_operator_06(self, Headers):
        filepath = "./resource/openapi/non-compliant/long_name.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        assert result[1][0]["status"] == "failed"

    @allure.title("注册算子，算子描述超过255个字符，注册失败")
    def test_register_operator_07(self, Headers):
        filepath = "./resource/openapi/non-compliant/long_desc.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        assert result[1][0]["status"] == "failed"

    @allure.title("openapi超出大小限制，注册算子，注册失败")
    def test_register_operator_08(self, Headers):
        filepath = "./resource/openapi/compliant/large_api.json"
        api_data = GetContent(filepath).jsonfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        # print(result)
        assert result[0] == 400

    @allure.title("运行模式为异步，注册数据源算子，注册失败")
    def test_register_operator_09(self, Headers):
        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "operator_info": {
                "execution_mode": "async",
                "is_data_source": True
            }
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 400

    @allure.title("运行模式为同步，注册数据源算子，注册成功")
    def test_register_operator_10(self, Headers):
        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "operator_info": {
                "execution_mode": "sync",
                "is_data_source": True
            }
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        assert result[1][0]["status"] == "success"