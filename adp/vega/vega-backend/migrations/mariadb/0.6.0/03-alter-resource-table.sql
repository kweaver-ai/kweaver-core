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

USE kweaver;

-- 删除原有的 uk_catalog_name 唯一索引
DROP INDEX uk_catalog_name ON t_resource;

-- 添加新的 uk_catalog_source_identifier 唯一索引
CREATE UNIQUE INDEX uk_catalog_source_identifier ON t_resource (f_catalog_id, f_source_identifier);

ALTER TABLE t_discover_task ADD COLUMN IF NOT EXISTS f_scheduled_id varchar(40) DEFAULT NULL AFTER f_catalog_id;
ALTER TABLE t_discover_task ADD COLUMN IF NOT EXISTS f_strategies varchar(100) DEFAULT NULL AFTER f_scheduled_id;

CREATE TABLE IF NOT EXISTS t_scheduled_discover_task (
    -- 主键与关联信息
    f_id                      VARCHAR(40) NOT NULL DEFAULT '' COMMENT '任务唯一标识',
    f_catalog_id              VARCHAR(40) NOT NULL DEFAULT '' COMMENT '所属catalog ID',
    f_cron_expr               VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'Cron表达式',
    -- 时间配置
    f_start_time              BIGINT(20) NOT NULL DEFAULT 0 COMMENT '开始时间（Unix毫秒时间戳）',
    f_end_time                BIGINT(20) NOT NULL DEFAULT 0 COMMENT '结束时间（Unix毫秒时间戳），0表示无结束时间',

    -- 任务状态
    f_enabled                 TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否启用: 0-禁用, 1-启用',
    f_strategies              VARCHAR(100) NOT NULL DEFAULT '' COMMENT '策略',

    f_last_run                BIGINT(20) NOT NULL DEFAULT 0 COMMENT '最后执行时间（Unix毫秒时间戳）',
    f_next_run                BIGINT(20) NOT NULL DEFAULT 0 COMMENT '下次执行时间（Unix毫秒时间戳）',

    -- 审计字段
    f_creator_id              VARCHAR(128) NOT NULL DEFAULT '' COMMENT '创建者id',
    f_creator_type            VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
    f_create_time             BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
    f_updater_id              VARCHAR(128) NOT NULL DEFAULT '' COMMENT '更新者id',
    f_updater_type            VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
    f_update_time             BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',

    -- 索引
    PRIMARY KEY (f_id),
    INDEX idx_catalog_id (f_catalog_id),
    INDEX idx_enabled (f_enabled),
    INDEX idx_next_run (f_next_run)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT='定时发现任务表，记录定时资源发现任务的配置和执行状态';
