class GlobalVariableManager:
    """全局变量管理器，用于在测试用例之间共享数据"""

    _instance = None
    _global_vars = {}

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance

    def set_var(self, key, value):
        """设置全局变量"""
        self._global_vars[key] = value
        return value

    def get_var(self, key, default=None):
        """获取全局变量"""
        return self._global_vars.get(key, default)

    def del_var(self, key):
        """删除全局变量"""
        if key in self._global_vars:
            del self._global_vars[key]
            return True
        return False

    def clear(self):
        """清空所有全局变量"""
        self._global_vars.clear()


# 全局变量管理器实例，供外部直接使用
global_vars = GlobalVariableManager()
