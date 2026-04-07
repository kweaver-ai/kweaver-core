import uuid
import aiohttp
import logging

logger = logging.getLogger(__name__)


async def zhipu_search_tool(inputs, props, resource, data_source_config, context=None):
    tool = "web-search-pro"
    url = "https://open.bigmodel.cn/api/paas/v4/tools"
    request_id = str(uuid.uuid4())
    data = {
        "request_id": request_id,
        "tool": tool,
        "stream": False,
        "messages": [{"role": "user", "content": inputs["query"]}],
    }
    headers = {"Authorization": props.get("api_key")}

    async with aiohttp.ClientSession() as session:
        try:
            async with session.post(url, json=data, headers=headers) as response:
                # Handle successful responses (2xx status codes)
                if 200 <= response.status < 300:
                    result = await response.json()
                    return result
                else:
                    error_text = await response.text()
                    logger.error(
                        f"zhipu_search_tool request failed with status {response.status}: {error_text}"
                    )
                    return {
                        "error": f"Request failed with status {response.status}",
                        "details": error_text,
                    }
        except aiohttp.ClientError as e:
            logger.error(f"zhipu_search_tool request error: {e}")
            return {"error": f"Request failed: {str(e)}"}
        except Exception as e:
            logger.error(f"zhipu_search_tool unexpected error: {e}")
            return {"error": f"Unexpected error: {str(e)}"}
