import os
import uvicorn
from app.utils.app_utils import create_app
from app.core.config import base_config
from app.logs.stand_log import StandLogger
# 导入Kafka优雅关闭处理器
import app.utils.kafka_shutdown


app = create_app()

if __name__ == '__main__':
    print(f"NH_DEBUG={os.getenv('NH_DEBUG')}")
    if os.getenv('NH_DEBUG') == "True":
        print("NH_DEBUG ---")
        import pydevd_pycharm

        pydevd_pycharm.settrace('127.0.0.1', port=9009, stdoutToServer=True, stderrToServer=True, suspend=False)

    StandLogger.info_log("启动 API 服务")
    uvicorn.run(app='main:app', host='0.0.0.0', port=base_config.PORTDEFAULT, limit_concurrency=500, reload=False)
