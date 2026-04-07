-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

SET SCHEMA adp;

INSERT INTO t_connector_type (f_type, f_name, f_description, f_mode, f_category, f_field_config, f_enabled)
SELECT 'postgresql', 'postgresql', 'PostgreSQL 关系型数据库连接器', 'local', 'table',
       '{
           "host":      {"name":"主机地址","type":"string","description":"数据库服务器主机地址","required":true,"encrypted":false},
           "port":      {"name":"端口号","type":"integer","description":"数据库服务器端口","required":true,"encrypted":false},
           "username":  {"name":"用户名","type":"string","description":"数据库用户名","required":true,"encrypted":false},
           "password":  {"name":"密码","type":"string","description":"数据库密码","required":true,"encrypted":true},
           "database":  {"name":"数据库名","type":"string","description":"PostgreSQL 连接目标 database","required":true,"encrypted":false},
           "schemas":   {"name":"Schema 列表","type":"array","description":"可选；为空则扫描当前库下除系统 schema 外的用户 schema；非空则仅扫描列出的 schema","required":false,"encrypted":false},
           "options":   {"name":"连接参数","type":"object","description":"连接参数（如 sslmode、connect_timeout 等）","required":false,"encrypted":false}
       }',
       1
FROM DUAL WHERE NOT EXISTS ( SELECT f_type FROM t_connector_type WHERE f_type = 'postgresql' );