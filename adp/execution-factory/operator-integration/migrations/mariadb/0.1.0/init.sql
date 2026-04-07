USE adp;

CREATE TABLE IF NOT EXISTS `t_metadata_api` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_summary` varchar(256) NOT NULL COMMENT '摘要', -- 接口摘要 (唯一索引)
    `f_version` varchar(40) NOT NULL COMMENT 'UUID',
    `f_svc_url` text NOT NULL COMMENT '地址', -- 编辑后生成新的版本
    `f_description` text COMMENT '描述', -- 单个接口描述（详情）
    `f_path` text NOT NULL COMMENT 'API路径',
    `f_method` varchar(50) NOT NULL COMMENT '请求方法',
    `f_api_spec` longtext DEFAULT NULL COMMENT 'API内容',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_version (`f_version`) USING BTREE
) ENGINE = InnoDB COMMENT = 'API元数据表';

CREATE TABLE IF NOT EXISTS `t_op_registry` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_op_id` varchar(40) NOT NULL COMMENT '算子ID,UUID',
    `f_name` varchar(512) NOT NULL COMMENT '算子名称', -- 默认为摘要
    `f_metadata_version` varchar(40) NOT NULL COMMENT '算子元数据版本(关联t_metadata_api.f_version)',
    `f_metadata_type` varchar(40) NOT NULL COMMENT '算子元数据类型(api/func/...)',
    `f_status` varchar(10) DEFAULT 0 COMMENT '算子状态',
    `f_operator_type` varchar(10) DEFAULT 0 COMMENT '算子类型, 0：基础算子, 1: 组合算子',
    `f_execution_mode` varchar(10) DEFAULT 0 COMMENT '执行模式, 0: 同步执行, 1: 异步执行',
    `f_category` varchar(50) DEFAULT 0 COMMENT '算子业务分类, 数据处理/算法模型',
    `f_source` varchar(50) DEFAULT '' COMMENT '算子来源,system/unknown',
    `f_execute_control` text DEFAULT NULL COMMENT '执行控制参数', -- 超时重试策略
    `f_extend_info` text DEFAULT NULL COMMENT '扩展信息', -- flow_id一类
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    `f_is_deleted` boolean DEFAULT 0 COMMENT '是否删除', -- 0: 未删除, 1: 待删除
    `f_is_internal` boolean DEFAULT 0 COMMENT '否为内置算子',
    `f_is_data_source` boolean DEFAULT 0 COMMENT '是否为数据源算子',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_op_id_version (f_op_id, f_metadata_version) USING BTREE,
    KEY idx_name_update (f_name, f_update_time) USING BTREE,
    KEY idx_status_update (f_status, f_update_time) USING BTREE,
    KEY idx_category_update (f_category, f_update_time) USING BTREE,
    KEY idx_create_user_update (f_create_user, f_update_time) USING BTREE,
    KEY idx_update_time (f_update_time) USING BTREE
) ENGINE = InnoDB COMMENT = '算子注册表';


CREATE TABLE IF NOT EXISTS `t_toolbox` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_box_id` varchar(40) NOT NULL COMMENT '工具箱ID',
    `f_name` varchar(50) NOT NULL COMMENT '工具箱名称',
    `f_description` longtext NOT NULL COMMENT '工具箱描述',
    `f_svc_url` text NOT NULL COMMENT '地址',
    `f_status` varchar(50) NOT NULL COMMENT '状态',
    `f_is_internal` boolean DEFAULT 0 COMMENT '是否为内置工具箱', -- 0: 不是, 1: 是
    `f_source` varchar(50) DEFAULT '' COMMENT '工具箱来源, DIP/Custom',
    `f_category` varchar(50) DEFAULT 0 COMMENT '工具箱分类, 数据处理/算法模型',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    `f_release_user` varchar(50) NOT NULL COMMENT '发布者',
    `f_release_time` bigint(20) NOT NULL COMMENT '发布时间',
    `f_metadata_type` varchar(50) NOT NULL DEFAULT 'openapi' COMMENT '工具箱内工具的类型，OpenAPI/Function',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_box_id (f_box_id) USING BTREE,
    KEY idx_name (f_name) USING BTREE,
    KEY idx_status (f_status) USING BTREE,
    KEY idx_category (f_category) USING BTREE,
    KEY idx_creator_status (f_create_user, f_status) USING BTREE,
    KEY idx_ctime (f_create_time) USING BTREE,
    KEY idx_utime (f_update_time) USING BTREE
) ENGINE = InnoDB COMMENT = '工具箱';


CREATE TABLE IF NOT EXISTS `t_tool` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_tool_id` varchar(40) NOT NULL COMMENT '工具ID',
    `f_box_id` varchar(40) NOT NULL,
    `f_name` varchar(256) NOT NULL COMMENT '名称',
    `f_description` longtext NOT NULL COMMENT '工具描述',
    `f_source_type` varchar(50) NOT NULL COMMENT '来源类型, 0: 算子, 1: OpenAPI',
    `f_source_id` varchar(40) NOT NULL COMMENT '来源ID',
    `f_status` varchar(40) DEFAULT 0 COMMENT '状态,启用/禁用',
    `f_use_count` bigint(20) NOT NULL COMMENT '使用统计',
    `f_use_rule` longtext DEFAULT NULL COMMENT '规则',
    `f_parameters` longtext DEFAULT NULL COMMENT '参数',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    `f_extend_info` text DEFAULT NULL COMMENT '扩展信息',
    `f_is_deleted` boolean DEFAULT 0 COMMENT '是否删除', -- 0: 未删除, 1: 待删除
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_tool_id (f_tool_id) USING BTREE,
    KEY idx_box_id (f_box_id) USING BTREE,
    KEY idx_name_update (f_name, f_update_time) USING BTREE
) ENGINE = InnoDB COMMENT = '工具表';


