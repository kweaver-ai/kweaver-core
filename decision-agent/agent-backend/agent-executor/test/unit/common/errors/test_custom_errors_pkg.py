"""单元测试 - common/errors/custom_errors_pkg 模块"""


class TestParamError:
    """测试ParamError函数"""

    def test_param_error_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.custom_errors_pkg import ParamError

        error = ParamError()

        # Check it has the expected attributes
        assert hasattr(error, "error_code")
        assert hasattr(error, "description")
        assert hasattr(error, "solution")

    def test_param_error_error_code(self):
        """测试错误码"""
        from app.common.errors.custom_errors_pkg import ParamError

        error = ParamError()

        assert "ParamError" in error.error_code

    def test_param_error_description_exists(self):
        """测试描述存在"""
        from app.common.errors.custom_errors_pkg import ParamError

        error = ParamError()

        assert error.description is not None
        assert len(error.description) > 0

    def test_param_error_solution_exists(self):
        """测试解决方案存在"""
        from app.common.errors.custom_errors_pkg import ParamError

        error = ParamError()

        assert error.solution is not None
        assert len(error.solution) > 0

    def test_param_error_to_dict(self):
        """测试转换为字典"""
        from app.common.errors.custom_errors_pkg import ParamError

        error = ParamError()
        error_dict = error.to_dict()

        assert "ErrorCode" in error_dict
        assert "Description" in error_dict
        assert "Solution" in error_dict


class TestAgentPermissionError:
    """测试AgentPermissionError函数"""

    def test_agent_permission_error_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.custom_errors_pkg import AgentPermissionError

        error = AgentPermissionError()

        assert hasattr(error, "error_code")
        assert hasattr(error, "description")
        assert hasattr(error, "solution")

    def test_agent_permission_error_error_code(self):
        """测试错误码"""
        from app.common.errors.custom_errors_pkg import AgentPermissionError

        error = AgentPermissionError()

        assert "PermissionError" in error.error_code

    def test_agent_permission_error_description(self):
        """测试描述"""
        from app.common.errors.custom_errors_pkg import AgentPermissionError

        error = AgentPermissionError()

        assert "permission" in error.description.lower()

    def test_agent_permission_error_solution(self):
        """测试解决方案"""
        from app.common.errors.custom_errors_pkg import AgentPermissionError

        error = AgentPermissionError()

        assert error.solution is not None

    def test_agent_permission_error_to_dict(self):
        """测试转换为字典"""
        from app.common.errors.custom_errors_pkg import AgentPermissionError

        error = AgentPermissionError()
        error_dict = error.to_dict()

        assert "ErrorCode" in error_dict


class TestDolphinSDKModelError:
    """测试DolphinSDKModelError函数"""

    def test_dolphin_sdk_model_error_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.custom_errors_pkg import DolphinSDKModelError

        error = DolphinSDKModelError()

        assert hasattr(error, "error_code")

    def test_dolphin_sdk_model_error_error_code(self):
        """测试错误码"""
        from app.common.errors.custom_errors_pkg import DolphinSDKModelError

        error = DolphinSDKModelError()

        # Error code contains "Model" (may be "ModelExecption" typo in source)
        assert "Model" in error.error_code

    def test_dolphin_sdk_model_error_has_description(self):
        """测试有描述"""
        from app.common.errors.custom_errors_pkg import DolphinSDKModelError

        error = DolphinSDKModelError()

        assert error.description is not None

    def test_dolphin_sdk_model_error_to_dict(self):
        """测试转换为字典"""
        from app.common.errors.custom_errors_pkg import DolphinSDKModelError

        error = DolphinSDKModelError()
        error_dict = error.to_dict()

        assert isinstance(error_dict, dict)


class TestDolphinSDKSkillError:
    """测试DolphinSDKSkillError函数"""

    def test_dolphin_sdk_skill_error_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.custom_errors_pkg import DolphinSDKSkillError

        error = DolphinSDKSkillError()

        assert hasattr(error, "error_code")

    def test_dolphin_sdk_skill_error_error_code(self):
        """测试错误码"""
        from app.common.errors.custom_errors_pkg import DolphinSDKSkillError

        error = DolphinSDKSkillError()

        # Error code contains "Skill"
        assert "Skill" in error.error_code

    def test_dolphin_sdk_skill_error_description(self):
        """测试描述"""
        from app.common.errors.custom_errors_pkg import DolphinSDKSkillError

        error = DolphinSDKSkillError()

        assert error.description is not None

    def test_dolphin_sdk_skill_error_solution(self):
        """测试解决方案"""
        from app.common.errors.custom_errors_pkg import DolphinSDKSkillError

        error = DolphinSDKSkillError()

        assert error.solution is not None


