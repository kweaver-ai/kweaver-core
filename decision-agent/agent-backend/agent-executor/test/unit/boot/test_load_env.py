"""单元测试 - boot/load_env 模块"""

from unittest.mock import patch

from app.boot.load_env import load_env


class TestLoadEnv:
    """测试 load_env 函数"""

    @patch("app.boot.load_env.load_dotenv")
    def test_load_env_calls_load_dotenv(self, mock_load_dotenv):
        """测试 load_env 调用 load_dotenv"""
        load_env()

        # 验证 load_dotenv 被调用
        assert mock_load_dotenv.called

    @patch("app.boot.load_env.load_dotenv")
    def test_load_env_with_correct_path(self, mock_load_dotenv):
        """测试 load_env 使用正确的路径"""
        load_env()

        # 获取调用参数
        call_kwargs = mock_load_dotenv.call_args[1]
        env_file = call_kwargs["dotenv_path"]

        # 检查路径以 .env 结尾
        assert str(env_file).endswith(".env")

    @patch("app.boot.load_env.load_dotenv")
    def test_load_env_override_is_false(self, mock_load_dotenv):
        """测试 load_env 使用 override=False，不覆盖已有环境变量"""
        load_env()

        call_kwargs = mock_load_dotenv.call_args[1]
        assert call_kwargs.get("override") is False
