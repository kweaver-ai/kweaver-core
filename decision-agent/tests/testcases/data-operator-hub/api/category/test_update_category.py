# -*- coding:UTF-8 -*-

import allure
import pytest

from lib.operator_internal import InternalOperator

@allure.feature("MCP服务管理接口测试：更新算子分类")
class TestUpdateCategory:
    
    client = InternalOperator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, UserHeaders):
        category = {
            "category_type": "update_type",
            "name": "待更新分类"
        }
        result = self.client.CreateCategory(category, UserHeaders)
        assert result[0] == 200

    @allure.title("更新算子分类，传参正确，更新成功")
    def test_update_category_01(self, UserHeaders):
        result = self.client.UpdateCategory("update_type", {"name": "等待更新"}, UserHeaders)
        assert result[0] == 200
        result = self.client.GetCategory(UserHeaders)
        assert result[0] == 200
        assert "待更新分类" not in str(result[1])
        assert "等待更新" in str(result[1])

    @allure.title("更新算子分类，分类不存在，更新失败")
    def test_update_category_02(self, UserHeaders):
        result = self.client.UpdateCategory("update_type_01", {"name": "更新"}, UserHeaders)
        assert result[0] == 404

    @allure.title("更新算子分类，分类名称存在同名，更新失败")
    def test_update_category_03(self, UserHeaders):
        result = self.client.UpdateCategory("update_type", {"name": "未分类"}, UserHeaders)
        assert result[0] == 400

        result = self.client.UpdateCategory("update_type", {"name": "系统工具"}, UserHeaders)
        assert result[0] == 400

    @allure.title("更新算子分类，分类名称不合法，更新失败")
    @pytest.mark.parametrize("name", ["invalid name","name~","name@","name`","name#","name$","name%","name^","name^","name&", 
                                      "name*","name()","name-","name+","name=","name[]","name{}","name|","name\\","name:",
                                      "name;","name'","name,","name.","name?","name/","name<","name>","name；","name“","name：",
                                      "name’","name【】","name《","name》","name？","name·","name、","name，","name。",
                                      "invalid_name:_more_then_50_characters_aaaaaaaaaaaaa"]) 
    def test_update_category_04(self, name, UserHeaders):
        result = self.client.UpdateCategory("update_type", {"name": name}, UserHeaders)
        assert result[0] == 400