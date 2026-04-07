import os

from dotenv import load_dotenv


# 在应用启动前加载环境变量
def load_env():
    env_file = os.path.join(os.path.dirname(__file__), "..", "..", ".env")

    # override=False 与原逻辑一致：不覆盖已存在的环境变量
    load_dotenv(dotenv_path=env_file, override=False)
