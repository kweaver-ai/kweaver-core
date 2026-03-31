-- Decision Agent 数据库初始化脚本
-- 创建数据库 adp（如果不存在）
CREATE DATABASE IF NOT EXISTS `adp` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `adp`;

-- 记忆历史表
CREATE TABLE IF NOT EXISTS `t_data_agent_memory_history` (
  `f_id`          varchar(40) NOT NULL COMMENT 'ID，唯一标识',
  `f_memory_id`   varchar(40) NOT NULL COMMENT '记忆ID',
  `f_old_memory`  text COMMENT '旧的记忆',
  `f_new_memory`  text  COMMENT '新的记忆',
  `f_event`       varchar(40) NOT NULL  COMMENT '对记忆的操作: ADD, UPDATE, DELETE',
  `f_created_at`  varchar(40) NOT NULL DEFAULT '' COMMENT '创建时间',
  `f_updated_at`  varchar(40) NOT NULL DEFAULT '' COMMENT '更新时间',
  `f_actor_id`    varchar(40) NOT NULL DEFAULT '' COMMENT '操作者',
  `f_role`        varchar(40) NOT NULL DEFAULT '' COMMENT '操作者角色',
  `f_create_by`   varchar(40) NOT NULL DEFAULT '' COMMENT '创建者',
  `f_update_by`   varchar(40) NOT NULL DEFAULT '' COMMENT '最后修改者',
  `f_is_deleted`  tinyint NOT NULL DEFAULT 0 COMMENT '是否删除：0 - 否 1 - 是',
  PRIMARY KEY (`f_id`),
  KEY `idx_memory_id` (`f_memory_id`)
) ENGINE=InnoDB COMMENT='记忆历史数据表';



-- data agent配置表 start
create table if not exists t_data_agent_config
(
    f_id              varchar(40)  not null comment 'id',
#     f_seq_no          bigint       not null auto_increment comment '自增序号' unique,
    f_name            varchar(128) not null comment 'agent名',
    f_profile         varchar(500) not null default '' comment '简介',

    f_key             varchar(64)  not null comment '唯一标识，支持自动生成和人工设置',
    f_is_built_in     tinyint(4)   not null default 0 comment '是否是内置的，枚举值：0-否，1-是',
    f_is_system_agent tinyint(4)   not null default 0 comment '是否是系统agent，枚举值：0-否，1-是',

    f_product_key     varchar(40)  not null default '' comment '产品标识',

    f_avatar_type     tinyint(4)   not null default 0 comment '头像类型，0-内置头像，1-用户上传头像，2-AI生成头像',
    f_avatar          varchar(256) not null default '' comment '头像信息。当头像类型为0时，为内置头像标识；当头像类型为1和2时，为头像存储信息',

    f_created_at      bigint       not null default 0 comment '创建时间',
    f_created_by      varchar(40)  not null default '' comment '创建者',

    f_updated_at      bigint       not null default 0 comment '最后修改时间',
    f_updated_by      varchar(40)  not null default '' comment '最后修改者',

    f_deleted_by      varchar(36)  not null default '' comment '删除者id',
    f_deleted_at      bigint       not null default 0 comment '删除时间',

    f_config          text         not null comment '配置',

    f_status          varchar(30)  not null default 'unpublished' comment '发布状态，unpublished-未发布，published-已发布',

    f_created_type    varchar(20)  not null default 'create' comment '创建方式，create-创建，copy-复制',

#     f_published_at      bigint  not null default 0 comment '发布时间',
#     f_published_by        varchar(40)  not null default '' comment '发布者id',

    f_create_from     varchar(20)  not null default 'dip' comment '创建来源服务',
    primary key (f_id),
    unique key uk_key_deleted_at (f_key, f_deleted_at),
    index idx_created_at (f_created_at),
    index idx_deleted_at (f_deleted_at),
    index idx_is_built_in (f_is_built_in),
    index idx_created_by (f_created_by),
    index idx_status (f_status)
) engine = innodb comment ='data agent表';
-- data agent配置表 end

