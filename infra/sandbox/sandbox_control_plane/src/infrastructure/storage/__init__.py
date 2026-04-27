"""
存储模块

提供 S3 兼容的对象存储实现（AWS S3、MinIO）
"""
from .s3_storage import S3Storage

__all__ = ["S3Storage"]
