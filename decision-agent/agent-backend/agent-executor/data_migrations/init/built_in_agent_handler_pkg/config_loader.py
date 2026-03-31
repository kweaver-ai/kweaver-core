"""
配置加载功能模块
"""


def load_agent_config(file_path) -> dict:
    """
    加载agent配置文件中的配置信息

    Args:
        file_path: 配置文件路径

    Returns:
        dict: agent配置信息，包含以下字段：
            - name: agent名称
            - description: agent描述
            - avatar_type: 头像类型
            - avatar: 头像标识
            - product_key: 产品标识
            - key: agent唯一标识
            - config: agent配置字典

    Raises:
        FileNotFoundError: 当配置文件不存在时
        Exception: 当配置文件解析失败时

    Examples:
        >>> config = load_agent_config("/path/to/agent_config.py")
        >>> print(f"Agent name: {config['name']}")
        >>> print(f"Config: {config['config']}")
    """

    # 1. 读取配置文件
    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    # 2. 使用exec执行文件内容，但限制在局部作用域中
    local_vars = {}
    exec(content, {}, local_vars)

    # 3. 从局部作用域中获取配置信息
    config = {
        "name": local_vars.get("name", ""),
        "description": local_vars.get("description", ""),
        "avatar_type": 1,
        "avatar": "1",
        "product_key": "dip",
        "key": local_vars.get("name", ""),
        "config": local_vars.get("config", {}),
    }
    return config
