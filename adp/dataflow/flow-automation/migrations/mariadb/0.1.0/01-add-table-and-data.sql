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
