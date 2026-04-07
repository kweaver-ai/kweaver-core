"""
执行请求值对象单元测试

测试 ExecutionRequest 的功能。
"""
import pytest

from src.domain.value_objects.execution_request import ExecutionRequest


class TestExecutionRequest:
    """执行请求测试"""

    def test_create_with_required_fields(self):
        """测试使用必填字段创建"""
        request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={"name": "World"},
            timeout=60,
            env_vars={"DEBUG": "true"}
        )

        assert request.code == "print('hello')"
        assert request.language == "python"
        assert request.event == {"name": "World"}
        assert request.timeout == 60
        assert request.env_vars == {"DEBUG": "true"}
        assert request.execution_id is None
        assert request.session_id is None

    def test_create_with_all_fields(self):
        """测试使用所有字段创建"""
        request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={"name": "World"},
            timeout=60,
            env_vars={"DEBUG": "true"},
            execution_id="exec-123",
            session_id="sess-456"
        )

        assert request.execution_id == "exec-123"
        assert request.session_id == "sess-456"

    def test_language_python(self):
        """测试 Python 语言"""
        request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={}
        )
        assert request.language == "python"

    def test_language_javascript(self):
        """测试 JavaScript 语言"""
        request = ExecutionRequest(
            code="console.log('hello')",
            language="javascript",
            event={},
            timeout=60,
            env_vars={}
        )
        assert request.language == "javascript"

    def test_language_shell(self):
        """测试 Shell 语言"""
        request = ExecutionRequest(
            code="echo hello",
            language="shell",
            event={},
            timeout=60,
            env_vars={}
        )
        assert request.language == "shell"

    def test_empty_code_raises_error(self):
        """测试空代码抛出错误"""
        with pytest.raises(ValueError, match="code cannot be empty"):
            ExecutionRequest(
                code="",
                language="python",
                event={},
                timeout=60,
                env_vars={}
            )

    def test_empty_language_raises_error(self):
        """测试空语言抛出错误"""
        with pytest.raises(ValueError, match="language cannot be empty"):
            ExecutionRequest(
                code="print('hello')",
                language="",
                event={},
                timeout=60,
                env_vars={}
            )

    def test_timeout_below_minimum_raises_error(self):
        """测试超时低于最小值抛出错误"""
        with pytest.raises(ValueError, match="timeout must be between 1 and 3600"):
            ExecutionRequest(
                code="print('hello')",
                language="python",
                event={},
                timeout=0,
                env_vars={}
            )

    def test_timeout_above_maximum_raises_error(self):
        """测试超时高于最大值抛出错误"""
        with pytest.raises(ValueError, match="timeout must be between 1 and 3600"):
            ExecutionRequest(
                code="print('hello')",
                language="python",
                event={},
                timeout=3601,
                env_vars={}
            )

    def test_timeout_negative_raises_error(self):
        """测试负超时抛出错误"""
        with pytest.raises(ValueError, match="timeout must be between 1 and 3600"):
            ExecutionRequest(
                code="print('hello')",
                language="python",
                event={},
                timeout=-1,
                env_vars={}
            )

    def test_unsupported_language_raises_error(self):
        """测试不支持的语言抛出错误"""
        with pytest.raises(ValueError, match="unsupported language"):
            ExecutionRequest(
                code="print('hello')",
                language="ruby",
                event={},
                timeout=60,
                env_vars={}
            )

    def test_timeout_boundary_minimum(self):
        """测试超时边界值（最小有效值）"""
        request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=1,
            env_vars={}
        )
        assert request.timeout == 1

    def test_timeout_boundary_maximum(self):
        """测试超时边界值（最大有效值）"""
        request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=3600,
            env_vars={}
        )
        assert request.timeout == 3600

    def test_env_vars_empty(self):
        """测试空环境变量"""
        request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={}
        )
        assert request.env_vars == {}

    def test_env_vars_with_values(self):
        """测试带值的环境变量"""
        request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={"DEBUG": "true", "API_KEY": "secret"}
        )
        assert request.env_vars == {"DEBUG": "true", "API_KEY": "secret"}

    def test_event_empty(self):
        """测试空事件"""
        request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={}
        )
        assert request.event == {}

    def test_event_with_nested_data(self):
        """测试带嵌套数据的事件"""
        event = {
            "name": "World",
            "data": {"key": "value"},
            "list": [1, 2, 3]
        }
        request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event=event,
            timeout=60,
            env_vars={}
        )
        assert request.event == event

    def test_is_dataclass(self):
        """测试是数据类"""
        from dataclasses import is_dataclass
        assert is_dataclass(ExecutionRequest)

    def test_dataclass_equality(self):
        """测试数据类相等性"""
        request1 = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={}
        )
        request2 = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={}
        )
        assert request1 == request2

    def test_dataclass_inequality(self):
        """测试数据类不等性"""
        request1 = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={}
        )
        request2 = ExecutionRequest(
            code="print('world')",
            language="python",
            event={},
            timeout=60,
            env_vars={}
        )
        assert request1 != request2

    def test_multiline_code(self):
        """测试多行代码"""
        code = '''
def greet(name):
    return f"Hello, {name}!"

print(greet("World"))
'''
        request = ExecutionRequest(
            code=code,
            language="python",
            event={},
            timeout=60,
            env_vars={}
        )
        assert "def greet" in request.code

    def test_optional_execution_id(self):
        """测试可选执行 ID"""
        request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={},
            execution_id="exec-123"
        )
        assert request.execution_id == "exec-123"

    def test_optional_session_id(self):
        """测试可选会话 ID"""
        request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={},
            session_id="sess-456"
        )
        assert request.session_id == "sess-456"
