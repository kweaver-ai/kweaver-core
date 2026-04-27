#!/usr/bin/python3
# -*- coding:utf-8 -*-

import json
import os
import httpx

from errors.errors import InternalErrException
from common.logger import logger
from common.configs import t4th_configs


class OCRAdapter:
    def __init__(self) -> None:
        self.addr = "{}://{}:{}".format(t4th_configs["protocol"], t4th_configs["host"], t4th_configs["port"])

    async def general(self, scene: str, encoded_content: str):
        target = "{}/lab/ocr/predict/general".format(self.addr)
        data = {'scene': scene, "image": encoded_content}
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.request("POST", target, json=data)
            if resp.status_code != httpx.codes.OK:
                logger.warn("[general] Built-in ocr recognition error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.content

    async def ticket(self, scene: str, encoded_content: str):
        target = "{}/lab/ocr/predict/ticket".format(self.addr)
        data = {'scene': scene, "image": encoded_content}
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.request("POST", target, json=data)
            if resp.status_code != httpx.codes.OK:
                logger.warn("[ticket] Built-in ocr recognition error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.content

    async def advanced_ocr_rec(self, taskType, uri, scene, doc_name, file_input, doc_type):
        target = "{}/service/submitTask?taskType={}&uri={}&scene={}".format(self.addr, taskType, uri, scene)
        files = [('file', (doc_name, file_input, doc_type))]
        async with httpx.AsyncClient(verify=False) as client:
            resp = await client.request("POST", target, files=files)
            if resp.status_code != httpx.codes.OK:
                logger.warn("[advanced_OCR_rec] Advanced ocr recognition error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            resp_content_obj = json.loads(resp.content)
            if resp_content_obj['msgCode'] == str(httpx.codes.INTERNAL_SERVER_ERROR):
                logger.warn("[advanced_OCR_rec] Advanced ocr recognition error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.content

    async def get_ocr_result(self, task_no):
        target = "{}/service/getResultDetail?taskNo={}".format(self.addr, task_no)
        async with httpx.AsyncClient(verify=False) as client:
            resp = await client.request("GET", target)
            if resp.status_code != httpx.codes.OK:
                logger.warn("[get_ocr_result] get ocr recognition result error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            resp_content_obj = json.loads(resp.content)
            if resp_content_obj['msgCode'] == str(httpx.codes.INTERNAL_SERVER_ERROR):
                logger.warn("[get_ocr_result] get ocr recognition result error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.content

    async def delete_ocr_task(self, task_nos):
        target = "{}/service/batchDeleteTask".format(self.addr)
        json_data = json.dumps(task_nos)
        headers = {'Content-Type': 'application/json'}
        async with httpx.AsyncClient(verify=False) as client:
            resp = await client.request("POST", target, json=json_data, headers=headers)
            if resp.status_code != httpx.codes.OK:
                logger.warn("[delete_ocr_task] delete ocr task error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.content
