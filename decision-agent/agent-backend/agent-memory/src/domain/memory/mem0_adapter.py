from mem0.memory.main import AsyncMemory
from src.config import memory_config, rerank_config
from typing import Any, Dict, List
from src.utils.logger import logger
from src.adaptee.mf_model_factory.rerank_model_client import RerankModelClient


class Mem0MemoryAdapter:
    """基于mem0的Memory能力的适配器"""

    _instance = None
    _initialized = False
    rerank_client: RerankModelClient = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance

    def __init__(self):
        if not self._initialized:
            self.memory = None
            self._initialized = True
            self.rerank_client = None

    @classmethod
    async def create(cls):
        """创建并初始化适配器实例"""
        instance = cls()
        if instance.memory is None:
            logger.info("Initializing Async Memory...")
            instance.memory = AsyncMemory(config=memory_config)
            logger.info("Async Memory initialized.")

        if instance.rerank_client is None:
            logger.info("Initializing Rerank Client")

            instance.rerank_client = RerankModelClient(config=rerank_config)
            logger.info("Rerank Client initialized.")
        return instance

    async def add(self, messages, **kwargs):
        if self.memory is None:
            await self.create()

        # Extract context from kwargs
        context = kwargs.get("context")

        # Remove context from kwargs to avoid passing it to mem0
        kwargs.pop("context", None)

        return await self.memory.add(messages, context=context, **kwargs)

    async def search(self, query: str, **kwargs):
        if self.memory is None:
            await self.create()

        # Use provided threshold (fallback 0.5)
        rerank_threshold = kwargs.get("rerank_threshold", 0.0)
        context = kwargs.get("context", {})

        # Remove context from kwargs to avoid passing it to mem0
        kwargs.pop("rerank_threshold", None)
        kwargs.pop("context", None)

        search_rt = await self.memory.search(query, context=context, **kwargs)
        if not search_rt:
            return search_rt

        try:
            # Extract results list from mem0 search output
            original_results: List[Dict[str, Any]] = (
                search_rt.get("results", [])
                if isinstance(search_rt, dict)
                else list(search_rt)
            )
            if not original_results:
                return search_rt

            # Build documents from memory field
            documents: List[str] = [item.get("memory", "") for item in original_results]
            logger.info(f"Reranking memory with threashold:{rerank_threshold}")
            # Call rerank service and filter by threshold
            reranked = await self.rerank_client.rerank_with_threshold(
                query, documents, threshold=rerank_threshold, context=context
            )
            logger.info(f"Reranked {len(reranked.results)} memories ")
            # Map indices from rerank back to original results and order by relevance desc
            sorted_indices = [
                r.index
                for r in sorted(
                    reranked.results, key=lambda x: x.relevance_score, reverse=True
                )
            ]
            # Create a mapping from index to relevance_score for easy lookup
            index_to_score = {r.index: r.relevance_score for r in reranked.results}

            # Deduplicate and keep valid indices only, adding relevance_score
            seen = set()
            filtered_sorted = []
            for idx in sorted_indices:
                if 0 <= idx < len(original_results) and idx not in seen:
                    result_item = original_results[idx].copy()
                    result_item["rerank_score"] = index_to_score[idx]
                    filtered_sorted.append(result_item)
                    seen.add(idx)

            # Return in the same structure as input
            if isinstance(search_rt, dict):
                search_rt["results"] = filtered_sorted
                return search_rt
            else:
                return filtered_sorted
        except Exception as e:
            logger.error(f"Error during reranking: {e}")
            return search_rt

    async def get(self, memory_id: str):
        if self.memory is None:
            await self.create()
        return await self.memory.get(memory_id)

    async def get_all(self, **kwargs):
        if self.memory is None:
            await self.create()
        return await self.memory.get_all(**kwargs)

    async def update(self, memory_id: str, data: Any):
        if self.memory is None:
            await self.create()
        return await self.memory.update(memory_id, data)

    async def delete(self, memory_id: str):
        if self.memory is None:
            await self.create()
        return await self.memory.delete(memory_id)

    async def history(self, memory_id: str):
        if self.memory is None:
            await self.create()
        return await self.memory.history(memory_id)
