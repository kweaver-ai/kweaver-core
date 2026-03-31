# -*- coding:UTF-8 -*-

import allure
import pytest
import random
import string
import uuid

from common.get_content import GetContent
from lib.tool_box import ToolBox

box_id = str(uuid.uuid4())

@allure.feature("工具注册与管理接口测试：内置工具")
class TestBuiltin:
    
    client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 200

    @allure.title("创建内置工具，传参正确，创建成功，默认工具箱状态为published，工具状态为enabled，类型为system")
    def test_Builtin_01(self, Headers):
        box_id = str(uuid.uuid4())
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 200
        assert box_id == result[1]["box_id"]
        assert name == result[1]["box_name"]
        tools = result[1]["tools"]
        for tool in tools:
            assert tool["status"] == "enabled"

        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"
        assert result[1]["is_internal"] == True
        assert result[1]["category_type"] == "system"

    @allure.title("更新内置工具，传参正确，更新成功")
    def test_Builtin_02(self, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        print(box_id, name)
        
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description update",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 200

        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id
        assert result[1]["box_name"] == name
        assert result[1]["box_desc"] == "test description update"
        assert result[1]["category_type"] == "system"

    @allure.title("更新内置工具，名称不合法，更新失败")
    @pytest.mark.parametrize("name", ["invalid: ~!@#$%^&*()_+{}|:'<>?.,《》？：“{}|【】、·-=——`", "invalid name: more then 50 characters, aaaaaaaaaaaa"])
    def test_Builtin_03(self, name, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description update",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 400

    @allure.title("更新内置工具，描述不合法，更新失败")
    def test_Builtin_04(self, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "invalid_desc: more then 255 characters, aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 400

    @allure.title("更新内置工具，版本格式错误，更新失败")
    @pytest.mark.parametrize("version", ["v1", "first_version"])
    def test_Builtin_06(self, version, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description update",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": version,
            "config_source": "auto"
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 400

    @allure.title("创建内置工具，存在同名已发布工具箱，创建失败")
    def test_Builtin_07(self, Headers):
        box_id = str(uuid.uuid4())
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 200
        box_id1 = result[1]["box_id"]
        update_data = {
            "status": "published"
        }
        result = self.client.UpdateToolboxStatus(box_id1, update_data, Headers)
        assert result[0] == 200

        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description update",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 400

    @allure.title("创建内置工具，工具存在同名，创建失败")
    def test_Builtin_08(self, Headers):
        filepath = "./resource/openapi/compliant/duplicate_tool.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 400

    @allure.title("手动更新内置工具，不加保护锁，自动更新后手动更新内容被覆盖")
    def test_Builtin_09(self, Headers):
        global box_id

        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        len(result[1]["tools"]) > 1

        # 手动更新内置工具，不加保护锁
        filepath = "./resource/openapi/compliant/test3.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        
        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description update 1",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "manual",
            "protected_flag": False
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 200
        # 验证更新后结果
        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id
        assert result[1]["box_name"] == name
        assert result[1]["box_desc"] == "test description update 1"
        assert len(result[1]["tools"]) == 1

        # 自动更新内置工具
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name1 = ''.join(random.choice(string.ascii_letters) for i in range(8))
        
        data = {
            "box_id": box_id,
            "box_name": name1,
            "box_desc": "test description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto",
            "protected_flag": False
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 200

        # 验证更新后结果，不保留手动更新内容
        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id
        assert result[1]["box_name"] == name1
        assert result[1]["box_desc"] == "test description"
        assert len(result[1]["tools"]) > 1

    @allure.title("手动更新内置工具，添加保护锁，自动更新后手动更新内容被保留")
    def test_Builtin_10(self, Headers):
        global box_id

        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        len(result[1]["tools"]) > 1

        # 手动更新内置工具，加保护锁
        filepath = "./resource/openapi/compliant/test3.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))

        data = {
            "box_id": box_id,
            "box_name": name,
            "box_desc": "test description update 1",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "manual",
            "protected_flag": True
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 200
        # 验证更新后结果
        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id
        assert result[1]["box_name"] == name
        assert result[1]["box_desc"] == "test description update 1"
        assert len(result[1]["tools"]) == 1

        # 自动更新内置工具
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name1 = ''.join(random.choice(string.ascii_letters) for i in range(8))
        
        data = {
            "box_id": box_id,
            "box_name": name1,
            "box_desc": "test description",
            "data": yaml_data,
            "metadata_type": "openapi",
            "source": "internal",
            "config_version": "1.0.0",
            "config_source": "auto",
            "protected_flag": False
        }
        result = self.client.Builtin(data, Headers)
        assert result[0] == 200

        # 验证更新后结果，保留手动更新内容
        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id
        assert result[1]["box_name"] == name
        assert result[1]["box_desc"] == "test description update 1"
        assert len(result[1]["tools"]) == 1
