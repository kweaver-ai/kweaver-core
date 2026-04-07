import os
from typing import Dict

import httpx

from errors.errors import InternalErrException
from common.logger import logger
from common.configs import hydra_configs
from common.configs import is_auth_enable


class Hydra:
    def __init__(self) -> None:
        self.addr = f"http://{hydra_configs['admin_host']}:{hydra_configs['admin_port']}"

    async def check_token(self, token: str) -> Dict:
        if not is_auth_enable():
            return {
                "active": True,
                "scope": "",
                "client_id": "mock-client-id",
                "sub": "mock-user-id",
                "ext": {"login_ip": "127.0.0.1", "udid": "mock-udid", "client_type": "web"},
                "visitor_type": "realname",
            }

        target = f"{self.addr}/admin/oauth2/introspect"
        headers = {'Content-Type': 'application/x-www-form-urlencoded'}
        body = 'token=' + token
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.request("POST", target, data=body, headers=headers)
            if resp.status_code < httpx.codes.OK or resp.status_code > httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[check_token] check token error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException()
            return resp.json()
