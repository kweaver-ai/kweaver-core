from typing import Optional
from pydantic import BaseModel, Field, model_validator
from app.domain.vo.agentvo import AgentConfigVo, AgentInputVo, AgentRunOptionsVo
from app.infra.common.helper.env_helper import is_aaron_local_dev


class V2RunAgentReq(BaseModel):
    agent_id: str = Field(
        default=None,
        title="agent_id",
        description="agent id",
        example="1830930776523276288",
    )

    agent_version: Optional[str] = Field(
        default=None,
        title="agent_version",
        description="agent版本号,与agent_id配合使用",
        example="latest",
    )

    agent_config: AgentConfigVo = Field(
        default=None,
        title="agent_config",
        description="agent配置",
        # example=docqa_template.agent_config,
    )

    agent_input: AgentInputVo = Field(
        ..., title="agent_input", description="agent输入", example={"query": "你好"}
    )

    options: AgentRunOptionsVo = Field(
        default=None,
        title="options",
        description="agent运行选项",
        alias="_options",
    )

    conversation_session_id: Optional[str] = Field(
        None,
        title="conversation_session_id",
        description="对话Session ID，如果提供则使用缓存的Agent",
        json_schema_extra={
            "example": "agent_executor:session_cache:user_123:conv_456:1234567890:uuid"
        },
    )

    @model_validator(mode="after")
    def validate_agent_params(self) -> "V2RunAgentReq":
        """验证agent_id和agent_config至少有一个"""
        if not self.agent_id and not self.agent_config:
            raise ValueError("agent_id和agent_config至少需要提供一个")

        if self.agent_id and not self.agent_version:
            # 在 aaron local dev 环境下跳过验证
            if is_aaron_local_dev():
                pass
            else:
                raise ValueError("当agent_id提供时，agent_version不能为空")

        return self
