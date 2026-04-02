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

     # 用户身份验证密码RSABase64加密
    def auth_Pwd_RSABase64(self, message):
        '''
        param: public_key_loc Path to public key
        param: message String to be encrypted
        return base64 encoded encrypted string
        '''

        key = os.path.join(RESOURCE_DIR,"rsa_public.key")
        with open(key, "r") as f:
            pubkey = f.read()
        public_key = RSA.import_key(pubkey)
        cipher = PKCS1_v1_5.new(public_key)
        encrypted_message = cipher.encrypt(message.encode('utf-8'))

        # Convert encrypted byte data to a base64 string
        result = b64encode(encrypted_message).decode('utf-8')
        return result

    def GetNew(self, account,auth_request, getnewPort, password, name, client_type, description, udids, id, content, ip, port, clientip):
        password = self.auth_Pwd_RSABase64(password)
        url = "%s://%s:%s/api/eacp/v1/auth1/getnew" % (auth_request,ip, getnewPort)
        data = {"account": account, "password": password,
                "device": {"name": name, "client_type": client_type, "description": description, "udids": udids}, \
                "vcode": {"id": id, "content": content}, "ip": clientip}
        r = requests.request('POST', url, json=data, verify=False)
        print("getnew response", r)
        return r.status_code, json.loads(r.content)
