-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 说明：
-- 1. 删除原有的 uk_catalog_source_identifier 唯一索引
-- 2. 调整anyshare connector的字段配置描述

SET SCHEMA kweaver;

-- 删除原有的 idx_t_resource_catalog_source_identifier 唯一索引
DROP INDEX idx_t_resource_catalog_source_identifier;

-- 更新anyshare连接器的doc_lib_type字段描述
UPDATE t_connector_type
SET f_field_config =
    '{
        "protocol":     {"name":"协议","type":"string","description":"http 或 https","required":true,"encrypted":false},
        "host":         {"name":"主机地址","type":"string","description":"AnyShare 服务主机","required":true,"encrypted":false},
        "port":         {"name":"端口","type":"integer","description":"服务端口","required":true,"encrypted":false},
        "auth_type":    {"name":"认证方式","type":"integer","description":"1=访问令牌 Token，2=AppID/AppSecret","required":true,"encrypted":false},
        "token":        {"name":"访问令牌","type":"string","description":"auth_type=1 时必填","required":false,"encrypted":true},
        "app_id":       {"name":"应用账户 ID","type":"string","description":"auth_type=2 时必填","required":false,"encrypted":false},
        "app_secret":   {"name":"应用密钥","type":"string","description":"auth_type=2 时必填","required":false,"encrypted":true},
        "doc_lib_type": {"name":"文档库类型","type":"integer","description":"1=知识库，2=文档库","required":true,"encrypted":false},
        "paths":        {"name":"路径列表","type":"array","description":"可选；按文档库名称路径解析起点，空则按文档库类型枚举","required":false,"encrypted":false}
    }'
WHERE f_type = 'anyshare';

