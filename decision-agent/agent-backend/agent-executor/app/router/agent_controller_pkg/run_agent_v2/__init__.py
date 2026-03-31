"""
run_agent_v2 包

将原 run_agent_v2.py 文件重构为包结构，每个函数独立为一个模块。

模块说明：
- run_agent.py: 主入口函数，处理Agent运行请求
- prepare.py: 准备Agent配置、输入和请求头
- handle_cache.py: 处理缓存初始化和状态管理
- safe_output_generator.py: 安全的输出生成器，处理客户端断开连接
- process_options.py: 处理Agent运行选项
"""

from .run_agent import run_agent

__all__ = [
    "run_agent",
]
