
SET SEARCH_PATH TO adp;


CREATE TABLE IF NOT EXISTS `t_resource_deploy` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_resource_id` VARCHAR(40) NOT NULL COMMENT '资源ID',
  `f_type` VARCHAR(40) NOT NULL COMMENT '资源类型',
  `f_version` INT(20) NOT NULL COMMENT '资源版本',
  `f_name` VARCHAR(40) NOT NULL COMMENT '资源名称',
  `f_description` LONGTEXT NOT NULL COMMENT '资源描述',
  `f_config` LONGTEXT NOT NULL COMMENT '资源配置',
  `f_status` VARCHAR(40) NOT NULL COMMENT '资源状态',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_resource_deploy_uk_resource_id` (f_resource_id, f_type, f_version)
);