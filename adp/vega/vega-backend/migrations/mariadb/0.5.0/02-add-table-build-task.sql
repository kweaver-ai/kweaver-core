-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

USE adp;

CREATE TABLE IF NOT EXISTS t_build_task (
    -- 主键与关联信息
    f_id                      VARCHAR(40) NOT NULL COMMENT '任务ID',
    f_resource_id             VARCHAR(40) NOT NULL COMMENT '资源ID',

    -- 任务状态
    f_status                  VARCHAR(20) NOT NULL COMMENT '任务状态: pending, running, completed, failed',
    f_mode                    VARCHAR(20) NOT NULL COMMENT '任务模式: full, incremental, realtime',
    f_total_count             BIGINT NOT NULL DEFAULT 0 COMMENT '总数',
    f_synced_count            BIGINT NOT NULL DEFAULT 0 COMMENT '已同步数',
    f_vectorized_count        BIGINT NOT NULL DEFAULT 0 COMMENT '已做向量数',
    f_synced_mark             VARCHAR(100) DEFAULT NULL COMMENT '同步标记',
    f_error_msg               TEXT DEFAULT NULL COMMENT '错误信息',

    -- 审计字段
    f_creator_id              VARCHAR(40) NOT NULL COMMENT '创建人ID',
    f_creator_type            VARCHAR(20) NOT NULL COMMENT '创建人类型',
    f_create_time             BIGINT NOT NULL COMMENT '创建时间',
    f_updater_id              VARCHAR(40) NOT NULL COMMENT '更新人ID',
    f_updater_type            VARCHAR(20) NOT NULL COMMENT '更新人类型',
    f_update_time             BIGINT NOT NULL COMMENT '更新时间',

    -- 索引
    PRIMARY KEY (f_id),
    INDEX idx_resource_id (f_resource_id),
    INDEX idx_status (f_status),
    INDEX idx_create_time (f_create_time)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT='构建任务表';