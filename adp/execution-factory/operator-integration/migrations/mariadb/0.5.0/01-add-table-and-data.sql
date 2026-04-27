USE adp;

CREATE TABLE IF NOT EXISTS `t_skill_repository` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_skill_id` varchar(40) NOT NULL COMMENT 'Skill ID',
    `f_name` varchar(255) NOT NULL COMMENT 'Skill 名称',
    `f_description` longtext NOT NULL COMMENT 'Skill 描述',
    `f_skill_content` longtext NOT NULL COMMENT 'Skill 指令正文',
    `f_version` varchar(40) NOT NULL COMMENT 'Skill 版本',
    `f_status` varchar(40) NOT NULL COMMENT 'Skill 状态',
    `f_source` varchar(50) NOT NULL DEFAULT '' COMMENT 'Skill 来源',
    `f_extend_info` longtext DEFAULT NULL COMMENT '扩展信息',
    `f_dependencies` longtext DEFAULT NULL COMMENT '依赖信息',
    `f_file_manifest` longtext DEFAULT NULL COMMENT '文件摘要清单',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    `f_delete_user` varchar(50) NOT NULL DEFAULT '' COMMENT '删除者',
    `f_delete_time` bigint(20) NOT NULL DEFAULT 0 COMMENT '删除时间',
    `f_category` varchar(50) DEFAULT '' COMMENT '工具箱分类, 数据处理/算法模型',
    `f_is_deleted` boolean DEFAULT 0 COMMENT '是否删除', -- 0: 未删除, 1: 待删除
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `uk_skill_id` (`f_skill_id`) USING BTREE,
    KEY `idx_status_update_time` (`f_status`, `f_update_time`) USING BTREE,
    KEY `idx_category_update_time` (`f_category`, `f_update_time`) USING BTREE,
    KEY `idx_create_user_update_time` (`f_create_user`, `f_update_time`) USING BTREE
) ENGINE = InnoDB COMMENT = 'Skill 主表';

CREATE TABLE IF NOT EXISTS `t_skill_file_index` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_skill_id` varchar(40) NOT NULL COMMENT 'Skill ID',
    `f_skill_version` varchar(40) NOT NULL COMMENT 'Skill 版本',
    `f_rel_path` varchar(512) NOT NULL COMMENT '文件相对路径',
    `f_path_hash` varchar(32) NOT NULL COMMENT '相对路径哈希',
    `f_storage_id` varchar(50) NOT NULL COMMENT '对象存储ID',
    `f_storage_key` text NOT NULL COMMENT '对象存储键',
    `f_file_type` varchar(40) NOT NULL COMMENT '文件类型',
    `f_content_sha256` varchar(64) NOT NULL COMMENT '文件内容 SHA256',
    `f_mime_type` varchar(128) NOT NULL DEFAULT '' COMMENT 'MIME 类型',
    `f_size` bigint(20) NOT NULL DEFAULT 0 COMMENT '文件大小',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `uk_skill_version_rel_path` (`f_skill_id`, `f_skill_version`, `f_rel_path`) USING BTREE,
    UNIQUE KEY `uk_skill_version_path_hash` (`f_skill_id`, `f_skill_version`, `f_path_hash`) USING BTREE
) ENGINE = InnoDB COMMENT = 'Skill 文件索引表';
