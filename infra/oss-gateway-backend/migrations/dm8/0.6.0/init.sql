SET SCHEMA kweaver;

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
