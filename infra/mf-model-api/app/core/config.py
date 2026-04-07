import ipaddress
import logging
import os
import sys

from app.utils.observability.observability_setting import ServerInfo, ObservabilitySetting, LogSetting, TraceSetting
import aiohttp


class BaseConfig(object):
    DEBUGDEFAULT = False
    aiohttp_timeout = aiohttp.ClientTimeout(
        total=1800,  # 总超时
        sock_connect=30  # 保留连接超时
    )
    DIPHOSTDEFAULT = "10.4.134.253"
    PORTDEFAULT = 9898
    # 结构化数据库相关
    RDSHOSTDEFAULT = DIPHOSTDEFAULT
    RDSPORTDEFAULT = 3330
    RDSDBNAMEDEFAULT = 'adp'
    RDSUSERDEFAULT = 'root'
    RDSPASSDEFAULT = 'password'

    # redis数据库相关
    REDISCLUSTERMODEDEFAULT = "master-slave"
    # 哨兵
    SENTINELMASTERDEFAULT = DIPHOSTDEFAULT
    SENTINELUSERDEFAULT = "root"
    SENTINELPASSDEFAULT = "password"
    REDISREADHOSTDEFAULT = DIPHOSTDEFAULT
    REDISREADPORTDEFAULT = 6379
    REDISREADUSERDEFAULT = 'root'
    REDISREADPASSDEFAULT = 'password'

    REDISWRITEHOSTDEFAULT = DIPHOSTDEFAULT
    REDISWRITEPORTDEFAULT = 6379
    REDISWRITEUSERDEFAULT = 'root'
    REDISWRITEPASSDEFAULT = 'password'
    # 主从
    REDISHOSTDEFAULT = DIPHOSTDEFAULT
    REDISPORTDEFAULT = 6379
    REDISUSERDEFAULT = 'root'
    REDISPASSDEFAULT = 'password'
    # 鉴权相关
    OAUTHADMINHOSTDEFAULT = DIPHOSTDEFAULT
    OAUTHADMINPORTDEFAULT = 4445
    USERMANAGEMENTPRIVATEHOSTDEFAULT = DIPHOSTDEFAULT
    USERMANAGEMENTPRIVATEPORTDEFAULT = 30980

    # 资源权限相关
    AUTHORIZATIONPRIVATEHOSTDEFAULT = DIPHOSTDEFAULT
    AUTHORIZATIONPRIVATEPORTDEFAULT = 30920

    # KAFKA相关
    KAFKAHOSTDEFAULT = DIPHOSTDEFAULT
    KAFKAPORTDEFAULT = 9097
    KAFKAUSERDEFAULT = "anyrobot"
    KAFKAPASSDEFAULT = "password"
    # app相关
    APP_PORT = int(os.getenv('PORT', PORTDEFAULT))
    DEBUG = True if os.getenv("DEBUG") else DEBUGDEFAULT
    LOG_LEVEL = logging.debug if DEBUG else logging.info

    # 结构化数据库相关
    RDSHOST = os.getenv("RDSHOST", RDSHOSTDEFAULT)
    RDSPORT = int(os.getenv("RDSPORT", RDSPORTDEFAULT))
    RDSDBNAME = os.getenv("RDSDBNAME", RDSDBNAMEDEFAULT)
    RDSUSER = os.getenv("RDSUSER", RDSUSERDEFAULT)
    RDSPASS = os.getenv("RDSPASS", RDSPASSDEFAULT)

    # redis数据库相关
    REDISCLUSTERMODE = os.getenv("REDISCLUSTERMODE", REDISCLUSTERMODEDEFAULT)
    REDISHOST = os.getenv("REDISHOST", DIPHOSTDEFAULT)
    REDISPORT = os.getenv("REDISPORT", REDISPORTDEFAULT)
    REDISUSER = os.getenv("REDISUSER", REDISUSERDEFAULT)
    REDISPASS = os.getenv("REDISPASS", REDISPASSDEFAULT)
    REDISREADHOST = os.getenv("REDISREADHOST", REDISREADHOSTDEFAULT)
    REDISREADPORT = os.getenv("REDISREADPORT", REDISREADPORTDEFAULT)
    REDISREADUSER = os.getenv("REDISREADUSER", REDISREADUSERDEFAULT)
    REDISREADPASS = os.getenv("REDISREADPASS", REDISREADPASSDEFAULT)
    REDISWRITEHOST = os.getenv("REDISWRITEHOST", REDISWRITEHOSTDEFAULT)
    REDISWRITEPORT = os.getenv("REDISWRITEPORT", REDISWRITEPORTDEFAULT)
    REDISWRITEUSER = os.getenv("REDISWRITEUSER", REDISWRITEUSERDEFAULT)
    REDISWRITEPASS = os.getenv("REDISWRITEPASS", REDISWRITEPASSDEFAULT)
    SENTINELMASTER = os.getenv("SENTINELMASTER", SENTINELMASTERDEFAULT)
    SENTINELUSER = os.getenv("SENTINELUSER", SENTINELUSERDEFAULT)
    SENTINELPASS = os.getenv("SENTINELPASS", SENTINELPASSDEFAULT)
    # 登录鉴权相关
    OAUTHADMINHOST = os.getenv("OAUTHADMINHOST", OAUTHADMINHOSTDEFAULT)
    OAUTHADMINPORT = os.getenv("OAUTHADMINPORT", OAUTHADMINPORTDEFAULT)
    USERMANAGEMENTPRIVATEHOST = os.getenv("USERMANAGEMENTPRIVATEHOST", USERMANAGEMENTPRIVATEHOSTDEFAULT)
    USERMANAGEMENTPRIVATEPORT = os.getenv("USERMANAGEMENTPRIVATEPORT", USERMANAGEMENTPRIVATEPORTDEFAULT)
    # 资源权限相关
    AUTHORIZATIONPRIVATEHOST = os.getenv("AUTHORIZATIONPRIVATEHOST", AUTHORIZATIONPRIVATEHOSTDEFAULT)
    AUTHORIZATIONPRIVATEPORT = os.getenv("AUTHORIZATIONPRIVATEPORT", AUTHORIZATIONPRIVATEPORTDEFAULT)

    # KAFKA相关
    KAFKAHOST = os.getenv('KAFKAHOST', KAFKAHOSTDEFAULT)
    KAFKAPORT = os.getenv('KAFKAPORT', KAFKAPORTDEFAULT)
    KAFKAUSER = os.getenv('KAFKAUSER', KAFKAUSERDEFAULT)
    KAFKAPASS = os.getenv('KAFKAPASS', KAFKAPASSDEFAULT)

