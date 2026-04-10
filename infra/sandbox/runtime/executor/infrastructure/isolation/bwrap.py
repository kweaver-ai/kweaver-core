"""

Bubblewrap Execution Adapter

Implements secure code execution using Bubblewrap for process isolation.
"""

import asyncio
import json
import os
import subprocess
import time
import shutil
from pathlib import Path
from typing import List, Optional
import structlog

from executor.domain.entities import Execution
from executor.domain.value_objects import ExecutionResult, ExecutionStatus, ExecutionMetrics
from executor.infrastructure.isolation.code_wrapper import normalize_shell_code
from executor.infrastructure.config import settings
from executor.infrastructure.isolation.result_parser import remove_markers_from_output


logger = structlog.get_logger(__name__)


def check_bwrap_available() -> bool:
    """
    Check if Bubblewrap is available on the system.

    Returns:
        True if bwrap is available, False otherwise

    Raises:
        RuntimeError: If bwrap is not found
    """
    if not shutil.which("bwrap"):
        raise RuntimeError("Bubblewrap (bwrap) is not installed or not in PATH")
    return True


def get_bwrap_version() -> str:
    """
    Get the Bubblewrap version.

    Returns:
        Version string (e.g., "1.7.0")

    Raises:
        RuntimeError: If bwrap is not available or version cannot be determined
    """
    try:
        result = subprocess.run(
            ["bwrap", "--version"],
            capture_output=True,
            text=True,
            timeout=5,
        )
        if result.returncode == 0:
            # Extract version from output like "bwrap 1.7.0"
            version = result.stdout.strip().split()[-1]
            return version
        raise RuntimeError("Failed to get bwrap version")
    except subprocess.TimeoutExpired:
        raise RuntimeError("Timeout getting bwrap version")
    except FileNotFoundError:
        raise RuntimeError("bwrap not found")



