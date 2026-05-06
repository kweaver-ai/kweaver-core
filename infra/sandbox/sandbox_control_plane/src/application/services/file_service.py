"""
文件应用服务

编排文件上传下载相关的用例。
"""

import io
import mimetypes
import re
import stat
import zipfile
from pathlib import PurePosixPath
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
        max_extracted_file_count: int = 10000,
        max_extracted_total_size_mb: int = 512,
    ):
        self._session_repo = session_repo
        self._storage_service = storage_service
        self._max_extracted_file_count = max_extracted_file_count
        self._max_extracted_total_size_bytes = max_extracted_total_size_mb * 1024 * 1024

    async def upload_file(
        self,
        session_id: str,
        path: str,
        content: bytes,
        content_type: str = "application/octet-stream",
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

        normalized_path = self._validate_relative_path(path)

        s3_path = f"{session.workspace_path}/{normalized_path}"
        await self._storage_service.upload_file(
            s3_path=s3_path, content=content, content_type=content_type
        )

        return normalized_path

    async def upload_and_extract_zip(
        self,
        session_id: str,
        path: str,
        content: bytes,
        overwrite: bool = False,
    ) -> Dict[str, Any]:
        """
        上传 ZIP 并解压到会话工作区。

        返回解压统计结果。
        """
        session = await self._session_repo.find_by_id(session_id)
        if not session:
            raise NotFoundError(f"Session not found: {session_id}")

        if not session.is_active():
            raise ValidationError(f"Session is not active: {session_id}")

        extract_path = self._validate_directory_path(path)

        try:
            archive = zipfile.ZipFile(io.BytesIO(content))
        except zipfile.BadZipFile as exc:
            raise ValidationError("Invalid ZIP archive") from exc

        with archive:
            entries: list[tuple[zipfile.ZipInfo, str, str]] = []
            total_uncompressed_size = 0

            for zip_info in archive.infolist():
                if zip_info.is_dir():
                    continue

                entry_path = self._validate_zip_entry_path(zip_info)
                total_uncompressed_size += zip_info.file_size

                if len(entries) + 1 > self._max_extracted_file_count:
                    raise ValidationError("ZIP archive contains too many files")
                if total_uncompressed_size > self._max_extracted_total_size_bytes:
                    raise ValidationError("ZIP archive uncompressed size exceeds limit")

                entries.append(
                    (
                        zip_info,
                        entry_path,
                        self._join_paths(extract_path, entry_path),
                    )
                )

            extracted_file_count = 0
            skipped_file_count = 0

            for zip_info, entry_path, destination_path in entries:
                s3_path = f"{session.workspace_path}/{destination_path}"

                if not overwrite and await self._storage_service.file_exists(s3_path):
                    skipped_file_count += 1
                    continue

                with archive.open(zip_info, "r") as member:
                    file_content = member.read()

                content_type = mimetypes.guess_type(entry_path)[0] or "application/octet-stream"
                await self._storage_service.upload_file(
                    s3_path=s3_path,
                    content=file_content,
                    content_type=content_type,
                )
                extracted_file_count += 1

        return {
            "mode": "archive_extract",
            "extract_path": extract_path,
            "extracted_file_count": extracted_file_count,
            "skipped_file_count": skipped_file_count,
            "size": len(content),
        }

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
        self, session_id: str, path: str = None, limit: int = 1000
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
                relative_path = key[len(s3_key_prefix) :].lstrip("/")
            else:
                relative_path = key.lstrip("/")

            # 过滤掉空路径（目录本身）和以斜杠结尾的目录标记
            if not relative_path or relative_path.endswith("/"):
                continue

            # 容器内挂载路径: /workspace/{relative_path}
            container_path = f"/workspace/{relative_path}"

            result.append(
                {
                    "name": relative_path,
                    "container_path": container_path,
                    "size": file["size"],
                    "modified_time": file.get("last_modified"),
                    "etag": file.get("etag"),
                }
            )

        return result

    def _validate_relative_path(self, path: str) -> str:
        if not path:
            raise ValidationError("Invalid file path")

        stripped = path.strip()
        if not stripped or stripped.startswith("/") or "\\" in stripped:
            raise ValidationError("Invalid file path")
        if re.match(r"^[A-Za-z]:", stripped):
            raise ValidationError("Invalid file path")

        normalized = PurePosixPath(stripped).as_posix()
        parts = PurePosixPath(normalized).parts
        if any(part == ".." for part in parts):
            raise ValidationError("Invalid file path")

        if normalized.startswith("./"):
            normalized = normalized[2:]
        if not normalized or normalized == ".":
            raise ValidationError("Invalid file path")
        return normalized

    def _validate_directory_path(self, path: str) -> str:
        normalized = self._validate_relative_path(path)
        return normalized.rstrip("/")

    def _validate_zip_entry_path(self, zip_info: zipfile.ZipInfo) -> str:
        path = zip_info.filename
        if not path:
            raise ValidationError("Invalid ZIP entry path")
        if path.startswith("/") or "\\" in path or re.match(r"^[A-Za-z]:", path):
            raise ValidationError("Invalid ZIP entry path")
        mode = zip_info.external_attr >> 16
        if stat.S_ISLNK(mode):
            raise ValidationError("Invalid ZIP entry path")

        normalized = PurePosixPath(path).as_posix()
        parts = PurePosixPath(normalized).parts
        if any(part in ("", ".", "..") for part in parts):
            raise ValidationError("Invalid ZIP entry path")
        return normalized

    def _join_paths(self, base: str, child: str) -> str:
        if not base:
            return child
        return f"{base}/{child}"
