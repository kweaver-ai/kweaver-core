from models.models import AutomationAdmin
from models.db import get_session
from errors.errors import InternalErrException, NotFoundException
from common.logger import logger
from models.models import BaseModel

class AutomationAdminModel(BaseModel):
    
    async def get_automation_admin(self, user_id: str) -> AutomationAdmin:
        sqlStr = f"select f_id, f_user_id, f_user_name from {self.db_name}.t_content_admin where f_user_id = :user_id"
        with get_session(expire_on_commit=False) as session:
            try:
                res = session.execute(sqlStr, {"user_id": user_id}).fetchone()
            except Exception as e:
                logger.warning(f"[get_automation_admin] failed to get automation admin, detail: {e}")
                raise InternalErrException(detail="db operation failed")
            if res is None:
                raise NotFoundException(detail=f"content admin {user_id} not found")
            row = dict(res)
            return AutomationAdmin(**row)

