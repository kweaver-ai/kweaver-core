
SET SEARCH_PATH TO adp;


CREATE TABLE IF NOT EXISTS `t_python_package` (
  `f_id` VARCHAR(32) NOT NULL COMMENT '主键ID',
  `f_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
  `f_oss_id` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'ossid',
  `f_oss_key` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'key',
  `f_creator_id` VARCHAR(36) NOT NULL DEFAULT '' COMMENT '创建者id',
  `f_creator_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '创建者名称',
  `f_created_at` BIGINT(20) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_python_package_uk_t_python_package_name` (`f_name`)
);


