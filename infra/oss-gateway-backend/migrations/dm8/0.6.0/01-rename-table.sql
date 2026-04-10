-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：在 kweaver schema 下创建 oss-gateway-backend 相关表，并从 adp schema 复制数据
-- ==========================================

SET SCHEMA kweaver;

-- ==========================================
-- 1. t_storage_config
-- ==========================================
CREATE TABLE if not exists t_storage_config
(
    f_storage_id        VARCHAR(50 CHAR)      not null,
    f_storage_name      VARCHAR(128 CHAR)     not null,
    f_vendor_type       VARCHAR(32 CHAR)      not null,
    f_endpoint          VARCHAR(256 CHAR)     not null,
    f_bucket_name       VARCHAR(128 CHAR)     not null,
    f_access_key_id     VARCHAR(256 CHAR)     not null,
    f_access_key        VARCHAR(512 CHAR)     not null,
    f_region            VARCHAR(64 CHAR)      default '' null,
    f_is_default        INT                   default 0 null,
    f_is_enabled        INT                   default 1 null,
    f_internal_endpoint VARCHAR(256 CHAR)     default '' null,
    f_site_id           VARCHAR(64 CHAR)      default '' null,
    f_created_at        datetime(6)           null,
    f_updated_at        datetime(6)           null,
    CLUSTER PRIMARY KEY (f_storage_id)
);

INSERT INTO kweaver."t_storage_config" (
    f_storage_id, f_storage_name, f_vendor_type,
    f_endpoint, f_bucket_name, f_access_key_id, f_access_key,
    f_region, f_is_default, f_is_enabled,
    f_internal_endpoint, f_site_id, f_created_at, f_updated_at
)
SELECT
    f_storage_id, f_storage_name, f_vendor_type,
    f_endpoint, f_bucket_name, f_access_key_id, f_access_key,
    f_region, f_is_default, f_is_enabled,
    f_internal_endpoint, f_site_id, f_created_at, f_updated_at
FROM adp."t_storage_config" s;

-- ==========================================
-- 2. t_multipart_upload_task
-- ==========================================
CREATE TABLE if not exists t_multipart_upload_task
(
    f_id          VARCHAR(50 CHAR)      not null,
    f_storage_id  VARCHAR(50 CHAR)      not null,
    f_object_key  VARCHAR(512 CHAR)     not null,
    f_upload_id   VARCHAR(256 CHAR)     not null,
    f_total_size  BIGINT                not null,
    f_part_size   INT                   not null,
    f_total_parts INT                   not null,
    f_status      SMALLINT              default 0 null,
    f_created_at  datetime(6)           null,
    f_expires_at  datetime(6)           not null,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS idx_storage_id ON t_multipart_upload_task(f_storage_id);
CREATE INDEX IF NOT EXISTS idx_status ON t_multipart_upload_task(f_status);
CREATE INDEX IF NOT EXISTS idx_expires_at ON t_multipart_upload_task(f_expires_at);

INSERT INTO kweaver."t_multipart_upload_task" (
    f_id, f_storage_id, f_object_key, f_upload_id,
    f_total_size, f_part_size, f_total_parts,
    f_status, f_created_at, f_expires_at
)
SELECT
    f_id, f_storage_id, f_object_key, f_upload_id,
    f_total_size, f_part_size, f_total_parts,
    f_status, f_created_at, f_expires_at
FROM adp."t_multipart_upload_task" s;
