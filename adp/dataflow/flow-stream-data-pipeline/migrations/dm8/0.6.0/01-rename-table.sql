SET SCHEMA kweaver;


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

-- t_stream_data_pipeline
INSERT INTO kweaver."t_stream_data_pipeline" ("f_pipeline_id", "f_pipeline_name", "f_tags", "f_comment", "f_builtin", "f_output_type", "f_index_base", "f_use_index_base_in_data", "f_pipeline_status", "f_pipeline_status_details", "f_deployment_config", "f_create_time", "f_update_time", "f_creator", "f_updater", "f_creator_type", "f_updater_type")
SELECT s."f_pipeline_id", s."f_pipeline_name", s."f_tags", s."f_comment", s."f_builtin", s."f_output_type", s."f_index_base", s."f_use_index_base_in_data", s."f_pipeline_status", s."f_pipeline_status_details", s."f_deployment_config", s."f_create_time", s."f_update_time", s."f_creator", s."f_updater", s."f_creator_type", s."f_updater_type"
FROM adp."t_stream_data_pipeline" s;
