"""
S3 存储实现

使用 boto3 实现 S3 兼容的对象存储，支持 AWS S3 和 MinIO。
"""
import asyncio
import logging
import os
from typing import Optional
from urllib.parse import urlparse

import boto3
from botocore.exceptions import ClientError

from src.domain.services.storage import IStorageService
from src.infrastructure.config.settings import get_settings

logger = logging.getLogger(__name__)


class S3Storage(IStorageService):
    """
    S3 兼容的存储实现

    支持：
    - AWS S3
    - MinIO（用于本地开发）
    - 任何 S3 兼容的存储（例如：Wasabi、DigitalOcean Spaces）
    """

    def __init__(self):
        """
        初始化 S3 客户端

        从 settings 中读取 S3 配置：
        - s3_endpoint_url: S3 端点 URL（MinIO 使用）
        - s3_access_key_id: 访问密钥 ID
        - s3_secret_access_key: 密钥
        - s3_region: 区域
        - s3_bucket: 存储桶名称
        """
        settings = get_settings()

        # 初始化 S3 客户端
        self._client = boto3.client(
            's3',
            endpoint_url=settings.s3_endpoint_url or None,  # AWS S3 不需要 endpoint_url
            aws_access_key_id=settings.s3_access_key_id,
            aws_secret_access_key=settings.s3_secret_access_key,
            region_name=settings.s3_region,
        )
        self._bucket = settings.s3_bucket

    async def initialize(self) -> None:
        """
        异步初始化，确保 bucket 存在

        在 control-plane 启动时调用此方法来确保 S3 bucket 已创建。
        """
        try:
            await self._ensure_bucket_exists()
            logger.info(f"S3 storage initialized successfully (bucket: {self._bucket})")
        except Exception as e:
            logger.error(f"Failed to initialize S3 storage: {e}")
            # 不抛出异常，允许系统在 MinIO 不可用时继续运行
            # 文件操作会失败，但不会阻止控制平面启动

    def _parse_s3_path(self, s3_path: str) -> tuple[str, str]:
        """
        解析 S3 路径，返回 bucket 和 key

        支持两种格式：
        1. s3://bucket/key
        2. bucket/key (相对路径)

        Args:
            s3_path: S3 对象路径

        Returns:
            (bucket, key) 元组
        """
        if s3_path.startswith("s3://"):
            parsed = urlparse(s3_path)
            bucket = parsed.netloc
            key = parsed.path.lstrip('/')
        else:
            # 相对路径，使用默认 bucket
            bucket = self._bucket
            key = s3_path.lstrip('/')

        return bucket, key

    def _build_s3_path(self, bucket: str, key: str) -> str:
        """
        构建 S3 路径

        Args:
            bucket: 存储桶名称
            key: 对象键

        Returns:
            S3 路径（s3://bucket/key 格式）
        """
        return f"s3://{bucket}/{key}"

    async def _ensure_bucket_exists(self) -> None:
        """确保存储桶存在，不存在则创建"""
        try:
            await asyncio.to_thread(
                self._client.head_bucket,
                Bucket=self._bucket
            )
        except ClientError as e:
            error_code = e.response.get('Error', {}).get('Code')
            if error_code == '404':
                # 存储桶不存在，创建它
                try:
                    if self._client.meta.region_name == 'us-east-1':
                        # us-east-1 不需要 LocationConstraint
                        await asyncio.to_thread(
                            self._client.create_bucket,
                            Bucket=self._bucket
                        )
                    else:
                        await asyncio.to_thread(
                            self._client.create_bucket,
                            Bucket=self._bucket,
                            CreateBucketConfiguration={
                                'LocationConstraint': self._client.meta.region_name
                            }
                        )
                    logger.info(f"Created S3 bucket: {self._bucket}")
                except ClientError as create_error:
                    logger.error(f"Failed to create bucket {self._bucket}: {create_error}")
                    raise
            else:
                logger.error(f"Error checking bucket {self._bucket}: {e}")
                raise

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
        await self._ensure_bucket_exists()

        bucket, key = self._parse_s3_path(s3_path)

        # 根据内容大小选择上传方式
        content_size = len(content)

        if content_size > 5 * 1024 * 1024:  # 大于 5MB 使用分片上传
            from boto3.s3.transfer import TransferConfig

            # 先写入临时文件
            import tempfile
            with tempfile.NamedTemporaryFile(delete=False) as tmp_file:
                tmp_file.write(content)
                tmp_file_path = tmp_file.name

            try:
                config = TransferConfig(
                    multipart_threshold=5 * 1024 * 1024,
                    max_concurrency=4
                )
                await asyncio.to_thread(
                    self._client.upload_file,
                    Filename=tmp_file_path,
                    Bucket=bucket,
                    Key=key,
                    ExtraArgs={'ContentType': content_type},
                    Config=config
                )
            finally:
                os.unlink(tmp_file_path)
        else:
            # 小文件直接上传
            await asyncio.to_thread(
                self._client.put_object,
                Bucket=bucket,
                Key=key,
                Body=content,
                ContentType=content_type
            )

        # 清理可能存在的目录标记 (s3fs 兼容性修复)
        # 当上传 test/test_data.csv 时，S3 可能会创建 test/ 目录标记
        # 这会导致 s3fs 将 test 显示为文件而非目录
        if '/' in key:
            dir_marker = key.rsplit('/', 1)[0] + '/'
            try:
                await asyncio.to_thread(
                    self._client.head_object,
                    Bucket=bucket,
                    Key=dir_marker
                )
                # 目录标记存在，删除它
                await asyncio.to_thread(
                    self._client.delete_object,
                    Bucket=bucket,
                    Key=dir_marker
                )
                logger.debug(f"Removed S3 directory marker for s3fs compatibility: {dir_marker}")
            except ClientError as e:
                error_code = e.response.get('Error', {}).get('Code')
                if error_code == '404':
                    # 目录标记不存在，无需处理
                    pass

        logger.debug(f"Uploaded file to {s3_path}, size={content_size}")

    async def download_file(self, s3_path: str) -> bytes:
        """
        下载文件

        Args:
            s3_path: S3 对象路径

        Returns:
            文件内容
        """
        bucket, key = self._parse_s3_path(s3_path)

        response = await asyncio.to_thread(
            self._client.get_object,
            Bucket=bucket,
            Key=key
        )

        content = response['Body'].read()
        logger.debug(f"Downloaded file from {s3_path}, size={len(content)}")

        return content

    async def file_exists(self, s3_path: str) -> bool:
        """
        检查文件是否存在

        Args:
            s3_path: S3 对象路径

        Returns:
            是否存在
        """
        bucket, key = self._parse_s3_path(s3_path)

        try:
            await asyncio.to_thread(
                self._client.head_object,
                Bucket=bucket,
                Key=key
            )
            return True
        except ClientError as e:
            error_code = e.response.get('Error', {}).get('Code')
            if error_code == '404':
                return False
            logger.error(f"Error checking file existence {s3_path}: {e}")
            raise

    async def get_file_info(self, s3_path: str) -> dict:
        """
        获取文件信息

        Args:
            s3_path: S3 对象路径

        Returns:
            文件信息字典，包含 size, content_type, last_modified 等
        """
        bucket, key = self._parse_s3_path(s3_path)

        response = await asyncio.to_thread(
            self._client.head_object,
            Bucket=bucket,
            Key=key
        )

        return {
            "size": response['ContentLength'],
            "content_type": response.get('ContentType', 'application/octet-stream'),
            "last_modified": response['LastModified'],
            "etag": response['ETag'].strip('"')
        }

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
        bucket, key = self._parse_s3_path(s3_path)

        url = await asyncio.to_thread(
            self._client.generate_presigned_url,
            'get_object',
            Params={'Bucket': bucket, 'Key': key},
            ExpiresIn=expiration_seconds
        )

        logger.debug(f"Generated presigned URL for {s3_path}, expires in {expiration_seconds}s")

        return url

    async def delete_file(self, s3_path: str) -> None:
        """
        删除文件

        Args:
            s3_path: S3 对象路径
        """
        bucket, key = self._parse_s3_path(s3_path)

        await asyncio.to_thread(
            self._client.delete_object,
            Bucket=bucket,
            Key=key
        )

        logger.debug(f"Deleted file {s3_path}")

    async def delete_prefix(self, prefix: str) -> int:
        """
        删除指定前缀的所有文件（用于会话清理）

        Args:
            prefix: S3 路径前缀（例如: "sessions/sess_abc123/" 或 "s3://bucket/sessions/sess_abc123/"）

        Returns:
            删除的文件数量
        """
        deleted_count = 0
        bucket = self._bucket

        # 如果 prefix 包含 bucket，提取出来
        if prefix.startswith("s3://"):
            parsed = urlparse(prefix)
            bucket = parsed.netloc
            prefix = parsed.path.lstrip('/')

        # 使用 asyncio.to_thread 执行同步的列表和删除操作
        def _delete_all_files():
            """同步函数，执行批量删除"""
            count = 0
            paginator = self._client.get_paginator('list_objects_v2')
            delete_chunks = []

            try:
                # 直接迭代 paginator（同步操作）
                for page in paginator.paginate(Bucket=bucket, Prefix=prefix):
                    if 'Contents' in page:
                        for obj in page['Contents']:
                            delete_chunks.append({'Key': obj['Key']})

                        # 当累积到 1000 个对象时删除
                        if len(delete_chunks) >= 1000:
                            self._client.delete_objects(
                                Bucket=bucket,
                                Delete={'Objects': delete_chunks}
                            )
                            count += len(delete_chunks)
                            delete_chunks = []

                # 删除剩余的对象
                if delete_chunks:
                    self._client.delete_objects(
                        Bucket=bucket,
                        Delete={'Objects': delete_chunks}
                    )
                    count += len(delete_chunks)

            except ClientError as e:
                logger.error(f"Error deleting files with prefix {prefix}: {e}")
            return count

        # 在线程池中执行同步操作
        deleted_count = await asyncio.to_thread(_delete_all_files)

        logger.info(f"Deleted {deleted_count} files with prefix {prefix} (bucket: {bucket})")

        return deleted_count

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
            文件列表，每个文件包含 key, size, last_modified
        """
        bucket = self._bucket

        # 如果 prefix 包含 bucket，提取出来
        if prefix.startswith("s3://"):
            parsed = urlparse(prefix)
            bucket = parsed.netloc
            prefix = parsed.path.lstrip('/')

        def _list_all_files():
            """同步函数，执行列表操作"""
            files = []
            paginator = self._client.get_paginator('list_objects_v2')

            try:
                # 直接迭代 paginator（同步操作）
                for page in paginator.paginate(Bucket=bucket, Prefix=prefix):
                    if 'Contents' in page:
                        for obj in page['Contents']:
                            files.append({
                                'key': obj['Key'],
                                'size': obj['Size'],
                                'last_modified': obj['LastModified'],
                                'etag': obj['ETag'].strip('"')
                            })
                            if limit and len(files) >= limit:
                                break
                    if limit and len(files) >= limit:
                        break
            except ClientError as e:
                logger.error(f"Error listing objects with prefix {prefix}: {e}")
            return files

        # 在线程池中执行同步操作
        files = await asyncio.to_thread(_list_all_files)

        return files
