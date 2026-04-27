#!/usr/bin/python3
# -*- coding:utf-8 -*-

import os
import shutil
import uuid
import tornado.ioloop
from typing import Dict, Any
from utils.redis import RedisLock
from utils.utils import generate_MD5_CRC32
from errors.errors import BadParameterException
from common.configs import MAX_FILE_SIZE, CacheConfig, DataSourceType, FileMimeType, InsertType
from .file_handlers.factory import FileHandlerFactory
from .file_handlers.cache import FileCacheManager
from .file_handlers.template import TemplateManager
from models.file_operator import FileOperatorRequest
from drivenadapters.efast import EfastAdapter


class FileOperationService:
    """文件操作服务"""
    
    def __init__(self, host: str, access_token: str) -> None:
        self.efast = EfastAdapter(host=host, access_token=access_token)
        self.cache_manager = FileCacheManager()
        self.template = TemplateManager()
        self.redis_lock = RedisLock()
    
    async def async_create_file(self, request_body: Dict[str, Any]):
        """异步创建文件"""
        return await tornado.ioloop.IOLoop.current().run_in_executor(
            None, self.create_file, request_body
        )
    
    def create_file(self, request_body: Dict[str, Any]):
        """创建文件"""
        update_request = FileOperatorRequest.from_dict_create(request_body)

        update_request.insert_type = InsertType.COVER.value
        return self._create_new_file_and_update(update_request)

    async def async_update_file(self, request_body: Dict[str, Any]):
        """异步更新文件"""
        return await tornado.ioloop.IOLoop.current().run_in_executor(
            None, self.update_file, request_body
        )
    
    def update_file(self, request_body: Dict[str, Any]):
        """
        更新文件
        
        Args:
            request_body: Dict[str, Any] - 更新请求体
        
        Returns:
           Str - 文件路径
        """
        update_request = FileOperatorRequest.from_dict_update(request_body)

        doc_id = update_request.doc_id

        lock_key = doc_id.split("/")[-1]
        with self.redis_lock.lock(key = lock_key, timeout = 30, blocking=False, auto_renewal=True):
            return self._handle_normal_update(update_request)

    def _create_new_file_and_update(self, update_request: FileOperatorRequest):
        """
        创建新文件并更新内容
        
        Args:
            update_request: 更新请求
            
        Returns:
            创建结果
        """
        try:
            is_directory = self._is_directory(update_request.doc_id)
            if not is_directory:
                raise BadParameterException(detail={
                    "info": f"document is not directory, not support create"
                })
        
            doc_type = update_request.file_type
            handler = FileHandlerFactory.get_handler(doc_type)
        
            # 创建新的空文件
            if update_request.doc_name.lower().endswith(handler.get_file_extension()):
                update_request.doc_name_with_ext = update_request.doc_name
            else:
                update_request.doc_name_with_ext = f"{update_request.doc_name}{handler.get_file_extension()}"

            random_dir = str(uuid.uuid4()).replace('-', '').upper()
            full_dir_path = os.path.join(CacheConfig.CACHE_ROOT_DIR, random_dir)

            if not os.path.exists(full_dir_path):
                os.mkdir(full_dir_path)

            source_path = os.path.join(full_dir_path, update_request.doc_name_with_ext)

            if update_request.source_type == DataSourceType.FULLTEXT.value:
                if (update_request.content is None or 
                    (isinstance(update_request.content, str) and update_request.content.strip() == '') or
                    (isinstance(update_request.content, list) and len(update_request.content) == 0)):
                    self.template.create_empty_file(update_request.file_type, source_path)
                else:
                    update_request.insert_type = update_request.get_actual_insert_type()
                    handler.update(source_path, update_request)
            else:
                self.cache_manager.download_from_url(source_path, update_request.content)
            
            result = self._upload_updated_file(source_path, update_request, False)

            return result
        except Exception as e:
            raise e
        finally:
            try:
                shutil.rmtree(full_dir_path)
            except:
                pass


    def _handle_normal_update(self, update_request: FileOperatorRequest):
        """
        更新文件
        1. 验证参数
        2. 获取文件（缓存或下载）
        3. 调用handler更新
        4. 上传结果
        """
        try:
            doc_id = update_request.doc_id
            
            doc_info = self.efast.file_info(doc_id)
            doc_type = update_request.file_type
            
            if f"{doc_info.get('name')}x".lower().endswith('.xlsx'):
                doc_type = 'xls'

            self._validate_file(doc_info, doc_id, doc_type)

            source_path, update_request.doc_name_with_ext = self.cache_manager.download_from_docid(doc_id, self.efast.file_download)
            
            handler = FileHandlerFactory.get_handler(doc_type)
            handler.update(source_path, update_request)
            
            result = self._upload_updated_file(source_path, update_request, True)

            return result
        finally:
            self.cache_manager.clear_cache(doc_id)


    def _is_directory(self, doc_id: str) -> bool:
        """
        判断doc_id是否为文件夹
        
        Args:
            doc_id: 文档ID
            
        Returns:
            是否为文件夹
        """
        # 调用文件系统接口判断
        info = self.efast.file_info(doc_id)
        if 'size' in info and info.get('size') == -1:
            return True
        
        return False

    def _validate_file(self, doc_info: Dict, doc_id: str, doc_type: str):
        """验证文件大小"""

        doc_info = self.efast.file_info(doc_id)

        if 'size' in doc_info and doc_info.get('size') == -1:
            raise BadParameterException(detail={
                "info": f"document is directory, not support update"
            })

        if not doc_info.get("name").lower().endswith(doc_type.lower()):
            raise BadParameterException(detail={
                "info": f"document type is not match",
                "doc_type": doc_type,
                "name": doc_info.get("name")
            })

        if doc_info["size"] > MAX_FILE_SIZE:
            raise BadParameterException(detail={
                "info": f"File size limited, docid: {doc_id}", 
                "size": doc_info['size'], 
                "max_size": {MAX_FILE_SIZE}
            })
    
    def _upload_updated_file(self, source_path: str, update_request: FileOperatorRequest, is_update: bool = True):
        """上传更新后的文件"""
        with open(source_path, 'rb') as f:
            slice_md5, crc32, data_length = generate_MD5_CRC32(f)
        
        res = self.efast.predupload(data_length=data_length, slice_md5=slice_md5)
        
        parent_doc_id = update_request.doc_id
        if is_update:
            parent_doc_id = os.path.dirname(update_request.doc_id)
        
        if res.match:
            return self.efast.dupload(
                crc32=crc32,
                docid=parent_doc_id,
                data_length=data_length,
                slice_md5=slice_md5,
                doc_name=update_request.doc_name_with_ext,
                ondup=update_request.ondup
            )
        
        file_metadata = {
            'name': os.path.basename(source_path),
            'mime_type': FileMimeType.get_mimetype(update_request.file_type),
        }

        return self.efast.file_upload(
            docid=parent_doc_id,
            data_length=data_length,
            doc_name=update_request.doc_name_with_ext,
            file_input=source_path,
            file_metadata=file_metadata,
            ondup=update_request.ondup or 3
        )