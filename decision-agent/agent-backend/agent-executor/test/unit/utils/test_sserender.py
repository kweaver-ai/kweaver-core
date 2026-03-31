"""单元测试 - utils/sserender 模块"""

import pytest
import json


class TestSSEInit:
    """测试 SSE 初始化"""

    def test_init_with_id_only(self):
        """测试仅带ID初始化"""
        from app.utils.sserender import SSE

        sse = SSE(ID="msg123")
        assert sse.ID == "msg123"
        assert sse.event is None
        assert sse.data is None
        assert sse.retry is None
        assert sse.comment is None

    def test_init_with_event_only(self):
        """测试仅带event初始化"""
        from app.utils.sserender import SSE

        sse = SSE(event="message")
        assert sse.ID is None
        assert sse.event == "message"
        assert sse.data is None
        assert sse.retry is None
        assert sse.comment is None

    def test_init_with_data_only(self):
        """测试仅带data初始化"""
        from app.utils.sserender import SSE

        sse = SSE(data="Hello World")
        assert sse.ID is None
        assert sse.event is None
        assert sse.data == "Hello World"
        assert sse.retry is None
        assert sse.comment is None

    def test_init_with_comment_only(self):
        """测试仅带comment初始化"""
        from app.utils.sserender import SSE

        sse = SSE(comment="This is a comment")
        assert sse.ID is None
        assert sse.event is None
        assert sse.data is None
        assert sse.retry is None
        assert sse.comment == "This is a comment"

    def test_init_with_all_fields(self):
        """测试带所有字段初始化"""
        from app.utils.sserender import SSE

        sse = SSE(
            ID="msg123",
            event="message",
            data="test data",
            comment="test comment",
            retry=5000,
        )

        assert sse.ID == "msg123"
        assert sse.event == "message"
        assert sse.data == "test data"
        assert sse.comment == "test comment"
        assert sse.retry == 5000

    def test_init_with_list_data(self):
        """测试带列表data初始化"""
        from app.utils.sserender import SSE

        sse = SSE(data=["line1", "line2", "line3"])
        assert sse.data == ["line1", "line2", "line3"]

    def test_init_with_list_comment(self):
        """测试带列表comment初始化"""
        from app.utils.sserender import SSE

        sse = SSE(comment=["comment1", "comment2"])
        assert sse.comment == ["comment1", "comment2"]

    def test_init_with_no_arguments_raises_error(self):
        """测试无参数初始化抛出错误"""
        from app.utils.sserender import SSE

        with pytest.raises(ValueError, match="at least one argument"):
            SSE()

    def test_init_with_invalid_retry_type_raises_error(self):
        """测试无效的retry类型抛出错误"""
        from app.utils.sserender import SSE

        with pytest.raises(TypeError, match="retry argument must be int"):
            SSE(data="test", retry="not_an_int")


class TestSSERender:
    """测试 SSE render 方法"""

    def test_render_id_only(self):
        """测试渲染仅带ID的消息"""
        from app.utils.sserender import SSE

        sse = SSE(ID="msg123")
        result = sse.render()

        assert "id: msg123" in result
        assert result.endswith("\r\n\r\n")

    def test_render_event_only(self):
        """测试渲染仅带event的消息"""
        from app.utils.sserender import SSE

        sse = SSE(event="message")
        result = sse.render()

        assert "event: message" in result
        assert result.endswith("\r\n\r\n")

    def test_render_data_only(self):
        """测试渲染仅带data的消息"""
        from app.utils.sserender import SSE

        sse = SSE(data="Hello World")
        result = sse.render()

        assert "data: Hello World" in result
        assert result.endswith("\r\n\r\n")

    def test_render_comment_only(self):
        """测试渲染仅带comment的消息"""
        from app.utils.sserender import SSE

        sse = SSE(comment="This is a comment")
        result = sse.render()

        assert ": This is a comment" in result
        assert result.endswith("\r\n\r\n")

    def test_render_retry_only(self):
        """测试渲染仅带retry的消息"""
        from app.utils.sserender import SSE

        sse = SSE(event="test", retry=5000)
        result = sse.render()

        assert "retry: 5000" in result
        assert result.endswith("\r\n\r\n")

    def test_render_with_all_fields(self):
        """测试渲染带所有字段的消息"""
        from app.utils.sserender import SSE

        sse = SSE(
            ID="msg123",
            event="message",
            data="test data",
            comment="test comment",
            retry=5000,
        )
        result = sse.render()

        assert "id: msg123" in result
        assert "event: message" in result
        assert "data: test data" in result
        assert ": test comment" in result
        assert "retry: 5000" in result

    def test_render_multiline_data(self):
        """测试渲染多行data"""
        from app.utils.sserender import SSE

        sse = SSE(data="line1\nline2\nline3")
        result = sse.render()

        assert "data: line1" in result
        assert "data: line2" in result
        assert "data: line3" in result

    def test_render_list_data(self):
        """测试渲染列表data"""
        from app.utils.sserender import SSE

        sse = SSE(data=["line1", "line2", "line3"])
        result = sse.render()

        assert "data: line1" in result
        assert "data: line2" in result
        assert "data: line3" in result

    def test_render_multiline_comment(self):
        """测试渲染多行comment"""
        from app.utils.sserender import SSE

        sse = SSE(comment="comment1\ncomment2", event="test")
        result = sse.render()

        assert ": comment1" in result
        assert ": comment2" in result

    def test_render_with_encode_true(self):
        """测试渲染并编码"""
        from app.utils.sserender import SSE

        sse = SSE(data="test")
        result = sse.render(with_encode=True)

        assert isinstance(result, bytes)
        assert b"data: test" in result

    def test_render_with_encode_false(self):
        """测试渲染不编码"""
        from app.utils.sserender import SSE

        sse = SSE(data="test")
        result = sse.render(with_encode=False)

        assert isinstance(result, str)
        assert "data: test" in result


