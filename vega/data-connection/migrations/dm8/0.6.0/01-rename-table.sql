-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：在 kweaver schema 下创建 data-connection 相关表，并从 adp schema 复制数据
-- ==========================================

SET SCHEMA kweaver;
-- ==========================================
-- data_source
-- ==========================================
CREATE TABLE IF NOT EXISTS "data_source" (
    "id" varchar(36 char) NOT NULL COMMENT '主键，生成规则:36位uuid',
    "name" varchar(128 char) NOT NULL COMMENT '数据源展示名称',
    "type_name" varchar(30 char) NOT NULL COMMENT '数据库类型',
    "bin_data" blob NOT NULL COMMENT '数据源配置信息',
    "comment" varchar(255 char) DEFAULT NULL COMMENT '描述',
    "created_by_uid" varchar(36 char) NOT NULL COMMENT '创建人',
    "created_at" timestamp NOT NULL COMMENT '创建时间',
    "updated_by_uid" varchar(36 char) DEFAULT NULL COMMENT '修改人',
    "updated_at" timestamp DEFAULT NULL COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("id")
);

INSERT INTO kweaver."data_source" (
    "id",
    "name",
    "type_name",
    "bin_data",
    "comment",
    "created_by_uid",
    "created_at",
    "updated_by_uid",
    "updated_at"
)
SELECT
    "id",
    "name",
    "type_name",
    "bin_data",
    "comment",
    "created_by_uid",
    "created_at",
    "updated_by_uid",
    "updated_at"
FROM adp."data_source" s;

-- ==========================================
-- t_data_source_info
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_data_source_info" (
    "f_id" varchar(36 char) NOT NULL COMMENT '主键，生成规则:36位uuid',
    "f_name" varchar(128 char) NOT NULL COMMENT '数据源展示名称',
    "f_type" varchar(30 char) NOT NULL COMMENT '数据库类型',
    "f_catalog" varchar(50 char) COMMENT '数据源catalog名称',
    "f_database" varchar(100 char) COMMENT '数据库名称',
    "f_schema" varchar(100 char) COMMENT '数据库模式',
    "f_connect_protocol" varchar(30 char) NOT NULL COMMENT '连接方式',
    "f_host" varchar(128 char) NOT NULL COMMENT '地址',
    "f_port" int NOT NULL COMMENT '端口',
    "f_account" varchar(128 char) COMMENT '账户',
    "f_password" varchar(1024 char) COMMENT '密码',
    "f_storage_protocol" varchar(30 char) COMMENT '存储介质',
    "f_storage_base" varchar(1024 char) COMMENT '存储路径',
    "f_token" varchar(100 char) COMMENT 'token认证',
    "f_replica_set" varchar(100 char) COMMENT '副本集名称',
    "f_is_built_in" tinyint NOT NULL DEFAULT '0' COMMENT '是否为内置数据源（0 特殊 1 非内置 2 内置），默认为0',
    "f_comment" varchar(255 char) COMMENT '描述',
    "f_created_by_uid" varchar(36 char) COMMENT '创建人',
    "f_created_at" datetime(3) COMMENT '创建时间',
    "f_updated_by_uid" varchar(36 char) COMMENT '修改人',
    "f_updated_at" datetime(3) COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("f_id")
);

INSERT INTO kweaver."t_data_source_info" (
    "f_id",
    "f_name",
    "f_type",
    "f_catalog",
    "f_database",
    "f_schema",
    "f_connect_protocol",
    "f_host",
    "f_port",
    "f_account",
    "f_password",
    "f_storage_protocol",
    "f_storage_base",
    "f_token",
    "f_replica_set",
    "f_is_built_in",
    "f_comment",
    "f_created_by_uid",
    "f_created_at",
    "f_updated_by_uid",
    "f_updated_at"
)
SELECT
    "f_id",
    "f_name",
    "f_type",
    "f_catalog",
    "f_database",
    "f_schema",
    "f_connect_protocol",
    "f_host",
    "f_port",
    "f_account",
    "f_password",
    "f_storage_protocol",
    "f_storage_base",
    "f_token",
    "f_replica_set",
    "f_is_built_in",
    "f_comment",
    "f_created_by_uid",
    "f_created_at",
    "f_updated_by_uid",
    "f_updated_at"
FROM adp."t_data_source_info" s;

-- ==========================================
-- t_task_scan
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_task_scan" (
  "id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
  "type" tinyint NOT NULL DEFAULT 0 COMMENT '扫描任务：0 :即时-数据源;1 :即时-数据表;2: 定时-数据源;3: 定时-数据表',
  "name" varchar(128 char) NOT NULL COMMENT '任务名称',
  "ds_id" varchar(36 char) DEFAULT NULL COMMENT '数据源唯一标识',
  "scan_status" tinyint DEFAULT NULL COMMENT '任务状态',
  "start_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务开始时间',
  "end_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务结束时间',
  "create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  "task_params_info" text DEFAULT NULL COMMENT '任务执行参数信息',
  "task_process_info" text DEFAULT NULL COMMENT '任务执行进度信息',
  "task_result_info" text DEFAULT NULL COMMENT '任务执行结果信息',
  "schedule_id" varchar(36 char) DEFAULT NULL COMMENT '定时任务配置id',
  CLUSTER PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "t_task_scan_ds_id_IDX" on "t_task_scan" ("ds_id");

INSERT INTO kweaver."t_task_scan" (
    "id",
    "type",
    "name",
    "ds_id",
    "scan_status",
    "start_time",
    "end_time",
    "create_user",
    "task_params_info",
    "task_process_info",
    "task_result_info",
    "schedule_id"
)
SELECT
    "id",
    "type",
    "name",
    "ds_id",
    "scan_status",
    "start_time",
    "end_time",
    "create_user",
    "task_params_info",
    "task_process_info",
    "task_result_info",
    "schedule_id"
FROM adp."t_task_scan" s;

-- ==========================================
-- t_task_scan_table
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_task_scan_table" (
  "id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
  "task_id" varchar(36 char) NOT NULL COMMENT '关联任务id',
  "ds_id" varchar(36 char) NOT NULL COMMENT '数据源唯一标识',
  "ds_name" varchar(128 char) NOT NULL COMMENT '数据源名称',
  "table_id" varchar(36 char) NOT NULL COMMENT 'table的唯一id',
  "table_name" varchar(128 char) NOT NULL COMMENT 'table的name',
  "schema_name" varchar(128 char) NOT NULL COMMENT 'schema的name',
  "scan_status" tinyint DEFAULT NULL COMMENT '任务状态：0 等待;1 进行中;2 成功;3 失败',
  "start_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务开始时间',
  "end_time" datetime DEFAULT NULL COMMENT '任务结束时间',
  "create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  "scan_params" text DEFAULT NULL COMMENT '任务执行参数信息',
  "scan_result_info" text DEFAULT NULL COMMENT '任务执行结果：',
  "error_stack" text DEFAULT NULL COMMENT '异常堆栈信息',
  "operation_type" tinyint DEFAULT NULL COMMENT '操作类型:0:insert 1:delete 2:update 3:unknown',
  CLUSTER PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "t_task_scan_table_task_id_IDX" on "t_task_scan_table" ("task_id");
