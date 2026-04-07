import json

from fastapi import APIRouter, Header
from sse_starlette import EventSourceResponse

from app.common.config import Config
from app.models.tool_requests import (
    ZhipuSearchRequest,
    OnlineSearchCiteRequest,
)
from app.models.tool_responses import (
    OnlineSearchCiteResponse,
    ZhipuSearchResponse,
)


router = APIRouter(prefix=Config.app.host_prefix + "/tools", tags=["internal-tools"])


@router.post(
    "/zhipu_search_tool", response_model=ZhipuSearchResponse, summary="智谱搜索"
)
async def zhipu_search(
    request: ZhipuSearchRequest,
    api_key: str = Header(..., description="智谱API密钥", alias="api_key"),
) -> ZhipuSearchResponse:
    """
    执行智谱搜索

    - **query**: 搜索查询词

    返回搜索结果内容
    """
    from app.logic.tool.zhipu_search_tool import zhipu_search_tool

    param = request.model_dump()
    res = await zhipu_search_tool(param, {"api_key": api_key}, None, None, None)

    return ZhipuSearchResponse(**res)


@router.post(
    "/online_search_cite_tool",
    response_model=OnlineSearchCiteResponse,
    summary="联网搜索添加引用工具",
)
async def online_search_cite_tool(
    request: OnlineSearchCiteRequest,
) -> OnlineSearchCiteResponse:
    """
    执行联网搜索并添加引用

    - **query**: 搜索查询词
    - **model_name**: 模型名称
    - **search_tool**: 搜索工具
    - **api_key**: 搜索工具API密钥
    - **user_id**: 用户id
    返回搜索结果内容
    """
    headers = {"x-account-id": request.user_id}
    if not request.stream:
        from app.logic.tool.online_search_cite_tool import online_search_cite_tool

        param = request.model_dump()
        res = await online_search_cite_tool(request=param, headers=headers)

        return OnlineSearchCiteResponse(**res)
    else:
        param = request.model_dump()

        async def generate_stream():
            from app.logic.tool.online_search_cite_tool import (
                get_answer,
                get_completion_stream,
                get_search_results,
            )

            search_results = await get_search_results(param, headers)
            final_references = []
            ref_list = search_results["choices"][0]["message"]["tool_calls"][1][
                "search_result"
            ]
            for index, ref in enumerate(ref_list):
                ref_item = {
                    "title": ref.get("title", "未知标题"),
                    "content": ref.get("content", ""),
                    "link": ref.get("link", ""),
                    "index": index,
                }
                final_references.append(ref_item)

            full_answer = ""
            current_response = OnlineSearchCiteResponse(
                answer=full_answer,
                references=final_references,
            )
            yield f"{json.dumps(current_response.model_dump(), ensure_ascii=False)}"

            answer, _ = await get_answer(param, headers, search_results)

            async for chunk in get_completion_stream(
                param, headers, answer, final_references
            ):
                full_answer += chunk
                current_response = OnlineSearchCiteResponse(
                    answer=full_answer,
                    references=final_references,
                )
                yield f"{json.dumps(current_response.model_dump(), ensure_ascii=False)}"

        return EventSourceResponse(generate_stream(), ping=3600)
