# -*- coding:UTF-8 -*-

import allure
import uuid
import string
import random
import pytest

from common.get_content import GetContent
from lib.operator import Operator

operator_id = str(uuid.uuid4())
name = ''.join(random.choice(string.ascii_letters) for i in range(8))

@allure.feature("算子注册与管理接口测试：内置算子注册与更新")
class TestBuiltinOperator:
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global operator_id, name
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }

        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 200

    @allure.title("注册内置算子，注册成功，算子状态为已发布，类型为system")
    def test_builtin_operator_01(self, Headers):
        operator_id = str(uuid.uuid4())
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }

        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        assert result[1]["status"] == "success"
        assert "version" in result[1]

        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        assert result[1]["version"] == result[1]["version"]
        assert result[1]["status"] == "published"
        assert result[1]["is_internal"] == True
        assert result[1]["operator_info"]["category"] == "system"

    @allure.title("批量注册多个内置算子，注册失败")
    def test_builtin_operator_02(self, Headers):
        filepath = "./resource/openapi/compliant/test0.json"
        api_data = GetContent(filepath).jsonfile()
        operator_id = str(uuid.uuid4())
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }

        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 400

    @allure.title("更新内置算子，传参正确，更新成功")
    def test_builtin_operator_03(self, Headers):
        global operator_id, name
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "composite",
            "execution_mode": "async",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        
        info_result = self.client.GetOperatorInfo(operator_id, Headers)
        assert info_result[0] == 200
        assert info_result[1]["name"] == name
        assert info_result[1]["operator_info"]["operator_type"] == "composite"
        assert info_result[1]["operator_info"]["execution_mode"] == "async"

    @allure.title("更新内置算子，名称不合法，更新失败")
    @pytest.mark.parametrize("name", ["invalid: ~!@#$%^&*()_+{}|:'<>?.,《》？：{}|【】、·-=——`", "invalid name: more then 50 characters, aaaaaaaaaaaa"])
    def test_builtin_operator_04(self, name, Headers):
        global operator_id
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 400
    
    @allure.title("更新内置工具，描述不合法，更新失败")
    def test_builtin_operator_05(self, Headers):
        global operator_id
        filepath = "./resource/openapi/non-compliant/long_desc.yaml"
        api_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 400

    @allure.title("更新内置算子，算子类型无效，更新失败")
    def test_builtin_operator_06(self, Headers):
        global operator_id
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "invalid_type",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 400

    @allure.title("更新内置算子，算子执行模式无效，更新失败")
    def test_builtin_operator_08(self, Headers):
        global operator_id
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "invalid_mode",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 400

    @allure.title("更新内置工具，版本格式错误，更新失败")
    @pytest.mark.parametrize("version", ["v1", "first_version"])
    def test_builtin_operator_09(self, version, Headers):
        global operator_id
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "async",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        data = {
            "operator_id": operator_id,
            "config_version": version
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 400

    @allure.title("注册内置算子，同名算子已存在，注册失败")
    def test_builtin_operator_10(self, Headers):
        global name
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        operator_id = str(uuid.uuid4())

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "composite",
            "execution_mode": "async",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0"
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 409

    @allure.title("手动更新内置算子，不加保护锁，自动更新后手动更新内容被覆盖")
    def test_builtin_operator_11(self, Headers):
        global operator_id
        # 手动更新内置算子，不加保护锁
        filepath = "./resource/openapi/compliant/edit-test1.yaml"
        api_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "composite",
            "execution_mode": "async",
            "source": "intenal",
            "config_source": "manual",
            "config_version": "1.0.0",
            "protected_flag": False
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        # 验证更新后算子信息
        info_result = self.client.GetOperatorInfo(operator_id, Headers)
        assert info_result[0] == 200
        assert info_result[1]["name"] == name
        assert info_result[1]["operator_info"]["operator_type"] == "composite"
        assert info_result[1]["operator_info"]["execution_mode"] == "async"
        summary = info_result[1]["metadata"]["summary"]

        # 自动更新
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        name1 = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "operator_id": operator_id,
            "name": name1,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0",
            "protected_flag": False
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        # 验证更新后算子信息，自动更新后手动更新内容不保留
        info_result = self.client.GetOperatorInfo(operator_id, Headers)
        assert info_result[0] == 200
        assert info_result[1]["name"] == name1
        assert info_result[1]["operator_info"]["operator_type"] == "basic"
        assert info_result[1]["operator_info"]["execution_mode"] == "sync"
        summary1 = info_result[1]["metadata"]["summary"]

        assert summary != summary1

    @allure.title("手动更新内置算子，加保护锁，自动更新手动更新内容被保留")
    def test_builtin_operator_12(self, Headers):
        global operator_id
        # 手动更新内置算子，加保护锁
        filepath = "./resource/openapi/compliant/edit-test1.yaml"
        api_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "composite",
            "execution_mode": "async",
            "source": "intenal",
            "config_source": "manual",
            "config_version": "1.0.0",
            "protected_flag": True
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        # 验证更新后算子信息
        info_result = self.client.GetOperatorInfo(operator_id, Headers)
        assert info_result[0] == 200
        assert info_result[1]["name"] == name
        assert info_result[1]["operator_info"]["operator_type"] == "composite"
        assert info_result[1]["operator_info"]["execution_mode"] == "async"
        summary = info_result[1]["metadata"]["summary"]

        # 自动更新
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        name1 = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "operator_id": operator_id,
            "name": name1,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0",
            "protected_flag": True
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        # 验证更新后算子信息，自动更新后手动更新内容保留
        info_result = self.client.GetOperatorInfo(operator_id, Headers)
        assert info_result[0] == 200
        assert info_result[1]["name"] == name
        assert info_result[1]["operator_info"]["operator_type"] == "composite"
        assert info_result[1]["operator_info"]["execution_mode"] == "async"
        summary1 = info_result[1]["metadata"]["summary"]

        assert summary == summary1

    @allure.title("更新内置算子，算子执行模式为同步，设置为数据源算子，更新成功")
    def test_builtin_operator_13(self, Headers):
        global operator_id
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "sync",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0",
            "is_data_source": True
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 200

    @allure.title("更新内置算子，算子执行模式为异步，设置为数据源算子，更新失败")
    def test_builtin_operator_14(self, Headers):
        global operator_id
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "operator_id": operator_id,
            "name": name,
            "data": api_data,
            "metadata_type": "openapi",
            "operator_type": "basic",
            "execution_mode": "async",
            "source": "intenal",
            "config_source": "auto",
            "config_version": "1.0.0",
            "is_data_source": True
        }
        result = self.client.RegisterBuiltinOperator(data, Headers)
        assert result[0] == 400
