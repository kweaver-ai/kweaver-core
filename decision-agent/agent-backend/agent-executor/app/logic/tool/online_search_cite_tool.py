import aiohttp
import json
from typing import AsyncGenerator
import uuid

answer_prompt = """

<references>
{references}
</references>

<query>
{query}
</query>

请根据<references>和<query>生成答案
要求：
1. 详细
2. 准确
"""


sys_prompt = """
<example>
    {{
        "index": 6,
        "link": "http://mp.weixin.qq.com/s?__biz=MzIyMjU3MDQ5NA==&mid=2247485368&idx=3&sn=16d66ba9277c7936d92172eaf14567f5",
        "media": "微信公众平台",
        "refer": "ref_7",
        "title": "golang: 用sync.Mutex实现线程安全",
        "content": "何为 sync.Mutex？如何使用 sync.Mutex？package confvarversionstringpackage confimport \"sync\"var ( version string mu sync.Mutex)func SetVersion(v string) mu.Lock() // 在修改 version 之前获取锁 version v mu.Unlock() // 修改完成后释放锁func GetVersion() string mu.Lock() // 在读取 version 之前获取锁 v : version mu.Unlock() // 读取完成后释放锁 return v注意事项读操作是并发安全此外,也可以使用 sync.RWMutex，它对于读操作是并发安全的，只有在写操作时才需要完全锁定。这在读多写少的场景下更有效。以下是如何使用 sync.RWMutex：package confimport \"sync\"var ( version string mu sync.RWMutex)func SetVersion(v string) mu.Lock() // 在修改 version 之前获取锁 version v mu.Unlock() // 修改完成后释放锁func GetVersion() string mu.RLock() // 在读取 version 之前获取读锁 v : version mu.RUnlock() // 读取完成后释放读锁 return v这样，当有多个 goroutine 同",
        "icon": "https://sfile.chatglm.cn/searchImage/mp_weixin_qq_com_icon_new.jpg"
    }}
</example>

<example_text_with_citation>
    Go语言提供了强大的并发编程能力，其中同步锁是控制并发访问共享资源的核心机制。本文将详细介绍Go语言中的同步锁实现，包括互斥锁和读写锁。

    ## 1. 互斥锁(sync.Mutex)

    ### 1.1 基本概念

    互斥锁(Mutex)是Go语言中最基础的同步原语，用于保护共享资源的独占访问。当一个goroutine获取了锁后，其他goroutine必须等待该锁被释放才能继续执行<i index="0"> 0 </i>。
</example_text_with_citation>

<ref_rule>
    网页资料的样式<example>中的样式一致，资料标记为index字段，资料内容为content字段。
    使用index字段作为引用下标，通过下标跳转到引用内容。
    引用格式必须严格使用<ref_style>中的模板：将数字替换为实际的index值，例如<i index="实际index"> 实际index </i>。不要省略<i>标签或index属性。
</ref_rule>

<synthesized_text>
{answer}
</synthesized_text>

<references>
{references}
</references>

<ref_style>
    <i index="1"> 1 </i>
</ref_style>


你是一位为报告添加正确引用的专家。你得到了一份联网搜索报告，该报告位于 <synthesized_text> 标签内，是基于提供的网页资料生成的。然而，这些资料并未在 <synthesized_text> 中被引用。你的任务是通过为这份报告生成正确且合适的引用，来增强用户的信任。
根据提供的引用数据<references>和规则<ref_rule>，在输入文本中添加引用，保证可以通过下标跳转到引用内容。请输出修改后的报告，除了添加的引用外，其余内容保持不变。
规则：
不要以任何方式修改 <synthesized_text> —— 保持所有内容 100% 一致，只添加引用
注意空白：不要添加或删除任何空白
只在源文件直接支持文本中的声明时添加引用
引用指南：
避免不必要的引用：并非每个陈述都需要引用。专注于引用关键事实、结论和实质性主张，这些内容与来源相关，而不是常识。优先引用读者想要验证的、增加论点可信度的或与特定来源明显相关的主张
引用有意义的语义单元：引用应涵盖完整的观点、发现或主张，这些内容作为独立的陈述是有意义的。避免引用单个单词或小短语片段，因为这些内容脱离上下文会失去意义；最好在句子末尾添加引用
最小化句子碎片化：避免在单个句子中添加多个引用，这会破坏句子的连贯性。只有在需要将句子中的特定主张归因于特定来源时，才在句子内的短语之间添加引用
避免相邻的冗余引用：不要在同一个句子中对同一来源进行多次引用，因为这是冗余且不必要的。如果一个句子包含来自同一来源的多个可引用主张，只需在句子末尾、句号之后使用一个引用即可
技术要求：
引用会在关闭标签处产生一个可视化的、可交互的元素。注意关闭标签的位置，不要不必要地拆分短语和句子
没有引用的文本将被收集并与 <synthesized_text> 中的原始报告进行比较。如果文本不一致，你的结果将被拒绝。
<example_text_with_citation>是参考格式，现在，请添加联网搜索报告的引用，并输出添加引用后的报告,不要输出任何其他内容。

"""


