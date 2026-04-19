# -*- coding: utf-8 -*-
"""单元测试 - input_handler_pkg/build_llm_config 模块"""

import importlib

import pytest
from unittest.mock import MagicMock, patch


def get_build_llm_config_module():
    return importlib.import_module(
        "app.logic.agent_core_logic_v2.input_handler_pkg.build_llm_config"
    )


class TestBuildLLMConfig:
    """测试 build_llm_config 函数"""

    @pytest.fixture
    def mock_agent_core(self):
        """创建 mock AgentCoreV2"""
        core = MagicMock()
        core.agent_config = MagicMock()
        core.agent_config.agent_run_id = "test_run_id"
        core.agent_config.agent_id = "test_agent_id"
        core.agent_config.llms = [
            {
                "is_default": True,
                "llm_config": {
                    "name": "gpt-4",
                    "temperature": 0.7,
                },
            }
        ]
        return core

    @pytest.mark.asyncio
    async def test_build_llm_config_basic(self, mock_agent_core):
        """测试基本LLM配置构建"""
        module = get_build_llm_config_module()

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "set_user_account_id"):
                with patch.object(module, "set_user_account_type"):
                    with patch.object(module.Config.local_dev, "is_use_outer_llm", False):
                        with patch.object(
                            mock_agent_core.agent_config,
                            "llms",
                            [
                                {
                                    "is_default": True,
                                    "llm_config": {
                                        "name": "gpt-4",
                                        "temperature": 0.7,
                                    },
                                }
                            ],
                        ):
                            build_llm_config = module.build_llm_config
                            result = await build_llm_config(
                                mock_agent_core,
                                user_id="user123",
                                visitor_type="standard",
                            )

                        assert "default" in result
                        assert "llms" in result
                        assert result["default"] == "gpt-4"

    @pytest.mark.asyncio
    async def test_build_llm_config_multiple_llms(self, mock_agent_core):
        """测试多个LLM配置"""
        mock_agent_core.agent_config.llms = [
            {
                "is_default": True,
                "llm_config": {
                    "name": "gpt-4",
                    "temperature": 0.7,
                },
            },
            {
                "is_default": False,
                "llm_config": {
                    "name": "claude-3",
                    "temperature": 0.5,
                },
            },
        ]

        module = get_build_llm_config_module()

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "set_user_account_id"):
                with patch.object(module, "set_user_account_type"):
                    with patch.object(module.Config.local_dev, "is_use_outer_llm", False):
                        build_llm_config = module.build_llm_config

                        result = await build_llm_config(
                            mock_agent_core, user_id="user123", visitor_type="standard"
                        )

                        assert result["default"] == "gpt-4"
                        assert "gpt-4" in result["llms"]
                        assert "claude-3" in result["llms"]

    @pytest.mark.asyncio
    async def test_build_llm_config_no_default(self, mock_agent_core):
        """测试没有默认LLM配置"""
        mock_agent_core.agent_config.llms = [
            {
                "is_default": False,
                "llm_config": {
                    "name": "gpt-4",
                    "temperature": 0.7,
                },
            }
        ]

        module = get_build_llm_config_module()

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "set_user_account_id"):
                with patch.object(module, "set_user_account_type"):
                    with patch.object(module.Config.local_dev, "is_use_outer_llm", False):
                        build_llm_config = module.build_llm_config

                        result = await build_llm_config(
                            mock_agent_core, user_id="user123", visitor_type="standard"
                        )

                        # 默认值应为空字符串
                        assert result["default"] == ""

    @pytest.mark.asyncio
    async def test_build_llm_config_empty_llms(self, mock_agent_core):
        """测试空LLM配置列表"""
        mock_agent_core.agent_config.llms = []

        module = get_build_llm_config_module()

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "set_user_account_id"):
                with patch.object(module, "set_user_account_type"):
                    build_llm_config = module.build_llm_config

                    result = await build_llm_config(
                        mock_agent_core, user_id="user123", visitor_type="standard"
                    )

                    assert result["default"] == ""
                    assert result["llms"] == {}

    @pytest.mark.asyncio
    async def test_build_llm_config_none_llms(self, mock_agent_core):
        """测试None LLM配置"""
        mock_agent_core.agent_config.llms = None

        module = get_build_llm_config_module()

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "set_user_account_id"):
                with patch.object(module, "set_user_account_type"):
                    build_llm_config = module.build_llm_config

                    result = await build_llm_config(
                        mock_agent_core, user_id="user123", visitor_type="standard"
                    )

                    assert result["default"] == ""
                    assert result["llms"] == {}


class TestGetLLMConfigFromCache:
    """测试 get_llm_config_from_cache 函数"""

    def test_get_llm_config_from_cache(self):
        """测试从缓存获取LLM配置"""
        mock_agent_core = MagicMock()
        mock_cache_handler = MagicMock()
        mock_cache_handler.get_llm_config.return_value = {"model": "gpt-4"}
        mock_agent_core.cache_handler = mock_cache_handler

        module = get_build_llm_config_module()

        with patch.object(
            module,
            "get_llm_config_from_cache",
            side_effect=lambda ac, llm_id: ac.cache_handler.get_llm_config(llm_id),
        ):
            get_llm_config_from_cache = module.get_llm_config_from_cache

            result = get_llm_config_from_cache(mock_agent_core, "llm123")
            assert result == {"model": "gpt-4"}
