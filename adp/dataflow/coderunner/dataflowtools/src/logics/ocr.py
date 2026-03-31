#!/usr/bin/python3
# -*- coding:utf-8 -*-

import base64
import io
import json
import time
from typing import List
from common.constants import *
from utils.utils import *
from drivenadapters.efast import EfastAdapter
from drivenadapters.ocr import OCRAdapter
from errors.errors import *
from utils.utils import run_in_thread

SUPPORTED_SCENE_TYPES=('general', 'eleinvoice', 'idcard')
SUPPORTED_FILE_TYPES={'.jpg': 'image/jpeg', '.jpeg': 'image/jpeg', '.png': 'image/png', '.bmp': 'image/bmp', '.tif': 'image/tiff', '.tiff': 'image/tiff', '.pdf': 'application/pdf'}
BUILT_IN_SUPPORTED_FILE_TYPES={'.jpg': 'image/jpeg', '.jpeg': 'image/jpeg', '.png': 'image/png', '.bmp': 'image/bmp', '.tif': 'image/tiff', '.tiff': 'image/tiff'}

class OCRService:
    def __init__(self, host=None, access_token=None) -> None:
        self.ocr = OCRAdapter()
        self.efast = EfastAdapter(host=host, access_token=access_token)

    async def buildin_ocr(self, requestBody) -> str:
        # 内置ocr逻辑
        doc_id, scene, rec_type = requestBody["docid"], requestBody["scene"], requestBody["rec_type"]
        content, doc_name = await run_in_thread(self.efast.file_download, doc_id)
        encoded_content = base64.b64encode(io.BytesIO(content).read()).decode()
        resp_content = b''
        doc_type = split_file_type(doc_name)
        if doc_type not in BUILT_IN_SUPPORTED_FILE_TYPES:
            raise BadParameterException(detail="unsupported file type, type: {}, supported types: {}".format(doc_type, BUILT_IN_SUPPORTED_FILE_TYPES))
        if rec_type == SUPPORTED_SCENE_TYPES[0]:
            resp_content = await self.ocr.general(scene=scene, encoded_content=encoded_content)
        elif rec_type == SUPPORTED_SCENE_TYPES[1] or rec_type == SUPPORTED_SCENE_TYPES[2]:
            resp_content = await self.ocr.ticket(scene=scene, encoded_content=encoded_content)
        else:
            raise BadParameterException(detail="unsupported rec type, type: {}, supported types: {}".format(rec_type, SUPPORTED_SCENE_TYPES))
        return self.parse_buildin_data(resp_content=resp_content, rec_type=rec_type)

    async def external_ocr(self, requestBody) -> str:
        doc_id, scene, rec_type = requestBody["docid"], requestBody["scene"], requestBody["rec_type"]
        content, doc_name = await run_in_thread(self.efast.file_download, doc_id)
        # 外置ocr逻辑
        doc_type = split_file_type(doc_name)
        if doc_type not in SUPPORTED_FILE_TYPES:
            raise BadParameterException(detail="unsupported file type, type: {}, supported types: {}".format(doc_type, SUPPORTED_FILE_TYPES))
        task_type, uri = requestBody['task_type'], requestBody['uri']
        resp_content = await self.ocr.advanced_ocr_rec(taskType=task_type, uri=uri, scene=scene, doc_name=doc_name, file_input=io.BytesIO(content), doc_type=SUPPORTED_FILE_TYPES[doc_type])
        resp_content_obj = json.loads(resp_content)
        resp_result = {'task_id': resp_content_obj["data"], "rec_type": rec_type}
        return json.dumps(resp_result)

    def parse_buildin_data(self, resp_content: bytes, rec_type: str) -> str:
        result = {}
        params_map = {}
        resp_content_obj = json.loads(resp_content)
        if rec_type == SUPPORTED_SCENE_TYPES[0]:
            resp_content_obj['data'] = {'json': resp_content_obj['data']['json']}
            results = [{'result': json.dumps(resp_content_obj, ensure_ascii=False)}]
            subTaskList = {"subTaskList": results}
            return json.dumps(subTaskList, ensure_ascii=False)
        elif rec_type == SUPPORTED_SCENE_TYPES[1]:
            params_map = eleinvoice_map
        elif rec_type == SUPPORTED_SCENE_TYPES[2]:
            params_map = idcard_map
        raw_result = resp_content_obj['data']['json']['raw_result']
        keys = raw_result['keys']
        texts = raw_result['texts']
        for index, key in enumerate(keys):
            if key in params_map:
                result[params_map[key]] = ''.join(texts[index])
            else:
                result[params_map[key]] = ''
        return json.dumps(result)

    async def get_ocr_result(self,task_id: str, rec_type: str):
        resp_content = await self.ocr.get_ocr_result(task_id)
        resp_content_obj = json.loads(resp_content)
        state = resp_content_obj['data']['state']

        if state == 32 or state == 40:
            raise InternalErrException()
        elif state == 30:
            return self.parse_external_data(resp_content=resp_content, rec_type=rec_type)
        else:
            raise RequestProcessingException()

    #  分开处理高级和普通ocr的数据
    def parse_external_data(self, resp_content: bytes, rec_type: str) -> str:
        result = {}
        params_map = {}
        resp_content_obj = json.loads(resp_content)
        subTaskLists = resp_content_obj['data']['subTaskList']
        if rec_type == SUPPORTED_SCENE_TYPES[0]:
            new_subTaskLists = []
            for subTaskList in subTaskLists:
                sub_result = json.loads(subTaskList['result'])
                sub_result['data'] = {'json':sub_result['data']['json']}
                subTaskList['result'] = json.dumps(sub_result, ensure_ascii=False)
                new_subTaskLists.append(subTaskList)
            resp_content_obj['data']['subTaskList'] = new_subTaskLists
            return json.dumps(resp_content_obj['data'], ensure_ascii=False)
        elif rec_type == SUPPORTED_SCENE_TYPES[1]:
            # 电子发票返回体
            params_map = eleinvoice_map
        elif rec_type == SUPPORTED_SCENE_TYPES[2]:
            # 身份证返回体
            params_map = idcard_map
        resp_content_obj = json.loads(subTaskLists[0]['result'])
        raw_result = resp_content_obj['data']['json']['raw_result']
        keys = raw_result['keys']
        texts = raw_result['texts']
        for index, key in enumerate(keys):
            if key in params_map:
                result[params_map[key]] = ''.join(texts[index])
            else:
                result[params_map[key]] = ''
        return json.dumps(result)


    async def delete_ocr_task(self, task_id: List):
        await self.ocr.delete_ocr_task(task_nos=task_id)