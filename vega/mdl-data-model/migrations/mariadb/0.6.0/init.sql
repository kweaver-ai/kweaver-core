-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

USE kweaver;

-- 指标模型
CREATE TABLE IF NOT EXISTS t_metric_model (
  f_model_id varchar(40) NOT NULL DEFAULT '' COMMENT '指标模型 id',
  f_model_name varchar(40) NOT NULL COMMENT '指标模型名称',
  f_tags varchar(255) DEFAULT NULL COMMENT '标签',
  f_comment varchar(255) DEFAULT NULL COMMENT '备注',
  f_catalog_id varchar(40) NOT NULL DEFAULT '' COMMENT '编目id',
  f_catalog_content text DEFAULT NULL COMMENT '编目内容',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_measure_name varchar(50) NOT NULL DEFAULT '' COMMENT '度量名称',
  f_metric_type varchar(20) NOT NULL COMMENT '指标类型',
  f_data_source varchar(255) NOT NULL COMMENT '数据源',
  f_query_type varchar(20) NOT NULL COMMENT '指标查询语言',
  f_formula text NOT NULL COMMENT '计算公式',
  f_formula_config text DEFAULT NULL COMMENT '计算公式配置化',
  f_analysis_dimessions varchar(8192) DEFAULT NULL COMMENT '分析维度',
  f_order_by_fields varchar(4096) DEFAULT NULL COMMENT '排序字段',
  f_having_condition varchar(2048) DEFAULT NULL COMMENT '值过滤',
  f_date_field varchar(255) DEFAULT NULL COMMENT '时间字段',
  f_measure_field varchar(255) NOT NULL COMMENT '度量字段',
  f_unit_type varchar(40) NOT NULL COMMENT '单位类型',
  f_unit varchar(20) NOT NULL COMMENT '度量单位',
  f_group_id varchar(40) NOT NULL DEFAULT '' COMMENT '指标模型分组 id',
  f_builtin tinyint(2) DEFAULT 0 COMMENT '内置模型标识: 0 非内置, 1 内置',
  f_calendar_interval tinyint(2) DEFAULT 0 COMMENT '是否日历间隔。0: 非日历间隔; 1: 日历间隔',
  PRIMARY KEY (f_model_id),
  UNIQUE KEY uk_model_name (f_group_id, f_model_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '指标模型';


-- 指标模型分组
CREATE TABLE IF NOT EXISTS t_metric_model_group (
  f_group_id varchar(40) NOT NULL DEFAULT '' COMMENT '指标模型分组 id',
  f_group_name varchar(40) NOT NULL COMMENT '指标模型分组名称',
  f_comment varchar(255) NOT NULL DEFAULT '' COMMENT '指标模型分组备注',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_builtin tinyint(2) NOT NULL DEFAULT 0 COMMENT '内置分组标识: 0 非内置, 1 内置',
  PRIMARY KEY (f_group_id),
  UNIQUE KEY uk_f_group_name (f_group_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '指标模型分组';


-- 指标模型持久化任务
CREATE TABLE IF NOT EXISTS t_metric_model_task(
  f_task_id varchar(40) NOT NULL DEFAULT '' COMMENT '任务 id',
  f_task_name varchar(40) NOT NULL COMMENT '任务名称',
  f_comment varchar(255) NOT NULL DEFAULT '' COMMENT '任务备注',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_module_type varchar(20) NOT NULL DEFAULT '' COMMENT '模块类型',
  f_model_id varchar(40) NOT NULL COMMENT '指标模型 id',
  f_schedule varchar(255) NOT NULL COMMENT '执行频率',
  f_variables text DEFAULT NULL COMMENT '变量过滤',
  f_time_windows varchar(1024) DEFAULT NULL COMMENT '时间窗口',
  f_steps varchar(255) NOT NULL DEFAULT '[]' COMMENT '持久化步长',
  f_plan_time bigint(20) NOT NULL DEFAULT 0 COMMENT '计划时间',
  f_index_base varchar(40) NOT NULL COMMENT '索引库类型',
  f_retrace_duration varchar(20) DEFAULT NULL COMMENT '追溯时长',
  f_schedule_sync_status tinyint(2) NOT NULL COMMENT '任务的同步状态。3: 完成',
  f_execute_status tinyint(2) DEFAULT 0 COMMENT '任务的执行状态。4: 执行成功; 5: 执行失败',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建任务的用户id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  PRIMARY KEY (f_task_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '指标持久化任务';


-- 指标索引库tsid的时间分割静态表
CREATE TABLE IF NOT EXISTS t_static_metric_index (
  f_id int(11) AUTO_INCREMENT COMMENT '唯一id编号',
  f_base_type varchar(40) NOT NULL COMMENT '指标索引库类型',
  f_split_time datetime DEFAULT current_timestamp() COMMENT '索引库的时间分割',
  PRIMARY KEY (f_id),
  UNIQUE KEY uk_f_index_base_type (f_base_type)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '指标索引库tsid的时间分割静态表';


-- 事件模型聚合规则
CREATE TABLE if not exists t_event_model_aggregate_rules (
  f_aggregate_rule_id varchar(40) NOT NULL DEFAULT '' COMMENT '聚合规则id',
  f_aggregate_rule_type varchar(40) NOT NULL COMMENT '聚合规则类型',
  f_aggregate_algo varchar(900) NOT NULL COMMENT '聚合算法',
  f_rule_priority int(11) NOT NULL COMMENT '规则优先级',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_group_fields varchar(255) DEFAULT '[]' COMMENT '分组字段',
  f_aggregate_analysis_algo varchar(1024) DEFAULT '{}' COMMENT "分析算法",
  PRIMARY KEY (f_aggregate_rule_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin;


-- 事件模型
CREATE TABLE if not exists t_event_models (
  f_event_model_id varchar(40) NOT NULL DEFAULT '' COMMENT '事件模型id',
  f_event_model_name varchar(255) NOT NULL COMMENT '事件模型名称',
  f_event_model_group_name varchar(40) NOT NULL DEFAULT '' COMMENT '事件模型分组名称',
  f_event_model_type varchar(40) NOT NULL COMMENT '事件模型类型',
  f_event_model_tags varchar(255) NOT NULL COMMENT '事件模型标签',
  f_event_model_comment varchar(255) DEFAULT NULL COMMENT '事件模型说明',
  f_data_source_type varchar(40) NOT NULL COMMENT '数据源类型',
  f_data_source varchar(900) DEFAULT NULL COMMENT '数据源对象id',
  f_detect_rule_id varchar(40) NOT NULL COMMENT '检测规则id',
  f_aggregate_rule_id varchar(40) NOT NULL COMMENT '聚合规则id',
  f_default_time_window varchar(40) NOT NULL COMMENT '默认时间窗口',
  f_is_active tinyint(2) DEFAULT 0 COMMENT '是否是定期执行模式',
  f_enable_subscribe tinyint(2) DEFAULT 0 COMMENT '是否是实时订阅模式',
  f_status tinyint(2) DEFAULT 0 COMMENT '是否启用',
  f_downstream_dependent_model varchar(1024) DEFAULT '' COMMENT '依赖模型',
  f_is_custom tinyint(2) NOT NULL COMMENT '是否个性化',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_event_model_id),
  UNIQUE KEY uk_f_model_name (f_event_model_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin;


-- 事件模型检测规则
CREATE TABLE if not exists t_event_model_detect_rules (
  f_detect_rule_id varchar(40) NOT NULL DEFAULT '' COMMENT '检测规则id',
  f_detect_rule_type varchar(40) NOT NULL COMMENT '检测规则类型',
  f_formula varchar(2014) DEFAULT NULL COMMENT '计算公式',
  f_detect_algo varchar(40) DEFAULT NULL  COMMENT '检测算法',
  f_detect_analysis_algo varchar(1024) DEFAULT '{}' COMMENT '分析算法',
  f_rule_priority int(11) NOT NULL COMMENT "规则优先级",
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_detect_rule_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin;


-- 事件模型持久化任务
CREATE TABLE IF NOT EXISTS t_event_model_task (
  f_task_id varchar(40) NOT NULL DEFAULT '' COMMENT '唯一id编号',
  f_model_id varchar(40) NOT NULL COMMENT '事件模型 id',
  f_storage_config varchar(255) NOT NULL COMMENT '存储配置',
  f_schedule varchar(255) NOT NULL COMMENT '执行频率',
  f_dispatch_config varchar(255) NOT NULL COMMENT '调度配置',
  f_execute_parameter varchar(255) NOT NULL COMMENT '执行参数',
  f_task_status tinyint(2) NOT NULL COMMENT '最近一次任务执行状态。4: 执行成功; 5: 执行失败',
  f_error_details varchar(2048) NOT NULL COMMENT '任务执行失败原因',
  f_status_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '执行状态更新时间',
  f_schedule_sync_status tinyint(2) NOT NULL COMMENT '任务的同步状态。3: 完成',
  f_downstream_dependent_task varchar(1024) DEFAULT '' COMMENT '依赖任务',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_task_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '事件模型持久化任务';


-- 事件模型异步任务记录
CREATE TABLE IF NOT EXISTS t_event_model_task_execution_records (
  f_run_id bigint(20) unsigned NOT NULL COMMENT '运行id',
  f_run_type varchar(40) NOT NULL COMMENT '任务类型',
  f_execute_parameter varchar(2048) NOT NULL COMMENT '执行参数',
  f_status varchar(40) DEFAULT '0' COMMENT '状态',
  f_error_details varchar(1024) NOT NULL COMMENT '错误原因',
  f_update_time datetime NOT NULL COMMENT '更新时间',
  f_create_time datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (f_run_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin;


-- 数据视图
CREATE TABLE IF NOT EXISTS t_data_view (
  f_view_id varchar(40) NOT NULL DEFAULT '' COMMENT '数据视图 id',
  f_view_name varchar(255) NOT NULL COMMENT '数据视图名称',
  f_technical_name varchar(255) NOT NULL DEFAULT '' COMMENT '技术名称',
  f_group_id varchar(40) NOT NULL DEFAULT '' COMMENT '数据视图分组 id',
  f_type varchar(10) NOT NULL DEFAULT '' COMMENT '视图类型',
  f_query_type varchar(10) NOT NULL DEFAULT '' COMMENT '查询类型',
  f_builtin tinyint(2) DEFAULT 0 COMMENT '内置视图标识: 0 非内置, 1 内置',
  f_tags varchar(255) NOT NULL DEFAULT '' COMMENT '标签',
  f_comment varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_data_source_type varchar(20) NOT NULL DEFAULT '' COMMENT '数据源类型',
  f_data_source_id varchar(40) NOT NULL DEFAULT '' COMMENT '数据源 id',
  f_file_name varchar(128) NOT NULL DEFAULT '' COMMENT '文件名',
  f_excel_config text DEFAULT NULL COMMENT 'excel 配置',
  f_data_scope longtext DEFAULT NULL COMMENT '数据范围',
  f_fields longtext DEFAULT NULL COMMENT '字段列表',
  f_status varchar(20) NOT NULL DEFAULT '' COMMENT '状态',
  f_metadata_form_id varchar(40) NOT NULL DEFAULT '' COMMENT '元数据表单 id',
  f_primary_keys varchar(255) NOT NULL DEFAULT '' COMMENT '主键列表',
  f_sql longtext DEFAULT NULL COMMENT '生成视图sql',
  f_meta_table_name varchar(1024) NOT NULL DEFAULT '' COMMENT '元数据表名',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_delete_time bigint(20) NOT NULL DEFAULT 0 COMMENT '删除时间',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_updater varchar(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type varchar(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  f_data_source text DEFAULT NULL COMMENT '废弃, 数据视图数据来源',
  f_field_scope tinyint(2) NOT NULL DEFAULT '0' COMMENT '废弃, 字段范围: 0 部分字段, 1 全部字段',
  f_filters text DEFAULT NULL COMMENT '废弃, 过滤条件',
  f_open_streaming tinyint(2) NOT NULL DEFAULT 0 COMMENT '废弃, 是否开启视图实时订阅任务: 0 不开启, 1 开启',
  f_job_id varchar(40) NOT NULL DEFAULT '' COMMENT '废弃, 订阅任务 id',
  f_loggroup_filters longtext DEFAULT NULL COMMENT '废弃, 日志分组过滤条件',
  PRIMARY KEY (f_view_id),
  UNIQUE KEY uk_f_view_name (f_group_id, f_view_name, f_delete_time)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据视图';


-- 数据视图分组
CREATE TABLE IF NOT EXISTS t_data_view_group (
  f_group_id varchar(40) NOT NULL DEFAULT '' COMMENT '数据视图分组 id',
  f_group_name varchar(40) NOT NULL COMMENT '数据视图分组名称',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_delete_time bigint(20) NOT NULL DEFAULT 0 COMMENT '删除时间',
  f_builtin tinyint(2) NOT NULL DEFAULT 0 COMMENT '内置视图标识: 0 非内置, 1 内置',
  PRIMARY KEY (f_group_id),
  UNIQUE KEY uk_f_group_name (f_builtin, f_group_name, f_delete_time)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据视图分组';


-- 扫描记录
CREATE TABLE IF NOT EXISTS t_scan_record (
    f_record_id varchar(40) NOT NULL DEFAULT '' COMMENT '扫描记录 id',
    f_data_source_id varchar(40) NOT NULL COMMENT '数据源 id',
    f_scanner varchar(40) NOT NULL COMMENT '扫描器',
    f_scan_time bigint(20) NOT NULL DEFAULT 0 COMMENT '扫描时间',
    f_data_source_status varchar(20) NOT NULL DEFAULT '' COMMENT '数据源状态: available 可用 scanning 扫描中',
    f_metadata_task_id varchar(128)  DEFAULT NULL COMMENT '元数据采集平台任务id',
    PRIMARY KEY (f_record_id),
    UNIQUE KEY uk_scan_record (f_data_source_id, f_scanner)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据源扫描记录表';


-- 视图行列规则表
CREATE TABLE IF NOT EXISTS t_data_view_row_column_rule (
  f_rule_id varchar(40) NOT NULL DEFAULT '' COMMENT '视图行列规则 id',
  f_rule_name varchar(255) NOT NULL COMMENT '视图行列规则名称',
  f_view_id varchar(40) NOT NULL COMMENT '视图 id',
  f_tags varchar(255) NOT NULL DEFAULT '' COMMENT '标签',
  f_comment varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_fields longtext NOT NULL COMMENT '列',
  f_row_filters text NOT NULL COMMENT '行过滤规则',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_updater varchar(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type varchar(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  PRIMARY KEY (f_rule_id),
  UNIQUE KEY uk_f_rule_name (f_rule_name, f_view_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据视图行列规则';


-- 数据字典
CREATE TABLE IF NOT EXISTS t_data_dict (
  f_dict_id varchar(40) NOT NULL COMMENT '数据字典id',
  f_dict_name varchar(255) NOT NULL COMMENT '数据字典名称',
  f_tags varchar(255) NOT NULL COMMENT '标签',
  f_comment varchar(255) DEFAULT NULL COMMENT '备注',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_dict_type varchar(20) NOT NULL DEFAULT 'kv_dict' COMMENT '数据字典类型',
  f_dict_store varchar(255) NOT NULL COMMENT '数据字典的项存放的对应表名称',
  f_dimension varchar(1500) NOT NULL COMMENT '数据字典维度关系',
  f_unique_key tinyint(2) NOT NULL DEFAULT 1 COMMENT '是否唯一键',
  PRIMARY KEY (f_dict_id),
  UNIQUE KEY uk_dict_name (f_dict_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据字典信息';


-- 数据字典项
CREATE TABLE IF NOT EXISTS t_data_dict_item (
  f_item_id varchar(40) NOT NULL COMMENT '数据字典项id',
  f_dict_id varchar(40) NOT NULL COMMENT '数据字典id',
  f_item_key varchar(3000) NOT NULL COMMENT '数据字典项key值',
  f_item_value varchar(3000) NOT NULL COMMENT '数据字典项value值',
  f_comment varchar(255) COMMENT '数据字典项说明',
  PRIMARY KEY (f_item_id),
  KEY idx_dict_id (f_dict_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据字典项信息表';


-- 数据连接
CREATE TABLE IF NOT EXISTS t_data_connection (
  f_connection_id varchar(40) NOT NULL COMMENT '唯一id编号',
  f_connection_name varchar(40) NOT NULL COMMENT '数据连接名称',
  f_tags varchar(255) DEFAULT '' COMMENT '标签',
  f_comment varchar(255) DEFAULT '' COMMENT '数据连接备注',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_data_source_type varchar(40) NOT NULL COMMENT '数据源类型',
  f_config text NOT NULL COMMENT '详细配置',
  f_config_md5 varchar(32) DEFAULT '' COMMENT '详细配置的唯一标识符',
  PRIMARY KEY (f_connection_id),
  UNIQUE KEY uk_f_connection_name (f_connection_name),
  KEY idx_f_data_source_type (f_data_source_type),
  KEY idx_f_config_md5 (f_config_md5)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据连接';


-- 数据连接状态
CREATE TABLE IF NOT EXISTS t_data_connection_status (
  f_connection_id varchar(40) NOT NULL COMMENT '数据连接id',
  f_status varchar(5) NOT NULL COMMENT '连接状态',
  f_detection_time bigint(20) NOT NULL DEFAULT 0 COMMENT '检测时间',
  PRIMARY KEY (f_connection_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据连接状态';


-- 链路模型
CREATE TABLE IF NOT EXISTS t_trace_model (
  f_model_id varchar(40) NOT NULL COMMENT '唯一id编号',
  f_model_name varchar(40) NOT NULL COMMENT '链路模型名称',
  f_tags varchar(255) DEFAULT '' COMMENT '标签',
  f_comment varchar(255) DEFAULT '' COMMENT '链路模型备注',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_span_source_type varchar(40) NOT NULL COMMENT 'span数据来源类型',
  f_span_config text NOT NULL COMMENT 'span配置',
  f_enabled_related_log tinyint(2) NOT NULL COMMENT '是否开启配置span关联日志配置, 0表示否, 1表示是',
  f_related_log_source_type varchar(40) NOT NULL COMMENT 'span关联日志数据来源类型',
  f_related_log_config text NOT NULL COMMENT 'span关联日志配置',
  PRIMARY KEY (f_model_id),
  UNIQUE KEY uk_f_model_name (f_model_name),
  KEY idx_f_span_source_type (f_span_source_type)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '链路模型';


-- 全局任务表
CREATE TABLE IF NOT EXISTS t_data_model_job (
  f_job_id varchar(40) NOT NULL DEFAULT '' COMMENT '任务 id',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_job_type varchar(40) NOT NULL COMMENT '任务类型',
  f_job_config text COMMENT '任务配置',
  f_job_status varchar(20) NOT NULL COMMENT '任务状态: running 正常, error 异常',
  f_job_status_details text NOT NULL COMMENT '任务状态详情',
  PRIMARY KEY (f_job_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '全局任务表';


-- 目标模型
CREATE TABLE IF NOT EXISTS t_objective_model (
  f_model_id varchar(40) NOT NULL DEFAULT '' COMMENT '目标模型 id',
  f_model_name varchar(40) NOT NULL COMMENT '目标模型名称',
  f_tags varchar(255) NOT NULL DEFAULT '' COMMENT '标签',
  f_comment varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_objective_type varchar(20) NOT NULL COMMENT '目标类型',
  f_objective_config text NOT NULL COMMENT '目标配置',
  PRIMARY KEY (f_model_id),
  UNIQUE KEY uk_t_objective_model (f_model_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '目标模型';


-- --------------------------------------- 初始化数据 --------------------------------------------
-- 未分组
INSERT INTO t_data_view_group (
  f_group_id,
  f_group_name,
  f_create_time,
  f_update_time,
  f_builtin
)
SELECT '', '', 1733903782147, 1733903782147, 0
FROM DUAL
WHERE NOT EXISTS(
  SELECT f_group_id
  FROM t_data_view_group
  WHERE f_group_id = ''
);

-- 索引库
INSERT INTO t_data_view_group (
  f_group_id,
  f_group_name,
  f_create_time,
  f_update_time,
  f_builtin
)
SELECT '__index_base', 'index_base', 1733903782147, 1733903782147, 1
FROM DUAL
WHERE NOT EXISTS(
  SELECT f_group_id
  FROM t_data_view_group
  WHERE f_group_id = '__index_base'
);

-- 未分组
INSERT INTO t_metric_model_group (
  f_group_id,
  f_group_name,
  f_create_time,
  f_update_time,
  f_builtin
)
SELECT '', '', 1733903782147, 1733903782147, 0
FROM DUAL
WHERE NOT EXISTS(
  SELECT f_group_id
  FROM t_metric_model_group
  WHERE f_group_id = ''
);
