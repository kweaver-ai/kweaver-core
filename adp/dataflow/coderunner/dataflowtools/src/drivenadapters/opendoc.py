#!/usr/bin/python3
# -*- coding:utf-8 -*-

import json
import os
import httpx

from errors.errors import InternalErrException
from common.logger import logger
from common.configs import open_doc_configs

class OpenDocAdapter:
    def __init__(self) -> None:
        print(open_doc_configs["host"], open_doc_configs["port"])
        self.addr = "http://{}:{}".format(open_doc_configs["host"], open_doc_configs["port"])

    async def file_download(self, doc: list, token: str):
        print(self.addr)
        print(token)
        target = "{}/api/open-doc/v1/file-download".format(self.addr)
        data = {"doc": doc}
        headers = {'Content-Type': 'application/json', 'Authorization': token}
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.request("POST", target, json=data, headers=headers)
            if resp.status_code != httpx.codes.OK:
                logger.warn("[file_download] download file error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.content
