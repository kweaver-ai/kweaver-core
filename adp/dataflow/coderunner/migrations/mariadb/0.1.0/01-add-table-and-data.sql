USE adp;

CREATE TABLE IF NOT EXISTS `t_python_package` (
  `f_id` varchar(32) NOT NULL COMMENT '主键ID',
  `f_name` varchar(255) NOT NULL DEFAULT '' COMMENT '名称',
  `f_oss_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'ossid',
  `f_oss_key` varchar(32) NOT NULL DEFAULT '' COMMENT 'key',
  `f_creator_id` varchar(36) NOT NULL DEFAULT '' COMMENT '创建者id',
  `f_creator_name` varchar(128) NOT NULL DEFAULT '' COMMENT '创建者名称',
  `f_created_at` bigint(20) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `uk_t_python_package_name` (`f_name`)
) ENGINE=InnoDB COMMENT='包管理表';