-- data agent配置模板表 start
create table if not exists t_data_agent_config_tpl
(
    f_id              bigint  not null auto_increment comment 'id',
    f_name            varchar(128) not null comment 'agent名',
    f_profile         varchar(256) not null default '' comment '简介',

    f_key             varchar(64)  not null comment '唯一标识，支持自动生成和人工设置',
    f_is_built_in     tinyint(4)   not null default 0 comment '是否是内置的，枚举值：0-否，1-是',

#     f_is_system_agent tinyint(4)   not null default 0 comment '是否是系统agent，枚举值：0-否，1-是',

    f_product_key     varchar(40)  not null default '' comment '产品标识',

    f_avatar_type     tinyint(4)   not null default 0 comment '头像类型，0-内置头像，1-用户上传头像，2-AI生成头像',
    f_avatar          varchar(256) not null default '' comment '头像信息。当头像类型为0时，为内置头像标识；当头像类型为1和2时，为头像存储信息',

    f_created_at      bigint       not null default 0 comment '创建时间',
    f_created_by      varchar(40)  not null default '' comment '创建者',

    f_updated_at      bigint       not null default 0 comment '最后修改时间',
    f_updated_by      varchar(40)  not null default '' comment '最后修改者',

    f_deleted_by      varchar(36)  not null default '' comment '删除者id',
    f_deleted_at      bigint       not null default 0 comment '删除时间',

    f_config          text         not null comment '配置',

    f_status          varchar(30)  not null default 'unpublished' comment '发布状态，unpublished-未发布，published-已发布',

    f_published_at    bigint       not null default 0 comment '发布时间',
    f_published_by    varchar(40)  not null default '' comment '发布人id',

#     f_is_last_one       tinyint(4)   not null default 1 comment '是否是最后一个（说明： 对于同一个模板（f_key相同） 1. 如果是未发布的，只有一个，这个就是最后一个；2. 如果是已发布的，并且发布后没有新的编辑，这个就是最后一个；3. 如果是已发布的，并且发布后有新的编辑，此时会有两条key一样的记录。已发布的这个不是最后一个，最新编辑的这个才是最后一个。 此字段用于在“我的模板列表”中，展示f_is_last_one=1的记录）枚举值：0-否，1-是',


#     f_category_id      varchar(40)  not null default '' comment '发布时选择的分类id，t_data_agent_release_category表的f_id',

    f_created_type    varchar(20)  not null comment '创建方式，copy_from_agent-从agent复制，copy_from_tpl-从模板复制，create-创建',

    f_create_from     varchar(20)  not null default 'dip' comment '创建来源服务',
    primary key (f_id),
    unique key uk_key_status_deleted_at (f_key, f_status, f_deleted_at),
    index idx_updated_at (f_updated_at),
    index idx_published_at (f_published_at),
    index idx_deleted_at (f_deleted_at),
    index idx_is_built_in (f_is_built_in),
    index idx_created_by (f_created_by),
#     index idx_is_last_one (f_is_last_one),
    index idx_status (f_status)
) engine = innodb comment ='data agent模板表';

create table if not exists t_data_agent_config_tpl_published
(
    f_id              bigint  not null auto_increment comment 'id',
    f_tpl_id      bigint not null comment 't_data_agent_tpl 表的f_id',
    f_name            varchar(128) not null comment 'agent模板名称',
    f_profile         varchar(256) not null default '' comment '简介',

    f_key             varchar(64)  not null comment '唯一标识，支持自动生成和人工设置',
    f_is_built_in     tinyint(4)   not null default 0 comment '是否是内置的，枚举值：0-否，1-是',

    f_product_key     varchar(40)  not null default '' comment '产品标识',

    f_avatar_type     tinyint(4)   not null default 0 comment '头像类型，0-内置头像，1-用户上传头像，2-AI生成头像',
    f_avatar          varchar(256) not null default '' comment '头像信息。当头像类型为0时，为内置头像标识；当头像类型为1和2时，为头像存储信息',


    f_config          text         not null comment '配置',

    f_published_at    bigint       not null default 0 comment '发布时间',
    f_published_by    varchar(40)  not null default '' comment '发布人id',



    primary key (f_id),
    unique key uk_key (f_key),
    index idx_published_at (f_published_at),
    index idx_is_built_in (f_is_built_in),
    unique key uk_tpl_id (f_tpl_id)
) engine = innodb comment ='data agent 已发布模板表';

