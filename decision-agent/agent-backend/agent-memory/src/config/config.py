from mem0.configs.base import MemoryConfig, LlmConfig, EmbedderConfig, VectorStoreConfig
from src.adaptee.mf_model_factory.rerank_model_client import Config as RerankConfig
from typing import Dict, Any
import yaml
import os
from pathlib import Path
import sys
from src.utils.env import getenv_int


class Config:
    _instance = None
    _initialized = False

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance

    def __init__(self):
        if not self._initialized:
            self._initialized = True
            if hasattr(sys, "_MEIPASS"):
                self.config_path = Path(sys._MEIPASS) / "config.yaml"
            else:
                self.config_path = Path(__file__).parent / "config.yaml"

            self.config = self._load_config()
            self._process_environment_variables()

    def _load_config(self) -> Dict[str, Any]:
        """加载配置文件"""
        with open(self.config_path, "r", encoding="utf-8") as f:
            config = yaml.safe_load(f)
        return config or {}

    def _process_environment_variables(self):
        """处理配置文件中的环境变量"""

        def replace_env_vars(value):
            if (
                isinstance(value, str)
                and value.startswith("${")
                and value.endswith("}")
            ):
                env_var = value[2:-1]
                return os.getenv(env_var) or value
            return value

        def process_dict(d: Dict[str, Any]):
            if not isinstance(d, dict):
                return
            for key, value in d.items():
                if isinstance(value, dict):
                    process_dict(value)
                else:
                    d[key] = replace_env_vars(value)

        process_dict(self.config)
        process_dict(self.config)

        # db config
        self.config["db"]["host"] = os.getenv("RDSHOST") or self.config["db"]["host"]
        self.config["db"]["port"] = getenv_int("RDSPORT") or self.config["db"]["port"]
        self.config["db"]["user"] = os.getenv("RDSUSER") or self.config["db"]["user"]
        self.config["db"]["password"] = (
            os.getenv("RDSPASS") or self.config["db"]["password"]
        )
        self.config["db"]["database"] = (
            os.getenv("RDSDBNAME") or self.config["db"]["database"]
        )

        # llm
        self.config["llm"]["base_url"] = (
            os.getenv("LLM_BASE_URL") or self.config["llm"]["base_url"]
        )
        self.config["llm"]["model"] = (
            os.getenv("LLM_MODEL") or self.config["llm"]["model"]
        )

        # embedding model
        self.config["embedder"]["model"] = (
            os.getenv("EMBEDDING_MODEL") or self.config["embedder"]["model"]
        )
        self.config["embedder"]["base_url"] = (
            os.getenv("EMBEDDING_MODEL_BASE_URL") or self.config["embedder"]["base_url"]
        )
        self.config["embedder"]["embedding_dims"] = (
            os.getenv("EMBEDDING_MODEL_DIMS")
            or self.config["embedder"]["embedding_dims"]
        )

        # vector store
        self.config["vector_store"]["host"] = (
            os.getenv("OPENSEARCH_HOST") or self.config["vector_store"]["host"]
        )
        self.config["vector_store"]["port"] = (
            os.getenv("OPENSEARCH_PORT") or self.config["vector_store"]["port"]
        )
        self.config["vector_store"]["user"] = (
            os.getenv("OPENSEARCH_USER") or self.config["vector_store"]["user"]
        )
        self.config["vector_store"]["password"] = (
            os.getenv("OPENSEARCH_PASS") or self.config["vector_store"]["password"]
        )
        self.config["vector_store"]["embedding_model_dims"] = (
            os.getenv("EMBEDDING_MODEL_DIMS")
            or self.config["vector_store"]["embedding_model_dims"]
        )

        # rerank
        self.config["rerank"]["rerank_url"] = (
            os.getenv("RERANK_URL") or self.config["rerank"]["rerank_url"]
        )
        self.config["rerank"]["rerank_model"] = (
            os.getenv("RERANK_MODEL") or self.config["rerank"]["rerank_model"]
        )

        # mem0
        if "mem0" in self.config:
            if "memory" not in self.config:
                self.config["memory"] = {}
            self.config["memory"]["max_tokens"] = int(self.config["mem0"]["max_tokens"])

        process_dict(self.config)

    def get(self, key: str, default: Any = None) -> Any:
        """获取配置值"""
        keys = key.split(".")
        value = self.config
        for k in keys:
            if isinstance(value, dict) and k in value:
                value = value[k]
            else:
                value = None
        return value or default

    def get_db_config(self) -> Dict[str, Any]:
        """获取数据库配置"""
        return {
            "host": self.config["db"]["host"],
            "port": self.config["db"]["port"],
            "user": self.config["db"]["user"],
            "password": self.config["db"]["password"],
            "database": self.config["db"]["database"],
        }

    def get_rerank_config(self) -> RerankConfig:
        """获取rerank配置"""
        return RerankConfig(
            rerank_url=self.config["rerank"]["rerank_url"],
            rerank_model=self.config["rerank"]["rerank_model"],
        )

    def _get_llm_config(self) -> LlmConfig:
        """获取LLM配置"""
        llm_config = self.get("llm", {})
        return LlmConfig(
            provider=llm_config.get("provider", "openai"),
            config={
                "model": llm_config.get("model", ""),
                "openai_base_url": llm_config.get("base_url", ""),
                "api_key": os.getenv("LLM_API_KEY") or llm_config.get("api_key", ""),
            },
        )

    def _get_embedder_config(self) -> EmbedderConfig:
        """获取Embedder配置"""
        embedder_config = self.get("embedder", {})
        return EmbedderConfig(
            provider=embedder_config.get("provider", "openai"),
            config={
                "model": embedder_config.get("model", "text-embedding-3-small"),
                "api_key": os.getenv("OPENAI_API_KEY")
                or embedder_config.get("api_key", ""),
                "openai_base_url": embedder_config.get("base_url", ""),
                "embedding_dims": embedder_config.get("embedding_dims", 768),
            },
        )

    def _get_vector_config(self) -> VectorStoreConfig:
        """获取VectorStore配置"""
        vector_config = self.get("vector_store", {})
        return VectorStoreConfig(
            provider=vector_config.get("provider", "opensearch"),
            config={
                "host": vector_config.get("host", ""),
                "port": vector_config.get("port", 9200),
                "collection_name": vector_config.get("collection_name", "mem0"),
                "user": vector_config.get("user", ""),
                "password": vector_config.get("password", ""),
                "embedding_model_dims": vector_config.get("embedding_model_dims", 768),
            },
        )

    def get_memory_config(self) -> MemoryConfig:
        """获取完整的Memory配置"""
        # 获取各个子配置
        llm_config = self._get_llm_config()
        embedder_config = self._get_embedder_config()
        vector_config = self._get_vector_config()
        memory_dict = self.config.get("memory", {})

        # 构建完整的Memory配置
        return MemoryConfig(
            llm=llm_config,
            embedder=embedder_config,
            vector_store=vector_config,
            history_db_path=memory_dict.get("history_db_path", "data/history.db"),
        )
