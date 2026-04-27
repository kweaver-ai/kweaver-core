# -*- coding: utf-8 -*-
"""单元测试 - agent_cache_manage_logic/__init__ 模块"""


class TestAgentCacheManageLogicInit:
    """测试 agent_cache_manage_logic 模块导入"""

    def test_import_agent_cache_service(self):
        """测试导入 AgentCacheService"""
        from app.logic.agent_core_logic_v2.agent_cache_manage_logic import (
            AgentCacheService,
        )

        assert AgentCacheService is not None

    def test_import_agent_cache_manager(self):
        """测试导入 AgentCacheManager"""
        from app.logic.agent_core_logic_v2.agent_cache_manage_logic import (
            AgentCacheManager,
        )

        assert AgentCacheManager is not None

    def test_module_all_exports(self):
        """测试模块 __all__ 导出"""
        from app.logic.agent_core_logic_v2.agent_cache_manage_logic import __all__

        assert "AgentCacheService" in __all__
        assert "AgentCacheManager" in __all__
