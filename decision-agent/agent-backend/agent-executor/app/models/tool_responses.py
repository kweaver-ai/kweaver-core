from typing import Any, Dict, List, Optional

from pydantic import BaseModel, Field


class ZhipuSearchResponse(BaseModel):
    """智谱搜索响应"""

    """样例
    {'choices': [{...}], 'created': 1749713107, 'id': '20250612152507b7e80ee738244e30', 'model': 'web-search-pro', 'request_id': 'a983a178-7f9a-4ef2-a4d0-6cb5e17570bf', 'usage': {'completion_tokens': 3000, 'prompt_tokens': 0, 'total_tokens': 3000}}
    """

    choices: List[Dict[str, Any]] = Field(..., description="搜索结果列表")
    created: int = Field(..., description="创建时间")
    id: str = Field(..., description="搜索ID")
    model: str = Field(..., description="模型")
    request_id: str = Field(..., description="请求ID")
    usage: Dict[str, Any] = Field(..., description="使用情况统计")


class ReferenceResult(BaseModel):
    """引用内容"""

    title: str = Field(..., description="引用标题")
    content: str = Field(..., description="引用内容")
    index: int = Field(..., description="引用索引")
    link: str = Field(..., description="引用链接")


class OnlineSearchCiteResponse(BaseModel):
    """联网搜索添加引用响应"""

    references: List[ReferenceResult] = Field(..., description="引用内容列表")
    answer: str = Field(..., description="添加引用标记的回答")


class NL2NGQLResponse(BaseModel):
    """NL2NGQL响应"""

    outputs: List[Dict[str, Any]] = Field(..., description="输出结果列表")


class SchemaInfo(BaseModel):
    """图数据库模式信息"""

    schema_data: Dict[str, Any] = Field(..., description="数据库模式", alias="schema")


class SkillLoadResponse(BaseModel):
    """内置工具：加载 Skill 包响应"""

    skill_id: str = Field(..., description="Skill 标识符")
    skill_md_content: str = Field(..., description="SKILL.md 文件的完整文本内容")
    available_scripts: List[str] = Field(default_factory=list, description="可执行脚本路径列表")
    available_references: List[str] = Field(default_factory=list, description="参考文件路径列表")
    source: str = Field(default="factory", description="数据来源标识")
    error: Optional[str] = Field(default=None, description="错误信息（成功时为 null）")


class SkillReadFileResponse(BaseModel):
    """内置工具：读取 Skill 包内文件响应"""

    skill_id: str = Field(..., description="Skill 标识符")
    file_path: str = Field(..., description="读取的文件相对路径")
    content: str = Field(..., description="文件文本内容")
    source: str = Field(default="factory", description="数据来源标识")
    error: Optional[str] = Field(default=None, description="错误信息（成功时为 null）")


class SkillExecuteScriptResponse(BaseModel):
    """内置工具：执行 Skill 包脚本响应"""

    skill_id: str = Field(..., description="Skill 标识符")
    entry_shell: str = Field(..., description="执行的入口 Shell 命令")
    command: str = Field(default="", description="实际执行的命令")
    stdout: str = Field(default="", description="标准输出内容")
    stderr: str = Field(default="", description="标准错误内容")
    exit_code: int = Field(default=-1, description="退出码（0 表示成功）")
    duration_ms: int = Field(default=0, description="执行耗时（毫秒）")
    artifacts: List[Dict[str, Any]] = Field(default_factory=list, description="产生的制品列表")
    source: str = Field(default="factory", description="数据来源标识")
    error: Optional[str] = Field(default=None, description="错误信息（成功时为 null）")
