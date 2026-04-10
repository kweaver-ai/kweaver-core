-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：修改 t_resource 表的索引结构
-- ==========================================
-- 说明：
-- 1. 删除原有的 idx_t_resource_catalog_name 唯一索引
-- 2. 添加新的 idx_t_resource_catalog_source_identifier 唯一索引

SET SCHEMA adp;

-- 删除原有的 idx_t_resource_catalog_name 唯一索引
DROP INDEX IF EXISTS t_resource.idx_t_resource_catalog_name;

-- 添加新的 idx_t_resource_catalog_source_identifier 唯一索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_t_resource_catalog_source_identifier ON t_resource(f_catalog_id, f_source_identifier);
