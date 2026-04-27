import os

# 项目根路径: /agent-executor
PROJECT_PATH = os.path.abspath(
    os.path.dirname(
        os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    )
)
# babel翻译的目录，默认是translations。这里可以修改成international
BABEL_TRANSLATION_DIRECTORIES = "app/common/international"
BABEL_TRANSLATION_DIRECTORIES = PROJECT_PATH + "/" + BABEL_TRANSLATION_DIRECTORIES
# babel LC_MESSAGES下面的messages.mo文件名的配置
BABEL_DOMAIN = "messages"
# babel默认的语言类型
BABEL_DEFAULT_LOCALE = "en"
# babel 支持的语言类型
BABEL_LANS = ["en", "zh"]


def pot_path():
    return "{0}/{1}.pot".format(BABEL_TRANSLATION_DIRECTORIES, BABEL_DOMAIN)


def po_path(locale):
    return "{}/{}/LC_MESSAGES/{}.po".format(
        BABEL_TRANSLATION_DIRECTORIES, locale, BABEL_DOMAIN
    )


def cfg_path():
    return "{}/babel.cfg".format(PROJECT_PATH)


"""
提取项目中所有需要翻译的文本，生成模板文件
"""


def extra_command():
    # pybabel extract -F babel.cfg -k _l -o messages.pot .
    return "pybabel extract -F {} -k _l -o {} {}/.".format(
        cfg_path(), pot_path(), PROJECT_PATH
    )


"""
初始化翻译语言文件夹，参数是语言缩写
"""


def init_command(locale):
    # pybabel init -i messages.pot -d app/translations -l zh
    return "pybabel init -i {} -d {} -l {}".format(
        pot_path(), BABEL_TRANSLATION_DIRECTORIES, locale
    )


"""
从指定的模板中更新msgid到po文件里面
"""


def update_command():
    return "pybabel update -i {}/{}.pot -d {}".format(
        BABEL_TRANSLATION_DIRECTORIES, BABEL_DOMAIN, BABEL_TRANSLATION_DIRECTORIES
    )


"""
将所有语言的翻译编译下，编译后的文件查找更快
"""


def compile_command():
    return f"pybabel  compile -d {BABEL_TRANSLATION_DIRECTORIES} -f"


"""
编译所有语言的翻译
"""


def compile_all():
    # 抽取标注的字段，生成翻译模板
    os.system(extra_command())
    # 根据模板生成所有语言翻译需要的文件夹
    for locale in BABEL_LANS:
        # 默认的语言类型不用翻译
        if locale == BABEL_DEFAULT_LOCALE:
            continue
        # 检查po文件
        poFilePath = po_path(locale)
        if not os.path.exists(poFilePath):
            # 创建po文件
            os.system(init_command(locale))
        else:
            # 更新翻译模板内容
            os.system(update_command())
        # 编译模板内容
        os.system(compile_command())
