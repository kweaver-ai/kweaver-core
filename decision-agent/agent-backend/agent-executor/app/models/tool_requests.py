from pydantic import BaseModel, Field


class ZhipuSearchRequest(BaseModel):
    """智谱搜索请求"""

    query: str = Field(..., description="搜索查询词", example="机器学习")


class SkillLoadRequest(BaseModel):
    """内置工具：加载 Skill 包请求"""

    skill_id: str = Field(..., description="Skill 标识符", example="my-skill-001")


class SkillReadFileRequest(BaseModel):
    """内置工具：读取 Skill 包内文件请求"""

    skill_id: str = Field(..., description="Skill 标识符", example="my-skill-001")
    file_path: str = Field(
        ..., description="Skill 包内文件的相对路径", example="references/guide.md"
    )


class SkillExecuteScriptRequest(BaseModel):
    """内置工具：执行 Skill 包脚本请求"""

    skill_id: str = Field(..., description="Skill 标识符", example="my-skill-001")
    entry_shell: str = Field(
        ...,
        description="来自 SKILL.md 的入口 Shell 命令",
        example="python scripts/analyze.py",
    )


class GetSchemaRequest(BaseModel):
    """获取模式请求"""

    database: str = Field(..., description="数据库名称", example="test_db")


class OnlineSearchCiteRequest(BaseModel):
    """联网搜索添加引用请求"""

    query: str = Field(..., description="搜索查询词", example="机器学习")
    model_name: str = Field(..., description="模型名称", example="deepseek-v3")
    search_tool: str = Field(..., description="搜索工具", example="zhipu_search_tool")
    api_key: str = Field(
        ...,
        description="搜索工具的API KEY",
        example="18286",
    )
    user_id: str = Field(..., description="userid", example="bdb7")
    stream: bool = Field(default=False, description="是否流式返回", example=False)
