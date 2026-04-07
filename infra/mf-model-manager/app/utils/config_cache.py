import json
import os
import threading
import asyncio
import queue
from typing import Dict, Any
from datetime import datetime
from time import monotonic
from app.dao.model_quota_dao import model_quota_dao
from app.mydb.ConnectUtil import redis_util, get_redis_util
from app.logs.stand_log import StandLogger


class ModelConfigNode:
    """模型配置数据节点类，支持直接属性访问"""

    # 定义类属性以支持IDE自动提示
    model_id: str
    billing_type: str
    input_tokens: int
    output_tokens: int
    currency_type: str
    referprice_in: float
    referprice_out: float
    num_type: list
    price_type: list

    def __init__(self, data: Dict[str, Any]):
        self.model_id = str(data.get("f_model_id"))
        self.billing_type = int(data.get("f_billing_type", 1))
        self.input_tokens = int(float(data.get("f_input_tokens", -1)))
        self.output_tokens = int(float(data.get("f_output_tokens", -1)))
        self.currency_type = int(data.get("f_currency_type", 0))
        self.referprice_in = float(data.get("f_referprice_in", -1))
        self.referprice_out = float(data.get("f_referprice_out", -1))

        raw_num_type = data.get("f_num_type", [1, 1])
        if isinstance(raw_num_type, str):
            try:
                raw_num_type = json.loads(raw_num_type)
            except Exception:
                raw_num_type = [1, 1]
        self.num_type = list(raw_num_type)

        raw_price_type = data.get("f_price_type", ["thousand", "thousand"])
        if isinstance(raw_price_type, str):
            try:
                raw_price_type = json.loads(raw_price_type)
            except Exception:
                raw_price_type = ["thousand", "thousand"]
        self.price_type = list(raw_price_type)

    def to_dict(self) -> Dict[str, Any]:
        return {
            "f_model_id": self.model_id,
            "f_billing_type": self.billing_type,
            "f_input_tokens": self.input_tokens,
            "f_output_tokens": self.output_tokens,
            "f_currency_type": self.currency_type,
            "f_referprice_in": self.referprice_in,
            "f_referprice_out": self.referprice_out,
            "f_num_type": self.num_type,
            "f_price_type": self.price_type,
        }


