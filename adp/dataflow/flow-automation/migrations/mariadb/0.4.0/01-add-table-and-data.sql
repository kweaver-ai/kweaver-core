USE adp;


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
