"""单元测试 - common/structs 模块 - Dataclass测试"""

from app.common.structs import (
    LogicBlock,
    AugmentBlock,
    RetrieverBlock,
)


class TestLogicBlock:
    """测试 LogicBlock 数据类"""

    def test_default_initialization(self):
        """测试默认初始化"""
        block = LogicBlock()
        assert block.id is None
        assert block.name is None
        assert block.type is None
        assert block.output is None
        assert block.llm_config is None

    def test_with_all_fields(self):
        """测试所有字段都有值"""
        block = LogicBlock(
            id="block_1",
            name="Test Block",
            type="llm_block",
            output="result",
            llm_config={"model": "gpt-4"},
        )
        assert block.id == "block_1"
        assert block.name == "Test Block"
        assert block.type == "llm_block"
        assert block.output == "result"
        assert block.llm_config == {"model": "gpt-4"}

    def test_partial_fields(self):
        """测试部分字段"""
        block = LogicBlock(id="block_1", type="retriever_block")
        assert block.id == "block_1"
        assert block.type == "retriever_block"
        assert block.name is None


class TestAugmentBlock:
    """测试 AugmentBlock 数据类"""

    def test_default_initialization(self):
        """测试默认初始化"""
        block = AugmentBlock()
        assert block.input == []
        assert block.augment_data_source == {}
        assert block.need_augment_content is False
        assert block.augment_entities == {}

    def test_with_input(self):
        """测试设置input"""
        block = AugmentBlock(input=["query1", "query2"])
        assert len(block.input) == 2
        assert block.input == ["query1", "query2"]

    def test_with_augment_data_source(self):
        """测试设置augment_data_source"""
        block = AugmentBlock(augment_data_source={"type": "concept"})
        assert block.augment_data_source == {"type": "concept"}

    def test_with_need_augment_content(self):
        """测试设置need_augment_content"""
        block = AugmentBlock(need_augment_content=True)
        assert block.need_augment_content is True

    def test_with_augment_entities(self):
        """测试设置augment_entities"""
        block = AugmentBlock(augment_entities={"entity1": "value1"})
        assert block.augment_entities == {"entity1": "value1"}

    def test_all_fields(self):
        """测试所有字段都有值"""
        block = AugmentBlock(
            input=["test"],
            augment_data_source={"type": "test"},
            need_augment_content=True,
            augment_entities={"key": "value"},
        )
        assert len(block.input) == 1
        assert block.need_augment_content is True
        assert block.augment_entities["key"] == "value"


class TestRetrieverBlock:
    """测试 RetrieverBlock 数据类"""

    def test_default_initialization(self):
        """测试默认初始化"""
        block = RetrieverBlock()
        # Inherited from LogicBlock
        assert block.id is None
        assert block.name is None
        assert block.type is None
        # RetrieverBlock specific fields
        assert block.input is None
        assert block.headers_info == {}
        assert block.body == {}
        assert block.data_source == {}
        assert block.augment_data_source == {}
        assert block.processed_query == {}
        assert block.retrival_slices == {}
        assert block.rank_slices == {}
        assert block.rank_rough_slices == {}
        assert block.rank_rough_slices_num == {}
        assert block.rank_accurate_slices == {}
        assert block.rank_accurate_slices_num == {}
        assert block.snippets_slices == {}
        assert block.cites_slices == {}
        assert block.format_out == []
        # Note: faq_retrival_qas and faq_rank_qas use default=list (class, not instance)
        assert block.faq_find_answer is False
        assert block.faq_format_out_qas == []
        assert block.security_token == set()

    def test_inherited_fields(self):
        """测试继承的字段"""
        block = RetrieverBlock(
            id="retriever_1",
            name="Test Retriever",
            type="retriever_block",
            output="search_result",
        )
        assert block.id == "retriever_1"
        assert block.name == "Test Retriever"
        assert block.type == "retriever_block"
        assert block.output == "search_result"

    def test_with_input(self):
        """测试设置input"""
        block = RetrieverBlock(input="test query")
        assert block.input == "test query"

    def test_with_headers_info(self):
        """测试设置headers_info"""
        block = RetrieverBlock(headers_info={"auth": "token"})
        assert block.headers_info == {"auth": "token"}

    def test_with_body(self):
        """测试设置body"""
        block = RetrieverBlock(body={"query": "test"})
        assert block.body == {"query": "test"}

    def test_with_data_source(self):
        """测试设置data_source"""
        block = RetrieverBlock(data_source={"type": "elasticsearch"})
        assert block.data_source == {"type": "elasticsearch"}

    def test_with_format_out(self):
        """测试设置format_out"""
        block = RetrieverBlock(format_out=["result1", "result2"])
        assert len(block.format_out) == 2
        assert block.format_out == ["result1", "result2"]

    def test_with_faq_fields(self):
        """测试FAQ相关字段"""
        block = RetrieverBlock(
            faq_retrival_qas=[("q1", "a1")],
            faq_rank_qas=[("q2", "a2")],
            faq_find_answer=True,
            faq_format_out_qas={"result": "data"},
        )
        assert len(block.faq_retrival_qas) == 1
        assert len(block.faq_rank_qas) == 1
        assert block.faq_find_answer is True
        assert block.faq_format_out_qas == {"result": "data"}

    def test_with_security_token(self):
        """测试设置security_token"""
        block = RetrieverBlock(security_token={"token1", "token2"})
        assert "token1" in block.security_token
        assert "token2" in block.security_token
        assert len(block.security_token) == 2

    def test_multiple_instances_independent(self):
        """测试多个实例独立性"""
        block1 = RetrieverBlock(id="block1")
        block2 = RetrieverBlock(id="block2")

        block1.input = "query1"
        block2.input = "query2"

        assert block1.input == "query1"
        assert block2.input == "query2"
        assert block1.input is not block2.input
