#!/usr/bin/python3
# -*- coding:utf-8 -*-

import json
import os
import httpx

from common.logger import logger
from errors.errors import InternalErrException
from common.configs import docset_configs

class DocSetAdapter:
    def __init__(self) -> None:
        self.private_addr = "http://{}:{}".format(docset_configs["private_host"], docset_configs["private_port"])

    async def get_fulltext_doc(self, docid: str):
        target = "{}/api/docset/v1/subdoc".format(self.private_addr)
        data = {"doc_id": docid, "type": "full_text"}
        headers = {'Content-Type': 'application/json'}
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.request("POST", target, json=data, headers=headers)
            if resp.status_code != httpx.codes.OK:
                logger.warn("[get_fulltext_doc] get fulltext doc error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.json()
