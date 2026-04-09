USE adp;

create table if not exists t_storage_config
(
    f_storage_id        varchar(50)       not null comment 'Storage ID (Snowflake ID)',
    f_storage_name      varchar(128)      not null comment 'Storage name',
    f_vendor_type       varchar(32)       not null comment 'Vendor type: OSS/OBS/ECEPH',
    f_endpoint          varchar(256)      not null comment 'Service endpoint URL',
    f_bucket_name       varchar(128)      not null comment 'Bucket name',
    f_access_key_id     varchar(256)      not null comment 'AccessKeyID (encrypted)',
    f_access_key        varchar(512)      not null comment 'AccessKeySecret (encrypted)',
    f_region            varchar(64)       default '' null comment 'Region (required for OSS/OBS, optional for ECEPH)',
    f_is_default        int(11)           default 0 null comment 'Is default storage',
    f_is_enabled        int(11)           default 1 null comment 'Is enabled',
    f_internal_endpoint varchar(256)      default '' null comment 'Internal access endpoint',
    f_site_id           varchar(64)       default '' null comment 'Site ID for multi-tenant isolation',
    f_created_at        datetime(6)       null comment 'Creation time',
    f_updated_at        datetime(6)       null comment 'Update time',
    primary key (f_storage_id)
);

create table if not exists t_multipart_upload_task
(
    f_id          varchar(50)       not null comment 'Task ID (Snowflake ID)',
    f_storage_id  varchar(50)       not null comment 'Associated storage ID',
    f_object_key  varchar(512)      not null comment 'Object key',
    f_upload_id   varchar(256)      not null comment 'Upload ID from vendor',
    f_total_size  bigint            not null comment 'Total file size',
    f_part_size   int               not null comment 'Part size in bytes',
    f_total_parts int               not null comment 'Total number of parts',
    f_status      smallint          default 0 null comment 'Status: 0=in progress, 1=completed, 2=cancelled',
    f_created_at  datetime(6)       null comment 'Creation time',
    f_expires_at  datetime(6)       not null comment 'Expiration time',
    primary key (f_id),
    key idx_storage_id (f_storage_id),
    key idx_status (f_status),
    key idx_expires_at (f_expires_at)
);
