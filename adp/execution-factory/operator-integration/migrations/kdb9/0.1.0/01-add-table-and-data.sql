
SET SEARCH_PATH TO adp;


CREATE TABLE IF NOT EXISTS `t_metadata_api` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_summary` VARCHAR(256) NOT NULL COMMENT '摘要',
  `f_version` VARCHAR(40) NOT NULL COMMENT 'UUID',
  `f_svc_url` TEXT NOT NULL COMMENT '地址',
  `f_description` TEXT COMMENT '描述',
  `f_path` TEXT NOT NULL COMMENT 'API路径',
  `f_method` VARCHAR(50) NOT NULL COMMENT '请求方法',
  `f_api_spec` LONGTEXT DEFAULT NULL COMMENT 'API内容',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_metadata_api_uk_version` (f_version)
);


CREATE TABLE IF NOT EXISTS `t_op_registry` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_op_id` VARCHAR(40) NOT NULL COMMENT '算子ID,UUID',
  `f_name` VARCHAR(512) NOT NULL COMMENT '算子名称',
  `f_metadata_version` VARCHAR(40) NOT NULL COMMENT '算子元数据版本(关联t_metadata_api.f_version)',
  `f_metadata_type` VARCHAR(40) NOT NULL COMMENT '算子元数据类型(api/func/...)',
  `f_status` VARCHAR(10) DEFAULT 0 COMMENT '算子状态',
  `f_operator_type` VARCHAR(10) DEFAULT 0 COMMENT '算子类型, 0：基础算子, 1: 组合算子',
  `f_execution_mode` VARCHAR(10) DEFAULT 0 COMMENT '执行模式, 0: 同步执行, 1: 异步执行',
  `f_category` VARCHAR(50) DEFAULT 0 COMMENT '算子业务分类, 数据处理/算法模型',
  `f_source` VARCHAR(50) DEFAULT '' COMMENT '算子来源,system/unknown',
  `f_execute_control` TEXT DEFAULT NULL COMMENT '执行控制参数',
  `f_extend_info` TEXT DEFAULT NULL COMMENT '扩展信息',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  `f_is_deleted` TINYINT(1) DEFAULT 0 COMMENT '是否删除',
  `f_is_internal` TINYINT(1) DEFAULT 0 COMMENT '否为内置算子',
  `f_is_data_source` TINYINT(1) DEFAULT 0 COMMENT '是否为数据源算子',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_op_registry_uk_op_id_version` (f_op_id, f_metadata_version)
);

CREATE INDEX IF NOT EXISTS `idx_t_op_registry_idx_name_update` ON `t_op_registry` (f_name, f_update_time);
CREATE INDEX IF NOT EXISTS `idx_t_op_registry_idx_status_update` ON `t_op_registry` (f_status, f_update_time);
CREATE INDEX IF NOT EXISTS `idx_t_op_registry_idx_category_update` ON `t_op_registry` (f_category, f_update_time);
CREATE INDEX IF NOT EXISTS `idx_t_op_registry_idx_create_user_update` ON `t_op_registry` (f_create_user, f_update_time);
CREATE INDEX IF NOT EXISTS `idx_t_op_registry_idx_update_time` ON `t_op_registry` (f_update_time);


CREATE TABLE IF NOT EXISTS `t_toolbox` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_box_id` VARCHAR(40) NOT NULL COMMENT '工具箱ID',
  `f_name` VARCHAR(50) NOT NULL COMMENT '工具箱名称',
  `f_description` LONGTEXT NOT NULL COMMENT '工具箱描述',
  `f_svc_url` TEXT NOT NULL COMMENT '地址',
  `f_status` VARCHAR(50) NOT NULL COMMENT '状态',
  `f_is_internal` TINYINT(1) DEFAULT 0 COMMENT '是否为内置工具箱',
  `f_source` VARCHAR(50) DEFAULT '' COMMENT '工具箱来源, DIP/Custom',
  `f_category` VARCHAR(50) DEFAULT 0 COMMENT '工具箱分类, 数据处理/算法模型',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  `f_release_user` VARCHAR(50) NOT NULL COMMENT '发布者',
  `f_release_time` BIGINT(20) NOT NULL COMMENT '发布时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_toolbox_uk_box_id` (f_box_id)
);

