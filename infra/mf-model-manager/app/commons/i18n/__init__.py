from typing import Dict, Optional
from . import en_us, zh_cn


async def get_error_message(code: str, lang) -> str:
    error_messages = {
        "en-us": en_us.error_messages,
        "zh-cn": zh_cn.error_messages,
        "zh-tw": zh_cn.error_messages
    }

    if lang not in error_messages or code not in error_messages[lang]:
        return ""
    return error_messages[lang][code]
