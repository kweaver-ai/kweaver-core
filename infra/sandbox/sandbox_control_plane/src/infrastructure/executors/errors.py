"""
执行器相关错误

定义与执行器通信时可能出现的错误。
"""


class ExecutorError(Exception):
    """执行器错误基类"""

    pass


class ExecutorConnectionError(ExecutorError):
    """无法连接到执行器"""

    def __init__(self, executor_url: str, reason: str = ""):
        self.executor_url = executor_url
        self.reason = reason
        super().__init__(f"Failed to connect to executor at {executor_url}: {reason}")


class ExecutorTimeoutError(ExecutorError):
    """执行器响应超时"""

    def __init__(self, executor_url: str, timeout: float):
        self.executor_url = executor_url
        self.timeout = timeout
        super().__init__(f"Executor at {executor_url} timed out after {timeout}s")


class ExecutorUnavailableError(ExecutorError):
    """执行器不可用（ unhealthy）"""

    def __init__(self, executor_url: str, status: str = ""):
        self.executor_url = executor_url
        self.status = status
        super().__init__(f"Executor at {executor_url} is unavailable: {status}")


class ExecutorResponseError(ExecutorError):
    """执行器返回错误响应"""

    def __init__(self, executor_url: str, status_code: int, message: str = ""):
        self.executor_url = executor_url
        self.status_code = status_code
        self.message = message
        super().__init__(
            f"Executor at {executor_url} returned error {status_code}: {message}"
        )


class ExecutorValidationError(ExecutorError):
    """执行器请求验证失败"""

    def __init__(self, executor_url: str, validation_errors: list):
        self.executor_url = executor_url
        self.validation_errors = validation_errors
        super().__init__(
            f"Executor at {executor_url} rejected request: {validation_errors}"
        )
