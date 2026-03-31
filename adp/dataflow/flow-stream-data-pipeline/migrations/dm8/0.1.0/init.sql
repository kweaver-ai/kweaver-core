SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS t_internal_app (
  f_app_id VARCHAR(40 CHAR) NOT NULL,
  f_app_name VARCHAR(40 CHAR) NOT NULL,
  f_app_secret VARCHAR(40 CHAR) NOT NULL,
  f_create_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_app_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_internal_app_uk_app_name ON t_internal_app(f_app_name);

CREATE TABLE IF NOT EXISTS t_stream_data_pipeline (
  f_pipeline_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_pipeline_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) NOT NULL,
  f_comment VARCHAR(255 CHAR),
  f_builtin TINYINT DEFAULT 0,
  f_output_type VARCHAR(20 CHAR) NOT NULL,
  f_index_base VARCHAR(255 CHAR) NOT NULL,
  f_use_index_base_in_data TINYINT DEFAULT 0,
  f_pipeline_status VARCHAR(10 CHAR) NOT NULL,
  f_pipeline_status_details text NOT NULL,
  f_deployment_config text NOT NULL,
  f_create_time BIGINT NOT NULL default 0,
  f_update_time BIGINT NOT NULL default 0,
  "f_creator" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    "f_updater" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    "f_creator_type" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    "f_updater_type" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY (f_pipeline_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_stream_data_pipeline_uk_name ON t_stream_data_pipeline(f_pipeline_name);

