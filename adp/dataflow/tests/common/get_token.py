# -*- coding:UTF-8 -*-

import os
import sys
import random
import requests
import string
import json
import urllib
import base64
import hashlib
import warnings 
warnings.filterwarnings("ignore")

from hashlib import md5
from urllib.parse import parse_qsl
from urllib.parse import urlsplit
from urllib.parse import unquote
from M2Crypto import RSA, BIO

class GetToken(object):
    def __init__(self, host):
        """
        """
        self.host = host
        self.hydra_public_port = 443
        self.eacp_svc_ip = "eacp-private.anyshare.svc.cluster.local"
        self.hydra_admin_svc_ip = "hydra-admin.anyshare.svc.cluster.local"
        self.hydra_admin_port = 4445
        self.public_url = 'https://%s:%s/'%(self.host, self.hydra_public_port)
        self.admin_url = 'http://%s:%s/'%(self.hydra_admin_svc_ip, self.hydra_admin_port)

        self.key = '''
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4E+eiWRwffhRIPQYvlXU
jf0b3HqCmosiCxbFCYI/gdfDBhrTUzbt3fL3o/gRQQBEPf69vhJMFH2ZMtaJM6oh
E3yQef331liPVM0YvqMOgvoID+zDa1NIZFObSsjOKhvZtv9esO0REeiVEPKNc+Dp
6il3x7TV9VKGEv0+iriNjqv7TGAexo2jVtLm50iVKTju2qmCDG83SnVHzsiNj70M
iviqiLpgz72IxjF+xN4bRw8I5dD0GwwO8kDoJUGWgTds+VckCwdtZA65oui9Osk5
t1a4pg6Xu9+HFcEuqwJTDxATvGAz1/YW0oUisjM0ObKTRDVSfnTYeaBsN6L+M+8g
CwIDAQAB
-----END PUBLIC KEY-----
        '''
        self.key1 = '''
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC7JL0DcaMUHumSdhxXTxqiABBC
DERhRJIsAPB++zx1INgSEKPGbexDt1ojcNAc0fI+G/yTuQcgH1EW8posgUni0mcT
E6CnjkVbv8ILgCuhy+4eu+2lApDwQPD9Tr6J8k21Ruu2sWV5Z1VRuQFqGm/c5vaT
OQE5VFOIXPVTaa25mQIDAQAB
-----END PUBLIC KEY-----
        '''

    def registerClient(self):
        #注册OAuth 2.0客户端
        requrl = self.public_url + 'oauth2/clients'
        redirect_uri = "https://%s:9010/callback" % self.host
        post_logout_redirect_uri = "https://%s:9010/successful-logout" % self.host
        data =  {
                    "grant_types": [
                        "authorization_code",
                        "implicit",
                        "refresh_token"
                    ],
                    "response_types": [
                        "code",
                        "token",
                        "token id_token"
                    ],
                    "scope": "offline openid all",
                    "redirect_uris": [redirect_uri],
                    "client_name": "test",
                    "post_logout_redirect_uris": [post_logout_redirect_uri],
                    "metadata": {
                        "device": {"client_type": "unknown"}
                    }
                }
        r = requests.request('POST', requrl, json=data, verify=False)
        if r.status_code < 200 or r.status_code > 299:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))
        return json.loads(r.content)

    def OAuthUser(self, response_type, client_id):
        #授权用户
        state = ''.join(random.sample(string.ascii_letters, 24))
        # redirect_uri = urllib.quote(redirect_uri)
        redirect_uri = "https://%s:9010/callback" % self.host
        query = 'client_id=%s&response_type=%s&scope=openid+offline+all&redirect_uri=%s&state=%s' % (client_id, response_type, redirect_uri, state)
        requrl = self.public_url + 'oauth2/auth?' + query
        # r = requests.request('GET', requrl, cookies=cookies, verify=False, allow_redirects=False)
        r = requests.request('GET', requrl, verify=False, allow_redirects=False)
        cookies = requests.utils.dict_from_cookiejar(r.cookies)
        if r.status_code == 302:
            login_challenge = dict(parse_qsl(urlsplit(r.headers["Location"]).query))['login_challenge']
            return {"login_challenge": login_challenge, "cookies": cookies, "headers": r.headers, "state": state}
        else:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))


    def GetLogin(self, login_challenge):
        #获取登陆请求
        # import pdb;pdb.set_trace()
        requrl = self.admin_url + 'admin/oauth2/auth/requests/login?login_challenge=%s' % login_challenge
        r = requests.request('GET', requrl, verify=False, allow_redirects=False)
     
        if r.status_code < 200 or r.status_code > 299:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))
        return json.loads(r.content)

    def AcceptLogin(self, login_challenge, subject, context):
        #接受登陆请求
        # import pdb;pdb.set_trace()
        requrl = self.admin_url + 'admin/oauth2/auth/requests/login/accept?login_challenge=%s' % login_challenge
        # data = {"acr": acr, "remember": remember, "remember_for": remember_for, "subject": subject, "context": context}
        data = {
                "acr": "string",
                "context": context,
                # "force_subject_identifier": "string",
                "remember": True,
                "remember_for": 3600,
                "subject": subject
                }
        r = requests.request('PUT', requrl, json=data, verify=False, allow_redirects=False)
        if r.status_code < 200 or r.status_code > 299:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))

        redirect_to = json.loads(r.content)["redirect_to"]
        redirect_to = unquote(redirect_to)
        return redirect_to

    def GetOAuth(self, requrl, cookies):
        #获取授权
        # import pdb;pdb.set_trace()
        r = requests.request('GET', requrl, cookies=cookies, verify=False, allow_redirects=False)
        cookies = requests.utils.dict_from_cookiejar(r.cookies)
        if r.status_code == 302:
            consent_challenge = dict(parse_qsl(urlsplit(r.headers["Location"]).query))['consent_challenge']
            return {"consent_challenge": consent_challenge, "cookies": cookies}
        else:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))

    def GetContent(self, consent_challenge, cookies):
        #获取授权请求
        requrl = self.admin_url + 'admin/oauth2/auth/requests/consent?consent_challenge=%s' % consent_challenge
        r = requests.request('GET', requrl, cookies=cookies, verify=False, allow_redirects=False)
        if r.status_code < 200 or r.status_code > 299:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))

        return json.loads(r.content)

    def AcceptConsent(self, consent_challenge, scope, context):
        #接受授权请求
        # import pdb;pdb.set_trace
        requrl = self.admin_url + 'admin/oauth2/auth/requests/consent/accept?consent_challenge=%s' % consent_challenge
        # data = {"grant_scope": grant_scope, "remember": remember, "remember_for": remember_for, "session": {"access_token": access_token}}

        data = {
                    "grant_access_token_audience": [
                        "string"
                    ],
                    "grant_scope": scope,
                    # "handled_at": "2020-02-19T10:34:52Z",
                    "remember": True,
                    "remember_for": 0,
                    "session": {
                        "access_token": context,
                        "id_token": {
                        }
                    }
                }
        r = requests.request('PUT', requrl, json=data, verify=False, allow_redirects=False)
        if r.status_code < 200 or r.status_code > 299:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))

        redirect_to = json.loads(r.content)["redirect_to"]
        redirect_to = unquote(redirect_to)
        return redirect_to

    def GetToken(self, requrl, cookies):
        #获取implicit token
        r = requests.request('GET', requrl, cookies=cookies, verify=False, allow_redirects=False)
        if r.status_code == 303:
            # https://127.0.0.1:9010/callback#access_token=HQPbKpIMB7Lerxz-XZB-rJywAlnybuWQrH-dl-OVFUY.EvbroXUc4Pw-skJyxc0sdTMswcIwCCUCa5ut4O3GHLM&expires_in=3600&scope=all&state=BEXaqxiVRPpeNCQfvZKHsdtr&token_type=bearer
            access_token = dict(parse_qsl(urlsplit(r.headers["Location"]).fragment))['access_token']
            return access_token
        else:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))

    def GetCode(self, requrl, cookies):
        #获取implicit token
        # import pdb;pdb.set_trace()
        r = requests.request('GET', requrl, cookies=cookies, verify=False, allow_redirects=False)
        if r.status_code == 303:
            # https://127.0.0.1:9010/callback?code=vnl5-ZO6ys33rdCkFlQl_GxOLAErHL5b0sIWrgrqKdY.bJnI_iNFnIfw0ej610Acvb2jAAKURTc_Y8HKhn_6rY0&scope=all&state=iHkXbAmZJnhPBNcWEzVYgeod
            code = dict(parse_qsl(urlsplit(r.headers["Location"]).query))['code']
            return code
        else:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))

    def ApplyToken(self, grant_type, code, client_id, client_secret):
        #申请令牌
        requrl = self.public_url + '/oauth2/token'
        headers = {"Content-Type": "application/x-www-form-urlencoded"}
        redirect_uri = "https://%s:9010/callback" % self.host
        if grant_type == 'authorization_code':
            data = {"grant_type": grant_type, "code": code, "redirect_uri": redirect_uri}
        # if grant_type == 'client_credentials':
        #     data = {"grant_type": grant_type, "scope": scope}
        # if grant_type == 'refresh_token':
        #     data = {"grant_type": grant_type, "refresh_token": refresh_token}
        # print data
        r = requests.request('POST', requrl, data=data, headers=headers, auth=(client_id, client_secret), verify=False)
        if r.status_code < 200 or r.status_code > 299:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))

        return json.loads(r.content)

    def auth_Pwd_RSABase64(self, key, message):
        '''
        param: public_key_loc Path to public key
        param: message String to be encrypted
        return base64 encoded encrypted string
        '''

        pubkey = str(key).encode('utf8')
        bio = BIO.MemoryBuffer(pubkey)
        rsa = RSA.load_pub_key_bio(bio)
        encrypted = rsa.public_encrypt(message.encode('utf8'), RSA.pkcs1_padding)
        # result = encrypted.encode('base64')
        result = base64.b64encode(encrypted)
        result = result.decode("utf-8")
        # print result
        return result

    def modifyAdminPwd(self, newpwd, oldpwd):  
        newpwd = self.auth_Pwd_RSABase64(self.key1, newpwd)
        oldpwd = self.auth_Pwd_RSABase64(self.key1, oldpwd)
        data = {
            "account": "admin",
            "newpwd": newpwd,
            "oldpwd": oldpwd
        }
        suffixofsign = "eisoo.com"
        signdata = json.dumps(data) + suffixofsign
        sign = hashlib.md5(signdata.encode('utf-8')).hexdigest()

        requrl = 'https://%s/api/eacp/v1/auth1/modifypassword??sign=%s' % (self.host, sign) 
        
        re = requests.request('POST', requrl, json=data, verify=False)
        if re.status_code != 200:
            content = json.loads(re.content)
            for key in content.keys():
                print("%s: %s" % (key, content[key]))
            return [re.status_code, json.loads(re.content)]
        else:
            return [re.status_code, re.content]

        # if r.status_code != 200:
        #     raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))
        # return json.loads(r.content)

    def ConsoleLogin(self, account, password):
        # import pdb;pdb.set_trace()
        password = self.auth_Pwd_RSABase64(self.key, password)
        # password = "oRa5L8dOest06rKRze6JfW9+tIPJP3q2wLCAzMuxx9Nu7omelGURtn8OnpxZjHepvGXXpRMtDIyD\n0ET09xsbh/zZpAZT/aYU/pZgsCej1OqqnkLet6w44SQSTRuJ/GqiUN/M6QHJ4srQKu/R+e0kc2rP\nnJq7BWvaDTW+IFpABNw=\n"
        requrl = "http://%s:9998/api/eacp/v1/auth1/consolelogin" % (self.eacp_svc_ip)
        data = {
            "device": {
                "udids": [],
                "client_type": "linux",
                "name": "",
                "description": ""
            },
            "credential": {
                "vcode": {
                    "content": "",
                    "id": ""
                },
                "account": account,
                "password": password,
                "type": "account"
            },
            "ip": self.host
        }
        r = requests.request('POST', requrl, json=data, verify=False)
        if r.status_code != 200:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))
        return json.loads(r.content)

    def Getnew(self, account, password):
        password = self.auth_Pwd_RSABase64(self.key, password)
        #password = "dgjY1D5AIC3+PLRKWZYIfYH9qEPaRILsIOWa0XstwPxf75VWuHcAUR+5GHAM0Pdu5k6WWy4HmK2S\nVm602rbtNMz98oihaaWgeWmzxpx/YllTN4cJUHBiX7HJj5+X8So2zjvXVZWXsrjOb2XOViLgcKjg\n6PGs1bxLoPqVC7tRcVg=\n"
        requrl = "http://%s:9998/api/eacp/v1/auth1/getnew" % (self.eacp_svc_ip)
        data = {"account":account,"password":password,"device":{"client_type":"web"}, "ip": self.host}
        r = requests.request('POST', requrl, json=data, verify=False)
        if r.status_code < 200 or r.status_code > 299:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))
        return json.loads(r.content)

    def get_token(self, host, account, password):
        '''
        管理员/普通用户获取access_token
        '''
        # import pdb;pdb.set_trace()
        register_data = self.registerClient()

        #response_type = 'token'
        response_type = 'code'
        oAuthUser_data = self.OAuthUser(response_type, register_data['client_id'])

        getLogin_data = self.GetLogin(dict(parse_qsl(urlsplit(oAuthUser_data['headers']['Location']).query))['login_challenge'])
        if account == "admin":
            data = self.ConsoleLogin(account, password)
        else:
            data = self.Getnew(account, password)

        redirect_to = self.AcceptLogin(oAuthUser_data['login_challenge'],data['user_id'], data['context'])

        getOAuth_data = self.GetOAuth(redirect_to, oAuthUser_data['cookies'])

        GetContent_data = self.GetContent(getOAuth_data['consent_challenge'], getOAuth_data['cookies'])

        acceptConsent_data = self.AcceptConsent(getOAuth_data['consent_challenge'], GetContent_data['requested_scope'], GetContent_data['context'])

        if response_type == 'token':
            print(self.GetToken(acceptConsent_data, getOAuth_data['cookies']))
        if response_type == 'code':
            code = self.GetCode(acceptConsent_data, getOAuth_data['cookies'])
            applyToken_code = self.ApplyToken('authorization_code', code, register_data["client_id"], register_data["client_secret"])
            #print applyToken_code
        return data['user_id'], applyToken_code["access_token"]

if __name__ == '__main__':
    client = GetToken("10.4.110.62")
    # mod = client.modifyAdminPwd("eisoo.com123", "eisoo.com")
    result = client.get_token("10.4.110.62", "A0", "111111")
    print(result)
