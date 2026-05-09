-- Source: dataflow/coderunner/migrations/mariadb/0.1.0/init.sql
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
-- Source: dataflow/flow-automation/migrations/mariadb/0.5.0/init.sql
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
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `uk_flow_storage_oss_id_object_key` (`f_oss_id`, `f_object_key`),
  KEY `idx_flow_storage_status` (`f_status`),
  KEY `idx_flow_storage_created_at` (`f_created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Dataflow存储文件表';

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
  PRIMARY KEY (`f_id`),
  KEY `idx_flow_file_dag_id` (`f_dag_id`),
  KEY `idx_flow_file_dag_instance_id` (`f_dag_instance_id`),
  KEY `idx_flow_file_storage_id` (`f_storage_id`),
  KEY `idx_flow_file_status` (`f_status`),
  KEY `idx_flow_file_expires_at` (`f_expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Dataflow业务文件表';

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
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `uk_flow_file_download_job_file_id` (`f_file_id`),
  KEY `idx_flow_file_download_job_status_retry` (`f_status`, `f_next_retry_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Dataflow文件下载任务表';

CREATE TABLE IF NOT EXISTS `t_flow_task_resume` (
  `f_id` BIGINT UNSIGNED NOT NULL COMMENT '主键ID',
  `f_task_instance_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '被阻塞的任务实例ID',
  `f_dag_instance_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属流程实例ID',
  `f_resource_type` VARCHAR(32) NOT NULL DEFAULT 'file' COMMENT '资源类型',
  `f_resource_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '资源ID，对文件场景即flow_file ID',
  `f_created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `f_updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `uk_flow_task_resume_task_instance_id` (`f_task_instance_id`),
  KEY `idx_flow_task_resume_resource` (`f_resource_type`, `f_resource_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Dataflow阻塞任务恢复表';

-- Source: dataflow/flow-stream-data-pipeline/migrations/mariadb/0.1.0/init.sql
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
