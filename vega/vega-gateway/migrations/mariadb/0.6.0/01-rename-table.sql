-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：将 vega-gateway 相关表从 adp 库迁移至 kweaver 库
-- ==========================================
USE kweaver;

RENAME TABLE adp.cache_table TO kweaver.cache_table;
RENAME TABLE adp.client_id TO kweaver.client_id;
RENAME TABLE adp.excel_column_type TO kweaver.excel_column_type;
RENAME TABLE adp.excel_table_config TO kweaver.excel_table_config;
RENAME TABLE adp.query_info TO kweaver.query_info;
RENAME TABLE adp.task_info TO kweaver.task_info;
