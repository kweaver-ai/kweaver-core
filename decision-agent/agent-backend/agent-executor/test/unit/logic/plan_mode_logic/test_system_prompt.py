"""单元测试 - logic/plan_mode_logic 模块"""


class TestGetSystemPromptWithPlan:
    """测试 get_system_prompt_with_plan 函数"""

    def test_with_empty_user_system_prompt(self):
        """测试空用户系统提示"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        result = get_system_prompt_with_plan("")

        assert "你是一个智能任务管理Agent" in result
        assert "工作流程：" in result

    def test_with_whitespace_only_user_system_prompt(self):
        """测试仅空白字符的用户系统提示"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        result = get_system_prompt_with_plan("   \n\t  ")

        assert "你是一个智能任务管理Agent" in result
        assert "其他要求：" not in result

    def test_with_none_user_system_prompt(self):
        """测试None用户系统提示"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        result = get_system_prompt_with_plan(None)

        assert "你是一个智能任务管理Agent" in result
        assert "其他要求：" not in result

    def test_with_valid_user_system_prompt(self):
        """测试有效用户系统提示"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = "请使用中文回答"
        result = get_system_prompt_with_plan(user_prompt)

        assert "你是一个智能任务管理Agent" in result
        assert "其他要求：" in result
        assert "请使用中文回答" in result

    def test_with_multiline_user_system_prompt(self):
        """测试多行用户系统提示"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = """
        第一条要求
        第二条要求
        """
        result = get_system_prompt_with_plan(user_prompt)

        assert "你是一个智能任务管理Agent" in result
        assert "其他要求：" in result
        assert "第一条要求" in result
        assert "第二条要求" in result

    def test_base_prompt_content(self):
        """测试基础提示内容"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        result = get_system_prompt_with_plan("")

        # Check for key sections in the plan system prompt
        assert "任务规划阶段" in result
        assert "任务执行阶段" in result
        assert "任务完成判断" in result
        assert "任务总结" in result
        assert "Task_Plan_Agent" in result

    def test_user_prompt_appended_after_base(self):
        """测试用户提示追加到基础提示之后"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        base_only = get_system_prompt_with_plan("")
        with_user = get_system_prompt_with_plan("额外要求")

        # The with_user version should start with the base prompt
        assert with_user.startswith(base_only.rstrip("\n"))
        # And should have extra requirements at the end
        assert "额外要求" in with_user

    def test_with_special_characters(self):
        """测试特殊字符"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = "Test !@#$%^&*()_+ {}|:<>?[]"
        result = get_system_prompt_with_plan(user_prompt)
        assert user_prompt in result

    def test_with_unicode(self):
        """测试unicode字符"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = "测试中文🎉 emoji"
        result = get_system_prompt_with_plan(user_prompt)
        assert user_prompt in result

    def test_with_tabs(self):
        """测试制表符"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = "\t\tIndented text\t\t"
        result = get_system_prompt_with_plan(user_prompt)
        assert "Indented text" in result

    def test_with_leading_trailing_spaces(self):
        """测试前后空格"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = "   leading and trailing   "
        result = get_system_prompt_with_plan(user_prompt)
        assert "leading and trailing" in result

    def test_returns_string(self):
        """测试返回字符串"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        result = get_system_prompt_with_plan(None)
        assert isinstance(result, str)

    def test_multiple_calls_same_result(self):
        """测试多次调用产生相同结果"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = "Test prompt"
        result1 = get_system_prompt_with_plan(user_prompt)
        result2 = get_system_prompt_with_plan(user_prompt)
        assert result1 == result2

    def test_with_markdown(self):
        """测试markdown格式"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = "# Header\n## Subheader\n- List item"
        result = get_system_prompt_with_plan(user_prompt)
        assert "# Header" in result
        assert "## Subheader" in result
        assert "- List item" in result

    def test_with_json(self):
        """测试JSON格式"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = '{"key": "value", "number": 123}'
        result = get_system_prompt_with_plan(user_prompt)
        assert user_prompt in result

    def test_with_code_snippet(self):
        """测试代码片段"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = "```python\ndef test():\n    pass\n```"
        result = get_system_prompt_with_plan(user_prompt)
        assert "def test():" in result

    def test_with_zero_length_string(self):
        """测试零长度字符串"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        result = get_system_prompt_with_plan("")
        assert "你是一个智能任务管理Agent" in result

    def test_with_single_space(self):
        """测试单个空格"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        result = get_system_prompt_with_plan(" ")
        assert "你是一个智能任务管理Agent" in result

    def test_with_only_newlines(self):
        """测试仅有换行符"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        result = get_system_prompt_with_plan("\n\n\n")
        assert "你是一个智能任务管理Agent" in result

    def test_with_mixed_whitespace(self):
        """测试混合空白字符"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        result = get_system_prompt_with_plan("  \n\t  \n  ")
        assert "你是一个智能任务管理Agent" in result

    def test_with_very_long(self):
        """测试非常长的提示"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = "A" * 10000
        result = get_system_prompt_with_plan(user_prompt)
        assert user_prompt in result

    def test_with_very_short(self):
        """测试非常短的提示"""
        from app.logic.plan_mode_logic import get_system_prompt_with_plan

        user_prompt = "Hi"
        result = get_system_prompt_with_plan(user_prompt)
        assert "Hi" in result
