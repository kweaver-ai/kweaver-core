# -*- coding: utf-8 -*-
"""
单元测试 - app/common/dependencies/dolphin_lazy_import.py
Dolphin延迟导入管理器提供 Dolphin SDK 的延迟导入功能。
"""

import pytest


class TestLazyDolphinImporterSingleton:
    """测试 LazyDolphinImporter 单例模式"""

    def test_singleton_returns_same_instance(self):
        """测试单例模式返回相同实例"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        instance1 = LazyDolphinImporter()
        instance2 = LazyDolphinImporter()

        assert instance1 is instance2

    def test_singleton_initialization(self):
        """测试单例初始化状态"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        instance = LazyDolphinImporter()

        assert hasattr(instance, "_import_cache")
        assert hasattr(instance, "_available")
        assert hasattr(instance, "_initialized")
        assert instance._initialized is True


class TestLazyDolphinImporterAvailableProperty:
    """测试 LazyDolphinImporter available 属性"""

    def test_available_returns_bool(self):
        """测试 available 属性返回布尔值"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        result = importer.available

        assert isinstance(result, bool)

    def test_available_caches_result(self):
        """测试 available 属性缓存结果"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        result1 = importer.available
        result2 = importer.available

        # 结果应该相同（被缓存）
        assert result1 == result2


class TestLazyDolphinImporterGetModule:
    """测试 LazyDolphinImporter get_module 方法"""

    def test_get_module_returns_something(self):
        """测试 get_module 返回值不为 None"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        result = importer.get_module("dolphin.core.common.exceptions")

        assert result is not None

    def test_get_module_caches_result(self):
        """测试 get_module 缓存结果"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        result1 = importer.get_module("dolphin.core.common.exceptions")
        result2 = importer.get_module("dolphin.core.common.exceptions")

        # 结果应该相同（被缓存）
        assert result1 is result2

    def test_get_module_different_paths(self):
        """测试 get_module 处理不同路径"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        module1 = importer.get_module("dolphin.core.common.exceptions")
        module2 = importer.get_module("dolphin.core.context.var_output")

        assert module1 is not None
        assert module2 is not None


class TestLazyDolphinImporterGetExceptionClass:
    """测试 LazyDolphinImporter get_exception_class 方法"""

    def test_get_exception_class_returns_exception(self):
        """测试 get_exception_class 返回异常类"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        _exception_class = importer.get_exception_class("ModelException")

        # 应该是一个异常类（可调用且是 Exception 的子类）
        assert callable(exception_class)
        assert issubclass(exception_class, Exception)

    def test_get_exception_class_has_correct_name(self):
        """测试 get_exception_class 返回的类名正确"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        _exception_class = importer.get_exception_class("TestException")

        assert exception_class.__name__ == "TestException"

    def test_get_exception_class_caches_result(self):
        """测试 get_exception_class 缓存结果"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        result1 = importer.get_exception_class("ModelException")
        result2 = importer.get_exception_class("ModelException")

        # 结果应该相同（被缓存）
        assert result1 is result2

    def test_get_exception_class_can_be_instantiated(self):
        """测试 get_exception_class 返回的类可以被实例化"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        _exception_class = importer.get_exception_class("ModelException")

        # 应该能创建异常实例
        exception = exception_class("Test message")
        assert str(exception) == "Test message"

    def test_get_exception_class_multiple_types(self):
        """测试 get_exception_class 处理多种异常类型"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()

        model_exc = importer.get_exception_class("ModelException")
        skill_exc = importer.get_exception_class("SkillException")
        dolphin_exc = importer.get_exception_class("DolphinException")

        assert issubclass(model_exc, Exception)
        assert issubclass(skill_exc, Exception)
        assert issubclass(dolphin_exc, Exception)


