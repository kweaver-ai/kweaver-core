# coding: utf-8
"""
支持能力：
- 随机数字：random_int(min, max)
- 随机英文/字母数字字符串：random_str(length, kind)
- 随机中文：random_cn(length)
- UUID：gen_uuid()
- 时间戳：ts_ms()/ts_s()
"""
from __future__ import annotations

import json
import random
import string
import time
import uuid
import base64
from typing import Optional
from faker import Faker
from src.common.global_var import global_vars

_FAKER_ZH = Faker("zh_CN")


def get_global(key: str, default: str = "") -> str:
    """读取全局变量管理器中的值并转为字符串。
    用法：${get_global(key)} 或 ${get_global(key, default)}
    """

    return global_vars.get_var(key, default)


def random_string(length=8, model=8):
    """
    生成指定长度的随机字符串

    :param length: 字符串长度
    :param model: 按位指定字符串类型，
                  1=string.whitespace
                  2=string.ascii_lowercase
                  4=string.ascii_uppercase
                  8=string.digits
                  16=string.hexdigits
                  32=string.octdigits
                  64=string.punctuation
    """
    chars = ""
    if model & 1: chars = chars + string.whitespace
    if (model >> 1) & 1: chars = chars + string.ascii_lowercase
    if (model >> 2) & 1: chars = chars + string.ascii_uppercase
    if (model >> 3) & 1: chars = chars + string.digits
    if (model >> 4) & 1: chars = chars + string.hexdigits
    if (model >> 5) & 1: chars = chars + string.octdigits
    if (model >> 6) & 1: chars = chars + string.punctuation + r"·！￥……（）【】、：；“‘《》，。？、"

    return json.dumps(''.join(random.choices(chars, k=length)))


def random_int(min_value: int = 0, max_value: int = 10 ** 9) -> int:
    """生成指定范围的随机整数（闭区间）。"""
    if min_value > max_value:
        min_value, max_value = max_value, min_value
    return random.randint(min_value, max_value)


def random_str(
        length: int = 8, kind: str = "alnum", alphabet: Optional[str] = None
) -> str:
    """生成指定长度的随机字符串。
    - kind: digits | letters | alnum | hex
    - alphabet: 若提供，优先使用该字符集
    """
    if length <= 0:
        return ""
    if alphabet:
        pool = alphabet
    else:
        kind = (kind or "alnum").lower()
        if kind == "digits":
            pool = string.digits
        elif kind == "letters":
            pool = string.ascii_letters
        elif kind == "hex":
            pool = "0123456789abcdef"
        else:
            pool = string.ascii_letters + string.digits
    return "".join(random.choice(pool) for _ in range(length)).lower()


# 高频常用汉字表（截取自公开语料的高频字，尽量避免生僻字导致的“口口/方框”显示）
_COMMON_HANZI = (
    "的一是在不了有和人这中大为上个国我以要他时来用们生到作地于出就分对成会可主发年动同工也能下过子说产种面而方后多定行学法所民得经十"
    "三之进等部度家电力里如水化高自二理起小物现实现加量都两体制机当使点从业本去把性好应开它合还因由其些然前外天政四日那主义事平形相全表间样与关各重新线内数正心反你明看原又么利比或但质气第向道路命此变条只没结解决问意建月公无系军很情者最立代想已通并提直题党程展五果料象员革位入常文总次品式活设置及管特件长求老头基资边流路级少图山统接知较将组见计别她手角期根论运农指几九区强放决西被干做必战先回则任选取据处队南给色光门即保治北造百规热领七海口东导器压志世金增争济阶油思术极交受联什认六共权收证改清念建"
)


