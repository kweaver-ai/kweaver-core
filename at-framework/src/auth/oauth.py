#!/usr/bin/env python
# coding:utf-8
import json
import random
import string
import urllib.parse

import requests
import urllib3
from urllib3.exceptions import InsecureRequestWarning

urllib3.disable_warnings(InsecureRequestWarning)


class OAuth:
    def __init__(self, namespace='dip'):
        self.namespace = namespace

    def AdminURL(self, auth_request, ip: str, adminUrlPort: str) -> str:
        """管理端 OAuth2 Admin 基础 URL，使用传入的 IP"""
        return f'{auth_request}://{ip}:{adminUrlPort}/admin/oauth2/'

    def PublicURL(self, ip: str) -> str:
        """对外 OAuth2 Public 基础 URL，使用传入的 IP"""
        return f'https://{ip}:443/oauth2/'

    # 授权相关
    def OAuthUser(self, ip, port, client_id, response_type, scope, redirect_uri, cookies):
        state = ''.join(random.sample(string.ascii_letters, 24))
        query = f'client_id={client_id}&response_type={response_type}&scope={scope}&redirect_uri={redirect_uri}&state={state}'
        req_url = self.PublicURL(ip) + 'auth?' + query
        r = requests.request('GET', req_url, cookies=cookies, verify=False, allow_redirects=False)
        cookies = requests.utils.dict_from_cookiejar(r.cookies)
        if r.status_code in (302, 303):
            location = dict(urllib.parse.parse_qsl(urllib.parse.urlsplit(r.headers.get("Location", "")).query))
            login_challenge = location.get('login_challenge', '')
            return r.status_code, {"login_challenge": login_challenge, "cookies": cookies, "headers": r.headers, "state": state}
        else:
            return r.status_code, r.content

    def GetOAuth(self, req_url, cookies):
        r = requests.request('GET', req_url, cookies=cookies, verify=False, allow_redirects=False)
        cookies = requests.utils.dict_from_cookiejar(r.cookies)
        if r.status_code == 302:
            location = dict(urllib.parse.parse_qsl(urllib.parse.urlsplit(r.headers.get("Location", "")).query))
            consent_challenge = location.get('consent_challenge', '')
            return r.status_code, {"consent_challenge": consent_challenge, "cookies": cookies}
        else:
            return r.status_code, r.content

    def GetCode(self, req_url, cookies):
        r = requests.request('GET', req_url, cookies=cookies, verify=False, allow_redirects=False)
        if r.status_code == 303:
            location = dict(urllib.parse.parse_qsl(urllib.parse.urlsplit(r.headers.get("Location", "")).query))
            code = location.get('code', '')
            return r.status_code, {"code": code, "headers": r.headers}
        else:
            return r.status_code, {"headers": r.headers}

    def GetConsent(self, auth_request, ip, adminUrlPort, consent_challenge, cookies):
        req_url = f'{self.AdminURL(auth_request, ip, adminUrlPort)}auth/requests/consent?consent_challenge={consent_challenge}'
        r = requests.request('GET', req_url, cookies=cookies, verify=False, allow_redirects=False)
        return r.status_code, json.loads(r.content)

    def AcceptConsent(self, auth_request, ip, adminUrlPort, consent_challenge, grant_scope, remember, remember_for, access_token):
        req_url = f'{self.AdminURL(auth_request, ip, adminUrlPort)}auth/requests/consent/accept?consent_challenge={consent_challenge}'
        data = {"grant_scope": grant_scope, "remember": remember, "remember_for": remember_for, "session": {"access_token": access_token}}
        r = requests.request('PUT', req_url, json=data, verify=False, allow_redirects=False)
        return r.status_code, json.loads(r.content)

    def ApplyToken(self, ip, client_id, client_secret, grant_type, code, refresh_token, scope, redirect_uri):
        req_url = self.PublicURL(ip) + '/token'
        headers = {"Content-Type": "application/x-www-form-urlencoded"}
        if grant_type == 'authorization_code':
            data = {"grant_type": grant_type, "code": code, "redirect_uri": redirect_uri}
        elif grant_type == 'client_credentials':
            data = {"grant_type": grant_type, "scope": scope}
        elif grant_type == 'refresh_token':
            data = {"grant_type": grant_type, "refresh_token": refresh_token}
        else:
            data = {"grant_type": grant_type}
        r = requests.request('POST', req_url, data=data, headers=headers, auth=(client_id, client_secret), verify=False)
        return r.status_code, json.loads(r.content)

    # 认证相关
    def GetLogin(self, auth_request, ip, adminUrlPort, login_challenge):
        req_url = self.AdminURL(auth_request, ip, adminUrlPort) + f'auth/requests/login?login_challenge={login_challenge}'
        r = requests.request('GET', req_url, verify=False, allow_redirects=False)
        return r.status_code, json.loads(r.content)

    def AcceptLogin(self, auth_request, ip, adminUrlPort, login_challenge, subject, remember, remember_for, acr, context):
        req_url = self.AdminURL(auth_request, ip, adminUrlPort) + f'auth/requests/login/accept?login_challenge={login_challenge}'
        data = {"acr": acr, "remember": remember, "remember_for": remember_for, "subject": subject, "context": context}
        r = requests.request('PUT', req_url, json=data, verify=False, allow_redirects=False)
        return r.status_code, json.loads(r.content)

    def RegisterClient1(self, ip, port, grant_types, response_types, scope, redirect_uris, metadata, post_logout_redirect_uris, client_name, body):
        req_url = self.PublicURL(ip) + 'clients'
        if not body:
            body = {"grant_types": grant_types, "response_types": response_types, "scope": scope, "redirect_uris": redirect_uris, "post_logout_redirect_uris": post_logout_redirect_uris, "client_name": client_name, "metadata": metadata}
        r = requests.request('POST', req_url, json=body, verify=False)
        return r.status_code, json.loads(r.content)
