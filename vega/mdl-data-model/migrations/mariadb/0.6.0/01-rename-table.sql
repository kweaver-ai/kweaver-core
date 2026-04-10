-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：将 mdl-data-model 相关表从 adp 库迁移至 kweaver 库
-- ==========================================
USE kweaver;

RENAME TABLE adp.t_metric_model TO kweaver.t_metric_model;
RENAME TABLE adp.t_metric_model_group TO kweaver.t_metric_model_group;
RENAME TABLE adp.t_metric_model_task TO kweaver.t_metric_model_task;
RENAME TABLE adp.t_static_metric_index TO kweaver.t_static_metric_index;
RENAME TABLE adp.t_event_model_aggregate_rules TO kweaver.t_event_model_aggregate_rules;
RENAME TABLE adp.t_event_models TO kweaver.t_event_models;
RENAME TABLE adp.t_event_model_detect_rules TO kweaver.t_event_model_detect_rules;
RENAME TABLE adp.t_event_model_task TO kweaver.t_event_model_task;
RENAME TABLE adp.t_event_model_task_execution_records TO kweaver.t_event_model_task_execution_records;
RENAME TABLE adp.t_data_view TO kweaver.t_data_view;
RENAME TABLE adp.t_data_view_group TO kweaver.t_data_view_group;
RENAME TABLE adp.t_data_view_row_column_rule TO kweaver.t_data_view_row_column_rule;
RENAME TABLE adp.t_data_dict TO kweaver.t_data_dict;
RENAME TABLE adp.t_data_dict_item TO kweaver.t_data_dict_item;
RENAME TABLE adp.t_data_connection TO kweaver.t_data_connection;
RENAME TABLE adp.t_data_connection_status TO kweaver.t_data_connection_status;
RENAME TABLE adp.t_trace_model TO kweaver.t_trace_model;
RENAME TABLE adp.t_data_model_job TO kweaver.t_data_model_job;
RENAME TABLE adp.t_objective_model TO kweaver.t_objective_model;
RENAME TABLE adp.t_scan_record TO kweaver.t_scan_record;
