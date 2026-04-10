-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

USE kweaver;

CREATE TABLE IF NOT EXISTS `data_source` (
    `id` char(36) NOT NULL COMMENT '主键，生成规则:36位uuid',
    `name` varchar(128) NOT NULL COMMENT '数据源展示名称',
    `type_name` varchar(30) NOT NULL COMMENT '数据库类型',
    `bin_data` blob NOT NULL COMMENT '数据源配置信息',
    `comment` varchar(255) DEFAULT NULL COMMENT '描述',
    `created_by_uid` char(36) NOT NULL COMMENT '创建人',
    `created_at` datetime(3) NOT NULL COMMENT '创建时间',
    `updated_by_uid` char(36) DEFAULT NULL COMMENT '修改人',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`)
    );

CREATE TABLE IF NOT EXISTS `t_data_source_info` (
      `f_id` varchar(36) NOT NULL COMMENT '主键，生成规则:36位uuid',
      `f_name` varchar(128) NOT NULL COMMENT '数据源展示名称',
      `f_type` varchar(30) NOT NULL COMMENT '数据库类型',
      `f_catalog` varchar(50) COMMENT '数据源catalog名称',
      `f_database` varchar(100) COMMENT '数据库名称',
      `f_schema` varchar(100) COMMENT '数据库模式',
      `f_connect_protocol` varchar(30) NOT NULL COMMENT '连接方式',
      `f_host` varchar(128) NOT NULL COMMENT '地址',
      `f_port` int NOT NULL COMMENT '端口',
      `f_account` varchar(128) COMMENT '账户',
      `f_password` varchar(1024) COMMENT '密码',
      `f_storage_protocol` varchar(30) COMMENT '存储介质',
      `f_storage_base` varchar(1024) COMMENT '存储路径',
      `f_token` varchar(100) COMMENT 'token认证',
      `f_replica_set` varchar(100) COMMENT '副本集名称',
      `f_is_built_in` tinyint NOT NULL DEFAULT '0' COMMENT '是否为内置数据源（0 特殊 1 非内置 2 内置），默认为0',
      `f_comment` varchar(255) COMMENT '描述',
      `f_created_by_uid` varchar(36) COMMENT '创建人',
      `f_created_at` datetime(3) COMMENT '创建时间',
      `f_updated_by_uid` varchar(36) COMMENT '修改人',
      `f_updated_at` datetime(3) COMMENT '更新时间',
      PRIMARY KEY (`f_id`)
);

INSERT INTO `t_data_source_info` (`f_id`,`f_name`,`f_type`,`f_connect_protocol`,`f_host`,`f_port`,`f_is_built_in`,`f_created_at`,`f_updated_at`)
SELECT 'cedb5294-07c3-45b1-a273-17baefa62800','索引库','index_base','http','mdl-index-base-svc',13013,2,current_timestamp(),current_timestamp()
FROM DUAL WHERE NOT EXISTS( SELECT `f_id` FROM `t_data_source_info` WHERE `f_id` = 'cedb5294-07c3-45b1-a273-17baefa62800' AND `f_type` = 'index_base');

CREATE TABLE IF NOT EXISTS `t_task_scan` (
  `id` char(36) NOT NULL COMMENT '唯一id，雪花算法',
  `type` tinyint NOT NULL DEFAULT 0 COMMENT '扫描任务：0 :即时-数据源;1 :即时-数据表;2: 定时-数据源;3: 定时-数据表',
  `name` varchar(128) NOT NULL COMMENT '任务名称',
  `ds_id` char(36) DEFAULT NULL COMMENT '数据源唯一标识',
  `scan_status` tinyint DEFAULT NULL COMMENT '任务状态',
  `start_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务开始时间',
  `end_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务结束时间',
  `create_user` varchar(100) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  `task_params_info` text DEFAULT NULL COMMENT '任务执行参数信息',
  `task_process_info` text DEFAULT NULL COMMENT '任务执行进度信息',
  `task_result_info` text DEFAULT NULL COMMENT '任务执行结果信息',
  `schedule_id` char(36) DEFAULT NULL COMMENT '定时任务配置id',
  PRIMARY KEY (`id`),
  KEY `t_task_scan_ds_id_IDX` (`ds_id`)
);

CREATE TABLE IF NOT EXISTS `t_task_scan_table` (
  `id` char(36) NOT NULL COMMENT '唯一id，雪花算法',
  `task_id` char(36) NOT NULL COMMENT '关联任务id',
  `ds_id` char(36) NOT NULL COMMENT '数据源唯一标识',
  `ds_name` varchar(128) NOT NULL COMMENT '数据源名称',
  `table_id` char(36) NOT NULL COMMENT 'table的唯一id',
  `table_name` varchar(128) NOT NULL COMMENT 'table的name',
  `schema_name` varchar(128) NOT NULL COMMENT 'schema的name',
  `scan_status` tinyint DEFAULT NULL COMMENT '任务状态：0 等待;1 进行中;2 成功;3 失败',
  `start_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务开始时间',
  `end_time` datetime  DEFAULT NULL COMMENT '任务结束时间',
  `create_user` varchar(100) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  `scan_params` text DEFAULT NULL COMMENT '任务执行参数信息',
  `scan_result_info` text DEFAULT NULL COMMENT '任务执行结果：',
  `error_stack` text DEFAULT NULL COMMENT '异常堆栈信息',
  `operation_type` tinyint DEFAULT NULL COMMENT '操作类型:0:insert 1:delete 2:update 3:unknown',
  PRIMARY KEY (`id`),
  KEY `t_task_scan_table_task_id_IDX` (`task_id`),
  KEY `t_task_scan_table_table_id_IDX` (`table_id`)
);

