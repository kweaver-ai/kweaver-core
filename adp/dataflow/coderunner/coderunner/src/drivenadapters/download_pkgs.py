import httpx
import os
from errors.errors import InternalErrException

class DownLoadPkgs:
    host = os.getenv("DATA_FLOW_TOOLS_PRIVATE_HOST", "dataflowtools-private")   
    port = os.getenv("DATA_FLOW_TOOLS_PRIVATE_PORT", "8086")

    def __init__(self, ):
        self.addr = f"http://{self.host}:{self.port}"

    async def download_pkg(self, pkg_name: str) -> bytes:
        target = f"{self.addr}/api/coderunner/v1/py-packages/{pkg_name}"
        headers = {'Content-Type': 'application/json'}
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.request("GET", target, headers=headers)
            if resp.status_code != httpx.codes.OK:
                raise InternalErrException("[download_pkg] download package error, status: {}, detail: {}".format(resp.status_code, resp.text))
            return resp.content
