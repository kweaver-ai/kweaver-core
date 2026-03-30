# encoding:utf-8
import urllib.parse

from src.auth.eacp_api import EACP_API
from src.auth.oauth import OAuth


class Login:
    def __init__(self, namespace='anyshare'):
        self.namespace = namespace

    def UserLogin(self, host, client_host, account, password, public_port=443, admin_port=9998, name="web", client_type="web", description="mi", udids=None):
        if udids is None:
            udids = ['0a-23-fd-dd-aa-dd-xc']
        ip = host
        client_ip = client_host
        grant_types = ['authorization_code', 'refresh_token', 'implicit']
        response_types = ['token id_token', 'code', 'token']
        redirect_uris = [f'https://{ip}:9010/callback']
        device = {'client_type': client_type}
        metadata = {'device': device}
        post_logout_redirect_uris = [f'https://{ip}:9010/success-logout']

        oauth2 = OAuth(self.namespace)
        eacp = EACP_API(self.namespace)

        # 动态注册客户端
        client = oauth2.RegisterClient1(ip, public_port, grant_types, response_types, 'openid offline all', redirect_uris, metadata, post_logout_redirect_uris, 'user', '')
        if client[0] != 201:
            return 'fail'
        client_id = client[1]["client_id"]
        client_secret = client[1]["client_secret"]

        # 授权用户
        oauth = oauth2.OAuthUser(ip, public_port, client_id, 'code', 'openid+offline+all', f'https://{ip}:9010/callback', '')
        if oauth[0] != 302:
            return 'fail'
        login_challenge = oauth[1]["login_challenge"]
        cookies = oauth[1]["cookies"]

        # 获取登录请求
        get_login = oauth2.GetLogin(ip, admin_port, login_challenge)
        if get_login[0] != 200:
            return 'fail'

        # 身份验证
        get_new = eacp.GetNew(account, password, name, client_type, description, udids, '', '', ip, 9080, client_ip)
        if get_new[0] != 200:
            return 'fail'
        user_id = get_new[1]["user_id"]
        context = get_new[1]["context"]

        # 接受登录请求
        accept_login = oauth2.AcceptLogin(ip, admin_port, login_challenge, user_id, True, 3600, '', context)
        if accept_login[0] != 200:
            return 'fail'
        req_url = urllib.parse.unquote(accept_login[1]["redirect_to"])

        # 获取授权
        get = oauth2.GetOAuth(req_url, cookies)
        if get[0] != 302:
            return 'fail'
        consent_challenge = get[1]["consent_challenge"]
        cookies = get[1]['cookies']

        # 获取授权请求
        get_consent = oauth2.GetConsent(ip, admin_port, consent_challenge, '')
        if get_consent[0] != 200:
            return 'fail'
        context = get_consent[1]['context']

        # 接受授权请求
        grant_scope = ['openid', 'offline', 'all']
        accept_consent = oauth2.AcceptConsent(ip, admin_port, consent_challenge, grant_scope, True, 3600, context)
        if accept_consent[0] != 200:
            return 'fail'
        req_url = urllib.parse.unquote(accept_consent[1]["redirect_to"])

        # 获取 code
        get_code = oauth2.GetCode(req_url, cookies)
        if get_code[0] != 303:
            return 'fail'
        code = get_code[1]["code"]

        # 申请令牌
        token = oauth2.ApplyToken(ip, client_id, client_secret, 'authorization_code', code, '', '', f'https://{ip}:9010/callback')
        if token[0] != 200 or len(token[1].get("access_token", "")) == 0:
            return 'fail'
        else:
            refresh_token = token[1]["refresh_token"]
            clients = {'clientId': client_id, 'client_secret': client_secret, 'cookies': cookies, 'refresh_token': refresh_token}
            result = [get_new[0], user_id, token[1]["access_token"], clients]
            return result
