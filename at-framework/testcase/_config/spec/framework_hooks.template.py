# 复制为 testcase/<模块名>/framework_hooks.py
# 框架在 pytest session 开始时若环境变量 AT_CLEAN_UP=1 且模块匹配，会调用 session_clean_up(config, allure)。
# dataflow、ontology、execution-factory 等模块若无会话级清理需求，可不创建此文件。


def session_clean_up(config, allure):
    """
    :param config: configparser 转成的 dict，含 'env','test_data','internal' 等键
    :param allure: allure 模块，失败时可 allure.attach(...)
    常用: config['server']['host'], config['server'].get('base_url','https://host'),
         config['test_data']['application_token']
    """
    pass
