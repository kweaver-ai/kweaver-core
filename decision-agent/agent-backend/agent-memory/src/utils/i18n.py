import sys
from typing import Dict, Optional
import tomli
from pathlib import Path


class I18nManager:
    """国际化管理器"""

    _instance = None
    _initialized = False

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance

    def __init__(self):
        if not self._initialized:
            self._initialized = True
            self._load_resources()

    def _load_resources(self):
        """加载国际化资源"""
        self.resources = {}
        if hasattr(sys, "_MEIPASS"):
            locale_dir = Path(sys._MEIPASS) / "locale"
        else:
            locale_dir = Path(__file__).parent.parent / "locale"

        # 遍历所有语言目录
        for lang_dir in locale_dir.iterdir():
            if lang_dir.is_dir():
                lang = lang_dir.name
                self.resources[lang] = {}

                # 加载该语言下的所有资源文件
                for resource_file in lang_dir.glob("*.toml"):
                    resource_name = resource_file.stem
                    with open(resource_file, "rb") as f:
                        self.resources[lang][resource_name] = tomli.load(f)

    def get_error_info(
        self,
        error_code: str,
        lang: str = "en_US",
        custom_description: Optional[str] = None,
    ) -> Dict[str, str]:
        """获取错误信息

        Args:
            error_code: 错误代码
            lang: 语言代码，默认为英文

        Returns:
            Dict[str, str]: 包含错误描述、解决方案和错误链接的字典
        """
        # 如果请求的语言不存在，默认使用英文
        if lang not in self.resources:
            lang = "en_US"

        keys = error_code.split(".")
        current = self.resources[lang]["errors"]
        for k in keys:
            if isinstance(current, dict) and k in current:
                current = current[k]
            else:
                raise KeyError(f"Key {k} not found in resources for language {lang}")

        description = current.get("description", "")
        if custom_description is not None:
            description = custom_description
        return {
            "description": description,
            "solution": current.get("solution", ""),
            "error_link": current.get("error_link", ""),
        }


# 创建全局国际化管理器实例
i18n_manager = I18nManager()
