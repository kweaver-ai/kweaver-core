-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：在 kweaver schema 下创建 mdl-data-model 相关表，并从 adp schema 复制数据
-- ==========================================

SET SCHEMA kweaver;

-- ==========================================
-- 1. t_metric_model 指标模型
-- ==========================================
CREATE TABLE IF NOT EXISTS t_metric_model (
  f_model_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_model_name VARCHAR(40 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(255 CHAR) DEFAULT NULL,
  f_catalog_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_catalog_content TEXT DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_measure_name VARCHAR(50 CHAR) NOT NULL DEFAULT '',
  f_metric_type VARCHAR(20 CHAR) NOT NULL,
  f_data_source VARCHAR(255 CHAR) NOT NULL,
  f_query_type VARCHAR(20 CHAR) NOT NULL,
  f_formula TEXT NOT NULL,
  f_formula_config TEXT DEFAULT NULL,
  f_analysis_dimessions VARCHAR(8192 CHAR) DEFAULT NULL,
  f_order_by_fields VARCHAR(4096 CHAR) DEFAULT NULL,
  f_having_condition VARCHAR(2048 CHAR) DEFAULT NULL,
  f_date_field VARCHAR(255 CHAR) DEFAULT NULL,
  f_measure_field VARCHAR(255 CHAR) NOT NULL,
  f_unit_type VARCHAR(40 CHAR) NOT NULL,
  f_unit VARCHAR(20 CHAR) NOT NULL,
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_builtin TINYINT DEFAULT 0,
  f_calendar_interval TINYINT DEFAULT 0,
  CLUSTER PRIMARY KEY (f_model_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_metric_model_uk_model_name ON t_metric_model(f_group_id, f_model_name);

INSERT INTO kweaver."t_metric_model" (
    "f_model_id", "f_model_name", "f_tags", "f_comment",
    "f_catalog_id", "f_catalog_content",
    "f_creator", "f_creator_type", "f_create_time", "f_update_time",
    "f_measure_name", "f_metric_type", "f_data_source", "f_query_type",
    "f_formula", "f_formula_config", "f_analysis_dimessions", "f_order_by_fields",
    "f_having_condition", "f_date_field", "f_measure_field",
    "f_unit_type", "f_unit", "f_group_id", "f_builtin", "f_calendar_interval"
)
SELECT
    s."f_model_id", s."f_model_name", s."f_tags", s."f_comment",
    s."f_catalog_id", s."f_catalog_content",
    s."f_creator", s."f_creator_type", s."f_create_time", s."f_update_time",
    s."f_measure_name", s."f_metric_type", s."f_data_source", s."f_query_type",
    s."f_formula", s."f_formula_config", s."f_analysis_dimessions", s."f_order_by_fields",
    s."f_having_condition", s."f_date_field", s."f_measure_field",
    s."f_unit_type", s."f_unit", s."f_group_id", s."f_builtin", s."f_calendar_interval"
FROM adp."t_metric_model" s;

-- ==========================================
-- 2. t_metric_model_group 指标模型分组
-- ==========================================
CREATE TABLE IF NOT EXISTS t_metric_model_group (
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_group_name VARCHAR(40 CHAR) NOT NULL,
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_builtin TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_group_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_metric_model_group_uk_f_group_name ON t_metric_model_group(f_group_name);

INSERT INTO kweaver."t_metric_model_group" (
    "f_group_id", "f_group_name", "f_comment",
    "f_create_time", "f_update_time", "f_builtin"
)
SELECT
    s."f_group_id", s."f_group_name", s."f_comment",
    s."f_create_time", s."f_update_time", s."f_builtin"
FROM adp."t_metric_model_group" s;

-- ==========================================
-- 3. t_metric_model_task 指标模型持久化任务
-- ==========================================
CREATE TABLE IF NOT EXISTS t_metric_model_task (
  f_task_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_task_name VARCHAR(40 CHAR) NOT NULL,
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_module_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_model_id VARCHAR(40 CHAR) NOT NULL,
  f_schedule VARCHAR(255 CHAR) NOT NULL,
  f_variables TEXT DEFAULT NULL,
  f_time_windows VARCHAR(1024 CHAR) DEFAULT NULL,
  f_steps VARCHAR(255 CHAR) NOT NULL DEFAULT '[]',
  f_plan_time BIGINT NOT NULL DEFAULT 0,
  f_index_base VARCHAR(40 CHAR) NOT NULL,
  f_retrace_duration VARCHAR(20 CHAR) DEFAULT NULL,
  f_schedule_sync_status TINYINT NOT NULL,
  f_execute_status TINYINT DEFAULT 0,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY (f_task_id)
);

INSERT INTO kweaver."t_metric_model_task" (
    "f_task_id", "f_task_name", "f_comment",
    "f_create_time", "f_update_time", "f_module_type",
    "f_model_id", "f_schedule", "f_variables", "f_time_windows",
    "f_steps", "f_plan_time", "f_index_base", "f_retrace_duration",
    "f_schedule_sync_status", "f_execute_status",
    "f_creator", "f_creator_type"
)
SELECT
    s."f_task_id", s."f_task_name", s."f_comment",
    s."f_create_time", s."f_update_time", s."f_module_type",
    s."f_model_id", s."f_schedule", s."f_variables", s."f_time_windows",
    s."f_steps", s."f_plan_time", s."f_index_base", s."f_retrace_duration",
    s."f_schedule_sync_status", s."f_execute_status",
    s."f_creator", s."f_creator_type"
FROM adp."t_metric_model_task" s;

-- ==========================================
-- 4. t_static_metric_index 指标索引静态表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_static_metric_index (
  f_id INT IDENTITY(1, 1),
  f_base_type VARCHAR(40 CHAR) NOT NULL,
  f_split_time datetime(0) DEFAULT current_timestamp(),
  CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_static_metric_index_uk_f_index_base_type ON t_static_metric_index(f_base_type);

SET IDENTITY_INSERT kweaver.t_static_metric_index ON;

INSERT INTO kweaver."t_static_metric_index" (
    "f_id", "f_base_type", "f_split_time"
)
SELECT
    s."f_id", s."f_base_type", s."f_split_time"
FROM adp."t_static_metric_index" s;

SET IDENTITY_INSERT kweaver.t_static_metric_index OFF;

-- ==========================================
-- 5. t_event_model_aggregate_rules 事件模型聚合规则
-- ==========================================
CREATE TABLE IF NOT EXISTS t_event_model_aggregate_rules (
  f_aggregate_rule_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_aggregate_rule_type VARCHAR(40 CHAR) NOT NULL,
  f_aggregate_algo VARCHAR(900 CHAR) NOT NULL,
  f_rule_priority INT NOT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_group_fields VARCHAR(255 CHAR) DEFAULT '[]',
  f_aggregate_analysis_algo VARCHAR(1024 CHAR) DEFAULT '{}',
  CLUSTER PRIMARY KEY (f_aggregate_rule_id)
);

INSERT INTO kweaver."t_event_model_aggregate_rules" (
    "f_aggregate_rule_id", "f_aggregate_rule_type", "f_aggregate_algo",
    "f_rule_priority", "f_create_time", "f_update_time",
    "f_group_fields", "f_aggregate_analysis_algo"
)
SELECT
    s."f_aggregate_rule_id", s."f_aggregate_rule_type", s."f_aggregate_algo",
    s."f_rule_priority", s."f_create_time", s."f_update_time",
    s."f_group_fields", s."f_aggregate_analysis_algo"
FROM adp."t_event_model_aggregate_rules" s;

-- ==========================================
-- 6. t_event_models 事件模型
-- ==========================================
CREATE TABLE IF NOT EXISTS t_event_models (
  f_event_model_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_event_model_name VARCHAR(255 CHAR) NOT NULL,
  f_event_model_group_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_event_model_type VARCHAR(40 CHAR) NOT NULL,
  f_event_model_tags VARCHAR(255 CHAR) NOT NULL,
  f_event_model_comment VARCHAR(255 CHAR) DEFAULT NULL,
  f_data_source_type VARCHAR(40 CHAR) NOT NULL,
  f_data_source VARCHAR(900 CHAR) DEFAULT NULL,
  f_detect_rule_id VARCHAR(40 CHAR) NOT NULL,
  f_aggregate_rule_id VARCHAR(40 CHAR) NOT NULL,
  f_default_time_window VARCHAR(40 CHAR) NOT NULL,
  f_is_active TINYINT DEFAULT 0,
  f_enable_subscribe TINYINT DEFAULT 0,
  f_status TINYINT DEFAULT 0,
  f_downstream_dependent_model VARCHAR(1024 CHAR) DEFAULT '',
  f_is_custom TINYINT NOT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_event_model_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_event_models_uk_f_model_name ON t_event_models(f_event_model_name);

INSERT INTO kweaver."t_event_models" (
    "f_event_model_id", "f_event_model_name", "f_event_model_group_name",
    "f_event_model_type", "f_event_model_tags", "f_event_model_comment",
    "f_data_source_type", "f_data_source",
    "f_detect_rule_id", "f_aggregate_rule_id", "f_default_time_window",
    "f_is_active", "f_enable_subscribe", "f_status",
    "f_downstream_dependent_model", "f_is_custom",
    "f_creator", "f_creator_type", "f_create_time", "f_update_time"
)
SELECT
    s."f_event_model_id", s."f_event_model_name", s."f_event_model_group_name",
    s."f_event_model_type", s."f_event_model_tags", s."f_event_model_comment",
    s."f_data_source_type", s."f_data_source",
    s."f_detect_rule_id", s."f_aggregate_rule_id", s."f_default_time_window",
    s."f_is_active", s."f_enable_subscribe", s."f_status",
    s."f_downstream_dependent_model", s."f_is_custom",
    s."f_creator", s."f_creator_type", s."f_create_time", s."f_update_time"
FROM adp."t_event_models" s;

-- ==========================================
-- 7. t_event_model_detect_rules 事件模型检测规则
-- ==========================================
CREATE TABLE IF NOT EXISTS t_event_model_detect_rules (
  f_detect_rule_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_detect_rule_type VARCHAR(40 CHAR) NOT NULL,
  f_formula VARCHAR(2014 CHAR) DEFAULT NULL,
  f_detect_algo VARCHAR(40 CHAR) DEFAULT NULL,
  f_detect_analysis_algo VARCHAR(1024 CHAR) DEFAULT '{}',
  f_rule_priority INT NOT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_detect_rule_id)
);

INSERT INTO kweaver."t_event_model_detect_rules" (
    "f_detect_rule_id", "f_detect_rule_type", "f_formula",
    "f_detect_algo", "f_detect_analysis_algo", "f_rule_priority",
    "f_create_time", "f_update_time"
)
SELECT
    s."f_detect_rule_id", s."f_detect_rule_type", s."f_formula",
    s."f_detect_algo", s."f_detect_analysis_algo", s."f_rule_priority",
    s."f_create_time", s."f_update_time"
FROM adp."t_event_model_detect_rules" s;

-- ==========================================
-- 8. t_event_model_task 事件模型持久化任务
-- ==========================================
CREATE TABLE IF NOT EXISTS t_event_model_task (
  f_task_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_model_id VARCHAR(40 CHAR) NOT NULL,
  f_storage_config VARCHAR(255 CHAR) NOT NULL,
  f_schedule VARCHAR(255 CHAR) NOT NULL,
  f_dispatch_config VARCHAR(255 CHAR) NOT NULL,
  f_execute_parameter VARCHAR(255 CHAR) NOT NULL,
  f_task_status TINYINT NOT NULL,
  f_error_details VARCHAR(2048 CHAR) NOT NULL,
  f_status_update_time BIGINT NOT NULL DEFAULT 0,
  f_schedule_sync_status TINYINT NOT NULL,
  f_downstream_dependent_task VARCHAR(1024 CHAR) DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_task_id)
);

INSERT INTO kweaver."t_event_model_task" (
    "f_task_id", "f_model_id",
    "f_storage_config", "f_schedule", "f_dispatch_config", "f_execute_parameter",
    "f_task_status", "f_error_details", "f_status_update_time",
    "f_schedule_sync_status", "f_downstream_dependent_task",
    "f_creator", "f_creator_type", "f_create_time", "f_update_time"
)
SELECT
    s."f_task_id", s."f_model_id",
    s."f_storage_config", s."f_schedule", s."f_dispatch_config", s."f_execute_parameter",
    s."f_task_status", s."f_error_details", s."f_status_update_time",
    s."f_schedule_sync_status", s."f_downstream_dependent_task",
    s."f_creator", s."f_creator_type", s."f_create_time", s."f_update_time"
FROM adp."t_event_model_task" s;

-- ==========================================
-- 9. t_event_model_task_execution_records 事件模型异步任务记录
-- ==========================================
CREATE TABLE IF NOT EXISTS t_event_model_task_execution_records (
  f_run_id BIGINT NOT NULL,
  f_run_type VARCHAR(40 CHAR) NOT NULL,
  f_execute_parameter VARCHAR(2048 CHAR) NOT NULL,
  f_status VARCHAR(40 CHAR) DEFAULT '0',
  f_error_details VARCHAR(1024 CHAR) NOT NULL,
  f_update_time datetime(0) NOT NULL,
  f_create_time datetime(0) NOT NULL,
  CLUSTER PRIMARY KEY (f_run_id)
);

INSERT INTO kweaver."t_event_model_task_execution_records" (
    "f_run_id", "f_run_type", "f_execute_parameter",
    "f_status", "f_error_details",
    "f_update_time", "f_create_time"
)
SELECT
    s."f_run_id", s."f_run_type", s."f_execute_parameter",
    s."f_status", s."f_error_details",
    s."f_update_time", s."f_create_time"
FROM adp."t_event_model_task_execution_records" s;

-- ==========================================
-- 10. t_data_view 数据视图
-- ==========================================
CREATE TABLE IF NOT EXISTS t_data_view (
  f_view_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_view_name VARCHAR(255 CHAR) NOT NULL,
  f_technical_name VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_type VARCHAR(10 CHAR) NOT NULL DEFAULT '',
  f_query_type VARCHAR(10 CHAR) NOT NULL DEFAULT '',
  f_builtin TINYINT DEFAULT 0,
  f_tags VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_data_source_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_data_source_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_file_name VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  f_excel_config TEXT DEFAULT NULL,
  f_data_scope TEXT DEFAULT NULL,
  f_fields TEXT DEFAULT NULL,
  f_status VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_metadata_form_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_primary_keys VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_sql TEXT DEFAULT NULL,
  f_meta_table_name VARCHAR(1024 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_delete_time BIGINT NOT NULL DEFAULT 0,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_data_source TEXT DEFAULT NULL,
  f_field_scope TINYINT NOT NULL DEFAULT '0',
  f_filters TEXT DEFAULT NULL,
  f_open_streaming TINYINT NOT NULL DEFAULT 0,
  f_job_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_loggroup_filters TEXT DEFAULT NULL,
  CLUSTER PRIMARY KEY (f_view_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_view_uk_f_view_name ON t_data_view(f_group_id, f_view_name, f_delete_time);

INSERT INTO kweaver."t_data_view" (
    "f_view_id", "f_view_name", "f_technical_name", "f_group_id",
    "f_type", "f_query_type", "f_builtin", "f_tags", "f_comment",
    "f_data_source_type", "f_data_source_id", "f_file_name",
    "f_excel_config", "f_data_scope", "f_fields",
    "f_status", "f_metadata_form_id", "f_primary_keys", "f_sql", "f_meta_table_name",
    "f_create_time", "f_update_time", "f_delete_time",
    "f_creator", "f_creator_type", "f_updater", "f_updater_type",
    "f_data_source", "f_field_scope", "f_filters",
    "f_open_streaming", "f_job_id", "f_loggroup_filters"
)
SELECT
    s."f_view_id", s."f_view_name", s."f_technical_name", s."f_group_id",
    s."f_type", s."f_query_type", s."f_builtin", s."f_tags", s."f_comment",
    s."f_data_source_type", s."f_data_source_id", s."f_file_name",
    s."f_excel_config", s."f_data_scope", s."f_fields",
    s."f_status", s."f_metadata_form_id", s."f_primary_keys", s."f_sql", s."f_meta_table_name",
    s."f_create_time", s."f_update_time", s."f_delete_time",
    s."f_creator", s."f_creator_type", s."f_updater", s."f_updater_type",
    s."f_data_source", s."f_field_scope", s."f_filters",
    s."f_open_streaming", s."f_job_id", s."f_loggroup_filters"
FROM adp."t_data_view" s;

-- ==========================================
-- 11. t_data_view_group 数据视图分组
-- ==========================================
CREATE TABLE IF NOT EXISTS t_data_view_group (
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_group_name VARCHAR(40 CHAR) NOT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_delete_time BIGINT NOT NULL DEFAULT 0,
  f_builtin TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_group_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS "t_data_view_group_uk_f_group_name" ON "t_data_view_group"(f_builtin, f_group_name, f_delete_time);

INSERT INTO kweaver."t_data_view_group" (
    "f_group_id", "f_group_name",
    "f_create_time", "f_update_time", "f_delete_time", "f_builtin"
)
SELECT
    s."f_group_id", s."f_group_name",
    s."f_create_time", s."f_update_time", s."f_delete_time", s."f_builtin"
FROM adp."t_data_view_group" s;

-- ==========================================
-- 12. t_data_view_row_column_rule 视图行列规则
-- ==========================================
CREATE TABLE IF NOT EXISTS t_data_view_row_column_rule (
  f_rule_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_rule_name VARCHAR(255 CHAR) NOT NULL,
  f_view_id VARCHAR(40 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_fields TEXT NOT NULL,
  f_row_filters TEXT NOT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY (f_rule_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS "t_data_view_row_column_rule_uk_f_rule_name" ON "t_data_view_row_column_rule" (f_rule_name, f_view_id);

INSERT INTO kweaver."t_data_view_row_column_rule" (
    "f_rule_id", "f_rule_name", "f_view_id",
    "f_tags", "f_comment", "f_fields", "f_row_filters",
    "f_create_time", "f_update_time",
    "f_creator", "f_creator_type", "f_updater", "f_updater_type"
)
SELECT
    s."f_rule_id", s."f_rule_name", s."f_view_id",
    s."f_tags", s."f_comment", s."f_fields", s."f_row_filters",
    s."f_create_time", s."f_update_time",
    s."f_creator", s."f_creator_type", s."f_updater", s."f_updater_type"
FROM adp."t_data_view_row_column_rule" s;

-- ==========================================
-- 13. t_data_dict 数据字典
-- ==========================================
CREATE TABLE IF NOT EXISTS t_data_dict (
  f_dict_id VARCHAR(40 CHAR) NOT NULL,
  f_dict_name VARCHAR(255 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) NOT NULL,
  f_comment VARCHAR(255 CHAR) DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_dict_type VARCHAR(20 CHAR) NOT NULL DEFAULT 'kv_dict',
  f_dict_store VARCHAR(255 CHAR) NOT NULL,
  f_dimension VARCHAR(1500 CHAR) NOT NULL,
  f_unique_key TINYINT NOT NULL DEFAULT 1,
  CLUSTER PRIMARY KEY (f_dict_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_dict_uk_dict_name ON t_data_dict(f_dict_name);

INSERT INTO kweaver."t_data_dict" (
    "f_dict_id", "f_dict_name", "f_tags", "f_comment",
    "f_creator", "f_creator_type", "f_create_time", "f_update_time",
    "f_dict_type", "f_dict_store", "f_dimension", "f_unique_key"
)
SELECT
    s."f_dict_id", s."f_dict_name", s."f_tags", s."f_comment",
    s."f_creator", s."f_creator_type", s."f_create_time", s."f_update_time",
    s."f_dict_type", s."f_dict_store", s."f_dimension", s."f_unique_key"
FROM adp."t_data_dict" s;

-- ==========================================
-- 14. t_data_dict_item 数据字典项
-- ==========================================
CREATE TABLE IF NOT EXISTS t_data_dict_item (
  f_item_id VARCHAR(40 CHAR) NOT NULL,
  f_dict_id VARCHAR(40 CHAR) NOT NULL,
  f_item_key VARCHAR(3000 CHAR) NOT NULL,
  f_item_value VARCHAR(3000 CHAR) NOT NULL,
  f_comment VARCHAR(255 CHAR),
  CLUSTER PRIMARY KEY (f_item_id)
);

CREATE INDEX IF NOT EXISTS t_data_dict_item_idx_dict_id ON t_data_dict_item(f_dict_id);

INSERT INTO kweaver."t_data_dict_item" (
    "f_item_id", "f_dict_id", "f_item_key", "f_item_value", "f_comment"
)
SELECT
    s."f_item_id", s."f_dict_id", s."f_item_key", s."f_item_value", s."f_comment"
FROM adp."t_data_dict_item" s;

-- ==========================================
-- 15. t_data_connection 数据连接
-- ==========================================
CREATE TABLE IF NOT EXISTS t_data_connection (
  f_connection_id VARCHAR(40 CHAR) NOT NULL,
  f_connection_name VARCHAR(40 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) DEFAULT '',
  f_comment VARCHAR(255 CHAR) DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_data_source_type VARCHAR(40 CHAR) NOT NULL,
  f_config TEXT NOT NULL,
  f_config_md5 VARCHAR(32 CHAR) DEFAULT '',
  CLUSTER PRIMARY KEY (f_connection_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_connection_uk_f_connection_name ON t_data_connection(f_connection_name);
CREATE INDEX IF NOT EXISTS t_data_connection_idx_f_data_source_type ON t_data_connection(f_data_source_type);
CREATE INDEX IF NOT EXISTS t_data_connection_idx_f_config_md5 ON t_data_connection(f_config_md5);

INSERT INTO kweaver."t_data_connection" (
    "f_connection_id", "f_connection_name", "f_tags", "f_comment",
    "f_create_time", "f_update_time",
    "f_data_source_type", "f_config", "f_config_md5"
)
SELECT
    s."f_connection_id", s."f_connection_name", s."f_tags", s."f_comment",
    s."f_create_time", s."f_update_time",
    s."f_data_source_type", s."f_config", s."f_config_md5"
FROM adp."t_data_connection" s;

-- ==========================================
-- 16. t_data_connection_status 数据连接状态
-- ==========================================
CREATE TABLE IF NOT EXISTS t_data_connection_status (
  f_connection_id VARCHAR(40 CHAR) NOT NULL,
  f_status VARCHAR(5 CHAR) NOT NULL,
  f_detection_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_connection_id)
);

INSERT INTO kweaver."t_data_connection_status" (
    "f_connection_id", "f_status", "f_detection_time"
)
SELECT
    s."f_connection_id", s."f_status", s."f_detection_time"
FROM adp."t_data_connection_status" s;

-- ==========================================
-- 17. t_trace_model 链路模型
-- ==========================================
CREATE TABLE IF NOT EXISTS t_trace_model (
  f_model_id VARCHAR(40 CHAR) NOT NULL,
  f_model_name VARCHAR(40 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) DEFAULT '',
  f_comment VARCHAR(255 CHAR) DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_span_source_type VARCHAR(40 CHAR) NOT NULL,
  f_span_config TEXT NOT NULL,
  f_enabled_related_log TINYINT NOT NULL,
  f_related_log_source_type VARCHAR(40 CHAR) NOT NULL,
  f_related_log_config TEXT NOT NULL,
  CLUSTER PRIMARY KEY (f_model_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_trace_model_uk_f_model_name ON t_trace_model(f_model_name);
CREATE INDEX IF NOT EXISTS t_trace_model_idx_f_span_source_type ON t_trace_model(f_span_source_type);

INSERT INTO kweaver."t_trace_model" (
    "f_model_id", "f_model_name", "f_tags", "f_comment",
    "f_creator", "f_creator_type", "f_create_time", "f_update_time",
    "f_span_source_type", "f_span_config",
    "f_enabled_related_log", "f_related_log_source_type", "f_related_log_config"
)
SELECT
    s."f_model_id", s."f_model_name", s."f_tags", s."f_comment",
    s."f_creator", s."f_creator_type", s."f_create_time", s."f_update_time",
    s."f_span_source_type", s."f_span_config",
    s."f_enabled_related_log", s."f_related_log_source_type", s."f_related_log_config"
FROM adp."t_trace_model" s;

-- ==========================================
-- 18. t_data_model_job 全局任务表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_data_model_job (
  f_job_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_job_type VARCHAR(40 CHAR) NOT NULL,
  f_job_config TEXT,
  f_job_status VARCHAR(20 CHAR) NOT NULL,
  f_job_status_details TEXT NOT NULL,
  CLUSTER PRIMARY KEY (f_job_id)
);

INSERT INTO kweaver."t_data_model_job" (
    "f_job_id", "f_creator", "f_creator_type",
    "f_create_time", "f_update_time",
    "f_job_type", "f_job_config", "f_job_status", "f_job_status_details"
)
SELECT
    s."f_job_id", s."f_creator", s."f_creator_type",
    s."f_create_time", s."f_update_time",
    s."f_job_type", s."f_job_config", s."f_job_status", s."f_job_status_details"
FROM adp."t_data_model_job" s;

-- ==========================================
-- 19. t_objective_model 目标模型
-- ==========================================
CREATE TABLE IF NOT EXISTS t_objective_model (
  f_model_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_model_name VARCHAR(40 CHAR) NOT NULL,
  f_tags VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  f_objective_type VARCHAR(20 CHAR) NOT NULL,
  f_objective_config TEXT NOT NULL,
  CLUSTER PRIMARY KEY (f_model_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_objective_model_uk_t_objective_model ON t_objective_model(f_model_name);

INSERT INTO kweaver."t_objective_model" (
    "f_model_id", "f_model_name", "f_tags", "f_comment",
    "f_creator", "f_creator_type", "f_create_time", "f_update_time",
    "f_objective_type", "f_objective_config"
)
SELECT
    s."f_model_id", s."f_model_name", s."f_tags", s."f_comment",
    s."f_creator", s."f_creator_type", s."f_create_time", s."f_update_time",
    s."f_objective_type", s."f_objective_config"
FROM adp."t_objective_model" s;

-- ==========================================
-- 20. t_scan_record 数据源扫描记录
-- ==========================================
CREATE TABLE IF NOT EXISTS t_scan_record (
  f_record_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_data_source_id VARCHAR(40 CHAR) NOT NULL,
  f_scanner VARCHAR(40 CHAR) NOT NULL,
  f_scan_time BIGINT NOT NULL DEFAULT 0,
  f_data_source_status VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_metadata_task_id VARCHAR(128 CHAR) DEFAULT NULL,
  CLUSTER PRIMARY KEY (f_record_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_scan_record_uk_scan_record ON t_scan_record(f_data_source_id, f_scanner);

INSERT INTO kweaver."t_scan_record" (
    "f_record_id", "f_data_source_id", "f_scanner",
    "f_scan_time", "f_data_source_status", "f_metadata_task_id"
)
SELECT
    s."f_record_id", s."f_data_source_id", s."f_scanner",
    s."f_scan_time", s."f_data_source_status", s."f_metadata_task_id"
FROM adp."t_scan_record" s;