create table if not exists t_data_agent_tpl_category_rel
(
    f_id          bigint      not null auto_increment,
    f_published_tpl_id      bigint not null comment 't_data_agent_config_tpl_published 表的f_id',
    f_category_id varchar(40) not null comment 't_data_agent_release_category 表的f_id',
    primary key (f_id),
    unique key uk_published_tpl_id_category_id (f_published_tpl_id, f_category_id),
    index idx_category_id (f_category_id)
) engine = innodb comment 'data agent模板和分类关联表';
-- data agent配置模板表 end

-- 产品相关表 start
create table if not exists t_product
(
    f_id         BIGINT       not null auto_increment comment '唯一id',
    f_name       varchar(64)  not null comment '产品显示名称',
    f_key        varchar(40)  not null comment '产品唯一标识符',
    f_profile    varchar(256) not null default '' comment '产品简介说明',

    f_created_by varchar(36)  not null default '' comment '创建者id',
    f_created_at bigint       not null default 0 comment '创建时间',

    f_updated_by varchar(36)  not null default '' comment '修改者id',
    f_updated_at bigint       not null default 0 comment '更新时间',

    f_deleted_by varchar(36)  not null default '' comment '删除者id',
    f_deleted_at bigint       not null default 0 comment '删除时间',

    unique key uk_key_deleted_at (f_key, f_deleted_at),
    key idx_deleted_at (f_deleted_at),
    key idx_created_at (f_created_at),
    primary key (f_id)
) engine = innodb comment '产品表';

-- 内置产品
insert into t_product (f_name, f_profile, f_key, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by,
                       f_deleted_at)
select 'AnyShare',
       'AnyShare',
       'anyshare',
       '',
       unix_timestamp() * 1000,
       '',
       0,
       '',
       0
from dual
where not exists (select 1 from t_product where f_key = 'anyshare');

insert into t_product (f_name, f_profile, f_key, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by,
                       f_deleted_at)
select 'DIP',
       'DIP',
       'dip',
       '',
       unix_timestamp() * 1000,
       '',
       0,
       '',
       0
from dual
where not exists (select 1 from t_product where f_key = 'dip');

insert into t_product (f_name, f_profile, f_key, f_created_by, f_created_at, f_updated_by, f_updated_at, f_deleted_by,
                       f_deleted_at)
select 'ChatBI',
       'ChatBI',
       'chatbi',
       '',
       unix_timestamp() * 1000,
       '',
       0,
       '',
       0
from dual
where not exists (select 1 from t_product where f_key = 'chatbi');

-- 产品相关表 end

-- 空间表 start

-- 1. 空间主表
create table if not exists t_custom_space
(
    f_id         varchar(40)  not null comment '唯一id',
    f_name       varchar(64)  not null comment '空间显示名称',
    f_key        varchar(64)  not null comment '空间唯一标识符，支持自动生成和人工设置',
    f_profile    varchar(256) not null default '' comment '空间简介说明',

    f_created_by varchar(36)  not null default '' comment '创建者id',
    f_created_at bigint       not null default 0 comment '创建时间',

    f_updated_by varchar(36)  not null default '' comment '修改者id',
    f_updated_at bigint       not null default 0 comment '更新时间',

    f_deleted_by varchar(36)  not null default '' comment '删除者id',
    f_deleted_at bigint       not null default 0 comment '删除时间',

    unique key uk_key_deleted_at (f_key, f_deleted_at),
    key idx_created_by (f_created_by),
    key idx_created_at (f_created_at),
    key idx_updated_at (f_updated_at),
    primary key (f_id)
) engine = innodb comment '空间主表';

