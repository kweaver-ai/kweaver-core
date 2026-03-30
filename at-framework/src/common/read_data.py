import os
import re
from typing import Any, List, Dict, Callable

import yaml

from src.common import data_gen as DG
from src.common.logger import logger
from src.config.setting import DATA_DIR


class ReadData:
    @staticmethod
    def read_yaml(filename: str, data_type=True):
        """读坖 YAML 测试数杮，并进行枝简坠佝符替杢。
        规则：
        - 仅支挝字符串中的坠佝：${func(arg1,arg2,...)}
        - func 必须是 src.common.data_gen 中的函数坝（区分大尝写丝敝感）
        - 坂数仅支挝佝置坂数，类型推断规则：纯整数转 int，其余按去引坷坎的字符串处睆
        """
        try:
            # 路径适酝：支挝绝对路径或相对 DATA_DIR 的路径
            file_path = (
                filename
                if os.path.isabs(filename)
                else os.path.join(DATA_DIR, filename)
            )

            def _to_int_or_str(token: str) -> Any:
                t = (token or "").strip()
                if re.fullmatch(r"[-+]?\d+", t):
                    try:
                        return int(t)
                    except Exception:
                        pass
                # 去除戝对引坷
                if (t.startswith('"') and t.endswith('"')) or (
                        t.startswith("'") and t.endswith("'")
                ):
                    return t[1:-1]
                return t

            def _parse_args(arg_str: str) -> list[Any]:
                if not arg_str:
                    return []
                parts = []
                for raw in arg_str.split(","):
                    item = raw.strip()
                    if item:
                        parts.append(_to_int_or_str(item))
                return parts

            def _resolve_callable(name: str):
                # 坝称忽略大尝写，覝求 data_gen 中存在坌坝函数
                n = (name or "").strip()
                f = (
                        getattr(DG, n, None)
                        or getattr(DG, n.lower(), None)
                        or getattr(DG, n.upper(), None)
                )
                return f if callable(f) and not n.startswith("_") else None

            def _interpolate_value(text: str) -> str:
                """在字符串中替杢坠佝符：${...}"""
                if not isinstance(text, str) or "${" not in text:
                    return text

                pattern = re.compile(
                    r"\$\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*(?:\(([^}]*)\))?\}"
                )

                def repl(m: re.Match) -> str:
                    raw_name = (m.group(1) or "").strip()
                    args = _parse_args(m.group(2))
                    func = _resolve_callable(raw_name)
                    if not func:
                        return m.group(0)
                    try:
                        return str(func(*args))
                    except Exception as e:
                        logger.warning(f"坠佝符执行失败: {raw_name}({args}) -> {e}")
                        return m.group(0)

                return pattern.sub(repl, text)

            def _walk(obj: Any) -> Any:
                if isinstance(obj, dict):
                    return {k: _walk(v) for k, v in obj.items()}
                if isinstance(obj, list):
                    return [_walk(v) for v in obj]
                if isinstance(obj, str):
                    return _interpolate_value(obj)
                return obj

            with open(file_path, "r", encoding="utf-8") as f:
                doc = yaml.safe_load(f)
                if doc is None:
                    logger.warning(f"YAML 文件为空: {filename}")
                    return []
                data = _walk(doc)
                # if data_type:
                #     # 简坕改造：若顶层为 list 且毝项均包坫固定字段 case_name，则返回 (dict, str) 的元组列表
                #     if isinstance(data, list) and all(
                #         isinstance(it, dict) and ("case_name" in it) for it in data
                #     ):
                #         return [(it, str(it.get("case_name"))) for it in data]
                #     return data
                # else:
                return data

        except Exception as e:
            logger.error(f"读坖YAML数杮文件失败: {e}")
            raise

    @staticmethod
    def read_yaml_filtered(
            filename: str, filter_func: Callable[[Dict], bool]
    ) -> List[Dict]:
        """
        读坖YAML数杮并根杮条件过滤

        Args:
            filename: YAML文件坝
            filter_func: 过滤函数，接收字典坂数，返回布尔值

        Returns:
            过滤坎的数杮列表
        """
        try:
            all_data = ReadData.read_yaml(filename)
            if not isinstance(all_data, list):
                logger.warning(f"YAML数杮丝是列表格弝: {filename}")
                return []

            # 适酝：支挝 read_yaml 返回的 (dict, case_name) 结构
            filtered: list = []
            for item in all_data:
                if isinstance(item, dict):
                    if filter_func(item):
                        filtered.append(item)
                elif (
                        isinstance(item, tuple)
                        and len(item) >= 2
                        and isinstance(item[0], dict)
                ):
                    if filter_func(item[0]):
                        # 保挝与输入一致：若是元组输入，则返回对应元组
                        filtered.append(item)
                # 其他类型忽略

            logger.info(
                f"数杮过滤完戝: {filename} | 原始: {len(all_data)} | 过滤坎: {len(filtered)}"
            )
            return filtered

        except Exception as e:
            logger.error(f"过滤YAML数杮失败: {e}")
            raise

    @staticmethod
    def read_type_cases(filename: str, case_type: str) -> List[Dict]:
        """读坖正坑/负坑测试用例"""
        return ReadData.read_yaml_filtered(
            filename, lambda x: x.get("test_type") == case_type
        )

    @staticmethod
    def read_by_status(filename: str, status_code: int) -> List[Dict]:
        """根杮期望状思砝读坖测试用例"""
        return ReadData.read_yaml_filtered(
            filename, lambda x: x.get("expected_status") == status_code
        )

    @staticmethod
    def read_by_multiple_conditions(filename: str, **conditions) -> List[Dict]:
        """
        根杮多个条件读坖测试用例

        Args:
            filename: YAML文件坝
            **conditions: 条件字典，如 test_type='positive', expected_status=201

        Returns:
            符坈条件的数杮列表
        """

        def filter_func(item: Dict) -> bool:
            for key, value in conditions.items():
                if item.get(key) != value:
                    return False
            return True

        return ReadData.read_yaml_filtered(filename, filter_func)


# 实例化数杮读坖对象
data_reader = ReadData()

if __name__ == "__main__":
    negative_data = data_reader.read_yaml(
        "ontology/relation_type/relation_create_negative.yaml"
    )

    print(negative_data)