CREATE INDEX IF NOT EXISTS `idx_t_toolbox_idx_name` ON `t_toolbox` (f_name);
CREATE INDEX IF NOT EXISTS `idx_t_toolbox_idx_status` ON `t_toolbox` (f_status);
CREATE INDEX IF NOT EXISTS `idx_t_toolbox_idx_category` ON `t_toolbox` (f_category);
CREATE INDEX IF NOT EXISTS `idx_t_toolbox_idx_creator_status` ON `t_toolbox` (f_create_user, f_status);
CREATE INDEX IF NOT EXISTS `idx_t_toolbox_idx_ctime` ON `t_toolbox` (f_create_time);
CREATE INDEX IF NOT EXISTS `idx_t_toolbox_idx_utime` ON `t_toolbox` (f_update_time);


CREATE TABLE IF NOT EXISTS `t_tool` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_tool_id` VARCHAR(40) NOT NULL COMMENT '工具ID',
  `f_box_id` VARCHAR(40) NOT NULL,
  `f_name` VARCHAR(256) NOT NULL COMMENT '名称',
  `f_description` LONGTEXT NOT NULL COMMENT '工具描述',
  `f_source_type` VARCHAR(50) NOT NULL COMMENT '来源类型, 0: 算子, 1: OpenAPI',
  `f_source_id` VARCHAR(40) NOT NULL COMMENT '来源ID',
  `f_status` VARCHAR(40) DEFAULT 0 COMMENT '状态,启用/禁用',
  `f_use_count` BIGINT(20) NOT NULL COMMENT '使用统计',
  `f_use_rule` LONGTEXT DEFAULT NULL COMMENT '规则',
  `f_parameters` LONGTEXT DEFAULT NULL COMMENT '参数',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  `f_extend_info` TEXT DEFAULT NULL COMMENT '扩展信息',
  `f_is_deleted` TINYINT(1) DEFAULT 0 COMMENT '是否删除',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_tool_uk_tool_id` (f_tool_id)
);

CREATE INDEX IF NOT EXISTS `idx_t_tool_idx_box_id` ON `t_tool` (f_box_id);
CREATE INDEX IF NOT EXISTS `idx_t_tool_idx_name_update` ON `t_tool` (f_name, f_update_time);


CREATE TABLE IF NOT EXISTS `t_mcp_server_config` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_mcp_id` VARCHAR(40) NOT NULL COMMENT 'mcp_id',
  `f_creation_type` VARCHAR(20) NOT NULL DEFAULT 'custom' COMMENT '创建类型',
  `f_name` VARCHAR(50) NOT NULL COMMENT 'MCP Server名称，全局唯一',
  `f_description` TEXT NOT NULL COMMENT '描述信息',
  `f_mode` VARCHAR(32) NOT NULL COMMENT '通信模式（sse、streamable、stdio_npx、stdio_uvx）',
  `f_url` TEXT NOT NULL COMMENT '通信地址,SSE/Streamable模式下的服务URL',
  `f_headers` TEXT NOT NULL COMMENT 'http请求头,JSON字符串',
  `f_command` TEXT NOT NULL COMMENT 'stdio模式下的命令,JSON字符串',
  `f_env` TEXT NOT NULL COMMENT '环境变量,JSON字符串',
  `f_args` TEXT NOT NULL COMMENT '命令参数,JSON字符串',
  `f_status` VARCHAR(30) NOT NULL DEFAULT 'unpublish' COMMENT '状态',
  `f_is_internal` TINYINT(1) DEFAULT 0 COMMENT '是否为内置',
  `f_source` VARCHAR(50) DEFAULT 0 COMMENT '服务来源',
  `f_category` VARCHAR(50) DEFAULT 0 COMMENT '分类',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  `f_version` INT(20) NOT NULL DEFAULT 0 COMMENT '版本号',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_mcp_server_config_uk_mcp_id` (f_mcp_id)
);

