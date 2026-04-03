from typing import Any, Dict
from fastapi import status
from fastapi.responses import JSONResponse

from app.commons.i18n import get_error_message


def error_response(
        status_code: int,
        code: str,
        detail: str,
        solution: str = "",
        link: str = "",
        language: str = "zh"
) -> JSONResponse:
    """Handle errors response
    Args:
        status_code: HTTP status code
        code: Error code
        detail: Error detail message
        solution: solution suggestion
        link: Reference link
        lang: Language code (en/zh)
    Returns:
        JSONResponse: FastAPI response object with errors structure
    """
    error_content: Dict[str, str] = {
        "code": code,
        "description": get_error_message(code, language),
        "solution": solution,
        "detail": detail,
        "link": link
    }

    return JSONResponse(
        status_code=status_code,
        content=error_content
    )


def correct_response(http_code: int = status.HTTP_200_OK, data: Any = None, ) -> JSONResponse:
    """Handle normal response
    Args:
        data: Response data
        http_code: HTTP status code
    Returns:
        JSONResponse: FastAPI response object
    """
    return JSONResponse(
        status_code=http_code,
        content=data
    )