def random_cn(length: int = 4, charset: str = "common") -> str:
    """生成指定长度的随机中文字符串。
    - charset = "common"（默认）：从常用汉字表中抽取，避免生僻字导致的乱码/方框
    - charset = "basic"：从基本汉字区 U+4E00 - U+9FA5 随机抽取（可能包含生僻字）
    """
    if length <= 0:
        return ""
    pool = None
    if isinstance(charset, str) and charset.lower() == "basic":
        start, end = 0x4E00, 0x9FA5
        return "".join(chr(random.randint(start, end)) for _ in range(length))
    # 默认或未识别：使用常用字表
    pool = _COMMON_HANZI if _COMMON_HANZI else None
    if pool:
        return "".join(random.choice(pool) for _ in range(length))
    # 兜底：basic 区域
    start, end = 0x4E00, 0x9FA5
    return "".join(chr(random.randint(start, end)) for _ in range(length))


def fake_cn_name() -> str:
    """生成中文姓名（依赖 Faker zh_CN，本地降级为 2~3 个常用字）。"""
    if _FAKER_ZH:
        try:
            return _FAKER_ZH.name()
        except Exception:
            pass
    return random_cn(random.randint(2, 3))


def fake_cn_company() -> str:
    """生成中文公司名（依赖 Faker，降级为常用字 + “科技有限公司”）。"""
    if _FAKER_ZH:
        try:
            return _FAKER_ZH.company()
        except Exception:
            pass
    return random_cn(3) + "科技有限公司"


def fake_cn_city() -> str:
    """生成中文城市名。"""
    if _FAKER_ZH:
        try:
            return _FAKER_ZH.city()
        except Exception:
            pass
    return random_cn(2) + "市"


def fake_cn_address() -> str:
    """生成中文地址。"""
    if _FAKER_ZH:
        try:
            return _FAKER_ZH.address()
        except Exception:
            pass
    return fake_cn_city() + random_cn(6) + "路" + str(random_int(1, 999)) + "号"


def fake_cn_sentence(nb_words: int = 8) -> str:
    """生成中文句子。参数 nb_words 仅作近似控制。"""
    if _FAKER_ZH:
        try:
            # Faker 的 sentence 不是严格按字数，这里按词数近似
            return _FAKER_ZH.sentence(nb_words=nb_words)
        except Exception:
            pass
    return random_cn(max(4, nb_words)) + "。"


def fake_cn_paragraph(nb_sentences: int = 3) -> str:
    """生成中文段落。"""
    if _FAKER_ZH:
        try:
            return _FAKER_ZH.paragraph(nb_sentences=nb_sentences)
        except Exception:
            pass
    return "".join(
        fake_cn_sentence(random.randint(6, 12)) for _ in range(max(1, nb_sentences))
    )


def fake_cn_phone() -> str:
    """生成中国大陆手机号。"""
    if _FAKER_ZH:
        try:
            return _FAKER_ZH.phone_number()
        except Exception:
            pass
    # 简易降级：1[3-9]xxxxxxxxx
    return (
            "1"
            + random.choice("3456789")
            + "".join(random.choice(string.digits) for _ in range(9))
    )


def ts_uuid(hex_only: bool = True, length: int = 8) -> str:
    """生成 UUID。默认返回指定长度的 hex，不含连字符。"""
    uuid_str = uuid.uuid4().hex if hex_only else str(uuid.uuid4()).replace('-', '')
    return uuid_str[:length]


def ts_ms() -> int:
    """当前时间戳（毫秒）。"""
    return int(time.time() * 1000)


def ts_s() -> int:
    """当前时间戳（秒）。"""
    return int(time.time())


def base64_encode(input_str: str) -> str:
    """
    对指定字符串进行base64加密

    参数:
        input_str: 需要加密的字符串

    返回:
        加密后的base64字符串
    """
    # 将字符串转换为字节（默认使用utf-8编码）
    input_bytes = input_str.encode("utf-8")

    # 进行base64编码
    encoded_bytes = base64.b64encode(input_bytes)

    # 将字节转换回字符串并返回
    return encoded_bytes.decode("utf-8")


