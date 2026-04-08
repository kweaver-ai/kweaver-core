"""
Unit tests for Code Wrapper module.

Tests code wrapper generation for Lambda-style handler execution.
"""

import pytest
from pathlib import Path

from executor.infrastructure.isolation.code_wrapper import (
    generate_python_wrapper,
    generate_javascript_wrapper,
    generate_shell_wrapper,
    normalize_shell_code,
    validate_python_handler,
    extract_handler_signature,
    wrap_for_execution,
    unwrap_python_code,
)


class TestGeneratePythonWrapper:
    """Tests for generate_python_wrapper function."""

    def test_generate_wrapper_basic(self):
        """Test basic Python wrapper generation."""
        user_code = 'def handler(event):\n    return {"message": "Hello"}'

        wrapper = generate_python_wrapper(user_code)

        assert "def handler(event):" in wrapper
        assert "===SANDBOX_RESULT===" in wrapper
        assert "===SANDBOX_RESULT_END===" in wrapper
        assert 'import sys' in wrapper
        assert 'import json' in wrapper

    def test_generate_wrapper_with_event_processing(self):
        """Test wrapper handles event processing."""
        user_code = 'def handler(event):\n    return event'

        wrapper = generate_python_wrapper(user_code)

        assert "json.loads(stdin_data)" in wrapper
        assert "handler(event)" in wrapper

    def test_generate_wrapper_includes_user_code(self):
        """Test that user code is included in wrapper."""
        user_code = '''def handler(event):
    # Custom logic here
    result = process(event)
    return result'''

        wrapper = generate_python_wrapper(user_code)

        assert "# Custom logic here" in wrapper
        assert "result = process(event)" in wrapper

    def test_generate_wrapper_includes_error_handling(self):
        """Test that wrapper includes error handling."""
        user_code = 'def handler(event):\n    pass'

        wrapper = generate_python_wrapper(user_code)

        assert "except Exception as e:" in wrapper
        assert "traceback.print_exc()" in wrapper


class TestGenerateJavascriptWrapper:
    """Tests for generate_javascript_wrapper function."""

    def test_generate_javascript_wrapper_basic(self):
        """Test basic JavaScript wrapper generation."""
        user_code = 'console.log("Hello");'

        wrapper = generate_javascript_wrapper(user_code)

        assert "console.log" in wrapper
        assert "try {" in wrapper
        assert "} catch (error) {" in wrapper

    def test_generate_javascript_wrapper_includes_error_handling(self):
        """Test JavaScript wrapper error handling."""
        user_code = 'throw new Error("test");'

        wrapper = generate_javascript_wrapper(user_code)

        assert "process.exit(1)" in wrapper


class TestGenerateShellWrapper:
    """Tests for generate_shell_wrapper function."""

    def test_generate_shell_wrapper_basic(self):
        """Test basic shell wrapper generation."""
        user_code = 'echo "Hello"'

        wrapper = generate_shell_wrapper(user_code)

        assert "echo" in wrapper
        assert "set -e" in wrapper

    def test_generate_shell_wrapper_preserves_code(self):
        """Test that shell wrapper preserves user code."""
        user_code = '''#!/bin/bash
echo "Step 1"
echo "Step 2"'''

        wrapper = generate_shell_wrapper(user_code)

        assert 'echo "Step 1"' in wrapper
        assert 'echo "Step 2"' in wrapper


class TestNormalizeShellCode:
    """Tests for normalize_shell_code."""

    def test_normalize_strips_bash_prefix_for_command(self):
        """Test stripping accidental bash prefix before a normal command."""
        normalized = normalize_shell_code(
            "bash python scripts/analyze_project.py",
            Path("/workspace/skill/mini-wiki"),
        )

        assert normalized == "python scripts/analyze_project.py"

    def test_normalize_strips_multiple_segments(self):
        """Test stripping accidental bash prefix across chained commands."""
        normalized = normalize_shell_code(
            "bash cd xx/xx/ & bash python xxx.py",
            Path("/workspace/skill/mini-wiki"),
        )

        assert normalized == "cd xx/xx/ & python xxx.py"

    def test_normalize_keeps_valid_script_invocation(self, tmp_path: Path):
        """Test preserving valid bash script invocation."""
        workspace = tmp_path / "workspace"
        workspace.mkdir()
        (workspace / "run.sh").write_text("#!/bin/sh\necho ok\n")

        normalized = normalize_shell_code("bash run.sh", workspace)

        assert normalized == "bash run.sh"


