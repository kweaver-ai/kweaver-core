
-- UserManagement

/*
MySQL: Database - user_management
*********************************************************************
*/
use user_management;

CREATE TABLE IF NOT EXISTS `t_group` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_group_id` char(40) NOT NULL COMMENT '用户组唯一标识',
    `f_group_name` varchar(512) NOT NULL COMMENT '用户组名',
    `f_created_time` bigint(40) NOT NULL COMMENT '用户组创建时间',
    `f_notes` varchar(1200) NOT NULL COMMENT '备注',
    PRIMARY KEY (`f_primary_id`)
) ENGINE=InnoDB COMMENT='AnyShare用户组信息表';


CREATE TABLE IF NOT EXISTS `t_group_member` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_group_id` char(40) NOT NULL COMMENT '用户组唯一标识',
    `f_member_id` char(40) NOT NULL COMMENT '用户唯一标识,可以是用户，可以是部门',
    `f_member_type` tinyint(4) NOT NULL COMMENT '用户类型,0：部门：1：用户',
    `f_added_time` bigint(40) NOT NULL COMMENT '用户组成员添加时间',
    PRIMARY KEY (`f_primary_id`)
) ENGINE=InnoDB COMMENT='AnyShare用户组成员信息表';

CREATE TABLE IF NOT EXISTS `t_anonymity` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_anonymity_id` char(40) NOT NULL COMMENT '匿名账户id',
    `f_password` varchar(100) NOT NULL COMMENT '访问密码',
    `f_expires_at` bigint(40) NOT NULL COMMENT '到期时间, 0为永久有效',
    `f_limited_times` bigint(40) NOT NULL COMMENT '访问限制次数, -1为无限制',
    `f_accessed_times` bigint(40) NOT NULL DEFAULT '0' COMMENT '已访问次数',
    `f_created_at` bigint(40) NOT NULL COMMENT '生成时间',
    `f_type`  char(40) NOT NULL COMMENT '匿名账户类型 example:document 文档匿名用户',
    `f_verify_mobile` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否手机验证 , 0: 否，1: 是',
    PRIMARY KEY (`f_primary_id`),
    UNIQUE KEY `idx_id` (`f_anonymity_id`),
    KEY `idx_expires_at` (`f_expires_at`)
) ENGINE=InnoDB COMMENT='匿名账户表';

CREATE TABLE IF NOT EXISTS `t_outbox` (
    `f_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `f_business_type` tinyint(4) NOT NULL COMMENT '业务类型',
    `f_message` longtext NOT NULL COMMENT '消息内容，json格式字符串',
    `f_create_time` bigint(20) NOT NULL COMMENT '消息创建时间',
    PRIMARY KEY (`f_id`),
    KEY `idx_business_type_and_create_time` (`f_business_type`, `f_create_time`)
) ENGINE=InnoDB COMMENT='outbox信息表';

CREATE TABLE IF NOT EXISTS `t_outbox_lock` (
    `f_business_type` tinyint(4) NOT NULL COMMENT '业务类型',
    PRIMARY KEY (`f_business_type`)
) ENGINE=InnoDB COMMENT='outbox分布式锁表';

INSERT INTO t_outbox_lock(f_business_type) SELECT 1 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 1);
INSERT INTO t_outbox_lock(f_business_type) SELECT 2 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 2);
INSERT INTO t_outbox_lock(f_business_type) SELECT 3 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 3);
INSERT INTO t_outbox_lock(f_business_type) SELECT 4 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 4);
INSERT INTO t_outbox_lock(f_business_type) SELECT 5 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 5);
INSERT INTO t_outbox_lock(f_business_type) SELECT 6 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 6);
INSERT INTO t_outbox_lock(f_business_type) SELECT 7 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 7);
INSERT INTO t_outbox_lock(f_business_type) SELECT 8 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 8);
INSERT INTO t_outbox_lock(f_business_type) SELECT 9 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 9);

CREATE TABLE IF NOT EXISTS `t_app` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_id` char(40) NOT NULL COMMENT '应用账户ID',
    `f_name` varchar(512) NOT NULL COMMENT '应用账户名称',
    `f_password` varchar(100) NOT NULL COMMENT '应用账户密码',           -- 使用BCrypt散列后固定占60位，暂定为varcher(100)
    `f_type` tinyint(4) NOT NULL COMMENT '应用账户类型',
    `f_created_time` bigint(40) NOT NULL COMMENT '应用账户创建时间',
    `f_credential_type` tinyint(4) NOT NULL DEFAULT '1' COMMENT '凭证类型,1: 密码,2: 令牌',
    PRIMARY KEY (`f_primary_id`),
    UNIQUE KEY `f_id` (`f_id`),
    UNIQUE KEY `f_name` (`f_name`)
) ENGINE=InnoDB COMMENT='应用账户信息表';

CREATE TABLE IF NOT EXISTS `t_org_perm_app` (                                                   -- 此表记录应用账户对组织架构管理的权限
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,                                          -- 自增主键
    `f_app_id` char(40) NOT NULL,                                                               -- 应用账户id
    `f_app_name` varchar(150) NOT NULL,                                                         -- 应用账户名称
    `f_org_type` tinyint(4) NOT NULL,                                                           -- 组织架构对象类型，1：用户，2：部门，3：用户组
    `f_perm_value` int(11) NOT NULL DEFAULT '0',                                                -- 权限值
    `f_end_time` bigint(20) DEFAULT '-1',                                                       -- 权限结束时间, 微秒的时间戳, -1标识永久有效
    `f_modify_time` bigint(20) NOT NULL DEFAULT '0',                                            -- 记录修改时间, 微秒的时间戳
    `f_create_time` bigint(20) NOT NULL,                                                        -- 记录创建时间, 微秒的时间戳
    PRIMARY KEY (`f_primary_id`),
    KEY `idx_f_app_id` (`f_app_id`),
    KEY `idx_f_end_time` (`f_end_time`),
    KEY `idx_f_org_type` (`f_org_type`)
) ENGINE=InnoDB COMMENT='应用账户组织架构管理权限表';

CREATE TABLE IF NOT EXISTS `t_avatar` (                                                   -- 此表记录用户头像信息
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,                                    -- 自增主键
    `f_user_id` char(40) NOT NULL,                                                        -- 用户ID
    `f_oss_id` char(40) NOT NULL,                                                         -- 对象存储ID
    `f_key` char(80) NOT NULL,                                                            -- 对象存储内文件KEY值
    `f_type` varchar(50) NOT NULL,                                                        -- 文件类型
    `f_status` tinyint(4) NOT NULL,                                                       -- 文件状态类型，0：未使用，1：已使用
    `f_time` bigint(20) NOT NULL,                                                         -- 记录创建时间, 微秒的时间戳
    PRIMARY KEY (`f_primary_id`),
    KEY `idx_f_user_id` (`f_user_id`),
    UNIQUE KEY `idx_f_key` (`f_key`),
    KEY `idx_f_time` (`f_time`)
) ENGINE=InnoDB COMMENT='用户头像信息表';


CREATE TABLE IF NOT EXISTS `t_internal_group` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_id` char(40) NOT NULL COMMENT '内部组唯一标识',
    `f_created_time` bigint(40) NOT NULL COMMENT '内部组创建时间',
    PRIMARY KEY (`f_primary_id`),
    KEY `idx_id` (`f_id`)
) ENGINE=InnoDB COMMENT='AnyShare内部组信息表';

CREATE TABLE IF NOT EXISTS `t_internal_group_member` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_internal_group_id` char(40) NOT NULL COMMENT '内部组唯一标识',
    `f_member_id` char(40) NOT NULL COMMENT '成员唯一标识',
    `f_member_type` tinyint(4) NOT NULL COMMENT '成员类型,1：用户',
    `f_added_time` bigint(40) NOT NULL COMMENT '内部组成员添加时间',
    PRIMARY KEY (`f_primary_id`),
    KEY `idx_id` (`f_internal_group_id`),
    KEY `idx_f_member_id` (`f_member_id`)
) ENGINE=InnoDB COMMENT='AnyShare内部组成员信息表';

CREATE TABLE IF NOT EXISTS `option` (
    `key` varchar(40) NOT NULL COMMENT '配置关键字',
    `value` varchar(150) NOT NULL COMMENT '配置值',
    PRIMARY KEY (`key`)
) ENGINE=InnoDB COMMENT='配置表';

CREATE TABLE IF NOT EXISTS `t_org_perm` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_id` char(40) NOT NULL COMMENT '账户id',
    `f_name` varchar(150) NOT NULL COMMENT '账户名称',
    `f_type` tinyint(4) NOT NULL COMMENT '账户类型，1：实名用户',
    `f_org_type` tinyint(4) NOT NULL COMMENT '组织架构对象类型，1：用户，2：部门，3：用户组',
    `f_perm_value` int(11) NOT NULL DEFAULT '0' COMMENT '权限值',
    `f_end_time` bigint(20) DEFAULT '-1' COMMENT '权限结束时间, 微秒的时间戳, -1标识永久有效',
    `f_modify_time` bigint(20) NOT NULL DEFAULT '0' COMMENT '记录修改时间, 微秒的时间戳',
    `f_create_time` bigint(20) NOT NULL COMMENT '记录创建时间, 微秒的时间戳',
    PRIMARY KEY (`f_primary_id`),
    KEY `idx_f_id` (`f_id`),
    KEY `idx_f_end_time` (`f_end_time`),
    KEY `idx_f_org_type` (`f_org_type`),
    KEY `idx_f_type` (`f_type`)
) ENGINE=InnoDB COMMENT='组织架构管理权限表';

INSERT INTO `option`(`key`,`value`) SELECT 'user_defalut_des_password','4SLXQjA5JbE=' FROM DUAL WHERE NOT EXISTS(SELECT `value` FROM `option` WHERE `key` = 'user_defalut_des_password');
INSERT INTO `option`(`key`,`value`) SELECT 'user_defalut_ntlm_password','32ed87bdb5fdc5e9cba88547376818d4' FROM DUAL WHERE NOT EXISTS(SELECT `value` FROM `option` WHERE `key` = 'user_defalut_ntlm_password');
INSERT INTO `option`(`key`,`value`) SELECT 'user_defalut_sha2_password','8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92' FROM DUAL WHERE NOT EXISTS(SELECT `value` FROM `option` WHERE `key` = 'user_defalut_sha2_password');
INSERT INTO `option`(`key`,`value`) SELECT 'user_defalut_md5_password','e10adc3949ba59abbe56e057f20f883e' FROM DUAL WHERE NOT EXISTS(SELECT `value` FROM `option` WHERE `key` = 'user_defalut_md5_password');

use sharemgnt_db;

