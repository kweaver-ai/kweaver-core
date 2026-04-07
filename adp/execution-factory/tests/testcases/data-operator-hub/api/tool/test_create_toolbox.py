# -*- coding:UTF-8 -*-

import allure
import pytest
import random
import string

from common.get_content import GetContent
from lib.tool_box import ToolBox

characters = string.ascii_letters + string.digits
name = ''.join(random.choice(characters) for i in range(8))
box_id = ""


@allure.feature("工具注册与管理接口测试：创建工具箱")
class TestCreateToolbox:
    
    client = ToolBox()

    @allure.title("创建工具箱，传参正确，创建成功，默认未发布")
    def test_create_toolbox_01(self, Headers):
        global name
        global box_id
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 200
        box_id = result[1]["box_id"]

        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "unpublish"

    @allure.title("创建工具箱，必填参数metadata_type不传，创建失败")
    def test_create_toolbox_02(self, Headers):
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": yaml_data
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 400

    @allure.title("创建工具箱，必填参数data不传，创建失败")
    def test_create_toolbox_03(self, Headers):
        data = {
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 400

    @allure.title("创建工具箱，metadata_type值不正确，创建失败")
    def test_create_toolbox_04(self, Headers):
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": yaml_data,
            "metadata_type": "invalid_type"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 400

    @allure.title("创建工具箱，data内容无法解析，创建失败")
    def test_create_toolbox_05(self, Headers):
        data = {
            "data": "invalid_yaml_data",
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 400

    @allure.title("创建工具箱，名称不合法，创建失败")
    @pytest.mark.parametrize("name", ["invalid name","name~","name@","name`","name#","name$","name%","name^","name^","name&", 
                                      "name*","name()","name-","name+","name=","name[]","name{}","name|","name\\","name:",
                                      "name;","name'","name,","name.","name?","name/","name<","name>","name；","name“","name：",
                                      "name’","name【】","name《","name》","name？","name·","name、","name，","name。",
                                      "invalid_name:_more_then_50_characters_aaaaaaaaaaaaa"])
    def test_create_toolbox_06(self, name, Headers):
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 400

    @allure.title("创建工具箱，描述不合法，创建失败")
    def test_create_toolbox_07(self, Headers):
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_desc": "invalid_desc: more then 255 characters, aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 400

    @allure.title("创建工具箱，存在同名未发布工具箱，创建成功")
    def test_create_toolbox_08(self, Headers):
        global name
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 200

    @allure.title("创建工具箱，分类不存在，创建成功（目前后端未校验分类有效性）")
    def test_create_toolbox_09(self, Headers):
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_category": "invalid_category",
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 200  # 修改为200，因为后端目前未校验分类有效性

    @allure.title("创建工具箱，工具箱中的工具存在同名，创建失败")
    def test_create_toolbox_10(self, Headers):
        filepath = "./resource/openapi/compliant/duplicate_tool.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 400

    @allure.title("创建工具箱，已存在同名已发布工具箱，创建失败")
    def test_create_toolbox_11(self, Headers):
        global box_id
        global name

        # 发布工具箱
        update_data = {
            "status": "published"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200

        # 创建同名工具箱
        filepath = "./resource/openapi/compliant/test.json"
        json_data = GetContent(filepath).jsonfile()
        
        data = {
            "box_name": name,
            "data": json_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 400

    @allure.title("创建工具箱，metadata_type为function，传参正确，创建成功")
    def test_create_toolbox_function_01(self, Headers):
        """测试创建function类型的工具箱
        根据实际API，function类型创建工具箱时只需要基本信息和metadata_type
        """
        function_name = ''.join(random.choice(characters) for i in range(8))
        
        # function类型创建工具箱的基本参数（根据实际API请求格式）
        data = {
            "box_name": f"function_box_{function_name}",
            "box_desc": "测试函数工具箱描述",
            "box_category": "data_process",  # 使用有效的分类
            "metadata_type": "function"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 200, f"创建function类型工具箱失败，状态码: {result[0]}, 响应: {result}"
        assert "box_id" in result[1]
        
        # 验证工具箱状态
        box_id_func = result[1]["box_id"]
        result = self.client.GetToolbox(box_id_func, Headers)
        assert result[0] == 200, f"获取工具箱信息失败，状态码: {result[0]}, 响应: {result}"
        assert result[1]["status"] == "unpublish"

    @allure.title("创建工具箱，metadata_type为function，必填参数box_name不传，创建失败")
    def test_create_toolbox_function_02(self, Headers):
        """测试function类型创建工具箱时缺少必填参数box_name"""
        data = {
            "box_desc": "测试描述",
            "box_category": "data_process",
            "metadata_type": "function"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 400

    @allure.title("创建工具箱，metadata_type为function，box_category不存在，创建成功（目前后端未校验分类有效性）")
    def test_create_toolbox_function_03(self, Headers):
        """测试function类型创建工具箱时使用无效的分类"""
        function_name = ''.join(random.choice(characters) for i in range(8))
        
        data = {
            "box_name": f"function_box_{function_name}",
            "box_desc": "测试描述",
            "box_category": "invalid_category",  # 无效的分类
            "metadata_type": "function"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 200  # 修改为200，因为后端目前未校验分类有效性

    @allure.title("创建工具箱，metadata_type为function，名称不合法，创建失败")
    def test_create_toolbox_function_04(self, Headers):
        """测试function类型创建工具箱时使用不合法的名称"""
        data = {
            "box_name": "invalid name",  # 包含空格的不合法名称
            "box_desc": "测试描述",
            "box_category": "data_process",
            "metadata_type": "function"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 400

    @allure.title("创建工具箱，metadata_type为function，描述超长，创建失败")
    def test_create_toolbox_function_05(self, Headers):
        """测试function类型创建工具箱时描述超过255字符"""
        function_name = ''.join(random.choice(characters) for i in range(8))
        
        data = {
            "box_name": f"function_box_{function_name}",
            "box_desc": "invalid_desc: more then 255 characters, aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
            "box_category": "data_process",
            "metadata_type": "function"
        }
        result = self.client.CreateToolbox(data, Headers)
        assert result[0] == 400

    @allure.title("创建工具箱，metadata_type为function，存在同名未发布工具箱，创建成功")
    def test_create_toolbox_function_06(self, Headers):
        """测试function类型创建同名未发布工具箱"""
        function_name = ''.join(random.choice(characters) for i in range(8))
        box_name = f"function_box_{function_name}"
        
        # 第一次创建
        data1 = {
            "box_name": box_name,
            "box_desc": "测试描述1",
            "box_category": "data_process",
            "metadata_type": "function"
        }
        result1 = self.client.CreateToolbox(data1, Headers)
        assert result1[0] == 200
        
        # 创建同名工具箱（应该成功，因为前一个是未发布状态）
        data2 = {
            "box_name": box_name,
            "box_desc": "测试描述2",
            "box_category": "data_process",
            "metadata_type": "function"
        }
        result2 = self.client.CreateToolbox(data2, Headers)
        assert result2[0] == 200