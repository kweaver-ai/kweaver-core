import os
import shutil
from typing import Callable, Optional, Tuple

import requests
from errors.errors import InternalErrException
from common.logger import logger
from common.configs import CacheConfig
from errors.errors import BadParameterException
from utils.utils import is_valid_url

class FileCacheManager:
    _instance = None
    
    def __new__(cls, *args, **kwargs):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._initialized = False
        return cls._instance
    
    def __init__(self):
        """
        文件缓存管理器的构造函数

        Args:
            host (str): EFAST的host
            access_token (str): EFAST的access_token
        """
        if self._initialized:
            return
        
        self._initialized = True
        self._cache_dir = CacheConfig.CACHE_ROOT_DIR

        self._init_cache_dir()

    def _init_cache_dir(self):
        """初始化缓存目录"""
        if not os.path.exists(self._cache_dir):
            os.makedirs(self._cache_dir, exist_ok=True)

    def download_from_docid(self, doc_id: str, download_fn: Optional[Callable[[str], Tuple[bytes, str]]] = None):
        """
        获取文件（缓存或下载）

        Args:
            doc_id: 文件 ID
            download_fn: 下载函数，接受 doc_id 参数，返回 (文件内容字节, 文件名)
        Returns:
            Tuple[str, str]: 文件路径和文件名
        """
        if not doc_id:
            return

        ids = doc_id.split('/')
        if len(ids) == 0:
            return

        cache_dir_name = ids[-1]
        cache_dir_path = os.path.join(self._cache_dir, cache_dir_name)

        # 确保缓存目录存在
        os.makedirs(cache_dir_path, exist_ok=True)

        if download_fn is None:
            raise InternalErrException(detail= f"download_from_docid failed, download_fn is None")
        content_bytes, name = download_fn(doc_id)

        file_path = os.path.join(cache_dir_path, name)
        with open(file_path, 'wb') as f:
            f.write(content_bytes)

        return file_path, name

    def download_from_url(self, file_path: str, url: str):
        # 判断url是否是一个合法的url
        if not is_valid_url(url):
            raise BadParameterException(detail={
                "info": "url is not valid",
                "url": url[0:30] if len(url) > 30 else url
            })

        with requests.get(url, stream=True) as resp:
            resp.raise_for_status()
            
            # 分块写入文件
            with open(file_path, 'wb') as file:
                for chunk in resp.iter_content(chunk_size=8192):
                    file.write(chunk)

    
    def clear_cache(self, doc_id: str):
        if not doc_id:
            return
        
        ids = doc_id.split('/')
        if len(ids) == 0:
            return

        cache_dir_name = ids[-1]
        cache_dir_path = os.path.join(self._cache_dir, cache_dir_name)
        try:
            shutil.rmtree(cache_dir_path)
        except Exception as e:
            logger.warning(f"clear cache failed, path: {cache_dir_path}, detail: {str(e)}")
        