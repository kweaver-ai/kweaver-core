"""
Pytest 配置文件

确保测试按顺序执行，并在测试之间添加延迟以避免系统过载。
"""
import asyncio
import time


def pytest_configure(config):
    """Pytest 配置钩子"""
    # 禁用并行测试
    config.pluginmanager.set_blocked("pytest-xdist")


def pytest_runtest_setup(item):
    """在每个测试开始前执行"""
    # 添加延迟，避免同时启动多个容器
    time.sleep(0.5)


def pytest_runtest_teardown(item, nextitem):
    """在每个测试结束后执行"""
    # 确保异步资源被清理
    try:
        loop = asyncio.get_event_loop()
        if loop and not loop.is_closed():
            # 运行所有待处理的任务
            pending = asyncio.all_tasks(loop)
            if pending:
                loop.run_until_complete(asyncio.gather(*pending, return_exceptions=True))
    except Exception:
        pass
