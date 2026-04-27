import os

import httpx

from errors.errors import InternalErrException
from common.logger import logger


class CodeRunner:
    host  = os.getenv("CODE_RUNNER_PRIVATE_HOST", "coderunner-private")
    port = os.getenv("CODE_RUNNER_PRIVATE_PORT", "8085")
    
    def __init__(self):
        self.addr = f"http://{self.host}:{self.port}"

    async def delete_pkg(self, pkg_name: str):
        target = f"{self.addr}/api/coderunner/v1/py-packages/{pkg_name}"
        headers = {'Content-Type': 'application/json'}
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.delete(target, headers=headers)
            if resp.status_code < httpx.codes.OK or resp.status_code > httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[delete_pkg] delete pkg error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
