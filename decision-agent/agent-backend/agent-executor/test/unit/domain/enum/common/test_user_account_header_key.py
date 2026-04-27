"""单元测试 - domain/enum/common/user_account_header_key 模块"""


class TestUserAccountHeaderKey:
    """测试 UserAccountHeaderKey 枚举"""

    def test_enum_values(self):
        """测试枚举值"""
        from app.domain.enum.common.user_account_header_key import UserAccountHeaderKey

        assert UserAccountHeaderKey.ACCOUNT_ID.value == "x-account-id"
        assert UserAccountHeaderKey.ACCOUNT_TYPE.value == "x-account-type"
        assert UserAccountHeaderKey.BIZ_DOMAIN_ID.value == "x-business-domain"
        assert UserAccountHeaderKey.ACCOUNT_ID_OLD.value == "x-user"
        assert UserAccountHeaderKey.ACCOUNT_TYPE_OLD.value == "x-visitor-type"

    def test_enum_is_string_enum(self):
        """测试枚举是字符串枚举"""
        from app.domain.enum.common.user_account_header_key import UserAccountHeaderKey

        assert isinstance(UserAccountHeaderKey.ACCOUNT_ID.value, str)
        assert isinstance(UserAccountHeaderKey.ACCOUNT_ID, str)


class TestGetUserAccountId:
    """测试 get_user_account_id 函数"""

    def test_get_new_account_id(self):
        """测试获取新格式的account_id"""
        from app.domain.enum.common.user_account_header_key import get_user_account_id

        headers = {"x-account-id": "user123"}
        result = get_user_account_id(headers)

        assert result == "user123"

    def test_get_old_account_id_when_new_missing(self):
        """测试新格式不存在时返回旧格式"""
        from app.domain.enum.common.user_account_header_key import get_user_account_id

        headers = {"x-user": "user456"}
        result = get_user_account_id(headers)

        assert result == "user456"

    def test_get_new_account_id_preferred(self):
        """测试新格式优先于旧格式"""
        from app.domain.enum.common.user_account_header_key import get_user_account_id

        headers = {"x-account-id": "new123", "x-user": "old456"}
        result = get_user_account_id(headers)

        assert result == "new123"


class TestGetUserAccountType:
    """测试 get_user_account_type 函数"""

    def test_get_new_account_type(self):
        """测试获取新格式的account_type"""
        from app.domain.enum.common.user_account_header_key import get_user_account_type

        headers = {"x-account-type": "app"}
        result = get_user_account_type(headers)

        assert result == "app"

    def test_get_old_account_type_when_new_missing(self):
        """测试新格式不存在时返回旧格式"""
        from app.domain.enum.common.user_account_header_key import get_user_account_type

        headers = {"x-visitor-type": "user"}
        result = get_user_account_type(headers)

        assert result == "user"

    def test_get_new_account_type_preferred(self):
        """测试新格式优先于旧格式"""
        from app.domain.enum.common.user_account_header_key import get_user_account_type

        headers = {"x-account-type": "app", "x-visitor-type": "user"}
        result = get_user_account_type(headers)

        assert result == "app"


class TestGetUserAccount:
    """测试 get_user_account 函数"""

    def test_get_user_account_tuple(self):
        """测试获取用户账号元组"""
        from app.domain.enum.common.user_account_header_key import get_user_account

        headers = {"x-account-id": "user123", "x-account-type": "user"}
        account_id, account_type = get_user_account(headers)

        assert account_id == "user123"
        assert account_type == "user"


class TestGetBizDomainId:
    """测试 get_biz_domain_id 函数"""

    def test_get_biz_domain_id_existing(self):
        """测试获取存在的业务域ID"""
        from app.domain.enum.common.user_account_header_key import get_biz_domain_id

        headers = {"x-business-domain": "domain123"}
        result = get_biz_domain_id(headers)

        assert result == "domain123"

    def test_get_biz_domain_id_missing(self):
        """测试获取不存在的业务域ID返回空字符串"""
        from app.domain.enum.common.user_account_header_key import get_biz_domain_id

        headers = {}
        result = get_biz_domain_id(headers)

        assert result == ""


