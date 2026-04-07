#!/usr/bin/env python3
"""
独立的 Kafka 消费者进程启动脚本
"""
import os
import sys
import signal
import multiprocessing
from app.logs.stand_log import StandLogger
from app.core.config import base_config
from app.utils.config_cache import quota_config_cache_tree  # 初始化配置缓存


class KafkaConsumerProcess:
    def __init__(self):
        self.process = None
        self.kafka_processor = None
        self.running = False

    def signal_handler(self, signum, frame):
        """信号处理器，用于优雅关闭"""
        StandLogger.info_log(f"收到信号 {signum}，开始优雅关闭 Kafka 消费者...")
        self.running = False
        # 停止全局的 kafka_processor
        try:
            from app.utils.kafka_streams_processor import kafka_processor
            if kafka_processor:
                kafka_processor.stop_consumer()
        except Exception as e:
            StandLogger.error(f"停止 Kafka 处理器时出错: {e}")

    def run_consumer(self):
        """运行 Kafka 消费者的函数"""
        try:
            StandLogger.info_log("Kafka 消费者进程启动")
            self.running = True
            
            # 注册信号处理器
            signal.signal(signal.SIGINT, self.signal_handler)
            signal.signal(signal.SIGTERM, self.signal_handler)
            StandLogger.info_log("信号处理器已注册")
            StandLogger.info_log("正在导入 Kafka Streams 处理器...")
            from app.utils.kafka_streams_processor import start_kafka_streams_processor
            StandLogger.info_log("Kafka Streams 处理器导入成功")
            StandLogger.info_log("开始启动 Kafka Streams 处理器...")
            start_kafka_streams_processor()
            StandLogger.info_log("Kafka Streams 处理器启动完成")
            
        except Exception as e:
            StandLogger.error(f"Kafka 消费者进程运行出错: {e}")
            import traceback
            StandLogger.error(f"详细错误信息: {traceback.format_exc()}")
            raise
        finally:
            StandLogger.info_log("Kafka 消费者进程结束")

    def start(self):
        """启动 Kafka 消费者进程"""
        try:
            # 直接运行消费者
            self.run_consumer()
        except KeyboardInterrupt:
            StandLogger.info_log("收到键盘中断信号，关闭 Kafka 消费者")
        except Exception as e:
            StandLogger.error(f"启动 Kafka 消费者进程失败: {e}")
            sys.exit(1)


def main():
    """主函数"""
    StandLogger.info_log("=== Kafka 消费者进程启动 ===")  # 控制台输出
    StandLogger.info_log("启动独立的 Kafka 消费者进程")
    
    # 创建并启动 Kafka 消费者
    StandLogger.info_log("创建 KafkaConsumerProcess 实例...")  # 控制台输出
    consumer_process = KafkaConsumerProcess()
    StandLogger.info_log("开始启动消费者...")  # 控制台输出
    consumer_process.start()


if __name__ == '__main__':
    main()
