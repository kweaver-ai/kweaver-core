# -*- coding: utf-8 -*-
"""
Unit tests for app/common/tool_v2/common module
"""

import pytest


class TestModuleImports:
    """Tests for module imports"""

    @pytest.mark.asyncio
    async def test_module_imports(self):
        """Test that common module can be imported"""
        from app.common.tool_v2.common import parse_kwargs

        assert callable(parse_kwargs)
