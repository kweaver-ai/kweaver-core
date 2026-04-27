"""
依赖解析工具单元测试

测试 shared/utils/dependencies.py 模块的功能。
"""
import pytest

from src.shared.utils.dependencies import (
    parse_dependencies_to_pip_specs,
    format_dependencies_for_script,
    build_dependency_install_script,
    format_dependency_install_script_for_shell,
)


class TestParseDependenciesToPipSpecs:
    """依赖解析测试"""

    def test_parse_empty_list(self):
        """测试解析空列表"""
        result = parse_dependencies_to_pip_specs([])
        assert result == []

    def test_parse_none(self):
        """测试解析 None"""
        result = parse_dependencies_to_pip_specs(None)
        assert result == []

    def test_parse_string_dependencies(self):
        """测试解析字符串格式的依赖"""
        deps = ["requests==2.31.0", "pandas>=2.0"]
        result = parse_dependencies_to_pip_specs(deps)

        assert result == ["requests==2.31.0", "pandas>=2.0"]

    def test_parse_dict_dependencies_with_version(self):
        """测试解析字典格式的依赖（带版本）"""
        deps = [
            {"name": "requests", "version": "==2.31.0"},
            {"name": "pandas", "version": ">=2.0"}
        ]
        result = parse_dependencies_to_pip_specs(deps)

        assert result == ["requests==2.31.0", "pandas>=2.0"]

    def test_parse_dict_dependencies_without_version(self):
        """测试解析字典格式的依赖（不带版本）"""
        deps = [
            {"name": "requests"},
            {"name": "pandas", "version": ""}
        ]
        result = parse_dependencies_to_pip_specs(deps)

        assert result == ["requests", "pandas"]

    def test_parse_mixed_dependencies(self):
        """测试解析混合格式的依赖"""
        deps = [
            "requests==2.31.0",
            {"name": "pandas", "version": ">=2.0"},
            "numpy"
        ]
        result = parse_dependencies_to_pip_specs(deps)

        assert result == ["requests==2.31.0", "pandas>=2.0", "numpy"]

    def test_parse_complex_version_specs(self):
        """测试解析复杂版本规格"""
        deps = [
            {"name": "requests", "version": ">=2.28.0,<3.0"},
            {"name": "django", "version": "==4.2.*"}
        ]
        result = parse_dependencies_to_pip_specs(deps)

        assert "requests>=2.28.0,<3.0" in result
        assert "django==4.2.*" in result


class TestFormatDependenciesForScript:
    """格式化依赖脚本测试"""

    def test_format_empty_list(self):
        """测试格式化空列表"""
        deps_json, deps_list = format_dependencies_for_script([])

        assert deps_json == ""
        assert deps_list == ""

    def test_format_none(self):
        """测试格式化 None"""
        deps_json, deps_list = format_dependencies_for_script(None)

        assert deps_json == ""
        assert deps_list == ""

    def test_format_string_dependencies(self):
        """测试格式化字符串依赖"""
        deps = ["requests==2.31.0", "pandas>=2.0"]
        deps_json, deps_list = format_dependencies_for_script(deps)

        assert "requests" in deps_json
        assert "pandas" in deps_json
        assert '"requests==2.31.0"' in deps_list
        assert '"pandas>=2.0"' in deps_list

    def test_format_dict_dependencies(self):
        """测试格式化字典依赖"""
        deps = [{"name": "requests", "version": "==2.31.0"}]
        deps_json, deps_list = format_dependencies_for_script(deps)

        assert '"name"' in deps_json
        assert '"requests"' in deps_json
        assert '"requests==2.31.0"' in deps_list


class TestBuildDependencyInstallScript:
    """构建依赖安装脚本测试"""

    def test_build_script(self):
        """测试构建安装脚本"""
        script = build_dependency_install_script()

        assert "pip3 install" in script
        assert "/opt/sandbox-venv" in script
        assert "pypi.org" in script


class TestFormatDependencyInstallScriptForShell:
    """格式化 Shell 脚本测试"""

    def test_format_empty_list(self):
        """测试格式化空列表"""
        result = format_dependency_install_script_for_shell([])

        assert result == ""

    def test_format_none(self):
        """测试格式化 None"""
        result = format_dependency_install_script_for_shell(None)

        assert result == ""

    def test_format_dependencies(self):
        """测试格式化依赖"""
        deps = [
            {"name": "requests", "version": "==2.31.0"},
            "pandas>=2.0"
        ]
        result = format_dependency_install_script_for_shell(deps)

        assert "📦 Installing dependencies" in result
        assert "pip3 install" in result
        assert "/opt/sandbox-venv" in result

    def test_format_includes_pip_specs(self):
        """测试格式化包含 pip 规格"""
        deps = ["requests==2.31.0"]
        result = format_dependency_install_script_for_shell(deps)

        assert "requests==2.31.0" in result

    def test_format_includes_success_message(self):
        """测试格式化包含成功消息"""
        deps = ["requests"]
        result = format_dependency_install_script_for_shell(deps)

        assert "✅" in result
        assert "❌" in result