CREATE TABLE IF NOT EXISTS `t_mcp_server_config` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_mcp_id` varchar(40) NOT NULL COMMENT 'mcp_id',
    `f_creation_type` varchar(20) NOT NULL DEFAULT 'custom' COMMENT '创建类型',
    `f_name` varchar(50) NOT NULL COMMENT 'MCP Server名称，全局唯一',
    `f_description` text NOT NULL COMMENT '描述信息',
    `f_mode` varchar(32) NOT NULL COMMENT '通信模式（sse、streamable、stdio_npx、stdio_uvx）',
    `f_url` text NOT NULL COMMENT '通信地址,SSE/Streamable模式下的服务URL',
    `f_headers` text NOT NULL COMMENT 'http请求头,JSON字符串',
    `f_command` text NOT NULL COMMENT 'stdio模式下的命令,JSON字符串',
    `f_env` text NOT NULL COMMENT '环境变量,JSON字符串',
    `f_args` text NOT NULL COMMENT '命令参数,JSON字符串',
    `f_status` varchar(30) NOT NULL DEFAULT 'unpublish' COMMENT '状态',
    `f_is_internal` boolean DEFAULT 0 COMMENT '是否为内置', -- 0: 不是, 1: 是
    `f_source` varchar(50) DEFAULT 0 COMMENT '服务来源',
    `f_category` varchar(50) DEFAULT 0 COMMENT '分类',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    `f_version` int(20) NOT NULL DEFAULT 0 COMMENT '版本号',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_mcp_id (f_mcp_id) USING BTREE,
    KEY `idx_name` (`f_name`) USING BTREE,
    KEY `idx_update_time` (`f_update_time`) USING BTREE,
    KEY `idx_status`(`f_status`) USING BTREE
) ENGINE = InnoDB COMMENT = 'MCP Server表';

CREATE TABLE IF NOT EXISTS `t_mcp_tool` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_mcp_tool_id` varchar(40) NOT NULL COMMENT 'mcp_tool_id',
    `f_mcp_id` varchar(40) NOT NULL COMMENT 'mcp_id',
    `f_mcp_version` int(20) NOT NULL COMMENT 'mcp版本',
    `f_box_id` varchar(40) NOT NULL COMMENT '工具箱ID',
    `f_box_name` varchar(50) COMMENT '工具箱名称',
    `f_tool_id` varchar(40) NOT NULL COMMENT '工具ID',
    `f_name` varchar(256)  COMMENT '工具名称',
    `f_description` longtext  COMMENT '工具描述',
    `f_use_rule` longtext  COMMENT '使用规则',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_mcp_tool_id (f_mcp_tool_id) USING BTREE,
    KEY idx_mcp_id_version (f_mcp_id, f_mcp_version) USING BTREE,
    KEY idx_name_update (f_name, f_update_time) USING BTREE
) ENGINE = InnoDB COMMENT = 'MCP工具表';

