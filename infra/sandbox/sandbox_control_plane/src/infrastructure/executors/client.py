"""
执行器 HTTP 客户端

用于与沙箱容器内的执行器进行 HTTP 通信。
"""
import asyncio
import logging
from typing import Optional

import httpx

from src.infrastructure.executors.dto import (
    ExecutorExecuteRequest,
    ExecutorExecuteResponse,
    ExecutorHealthResponse,
    ExecutorSyncSessionConfigRequest,
    ExecutorSyncSessionConfigResponse,
)
from src.infrastructure.executors.errors import (
    ExecutorConnectionError,
    ExecutorTimeoutError,
    ExecutorUnavailableError,
    ExecutorResponseError,
    ExecutorValidationError,
)

logger = logging.getLogger(__name__)


class ExecutorClient:
    """
    执行器 HTTP 客户端

    通过 HTTP 与运行在容器内的 sandbox-executor 通信。
    """

    def __init__(
        self,
        timeout: float = 30.0,
        max_retries: int = 3,
        retry_delay: float = 0.5,
    ):
        """
        初始化执行器客户端

        Args:
            timeout: 请求超时时间（秒）
            max_retries: 最大重试次数
            retry_delay: 重试延迟（秒）
        """
        self._timeout = timeout
        self._max_retries = max_retries
        self._retry_delay = retry_delay
        self._client: Optional[httpx.AsyncClient] = None

    async def __aenter__(self):
        """进入上下文管理器"""
        self._client = httpx.AsyncClient(timeout=self._timeout)
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """退出上下文管理器"""
        if self._client:
            await self._client.aclose()

    def _get_client(self) -> httpx.AsyncClient:
        """获取 HTTP 客户端实例"""
        if self._client is None:
            self._client = httpx.AsyncClient(timeout=self._timeout)
        return self._client

    async def submit_execution(
        self,
        executor_url: str,
        execution_id: str,
        session_id: str,
        code: str,
        language: str,
        event: dict,
        timeout: int,
        env_vars: dict,
    ) -> str:
        """
        提交执行请求到执行器

        Args:
            executor_url: 执行器 URL (e.g., "http://container-name:8080")
            execution_id: 执行 ID
            session_id: 会话 ID
            code: 要执行的代码
            language: 编程语言
            event: 事件数据
            timeout: 超时时间（秒）
            env_vars: 环境变量

        Returns:
            execution_id: 执行任务 ID

        Raises:
            ExecutorConnectionError: 无法连接到执行器
            ExecutorTimeoutError: 执行器响应超时
            ExecutorValidationError: 请求验证失败
            ExecutorResponseError: 执行器返回错误
        """
        client = self._get_client()
        url = f"{executor_url}/execute"

        request = ExecutorExecuteRequest(
            execution_id=execution_id,
            session_id=session_id,
            code=code,
            language=language,
            event=event,
            timeout=timeout,
            env_vars=env_vars,
        )

        logger.info(f"Submitting execution request: executor_url={executor_url}, execution_id={execution_id}, language={language}")

        for attempt in range(self._max_retries):
            try:
                response = await client.post(
                    url,
                    json=request.model_dump(),
                    headers={"Content-Type": "application/json"},
                )

                if response.status_code == 200:
                    result = ExecutorExecuteResponse(**response.json())
                    logger.info(f"Execution submitted successfully: execution_id={execution_id}, status={result.status}")
                    return result.execution_id

                elif response.status_code == 400:
                    # Validation error - don't retry
                    raise ExecutorValidationError(
                        executor_url, response.json().get("errors", [])
                    )

                elif response.status_code >= 500:
                    # Server error - retry
                    if attempt < self._max_retries - 1:
                        logger.warning(f"Executor returned {response.status_code}, retrying... attempt={attempt + 1}")
                        await asyncio.sleep(self._retry_delay * (attempt + 1))
                        continue
                    else:
                        raise ExecutorResponseError(
                            executor_url,
                            response.status_code,
                            response.text,
                        )

                else:
                    raise ExecutorResponseError(
                        executor_url,
                        response.status_code,
                        response.text,
                    )

            except httpx.ConnectError as e:
                if attempt < self._max_retries - 1:
                    logger.warning(f"Failed to connect to executor, retrying... executor_url={executor_url}, attempt={attempt + 1}, error={e}")
                    await asyncio.sleep(self._retry_delay * (attempt + 1))
                    continue
                else:
                    raise ExecutorConnectionError(executor_url, str(e))

            except httpx.TimeoutException as e:
                raise ExecutorTimeoutError(executor_url, self._timeout)

            except httpx.HTTPStatusError as e:
                raise ExecutorResponseError(executor_url, e.response.status_code, str(e))

        # Should not reach here
        raise ExecutorConnectionError(executor_url, "Max retries exceeded")

    async def health_check(self, executor_url: str) -> ExecutorHealthResponse:
        """
        检查执行器健康状态

        Args:
            executor_url: 执行器 URL

        Returns:
            健康状态响应

        Raises:
            ExecutorConnectionError: 无法连接到执行器
            ExecutorUnavailableError: 执行器不健康
        """
        client = self._get_client()
        url = f"{executor_url}/health"

        try:
            response = await client.get(url)

            if response.status_code == 200:
                return ExecutorHealthResponse(**response.json())
            else:
                raise ExecutorUnavailableError(
                    executor_url, f"status_code={response.status_code}"
                )

        except httpx.ConnectError as e:
            raise ExecutorConnectionError(executor_url, str(e))
        except httpx.TimeoutException as e:
            raise ExecutorTimeoutError(executor_url, self._timeout)

    async def sync_session_config(
        self,
        executor_url: str,
        session_id: str,
        language_runtime: str,
        python_package_index_url: str,
        dependencies: list[str],
        sync_mode: str,
    ) -> ExecutorSyncSessionConfigResponse:
        """同步会话依赖配置到 executor。"""
        client = self._get_client()
        url = f"{executor_url}/internal/session-config/sync"
        request = ExecutorSyncSessionConfigRequest(
            session_id=session_id,
            language_runtime=language_runtime,
            python_package_index_url=python_package_index_url,
            dependencies=dependencies,
            sync_mode=sync_mode,
        )

        try:
            response = await client.post(
                url,
                json=request.model_dump(),
                headers={"Content-Type": "application/json"},
            )
        except httpx.ConnectError as e:
            raise ExecutorConnectionError(executor_url, str(e))
        except httpx.TimeoutException:
            raise ExecutorTimeoutError(executor_url, self._timeout)

        if response.status_code == 200:
            return ExecutorSyncSessionConfigResponse(**response.json())
        if response.status_code == 400:
            raise ExecutorValidationError(executor_url, response.json())
        if response.status_code in (422, 500):
            raise ExecutorResponseError(executor_url, response.status_code, response.text)
        if response.status_code == 503:
            raise ExecutorUnavailableError(executor_url, response.text)

        raise ExecutorResponseError(executor_url, response.status_code, response.text)

    async def close(self) -> None:
        """关闭客户端"""
        if self._client:
            await self._client.aclose()
            self._client = None
