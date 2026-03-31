"""单元测试 - config/builtin_ids_class 模块"""


class TestBuiltinIdsConfig:
    """测试 BuiltinIdsConfig 类"""

    def test_init_creates_default_agent_ids(self):
        """测试初始化创建默认agent IDs"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()

        assert "deepsearch" in config.agent_ids
        assert "OnlineSearch_Agent" in config.agent_ids
        assert "Plan_Agent" in config.agent_ids
        assert "SimpleChat_Agent" in config.agent_ids
        assert "Summary_Agent" in config.agent_ids

    def test_init_creates_default_tool_ids(self):
        """测试初始化创建默认tool IDs"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()

        assert "zhipu_search_tool" in config.tool_ids
        assert "check" in config.tool_ids
        assert "pass" in config.tool_ids

    def test_init_creates_default_tool_box_ids(self):
        """测试初始化创建默认toolbox IDs"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()

        assert "搜索工具" in config.tool_box_ids
        assert "数据处理工具" in config.tool_box_ids
        assert "文件处理工具" in config.tool_box_ids

    def test_get_agent_id_existing(self):
        """测试获取存在的agent ID"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        agent_id = config.get_agent_id("deepsearch")

        assert agent_id == "deepsearch"

    def test_get_agent_id_non_existing(self):
        """测试获取不存在的agent ID返回名称本身"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        agent_id = config.get_agent_id("nonexistent")

        assert agent_id == "nonexistent"

    def test_get_tool_id_existing(self):
        """测试获取存在的tool ID"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        tool_id = config.get_tool_id("zhipu_search_tool")

        assert tool_id == "zhipu_search_tool"

    def test_get_tool_id_non_existing(self):
        """测试获取不存在的tool ID返回名称本身"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        tool_id = config.get_tool_id("nonexistent")

        assert tool_id == "nonexistent"

    def test_get_tool_box_id_existing(self):
        """测试获取存在的toolbox ID"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        toolbox_id = config.get_tool_box_id("搜索工具")

        assert toolbox_id == "搜索工具"

    def test_get_tool_box_id_non_existing(self):
        """测试获取不存在的toolbox ID返回名称本身"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        toolbox_id = config.get_tool_box_id("nonexistent")

        assert toolbox_id == "nonexistent"

    def test_set_agent_id(self):
        """测试设置agent ID"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        config.set_agent_id("deepsearch", "new_deepsearch_id")

        assert config.get_agent_id("deepsearch") == "new_deepsearch_id"

    def test_set_tool_id(self):
        """测试设置tool ID"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        config.set_tool_id("zhipu_search_tool", "new_tool_id")

        assert config.get_tool_id("zhipu_search_tool") == "new_tool_id"

    def test_set_tool_box_id(self):
        """测试设置toolbox ID"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        config.set_tool_box_id("搜索工具", "new_toolbox_id")

        assert config.get_tool_box_id("搜索工具") == "new_toolbox_id"

    def test_get_all_agent_ids(self):
        """测试获取所有agent IDs"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        all_ids = config.get_all_agent_ids()

        assert isinstance(all_ids, dict)
        assert "deepsearch" in all_ids
        assert len(all_ids) == 5

    def test_get_all_tool_ids(self):
        """测试获取所有tool IDs"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        all_ids = config.get_all_tool_ids()

        assert isinstance(all_ids, dict)
        assert "zhipu_search_tool" in all_ids

    def test_get_all_tool_box_ids(self):
        """测试获取所有toolbox IDs"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        all_ids = config.get_all_tool_box_ids()

        assert isinstance(all_ids, dict)
        assert "搜索工具" in all_ids

    def test_get_all_returns_copy(self):
        """测试获取所有ID返回副本"""
        from app.config.builtin_ids_class import BuiltinIdsConfig

        config = BuiltinIdsConfig()
        all_ids = config.get_all_agent_ids()
        all_ids["deepsearch"] = "modified"

        # Original should not be modified
        assert config.agent_ids["deepsearch"] == "deepsearch"
