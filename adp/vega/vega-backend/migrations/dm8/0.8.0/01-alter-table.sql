-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：修改 t_build_task 表的索引结构
-- ==========================================

SET SCHEMA kweaver;

ALTER TABLE t_build_task ADD COLUMN IF NOT EXISTS f_catalog_id VARCHAR(40 CHAR) NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_t_build_task_catalog_id ON t_build_task(f_catalog_id);
UPDATE t_build_task bt JOIN t_resource r ON bt.f_resource_id = r.f_id SET bt.f_catalog_id = r.f_catalog_id;

-- ==========================================
-- 迁移脚本：删除未使用的 t_catalog_discover_policy 表
-- ==========================================
DROP TABLE IF EXISTS t_catalog_discover_policy;

-- ==========================================
-- 迁移脚本：t_scheduled_discover_task → t_discover_schedule
-- 表重命名 + 审计字段统一（f_creator_id / f_updater_id → f_creator / f_updater，长度 128 → 40）
-- ==========================================
ALTER TABLE t_scheduled_discover_task RENAME TO t_discover_schedule;
ALTER TABLE t_discover_schedule ALTER COLUMN f_creator_id RENAME TO f_creator;
ALTER TABLE t_discover_schedule ALTER COLUMN f_updater_id RENAME TO f_updater;
ALTER TABLE t_discover_schedule MODIFY f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '';
ALTER TABLE t_discover_schedule MODIFY f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '';

-- ==========================================
-- 迁移脚本：t_discover_task 列改名 + 加索引 + 审计字段长度统一
-- ==========================================
ALTER TABLE t_discover_task ALTER COLUMN f_scheduled_id RENAME TO f_schedule_id;
CREATE INDEX IF NOT EXISTS idx_t_discover_task_schedule_id ON t_discover_task(f_schedule_id);
ALTER TABLE t_discover_task MODIFY f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '';

-- ==========================================
-- 迁移脚本：t_catalog / t_resource 审计字段长度统一（128 → 40）
-- ==========================================
ALTER TABLE t_catalog MODIFY f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '';
ALTER TABLE t_catalog MODIFY f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '';
ALTER TABLE t_resource MODIFY f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '';
ALTER TABLE t_resource MODIFY f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '';

-- ==========================================
-- 迁移脚本：t_build_task 审计字段对齐
-- f_creator_id / f_updater_id → f_creator / f_updater；补 DEFAULT
-- ==========================================
ALTER TABLE t_build_task ALTER COLUMN f_creator_id RENAME TO f_creator;
ALTER TABLE t_build_task ALTER COLUMN f_updater_id RENAME TO f_updater;
ALTER TABLE t_build_task MODIFY f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '';
ALTER TABLE t_build_task MODIFY f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '';
ALTER TABLE t_build_task MODIFY f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '';
ALTER TABLE t_build_task MODIFY f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '';
ALTER TABLE t_build_task MODIFY f_create_time BIGINT NOT NULL DEFAULT 0;
ALTER TABLE t_build_task MODIFY f_update_time BIGINT NOT NULL DEFAULT 0;
