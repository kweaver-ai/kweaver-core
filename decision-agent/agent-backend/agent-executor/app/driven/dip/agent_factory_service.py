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

        # Skill API is served by agent-operator-integration, not agent-factory.
        # skill_api.yaml server: /api/agent-operator-integration/internal-v1
        _oi_host = Config.services.agent_operator_integration.host
        _oi_port = Config.services.agent_operator_integration.port
        self._skill_api_base_url = (
            "http://{}:{}/api/agent-operator-integration/internal-v1".format(
                _oi_host, _oi_port
            )
        )

    def set_headers(self, headers):
        self.headers = headers

    # Hop-by-hop headers that must NOT be forwarded to downstream service calls.
    # Forwarding Content-Length from the executor's incoming request body to a
    # downstream POST request with a completely different (smaller) body causes
    # FastAPI/uvicorn to fail to read the downstream request body, resulting in
    # a 422 validation error before the handler is ever entered.
    _HOP_BY_HOP_HEADERS = frozenset(
        {
            "content-length",
            "transfer-encoding",
            "connection",
            "keep-alive",
            "host",
            "upgrade",
            "proxy-authenticate",
            "proxy-authorization",
            "te",
            "trailers",
        }
    )

    def _effective_headers(self, request_headers: dict) -> dict:
        """Return per-call headers when provided, falling back to the shared singleton headers.

        skill API methods accept an explicit ``request_headers`` argument so that
        concurrent async invocations from different requests each carry their own
        identity (x-user-account-id, x-business-domain, etc.) without mutating the
        singleton's ``self.headers`` between an ``await`` suspension point and the
        next I/O call.

        Hop-by-hop headers (Content-Length, Transfer-Encoding, Host, etc.) are
        stripped from forwarded headers so that aiohttp can set the correct
        values for each downstream request body independently.
        """
        source = request_headers if request_headers is not None else self.headers
        return {k: v for k, v in source.items() if k.lower() not in self._HOP_BY_HOP_HEADERS}

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


    # =====================================================================
    # Skill API methods
    # These are the only HTTP entry points for remote skill access.
    # =====================================================================

    async def get_skill_content(
        self, skill_id: str, request_headers: dict = None
    ) -> dict:
        """Call GET /skills/{skill_id}/content.

        Returns the SKILL.md download URL, full file list, and skill status.

        Args:
            skill_id: Execution-factory skill identifier
            request_headers: Per-request headers that override the singleton's
                self.headers for this call, preventing cross-request header bleed
                under concurrent async execution.

        Returns:
            Response data dict containing 'url', 'files', 'status', 'skill_id'
        """
        url = self._skill_api_base_url + f"/skills/{skill_id}/content"
        timeout = aiohttp.ClientTimeout(total=HTTP_REQUEST_TIMEOUT)
        StandLogger.info(f"get_skill_content header={self._effective_headers(request_headers)},url={url}")
        async with aiohttp.ClientSession(
            headers=self._effective_headers(request_headers), timeout=timeout
        ) as session:
            async with session.get(url) as response:
                if response.status != 200:
                    err = f"get_skill_content error [{response.status}]: {await response.text()}"
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ExternalServiceError(), err)
                res = await response.json()
                # Log the complete response for debugging
                StandLogger.info(
                    f"get_skill_content complete response for skill '{skill_id}': {res}"
                )
                data = res.get("data", {})
                # Log the raw response for debugging
                StandLogger.info(
                    f"get_skill_content raw response for skill '{skill_id}': "
                    f"code={res.get('code')}, msg={res.get('msg')!r}, "
                    f"data keys={list(data.keys())}, "
                    f"url_present={'url' in data}, url_value={data.get('url')!r}, "
                    f"url_empty={data.get('url') == ''}, url_is_none={data.get('url') is None}"
                )
                return data

    async def read_skill_file_meta(
        self, skill_id: str, rel_path: str, request_headers: dict = None
    ) -> dict:
        """Call POST /skills/{skill_id}/files/read.

        Returns the download URL and metadata for a single skill file.
        Does NOT return file content directly — the caller must download
        the returned URL separately.

        Args:
            skill_id: Execution-factory skill identifier
            rel_path: Relative path inside the skill package (e.g. references/foo.md)
            request_headers: Per-request headers that override self.headers for this
                call (see get_skill_content for rationale).

        Returns:
            Response data dict containing 'url', 'rel_path', 'mime_type', 'file_type'
        """
        url = self._skill_api_base_url + f"/skills/{skill_id}/files/read"
        payload = {"rel_path": rel_path}
        timeout = aiohttp.ClientTimeout(total=HTTP_REQUEST_TIMEOUT)
        StandLogger.info(f"read_skill_file_meta header={self._effective_headers(request_headers)},url={url},payload={payload}")
        async with aiohttp.ClientSession(
            headers=self._effective_headers(request_headers), timeout=timeout
        ) as session:
            async with session.post(url, json=payload) as response:
                if response.status != 200:
                    err = (
                        f"read_skill_file_meta error [{response.status}]: "
                        f"{await response.text()}"
                    )
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ExternalServiceError(), err)
                res = await response.json()
                # Log the complete response for debugging
                StandLogger.info(
                    f"read_skill_file_meta complete response for skill '{skill_id}', path '{rel_path}': {res}"
                )
                data = res.get("data", {})
                # Log the raw response for debugging
                StandLogger.info(
                    f"read_skill_file_meta raw response for skill '{skill_id}', path '{rel_path}': "
                    f"code={res.get('code')}, msg={res.get('msg')!r}, "
                    f"data keys={list(data.keys())}, "
                    f"url_value={data.get('url')!r}, url_present={'url' in data}"
                )
                return data

    async def download_text_by_url(self, url: str) -> str:
        """Download raw text content from an object-storage URL.

        Used to fetch SKILL.md and other text files after obtaining their
        download URLs from the skill API.

        Args:
            url: Pre-signed or direct download URL

        Returns:
            Raw text content as a string
        """
        timeout = aiohttp.ClientTimeout(total=HTTP_REQUEST_TIMEOUT)
        async with aiohttp.ClientSession(timeout=timeout) as session:
            async with session.get(url) as response:
                if response.status != 200:
                    err = f"download_text_by_url error [{response.status}]: {url}"
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ExternalServiceError(), err)
                return await response.text(encoding="utf-8")

    # Allowlist of file extensions that are guaranteed to be readable as plain
    # UTF-8 text.  Any file outside this set is treated as binary and rejected
    # before a download is attempted, satisfying the design's "text files only"
    # boundary (§ 5.3 of the factory_skill_execution_design).
    _TEXT_EXTENSIONS = frozenset(
        [".md", ".txt", ".json", ".yaml", ".yml", ".py", ".sh", ".js", ".ts"]
    )

    # MIME type prefixes that indicate binary (non-text) content.  If the API
    # returns a mime_type matching any of these, the file is rejected even when
    # the extension check would otherwise pass.
    _BINARY_MIME_PREFIXES = (
        "image/",
        "audio/",
        "video/",
        "application/octet-stream",
        "application/zip",
        "application/x-tar",
        "application/x-gzip",
        "application/pdf",
        "font/",
    )

    async def read_downloaded_skill_text(
        self,
        url: str,
        file_path: str,
        mime_type: str = None,
        file_type: str = None,
    ) -> str:
        """Download a skill file and return its text content.

        Only files with extensions in _TEXT_EXTENSIONS are allowed (design § 5.3).
        Binary files are rejected before any network request is made.

        Args:
            url: Download URL returned by read_skill_file_meta or get_skill_content
            file_path: Original relative path, used for extension detection
            mime_type: Optional MIME type hint from the API response
            file_type: Optional file type hint from the API response

        Returns:
            File content as a plain text string

        Raises:
            CodeException: When the file extension or MIME type indicates binary
                content that is not supported at this stage.
        """
        import os

        ext = os.path.splitext(file_path)[1].lower()

        if ext and ext not in self._TEXT_EXTENSIONS:
            err = (
                f"Unsupported file type '{ext}' for skill file '{file_path}'. "
                f"Only text formats are supported: {sorted(self._TEXT_EXTENSIONS)}"
            )
            error_log = log_oper.get_error_log(err, sys._getframe())
            StandLogger.error(error_log, log_oper.SYSTEM_LOG)
            raise CodeException(errors.ParamException(), err)

        if mime_type:
            for binary_prefix in self._BINARY_MIME_PREFIXES:
                if mime_type.lower().startswith(binary_prefix):
                    err = (
                        f"Binary MIME type '{mime_type}' is not supported for skill "
                        f"file '{file_path}'. Only plain text files can be read."
                    )
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ParamException(), err)
        StandLogger.info(f"read_downloaded_skill_text url={url}")
        return await self.download_text_by_url(url)

    async def execute_skill_script(
        self,
        skill_id: str,
        entry_shell: str,
        extra: dict = None,
        request_headers: dict = None,
    ) -> dict:
        """Call POST /skills/{skill_id}/execute (skill_api.yaml).

        Passes entry_shell (obtained by the LLM from SKILL.md) directly to
        the factory API.  The API response uses 'execution_time' (ms); this
        method normalises it to 'duration_ms' for the runtime layer.

        Args:
            skill_id: Execution-factory skill identifier
            entry_shell: Shell command from SKILL.md, e.g. 'python scripts/foo.py'
            extra: Optional dict; honours 'timeout' key (int, seconds).
            request_headers: Per-request headers that override self.headers.

        Returns:
            Response data dict containing stdout, stderr, exit_code,
            duration_ms (normalised from execution_time), and other fields.
        """
        url = self._skill_api_base_url + f"/skills/{skill_id}/execute"
        payload: dict = {"entry_shell": entry_shell}
        if extra and isinstance(extra.get("timeout"), int):
            payload["timeout"] = extra["timeout"]

        timeout = aiohttp.ClientTimeout(total=HTTP_REQUEST_TIMEOUT)
        StandLogger.info(f"execute_skill_script header={self._effective_headers(request_headers)},url={url},payload={payload}")
        async with aiohttp.ClientSession(
            headers=self._effective_headers(request_headers), timeout=timeout
        ) as session:
            async with session.post(url, json=payload) as response:
                if response.status != 200:
                    err = (
                        f"execute_skill_script error [{response.status}]: "
                        f"{await response.text()}"
                    )
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ExternalServiceError(), err)
                res = await response.json()
                # Log the complete response for debugging
                StandLogger.info(
                    f"execute_skill_script complete response for skill '{skill_id}': {res}"
                )
                data = res.get("data", {})
                # Log the raw response for debugging
                StandLogger.info(
                    f"execute_skill_script raw response for skill '{skill_id}': "
                    f"code={res.get('code')}, msg={res.get('msg')!r}, "
                    f"data keys={list(data.keys())}, "
                    f"exit_code={data.get('exit_code')}, "
                    f"execution_time={data.get('execution_time')}"
                )
                # Normalise API field name -> contract field name expected by runtime
                if "execution_time" in data and "duration_ms" not in data:
                    data["duration_ms"] = data["execution_time"]
                return data


agent_factory_service = AgentFactoryService()
