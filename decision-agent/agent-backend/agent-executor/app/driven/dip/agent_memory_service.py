from http import HTTPStatus
import asyncio
from typing import Optional, Any
import json
import aiohttp
from aiohttp import ClientTimeout
from circuitbreaker import circuit

from app.common import errors
from app.common.config import Config
from app.common.errors import CodeException
from app.utils.common import GetFailureThreshold, GetRecoveryTimeout


class AgentMemoryService:
    def __init__(self):
        self._host = Config.services.agent_memory.host
        self._port = Config.services.agent_memory.port
        self._basic_url = "http://{}:{}".format(self._host, self._port)

    @circuit(
        failure_threshold=GetFailureThreshold(), recovery_timeout=GetRecoveryTimeout()
    )
    async def build_memory(
        self,
        messages: list,
        user_id: Optional[str] = None,
        agent_id: Optional[str] = None,
        run_id: Optional[str] = None,
        metadata: Optional[dict] = None,
        headers: Optional[dict] = None,
    ):
        url = "{basic_url}/api/agent-memory/internal/v1/memory".format(
            basic_url=self._basic_url
        )

        body: dict[str, Any] = {
            "messages": messages,
        }

        if user_id:
            body["user_id"] = user_id

        if agent_id:
            body["agent_id"] = agent_id

        if run_id:
            body["run_id"] = run_id

        if metadata:
            body["metadata"] = metadata

        try:
            async with aiohttp.ClientSession(headers=headers) as session:
                timeout = ClientTimeout(total=60)

                async with session.post(
                    url, json=body, ssl=False, timeout=timeout
                ) as response:
                    if response.status != HTTPStatus.NO_CONTENT:
                        response_text = await response.text()
                        err = f"{self._host} build_memory error: status={response.status}, response={response_text}"
                        raise CodeException(errors.ExternalServiceError(), err)
        except asyncio.TimeoutError as e:
            # 处理超时异常
            err = (
                f"build_memory request timeout: url={url}, timeout=60s, error={str(e)}"
            )
            raise CodeException(errors.ExternalServiceError(), err)

        except aiohttp.ClientError as e:
            # 处理其他 aiohttp 客户端错误
            err = f"build_memory client error: url={url}, error={str(e)}"
            raise CodeException(errors.ExternalServiceError(), err)

        except Exception as e:
            # 处理其他未预期的异常
            err = f"build_memory unexpected error: url={url}, request={json.dumps(body, ensure_ascii=False)}, error={str(e)}"
            raise CodeException(errors.ExternalServiceError(), err)


agent_memory_service = AgentMemoryService()
