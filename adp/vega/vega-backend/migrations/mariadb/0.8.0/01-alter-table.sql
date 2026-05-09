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

-- ==========================================
-- 迁移脚本：删除未使用的 t_catalog_discover_policy 表
-- ==========================================
DROP TABLE IF EXISTS t_catalog_discover_policy;

-- ==========================================
-- 迁移脚本：t_scheduled_discover_task → t_discover_schedule
-- 表重命名 + 审计字段统一（f_creator_id / f_updater_id → f_creator / f_updater，长度 128 → 40）
-- ==========================================
RENAME TABLE t_scheduled_discover_task TO t_discover_schedule;
ALTER TABLE t_discover_schedule CHANGE COLUMN f_creator_id f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id';
ALTER TABLE t_discover_schedule CHANGE COLUMN f_updater_id f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id';

-- ==========================================
-- 迁移脚本：t_discover_task 列改名 + 加索引 + 审计字段长度统一
-- ==========================================
ALTER TABLE t_discover_task CHANGE COLUMN f_scheduled_id f_schedule_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '关联的 DiscoverSchedule ID';
ALTER TABLE t_discover_task ADD INDEX IF NOT EXISTS idx_schedule_id (f_schedule_id);
ALTER TABLE t_discover_task MODIFY COLUMN f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id';

-- ==========================================
-- 迁移脚本：t_catalog / t_resource 审计字段长度统一（128 → 40）
-- ==========================================
ALTER TABLE t_catalog MODIFY COLUMN f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id';
ALTER TABLE t_catalog MODIFY COLUMN f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id';
ALTER TABLE t_resource MODIFY COLUMN f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id';
ALTER TABLE t_resource MODIFY COLUMN f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id';

-- ==========================================
-- 迁移脚本：t_build_task 审计字段对齐
-- f_creator_id / f_updater_id → f_creator / f_updater
-- 时间字段补 BIGINT(20) DEFAULT 0
-- ==========================================
ALTER TABLE t_build_task CHANGE COLUMN f_creator_id f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id';
ALTER TABLE t_build_task CHANGE COLUMN f_updater_id f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id';
ALTER TABLE t_build_task MODIFY COLUMN f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型';
ALTER TABLE t_build_task MODIFY COLUMN f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型';
ALTER TABLE t_build_task MODIFY COLUMN f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间';
ALTER TABLE t_build_task MODIFY COLUMN f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间';
