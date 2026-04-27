import json
from common.logger import logger
from common.constants import page_not_found
from errors.errors import ErrorInfo, ErrorBuilder, INTERNAL_ERROR, BAD_REQUEST_ERROR, INTERNAL_ERROR

class ErrorHandler:
    def __init__(self, handler):
        self.handler = handler

    def _write_error(self, http_status: int, error_info: ErrorInfo, e: Exception):
        logger.warning(e)
        self.handler.set_status(http_status)

        if isinstance(e, ErrorBuilder):
            response_content = e.serialize()
        else:
            if error_info is None:
                error_info = INTERNAL_ERROR
            detail = {"exception": e.__class__.__name__}
            # 尝试解析异常的 body 属性
            try:
                body = getattr(e, 'body', None)
                if body is not None:
                    parsed = json.loads(str(body))
                    if isinstance(parsed, dict):
                        detail.update(parsed)
                    else:
                        detail["cause"] = str(e)
                else:
                    detail["cause"] = str(e)
            except Exception:
                detail["cause"] = str(e)

            response_content = {
                "code": error_info.code,
                "message": error_info.message,
                "cause": error_info.cause,
                "detail": detail
            }

        response_content_str = json.dumps(response_content, ensure_ascii=False)
        self.handler.set_header("Content-Type", "application/json")
        self.handler.write(response_content_str)

    def BAD_REQUEST(self, e: Exception):
        self._write_error(400, BAD_REQUEST_ERROR, e)

    def INTERNAL_ERROR(self, e: Exception):
        self._write_error(500, INTERNAL_ERROR, e)

    def CUSTOM_ERROR(self, code: int, error_info: ErrorInfo, e: Exception):
        self._write_error(code, error_info, e)

    def PAGE_NOT_FOUND(self):
        self.handler.set_status(404)
        self.handler.write(page_not_found)