CREATE INDEX IF NOT EXISTS "t_task_scan_table_table_id_IDX" on "t_task_scan_table" ("table_id");

INSERT INTO kweaver."t_task_scan_table" (
    "id",
    "task_id",
    "ds_id",
    "ds_name",
    "table_id",
    "table_name",
    "schema_name",
    "scan_status",
    "start_time",
    "end_time",
    "create_user",
    "scan_params",
    "scan_result_info",
    "error_stack",
    "operation_type"
)
SELECT
    "id",
    "task_id",
    "ds_id",
    "ds_name",
    "table_id",
    "table_name",
    "schema_name",
    "scan_status",
    "start_time",
    "end_time",
    "create_user",
    "scan_params",
    "scan_result_info",
    "error_stack",
    "operation_type"
FROM adp."t_task_scan_table" s;

-- ==========================================
-- t_table_scan
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_table_scan" (
  "f_id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
  "f_name" varchar(128 char) NOT NULL COMMENT '表名称',
  "f_advanced_params" text DEFAULT NULL COMMENT '高级参数，格式为"{key(1): value(1), ... , key(n): value(n)}"',
  "f_description" varchar(2048 char) DEFAULT NULL COMMENT '表注释',
  "f_table_rows" bigint NOT NULL DEFAULT 0 COMMENT '表数据量，默认0',
  "f_data_source_id" varchar(36 char) NOT NULL COMMENT '数据源唯一标识',
  "f_data_source_name" varchar(128 char) NOT NULL COMMENT '冗余字段，数据源名称',
  "f_schema_name" varchar(128 char) NOT NULL COMMENT '冗余字段，schema名称',
  "f_task_id" varchar(36 char) NOT NULL COMMENT '关联任务id',
  "f_version" int NOT NULL DEFAULT 1 COMMENT '版本号',
  "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  "f_operation_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
  "f_operation_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
  "f_operation_type" tinyint NOT NULL DEFAULT 0 COMMENT '状态：0新增1删除2更新',
  "f_status" tinyint NOT NULL DEFAULT 3 COMMENT '任务状态：0 等待;1 进行中;2 成功;3 失败',
  "f_status_change" tinyint NOT NULL DEFAULT 0 COMMENT '状态是否发生变化：0 否1 是',
  "f_scan_source" tinyint DEFAULT NULL COMMENT '扫描来源',
CLUSTER PRIMARY KEY ("f_id")
);
CREATE INDEX IF NOT EXISTS "t_table_scan_f_data_source_id_IDX" on "t_table_scan" ("f_task_id");

INSERT INTO kweaver."t_table_scan" (
    "f_id",
    "f_name",
    "f_advanced_params",
    "f_description",
    "f_table_rows",
    "f_data_source_id",
    "f_data_source_name",
    "f_schema_name",
    "f_task_id",
    "f_version",
    "f_create_time",
    "f_create_user",
    "f_operation_time",
    "f_operation_user",
    "f_operation_type",
    "f_status",
    "f_status_change",
    "f_scan_source"
)
SELECT
    "f_id",
    "f_name",
    "f_advanced_params",
    "f_description",
    "f_table_rows",
    "f_data_source_id",
    "f_data_source_name",
    "f_schema_name",
    "f_task_id",
    "f_version",
    "f_create_time",
    "f_create_user",
    "f_operation_time",
    "f_operation_user",
    "f_operation_type",
    "f_status",
    "f_status_change",
    "f_scan_source"
FROM adp."t_table_scan" s;

-- ==========================================
-- t_table_field_scan
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_table_field_scan" (
  "f_id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
  "f_field_name" varchar(128 char) NOT NULL COMMENT '字段名',
  "f_table_id" varchar(36 char) NOT NULL COMMENT 'Table唯一标识',
  "f_table_name" varchar(128 char) NOT NULL COMMENT '表名',
  "f_field_type" varchar(128 char) DEFAULT NULL COMMENT '字段类型',
  "f_field_length" int DEFAULT NULL COMMENT '字段长度',
  "f_field_precision" int DEFAULT NULL COMMENT '字段精度',
  "f_field_comment" varchar(2048 char) DEFAULT NULL COMMENT '字段注释',
  "f_field_order_no" int DEFAULT NULL,
  "f_advanced_params" varchar(2048 char) DEFAULT NULL COMMENT '字段高级参数',
  "f_version" int NOT NULL DEFAULT 1 COMMENT '版本号',
  "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  "f_operation_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
  "f_operation_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
  "f_operation_type" tinyint NOT NULL DEFAULT 0 COMMENT '状态：0新增1删除2更新',
  "f_status_change" tinyint NOT NULL DEFAULT 0 COMMENT '状态是否发生变化：0 否1 是',
  CLUSTER PRIMARY KEY ("f_id")
);
CREATE INDEX IF NOT EXISTS "t_table_field_scan_f_table_id_IDX" on "t_table_field_scan" ("f_table_id");

INSERT INTO kweaver."t_table_field_scan" (
    "f_id",
    "f_field_name",
    "f_table_id",
    "f_table_name",
    "f_field_type",
    "f_field_length",
    "f_field_precision",
    "f_field_comment",
    "f_field_order_no",
    "f_advanced_params",
    "f_version",
    "f_create_time",
    "f_create_user",
    "f_operation_time",
    "f_operation_user",
    "f_operation_type",
    "f_status_change"
)
SELECT
    "f_id",
    "f_field_name",
    "f_table_id",
    "f_table_name",
    "f_field_type",
    "f_field_length",
    "f_field_precision",
    "f_field_comment",
    "f_field_order_no",
    "f_advanced_params",
    "f_version",
    "f_create_time",
    "f_create_user",
    "f_operation_time",
    "f_operation_user",
    "f_operation_type",
    "f_status_change"
FROM adp."t_table_field_scan" s;

