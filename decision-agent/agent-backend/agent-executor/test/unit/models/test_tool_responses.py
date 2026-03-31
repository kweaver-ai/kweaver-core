"""单元测试 - models/tool_responses 模块"""


class TestZhipuSearchResponse:
    """测试 ZhipuSearchResponse 模型"""

    def test_default_initialization(self):
        """测试默认初始化"""
        from app.models.tool_responses import ZhipuSearchResponse

        response = ZhipuSearchResponse(
            choices=[{"result": "test"}],
            created=1234567890,
            id="test_id",
            model="test_model",
            request_id="req_id",
            usage={"tokens": 100},
        )

        assert response.choices == [{"result": "test"}]
        assert response.created == 1234567890
        assert response.id == "test_id"
        assert response.model == "test_model"
        assert response.request_id == "req_id"
        assert response.usage == {"tokens": 100}

    def test_with_empty_choices(self):
        """测试空choices"""
        from app.models.tool_responses import ZhipuSearchResponse

        response = ZhipuSearchResponse(
            choices=[],
            created=1234567890,
            id="test_id",
            model="test_model",
            request_id="req_id",
            usage={},
        )

        assert response.choices == []

    def test_model_dump(self):
        """测试模型序列化"""
        from app.models.tool_responses import ZhipuSearchResponse

        response = ZhipuSearchResponse(
            choices=[{"result": "test"}],
            created=1234567890,
            id="test_id",
            model="test_model",
            request_id="req_id",
            usage={},
        )
        data = response.model_dump()

        assert data["choices"] == [{"result": "test"}]
        assert data["id"] == "test_id"


class TestReferenceResult:
    """测试 ReferenceResult 模型"""

    def test_default_initialization(self):
        """测试默认初始化"""
        from app.models.tool_responses import ReferenceResult

        ref = ReferenceResult(
            title="Test Title",
            content="Test Content",
            index=0,
            link="https://example.com",
        )

        assert ref.title == "Test Title"
        assert ref.content == "Test Content"
        assert ref.index == 0
        assert ref.link == "https://example.com"

    def test_with_multiple_references(self):
        """测试多个引用"""
        from app.models.tool_responses import ReferenceResult

        ref1 = ReferenceResult(
            title="Title 1", content="Content 1", index=0, link="https://example.com/1"
        )

        ref2 = ReferenceResult(
            title="Title 2", content="Content 2", index=1, link="https://example.com/2"
        )

        assert ref1.index == 0
        assert ref2.index == 1

    def test_model_dump(self):
        """测试模型序列化"""
        from app.models.tool_responses import ReferenceResult

        ref = ReferenceResult(
            title="Test", content="Content", index=0, link="https://test.com"
        )
        data = ref.model_dump()

        assert data["title"] == "Test"
        assert data["index"] == 0


class TestOnlineSearchCiteResponse:
    """测试 OnlineSearchCiteResponse 模型"""

    def test_default_initialization(self):
        """测试默认初始化"""
        from app.models.tool_responses import OnlineSearchCiteResponse, ReferenceResult

        refs = [
            ReferenceResult(
                title="Test", content="Content", index=0, link="https://test.com"
            )
        ]

        response = OnlineSearchCiteResponse(references=refs, answer="Test answer")

        assert len(response.references) == 1
        assert response.answer == "Test answer"

    def test_with_multiple_references(self):
        """测试多个引用"""
        from app.models.tool_responses import OnlineSearchCiteResponse, ReferenceResult

        refs = [
            ReferenceResult(
                title=f"Title {i}",
                content=f"Content {i}",
                index=i,
                link=f"https://test.com/{i}",
            )
            for i in range(3)
        ]

        response = OnlineSearchCiteResponse(
            references=refs, answer="Answer with multiple references"
        )

        assert len(response.references) == 3

    def test_with_empty_references(self):
        """测试空引用列表"""
        from app.models.tool_responses import OnlineSearchCiteResponse

        response = OnlineSearchCiteResponse(
            references=[], answer="Answer without references"
        )

        assert response.references == []

    def test_model_dump(self):
        """测试模型序列化"""
        from app.models.tool_responses import OnlineSearchCiteResponse, ReferenceResult

        refs = [
            ReferenceResult(
                title="Test", content="Content", index=0, link="https://test.com"
            )
        ]

        response = OnlineSearchCiteResponse(references=refs, answer="Answer")
        data = response.model_dump()

        assert data["answer"] == "Answer"
        assert len(data["references"]) == 1


