import asyncio
import json
import threading
import time
from collections import defaultdict

from app.commons.snow_id import worker
from app.mydb.ConnectUtil import MyKafkaClient, get_redis_util, redis_util
from app.logs.stand_log import StandLogger
from app.dao.model_used_audit_dao import model_op_dao
from app.interfaces.dbaccess import ModelUsedAuditInfo
import datetime

from app.utils.config_cache import quota_config_cache_tree


class KafkaStreamsProcessor:
    def __init__(self, topic_name='tenant_a.dip.model_manager.quota_data', group_id='quota_data_group_new',
                 consume_from_beginning=False):
        # 使用新的消费者组ID，确保能够消费到所有消息
        self.kafka_client = MyKafkaClient(topic_name)
        self.topic_name = topic_name
        self.group_id = group_id
        self.consume_from_beginning = consume_from_beginning  # 是否从最早的消息开始消费
        # 存储按model_id和user_id联合主键汇总的数据
        self.aggregated_data = defaultdict(lambda: {
            'input_tokens': 0,
            'output_tokens': 0,
            'conf_id': '',
            'model_id': '',
            'user_id': '',
            'total_price': 0.0,
            'currency_type': 0,
            'price_type': [],
            'referprice_in': 0.0,
            'referprice_out': 0.0,
            'total_count': 0,
            'failed_count': 0,
            'average_total_time': 0.0,
            'average_first_time': 0.0,
            'total_time': 0.0,
            'first_time': 0.0
        })
        self.model_op_dao = model_op_dao
        self.lock = threading.Lock()
        self.running = True  # 添加运行状态标志
        self.processed_messages = set()  # 用于记录已处理的消息，防止重复处理
        self.is_processing = False  # 添加处理状态标志，防止重复执行

    def _connect_consumer_with_custom_config(self):
        """使用自定义配置连接消费者"""
        from confluent_kafka import Consumer
        from app.core.config import base_config
        import socket

        # 使用固定的消费者组ID，确保多实例之间实现分区协作而不是重复消费
        hostname = socket.gethostname()
        import os
        pid = os.getpid()

        # 自定义消费者配置，避免重复消费
        offset_reset = 'earliest' if self.consume_from_beginning else 'latest'
        consumer_config = {
            'bootstrap.servers': '{}:{}'.format(base_config.KAFKAHOST, base_config.KAFKAPORT),
            'security.protocol': 'sasl_plaintext',
            'enable.ssl.certificate.verification': 'false',
            'sasl.mechanism': 'PLAIN',
            'sasl.username': base_config.KAFKAUSER,
            'sasl.password': base_config.KAFKAPASS,
            'group.id': self.group_id,  # 使用固定消费者组，避免多实例重复消费
            'auto.offset.reset': offset_reset,  # 根据配置决定从何处开始消费
            'enable.auto.commit': True,
            'auto.commit.interval.ms': 5000,  # 增加提交间隔
            'session.timeout.ms': 30000,  # 增加会话超时时间
            'heartbeat.interval.ms': 10000,  # 心跳间隔
            'max.poll.interval.ms': 300000,  # 最大轮询间隔
            'fetch.wait.max.ms': 500,  # 最大等待时间
            'fetch.min.bytes': 1,  # 最小字节数
            'fetch.max.bytes': 52428800  # 最大字节数
        }

        StandLogger.info_log(f"消费者配置: group.id={self.group_id}, auto.offset.reset={offset_reset}")
        StandLogger.info_log(f"主机名: {hostname}, 进程ID: {pid}")

        # 创建消费者
        self.kafka_client.consumer = Consumer(consumer_config)
        # 订阅topic
        self.kafka_client.consumer.subscribe([self.topic_name])

        StandLogger.info_log(f"消费者已订阅 Topic: {self.topic_name}")

    def start_consumer(self):
        """启动Kafka消费者"""
        StandLogger.info_log(f"启动Kafka消费者... Topic: {self.topic_name}, Group ID: {self.group_id}")

        try:
            StandLogger.info_log("正在连接Kafka消费者...")
            # 使用自定义配置连接消费者
            self._connect_consumer_with_custom_config()
            StandLogger.info_log("Kafka消费者连接成功")
        except Exception as e:
            StandLogger.error(f"连接Kafka消费者失败: {e}")
            raise

        # 启动定时任务（仅启动一次）
        # 使用线程来运行定时任务
        if not hasattr(self, "_timer_thread") or self._timer_thread is None or not self._timer_thread.is_alive():
            StandLogger.info_log("启动定时数据处理任务...")
            import threading
            self._timer_thread = threading.Thread(target=self._run_periodic_processing, daemon=True)
            self._timer_thread.start()
            StandLogger.info_log("定时数据处理任务已启动")
        else:
            StandLogger.info_log("定时数据处理任务已在运行，跳过重复启动")

        # 持续消费消息（批量）
        message_count = 0
        StandLogger.info_log("开始消费Kafka消息...")
        while self.running:
            try:
                batch = self.kafka_client.consume_batch(num_messages=500, timeout=0.2)
                if batch:
                    for message in batch:
                        message_count += 1
                        # StandLogger.info_log(
                        #     f"收到第{message_count}条消息: topic={message['topic']}, partition={message['partition']}, offset={message['offset']}")
                        self._process_message(message)
            except Exception as e:
                StandLogger.error(f"消费Kafka消息时出错: {e}")
                import time
                time.sleep(1)

        StandLogger.info_log("Kafka消费者已停止")

    def _process_message(self, message):
        """处理单条Kafka消息"""
        try:
            # 生成消息唯一标识符（基于topic、partition、offset）
            message_id = f"{message['topic']}_{message['partition']}_{message['offset']}"

            # 检查是否已经处理过此消息（幂等性检查）
            with self.lock:
                if message_id in self.processed_messages:
                    StandLogger.info_log(f"消息已处理过，跳过: {message_id}")
                    return
                self.processed_messages.add(message_id)

                # 限制已处理消息集合的大小，避免内存泄漏
                if len(self.processed_messages) > 10000:
                    # 保留最近5000个消息ID
                    self.processed_messages = set(list(self.processed_messages)[-5000:])

            # 解析消息内容
            value = message['value']
            if isinstance(value, bytes):
                value = value.decode('utf-8')

            data = json.loads(value)
            StandLogger.info_log(f"接收到消息: {data}")
            # 提取关键字段
            model_id = data.get('model_id', '')
            user_id = data.get('user_id', '')
            status = data.get('status', '')

            if not model_id or not user_id or not status:
                StandLogger.warn(f"消息缺少model_id或user_id或status: {data}")
                return

            # 使用model_id和user_id作为联合主键
            key = f"{model_id}_{user_id}_{status}"
            # 累加input_tokens和output_tokens
            with self.lock:
                self.aggregated_data[key]['input_tokens'] += data.get('input_tokens', 0)
                self.aggregated_data[key]['output_tokens'] += data.get('output_tokens', 0)
                self.aggregated_data[key]['total_count'] += 1
                if status == "failed":
                    self.aggregated_data[key]['failed_count'] += 1
                # 累加总时间和首字时间（仅统计成功请求）
                if status != "failed":
                    self.aggregated_data[key]['total_time'] += data.get('total_time', 0.0)
                    self.aggregated_data[key]['first_time'] += data.get('first_time', 0.0)
                # 保存其他字段（假设同一种model_id和user_id组合的其他字段是相同的）
                self.aggregated_data[key]['conf_id'] = data.get('conf_id', '')
                self.aggregated_data[key]['model_id'] = model_id
                self.aggregated_data[key]['user_id'] = user_id
                price_dict = {
                    "thousand": 1000,
                    "million": 1000000
                }
                price_type = quota_config_cache_tree[model_id].price_type
                if quota_config_cache_tree[model_id].billing_type == 1:
                    total_price = data.get('input_tokens', 0) * (
                            quota_config_cache_tree[model_id].referprice_in / price_dict.get(price_type[0],
                                                                                             1000)) + data.get(
                        'output_tokens', 0) * (quota_config_cache_tree[model_id].referprice_out / price_dict.get(
                        price_type[1], 1000))
                else:
                    total_price = (data.get('input_tokens', 0) + data.get('output_tokens', 0)) * \
                                  quota_config_cache_tree[model_id].referprice_in / price_dict.get(price_type[0], 1000)
                self.aggregated_data[key]['total_price'] += total_price
                self.aggregated_data[key]['currency_type'] = quota_config_cache_tree[model_id].currency_type
                self.aggregated_data[key]['price_type'] = price_type
                self.aggregated_data[key]['referprice_in'] = quota_config_cache_tree[model_id].referprice_in
                self.aggregated_data[key]['referprice_out'] = quota_config_cache_tree[model_id].referprice_out
            # StandLogger.info_log(
            #     f"处理消息: model_id={model_id}, user_id={user_id}, conf_id={conf_id}, input_tokens={data.get('input_tokens', 0)}, output_tokens={data.get('output_tokens', 0)}")
        except json.JSONDecodeError as e:
            StandLogger.error(f"解析Kafka消息失败: {e}")
        except Exception as e:
            StandLogger.error(f"处理Kafka消息时出错: {e}")

    def _run_periodic_processing(self):
        """运行定时数据处理的线程函数"""
        StandLogger.info_log("创建定时汇总数据任务成功")
        import time

        while self.running:
            try:
                StandLogger.info_log("300秒后开始执行定时汇总任务")

                # 等待300秒
                time.sleep(300)

                if not self.running:
                    break

                # 处理数据
                self._process_aggregated_data()
            except Exception as e:
                StandLogger.error(f"定期处理数据时出错: {e}")
                time.sleep(1)

    def _process_aggregated_data(self):
        """处理汇总数据并存入数据库"""
        # 检查是否正在处理中，防止重复执行
        with self.lock:
            if self.is_processing:
                StandLogger.info_log("数据正在处理中，跳过本次执行")
                return
            self.is_processing = True

        try:
            if not self.aggregated_data:
                StandLogger.info_log("没有需要处理的汇总数据")
                return

            StandLogger.info_log(f"开始处理{len(self.aggregated_data)}条汇总数据")

            # 使用锁保护复制和清空操作
            with self.lock:
                # 复制当前数据并清空原字典
                data_to_process = dict(self.aggregated_data)
                self.aggregated_data.clear()

            # 将数据分批，每批最多500条
            batch_size = 500
            data_items = list(data_to_process.items())

            # 分批处理数据
            for i in range(0, len(data_items), batch_size):
                batch_items = data_items[i:i + batch_size]
                batch_data = []

                # 收集一批数据
                for key, data in batch_items:
                    try:
                        # 计算平均时间
                        success_count = data['total_count'] - data['failed_count']
                        average_total_time = data['total_time'] / success_count if success_count > 0 else 0.0
                        average_first_time = data['first_time'] / success_count if success_count > 0 else 0.0
                        
                        audit_info = ModelUsedAuditInfo(
                            conf_id=data['conf_id'],
                            model_id=data['model_id'],
                            user_id=data['user_id'],
                            input_tokens=data['input_tokens'],
                            output_tokens=data['output_tokens'],
                            total_price=data['total_price'],
                            create_time=datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                            currency_type=data['currency_type'],
                            price_type=json.loads(data['price_type']) if isinstance(data['price_type'],
                                                                                    str) else data['price_type'],
                            referprice_in=data['referprice_in'],
                            referprice_out=data['referprice_out'],
                            total_count=data['total_count'],
                            failed_count=data['failed_count'],
                            average_total_time=average_total_time,
                            average_first_time=average_first_time
                        )
                        batch_data.append(audit_info)
                    except Exception as e:
                        StandLogger.error(f"创建ModelUsedAuditInfo对象时出错: {e}")

                # 批量保存到数据库（使用INSERT ... ON DUPLICATE KEY UPDATE实现幂等性）
                if batch_data:
                    try:
                        # 使用批量插入或更新，避免重复数据
                        affected = self.model_op_dao.batch_add_model_used_log(batch_data)
                        StandLogger.info_log(f"成功批量保存/更新{affected}条数据, 收集到{len(batch_data)}条")
                    except Exception as e:
                        StandLogger.error(f"批量保存数据到数据库时出错: {e}")

        # # 将处理失败的数据重新放回聚合字典
        # if failed_data:
        #     with self.lock:
        #         for key, data in failed_data.items():
        #             # 如果该key已存在，则累加tokens，否则直接赋值
        #             if key in self.aggregated_data:
        #                 self.aggregated_data[key]['input_tokens'] += data['input_tokens']
        #                 self.aggregated_data[key]['output_tokens'] += data['output_tokens']
        #                 # 其他字段保持不变，因为假设相同key的数据其他字段是相同的
        #             else:
        #                 self.aggregated_data[key] = data
        #     StandLogger.info_log(f"重新放回{len(failed_data)}条处理失败的数据到聚合字典")

        finally:
            # 重置处理状态标志
            with self.lock:
                self.is_processing = False

    def stop_consumer(self):
        """停止Kafka消费者"""
        StandLogger.info_log("停止Kafka消费者...")
        self.running = False
        self.kafka_client.close_consumer()
        # 等待定时线程退出
        try:
            if hasattr(self, "_timer_thread") and self._timer_thread is not None:
                self._timer_thread.join(timeout=2)
        except Exception:
            pass


# 全局实例
kafka_processor = None


def start_kafka_streams_processor():
    """启动Kafka Streams处理器"""
    global kafka_processor
    StandLogger.info_log("开始启动Kafka Streams处理器...")
    if kafka_processor is None:
        StandLogger.info_log("创建KafkaStreamsProcessor实例...")
        kafka_processor = KafkaStreamsProcessor()
        StandLogger.info_log("KafkaStreamsProcessor实例创建成功")

        StandLogger.info_log("开始调用start_consumer()方法...")
        # 直接运行消费者
        kafka_processor.start_consumer()
        StandLogger.info_log("Kafka Streams处理器已启动")
    else:
        StandLogger.info_log("KafkaStreamsProcessor实例已存在，跳过创建")