def random_color() -> str:
    """
    返回随机颜色
    """
    colorama = [
        "#08979C",
        "#0E5FC5",
        "#323232",
        "#36CFC9",
        "#3A93FF",
        "#52C41A",
        "#8C8C8C",
        "#9254DE",
        "#A0D911",
        "#EB2F96",
        "#FAAD14",
        "#FADB14",
        "#FF4D4F",
        "#FF7A45",
    ]
    return random.choice(colorama)


def random_icon() -> str:
    """
    返回随机标签
    """
    icons = [
        "icon-dip-shijianchuoxing",
        "icon-dip-riqixing",
        "icon-dip-riqishijianxing",
        "icon-dip-xiaoshuxing",
        "icon-dip-buerxing",
        "icon-dip-wenbenxing",
        "icon-dip-zhengshuxing",
        "icon-dip-erjinzhi",
        "icon-dip-kinship-full",
        "icon-dip-kuaizhao",
        "icon-dip-ziduanmoxing",
        "icon-dip-suoyinguanli",
        "icon-dip-mubiaomoxing",
        "icon-dip-lianlumoxing",
        "icon-dip-zhibiaomoxing",
        "icon-dip-zhishitiaomu",
        "icon-dip-a-shujumoxing1",
        "icon-dip-shujushitu",
        "icon-dip-wendang",
        "icon-dip-suoyin",
        "icon-dip-shujulianjie",
        "icon-dip-cepingzhibiao",
        "icon-dip-cepingshuju",
        "icon-dip-xiaoguoceping",
        "icon-dip-chakanbangdan",
        "icon-dip-tishicikaifa",
        "icon-dip-moxinggongchang",
        "icon-dip-bentiyinqing",
        "icon-dip-chaojizhushou",
        "icon-dip-duomotaishujuhu",
        "icon-dip-zhinengtigongchang",
        "icon-dip-gongju",
        "icon-dip-cebianlan",
        "icon-dip-saomiao1",
        "icon-dip-DATALake",
        "icon-dip-table",
        "icon-dip-echats",
        "icon-dip-chenggong",
        "icon-dip-KG1",
        "icon-dip-data",
        "icon-dip-upload",
        "icon-dip-daochu",
        "icon-dip-history",
        "icon-dip-chat1",
        "icon-dip-gengduo",
        "icon-dip-a-yichu1",
        "icon-dip-tishi",
        "icon-dip-shouquan",
        "icon-dip-bianji",
        "icon-dip-Role",
        "icon-dip-Organization",
        "icon-dip-a-UserGroup",
        "icon-dip-User",
        "icon-dip-a-ApplicationAccount",
        "icon-dip-Department",
        "icon-dip-list",
        "icon-dip-yingyong",
        "icon-dip-copy",
        "icon-dip-export",
        "icon-dip-back",
        "icon-dip-UserAgreement",
        "icon-dip-privacypolicy",
        "icon-dip-help",
        "icon-dip-language-outline",
        "icon-dip-about",
        "icon-dip-think",
        "icon-dip-modelm",
        "icon-dip-lianbangAI",
        "icon-dip-KG",
        "icon-dip-expand",
        "icon-dip-chat2",
        "icon-dip-net",
        "icon-dip-write",
        "icon-dip-doc-qa",
        "icon-dip-task-list",
        "icon-dip-graph-qa",
        "icon-dip-check",
        "icon-dip-sort-descending",
        "icon-dip-right",
        "icon-dip-arrow-up",
        "icon-dip-search",
        "icon-dip-left",
        "icon-dip-filter",
        "icon-dip-close",
        "icon-dip-refresh",
        "icon-dip-calendar",
        "icon-dip-person",
        "icon-dip-add",
        "icon-dip-trash",
        "icon-dip-introduction",
        "icon-dip-move",
        "icon-dip-chat",
        "icon-dip-deep-thinking",
        "icon-dip-attachment",
    ]
    return random.choice(icons)


if __name__ == "__main__":
    dd = random_str()
    print(dd)