-- ==========================================
-- t_task_scan_schedule
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_task_scan_schedule" (
  "id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
  "type" tinyint NOT NULL DEFAULT 0 COMMENT '扫描任务：0 :即时-数据源;1 :即时-数据表;2: 定时-数据源',
  "name" varchar(128 char) NOT NULL COMMENT '任务名称',
  "cron_expression" varchar(64 char) NOT NULL COMMENT 'cron表达式',
  "scan_strategy" varchar(64 char) DEFAULT NULL COMMENT '快速扫描策略',
  "task_status" tinyint DEFAULT 0 COMMENT '定时扫描任务:0 close 1 open',
  "ds_id" varchar(36 char) DEFAULT NULL COMMENT '数据源唯一标识',
  "create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  "create_user" varchar(64 char) DEFAULT NULL COMMENT '创建用户',
  "operation_time" datetime DEFAULT NULL,
  "operation_user" varchar(64 char) DEFAULT NULL,
  "operation_type" tinyint NOT NULL DEFAULT 0 COMMENT '状态：0新增1删除2更新',
  CLUSTER PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "t_task_scan_ds_id_IDX" on "t_task_scan_schedule" ("ds_id");

INSERT INTO kweaver."t_task_scan_schedule" (
    "id",
    "type",
    "name",
    "cron_expression",
    "scan_strategy",
    "task_status",
    "ds_id",
    "create_time",
    "create_user",
    "operation_time",
    "operation_user",
    "operation_type"
)
SELECT
    "id",
    "type",
    "name",
    "cron_expression",
    "scan_strategy",
    "task_status",
    "ds_id",
    "create_time",
    "create_user",
    "operation_time",
    "operation_user",
    "operation_type"
FROM adp."t_task_scan_schedule" s;

-- ==========================================
-- t_data_quality_model
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_data_quality_model" (
                                                      "f_id" bigint NOT NULL COMMENT '主键id，唯一标识',
                                                      "f_ds_id" bigint NOT NULL COMMENT '数据源ID',
                                                      "f_dolphinscheduler_ds_id" bigint NOT NULL COMMENT 'dolphinscheduler数据源ID',
                                                      "f_db_type" varchar(50 char) NOT NULL COMMENT '数据库类型',
    "f_tb_name" varchar(512 char) NOT NULL COMMENT '表名称',
    "f_process_definition_code" bigint NOT NULL COMMENT '工作流定义ID',
    "f_crontab" varchar(128 char) DEFAULT NULL COMMENT '定时任务表达式',
    CLUSTER PRIMARY KEY ("f_id")
    );

INSERT INTO kweaver."t_data_quality_model" (
    "f_id",
    "f_ds_id",
    "f_dolphinscheduler_ds_id",
    "f_db_type",
    "f_tb_name",
    "f_process_definition_code",
    "f_crontab"
)
SELECT
    "f_id",
    "f_ds_id",
    "f_dolphinscheduler_ds_id",
    "f_db_type",
    "f_tb_name",
    "f_process_definition_code",
    "f_crontab"
FROM adp."t_data_quality_model" s;

-- ==========================================
-- t_data_quality_rule
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_data_quality_rule" (
                                                     "f_id" bigint NOT NULL COMMENT '主键id，唯一标识',
                                                     "f_field_name" varchar(512 char) NOT NULL COMMENT '字段名称',
    "f_rule_id" tinyint NOT NULL COMMENT '质量规则ID：1-空值检测，，2-自定义SQL，5-字段长度校验，6-唯一性校验，7-正则表达式，9-枚举值校验，10-表行数校验',
    "f_threshold" double DEFAULT NULL COMMENT '阈值，默认0',
    "f_check_val" varchar(10240 char) DEFAULT NULL COMMENT '1、自定义sql：填写sql语句；2、字段长度校验：填写字段长度；3、正则表达式：填写正则表达式；4、枚举值校验：填写枚举值，逗号分割；5、表行数校验：填写表行数。',
    "f_check_val_name" varchar(128 char) DEFAULT NULL COMMENT '自定义sql时，填写的实际值名',
    "f_model_id" bigint NOT NULL COMMENT '质量模型ID',
    CLUSTER PRIMARY KEY ("f_id")
);

INSERT INTO kweaver."t_data_quality_rule" (
    "f_id",
    "f_field_name",
    "f_rule_id",
    "f_threshold",
    "f_check_val",
    "f_check_val_name",
    "f_model_id"
)
SELECT
    "f_id",
    "f_field_name",
    "f_rule_id",
    "f_threshold",
    "f_check_val",
    "f_check_val_name",
    "f_model_id"
FROM adp."t_data_quality_rule" s;

-- ==========================================
-- t_data_source
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_data_source" (
    "f_id" varchar(36 char) NOT NULL COMMENT '唯一id，雪花算法',
    "f_name" varchar(128 char) NOT NULL COMMENT '数据源名称',
    "f_data_source_type" tinyint NOT NULL COMMENT '类型，关联字典表f_dict_type为1时的f_dict_key',
    "f_data_source_type_name" varchar(256 char) NOT NULL COMMENT '类型名称，对应字典表f_dict_type为1时的f_dict_value',
    "f_user_name" varchar(128 char) NOT NULL COMMENT '用户名',
    "f_password" varchar(1024 char) NOT NULL COMMENT '密码',
    "f_description" varchar(255 char) NOT NULL DEFAULT '' COMMENT '描述',
    "f_extend_property" varchar(255 char) NOT NULL DEFAULT '' COMMENT '扩展属性，默认为空字符串',
    "f_host" varchar(128 char) NOT NULL COMMENT 'HOST',
    "f_port" int NOT NULL COMMENT '端口',
    "f_enable_status" tinyint NOT NULL DEFAULT 1 COMMENT '禁用/启用状态，1 启用，2 停用，默认为启用',
    "f_connect_status" tinyint NOT NULL DEFAULT 1 COMMENT '连接状态，1 成功，2 失败，默认为成功',
    "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
    "f_update_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
    "f_update_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
    "f_database" varchar(100 char) DEFAULT NULL COMMENT '数据库名称',
    "f_info_system_id" varchar(128 char) DEFAULT NULL COMMENT '信息系统id',
    "f_dolphin_id" bigint DEFAULT NULL COMMENT 'dolphin数据元id',
    "f_delete_code" bigint DEFAULT 0 COMMENT '逻辑删除标识码',
    "f_live_update_status" tinyint NOT NULL DEFAULT 0 COMMENT '实时更新标识（0无需更新，1待更新，2更新中，3连接不可用，4无权限，5待广播',
    "f_live_update_benchmark" varchar(255 char) DEFAULT NULL COMMENT '实时更新基准',
    "f_live_update_time" datetime DEFAULT current_timestamp() COMMENT '实时更新时间',
    CLUSTER PRIMARY KEY ("f_id")
    );
CREATE UNIQUE INDEX IF NOT EXISTS "t_data_source_un" on "t_data_source" ("f_name","f_create_user","f_info_system_id","f_delete_code");

