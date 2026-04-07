name = "Task_Plan_Agent"

description = "可以分析用户的请求，将其分解为具体的可执行任务"

dolphin = """
/prompt/你是一个专业的任务规划代理。请分析用户的请求，将其分解为具体的可执行任务。

用户请求: {$query}

请按照以下格式返回JSON格式的任务列表(纯json字符串，不包含markdown语法等任何其他内容):
{
    "tasks": [
        {
            "id": "task_1",
            "action": "具体的操作描述",
            "description": "详细的任务描述",
            "type": "tool|agent|llm",
            "dependencies": []
        }
    ],
    "reasoning": "分解任务的reasoning过程"
}

任务类型说明:
- tool: 需要调用具体工具的任务
- agent: 需要调用其他代理的任务
- llm: 需要LLM直接处理的任务

请确保任务按执行顺序排列，每个任务都是明确可执行的。->plan_res

$plan_res.answer -> task_list
"""

tools = []

llms = [
    {
        "is_default": True,
        "llm_config": {
            "id": "1922491455683825664",
            "name": "Tome-pro",
            "temperature": 0.7,
            "top_p": 0.8,
            "top_k": 40,
            "frequency_penalty": 0.5,
            "presence_penalty": 0.3,
            "max_tokens": 200,
            "retrieval_max_tokens": 8,
        },
    }
]

""" built_in_can_edit_fields:
- 内置agent可编辑字段
- 当is_built_in为1时有效
- 用途：
    - 内置agent创建者可以根据需要来设置此字段，来控制用户可以编辑的agent配置字段
    - 前端根据此字段来控制用户可以编辑的agent配置字段
    - agent后端编辑接口根据此字段来验证前端传过来的字段是否符合要求
"""
built_in_can_edit_fields = {
    "name": False,
    "avatar": False,
    "profile": False,
    "input_config": False,
    "system_prompt": False,
    "knowledge_source": False,
    "model": True,
    "skills": False,
    "opening_remark_config": False,
    "preset_questions": False,
    "skills.tools.tool_input": False,
}

config = {
    "input": {
        "fields": [{"name": "query", "type": "string"}],
        "is_temp_zone_enabled": 0,
    },
    "system_prompt": "",
    "dolphin": dolphin,
    "pre_dolphin": [],
    "post_dolphin": [],
    "is_dolphin_mode": 1,
    "skills": {
        "tools": tools,
    },
    "llms": llms,
    "is_data_flow_set_enabled": 0,
    "output": {
        "variables": {"answer_var": "task_list"},
        "default_format": "markdown",
    },
    "built_in_can_edit_fields": built_in_can_edit_fields,
    "memory": {"is_enabled": False},
    "related_question": {"is_enabled": False},
    "plan_mode": {"is_enabled": False},
}
