"""
Simple subprocess isolation runner.

This is a fallback isolation mechanism for development environments
where Bubblewrap is not available (e.g., macOS Docker containers).

WARNING: This provides NO security isolation and should ONLY be used
for development purposes.
"""
import asyncio
import time
import json
import logging
from pathlib import Path
from typing import List, Tuple

from executor.domain.entities import Execution
from executor.domain.value_objects import ExecutionResult, ExecutionStatus, ExecutionMetrics
from executor.infrastructure.config import settings
from executor.infrastructure.isolation.code_wrapper import normalize_shell_code

logger = logging.getLogger(__name__)


class SubprocessRunner:
    """
    Executes code using subprocess without isolation.

    WARNING: This is a DEVELOPMENT-ONLY fallback that provides NO
    security isolation. Never use this in production.
    """

    def __init__(self, workspace_path: Path):
        """
        Initialize the subprocess runner.

        Args:
            workspace_path: Path to the workspace directory (can be S3 path)
        """
        # Store original workspace path for reference
        self.original_workspace_path = workspace_path
        workspace_str = str(workspace_path)

        # If workspace is S3 path, use local /workspace directory for execution
        # S3 paths are handled by the artifact storage layer
        if workspace_str.startswith("s3:/") or workspace_str.startswith("s3://"):
            logger.warning(f"Workspace is S3 path ({workspace_str}), using local /workspace for execution")
            self.workspace_path = Path("/workspace")
        else:
            self.workspace_path = workspace_path

        logger.warning(f"SubprocessRunner initialized - NO SECURITY ISOLATION, workspace_path={self.workspace_path}")

    async def execute(self, execution: Execution) -> ExecutionResult:
        """
        Execute code without isolation.

        Args:
            execution: Execution entity with code and context

        Returns:
            ExecutionResult with stdout, stderr, exit code, timing, return_value, and metrics
        """
        start_time = time.perf_counter()
        start_cpu = time.process_time()
        logger.info(f"Executing code without isolation (DEVELOPMENT MODE), execution_id={execution.execution_id}, language={execution.language}")

        try:
            # Build language-specific command and environment
            cmd, env_args = self._build_command(execution)
            cwd_path = execution.context.resolve_working_directory_path()

            # Execute using asyncio subprocess
            process = await asyncio.create_subprocess_exec(
                *cmd,
                cwd=str(cwd_path),
                env=env_args,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
            )

            stdout, stderr = await asyncio.wait_for(
                process.communicate(),
                timeout=30  # Default timeout, outer layer handles actual timeout
            )

            duration = time.perf_counter() - start_time
            cpu_time = time.process_time() - start_cpu

            stdout_str = stdout.decode("utf-8", errors="replace")
            stderr_str = stderr.decode("utf-8", errors="replace")

            # Parse return value from stdout (for Lambda handlers)
            return_value = None
            language = execution.language.lower()
            if language in ("python", "python3") and process.returncode == 0:
                try:
                    # Lambda handler writes return value as JSON to stdout
                    # But there might be print() statements before the return value
                    # Strategy: extract the last valid JSON as return_value, keep everything else as stdout
                    lines = stdout_str.strip().split('\n')

                    # Try to find the last valid JSON line as the return value
                    for i in range(len(lines) - 1, -1, -1):
                        line = lines[i].strip()
                        if line:  # Non-empty line
                            try:
                                return_value = json.loads(line)
                                # Found valid JSON - remove this line from stdout
                                stdout_str = '\n'.join(lines[:i])
                                if stdout_str:
                                    # Ensure stdout ends with a single newline if there were previous lines
                                    if not stdout_str.endswith('\n'):
                                        stdout_str += '\n'
                                break
                            except json.JSONDecodeError:
                                # Not valid JSON, continue searching
                                continue
                    else:
                        # No valid JSON found, keep the original stdout
                        pass
                except Exception as e:
                    # If parsing fails, keep the original stdout
                    logger.debug(f"Failed to parse return value from stdout: {e}")
                    pass

            logger.info(f"Execution completed, execution_id={execution.execution_id}, exit_code={process.returncode}, duration_ms={duration * 1000}")

            return ExecutionResult(
                status=ExecutionStatus.COMPLETED if process.returncode == 0 else ExecutionStatus.FAILED,
                stdout=stdout_str,
                stderr=stderr_str,
                exit_code=process.returncode,
                execution_time_ms=duration * 1000,
                return_value=return_value,
                metrics=ExecutionMetrics(
                    duration_ms=duration * 1000,
                    cpu_time_ms=cpu_time * 1000,
                    peak_memory_mb=None,
                    io_read_bytes=None,
                    io_write_bytes=None,
                ),
            )

        except asyncio.TimeoutError:
            logger.warning(f"Execution timed out, execution_id={execution.execution_id}")
            return ExecutionResult(
                status=ExecutionStatus.TIMEOUT,
                stdout="",
                stderr="Execution timed out",
                exit_code=124,
                execution_time_ms=30000,
                return_value=None,
                metrics=None,
            )
        except Exception as e:
            logger.error(f"Execution failed, execution_id={execution.execution_id}, error={e}")
            return ExecutionResult(
                status=ExecutionStatus.ERROR,
                stdout="",
                stderr=str(e),
                exit_code=-1,
                execution_time_ms=0,
                return_value=None,
                metrics=None,
            )

    def _build_command(self, execution: Execution) -> Tuple[List[str], dict]:
        """
        Build language-specific command and environment.

        Args:
            execution: Execution entity with code and context

        Returns:
            Tuple of (command_list, environment_dict)
        """
        import os
        language = execution.language.lower()
        code = execution.code

        # Build environment variables - inherit from current process
        env_args = os.environ.copy()
        # Override specific variables
        env_args.update({
            "PATH": "/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin",
            "HOME": str(self.workspace_path),
            "USER": "sandbox",
            "WORKSPACE_PATH": str(self.workspace_path),
            "PYTHONPATH": self._build_pythonpath(env_args.get("PYTHONPATH")),
        })

        # Add user-provided environment variables
        if execution.context.env_vars:
            env_args.update(execution.context.env_vars)

        # Get event data as JSON
        event_json = json.dumps(execution.context.event or {})

        # Language-specific commands
        if language in ("python", "python3"):
            # For Python Lambda handlers, wrap the code to:
            # 1. Define the handler function from user code
            # 2. Call the handler with the event data
            # 3. Serialize and print the return value as JSON
            wrapped_code = f'''
import json
import sys

# User's code
{code}

# Execute the Lambda handler
if __name__ == "__main__":
    event_data = json.loads({json.dumps(event_json)})
    result = handler(event_data)
    print(json.dumps(result))
'''
            cmd = ["python3", "-c", wrapped_code]
        elif language == "javascript":
            # For JavaScript handlers (Node.js)
            wrapped_code = f'''
// User's code
{code}

// Execute the Lambda handler
const eventData = {json.dumps(event_json)};
const result = handler(eventData);
console.log(JSON.stringify(result));
'''
            cmd = ["node", "-e", wrapped_code]
        elif language == "shell":
            code = normalize_shell_code(code, execution.context.resolve_working_directory_path())
            cmd = ["sh", "-c", code]
        else:
            raise ValueError(f"Unsupported language: {language}")

        return cmd, env_args

    def _build_pythonpath(self, existing_pythonpath: str | None) -> str:
        dependency_path = settings.dependency_install_path
        if existing_pythonpath:
            return f"{dependency_path}:{existing_pythonpath}"
        return dependency_path
