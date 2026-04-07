
SET SEARCH_PATH TO adp;


CREATE TABLE IF NOT EXISTS t_internal_app (
  f_app_id VARCHAR(40) NOT NULL COMMENT 'app_id',
  f_app_name VARCHAR(40) NOT NULL COMMENT 'app名称',
  f_app_secret VARCHAR(40) NOT NULL COMMENT 'app_secret',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (f_app_id),
  UNIQUE KEY `idx_t_internal_app_uk_app_name` (f_app_name)
);



CREATE TABLE IF NOT EXISTS t_stream_data_pipeline (
  f_pipeline_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '管道 id',
  f_pipeline_name VARCHAR(40) NOT NULL DEFAULT '' COMMENT '管道名称',
  f_tags VARCHAR(255) NOT NULL COMMENT '标签',
  f_comment VARCHAR(255) COMMENT '备注',
  f_builtin TINYINT(1) DEFAULT 0 COMMENT '内置管道标识: 0 非内置, 1 内置',
  f_output_type VARCHAR(20) NOT NULL COMMENT '数据输出类型',
  f_index_base VARCHAR(255) NOT NULL COMMENT '索引库类型',
  f_use_index_base_in_data TINYINT(1) DEFAULT 0 COMMENT '是否使用数据里的索引库: 0 否, 1 是',
  f_pipeline_status VARCHAR(10) NOT NULL COMMENT '管道状态: failed 失败, running 运行中, close 关闭',
  f_pipeline_status_details TEXT NOT NULL COMMENT '管道状态详情',
  f_deployment_config TEXT NOT NULL COMMENT '资源配置信息',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  PRIMARY KEY (f_pipeline_id),
  UNIQUE KEY `idx_t_stream_data_pipeline_uk_name` (f_pipeline_name)
);


