from contextlib import contextmanager
from typing import Dict

from sqlalchemy import create_engine
from sqlalchemy.engine.url import URL
from sqlalchemy.ext.declarative import DeclarativeMeta, declarative_base
from sqlalchemy.orm import Session, sessionmaker
from sqlalchemy import MetaData

_engine = None  # pylint: disable=invalid-name
_session_factory = sessionmaker()  # pylint: disable=invalid-name
session_factory = _session_factory

def get_base(schema: str, dbtype):
    if dbtype == 'DM8':
        metadata = MetaData(
            schema=schema
        )
        Base: DeclarativeMeta = declarative_base(metadata=metadata)
    else:
        Base: DeclarativeMeta = declarative_base()
    return Base


def init_db(Base: DeclarativeMeta, **db_configs: Dict):  # pylint: disable=invalid-name
    global _engine  # pylint: disable=global-statement, invalid-name
    driver = db_configs.get("driver", "")
    database = db_configs['name']

    if driver == 'sqlite':
        url = URL(driver,
                  database=database,
                  query={'check_same_thread': False})
        kwargs = {}
        if _engine is not None:
            return

        _engine = create_engine(url, **kwargs)
        Base.metadata.bind = _engine
        _session_factory.configure(bind=_engine)
        Base.metadata.create_all()
    else:
        dbtype = db_configs['type']
        if dbtype == "DM8":
            driver = "dm+dmPython"
            url = URL(driver,
                      host=db_configs['host'],
                      port=db_configs['port'],
                      username=db_configs['user'],
                      password=db_configs['password'],
                      )
        else:
            driver = "mysql+pymysql"
            url = URL(driver,
                      host=db_configs['host'],
                      port=db_configs['port'],
                      username=db_configs['user'],
                      password=db_configs['password'],
                      database=database,
                      query={
                          "charset": db_configs['charset'],
                          "binary_prefix": True
                      })
        kwargs = {
            "pool_size": 20,
            "pool_pre_ping": True,
            "max_overflow": 10
        }
        if _engine is not None:
            return
        _engine = create_engine(url, **kwargs)
        Base.metadata.bind = _engine
        _session_factory.configure(bind=_engine)


@contextmanager
def get_session(**kwargs) -> Session: # type: ignore
    session = _session_factory(**kwargs)
    try:
        yield session
        session.commit()
    except:
        session.rollback()
        raise
    finally:
        session.close()
