"""
文件应用服务

编排文件上传下载相关的用例。
"""
from typing import Dict, List, Any
from urllib.parse import urlparse

from src.domain.repositories.session_repository import ISessionRepository
from src.domain.services.storage import IStorageService
from src.shared.errors.domain import NotFoundError, ValidationError


class FileService:
    """
    文件应用服务

    编排文件上传、下载等用例。
    """

    def __init__(
        self,
        session_repo: ISessionRepository,
        storage_service: IStorageService,
    ):
        self._session_repo = session_repo
        self._storage_service = storage_service

    async def upload_file(
        self,
        session_id: str,
        path: str,
        content: bytes,
        content_type: str = "application/octet-stream"
    ) -> str:
        """
        上传文件用例

        流程：
        1. 验证会话存在且运行中
        2. 验证路径格式
        3. 上传到存储
        4. 返回文件路径
        """
        session = await self._session_repo.find_by_id(session_id)
        if not session:
            raise NotFoundError(f"Session not found: {session_id}")

        if not session.is_active():
            raise ValidationError(f"Session is not active: {session_id}")

        if not path or path.startswith("/"):
            raise ValidationError("Invalid file path")

        s3_path = f"{session.workspace_path}/{path}"
        await self._storage_service.upload_file(
            s3_path=s3_path,
            content=content,
            content_type=content_type
        )

        return path

    async def download_file(self, session_id: str, path: str) -> Dict:
        """
        下载文件用例

        流程：
        1. 验证会话存在
        2. 验证文件存在
        3. 返回文件内容或预签名 URL
        """
        session = await self._session_repo.find_by_id(session_id)
        if not session:
            raise NotFoundError(f"Session not found: {session_id}")

        s3_path = f"{session.workspace_path}/{path}"
        file_exists = await self._storage_service.file_exists(s3_path)
        if not file_exists:
            raise NotFoundError(f"File not found: {path}")

        file_info = await self._storage_service.get_file_info(s3_path)
        file_size = file_info["size"]

        # 小文件（<10MB）直接返回内容，大文件返回预签名 URL
        SMALL_FILE_THRESHOLD = 10 * 1024 * 1024  # 10MB

        if file_size < SMALL_FILE_THRESHOLD:
            content = await self._storage_service.download_file(s3_path)
            return {
                "content": content,
                "content_type": file_info.get("content_type", "application/octet-stream"),
                "size": file_size,
            }

        presigned_url = await self._storage_service.generate_presigned_url(s3_path)
        return {
            "presigned_url": presigned_url,
            "size": file_size,
        }

    async def list_files(
        self,
        session_id: str,
        path: str = None,
        limit: int = 1000
    ) -> List[Dict[str, Any]]:
        """
        列出 session 下的文件

        Args:
            session_id: Session ID
            path: 可选，指定目录路径（相对于 workspace 根目录）
            limit: 最大返回文件数

        Returns:
            文件列表，每个文件包含 name, size, modified_time, container_path 等
        """
        session = await self._session_repo.find_by_id(session_id)
        if not session:
            raise NotFoundError(f"Session not found: {session_id}")

        # 解析 workspace_path，提取 S3 key 前缀
        # workspace_path 格式: s3://bucket/sessions/{session_id}/
        # S3 key 格式: sessions/{session_id}/...
        parsed = urlparse(session.workspace_path)
        s3_key_prefix = parsed.path.lstrip("/")  # 去掉开头的 /，得到 "sessions/{session_id}/"

        # 构建 S3 查询前缀
        if path:
            normalized_path = path.strip().strip("/")
            if normalized_path:
                # 确保 s3_key_prefix 以 / 结尾
                base = s3_key_prefix.rstrip("/")
                prefix = f"{base}/{normalized_path}"
            else:
                prefix = s3_key_prefix.rstrip("/")
        else:
            prefix = s3_key_prefix.rstrip("/")

        files = await self._storage_service.list_files(prefix, limit)

        result = []

        for file in files:
            key = file["key"]

            # 提取相对于 session workspace 的路径
            # key 格式: sessions/{session_id}/conversation-1231/uploads/temparea/test.csv
            # s3_key_prefix 格式: sessions/{session_id}/
            if key.startswith(s3_key_prefix):
                relative_path = key[len(s3_key_prefix):].lstrip("/")
            else:
                relative_path = key.lstrip("/")

            # 过滤掉空路径（目录本身）和以斜杠结尾的目录标记
            if not relative_path or relative_path.endswith("/"):
                continue

            # 容器内挂载路径: /workspace/{relative_path}
            container_path = f"/workspace/{relative_path}"

            result.append({
                "name": relative_path,
                "container_path": container_path,
                "size": file["size"],
                "modified_time": file.get("last_modified"),
                "etag": file.get("etag")
            })

        return result
