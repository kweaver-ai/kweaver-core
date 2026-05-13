import json
import sys
from http import HTTPStatus
from typing import Optional

import aiohttp
from circuitbreaker import circuit

import app.common.stand_log as log_oper
from app.common import errors
from app.common.config import Config
from app.common.errors import CodeException
from app.common.stand_log import StandLogger
from app.common.struct_logger.error_log_class import get_error_log_json
from app.utils.common import GetFailureThreshold, GetRecoveryTimeout
from app.infra.common.infra_constant.const import HTTP_REQUEST_TIMEOUT


class AgentOperatorIntegrationService:
    # Hop-by-hop headers that must NOT be forwarded to downstream service calls.
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

    # Allowlist of file extensions that are guaranteed to be readable as plain UTF-8 text.
    _TEXT_EXTENSIONS = frozenset(
        [".md", ".txt", ".json", ".yaml", ".yml", ".py", ".sh", ".js", ".ts"]
    )

    # MIME type prefixes that indicate binary (non-text) content.
    _BINARY_MIME_PREFIXES = (
        "image/",
        "audio/",
        "video/",
        "application/zip",
        "application/x-tar",
        "application/x-gzip",
        "application/pdf",
        "font/",
    )

    def __init__(self):
        self._host = Config.services.agent_operator_integration.host
        self._port = Config.services.agent_operator_integration.port
        self._basic_url = "http://{}:{}".format(self._host, self._port)
        self._skill_api_base_url = self._basic_url + "/api/agent-operator-integration/internal-v1"

        self.headers = {}

    def set_headers(self, headers):
        self.headers = headers

    @circuit(
        failure_threshold=GetFailureThreshold(), recovery_timeout=GetRecoveryTimeout()
    )
    async def get_tool_box_list(self) -> dict:
        url = "{basic_url}/api/agent-operator-integration/internal-v1/tool-box/list".format(
            basic_url=self._basic_url
        )

        params = {
            "page": 1,
            "page_size": 100,
        }

        async with aiohttp.ClientSession(headers=self.headers) as session:
            async with session.get(url, params=params, ssl=False) as response:
                if response.status != HTTPStatus.OK:
                    err = self._host + " get_tool_box_list error: {}".format(
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
    async def get_tool_list(self, box_id) -> dict:
        url = "{basic_url}/api/agent-operator-integration/internal-v1/tool-box/{box_id}/tools/list".format(
            basic_url=self._basic_url, box_id=box_id
        )

        params = {
            "page": 1,
            "page_size": 100,
        }

        async with aiohttp.ClientSession(headers=self.headers) as session:
            async with session.get(url, params=params, ssl=False) as response:
                if response.status != HTTPStatus.OK:
                    err = self._host + " get_tool_list error: {}".format(
                        await response.text()
                    )
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ExternalServiceError(), err)

                res = await response.json()

        return res

    def get_mock_tool_info(self):
        with open(".local/tool_info.json", "r") as f:
            return json.load(f)

    @circuit(
        failure_threshold=GetFailureThreshold(), recovery_timeout=GetRecoveryTimeout()
    )
    async def get_tool_info(self, box_id, tool_id) -> dict | None:
        # if Config.LOCAL_DEV_AARON:
        #     return self.get_mock_tool_info()

        url = "{basic_url}/api/agent-operator-integration/internal-v1/tool-box/{box_id}/tool/{tool_id}".format(
            basic_url=self._basic_url, box_id=box_id, tool_id=tool_id
        )

        try:
            async with aiohttp.ClientSession(headers=self.headers) as session:
                async with session.get(url, ssl=False) as response:
                    if response.status != HTTPStatus.OK:
                        err = (
                            "tool unavailable, tool_box_id: {box_id}, tool_id: {tool_id}, req_url: {url}"
                            "\nget_tool_info error: {response_text}"
                            "\nresponse_code: {response_status}"
                        ).format(
                            box_id=box_id,
                            tool_id=tool_id,
                            url=url,
                            response_text=await response.text(),
                            response_status=response.status,
                        )

                        error_log = get_error_log_json(err, sys._getframe())
                        StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                        return None

                    res = await response.json()
        except Exception as exc:
            err = (
                "tool unavailable, tool_box_id: {box_id}, tool_id: {tool_id}, req_url: {url}"
                "\nget_tool_info exception: {exception}"
            ).format(box_id=box_id, tool_id=tool_id, url=url, exception=repr(exc))
            error_log = get_error_log_json(err, sys._getframe())
            StandLogger.error(error_log, log_oper.SYSTEM_LOG)
            return None

        return res

    @circuit(
        failure_threshold=GetFailureThreshold(), recovery_timeout=GetRecoveryTimeout()
    )
    async def get_mcp_tools(self, mcp_server_id) -> dict:
        # if Config.is_local_dev():
        #     return {"tools": [{"name": "test"}]}

        url = "{basic_url}/api/agent-operator-integration/internal-v1/mcp/proxy/{mcp_server_id}/tools".format(
            basic_url=self._basic_url, mcp_server_id=mcp_server_id
        )

        async with aiohttp.ClientSession(headers=self.headers) as session:
            async with session.get(url, ssl=False) as response:
                if response.status != HTTPStatus.OK:
                    err = self._host + " get_mcp_tools error: {}".format(
                        await response.text()
                    )
                    error_log = get_error_log_json(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ExternalServiceError(), err)

                res = await response.json()

        return res

    # =========================================================================
    # Skill API methods
    # Encapsulate all remote skill access for load / read / execute operations.
    # =========================================================================

    def _effective_headers(self, request_headers: Optional[dict]) -> dict:
        """Return per-call headers, stripping hop-by-hop entries.

        Accepts an explicit ``request_headers`` dict so that concurrent async
        invocations each carry their own identity (x-user-account-id, etc.)
        without mutating the shared singleton's ``self.headers``.
        """
        source = request_headers if request_headers is not None else self.headers
        return {k: v for k, v in source.items() if k.lower() not in self._HOP_BY_HOP_HEADERS}

    async def get_skill_content(
        self, skill_id: str, request_headers: Optional[dict] = None
    ) -> dict:
        """Call GET /skills/{skill_id}/content.

        Returns the SKILL.md download URL, full file list, and skill status.
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
                StandLogger.info(
                    f"get_skill_content complete response for skill '{skill_id}': {res}"
                )
                data = res.get("data") or res
                StandLogger.info(
                    f"get_skill_content extracted data for skill '{skill_id}': "
                    f"data keys={list(data.keys())}, "
                    f"url_present={'url' in data}, url_value={data.get('url')!r}"
                )
                return data

    async def read_skill_file_meta(
        self, skill_id: str, rel_path: str, request_headers: Optional[dict] = None
    ) -> dict:
        """Call POST /skills/{skill_id}/files/read.

        Returns the download URL and metadata for a single skill file.
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
                StandLogger.info(
                    f"read_skill_file_meta complete response for skill '{skill_id}', path '{rel_path}': {res}"
                )
                data = res.get("data") or res
                StandLogger.info(
                    f"read_skill_file_meta extracted data for skill '{skill_id}', path '{rel_path}': "
                    f"data keys={list(data.keys())}, "
                    f"url_present={'url' in data}, url_value={data.get('url')!r}"
                )
                return data

    async def download_text_by_url(self, url: str) -> str:
        """Download raw text content from an object-storage URL."""
        timeout = aiohttp.ClientTimeout(total=HTTP_REQUEST_TIMEOUT)
        # SSL verification is disabled because the pre-signed URL already carries
        # an HMAC signature guaranteeing authenticity and integrity of the content.
        connector = aiohttp.TCPConnector(ssl=False)
        async with aiohttp.ClientSession(timeout=timeout, connector=connector) as session:
            async with session.get(url) as response:
                if response.status != 200:
                    err = f"download_text_by_url error [{response.status}]: {url}"
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ExternalServiceError(), err)
                return await response.text(encoding="utf-8")

    async def read_downloaded_skill_text(
        self,
        url: str,
        file_path: str,
        mime_type: Optional[str] = None,
        file_type: Optional[str] = None,
    ) -> str:
        """Download a skill file and return its text content.

        Validates the file extension (primary) and MIME type (fallback) before
        downloading to reject binary files early.
        """
        import os

        ext = os.path.splitext(file_path)[1].lower()

        if ext and ext in self._TEXT_EXTENSIONS:
            StandLogger.info(
                f"read_downloaded_skill_text: File '{file_path}' approved by "
                f"extension check (ext={ext}), proceeding with download from url={url}"
            )
            return await self.download_text_by_url(url)

        if ext:
            err = (
                f"Unsupported file type '{ext}' for skill file '{file_path}'. "
                f"Only text formats are supported: {sorted(self._TEXT_EXTENSIONS)}"
            )
            error_log = log_oper.get_error_log(err, sys._getframe())
            StandLogger.error(error_log, log_oper.SYSTEM_LOG)
            raise CodeException(errors.ParamError(), err)

        if mime_type:
            for binary_prefix in self._BINARY_MIME_PREFIXES:
                if mime_type.lower().startswith(binary_prefix):
                    err = (
                        f"Binary MIME type '{mime_type}' is not supported for skill "
                        f"file '{file_path}'. Only plain text files can be read."
                    )
                    error_log = log_oper.get_error_log(err, sys._getframe())
                    StandLogger.error(error_log, log_oper.SYSTEM_LOG)
                    raise CodeException(errors.ParamError(), err)

        StandLogger.info(f"read_downloaded_skill_text url={url}")
        return await self.download_text_by_url(url)

    async def execute_skill_script(
        self,
        skill_id: str,
        entry_shell: str,
        extra: Optional[dict] = None,
        request_headers: Optional[dict] = None,
    ) -> dict:
        """Call POST /skills/{skill_id}/execute.

        Passes entry_shell directly to the factory API and normalises the
        response field 'execution_time' -> 'duration_ms'.
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
                StandLogger.info(
                    f"execute_skill_script complete response for skill '{skill_id}': {res}"
                )
                data = res.get("data") or res
                StandLogger.info(
                    f"execute_skill_script extracted data for skill '{skill_id}': "
                    f"data keys={list(data.keys())}, "
                    f"exit_code={data.get('exit_code')}, "
                    f"execution_time={data.get('execution_time')}"
                )
                if "execution_time" in data and "duration_ms" not in data:
                    data["duration_ms"] = data["execution_time"]
                return data


agent_operator_integration_service = AgentOperatorIntegrationService()