CREATE TABLE IF NOT EXISTS `t_table_scan` (
  `f_id` char(36) NOT NULL COMMENT '唯一id，雪花算法',
  `f_name` varchar(128) NOT NULL COMMENT '表名称',
  `f_advanced_params` text DEFAULT NULL COMMENT '高级参数，格式为"{key(1): value(1), ... , key(n): value(n)}"',
  `f_description` varchar(2048) DEFAULT NULL COMMENT '表注释',
  `f_table_rows` bigint NOT NULL DEFAULT 0 COMMENT '表数据量，默认0',
  `f_data_source_id` char(36) NOT NULL COMMENT '数据源唯一标识',
  `f_data_source_name` varchar(128) NOT NULL COMMENT '冗余字段，数据源名称',
  `f_schema_name` varchar(128) NOT NULL COMMENT '冗余字段，schema名称',
  `f_task_id` char(36) NOT NULL COMMENT '关联任务id',
  `f_version` int NOT NULL DEFAULT 1 COMMENT '版本号',
  `f_create_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  `f_create_user` varchar(100) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  `f_operation_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
  `f_operation_user` varchar(100) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
  `f_operation_type` tinyint NOT NULL DEFAULT 0 COMMENT '状态：0新增1删除2更新',
  `f_status` tinyint NOT NULL DEFAULT 3 COMMENT '任务状态：0 等待;1 进行中;2 成功;3 失败',
  `f_status_change` tinyint NOT NULL DEFAULT 0 COMMENT '状态是否发生变化：0 否1 是',
  `f_scan_source` tinyint(4) DEFAULT NULL COMMENT '扫描来源',
  PRIMARY KEY (`f_id`),
  KEY `t_table_scan_f_data_source_id_IDX` (`f_task_id`)
);

CREATE TABLE IF NOT EXISTS `t_table_field_scan` (
  `f_id` char(36) NOT NULL COMMENT '唯一id，雪花算法',
  `f_field_name` varchar(128) NOT NULL COMMENT '字段名',
  `f_table_id` char(36) NOT NULL COMMENT 'Table唯一标识',
  `f_table_name` varchar(128) NOT NULL COMMENT '表名',
  `f_field_type` varchar(128) DEFAULT NULL COMMENT '字段类型',
  `f_field_length` int DEFAULT NULL COMMENT '字段长度',
  `f_field_precision` int DEFAULT NULL COMMENT '字段精度',
  `f_field_comment` varchar(2048) DEFAULT NULL COMMENT '字段注释',
  `f_field_order_no` int DEFAULT NULL,
  `f_advanced_params` varchar(2048) DEFAULT NULL COMMENT '字段高级参数',
  `f_version` int NOT NULL DEFAULT 1 COMMENT '版本号',
  `f_create_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  `f_create_user` varchar(100) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
  `f_operation_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
  `f_operation_user` varchar(100) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
  `f_operation_type` tinyint NOT NULL DEFAULT 0 COMMENT '状态：0新增1删除2更新',
  `f_status_change` tinyint NOT NULL DEFAULT 0 COMMENT '状态是否发生变化：0 否1 是',
  PRIMARY KEY (`f_id`),
  KEY `t_table_field_scan_f_table_id_IDX` (`f_table_id`)
);