class TestLazyDolphinImporterGetVarOutputClass:
    """测试 LazyDolphinImporter get_var_output_class 方法"""

    def test_get_var_output_class_returns_class(self):
        """测试 get_var_output_class 返回类"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        var_output_class = importer.get_var_output_class()

        assert var_output_class is not None
        assert callable(var_output_class)

    def test_get_var_output_class_can_be_instantiated(self):
        """测试 get_var_output_class 返回的类可以被实例化"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        var_output_class = importer.get_var_output_class()

        instance = var_output_class()
        assert instance is not None

    def test_get_var_output_class_caches_result(self):
        """测试 get_var_output_class 缓存结果"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        result1 = importer.get_var_output_class()
        result2 = importer.get_var_output_class()

        # 结果应该相同（被缓存）
        assert result1 is result2


class TestVarOutputMockImplementation:
    """测试 VarOutput Mock 实现"""

    def test_var_output_has_set_method(self):
        """测试 VarOutput 有 set 方法"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        var_output_class = importer.get_var_output_class()
        instance = var_output_class()

        # 应该有 set 方法
        assert hasattr(instance, "set")
        assert callable(instance.set)

    def test_var_output_has_get_method(self):
        """测试 VarOutput 有 get 方法"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        var_output_class = importer.get_var_output_class()
        instance = var_output_class()

        # 应该有 get 方法
        assert hasattr(instance, "get")
        assert callable(instance.get)

    def test_var_output_has_delete_method(self):
        """测试 VarOutput 有 delete 方法"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        var_output_class = importer.get_var_output_class()
        instance = var_output_class()

        # 应该有 delete 方法
        assert hasattr(instance, "delete")
        assert callable(instance.delete)

    def test_var_output_set_get(self):
        """测试 VarOutput set 和 get 功能"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        var_output_class = importer.get_var_output_class()
        instance = var_output_class()

        # 设置值
        instance.set("test_key", "test_value")
        result = instance.get("test_key")

        assert result == "test_value"

    def test_var_output_get_default(self):
        """测试 VarOutput get 默认值"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        var_output_class = importer.get_var_output_class()
        instance = var_output_class()

        result = instance.get("nonexistent_key", "default_value")
        assert result == "default_value"

    def test_var_output_delete(self):
        """测试 VarOutput delete 功能"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        var_output_class = importer.get_var_output_class()
        instance = var_output_class()

        # 设置值然后删除
        instance.set("test_key", "test_value")
        instance.delete("test_key")
        result = instance.get("test_key", "default")

        assert result == "default"


class TestGlobalFunctions:
    """测试全局函数"""

    def test_is_dolphin_available_returns_bool(self):
        """测试 is_dolphin_available 返回布尔值"""
        from app.common.dependencies.dolphin_lazy_import import is_dolphin_available

        result = is_dolphin_available()
        assert isinstance(result, bool)

    def test_get_dolphin_exception_returns_exception_class(self):
        """测试 get_dolphin_exception 返回异常类"""
        from app.common.dependencies.dolphin_lazy_import import get_dolphin_exception

        exception_class = get_dolphin_exception("ModelException")
        assert issubclass(exception_class, Exception)

    def test_get_dolphin_exception_different_names(self):
        """测试 get_dolphin_exception 处理不同名称"""
        from app.common.dependencies.dolphin_lazy_import import get_dolphin_exception

        exc1 = get_dolphin_exception("ModelException")
        exc2 = get_dolphin_exception("SkillException")
        exc3 = get_dolphin_exception("CustomException")

        assert issubclass(exc1, Exception)
        assert issubclass(exc2, Exception)
        assert issubclass(exc3, Exception)

    def test_get_dolphin_var_output_class_returns_class(self):
        """测试 get_dolphin_var_output_class 返回类"""
        from app.common.dependencies.dolphin_lazy_import import (
            get_dolphin_var_output_class,
        )

        var_output_class = get_dolphin_var_output_class()
        assert callable(var_output_class)

    def test_create_dolphin_exception_creates_instance(self):
        """测试 create_dolphin_exception 创建异常实例"""
        from app.common.dependencies.dolphin_lazy_import import create_dolphin_exception

        exception = create_dolphin_exception("ModelException", "Test error message")

        assert isinstance(exception, Exception)
        assert str(exception) == "Test error message"

    def test_create_dolphin_exception_different_types(self):
        """测试 create_dolphin_exception 创建不同类型异常"""
        from app.common.dependencies.dolphin_lazy_import import create_dolphin_exception

        exc1 = create_dolphin_exception("ModelException", "Model error")
        exc2 = create_dolphin_exception("SkillException", "Skill error")

        assert isinstance(exc1, Exception)
        assert isinstance(exc2, Exception)
        assert str(exc1) == "Model error"
        assert str(exc2) == "Skill error"


class TestLazyImportDolphinDecorator:
    """测试 lazy_import_dolphin 装饰器"""

    def test_decorator_wraps_function(self):
        """测试装饰器正确包装函数"""
        from app.common.dependencies.dolphin_lazy_import import lazy_import_dolphin

        @lazy_import_dolphin
        def test_function():
            return "test_result"

        result = test_function()
        assert result == "test_result"

    def test_decorator_preserves_function_name(self):
        """测试装饰器保留函数名"""
        from app.common.dependencies.dolphin_lazy_import import lazy_import_dolphin

        @lazy_import_dolphin
        def my_function():
            return "result"

        assert my_function.__name__ == "my_function"

    def test_decorator_with_arguments(self):
        """测试装饰器处理带参数的函数"""
        from app.common.dependencies.dolphin_lazy_import import lazy_import_dolphin

        @lazy_import_dolphin
        def test_function(arg1, arg2, kwarg=None):
            return arg1 + arg2 + (kwarg or 0)

        result = test_function(1, 2, kwarg=3)
        assert result == 6

    def test_decorator_multiple_uses(self):
        """测试装饰器可以多次使用"""
        from app.common.dependencies.dolphin_lazy_import import lazy_import_dolphin

        @lazy_import_dolphin
        def func1():
            return "func1"

        @lazy_import_dolphin
        def func2():
            return "func2"

        assert func1() == "func1"
        assert func2() == "func2"


class TestModuleLevelExceptions:
    """测试模块级异常类"""

    def test_model_exception_exists(self):
        """测试 ModelException 存在"""
        from app.common.dependencies.dolphin_lazy_import import ModelException

        assert ModelException is not None
        assert issubclass(ModelException, Exception)

    def test_skill_exception_exists(self):
        """测试 SkillException 存在"""
        from app.common.dependencies.dolphin_lazy_import import SkillException

        assert SkillException is not None
        assert issubclass(SkillException, Exception)

    def test_dolphin_exception_exists(self):
        """测试 DolphinException 存在"""
        from app.common.dependencies.dolphin_lazy_import import DolphinException

        assert DolphinException is not None
        assert issubclass(DolphinException, Exception)

    def test_model_exception_can_be_raised(self):
        """测试 ModelException 可以被抛出"""
        from app.common.dependencies.dolphin_lazy_import import ModelException

        with pytest.raises(ModelException):
            raise ModelException("Test error")

    def test_skill_exception_can_be_raised(self):
        """测试 SkillException 可以被抛出"""
        from app.common.dependencies.dolphin_lazy_import import SkillException

        with pytest.raises(SkillException):
            raise SkillException("Test error")

    def test_dolphin_exception_can_be_raised(self):
        """测试 DolphinException 可以被抛出"""
        from app.common.dependencies.dolphin_lazy_import import DolphinException

        with pytest.raises(DolphinException):
            raise DolphinException("Test error")

    def test_model_exception_has_correct_name(self):
        """测试 ModelException 有正确的名称"""
        from app.common.dependencies.dolphin_lazy_import import ModelException

        assert ModelException.__name__ == "ModelException"

    def test_skill_exception_has_correct_name(self):
        """测试 SkillException 有正确的名称"""
        from app.common.dependencies.dolphin_lazy_import import SkillException

        assert SkillException.__name__ == "SkillException"

    def test_dolphin_exception_has_correct_name(self):
        """测试 DolphinException 有正确的名称"""
        from app.common.dependencies.dolphin_lazy_import import DolphinException

        assert DolphinException.__name__ == "DolphinException"


class TestLazyDolphinImporterEdgeCases:
    """测试边界情况"""

    def test_empty_exception_name(self):
        """测试空异常名称"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        _exception_class = importer.get_exception_class("")

        # 即使名称为空，也应该返回一个有效的异常类
        assert issubclass(exception_class, Exception)

    def test_special_characters_in_exception_name(self):
        """测试异常名称中的特殊字符"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        _exception_class = importer.get_exception_class("MyCustom_Exception_123")

        assert issubclass(exception_class, Exception)
        assert exception_class.__name__ == "MyCustom_Exception_123"

    def test_unicode_exception_name(self):
        """测试 Unicode 异常名称"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        _exception_class = importer.get_exception_class("异常类")

        assert issubclass(exception_class, Exception)
        assert exception_class.__name__ == "异常类"

    def test_var_output_multiple_instances(self):
        """测试 VarOutput 多个实例独立"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        var_output_class = importer.get_var_output_class()

        instance1 = var_output_class()
        instance2 = var_output_class()

        # 两个实例应该独立
        instance1.set("key", "value1")
        instance2.set("key", "value2")

        assert instance1.get("key") == "value1"
        assert instance2.get("key") == "value2"

    def test_var_output_overwrite_key(self):
        """测试 VarOutput 覆盖键值"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        var_output_class = importer.get_var_output_class()
        instance = var_output_class()

        instance.set("key", "value1")
        instance.set("key", "value2")

        assert instance.get("key") == "value2"

    def test_var_output_get_none_default(self):
        """测试 VarOutput get 返回 None 作为默认值"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()
        var_output_class = importer.get_var_output_class()
        instance = var_output_class()

        result = instance.get("nonexistent_key")
        assert result is None


class TestImportCacheIsolation:
    """测试导入缓存隔离"""

    def test_module_cache_separate_from_exception_cache(self):
        """测试模块缓存与异常缓存分离"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()

        # 获取模块和异常类
        _module = importer.get_module("dolphin.core.common.exceptions")
        _exception_class = importer.get_exception_class("ModelException")

        # 检查缓存中有不同的键
        assert "dolphin.core.common.exceptions" in importer._import_cache
        assert "exception_ModelException" in importer._import_cache

    def test_cache_keys_unique(self):
        """测试缓存键唯一性"""
        from app.common.dependencies.dolphin_lazy_import import LazyDolphinImporter

        importer = LazyDolphinImporter()

        importer.get_exception_class("ModelException")
        importer.get_module("dolphin.core.common.exceptions")
        importer.get_var_output_class()

        # 验证三个不同的缓存项
        assert len(importer._import_cache) >= 3