class TestSSEFromContent:
    """测试 SSE from_content 方法"""

    def test_from_content_with_id(self):
        """测试解析带ID的内容"""
        from app.utils.sserender import SSE

        content = "id: msg123\r\n\r\n"
        sse = SSE.from_content(content)

        assert sse.ID == "msg123"

    def test_from_content_with_event(self):
        """测试解析带event的内容"""
        from app.utils.sserender import SSE

        content = "event: message\r\n\r\n"
        sse = SSE.from_content(content)

        assert sse.event == "message"

    def test_from_content_with_data(self):
        """测试解析带data的内容"""
        from app.utils.sserender import SSE

        content = "data: Hello World\r\n\r\n"
        sse = SSE.from_content(content)

        assert sse.data == ["Hello World"]

    def test_from_content_with_retry(self):
        """测试解析带retry的内容"""
        from app.utils.sserender import SSE

        content = "event: test\r\nretry: 5000\r\n\r\n"
        sse = SSE.from_content(content)

        assert sse.retry == 5000

    def test_from_content_with_comment(self):
        """测试解析带comment的内容"""
        from app.utils.sserender import SSE

        content = ": This is a comment\r\n\r\n"
        sse = SSE.from_content(content)

        assert sse.comment == ["This is a comment"]

    def test_from_content_with_all_fields(self):
        """测试解析带所有字段的内容"""
        from app.utils.sserender import SSE

        content = "id: msg123\r\nevent: message\r\ndata: Hello\r\n: comment\r\nretry: 5000\r\n\r\n"
        sse = SSE.from_content(content)

        assert sse.ID == "msg123"
        assert sse.event == "message"
        assert sse.data == ["Hello"]
        assert sse.comment == ["comment"]
        assert sse.retry == 5000

    def test_from_content_with_multiline_data(self):
        """测试解析多行data"""
        from app.utils.sserender import SSE

        content = "data: line1\r\ndata: line2\r\ndata: line3\r\n\r\n"
        sse = SSE.from_content(content)

        assert sse.data == ["line1", "line2", "line3"]

    def test_from_content_with_bytes(self):
        """测试解析字节内容"""
        from app.utils.sserender import SSE

        content = b"data: test\r\n\r\n"
        sse = SSE.from_content(content)

        assert sse.data == ["test"]

    def test_from_content_strict_mode_valid(self):
        """测试严格模式有效内容"""
        from app.utils.sserender import SSE

        content = "data: test\r\n\r\n"
        sse = SSE.from_content(content, strict=True)

        assert sse.data == ["test"]

    def test_from_content_strict_mode_invalid_raises_error(self):
        """测试严格模式无效内容抛出错误"""
        from app.utils.sserender import SSE

        content = "data: test"  # Missing \r\n\r\n ending

        with pytest.raises(ValueError, match="not end with"):
            SSE.from_content(content, strict=True)

    def test_from_content_with_list(self):
        """测试从列表解析"""
        from app.utils.sserender import SSE

        content = ["id: msg123", "event: message", "data: test"]
        sse = SSE.from_content(content)

        assert sse.ID == "msg123"
        assert sse.event == "message"
        assert sse.data == ["test"]


class TestSSEDataStr:
    """测试 SSE data_str 方法"""

    def test_data_str_with_string_data(self):
        """测试字符串data获取"""
        from app.utils.sserender import SSE

        sse = SSE(data="test data")
        result = sse.data_str()

        assert result == "test data"

    def test_data_str_with_list_data_all(self):
        """测试列表data获取全部"""
        from app.utils.sserender import SSE

        sse = SSE(data=["line1", "line2", "line3"])
        result = sse.data_str()

        assert result == "line1line2line3"

    def test_data_str_with_list_data_slice_end(self):
        """测试列表data获取到指定位置"""
        from app.utils.sserender import SSE

        sse = SSE(data=["line1", "line2", "line3"])
        result = sse.data_str(end=2)

        assert result == "line1line2"

    def test_data_str_with_list_data_slice_start(self):
        """测试列表data从指定位置获取"""
        from app.utils.sserender import SSE

        sse = SSE(data=["line1", "line2", "line3"])
        result = sse.data_str(start=1)

        assert result == "line2line3"

    def test_data_str_with_list_data_slice_both(self):
        """测试列表data获取范围"""
        from app.utils.sserender import SSE

        sse = SSE(data=["line1", "line2", "line3", "line4"])
        result = sse.data_str(start=1, end=3)

        assert result == "line2line3"


class TestSSEInfo:
    """测试 SSE info 方法"""

    def test_info_with_valid_info_line(self):
        """测试带有效info行"""
        from app.utils.sserender import SSE

        info_dict = {"key": "value", "count": 42}
        sse = SSE(data=["line1", f"--info--{json.dumps(info_dict)}", "--end--"])

        result = sse.info()

        assert result == info_dict

    def test_info_without_info_line(self):
        """测试不带info行"""
        from app.utils.sserender import SSE

        sse = SSE(data=["line1", "line2"])

        result = sse.info()

        assert result == {}

    def test_info_with_insufficient_data(self):
        """测试data不足两行会抛出IndexError"""
        from app.utils.sserender import SSE

        sse = SSE(data=["line1"])

        # The actual code has a bug - it doesn't check length before accessing [-2]
        with pytest.raises(IndexError):
            sse.info()