INSERT INTO kweaver."t_data_source" (
    "f_id",
    "f_name",
    "f_data_source_type",
    "f_data_source_type_name",
    "f_user_name",
    "f_password",
    "f_description",
    "f_extend_property",
    "f_host",
    "f_port",
    "f_enable_status",
    "f_connect_status",
    "f_authority_id",
    "f_create_time",
    "f_create_user",
    "f_update_time",
    "f_update_user",
    "f_database",
    "f_info_system_id",
    "f_dolphin_id",
    "f_delete_code",
    "f_live_update_status",
    "f_live_update_benchmark",
    "f_live_update_time"
)
SELECT
    "f_id",
    "f_name",
    "f_data_source_type",
    "f_data_source_type_name",
    "f_user_name",
    "f_password",
    "f_description",
    "f_extend_property",
    "f_host",
    "f_port",
    "f_enable_status",
    "f_connect_status",
    "f_authority_id",
    "f_create_time",
    "f_create_user",
    "f_update_time",
    "f_update_user",
    "f_database",
    "f_info_system_id",
    "f_dolphin_id",
    "f_delete_code",
    "f_live_update_status",
    "f_live_update_benchmark",
    "f_live_update_time"
FROM adp."t_data_source" s;

-- ==========================================
-- t_dict
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_dict" (
                                        "f_id" int NOT NULL IDENTITY(1,1) COMMENT '唯一id，自增ID',
    "f_dict_type" tinyint NOT NULL COMMENT '字典类型\n1：数据源类型\n2：Oracle字段类型\n3：MySQL字段类型\n4：PostgreSQL字段类型\n5：SqlServer字段类型\n6：Hive字段类型\n7：HBase字段类型\n8：MongoDB字段类型\n9：FTP字段类型\n10：HDFS字段类型\n11：SFTP字段类型\n12：CMQ字段类型\n13：Kafka字段类型\n14：API字段类型',
    "f_dict_key" tinyint NOT NULL COMMENT '枚举值',
    "f_dict_value" varchar(256 char) NOT NULL COMMENT '枚举对应描述',
    "f_extend_property" varchar(1024 char) NOT NULL COMMENT '扩展属性',
    "f_enable_status" tinyint NOT NULL DEFAULT 2 COMMENT '启用状态，1 启用，2 停用，默认为停用',
    CLUSTER PRIMARY KEY ("f_id")
    );
CREATE UNIQUE INDEX IF NOT EXISTS "t_dict_un" on "t_dict" ("f_dict_type","f_dict_key");

SET IDENTITY_INSERT kweaver.t_dict ON;

INSERT INTO kweaver."t_dict" (
    "f_id",
    "f_dict_type",
    "f_dict_key",
    "f_dict_value",
    "f_extend_property",
    "f_enable_status"
)
SELECT
    "f_id",
    "f_dict_type",
    "f_dict_key",
    "f_dict_value",
    "f_extend_property",
    "f_enable_status"
FROM adp."t_dict" s;

SET IDENTITY_INSERT kweaver.t_dict OFF;

-- ==========================================
-- t_indicator
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_indicator" (
                                             "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                             "f_indicator_name" varchar(128 char) NOT NULL COMMENT '指标名称',
    "f_indicator_type" varchar(128 char) NOT NULL COMMENT '指标类型',
    "f_indicator_value" bigint NOT NULL COMMENT '指标数值',
    "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    "f_indicator_object_id" bigint DEFAULT NULL COMMENT '关联对象ID',
    "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    "f_advanced_params" varchar(255 char) NOT NULL DEFAULT '[]' COMMENT '指标高级参数',
    CLUSTER PRIMARY KEY ("f_id","f_create_time")
    );

INSERT INTO kweaver."t_indicator" (
    "f_id",
    "f_indicator_name",
    "f_indicator_type",
    "f_indicator_value",
    "f_create_time",
    "f_indicator_object_id",
    "f_authority_id",
    "f_advanced_params"
)
SELECT
    "f_id",
    "f_indicator_name",
    "f_indicator_type",
    "f_indicator_value",
    "f_create_time",
    "f_indicator_object_id",
    "f_authority_id",
    "f_advanced_params"
FROM adp."t_indicator" s;

-- ==========================================
-- t_lineage_edge_column
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_lineage_edge_column" (
    "f_id" varchar(64 char) NOT NULL COMMENT '主键ID，根据f_table_id和f_column_id值MD5计算得到',
    "f_parent_id" varchar(64 char) NOT NULL COMMENT '源字段ID',
    "f_child_id" varchar(64 char) NOT NULL COMMENT '目标字段ID',
    "f_create_type" varchar(20 char) DEFAULT NULL COMMENT '创建类型： HIVE/DATAX/SPARK/USER_REPORT',
    "f_query_text" text DEFAULT NULL COMMENT '生成血缘的sql或者脚本说明',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "deleted_at" bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    "f_create_time" timestamp NULL DEFAULT NULL COMMENT '创建时间，时间戳',
    CLUSTER PRIMARY KEY ("f_id")
    );

INSERT INTO kweaver."t_lineage_edge_column" (
    "f_id",
    "f_parent_id",
    "f_child_id",
    "f_create_type",
    "f_query_text",
    "created_at",
    "updated_at",
    "deleted_at",
    "f_create_time"
)
SELECT
    "f_id",
    "f_parent_id",
    "f_child_id",
    "f_create_type",
    "f_query_text",
    "created_at",
    "updated_at",
    "deleted_at",
    "f_create_time"
FROM adp."t_lineage_edge_column" s;

-- ==========================================
-- t_lineage_edge_column_table_relation
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_lineage_edge_column_table_relation" (
    "f_id" varchar(64 char) NOT NULL COMMENT '主键ID，根据f_table_id和f_column_id值MD5计算得到',
    "f_table_id" varchar(64 char) NOT NULL COMMENT '表ID',
    "f_column_id" varchar(64 char) NOT NULL COMMENT '字段ID',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "deleted_at" bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    CLUSTER PRIMARY KEY ("f_id")
    );

INSERT INTO kweaver."t_lineage_edge_column_table_relation" (
    "f_id",
    "f_table_id",
    "f_column_id",
    "created_at",
    "updated_at",
    "deleted_at"
)
SELECT
    "f_id",
    "f_table_id",
    "f_column_id",
    "created_at",
    "updated_at",
    "deleted_at"
FROM adp."t_lineage_edge_column_table_relation" s;

-- ==========================================
-- t_lineage_edge_table
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_lineage_edge_table" (
    "f_id" varchar(64 char) NOT NULL COMMENT '主键ID，根据f_table_id和f_column_id值MD5计算得到',
    "f_parent_id" varchar(64 char) NOT NULL COMMENT '源ID',
    "f_child_id" varchar(64 char) NOT NULL COMMENT '目标ID',
    "f_create_type" varchar(20 char) DEFAULT NULL COMMENT '创建类型： HIVE/DATAX/SPARK/USER_REPORT',
    "f_query_text" text DEFAULT NULL COMMENT '生成血缘的sql或者脚本说明',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "deleted_at" bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    "f_create_time" timestamp NULL DEFAULT NULL COMMENT '创建时间，时间戳',
    CLUSTER PRIMARY KEY ("f_id")
    );

