-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：在 kweaver schema 下创建 vega 相关表，并从 adp schema 复制数据
-- 使用 WHERE NOT EXISTS 判重逻辑，避免插入重复记录（基于主键判断）
-- ==========================================

SET SCHEMA kweaver;

-- ==========================================
-- 1. t_catalog 主表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_catalog (
    f_id                      VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_name                    VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    f_tags                    VARCHAR(255 CHAR) NOT NULL DEFAULT '[]',
    f_description             VARCHAR(1000 CHAR) NOT NULL DEFAULT '',

    f_type                    VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_enabled                 TINYINT NOT NULL DEFAULT 1,

    f_connector_type          VARCHAR(50 CHAR) NOT NULL DEFAULT '',
    f_connector_config        TEXT NOT NULL,
    f_metadata                TEXT NOT NULL,

    f_health_check_enabled    TINYINT NOT NULL DEFAULT 1,
    f_health_check_status     VARCHAR(20 CHAR) NOT NULL DEFAULT 'healthy',
    f_last_check_time         BIGINT NOT NULL DEFAULT 0,
    f_health_check_result     TEXT NOT NULL,

    f_creator                 VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_creator_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_create_time             BIGINT NOT NULL DEFAULT 0,
    f_updater                 VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_updater_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_update_time             BIGINT NOT NULL DEFAULT 0,

    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_catalog_name ON t_catalog(f_name);
CREATE INDEX IF NOT EXISTS idx_t_catalog_type ON t_catalog(f_type);
CREATE INDEX IF NOT EXISTS idx_t_catalog_connector_type ON t_catalog(f_connector_type);
CREATE INDEX IF NOT EXISTS idx_t_catalog_health_check_status ON t_catalog(f_health_check_status);

INSERT INTO kweaver."t_catalog" (
    "f_id", "f_name", "f_tags", "f_description",
    "f_type", "f_enabled",
    "f_connector_type", "f_connector_config", "f_metadata",
    "f_health_check_enabled", "f_health_check_status", "f_last_check_time", "f_health_check_result",
    "f_creator", "f_creator_type", "f_create_time",
    "f_updater", "f_updater_type", "f_update_time"
)
SELECT
    s."f_id", s."f_name", s."f_tags", s."f_description",
    s."f_type", s."f_enabled",
    s."f_connector_type", s."f_connector_config", s."f_metadata",
    s."f_health_check_enabled", s."f_health_check_status", s."f_last_check_time", s."f_health_check_result",
    s."f_creator", s."f_creator_type", s."f_create_time",
    s."f_updater", s."f_updater_type", s."f_update_time"
FROM adp."t_catalog" s;

-- ==========================================
-- 2. t_catalog_discover_policy 发现与变更策略表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_catalog_discover_policy (
    f_id                      VARCHAR(40 CHAR) NOT NULL DEFAULT '',

    f_enabled                 TINYINT NOT NULL DEFAULT 0,

    f_discover_mode          VARCHAR(20 CHAR) NOT NULL DEFAULT 'manual',
    f_discover_cron          VARCHAR(100 CHAR) NOT NULL DEFAULT '',
    f_discover_config        TEXT NOT NULL,

    f_on_resource_added       VARCHAR(20 CHAR) NOT NULL DEFAULT 'auto_register',
    f_on_resource_removed     VARCHAR(20 CHAR) NOT NULL DEFAULT 'mark_stale',
    f_on_schema_changed       VARCHAR(20 CHAR) NOT NULL DEFAULT 'auto_update',
    f_on_file_content_changed VARCHAR(20 CHAR) NOT NULL DEFAULT 'pending_review',

    f_change_policy_config    TEXT NOT NULL,

    CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS idx_t_catalog_discover_policy_discover_mode ON t_catalog_discover_policy(f_discover_mode);
CREATE INDEX IF NOT EXISTS idx_t_catalog_discover_policy_enabled ON t_catalog_discover_policy(f_enabled);

INSERT INTO kweaver."t_catalog_discover_policy" (
    "f_id", "f_enabled",
    "f_discover_mode", "f_discover_cron", "f_discover_config",
    "f_on_resource_added", "f_on_resource_removed", "f_on_schema_changed", "f_on_file_content_changed",
    "f_change_policy_config"
)
SELECT
    s."f_id", s."f_enabled",
    s."f_discover_mode", s."f_discover_cron", s."f_discover_config",
    s."f_on_resource_added", s."f_on_resource_removed", s."f_on_schema_changed", s."f_on_file_content_changed",
    s."f_change_policy_config"
FROM adp."t_catalog_discover_policy" s;

-- ==========================================
-- 3. t_resource 数据资源主表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_resource (
    f_id                      VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_catalog_id              VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_name                    VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    f_tags                    VARCHAR(255 CHAR) NOT NULL DEFAULT '[]',
    f_description             VARCHAR(1000 CHAR) NOT NULL DEFAULT '',

    f_category                VARCHAR(20 CHAR) NOT NULL DEFAULT '',

    f_status                  VARCHAR(20 CHAR) NOT NULL DEFAULT 'active',
    f_status_message          VARCHAR(500 CHAR) NOT NULL DEFAULT '',

    f_database                VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_source_identifier       VARCHAR(500 CHAR) NOT NULL DEFAULT '',
    f_source_metadata         TEXT NOT NULL,

    f_schema_definition       TEXT NOT NULL,

    f_logic_type              VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_logic_definition        TEXT NOT NULL,

    f_local_enabled           TINYINT NOT NULL DEFAULT 0,
    f_local_storage_engine    VARCHAR(50 CHAR) NOT NULL DEFAULT '',
    f_local_storage_config    TEXT NOT NULL,
    f_local_index_name        VARCHAR(255 CHAR) NOT NULL DEFAULT '',

    f_sync_strategy           VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_sync_config             TEXT NOT NULL,
    f_sync_status             VARCHAR(20 CHAR) NOT NULL DEFAULT 'not_synced',
    f_last_sync_time          BIGINT NOT NULL DEFAULT 0,
    f_sync_error_message      TEXT NOT NULL,

    f_creator                 VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_creator_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_create_time             BIGINT NOT NULL DEFAULT 0,
    f_updater                 VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_updater_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_update_time             BIGINT NOT NULL DEFAULT 0,

    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_t_resource_catalog_source_identifier ON t_resource(f_catalog_id, f_source_identifier);
CREATE INDEX IF NOT EXISTS idx_t_resource_category ON t_resource(f_category);
CREATE INDEX IF NOT EXISTS idx_t_resource_status ON t_resource(f_status);

INSERT INTO kweaver."t_resource" (
    "f_id", "f_catalog_id", "f_name", "f_tags", "f_description",
    "f_category",
    "f_status", "f_status_message",
    "f_database", "f_source_identifier", "f_source_metadata",
    "f_schema_definition",
    "f_logic_type", "f_logic_definition",
    "f_local_enabled", "f_local_storage_engine", "f_local_storage_config", "f_local_index_name",
    "f_sync_strategy", "f_sync_config", "f_sync_status", "f_last_sync_time", "f_sync_error_message",
    "f_creator", "f_creator_type", "f_create_time",
    "f_updater", "f_updater_type", "f_update_time"
)
SELECT
    s."f_id", s."f_catalog_id", s."f_name", s."f_tags", s."f_description",
    s."f_category",
    s."f_status", s."f_status_message",
    s."f_database", s."f_source_identifier", s."f_source_metadata",
    s."f_schema_definition",
    s."f_logic_type", s."f_logic_definition",
    s."f_local_enabled", s."f_local_storage_engine", s."f_local_storage_config", s."f_local_index_name",
    s."f_sync_strategy", s."f_sync_config", s."f_sync_status", s."f_last_sync_time", s."f_sync_error_message",
    s."f_creator", s."f_creator_type", s."f_create_time",
    s."f_updater", s."f_updater_type", s."f_update_time"
FROM adp."t_resource" s;

-- ==========================================
-- 4. t_resource_schema_history Schema历史表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_resource_schema_history (
    f_id                      VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_resource_id             VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_schema_version          VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_schema_definition       TEXT NOT NULL,

    f_change_type             VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_change_summary          VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
    f_schema_inferred         TINYINT NOT NULL DEFAULT 0,
    f_change_time             BIGINT NOT NULL DEFAULT 0,

    CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS idx_t_resource_schema_history_resource_id ON t_resource_schema_history(f_resource_id);

INSERT INTO kweaver."t_resource_schema_history" (
    "f_id", "f_resource_id", "f_schema_version", "f_schema_definition",
    "f_change_type", "f_change_summary", "f_schema_inferred", "f_change_time"
)
SELECT
    s."f_id", s."f_resource_id", s."f_schema_version", s."f_schema_definition",
    s."f_change_type", s."f_change_summary", s."f_schema_inferred", s."f_change_time"
FROM adp."t_resource_schema_history" s;

-- ==========================================
-- 5. t_connector_type Connector 类型注册表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_connector_type (
    f_type                    VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_name                    VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    f_tags                    VARCHAR(255 CHAR) NOT NULL DEFAULT '[]',
    f_description             VARCHAR(1000 CHAR) NOT NULL DEFAULT '',

    f_mode                    VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_category                VARCHAR(32 CHAR) NOT NULL DEFAULT '',

    f_endpoint                VARCHAR(512 CHAR) NOT NULL DEFAULT '',

    f_field_config            TEXT NOT NULL,

    f_enabled                 TINYINT NOT NULL DEFAULT 1,

    CLUSTER PRIMARY KEY (f_type)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_connector_type_name ON t_connector_type(f_name);
CREATE INDEX IF NOT EXISTS idx_t_connector_type_mode ON t_connector_type(f_mode);
CREATE INDEX IF NOT EXISTS idx_t_connector_type_category ON t_connector_type(f_category);
CREATE INDEX IF NOT EXISTS idx_t_connector_type_enabled ON t_connector_type(f_enabled);

INSERT INTO kweaver."t_connector_type" (
    "f_type", "f_name", "f_tags", "f_description",
    "f_mode", "f_category",
    "f_endpoint", "f_field_config", "f_enabled"
)
SELECT
    s."f_type", s."f_name", s."f_tags", s."f_description",
    s."f_mode", s."f_category",
    s."f_endpoint", s."f_field_config", s."f_enabled"
FROM adp."t_connector_type" s;

-- ==========================================
-- 6. t_discover_task 发现任务表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_discover_task (
    f_id                      VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_catalog_id              VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_scheduled_id            VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_strategies              VARCHAR(100 CHAR) NOT NULL DEFAULT '',
    f_trigger_type            VARCHAR(20 CHAR) NOT NULL DEFAULT 'manual',

    f_status                  VARCHAR(20 CHAR) NOT NULL DEFAULT 'pending',
    f_progress                INT NOT NULL DEFAULT 0,
    f_message                 VARCHAR(1000 CHAR) NOT NULL DEFAULT '',

    f_start_time              BIGINT NOT NULL DEFAULT 0,
    f_finish_time             BIGINT NOT NULL DEFAULT 0,

    f_result                  TEXT NOT NULL,

    f_creator                 VARCHAR(128 CHAR) NOT NULL DEFAULT '',
    f_creator_type            VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_create_time             BIGINT NOT NULL DEFAULT 0,

    CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS idx_t_discover_task_catalog_id ON t_discover_task (f_catalog_id);
CREATE INDEX IF NOT EXISTS idx_t_discover_task_status ON t_discover_task (f_status);

INSERT INTO kweaver."t_discover_task" (
    "f_id", "f_catalog_id", "f_trigger_type",
    "f_status", "f_progress", "f_message",
    "f_start_time", "f_finish_time",
    "f_result",
    "f_creator", "f_creator_type", "f_create_time"
)
SELECT
    s."f_id", s."f_catalog_id", s."f_trigger_type",
    s."f_status", s."f_progress", s."f_message",
    s."f_start_time", s."f_finish_time",
    s."f_result",
    s."f_creator", s."f_creator_type", s."f_create_time"
FROM adp."t_discover_task" s;

-- ==========================================
-- 7. t_build_task 构建任务表
-- ==========================================
CREATE TABLE IF NOT EXISTS t_build_task (
    f_id                      VARCHAR(40 CHAR) NOT NULL,
    f_resource_id             VARCHAR(40 CHAR) NOT NULL,

    f_status                  VARCHAR(20 CHAR) NOT NULL,
    f_mode                    VARCHAR(20 CHAR) NOT NULL,
    f_total_count             BIGINT NOT NULL DEFAULT 0,
    f_synced_count            BIGINT NOT NULL DEFAULT 0,
    f_vectorized_count        BIGINT NOT NULL DEFAULT 0,
    f_synced_mark             VARCHAR(100 CHAR) DEFAULT NULL,
    f_error_msg               TEXT DEFAULT NULL,

    f_creator_id              VARCHAR(40 CHAR) NOT NULL,
    f_creator_type            VARCHAR(20 CHAR) NOT NULL,
    f_create_time             BIGINT NOT NULL,
    f_updater_id              VARCHAR(40 CHAR) NOT NULL,
    f_updater_type            VARCHAR(20 CHAR) NOT NULL,
    f_update_time             BIGINT NOT NULL,

    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_t_build_task_resource_id ON t_build_task(f_resource_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_t_build_task_status ON t_build_task(f_status);

INSERT INTO kweaver."t_build_task" (
    "f_id", "f_resource_id",
    "f_status", "f_mode",
    "f_total_count", "f_synced_count", "f_vectorized_count",
    "f_synced_mark", "f_error_msg",
    "f_creator_id", "f_creator_type", "f_create_time",
    "f_updater_id", "f_updater_type", "f_update_time"
)
SELECT
    s."f_id", s."f_resource_id",
    s."f_status", s."f_mode",
    s."f_total_count", s."f_synced_count", s."f_vectorized_count",
    s."f_synced_mark", s."f_error_msg",
    s."f_creator_id", s."f_creator_type", s."f_create_time",
    s."f_updater_id", s."f_updater_type", s."f_update_time"
FROM adp."t_build_task" s;
