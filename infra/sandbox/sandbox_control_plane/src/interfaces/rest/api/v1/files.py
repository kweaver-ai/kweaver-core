"""
文件操作 REST API 路由

定义文件上传下载相关的 HTTP 端点。
"""
import fastapi
from fastapi import APIRouter, Depends, HTTPException, status, UploadFile, File, Query
from typing import Optional

from src.application.services.file_service import FileService
from src.interfaces.rest.schemas.response import ErrorResponse
from src.infrastructure.dependencies import get_file_service_db

router = APIRouter(prefix="/sessions/{session_id}/files", tags=["files"])


@router.get("")
async def list_files(
    session_id: str,
    path: Optional[str] = Query(None, description="指定目录路径（相对于 workspace 根目录），不指定则列出所有文件"),
    limit: int = Query(1000, ge=1, le=10000, description="最大返回文件数"),
    service: FileService = Depends(get_file_service_db)
):
    """
    列出 session 下的文件

    返回该 session workspace 中的文件列表，支持指定目录路径

    - **path**: 可选，指定目录路径（如 "src/" 或 "src/utils"），不指定则列出所有文件
    - **limit**: 最大返回文件数 (1-10000)
    """
    try:
        files = await service.list_files(
            session_id=session_id,
            path=path,
            limit=limit
        )

        return {
            "session_id": session_id,
            "files": files,
            "count": len(files)
        }

    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )


@router.post("/upload")
async def upload_file(
    session_id: str,
    path: str,
    file: UploadFile = File(...),
    service: FileService = Depends(get_file_service_db)
):
    """
    上传文件到会话工作区

    - **path**: 文件在工作区中的路径
    - **file**: 要上传的文件（最大 100MB）
    """
    try:
        # 验证文件大小
        content = await file.read()
        if len(content) > 100 * 1024 * 1024:  # 100MB
            raise HTTPException(
                status_code=status.HTTP_413_REQUEST_ENTITY_TOO_LARGE,
                detail="File size exceeds 100MB limit"
            )

        file_path = await service.upload_file(
            session_id=session_id,
            path=path,
            content=content,
            content_type=file.content_type
        )

        return {
            "session_id": session_id,
            "file_path": file_path,
            "size": len(content)
        }

    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )


@router.get("/{file_path:path}")
async def download_file(
    session_id: str,
    file_path: str,
    service: FileService = Depends(get_file_service_db)
):
    """
    从会话工作区下载文件

    - **file_path**: 文件在工作区中的路径
    """
    try:
        file_data = await service.download_file(
            session_id=session_id,
            path=file_path
        )

        if file_data.get("presigned_url"):
            return {
                "session_id": session_id,
                "file_path": file_path,
                "presigned_url": file_data["presigned_url"],
                "size": file_data["size"]
            }
        else:
            from fastapi.responses import Response
            return Response(
                content=file_data["content"],
                media_type=file_data["content_type"],
                headers={
                    "Content-Disposition": f'attachment; filename="{file_path}"'
                }
            )

    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e)
        )
