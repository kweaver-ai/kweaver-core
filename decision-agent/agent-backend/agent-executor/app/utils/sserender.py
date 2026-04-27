import io
import json
import re
from typing import Optional, Union, List


class SSE:
    """SSE解析器"""

    _DEFAULT_SEPARATOR: str = "\r\n"
    _LINE_SEP_EXPR = re.compile(r"\r\n|\r|\n")
    ID: Optional[str]
    event: Optional[str]
    data: Optional[Union[str, List[str]]]
    retry: Optional[int]
    comment: Optional[Union[str, List[str]]]

    __slots__ = ["ID", "event", "data", "retry", "comment"]

    def __init__(
        self,
        ID: Optional[str] = None,
        event: Optional[str] = None,
        data: Optional[Union[str, List[str]]] = None,
        comment: Optional[Union[str, List[str]]] = None,
        retry: Optional[int] = None,
    ) -> None:
        """构造sse消息.

        除了retry外至少要有1个参数.

        Args:
            ID (Optional[str], optional): 消息id. Defaults to None.
            event (Optional[str], optional): 消息事件. Defaults to None.
            data (Optional[str], optional): 消息数据. Defaults to None.
            comment (Optional[str], optional): 消息注释. Defaults to None.
            retry (Optional[int], optional): 重连间隔时间,单位ms. Defaults to None.

        Raises:
            TypeError: [description]
            ValueError: [description]
        """
        if any([ID, event, data, comment]):
            self.ID = ID
            self.event = event
            self.data = data
            self.retry = retry
            self.comment = comment
            if self.retry and not isinstance(self.retry, int):
                raise TypeError("retry argument must be int")
        else:
            raise ValueError("at least one argument")

    def render(self, *, with_encode: bool = False) -> Union[str, bytes]:
        """渲染sse消息对象

        Args:
            with_encode (bool, optional): 是否编码. Defaults to False.

        Returns:
            Union[str, bytes]: with_encode为True则返回为bytes否则为str
        """
        buffer = io.StringIO()
        if self.comment is not None:
            if isinstance(self.comment, list):
                self.comment = "\n".join(self.comment)
            for chunk in self._LINE_SEP_EXPR.split(self.comment):
                buffer.write(": {}".format(chunk))
                buffer.write(self._DEFAULT_SEPARATOR)
        if self.ID is not None:
            buffer.write(self._LINE_SEP_EXPR.sub("", "id: {}".format(self.ID)))
            buffer.write(self._DEFAULT_SEPARATOR)

        if self.event is not None:
            buffer.write(self._LINE_SEP_EXPR.sub("", "event: {}".format(self.event)))
            buffer.write(self._DEFAULT_SEPARATOR)
        if self.data is not None:
            if isinstance(self.data, list):
                self.data = "\n".join(self.data)
            for chunk in self._LINE_SEP_EXPR.split(self.data):
                buffer.write("data: {}".format(chunk))
                buffer.write(self._DEFAULT_SEPARATOR)

        if self.retry is not None:
            buffer.write("retry: {}".format(self.retry))
            buffer.write(self._DEFAULT_SEPARATOR)

        buffer.write(self._DEFAULT_SEPARATOR)
        result = buffer.getvalue()
        if with_encode:
            return result.encode("utf-8")
        return result

    @classmethod
    def from_content(
        clz,
        content: Union[str, bytes, list],
        *,
        strict: bool = False,
        spearator: str = None,
    ) -> "SSE":
        """从字符串或者字节流中解析出sse消息.

        Args:
            content (Union[str, bytes]): 待解析的内容
            strict (bool, optional): 严格模式,校验结尾是否为`\r\n\r\n`. Defaults to False.

        Raises:
            ValueError: not end with `\r\n\r\n`

        Returns:
            [SSE]: sse消息对象
        """
        if spearator is None:
            spearator = clz._DEFAULT_SEPARATOR
        if isinstance(content, list):
            lines = content
        else:
            if isinstance(content, bytes):
                content = content.decode("utf-8")
            if strict:
                end = spearator + spearator
                if not content.endswith(end):
                    raise ValueError(f"not end with {end}")
            lines = content.strip().split(spearator)
        ID = ""
        event = ""
        data = []
        retry = 0
        comment = []

        for line in lines:
            if line.startswith(": "):
                c = line[2:]
                comment.append(c)
            if line.startswith("id: "):
                ID = line[4:]
            if line.startswith("event: "):
                event = line[7:]
            if line.startswith("data: "):
                c = line[6:]
                data.append(c)
            if line.startswith("retry: "):
                retry = int(line[7:].strip())
        args = {}
        if ID:
            args.update({"ID": ID})
        if event:
            args.update({"event": event})
        if data:
            args.update({"data": data})
        if retry:
            args.update({"retry": retry})
        if comment:
            args.update({"comment": comment})
        return clz(**args)

    def data_str(self, start=None, end=None):
        """
        获取data数据
        tail_trim=2: 最后两条信息是 --info-- 和 --end--, 需要排除掉
        """
        if isinstance(self.data, list):
            if start is None and end is None:
                data = self.data
            elif start is None and end is not None:
                data = self.data[:end]
            elif start is not None and end is None:
                data = self.data[start:]
            elif start is not None and end is not None:
                data = self.data[start:end]
            return "".join(data)
        return self.data

    def info(self) -> dict:
        if isinstance(self.data, list) and self.data[-2].startswith("--info--"):
            return json.loads(self.data[-2][8:])
        return {}
