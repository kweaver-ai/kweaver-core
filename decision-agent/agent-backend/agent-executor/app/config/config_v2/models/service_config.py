"""
服务地址相关配置
"""

from dataclasses import dataclass


@dataclass
class ServiceEndpoint:
    """服务端点配置"""

    host: str = ""
    port: str = ""


@dataclass
class ServicesConfig:
    """依赖服务配置"""

    # 模型相关服务
    mf_model_api: ServiceEndpoint = None

    # Agent相关服务
    agent_executor: ServiceEndpoint = None
    agent_factory: ServiceEndpoint = None
    agent_operator_integration: ServiceEndpoint = None
    agent_memory: ServiceEndpoint = None

    def __post_init__(self):
        """初始化默认值"""
        if self.mf_model_api is None:
            self.mf_model_api = ServiceEndpoint("mf-model-api", "9898")
        if self.agent_executor is None:
            self.agent_executor = ServiceEndpoint("agent-executor", "30778")
        if self.agent_factory is None:
            self.agent_factory = ServiceEndpoint("agent-factory", "13020")
        if self.agent_operator_integration is None:
            self.agent_operator_integration = ServiceEndpoint(
                "agent-operator-integration", "9000"
            )
        if self.agent_memory is None:
            self.agent_memory = ServiceEndpoint("agent-memory", "30790")

    @classmethod
    def from_dict(cls, data: dict) -> "ServicesConfig":
        """从字典创建配置对象"""

        def get_endpoint(service_data: dict) -> ServiceEndpoint:
            return ServiceEndpoint(
                host=service_data.get("host", ""),
                port=str(service_data.get("port", "")),
            )

        return cls(
            mf_model_api=get_endpoint(data.get("mf_model_api", {})),
            agent_executor=get_endpoint(data.get("agent_executor", {})),
            agent_factory=get_endpoint(data.get("agent_factory", {})),
            agent_operator_integration=get_endpoint(
                data.get("agent_operator_integration", {})
            ),
            agent_memory=get_endpoint(data.get("agent_memory", {})),
        )