CREATE TABLE IF NOT EXISTS `t_reserved_name` (
  `f_id` char(40) NOT NULL COMMENT 'id',
  `f_name` char(150) NOT NULL COMMENT '名称',
  `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
  `f_update_time` bigint(20) NOT NULL COMMENT '修改时间',
  PRIMARY KEY (`f_id`),
  KEY `idx_name` (`f_name`)
) ENGINE=InnoDB COMMENT='保留名称表';

CREATE TABLE IF NOT EXISTS `t_sharemgnt_config` (
  `f_key` char(32) NOT NULL,                                        -- 配置关键字
  `f_value` varchar(1024) NOT NULL,                                 -- 配置的值
  PRIMARY KEY (`f_key`)
) ENGINE=InnoDB;

INSERT INTO `t_sharemgnt_config`(`f_key`, `f_value`) SELECT 'user_expired_disable_lock', 'locked' FROM DUAL WHERE NOT EXISTS (SELECT `f_key` FROM `t_sharemgnt_config` WHERE `f_key` = 'user_expired_disable_lock');
INSERT INTO `t_sharemgnt_config`(`f_key`, `f_value`) SELECT 'user_not_login_disable_lock', 'locked' FROM DUAL WHERE NOT EXISTS (SELECT `f_key` FROM `t_sharemgnt_config` WHERE `f_key` = 'user_not_login_disable_lock');


-- hydra

-- Migration generated by the command below; DO NOT EDIT.
-- hydra:generate hydra migrate gen
SET FOREIGN_KEY_CHECKS = 0;

use hydra_v2;

CREATE TABLE IF NOT EXISTS `networks` (
  `id` char(36) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_client` (
  `id` varchar(255) NOT NULL,
  `client_name` text NOT NULL,
  `client_secret` text NOT NULL,
  `scope` text NOT NULL,
  `owner` text NOT NULL,
  `policy_uri` text NOT NULL,
  `tos_uri` text NOT NULL,
  `client_uri` text NOT NULL,
  `logo_uri` text NOT NULL,
  `client_secret_expires_at` int NOT NULL DEFAULT '0',
  `sector_identifier_uri` text NOT NULL,
  `jwks` text NOT NULL,
  `jwks_uri` text NOT NULL,
  `token_endpoint_auth_method` varchar(25) NOT NULL DEFAULT '',
  `request_object_signing_alg` varchar(10) NOT NULL DEFAULT '',
  `userinfo_signed_response_alg` varchar(10) NOT NULL DEFAULT '',
  `subject_type` varchar(15) NOT NULL DEFAULT '',
  `pk_deprecated` int unsigned DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `frontchannel_logout_uri` text NOT NULL,
  `frontchannel_logout_session_required` tinyint(1) NOT NULL DEFAULT '0',
  `backchannel_logout_uri` text NOT NULL,
  `backchannel_logout_session_required` tinyint(1) NOT NULL DEFAULT '0',
  `metadata` text NOT NULL,
  `token_endpoint_auth_signing_alg` varchar(10) NOT NULL DEFAULT '',
  `authorization_code_grant_access_token_lifespan` bigint DEFAULT NULL,
  `authorization_code_grant_id_token_lifespan` bigint DEFAULT NULL,
  `authorization_code_grant_refresh_token_lifespan` bigint DEFAULT NULL,
  `client_credentials_grant_access_token_lifespan` bigint DEFAULT NULL,
  `implicit_grant_access_token_lifespan` bigint DEFAULT NULL,
  `implicit_grant_id_token_lifespan` bigint DEFAULT NULL,
  `jwt_bearer_grant_access_token_lifespan` bigint DEFAULT NULL,
  `password_grant_access_token_lifespan` bigint DEFAULT NULL,
  `password_grant_refresh_token_lifespan` bigint DEFAULT NULL,
  `refresh_token_grant_id_token_lifespan` bigint DEFAULT NULL,
  `refresh_token_grant_access_token_lifespan` bigint DEFAULT NULL,
  `refresh_token_grant_refresh_token_lifespan` bigint DEFAULT NULL,
  `pk` char(36) NOT NULL,
  `registration_access_token_signature` varchar(128) NOT NULL DEFAULT '',
  `nid` char(36) NOT NULL,
  `redirect_uris` json NOT NULL,
  `grant_types` json NOT NULL,
  `response_types` json NOT NULL,
  `audience` json NOT NULL,
  `allowed_cors_origins` json NOT NULL,
  `contacts` json NOT NULL,
  `request_uris` json NOT NULL,
  `post_logout_redirect_uris` json NOT NULL,
  `access_token_strategy` varchar(10) NOT NULL DEFAULT '',
  `skip_consent` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`pk`),
  UNIQUE KEY `hydra_client_id_key` (`id`,`nid`),
  KEY `pk_deprecated` (`pk_deprecated`),
  KEY `hydra_client_nid_fk_idx` (`nid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_jwk` (
  `sid` varchar(255) NOT NULL,
  `kid` varchar(255) NOT NULL,
  `version` int NOT NULL DEFAULT '0',
  `keydata` text NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `pk_deprecated` int unsigned NOT NULL AUTO_INCREMENT,
  `pk` char(36) NOT NULL,
  `nid` char(36) NOT NULL,
  PRIMARY KEY (`pk`),
  UNIQUE KEY `hydra_jwk_sid_kid_nid_key` (`sid`,`kid`,`nid`),
  KEY `pk_deprecated` (`pk_deprecated`),
  KEY `hydra_jwk_nid_fk_idx` (`nid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_oauth2_authentication_session` (
  `id` varchar(40) NOT NULL,
  `authenticated_at` timestamp NULL DEFAULT NULL,
  `subject` varchar(255) NOT NULL,
  `remember` tinyint(1) NOT NULL DEFAULT '0',
  `nid` char(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `hydra_oauth2_authentication_session_nid_fk_idx` (`nid`),
  KEY `hydra_oauth2_authentication_session_subject_nid_idx` (`subject`,`nid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_oauth2_flow` (
  `login_challenge` varchar(40) NOT NULL,
  `login_verifier` varchar(40) NOT NULL,
  `login_csrf` varchar(40) NOT NULL,
  `subject` varchar(255) NOT NULL,
  `request_url` text NOT NULL,
  `login_skip` tinyint(1) NOT NULL,
  `client_id` varchar(255) NOT NULL,
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `login_initialized_at` timestamp NULL DEFAULT NULL,
  `oidc_context` json NOT NULL,
  `login_session_id` varchar(40),
  `state` smallint NOT NULL,
  `login_remember` tinyint(1) NOT NULL DEFAULT '0',
  `login_remember_for` int NOT NULL,
  `login_error` text,
  `acr` text NOT NULL,
  `login_authenticated_at` timestamp NULL DEFAULT NULL,
  `login_was_used` tinyint(1) NOT NULL DEFAULT '0',
  `forced_subject_identifier` varchar(255) NOT NULL DEFAULT '',
  `context` json NOT NULL,
  `consent_challenge_id` varchar(40) DEFAULT NULL,
  `consent_skip` tinyint(1) NOT NULL DEFAULT '0',
  `consent_verifier` varchar(40) DEFAULT NULL,
  `consent_csrf` varchar(40) DEFAULT NULL,
  `consent_remember` tinyint(1) NOT NULL DEFAULT '0',
  `consent_remember_for` int DEFAULT NULL,
  `consent_handled_at` timestamp NULL DEFAULT NULL,
  `consent_error` text,
  `session_access_token` json NOT NULL,
  `session_id_token` json NOT NULL,
  `consent_was_used` tinyint(1) DEFAULT NULL,
  `nid` char(36) NOT NULL,
  `requested_scope` json NOT NULL,
  `requested_at_audience` json DEFAULT NULL,
  `amr` json DEFAULT NULL,
  `granted_scope` json DEFAULT NULL,
  `granted_at_audience` json DEFAULT NULL,
  `login_extend_session_lifespan` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`login_challenge`),
  UNIQUE KEY `hydra_oauth2_flow_login_verifier_idx` (`login_verifier`),
  UNIQUE KEY `hydra_oauth2_flow_consent_challenge_idx` (`consent_challenge_id`),
  UNIQUE KEY `hydra_oauth2_flow_consent_verifier_idx` (`consent_verifier`),
  KEY `hydra_oauth2_flow_login_session_id_idx` (`login_session_id`),
  KEY `hydra_oauth2_flow_nid_fk_idx` (`nid`),
  KEY `hydra_oauth2_flow_client_id_subject_idx` (`client_id`,`nid`,`subject`),
  KEY `hydra_oauth2_flow_sub_idx` (`subject`,`nid`),
  KEY `hydra_oauth2_flow_multi_query_idx` (`consent_error`(2),`state`,`subject`,`client_id`,`consent_skip`,`consent_remember`,`nid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_oauth2_access` (
  `signature` varchar(255) NOT NULL,
  `request_id` varchar(40) NOT NULL DEFAULT '',
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_id` varchar(255) NOT NULL DEFAULT '',
  `scope` text NOT NULL,
  `granted_scope` text NOT NULL,
  `form_data` text NOT NULL,
  `session_data` text NOT NULL,
  `subject` varchar(255) NOT NULL DEFAULT '',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `requested_audience` text NOT NULL,
  `granted_audience` text NOT NULL,
  `challenge_id` varchar(40) DEFAULT NULL,
  `nid` char(36) NOT NULL,
  CONSTRAINT `hydra_oauth2_access_challenge_id_fk` FOREIGN KEY (`challenge_id`) REFERENCES `hydra_oauth2_flow` (`consent_challenge_id`) ON DELETE CASCADE,
  PRIMARY KEY (`signature`),
  KEY `hydra_oauth2_access_challenge_id_idx` (`challenge_id`),
  KEY `hydra_oauth2_access_nid_fk_idx` (`nid`),
  KEY `hydra_oauth2_access_client_id_fk` (`client_id`,`nid`),
  KEY `hydra_oauth2_access_requested_at_idx` (`requested_at`,`nid`),
  KEY `hydra_oauth2_access_client_id_subject_nid_idx` (`client_id`,`subject`,`nid`),
  KEY `hydra_oauth2_access_request_id_idx` (`request_id`,`nid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_oauth2_code` (
  `signature` varchar(255) NOT NULL,
  `request_id` varchar(40) NOT NULL DEFAULT '',
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_id` varchar(255) NOT NULL DEFAULT '',
  `scope` text NOT NULL,
  `granted_scope` text NOT NULL,
  `form_data` text NOT NULL,
  `session_data` text NOT NULL,
  `subject` varchar(255) NOT NULL DEFAULT '',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `requested_audience` text NOT NULL,
  `granted_audience` text NOT NULL,
  `challenge_id` varchar(40) DEFAULT NULL,
  `nid` char(36) NOT NULL,
  CONSTRAINT `hydra_oauth2_code_challenge_id_fk` FOREIGN KEY (`challenge_id`) REFERENCES `hydra_oauth2_flow` (`consent_challenge_id`) ON DELETE CASCADE,
  PRIMARY KEY (`signature`),
  KEY `hydra_oauth2_code_challenge_id_idx` (`challenge_id`),
  KEY `hydra_oauth2_code_nid_fk_idx` (`nid`),
  KEY `hydra_oauth2_code_client_id_fk` (`client_id`,`nid`),
  KEY `hydra_oauth2_code_request_id_idx` (`request_id`,`nid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_oauth2_jti_blacklist` (
  `signature` varchar(64) NOT NULL,
  `expires_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `nid` char(36) NOT NULL,
  PRIMARY KEY (`signature`,`nid`),
  KEY `hydra_oauth2_jti_blacklist_nid_fk_idx` (`nid`),
  KEY `hydra_oauth2_jti_blacklist_expiry` (`expires_at`,`nid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_oauth2_logout_request` (
  `challenge` varchar(36) NOT NULL,
  `verifier` varchar(36) NOT NULL,
  `subject` varchar(255) NOT NULL,
  `sid` varchar(36) NOT NULL,
  `client_id` varchar(255) DEFAULT NULL,
  `request_url` text NOT NULL,
  `redir_url` text NOT NULL,
  `was_used` tinyint(1) NOT NULL DEFAULT '0',
  `accepted` tinyint(1) NOT NULL DEFAULT '0',
  `rejected` tinyint(1) NOT NULL DEFAULT '0',
  `rp_initiated` tinyint(1) NOT NULL DEFAULT '0',
  `nid` char(36) NOT NULL,
  PRIMARY KEY (`challenge`),
  UNIQUE KEY `hydra_oauth2_logout_request_veri_idx` (`verifier`),
  KEY `hydra_oauth2_logout_request_nid_fk_idx` (`nid`),
  KEY `hydra_oauth2_logout_request_client_id_fk` (`client_id`,`nid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_oauth2_obfuscated_authentication_session` (
  `subject` varchar(255) NOT NULL,
  `client_id` varchar(255) NOT NULL,
  `subject_obfuscated` varchar(255) NOT NULL,
  `nid` char(36) NOT NULL,
  PRIMARY KEY (`subject`,`client_id`,`nid`),
  UNIQUE KEY `hydra_oauth2_obfuscated_authentication_session_so_nid_idx` (`client_id`,`subject_obfuscated`,`nid`),
  KEY `hydra_oauth2_obfuscated_authentication_session_nid_fk_idx` (`nid`),
  KEY `hydra_oauth2_obfuscated_authentication_session_client_id_fk` (`client_id`,`nid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_oauth2_oidc` (
  `signature` varchar(255) NOT NULL,
  `request_id` varchar(40) NOT NULL DEFAULT '',
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_id` varchar(255) NOT NULL DEFAULT '',
  `scope` text NOT NULL,
  `granted_scope` text NOT NULL,
  `form_data` text NOT NULL,
  `session_data` text NOT NULL,
  `subject` varchar(255) NOT NULL DEFAULT '',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `requested_audience` text NOT NULL,
  `granted_audience` text NOT NULL,
  `challenge_id` varchar(40) DEFAULT NULL,
  `nid` char(36) NOT NULL,
  CONSTRAINT `hydra_oauth2_oidc_challenge_id_fk` FOREIGN KEY (`challenge_id`) REFERENCES `hydra_oauth2_flow` (`consent_challenge_id`) ON DELETE CASCADE,
  PRIMARY KEY (`signature`),
  KEY `hydra_oauth2_oidc_challenge_id_idx` (`challenge_id`),
  KEY `hydra_oauth2_oidc_nid_fk_idx` (`nid`),
  KEY `hydra_oauth2_oidc_client_id_fk` (`client_id`,`nid`),
  KEY `hydra_oauth2_oidc_request_id_idx` (`request_id`,`nid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_oauth2_pkce` (
  `signature` varchar(255) NOT NULL,
  `request_id` varchar(40) NOT NULL DEFAULT '',
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_id` varchar(255) NOT NULL DEFAULT '',
  `scope` text NOT NULL,
  `granted_scope` text NOT NULL,
  `form_data` text NOT NULL,
  `session_data` text NOT NULL,
  `subject` varchar(255) NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `requested_audience` text NOT NULL,
  `granted_audience` text NOT NULL,
  `challenge_id` varchar(40) DEFAULT NULL,
  `nid` char(36) NOT NULL,
  CONSTRAINT `hydra_oauth2_pkce_challenge_id_fk` FOREIGN KEY (`challenge_id`) REFERENCES `hydra_oauth2_flow` (`consent_challenge_id`) ON DELETE CASCADE,
  PRIMARY KEY (`signature`),
  KEY `hydra_oauth2_pkce_challenge_id_idx` (`challenge_id`),
  KEY `hydra_oauth2_pkce_nid_fk_idx` (`nid`),
  KEY `hydra_oauth2_pkce_client_id_fk` (`client_id`,`nid`),
  KEY `hydra_oauth2_pkce_request_id_idx` (`request_id`,`nid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_oauth2_refresh` (
  `signature` varchar(255) NOT NULL,
  `request_id` varchar(40) NOT NULL DEFAULT '',
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_id` varchar(255) NOT NULL DEFAULT '',
  `scope` text NOT NULL,
  `granted_scope` text NOT NULL,
  `form_data` text NOT NULL,
  `session_data` text NOT NULL,
  `subject` varchar(255) NOT NULL DEFAULT '',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `requested_audience` text NOT NULL,
  `granted_audience` text NOT NULL,
  `challenge_id` varchar(40) DEFAULT NULL,
  `nid` char(36) NOT NULL,
  CONSTRAINT `hydra_oauth2_refresh_challenge_id_fk` FOREIGN KEY (`challenge_id`) REFERENCES `hydra_oauth2_flow` (`consent_challenge_id`) ON DELETE CASCADE,
  PRIMARY KEY (`signature`),
  KEY `hydra_oauth2_refresh_challenge_id_idx` (`challenge_id`),
  KEY `hydra_oauth2_refresh_client_id_fk` (`client_id`,`nid`),
  KEY `hydra_oauth2_refresh_client_id_subject_idx` (`client_id`,`subject`),
  KEY `hydra_oauth2_refresh_request_id_idx` (`request_id`),
  KEY `hydra_oauth2_refresh_requested_at_idx` (`nid`,`requested_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `hydra_oauth2_trusted_jwt_bearer_issuer` (
  `id` varchar(36) NOT NULL,
  `issuer` varchar(255) NOT NULL,
  `subject` varchar(255) NOT NULL,
  `scope` text NOT NULL,
  `key_set` varchar(255) NOT NULL,
  `key_id` varchar(255) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `expires_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `nid` char(36) NOT NULL,
  `allow_any_subject` tinyint(1) NOT NULL DEFAULT '0',
  CONSTRAINT `hydra_oauth2_trusted_jwt_bearer_issuer_ibfk_1` FOREIGN KEY (`key_set`, `key_id`, `nid`) REFERENCES `hydra_jwk` (`sid`, `kid`, `nid`) ON DELETE CASCADE,
  PRIMARY KEY (`id`),
  UNIQUE KEY `issuer` (`issuer`,`subject`,`key_id`),
  KEY `hydra_oauth2_trusted_jwt_bearer_issuer_nid_fk_idx` (`nid`),
  KEY `hydra_oauth2_trusted_jwt_bearer_issuer_ibfk_1` (`key_set`,`key_id`,`nid`),
  KEY `hydra_oauth2_trusted_jwt_bearer_issuer_expires_at_idx` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

SET FOREIGN_KEY_CHECKS = 1;

INSERT INTO `networks`
(`id`, `created_at`, `updated_at`)
SELECT
    (SELECT LOWER(UUID())),
    '2013-10-07 08:23:19',
    '2013-10-07 08:23:19'
FROM DUAL
WHERE NOT EXISTS (SELECT `created_at` FROM `networks` WHERE `created_at` = '2013-10-07 08:23:19');

-- ShareMgnt

/*
MySQL: Database - sharemgnt
*********************************************************************
*/
use sharemgnt_db;

CREATE TABLE IF NOT EXISTS `t_user` (
    `f_user_id` char(40) NOT NULL,                                  -- 用户id
    `f_login_name` CHAR(150) NOT NULL,                              -- 登录名
    `f_display_name`  CHAR(150) NOT NULL,                           -- 显示名
    `f_remark` char(128),                                           -- 备注
    `f_idcard_number` char(32),                                     -- 身份证号
    `f_password` char(32) NOT NULL,                                 -- 密码的校验值
    `f_des_password` char(150) DEFAULT '',                          -- 管控密码的校验值
    `f_ntlm_password` char(32) DEFAULT '',                          -- ntlm密码的校验值
    `f_sha2_password` char(64) DEFAULT '',                          -- sha2密码的校验值
    `f_mail_address` char(150) NOT NULL,                            -- 邮箱地址
    `f_auth_type` tinyint(4) NOT NULL DEFAULT '0',                  -- 认证类型, 1为本地用户, 2为域用户, 3为第三方用户
    `f_status` tinyint(4) NOT NULL DEFAULT '0',                     -- 用户禁用状态, 0为正常使用, 1为禁用
    `f_freeze_status` tinyint(4) NOT NULL DEFAULT '0',              -- 用户冻结状态, 0为不冻结, 1为冻结
    `f_pwd_timestamp` datetime,                                     -- 密码修改时间
    `f_pwd_error_latest_timestamp` datetime,                        -- 上次密码输入错误的时间
    `f_pwd_error_cnt` tinyint(4) NOT NULL DEFAULT '0',              -- 密码错误次数
    `f_domain_object_guid`char(100) DEFAULT '',                     -- 域对象的guid
    `f_domain_path` char(255) DEFAULT '',                           -- 域路径
    `f_ldap_server_type` tinyint(4) NOT NULL DEFAULT '0',           -- ldap服务器类型
    `f_third_party_id` char(255),                                   -- 第三方系统中的id
    `f_third_party_depart_id` varchar(255),                         -- 第三方系统中的部门id
    `f_priority` smallint(6) NOT NULL DEFAULT '999',                -- 用户优先级
    `f_csf_level` tinyint(4) NOT NULL DEFAULT '5',                  -- 用户密级
    `f_pwd_control` tinyint(1) NOT NULL DEFAULT '0',                -- 本地用户的密码管控, 0为不使用密码管控, 1为使用
    `f_oss_id` char(40),                                            -- 用户归属对象存储
    `f_create_time` datetime DEFAULT now(),                         -- 用户创建时间
    `f_last_request_time` datetime DEFAULT now(),                   -- 用户最后一次请求的时间
    `f_last_client_request_time` datetime NOT NULL DEFAULT now(),   -- 用户在客户端最后一次请求的时间
    `f_auto_disable_status` tinyint(4) NOT NULL DEFAULT '0',        -- 用户自动禁用状态, 1为长时间不登录禁用, 2为用户过期禁用
    `f_agreed_to_terms_of_use` tinyint(4) NOT NULL DEFAULT '0',     -- 本地用户是否同意用户使用协议
    `f_real_name_auth_status` tinyint(4) NOT NULL DEFAULT '0',      -- 实名状态
    `f_tel_number` char(40) DEFAULT NULL,                           -- 电话号码
    `f_is_activate` tinyint(4) NOT NULL DEFAULT '0',                -- 是否激活
    `f_activate_status` tinyint(4) NOT NULL DEFAULT '0',            -- 用户是否登录过系统
    `f_third_party_attr` varchar(255) NOT NULL DEFAULT '',          -- 第三方应用属性
    `f_expire_time` int(11) NOT NULL DEFAULT '-1',                  -- 用户账号有效期, 单位为秒, 默认永久有效
    `f_user_document_read_status` bigint(20) DEFAULT '0',           -- 用户的文档已读状态，目前用于快速入门
    `f_manager_id` varchar(40) NOT NULL DEFAULT '' COMMENT '用户上级ID',
    `f_code` varchar(255) NOT NULL DEFAULT '' COMMENT '用户编码',
    `f_position` varchar(50) NOT NULL DEFAULT '' COMMENT '岗位',
    `f_csf_level2` tinyint(4) NOT NULL DEFAULT '51' COMMENT '用户密级2',
    PRIMARY KEY (`f_user_id`),
    KEY `f_mail_address_index` (`f_mail_address`),
    KEY `f_domain_object_guid` (`f_domain_object_guid`),
    UNIQUE KEY `f_login_name` (`f_login_name`),
    KEY `f_display_name_index` (`f_display_name`),
    KEY `f_remark_index` (`f_remark`),
    KEY `f_idcard_number_index` (`f_idcard_number`),
    KEY `f_third_party_id_index` (`f_third_party_id`),
    KEY `f_tel_number_index` (`f_tel_number`),
    KEY `f_priority_name_index` (`f_priority`,`f_display_name`),
    KEY `idx_t_user_code` (`f_code`),
    KEY `idx_t_user_position` (`f_position`),
    KEY `idx_t_user_manager_id` (`f_manager_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_third_party_auth` (
    `f_id` int(11) NOT NULL AUTO_INCREMENT,                         -- 自增主键
    `f_app_id` varchar(128) NOT NULL,                                -- 第三方App Id
    `f_app_name` varchar(128) NOT NULL DEFAULT '',                  -- 第三方App名
    `f_enable` tinyint(1) NOT NULL DEFAULT 0,                       -- 是否启用, 1为启用, 0为禁用
    `f_config` text,                                                -- 第三方配置, 外部可见
    `f_internal_config` text,                                       -- 第三方配置, 内部使用, 外部不可见
    `f_plugin_name` varchar(255) NOT NULL,                           -- 第三方插件名称
    `f_plugin_type` tinyint(4) NOT NULL,                            -- 第三方种类, 0: 认证, 1: 消息
    `f_object_id` char(110) NOT NULL,                               -- 文件ID, 用于确定在存储中的位置(兼容旧版本文件在对象存储的key，旧版本为三段结构(evfs前缀/cid/object_id)，新版本为一段(object_id))
    `f_oss_id` char(40) NOT NULL,                                   -- 第三方插件上传对象存储
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `f_app_id` (`f_app_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_person_group` (
    `f_group_id` char(40) NOT NULL,                                 -- 联系人组id
    `f_user_id` char(40) NOT NULL,                                  -- 用户id
    `f_group_name` char(128) NOT NULL,                              -- 联系人组名
    `f_person_count` bigint(20) NOT NULL,                           -- 人员总数
    PRIMARY KEY (`f_group_id`),
    KEY `f_user_id` (`f_user_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_contact_person` (
    `f_id` int(11) NOT NULL AUTO_INCREMENT,                         -- 自增主键
    `f_group_id` char(40) NOT NULL,                                 -- 联系人组id
    `f_user_id` char(40) NOT NULL,                                  -- 用户id
    CONSTRAINT `t_contact_person_ibfk_1` FOREIGN KEY (`f_group_id`) REFERENCES `t_person_group` (`f_group_id`),
    PRIMARY KEY (`f_id`),
    KEY `f_group_id` (`f_group_id`,`f_user_id`),
    KEY `f_user_id` (`f_user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1;

CREATE TABLE IF NOT EXISTS `t_domain` (
    `f_domain_id` bigint(20) NOT NULL AUTO_INCREMENT,               -- 域标识id
    `f_domain_name` varchar(253) NOT NULL,                          -- 域名
    `f_ip_address` varchar(253) NOT NULL,                           -- IP
    `f_port` bigint(20) NOT NULL,                                   -- 端口
    `f_administrator` char(50) NOT NULL,                            -- 账户
    `f_password` char(255) NOT NULL,                                -- 密码
    `f_parent_domain_id` bigint(20) NOT NULL,                       -- 父域id
    `f_domain_type` tinyint(4) NOT NULL,                            -- 域类型
    `f_status` tinyint(4) NOT NULL,                                 -- 状态, 0为禁用, 1为启用
    `f_ldap_server_type` tinyint(4) NOT NULL DEFAULT '1',           -- LDAP服务器类型
    `f_sync` tinyint(4) NOT NULL DEFAULT 0,                         -- 同步状态, -1为关闭域同步, 0为开启正向同步, 1为开启反相同步
    `f_use_ssl` tinyint(4) NOT NULL DEFAULT 0,                      -- 是否使用SSL
    `f_config` text,                                                -- 配置
    `f_key_config` text,                                            -- 关键配置
    PRIMARY KEY (`f_domain_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_failover_domain` (
    `f_domain_id` bigint(20) NOT NULL AUTO_INCREMENT,               -- 域标识id
    `f_parent_domain_id` bigint(20) NOT NULL,                       -- 主域id
    `f_ip_address` char(50) NOT NULL,                               -- IP
    `f_port` bigint(20) NOT NULL,                                   -- 端口
    `f_administrator` char(50) NOT NULL,                            -- 账户
    `f_password` char(255) NOT NULL,                                -- 密码
    `f_use_ssl` tinyint(4) NOT NULL DEFAULT 0,                      -- 是否使用SSL
    PRIMARY KEY (`f_domain_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_department` (
    `f_department_id` char(40) NOT NULL,                            -- 部门id
    `f_auth_type` tinyint(4) NOT NULL,                              -- 认证类型, 1为本地创建的部门, 2为域控导入的部门, 3为第三方导入的部门
    `f_name` char(128) NOT NULL,                                    -- 部门名
    `f_domain_object_guid` char(100) DEFAULT '',                    -- 域对象的guid
    `f_domain_path` char(255) DEFAULT '',                           -- 域路径
    `f_is_enterprise` tinyint(4) NOT NULL,                          -- 是否为组织
    `f_third_party_id` char(255) DEFAULT '',                        -- 第三方系统中的id
    `f_priority` mediumint(9) NOT NULL DEFAULT '999999',            -- 部门优先级
    `f_oss_id` char(40),                                            -- 部门下用户的对象存储
    `f_mail_address` char(150) NOT NULL DEFAULT '',                 -- 邮箱地址
    `f_path` text NOT NULL,                                         -- 部门全路径
    `f_manager_id` varchar(40) NOT NULL DEFAULT '' COMMENT '负责人ID',
    `f_status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '启用状态，1：启用 2：停用',
    `f_code` varchar(255) NOT NULL DEFAULT '' COMMENT '部门编码',
    `f_remark` varchar(128) NOT NULL DEFAULT '' COMMENT '部门备注',
    PRIMARY KEY (`f_department_id`),
    UNIQUE KEY `f_department_id_index` (`f_department_id`),
    KEY `f_name_index` (`f_name`),
    KEY `f_mail_address_index` (`f_mail_address`),
    KEY `f_is_enterprise_index` (`f_is_enterprise`),
    KEY `f_third_party_id_index` (`f_third_party_id`),
    KEY `f_path_index` (`f_path`(480)),                              -- 根据中铁30w部门层级最大深度13建立索引长度，480 = 37*13-1
    KEY `idx_t_department_manager_id` (`f_manager_id`),
    KEY `idx_t_department_status` (`f_status`),
    KEY `idx_t_department_code` (`f_code`),
    KEY `idx_t_department_remark` (`f_remark`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_department_relation` (
    `f_relation_id` bigint(20) NOT NULL AUTO_INCREMENT,             -- 自增主键
    `f_department_id` char(40) NOT NULL,                            -- 部门id
    `f_parent_department_id` char(40) NOT NULL,                     -- 父部门id
    PRIMARY KEY (`f_relation_id`),
    UNIQUE KEY `f_department_id_index` (`f_department_id`),
    KEY `f_parent_department_id_index` (`f_parent_department_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_user_department_relation` (
    `f_relation_id` bigint(20) NOT NULL AUTO_INCREMENT,             -- 自增主键
    `f_user_id` char(40) NOT NULL,                                  -- 用户id
    `f_department_id` char(40) NOT NULL,                            -- 用户所在部门id
    `f_path` text NOT NULL,                                         -- 用户所在部门全路径
    PRIMARY KEY (`f_relation_id`),
    KEY `f_user_id_index` (`f_user_id`),
    KEY `f_department_id_index` (`f_department_id`),
    KEY `f_path_index` (`f_path`(480))                              -- 根据中铁30w部门层级最大深度13建立索引长度，480 = 37*13-1
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_online_user_real_time` (
    `f_time` char(40) NOT NULL,                                     -- 时间字符串
    `f_count` bigint(20) NOT NULL,                                  -- 总数
    `f_uuid` char(40) NOT NULL,                                     -- 记录的唯一标识
    PRIMARY KEY (`f_uuid`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_max_online_user_day` (
`f_time` char(40) NOT NULL,                                     -- 时间字符串
`f_count` bigint(20) NOT NULL,                                  -- 总数
`f_uuid` char(40) NOT NULL,                                     -- 记录的唯一标识
PRIMARY KEY (`f_uuid`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_max_online_user_month` (
    `f_time` char(40) NOT NULL,                                     -- 时间字符串
    `f_count` bigint(20) NOT NULL,                                  -- 总数
    `f_uuid` char(40) NOT NULL,                                     -- 记录的唯一标识
    PRIMARY KEY (`f_uuid`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_ou_user` (
    `f_id` int(11) NOT NULL AUTO_INCREMENT,                         -- 自增主键
    `f_user_id` char(40) NOT NULL,                                  -- 用户id
    `f_ou_id` char(40) NOT NULL,                                    -- OU中标识id
    PRIMARY KEY (`f_id`),
    KEY `f_user_id_index` (`f_user_id`),
    KEY `f_ou_id_index` (`f_ou_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_ou_department` (
    `f_id` int(11) NOT NULL AUTO_INCREMENT,                         -- 自增主键
    `f_department_id` char(40) NOT NULL,                            -- 部门id
    `f_ou_id` char(40) NOT NULL,                                    -- OU中标识id
    PRIMARY KEY (`f_id`),
    KEY `f_department_id_index` (`f_department_id`),
    KEY `f_ou_id_index` (`f_ou_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_license` (
  `f_license_value` char(255) NOT NULL,                             -- 许可证值
  `f_active` tinyint(4) NOT NULL DEFAULT '0',                       -- 是否激活
  `f_type` tinyint(4) NOT NULL DEFAULT '0',                         -- 类型
  `f_version` tinyint(4) NOT NULL DEFAULT '0',                      -- 许可证版本, 5: 5.0许可证, 6: 6.0许可证
  PRIMARY KEY (`f_license_value`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_license_used` (
  `f_license_value` char(255) NOT NULL,                             -- 许可证值
  `f_active_time` bigint(20) NOT NULL DEFAULT '-1',                 -- 激活时间, 微秒的时间戳
  `f_type` tinyint(4) NOT NULL DEFAULT '0',                         -- 类型
  PRIMARY KEY (`f_license_value`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_oem_config` (
  `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,               -- 自增主键
  `f_section` char(32) NOT NULL,                                    -- 内容分类
  `f_option` char(32) NOT NULL,                                     -- 选项
  `f_value` mediumblob NOT NULL,                                    -- 值
  PRIMARY KEY (`f_primary_id`),
  UNIQUE KEY `f_index_section_option` (`f_section`,`f_option`) USING BTREE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_sharemgnt_config` (
  `f_key` char(32) NOT NULL,                                        -- 配置关键字
  `f_value` varchar(1024) NOT NULL,                                 -- 配置的值
  PRIMARY KEY (`f_key`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_perm_share_strategy` (
    `f_index` int(11) NOT NULL AUTO_INCREMENT,                      -- 自增主键
    `f_strategy_id` char(40) NOT NULL,                              -- 策略id
    `f_obj_id` char(40) NOT NULL,                                   -- 对象id
    `f_obj_type` tinyint(4) NOT NULL,                               -- 对象类型
    `f_parent_id` char(40),                                         -- 父对象id
    `f_sharer_or_scope` tinyint(4) NOT NULL,                        -- 共享者或共享范围
    `f_status` tinyint(4) NOT NULL DEFAULT '0',                     -- 策略状态
    PRIMARY KEY (`f_index`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_link_share_strategy` (
    `f_sharer_id` char(40) NOT NULL,                                -- 共享者id
    `f_sharer_type` tinyint(4) NOT NULL,                            -- 共享者类型
    PRIMARY KEY (`f_sharer_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_find_share_strategy` (
    `f_sharer_id` char(40) NOT NULL,                                -- 共享者id
    `f_sharer_type` tinyint(4) NOT NULL,                            -- 共享者类型
    PRIMARY KEY (`f_sharer_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_leak_proof_strategy` (
  `f_strategy_id` bigint(20) NOT NULL AUTO_INCREMENT,               -- 自增主键
  `f_accessor_id` char(40) NOT NULL,                                -- 访问者id
  `f_accessor_type` tinyint(4) NOT NULL,                            -- 访问者类型
  `f_perm_value` int(11) DEFAULT NULL,                              -- 权限值
  PRIMARY KEY (`f_strategy_id`),
  UNIQUE KEY `f_accessor_id_index` (`f_accessor_id`) USING HASH
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_cert` (
  `f_key` char(32) COLLATE utf8mb4_bin NOT NULL,                    -- 配置关键字
  `f_value` varchar(8192) COLLATE utf8mb4_bin NOT NULL,             -- 配置的值
  PRIMARY KEY (`f_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS `t_client_update_package` (
  `f_id` int  NOT NULL AUTO_INCREMENT,                              -- 自增主键
  `f_name` varchar(150) NOT NULL,                                   -- 客户端包名
  `f_os` int  NOT NULL ,                                            -- 系统类型
  `f_size` bigint(20) NOT NULL,                                     -- 安装包大小
  `f_version` varchar(50) NOT NULL,                                 -- 安装包版本
  `f_time` varchar(50) NOT NULL,                                    -- 安装包上传时间
  `f_mode` tinyint(1) NOT NULL,                                     -- 升级类型
  `f_pkg_location` tinyint(4) NOT NULL DEFAULT '1',                 -- 升级包位置，1表示本地上传到对象存储，2表示独立配置升级包下载地址
  `f_url` text NOT NULL,                                            -- 下载地址
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `f_os_index` (`f_os`) USING BTREE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_site_info` (
  `f_site_id` varchar(36) NOT NULL,                                 -- 站点id
  `f_site_ip` varchar(64) DEFAULT NULL,                             -- 站点IP
  `f_site_name` varchar(128) NOT NULL,                              -- 站点名称
  `f_site_type` tinyint(3) NOT NULL,                                -- 站点类型, 0为普通站点, 1为总站点, 2为分站点
  `f_site_link_status` tinyint(1) DEFAULT NULL,                     -- 站点连接状态
  `f_site_status` tinyint(1) NOT NULL DEFAULT '1',                  -- 站点启用状态, 1为启用, 2为禁用
  `f_site_used_space` bigint(20) NOT NULL DEFAULT '0',              -- 站点已用存储空间
  `f_site_total_space` bigint(20) NOT NULL DEFAULT '0',             -- 站点总存储空间
  `f_site_key` varchar(10) NOT NULL,                                -- 站点标识key
  `f_site_master_ip` varchar(64) DEFAULT NULL,                      -- 主站点IP
  `f_site_is_sync` tinyint(1) NOT NULL DEFAULT '0',                 -- 站点信息是否同步
  `f_site_heart_rate` bigint(20) DEFAULT NULL,                      -- 心跳信息
  `f_uniq_index` int  NOT NULL AUTO_INCREMENT,                      -- 自增主键
  `f_site_master_db_ip` varchar(64) DEFAULT NULL,                   -- 主站点数据库ip
  `f_site_need_update_virusdb` tinyint(1) NOT NULL DEFAULT '0',     -- 站点病毒库的更新状态
  PRIMARY KEY (`f_site_id`),
  UNIQUE KEY `f_uniq_index_index` (`f_uniq_index`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS `t_third_party_db` (
    `f_third_db_id` char(50) NOT NULL,                              -- 第三方数据库标识id
    `f_name` char(50) DEFAULT "",                                   -- 第三方名称
    `f_ip` char(50) NOT NULL,                                       -- 数据库IP
    `f_port` bigint(20) NOT NULL,                                   -- 数据库端口
    `f_admin` char(50) NOT NULL,                                    -- 数据库用户名
    `f_password` char(50) NOT NULL,                                 -- 数据库密码
    `f_database` char(50) NOT NULL,                                 -- 数据库
    `f_db_type` tinyint(4) NOT NULL,                                -- 数据库类型
    `f_charset` char(50) NOT NULL DEFAULT '',                       -- 数据库字符集
    `f_status` tinyint(4) NOT NULL DEFAULT 0,                       -- 状态
    `f_parent_department_id` char(50) DEFAULT "",                   -- 父部门id
    `f_third_root_name` char(50) DEFAULT "",                        -- 第三方根对象名
    `f_third_root_id` char(50) DEFAULT "",                          -- 第三方根对象id
    `f_sync_interval` int(10) DEFAULT 3600,                         -- 同步周期
    `f_space_size` char(40) DEFAULT '5368709120',                   -- 用户空间配额
    `f_user_type` tinyint(4) DEFAULT 3,                             -- 用户类型
    PRIMARY KEY (`f_third_db_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_third_depart_table` (
    `f_table_id` char(50) NOT NULL,                                 -- 第三方数据表标识id
    `f_third_db_id` char(50) NOT NULL,                              -- 第三方数据库标识id
    `f_table_name` char(50) NOT NULL,                               -- 第三方数据表名
    `f_department_id` char(50) DEFAULT "",                          -- 部门id
    `f_department_name` char(50) DEFAULT "",                        -- 部门名
    `f_deparment_priority` char(50) DEFAULT "",                     -- 部门优先级
    `f_filter` text,                                                -- 过滤条件
    `f_sub_group` text,                                             -- 下级组
    PRIMARY KEY (`f_table_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_third_depart_relation_table` (
    `f_table_id` char(50) NOT NULL,                                 -- 第三方数据表标识id
    `f_third_db_id` char(50) NOT NULL,                              -- 第三方数据库标识id
    `f_table_name` char(50) NOT NULL,                               -- 第三方数据表名
    `f_department_id` char(50) DEFAULT "",                          -- 部门id
    `f_parent_department_id` char(50) DEFAULT "",                   -- 上级部门id
    `f_parent_group_table_id` char(50) DEFAULT "",                  -- 上级组表标识id
    `f_parent_group_name` char(50) DEFAULT "",                      -- 上级组名
    `f_filter` text,                                                -- 过滤条件
    PRIMARY KEY (`f_table_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_third_user_table` (
    `f_table_id` char(50) NOT NULL,                                 -- 第三方数据表标识id
    `f_third_db_id` char(50) NOT NULL,                              -- 第三方数据库标识id
    `f_table_name` char(50) NOT NULL,                               -- 第三方数据表名
    `f_user_id` char(50) DEFAULT "",                                -- 用户id
    `f_user_login_name` char(50) DEFAULT "",                        -- 用户登录名
    `f_user_display_name` char(50) DEFAULT "",                      -- 用户显示名
    `f_user_email` char(50) DEFAULT "",                             -- 用户邮箱地址
    `f_user_password` char(50) DEFAULT "",                          -- 用户密码
    `f_user_status` char(50) DEFAULT "",                            -- 用户状态
    `f_user_priority` char(50) DEFAULT "",                          -- 用户优先级
    `f_filter` text,                                                -- 过滤条件
    PRIMARY KEY (`f_table_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_third_user_relation_table` (
    `f_table_id` char(50) NOT NULL,                                 -- 第三方数据表标识id
    `f_third_db_id` char(50) NOT NULL,                              -- 第三方数据库标识id
    `f_table_name` char(50) NOT NULL,                               -- 第三方数据表名
    `f_user_id` char(50) DEFAULT "",                                -- 用户id
    `f_parent_department_id` char(50) DEFAULT "",                   -- 上级部门id
    `f_parent_group_table_id` char(50) DEFAULT "",                  -- 上级组表id
    `f_parent_group_name` char(50) DEFAULT "",                      -- 上级组名
    `f_filter` text,                                                -- 过滤条件
    PRIMARY KEY (`f_table_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_third_auth_info` (
    `f_app_id` varchar(50) NOT NULL,                                -- 第三方App Id
    `f_app_key` char(36) NOT NULL,                                  -- 第三方App Key
    `f_enabled` tinyint(1) NOT NULL DEFAULT 1,                      -- 是否启用, 1为启用, 0为禁用
    PRIMARY KEY (`f_app_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_third_party_tool_config` (
    `f_tool_id` char(128) NOT NULL,                                 -- 工具唯一标识id
    `f_enabled` tinyint(1) NOT NULL DEFAULT 0,                      -- 是否启用, 0为禁用, 1为启用
    `f_url` text,                                                   -- url访问地址
    `f_tool_name` char(128) NOT NULL,                               -- 第三方工具名称, 仅在工具标识为"CAD"时保存, 合法名称为"hc"或"mx"
    `f_app_id` char(50),                                            -- 鉴权唯一标识
    `f_app_key` char(150),                                          -- 鉴权密钥
    PRIMARY KEY (`f_tool_id`)
)ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_net_accessors_info` (
    `f_index` int(11) NOT NULL AUTO_INCREMENT,                      -- 自增主键
    `f_id` char(40) NOT NULL,                                       -- 记录id
    `f_ip` char(15) NOT NULL,                                       -- IP
    `f_sub_net_mask` char(15) NOT NULL,                             -- 子网掩码
    `f_accessor_id` char(40),                                       -- 访问者id
    `f_accessor_type` tinyint(4) NOT NULL,                          -- 访问者类型
    PRIMARY KEY (`f_index`)
)ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_nas_node` (
  `f_uuid` varchar(40) NOT NULL,                                    -- 节点标识, UUID
  PRIMARY KEY (`f_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS `t_limit_rate` (
  `f_id` varchar(40) NOT NULL,                                      -- 限速规则id
  `f_obj_id` varchar(40) NOT NULL,                                  -- 对象id
  `f_obj_type` tinyint(4) NOT NULL,                                 -- 对象类型, 1为用户, 2为部门
  `f_limit_type` tinyint(4) NOT NULL,                               -- 限速类型, 0为用户, 1为部门
  `f_upload_rate` int NOT NULL,                                     -- 上传限速值
  `f_download_rate` int NOT NULL,                                   -- 下载限速值
  PRIMARY KEY (`f_id`, `f_obj_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS `t_nginx_user_rate` (
  `f_userid` varchar(40) NOT NULL,                                  -- 用户id
  `f_parent_deptids` text NOT NULL,                                 -- 上一层有规则的父部门id, 没有时置为-1, 用户级别限速时为空
  `f_download_req_cnt` int DEFAULT 0,                               -- 下载请求数量
  `f_upload_req_cnt` int DEFAULT 0,                                 -- 上传请求数量
  `f_upload_rate` int DEFAULT 0,                                    -- 上传限速
  `f_download_rate` int DEFAULT 0,                                  -- 下载限速
  PRIMARY KEY (`f_userid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS `t_department_responsible_person` (
  `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,               -- 自增主键
  `f_department_id` char(40) NOT NULL,                              -- 部门id
  `f_user_id` char(40) NOT NULL,                                    -- 管理者的用户id
  PRIMARY KEY (`f_primary_id`),
  UNIQUE KEY `responsible_person_depart_index` (`f_user_id`,`f_department_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_watermark_config` (
    `f_for_user_doc` tinyint(4) NOT NULL DEFAULT '0',               -- 是否用于个人文档
    `f_for_custom_doc` tinyint(4) NOT NULL DEFAULT '0',             -- 是否用于自定义文档库
    `f_for_archive_doc` tinyint(4) NOT NULL DEFAULT '0',            -- 是否用于归档库
    `f_config` mediumblob NOT NULL,                                 -- 配置内容
    `f_index` int(11) NOT NULL AUTO_INCREMENT,                      -- 自增主键
    PRIMARY KEY (`f_index`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_watermark_doc` (
    `f_obj_id` char(255) NOT NULL,                                  -- 文档库对象id
    `f_watermark_type` tinyint(4) NOT NULL,                         -- 水印类型
    `f_time` bigint(20) NOT NULL,                                   -- 记录的时间
    PRIMARY KEY (`f_obj_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_link_template` (
  `f_index` int(11) NOT NULL AUTO_INCREMENT,                        -- 自增主键
  `f_template_id` char(40) NOT NULL,                                -- 模板id
  `f_template_type` tinyint(4) NOT NULL,                            -- 模板类型
  `f_sharer_id` char(40) NOT NULL,                                  -- 共享者id
  `f_sharer_type` tinyint(1) NOT NULL,                              -- 共享者类型
  `f_create_time` bigint(20) NOT NULL,                              -- 记录创建时间, 微秒的时间戳
  `f_config` text NOT NULL,                                         -- 配置信息
  PRIMARY KEY (`f_index`),
  KEY `f_template_id_index` (`f_template_id`),
  KEY `f_sharer_id_index` (`f_sharer_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_net_docs_limit_info` (
    `f_index` int(11) NOT NULL AUTO_INCREMENT,                      -- 自增主键
    `f_id` char(40) NOT NULL,                                       -- 记录id
    `f_ip` char(15) NOT NULL,                                       -- IP
    `f_sub_net_mask` char(15) NOT NULL,                             -- 子网掩码
    `f_doc_id` char(40),                                            -- 文档入口id
    PRIMARY KEY (`f_index`),
    KEY `f_doc_id_index` (`f_doc_id`)
)ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_user_verification_code` (
  `f_user_id` char(40) NOT NULL,                                    -- 用户id
  `f_vcode` varchar(40) NOT NULL,                                   -- 验证码值
  `f_create_time` bigint(20) NOT NULL,                              -- 验证码创建时间, 微秒的时间戳
  PRIMARY KEY (`f_user_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_antivirus_admin` (
    `f_user_id` char(40) NOT NULL,                                  -- 用户id
    PRIMARY KEY (`f_user_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_hide_ou` (
    `f_department_id` char(40) NOT NULL,                            -- 部门id
    PRIMARY KEY (`f_department_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_vcode` (
    `f_uuid` char(40) NOT NULL,                                     -- 记录标识, UUID
    `f_vcode` char(40) NOT NULL,                                    -- 校验码的值
    `f_vcode_type` int(4) default 1,                                -- 校验码类型, 1: 其他情况, 2: 忘记密码时创建的验证码
    `f_vcode_error_cnt` int(4) default 0,                           -- 验证码输入错误次数
    `f_createtime` datetime DEFAULT now(),                          -- 创建时间
    PRIMARY KEY (`f_uuid`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_recycle` (
    `f_cid` char(40) NOT NULL,                                      -- 文档入口CID
    `f_gns` char(80) NOT NULL,                                      -- 文档路径标识
    `f_setter` char(40) NOT NULL,                                   -- 配置者id
    `f_retention_days` int DEFAULT -1,                              -- 保留天数, -1为不自动清理
    PRIMARY KEY (`f_cid`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_sms_code` (
    `f_tel_number` char(40) NOT NULL,                               -- 电话号码
    `f_verify_code` char(6) NOT NULL,                               -- 确认码
    `f_create_time` datetime DEFAULT now(),                         -- 创建时间
    PRIMARY KEY (`f_tel_number`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_copy_limit_rate` (
  `f_obj_id` varchar(40) NOT NULL,                                  -- 限速规则id
  `f_parent_id` varchar(40) NOT NULL,                               -- 上一级有规则的父部门id
  `f_obj_type` tinyint(4) NOT NULL,                                 -- 限速对象类型, 1为用户, 2为部门
  `f_upload_rate` int NOT NULL,                                     -- 上传限速值
  `f_download_rate` int NOT NULL,                                   -- 下载限速值
  PRIMARY KEY (`f_obj_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS `t_active_user_day` (
    `f_id` bigint(20) NOT NULL AUTO_INCREMENT,                      -- 自增主键
    `f_time` char(40) NOT NULL,                                     -- 时间字符串
    `f_active_count` bigint(20) NOT NULL DEFAULT '0',               -- 活跃用户数
    `f_activate_count` bigint(20) NOT NULL DEFAULT '0',             -- 激活用户数
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `f_time_index` (`f_time`) USING BTREE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_active_user_month` (
    `f_id` bigint(20) NOT NULL AUTO_INCREMENT,                      -- 自增主键
    `f_time` char(40) NOT NULL,                                     -- 时间字符串
    `f_active_count` bigint(20) NOT NULL DEFAULT '0',               -- 活跃用户数
    `f_total_count` bigint(20) NOT NULL DEFAULT '0',                -- 当月用户总数
    `f_activate_count` bigint(20) NOT NULL DEFAULT '0',             -- 激活用户数
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `f_time_index` (`f_time`) USING BTREE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_active_user_year` (
    `f_id` bigint(20) NOT NULL AUTO_INCREMENT,                      -- 自增主键
    `f_time` char(40) NOT NULL,                                     -- 时间字符串
    `f_total_count` bigint(20) NOT NULL DEFAULT '0',                -- 当年用户总数
    `f_activate_count` bigint(20) NOT NULL DEFAULT '0',             -- 激活用户数
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `f_time_index` (`f_time`) USING BTREE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_operation_problem` (
    `f_id` bigint(20) NOT NULL AUTO_INCREMENT,                      -- 自增主键
    `f_ip` char(40) NOT NULL,                                       -- IP
    `f_time` bigint(20) NOT NULL,                                   -- 异常开始至结束的时间中间值
    `f_time_from` bigint(20) NOT NULL DEFAULT '0',                  -- 异常开始时间
    `f_time_util` bigint(20) NOT NULL DEFAULT '0',                  -- 异常结束时间
    `f_obj_id` char(40) NOT NULL,                                   -- 触发异常的triggerid或历史数据异常的itemid
    `f_type` tinyint(4) NOT NULL,                                   -- 异常类型
    `f_description` text NOT NULL,                                  -- trigger或item的名称
    `f_monitoring_range` char(40) DEFAULT NULL,                     -- 异常监控值范围
    PRIMARY KEY (`f_id`),
    KEY `f_time_index` (`f_time`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_role`(
    `f_role_id` char(40) NOT NULL,                                  -- 角色id
    `f_name` char(150) NOT NULL DEFAULT '',                         -- 角色名称
    `f_description` text NOT NULL,                                  -- 角色职能描述
    `f_creator_id` char(40) NOT NULL DEFAULT '',                    -- 角色创建者id
    `f_priority` smallint(6) NOT NULL DEFAULT '999',                -- 角色权重值
    PRIMARY KEY (`f_role_id`),
    KEY `f_creator_id_index` (`f_creator_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_user_role_relation` (
    `f_user_id` char(40) NOT NULL,                                  -- 用户id
    `f_role_id` char(40) NOT NULL,                                  -- 角色id
    PRIMARY KEY (`f_user_id`, `f_role_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_department_audit_person` (
    `f_user_id` char(40) NOT NULL,                                  -- 用户id
    `f_department_id` char(40) NOT NULL,                            -- 部门id
    PRIMARY KEY (`f_user_id`, `f_department_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_user_role_attribute` (
    `f_user_id` char(40) NOT NULL,                                  -- 用户id
    `f_mail_address` varchar(1024) NOT NULL,                        -- 用户角色邮箱列表
    PRIMARY KEY (`f_user_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_file_crawl_strategy`(
    `f_strategy_id` int(8) NOT NULL AUTO_INCREMENT,                 -- 自增主键
    `f_user_id` char(40) NOT NULL,                                  -- 用户id
    `f_doc_id` char(40) NOT NULL,                                   -- 文档库路径
    `f_file_crawl_type` text NOT NULL,                              -- 抓取类型, 后缀名+空格组成
    PRIMARY KEY (`f_strategy_id`),
UNIQUE KEY `f_user_id_index` (`f_user_id`) USING BTREE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_doc_auto_archive_strategy` (
  `f_index` bigint(20) NOT NULL AUTO_INCREMENT,                     -- 唯一自增id
  `f_strategy_id` char(40) NOT NULL,                                -- 策略id，不同用户、部门的策略id可能相同，返回给前端时会合并
  `f_obj_id` char(40) NOT NULL,                                     -- 对象id，可能是用户、部门id
  `f_obj_type` tinyint(4) NOT NULL,                                 -- 对象类型，1：用户 2：部门
  `f_archive_dest_doc_id` char(40) NOT NULL,                        -- 目的归档库gns
  `f_archive_cycle` bigint(20) NOT NULL,                            -- 归档周期，天数
  `f_archive_cycle_modify_time` bigint(20) NOT NULL,                -- 归档周期的变更时间
  `f_create_time` bigint(20) NOT NULL,                              -- 记录创建时间
  PRIMARY KEY (`f_index`),
  UNIQUE KEY `f_obj_id` (`f_obj_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_doc_auto_clean_strategy` (
`f_index` bigint(20) NOT NULL AUTO_INCREMENT,       -- 唯一自增id
`f_strategy_id` char(40) NOT NULL,                  -- 使用类似于生成t_user表中f_user_id的方法生成
`f_obj_id` char(40) NOT NULL,                       -- 使用该策略的用户/部门/角色(6.0)id
`f_obj_type` tinyint(4) NOT NULL,                   -- 使用该策略的id的类型，用户1/部门2/角色4(6.0)
`f_enable_remain_hours` tinyint(4) NOT NULL,        -- 启用数据保留时间
`f_remain_hours` bigint(20) NOT NULL,               -- 数据在正常位置的保留时间
`f_clean_cycle_days` bigint(20) NOT NULL,           -- 清理周期
`f_clean_cycle_modify_time` bigint(20) NOT NULL,    -- 清理周期的变更时间
`f_create_time` bigint(20) NOT NULL,                -- 策略的创建时间
`f_status` tinyint(4) NOT NULL,                     -- 策略的启用/禁用标志位
PRIMARY KEY (`f_index`),
UNIQUE KEY `f_obj_id` (`f_obj_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_local_sync_strategy` (
    `f_index` int(11) NOT NULL AUTO_INCREMENT,                      -- 唯一自增id
    `f_strategy_id` char(40) NOT NULL,                              -- 策略id
    `f_obj_id` char(40) NOT NULL,                                   -- 对象id
    `f_obj_type` tinyint(4) NOT NULL,                               -- 对象类型, 1: 用户, 2: 部门
    `f_open_status` tinyint(4) NOT NULL,                            -- 本地同步策略开启状态
    `f_delete_status` tinyint(4) NOT NULL,                          -- 是否允许删除配置的同步任务
    `f_create_time` bigint(20) NOT NULL,                            -- 策略创建时间, 微秒的时间戳
    PRIMARY KEY (`f_index`),
    UNIQUE KEY `f_obj_id` (`f_obj_id`),
    KEY `f_strategy_id_index` (`f_strategy_id`),
    KEY `f_create_time_index` (`f_create_time`)
) ENGINE=InnoDB AUTO_INCREMENT=41;

CREATE TABLE IF NOT EXISTS `t_user_custom_attr` (
    `f_id` char(26) NOT NULL COMMENT '自定义属性id(主键)',
    `f_user_id` char(36) NOT NULL COMMENT '用户id(唯一索引)',
    `f_custom_attr` longtext NOT NULL COMMENT '自定义属性(json)',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `uk_user_id_index` (`f_user_id`) USING BTREE
) ENGINE=InnoDB COMMENT '用户自定义属性表';

INSERT INTO `t_sharemgnt_config`(`f_key`, `f_value`) SELECT 'reserved_name_lock', 'locked' FROM DUAL WHERE NOT EXISTS (SELECT `f_key` FROM `t_sharemgnt_config` WHERE `f_key` = 'reserved_name_lock');

CREATE TABLE IF NOT EXISTS `t_distributed_lock` (
    `f_lock_key` varchar(26) NOT NULL COMMENT '锁key',
    `f_hoder_id` char(36) NOT NULL COMMENT '锁holder id',
    `f_expire_time` datetime COMMENT '锁过期时间',
    PRIMARY KEY (`f_lock_key`)
) ENGINE=InnoDB COMMENT '分布式锁表';
-- PolicyManagement

/*
MySQL: Database - policy_mgnt
*********************************************************************
*/
use policy_mgnt;

CREATE TABLE IF NOT EXISTS `t_policies` (
  `f_name` varchar(255) NOT NULL,
  `f_default` text NOT NULL,
  `f_value` text NOT NULL,
  `f_locked` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`f_name`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_network_restriction` (
  `f_id` varchar(36) NOT NULL,
  `f_name` varchar(128) DEFAULT NULL,
  `f_start_ip` varchar(40) NOT NULL,
  `f_end_ip` varchar(40) NOT NULL,
  `f_ip_address` varchar(40) NOT NULL,
  `f_ip_mask` varchar(15) NOT NULL,
  `f_segment_start` varchar(128) NOT NULL,
  `f_segment_end` varchar(128) NOT NULL,
  `f_type` varchar(15) NOT NULL,
  `f_ip_type` varchar(15) NOT NULL DEFAULT 'ipv4',
  `f_created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `uix_t_network_restriction_f_name` (`f_name`),
  KEY `idx_t_network_restriction_f_start_ip` (`f_start_ip`),
  KEY `idx_t_network_restriction_f_end_ip` (`f_end_ip`),
  KEY `idx_t_network_restriction_f_ip_address` (`f_ip_address`),
  KEY `idx_t_network_restriction_f_type` (`f_type`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_network_accessor_relation` (
  `f_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `f_network_id` varchar(36) NOT NULL,
  `f_accessor_id` varchar(36) NOT NULL,
  `f_accessor_type` varchar(10) NOT NULL,
  `f_created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_net_acc` (`f_network_id`,`f_accessor_id`),
  KEY `idx_t_network_accessor_relation_f_accessor_type` (`f_accessor_type`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_event_store` (
  `f_id` bigint(20) NOT NULL,
  `f_dispatched` tinyint(1) NOT NULL DEFAULT 0,
  `f_dispatched_at` datetime DEFAULT NULL,
  `f_payload` longblob NOT NULL,
  `f_options` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
  `f_headers` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
  PRIMARY KEY (`f_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_product_relation` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',                     
    `f_account_id` varchar(40) NOT NULL COMMENT '授权对象id',
    `f_account_type` tinyint(4) NOT NULL COMMENT '授权类型，0：未知，1：普通用户',                              
    `f_product` varchar(255) NOT NULL COMMENT '产品名称，由license规定',                          
    PRIMARY KEY (`f_primary_id`) COMMENT '主键',
    KEY `idx_t_product_relation_f_account_id_f_product` (`f_account_id`, `f_product`),
    KEY `idx_t_product_relation_f_product_f_account_id` (`f_product`, `f_account_id`)
) ENGINE=InnoDB COMMENT='产品授权关系表';

CREATE TABLE IF NOT EXISTS `t_outbox` (
    `f_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `f_business_type` tinyint(4) NOT NULL COMMENT '业务类型',
    `f_message` longtext NOT NULL COMMENT '消息内容，json格式字符串',
    `f_create_time` bigint(20) NOT NULL COMMENT '消息创建时间',
    PRIMARY KEY (`f_id`),
    KEY `idx_business_type_and_create_time` (`f_business_type`, `f_create_time`)
) ENGINE=InnoDB COMMENT='outbox信息表';

CREATE TABLE IF NOT EXISTS `t_outbox_lock` (
    `f_business_type` tinyint(4) NOT NULL COMMENT '业务类型',
    PRIMARY KEY (`f_business_type`)
) ENGINE=InnoDB COMMENT='outbox分布式锁表';

INSERT INTO t_outbox_lock(f_business_type) SELECT 1 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 1);

-- ShareServer

/*
MySQL: Database - anyshare-eacp
*********************************************************************
*/
use anyshare;

CREATE TABLE IF NOT EXISTS `t_acs_custom_perm` (                                              -- 此表记录文档权限
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,                                          -- 自增主键
    `f_doc_id` text NOT NULL,                                                                   -- 文档路径
    `f_accessor_id` char(40) NOT NULL,                                                          -- 被配置权限的对象id
    `f_accessor_name` varchar(150) NOT NULL,                                                    -- 被配置权限的显示名
    `f_accessor_type` tinyint(4) NOT NULL,                                                      -- 权限所有者类型, 1: 用户, 2: 组织/部门, 3: 联系人组, 4: 匿名用户
    `f_type` tinyint(4) NOT NULL,                                                               -- 权限类型, 1: 拒绝, 2: 允许 3: 禁用继承
    `f_perm_value` int(11) NOT NULL DEFAULT '1',                                                -- 权限值
    `f_source` tinyint(4) NOT NULL DEFAULT '1',                                                 -- 权限来源  1: 用户配置, 2: 系统内部配置
    `f_end_time` bigint(20) DEFAULT '-1',                                                       -- 权限结束时间, 微秒的时间戳, -1标识永久有效
    `f_modify_time` bigint(20) NOT NULL DEFAULT '0',                                            -- 记录修改时间, 微秒的时间戳
    `f_create_time` bigint(20) NOT NULL,                                                        -- 记录创建时间, 微秒的时间戳
    PRIMARY KEY (`f_primary_id`),
    KEY `t_perm_f_doc_id_index` (`f_doc_id`(120)),
    KEY `t_perm_f_accessor_id_index` (`f_accessor_id`),
    KEY `t_perm_f_accessor_type_index` (`f_accessor_type`),
    KEY `t_perm_f_type_index` (`f_type`),
    KEY `t_perm_f_end_time_index` (`f_end_time`) USING BTREE
  ) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_acs_doc` (                        -- 此表保存文档入口视图
    `f_doc_id` char(40) NOT NULL,                                 -- 文档入口id
    `f_doc_type` tinyint(4) NOT NULL,                             -- 文档入口类型, 1: 个人文档, 3: 文档库
    `f_type_name` char(128) NOT NULL,                             -- 文档入口类型名
    `f_status` int(11) DEFAULT '1',                               -- 文档入口状态, 1: 启用, 其他: 禁用
    `f_create_time` bigint(20) NOT NULL DEFAULT '0',              -- 文档入口创建时间, 微秒的时间戳
    `f_delete_time` bigint(20) DEFAULT '0',                       -- 文档入口被删除的时间
    `f_deleter_id` char(40) NOT NULL DEFAULT '',                  -- 文档入口删除者id
    `f_obj_id` char(40) NOT NULL,                                 -- 文档入口标识id
    `f_name` char(128) NOT NULL,                                  -- 文档入口名
    `f_creater_id` char(40) NOT NULL,                             -- 文档入口的创建者id
    `f_creater_name` varchar(150) NOT NULL,                       -- 文档入口的创建者名称
    `f_creater_type` tinyint(4) NOT NULL,                         -- 文档入口的创建者类型
    `f_oss_id` char(150) NOT NULL DEFAULT '',                     -- 文档入口所属对象存储id
    `f_relate_depart_id` char(40) NOT NULL DEFAULT '',            -- 关联部门id
    `f_subtype_id` char(40) NOT NULL DEFAULT '',                  -- 所属文档库分类id
    `f_display_order` MEDIUMINT DEFAULT -1,                       -- 自定义文档库显示顺序
    `f_owners_id` text NOT NULL,                                  -- 文档库所有者id
    `f_owners_name` text NOT NULL,                                -- 文档库所有者名称
    `f_depart_manager_as_owner` tinyint(4) NOT NULL DEFAULT 0,    -- 部门负责人为所有者  0：否   1：是
    PRIMARY KEY (`f_doc_id`),
    KEY `t_doc_f_doc_type_index` (`f_doc_type`) USING BTREE,
    KEY `t_doc_f_obj_id_index` (`f_obj_id`) USING BTREE,
    KEY `t_doc_f_name_index` (`f_name`) USING BTREE,
    KEY `t_doc_f_type_name_index` (`f_type_name`) USING BTREE,
    KEY `t_doc_f_relate_depart_id_index` (`f_relate_depart_id`) USING BTREE,
    KEY `t_display_order_index` (`f_display_order`) USING BTREE,
    KEY `t_doc_f_creater_id_index` (`f_creater_id`) USING BTREE
  ) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_acs_doc_quit` (                   -- 此表记录屏蔽共享信息
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,            -- 自增主键
    `f_user_id` char(40) NOT NULL,                                -- 用户id
    `f_doc_id` char(40) NOT NULL,                                 -- 入口文档id
    PRIMARY KEY (`f_primary_id`),
    UNIQUE KEY `t_acs_doc_unique_index` (`f_user_id`,`f_doc_id`) USING HASH,
    KEY `t_acs_doc_quit_f_doc_id_index` (`f_doc_id`) USING BTREE
  ) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_acs_owner` (                                                      -- 此表保存文档入口所有者信息
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,                                            -- 自增主键
    `f_gns_path` text NOT NULL,                                                                   -- 文档路径
    `f_owner_id` char(40) NOT NULL,                                                               -- 所有者id
    `f_owner_name` varchar(150) NOT NULL,                                                         -- 所有者显示名
    `f_type` tinyint(4) NOT NULL,                                                                 -- 用户类型, 1: 用户, 2: 组织/部门, 3: 联系人组, 4: 匿名用户, 5: 用户组， 6: 应用账户
    `f_modify_time` bigint(20) NOT NULL DEFAULT '0',                                              -- 记录修改时间, 微秒的时间戳
    `f_deletable` tinyint(1) NOT NULL,                                                            -- 允许删除标记, 1: 允许删除, 0: 禁止删除
    PRIMARY KEY (`f_primary_id`),
    KEY `t_owner_f_gns_path_index` (`f_gns_path`(120)),
    KEY `t_owner_f_owner_id_index` (`f_owner_id`)
  ) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_device` (                         -- 此表记录设备信息
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,            -- 自增主键
    `f_user_id` char(40) NOT NULL,                                -- 用户id
    `f_udid` char(40) NOT NULL,                                   -- 用户设备标识
    `f_name` char(128) NOT NULL,                                  -- 设备名
    `f_os_type` tinyint(4) NOT NULL,                              -- 系统类型
    `f_device_type` char(128) NOT NULL,                           -- 设备类型
    `f_last_login_ip` char(40) NOT NULL,                          -- 最后登录IP
    `f_last_login_time` bigint(20) NOT NULL,                      -- 最后登录时间, 微秒的时间戳
    `f_erase_flag` tinyint(4) NOT NULL DEFAULT '0',               -- 设备擦除标记
    `f_last_erase_time` bigint(20) NOT NULL DEFAULT '0',          -- 最后擦除时间, 微秒的时间戳
    `f_disable_flag` tinyint(4) NOT NULL DEFAULT '0',             -- 设备禁用标记
    `f_bind_flag` tinyint(4) NOT NULL DEFAULT '0',                -- 设备绑定标记
    PRIMARY KEY (`f_primary_id`),
    UNIQUE KEY `unique_index` (`f_user_id`,`f_udid`) USING BTREE,
    KEY `f_udid_index` (`f_udid`) USING BTREE
  ) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_lock` (                           -- 此表记录锁定的文件
    `f_primary_id` char(40) NOT NULL,                             -- 自增主键
    `f_doc_id` text NOT NULL,                                     -- 文件路径标识
    `f_user_id` char(40) NOT NULL,                                -- 用户id
    `f_user_name` char(150) NOT NULL,                             -- 用户显示名
    `f_user_type` tinyint(4) NOT NULL,                            -- 用户类型  1: 用户, 6: 应用账户
    `f_source` tinyint(4) NOT NULL DEFAULT '1',                   -- 锁记录来源  1: 用户配置, 2: 系统内部配置
    `f_create_date` bigint(20) NOT NULL DEFAULT '-1',             -- 锁创建时间
    `f_refresh_date` bigint(20) NOT NULL DEFAULT '-1',            -- 锁刷新时间
    `f_expire_time` bigint(20) NOT NULL DEFAULT '-2',             -- 锁过期时间, -1: 永久有效, -2: 服务器配置的超期间隔(单位: 秒)
    PRIMARY KEY (`f_primary_id`),
    KEY `t_finder_f_doc_id_index` (`f_doc_id`(120)) USING BTREE,
    KEY `t_lock_f_refresh_date_index` (`f_refresh_date`) USING BTREE,
    KEY `t_lock_f_expire_time_index` (`f_expire_time`) USING BTREE
  ) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_audit` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT  COMMENT '自增主键',
    `f_apply_id` char(40) NOT NULL COMMENT '申请唯一标识',
    `f_apply_type` tinyint(4) NOT NULL COMMENT '申请的类型',
    `f_doc_id` text NOT NULL COMMENT '文档路径标识',
    `f_sharer_id` char(40) NOT NULL COMMENT '申请者id',
    `f_create_date` bigint(20) NOT NULL COMMENT '申请创建时间, 微秒的时间戳',
    `f_accessor_id` char(40) NOT NULL COMMENT '被共享者id',
    `f_accessor_name` char(150) NOT NULL COMMENT '被共享者名字',
    `f_accessor_type` tinyint(4) NOT NULL COMMENT '被共享者类型',
    `f_detail` text NOT NULL COMMENT '申请详情',
    PRIMARY KEY (`f_primary_id`),
    UNIQUE KEY `uk_apply_id` (`f_apply_id`),
    KEY `idx_doc_id` (`f_doc_id`(137)) COMMENT '大小依据 4层目录+分隔符+gns前缀',
    KEY `idx_sharer_id` (`f_sharer_id`),
    KEY `idx_accessor_id` (`f_accessor_id`)
  ) ENGINE=InnoDB COMMENT='审核申请信息表';

CREATE TABLE IF NOT EXISTS `t_active_user_info` (               -- 此表记录活跃用户信息
    `f_id` bigint(20) NOT NULL AUTO_INCREMENT,                    -- 自增主键
    `f_time` char(40) NOT NULL,                                   -- 统计时间, 如: 2018-08-08
    `f_user_id` char(40) NOT NULL,                                -- 用户id
    PRIMARY KEY (`f_id`),
    KEY `idx_userid` (`f_user_id`),
    KEY `idx_time` (`f_time`)
  ) ENGINE=InnoDB;

  CREATE TABLE IF NOT EXISTS `t_conf` (                           -- 此表记录基本配置信息
    `f_key` char(32) NOT NULL,                                    -- 配置关键字
    `f_value` char(255) NOT NULL,                                 -- 配置的值
    PRIMARY KEY (`f_key`)
  ) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_eacp_outbox` (
    `f_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `f_message` longtext NOT NULL COMMENT '消息内容，json格式字符串',
    `f_create_time` bigint(20) NOT NULL COMMENT '消息创建时间',
    PRIMARY KEY (`f_id`),
    KEY `idx_create_time` (`f_create_time`)
  ) ENGINE=InnoDB COMMENT='outbox信息表';

INSERT INTO t_conf(f_key,f_value) SELECT 'auto_lock','true' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'auto_lock');
INSERT INTO t_conf(f_key,f_value) SELECT 'oem_allow_auth_low_csf_user','true' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key ='oem_allow_auth_low_csf_user');
INSERT INTO t_conf(f_key,f_value) SELECT 'oem_allow_owner','true' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'oem_allow_owner');
INSERT INTO t_conf(f_key,f_value) SELECT 'oem_client_logout_time','-1' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'oem_client_logout_time');
INSERT INTO t_conf(f_key,f_value) SELECT 'oem_indefinite_perm','true' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'oem_indefinite_perm');
INSERT INTO t_conf(f_key,f_value) SELECT 'oem_max_pass_expired_days','-1' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'oem_max_pass_expired_days');
INSERT INTO t_conf(f_key,f_value) SELECT 'oem_remember_pass','true' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'oem_remember_pass');
INSERT INTO t_conf(f_key,f_value) SELECT 'web_client_host','' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'web_client_host');
INSERT INTO t_conf(f_key,f_value) SELECT 'web_client_port','443' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'web_client_port');
INSERT INTO t_conf(f_key,f_value) SELECT 'web_client_http_port','80' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key ='web_client_http_port');
INSERT INTO t_conf(f_key,f_value) SELECT 'eacp_https_port','9999' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'eacp_https_port');
INSERT INTO t_conf(f_key,f_value) SELECT 'efast_https_port','9124' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'efast_https_port');
INSERT INTO t_conf(f_key,f_value) SELECT 'oem_default_perm_expired_days','-1' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'oem_default_perm_expired_days');
INSERT INTO t_conf(f_key,f_value) SELECT 'auto_lock_expired_interval','180' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'auto_lock_expired_interval');

-- Authentication

/*
MySQL: Database - authentication
*********************************************************************
*/

use authentication;

CREATE TABLE IF NOT EXISTS `t_session` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_login_session_id` varchar(255) NOT NULL COMMENT 'session id',
    `f_subject` varchar(255) NOT NULL COMMENT '用户ID',
    `f_client_id` varchar(255) NOT NULL COMMENT '客户端ID',
    `f_exp` bigint(20) NOT NULL COMMENT 'Context到期时间戳',
    `f_session_access_token` text NOT NULL COMMENT 'Context信息',
    PRIMARY KEY (`f_primary_id`),
    UNIQUE KEY (`f_login_session_id`)
) ENGINE=InnoDB COMMENT='Context信息表';

CREATE TABLE IF NOT EXISTS `t_client_public` (
`id` varchar(255) NOT NULL COMMENT '客户端ID',
`client_name` text NOT NULL COMMENT '客户端名称',
`client_secret` text NOT NULL COMMENT '客户端密钥',
`redirect_uris` text NOT NULL COMMENT '客户端回调地址',
`grant_types` text NOT NULL COMMENT '客户端授权模式',
`response_types` text NOT NULL COMMENT '客户端接收响应类型',
`scope` text NOT NULL COMMENT '客户端申请权限范围',
`pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
`post_logout_redirect_uris` text NOT NULL COMMENT '客户端登出成功地址',
`metadata` text NOT NULL COMMENT '客户端元数据',
`created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '客户端创建时间',
PRIMARY KEY (`pk`),
UNIQUE KEY `hydra_client_idx_id_uq` (`id`)
) ENGINE=InnoDB COMMENT='公开注册客户端信息表';

CREATE TABLE IF NOT EXISTS `t_conf` (
  `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `f_key` char(32) NOT NULL COMMENT '键',
  `f_value` varchar(1024) NOT NULL COMMENT '值',
  PRIMARY KEY (`f_primary_id`),
  UNIQUE KEY `uk_conf` (`f_key`)
) ENGINE=InnoDB COMMENT='认证配置表';

INSERT INTO t_conf(f_key,f_value) SELECT 'remember_for','2592000' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'remember_for');
INSERT INTO t_conf(f_key,f_value) SELECT 'remember_visible','true' FROM DUAL WHERE NOT EXISTS(SELECT f_value FROM t_conf WHERE f_key = 'remember_visible');

CREATE TABLE IF NOT EXISTS `t_access_token_perm` (
  `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `f_app_id` char(36) NOT NULL COMMENT '应用账户id',
  `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`f_primary_id`),
  KEY `idx_f_app_id` (`f_app_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_outbox` (
    `f_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `f_business_type` tinyint(4) NOT NULL COMMENT '业务类型',
    `f_message` longtext NOT NULL COMMENT '消息内容，json格式字符串',
    `f_create_time` bigint(20) NOT NULL COMMENT '消息创建时间',
    PRIMARY KEY (`f_id`),
    KEY `idx_business_type_and_create_time` (`f_business_type`, `f_create_time`)
) ENGINE=InnoDB COMMENT='outbox信息表';

CREATE TABLE IF NOT EXISTS `t_outbox_lock` (
    `f_business_type` tinyint(4) NOT NULL COMMENT '业务类型',
    PRIMARY KEY (`f_business_type`)
) ENGINE=InnoDB COMMENT='outbox分布式锁表';

CREATE TABLE IF NOT EXISTS t_anonymous_sms_vcode (
    f_id char(26) NOT NULL COMMENT '验证码唯一标识',
    f_phone_number varchar(150) NOT NULL COMMENT '加密手机号',
    f_anonymity_id char(40) NOT NULL COMMENT '匿名账户id',
    f_content char(8) NOT NULL COMMENT '验证码内容',
    f_create_time timestamp NOT NULL COMMENT '创建时间',
    PRIMARY KEY (f_id),
    KEY idx_phone_number_anonymity_id (f_phone_number, f_anonymity_id),
    KEY idx_create_time (f_create_time)
) ENGINE=InnoDB COMMENT='匿名认证短信验证码存储表';

CREATE TABLE IF NOT EXISTS `t_distributed_lock` (
    `f_business_type` tinyint(4) NOT NULL COMMENT '业务类型',
    PRIMARY KEY (`f_business_type`)
) ENGINE=InnoDB COMMENT='分布式锁表';

CREATE TABLE IF NOT EXISTS `t_ticket` (
    `f_id` char(26) NOT NULL COMMENT '凭据唯一标识',
    `f_user_id` char(40) NOT NULL COMMENT '用户唯一标识',
    `f_client_id` varchar(255) NOT NULL COMMENT 'OAuth2客户端唯一标识',
    `f_create_time` bigint(10) NOT NULL COMMENT '凭据创建时间',
    PRIMARY KEY (`f_id`),
    KEY `idx_create_time` (`f_create_time`)
) ENGINE=InnoDB COMMENT='单点登录凭据表';

INSERT INTO t_outbox_lock(f_business_type) SELECT 1 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 1);
INSERT INTO t_outbox_lock(f_business_type) SELECT 2 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 2);
INSERT INTO t_outbox_lock(f_business_type) SELECT 3 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 3);
INSERT INTO t_outbox_lock(f_business_type) SELECT 4 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_outbox_lock WHERE f_business_type = 4);

INSERT INTO t_distributed_lock(f_business_type) SELECT 1 FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_distributed_lock WHERE f_business_type = 1);

INSERT INTO t_conf(f_key, f_value) SELECT 'anonymous_sms_expiration', '2' FROM DUAL WHERE NOT EXISTS(SELECT f_key FROM t_conf WHERE f_key = 'anonymous_sms_expiration');

CREATE TABLE IF NOT EXISTS t_outbox_unordered (
  id bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  f_message text NOT NULL COMMENT '信息',
  f_status tinyint(11) NOT NULL DEFAULT 0 COMMENT '状态(0 未开始,1 处理中)',
  f_created_at bigint(20) NOT NULL COMMENT '创建时间',
  f_updated_at bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_f_status(f_status),
  KEY idx_f_created_at(f_created_at),
  KEY idx_f_updated_at(f_updated_at)
)ENGINE = InnoDB COMMENT='无序outbox信息表';

-- AuditLog

/*
MySQL: Database - anyshare
*********************************************************************
*/
use anyshare;

CREATE TABLE IF NOT EXISTS `t_log_login` (
  `f_log_id` bigint(20) NOT NULL COMMENT '日志id',
  `f_user_id` char(40) NOT NULL COMMENT '用户id',
  `f_user_name` char(128) NOT NULL COMMENT '用户显示名',
  `f_user_type` varchar(32) NOT NULL DEFAULT 'authenticated_user' COMMENT '用户类型',
  `f_obj_id` char(40) NOT NULL COMMENT '对象id',
  `f_additional_info` text NOT NULL COMMENT '附加信息',
  `f_level` tinyint(4) NOT NULL COMMENT '日志级别, 1: 信息, 2: 警告',
  `f_op_type` tinyint(4) NOT NULL COMMENT '操作类型',
  `f_date` bigint(20) NOT NULL COMMENT '日志记录时间, 微秒的时间戳',
  `f_ip` char(40) NOT NULL COMMENT '访问者的IP',
  `f_mac` char(40) NOT NULL DEFAULT '' COMMENT '文档入口属于哪个站点',
  `f_msg` text NOT NULL COMMENT '日志描述',
  `f_exmsg` text NOT NULL COMMENT '日志附加描述',
  `f_user_agent` varchar(1024) NOT NULL DEFAULT '' COMMENT '用户代理',
  `f_user_paths` text COMMENT '用户所属部门信息',
  `f_obj_name` char(128) NOT NULL DEFAULT '' COMMENT '对象名称',
  `f_obj_type` tinyint(4) NOT NULL DEFAULT 0 COMMENT '对象类型',
  PRIMARY KEY (`f_log_id`),
  KEY `t_log_f_user_id_index` (`f_user_id`) USING BTREE,
  KEY `t_log_f_user_name_index` (`f_user_name`) USING BTREE,
  KEY `t_log_f_op_type_index` (`f_op_type`) USING BTREE,
  KEY `t_log_f_date_index` (`f_date`) USING BTREE,
  KEY `t_log_f_ip_index` (`f_ip`) USING BTREE,
  KEY `t_log_f_mac_index` (`f_mac`) USING BTREE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_log_management` (
  `f_log_id` bigint(20) NOT NULL COMMENT '日志id',
  `f_user_id` char(40) NOT NULL COMMENT '用户id',
  `f_user_name` char(128) NOT NULL COMMENT '用户显示名',
  `f_user_type` varchar(32) NOT NULL DEFAULT 'authenticated_user' COMMENT '用户类型',
  `f_obj_id` char(40) NOT NULL COMMENT '对象id',
  `f_additional_info` text NOT NULL COMMENT '附加信息',
  `f_level` tinyint(4) NOT NULL COMMENT '日志级别, 1: 信息, 2: 警告',
  `f_op_type` tinyint(4) NOT NULL COMMENT '操作类型',
  `f_date` bigint(20) NOT NULL COMMENT '日志记录时间',
  `f_ip` char(40) NOT NULL COMMENT '访问者IP',
  `f_mac` char(40) NOT NULL DEFAULT '' COMMENT '文档入口所属站点',
  `f_msg` text NOT NULL COMMENT '日志描述',
  `f_exmsg` text NOT NULL COMMENT '日志附加描述',
  `f_user_agent` varchar(1024) NOT NULL DEFAULT '' COMMENT '用户代理',
  `f_user_paths` text COMMENT '用户所属部门信息',
  `f_obj_name` char(128) NOT NULL DEFAULT '' COMMENT '对象名称',
  `f_obj_type` tinyint(4) NOT NULL DEFAULT 0 COMMENT '对象类型',
  PRIMARY KEY (`f_log_id`),
  KEY `t_log_f_user_id_index` (`f_user_id`) USING BTREE,
  KEY `t_log_f_user_name_index` (`f_user_name`) USING BTREE,
  KEY `t_log_f_op_type_index` (`f_op_type`) USING BTREE,
  KEY `t_log_f_date_index` (`f_date`) USING BTREE,
  KEY `t_log_f_ip_index` (`f_ip`) USING BTREE,
  KEY `t_log_f_mac_index` (`f_mac`) USING BTREE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_log_operation` (
  `f_log_id` bigint(20) NOT NULL COMMENT '日志id',
  `f_user_id` char(40) NOT NULL COMMENT '用户id',
  `f_user_name` char(128) NOT NULL COMMENT '用户显示名',
  `f_user_type` varchar(32) NOT NULL DEFAULT 'authenticated_user' COMMENT '用户类型',
  `f_obj_id` char(40) NOT NULL COMMENT '对象id',
  `f_additional_info` text NOT NULL COMMENT '附加信息',
  `f_level` tinyint(4) NOT NULL COMMENT '日志级别, 1: 信息, 2: 警告',
  `f_op_type` tinyint(4) NOT NULL COMMENT '日志类型',
  `f_date` bigint(20) NOT NULL COMMENT '日志记录时间',
  `f_ip` char(40) NOT NULL COMMENT '访问者IP',
  `f_mac` char(40) NOT NULL DEFAULT '' COMMENT '文档入口所属站点',
  `f_msg` text NOT NULL COMMENT '日志描述',
  `f_exmsg` text NOT NULL COMMENT '日志附加描述',
  `f_user_agent` varchar(1024) NOT NULL DEFAULT '' COMMENT '用户代理',
  `f_user_paths` text COMMENT '用户所属部门信息',
  `f_obj_name` char(128) NOT NULL DEFAULT '' COMMENT '对象名称',
  `f_obj_type` tinyint(4) NOT NULL DEFAULT 0 COMMENT '对象类型',
  PRIMARY KEY (`f_log_id`),
  KEY `t_log_f_user_id_index` (`f_user_id`) USING BTREE,
  KEY `t_log_f_user_name_index` (`f_user_name`) USING BTREE,
  KEY `t_log_f_op_type_index` (`f_op_type`) USING BTREE,
  KEY `t_log_f_date_index` (`f_date`) USING BTREE,
  KEY `t_log_f_ip_index` (`f_ip`) USING BTREE,
  KEY `t_log_f_mac_index` (`f_mac`) USING BTREE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_history_log_info` (
  `f_id` char(128) NOT NULL COMMENT '唯一标识',
  `f_name` char(128) NOT NULL COMMENT '日志记录名',
  `f_size` bigint(20) NOT NULL COMMENT '日志大小',
  `f_type` tinyint(4) NOT NULL COMMENT '记录类型, 10: 登录日志, 11: 管理日志, 12: 操作日志',
  `f_date` bigint(20) NOT NULL COMMENT '记录时间',
  `f_dump_date` bigint(20) NOT NULL COMMENT '转存时间',
  `f_oss_id` char(40) NOT NULL COMMENT '历史日志所属的对象存储ID',
  PRIMARY KEY (`f_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `t_log_config` (
  `f_key` char(40) NOT NULL,
  `f_value` char(40) NOT NULL,
  PRIMARY KEY (`f_key`)
) ENGINE=InnoDB COMMENT='日志配置';

-- 转存周期
INSERT INTO t_log_config (f_key, f_value) SELECT 'retention_period', 1 FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM t_log_config WHERE f_key = 'retention_period');
-- 转存周期单位
INSERT INTO t_log_config (f_key, f_value) SELECT 'retention_period_unit', 'year' FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM t_log_config WHERE f_key = 'retention_period_unit');
-- 转存时间
INSERT INTO t_log_config (f_key, f_value) SELECT 'dump_time', '03:00:00' FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM t_log_config WHERE f_key = 'dump_time');
-- 转存格式
INSERT INTO t_log_config (f_key, f_value) SELECT 'dump_format', 'csv' FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM t_log_config WHERE f_key = 'dump_format');
-- 历史日志导出是否加密
INSERT INTO t_log_config (f_key, f_value) SELECT 'history_log_export_with_pwd', 0 FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM t_log_config WHERE f_key = 'history_log_export_with_pwd');

CREATE TABLE IF NOT EXISTS `t_auditlog_outbox` (
    `f_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `f_business_type` varchar(20) NOT NULL COMMENT '业务类型',
    `f_message` longtext NOT NULL COMMENT '消息内容，json格式字符串',
    `f_create_time` bigint(20) NOT NULL COMMENT '消息创建时间',
    PRIMARY KEY (`f_id`),
    KEY `idx_business_type_and_create_time` (`f_business_type`, `f_create_time`)
  ) ENGINE=InnoDB COMMENT='outbox信息表';

CREATE TABLE IF NOT EXISTS `t_auditlog_outbox_lock` (
    `f_business_type` varchar(20) NOT NULL COMMENT '业务类型',
    PRIMARY KEY (`f_business_type`)
) ENGINE=InnoDB COMMENT='outbox分布式锁表';

INSERT INTO t_auditlog_outbox_lock(f_business_type) SELECT 'client_log' FROM DUAL WHERE NOT EXISTS(SELECT f_business_type FROM t_auditlog_outbox_lock WHERE f_business_type = 'client_log');

CREATE TABLE IF NOT EXISTS `t_log_scope_strategy` (
  `f_id` bigint(20) NOT NULL,
  `f_created_at` bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  `f_created_by` varchar(64) NOT NULL DEFAULT '' COMMENT '创建人员',
  `f_updated_at` bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  `f_updated_by` varchar(64) NOT NULL DEFAULT '' COMMENT '更新人员',
  `f_log_type` tinyint(4) NOT NULL COMMENT '日志类型',
  `f_log_category` tinyint(4) NOT NULL COMMENT '日志分类',
  `f_role` char(128) NOT NULL COMMENT '查看者角色名',
  `f_scope` varchar(1024) NOT NULL COMMENT '查看范围',
  PRIMARY KEY (`f_id`),
  KEY `idx_log_type` (`f_log_type`),
  KEY `idx_log_category` (`f_log_category`),
  KEY `idx_role` (`f_role`)
) ENGINE=InnoDB COMMENT='日志查看范围策略';

-- 安全管理员查看活跃访问日志
INSERT INTO t_log_scope_strategy(f_id, f_log_type, f_log_category, f_role, f_scope)
SELECT 111000222000333000, 10, 1, 'sec_admin', 'audit_admin,normal_user'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM t_log_scope_strategy
  WHERE f_id = 111000222000333000 OR (f_log_type = 10 AND f_log_category = 1 AND f_role = 'sec_admin')
);
-- 审计管理员查看活跃访问日志
INSERT INTO t_log_scope_strategy(f_id, f_log_type, f_log_category, f_role, f_scope)
SELECT 111000222000333001, 10, 1, 'audit_admin', 'sys_admin,sec_admin'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM t_log_scope_strategy
  WHERE f_id = 111000222000333001 OR (f_log_type = 10 AND f_log_category = 1 AND f_role = 'audit_admin')
);
-- 安全管理员查看活跃管理日志
INSERT INTO t_log_scope_strategy(f_id, f_log_type, f_log_category, f_role, f_scope)
SELECT 111000222000333002, 11, 1, 'sec_admin', 'audit_admin,normal_user'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM t_log_scope_strategy
  WHERE f_id = 111000222000333002 OR (f_log_type = 11 AND f_log_category = 1 AND f_role = 'sec_admin')
);
-- 审计管理员查看活跃管理日志
INSERT INTO t_log_scope_strategy(f_id, f_log_type, f_log_category, f_role, f_scope)
SELECT 111000222000333003, 11, 1, 'audit_admin', 'sys_admin,sec_admin'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM t_log_scope_strategy
  WHERE f_id = 111000222000333003 OR (f_log_type = 11 AND f_log_category = 1 AND f_role = 'audit_admin')
);
-- 安全管理员查看活跃操作日志
INSERT INTO t_log_scope_strategy(f_id, f_log_type, f_log_category, f_role, f_scope)
SELECT 111000222000333004, 12, 1, 'sec_admin', 'normal_user'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM t_log_scope_strategy
  WHERE f_id = 111000222000333004 OR (f_log_type = 12 AND f_log_category = 1 AND f_role = 'audit_admin')
);
-- 安全管理员查看历史访问日志
INSERT INTO t_log_scope_strategy(f_id, f_log_type, f_log_category, f_role, f_scope)
SELECT 111000222000333005, 10, 2, 'sec_admin', ''
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM t_log_scope_strategy
  WHERE f_id = 111000222000333005 OR (f_log_type= 10 AND f_log_category = 2 AND f_role = 'sec_admin')
);
-- 安全管理员查看历史管理日志
INSERT INTO t_log_scope_strategy(f_id, f_log_type, f_log_category, f_role, f_scope)
SELECT 111000222000333006, 11, 2, 'sec_admin', ''
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM t_log_scope_strategy
  WHERE f_id = 111000222000333006 OR (f_log_type= 11 AND f_log_category = 2 AND f_role = 'sec_admin')
);
-- 审计管理员查看历史操作日志
INSERT INTO t_log_scope_strategy(f_id, f_log_type, f_log_category, f_role, f_scope)
SELECT 111000222000333007, 12, 2, 'sec_admin', ''
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM t_log_scope_strategy
  WHERE f_id = 111000222000333007 OR (f_log_type= 12 AND f_log_category = 2 AND f_role = 'sec_admin')
);

-- 暂时只用于redis分布式锁的value，保证value的唯一性
-- 【注意】这个和Personalization共用一张表，如果调整，两边都注意下是否一起调整相应地方
create table if not exists t_pers_rec_unique_id
(
    f_id        char(36) not null comment 'ulid生成的id',
    f_flag tinyint not null  comment '使用场景（1：数据库的主键，2：redis分布式锁value）',
    primary key (f_id, f_flag)
) ENGINE = InnoDB comment '个性化推荐 唯一id';

CREATE TABLE IF NOT EXISTS t_pers_rec_svc_config
(
    f_id         bigint        not null auto_increment,
    f_key        varchar(64)   not null comment '配置key',
    f_value      varchar(2048) not null comment '配置value',
    f_created_at bigint        not null comment '创建时间',
    f_updated_at bigint        not null default 0 comment '更新时间',
    primary key (f_id),
    unique key uk_key (f_key)
) ENGINE = InnoDB COMMENT '个性化推荐 服务配置（用于存储一些配置或标识等）';

-- Authorization

/*
MySQL: Database - anyshare
*********************************************************************
*/
use anyshare;

CREATE TABLE IF NOT EXISTS `t_resource_type`
(
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_id`  char(40) NOT NULL COMMENT '资源类型唯一标识',
    `f_name`  char(255) NOT NULL COMMENT '资源类型名称',
    `f_description`  text NOT NULL COMMENT '资源类型描述',
    `f_instance_url`  text NOT NULL COMMENT '资源类型实例URL',
    `f_data_struct`  char(40) NOT NULL COMMENT '数据结构, 支持tree、array、string',
    `f_operation`   longtext NOT NULL COMMENT '操作, 内容是json数组',
    `f_hidden` tinyint(4) NOT NULL COMMENT '是否隐藏, 0: 不隐藏, 1: 隐藏',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_modify_time` bigint(20) NOT NULL COMMENT '修改时间',
    UNIQUE KEY `uk_id` (`f_id`),
    PRIMARY KEY (`f_primary_id`)
) ENGINE = InnoDB COMMENT='资源类型表';


CREATE TABLE IF NOT EXISTS `t_policy`
(
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_id` char(40) NOT NULL COMMENT '策略ID',
    `f_resource_id`  char(40) NOT NULL COMMENT '资源实例ID',
    `f_resource_type` char(40) NOT NULL COMMENT '资源类型',
    `f_resource_name`  char(255) NOT NULL COMMENT '资源名称',
    `f_accessor_id`  char(40) NOT NULL COMMENT '访问者',
    `f_accessor_type`  tinyint(4) NOT NULL COMMENT '访问者类型 1: 用户, 2: 组织/部门,  5: 用户组， 6: 应用账户 7: 角色' ,
    `f_accessor_name`  varchar(150) NOT NULL COMMENT '访问者名称',
    `f_operation`     longtext   NOT NULL COMMENT '操作',
    `f_condition`     longtext   NOT NULL COMMENT '条件',
    `f_ancestors`     longtext   NOT NULL COMMENT '祖先信息',
    `f_end_time` bigint(20) NOT NULL COMMENT '过期时间',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_modify_time` bigint(20) NOT NULL COMMENT '修改时间',
    KEY `idx_f_id` (`f_id`),
    KEY `idx_accessor_type` (`f_accessor_id`, `f_accessor_type`),
    KEY `idx_resource_type_accessor` (`f_resource_type`, `f_accessor_id`),
    KEY `idx_resource_id_type_accessor` (`f_resource_id`, `f_resource_type`, `f_accessor_id`),
    PRIMARY KEY (`f_primary_id`)
) ENGINE = InnoDB COMMENT ='策略配置表';


CREATE TABLE IF NOT EXISTS `t_role` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_id` char(40) NOT NULL COMMENT '角色唯一标识',
    `f_name` varchar(512) NOT NULL COMMENT '角色名称',
    `f_description` text NOT NULL COMMENT '描述',
    `f_source` tinyint(4) NOT NULL DEFAULT 3 COMMENT '角色来源, 1: 系统, 2: 业务内置, 3: 用户自定义',
    `f_visibility` tinyint(4) NOT NULL COMMENT '是否可见, 0: 不可见, 1: 可见',
    `f_resource_scope` longtext   NOT NULL COMMENT '资源范围',
    `f_created_time` bigint(40) NOT NULL COMMENT '创建时间',
    `f_modify_time`  bigint(20) NOT NULL COMMENT '修改时间',
    KEY `idx_t_role_f_id` (`f_id`),
    PRIMARY KEY (`f_primary_id`)
) ENGINE=InnoDB COMMENT='角色表';


CREATE TABLE IF NOT EXISTS `t_role_member` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_role_id` char(40) NOT NULL COMMENT '角色唯一标识',
    `f_member_id` char(40) NOT NULL COMMENT '成员唯一标识',
    `f_member_type` tinyint(4) NOT NULL COMMENT '成员类型,1: 用户, 2: 组织/部门, 5: 用户组 6: 应用账户',
    `f_member_name` varchar(150) NOT NULL COMMENT '成员名称',
    `f_created_time` bigint(40) NOT NULL COMMENT '创建时间',
    `f_modify_time`  bigint(20) NOT NULL COMMENT '修改时间',
    KEY `idx_f_role_id` (`f_role_id`),
    KEY `idx_f_member_id` (`f_member_id`),
    PRIMARY KEY (`f_primary_id`)
) ENGINE=InnoDB COMMENT='角色成员表';

CREATE TABLE IF NOT EXISTS `t_obligation_type` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_id`  char(255) NOT NULL COMMENT '义务类型唯一标识',
    `f_name`  char(255) NOT NULL COMMENT '义务类型名称',
    `f_description`  text NOT NULL COMMENT '义务类型描述',
    `f_applicable_resource_types`   longtext NOT NULL COMMENT '资源类型范围, 格式是json',
    `f_schema`   longtext NOT NULL COMMENT '参数配置，格式为JSON Schema',
    `f_ui_schema`   longtext NOT NULL COMMENT 'uiSchema, 格式是json',
    `f_default_value`   longtext NOT NULL COMMENT '义务类型默认值, 格式是json',
    `f_created_at` bigint(40) NOT NULL COMMENT '创建时间',
    `f_modified_at`  bigint(20) NOT NULL COMMENT '修改时间',
    KEY `idx_f_id` (`f_id`),
    PRIMARY KEY (`f_primary_id`)
) ENGINE = InnoDB COMMENT='义务类型表';

CREATE TABLE IF NOT EXISTS `t_obligation` (
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_id`  char(40) NOT NULL COMMENT '义务唯一标识',
    `f_type_id`  char(255) NOT NULL COMMENT '义务类型',
    `f_name`  varchar(255) NOT NULL COMMENT '义务名称',
    `f_description`  text NOT NULL COMMENT '义务描述',
    `f_value`   longtext NOT NULL COMMENT '义务配置, 格式是json',
    `f_created_at` bigint(40) NOT NULL COMMENT '创建时间',
    `f_modified_at`  bigint(20) NOT NULL COMMENT '修改时间',
    KEY `idx_f_id` (`f_id`),
    KEY `idx_f_type_id` (`f_type_id`),
    PRIMARY KEY (`f_primary_id`)
) ENGINE = InnoDB COMMENT='义务表';


CREATE TABLE IF NOT EXISTS `t_resource_type_hierarchy`
(
    `f_primary_id` bigint(20) NOT NULL AUTO_INCREMENT,
    `f_resource_type_id`  char(40) NOT NULL COMMENT '根节点的资源类型唯一标识',
    `f_children`          longtext NOT NULL COMMENT '下级节点信息',
    `f_created_at` bigint(20) NOT NULL COMMENT '创建时间',
    `f_modified_at` bigint(20) NOT NULL COMMENT '修改时间',
    UNIQUE KEY `uk_resource_type_id` (`f_resource_type_id`),
    PRIMARY KEY (`f_primary_id`)
) ENGINE = InnoDB COMMENT='资源类型层级关系表';