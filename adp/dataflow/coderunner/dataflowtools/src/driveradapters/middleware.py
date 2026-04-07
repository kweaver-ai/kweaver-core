from tornado.web import RequestHandler
from common.configs import is_auth_enable
from models.driven_models import UserInfo
from logics.automation_admin import AutomationAdminService
from errors.errors import UnauthorizedException, NoPermissionException
from drivenadapters.hydra import Hydra
from drivenadapters.user_management import UserManagement
class Middleware:
    def __init__(self, handler: RequestHandler):
        self.handler = handler

    def process_request(self):
        pass

    def process_response(self):
        pass

class CheckToken(Middleware):
    '''
    检查token是否有效
    '''
    hydra = Hydra()
    user_management = UserManagement()

    def build_mock_user_info(self):
        user_agent = self.handler.request.headers.get("User-Agent", "")
        payload = {
            "active": True,
            "scope": "",
            "client_id": "mock-client-id",
            "sub": "mock-user-id",
            "ext": {"login_ip": "127.0.0.1", "udid": "mock-udid", "client_type": "web"},
            "user_name": "mock-user",
            "account_type": "user",
            "parent_deps": [],
            "roles": ["admin"],
            "visitor_type": "authenticated_user",
            "login_ip": "127.0.0.1",
            "user_agent": user_agent,
        }
        return UserInfo(**payload)

    async def check_token(self):
        if not is_auth_enable():
            return self.build_mock_user_info()
        header = self.handler.request.headers.get('Authorization', "")
        if header == "":
            raise UnauthorizedException(cause="token empty")
        try:
            prefix = header.split(' ', 1)[0].lower()
            if prefix != "bearer":
                raise UnauthorizedException(cause="token invalid")
            access_token = header.split(' ', 1)[1]
        except:
            raise UnauthorizedException(cause="token invalid")
        user_agent = self.handler.request.headers.get('User-Agent', "")
        res = await self.hydra.check_token(access_token)
        if not res:
            raise UnauthorizedException(cause="token invalid")
        if res.get("active") is False:
            raise UnauthorizedException(cause="token does not active")
        
        res["user_agent"] = user_agent
        res["login_ip"] = res.get("ext").get("login_ip")
        
        if res.get("sub") != res.get("client_id"):
            res["account_type"] = "user"
            visitor_type = res.get("visitor_type") 
            if visitor_type == "realname":
                res["visitor_type"] = "authenticated_user"
            elif visitor_type == "anonymous":
                res["visitor_type"] = "anonymous_user"
                return UserInfo(**res)

            user_info = await self.user_management.get_user_info(res.get("sub"))
            if not user_info:
                raise UnauthorizedException(cause="token invalid")
            
            res["user_name"] = user_info[0].get("name")
            res['parent_deps'] = user_info[0].get("parent_deps")
            res['roles'] = user_info[0].get("roles")   
            

        else:
            res["user_name"] = ""
            res["account_type"] = "app"
            
        return UserInfo(**res)

    async def process_request(self):
        return await self.check_token()

class CheckAutomationAdmin(Middleware):
    '''
    检查用户是否是自动化管理员
    '''
    automation_admin_service = AutomationAdminService()

    async def check_automation_admin(self):
        if not is_auth_enable():
            return True
        user_id = self.handler.user_info.user_id
        res = await self.automation_admin_service.check_automation_admin(user_id)
        if not res:
            raise NoPermissionException(detail= "user is not automation admin")

    async def process_request(self):
        return await self.check_automation_admin()
