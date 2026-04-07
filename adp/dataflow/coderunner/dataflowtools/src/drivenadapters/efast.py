import requests
from typing import Tuple
from aishu_anyshare_api.api_client import ApiClient
from aishu_anyshare_api_efast.models import *
from utils.utils import *
from errors.errors import InternalErrException
from common.logger import logger


class EfastAdapter:
    def __init__(self,host,access_token) -> None:
        self.DEFAULT_CSF_LEVEL = 0
        self.client = ApiClient(host = host, access_token = access_token, verify_ssl = False)

    def file_info(self, doc_id: str):
        resp = self.client.efast.efast_v1_file_attribute_post(FileAttributeReq.from_dict({"docid":doc_id}))
        doc_info = {"size": resp.size, "rev": resp.rev, "name": resp.name}
        return doc_info

    def gns_to_path(self, doc_id: str):
        resp = self.client.efast.efast_v1_file_convertpath_post(FileConvertpathReq.from_dict({"docid":doc_id}))
        return resp.namepath

    def path_to_gns(self, path: str):
        resp = self.client.efast.efast_v1_file_getinfobypath_post(FileGetinfobypathReq.from_dict({"namepath":path}))
        return resp.docid

    def file_download(self, doc_id: str) -> Tuple[bytes, str]:
        resp = self.client.efast.efast_v1_file_osdownload_post(FileOsdownloadReq.from_dict({"docid":doc_id}))
        method, url, headers, doc_name = resp.authrequest[0], resp.authrequest[1], resp.authrequest[2:], resp.name
        header = {}
        for val in headers:
            arr = val.split(": ")
            header[arr[0]] = arr[1].strip()
        resp = requests.request(method, url, headers=header, verify=False)
        if resp.status_code != requests.codes.ok:
            logger.warning("[file_download] download file failed, status: {}, detail: {}".format(resp.status_code, resp.text))
            raise InternalErrException(detail=resp.text)
        return resp.content, doc_name

    def predupload(self, data_length: int, slice_md5: str) -> FilePreduploadRes:
        params = FilePreduploadReq.from_dict({"length": data_length,"slice_md5": slice_md5})
        res = self.client.efast.efast_v1_file_predupload_post(params)
        return res

    def dupload(self, crc32: str, docid: str, data_length: int, slice_md5: str, doc_name: str, ondup: int) -> str:
        timestamp = generate_timestamp()
        # 如果匹配秒传，则直接使用秒传接口创建文件
        params = FileDuploadReq.from_dict({"client_mtime": timestamp, "crc32": crc32, "csflevel": self.DEFAULT_CSF_LEVEL,
            "docid": docid, "length": data_length, "md5": slice_md5, "name": doc_name, "ondup": ondup})
        res = self.client.efast.efast_v1_file_dupload_post(params)
        if not res.success:
            raise InternalErrException(detail="[dupload] dupload failed")
        return res.docid # type: ignore

    def file_upload(self, docid: str, data_length: int, doc_name: str, file_input, file_metadata: dict, ondup: int) -> str:
        # 开始上传文件协议
        timestamp = generate_timestamp()
        params = FileOsbeginuploadReq.from_dict({"client_mtime": timestamp, "docid": docid, "length": data_length, "name": doc_name, "ondup": ondup, "reqmethod": "POST"})
        res = self.client.efast.efast_v1_file_osbeginupload_post(params)
        authrequest, new_doc_id, rev = res.authrequest, res.docid, res.rev
        payload = {}
        for val in authrequest[2:]:
            vals = val.split(": ")
            if len(vals) < 2:
                continue
            payload[vals[0]] = vals[1]

        # 打开文件流（支持路径或流对象）
        if isinstance(file_input, str):
            file_obj = open(file_input, 'rb')
            should_close = True
        else:
            file_obj = file_input
            should_close = False
        try:
            # 创建form-data文件二进制数据
            files = [('file', (file_metadata.get('name'), file_obj, file_metadata.get('mime_type')))]
            resp = requests.request(authrequest[0], authrequest[1], data=payload, files=files, verify=False)
            if resp.status_code != requests.codes.no_content:
                logger.warning("[file_upload] upload to oss failed, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)

            # 完成上传文件协议
            params = FileOsenduploadReq.from_dict({"csflevel": self.DEFAULT_CSF_LEVEL, "docid": new_doc_id, "rev": rev})
            res = self.client.efast.efast_v1_file_osendupload_post_with_http_info(params)
            if res.status_code != requests.codes.ok:
                logger.warning("[file_upload] oss end upload failed, status: {}, detail: {}".format(res.status_code, res.raw_data))
                raise InternalErrException(detail=res.raw_data)
        finally:
            if should_close:
                file_obj.close()
        return new_doc_id
