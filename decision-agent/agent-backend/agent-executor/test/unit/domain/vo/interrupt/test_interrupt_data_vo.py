"""单元测试 - domain/vo/interrupt/interrupt_data 模块"""

from unittest import TestCase

from app.domain.vo.interrupt.interrupt_data import (
    ToolArg,
    InterruptConfig,
    InterruptData,
)


class TestToolArg(TestCase):
    """测试 ToolArg 类"""

    def test_init(self):
        """测试初始化"""
        arg = ToolArg(key="param1", value="value1", type="string")
        self.assertEqual(arg.key, "param1")
        self.assertEqual(arg.value, "value1")
        self.assertEqual(arg.type, "string")

    def test_model_dump(self):
        """测试序列化"""
        arg = ToolArg(key="param1", value="value1", type="string")
        data = arg.model_dump()
        self.assertEqual(data["key"], "param1")
        self.assertEqual(data["value"], "value1")
        self.assertEqual(data["type"], "string")

    def test_init_with_complex_value(self):
        """测试复杂值类型"""
        arg = ToolArg(key="param1", value={"nested": "value"}, type="object")
        self.assertEqual(arg.value, {"nested": "value"})


class TestInterruptConfig(TestCase):
    """测试 InterruptConfig 类"""

    def test_init_true(self):
        """测试需要确认的配置"""
        config = InterruptConfig(
            requires_confirmation=True,
            confirmation_message="Please confirm",
        )
        self.assertTrue(config.requires_confirmation)
        self.assertEqual(config.confirmation_message, "Please confirm")

    def test_init_false(self):
        """测试不需要确认的配置"""
        config = InterruptConfig(requires_confirmation=False, confirmation_message="")
        self.assertFalse(config.requires_confirmation)
        self.assertEqual(config.confirmation_message, "")

    def test_model_dump(self):
        """测试序列化"""
        config = InterruptConfig(
            requires_confirmation=True, confirmation_message="Confirm?"
        )
        data = config.model_dump()
        self.assertTrue(data["requires_confirmation"])
        self.assertEqual(data["confirmation_message"], "Confirm?")

    def test_init_with_long_message(self):
        """测试长消息"""
        long_message = "A" * 1000
        config = InterruptConfig(
            requires_confirmation=True, confirmation_message=long_message
        )
        self.assertEqual(config.confirmation_message, long_message)


class TestInterruptData(TestCase):
    """测试 InterruptData 类"""

    def test_init_minimal(self):
        """测试最小初始化"""
        data = InterruptData(tool_name="test_tool")
        self.assertEqual(data.tool_name, "test_tool")
        self.assertIsNone(data.tool_description)
        self.assertEqual(data.tool_args, [])
        self.assertIsNone(data.interrupt_config)

    def test_init_with_all_fields(self):
        """测试带所有字段初始化"""
        args = [
            ToolArg(key="param1", value="value1", type="string"),
            ToolArg(key="param2", value="value2", type="number"),
        ]
        config = InterruptConfig(
            requires_confirmation=True, confirmation_message="Please confirm"
        )
        data = InterruptData(
            tool_name="test_tool",
            tool_description="Test tool description",
            tool_args=args,
            interrupt_config=config,
        )
        self.assertEqual(data.tool_name, "test_tool")
        self.assertEqual(data.tool_description, "Test tool description")
        self.assertEqual(len(data.tool_args), 2)
        self.assertIsNotNone(data.interrupt_config)

    def test_init_with_description(self):
        """测试带工具描述初始化"""
        data = InterruptData(tool_name="test_tool", tool_description="Tool description")
        self.assertEqual(data.tool_name, "test_tool")
        self.assertEqual(data.tool_description, "Tool description")

    def test_init_with_args(self):
        """测试带工具参数初始化"""
        args = [
            ToolArg(key="param1", value="value1", type="string"),
            ToolArg(key="param2", value=123, type="number"),
        ]
        data = InterruptData(tool_name="test_tool", tool_args=args)
        self.assertEqual(len(data.tool_args), 2)
        self.assertEqual(data.tool_args[0].key, "param1")
        self.assertEqual(data.tool_args[1].value, 123)

    def test_init_with_config(self):
        """测试带中断配置初始化"""
        config = InterruptConfig(
            requires_confirmation=True, confirmation_message="Confirm?"
        )
        data = InterruptData(tool_name="test_tool", interrupt_config=config)
        self.assertIsNotNone(data.interrupt_config)
        self.assertTrue(data.interrupt_config.requires_confirmation)

    def test_model_dump_all_fields(self):
        """测试序列化所有字段"""
        args = [ToolArg(key="param", value="value", type="string")]
        config = InterruptConfig(
            requires_confirmation=True, confirmation_message="Confirm"
        )
        data = InterruptData(
            tool_name="test_tool",
            tool_description="Description",
            tool_args=args,
            interrupt_config=config,
        )
        dumped = data.model_dump()
        self.assertEqual(dumped["tool_name"], "test_tool")
        self.assertEqual(dumped["tool_description"], "Description")
        self.assertEqual(len(dumped["tool_args"]), 1)
        self.assertIsNotNone(dumped["interrupt_config"])

    def test_model_dump_minimal(self):
        """测试序列化最小字段"""
        data = InterruptData(tool_name="test_tool")
        dumped = data.model_dump()
        self.assertEqual(dumped["tool_name"], "test_tool")
        self.assertEqual(dumped["tool_args"], [])

    def test_multiple_args(self):
        """测试多个工具参数"""
        args = [
            ToolArg(key="arg1", value="val1", type="string"),
            ToolArg(key="arg2", value="val2", type="string"),
            ToolArg(key="arg3", value="val3", type="string"),
        ]
        data = InterruptData(tool_name="test_tool", tool_args=args)
        self.assertEqual(len(data.tool_args), 3)

    def test_empty_args_list(self):
        """测试空参数列表"""
        data = InterruptData(tool_name="test_tool", tool_args=[])
        self.assertEqual(data.tool_args, [])
