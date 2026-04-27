from app.routers import llm_router, open_api_json, small_model_router, private_route, prompt_router, model_quota_router
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from fastapi import Request

api_version_public_v1 = "/api/mf-model-manager/v1"
api_version_private_v1 = "/api/private/mf-model-manager/v1"
api_version_health = "/api/v1"


def router_init(app):
    app.include_router(
        llm_router.health_route,
        prefix=api_version_health,
        tags=["Factory"],
        responses={404: {"description": "Not found"}},
    )
    app.include_router(
        llm_router.llm_route,
        prefix=api_version_public_v1,
        tags=["Factory"],
        responses={404: {"description": "Not found"}},
    )
    app.include_router(
        open_api_json.open_json_router,
        prefix=api_version_public_v1,
        tags=["Factory"],
        responses={404: {"description": "Not found"}},
    )

    app.include_router(
        small_model_router.small_model_router,
        prefix=api_version_public_v1,
        tags=["Factory"],
        responses={404: {"description": "Not found"}},
    )
    app.include_router(
        private_route.private_route,
        prefix=api_version_private_v1,
        tags=["Factory"],
        responses={404: {"description": "Not found"}},
    )
    app.include_router(
        prompt_router.prompt_route,
        prefix=api_version_public_v1,
        tags=["Factory"],
        responses={404: {"description": "Not found"}},
    )
    app.include_router(
        model_quota_router.model_quota_router,
        prefix=api_version_public_v1,
        tags=["Factory"],
        responses={404: {"description": "Not found"}},
    )
    @app.exception_handler(RequestValidationError)
    async def exception_handler(request: Request, exc: RequestValidationError):
        print("errors:")
        print(exc.errors())
        for error in exc.errors():
            paramName = ' '.join(map(str, error["loc"][1:]))
            if error["type"] == "value_error.missing":
                content = {"code": "ModelFactory.Router.ParamError.ParamMissing",
                           "description": "参数缺失",
                           "detail": "{0} 参数缺失".format(paramName),
                           "solution": "请检查填写的参数是否正确。",
                           "link": ""}
                return JSONResponse(status_code=400, content=content)
            else:
                content = {"code": "ModelFactory.Router.ParamError.FormatError",
                           "description": "参数错误",
                           "detail": f"{error.get('msg', '')}",
                           "solution": "请检查输入内容格式是否符合要求",
                           "link": ""}
                return JSONResponse(status_code=400, content=content)

    app.add_exception_handler(RequestValidationError, exception_handler)
