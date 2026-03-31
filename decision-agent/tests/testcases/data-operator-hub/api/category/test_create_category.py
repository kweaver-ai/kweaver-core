# -*- coding:UTF-8 -*-

import allure
import pytest

from lib.operator_internal import InternalOperator

@allure.feature("MCP服务管理接口测试：新建算子分类")
class TestCreateCategory:
    
    client = InternalOperator()

    @allure.title("新建算子分类，传参正确，新建成功")
    def test_create_category_01(self, UserHeaders):
        category = {
            "category_type": "debug",
            "name": "调试"
        }
        result = self.client.CreateCategory(category, UserHeaders)
        assert result[0] == 200

    @allure.title("新建算子分类，分类名称已存在，新建失败")
    @pytest.mark.parametrize("name", ["数据处理", "未分类", "系统工具"])
    def test_create_category_02(self, name, UserHeaders):
        category = {
            "category_type": "process",
            "name": name
        }
        result = self.client.CreateCategory(category, UserHeaders)
        assert result[0] == 400

    @allure.title("新建算子分类，分类类型已存在，新建失败")
    @pytest.mark.parametrize("catogory_type", ["data_process", "other_category", "system"])
    def test_create_category_03(self, catogory_type, UserHeaders):
        category = {
            "category_type": catogory_type,
            "name": "测试数据处理"
        }
        result = self.client.CreateCategory(category, UserHeaders)
        assert result[0] == 400

    @allure.title("新建算子分类，分类名称不合法，新建失败")
    @pytest.mark.parametrize("name", ["invalid name","name~","name@","name`","name#","name$","name%","name^","name^","name&", 
                                      "name*","name()","name-","name+","name=","name[]","name{}","name|","name\\","name:",
                                      "name;","name'","name,","name.","name?","name/","name<","name>","name；","name“","name：",
                                      "name’","name【】","name《","name》","name？","name·","name、","name，","name。",
                                      "invalid_name:_more_then_50_characters_aaaaaaaaaaaaa"]) 
    def test_create_category_04(self, name, UserHeaders):
        category = {
            "category_type": "data_process",
            "name": name
        }
        result = self.client.CreateCategory(category, UserHeaders)
        assert result[0] == 400