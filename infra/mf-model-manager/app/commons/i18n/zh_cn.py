# 简体中文
from app.commons.errors.codes import ParamValidationErrors

error_messages = {
    ParamValidationErrors.ParamMissing: {
        "code": ParamValidationErrors.ParamMissing,
        "description": "参数缺失",
        "detail": "",
        "solution": "请阅读API文档填写正确的参数",
        "link": ""
    },
    ParamValidationErrors.ParamTypeError: {
        "code": ParamValidationErrors.ParamTypeError,
        "description": "参数类型错误",
        "solution": "请阅读API文档填写正确的参数",
        "detail": "",
        "link": ""

    }
}
