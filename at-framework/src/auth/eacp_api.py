# coding: utf-8
# 从 DP_AT 迁移，依赖 resource/rsa_public.key
import json
import os

import requests
import urllib3
from base64 import b64encode
from Crypto.Cipher import PKCS1_v1_5
from Crypto.PublicKey import RSA
from urllib3.exceptions import InsecureRequestWarning

from src.config.setting import RESOURCE_DIR

urllib3.disable_warnings(InsecureRequestWarning)


class EACP_API:
    def __init__(self, namespace='anyshare'):
        self.namespace = namespace

    def auth_Pwd_RSABase64(self, message: str) -> str:
        """使用项目 resource 目录下的 rsa_public.key 进行 RSA 加密并 base64 编码"""
        key = os.path.join(RESOURCE_DIR, "rsa_public.key")
        with open(key, "r", encoding="utf-8") as f:
            pubkey = f.read()
        public_key = RSA.import_key(pubkey)
        cipher = PKCS1_v1_5.new(public_key)
        encrypted_message = cipher.encrypt(message.encode('utf-8'))
        return b64encode(encrypted_message).decode('utf-8')

    def GetNew(self, account, password, name, client_type, description, udids, id, content, ip, port, clientip):
        """EACP 用户身份验证，返回 user_id 与 context 等信息"""
        password = self.auth_Pwd_RSABase64(password)
        port = '9998'
        url = f"http://{ip}:{port}/api/eacp/v1/auth1/getnew"
        data = {"account": account, "password": password, "device": {"name": name, "client_type": client_type, "description": description, "udids": udids}, "vcode": {"id": id, "content": content}, "ip": clientip}
        r = requests.request('POST', url, json=data, verify=False)
        return r.status_code, json.loads(r.content)
