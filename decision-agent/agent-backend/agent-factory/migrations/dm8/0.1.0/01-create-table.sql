SET SCHEMA adp;

CREATE TABLE if not exists t_data_agent_config
(
    f_id              VARCHAR(40 CHAR)  not null,
    f_name            VARCHAR(128 CHAR) not null,
    f_profile         VARCHAR(500 CHAR) not null default '',
    f_key             VARCHAR(64 CHAR)  not null,
    f_is_built_in     TINYINT   not null default 0,
    f_is_system_agent TINYINT   not null default 0,
    f_product_key     VARCHAR(40 CHAR)  not null default '',
    f_avatar_type     TINYINT   not null default 0,
    f_avatar          VARCHAR(256 CHAR) not null default '',
    f_created_at      BIGINT       not null default 0,
    f_created_by      VARCHAR(40 CHAR)  not null default '',
    f_updated_at      BIGINT       not null default 0,
    f_updated_by      VARCHAR(40 CHAR)  not null default '',
    f_deleted_by      VARCHAR(36 CHAR)  not null default '',
    f_deleted_at      BIGINT       not null default 0,
    f_config          text         not null,
    f_status          VARCHAR(30 CHAR)  not null default 'unpublished',
    f_created_type    VARCHAR(20 CHAR)  not null default 'create',
    f_create_from     VARCHAR(20 CHAR)  not null default 'dip',
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_config_uk_key_deleted_at ON t_data_agent_config(f_key, f_deleted_at);
CREATE INDEX IF NOT EXISTS t_data_agent_config_idx_created_at ON t_data_agent_config(f_created_at);
CREATE INDEX IF NOT EXISTS t_data_agent_config_idx_deleted_at ON t_data_agent_config(f_deleted_at);
CREATE INDEX IF NOT EXISTS t_data_agent_config_idx_is_built_in ON t_data_agent_config(f_is_built_in);
CREATE INDEX IF NOT EXISTS t_data_agent_config_idx_created_by ON t_data_agent_config(f_created_by);
CREATE INDEX IF NOT EXISTS t_data_agent_config_idx_status ON t_data_agent_config(f_status);



CREATE TABLE if not exists t_data_agent_config_tpl
(
    f_id              BIGINT  not null IDENTITY(1, 1),
    f_name            VARCHAR(128 CHAR) not null,
    f_profile         VARCHAR(256 CHAR) not null default '',
    f_key             VARCHAR(64 CHAR)  not null,
    f_is_built_in     TINYINT   not null default 0,
    f_product_key     VARCHAR(40 CHAR)  not null default '',
    f_avatar_type     TINYINT   not null default 0,
    f_avatar          VARCHAR(256 CHAR) not null default '',
    f_created_at      BIGINT       not null default 0,
    f_created_by      VARCHAR(40 CHAR)  not null default '',
    f_updated_at      BIGINT       not null default 0,
    f_updated_by      VARCHAR(40 CHAR)  not null default '',
    f_deleted_by      VARCHAR(36 CHAR)  not null default '',
    f_deleted_at      BIGINT       not null default 0,
    f_config          text         not null,
    f_status          VARCHAR(30 CHAR)  not null default 'unpublished',
    f_published_at    BIGINT       not null default 0,
    f_published_by    VARCHAR(40 CHAR)  not null default '',
    f_created_type    VARCHAR(20 CHAR)  not null,
    f_create_from     VARCHAR(20 CHAR)  not null default 'dip',
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_config_tpl_uk_key_status_deleted_at ON t_data_agent_config_tpl(f_key, f_status, f_deleted_at);
CREATE INDEX IF NOT EXISTS t_data_agent_config_tpl_idx_updated_at ON t_data_agent_config_tpl(f_updated_at);
CREATE INDEX IF NOT EXISTS t_data_agent_config_tpl_idx_published_at ON t_data_agent_config_tpl(f_published_at);
CREATE INDEX IF NOT EXISTS t_data_agent_config_tpl_idx_deleted_at ON t_data_agent_config_tpl(f_deleted_at);
CREATE INDEX IF NOT EXISTS t_data_agent_config_tpl_idx_is_built_in ON t_data_agent_config_tpl(f_is_built_in);
CREATE INDEX IF NOT EXISTS t_data_agent_config_tpl_idx_created_by ON t_data_agent_config_tpl(f_created_by);
CREATE INDEX IF NOT EXISTS t_data_agent_config_tpl_idx_status ON t_data_agent_config_tpl(f_status);



CREATE TABLE if not exists t_data_agent_config_tpl_published
(
    f_id              BIGINT  not null IDENTITY(1, 1),
    f_tpl_id      BIGINT not null,
    f_name            VARCHAR(128 CHAR) not null,
    f_profile         VARCHAR(256 CHAR) not null default '',
    f_key             VARCHAR(64 CHAR)  not null,
    f_is_built_in     TINYINT   not null default 0,
    f_product_key     VARCHAR(40 CHAR)  not null default '',
    f_avatar_type     TINYINT   not null default 0,
    f_avatar          VARCHAR(256 CHAR) not null default '',
    f_config          text         not null,
    f_published_at    BIGINT       not null default 0,
    f_published_by    VARCHAR(40 CHAR)  not null default '',
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_config_tpl_published_uk_key ON t_data_agent_config_tpl_published(f_key);
CREATE INDEX IF NOT EXISTS t_data_agent_config_tpl_published_idx_published_at ON t_data_agent_config_tpl_published(f_published_at);
CREATE INDEX IF NOT EXISTS t_data_agent_config_tpl_published_idx_is_built_in ON t_data_agent_config_tpl_published(f_is_built_in);
CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_config_tpl_published_uk_tpl_id ON t_data_agent_config_tpl_published(f_tpl_id);



CREATE TABLE if not exists t_data_agent_tpl_category_rel
(
    f_id          BIGINT      not null IDENTITY(1, 1),
    f_published_tpl_id      BIGINT not null,
    f_category_id VARCHAR(40 CHAR) not null,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_tpl_category_rel_uk_published_tpl_id_category_id ON t_data_agent_tpl_category_rel(f_published_tpl_id, f_category_id);
CREATE INDEX IF NOT EXISTS t_data_agent_tpl_category_rel_idx_category_id ON t_data_agent_tpl_category_rel(f_category_id);



CREATE TABLE if not exists t_product
(
    f_id         BIGINT       not null IDENTITY(1, 1),
    f_name       VARCHAR(64 CHAR)  not null,
    f_key        VARCHAR(40 CHAR)  not null,
    f_profile    VARCHAR(256 CHAR) not null default '',
    f_created_by VARCHAR(36 CHAR)  not null default '',
    f_created_at BIGINT       not null default 0,
    f_updated_by VARCHAR(36 CHAR)  not null default '',
    f_updated_at BIGINT       not null default 0,
    f_deleted_by VARCHAR(36 CHAR)  not null default '',
    f_deleted_at BIGINT       not null default 0,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_product_uk_key_deleted_at ON t_product(f_key, f_deleted_at);
CREATE INDEX IF NOT EXISTS t_product_idx_deleted_at ON t_product(f_deleted_at);
CREATE INDEX IF NOT EXISTS t_product_idx_created_at ON t_product(f_created_at);


INSERT INTO t_product (f_name, f_profile, f_key, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by,

                       f_deleted_at)

select 'AnyShare',

       'AnyShare',

       'anyshare',

       '',

       unix_timestamp(sysdate) * 1000,

       '',

       0,

       '',

       0

from dual

where not exists (select 1 from t_product where f_key = 'anyshare');


INSERT INTO t_product (f_name, f_profile, f_key, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by,

                       f_deleted_at)

select 'DIP',

       'DIP',

       'dip',

       '',

       unix_timestamp(sysdate) * 1000,

       '',

       0,

       '',

       0

from dual

where not exists (select 1 from t_product where f_key = 'dip');


INSERT INTO t_product (f_name, f_profile, f_key, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by,

                       f_deleted_at)

select 'ChatBI',

       'ChatBI',

       'chatbi',

       '',

       unix_timestamp(sysdate) * 1000,

       '',

       0,

       '',

       0

from dual

where not exists (select 1 from t_product where f_key = 'chatbi');





CREATE TABLE if not exists t_custom_space
(
    f_id         VARCHAR(40 CHAR)  not null,
    f_name       VARCHAR(64 CHAR)  not null,
    f_key        VARCHAR(64 CHAR)  not null,
    f_profile    VARCHAR(256 CHAR) not null default '',
    f_created_by VARCHAR(36 CHAR)  not null default '',
    f_created_at BIGINT       not null default 0,
    f_updated_by VARCHAR(36 CHAR)  not null default '',
    f_updated_at BIGINT       not null default 0,
    f_deleted_by VARCHAR(36 CHAR)  not null default '',
    f_deleted_at BIGINT       not null default 0,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_custom_space_uk_key_deleted_at ON t_custom_space(f_key, f_deleted_at);
CREATE INDEX IF NOT EXISTS t_custom_space_idx_created_by ON t_custom_space(f_created_by);
CREATE INDEX IF NOT EXISTS t_custom_space_idx_created_at ON t_custom_space(f_created_at);
CREATE INDEX IF NOT EXISTS t_custom_space_idx_updated_at ON t_custom_space(f_updated_at);



CREATE TABLE if not exists t_custom_space_member
(
    f_id         BIGINT      not null IDENTITY(1, 1),
    f_space_id   VARCHAR(40 CHAR) not null,
    f_space_key  VARCHAR(64 CHAR) not null,
    f_obj_type   VARCHAR(32 CHAR) not null,
    f_obj_id     VARCHAR(64 CHAR) not null,
    f_created_at BIGINT      not null default 0,
    f_created_by VARCHAR(36 CHAR) not null default '',
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_custom_space_member_uk_space_id_obj_type_obj_id ON t_custom_space_member(f_space_id, f_obj_type, f_obj_id);



CREATE TABLE if not exists t_custom_space_resource
(
    f_id            BIGINT      not null IDENTITY(1, 1),
    f_space_id      VARCHAR(40 CHAR) not null,
    f_space_key     VARCHAR(64 CHAR) not null,
    f_resource_type VARCHAR(32 CHAR) not null,
    f_resource_id   VARCHAR(64 CHAR) not null,
    f_created_at    BIGINT      not null default 0,
    f_created_by    VARCHAR(36 CHAR) not null default '',
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_custom_space_resource_uk_space_id_resource_type_resource_id ON t_custom_space_resource(f_space_id, f_resource_type, f_resource_id);




CREATE TABLE if not exists t_data_agent_datasource_dataset_assoc
(
    f_id            BIGINT      not null IDENTITY(1, 1),
    f_agent_id      VARCHAR(40 CHAR) not null,
    f_agent_version VARCHAR(32 CHAR) not null,
    f_dataset_id    VARCHAR(40 CHAR) not null,
    f_created_at    BIGINT      not null default 0,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_datasource_dataset_assoc_uk_agent_id_agent_version ON t_data_agent_datasource_dataset_assoc(f_agent_id, f_agent_version);
CREATE INDEX IF NOT EXISTS t_data_agent_datasource_dataset_assoc_idx_dataset_id ON t_data_agent_datasource_dataset_assoc(f_dataset_id);



CREATE TABLE if not exists t_data_agent_datasource_dataset
(
    f_id          VARCHAR(40 CHAR) not null,
    f_hash_sha256 VARCHAR(64 CHAR) not null,
    f_created_at  BIGINT      not null default 0,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE INDEX IF NOT EXISTS t_data_agent_datasource_dataset_idx_hash_sha256 ON t_data_agent_datasource_dataset(f_hash_sha256);



CREATE TABLE if not exists t_data_agent_datasource_dataset_obj
(
    f_id          BIGINT      not null IDENTITY(1, 1),
    f_dataset_id  VARCHAR(40 CHAR) not null,
    f_object_id   VARCHAR(40 CHAR) not null,
    f_object_type VARCHAR(32 CHAR)          default 'dir' not null,
    f_created_at  BIGINT      not null default 0,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_datasource_dataset_obj_uk_dataset_id_object_id_object_type ON t_data_agent_datasource_dataset_obj(f_dataset_id, f_object_type, f_object_id);



CREATE TABLE if not exists t_data_agent_release
(
    f_id                 VARCHAR(40 CHAR)  not null,
    f_agent_id           VARCHAR(40 CHAR)  not null,
    f_agent_name         VARCHAR(128 CHAR) not null,
    f_agent_config       text     not null,
    f_agent_version      VARCHAR(32 CHAR)  not null,
    f_agent_desc         VARCHAR(255 CHAR) not null default '',
    f_is_api_agent       TINYINT   not null default 0,
    f_is_web_sdk_agent   TINYINT   not null default 0,
    f_is_skill_agent     TINYINT   not null default 0,
    f_is_data_flow_agent TINYINT   not null default 0,
    f_is_to_custom_space TINYINT   not null default 0,
    f_is_to_square       TINYINT   not null default 0,
    f_is_pms_ctrl        TINYINT   not null default 0,
    f_create_time        BIGINT       not null default 0,
    f_update_time        BIGINT       not null default 0,
    f_create_by          VARCHAR(40 CHAR)  not null default '',
    f_update_by          VARCHAR(40 CHAR)  not null default '',
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_release_uk_agent_id ON t_data_agent_release(f_agent_id);
CREATE INDEX IF NOT EXISTS t_data_agent_release_idx_update_time ON t_data_agent_release(f_update_time);
CREATE INDEX IF NOT EXISTS t_data_agent_release_idx_is_api_agent ON t_data_agent_release(f_is_api_agent);
CREATE INDEX IF NOT EXISTS t_data_agent_release_idx_is_web_sdk_agent ON t_data_agent_release(f_is_web_sdk_agent);
CREATE INDEX IF NOT EXISTS t_data_agent_release_idx_is_skill_agent ON t_data_agent_release(f_is_skill_agent);
CREATE INDEX IF NOT EXISTS t_data_agent_release_idx_is_data_flow_agent ON t_data_agent_release(f_is_data_flow_agent);
CREATE INDEX IF NOT EXISTS t_data_agent_release_idx_is_to_custom_space ON t_data_agent_release(f_is_to_custom_space);
CREATE INDEX IF NOT EXISTS t_data_agent_release_idx_is_to_square ON t_data_agent_release(f_is_to_square);
CREATE INDEX IF NOT EXISTS t_data_agent_release_idx_is_pms_ctrl ON t_data_agent_release(f_is_pms_ctrl);




CREATE TABLE if not exists t_data_agent_release_history
(
    f_id            VARCHAR(40 CHAR)  not null,
    f_agent_id      VARCHAR(40 CHAR)  not null,
    f_agent_config  text     not null,
    f_agent_version VARCHAR(32 CHAR)  not null,
    f_agent_desc    VARCHAR(255 CHAR) not null default '',
    f_create_time   BIGINT       not null default 0,
    f_update_time   BIGINT       not null default 0,
    f_create_by     VARCHAR(40 CHAR)  not null default '',
    f_update_by     VARCHAR(40 CHAR)  not null default '',
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_release_history_uk_agent_id_agent_version ON t_data_agent_release_history(f_agent_id, f_agent_version);



CREATE TABLE if not exists t_data_agent_release_category
(
    f_id          VARCHAR(40 CHAR)  not null,
    f_name        VARCHAR(128 CHAR) not null,
    f_description VARCHAR(256 CHAR) not null default '',
    f_create_time BIGINT       not null default 0,
    f_update_time BIGINT       not null default 0,
    f_create_by   VARCHAR(40 CHAR)  not null default '',
    f_update_by   VARCHAR(40 CHAR)  not null default '',
    CLUSTER PRIMARY KEY (f_id)
);




CREATE TABLE if not exists t_data_agent_release_category_rel
(
    f_id          VARCHAR(40 CHAR) not null,
    f_release_id  VARCHAR(40 CHAR) not null,
    f_category_id VARCHAR(40 CHAR) not null,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_release_category_rel_uk_release_id_category_id ON t_data_agent_release_category_rel(f_release_id, f_category_id);




CREATE TABLE if not exists t_data_agent_release_permission
(
    f_id         BIGINT      not null IDENTITY(1, 1),
    f_release_id VARCHAR(40 CHAR) not null,
    f_obj_type   VARCHAR(32 CHAR) not null,
    f_obj_id     VARCHAR(64 CHAR) not null,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_release_permission_uk_release_id_obj_type_obj_id ON t_data_agent_release_permission(f_release_id, f_obj_type, f_obj_id);





CREATE TABLE if not exists t_data_agent_visit_history
(
    f_id            VARCHAR(40 CHAR) not null,
    f_agent_id      VARCHAR(40 CHAR) not null,
    f_agent_version VARCHAR(32 CHAR) not null,
    f_custom_space_id VARCHAR(40 CHAR) not null default '',
    f_visit_count   INT         not null default 1,
    f_create_time   BIGINT      not null default 0,
    f_update_time   BIGINT      not null default 0,
    f_create_by     VARCHAR(40 CHAR) not null default '',
    f_update_by     VARCHAR(40 CHAR) not null default '',
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_data_agent_visit_history_uk_user_agent ON t_data_agent_visit_history(f_create_by, f_agent_id, f_agent_version);
CREATE INDEX IF NOT EXISTS t_data_agent_visit_history_idx_custom_space_id ON t_data_agent_visit_history(f_custom_space_id);




CREATE TABLE if not exists t_biz_domain_agent_rel
(
    f_id            BIGINT      not null IDENTITY(1, 1),
    f_biz_domain_id VARCHAR(40 CHAR) not null,
    f_agent_id      VARCHAR(40 CHAR) not null,
    f_created_at    BIGINT      not null default 0,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_biz_domain_agent_rel_uk_biz_domain_id_agent_id ON t_biz_domain_agent_rel(f_biz_domain_id, f_agent_id);
CREATE INDEX IF NOT EXISTS t_biz_domain_agent_rel_idx_agent_id ON t_biz_domain_agent_rel(f_agent_id);



CREATE TABLE if not exists t_biz_domain_agent_tpl_rel
(
    f_id            BIGINT      not null IDENTITY(1, 1),
    f_biz_domain_id VARCHAR(40 CHAR) not null,
    f_agent_tpl_id  BIGINT      not null,
    f_created_at    BIGINT      not null default 0,
    CLUSTER PRIMARY KEY (f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_biz_domain_agent_tpl_rel_uk_biz_domain_id_agent_tpl_id ON t_biz_domain_agent_tpl_rel(f_biz_domain_id, f_agent_tpl_id);
CREATE INDEX IF NOT EXISTS t_biz_domain_agent_tpl_rel_idx_agent_tpl_id ON t_biz_domain_agent_tpl_rel(f_agent_tpl_id);




INSERT INTO t_data_agent_release_category(f_id, f_name, f_description)

SELECT '01JRYRKP0M8VYHQSX4FXR5CKGX', '情报分析', '情报分析'

FROM DUAL

WHERE NOT EXISTS (SELECT f_id FROM t_data_agent_release_category WHERE f_id = '01JRYRKP0M8VYHQSX4FXR5CKGX');


INSERT INTO t_data_agent_release_category(f_id, f_name, f_description)

SELECT '01JRYRKP0M8VYHQSX4FXR5CKGY', '智能洞察', '智能洞察'

FROM DUAL

WHERE NOT EXISTS (SELECT f_id FROM t_data_agent_release_category WHERE f_id = '01JRYRKP0M8VYHQSX4FXR5CKGY');


INSERT INTO t_data_agent_release_category(f_id, f_name, f_description)

SELECT '01JRYRKP0M8VYHQSX4FXR5CKGZ', '分析助手', '分析助手'

FROM DUAL

WHERE NOT EXISTS (SELECT f_id FROM t_data_agent_release_category WHERE f_id = '01JRYRKP0M8VYHQSX4FXR5CKGZ');


INSERT INTO t_data_agent_release_category(f_id, f_name, f_description)

SELECT '01JRYRKP0M8VYHQSX4FXR5CKG1', '辅助阅读', '辅助阅读'

FROM DUAL

WHERE NOT EXISTS (SELECT f_id FROM t_data_agent_release_category WHERE f_id = '01JRYRKP0M8VYHQSX4FXR5CKG1');


INSERT INTO t_data_agent_release_category(f_id, f_name, f_description)

SELECT '01JRYRKP0M8VYHQSX4FXR5CKG2', '事件感知', '事件感知'

FROM DUAL

WHERE NOT EXISTS (SELECT f_id FROM t_data_agent_release_category WHERE f_id = '01JRYRKP0M8VYHQSX4FXR5CKG2');


INSERT INTO t_data_agent_release_category(f_id, f_name, f_description)

SELECT '01JRYRKP0M8VYHQSX4FXR5CKG3', '报告生成', '报告生成'

FROM DUAL

WHERE NOT EXISTS (SELECT f_id FROM t_data_agent_release_category WHERE f_id = '01JRYRKP0M8VYHQSX4FXR5CKG3');


INSERT INTO t_data_agent_release_category(f_id, f_name, f_description)

SELECT '01JRYRKP0M8VYHQSX4FXR5CKG4', '辅助决策', '辅助决策'

FROM DUAL

WHERE NOT EXISTS (SELECT f_id FROM t_data_agent_release_category WHERE f_id = '01JRYRKP0M8VYHQSX4FXR5CKG4');


INSERT INTO t_data_agent_release_category(f_id, f_name, f_description)

SELECT '01JRYRKP0M8VYHQSX4FXR5CKG5', '数据处理', '数据处理'

FROM DUAL

WHERE NOT EXISTS (SELECT f_id FROM t_data_agent_release_category WHERE f_id = '01JRYRKP0M8VYHQSX4FXR5CKG5');


