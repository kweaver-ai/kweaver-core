import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from src.adaptee.mf_model_factory.rerank_model_client import (
    RerankResponseUsage,
    RerankResponseResult,
    RerankResponse,
    Config,
    RerankModelClient,
)


class TestRerankResponseUsage:
    def test_creation(self):
        """Test creating RerankResponseUsage"""
        usage = RerankResponseUsage(prompt_tokens=10, total_tokens=20)
        assert usage.prompt_tokens == 10
        assert usage.total_tokens == 20

    def test_creation_with_zero(self):
        """Test creating RerankResponseUsage with zero values"""
        usage = RerankResponseUsage(prompt_tokens=0, total_tokens=0)
        assert usage.prompt_tokens == 0
        assert usage.total_tokens == 0

    def test_creation_with_large_values(self):
        """Test creating RerankResponseUsage with large values"""
        usage = RerankResponseUsage(prompt_tokens=10000, total_tokens=20000)
        assert usage.prompt_tokens == 10000
        assert usage.total_tokens == 20000


class TestRerankResponseResult:
    def test_creation(self):
        """Test creating RerankResponseResult"""
        result = RerankResponseResult(
            relevance_score=0.9, index=0, document="Test document"
        )
        assert result.relevance_score == 0.9
        assert result.index == 0
        assert result.document == "Test document"

    def test_creation_without_document(self):
        """Test creating RerankResponseResult without document"""
        result = RerankResponseResult(relevance_score=0.8, index=1)
        assert result.document is None

    def test_creation_with_zero_score(self):
        """Test creating RerankResponseResult with zero score"""
        result = RerankResponseResult(relevance_score=0.0, index=0, document="Test")
        assert result.relevance_score == 0.0

    def test_creation_with_negative_score(self):
        """Test creating RerankResponseResult with negative score"""
        result = RerankResponseResult(relevance_score=-0.5, index=0, document="Test")
        assert result.relevance_score == -0.5

    def test_creation_with_score_greater_than_one(self):
        """Test creating RerankResponseResult with score > 1"""
        result = RerankResponseResult(relevance_score=1.5, index=0, document="Test")
        assert result.relevance_score == 1.5


class TestRerankResponse:
    def test_creation(self):
        """Test creating RerankResponse"""
        usage = RerankResponseUsage(prompt_tokens=10, total_tokens=20)
        results = [
            RerankResponseResult(relevance_score=0.9, index=0, document="Doc1"),
            RerankResponseResult(relevance_score=0.8, index=1, document="Doc2"),
        ]

        response = RerankResponse(
            id="test_id", usage=usage, results=results, created=1234567890
        )

        assert response.id == "test_id"
        assert response.object == "rerank"
        assert response.model == "reranker"
        assert response.usage == usage
        assert len(response.results) == 2
        assert response.created == 1234567890

    def test_creation_with_custom_object_and_model(self):
        """Test creating RerankResponse with custom object and model"""
        usage = RerankResponseUsage(prompt_tokens=5, total_tokens=10)
        results = []

        response = RerankResponse(
            id="custom_id",
            object="custom_object",
            model="custom_model",
            usage=usage,
            results=results,
            created=1234567890,
        )

        assert response.object == "custom_object"
        assert response.model == "custom_model"

    def test_filter_by_threshold(self):
        """Test filtering results by threshold"""
        usage = RerankResponseUsage(prompt_tokens=10, total_tokens=20)
        results = [
            RerankResponseResult(relevance_score=0.9, index=0, document="Doc1"),
            RerankResponseResult(relevance_score=0.5, index=1, document="Doc2"),
            RerankResponseResult(relevance_score=0.7, index=2, document="Doc3"),
            RerankResponseResult(relevance_score=0.6, index=3, document="Doc4"),
        ]

        response = RerankResponse(
            id="test_id", usage=usage, results=results, created=1234567890
        )

        filtered = response.filter_by_threshold(0.7)

        assert len(filtered.results) == 2
        assert all(r.relevance_score >= 0.7 for r in filtered.results)
        assert filtered.id == "test_id"
        assert filtered.created == 1234567890

    def test_filter_by_threshold_no_results(self):
        """Test filtering with threshold higher than all scores"""
        usage = RerankResponseUsage(prompt_tokens=10, total_tokens=20)
        results = [
            RerankResponseResult(relevance_score=0.5, index=0, document="Doc1"),
            RerankResponseResult(relevance_score=0.6, index=1, document="Doc2"),
        ]

        response = RerankResponse(
            id="test_id", usage=usage, results=results, created=1234567890
        )

        filtered = response.filter_by_threshold(0.9)

        assert len(filtered.results) == 0

    def test_filter_by_threshold_all_results(self):
        """Test filtering with threshold lower than all scores"""
        usage = RerankResponseUsage(prompt_tokens=10, total_tokens=20)
        results = [
            RerankResponseResult(relevance_score=0.9, index=0, document="Doc1"),
            RerankResponseResult(relevance_score=0.8, index=1, document="Doc2"),
        ]

        response = RerankResponse(
            id="test_id", usage=usage, results=results, created=1234567890
        )

        filtered = response.filter_by_threshold(0.5)

        assert len(filtered.results) == 2

    def test_filter_by_threshold_exact_match(self):
        """Test filtering with exact threshold match"""
        usage = RerankResponseUsage(prompt_tokens=10, total_tokens=20)
        results = [
            RerankResponseResult(relevance_score=0.7, index=0, document="Doc1"),
            RerankResponseResult(relevance_score=0.6, index=1, document="Doc2"),
        ]

        response = RerankResponse(
            id="test_id", usage=usage, results=results, created=1234567890
        )

        filtered = response.filter_by_threshold(0.7)

        assert len(filtered.results) == 1
        assert filtered.results[0].relevance_score == 0.7

    def test_filter_returns_new_instance(self):
        """Test that filter returns a new instance"""
        usage = RerankResponseUsage(prompt_tokens=10, total_tokens=20)
        results = [RerankResponseResult(relevance_score=0.9, index=0, document="Doc1")]

        response = RerankResponse(
            id="test_id", usage=usage, results=results, created=1234567890
        )

        filtered = response.filter_by_threshold(0.5)

        assert response is not filtered
        assert len(response.results) == 1
        assert len(filtered.results) == 1


