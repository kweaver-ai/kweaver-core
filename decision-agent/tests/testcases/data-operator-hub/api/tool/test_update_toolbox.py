# -*- coding:UTF-8 -*-

import allure
import pytest
import random
import string
import uuid

from common.get_content import GetContent
from lib.tool_box import ToolBox

box_id = ""
characters = string.ascii_letters + string.digits
name = ''.join(random.choice(characters) for i in range(8))

@allure.feature("工具注册与管理接口测试：更新工具箱")
class TestUpdateToolbox:
    
    client = ToolBox()
    
    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_id
        global name
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"  
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]

    @allure.title("更新工具箱，传参正确，更新成功")
    def test_update_toolbox_01(self, Headers):
        global box_id
        global name

        update_data = {
            "box_name": name,
            "box_desc": "test toolbox update description",
            "box_svc_url": "http://test.com",
            "box_icon": "icon-color-tool-FADB14",
            "box_category": "data_process"
        }
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 200

    @allure.title("更新工具箱，工具箱不存在，更新失败")
    def test_update_toolbox_02(self, Headers):
        update_data = {
            "box_name": "test_toolbox_update",
            "box_desc": "test toolbox update description",
            "box_svc_url": "http://test.com",
            "box_icon": "icon-color-tool-FADB14",
            "box_category": "data_process"
        }
        box_id = str(uuid.uuid4())
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("更新工具箱，名称不合法，更新失败")
    @pytest.mark.parametrize("name", ["invalid name","name~","name@","name`","name#","name$","name%","name^","name^","name&", 
                                      "name*","name()","name-","name+","name=","name[]","name{}","name|","name\\","name:",
                                      "name;","name'","name,","name.","name?","name/","name<","name>","name；","name“","name：",
                                      "name’","name【】","name《","name》","name？","name·","name、","name，","name。",
                                      "invalid_name:_more_then_50_characters_aaaaaaaaaaaaa"])    
    def test_update_toolbox_03(self, name, Headers):
        global box_id
        update_data = {
            "box_name": name,
            "box_desc": "test toolbox update description",
            "box_svc_url": "http://test.com",
            "box_icon": "icon-color-tool-FADB14",
            "box_category": "data_process"
        }
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("更新工具箱，描述不合法，更新失败")
    def test_update_toolbox_04(self, Headers):
        global box_id
        update_data = {
            "box_name": "test_toolbox_update",
            "box_desc": "invalid_desc: more then 255 characters, aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
            "box_svc_url": "http://test.com",
            "box_icon": "icon-color-tool-FADB14",
            "box_category": "data_process"
        }
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("更新工具箱，分类不存在，更新失败")
    def test_update_toolbox_05(self, Headers):
        global box_id
        global name
        update_data = {
            "box_name": name,
            "box_desc": "test toolbox update description",
            "box_svc_url": "http://test.com",
            "box_icon": "icon-color-tool-FADB14",
            "box_category": "invalid_category"
        }
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("更新已发布工具箱，更新成功")
    def test_update_toolbox_06(self, Headers):
        global box_id

        # 发布工具箱
        update_data = {
            "status": "published"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id
        assert result[1]["status"] == "published"

        # 编辑工具箱
        name = ''.join(random.choice(characters) for i in range(8))
        update_data = {
            "box_name": name,
            "box_desc": "test toolbox update description 11",
            "box_svc_url": "http://127.0.0.1",
            "box_icon": "icon-color-tool-FADB14",
            "box_category": "data_process"
        }
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 200

        # 校验工具箱
        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["box_name"] == name
        assert result[1]["status"] == "published"

    @allure.title("更新已下架工具箱，更新成功")
    def test_update_toolbox_07(self, Headers):
        global box_id

        # 下架工具箱
        update_data = {
            "status": "offline"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id
        assert result[1]["status"] == "offline"

        # 编辑工具箱
        name = ''.join(random.choice(characters) for i in range(8))
        update_data = {
            "box_name": name,
            "box_desc": "test toolbox update description 22",
            "box_svc_url": "http://127.0.0.1",
            "box_icon": "icon-color-tool-FADB14",
            "box_category": "data_process"
        }
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 200

        # 校验工具箱
        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["box_name"] == name
        assert result[1]["status"] == "offline"

    @allure.title("更新工具箱元数据，metadata_type为其他类型，更新失败")
    def test_update_toolbox_08(self, Headers):
        global box_id
        update_data = {
            "metadata_type": "toolbox",
            "box_name": "test_toolbox_update_08",
            "box_desc": "test toolbox update description",
            "box_svc_url": "http://test.com",
            "box_category": "data_process"
        }
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("更新工具箱元数据，未匹配到当前工具箱中的任何工具，更新失败")
    def test_update_toolbox_09(self, Headers):
        name = ''.join(random.choice(string.ascii_letters + string.digits) for i in range(12))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"  
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]
        filepath = "./resource/openapi/compliant/relations.yaml"
        relations_data = GetContent(filepath).yamlfile()
        update_data = {
            "box_name": name,
            "box_desc": "test toolbox update description",
            "box_svc_url": "http://test.com",
            "box_category": "data_process",
            "metadata_type": "openapi",
            "data": relations_data
        }
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 404

    @allure.title("更新工具箱元数据，openapi中包含多个工具，可匹配到当前工具箱中的工具，更新成功。返回更新列表")
    def test_update_toolbox_10(self, Headers):
        name = ''.join(random.choice(string.ascii_letters + string.digits) for i in range(12))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"  
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]
        tool_ids = []
        result = self.client.GetBoxToolsList(box_id, None, Headers)
        tools = result[1]["tools"]
        for tool in tools:
            tool_ids.append(tool["tool_id"])
        filepath = "./resource/openapi/compliant/mcp_update.yaml"
        mcp_data = GetContent(filepath).yamlfile()
        update_data = {
            "box_name": name,
            "box_desc": "test toolbox update description",
            "box_svc_url": "http://test.com",
            "box_category": "data_process",
            "metadata_type": "openapi",
            "data": mcp_data
        }
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 200
        edit_tools = result[1]["edit_tools"]
        for tool in edit_tools:
            assert tool["tool_id"] in tool_ids

    @allure.title("更新工具箱元数据，openapi中仅包含一个工具，可匹配到当前工具箱中的工具，更新成功")
    def test_update_toolbox_11(self, Headers):
        name = ''.join(random.choice(string.ascii_letters + string.digits) for i in range(12))
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"  
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]
        tool_ids = []
        result = self.client.GetBoxToolsList(box_id, None, Headers)
        tools = result[1]["tools"]
        for tool in tools:
            tool_ids.append(tool["tool_id"])
        filepath = "./resource/openapi/compliant/mcp_update_01.yaml"
        mcp_data = GetContent(filepath).yamlfile()
        update_data = {
            "box_name": name,
            "box_desc": "test toolbox update description",
            "box_svc_url": "http://test.com",
            "box_category": "data_process",
            "metadata_type": "openapi",
            "data": mcp_data
        }
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 200
        edit_tools = result[1]["edit_tools"]
        for tool in edit_tools:
            assert tool["tool_id"] in tool_ids