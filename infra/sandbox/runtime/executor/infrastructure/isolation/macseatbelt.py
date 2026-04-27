"""
macOS Seatbelt Sandbox Execution Adapter

Implements secure code execution using macOS sandbox-exec (Seatbelt).
"""

import asyncio
import json
import subprocess
import time
import shutil
import tempfile
import os
from pathlib import Path
from typing import Optional
import structlog

from executor.domain.entities import Execution
from executor.domain.value_objects import ExecutionResult, ExecutionStatus, ExecutionMetrics
from executor.infrastructure.config import settings
from executor.infrastructure.isolation.code_wrapper import normalize_shell_code
from executor.infrastructure.isolation.result_parser import remove_markers_from_output


logger = structlog.get_logger(__name__)


# macOS Seatbelt sandbox profile for code execution
# Based on system profile syntax from /usr/share/sandbox/
SANDBOX_PROFILE = """
(version 1)
(deny default)
(debug deny)

(allow process-exec)
(allow file-read*)
(allow file-write* (subpath "/tmp"))
(allow system*)
"""


def check_sandbox_available() -> bool:
    """
    Check if sandbox-exec is available on the system.

    Returns:
        True if sandbox-exec is available, False otherwise

    Raises:
        RuntimeError: If sandbox-exec is not found
    """
    if not shutil.which("sandbox-exec"):
        raise RuntimeError("sandbox-exec is not installed or not in PATH")
    return True


def get_sandbox_version() -> str:
    """
    Get the sandbox-exec version.

    Returns:
        Version string (e.g., "sandbox-exec (macOS)")
    """
    try:
        import platform
        macos_version = platform.mac_ver()[0]
        # Just check if sandbox-exec exists, don't test execution
        if shutil.which("sandbox-exec"):
            return f"sandbox-exec (macOS {macos_version})"
        raise RuntimeError("sandbox-exec not found")
    except Exception as e:
        raise RuntimeError(f"Failed to get sandbox-exec version: {e}")


