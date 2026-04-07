# 简体中文
from app.commons.errors.codes import ParamValidationErrors

error_messages = {
    ParamValidationErrors.ParamMissing: {
        "code": ParamValidationErrors.ParamMissing,
        "description": "参数缺失",
        "detail": "",
        "solution": "Please read the API documentation and pass the correct parameters",
        "link": ""
    },
    ParamValidationErrors.ParamTypeError: {
        "code": ParamValidationErrors.ParamTypeError,
        "description": "参数类型错误",
        "solution": "Please read the API documentation and pass the correct parameters",
        "detail": "",
        "link": ""

    }
}
