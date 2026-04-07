"""
字典路径解析工具包

提供强大的字典和列表路径解析功能
"""

from .dict_path_parser import (
    DictPathParser,
    DictPathParserFlat,
    get_dict_val_by_path,
    get_dic_val_by_path_flat,
    set_dict_val_by_path,
)

__all__ = [
    "DictPathParser",
    "DictPathParserFlat",
    "get_dict_val_by_path",
    "get_dic_val_by_path_flat",
    "set_dict_val_by_path",
]
