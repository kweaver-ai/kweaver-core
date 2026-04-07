# -*- coding:UTF-8 -*-

import allure
import string
import random
import uuid
import pytest

from common.get_content import GetContent
from lib.tool_box import ToolBox
from lib.operator import Operator

box_id = ""
operator_id = ""
operator_version = ""

@allure.feature("工具注册与管理接口测试：算子转换成工具")
class TestConvertOperator:
    
    client = ToolBox()
    client1 = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_id
        global operator_id
        global operator_version

        # 创建工具箱
        filepath = "./resource/openapi/compliant/test.json"
        json_data = GetContent(filepath).jsonfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": json_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]

        # 注册算子
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()

        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }

        result = self.client1.RegisterOperator(data, Headers)
        assert result[0] == 200
        operators = result[1]
        for operator in operators:
            assert operator["status"] == "success"
            operator_id = operator["operator_id"]
            operator_version = operator["version"]

    @allure.title("算子转换成工具，传参正确，算子未发布，转换失败")
    def test_convert_operator_01(self, Headers):
        global box_id
        global operator_id
        global operator_version

        # 转换算子为工具
        convert_data = {
            "box_id": box_id,
            "operator_id": operator_id,
            "operator_version": operator_version
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 404

    @allure.title("算子转换成工具，传参正确，算子已发布，转换成功")
    def test_convert_operator_02(self, Headers):
        global box_id
        global operator_id
        global operator_version

        update_data = [
            {
                "operator_id": operator_id,
                "version": operator_version,
                "status": "published"
            }
        ] 
        result = self.client1.UpdateOperatorStatus(update_data, Headers)
        assert result[0] == 200

        # 转换算子为工具
        convert_data = {
            "box_id": box_id,
            "operator_id": operator_id,
            "operator_version": operator_version
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 200
        assert "tool_id" in result[1]

    @allure.title("算子转换成工具，工具箱不存在，转换失败")
    def test_convert_operator_03(self, Headers):
        global operator_id
        global operator_version

        box_id = str(uuid.uuid4())
        convert_data = {
            "box_id": box_id,
            "operator_id": operator_id,
            "operator_version": operator_version
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 400

    @allure.title("算子转换成工具，算子不存在，转换失败")
    def test_convert_operator_04(self, Headers):
        global box_id
        global operator_version

        # 转换不存在的算子
        operator_id = str(uuid.uuid4())
        convert_data = {
            "box_id": box_id,
            "operator_id": operator_id,
            "operator_version": operator_version
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 404

    @allure.title("算子转换成工具，必填参数operator_version不传，转换失败")
    def test_convert_operator_05(self, Headers):
        global box_id
        global operator_id

        # 转换算子，不传必填参数
        convert_data = {
            "box_id": box_id,
            "operator_id": operator_id
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 400 

    @allure.title("算子转换成工具，必填参数operator_id不传，转换失败")
    def test_convert_operator_06(self, Headers):
        global box_id
        global operator_version

        # 转换算子，不传必填参数
        convert_data = {
            "box_id": box_id,
            "operator_version": operator_version
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 400 

    @allure.title("算子转换成工具，必填参数box_id不传，转换失败")
    def test_convert_operator_07(self, Headers):
        global operator_id
        global operator_version

        # 转换算子，不传必填参数
        convert_data = {
            "operator_id": operator_id,
            "operator_version": operator_version
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 400 

    @allure.title("算子转换成工具，传参正确，算子已下架，转换失败")
    def test_convert_operator_08(self, Headers):
        global box_id
        global operator_id
        global operator_version

        update_data = [
            {
                "operator_id": operator_id,
                "version": operator_version,
                "status": "offline"
            }
        ] 
        result = self.client1.UpdateOperatorStatus(update_data, Headers)
        assert result[0] == 200

        # 转换算子为工具
        convert_data = {
            "box_id": box_id,
            "operator_id": operator_id,
            "operator_version": operator_version
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 400