class MacSeatbeltRunner:
    """
    Executes code using macOS Seatbelt sandbox for process isolation.

    Provides sandbox isolation on macOS using the native sandbox-exec tool,
    which enforces filesystem access controls and process execution restrictions.
    """

    def __init__(self, workspace_path: Path):
        """
        Initialize the macOS Seatbelt runner.

        Args:
            workspace_path: Path to the workspace directory
        """
        self.workspace_path = workspace_path

    async def execute(self, execution: Execution) -> ExecutionResult:
        """
        Execute code within Seatbelt sandbox isolation.

        Args:
            execution: Execution entity with code and context

        Returns:
            ExecutionResult with stdout, stderr, exit code, timing, and metrics
        """
        start_time = time.perf_counter()
        start_cpu = time.process_time()
        profile_file = None

        logger.info(
            "Executing code in macOS sandbox",
            execution_id=execution.execution_id,
            language=execution.language,
        )

        try:
            # Write sandbox profile to a temporary file
            profile_file = tempfile.NamedTemporaryFile(
                mode='w',
                suffix='.sb',
                delete=False
            )
            profile_file.write(SANDBOX_PROFILE)
            profile_file.close()
            profile_path = profile_file.name

            # Build language-specific command with profile path
            cmd, env_args = self._build_command(execution, profile_path)

            # Prepare environment - merge with current environment
            exec_env = os.environ.copy()
            exec_env.update(env_args)
            exec_env["PYTHONPATH"] = self._build_pythonpath(exec_env.get("PYTHONPATH"))

            # Determine working directory
            cwd = str(execution.context.resolve_working_directory_path())

            # Execute with asyncio subprocess (non-blocking)
            process = await asyncio.create_subprocess_exec(
                *cmd,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
                cwd=cwd,
                env=exec_env,
            )

            # Wait for process to complete and capture output
            stdout_bytes, stderr_bytes = await process.communicate()

            # Convert bytes to string
            stdout = stdout_bytes.decode('utf-8')
            stderr = stderr_bytes.decode('utf-8')

            duration_ms = (time.perf_counter() - start_time) * 1000
            cpu_time_ms = (time.process_time() - start_cpu) * 1000

            # Parse output for return value (Python handler mode)
            return_value = None
            if execution.language.lower() == "python":
                return_value = self._parse_return_value(stdout)

            # Clean stdout by removing return value markers
            clean_stdout = remove_markers_from_output(stdout)

            # Collect performance metrics
            metrics = ExecutionMetrics(
                duration_ms=round(duration_ms, 2),
                cpu_time_ms=round(cpu_time_ms, 2),
            )

            execution_result = ExecutionResult(
                status=ExecutionStatus.COMPLETED if process.returncode == 0 else ExecutionStatus.FAILED,
                stdout=clean_stdout,
                stderr=stderr,
                exit_code=process.returncode,
                execution_time_ms=duration_ms,
                return_value=return_value,
                metrics=metrics,
            )

            logger.info(
                "Sandbox execution completed",
                execution_id=execution.execution_id,
                exit_code=process.returncode,
                duration_ms=duration_ms,
            )

            return execution_result

        except asyncio.TimeoutError:
            duration_ms = (time.perf_counter() - start_time) * 1000
            logger.warning("Sandbox execution timeout", execution_id=execution.execution_id)

            # Clean up: terminate the subprocess if still running
            if 'process' in locals() and process.returncode is None:
                try:
                    process.kill()
                    await process.wait()
                except Exception:
                    pass

            return ExecutionResult(
                status=ExecutionStatus.TIMEOUT,
                stdout="",
                stderr="Execution timeout",
                exit_code=-1,
                execution_time_ms=duration_ms,
                metrics=ExecutionMetrics(duration_ms=round(duration_ms, 2), cpu_time_ms=0),
            )

        except Exception as e:
            duration_ms = (time.perf_counter() - start_time) * 1000
            logger.error(
                "Sandbox execution error",
                execution_id=execution.execution_id,
                error=str(e),
            )
            return ExecutionResult(
                status=ExecutionStatus.FAILED,
                stdout="",
                stderr=str(e),
                exit_code=-1,
                execution_time_ms=duration_ms,
                error=str(e),
                metrics=ExecutionMetrics(duration_ms=round(duration_ms, 2), cpu_time_ms=0),
            )

        finally:
            # Clean up profile file
            if profile_file and profile_file.name:
                try:
                    os.unlink(profile_file.name)
                except Exception:
                    pass

    def _build_command(self, execution: Execution, profile_path: str) -> tuple[list, dict]:
        """
        Build the complete command for executing code in sandbox.

        Args:
            execution: Execution entity
            profile_path: Path to the sandbox profile file

        Returns:
            Tuple of (command list, environment variables)
        """
        lang = execution.language.lower()

        if lang == "python":
            return self._build_python_command(execution, profile_path)
        elif lang in ["javascript", "nodejs", "node"]:
            return self._build_node_command(execution, profile_path)
        elif lang in ["bash", "shell"]:
            return self._build_shell_command(execution, profile_path)
        else:
            raise ValueError(f"Unsupported language: {execution.language}")

    def _build_python_command(self, execution: Execution, profile_path: str) -> tuple[list, dict]:
        """Build command for Python handler execution."""
        # Wrap user code in AWS Lambda handler pattern
        wrapped_code = f'''
import json
import sys
import os

{execution.code}

if __name__ == "__main__":
    # Read event from environment variable
    event_json = os.environ.get("EVENT_JSON", "{{}}")
    event = json.loads(event_json)

    # Call handler function
    result = handler(event)

    # Print result with marker for parsing
    print("===SANDBOX_RESULT===" + json.dumps(result) + "===SANDBOX_RESULT_END===")
'''
        # Build command - pass env vars through shell environment
        cmd = [
            "sandbox-exec",
            "-f", profile_path,
            "--",
            "python3",
            "-c",
            wrapped_code,
        ]

        # Build environment dict to return
        env = {}
        if execution.context.event:
            env["EVENT_JSON"] = json.dumps(execution.context.event)
        env.update(execution.context.env_vars)

        return cmd, env

    def _build_node_command(self, execution: Execution, profile_path: str) -> tuple[list, dict]:
        """Build command for Node.js handler execution."""
        # Wrap user code in AWS Lambda handler pattern
        wrapped_code = f'''
{execution.code}

const eventJson = process.env.EVENT_JSON || '{{}}';
const event = JSON.parse(eventJson);

const result = handler(event, {{}});

console.log('===SANDBOX_RESULT===' + JSON.stringify(result) + '===SANDBOX_RESULT_END===');
'''
        cmd = [
            "sandbox-exec",
            "-f", profile_path,
            "--",
            "node",
            "-e",
            wrapped_code,
        ]

        # Build environment dict to return
        env = {}
        if execution.context.event:
            env["EVENT_JSON"] = json.dumps(execution.context.event)
        env.update(execution.context.env_vars)

        return cmd, env

    def _build_shell_command(self, execution: Execution, profile_path: str) -> tuple[list, dict]:
        """Build command for shell execution."""
        normalized_code = normalize_shell_code(
            execution.code,
            execution.context.resolve_working_directory_path(),
        )
        cmd = [
            "sandbox-exec",
            "-f", profile_path,
            "--",
            "bash",
            "-c",
            normalized_code,
        ]

        # Build environment dict to return
        env = {}
        if execution.context.event:
            env["EVENT_JSON"] = json.dumps(execution.context.event)
        env.update(execution.context.env_vars)

        return cmd, env

    def _parse_return_value(self, stdout: str) -> Optional[dict]:
        """
        Parse return value from stdout.

        Args:
            stdout: Standard output from execution

        Returns:
            Parsed return value dict, or None if not found
        """
        try:
            # Look for result markers
            if "===SANDBOX_RESULT===" in stdout:
                # Extract JSON between markers
                start = stdout.find("===SANDBOX_RESULT===") + len("===SANDBOX_RESULT===")
                end = stdout.find("===SANDBOX_RESULT_END===")
                if start > 0 and end > start:
                    json_str = stdout[start:end].strip()
                    return json.loads(json_str)
        except (json.JSONDecodeError, ValueError) as e:
            logger.warning("Failed to parse return value", error=str(e))
        return None

    def is_available(self) -> bool:
        """
        Check if sandbox is available.

        Returns:
            True if sandbox-exec is available
        """
        try:
            return check_sandbox_available()
        except RuntimeError:
            return False

    def get_version(self) -> str:
        """
        Get sandbox version information.

        Returns:
            Version string
        """
        return get_sandbox_version()

    def _build_pythonpath(self, existing_pythonpath: str | None) -> str:
        dependency_path = settings.dependency_install_path
        if existing_pythonpath:
            return f"{dependency_path}:{existing_pythonpath}"
        return dependency_path
