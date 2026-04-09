-- AnyShare fileset connector (local)
INSERT INTO t_connector_type (f_type, f_name, f_description, f_mode, f_category, f_field_config, f_enabled)
SELECT 'anyshare', 'anyshare', 'AnyShare 连接器', 'local', 'fileset',
    '{
        "protocol":     {"name":"协议","type":"string","description":"http 或 https","required":true,"encrypted":false},
        "host":         {"name":"主机地址","type":"string","description":"AnyShare 服务主机","required":true,"encrypted":false},
        "port":         {"name":"端口","type":"integer","description":"服务端口","required":true,"encrypted":false},
        "auth_type":    {"name":"认证方式","type":"integer","description":"1=访问令牌 Token，2=AppID/AppSecret","required":true,"encrypted":false},
        "token":        {"name":"访问令牌","type":"string","description":"auth_type=1 时必填","required":false,"encrypted":true},
        "app_id":       {"name":"应用账户 ID","type":"string","description":"auth_type=2 时必填","required":false,"encrypted":false},
        "app_secret":   {"name":"应用密钥","type":"string","description":"auth_type=2 时必填","required":false,"encrypted":true},
        "doc_lib_type": {"name":"文档库类型","type":"integer","description":"1=知识库","required":true,"encrypted":false},
        "paths":        {"name":"路径列表","type":"array","description":"可选；按文档库名称路径解析起点，空则按文档库类型枚举","required":false,"encrypted":false}
    }',
    1
FROM DUAL WHERE NOT EXISTS ( SELECT f_type FROM t_connector_type WHERE f_type = 'anyshare' );