class TestSetUserAccount:
    """测试 set_user_account 函数"""

    def test_set_user_account(self):
        """测试设置用户账号"""
        from app.domain.enum.common.user_account_header_key import set_user_account

        headers = {}
        set_user_account(headers, "user123", "user")

        assert headers["x-account-id"] == "user123"
        assert headers["x-account-type"] == "user"
        assert headers["x-user"] == "user123"
        assert headers["x-visitor-type"] == "user"


class TestSetUserAccountId:
    """测试 set_user_account_id 函数"""

    def test_set_user_account_id(self):
        """测试设置用户账号ID"""
        from app.domain.enum.common.user_account_header_key import set_user_account_id

        headers = {}
        set_user_account_id(headers, "user123")

        assert headers["x-account-id"] == "user123"
        assert headers["x-user"] == "user123"


class TestSetUserAccountType:
    """测试 set_user_account_type 函数"""

    def test_set_user_account_type(self):
        """测试设置用户账号类型"""
        from app.domain.enum.common.user_account_header_key import set_user_account_type

        headers = {}
        set_user_account_type(headers, "app")

        assert headers["x-account-type"] == "app"
        assert headers["x-visitor-type"] == "app"


class TestSetBizDomainId:
    """测试 set_biz_domain_id 函数"""

    def test_set_biz_domain_id(self):
        """测试设置业务域ID"""
        from app.domain.enum.common.user_account_header_key import set_biz_domain_id

        headers = {}
        set_biz_domain_id(headers, "domain123")

        assert headers["x-business-domain"] == "domain123"


class TestHasUserAccount:
    """测试 has_user_account 函数"""

    def test_has_user_account_with_new_key(self):
        """测试新格式存在时返回True"""
        from app.domain.enum.common.user_account_header_key import has_user_account

        headers = {"x-account-id": "user123"}
        result = has_user_account(headers)

        assert result

    def test_has_user_account_with_old_key(self):
        """测试旧格式存在时返回True"""
        from app.domain.enum.common.user_account_header_key import has_user_account

        headers = {"x-user": "user456"}
        result = has_user_account(headers)

        assert result

    def test_has_user_account_with_both_keys(self):
        """测试两种格式都存在时返回True"""
        from app.domain.enum.common.user_account_header_key import has_user_account

        headers = {"x-account-id": "new123", "x-user": "old456"}
        result = has_user_account(headers)

        assert result

    def test_has_user_account_missing(self):
        """测试账号不存在时返回False"""
        from app.domain.enum.common.user_account_header_key import has_user_account

        headers = {}
        result = has_user_account(headers)

        assert not result


class TestHasUserAccountType:
    """测试 has_user_account_type 函数"""

    def test_has_user_account_type_with_new_key(self):
        """测试新格式存在时返回True"""
        from app.domain.enum.common.user_account_header_key import has_user_account_type

        headers = {"x-account-type": "app"}
        result = has_user_account_type(headers)

        assert result

    def test_has_user_account_type_with_old_key(self):
        """测试旧格式存在时返回True"""
        from app.domain.enum.common.user_account_header_key import has_user_account_type

        headers = {"x-visitor-type": "user"}
        result = has_user_account_type(headers)

        assert result

    def test_has_user_account_type_with_both_keys(self):
        """测试两种格式都存在时返回True"""
        from app.domain.enum.common.user_account_header_key import has_user_account_type

        headers = {"x-account-type": "app", "x-visitor-type": "user"}
        result = has_user_account_type(headers)

        assert result

    def test_has_user_account_type_missing(self):
        """测试账号类型不存在时返回False"""
        from app.domain.enum.common.user_account_header_key import has_user_account_type

        headers = {}
        result = has_user_account_type(headers)

        assert not result
