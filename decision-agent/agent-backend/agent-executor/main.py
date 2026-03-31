# -*- coding:utf-8 -*-
import sys
import signal
import asyncio
import logging
import uvicorn

# 1. 服务启动前的boot
from app.boot import boot
from app.common.config import Config
from app.router import app

boot.on_boot_run()


server_instance = None


def signal_handler(signum, frame):
    """信号处理器"""
    print("\n接收到停止信号，正在关闭服务...")

    if server_instance:
        if server_instance.should_exit:
            server_instance.force_exit = True
        else:
            server_instance.should_exit = True


def asyncio_exception_handler(loop, context):
    """Asyncio 异常处理器 - 抑制 KeyboardInterrupt 警告"""
    exception = context.get("exception")

    # 忽略 KeyboardInterrupt 和 CancelledError
    if isinstance(exception, (KeyboardInterrupt, asyncio.CancelledError)):
        return

    # 其他异常正常处理
    if exception:
        print(f"Asyncio 异常: {exception}")


class HealthCheckFilter(logging.Filter):
    """过滤健康检查端点的日志"""

    def filter(self, record: logging.LogRecord) -> bool:
        # 过滤掉健康检查端点的访问日志
        message = record.getMessage()
        if "/health/alive" in message or "/health/ready" in message:
            return False
        return True


class ASGIExceptionFilter(logging.Filter):
    """过滤 uvicorn 的 ASGI 异常日志（因为我们已经用 error_logger 处理了）"""

    def filter(self, record: logging.LogRecord) -> bool:
        message = record.getMessage()
        # 过滤掉 "Exception in ASGI application" 日志
        # 因为我们已经在 enhanced_unknown_handler 中用 error_logger 记录了详细信息
        if "Exception in ASGI application" in message:
            return False
        return True


def main():
    # 设置 asyncio 异常处理器
    try:
        loop = asyncio.get_event_loop()
        loop.set_exception_handler(asyncio_exception_handler)
    except RuntimeError:
        # 如果没有事件循环，忽略
        pass

    global server_instance

    # 配置 Uvicorn 日志格式，添加时间信息
    log_config = uvicorn.config.LOGGING_CONFIG
    log_config["formatters"]["default"]["fmt"] = (
        "%(asctime)s - %(levelname)s - %(message)s"
    )
    log_config["formatters"]["default"]["datefmt"] = "%Y-%m-%d %H:%M:%S"
    log_config["formatters"]["access"]["fmt"] = (
        '%(asctime)s - %(levelname)s - %(client_addr)s - "%(request_line)s" %(status_code)s'
    )
    log_config["formatters"]["access"]["datefmt"] = "%Y-%m-%d %H:%M:%S"

    # 添加日志过滤器
    log_config["filters"] = {
        "health_check_filter": {
            "()": HealthCheckFilter,
        },
        "asgi_exception_filter": {
            "()": ASGIExceptionFilter,
        },
    }
    log_config["handlers"]["access"]["filters"] = ["health_check_filter"]
    log_config["handlers"]["default"]["filters"] = ["asgi_exception_filter"]

    config = uvicorn.Config(
        app,
        host=Config.app.host_ip,
        port=Config.app.port,
        log_level=Config.app.log_level,
        log_config=log_config,
    )
    server_instance = uvicorn.Server(config)

    # 注册信号处理器
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    try:
        # 运行服务
        server_instance.run()
    except (KeyboardInterrupt, SystemExit):
        print("\n服务已停止")
    except Exception as e:
        print(f"\n服务异常退出: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
