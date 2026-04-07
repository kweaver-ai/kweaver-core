from typing import Optional
from pydantic import BaseModel, Field


class V2RunAgentResponse(BaseModel):
    answer: dict = Field(..., title="answer", description="agent最终输出")
    status: str = Field(
        ...,
        title="status",
        description='"True": 流式信息已结束; "False": 流式信息未结束，正在返回; "Error": 失败',
    )
    ttft: Optional[int] = Field(
        None, title="ttft", description="Time to First Token（毫秒）"
    )
