from gettext import gettext as _l
from typing import Dict

from jsonschema import ValidationError

from app.utils.common import get_lang


class RegexPatterns:
    Chinese_and_English_numbers_and_underline = "^[_a-zA-Z0-9\u4e00-\u9fa5]+$"
    Chinese_and_English_numbers_and_special_symbols_on_the_keyboard = "^[-_a-zA-Z0-9 =~!@#$&%^&*'\"“、‘()_+`{}\\\;:,.·?<>|/~！$\s/\n\\\s@#￥%…&*“”（）—+。={}|【】：；、《》？，。\u4e00-\u9fa5]+$"
    Chinese_and_English_numbers_and_special_symbols_on_the_keyboard_allow_empty = "^([-_a-zA-Z0-9 =~!@#$&%^&*'\"“、‘()_+`{}\\\;:,.·?<>|/~！$\s/\n\\\s@#￥%…&*“”（）—+。={}|【】：；、《》？，。\u4e00-\u9fa5]+)?$"
    Chinese_and_English_numbers_and_some_keyboard_symbols_excluding_special_chars = "^[-_a-zA-Z0-9 =~!@$&%^&'“、‘()_+`{}\\\;,.·~！$\s\n\s@￥%…&“”（）—+。={}【】：；、《》？，。\u4e00-\u9fa5]+$"
    Positive_integer = "^[1-9][0-9]*$"
    Positive_integer_with_minus_1 = r"^(-1|[1-9]\d*)$"
    Positive_integer_with_0 = r"^(\d*)$"
    English_numbers_and_hyphen = "^[a-zA-Z0-9-]+$"
    oss_id_pattern_allow_empty = "^(AD-\d{19}-\d{19})?$"
    snow_id_pattern = "^\d{19}$"
    snow_id_pattern_allow_empty = "^(\d{19})?$"
    uuid_pattern = "^[0-9a-fA-F]{32}$"
    Variable_in_curly_braces = r"\{{(.*?)\}}"
    """
    变量使用花括号 `{{}}` 表示
    """
    Simple_variable_with_dollar_sign = r"\$([a-zA-Z_][a-zA-Z0-9_]*)"
    """
    变量以`$`开头，只匹配简单变量: `$变量名`
    """
    Complex_variable_with_dollar_sign = (
        r"\$([a-zA-Z_][a-zA-Z0-9_]*(?:(?:\.[a-zA-Z_][a-zA-Z0-9_]*)|(?:\[\d+\]))*)"
    )
    """
    变量以`$`开头，支持以下形式:
    简单变量: `$变量名`
    数组索引: `$变量名[index]`
    嵌套属性: `$变量名.key1.key2`
    示例:
    ```
    $x
    $result[0]
    $a.b.c
    ```
    """


regexErrorMessage: Dict[str, str] = {
    RegexPatterns.Chinese_and_English_numbers_and_underline: _l(
        " must be Chinese and English, numbers and underline!"
    ),
    RegexPatterns.Chinese_and_English_numbers_and_special_symbols_on_the_keyboard: _l(
        " must be Chinese and English, numbers and special symbols on the keyboard and cannot be empty!"
    ),
    RegexPatterns.Chinese_and_English_numbers_and_special_symbols_on_the_keyboard_allow_empty: _l(
        " must be Chinese and English, numbers and special symbols on the keyboard!"
    ),
    RegexPatterns.Chinese_and_English_numbers_and_some_keyboard_symbols_excluding_special_chars: _l(
        " must be Chinese and English, numbers and some keyboard symbols, excluding '#' '/' ':' '*' '?' '\"' '<' '>' '|'!"
    ),
    RegexPatterns.Positive_integer: _l(" must be positive integer!"),
    RegexPatterns.Positive_integer_with_minus_1: _l(" must be positive integer or -1!"),
    RegexPatterns.Positive_integer_with_0: _l(" must be positive integer or 0!"),
    RegexPatterns.English_numbers_and_hyphen: _l(
        " must be English, numbers and hyphen!"
    ),
    RegexPatterns.oss_id_pattern_allow_empty: _l(" must be oss_id pattern!"),
    RegexPatterns.snow_id_pattern: _l(
        " must be a 19 digit number!"
    ),  # 必须是19位的数字
    RegexPatterns.snow_id_pattern_allow_empty: _l(
        " must be a 19 digit number!"
    ),  # 必须是19位的数字
    RegexPatterns.uuid_pattern: _l(" must be uuid pattern!"),
}


def GetErrorMessageByRegex(key: str) -> str:
    if key in regexErrorMessage:
        return regexErrorMessage[key]
    return _l(" is invalid!")


def handleJsonSchemaError(exc: ValidationError) -> str:
    """对jsonschema校验的异常处理报错语句"""
    _l = get_lang()
    errorMessage = exc.message
    paramName = ""  # 出错的字段
    if len(exc.absolute_path) > 0:
        paramName = exc.absolute_path[-1]
        # 如果是array类型，则paramName有可能是数组序号
        i = -1
        while isinstance(paramName, int):
            i -= 1
            try:
                paramName = exc.absolute_path[i]
                paramName = _l(" parameter: ") + paramName
            except IndexError:
                paramName = ""
                break
        errorMessage = "{} {}".format(paramName, errorMessage)
    error_type = ""
    if len(exc.absolute_schema_path) > 0:
        error_type = exc.absolute_schema_path[-1]
    if error_type == "required":
        paramName = ", ".join(set(exc.validator_value) - set(exc.instance.keys()))
        errorMessage = _l(" parameter: ") + paramName + _l(" is required!")
    elif error_type == "type":
        if exc.validator_value == "string":
            errorMessage = paramName + _l(" must be a string type!")
        elif exc.validator_value == "integer":
            errorMessage = paramName + _l(" must be an int type!")
        elif exc.validator_value == "float":
            errorMessage = paramName + _l(" must be a float type!")
        elif exc.validator_value == "array":
            errorMessage = paramName + _l(" must be a list!")
        elif exc.validator_value == "object":
            errorMessage = paramName + _l(" must be a dict!")
        elif exc.validator_value == "boolean":
            errorMessage = paramName + _l(" must be a bool type!")
    elif error_type in ["minLength", "minItems"]:
        errorMessage = (
            paramName + _l(" length is at least ") + str(exc.validator_value) + _l("!")
        )
        if len(exc.instance) == 0:
            errorMessage = paramName + _l(" cannot be empty. ")
    elif error_type in ["maxLength", "maxItems"]:
        errorMessage = (
            paramName + _l(" length is up to ") + str(exc.validator_value) + _l("!")
        )
    elif error_type == "pattern":
        errorMessage = paramName + _l(GetErrorMessageByRegex(str(exc.validator_value)))
    elif error_type == "uniqueItems":
        errorMessage = paramName + _l(" has duplicated items!")
    elif error_type == "enum":
        errorMessage = paramName + _l(" must be in {}!").format(
            str(exc.validator_value)
        )
    return errorMessage