INSERT INTO kweaver."t_lineage_edge_table" (
    "f_id",
    "f_parent_id",
    "f_child_id",
    "f_create_type",
    "f_query_text",
    "created_at",
    "updated_at",
    "deleted_at",
    "f_create_time"
)
SELECT
    "f_id",
    "f_parent_id",
    "f_child_id",
    "f_create_type",
    "f_query_text",
    "created_at",
    "updated_at",
    "deleted_at",
    "f_create_time"
FROM adp."t_lineage_edge_table" s;

-- ==========================================
-- t_lineage_graph_info
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_lineage_graph_info" (
    "app_id" varchar(20 char) NOT NULL COMMENT '图谱appId',
    "graph_id" bigint DEFAULT NULL COMMENT '图谱graphId',
    CLUSTER PRIMARY KEY ("app_id")
    );

INSERT INTO kweaver."t_lineage_graph_info" (
    "app_id",
    "graph_id"
)
SELECT
    "app_id",
    "graph_id"
FROM adp."t_lineage_graph_info" s;

-- ==========================================
-- t_lineage_log
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_lineage_log" (
    "id" varchar(36 char) NOT NULL DEFAULT LOWER(RAWTOHEX(SYS_GUID())),
    "class_id" varchar(36 char) NOT NULL COMMENT '实体的主键id',
    "class_type" varchar(36 char) NOT NULL COMMENT '实体类型',
    "action_type" varchar(10 char) NOT NULL COMMENT '操作类型：insert update delete',
    "class_data" text NOT NULL COMMENT '血缘实体json',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    CLUSTER PRIMARY KEY ("id")
    );

INSERT INTO kweaver."t_lineage_log" (
    "id",
    "class_id",
    "class_type",
    "action_type",
    "class_data",
    "created_at"
)
SELECT
    "id",
    "class_id",
    "class_type",
    "action_type",
    "class_data",
    "created_at"
FROM adp."t_lineage_log" s;

-- ==========================================
-- t_lineage_relation
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_lineage_relation" (
    "unique_id" varchar(255 char) NOT NULL COMMENT '实体ID',
    "class_type" tinyint DEFAULT NULL COMMENT '类型，1:column,2:indicator',
    "parent" text DEFAULT NULL COMMENT '上一个节点',
    "child" text DEFAULT NULL COMMENT '下一个节点',
    "created_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("unique_id")
    );

INSERT INTO kweaver."t_lineage_relation" (
    "unique_id",
    "class_type",
    "parent",
    "child",
    "created_at",
    "updated_at"
)
SELECT
    "unique_id",
    "class_type",
    "parent",
    "child",
    "created_at",
    "updated_at"
FROM adp."t_lineage_relation" s;

-- ==========================================
-- t_lineage_tag_column
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_lineage_tag_column" (
    "f_id" varchar(64 char) NOT NULL COMMENT '主键ID，根据f_table_id和f_column值MD5计算得到',
    "f_table_id" varchar(64 char) NOT NULL COMMENT 't_lineage_tag_table表ID',
    "f_column" varchar(255 char) NOT NULL COMMENT '字段名称',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "deleted_at" bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    CLUSTER PRIMARY KEY ("f_id")
    );

INSERT INTO kweaver."t_lineage_tag_column" (
    "f_id",
    "f_table_id",
    "f_column",
    "created_at",
    "updated_at",
    "deleted_at"
)
SELECT
    "f_id",
    "f_table_id",
    "f_column",
    "created_at",
    "updated_at",
    "deleted_at"
FROM adp."t_lineage_tag_column" s;

-- ==========================================
-- t_lineage_tag_table
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_lineage_tag_table" (
    "f_id" varchar(64 char) NOT NULL COMMENT '主键ID，根据f_db_type、f_ds_id、f_jdbc_url、f_jdbc_user、f_db_name、f_db_schema、f_tb_name值MD5计算得到',
    "f_db_type" varchar(64 char) NOT NULL COMMENT '数据库类型',
    "f_ds_id" varchar(64 char) DEFAULT NULL COMMENT '数据源ID',
    "f_jdbc_url" varchar(255 char) DEFAULT NULL COMMENT '数据库连接URL',
    "f_jdbc_user" varchar(255 char) DEFAULT NULL COMMENT '数据库JDBC 用户名',
    "f_db_name" varchar(255 char) DEFAULT NULL COMMENT '数据库名称',
    "f_db_schema" varchar(255 char) DEFAULT NULL COMMENT '模式名称',
    "f_tb_name" varchar(255 char) NOT NULL COMMENT '表名称',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "deleted_at" bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    CLUSTER PRIMARY KEY ("f_id")
    );

INSERT INTO kweaver."t_lineage_tag_table" (
    "f_id",
    "f_db_type",
    "f_ds_id",
    "f_jdbc_url",
    "f_jdbc_user",
    "f_db_name",
    "f_db_schema",
    "f_tb_name",
    "created_at",
    "updated_at",
    "deleted_at"
)
SELECT
    "f_id",
    "f_db_type",
    "f_ds_id",
    "f_jdbc_url",
    "f_jdbc_user",
    "f_db_name",
    "f_db_schema",
    "f_tb_name",
    "created_at",
    "updated_at",
    "deleted_at"
FROM adp."t_lineage_tag_table" s;

-- ==========================================
-- t_indicator2
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_indicator2" (
                                              "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                              "f_indicator_name" varchar(128 char) NOT NULL COMMENT '指标名称',
    "f_indicator_type" varchar(128 char) NOT NULL COMMENT '指标类型',
    "f_indicator_value" bigint NOT NULL COMMENT '指标数值',
    "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    "f_indicator_object_id" bigint DEFAULT NULL COMMENT '关联对象ID',
    "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    "f_advanced_params" varchar(255 char) NOT NULL DEFAULT '[]' COMMENT '指标高级参数',
    CLUSTER PRIMARY KEY ("f_id","f_create_time")
    );

INSERT INTO kweaver."t_indicator2" (
    "f_id",
    "f_indicator_name",
    "f_indicator_type",
    "f_indicator_value",
    "f_create_time",
    "f_indicator_object_id",
    "f_authority_id",
    "f_advanced_params"
)
SELECT
    "f_id",
    "f_indicator_name",
    "f_indicator_type",
    "f_indicator_value",
    "f_create_time",
    "f_indicator_object_id",
    "f_authority_id",
    "f_advanced_params"
FROM adp."t_indicator2" s;

