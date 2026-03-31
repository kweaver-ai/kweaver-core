
SET SEARCH_PATH TO adp;


CREATE TABLE IF NOT EXISTS `t_skill_repository` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_skill_id` VARCHAR(40) NOT NULL COMMENT 'Skill ID',
  `f_name` VARCHAR(255) NOT NULL COMMENT 'Skill 名称',
  `f_description` LONGTEXT NOT NULL COMMENT 'Skill 描述',
  `f_skill_content` LONGTEXT NOT NULL COMMENT 'Skill 指令正文',
  `f_version` VARCHAR(40) NOT NULL COMMENT 'Skill 版本',
  `f_status` VARCHAR(40) NOT NULL COMMENT 'Skill 状态',
  `f_source` VARCHAR(50) NOT NULL DEFAULT '' COMMENT 'Skill 来源',
  `f_extend_info` TEXT DEFAULT NULL COMMENT '扩展信息',
  `f_dependencies` TEXT DEFAULT NULL COMMENT '依赖信息',
  `f_file_manifest` LONGTEXT DEFAULT NULL COMMENT '文件摘要清单',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  `f_delete_user` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '删除者',
  `f_delete_time` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '删除时间',
  `f_category` VARCHAR(50 CHAR) DEFAULT '' COMMENT '工具箱分类, 数据处理/算法模型',
  `f_is_deleted` BOOLEAN DEFAULT 0 COMMENT '是否删除', -- 0: 未删除, 1: 待删除
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_skill_repository_uk_skill_id` (f_skill_id)
);

CREATE INDEX IF NOT EXISTS `idx_t_skill_repository_idx_status_update_time` ON `t_skill_repository` (f_status, f_update_time);
CREATE INDEX IF NOT EXISTS `idx_t_skill_repository_idx_category_update_time` ON `t_skill_repository` (f_category, f_update_time);
CREATE INDEX IF NOT EXISTS `idx_t_skill_repository_idx_create_user_update_time` ON `t_skill_repository` (f_create_user, f_update_time);


CREATE TABLE IF NOT EXISTS `t_skill_file_index` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_skill_id` VARCHAR(40) NOT NULL COMMENT 'Skill ID',
  `f_skill_version` VARCHAR(40) NOT NULL COMMENT 'Skill 版本',
  `f_rel_path` VARCHAR(512) NOT NULL COMMENT '文件相对路径',
  `f_path_hash` VARCHAR(32) NOT NULL COMMENT '相对路径哈希',
  `f_storage_id` VARCHAR(50) NOT NULL COMMENT '对象存储ID',
  `f_storage_key` TEXT NOT NULL COMMENT '对象存储键',
  `f_file_type` VARCHAR(40) NOT NULL COMMENT '文件类型',
  `f_content_sha256` VARCHAR(64) NOT NULL COMMENT '文件内容 SHA256',
  `f_mime_type` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'MIME 类型',
  `f_size` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '文件大小',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_skill_file_index_uk_skill_version_rel_path` (f_skill_id, f_skill_version, f_rel_path),
  UNIQUE KEY `idx_t_skill_file_index_uk_skill_version_path_hash` (f_skill_id, f_skill_version, f_path_hash)
);
