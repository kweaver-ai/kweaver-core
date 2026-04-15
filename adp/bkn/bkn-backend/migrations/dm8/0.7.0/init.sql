-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

SET SCHEMA kweaver;


-- 业务知识网络
CREATE TABLE IF NOT EXISTS t_knowledge_network (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
<<<<<<< HEAD
  f_comment TEXT NOT NULL,
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT NOT NULL,
  f_skill_content TEXT NOT NULL,
=======
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
>>>>>>> eb17aa73 ([feature] 198-issue 【BKN】bkn支持指标模型，以实现基于业务语义的指标定义-方案)
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_business_domain VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id,f_branch)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_knowledge_network_kn_name ON t_knowledge_network(f_name,f_branch);


-- 对象类
CREATE TABLE IF NOT EXISTS t_object_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
<<<<<<< HEAD
  f_comment TEXT NOT NULL,
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT NOT NULL,
=======
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
>>>>>>> eb17aa73 ([feature] 198-issue 【BKN】bkn支持指标模型，以实现基于业务语义的指标定义-方案)
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_data_source VARCHAR(255 CHAR) NOT NULL,
  f_data_properties TEXT DEFAULT NULL,
  f_logic_properties TEXT DEFAULT NULL,
  f_primary_keys VARCHAR(8192 CHAR) DEFAULT NULL,
  f_display_key VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_incremental_key VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_object_type_ot_name ON t_object_type(f_kn_id,f_branch,f_name);


-- 对象类状态
CREATE TABLE IF NOT EXISTS t_object_type_status (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_incremental_key VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_incremental_value VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_index VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_index_available BIT NOT NULL DEFAULT 0,
  f_doc_count BIGINT NOT NULL DEFAULT 0,
  f_storage_size BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);


-- 关系类
CREATE TABLE IF NOT EXISTS t_relation_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
<<<<<<< HEAD
  f_comment TEXT NOT NULL,
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT NOT NULL,
=======
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
>>>>>>> eb17aa73 ([feature] 198-issue 【BKN】bkn支持指标模型，以实现基于业务语义的指标定义-方案)
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_source_object_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_target_object_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_mapping_rules TEXT DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);


-- 行动类
CREATE TABLE IF NOT EXISTS t_action_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
<<<<<<< HEAD
  f_comment TEXT NOT NULL,
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT NOT NULL,
=======
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
>>>>>>> eb17aa73 ([feature] 198-issue 【BKN】bkn支持指标模型，以实现基于业务语义的指标定义-方案)
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_action_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_object_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_condition TEXT DEFAULT NULL,
  f_affect TEXT DEFAULT NULL,
  f_action_source VARCHAR(255 CHAR) NOT NULL,
  f_parameters TEXT DEFAULT NULL,
  f_schedule VARCHAR(255 CHAR) DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_action_type_at_name ON t_action_type(f_kn_id,f_branch,f_name);


-- 任务管理
CREATE TABLE IF NOT EXISTS t_kn_job (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_job_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_job_concept_config TEXT DEFAULT NULL,
  f_state VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_state_detail TEXT DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_finish_time BIGINT NOT NULL DEFAULT 0,
  f_time_cost BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
);


-- 子任务管理
CREATE TABLE IF NOT EXISTS t_kn_task (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_job_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_index VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_doc_count BIGINT NOT NULL DEFAULT 0,
  f_state VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_state_detail TEXT DEFAULT NULL,
  f_start_time BIGINT NOT NULL DEFAULT 0,
  f_finish_time BIGINT NOT NULL DEFAULT 0,
  f_time_cost BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
);


-- 概念分组
CREATE TABLE IF NOT EXISTS t_concept_group (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
<<<<<<< HEAD
  f_comment TEXT NOT NULL,
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT NOT NULL,
=======
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
>>>>>>> eb17aa73 ([feature] 198-issue 【BKN】bkn支持指标模型，以实现基于业务语义的指标定义-方案)
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_concept_group_name ON t_concept_group(f_kn_id,f_branch,f_name);


-- 分组与概念对应表
CREATE TABLE IF NOT EXISTS t_concept_group_relation (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_concept_group_relation ON t_concept_group_relation(f_kn_id,f_branch,f_group_id,f_concept_type,f_concept_id);


-- Action Schedule Management
-- Supports cron-based scheduled action execution with distributed locking
CREATE TABLE IF NOT EXISTS t_action_schedule (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(100 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_action_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_cron_expression VARCHAR(100 CHAR) NOT NULL DEFAULT '',
  f_instance_identities TEXT DEFAULT NULL,
  f_dynamic_params TEXT DEFAULT NULL,
  f_status VARCHAR(20 CHAR) NOT NULL DEFAULT 'inactive',
  f_last_run_time BIGINT NOT NULL DEFAULT 0,
  f_next_run_time BIGINT NOT NULL DEFAULT 0,
  f_lock_holder VARCHAR(64 CHAR) DEFAULT NULL,
  f_lock_time BIGINT NOT NULL DEFAULT 0,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS idx_action_schedule_kn_branch ON t_action_schedule(f_kn_id, f_branch);
CREATE INDEX IF NOT EXISTS idx_action_schedule_status_next_run ON t_action_schedule(f_status, f_next_run_time);
CREATE INDEX IF NOT EXISTS idx_action_schedule_action_type ON t_action_schedule(f_action_type_id);


-- Risk Type
CREATE TABLE IF NOT EXISTS t_risk_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
<<<<<<< HEAD
  f_comment TEXT NOT NULL,
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT NOT NULL,
=======
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
>>>>>>> eb17aa73 ([feature] 198-issue 【BKN】bkn支持指标模型，以实现基于业务语义的指标定义-方案)
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id, f_branch, f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_risk_type_name ON t_risk_type(f_kn_id, f_branch, f_name);
<<<<<<< HEAD
=======


-- BKN 指标定义（0.7.0 相对 0.6.0 新增；与 01-metric_definition.sql 一致）
CREATE TABLE IF NOT EXISTS t_metric_definition (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_unit_type VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  f_unit VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  f_metric_type VARCHAR(32 CHAR) NOT NULL DEFAULT 'atomic',
  f_scope_type VARCHAR(32 CHAR) NOT NULL DEFAULT 'object_type',
  f_scope_ref VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_time_dimension TEXT DEFAULT NULL,
  f_calculation_formula TEXT NOT NULL,
  f_analysis_dimensions TEXT DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id, f_branch, f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_metric_name ON t_metric_definition(f_kn_id, f_branch, f_name);
>>>>>>> eb17aa73 ([feature] 198-issue 【BKN】bkn支持指标模型，以实现基于业务语义的指标定义-方案)
