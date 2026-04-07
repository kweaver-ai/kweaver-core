
SET SEARCH_PATH TO adp;


CREATE TABLE IF NOT EXISTS `t_model` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '模型名称',
  `f_description` VARCHAR(300) NOT NULL DEFAULT '' COMMENT '模型描述',
  `f_train_status` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '模型训练状态',
  `f_status` TINYINT NOT NULL COMMENT '状态',
  `f_rule` TEXT DEFAULT NULL COMMENT '数据标签',
  `f_userid` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '用户id',
  `f_type` TINYINT NOT NULL DEFAULT -1 COMMENT '模型类型',
  `f_created_at` BIGINT DEFAULT NULL COMMENT '创建时间',
  `f_updated_at` BIGINT DEFAULT NULL COMMENT '更新时间',
  `f_scope` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '用户作用域',
  PRIMARY KEY (`f_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_model_f_name` ON `t_model` (f_name);
CREATE INDEX IF NOT EXISTS `idx_t_model_f_userid_status` ON `t_model` (f_userid, f_status);
CREATE INDEX IF NOT EXISTS `idx_t_model_f_status_type` ON `t_model` (f_status, f_type);


CREATE TABLE IF NOT EXISTS `t_train_file` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_train_id` BIGINT UNSIGNED NOT NULL COMMENT '训练记录id',
  `f_oss_id` VARCHAR(36) DEFAULT '' COMMENT '应用存储的ossid',
  `f_key` VARCHAR(36) DEFAULT '' COMMENT '训练文件对象存储key',
  `f_created_at` BIGINT DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`f_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_train_file_f_train_id` ON `t_train_file` (f_train_id);


CREATE TABLE IF NOT EXISTS `t_automation_executor` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_name` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '节点名称',
  `f_description` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '节点描述',
  `f_creator_id` VARCHAR(40) NOT NULL COMMENT '创建者ID',
  `f_status` TINYINT NOT NULL COMMENT '状态 0 禁用 1 启用',
  `f_created_at` BIGINT DEFAULT NULL COMMENT '创建时间',
  `f_updated_at` BIGINT DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`f_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_automation_executor_name` ON `t_automation_executor` (`f_name`);
CREATE INDEX IF NOT EXISTS `idx_t_automation_executor_creator_id` ON `t_automation_executor` (`f_creator_id`);
CREATE INDEX IF NOT EXISTS `idx_t_automation_executor_status` ON `t_automation_executor` (`f_status`);


CREATE TABLE IF NOT EXISTS `t_automation_executor_accessor` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_executor_id` BIGINT UNSIGNED NOT NULL COMMENT '节点ID',
  `f_accessor_id` VARCHAR(40) NOT NULL COMMENT '访问者ID',
  `f_accessor_type` VARCHAR(20) NOT NULL COMMENT '访问者类型 user, department, group, contact',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_automation_executor_accessor_uk_executor_accessor` (`f_executor_id`, `f_accessor_id`, `f_accessor_type`)
);

CREATE INDEX IF NOT EXISTS `idx_t_automation_executor_accessor` ON `t_automation_executor_accessor` (`f_executor_id`, `f_accessor_id`, `f_accessor_type`);


