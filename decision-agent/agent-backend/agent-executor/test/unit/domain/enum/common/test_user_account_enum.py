"""单元测试 - domain/enum/common 模块"""

from unittest import TestCase

from app.domain.enum.common.user_account_type import (
    UserAccountType,
    check_user_account_type,
)
from app.domain.enum.common.user_account_header_key import (
    UserAccountHeaderKey,
    get_user_account_id,
    get_user_account_type,
    get_user_account,
    get_biz_domain_id,
    set_user_account,
    set_user_account_id,
    set_user_account_type,
    set_biz_domain_id,
    has_user_account,
    has_user_account_type,
)


class TestUserAccountType(TestCase):
    """测试 UserAccountType 枚举"""

    def test_app_value(self):
        """测试APP账号类型"""
        self.assertEqual(UserAccountType.APP.value, "app")

    def test_user_value(self):
        """测试USER账号类型"""
        self.assertEqual(UserAccountType.USER.value, "user")

    def test_anonymous_value(self):
        """测试ANONYMOUS账号类型"""
        self.assertEqual(UserAccountType.ANONYMOUS.value, "anonymous")

    def test_check_user_account_type_app(self):
        """测试检查APP类型"""
        self.assertTrue(check_user_account_type("app"))

    def test_check_user_account_type_user(self):
        """测试检查USER类型"""
        self.assertTrue(check_user_account_type("user"))

    def test_check_user_account_type_invalid(self):
        """测试检查无效类型"""
        self.assertFalse(check_user_account_type("invalid"))

    def test_check_user_account_type_anonymous(self):
        """测试检查ANONYMOUS类型"""
        self.assertFalse(check_user_account_type("anonymous"))


class TestUserAccountHeaderKey(TestCase):
    """测试 UserAccountHeaderKey 枚举和函数"""

    def test_account_id_value(self):
        """测试ACCOUNT_ID"""
        self.assertEqual(UserAccountHeaderKey.ACCOUNT_ID.value, "x-account-id")

    def test_account_type_value(self):
        """测试ACCOUNT_TYPE"""
        self.assertEqual(UserAccountHeaderKey.ACCOUNT_TYPE.value, "x-account-type")

    def test_biz_domain_id_value(self):
        """测试BIZ_DOMAIN_ID"""
        self.assertEqual(UserAccountHeaderKey.BIZ_DOMAIN_ID.value, "x-business-domain")

    def test_account_id_old_value(self):
        """测试ACCOUNT_ID_OLD"""
        self.assertEqual(UserAccountHeaderKey.ACCOUNT_ID_OLD.value, "x-user")

    def test_account_type_old_value(self):
        """测试ACCOUNT_TYPE_OLD"""
        self.assertEqual(UserAccountHeaderKey.ACCOUNT_TYPE_OLD.value, "x-visitor-type")

    def test_get_user_account_id_new(self):
        """测试获取新版本账号ID"""
        headers = {"x-account-id": "test_id"}
        result = get_user_account_id(headers)
        self.assertEqual(result, "test_id")

    def test_get_user_account_id_old(self):
        """测试获取旧版本账号ID"""
        headers = {"x-user": "test_id"}
        result = get_user_account_id(headers)
        self.assertEqual(result, "test_id")

    def test_get_user_account_id_both(self):
        """测试新旧都存在时优先新版本"""
        headers = {"x-account-id": "new_id", "x-user": "old_id"}
        result = get_user_account_id(headers)
        self.assertEqual(result, "new_id")

    def test_get_user_account_type_new(self):
        """测试获取新版本账号类型"""
        headers = {"x-account-type": "app"}
        result = get_user_account_type(headers)
        self.assertEqual(result, "app")

    def test_get_user_account_type_old(self):
        """测试获取旧版本账号类型"""
        headers = {"x-visitor-type": "user"}
        result = get_user_account_type(headers)
        self.assertEqual(result, "user")

    def test_get_user_account_type_both(self):
        """测试新旧都存在时优先新版本"""
        headers = {"x-account-type": "app", "x-visitor-type": "user"}
        result = get_user_account_type(headers)
        self.assertEqual(result, "app")

    def test_get_user_account(self):
        """测试获取用户账号信息"""
        headers = {"x-account-id": "test_id", "x-account-type": "app"}
        account_id, account_type = get_user_account(headers)
        self.assertEqual(account_id, "test_id")
        self.assertEqual(account_type, "app")

    def test_get_biz_domain_id(self):
        """测试获取业务域ID"""
        headers = {"x-business-domain": "domain_123"}
        result = get_biz_domain_id(headers)
        self.assertEqual(result, "domain_123")

    def test_get_biz_domain_id_missing(self):
        """测试获取业务域ID不存在"""
        headers = {}
        result = get_biz_domain_id(headers)
        self.assertEqual(result, "")

    def test_set_user_account(self):
        """测试设置用户账号信息"""
        headers = {}
        set_user_account(headers, "test_id", "app")
        self.assertEqual(headers["x-account-id"], "test_id")
        self.assertEqual(headers["x-account-type"], "app")
        self.assertEqual(headers["x-user"], "test_id")
        self.assertEqual(headers["x-visitor-type"], "app")

    def test_set_user_account_id(self):
        """测试设置用户账号ID"""
        headers = {}
        set_user_account_id(headers, "test_id")
        self.assertEqual(headers["x-account-id"], "test_id")
        self.assertEqual(headers["x-user"], "test_id")

    def test_set_user_account_type(self):
        """测试设置用户账号类型"""
        headers = {}
        set_user_account_type(headers, "app")
        self.assertEqual(headers["x-account-type"], "app")
        self.assertEqual(headers["x-visitor-type"], "app")

    def test_set_biz_domain_id(self):
        """测试设置业务域ID"""
        headers = {}
        set_biz_domain_id(headers, "domain_123")
        self.assertEqual(headers["x-business-domain"], "domain_123")

    def test_has_user_account_new(self):
        """测试是否有新版本用户账号"""
        headers = {"x-account-id": "test_id"}
        result = has_user_account(headers)
        self.assertTrue(result)

    def test_has_user_account_old(self):
        """测试是否有旧版本用户账号"""
        headers = {"x-user": "test_id"}
        result = has_user_account(headers)
        self.assertTrue(result)

    def test_has_user_account_none(self):
        """测试没有用户账号"""
        headers = {}
        result = has_user_account(headers)
        self.assertFalse(result)

    def test_has_user_account_type_new(self):
        """测试是否有新版本账号类型"""
        headers = {"x-account-type": "app"}
        result = has_user_account_type(headers)
        self.assertTrue(result)

    def test_has_user_account_type_old(self):
        """测试是否有旧版本账号类型"""
        headers = {"x-visitor-type": "user"}
        result = has_user_account_type(headers)
        self.assertTrue(result)

    def test_has_user_account_type_none(self):
        """测试没有账号类型"""
        headers = {}
        result = has_user_account_type(headers)
        self.assertFalse(result)
