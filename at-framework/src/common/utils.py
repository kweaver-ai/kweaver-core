import os
import re
import string
from base64 import b64encode
from typing import List, Dict, Any, Optional

import pandas
from Crypto.Cipher import PKCS1_v1_5
from Crypto.PublicKey import RSA
from deepmerge import Merger

from src.config.setting import RESOURCE_DIR, SHEET_TITLE_MAPPING


def load_case(file):
    """
    读取excel中的用例配置
    """
    global_params = read_excel(file,
                               sheet=SHEET_TITLE_MAPPING["global"]["sheet"],
                               index=SHEET_TITLE_MAPPING["global"]["index"],
                               **SHEET_TITLE_MAPPING["global"]["mapping"])
    api_params = read_excel(file,
                            sheet=SHEET_TITLE_MAPPING["api"]["sheet"],
                            index=SHEET_TITLE_MAPPING["api"]["index"],
                            **SHEET_TITLE_MAPPING["api"]["mapping"])
    suite_params = read_excel(file, fill=True,
                              sheet=SHEET_TITLE_MAPPING["suite"]["sheet"],
                              **SHEET_TITLE_MAPPING["suite"]["mapping"])

    # 依据suite配置，加载用例
    case_list = [{**y, **{"feature": x["feature"], "story": x["story"]}}
                 for x in suite_params if str(x["switch"]).lower() == 'y'
                 for y in read_excel(file, sheet=x["story"],
                                     **SHEET_TITLE_MAPPING["case"]["mapping"])]

    # 替换API信息
    case_list = [{**x, **api_params[x["url"]]} for x in case_list if x["url"] in api_params]

    # 全局变量中存在嵌套引用, 需对params额外执行一次参数替换
    params = {k: v["value"] for k, v in global_params.items()}
    params = {k: string.Template(str(v)).safe_substitute(**params) for k, v in params.items()}
    global_params = {k: string.Template(str(v)).safe_substitute(**params) for k, v in params.items()}

    # 替换用例中的全局变量
    case_list = [replace_params(x, **global_params) for x in case_list]

    return case_list


def read_excel(file, sheet, fill=False, index=None, **kwargs):
    sheet_info = []

    '''
    自动填充时，需要设置keep_default_na=True，否则无法识别填充项
    固定为前向填充
    '''
    if fill:
        df = pandas.read_excel(file, sheet_name=sheet, engine="openpyxl").ffill()
    else:
        df = pandas.read_excel(file, sheet_name=sheet, keep_default_na=False, engine="openpyxl")

    if kwargs:
        for line in df.itertuples():
            sheet_info.append({k: getattr(line, v) for k, v in kwargs.items()})

    # 指定索引字段时, 转换输出格式
    if index:
        sheet_info = pandas.DataFrame(sheet_info).set_index(index).to_dict("index")

    return sheet_info


def replace_params(input_case, **kwargs):
    """
    使用JinJa2.Template会导致未配置的参数项置空
    故此处应用string.Template，仅替换必要信息。用例执行时渲染其余参数
    """
    output_case = {k: string.Template(str(v)).safe_substitute(kwargs) for k, v in input_case.items()}
    return output_case


merger = Merger(
    # 传递原始策略
    [(dict, ["merge"])],
    # 回退策略
    ["override"],
    # 类型冲突策略
    ["override"]
)


def merge_with_deepmerge(dict1, dict2):
    """使用 deepmerge 库合并"""
    return merger.merge(dict1, dict2)


