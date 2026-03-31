# 当前所在包的入口文件，所有的boot方法都会在这里调用


def on_boot_run():
    # 1. 初始化内置agent和工具
    from . import built_in

    built_in.handle_built_in()

    # 2. 启动时输出Config信息
    from app.common.config import Config

    if Config.local_dev.is_show_config_on_boot:
        from app.common.struct_logger import struct_logger

        struct_logger.console_logger.debug(
            "Config配置信息", config_info=Config.to_dict()
        )
