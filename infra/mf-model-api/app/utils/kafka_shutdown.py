"""
Kafka异步生产者优雅关闭工具
"""
import atexit
import signal
import sys
from app.mydb.ConnectUtil import kafka_client
from app.logs.stand_log import StandLogger


def graceful_shutdown():
    """优雅关闭Kafka异步生产者"""
    try:
        StandLogger.info("开始优雅关闭Kafka异步生产者...")
        kafka_client.shutdown_async_producer()
        StandLogger.info("Kafka异步生产者已成功关闭")
    except Exception as e:
        StandLogger.error(f"关闭Kafka异步生产者时出错: {e}")


def signal_handler(signum, frame):
    """信号处理器"""
    StandLogger.info(f"收到信号 {signum}，开始优雅关闭...")
    graceful_shutdown()
    sys.exit(0)


def register_shutdown_handlers():
    """注册关闭处理器"""
    # 注册程序退出时的清理函数
    atexit.register(graceful_shutdown)
    
    # 注册信号处理器
    signal.signal(signal.SIGINT, signal_handler)   # Ctrl+C
    signal.signal(signal.SIGTERM, signal_handler)  # 终止信号
    
    StandLogger.info("Kafka优雅关闭处理器已注册")


# 自动注册关闭处理器
register_shutdown_handlers()