-- ==========================================
-- t_lineage_tag_column2
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_lineage_tag_column2" (
    "unique_id" varchar(255 char) NOT NULL COMMENT '列的唯一id',
    "uuid" varchar(36 char) DEFAULT NULL COMMENT '字段的uuid',
    "technical_name" varchar(255 char) DEFAULT NULL COMMENT '列技术名称',
    "business_name" varchar(255 char) DEFAULT NULL COMMENT '列业务名称',
    "comment" varchar(300 char) DEFAULT NULL COMMENT '字段注释',
    "data_type" varchar(255 char) DEFAULT NULL COMMENT '字段的数据类型',
    "primary_key" tinyint DEFAULT NULL COMMENT '是否主键',
    "table_unique_id" varchar(36 char) DEFAULT NULL COMMENT '属于血缘表的uuid',
    "expression_name" text DEFAULT NULL COMMENT 'column的生成表达式',
    "column_unique_ids" varchar(1024 char) DEFAULT '' COMMENT 'column的生成依赖的column的uid',
    "action_type" varchar(10 char) DEFAULT NULL COMMENT '操作类型:insertupdatedelete',
    "created_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("unique_id")
    );

INSERT INTO kweaver."t_lineage_tag_column2" (
    "unique_id",
    "uuid",
    "technical_name",
    "business_name",
    "comment",
    "data_type",
    "primary_key",
    "table_unique_id",
    "expression_name",
    "column_unique_ids",
    "action_type",
    "created_at",
    "updated_at"
)
SELECT
    "unique_id",
    "uuid",
    "technical_name",
    "business_name",
    "comment",
    "data_type",
    "primary_key",
    "table_unique_id",
    "expression_name",
    "column_unique_ids",
    "action_type",
    "created_at",
    "updated_at"
FROM adp."t_lineage_tag_column2" s;

-- ==========================================
-- t_lineage_tag_indicator2
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_lineage_tag_indicator2" (
    "uuid" varchar(36 char) NOT NULL COMMENT '指标的uuid',
    "name" varchar(128 char) NOT NULL COMMENT '指标名称',
    "description" varchar(300 char) DEFAULT NULL COMMENT '指标名称描述',
    "code" varchar(128 char) NOT NULL COMMENT '指标编号',
    "indicator_type" varchar(10 char) NOT NULL COMMENT '指标类型:atomic原子derived衍生composite复合',
    "expression" text DEFAULT NULL COMMENT '指标表达式，如果指标是原子或复合指标时',
    "indicator_uuids" varchar(1024 char) DEFAULT '' COMMENT '引用的指标uuid',
    "time_restrict" text DEFAULT NULL COMMENT '时间限定表达式，如果指标是衍生指标时',
    "modifier_restrict" text DEFAULT NULL COMMENT '普通限定表达式，如果指标是衍生指标时',
    "owner_uid" varchar(50 char) DEFAULT NULL COMMENT '数据ownerID',
    "owner_name" varchar(128 char) DEFAULT NULL COMMENT '数据owner名称',
    "department_id" varchar(36 char) DEFAULT NULL COMMENT '所属部门id',
    "department_name" varchar(128 char) DEFAULT NULL COMMENT '所属部门名称',
    "column_unique_ids" varchar(1024 char) DEFAULT '' COMMENT '依赖的字段的unique_id',
    "action_type" varchar(10 char) NOT NULL COMMENT '操作类型:insertupdatedelete',
    "created_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) DEFAULT current_timestamp(3) COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("uuid")
    );

INSERT INTO kweaver."t_lineage_tag_indicator2" (
    "uuid",
    "name",
    "description",
    "code",
    "indicator_type",
    "expression",
    "indicator_uuids",
    "time_restrict",
    "modifier_restrict",
    "owner_uid",
    "owner_name",
    "department_id",
    "department_name",
    "column_unique_ids",
    "action_type",
    "created_at",
    "updated_at"
)
SELECT
    "uuid",
    "name",
    "description",
    "code",
    "indicator_type",
    "expression",
    "indicator_uuids",
    "time_restrict",
    "modifier_restrict",
    "owner_uid",
    "owner_name",
    "department_id",
    "department_name",
    "column_unique_ids",
    "action_type",
    "created_at",
    "updated_at"
FROM adp."t_lineage_tag_indicator2" s;

-- ==========================================
-- t_lineage_tag_table2
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_lineage_tag_table2" (
    "unique_id" varchar(255 char) NOT NULL COMMENT '唯一id',
    "uuid" varchar(36 char) NOT NULL COMMENT '表的uuid',
    "technical_name" varchar(255 char) NOT NULL COMMENT '表技术名称',
    "business_name" varchar(255 char) DEFAULT NULL COMMENT '表业务名称',
    "comment" varchar(300 char) DEFAULT NULL COMMENT '表注释',
    "table_type" varchar(36 char) NOT NULL COMMENT '表类型',
    "datasource_id" varchar(36 char) DEFAULT NULL COMMENT '数据源id',
    "datasource_name" varchar(255 char) DEFAULT NULL COMMENT '数据源名称',
    "owner_id" varchar(36 char) DEFAULT NULL COMMENT '数据Ownerid',
    "owner_name" varchar(128 char) DEFAULT NULL COMMENT '数据OwnerName',
    "department_id" varchar(36 char) DEFAULT NULL COMMENT '所属部门id',
    "department_name" varchar(128 char) DEFAULT NULL COMMENT '所属部门mame',
    "info_system_id" varchar(36 char) DEFAULT NULL COMMENT '信息系统id',
    "info_system_name" varchar(128 char) DEFAULT NULL COMMENT '信息系统名称',
    "database_name" varchar(128 char) NOT NULL COMMENT '数据库名称',
    "catalog_name" varchar(255 char) NOT NULL DEFAULT '' COMMENT '数据源catalog名称',
    "catalog_addr" varchar(1024 char) NOT NULL DEFAULT '' COMMENT '数据源地址',
    "catalog_type" varchar(128 char) NOT NULL COMMENT '数据库类型名称',
    "task_execution_info" varchar(128 char) DEFAULT NULL COMMENT '表加工任务的相关名称',
    "action_type" varchar(10 char) NOT NULL COMMENT '操作类型:insertupdatedelete',
    "created_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    "updated_at" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '更新时间',
    CLUSTER PRIMARY KEY ("unique_id")
    );

