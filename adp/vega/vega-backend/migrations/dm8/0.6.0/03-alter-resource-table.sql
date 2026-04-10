-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：修改 t_resource 表的索引结构
-- ==========================================

SET SCHEMA kweaver;

-- ==========================================
-- 9. t_scheduled_discover_task 定时发现任务表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_scheduled_discover_task (
    -- 主键与关联信息
    f_id                      VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_catalog_id              VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_cron_expr               VARCHAR(100 CHAR) NOT NULL DEFAULT '',

    -- 时间配置
    f_start_time              BIGINT NOT NULL DEFAULT 0,
    f_end_time                BIGINT NOT NULL DEFAULT 0,

    -- 任务状态
    f_enabled                 TINYINT NOT NULL DEFAULT 0,
    f_strategies              VARCHAR(100 CHAR) NOT NULL DEFAULT '',

    f_last_run                BIGINT NOT NULL DEFAULT 0,
    f_next_run                BIGINT NOT NULL DEFAULT 0,

    -- 审计字段
    f_creator_id              VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_creator_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_create_time             BIGINT NOT NULL DEFAULT 0,
    f_updater_id              VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_updater_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_update_time             BIGINT NOT NULL DEFAULT 0,
    -- 索引
    CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS idx_t_scheduled_discover_task_catalog_id ON t_scheduled_discover_task (f_catalog_id);
CREATE INDEX IF NOT EXISTS idx_t_scheduled_discover_task_enabled ON t_scheduled_discover_task (f_enabled);
CREATE INDEX IF NOT EXISTS idx_t_scheduled_discover_task_next_run ON t_scheduled_discover_task (f_next_run);