CREATE INDEX IF NOT EXISTS `idx_t_mcp_server_config_idx_name` ON `t_mcp_server_config` (`f_name`);
CREATE INDEX IF NOT EXISTS `idx_t_mcp_server_config_idx_update_time` ON `t_mcp_server_config` (`f_update_time`);
CREATE INDEX IF NOT EXISTS `idx_t_mcp_server_config_idx_status` ON `t_mcp_server_config` (`f_status`);


CREATE TABLE IF NOT EXISTS `t_mcp_tool` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_mcp_tool_id` VARCHAR(40) NOT NULL COMMENT 'mcp_tool_id',
  `f_mcp_id` VARCHAR(40) NOT NULL COMMENT 'mcp_id',
  `f_mcp_version` INT(20) NOT NULL COMMENT 'mcp版本',
  `f_box_id` VARCHAR(40) NOT NULL COMMENT '工具箱ID',
  `f_box_name` VARCHAR(50) COMMENT '工具箱名称',
  `f_tool_id` VARCHAR(40) NOT NULL COMMENT '工具ID',
  `f_name` VARCHAR(256)  COMMENT '工具名称',
  `f_description` LONGTEXT  COMMENT '工具描述',
  `f_use_rule` LONGTEXT  COMMENT '使用规则',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_mcp_tool_uk_mcp_tool_id` (f_mcp_tool_id)
);

CREATE INDEX IF NOT EXISTS `idx_t_mcp_tool_idx_mcp_id_version` ON `t_mcp_tool` (f_mcp_id, f_mcp_version);
CREATE INDEX IF NOT EXISTS `idx_t_mcp_tool_idx_name_update` ON `t_mcp_tool` (f_name, f_update_time);


CREATE TABLE IF NOT EXISTS `t_mcp_server_release` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_mcp_id` VARCHAR(40) NOT NULL COMMENT 'mcp_id',
  `f_creation_type` VARCHAR(20) NOT NULL DEFAULT 'custom' COMMENT '创建类型',
  `f_name` VARCHAR(50) NOT NULL COMMENT 'MCP Server名称，全局唯一',
  `f_description` TEXT NOT NULL COMMENT '描述信息',
  `f_mode` VARCHAR(32) NOT NULL COMMENT '通信模式（sse、streamable、stdio_npx、stdio_uvx）',
  `f_url` TEXT NOT NULL COMMENT '通信地址,SSE/Streamable模式下的服务URL',
  `f_headers` TEXT NOT NULL COMMENT 'http请求头,JSON字符串',
  `f_command` TEXT NOT NULL COMMENT 'stdio模式下的命令,JSON字符串',
  `f_env` TEXT NOT NULL COMMENT '环境变量,JSON字符串',
  `f_args` TEXT NOT NULL COMMENT '命令参数,JSON字符串',
  `f_status` VARCHAR(30) NOT NULL DEFAULT 'draft' COMMENT '状态',
  `f_is_internal` TINYINT(1) DEFAULT 0 COMMENT '是否为内置',
  `f_source` VARCHAR(50) DEFAULT 0 COMMENT '服务来源',
  `f_category` VARCHAR(50) DEFAULT 0 COMMENT '分类',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  `f_version` INT(20) NOT NULL COMMENT '发布版本号',
  `f_release_desc` VARCHAR(50) NOT NULL COMMENT '发布描述',
  `f_release_user` VARCHAR(50) NOT NULL COMMENT '发布者',
  `f_release_time` BIGINT(20) NOT NULL COMMENT '发布时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_mcp_server_release_uk_mcp` (f_mcp_id, f_version)
);

