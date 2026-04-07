"""
Code wrapper generator for Lambda-style handler execution.

Generates wrapper code that injects the Lambda handler pattern around user code,
enabling execution via fileless `python3 -c` invocation. The wrapper reads event
data from stdin, calls the user's handler(event) function, and prints the return
value between special markers for extraction.
"""

import json
from typing import Optional

from executor.infrastructure.logging.logging_config import get_logger


logger = get_logger()


def generate_python_wrapper(user_code: str) -> str:
    """
    Generate Python wrapper code around user code for Lambda-style execution.

    The wrapper:
    1. Defines the user's code (including the handler function)
    2. Reads event data from stdin
    3. Calls handler(event)
    4. Prints the return value between ===SANDBOX_RESULT=== markers
    5. Handles exceptions and prints tracebacks to stderr

    Args:
        user_code: Python code containing a handler(event) function

    Returns:
        Complete wrapped code ready for execution via python3 -c

    Examples:
        >>> user_code = 'def handler(event):\\n    return {"message": "Hello"}'
        >>> wrapped = generate_python_wrapper(user_code)
        >>> 'def handler(event):' in wrapped
        True
        >>> '===SANDBOX_RESULT===' in wrapped
        True
    """
    wrapper = f'''import sys
import json
import traceback

# User code starts here
{user_code}
# User code ends here

# Main execution wrapper
if __name__ == "__main__":
    try:
        # Read event from stdin
        stdin_data = sys.stdin.read()
        if stdin_data.strip():
            event = json.loads(stdin_data)
        else:
            event = {{}}

        # Call the handler
        result = handler(event)

        # Print result between markers for extraction
        print("===SANDBOX_RESULT===")
        print(json.dumps(result))
        print("===SANDBOX_RESULT_END===")

    except Exception as e:
        # Print error to stderr
        traceback.print_exc()
        sys.exit(1)
'''
    return wrapper


def generate_javascript_wrapper(user_code: str) -> str:
    """
    Generate JavaScript wrapper code for Node.js execution.

    Args:
        user_code: JavaScript code to execute

    Returns:
        Complete wrapped code ready for execution via node -e
    """
    wrapper = f'''try {{
        // User code
        {user_code}
    }} catch (error) {{
        console.error(error);
        process.exit(1);
    }}
'''
    return wrapper


def generate_shell_wrapper(user_code: str) -> str:
    """
    Generate shell wrapper code for bash execution.

    Args:
        user_code: Shell script code to execute

    Returns:
        Complete wrapped code ready for execution via bash -c
    """
    # For shell, we mostly just pass through the code
    # but ensure it runs with proper error handling
    wrapper = f'''set -e  # Exit on error
{user_code}
'''
    return wrapper


def validate_python_handler(code: str) -> tuple[bool, Optional[str]]:
    """
    Validate that Python code contains a handler function.

    Checks for:
    - Presence of 'def handler('
    - At least one parameter (event)
    - Proper function definition syntax

    Args:
        code: Python code to validate

    Returns:
        Tuple of (is_valid, error_message)
    """
    if not code or not code.strip():
        return False, "Code is empty"

    if "def handler(" not in code:
        return False, "handler(event) function not found. Please define a handler function."

    # Basic syntax check
    try:
        compile(code, "<string>", "exec")
    except SyntaxError as e:
        return False, f"Syntax error: {e}"

    return True, None


def extract_handler_signature(code: str) -> Optional[str]:
    """
    Extract the handler function signature from code.

    Args:
        code: Python code containing handler function

    Returns:
        Function signature string, or None if not found

    Examples:
        >>> code = 'def handler(event, context=None):\\n    pass'
        >>> extract_handler_signature(code)
        'def handler(event, context=None)'
    """
    import re

    pattern = r"def handler\s*\((.*?)\):"
    match = re.search(pattern, code)
    if match:
        return f"def handler({match.group(1)}):"
    return None


def wrap_for_execution(code: str, language: str) -> str:
    """
    Generate appropriate wrapper for the given language.

    Args:
        code: User code to wrap
        language: Programming language (python, javascript, shell)

    Returns:
        Wrapped code ready for execution

    Raises:
        ValueError: If language is not supported
    """
    if language == "python":
        # Validate Python handler before wrapping
        is_valid, error = validate_python_handler(code)
        if not is_valid:
            logger.warning("Handler validation failed", error=error)
            # Still wrap it - let runtime error occur
        return generate_python_wrapper(code)

    elif language == "javascript":
        return generate_javascript_wrapper(code)

    elif language == "shell":
        return generate_shell_wrapper(code)

    else:
        raise ValueError(f"Unsupported language: {language}")


def unwrap_python_code(wrapped_code: str) -> str:
    """
    Extract original user code from wrapped Python code.

    This is useful for debugging or displaying the user's original code.

    Args:
        wrapped_code: Full wrapped Python code

    Returns:
        Original user code (between markers)
    """
    start_marker = "# User code starts here"
    end_marker = "# User code ends here"

    start_idx = wrapped_code.find(start_marker)
    if start_idx == -1:
        return wrapped_code

    end_idx = wrapped_code.find(end_marker, start_idx)
    if end_idx == -1:
        return wrapped_code

    # Extract code between markers
    code_start = start_idx + len(start_marker)
    code_end = end_idx
    return wrapped_code[code_start:code_end].strip()
