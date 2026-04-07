from models.automation_admin import AutomationAdminModel
from errors.errors import InternalErrException, NotFoundException
from common.logger import logger

class AutomationAdminService:
    def __init__(self):
        self.automation_admin_model = AutomationAdminModel()

    async def check_automation_admin(self, user_id: str) -> bool:
        try:
            admin = await self.automation_admin_model.get_automation_admin(user_id)
            if admin is not None:
                return True
        except NotFoundException:
            return False
        except Exception as e:
            logger.warning(f"[check_automation_admin] failed to check automation admin, detail: {e}")
            return False