CREATE TABLE IF NOT EXISTS `t_task_scan_schedule` (
  `id` char(36) NOT NULL COMMENT '唯一id，雪花算法',
  `type` tinyint NOT NULL DEFAULT 0 COMMENT '扫描任务：0 :即时-数据源;1 :即时-数据表;2: 定时-数据源',
  `name` varchar(128) NOT NULL COMMENT '任务名称',
  `cron_expression` varchar(64) NOT NULL COMMENT 'cron表达式',
  `scan_strategy` varchar(64) DEFAULT NULL COMMENT '快速扫描策略',
  `task_status` tinyint DEFAULT 0 COMMENT '定时扫描任务:0 close 1 open',
  `ds_id` char(36) DEFAULT NULL COMMENT '数据源唯一标识',
  `create_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  `create_user` varchar(64) DEFAULT NULL COMMENT '创建用户',
  `operation_time` datetime DEFAULT NULL,
  `operation_user` varchar(64) DEFAULT NULL,
  `operation_type` tinyint NOT NULL DEFAULT 0 COMMENT '状态：0新增1删除2更新',
PRIMARY KEY (`id`),
KEY `t_task_scan_ds_id_IDX` (`ds_id`)
);


-- metadata tables


CREATE TABLE IF NOT EXISTS `t_data_quality_model` (
                                                      `f_id` bigint NOT NULL COMMENT '主键id，唯一标识',
                                                      `f_ds_id` bigint NOT NULL COMMENT '数据源ID',
                                                      `f_dolphinscheduler_ds_id` bigint NOT NULL COMMENT 'dolphinscheduler数据源ID',
                                                      `f_db_type` varchar(50) NOT NULL COMMENT '数据库类型',
    `f_tb_name` varchar(512) NOT NULL COMMENT '表名称',
    `f_process_definition_code` bigint NOT NULL COMMENT '工作流定义ID',
    `f_crontab` varchar(128) DEFAULT NULL COMMENT '定时任务表达式',
    PRIMARY KEY (`f_id`)
    );


CREATE TABLE IF NOT EXISTS `t_data_quality_rule` (
                                                     `f_id` bigint NOT NULL COMMENT '主键id，唯一标识',
                                                     `f_field_name` varchar(512) NOT NULL COMMENT '字段名称',
    `f_rule_id` tinyint NOT NULL COMMENT '质量规则ID：1-空值检测，，2-自定义SQL，5-字段长度校验，6-唯一性校验，7-正则表达式，9-枚举值校验，10-表行数校验',
    `f_threshold` double DEFAULT NULL COMMENT '阈值，默认0',
    `f_check_val` varchar(10240) DEFAULT NULL COMMENT '1、自定义sql：填写sql语句；2、字段长度校验：填写字段长度；3、正则表达式：填写正则表达式；4、枚举值校验：填写枚举值，逗号分割；5、表行数校验：填写表行数。',
    `f_check_val_name` varchar(128) DEFAULT NULL COMMENT '自定义sql时，填写的实际值名',
    `f_model_id` bigint NOT NULL COMMENT '质量模型ID',
    PRIMARY KEY (`f_id`)
    );


CREATE TABLE IF NOT EXISTS `t_data_source` (
    `f_id` char(36) NOT NULL COMMENT '唯一id，雪花算法',
    `f_name` varchar(128) NOT NULL COMMENT '数据源名称',
    `f_data_source_type` tinyint NOT NULL COMMENT '类型，关联字典表f_dict_type为1时的f_dict_key',
    `f_data_source_type_name` varchar(256) NOT NULL COMMENT '类型名称，对应字典表f_dict_type为1时的f_dict_value',
    `f_user_name` varchar(128) NOT NULL COMMENT '用户名',
    `f_password` varchar(1024) NOT NULL COMMENT '密码',
    `f_description` varchar(255) NOT NULL DEFAULT '' COMMENT '描述',
    `f_extend_property` varchar(255) NOT NULL DEFAULT '' COMMENT '扩展属性，默认为空字符串',
    `f_host` varchar(128) NOT NULL COMMENT 'HOST',
    `f_port` int NOT NULL COMMENT '端口',
    `f_enable_status` tinyint NOT NULL DEFAULT 1 COMMENT '禁用/启用状态，1 启用，2 停用，默认为启用',
    `f_connect_status` tinyint NOT NULL DEFAULT 1 COMMENT '连接状态，1 成功，2 失败，默认为成功',
    `f_authority_id` bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    `f_create_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `f_create_user` varchar(100) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
    `f_update_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
    `f_update_user` varchar(100) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
    `f_database` varchar(100) DEFAULT NULL COMMENT '数据库名称',
    `f_info_system_id` varchar(128) DEFAULT NULL COMMENT '信息系统id',
    `f_dolphin_id` bigint DEFAULT NULL COMMENT 'dolphin数据元id',
    `f_delete_code` bigint DEFAULT 0 COMMENT '逻辑删除标识码',
    `f_live_update_status` tinyint NOT NULL DEFAULT 0 COMMENT '实时更新标识（0无需更新，1待更新，2更新中，3连接不可用，4无权限，5待广播',
    `f_live_update_benchmark` varchar(255) DEFAULT NULL COMMENT '实时更新基准',
    `f_live_update_time` datetime DEFAULT current_timestamp() COMMENT '实时更新时间',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `t_data_source_un` (`f_name`,`f_create_user`,`f_info_system_id`,`f_delete_code`)
    );


CREATE TABLE IF NOT EXISTS `t_dict` (
                                        `f_id` int NOT NULL AUTO_INCREMENT COMMENT '唯一id，自增ID',
                                        `f_dict_type` tinyint NOT NULL COMMENT '字典类型\n1：数据源类型\n2：Oracle字段类型\n3：MySQL字段类型\n4：PostgreSQL字段类型\n5：SqlServer字段类型\n6：Hive字段类型\n7：HBase字段类型\n8：MongoDB字段类型\n9：FTP字段类型\n10：HDFS字段类型\n11：SFTP字段类型\n12：CMQ字段类型\n13：Kafka字段类型\n14：API字段类型',
                                        `f_dict_key` tinyint NOT NULL COMMENT '枚举值',
                                        `f_dict_value` varchar(256) NOT NULL COMMENT '枚举对应描述',
    `f_extend_property` varchar(1024) NOT NULL COMMENT '扩展属性',
    `f_enable_status` tinyint NOT NULL DEFAULT 2 COMMENT '启用状态，1 启用，2 停用，默认为停用',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `t_dict_un` (`f_dict_type`,`f_dict_key`)
    );


CREATE TABLE IF NOT EXISTS `t_indicator` (
                                             `f_id` bigint NOT NULL COMMENT '唯一id，雪花算法',
                                             `f_indicator_name` varchar(128) NOT NULL COMMENT '指标名称',
    `f_indicator_type` varchar(128) NOT NULL COMMENT '指标类型',
    `f_indicator_value` bigint NOT NULL COMMENT '指标数值',
    `f_create_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `f_indicator_object_id` bigint DEFAULT NULL COMMENT '关联对象ID',
    `f_authority_id` bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    `f_advanced_params` varchar(255) NOT NULL DEFAULT '[]' COMMENT '指标高级参数',
    PRIMARY KEY (`f_id`,`f_create_time`)
    );


CREATE TABLE IF NOT EXISTS `t_lineage_edge_column` (
    `f_id` varchar(64) NOT NULL COMMENT '主键ID，根据f_table_id和f_column_id值MD5计算得到',
    `f_parent_id` varchar(64) NOT NULL COMMENT '源字段ID',
    `f_child_id` varchar(64) NOT NULL COMMENT '目标字段ID',
    `f_create_type` varchar(20) DEFAULT NULL COMMENT '创建类型： HIVE/DATAX/SPARK/USER_REPORT',
    `f_query_text` text DEFAULT NULL COMMENT '生成血缘的sql或者脚本说明',
    `created_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    `updated_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    `deleted_at` bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    `f_create_time` timestamp NULL DEFAULT NULL COMMENT '创建时间，时间戳',
    PRIMARY KEY (`f_id`)
    );


CREATE TABLE IF NOT EXISTS `t_lineage_edge_column_table_relation` (
    `f_id` varchar(64) NOT NULL COMMENT '主键ID，根据f_table_id和f_column_id值MD5计算得到',
    `f_table_id` varchar(64) NOT NULL COMMENT '表ID',
    `f_column_id` varchar(64) NOT NULL COMMENT '字段ID',
    `created_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    `updated_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    `deleted_at` bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    PRIMARY KEY (`f_id`)
    );