-- 2. 空间-成员关联表
create table if not exists t_custom_space_member
(
    f_id         bigint      not null auto_increment,
    f_space_id   varchar(40) not null comment '空间id，t_space表的f_id',
    f_space_key  varchar(64) not null comment '空间唯一标识',
    f_obj_type   varchar(32) not null comment '组织对象类型，枚举值：user-个人，dept-部门，user_group-用户组',
    f_obj_id     varchar(64) not null comment '组织对象的唯一标识',

    f_created_at bigint      not null default 0 comment '创建时间',
    f_created_by varchar(36) not null default '' comment '创建者id',

    unique key uk_space_id_obj_type_obj_id (f_space_id, f_obj_type, f_obj_id),
    primary key (f_id)
) engine = innodb comment '空间-成员关联表';

-- 3. 空间-资源对象关联表
create table if not exists t_custom_space_resource
(
    f_id            bigint      not null auto_increment,
    f_space_id      varchar(40) not null comment '空间id，t_space表的f_id',
    f_space_key     varchar(64) not null comment '空间唯一标识',
    f_resource_type varchar(32) not null comment '资源对象类型，如：data_agent等',
    f_resource_id   varchar(64) not null comment '资源的唯一标识',

    f_created_at    bigint      not null default 0 comment '创建时间',
    f_created_by    varchar(36) not null default '' comment '创建者id',

    unique key uk_space_id_resource_type_resource_id (f_space_id, f_resource_type, f_resource_id),
    primary key (f_id)
) engine = innodb comment '空间-资源对象关联表';

-- 空间表 end

-- 发布表
create table if not exists t_data_agent_release
(
    f_id                 varchar(40)  not null comment 'id',
#     f_seq_no             bigint       not null auto_increment comment '自增序号' unique,
    f_agent_id           varchar(40)  not null comment 'agent_id',
    f_agent_name         varchar(128) not null comment '发布时agent的名称',

    f_agent_config       longtext     not null comment 'agent配置',

    f_agent_version      varchar(32)  not null comment 'agent版本',
    f_agent_desc         varchar(255) not null default '' comment 'agent描述',

    f_is_api_agent       tinyint(4)   not null default 0 comment '是否发布为api agent，枚举值：0-否，1-是',
    f_is_web_sdk_agent   tinyint(4)   not null default 0 comment '是否发布为web sdk agent，枚举值：0-否，1-是',
    f_is_skill_agent     tinyint(4)   not null default 0 comment '是否发布为技能 agent，枚举值：0-否，1-是',
    f_is_data_flow_agent tinyint(4)   not null default 0 comment '是否发布为数据流 agent，枚举值：0-否，1-是',

    f_is_to_custom_space tinyint(4)   not null default 0 comment '是否发布到自定义空间，枚举值：0-否，1-是',
    f_is_to_square       tinyint(4)   not null default 0 comment '是否发布到广场，枚举值：0-否，1-是',

    f_is_pms_ctrl        tinyint(4)   not null default 0 comment '是否进行“使用权限”的管控，枚举值：0-否，1-是。开启管控是，会在t_data_agent_release_permission表中记录哪些对象有权限使用此data agent',

    f_create_time        bigint       not null default 0 comment '创建时间',
    f_update_time        bigint       not null default 0 comment '最后修改时间',
    f_create_by          varchar(40)  not null default '' comment '创建者',
    f_update_by          varchar(40)  not null default '' comment '最后修改者',


    primary key (f_id),
    unique key uk_agent_id (f_agent_id),
    index idx_update_time (f_update_time),

    index idx_is_api_agent (f_is_api_agent),
    index idx_is_web_sdk_agent (f_is_web_sdk_agent),
    index idx_is_skill_agent (f_is_skill_agent),
    index idx_is_data_flow_agent (f_is_data_flow_agent),

    index idx_is_to_custom_space (f_is_to_custom_space),
    index idx_is_to_square (f_is_to_square),
    
    index idx_is_pms_ctrl (f_is_pms_ctrl)
) engine = innodb comment ='data agent发布表';


