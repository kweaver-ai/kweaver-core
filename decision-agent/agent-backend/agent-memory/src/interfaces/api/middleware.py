from fastapi import Request
from fastapi.responses import JSONResponse
from src.utils.i18n import i18n_manager
from src.interfaces.api.exceptions import MemoryException
from src.interfaces.api.schemas import ErrorResponse
from fastapi.exceptions import RequestValidationError
from pydantic_core import ValidationError as PydanticValidationError
from src.utils.logger import logger


async def error_handler_middleware(request: Request, call_next):
    # 从请求头获取语言设置，默认为中文
    language = request.headers.get("X-Language", "zh_CN")

    try:
        return await call_next(request)

    except (RequestValidationError, PydanticValidationError) as e:
        logger.errorf(
            "request validate error::%s",
            str(e),
            exc_info=True,
            extra={"details": e.errors()},
        )
        # 处理请求参数验证错误
        error_info = i18n_manager.get_error_info(
            error_code="AgentMemory.Validation.Error",
            lang=language,
            custom_description=str(e),
        )

        error_response = ErrorResponse(
            error_code="AgentMemory.Validation.Error",
            description=error_info["description"],
            solution=error_info["solution"],
            error_link=error_info["error_link"],
        )
        return JSONResponse(status_code=400, content=error_response.model_dump())

    except MemoryException as e:
        logger.errorf(
            "Memory Exception error: :%s",
            str(e),
            exc_info=True,
            extra={"details": str(e)},
        )
        # 获取错误信息
        error_info = i18n_manager.get_error_info(e.error_code, language)

        # 构建错误响应
        error_response = ErrorResponse(
            error_code=e.error_code,
            description=error_info["description"],
            solution=error_info["solution"],
            error_link=error_info["error_link"],
            error_details={"error": str(e)},
        )

        return JSONResponse(
            status_code=e.status_code, content=error_response.model_dump()
        )
    except Exception as e:
        logger.errorf(
            "Unhandled error:%s,", str(e), exc_info=True, extra={"details": str(e)}
        )
        # 处理未预期的错误
        error_info = i18n_manager.get_error_info("AgentMemory.Internal.Error", language)

        error_response = ErrorResponse(
            error_code="AgentMemory.Internal.Error",
            description=error_info["description"],
            solution=error_info["solution"],
            error_link=error_info["error_link"],
            error_details={"error": str(e)},
        )

        return JSONResponse(status_code=500, content=error_response.model_dump())
