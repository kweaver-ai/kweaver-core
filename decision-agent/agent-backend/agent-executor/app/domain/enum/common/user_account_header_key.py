from enum import Enum


class UserAccountHeaderKey(str, Enum):
    """用户账号header键"""

    ACCOUNT_ID = "x-account-id"  # 账号id [当account_type=user时，account_id为用户id；当account_type=app时，account_id为应用id]
    ACCOUNT_TYPE = "x-account-type"  # 账号类型[app-应用账号, user-普通用户,anonymous-匿名用户（暂不支持）]
    BIZ_DOMAIN_ID = "x-business-domain"  # 业务域id

    ACCOUNT_ID_OLD = "x-user"  # 账号id
    ACCOUNT_TYPE_OLD = "x-visitor-type"  # 账号类型


# ---get---start---


def get_user_account_id(headers: dict) -> str:
    if headers.get(UserAccountHeaderKey.ACCOUNT_ID.value):
        return headers[UserAccountHeaderKey.ACCOUNT_ID.value]
    return headers[UserAccountHeaderKey.ACCOUNT_ID_OLD.value]


def get_user_account_type(headers: dict) -> str:
    if headers.get(UserAccountHeaderKey.ACCOUNT_TYPE.value):
        return headers[UserAccountHeaderKey.ACCOUNT_TYPE.value]
    return headers[UserAccountHeaderKey.ACCOUNT_TYPE_OLD.value]


def get_user_account(headers: dict) -> tuple[str, str]:
    return get_user_account_id(headers), get_user_account_type(headers)


def get_biz_domain_id(headers: dict) -> str:
    return headers.get(UserAccountHeaderKey.BIZ_DOMAIN_ID.value, "")


# ---get---end---


# ---set---start---
def set_user_account(headers: dict, account_id: str, account_type: str) -> None:
    headers[UserAccountHeaderKey.ACCOUNT_ID.value] = account_id
    headers[UserAccountHeaderKey.ACCOUNT_TYPE.value] = account_type

    headers[UserAccountHeaderKey.ACCOUNT_ID_OLD.value] = account_id
    headers[UserAccountHeaderKey.ACCOUNT_TYPE_OLD.value] = account_type


def set_user_account_id(headers: dict, account_id: str) -> None:
    headers[UserAccountHeaderKey.ACCOUNT_ID.value] = account_id
    headers[UserAccountHeaderKey.ACCOUNT_ID_OLD.value] = account_id


def set_user_account_type(headers: dict, account_type: str) -> None:
    headers[UserAccountHeaderKey.ACCOUNT_TYPE.value] = account_type
    headers[UserAccountHeaderKey.ACCOUNT_TYPE_OLD.value] = account_type


def set_biz_domain_id(headers: dict, biz_domain_id: str) -> None:
    headers[UserAccountHeaderKey.BIZ_DOMAIN_ID.value] = biz_domain_id


# ---set---end---


# ---has---start---
def has_user_account(headers: dict) -> bool:
    return headers.get(UserAccountHeaderKey.ACCOUNT_ID.value) or headers.get(
        UserAccountHeaderKey.ACCOUNT_ID_OLD.value
    )


def has_user_account_type(headers: dict) -> bool:
    return headers.get(UserAccountHeaderKey.ACCOUNT_TYPE.value) or headers.get(
        UserAccountHeaderKey.ACCOUNT_TYPE_OLD.value
    )


# ---has---end---
