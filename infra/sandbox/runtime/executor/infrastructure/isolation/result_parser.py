"""
Result parser for extracting return values from sandbox execution output.

Parses stdout to extract the handler return value using marker-based delimiters.
The Lambda-style wrapper prints return values between ===SANDBOX_RESULT=== and
===SANDBOX_RESULT_END=== markers.
"""

import re
import json
from typing import Optional, Any

from executor.infrastructure.logging.logging_config import get_logger


logger = get_logger()


def parse_return_value(stdout: str) -> Optional[Any]:
    """
    Extract return value from stdout using marker-based parsing.

    The Lambda-style code wrapper prints the handler return value between markers:
        ===SANDBOX_RESULT===<json>===SANDBOX_RESULT_END===

    Args:
        stdout: Standard output from code execution

    Returns:
        Parsed return value (dict, list, string, number, bool, or None)
        Returns None if markers are not found or JSON is invalid

    Examples:
        >>> parse_return_value("===SANDBOX_RESULT===\\n{\"status\": \"ok\"}\\n===SANDBOX_RESULT_END===")
        {'status': 'ok'}

        >>> parse_return_value("===SANDBOX_RESULT===\\n42\\n===SANDBOX_RESULT_END===")
        42

        >>> parse_return_value("No markers here")
        None
    """
    if not stdout:
        logger.debug("Empty stdout, no return value to parse")
        return None

    # Define marker pattern
    start_marker = "===SANDBOX_RESULT==="
    end_marker = "===SANDBOX_RESULT_END==="

    # Find start and end of result
    start_idx = stdout.find(start_marker)
    if start_idx == -1:
        logger.debug("Start marker not found in stdout")
        return None

    end_idx = stdout.find(end_marker, start_idx)
    if end_idx == -1:
        logger.warning("Start marker found but end marker missing", stdout_length=len(stdout))
        return None

    # Extract JSON content between markers
    result_start = start_idx + len(start_marker)
    result_content = stdout[result_start:end_idx].strip()

    if not result_content:
        logger.warning("Empty content between markers")
        return None

    # Parse JSON
    try:
        return_value = json.loads(result_content)
        logger.debug(
            "Successfully parsed return value",
            type=type(return_value).__name__,
            size=len(str(return_value)),
        )
        return return_value
    except json.JSONDecodeError as e:
        logger.error(
            "Failed to parse return value JSON",
            error=str(e),
            content_preview=result_content[:200],
        )
        return None


def has_result_markers(stdout: str) -> bool:
    """
    Check if stdout contains the result markers.

    Args:
        stdout: Standard output from code execution

    Returns:
        True if both start and end markers are present
    """
    if not stdout:
        return False

    return (
        "===SANDBOX_RESULT===" in stdout and "===SANDBOX_RESULT_END===" in stdout
    )


def remove_markers_from_output(stdout: str) -> str:
    """
    Remove result markers and their content from stdout.

    This is useful when you want to display stdout to the user without
    including the serialized return value.

    Args:
        stdout: Standard output from code execution

    Returns:
        Stdout with markers and content removed

    Examples:
        >>> remove_markers_from_output("Hello\\n===SANDBOX_RESULT===\\n{}\\n===SANDBOX_RESULT_END===\\nWorld")
        'Hello\\n\\nWorld'
    """
    if not stdout:
        return stdout

    start_marker = "===SANDBOX_RESULT==="
    end_marker = "===SANDBOX_RESULT_END==="

    start_idx = stdout.find(start_marker)
    if start_idx == -1:
        return stdout

    end_idx = stdout.find(end_marker, start_idx)
    if end_idx == -1:
        return stdout

    # Remove everything from start marker to after end marker
    result = stdout[:start_idx] + stdout[end_idx + len(end_marker):]
    return result


def extract_artifact_paths(stdout: str) -> list[str]:
    """
    Extract artifact file paths from stdout.

    Artifacts can be reported via special marker comments:
        # SANDBOX_ARTIFACT: path/to/file.txt

    Args:
        stdout: Standard output from code execution

    Returns:
        List of artifact file paths

    Examples:
        >>> extract_artifact_paths("Processing...\\n# SANDBOX_ARTIFACT: output/result.csv\\nDone")
        ['output/result.csv']
    """
    if not stdout:
        return []

    artifact_pattern = r"# SANDBOX_ARTIFACT:\s*(.+)"
    matches = re.findall(artifact_pattern, stdout)

    artifacts = [match.strip() for match in matches if match.strip()]
    if artifacts:
        logger.debug("Found artifact markers", count=len(artifacts))

    return artifacts
