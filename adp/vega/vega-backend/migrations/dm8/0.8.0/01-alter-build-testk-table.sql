-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：修改 t_build_task 表的索引结构
-- ==========================================

SET SCHEMA kweaver;

CREATE INDEX IF NOT EXISTS idx_t_build_task_catalog_id ON t_build_task(f_catalog_id);
UPDATE t_build_task bt JOIN t_resource r ON bt.f_resource_id = r.f_id SET bt.f_catalog_id = r.f_catalog_id;