CREATE TABLE IF NOT EXISTS `t_lineage_edge_table` (
    `f_id` varchar(64) NOT NULL COMMENT '主键ID，根据f_table_id和f_column_id值MD5计算得到',
    `f_parent_id` varchar(64) NOT NULL COMMENT '源ID',
    `f_child_id` varchar(64) NOT NULL COMMENT '目标ID',
    `f_create_type` varchar(20) DEFAULT NULL COMMENT '创建类型： HIVE/DATAX/SPARK/USER_REPORT',
    `f_query_text` text DEFAULT NULL COMMENT '生成血缘的sql或者脚本说明',
    `created_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    `updated_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    `deleted_at` bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    `f_create_time` timestamp NULL DEFAULT NULL COMMENT '创建时间，时间戳',
    PRIMARY KEY (`f_id`)
    );


CREATE TABLE IF NOT EXISTS `t_lineage_graph_info` (
    `app_id` char(20) NOT NULL COMMENT '图谱appId',
    `graph_id` bigint DEFAULT NULL COMMENT '图谱graphId',
    PRIMARY KEY (`app_id`)
    );


CREATE TABLE IF NOT EXISTS `t_lineage_log` (
    `id` char(36) NOT NULL DEFAULT (uuid()),
    `class_id` char(36) NOT NULL COMMENT '实体的主键id',
    `class_type` char(36) NOT NULL COMMENT '实体类型',
    `action_type` char(10) NOT NULL COMMENT '操作类型：insert update delete',
    `class_data` text NOT NULL COMMENT '血缘实体json',
    `created_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    PRIMARY KEY (`id`)
    );


CREATE TABLE IF NOT EXISTS `t_lineage_relation` (
    `unique_id` varchar(255) NOT NULL COMMENT '实体ID',
    `class_type` tinyint DEFAULT NULL COMMENT '类型，1:column,2:indicator',
    `parent` text DEFAULT NULL COMMENT '上一个节点',
    `child` text DEFAULT NULL COMMENT '下一个节点',
    `created_at` datetime(3) DEFAULT current_timestamp(3) COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT current_timestamp(3) COMMENT '更新时间',
    PRIMARY KEY (`unique_id`)
    );


CREATE TABLE IF NOT EXISTS `t_lineage_tag_column` (
    `f_id` varchar(64) NOT NULL COMMENT '主键ID，根据f_table_id和f_column值MD5计算得到',
    `f_table_id` varchar(64) NOT NULL COMMENT 't_lineage_tag_table表ID',
    `f_column` varchar(255) NOT NULL COMMENT '字段名称',
    `created_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    `updated_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    `deleted_at` bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    PRIMARY KEY (`f_id`)
    );


CREATE TABLE IF NOT EXISTS `t_lineage_tag_table` (
    `f_id` varchar(64) NOT NULL COMMENT '主键ID，根据f_db_type、f_ds_id、f_jdbc_url、f_jdbc_user、f_db_name、f_db_schema、f_tb_name值MD5计算得到',
    `f_db_type` varchar(64) NOT NULL COMMENT '数据库类型',
    `f_ds_id` varchar(64) DEFAULT NULL COMMENT '数据源ID',
    `f_jdbc_url` varchar(255) DEFAULT NULL COMMENT '数据库连接URL',
    `f_jdbc_user` varchar(255) DEFAULT NULL COMMENT '数据库JDBC 用户名',
    `f_db_name` varchar(255) DEFAULT NULL COMMENT '数据库名称',
    `f_db_schema` varchar(255) DEFAULT NULL COMMENT '模式名称',
    `f_tb_name` varchar(255) NOT NULL COMMENT '表名称',
    `created_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    `updated_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    `deleted_at` bigint DEFAULT 0 COMMENT '删除时间（逻辑删除）',
    PRIMARY KEY (`f_id`)
    );


CREATE TABLE IF NOT EXISTS `t_indicator2` (
                                              `f_id` bigint NOT NULL COMMENT '唯一id，雪花算法',
                                              `f_indicator_name` varchar(128) NOT NULL COMMENT '指标名称',
    `f_indicator_type` varchar(128) NOT NULL COMMENT '指标类型',
    `f_indicator_value` bigint NOT NULL COMMENT '指标数值',
    `f_create_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `f_indicator_object_id` bigint DEFAULT NULL COMMENT '关联对象ID',
    `f_authority_id` bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    `f_advanced_params` varchar(255) NOT NULL DEFAULT '[]' COMMENT '指标高级参数',
    PRIMARY KEY (`f_id`,`f_create_time`)
    );


CREATE TABLE IF NOT EXISTS `t_lineage_tag_column2` (
    `unique_id` varchar(255) NOT NULL COMMENT '列的唯一id',
    `uuid` char(36) DEFAULT NULL COMMENT '字段的uuid',
    `technical_name` varchar(255) DEFAULT NULL COMMENT '列技术名称',
    `business_name` varchar(255) DEFAULT NULL COMMENT '列业务名称',
    `comment` varchar(300) DEFAULT NULL COMMENT '字段注释',
    `data_type` varchar(255) DEFAULT NULL COMMENT '字段的数据类型',
    `primary_key` tinyint DEFAULT NULL COMMENT '是否主键',
    `table_unique_id` char(36) DEFAULT NULL COMMENT '属于血缘表的uuid',
    `expression_name` text DEFAULT NULL COMMENT 'column的生成表达式',
    `column_unique_ids` varchar(1024) DEFAULT '' COMMENT 'column的生成依赖的column的uid',
    `action_type` varchar(10) DEFAULT NULL COMMENT '操作类型:insertupdatedelete',
    `created_at` datetime(3) DEFAULT current_timestamp(3) COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT current_timestamp(3) COMMENT '更新时间',
    PRIMARY KEY (`unique_id`)
    );


