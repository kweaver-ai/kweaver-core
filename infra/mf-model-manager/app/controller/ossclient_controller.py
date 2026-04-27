import requests

from app.logs.stand_log import StandLogger
from pydantic import BaseModel
import abc

# 对象存储交互请求信息
class OSSRequestInfo(BaseModel):
    Method:     str = None
    URL:        str = None
    Headers:    dict = None
    BodyDict:   dict = None
    BodyStr:    str = None

class DrivenOsOperation(metaclass=abc.ABCMeta):
    # 删除对象存储中文件
    @abc.abstractmethod
    def GetFileInfo(self, ossRequestInfo:OSSRequestInfo): pass

class OssOperation(DrivenOsOperation):
    start = 0
    block_size = 1024*1024*5
    def GetFileInfo(self, ossRequestInfo:OSSRequestInfo,streamDownload: bool):

        url = ossRequestInfo.URL
        method = ossRequestInfo.Method
        headers = ossRequestInfo.Headers
        if streamDownload:
            headers['Range'] = f"bytes={self.start}-{self.block_size}"
        response = requests.request(method=method, url=url, headers=headers, verify=False)
        if response.status_code < 200 or response.status_code >= 300:
            err = "OsOperation GetFileInfo errors: {}".format(response.content)
            StandLogger.error(err)
            raise Exception(err)

        return response