CREATE TABLE IF NOT EXISTS `t_automation_executor_action` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_executor_id` BIGINT UNSIGNED NOT NULL COMMENT '节点ID',
  `f_operator` VARCHAR(64) NOT NULL COMMENT '动作标识',
  `f_name` VARCHAR(256) NOT NULL COMMENT '动作名称',
  `f_description` VARCHAR(1024) NOT NULL COMMENT '动作描述',
  `f_group` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '分组',
  `f_type` VARCHAR(16) NOT NULL DEFAULT 'python' COMMENT '节点类型',
  `f_inputs` MEDIUMTEXT COMMENT '节点输入',
  `f_outputs` MEDIUMTEXT COMMENT '节点输出',
  `f_config` MEDIUMTEXT COMMENT '节点配置',
  `f_created_at` BIGINT DEFAULT NULL COMMENT '创建时间',
  `f_updated_at` BIGINT DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`f_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_automation_executor_action_executor_id` ON `t_automation_executor_action` (`f_executor_id`);
CREATE INDEX IF NOT EXISTS `idx_t_automation_executor_action_operator` ON `t_automation_executor_action` (`f_operator`);
CREATE INDEX IF NOT EXISTS `idx_t_automation_executor_action_name` ON `t_automation_executor_action` (`f_name`);


CREATE TABLE IF NOT EXISTS `t_content_admin` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_user_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '用户id',
  `f_user_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '用户名称',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_content_admin_uk_f_user_id` (`f_user_id`)
);



CREATE TABLE IF NOT EXISTS `t_audio_segments` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_task_id` VARCHAR(32) NOT NULL COMMENT '任务id',
  `f_object` VARCHAR(1024) NOT NULL COMMENT '文件对象信息',
  `f_summary_type` VARCHAR(12) NOT NULL COMMENT '总结类型',
  `f_max_segments` TINYINT NOT NULL COMMENT '最大分段数',
  `f_max_segments_type` VARCHAR(12) NOT NULL COMMENT '分段类型',
  `f_need_abstract` TINYINT NOT NULL COMMENT '是否需要摘要',
  `f_abstract_type` VARCHAR(12) NOT NULL COMMENT '摘要总结方式',
  `f_callback` VARCHAR(1024) NOT NULL COMMENT '回调地址',
  `f_created_at` BIGINT DEFAULT NULL COMMENT '创建时间',
  `f_updated_at` BIGINT DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`f_id`)
);



CREATE TABLE IF NOT EXISTS `t_automation_conf` (
  `f_key` CHAR(32) NOT NULL,
  `f_value` CHAR(255) NOT NULL,
  PRIMARY KEY (`f_key`)
);


INSERT INTO `t_automation_conf` (f_key, f_value) SELECT 'process_template', 1 FROM DUAL WHERE NOT EXISTS(SELECT `f_key`, `f_value` FROM `t_automation_conf` WHERE `f_key`='process_template');

INSERT INTO `t_automation_conf` (f_key, f_value) SELECT 'ai_capabilities', 1 FROM DUAL WHERE NOT EXISTS(SELECT `f_key`, `f_value` FROM `t_automation_conf` WHERE `f_key`='ai_capabilities');


CREATE TABLE IF NOT EXISTS `t_automation_agent` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'Agent 名称',
  `f_agent_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Agent ID',
  `f_version` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'Agent 版本',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_automation_agent_uk_t_automation_agent_name` (`f_name`)
);

CREATE INDEX IF NOT EXISTS `idx_t_automation_agent_agent_id` ON `t_automation_agent` (`f_agent_id`);