def filter_response_data(
        data: List[Dict[str, Any]], conditions: Dict[str, Any], match_mode: str = "all"
) -> List[Dict[str, Any]]:
    """
    通用响应数据筛选方法

    Args:
        data: 响应数据列表
        conditions: 筛选条件字典
        match_mode: 匹配模式 - "all"(所有条件都满足) 或 "any"(任一条件满足)

    Returns:
        符合条件的数据项列表

    Example:
        # 查找name以auto_test开头，host为指定IP，database_name为指定库的数据
        conditions = {
            "name": {"startswith": "auto_test"},
            "bin_data.host": {"equals": "10.4.174.236"},
            "bin_data.database_name": {"equals": "ldbc_snb_01"}
        }
        result = filter_response_data(response_data, conditions)
    """

    def check_condition(
            item: Dict[str, Any], field_path: str, condition: Dict[str, Any]
    ) -> bool:
        """检查单个条件是否满足"""
        # 获取嵌套字段值
        value = get_nested_value(item, field_path)

        if value is None:
            return False

        # 根据条件类型进行判断
        for condition_type, expected_value in condition.items():
            if condition_type == "equals":
                if value != expected_value:
                    return False
            elif condition_type == "startswith":
                if not str(value).startswith(str(expected_value)):
                    return False
            elif condition_type == "endswith":
                if not str(value).endswith(str(expected_value)):
                    return False
            elif condition_type == "contains":
                if str(expected_value) not in str(value):
                    return False
            elif condition_type == "regex":
                if not re.search(str(expected_value), str(value)):
                    return False
            elif condition_type == "in":
                if value not in expected_value:
                    return False
            elif condition_type == "not_in":
                if value in expected_value:
                    return False
            elif condition_type == "gt":
                if not (isinstance(value, (int, float)) and value > expected_value):
                    return False
            elif condition_type == "lt":
                if not (isinstance(value, (int, float)) and value < expected_value):
                    return False
            elif condition_type == "gte":
                if not (isinstance(value, (int, float)) and value >= expected_value):
                    return False
            elif condition_type == "lte":
                if not (isinstance(value, (int, float)) and value <= expected_value):
                    return False
            else:
                raise ValueError(f"不支持的条件类型: {condition_type}")

        return True

    def get_nested_value(data: Dict[str, Any], path: str) -> Any:
        """获取嵌套字段的值，支持点号分隔的路径"""
        keys = path.split(".")
        current = data

        for key in keys:
            if isinstance(current, dict) and key in current:
                current = current[key]
            else:
                return None

        return current

    def item_matches_conditions(item: Dict[str, Any]) -> bool:
        """检查数据项是否满足所有条件"""
        results = []

        for field_path, condition in conditions.items():
            result = check_condition(item, field_path, condition)
            results.append(result)

        # 根据匹配模式返回结果
        if match_mode == "all":
            return all(results)
        elif match_mode == "any":
            return any(results)
        else:
            raise ValueError(f"不支持的匹配模式: {match_mode}")

    # 筛选数据
    return [item for item in data if item_matches_conditions(item)]


def find_first_match(
        data: List[Dict[str, Any]], conditions: Dict[str, Any]
) -> Optional[Dict[str, Any]]:
    """
    查找第一个符合条件的数据项

    Args:
        data: 响应数据列表
        conditions: 筛选条件字典

    Returns:
        第一个符合条件的数据项，如果没有则返回None
    """
    results = filter_response_data(data, conditions)
    return results[0] if results else None


def convert_db_type_to_support_type(db_type: str) -> str | None:
    """
    将数据库数据类型转换为业务知识网络支持的类型

    参数:
        db_type: 数据库原始数据类型（如 'VARCHAR'、'tinyint'、'timestamp' 等）

    返回:
        转换后的目标类型（如 'varchar'、'short'、'timestamp' 等）；若未匹配则返回 None
    """
    # 统一转为小写，避免大小写差异导致匹配失败
    db_type_lower = db_type.lower()

    # 定义「数据库类型 -> 业务支持类型」的映射关系
    type_mapping = {
        # 布尔类型
        "bool": "boolean",
        "boolean": "boolean",
        # 短整型（对应 smallint/tinyint 等）
        "tinyint": "short",
        "smallint": "short",
        # 整型
        "int": "integer",
        "integer": "integer",
        # 长整型
        "bigint": "long",
        # 浮点型
        "float": "float",
        # 双精度浮点
        "double": "double",
        "double precision": "double",
        # 定点数（decimal/numeric）
        "decimal": "decimal",
        "numeric": "decimal",
        # 字符串类型
        "varchar": "sting",
        "char": "sting",
        "character varying": "sting",
        "keyword": "keyword",  # 若数据库直接支持 keyword 类型
        "string": "string",
        # 文本类型
        "text": "text",
        "longtext": "text",
        # 时间类型
        "timestamp": "timestamp",
        "datetime": "datetime",
        "date": "date",
        # 向量类型
        "vector": "vector",
    }

    # 遍历映射表，匹配数据库类型
    for db_type_pattern, target_type in type_mapping.items():
        if db_type_pattern in db_type_lower:
            return target_type

    # 无匹配类型时返回 None（或根据需求抛异常/返回默认值）
    return None