CREATE TABLE IF NOT EXISTS `t_lineage_tag_indicator2` (
    `uuid` char(36) NOT NULL COMMENT '指标的uuid',
    `name` varchar(128) NOT NULL COMMENT '指标名称',
    `description` varchar(300) DEFAULT NULL COMMENT '指标名称描述',
    `code` varchar(128) NOT NULL COMMENT '指标编号',
    `indicator_type` varchar(10) NOT NULL COMMENT '指标类型:atomic原子derived衍生composite复合',
    `expression` text DEFAULT NULL COMMENT '指标表达式，如果指标是原子或复合指标时',
    `indicator_uuids` varchar(1024) DEFAULT '' COMMENT '引用的指标uuid',
    `time_restrict` text DEFAULT NULL COMMENT '时间限定表达式，如果指标是衍生指标时',
    `modifier_restrict` text DEFAULT NULL COMMENT '普通限定表达式，如果指标是衍生指标时',
    `owner_uid` varchar(50) DEFAULT NULL COMMENT '数据ownerID',
    `owner_name` varchar(128) DEFAULT NULL COMMENT '数据owner名称',
    `department_id` char(36) DEFAULT NULL COMMENT '所属部门id',
    `department_name` varchar(128) DEFAULT NULL COMMENT '所属部门名称',
    `column_unique_ids` varchar(1024) DEFAULT '' COMMENT '依赖的字段的unique_id',
    `action_type` varchar(10) NOT NULL COMMENT '操作类型:insertupdatedelete',
    `created_at` datetime(3) DEFAULT current_timestamp(3) COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT current_timestamp(3) COMMENT '更新时间',
    PRIMARY KEY (`uuid`)
    );


CREATE TABLE IF NOT EXISTS `t_lineage_tag_table2` (
    `unique_id` varchar(255) NOT NULL COMMENT '唯一id',
    `uuid` char(36) NOT NULL COMMENT '表的uuid',
    `technical_name` varchar(255) NOT NULL COMMENT '表技术名称',
    `business_name` varchar(255) DEFAULT NULL COMMENT '表业务名称',
    `comment` varchar(300) DEFAULT NULL COMMENT '表注释',
    `table_type` varchar(36) NOT NULL COMMENT '表类型',
    `datasource_id` char(36) DEFAULT NULL COMMENT '数据源id',
    `datasource_name` varchar(255) DEFAULT NULL COMMENT '数据源名称',
    `owner_id` char(36) DEFAULT NULL COMMENT '数据Ownerid',
    `owner_name` varchar(128) DEFAULT NULL COMMENT '数据OwnerName',
    `department_id` char(36) DEFAULT NULL COMMENT '所属部门id',
    `department_name` varchar(128) DEFAULT NULL COMMENT '所属部门mame',
    `info_system_id` char(36) DEFAULT NULL COMMENT '信息系统id',
    `info_system_name` varchar(128) DEFAULT NULL COMMENT '信息系统名称',
    `database_name` varchar(128) NOT NULL COMMENT '数据库名称',
    `catalog_name` varchar(255) NOT NULL DEFAULT '' COMMENT '数据源catalog名称',
    `catalog_addr` varchar(1024) NOT NULL DEFAULT '' COMMENT '数据源地址',
    `catalog_type` varchar(128) NOT NULL COMMENT '数据库类型名称',
    `task_execution_info` varchar(128) DEFAULT NULL COMMENT '表加工任务的相关名称',
    `action_type` varchar(10) NOT NULL COMMENT '操作类型:insertupdatedelete',
    `created_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '创建时间',
    `updated_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '更新时间',
    PRIMARY KEY (`unique_id`)
    );


