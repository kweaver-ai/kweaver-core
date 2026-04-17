USE kweaver;

CREATE TABLE IF NOT EXISTS `t_skill_release` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_skill_id` varchar(40) NOT NULL COMMENT '技能ID',
    `f_name` varchar(512) NOT NULL COMMENT '技能名称',
    `f_description` longtext NOT NULL COMMENT '技能描述',
    `f_skill_content` longtext NOT NULL COMMENT '技能主体内容',
    `f_version` varchar(40) NOT NULL COMMENT '发布快照版本',
    `f_category` varchar(50) DEFAULT '' COMMENT '技能分类',
    `f_source` varchar(50) DEFAULT '' COMMENT '技能来源',
    `f_extend_info` longtext DEFAULT NULL COMMENT '扩展信息',
    `f_dependencies` longtext DEFAULT NULL COMMENT '依赖信息',
    `f_file_manifest` longtext DEFAULT NULL COMMENT '文件清单',
    `f_status` varchar(20) NOT NULL COMMENT '技能状态',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    `f_release_user` varchar(50) NOT NULL COMMENT '发布者',
    `f_release_time` bigint(20) NOT NULL COMMENT '发布时间',
    `f_release_desc` varchar(255) NOT NULL DEFAULT '' COMMENT '发布描述',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_skill_release_skill_id (`f_skill_id`) USING BTREE,
    KEY `idx_skill_release_status_update_time` (`f_status`, `f_update_time`) USING BTREE,
    KEY `idx_skill_release_release_time` (`f_release_time`) USING BTREE
) ENGINE = InnoDB COMMENT = '技能当前发布快照表';

CREATE TABLE IF NOT EXISTS `t_skill_release_history` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_skill_id` varchar(40) NOT NULL COMMENT '技能ID',
    `f_version` varchar(40) NOT NULL COMMENT '发布快照版本',
    `f_skill_release` longtext NOT NULL COMMENT '技能发布快照',
    `f_release_desc` varchar(255) NOT NULL DEFAULT '' COMMENT '发布描述',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_skill_release_history (`f_skill_id`, `f_version`) USING BTREE,
    KEY `idx_skill_release_history_skill_id_create_time` (`f_skill_id`, `f_create_time`) USING BTREE
) ENGINE = InnoDB COMMENT = '技能发布历史表';