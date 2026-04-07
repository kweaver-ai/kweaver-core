import os
import sys
from typing import Tuple
import requests
from circuitbreaker import circuit

from app.utils.stand_log import StandLogger

class OSSRequestInfo:
    Method:     str = None
    URL:        str = None
    Headers:    dict = None
    BodyDict:   dict = None
    BodyStr:    str = None

class OSSGateway:
    _host = os.getenv("OSSGATEWAYMANAGER_PRIVATE_HOST", "ossgatewaymanager-private")
    _port = os.getenv("OSSGATEWAYMANAGER_PRIVATE_PORT", "9002")
    _basicUrl = f"http://{_host}:{_port}"
    _bucketID = ""

    # 下载
    @circuit(failure_threshold=5, recovery_timeout=60)
    def ownload(self, key: str, name: str):
        url = "%s/api/ossgateway/v1/download/%s/%s?user=%s&type=query_string&internal_request=false&save_name=%s" % \
              (self._basicUrl, self._bucketID, key, self._bucketID, name)
        response = requests.get(url, verify=False)
        if response.status_code != 200:
            StandLogger.error("下载失败")
            raise Exception("Builder.OssGateway.Download.InternalError")
        res = response.json()
        ossRequestInfo = OSSRequestInfo()
        ossRequestInfo.Method = res.get("method", None)
        ossRequestInfo.URL = res.get("url", None)
        ossRequestInfo.Headers = res.get("headers", None)
        return ossRequestInfo