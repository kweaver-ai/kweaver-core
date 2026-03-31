-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS t_build_task (
    -- 主键与关联信息
    f_id                      VARCHAR(40 CHAR) NOT NULL,
    f_resource_id             VARCHAR(40 CHAR) NOT NULL,

    -- 任务状态
    f_status                  VARCHAR(20 CHAR) NOT NULL,
    f_mode                    VARCHAR(20 CHAR) NOT NULL,
    f_total_count             BIGINT NOT NULL DEFAULT 0,
    f_synced_count            BIGINT NOT NULL DEFAULT 0,
    f_vectorized_count        BIGINT NOT NULL DEFAULT 0,
    f_synced_mark             VARCHAR(100 CHAR) DEFAULT NULL,
    f_error_msg               TEXT DEFAULT NULL,

    -- 审计字段
    f_creator_id              VARCHAR(40 CHAR) NOT NULL,
    f_creator_type            VARCHAR(20 CHAR) NOT NULL,
    f_create_time             BIGINT NOT NULL,
    f_updater_id              VARCHAR(40 CHAR) NOT NULL,
    f_updater_type            VARCHAR(20 CHAR) NOT NULL,
    f_update_time             BIGINT NOT NULL,

    -- 索引
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_t_build_task_resource_id ON t_build_task(f_resource_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_t_build_task_status ON t_build_task(f_status);