class TestNL2NGQLResponse:
    """测试 NL2NGQLResponse 模型"""

    def test_default_initialization(self):
        """测试默认初始化"""
        from app.models.tool_responses import NL2NGQLResponse

        response = NL2NGQLResponse(outputs=[{"query": "MATCH (n) RETURN n"}])

        assert response.outputs == [{"query": "MATCH (n) RETURN n"}]

    def test_with_multiple_outputs(self):
        """测试多个输出"""
        from app.models.tool_responses import NL2NGQLResponse

        response = NL2NGQLResponse(
            outputs=[
                {"query": "MATCH (n) RETURN n"},
                {"query": "MATCH (n:Person) RETURN n"},
            ]
        )

        assert len(response.outputs) == 2

    def test_with_empty_outputs(self):
        """测试空输出"""
        from app.models.tool_responses import NL2NGQLResponse

        response = NL2NGQLResponse(outputs=[])

        assert response.outputs == []

    def test_model_dump(self):
        """测试模型序列化"""
        from app.models.tool_responses import NL2NGQLResponse

        response = NL2NGQLResponse(outputs=[{"result": "test"}])
        data = response.model_dump()

        assert data["outputs"] == [{"result": "test"}]


class TestSchemaInfo:
    """测试 SchemaInfo 模型"""

    def test_default_initialization(self):
        """测试默认初始化"""
        from app.models.tool_responses import SchemaInfo

        # Use 'schema' as the field name (it's an alias for schema_data)
        schema = SchemaInfo(schema={"nodes": ["Person", "Company"]})

        assert schema.schema_data == {"nodes": ["Person", "Company"]}

    def test_with_complex_schema(self):
        """测试复杂模式"""
        from app.models.tool_responses import SchemaInfo

        schema = SchemaInfo(
            schema={
                "nodes": ["Person", "Company"],
                "edges": ["WORKS_AT", "KNOWS"],
                "properties": {
                    "Person": ["name", "age"],
                    "Company": ["name", "founded"],
                },
            }
        )

        assert "Person" in schema.schema_data["nodes"]
        assert "WORKS_AT" in schema.schema_data["edges"]

    def test_alias_works(self):
        """测试别名工作"""
        from app.models.tool_responses import SchemaInfo

        # Using 'schema' as alias for 'schema_data'
        schema = SchemaInfo(schema={"test": "value"})

        assert schema.schema_data == {"test": "value"}

    def test_model_dump(self):
        """测试模型序列化"""
        from app.models.tool_responses import SchemaInfo

        schema = SchemaInfo(schema={"test": "value"})
        data = schema.model_dump()

        # Should use the original field name
        assert data["schema_data"] == {"test": "value"}

    def test_model_dump_json(self):
        """测试JSON序列化"""
        from app.models.tool_responses import SchemaInfo

        schema = SchemaInfo(schema={"test": "value"})
        json_str = schema.model_dump_json()

        assert "test" in json_str

    def test_model_dump_by_alias(self):
        """测试使用别名序列化"""
        from app.models.tool_responses import SchemaInfo

        schema = SchemaInfo(schema={"test": "value"})
        data = schema.model_dump(by_alias=True)

        # Should use alias 'schema' when by_alias=True
        assert "schema" in data


# Extended tests for coverage improvement


class TestZhipuSearchResponseExtended:
    """测试 ZhipuSearchResponse 模型的扩展测试"""

    def test_unicode_in_choices(self):
        """测试choices中的Unicode内容"""
        from app.models.tool_responses import ZhipuSearchResponse

        response = ZhipuSearchResponse(
            choices=[{"result": "测试结果", "content": "中文内容"}],
            created=1234567890,
            id="test_id",
            model="test_model",
            request_id="req_id",
            usage={},
        )
        assert "测试结果" in str(response.choices[0]["result"])

    def test_nested_dict_in_choices(self):
        """测试嵌套字典的choices"""
        from app.models.tool_responses import ZhipuSearchResponse

        response = ZhipuSearchResponse(
            choices=[
                {
                    "result": "test",
                    "metadata": {"key": "value", "nested": {"deep": "value"}},
                    "items": [1, 2, 3],
                }
            ],
            created=1234567890,
            id="test_id",
            model="test_model",
            request_id="req_id",
            usage={},
        )
        assert response.choices[0]["metadata"]["nested"]["deep"] == "value"

    def test_empty_usage_dict(self):
        """测试空的usage字典"""
        from app.models.tool_responses import ZhipuSearchResponse

        response = ZhipuSearchResponse(
            choices=[],
            created=1234567890,
            id="test_id",
            model="test_model",
            request_id="req_id",
            usage={},
        )
        assert response.usage == {}