base_config = BaseConfig()
server_info = ServerInfo(
    server_name="agent-executor",
    server_version="1.0.0",
    language="python",
    python_version=sys.version,
)

observability_config = ObservabilitySetting(
    log=LogSetting(
        log_enabled=os.getenv("LOG_ENABLED", "false") == "true",
        log_exporter=os.getenv("LOG_EXPORTER", "http"),
        log_load_interval=int(os.getenv("LOG_LOAD_INTERNAL", "10")),
        log_load_max_log=int(os.getenv("LOG_LOAD_MAX_LOG", "1000")),
        http_log_feed_ingester_url=os.getenv("httpLogFeedIngesterUrl",
                                             "http://feed-ingester-service:13031/api/feed_ingester/v1/jobs/dip-o11y-log/events"),
    )
    # trace=TraceSetting(
    #     trace_enabled=os.getenv("O11Y_TRACE_ENABLED", "false") == "true",
    #     trace_provider=os.getenv("O11Y_TRACE_PROVIDER", "http"),
    #     trace_max_queue_size=int(os.getenv("O11Y_TRACE_MAX_QUEUE_SIZE", "512")),
    #     max_export_batch_size=int(os.getenv("O11Y_TRACE_MAX_EXPORT_BATCH_SIZE", "512")),
    #     http_trace_feed_ingester_url=os.getenv("O11Y_HTTP_TRACE_FEED_INGESTER_URL",
    #                                            "http://feed-ingester-service:13031/api/feed_ingester/v1/jobs/dip-o11y-trace/events"),
    # )
)
