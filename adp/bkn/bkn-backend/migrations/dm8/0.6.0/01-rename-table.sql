-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：在 kweaver schema 下创建 bkn 相关表，并从 adp schema 复制数据
-- ==========================================

SET SCHEMA kweaver;

-- ==========================================
-- 1. t_knowledge_network 业务知识网络
-- ==========================================
CREATE TABLE IF NOT EXISTS t_knowledge_network (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
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

INSERT INTO kweaver."t_knowledge_network" (
    "f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
    "f_branch", "f_business_domain",
    "f_creator", "f_creator_type", "f_create_time",
    "f_updater", "f_updater_type", "f_update_time"
)
SELECT
    s."f_id", s."f_name", s."f_tags", s."f_comment", s."f_icon", s."f_color", s."f_bkn_raw_content",
    s."f_branch", s."f_business_domain",
    s."f_creator", s."f_creator_type", s."f_create_time",
    s."f_updater", s."f_updater_type", s."f_update_time"
FROM adp."t_knowledge_network" s;

-- ==========================================
-- 2. t_object_type 对象类
-- ==========================================
CREATE TABLE IF NOT EXISTS t_object_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
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

INSERT INTO kweaver."t_object_type" (
    "f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
    "f_kn_id", "f_branch",
    "f_data_source", "f_data_properties", "f_logic_properties", "f_primary_keys",
    "f_display_key", "f_incremental_key",
    "f_creator", "f_creator_type", "f_create_time",
    "f_updater", "f_updater_type", "f_update_time"
)
SELECT
    s."f_id", s."f_name", s."f_tags", s."f_comment", s."f_icon", s."f_color", s."f_bkn_raw_content",
    s."f_kn_id", s."f_branch",
    s."f_data_source", s."f_data_properties", s."f_logic_properties", s."f_primary_keys",
    s."f_display_key", s."f_incremental_key",
    s."f_creator", s."f_creator_type", s."f_create_time",
    s."f_updater", s."f_updater_type", s."f_update_time"
FROM adp."t_object_type" s;

-- ==========================================
-- 3. t_object_type_status 对象类状态
-- ==========================================
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

INSERT INTO kweaver."t_object_type_status" (
    "f_id", "f_kn_id", "f_branch",
    "f_incremental_key", "f_incremental_value",
    "f_index", "f_index_available", "f_doc_count", "f_storage_size", "f_update_time"
)
SELECT
    s."f_id", s."f_kn_id", s."f_branch",
    s."f_incremental_key", s."f_incremental_value",
    s."f_index", s."f_index_available", s."f_doc_count", s."f_storage_size", s."f_update_time"
FROM adp."t_object_type_status" s;

-- ==========================================
-- 4. t_relation_type 关系类
-- ==========================================
CREATE TABLE IF NOT EXISTS t_relation_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
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

INSERT INTO kweaver."t_relation_type" (
    "f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
    "f_kn_id", "f_branch",
    "f_source_object_type_id", "f_target_object_type_id", "f_type", "f_mapping_rules",
    "f_creator", "f_creator_type", "f_create_time",
    "f_updater", "f_updater_type", "f_update_time"
)
SELECT
    s."f_id", s."f_name", s."f_tags", s."f_comment", s."f_icon", s."f_color", s."f_bkn_raw_content",
    s."f_kn_id", s."f_branch",
    s."f_source_object_type_id", s."f_target_object_type_id", s."f_type", s."f_mapping_rules",
    s."f_creator", s."f_creator_type", s."f_create_time",
    s."f_updater", s."f_updater_type", s."f_update_time"
FROM adp."t_relation_type" s;

-- ==========================================
-- 5. t_action_type 行动类
-- ==========================================
CREATE TABLE IF NOT EXISTS t_action_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
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

INSERT INTO kweaver."t_action_type" (
    "f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
    "f_kn_id", "f_branch",
    "f_action_type", "f_object_type_id", "f_condition", "f_affect",
    "f_action_source", "f_parameters", "f_schedule",
    "f_creator", "f_creator_type", "f_create_time",
    "f_updater", "f_updater_type", "f_update_time"
)
SELECT
    s."f_id", s."f_name", s."f_tags", s."f_comment", s."f_icon", s."f_color", s."f_bkn_raw_content",
    s."f_kn_id", s."f_branch",
    s."f_action_type", s."f_object_type_id", s."f_condition", s."f_affect",
    s."f_action_source", s."f_parameters", s."f_schedule",
    s."f_creator", s."f_creator_type", s."f_create_time",
    s."f_updater", s."f_updater_type", s."f_update_time"
FROM adp."t_action_type" s;

-- ==========================================
-- 6. t_kn_job 任务管理
-- ==========================================
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

INSERT INTO kweaver."t_kn_job" (
    "f_id", "f_name", "f_kn_id", "f_branch",
    "f_job_type", "f_job_concept_config", "f_state", "f_state_detail",
    "f_creator", "f_creator_type", "f_create_time", "f_finish_time", "f_time_cost"
)
SELECT
    s."f_id", s."f_name", s."f_kn_id", s."f_branch",
    s."f_job_type", s."f_job_concept_config", s."f_state", s."f_state_detail",
    s."f_creator", s."f_creator_type", s."f_create_time", s."f_finish_time", s."f_time_cost"
FROM adp."t_kn_job" s;

-- ==========================================
-- 7. t_kn_task 子任务管理
-- ==========================================
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

INSERT INTO kweaver."t_kn_task" (
    "f_id", "f_name", "f_job_id",
    "f_concept_type", "f_concept_id",
    "f_index", "f_doc_count",
    "f_state", "f_state_detail",
    "f_start_time", "f_finish_time", "f_time_cost"
)
SELECT
    s."f_id", s."f_name", s."f_job_id",
    s."f_concept_type", s."f_concept_id",
    s."f_index", s."f_doc_count",
    s."f_state", s."f_state_detail",
    s."f_start_time", s."f_finish_time", s."f_time_cost"
FROM adp."t_kn_task" s;

-- ==========================================
-- 8. t_concept_group 概念分组
-- ==========================================
CREATE TABLE IF NOT EXISTS t_concept_group (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_bkn_raw_content TEXT DEFAULT NULL,
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

INSERT INTO kweaver."t_concept_group" (
    "f_id", "f_name", "f_tags", "f_comment", "f_icon", "f_color", "f_bkn_raw_content",
    "f_kn_id", "f_branch",
    "f_creator", "f_creator_type", "f_create_time",
    "f_updater", "f_updater_type", "f_update_time"
)
SELECT
    s."f_id", s."f_name", s."f_tags", s."f_comment", s."f_icon", s."f_color", s."f_bkn_raw_content",
    s."f_kn_id", s."f_branch",
    s."f_creator", s."f_creator_type", s."f_create_time",
    s."f_updater", s."f_updater_type", s."f_update_time"
FROM adp."t_concept_group" s;

-- ==========================================
-- 9. t_concept_group_relation 分组与概念对应表
-- ==========================================
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

INSERT INTO kweaver."t_concept_group_relation" (
    "f_id", "f_kn_id", "f_branch",
    "f_group_id", "f_concept_type", "f_concept_id", "f_create_time"
)
SELECT
    s."f_id", s."f_kn_id", s."f_branch",
    s."f_group_id", s."f_concept_type", s."f_concept_id", s."f_create_time"
FROM adp."t_concept_group_relation" s;

-- ==========================================
-- 10. t_action_schedule Action Schedule
-- ==========================================
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

INSERT INTO kweaver."t_action_schedule" (
    "f_id", "f_name", "f_kn_id", "f_branch",
    "f_action_type_id", "f_cron_expression",
    "f_instance_identities", "f_dynamic_params",
    "f_status", "f_last_run_time", "f_next_run_time",
    "f_lock_holder", "f_lock_time",
    "f_creator", "f_creator_type", "f_create_time",
    "f_updater", "f_updater_type", "f_update_time"
)
SELECT
    s."f_id", s."f_name", s."f_kn_id", s."f_branch",
    s."f_action_type_id", s."f_cron_expression",
    s."f_instance_identities", s."f_dynamic_params",
    s."f_status", s."f_last_run_time", s."f_next_run_time",
    s."f_lock_holder", s."f_lock_time",
    s."f_creator", s."f_creator_type", s."f_create_time",
    s."f_updater", s."f_updater_type", s."f_update_time"
FROM adp."t_action_schedule" s;

-- ==========================================
-- 11. t_risk_type 风险类
-- ==========================================
CREATE TABLE IF NOT EXISTS t_risk_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
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

INSERT INTO kweaver."t_risk_type" (
    "f_id", "f_name", "f_comment", "f_tags", "f_icon", "f_color",
    "f_kn_id", "f_branch",
    "f_creator", "f_creator_type", "f_create_time",
    "f_updater", "f_updater_type", "f_update_time"
)
SELECT
    s."f_id", s."f_name", s."f_comment", s."f_tags", s."f_icon", s."f_color",
    s."f_kn_id", s."f_branch",
    s."f_creator", s."f_creator_type", s."f_create_time",
    s."f_updater", s."f_updater_type", s."f_update_time"
FROM adp."t_risk_type" s;
