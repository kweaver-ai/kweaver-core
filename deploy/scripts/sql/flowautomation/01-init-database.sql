-- Source: autoflow/coderunner/migrations/mariadb/0.1.0/pre/init.sql
USE adp;

CREATE TABLE IF NOT EXISTS `t_python_package` (
  `f_id` varchar(32) NOT NULL COMMENT '主键ID',
  `f_name` varchar(255) NOT NULL DEFAULT '' COMMENT '名称',
  `f_oss_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'ossid',
  `f_oss_key` varchar(32) NOT NULL DEFAULT '' COMMENT 'key',
  `f_creator_id` varchar(36) NOT NULL DEFAULT '' COMMENT '创建者id',
  `f_creator_name` varchar(128) NOT NULL DEFAULT '' COMMENT '创建者名称',
  `f_created_at` bigint(20) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `uk_t_python_package_name` (`f_name`)
) ENGINE=InnoDB COMMENT='包管理表';
-- Source: autoflow/ecron/migrations/mariadb/0.1.0/pre/init.sql
/*
MySQL: Database - ecron
create table
*********************************************************************
*/
use adp;
CREATE TABLE IF NOT EXISTS `t_cron_job`
(
    `f_key_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增长ID',
    `f_job_id` varchar(36) NOT NULL COMMENT '任务ID',
    `f_job_name` varchar(64) NOT NULL COMMENT '任务名称',
    `f_job_cron_time` varchar(32) NOT NULL COMMENT '时间计划，cron表达式',
    `f_job_type` tinyint(4) NOT NULL COMMENT '任务类型，参考数据字典',
    `f_job_context` varchar(10240) COMMENT '参考任务上下文数据结构',
    `f_tenant_id` varchar(36) COMMENT '任务来源ID',
    `f_enabled` tinyint(2) NOT NULL DEFAULT 1 COMMENT '启用/禁用标识',
    `f_remarks` varchar(256) COMMENT '备注',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_time` bigint(20) NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`f_key_id`),
    UNIQUE KEY `index_job_id`(`f_job_id`) USING BTREE,
    UNIQUE KEY `index_job_name`(`f_job_name`, `f_tenant_id`) USING BTREE,
    KEY `index_tenant_id`(`f_tenant_id`) USING BTREE,
    KEY `index_time`(`f_create_time`, `f_update_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '定时任务信息表';

CREATE TABLE IF NOT EXISTS `t_cron_job_status`
(
    `f_key_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增长ID',
    `f_execute_id` varchar(36) NOT NULL COMMENT '执行编号，流水号',
    `f_job_id` varchar(36) NOT NULL COMMENT '任务ID',
    `f_job_type` tinyint(4) NOT NULL COMMENT '任务类型',
    `f_job_name` varchar(64) NOT NULL COMMENT '任务名称',
    `f_job_status` tinyint(4) NOT NULL COMMENT '任务状态，参考数据字典',
    `f_begin_time` bigint(20) COMMENT '任务本次执行开始时间',
    `f_end_time` bigint(20) COMMENT '任务本次执行结束时间',
    `f_executor` varchar(1024) COMMENT '任务执行者',
    `f_execute_times` int COMMENT '任务执行次数',
    `f_ext_info` varchar(1024) COMMENT '扩展信息',
    PRIMARY KEY (`f_key_id`),
    UNIQUE KEY `index_execute_id`(`f_execute_id`) USING BTREE,
    KEY `index_job_id`(`f_job_id`) USING BTREE,
    KEY `index_job_status`(`f_job_status`) USING BTREE,
    KEY `index_time`(`f_begin_time`,`f_end_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '定时任务状态表';

-- Source: autoflow/flow-automation/migrations/mariadb/0.1.0/pre/init.sql
USE adp;

-- ----------------------------
-- workflow.t_model definition
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_model` (
  `f_id` bigint unsigned NOT NULL COMMENT '主键id',
  `f_name` varchar(255) NOT NULL DEFAULT '' COMMENT '模型名称',
  `f_description` varchar(300) NOT NULL DEFAULT '' COMMENT '模型描述',
  `f_train_status` varchar(16) NOT NULL DEFAULT '' COMMENT '模型训练状态',
  `f_status` tinyint NOT NULL COMMENT '状态',
  `f_rule` text DEFAULT NULL COMMENT '数据标签',
  `f_userid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户id',
  `f_type` tinyint NOT NULL DEFAULT -1 COMMENT '模型类型',
  `f_created_at` bigint DEFAULT NULL COMMENT '创建时间',
  `f_updated_at` bigint DEFAULT NULL COMMENT '更新时间',
  `f_scope` varchar(40) NOT NULL DEFAULT '' COMMENT '用户作用域',
  PRIMARY KEY (`f_id`),
  KEY idx_t_model_f_name (f_name),
  KEY idx_t_model_f_userid_status (f_userid, f_status),
  KEY idx_t_model_f_status_type (f_status, f_type)
) ENGINE=InnoDB COMMENT '模型记录表';


-- ----------------------------
-- workflow.t_train_file definition
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_train_file` (
  `f_id` bigint unsigned NOT NULL COMMENT '主键id',
  `f_train_id` bigint unsigned NOT NULL COMMENT '训练记录id',
  `f_oss_id` varchar(36) DEFAULT '' COMMENT '应用存储的ossid',
  `f_key` varchar(36) DEFAULT '' COMMENT '训练文件对象存储key',
  `f_created_at` bigint DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`f_id`),
  KEY idx_t_train_file_f_train_id (f_train_id)
) ENGINE=InnoDB COMMENT '模型训练文件记录表';


CREATE TABLE IF NOT EXISTS `t_automation_executor` (
  `f_id` bigint unsigned NOT NULL COMMENT '主键id',
  `f_name` varchar(256) NOT NULL DEFAULT '' COMMENT '节点名称',
  `f_description` varchar(1024) NOT NULL DEFAULT '' COMMENT '节点描述',
  `f_creator_id` varchar(40) NOT NULL COMMENT '创建者ID',
  `f_status` tinyint NOT NULL COMMENT '状态 0 禁用 1 启用',
  `f_created_at` bigint DEFAULT NULL COMMENT '创建时间',
  `f_updated_at` bigint DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`f_id`),
  KEY `idx_t_automation_executor_name` (`f_name`),
  KEY `idx_t_automation_executor_creator_id` (`f_creator_id`),
  KEY `idx_t_automation_executor_status` (`f_status`)
) ENGINE=InnoDB COMMENT '自定义节点';

CREATE TABLE IF NOT EXISTS `t_automation_executor_accessor` (
  `f_id` bigint unsigned NOT NULL COMMENT '主键id',
  `f_executor_id` bigint unsigned NOT NULL COMMENT '节点ID',
  `f_accessor_id` varchar(40) NOT NULL COMMENT '访问者ID',
  `f_accessor_type` varchar(20) NOT NULL COMMENT '访问者类型 user, department, group, contact',
  PRIMARY KEY (`f_id`),
  KEY `idx_t_automation_executor_accessor` (`f_executor_id`, `f_accessor_id`, `f_accessor_type`),
  UNIQUE KEY `uk_executor_accessor` (`f_executor_id`, `f_accessor_id`, `f_accessor_type`)
) ENGINE=InnoDB COMMENT '自定义节点访问者';

CREATE TABLE IF NOT EXISTS `t_automation_executor_action` (
  `f_id` bigint unsigned NOT NULL COMMENT '主键id',
  `f_executor_id` bigint unsigned NOT NULL COMMENT '节点ID',
  `f_operator` varchar(64) NOT NULL COMMENT '动作标识',
  `f_name` varchar(256) NOT NULL COMMENT '动作名称',
  `f_description` varchar(1024) NOT NULL COMMENT '动作描述',
  `f_group` varchar(64) NOT NULL DEFAULT '' COMMENT '分组',
  `f_type` varchar(16) NOT NULL DEFAULT 'python' COMMENT '节点类型',
  `f_inputs` mediumtext COMMENT '节点输入',
  `f_outputs` mediumtext COMMENT '节点输出',
  `f_config` mediumtext COMMENT '节点配置',
  `f_created_at` bigint DEFAULT NULL COMMENT '创建时间',
  `f_updated_at` bigint DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`f_id`),
  KEY `idx_t_automation_executor_action_executor_id` (`f_executor_id`),
  KEY `idx_t_automation_executor_action_operator` (`f_operator`),
  KEY `idx_t_automation_executor_action_name` (`f_name`)
) ENGINE=InnoDB COMMENT '节点动作';

CREATE TABLE IF NOT EXISTS `t_content_admin` (
  `f_id` bigint unsigned NOT NULL COMMENT '主键id',
  `f_user_id` varchar(40) NOT NULL DEFAULT '' COMMENT '用户id',
  `f_user_name` varchar(128) NOT NULL DEFAULT '' COMMENT '用户名称',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `uk_f_user_id` (`f_user_id`)
) ENGINE=InnoDB COMMENT='管理员表';

CREATE TABLE IF NOT EXISTS `t_audio_segments` (
  `f_id` bigint unsigned NOT NULL COMMENT '主键id',
  `f_task_id` varchar(32) NOT NULL COMMENT '任务id',
  `f_object` varchar(1024) NOT NULL COMMENT '文件对象信息',
  `f_summary_type` varchar(12) NOT NULL COMMENT '总结类型',
  `f_max_segments` tinyint NOT NULL COMMENT '最大分段数',
  `f_max_segments_type` varchar(12) NOT NULL COMMENT '分段类型',
  `f_need_abstract` tinyint NOT NULL COMMENT '是否需要摘要',
  `f_abstract_type` varchar(12) NOT NULL COMMENT '摘要总结方式',
  `f_callback` varchar(1024) NOT NULL COMMENT '回调地址',
  `f_created_at` bigint DEFAULT NULL COMMENT '创建时间',
  `f_updated_at` bigint DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`f_id`)
) ENGINE=InnoDB COMMENT '音频转换任务记录表';


CREATE TABLE  IF NOT EXISTS `t_automation_conf` (
  `f_key` char(32) NOT NULL,
  `f_value` char(255) NOT NULL,
  PRIMARY KEY (`f_key`)
) ENGINE=InnoDB COMMENT '自动化配置';

INSERT INTO `t_automation_conf` (f_key, f_value) SELECT 'process_template', 1 FROM DUAL WHERE NOT EXISTS(SELECT `f_key`, `f_value` FROM `t_automation_conf` WHERE `f_key`='process_template');
INSERT INTO `t_automation_conf` (f_key, f_value) SELECT 'ai_capabilities', 1 FROM DUAL WHERE NOT EXISTS(SELECT `f_key`, `f_value` FROM `t_automation_conf` WHERE `f_key`='ai_capabilities');

CREATE TABLE IF NOT EXISTS `t_automation_agent` (
  `f_id` bigint unsigned NOT NULL COMMENT '主键id',
  `f_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'Agent 名称',
  `f_agent_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'Agent ID',
  `f_version` varchar(32) NOT NULL DEFAULT '' COMMENT 'Agent 版本',
  PRIMARY KEY (`f_id`),
  KEY `idx_t_automation_agent_agent_id` (`f_agent_id`),
  UNIQUE KEY `uk_t_automation_agent_name` (`f_name`)
) ENGINE=InnoDB COMMENT 'Agent';

CREATE TABLE IF NOT EXISTS `t_alarm_rule` (
  `f_id` bigint unsigned NOT NULL COMMENT '主键id',
  `f_rule_id` bigint unsigned NOT NULL COMMENT '告警规则ID',
  `f_dag_id` bigint unsigned NOT NULL COMMENT '流程ID',
  `f_frequency` smallint(6) unsigned NOT NULL COMMENT '频率',
  `f_threshold` mediumint(9) unsigned NOT NULL COMMENT '阈值',
  `f_created_at` bigint DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`f_id`),
  KEY `idx_t_alarm_rule_rule_id` (`f_rule_id`)
) ENGINE=InnoDB COMMENT '告警规则';

CREATE TABLE IF NOT EXISTS `t_alarm_user` (
  `f_id` bigint unsigned NOT NULL COMMENT '主键id',
  `f_rule_id` bigint unsigned NOT NULL COMMENT '告警规则ID',
  `f_user_id` varchar(36) NOT NULL COMMENT '用户ID',
  `f_user_name` varchar(128) NOT NULL COMMENT '用户名称',
  `f_user_type` varchar(10) NOT NULL COMMENT '用户类型,取值: user,group',
  PRIMARY KEY (`f_id`),
  KEY `idx_t_alarm_user_rule_id` (`f_rule_id`)
) ENGINE=InnoDB COMMENT '告警用户';


CREATE TABLE IF NOT EXISTS `t_automation_dag_instance_ext_data` (
  `f_id` VARCHAR(64) NOT NULL COMMENT '主键id',
  `f_created_at` BIGINT DEFAULT NULL COMMENT '创建时间',
  `f_updated_at` BIGINT DEFAULT NULL COMMENT '更新时间',
  `f_dag_id` VARCHAR(64) COMMENT 'DAG id',
  `f_dag_ins_id` VARCHAR(64) COMMENT 'DAG实例id',
  `f_field` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '字段名称',
  `f_oss_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'OSS存储id',
  `f_oss_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'OSS存储key',
  `f_size` BIGINT unsigned DEFAULT NULL COMMENT '文件大小',
  `f_removed` BOOLEAN NOT NULL DEFAULT 1 COMMENT '是否删除(1:未删除,0:已删除)',
  PRIMARY KEY (`f_id`),
  KEY `idx_t_automation_dag_instance_ext_data_dag_ins_id` (`f_dag_ins_id`)
) ENGINE=InnoDB COMMENT 'DagInstanceExtData';


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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  UNIQUE KEY `uk_hash` (`f_hash`),
  KEY `idx_expire_time` (`f_expire_time`)
) ENGINE=InnoDB COMMENT 'ContentPipeline 任务';

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
  `f_inline` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否内联',
  `f_visibility` TINYINT(2) NOT NULL DEFAULT '0' COMMENT '可见性(0: private, 1: public)',
  `f_timestamp` BIGINT(20) NOT NULL DEFAULT '0' COMMENT '时间戳',
  PRIMARY KEY (`f_id`),
  KEY `idx_instance_id` (`f_instance_id`, `f_id`),
  KEY `idx_instance_type_vis` (`f_instance_id`, `f_type`, `f_visibility`, `f_id`),
  KEY `idx_instance_name_type` (`f_instance_id`, `f_name`, `f_type`, `f_id`)
) ENGINE=InnoDB COMMENT='DAG实例事件日志表';

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
 `f_removed` BOOLEAN NOT NULL DEFAULT 0 COMMENT '删除标记',
 `f_emails` TEXT DEFAULT NULL COMMENT '通知邮箱',
 `f_template` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '模板标识',
 `f_published` BOOLEAN NOT NULL DEFAULT 0 COMMENT '发布标记',
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
 `f_is_debug` BOOLEAN NOT NULL DEFAULT 0 COMMENT '调试标记',
 `f_debug_id` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '调试ID',
 `f_biz_domain_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务域ID',
  PRIMARY KEY (`f_id`),
  KEY `idx_dag_user_id` (`f_user_id`),
  KEY `idx_dag_type` (`f_type`),
  KEY `idx_dag_trigger` (`f_trigger`),
  KEY `idx_dag_name` (`f_name`),
  KEY `idx_dag_biz_domain` (`f_biz_domain_id`)
) ENGINE=InnoDB COMMENT 'DAG流程定义表';

CREATE TABLE IF NOT EXISTS `t_flow_dag_var` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_dag_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'DAG ID',
 `f_var_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变量名',
 `f_default_value` TEXT DEFAULT NULL COMMENT '默认值',
 `f_var_type` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '变量类型',
 `f_description` TEXT DEFAULT NULL COMMENT '变量描述',
  PRIMARY KEY (`f_id`),
  KEY `idx_dag_vars_dag_id` (`f_dag_id`)
) ENGINE=InnoDB COMMENT 'DAG变量定义表';

CREATE TABLE IF NOT EXISTS `t_flow_dag_instance_keyword` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_dag_ins_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'DAG实例ID',
 `f_keyword` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关键词',
  PRIMARY KEY (`f_id`),
  KEY `idx_dag_ins_kw` (`f_dag_ins_id`, `f_keyword`)
) ENGINE=InnoDB COMMENT 'DAG实例关键词表';

CREATE TABLE IF NOT EXISTS `t_flow_dag_step` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_dag_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'DAG ID',
 `f_operator` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作符',
 `f_source_id` TEXT NOT NULL COMMENT '来源ID',
 `f_has_datasource` BOOLEAN NOT NULL DEFAULT 0 COMMENT '是否有数据源',
  PRIMARY KEY (`f_id`),
  KEY `idx_dag_step_op` (`f_operator`),
  KEY `idx_dag_step_op_dag` (`f_dag_id`, `f_operator`),
  KEY `idx_dag_step_has_ds_dag` (`f_dag_id`, `f_has_datasource`)
) ENGINE=InnoDB COMMENT 'DAG步骤定义表';

CREATE TABLE IF NOT EXISTS `t_flow_dag_accessor` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_dag_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'DAG ID',
 `f_accessor_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '访问者ID',
  PRIMARY KEY (`f_id`),
  KEY `idx_dag_accessor_id_dag` (`f_accessor_id`, `f_dag_id`)
) ENGINE=InnoDB COMMENT 'DAG访问者定义表';

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
  PRIMARY KEY (`f_id`),
  KEY `idx_dag_versions_dag_version` (`f_version_id`, `f_dag_id`),
  KEY `idx_dag_versions_dag_sort` (`f_dag_id`, `f_sort_time`)
) ENGINE=InnoDB COMMENT 'DAG版本定义表';

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
 `f_event_persistence` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '事件持久化标记',
 `f_event_oss_path` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '事件OSS路径',
 `f_share_data` MEDIUMTEXT DEFAULT NULL COMMENT '共享数据',
 `f_share_data_ext` MEDIUMTEXT DEFAULT NULL COMMENT '共享数据扩展',
 `f_status` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '状态',
 `f_reason` MEDIUMTEXT DEFAULT NULL COMMENT '原因',
 `f_cmd` TEXT DEFAULT NULL COMMENT '命令',
 `f_has_cmd` BOOLEAN NOT NULL DEFAULT 0 COMMENT '是否包含命令',
 `f_batch_run_id` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '批次运行ID',
 `f_user_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '用户ID',
 `f_ended_at` BIGINT NOT NULL DEFAULT 0 COMMENT '结束时间',
 `f_dag_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'DAG类型',
 `f_policy_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '策略类型',
 `f_appinfo` TEXT DEFAULT NULL COMMENT '应用信息',
 `f_priority` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '优先级',
 `f_mode` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '模式',
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
  PRIMARY KEY (`f_id`),
  KEY `idx_dag_ins_dag_status` (`f_dag_id`, `f_status`),
  KEY `idx_dag_ins_status_upd` (`f_status`, `f_updated_at`),
  KEY `idx_dag_ins_status_user_pri` (`f_status`, `f_user_id`, `f_priority`),
  KEY `idx_dag_ins_user_id` (`f_user_id`),
  KEY `idx_dag_ins_batch_run` (`f_batch_run_id`),
  KEY `idx_dag_ins_worker` (`f_worker`)
) ENGINE=InnoDB COMMENT 'DAG实例定义表';

CREATE TABLE IF NOT EXISTS `t_flow_inbox` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_msg` MEDIUMTEXT DEFAULT NULL COMMENT '消息内容',
 `f_topic` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '主题',
 `f_docid` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文档ID',
 `f_dag` TEXT DEFAULT NULL COMMENT 'DAG列表',
  PRIMARY KEY (`f_id`),
  KEY `idx_inbox_docid` (`f_docid`),
  KEY `idx_inbox_topic_created` (`f_topic`, `f_created_at`)
) ENGINE=InnoDB COMMENT '事件触发流程记录表';