class TestConfig:
    def test_creation(self):
        """Test creating Config"""
        config = Config(rerank_url="http://test.com/rerank")
        assert config.rerank_url == "http://test.com/rerank"
        assert config.rerank_model == "reranker"

    def test_creation_with_custom_model(self):
        """Test creating Config with custom model"""
        config = Config(
            rerank_url="http://test.com/rerank", rerank_model="custom_model"
        )
        assert config.rerank_url == "http://test.com/rerank"
        assert config.rerank_model == "custom_model"

    def test_creation_with_empty_model(self):
        """Test creating Config with empty model"""
        config = Config(rerank_url="http://test.com/rerank", rerank_model="")
        assert config.rerank_model == ""


class TestRerankModelClient:
    def test_init(self):
        """Test client initialization"""
        config = Config(rerank_url="http://test.com/rerank")
        client = RerankModelClient(config)

        assert client.config is config

    def test_prepare_request_body(self):
        """Test preparing request body"""
        config = Config(rerank_url="http://test.com/rerank")
        client = RerankModelClient(config)

        query = "test query"
        documents = ["doc1", "doc2", "doc3"]
        body = client._prepare_request_body(query, documents)

        assert body == {"model": "reranker", "query": query, "documents": documents}

    def test_prepare_request_body_with_custom_model(self):
        """Test preparing request body with custom model"""
        config = Config(
            rerank_url="http://test.com/rerank", rerank_model="custom_model"
        )
        client = RerankModelClient(config)

        query = "test query"
        documents = ["doc1"]
        body = client._prepare_request_body(query, documents)

        assert body["model"] == "custom_model"

    def test_prepare_request_body_empty_documents(self):
        """Test preparing request body with empty documents list"""
        config = Config(rerank_url="http://test.com/rerank")
        client = RerankModelClient(config)

        body = client._prepare_request_body("query", [])

        assert body["documents"] == []

    def test_prepare_request_body_special_characters(self):
        """Test preparing request body with special characters"""
        config = Config(rerank_url="http://test.com/rerank")
        client = RerankModelClient(config)

        query = "Query with special chars: <>&\"'"
        documents = ["Document with <html>"]
        body = client._prepare_request_body(query, documents)

        assert body["query"] == query
        assert body["documents"] == documents
