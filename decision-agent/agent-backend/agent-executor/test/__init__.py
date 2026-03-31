import json


class MockResponse:
    """
    mock requests返回值
    """

    def __init__(self, status_code, json_data=None, **kwargs):
        self.status_code = status_code
        self.json_data = json_data
        for key, value in kwargs.items():
            setattr(self, key, value)

    def json(self):
        return self.json_data

    def iter_lines(self):
        data = self.json_data
        if isinstance(data, str):
            lines = data.split("\n")
            data = []
            for line in lines:
                line = line.strip()
                if line.startswith("data:"):
                    line = line[6:]
                    data.append(line)
        elif isinstance(data, list):
            data = data
        else:
            data = []
        for chunk in data:
            chunk = "data: {}\n\n".format(chunk)
            yield chunk.encode()


class MockAsyncResponse:
    """
    mock aiohttp返回值
    """

    async def __aenter__(self):
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        pass

    def __init__(self, status, json_data=None, **kwargs):
        self.status = status
        self.json_data = json_data
        for key, value in kwargs.items():
            setattr(self, key, value)

    async def json(self):
        return self.json_data

    async def text(self) -> str:
        if isinstance(self.json_data, str):
            return self.json_data
        try:
            return json.dumps(self.json_data)
        except (TypeError, ValueError):
            return str(self.json_data)

    async def read(self):
        return (await self.text()).encode()

    def release(self):
        pass


class MockStreamContent:
    def __init__(self, data):
        if isinstance(data, str):
            lines = data.split("\n")
            self.data = []
            for line in lines:
                line = line.strip()
                if line.startswith("data:"):
                    line = line[6:]
                    self.data.append(line)
        elif isinstance(data, list):
            self.data = data
        else:
            self.data = []

    async def readline(self):
        if not self.data:
            return b""
        chunk = "data: {}".format(self.data.pop(0))
        return chunk.encode()

    async def iter_chunked(self, n: int):
        for chunk in self.data:
            chunk = "data: {}\n\n".format(chunk)
            yield chunk.encode()


class MockSession:
    """
    模拟 aiohttp.ClientSession 和它的用法
    """

    async def __aenter__(self):
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        pass

    def __init__(self, status=200, json_data=None, **kwargs):
        self.status = status
        self.json_data = json_data
        self.kwargs = kwargs
        self.side_effect = kwargs.get("side_effect", [])

    def request(self, *args, **kwargs):
        if self.side_effect:
            return self.side_effect.pop(0)
        return MockAsyncResponse(self.status, self.json_data, **self.kwargs)

    def get(self, *args, **kwargs):
        if self.side_effect:
            return self.side_effect.pop(0)
        return MockAsyncResponse(self.status, self.json_data, **self.kwargs)

    def post(self, *args, **kwargs):
        if self.side_effect:
            return self.side_effect.pop(0)
        return MockAsyncResponse(self.status, self.json_data, **self.kwargs)