-- 历史发布表
create table if not exists t_data_agent_release_history
(
    f_id            varchar(40)  not null comment 'id',
    f_agent_id      varchar(40)  not null comment 'agent_id',

    f_agent_config  longtext     not null comment 'agent配置',

    f_agent_version varchar(32)  not null comment 'agent版本',
    f_agent_desc    varchar(255) not null default '' comment 'agent描述',

    f_create_time   bigint       not null default 0 comment '创建时间',
    f_update_time   bigint       not null default 0 comment '最后修改时间',
    f_create_by     varchar(40)  not null default '' comment '创建者',
    f_update_by     varchar(40)  not null default '' comment '最后修改者',

    primary key (f_id),
    unique key uk_agent_id_agent_version (f_agent_id, f_agent_version)
) engine = innodb comment ='data agent发布历史表';

-- 分类表
create table if not exists t_data_agent_release_category
(
    f_id          varchar(40)  not null comment 'id',
    f_name        varchar(128) not null comment '分类名',
    f_description varchar(256) not null default '' comment '描述',

    f_create_time bigint       not null default 0 comment '创建时间',
    f_update_time bigint       not null default 0 comment '最后修改时间',
    f_create_by   varchar(40)  not null default '' comment '创建者',
    f_update_by   varchar(40)  not null default '' comment '最后修改者',
    primary key (f_id)
) engine = innodb comment 'data agent发布分类表';


-- release和分类关联表
create table if not exists t_data_agent_release_category_rel
(
    f_id          varchar(40) not null comment 'id',
    f_release_id  varchar(40) not null comment 't_data_agent_release 表的f_id',
    f_category_id varchar(40) not null comment 't_data_agent_release_category 表的f_id',
    primary key (f_id),
    unique key uk_release_id_category_id (f_release_id, f_category_id)
) engine = innodb comment 'data agent 发布分类关联表';


-- 适用范围
create table if not exists t_data_agent_release_permission
(
    f_id         bigint      not null auto_increment,
    f_release_id varchar(40) not null comment 't_data_agent_release 表的f_id',
    f_obj_type   varchar(32) not null comment '适用对象类型，枚举值：user-个人，dept-组织，user_group-用户组，role-角色，app_account-应用账号',
    f_obj_id     varchar(64) not null comment '适用对象ID',
    unique key uk_release_id_obj_type_obj_id (f_release_id, f_obj_type, f_obj_id),
    primary key (f_id)
) engine = innodb comment 'data agent配置-适用范围（哪些对象有权限使用此data agent）';


-- 最近访问

create table if not exists t_data_agent_visit_history
(
    f_id            varchar(40) not null comment 'id',
    f_agent_id      varchar(40) not null comment 'agent_id',
    f_agent_version varchar(32) not null comment 'agent版本',

    f_custom_space_id varchar(40) not null default '' comment '自定义空间id',

    f_visit_count   int         not null default 1 comment '访问次数',

    f_create_time   bigint      not null default 0 comment '创建时间（首次访问时间）',
    f_update_time   bigint      not null default 0 comment '最后修改时间（最近访问时间）',
    f_create_by     varchar(40) not null default '' comment '创建者',
    f_update_by     varchar(40) not null default '' comment '最后修改者',
    primary key (f_id),
    unique key uk_user_agent (f_create_by, f_agent_id, f_agent_version),
    index idx_custom_space_id (f_custom_space_id)
) engine = innodb comment '用户最近访问Agent历史记录表';

-- 业务域和agent、agent模板的关联 start--

-- 业务域与agent关联表
create table if not exists t_biz_domain_agent_rel
(
    f_id            bigint      not null auto_increment comment '自增ID',
    f_biz_domain_id varchar(40) not null comment '业务域ID',
    f_agent_id      varchar(40) not null comment 'agent ID，对应t_data_agent_config表的f_id',
    f_created_at    bigint      not null default 0 comment '创建时间',
    primary key (f_id),
    unique key uk_biz_domain_id_agent_id (f_biz_domain_id, f_agent_id),
    index idx_agent_id (f_agent_id)
) engine = innodb comment '业务域与agent关联表';

