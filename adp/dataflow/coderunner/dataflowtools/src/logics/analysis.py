#!/usr/bin/python3
# -*- coding:utf-8 -*-

import requests
import re
import httpx
from common.constants import *
from utils.utils import *
from drivenadapters.opendoc import OpenDocAdapter
from drivenadapters.docset import DocSetAdapter
from errors.errors import *
from common.logger import logger

class AnalysisService:
    def __init__(self) -> None:
        self.opendoc = OpenDocAdapter()
        self.docset = DocSetAdapter()

    async def parse_doc(self, docid, token) -> dict:
        res = await self.docset.get_fulltext_doc(docid)
        if res["status"] != "completed":
            return res
        try:
            async with httpx.AsyncClient(timeout=900, verify=False) as client:
                resp = await client.get(res["url"])
            if resp.status_code != httpx.codes.OK:
                logger.warn("[parse_doc] parse doc error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)

            # 获取文件内容
            con = resp.content
            decoded_string = con.decode('utf-8')
            pattern = r"\{\{ (.+?) \}\}"
            matches = re.findall(pattern, decoded_string)
            slots = []
            for index, item in enumerate(matches):
                slots.append({
                    "id": index,
                    "value": item
                })
            result = {
                "docid": res["doc_id"],
                "version": res["version"],
                "status": res["status"],
                "slots": slots
            }
            return result
        except requests.RequestException as e:
            logger.warn(f"请求错误: {e}")
            return None
        
    # async def 