class BubblewrapRunner:
    """
    Executes code using Bubblewrap for process isolation.

    Provides the second layer of security isolation within the container,
    using Linux namespaces and seccomp filters.
    """

    def __init__(self, workspace_path: Path):
        """
        Initialize the Bubblewrap runner.

        Args:
            workspace_path: Path to the workspace directory
        """
        self.workspace_path = workspace_path
        self._base_args = self._build_base_args()

    def _build_base_args(self, container_working_directory: str = "/workspace") -> List[str]:
        """
        Build base Bubblewrap arguments for isolation.

        Returns:
            List of bwrap command arguments
        """
        dependency_path = settings.dependency_install_path
        return [
            "bwrap",
            # Filesystem isolation
            "--ro-bind", "/usr", "/usr",
            "--ro-bind", "/lib", "/lib",
            "--ro-bind", "/lib64", "/lib64",
            "--ro-bind", "/bin", "/bin",
            "--ro-bind", "/sbin", "/sbin",
            # Session-installed third-party dependencies remain read-only during execution.
            "--ro-bind", dependency_path, dependency_path,
            # Workspace (writable)
            "--bind", str(self.workspace_path), "/workspace",
            "--chdir", container_working_directory,
            # Temporary directory (tmpfs)
            "--tmpfs", "/tmp",
            # Minimal /proc and /dev
            "--proc", "/proc",
            "--dev", "/dev",
            # Namespace isolation
            "--unshare-all",
            "--unshare-net",  # Network isolation
            # Process management
            "--die-with-parent",
            "--new-session",
            # Environment
            "--clearenv",
            "--setenv", "PATH", "/usr/local/bin:/usr/bin:/bin",
            "--setenv", "HOME", "/workspace",
            "--setenv", "TMPDIR", "/tmp",
            # Security (Note: --cap-drop and --no-new-privs not available in bwrap 0.11.0)
            # These are handled by container-level security (non-privileged user, namespaces)
        ]

    async def execute(self, execution: Execution) -> ExecutionResult:
        """
        Execute code within Bubblewrap isolation.

        Args:
            execution: Execution entity with code and context

        Returns:
            ExecutionResult with stdout, stderr, exit code, timing, return_value, and metrics
        """
        start_time = time.perf_counter()
        start_cpu = time.process_time()
        logger.info(
            "Executing code in bwrap",
            execution_id=execution.execution_id,
            language=execution.language,
        )

        try:
            # Build language-specific command and environment
            cmd, env_args = self._build_command(execution)
            if env_args:
                cmd = self._inject_env_args(cmd, env_args)

            # Prepare environment with event data
            env = os.environ.copy()
            env["PYTHONPATH"] = self._build_pythonpath(env.get("PYTHONPATH"))

            # Execute with asyncio subprocess (non-blocking)
            process = await asyncio.create_subprocess_exec(
                *cmd,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
                cwd=str(self.workspace_path),
                env=env,
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

            # Try to collect memory metrics
            # Note: This is a placeholder - actual implementation would monitor /proc/{pid}/status
            # For now, we don't have direct access to the child process's memory usage

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
                "Execution completed",
                execution_id=execution.execution_id,
                exit_code=process.returncode,
                duration_ms=duration_ms,
            )

            return execution_result

        except asyncio.TimeoutError:
            duration_ms = (time.perf_counter() - start_time) * 1000
            logger.warning("Bwrap execution timeout", execution_id=execution.execution_id)

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
                "Bwrap execution error",
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

    def _generate_wrapper_code(self, user_code: str) -> str:
        """
        Generate wrapper code for Lambda-style handler execution.

        Args:
            user_code: User's Python code

        Returns:
            Complete wrapper script
        """
        return f"""
import json
import sys
import os

# User code
{user_code}

# Read event from environment variable
try:
    event_json = os.environ.get("EVENT_JSON", "{{}}")
    event = json.loads(event_json) if event_json.strip() else {{}}
except json.JSONDecodeError as e:
    print(f"Error parsing event JSON: {{e}}", file=sys.stderr)
    sys.exit(1)

# Call handler
try:
    if 'handler' not in globals():
        raise ValueError("必须定义 handler(event) 函数")

    result = handler(event)

    # Output result with markers
    print("\n===SANDBOX_RESULT===")
    print(json.dumps(result))
    print("\n===SANDBOX_RESULT_END===")

except Exception as e:
    import traceback
    print("\n===SANDBOX_ERROR===")
    print(traceback.format_exc())
    print("\n===SANDBOX_ERROR_END===")
    sys.exit(1)
"""

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

    def _build_command(self, execution: Execution) -> tuple[List[str], dict]:
        """
        Build the complete command for executing code.

        Args:
            execution: Execution entity

        Returns:
            Tuple of (command list, environment variables dict)
        """
        lang = execution.language.lower()

        if lang == "python":
            return self._build_python_command(execution)
        elif lang in ["javascript", "nodejs", "node"]:
            return self._build_node_command(execution)
        elif lang in ["bash", "shell"]:
            return self._build_shell_command(execution)
        else:
            raise ValueError(f"Unsupported language: {execution.language}")

    def _build_python_command(self, execution: Execution) -> tuple[List[str], dict]:
        """
        Build command for Python execution using fileless approach.

        Uses python3 -c to execute code directly in memory with Lambda-style wrapper.
        """
        execution.context.resolve_working_directory_path()
        # Generate wrapper code for handler execution
        wrapper_code = self._generate_wrapper_code(execution.code)

        cmd = self._build_base_args(execution.context.container_working_directory()) + [
            "--",
            "python3",
            "-c",
            wrapper_code,
        ]
        return cmd, self._build_execution_env(execution)

    def _build_node_command(self, execution: Execution) -> tuple[List[str], dict]:
        """Build command for Node.js execution."""
        execution.context.resolve_working_directory_path()
        # Wrap user code in AWS Lambda handler pattern
        wrapper_code = f'''
{execution.code}

const eventJson = process.env.EVENT_JSON || '{{}}';
const event = JSON.parse(eventJson);

const result = handler(event, {{}});

console.log('===SANDBOX_RESULT===' + JSON.stringify(result) + '===SANDBOX_RESULT_END===');
'''

        # Write code to temporary file
        code_file = self.workspace_path / "user_code.js"
        code_file.write_text(wrapper_code)

        cmd = self._build_base_args(execution.context.container_working_directory()) + [
            "--ro-bind", str(code_file), "/workspace/user_code.js",
            "--",
            "node",
            "/workspace/user_code.js",
        ]
        return cmd, self._build_execution_env(execution)

    def _build_shell_command(self, execution: Execution) -> tuple[List[str], dict]:
        """Build command for shell execution."""
        resolved_cwd = execution.context.resolve_working_directory_path()
        normalized_code = normalize_shell_code(execution.code, resolved_cwd)
        cmd = self._build_base_args(execution.context.container_working_directory()) + [
            "--",
            "bash",
            "-c",
            normalized_code,
        ]
        return cmd, self._build_execution_env(execution)

    def _build_pythonpath(self, existing_pythonpath: str | None) -> str:
        dependency_path = settings.dependency_install_path
        if existing_pythonpath:
            return f"{dependency_path}:{existing_pythonpath}"
        return dependency_path

    def _build_execution_env(self, execution: Execution) -> dict[str, str]:
        env_args: dict[str, str] = {
            "PYTHONPATH": self._build_pythonpath(os.environ.get("PYTHONPATH")),
        }
        if execution.context.event:
            env_args["EVENT_JSON"] = json.dumps(execution.context.event)
        if execution.context.env_vars:
            env_args.update(execution.context.env_vars)
        return env_args

    def _inject_env_args(self, cmd: List[str], env_args: dict[str, str]) -> List[str]:
        if "--" not in cmd:
            return cmd
        separator_index = cmd.index("--")
        prefix = cmd[:separator_index]
        suffix = cmd[separator_index:]
        for key, value in env_args.items():
            prefix.extend(["--setenv", key, value])
        return prefix + suffix
