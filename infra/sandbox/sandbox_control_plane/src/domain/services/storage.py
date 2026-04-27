"""
存储领域服务接口

定义存储的抽象接口，负责文件存储操作。
"""
from abc import ABC, abstractmethod
from typing import Dict, Optional


class IStorageService(ABC):
    """
    存储服务接口

    这是领域层定义的 Port，由基础设施层实现 Adapter。
    负责与 S3 兼容的对象存储进行交互。
    """

    @abstractmethod
    async def upload_file(
        self,
        s3_path: str,
        content: bytes,
        content_type: str = "application/octet-stream"
    ) -> None:
        """
        上传文件

        Args:
            s3_path: S3 对象路径
            content: 文件内容
            content_type: MIME 类型
        """
        pass

    @abstractmethod
    async def download_file(self, s3_path: str) -> bytes:
        """
        下载文件

        Args:
            s3_path: S3 对象路径

        Returns:
            文件内容
        """
        pass

    @abstractmethod
    async def file_exists(self, s3_path: str) -> bool:
        """
        检查文件是否存在

        Args:
            s3_path: S3 对象路径

        Returns:
            是否存在
        """
        pass

    @abstractmethod
    async def get_file_info(self, s3_path: str) -> Dict:
        """
        获取文件信息

        Args:
            s3_path: S3 对象路径

        Returns:
            文件信息字典，包含 size, content_type, last_modified 等
        """
        pass

    @abstractmethod
    async def generate_presigned_url(
        self,
        s3_path: str,
        expiration_seconds: int = 3600
    ) -> str:
        """
        生成预签名 URL

        Args:
            s3_path: S3 对象路径
            expiration_seconds: 过期时间（秒）

        Returns:
            预签名 URL
        """
        pass

    @abstractmethod
    async def delete_file(self, s3_path: str) -> None:
        """
        删除文件

        Args:
            s3_path: S3 对象路径
        """
        pass

    @abstractmethod
    async def list_files(
        self,
        prefix: str,
        limit: int = 1000
    ) -> list:
        """
        列出文件

        Args:
            prefix: S3 路径前缀
            limit: 最大返回数量

        Returns:
            文件列表
        """
        pass

    @abstractmethod
    async def delete_prefix(
        self,
        prefix: str
    ) -> int:
        """
        删除指定前缀的所有文件（用于会话清理）

        Args:
            prefix: S3 路径前缀（例如: "sessions/sess_abc123/"）

        Returns:
            删除的文件数量
        """
        pass
