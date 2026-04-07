#!/usr/bin/python3
# -*- coding:utf-8 -*-

import io
import json
import math
import os
from typing import Tuple
import httpx
import time

from errors.errors import InternalErrException, NotFoundException
from utils.utils import *
from common.logger import logger
from common.configs import oss_management_configs

class OSSManagement:
    def __init__(self) -> None:
        self.addr = f"http://{oss_management_configs['private_host']}:{oss_management_configs['private_port']}"
        self.file_size_threshold = 20 * 1024 * 1024

    async def get_default_oss(self):
        target = "{}/api/ossgateway/v1/default-storage".format(self.addr)
        headers = {'Content-Type': 'application/json'}
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.request("GET", target, headers=headers)
            if resp.status_code != httpx.codes.OK:
                if resp.status_code == httpx.codes.NOT_FOUND:
                    raise NotFoundException()
                logger.warn("[get_default_oss] get default oss error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.content

    async def get_object_storage_infos(self, biz_type = "as"):
        target = "{}/api/ossgateway/v1/objectstorageinfo?app={}".format(self.addr, biz_type)
        headers = {'Content-Type': 'application/json'}
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.request("GET", target, headers=headers)
            if resp.status_code != httpx.codes.OK:
                logger.warn("[get_object_storage_infos] get object storage infos error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.content

    async def get_availd_oss(self) -> str:
        try:
            res = await self.get_default_oss()
            resJson = json.loads(res)
            return resJson.get("storage_id")
        except NotFoundException:
            res = await self.get_object_storage_infos(biz_type="as")
            json_array = json.loads(res)
            for item in json_array:
                if not item.get("enabled"):
                    continue
                return item.get("id")
            raise NotFoundException(detail="no availd oss")
    
    async def upload_file(self, oss_id: str, key: str, internal_request: bool, file: io.BufferedReader, size: int):
        if size < self.file_size_threshold:
            await self.simple_upload(oss_id, key, internal_request, file)
        else:
            await self.multi_upload_file(oss_id, key, internal_request, file, size)

    # simple_upload 简单上传文件
    async def simple_upload(self, oss_id: str, key: str, internal_request: bool, file: io.BufferedReader):
        target = f"{self.addr}/api/ossgateway/v1/upload/{oss_id}/{key}?request_method=PUT&internal_request={internal_request}"
        request_body = {}

        try:
            data = file.read()
        except Exception as err:
            raise InternalErrException(detail=str(err))

        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.get(target)
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[simple_upload] simple upload error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            request_body = resp.json()

            resp = await client.request(request_body['method'], request_body['url'], headers=request_body['headers'], data=data)
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[simple_upload] simple upload error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)

    # multi_upload_file 分片上传文件
    async def multi_upload_file(self, oss_id: str, key: str, internal_request: bool, file: io.BufferedReader, size: int):
        part_min_size = 20 * 1024 * 1024
        part_max_size = 20 * 1024 * 1024
        part_max_num = 10000
        part_size = 0
        part_count = 0
        file_size = size
        e_tags = {}

        # Calculate total parts
        while True:
            part_size += part_min_size
            if part_size > part_max_size:
                raise InternalErrException(detail="file too long")
            part_count = file_size // part_size
            if file_size == 0 or file_size % part_size != 0:
                part_count += 1
            if part_count <= part_max_num:
                break

        # Get upload info
        upload_info = await self.init_multi_upload(oss_id, key, internal_request, size)
        part_file = bytearray(part_size)
        upload_id = upload_info.get("upload_id")

        for i in range(1, part_count + 1):
            part_file_size = file.readinto(part_file)
            if part_file_size < 0:
                raise InternalErrException(detail="read file error")

            e_tag = await self.upload_frag_file(oss_id, key, upload_id, i, 1, bytes(part_file[:part_file_size]), internal_request)
            if e_tag is None:
                raise InternalErrException(detail="upload frag file error")

            # Store eTag information
            e_tags[str(i)] = e_tag

        # Complete upload protocol
        # if not await self.complete_multi_upload(oss_id, key, upload_id, e_tags, internal_request):
            # raise InternalErrException(detail="complete multi upload error")
        await self.complete_multi_upload(oss_id, key, upload_id, e_tags, internal_request)

    # init_multi_upload 初始化分片上传
    async def init_multi_upload(self, oss_id: str, key: str, internal_request: bool, size: int) -> dict:
        target = f"{self.addr}/api/ossgateway/v1/initmultiupload/{oss_id}/{key}?size={size}&internal_request={internal_request}"
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.get(target)
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[init_multi_upload] init multi upload error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.json()

    # upload_frag_file 上传分片文件
    async def upload_frag_file(self, oss_id: str, key: str, upload_id: str, part_id: int, part_num: int, part_file: bytes, internal_request: bool) -> str:
        target = f"{self.addr}/api/ossgateway/v1/uploadpart/{oss_id}/{key}?part_id={part_id}&part_num={part_num}&upload_id={upload_id}&internal_request={internal_request}"
        # 获取开始上传协议，分块大小和上传id
        out = {}
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.get(target)
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                raise InternalErrException("[upload_frag_file] get upload frag file info error, status: {}, detail: {}".format(resp.status_code, resp.text))
            out = resp.json()
            
            str_part_id = str(part_id)
            request_body = out[str_part_id]
            resp = await client.request(request_body['method'], request_body['url'], headers=request_body['headers'], data=part_file)
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                raise InternalErrException("[upload_frag_file] upload frag file error, status: {}, detail: {}".format(resp.status_code, resp.text))
            return resp.headers.get("Etag").strip('"')

    # complete_multi_upload 完成分片上传
    async def complete_multi_upload(self, oss_id: str, key: str, upload_id: str, e_tags: dict, internal_request: bool) -> bool:
        target = f"{self.addr}/api/ossgateway/v1/completeupload/{oss_id}/{key}?upload_id={upload_id}&internal_request={internal_request}"
        request_body = {}
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.post(target, json=e_tags)
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[complete_multi_upload] get complete upload info error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            request_body = resp.json()

            resp = await client.request(request_body['method'], request_body['url'], headers=request_body['headers'], data=request_body['request_body'])
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[complete_multi_upload] complete multi upload error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return True
    
    async def get_object_meta(self, oss_id: str, key: str, internal_request: bool) -> int:
        target = f"{self.addr}/api/ossgateway/v1/head/{oss_id}/{key}?internal_request={internal_request}"
        request_body = {}

        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.get(target)
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[get_object_meta] get object meta info error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            request_body = resp.json()

            resp = await client.request(request_body['method'], request_body['url'], headers=request_body['headers'], data=request_body.get('body'))
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[get_object_meta] get object meta error, status: {}, detail: {}".format(resp.status_code, resp.text))
                if resp.status_code == httpx.codes.NOT_FOUND:
                    raise NotFoundException()
                raise InternalErrException(detail=resp.text)   

            content_length = -1
            if 'Content-Length' in resp.headers:
                content_length = int(resp.headers['Content-Length'])
            
            return content_length
        
    async def __download_file_by_frag(self, oss_id: str, key: str, internal_request: bool, start: int, end: int, part_size: int, file_size: int) -> Tuple[bytearray, bool]:
        buff = bytearray()
        target = f"{self.addr}/api/ossgateway/v1/download/{oss_id}/{key}?internal_request={internal_request}"
        request_body = {}

        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.get(target)
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[__download_file_by_frag] get download frag file info error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            request_body = resp.json()

            # 文件小于4M时，下载原本大小
            if file_size <= end:
                end = file_size - 1

            body_byte = bytes()
            if not request_body.get('body') is None:
                body_byte = bytes(request_body['body'])

            # 分片范围
            request_body['headers']["Range"] = f"bytes={start}-{end}"
            resp = await client.request(request_body['method'], request_body['url'], headers=request_body['headers'], data=body_byte)
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:       
                logger.warn("[__download_file_by_frag] download file by frag error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            buff = resp.content


            # 检验下载的分片文件是否完整
            is_loss = is_byte_loss(len(buff), start, end, part_size, file_size)
            return buff, is_loss
    
    async def download_file(self, oss_id: str, key: str, internal_request: bool) -> bytearray:
        i = 1
        start = 0
        end = 4194303
        part_size = 4194304
        retry = 0
        buff = bytearray()

        # 获取文件大小
        file_size = await self.get_object_meta(oss_id, key, internal_request)
        if file_size == -1:
            raise InternalErrException(detail="get object meta info failed")

        # 下载次数
        download_time = math.ceil(file_size / part_size)

        while i <= download_time:
            data, is_loss = await self.__download_file_by_frag(oss_id, key, internal_request, start, end, part_size, file_size)
            if is_loss:
                # 下载过程出错，直接返回
                return buff, InternalErrException(detail="fragment download file byte loss")
            
            # 判断字节存在缺失，重新下载，重试次数3次
            if is_loss:
                retry += 1
                if retry == 3:
                    raise InternalErrException(detail="fragment download file byte loss")
                time.sleep(0.1)
                continue

            # 下载成功重置重试次数
            retry = 0
            # 添加下载内容进缓冲区
            buff.extend(data)
            start = end + 1
            end += part_size

            # 最后一次下载
            if i == download_time - 1:
                end = file_size
            i += 1

        if len(buff) != file_size:
            raise InternalErrException(detail=f"download file may be broken, filesize: {file_size}, download size: {len(buff)}")

        return buff
    
    async def download_file_to_local(self, oss_id: str, key: str, internal_request: bool, file_path: str) -> int:
        data = await self.download_file(oss_id, key, internal_request)
        file_size = len(data)
        if file_size == -1:
            raise InternalErrException(detail="file download failed when get object meta info")
        
        try:
            with open(file_path, 'wb') as f:
                f.write(data)
        except Exception as e:
            raise InternalErrException(detail=f"file download failed when create tmp file, detail: {str(e)}")

        return file_size