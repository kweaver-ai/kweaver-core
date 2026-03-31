-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

USE adp;

-- 业务知识网络
CREATE TABLE IF NOT EXISTS t_knowledge_network (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络id',
  f_name VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络名称',
  f_tags VARCHAR(255) DEFAULT NULL COMMENT '标签',
  f_comment VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '备注',
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
  f_comment VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '备注', 
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
  f_comment VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '备注', 
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
  f_comment VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '备注', 
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
  f_comment VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '备注',
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

-- Action Schedule Management
-- Supports cron-based scheduled action execution with distributed locking

CREATE TABLE IF NOT EXISTS t_action_schedule (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Schedule ID',
  f_name VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'Schedule name',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Knowledge network ID',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Branch',
  f_action_type_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Action type ID to execute',
  f_cron_expression VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'Standard 5-field cron expression (min hour dom mon dow)',
  f_unique_identities MEDIUMTEXT DEFAULT NULL COMMENT 'JSON array of target object unique identities',
  f_dynamic_params MEDIUMTEXT DEFAULT NULL COMMENT 'JSON object of dynamic parameters',
  f_status VARCHAR(20) NOT NULL DEFAULT 'inactive' COMMENT 'Schedule status: active or inactive',
  f_last_run_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'Last execution timestamp (ms)',
  f_next_run_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'Next scheduled run timestamp (ms)',
  f_lock_holder VARCHAR(64) DEFAULT NULL COMMENT 'Pod ID holding execution lock (NULL = unlocked)',
  f_lock_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'Lock acquisition timestamp (ms) for timeout detection',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Creator ID',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT 'Creator type',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'Create timestamp (ms)',
  f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT 'Updater ID',
  f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT 'Updater type',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'Update timestamp (ms)',
  PRIMARY KEY (f_id),
  KEY idx_kn_branch (f_kn_id, f_branch),
  KEY idx_status_next_run (f_status, f_next_run_time),
  KEY idx_action_type (f_action_type_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = 'Action schedule for cron-based execution';
