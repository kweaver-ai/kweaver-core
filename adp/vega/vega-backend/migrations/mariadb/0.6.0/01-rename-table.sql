-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：将 vega 相关表从 adp 库迁移至 kweaver 库
-- ==========================================
USE kweaver;

RENAME TABLE adp.t_catalog TO kweaver.t_catalog;
RENAME TABLE adp.t_catalog_discover_policy TO kweaver.t_catalog_discover_policy;
RENAME TABLE adp.t_resource TO kweaver.t_resource;
RENAME TABLE adp.t_resource_schema_history TO kweaver.t_resource_schema_history;
RENAME TABLE adp.t_connector_type TO kweaver.t_connector_type;
RENAME TABLE adp.t_discover_task TO kweaver.t_discover_task;
RENAME TABLE adp.t_build_task TO kweaver.t_build_task;