CREATE INDEX IF NOT EXISTS `idx_t_mcp_server_release_idx_mcp_id_create_time` ON `t_mcp_server_release` (`f_mcp_id`, `f_create_time`);
CREATE INDEX IF NOT EXISTS `idx_t_mcp_server_release_idx_status_update_time` ON `t_mcp_server_release` (`f_status`, `f_update_time`);


CREATE TABLE IF NOT EXISTS `t_mcp_server_release_history` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_mcp_id` VARCHAR(40) NOT NULL COMMENT 'mcp_id',
  `f_mcp_release` LONGTEXT NOT NULL COMMENT 'mcp server 发布信息',
  `f_version` INT NOT NULL COMMENT '发布版本',
  `f_release_desc` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '发布描述',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_mcp_server_release_history_uk_mcp` (f_mcp_id, f_version)
);

CREATE INDEX IF NOT EXISTS `idx_t_mcp_server_release_history_idx_mcp_id_create_time` ON `t_mcp_server_release_history` (`f_mcp_id`, `f_create_time`);


CREATE TABLE IF NOT EXISTS `t_internal_component_config` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_component_type` VARCHAR(50) NOT NULL COMMENT '组件类型, toolbox/mcp/operator',
  `f_component_id` VARCHAR(40) NOT NULL COMMENT '组件ID',
  `f_config_version` VARCHAR(40) NOT NULL COMMENT '配置版本: x.y.z格式',
  `f_config_source` VARCHAR(40) NOT NULL COMMENT '配置来源: (auto自动/manual手动)',
  `f_protected_flag` TINYINT(1) DEFAULT 0 COMMENT '手动配置保护锁',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_internal_component_config_uk_comp_type_id` (`f_component_type`,`f_component_id`)
);



CREATE TABLE IF NOT EXISTS `t_operator_release` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_op_id` VARCHAR(40) NOT NULL COMMENT '算子ID,UUID',
  `f_name` VARCHAR(512) NOT NULL COMMENT '算子名称',
  `f_metadata_version` VARCHAR(40) NOT NULL COMMENT '算子元数据版本(关联t_metadata_api.f_version)',
  `f_metadata_type` VARCHAR(40) NOT NULL COMMENT '算子元数据类型(api/func/...)',
  `f_status` VARCHAR(10) DEFAULT 0 COMMENT '算子状态',
  `f_operator_type` VARCHAR(10) DEFAULT 0 COMMENT '算子类型, 0：基础算子, 1: 组合算子',
  `f_execution_mode` VARCHAR(10) DEFAULT 0 COMMENT '执行模式, 0: 同步执行, 1: 异步执行',
  `f_category` VARCHAR(50) DEFAULT 0 COMMENT '算子业务分类, 数据处理/算法模型',
  `f_source` VARCHAR(50) DEFAULT '' COMMENT '算子来源,system/unknown',
  `f_execute_control` TEXT DEFAULT NULL COMMENT '执行控制参数',
  `f_extend_info` TEXT DEFAULT NULL COMMENT '扩展信息',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  `f_tag` INT(20) NOT NULL COMMENT '发布版本号',
  `f_release_user` VARCHAR(50) NOT NULL COMMENT '发布者',
  `f_release_time` BIGINT(20) NOT NULL COMMENT '发布时间',
  `f_is_internal` TINYINT(1) DEFAULT 0 COMMENT '否为内置算子',
  `f_is_data_source` TINYINT(1) DEFAULT 0 COMMENT '是否为数据源算子',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_operator_release_uk_op` (f_op_id, f_tag)
);

CREATE INDEX IF NOT EXISTS `idx_t_operator_release_idx_op_id_create_time` ON `t_operator_release` (`f_op_id`, `f_create_time`);
CREATE INDEX IF NOT EXISTS `idx_t_operator_release_idx_status_update_time` ON `t_operator_release` (`f_status`, `f_update_time`);