class TestLazyDolphinImporterIntegration:
    """集成测试"""

    def test_full_workflow_exception_creation(self):
        """测试完整的异常创建工作流"""
        from app.common.dependencies.dolphin_lazy_import import (
            get_dolphin_exception,
            create_dolphin_exception,
        )

        # 获取异常类
        exception_class = get_dolphin_exception("TestException")

        # 创建异常实例
        exception = create_dolphin_exception("TestException", "Test message")

        assert isinstance(exception, exception_class)
        assert str(exception) == "Test message"

    def test_full_workflow_var_output(self):
        """测试完整的 VarOutput 工作流"""
        from app.common.dependencies.dolphin_lazy_import import (
            get_dolphin_var_output_class,
        )

        var_output_class = get_dolphin_var_output_class()
        instance = var_output_class()

        # 设置多个值
        instance.set("key1", "value1")
        instance.set("key2", "value2")
        instance.set("key3", "value3")

        # 获取值
        assert instance.get("key1") == "value1"
        assert instance.get("key2") == "value2"
        assert instance.get("key3") == "value3"

        # 删除一个值
        instance.delete("key2")
        assert instance.get("key2", "default") == "default"

        # 其他值仍然存在
        assert instance.get("key1") == "value1"
        assert instance.get("key3") == "value3"
