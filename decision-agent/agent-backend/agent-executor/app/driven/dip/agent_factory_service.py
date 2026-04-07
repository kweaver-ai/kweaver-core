import sys

import aiohttp
from circuitbreaker import circuit

import app.common.stand_log as log_oper
from app.common import errors
from app.common.config import Config
from app.common.errors import CodeException
from app.common.stand_log import StandLogger
from app.utils.common import GetFailureThreshold, GetRecoveryTimeout
from app.utils.observability.trace_wrapper import internal_span
from app.infra.common.infra_constant.const import HTTP_REQUEST_TIMEOUT
from opentelemetry.trace import Span


class AgentFactoryService:
    def __init__(self):
        self._host = Config.services.agent_factory.host
        self._port = Config.services.agent_factory.port
        self._basic_url = "http://{}:{}".format(self._host, self._port)
        self.headers = {}

    def set_headers(self, headers):
        self.headers = headers

    @circuit(
        failure_threshold=GetFailureThreshold(), recovery_timeout=GetRecoveryTimeout()
    )
    async def get_agent_config(self, agent_id) -> dict:
        """
        获取agent配置
        """
        url = self._basic_url + "/api/agent-factory/internal/v3/agent/{}".format(
            agent_id
        )

        timeout = aiohttp.ClientTimeout(total=HTTP_REQUEST_TIMEOUT)
        async with aiohttp.ClientSession(
            headers=self.headers, timeout=timeout
        ) as session:
            async with session.get(url) as response:
                if response.status != 200:
                    err = self._host + " get_agent_config error: {}".format(
                        await response.text()
                    )
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ExternalServiceError(), err)

                res = await response.json()
        """
        res = {
    "id": "01JS4Z1DGERV8HBY140PS1PH8N",
    "key": "01JS4Z1DGERV8HBY140N8TQK9T",
    "is_built_in": 0,
    "name": "简单问答Agent",
    "profile": "支持与大模型进行对话问答，不支持数据源召回，不支持临时区，不支持调用工具",
    "avatar_type": 1,
    "avatar": "1",
    "product_id": 1,
    "product_name": "AnyShare",
    "config": {
        "input": {
            "fields": [
                {
                    "name": "query",
                    "type": "string"
                }
            ],
            "rewrite": null,
            "augment": null,
            "is_temp_zone_enabled": 0,
            "temp_zone_config": null
        },
        "system_prompt": "/prompt/$query -> answer",
        "dolphin": "string",
        "is_dolphin_mode": 1,
        "data_source": null,
        "tools": [],
        "llms": [
            {
                "is_default": true,
                "llm_config": {
                    "id": "1916319990936637440",
                    "name": "Tome-pro",
                    "temperature": 0,
                    "top_p": 0.95,
                    "top_k": 1,
                    "frequency_penalty": 0,
                    "presence_penalty": 0,
                    "max_tokens": 500
                }
            }
        ],
        "is_data_flow_set_enabled": 0,
        "opening_remark_config": null,
        "preset_questions": null,
        "output": {
            "answer_variables": [
                "answer"
            ]
        }
    },
    "status": "published"
}
        """

        return res

    @circuit(
        failure_threshold=GetFailureThreshold(), recovery_timeout=GetRecoveryTimeout()
    )
    async def get_agent_config_by_key(self, agent_key) -> dict:
        """
        根据agent_key获取agent配置
        """
        url = self._basic_url + "/api/agent-factory/internal/v3/agent/by-key/{}".format(
            agent_key
        )

        timeout = aiohttp.ClientTimeout(total=HTTP_REQUEST_TIMEOUT)
        async with aiohttp.ClientSession(
            headers=self.headers, timeout=timeout
        ) as session:
            async with session.get(url) as response:
                if response.status != 200:
                    err = self._host + " get_agent_config_by_key error: {}".format(
                        await response.text()
                    )
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ExternalServiceError(), err)

                res = await response.json()
                return res

    @circuit(
        failure_threshold=GetFailureThreshold(), recovery_timeout=GetRecoveryTimeout()
    )
    @internal_span()
    async def check_agent_permission(
        self, agent_id, user_id, visitor_type, span: Span = None
    ) -> bool:
        if span and span.is_recording():
            span.set_attribute("agent_id", agent_id)
            span.set_attribute("user_id", user_id)
            span.set_attribute("visitor_type", visitor_type)
        """检查非个人空间下的某个agent是否有执行（使用）权限"""

        if Config.local_dev.is_skip_pms_check:
            return True

        url = (
            self._basic_url + "/api/agent-factory/internal/v3/agent-permission/execute"
        )

        if visitor_type and visitor_type == "app":
            data = {"agent_id": agent_id, "app_account_id": user_id}
        else:  # visitor_type == 'user'
            data = {"agent_id": agent_id, "user_id": user_id}

        timeout = aiohttp.ClientTimeout(total=HTTP_REQUEST_TIMEOUT)
        async with aiohttp.ClientSession(
            headers=self.headers, timeout=timeout
        ) as session:
            async with session.post(url, json=data) as response:
                if response.status != 200:
                    err = self._host + " check_agent_permission error: {}".format(
                        await response.text()
                    )
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ExternalServiceError(), err)

                res = await response.json()
                return res.get("is_allowed", False)

    @circuit(
        failure_threshold=GetFailureThreshold(), recovery_timeout=GetRecoveryTimeout()
    )
    async def get_agent_config_by_agent_id_and_version(
        self, agent_id: str, version: str
    ) -> dict:
        """
        根据agent_id和version获取agent配置
        """
        url = (
            self._basic_url
            + "/api/agent-factory/internal/v3/agent-market/agent/{}/version/{}".format(
                agent_id, version
            )
        )

        timeout = aiohttp.ClientTimeout(total=HTTP_REQUEST_TIMEOUT)
        async with aiohttp.ClientSession(
            headers=self.headers, timeout=timeout
        ) as session:
            async with session.get(url) as response:
                if response.status != 200:
                    err = (
                        self._host
                        + " get_agent_config_by_agent_id_and_version error: {}".format(
                            await response.text()
                        )
                    )
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ExternalServiceError(), err)

                res = await response.json()
                return res


agent_factory_service = AgentFactoryService()
