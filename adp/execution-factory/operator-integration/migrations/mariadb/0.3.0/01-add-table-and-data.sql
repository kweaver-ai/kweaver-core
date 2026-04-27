USE adp;

CREATE TABLE IF NOT EXISTS `t_resource_deploy` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_resource_id` varchar(40) NOT NULL COMMENT '资源ID',
    `f_type` varchar(40) NOT NULL COMMENT '资源类型',
    `f_version` int(20) NOT NULL COMMENT '资源版本',
    `f_name` varchar(40) NOT NULL COMMENT '资源名称',
    `f_description` longtext NOT NULL COMMENT '资源描述',
    `f_config` longtext NOT NULL COMMENT '资源配置',
    `f_status` varchar(40) NOT NULL COMMENT '资源状态',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_resource_id (f_resource_id, f_type, f_version) USING BTREE
) ENGINE = InnoDB COMMENT = '资源部署表';