CREATE TABLE IF NOT EXISTS `t_flow_outbox` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_msg` MEDIUMTEXT DEFAULT NULL COMMENT '消息内容',
 `f_topic` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '主题',
  PRIMARY KEY (`f_id`),
  KEY `idx_outbox_created` (`f_created_at`)
) ENGINE=InnoDB COMMENT '消息发件箱';

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
  PRIMARY KEY (`f_id`),
  KEY `idx_task_ins_dag_ins_id` (`f_dag_ins_id`),
  KEY `idx_task_ins_hash` (`f_hash`),
  KEY `idx_task_ins_action` (`f_action_name`),
  KEY `idx_task_ins_status_expire` (`f_status`, `f_expired_at`),
  KEY `idx_task_ins_status_upd_id` (`f_status`, `f_updated_at`, `f_id`)
) ENGINE=InnoDB COMMENT 'Task实例定义表';

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
 `f_is_app` BOOLEAN NOT NULL DEFAULT 0 COMMENT '是否应用',
  PRIMARY KEY (`f_id`),
  KEY `idx_token_user_id` (`f_user_id`)
) ENGINE=InnoDB COMMENT 'Token定义表';

CREATE TABLE IF NOT EXISTS `t_flow_client` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_client_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端名称',
 `f_client_id` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '客户端ID',
 `f_client_secret` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '客户端密钥',
  PRIMARY KEY (`f_id`),
  KEY `idx_client_name` (`f_client_name`)
) ENGINE=InnoDB COMMENT 'Client定义表';

CREATE TABLE IF NOT EXISTS `t_flow_switch` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '开关名称',
 `f_status` BOOLEAN NOT NULL DEFAULT 0 COMMENT '开关状态',
  PRIMARY KEY (`f_id`),
  KEY `idx_switch_name` (`f_name`)
) ENGINE=InnoDB COMMENT '开关定义表';

CREATE TABLE IF NOT EXISTS `t_flow_log` (
 `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
 `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
 `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
 `f_ossid` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'OSS ID',
 `f_key` VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'OSS Key',
 `f_filename` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '文件名',
  PRIMARY KEY (`f_id`)
) ENGINE=InnoDB COMMENT '日志定义表';

-- Source: autoflow/flow-stream-data-pipeline/migrations/mariadb/0.1.0/pre/init.sql
USE adp;