CREATE TABLE IF NOT EXISTS `t_operator_release_history` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_op_id` VARCHAR(40) NOT NULL COMMENT '算子ID',
  `f_op_release` LONGTEXT NOT NULL COMMENT '算子发布信息',
  `f_metadata_version` VARCHAR(40) NOT NULL COMMENT '元数据版本(关联t_metadata_api.f_version)',
  `f_metadata_type` VARCHAR(40) NOT NULL COMMENT '元数据类型(api/func/...)',
  `f_tag` INT NOT NULL COMMENT '发布版本',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_operator_release_history_uk_op` (f_op_id, f_tag)
);

CREATE INDEX IF NOT EXISTS `idx_t_operator_release_history_idx_op_id_create_time` ON `t_operator_release_history` (`f_op_id`, `f_create_time`);
CREATE INDEX IF NOT EXISTS `idx_t_operator_release_history_idx_op_id_metadata_version` ON `t_operator_release_history` (`f_op_id`, `f_metadata_version`);


CREATE TABLE IF NOT EXISTS `t_category` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_category_id` VARCHAR(40) NOT NULL COMMENT '分类ID',
  `f_category_name` VARCHAR(50) NOT NULL COMMENT '分类名称',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_category_uk_category_id` (f_category_id),
  UNIQUE KEY `idx_t_category_uk_category_name` (f_category_name)
);



CREATE TABLE IF NOT EXISTS `t_outbox_message` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_event_id` VARCHAR(40) NOT NULL COMMENT '事件ID',
  `f_event_type` VARCHAR(40) NOT NULL COMMENT '事件类型',
  `f_topic` TEXT NOT NULL COMMENT '主题',
  `f_payload` LONGTEXT NOT NULL COMMENT '消息体',
  `f_status` VARCHAR(40) NOT NULL COMMENT '状态',
  `f_created_at` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_updated_at` BIGINT(20) NOT NULL COMMENT '更新时间',
  `f_next_retry_at` BIGINT(20) NOT NULL COMMENT '下次重试时间',
  `f_retry_count` INT(20) NOT NULL COMMENT '重试次数',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_outbox_message_uk_event_id` (f_event_id)
);

CREATE INDEX IF NOT EXISTS `idx_t_outbox_message_idx_event_type` ON `t_outbox_message` (f_event_type);
CREATE INDEX IF NOT EXISTS `idx_t_outbox_message_idx_status_next_retry` ON `t_outbox_message` (f_status,f_next_retry_at);


CREATE TABLE IF NOT EXISTS `t_metadata_function` (
  `f_id` BIGSERIAL NOT NULL COMMENT '自增主键',
  `f_summary` VARCHAR(256) NOT NULL COMMENT '摘要',
  `f_version` VARCHAR(40) NOT NULL COMMENT 'UUID',
  `f_svc_url` TEXT NOT NULL COMMENT '地址',
  `f_description` TEXT COMMENT '描述',
  `f_path` TEXT NOT NULL COMMENT 'API路径',
  `f_method` VARCHAR(50) NOT NULL COMMENT '请求方法',
  `f_code` LONGTEXT NOT NULL COMMENT '函数代码',
  `f_script_type` VARCHAR(50) NOT NULL COMMENT '脚本类型',
  `f_dependencies` LONGTEXT DEFAULT NULL COMMENT '依赖库',
  `f_api_spec` LONGTEXT DEFAULT NULL COMMENT 'API内容',
  `f_create_user` VARCHAR(50) NOT NULL COMMENT '创建者',
  `f_update_user` VARCHAR(50) NOT NULL COMMENT '编辑者',
  `f_create_time` BIGINT(20) NOT NULL COMMENT '创建时间',
  `f_update_time` BIGINT(20) NOT NULL COMMENT '编辑时间',
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `idx_t_metadata_function_uk_version` (f_version)
);