CREATE TABLE IF NOT EXISTS `t_alarm_rule` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_rule_id` BIGINT UNSIGNED NOT NULL COMMENT '告警规则ID',
  `f_dag_id` BIGINT UNSIGNED NOT NULL COMMENT '流程ID',
  `f_frequency` SMALLINT(6) UNSIGNED NOT NULL COMMENT '频率',
  `f_threshold` MEDIUMINT(9) UNSIGNED NOT NULL COMMENT '阈值',
  `f_created_at` BIGINT DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`f_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_alarm_rule_rule_id` ON `t_alarm_rule` (`f_rule_id`);


CREATE TABLE IF NOT EXISTS `t_alarm_user` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_rule_id` BIGINT UNSIGNED NOT NULL COMMENT '告警规则ID',
  `f_user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
  `f_user_name` VARCHAR(128) NOT NULL COMMENT '用户名称',
  `f_user_type` VARCHAR(10) NOT NULL COMMENT '用户类型,取值: user,group',
  PRIMARY KEY (`f_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_alarm_user_rule_id` ON `t_alarm_user` (`f_rule_id`);


CREATE TABLE IF NOT EXISTS `t_automation_dag_instance_ext_data` (
  `f_id` VARCHAR(64) NOT NULL COMMENT '主键id',
  `f_created_at` BIGINT DEFAULT NULL COMMENT '创建时间',
  `f_updated_at` BIGINT DEFAULT NULL COMMENT '更新时间',
  `f_dag_id` VARCHAR(64) COMMENT 'DAG id',
  `f_dag_ins_id` VARCHAR(64) COMMENT 'DAG实例id',
  `f_field` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '字段名称',
  `f_oss_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'OSS存储id',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_size` BIGINT UNSIGNED DEFAULT NULL COMMENT '文件大小',
  `f_removed` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否删除(1:未删除,0:已删除)',
  PRIMARY KEY (`f_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_automation_dag_instance_ext_data_dag_ins_id` ON `t_automation_dag_instance_ext_data` (`f_dag_ins_id`);


CREATE TABLE IF NOT EXISTS `t_task_cache_0` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_0_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_0_idx_expire_time` ON `t_task_cache_0` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_1` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_1_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_1_idx_expire_time` ON `t_task_cache_1` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_2` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_2_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_2_idx_expire_time` ON `t_task_cache_2` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_3` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_3_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_3_idx_expire_time` ON `t_task_cache_3` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_4` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_4_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_4_idx_expire_time` ON `t_task_cache_4` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_5` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_5_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_5_idx_expire_time` ON `t_task_cache_5` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_6` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_6_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_6_idx_expire_time` ON `t_task_cache_6` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_7` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_7_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_7_idx_expire_time` ON `t_task_cache_7` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_8` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_8_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_8_idx_expire_time` ON `t_task_cache_8` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_9` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_9_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_9_idx_expire_time` ON `t_task_cache_9` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_a` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_a_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_a_idx_expire_time` ON `t_task_cache_a` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_b` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_b_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_b_idx_expire_time` ON `t_task_cache_b` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_c` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_c_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_c_idx_expire_time` ON `t_task_cache_c` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_d` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_d_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_d_idx_expire_time` ON `t_task_cache_d` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_e` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_e_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_e_idx_expire_time` ON `t_task_cache_e` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_task_cache_f` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_hash` CHAR(40) NOT NULL DEFAULT '' COMMENT '任务hash',
  `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务类型',
  `f_status` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '任务状态(1 处理中, 2 成功, 3 失败)',
  `f_oss_id` CHAR(36) NOT NULL DEFAULT '' COMMENT '对象存储ID',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_ext` CHAR(20) NOT NULL DEFAULT '' COMMENT '副文档后缀名',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '副文档大小',
  `f_err_msg` TEXT NULL DEFAULT NULL COMMENT '错误信息',
  `f_create_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '创建时间',
  `f_modify_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `f_expire_time` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_task_cache_f_uk_hash` (`f_hash`)
);

CREATE INDEX IF NOT EXISTS `idx_t_task_cache_f_idx_expire_time` ON `t_task_cache_f` (`f_expire_time`);


CREATE TABLE IF NOT EXISTS `t_dag_instance_event` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键id',
  `f_type` TINYINT(4) NOT NULL DEFAULT '0' COMMENT '事件类型',
  `f_instance_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'DAG实例ID',
  `f_operator` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '节点标识',
  `f_task_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '任务ID',
  `f_status` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '任务状态',
  `f_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '变量名称',
  `f_data` LONGTEXT NOT NULL COMMENT '数据',
  `f_size` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '数据大小',
  `f_inline` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '是否内联',
  `f_visibility` TINYINT(2) NOT NULL DEFAULT '0' COMMENT '可见性(0: private, 1: public)',
  `f_timestamp` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '时间戳',
  PRIMARY KEY (`f_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_dag_instance_event_idx_instance_id` ON `t_dag_instance_event` (`f_instance_id`, `f_id`);
CREATE INDEX IF NOT EXISTS `idx_t_dag_instance_event_idx_instance_type_vis` ON `t_dag_instance_event` (`f_instance_id`, `f_type`, `f_visibility`, `f_id`);
CREATE INDEX IF NOT EXISTS `idx_t_dag_instance_event_idx_instance_name_type` ON `t_dag_instance_event` (`f_instance_id`, `f_name`, `f_type`, `f_id`);


CREATE TABLE IF NOT EXISTS `t_cron_job`
  (
  `f_key_id` BIGSERIAL NOT NULL COMMENT '自增长ID',
  `f_job_id` VARCHAR(36) NOT NULL COMMENT '任务ID',
  `f_job_name` VARCHAR(64) NOT NULL COMMENT '任务名称',
  `f_job_cron_time` VARCHAR(32) NOT NULL COMMENT '时间计划，cron表达式',
  `f_job_type` TINYINT(4) NOT NULL COMMENT '任务类型，参考数据字典',
  `f_job_context` VARCHAR(10240) COMMENT '参考任务上下文数据结构',
  `f_tenant_id` VARCHAR(36) COMMENT '任务来源ID',
  `f_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '启用/禁用标识',
  `f_remarks` VARCHAR(256) COMMENT '备注',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`f_key_id`),
  UNIQUE KEY `idx_t_cron_job_index_job_id` (`f_job_id`),
  UNIQUE KEY `idx_t_cron_job_index_job_name` (`f_job_name`, `f_tenant_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_cron_job_index_tenant_id` ON `t_cron_job` (`f_tenant_id`);
CREATE INDEX IF NOT EXISTS `idx_t_cron_job_index_time` ON `t_cron_job` (`f_create_time`, `f_update_time`);


CREATE TABLE IF NOT EXISTS `t_cron_job_status`
  (
  `f_key_id` BIGSERIAL NOT NULL COMMENT '自增长ID',
  `f_execute_id` VARCHAR(36) NOT NULL COMMENT '执行编号，流水号',
  `f_job_id` VARCHAR(36) NOT NULL COMMENT '任务ID',
  `f_job_type` TINYINT(4) NOT NULL COMMENT '任务类型',
  `f_job_name` VARCHAR(64) NOT NULL COMMENT '任务名称',
  `f_job_status` TINYINT(4) NOT NULL COMMENT '任务状态，参考数据字典',
  `f_begin_time` BIGINT(20) COMMENT '任务本次执行开始时间',
  `f_end_time` BIGINT(20) COMMENT '任务本次执行结束时间',
  `f_executor` VARCHAR(1024) COMMENT '任务执行者',
  `f_execute_times` INT COMMENT '任务执行次数',
  `f_ext_info` VARCHAR(1024) COMMENT '扩展信息',
  PRIMARY KEY (`f_key_id`),
  UNIQUE KEY `idx_t_cron_job_status_index_execute_id` (`f_execute_id`)
);

CREATE INDEX IF NOT EXISTS `idx_t_cron_job_status_index_job_id` ON `t_cron_job_status` (`f_job_id`);
CREATE INDEX IF NOT EXISTS `idx_t_cron_job_status_index_job_status` ON `t_cron_job_status` (`f_job_status`);
CREATE INDEX IF NOT EXISTS `idx_t_cron_job_status_index_time` ON `t_cron_job_status` (`f_begin_time`,`f_end_time`);

CREATE TABLE IF NOT EXISTS `t_flow_dag` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_user_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '用户ID',
 `f_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'DAG名称',
 `f_desc` VARCHAR(310) NOT NULL DEFAULT '' COMMENT 'DAG描述',
 `f_trigger` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '触发器配置',
 `f_cron` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Cron表达式',
 `f_vars` MEDIUMTEXT DEFAULT NULL COMMENT '变量定义',
 `f_status` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '状态',
 `f_tasks` MEDIUMTEXT DEFAULT NULL COMMENT '任务配置',
 `f_steps` MEDIUMTEXT DEFAULT NULL COMMENT '步骤配置',
 `f_description` VARCHAR(310) NOT NULL DEFAULT '' COMMENT '详细描述',
 `f_shortcuts` TEXT DEFAULT NULL COMMENT '快捷配置',
 `f_accessors` TEXT DEFAULT NULL COMMENT '访问者列表',
 `f_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'DAG类型',
 `f_policy_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '策略类型',
 `f_appinfo` TEXT DEFAULT NULL COMMENT '应用信息',
 `f_priority` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '优先级',
 `f_removed` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '删除标记',
 `f_emails` TEXT DEFAULT NULL COMMENT '通知邮箱',
 `f_template` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '模板标识',
 `f_published` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '发布标记',
 `f_trigger_config` TEXT DEFAULT NULL COMMENT '触发器配置详情',
 `f_sub_ids` TEXT DEFAULT NULL COMMENT '子ID列表',
 `f_exec_mode` VARCHAR(8) NOT NULL DEFAULT '' COMMENT '执行模式',
 `f_category` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '分类',
 `f_outputs` MEDIUMTEXT DEFAULT NULL COMMENT '输出定义',
 `f_instructions` MEDIUMTEXT DEFAULT NULL COMMENT '操作说明',
 `f_operator_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '操作人ID',
 `f_inc_values` VARCHAR(4096) DEFAULT NULL COMMENT '增量值',
 `f_version` VARCHAR(64) DEFAULT NULL COMMENT '版本信息',
 `f_version_id` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '版本ID',
 `f_modify_by` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '修改人',
 `f_is_debug` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '调试标记',
 `f_debug_id` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '调试ID',
 `f_biz_domain_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务域ID',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_dag_user_id` ON `t_flow_dag` (`f_user_id`);
CREATE INDEX IF NOT EXISTS `idx_dag_type` ON `t_flow_dag` (`f_type`);
CREATE INDEX IF NOT EXISTS `idx_dag_trigger` ON `t_flow_dag` (`f_trigger`);
CREATE INDEX IF NOT EXISTS `idx_dag_name` ON `t_flow_dag` (`f_name`);
CREATE INDEX IF NOT EXISTS `idx_dag_biz_domain` ON `t_flow_dag` (`f_biz_domain_id`);

CREATE TABLE IF NOT EXISTS `t_flow_dag_var` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_dag_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'DAG ID',
 `f_var_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变量名',
 `f_default_value` TEXT DEFAULT NULL COMMENT '默认值',
 `f_var_type` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '变量类型',
 `f_description` TEXT DEFAULT NULL COMMENT '变量描述',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_dag_vars_dag_id` ON `t_flow_dag_var` (`f_dag_id`);

CREATE TABLE IF NOT EXISTS `t_flow_dag_instance_keyword` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_dag_ins_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'DAG实例ID',
 `f_keyword` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关键词',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_dag_ins_kw` ON `t_flow_dag_instance_keyword` (`f_dag_ins_id`, `f_keyword`);

CREATE TABLE IF NOT EXISTS `t_flow_dag_step` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_dag_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'DAG ID',
 `f_operator` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作符',
 `f_source_id` TEXT NOT NULL COMMENT '来源ID',
 `f_has_datasource` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否有数据源',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_dag_step_op` ON `t_flow_dag_step` (`f_operator`);
CREATE INDEX IF NOT EXISTS `idx_dag_step_op_dag` ON `t_flow_dag_step` (`f_dag_id`, `f_operator`);
CREATE INDEX IF NOT EXISTS `idx_dag_step_has_ds_dag` ON `t_flow_dag_step` (`f_dag_id`, `f_has_datasource`);

CREATE TABLE IF NOT EXISTS `t_flow_dag_accessor` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_dag_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'DAG ID',
 `f_accessor_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '访问者ID',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_dag_accessor_id_dag` ON `t_flow_dag_accessor` (`f_accessor_id`, `f_dag_id`);

CREATE TABLE IF NOT EXISTS `t_flow_dag_version` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_dag_id` VARCHAR(20) NOT NULL DEFAULT '' COMMENT 'DAG ID',
 `f_user_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '用户ID',
 `f_version` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '版本号',
 `f_version_id` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '版本ID',
 `f_change_log` VARCHAR(512) DEFAULT NULL COMMENT '变更记录',
 `f_config` LONGTEXT DEFAULT NULL COMMENT '配置内容',
 `f_sort_time` BIGINT NOT NULL DEFAULT 0 COMMENT '排序时间',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_dag_versions_dag_version` ON `t_flow_dag_version` (`f_version_id`, `f_dag_id`);
CREATE INDEX IF NOT EXISTS `idx_dag_versions_dag_sort` ON `t_flow_dag_version` (`f_dag_id`, `f_sort_time`);

CREATE TABLE IF NOT EXISTS `t_flow_dag_instance` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_dag_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'DAG ID',
 `f_trigger` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '触发器配置',
 `f_worker` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '执行节点',
 `f_source` TEXT DEFAULT NULL COMMENT '来源',
 `f_vars` MEDIUMTEXT DEFAULT NULL COMMENT '变量',
 `f_keywords` TEXT DEFAULT NULL COMMENT '关键词',
 `f_event_persistence` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '事件持久化标记',
 `f_event_oss_path` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '事件OSS路径',
 `f_share_data` MEDIUMTEXT DEFAULT NULL COMMENT '共享数据',
 `f_share_data_ext` MEDIUMTEXT DEFAULT NULL COMMENT '共享数据扩展',
 `f_status` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '状态',
 `f_reason` MEDIUMTEXT DEFAULT NULL COMMENT '原因',
 `f_cmd` TEXT DEFAULT NULL COMMENT '命令',
 `f_has_cmd` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否包含命令',
 `f_batch_run_id` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '批次运行ID',
 `f_user_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '用户ID',
 `f_ended_at` BIGINT NOT NULL DEFAULT 0 COMMENT '结束时间',
 `f_dag_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'DAG类型',
 `f_policy_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '策略类型',
 `f_appinfo` TEXT DEFAULT NULL COMMENT '应用信息',
 `f_priority` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '优先级',
 `f_mode` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '模式',
 `f_dump` LONGTEXT DEFAULT NULL COMMENT 'Dump数据',
 `f_dump_ext` LONGTEXT DEFAULT NULL COMMENT 'Dump扩展',
 `f_success_callback` VARCHAR(1024) DEFAULT NULL COMMENT '成功回调',
 `f_error_callback` VARCHAR(1024) DEFAULT NULL COMMENT '失败回调',
 `f_call_chain` TEXT DEFAULT NULL COMMENT '调用链',
 `f_resume_data` TEXT DEFAULT NULL COMMENT '恢复数据',
 `f_resume_status` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '恢复状态',
 `f_version` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '版本号',
 `f_version_id` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '版本ID',
 `f_biz_domain_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务域ID',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_dag_ins_dag_status` ON `t_flow_dag_instance` (`f_dag_id`, `f_status`);
CREATE INDEX IF NOT EXISTS `idx_dag_ins_status_upd` ON `t_flow_dag_instance` (`f_status`, `f_updated_at`);
CREATE INDEX IF NOT EXISTS `idx_dag_ins_status_user_pri` ON `t_flow_dag_instance` (`f_status`, `f_user_id`, `f_priority`);
CREATE INDEX IF NOT EXISTS `idx_dag_ins_user_id` ON `t_flow_dag_instance` (`f_user_id`);
CREATE INDEX IF NOT EXISTS `idx_dag_ins_batch_run` ON `t_flow_dag_instance` (`f_batch_run_id`);
CREATE INDEX IF NOT EXISTS `idx_dag_ins_worker` ON `t_flow_dag_instance` (`f_worker`);

CREATE TABLE IF NOT EXISTS `t_flow_inbox` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_msg` MEDIUMTEXT DEFAULT NULL COMMENT '消息内容',
 `f_topic` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '主题',
 `f_docid` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文档ID',
 `f_dag` TEXT DEFAULT NULL COMMENT 'DAG列表',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_inbox_docid` ON `t_flow_inbox` (`f_docid`);
CREATE INDEX IF NOT EXISTS `idx_inbox_topic_created` ON `t_flow_inbox` (`f_topic`, `f_created_at`);

CREATE TABLE IF NOT EXISTS `t_flow_outbox` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_msg` MEDIUMTEXT DEFAULT NULL COMMENT '消息内容',
 `f_topic` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '主题',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_outbox_created` ON `t_flow_outbox` (`f_created_at`);

CREATE TABLE IF NOT EXISTS `t_flow_task_instance` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_expired_at` BIGINT NOT NULL DEFAULT 0 COMMENT '过期时间',
 `f_task_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '任务ID',
 `f_dag_ins_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'DAG实例ID',
 `f_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '任务名称',
 `f_depend_on` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '依赖关系',
 `f_action_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '动作名称',
 `f_timeout_secs` BIGINT NOT NULL DEFAULT 0 COMMENT '超时时间(秒)',
 `f_params` MEDIUMTEXT DEFAULT NULL COMMENT '参数',
 `f_traces` MEDIUMTEXT DEFAULT NULL COMMENT '链路信息',
 `f_status` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '状态',
 `f_reason` MEDIUMTEXT DEFAULT NULL COMMENT '原因',
 `f_pre_checks` TEXT DEFAULT NULL COMMENT '预检查',
 `f_results` MEDIUMTEXT DEFAULT NULL COMMENT '结果',
 `f_steps` MEDIUMTEXT DEFAULT NULL COMMENT '步骤',
 `f_last_modified_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最后修改时间',
 `f_rendered_params` LONGTEXT DEFAULT NULL COMMENT '渲染参数',
 `f_hash` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '哈希',
 `f_settings` LONGTEXT DEFAULT NULL COMMENT '配置',
 `f_metadata` LONGTEXT DEFAULT NULL COMMENT '元数据',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_task_ins_dag_ins_id` ON `t_flow_task_instance` (`f_dag_ins_id`);
CREATE INDEX IF NOT EXISTS `idx_task_ins_hash` ON `t_flow_task_instance` (`f_hash`);
CREATE INDEX IF NOT EXISTS `idx_task_ins_action` ON `t_flow_task_instance` (`f_action_name`);
CREATE INDEX IF NOT EXISTS `idx_task_ins_status_expire` ON `t_flow_task_instance` (`f_status`, `f_expired_at`);
CREATE INDEX IF NOT EXISTS `idx_task_ins_status_upd_id` ON `t_flow_task_instance` (`f_status`, `f_updated_at`, `f_id`);

CREATE TABLE IF NOT EXISTS `t_flow_token` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_user_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '用户ID',
 `f_user_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '用户名',
 `f_refresh_token` TEXT DEFAULT NULL COMMENT '刷新令牌',
 `f_token` TEXT DEFAULT NULL COMMENT '访问令牌',
 `f_expires_in` INT NOT NULL DEFAULT 0 COMMENT '过期时间(秒)',
 `f_login_ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '登录IP',
 `f_is_app` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否应用',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_token_user_id` ON `t_flow_token` (`f_user_id`);

CREATE TABLE IF NOT EXISTS `t_flow_client` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_client_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端名称',
 `f_client_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '客户端ID',
 `f_client_secret` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '客户端密钥',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_client_name` ON `t_flow_client` (`f_client_name`);

CREATE TABLE IF NOT EXISTS `t_flow_switch` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '开关名称',
 `f_status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '开关状态',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_switch_name` ON `t_flow_switch` (`f_name`);

CREATE TABLE IF NOT EXISTS `t_flow_log` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_ossid` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'OSS ID',
 `f_key` VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'OSS Key',
 `f_filename` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '文件名',
  PRIMARY KEY (`f_id`)
);

CREATE TABLE IF NOT EXISTS `t_flow_storage` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_oss_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'OssGateway存储ID',
 `f_object_key` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '对象存储key',
 `f_name` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '原始文件名',
 `f_content_type` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'MIME类型',
 `f_size` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件大小',
 `f_etag` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '文件etag/hash',
 `f_status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 2待删除 3已删除',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间 0表示未删除',
  PRIMARY KEY (`f_id`)
);
CREATE UNIQUE INDEX IF NOT EXISTS `uk_flow_storage_oss_id_object_key` ON `t_flow_storage` (`f_oss_id`, `f_object_key`);
CREATE INDEX IF NOT EXISTS `idx_flow_storage_status` ON `t_flow_storage` (`f_status`);
CREATE INDEX IF NOT EXISTS `idx_flow_storage_created_at` ON `t_flow_storage` (`f_created_at`);

CREATE TABLE IF NOT EXISTS `t_flow_file` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID，对应 dfs://<id>',
 `f_dag_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '流程定义ID',
 `f_dag_instance_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '流程实例ID',
 `f_storage_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '存储文件ID，未落OSS时为0',
 `f_status` TINYINT NOT NULL DEFAULT 1 COMMENT '业务状态 1待就绪 2就绪 3失效 4已过期',
 `f_name` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '文件名',
 `f_expires_at` BIGINT NOT NULL DEFAULT 0 COMMENT '过期时间 0表示不过期',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`f_id`)
);
CREATE INDEX IF NOT EXISTS `idx_flow_file_dag_id` ON `t_flow_file` (`f_dag_id`);
CREATE INDEX IF NOT EXISTS `idx_flow_file_dag_instance_id` ON `t_flow_file` (`f_dag_instance_id`);
CREATE INDEX IF NOT EXISTS `idx_flow_file_storage_id` ON `t_flow_file` (`f_storage_id`);
CREATE INDEX IF NOT EXISTS `idx_flow_file_status` ON `t_flow_file` (`f_status`);
CREATE INDEX IF NOT EXISTS `idx_flow_file_expires_at` ON `t_flow_file` (`f_expires_at`);

CREATE TABLE IF NOT EXISTS `t_flow_file_download_job` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_file_id` BIGINT UNSIGNED NOT NULL COMMENT '关联flow_file ID',
 `f_status` TINYINT NOT NULL DEFAULT 1 COMMENT '任务状态 1待执行 2执行中 3成功 4失败 5取消',
 `f_retry_count` INT NOT NULL DEFAULT 0 COMMENT '已重试次数',
 `f_max_retry` INT NOT NULL DEFAULT 3 COMMENT '最大重试次数',
 `f_next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间',
 `f_error_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '错误码',
 `f_error_message` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '错误信息',
 `f_download_url` VARCHAR(2048) NOT NULL DEFAULT '' COMMENT '源文件URL',
 `f_started_at` BIGINT NOT NULL DEFAULT 0 COMMENT '开始时间',
 `f_finished_at` BIGINT NOT NULL DEFAULT 0 COMMENT '结束时间',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`f_id`)
);
CREATE UNIQUE INDEX IF NOT EXISTS `uk_flow_file_download_job_file_id` ON `t_flow_file_download_job` (`f_file_id`);
CREATE INDEX IF NOT EXISTS `idx_flow_file_download_job_status_retry` ON `t_flow_file_download_job` (`f_status`, `f_next_retry_at`);

CREATE TABLE IF NOT EXISTS `t_flow_task_resume` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_task_instance_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '被阻塞的任务实例ID',
 `f_dag_instance_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属流程实例ID',
 `f_resource_type` VARCHAR(32) NOT NULL DEFAULT 'file' COMMENT '资源类型',
 `f_resource_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '资源ID，对文件场景即flow_file ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`f_id`)
);
CREATE UNIQUE INDEX IF NOT EXISTS `uk_flow_task_resume_task_instance_id` ON `t_flow_task_resume` (`f_task_instance_id`);
CREATE INDEX IF NOT EXISTS `idx_flow_task_resume_resource` ON `t_flow_task_resume` (`f_resource_type`, `f_resource_id`);