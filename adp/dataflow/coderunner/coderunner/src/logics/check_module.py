import asyncio
import os
import re
import shutil
import subprocess
import sys
import glob
import tarfile
import threading
import time

from drivenadapters.download_pkgs import DownLoadPkgs
from errors.errors import InternalErrException
from common.logger import logger
class CheckModule:

    _instance = None
    _lock = threading.Lock()

    # 单例模式
    def __new__(cls, *args, **kwargs):
        with cls._lock:
            if not cls._instance:
                cls._instance = super(CheckModule, cls).__new__(cls)
        return cls._instance
    
    # 仅初始化一次
    def __init__(self):
        if not hasattr(self, '_initialized'):
            self.download_pkgs = DownLoadPkgs()
            self.installed_pkgs = {}
            self.lock = threading.Lock()
            self.pip_lock = threading.Lock()
            self._initialized = True

    def extract_imports_from_string(self, code: str):
        pattern = re.compile(r'^(?:from\s+(\S+)\s+import|import\s+(.+))', re.MULTILINE)
        packages = set()
    
        # 按行分割代码
        lines = code.splitlines()
        
        for line in lines:
            # 匹配每一行
            match = pattern.match(line.strip())
            if match:
                # 处理 'from module import ...' 格式
                if match.group(1):
                    packages.add(match.group(1).split('.')[0])
                # 处理 'import module1, module2' 格式
                elif match.group(2):
                    modules = [m.split('.')[0] for m in match.group(2).split(',')]
                    for module in modules:
                        module = module.strip()
                        if not module:
                            continue
                        if ' as ' in module:
                            module = module.split(' as ')[0]
                        packages.add(module)

        return packages

    def check_module(self, module):
        try:
            subprocess.check_call([sys.executable, "-c", f"import {module}"])
            return True
        except subprocess.CalledProcessError:
            return False

    async def install_module(self, module_name):
        '''
        安装python包
        安装状态
        -1: 安装成功
        -2: 安装失败
        '''
        try:
            module_name = module_name.lower().strip()
            workspace = "/tmp/pkgs"
            tar_path = f"{workspace}/{module_name}.tar"
            package_path = f"{workspace}/{module_name}"

            # 判断目录是否存在，不存在则创建
            if not os.path.exists(workspace):
                os.makedirs(workspace)

            await self.download_pkg(module_name, tar_path)
            with tarfile.open(tar_path, 'r') as tar:
                tar.extractall(path=workspace)

            packages = glob.glob(f"{package_path}/*")
            # 单线程安装包，避免并发安装包出现依赖冲突
            with self.pip_lock:
                subprocess.check_call([sys.executable, "-m", "pip", "install"] + packages)
            self.installed_pkgs[module_name] = -1
            return True
        except subprocess.CalledProcessError as e:
            logger.warning(f"[install_module] module {module_name} install failed, detail: {str(e)}")
            self.installed_pkgs[module_name] = -2
            return False
        except Exception as e:
            logger.warning(f"[install_module] module {module_name} install failed, detail: {str(e)}")
            self.installed_pkgs[module_name] = -2
            return False
        finally:
            # 删除解压后的文件夹
            os.remove(tar_path)
            shutil.rmtree(package_path)
    
    async def uninstall_module(self, module_name: str):
        try:
            subprocess.check_call([sys.executable, "-m", "pip", "uninstall", "-y", module_name])
        except Exception as e:
            raise InternalErrException(detail= f"[uninstall_module] module {module_name} uninstall failed, detail: {str(e)}")

    async def download_pkg(self, module_name: str, file_path: str):
        data = await self.download_pkgs.download_pkg(module_name)
        file_size = len(data)
        if file_size == -1:
            raise InternalErrException(detail= "[download_pkg] file download failed when get object meta info")
        
        try:
            with open(file_path, 'wb') as f:
                f.write(data)
        except Exception as e:
            raise InternalErrException(detail= f"[download_pkg] file download failed when create tmp file, detail: {str(e)}")
        
    def check_and_install_modules(self, code: str):
        modules = self.extract_imports_from_string(code)
        for module in modules:
            with self.lock:
                if module not in self.installed_pkgs:
                    # 首次安装初始状态
                    self.installed_pkgs[module] = 1
                elif self.installed_pkgs[module] >= 1:
                    # 已经在安装中的包，则版本加1标识即可
                    self.installed_pkgs[module] += 1
                elif self.installed_pkgs[module] == -2:
                    # 若上一次安装失败，则进行重新安装时重置状态
                    self.installed_pkgs[module] = 1

            #  已经处于安装中的包，等待安装完成
            version = self.installed_pkgs[module]
            while module in self.installed_pkgs and version > 1 :
                time.sleep(1)
                version = self.installed_pkgs[module]

            # 安装失败，结束安装
            if self.installed_pkgs[module] == -2:
                raise InternalErrException(detail= f"[check_and_install_modules] module {module} install failed")
            
            if self.check_module(module):
                self.installed_pkgs[module] = -1
                continue

            asyncio.run(self.install_module(module))