INSERT INTO kweaver."t_lineage_tag_table2" (
    "unique_id",
    "uuid",
    "technical_name",
    "business_name",
    "comment",
    "table_type",
    "datasource_id",
    "datasource_name",
    "owner_id",
    "owner_name",
    "department_id",
    "department_name",
    "info_system_id",
    "info_system_name",
    "database_name",
    "catalog_name",
    "catalog_addr",
    "catalog_type",
    "task_execution_info",
    "action_type",
    "created_at",
    "updated_at"
)
SELECT
    "unique_id",
    "uuid",
    "technical_name",
    "business_name",
    "comment",
    "table_type",
    "datasource_id",
    "datasource_name",
    "owner_id",
    "owner_name",
    "department_id",
    "department_name",
    "info_system_id",
    "info_system_name",
    "database_name",
    "catalog_name",
    "catalog_addr",
    "catalog_type",
    "task_execution_info",
    "action_type",
    "created_at",
    "updated_at"
FROM adp."t_lineage_tag_table2" s;

-- ==========================================
-- t_live_ddl
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_live_ddl" (
                                            "f_id" bigint NOT NULL IDENTITY(1,1) COMMENT '唯一标识',
    "f_data_source_id" bigint NOT NULL DEFAULT 0 COMMENT '数据源ID',
    "f_data_source_name" varchar(255 char) NOT NULL DEFAULT '' COMMENT '数据源名称',
    "f_origin_catalog" varchar(255 char) DEFAULT NULL COMMENT '物理catalog',
    "f_virtual_catalog" varchar(255 char) DEFAULT NULL COMMENT '虚拟化catalog',
    "f_schema_id" bigint DEFAULT NULL COMMENT 'schemaID',
    "f_schema_name" varchar(255 char) DEFAULT NULL COMMENT 'schema名称',
    "f_table_id" bigint DEFAULT NULL COMMENT 'tableID',
    "f_table_name" varchar(255 char) DEFAULT NULL COMMENT 'table名称',
    "f_sql_type" varchar(100 char) DEFAULT NULL COMMENT 'sql类型(AlterTable,AlterColumn,CreateTable,CommentTable,CommentColumn,DropTable,RenameTable)',
    "f_sql_text" text NOT NULL COMMENT 'sql文本',
    "f_live_update_benchmark" varchar(255 char) NOT NULL DEFAULT '' COMMENT '实时更新基准',
    "f_monitor_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '监听时间，默认当前时间',
    "f_update_status" tinyint DEFAULT NULL COMMENT '更新状态（0全量更新，1增量更新，2忽略更新，3待更新，4解析失败，5更新失败）',
    "f_update_message" varchar(2000 char) DEFAULT NULL COMMENT '更新信息',
    "f_push_status" tinyint DEFAULT NULL COMMENT '0不推送,1待推送,2已推送',
    CLUSTER PRIMARY KEY ("f_id")
    );

SET IDENTITY_INSERT kweaver.t_live_ddl ON;

INSERT INTO kweaver."t_live_ddl" (
    "f_id",
    "f_data_source_id",
    "f_data_source_name",
    "f_origin_catalog",
    "f_virtual_catalog",
    "f_schema_id",
    "f_schema_name",
    "f_table_id",
    "f_table_name",
    "f_sql_type",
    "f_sql_text",
    "f_live_update_benchmark",
    "f_monitor_time",
    "f_update_status",
    "f_update_message",
    "f_push_status"
)
SELECT
    "f_id",
    "f_data_source_id",
    "f_data_source_name",
    "f_origin_catalog",
    "f_virtual_catalog",
    "f_schema_id",
    "f_schema_name",
    "f_table_id",
    "f_table_name",
    "f_sql_type",
    "f_sql_text",
    "f_live_update_benchmark",
    "f_monitor_time",
    "f_update_status",
    "f_update_message",
    "f_push_status"
FROM adp."t_live_ddl" s;

SET IDENTITY_INSERT kweaver.t_live_ddl OFF;

-- ==========================================
-- t_schema
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_schema" (
                                          "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                          "f_name" varchar(128 char) NOT NULL COMMENT 'schema名称',
    "f_data_source_id" varchar(36 char) NOT NULL COMMENT '数据源唯一标识',
    "f_data_source_name" varchar(128 char) NOT NULL COMMENT '冗余字段，数据源名称',
    "f_data_source_type" tinyint NOT NULL COMMENT '冗余字段，数据源类型，关联字典表f_dict_type为1时的f_dict_key',
    "f_data_source_type_name" varchar(256 char) NOT NULL COMMENT '冗余字段，数据源类型名称，对应字典表f_dict_type为1时的f_dict_value',
    "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
    "f_update_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
    "f_update_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
    CLUSTER PRIMARY KEY ("f_id")
    );
CREATE UNIQUE INDEX IF NOT EXISTS "t_schema_un" on "t_schema" ("f_data_source_id","f_name");

INSERT INTO kweaver."t_schema" (
    "f_id",
    "f_name",
    "f_data_source_id",
    "f_data_source_name",
    "f_data_source_type",
    "f_data_source_type_name",
    "f_authority_id",
    "f_create_time",
    "f_create_user",
    "f_update_time",
    "f_update_user"
)
SELECT
    "f_id",
    "f_name",
    "f_data_source_id",
    "f_data_source_name",
    "f_data_source_type",
    "f_data_source_type_name",
    "f_authority_id",
    "f_create_time",
    "f_create_user",
    "f_update_time",
    "f_update_user"
FROM adp."t_schema" s;

-- ==========================================
-- t_table
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_table" (
                                         "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                         "f_name" varchar(128 char) NOT NULL COMMENT '表名称',
    "f_advanced_params" text NOT NULL COMMENT '高级参数，默认为"{}"，格式为"{key(1): value(1), ... , key(n): value(n)}"',
    "f_description" varchar(2048 char) DEFAULT NULL COMMENT '表注释',
    "f_table_rows" bigint NOT NULL DEFAULT 0 COMMENT '表数据量，默认0',
    "f_schema_id" bigint NOT NULL COMMENT 'schema唯一标识',
    "f_schema_name" varchar(128 char) NOT NULL COMMENT '冗余字段，schema名称',
    "f_data_source_id" varchar(36 char) NOT NULL COMMENT '数据源唯一标识',
    "f_data_source_name" varchar(128 char) NOT NULL COMMENT '冗余字段，数据源名称',
    "f_data_source_type" tinyint NOT NULL COMMENT '冗余字段，数据源类型，关联字典表f_dict_type为1时的f_dict_key',
    "f_data_source_type_name" varchar(256 char) NOT NULL COMMENT '冗余字段，数据源类型名称，对应字典表f_dict_type为1时的f_dict_value',
    "f_version" int NOT NULL DEFAULT 1 COMMENT '版本号',
    "f_authority_id" varchar(100 char) NOT NULL DEFAULT '' COMMENT '权限域（目前为预留字段），默认0',
    "f_create_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
    "f_update_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
    "f_update_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
    "f_delete_flag" tinyint NOT NULL DEFAULT 0 COMMENT '逻辑删除标识',
    "f_delete_time" datetime DEFAULT NULL COMMENT '逻辑删除时间',
    "f_scan_source" tinyint DEFAULT NULL  COMMENT '扫描来源',
    CLUSTER PRIMARY KEY ("f_id")
    );
