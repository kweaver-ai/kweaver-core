-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：将 data-connection 相关表从 adp 库迁移至 kweaver 库
-- ==========================================
USE kweaver;

RENAME TABLE adp.data_source TO kweaver.data_source;
RENAME TABLE adp.t_data_source_info TO kweaver.t_data_source_info;
RENAME TABLE adp.t_task_scan TO kweaver.t_task_scan;
RENAME TABLE adp.t_task_scan_table TO kweaver.t_task_scan_table;
RENAME TABLE adp.t_table_scan TO kweaver.t_table_scan;
RENAME TABLE adp.t_table_field_scan TO kweaver.t_table_field_scan;
RENAME TABLE adp.t_task_scan_schedule TO kweaver.t_task_scan_schedule;
RENAME TABLE adp.t_data_quality_model TO kweaver.t_data_quality_model;
RENAME TABLE adp.t_data_quality_rule TO kweaver.t_data_quality_rule;
RENAME TABLE adp.t_data_source TO kweaver.t_data_source;
RENAME TABLE adp.t_dict TO kweaver.t_dict;
RENAME TABLE adp.t_indicator TO kweaver.t_indicator;
RENAME TABLE adp.t_lineage_edge_column TO kweaver.t_lineage_edge_column;
RENAME TABLE adp.t_lineage_edge_column_table_relation TO kweaver.t_lineage_edge_column_table_relation;
RENAME TABLE adp.t_lineage_edge_table TO kweaver.t_lineage_edge_table;
RENAME TABLE adp.t_lineage_graph_info TO kweaver.t_lineage_graph_info;
RENAME TABLE adp.t_lineage_log TO kweaver.t_lineage_log;
RENAME TABLE adp.t_lineage_relation TO kweaver.t_lineage_relation;
RENAME TABLE adp.t_lineage_tag_column TO kweaver.t_lineage_tag_column;
RENAME TABLE adp.t_lineage_tag_table TO kweaver.t_lineage_tag_table;
RENAME TABLE adp.t_indicator2 TO kweaver.t_indicator2;
RENAME TABLE adp.t_lineage_tag_column2 TO kweaver.t_lineage_tag_column2;
RENAME TABLE adp.t_lineage_tag_indicator2 TO kweaver.t_lineage_tag_indicator2;
RENAME TABLE adp.t_lineage_tag_table2 TO kweaver.t_lineage_tag_table2;
RENAME TABLE adp.t_live_ddl TO kweaver.t_live_ddl;
RENAME TABLE adp.t_schema TO kweaver.t_schema;
RENAME TABLE adp.t_table TO kweaver.t_table;
RENAME TABLE adp.t_table_field TO kweaver.t_table_field;
RENAME TABLE adp.t_table_field_his TO kweaver.t_table_field_his;
RENAME TABLE adp.t_task TO kweaver.t_task;
RENAME TABLE adp.t_task_log TO kweaver.t_task_log;
