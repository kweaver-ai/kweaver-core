"""单元测试 - utils/dict_util 模块"""

import pytest


class TestDictPathParser:
    """测试 DictPathParser 类"""

    def test_init_with_data(self):
        """测试使用数据初始化"""
        from app.utils.dict_util import DictPathParser

        data = {"a": {"b": {"c": 1}}}
        parser = DictPathParser(data)

        assert parser.data == data

    def test_init_without_data(self):
        """测试不使用数据初始化"""
        from app.utils.dict_util import DictPathParser

        parser = DictPathParser()

        assert parser.data == {}

    def test_get_simple_path(self):
        """测试获取简单路径"""
        from app.utils.dict_util import DictPathParser

        data = {"a": {"b": {"c": 1}}}
        parser = DictPathParser(data)

        result = parser.get("a.b.c")

        assert result == 1

    def test_get_empty_path_returns_data(self):
        """测试空路径返回原始数据"""
        from app.utils.dict_util import DictPathParser

        data = {"a": 1}
        parser = DictPathParser(data)

        result = parser.get("")

        assert result == data

    def test_get_array_with_wildcard(self):
        """测试获取数组元素（通配符）"""
        from app.utils.dict_util import DictPathParser

        data = {"a": [{"b": 1}, {"b": 2}]}
        parser = DictPathParser(data)

        result = parser.get("a[*].b")

        assert result == [1, 2]

    def test_get_array_with_index(self):
        """测试获取数组元素（索引）"""
        from app.utils.dict_util import DictPathParser

        data = {"a": [{"b": 1}, {"b": 2}]}
        parser = DictPathParser(data)

        result = parser.get("a[0].b")

        assert result == 1

    def test_get_flat(self):
        """测试获取扁平化结果"""
        from app.utils.dict_util import DictPathParser

        data = {"a": {"b": [{"c": 1}, {"c": 2}]}}
        parser = DictPathParser(data)

        result = parser.get_flat("a.b[*].c")

        assert result == [1, 2]

    def test_set_simple_path(self):
        """测试设置简单路径"""
        from app.utils.dict_util import DictPathParser

        parser = DictPathParser()
        parser.set("a.b.c", 1)

        assert parser.data == {"a": {"b": {"c": 1}}}

    def test_set_empty_path_sets_data(self):
        """测试设置空路径设置数据"""
        from app.utils.dict_util import DictPathParser

        parser = DictPathParser()
        parser.set("", "new_data")

        assert parser.data == "new_data"

    def test_has_existing_path(self):
        """测试检查存在的路径"""
        from app.utils.dict_util import DictPathParser

        data = {"a": {"b": {"c": 1}}}
        parser = DictPathParser(data)

        assert parser.has("a.b.c") is True

    def test_has_non_existing_path(self):
        """测试检查不存在的路径"""
        from app.utils.dict_util import DictPathParser

        data = {"a": {"b": {"c": 1}}}
        parser = DictPathParser(data)

        assert parser.has("a.b.d") is False

    def test_delete_existing_path(self):
        """测试删除存在的路径"""
        from app.utils.dict_util import DictPathParser

        data = {"a": {"b": {"c": 1, "d": 2}}}
        parser = DictPathParser(data)

        result = parser.delete("a.b.c")

        assert result is True
        assert parser.data == {"a": {"b": {"d": 2}}}

    def test_delete_non_existing_path(self):
        """测试删除不存在的路径"""
        from app.utils.dict_util import DictPathParser

        data = {"a": {"b": {"c": 1}}}
        parser = DictPathParser(data)

        result = parser.delete("a.b.d")

        assert result is False

    def test_get_all_paths_from_dict(self):
        """测试获取字典的所有路径"""
        from app.utils.dict_util import DictPathParser

        data = {"a": 1, "b": {"c": 2}}
        parser = DictPathParser(data)

        paths = parser.get_all_paths()

        assert "a" in paths
        assert "b" in paths
        assert "b.c" in paths


class TestGetDictValByPath:
    """测试 get_dict_val_by_path 函数"""

    def test_get_value_from_nested_dict(self):
        """测试从嵌套字典获取值"""
        from app.utils.dict_util import get_dict_val_by_path

        data = {"a": {"b": {"c": 1}}}
        result = get_dict_val_by_path(data, "a.b.c")

        assert result == 1

    def test_get_value_with_array_wildcard(self):
        """测试使用通配符获取数组值"""
        from app.utils.dict_util import get_dict_val_by_path

        data = {"items": [{"value": 1}, {"value": 2}]}
        result = get_dict_val_by_path(data, "items[*].value")

        assert result == [1, 2]

    def test_get_non_existing_path_raises_error(self):
        """测试获取不存在的路径抛出异常"""
        from app.utils.dict_util import get_dict_val_by_path

        data = {"a": {"b": 1}}

        with pytest.raises(KeyError):
            get_dict_val_by_path(data, "a.c")


class TestSetDictValByPath:
    """测试 set_dict_val_by_path 函数"""

    def test_set_value_returns_new_data(self):
        """测试设置值返回新数据"""
        from app.utils.dict_util import set_dict_val_by_path

        data = {}
        result = set_dict_val_by_path(data, "a.b.c", 1)

        # The function returns a new data structure
        assert result == {"a": {"b": {"c": 1}}}

    def test_set_value_returns_modified_data(self):
        """测试设置值返回修改后的数据"""
        from app.utils.dict_util import set_dict_val_by_path

        data = {"a": {"b": {"c": 1}}}
        result = set_dict_val_by_path(data, "a.b.c", 2)

        # The function returns a new data structure with modified value
        assert result == {"a": {"b": {"c": 2}}}


class TestModuleExports:
    """测试模块导出"""

    def test_module_exports_dict_path_parser(self):
        """测试模块导出 DictPathParser"""
        from app.utils.dict_util import DictPathParser

        assert DictPathParser is not None

    def test_module_exports_dict_path_parser_flat(self):
        """测试模块导出 DictPathParserFlat"""
        from app.utils.dict_util import DictPathParserFlat

        assert DictPathParserFlat is not None

    def test_module_exports_get_dict_val_by_path(self):
        """测试模块导出 get_dict_val_by_path"""
        from app.utils.dict_util import get_dict_val_by_path

        assert get_dict_val_by_path is not None

    def test_module_exports_get_dic_val_by_path_flat(self):
        """测试模块导出 get_dic_val_by_path_flat"""
        from app.utils.dict_util import get_dic_val_by_path_flat

        assert get_dic_val_by_path_flat is not None

    def test_module_exports_set_dict_val_by_path(self):
        """测试模块导出 set_dict_val_by_path"""
        from app.utils.dict_util import set_dict_val_by_path

        assert set_dict_val_by_path is not None
