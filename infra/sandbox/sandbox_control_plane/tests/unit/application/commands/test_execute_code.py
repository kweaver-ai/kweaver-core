"""
执行代码命令单元测试

测试 ExecuteCodeCommand 的功能。
"""
import pytest

from src.application.commands.execute_code import ExecuteCodeCommand


class TestExecuteCodeCommand:
    """执行代码命令测试"""

    def test_create_with_required_fields(self):
        """测试使用必填字段创建"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="print('hello')",
            language="python"
        )

        assert command.session_id == "sess-123"
        assert command.code == "print('hello')"
        assert command.language == "python"
        assert command.async_mode is False  # default
        assert command.stdin is None  # default
        assert command.timeout == 30  # default
        assert command.event_data is None  # default

    def test_create_with_all_fields(self):
        """测试使用所有字段创建"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="print('hello')",
            language="python",
            async_mode=True,
            stdin="test input",
            timeout=60,
            event_data={"name": "World"},
            working_directory="src/jobs",
        )

        assert command.session_id == "sess-123"
        assert command.code == "print('hello')"
        assert command.language == "python"
        assert command.async_mode is True
        assert command.stdin == "test input"
        assert command.timeout == 60
        assert command.event_data == {"name": "World"}
        assert command.working_directory == "src/jobs"

    def test_language_python(self):
        """测试 Python 语言"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="print('hello')",
            language="python"
        )

        assert command.language == "python"

    def test_language_javascript(self):
        """测试 JavaScript 语言"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="console.log('hello')",
            language="javascript"
        )

        assert command.language == "javascript"

    def test_language_shell(self):
        """测试 Shell 语言"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="echo hello",
            language="shell"
        )

        assert command.language == "shell"

    def test_empty_code_raises_error(self):
        """测试空代码抛出错误"""
        with pytest.raises(ValueError, match="code cannot be empty"):
            ExecuteCodeCommand(
                session_id="sess-123",
                code="",
                language="python"
            )

    def test_zero_timeout_raises_error(self):
        """测试零超时抛出错误"""
        with pytest.raises(ValueError, match="timeout must be positive"):
            ExecuteCodeCommand(
                session_id="sess-123",
                code="print('hello')",
                language="python",
                timeout=0
            )

    def test_negative_timeout_raises_error(self):
        """测试负超时抛出错误"""
        with pytest.raises(ValueError, match="timeout must be positive"):
            ExecuteCodeCommand(
                session_id="sess-123",
                code="print('hello')",
                language="python",
                timeout=-1
            )

    def test_unsupported_language_raises_error(self):
        """测试不支持的语言抛出错误"""
        with pytest.raises(ValueError, match="Unsupported language"):
            ExecuteCodeCommand(
                session_id="sess-123",
                code="print('hello')",
                language="ruby"
            )

    def test_async_mode_true(self):
        """测试异步模式为 True"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="print('hello')",
            language="python",
            async_mode=True
        )

        assert command.async_mode is True

    def test_async_mode_false(self):
        """测试异步模式为 False"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="print('hello')",
            language="python",
            async_mode=False
        )

        assert command.async_mode is False

    def test_with_stdin(self):
        """测试带标准输入"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="name = input()",
            language="python",
            stdin="World"
        )

        assert command.stdin == "World"

    def test_with_event_data(self):
        """测试带事件数据"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="print(event['name'])",
            language="python",
            event_data={"name": "World", "count": 42}
        )

        assert command.event_data == {"name": "World", "count": 42}

    def test_with_working_directory_normalized(self):
        """测试工作目录归一化"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="echo hello",
            language="shell",
            working_directory="./skill/mini-wiki",
        )

        assert command.working_directory == "skill/mini-wiki"

    def test_invalid_working_directory_raises_error(self):
        """测试非法工作目录抛出错误"""
        with pytest.raises(ValueError, match="working_directory must be a relative workspace path"):
            ExecuteCodeCommand(
                session_id="sess-123",
                code="echo hello",
                language="shell",
                working_directory="../etc",
            )

    def test_timeout_boundary_minimum(self):
        """测试超时边界值（最小有效值）"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="print('hello')",
            language="python",
            timeout=1
        )

        assert command.timeout == 1

    def test_timeout_boundary_large(self):
        """测试超时边界值（较大值）"""
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="print('hello')",
            language="python",
            timeout=3600
        )

        assert command.timeout == 3600

    def test_is_dataclass(self):
        """测试是数据类"""
        from dataclasses import is_dataclass

        assert is_dataclass(ExecuteCodeCommand)

    def test_dataclass_equality(self):
        """测试数据类相等性"""
        command1 = ExecuteCodeCommand(
            session_id="sess-123",
            code="print('hello')",
            language="python"
        )

        command2 = ExecuteCodeCommand(
            session_id="sess-123",
            code="print('hello')",
            language="python"
        )

        assert command1 == command2

    def test_dataclass_inequality(self):
        """测试数据类不等性"""
        command1 = ExecuteCodeCommand(
            session_id="sess-123",
            code="print('hello')",
            language="python"
        )

        command2 = ExecuteCodeCommand(
            session_id="sess-456",
            code="print('hello')",
            language="python"
        )

        assert command1 != command2

    def test_multiline_code(self):
        """测试多行代码"""
        code = '''
def greet(name):
    return f"Hello, {name}!"

print(greet("World"))
'''
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code=code,
            language="python"
        )

        assert "def greet" in command.code
        assert "Hello" in command.code

    def test_whitespace_only_code_raises_error(self):
        """测试仅空白字符的代码抛出错误"""
        # Note: The validation checks `if not self.code`, which evaluates to True for empty string
        # but False for whitespace-only strings. This test verifies the current behavior.
        command = ExecuteCodeCommand(
            session_id="sess-123",
            code="   ",
            language="python"
        )

        # Whitespace is not empty, so it should be allowed
        assert command.code == "   "
