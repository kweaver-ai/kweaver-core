from sqlalchemy import (BIGINT, VARCHAR, Column,
                        DateTime, Integer, String, UniqueConstraint)

from models.db import get_base
from common.configs import db_configs

Base = get_base(db_configs["name"], db_configs["type"])
  
class PythonPackage(Base): # type: ignore
    __tablename__ = "t_python_package"

    f_id = Column(VARCHAR(32), primary_key=True)
    f_created_at = Column(Integer, nullable=False)
    f_name = Column(VARCHAR(255), nullable=False, unique=True)
    f_oss_id = Column(VARCHAR(32), nullable=False)
    f_oss_key = Column(VARCHAR(32), nullable=False)
    f_creator_id = Column(VARCHAR(36), nullable=False)
    f_creator_name = Column(VARCHAR(128), nullable=False)

class AutomationAdmin(Base):
    __tablename__ = "t_content_admin"

    f_id = Column(BIGINT, primary_key=True)
    f_user_id = Column(VARCHAR(40), nullable=False)
    f_user_name = Column(VARCHAR(128), nullable=False)

class BaseModel:
    db_name = db_configs["name"]