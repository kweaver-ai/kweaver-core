from app.common.config import Config
from data_migrations.init.manage_built_in_agent_and_tool import (
    main as init_built_in_agent_and_tool,
)


def handle_built_in():
    if Config.local_dev.do_not_init_built_in_agent_and_tool:
        return
    init_built_in_agent_and_tool()