class TestValidatePythonHandler:
    """Tests for validate_python_handler function."""

    def test_validate_valid_handler(self):
        """Test validation of valid handler."""
        code = 'def handler(event):\n    return {"status": "ok"}'

        is_valid, error = validate_python_handler(code)

        assert is_valid is True
        assert error is None

    def test_validate_empty_code(self):
        """Test validation of empty code."""
        is_valid, error = validate_python_handler("")

        assert is_valid is False
        assert "empty" in error.lower()

    def test_validate_missing_handler(self):
        """Test validation when handler is missing."""
        code = 'def some_function():\n    pass'

        is_valid, error = validate_python_handler(code)

        assert is_valid is False
        assert "handler" in error.lower()

    def test_validate_syntax_error(self):
        """Test validation of code with syntax error."""
        code = 'def handler(event):\n    return {'  # Missing closing brace

        is_valid, error = validate_python_handler(code)

        assert is_valid is False
        assert "Syntax error" in error

    def test_validate_handler_with_context(self):
        """Test validation of handler with context parameter."""
        code = 'def handler(event, context=None):\n    pass'

        is_valid, error = validate_python_handler(code)

        assert is_valid is True


class TestExtractHandlerSignature:
    """Tests for extract_handler_signature function."""

    def test_extract_simple_signature(self):
        """Test extracting simple handler signature."""
        code = 'def handler(event):\n    pass'

        signature = extract_handler_signature(code)

        assert signature == "def handler(event):"

    def test_extract_signature_with_context(self):
        """Test extracting handler signature with context."""
        code = 'def handler(event, context=None):\n    return {}'

        signature = extract_handler_signature(code)

        assert signature == "def handler(event, context=None):"

    def test_extract_signature_not_found(self):
        """Test when handler signature is not found."""
        code = 'def other_function():\n    pass'

        signature = extract_handler_signature(code)

        assert signature is None

    def test_extract_signature_with_type_hints(self):
        """Test extracting signature with type hints and return type.

        Note: The current regex pattern expects ':' immediately after ')',
        so signatures with return type hints (e.g., '-> dict') won't match.
        This test documents the current behavior.
        """
        code = 'def handler(event: dict) -> dict:\n    return event'

        signature = extract_handler_signature(code)

        # Current implementation doesn't handle return type hints
        # because pattern expects ':' immediately after ')'
        assert signature is None  # Documents current behavior


class TestWrapForExecution:
    """Tests for wrap_for_execution function."""

    def test_wrap_python_code(self):
        """Test wrapping Python code."""
        code = 'def handler(event):\n    return event'

        wrapped = wrap_for_execution(code, "python")

        assert "===SANDBOX_RESULT===" in wrapped
        assert "def handler(event):" in wrapped

    def test_wrap_javascript_code(self):
        """Test wrapping JavaScript code."""
        code = 'console.log("test");'

        wrapped = wrap_for_execution(code, "javascript")

        assert "try {" in wrapped
        assert 'console.log("test")' in wrapped

    def test_wrap_shell_code(self):
        """Test wrapping shell code."""
        code = 'echo "test"'

        wrapped = wrap_for_execution(code, "shell")

        assert "set -e" in wrapped
        assert 'echo "test"' in wrapped

    def test_wrap_unsupported_language(self):
        """Test wrapping unsupported language raises error."""
        code = 'some code'

        with pytest.raises(ValueError, match="Unsupported language"):
            wrap_for_execution(code, "ruby")


class TestUnwrapPythonCode:
    """Tests for unwrap_python_code function."""

    def test_unwrap_python_code(self):
        """Test unwrapping Python code."""
        user_code = 'def handler(event):\n    return event'
        wrapped = generate_python_wrapper(user_code)

        unwrapped = unwrap_python_code(wrapped)

        assert "def handler(event):" in unwrapped
        assert "return event" in unwrapped

    def test_unwrap_without_markers(self):
        """Test unwrapping code without markers."""
        code = 'def handler(event):\n    pass'

        unwrapped = unwrap_python_code(code)

        # Should return original code if markers not found
        assert unwrapped == code

    def test_unwrap_empty_input(self):
        """Test unwrapping empty input."""
        unwrapped = unwrap_python_code("")

        assert unwrapped == ""