-- 内部应用
CREATE TABLE IF NOT EXISTS t_internal_app (
  f_app_id varchar(40) NOT NULL COMMENT 'app_id',
  f_app_name varchar(40) NOT NULL COMMENT 'app名称',
  f_app_secret varchar(40) NOT NULL COMMENT 'app_secret',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (f_app_id),
  UNIQUE KEY uk_app_name (f_app_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '内部应用';


CREATE TABLE IF NOT EXISTS t_stream_data_pipeline (
  f_pipeline_id varchar(40) NOT NULL DEFAULT '' COMMENT '管道 id',
  f_pipeline_name varchar(40) NOT NULL DEFAULT '' COMMENT '管道名称',
  f_tags varchar(255) NOT NULL COMMENT '标签',
  f_comment varchar(255) COMMENT '备注',
  f_builtin boolean DEFAULT 0 COMMENT '内置管道标识: 0 非内置, 1 内置',
  f_output_type varchar(20) NOT NULL COMMENT '数据输出类型',
  f_index_base varchar(255) NOT NULL COMMENT '索引库类型',
  f_use_index_base_in_data boolean DEFAULT 0 COMMENT '是否使用数据里的索引库: 0 否, 1 是',
  f_pipeline_status varchar(10) NOT NULL COMMENT '管道状态: failed 失败, running 运行中, close 关闭',
  f_pipeline_status_details text NOT NULL COMMENT '管道状态详情',
  f_deployment_config text NOT NULL COMMENT '资源配置信息',
  f_create_time bigint(20) NOT NULL default 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL default 0 COMMENT '更新时间',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_updater varchar(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type varchar(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  PRIMARY KEY (f_pipeline_id),
  UNIQUE KEY uk_name (f_pipeline_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '流式数据管道信息';
-- Source: autoflow/workflow/migrations/mariadb/0.1.0/pre/init.sql
/*
 Navicat Premium Data Transfer

 Source Server         : 192.168.1.170-workflow
 Source Server Type    : MySQL
 Source Server Version : 100604
 Source Host           : 192.168.1.170:3306
 Source Schema         : workflow

 Target Server Type    : MySQL
 Target Server Version : 100604
 File Encoding         : 65001

 Date: 12/04/2023 20:51:08
*/

USE adp;
-- SET NAMES utf8mb4;
-- SET FOREIGN_KEY_CHECKS = 0;


-- ----------------------------
-- Table structure for t_wf_activity_info_config
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_activity_info_config`  (
  `activity_def_id` varchar(100) NOT NULL COMMENT '流程环节定义ID',
  `activity_def_name` varchar(100) NULL DEFAULT NULL COMMENT '流程环节定义名称',
  `process_def_id` varchar(100) NOT NULL COMMENT '流程定义ID',
  `process_def_name` varchar(500) NULL DEFAULT NULL COMMENT '流程定义名称',
  `activity_page_url` varchar(500) NULL DEFAULT NULL COMMENT '流程环节表单URL',
  `activity_page_info` mediumtext NULL COMMENT '流程环节表单数据',
  `activity_operation_roleid` varchar(4000) NULL DEFAULT NULL COMMENT '流程环节绑定操作权限ID',
  `remark` varchar(500) NULL DEFAULT NULL COMMENT '备注',
  `jump_type` varchar(10) NULL DEFAULT NULL COMMENT '环节跳转类型，AUTO：自动路径跳转；MANUAL：人工选择跳转、FREE：自由选择跳转',
  `activity_status_name` varchar(100) NULL DEFAULT NULL COMMENT '流程环节状态名称(默认与环节名称一致)',
  `activity_order` decimal(10, 0) NULL DEFAULT NULL COMMENT '环节排序',
  `activity_limit_time` decimal(10, 0) NULL DEFAULT NULL COMMENT '环节时限',
  `idea_display_area` varchar(50) NULL DEFAULT NULL COMMENT '意见分栏',
  `is_show_idea` varchar(10) NULL DEFAULT NULL COMMENT '是否显示意见输入区域,默认启用ENABLED,否则禁用DISABLE',
  `activity_def_child_type` varchar(20) NULL DEFAULT NULL COMMENT '环节子类型，through:流程贯穿,inside:内部流程',
  `activity_def_deal_type` varchar(20) NULL DEFAULT NULL COMMENT '环节处理类型，单人多人',
  `activity_def_type` varchar(20) NULL DEFAULT NULL COMMENT '环节类型',
  `is_start_usertask` varchar(4) NULL DEFAULT NULL COMMENT '是否是开始节点  是为Y  否为N',
  `c_protocl` varchar(50) NULL DEFAULT NULL COMMENT 'PC端协议',
  `m_protocl` varchar(50) NULL DEFAULT NULL COMMENT '移动端协议',
  `m_url` varchar(500) NULL DEFAULT NULL COMMENT '手机端待办地址',
  `other_sys_deal_status` varchar(10) NULL DEFAULT NULL COMMENT '其它系统处理状态   0 不可处理；1 仅阅读；2可处理',
  PRIMARY KEY (`activity_def_id`, `process_def_id`) USING BTREE
) ENGINE = InnoDB;

-- -- ----------------------------
-- -- Records of t_wf_activity_info_config
-- -- ----------------------------
-- INSERT INTO `t_wf_activity_info_config` VALUES ('UserTask_0zz6lcw', '审核', 'Process_SHARE001:1:52af48f0-d930-11ed-8086-00ff0fa3e6a7', '实名共享审核工作流', NULL, NULL, NULL, NULL, 'MANUAL', NULL, 1, NULL, NULL, 'ENABLED', NULL, 'tjsh', 'userTask', NULL, NULL, NULL, NULL, NULL);
-- INSERT INTO `t_wf_activity_info_config` VALUES ('UserTask_0zz6lcw', '审核', 'Process_SHARE002:1:52e489c3-d930-11ed-8086-00ff0fa3e6a7', '匿名共享审核工作流', NULL, NULL, NULL, NULL, 'MANUAL', NULL, 1, NULL, NULL, 'ENABLED', NULL, 'tjsh', 'userTask', NULL, NULL, NULL, NULL, NULL);

-- ----------------------------
-- Table structure for t_wf_activity_rule
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_activity_rule`  (
  `rule_id` varchar(50) NOT NULL COMMENT '规则ID',
  `rule_name` varchar(250) NOT NULL COMMENT '规则名称',
  `proc_def_id` varchar(100) NULL DEFAULT NULL COMMENT '流程定义ID',
  `source_act_id` varchar(50) NULL DEFAULT NULL COMMENT '源环节ID',
  `target_act_id` varchar(50) NULL DEFAULT NULL COMMENT '目标环节ID',
  `rule_script` mediumtext NOT NULL COMMENT '规则脚本',
  `rule_priority` decimal(10, 0) NULL DEFAULT NULL COMMENT '优先级',
  `rule_type` varchar(5) NULL DEFAULT NULL COMMENT '规则类型：A：环节，R：资源，F:多实例环节完成条件',
  `tenant_id` varchar(255) NOT NULL COMMENT '租户ID',
  `rule_remark` varchar(2000) NULL DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`rule_id`) USING BTREE
) ENGINE = InnoDB COMMENT = '环节规则表';

-- ----------------------------
-- Records of t_wf_activity_rule
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_application
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_application`  (
  `app_id` varchar(50) NOT NULL COMMENT '应用系统ID',
  `app_name` varchar(50) NULL DEFAULT NULL COMMENT '应用系统名称',
  `app_type` varchar(20) NULL DEFAULT NULL COMMENT '应用系统分类',
  `app_access_url` varchar(300) NULL DEFAULT NULL COMMENT '应用系统访问地址',
  `app_create_time` datetime(0) NULL DEFAULT NULL COMMENT '应用系统创建时间',
  `app_update_time` datetime(0) NULL DEFAULT NULL COMMENT '应用系统更新时间',
  `app_creator_id` varchar(50) NULL DEFAULT NULL COMMENT '应用系统创建人ID',
  `app_updator_id` varchar(50) NULL DEFAULT NULL COMMENT '应用系统更新人ID',
  `app_status` varchar(2) NULL DEFAULT NULL COMMENT '应用系统状态',
  `app_desc` varchar(300) NULL DEFAULT NULL COMMENT '应用系统描述',
  `app_provider` varchar(100) NULL DEFAULT NULL COMMENT '应用系统开发厂商',
  `app_linkman` varchar(50) NULL DEFAULT NULL COMMENT '应用系统联系人',
  `app_phone` varchar(30) NULL DEFAULT NULL COMMENT '应用系统联系电话',
  `app_unitework_check_url` varchar(300) NULL DEFAULT NULL COMMENT '应用系统检查路径',
  `app_sort` decimal(10, 0) NULL DEFAULT NULL COMMENT '应用系统排序号',
  `app_shortname` char(50) NULL DEFAULT NULL COMMENT '应用系统简洁名称',
  PRIMARY KEY (`app_id`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_application
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_application_user
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_application_user`  (
  `app_id` varchar(50) NOT NULL COMMENT '租户ID',
  `user_id` varchar(50) NOT NULL COMMENT '用户ID',
  `remark` varchar(300) NULL DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`app_id`, `user_id`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_application_user
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_dict
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_dict`  (
  `id` varchar(50) NOT NULL COMMENT '主键id',
  `dict_code` varchar(50) NULL DEFAULT NULL COMMENT '字典编码',
  `dict_parent_id` varchar(50) NULL DEFAULT NULL COMMENT '字典上级主键id',
  `dict_name` text NULL COMMENT '字典名称',
  `sort` decimal(10, 0) NULL DEFAULT NULL COMMENT '排序号',
  `status` varchar(2) NULL DEFAULT NULL COMMENT '状态',
  `creator_id` varchar(50) NULL DEFAULT NULL COMMENT '创建人',
  `create_date` datetime(0) NULL DEFAULT NULL COMMENT '创建时间',
  `updator_id` varchar(50) NULL DEFAULT NULL COMMENT '更新人',
  `update_date` datetime(0) NULL DEFAULT NULL COMMENT '最后更新时间',
  `app_id` varchar(50) NOT NULL COMMENT '应用id',
  `dict_value` varchar(4000) NULL DEFAULT NULL COMMENT '字典值',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB;

-- -- ----------------------------
-- -- Records of t_wf_dict
-- -- ----------------------------
-- INSERT INTO `t_wf_dict` VALUES ('492b9f23-d930-11ed-81d7-20040ff2c754', 'free_audit_secret_level', NULL, '6', NULL, 'Y', NULL, NULL, NULL, NULL, 'as_workflow', '');
-- INSERT INTO `t_wf_dict` VALUES ('492d9405-d930-11ed-81d7-20040ff2c754', 'self_dept_free_audit', NULL, 'Y', NULL, 'Y', NULL, NULL, NULL, NULL, 'as_workflow', '');
-- INSERT INTO `t_wf_dict` VALUES ('492f6745-d930-11ed-81d7-20040ff2c754', 'free_audit_secret_level_enum', NULL, '{\"非密\": 5,\"内部\": 6, \"秘密\": 7,\"机密\": 8}', NULL, 'Y', NULL, NULL, NULL, NULL, 'as_workflow', '');
-- INSERT INTO `t_wf_dict` VALUES ('4931465a-d930-11ed-81d7-20040ff2c754', 'anonymity_auto_audit_switch', NULL, 'n', NULL, 'Y', NULL, NULL, NULL, NULL, 'as_workflow', NULL);
-- INSERT INTO `t_wf_dict` VALUES ('49331202-d930-11ed-81d7-20040ff2c754', 'rename_auto_audit_switch', NULL, 'n', NULL, 'Y', NULL, NULL, NULL, NULL, 'as_workflow', NULL);

-- ----------------------------
-- Table structure for t_wf_doc_audit_apply
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_doc_audit_apply`  (
  `id` varchar(50) NOT NULL COMMENT '申请ID',
  `biz_id` varchar(50) NOT NULL COMMENT '业务关联ID，如：AS共享申请ID',
  `doc_id` text NULL COMMENT '文档ID，如：gns://xxx/xxx',
  `doc_path` varchar(1000) NULL DEFAULT NULL COMMENT '文档路径，如：/name/xxx.txt',
  `doc_type` varchar(10) NULL DEFAULT NULL COMMENT '文档类型（folder文件夹,file文件,doc_lib文档库）',
  `csf_level` int(2) NULL DEFAULT NULL COMMENT '文件密级,5~15，如果是文件夹，则为0',
  `biz_type` varchar(100) NULL DEFAULT NULL COMMENT '业务类型（realname共享给指定用户的申请，anonymous共享给任意用户的申请，sync同步申请，flow流转申请，security定密申请）',
  `apply_type` varchar(100) NOT NULL COMMENT '申请类型（sync同步申请，flow流转申请，perm共享申请，anonymous匿名申请，owner所有者申请，security定密申请，inherit更改继承申请）',
  `apply_detail` mediumtext NOT NULL COMMENT '申请明细（docLibType文档库ID，accessorId访问者ID，accessorName访问者名称，accessorType访问者类型，allowValue允许权限，denyValue拒绝权限，inherit是否继承权限，expiresAt有效期，opType操作类型，linkUrl链接地址，title链接标题，password密码，accessLimit访问次数）',
  `proc_def_id` varchar(100) NULL DEFAULT NULL COMMENT '流程定义ID',
  `proc_def_name` varchar(300) NULL DEFAULT NULL COMMENT '流程定义名称',
  `proc_inst_id` varchar(100) NULL DEFAULT NULL COMMENT '流程实例ID',
  `audit_type` varchar(10) NULL DEFAULT NULL COMMENT '审核模式（tjsh-同级审核，hqsh-汇签审核，zjsh-逐级审核）',
  `auditor` text NULL COMMENT '审核员，冗余字段，用于页面展示(id审核人ID，name审核人名称，status审核状态，auditDate审核时间)',
  `apply_user_id` varchar(50) NOT NULL COMMENT '申请人ID',
  `apply_user_name` varchar(150) NULL DEFAULT NULL COMMENT '申请人名称',
  `apply_time` datetime(0) NOT NULL COMMENT '申请时间',
  `doc_names` varchar(2000) NULL DEFAULT NULL COMMENT '文档名称',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `AK_t_wf_doc_audit_application_proc_inst_idx` (`proc_inst_id`) USING BTREE,
  KEY `AK_t_wf_doc_audit_application_apply_user_idx` (`apply_user_id`) USING BTREE,
  KEY `AK_t_wf_doc_audit_application_biz_idx` (`biz_id`) USING BTREE,
  KEY `AK_t_wf_doc_audit_application_biz_typex` (`biz_type`) USING BTREE
) ENGINE = InnoDB COMMENT = '文档审核申请表';

-- ----------------------------
-- Records of t_wf_doc_audit_apply
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_doc_audit_detail
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_doc_audit_detail`  (
  `id` varchar(50) NOT NULL COMMENT '主键ID',
  `apply_id` varchar(50) NOT NULL COMMENT '申请ID',
  `doc_id` text NOT NULL COMMENT '文档ID，如：gns://xxx/xxx',
  `doc_path` varchar(1000) NOT NULL COMMENT '文档路径，如：/name/xxx.txt',
  `doc_type` varchar(10) NOT NULL COMMENT '文档类型（folder文件夹,file文件,doc_lib文档库）',
  `csf_level` int(2) NULL DEFAULT NULL COMMENT '文件密级,5~15，如果是文件夹，则为0',
  `doc_name` varchar(1000) NULL DEFAULT NULL COMMENT '文件名称',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `AK_t_wf_doc_audit_detail_apply_idx` (`apply_id`) USING BTREE
) ENGINE = InnoDB COMMENT = '文档审核申请明细表';

-- ----------------------------
-- Records of t_wf_doc_audit_detail
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_doc_audit_history
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_doc_audit_history`  (
  `id` varchar(50) NOT NULL COMMENT '申请ID',
  `biz_id` varchar(50) NOT NULL COMMENT '业务关联ID，如：AS共享申请ID',
  `doc_id` text NULL COMMENT '文档ID，如：gns://xxx/xxx',
  `doc_path` varchar(1000) NULL DEFAULT NULL COMMENT '文档路径，如：/name/xxx.txt',
  `doc_type` varchar(10) NULL DEFAULT NULL COMMENT '文档类型（folder文件夹,file文件,doc_lib文档库）',
  `csf_level` int(2) NULL DEFAULT NULL COMMENT '文件密级,5~15，如果是文件夹，则为0',
  `biz_type` varchar(100) NULL DEFAULT NULL COMMENT '业务类型（realname共享给指定用户的申请，anonymous共享给任意用户的申请，sync同步申请，flow流转申请，security定密申请）',
  `apply_type` varchar(100) NOT NULL COMMENT '申请类型（sync同步申请，flow流转申请，perm共享申请，anonymous匿名申请，owner所有者申请，security更改密级申请，inherit更改继承申请）',
  `apply_detail` mediumtext NOT NULL COMMENT '申请明细（docLibType文档库ID，accessorId访问者ID，accessorName访问者名称，accessorType访问者类型，allowValue允许权限，denyValue拒绝权限，inherit是否继承权限，expiresAt有效期，opType操作类型，linkUrl链接地址，title链接标题，password密码，accessLimit访问次数）',
  `proc_def_id` varchar(100) NULL DEFAULT NULL COMMENT '流程定义ID',
  `proc_def_name` varchar(300) NULL DEFAULT NULL COMMENT '流程定义名称',
  `proc_inst_id` varchar(100) NULL DEFAULT NULL COMMENT '流程实例ID',
  `apply_user_id` varchar(50) NOT NULL COMMENT '申请人ID',
  `apply_user_name` varchar(150) NULL DEFAULT NULL COMMENT '申请人名称',
  `apply_time` datetime(0) NOT NULL COMMENT '申请时间',
  `audit_status` int(10) NOT NULL COMMENT '审核状态，1-审核中 2-已拒绝 3-已通过 4-自动审核通过 5-作废  6-发起失败',
  `audit_result` varchar(10) NULL DEFAULT NULL COMMENT '审核结果，pass-通过 reject-拒绝',
  `audit_msg` varchar(2400) NULL DEFAULT NULL COMMENT '最后一次审核意见，默认为审核意见，当审核状态为6时代表发起失败异常码，异常码包含：S0001未配置审核策略;S0002无匹配审核员;S0003无匹配的审核员（发起人与审核人为同一人);S0004无匹配密级审核员',
  `audit_type` varchar(10) NULL DEFAULT NULL COMMENT '审核模式（tjsh-同级审核，hqsh-汇签审核，zjsh-逐级审核）',
  `auditor` text NULL COMMENT '审核员，冗余字段，用于页面展示(id审核人ID，name审核人名称，status审核状态，auditDate审核时间)',
  `last_update_time` datetime(0) NULL DEFAULT NULL COMMENT '最后一次修改时间',
  `doc_names` varchar(2000) NULL DEFAULT NULL COMMENT '文档名称',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `AK_t_wf_doc_audit_history_proc_inst_idx` (`proc_inst_id`) USING BTREE,
  KEY `AK_t_wf_doc_audit_history_biz_idx` (`biz_id`) USING BTREE,
  KEY `AK_t_wf_doc_audit_history_apply_user_id_idx` (`apply_user_id`, `audit_status`, `biz_type`, `last_update_time`) USING BTREE
) ENGINE = InnoDB COMMENT = '文档审核历史表';

-- ----------------------------
-- Records of t_wf_doc_audit_history
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_doc_share_strategy
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_doc_share_strategy`  (
  `id` varchar(50) NOT NULL COMMENT '主键',
  `doc_id` varchar(200) NULL DEFAULT NULL COMMENT '文档库ID',
  `doc_name` varchar(300) NULL DEFAULT NULL COMMENT '文档库名称',
  `doc_type` varchar(100) NULL DEFAULT NULL COMMENT '文档库类型，个人文档库：user_doc_lib；部门文档库：department_doc_lib；自定义文档库：custom_doc_lib; 知识库: knowledge_doc_lib',
  `audit_model` varchar(100) NOT NULL COMMENT '审核模式，同级审核：tjsh；汇签审核：hqsh；逐级审核：zjsh；',
  `proc_def_id` varchar(300) NOT NULL COMMENT '流程定义ID',
  `proc_def_name` varchar(300) NOT NULL COMMENT '流程定义名称',
  `act_def_id` varchar(100) NULL DEFAULT NULL COMMENT '流程环节ID',
  `act_def_name` varchar(300) NULL DEFAULT NULL COMMENT '流程环节名称',
  `create_user_id` varchar(100) NULL DEFAULT NULL COMMENT '创建人ID',
  `create_user_name` varchar(100) NULL DEFAULT NULL COMMENT '创建人名称',
  `create_time` datetime(0) NOT NULL COMMENT '创建时间',
  `strategy_type` varchar(50) NULL DEFAULT NULL COMMENT '策略类型，指定用户审核：named_auditor；部门审核员：dept_auditor；连续多级部门审核：multilevel',
  `rule_type` varchar(50) NULL DEFAULT NULL COMMENT '规则类型，角色：role',
  `rule_id` varchar(100) NULL DEFAULT NULL COMMENT '规则ID',
  `level_type` varchar(100) NULL DEFAULT NULL COMMENT '匹配级别类型，直属部门向上一级：belongUp1；直属部门向上二级：belongUp2；直属部门向上三级：belongUp3；直属部门向上四级：belongUp4；直属部门向上五级：belongUp5；直属部门向上六级：belongUp6；直属部门向上七级：belongUp7；直属部门向上八级：belongUp8；直属部门向上九级：belongUp9；直属部门向上十级：belongUp10；最高级部门审核员：highestLevel；最高级部门向下一级：highestDown1；最高级部门向下二级：highestDown2；最高级部门向下三级：highestDown3；最高级部门向下四级：highestDown4；最高级部门向下五级：highestDown5；最高级部门向下六级：highestDown6；最高级部门向下七级：highestDown7；最高级部门向下八级：highestDown8；最高级部门向下九级：highestDown9；最高级部门向下十级：highestDown10；',
  `no_auditor_type` varchar(50) NULL DEFAULT NULL COMMENT '未匹配到部门审核员类型，自动拒绝：auto_reject；自动通过：auto_pass',
  `repeat_audit_type` varchar(50) NULL DEFAULT NULL COMMENT '同一审核员重复审核类型，只需审核一次：once；每次都需要审核：always',
  `own_auditor_type` varchar(50) NULL DEFAULT NULL COMMENT '审核员为发起人自己时审核类型，自动拒绝：auto_reject；自动通过：auto_pass',
  `countersign_switch` varchar(10) NULL DEFAULT NULL COMMENT '是否允许加签 Y-是',
  `countersign_count` varchar(10) NULL DEFAULT NULL COMMENT '允许最大加签次数',
  `countersign_auditors` varchar(10) NULL DEFAULT NULL COMMENT '允许最大加签人数',
  `transfer_switch` varchar(10) NULL DEFAULT NULL COMMENT '转审开关',
  `transfer_count` varchar(10) NULL DEFAULT NULL COMMENT '最大转审次数',
  `perm_config` varchar(64) NOT NULL DEFAULT '' COMMENT '申请人权限配置',
  `strategy_configs` varchar(128) NOT NULL DEFAULT '' COMMENT '新增高级配置统一存放位置',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `AK_T_WF_DOC_SHARE_STRATEGY_PROC_DEF_ID_IDX` (`proc_def_id`) USING BTREE,
  KEY `AK_T_WF_DOC_SHARE_STRATEGY_ACT_DEF_ID_IDX` (`act_def_id`) USING BTREE
) ENGINE = InnoDB COMMENT = '文档共享审核策略';

-- ----------------------------
-- Records of t_wf_doc_share_strategy
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_doc_share_strategy_auditor
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_doc_share_strategy_auditor`  (
  `id` varchar(50) NOT NULL COMMENT '主键',
  `user_id` varchar(300) NOT NULL COMMENT '审核人',
  `user_code` varchar(300) NOT NULL COMMENT '审核人账号',
  `user_name` varchar(300) NOT NULL COMMENT '审核人名称',
  `user_dept_id` varchar(300) NULL DEFAULT NULL COMMENT '审核人部门ID',
  `user_dept_name` varchar(300) NULL DEFAULT NULL COMMENT '审核人部门名称',
  `audit_strategy_id` varchar(100) NULL DEFAULT NULL COMMENT '审核策略ID（t_wf_doc_audit_strategy主键）',
  `audit_sort` int(11) NULL DEFAULT NULL COMMENT '审核人排序',
  `create_user_id` varchar(100) NULL DEFAULT NULL COMMENT '创建人ID',
  `create_user_name` varchar(100) NULL DEFAULT NULL COMMENT '创建人名称',
  `create_time` datetime(0) NOT NULL COMMENT '创建时间',
  `org_type` varchar(32) NOT NULL DEFAULT '' COMMENT '组织类型',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `AK_T_WF_DOC_SHARE_STRATEGY_AUDITOR_STRATEGY_ID_IDX` (`audit_strategy_id`) USING BTREE,
  KEY `AK_T_WF_DOC_SHARE_STRATEGY_AUDITOR_USER_ID_IDX` (`user_id`) USING BTREE
) ENGINE = InnoDB COMMENT = '文档共享审核员表';

-- ----------------------------
-- Records of t_wf_doc_share_strategy_auditor
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_evt_log
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_evt_log`  (
  `log_nr_` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `type_` varchar(64) NULL DEFAULT NULL COMMENT '类型',
  `proc_def_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程定义ID',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `execution_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程执行ID',
  `task_id_` varchar(64) NULL DEFAULT NULL COMMENT '任务ID',
  `time_stamp_` timestamp(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '日志时间',
  `user_id_` varchar(255) NULL DEFAULT NULL COMMENT '用户ID',
  `data_` longblob NULL COMMENT '内容',
  `lock_owner_` varchar(255) NULL DEFAULT NULL,
  `lock_time_` timestamp(0) NULL DEFAULT NULL,
  `is_processed_` tinyint(4) NULL DEFAULT 0,
  PRIMARY KEY (`log_nr_`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1;

-- ----------------------------
-- Records of t_wf_evt_log
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_free_audit
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_free_audit`  (
  `id` varchar(50) NOT NULL COMMENT '主键ID',
  `process_def_key` varchar(30) NULL DEFAULT NULL COMMENT '流程定义key',
  `department_id` varchar(50) NOT NULL COMMENT '部门id',
  `department_name` varchar(600) NOT NULL COMMENT '部门名称',
  `create_user_id` varchar(50) NOT NULL COMMENT '创建人id',
  `create_time` datetime(0) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `AK_t_wf_free_audit_process_def_keyx` (`process_def_key`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_free_audit
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_ge_bytearray
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_ge_bytearray`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '部署的文件名称',
  `deployment_id_` varchar(64) NULL DEFAULT NULL COMMENT '部署的ID',
  `bytes_` longblob NULL COMMENT '大文本类型，存储文本字节流',
  `generated_` tinyint(4) NULL DEFAULT NULL COMMENT '是否是引擎生成 0为用户生成 1为Activiti生成',
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_FK_BYTEARR_DEPL` (`deployment_id_`) USING BTREE
) ENGINE = InnoDB COMMENT = '内核-流程数据存储表（流程定义xml、流程执行变量序列化数据）';

-- -- ----------------------------
-- -- Records of t_wf_ge_bytearray
-- -- ----------------------------
-- INSERT INTO `t_wf_ge_bytearray` VALUES ('52614fff-d930-11ed-8086-00ff0fa3e6a7', 1, '实名共享审核工作流.bpmn20.xml', '526128ee-d930-11ed-8086-00ff0fa3e6a7', 0x3C3F786D6C2076657273696F6E3D22312E302220656E636F64696E673D225554462D38223F3E0A3C646566696E6974696F6E7320786D6C6E733D22687474703A2F2F7777772E6F6D672E6F72672F737065632F42504D4E2F32303130303532342F4D4F44454C2220786D6C6E733A7873693D22687474703A2F2F7777772E77332E6F72672F323030312F584D4C536368656D612D696E7374616E63652220786D6C6E733A62706D6E64693D22687474703A2F2F7777772E6F6D672E6F72672F737065632F42504D4E2F32303130303532342F44492220786D6C6E733A6F6D6764633D22687474703A2F2F7777772E6F6D672E6F72672F737065632F44442F32303130303532342F44432220786D6C6E733A64693D22687474703A2F2F7777772E6F6D672E6F72672F737065632F44442F32303130303532342F44492220786D6C6E733A61637469766974693D22687474703A2F2F61637469766974692E6F72672F62706D6E2220786D6C6E733A7873643D22687474703A2F2F7777772E77332E6F72672F323030312F584D4C536368656D6122207461726765744E616D6573706163653D22687474703A2F2F7777772E61637469766974692E6F72672F74657374223E0A20203C70726F636573732069643D2250726F636573735F534841524530303122206E616D653D22E5AE9EE5908DE585B1E4BAABE5AEA1E6A0B8E5B7A5E4BD9CE6B5812220697345786563757461626C653D2274727565223E0A202020203C73746172744576656E742069643D227369642D34363538384541412D333842372D344642432D383044442D34364135454645323643464122206E616D653D22E58F91E8B5B7223E0A2020202020203C6F7574676F696E673E53657175656E6365466C6F775F306A66656E64773C2F6F7574676F696E673E0A202020203C2F73746172744576656E743E0A202020203C757365725461736B2069643D22557365725461736B5F307A7A366C637722206E616D653D22E5AEA1E6A0B8222061637469766974693A61737369676E65653D22247B61737369676E65657D222061637469766974693A63616E64696461746555736572733D22223E0A2020202020203C657874656E73696F6E456C656D656E74733E0A20202020202020203C61637469766974693A657870616E6450726F70657274792069643D226465616C54797065222076616C75653D22746A736822202F3E0A2020202020203C2F657874656E73696F6E456C656D656E74733E0A2020202020203C696E636F6D696E673E53657175656E6365466C6F775F306A66656E64773C2F696E636F6D696E673E0A2020202020203C6F7574676F696E673E53657175656E6365466C6F775F303871637962793C2F6F7574676F696E673E0A2020202020203C6D756C7469496E7374616E63654C6F6F7043686172616374657269737469637320697353657175656E7469616C3D2266616C7365222061637469766974693A636F6C6C656374696F6E3D22247B61737369676E65654C6973747D222061637469766974693A656C656D656E745661726961626C653D2261737369676E656522202F3E0A202020203C2F757365725461736B3E0A202020203C656E644576656E742069643D22456E644576656E745F3177716769707022206E616D653D22E7BB93E69D9F223E0A2020202020203C696E636F6D696E673E53657175656E6365466C6F775F303871637962793C2F696E636F6D696E673E0A202020203C2F656E644576656E743E0A202020203C73657175656E6365466C6F772069643D2253657175656E6365466C6F775F306A66656E64772220736F757263655265663D227369642D34363538384541412D333842372D344642432D383044442D34364135454645323643464122207461726765745265663D22557365725461736B5F307A7A366C637722202F3E0A202020203C73657175656E6365466C6F772069643D2253657175656E6365466C6F775F303871637962792220736F757263655265663D22557365725461736B5F307A7A366C637722207461726765745265663D22456E644576656E745F3177716769707022202F3E0A20203C2F70726F636573733E0A20203C62706D6E64693A42504D4E4469616772616D2069643D2242504D4E4469616772616D5F64656D6F5F7A6468746136393636363333333636223E0A202020203C62706D6E64693A42504D4E506C616E652069643D2242504D4E506C616E655F64656D6F5F7A6468746136393636363333333636222062706D6E456C656D656E743D2250726F636573735F5348415245303031223E0A2020202020203C62706D6E64693A42504D4E53686170652069643D2242504D4E53686170655F7369642D34363538384541412D333842372D344642432D383044442D343641354546453236434641222062706D6E456C656D656E743D227369642D34363538384541412D333842372D344642432D383044442D343641354546453236434641223E0A20202020202020203C6F6D6764633A426F756E647320783D222D31352220793D222D323335222077696474683D22353022206865696768743D22353022202F3E0A20202020202020203C62706D6E64693A42504D4E4C6162656C3E0A202020202020202020203C6F6D6764633A426F756E647320783D222D312220793D222D323135222077696474683D22323222206865696768743D22313422202F3E0A20202020202020203C2F62706D6E64693A42504D4E4C6162656C3E0A2020202020203C2F62706D6E64693A42504D4E53686170653E0A2020202020203C62706D6E64693A42504D4E53686170652069643D22557365725461736B5F307A7A366C63775F6469222062706D6E456C656D656E743D22557365725461736B5F307A7A366C6377223E0A20202020202020203C6F6D6764633A426F756E647320783D222D36302220793D222D3630222077696474683D2231343022206865696768743D2231303022202F3E0A2020202020203C2F62706D6E64693A42504D4E53686170653E0A2020202020203C62706D6E64693A42504D4E53686170652069643D22456E644576656E745F317771676970705F6469222062706D6E456C656D656E743D22456E644576656E745F31777167697070223E0A20202020202020203C6F6D6764633A426F756E647320783D222D31352220793D22313632222077696474683D22353022206865696768743D22353022202F3E0A20202020202020203C62706D6E64693A42504D4E4C6162656C3E0A202020202020202020203C6F6D6764633A426F756E647320783D222D312220793D22313830222077696474683D22323222206865696768743D22313422202F3E0A20202020202020203C2F62706D6E64693A42504D4E4C6162656C3E0A2020202020203C2F62706D6E64693A42504D4E53686170653E0A2020202020203C62706D6E64693A42504D4E456467652069643D2253657175656E6365466C6F775F306A66656E64775F6469222062706D6E456C656D656E743D2253657175656E6365466C6F775F306A66656E6477223E0A20202020202020203C64693A776179706F696E7420783D2231302220793D222D31383522202F3E0A20202020202020203C64693A776179706F696E7420783D2231302220793D222D363022202F3E0A2020202020203C2F62706D6E64693A42504D4E456467653E0A2020202020203C62706D6E64693A42504D4E456467652069643D2253657175656E6365466C6F775F303871637962795F6469222062706D6E456C656D656E743D2253657175656E6365466C6F775F30387163796279223E0A20202020202020203C64693A776179706F696E7420783D2231302220793D22343022202F3E0A20202020202020203C64693A776179706F696E7420783D2231302220793D2231363022202F3E0A2020202020203C2F62706D6E64693A42504D4E456467653E0A202020203C2F62706D6E64693A42504D4E506C616E653E0A20203C2F62706D6E64693A42504D4E4469616772616D3E0A3C2F646566696E6974696F6E733E, 0);
-- INSERT INTO `t_wf_ge_bytearray` VALUES ('52e06b12-d930-11ed-8086-00ff0fa3e6a7', 1, '匿名共享审核工作流.bpmn20.xml', '52e06b11-d930-11ed-8086-00ff0fa3e6a7', 0x3C3F786D6C2076657273696F6E3D22312E302220656E636F64696E673D225554462D38223F3E0A3C646566696E6974696F6E7320786D6C6E733D22687474703A2F2F7777772E6F6D672E6F72672F737065632F42504D4E2F32303130303532342F4D4F44454C2220786D6C6E733A7873693D22687474703A2F2F7777772E77332E6F72672F323030312F584D4C536368656D612D696E7374616E63652220786D6C6E733A62706D6E64693D22687474703A2F2F7777772E6F6D672E6F72672F737065632F42504D4E2F32303130303532342F44492220786D6C6E733A6F6D6764633D22687474703A2F2F7777772E6F6D672E6F72672F737065632F44442F32303130303532342F44432220786D6C6E733A64693D22687474703A2F2F7777772E6F6D672E6F72672F737065632F44442F32303130303532342F44492220786D6C6E733A61637469766974693D22687474703A2F2F61637469766974692E6F72672F62706D6E2220786D6C6E733A7873643D22687474703A2F2F7777772E77332E6F72672F323030312F584D4C536368656D6122207461726765744E616D6573706163653D22687474703A2F2F7777772E61637469766974692E6F72672F74657374223E0A20203C70726F636573732069643D2250726F636573735F534841524530303222206E616D653D22E58CBFE5908DE585B1E4BAABE5AEA1E6A0B8E5B7A5E4BD9CE6B5812220697345786563757461626C653D2274727565223E0A202020203C73746172744576656E742069643D227369642D34363538384541412D333842372D344642432D383044442D34364135454645323643464122206E616D653D22E58F91E8B5B7223E0A2020202020203C6F7574676F696E673E53657175656E6365466C6F775F306A66656E64773C2F6F7574676F696E673E0A202020203C2F73746172744576656E743E0A202020203C757365725461736B2069643D22557365725461736B5F307A7A366C637722206E616D653D22E5AEA1E6A0B8222061637469766974693A61737369676E65653D22247B61737369676E65657D222061637469766974693A63616E64696461746555736572733D22223E0A2020202020203C657874656E73696F6E456C656D656E74733E0A20202020202020203C61637469766974693A657870616E6450726F70657274792069643D226465616C54797065222076616C75653D22746A736822202F3E0A2020202020203C2F657874656E73696F6E456C656D656E74733E0A2020202020203C696E636F6D696E673E53657175656E6365466C6F775F306A66656E64773C2F696E636F6D696E673E0A2020202020203C6F7574676F696E673E53657175656E6365466C6F775F303871637962793C2F6F7574676F696E673E0A2020202020203C6D756C7469496E7374616E63654C6F6F7043686172616374657269737469637320697353657175656E7469616C3D2266616C7365222061637469766974693A636F6C6C656374696F6E3D22247B61737369676E65654C6973747D222061637469766974693A656C656D656E745661726961626C653D2261737369676E656522202F3E0A202020203C2F757365725461736B3E0A202020203C656E644576656E742069643D22456E644576656E745F3177716769707022206E616D653D22E7BB93E69D9F223E0A2020202020203C696E636F6D696E673E53657175656E6365466C6F775F303871637962793C2F696E636F6D696E673E0A202020203C2F656E644576656E743E0A202020203C73657175656E6365466C6F772069643D2253657175656E6365466C6F775F306A66656E64772220736F757263655265663D227369642D34363538384541412D333842372D344642432D383044442D34364135454645323643464122207461726765745265663D22557365725461736B5F307A7A366C637722202F3E0A202020203C73657175656E6365466C6F772069643D2253657175656E6365466C6F775F303871637962792220736F757263655265663D22557365725461736B5F307A7A366C637722207461726765745265663D22456E644576656E745F3177716769707022202F3E0A20203C2F70726F636573733E0A20203C62706D6E64693A42504D4E4469616772616D2069643D2242504D4E4469616772616D5F64656D6F5F7A6468746136393636363333333636223E0A202020203C62706D6E64693A42504D4E506C616E652069643D2242504D4E506C616E655F64656D6F5F7A6468746136393636363333333636222062706D6E456C656D656E743D2250726F636573735F5348415245303032223E0A2020202020203C62706D6E64693A42504D4E53686170652069643D2242504D4E53686170655F7369642D34363538384541412D333842372D344642432D383044442D343641354546453236434641222062706D6E456C656D656E743D227369642D34363538384541412D333842372D344642432D383044442D343641354546453236434641223E0A20202020202020203C6F6D6764633A426F756E647320783D222D31352220793D222D323335222077696474683D22353022206865696768743D22353022202F3E0A20202020202020203C62706D6E64693A42504D4E4C6162656C3E0A202020202020202020203C6F6D6764633A426F756E647320783D222D312220793D222D323135222077696474683D22323222206865696768743D22313422202F3E0A20202020202020203C2F62706D6E64693A42504D4E4C6162656C3E0A2020202020203C2F62706D6E64693A42504D4E53686170653E0A2020202020203C62706D6E64693A42504D4E53686170652069643D22557365725461736B5F307A7A366C63775F6469222062706D6E456C656D656E743D22557365725461736B5F307A7A366C6377223E0A20202020202020203C6F6D6764633A426F756E647320783D222D36302220793D222D3630222077696474683D2231343022206865696768743D2231303022202F3E0A2020202020203C2F62706D6E64693A42504D4E53686170653E0A2020202020203C62706D6E64693A42504D4E53686170652069643D22456E644576656E745F317771676970705F6469222062706D6E456C656D656E743D22456E644576656E745F31777167697070223E0A20202020202020203C6F6D6764633A426F756E647320783D222D31352220793D22313632222077696474683D22353022206865696768743D22353022202F3E0A20202020202020203C62706D6E64693A42504D4E4C6162656C3E0A202020202020202020203C6F6D6764633A426F756E647320783D222D312220793D22313830222077696474683D22323222206865696768743D22313422202F3E0A20202020202020203C2F62706D6E64693A42504D4E4C6162656C3E0A2020202020203C2F62706D6E64693A42504D4E53686170653E0A2020202020203C62706D6E64693A42504D4E456467652069643D2253657175656E6365466C6F775F306A66656E64775F6469222062706D6E456C656D656E743D2253657175656E6365466C6F775F306A66656E6477223E0A20202020202020203C64693A776179706F696E7420783D2231302220793D222D31383522202F3E0A20202020202020203C64693A776179706F696E7420783D2231302220793D222D363022202F3E0A2020202020203C2F62706D6E64693A42504D4E456467653E0A2020202020203C62706D6E64693A42504D4E456467652069643D2253657175656E6365466C6F775F303871637962795F6469222062706D6E456C656D656E743D2253657175656E6365466C6F775F30387163796279223E0A20202020202020203C64693A776179706F696E7420783D2231302220793D22343022202F3E0A20202020202020203C64693A776179706F696E7420783D2231302220793D2231363022202F3E0A2020202020203C2F62706D6E64693A42504D4E456467653E0A202020203C2F62706D6E64693A42504D4E506C616E653E0A20203C2F62706D6E64693A42504D4E4469616772616D3E0A3C2F646566696E6974696F6E733E, 0);

-- ----------------------------
-- Table structure for t_wf_ge_property
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_ge_property`  (
  `name_` varchar(64) NOT NULL COMMENT '名称',
  `value_` varchar(300) NULL DEFAULT NULL COMMENT '值',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  PRIMARY KEY (`name_`) USING BTREE
) ENGINE = InnoDB;

-- -- ----------------------------
-- -- Records of t_wf_ge_property
-- -- ----------------------------
-- INSERT INTO `t_wf_ge_property` VALUES ('next.dbid', '1', 1);
-- INSERT INTO `t_wf_ge_property` VALUES ('schema.history', 'create(7.0.4.7.0)', 1);
-- INSERT INTO `t_wf_ge_property` VALUES ('schema.version', '7.0.4.7.0', 1);

-- ----------------------------
-- Table structure for t_wf_hi_actinst
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_hi_actinst`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `proc_def_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程定义ID',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `execution_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程执行ID',
  `act_id_` varchar(255) NULL DEFAULT NULL COMMENT '活动ID',
  `task_id_` varchar(64) NULL DEFAULT NULL COMMENT '任务ID',
  `call_proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '调用外部流程的流程实例ID',
  `act_name_` varchar(255) NULL DEFAULT NULL COMMENT '活动名称',
  `act_type_` varchar(255) NULL DEFAULT NULL COMMENT '活动类型 如startEvent、userTask',
  `assignee_` varchar(255) NULL DEFAULT NULL COMMENT '代理人员',
  `start_time_` datetime(0) NULL DEFAULT NULL COMMENT '开始时间',
  `end_time_` datetime(0) NULL DEFAULT NULL COMMENT '结束时间',
  `duration_` bigint(20) NULL DEFAULT NULL COMMENT '时长，耗时',
  `tenant_id_` varchar(255) NULL DEFAULT NULL COMMENT '租户ID',
  `proc_def_name` varchar(500) NULL DEFAULT NULL COMMENT '流程定义名称',
  `proc_title` varchar(300) NULL DEFAULT NULL COMMENT '流程标题',
  `pre_act_id` varchar(255) NULL DEFAULT NULL COMMENT '父级活动ID',
  `pre_act_name` varchar(255) NULL DEFAULT NULL COMMENT '父级活动名称',
  `pre_act_inst_id` varchar(255) NULL DEFAULT NULL COMMENT '父级活动实例ID',
  `create_time_` timestamp(0) NULL DEFAULT NULL COMMENT '创建时间',
  `last_updated_time_` timestamp(0) NULL DEFAULT NULL COMMENT '最后更新时间',
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_IDX_HI_ACT_INST_START` (`start_time_`) USING BTREE,
  KEY `ACT_IDX_HI_ACT_INST_END` (`end_time_`) USING BTREE,
  KEY `ACT_IDX_HI_ACT_INST_PROCINST` (`proc_inst_id_`, `act_id_`) USING BTREE,
  KEY `ACT_IDX_HI_ACT_INST_EXEC` (`execution_id_`, `act_id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_hi_actinst
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_hi_attachment
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_hi_attachment`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `user_id_` varchar(255) NULL DEFAULT NULL COMMENT '用户ID',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '附件名称',
  `description_` varchar(4000) NULL DEFAULT NULL COMMENT '描述',
  `type_` varchar(255) NULL DEFAULT NULL COMMENT '附件类型',
  `task_id_` varchar(64) NULL DEFAULT NULL COMMENT '任务Id',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `url_` varchar(4000) NULL DEFAULT NULL COMMENT '附件地址',
  `content_id_` varchar(64) NULL DEFAULT NULL COMMENT '内容Id（字节表的ID）',
  `time_` datetime(0) NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_hi_attachment
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_hi_comment
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_hi_comment`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `type_` varchar(255) NULL DEFAULT NULL COMMENT '类型：event（事件）comment（意见）',
  `time_` datetime(0) NULL DEFAULT NULL COMMENT '填写时间',
  `user_id_` varchar(255) NULL DEFAULT NULL COMMENT '填写人',
  `task_id_` varchar(64) NULL DEFAULT NULL COMMENT '任务Id',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `action_` varchar(255) NULL DEFAULT NULL COMMENT '行为类型 （值为下列内容中的一种：AddUserLink、DeleteUserLink、AddGroupLink、DeleteGroupLink、AddComment、AddAttachment、DeleteAttachment）',
  `message_` varchar(4000) NULL DEFAULT NULL COMMENT '处理意见',
  `full_msg_` longblob NULL COMMENT '全部消息',
  `display_area` varchar(500) NULL DEFAULT NULL,
  `top_proc_inst_id_` varchar(100) NULL DEFAULT NULL COMMENT '顶级流程实例ID',
  PRIMARY KEY (`id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_hi_comment
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_hi_detail
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_hi_detail`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `type_` varchar(255) NULL DEFAULT NULL COMMENT '类型:（表单：FormProperty；参数：VariableUpdate）',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `execution_id_` varchar(64) NULL DEFAULT NULL COMMENT '执行实例ID',
  `task_id_` varchar(64) NULL DEFAULT NULL COMMENT '任务实例ID',
  `act_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '活动实例Id',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '名称',
  `var_type_` varchar(255) NULL DEFAULT NULL COMMENT '变量类型',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `time_` datetime(0) NULL DEFAULT NULL COMMENT '创建时间',
  `bytearray_id_` varchar(64) NULL DEFAULT NULL COMMENT '字节数组Id',
  `double_` double NULL DEFAULT NULL COMMENT '存储变量类型为Double',
  `long_` bigint(20) NULL DEFAULT NULL COMMENT '存储变量类型为long',
  `text_` varchar(4000) NULL DEFAULT NULL COMMENT '存储变量值类型为String',
  `text2_` varchar(4000) NULL DEFAULT NULL COMMENT '此处存储的是JPA持久化对象时，才会有值。此值为对象ID',
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_IDX_HI_DETAIL_PROC_INST` (`proc_inst_id_`) USING BTREE,
  KEY `ACT_IDX_HI_DETAIL_ACT_INST` (`act_inst_id_`) USING BTREE,
  KEY `ACT_IDX_HI_DETAIL_TIME` (`time_`) USING BTREE,
  KEY `ACT_IDX_HI_DETAIL_NAME` (`name_`) USING BTREE,
  KEY `ACT_IDX_HI_DETAIL_TASK_ID` (`task_id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_hi_detail
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_hi_identitylink
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_hi_identitylink`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `group_id_` varchar(255) NULL DEFAULT NULL COMMENT '用户组ID',
  `type_` varchar(255) NULL DEFAULT NULL COMMENT '用户组类型',
  `user_id_` varchar(255) NULL DEFAULT NULL COMMENT '用户ID',
  `task_id_` varchar(64) NULL DEFAULT NULL COMMENT '任务Id',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_IDX_HI_IDENT_LNK_USER` (`user_id_`) USING BTREE,
  KEY `ACT_IDX_HI_IDENT_LNK_TASK` (`task_id_`) USING BTREE,
  KEY `ACT_IDX_HI_IDENT_LNK_PROCINST` (`proc_inst_id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_hi_identitylink
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_hi_procinst
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_hi_procinst`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `business_key_` text NULL COMMENT '文档ID，如：gns://xxx/xxx',
  `proc_def_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程定义Id',
  `start_time_` datetime(0) NULL DEFAULT NULL COMMENT '开始时间',
  `end_time_` datetime(0) NULL DEFAULT NULL COMMENT '结束时间',
  `duration_` bigint(20) NULL DEFAULT NULL COMMENT '时长',
  `start_user_id_` varchar(255) NULL DEFAULT NULL COMMENT '发起人员Id',
  `start_act_id_` varchar(255) NULL DEFAULT NULL COMMENT '开始节点',
  `end_act_id_` varchar(255) NULL DEFAULT NULL COMMENT '结束节点',
  `super_process_instance_id_` varchar(64) NULL DEFAULT NULL COMMENT '超级流程实例Id',
  `delete_reason_` varchar(4000) NULL DEFAULT NULL COMMENT '删除理由',
  `tenant_id_` varchar(255) NULL DEFAULT NULL COMMENT '租户ID',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '名称',
  `proc_state` int(11) NULL DEFAULT NULL COMMENT '流程状态',
  `proc_def_name` varchar(500) NULL DEFAULT NULL COMMENT '流程定义名称',
  `start_user_name` varchar(100) NULL DEFAULT NULL COMMENT '发起人名称',
  `starter_org_id` varchar(100) NULL DEFAULT NULL COMMENT '发起人组织ID',
  `starter_org_name` varchar(100) NULL DEFAULT NULL COMMENT '发起人组织名称',
  `starter` varchar(100) NULL DEFAULT NULL COMMENT '发起人',
  `top_process_instance_id_` varchar(100) NULL DEFAULT NULL COMMENT '顶级流程实例ID',
  PRIMARY KEY (`id_`) USING BTREE,
  UNIQUE KEY `PROC_INST_ID_` (`proc_inst_id_`) USING BTREE,
  KEY `ACT_IDX_HI_PRO_INST_END` (`end_time_`) USING BTREE,
  KEY `ACT_IDX_HI_PRO_I_BUSKEY` (`business_key_`(50)) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_hi_procinst
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_hi_taskinst
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_hi_taskinst`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `proc_def_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程定义ID',
  `task_def_key_` varchar(255) NULL DEFAULT NULL COMMENT '任务定义Key',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `execution_id_` varchar(64) NULL DEFAULT NULL COMMENT '执行实例ID',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '名称',
  `parent_task_id_` varchar(64) NULL DEFAULT NULL COMMENT '父节点实例ID',
  `description_` varchar(500) NULL DEFAULT NULL,
  `owner_` varchar(255) NULL DEFAULT NULL COMMENT '实际签收人 任务的拥有者',
  `assignee_` varchar(255) NULL DEFAULT NULL COMMENT '代理人',
  `start_time_` datetime(0) NULL DEFAULT NULL COMMENT '开始时间',
  `claim_time_` datetime(0) NULL DEFAULT NULL COMMENT '提醒时间',
  `end_time_` datetime(0) NULL DEFAULT NULL COMMENT '结束时间',
  `duration_` bigint(20) NULL DEFAULT NULL COMMENT '时长',
  `delete_reason_` varchar(500) NULL DEFAULT NULL,
  `priority_` int(11) NULL DEFAULT NULL COMMENT '优先级',
  `due_date_` datetime(0) NULL DEFAULT NULL COMMENT '应完成时间',
  `form_key_` varchar(255) NULL DEFAULT NULL COMMENT '表单key',
  `category_` varchar(255) NULL DEFAULT NULL,
  `tenant_id_` varchar(255) NULL DEFAULT NULL COMMENT '租户ID',
  `proc_title` varchar(2000) NULL DEFAULT NULL COMMENT '流程标题',
  `sender` varchar(64) NULL DEFAULT NULL,
  `pre_task_def_key` varchar(64) NULL DEFAULT NULL COMMENT '父级任务定义key',
  `pre_task_id` varchar(64) NULL DEFAULT NULL COMMENT '父级任务ID',
  `pre_task_def_name` varchar(255) NULL DEFAULT NULL COMMENT '父级任务名称',
  `action_type` varchar(64) NULL DEFAULT NULL,
  `top_execution_id_` varchar(64) NULL DEFAULT NULL COMMENT '顶级执行ID',
  `sender_org_id` varchar(100) NULL DEFAULT NULL COMMENT '发送人组织ID',
  `assignee_org_id` varchar(100) NULL DEFAULT NULL COMMENT '代理人组织ID',
  `proc_def_name` varchar(500) NULL DEFAULT NULL COMMENT '流程定义名称',
  `status` varchar(100) NULL DEFAULT NULL COMMENT '审核状态，1-审核中 2-已拒绝 3-已通过 4-自动审核通过 5-作废 6-发起失败 70-已撤销',
  `biz_id` varchar(100) NULL DEFAULT NULL COMMENT '业务主键',
  `doc_id` text NULL COMMENT '文档ID，如：gns://xxx/xxx',
  `doc_name` varchar(1000) NULL DEFAULT NULL COMMENT '文档名称',
  `doc_path` varchar(1000) NULL DEFAULT NULL COMMENT '文档路径',
  `addition` mediumtext NULL COMMENT '业务字段',
  `message_id` varchar(64) NOT NULL DEFAULT '' COMMENT '消息ID',
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_IDX_HI_TASK_INST_PROCINST` (`proc_inst_id_`) USING BTREE,
  KEY `ACT_IDX_HI_TASK_INST_END_TIME` (`end_time_`) USING BTREE,
  KEY `ACT_IDX_HI_TASK_DELETE_REASON_` (`assignee_`, `delete_reason_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_hi_taskinst
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_hi_varinst
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_hi_varinst`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `execution_id_` varchar(64) NULL DEFAULT NULL COMMENT '执行实例ID',
  `task_id_` varchar(64) NULL DEFAULT NULL COMMENT '任务Id',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '名称',
  `var_type_` varchar(100) NULL DEFAULT NULL COMMENT '变量类型',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `bytearray_id_` varchar(64) NULL DEFAULT NULL COMMENT '字节数组ID',
  `double_` double NULL DEFAULT NULL COMMENT '存储DoubleType类型的数据',
  `long_` bigint(20) NULL DEFAULT NULL COMMENT '存储LongType类型的数据',
  `text_` varchar(4000) NULL DEFAULT NULL COMMENT '存储变量值类型为String，如此处存储持久化对象时，值jpa对象的class',
  `text2_` varchar(4000) NULL DEFAULT NULL COMMENT '此处存储的是JPA持久化对象时，才会有值。此值为对象ID',
  `create_time_` datetime(0) NULL DEFAULT NULL COMMENT '创建时间',
  `last_updated_time_` datetime(0) NULL DEFAULT NULL COMMENT '最后更新时间',
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_IDX_HI_PROCVAR_PROC_INST` (`proc_inst_id_`) USING BTREE,
  KEY `ACT_IDX_HI_PROCVAR_NAME_TYPE` (`name_`, `var_type_`) USING BTREE,
  KEY `ACT_IDX_HI_PROCVAR_TASK_ID` (`task_id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_hi_varinst
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_org
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_org`  (
  `org_id` varchar(50) NOT NULL COMMENT '组织编码',
  `org_name` varchar(200) NOT NULL COMMENT '组织简称',
  `org_full_name` varchar(500) NOT NULL COMMENT '组织全称',
  `org_full_path_name` varchar(4000) NULL DEFAULT NULL COMMENT '组织全路径名称',
  `org_full_path_id` varchar(4000) NULL DEFAULT NULL COMMENT '组织全路径ID',
  `org_parent_id` varchar(50) NOT NULL COMMENT '上级组织编码',
  `org_type` varchar(10) NOT NULL COMMENT '组织类型',
  `org_level` int(11) NOT NULL COMMENT '组织级别',
  `org_area_type` varchar(10) NOT NULL COMMENT '政府单位区域类别',
  `org_sort` int(11) NULL DEFAULT NULL COMMENT '组织排序号',
  `org_work_phone` varchar(50) NULL DEFAULT NULL COMMENT '组织工作手机号',
  `org_work_address` varchar(1000) NULL DEFAULT NULL COMMENT '组织工作地址',
  `org_principal` varchar(100) NULL DEFAULT NULL COMMENT '组织负责人',
  `org_status` varchar(10) NOT NULL COMMENT '组织状态',
  `org_create_time` date NULL DEFAULT NULL COMMENT '组织创建时间',
  `remark` varchar(500) NULL DEFAULT NULL COMMENT '备注',
  `fund_code` varchar(20) NULL DEFAULT NULL COMMENT '基金编码',
  `fund_name` varchar(100) NULL DEFAULT NULL COMMENT '基金名称',
  `company_id` varchar(50) NULL DEFAULT NULL COMMENT '公司ID',
  `dept_id` varchar(50) NULL DEFAULT NULL COMMENT '部门ID',
  `dept_name` varchar(100) NULL DEFAULT NULL COMMENT '部门名称',
  `company_name` varchar(100) NULL DEFAULT NULL COMMENT '公司名称',
  `org_branch_leader` varchar(50) NULL DEFAULT NULL COMMENT '分管领导ID',
  PRIMARY KEY (`org_id`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_org
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_procdef_info
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_procdef_info`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `proc_def_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程定义ID',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `info_json_id_` varchar(64) NULL DEFAULT NULL,
  CONSTRAINT `ACT_FK_INFO_JSON_BA` FOREIGN KEY (`info_json_id_`) REFERENCES `t_wf_ge_bytearray` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  PRIMARY KEY (`id_`) USING BTREE,
  UNIQUE KEY `ACT_UNIQ_INFO_PROCDEF` (`proc_def_id_`) USING BTREE,
  KEY `ACT_IDX_INFO_PROCDEF` (`proc_def_id_`) USING BTREE,
  KEY `ACT_FK_INFO_JSON_BA` (`info_json_id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_procdef_info
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_process_error_log
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_process_error_log`  (
  `pelog_id` varchar(36) NOT NULL COMMENT '主键，GUID',
  `process_instance_id` varchar(50) NULL DEFAULT NULL COMMENT '流程实例ID',
  `process_title` varchar(500) NULL DEFAULT NULL COMMENT '流程标题',
  `creator` varchar(50) NULL DEFAULT NULL COMMENT '流程发送人',
  `action_type` varchar(50) NULL DEFAULT NULL COMMENT '流程操作类型',
  `process_msg` mediumtext NULL COMMENT '流程消息内容',
  `pelog_create_time` datetime(0) NULL DEFAULT NULL COMMENT '记录时间',
  `receivers` varchar(4000) NULL DEFAULT NULL COMMENT '任务接收者',
  `process_def_name` varchar(500) NULL DEFAULT NULL COMMENT '流程定义名称',
  `app_id` varchar(100) NULL DEFAULT NULL COMMENT '应用ID',
  `process_log_level` varchar(20) NULL DEFAULT NULL COMMENT '日志级别，INFO-信息，ERROR异常',
  `retry_status` varchar(2) NULL DEFAULT NULL COMMENT '错误重试状态,y:已处理：n：未处理',
  `error_msg` mediumtext NULL COMMENT '异常信息',
  `user_time` varchar(100) NULL DEFAULT NULL COMMENT '耗时（毫秒）',
  PRIMARY KEY (`pelog_id`) USING BTREE,
  KEY `INDEX_T_WF_ERROR_LOG_APPID` (`app_id`, `process_log_level`, `pelog_create_time`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_process_error_log
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_process_info_config
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_process_info_config`  (
  `process_def_id` varchar(100) NOT NULL COMMENT '流程定义ID',
  `process_def_name` varchar(500) NULL DEFAULT NULL COMMENT '流程定义名称',
  `process_def_key` varchar(100) NULL DEFAULT NULL COMMENT '流程KEY',
  `process_type_id` varchar(50) NULL DEFAULT NULL COMMENT '流程类型ID',
  `process_type_name` varchar(100) NULL DEFAULT NULL COMMENT '流程类型名称',
  `process_page_url` varchar(1000) NULL DEFAULT NULL COMMENT '流程表单URL',
  `process_page_info` varchar(4000) NULL DEFAULT NULL COMMENT '流程表单数据',
  `process_start_auth` varchar(500) NULL DEFAULT NULL COMMENT '流程表单起草权限',
  `process_start_isshow` varchar(10) NULL DEFAULT NULL COMMENT '新建流程是否可见，Y-可见，N-不可见',
  `remark` varchar(1000) NULL DEFAULT NULL COMMENT '备注',
  `page_isshow_select_usertree` decimal(10, 0) NULL DEFAULT NULL COMMENT '是否展示选人组件',
  `process_handler_class_path` varchar(500) NULL DEFAULT NULL COMMENT '流程处理程序类路径',
  `process_start_order` decimal(10, 0) NULL DEFAULT NULL COMMENT '排序号',
  `deployment_id` varchar(64) NULL DEFAULT NULL COMMENT '部署ID',
  `create_time` datetime(0) NULL DEFAULT NULL COMMENT '创建时间',
  `last_update_time` datetime(0) NULL DEFAULT NULL COMMENT '最后一次修改时间',
  `create_user` varchar(50) NULL DEFAULT NULL COMMENT '创建人',
  `create_user_name` varchar(150) NULL DEFAULT NULL COMMENT '创建人名称',
  `last_update_user` varchar(50) NULL DEFAULT NULL COMMENT '最后一次修改人',
  `tenant_id` varchar(255) NULL DEFAULT NULL COMMENT '租户ID',
  `process_mgr_state` varchar(20) NULL DEFAULT NULL COMMENT '流程定义管理状态，UNRELEASE-未发布，UPDATE-修订中，RELEASE-已发布',
  `process_model_sync_state` varchar(10) NULL DEFAULT NULL COMMENT '流程定义与模型同步状态，Y:已同步,N:未同步',
  `process_mgr_isshow` varchar(10) NULL DEFAULT NULL COMMENT '流程定义管理状态：Y:可见，N:不可见',
  `aris_code` varchar(100) NULL DEFAULT NULL COMMENT 'arisr流程编码',
  `c_protocl` varchar(50) NULL DEFAULT NULL COMMENT 'PC端协议',
  `m_protocl` varchar(50) NULL DEFAULT NULL COMMENT '移动端协议',
  `m_url` varchar(500) NULL DEFAULT NULL COMMENT '移动端待办地址',
  `other_sys_deal_status` varchar(10) NULL DEFAULT NULL COMMENT '移动端处理状态',
  `template` varchar(10) NULL DEFAULT NULL COMMENT '是否是流程模板 Y-是',
  PRIMARY KEY (`process_def_id`) USING BTREE
) ENGINE = InnoDB;

-- -- ----------------------------
-- -- Records of t_wf_process_info_config
-- -- ----------------------------
-- INSERT INTO `t_wf_process_info_config` VALUES ('Process_SHARE001:1:52af48f0-d930-11ed-8086-00ff0fa3e6a7', '实名共享审核工作流', 'Process_SHARE001', 'doc_share', '文档共享审核', NULL, NULL, NULL, 'Y', NULL, NULL, NULL, NULL, '526128ee-d930-11ed-8086-00ff0fa3e6a7', '2023-04-12 20:48:28', '2023-04-12 20:48:28', '', NULL, NULL, 'as_workflow', 'UNRELEASE', 'Y', 'Y', NULL, 'hnzy:workitem', 'none', NULL, 'yes', NULL);
-- INSERT INTO `t_wf_process_info_config` VALUES ('Process_SHARE002:1:52e489c3-d930-11ed-8086-00ff0fa3e6a7', '匿名共享审核工作流', 'Process_SHARE002', 'doc_share', '文档共享审核', NULL, NULL, NULL, 'Y', NULL, NULL, NULL, NULL, '52e06b11-d930-11ed-8086-00ff0fa3e6a7', '2023-04-12 20:48:29', '2023-04-12 20:48:29', '', NULL, NULL, 'as_workflow', 'UNRELEASE', 'Y', 'Y', NULL, 'hnzy:workitem', 'none', NULL, 'yes', NULL);

-- ----------------------------
-- Table structure for t_wf_re_deployment
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_re_deployment`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '部署包的名称',
  `category_` varchar(255) NULL DEFAULT NULL COMMENT '类型',
  `tenant_id_` varchar(255) NULL DEFAULT NULL COMMENT '租户',
  `deploy_time_` timestamp(0) NULL DEFAULT NULL COMMENT '部署时间',
  PRIMARY KEY (`id_`) USING BTREE
) ENGINE = InnoDB;

-- -- ----------------------------
-- -- Records of t_wf_re_deployment
-- -- ----------------------------
-- INSERT INTO `t_wf_re_deployment` VALUES ('526128ee-d930-11ed-8086-00ff0fa3e6a7', '实名共享审核工作流', 'doc_share', 'as_workflow', '2023-04-12 20:48:28');
-- INSERT INTO `t_wf_re_deployment` VALUES ('52e06b11-d930-11ed-8086-00ff0fa3e6a7', '匿名共享审核工作流', 'doc_share', 'as_workflow', '2023-04-12 20:48:29');

-- ----------------------------
-- Table structure for t_wf_re_model
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_re_model`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '模型的名称：比如：收文管理',
  `key_` varchar(255) NULL DEFAULT NULL COMMENT '模型的关键字，流程引擎用到。比如：FTOA_SWGL',
  `category_` varchar(255) NULL DEFAULT NULL COMMENT '类型，用户自己对流程模型的分类。',
  `create_time_` timestamp(0) NULL DEFAULT NULL COMMENT '创建时间',
  `last_update_time_` timestamp(0) NULL DEFAULT NULL COMMENT '最后修改时间',
  `version_` int(11) NULL DEFAULT NULL COMMENT '版本，从1开始。',
  `meta_info_` varchar(4000) NULL DEFAULT NULL COMMENT '数据源信息，比如：\n            {\"name\":\"FTOA_SWGL\",\"revision\":1,\"description\":\"丰台财政局OA，收文管理流程\"}',
  `deployment_id_` varchar(64) NULL DEFAULT NULL COMMENT '部署ID',
  `editor_source_value_id_` varchar(64) NULL DEFAULT NULL COMMENT '编辑源值ID',
  `editor_source_extra_value_id_` varchar(64) NULL DEFAULT NULL COMMENT '编辑源额外值ID（外键ACT_GE_BYTEARRAY ）',
  `tenant_id_` varchar(255) NULL DEFAULT NULL COMMENT '租户',
  `model_state` varchar(10) NULL DEFAULT NULL COMMENT '状态',
  CONSTRAINT `ACT_FK_MODEL_DEPLOYMENT` FOREIGN KEY (`deployment_id_`) REFERENCES `t_wf_re_deployment` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `ACT_FK_MODEL_SOURCE` FOREIGN KEY (`editor_source_value_id_`) REFERENCES `t_wf_ge_bytearray` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `ACT_FK_MODEL_SOURCE_EXTRA` FOREIGN KEY (`editor_source_extra_value_id_`) REFERENCES `t_wf_ge_bytearray` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_FK_MODEL_SOURCE` (`editor_source_value_id_`) USING BTREE,
  KEY `ACT_FK_MODEL_SOURCE_EXTRA` (`editor_source_extra_value_id_`) USING BTREE,
  KEY `ACT_FK_MODEL_DEPLOYMENT` (`deployment_id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_re_model
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_re_procdef
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_re_procdef`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `category_` varchar(255) NULL DEFAULT NULL COMMENT '流程命名空间（该编号就是流程文件targetNamespace的属性值）',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '流程名称（该编号就是流程文件process元素的name属性值）',
  `key_` varchar(255) NULL DEFAULT NULL COMMENT '流程编号（该编号就是流程文件process元素的id属性值）',
  `version_` int(11) NOT NULL COMMENT '流程版本号（由程序控制，新增即为1，修改后依次加1来完成的）',
  `deployment_id_` varchar(64) NULL DEFAULT NULL COMMENT '部署表ID',
  `resource_name_` varchar(4000) NULL DEFAULT NULL COMMENT '资源文件名称',
  `dgrm_resource_name_` varchar(4000) NULL DEFAULT NULL COMMENT '图片资源文件名称',
  `description_` varchar(4000) NULL DEFAULT NULL COMMENT '描述信息',
  `has_start_form_key_` tinyint(4) NULL DEFAULT NULL COMMENT '是否从key启动（start节点是否存在formKey 0否  1是）',
  `has_graphical_notation_` tinyint(4) NULL DEFAULT NULL,
  `suspension_state_` int(11) NULL DEFAULT NULL COMMENT '是否挂起 1激活 2挂起',
  `tenant_id_` varchar(255) NULL DEFAULT NULL COMMENT '租户ID',
  `org_id_` varchar(100) NULL DEFAULT NULL COMMENT '组织ID',
  PRIMARY KEY (`id_`) USING BTREE,
  UNIQUE KEY `ACT_UNIQ_PROCDEF` (`key_`, `version_`, `tenant_id_`) USING BTREE
) ENGINE = InnoDB;

-- -- ----------------------------
-- -- Records of t_wf_re_procdef
-- -- ----------------------------
-- INSERT INTO `t_wf_re_procdef` VALUES ('Process_SHARE001:1:52af48f0-d930-11ed-8086-00ff0fa3e6a7', 1, 'doc_share', '实名共享审核工作流', 'Process_SHARE001', 1, '526128ee-d930-11ed-8086-00ff0fa3e6a7', '实名共享审核工作流.bpmn20.xml', NULL, NULL, 0, 1, 1, 'as_workflow', NULL);
-- INSERT INTO `t_wf_re_procdef` VALUES ('Process_SHARE002:1:52e489c3-d930-11ed-8086-00ff0fa3e6a7', 1, 'doc_share', '匿名共享审核工作流', 'Process_SHARE002', 1, '52e06b11-d930-11ed-8086-00ff0fa3e6a7', '匿名共享审核工作流.bpmn20.xml', NULL, NULL, 0, 1, 1, 'as_workflow', NULL);

-- ----------------------------
-- Table structure for t_wf_role
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_role`  (
  `role_id` varchar(50) NOT NULL COMMENT '角色ID',
  `role_name` varchar(500) NULL DEFAULT NULL COMMENT '角色名称',
  `role_type` varchar(20) NULL DEFAULT NULL COMMENT '角色类型',
  `role_sort` int(11) NULL DEFAULT NULL COMMENT '角色排序号',
  `role_org_id` int(11) NULL DEFAULT NULL COMMENT '角色组织ID',
  `role_app_id` varchar(500) NULL DEFAULT NULL COMMENT '角色所属租户',
  `role_status` varchar(10) NOT NULL COMMENT '角色状态',
  `role_create_time` datetime(0) NULL DEFAULT NULL COMMENT '角色创建时间',
  `role_creator` varchar(50) NULL DEFAULT NULL COMMENT '角色创建者',
  `remark` varchar(500) NULL DEFAULT NULL COMMENT '备注',
  `template` varchar(10) NULL DEFAULT NULL COMMENT '是否是流程模板 Y-是',
  PRIMARY KEY (`role_id`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_role
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_ru_event_subscr
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_ru_event_subscr`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `event_type_` varchar(255) NULL DEFAULT NULL COMMENT '事件类型',
  `event_name_` varchar(255) NULL DEFAULT NULL COMMENT '事件名称',
  `execution_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程执行ID',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `activity_id_` varchar(64) NULL DEFAULT NULL COMMENT '活动ID',
  `configuration_` varchar(255) NULL DEFAULT NULL COMMENT '配置信息',
  `created_` timestamp(0) NOT NULL DEFAULT current_timestamp(0) COMMENT '创建时间',
  `proc_def_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程定义ID',
  `tenant_id_` varchar(255) NULL DEFAULT NULL COMMENT '租户ID',
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_IDX_EVENT_SUBSCR_CONFIG_` (`configuration_`) USING BTREE,
  KEY `ACT_FK_EVENT_EXEC` (`execution_id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_ru_event_subscr
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_ru_execution
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_ru_execution`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `business_key_` text NULL COMMENT '文档ID，如：gns://xxx/xxx',
  `parent_id_` varchar(64) NULL DEFAULT NULL COMMENT '父节点实例ID',
  `proc_def_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程定义ID',
  `super_exec_` varchar(64) NULL DEFAULT NULL,
  `act_id_` varchar(255) NULL DEFAULT NULL COMMENT '节点实例ID即ACT_HI_ACTINST中ID',
  `is_active_` tinyint(4) NULL DEFAULT NULL COMMENT '是否存活',
  `is_concurrent_` tinyint(4) NULL DEFAULT NULL COMMENT '是否为并行',
  `is_scope_` tinyint(4) NULL DEFAULT NULL,
  `is_event_scope_` tinyint(4) NULL DEFAULT NULL,
  `suspension_state_` int(11) NULL DEFAULT NULL COMMENT '挂起状态   1激活 2挂起',
  `cached_ent_state_` int(11) NULL DEFAULT NULL COMMENT '缓存结束状态',
  `tenant_id_` varchar(255) NULL DEFAULT NULL COMMENT '租户ID',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '名称',
  `lock_time_` timestamp(0) NULL DEFAULT NULL,
  `top_process_instance_id_` varchar(100) NULL DEFAULT NULL COMMENT '顶级流程实例ID',
  CONSTRAINT `ACT_FK_EXE_PROCDEF` FOREIGN KEY (`proc_def_id_`) REFERENCES `t_wf_re_procdef` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `ACT_FK_EXE_SUPER` FOREIGN KEY (`super_exec_`) REFERENCES `t_wf_ru_execution` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_FK_EXE_PROCINST` (`proc_inst_id_`) USING BTREE,
  KEY `ACT_FK_EXE_PARENT` (`parent_id_`) USING BTREE,
  KEY `ACT_FK_EXE_SUPER` (`super_exec_`) USING BTREE,
  KEY `ACT_FK_EXE_PROCDEF` (`proc_def_id_`) USING BTREE,
  KEY `ACT_IDX_EXEC_BUSKEY` (`business_key_`(50)) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_ru_execution
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_ru_identitylink
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_ru_identitylink`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `group_id_` varchar(255) NULL DEFAULT NULL COMMENT '用户组ID',
  `type_` varchar(255) NULL DEFAULT NULL COMMENT '用户组类型（主要分为以下几种：assignee、candidate、\n\n            owner、starter、participant。即：受让人,候选人,所有者、起动器、参与者）',
  `user_id_` varchar(255) NULL DEFAULT NULL COMMENT '用户ID',
  `task_id_` varchar(64) NULL DEFAULT NULL COMMENT '任务Id',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `proc_def_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程定义Id',
  `org_id_` varchar(100) NULL DEFAULT NULL COMMENT '组织ID',
  CONSTRAINT `ACT_FK_ATHRZ_PROCEDEF` FOREIGN KEY (`proc_def_id_`) REFERENCES `t_wf_re_procdef` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `ACT_FK_IDL_PROCINST` FOREIGN KEY (`proc_inst_id_`) REFERENCES `t_wf_ru_execution` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_IDX_IDENT_LNK_USER` (`user_id_`) USING BTREE,
  KEY `ACT_IDX_IDENT_LNK_GROUP` (`group_id_`) USING BTREE,
  KEY `ACT_IDX_ATHRZ_PROCEDEF` (`proc_def_id_`) USING BTREE,
  KEY `ACT_FK_TSKASS_TASK` (`task_id_`) USING BTREE,
  KEY `ACT_FK_IDL_PROCINST` (`proc_inst_id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_ru_identitylink
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_ru_job
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_ru_job`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `type_` varchar(255) NULL DEFAULT NULL COMMENT '类型',
  `lock_exp_time_` timestamp(0) NULL DEFAULT NULL COMMENT '锁定释放时间',
  `lock_owner_` varchar(255) NULL DEFAULT NULL COMMENT '挂起者',
  `exclusive_` boolean NULL DEFAULT NULL,
  `execution_id_` varchar(64) NULL DEFAULT NULL COMMENT '执行实例ID',
  `process_instance_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `proc_def_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程定义ID',
  `retries_` int(11) NULL DEFAULT NULL,
  `exception_stack_id_` varchar(64) NULL DEFAULT NULL COMMENT '异常信息ID',
  `exception_msg_` varchar(4000) NULL DEFAULT NULL COMMENT '异常信息',
  `duedate_` timestamp(0) NULL DEFAULT NULL COMMENT '到期时间',
  `repeat_` varchar(255) NULL DEFAULT NULL COMMENT '重复',
  `handler_type_` varchar(255) NULL DEFAULT NULL COMMENT '处理类型',
  `handler_cfg_` varchar(4000) NULL DEFAULT NULL COMMENT '标识',
  `tenant_id_` varchar(255) NULL DEFAULT NULL COMMENT '租户ID',
  CONSTRAINT `ACT_FK_JOB_EXCEPTION` FOREIGN KEY (`exception_stack_id_`) REFERENCES `t_wf_ge_bytearray` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_FK_JOB_EXCEPTION` (`exception_stack_id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_ru_job
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_ru_task
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_ru_task`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `execution_id_` varchar(64) NULL DEFAULT NULL COMMENT '执行实例ID',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例ID',
  `proc_def_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程定义ID',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '任务名称',
  `parent_task_id_` varchar(64) NULL DEFAULT NULL COMMENT '父节任务ID',
  `description_` varchar(4000) NULL DEFAULT NULL COMMENT '任务描述',
  `task_def_key_` varchar(255) NULL DEFAULT NULL COMMENT '任务定义key',
  `owner_` varchar(255) NULL DEFAULT NULL COMMENT '拥有者',
  `assignee_` varchar(255) NULL DEFAULT NULL COMMENT '代理人',
  `delegation_` varchar(64) NULL DEFAULT NULL COMMENT '委托类型，DelegationState分为两种：PENDING，RESOLVED。如无委托则为空',
  `priority_` int(11) NULL DEFAULT NULL COMMENT '优先级别',
  `create_time_` timestamp(0) NULL DEFAULT NULL COMMENT '创建时间',
  `due_date_` datetime(0) NULL DEFAULT NULL COMMENT '执行时间',
  `category_` varchar(255) NULL DEFAULT NULL,
  `suspension_state_` int(11) NULL DEFAULT NULL COMMENT '暂停状态 1代表激活 2代表挂起',
  `tenant_id_` varchar(255) NULL DEFAULT NULL COMMENT '租户ID',
  `form_key_` varchar(255) NULL DEFAULT NULL,
  `proc_title` varchar(2000) NULL DEFAULT NULL COMMENT '流程标题',
  `sender` varchar(64) NULL DEFAULT NULL,
  `pre_task_def_key` varchar(64) NULL DEFAULT NULL COMMENT '父级任务key',
  `pre_task_id` varchar(64) NULL DEFAULT NULL COMMENT '父级任务ID',
  `pre_task_def_name` varchar(255) NULL DEFAULT NULL COMMENT '父级任务名称',
  `action_type` varchar(64) NULL DEFAULT NULL,
  `sender_org_id` varchar(100) NULL DEFAULT NULL COMMENT '发送人组织ID',
  `assignee_org_id` varchar(100) NULL DEFAULT NULL COMMENT '代理人组织ID',
  `proc_def_name` varchar(500) NULL DEFAULT NULL COMMENT '流程定义名称',
  `biz_id` varchar(100) NULL DEFAULT NULL COMMENT '业务主键',
  `doc_id` text NULL COMMENT '文档ID，如：gns://xxx/xxx',
  `doc_name` varchar(1000) NULL DEFAULT NULL COMMENT '文档名称',
  `doc_path` varchar(1000) NULL DEFAULT NULL COMMENT '文档路径',
  `addition` mediumtext NULL COMMENT '业务字段',
  CONSTRAINT `ACT_FK_TASK_EXE` FOREIGN KEY (`execution_id_`) REFERENCES `t_wf_ru_execution` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `ACT_FK_TASK_PROCDEF` FOREIGN KEY (`proc_def_id_`) REFERENCES `t_wf_re_procdef` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `ACT_FK_TASK_PROCINST` FOREIGN KEY (`proc_inst_id_`) REFERENCES `t_wf_ru_execution` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_IDX_TASK_PARENT_TASK_ID` (`parent_task_id_`) USING BTREE,
  KEY `ACT_FK_TASK_EXE` (`execution_id_`) USING BTREE,
  KEY `ACT_FK_TASK_PROCINST` (`proc_inst_id_`) USING BTREE,
  KEY `ACT_FK_TASK_PROCDEF` (`proc_def_id_`) USING BTREE,
  KEY `ACT_IDX_TASK_ASSIGNEE__LIST_IDX` (`assignee_`, `create_time_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_ru_task
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_ru_variable
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_ru_variable`  (
  `id_` varchar(64) NOT NULL COMMENT '主键',
  `rev_` int(11) NULL DEFAULT NULL COMMENT '版本号',
  `type_` varchar(255) NULL DEFAULT NULL COMMENT '编码类型',
  `name_` varchar(255) NULL DEFAULT NULL COMMENT '变量名称',
  `execution_id_` varchar(64) NULL DEFAULT NULL COMMENT '执行实例ID,4.9版本改为与task_id_组成联合索引',
  `proc_inst_id_` varchar(64) NULL DEFAULT NULL COMMENT '流程实例Id',
  `task_id_` varchar(64) NULL DEFAULT NULL COMMENT '任务id',
  `bytearray_id_` varchar(64) NULL DEFAULT NULL COMMENT '字节组ID',
  `double_` double NULL DEFAULT NULL COMMENT '存储变量类型为Double',
  `long_` bigint(20) NULL DEFAULT NULL COMMENT '存储变量类型为long',
  `text_` varchar(4000) NULL DEFAULT NULL COMMENT '存储变量值类型为String\n\n            如此处存储持久化对象时，值jpa对象的class',
  `text2_` varchar(4000) NULL DEFAULT NULL COMMENT '此处存储的是JPA持久化对象时，才会有值。此值为对象ID',
  CONSTRAINT `ACT_FK_VAR_BYTEARRAY` FOREIGN KEY (`bytearray_id_`) REFERENCES `t_wf_ge_bytearray` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `ACT_FK_VAR_EXE` FOREIGN KEY (`execution_id_`) REFERENCES `t_wf_ru_execution` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `ACT_FK_VAR_PROCINST` FOREIGN KEY (`proc_inst_id_`) REFERENCES `t_wf_ru_execution` (`id_`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  PRIMARY KEY (`id_`) USING BTREE,
  KEY `ACT_IDX_VARIABLE_TASK_ID` (`task_id_`) USING BTREE,
  KEY `ACT_FK_VAR_EXE` (`execution_id_`) USING BTREE,
  KEY `ACT_FK_VAR_PROCINST` (`proc_inst_id_`) USING BTREE,
  KEY `ACT_FK_VAR_BYTEARRAY` (`bytearray_id_`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_ru_variable
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_sys_log
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_sys_log`  (
  `id` varchar(50) NOT NULL COMMENT '主键id',
  `type` varchar(10) NOT NULL COMMENT '日志类型 info信息，warn警告，error异常',
  `url` varchar(500) NULL DEFAULT NULL COMMENT '接口地址',
  `system_name` varchar(20) NULL DEFAULT NULL COMMENT '系统名称',
  `user_id` varchar(50) NULL DEFAULT NULL COMMENT '访问人ID',
  `msg` varchar(500) NOT NULL COMMENT '信息',
  `ex_msg` text NULL COMMENT '附加信息',
  `create_time` datetime(0) NOT NULL DEFAULT current_timestamp(0) COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_sys_log
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_type
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_type`  (
  `type_id` varchar(50) NOT NULL COMMENT '类型ID',
  `type_name` varchar(50) NOT NULL COMMENT '类型名称',
  `type_parent_id` varchar(50) NULL DEFAULT NULL COMMENT '父类型ID',
  `type_sort` decimal(10, 0) NULL DEFAULT NULL COMMENT '排序',
  `app_key` varchar(50) NOT NULL COMMENT '应用key',
  `type_remark` varchar(500) NULL DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`type_id`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_type
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_user
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_user`  (
  `user_id` varchar(50) NOT NULL COMMENT '用户ID',
  `user_code` varchar(50) NOT NULL COMMENT '用户编码',
  `user_name` varchar(50) NOT NULL COMMENT '用户姓名',
  `user_sex` varchar(2) NULL DEFAULT NULL COMMENT '用户性别',
  `user_age` int(11) NULL DEFAULT NULL COMMENT '用户年龄',
  `company_id` varchar(50) NOT NULL COMMENT '直属单位编码',
  `org_id` varchar(50) NOT NULL COMMENT '所属单位编码',
  `user_mobile` varchar(50) NULL DEFAULT NULL COMMENT '用户手机号码',
  `user_mail` varchar(100) NULL DEFAULT NULL COMMENT '用户邮箱',
  `user_work_address` varchar(500) NULL DEFAULT NULL COMMENT '用户工作地址',
  `user_work_phone` varchar(100) NULL DEFAULT NULL COMMENT '用户工作手机号',
  `user_home_addree` varchar(500) NULL DEFAULT NULL COMMENT '用户家庭地址',
  `user_home_phone` varchar(100) NULL DEFAULT NULL COMMENT '用户家庭手机号',
  `position_id` varchar(50) NULL DEFAULT NULL COMMENT '主要岗位',
  `plurality_position_id` varchar(50) NULL DEFAULT NULL COMMENT '兼职岗位',
  `title_id` varchar(1000) NULL DEFAULT NULL COMMENT '主要职务',
  `plurality_title_id` varchar(100) NULL DEFAULT NULL COMMENT '兼职职务',
  `user_type` varchar(10) NULL DEFAULT NULL COMMENT '用户类别',
  `user_status` varchar(10) NOT NULL COMMENT '用户状态',
  `user_sort` int(11) NULL DEFAULT NULL COMMENT '用户排序',
  `user_pwd` varchar(20) NULL DEFAULT '123456' COMMENT '用户密码',
  `user_create_time` date NULL DEFAULT NULL COMMENT '用户创建时间',
  `user_update_time` timestamp(0) NOT NULL DEFAULT current_timestamp(0) COMMENT '用户修改时间',
  `user_creator` varchar(30) NULL DEFAULT NULL COMMENT '用户创建者',
  `remark` varchar(500) NULL DEFAULT NULL COMMENT '备注',
  `dept_id` varchar(50) NULL DEFAULT NULL COMMENT '直属单位编码',
  `company_name` varchar(100) NULL DEFAULT NULL COMMENT '直属单位名称',
  `dept_name` varchar(100) NULL DEFAULT NULL COMMENT '直属单位名称',
  `org_name` varchar(100) NULL DEFAULT NULL COMMENT '所属单位名称',
  PRIMARY KEY (`user_id`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_user
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_user2role
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_user2role`  (
  `role_id` varchar(50) NOT NULL COMMENT '角色ID',
  `user_id` varchar(500) NOT NULL COMMENT '用户ID',
  `remark` varchar(500) NULL DEFAULT NULL COMMENT '备注',
  `user_code` varchar(500) NULL DEFAULT NULL COMMENT '用户编码',
  `user_name` varchar(500) NULL DEFAULT NULL COMMENT '用户名称',
  `org_id` varchar(50) NOT NULL COMMENT '组织ID',
  `org_name` varchar(500) NULL DEFAULT NULL COMMENT '组织名称',
  `sort` int(11) NULL DEFAULT NULL COMMENT '排序',
  `create_user_id` varchar(100) NULL DEFAULT NULL COMMENT '创建人ID',
  `create_user_name` varchar(100) NULL DEFAULT NULL COMMENT '创建人名称',
  `create_time` datetime(0) NULL DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`role_id`, `user_id`, `org_id`) USING BTREE
) ENGINE = InnoDB;

-- ----------------------------
-- Records of t_wf_user2role
-- ----------------------------

-- ----------------------------
-- Table structure for t_wf_countersign_info
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_countersign_info`  (
  `id` varchar(50) NOT NULL COMMENT '主键ID',
  `proc_inst_id` varchar(50) NULL DEFAULT NULL COMMENT '流程实例ID',
  `task_id` varchar(100) NULL DEFAULT NULL COMMENT '任务ID',
  `task_def_key` varchar(100) NULL DEFAULT NULL COMMENT '任务定义KEY',
  `countersign_auditor` varchar(100) NULL DEFAULT NULL COMMENT '加签的审核员',
  `countersign_auditor_name` varchar(100) NULL DEFAULT NULL COMMENT '加签的审核员名称',
  `countersign_by` varchar(100) NULL DEFAULT NULL COMMENT '加签人',
  `countersign_by_name` varchar(100) NULL DEFAULT NULL COMMENT '加签人名称',
  `reason` varchar(1000) NULL DEFAULT NULL COMMENT '加签原因',
  `batch` decimal(10, 0) NULL DEFAULT NULL COMMENT '批次',
  `create_time` datetime(0) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_t_wf_countersign_info_inst_id_def_key` (`proc_inst_id`, `task_def_key`)
) ENGINE = InnoDB;

-- ----------------------------
-- Table structure for t_wf_transfer_info
-- ----------------------------
CREATE TABLE IF NOT EXISTS `t_wf_transfer_info` (
  `id` varchar(50) NOT NULL COMMENT '主键ID',
  `proc_inst_id` varchar(50) NULL DEFAULT NULL COMMENT '流程实例ID',
  `task_id` varchar(100) NULL DEFAULT NULL COMMENT '任务ID',
  `task_def_key` varchar(100) NULL DEFAULT NULL COMMENT '任务定义KEY',
  `transfer_auditor` varchar(100) NULL DEFAULT NULL COMMENT '转审的审核员',
  `transfer_auditor_name` varchar(100) NULL DEFAULT NULL COMMENT '转审的审核员名称',
  `transfer_by` varchar(100) NULL DEFAULT NULL COMMENT '转审人',
  `transfer_by_name` varchar(100) NULL DEFAULT NULL COMMENT '转审人名称',
  `reason` varchar(1000) NULL DEFAULT NULL COMMENT '转审原因',
  `batch` decimal(10,0) NULL DEFAULT NULL COMMENT '批次',
  `create_time` datetime(0) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_t_wf_transfer_info_inst_id_def_key` (`proc_inst_id`, `task_def_key`)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS `t_wf_outbox` (
  `f_id` varchar(50) NOT NULL COMMENT '主键ID',
  `f_topic` varchar(128) NOT NULL COMMENT '消息topic',
  `f_message` longtext NOT NULL COMMENT '消息内容,json格式字符串',
  `f_create_time` datetime(0) NOT NULL COMMENT '消息创建时间',
  PRIMARY KEY (`f_id`),
  KEY `idx_t_wf_outbox_f_create_time` (`f_create_time`) USING BTREE
) ENGINE=InnoDB COMMENT='outbox信息表';

CREATE TABLE IF NOT EXISTS `t_wf_internal_group` (
  `f_id` varchar(40) NOT NULL COMMENT '主键id',
  `f_apply_id` varchar(50) NOT NULL COMMENT '申请id',
  `f_apply_user_id` varchar(40) NOT NULL COMMENT '申请人id',
  `f_group_id` varchar(40) NOT NULL COMMENT '内部组id',
  `f_expired_at` bigint DEFAULT -1 COMMENT '创内部组过期时间',
  `f_created_at` bigint DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`f_id`),
  KEY idx_t_wf_internal_group_apply_id (f_apply_id),
  KEY idx_t_wf_internal_group_expired_at (f_expired_at)
) ENGINE=InnoDB COMMENT='内部组账号信息';

CREATE TABLE IF NOT EXISTS t_wf_doc_audit_message (
  id VARCHAR(64) NOT NULL COMMENT '主键ID',
  proc_inst_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '流程实例ID',
  chan VARCHAR(255) NOT NULL DEFAULT '' COMMENT '消息 channel',
  payload MEDIUMTEXT NULL DEFAULT NULL COMMENT '消息 payload',
  ext_message_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '消息中心消息ID',
  PRIMARY KEY (id),
  KEY idx_t_wf_doc_audit_message_proc_inst_id (proc_inst_id)
) ENGINE=InnoDB COMMENT='审核消息';

CREATE TABLE IF NOT EXISTS t_wf_doc_audit_message_receiver (
  id VARCHAR(64) NOT NULL COMMENT '主键ID',
  message_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '消息ID',
  receiver_id  VARCHAR(255) NOT NULL DEFAULT '' COMMENT '接收者ID',
  handler_id VARCHAR(255) NOT NULL DEFAULT '' COMMENT '处理者ID',
  audit_status VARCHAR(10) NOT NULL DEFAULT '' COMMENT '处理状态',
  PRIMARY KEY (id),
  KEY idx_t_wf_doc_audit_message_receiver_message_id (message_id),
  KEY idx_t_wf_doc_audit_message_receiver_receiver_id (receiver_id),
  KEY idx_t_wf_doc_audit_message_receiver_handler_id (handler_id)
) ENGINE=InnoDB COMMENT='审核消息接收者';

CREATE TABLE IF NOT EXISTS `t_wf_doc_share_strategy_config` (
  `f_id` varchar(40) NOT NULL COMMENT '主键id',
  `f_proc_def_id` varchar(300) NOT NULL COMMENT '流程定义ID',
  `f_act_def_id` varchar(100) NOT NULL COMMENT '流程环节ID',
  `f_name` varchar(64) NOT NULL COMMENT '字段名称',
  `f_value` varchar(64) NOT NULL COMMENT '字段值',
  PRIMARY KEY (`f_id`),
  KEY idx_t_wf_doc_share_strategy_config_proc_act_def_id (f_proc_def_id, f_act_def_id),
  KEY idx_t_wf_doc_share_strategy_config_proc_def_id_name (f_proc_def_id, f_name),
  KEY idx_t_wf_doc_share_strategy_config_name (f_name)
) ENGINE=InnoDB COMMENT='审核流程高级配置表';

CREATE TABLE IF NOT EXISTS `t_wf_doc_audit_sendback_message` (
  `f_id` varchar(64) NOT NULL COMMENT '主键ID',
  `f_proc_inst_id` varchar(64) NOT NULL DEFAULT '' COMMENT '流程实例ID',
  `f_message_id` varchar(64) NOT NULL DEFAULT '' COMMENT '消息中心消息ID',
  `f_created_at` datetime NOT NULL COMMENT '创建时间',
  `f_updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`f_id`),
  KEY `idx_t_wf_doc_audit_sendback_message_proc_inst_id` (f_proc_inst_id)
) ENGINE=InnoDB COMMENT='审核退回消息';

CREATE TABLE IF NOT EXISTS `t_wf_inbox` (
  `f_id` varchar(50) NOT NULL COMMENT '主键ID',
  `f_topic` varchar(128) NOT NULL COMMENT '消息topic',
  `f_message` longtext NOT NULL COMMENT '消息内容,json格式字符串',
  `f_create_time` datetime(0) NOT NULL COMMENT '消息创建时间',
  PRIMARY KEY (`f_id`),
  KEY `idx_t_wf_inbox_f_create_time` (`f_create_time`) USING BTREE
) ENGINE=InnoDB COMMENT='inbox信息表';

-- SET FOREIGN_KEY_CHECKS = 1;

-- ----------------------------
-- Records of t_wf_ge_property
-- ----------------------------
INSERT INTO `t_wf_ge_property` SELECT 'next.dbid', '1', 1 FROM DUAL WHERE NOT EXISTS(SELECT `value_`, `rev_` FROM `t_wf_ge_property` WHERE `name_`='next.dbid');
INSERT INTO `t_wf_ge_property` SELECT 'schema.history', 'create(7.0.4.7.0)', 1 FROM DUAL WHERE NOT EXISTS(SELECT `value_`, `rev_` FROM `t_wf_ge_property` WHERE `name_`='schema.history');
INSERT INTO `t_wf_ge_property` SELECT 'schema.version', '7.0.4.7.0', 1 FROM DUAL WHERE NOT EXISTS(SELECT `value_`, `rev_` FROM `t_wf_ge_property` WHERE `name_`='schema.version');

-- ----------------------------
-- Records of t_wf_dict
-- ----------------------------
INSERT INTO `t_wf_dict` SELECT 'dc10b959-1bb4-4182-baf7-ab16d9409989', 'free_audit_secret_level', NULL, '6', NULL, 'Y', NULL, NULL, NULL, NULL, 'as_workflow', '' FROM DUAL WHERE NOT EXISTS(SELECT `dict_code`, `dict_parent_id`, `dict_name`, `sort`, `status`, `creator_id`, `create_date`, `updator_id`, `update_date`, `app_id`, `dict_value` FROM `t_wf_dict` WHERE `dict_code`='free_audit_secret_level');
INSERT INTO `t_wf_dict` SELECT '3d89e740-df13-4212-92a0-29e674da0e17', 'self_dept_free_audit', NULL, 'Y', NULL, 'Y', NULL, NULL, NULL, NULL, 'as_workflow', '' FROM DUAL WHERE NOT EXISTS(SELECT `dict_code`, `dict_parent_id`, `dict_name`, `sort`, `status`, `creator_id`, `create_date`, `updator_id`, `update_date`, `app_id`, `dict_value` FROM `t_wf_dict` WHERE `dict_code`='self_dept_free_audit');
INSERT INTO `t_wf_dict` SELECT 'bfc1c6cd-1bda-4057-992e-feb624915b0e', 'free_audit_secret_level_enum', NULL, '{\"非密\": 5,\"内部\": 6, \"秘密\": 7,\"机密\": 8}', NULL, 'Y', NULL, NULL, NULL, NULL, 'as_workflow', '' FROM DUAL WHERE NOT EXISTS(SELECT `dict_code`, `dict_parent_id`, `dict_name`, `sort`, `status`, `creator_id`, `create_date`, `updator_id`, `update_date`, `app_id`, `dict_value` FROM `t_wf_dict` WHERE `dict_code`='free_audit_secret_level_enum');
INSERT INTO `t_wf_dict` SELECT 'eaa1b91c-c53c-4113-a066-3e2690c36eae', 'anonymity_auto_audit_switch', NULL, 'n', NULL, 'Y', NULL, NULL, NULL, NULL, 'as_workflow', NULL FROM DUAL WHERE NOT EXISTS(SELECT `dict_code`, `dict_parent_id`, `dict_name`, `sort`, `status`, `creator_id`, `create_date`, `updator_id`, `update_date`, `app_id`, `dict_value` FROM `t_wf_dict` WHERE `dict_code`='anonymity_auto_audit_switch');
INSERT INTO `t_wf_dict` SELECT '706601cd-948b-4e4b-9265-3ada83d23326', 'rename_auto_audit_switch', NULL, 'n', NULL, 'Y', NULL, NULL, NULL, NULL, 'as_workflow', NULL FROM DUAL WHERE NOT EXISTS(SELECT `dict_code`, `dict_parent_id`, `dict_name`, `sort`, `status`, `creator_id`, `create_date`, `updator_id`, `update_date`, `app_id`, `dict_value` FROM `t_wf_dict` WHERE `dict_code`='rename_auto_audit_switch');
-- Source: ontology/ontology-manager/migrations/mariadb/0.1.0/pre/init.sql
USE adp;

-- 业务知识网络
CREATE TABLE IF NOT EXISTS t_knowledge_network (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_name VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络名称',
  f_tags VARCHAR(255) DEFAULT NULL COMMENT '标签',
  f_comment VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_icon VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图标',
  f_color VARCHAR(40) NOT NULL DEFAULT '' COMMENT '颜色',
  f_detail MEDIUMTEXT DEFAULT NULL COMMENT '概览',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_business_domain VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务域',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_id,f_branch),
  UNIQUE KEY uk_kn_name (f_name,f_branch)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '业务知识网络';

-- 对象类
CREATE TABLE IF NOT EXISTS t_object_type (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象类id',
  f_name VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象类名称',
  f_tags VARCHAR(255) DEFAULT NULL COMMENT '标签',
  f_comment VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_icon VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图标',
  f_color VARCHAR(40) NOT NULL DEFAULT '' COMMENT '颜色',
  f_detail MEDIUMTEXT DEFAULT NULL COMMENT '概览',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_data_source VARCHAR(255) NOT NULL COMMENT '数据来源，当前只有视图',
  f_data_properties LONGTEXT DEFAULT NULL COMMENT '数据属性',
  f_logic_properties MEDIUMTEXT DEFAULT NULL COMMENT '逻辑属性',
  f_primary_keys VARCHAR(8192) DEFAULT NULL COMMENT '对象类主键',
  f_display_key VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象实例的显示属性',
  f_incremental_key VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象类增量键',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_kn_id,f_branch,f_id),
  UNIQUE KEY uk_object_type_name (f_kn_id,f_branch,f_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '对象类';

-- 对象类状态
CREATE TABLE IF NOT EXISTS t_object_type_status (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象类id',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_incremental_key VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象类增量键',
  f_incremental_value VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象类当前增量值',
  f_index VARCHAR(255) NOT NULL DEFAULT '' COMMENT '索引名称',
  f_index_available BOOLEAN NOT NULL DEFAULT 0 COMMENT '索引是否可用',
  f_doc_count BIGINT(20) NOT NULL DEFAULT 0 COMMENT '文档数量',
  f_storage_size BIGINT(20) NOT NULL DEFAULT 0 COMMENT '存储大小',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_kn_id,f_branch,f_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '对象类状态';

-- 关系类
CREATE TABLE IF NOT EXISTS t_relation_type (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '关系类id',
  f_name VARCHAR(40) NOT NULL DEFAULT '' COMMENT '关系类名称',
  f_tags VARCHAR(255) DEFAULT NULL COMMENT '标签',
  f_comment VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_icon VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图标',
  f_color VARCHAR(40) NOT NULL DEFAULT '' COMMENT '颜色',
  f_detail MEDIUMTEXT DEFAULT NULL COMMENT '概览',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_source_object_type_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '起点对象类',
  f_target_object_type_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '终点对象类',
  f_type VARCHAR(40) NOT NULL DEFAULT '' COMMENT '关联类型',
  f_mapping_rules TEXT DEFAULT NULL COMMENT '关联规则',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_kn_id,f_branch,f_id),
  UNIQUE KEY uk_relation_type_name (f_kn_id,f_branch,f_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '关系类';

-- 行动类
CREATE TABLE IF NOT EXISTS t_action_type (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '行动类id',
  f_name VARCHAR(40) NOT NULL DEFAULT '' COMMENT '行动类名称',
  f_tags VARCHAR(255) DEFAULT NULL COMMENT '标签',
  f_comment VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_icon VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图标',
  f_color VARCHAR(40) NOT NULL DEFAULT '' COMMENT '颜色',
  f_detail MEDIUMTEXT DEFAULT NULL COMMENT '概览',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_action_type VARCHAR(40) NOT NULL DEFAULT '' COMMENT '行动类型',
  f_object_type_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象类',
  f_condition TEXT DEFAULT NULL COMMENT '行动条件',
  f_affect TEXT DEFAULT NULL COMMENT '行动影响',
  f_action_source VARCHAR(255) NOT NULL COMMENT '行动资源',
  f_parameters TEXT DEFAULT NULL COMMENT '行动参数',
  f_schedule VARCHAR(255) DEFAULT NULL COMMENT '行动监听',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_kn_id,f_branch,f_id),
  UNIQUE KEY uk_action_type_name (f_kn_id,f_branch,f_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '行动类';


-- 任务管理
CREATE TABLE IF NOT EXISTS t_kn_job (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '任务id',
  f_name VARCHAR(40) NOT NULL DEFAULT '' COMMENT '任务名称',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_job_type VARCHAR(40) NOT NULL DEFAULT '' COMMENT '任务类型',
  f_job_concept_config MEDIUMTEXT DEFAULT NULL COMMENT '任务概念配置',
  f_state VARCHAR(40) NOT NULL DEFAULT '' COMMENT '状态',
  f_state_detail TEXT DEFAULT NULL COMMENT '状态详情',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_finish_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '完成时间',
  f_time_cost BIGINT(20) NOT NULL DEFAULT 0 COMMENT '耗时',
  PRIMARY KEY (f_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '任务';


-- 子任务管理
CREATE TABLE IF NOT EXISTS t_kn_task (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '子任务id',
  f_name VARCHAR(40) NOT NULL DEFAULT '' COMMENT '子任务名称',
  f_job_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '任务id',
  f_concept_type VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念类型',
  f_concept_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念id',
  f_index VARCHAR(255) NOT NULL DEFAULT '' COMMENT '索引名称',
  f_doc_count BIGINT(20) NOT NULL DEFAULT 0 COMMENT '文档数量',
  f_state VARCHAR(40) NOT NULL DEFAULT '' COMMENT '状态',
  f_state_detail TEXT DEFAULT NULL COMMENT '状态详情',
  f_start_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '开始时间',
  f_finish_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '完成时间',
  f_time_cost BIGINT(20) NOT NULL DEFAULT 0 COMMENT '耗时',
  PRIMARY KEY (f_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '子任务';

-- 概念分组
CREATE TABLE IF NOT EXISTS t_concept_group (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念分组id',
  f_name VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念分组名称',
  f_tags VARCHAR(255) DEFAULT NULL COMMENT '标签',
  f_comment VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_icon VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图标',
  f_color VARCHAR(40) NOT NULL DEFAULT '' COMMENT '颜色',
  f_detail MEDIUMTEXT DEFAULT NULL COMMENT '概览',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_kn_id,f_branch,f_id),
  UNIQUE KEY uk_concept_group_name (f_kn_id,f_branch,f_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '概念分组';

-- 分组与概念对应表
CREATE TABLE IF NOT EXISTS t_concept_group_relation (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '主键id',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_group_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念分组id',
  f_concept_type VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念类型',
  f_concept_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '概念id',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (f_id),
  UNIQUE KEY uk_concept_group_relation (f_kn_id,f_branch,f_group_id,f_concept_type,f_concept_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '分组与概念对应表';

-- Source: vega/data-connection/migrations/mariadb/0.1.0/pre/init.sql
USE adp;

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
  `f_status` tinyint NOT NULL DEFAULT 3 COMMENT '任务状态：0 成功1 失败2 进行中 3 初始化',
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
-- Source: vega/mdl-data-model/migrations/mariadb/0.1.0/pre/init.sql
USE adp;

-- 指标模型
CREATE TABLE IF NOT EXISTS t_metric_model (
  f_model_id varchar(40) NOT NULL DEFAULT '' COMMENT '指标模型 id',
  f_model_name varchar(40) NOT NULL COMMENT '指标模型名称',
  f_tags varchar(255) DEFAULT NULL COMMENT '标签',
  f_comment varchar(255) DEFAULT NULL COMMENT '备注',
  f_catalog_id varchar(40) NOT NULL DEFAULT '' COMMENT '编目id',
  f_catalog_content text DEFAULT NULL COMMENT '编目内容',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_measure_name varchar(50) NOT NULL DEFAULT '' COMMENT '度量名称',
  f_metric_type varchar(20) NOT NULL COMMENT '指标类型',
  f_data_source varchar(255) NOT NULL COMMENT '数据源',
  f_query_type varchar(20) NOT NULL COMMENT '指标查询语言',
  f_formula text NOT NULL COMMENT '计算公式',
  f_formula_config text DEFAULT NULL COMMENT '计算公式配置化',
  f_analysis_dimessions varchar(8192) DEFAULT NULL COMMENT '分析维度',
  f_order_by_fields varchar(4096) DEFAULT NULL COMMENT '排序字段',
  f_having_condition varchar(2048) DEFAULT NULL COMMENT '值过滤',
  f_date_field varchar(255) DEFAULT NULL COMMENT '时间字段',
  f_measure_field varchar(255) NOT NULL COMMENT '度量字段',
  f_unit_type varchar(40) NOT NULL COMMENT '单位类型',
  f_unit varchar(20) NOT NULL COMMENT '度量单位',
  f_group_id varchar(40) NOT NULL DEFAULT '' COMMENT '指标模型分组 id',
  f_builtin tinyint(2) DEFAULT 0 COMMENT '内置模型标识: 0 非内置, 1 内置',
  f_calendar_interval tinyint(2) DEFAULT 0 COMMENT '是否日历间隔。0: 非日历间隔; 1: 日历间隔',
  PRIMARY KEY (f_model_id),
  UNIQUE KEY uk_model_name (f_group_id, f_model_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '指标模型';

-- 指标模型分组
CREATE TABLE IF NOT EXISTS t_metric_model_group (
  f_group_id varchar(40) NOT NULL DEFAULT '' COMMENT '指标模型分组 id',
  f_group_name varchar(40) NOT NULL COMMENT '指标模型分组名称',
  f_comment varchar(255) NOT NULL DEFAULT '' COMMENT '指标模型分组备注',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_builtin tinyint(2) NOT NULL DEFAULT 0 COMMENT '内置分组标识: 0 非内置, 1 内置',
  PRIMARY KEY (f_group_id),
  UNIQUE KEY uk_f_group_name (f_group_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '指标模型分组';


-- 指标模型持久化任务
CREATE TABLE IF NOT EXISTS t_metric_model_task(
  f_task_id varchar(40) NOT NULL DEFAULT '' COMMENT '任务 id',
  f_task_name varchar(40) NOT NULL COMMENT '任务名称',
  f_comment varchar(255) NOT NULL DEFAULT '' COMMENT '任务备注',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_module_type varchar(20) NOT NULL DEFAULT '' COMMENT '模块类型',
  f_model_id varchar(40) NOT NULL COMMENT '指标模型 id',
  f_schedule varchar(255) NOT NULL COMMENT '执行频率',
  f_variables text DEFAULT NULL COMMENT '变量过滤',
  f_time_windows varchar(1024) DEFAULT NULL COMMENT '时间窗口',
  f_steps varchar(255) NOT NULL DEFAULT '[]' COMMENT '持久化步长',
  f_plan_time bigint(20) NOT NULL DEFAULT 0 COMMENT '计划时间',
  f_index_base varchar(40) NOT NULL COMMENT '索引库类型',
  f_retrace_duration varchar(20) DEFAULT NULL COMMENT '追溯时长',
  f_schedule_sync_status tinyint(2) NOT NULL COMMENT '任务的同步状态。3: 完成',
  f_execute_status tinyint(2) DEFAULT 0 COMMENT '任务的执行状态。4: 执行成功; 5: 执行失败',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建任务的用户id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  PRIMARY KEY (f_task_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '指标持久化任务';


-- 指标索引的静态表，记录指标类索引库的分割时间点，便于升级到__tsid后，指标模型查询的兼容
CREATE TABLE IF NOT EXISTS t_static_metric_index (
  f_id int(11) AUTO_INCREMENT COMMENT '唯一id编号',
  f_base_type varchar(40) NOT NULL COMMENT '指标索引库类型',
  f_split_time datetime DEFAULT current_timestamp() COMMENT '索引库的时间分割',
  PRIMARY KEY (f_id),
  UNIQUE KEY uk_f_index_base_type (f_base_type)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '指标索引库tsid的时间分割静态表';


-- 事件模型聚合规则
CREATE TABLE if not exists t_event_model_aggregate_rules (
  f_aggregate_rule_id varchar(40) NOT NULL DEFAULT '' COMMENT '聚合规则id',
  f_aggregate_rule_type varchar(40) NOT NULL COMMENT '聚合规则类型',
  f_aggregate_algo varchar(900) NOT NULL COMMENT '聚合算法',
  f_rule_priority int(11) NOT NULL COMMENT '规则优先级',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_group_fields varchar(255) DEFAULT '[]' COMMENT '分组字段',
  f_aggregate_analysis_algo varchar(1024) DEFAULT '{}' COMMENT "分析算法",
  PRIMARY KEY (f_aggregate_rule_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin;


-- 事件模型
CREATE TABLE if not exists t_event_models (
  f_event_model_id varchar(40) NOT NULL DEFAULT '' COMMENT '事件模型id',
  f_event_model_name varchar(255) NOT NULL COMMENT '事件模型名称',
  f_event_model_group_name varchar(40) NOT NULL DEFAULT '' COMMENT '事件模型分组名称',
  f_event_model_type varchar(40) NOT NULL COMMENT '事件模型类型',
  f_event_model_tags varchar(255) NOT NULL COMMENT '事件模型标签',
  f_event_model_comment varchar(255) DEFAULT NULL COMMENT '事件模型说明',
  f_data_source_type varchar(40) NOT NULL COMMENT '数据源类型',
  f_data_source varchar(900) DEFAULT NULL COMMENT '数据源对象id',
  f_detect_rule_id varchar(40) NOT NULL COMMENT '检测规则id',
  f_aggregate_rule_id varchar(40) NOT NULL COMMENT '聚合规则id',
  f_default_time_window varchar(40) NOT NULL COMMENT '默认时间窗口',
  f_is_active tinyint(2) DEFAULT 0 COMMENT '是否是定期执行模式',
  f_enable_subscribe tinyint(2) DEFAULT 0 COMMENT '是否是实时订阅模式',
  f_status tinyint(2) DEFAULT 0 COMMENT '是否启用',
  f_downstream_dependent_model varchar(1024) DEFAULT '' COMMENT '依赖模型',
  f_is_custom tinyint(2) NOT NULL COMMENT '是否个性化',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_event_model_id),
  UNIQUE KEY uk_f_model_name (f_event_model_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin;


-- 事件模型检测规则
CREATE TABLE if not exists t_event_model_detect_rules (
  f_detect_rule_id varchar(40) NOT NULL DEFAULT '' COMMENT '检测规则id',
  f_detect_rule_type varchar(40) NOT NULL COMMENT '检测规则类型',
  f_formula varchar(2014) DEFAULT NULL COMMENT '计算公式',
  f_detect_algo varchar(40) DEFAULT NULL  COMMENT '检测算法',
  f_detect_analysis_algo varchar(1024) DEFAULT '{}' COMMENT '分析算法',
  f_rule_priority int(11) NOT NULL COMMENT "规则优先级",
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_detect_rule_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin;


-- 事件模型持久化任务
CREATE TABLE IF NOT EXISTS t_event_model_task (
  f_task_id varchar(40) NOT NULL DEFAULT '' COMMENT '唯一id编号',
  f_model_id varchar(40) NOT NULL COMMENT '事件模型 id',
  f_storage_config varchar(255) NOT NULL COMMENT '存储配置',
  f_schedule varchar(255) NOT NULL COMMENT '执行频率',
  f_dispatch_config varchar(255) NOT NULL COMMENT '调度配置',
  f_execute_parameter varchar(255) NOT NULL COMMENT '执行参数',
  f_task_status tinyint(2) NOT NULL COMMENT '最近一次任务执行状态。4: 执行成功; 5: 执行失败',
  f_error_details varchar(2048) NOT NULL COMMENT '任务执行失败原因',
  f_status_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '执行状态更新时间',
  f_schedule_sync_status tinyint(2) NOT NULL COMMENT '任务的同步状态。3: 完成',
  f_downstream_dependent_task varchar(1024) DEFAULT '' COMMENT '依赖任务',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_task_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '事件模型持久化任务';


-- 事件模型异步任务记录
CREATE TABLE IF NOT EXISTS t_event_model_task_execution_records (
  f_run_id bigint(20) unsigned NOT NULL COMMENT '运行id',
  f_run_type varchar(40) NOT NULL COMMENT '任务类型',
  f_execute_parameter varchar(2048) NOT NULL COMMENT '执行参数',
  f_status varchar(40) DEFAULT '0' COMMENT '状态',
  f_error_details varchar(1024) NOT NULL COMMENT '错误原因',
  f_update_time datetime NOT NULL COMMENT '更新时间',
  f_create_time datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (f_run_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin;


-- 数据视图
CREATE TABLE IF NOT EXISTS t_data_view (
  f_view_id varchar(40) NOT NULL DEFAULT '' COMMENT '数据视图 id',
  f_view_name varchar(255) NOT NULL COMMENT '数据视图名称',
  f_technical_name varchar(255) NOT NULL DEFAULT '' COMMENT '技术名称',
  f_group_id varchar(40) NOT NULL DEFAULT '' COMMENT '数据视图分组 id',
  f_type varchar(10) NOT NULL DEFAULT '' COMMENT '视图类型',
  f_query_type varchar(10) NOT NULL DEFAULT '' COMMENT '查询类型',
  f_builtin tinyint(2) DEFAULT 0 COMMENT '内置视图标识: 0 非内置, 1 内置',
  f_tags varchar(255) NOT NULL DEFAULT '' COMMENT '标签',
  f_comment varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_data_source_type varchar(20) NOT NULL DEFAULT '' COMMENT '数据源类型',
  f_data_source_id varchar(40) NOT NULL DEFAULT '' COMMENT '数据源 id',
  f_file_name varchar(128) NOT NULL DEFAULT '' COMMENT '文件名',
  f_excel_config text DEFAULT NULL COMMENT 'excel 配置',
  f_data_scope longtext DEFAULT NULL COMMENT '数据范围',
  f_fields longtext DEFAULT NULL COMMENT '字段列表',
  f_status varchar(20) NOT NULL DEFAULT '' COMMENT '状态',
  f_metadata_form_id varchar(40) NOT NULL DEFAULT '' COMMENT '元数据表单 id',
  f_primary_keys varchar(255) NOT NULL DEFAULT '' COMMENT '主键列表',
  f_sql longtext DEFAULT NULL COMMENT '生成视图sql',
  f_meta_table_name varchar(1024) NOT NULL DEFAULT '' COMMENT '元数据表名',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_delete_time bigint(20) NOT NULL DEFAULT 0 COMMENT '删除时间',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_updater varchar(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type varchar(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  f_data_source text DEFAULT NULL COMMENT '废弃, 数据视图数据来源',
  f_field_scope tinyint(2) NOT NULL DEFAULT '0' COMMENT '废弃, 字段范围: 0 部分字段, 1 全部字段',
  f_filters text DEFAULT NULL COMMENT '废弃, 过滤条件',
  f_open_streaming tinyint(2) NOT NULL DEFAULT 0 COMMENT '废弃, 是否开启视图实时订阅任务: 0 不开启, 1 开启',
  f_job_id varchar(40) NOT NULL DEFAULT '' COMMENT '废弃, 订阅任务 id',
  f_loggroup_filters longtext DEFAULT NULL COMMENT '废弃, 日志分组过滤条件',
  PRIMARY KEY (f_view_id),
  UNIQUE KEY uk_f_view_name (f_group_id, f_view_name, f_delete_time)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据视图';

-- 数据视图分组
CREATE TABLE IF NOT EXISTS t_data_view_group (
  f_group_id varchar(40) NOT NULL DEFAULT '' COMMENT '数据视图分组 id',
  f_group_name varchar(40) NOT NULL COMMENT '数据视图分组名称',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_delete_time bigint(20) NOT NULL DEFAULT 0 COMMENT '删除时间',
  f_builtin tinyint(2) NOT NULL DEFAULT 0 COMMENT '内置视图标识: 0 非内置, 1 内置',
  PRIMARY KEY (f_group_id),
  UNIQUE KEY uk_f_group_name (f_builtin, f_group_name, f_delete_time)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据视图分组';

-- 扫描记录
CREATE TABLE IF NOT EXISTS t_scan_record (
    f_record_id varchar(40) NOT NULL DEFAULT '' COMMENT '扫描记录 id',
    f_data_source_id varchar(40) NOT NULL COMMENT '数据源 id',
    f_scanner varchar(40) NOT NULL COMMENT '扫描器',
    f_scan_time bigint(20) NOT NULL DEFAULT 0 COMMENT '扫描时间',
    f_data_source_status varchar(20) NOT NULL DEFAULT '' COMMENT '数据源状态: available 可用 scanning 扫描中',
    f_metadata_task_id varchar(128)  DEFAULT NULL COMMENT '元数据采集平台任务id',
    PRIMARY KEY (f_record_id),
    UNIQUE KEY uk_scan_record (f_data_source_id, f_scanner)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据源扫描记录表';

-- 视图行列规则表
CREATE TABLE IF NOT EXISTS t_data_view_row_column_rule (
  f_rule_id varchar(40) NOT NULL DEFAULT '' COMMENT '视图行列规则 id',
  f_rule_name varchar(255) NOT NULL COMMENT '视图行列规则名称',
  f_view_id varchar(40) NOT NULL COMMENT '视图 id',
  f_tags varchar(255) NOT NULL DEFAULT '' COMMENT '标签',
  f_comment varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_fields longtext NOT NULL COMMENT '列',
  f_row_filters text NOT NULL COMMENT '行过滤规则',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_updater varchar(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type varchar(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  PRIMARY KEY (f_rule_id),
  UNIQUE KEY uk_f_rule_name (f_rule_name, f_view_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据视图行列规则';

-- 数据字典
CREATE TABLE IF NOT EXISTS t_data_dict (
  f_dict_id varchar(40) NOT NULL COMMENT '数据字典id',
  f_dict_name varchar(255) NOT NULL COMMENT '数据字典名称',
  f_tags varchar(255) NOT NULL COMMENT '标签',
  f_comment varchar(255) DEFAULT NULL COMMENT '备注',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_dict_type varchar(20) NOT NULL DEFAULT 'kv_dict' COMMENT '数据字典类型',
  f_dict_store varchar(255) NOT NULL COMMENT '数据字典的项存放的对应表名称',
  f_dimension varchar(1500) NOT NULL COMMENT '数据字典维度关系',
  f_unique_key tinyint(2) NOT NULL DEFAULT 1 COMMENT '是否唯一键',
  PRIMARY KEY (f_dict_id),
  UNIQUE KEY uk_dict_name (f_dict_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据字典信息';

-- 数据字典项
CREATE TABLE IF NOT EXISTS t_data_dict_item (
  f_item_id varchar(40) NOT NULL COMMENT '数据字典项id',
  f_dict_id varchar(40) NOT NULL COMMENT '数据字典id',
  f_item_key varchar(3000) NOT NULL COMMENT '数据字典项key值',
  f_item_value varchar(3000) NOT NULL COMMENT '数据字典项value值',
  f_comment varchar(255) COMMENT '数据字典项说明',
  PRIMARY KEY (f_item_id),
  KEY idx_dict_id (f_dict_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据字典项信息表';

-- 数据连接
CREATE TABLE IF NOT EXISTS t_data_connection (
  f_connection_id varchar(40) NOT NULL COMMENT '唯一id编号',
  f_connection_name varchar(40) NOT NULL COMMENT '数据连接名称',
  f_tags varchar(255) DEFAULT '' COMMENT '标签',
  f_comment varchar(255) DEFAULT '' COMMENT '数据连接备注',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_data_source_type varchar(40) NOT NULL COMMENT '数据源类型',
  f_config text NOT NULL COMMENT '详细配置',
  f_config_md5 varchar(32) DEFAULT '' COMMENT '详细配置的唯一标识符',
  PRIMARY KEY (f_connection_id),
  UNIQUE KEY uk_f_connection_name (f_connection_name),
  KEY idx_f_data_source_type (f_data_source_type),
  KEY idx_f_config_md5 (f_config_md5)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据连接';

-- 数据连接状态
CREATE TABLE IF NOT EXISTS t_data_connection_status (
  f_connection_id varchar(40) NOT NULL COMMENT '数据连接id',
  f_status varchar(5) NOT NULL COMMENT '连接状态',
  f_detection_time bigint(20) NOT NULL DEFAULT 0 COMMENT '检测时间',
  PRIMARY KEY (f_connection_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据连接状态';

-- 链路模型
CREATE TABLE IF NOT EXISTS t_trace_model (
  f_model_id varchar(40) NOT NULL COMMENT '唯一id编号',
  f_model_name varchar(40) NOT NULL COMMENT '链路模型名称',
  f_tags varchar(255) DEFAULT '' COMMENT '标签',
  f_comment varchar(255) DEFAULT '' COMMENT '链路模型备注',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_span_source_type varchar(40) NOT NULL COMMENT 'span数据来源类型',
  f_span_config text NOT NULL COMMENT 'span配置',
  f_enabled_related_log tinyint(2) NOT NULL COMMENT '是否开启配置span关联日志配置, 0表示否, 1表示是',
  f_related_log_source_type varchar(40) NOT NULL COMMENT 'span关联日志数据来源类型',
  f_related_log_config text NOT NULL COMMENT 'span关联日志配置',
  PRIMARY KEY (f_model_id),
  UNIQUE KEY uk_f_model_name (f_model_name),
  KEY idx_f_span_source_type (f_span_source_type)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '链路模型';

-- global data-model-job
CREATE TABLE IF NOT EXISTS t_data_model_job (
  f_job_id varchar(40) NOT NULL DEFAULT '' COMMENT '任务 id',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_job_type varchar(40) NOT NULL COMMENT '任务类型',
  f_job_config text COMMENT '任务配置',
  f_job_status varchar(20) NOT NULL COMMENT '任务状态: running 正常, error 异常',
  f_job_status_details text NOT NULL COMMENT '任务状态详情',
  PRIMARY KEY (f_job_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '全局任务表';


-- 目标模型
CREATE TABLE IF NOT EXISTS t_objective_model (
  f_model_id varchar(40) NOT NULL DEFAULT '' COMMENT '目标模型 id',
  f_model_name varchar(40) NOT NULL COMMENT '目标模型名称',
  f_tags varchar(255) NOT NULL DEFAULT '' COMMENT '标签',
  f_comment varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_objective_type varchar(20) NOT NULL COMMENT '目标类型',
  f_objective_config text NOT NULL COMMENT '目标配置',
  PRIMARY KEY (f_model_id),
  UNIQUE KEY uk_t_objective_model (f_model_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '目标模型';


-- --------------------------------------- 初始化数据 --------------------------------------------
-- 未分组
INSERT INTO t_data_view_group (
  f_group_id,
  f_group_name,
  f_create_time,
  f_update_time,
  f_builtin
)
SELECT '', '', 1733903782147, 1733903782147, 0
FROM DUAL
WHERE NOT EXISTS(
  SELECT f_group_id
  FROM t_data_view_group
  WHERE f_group_id = ''
);

-- 索引库
INSERT INTO t_data_view_group (
  f_group_id,
  f_group_name,
  f_create_time,
  f_update_time,
  f_builtin
)
SELECT '__index_base', 'index_base', 1733903782147, 1733903782147, 1
FROM DUAL
WHERE NOT EXISTS(
  SELECT f_group_id
  FROM t_data_view_group
  WHERE f_group_id = '__index_base'
);

-- 未分组
INSERT INTO t_metric_model_group (
  f_group_id,
  f_group_name,
  f_create_time,
  f_update_time,
  f_builtin
)
SELECT '', '', 1733903782147, 1733903782147, 0
FROM DUAL
WHERE NOT EXISTS(
  SELECT f_group_id
  FROM t_metric_model_group
  WHERE f_group_id = ''
);

-- Source: vega/vega-gateway/migrations/mariadb/0.1.0/pre/init.sql
USE adp;


CREATE TABLE IF NOT EXISTS `cache_table` (
  `id` char(36) NOT NULL COMMENT '主键',
  `catalog_name` char(36) NOT NULL COMMENT '对应的逻辑视图的catalog名称',
  `schema_name` char(36) NOT NULL COMMENT '对应的逻辑视图的schema名称',
  `table_name` char(36) NOT NULL COMMENT '对应的逻辑视图的table名称',
  `cts_sql` text DEFAULT NULL COMMENT '表的建表sql',
  `source_create_sql` text DEFAULT NULL COMMENT '样例数据查询sql',
  `current_view_original_text` text DEFAULT NULL COMMENT '最近一次的原始加密sql',
  `status` char(36) NOT NULL COMMENT '可用；异常；正在初始化',
  `mid_status` char(36) DEFAULT NULL COMMENT '在FSM任务的时候的中间状态',
  `deps` varchar(255) DEFAULT '' COMMENT '生成的结果缓存表的id用,分隔',
  `create_time` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '更新时间',
  `update_time` datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '更新时间',
  PRIMARY KEY (`id`)
);


CREATE TABLE IF NOT EXISTS `client_id` (
  `id` int NOT NULL COMMENT '主键id',
  `client_name` varchar(128) DEFAULT NULL COMMENT '客户端名称',
  `client_id` varchar(64) DEFAULT NULL COMMENT '客户端id',
  `client_secret` varchar(64) DEFAULT NULL COMMENT '客户端密码',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
);


CREATE TABLE IF NOT EXISTS `excel_column_type` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `catalog` varchar(256) NOT NULL COMMENT '数据源',
  `vdm_catalog` varchar(256) DEFAULT NULL COMMENT 'vdm数据源',
  `schema_name` varchar(256) NOT NULL COMMENT '库名',
  `table_name` varchar(512) NOT NULL COMMENT '表名',
  `column_name` varchar(128) NOT NULL COMMENT '列名',
  `column_comment` varchar(512) DEFAULT NULL COMMENT '列注释',
  `type` varchar(128) NOT NULL COMMENT '字段类型',
  `order_no` int NOT NULL COMMENT '列序号',
  `create_time` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '更新时间',
  PRIMARY KEY (`id`)
);


CREATE TABLE IF NOT EXISTS `excel_table_config` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `catalog` varchar(256) NOT NULL COMMENT '数据源',
  `vdm_catalog` varchar(256) DEFAULT NULL COMMENT 'vdm数据源',
  `schema_name` varchar(256) NOT NULL COMMENT '库名',
  `file_name` varchar(512) NOT NULL COMMENT 'excel文件名',
  `table_name` varchar(512) NOT NULL COMMENT '表名',
  `table_comment` varchar(512) DEFAULT NULL COMMENT '表注释',
  `sheet` varchar(128) DEFAULT NULL COMMENT 'sheet名称',
  `all_sheet` tinyint NOT NULL DEFAULT 0 COMMENT '是否加载所有sheet',
  `sheet_as_new_column` tinyint NOT NULL DEFAULT 0 COMMENT 'sheet是否作为列 1:是 0:否',
  `start_cell` varchar(32) DEFAULT NULL COMMENT '起始单元格',
  `end_cell` varchar(32) DEFAULT NULL COMMENT '结束单元格',
  `has_headers` tinyint NOT NULL DEFAULT 1 COMMENT '是否有表头  1：有； 0：没有',
  `create_time` timestamp NULL DEFAULT current_timestamp() COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT current_timestamp() COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `excel_table_config_vdm_table_uindex` (`catalog`,`table_name`)
);


CREATE TABLE IF NOT EXISTS `query_info` (
  `query_id` varchar(30) NOT NULL COMMENT 'query id',
  `result` longtext DEFAULT NULL COMMENT '查询结果集',
  `msg` varchar(500) DEFAULT NULL COMMENT '错误详情',
  `task_id` varchar(200) NOT NULL COMMENT '任务Id',
  `state` varchar(30) NOT NULL COMMENT '状态',
  `create_time` varchar(30) NOT NULL COMMENT '创建时间',
  `update_time` varchar(30) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`query_id`)
);


CREATE TABLE IF NOT EXISTS `task_info` (
  `task_id` varchar(200) NOT NULL COMMENT '主键taskid',
  `state` varchar(30) DEFAULT NULL COMMENT 'task状态',
  `query` longtext DEFAULT NULL,
  `create_time` varchar(30) DEFAULT NULL COMMENT '创建时间',
  `update_time` varchar(30) DEFAULT NULL COMMENT '修改时间',
  `topic` varchar(100) DEFAULT NULL COMMENT 'topic名称',
  `sub_task_id` varchar(200) NOT NULL COMMENT '子任务Id',
  `type` int NOT NULL DEFAULT 1 COMMENT '类型,0:异步查询,1:字段探查',
  `elapsed_time` varchar(30) NOT NULL COMMENT '总耗时',
  `update_count` text NOT NULL COMMENT '结果集大小,只针对insert into或create table as记录大小',
  `schedule_time` varchar(30) NOT NULL COMMENT '调度耗时',
  `queued_time` varchar(30) NOT NULL COMMENT '队列耗时',
  `cpu_time` varchar(30) NOT NULL COMMENT 'cpu耗时',
  PRIMARY KEY (`task_id`,`sub_task_id`)
);


-- Source: vega/vega-metadata/migrations/mariadb/0.1.0/pre/init.sql
USE adp;


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