def filter_tables_by_name(data_list: list, target_tables: list) -> list:
    """
    根据字典的 'name' 字段，从原始数据列表中筛选出目标表名对应的字典

    参数:
        data_list: list - 原始数据列表（即用户提供的包含表结构的大列表）
        target_tables: list - 目标表名列表（如 ['person', 'post']）

    返回:
        list - 筛选后的数据列表（仅包含 'name' 在 target_tables 中的字典）
    """
    # 列表推导式高效筛选：遍历原始数据，匹配name字段，同时用get避免KeyError
    filtered_result = [
        table_dict
        for table_dict in data_list
        if table_dict.get("technical_name")
           in target_tables  # get('name')确保无name键时不报错
    ]
    return filtered_result


def pwd_RSABase64(message: str) -> str:
    """使用项目 resource 目录下的 rsa_public.key 进行 RSA 加密并 base64 编码"""
    key = os.path.join(RESOURCE_DIR, "dataconnect_rsa_piblick.key")
    with open(key, "r") as f:
        pubkey = f.read()
    public_key = RSA.import_key(pubkey)
    cipher = PKCS1_v1_5.new(public_key)
    encrypted_message = cipher.encrypt(message.encode("utf-8"))
    return b64encode(encrypted_message).decode("utf-8")


def get_unique_identities(d: dict):
    return [
        obj["_instance_identity"] for obj in d.values() if "_instance_identity" in obj
    ]


def get_id_data(d: dict):
    id_kv_list = []
    for item in d:
        # 从_instance_identity字段中提取数据
        instance_identity = item.get("_instance_identity", {})
        # 遍历_instance_identity字典的每个键值对，筛选键包含"id"的项
        for key, value in instance_identity.items():
            if "id" in key.lower():  # 忽略大小写（如ID、Id也会被匹配）
                id_kv_list.append({key: value})
    return id_kv_list


def normalize_list_of_dicts(lst):
    # 每个 dict -> 排好序的 (key, value) 元组，再组成列表
    return sorted([tuple(sorted(d.items())) for d in lst])


def check_file_exists_with_os(file_path):
    """
    使用os.path检查文件是否存在
    :param file_path: 文件路径（相对/绝对）
    :return: True/False
    """
    # 第一步：检查路径是否存在（文件/文件夹都算）
    if not os.path.exists(file_path):
        return False
    # 第二步：进一步判断是否是文件（避免路径是文件夹的情况）
    if os.path.isfile(file_path):
        return True
    else:
        return False


# 使用示例
if __name__ == "__main__":
    import json
    from jinja2 import Environment
    from src.common.data_gen import random_string

    env = Environment()
    env.globals['random_string'] = random_string

    case_list = load_case("../../test_data/etrino/etrino_case.xlsx")
    for case_info in case_list:
        case_info = {k: env.from_string(v).render() for k, v in case_info.items()}
        case_body_params = json.loads(json.dumps(case_info["body_params"], ensure_ascii=False)) if case_info["body_params"] != '' else {}
        print("=================================")
        print(case_body_params)


