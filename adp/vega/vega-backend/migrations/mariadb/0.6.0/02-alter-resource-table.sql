-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：修改 t_resource 表的索引结构
-- ==========================================
-- 说明：
-- 1. 删除原有的 uk_catalog_name 唯一索引
-- 2. 添加新的 uk_catalog_source_identifier 唯一索引

USE adp;

-- 删除原有的 uk_catalog_name 唯一索引
ALTER TABLE t_resource DROP INDEX uk_catalog_name;

-- 添加新的 uk_catalog_source_identifier 唯一索引
ALTER TABLE t_resource ADD UNIQUE INDEX uk_catalog_source_identifier (f_catalog_id, f_source_identifier);
