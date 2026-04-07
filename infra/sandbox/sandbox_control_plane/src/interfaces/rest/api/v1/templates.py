"""
模板 REST API 路由

定义模板相关的 HTTP 端点。
"""
from fastapi import APIRouter, Depends, HTTPException, status
from typing import List

from src.application.services.template_service import TemplateService
from src.application.commands.create_template import CreateTemplateCommand
from src.application.commands.update_template import UpdateTemplateCommand
from src.application.queries.get_template import GetTemplateQuery
from src.application.dtos.template_dto import TemplateDTO
from src.interfaces.rest.schemas.request import CreateTemplateRequest, UpdateTemplateRequest
from src.interfaces.rest.schemas.response import TemplateResponse, ErrorResponse
from src.infrastructure.dependencies import get_template_service_db

router = APIRouter(prefix="/templates", tags=["templates"])


@router.post("", response_model=TemplateResponse, status_code=status.HTTP_201_CREATED)
async def create_template(
    request: CreateTemplateRequest,
    service: TemplateService = Depends(get_template_service_db)
):
    """
    创建模板

    - **id**: 模板 ID
    - **name**: 模板名称
    - **image_url**: 镜像 URL
    - **runtime_type**: 运行时类型 (python3.11, nodejs20, java17, go1.21)
    - **default_cpu_cores**: 默认 CPU 核心数
    - **default_memory_mb**: 默认内存（MB）
    - **default_disk_mb**: 默认磁盘（MB）
    - **default_timeout_sec**: 默认超时时间（秒）
    - **default_env_vars**: 默认环境变量
    """
    command = CreateTemplateCommand(
        template_id=request.id,
        name=request.name,
        image_url=request.image_url,
        runtime_type=request.runtime_type,
        default_cpu_cores=request.default_cpu_cores,
        default_memory_mb=request.default_memory_mb,
        default_disk_mb=request.default_disk_mb,
        default_timeout_sec=request.default_timeout,
        default_env_vars=request.default_env_vars
    )

    template_dto = await service.create_template(command)
    return _map_dto_to_response(template_dto)


@router.get("", response_model=List[TemplateResponse])
async def list_templates(
    limit: int = 50,
    offset: int = 0,
    service: TemplateService = Depends(get_template_service_db)
):
    """列出所有模板"""
    templates = await service.list_templates(limit=limit, offset=offset)
    return [_map_dto_to_response(t) for t in templates]


@router.get("/{template_id}", response_model=TemplateResponse)
async def get_template(
    template_id: str,
    service: TemplateService = Depends(get_template_service_db)
):
    """获取模板详情"""
    query = GetTemplateQuery(template_id=template_id)
    template_dto = await service.get_template(query)
    return _map_dto_to_response(template_dto)


@router.put("/{template_id}", response_model=TemplateResponse)
async def update_template(
    template_id: str,
    request: UpdateTemplateRequest,
    service: TemplateService = Depends(get_template_service_db)
):
    """更新模板"""
    command = UpdateTemplateCommand(
        template_id=template_id,
        name=request.name,
        image_url=request.image_url,
        default_cpu_cores=request.default_cpu_cores,
        default_memory_mb=request.default_memory_mb,
        default_disk_mb=request.default_disk_mb,
        default_timeout_sec=request.default_timeout,
        default_env_vars=request.default_env_vars
    )

    template_dto = await service.update_template(command)
    return _map_dto_to_response(template_dto)


@router.delete("/{template_id}")
async def delete_template(
    template_id: str,
    service: TemplateService = Depends(get_template_service_db)
):
    """删除模板"""
    await service.delete_template(template_id)
    return {"message": "Template deleted successfully"}


def _map_dto_to_response(dto: TemplateDTO) -> TemplateResponse:
    """将 TemplateDTO 映射为 TemplateResponse"""
    return TemplateResponse(
        id=dto.id,
        name=dto.name,
        image_url=dto.image_url,
        runtime_type=dto.runtime_type,
        default_cpu_cores=dto.default_cpu_cores,
        default_memory_mb=dto.default_memory_mb,
        default_disk_mb=dto.default_disk_mb,
        default_timeout_sec=dto.default_timeout_sec,
        default_env_vars=dto.default_env_vars,
        is_active=dto.is_active,
        created_at=dto.created_at,
        updated_at=dto.updated_at
    )
