USE adp;

-- 内部应用
CREATE TABLE IF NOT EXISTS t_internal_app (
  f_app_id varchar(40) NOT NULL COMMENT 'app_id',
  f_app_name varchar(40) NOT NULL COMMENT 'app名称',
  f_app_secret varchar(40) NOT NULL COMMENT 'app_secret',
  f_create_time bigint(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (f_app_id),
  UNIQUE KEY uk_app_name (f_app_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '内部应用';


CREATE TABLE IF NOT EXISTS t_stream_data_pipeline (
  f_pipeline_id varchar(40) NOT NULL DEFAULT '' COMMENT '管道 id',
  f_pipeline_name varchar(40) NOT NULL DEFAULT '' COMMENT '管道名称',
  f_tags varchar(255) NOT NULL COMMENT '标签',
  f_comment varchar(255) COMMENT '备注',
  f_builtin boolean DEFAULT 0 COMMENT '内置管道标识: 0 非内置, 1 内置',
  f_output_type varchar(20) NOT NULL COMMENT '数据输出类型',
  f_index_base varchar(255) NOT NULL COMMENT '索引库类型',
  f_use_index_base_in_data boolean DEFAULT 0 COMMENT '是否使用数据里的索引库: 0 否, 1 是',
  f_pipeline_status varchar(10) NOT NULL COMMENT '管道状态: failed 失败, running 运行中, close 关闭',
  f_pipeline_status_details text NOT NULL COMMENT '管道状态详情',
  f_deployment_config text NOT NULL COMMENT '资源配置信息',
  f_create_time bigint(20) NOT NULL default 0 COMMENT '创建时间',
  f_update_time bigint(20) NOT NULL default 0 COMMENT '更新时间',
  f_creator varchar(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type varchar(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_updater varchar(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type varchar(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  PRIMARY KEY (f_pipeline_id),
  UNIQUE KEY uk_name (f_pipeline_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '流式数据管道信息';