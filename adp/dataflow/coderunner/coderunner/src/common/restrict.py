import builtins
import resource

class RestrictedImport:
    not_allowed_packages = ['os', 'subprocess', 'pickle', 'sys']

    def __call__(self, name, globals=None, locals=None, fromlist=(), level=0):
        if name not in self.not_allowed_packages:
            return builtins.__import__(name, globals, locals, fromlist, level)
        else:
            raise Exception(f"Importing package '{name}' is not allowed")
        
# 设置资源限制的函数
def set_resource_limits():
    # 设置内存限制
    resource.setrlimit(resource.RLIMIT_AS, (80 * 1024 * 1024, 80 * 1024 * 1024))