# 调用搜索工具获取网页搜索结果 - 非流式版本
async def get_search_results(request, headers):
    if request["search_tool"] == "zhipu_search_tool":
        tool = "web-search-pro"
        url = "https://open.bigmodel.cn/api/paas/v4/tools"
        request_id = str(uuid.uuid4())
        data = {
            "request_id": request_id,
            "tool": tool,
            "stream": False,
            "messages": [{"role": "user", "content": request["query"]}],
        }
        headers = {"Authorization": request["api_key"]}
        async with aiohttp.ClientSession() as session:
            async with session.post(url, headers=headers, json=data) as response:
                return await response.json()
    else:
        raise ValueError(f"不支持的搜索工具: {request['search_tool']}")


# 根据query和search_results生成answer
async def get_answer(request, headers, search_results) -> (str, list):
    references = []
    references_str = ""

    # Validate search_results structure before accessing nested elements
    try:
        ref_list = search_results["choices"][0]["message"]["tool_calls"][1][
            "search_result"
        ]
    except (KeyError, IndexError, TypeError) as e:
        raise ValueError(f"Invalid search_results structure: {e}") from e

    count = 0
    for ref in ref_list:
        # 修改为创建新 dict，并为缺失字段提供默认值，以匹配 ReferenceResult 模型
        ref_item = {
            "title": ref.get("title", "未知标题"),
            "content": ref.get("content", ""),
            "link": ref.get("link", ""),
            "index": count,
        }
        count = count + 1
        references.append(ref_item)
        references_str += ref["content"] + "\n"

    formatted_prompt = answer_prompt.format(
        query=request["query"], references=references_str
    )

    # 使用 model_api_service 调用大模型
    from app.driven.dip.model_api_service import model_api_service

    messages = [{"role": "user", "content": formatted_prompt}]
    result = await model_api_service.call(
        model=request["model_name"],
        messages=messages,
        max_tokens=10000,
        userid=headers.get("userid", ""),
    )
    return result, references


async def data_pre_answer(answer, references) -> (str, str):
    result = answer
    references = references
    # sorted_references = []
    # count = 0
    # for reference in references:
    #     reference["index"] = count
    #     count += 1
    #     sorted_references.append(reference)
    # references = sorted_references
    return result, json.dumps(references, ensure_ascii=False)


async def get_completion(request, headers, answer, references):
    answer_str, references_str = await data_pre_answer(answer, references)
    # 保存format的结果
    formatted_prompt = sys_prompt.format(answer=answer_str, references=references_str)

    # 使用 model_api_service 调用大模型
    from app.driven.dip.model_api_service import model_api_service

    messages = [{"role": "user", "content": formatted_prompt}]
    result = await model_api_service.call(
        model=request["model_name"],
        messages=messages,
        max_tokens=10000,
        userid=headers.get("userid", ""),
    )
    return result


async def get_completion_stream(
    request, headers, answer, references
) -> AsyncGenerator[str, None]:
    answer_str, references_str = await data_pre_answer(answer, references)
    # 保存format的结果
    formatted_prompt = sys_prompt.format(answer=answer_str, references=references_str)

    # 使用 model_api_service 调用大模型
    from app.driven.dip.model_api_service import model_api_service

    messages = [{"role": "user", "content": formatted_prompt}]

    async for content in model_api_service.call_stream(
        model=request["model_name"],
        messages=messages,
        max_tokens=10000,
        userid=headers.get("userid", ""),
    ):
        yield content


# 非流式
async def online_search_cite_tool(request, headers):
    # 1. 调用搜索工具获取网页搜索结果，获取搜索结果
    search_results = await get_search_results(request, headers)
    # 2. 根据搜索结果生成answer
    answer, references = await get_answer(request, headers, search_results)
    # 3. 添加引用并返回answer和引用内容
    result = await get_completion(request, headers, answer, references)
    return dict(answer=result, references=references)


# 流式模式已移至控制器中直接处理，避免重复获取搜索结果
