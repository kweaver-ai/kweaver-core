import os


languages = ["zh_cn", "zh_tw", "en_us"]

class Locale:
    
    console_log_map = {
        "upload_pkg": {
            languages[0]: [f"上传 Python依赖包 “<%s>” 成功", ""],
            languages[1]: [f"上載 Python依賴套件 “<%s>” 成功", ""],
            languages[2]: [f"The Python dependency package \"<%s>\" is uploaded successfully.", ""]

        },
        "delete_pkg": {
            languages[0]: [f"删除 Python依赖包 “<%s>” 成功", ""],
            languages[1]: [f"刪除 Python依賴套件 “<%s>” 成功", ""],
            languages[2]: [f"The Python dependency package \"<%s>\" is deleted successfully.", ""]
        }
    }

    @classmethod
    def get_console_log(cls, key, detail_params, ext_params):
        lang = os.environ.get("SERVICE_LANG", "zh_cn").lower()
        
        body = cls.console_log_map.get(key, {}).get(lang)
        if body is None:
            return "", ""

        if body[0] == "":
            detail = "" 
        else:
            detail = body[0] % detail_params

        if body[1] == "":
            ext_msg = ""
        else:
            ext_msg = body[1] % ext_params

        return detail, ext_msg
