from typing import List, Optional, Dict, Any
from dataclasses import dataclass
import aiohttp
import asyncio


@dataclass
class RerankResponseUsage:
    prompt_tokens: int
    total_tokens: int


@dataclass
class RerankResponseResult:
    relevance_score: float
    index: int
    document: str = None


@dataclass
class RerankResponse:
    id: str

    usage: RerankResponseUsage
    results: List[RerankResponseResult]
    created: int
    object: str = "rerank"
    model: str = "reranker"

    def filter_by_threshold(self, threshold: float) -> "RerankResponse":
        """
        筛选relevance_score大于等于指定阈值的结果，返回一个新的RerankResponse对象

        Args:
            threshold: 阈值

        Returns:
            包含符合条件结果的新的RerankResponse对象
        """
        filtered_results = [
            result for result in self.results if result.relevance_score >= threshold
        ]
        return RerankResponse(
            id=self.id,
            object=self.object,
            model=self.model,
            usage=self.usage,
            results=filtered_results,
            created=self.created,
        )


@dataclass
class Config:
    rerank_url: str
    # 请求模型工厂，模型名称使用默认的 reranker
    rerank_model: str = "reranker"


class RerankModelClient:
    def __init__(self, config: Config):
        self.config = config

    def _prepare_request_body(self, query: str, documents: List[str]) -> dict:
        return {
            "model": self.config.rerank_model,
            "query": query,
            "documents": documents,
        }

    async def rerank(
        self, query: str, documents: List[str], context: Optional[Dict[str, Any]] = None
    ) -> RerankResponse:
        """
        Rerank documents based on a query.

        Args:
            query (str): Query to search for
            documents (List[str]): List of documents to rerank

        Returns:
            RerankResponse: Reranked documents
        """

        headers = {}
        if context:
            if context.get("user_id"):
                headers["x-account-id"] = context["user_id"]
            if context.get("visitor_type"):
                headers["x-account-type"] = context["visitor_type"]

        try:
            async with aiohttp.ClientSession() as session:
                context_manager = session.post(
                    self.config.rerank_url,
                    json=self._prepare_request_body(query, documents),
                    headers=headers,
                )
                if asyncio.iscoroutine(context_manager):
                    context_manager = await context_manager
                async with context_manager as response:
                    if response.status != 200:
                        raise Exception(f"Error reranking documents: {response.status}")
                    response_data = await response.json()

                    # Convert the response data to RerankResponse object
                    usage_data = response_data.get("usage", {})
                    usage = RerankResponseUsage(
                        prompt_tokens=usage_data.get("prompt_tokens", 0),
                        total_tokens=usage_data.get("total_tokens", 0),
                    )

                    results_data = response_data.get("results", [])
                    results = [
                        RerankResponseResult(
                            relevance_score=result.get("relevance_score") or 0.0,
                            index=result.get("index", 0),
                            document=result.get("document", ""),
                        )
                        for result in results_data
                    ]

                    return RerankResponse(
                        id=response_data.get("id", ""),
                        object=response_data.get("object", "rerank"),
                        model=response_data.get("model", "reranker"),
                        usage=usage,
                        results=results,
                        created=response_data.get("created", 0),
                    )
        except Exception as e:
            raise Exception(f"Error reranking documents: {e}")

    async def rerank_with_threshold(
        self,
        query: str,
        documents: List[str],
        threshold: float = 0.1,
        context: Optional[Dict[str, Any]] = None,
    ) -> RerankResponse:
        """
        筛选relevance_score大于等于指定阈值的结果
        Args:
            query (str): Query to search for
            documents (List[str]): List of documents to rerank
            threshold: 阈值， 返回大于等于阈值的结果
            Returns:
                包含符合条件结果的新的RerankResponse对象
        """
        rerank_response = await self.rerank(query, documents, context=context)
        return rerank_response.filter_by_threshold(threshold)
