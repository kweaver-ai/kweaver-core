import os
import sys
import pytest
from unittest.mock import patch, AsyncMock, MagicMock

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from src.adaptee.mf_model_factory.rerank_model_client import (
    RerankModelClient,
    Config,
    RerankResponse,
    RerankResponseUsage,
    RerankResponseResult,
)
import asyncio


def test_init_and_prepare_request_body():
    config = Config(rerank_url="http://test.com/rerank")
    client = RerankModelClient(config)
    assert client.config is config

    query = "test query"
    documents = ["doc1", "doc2", "doc3"]
    body = client._prepare_request_body(query, documents)
    assert body == {
        "model": "reranker",
        "query": query,
        "documents": documents,
    }


def test_rerank_response_filter_by_threshold():
    usage = RerankResponseUsage(prompt_tokens=10, total_tokens=20)
    results = [
        RerankResponseResult(relevance_score=0.9, index=0, document="doc1"),
        RerankResponseResult(relevance_score=0.5, index=1, document="doc2"),
        RerankResponseResult(relevance_score=0.8, index=2, document="doc3"),
    ]
    response = RerankResponse(
        id="test_id",
        object="rerank",
        model="reranker",
        usage=usage,
        results=results,
        created=1234567890,
    )

    filtered = response.filter_by_threshold(0.7)
    assert filtered.id == "test_id"
    assert filtered.object == "rerank"
    assert filtered.model == "reranker"
    assert filtered.usage is usage
    assert filtered.created == 1234567890
    assert [r.relevance_score for r in filtered.results] == [0.9, 0.8]


@pytest.mark.asyncio
async def test_rerank_success():
    config = Config(rerank_url="http://test.com/rerank")
    client = RerankModelClient(config)

    payload = {
        "id": "test_response_id",
        "object": "rerank",
        "model": "reranker",
        "usage": {"prompt_tokens": 10, "total_tokens": 20},
        "results": [
            {"relevance_score": 0.9, "index": 0, "document": "doc1"},
            {"relevance_score": 0.8, "index": 1, "document": "doc2"},
        ],
        "created": 1234567890,
    }

    mock_response = AsyncMock()
    mock_response.status = 200
    mock_response.json = AsyncMock(return_value=payload)

    # Use a MagicMock as the context manager to avoid being detected as coroutine
    mock_post_cm = MagicMock()
    mock_post_cm.__aenter__ = AsyncMock(return_value=mock_response)
    mock_post_cm.__aexit__ = AsyncMock(return_value=None)

    mock_session = AsyncMock()
    mock_session.post.return_value = mock_post_cm

    with patch("aiohttp.ClientSession") as mock_session_cls:
        mock_session_cls.return_value.__aenter__.return_value = mock_session

        result = await client.rerank("test query", ["doc1", "doc2"])

        assert isinstance(result, RerankResponse)
        assert result.id == "test_response_id"
        assert result.object == "rerank"
        assert result.model == "reranker"
        assert result.usage.prompt_tokens == 10
        assert result.usage.total_tokens == 20
        assert [r.relevance_score for r in result.results] == [0.9, 0.8]

        mock_session.post.assert_called_once_with(
            config.rerank_url,
            json={
                "model": "reranker",
                "query": "test query",
                "documents": ["doc1", "doc2"],
            },
        )


@pytest.mark.asyncio
async def test_rerank_http_error():
    config = Config(rerank_url="http://test.com/rerank")
    client = RerankModelClient(config)

    mock_response = AsyncMock()
    mock_response.status = 500

    mock_post_cm = MagicMock()
    mock_post_cm.__aenter__ = AsyncMock(return_value=mock_response)
    mock_post_cm.__aexit__ = AsyncMock(return_value=None)

    mock_session = AsyncMock()
    mock_session.post.return_value = mock_post_cm

    with patch("aiohttp.ClientSession") as mock_session_cls:
        mock_session_cls.return_value.__aenter__.return_value = mock_session

        with pytest.raises(Exception) as exc:
            await client.rerank("q", ["d1"])
        # The raised exception message should include the inner error
        assert "Error reranking documents" in str(exc.value)
        assert "500" in str(exc.value)


@pytest.mark.asyncio
async def test_rerank_with_threshold():
    config = Config(rerank_url="http://test.com/rerank")
    client = RerankModelClient(config)

    usage = RerankResponseUsage(prompt_tokens=10, total_tokens=20)
    results = [
        RerankResponseResult(relevance_score=0.9, index=0, document="doc1"),
        RerankResponseResult(relevance_score=0.5, index=1, document="doc2"),
        RerankResponseResult(relevance_score=0.8, index=2, document="doc3"),
    ]
    mock_response = RerankResponse(
        id="test_id",
        object="rerank",
        model="reranker",
        usage=usage,
        results=results,
        created=1234567890,
    )

    with patch.object(
        client, "rerank", new=AsyncMock(return_value=mock_response)
    ) as mock_rerank:
        filtered = await client.rerank_with_threshold("q", ["d1", "d2"], threshold=0.7)
        assert isinstance(filtered, RerankResponse)
        assert [r.relevance_score for r in filtered.results] == [0.9, 0.8]
        mock_rerank.assert_awaited_once()


if __name__ == "__main__":
    from src.config import rerank_config

    client = RerankModelClient(rerank_config)
    result = asyncio.run(
        client.rerank_with_threshold(
            "我叫什么",
            ["周敏时国家二级心里咨询师", "我的姓名是郭晨光", "无关"],
            threshold=0.1,
        )
    )

    print(result)