class TestReferenceResultExtended:
    """测试 ReferenceResult 模型的扩展测试"""

    def test_unicode_content(self):
        """测试Unicode内容"""
        from app.models.tool_responses import ReferenceResult

        ref = ReferenceResult(
            title="测试标题",
            content="测试内容",
            index=0,
            link="https://example.com/测试",
        )
        assert ref.title == "测试标题"
        assert ref.content == "测试内容"

    def test_special_characters_in_title(self):
        """测试标题中的特殊字符"""
        from app.models.tool_responses import ReferenceResult

        ref = ReferenceResult(
            title="Title <script>alert('xss')</script>",
            content="Content",
            index=0,
            link="https://example.com",
        )
        assert "<script>" in ref.title

    def test_negative_index(self):
        """测试负数索引"""
        from app.models.tool_responses import ReferenceResult

        ref = ReferenceResult(
            title="Test", content="Content", index=-1, link="https://example.com"
        )
        assert ref.index == -1

    def test_url_with_special_chars(self):
        """测试包含特殊字符的URL"""
        from app.models.tool_responses import ReferenceResult

        ref = ReferenceResult(
            title="Test",
            content="Content",
            index=0,
            link="https://example.com/path?query=value&other=123#fragment",
        )
        assert "?" in ref.link
        assert "#" in ref.link


class TestOnlineSearchCiteResponseExtended:
    """测试 OnlineSearchCiteResponse 模型的扩展测试"""

    def test_unicode_answer(self):
        """测试Unicode答案"""
        from app.models.tool_responses import OnlineSearchCiteResponse

        response = OnlineSearchCiteResponse(
            references=[], answer="这是中文答案，带有测试内容。"
        )
        assert "中文答案" in response.answer

    def test_answer_with_newlines(self):
        """测试包含换行符的答案"""
        from app.models.tool_responses import OnlineSearchCiteResponse

        response = OnlineSearchCiteResponse(
            references=[], answer="Line 1\nLine 2\nLine 3"
        )
        assert "\n" in response.answer

    def test_answer_with_special_characters(self):
        """测试包含特殊字符的答案"""
        from app.models.tool_responses import OnlineSearchCiteResponse

        response = OnlineSearchCiteResponse(
            references=[], answer="Answer with <tags> & \"quotes\" and 'apostrophes'"
        )
        assert "<tags>" in response.answer


class TestNL2NGQLResponseExtended:
    """测试 NL2NGQLResponse 模型的扩展测试"""

    def test_unicode_in_outputs(self):
        """测试outputs中的Unicode内容"""
        from app.models.tool_responses import NL2NGQLResponse

        response = NL2NGQLResponse(outputs=[{"query": "MATCH (n:人名) RETURN n"}])
        assert "人名" in response.outputs[0]["query"]

    def test_nested_dict_outputs(self):
        """测试嵌套字典的outputs"""
        from app.models.tool_responses import NL2NGQLResponse

        response = NL2NGQLResponse(
            outputs=[
                {
                    "query": "MATCH (n) RETURN n",
                    "params": {"name": "测试", "count": 123},
                    "nested": {"key": {"deep": "value"}},
                }
            ]
        )
        assert response.outputs[0]["params"]["name"] == "测试"

    def test_empty_dict_in_outputs(self):
        """测试空字典在outputs中"""
        from app.models.tool_responses import NL2NGQLResponse

        response = NL2NGQLResponse(outputs=[{}])
        assert response.outputs == [{}]


class TestSchemaInfoExtended:
    """测试 SchemaInfo 模型的扩展测试"""

    def test_unicode_in_schema(self):
        """测试schema中的Unicode内容"""
        from app.models.tool_responses import SchemaInfo

        schema = SchemaInfo(
            schema={
                "节点类型": ["人", "公司"],
                "属性": {"姓名": "string", "年龄": "int"},
            }
        )
        assert "节点类型" in schema.schema_data

    def test_nested_structure(self):
        """测试嵌套结构"""
        from app.models.tool_responses import SchemaInfo

        schema = SchemaInfo(schema={"level1": {"level2": {"level3": "deep value"}}})
        assert schema.schema_data["level1"]["level2"]["level3"] == "deep value"

    def test_list_values_in_schema(self):
        """测试schema中的列表值"""
        from app.models.tool_responses import SchemaInfo

        schema = SchemaInfo(schema={"items": [1, 2, 3, "four", {"five": 5}]})
        assert len(schema.schema_data["items"]) == 5