CREATE TABLE IF NOT EXISTS `t_mcp_server_release` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_mcp_id` varchar(40) NOT NULL COMMENT 'mcp_id',
    `f_creation_type` varchar(20) NOT NULL DEFAULT 'custom' COMMENT '创建类型',
    `f_name` varchar(50) NOT NULL COMMENT 'MCP Server名称，全局唯一',
    `f_description` text NOT NULL COMMENT '描述信息',
    `f_mode` varchar(32) NOT NULL COMMENT '通信模式（sse、streamable、stdio_npx、stdio_uvx）',
    `f_url` text NOT NULL COMMENT '通信地址,SSE/Streamable模式下的服务URL',
    `f_headers` text NOT NULL COMMENT 'http请求头,JSON字符串',
    `f_command` text NOT NULL COMMENT 'stdio模式下的命令,JSON字符串',
    `f_env` text NOT NULL COMMENT '环境变量,JSON字符串',
    `f_args` text NOT NULL COMMENT '命令参数,JSON字符串',
    `f_status` varchar(30) NOT NULL DEFAULT 'draft' COMMENT '状态',
    `f_is_internal` boolean DEFAULT 0 COMMENT '是否为内置', -- 0: 不是, 1: 是
    `f_source` varchar(50) DEFAULT 0 COMMENT '服务来源',
    `f_category` varchar(50) DEFAULT 0 COMMENT '分类',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    `f_version` int(20) NOT NULL COMMENT '发布版本号',
    `f_release_desc` varchar(50) NOT NULL COMMENT '发布描述',
    `f_release_user` varchar(50) NOT NULL COMMENT '发布者',
    `f_release_time` bigint(20) NOT NULL COMMENT '发布时间',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_mcp (f_mcp_id, f_version) USING BTREE,
    KEY `idx_mcp_id_create_time` (`f_mcp_id`, `f_create_time`) USING BTREE,
    KEY  `idx_status_update_time` (`f_status`, `f_update_time`) USING BTREE
) ENGINE = InnoDB COMMENT = 'MCP Server发布表';

CREATE TABLE IF NOT EXISTS `t_mcp_server_release_history` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_mcp_id` varchar(40) NOT NULL COMMENT 'mcp_id',
    `f_mcp_release` longtext NOT NULL COMMENT 'mcp server 发布信息',
    `f_version` int NOT NULL COMMENT '发布版本',
    `f_release_desc` varchar(255) NOT NULL DEFAULT '' COMMENT '发布描述',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_mcp (f_mcp_id, f_version) USING BTREE,
    KEY `idx_mcp_id_create_time` (`f_mcp_id`, `f_create_time`) USING BTREE
) ENGINE = InnoDB COMMENT = 'MCP Server发布历史表';

CREATE TABLE IF NOT EXISTS `t_internal_component_config` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_component_type` varchar(50) NOT NULL COMMENT '组件类型, toolbox/mcp/operator',
    `f_component_id` varchar(40) NOT NULL COMMENT '组件ID',
    `f_config_version` varchar(40) NOT NULL COMMENT '配置版本: x.y.z格式',
    `f_config_source` varchar(40) NOT NULL COMMENT '配置来源: (auto自动/manual手动)',
    `f_protected_flag` boolean DEFAULT 0 COMMENT '手动配置保护锁',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY `uk_comp_type_id` (`f_component_type`,`f_component_id`) USING BTREE
) ENGINE = InnoDB COMMENT = '内置组件配置表';


CREATE TABLE IF NOT EXISTS `t_operator_release` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_op_id` varchar(40) NOT NULL COMMENT '算子ID,UUID',
    `f_name` varchar(512) NOT NULL COMMENT '算子名称', -- 默认为摘要
    `f_metadata_version` varchar(40) NOT NULL COMMENT '算子元数据版本(关联t_metadata_api.f_version)',
    `f_metadata_type` varchar(40) NOT NULL COMMENT '算子元数据类型(api/func/...)',
    `f_status` varchar(10) DEFAULT 0 COMMENT '算子状态',
    `f_operator_type` varchar(10) DEFAULT 0 COMMENT '算子类型, 0：基础算子, 1: 组合算子',
    `f_execution_mode` varchar(10) DEFAULT 0 COMMENT '执行模式, 0: 同步执行, 1: 异步执行',
    `f_category` varchar(50) DEFAULT 0 COMMENT '算子业务分类, 数据处理/算法模型',
    `f_source` varchar(50) DEFAULT '' COMMENT '算子来源,system/unknown',
    `f_execute_control` text DEFAULT NULL COMMENT '执行控制参数', -- 超时重试策略
    `f_extend_info` text DEFAULT NULL COMMENT '扩展信息', -- flow_id一类
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    `f_tag` int(20) NOT NULL COMMENT '发布版本号',
    `f_release_user` varchar(50) NOT NULL COMMENT '发布者',
    `f_release_time` bigint(20) NOT NULL COMMENT '发布时间',
    `f_is_internal` boolean DEFAULT 0 COMMENT '否为内置算子', -- 0: 不是, 1: 是
    `f_is_data_source` boolean DEFAULT 0 COMMENT '是否为数据源算子', -- 0: 不是, 1: 是
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_op (f_op_id, f_tag) USING BTREE,
    KEY `idx_op_id_create_time` (`f_op_id`, `f_create_time`) USING BTREE,
    KEY `idx_status_update_time` (`f_status`, `f_update_time`) USING BTREE
) ENGINE = InnoDB COMMENT = '算子发布表';