CREATE TABLE IF NOT EXISTS `t_live_ddl` (
                                            `f_id` bigint NOT NULL AUTO_INCREMENT COMMENT '唯一标识',
                                            `f_data_source_id` bigint NOT NULL DEFAULT 0 COMMENT '数据源ID',
                                            `f_data_source_name` varchar(255) NOT NULL DEFAULT '' COMMENT '数据源名称',
    `f_origin_catalog` varchar(255) DEFAULT NULL COMMENT '物理catalog',
    `f_virtual_catalog` varchar(255) DEFAULT NULL COMMENT '虚拟化catalog',
    `f_schema_id` bigint DEFAULT NULL COMMENT 'schemaID',
    `f_schema_name` varchar(255) DEFAULT NULL COMMENT 'schema名称',
    `f_table_id` bigint DEFAULT NULL COMMENT 'tableID',
    `f_table_name` varchar(255) DEFAULT NULL COMMENT 'table名称',
    `f_sql_type` varchar(100) DEFAULT NULL COMMENT 'sql类型(AlterTable,AlterColumn,CreateTable,CommentTable,CommentColumn,DropTable,RenameTable)',
    `f_sql_text` text NOT NULL COMMENT 'sql文本',
    `f_live_update_benchmark` varchar(255) NOT NULL DEFAULT '' COMMENT '实时更新基准',
    `f_monitor_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '监听时间，默认当前时间',
    `f_update_status` tinyint DEFAULT NULL COMMENT '更新状态（0全量更新，1增量更新，2忽略更新，3待更新，4解析失败，5更新失败）',
    `f_update_message` varchar(2000) DEFAULT NULL COMMENT '更新信息',
    `f_push_status` tinyint DEFAULT NULL COMMENT '0不推送,1待推送,2已推送',
    PRIMARY KEY (`f_id`)
    );


CREATE TABLE IF NOT EXISTS `t_schema` (
                                          `f_id` bigint NOT NULL COMMENT '唯一id，雪花算法',
                                          `f_name` varchar(128) NOT NULL COMMENT 'schema名称',
    `f_data_source_id` char(36) NOT NULL COMMENT '数据源唯一标识',
    `f_data_source_name` varchar(128) NOT NULL COMMENT '冗余字段，数据源名称',
    `f_data_source_type` tinyint NOT NULL COMMENT '冗余字段，数据源类型，关联字典表f_dict_type为1时的f_dict_key',
    `f_data_source_type_name` varchar(256) NOT NULL COMMENT '冗余字段，数据源类型名称，对应字典表f_dict_type为1时的f_dict_value',
    `f_authority_id` bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    `f_create_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `f_create_user` varchar(100) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
    `f_update_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
    `f_update_user` varchar(100) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `t_schema_un` (`f_data_source_id`,`f_name`)
    );


CREATE TABLE IF NOT EXISTS `t_table` (
                                         `f_id` bigint NOT NULL COMMENT '唯一id，雪花算法',
                                         `f_name` varchar(128) NOT NULL COMMENT '表名称',
    `f_advanced_params` text NOT NULL COMMENT '高级参数，默认为"{}"，格式为"{key(1): value(1), ... , key(n): value(n)}"',
    `f_description` varchar(2048) DEFAULT NULL COMMENT '表注释',
    `f_table_rows` bigint NOT NULL DEFAULT 0 COMMENT '表数据量，默认0',
    `f_schema_id` bigint NOT NULL COMMENT 'schema唯一标识',
    `f_schema_name` varchar(128) NOT NULL COMMENT '冗余字段，schema名称',
    `f_data_source_id` char(36) NOT NULL COMMENT '数据源唯一标识',
    `f_data_source_name` varchar(128) NOT NULL COMMENT '冗余字段，数据源名称',
    `f_data_source_type` tinyint NOT NULL COMMENT '冗余字段，数据源类型，关联字典表f_dict_type为1时的f_dict_key',
    `f_data_source_type_name` varchar(256) NOT NULL COMMENT '冗余字段，数据源类型名称，对应字典表f_dict_type为1时的f_dict_value',
    `f_version` int NOT NULL DEFAULT 1 COMMENT '版本号',
    `f_authority_id` varchar(100) NOT NULL DEFAULT '' COMMENT '权限域（目前为预留字段），默认0',
    `f_create_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
    `f_create_user` varchar(100) NOT NULL DEFAULT '' COMMENT '创建用户（ID），默认空字符串',
    `f_update_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '修改时间，默认当前时间',
    `f_update_user` varchar(100) NOT NULL DEFAULT '' COMMENT '修改用户（ID），默认空字符串',
    `f_delete_flag` tinyint NOT NULL DEFAULT 0 COMMENT '逻辑删除标识',
    `f_delete_time` datetime DEFAULT NULL COMMENT '逻辑删除时间',
    `f_scan_source` tinyint DEFAULT NULL COMMENT '扫描来源',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `t_table_un` (`f_data_source_id`,`f_schema_id`,`f_name`)
    );


CREATE TABLE IF NOT EXISTS `t_table_field` (
                                               `f_table_id` bigint NOT NULL COMMENT 'Table唯一标识',
                                               `f_field_name` varchar(128) NOT NULL COMMENT '字段名',
    `f_field_type` varchar(128) DEFAULT NULL COMMENT '字段类型',
    `f_field_length` int DEFAULT NULL COMMENT '字段长度',
    `f_field_precision` int DEFAULT NULL COMMENT '字段精度',
    `f_field_comment` varchar(2048) DEFAULT NULL COMMENT '字段注释',
    `f_advanced_params` varchar(2048) NOT NULL DEFAULT '[]' COMMENT '字段高级参数',
    `f_update_flag` tinyint NOT NULL DEFAULT 0 COMMENT '更新标识',
    `f_update_time` datetime DEFAULT NULL COMMENT '更新时间',
    `f_delete_flag` tinyint NOT NULL DEFAULT 0 COMMENT '逻辑删除标识',
    `f_delete_time` datetime DEFAULT NULL COMMENT '逻辑删除时间',
    PRIMARY KEY (`f_table_id`,`f_field_name`)
    );


CREATE TABLE IF NOT EXISTS `t_table_field_his` (
                                                   `f_id` bigint NOT NULL COMMENT '唯一id，雪花算法',
                                                   `f_field_name` varchar(128) NOT NULL COMMENT '字段名',
    `f_field_type` varchar(128) DEFAULT NULL COMMENT '字段类型',
    `f_field_length` int DEFAULT NULL COMMENT '字段长度',
    `f_field_precision` int DEFAULT NULL COMMENT '字段精度',
    `f_field_comment` varchar(2048) DEFAULT NULL COMMENT '字段注释',
    `f_table_id` bigint NOT NULL COMMENT 'Table唯一标识',
    `f_version` int NOT NULL DEFAULT 1 COMMENT '版本号',
    `f_advanced_params` varchar(255) NOT NULL DEFAULT '[]' COMMENT '字段高级参数',
    PRIMARY KEY (`f_id`,`f_version`)
    );


CREATE TABLE IF NOT EXISTS `t_task` (
                                        `f_id` bigint NOT NULL COMMENT '唯一id，雪花算法',
                                        `f_object_id` char(36) DEFAULT NULL COMMENT '任务对象id',
    `f_object_type` tinyint DEFAULT NULL COMMENT '任务对象类型1数据源、2数据表',
    `f_name` varchar(255) DEFAULT NULL COMMENT '任务名称',
    `f_status` tinyint NOT NULL COMMENT '任务状态：0成功，1失败，2进行中',
    `f_start_time` datetime NOT NULL DEFAULT current_timestamp() COMMENT '任务开始时间',
    `f_end_time` datetime DEFAULT NULL COMMENT '任务结束时间',
    `f_create_user` varchar(100) NOT NULL DEFAULT '' COMMENT '创建用户',
    `f_authority_id` bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
    `f_advanced_params` varchar(255) NOT NULL DEFAULT '[]' COMMENT '任务高级参数',
    PRIMARY KEY (`f_id`)
    );


CREATE TABLE IF NOT EXISTS `t_task_log` (
                                            `f_id` bigint NOT NULL COMMENT '唯一id，雪花算法',
                                            `f_task_id` bigint DEFAULT NULL COMMENT '任务id',
                                            `f_log` text DEFAULT NULL COMMENT '任务日志文本',
                                            `f_authority_id` bigint NOT NULL DEFAULT 0 COMMENT '权限域（目前为预留字段），默认0',
                                            PRIMARY KEY (`f_id`)
    );




-- 添加虚拟化数据源类型
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,1,'Oracle','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,2,'MySQL','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,3,'PostgreSQL','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,4,'SqlServer','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,5,'Hive','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,6,'HBase','{"dbCatalogName": 非关系型数据库}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,7,'MongoDB','{"dbCatalogName": 非关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,9,'HDFS','{"dbCatalogName": 文件系统}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,10,'SFTP','{"dbCatalogName": 文件系统}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,11,'CMQ','{"dbCatalogName": 消息队列}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,12,'Kafka','{"dbCatalogName": 消息队列}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,13,'API','{"dbCatalogName": 其他}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,14,'CLICKHOUSE', '{"dbCatalogName": 非关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 14);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,15,'doris','{"dbCatalogName": 非关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 15);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,16,'mariadb','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 16);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,17,'dm','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 17);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,18,'maxcompute','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 18);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,1,'VARCHAR2','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,2,'NUMBER','{"jdbcType":2,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,3,'NVARCHAR2','{"jdbcType":-9,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,4,'DATE','{"jdbcType":91,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,5,'NCLOB','{"jdbcType":2011,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,6,'TIMESTAMP','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,7,'CLOB','{"jdbcType":2005,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,8,'TIMESTAMP(6)','{"jdbcType":10,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,9,'CHAR','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,10,'VARCHAR','{"jdbcType":11,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,11,'FLOAT','{"jdbcType":6,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,12,'BLOB','{"jdbcType":2004,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,13,'DECIMAL','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,14,'LONG','{"jdbcType":-1,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 14);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,15,'INT','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 15);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,16,'RAW','{"jdbcType":-3,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 16);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,17,'TIMESTAMP(4)','{"jdbcType":3,"javaColumnType":"8","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 17);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,18,'ROWID','{"jdbcType":-8,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 18);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,19,'AQ$_SUBSCRIBERS','{"jdbcType":2003,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 19);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,20,'LONG RAW','{"jdbcType":-4,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 20);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,21,'NCHAR','{"jdbcType":-15,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 21);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,1,'DOUBLE','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,2,'TINYINT','{"jdbcType":-6,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,3,'BOOLEAN','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,4,'INTEGER','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,5,'VARCHAR','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,6,'CHAR','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,7,'BLOB','{"jdbcType":-4,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,8,'SMALLINT','{"jdbcType":5,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,9,'MEDIUMINT','{"jdbcType":4,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,10,'BIT','{"jdbcType":-7,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,11,'FLOAT','{"jdbcType":7,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,12,'DECIMAL','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,13,'DATE','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,14,'TIME','{"jdbcType":92,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 14);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,15,'DATETIME','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 15);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,16,'TIMESTAMP','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 16);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,17,'YEAR','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 17);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,18,'INT','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 18);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,19,'BIGINT','{"jdbcType":-5,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 19);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,20,'LONGTEXT','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 20);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,21,'TEXT','{"jdbcType":-1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 21);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,22,'JSON','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 22);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,23,'MEDIUMTEXT','{"jdbcType":13,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 23);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,24,'SET','{"jdbcType":11,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 24);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,25,'MEDIUMBLOB','{"jdbcType":11,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 25);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,26,'LONGBLOB','{"jdbcType":11,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 26);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,27,'VARBINARY','{"jdbcType":-3,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 27);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,28,'BINARY','{"jdbcType":-2,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 28);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,29,'ENUM','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 29);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,30,'TINYTEXT','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 30);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,1,'int2','{"jdbcType":12,"javaColumnType":"3","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,2,'int4','{"jdbcType":12,"javaColumnType":"3","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,3,'int8','{"jdbcType":-5,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,4,'varchar','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,5,'date','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,6,'timestamp','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,7,'timestampwithouttimezone','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,8,'bigint','{"jdbcType":11,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,9,'decimal','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,10,'TEXT','{"jdbcType":5,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,11,'bigserial','{"jdbcType":11,"javaColumnType":"2","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,12,'timestampwithtimezone','{"jdbcType":99,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,13,'numeric','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,14,'float8','{"jdbcType":12,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 14);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,15,'float4','{"jdbcType":12,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 15);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,16,'bpchar','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 16);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,17,'serial','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 17);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,18,'json','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 18);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,19,'bytea','{"jdbcType":2004,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 19);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,20,'bool','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 20);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,21,'array','{"jdbcType":2003}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 21);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,22,'numeric','{"jdbcType":2,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 22);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,1,'varchar','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,2,'int','{"jdbcType":4,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,3,'datetime','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,4,'smallint','{"jdbcType":5,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,5,'sysname','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,6,'date','{"jdbcType":93,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,7,'datetime2','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,8,'nvarchar','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,9,'intidentity','{"jdbcType":11,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,10,'decimal','{"jdbcType":11,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,11,'char','{"jdbcType":11,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,12,'bigint','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,13,'uniqueidentifier','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,14,'tinyint','{"jdbcType":4,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 14);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,15,'real','{"jdbcType":11,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 15);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,16,'nchar','{"jdbcType":11,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 16);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,1,'INT','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,2,'STRING','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,3,'BOOLEAN','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,4,'DECIMAL','{"jdbcType":3,"javaColumnType":"3","sparkColumnType":"3","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,5,'TIMESTAMP','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"string","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,6,'SMALLINT','{"jdbcType":5,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,7,'TINYINT','{"jdbcType":-6,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,8,'BINARY','{"jdbcType":-2,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,9,'VARCHAR','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,10,'CHAR','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,11,'DATE','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,12,'DOUBLE','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,13,'BIGINT','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 13);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,14,'FLOAT','{"jdbcType":6,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 14);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,1,'hbase-int','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,2,'hbase-string','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,3,'hbase-date','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,4,'hbase_long','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,5,'hbase_bytes','{"jdbcType":-2,"javaColumnType":"7","sparkColumnType":"8","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,6,'hbase_boolean','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,7,'hbase_double','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,8,'hbase_timestamp','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 8);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,1,'DOUBLE','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,2,'STRING','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,3,'OBJECT','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,4,'ARRAY','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,5,'DATE','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,6,'INT','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,7,'OBJECTID','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 7);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,8,'LONG','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 8);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,9,'BASICDBOBJECT','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 9);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,10,'INTERGER','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 10);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,11,'NULL','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 11);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,12,'INTEGER','{"jdbcType":11,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 12);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,13,'BASICDBLIST','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"非关系型数据库"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 13);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 9,4,'FTPLONG','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 9 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 9,5,'FTPDATE','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 9 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 9,6,'FTPTIMESTAMP','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 9 AND f_dict_key = 6);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 10,1,'HDFSSTRING','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 10 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 10,2,'HDFSINT','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 10 AND f_dict_key = 2);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,1,'SFTPLONG','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,2,'SFTPTIMESTAMP','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,3,'SFTPDATE','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,4,'SFTPDOUBLE','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,5,'SFTPSTRING','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 11,6,'FTPINT','{"jdbcType":6,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"文件系统"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 11 AND f_dict_key = 6);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,1,'kafkaString','{"jdbcType":12,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,2,'kafkaInteger','{"jdbcType":4,"javaColumnType":"1","sparkColumnType":"1","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,3,'kafkaLong','{"jdbcType":-5,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,4,'kafkaDouble','{"jdbcType":8,"javaColumnType":"3","sparkColumnType":"4","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,5,'kafkaBool','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,6,'kafkaDate','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 6);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 13,7,'kafkaTimestamp','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"消息队列"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 13 AND f_dict_key = 7);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 14,1,'api_boolean','{"jdbcType":16,"javaColumnType":"5","sparkColumnType":"6","flinkColumnType":"4","dbCatalogName":"其他"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 14 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 14,2,'api_date','{"jdbcType":91,"javaColumnType":"6","sparkColumnType":"7","flinkColumnType":"4","dbCatalogName":"其他"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 14 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 14,3,'api_timestamp','{"jdbcType":93,"javaColumnType":"8","sparkColumnType":"9","flinkColumnType":"4","dbCatalogName":"其他"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 14 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 14,4,'api_long','{"jdbcType":4,"javaColumnType":"2","sparkColumnType":"2","flinkColumnType":"4","dbCatalogName":"其他"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 14 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 14,5,'api_string','{"jdbcType":1,"javaColumnType":"4","sparkColumnType":"5","flinkColumnType":"4","dbCatalogName":"其他"}',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 14 AND f_dict_key = 5);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,1,'oracle.jdbc.OracleDriver','oracle驱动类',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,2,'com.mysql.cj.jdbc.Driver','mysql驱动类',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,3,'org.postgresql.Driver','postgresql驱动类',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,4,'com.microsoft.sqlserver.jdbc.SQLServerDriver','sqlserver驱动类',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,5,'org.apache.hive.jdbc.HiveDriver','hive驱动类',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 15,14,'jdbc:clickhouse://', 'clickhouse-jdbc前缀', 1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 15 AND f_dict_key = 14);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,1,'jdbc:oracle:thin:@//','oracle-jdbc前缀',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,2,'jdbc:mysql://','mysql-jdbc前缀',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,3,'jdbc:postgresql://','postgresql-jdbc前缀',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,4,'jdbc:sqlserver://','sqlserver-jdbc前缀',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,5,'jdbc:hive2://','hive2-jdbc前缀',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 16,14,'com.clickhouse.jdbc.ClickHouseDriver','clickhouse驱动类',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 16 AND f_dict_key = 14);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,1,'select 1 from dual','oracle-有效性检测',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 1);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,2,'select 1','mysql-有效性检测',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 2);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,3,'select version()','postgresql-有效性检测',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 3);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,4,'select 1','sqlserver-有效性检测',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 4);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,5,'select 1','hive2-有效性检测',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 5);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 17,14,'select 1','clickhouse有效性检测',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 17 AND f_dict_key = 14);


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 2,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 2 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 3,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 3 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 4,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 4 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 5,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 5 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 6,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 6 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 7,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 7 AND f_dict_key = 0);
INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 8,0,'UNKNOWN','',2 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 8 AND f_dict_key = 0);


-- 当前血缘初始化依赖于 AnyFabric，引擎从 AnyFabric 剥离后，临时跳过血缘的初始化逻辑。后续考虑整体方案。
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '1',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 1);
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '2',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 2);
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '3',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 3);
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '4',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 4);
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '5',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 5);
INSERT INTO t_lineage_graph_info (app_id,graph_id) SELECT '6',0 FROM DUAL WHERE NOT EXISTS ( SELECT app_id from t_lineage_graph_info where app_id = 6);