CREATE UNIQUE INDEX IF NOT EXISTS "t_table_un" on "t_table" ("f_data_source_id","f_schema_id","f_name");

INSERT INTO kweaver."t_table" (
    "f_id",
    "f_name",
    "f_advanced_params",
    "f_description",
    "f_table_rows",
    "f_schema_id",
    "f_schema_name",
    "f_data_source_id",
    "f_data_source_name",
    "f_data_source_type",
    "f_data_source_type_name",
    "f_version",
    "f_authority_id",
    "f_create_time",
    "f_create_user",
    "f_update_time",
    "f_update_user",
    "f_delete_flag",
    "f_delete_time",
    "f_scan_source"
)
SELECT
    "f_id",
    "f_name",
    "f_advanced_params",
    "f_description",
    "f_table_rows",
    "f_schema_id",
    "f_schema_name",
    "f_data_source_id",
    "f_data_source_name",
    "f_data_source_type",
    "f_data_source_type_name",
    "f_version",
    "f_authority_id",
    "f_create_time",
    "f_create_user",
    "f_update_time",
    "f_update_user",
    "f_delete_flag",
    "f_delete_time",
    "f_scan_source"
FROM adp."t_table" s;

-- ==========================================
-- t_table_field
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_table_field" (
                                               "f_table_id" bigint NOT NULL COMMENT 'Table唯一标识',
                                               "f_field_name" varchar(128 char) NOT NULL COMMENT '字段名',
    "f_field_type" varchar(128 char) DEFAULT NULL COMMENT '字段类型',
    "f_field_length" int DEFAULT NULL COMMENT '字段长度',
    "f_field_precision" int DEFAULT NULL COMMENT '字段精度',
    "f_field_comment" varchar(2048 char) DEFAULT NULL COMMENT '字段注释',
    "f_advanced_params" varchar(2048 char) NOT NULL DEFAULT '[]' COMMENT '字段高级参数',
    "f_update_flag" tinyint NOT NULL DEFAULT 0 COMMENT '更新标识',
    "f_update_time" datetime DEFAULT NULL COMMENT '更新时间',
    "f_delete_flag" tinyint NOT NULL DEFAULT 0 COMMENT '逻辑删除标识',
    "f_delete_time" datetime DEFAULT NULL COMMENT '逻辑删除时间',
    CLUSTER PRIMARY KEY ("f_table_id","f_field_name")
    );

INSERT INTO kweaver."t_table_field" (
    "f_table_id",
    "f_field_name",
    "f_field_type",
    "f_field_length",
    "f_field_precision",
    "f_field_comment",
    "f_advanced_params",
    "f_update_flag",
    "f_update_time",
    "f_delete_flag",
    "f_delete_time"
)
SELECT
    "f_table_id",
    "f_field_name",
    "f_field_type",
    "f_field_length",
    "f_field_precision",
    "f_field_comment",
    "f_advanced_params",
    "f_update_flag",
    "f_update_time",
    "f_delete_flag",
    "f_delete_time"
FROM adp."t_table_field" s;

-- ==========================================
-- t_table_field_his
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_table_field_his" (
                                                   "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                                   "f_field_name" varchar(128 char) NOT NULL COMMENT '字段名',
    "f_field_type" varchar(128 char) DEFAULT NULL COMMENT '字段类型',
    "f_field_length" int DEFAULT NULL COMMENT '字段长度',
    "f_field_precision" int DEFAULT NULL COMMENT '字段精度',
    "f_field_comment" varchar(2048 char) DEFAULT NULL COMMENT '字段注释',
    "f_table_id" bigint NOT NULL COMMENT 'Table唯一标识',
    "f_version" int NOT NULL DEFAULT 1 COMMENT '版本号',
    "f_advanced_params" varchar(255 char) NOT NULL DEFAULT '[]' COMMENT '字段高级参数',
    CLUSTER PRIMARY KEY ("f_id","f_version")
    );

INSERT INTO kweaver."t_table_field_his" (
    "f_id",
    "f_field_name",
    "f_field_type",
    "f_field_length",
    "f_field_precision",
    "f_field_comment",
    "f_table_id",
    "f_version",
    "f_advanced_params"
)
SELECT
    "f_id",
    "f_field_name",
    "f_field_type",
    "f_field_length",
    "f_field_precision",
    "f_field_comment",
    "f_table_id",
    "f_version",
    "f_advanced_params"
FROM adp."t_table_field_his" s;

-- ==========================================
-- t_task
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_task" (
                                        "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                        "f_object_id" varchar(36 char) DEFAULT NULL COMMENT '任务对象id',
    "f_object_type" tinyint DEFAULT NULL COMMENT '任务对象类型1数据源、2数据表',
    "f_name" varchar(255 char) DEFAULT NULL COMMENT '任务名称',
    "f_status" tinyint NOT NULL COMMENT '任务状态：0成功，1失败，2进行中',
    "f_start_time" datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务开始时间',
    "f_end_time" datetime DEFAULT NULL COMMENT '任务结束时间',
    "f_create_user" varchar(100 char) NOT NULL DEFAULT '' COMMENT '创建用户',
    "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    "f_advanced_params" varchar(255 char) NOT NULL DEFAULT '[]' COMMENT '任务高级参数',
    CLUSTER PRIMARY KEY ("f_id")
    );

INSERT INTO kweaver."t_task" (
    "f_id",
    "f_object_id",
    "f_object_type",
    "f_name",
    "f_status",
    "f_start_time",
    "f_end_time",
    "f_create_user",
    "f_authority_id",
    "f_advanced_params"
)
SELECT
    "f_id",
    "f_object_id",
    "f_object_type",
    "f_name",
    "f_status",
    "f_start_time",
    "f_end_time",
    "f_create_user",
    "f_authority_id",
    "f_advanced_params"
FROM adp."t_task" s;

-- ==========================================
-- t_task_log
-- ==========================================
CREATE TABLE IF NOT EXISTS "t_task_log" (
                                            "f_id" bigint NOT NULL COMMENT '唯一id，雪花算法',
                                            "f_task_id" bigint DEFAULT NULL COMMENT '任务id',
                                            "f_log" text DEFAULT NULL COMMENT '任务日志文本',
                                            "f_authority_id" bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
                                            CLUSTER PRIMARY KEY ("f_id")
    );

INSERT INTO kweaver."t_task_log" (
    "f_id",
    "f_task_id",
    "f_log",
    "f_authority_id"
)
SELECT
    "f_id",
    "f_task_id",
    "f_log",
    "f_authority_id"
FROM adp."t_task_log" s;