CREATE TABLE IF NOT EXISTS `t_operator_release_history` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_op_id` varchar(40) NOT NULL COMMENT '算子ID',
    `f_op_release` longtext NOT NULL COMMENT '算子发布信息',
    `f_metadata_version` varchar(40) NOT NULL COMMENT '元数据版本(关联t_metadata_api.f_version)',
    `f_metadata_type` varchar(40) NOT NULL COMMENT '元数据类型(api/func/...)',
    `f_tag` int NOT NULL COMMENT '发布版本',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_op (f_op_id, f_tag) USING BTREE,
    KEY `idx_op_id_create_time` (`f_op_id`, `f_create_time`) USING BTREE,
    KEY `idx_op_id_metadata_version` (`f_op_id`, `f_metadata_version`) USING BTREE
) ENGINE = InnoDB COMMENT = '算子发布历史表';

CREATE TABLE IF NOT EXISTS `t_category` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_category_id` varchar(40) NOT NULL COMMENT '分类ID',
    `f_category_name` varchar(50) NOT NULL COMMENT '分类名称',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_category_id (f_category_id) USING BTREE,
    UNIQUE KEY uk_category_name (f_category_name) USING BTREE
) ENGINE = InnoDB COMMENT = '分类表';

CREATE TABLE IF NOT EXISTS `t_outbox_message` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_event_id` varchar(40) NOT NULL COMMENT '事件ID',
    `f_event_type` varchar(40) NOT NULL COMMENT '事件类型',
    `f_topic` text NOT NULL COMMENT '主题',
    `f_payload` longtext NOT NULL COMMENT '消息体',
    `f_status` varchar(40) NOT NULL COMMENT '状态',
    `f_created_at` bigint(20) NOT NULL COMMENT '创建时间',
    `f_updated_at` bigint(20) NOT NULL COMMENT '更新时间',
    `f_next_retry_at` bigint(20) NOT NULL COMMENT '下次重试时间',
    `f_retry_count` int(20) NOT NULL COMMENT '重试次数',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_event_id (f_event_id) USING BTREE,
    KEY idx_event_type (f_event_type) USING BTREE,
    KEY idx_status_next_retry (f_status,f_next_retry_at) USING BTREE
) ENGINE = InnoDB COMMENT = 'outbox消息表';

CREATE TABLE IF NOT EXISTS `t_metadata_function` (
    `f_id` bigint AUTO_INCREMENT NOT NULL COMMENT '自增主键',
    `f_summary` varchar(256) NOT NULL COMMENT '摘要', -- 接口摘要 (唯一索引)
    `f_version` varchar(40) NOT NULL COMMENT 'UUID',
    `f_svc_url` text NOT NULL COMMENT '地址', -- 编辑后生成新的版本
    `f_description` text COMMENT '描述', -- 单个接口描述（详情）
    `f_path` text NOT NULL COMMENT 'API路径',
    `f_method` varchar(50) NOT NULL COMMENT '请求方法',
    `f_code` longtext NOT NULL COMMENT '函数代码',
    `f_script_type` varchar(50) NOT NULL COMMENT '脚本类型',
    `f_dependencies` longtext DEFAULT NULL COMMENT '依赖库',
    `f_api_spec` longtext DEFAULT NULL COMMENT 'API内容',
    `f_create_user` varchar(50) NOT NULL COMMENT '创建者',
    `f_update_user` varchar(50) NOT NULL COMMENT '编辑者',
    `f_create_time` bigint(20) NOT NULL COMMENT '创建时间',
    `f_update_time` bigint(20) NOT NULL COMMENT '编辑时间',
    PRIMARY KEY (`f_id`),
    UNIQUE KEY uk_version (f_version) USING BTREE
) ENGINE = InnoDB COMMENT = '函数元数据表';