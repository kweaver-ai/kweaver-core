-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：修改 t_build_task 表的索引结构
-- ==========================================

USE kweaver;

ALTER TABLE t_build_task ADD COLUMN IF NOT EXISTS f_catalog_id VARCHAR(40) NOT NULL DEFAULT '';
DROP INDEX IF EXISTS idx_create_time on t_build_task;
ALTER TABLE t_build_task ADD INDEX IF NOT EXISTS idx_catalog_id (f_catalog_id);
UPDATE t_build_task bt JOIN t_resource r ON bt.f_resource_id = r.f_id SET bt.f_catalog_id = r.f_catalog_id;