class TestDolphinSDKBaseError:
    """测试DolphinSDKBaseError函数"""

    def test_dolphin_sdk_base_error_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.custom_errors_pkg import DolphinSDKBaseError

        error = DolphinSDKBaseError()

        assert hasattr(error, "error_code")

    def test_dolphin_sdk_base_error_error_code(self):
        """测试错误码"""
        from app.common.errors.custom_errors_pkg import DolphinSDKBaseError

        error = DolphinSDKBaseError()

        # Error code contains "Base"
        assert "Base" in error.error_code

    def test_dolphin_sdk_base_error_fields(self):
        """测试字段"""
        from app.common.errors.custom_errors_pkg import DolphinSDKBaseError

        error = DolphinSDKBaseError()

        assert error.description is not None
        assert error.solution is not None


class TestConversationRunningError:
    """测试ConversationRunningError函数"""

    def test_conversation_running_error_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.custom_errors_pkg import ConversationRunningError

        error = ConversationRunningError()

        assert hasattr(error, "error_code")

    def test_conversation_running_error_error_code(self):
        """测试错误码"""
        from app.common.errors.custom_errors_pkg import ConversationRunningError

        error = ConversationRunningError()

        # Error code contains "Running"
        assert "Running" in error.error_code

    def test_conversation_running_error_to_dict(self):
        """测试转换为字典"""
        from app.common.errors.custom_errors_pkg import ConversationRunningError

        error = ConversationRunningError()
        error_dict = error.to_dict()

        assert "ErrorCode" in error_dict
        assert "Description" in error_dict
        assert "Solution" in error_dict


class TestCustomErrorsComparison:
    """测试自定义错误比较"""

    def test_different_error_codes(self):
        """测试不同错误码"""
        from app.common.errors.custom_errors_pkg import (
            ParamError,
            AgentPermissionError,
            DolphinSDKModelError,
        )

        param_error = ParamError()
        perm_error = AgentPermissionError()
        model_error = DolphinSDKModelError()

        assert param_error.error_code != perm_error.error_code
        assert perm_error.error_code != model_error.error_code
        assert param_error.error_code != model_error.error_code

    def test_all_have_description(self):
        """测试所有错误都有描述"""
        from app.common.errors.custom_errors_pkg import (
            ParamError,
            AgentPermissionError,
            DolphinSDKModelError,
            DolphinSDKSkillError,
            DolphinSDKBaseError,
            ConversationRunningError,
        )

        errors = [
            ParamError(),
            AgentPermissionError(),
            DolphinSDKModelError(),
            DolphinSDKSkillError(),
            DolphinSDKBaseError(),
            ConversationRunningError(),
        ]

        for error in errors:
            assert error.description is not None
            assert len(error.description) > 0

    def test_all_have_solution(self):
        """测试所有错误都有解决方案"""
        from app.common.errors.custom_errors_pkg import (
            ParamError,
            AgentPermissionError,
            DolphinSDKModelError,
            DolphinSDKSkillError,
            DolphinSDKBaseError,
            ConversationRunningError,
        )

        errors = [
            ParamError(),
            AgentPermissionError(),
            DolphinSDKModelError(),
            DolphinSDKSkillError(),
            DolphinSDKBaseError(),
            ConversationRunningError(),
        ]

        for error in errors:
            assert error.solution is not None
            assert len(error.solution) > 0

    def test_all_to_dict_works(self):
        """测试所有错误可以转换为字典"""
        from app.common.errors.custom_errors_pkg import (
            ParamError,
            AgentPermissionError,
            DolphinSDKModelError,
            DolphinSDKSkillError,
            DolphinSDKBaseError,
            ConversationRunningError,
        )

        errors = [
            ParamError(),
            AgentPermissionError(),
            DolphinSDKModelError(),
            DolphinSDKSkillError(),
            DolphinSDKBaseError(),
            ConversationRunningError(),
        ]

        for error in errors:
            error_dict = error.to_dict()
            assert isinstance(error_dict, dict)
            assert "ErrorCode" in error_dict


