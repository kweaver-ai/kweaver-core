# encoding:utf-8
import urllib.parse

from src.auth.eacp_api import EACP_API
from src.auth.oauth import OAuth
from src.config.setting import config as setting_config


class Login:
    def __init__(self, namespace='anyshare'):
        self.namespace = namespace

    def UserLogin(self, host, client_host, account, password, name="web", client_type="web", description="mi", udids=None):
        if udids is None:
            udids = ['0a-23-fd-dd-aa-dd-xc']
        ip = host
        client_ip = client_host
        admin_url_port = str(setting_config.get("server", "adminUrlPort")).strip()
        getnew_port = str(setting_config.get("server", "getnewPort")).strip()
        public_port = str(setting_config.get("server", "public_port")).strip()
        auth_request = str(setting_config.get("server", "auth_request")).strip()
        grant_types = ['authorization_code', 'refresh_token', 'implicit']
        response_types = ['token id_token', 'code', 'token']
        redirect_uris = [f'https://{ip}:9010/callback']
        device = {'client_type': client_type}
        metadata = {'device': device}
        post_logout_redirect_uris = [f'https://{ip}:9010/success-logout']

        oauth2 = OAuth(self.namespace)
        eacp = EACP_API(self.namespace)

        # 动态注册客户端
        print("RegisterClient1")
        client = oauth2.RegisterClient1(ip, public_port, grant_types, response_types, 'openid offline all', redirect_uris, metadata, post_logout_redirect_uris, 'user', '')
        if client[0] != 201:
            print("RegisterClient1 fail", client)
            return 'fail'
        client_id = client[1]["client_id"]
        client_secret = client[1]["client_secret"]

        # 授权用户
        print("OAuthUser")
        oauth = oauth2.OAuthUser(ip, public_port, client_id, 'code', 'openid+offline+all', f'https://{ip}:9010/callback', '')
        if oauth[0] != 302:
            print("OAuthUser fail", oauth)
            return 'fail'
        login_challenge = oauth[1]["login_challenge"]
        cookies = oauth[1]["cookies"]

        # 获取登录请求
        print("GetLogin")
        get_login = oauth2.GetLogin(auth_request, ip, admin_url_port, login_challenge)
        if get_login[0] != 200:
            print("GetLogin fail", get_login)
            return 'fail'

        # 身份验证
        print("GetNew")
        get_new = eacp.GetNew(account, auth_request, getnew_port, password, name, client_type, description, udids, '', '', ip, 9080, client_ip)
        print("get_new", get_new)
        if get_new[0] != 200:
            print("get_new", get_new)
            print("GetNew fail", get_new)
            return 'fail'
        user_id = get_new[1]["user_id"]
        context = get_new[1]["context"]

        # 接受登录请求
        print("AcceptLogin")
        accept_login = oauth2.AcceptLogin(auth_request, ip, admin_url_port, login_challenge, user_id, True, 3600, '', context)
        if accept_login[0] != 200:
            print("AcceptLogin fail", accept_login)
            return 'fail'
        req_url = urllib.parse.unquote(accept_login[1]["redirect_to"])

        # 获取授权
        print("GetOAuth")
        get = oauth2.GetOAuth(req_url, cookies)
        if get[0] != 302:
            print("GetOAuth fail", get)
            return 'fail'
        consent_challenge = get[1]["consent_challenge"]
        cookies = get[1]['cookies']

        # 获取授权请求
        print("GetConsent")
        get_consent = oauth2.GetConsent(auth_request, ip, admin_url_port, consent_challenge, '')
        if get_consent[0] != 200:
            print("GetConsent fail", get_consent)
            return 'fail'
        context = get_consent[1]['context']

        # 接受授权请求
        grant_scope = ['openid', 'offline', 'all']
        print("AcceptConsent")
        accept_consent = oauth2.AcceptConsent(auth_request, ip, admin_url_port, consent_challenge, grant_scope, True, 3600, context)
        if accept_consent[0] != 200:
            print("AcceptConsent fail", accept_consent)
            return 'fail'
        req_url = urllib.parse.unquote(accept_consent[1]["redirect_to"])

        # 获取 code
        print("GetCode")
        get_code = oauth2.GetCode(req_url, cookies)
        if get_code[0] != 303:
            print("GetCode fail", get_code)
            return 'fail'
        code = get_code[1]["code"]

        # 申请令牌
        print("ApplyToken")
        token = oauth2.ApplyToken(ip, client_id, client_secret, 'authorization_code', code, '', '', f'https://{ip}:9010/callback')
        if token[0] != 200 or len(token[1].get("access_token", "")) == 0:
            print("ApplyToken fail", token)
            return 'fail'
        else:
            refresh_token = token[1]["refresh_token"]
            clients = {'clientId': client_id, 'client_secret': client_secret, 'cookies': cookies, 'refresh_token': refresh_token}
            result = [get_new[0], user_id, token[1]["access_token"], clients]
            return result