class QuotaConfigCacheTree:
    """配置缓存树形结构类"""

    def __init__(self):
        # 使用树形结构存储数据，第一层为model_id，第二层为具体配置属性
        self._root: Dict[str, ModelConfigNode] = {}
        self._lock = threading.RLock()
        self._snapshot_path = os.path.join(os.path.dirname(__file__), "quota_config_cache.snapshot.json")
        self._snapshot_mtime = 0.0
        self._boot_source = "unknown"  # db | snapshot | empty
        self._load_from_snapshot_or_db()
        StandLogger.info_log("加载大模型配额相关配置成功")
        # Redis 多副本同步设置（基于全局单例）
        self._redis_channel = "quota_config_change"
        self._redis_hash_key = "quota_config_hash"
        self._redis_cmd_queue: "queue.Queue" = queue.Queue()
        try:
            self._redis_worker_thread = threading.Thread(target=self._run_redis_worker_and_listener, daemon=True)
            self._redis_worker_thread.start()
        except Exception as e:
            StandLogger.error(f"启动Redis工作线程失败: {e}")

    def _load_data(self):
        """从数据库加载所有配置数据到树形结构"""
        try:
            model_quota_config = model_quota_dao.get_all_model_quota_config()
            loaded = False
            for line in model_quota_config:
                model_id = str(line["f_model_id"])
                billing_type = line["f_billing_type"]
                if not billing_type:
                    line = {"f_model_id": model_id}
                self._root[model_id] = ModelConfigNode(line)
                loaded = True
            if loaded:
                self._save_snapshot()
                self._boot_source = "db"
                StandLogger.info_log(f"配额配置从DB加载，模型ID={self.list_all_model_ids()}")
            else:
                self._boot_source = "empty"
                StandLogger.info_log("DB未返回任何配额配置")
        except Exception as e:
            self._boot_source = "empty"
            StandLogger.info(f"加载配置数据时出错: {e}")

    def _load_from_snapshot_or_db(self) -> None:
        # 优先尝试从DB加载；如果DB为空/失败，则尝试快照；最后保持空
        with self._lock:
            self._root.clear()
            try:
                self._load_data()
            except Exception:
                pass
            if self._boot_source != "db":
                # DB为空或失败，尝试快照
                if os.path.exists(self._snapshot_path):
                    try:
                        self._load_snapshot()
                        self._boot_source = "snapshot"
                        StandLogger.info_log(f"配额配置从快照加载，模型ID={self.list_all_model_ids()}")
                        return
                    except Exception as e:
                        StandLogger.error(f"加载快照失败，保持空: {e}")
                # 最终为空
                self._root = {}

    def _load_snapshot(self) -> None:
        with self._lock:
            with open(self._snapshot_path, "r", encoding="utf-8") as f:
                data = json.load(f)
            new_root: Dict[str, ModelConfigNode] = {}
            for model_id, cfg in data.items():
                new_root[str(model_id)] = ModelConfigNode(cfg)
            self._root = new_root
            try:
                self._snapshot_mtime = os.path.getmtime(self._snapshot_path)
            except Exception:
                self._snapshot_mtime = 0.0

    def _save_snapshot(self) -> None:
        with self._lock:
            directory = os.path.dirname(self._snapshot_path)
            if directory and not os.path.exists(directory):
                os.makedirs(directory, exist_ok=True)
            tmp_path = self._snapshot_path
            for mid, node in self._root.items():
                StandLogger.info(f"mid={mid},node={node}")
            serializable = {mid: node.to_dict() for mid, node in self._root.items()}
            with open(tmp_path, "w", encoding="utf-8") as f:
                json.dump(serializable, f, ensure_ascii=False)
                f.flush()
                os.fsync(f.fileno())
            os.replace(tmp_path, self._snapshot_path)
            try:
                self._snapshot_mtime = os.path.getmtime(self._snapshot_path)
            except Exception:
                self._snapshot_mtime = 0.0

    def _refresh_from_snapshot_if_modified(self) -> None:
        try:
            if not os.path.exists(self._snapshot_path):
                return
            current_mtime = os.path.getmtime(self._snapshot_path)
            if current_mtime > self._snapshot_mtime:
                self._load_snapshot()
        except Exception as e:
            StandLogger.error(f"检查/加载快照失败: {e}")

    # ---------------- Redis 同步（基于全局单例）：发布/订阅与全量同步 ----------------
    def _enqueue_cmd(self, cmd: Dict[str, Any]) -> None:
        try:
            self._redis_cmd_queue.put_nowait(cmd)
        except Exception:
            pass

    def _run_redis_worker_and_listener(self) -> None:
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)

        async def startup_and_run():
            # 初始化全局redis单例
            try:
                global redis_util
                if redis_util is None:
                    redis_util = await get_redis_util()
            except Exception as e:
                StandLogger.error(f"初始化Redis单例失败: {e}")
                return

            # 启动前根据来源决定同步方向
            try:
                if self._boot_source == "db":
                    # 以DB为准，广播全量覆盖到Redis
                    self._enqueue_cmd({"op": "full_reload"})
                    StandLogger.info_log("以DB为准，已触发全量覆盖写入Redis")
                else:
                    # 本地非DB来源，尝试从Redis回填本地
                    await self._async_load_all_from_redis_hash()
            except Exception as e:
                StandLogger.error(f"启动前同步阶段异常: {e}")

            # 并行启动：订阅循环 + 命令处理循环
            await asyncio.gather(
                self._async_subscribe_loop(),
                self._async_cmd_loop()
            )

        try:
            loop.run_until_complete(startup_and_run())
        finally:
            try:
                loop.stop()
            except Exception:
                pass

    async def _async_load_all_from_redis_hash(self) -> None:
        try:
            global redis_util
            all_map = await redis_util.read_conn.hgetall(self._redis_hash_key)
            if not all_map:
                return
            new_root: Dict[str, ModelConfigNode] = {}
            for k, v in all_map.items():
                model_id = k.decode() if isinstance(k, (bytes, bytearray)) else str(k)
                raw = v.decode() if isinstance(v, (bytes, bytearray)) else v
                try:
                    cfg = json.loads(raw)
                    new_root[model_id] = ModelConfigNode(cfg)
                except Exception:
                    continue
            if new_root:
                with self._lock:
                    self._root = new_root
                    self._save_snapshot()
                StandLogger.info_log(f"已从Redis Hash全量同步{len(new_root)}条配置")
        except Exception as e:
            StandLogger.error(f"异步从Redis Hash全量加载失败: {e}")

    async def _async_publish(self, payload: Dict[str, Any]) -> None:
        try:
            global redis_util
            await redis_util.write_conn.publish(self._redis_channel, json.dumps(payload, ensure_ascii=False))
        except Exception as e:
            StandLogger.error(f"异步发布Redis变更事件失败: {e}")

    async def _async_cmd_loop(self) -> None:
        loop = asyncio.get_running_loop()
        while True:
            try:
                cmd = await loop.run_in_executor(None, self._redis_cmd_queue.get)
                op = cmd.get("op")
                global redis_util
                if op == "upsert":
                    model_id = cmd.get("model_id")
                    cfg = cmd.get("config")
                    await redis_util.write_conn.hset(self._redis_hash_key, model_id,
                                                     json.dumps(cfg, ensure_ascii=False))
                    await self._async_publish({"op": "upsert", "config": cfg})
                elif op == "delete":
                    model_id = cmd.get("model_id")
                    await redis_util.write_conn.hdel(self._redis_hash_key, model_id)
                    await self._async_publish({"op": "delete", "model_id": model_id})
                elif op == "full_reload":
                    # 覆盖写入Hash
                    pipe = redis_util.write_conn.pipeline(transaction=True)
                    await pipe.delete(self._redis_hash_key)
                    for mid, node in self._root.items():
                        await pipe.hset(self._redis_hash_key, mid, json.dumps(node.to_dict(), ensure_ascii=False))
                    await pipe.execute()
                    await self._async_publish({"op": "full_reload"})
            except Exception as e:
                StandLogger.error(f"Redis命令处理异常: {e}")

    async def _async_subscribe_loop(self) -> None:
        backoff = 0.5
        while True:
            try:
                global redis_util
                pubsub = redis_util.read_conn.pubsub()
                await pubsub.subscribe(self._redis_channel)
                while True:
                    msg = await pubsub.get_message(ignore_subscribe_messages=True, timeout=1.0)
                    if not msg:
                        await asyncio.sleep(0.05)
                        continue
                    data = msg.get("data")
                    if isinstance(data, (bytes, bytearray)):
                        data = data.decode()
                    try:
                        payload = json.loads(data)
                    except Exception:
                        continue
                    op = payload.get("op")
                    if op == "upsert":
                        cfg = payload.get("config") or {}
                        model_id = str(cfg.get("f_model_id") or payload.get("model_id"))
                        if model_id:
                            with self._lock:
                                self._root[model_id] = ModelConfigNode(cfg)
                                self._save_snapshot()
                    elif op == "delete":
                        model_id = str(payload.get("model_id", ""))
                        if model_id:
                            with self._lock:
                                if model_id in self._root:
                                    del self._root[model_id]
                                    self._save_snapshot()
                    elif op == "full_reload":
                        await self._async_load_all_from_redis_hash()
                backoff = 0.5
            except Exception as e:
                StandLogger.error(f"Redis订阅循环异常: {e}")
                try:
                    await asyncio.sleep(backoff)
                    backoff = min(backoff * 2, 10)
                except Exception:
                    pass

    def __getattr__(self, model_id: str) -> ModelConfigNode:
        """支持直接通过属性访问获取模型配置数据"""
        self._refresh_from_snapshot_if_modified()
        if model_id in self._root:
            return self._root[model_id]
        raise AttributeError(f"模型配置 '{model_id}' 不存在")

    def __getitem__(self, model_id: str) -> ModelConfigNode:
        """支持通过索引访问获取模型配置数据"""
        self._refresh_from_snapshot_if_modified()
        if model_id in self._root:
            return self._root[model_id]
        raise KeyError(f"模型配置 '{model_id}' 不存在")

    def add(self, data: Dict[str, Any]) -> None:
        """添加新的配置数据到树形结构"""
        with self._lock:
            model_id = str(data["f_model_id"])
            self._root[model_id] = ModelConfigNode(data)
            self._save_snapshot()
            # 异步同步到Redis并发布变更（通过队列交给工作线程处理）
            self._enqueue_cmd({"op": "upsert", "model_id": model_id, "config": self._root[model_id].to_dict()})
        return True

    def update(self, data: Dict[str, Any]) -> bool:
        model_id = str(data["f_model_id"])
        """更新指定model_id的配置数据"""
        with self._lock:
            self._root[model_id] = ModelConfigNode(data)
            self._save_snapshot()
            # 异步同步到Redis并发布变更
            self._enqueue_cmd({"op": "upsert", "model_id": model_id, "config": self._root[model_id].to_dict()})
        return True

    def delete(self, model_id: str) -> bool:
        """从树形结构中删除指定model_id的配置数据"""
        with self._lock:
            if model_id in self._root:
                del self._root[model_id]
                self._save_snapshot()
                # 异步同步删除
                self._enqueue_cmd({"op": "delete", "model_id": model_id})
                return True
        return False

    def delete_batch(self, model_ids: []) -> bool:
        """从树形结构中批量删除指定model_id的配置数据"""
        with self._lock:
            for model_id in model_ids:
                if model_id in self._root:
                    del self._root[model_id]
            self._save_snapshot()
            # 异步广播全量刷新
            self._enqueue_cmd({"op": "full_reload"})
        return True

    def reload(self) -> None:
        """重新从数据库加载所有配置数据"""
        with self._lock:
            self._root.clear()
            self._load_data()
            # 异步广播全量刷新
            self._enqueue_cmd({"op": "full_reload"})

    def list_all_model_ids(self) -> list:
        """获取所有model_id列表"""
        self._refresh_from_snapshot_if_modified()
        return list(self._root.keys())

    def get_all_configs(self) -> Dict[str, ModelConfigNode]:
        """获取所有配置数据"""
        self._refresh_from_snapshot_if_modified()
        return self._root


# 全局配置缓存树形结构实例
quota_config_cache_tree = QuotaConfigCacheTree()