-- 业务域与agent模板关联表
create table if not exists t_biz_domain_agent_tpl_rel
(
    f_id            bigint      not null auto_increment comment '自增ID',
    f_biz_domain_id varchar(40) not null comment '业务域ID',
    f_agent_tpl_id  bigint      not null comment 'agent模板ID，对应t_data_agent_config_tpl表的f_id',
    f_created_at    bigint      not null default 0 comment '创建时间',
    primary key (f_id),
    unique key uk_biz_domain_id_agent_tpl_id (f_biz_domain_id, f_agent_tpl_id),
    index idx_agent_tpl_id (f_agent_tpl_id)
) engine = innodb comment '业务域与agent模板关联表';

-- 业务域和agent、agent模板的关联 end--


-- 初始化分类
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


-- data agent 会话表
create table if not exists t_data_agent_conversation
(
    f_id                 varchar(40)  not null comment '会话 ID，会话唯一标识',
    f_agent_app_key      varchar(40)  not null comment 'agent app key',

    f_title              varchar(255) not null comment '会话标题，默认使用首次用户提问消息的前20个字符，支持修改标题',
    f_origin             varchar(40) not null default 'web_chat' comment '用于标记会话发起源：1. web_chat: 通过浏览器对话发起 2. api_call: api 调用发起（当前 API 暂不记录会话，只是预留未来扩展）',
    f_message_index      int not null default 0 comment '最新消息下标，会话消息下标从0开始，每产生一条新消息，下标 +1',
    f_read_message_index int not null default 0 comment '最新已读消息下标，用于实现未读消息提醒功能，当前已读会话消息下标 < 最新会话消息下标时，表示有未读的消息',
    f_ext                mediumtext not null  comment '预留扩展字段',

    f_create_time        bigint  not null default 0 comment '创建时间',
    f_update_time        bigint  not null default 0 comment '最后修改时间',
    f_create_by          varchar(40)  not null default '' comment '创建者',
    f_update_by          varchar(40)  not null default '' comment '最后修改者',
    f_is_deleted         tinyint(4) not null default 0 comment '是否删除：0-否 1-是',

    primary key (f_id),
    index idx_agent_app_key (f_agent_app_key)
) engine = innodb comment ='data agent 会话表';

-- data agent 会话消息表
create table if not exists t_data_agent_conversation_message (
    f_id              varchar(40) not null comment '消息ID，消息唯一标识',
    f_agent_app_key   varchar(40)  not null comment 'agent app key',
    f_conversation_id varchar(40) not null default '' comment '会话ID，会话唯一标识',
    f_agent_id        varchar(40)  not null comment 'agent ID',
    f_agent_version   varchar(32)  not null comment 'agent版本',
    f_reply_id        varchar(40) not null default '' comment '回复消息ID，用于关联问答消息',

    f_index           int not null comment '消息下标，用于标记消息在整个会话中的位置、顺序，比如基于Index正序在前端按照时间线展示对话消息',
    f_role            varchar(255) not null comment '产生消息的角色，支持一下角色：User: 用户；Assistant: 助手',
    f_content         mediumtext  not null comment '消息内容，结构随角色类型变化。当角色为User时，用户输入包括文字和临时区文件（图片、文档、音视频）；当角色为Assistant时，包括最终返回结果和中间结果',
    f_content_type    varchar(32) comment '内容类型',
    f_status          varchar(32) comment '消息状态，随Role类型变化，Role为 User时：Received :  已接收(消息成功接收并持久化， 初始状态)Processed: 处理完成（成功触发后续的Agent Call）；Role为Assistant时：Processing： 生成中（消息正在生成中 ， 初始状态）Succeded： 生成成功（消息处理完成，返回成功）Failed： 生成失败（消息生成失败）Cancelled: 取消生成（用户、系统终止会话）',
    f_ext             mediumtext  not null  comment '预留扩展字段',

    f_create_time        bigint  not null default 0 comment '创建时间',
    f_update_time        bigint  not null default 0 comment '最后修改时间',
    f_create_by          varchar(40)  not null default '' comment '创建者',
    f_update_by          varchar(40)  not null default '' comment '最后修改者',
    f_is_deleted         tinyint(4) not null default 0 comment '是否删除：0-否 1-是',

    primary key (f_id),
    index idx_agent_app_key (f_agent_app_key),
    index idx_conversation_id (f_conversation_id)
) engine = innodb comment = '会话消息表';