class TestCustomErrorsStr:
    """测试自定义错误字符串表示"""

    def test_param_error_str(self):
        """测试ParamError字符串"""
        from app.common.errors.custom_errors_pkg import ParamError

        error = ParamError()
        error_str = str(error)

        assert isinstance(error_str, str)

    def test_agent_permission_error_str(self):
        """测试AgentPermissionError字符串"""
        from app.common.errors.custom_errors_pkg import AgentPermissionError

        error = AgentPermissionError()
        error_str = str(error)

        assert isinstance(error_str, str)

    def test_dolphin_sdk_model_error_str(self):
        """测试DolphinSDKModelError字符串"""
        from app.common.errors.custom_errors_pkg import DolphinSDKModelError

        error = DolphinSDKModelError()
        error_str = str(error)

        assert isinstance(error_str, str)


class TestCustomErrorsRepr:
    """测试自定义错误repr"""

    def test_param_error_repr(self):
        """测试ParamError repr"""
        from app.common.errors.custom_errors_pkg import ParamError

        error = ParamError()
        error_repr = repr(error)

        assert isinstance(error_repr, str)
        assert "Error(error_code=" in error_repr

    def test_agent_permission_error_repr(self):
        """测试AgentPermissionError repr"""
        from app.common.errors.custom_errors_pkg import AgentPermissionError

        error = AgentPermissionError()
        error_repr = repr(error)

        assert isinstance(error_repr, str)

    def test_all_errors_repr_format(self):
        """测试所有错误repr格式"""
        from app.common.errors.custom_errors_pkg import (
            ParamError,
            AgentPermissionError,
            DolphinSDKModelError,
            DolphinSDKSkillError,
            DolphinSDKBaseError,
            ConversationRunningError,
        )

        errors = [
            ParamError(),
            AgentPermissionError(),
            DolphinSDKModelError(),
            DolphinSDKSkillError(),
            DolphinSDKBaseError(),
            ConversationRunningError(),
        ]

        for error in errors:
            error_repr = repr(error)
            assert "Error(error_code=" in error_repr


class TestCustomErrorsPackage:
    """测试自定义错误包"""

    def test_import_all_errors(self):
        """测试导入所有错误"""
        from app.common.errors.custom_errors_pkg import (
            ParamError,
            AgentPermissionError,
            DolphinSDKModelError,
            DolphinSDKSkillError,
            DolphinSDKBaseError,
            ConversationRunningError,
        )

        assert callable(ParamError)
        assert callable(AgentPermissionError)
        assert callable(DolphinSDKModelError)
        assert callable(DolphinSDKSkillError)
        assert callable(DolphinSDKBaseError)
        assert callable(ConversationRunningError)

    def test_package_all_list(self):
        """测试包的__all__列表"""
        from app.common.errors.custom_errors_pkg import __all__

        assert "ParamError" in __all__
        assert "AgentPermissionError" in __all__
        assert "DolphinSDKModelError" in __all__
        assert "DolphinSDKSkillError" in __all__
        assert "DolphinSDKBaseError" in __all__
        assert "ConversationRunningError" in __all__

    def test_all_exported_errors_work(self):
        """测试所有导出的错误都可以工作"""
        from app.common.errors.custom_errors_pkg import __all__

        for error_name in __all__:
            error_func = getattr(
                __import__(
                    "app.common.errors.custom_errors_pkg", fromlist=[error_name]
                ),
                error_name,
            )
            error = error_func()
            assert hasattr(error, "error_code")
            assert hasattr(error, "description")
            assert hasattr(error, "solution")


class TestCustomErrorsRoundTrip:
    """测试往返转换"""

    def test_param_error_roundtrip(self):
        """测试ParamError往返"""
        from app.common.errors.api_error_class import APIError
        from app.common.errors.custom_errors_pkg import ParamError

        original = ParamError()
        error_dict = original.to_dict()
        restored = APIError.from_dict(error_dict)

        assert restored.error_code == original.error_code

    def test_agent_permission_error_roundtrip(self):
        """测试AgentPermissionError往返"""
        from app.common.errors.api_error_class import APIError
        from app.common.errors.custom_errors_pkg import AgentPermissionError

        original = AgentPermissionError()
        error_dict = original.to_dict()
        restored = APIError.from_dict(error_dict)

        assert restored.error_code == original.error_code
